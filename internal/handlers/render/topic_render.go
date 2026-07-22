package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/pkg/markdown"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/arrays"
)

func BuildTopic(ctx *gin.Context, topic *models.Topic) *resp.TopicResponse {
	rsp := _buildTopic(topic, true)
	if rsp == nil {
		return nil
	}

	if currentUser := common.GetCurrentUser(ctx); currentUser != nil {
		rsp.Liked = services.UserLikeService.Exists(currentUser.Id, constants.EntityTopic, topic.Id)
		rsp.Favorited = services.FavoriteService.IsFavorited(currentUser.Id, constants.EntityTopic, topic.Id)
	}

	if vote := services.VoteService.Get(topic.VoteId); vote != nil {
		rsp.Vote = BuildVote(ctx, vote)
	}

	// 附件仅在帖子详情接口返回。
	list := services.AttachmentService.ListByTopicId(topic.Id)
	if len(list) > 0 {
		rsp.Attachments = BuildAttachmentResponses(list)
	}

	return rsp
}

// BuildAttachmentResponses 将附件列表转为 AttachmentResponse 列表。
func BuildAttachmentResponses(list []models.Attachment) []resp.AttachmentResponse {
	if len(list) == 0 {
		return nil
	}
	atts := make([]resp.AttachmentResponse, 0, len(list))
	for _, att := range list {
		atts = append(atts, resp.AttachmentResponse{
			Id:            att.Id,
			FileName:      att.FileName,
			FileSize:      att.FileSize,
			DownloadCount: att.DownloadCount,
		})
	}
	return atts
}

func BuildSimpleTopic(topic *models.Topic) *resp.TopicResponse {
	return _buildTopic(topic, false)
}

func BuildSimpleTopics(ctx *gin.Context, topics []models.Topic) []resp.TopicResponse {
	if len(topics) == 0 {
		return nil
	}

	var likedTopicIds []int64
	if currentUser := common.GetCurrentUser(ctx); currentUser != nil {
		var topicIds []int64
		for _, topic := range topics {
			topicIds = append(topicIds, topic.Id)
		}
		likedTopicIds = services.UserLikeService.IsLiked(currentUser.Id, constants.EntityTopic, topicIds)
	}

	var responses []resp.TopicResponse
	for _, topic := range topics {
		item := BuildSimpleTopic(&topic)
		item.Liked = arrays.Contains(topic.Id, likedTopicIds)
		if vote := services.VoteService.Get(topic.VoteId); vote != nil {
			item.Vote = BuildVote(ctx, vote)
		}
		responses = append(responses, *item)
	}
	return responses
}

func _buildTopic(topic *models.Topic, buildContent bool) *resp.TopicResponse {
	if topic == nil {
		return nil
	}

	rsp := &resp.TopicResponse{}

	rsp.Id = idcodec.Encode(topic.Id)
	rsp.Type = topic.Type
	rsp.QaStatus = topic.QaStatus
	rsp.AcceptedCommentId = topic.AcceptedCommentId
	rsp.SolvedAt = topic.SolvedAt
	rsp.Title = topic.Title
	rsp.User = BuildUserInfoDefaultIfNull(topic.UserId)
	rsp.LastCommentTime = topic.LastCommentTime
	rsp.CreateTime = topic.CreateTime
	rsp.UpdateTime = topic.UpdateTime
	rsp.ViewCount = topic.ViewCount
	rsp.CommentCount = topic.CommentCount
	rsp.LikeCount = topic.LikeCount
	rsp.Recommend = topic.Recommend
	rsp.RecommendTime = topic.RecommendTime
	rsp.Sticky = topic.Sticky
	rsp.StickyTime = topic.StickyTime
	rsp.Status = topic.Status
	rsp.IpLocation = topic.IpLocation

	// 构建内容
	if buildContent {
		contentHtml := topic.Content
		if topic.ContentType == constants.ContentTypeMarkdown {
			contentHtml = markdown.ToHTML(topic.Content)
		}
		rsp.Content, rsp.Toc = handleTopicHtmlContent(contentHtml)
	} else {
		contentHtml := topic.Content
		if topic.ContentType == constants.ContentTypeMarkdown {
			contentHtml = markdown.ToHTML(topic.Content)
		}
		rsp.Summary = common.GetSummary(constants.ContentTypeHtml, contentHtml)
	}

	if topic.CategoryId > 0 {
		category := services.CategoryService.Get(topic.CategoryId)
		rsp.Category = BuildCategory(category)
	}

	tags := services.TopicService.GetTopicTags(topic.Id)
	rsp.Tags = BuildTags(tags)

	return rsp
}
