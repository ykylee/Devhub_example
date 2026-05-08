package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

type memoryRBACPolicyStore struct {
	policy domain.RBACPolicy
	input  domain.ReplaceRBACPolicyInput
}

func (s *memoryRBACPolicyStore) GetActiveRBACPolicy(_ context.Context) (domain.RBACPolicy, error) {
	if s.policy.PolicyVersion == "" {
		return domain.DefaultRBACPolicy(), nil
	}
	return s.policy, nil
}

func (s *memoryRBACPolicyStore) ReplaceRBACPolicy(_ context.Context, input domain.ReplaceRBACPolicyInput) (domain.RBACPolicy, error) {
	s.input = input
	policy := input.Policy
	if policy.PolicyVersion == "" {
		policy.PolicyVersion = "rbac_test"
	}
	policy.Source = "memory"
	policy.Editable = true
	s.policy = policy
	return policy, nil
}

func TestGetRBACPolicyReturnsDefaultMatrix(t *testing.T) {
	router := NewRouter(RouterConfig{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policy", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Roles []struct {
				Role  string `json:"role"`
				Label string `json:"label"`
			} `json:"roles"`
			Resources []struct {
				Resource string `json:"resource"`
				Label    string `json:"label"`
			} `json:"resources"`
			Permissions []struct {
				Permission string `json:"permission"`
				Rank       int    `json:"rank"`
			} `json:"permissions"`
			Matrix map[string]map[string]string `json:"matrix"`
		} `json:"data"`
		Meta struct {
			PolicyVersion string `json:"policy_version"`
			Source        string `json:"source"`
			Editable      bool   `json:"editable"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)

	if resp.Status != "ok" {
		t.Fatalf("expected ok status, got %q", resp.Status)
	}
	if len(resp.Data.Roles) != 3 || len(resp.Data.Resources) != 6 || len(resp.Data.Permissions) != 4 {
		t.Fatalf("unexpected policy dimensions: roles=%d resources=%d permissions=%d", len(resp.Data.Roles), len(resp.Data.Resources), len(resp.Data.Permissions))
	}
	if got := resp.Data.Matrix["developer"]["commands"]; got != "none" {
		t.Fatalf("expected developer commands permission none, got %q", got)
	}
	if got := resp.Data.Matrix["manager"]["risks"]; got != "write" {
		t.Fatalf("expected manager risks permission write, got %q", got)
	}
	if got := resp.Data.Matrix["system_admin"]["system_config"]; got != "admin" {
		t.Fatalf("expected system_admin system_config permission admin, got %q", got)
	}
	if resp.Meta.PolicyVersion == "" || resp.Meta.Source != "static_default_policy" || resp.Meta.Editable {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}

func TestGetRBACPolicyUsesConfiguredStore(t *testing.T) {
	policy := domain.DefaultRBACPolicy()
	policy.PolicyVersion = "stored"
	policy.Source = "memory"
	policy.Editable = true
	policy.Matrix["developer"]["commands"] = domain.RBACPermissionRead
	router := NewRouter(RouterConfig{RBACPolicyStore: &memoryRBACPolicyStore{policy: policy}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policy", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Matrix map[string]map[string]string `json:"matrix"`
		} `json:"data"`
		Meta struct {
			PolicyVersion string `json:"policy_version"`
			Source        string `json:"source"`
			Editable      bool   `json:"editable"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Meta.PolicyVersion != "stored" || resp.Meta.Source != "memory" || !resp.Meta.Editable {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
	if got := resp.Data.Matrix["developer"]["commands"]; got != "read" {
		t.Fatalf("expected stored policy matrix, got %q", got)
	}
}

func TestReplaceRBACPolicyPersistsAndAudits(t *testing.T) {
	rbacStore := &memoryRBACPolicyStore{}
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{RBACPolicyStore: rbacStore, AuditStore: audits})

	body := []byte(`{
		"policy_version": "rbac-2026-05-07-test",
		"reason": "Allow developers to create draft commands",
		"matrix": {
			"developer": {
				"repositories": "read",
				"ci_runs": "read",
				"risks": "read",
				"commands": "read",
				"organization": "none",
				"system_config": "none"
			},
			"manager": {
				"repositories": "write",
				"ci_runs": "read",
				"risks": "write",
				"commands": "write",
				"organization": "read",
				"system_config": "none"
			},
			"system_admin": {
				"repositories": "admin",
				"ci_runs": "admin",
				"risks": "admin",
				"commands": "admin",
				"organization": "admin",
				"system_config": "admin"
			}
		}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "admin")
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rbacStore.input.ActorLogin != "admin" || rbacStore.input.Reason != "Allow developers to create draft commands" {
		t.Fatalf("unexpected replace input: %+v", rbacStore.input)
	}
	if got := rbacStore.policy.Matrix["developer"]["commands"]; got != domain.RBACPermissionRead {
		t.Fatalf("expected updated developer commands permission, got %q", got)
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log, got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.Action != "rbac.policy_replaced" || log.TargetType != "rbac_policy" || log.TargetID != "rbac-2026-05-07-test" {
		t.Fatalf("unexpected audit log: %+v", log)
	}
	if rec.Header().Get("X-Devhub-Role-Deprecated") != "true" {
		t.Fatalf("expected X-Devhub-Role deprecation header")
	}
}

func TestReplaceRBACPolicyRejectsMissingMatrixResource(t *testing.T) {
	router := NewRouter(RouterConfig{RBACPolicyStore: &memoryRBACPolicyStore{}})
	body := []byte(`{
		"reason": "bad policy",
		"matrix": {
			"developer": {},
			"manager": {},
			"system_admin": {}
		}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReplaceRBACPolicyRequiresActorRole(t *testing.T) {
	router := NewRouter(RouterConfig{RBACPolicyStore: &memoryRBACPolicyStore{}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader([]byte(`{}`))))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReplaceRBACPolicyRejectsInsufficientPermission(t *testing.T) {
	router := NewRouter(RouterConfig{RBACPolicyStore: &memoryRBACPolicyStore{}})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Devhub-Role", "manager")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReplaceRBACPolicyUsesMappedUserRoleBeforeHeaderRole(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	orgs.users["u-admin"] = domain.AppUser{
		UserID: "u-admin",
		Role:   domain.AppRoleSystemAdmin,
		Status: domain.UserStatusActive,
	}
	rbacStore := &memoryRBACPolicyStore{}
	router := NewRouter(RouterConfig{
		OrganizationStore: orgs,
		RBACPolicyStore:   rbacStore,
	})

	body := []byte(`{
		"reason": "mapped role wins",
		"matrix": {
			"developer": {
				"repositories": "read",
				"ci_runs": "read",
				"risks": "read",
				"commands": "none",
				"organization": "none",
				"system_config": "none"
			},
			"manager": {
				"repositories": "write",
				"ci_runs": "read",
				"risks": "write",
				"commands": "write",
				"organization": "read",
				"system_config": "none"
			},
			"system_admin": {
				"repositories": "admin",
				"ci_runs": "admin",
				"risks": "admin",
				"commands": "admin",
				"organization": "admin",
				"system_config": "admin"
			}
		}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "u-admin")
	req.Header.Set("X-Devhub-Role", "manager")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 because users.role is system_admin, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rbacStore.input.ActorLogin != "u-admin" {
		t.Fatalf("expected actor u-admin, got %+v", rbacStore.input)
	}
}

func TestReplaceRBACPolicyRejectsUnmappedAuthenticatedActorBeforeHeaderRole(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	router := NewRouter(RouterConfig{
		OrganizationStore: orgs,
		RBACPolicyStore:   &memoryRBACPolicyStore{},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policy", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Devhub-Actor", "missing-user")
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 because authenticated actor is not mapped, got %d body=%s", rec.Code, rec.Body.String())
	}
}
