package services

import (
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var TopicTagService = newTopicTagService()

func newTopicTagService() *topicTagService {
	return &topicTagService{BaseService: newBaseService[models.TopicTag, crudRepository[models.TopicTag]](repositories.TopicTagRepository)}
}

type topicTagService struct {
	BaseService[models.TopicTag, crudRepository[models.TopicTag]]
}

func (s *topicTagService) HardDeleteTopicTags(ctx *sqls.TxContext, topicId int64) error {
	if topicId <= 0 {
		return nil
	}
	return ctx.Tx.Where("topic_id = ?", topicId).Delete(models.TopicTag{}).Error
}

func (s *topicTagService) DeleteByTopicId(ctx *sqls.TxContext, topicId int64) error {
	return ctx.Tx.Model(models.TopicTag{}).Where("topic_id = ?", topicId).UpdateColumn("status", constants.StatusDeleted).Error
}

func (s *topicTagService) UndeleteByTopicId(ctx *sqls.TxContext, topicId int64) error {
	return ctx.Tx.Model(models.TopicTag{}).Where("topic_id = ?", topicId).UpdateColumn("status", constants.StatusOk).Error
}
