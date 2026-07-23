package admin

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"

	"bbs-go/internal/handlers/render"
	"bbs-go/internal/models/constants"
	modelReq "bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/services"
)

// 推荐
// 取消推荐
func TopicDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.TopicService.Get(id)
	if t == nil || !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func TopicList(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "id",
			Op:        params.Eq,
			ValueWrapper: func(origin string) string {
				if id := idcodec.Decode(origin); id > 0 {
					return strconv.FormatInt(id, 10)
				}
				return ""
			},
		},
		params.QueryFilter{
			ParamName: "userId",
			Op:        params.Eq,
			ValueWrapper: func(origin string) string {
				if id := idcodec.Decode(origin); id > 0 {
					return strconv.FormatInt(id, 10)
				}
				return ""
			},
		},
		params.QueryFilter{
			ParamName: "status",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "type",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "qaStatus",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "format",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "recommend",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "title",
			Op:        params.Like,
		},
	).Desc("id")
	allowed := services.ContentAccessService.GetAllowedCategoryIds(user)
	if len(allowed) == 0 {
		cnd.Eq("id", -1)
	} else {
		cnd.In("category_id", allowed)
	}
	list, paging := services.TopicService.FindPageByCnd(cnd)

	var results []map[string]interface{}
	for _, topic := range list {
		item := render.BuildSimpleTopic(&topic)
		builder := web.NewRspBuilder(item)
		builder.Put("id", topic.Id) // Admin 端保持使用明文 ID，避免后台管理接口参数不兼容
		builder.Put("idEncode", idcodec.Encode(topic.Id))
		builder.Put("status", topic.Status)
		if vote := services.VoteService.Get(topic.VoteId); vote != nil {
			builder.Put("vote", render.BuildVote(ctx, vote))
		}
		results = append(results, builder.Build())
	}

	ginx.WriteJSON(ctx, &web.PageResult{Results: results, Page: paging})

}

func TopicRecommend(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	err = services.TopicService.SetRecommend(id, true)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, id, "设置话题推荐", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func TopicRemoveRecommend(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	err = services.TopicService.SetRecommend(id, false)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, id, "取消话题推荐", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func TopicSticky(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	user := common.GetCurrentUser(ctx)
	topic := services.TopicService.Get(id)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(user, topic) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	sticky := params.FormValueBoolDefault(ctx, "sticky", false)
	if err := services.TopicService.SetSticky(id, sticky); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityTopic, id,
		fmt.Sprintf("%s话题置顶", map[bool]string{true: "设置", false: "取消"}[sticky]), ctx.Request)
	ginx.WriteJSON(ctx, nil)
}

func TopicRemove(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	err = services.TopicService.Delete(id, user.Id, ctx.Request)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)

}

func TopicUndelete(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	err = services.TopicService.Undelete(id)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, id, "恢复话题", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func TopicAcceptAnswer(ctx *gin.Context) {
	var req modelReq.TopicAcceptAnswerReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if req.Id <= 0 || req.CommentId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id/commentId is required"))
		return
	}
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(req.Id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	if err := services.TopicService.AcceptAnswer(req.Id, req.CommentId, user.Id, true); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityTopic, req.Id,
		fmt.Sprintf("采纳答案评论：%d", req.CommentId), ctx.Request)
	ginx.WriteJSON(ctx, nil)

}

func TopicUnacceptAnswer(ctx *gin.Context) {
	id, _ := params.GetInt64(ctx, "id")
	if id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id is required"))
		return
	}
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	if err := services.TopicService.UnacceptAnswer(id, user.Id, true); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityTopic, id,
		"取消采纳答案", ctx.Request)
	ginx.WriteJSON(ctx, nil)

}

func TopicMarkSolved(ctx *gin.Context) {
	id, _ := params.GetInt64(ctx, "id")
	if id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id is required"))
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	if err := services.TopicService.ForceSetQaStatus(id, constants.QaStatusSolved); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, id, "标记话题已解决", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func TopicMarkUnsolved(ctx *gin.Context) {
	id, _ := params.GetInt64(ctx, "id")
	if id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id is required"))
		return
	}
	if !services.ContentAccessService.CanAccessTopicCategory(common.GetCurrentUser(ctx), services.TopicService.Get(id)) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	if err := services.TopicService.ForceSetQaStatus(id, constants.QaStatusUnsolved); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, id, "取消话题已解决状态", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}
