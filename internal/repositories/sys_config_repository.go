package repositories

import (
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var SysConfigRepository = newSysConfigRepository()

func newSysConfigRepository() *sysConfigRepository {
	return &sysConfigRepository{}
}

type sysConfigRepository struct {
	BaseRepository[models.SysConfig]
}

func (r *sysConfigRepository) GetByKey(db *gorm.DB, key string) *models.SysConfig {
	if len(key) == 0 {
		return nil
	}
	return r.Take(db, &models.SysConfig{Key: key})
}
