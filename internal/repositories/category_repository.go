package repositories

import (
	"bbs-go/internal/models"
)

var CategoryRepository = newCategoryRepository()

func newCategoryRepository() *categoryRepository {
	return &categoryRepository{}
}

type categoryRepository struct {
	BaseRepository[models.Category]
}
