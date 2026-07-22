package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func migrate_notification_types_defaults() error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		now := dates.NowTimestamp()
		existing := repositories.SysConfigRepository.GetByKey(tx, constants.SysConfigNotificationTypes)
		if existing != nil && existing.Value != "" {
			return nil
		}
		defaults := map[string]dto.NoticeTypeConfig{
			"topicComment":     {Site: true},
			"commentReply":     {Site: true},
			"topicLike":        {Site: true},
			"topicFavorite":    {Site: true},
			"topicRecommend":   {Site: true},
			"topicDelete":      {Site: true},
			"articleComment":   {Site: true},
			"qaAnswerAccepted": {Site: true},
		}
		value := jsons.ToJsonStr(defaults)
		if existing == nil {
			return repositories.SysConfigRepository.Create(tx, &models.SysConfig{
				Key:         constants.SysConfigNotificationTypes,
				Value:       value,
				Name:        "通知类型配置",
				Description: "各消息类型的站内信开关",
				CreateTime:  now,
				UpdateTime:  now,
			})
		}
		existing.Value = value
		existing.UpdateTime = now
		return repositories.SysConfigRepository.Update(tx, existing)
	})
}
