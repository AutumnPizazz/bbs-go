package repositories

import (
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"
)

var AttachmentRepository = newAttachmentRepository()

func newAttachmentRepository() *attachmentRepository {
	return &attachmentRepository{}
}

type attachmentRepository struct {
	BaseRepository[models.Attachment]
}

func (r *attachmentRepository) UpdateColumns(db *gorm.DB, topicId int64, columns map[string]interface{}) error {
	return db.Model(&models.Attachment{}).Where("topic_id = ?", topicId).Updates(columns).Error
}

func (r *attachmentRepository) ListByTopicId(db *gorm.DB, topicId int64) []models.Attachment {
	return r.Find(db, sqls.NewCnd().Eq("topic_id", topicId).Eq("status", constants.StatusOk).Asc("create_time"))
}

func (r *attachmentRepository) IncrDownloadCount(db *gorm.DB, id string) error {
	return r.UpdateColumn(db, id, "download_count", gorm.Expr("download_count + 1"))
}
