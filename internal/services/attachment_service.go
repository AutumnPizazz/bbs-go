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
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/respath"
	"bbs-go/internal/pkg/uploader"
	"bbs-go/internal/repositories"
)

var AttachmentService = new(attachmentService)

type attachmentService struct {
	BaseService[models.Attachment, crudRepository[models.Attachment]]
}

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
func (s *attachmentService) Upload(userId, categoryId int64, filename string, content io.Reader, contentLength int64, contentType string) (*models.Attachment, error) {
	user := UserService.Get(userId)
	if user == nil {
		return nil, errors.New(locales.Get("errors.not_login"))
	}
	category := CategoryService.Get(categoryId)
	if category == nil || category.Status != constants.StatusOk {
		return nil, errors.New(locales.Get("topic.category_not_found"))
	}
	cfg := CategoryService.GetAttachmentConfig(categoryId)
	if err := s.validateAttachmentPolicy(cfg, filename, contentLength); err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(filename))
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
		StorageKey: key,
		FileSize:   contentLength,
		FileType:   contentType,
		Status:     constants.StatusOk,
		CreateTime: dates.NowTimestamp(),
		UpdateTime: dates.NowTimestamp(),
	}
	if err := repositories.AttachmentRepository.Create(sqls.DB(), att); err != nil {
		_ = UploadService.DeleteObject(key)
		return nil, err
	}
	return att, nil
}

func (s *attachmentService) validateAttachmentPolicy(cfg dto.AttachmentConfig, filename string, size int64) error {
	if !cfg.Enabled {
		return errors.New(locales.Get("attachment.disabled"))
	}
	if cfg.MaxSizeMB > 0 && size > int64(cfg.MaxSizeMB)*1024*1024 {
		return errors.New(locales.Getf("attachment.too_large", cfg.MaxSizeMB))
	}
	if !s.extAllowed(filepath.Ext(filename), cfg.AllowedTypes) {
		return errors.New(locales.Get("attachment.ext_not_allowed"))
	}
	return nil
}

// Get 根据 ID 获取附件（仅返回存在且正常的）
func (s *attachmentService) Get(id string) *models.Attachment {
	att := repositories.AttachmentRepository.Get(sqls.DB(), id)
	if att == nil || att.Status != constants.StatusOk {
		return nil
	}
	return att
}

// GetAny is used by the management console so deleted records remain traceable.
func (s *attachmentService) GetAny(id string) *models.Attachment {
	return repositories.AttachmentRepository.Get(sqls.DB(), id)
}

func (s *attachmentService) SoftDelete(id string) error {
	return repositories.AttachmentRepository.Updates(sqls.DB(), id, map[string]interface{}{
		"status":      constants.StatusDeleted,
		"update_time": dates.NowTimestamp(),
	})
}

// Undelete restores a soft-deleted attachment.
func (s *attachmentService) Undelete(id string) error {
	att := s.GetAny(id)
	if att == nil {
		return errors.New("attachment not found")
	}
	if att.Status != constants.StatusDeleted {
		return errors.New("attachment is not deleted")
	}
	return repositories.AttachmentRepository.Updates(sqls.DB(), id, map[string]interface{}{
		"status":      constants.StatusOk,
		"update_time": dates.NowTimestamp(),
	})
}

// Delete removes an unbound object from storage, then marks its record deleted.
func (s *attachmentService) Delete(id string) error {
	att := s.GetAny(id)
	if att == nil {
		return errors.New("attachment not found")
	}
	if att.TopicId > 0 {
		return errors.New("referenced attachments cannot be physically deleted")
	}
	key, err := attachmentStorageKey(att)
	if err != nil {
		return err
	}
	if key != "" {
		if err := UploadService.DeleteObject(key); err != nil {
			return err
		}
	}
	return s.SoftDelete(id)
}

// CleanupOrphans physically removes unbound uploads older than before.
func (s *attachmentService) CleanupOrphans(before int64) (int64, error) {
	var attachments []models.Attachment
	if err := sqls.DB().Where("topic_id = ? AND status = ? AND create_time < ?", 0, constants.StatusOk, before).Find(&attachments).Error; err != nil {
		return 0, err
	}
	var cleaned int64
	for i := range attachments {
		if err := s.Delete(attachments[i].Id); err != nil {
			return cleaned, err
		}
		cleaned++
	}
	return cleaned, nil
}

func attachmentStorageKey(att *models.Attachment) (string, error) {
	if att == nil {
		return "", nil
	}
	if strings.TrimSpace(att.StorageKey) != "" {
		key, err := uploader.NormalizeStorageKey(att.StorageKey)
		if err != nil {
			return "", err
		}
		return key, nil
	}
	if strings.TrimSpace(att.FileUrl) == "" {
		return "", nil
	}
	if strings.HasPrefix(strings.TrimSpace(att.FileUrl), respath.UploadsURLPrefix) {
		path := strings.TrimPrefix(strings.TrimLeft(strings.TrimSpace(att.FileUrl), "/"), strings.Trim(respath.UploadsURLPrefix, "/")+"/")
		return uploader.NormalizeStorageKey(path)
	}
	return UploadService.StorageKeyFromURL(att.FileUrl)
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
	if topic == nil || !ContentAccessService.CanAccessTopic(UserService.Get(userId), topic) {
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

// BindToComment binds uploaded attachments to a comment.
func (s *attachmentService) BindToComment(ctx *sqls.TxContext, commentId, userId int64, attachmentIds []string) error {
	for _, aid := range attachmentIds {
		if strs.IsBlank(aid) {
			continue
		}
		att := repositories.AttachmentRepository.Get(ctx.Tx, aid)
		if att == nil || att.UserId != userId {
			return errors.New(locales.Get("attachment.no_permission"))
		}
		if att.EntityId != 0 && att.EntityId != commentId {
			return errors.New(locales.Get("attachment.already_bound"))
		}
		if err := repositories.AttachmentRepository.Updates(ctx.Tx, aid, map[string]interface{}{
			"entity_type": constants.EntityComment,
			"entity_id":   commentId,
			"status":      constants.StatusOk,
			"update_time": dates.NowTimestamp(),
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListByCommentId returns attachments bound to a comment.
func (s *attachmentService) ListByCommentId(commentId int64) []models.Attachment {
	var list []models.Attachment
	sqls.DB().Where("entity_type = ? AND entity_id = ? AND status = ?",
		constants.EntityComment, commentId, constants.StatusOk).
		Order("create_time asc").Find(&list)
	return list
}

// SoftDeleteByCommentId soft-deletes attachments when a comment is deleted.
func (s *attachmentService) SoftDeleteByCommentId(ctx *sqls.TxContext, commentId int64) error {
	return ctx.Tx.Model(&models.Attachment{}).
		Where("entity_type = ? AND entity_id = ?", constants.EntityComment, commentId).
		Updates(map[string]interface{}{
			"status": constants.StatusDeleted, "update_time": dates.NowTimestamp(),
		}).Error
}

// ReplaceTopicAttachments 编辑帖时全量替换附件
func (s *attachmentService) ReplaceTopicAttachments(ctx *sqls.TxContext, topicId, userId, categoryId int64, attachmentIds []string) error {
	newSet := make(map[string]bool, len(attachmentIds))
	for _, id := range attachmentIds {
		if strs.IsNotBlank(id) {
			newSet[id] = true
		}
	}
	cfg := CategoryService.GetAttachmentConfig(categoryId)
	attachments := make(map[string]*models.Attachment, len(newSet))
	for aid := range newSet {
		att := repositories.AttachmentRepository.Get(ctx.Tx, aid)
		if att == nil || att.UserId != userId {
			return errors.New(locales.Get("attachment.no_permission"))
		}
		if att.TopicId != 0 && att.TopicId != topicId {
			return errors.New(locales.Get("attachment.already_bound"))
		}
		attachments[aid] = att
	}
	if len(attachments) > 0 {
		if !cfg.Enabled {
			return errors.New(locales.Get("attachment.disabled"))
		}
		if cfg.MaxCount > 0 && len(attachments) > cfg.MaxCount {
			return errors.New(locales.Getf("attachment.too_many", cfg.MaxCount))
		}
		for _, att := range attachments {
			if err := s.validateAttachmentPolicy(cfg, att.FileName, att.FileSize); err != nil {
				return err
			}
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
		if err := repositories.AttachmentRepository.Updates(ctx.Tx, aid, map[string]interface{}{
			"topic_id": topicId, "status": constants.StatusOk, "update_time": dates.NowTimestamp(),
		}); err != nil {
			return err
		}
	}
	return nil
}

// CheckAttachmentsExistAndOwned 检查附件归属、绑定状态和目标节点策略。
func (s *attachmentService) CheckAttachmentsExistAndOwned(ctx *sqls.TxContext, userId, categoryId int64, attachmentIds []string, topicId int64) error {
	uniqueIds := make(map[string]struct{}, len(attachmentIds))
	attachments := make(map[string]*models.Attachment, len(attachmentIds))
	for _, aid := range attachmentIds {
		if strs.IsBlank(aid) {
			continue
		}
		uniqueIds[aid] = struct{}{}
		att := repositories.AttachmentRepository.Get(ctx.Tx, aid)
		if att == nil || att.UserId != userId {
			return errors.New(locales.Get("attachment.no_permission"))
		}
		if att.TopicId != 0 && att.TopicId != topicId {
			return errors.New(locales.Get("attachment.already_bound"))
		}
		attachments[aid] = att
	}
	if len(attachments) == 0 {
		return nil
	}
	cfg := CategoryService.GetAttachmentConfig(categoryId)
	if !cfg.Enabled {
		return errors.New(locales.Get("attachment.disabled"))
	}
	if cfg.MaxCount > 0 && len(uniqueIds) > cfg.MaxCount {
		return errors.New(locales.Getf("attachment.too_many", cfg.MaxCount))
	}
	for _, att := range attachments {
		if err := s.validateAttachmentPolicy(cfg, att.FileName, att.FileSize); err != nil {
			return err
		}
	}
	return nil
}

// ValidateTopicAttachments checks existing attachments when a topic moves to another category.
func (s *attachmentService) ValidateTopicAttachments(topicId, categoryId int64) error {
	attachments := s.ListByTopicId(topicId)
	if len(attachments) == 0 {
		return nil
	}
	cfg := CategoryService.GetAttachmentConfig(categoryId)
	if !cfg.Enabled {
		return errors.New(locales.Get("attachment.disabled"))
	}
	if cfg.MaxCount > 0 && len(attachments) > cfg.MaxCount {
		return errors.New(locales.Getf("attachment.too_many", cfg.MaxCount))
	}
	for _, att := range attachments {
		if err := s.validateAttachmentPolicy(cfg, att.FileName, att.FileSize); err != nil {
			return err
		}
	}
	return nil
}
