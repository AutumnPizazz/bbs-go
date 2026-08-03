package repositories

import (
	"bbs-go/internal/models"
)

var FavoriteRepository = newFavoriteRepository()

func newFavoriteRepository() *favoriteRepository {
	return &favoriteRepository{}
}

type favoriteRepository struct {
	BaseRepository[models.Favorite]
}
