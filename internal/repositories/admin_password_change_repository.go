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

type adminPasswordChangeRepository struct {
	BaseRepository[models.AdminPasswordChange]
}

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

// Delete 按 userId 删除（区别于 BaseRepository.Delete 按主键删除）
func (r *adminPasswordChangeRepository) Delete(db *gorm.DB, userId int64) error {
	return db.Where("user_id = ?", userId).Delete(&models.AdminPasswordChange{}).Error
}
