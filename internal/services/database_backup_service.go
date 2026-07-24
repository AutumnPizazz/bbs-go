package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/glebarez/sqlite"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/pkg/config"
)

const (
	BackupTriggerManual        = "manual"
	BackupTriggerScheduled     = "scheduled"
	BackupTriggerRestoreSafety = "restore-safety"
	BackupConfirmText          = "CREATE_BACKUP"
	BackupDeleteConfirmText    = "DELETE_BACKUP"
	BackupDownloadConfirmText  = "DOWNLOAD_BACKUP"
	BackupRestoreConfirmText   = "RESTORE_BACKUP"
	maxBackupErrorLength       = 1024

	backupPhaseDumping       = "dumping"
	backupPhaseVerifying     = "verifying"
	backupPhaseFailed        = "failed"
	backupPhaseCompleted     = "completed"
	restorePhasePreparing    = "preparing"
	restorePhaseSafetyBackup = "safety-backup"
	restorePhaseRestoring    = "restoring"
	restorePhaseFailed       = "failed"
	restorePhaseCompleted    = "completed"
	backupMetadataTable      = "database_backups"
	restoreMetadataTable     = "database_restores"
)

type DatabaseBackupStatus struct {
	Running       bool                  `json:"running"`
	LastSuccessAt int64                 `json:"lastSuccessAt"`
	LastFailureAt int64                 `json:"lastFailureAt"`
	LastError     string                `json:"lastError,omitempty"`
	Retention     int                   `json:"retention"`
	Directory     string                `json:"directory"`
	Progress      int                   `json:"progress"`
	Phase         string                `json:"phase,omitempty"`
	Restore       DatabaseRestoreStatus `json:"restore"`
}

type DatabaseRestoreStatus struct {
	Running       bool   `json:"running"`
	LastSuccessAt int64  `json:"lastSuccessAt"`
	LastFailureAt int64  `json:"lastFailureAt"`
	LastError     string `json:"lastError,omitempty"`
	LastBackupId  int64  `json:"lastBackupId"`
	Progress      int    `json:"progress"`
	Phase         string `json:"phase,omitempty"`
}

type DatabaseBackupConfig struct {
	Enabled   bool   `json:"enabled"`
	Schedule  string `json:"schedule"`
	Retention int    `json:"retention"`
	Directory string `json:"directory"`
}

type databaseBackupService struct {
	mu sync.Mutex
}

var DatabaseBackupService = &databaseBackupService{}

func (s *databaseBackupService) Config() DatabaseBackupConfig {
	if config.Instance == nil {
		return DatabaseBackupConfig{Schedule: config.DefaultBackupSchedule, Retention: config.DefaultBackupRetention, Directory: config.DefaultBackupDirectory}
	}
	cfg := config.Instance.Backup
	config.SetBackupDefaults(&cfg)
	return DatabaseBackupConfig{Enabled: cfg.Enabled, Schedule: cfg.Schedule, Retention: cfg.Retention, Directory: cfg.Directory}
}

func (s *databaseBackupService) Status() DatabaseBackupStatus {
	status := DatabaseBackupStatus{}
	if db := sqls.DB(); db != nil {
		var running models.DatabaseBackup
		if db.Where("status = ?", models.DatabaseBackupRunning).Order("id desc").First(&running).Error == nil {
			status.Running = true
			status.Progress = running.Progress
			status.Phase = running.Phase
		}
		var successful models.DatabaseBackup
		if db.Where("status = ?", models.DatabaseBackupSuccess).Order("finished_at desc").First(&successful).Error == nil {
			status.LastSuccessAt = successful.FinishedAt
		}
		var failed models.DatabaseBackup
		if db.Where("status = ?", models.DatabaseBackupFailed).Order("finished_at desc").First(&failed).Error == nil {
			status.LastFailureAt = failed.FinishedAt
			status.LastError = failed.Error
		}
		status.Restore = s.restoreStatus(db)
		status.Running = status.Running || status.Restore.Running
	}
	cfg := s.Config()
	status.Retention = cfg.Retention
	status.Directory = cfg.Directory
	return status
}

func (s *databaseBackupService) ValidateConfig(next config.BackupConfig) error {
	config.SetBackupDefaults(&next)
	if next.Retention < 1 || next.Retention > 100 {
		return errors.New("backup retention must be between 1 and 100")
	}
	if _, err := cron.ParseStandard(next.Schedule); err != nil {
		return fmt.Errorf("invalid backup schedule: %w", err)
	}
	if _, err := s.directory(next.Directory); err != nil {
		return err
	}
	return nil
}

func (s *databaseBackupService) SaveConfig(next config.BackupConfig) error {
	if err := s.ValidateConfig(next); err != nil {
		return err
	}
	return config.SaveBackupConfig(next)
}

func (s *databaseBackupService) StartManual(requestedBy int64) (*models.DatabaseBackup, error) {
	return s.start(BackupTriggerManual, requestedBy)
}

func (s *databaseBackupService) StartScheduled() (*models.DatabaseBackup, error) {
	if !s.Config().Enabled {
		return nil, errors.New("scheduled database backups are disabled")
	}
	return s.start(BackupTriggerScheduled, 0)
}

func (s *databaseBackupService) StartRestore(backupId, requestedBy int64) (*models.DatabaseRestore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := sqls.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if s.databaseType() != config.DbTypeMySQL {
		return nil, errors.New("database restore is only supported for MySQL")
	}
	backup := s.get(backupId)
	if backup == nil || backup.Status != models.DatabaseBackupSuccess {
		return nil, errors.New("backup not found or is not successful")
	}
	if !strings.EqualFold(backup.DatabaseType, config.DbTypeMySQL) {
		return nil, errors.New("backup database type is not compatible with MySQL")
	}
	if !strings.EqualFold(filepath.Ext(backup.FileName), ".sql") {
		return nil, errors.New("MySQL restore requires a .sql backup file")
	}
	path, err := s.pathForBackup(backup)
	if err != nil {
		return nil, err
	}
	if err := s.verifyFile(path, config.DbTypeMySQL, backup.Checksum); err != nil {
		return nil, fmt.Errorf("backup validation failed: %w", err)
	}
	if s.backupRunning(db) {
		return nil, errors.New("another database backup is already running")
	}
	if s.restoreRunning(db) {
		return nil, errors.New("another database restore is already running")
	}

	now := dates.NowTimestamp()
	task := &models.DatabaseRestore{
		BackupId: backupId, DatabaseType: config.DbTypeMySQL, Status: models.DatabaseRestoreRunning,
		Progress: 5, Phase: restorePhasePreparing,
		RequestedBy: requestedBy, StartedAt: now, CreateTime: now, UpdateTime: now,
	}
	if err := db.Create(task).Error; err != nil {
		return nil, err
	}
	go s.runRestore(task.Id)
	return task, nil
}

// RecoverStale marks interrupted backups as failed so a restart cannot leave
// the persistent queue blocked forever.
func (s *databaseBackupService) RecoverStale() {
	db := sqls.DB()
	if db == nil {
		return
	}
	var backups []models.DatabaseBackup
	if err := db.Where("status = ?", models.DatabaseBackupRunning).Find(&backups).Error; err != nil {
		slog.Warn("failed to load interrupted database backups", slog.Any("error", err))
		return
	}
	for _, backup := range backups {
		if path, err := s.pathForBackup(&backup); err == nil {
			_ = os.Remove(path)
		}
		if err := db.Model(&models.DatabaseBackup{}).Where("id = ? AND status = ?", backup.Id, models.DatabaseBackupRunning).Updates(map[string]interface{}{
			"status": models.DatabaseBackupFailed, "phase": backupPhaseFailed, "error": "backup interrupted by process restart", "finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
		}).Error; err != nil {
			slog.Warn("failed to mark interrupted database backup", slog.Int64("id", backup.Id), slog.Any("error", err))
		}
	}
	var restores []models.DatabaseRestore
	if err := db.Where("status = ?", models.DatabaseRestoreRunning).Find(&restores).Error; err != nil {
		slog.Warn("failed to load interrupted database restores", slog.Any("error", err))
		return
	}
	for _, restore := range restores {
		if err := db.Model(&models.DatabaseRestore{}).Where("id = ? AND status = ?", restore.Id, models.DatabaseRestoreRunning).Updates(map[string]interface{}{
			"status": models.DatabaseRestoreFailed, "phase": restorePhaseFailed, "error": "restore interrupted by process restart", "finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
		}).Error; err != nil {
			slog.Warn("failed to mark interrupted database restore", slog.Int64("id", restore.Id), slog.Any("error", err))
		}
	}
}

func (s *databaseBackupService) start(trigger string, requestedBy int64) (*models.DatabaseBackup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db := sqls.DB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if s.backupRunning(db) {
		return nil, errors.New("another database backup is already running")
	}
	if s.restoreRunning(db) {
		return nil, errors.New("another database restore is already running")
	}
	backup, err := s.createBackupRecord(trigger, requestedBy)
	if err != nil {
		return nil, err
	}
	go s.run(backup.Id)
	return backup, nil
}

func (s *databaseBackupService) run(id int64) {
	backup := s.get(id)
	if backup == nil {
		return
	}
	if err := s.executeBackup(backup); err == nil {
		s.prune()
	}
}

func (s *databaseBackupService) createBackupRecord(trigger string, requestedBy int64) (*models.DatabaseBackup, error) {
	now := dates.NowTimestamp()
	backup := &models.DatabaseBackup{
		TriggerType: trigger, DatabaseType: s.databaseType(),
		FileName: s.newFileName(trigger), Directory: s.Config().Directory,
		Status: models.DatabaseBackupRunning, Progress: 5, Phase: backupPhaseDumping,
		RequestedBy: requestedBy, StartedAt: now, CreateTime: now, UpdateTime: now,
	}
	if err := sqls.DB().Create(backup).Error; err != nil {
		return nil, err
	}
	return backup, nil
}

func (s *databaseBackupService) executeBackup(backup *models.DatabaseBackup) error {
	s.updateBackupProgress(backup.Id, 10, backupPhaseDumping)
	path, size, checksum, err := s.createFile(backup.FileName, backup.Directory)
	if err != nil {
		s.markBackupFailed(backup.Id, err)
		return err
	}
	s.updateBackupProgress(backup.Id, 85, backupPhaseVerifying)
	if err := s.verifyFile(path, backup.DatabaseType, checksum); err != nil {
		_ = os.Remove(path)
		s.markBackupFailed(backup.Id, err)
		return err
	}
	updates := map[string]interface{}{
		"status": models.DatabaseBackupSuccess, "phase": backupPhaseCompleted, "progress": 100, "error": "", "size": size,
		"checksum": checksum, "finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
	}
	if err := sqls.DB().Model(&models.DatabaseBackup{}).Where("id = ?", backup.Id).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (s *databaseBackupService) markBackupFailed(id int64, cause error) {
	_ = sqls.DB().Model(&models.DatabaseBackup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": models.DatabaseBackupFailed, "phase": backupPhaseFailed, "error": truncateBackupError(cause.Error()),
		"finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
	}).Error
}

func (s *databaseBackupService) runRestore(id int64) {
	s.mu.Lock()
	restoreSucceeded := false
	defer func() {
		s.mu.Unlock()
		if restoreSucceeded {
			s.prune()
		}
	}()

	task := s.getRestore(id)
	if task == nil {
		return
	}
	s.updateRestoreProgress(id, 10, restorePhasePreparing)
	backup := s.get(task.BackupId)
	if backup == nil || backup.Status != models.DatabaseBackupSuccess {
		s.markRestoreFailed(id, errors.New("backup not found or is not successful"))
		return
	}
	path, err := s.pathForBackup(backup)
	if err != nil {
		s.markRestoreFailed(id, err)
		return
	}
	if err := s.verifyFile(path, config.DbTypeMySQL, backup.Checksum); err != nil {
		s.markRestoreFailed(id, fmt.Errorf("backup validation failed: %w", err))
		return
	}

	// The safety snapshot is completed before the selected dump is imported.
	s.updateRestoreProgress(id, 25, restorePhaseSafetyBackup)
	safetyBackup, err := s.createSafetyBackup(task.RequestedBy)
	if err != nil {
		s.markRestoreFailed(id, fmt.Errorf("safety backup failed: %w", err))
		return
	}
	if err := sqls.DB().Model(&models.DatabaseRestore{}).Where("id = ?", id).Updates(map[string]interface{}{
		"safety_backup_id": safetyBackup.Id, "update_time": dates.NowTimestamp(),
	}).Error; err != nil {
		s.markRestoreFailed(id, err)
		return
	}

	s.updateRestoreProgress(id, 55, restorePhaseRestoring)
	if err := s.restoreMySQL(path); err != nil {
		s.markRestoreFailed(id, err)
		return
	}
	if err := s.finalizeRestoreMetadata(task, backup, safetyBackup); err != nil {
		s.markRestoreFailed(id, fmt.Errorf("database restore completed but status finalization failed: %w", err))
		return
	}
	restoreSucceeded = true
}

func (s *databaseBackupService) createSafetyBackup(requestedBy int64) (*models.DatabaseBackup, error) {
	backup, err := s.createBackupRecord(BackupTriggerRestoreSafety, requestedBy)
	if err != nil {
		return nil, err
	}
	if err := s.executeBackup(backup); err != nil {
		return backup, err
	}
	completed := s.get(backup.Id)
	if completed == nil {
		return nil, errors.New("safety backup record was not found after completion")
	}
	return completed, nil
}

func (s *databaseBackupService) restoreMySQL(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	args, env, err := s.restoreCommandArgs()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = file
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return errors.New(message)
	}
	return nil
}

func (s *databaseBackupService) restoreCommandArgs() ([]string, []string, error) {
	if s.databaseType() != config.DbTypeMySQL || config.Instance == nil {
		return nil, nil, errors.New("database restore is only supported for MySQL")
	}
	cfg, err := mysql.ParseDSN(config.Instance.DB.Url)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid MySQL connection string: %w", err)
	}
	args := []string{"--binary-mode", "--user=" + cfg.User, "--database=" + cfg.DBName}
	if cfg.Net == "unix" {
		args = append(args, "--socket="+cfg.Addr)
	} else {
		host, port := splitMySQLAddress(cfg.Addr)
		args = append(args, "--host="+host)
		if port != "" {
			args = append(args, "--port="+port)
		}
	}
	return args, []string{"MYSQL_PWD=" + cfg.Passwd}, nil
}

func (s *databaseBackupService) markRestoreFailed(id int64, cause error) {
	_ = sqls.DB().Model(&models.DatabaseRestore{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": models.DatabaseRestoreFailed, "phase": restorePhaseFailed, "error": truncateBackupError(cause.Error()),
		"finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
	}).Error
}

// finalizeRestoreMetadata repairs the task tables when an older dump included
// them and replaced the records created for the current restore.
func (s *databaseBackupService) finalizeRestoreMetadata(task *models.DatabaseRestore, backup, safetyBackup *models.DatabaseBackup) error {
	db := sqls.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	if err := db.AutoMigrate(&models.DatabaseBackup{}, &models.DatabaseRestore{}); err != nil {
		return err
	}

	restoredBackup, err := s.ensureBackupRecord(db, backup)
	if err != nil {
		return fmt.Errorf("restore selected backup metadata: %w", err)
	}
	restoredSafetyBackup, err := s.ensureBackupRecord(db, safetyBackup)
	if err != nil {
		return fmt.Errorf("restore safety backup metadata: %w", err)
	}

	now := dates.NowTimestamp()
	if result := db.Model(&models.DatabaseRestore{}).Where("status = ? AND id <> ?", models.DatabaseRestoreRunning, task.Id).Updates(map[string]interface{}{
		"status": models.DatabaseRestoreFailed, "phase": restorePhaseFailed,
		"error": "restore metadata was replaced during a database restore", "finished_at": now, "update_time": now,
	}); result.Error != nil {
		return result.Error
	}

	updates := map[string]interface{}{
		"backup_id": restoredBackup.Id, "database_type": task.DatabaseType, "status": models.DatabaseRestoreSuccess,
		"requested_by": task.RequestedBy, "safety_backup_id": restoredSafetyBackup.Id,
		"progress": 100, "phase": restorePhaseCompleted, "error": "", "started_at": task.StartedAt,
		"finished_at": now, "update_time": now,
	}
	result := db.Model(&models.DatabaseRestore{}).Where("id = ?", task.Id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	restoredTask := *task
	restoredTask.Id = 0
	restoredTask.BackupId = restoredBackup.Id
	restoredTask.Status = models.DatabaseRestoreSuccess
	restoredTask.SafetyBackupId = restoredSafetyBackup.Id
	restoredTask.Progress = 100
	restoredTask.Phase = restorePhaseCompleted
	restoredTask.Error = ""
	restoredTask.FinishedAt = now
	restoredTask.UpdateTime = now
	if err := db.Create(&restoredTask).Error; err != nil {
		return err
	}
	return nil
}

func (s *databaseBackupService) ensureBackupRecord(db *gorm.DB, source *models.DatabaseBackup) (*models.DatabaseBackup, error) {
	if source == nil {
		return nil, errors.New("backup metadata is missing")
	}
	var existing models.DatabaseBackup
	err := db.Where("file_name = ?", source.FileName).First(&existing).Error
	if err == nil {
		restored := *source
		restored.Id = existing.Id
		if err := db.Save(&restored).Error; err != nil {
			return nil, err
		}
		return &restored, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	restored := *source
	restored.Id = 0
	if err := db.Create(&restored).Error; err != nil {
		return nil, err
	}
	return &restored, nil
}

func (s *databaseBackupService) updateBackupProgress(id int64, progress int, phase string) {
	if db := sqls.DB(); db != nil {
		_ = db.Model(&models.DatabaseBackup{}).Where("id = ? AND status = ?", id, models.DatabaseBackupRunning).Updates(map[string]interface{}{
			"progress": clampBackupProgress(progress), "phase": phase, "update_time": dates.NowTimestamp(),
		}).Error
	}
}

func (s *databaseBackupService) updateRestoreProgress(id int64, progress int, phase string) {
	if db := sqls.DB(); db != nil {
		_ = db.Model(&models.DatabaseRestore{}).Where("id = ? AND status = ?", id, models.DatabaseRestoreRunning).Updates(map[string]interface{}{
			"progress": clampBackupProgress(progress), "phase": phase, "update_time": dates.NowTimestamp(),
		}).Error
	}
}

func clampBackupProgress(progress int) int {
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}
	return progress
}

func (s *databaseBackupService) getRestore(id int64) *models.DatabaseRestore {
	if db := sqls.DB(); db != nil {
		var restore models.DatabaseRestore
		if db.First(&restore, id).Error == nil {
			return &restore
		}
	}
	return nil
}

func (s *databaseBackupService) backupRunning(db *gorm.DB) bool {
	var count int64
	return db.Model(&models.DatabaseBackup{}).Where("status = ?", models.DatabaseBackupRunning).Count(&count).Error == nil && count > 0
}

func (s *databaseBackupService) restoreRunning(db *gorm.DB) bool {
	var count int64
	return db.Model(&models.DatabaseRestore{}).Where("status = ?", models.DatabaseRestoreRunning).Count(&count).Error == nil && count > 0
}

func (s *databaseBackupService) restoreStatus(db *gorm.DB) DatabaseRestoreStatus {
	status := DatabaseRestoreStatus{}
	var running models.DatabaseRestore
	if db.Where("status = ?", models.DatabaseRestoreRunning).Order("id desc").First(&running).Error == nil {
		status.Running = true
		status.Progress = running.Progress
		status.Phase = running.Phase
	}
	var successful models.DatabaseRestore
	if db.Where("status = ?", models.DatabaseRestoreSuccess).Order("finished_at desc").First(&successful).Error == nil {
		status.LastSuccessAt = successful.FinishedAt
		status.LastBackupId = successful.BackupId
	}
	var failed models.DatabaseRestore
	if db.Where("status = ?", models.DatabaseRestoreFailed).Order("finished_at desc").First(&failed).Error == nil {
		status.LastFailureAt = failed.FinishedAt
		status.LastError = failed.Error
		status.LastBackupId = failed.BackupId
	}
	return status
}

func (s *databaseBackupService) List(limit int) []models.DatabaseBackup {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	var backups []models.DatabaseBackup
	if db := sqls.DB(); db != nil {
		db.Order("id desc").Limit(limit).Find(&backups)
	}
	return backups
}

func (s *databaseBackupService) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup := s.get(id)
	if backup == nil {
		return errors.New("backup not found")
	}
	if backup.Status == models.DatabaseBackupRunning {
		return errors.New("running backups cannot be deleted")
	}
	if db := sqls.DB(); db != nil {
		var runningRestore int64
		if err := db.Model(&models.DatabaseRestore{}).Where("backup_id = ? AND status = ?", id, models.DatabaseRestoreRunning).Count(&runningRestore).Error; err == nil && runningRestore > 0 {
			return errors.New("backup is being used by a running restore")
		}
	}
	path, err := s.pathForBackup(backup)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return sqls.DB().Delete(&models.DatabaseBackup{}, id).Error
}

func (s *databaseBackupService) Get(id int64) *models.DatabaseBackup { return s.get(id) }

func (s *databaseBackupService) FilePath(backup *models.DatabaseBackup) (string, error) {
	if backup == nil {
		return "", errors.New("backup not found")
	}
	return s.pathForBackup(backup)
}

func (s *databaseBackupService) StartScheduler(c *cron.Cron) {
	if config.Instance == nil {
		return
	}
	cfg := config.Instance.Backup
	config.SetBackupDefaults(&cfg)
	if !cfg.Enabled {
		return
	}
	if _, err := cron.ParseStandard(cfg.Schedule); err != nil {
		return
	}
	_, _ = c.AddFunc(cfg.Schedule, func() {
		_, _ = s.StartScheduled()
	})
}

func (s *databaseBackupService) createFile(fileName, directory string) (string, int64, string, error) {
	path, err := s.pathForFileNameInDirectory(fileName, directory)
	if err != nil {
		return "", 0, "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", 0, "", err
	}
	tmp, err := os.CreateTemp(dir, ".bbs-go-backup-*.tmp")
	if err != nil {
		return "", 0, "", err
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", 0, "", err
	}
	_ = os.Chmod(tmpPath, 0o600)
	defer func() { _ = os.Remove(tmpPath) }()

	if err := s.dump(tmpPath); err != nil {
		return "", 0, "", err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return "", 0, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, "", err
	}
	checksum, err := fileChecksum(path)
	if err != nil {
		return "", 0, "", err
	}
	return path, info.Size(), checksum, nil
}

func (s *databaseBackupService) dump(path string) error {
	dbType := s.databaseType()
	switch dbType {
	case config.DbTypeSQLite:
		return s.dumpSQLite(path)
	case config.DbTypeMySQL:
		return s.dumpCommand(path, "mysqldump")
	case config.DbTypePostgreSQL:
		return s.dumpCommand(path, "pg_dump")
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func (s *databaseBackupService) dumpSQLite(path string) error {
	db := sqls.DB()
	if db == nil {
		return errors.New("database is not initialized")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	quoted := strings.ReplaceAll(path, "'", "''")
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err = sqlDB.Exec("VACUUM INTO '" + quoted + "'")
	return err
}

func (s *databaseBackupService) dumpCommand(path, commandName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	args, env, err := s.dumpCommandArgs(commandName, path)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Env = append(os.Environ(), env...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return errors.New(message)
	}
	return nil
}

func (s *databaseBackupService) dumpCommandArgs(commandName, path string) ([]string, []string, error) {
	dbType := s.databaseType()
	if config.Instance == nil {
		return nil, nil, errors.New("database configuration is not initialized")
	}
	if dbType == config.DbTypeMySQL && commandName == "mysqldump" {
		cfg, err := mysql.ParseDSN(config.Instance.DB.Url)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid MySQL connection string: %w", err)
		}
		if cfg.DBName == "" {
			return nil, nil, errors.New("MySQL database name is required")
		}
		args := []string{"--single-transaction", "--routines", "--events", "--triggers", "--hex-blob", "--no-tablespaces", "--result-file=" + path, "--user=" + cfg.User}
		args = append(args, "--ignore-table="+cfg.DBName+"."+backupMetadataTable, "--ignore-table="+cfg.DBName+"."+restoreMetadataTable, cfg.DBName)
		if cfg.Net == "unix" {
			args = append(args, "--socket="+cfg.Addr)
		} else {
			host, port := splitMySQLAddress(cfg.Addr)
			args = append(args, "--host="+host)
			if port != "" {
				args = append(args, "--port="+port)
			}
		}
		return args, []string{"MYSQL_PWD=" + cfg.Passwd}, nil
	}
	if dbType == config.DbTypePostgreSQL && commandName == "pg_dump" {
		cfg, err := pgx.ParseConfig(config.Instance.DB.Url)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid PostgreSQL connection string: %w", err)
		}
		args := []string{"--format=plain", "--no-owner", "--file=" + path, "--host=" + cfg.Host, "--port=" + fmt.Sprint(cfg.Port), "--username=" + cfg.User, "--dbname=" + cfg.Database}
		if sslMode := cfg.RuntimeParams["sslmode"]; sslMode != "" {
			args = append(args, "--sslmode="+sslMode)
		}
		return args, []string{"PGPASSWORD=" + cfg.Password}, nil
	}
	return nil, nil, errors.New("database dump command does not match configured database")
}

func splitMySQLAddress(address string) (string, string) {
	if strings.HasPrefix(address, "/") {
		return address, ""
	}
	if host, port, err := net.SplitHostPort(address); err == nil {
		return host, port
	}
	host, port, ok := strings.Cut(address, ":")
	if ok && host != "" && port != "" {
		return host, port
	}
	return address, ""
}

func (s *databaseBackupService) verifyFile(path, dbType, checksum string) error {
	actual, err := fileChecksum(path)
	if err != nil {
		return err
	}
	if actual != checksum {
		return errors.New("backup checksum verification failed")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return errors.New("backup file is empty or invalid")
	}
	if dbType == config.DbTypeSQLite {
		return verifySQLite(path)
	}
	if dbType == config.DbTypeMySQL {
		return verifyMySQLDump(path)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var header [256]byte
	count, err := file.Read(header[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if count == 0 {
		return errors.New("backup file is empty")
	}
	return nil
}

func verifyMySQLDump(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, 64*1024))
	if err != nil {
		return err
	}
	if len(content) == 0 || !utf8.Valid(content) || strings.ContainsRune(string(content), '\x00') {
		return errors.New("MySQL backup is not a valid text dump")
	}
	upper := strings.ToUpper(string(content))
	for _, marker := range []string{"MYSQL DUMP", "MARIADB DUMP", "CREATE TABLE", "INSERT INTO", "SET "} {
		if strings.Contains(upper, marker) {
			return nil
		}
	}
	return errors.New("MySQL backup does not contain a recognized SQL dump header or statement")
}

func verifySQLite(path string) error {
	db, err := gorm.Open(sqlite.Open("file:"+path+"?mode=ro"), &gorm.Config{})
	if err != nil {
		return err
	}
	var result string
	if err := db.Raw("PRAGMA integrity_check").Scan(&result).Error; err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(result)) != "ok" {
		return fmt.Errorf("sqlite integrity check failed: %s", result)
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
	return nil
}

func (s *databaseBackupService) prune() {
	retention := s.Config().Retention
	list := s.List(100)
	successful := make([]models.DatabaseBackup, 0)
	for _, backup := range list {
		if backup.Status == models.DatabaseBackupSuccess {
			successful = append(successful, backup)
		}
	}
	if len(successful) <= retention {
		return
	}
	for _, backup := range successful[retention:] {
		_ = s.Delete(backup.Id)
	}
}

func (s *databaseBackupService) directory(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || filepath.IsAbs(value) {
		return "", errors.New("backup directory must be a relative path")
	}
	clean := filepath.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("backup directory must stay inside the application directory")
	}
	root, err := filepath.Abs(config.GetConfigDir())
	if err != nil {
		return "", err
	}
	dir, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (s *databaseBackupService) pathForFileName(fileName string) (string, error) {
	return s.pathForFileNameInDirectory(fileName, s.Config().Directory)
}

func (s *databaseBackupService) pathForBackup(backup *models.DatabaseBackup) (string, error) {
	directory := backup.Directory
	if directory == "" {
		directory = s.Config().Directory
	}
	return s.pathForFileNameInDirectory(backup.FileName, directory)
}

func (s *databaseBackupService) pathForFileNameInDirectory(fileName, directory string) (string, error) {
	if fileName == "" || filepath.Base(fileName) != fileName {
		return "", errors.New("invalid backup file name")
	}
	dir, err := s.directory(directory)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

func (s *databaseBackupService) get(id int64) *models.DatabaseBackup {
	if db := sqls.DB(); db != nil {
		var backup models.DatabaseBackup
		if db.First(&backup, id).Error == nil {
			return &backup
		}
	}
	return nil
}

func (s *databaseBackupService) newFileName(trigger string) string {
	var suffix [4]byte
	_, _ = rand.Read(suffix[:])
	extension := ".sql"
	if s.databaseType() == config.DbTypeSQLite {
		extension = ".sqlite3"
	}
	return fmt.Sprintf("bbs-go-%s-%s-%s%s", time.Now().UTC().Format("20060102-150405"), trigger, hex.EncodeToString(suffix[:]), extension)
}

func (s *databaseBackupService) databaseType() string {
	if config.Instance == nil || strings.TrimSpace(config.Instance.DB.Type) == "" {
		return config.DbTypeMySQL
	}
	return strings.ToLower(config.Instance.DB.Type)
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func truncateBackupError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > maxBackupErrorLength {
		return message[:maxBackupErrorLength]
	}
	return message
}
