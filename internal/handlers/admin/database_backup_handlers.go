package admin

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
)

type databaseBackupConfigForm struct {
	Enabled   bool   `json:"enabled" form:"enabled"`
	Schedule  string `json:"schedule" form:"schedule"`
	Retention int    `json:"retention" form:"retention"`
	Directory string `json:"directory" form:"directory"`
	Confirm   string `json:"confirm" form:"confirm"`
}

func DatabaseBackupConfig(ctx *gin.Context) {
	ginx.WriteJSON(ctx, map[string]interface{}{"config": services.DatabaseBackupService.Config(), "status": services.DatabaseBackupService.Status()})
}

func DatabaseBackupSaveConfig(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	form := &databaseBackupConfigForm{}
	if err := ginx.Bind(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if strings.TrimSpace(form.Confirm) != services.BackupConfirmText {
		err := ginx.ErrorMessage("confirmation text does not match")
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackupConfig", 0, "保存数据库备份配置", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	next := config.BackupConfig{Enabled: form.Enabled, Schedule: form.Schedule, Retention: form.Retention, Directory: form.Directory}
	if err := services.DatabaseBackupService.SaveConfig(next); err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackupConfig", 0, "保存数据库备份配置", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "databaseBackupConfig", 0, "保存数据库备份配置", ctx.Request)
	ginx.WriteJSON(ctx, services.DatabaseBackupService.Config())
}

func DatabaseBackupList(ctx *gin.Context) {
	limit := params.FormValueIntDefault(ctx, "limit", 50)
	ginx.WriteJSON(ctx, map[string]interface{}{"results": services.DatabaseBackupService.List(limit), "status": services.DatabaseBackupService.Status()})
}

func DatabaseBackupCreate(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if params.FormValue(ctx, "confirm") != services.BackupConfirmText {
		err := ginx.ErrorMessage("confirmation text does not match")
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeCreate, "databaseBackup", 0, "创建数据库备份", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	backup, err := services.DatabaseBackupService.StartManual(operator.Id)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeCreate, "databaseBackup", 0, "创建数据库备份", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, "databaseBackup", backup.Id, "创建数据库备份", ctx.Request)
	ginx.WriteJSON(ctx, backup)
}

func DatabaseBackupRemove(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	id, err := strconv.ParseInt(params.FormValue(ctx, "id"), 10, 64)
	if err != nil || id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid backup id"))
		return
	}
	if params.FormValue(ctx, "confirm") != services.BackupDeleteConfirmText {
		err := ginx.ErrorMessage("confirmation text does not match")
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeDelete, "databaseBackup", id, "删除数据库备份", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.DatabaseBackupService.Delete(id); err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeDelete, "databaseBackup", id, "删除数据库备份", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, "databaseBackup", id, "删除数据库备份", ctx.Request)
	ginx.WriteJSON(ctx, nil)
}

func DatabaseBackupDownload(ctx *gin.Context) {
	operator, loginErr := common.CheckLogin(ctx)
	if loginErr != nil {
		ginx.WriteJSON(ctx, loginErr)
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 || ctx.Query("confirm") != services.BackupDownloadConfirmText {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackup", id, "下载数据库备份", ginx.ErrorMessage("backup download confirmation is required"), ctx.Request)
		ginx.WriteHttpStatusJSON(ctx, http.StatusBadRequest, ginx.ErrorMessage("backup download confirmation is required"))
		return
	}
	backup := services.DatabaseBackupService.Get(id)
	if backup == nil || backup.Status != models.DatabaseBackupSuccess {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackup", id, "下载数据库备份", ginx.ErrorMessage("backup not found"), ctx.Request)
		ginx.WriteHttpStatusJSON(ctx, http.StatusNotFound, ginx.ErrorMessage("backup not found"))
		return
	}
	path, err := services.DatabaseBackupService.FilePath(backup)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackup", id, "下载数据库备份", err, ctx.Request)
		ginx.WriteHttpStatusJSON(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Content-Disposition", `attachment; filename="`+backup.FileName+`"`)
	if _, err := os.Stat(path); err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "databaseBackup", id, "下载数据库备份", err, ctx.Request)
		ginx.WriteHttpStatusJSON(ctx, http.StatusNotFound, ginx.ErrorMessage("backup file not found"))
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "databaseBackup", id, "下载数据库备份", ctx.Request)
	ctx.File(path)
}
