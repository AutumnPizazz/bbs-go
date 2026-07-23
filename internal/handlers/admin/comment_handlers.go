package admin

import (
	"strconv"

	"bbs-go/internal/handlers/render"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

func CommentDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid comment id"))
		return
	}
	comment := services.CommentService.Get(id)
	if comment == nil || !canAccessCommentScope(common.GetCurrentUser(ctx), comment) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("comment not found"))
		return
	}
	ginx.WriteJSON(ctx, buildAdminComment(comment))
}

func CommentList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	allowed := services.ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}

	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{ParamName: "id", Op: params.Eq},
		params.QueryFilter{ParamName: "userId", Op: params.Eq},
		params.QueryFilter{ParamName: "entityType", Op: params.Eq},
		params.QueryFilter{ParamName: "entityId", Op: params.Eq},
		params.QueryFilter{ParamName: "status", Op: params.Eq},
		params.QueryFilter{ParamName: "content", Op: params.Like},
	).Desc("id")
	if from := params.QueryValue(ctx, "from"); from != "" {
		cnd.Gte("create_time", from)
	}
	if to := params.QueryValue(ctx, "to"); to != "" {
		cnd.Lte("create_time", to)
	}
	// A reply is visible when its root comment belongs to an allowed topic.
	cnd.Where("(entity_type = ? AND entity_id IN (SELECT id FROM t_topic WHERE category_id IN (?))) OR (entity_type = ? AND entity_id IN (SELECT id FROM t_comment WHERE entity_type = ? AND entity_id IN (SELECT id FROM t_topic WHERE category_id IN (?))))",
		constants.EntityTopic, allowed, constants.EntityComment, constants.EntityTopic, allowed)
	list, paging := services.CommentService.FindPageByCnd(cnd)
	results := make([]map[string]interface{}, 0, len(list))
	for index := range list {
		results = append(results, buildAdminComment(&list[index]))
	}
	ginx.WriteJSON(ctx, &web.PageResult{Results: results, Page: paging})
}

func CommentRemove(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		id, err = params.FormValueInt64(ctx, "id")
	}
	if err != nil || id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid comment id"))
		return
	}
	operator := common.GetCurrentUser(ctx)
	comment := services.CommentService.Get(id)
	if comment == nil || !canAccessCommentScope(operator, comment) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	if err := services.CommentService.DeleteByAdmin(operator, id, ctx.Request); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

func canAccessCommentScope(user *models.User, comment *models.Comment) bool {
	if user == nil || comment == nil {
		return false
	}
	for depth := 0; depth < 32 && comment != nil; depth++ {
		switch comment.EntityType {
		case constants.EntityTopic:
			return services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(comment.EntityId))
		case constants.EntityComment:
			comment = services.CommentService.Get(comment.EntityId)
		default:
			return false
		}
	}
	return false
}

func buildAdminComment(comment *models.Comment) map[string]interface{} {
	item := map[string]interface{}{
		"id":         comment.Id,
		"comment":    render.BuildComment(comment),
		"content":    comment.Content,
		"status":     comment.Status,
		"userId":     comment.UserId,
		"entityType": comment.EntityType,
		"entityId":   comment.EntityId,
		"quoteId":    comment.QuoteId,
		"createTime": comment.CreateTime,
	}
	if comment.EntityType == constants.EntityComment {
		item["parent"] = services.CommentService.Get(comment.EntityId)
	}
	if comment.EntityType == constants.EntityTopic {
		item["topic"] = services.TopicService.Get(comment.EntityId)
	} else if parent := services.CommentService.Get(comment.EntityId); parent != nil && parent.EntityType == constants.EntityTopic {
		item["topic"] = services.TopicService.Get(parent.EntityId)
	}
	return item
}
