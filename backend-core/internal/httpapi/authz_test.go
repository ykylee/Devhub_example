package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

func TestRoleMeetsMin(t *testing.T) {
	cases := []struct {
		actor string
		min   domain.AppRole
		want  bool
	}{
		{"system_admin", domain.AppRoleSystemAdmin, true},
		{"system_admin", domain.AppRoleManager, true},
		{"system_admin", domain.AppRoleDeveloper, true},
		{"manager", domain.AppRoleSystemAdmin, false},
		{"manager", domain.AppRoleManager, true},
		{"manager", domain.AppRoleDeveloper, true},
		{"developer", domain.AppRoleSystemAdmin, false},
		{"developer", domain.AppRoleManager, false},
		{"developer", domain.AppRoleDeveloper, true},
		{"", domain.AppRoleSystemAdmin, false},
		{"", domain.AppRoleManager, false},
		{"", domain.AppRoleDeveloper, false},
		{"unknown", domain.AppRoleSystemAdmin, false},
		{"unknown", domain.AppRoleManager, false},
		{"unknown", domain.AppRoleDeveloper, false},
	}
	for _, tc := range cases {
		if got := roleMeetsMin(tc.actor, tc.min); got != tc.want {
			t.Errorf("roleMeetsMin(%q, %q) = %v, want %v", tc.actor, tc.min, got, tc.want)
		}
	}
}

func TestRequireMinRoleBypassesWhenDevFallback(t *testing.T) {
	router := testRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore()})

	body := []byte(`{
		"user_id": "u-dev",
		"email": "dev@example.com",
		"display_name": "Dev User",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 (dev fallback bypasses role guard), got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequireMinRoleBlocksInsufficientRole(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "dev-user",
		Subject: "user-dev",
		Role:    "developer",
	}}
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/some-user", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for developer deleting users, got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log for role denial, got %d", len(audits.logs))
	}
	if audits.logs[0].Action != "auth.role_denied" {
		t.Fatalf("expected auth.role_denied audit, got %q", audits.logs[0].Action)
	}
	if audits.logs[0].Payload["required_role"] != "system_admin" {
		t.Fatalf("expected required_role=system_admin in audit payload, got %+v", audits.logs[0].Payload)
	}
}

func TestRequireMinRoleAllowsSufficientRole(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "manager-user",
		Subject: "user-manager",
		Role:    "manager",
	}}
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer manager-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for manager reading audit-logs, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouteRoleMatrix(t *testing.T) {
	type routeCase struct {
		method     string
		path       string
		body       []byte
		minRole    domain.AppRole
		paramRoles map[string]int
	}

	body := []byte(`{}`)
	routes := []routeCase{
		{http.MethodPost, "/api/v1/admin/service-actions", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodPost, "/api/v1/risks/r-1/mitigations", body, domain.AppRoleManager, nil},
		{http.MethodPost, "/api/v1/users", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodPatch, "/api/v1/users/u-1", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodDelete, "/api/v1/users/u-1", nil, domain.AppRoleSystemAdmin, nil},
		{http.MethodPost, "/api/v1/organization/units", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodPatch, "/api/v1/organization/units/unit-1", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodDelete, "/api/v1/organization/units/unit-1", nil, domain.AppRoleSystemAdmin, nil},
		{http.MethodPut, "/api/v1/organization/units/unit-1/members", body, domain.AppRoleSystemAdmin, nil},
		{http.MethodGet, "/api/v1/audit-logs", nil, domain.AppRoleManager, nil},
	}

	roles := []string{"developer", "manager", "system_admin"}
	for _, rc := range routes {
		for _, role := range roles {
			verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
				Login:   role + "-user",
				Subject: "user-" + role,
				Role:    role,
			}}
			router := NewRouter(RouterConfig{
				CommandStore:        &memoryCommandStore{},
				OrganizationStore:   newMemoryOrganizationStore(),
				AuditStore:          &memoryAuditStore{},
				BearerTokenVerifier: verifier,
			})

			var reqBody *bytes.Reader
			if rc.body != nil {
				reqBody = bytes.NewReader(rc.body)
			}
			var req *http.Request
			if reqBody != nil {
				req = httptest.NewRequest(rc.method, rc.path, reqBody)
			} else {
				req = httptest.NewRequest(rc.method, rc.path, nil)
			}
			req.Header.Set("Authorization", "Bearer t")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			expectAllowed := roleRank(role) >= roleRank(string(rc.minRole))
			if expectAllowed && rec.Code == http.StatusForbidden {
				t.Errorf("%s %s with role=%s: expected to pass role gate (min=%s), got 403", rc.method, rc.path, role, rc.minRole)
			}
			if !expectAllowed && rec.Code != http.StatusForbidden {
				t.Errorf("%s %s with role=%s: expected 403 (min=%s), got %d body=%s", rc.method, rc.path, role, rc.minRole, rec.Code, rec.Body.String())
			}
		}
	}
}
