package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

func migrate_remove_category_type() error {
	db := modelsDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := dropLegacyIndex(tx, "t_category", "idx_category_type"); err != nil {
			return err
		}
		return dropColumn(tx, "t_category", "type")
	})
}
