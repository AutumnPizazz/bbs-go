package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var UserLikeRepository = newUserLikeRepository()

func newUserLikeRepository() *userLikeRepository {
	return &userLikeRepository{}
}

type userLikeRepository struct {
	BaseRepository[models.UserLike]
}

func (r *userLikeRepository) Exists(db *gorm.DB, userId int64, entityType string, entityId int64) bool {
	return r.FindOne(db, sqls.NewCnd().Eq("user_id", userId).Eq("entity_id", entityId).Eq("entity_type", entityType)) != nil
}
