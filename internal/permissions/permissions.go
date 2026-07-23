package permissions

import "strings"

const (
	TypeDashboard = "dashboard"
)

const (
	GroupWorkspace = "workspace"
	GroupContent   = "content"
	GroupCommunity = "community"
	GroupSystem    = "system"
	GroupLogs      = "logs"
)

type PermissionDefinition struct {
	Type        string
	Code        string
	GroupName   string
	SortNo      int
	NameEn      string
	NameZh      string
	Description string
}

func (p PermissionDefinition) String() string {
	return p.Code
}

func (p PermissionDefinition) IsValid() bool {
	return p.Type != "" &&
		p.Code != "" &&
		p.GroupName != "" &&
		p.SortNo > 0 &&
		p.NameEn != "" &&
		p.NameZh != "" &&
		strings.HasPrefix(p.Code, p.Type+".")
}

var (
	PermissionDashboardView = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.view", GroupName: GroupWorkspace, SortNo: 10, NameEn: "Dashboard Access", NameZh: "进入后台"}

	PermissionTopicView         = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.view", GroupName: GroupContent, SortNo: 100, NameEn: "View Topics", NameZh: "查看话题"}
	PermissionTopicRecommend    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.recommend", GroupName: GroupContent, SortNo: 110, NameEn: "Recommend Topics", NameZh: "推荐话题"}
	PermissionTopicSticky       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.sticky", GroupName: GroupContent, SortNo: 115, NameEn: "Sticky Topics", NameZh: "置顶话题"}
	PermissionTopicDelete       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.delete", GroupName: GroupContent, SortNo: 130, NameEn: "Delete Topics", NameZh: "删除话题"}
	PermissionTopicSolve        = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.solve", GroupName: GroupContent, SortNo: 140, NameEn: "Solve Topics", NameZh: "标记问答解决"}
	PermissionTopicAcceptAnswer = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.acceptAnswer", GroupName: GroupContent, SortNo: 150, NameEn: "Manage Accepted Answers", NameZh: "管理采纳答案"}

	PermissionCommentView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.comment.view", GroupName: GroupContent, SortNo: 300, NameEn: "View Comments", NameZh: "查看评论"}
	PermissionCommentDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.comment.delete", GroupName: GroupContent, SortNo: 310, NameEn: "Delete Comments", NameZh: "删除评论"}

	PermissionCategoryView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.category.view", GroupName: GroupContent, SortNo: 400, NameEn: "View Categories", NameZh: "查看分类"}
	PermissionCategoryCreate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.category.create", GroupName: GroupContent, SortNo: 410, NameEn: "Create Categories", NameZh: "创建分类"}
	PermissionCategoryUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.category.update", GroupName: GroupContent, SortNo: 420, NameEn: "Update Categories", NameZh: "编辑分类"}
	PermissionCategoryDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.category.delete", GroupName: GroupContent, SortNo: 430, NameEn: "Delete Categories", NameZh: "删除分类"}
	PermissionCategorySort   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.category.sort", GroupName: GroupContent, SortNo: 440, NameEn: "Sort Categories", NameZh: "排序分类"}

	PermissionLinkView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.link.view", GroupName: GroupContent, SortNo: 500, NameEn: "View Links", NameZh: "查看链接"}
	PermissionLinkCreate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.link.create", GroupName: GroupContent, SortNo: 510, NameEn: "Create Links", NameZh: "创建链接"}
	PermissionLinkUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.link.update", GroupName: GroupContent, SortNo: 520, NameEn: "Update Links", NameZh: "编辑链接"}
	PermissionLinkDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.link.delete", GroupName: GroupContent, SortNo: 530, NameEn: "Delete Links", NameZh: "删除链接"}

	PermissionTagView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.tag.view", GroupName: GroupContent, SortNo: 600, NameEn: "View Tags", NameZh: "查看标签"}
	PermissionTagCreate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.tag.create", GroupName: GroupContent, SortNo: 610, NameEn: "Create Tags", NameZh: "创建标签"}
	PermissionTagUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.tag.update", GroupName: GroupContent, SortNo: 620, NameEn: "Update Tags", NameZh: "编辑标签"}

	PermissionUserView             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.view", GroupName: GroupCommunity, SortNo: 700, NameEn: "View Users", NameZh: "查看用户"}
	PermissionUserCreate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.create", GroupName: GroupCommunity, SortNo: 710, NameEn: "Create Users", NameZh: "创建用户"}
	PermissionUserUpdate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.update", GroupName: GroupCommunity, SortNo: 720, NameEn: "Update Users", NameZh: "编辑用户"}
	PermissionUserForbidden        = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.forbidden", GroupName: GroupCommunity, SortNo: 730, NameEn: "Forbid Users", NameZh: "禁言用户"}
	PermissionUserForbiddenForever = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.forbiddenForever", GroupName: GroupCommunity, SortNo: 735, NameEn: "Forbid Users Permanently", NameZh: "永久禁言用户"}
	PermissionUserUpdatePassword   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.updatePassword", GroupName: GroupCommunity, SortNo: 740, NameEn: "Update Own Password", NameZh: "修改自己的密码"}
	PermissionUserResetPassword    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.resetPassword", GroupName: GroupCommunity, SortNo: 750, NameEn: "Reset User Password", NameZh: "重置用户密码"}
	PermissionUserAccessScope      = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.accessScope", GroupName: GroupCommunity, SortNo: 760, NameEn: "Manage User Content Scope", NameZh: "管理用户内容范围"}

	PermissionSettingView            = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.setting.view", GroupName: GroupSystem, SortNo: 1100, NameEn: "View Settings", NameZh: "查看设置"}
	PermissionSettingUpdate          = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.setting.update", GroupName: GroupSystem, SortNo: 1110, NameEn: "Update Settings", NameZh: "编辑设置"}
	PermissionSettingSensitiveUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.setting.sensitive.update", GroupName: GroupSystem, SortNo: 1115, NameEn: "Update Sensitive Settings", NameZh: "编辑敏感设置"}
	PermissionSearchReindex          = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.search.reindex", GroupName: GroupSystem, SortNo: 1120, NameEn: "Rebuild Search Index", NameZh: "重建搜索索引"}
	PermissionSitemapGenerate        = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.sitemap.generate", GroupName: GroupSystem, SortNo: 1130, NameEn: "Generate Sitemap", NameZh: "生成 Sitemap"}

	PermissionRoleView             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.view", GroupName: GroupSystem, SortNo: 1200, NameEn: "View Roles", NameZh: "查看角色"}
	PermissionRoleCreate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.create", GroupName: GroupSystem, SortNo: 1210, NameEn: "Create Roles", NameZh: "创建角色"}
	PermissionRoleUpdate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.update", GroupName: GroupSystem, SortNo: 1220, NameEn: "Update Roles", NameZh: "编辑角色"}
	PermissionRoleDelete           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.delete", GroupName: GroupSystem, SortNo: 1230, NameEn: "Delete Roles", NameZh: "删除角色"}
	PermissionRoleSort             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.sort", GroupName: GroupSystem, SortNo: 1240, NameEn: "Sort Roles", NameZh: "排序角色"}
	PermissionRolePermissionUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.permission.update", GroupName: GroupSystem, SortNo: 1250, NameEn: "Update Role Permissions", NameZh: "编辑角色权限"}

	PermissionUserReportView    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.userReport.view", GroupName: GroupCommunity, SortNo: 790, NameEn: "View User Reports", NameZh: "查看用户举报"}
	PermissionUserReportUpdate  = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.userReport.update", GroupName: GroupCommunity, SortNo: 792, NameEn: "Update User Reports", NameZh: "编辑用户举报"}
	PermissionUserReportProcess = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.userReport.process", GroupName: GroupCommunity, SortNo: 795, NameEn: "Process User Reports", NameZh: "处理用户举报"}

	PermissionOperateLogView = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.operateLog.view", GroupName: GroupSystem, SortNo: 1310, NameEn: "View Operation Logs", NameZh: "查看操作日志"}

	PermissionFavoriteView     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.favorite.view", GroupName: GroupCommunity, SortNo: 800, NameEn: "View Favorites", NameZh: "查看收藏"}
	PermissionFavoriteManage   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.favorite.manage", GroupName: GroupCommunity, SortNo: 810, NameEn: "Manage Favorites", NameZh: "管理收藏"}
	PermissionVoteView         = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.vote.view", GroupName: GroupContent, SortNo: 900, NameEn: "View Votes", NameZh: "查看投票"}
	PermissionVoteCreate       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.vote.create", GroupName: GroupContent, SortNo: 910, NameEn: "Create Votes", NameZh: "创建投票"}
	PermissionVoteUpdate       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.vote.update", GroupName: GroupContent, SortNo: 920, NameEn: "Update Votes", NameZh: "编辑投票"}
	PermissionVoteDelete       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.vote.delete", GroupName: GroupContent, SortNo: 930, NameEn: "Delete Votes", NameZh: "删除投票"}
	PermissionVoteOptionView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteOption.view", GroupName: GroupContent, SortNo: 940, NameEn: "View Vote Options", NameZh: "查看投票选项"}
	PermissionVoteOptionCreate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteOption.create", GroupName: GroupContent, SortNo: 950, NameEn: "Create Vote Options", NameZh: "创建投票选项"}
	PermissionVoteOptionUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteOption.update", GroupName: GroupContent, SortNo: 960, NameEn: "Update Vote Options", NameZh: "编辑投票选项"}
	PermissionVoteOptionDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteOption.delete", GroupName: GroupContent, SortNo: 970, NameEn: "Delete Vote Options", NameZh: "删除投票选项"}
	PermissionVoteRecordView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteRecord.view", GroupName: GroupContent, SortNo: 980, NameEn: "View Vote Records", NameZh: "查看投票记录"}
	PermissionVoteRecordManage = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.voteRecord.manage", GroupName: GroupContent, SortNo: 990, NameEn: "Manage Vote Records", NameZh: "管理投票记录"}

	PermissionDictTypeView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dictType.view", GroupName: GroupSystem, SortNo: 1350, NameEn: "View Dictionary Types", NameZh: "查看字典类型"}
	PermissionDictTypeCreate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dictType.create", GroupName: GroupSystem, SortNo: 1360, NameEn: "Create Dictionary Types", NameZh: "创建字典类型"}
	PermissionDictTypeUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dictType.update", GroupName: GroupSystem, SortNo: 1370, NameEn: "Update Dictionary Types", NameZh: "编辑字典类型"}
	PermissionDictTypeDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dictType.delete", GroupName: GroupSystem, SortNo: 1380, NameEn: "Delete Dictionary Types", NameZh: "删除字典类型"}
	PermissionDictView       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dict.view", GroupName: GroupSystem, SortNo: 1390, NameEn: "View Dictionary Items", NameZh: "查看字典项"}
	PermissionDictCreate     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dict.create", GroupName: GroupSystem, SortNo: 1400, NameEn: "Create Dictionary Items", NameZh: "创建字典项"}
	PermissionDictUpdate     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dict.update", GroupName: GroupSystem, SortNo: 1410, NameEn: "Update Dictionary Items", NameZh: "编辑字典项"}
	PermissionDictDelete     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dict.delete", GroupName: GroupSystem, SortNo: 1420, NameEn: "Delete Dictionary Items", NameZh: "删除字典项"}
	PermissionDictSort       = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.dict.sort", GroupName: GroupSystem, SortNo: 1430, NameEn: "Sort Dictionary Items", NameZh: "排序字典项"}

	PermissionAttachmentView    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.attachment.view", GroupName: GroupContent, SortNo: 1000, NameEn: "View Attachments", NameZh: "查看附件"}
	PermissionAttachmentDelete  = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.attachment.delete", GroupName: GroupContent, SortNo: 1010, NameEn: "Delete Attachments", NameZh: "删除附件"}
	PermissionAttachmentCleanup = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.attachment.cleanup", GroupName: GroupSystem, SortNo: 1440, NameEn: "Clean Up Orphan Attachments", NameZh: "清理孤儿附件"}

	PermissionMessageView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.message.view", GroupName: GroupCommunity, SortNo: 820, NameEn: "View Messages", NameZh: "查看站内消息"}
	PermissionMessageSend   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.message.send", GroupName: GroupCommunity, SortNo: 830, NameEn: "Send Messages", NameZh: "发送站内消息"}
	PermissionMessageDelete = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.message.delete", GroupName: GroupCommunity, SortNo: 840, NameEn: "Delete Messages", NameZh: "删除站内消息"}

	PermissionHealthView = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.health.view", GroupName: GroupSystem, SortNo: 1450, NameEn: "View System Health", NameZh: "查看系统健康状态"}
	PermissionTaskView   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.task.view", GroupName: GroupSystem, SortNo: 1460, NameEn: "View Background Tasks", NameZh: "查看后台任务"}
)

var Permissions = []PermissionDefinition{
	PermissionDashboardView,
	PermissionTopicView,
	PermissionTopicRecommend,
	PermissionTopicSticky,
	PermissionTopicDelete,
	PermissionTopicSolve,
	PermissionTopicAcceptAnswer,
	PermissionCommentView,
	PermissionCommentDelete,
	PermissionCategoryView,
	PermissionCategoryCreate,
	PermissionCategoryUpdate,
	PermissionCategoryDelete,
	PermissionCategorySort,
	PermissionLinkView,
	PermissionLinkCreate,
	PermissionLinkUpdate,
	PermissionLinkDelete,
	PermissionTagView,
	PermissionTagCreate,
	PermissionTagUpdate,
	PermissionUserView,
	PermissionUserCreate,
	PermissionUserUpdate,
	PermissionUserForbidden,
	PermissionUserForbiddenForever,
	PermissionUserUpdatePassword,
	PermissionUserResetPassword,
	PermissionUserAccessScope,
	PermissionRoleView,
	PermissionRoleCreate,
	PermissionRoleUpdate,
	PermissionRoleDelete,
	PermissionRoleSort,
	PermissionRolePermissionUpdate,
	PermissionSettingView,
	PermissionSettingUpdate,
	PermissionSettingSensitiveUpdate,
	PermissionSearchReindex,
	PermissionSitemapGenerate,
	PermissionUserReportView,
	PermissionUserReportUpdate,
	PermissionUserReportProcess,
	PermissionOperateLogView,
	PermissionFavoriteView,
	PermissionFavoriteManage,
	PermissionVoteView,
	PermissionVoteCreate,
	PermissionVoteUpdate,
	PermissionVoteDelete,
	PermissionVoteOptionView,
	PermissionVoteOptionCreate,
	PermissionVoteOptionUpdate,
	PermissionVoteOptionDelete,
	PermissionVoteRecordView,
	PermissionVoteRecordManage,
	PermissionDictTypeView,
	PermissionDictTypeCreate,
	PermissionDictTypeUpdate,
	PermissionDictTypeDelete,
	PermissionDictView,
	PermissionDictCreate,
	PermissionDictUpdate,
	PermissionDictDelete,
	PermissionDictSort,
	PermissionAttachmentView,
	PermissionAttachmentDelete,
	PermissionAttachmentCleanup,
	PermissionMessageView,
	PermissionMessageSend,
	PermissionMessageDelete,
	PermissionHealthView,
	PermissionTaskView,
}

func FindByCode(code string) (PermissionDefinition, bool) {
	for _, permission := range Permissions {
		if permission.Code == code {
			return permission, true
		}
	}
	return PermissionDefinition{}, false
}
