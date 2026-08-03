package repositories

import (
	"github.com/mlogclub/simple/common/dates"
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var TopicTagRepository = newTopicTagRepository()

func newTopicTagRepository() *topicTagRepository {
	return &topicTagRepository{}
}

type topicTagRepository struct {
	BaseRepository[models.TopicTag]
}

func (r *topicTagRepository) AddTopicTags(db *gorm.DB, topicId int64, tagIds []int64) error {
	if topicId <= 0 || len(tagIds) == 0 {
		return nil
	}
	for _, tagId := range tagIds {
		if err := r.Create(db, &models.TopicTag{
			TopicId:    topicId,
			TagId:      tagId,
			CreateTime: dates.NowTimestamp(),
		}); err != nil {
			return err
		}
	}
	return nil
}
