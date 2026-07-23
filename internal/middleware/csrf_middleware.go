package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/ginx"

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

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
