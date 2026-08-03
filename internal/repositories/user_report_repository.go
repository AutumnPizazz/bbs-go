package repositories

import (
	"bbs-go/internal/models"
)

var UserReportRepository = newUserReportRepository()

func newUserReportRepository() *userReportRepository {
	return &userReportRepository{}
}

type userReportRepository struct {
	BaseRepository[models.UserReport]
}
