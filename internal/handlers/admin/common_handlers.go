package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"
	"time"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

type dashboardRecentItem struct {
	Id         int64  `json:"id"`
	Title      string `json:"title,omitempty"`
	Content    string `json:"content,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	CreateTime int64  `json:"createTime"`
}

func buildRecentTopicItems(topics []models.Topic) []dashboardRecentItem {
	items := make([]dashboardRecentItem, 0, len(topics))
	for _, topic := range topics {
		items = append(items, dashboardRecentItem{
			Id:         topic.Id,
			Title:      topic.Title,
			CreateTime: topic.CreateTime,
		})
	}
	return items
}

func buildRecentUserItems(users []models.User) []dashboardRecentItem {
	items := make([]dashboardRecentItem, 0, len(users))
	for _, user := range users {
		items = append(items, dashboardRecentItem{
			Id:         user.Id,
			Nickname:   user.Nickname,
			CreateTime: user.CreateTime,
		})
	}
	return items
}

func CommonOverview(ctx *gin.Context) {

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	db := sqls.DB()

	metrics := map[string]int64{
		"totalUsers":    repositories.UserRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk)),
		"totalTopics":   repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk)),
		"totalArticles": repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Eq("format", constants.TopicFormatArticle)),
		"todayUsers":    repositories.UserRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", todayStart)),
		"todayTopics":   repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", todayStart)),
	}

	pending := map[string]int64{
		"pendingTopics":   repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusReview)),
		"pendingArticles": repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusReview).Eq("format", constants.TopicFormatArticle)),
		"pendingReports":  repositories.UserReportRepository.Count(db, sqls.NewCnd().Eq("audit_status", 0)),
	}

	recentTopics := repositories.TopicRepository.Find(db, sqls.NewCnd().Eq("status", constants.StatusOk).Desc("id").Limit(5))
	recentUsers := repositories.UserRepository.Find(db, sqls.NewCnd().Eq("status", constants.StatusOk).Desc("id").Limit(5))

	ginx.WriteJSON(ctx, web.NewEmptyRspBuilder().
		Put("metrics", metrics).
		Put("pending", pending).
		Put("recent", map[string]interface{}{
			"topics": buildRecentTopicItems(recentTopics),
			"users":  buildRecentUserItems(recentUsers),
		}).
		JsonResult())

}
