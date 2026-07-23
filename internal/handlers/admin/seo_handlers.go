package admin

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
)

func SeoSitemapGenerate(ctx *gin.Context) {
	status, started := services.SeoSitemapService.StartGenerate()
	if started {
		if operator := common.GetCurrentUser(ctx); operator != nil {
			services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "sitemap", 0, "启动 Sitemap 生成", ctx.Request)
		}
	}
	ginx.WriteJSON(ctx, status)
}

func SeoSitemapStatus(ctx *gin.Context) {
	ginx.WriteJSON(ctx, services.SeoSitemapService.Status())
}
