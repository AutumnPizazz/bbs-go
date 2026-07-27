package repositories

import (
	"bbs-go/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var AdminPasswordChangeRepository = newAdminPasswordChangeRepository()

func newAdminPasswordChangeRepository() *adminPasswordChangeRepository {
	return &adminPasswordChangeRepository{}
}

type adminPasswordChangeRepository struct{}

func (r *adminPasswordChangeRepository) GetByUserId(db *gorm.DB, userId int64) *models.AdminPasswordChange {
	return r.get(db, userId, false)
}

func (r *adminPasswordChangeRepository) GetByUserIdForUpdate(db *gorm.DB, userId int64) *models.AdminPasswordChange {
	return r.get(db, userId, true)
}

func (r *adminPasswordChangeRepository) get(db *gorm.DB, userId int64, forUpdate bool) *models.AdminPasswordChange {
	ret := &models.AdminPasswordChange{}
	query := db.Where("user_id = ?", userId)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(ret).Error; err != nil {
		return nil
	}
	return ret
}

func (r *adminPasswordChangeRepository) Create(db *gorm.DB, change *models.AdminPasswordChange) error {
	return db.Create(change).Error
}

func (r *adminPasswordChangeRepository) Update(db *gorm.DB, change *models.AdminPasswordChange) error {
	return db.Save(change).Error
}

func (r *adminPasswordChangeRepository) Delete(db *gorm.DB, userId int64) error {
	return db.Where("user_id = ?", userId).Delete(&models.AdminPasswordChange{}).Error
}
