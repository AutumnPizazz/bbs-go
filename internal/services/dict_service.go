package services

import (
	"strconv"
	"strings"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"

	"bbs-go/internal/pkg/secureconfig"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var DictService = newDictService()

func newDictService() *dictService {
	return &dictService{BaseService: newBaseService[models.Dict, crudRepository[models.Dict]](repositories.DictRepository)}
}

type dictService struct {
	BaseService[models.Dict, crudRepository[models.Dict]]
}

func (s *dictService) Delete(id int64) {
	repositories.DictRepository.Delete(sqls.DB(), id)
}

func (s *dictService) ReferenceCount(id int64) int64 {
	count := s.Count(sqls.NewCnd().Eq("parent_id", id))
	needle := strconv.FormatInt(id, 10)
	for _, config := range SysConfigService.GetAll() {
		value, err := secureconfig.Decrypt(config.Value)
		if err != nil {
			value = config.Value
		}
		if strings.Contains(value, `"dictId":`+needle) || strings.Contains(value, `"dict_id":`+needle) {
			count++
		}
	}
	return count
}

func (s *dictService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *dictService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.DictRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *dictService) FindByTypeId(typeId int64) []models.Dict {
	return s.Find(sqls.NewCnd().Eq("type_id", typeId).Eq("status", constants.StatusOk).Asc("sort_no").Desc("id"))
}

func (s *dictService) GetBy(typeId int64, name string) *models.Dict {
	return s.FindOne(sqls.NewCnd().Where("type_id = ? and name = ?", typeId, name))
}
