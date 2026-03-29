package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(middlewares...)
	r.GET("/test", func(c *gin.Context) {
		ac, ok := GetAuthContext(c)
		if !ok {
			c.JSON(200, gin.H{"authenticated": false})
			return
		}
		c.JSON(200, gin.H{
			"authenticated": true,
			"key_id":       ac.KeyID,
			"key_name":     ac.KeyName,
			"corp_name":    ac.CorpName,
			"app_name":     ac.AppName,
			"permissions":  ac.Permissions,
			"is_admin":     ac.IsAdmin,
		})
	})
	return r
}

// mockAuthenticator implements Authenticator for testing
type mockAuthenticator struct {
	authCtx *AuthContext
	authErr error
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, rawKey string) (*AuthContext, error) {
	if m.authErr != nil {
		return nil, m.authErr
	}
	return m.authCtx, nil
}

func TestGinMiddleware_MissingAuthHeader(t *testing.T) {
	auth := &mockAuthenticator{}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(40101) {
		t.Errorf("expected code 40101, got %v", resp["code"])
	}
}

func TestGinMiddleware_InvalidAuthFormat(t *testing.T) {
	auth := &mockAuthenticator{}
	r := setupRouter(GinMiddleware(auth))

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "Token wgk_test"},
		{"bearer only", "Bearer"},
		{"wrong prefix", "Basic abc123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", w.Code)
			}
		})
	}
}

func TestGinMiddleware_EmptyAPIKey(t *testing.T) {
	auth := &mockAuthenticator{}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer  ")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGinMiddleware_ValidAuth(t *testing.T) {
	auth := &mockAuthenticator{
		authCtx: &AuthContext{
			KeyID:       "key-1",
			KeyName:     "TestKey",
			Permissions: []string{"calendar:read"},
			CorpName:    "testcorp",
			AppName:     "oa",
		},
	}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_valid_key")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["authenticated"] != true {
		t.Error("expected authenticated true")
	}
	if resp["key_id"] != "key-1" {
		t.Errorf("expected key_id key-1, got %v", resp["key_id"])
	}
}

func TestGinMiddleware_InvalidAPIKey(t *testing.T) {
	auth := &mockAuthenticator{authErr: ErrInvalidAPIKey}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_invalid")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGinMiddleware_DisabledAPIKey(t *testing.T) {
	auth := &mockAuthenticator{authErr: ErrAPIKeyDisabled}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_disabled")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(40301) {
		t.Errorf("expected code 40301, got %v", resp["code"])
	}
}

func TestGinMiddleware_ExpiredAPIKey(t *testing.T) {
	auth := &mockAuthenticator{authErr: ErrAPIKeyExpired}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_expired")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGinMiddleware_UnknownError(t *testing.T) {
	auth := &mockAuthenticator{authErr: ErrPermissionDenied}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_unknown")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequirePermission_NoAuthContext(t *testing.T) {
	r := gin.New()
	r.Use(RequirePermission("calendar:read"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequirePermission_InvalidAuthContext(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, "not an auth context")
		c.Next()
	})
	r.Use(RequirePermission("calendar:read"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			Permissions: []string{"message:send"},
		})
		c.Next()
	})
	r.Use(RequirePermission("calendar:read"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequirePermission_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			Permissions: []string{"calendar:read"},
		})
		c.Next()
	})
	r.Use(RequirePermission("calendar:read"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequirePermission_AdminWildcard(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			Permissions: []string{"*"},
			IsAdmin:     true,
		})
		c.Next()
	})
	r.Use(RequirePermission("any:permission"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAnyPermission_NoAuthContext(t *testing.T) {
	r := gin.New()
	r.Use(RequireAnyPermission("calendar:read", "message:send"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAnyPermission_NoneMatch(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			Permissions: []string{"other:perm"},
		})
		c.Next()
	})
	r.Use(RequireAnyPermission("calendar:read", "message:send"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAnyPermission_OneMatch(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			Permissions: []string{"message:send"},
		})
		c.Next()
	})
	r.Use(RequireAnyPermission("calendar:read", "message:send"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAnyPermission_Admin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			IsAdmin: true,
		})
		c.Next()
	})
	r.Use(RequireAnyPermission("calendar:read", "message:send"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAdmin_NoAuthContext(t *testing.T) {
	r := gin.New()
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAdmin_NotAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			IsAdmin: false,
		})
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "admin privileges required" {
		t.Errorf("expected 'admin privileges required', got %v", resp["message"])
	}
}

func TestRequireAdmin_IsAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, &AuthContext{
			IsAdmin: true,
		})
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetAuthContext_NotSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ac, ok := GetAuthContext(c)
	if ok {
		t.Error("expected ok to be false")
	}
	if ac != nil {
		t.Error("expected ac to be nil")
	}
}

func TestGetAuthContext_InvalidType(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(ContextKey, "not an auth context")
	ac, ok := GetAuthContext(c)
	if ok {
		t.Error("expected ok to be false")
	}
	if ac != nil {
		t.Error("expected ac to be nil")
	}
}

func TestGetAuthContext_Valid(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	expected := &AuthContext{KeyID: "key-1", KeyName: "Test"}
	c.Set(ContextKey, expected)

	ac, ok := GetAuthContext(c)
	if !ok {
		t.Error("expected ok to be true")
	}
	if ac.KeyID != "key-1" {
		t.Errorf("expected KeyID key-1, got %s", ac.KeyID)
	}
}

func TestHasPermission_NonAdmin(t *testing.T) {
	ac := &AuthContext{
		Permissions: []string{"calendar:read", "message:send"},
		IsAdmin:     false,
	}

	if !ac.HasPermission("calendar:read") {
		t.Error("expected to have calendar:read permission")
	}
	if !ac.HasPermission("message:send") {
		t.Error("expected to have message:send permission")
	}
	if ac.HasPermission("admin:manage") {
		t.Error("expected not to have admin:manage permission")
	}
}

func TestHasPermission_WildcardInPermissions(t *testing.T) {
	ac := &AuthContext{
		Permissions: []string{"*"},
		IsAdmin:     false,
	}

	if !ac.HasPermission("anything:here") {
		t.Error("wildcard permission should grant all permissions")
	}
}

func TestHasAnyPermission(t *testing.T) {
	ac := &AuthContext{
		Permissions: []string{"calendar:read"},
		IsAdmin:     false,
	}

	if !ac.HasAnyPermission([]string{"calendar:read", "message:send"}) {
		t.Error("expected to have at least one matching permission")
	}
	if ac.HasAnyPermission([]string{"admin:manage", "other:perm"}) {
		t.Error("expected no matching permissions")
	}
}

func TestGinMiddleware_BearerPrefixCaseInsensitive(t *testing.T) {
	auth := &mockAuthenticator{
		authCtx: &AuthContext{KeyID: "key-1"},
	}
	r := setupRouter(GinMiddleware(auth))

	// Test "bearer" (lowercase)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer wgk_test")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for lowercase bearer, got %d", w.Code)
	}

	// Test "BEARER" (uppercase)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "BEARER wgk_test")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for uppercase bearer, got %d", w.Code)
	}
}

func TestGinMiddleware_AuthContextStored(t *testing.T) {
	expectedCtx := &AuthContext{
		KeyID:       "key-abc",
		KeyName:     "MyKey",
		Permissions: []string{"calendar:read", "calendar:write"},
		CorpName:    "testcorp",
		AppName:     "oa",
		IsAdmin:     false,
	}
	auth := &mockAuthenticator{authCtx: expectedCtx}
	r := setupRouter(GinMiddleware(auth))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wgk_test")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["key_name"] != "MyKey" {
		t.Errorf("expected key_name MyKey, got %v", resp["key_name"])
	}
	if resp["corp_name"] != "testcorp" {
		t.Errorf("expected corp_name testcorp, got %v", resp["corp_name"])
	}

	perms, ok := resp["permissions"].([]interface{})
	if !ok {
		t.Fatalf("expected permissions to be array, got %T", resp["permissions"])
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(perms))
	}
	if !strings.Contains(perms[0].(string), "calendar") {
		t.Errorf("expected calendar permission, got %v", perms[0])
	}
}

func TestRequirePermission_InvalidTypeInContext(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKey, "string value")
		c.Next()
	})
	r.Use(RequirePermission("any"))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
