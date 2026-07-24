package services

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/common/urls"

	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/bbsurls"
	"bbs-go/internal/pkg/respath"
	"bbs-go/internal/pkg/uploader"
)

var UploadService = newUploadService()

type uploadService struct {
	uploaderMap map[dto.UploadMethod]uploader.Uploader
	once        sync.Once
}

func newUploadService() *uploadService {
	return &uploadService{
		uploaderMap: make(map[dto.UploadMethod]uploader.Uploader),
	}
}

func (s *uploadService) putObject(key string, body io.Reader, opts *uploader.PutOptions) (string, error) {
	u, err := s.getUploader()
	if err != nil {
		return "", err
	}
	cfg := SysConfigService.GetUploadConfig()
	return u.PutObject(cfg, key, body, opts)
}

// PutObject 按 key 流式上传；opts 可设置 ContentType、ContentDisposition、ContentLength。
func (s *uploadService) PutObject(key string, body io.Reader, opts *uploader.PutOptions) (string, error) {
	return s.putObject(key, body, opts)
}

func (s *uploadService) DeleteObject(key string) error {
	u, err := s.getUploader()
	if err != nil {
		return err
	}
	return u.DeleteObject(SysConfigService.GetUploadConfig(), key)
}

// StorageKeyFromURL derives a legacy storage key only when the URL belongs to
// the currently configured provider. This prevents an arbitrary URL from
// being interpreted as a deletable object key.
func (s *uploadService) StorageKeyFromURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	cfg := SysConfigService.GetUploadConfig()
	if strs.IsBlank(string(cfg.EnableUploadMethod)) {
		cfg.EnableUploadMethod = dto.Local
	}
	basePath := ""
	switch cfg.EnableUploadMethod {
	case dto.Local:
		basePath = strings.Trim(respath.UploadsURLPrefix, "/")
		if parsed.Host != "" {
			base, parseErr := url.Parse(SysConfigService.GetBaseURL())
			if parseErr != nil || base.Host == "" || !strings.EqualFold(base.Host, parsed.Host) {
				return "", fmt.Errorf("attachment URL does not belong to local storage")
			}
		}
		if !strings.HasPrefix(strings.TrimLeft(parsed.Path, "/"), basePath+"/") {
			return "", fmt.Errorf("attachment URL does not belong to local storage")
		}
	case dto.AliyunOss:
		basePath, err = requireStorageHost(parsed, cfg.AliyunOss.Host)
	case dto.TencentCos:
		expected := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.TencentCos.Bucket, cfg.TencentCos.Region)
		basePath, err = requireStorageHost(parsed, expected)
	case dto.AwsS3:
		expected := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.AwsS3.Bucket, cfg.AwsS3.Region)
		basePath, err = requireStorageHost(parsed, expected)
	default:
		return "", fmt.Errorf("unsupported upload method: %s", cfg.EnableUploadMethod)
	}
	if err != nil {
		return "", err
	}
	key := strings.TrimPrefix(strings.TrimLeft(parsed.Path, "/"), strings.Trim(basePath, "/"))
	key = strings.TrimLeft(key, "/")
	return uploader.NormalizeStorageKey(key)
}

func requireStorageHost(parsed *url.URL, expectedRaw string) (string, error) {
	expected, err := url.Parse(strings.TrimSpace(expectedRaw))
	if err != nil || expected.Host == "" || parsed.Host == "" || !strings.EqualFold(expected.Host, parsed.Host) {
		return "", fmt.Errorf("attachment URL does not belong to configured storage")
	}
	return strings.Trim(expected.Path, "/"), nil
}

func (s *uploadService) CheckConnectivity() error {
	u, err := s.getUploader()
	if err != nil {
		return err
	}
	return u.CheckConnectivity(SysConfigService.GetUploadConfig())
}

func (s *uploadService) ObjectURL(key string) string {
	cfg := SysConfigService.GetUploadConfig()
	if strs.IsBlank(string(cfg.EnableUploadMethod)) {
		cfg.EnableUploadMethod = dto.Local
	}

	switch cfg.EnableUploadMethod {
	case dto.AliyunOss:
		return bbsurls.UrlJoin(cfg.AliyunOss.Host, key)
	case dto.TencentCos:
		return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cfg.TencentCos.Bucket, cfg.TencentCos.Region, key)
	case dto.AwsS3:
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.AwsS3.Bucket, cfg.AwsS3.Region, key)
	default:
		return respath.UploadsURLPrefix + key
	}
}

// PutImage 上传图片（已有完整字节）；key 使用内容 MD5，供 CopyImage 等场景。
func (s *uploadService) PutImage(data []byte, contentType string) (string, error) {
	contentType = uploader.NormalizeImageContentType(contentType)
	key := uploader.GenerateImageKey(data, contentType)
	opts := &uploader.PutOptions{ContentType: contentType, ContentLength: int64(len(data))}
	return s.putObject(key, bytes.NewReader(data), opts)
}

// PutImageStream 流式上传图片；key 使用 UUID，无需先读完整 body。
func (s *uploadService) PutImageStream(body io.Reader, contentLength int64, contentType string) (string, error) {
	contentType = uploader.NormalizeImageContentType(contentType)
	key := uploader.GenerateImageKeyByContentType(contentType)
	opts := &uploader.PutOptions{ContentType: contentType, ContentLength: contentLength}
	return s.putObject(key, body, opts)
}

func (s *uploadService) CopyImage(url string) (string, error) {
	u, err := s.getUploader()
	if err != nil {
		return "", err
	}
	u1 := urls.ParseUrl(url).GetURL()
	u2 := urls.ParseUrl(SysConfigService.GetBaseURL()).GetURL()
	if u1.Host == u2.Host {
		return url, nil
	}
	cfg := SysConfigService.GetUploadConfig()
	return u.CopyImage(cfg, url)
}

func (s *uploadService) getUploader() (uploader.Uploader, error) {
	s.once.Do(func() {
		s.uploaderMap[dto.Local] = &uploader.LocalUploader{}
		s.uploaderMap[dto.AliyunOss] = &uploader.AliyunOssUploader{}
		s.uploaderMap[dto.TencentCos] = &uploader.TencentCosUploader{}
		s.uploaderMap[dto.AwsS3] = &uploader.AwsS3Uploader{}
	})
	cfg := SysConfigService.GetUploadConfig()

	if strs.IsBlank(string(cfg.EnableUploadMethod)) {
		cfg.EnableUploadMethod = dto.Local
	}

	u, ok := s.uploaderMap[cfg.EnableUploadMethod]
	if !ok {
		return nil, fmt.Errorf("error: Upload method: %s not found", cfg.EnableUploadMethod)
	}
	return u, nil
}
