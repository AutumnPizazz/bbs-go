package admin

import (
	"strings"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
)

type announcementForm struct {
	Content string `json:"content" form:"content"`
}

func AnnouncementPreview(ctx *gin.Context) {
	form := &announcementForm{}
	if err := ginx.Bind(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, map[string]interface{}{
		"content":        strings.TrimSpace(form.Content),
		"currentContent": services.SysConfigService.GetSiteNotification(),
	})
}

func AnnouncementPublish(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	form := &announcementForm{}
	if err := ginx.Bind(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	record, err := services.AnnouncementService.Publish(operator.Id, form.Content)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, "announcement", 0, "发布站点公告", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "announcement", record.Id, "发布站点公告", ctx.Request)
	ginx.WriteJSON(ctx, record)
}

func AnnouncementHistory(ctx *gin.Context) {
	cnd := params.NewPagedSqlCnd(ctx, params.QueryFilter{ParamName: "status", Op: params.Eq}).Desc("id")
	list, paging := services.AnnouncementService.FindPage(cnd)
	ginx.WriteJSON(ctx, map[string]interface{}{"results": list, "page": paging})
}
