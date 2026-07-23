package permissions

import "testing"

func TestAdminPermissionRegistryMatchesMethodAndPath(t *testing.T) {
	code, ok := GetAdminPermissionCode("POST", "/api/admin/topic/list")
	if !ok {
		t.Fatalf("expected topic list permission to be registered")
	}
	if code != PermissionTopicView.Code {
		t.Fatalf("expected %s, got %s", PermissionTopicView.Code, code)
	}
}

func TestAdminPermissionRegistryAllowsRoleOptionsFromUserUpdate(t *testing.T) {
	codes, ok := GetAdminPermissionCodes("GET", "/api/admin/role/roles")
	if !ok {
		t.Fatalf("expected role options permission to be registered")
	}
	expected := []string{PermissionRoleView.Code, PermissionUserUpdate.Code}
	if len(codes) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, codes)
	}
	for i, expectedCode := range expected {
		if codes[i] != expectedCode {
			t.Fatalf("expected %#v, got %#v", expected, codes)
		}
	}
}

func TestAdminPermissionRegistryAllowsEitherUserForbiddenPermission(t *testing.T) {
	codes, ok := GetAdminPermissionCodes("POST", "/api/admin/user/forbidden")
	if !ok {
		t.Fatalf("expected user forbidden permission to be registered")
	}
	expected := []string{PermissionUserForbidden.Code, PermissionUserForbiddenForever.Code}
	if len(codes) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, codes)
	}
	for i, expectedCode := range expected {
		if codes[i] != expectedCode {
			t.Fatalf("expected %#v, got %#v", expected, codes)
		}
	}
}

func TestAdminPermissionRegistryProtectsLinkDelete(t *testing.T) {
	code, ok := GetAdminPermissionCode("POST", "/api/admin/link/delete")
	if !ok {
		t.Fatalf("expected link delete permission to be registered")
	}
	if code != PermissionLinkDelete.Code {
		t.Fatalf("expected %s, got %s", PermissionLinkDelete.Code, code)
	}
}

func TestAdminPermissionRegistryProtectsLinkUpdateSort(t *testing.T) {
	code, ok := GetAdminPermissionCode("POST", "/api/admin/link/update_sort")
	if !ok {
		t.Fatalf("expected link sort permission to be registered")
	}
	if code != PermissionLinkUpdate.Code {
		t.Fatalf("expected %s, got %s", PermissionLinkUpdate.Code, code)
	}
}

func TestAdminPermissionRegistryProtectsUserReportProcess(t *testing.T) {
	code, ok := GetAdminPermissionCode("POST", "/api/admin/user-report/process")
	if !ok {
		t.Fatalf("expected user report process permission to be registered")
	}
	if code != PermissionUserReportProcess.Code {
		t.Fatalf("expected %s, got %s", PermissionUserReportProcess.Code, code)
	}
}

func TestAdminPermissionRegistryRejectsUnknownAdminPath(t *testing.T) {
	if code, ok := GetAdminPermissionCode("POST", "/api/admin/unknown/action"); ok {
		t.Fatalf("expected unknown admin path to be rejected, got %s", code)
	}
}

func TestAdminPermissionRegistryProtectsCommentManagementPaths(t *testing.T) {
	paths := []struct {
		method string
		path   string
	}{
		{method: "GET", path: "/api/admin/comment/1"},
		{method: "POST", path: "/api/admin/comment/list"},
		{method: "POST", path: "/api/admin/comment/delete"},
		{method: "DELETE", path: "/api/admin/comment/1"},
	}

	for _, path := range paths {
		if code, ok := GetAdminPermissionCode(path.method, path.path); !ok || code == "" {
			t.Fatalf("expected %s %s to be protected", path.method, path.path)
		}
	}
}

func TestAdminPermissionRegistryCoversEveryRegisteredAdminRoute(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/common/overview"},
		{"POST", "/api/admin/comment/list"}, {"POST", "/api/admin/comment/delete"}, {"GET", "/api/admin/comment/1"}, {"DELETE", "/api/admin/comment/1"},
		{"GET", "/api/admin/role/roles"}, {"GET", "/api/admin/role/permissions"}, {"GET", "/api/admin/role/1"},
		{"POST", "/api/admin/role/list"}, {"POST", "/api/admin/role/create"}, {"POST", "/api/admin/role/update"},
		{"POST", "/api/admin/role/update_permissions"}, {"POST", "/api/admin/role/delete"}, {"POST", "/api/admin/role/update_sort"},
		{"GET", "/api/admin/dict-type/1"}, {"POST", "/api/admin/dict-type/list"}, {"POST", "/api/admin/dict-type/create"},
		{"POST", "/api/admin/dict-type/update"}, {"POST", "/api/admin/dict-type/delete"},
		{"GET", "/api/admin/dict/1"}, {"GET", "/api/admin/dict/dicts"}, {"POST", "/api/admin/dict/list"},
		{"POST", "/api/admin/dict/create"}, {"POST", "/api/admin/dict/update"}, {"POST", "/api/admin/dict/delete"}, {"POST", "/api/admin/dict/update_sort"},
		{"GET", "/api/admin/user/1"}, {"GET", "/api/admin/user/synccount"}, {"POST", "/api/admin/user/list"},
		{"POST", "/api/admin/user/create"}, {"POST", "/api/admin/user/update"}, {"POST", "/api/admin/user/forbidden"}, {"POST", "/api/admin/user/batch/preview"}, {"POST", "/api/admin/user/batch"},
		{"POST", "/api/admin/user/update_password"}, {"POST", "/api/admin/user/reset_password"},
		{"GET", "/api/admin/tag/1"}, {"GET", "/api/admin/tag/autocomplete"}, {"GET", "/api/admin/tag/tags"},
		{"POST", "/api/admin/tag/list"}, {"POST", "/api/admin/tag/create"}, {"POST", "/api/admin/tag/update"},
		{"GET", "/api/admin/attachment/abc"}, {"POST", "/api/admin/attachment/list"}, {"POST", "/api/admin/attachment/delete"}, {"POST", "/api/admin/attachment/cleanup-orphans"},
		{"GET", "/api/admin/favorite/1"}, {"POST", "/api/admin/favorite/list"}, {"POST", "/api/admin/favorite/create"}, {"POST", "/api/admin/favorite/update"},
		{"GET", "/api/admin/topic/1"}, {"POST", "/api/admin/topic/list"}, {"POST", "/api/admin/topic/recommend"}, {"DELETE", "/api/admin/topic/recommend"}, {"POST", "/api/admin/topic/batch/preview"}, {"POST", "/api/admin/topic/batch"},
		{"POST", "/api/admin/topic/sticky"},
		{"POST", "/api/admin/topic/delete"}, {"POST", "/api/admin/topic/undelete"}, {"POST", "/api/admin/topic/accept_answer"}, {"POST", "/api/admin/topic/unaccept_answer"},
		{"POST", "/api/admin/topic/mark_solved"}, {"POST", "/api/admin/topic/mark_unsolved"},
		{"GET", "/api/admin/category/1"}, {"GET", "/api/admin/category/options"}, {"POST", "/api/admin/category/list"},
		{"POST", "/api/admin/category/create"}, {"POST", "/api/admin/category/update"}, {"POST", "/api/admin/category/update_sort"}, {"POST", "/api/admin/category/delete"},
		{"GET", "/api/admin/sys-config/1"}, {"GET", "/api/admin/sys-config/configs"}, {"POST", "/api/admin/sys-config/list"},
		{"POST", "/api/admin/sys-config/save"}, {"POST", "/api/admin/sys-config/save-sensitive"},
		{"GET", "/api/admin/search/reindex/status"}, {"POST", "/api/admin/search/reindex"}, {"GET", "/api/admin/seo/sitemap/status"}, {"POST", "/api/admin/seo/sitemap/generate"},
		{"GET", "/api/admin/link/1"}, {"POST", "/api/admin/link/list"}, {"POST", "/api/admin/link/create"}, {"POST", "/api/admin/link/update"}, {"POST", "/api/admin/link/delete"}, {"POST", "/api/admin/link/update_sort"},
		{"GET", "/api/admin/operate-log/1"}, {"POST", "/api/admin/operate-log/list"},
		{"GET", "/api/admin/message/1"}, {"POST", "/api/admin/message/list"}, {"POST", "/api/admin/message/create"}, {"POST", "/api/admin/message/update"}, {"POST", "/api/admin/message/delete"},
		{"GET", "/api/admin/health"}, {"GET", "/api/admin/tasks/status"},
		{"GET", "/api/admin/user-report/1"}, {"POST", "/api/admin/user-report/list"}, {"POST", "/api/admin/user-report/create"}, {"POST", "/api/admin/user-report/update"}, {"POST", "/api/admin/user-report/process"}, {"POST", "/api/admin/user-report/action"}, {"POST", "/api/admin/user-report/batch/preview"}, {"POST", "/api/admin/user-report/batch"},
		{"GET", "/api/admin/vote/1"}, {"POST", "/api/admin/vote/list"}, {"POST", "/api/admin/vote/create"}, {"POST", "/api/admin/vote/update"}, {"POST", "/api/admin/vote/delete"},
		{"GET", "/api/admin/vote-option/1"}, {"POST", "/api/admin/vote-option/list"}, {"POST", "/api/admin/vote-option/create"}, {"POST", "/api/admin/vote-option/update"}, {"POST", "/api/admin/vote-option/delete"},
		{"GET", "/api/admin/vote-record/1"}, {"POST", "/api/admin/vote-record/list"}, {"POST", "/api/admin/vote-record/create"}, {"POST", "/api/admin/vote-record/update"}, {"POST", "/api/admin/vote-record/delete"},
	}

	for _, route := range routes {
		if codes, ok := GetAdminPermissionCodes(route.method, route.path); !ok || len(codes) == 0 {
			t.Errorf("missing permission rule for %s %s", route.method, route.path)
		}
	}
}
