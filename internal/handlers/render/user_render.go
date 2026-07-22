package render

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/bbsurls"
	"bbs-go/internal/pkg/idcodec"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/services"
	"strconv"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/spf13/cast"
)

func BuildUserInfoDefaultIfNull(id int64) *resp.UserInfo {
	user := cache.UserCache.Get(id)
	if user == nil {
		user = &models.User{}
		user.Id = id
		user.Username = sqls.SqlNullString(strconv.FormatInt(id, 10))
		user.Nickname = locales.Getf("user.anonymous", id)
		user.CreateTime = dates.NowTimestamp()
	}
	return BuildUserInfo(user)
}

func BuildUserInfo(user *models.User) *resp.UserInfo {
	if user == nil {
		return nil
	}
	ret := &resp.UserInfo{
		Id:           idcodec.Encode(user.Id),
		Nickname:     user.Nickname,
		Gender:       user.Gender,
		Birthday:     user.Birthday,
		TopicCount:   user.TopicCount,
		CommentCount: user.CommentCount,
		FansCount:    user.FansCount,
		FollowCount:  user.FollowCount,
		Description:  user.Description,
		CreateTime:   user.CreateTime,
		Forbidden:    user.IsForbidden(),
	}
	if strs.IsNotBlank(user.Avatar) {
		ret.Avatar = user.Avatar
		ret.SmallAvatar = HandleOssImageStyleAvatar(user.Avatar)
	} else {
		// avatar := RandomAvatar(user.Id)
		// ret.Avatar = avatar
		// ret.SmallAvatar = avatar
	}

	if len(ret.Description) == 0 {
		ret.Description = locales.Get("user.default_description")
	}
	if user.Status == constants.StatusDeleted {
		ret.Nickname = locales.Get("user.blacklist")
		ret.Description = ""
		ret.Forbidden = true
	} else if ret.Forbidden {
		redactForbiddenUserInfo(ret)
	}
	return ret
}

func redactForbiddenUserInfo(userInfo *resp.UserInfo) {
	userInfo.Nickname = locales.Get("user.forbidden_nickname")
	userInfo.Avatar = ""
	userInfo.SmallAvatar = ""
	userInfo.Description = ""
}

func BuildUserDetail(user *models.User) *resp.UserDetail {
	if user == nil {
		return nil
	}
	backgroundImage := user.BackgroundImage
	if strs.IsBlank(backgroundImage) {
		backgroundImage = "/res/images/default_user_bg.jpg"
	}
	ret := &resp.UserDetail{
		UserInfo:             *BuildUserInfo(user),
		Username:             user.Username.String,
		BackgroundImage:      backgroundImage,
		SmallBackgroundImage: HandleOssImageStyleSmall(backgroundImage),
		HomePage:             user.HomePage,
		Status:               user.Status,
	}
	if user.Status == constants.StatusDeleted {
		ret.Username = "blacklist"
		ret.HomePage = ""
	} else if ret.Forbidden {
		ret.Username = ""
	}
	return ret
}

func BuildUserProfile(user *models.User) *resp.UserProfile {
	if user == nil {
		return nil
	}
	ret := &resp.UserProfile{
		UserDetail:  *BuildUserDetail(user),
		PasswordSet: len(user.Password) > 0,
	}

	ret.Username = user.Username.String
	ret.Nickname = user.Nickname
	ret.Avatar = user.Avatar
	if strs.IsNotBlank(user.Avatar) {
		ret.SmallAvatar = HandleOssImageStyleAvatar(user.Avatar)
	}
	ret.Description = user.Description

	if strs.IsNotBlank(user.Roles) {
		ret.Roles = strings.Split(user.Roles, ",")
	}
	ret.Permissions = services.PermissionService.GetUserPermissionCodes(user)
	return ret
}

func RandomAvatar(userId int64) string {
	avatarCount := 128
	avatarIndex := userId % int64(avatarCount)
	return bbsurls.AbsUrl("/res/images/avatars/" + cast.ToString(avatarIndex) + ".png")
}
