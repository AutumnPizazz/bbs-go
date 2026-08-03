package repositories

import (
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var UserTokenRepository = newUserTokenRepository()

func newUserTokenRepository() *userTokenRepository {
	return &userTokenRepository{}
}

type userTokenRepository struct {
	BaseRepository[models.UserToken]
}

func (r *userTokenRepository) GetByToken(db *gorm.DB, token string) *models.UserToken {
	if len(token) == 0 {
		return nil
	}
	return r.Take(db, "token = ?", token)
}
