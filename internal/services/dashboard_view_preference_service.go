package services

import (
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

const maxDashboardViewJSONLength = 64 * 1024

func MaxDashboardViewJSONLength() int { return maxDashboardViewJSONLength }

type dashboardViewPreferenceService struct{}

var DashboardViewPreferenceService = &dashboardViewPreferenceService{}

func (s *dashboardViewPreferenceService) Get(userId int64, viewKey string) *models.DashboardViewPreference {
	var preference models.DashboardViewPreference
	if err := sqls.DB().Where("user_id = ? AND view_key = ?", userId, viewKey).First(&preference).Error; err != nil {
		return nil
	}
	return &preference
}

func (s *dashboardViewPreferenceService) Save(userId int64, viewKey, views string) error {
	if userId <= 0 || strings.TrimSpace(viewKey) == "" || len(views) > maxDashboardViewJSONLength {
		return errors.New("invalid dashboard view preference")
	}
	now := dates.NowTimestamp()
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		var preference models.DashboardViewPreference
		result := tx.Where("user_id = ? AND view_key = ?", userId, viewKey).First(&preference)
		if result.Error == nil {
			return tx.Model(&preference).Updates(map[string]interface{}{"views": views, "update_time": now}).Error
		}
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		return tx.Create(&models.DashboardViewPreference{UserId: userId, ViewKey: viewKey, Views: views, UpdateTime: now}).Error
	})
}
