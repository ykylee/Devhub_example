package view

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type fakeRBACAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeRBACAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_rbac_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewRBACHandler_NonNil(t *testing.T) {
	h := NewRBACHandler(RBACConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewRBACHandler_ConfigPropagation(t *testing.T) {
	cfg := RBACConfig{AuthDevFallback: true, AuditStore: &fakeRBACAuditStore{}}
	h := NewRBACHandler(cfg)
	if !h.cfg.AuthDevFallback {
		t.Fatal("AuthDevFallback not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRBACHandler(RBACConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "action", "type", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRBACAuditStore{}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "rbac.test", "role", "role-1", nil)
	if got.AuditID != "audit_rbac_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "rbac.test" || c0.TargetType != "role" || c0.TargetID != "role-1" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRBACAuditStore{}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)

	h.recordAuditBestEffort(c, "a", "t", "id", map[string]any{"existing": "value"})
	created := store.created[0]
	if created.Payload["existing"] != "value" {
		t.Fatalf("existing payload key lost: %+v", created.Payload)
	}
	if _, ok := created.Payload["actor_source"]; !ok {
		t.Fatal("actor_source must be augmented")
	}
}

func TestRecordAuditBestEffort_PersistFailureLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logBuf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(orig) })

	store := &fakeRBACAuditStore{err: errors.New("db_down")}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set(httphelp.CtxKeyRequestID, "req_x")

	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatal("expected zero audit on err")
	}
	if !strings.Contains(logBuf.String(), "audit log persistence failed") {
		t.Fatalf("expected log msg, got %q", logBuf.String())
	}
}

func TestAddAuditMeta(t *testing.T) {
	t.Run("with audit id stamps key", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{AuditID: "audit_x"})
		if resp["audit_log_id"] != "audit_x" {
			t.Fatalf("got %v", resp["audit_log_id"])
		}
	})
	t.Run("without audit id no-op", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{})
		if _, ok := resp["audit_log_id"]; ok {
			t.Fatal("must not stamp empty id")
		}
	})
}

type fakeRBACStore struct {
	listRBACRolesFunc             func(ctx context.Context) ([]domain.RBACRole, error)
	getRBACRoleFunc               func(ctx context.Context, roleID string) (domain.RBACRole, error)
	createRBACRoleFunc             func(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error)
	updateRBACRolePermissionsFunc func(ctx context.Context, roleID string, perms domain.PermissionMatrix) (domain.RBACRole, error)
	updateRBACRoleMetadataFunc    func(ctx context.Context, roleID, name, description string) (domain.RBACRole, error)
	deleteRBACRoleFunc            func(ctx context.Context, roleID string) error
}

func (f *fakeRBACStore) ListRBACRoles(ctx context.Context) ([]domain.RBACRole, error) {
	if f.listRBACRolesFunc != nil {
		return f.listRBACRolesFunc(ctx)
	}
	return nil, nil
}
func (f *fakeRBACStore) GetRBACRole(ctx context.Context, roleID string) (domain.RBACRole, error) {
	if f.getRBACRoleFunc != nil {
		return f.getRBACRoleFunc(ctx, roleID)
	}
	return domain.RBACRole{}, nil
}
func (f *fakeRBACStore) CreateRBACRole(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
	if f.createRBACRoleFunc != nil {
		return f.createRBACRoleFunc(ctx, role)
	}
	return role, nil
}
func (f *fakeRBACStore) UpdateRBACRolePermissions(ctx context.Context, roleID string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
	if f.updateRBACRolePermissionsFunc != nil {
		return f.updateRBACRolePermissionsFunc(ctx, roleID, perms)
	}
	return domain.RBACRole{}, nil
}
func (f *fakeRBACStore) UpdateRBACRoleMetadata(ctx context.Context, roleID, name, description string) (domain.RBACRole, error) {
	if f.updateRBACRoleMetadataFunc != nil {
		return f.updateRBACRoleMetadataFunc(ctx, roleID, name, description)
	}
	return domain.RBACRole{}, nil
}
func (f *fakeRBACStore) DeleteRBACRole(ctx context.Context, roleID string) error {
	if f.deleteRBACRoleFunc != nil {
		return f.deleteRBACRoleFunc(ctx, roleID)
	}
	return nil
}

func TestRBAC_API(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetRBACPolicyLegacyGone", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		h.GetRBACPolicyLegacyGone(c)
		if rec.Code != 410 {
			t.Fatalf("expected 410, got %d", rec.Code)
		}
	})

	t.Run("ListRBACPolicies - store nil", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		h.ListRBACPolicies(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("ListRBACPolicies - DB error", func(t *testing.T) {
		storeI := &fakeRBACStore{
			listRBACRolesFunc: func(ctx context.Context) ([]domain.RBACRole, error) {
				return nil, errors.New("db down")
			},
		}
		h := NewRBACHandler(RBACConfig{RBACStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/policies", nil)
		h.ListRBACPolicies(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ListRBACPolicies - Success", func(t *testing.T) {
		storeI := &fakeRBACStore{
			listRBACRolesFunc: func(ctx context.Context) ([]domain.RBACRole, error) {
				return []domain.RBACRole{
					{ID: "role1", Name: "Role 1", Description: "Desc 1", System: false, Permissions: domain.PermissionMatrix{}},
				}, nil
			},
		}
		h := NewRBACHandler(RBACConfig{RBACStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/policies", nil)
		h.ListRBACPolicies(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"id":"role1"`) {
			t.Fatalf("expected role1 inside, got %s", rec.Body.String())
		}
	})

	t.Run("CreateRBACPolicy - negative and error paths", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		h.CreateRBACPolicy(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeRBACStore{}
		h = NewRBACHandler(RBACConfig{RBACStore: storeI})

		// 2. invalid json -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/policies", strings.NewReader("invalid-json"))
		h.CreateRBACPolicy(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. empty id/name -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "", "name": ""}`))
		h.CreateRBACPolicy(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		// 4. system role reserved -> 422
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "system_admin", "name": "Admin"}`))
		h.CreateRBACPolicy(c4)
		if rec4.Code != 422 {
			t.Fatalf("expected 422, got %d", rec4.Code)
		}

		// 5. invalid id regex -> 400
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "INVALID ROLE ID", "name": "Name"}`))
		h.CreateRBACPolicy(c5)
		if rec5.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec5.Code, rec5.Body.String())
		}

		// 6. conflict role id -> 409
		storeI6 := &fakeRBACStore{
			createRBACRoleFunc: func(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrConflict
			},
		}
		h6 := NewRBACHandler(RBACConfig{RBACStore: storeI6})
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "custom-role", "name": "Name"}`))
		h6.CreateRBACPolicy(c6)
		if rec6.Code != 409 {
			t.Fatalf("expected 409, got %d", rec6.Code)
		}

		// 7. audit invariant violation -> 422
		storeI7 := &fakeRBACStore{
			createRBACRoleFunc: func(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrAuditInvariantViolation
			},
		}
		h7 := NewRBACHandler(RBACConfig{RBACStore: storeI7})
		rec7 := httptest.NewRecorder()
		c7, _ := gin.CreateTestContext(rec7)
		c7.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "custom-role", "name": "Name"}`))
		h7.CreateRBACPolicy(c7)
		if rec7.Code != 422 {
			t.Fatalf("expected 422, got %d", rec7.Code)
		}

		// 8. normal db error -> 500
		storeI8 := &fakeRBACStore{
			createRBACRoleFunc: func(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
				return domain.RBACRole{}, errors.New("db error")
			},
		}
		h8 := NewRBACHandler(RBACConfig{RBACStore: storeI8})
		rec8 := httptest.NewRecorder()
		c8, _ := gin.CreateTestContext(rec8)
		c8.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "custom-role", "name": "Name"}`))
		h8.CreateRBACPolicy(c8)
		if rec8.Code != 500 {
			t.Fatalf("expected 500, got %d", rec8.Code)
		}
	})

	t.Run("CreateRBACPolicy - Success", func(t *testing.T) {
		storeI := &fakeRBACStore{
			createRBACRoleFunc: func(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
				role.System = false
				return role, nil
			},
		}
		cache := NewPermissionCache(storeI)
		auditStore := &fakeRBACAuditStore{}
		h := NewRBACHandler(RBACConfig{RBACStore: storeI, PermissionCache: cache, AuditStore: auditStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/policies", strings.NewReader(`{"id": "custom-role", "name": "Name", "description": "desc"}`))
		h.CreateRBACPolicy(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(auditStore.created) != 1 {
			t.Fatal("expected audit log generated")
		}
		if !strings.Contains(rec.Body.String(), `"audit_log_id":"audit_rbac_id"`) {
			t.Fatalf("expected audit meta, got %s", rec.Body.String())
		}
	})

	t.Run("UpdateRBACPolicies - negative and validation paths", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		h.UpdateRBACPolicies(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeRBACStore{}
		h = NewRBACHandler(RBACConfig{RBACStore: storeI})

		// 2. bind failure -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader("invalid-json"))
		h.UpdateRBACPolicies(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. empty roles -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": []}`))
		h.UpdateRBACPolicies(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		// 4. empty role id -> 400
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": ""}]}`))
		h.UpdateRBACPolicies(c4)
		if rec4.Code != 400 {
			t.Fatalf("expected 400, got %d", rec4.Code)
		}

		// 5. role not found -> 404
		storeI5 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrNotFound
			},
		}
		h5 := NewRBACHandler(RBACConfig{RBACStore: storeI5})
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "ghost-role"}]}`))
		h5.UpdateRBACPolicies(c5)
		if rec5.Code != 404 {
			t.Fatalf("expected 404, got %d", rec5.Code)
		}

		// 6. DB error on lookup -> 500
		storeI6 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{}, errors.New("db error")
			},
		}
		h6 := NewRBACHandler(RBACConfig{RBACStore: storeI6})
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "role"}]}`))
		h6.UpdateRBACPolicies(c6)
		if rec6.Code != 500 {
			t.Fatalf("expected 500, got %d", rec6.Code)
		}

		// 7. audit invariant violation checked in loop -> 422
		storeI7 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
		}
		h7 := NewRBACHandler(RBACConfig{RBACStore: storeI7})
		rec7 := httptest.NewRecorder()
		c7, _ := gin.CreateTestContext(rec7)
		c7.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "role", "permissions": {"audit": {"create": true}}}]}`))
		h7.UpdateRBACPolicies(c7)
		if rec7.Code != 422 {
			t.Fatalf("expected 422, got %d. Body: %s", rec7.Code, rec7.Body.String())
		}

		// 8. system role metadata update -> 422
		storeI8 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, System: true}, nil
			},
		}
		h8 := NewRBACHandler(RBACConfig{RBACStore: storeI8})
		rec8 := httptest.NewRecorder()
		c8, _ := gin.CreateTestContext(rec8)
		newName := "changed"
		req8Body, _ := json.Marshal(rbacUpdatePoliciesRequest{
			Roles: []rbacUpdateRoleWire{{ID: "system_admin", Name: &newName}},
		})
		c8.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(string(req8Body)))
		h8.UpdateRBACPolicies(c8)
		if rec8.Code != 422 {
			t.Fatalf("expected 422, got %d. Body: %s", rec8.Code, rec8.Body.String())
		}
	})

	t.Run("UpdateRBACPolicies - DB update exception paths", func(t *testing.T) {
		// 1. permissions update ErrNotFound -> 404
		storeI1 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			updateRBACRolePermissionsFunc: func(ctx context.Context, id string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrNotFound
			},
		}
		h1 := NewRBACHandler(RBACConfig{RBACStore: storeI1})
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "role", "permissions": {}}]}`))
		h1.UpdateRBACPolicies(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// 2. permissions update ErrAuditInvariantViolation -> 422
		storeI2 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			updateRBACRolePermissionsFunc: func(ctx context.Context, id string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrAuditInvariantViolation
			},
		}
		h2 := NewRBACHandler(RBACConfig{RBACStore: storeI2})
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "role", "permissions": {}}]}`))
		h2.UpdateRBACPolicies(c2)
		if rec2.Code != 422 {
			t.Fatalf("expected 422, got %d", rec2.Code)
		}

		// 3. permissions update general db err -> 500
		storeI3 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			updateRBACRolePermissionsFunc: func(ctx context.Context, id string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
				return domain.RBACRole{}, errors.New("db error")
			},
		}
		h3 := NewRBACHandler(RBACConfig{RBACStore: storeI3})
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(`{"roles": [{"id": "role", "permissions": {}}]}`))
		h3.UpdateRBACPolicies(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}

		// 4. metadata update ErrSystemRoleImmutable -> 422
		newName := "Name"
		storeI4 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, System: false}, nil
			},
			updateRBACRoleMetadataFunc: func(ctx context.Context, id, name, desc string) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrSystemRoleImmutable
			},
		}
		h4 := NewRBACHandler(RBACConfig{RBACStore: storeI4})
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		req4Body, _ := json.Marshal(rbacUpdatePoliciesRequest{
			Roles: []rbacUpdateRoleWire{{ID: "role", Name: &newName}},
		})
		c4.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(string(req4Body)))
		h4.UpdateRBACPolicies(c4)
		if rec4.Code != 422 {
			t.Fatalf("expected 422, got %d", rec4.Code)
		}

		// 5. metadata update ErrNotFound -> 404
		storeI5 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, System: false}, nil
			},
			updateRBACRoleMetadataFunc: func(ctx context.Context, id, name, desc string) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrNotFound
			},
		}
		h5 := NewRBACHandler(RBACConfig{RBACStore: storeI5})
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(string(req4Body)))
		h5.UpdateRBACPolicies(c5)
		if rec5.Code != 404 {
			t.Fatalf("expected 404, got %d", rec5.Code)
		}

		// 6. metadata update general db err -> 500
		storeI6 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, System: false}, nil
			},
			updateRBACRoleMetadataFunc: func(ctx context.Context, id, name, desc string) (domain.RBACRole, error) {
				return domain.RBACRole{}, errors.New("db error")
			},
		}
		h6 := NewRBACHandler(RBACConfig{RBACStore: storeI6})
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(string(req4Body)))
		h6.UpdateRBACPolicies(c6)
		if rec6.Code != 500 {
			t.Fatalf("expected 500, got %d", rec6.Code)
		}
	})

	t.Run("UpdateRBACPolicies - Success", func(t *testing.T) {
		newName := "CustomRoleName"
		storeI := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, Name: "OldName", System: false}, nil
			},
			updateRBACRolePermissionsFunc: func(ctx context.Context, id string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
				return domain.RBACRole{ID: id, System: false, Permissions: perms}, nil
			},
			updateRBACRoleMetadataFunc: func(ctx context.Context, id, name, desc string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: id, Name: name, Description: desc, System: false}, nil
			},
			listRBACRolesFunc: func(ctx context.Context) ([]domain.RBACRole, error) {
				return []domain.RBACRole{
					{ID: "custom", Name: "CustomRoleName", System: false},
				}, nil
			},
		}
		cache := NewPermissionCache(storeI)
		h := NewRBACHandler(RBACConfig{RBACStore: storeI, PermissionCache: cache})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		reqBody, _ := json.Marshal(rbacUpdatePoliciesRequest{
			Roles: []rbacUpdateRoleWire{{ID: "custom", Name: &newName, Permissions: domain.PermissionMatrix{}}},
		})
		c.Request = httptest.NewRequest("PUT", "/policies", strings.NewReader(string(reqBody)))
		h.UpdateRBACPolicies(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DeleteRBACPolicy - error paths", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		h.DeleteRBACPolicy(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeRBACStore{}
		h = NewRBACHandler(RBACConfig{RBACStore: storeI})

		// 2. empty role_id param -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		h.DeleteRBACPolicy(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. lookup role not found -> 404
		storeI3 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{}, store.ErrNotFound
			},
		}
		h3 := NewRBACHandler(RBACConfig{RBACStore: storeI3})
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c3.Params = gin.Params{{Key: "role_id", Value: "ghost"}}
		h3.DeleteRBACPolicy(c3)
		if c3.Writer.Status() != 404 {
			t.Fatalf("expected 404, got %d", c3.Writer.Status())
		}

		// 4. lookup role DB err -> 500
		storeI4 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{}, errors.New("db down")
			},
		}
		h4 := NewRBACHandler(RBACConfig{RBACStore: storeI4})
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c4.Params = gin.Params{{Key: "role_id", Value: "some-role"}}
		h4.DeleteRBACPolicy(c4)
		if c4.Writer.Status() != 500 {
			t.Fatalf("expected 500, got %d", c4.Writer.Status())
		}

		// 5. delete ErrSystemRoleImmutable -> 422
		storeI5 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, System: true}, nil
			},
			deleteRBACRoleFunc: func(ctx context.Context, id string) error {
				return store.ErrSystemRoleImmutable
			},
		}
		h5 := NewRBACHandler(RBACConfig{RBACStore: storeI5})
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c5.Params = gin.Params{{Key: "role_id", Value: "system_admin"}}
		h5.DeleteRBACPolicy(c5)
		if c5.Writer.Status() != 422 {
			t.Fatalf("expected 422, got %d", c5.Writer.Status())
		}

		// 6. delete ErrRoleInUse -> 422
		storeI6 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			deleteRBACRoleFunc: func(ctx context.Context, id string) error {
				return store.ErrRoleInUse
			},
		}
		h6 := NewRBACHandler(RBACConfig{RBACStore: storeI6})
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c6.Params = gin.Params{{Key: "role_id", Value: "in-use"}}
		h6.DeleteRBACPolicy(c6)
		if c6.Writer.Status() != 422 {
			t.Fatalf("expected 422, got %d", c6.Writer.Status())
		}

		// 7. delete ErrNotFound -> 404
		storeI7 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			deleteRBACRoleFunc: func(ctx context.Context, id string) error {
				return store.ErrNotFound
			},
		}
		h7 := NewRBACHandler(RBACConfig{RBACStore: storeI7})
		rec7 := httptest.NewRecorder()
		c7, _ := gin.CreateTestContext(rec7)
		c7.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c7.Params = gin.Params{{Key: "role_id", Value: "vanished"}}
		h7.DeleteRBACPolicy(c7)
		if c7.Writer.Status() != 404 {
			t.Fatalf("expected 404, got %d", c7.Writer.Status())
		}

		// 8. delete general DB err -> 500
		storeI8 := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID}, nil
			},
			deleteRBACRoleFunc: func(ctx context.Context, id string) error {
				return errors.New("db error")
			},
		}
		h8 := NewRBACHandler(RBACConfig{RBACStore: storeI8})
		rec8 := httptest.NewRecorder()
		c8, _ := gin.CreateTestContext(rec8)
		c8.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c8.Params = gin.Params{{Key: "role_id", Value: "role"}}
		h8.DeleteRBACPolicy(c8)
		if c8.Writer.Status() != 500 {
			t.Fatalf("expected 500, got %d", c8.Writer.Status())
		}
	})

	t.Run("DeleteRBACPolicy - Success", func(t *testing.T) {
		storeI := &fakeRBACStore{
			getRBACRoleFunc: func(ctx context.Context, roleID string) (domain.RBACRole, error) {
				return domain.RBACRole{ID: roleID, Name: "Custom"}, nil
			},
		}
		cache := NewPermissionCache(storeI)
		h := NewRBACHandler(RBACConfig{RBACStore: storeI, PermissionCache: cache})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("DELETE", "/policies", nil)
		c.Params = gin.Params{{Key: "role_id", Value: "custom"}}
		h.DeleteRBACPolicy(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestRBAC_Middlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("requireMinRole - DevFallback bypass", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/x", nil)
		c.Set("devhub_auth_dev_fallback", true)
		
		middleware := h.requireMinRole(domain.AppRoleSystemAdmin)
		middleware(c)
		
		if c.Writer.Status() == 403 {
			t.Fatal("should bypass 403 under dev fallback")
		}
	})

	t.Run("requireMinRole - Forbidden paths", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. no actor role -> 403
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("GET", "/x", nil)
		middleware := h.requireMinRole(domain.AppRoleDeveloper)
		middleware(c1)
		if rec1.Code != 403 {
			t.Fatalf("expected 403, got %d", rec1.Code)
		}

		// 2. role developer meets min developer -> pass (status not 403)
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("GET", "/x", nil)
		c2.Set("devhub_actor_role", string(domain.AppRoleDeveloper))
		middleware(c2)
		if c2.Writer.Status() == 403 {
			t.Fatal("developer should pass min developer check")
		}

		// 3. role developer fails min manager -> 403
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("GET", "/x", nil)
		c3.Set("devhub_actor_role", string(domain.AppRoleDeveloper))
		middleware3 := h.requireMinRole(domain.AppRoleManager)
		middleware3(c3)
		if rec3.Code != 403 {
			t.Fatalf("expected 403, got %d", rec3.Code)
		}
	})

	t.Run("EnforceRoutePermission - DevFallback and Unmapped", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. Dev fallback bypass
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("devhub_auth_dev_fallback", true)
			c.Next()
		})
		r.GET("/api/v1/unknown-path", h.EnforceRoutePermission, func(c *gin.Context) {
			c.Status(200)
		})
		rec1 := httptest.NewRecorder()
		req1 := httptest.NewRequest("GET", "/api/v1/unknown-path", nil)
		r.ServeHTTP(rec1, req1)
		if rec1.Code == 403 {
			t.Fatal("should bypass route check on dev fallback")
		}

		// 2. Unmapped route -> 403
		r2 := gin.New()
		r2.GET("/api/v1/unknown-path", h.EnforceRoutePermission, func(c *gin.Context) {
			c.Status(200)
		})
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/v1/unknown-path", nil)
		r2.ServeHTTP(rec2, req2)
		if rec2.Code != 403 {
			t.Fatalf("expected 403, got %d", rec2.Code)
		}
	})

	t.Run("EnforceRoutePermission - Bypass, Deny and Allow", func(t *testing.T) {
		storeI := &fakeRBACStore{
			listRBACRolesFunc: func(ctx context.Context) ([]domain.RBACRole, error) {
				return domain.SystemRoles(), nil
			},
		}
		cache := NewPermissionCache(storeI)
		h := NewRBACHandler(RBACConfig{RBACStore: storeI, PermissionCache: cache})
		
		// 1. Bypass route
		r1 := gin.New()
		r1.GET("/api/v1/me", h.EnforceRoutePermission, func(c *gin.Context) {
			c.Status(200)
		})
		rec1 := httptest.NewRecorder()
		req1 := httptest.NewRequest("GET", "/api/v1/me", nil)
		r1.ServeHTTP(rec1, req1)
		if rec1.Code == 403 {
			t.Fatal("bypass route must pass")
		}

		// 2. Lacks permission -> 403
		r2 := gin.New()
		r2.Use(func(c *gin.Context) {
			c.Set("devhub_actor_role", "developer")
			c.Next()
		})
		r2.GET("/api/v1/audit-logs", h.EnforceRoutePermission, func(c *gin.Context) {
			c.Status(200)
		})
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/v1/audit-logs", nil)
		r2.ServeHTTP(rec2, req2)
		if rec2.Code != 403 {
			t.Fatalf("expected 403, got %d", rec2.Code)
		}

		// 3. Allows permission -> pass
		r3 := gin.New()
		r3.Use(func(c *gin.Context) {
			c.Set("devhub_actor_role", "system_admin")
			c.Next()
		})
		r3.GET("/api/v1/audit-logs", h.EnforceRoutePermission, func(c *gin.Context) {
			c.Status(200)
		})
		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest("GET", "/api/v1/audit-logs", nil)
		r3.ServeHTTP(rec3, req3)
		if rec3.Code == 403 {
			t.Fatal("admin should be allowed view audit logs")
		}
	})

	t.Run("EnforceRowOwnership - various paths", func(t *testing.T) {
		h := NewRBACHandler(RBACConfig{})
		
		// 1. Dev fallback bypass
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("POST", "/x", nil)
		c1.Set("devhub_auth_dev_fallback", true)
		if !h.EnforceRowOwnership(c1, "owner1") {
			t.Fatal("dev fallback must bypass ownership")
		}

		// 2. System admin pass
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/x", nil)
		c2.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		if !h.EnforceRowOwnership(c2, "owner1") {
			t.Fatal("system admin must pass ownership")
		}

		// 3. Allowed roles pass
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/x", nil)
		c3.Set("devhub_actor_role", "pmo_manager")
		if !h.EnforceRowOwnership(c3, "owner1", "pmo_manager") {
			t.Fatal("allowed role must pass ownership")
		}

		// 4. Owner self pass
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("POST", "/x", nil)
		c4.Set("devhub_actor_login", "owner1")
		if !h.EnforceRowOwnership(c4, "owner1") {
			t.Fatal("owner self must pass ownership")
		}

		// 5. ownerUserID is empty -> fail
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("POST", "/x", nil)
		c5.Set("devhub_actor_login", "")
		if h.EnforceRowOwnership(c5, "") {
			t.Fatal("empty ownerID should fail")
		}

		// 6. Denied path -> 403
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("POST", "/x", nil)
		c6.Set("devhub_actor_login", "attacker")
		c6.Set("devhub_actor_role", "developer")
		if h.EnforceRowOwnership(c6, "owner1") {
			t.Fatal("mismatch owner must fail ownership")
		}
		if rec6.Code != 403 {
			t.Fatalf("expected 403, got %d", rec6.Code)
		}
	})
}


func TestRoleRank_UnknownRoleReturnsZero(t *testing.T) {
	if got := roleRank("unknown_role_xyz"); got != 0 {
		t.Fatalf("expected 0 for unknown role, got %d", got)
	}
}

func TestRoleRank_KnownRoles(t *testing.T) {
	if got := roleRank("system_admin"); got == 0 {
		t.Fatal("expected non-zero rank for system_admin")
	}
}
