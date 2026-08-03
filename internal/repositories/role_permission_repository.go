package repositories

import (
	"bbs-go/internal/models"
)

var RolePermissionRepository = newRolePermissionRepository()

func newRolePermissionRepository() *rolePermissionRepository {
	return &rolePermissionRepository{}
}

type rolePermissionRepository struct {
	BaseRepository[models.RolePermission]
}
