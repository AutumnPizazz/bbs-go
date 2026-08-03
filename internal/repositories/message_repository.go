package repositories

import (
	"bbs-go/internal/models"
)

var MessageRepository = newMessageRepository()

func newMessageRepository() *messageRepository {
	return &messageRepository{}
}

type messageRepository struct {
	BaseRepository[models.Message]
}
