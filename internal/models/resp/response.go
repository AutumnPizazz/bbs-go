package resp

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"time"

	"github.com/mlogclub/simple/web"
)

// UserInfo 用户简单信息
type UserInfo struct {
	Id           string           `json:"id"`
	Nickname     string           `json:"nickname"`
	Avatar       string           `json:"avatar"`
	SmallAvatar  string           `json:"smallAvatar"`
	Gender       constants.Gender `json:"gender"`
	Birthday     *time.Time       `json:"birthday"`
	TopicCount   int              `json:"topicCount"`   // 话题数量
	CommentCount int              `json:"commentCount"` // 跟帖数量
	FansCount    int              `json:"fansCount"`    // 粉丝数量
	FollowCount  int              `json:"followCount"`  // 关注数量
	Description  string           `json:"description"`
	CreateTime   int64            `json:"createTime"`

	Forbidden bool `json:"forbidden"` // 是否禁言
	Followed  bool `json:"followed"`  // 是否关注

}

// UserDetail 用户详细信息
type UserDetail struct {
	UserInfo
	Username             string `json:"username"`
	BackgroundImage      string `json:"backgroundImage"`
	SmallBackgroundImage string `json:"smallBackgroundImage"`
	HomePage             string `json:"homePage"`
	Status               int    `json:"status"`
}

// UserProfile 用户个人信息
type UserProfile struct {
	UserDetail
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	PasswordSet bool     `json:"passwordSet"` // 密码已设置
}

type TagResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryResponse struct {
	Id               int64                `json:"id"`
	ParentId         int64                `json:"parentId"` // 父节点ID，0=一级
	Name             string               `json:"name"`
	Logo             string               `json:"logo"`
	Description      string               `json:"description"`
	AttachmentConfig dto.AttachmentConfig `json:"attachmentConfig"`
	Children         []CategoryResponse   `json:"children,omitempty"` // 子节点（发帖可选时用）
}

// CategoryTreeItem 后台节点树形列表项（含 sortNo/status/createTime，children 始终存在以兼容 Arco Table）
type CategoryTreeItem struct {
	Id          int64              `json:"id"`
	ParentId    int64              `json:"parentId"`
	Name        string             `json:"name"`
	Logo        string             `json:"logo"`
	Description string             `json:"description"`
	SortNo      int                `json:"sortNo"`
	Status      int                `json:"status"`
	CreateTime  int64              `json:"createTime"`
	Children    []CategoryTreeItem `json:"children"` // 子节点，叶子节点为 []，保证 Arco Table 树形展示
}

type SearchTopicResponse struct {
	Id         int64             `json:"id"`
	User       *UserInfo         `json:"user"`
	Category   *CategoryResponse `json:"category"`
	Tags       *[]TagResponse    `json:"tags"`
	Title      string            `json:"title"`
	Summary    string            `json:"summary"`
	CreateTime int64             `json:"createTime"`
}

type SearchUserResponse struct {
	User        *UserInfo `json:"user"`
	Nickname    string    `json:"nickname"`
	Username    string    `json:"username"`
	Description string    `json:"description"`
	CreateTime  int64     `json:"createTime"`
}

type TopicTocItem struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Level int    `json:"level"`
}

// 帖子列表返回实体
type TopicResponse struct {
	Id                string                `json:"id"`
	Type              constants.TopicType   `json:"type"`
	QaStatus          constants.QaStatus    `json:"qaStatus"`
	AcceptedCommentId int64                 `json:"acceptedCommentId"`
	SolvedAt          int64                 `json:"solvedAt"`
	User              *UserInfo             `json:"user"`
	Category          *CategoryResponse     `json:"category"`
	Tags              *[]TagResponse        `json:"tags"`
	Title             string                `json:"title"`
	Summary           string                `json:"summary"`
	Content           string                `json:"content"`
	Toc               []TopicTocItem        `json:"toc,omitempty"`
	LastCommentTime   int64                 `json:"lastCommentTime"`
	ViewCount         int64                 `json:"viewCount"`
	CommentCount      int64                 `json:"commentCount"`
	LikeCount         int64                 `json:"likeCount"`
	Liked             bool                  `json:"liked"`
	CreateTime        int64                 `json:"createTime"`
	UpdateTime        int64                 `json:"updateTime"`
	Recommend         bool                  `json:"recommend"`
	RecommendTime     int64                 `json:"recommendTime"`
	Sticky            bool                  `json:"sticky"`
	StickyTime        int64                 `json:"stickyTime"`
	Status            int                   `json:"status"`
	Favorited         bool                  `json:"favorited"`
	IpLocation        string                `json:"ipLocation"`
	Vote              *VoteResponse         `json:"vote"`
	Attachments       []AttachmentResponse  `json:"attachments,omitempty"`
}

// AttachmentResponse 附件返回（不包含直链）
type AttachmentResponse struct {
	Id            string `json:"id"`            // ID
	FileName      string `json:"fileName"`      // 原始文件名
	FileSize      int64  `json:"fileSize"`      // 文件大小（字节）
	DownloadCount int    `json:"downloadCount"` // 下载次数
}

type VoteResponse struct {
	Id          int64                `json:"id"`
	Type        constants.VoteType   `json:"type"`
	Title       string               `json:"title"`
	ExpiredAt   int64                `json:"expiredAt"`
	VoteNum     int                  `json:"voteNum"`
	OptionCount int                  `json:"optionCount"`
	VoteCount   int                  `json:"voteCount"`
	Expired     bool                 `json:"expired"`
	Voted       bool                 `json:"voted"`
	OptionIds   []int64              `json:"optionIds"`
	Options     []VoteOptionResponse `json:"options"`
}

type VoteOptionResponse struct {
	Id        int64   `json:"id"`
	Content   string  `json:"content"`
	SortNo    int     `json:"sortNo"`
	VoteCount int     `json:"voteCount"`
	Percent   float64 `json:"percent"`
	Voted     bool    `json:"voted"`
}

// CommentResponse 评论返回数据
type CommentResponse struct {
	Id           int64                 `json:"id"`
	User         *UserInfo             `json:"user"`
	EntityType   string                `json:"entityType"`
	EntityId     int64                 `json:"entityId"`
	ContentType  constants.ContentType `json:"contentType"`
	Content      string                `json:"content"`
	ImageList    []ImageInfo           `json:"imageList"`
	LikeCount    int64                 `json:"likeCount"`
	CommentCount int64                 `json:"commentCount"`
	Liked        bool                  `json:"liked"`
	QuoteId      int64                 `json:"quoteId"`
	Quote        *CommentResponse      `json:"quote"`
	Replies      *web.CursorResult     `json:"replies"`
	IpLocation   string                `json:"ipLocation"`
	Status       int                   `json:"status"`
	CreateTime   int64                 `json:"createTime"`
}

// 收藏返回数据
type FavoriteResponse struct {
	Id         int64     `json:"id"`
	EntityType string    `json:"entityType"`
	EntityId   int64     `json:"entityId"`
	Deleted    bool      `json:"deleted"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	User       *UserInfo `json:"user"`
	Url        string    `json:"url"`
	CreateTime int64     `json:"createTime"`
}

// 消息
type MessageResponse struct {
	Id           int64     `json:"id"`
	From         *UserInfo `json:"from"`    // 消息发送人
	UserId       int64     `json:"userId"`  // 消息接收人编号
	Title        string    `json:"title"`   // 标题
	Content      string    `json:"content"` // 消息内容
	QuoteContent string    `json:"quoteContent"`
	Type         int       `json:"type"`
	DetailUrl    string    `json:"detailUrl"` // 消息详情url
	ExtraData    string    `json:"extraData"`
	Status       int       `json:"status"`
	CreateTime   int64     `json:"createTime"`
}

// 图片
type ImageInfo struct {
	Url     string `json:"url"`
	Preview string `json:"preview"`
}

type TreeNode struct {
	Id       int64      `json:"id"`
	Key      int64      `json:"key"`
	Title    string     `json:"title"`
	Children []TreeNode `json:"children"`
}

type MenuResponse struct {
	Id         int64  `json:"id"`
	ParentId   *int64 `json:"parentId"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	SortNo     int    `json:"sortNo"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
}

type MenuTreeResponse struct {
	MenuResponse
	Level    int                `json:"level"`
	Children []MenuTreeResponse `json:"children"`
}

type DictResponse struct {
	Id         int64  `json:"id"`
	TypeId     int64  `json:"typeId"`
	ParentId   *int64 `json:"parentId"`   // 上级分类
	Name       string `json:"name"`       // 名称
	Label      string `json:"label"`      // 标题
	Value      string `json:"value"`      // 值
	SortNo     int    `json:"sortNo"`     // 排序
	Status     int    `json:"status"`     // 状态
	CreateTime int64  `json:"createTime"` // 创建时间
	UpdateTime int64  `json:"updateTime"` // 更新时间
}

type DictListResponse struct {
	DictResponse
	Children []DictListResponse `json:"children"`
}
