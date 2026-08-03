package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var UserFeedService = newUserFeedService()

func newUserFeedService() *userFeedService {
	return &userFeedService{BaseService: newBaseService[models.UserFeed, crudRepository[models.UserFeed]](repositories.UserFeedRepository)}
}

type userFeedService struct {
	BaseService[models.UserFeed, crudRepository[models.UserFeed]]
}

func (s *userFeedService) Delete(id int64) {
	repositories.UserFeedRepository.Delete(sqls.DB(), id)
}

func (s *userFeedService) DeleteByUser(userId, authorId int64) {
	sqls.DB().Where("user_id = ? and author_id = ?", userId, authorId).Delete(models.UserFeed{})
}

func (s *userFeedService) DeleteByDataId(dataId int64, dataType string) {
	sqls.DB().Where("data_id = ? and data_type = ?", dataId, dataType).Delete(models.UserFeed{})
}
