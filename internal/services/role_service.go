package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var RoleService = newRoleService()

func newRoleService() *roleService {
	return &roleService{BaseService: newBaseService[models.Role, crudRepository[models.Role]](repositories.RoleRepository)}
}

type roleService struct {
	BaseService[models.Role, crudRepository[models.Role]]
}

func (s *roleService) Update(t *models.Role) error {
	err := repositories.RoleRepository.Update(sqls.DB(), t)
	if err == nil {
		PermissionService.ClearCache()
	}
	return err
}

func (s *roleService) Updates(id int64, columns map[string]interface{}) error {
	err := repositories.RoleRepository.Updates(sqls.DB(), id, columns)
	if err == nil {
		PermissionService.ClearCache()
	}
	return err
}

func (s *roleService) UpdateColumn(id int64, name string, value interface{}) error {
	err := repositories.RoleRepository.UpdateColumn(sqls.DB(), id, name, value)
	if err == nil {
		PermissionService.ClearCache()
	}
	return err
}

func (s *roleService) Delete(id int64) {
	repositories.RoleRepository.Delete(sqls.DB(), id)
}

func (s *roleService) GetByCode(code string) *models.Role {
	return s.FindOne(sqls.NewCnd().Eq("code", code))
}

func (s *roleService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *roleService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.RoleRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}
