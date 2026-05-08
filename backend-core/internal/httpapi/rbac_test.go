package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// fakeRBACStore is an in-memory RBACStore for handler tests. It mimics the
// invariants of the real postgres store closely enough for endpoint behavior
// verification (audit invariant, system role immutability, role-in-use, FK).
type fakeRBACStore struct {
	roles    map[string]domain.RBACRole
	subjects map[string]string
	listErr  error
}

func newFakeRBACStore() *fakeRBACStore {
	roles := make(map[string]domain.RBACRole, 3)
	for _, role := range domain.SystemRoles() {
		role.CreatedAt = time.Now().UTC()
		role.UpdatedAt = role.CreatedAt
		roles[role.ID] = role
	}
	return &fakeRBACStore{roles: roles, subjects: map[string]string{}}
}

func (f *fakeRBACStore) ListRBACRoles(ctx context.Context) ([]domain.RBACRole, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]domain.RBACRole, 0, len(f.roles))
	for _, role := range f.roles {
		out = append(out, role)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].System != out[j].System {
			return out[i].System
		}
		rank := func(id string) int {
			switch id {
			case "developer":
				return 0
			case "manager":
				return 1
			case "system_admin":
				return 2
			default:
				return 3
			}
		}
		if out[i].System {
			return rank(out[i].ID) < rank(out[j].ID)
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (f *fakeRBACStore) GetRBACRole(ctx context.Context, roleID string) (domain.RBACRole, error) {
	role, ok := f.roles[roleID]
	if !ok {
		return domain.RBACRole{}, store.ErrNotFound
	}
	return role, nil
}

func (f *fakeRBACStore) CreateRBACRole(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
	if domain.IsSystemRole(role.ID) {
		return domain.RBACRole{}, store.ErrConflict
	}
	if _, exists := f.roles[role.ID]; exists {
		return domain.RBACRole{}, store.ErrConflict
	}
	role.System = false
	role.Permissions = domain.EnforceAuditInvariant(role.Permissions)
	role.CreatedAt = time.Now().UTC()
	role.UpdatedAt = role.CreatedAt
	f.roles[role.ID] = role
	return role, nil
}

func (f *fakeRBACStore) UpdateRBACRolePermissions(ctx context.Context, roleID string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
	role, ok := f.roles[roleID]
	if !ok {
		return domain.RBACRole{}, store.ErrNotFound
	}
	role.Permissions = domain.EnforceAuditInvariant(perms)
	role.UpdatedAt = time.Now().UTC()
	f.roles[roleID] = role
	return role, nil
}

func (f *fakeRBACStore) UpdateRBACRoleMetadata(ctx context.Context, roleID, name, description string) (domain.RBACRole, error) {
	role, ok := f.roles[roleID]
	if !ok {
		return domain.RBACRole{}, store.ErrNotFound
	}
	if role.System {
		return domain.RBACRole{}, store.ErrSystemRoleImmutable
	}
	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now().UTC()
	f.roles[roleID] = role
	return role, nil
}

func (f *fakeRBACStore) DeleteRBACRole(ctx context.Context, roleID string) error {
	role, ok := f.roles[roleID]
	if !ok {
		return store.ErrNotFound
	}
	if role.System {
		return store.ErrSystemRoleImmutable
	}
	for _, assigned := range f.subjects {
		if assigned == roleID {
			return store.ErrRoleInUse
		}
	}
	delete(f.roles, roleID)
	return nil
}

func (f *fakeRBACStore) GetSubjectRoles(ctx context.Context, userID string) ([]string, error) {
	role, ok := f.subjects[userID]
	if !ok {
		return nil, store.ErrNotFound
	}
	return []string{role}, nil
}

func (f *fakeRBACStore) SetSubjectRole(ctx context.Context, userID, roleID string) error {
	if _, ok := f.roles[roleID]; !ok {
		return store.ErrNotFound
	}
	if _, ok := f.subjects[userID]; !ok {
		return store.ErrNotFound
	}
	f.subjects[userID] = roleID
	return nil
}

func TestGetRBACPolicyLegacy_Returns410Gone(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policy", nil))
	if rec.Code != http.StatusGone {
		t.Fatalf("legacy endpoint code = %d, want 410 Gone", rec.Code)
	}
	var resp struct {
		Status string `json:"status"`
		Meta   struct {
			Replacement string `json:"replacement"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Status != "gone" {
		t.Errorf("status = %q, want \"gone\"", resp.Status)
	}
	if resp.Meta.Replacement != "/api/v1/rbac/policies" {
		t.Errorf("replacement pointer missing: meta=%+v", resp.Meta)
	}
}

func TestListRBACPolicies_SystemRolesFirstWithDefaults(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policies", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Status string         `json:"status"`
		Data   []rbacRoleWire `json:"data"`
		Meta   struct {
			PolicyVersion string   `json:"policy_version"`
			Editable      bool     `json:"editable"`
			SystemRoles   []string `json:"system_roles"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if len(resp.Data) != 3 {
		t.Fatalf("got %d roles, want 3", len(resp.Data))
	}
	wantOrder := []string{"developer", "manager", "system_admin"}
	for i, want := range wantOrder {
		if resp.Data[i].ID != want {
			t.Errorf("data[%d].id = %q, want %q", i, resp.Data[i].ID, want)
		}
		if !resp.Data[i].System {
			t.Errorf("data[%d].system = false, want true", i)
		}
	}
	dev := resp.Data[0]
	if !dev.Permissions[domain.ResourceSecurity].View {
		t.Error("developer security view should be true")
	}
	if dev.Permissions[domain.ResourceAudit].View {
		t.Error("developer audit view should be false")
	}
	mgr := resp.Data[1]
	if !mgr.Permissions[domain.ResourceSecurity].Create {
		t.Error("manager security create should be true (matches POST /risks/:id/mitigations)")
	}
	sysadmin := resp.Data[2]
	if sysadmin.Permissions[domain.ResourceAudit].Create || sysadmin.Permissions[domain.ResourceAudit].Edit || sysadmin.Permissions[domain.ResourceAudit].Delete {
		t.Errorf("system_admin audit invariant violated: %+v", sysadmin.Permissions[domain.ResourceAudit])
	}
	if !resp.Meta.Editable || resp.Meta.PolicyVersion == "" || len(resp.Meta.SystemRoles) != 3 {
		t.Errorf("meta = %+v", resp.Meta)
	}
}

func TestListRBACPolicies_StoreUnavailable(t *testing.T) {
	router := testRouter(RouterConfig{}) // no RBACStore
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policies", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", rec.Code)
	}
}

func TestCreateRBACPolicy_Custom(t *testing.T) {
	store := newFakeRBACStore()
	router := testRouter(RouterConfig{RBACStore: store})
	body := bytes.NewBufferString(`{
        "id": "custom-test",
        "name": "Test Role",
        "description": "PR-G4 test",
        "permissions": {"infrastructure": {"view": true}}
    }`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/rbac/policies", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := store.roles["custom-test"]; !ok {
		t.Errorf("role not stored")
	}
}

func TestCreateRBACPolicy_RejectsSystemID(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	body := bytes.NewBufferString(`{"id": "developer", "name": "Hijack"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/rbac/policies", body))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d, want 422", rec.Code)
	}
}

func TestCreateRBACPolicy_InvalidIDFormat(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	body := bytes.NewBufferString(`{"id": "random-no-prefix", "name": "x"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/rbac/policies", body))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("code = %d, want 400", rec.Code)
	}
}

func TestUpdateRBACPolicies_PermissionUpdate(t *testing.T) {
	store := newFakeRBACStore()
	router := testRouter(RouterConfig{RBACStore: store})
	body := bytes.NewBufferString(`{
        "roles": [{
            "id": "manager",
            "permissions": {
                "infrastructure": {"view": true, "create": true},
                "pipelines":      {"view": true},
                "organization":   {"view": true},
                "security":       {"view": true, "create": true},
                "audit":          {"view": true}
            }
        }]
    }`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policies", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d body=%s", rec.Code, rec.Body.String())
	}
	mgr := store.roles["manager"]
	if !mgr.Permissions[domain.ResourceInfrastructure].Create {
		t.Errorf("manager infra create not applied: %+v", mgr.Permissions[domain.ResourceInfrastructure])
	}
}

func TestUpdateRBACPolicies_RejectsSystemRoleMetadataChange(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	body := bytes.NewBufferString(`{
        "roles": [{ "id": "developer", "name": "Hacked" }]
    }`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/policies", body))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d, want 422", rec.Code)
	}
}

func TestDeleteRBACPolicy_SystemRoleNotDeletable(t *testing.T) {
	router := testRouter(RouterConfig{RBACStore: newFakeRBACStore()})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/policies/manager", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d, want 422", rec.Code)
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "system_role_not_deletable" {
		t.Errorf("code = %q, want system_role_not_deletable", resp.Code)
	}
}

func TestDeleteRBACPolicy_RoleInUse(t *testing.T) {
	store := newFakeRBACStore()
	store.roles["custom-x"] = domain.RBACRole{ID: "custom-x", Name: "X"}
	store.subjects["u1"] = "custom-x"
	router := testRouter(RouterConfig{RBACStore: store})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/policies/custom-x", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d, want 422", rec.Code)
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "role_in_use" {
		t.Errorf("code = %q, want role_in_use", resp.Code)
	}
}

func TestDeleteRBACPolicy_CustomRole(t *testing.T) {
	store := newFakeRBACStore()
	store.roles["custom-x"] = domain.RBACRole{ID: "custom-x", Name: "X"}
	router := testRouter(RouterConfig{RBACStore: store})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/policies/custom-x", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("code = %d body=%s", rec.Code, rec.Body.String())
	}
	if _, exists := store.roles["custom-x"]; exists {
		t.Errorf("role still present after delete")
	}
}

func TestSubjectRoles_GetAndPut(t *testing.T) {
	rbacStore := newFakeRBACStore()
	rbacStore.subjects["u1"] = "developer"
	router := testRouter(RouterConfig{RBACStore: rbacStore})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/subjects/u1/roles", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get code = %d", rec.Code)
	}
	var getResp struct {
		Data []string `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &getResp)
	if len(getResp.Data) != 1 || getResp.Data[0] != "developer" {
		t.Errorf("get data = %v, want [developer]", getResp.Data)
	}

	body := bytes.NewBufferString(`{"roles": ["manager"]}`)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/subjects/u1/roles", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("put code = %d body=%s", rec.Code, rec.Body.String())
	}
	if rbacStore.subjects["u1"] != "manager" {
		t.Errorf("subject role = %q, want manager", rbacStore.subjects["u1"])
	}
}

func TestSubjectRoles_PutSingleRoleRequired(t *testing.T) {
	rbacStore := newFakeRBACStore()
	rbacStore.subjects["u1"] = "developer"
	router := testRouter(RouterConfig{RBACStore: rbacStore})

	body := bytes.NewBufferString(`{"roles": ["developer", "manager"]}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/subjects/u1/roles", body))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("code = %d, want 422", rec.Code)
	}
}

func TestSubjectRoles_PutRoleNotFound(t *testing.T) {
	rbacStore := newFakeRBACStore()
	rbacStore.subjects["u1"] = "developer"
	router := testRouter(RouterConfig{RBACStore: rbacStore})

	body := bytes.NewBufferString(`{"roles": ["custom-missing"]}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/rbac/subjects/u1/roles", body))
	if rec.Code != http.StatusNotFound {
		t.Errorf("code = %d, want 404", rec.Code)
	}
}

func TestListRBACPolicies_StoreError(t *testing.T) {
	rbacStore := newFakeRBACStore()
	rbacStore.listErr = errors.New("db down")
	router := testRouter(RouterConfig{RBACStore: rbacStore})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policies", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("code = %d, want 500", rec.Code)
	}
	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "internal error" {
		t.Errorf("error body leaks underlying message: %q", resp.Error)
	}
}
