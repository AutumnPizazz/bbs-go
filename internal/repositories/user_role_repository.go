package repositories

import (
	"bbs-go/internal/models"
)

var UserRoleRepository = newUserRoleRepository()

func newUserRoleRepository() *userRoleRepository {
	return &userRoleRepository{}
}

type userRoleRepository struct {
	BaseRepository[models.UserRole]
}
