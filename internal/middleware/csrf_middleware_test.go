package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bbs-go/internal/models/constants"

	"github.com/gin-gonic/gin"
)

func TestEnsureCSRFCookieUpgradesLegacyCookieSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	request := httptest.NewRequest(http.MethodGet, "/api/user/current", nil)
	request.AddCookie(&http.Cookie{Name: constants.CookieTokenKey, Value: "legacy-token"})
	ctx.Request = request

	ensureCSRFCookie(ctx)

	var csrfCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == constants.CookieCSRFTokenKey {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatal("expected a CSRF cookie for a legacy cookie session")
	}
	if csrfCookie.HttpOnly {
		t.Fatal("expected the CSRF cookie to be readable by browser JavaScript")
	}
}

func TestCSRFMiddlewareRejectsMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	request := httptest.NewRequest(http.MethodPost, "/api/admin/tag/list", strings.NewReader(""))
	request.AddCookie(&http.Cookie{Name: constants.CookieTokenKey, Value: "token"})
	request.AddCookie(&http.Cookie{Name: constants.CookieCSRFTokenKey, Value: "cookie-token"})
	ctx.Request = request

	CSRFMiddleware(ctx)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
	if !strings.Contains(response.Body.String(), "CSRF token is missing or invalid") {
		t.Fatalf("expected CSRF error response, got %s", response.Body.String())
	}
}
