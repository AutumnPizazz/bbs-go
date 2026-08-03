package repositories

import (
	"bbs-go/internal/models"
)

var UserFeedRepository = newUserFeedRepository()

func newUserFeedRepository() *userFeedRepository {
	return &userFeedRepository{}
}

type userFeedRepository struct {
	BaseRepository[models.UserFeed]
}
