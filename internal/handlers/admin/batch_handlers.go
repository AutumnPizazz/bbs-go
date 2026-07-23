package admin

import (
	"fmt"
	"strings"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	modelReq "bbs-go/internal/models/req"
	"bbs-go/internal/permissions"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"
)

const adminBatchMaxSize = 100

type adminBatchPreview struct {
	Action          string  `json:"action"`
	RequestedCount  int     `json:"requestedCount"`
	EligibleCount   int     `json:"eligibleCount"`
	IneligibleCount int     `json:"ineligibleCount"`
	EligibleIds     []int64 `json:"eligibleIds"`
	ConfirmText     string  `json:"confirmText"`
}

type adminBatchFailure struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

type adminBatchResult struct {
	Action         string              `json:"action"`
	RequestedCount int                 `json:"requestedCount"`
	EligibleCount  int                 `json:"eligibleCount"`
	ProcessedCount int                 `json:"processedCount"`
	SkippedCount   int                 `json:"skippedCount"`
	FailedCount    int                 `json:"failedCount"`
	Failures       []adminBatchFailure `json:"failures"`
}

func bindAdminBatch(ctx *gin.Context) (modelReq.AdminBatchReq, []int64, error) {
	var req modelReq.AdminBatchReq
	if err := ginx.Bind(ctx, &req); err != nil {
		return req, nil, err
	}
	ids := req.ParsedIds()
	if len(ids) == 0 {
		return req, nil, ginx.ErrorMessage("ids is required")
	}
	if len(ids) > adminBatchMaxSize {
		return req, nil, ginx.ErrorMessage(fmt.Sprintf("at most %d ids can be processed at once", adminBatchMaxSize))
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return req, nil, ginx.ErrorMessage("ids must contain positive integers")
		}
		if _, ok := seen[id]; ok {
			return req, nil, ginx.ErrorMessage("ids must be unique")
		}
		seen[id] = struct{}{}
	}
	return req, ids, nil
}

func adminBatchConfirmation(action string, count int) string {
	return fmt.Sprintf("BATCH %s %d", strings.ToUpper(action), count)
}

func newAdminBatchPreview(action string, ids []int64, eligible []int64) adminBatchPreview {
	return adminBatchPreview{
		Action:          action,
		RequestedCount:  len(ids),
		EligibleCount:   len(eligible),
		IneligibleCount: len(ids) - len(eligible),
		EligibleIds:     eligible,
		ConfirmText:     adminBatchConfirmation(action, len(eligible)),
	}
}

func newAdminBatchResult(action string, ids []int64, eligibleCount int) adminBatchResult {
	return adminBatchResult{
		Action:         action,
		RequestedCount: len(ids),
		EligibleCount:  eligibleCount,
		SkippedCount:   len(ids) - eligibleCount,
		Failures:       make([]adminBatchFailure, 0),
	}
}

func appendAdminBatchFailure(result *adminBatchResult, id int64, err error) {
	result.FailedCount++
	result.Failures = append(result.Failures, adminBatchFailure{Id: id, Message: err.Error()})
}

func validateBatchConfirmation(req modelReq.AdminBatchReq, action string, eligibleCount int) error {
	expected := adminBatchConfirmation(action, eligibleCount)
	if strings.TrimSpace(req.ConfirmText) != expected {
		return ginx.ErrorMessage("confirmation text does not match the current eligible scope; request a new preview")
	}
	return nil
}

func canUseBatchPermission(user *models.User, permission permissions.PermissionDefinition) bool {
	return user != nil && (user.IsOwner() || services.PermissionService.HasPermission(user, permission.Code))
}

func topicBatchPermission(action string) (permissions.PermissionDefinition, error) {
	switch action {
	case "recommend":
		return permissions.PermissionTopicBatchRecommend, nil
	case "delete", "restore":
		return permissions.PermissionTopicBatchDelete, nil
	default:
		return permissions.PermissionDefinition{}, ginx.ErrorMessage("unsupported topic batch action")
	}
}

func topicBatchEligible(topic *models.Topic, action string) bool {
	if topic == nil {
		return false
	}
	switch action {
	case "recommend":
		return topic.Status == constants.StatusOk && !topic.Recommend
	case "delete":
		return topic.Status == constants.StatusOk
	case "restore":
		return topic.Status == constants.StatusDeleted
	default:
		return false
	}
}

func collectTopicBatch(user *models.User, ids []int64, action string) ([]*models.Topic, []int64) {
	eligible := make([]*models.Topic, 0, len(ids))
	eligibleIds := make([]int64, 0, len(ids))
	for _, id := range ids {
		topic := services.TopicService.Get(id)
		if !services.ContentAccessService.CanAccessTopicCategory(user, topic) || !topicBatchEligible(topic, action) {
			continue
		}
		eligible = append(eligible, topic)
		eligibleIds = append(eligibleIds, id)
	}
	return eligible, eligibleIds
}

func TopicBatchPreview(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	permission, err := topicBatchPermission(req.Action)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !canUseBatchPermission(operator, permission) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	_, eligibleIds := collectTopicBatch(operator, ids, req.Action)
	ginx.WriteJSON(ctx, newAdminBatchPreview(req.Action, ids, eligibleIds))
}

func TopicBatch(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	permission, err := topicBatchPermission(req.Action)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !canUseBatchPermission(operator, permission) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	eligible, _ := collectTopicBatch(operator, ids, req.Action)
	if err := validateBatchConfirmation(req, req.Action, len(eligible)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	result := newAdminBatchResult(req.Action, ids, len(eligible))
	for _, topic := range eligible {
		var operationErr error
		var operationType string
		var description string
		switch req.Action {
		case "recommend":
			operationErr = services.TopicService.SetRecommend(topic.Id, true)
			operationType = constants.OpTypeUpdate
			description = "批量设置话题推荐"
		case "delete":
			operationErr = services.TopicService.Delete(topic.Id, operator.Id, ctx.Request)
			operationType = constants.OpTypeDelete
			description = "批量删除话题"
		case "restore":
			operationErr = services.TopicService.Undelete(topic.Id)
			operationType = constants.OpTypeUpdate
			description = "批量恢复话题"
		}
		if operationErr != nil {
			appendAdminBatchFailure(&result, topic.Id, operationErr)
			continue
		}
		result.ProcessedCount++
		services.OperateLogService.AddOperateLog(operator.Id, operationType, constants.EntityTopic, topic.Id, description, ctx.Request)
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityTopic, 0,
		fmt.Sprintf("批量%s话题，成功%d，失败%d", req.Action, result.ProcessedCount, result.FailedCount), ctx.Request)
	ginx.WriteJSON(ctx, result)
}

func userBatchPermission(action string, days int) (permissions.PermissionDefinition, error) {
	switch action {
	case "forbid":
		if days == -1 {
			return permissions.PermissionUserBatchForbiddenForever, nil
		}
		if days > 0 {
			return permissions.PermissionUserBatchForbidden, nil
		}
		return permissions.PermissionDefinition{}, ginx.ErrorMessage("days must be a positive number or -1")
	case "removeForbidden":
		return permissions.PermissionUserBatchForbidden, nil
	default:
		return permissions.PermissionDefinition{}, ginx.ErrorMessage("unsupported user batch action")
	}
}

func canUseUserBatchPermission(user *models.User, action string, days int) bool {
	if action == "removeForbidden" {
		return canUseBatchPermission(user, permissions.PermissionUserBatchForbidden) ||
			canUseBatchPermission(user, permissions.PermissionUserBatchForbiddenForever)
	}
	permission, err := userBatchPermission(action, days)
	return err == nil && canUseBatchPermission(user, permission)
}

func userBatchEligible(user *models.User, action string) bool {
	if user == nil {
		return false
	}
	if action == "forbid" {
		return !user.IsForbidden()
	}
	return action == "removeForbidden" && user.IsForbidden()
}

func collectUserBatch(ids []int64, action string) ([]*models.User, []int64) {
	eligible := make([]*models.User, 0, len(ids))
	eligibleIds := make([]int64, 0, len(ids))
	for _, id := range ids {
		user := services.UserService.Get(id)
		if !userBatchEligible(user, action) {
			continue
		}
		eligible = append(eligible, user)
		eligibleIds = append(eligibleIds, id)
	}
	return eligible, eligibleIds
}

func UserBatchPreview(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	if _, validationErr := userBatchPermission(req.Action, req.Days); validationErr != nil {
		ginx.WriteJSON(ctx, validationErr)
		return
	}
	if !canUseUserBatchPermission(operator, req.Action, req.Days) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	_, eligibleIds := collectUserBatch(ids, req.Action)
	ginx.WriteJSON(ctx, newAdminBatchPreview(req.Action, ids, eligibleIds))
}

func UserBatch(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	if _, validationErr := userBatchPermission(req.Action, req.Days); validationErr != nil {
		ginx.WriteJSON(ctx, validationErr)
		return
	}
	if !canUseUserBatchPermission(operator, req.Action, req.Days) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	eligible, _ := collectUserBatch(ids, req.Action)
	if err := validateBatchConfirmation(req, req.Action, len(eligible)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	result := newAdminBatchResult(req.Action, ids, len(eligible))
	for _, user := range eligible {
		if req.Action == "forbid" {
			if err := services.UserService.Forbidden(operator.Id, user.Id, req.Days, req.Reason, ctx.Request); err != nil {
				appendAdminBatchFailure(&result, user.Id, err)
				continue
			}
		} else {
			services.UserService.RemoveForbidden(operator.Id, user.Id, ctx.Request)
		}
		result.ProcessedCount++
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityUser, 0,
		fmt.Sprintf("批量%s用户，成功%d，失败%d", req.Action, result.ProcessedCount, result.FailedCount), ctx.Request)
	ginx.WriteJSON(ctx, result)
}

func reportBatchStatus(action string) (int, error) {
	switch action {
	case "process":
		return 1, nil
	case "ignore":
		return 2, nil
	default:
		return 0, ginx.ErrorMessage("unsupported user report batch action")
	}
}

func collectReportBatch(user *models.User, ids []int64, processStatus int) ([]*models.UserReport, []int64) {
	eligible := make([]*models.UserReport, 0, len(ids))
	eligibleIds := make([]int64, 0, len(ids))
	for _, id := range ids {
		report := services.UserReportService.Get(id)
		if report == nil || report.ProcessStatus == int64(processStatus) || !userReportTargetAccessible(user, report) {
			continue
		}
		eligible = append(eligible, report)
		eligibleIds = append(eligibleIds, id)
	}
	return eligible, eligibleIds
}

func UserReportBatchPreview(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	if !canUseBatchPermission(operator, permissions.PermissionUserReportBatchProcess) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	processStatus, err := reportBatchStatus(req.Action)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	_, eligibleIds := collectReportBatch(operator, ids, processStatus)
	ginx.WriteJSON(ctx, newAdminBatchPreview(req.Action, ids, eligibleIds))
}

func UserReportBatch(ctx *gin.Context) {
	req, ids, err := bindAdminBatch(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	operator := common.GetCurrentUser(ctx)
	if !canUseBatchPermission(operator, permissions.PermissionUserReportBatchProcess) {
		ginx.WriteJSON(ctx, errs.NoPermission())
		return
	}
	processStatus, err := reportBatchStatus(req.Action)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	eligible, _ := collectReportBatch(operator, ids, processStatus)
	if err := validateBatchConfirmation(req, req.Action, len(eligible)); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	result := newAdminBatchResult(req.Action, ids, len(eligible))
	for _, report := range eligible {
		if err := services.UserReportService.Updates(report.Id, map[string]interface{}{
			"process_status":  processStatus,
			"process_time":    dates.NowTimestamp(),
			"process_user_id": operator.Id,
		}); err != nil {
			appendAdminBatchFailure(&result, report.Id, err)
			continue
		}
		result.ProcessedCount++
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "userReport", report.Id,
			fmt.Sprintf("批量更新举报处理状态：%d", processStatus), ctx.Request)
	}
	services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, "userReport", 0,
		fmt.Sprintf("批量%s举报，成功%d，失败%d", req.Action, result.ProcessedCount, result.FailedCount), ctx.Request)
	ginx.WriteJSON(ctx, result)
}
