package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"
	"bbs-go/internal/services"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
		"totalUsers":     repositories.UserRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk)),
		"totalTopics":    repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk)),
		"todayUsers":     repositories.UserRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", todayStart)),
		"todayTopics":    repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", todayStart)),
		"todayComments":  repositories.CommentRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", todayStart)),
		"forbiddenUsers": repositories.UserRepository.Count(db, sqls.NewCnd().Where("forbidden_end_time = ? OR forbidden_end_time > ?", -1, todayStart)),
	}
	metrics["activeUsers"] = countActiveUsers(db, todayStart)
	searchStatus := services.SearchReindexService.Status()
	sitemapStatus := services.SeoSitemapService.Status()
	if searchStatus.Error != "" || sitemapStatus.Error != "" {
		metrics["failedTasks"] = 1
	} else {
		metrics["failedTasks"] = 0
	}

	pending := map[string]int64{
		"pendingReports": repositories.UserReportRepository.Count(db, sqls.NewCnd().Eq("process_status", 0)),
	}

	recentTopics := repositories.TopicRepository.Find(db, sqls.NewCnd().Eq("status", constants.StatusOk).Desc("id").Limit(5))
	recentUsers := repositories.UserRepository.Find(db, sqls.NewCnd().Eq("status", constants.StatusOk).Desc("id").Limit(5))
	trend := buildDashboardTrend(db, now)

	ginx.WriteJSON(ctx, web.NewEmptyRspBuilder().
		Put("metrics", metrics).
		Put("pending", pending).
		Put("trend", trend).
		Put("recent", map[string]interface{}{
			"topics": buildRecentTopicItems(recentTopics),
			"users":  buildRecentUserItems(recentUsers),
		}).
		JsonResult())

}

func countActiveUsers(db *gorm.DB, start int64) int64 {
	var count int64
	err := db.Raw(`SELECT COUNT(DISTINCT user_id) FROM (
		SELECT user_id FROM t_topic WHERE status = ? AND create_time >= ?
		UNION SELECT user_id FROM t_comment WHERE status = ? AND create_time >= ?
	) active_users`, constants.StatusOk, start, constants.StatusOk, start).Scan(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func buildDashboardTrend(db *gorm.DB, now time.Time) []map[string]interface{} {
	trend := make([]map[string]interface{}, 0, 7)
	for offset := 6; offset >= 0; offset-- {
		day := now.AddDate(0, 0, -offset)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		end := start.AddDate(0, 0, 1)
		startMs, endMs := start.UnixMilli(), end.UnixMilli()
		trend = append(trend, map[string]interface{}{
			"date":     fmt.Sprintf("%04d-%02d-%02d", start.Year(), start.Month(), start.Day()),
			"users":    repositories.UserRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", startMs).Lt("create_time", endMs)),
			"topics":   repositories.TopicRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", startMs).Lt("create_time", endMs)),
			"comments": repositories.CommentRepository.Count(db, sqls.NewCnd().Eq("status", constants.StatusOk).Gte("create_time", startMs).Lt("create_time", endMs)),
		})
	}
	return trend
}
