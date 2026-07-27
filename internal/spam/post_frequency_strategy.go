package spam

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/req"
)

// PostFrequencyStrategy 发表频率限制（已禁用，始终放行）
type PostFrequencyStrategy struct{}

func (PostFrequencyStrategy) Name() string {
	return "PostFrequencyStrategy"
}

func (PostFrequencyStrategy) CheckTopic(user *models.User, topic req.CreateTopicReq) error {
	return nil
}

func (s PostFrequencyStrategy) CheckComment(user *models.User, form req.CreateCommentReq) error {
	return nil
}
