package services

import (
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var LinkService = newLinkService()

func newLinkService() *linkService {
	return &linkService{BaseService: newBaseService[models.Link, crudRepository[models.Link]](repositories.LinkRepository)}
}

type linkService struct {
	BaseService[models.Link, crudRepository[models.Link]]
}

func (s *linkService) Delete(id int64) {
	repositories.LinkRepository.Delete(sqls.DB(), id)
}

func (s *linkService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *linkService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.LinkRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}
