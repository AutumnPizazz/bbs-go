package admin

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/web"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"
	"bbs-go/internal/services"
)

func AttachmentList(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)
	allowed := services.ContentAccessService.GetAllowedCategoryIds(operator)
	if len(allowed) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("no content categories are assigned to this administrator"))
		return
	}
	query := params.NewQueryParams(ctx).
		LikeByReq("file_name").
		LikeByReq("file_type").
		EqByReq("user_id").
		EqByReq("topic_id").
		EqByReq("status").
		PageByReq().Desc("id")
	query.Where("topic_id = ? OR topic_id IN (SELECT id FROM t_topic WHERE category_id IN ?)", 0, allowed)
	list, paging := services.AttachmentService.FindPageByParams(query)
	ginx.WriteJSON(ctx, &web.PageResult{Results: list, Page: paging})
}

func AttachmentDetail(ctx *gin.Context) {
	id := strings.TrimSpace(ctx.Param("id"))
	if id == "" {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("attachment id is required"))
		return
	}
	attachment := services.AttachmentService.GetAny(id)
	if attachment == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("attachment not found"))
		return
	}
	if !canAccessAttachment(common.GetCurrentUser(ctx), attachment) {
		ginx.WriteJSON(ctx, errs.ContentAccessDenied())
		return
	}
	ginx.WriteJSON(ctx, attachment)
}

func AttachmentRemove(ctx *gin.Context) {
	ids := splitAttachmentIDs(params.FormValue(ctx, "ids"))
	if len(ids) == 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("attachment ids are required"))
		return
	}
	operator := common.GetCurrentUser(ctx)
	for _, id := range ids {
		attachment := services.AttachmentService.GetAny(id)
		if attachment == nil {
			ginx.WriteJSON(ctx, ginx.ErrorMessage("attachment not found: "+id))
			return
		}
		if !canAccessAttachment(operator, attachment) {
			ginx.WriteJSON(ctx, errs.ContentAccessDenied())
			return
		}
		if attachment.Status == constants.StatusDeleted {
			continue
		}
		if err := services.AttachmentService.SoftDelete(id); err != nil {
			ginx.WriteJSON(ctx, err)
			return
		}
	}
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, constants.EntityAttachment, 0,
			"删除附件，数量："+stringCount(len(ids)), ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)
}

func canAccessAttachment(user *models.User, attachment *models.Attachment) bool {
	if attachment == nil || user == nil {
		return false
	}
	if attachment.TopicId <= 0 {
		return len(services.ContentAccessService.GetAllowedCategoryIds(user)) > 0
	}
	return services.ContentAccessService.CanAccessTopic(user, services.TopicService.Get(attachment.TopicId))
}

func AttachmentCleanupOrphans(ctx *gin.Context) {
	before, ok := params.GetInt64(ctx, "before")
	if !ok || before <= 0 || before > dates.NowTimestamp() {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("before must be a past timestamp"))
		return
	}
	count, err := services.AttachmentService.CleanupOrphans(before)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if operator := common.GetCurrentUser(ctx); operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeDelete, constants.EntityAttachment, 0,
			"清理孤儿附件，数量："+stringCount(int(count)), ctx.Request)
	}
	ginx.WriteJSON(ctx, map[string]interface{}{"count": count})
}

func splitAttachmentIDs(value string) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' || r == '\r' || r == '\t' }) {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func stringCount(value int) string {
	return strconv.Itoa(value)
}
