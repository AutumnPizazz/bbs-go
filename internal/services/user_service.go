package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/search"
	"bbs-go/internal/pkg/validate"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"gorm.io/gorm"

	"bbs-go/internal/cache"

	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var UserService = newUserService()

func newUserService() *userService {
	return &userService{}
}

type userService struct {
}

func (s *userService) Get(id int64) *models.User {
	return repositories.UserRepository.Get(sqls.DB(), id)
}

func (s *userService) Take(where ...interface{}) *models.User {
	return repositories.UserRepository.Take(sqls.DB(), where...)
}

func (s *userService) Find(cnd *sqls.Cnd) []models.User {
	return repositories.UserRepository.Find(sqls.DB(), cnd)
}

func (s *userService) FindOne(cnd *sqls.Cnd) *models.User {
	return repositories.UserRepository.FindOne(sqls.DB(), cnd)
}

func (s *userService) FindPageByParams(params *params.QueryParams) (list []models.User, paging *sqls.Paging) {
	return repositories.UserRepository.FindPageByParams(sqls.DB(), params)
}

func (s *userService) FindPageByCnd(cnd *sqls.Cnd) (list []models.User, paging *sqls.Paging) {
	return repositories.UserRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *userService) Create(t *models.User) error {
	return errs.RegistrationClosed()
}

func (s *userService) Update(t *models.User) error {
	err := repositories.UserRepository.Update(sqls.DB(), t)
	cache.UserCache.Invalidate(t.Id)
	if err == nil {
		search.UpdateUserIndex(t)
	}
	return err
}

func (s *userService) Updates(id int64, columns map[string]interface{}) error {
	err := repositories.UserRepository.Updates(sqls.DB(), id, columns)
	cache.UserCache.Invalidate(id)
	if err == nil {
		search.UpdateUserIndex(s.Get(id))
	}
	return err
}

func (s *userService) UpdateColumn(id int64, name string, value interface{}) error {
	err := repositories.UserRepository.UpdateColumn(sqls.DB(), id, name, value)
	cache.UserCache.Invalidate(id)
	if err == nil {
		search.UpdateUserIndex(s.Get(id))
	}
	return err
}

func (s *userService) Delete(id int64) {
	repositories.UserRepository.Delete(sqls.DB(), id)
	cache.UserCache.Invalidate(id)
	_ = search.DeleteUserIndex(id)
}

// Scan 扫描
func (s *userService) Scan(callback func(users []models.User)) {
	var cursor int64
	for {
		list := repositories.UserRepository.Find(sqls.DB(), sqls.NewCnd().Where("id > ?", cursor).Asc("id").Limit(100))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

// Forbidden 禁言
func (s *userService) Forbidden(operatorId, userId int64, days int, reason string, r *http.Request) error {
	var forbiddenEndTime int64
	if days == -1 { // 永久禁言
		forbiddenEndTime = -1
	} else if days > 0 {
		forbiddenEndTime = dates.Timestamp(time.Now().Add(time.Hour * 24 * time.Duration(days)))
	} else {
		return errors.New(locales.Get("user.forbidden_time_invalid"))
	}
	if repositories.UserRepository.UpdateColumn(sqls.DB(), userId, "forbidden_end_time", forbiddenEndTime) == nil {
		cache.UserCache.Invalidate(userId)
		description := ""
		if strs.IsNotBlank(reason) {
			description = "禁言原因：" + reason
		}
		OperateLogService.AddOperateLog(operatorId, constants.OpTypeForbidden, constants.EntityUser, userId,
			description, r)

		// 永久禁言
		if days == -1 {
			go func() {
				// 删除话题
				TopicService.ScanByUser(userId, func(topics []models.Topic) {
					for _, topic := range topics {
						if topic.Status != constants.StatusDeleted {
							_ = TopicService.Delete(topic.Id, operatorId, nil)
						}
					}
				})

				// 删除评论
				CommentService.ScanByUser(userId, func(comments []models.Comment) {
					for _, comment := range comments {
						if comment.Status != constants.StatusDeleted {
							_ = CommentService.Delete(comment.Id)
						}
					}
				})

			}()
		}
	}
	return nil
}

// RemoveForbidden 移除禁言
func (s *userService) RemoveForbidden(operatorId, userId int64, r *http.Request) {
	user := s.Get(userId)
	if user == nil || !user.IsForbidden() {
		return
	}
	if repositories.UserRepository.UpdateColumn(sqls.DB(), userId, "forbidden_end_time", 0) == nil {
		cache.UserCache.Invalidate(user.Id)
		OperateLogService.AddOperateLog(operatorId, constants.OpTypeRemoveForbidden, constants.EntityUser, userId, "", r)
	}
}

// GetByUsername 根据用户名查找
func (s *userService) GetByUsername(username string) *models.User {
	return repositories.UserRepository.GetByUsername(sqls.DB(), username)
}

// SignIn 登录
func (s *userService) SignIn(username, password string) (*models.User, error) {
	if strs.IsBlank(username) {
		return nil, errors.New(locales.Get("user.username_required"))
	}
	if strs.IsBlank(password) {
		return nil, errors.New(locales.Get("user.password_required"))
	}
	if err := validate.IsPassword(password); err != nil {
		return nil, err
	}
	user := s.GetByUsername(username)
	if user == nil || user.Status != constants.StatusOk {
		return nil, errors.New(locales.Get("user.password_login_failed"))
	}
	if !passwd.ValidatePassword(user.Password, password) {
		return nil, errors.New(locales.Get("user.password_login_failed"))
	}
	return user, nil
}

// isUsernameExists 用户名是否存在
func (s *userService) isUsernameExists(username string) bool {
	return s.GetByUsername(username) != nil
}

// UpdateAvatar 更新头像
func (s *userService) UpdateAvatar(userId int64, avatar string) error {
	return s.UpdateColumn(userId, "avatar", avatar)
}

// UpdateNickname 更新昵称
func (s *userService) UpdateNickname(userId int64, nickname string) error {
	return s.UpdateColumn(userId, "nickname", nickname)
}

// UpdateDescription 更新简介
func (s *userService) UpdateDescription(userId int64, description string) error {
	return s.UpdateColumn(userId, "description", description)
}

// UpdateGender 修改性别
func (s *userService) UpdateGender(userId int64, gender string) error {
	if strs.IsBlank(gender) {
		return s.UpdateColumn(userId, "gender", "")
	} else {
		if gender != string(constants.GenderMale) && gender != string(constants.GenderFemale) {
			return errors.New("invalidate gender value")
		}
		return s.UpdateColumn(userId, "gender", gender)
	}
}

// UpdateBirthday 修改生日
func (s *userService) UpdateBirthday(userId int64, birthdayStr string) error {
	if strs.IsBlank(birthdayStr) {
		return s.UpdateColumn(userId, "birthday", "")
	} else {
		birthday, err := dates.Parse(birthdayStr, dates.FmtDate)
		if err != nil {
			return err
		}
		return s.UpdateColumn(userId, "birthday", birthday)
	}
}

// UpdateBackgroundImage 修改背景图
func (s *userService) UpdateBackgroundImage(userId int64, backgroundImage string) error {
	return s.UpdateColumn(userId, "background_image", backgroundImage)
}

// SetUsername 设置用户名
func (s *userService) SetUsername(userId int64, username string) error {
	username = strings.TrimSpace(username)
	if err := validate.IsUsername(username); err != nil {
		return err
	}

	user := s.Get(userId)
	if len(user.Username.String) > 0 {
		return errors.New(locales.Get("user.username_already_set"))
	}
	if s.isUsernameExists(username) {
		return errors.New(locales.Getf("user.username_occupied", username))
	}
	return s.UpdateColumn(userId, "username", username)
}

// IncrTopicCount topic_count + 1
func (s *userService) IncrTopicCount(ctx *sqls.TxContext, userId int64) error {
	if err := repositories.UserRepository.UpdateColumn(ctx.Tx, userId, "topic_count", gorm.Expr("topic_count + 1")); err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
		return err
	}
	ctx.RegisterCallback(func() {
		cache.UserCache.Invalidate(userId)
	})
	return nil
}

// IncrCommentCount comment_count + 1
func (s *userService) IncrCommentCount(userId int64) int {
	t := repositories.UserRepository.Get(sqls.DB(), userId)
	if t == nil {
		return 0
	}
	commentCount := t.CommentCount + 1
	if err := repositories.UserRepository.UpdateColumn(sqls.DB(), userId, "comment_count", commentCount); err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
	} else {
		cache.UserCache.Invalidate(userId)
	}
	return commentCount
}

// CheckPostStatus 用于在发表内容时检查用户状态
func (s *userService) CheckPostStatus(user *models.User) error {
	if user == nil {
		return errs.NotLogin()
	}
	if user.Status != constants.StatusOk {
		return errs.UserDisabled()
	}
	if user.IsForbidden() {
		return errs.ForbiddenError()
	}
	observeSeconds := cache.SysConfigCache.GetInt(constants.SysConfigUserObserveSeconds)
	if user.InObservationPeriod(observeSeconds) {
		return web.NewError(errs.CodeInObservationPeriod, locales.Getf("errors.in_observation", observeSeconds))
	}
	return nil
}
