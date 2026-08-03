package repositories

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var ThirdUserRepository = newThirdUserRepository()

func newThirdUserRepository() *thirdUserRepository {
	return &thirdUserRepository{}
}

type thirdUserRepository struct {
	BaseRepository[models.ThirdUser]
}

func (r *thirdUserRepository) GetByOpenId(db *gorm.DB, openId string, thirdType constants.ThirdType) *models.ThirdUser {
	return r.FindOne(db, sqls.NewCnd().Where("open_id = ? and third_type = ?", openId, thirdType))
}

func (r *thirdUserRepository) GetByUserId(db *gorm.DB, userId int64, thirdType constants.ThirdType) *models.ThirdUser {
	return r.FindOne(db, sqls.NewCnd().Where("user_id = ? and third_type = ?", userId, thirdType))
}
