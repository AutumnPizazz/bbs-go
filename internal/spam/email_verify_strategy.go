package spam

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"
)

type EmailVerifyStrategy struct{}

func (EmailVerifyStrategy) Name() string {
	return "EmailVerifyStrategy"
}

func (EmailVerifyStrategy) CheckTopic(user *models.User, form req.CreateTopicReq) error {
	if constants.IsArticleTopicFormat(form.Format) {
		if services.SysConfigService.IsCreateArticleEmailVerified() && !user.EmailVerified {
			return errs.EmailNotVerified()
		}
		return nil
	}
	if services.SysConfigService.IsCreateTopicEmailVerified() && !user.EmailVerified {
		return errs.EmailNotVerified()
	}
	return nil
}

func (EmailVerifyStrategy) CheckComment(user *models.User, form req.CreateCommentReq) error {
	if services.SysConfigService.IsCreateCommentEmailVerified() && !user.EmailVerified {
		return errs.EmailNotVerified()
	}
	return nil
}
