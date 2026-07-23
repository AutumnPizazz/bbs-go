package admin

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models/constants"
	modelReq "bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/repositories"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/models"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/spf13/cast"

	"bbs-go/internal/services"
)

// 禁言
// 修改自己的密码
// PostResetPassword 重置密码
func userBuildUserItem(user *models.User, buildRoleIds bool) map[string]interface{} {
	mode := user.ContentAccessMode
	if user.IsOwner() {
		mode = constants.ContentAccessModeAll
	}
	if !constants.IsContentAccessModeValid(mode) {
		mode = constants.ContentAccessModeAssignedCategories
	}
	b := web.NewRspBuilder(user).
		Put("idEncode", idcodec.Encode(user.Id)).
		Put("roles", user.GetRoles()).
		Put("username", user.Username.String).
		Put("contentAccessMode", mode).
		Put("categoryIds", services.ContentAccessService.GetAssignedCategoryIds(user.Id)).
		Put("forbiddenEndTime", user.ForbiddenEndTime).
		Put("forbidden", user.IsForbidden())
	if buildRoleIds {
		b.Put("roleIds", services.UserRoleService.GetUserRoleIds(user.Id))
	}
	return b.Build()
}

func UserSynccount(ctx *gin.Context) {
	operator := common.GetCurrentUser(ctx)

	go func() {
		services.UserService.Scan(func(users []models.User) {
			for _, user := range users {
				topicCount := repositories.TopicRepository.Count(sqls.DB(), sqls.NewCnd().Eq("user_id", user.Id).Eq("status", constants.StatusOk))
				commentCount := repositories.CommentRepository.Count(sqls.DB(), sqls.NewCnd().Eq("user_id", user.Id).Eq("status", constants.StatusOk))
				_ = repositories.UserRepository.UpdateColumn(sqls.DB(), user.Id, "topic_count", topicCount)
				_ = repositories.UserRepository.UpdateColumn(sqls.DB(), user.Id, "comment_count", commentCount)
				cache.UserCache.Invalidate(user.Id)
			}
		})
	}()
	if operator != nil {
		services.OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityUser, 0, "启动用户内容计数同步", ctx.Request)
	}
	ginx.WriteJSON(ctx, nil)

}

func UserDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	t := services.UserService.Get(id)
	if t == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("Not found, id="+strconv.FormatInt(id, 10)))
		return
	}
	ginx.WriteJSON(ctx, userBuildUserItem(t, true))

}

func UserList(ctx *gin.Context) {
	cnd := params.NewPagedSqlCnd(ctx,
		params.QueryFilter{
			ParamName: "id",
			Op:        params.Eq,
			ValueWrapper: func(origin string) string {
				if id := idcodec.Decode(origin); id > 0 {
					return cast.ToString(id)
				}
				return ""
			},
		},
		params.QueryFilter{
			ParamName: "nickname",
			Op:        params.Like,
		},
		params.QueryFilter{
			ParamName: "username",
			Op:        params.Eq,
		},
	)
	if forbiddenValue := params.QueryValue(ctx, "forbidden"); forbiddenValue != "" {
		now := dates.NowTimestamp()
		if cast.ToBool(forbiddenValue) {
			cnd.Where("(forbidden_end_time = ? OR forbidden_end_time > ?)", -1, now)
		} else {
			cnd.Where("(forbidden_end_time >= ? AND forbidden_end_time <= ?)", 0, now)
		}
	}
	list, paging := services.UserService.FindPageByCnd(cnd.Desc("id"))
	var itemList []map[string]interface{}
	for _, user := range list {
		itemList = append(itemList, userBuildUserItem(&user, false))
	}
	ginx.WriteJSON(ctx, &web.PageResult{Results: itemList, Page: paging})

}

func UserCreate(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var req modelReq.AdminUserCreateReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	user, err := services.UserService.CreateManagedUser(operator, req, ctx.Request)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, userBuildUserItem(user, true))

}

func UserUpdate(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var req modelReq.AdminUserUpdateReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	user, err := services.UserService.UpdateManagedUser(operator, req, ctx.Request)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, userBuildUserItem(user, true))

}

func UserForbidden(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	var req modelReq.AdminUserForbiddenReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !services.PermissionService.CanForbiddenUser(user, req.Days) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(locales.Get("errors.no_permission")))
		return
	}
	if req.UserId < 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage(locales.Get("admin.user_id_required")))
		return
	}
	if req.Days == 0 {
		services.UserService.RemoveForbidden(user.Id, req.UserId, ctx.Request)
	} else {
		if err := services.UserService.Forbidden(user.Id, req.UserId, req.Days, req.Reason, ctx.Request); err != nil {
			ginx.WriteJSON(ctx, err)
			return
		}
	}
	ginx.WriteJSON(ctx, nil)

}

func UserUpdatePassword(ctx *gin.Context) {
	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var req modelReq.AdminPasswordUpdateReq
	if err := ginx.Bind(ctx, &req); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.UserService.UpdatePasswordByAdmin(operator, req.UserId, req.Password, req.RePassword, ctx.Request); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)

}

func UserResetPassword(ctx *gin.Context) {
	userId, _ := params.GetInt64(ctx, "userId")

	if userId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("invalid param: userId"))
		return
	}

	operator, err := common.CheckLogin(ctx)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	newPassword, err := services.UserService.ResetPasswordByAdmin(operator, userId, ctx.Request)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	ginx.WriteJSON(ctx, map[string]interface{}{
		"password": newPassword,
	})

}
