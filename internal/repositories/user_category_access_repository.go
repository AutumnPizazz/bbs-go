package repositories

import (
	"bbs-go/internal/models"

	"gorm.io/gorm"
)

var UserCategoryAccessRepository = newUserCategoryAccessRepository()

type userCategoryAccessRepository struct{}

func newUserCategoryAccessRepository() *userCategoryAccessRepository {
	return &userCategoryAccessRepository{}
}

func (r *userCategoryAccessRepository) FindByUserId(db *gorm.DB, userId int64) []models.UserCategoryAccess {
	var list []models.UserCategoryAccess
	db.Where("user_id = ?", userId).Find(&list)
	return list
}

func (r *userCategoryAccessRepository) DeleteByUserId(db *gorm.DB, userId int64) error {
	return db.Where("user_id = ?", userId).Delete(&models.UserCategoryAccess{}).Error
}

func (r *userCategoryAccessRepository) Create(db *gorm.DB, access *models.UserCategoryAccess) error {
	return db.Create(access).Error
}
