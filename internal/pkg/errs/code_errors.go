package errs

import (
	"bbs-go/internal/pkg/locales"

	"github.com/mlogclub/simple/web"
)

func JsonError(err *web.CodeError) *web.JsonResult {
	if err == nil {
		return web.JsonSuccess()
	}
	return web.JsonErrorCode(err.Code, err.Message)
}

const (
	CodeNotLogin            = 1
	CodeNoPermission        = 2
	CodeCaptchaError        = 1000
	CodeForbiddenError      = 1001
	CodeUserDisabled        = 1002
	CodeInObservationPeriod = 1003
	CodeEmailNotVerified    = 1004
	CodeRegistrationClosed  = 1005
	CodeContentAccessDenied = 1006
)

// NewError 创建错误
func NewError(code int) *web.CodeError {
	var message string
	switch code {
	case CodeNotLogin:
		message = locales.Get("errors.not_login")
	case CodeNoPermission:
		message = locales.Get("errors.no_permission")
	case CodeCaptchaError:
		message = locales.Get("errors.captcha_error")
	case CodeForbiddenError:
		message = locales.Get("errors.forbidden")
	case CodeUserDisabled:
		message = locales.Get("errors.user_disabled")
	case CodeEmailNotVerified:
		message = locales.Get("errors.email_not_verified")
	case CodeRegistrationClosed:
		message = locales.Get("errors.registration_closed")
	case CodeContentAccessDenied:
		message = locales.Get("errors.content_access_denied")
	default:
		message = "Unknown error"
	}
	return web.NewError(code, message)
}

// 预定义的错误创建函数
var (
	NotLogin            = func() *web.CodeError { return NewError(CodeNotLogin) }
	NoPermission        = func() *web.CodeError { return NewError(CodeNoPermission) }
	CaptchaError        = func() *web.CodeError { return NewError(CodeCaptchaError) }
	ForbiddenError      = func() *web.CodeError { return NewError(CodeForbiddenError) }
	UserDisabled        = func() *web.CodeError { return NewError(CodeUserDisabled) }
	EmailNotVerified    = func() *web.CodeError { return NewError(CodeEmailNotVerified) }
	RegistrationClosed  = func() *web.CodeError { return NewError(CodeRegistrationClosed) }
	ContentAccessDenied = func() *web.CodeError { return NewError(CodeContentAccessDenied) }
)
