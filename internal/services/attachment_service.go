package services

import (
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/uploader"
	"bbs-go/internal/repositories"
)

var AttachmentService = new(attachmentService)

type attachmentService struct{}

func (s *attachmentService) extAllowed(ext string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return false
	}
	ext = strings.ToLower(ext)
	for _, a := range allowedTypes {
		allowed := strings.ToLower(strings.TrimSpace(a))
		if allowed == "*" || allowed == "*/*" || allowed == ext {
			return true
		}
	}
	return false
}

// Upload 流式上传附件；content 为数据流，contentLength 为文件大小（用于存储 FileSize 与上传 ContentLength）。
func (s *attachmentService) Upload(userId int64, filename string, content io.Reader, contentLength int64, contentType string) (*models.Attachment, error) {
	cfg := SysConfigService.GetAttachmentConfig()
	ext := strings.ToLower(filepath.Ext(filename))
	if !s.extAllowed(ext, cfg.AllowedTypes) {
		return nil, errors.New(locales.Get("attachment.ext_not_allowed"))
	}
	var (
		attId       = strs.UUID()
		key         = uploader.GenerateAttachmentKey(attId, ext)
		disposition = "attachment; filename=\"" + url.QueryEscape(filepath.Base(filename)) + "\""
	)
	fileUrl, err := UploadService.PutObject(key, content, &uploader.PutOptions{ContentType: contentType, ContentDisposition: disposition, ContentLength: contentLength})
	if err != nil {
		return nil, err
	}
	att := &models.Attachment{
		Id:         attId,
		TopicId:    0,
		UserId:     userId,
		FileName:   filename,
		FileUrl:    fileUrl,
		FileSize:   contentLength,
		FileType:   contentType,
		Status:     constants.StatusOk,
		CreateTime: dates.NowTimestamp(),
		UpdateTime: dates.NowTimestamp(),
	}
	if err := repositories.AttachmentRepository.Create(sqls.DB(), att); err != nil {
		return nil, err
	}
	return att, nil
}

// Get 根据 ID 获取附件（仅返回存在且正常的）
func (s *attachmentService) Get(id string) *models.Attachment {
	att := repositories.AttachmentRepository.Get(sqls.DB(), id)
	if att == nil || att.Status != constants.StatusOk {
		return nil
	}
	return att
}

// ListByTopicId 按帖子查询正常状态的附件
func (s *attachmentService) ListByTopicId(topicId int64) []models.Attachment {
	return repositories.AttachmentRepository.ListByTopicId(sqls.DB(), topicId)
}

// GetDownloadRedirectUrl 根据附件访问地址生成 302 目标 URL（Local 需拼 baseURL）
func (s *attachmentService) GetDownloadRedirectUrl(att *models.Attachment) string {
	return att.FileUrl
}

// Download 鉴权并返回下载重定向 URL。
func (s *attachmentService) Download(attachmentId string, userId int64) (redirectURL string, err error) {
	if strs.IsBlank(attachmentId) {
		return "", errors.New(locales.Get("attachment.not_found"))
	}
	att := repositories.AttachmentRepository.Get(sqls.DB(), attachmentId)
	if att == nil || att.Status != constants.StatusOk {
		return "", errors.New(locales.Get("attachment.not_found"))
	}
	if att.TopicId <= 0 {
		return "", errors.New(locales.Get("attachment.not_found"))
	}

	topic := repositories.TopicRepository.Get(sqls.DB(), att.TopicId)
	if topic == nil || topic.Status == constants.StatusDeleted {
		return "", errors.New(locales.Get("attachment.not_found"))
	}

	redirectURL = s.GetDownloadRedirectUrl(att)
	if strs.IsNotBlank(redirectURL) {
		repositories.AttachmentRepository.IncrDownloadCount(sqls.DB(), att.Id)
	}
	return redirectURL, nil
}

// SoftDeleteByTopicId 帖子删除时软删除其下所有附件
func (s *attachmentService) SoftDeleteByTopicId(ctx *sqls.TxContext, topicId int64) error {
	return repositories.AttachmentRepository.UpdateColumns(ctx.Tx, topicId, map[string]interface{}{
		"status":      constants.StatusDeleted,
		"update_time": dates.NowTimestamp(),
	})
}

// ReplaceTopicAttachments 编辑帖时全量替换附件
func (s *attachmentService) ReplaceTopicAttachments(ctx *sqls.TxContext, topicId, userId int64, attachmentIds []string) error {
	newSet := make(map[string]bool)
	for _, id := range attachmentIds {
		if strs.IsNotBlank(id) {
			newSet[id] = true
		}
	}

	// 从当前中移除的：解绑 + 软删除
	current := repositories.AttachmentRepository.ListByTopicId(ctx.Tx, topicId)
	for _, att := range current {
		if !newSet[att.Id] {
			if err := repositories.AttachmentRepository.Updates(ctx.Tx, att.Id, map[string]interface{}{
				"topic_id": 0, "status": constants.StatusDeleted, "update_time": dates.NowTimestamp(),
			}); err != nil {
				return err
			}
		}
	}

	// 新列表中的：校验归属且未绑其他帖，再绑定
	for _, aid := range attachmentIds {
		if strs.IsBlank(aid) {
			continue
		}
		att := repositories.AttachmentRepository.Get(ctx.Tx, aid)
		if att == nil || att.UserId != userId {
			return errors.New(locales.Get("attachment.no_permission"))
		}
		if att.TopicId != 0 && att.TopicId != topicId {
			return errors.New(locales.Get("attachment.already_bound"))
		}
		if err := repositories.AttachmentRepository.Updates(ctx.Tx, aid, map[string]interface{}{
			"topic_id": topicId, "status": constants.StatusOk, "update_time": dates.NowTimestamp(),
		}); err != nil {
			return err
		}
	}
	return nil
}

// CheckAttachmentsExistAndOwned 检查 attachmentIds 是否存在且均属于 userId，且未绑定其他帖子（或仅绑定 topicId）
func (s *attachmentService) CheckAttachmentsExistAndOwned(ctx *sqls.TxContext, userId int64, attachmentIds []string, topicId int64) error {
	for _, aid := range attachmentIds {
		if strs.IsBlank(aid) {
			continue
		}
		att := repositories.AttachmentRepository.Get(ctx.Tx, aid)
		if att == nil || att.UserId != userId {
			return errors.New(locales.Get("attachment.no_permission"))
		}
		if att.TopicId != 0 && att.TopicId != topicId {
			return errors.New(locales.Get("attachment.already_bound"))
		}
	}
	return nil
}
