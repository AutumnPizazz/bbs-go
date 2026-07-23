package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"
)

func VoteOptionDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.VoteOptionService.Get(id)
	if t == nil || !canAccessVoteOption(common.GetCurrentUser(ctx), t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func VoteOptionList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	allowed := services.ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}
	list, paging := services.VoteOptionService.FindPageByCnd(params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "id",
		},
		params.QueryFilter{ParamName: "voteId", Op: params.Eq},
	).Where("vote_id IN (SELECT id FROM t_vote WHERE topic_id IN (?))", allowed).Desc("id"))
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func VoteOptionCreate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	t := &models.VoteOption{}
	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !canAccessVoteOption(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteOptionService.Create(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, "voteOption", t.Id, "创建投票选项", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func VoteOptionUpdate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	id, _ := params.GetInt64(ctx, "id")
	t := services.VoteOptionService.Get(id)
	if t == nil || !canAccessVoteOption(operator, t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !canAccessVoteOption(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteOptionService.Update(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "voteOption", t.Id, "更新投票选项", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func VoteOptionRemove(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	ids := params.GetInt64Arr(ctx, "ids")
	if len(ids) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("delete ids is empty"))
		return
	}
	for _, id := range ids {
		if option := services.VoteOptionService.Get(id); option == nil || !canAccessVoteOption(operator, option) {
			ginx.WriteJSON(ctx, errs.ContentAccessDenied())
			return
		}
		services.VoteOptionService.Delete(id)
		if operator != nil {
			services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, "voteOption", id, "删除投票选项", ctx.Request)
		}
	}
	ginx.WriteJSON(ctx, nil)

}

func canAccessVoteOption(user *models.User, option *models.VoteOption) bool {
	if option == nil {
		return false
	}
	vote := services.VoteService.Get(option.VoteId)
	return vote != nil && services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(vote.TopicId))
}
