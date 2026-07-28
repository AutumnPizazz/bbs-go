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
func (s *topicAssignmentService) Claim(topicId int64, user *models.User) (*models.TopicAssignment, error) {
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
			return &existing, nil
		}
		return &existing, nil
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
	return claim, nil
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

// ListClaims returns all claims for a topic.
func (s *topicAssignmentService) ListClaims(topicId int64) []models.TopicAssignment {
	var list []models.TopicAssignment
	sqls.DB().Where("topic_id = ?", topicId).Order("id asc").Find(&list)
	return list
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
