package api

import (
	"bbs-go/internal/handlers/render"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/search"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/spf13/cast"
)

func SearchTopic(ctx *gin.Context) {
	var (
		cursor     = params.FormValueIntDefault(ctx, "cursor", 1)
		keyword    = params.FormValue(ctx, "keyword")
		categoryId = params.FormValueInt64Default(ctx, "categoryId", 0)
		timeRange  = params.FormValueIntDefault(ctx, "timeRange", 0)
		format     = params.FormValue(ctx, "format")
		limit      = 20
	)
	var categoryIds []int64
	allowedCategoryIds := services.ContentAccessService.GetAllowedCategoryIds(common.GetCurrentUser(ctx))
	if len(allowedCategoryIds) == 0 {
		ginx.WriteJSON(ctx, ginx.CursorData([]search.TopicDocument{}, cast.ToString(cursor+1), false))
		return
	}
	if categoryId > 0 {
		requested := services.CategoryService.GetCategoryIdsForList(categoryId)
		allowed := make(map[int64]struct{}, len(allowedCategoryIds))
		for _, id := range allowedCategoryIds {
			allowed[id] = struct{}{}
		}
		for _, id := range requested {
			if _, ok := allowed[id]; ok {
				categoryIds = append(categoryIds, id)
			}
		}
		if len(categoryIds) == 0 {
			ginx.WriteJSON(ctx, ginx.CursorData([]search.TopicDocument{}, cast.ToString(cursor+1), false))
			return
		}
	} else {
		categoryIds = allowedCategoryIds
	}
	list, _, err := search.SearchTopic(keyword, categoryId, categoryIds, timeRange, format, cursor, limit)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, ginx.CursorData(render.BuildSearchTopics(list), cast.ToString(cursor+1), len(list) >= limit))

}

func SearchUser(ctx *gin.Context) {
	var (
		cursor  = params.FormValueIntDefault(ctx, "cursor", 1)
		keyword = params.FormValue(ctx, "keyword")
		limit   = 20
	)
	list, _, err := search.SearchUser(keyword, cursor, limit)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, ginx.CursorData(render.BuildSearchUsers(list), cast.ToString(cursor+1), len(list) >= limit))
}
