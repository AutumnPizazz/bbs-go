package repositories

import (
	"bbs-go/internal/models"
)

var MigrationRepository = newMigrationRepository()

func newMigrationRepository() *migrationRepository {
	return &migrationRepository{}
}

type migrationRepository struct {
	BaseRepository[models.Migration]
}
