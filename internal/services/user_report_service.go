package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var UserReportService = newUserReportService()

func newUserReportService() *userReportService {
	return &userReportService{BaseService: newBaseService[models.UserReport, crudRepository[models.UserReport]](repositories.UserReportRepository)}
}

type userReportService struct {
	BaseService[models.UserReport, crudRepository[models.UserReport]]
}

func (s *userReportService) FindByObject(dataType string, dataId int64) []models.UserReport {
	return repositories.UserReportRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("data_type", dataType).
		Eq("data_id", dataId).
		Desc("id"))
}

func (s *userReportService) Delete(id int64) {
	repositories.UserReportRepository.Delete(sqls.DB(), id)
}
