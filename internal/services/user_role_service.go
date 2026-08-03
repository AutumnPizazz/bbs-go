package services

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var UserRoleService = newUserRoleService()

func newUserRoleService() *userRoleService {
	return &userRoleService{BaseService: newBaseService[models.UserRole, crudRepository[models.UserRole]](repositories.UserRoleRepository)}
}

type userRoleService struct {
	BaseService[models.UserRole, crudRepository[models.UserRole]]
}

func (s *userRoleService) IsRoleInUse(roleId int64) bool {
	if roleId <= 0 {
		return false
	}
	return s.Count(sqls.NewCnd().Eq("role_id", roleId)) > 0
}

func (s *userRoleService) Delete(id int64) {
	repositories.UserRoleRepository.Delete(sqls.DB(), id)
}

func (s *userRoleService) UpdateUserRoles(userId int64, roleIds []int64) error {
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		var roles []models.Role
		if len(roleIds) > 0 {
			roles = repositories.RoleRepository.Find(tx, sqls.NewCnd().In("id", roleIds))
		}

		var roleCodes []string
		for _, role := range roles {
			roleCodes = append(roleCodes, role.Code)
		}

		if err := tx.Delete(&models.UserRole{}, "user_id = ?", userId).Error; err != nil {
			return err
		}
		if len(roles) == 0 {
			return repositories.UserRepository.UpdateColumn(tx, userId, "roles", "")
		} else {
			for _, role := range roles {
				if err := repositories.UserRoleRepository.Create(tx, &models.UserRole{
					UserId:     userId,
					RoleId:     role.Id,
					CreateTime: dates.NowTimestamp(),
				}); err != nil {
					return err
				}
			}
			return repositories.UserRepository.UpdateColumn(tx, userId, "roles", strings.Join(roleCodes, ","))
		}
	})
	if err != nil {
		return err
	}
	cache.UserCache.Invalidate(userId)
	PermissionService.InvalidateUser(userId)
	ContentAccessService.InvalidateUser(userId)
	return nil
}

func (s *userRoleService) GetUserRoleIds(userId int64) (roleIds []int64) {
	list := s.Find(sqls.NewCnd().Eq("user_id", userId))
	for _, userRole := range list {
		roleIds = append(roleIds, userRole.RoleId)
	}
	return roleIds
}
