package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestPermissionCache_DefaultsToSystemRolesWhenStoreNil(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	cases := []struct {
		role     string
		resource domain.Resource
		action   domain.Action
		want     bool
	}{
		{"developer", domain.ResourceSecurity, domain.ActionView, true},
		{"developer", domain.ResourceSecurity, domain.ActionCreate, false},
		{"developer", domain.ResourceAudit, domain.ActionView, false},
		{"manager", domain.ResourceSecurity, domain.ActionCreate, true},
		{"manager", domain.ResourceAudit, domain.ActionView, true},
		{"manager", domain.ResourceOrganization, domain.ActionEdit, false},
		{"system_admin", domain.ResourceOrganization, domain.ActionDelete, true},
		{"system_admin", domain.ResourceAudit, domain.ActionCreate, false}, // invariant
		{"unknown-role", domain.ResourceInfrastructure, domain.ActionView, false},
	}
	for _, tc := range cases {
		got, err := cache.Allows(ctx, tc.role, tc.resource, tc.action)
		if err != nil {
			t.Errorf("Allows(%s, %s, %s) err: %v", tc.role, tc.resource, tc.action, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Allows(%s, %s, %s) = %v, want %v", tc.role, tc.resource, tc.action, got, tc.want)
		}
	}
}

func TestPermissionCache_InvalidateReloadsFromStore(t *testing.T) {
	store := newFakeRBACStore()
	cache := NewPermissionCache(store)
	ctx := context.Background()

	if got, _ := cache.Allows(ctx, "developer", domain.ResourceAudit, domain.ActionView); got {
		t.Fatal("developer should not have audit:view by default")
	}

	dev := store.roles["developer"]
	dev.Permissions[domain.ResourceAudit] = domain.ResourcePermissions{View: true}
	store.roles["developer"] = dev

	if got, _ := cache.Allows(ctx, "developer", domain.ResourceAudit, domain.ActionView); got {
		t.Fatal("cache should serve stale value before invalidation")
	}

	cache.Invalidate()

	if got, _ := cache.Allows(ctx, "developer", domain.ResourceAudit, domain.ActionView); !got {
		t.Fatal("after invalidate, developer should now have audit:view")
	}
}

func TestPermissionCache_LoadError(t *testing.T) {
	store := newFakeRBACStore()
	store.listErr = errors.New("db down")
	cache := NewPermissionCache(store)
	if _, err := cache.Allows(context.Background(), "developer", domain.ResourceSecurity, domain.ActionView); err == nil {
		t.Fatal("expected load error from Allows")
	}
}

func TestRoutePermissionTable_CoversAllProtectedV1Routes(t *testing.T) {
	router := NewRouter(RouterConfig{})

	bypassRoot := map[string]struct{}{
		"GET /health": {},
	}

	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, isRoot := bypassRoot[key]; isRoot {
			continue
		}
		if !startsWith(route.Path, "/api/v1/") {
			continue
		}
		if _, ok := lookupRoutePolicy(route.Method, route.Path); !ok {
			t.Errorf("route %s is registered but has no entry in routePermissionTable", key)
		}
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func TestEnforceRoutePermission_DenyByDefaultUnmappedRoute(t *testing.T) {
	// Build a router manually so the orphan route lives inside the v1 group
	// alongside the enforceRoutePermission middleware. NewRouter does not
	// expose the v1 group, so we cannot retrofit the middleware onto a new
	// orphan route after the fact.
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "dev-user",
		Subject: "user-dev",
		Role:    "developer",
	}}
	audits := &memoryAuditStore{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := RouterConfig{
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	}
	cfg.PermissionCache = NewPermissionCache(cfg.RBACStore)
	handler := Handler{cfg: cfg}
	v1 := router.Group("/api/v1")
	v1.Use(handler.authenticateActor)
	v1.Use(handler.enforceRoutePermission)
	v1.GET("/orphan", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orphan", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("unmapped route code = %d, want 403", rec.Code)
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "auth_policy_unmapped" {
		t.Errorf("code = %q, want auth_policy_unmapped", resp.Code)
	}

	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.policy_unmapped" {
		t.Errorf("expected one auth.policy_unmapped audit, got %+v", audits.logs)
	}
}

func TestEnforceRoutePermission_RoleAllowedAndDenied(t *testing.T) {
	cases := []struct {
		name      string
		role      string
		method    string
		path      string
		wantDenied bool
	}{
		{"developer cannot delete users", "developer", http.MethodDelete, "/api/v1/users/u-1", true},
		{"manager creates mitigation gate passes", "manager", http.MethodPost, "/api/v1/risks/r-1/mitigations", false},
		{"manager cannot create service-action", "manager", http.MethodPost, "/api/v1/admin/service-actions", true},
		{"developer cannot view audit-logs", "developer", http.MethodGet, "/api/v1/audit-logs", true},
		{"developer can view risks", "developer", http.MethodGet, "/api/v1/risks", false},
		{"system_admin delete users gate passes", "system_admin", http.MethodDelete, "/api/v1/users/u-1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
				Login:   tc.role + "-user",
				Subject: "user-" + tc.role,
				Role:    tc.role,
			}}
			router := NewRouter(RouterConfig{
				CommandStore:        &memoryCommandStore{},
				OrganizationStore:   newMemoryOrganizationStore(),
				AuditStore:          &memoryAuditStore{},
				BearerTokenVerifier: verifier,
			})
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer t")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gateDenied := rec.Code == http.StatusForbidden
			if tc.wantDenied && !gateDenied {
				t.Errorf("expected 403 (RBAC gate denial), got %d body=%s", rec.Code, rec.Body.String())
			}
			if !tc.wantDenied && gateDenied {
				t.Errorf("expected gate to pass, got 403 body=%s", rec.Body.String())
			}
		})
	}
}

func TestEnforceRoutePermission_BypassesMeAndWebhook(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "dev-user",
		Subject: "user-dev",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusForbidden {
		t.Errorf("/api/v1/me should bypass RBAC gate (got 403 body=%s)", rec.Body.String())
	}
}

// --- ADR-0011 §4.2 enforceRowOwnership ---
//
// helper 의 세 allow 규칙(system_admin / allowedRoles / owner-self) 과 deny
// 시의 audit + 403 envelope 을 검증한다. handler 단독 unit 호출이므로
// gin.CreateTestContext 로 컨텍스트만 만들고 c.Set 으로 actor 를 주입한다.

func newOwnershipTestContext(t *testing.T, login, role string) (*gin.Context, *memoryAuditStore, *httptest.ResponseRecorder, Handler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	audits := &memoryAuditStore{}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/applications/app-1", nil)
	c.Set("devhub_actor_login", login)
	c.Set("devhub_actor_role", role)
	handler := Handler{cfg: RouterConfig{AuditStore: audits}}
	return c, audits, rec, handler
}

func TestEnforceRowOwnership_SystemAdminAllowed(t *testing.T) {
	c, audits, rec, h := newOwnershipTestContext(t, "charlie", "system_admin")

	if got := h.enforceRowOwnership(c, "alice", "pmo_manager"); !got {
		t.Fatal("system_admin should always be allowed")
	}
	if rec.Code != 0 && rec.Code != http.StatusOK {
		t.Errorf("expected no abort, got status=%d", rec.Code)
	}
	if c.IsAborted() {
		t.Error("expected context not aborted on allow")
	}
	if len(audits.logs) != 0 {
		t.Errorf("expected no audit on allow, got %+v", audits.logs)
	}
}

func TestEnforceRowOwnership_AllowedRoleWhitelist(t *testing.T) {
	c, audits, _, h := newOwnershipTestContext(t, "bob", "pmo_manager")

	if got := h.enforceRowOwnership(c, "alice", "pmo_manager"); !got {
		t.Fatal("pmo_manager in allowedRoles should be allowed")
	}
	if c.IsAborted() {
		t.Error("expected context not aborted on allow")
	}
	if len(audits.logs) != 0 {
		t.Errorf("expected no audit on allow, got %+v", audits.logs)
	}
}

func TestEnforceRowOwnership_OwnerSelfAllowed(t *testing.T) {
	c, _, _, h := newOwnershipTestContext(t, "alice", "developer")

	if got := h.enforceRowOwnership(c, "alice"); !got {
		t.Fatal("owner-self should be allowed even without allowedRoles")
	}
	if c.IsAborted() {
		t.Error("expected context not aborted on owner-self allow")
	}
}

func TestEnforceRowOwnership_DeniedEmitsAuditAndForbidden(t *testing.T) {
	c, audits, rec, h := newOwnershipTestContext(t, "bob", "developer")

	if got := h.enforceRowOwnership(c, "alice", "pmo_manager"); got {
		t.Fatal("non-owner non-allowed role should be denied")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !c.IsAborted() {
		t.Error("expected context aborted on deny")
	}
	var resp struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Code   string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "auth_row_denied" {
		t.Errorf("envelope code = %q, want auth_row_denied", resp.Code)
	}
	if resp.Status != "forbidden" {
		t.Errorf("envelope status = %q, want forbidden", resp.Status)
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected exactly one audit entry, got %d: %+v", len(audits.logs), audits.logs)
	}
	got := audits.logs[0]
	if got.Action != "auth.row_denied" {
		t.Errorf("action = %q, want auth.row_denied", got.Action)
	}
	// payload 키 검증.
	payload := got.Payload
	if payload["denied_reason"] != "owner_mismatch" {
		t.Errorf("payload.denied_reason = %v, want owner_mismatch", payload["denied_reason"])
	}
	if payload["actor_role"] != "developer" {
		t.Errorf("payload.actor_role = %v, want developer", payload["actor_role"])
	}
	if payload["owner_user_id"] != "alice" {
		t.Errorf("payload.owner_user_id = %v, want alice", payload["owner_user_id"])
	}
}

func TestEnforceRowOwnership_EmptyOwnerDisablesSelfRule(t *testing.T) {
	// ownerUserID 가 "" 일 때 actor 가 "" 라고 해서 owner-self 가 통과되면 안 됨.
	c, _, rec, h := newOwnershipTestContext(t, "", "developer")

	if got := h.enforceRowOwnership(c, ""); got {
		t.Fatal("empty ownerUserID should not match empty actor login (rule disabled)")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestEnforceRowOwnership_DevFallbackBypasses(t *testing.T) {
	// AuthDevFallback=true 환경에서는 actor 가 컨텍스트에 없어도 통과해야 한다.
	// enforceRoutePermission 과 동일한 정책 (sprint claude/work_260515-d 도입).
	gin.SetMode(gin.TestMode)
	audits := &memoryAuditStore{}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/applications/app-1", nil)
	c.Set("devhub_auth_dev_fallback", true)
	h := Handler{cfg: RouterConfig{AuditStore: audits}}

	if got := h.enforceRowOwnership(c, "alice", "pmo_manager"); !got {
		t.Fatal("dev fallback should bypass and return true")
	}
	if c.IsAborted() {
		t.Error("dev fallback should not abort")
	}
	if len(audits.logs) != 0 {
		t.Errorf("dev fallback should not emit audit, got %+v", audits.logs)
	}
}

func TestEnforceRowOwnership_NoActorContextDenied(t *testing.T) {
	// actor 키가 컨텍스트에 없으면 (auth middleware 누락) deny 가 안전한 기본 동작.
	gin.SetMode(gin.TestMode)
	audits := &memoryAuditStore{}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/applications/app-1", nil)
	h := Handler{cfg: RouterConfig{AuditStore: audits}}

	if got := h.enforceRowOwnership(c, "alice", "pmo_manager"); got {
		t.Fatal("missing actor context should deny")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
