package admin

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/msg"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"

	"bbs-go/internal/models"
	"bbs-go/internal/services"
)

type adminMessageForm struct {
	UserId  int64  `form:"userId" json:"userId"`
	Title   string `form:"title" json:"title"`
	Content string `form:"content" json:"content"`
}

func MessageDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.MessageService.Get(id)
	if t == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func MessageList(ctx *gin.Context) {
	list, paging := services.MessageService.FindPageByParams(params.NewQueryParams(ctx).
		EqByReq("user_id").EqByReq("type").EqByReq("status").LikeByReq("title").PageByReq().Desc("id"))
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func MessageCreate(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	form := &adminMessageForm{}
	if err := ginx.Bind(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := validateAdminMessage(form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	recipient := services.UserService.Get(form.UserId)
	if recipient == nil || recipient.Status == constants.StatusDeleted {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("recipient not found"))
		return
	}
	t := &models.Message{
		FromId:     operator.Id,
		UserId:     form.UserId,
		Title:      strings.TrimSpace(form.Title),
		Content:    strings.TrimSpace(form.Content),
		Type:       int(msg.TypeAdminAnnouncement),
		Status:     msg.StatusUnread,
		CreateTime: dates.NowTimestamp(),
	}

	err = services.MessageService.Create(t)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, "message", t.Id, "发送站内消息", ctx.Request)
	ginx.WriteJSON(ctx, t)

}

func MessageUpdate(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	t := services.MessageService.Get(id)
	if t == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	form := &adminMessageForm{}
	if err := ginx.Bind(ctx, form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := validateAdminMessage(form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	err = services.MessageService.Updates(t.Id, map[string]interface{}{
		"title":   strings.TrimSpace(form.Title),
		"content": strings.TrimSpace(form.Content),
	})
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "message", t.Id, "编辑站内消息", ctx.Request)
	}
	t.Title = strings.TrimSpace(form.Title)
	t.Content = strings.TrimSpace(form.Content)
	ginx.WriteJSON(ctx, t)

}

func MessageRemove(ctx *gin.Context) {
	ids := params.FormValueInt64Array(ctx, "ids")
	if len(ids) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("message ids are required"))
		return
	}
	operator := common.GetCurrentUser(ctx)
	for _, id := range ids {
		if services.MessageService.Get(id) == nil {
			ginx.WriteJSON(ctx, ginx.ErrorMessage("message not found"))
			return
		}
		if err := services.MessageService.Updates(id, map[string]interface{}{"status": msg.StatusDeleted}); err != nil {
			ginx.WriteJSON(ctx, err)
			return
		}
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, "message", 0,
			"删除站内消息，数量："+strconv.Itoa(len(ids)), ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)
}

func validateAdminMessage(form *adminMessageForm) error {
	form.Title = strings.TrimSpace(form.Title)
	form.Content = strings.TrimSpace(form.Content)
	if form.UserId <= 0 {
		return ginx.ErrorMessage("recipient is required")
	}
	if form.Title == "" {
		return ginx.ErrorMessage("title is required")
	}
	if form.Content == "" {
		return ginx.ErrorMessage("content is required")
	}
	if len(form.Title) > 1024 {
		return ginx.ErrorMessage("title is too long")
	}
	return nil
}
