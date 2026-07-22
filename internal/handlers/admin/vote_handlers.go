package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"
)

func VoteDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.VoteService.Get(id)
	if t == nil || !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(t.TopicId)) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func VoteList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	list, paging := services.VoteService.FindPageByCnd(params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "id",
		},
	).Where("topic_id IN (?)", services.ContentAccessService.AllowedTopicSubquery(user)).Desc("id"))
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func VoteCreate(ctx *gin.Context) {
	t := &models.Vote{}
	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(t.TopicId)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteService.Create(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func VoteUpdate(ctx *gin.Context) {
	id, _ := params.GetInt64(ctx, "id")
	t := services.VoteService.Get(id)
	if t == nil || !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(t.TopicId)) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(t.TopicId)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteService.Update(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func VoteRemove(ctx *gin.Context) {
	ids := params.GetInt64Arr(ctx, "ids")
	if len(ids) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("delete ids is empty"))
		return
	}
	for _, id := range ids {
		vote := services.VoteService.Get(id)
		if vote == nil || !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(vote.TopicId)) {
			ginx.WriteJSON(ctx, errs.ContentAccessDenied())
			return
		}
		services.VoteService.Delete(id)
	}
	ginx.WriteJSON(ctx, nil)

}
