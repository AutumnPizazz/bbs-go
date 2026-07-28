package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
)

// TrashTopics lists soft-deleted topics scoped to the operator's accessible categories.
func TrashTopics(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{ParamName: "id"},
	).Eq("status", constants.StatusDeleted).Desc("id")

	allowedIds := services.ContentAccessService.GetAllowedCategoryIds(operator)
	if len(allowedIds) > 0 {
		cnd = cnd.In("category_id", allowedIds)
	}

	list, paging := services.TopicService.FindPageByCnd(cnd)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})
}

// TrashComments lists soft-deleted comments.
func TrashComments(ctx *gin.Context) {
	_, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{ParamName: "id"},
	).Eq("status", constants.StatusDeleted).Desc("id")

	list, paging := services.CommentService.FindPageByCnd(cnd)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})
}

// TrashAttachments lists soft-deleted attachments.
func TrashAttachments(ctx *gin.Context) {
	_, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	query := params.NewQueryParams(ctx).
		EqByReq("status").
		PageByReq().Desc("id")
	// Override status to only show deleted
	query.Cnd.Eq("status", constants.StatusDeleted)

	list, paging := services.AttachmentService.FindPageByParams(query)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})
}

// TrashRestoreComment restores a soft-deleted comment.
func TrashRestoreComment(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	id, err := strconv.ParseInt(params.FormValue(ctx, "id"), 10, 64)
	if err != nil || id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid comment id"))
		return
	}
	if err := services.CommentService.Undelete(id); err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, constants.EntityComment, id, "恢复评论", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityComment, id, "恢复评论", ctx.Request)
	ginx.WriteJSON(ctx, nil)
}

// TrashRestoreAttachment restores a soft-deleted attachment.
func TrashRestoreAttachment(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	id := params.FormValue(ctx, "id")
	if id == "" {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid attachment id"))
		return
	}
	if err := services.AttachmentService.Undelete(id); err != nil {
		services.OperateLogService.AddOperateLogFailure(operator.Id, constants.OpTypeUpdate, constants.EntityAttachment, 0, "恢复附件", err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityAttachment, 0, "恢复附件", ctx.Request)
	ginx.WriteJSON(ctx, nil)
}
