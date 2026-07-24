package admin

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/services"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/web"
)

func UserReportDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	t := services.UserReportService.Get(id)
	if t == nil || !userReportTargetAccessible(common.GetCurrentUser(ctx), t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, buildUserReportDetail(t))

}

func UserReportList(ctx *gin.Context) {
	allowed := services.ContentAccessService.GetAllowedCategoryIds(common.GetCurrentUser(ctx))
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "dataType",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "dataId",
			Op:        params.Eq,
		},
		params.QueryFilter{
			ParamName: "processStatus",
			Op:        params.Eq,
		},
	).Desc("id")
	cnd.Where("(data_type = ? AND data_id IN (SELECT id FROM t_topic WHERE category_id IN (?))) OR (data_type = ? AND data_id IN (SELECT id FROM t_comment WHERE entity_type = ? AND entity_id IN (SELECT id FROM t_topic WHERE category_id IN (?)))) OR data_type = ?",
		constants.EntityTopic, allowed, constants.EntityComment, constants.EntityTopic, allowed, constants.EntityUser)
	list, paging := services.UserReportService.FindPageByCnd(cnd)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})
}

func UserReportCreate(ctx *gin.Context) {
	t := &models.UserReport{}
	err := ginx.Bind(ctx, t)
	if err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}

	err = services.UserReportService.Create(t)
	if err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	ginx.WriteJSON(ctx, t)

}

func UserReportProcess(ctx *gin.Context) {
	id, _ := params.GetInt64(ctx, "id")
	if id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id is required"))
		return
	}

	processStatus, _ := params.GetInt64(ctx, "processStatus")
	if processStatus != 1 && processStatus != 2 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("processStatus must be 1 or 2"))
		return
	}

	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}

	t := services.UserReportService.Get(id)
	if t == nil || !userReportTargetAccessible(user, t) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}
	scope, scopeErr := userReportProcessScope(ctx)
	if scopeErr != nil {
		ginx.WriteJSON(ctx, scopeErr)
		return
	}

	processed, err := processRelatedUserReports(t, processStatus, scope, user.Id)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(user.Id, constants.OpTypeUpdate, "userReport", t.Id, "更新举报处理状态", err, ctx.Request)
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, "userReport", t.Id,
		fmt.Sprintf("更新举报处理状态：%d，范围：%s，影响%d条", processStatus, scope, processed), ctx.Request)
	ginx.WriteJSON(ctx, t)

}

func UserReportAction(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	id, _ := params.GetInt64(ctx, "id")
	if id <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("id is required"))
		return
	}
	report := services.UserReportService.Get(id)
	if report == nil || !userReportTargetAccessible(user, report) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("report not found"))
		return
	}
	action, _ := params.Get(ctx, "action")
	if action == "" {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("action is required"))
		return
	}
	scope, err := userReportProcessScope(ctx)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(user.Id, constants.OpTypeUpdate, "userReport", report.Id, "举报处置动作："+action, err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}

	switch action {
	case "delete":
		switch report.DataType {
		case constants.EntityTopic:
			err = services.TopicService.Delete(report.DataId, user.Id, ctx.Request)
		case constants.EntityComment:
			err = services.CommentService.DeleteByAdmin(user, report.DataId, ctx.Request)
		default:
			err = ginx.ErrorMessage("delete action is not supported for this report")
		}
	case "restore":
		if report.DataType != constants.EntityTopic {
			err = ginx.ErrorMessage("restore action is only supported for topic reports")
		} else {
			err = services.TopicService.Undelete(report.DataId)
			if err == nil {
				services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityTopic, report.DataId, "从举报处置中恢复话题", ctx.Request)
			}
		}
	case "forbid":
		if report.DataType != constants.EntityUser {
			err = ginx.ErrorMessage("forbid action is only supported for user reports")
		} else {
			days, _ := params.GetInt(ctx, "days")
			if days == 0 {
				days = 7
			}
			if !services.PermissionService.CanForbiddenUser(user, days) {
				err = errs.NoPermission()
			} else {
				err = services.UserService.Forbidden(user.Id, report.DataId, days, report.Reason, ctx.Request)
			}
		}
	case "removeForbidden":
		if report.DataType != constants.EntityUser {
			err = ginx.ErrorMessage("removeForbidden action is only supported for user reports")
		} else {
			services.UserService.RemoveForbidden(user.Id, report.DataId, ctx.Request)
		}
	default:
		err = ginx.ErrorMessage("unsupported report action")
	}
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	processed, err := processRelatedUserReports(report, 1, scope, user.Id)
	if err != nil {
		services.OperateLogService.AddOperateLogFailure(user.Id, constants.OpTypeUpdate, "userReport", report.Id, "举报处置动作："+action, err, ctx.Request)
		ginx.WriteJSON(ctx, err)
		return
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, "userReport", report.Id,
		fmt.Sprintf("举报处置动作：%s，范围：%s，影响%d条", action, scope, processed), ctx.Request)
	ginx.WriteJSON(ctx, services.UserReportService.Get(report.Id))
}

func userReportProcessScope(ctx *gin.Context) (string, error) {
	scope, _ := params.Get(ctx, "scope")
	if scope == "" {
		return "current", nil
	}
	if scope != "current" && scope != "object" {
		return "", ginx.ErrorMessage("scope must be current or object")
	}
	return scope, nil
}

func processRelatedUserReports(report *models.UserReport, processStatus int64, scope string, userId int64) (int, error) {
	reports := []*models.UserReport{report}
	if scope == "object" {
		relatedReports := services.UserReportService.FindByObject(report.DataType, report.DataId)
		for i := range relatedReports {
			related := &relatedReports[i]
			if related.Id != report.Id && related.ProcessStatus == 0 {
				reports = append(reports, related)
			}
		}
	}
	processed := 0
	for _, item := range reports {
		if err := services.UserReportService.Updates(item.Id, map[string]interface{}{
			"process_status": processStatus, "process_time": dates.NowTimestamp(), "process_user_id": userId,
		}); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func userReportTargetAccessible(user *models.User, report *models.UserReport) bool {
	if user == nil || report == nil {
		return false
	}
	switch report.DataType {
	case constants.EntityTopic:
		return services.ContentAccessService.CanAccessTopicCategory(user, services.TopicService.Get(report.DataId))
	case constants.EntityComment:
		return canAccessCommentScope(user, services.CommentService.Get(report.DataId))
	case constants.EntityUser:
		return true
	default:
		return false
	}
}

func buildUserReportDetail(report *models.UserReport) map[string]interface{} {
	detail := web.NewRspBuilder(report).Build()
	detail["target"] = buildUserReportTarget(report)
	related := services.UserReportService.FindByObject(report.DataType, report.DataId)
	relatedReports := make([]map[string]interface{}, 0, len(related))
	pendingCount := 0
	for _, item := range related {
		if item.Id == report.Id {
			continue
		}
		if item.ProcessStatus == 0 {
			pendingCount++
		}
		relatedReports = append(relatedReports, web.NewRspBuilder(&item).Build())
	}
	detail["relatedReports"] = relatedReports
	detail["relatedReportCount"] = len(relatedReports)
	detail["pendingRelatedReportCount"] = pendingCount
	return detail
}

func buildUserReportTarget(report *models.UserReport) map[string]interface{} {
	if report == nil {
		return nil
	}

	target := map[string]interface{}{
		"type": report.DataType,
		"id":   report.DataId,
	}

	switch report.DataType {
	case "topic":
		if topic := services.TopicService.Get(report.DataId); topic != nil {
			target["title"] = topic.Title
			target["content"] = topic.Content
			target["contentType"] = topic.ContentType
			target["userId"] = topic.UserId
			target["status"] = topic.Status
			target["url"] = "/topic/" + idcodec.Encode(topic.Id)
			return target
		}
	case "comment":
		if comment := services.CommentService.Get(report.DataId); comment != nil {
			target["content"] = comment.Content
			target["contentType"] = comment.ContentType
			target["userId"] = comment.UserId
			target["entityType"] = comment.EntityType
			target["entityId"] = comment.EntityId
			target["quoteId"] = comment.QuoteId
			target["status"] = comment.Status
			if comment.EntityType == constants.EntityTopic {
				target["url"] = "/topic/" + idcodec.Encode(comment.EntityId)
			} else if parent := services.CommentService.Get(comment.EntityId); parent != nil && parent.EntityType == constants.EntityTopic {
				target["url"] = "/topic/" + idcodec.Encode(parent.EntityId)
			}
			return target
		}
	case "user":
		if user := services.UserService.Get(report.DataId); user != nil {
			target["title"] = user.Nickname
			target["username"] = user.Username.String
			target["nickname"] = user.Nickname
			target["description"] = user.Description
			target["avatar"] = user.Avatar
			target["status"] = user.Status
			target["url"] = "/user/" + idcodec.Encode(user.Id)
			return target
		}
	}

	target["missing"] = true
	return target
}

func UserReportUpdate(ctx *gin.Context) {
	id, err := params.FormValueInt64(ctx, "id")
	if err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	t := services.UserReportService.Get(id)
	if t == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("entity not found"))
		return
	}

	err = ginx.Bind(ctx, t)
	if err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}

	err = services.UserReportService.Update(t)
	if err != nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(err.Error()))
		return
	}
	ginx.WriteJSON(ctx, t)

}
