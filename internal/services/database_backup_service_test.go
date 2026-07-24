package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/pkg/config"
)

func TestDatabaseBackupValidateConfig(t *testing.T) {
	service := &databaseBackupService{}
	if err := service.ValidateConfig(config.BackupConfig{Retention: 7, Schedule: "0 3 * * *", Directory: "backups"}); err != nil {
		t.Fatalf("expected valid backup config: %v", err)
	}
	for _, invalid := range []config.BackupConfig{
		{Retention: -1, Schedule: "0 3 * * *", Directory: "backups"},
		{Retention: 7, Schedule: "not a cron", Directory: "backups"},
		{Retention: 7, Schedule: "0 3 * * *", Directory: "../outside"},
	} {
		if err := service.ValidateConfig(invalid); err == nil {
			t.Fatalf("expected invalid backup config to fail: %#v", invalid)
		}
	}
}

func TestDatabaseBackupFileNameAndDirectoryAreConstrained(t *testing.T) {
	service := &databaseBackupService{}
	fileName := service.newFileName(BackupTriggerManual)
	if filepath.Base(fileName) != fileName || !strings.HasPrefix(fileName, "bbs-go-") {
		t.Fatalf("unexpected generated backup file name: %q", fileName)
	}
	if _, err := service.pathForFileNameInDirectory(fileName, "../outside"); err == nil {
		t.Fatal("expected traversal backup directory to be rejected")
	}
	if _, err := service.pathForFileName("../backup.sql"); err == nil {
		t.Fatal("expected path traversal file name to be rejected")
	}
	if _, err := service.pathForFileName(filepath.Join("nested", "backup.sql")); err == nil {
		t.Fatal("expected nested file name to be rejected")
	}
}

func TestDatabaseBackupDumpArgsDoNotExposePasswords(t *testing.T) {
	original := config.Instance
	t.Cleanup(func() { config.Instance = original })
	config.Instance = &config.Config{DB: config.DBConfig{
		Type: config.DbTypeMySQL,
		Url:  "backup-user:secret-password@tcp(db.example:3307)/bbsgo?parseTime=true",
	}}
	service := &databaseBackupService{}
	args, env, err := service.dumpCommandArgs("mysqldump", "C:\\backups\\dump.sql")
	if err != nil {
		t.Fatalf("expected mysql dump args: %v", err)
	}
	for _, arg := range args {
		if strings.Contains(arg, "secret-password") {
			t.Fatalf("password leaked into dump args: %q", arg)
		}
	}
	if len(env) != 1 || env[0] != "MYSQL_PWD=secret-password" {
		t.Fatalf("expected password to be passed through MYSQL_PWD, got %#v", env)
	}
	host, port := splitMySQLAddress("db.example:3307")
	if host != "db.example" || port != "3307" {
		t.Fatalf("unexpected mysql host parsing: %q:%q", host, port)
	}
}

func TestDatabaseBackupSQLiteCreatesAndVerifiesSnapshot(t *testing.T) {
	previousDB := sqls.DB()
	previousConfig := config.Instance
	t.Cleanup(func() {
		sqls.SetDB(previousDB)
		config.Instance = previousConfig
	})

	sourcePath := filepath.Join(t.TempDir(), "source.db")
	db, err := gorm.Open(sqlite.Open(sourcePath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open source sqlite database: %v", err)
	}
	if err := db.Exec("CREATE TABLE backup_test_rows (id INTEGER PRIMARY KEY, value TEXT NOT NULL)").Error; err != nil {
		t.Fatalf("create source table: %v", err)
	}
	if err := db.Exec("INSERT INTO backup_test_rows (value) VALUES (?)", "snapshot").Error; err != nil {
		t.Fatalf("insert source row: %v", err)
	}
	sqls.SetDB(db)

	directory := "backup-test-" + strings.ReplaceAll(t.Name(), "/", "-")
	config.Instance = &config.Config{
		DB:     config.DBConfig{Type: config.DbTypeSQLite},
		Backup: config.BackupConfig{Directory: directory, Retention: 7, Schedule: config.DefaultBackupSchedule},
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(config.GetConfigDir(), directory)) })
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get source sqlite connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	service := &databaseBackupService{}
	path, size, checksum, err := service.createFile("snapshot.sqlite", directory)
	if err != nil {
		t.Fatalf("create sqlite backup: %v", err)
	}
	if size == 0 || checksum == "" {
		t.Fatalf("expected non-empty checksumed backup, size=%d checksum=%q", size, checksum)
	}
	if err := service.verifyFile(path, config.DbTypeSQLite, checksum); err != nil {
		t.Fatalf("verify sqlite backup: %v", err)
	}
	if err := db.AutoMigrate(&models.DatabaseBackup{}); err != nil {
		t.Fatalf("migrate backup table: %v", err)
	}
	stale := &models.DatabaseBackup{
		TriggerType: BackupTriggerManual, DatabaseType: config.DbTypeSQLite,
		FileName: "stale.sqlite", Directory: directory, Status: models.DatabaseBackupRunning,
		CreateTime: 1, UpdateTime: 1,
	}
	if err := db.Create(stale).Error; err != nil {
		t.Fatalf("create stale backup record: %v", err)
	}
	stalePath, err := service.pathForBackup(stale)
	if err != nil {
		t.Fatalf("resolve stale backup path: %v", err)
	}
	if err := os.WriteFile(stalePath, []byte("partial"), 0o600); err != nil {
		t.Fatalf("create stale backup file: %v", err)
	}
	service.RecoverStale()
	var recovered models.DatabaseBackup
	if err := db.First(&recovered, stale.Id).Error; err != nil {
		t.Fatalf("load recovered backup: %v", err)
	}
	if recovered.Status != models.DatabaseBackupFailed || recovered.Error == "" {
		t.Fatalf("expected stale backup to be marked failed, got %#v", recovered)
	}
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale backup file to be removed, stat err=%v", err)
	}
}
