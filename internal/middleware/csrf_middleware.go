package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/ginx"
	"github.com/mlogclub/simple/common/strs"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware protects cookie-authenticated admin mutations. Header-token
// clients are not browser cookie clients and continue to use their explicit
// bearer-style credential.
func CSRFMiddleware(ctx *gin.Context) {
	if isSafeMethod(ctx.Request.Method) || ginx.GetCookie(ctx, constants.CookieTokenKey) == "" {
		ctx.Next()
		return
	}
	if ctx.GetHeader("Authorization") != "" || ctx.GetHeader("X-User-Token") != "" {
		ctx.Next()
		return
	}
	cookieToken := ginx.GetCookie(ctx, constants.CookieCSRFTokenKey)
	headerToken := ctx.GetHeader("X-CSRF-Token")
	if cookieToken == "" || headerToken == "" || subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		ginx.WriteHttpStatusJSON(ctx, http.StatusForbidden, ginx.ErrorMessage("CSRF token is missing or invalid"))
		ctx.Abort()
		return
	}
	ctx.Next()
}

// ensureCSRFCookie upgrades existing cookie sessions created before CSRF was
// enabled. Header-token clients do not need a browser CSRF cookie.
func ensureCSRFCookie(ctx *gin.Context) {
	if ginx.GetCookie(ctx, constants.CookieTokenKey) == "" ||
		ctx.GetHeader("Authorization") != "" ||
		ctx.GetHeader("X-User-Token") != "" ||
		ginx.GetCookie(ctx, constants.CookieCSRFTokenKey) != "" {
		return
	}

	ginx.SetCookieKV(
		ctx,
		constants.CookieCSRFTokenKey,
		strs.UUID(),
		ginx.CookieExpires(365*time.Hour*24),
	)
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
