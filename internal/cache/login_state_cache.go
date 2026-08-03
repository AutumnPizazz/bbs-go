package cache

import (
	"time"

	"bbs-go/internal/pkg/github"
	"bbs-go/internal/pkg/google"

	"github.com/goburrow/cache"
	"github.com/silenceper/wechat/v2/officialaccount/oauth"
)

// loginStateCache 第三方登录（GitHub/Google/微信）state 暂存缓存，
// 30 分钟无访问自动过期。
type loginStateCache[T any] struct {
	cache cache.Cache
}

func newLoginStateCache[T any]() *loginStateCache[T] {
	return &loginStateCache[T]{
		cache: cache.New(
			cache.WithMaximumSize(10000),
			cache.WithExpireAfterAccess(30*time.Minute),
		),
	}
}

func (c *loginStateCache[T]) Get(state string) *T {
	val, found := c.cache.GetIfPresent(state)
	if !found {
		return nil
	}
	return val.(*T)
}

func (c *loginStateCache[T]) Put(state string, data *T) {
	c.cache.Put(state, data)
}

type GithubLoginStateData struct {
	Redirect string
	Bind     bool // 表明当前是不是绑定流程
	UserInfo *github.GithubUserInfo
}

var GithubLoginStateCache = newLoginStateCache[GithubLoginStateData]()

type GoogleLoginStateData struct {
	Redirect string
	Bind     bool // 表明当前是不是绑定流程
	UserInfo *google.GoogleUserInfo
}

var GoogleLoginStateCache = newLoginStateCache[GoogleLoginStateData]()

type WxLoginStateData struct {
	Redirect string
	Bind     bool // 表明当前是不是绑定流程
	UserInfo oauth.UserInfo
}

var WxLoginStateCache = newLoginStateCache[WxLoginStateData]()
