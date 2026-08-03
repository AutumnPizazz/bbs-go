package repositories

import (
	"bbs-go/internal/models"
)

var PermissionRepository = newPermissionRepository()

func newPermissionRepository() *permissionRepository {
	return &permissionRepository{}
}

type permissionRepository struct {
	BaseRepository[models.Permission]
}
