package services

import (
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
)

const maxAnnouncementLength = 100 * 1024

type announcementService struct{}

var AnnouncementService = &announcementService{}

func (s *announcementService) Publish(operatorId int64, content string) (*models.AnnouncementPublishRecord, error) {
	content = strings.TrimSpace(content)
	if len(content) > maxAnnouncementLength {
		return nil, errors.New("announcement is too long")
	}
	now := dates.NowTimestamp()
	previous := cache.SysConfigCache.GetStr(constants.SysConfigSiteNotification)
	record := &models.AnnouncementPublishRecord{
		OperatorId: operatorId, PreviousContent: previous, Content: content,
		Status: "published", PublishTime: now, CreateTime: now,
	}
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := SysConfigService.setSingle(tx, constants.SysConfigSiteNotification, content, "", ""); err != nil {
			return err
		}
		return tx.Create(record).Error
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *announcementService) FindPage(cnd *sqls.Cnd) ([]models.AnnouncementPublishRecord, *sqls.Paging) {
	if cnd == nil {
		cnd = sqls.NewCnd().Desc("id")
	}
	var records []models.AnnouncementPublishRecord
	cnd.Find(sqls.DB(), &records)
	return records, cnd.Paging
}
