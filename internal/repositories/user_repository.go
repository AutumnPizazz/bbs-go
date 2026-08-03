package repositories

import (
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var UserRepository = newUserRepository()

func newUserRepository() *userRepository {
	return &userRepository{}
}

type userRepository struct {
	BaseRepository[models.User]
}

func (r *userRepository) GetByUsername(db *gorm.DB, username string) *models.User {
	return r.Take(db, "username = ?", username)
}
