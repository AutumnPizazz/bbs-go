package api

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/sqls"

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

// MySummary returns aggregated stats and recent activity for the current user.
func MySummary(ctx *gin.Context) {
	currentUser, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	type ActivityItem struct {
		Type      string `json:"type"`
		TopicId   int64  `json:"topicId"`
		TopicTitle string `json:"topicTitle"`
		Time      int64  `json:"time"`
	}

	db := sqls.DB()
	userId := currentUser.Id

	// Count accepted answers (comments where user's comment was marked as accepted on a QA topic)
	var acceptedCount int64
	db.Raw(`SELECT COUNT(*) FROM t_comment c
		INNER JOIN t_topic t ON t.accepted_comment_id = c.id
		WHERE c.user_id = ? AND c.status = 0 AND t.status = 0`, userId).Scan(&acceptedCount)

	activeClaims := services.TopicAssignmentService.CountActiveClaims(userId)

	// Build recent activity by merging three sources in Go.
	var activities []ActivityItem

	// 1. Recent claims
	type claimRow struct {
		TopicId    int64
		AssignedAt int64
		Title      string
	}
	var claims []claimRow
	db.Raw(`SELECT a.topic_id, a.assigned_at, t.title
		FROM t_topic_assignment a
		INNER JOIN t_topic t ON t.id = a.topic_id
		WHERE a.user_id = ? AND t.status = 0
		ORDER BY a.assigned_at DESC LIMIT 30`, userId).Scan(&claims)
	for _, c := range claims {
		activities = append(activities, ActivityItem{
			Type: "claimed", TopicId: c.TopicId, TopicTitle: c.Title, Time: c.AssignedAt,
		})
	}

	// 2. Recent answers (comments on QA topics)
	type answerRow struct {
		TopicId    int64
		CreateTime int64
		Title      string
	}
	var answers []answerRow
	db.Raw(`SELECT c.entity_id AS topic_id, c.create_time, t.title
		FROM t_comment c
		INNER JOIN t_topic t ON t.id = c.entity_id
		WHERE c.user_id = ? AND c.entity_type = 'topic' AND c.status = 0 AND t.status = 0 AND t.type = 2
		ORDER BY c.create_time DESC LIMIT 30`, userId).Scan(&answers)
	for _, a := range answers {
		activities = append(activities, ActivityItem{
			Type: "answered", TopicId: a.TopicId, TopicTitle: a.Title, Time: a.CreateTime,
		})
	}

	// 3. Accepted answers
	type acceptedRow struct {
		TopicId  int64
		SolvedAt int64
		Title    string
	}
	var accepted []acceptedRow
	db.Raw(`SELECT t.id AS topic_id, t.solved_at, t.title
		FROM t_comment c
		INNER JOIN t_topic t ON t.accepted_comment_id = c.id
		WHERE c.user_id = ? AND c.status = 0 AND t.status = 0
		ORDER BY t.solved_at DESC LIMIT 30`, userId).Scan(&accepted)
	for _, a := range accepted {
		activities = append(activities, ActivityItem{
			Type: "accepted", TopicId: a.TopicId, TopicTitle: a.Title, Time: a.SolvedAt,
		})
	}

	// 4. Recent topics created
	type topicRow struct {
		Id         int64
		CreateTime int64
		Title      string
	}
	var topics []topicRow
	db.Raw(`SELECT id, create_time, title FROM t_topic
		WHERE user_id = ? AND status = 0
		ORDER BY create_time DESC LIMIT 30`, userId).Scan(&topics)
	for _, t := range topics {
		activities = append(activities, ActivityItem{
			Type: "topic_created", TopicId: t.Id, TopicTitle: t.Title, Time: t.CreateTime,
		})
	}

	// Sort merged activities by time descending, take top 30.
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Time > activities[j].Time
	})
	if len(activities) > 30 {
		activities = activities[:30]
	}

	ginx.WriteJSON(ctx, map[string]interface{}{
		"topicCount":    currentUser.TopicCount,
		"commentCount":  currentUser.CommentCount,
		"acceptedCount": acceptedCount,
		"activeClaims":  activeClaims,
		"recentActivity": activities,
	})
}
