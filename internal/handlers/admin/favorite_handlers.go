package admin

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"

	"bbs-go/internal/models"
	"bbs-go/internal/services"
)

func FavoriteDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.FavoriteService.Get(id)
	if t == nil || !canAccessFavorite(common.GetCurrentUser(ctx), t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func FavoriteList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	allowed := services.ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{ParamName: "userId", Op: params.Eq},
		params.QueryFilter{ParamName: "entityType", Op: params.Eq},
		params.QueryFilter{ParamName: "entityId", Op: params.Eq},
	).Where("entity_type = ? AND entity_id IN (SELECT id FROM t_topic WHERE category_id IN (?))", constants.EntityTopic, allowed).Desc("id")
	list, paging := services.FavoriteService.FindPageByCnd(cnd)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func FavoriteCreate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	t := &models.Favorite{}
	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	if !canAccessFavorite(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	err := services.FavoriteService.Create(t)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, "favorite", t.Id, "创建收藏记录", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func FavoriteUpdate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	t := services.FavoriteService.Get(id)
	if t == nil || !canAccessFavorite(operator, t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !canAccessFavorite(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	err = services.FavoriteService.Update(t)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "favorite", t.Id, "更新收藏记录", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func canAccessFavorite(user *models.User, favorite *models.Favorite) bool {
	return favorite != nil && favorite.EntityType == constants.EntityTopic &&
		services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(favorite.EntityId))
}
