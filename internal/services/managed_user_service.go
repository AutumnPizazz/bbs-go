package services

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	modelReq "bbs-go/internal/models/req"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/search"
	"bbs-go/internal/pkg/str"
	"bbs-go/internal/pkg/validate"
	"bbs-go/internal/repositories"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func (s *userService) CreateInitialOwner(user *models.User) error {
	if user == nil {
		return errors.New(locales.Get("user.not_found"))
	}
	if !constants.IsContentAccessModeValid(user.ContentAccessMode) {
		user.ContentAccessMode = constants.ContentAccessModeAll
	}
	if err := repositories.UserRepository.Create(sqls.DB(), user); err != nil {
		return err
	}
	cache.UserCache.Invalidate(user.Id)
	search.UpdateUserIndex(user)
	return nil
}

func (s *userService) CreateManagedUser(operator *models.User, form modelReq.AdminUserCreateReq, r *http.Request) (*models.User, error) {
	mode := constants.ContentAccessMode(strings.TrimSpace(form.ContentAccessMode))
	if mode == "" {
		mode = constants.ContentAccessModeAssignedCategories
	}
	categoryIds := modelReq.SplitCommaInt64s(form.CategoryIds)
	roleIds := modelReq.SplitCommaInt64s(form.RoleIds)
	if err := validateManagedUser(operator, mode, categoryIds, roleIds); err != nil {
		return nil, err
	}

	username := strings.TrimSpace(form.Username)
	phone := strings.TrimSpace(form.Phone)
	nickname := strings.TrimSpace(form.Nickname)
	if username == "" {
		return nil, errors.New(locales.Get("user.username_required"))
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if s.GetByUsername(username) != nil {
		return nil, errors.New(locales.Getf("user.username_occupied", username))
	}
	if phone != "" && s.GetByPhone(phone) != nil {
		return nil, errors.New(locales.Getf("user.phone_occupied", phone))
	}
	if nickname == "" {
		return nil, errors.New(locales.Get("user.nickname_required"))
	}
	if err := validateManagedPassword(form.Password); err != nil {
		return nil, err
	}

	user := &models.User{
		Username:          sqls.SqlNullString(username),
		Phone:             sqls.SqlNullString(phone),
		Nickname:          nickname,
		Password:          passwd.EncodePassword(form.Password),
		Status:            form.Status,
		ContentAccessMode: mode,
		CreateTime:        dates.NowTimestamp(),
		UpdateTime:        dates.NowTimestamp(),
	}

	if err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.UserRepository.Create(tx, user); err != nil {
			return err
		}
		if err := writeManagedUserRoles(tx, user.Id, roleIds); err != nil {
			return err
		}
		return writeUserCategoryAccess(tx, user.Id, mode, categoryIds)
	}); err != nil {
		return nil, err
	}

	cache.UserCache.Invalidate(user.Id)
	ContentAccessService.InvalidateUser(user.Id)
	PermissionService.InvalidateUser(user.Id)
	search.UpdateUserIndex(user)
	OperateLogService.AddOperateLog(operator.Id, constants.OpTypeCreate, constants.EntityUser, user.Id,
		fmt.Sprintf("创建用户并设置内容访问范围：mode=%s categories=%v", mode, categoryIds), r)
	return user, nil
}

func (s *userService) UpdateManagedUser(operator *models.User, form modelReq.AdminUserUpdateReq, r *http.Request) (*models.User, error) {
	target := s.Get(form.Id)
	if target == nil {
		return nil, errors.New(locales.Get("user.not_found"))
	}
	if target.IsOwner() && (operator == nil || !operator.IsOwner()) {
		return nil, errs.NoPermission()
	}
	mode := constants.ContentAccessMode(strings.TrimSpace(form.ContentAccessMode))
	if mode == "" {
		mode = target.ContentAccessMode
		if mode == "" {
			mode = constants.ContentAccessModeAssignedCategories
		}
	}
	categoryIds := modelReq.SplitCommaInt64s(form.CategoryIds)
	roleIds := modelReq.SplitCommaInt64s(form.RoleIds)
	if err := validateManagedUser(operator, mode, categoryIds, roleIds); err != nil {
		return nil, err
	}

	username := strings.TrimSpace(form.Username)
	phone := strings.TrimSpace(form.Phone)
	if username == "" {
		return nil, errors.New(locales.Get("user.username_required"))
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if other := s.GetByUsername(username); other != nil && other.Id != target.Id {
		return nil, errors.New(locales.Getf("user.username_occupied", username))
	}
	if phone != "" {
		if other := s.GetByPhone(phone); other != nil && other.Id != target.Id {
			return nil, errors.New(locales.Getf("user.phone_occupied", phone))
		}
	}

	oldMode := target.ContentAccessMode
	if target.IsOwner() {
		oldMode = constants.ContentAccessModeAll
	}
	oldCategoryIds := ContentAccessService.GetAssignedCategoryIds(target.Id)
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"username":            sqls.SqlNullString(username),
			"phone":               sqls.SqlNullString(phone),
			"nickname":            form.Nickname,
			"avatar":              form.Avatar,
			"gender":              constants.Gender(form.Gender),
			"home_page":           form.HomePage,
			"description":         form.Description,
			"status":              form.Status,
			"content_access_mode": mode,
			"update_time":         dates.NowTimestamp(),
		}
		if err := repositories.UserRepository.Updates(tx, target.Id, updates); err != nil {
			return err
		}
		if err := writeManagedUserRoles(tx, target.Id, roleIds); err != nil {
			return err
		}
		return writeUserCategoryAccess(tx, target.Id, mode, categoryIds)
	})
	if err != nil {
		return nil, err
	}

	cache.UserCache.Invalidate(target.Id)
	ContentAccessService.InvalidateUser(target.Id)
	PermissionService.InvalidateUser(target.Id)
	updated := s.Get(target.Id)
	search.UpdateUserIndex(updated)
	newCategoryIds := categoryIds
	if mode == constants.ContentAccessModeAll {
		newCategoryIds = nil
	}
	if oldMode != mode || !sameInt64Set(oldCategoryIds, newCategoryIds) {
		OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityUser, target.Id,
			fmt.Sprintf("更新用户内容访问范围：old_mode=%s old_categories=%v new_mode=%s new_categories=%v", oldMode, oldCategoryIds, mode, newCategoryIds), r)
	}
	return updated, nil
}

func (s *userService) UpdatePasswordByAdmin(operator *models.User, targetUserId int64, password, rePassword string, r *http.Request) error {
	target, err := s.validatePasswordTarget(operator, targetUserId)
	if err != nil {
		return err
	}
	if err := validate.IsValidPassword(password, rePassword); err != nil {
		return err
	}
	if err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.UserRepository.UpdateColumn(tx, target.Id, "password", passwd.EncodePassword(password)); err != nil {
			return err
		}
		return tx.Model(&models.UserToken{}).Where("user_id = ? AND status = ?", target.Id, constants.StatusOk).
			Update("status", constants.StatusDeleted).Error
	}); err != nil {
		return err
	}
	s.invalidateUserTokens(target.Id)
	OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityUser, target.Id, "管理员修改用户密码", r)
	return nil
}

func (s *userService) ResetPasswordByAdmin(operator *models.User, targetUserId int64, r *http.Request) (string, error) {
	target, err := s.validatePasswordTarget(operator, targetUserId)
	if err != nil {
		return "", err
	}
	newPassword := str.GenerateRandomPassword()
	if err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.UserRepository.UpdateColumn(tx, target.Id, "password", passwd.EncodePassword(newPassword)); err != nil {
			return err
		}
		return tx.Model(&models.UserToken{}).Where("user_id = ? AND status = ?", target.Id, constants.StatusOk).
			Update("status", constants.StatusDeleted).Error
	}); err != nil {
		return "", err
	}
	s.invalidateUserTokens(target.Id)
	OperateLogService.AddOperateLog(operator.Id, constants.OpTypeUpdate, constants.EntityUser, target.Id, "管理员重置用户密码", r)
	return newPassword, nil
}

func (s *userService) validatePasswordTarget(operator *models.User, targetUserId int64) (*models.User, error) {
	if operator == nil {
		return nil, errs.NotLogin()
	}
	target := s.Get(targetUserId)
	if target == nil {
		return nil, errors.New(locales.Get("user.not_found"))
	}
	if target.IsOwner() && !operator.IsOwner() {
		return nil, errs.NoPermission()
	}
	return target, nil
}

func (s *userService) invalidateUserTokens(userId int64) {
	activeTokens := repositories.UserTokenRepository.Find(sqls.DB(), sqls.NewCnd().Eq("user_id", userId).Eq("status", constants.StatusOk))
	for _, token := range activeTokens {
		cache.UserTokenCache.Invalidate(token.Token)
	}
	cache.UserCache.Invalidate(userId)
}

func validateManagedUser(operator *models.User, mode constants.ContentAccessMode, categoryIds, roleIds []int64) error {
	if operator == nil {
		return errs.NotLogin()
	}
	if !constants.IsContentAccessModeValid(mode) {
		return errors.New("invalid content access mode")
	}
	if mode == constants.ContentAccessModeAll && !operator.IsOwner() && !PermissionService.HasPermission(operator, "dashboard.user.accessScope") {
		return errs.NoPermission()
	}
	if mode == constants.ContentAccessModeAssignedCategories && len(categoryIds) == 0 {
		return errors.New("at least one category is required")
	}
	return validateManagedRoleAssignments(operator, roleIds)
}

func validateManagedRoleAssignments(operator *models.User, roleIds []int64) error {
	if len(roleIds) == 0 || operator.IsOwner() {
		return nil
	}
	roles := repositories.RoleRepository.Find(sqls.DB(), sqls.NewCnd().In("id", roleIds).Eq("status", constants.StatusOk))
	if len(roles) != len(roleIds) {
		return errors.New("invalid role")
	}
	operatorPermissions := make(map[string]struct{})
	for _, code := range PermissionService.GetUserPermissionCodes(operator) {
		operatorPermissions[code] = struct{}{}
	}
	for _, role := range roles {
		if role.Code == constants.RoleOwner {
			return errs.NoPermission()
		}
		for _, code := range RolePermissionService.GetRolePermissionCodes(role.Id) {
			if _, ok := operatorPermissions[code]; !ok {
				return errs.NoPermission()
			}
		}
	}
	return nil
}

func validateManagedPassword(password string) error {
	if strs.IsBlank(password) {
		return errors.New(locales.Get("user.password_required"))
	}
	return validate.IsValidPassword(password, password)
}

func validateUsername(username string) error {
	return validate.IsUsername(username)
}

func writeManagedUserRoles(tx *gorm.DB, userId int64, roleIds []int64) error {
	var roles []models.Role
	if len(roleIds) > 0 {
		roles = repositories.RoleRepository.Find(tx, sqls.NewCnd().In("id", roleIds))
	}
	if len(roleIds) != len(roles) {
		return errors.New("invalid role")
	}
	if err := tx.Where("user_id = ?", userId).Delete(&models.UserRole{}).Error; err != nil {
		return err
	}
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
		if err := repositories.UserRoleRepository.Create(tx, &models.UserRole{UserId: userId, RoleId: role.Id, CreateTime: dates.NowTimestamp()}); err != nil {
			return err
		}
	}
	return repositories.UserRepository.UpdateColumn(tx, userId, "roles", strings.Join(codes, ","))
}

func writeUserCategoryAccess(tx *gorm.DB, userId int64, mode constants.ContentAccessMode, categoryIds []int64) error {
	if err := repositories.UserCategoryAccessRepository.DeleteByUserId(tx, userId); err != nil {
		return err
	}
	if mode == constants.ContentAccessModeAll {
		return nil
	}
	categories := repositories.CategoryRepository.Find(tx, sqls.NewCnd().In("id", categoryIds).Eq("status", constants.StatusOk))
	if len(categories) != len(uniqueInt64s(categoryIds)) {
		return errors.New("invalid category")
	}
	for _, categoryId := range uniqueInt64s(categoryIds) {
		if err := repositories.UserCategoryAccessRepository.Create(tx, &models.UserCategoryAccess{UserId: userId, CategoryId: categoryId, CreateTime: dates.NowTimestamp(), UpdateTime: dates.NowTimestamp()}); err != nil {
			return err
		}
	}
	return nil
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	ret := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	return ret
}

func sameInt64Set(left, right []int64) bool {
	leftSet := make(map[int64]struct{}, len(left))
	for _, value := range uniqueInt64s(left) {
		leftSet[value] = struct{}{}
	}
	rightSet := make(map[int64]struct{}, len(right))
	for _, value := range uniqueInt64s(right) {
		rightSet[value] = struct{}{}
	}
	if len(leftSet) != len(rightSet) {
		return false
	}
	for value := range leftSet {
		if _, ok := rightSet[value]; !ok {
			return false
		}
	}
	return true
}
