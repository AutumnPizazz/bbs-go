package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/msg"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/tidwall/gjson"

	"gorm.io/gorm"

	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
)

var SysConfigService = newSysConfigService()

const (
	maxScriptInjectionCount   = 20
	maxScriptInjectionCodeLen = 20 * 1024
	maxScriptInjectionNameLen = 200
)

var adminConfigKeys = map[string]struct{}{
	constants.SysConfigSiteTitle: {}, constants.SysConfigSiteDescription: {}, constants.SysConfigBaseURL: {},
	constants.SysConfigSiteKeywords: {}, constants.SysConfigSiteLogo: {}, constants.SysConfigSiteNavs: {},
	constants.SysConfigSiteNotification: {}, constants.SysConfigAboutPageConfig: {}, constants.SysConfigFooterLinks: {},
	constants.SysConfigRecommendTags: {}, constants.SysConfigUrlRedirect: {}, constants.SysConfigDefaultCategoryId: {},
	constants.SysConfigTopicCaptcha: {}, constants.SysConfigUserObserveSeconds: {}, constants.SysConfigTokenExpireDays: {},
	constants.SysConfigEnableHideContent: {}, constants.SysConfigModules: {}, constants.SysConfigLoginConfig: {},
	constants.SysConfigUploadConfig: {}, constants.SysConfigAttachmentConfig: {}, constants.SysConfigScriptInjections: {},
	constants.SysConfigTopicListStyle: {}, constants.SysConfigNotificationTypes: {},
}

var sensitiveAdminConfigKeys = map[string]struct{}{
	constants.SysConfigLoginConfig:  {},
	constants.SysConfigUploadConfig: {},
}

var sensitiveConfigPaths = map[string]struct{}{
	"loginConfig.weixinLogin.appSecret":      {},
	"loginConfig.googleLogin.clientSecret":   {},
	"loginConfig.githubLogin.clientSecret":   {},
	"uploadConfig.aliyunOss.accessKeyId":     {},
	"uploadConfig.aliyunOss.accessKeySecret": {},
	"uploadConfig.tencentCos.secretId":       {},
	"uploadConfig.tencentCos.secretKey":      {},
	"uploadConfig.awsS3.accessKeyId":         {},
	"uploadConfig.awsS3.accessKeySecret":     {},
}

func newSysConfigService() *sysConfigService {
	return &sysConfigService{}
}

type sysConfigService struct {
}

func (s *sysConfigService) Get(id int64) *models.SysConfig {
	return repositories.SysConfigRepository.Get(sqls.DB(), id)
}

func (s *sysConfigService) Take(where ...interface{}) *models.SysConfig {
	return repositories.SysConfigRepository.Take(sqls.DB(), where...)
}

func (s *sysConfigService) Find(cnd *sqls.Cnd) []models.SysConfig {
	return repositories.SysConfigRepository.Find(sqls.DB(), cnd)
}

func (s *sysConfigService) FindOne(cnd *sqls.Cnd) *models.SysConfig {
	return repositories.SysConfigRepository.FindOne(sqls.DB(), cnd)
}

func (s *sysConfigService) FindPageByParams(params *params.QueryParams) (list []models.SysConfig, paging *sqls.Paging) {
	return repositories.SysConfigRepository.FindPageByParams(sqls.DB(), params)
}

func (s *sysConfigService) FindPageByCnd(cnd *sqls.Cnd) (list []models.SysConfig, paging *sqls.Paging) {
	return repositories.SysConfigRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *sysConfigService) GetAll() []models.SysConfig {
	return repositories.SysConfigRepository.Find(sqls.DB(), sqls.NewCnd().Asc("id"))
}

func (s *sysConfigService) SetAll(configStr string) error {
	configs, err := decodeAdminConfig(configStr, false)
	if err != nil {
		return err
	}
	return s.saveConfigValues(configs)
}

// SetSensitive updates OAuth and storage credentials without returning their
// values. Blank credential fields keep the stored value; clearSensitive is the
// explicit operation required to remove one.
func (s *sysConfigService) SetSensitive(configStr string) error {
	var input map[string]json.RawMessage
	if err := json.Unmarshal([]byte(configStr), &input); err != nil || input == nil {
		return errors.New(locales.Get("settings.invalid_format"))
	}
	clearPaths := []string{}
	if raw, ok := input["clearSensitive"]; ok {
		if err := json.Unmarshal(raw, &clearPaths); err != nil {
			return errors.New("clearSensitive must be an array")
		}
		for _, path := range clearPaths {
			if _, ok := sensitiveConfigPaths[path]; !ok {
				return fmt.Errorf("unknown sensitive config path: %s", path)
			}
		}
		delete(input, "clearSensitive")
	}
	if len(input) == 0 {
		return errors.New("sensitive configuration is required")
	}
	for key := range input {
		if _, ok := sensitiveAdminConfigKeys[key]; !ok {
			return fmt.Errorf("configuration key is not sensitive or not writable: %s", key)
		}
	}

	for _, key := range []string{constants.SysConfigLoginConfig, constants.SysConfigUploadConfig} {
		raw, ok := input[key]
		if !ok {
			continue
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil || object == nil {
			return fmt.Errorf("%s must be an object", key)
		}
		if key == constants.SysConfigLoginConfig {
			var cfg dto.LoginConfig
			if err := json.Unmarshal(raw, &cfg); err != nil {
				return fmt.Errorf("invalid login configuration: %w", err)
			}
		} else {
			var cfg dto.UploadConfig
			if err := json.Unmarshal(raw, &cfg); err != nil {
				return fmt.Errorf("invalid upload configuration: %w", err)
			}
		}
	}

	configs := make(map[string]string, len(input))
	for key, raw := range input {
		stored := cache.SysConfigCache.GetStr(key)
		merged, err := mergeSensitiveConfig(stored, string(raw), key, clearPaths)
		if err != nil {
			return err
		}
		configs[key] = merged
	}
	return s.saveConfigValues(configs)
}

func decodeAdminConfig(configStr string, allowSensitive bool) (map[string]string, error) {
	var input map[string]json.RawMessage
	if err := json.Unmarshal([]byte(configStr), &input); err != nil || input == nil {
		return nil, errors.New(locales.Get("settings.invalid_format"))
	}
	configs := make(map[string]string, len(input))
	for key, raw := range input {
		if _, ok := adminConfigKeys[key]; !ok {
			return nil, fmt.Errorf("unknown sys config key: %s", key)
		}
		if _, sensitive := sensitiveAdminConfigKeys[key]; sensitive && !allowSensitive {
			return nil, fmt.Errorf("sensitive configuration requires a separate permission: %s", key)
		}
		value, err := rawConfigValue(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", key, err)
		}
		configs[key] = value
	}
	if value, ok := configs[constants.SysConfigSiteNavs]; ok {
		if err := validateSiteNavs(value); err != nil {
			return nil, err
		}
	}
	if value, ok := configs[constants.SysConfigAboutPageConfig]; ok {
		if err := validateAboutPageConfig(value); err != nil {
			return nil, err
		}
	}
	if value, ok := configs[constants.SysConfigFooterLinks]; ok {
		if err := validateFooterLinks(value); err != nil {
			return nil, err
		}
	}
	if value, ok := configs[constants.SysConfigScriptInjections]; ok {
		if err := validateScriptInjections(value); err != nil {
			return nil, err
		}
	}
	return configs, nil
}

func rawConfigValue(raw json.RawMessage) (string, error) {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	if text, ok := value.(string); ok {
		return text, nil
	}
	return string(raw), nil
}

func (s *sysConfigService) saveConfigValues(configs map[string]string) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			if err := s.setSingle(tx, key, value, "", ""); err != nil {
				return err
			}
		}
		return nil
	})
}

func mergeSensitiveConfig(existing, incoming, root string, clearPaths []string) (string, error) {
	var current, next map[string]json.RawMessage
	if err := json.Unmarshal([]byte(existing), &current); err != nil || current == nil {
		current = map[string]json.RawMessage{}
	}
	if err := json.Unmarshal([]byte(incoming), &next); err != nil || next == nil {
		return "", errors.New("invalid sensitive configuration")
	}
	mergeSensitiveMaps(current, next, root, clearPaths)
	for _, path := range clearPaths {
		if strings.HasPrefix(path, root+".") {
			setSensitivePath(current, strings.Split(strings.TrimPrefix(path, root+"."), "."), json.RawMessage(`""`))
		}
	}
	var normalized interface{}
	if root == constants.SysConfigLoginConfig {
		var cfg dto.LoginConfig
		if err := marshalMapInto(nextMap(current), &cfg); err != nil {
			return "", err
		}
		normalized = cfg
	} else {
		var cfg dto.UploadConfig
		if err := marshalMapInto(nextMap(current), &cfg); err != nil {
			return "", err
		}
		normalized = cfg
	}
	encoded, err := json.Marshal(normalized)
	return string(encoded), err
}

func nextMap(value map[string]json.RawMessage) map[string]json.RawMessage { return value }

func marshalMapInto(value map[string]json.RawMessage, target interface{}) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

func mergeSensitiveMaps(current, incoming map[string]json.RawMessage, path string, clearPaths []string) {
	for key, value := range incoming {
		fullPath := path + "." + key
		var currentObject, incomingObject map[string]json.RawMessage
		if json.Unmarshal(value, &incomingObject) == nil && incomingObject != nil {
			if json.Unmarshal(current[key], &currentObject) != nil || currentObject == nil {
				currentObject = map[string]json.RawMessage{}
			}
			mergeSensitiveMaps(currentObject, incomingObject, fullPath, clearPaths)
			encoded, _ := json.Marshal(currentObject)
			current[key] = encoded
			continue
		}
		if _, sensitive := sensitiveConfigPaths[fullPath]; sensitive && isBlankJSON(value) && !containsConfigString(clearPaths, fullPath) {
			continue
		}
		current[key] = value
	}
}

func setSensitivePath(current map[string]json.RawMessage, parts []string, value json.RawMessage) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		current[parts[0]] = value
		return
	}
	var child map[string]json.RawMessage
	if json.Unmarshal(current[parts[0]], &child) != nil || child == nil {
		child = map[string]json.RawMessage{}
	}
	setSensitivePath(child, parts[1:], value)
	encoded, _ := json.Marshal(child)
	current[parts[0]] = encoded
}

func isBlankJSON(raw json.RawMessage) bool {
	var value string
	return json.Unmarshal(raw, &value) == nil && strings.TrimSpace(value) == ""
}

func containsConfigString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// Set 设置配置，如果配置不存在，那么创建
func (s *sysConfigService) Set(key, value string) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.setSingle(tx, key, value, "", ""); err != nil {
			return err
		}
		return nil
	})
}

func (s *sysConfigService) setSingle(db *gorm.DB, key, value, name, description string) error {
	if len(key) == 0 {
		return errors.New("sys config key is null")
	}
	sysConfig := repositories.SysConfigRepository.GetByKey(db, key)
	if sysConfig == nil {
		sysConfig = &models.SysConfig{
			CreateTime: dates.NowTimestamp(),
		}
	}
	sysConfig.Key = key
	sysConfig.Value = value
	sysConfig.UpdateTime = dates.NowTimestamp()

	if strs.IsNotBlank(name) {
		sysConfig.Name = name
	}
	if strs.IsNotBlank(description) {
		sysConfig.Description = description
	}

	var err error
	if sysConfig.Id > 0 {
		err = repositories.SysConfigRepository.Update(db, sysConfig)
	} else {
		err = repositories.SysConfigRepository.Create(db, sysConfig)
	}
	if err != nil {
		return err
	} else {
		cache.SysConfigCache.Invalidate(key)
		return nil
	}
}

func (s *sysConfigService) GetTokenExpireDays() int {
	tokenExpireDays := cache.SysConfigCache.GetInt(constants.SysConfigTokenExpireDays)
	if tokenExpireDays <= 0 {
		tokenExpireDays = constants.DefaultTokenExpireDays
	}
	return tokenExpireDays
}

func (s *sysConfigService) IsEnableHideContent() bool {
	return cache.SysConfigCache.GetBool(constants.SysConfigEnableHideContent)
}

func (s *sysConfigService) IsTopicCaptcha() bool {
	return cache.SysConfigCache.GetBool(constants.SysConfigTopicCaptcha)
}

func (s *sysConfigService) GetDefaultCategoryId() int64 {
	return cache.SysConfigCache.GetInt64(constants.SysConfigDefaultCategoryId)
}

func (s *sysConfigService) GetTopicListStyle() string {
	return normalizeTopicListStyle(cache.SysConfigCache.GetStr(constants.SysConfigTopicListStyle))
}

func normalizeTopicListStyle(style string) string {
	switch strings.TrimSpace(style) {
	case constants.TopicListStyleCompact:
		return constants.TopicListStyleCompact
	default:
		return constants.TopicListStyleDefault
	}
}

func (s *sysConfigService) GetSiteNavs() []dto.ActionLink {
	siteNavs := cache.SysConfigCache.GetStr(constants.SysConfigSiteNavs)
	var siteNavsArr []dto.ActionLink
	if strs.IsNotBlank(siteNavs) {
		if err := jsons.Parse(siteNavs, &siteNavsArr); err != nil {
			slog.Warn("站点导航数据错误", slog.Any("err", err))
		}
	}
	return siteNavsArr
}

func (s *sysConfigService) GetModules() dto.ModulesConfig {
	return parseModulesConfig(cache.SysConfigCache.GetStr(constants.SysConfigModules))
}

func defaultModulesConfig() dto.ModulesConfig {
	return dto.ModulesConfig{
		Topic: true,
		QA:    true,
	}
}

func parseModulesConfig(str string) dto.ModulesConfig {
	modulesConfig := defaultModulesConfig()
	if strs.IsBlank(str) {
		return modulesConfig
	}

	if err := jsons.Parse(str, &modulesConfig); err != nil {
		slog.Warn("启用模块配置错误", slog.Any("err", err))
		return defaultModulesConfig()
	}

	// 兼容历史配置：以前只有 topic 开关，提问跟随帖子。
	// 新配置存在 qa 字段时，提问可以独立于帖子开关控制。
	if !gjson.Get(str, "qa").Exists() {
		modulesConfig.QA = modulesConfig.Topic
	}
	return modulesConfig
}

func (s *sysConfigService) GetAboutPageConfig() dto.AboutPageConfig {
	str := cache.SysConfigCache.GetStr(constants.SysConfigAboutPageConfig)
	cfg := dto.AboutPageConfig{
		Content: dto.LocalizedText{},
	}
	if strings.TrimSpace(str) == "" {
		return cfg
	}
	if err := jsons.Parse(str, &cfg); err != nil {
		slog.Warn("关于页配置错误", slog.Any("err", err))
		return dto.AboutPageConfig{Content: dto.LocalizedText{}}
	}
	if cfg.Content == nil {
		cfg.Content = dto.LocalizedText{}
	}
	return cfg
}

func (s *sysConfigService) GetFooterLinks() []dto.FooterLink {
	str := cache.SysConfigCache.GetStr(constants.SysConfigFooterLinks)
	cfg := []dto.FooterLink{}
	if strings.TrimSpace(str) == "" {
		return cfg
	}
	if err := jsons.Parse(str, &cfg); err != nil {
		slog.Warn("底部链接配置错误", slog.Any("err", err))
		return []dto.FooterLink{}
	}
	return cfg
}

// GetNotificationTypes 各消息类型的站内信开关，缺省为全部开启
func (s *sysConfigService) GetNotificationTypes() map[string]dto.NoticeTypeConfig {
	str := cache.SysConfigCache.GetStr(constants.SysConfigNotificationTypes)
	out := make(map[string]dto.NoticeTypeConfig)
	if strs.IsNotBlank(str) {
		_ = jsons.Parse(str, &out)
	}
	// 默认补全缺失类型
	allKeys := []string{"topicComment", "commentReply", "topicLike", "topicFavorite", "topicRecommend", "topicDelete", "qaAnswerAccepted"}
	for _, k := range allKeys {
		if _, ok := out[k]; !ok {
			out[k] = dto.NoticeTypeConfig{Site: true}
		}
	}
	return out
}

// msgTypeToKey msg.Type 与 notificationTypes 的 key 对应
func msgTypeToKey(t msg.Type) string {
	switch t {
	case msg.TypeTopicComment:
		return "topicComment"
	case msg.TypeCommentReply:
		return "commentReply"
	case msg.TypeTopicLike:
		return "topicLike"
	case msg.TypeTopicFavorite:
		return "topicFavorite"
	case msg.TypeTopicRecommend:
		return "topicRecommend"
	case msg.TypeTopicDelete:
		return "topicDelete"
	case msg.TypeQaAnswerAccepted:
		return "qaAnswerAccepted"
	default:
		return ""
	}
}

// IsSiteNoticeEnabled 该消息类型是否发站内信
func (s *sysConfigService) IsSiteNoticeEnabled(msgType msg.Type) bool {
	key := msgTypeToKey(msgType)
	if key == "" {
		return true
	}
	types := s.GetNotificationTypes()
	c, ok := types[key]
	if !ok {
		return true
	}
	return c.Site
}

func (s *sysConfigService) IsUrlRedirect() bool {
	return cache.SysConfigCache.GetBool(constants.SysConfigUrlRedirect)
}

func (s *sysConfigService) GetLoginConfig() dto.LoginConfig {
	str := cache.SysConfigCache.GetStr(constants.SysConfigLoginConfig)
	var loginConfig dto.LoginConfig
	if err := jsons.Parse(str, &loginConfig); err != nil {
		slog.Warn("登录配置错误", slog.Any("err", err))
	}

	// 如果全部禁用，那么默认启用密码登录
	if loginConfig.IsAllDisabled() {
		loginConfig.PasswordLogin.Enabled = true
	}

	return loginConfig
}

func (s *sysConfigService) GetLoginConfigAdmin() dto.LoginConfigAdmin {
	cfg := s.GetLoginConfig()
	return dto.LoginConfigAdmin{
		PasswordLogin: cfg.PasswordLogin,
		WeixinLogin: dto.WeixinLoginAdmin{
			Enabled: cfg.WeixinLogin.Enabled, AppId: cfg.WeixinLogin.AppId,
			AppSecretConfigured: strings.TrimSpace(cfg.WeixinLogin.AppSecret) != "",
		},
		GoogleLogin: dto.OAuthLoginAdmin{
			Enabled: cfg.GoogleLogin.Enabled, ClientId: cfg.GoogleLogin.ClientId,
			ClientSecretConfigured: strings.TrimSpace(cfg.GoogleLogin.ClientSecret) != "",
		},
		GithubLogin: dto.OAuthLoginAdmin{
			Enabled: cfg.GithubLogin.Enabled, ClientId: cfg.GithubLogin.ClientId,
			ClientSecretConfigured: strings.TrimSpace(cfg.GithubLogin.ClientSecret) != "",
		},
	}
}

func (s *sysConfigService) GetUploadConfig() dto.UploadConfig {
	str := cache.SysConfigCache.GetStr(constants.SysConfigUploadConfig)
	var uploadConfig dto.UploadConfig
	if err := jsons.Parse(str, &uploadConfig); err != nil {
		slog.Warn("上传配置错误", slog.Any("err", err))
	}
	return uploadConfig
}

func (s *sysConfigService) GetUploadConfigAdmin() dto.UploadConfigAdmin {
	cfg := s.GetUploadConfig()
	return dto.UploadConfigAdmin{
		EnableUploadMethod: cfg.EnableUploadMethod,
		AliyunOss: dto.AliyunOssUploadConfigAdmin{
			Host: cfg.AliyunOss.Host, Bucket: cfg.AliyunOss.Bucket, Endpoint: cfg.AliyunOss.Endpoint,
			AccessKeyIdConfigured:     strings.TrimSpace(cfg.AliyunOss.AccessKeyId) != "",
			AccessKeySecretConfigured: strings.TrimSpace(cfg.AliyunOss.AccessKeySecret) != "",
			StyleSplitter:             cfg.AliyunOss.StyleSplitter, StyleAvatar: cfg.AliyunOss.StyleAvatar,
			StylePreview: cfg.AliyunOss.StylePreview, StyleSmall: cfg.AliyunOss.StyleSmall, StyleDetail: cfg.AliyunOss.StyleDetail,
		},
		TencentCos: dto.TencentCosUploadConfigAdmin{
			Bucket: cfg.TencentCos.Bucket, Region: cfg.TencentCos.Region,
			SecretIdConfigured:  strings.TrimSpace(cfg.TencentCos.SecretId) != "",
			SecretKeyConfigured: strings.TrimSpace(cfg.TencentCos.SecretKey) != "",
		},
		AwsS3: dto.AwsS3UploadConfigAdmin{
			Region: cfg.AwsS3.Region, Bucket: cfg.AwsS3.Bucket,
			AccessKeyIdConfigured:     strings.TrimSpace(cfg.AwsS3.AccessKeyId) != "",
			AccessKeySecretConfigured: strings.TrimSpace(cfg.AwsS3.AccessKeySecret) != "",
		},
	}
}

// GetAttachmentConfig 附件配置（帖子附件）
func (s *sysConfigService) GetAttachmentConfig() dto.AttachmentConfig {
	str := cache.SysConfigCache.GetStr(constants.SysConfigAttachmentConfig)
	cfg := defaultAttachmentConfig()
	if strings.TrimSpace(str) == "" {
		return cfg
	}
	if err := jsons.Parse(str, &cfg); err != nil {
		slog.Warn("附件配置解析错误", slog.Any("err", err))
	}
	return normalizeAttachmentConfig(cfg)
}

func normalizeAttachmentConfig(cfg dto.AttachmentConfig) dto.AttachmentConfig {
	if len(cfg.AllowedTypes) == 0 {
		cfg.AllowedTypes = defaultAttachmentConfig().AllowedTypes
	}
	return cfg
}

func defaultAttachmentConfig() dto.AttachmentConfig {
	return dto.AttachmentConfig{
		Enabled:      true,
		AllowedTypes: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv", ".zip", ".rar", ".7z", ".tar", ".gz"},
		MaxSizeMB:    10,
		MaxCount:     5,
	}
}

func (s *sysConfigService) GetScriptInjections() []dto.ScriptInjection {
	str := cache.SysConfigCache.GetStr(constants.SysConfigScriptInjections)
	if strings.TrimSpace(str) == "" {
		return []dto.ScriptInjection{}
	}

	var injections []dto.ScriptInjection
	if err := jsons.Parse(str, &injections); err != nil {
		slog.Warn("脚本注入配置错误", slog.Any("err", err))
		return []dto.ScriptInjection{}
	}
	return injections
}

func (s *sysConfigService) GetBaseURL() string {
	baseURL := strings.TrimSpace(cache.SysConfigCache.GetStr(constants.SysConfigBaseURL))
	if baseURL == "" {
		return "/"
	}
	for len(baseURL) > 1 && strings.HasSuffix(baseURL, "/") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	return baseURL
}

func validateSiteNavs(siteNavsJson string) error {
	if strings.TrimSpace(siteNavsJson) == "" {
		return nil
	}
	var navs []dto.ActionLink
	if err := jsons.Parse(siteNavsJson, &navs); err != nil {
		return errors.New("invalid site navigation data format")
	}
	return validateActionLinks(navs, 1)
}

func validateAboutPageConfig(aboutPageConfigJSON string) error {
	if strings.TrimSpace(aboutPageConfigJSON) == "" {
		return nil
	}

	var cfg dto.AboutPageConfig
	if err := jsons.Parse(aboutPageConfigJSON, &cfg); err != nil {
		return errors.New("invalid about page config data format")
	}
	return nil
}

func validateFooterLinks(footerLinksJSON string) error {
	if strings.TrimSpace(footerLinksJSON) == "" {
		return nil
	}

	var links []dto.FooterLink
	if err := jsons.Parse(footerLinksJSON, &links); err != nil {
		return errors.New("invalid footer links data format")
	}
	for idx, item := range links {
		if !hasLocalizedText(item.Text) {
			return fmt.Errorf("footer link text is required at item %d", idx+1)
		}
	}
	return nil
}

func hasLocalizedText(text dto.LocalizedText) bool {
	for _, value := range text {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func validateActionLinks(navs []dto.ActionLink, depth int) error {
	if depth > 2 {
		return errors.New("site navigation supports at most two levels")
	}
	for idx, nav := range navs {
		if strings.TrimSpace(nav.Title) == "" {
			return fmt.Errorf("navigation title is required at item %d", idx+1)
		}
		if depth == 1 {
			if len(nav.Children) == 0 && strings.TrimSpace(nav.Url) == "" {
				return fmt.Errorf("primary navigation URL is required at item %d", idx+1)
			}
			if len(nav.Children) > 0 {
				if err := validateActionLinks(nav.Children, depth+1); err != nil {
					return err
				}
			}
			continue
		}
		if strings.TrimSpace(nav.Url) == "" {
			return fmt.Errorf("secondary navigation URL is required at item %d", idx+1)
		}
		if len(nav.Children) > 0 {
			return errors.New("site navigation supports at most two levels")
		}
	}
	return nil
}

func validateScriptInjections(scriptInjectionsJSON string) error {
	if strings.TrimSpace(scriptInjectionsJSON) == "" {
		return nil
	}

	var injections []dto.ScriptInjection
	if err := jsons.Parse(scriptInjectionsJSON, &injections); err != nil {
		return errors.New("invalid script injections data format")
	}

	if len(injections) > maxScriptInjectionCount {
		return fmt.Errorf("too many script injections, max %d", maxScriptInjectionCount)
	}

	for idx, injection := range injections {
		scriptName := strings.TrimSpace(injection.ScriptName)
		if scriptName == "" {
			return fmt.Errorf("script injection name is required at item %d", idx+1)
		}
		if len([]rune(scriptName)) > maxScriptInjectionNameLen {
			return fmt.Errorf("script injection name is too long at item %d", idx+1)
		}
		injectionType := strings.TrimSpace(injection.Type)
		switch injectionType {
		case "external":
			if strings.TrimSpace(injection.Src) == "" {
				return fmt.Errorf("script injection src is required at item %d", idx+1)
			}
		case "inline":
			if strings.TrimSpace(injection.Code) == "" {
				return fmt.Errorf("script injection code is required at item %d", idx+1)
			}
			if len(injection.Code) > maxScriptInjectionCodeLen {
				return fmt.Errorf("script injection code is too long at item %d", idx+1)
			}
		default:
			return fmt.Errorf("script injection type must be external or inline at item %d", idx+1)
		}
	}
	return nil
}
