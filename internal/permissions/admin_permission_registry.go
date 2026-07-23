package permissions

import (
	"bbs-go/internal/pkg/urls"
	"strings"
)

type adminPermissionRule struct {
	Method      string
	Pattern     string
	Permissions []PermissionDefinition
}

var (
	adminPathMatcher = urls.NewAntPathMatcher()

	adminPermissionRules = []adminPermissionRule{
		{Method: "GET", Pattern: "/api/admin/common/**", Permissions: []PermissionDefinition{PermissionDashboardView}},

		{Method: "GET", Pattern: "/api/admin/comment/*", Permissions: []PermissionDefinition{PermissionCommentView}},
		{Method: "POST", Pattern: "/api/admin/comment/list", Permissions: []PermissionDefinition{PermissionCommentView}},
		{Method: "POST", Pattern: "/api/admin/comment/delete", Permissions: []PermissionDefinition{PermissionCommentDelete}},
		{Method: "DELETE", Pattern: "/api/admin/comment/*", Permissions: []PermissionDefinition{PermissionCommentDelete}},

		{Method: "GET", Pattern: "/api/admin/topic/*", Permissions: []PermissionDefinition{PermissionTopicView}},
		{Method: "POST", Pattern: "/api/admin/topic/list", Permissions: []PermissionDefinition{PermissionTopicView}},
		{Method: "POST", Pattern: "/api/admin/topic/recommend", Permissions: []PermissionDefinition{PermissionTopicRecommend}},
		{Method: "DELETE", Pattern: "/api/admin/topic/recommend", Permissions: []PermissionDefinition{PermissionTopicRecommend}},
		{Method: "POST", Pattern: "/api/admin/topic/sticky", Permissions: []PermissionDefinition{PermissionTopicSticky}},
		{Method: "POST", Pattern: "/api/admin/topic/delete", Permissions: []PermissionDefinition{PermissionTopicDelete}},
		{Method: "POST", Pattern: "/api/admin/topic/undelete", Permissions: []PermissionDefinition{PermissionTopicDelete}},
		{Method: "POST", Pattern: "/api/admin/topic/accept_answer", Permissions: []PermissionDefinition{PermissionTopicAcceptAnswer}},
		{Method: "POST", Pattern: "/api/admin/topic/unaccept_answer", Permissions: []PermissionDefinition{PermissionTopicAcceptAnswer}},
		{Method: "POST", Pattern: "/api/admin/topic/mark_solved", Permissions: []PermissionDefinition{PermissionTopicSolve}},
		{Method: "POST", Pattern: "/api/admin/topic/mark_unsolved", Permissions: []PermissionDefinition{PermissionTopicSolve}},

		{Method: "GET", Pattern: "/api/admin/category/*", Permissions: []PermissionDefinition{PermissionCategoryView}},
		{Method: "GET", Pattern: "/api/admin/category/options", Permissions: []PermissionDefinition{PermissionCategoryView}},
		{Method: "POST", Pattern: "/api/admin/category/list", Permissions: []PermissionDefinition{PermissionCategoryView}},
		{Method: "POST", Pattern: "/api/admin/category/create", Permissions: []PermissionDefinition{PermissionCategoryCreate}},
		{Method: "POST", Pattern: "/api/admin/category/update", Permissions: []PermissionDefinition{PermissionCategoryUpdate}},
		{Method: "POST", Pattern: "/api/admin/category/delete", Permissions: []PermissionDefinition{PermissionCategoryDelete}},
		{Method: "POST", Pattern: "/api/admin/category/update_sort", Permissions: []PermissionDefinition{PermissionCategorySort}},

		{Method: "GET", Pattern: "/api/admin/link/*", Permissions: []PermissionDefinition{PermissionLinkView}},
		{Method: "POST", Pattern: "/api/admin/link/list", Permissions: []PermissionDefinition{PermissionLinkView}},
		{Method: "POST", Pattern: "/api/admin/link/create", Permissions: []PermissionDefinition{PermissionLinkCreate}},
		{Method: "POST", Pattern: "/api/admin/link/update", Permissions: []PermissionDefinition{PermissionLinkUpdate}},
		{Method: "POST", Pattern: "/api/admin/link/delete", Permissions: []PermissionDefinition{PermissionLinkDelete}},
		{Method: "POST", Pattern: "/api/admin/link/update_sort", Permissions: []PermissionDefinition{PermissionLinkUpdate}},

		{Method: "GET", Pattern: "/api/admin/tag/*", Permissions: []PermissionDefinition{PermissionTagView}},
		{Method: "POST", Pattern: "/api/admin/tag/list", Permissions: []PermissionDefinition{PermissionTagView}},
		{Method: "POST", Pattern: "/api/admin/tag/create", Permissions: []PermissionDefinition{PermissionTagCreate}},
		{Method: "POST", Pattern: "/api/admin/tag/update", Permissions: []PermissionDefinition{PermissionTagUpdate}},

		{Method: "GET", Pattern: "/api/admin/dict-type/*", Permissions: []PermissionDefinition{PermissionDictTypeView}},
		{Method: "GET", Pattern: "/api/admin/dict/*", Permissions: []PermissionDefinition{PermissionDictView}},
		{Method: "POST", Pattern: "/api/admin/dict-type/list", Permissions: []PermissionDefinition{PermissionDictTypeView}},
		{Method: "POST", Pattern: "/api/admin/dict-type/create", Permissions: []PermissionDefinition{PermissionDictTypeCreate}},
		{Method: "POST", Pattern: "/api/admin/dict-type/update", Permissions: []PermissionDefinition{PermissionDictTypeUpdate}},
		{Method: "POST", Pattern: "/api/admin/dict-type/delete", Permissions: []PermissionDefinition{PermissionDictTypeDelete}},
		{Method: "POST", Pattern: "/api/admin/dict/list", Permissions: []PermissionDefinition{PermissionDictView}},
		{Method: "POST", Pattern: "/api/admin/dict/create", Permissions: []PermissionDefinition{PermissionDictCreate}},
		{Method: "POST", Pattern: "/api/admin/dict/update", Permissions: []PermissionDefinition{PermissionDictUpdate}},
		{Method: "POST", Pattern: "/api/admin/dict/delete", Permissions: []PermissionDefinition{PermissionDictDelete}},
		{Method: "POST", Pattern: "/api/admin/dict/update_sort", Permissions: []PermissionDefinition{PermissionDictSort}},
		{Method: "GET", Pattern: "/api/admin/dict/dicts", Permissions: []PermissionDefinition{PermissionDictView}},

		{Method: "GET", Pattern: "/api/admin/attachment/*", Permissions: []PermissionDefinition{PermissionAttachmentView}},
		{Method: "POST", Pattern: "/api/admin/attachment/list", Permissions: []PermissionDefinition{PermissionAttachmentView}},
		{Method: "POST", Pattern: "/api/admin/attachment/delete", Permissions: []PermissionDefinition{PermissionAttachmentDelete}},
		{Method: "POST", Pattern: "/api/admin/attachment/cleanup-orphans", Permissions: []PermissionDefinition{PermissionAttachmentCleanup}},

		{Method: "GET", Pattern: "/api/admin/user/*", Permissions: []PermissionDefinition{PermissionUserView}},
		{Method: "GET", Pattern: "/api/admin/user/synccount", Permissions: []PermissionDefinition{PermissionUserUpdate}},
		{Method: "POST", Pattern: "/api/admin/user/list", Permissions: []PermissionDefinition{PermissionUserView}},
		{Method: "POST", Pattern: "/api/admin/user/create", Permissions: []PermissionDefinition{PermissionUserCreate}},
		{Method: "POST", Pattern: "/api/admin/user/update", Permissions: []PermissionDefinition{PermissionUserUpdate}},
		{Method: "POST", Pattern: "/api/admin/user/forbidden", Permissions: []PermissionDefinition{PermissionUserForbidden, PermissionUserForbiddenForever}},
		{Method: "POST", Pattern: "/api/admin/user/update_password", Permissions: []PermissionDefinition{PermissionUserUpdatePassword}},
		{Method: "POST", Pattern: "/api/admin/user/reset_password", Permissions: []PermissionDefinition{PermissionUserResetPassword}},

		{Method: "GET", Pattern: "/api/admin/role/roles", Permissions: []PermissionDefinition{PermissionRoleView, PermissionUserUpdate}},
		{Method: "GET", Pattern: "/api/admin/role/*", Permissions: []PermissionDefinition{PermissionRoleView}},
		{Method: "POST", Pattern: "/api/admin/role/list", Permissions: []PermissionDefinition{PermissionRoleView}},
		{Method: "POST", Pattern: "/api/admin/role/create", Permissions: []PermissionDefinition{PermissionRoleCreate}},
		{Method: "POST", Pattern: "/api/admin/role/update", Permissions: []PermissionDefinition{PermissionRoleUpdate}},
		{Method: "POST", Pattern: "/api/admin/role/delete", Permissions: []PermissionDefinition{PermissionRoleDelete}},
		{Method: "POST", Pattern: "/api/admin/role/update_sort", Permissions: []PermissionDefinition{PermissionRoleSort}},
		{Method: "POST", Pattern: "/api/admin/role/update_permissions", Permissions: []PermissionDefinition{PermissionRolePermissionUpdate}},

		{Method: "GET", Pattern: "/api/admin/sys-config/**", Permissions: []PermissionDefinition{PermissionSettingView}},
		{Method: "POST", Pattern: "/api/admin/sys-config/list", Permissions: []PermissionDefinition{PermissionSettingView}},
		{Method: "POST", Pattern: "/api/admin/sys-config/save", Permissions: []PermissionDefinition{PermissionSettingUpdate}},
		{Method: "POST", Pattern: "/api/admin/sys-config/save-sensitive", Permissions: []PermissionDefinition{PermissionSettingSensitiveUpdate}},
		{Method: "GET", Pattern: "/api/admin/search/reindex/status", Permissions: []PermissionDefinition{PermissionSearchReindex}},
		{Method: "POST", Pattern: "/api/admin/search/reindex", Permissions: []PermissionDefinition{PermissionSearchReindex}},
		{Method: "GET", Pattern: "/api/admin/seo/sitemap/status", Permissions: []PermissionDefinition{PermissionSitemapGenerate}},
		{Method: "POST", Pattern: "/api/admin/seo/sitemap/generate", Permissions: []PermissionDefinition{PermissionSitemapGenerate}},

		{Method: "GET", Pattern: "/api/admin/user-report/*", Permissions: []PermissionDefinition{PermissionUserReportView}},
		{Method: "POST", Pattern: "/api/admin/user-report/list", Permissions: []PermissionDefinition{PermissionUserReportView}},
		{Method: "POST", Pattern: "/api/admin/user-report/create", Permissions: []PermissionDefinition{PermissionUserReportUpdate}},
		{Method: "POST", Pattern: "/api/admin/user-report/update", Permissions: []PermissionDefinition{PermissionUserReportUpdate}},
		{Method: "POST", Pattern: "/api/admin/user-report/process", Permissions: []PermissionDefinition{PermissionUserReportProcess}},
		{Method: "POST", Pattern: "/api/admin/user-report/action", Permissions: []PermissionDefinition{PermissionUserReportProcess}},
		{Method: "GET", Pattern: "/api/admin/operate-log/*", Permissions: []PermissionDefinition{PermissionOperateLogView}},
		{Method: "POST", Pattern: "/api/admin/operate-log/list", Permissions: []PermissionDefinition{PermissionOperateLogView}},
		{Method: "GET", Pattern: "/api/admin/message/*", Permissions: []PermissionDefinition{PermissionMessageView}},
		{Method: "POST", Pattern: "/api/admin/message/list", Permissions: []PermissionDefinition{PermissionMessageView}},
		{Method: "POST", Pattern: "/api/admin/message/create", Permissions: []PermissionDefinition{PermissionMessageSend}},
		{Method: "POST", Pattern: "/api/admin/message/update", Permissions: []PermissionDefinition{PermissionMessageSend}},
		{Method: "POST", Pattern: "/api/admin/message/delete", Permissions: []PermissionDefinition{PermissionMessageDelete}},
		{Method: "GET", Pattern: "/api/admin/health", Permissions: []PermissionDefinition{PermissionHealthView}},
		{Method: "GET", Pattern: "/api/admin/tasks/**", Permissions: []PermissionDefinition{PermissionHealthView, PermissionTaskView}},

		{Method: "GET", Pattern: "/api/admin/favorite/*", Permissions: []PermissionDefinition{PermissionFavoriteView}},
		{Method: "POST", Pattern: "/api/admin/favorite/list", Permissions: []PermissionDefinition{PermissionFavoriteView}},
		{Method: "POST", Pattern: "/api/admin/favorite/create", Permissions: []PermissionDefinition{PermissionFavoriteManage}},
		{Method: "POST", Pattern: "/api/admin/favorite/update", Permissions: []PermissionDefinition{PermissionFavoriteManage}},

		{Method: "GET", Pattern: "/api/admin/vote/*", Permissions: []PermissionDefinition{PermissionVoteView}},
		{Method: "GET", Pattern: "/api/admin/vote-option/*", Permissions: []PermissionDefinition{PermissionVoteOptionView}},
		{Method: "GET", Pattern: "/api/admin/vote-record/*", Permissions: []PermissionDefinition{PermissionVoteRecordView}},
		{Method: "POST", Pattern: "/api/admin/vote/list", Permissions: []PermissionDefinition{PermissionVoteView}},
		{Method: "POST", Pattern: "/api/admin/vote/create", Permissions: []PermissionDefinition{PermissionVoteCreate}},
		{Method: "POST", Pattern: "/api/admin/vote/update", Permissions: []PermissionDefinition{PermissionVoteUpdate}},
		{Method: "POST", Pattern: "/api/admin/vote/delete", Permissions: []PermissionDefinition{PermissionVoteDelete}},
		{Method: "POST", Pattern: "/api/admin/vote-option/list", Permissions: []PermissionDefinition{PermissionVoteOptionView}},
		{Method: "POST", Pattern: "/api/admin/vote-option/create", Permissions: []PermissionDefinition{PermissionVoteOptionCreate}},
		{Method: "POST", Pattern: "/api/admin/vote-option/update", Permissions: []PermissionDefinition{PermissionVoteOptionUpdate}},
		{Method: "POST", Pattern: "/api/admin/vote-option/delete", Permissions: []PermissionDefinition{PermissionVoteOptionDelete}},
		{Method: "POST", Pattern: "/api/admin/vote-record/list", Permissions: []PermissionDefinition{PermissionVoteRecordView}},
		{Method: "POST", Pattern: "/api/admin/vote-record/create", Permissions: []PermissionDefinition{PermissionVoteRecordManage}},
		{Method: "POST", Pattern: "/api/admin/vote-record/update", Permissions: []PermissionDefinition{PermissionVoteRecordManage}},
		{Method: "POST", Pattern: "/api/admin/vote-record/delete", Permissions: []PermissionDefinition{PermissionVoteRecordManage}},
	}
)

func GetAdminPermissionCode(method, path string) (string, bool) {
	codes, ok := GetAdminPermissionCodes(method, path)
	if !ok || len(codes) == 0 {
		return "", false
	}
	return codes[0], true
}

func GetAdminPermissionCodes(method, path string) ([]string, bool) {
	method = strings.ToUpper(method)
	for _, rule := range adminPermissionRules {
		if rule.Method == method && adminPathMatcher.Match(rule.Pattern, path) {
			codes := make([]string, 0, len(rule.Permissions))
			for _, permission := range rule.Permissions {
				codes = append(codes, permission.Code)
			}
			return codes, true
		}
	}
	return nil, false
}
