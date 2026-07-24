package admin

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"
)

type healthComponent struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func Health(ctx *gin.Context) {
	components := map[string]healthComponent{}
	status := "ok"

	database := healthComponent{Status: "ok"}
	if db := sqls.DB(); db == nil {
		database = healthComponent{Status: "error", Message: "database is not initialized"}
	} else if sqlDB, err := db.DB(); err != nil {
		database = healthComponent{Status: "error", Message: err.Error()}
	} else {
		pingContext, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		err := sqlDB.PingContext(pingContext)
		cancel()
		if err != nil {
			database = healthComponent{Status: "error", Message: err.Error()}
		}
	}
	components["database"] = database
	if database.Status != "ok" {
		status = "degraded"
	}

	uploadConfig := services.SysConfigService.GetUploadConfig()
	storage := healthComponent{Status: "ok", Message: string(uploadConfig.EnableUploadMethod)}
	if storage.Message == "" {
		storage.Message = "local"
	}
	if err := services.UploadService.CheckConnectivity(); err != nil {
		storage.Status = "error"
		storage.Message = err.Error()
	}
	components["storage"] = storage
	if storage.Status != "ok" {
		status = "degraded"
	}

	searchStatus := services.SearchReindexService.Status()
	sitemapStatus := services.SeoSitemapService.Status()
	if searchStatus.Error != "" || sitemapStatus.Error != "" {
		status = "degraded"
	}
	components["search"] = healthComponent{Status: componentStatus(searchStatus.Error), Message: searchStatus.Error}
	components["sitemap"] = healthComponent{Status: componentStatus(sitemapStatus.Error), Message: sitemapStatus.Error}

	ginx.WriteJSON(ctx, map[string]interface{}{
		"status":     status,
		"checkedAt":  dates.NowTimestamp(),
		"components": components,
	})
}

func TaskStatus(ctx *gin.Context) {
	search := services.SearchReindexService.Status()
	sitemap := services.SeoSitemapService.Status()
	attachmentCleanup := services.AttachmentCleanupTaskService.Status()
	ginx.WriteJSON(ctx, map[string]interface{}{
		"tasks": map[string]interface{}{
			"searchReindex":     search,
			"sitemap":           sitemap,
			"attachmentCleanup": attachmentCleanup,
		},
		"failed":       search.Error != "" || sitemap.Error != "" || attachmentCleanup.Error != "",
		"active":       search.Running || sitemap.Running || attachmentCleanup.Running,
		"checkedAt":    dates.NowTimestamp(),
		"retrySupport": map[string]bool{"searchReindex": true, "sitemap": true, "attachmentCleanup": false},
	})
}

func componentStatus(message string) string {
	if message != "" {
		return "error"
	}
	return "ok"
}
