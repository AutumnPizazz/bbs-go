package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/event"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/search"
	"errors"
	"math"
	"net/http"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"strings"

	"gorm.io/gorm"

	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var TopicService = newTopicService()

func newTopicService() *topicService {
	return &topicService{}
}

type topicService struct{}

func (s *topicService) Get(id int64) *models.Topic {
	return repositories.TopicRepository.Get(sqls.DB(), id)
}

func (s *topicService) Take(where ...interface{}) *models.Topic {
	return repositories.TopicRepository.Take(sqls.DB(), where...)
}

func (s *topicService) Find(cnd *sqls.Cnd) []models.Topic {
	return repositories.TopicRepository.Find(sqls.DB(), cnd)
}

func (s *topicService) FindOne(cnd *sqls.Cnd) *models.Topic {
	return repositories.TopicRepository.FindOne(sqls.DB(), cnd)
}

func (s *topicService) FindPageByParams(params *params.QueryParams) (list []models.Topic, paging *sqls.Paging) {
	return repositories.TopicRepository.FindPageByParams(sqls.DB(), params)
}

func (s *topicService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Topic, paging *sqls.Paging) {
	return repositories.TopicRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *topicService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TopicRepository.Count(sqls.DB(), cnd)
}

func (s *topicService) Updates(id int64, columns map[string]interface{}) error {
	if err := repositories.TopicRepository.Updates(sqls.DB(), id, columns); err != nil {
		return err
	}

	// 添加索引
	search.UpdateTopicIndex(s.Get(id))

	return nil
}

func (s *topicService) UpdateColumn(id int64, name string, value interface{}) error {
	if err := repositories.TopicRepository.UpdateColumn(sqls.DB(), id, name, value); err != nil {
		return err
	}

	// 添加索引
	search.UpdateTopicIndex(s.Get(id))

	return nil
}

// Delete 删除
func (s *topicService) Delete(topicId, deleteUserId int64, r *http.Request) error {
	topic := s.Get(topicId)
	if topic == nil {
		return nil
	}
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TopicRepository.UpdateColumn(ctx.Tx, topicId, "status", constants.StatusDeleted); err != nil {
			return err
		}
		// 话题标签软删除
		if err := TopicTagService.DeleteByTopicId(ctx, topicId); err != nil {
			return err
		}
		// 附件软删除（同一事务内执行，避免 SQLite 卡住）
		if err := AttachmentService.SoftDeleteByTopicId(ctx, topicId); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	search.DeleteTopicIndex(topicId)
	event.Send(event.TopicDeleteEvent{
		UserId:       topic.UserId,
		TopicId:      topic.Id,
		DeleteUserId: deleteUserId,
	})
	return nil
}

// Undelete 取消删除
func (s *topicService) Undelete(id int64) error {
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TopicRepository.UpdateColumn(ctx.Tx, id, "status", constants.StatusOk); err != nil {
			return err
		}
		if err := TopicTagService.UndeleteByTopicId(ctx, id); err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		search.UpdateTopicIndex(s.Get(id))
	}
	return err
}

// 更新
func (s *topicService) Edit(userId, topicId int64, form req.EditTopicReq) error {
	if len(form.Title) == 0 {
		return errors.New(locales.Get("topic.title_required"))
	}

	if strs.RuneLen(form.Title) > 128 {
		return errors.New(locales.Get("topic.title_too_long"))
	}

	category := repositories.CategoryRepository.Get(sqls.DB(), form.CategoryId)
	if category == nil || category.Status != constants.StatusOk {
		return errors.New(locales.Get("topic.category_not_found"))
	}
	topic := repositories.TopicRepository.Get(sqls.DB(), topicId)
	if topic == nil {
		return errors.New(locales.Get("common.not_found"))
	}
	viewer := UserService.Get(userId)
	if !ContentAccessService.CanAccessTopic(viewer, topic) || !ContentAccessService.CanWriteCategory(viewer, form.CategoryId) {
		return errs.ContentAccessDenied()
	}
	// 编辑时附件数量校验（仅帖子类型）
	if topic.Type == constants.TopicTypeTopic && form.AttachmentIds != nil {
		attCfg := SysConfigService.GetAttachmentConfig()
		if !attCfg.Enabled {
			return errors.New(locales.Get("attachment.disabled"))
		}
		if attCfg.MaxCount > 0 && len(form.AttachmentIds) > attCfg.MaxCount {
			return errors.New(locales.Getf("attachment.too_many", attCfg.MaxCount))
		}
	}
	if !category.Type.Supports(topic.Type) || (constants.IsArticleTopicFormat(topic.Format) && category.Type != constants.CategoryTypeNormal) {
		return errors.New(locales.Get("topic.category_type_mismatch"))
	}
	if constants.IsArticleTopicFormat(topic.Format) && len(form.AttachmentIds) > 0 {
		return errors.New(locales.Get("topic.type_not_supported"))
	}

	hideContent := form.HideContent
	if topic.Type == constants.TopicTypeQA || constants.IsArticleTopicFormat(topic.Format) {
		// QA 话题忽略隐藏内容变更。
		hideContent = ""
	}
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		var (
			tagIds []int64
			err    error
		)
		updates := map[string]interface{}{
			"category_id":  form.CategoryId,
			"title":        form.Title,
			"summary":      form.Summary,
			"content":      form.Content,
			"source_url":   strings.TrimSpace(form.SourceUrl),
			"hide_content": hideContent,
			"update_time":  dates.NowTimestamp(),
		}
		if form.Cover != nil {
			if strings.TrimSpace(form.Cover.Url) == "" {
				updates["cover"] = ""
			} else {
				updates["cover"] = jsons.ToJsonStr(form.Cover)
			}
		}
		if err = repositories.TopicRepository.Updates(ctx.Tx, topicId, updates); err != nil {
			return err
		}

		// 创建帖子对应标签
		if tagIds, err = repositories.TagRepository.GetOrCreates(ctx.Tx, form.Tags); err != nil {
			return err
		}

		// 先删掉所有的标签
		if err := TopicTagService.HardDeleteTopicTags(ctx, topicId); err != nil {
			return err
		}
		// 然后重新添加标签
		if err := repositories.TopicTagRepository.AddTopicTags(ctx.Tx, topicId, tagIds); err != nil {
			return err
		}

		// 附件全量替换（仅当请求中带 attachmentIds 时，同一事务内执行避免 SQLite 卡住）
		if form.AttachmentIds != nil {
			if err := AttachmentService.ReplaceTopicAttachments(ctx, topicId, userId, form.AttachmentIds); err != nil {
				return err
			}
		}
		return nil
	})

	// 添加索引
	search.UpdateTopicIndex(s.Get(topicId))

	event.Send(event.TopicUpdateEvent{
		UserId:  userId,
		TopicId: topicId,
	})

	return err
}

// 推荐
func (s *topicService) SetRecommend(topicId int64, recommend bool) error {
	topic := s.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return errors.New(locales.Get("topic.topic_not_found"))
	}
	if topic.Recommend == recommend { // 推荐状态没变更
		return nil
	}
	if recommend {
		if err := s.Updates(topicId, map[string]interface{}{
			"recommend":      recommend,
			"recommend_time": dates.NowTimestamp(),
		}); err != nil {
			return err
		}
	} else {
		if err := s.UpdateColumn(topicId, "recommend", recommend); err != nil {
			return err
		}
	}

	// 发送事件
	event.Send(event.TopicRecommendEvent{
		TopicId:   topicId,
		Recommend: recommend,
	})

	// 添加索引
	search.UpdateTopicIndex(s.Get(topicId))

	return nil
}

// GetTopicTags 话题的标签
func (s *topicService) GetTopicTags(topicId int64) []models.Tag {
	topicTags := repositories.TopicTagRepository.Find(sqls.DB(), sqls.NewCnd().Where("topic_id = ?", topicId))

	var tagIds []int64
	for _, topicTag := range topicTags {
		tagIds = append(tagIds, topicTag.TagId)
	}
	return cache.TagCache.GetList(tagIds)
}

// GetTopics 帖子列表（最新、推荐、关注、节点）
func (s *topicService) GetTopics(user *models.User, categoryId, cursor int64, qaStatus, sort string, format constants.TopicFormat) (topics []models.Topic, nextCursor int64, hasMore bool) {
	limit := constants.TopicListPageSize
	if categoryId == constants.CategoryIdFollow {
		if user != nil {
			return s._GetFollowTopics(user.Id, cursor, format)
		}
		return
	} else {
		return s._GetCategoryTopics(user, categoryId, cursor, limit, qaStatus, sort, format)
	}
}

// _GetCategoryTopics 帖子列表（最新、推荐、节点）
func (s *topicService) _GetCategoryTopics(user *models.User, categoryId, cursor int64, limit int, qaStatus, sort string, format constants.TopicFormat) (topics []models.Topic, nextCursor int64, hasMore bool) {
	cnd := sqls.NewCnd()
	allowedCategoryIds := ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowedCategoryIds) == 0 {
		cnd.Eq("id", -1)
	} else {
		cnd.In("category_id", allowedCategoryIds)
	}
	if format != "" {
		cnd.Eq("format", format)
	}
	if categoryId > 0 {
		categoryIds := CategoryService.GetCategoryIdsForList(categoryId)
		if len(categoryIds) > 0 {
			cnd.In("category_id", categoryIds)
		} else {
			cnd.Eq("category_id", categoryId)
		}
	}
	if categoryId == constants.CategoryIdRecommend {
		cnd.Eq("recommend", true)
	}
	if qaStatus != "" {
		cnd.Eq("type", constants.TopicTypeQA)
		cnd.Eq("qa_status", qaStatus)
	}
	if sort == "latestPublish" {
		if cursor > 0 {
			cnd.Lt("id", cursor)
		}
		cnd.Eq("status", constants.StatusOk).Desc("id").Limit(limit)
	} else {
		if cursor > 0 {
			cnd.Lt("last_comment_time", cursor)
		}
		cnd.Eq("status", constants.StatusOk).Desc("last_comment_time").Limit(limit)
	}
	topics = repositories.TopicRepository.Find(sqls.DB(), cnd)
	if len(topics) > 0 {
		if sort == "latestPublish" {
			nextCursor = topics[len(topics)-1].Id
		} else {
			nextCursor = topics[len(topics)-1].LastCommentTime
		}
		hasMore = len(topics) >= limit
	} else {
		nextCursor = cursor
	}
	return
}

// _GetFollowTopics 关注帖子列表
func (s *topicService) _GetFollowTopics(userId int64, cursor int64, format constants.TopicFormat) (topics []models.Topic, nextCursor int64, hasMore bool) {
	limit := constants.TopicListPageSize
	allowedCategoryIds := ContentAccessService.GetAllowedCategoryIds(UserService.Get(userId))
	if len(allowedCategoryIds) == 0 {
		return nil, cursor, false
	}
	query := sqls.DB().Table("t_user_feed AS f").
		Select("f.*").
		Joins("JOIN t_topic AS t ON t.id = f.data_id").
		Where("f.user_id = ? AND f.data_type = ? AND t.status = ? AND t.category_id IN ?", userId, constants.EntityTopic, constants.StatusOk, allowedCategoryIds)
	if format != "" {
		query = query.Where("t.format = ?", format)
	}
	if cursor > 0 {
		query = query.Where("f.create_time < ?", cursor)
	}
	query = query.Order("f.create_time DESC").Limit(limit)

	var userFeeds []models.UserFeed
	query.Find(&userFeeds)
	if len(userFeeds) > 0 {
		nextCursor = userFeeds[len(userFeeds)-1].CreateTime
		hasMore = len(userFeeds) >= limit
	} else {
		nextCursor = cursor
	}

	var topicIds []int64
	for _, item := range userFeeds {
		topicIds = append(topicIds, item.DataId)
	}
	topics = TopicService.GetTopicByIds(topicIds)
	viewer := UserService.Get(userId)
	topics = s.filterVisibleTopics(viewer, topics)
	return
}

// 指定标签下话题列表
func (s *topicService) GetTagTopics(user *models.User, tagId, cursor int64, format constants.TopicFormat) (topics []models.Topic, nextCursor int64, hasMore bool) {
	limit := constants.TopicListPageSize
	allowedCategoryIds := ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowedCategoryIds) == 0 {
		return nil, cursor, false
	}
	var topicTags []models.TopicTag
	query := sqls.DB().Table("t_topic_tag AS tt").
		Select("tt.*").
		Joins("JOIN t_topic AS t ON t.id = tt.topic_id").
		Where("tt.tag_id = ? AND tt.status = ? AND t.status = ? AND t.category_id IN ?", tagId, constants.StatusOk, constants.StatusOk, allowedCategoryIds).
		Order("tt.last_comment_time DESC").Limit(limit)
	if format != "" {
		query = query.Where("t.format = ?", format)
	}
	if cursor > 0 {
		query = query.Where("tt.last_comment_time < ?", cursor)
	}
	query.Find(&topicTags)
	if len(topicTags) > 0 {
		nextCursor = topicTags[len(topicTags)-1].LastCommentTime

		var topicIds []int64
		for _, topicTag := range topicTags {
			topicIds = append(topicIds, topicTag.TopicId)
		}

		topicsMap := s.GetTopicInIds(topicIds)
		if topicsMap != nil {
			for _, topicTag := range topicTags {
				if topic, found := topicsMap[topicTag.TopicId]; found {
					topics = append(topics, topic)
				}
			}
		}
	} else {
		nextCursor = cursor
	}
	hasMore = len(topicTags) >= limit
	return
}

func (s *topicService) GetTopicByIds(topicIds []int64) (topics []models.Topic) {
	topicsMap := s.GetTopicInIds(topicIds)
	for _, topicId := range topicIds {
		topic, found := topicsMap[topicId]
		if found {
			topics = append(topics, topic)
		}
	}
	return
}

// GetTopicInIds 根据编号批量获取主题
func (s *topicService) GetTopicInIds(topicIds []int64) map[int64]models.Topic {
	if len(topicIds) == 0 {
		return nil
	}
	var topics []models.Topic
	sqls.DB().Where("id in (?)", topicIds).Find(&topics)

	topicsMap := make(map[int64]models.Topic, len(topics))
	for _, topic := range topics {
		topicsMap[topic.Id] = topic
	}
	return topicsMap
}

// 浏览数+1
func (s *topicService) IncrViewCount(topicId int64) {
	sqls.DB().Exec("update t_topic set view_count = view_count + 1 where id = ?", topicId)
}

// 当帖子被评论的时候，更新最后回复时间、回复数量+1
func (s *topicService) onComment(tx *gorm.DB, topicId int64, comment *models.Comment) error {
	if err := repositories.TopicRepository.Updates(tx, topicId, map[string]interface{}{
		"last_comment_time":    comment.CreateTime,
		"last_comment_user_id": comment.UserId,
		"comment_count":        gorm.Expr("comment_count + 1"),
	}); err != nil {
		return err
	}
	if err := tx.Exec("update t_topic_tag set last_comment_time = ?, last_comment_user_id = ? where topic_id = ?",
		comment.CreateTime, comment.UserId, topicId).Error; err != nil {
		return err
	}
	return nil
}

func (s *topicService) ScanByUser(userId int64, callback func(topics []models.Topic)) {
	var cursor int64 = 0
	for {
		list := repositories.TopicRepository.Find(sqls.DB(), sqls.NewCnd().
			Eq("user_id", userId).Gt("id", cursor).Asc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

func (s *topicService) Scan(callback func(topics []models.Topic)) {
	var cursor int64 = 0
	for {
		list := repositories.TopicRepository.Find(sqls.DB(), sqls.NewCnd().
			Gt("id", cursor).Asc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

// 倒序扫描
func (s *topicService) ScanDesc(callback func(topics []models.Topic)) {
	var cursor int64 = math.MaxInt64
	for {
		list := repositories.TopicRepository.Find(sqls.DB(), sqls.NewCnd().
			Lt("id", cursor).Desc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

// 倒序扫描
func (s *topicService) ScanDescWithDate(dateFrom, dateTo int64, callback func(topics []models.Topic)) {
	var cursor int64 = math.MaxInt64
	for {
		list := repositories.TopicRepository.Find(sqls.DB(), sqls.NewCnd().
			Cols("id", "status", "create_time", "update_time").
			Lt("id", cursor).Gte("create_time", dateFrom).Lt("create_time", dateTo).Desc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

func (s *topicService) GetUserTopics(viewer *models.User, userId, cursor int64, format constants.TopicFormat) (topics []models.Topic, nextCursor int64, hasMore bool) {
	limit := constants.TopicListPageSize
	cnd := sqls.NewCnd()
	allowedCategoryIds := ContentAccessService.GetAllowedCategoryIds(viewer)
	if len(allowedCategoryIds) == 0 {
		cnd.Eq("id", -1)
	} else {
		cnd.In("category_id", allowedCategoryIds)
	}
	if userId > 0 {
		cnd.Eq("user_id", userId)
	}
	if cursor > 0 {
		cnd.Lt("id", cursor)
	}
	cnd.Eq("status", constants.StatusOk).Desc("id").Limit(limit)
	if format != "" {
		cnd.Eq("format", format)
	}
	topics = repositories.TopicRepository.Find(sqls.DB(), cnd)
	if len(topics) > 0 {
		nextCursor = topics[len(topics)-1].Id
		hasMore = len(topics) >= limit
	} else {
		nextCursor = cursor
	}
	return
}

func (s *topicService) GetStickyTopics(user *models.User, categoryId int64, limit int, qaStatus string, format constants.TopicFormat) []models.Topic {
	cnd := sqls.NewCnd().Eq("sticky", true).Eq("status", constants.StatusOk).Desc("sticky_time").Limit(limit)
	allowedCategoryIds := ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowedCategoryIds) == 0 {
		cnd.Eq("id", -1)
	} else {
		cnd.In("category_id", allowedCategoryIds)
	}
	if format != "" {
		cnd.Eq("format", format)
	}
	if categoryId > 0 {
		categoryIds := CategoryService.GetCategoryIdsForList(categoryId)
		if len(categoryIds) > 0 {
			cnd.In("category_id", categoryIds)
		} else {
			cnd.Eq("category_id", categoryId)
		}
	}
	if qaStatus != "" {
		cnd.Eq("type", constants.TopicTypeQA)
		cnd.Eq("qa_status", qaStatus)
	}
	return s.Find(cnd)
}

func (s *topicService) filterVisibleTopics(user *models.User, topics []models.Topic) []models.Topic {
	if len(topics) == 0 {
		return topics
	}
	ret := make([]models.Topic, 0, len(topics))
	for _, topic := range topics {
		if ContentAccessService.CanAccessTopic(user, &topic) {
			ret = append(ret, topic)
		}
	}
	return ret
}

func (s *topicService) SetSticky(topicId int64, sticky bool) error {
	topic := s.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return errors.New(locales.Get("topic.topic_not_found"))
	}
	if topic.Sticky == sticky {
		return nil
	}
	if sticky {
		return s.Updates(topicId, map[string]interface{}{
			"sticky":      true,
			"sticky_time": dates.NowTimestamp(),
		})
	} else {
		return s.Updates(topicId, map[string]interface{}{
			"sticky": false,
		})
	}
}

func (s *topicService) AcceptAnswer(topicId, commentId, userId int64, isAdmin bool) error {
	topic := s.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return errors.New(locales.Get("common.not_found"))
	}
	if topic.Type != constants.TopicTypeQA {
		return errors.New(locales.Get("topic.type_not_supported"))
	}
	if topic.UserId != userId && !isAdmin {
		return errors.New(locales.Get("topic.no_permission"))
	}

	comment := CommentService.Get(commentId)
	if comment == nil || comment.Status != constants.StatusOk ||
		comment.EntityType != constants.EntityTopic || comment.EntityId != topic.Id {
		return errors.New(locales.Get("common.not_found"))
	}

	now := dates.NowTimestamp()
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TopicRepository.Updates(ctx.Tx, topic.Id, map[string]interface{}{
			"accepted_comment_id": comment.Id,
			"qa_status":           constants.QaStatusSolved,
			"solved_at":           now,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	search.UpdateTopicIndex(s.Get(topic.Id))

	event.Send(event.QaAnswerAcceptedEvent{
		UserId:     comment.UserId,
		TopicId:    topic.Id,
		CommentId:  comment.Id,
		CreateTime: now,
	})
	return nil
}

func (s *topicService) UnacceptAnswer(topicId, userId int64, isAdmin bool) error {
	topic := s.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return errors.New(locales.Get("common.not_found"))
	}
	if topic.Type != constants.TopicTypeQA {
		return errors.New(locales.Get("topic.type_not_supported"))
	}
	if topic.UserId != userId && !isAdmin {
		return errors.New(locales.Get("topic.no_permission"))
	}

	return s.Updates(topic.Id, map[string]interface{}{
		"accepted_comment_id": 0,
		"qa_status":           constants.QaStatusUnsolved,
		"solved_at":           0,
	})
}

func (s *topicService) ForceSetQaStatus(topicId int64, qaStatus constants.QaStatus) error {
	topic := s.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return errors.New(locales.Get("common.not_found"))
	}
	if topic.Type != constants.TopicTypeQA {
		return errors.New(locales.Get("topic.type_not_supported"))
	}

	columns := map[string]interface{}{
		"qa_status": qaStatus,
	}
	if qaStatus == constants.QaStatusSolved {
		columns["solved_at"] = dates.NowTimestamp()
	} else {
		columns["solved_at"] = 0
		columns["accepted_comment_id"] = 0
	}
	return s.Updates(topic.Id, columns)
}
