package services

import (
	"errors"
	"fmt"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
)

const (
	ClaimStatusActive    = "active"
	ClaimStatusResolved  = "resolved"
	ClaimStatusDismissed = "dismissed"
)

var TopicAssignmentService = &topicAssignmentService{}

type topicAssignmentService struct{}

// Claim lets a logged-in user volunteer to answer a Q&A topic.
// Any user can claim; the topic author or admin can also dismiss.
func (s *topicAssignmentService) Claim(topicId int64, user *models.User) (*ClaimResponse, error) {
	if user == nil {
		return nil, errors.New("login required")
	}
	topic := TopicService.Get(topicId)
	if topic == nil || topic.Type != constants.TopicTypeQA {
		return nil, errors.New("only Q&A topics can be claimed")
	}
	if topic.Status != constants.StatusOk {
		return nil, errors.New("this topic has been deleted")
	}

	db := sqls.DB()

	// Check existing
	var existing models.TopicAssignment
	err := db.Where("topic_id = ? AND user_id = ?", topicId, user.Id).First(&existing).Error
	if err == nil {
		if existing.Status == ClaimStatusDismissed {
			db.Model(&existing).Updates(map[string]interface{}{
				"status": ClaimStatusActive, "update_time": dates.NowTimestamp(),
			})
			existing.Status = ClaimStatusActive
		}
		return &ClaimResponse{
			Id: existing.Id, TopicId: existing.TopicId, UserId: existing.UserId,
			AssignedBy: existing.AssignedBy, Status: existing.Status,
			AssignedAt: existing.AssignedAt, ResolvedAt: existing.ResolvedAt,
			CreateTime: existing.CreateTime,
			Nickname: user.Nickname, Username: user.Username.String,
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check claim: %w", err)
	}

	now := dates.NowTimestamp()
	claim := &models.TopicAssignment{
		TopicId:    topicId,
		UserId:     user.Id,
		AssignedBy: user.Id, // self-claimed
		Status:     ClaimStatusActive,
		AssignedAt: now,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := db.Create(claim).Error; err != nil {
		return nil, err
	}
	return &ClaimResponse{
		Id: claim.Id, TopicId: claim.TopicId, UserId: claim.UserId,
		AssignedBy: claim.AssignedBy, Status: claim.Status,
		AssignedAt: claim.AssignedAt, ResolvedAt: claim.ResolvedAt,
		CreateTime: claim.CreateTime,
		Nickname: user.Nickname, Username: user.Username.String,
	}, nil
}

// Unclaim removes a claim. The claiming user can unclaim themselves,
// or the topic author/admin can remove any claim.
func (s *topicAssignmentService) Unclaim(topicId int64, operator *models.User, userId int64) error {
	if operator == nil {
		return errors.New("login required")
	}
	topic := TopicService.Get(topicId)
	if topic == nil {
		return errors.New("topic not found")
	}
	// Self-unclaim always allowed; author/admin can unclaim others
	if operator.Id != userId && topic.UserId != operator.Id && !operator.IsOwner() {
		return errors.New("you can only remove your own claim")
	}

	result := sqls.DB().Where("topic_id = ? AND user_id = ?", topicId, userId).Delete(&models.TopicAssignment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("claim not found")
	}
	return nil
}

// Dismiss marks a claim as dismissed (author rejects a claim).
func (s *topicAssignmentService) Dismiss(topicId, userId int64, operator *models.User) error {
	if operator == nil {
		return errors.New("login required")
	}
	topic := TopicService.Get(topicId)
	if topic == nil || (topic.UserId != operator.Id && !operator.IsOwner()) {
		return errors.New("only the topic author can dismiss a claim")
	}
	return s.updateStatus(topicId, userId, ClaimStatusDismissed)
}

// Resolve marks all active claims on a topic as resolved.
func (s *topicAssignmentService) Resolve(topicId int64) {
	now := dates.NowTimestamp()
	sqls.DB().Model(&models.TopicAssignment{}).
		Where("topic_id = ? AND status = ?", topicId, ClaimStatusActive).
		Updates(map[string]interface{}{
			"status": ClaimStatusResolved, "resolved_at": now, "update_time": now,
		})
}

// ListClaims returns all claims for a topic, enriched with user info.
func (s *topicAssignmentService) ListClaims(topicId int64) []ClaimResponse {
	var list []models.TopicAssignment
	sqls.DB().Where("topic_id = ?", topicId).Order("id asc").Find(&list)

	result := make([]ClaimResponse, 0, len(list))
	for _, c := range list {
		user := UserService.Get(c.UserId)
		nickname := ""
		username := ""
		if user != nil {
			nickname = user.Nickname
			username = user.Username.String
		}
		result = append(result, ClaimResponse{
			Id:         c.Id,
			TopicId:    c.TopicId,
			UserId:     c.UserId,
			AssignedBy: c.AssignedBy,
			Status:     c.Status,
			AssignedAt: c.AssignedAt,
			ResolvedAt: c.ResolvedAt,
			CreateTime: c.CreateTime,
			Nickname:   nickname,
			Username:   username,
		})
	}
	return result
}

type ClaimResponse struct {
	Id         int64  `json:"id"`
	TopicId    int64  `json:"topicId"`
	UserId     int64  `json:"userId"`
	AssignedBy int64  `json:"assignedBy"`
	Status     string `json:"status"`
	AssignedAt int64  `json:"assignedAt"`
	ResolvedAt int64  `json:"resolvedAt"`
	CreateTime int64  `json:"createTime"`
	Nickname   string `json:"nickname"`
	Username   string `json:"username"`
}

// ListClaimedTopics returns claims by a user.
func (s *topicAssignmentService) ListClaimedTopics(userId int64, status string, limit int) []models.TopicAssignment {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	var list []models.TopicAssignment
	query := sqls.DB().Where("user_id = ?", userId)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Order("assigned_at desc").Limit(limit).Find(&list)
	return list
}

// CountActiveClaims returns the count of active claims for a user.
func (s *topicAssignmentService) CountActiveClaims(userId int64) int64 {
	var count int64
	sqls.DB().Model(&models.TopicAssignment{}).
		Where("user_id = ? AND status = ?", userId, ClaimStatusActive).
		Count(&count)
	return count
}

func (s *topicAssignmentService) updateStatus(topicId, userId int64, status string) error {
	return sqls.DB().Model(&models.TopicAssignment{}).
		Where("topic_id = ? AND user_id = ?", topicId, userId).
		Updates(map[string]interface{}{
			"status": status, "update_time": dates.NowTimestamp(),
		}).Error
}
