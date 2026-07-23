package constants

const (
	DefaultTokenExpireDays       = 7   // 用户登录token默认有效期
	SummaryLen                   = 256 // 摘要长度
	UploadMaxM                   = 10
	UploadMaxBytes         int64 = 1024 * 1024 * 1024 * UploadMaxM
	CookieTokenKey               = "bbsgo_token"
	TopicListPageSize            = 30 // 帖子列表分页大小
)

// 昵称长度限制
const (
	NicknameMinLengthZhCN = 2
	NicknameMaxLengthZhCN = 12
	NicknameMinLengthEnUS = 2
	NicknameMaxLengthEnUS = 20
)

// 系统配置
const (
	SysConfigSiteTitle          = "siteTitle"          // 站点标题
	SysConfigSiteDescription    = "siteDescription"    // 站点描述
	SysConfigBaseURL            = "baseURL"            // 网站URL
	SysConfigSiteKeywords       = "siteKeywords"       // 站点关键字
	SysConfigSiteLogo           = "siteLogo"           // 站点Logo
	SysConfigSiteNavs           = "siteNavs"           // 站点导航
	SysConfigSiteNotification   = "siteNotification"   // 站点公告
	SysConfigAboutPageConfig    = "aboutPageConfig"    // 关于页配置
	SysConfigFooterLinks        = "footerLinks"        // 底部链接
	SysConfigRecommendTags      = "recommendTags"      // 推荐标签
	SysConfigUrlRedirect        = "urlRedirect"        // 是否开启链接跳转
	SysConfigDefaultCategoryId  = "defaultCategoryId"  // 发帖默认节点
	SysConfigTopicCaptcha       = "topicCaptcha"       // 是否开启发帖验证码
	SysConfigUserObserveSeconds = "userObserveSeconds" // 新用户观察期
	SysConfigTokenExpireDays    = "tokenExpireDays"    // 登录Token有效天数
	SysConfigEnableHideContent  = "enableHideContent"  // 启用回复可见功能
	SysConfigModules            = "modules"            // 功能模块
	SysConfigLoginConfig        = "loginConfig"        // 登录配置
	SysConfigUploadConfig       = "uploadConfig"       // 上传配置
	SysConfigAttachmentConfig   = "attachmentConfig"   // 附件配置（帖子附件）
	SysConfigScriptInjections   = "scriptInjections"   // head脚本注入配置
	SysConfigTopicListStyle     = "topicListStyle"     // 帖子列表样式
	SysConfigNotificationTypes  = "notificationTypes"  // 站内信通知类型配置
)

const (
	TopicListStyleDefault = "default"
	TopicListStyleCompact = "compact"
)

// EntityType
const (
	EntityTopic      = "topic"
	EntityComment    = "comment"
	EntityUser       = "user"
	EntityAttachment = "attachment"
)

// 用户角色
const (
	RoleOwner = "owner" // 站长
)

// ContentAccessMode controls the node scope used for content reads.
type ContentAccessMode string

const (
	ContentAccessModeAll                ContentAccessMode = "all"
	ContentAccessModeAssignedCategories ContentAccessMode = "assigned_categories"
)

func IsContentAccessModeValid(mode ContentAccessMode) bool {
	return mode == ContentAccessModeAll || mode == ContentAccessModeAssignedCategories
}

// 操作类型
const (
	OpTypeCreate          = "create"
	OpTypeDelete          = "delete"
	OpTypeUpdate          = "update"
	OpTypeForbidden       = "forbidden"
	OpTypeRemoveForbidden = "removeForbidden"
)

// 状态
const (
	StatusOk      = 0 // 正常
	StatusDeleted = 1 // 删除
)

// 角色类型
const (
	RoleTypeSystem = 0 // 系统角色
	RoleTypeCustom = 1 // 自定义角色
)

// 内容类型
type ContentType string

const (
	ContentTypeHtml     ContentType = "html"
	ContentTypeMarkdown ContentType = "markdown"
	ContentTypeText     ContentType = "text"
)

type TopicType int

const (
	TopicTypeTopic TopicType = 0 // 帖子
	TopicTypeQA    TopicType = 2 // 问答
)

type QaStatus string

const (
	QaStatusUnsolved QaStatus = "unsolved"
	QaStatusSolved   QaStatus = "solved"
)

func IsPostTopicType(topicType TopicType) bool {
	return topicType == TopicTypeTopic || topicType == TopicTypeQA
}

type VoteType int

const (
	VoteTypeSingle   VoteType = 1 // 单选
	VoteTypeMultiple VoteType = 2 // 多选
)

type ThirdType string

const (
	ThirdTypeWeixin ThirdType = "weixin"
	ThirdTypeGoogle ThirdType = "google"
	ThirdTypeGithub ThirdType = "github"
)

const (
	FollowStatusNONE   = 0
	FollowStatusFollow = 1
	FollowStatusBoth   = 2
)

const (
	CategoryIdNewest    int64 = 0
	CategoryIdRecommend int64 = -1
	CategoryIdFollow    int64 = -2
)

type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

// 模块
const (
	ModuleTopic   = "topic"
)
