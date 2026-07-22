package api

import (
	"bbs-go/internal/handlers/render"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
)

func VoteDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	voteId := id

	vote := services.VoteService.Get(voteId)
	if vote == nil || !services.ContentAccessService.CanAccessTopic(common.GetCurrentUser(ctx), services.TopicService.Get(vote.TopicId)) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("vote not found"))
		return
	}
	ginx.WriteJSON(ctx, render.BuildVote(ctx, vote))

}

func VoteCast(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}

	var form req.VoteCastReq
	if err := ginx.BindJSON(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	vote := services.VoteService.Get(form.VoteId)
	if vote == nil || !services.ContentAccessService.CanAccessTopic(user, services.TopicService.Get(vote.TopicId)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}

	err := services.VoteService.Cast(user.Id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	vote = services.VoteService.Get(form.VoteId)
	ginx.WriteJSON(ctx, render.BuildVote(ctx, vote))

}
