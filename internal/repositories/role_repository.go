package repositories

import (
	"bbs-go/internal/models"
)

var RoleRepository = newRoleRepository()

func newRoleRepository() *roleRepository {
	return &roleRepository{}
}

type roleRepository struct {
	BaseRepository[models.Role]
}
