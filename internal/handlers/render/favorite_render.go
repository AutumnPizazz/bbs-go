package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/bbsurls"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"
)

func BuildFavorite(favorite *models.Favorite) *resp.FavoriteResponse {
	rsp := &resp.FavoriteResponse{}
	rsp.Id = favorite.Id
	rsp.EntityType = favorite.EntityType
	rsp.CreateTime = favorite.CreateTime

	topic := services.TopicService.Get(favorite.EntityId)
	if topic == nil || topic.Status != constants.StatusOk {
		rsp.Deleted = true
	} else {
		rsp.Url = bbsurls.TopicUrl(topic.Id)
		rsp.User = BuildUserInfoDefaultIfNull(topic.UserId)
		rsp.Title = topic.Title
	rsp.Content = common.GetSummary(topic.ContentType, topic.Content)
	}
	return rsp
}

func BuildFavorites(favorites []models.Favorite) []resp.FavoriteResponse {
	if len(favorites) == 0 {
		return nil
	}
	var responses []resp.FavoriteResponse
	for _, favorite := range favorites {
		responses = append(responses, *BuildFavorite(&favorite))
	}
	return responses
}
