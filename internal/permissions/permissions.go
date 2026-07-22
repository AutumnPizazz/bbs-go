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

	PermissionTopicView      = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.view", GroupName: GroupContent, SortNo: 100, NameEn: "View Topics", NameZh: "查看话题"}
	PermissionTopicRecommend = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.recommend", GroupName: GroupContent, SortNo: 110, NameEn: "Recommend Topics", NameZh: "推荐话题"}
	PermissionTopicSticky    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.sticky", GroupName: GroupContent, SortNo: 115, NameEn: "Sticky Topics", NameZh: "置顶话题"}
	PermissionTopicDelete    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.delete", GroupName: GroupContent, SortNo: 130, NameEn: "Delete Topics", NameZh: "删除话题"}
	PermissionTopicSolve     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.topic.solve", GroupName: GroupContent, SortNo: 140, NameEn: "Solve Topics", NameZh: "标记问答解决"}

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


	PermissionUserView             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.view", GroupName: GroupCommunity, SortNo: 700, NameEn: "View Users", NameZh: "查看用户"}
	PermissionUserCreate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.create", GroupName: GroupCommunity, SortNo: 710, NameEn: "Create Users", NameZh: "创建用户"}
	PermissionUserUpdate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.update", GroupName: GroupCommunity, SortNo: 720, NameEn: "Update Users", NameZh: "编辑用户"}
	PermissionUserForbidden        = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.forbidden", GroupName: GroupCommunity, SortNo: 730, NameEn: "Forbid Users", NameZh: "禁言用户"}
	PermissionUserForbiddenForever = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.forbiddenForever", GroupName: GroupCommunity, SortNo: 735, NameEn: "Forbid Users Permanently", NameZh: "永久禁言用户"}
	PermissionUserUpdatePassword   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.updatePassword", GroupName: GroupCommunity, SortNo: 740, NameEn: "Update Own Password", NameZh: "修改自己的密码"}
	PermissionUserResetPassword    = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.resetPassword", GroupName: GroupCommunity, SortNo: 750, NameEn: "Reset User Password", NameZh: "重置用户密码"}
	PermissionUserAccessScope      = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.user.accessScope", GroupName: GroupCommunity, SortNo: 760, NameEn: "Manage User Content Scope", NameZh: "管理用户内容范围"}

	PermissionSettingView     = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.setting.view", GroupName: GroupSystem, SortNo: 1100, NameEn: "View Settings", NameZh: "查看设置"}
	PermissionSettingUpdate   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.setting.update", GroupName: GroupSystem, SortNo: 1110, NameEn: "Update Settings", NameZh: "编辑设置"}
	PermissionSearchReindex   = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.search.reindex", GroupName: GroupSystem, SortNo: 1120, NameEn: "Rebuild Search Index", NameZh: "重建搜索索引"}
	PermissionSitemapGenerate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.sitemap.generate", GroupName: GroupSystem, SortNo: 1130, NameEn: "Generate Sitemap", NameZh: "生成 Sitemap"}

	PermissionRoleView             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.view", GroupName: GroupSystem, SortNo: 1200, NameEn: "View Roles", NameZh: "查看角色"}
	PermissionRoleCreate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.create", GroupName: GroupSystem, SortNo: 1210, NameEn: "Create Roles", NameZh: "创建角色"}
	PermissionRoleUpdate           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.update", GroupName: GroupSystem, SortNo: 1220, NameEn: "Update Roles", NameZh: "编辑角色"}
	PermissionRoleDelete           = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.delete", GroupName: GroupSystem, SortNo: 1230, NameEn: "Delete Roles", NameZh: "删除角色"}
	PermissionRoleSort             = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.sort", GroupName: GroupSystem, SortNo: 1240, NameEn: "Sort Roles", NameZh: "排序角色"}
	PermissionRolePermissionUpdate = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.role.permission.update", GroupName: GroupSystem, SortNo: 1250, NameEn: "Update Role Permissions", NameZh: "编辑角色权限"}

	PermissionUserReportView  = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.userReport.view", GroupName: GroupCommunity, SortNo: 790, NameEn: "View User Reports", NameZh: "查看用户举报"}
	PermissionUserReportProcess = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.userReport.process", GroupName: GroupCommunity, SortNo: 795, NameEn: "Process User Reports", NameZh: "处理用户举报"}

	PermissionOperateLogView = PermissionDefinition{Type: TypeDashboard, Code: "dashboard.operateLog.view", GroupName: GroupSystem, SortNo: 1310, NameEn: "View Operation Logs", NameZh: "查看操作日志"}
)

var Permissions = []PermissionDefinition{
	PermissionDashboardView,
	PermissionTopicView,
	PermissionTopicRecommend,
	PermissionTopicSticky,
	PermissionTopicDelete,
	PermissionTopicSolve,
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
	PermissionSearchReindex,
	PermissionSitemapGenerate,
	PermissionUserReportView,
	PermissionUserReportProcess,
	PermissionOperateLogView,
}

func FindByCode(code string) (PermissionDefinition, bool) {
	for _, permission := range Permissions {
		if permission.Code == code {
			return permission, true
		}
	}
	return PermissionDefinition{}, false
}
