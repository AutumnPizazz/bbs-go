package api

import (
	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
)

// TopicClaim lets the current user claim a Q&A topic.
func TopicClaim(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	topicId := idcodec.Decode(ctx.Param("id"))
	if topicId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid topic id"))
		return
	}
	result, err := services.TopicAssignmentService.Claim(topicId, currentUser)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, result)
}

// TopicUnclaim removes a claim. Self or author/admin.
func TopicUnclaim(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	topicId := idcodec.Decode(ctx.Param("id"))
	if topicId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid topic id"))
		return
	}
	userId := params.FormValueInt64Default(ctx, "userId", currentUser.Id)
	if err := services.TopicAssignmentService.Unclaim(topicId, currentUser, userId); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

// TopicDismissClaim lets the topic author dismiss someone's claim.
func TopicDismissClaim(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	topicId := idcodec.Decode(ctx.Param("id"))
	if topicId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid topic id"))
		return
	}
	userId := params.FormValueInt64Default(ctx, "userId", 0)
	if userId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("userId is required"))
		return
	}
	if err := services.TopicAssignmentService.Dismiss(topicId, userId, currentUser); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

// TopicClaims returns all claims for a topic.
func TopicClaims(ctx *gin.Context) {
	topicId := idcodec.Decode(ctx.Param("id"))
	if topicId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid topic id"))
		return
	}
	list := services.TopicAssignmentService.ListClaims(topicId)
	ginx.WriteJSON(ctx, list)
}

// MyClaimedTopics returns topics claimed by the current user.
func MyClaimedTopics(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	status := params.FormValue(ctx, "status")
	limit := params.FormValueIntDefault(ctx, "limit", 20)
	list := services.TopicAssignmentService.ListClaimedTopics(currentUser.Id, status, limit)
	ginx.WriteJSON(ctx, list)
}

// MyClaimedCount returns the count of active claims.
func MyClaimedCount(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	count := services.TopicAssignmentService.CountActiveClaims(currentUser.Id)
	ginx.WriteJSON(ctx, map[string]interface{}{"count": count})
}
