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

func VoteRecordDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.VoteRecordService.Get(id)
	if t == nil || !canAccessVoteRecord(common.GetCurrentUser(ctx), t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func VoteRecordList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	allowed := services.ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}
	list, paging := services.VoteRecordService.FindPageByCnd(params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "id",
		},
		params.QueryFilter{ParamName: "userId", Op: params.Eq},
		params.QueryFilter{ParamName: "voteId", Op: params.Eq},
	).Where("vote_id IN (SELECT id FROM t_vote WHERE topic_id IN (?))", allowed).Desc("id"))
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})

}

func VoteRecordCreate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	t := &models.VoteRecord{}
	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !canAccessVoteRecord(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteRecordService.Create(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, "voteRecord", t.Id, "创建投票记录", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func VoteRecordUpdate(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	id, _ := params.GetInt64(ctx, "id")
	t := services.VoteRecordService.Get(id)
	if t == nil || !canAccessVoteRecord(operator, t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	if err := ginx.Bind(ctx, t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if !canAccessVoteRecord(operator, t) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	if err := services.VoteRecordService.Update(t); err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "voteRecord", t.Id, "更新投票记录", ctx.Request)
	}
	ginx.WriteJSON(ctx, t)

}

func VoteRecordRemove(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	ids := params.GetInt64Arr(ctx, "ids")
	if len(ids) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("delete ids is empty"))
		return
	}
	for _, id := range ids {
		if record := services.VoteRecordService.Get(id); record == nil || !canAccessVoteRecord(operator, record) {
			ginx.WriteJSON(ctx, errs.ContentAccessDenied())
			return
		}
		services.VoteRecordService.Delete(id)
		if operator != nil {
			services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, "voteRecord", id, "删除投票记录", ctx.Request)
		}
	}
	ginx.WriteJSON(ctx, nil)

}

func canAccessVoteRecord(user *models.User, record *models.VoteRecord) bool {
	if record == nil {
		return false
	}
	vote := services.VoteService.Get(record.VoteId)
	return vote != nil && services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(vote.TopicId))
}
