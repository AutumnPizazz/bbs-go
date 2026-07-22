package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/event"
	"bbs-go/internal/pkg/iplocator"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/search"
	"bbs-go/internal/repositories"
	"errors"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

var TopicPublishService = new(topicPublishService)

type topicPublishService struct{}

// Publish 发表
func (s *topicPublishService) Publish(userId int64, form req.CreateTopicReq) (*models.Topic, error) {
	if err := s.checkParams(userId, form); err != nil {
		return nil, err
	}
	// QA 话题不处理隐藏内容和投票，前端即使传入也忽略。
	if form.Type == constants.TopicTypeQA {
		form.HideContent = ""
		form.Vote = nil
	}

	now := dates.NowTimestamp()
	topic := &models.Topic{
		Type:            form.Type,
		QaStatus:        constants.QaStatusUnsolved,
		UserId:          userId,
		CategoryId:      form.CategoryId,
		Title:           form.Title,
		ContentType:     form.ContentType,
		Content:         form.Content,
		HideContent:     form.HideContent,
		Status:          constants.StatusOk,
		UserAgent:       form.UserAgent,
		Ip:              form.Ip,
		IpLocation:      iplocator.IpLocation(form.Ip),
		LastCommentTime: now,
		CreateTime:      now,
		UpdateTime:      now,
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		var (
			tagIds []int64
			err    error
		)
		// 帖子
		if err = repositories.TopicRepository.Create(ctx.Tx, topic); err != nil {
			return err
		}
		// 投票
		if form.Vote != nil {
			vote, voteErr := VoteService.CreateWithOptionsTx(ctx, topic.Id, userId, form.Vote, now)
			if voteErr != nil {
				return voteErr
			}
			if vote != nil {
				topic.VoteId = vote.Id
				if err = repositories.TopicRepository.UpdateColumn(ctx.Tx, topic.Id, "vote_id", vote.Id); err != nil {
					return err
				}
			}
		}

		// 标签
		if tagIds, err = repositories.TagRepository.GetOrCreates(ctx.Tx, form.Tags); err != nil {
			return err
		}
		if err = repositories.TopicTagRepository.AddTopicTags(ctx.Tx, topic.Id, tagIds); err != nil {
			return err
		}

		// 用户计数
		if err = UserService.IncrTopicCount(ctx, userId); err != nil {
			return err
		}

		// 附件绑定（同一事务内校验与更新，避免 SQLite 卡住）
		if len(form.AttachmentIds) > 0 {
			if err = AttachmentService.CheckAttachmentsExistAndOwned(ctx, userId, form.CategoryId, form.AttachmentIds, topic.Id); err != nil {
				return err
			}
			for _, aid := range form.AttachmentIds {
				if err = repositories.AttachmentRepository.UpdateColumn(ctx.Tx, aid, "topic_id", topic.Id); err != nil {
					return err
				}
				repositories.AttachmentRepository.UpdateColumn(ctx.Tx, aid, "update_time", now)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 添加索引
	search.UpdateTopicIndexAsync(topic)
	// 发送事件
	event.Send(event.TopicCreateEvent{
		UserId:     topic.UserId,
		TopicId:    topic.Id,
		TopicType:  int(topic.Type),
		CreateTime: topic.CreateTime,
	})
	return topic, nil
}

func (s topicPublishService) checkParams(userId int64, form req.CreateTopicReq) (err error) {
	modules := SysConfigService.GetModules()
	if form.Type == constants.TopicTypeTopic {
		if !modules.Topic {
			return errors.New(locales.Get("topic.discussions_disabled"))
		}
		if strs.IsBlank(form.Title) {
			return errors.New(locales.Get("topic.title_required"))
		}

		if strs.IsBlank(form.Content) {
			return errors.New(locales.Get("topic.content_required"))
		}

		if strs.RuneLen(form.Title) > 128 {
			return errors.New(locales.Get("topic.title_too_long"))
		}
	} else if form.Type == constants.TopicTypeQA {
		if !modules.QA {
			return errors.New(locales.Get("topic.qa_disabled"))
		}
		if strs.IsBlank(form.Title) {
			return errors.New(locales.Get("topic.title_required"))
		}

		if strs.IsBlank(form.Content) {
			return errors.New(locales.Get("topic.content_required"))
		}

		if strs.RuneLen(form.Title) > 128 {
			return errors.New(locales.Get("topic.title_too_long"))
		}
	} else {
		return errors.New(locales.Get("topic.type_not_supported"))
	}

	if form.CategoryId <= 0 {
		return errors.New(locales.Get("topic.category_required"))
	}
	if !ContentAccessService.CanWriteCategory(UserService.Get(userId), form.CategoryId) {
		return errs.ContentAccessDenied()
	}

	category := repositories.CategoryRepository.Get(sqls.DB(), form.CategoryId)
	if category == nil || category.Status != constants.StatusOk {
		return errors.New(locales.Get("topic.category_not_found"))
	}
	if !category.Type.Supports(form.Type) {
		return errors.New(locales.Get("topic.category_type_mismatch"))
	}
	if form.Type == constants.TopicTypeQA {
		form.Vote = nil
	}
	if err = VoteService.CheckCreateForm(form.Vote); err != nil {
		return err
	}

	return nil
}
