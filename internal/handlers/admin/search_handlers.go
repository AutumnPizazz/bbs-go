package admin

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
)

func SearchReindex(ctx *gin.Context) {
	status, started := services.SearchReindexService.Start()
	if started {
		if operator := common.GetCurrentUser(ctx); operator != nil {
			services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "searchReindex", 0, "启动搜索索引重建", ctx.Request)
		}
	}
	ginx.WriteJSON(ctx, status)
}

func SearchReindexStatus(ctx *gin.Context) {
	ginx.WriteJSON(ctx, services.SearchReindexService.Status())
}
