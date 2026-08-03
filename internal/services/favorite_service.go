package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/event"
	"bbs-go/internal/pkg/locales"
	"errors"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var FavoriteService = newFavoriteService()

func newFavoriteService() *favoriteService {
	return &favoriteService{BaseService: newBaseService[models.Favorite, crudRepository[models.Favorite]](repositories.FavoriteRepository)}
}

type favoriteService struct {
	BaseService[models.Favorite, crudRepository[models.Favorite]]
}

func (s *favoriteService) Delete(id int64) {
	repositories.FavoriteRepository.Delete(sqls.DB(), id)
}

func (s *favoriteService) IsFavorited(userId int64, entityType string, entityId int64) bool {
	return repositories.FavoriteRepository.Take(sqls.DB(), "user_id = ? and entity_type = ? and entity_id = ?",
		userId, entityType, entityId) != nil
}

func (s *favoriteService) GetBy(userId int64, entityType string, entityId int64) *models.Favorite {
	return repositories.FavoriteRepository.Take(sqls.DB(), "user_id = ? and entity_type = ? and entity_id = ?",
		userId, entityType, entityId)
}

// AddTopicFavorite 收藏主题
func (s *favoriteService) AddTopicFavorite(userId, topicId int64) error {
	topic := repositories.TopicRepository.Get(sqls.DB(), topicId)
	if topic == nil || !ContentAccessService.CanAccessTopic(UserService.Get(userId), topic) {
		return errors.New(locales.Get("favorite.topic_not_found"))
	}
	return s.addFavorite(userId, constants.EntityTopic, topicId)
}

func (s *favoriteService) addFavorite(userId int64, entityType string, entityId int64) error {
	if s.IsFavorited(userId, entityType, entityId) { // 已经收藏
		return nil
	}
	if err := repositories.FavoriteRepository.Create(sqls.DB(), &models.Favorite{
		UserId:     userId,
		EntityType: entityType,
		EntityId:   entityId,
		CreateTime: dates.NowTimestamp(),
	}); err != nil {
		return err
	}

	// 发送事件
	event.Send(event.UserFavoriteEvent{
		UserId:     userId,
		EntityId:   entityId,
		EntityType: entityType,
	})
	return nil
}
