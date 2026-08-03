package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var DictTypeService = newDictTypeService()

func newDictTypeService() *dictTypeService {
	return &dictTypeService{BaseService: newBaseService[models.DictType, crudRepository[models.DictType]](repositories.DictTypeRepository)}
}

type dictTypeService struct {
	BaseService[models.DictType, crudRepository[models.DictType]]
}

func (s *dictTypeService) Delete(id int64) {
	repositories.DictTypeRepository.Delete(sqls.DB(), id)
}

func (s *dictTypeService) GetByCode(code string) *models.DictType {
	return s.FindOne(sqls.NewCnd().Eq("code", code))
}
