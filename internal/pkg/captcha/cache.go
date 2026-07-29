package captcha

import (
	"time"

	"github.com/goburrow/cache"
	"github.com/wenlng/go-captcha/v2/slide"
)

var captchaCache cache.Cache

func init() {
	captchaCache = cache.New(
		cache.WithMaximumSize(1000),
		cache.WithExpireAfterAccess(10*time.Minute),
	)
}

func Get(captchaId string) *slide.Block {
	if v, ok := captchaCache.GetIfPresent(captchaId); ok {
		return v.(*slide.Block)
	}
	return nil
}

func Set(captchaId string, captcha *slide.Block) {
	captchaCache.Put(captchaId, captcha)
}

func Delete(captchaId string) {
	captchaCache.Invalidate(captchaId)
}
