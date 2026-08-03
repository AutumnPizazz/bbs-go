package repositories

import (
	"bbs-go/internal/models"
)

var UserFollowRepository = newUserFollowRepository()

func newUserFollowRepository() *userFollowRepository {
	return &userFollowRepository{}
}

type userFollowRepository struct {
	BaseRepository[models.UserFollow]
}
