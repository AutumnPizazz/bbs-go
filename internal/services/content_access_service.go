package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/msg"
	"bbs-go/internal/repositories"
	"sync"

	"github.com/mlogclub/simple/sqls"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

// ContentAccessService is the single source of truth for node-scoped content access.
var ContentAccessService = newContentAccessService()

type contentAccessService struct {
	mu    sync.RWMutex
	cache map[int64][]int64
}

func newContentAccessService() *contentAccessService {
	return &contentAccessService{cache: make(map[int64][]int64)}
}

// GetAllowedCategoryIds returns active categories visible to user. An empty result means no content access.
func (s *contentAccessService) GetAllowedCategoryIds(user *models.User) []int64 {
	if user == nil || user.Id <= 0 {
		return nil
	}

	s.mu.RLock()
	if ids, ok := s.cache[user.Id]; ok {
		ret := append([]int64(nil), ids...)
		s.mu.RUnlock()
		return ret
	}
	s.mu.RUnlock()

	active := CategoryService.GetCategories()
	if user.IsOwner() || user.ContentAccessMode == constants.ContentAccessModeAll {
		ids := make([]int64, 0, len(active))
		for _, category := range active {
			ids = append(ids, category.Id)
		}
		return s.store(user.Id, ids)
	}

	assigned := repositories.UserCategoryAccessRepository.FindByUserId(sqls.DB(), user.Id)
	roots := make(map[int64]struct{}, len(assigned))
	for _, access := range assigned {
		roots[access.CategoryId] = struct{}{}
	}

	byId := make(map[int64]models.Category, len(active))
	for _, category := range active {
		byId[category.Id] = category
	}
	ids := make([]int64, 0)
	for _, category := range active {
		current := category.Id
		for current > 0 {
			if _, ok := roots[current]; ok {
				ids = append(ids, category.Id)
				break
			}
			parent, ok := byId[current]
			if !ok || parent.ParentId == current {
				break
			}
			current = parent.ParentId
		}
	}
	return s.store(user.Id, ids)
}

func (s *contentAccessService) store(userId int64, ids []int64) []int64 {
	ret := append([]int64(nil), ids...)
	s.mu.Lock()
	s.cache[userId] = ret
	s.mu.Unlock()
	return append([]int64(nil), ret...)
}

func (s *contentAccessService) InvalidateUser(userId int64) {
	s.mu.Lock()
	delete(s.cache, userId)
	s.mu.Unlock()
}

func (s *contentAccessService) InvalidateAll() {
	s.mu.Lock()
	s.cache = make(map[int64][]int64)
	s.mu.Unlock()
}

func (s *contentAccessService) GetAssignedCategoryIds(userId int64) []int64 {
	if userId <= 0 {
		return nil
	}
	accesses := repositories.UserCategoryAccessRepository.FindByUserId(sqls.DB(), userId)
	ids := make([]int64, 0, len(accesses))
	for _, access := range accesses {
		ids = append(ids, access.CategoryId)
	}
	return ids
}

func (s *contentAccessService) CanAccessCategory(user *models.User, categoryId int64) bool {
	if categoryId <= 0 || user == nil {
		return false
	}
	for _, id := range s.GetAllowedCategoryIds(user) {
		if id == categoryId {
			return true
		}
	}
	return false
}

func (s *contentAccessService) CanAccessTopic(user *models.User, topic *models.Topic) bool {
	return topic != nil && topic.Status == constants.StatusOk && s.CanAccessCategory(user, topic.CategoryId)
}

func (s *contentAccessService) CanAccessTopicCategory(user *models.User, topic *models.Topic) bool {
	return topic != nil && s.CanAccessCategory(user, topic.CategoryId)
}

func (s *contentAccessService) CanAccessComment(user *models.User, comment *models.Comment) bool {
	if comment == nil || comment.Status == constants.StatusDeleted {
		return false
	}
	if comment.EntityType == constants.EntityTopic {
		return s.CanAccessTopic(user, TopicService.Get(comment.EntityId))
	}
	if comment.EntityType == constants.EntityComment {
		return s.CanAccessComment(user, CommentService.Get(comment.EntityId))
	}
	return false
}

func (s *contentAccessService) CanAccessEntity(user *models.User, entityType string, entityId int64) bool {
	switch entityType {
	case constants.EntityTopic:
		return s.CanAccessTopic(user, TopicService.Get(entityId))
	case constants.EntityComment:
		return s.CanAccessComment(user, CommentService.Get(entityId))
	default:
		return false
	}
}

// FilterCategoryTree retains only enabled categories in the user's scope.
func (s *contentAccessService) FilterCategoryTree(user *models.User, categories []models.Category) []models.Category {
	allowed := make(map[int64]struct{})
	for _, id := range s.GetAllowedCategoryIds(user) {
		allowed[id] = struct{}{}
	}
	ret := make([]models.Category, 0, len(categories))
	for _, category := range categories {
		if category.Status != constants.StatusOk {
			continue
		}
		if _, ok := allowed[category.Id]; !ok {
			continue
		}
		if category.ParentId > 0 {
			if _, ok := allowed[category.ParentId]; !ok {
				category.ParentId = 0
			}
		}
		ret = append(ret, category)
	}
	return ret
}

func (s *contentAccessService) CanWriteCategory(user *models.User, categoryId int64) bool {
	return s.CanAccessCategory(user, categoryId)
}

func (s *contentAccessService) CanAccessMessage(user *models.User, message *models.Message) bool {
	if message == nil {
		return false
	}
	switch msg.Type(message.Type) {
	case msg.TypeTopicComment:
		return s.canAccessMessageEntity(user, message.ExtraData, "entityType", "entityId")
	case msg.TypeCommentReply:
		return s.canAccessMessageEntity(user, message.ExtraData, "rootEntityType", "rootEntityId")
	case msg.TypeTopicLike, msg.TypeTopicFavorite, msg.TypeTopicRecommend, msg.TypeTopicDelete, msg.TypeQaAnswerAccepted:
		topicId := gjson.Get(message.ExtraData, "topicId").Int()
		return topicId > 0 && s.CanAccessTopic(user, TopicService.Get(topicId))
	default:
		return true
	}
}

func (s *contentAccessService) canAccessMessageEntity(user *models.User, data, typePath, idPath string) bool {
	entityType := gjson.Get(data, typePath).String()
	entityId := gjson.Get(data, idPath).Int()
	if entityId <= 0 {
		return false
	}
	switch entityType {
	case constants.EntityTopic:
		return s.CanAccessTopic(user, TopicService.Get(entityId))
	case constants.EntityComment:
		return s.CanAccessComment(user, CommentService.Get(entityId))
	default:
		return false
	}
}

func (s *contentAccessService) FilterMessages(user *models.User, messages []models.Message) []models.Message {
	if len(messages) == 0 {
		return nil
	}
	ret := make([]models.Message, 0, len(messages))
	for i := range messages {
		if s.CanAccessMessage(user, &messages[i]) {
			ret = append(ret, messages[i])
		}
	}
	return ret
}

// AllowedTopicSubquery constrains topic_id queries to the operator's active category scope.
func (s *contentAccessService) AllowedTopicSubquery(user *models.User) *gorm.DB {
	allowed := s.GetAllowedCategoryIds(user)
	query := sqls.DB().Model(&models.Topic{}).Select("id")
	if len(allowed) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("category_id IN ?", allowed)
}
