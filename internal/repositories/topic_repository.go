package repositories

import (
	"bbs-go/internal/models"
)

var TopicRepository = newTopicRepository()

func newTopicRepository() *topicRepository {
	return &topicRepository{}
}

type topicRepository struct {
	BaseRepository[models.Topic]
}
