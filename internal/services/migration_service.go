package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var MigrationService = newMigrationService()

func newMigrationService() *migrationService {
	return &migrationService{BaseService: newBaseService[models.Migration, crudRepository[models.Migration]](repositories.MigrationRepository)}
}

type migrationService struct {
	BaseService[models.Migration, crudRepository[models.Migration]]
}

func (s *migrationService) Delete(id int64) {
	repositories.MigrationRepository.Delete(sqls.DB(), id)
}

func (s *migrationService) GetBy(version string) *models.Migration {
	return s.FindOne(sqls.NewCnd().Where("version = ?", version))
}
