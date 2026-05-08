package store_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

func newTestRBACStore(t *testing.T) (*store.PostgresStore, context.Context) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	t.Cleanup(pgStore.Close)
	return pgStore, ctx
}

func TestRBAC_ListRoles_SeedsThreeSystemRoles(t *testing.T) {
	s, ctx := newTestRBACStore(t)

	roles, err := s.ListRBACRoles(ctx)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) < 3 {
		t.Fatalf("expected at least 3 system roles, got %d", len(roles))
	}

	wantOrder := []string{"developer", "manager", "system_admin"}
	for i, want := range wantOrder {
		if roles[i].ID != want {
			t.Errorf("roles[%d].ID = %q, want %q (system roles must come first in fixed order)", i, roles[i].ID, want)
		}
		if !roles[i].System {
			t.Errorf("roles[%d].System = false for %q, want true", i, roles[i].ID)
		}
	}

	for _, want := range domain.SystemRoles() {
		got, err := s.GetRBACRole(ctx, want.ID)
		if err != nil {
			t.Fatalf("get role %s: %v", want.ID, err)
		}
		for _, r := range domain.AllResources() {
			if got.Permissions[r] != want.Permissions[r] {
				t.Errorf("role %s resource %s permissions = %+v, want %+v", want.ID, r, got.Permissions[r], want.Permissions[r])
			}
		}
		if got.Name != want.Name {
			t.Errorf("role %s name = %q, want %q", want.ID, got.Name, want.Name)
		}
	}
}

func TestRBAC_GetRole_NotFound(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	_, err := s.GetRBACRole(ctx, "custom-does-not-exist")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRBAC_CreateUpdateDeleteCustomRole(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	id := fmt.Sprintf("custom-test-%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = s.DeleteRBACRole(context.Background(), id) })

	created, err := s.CreateRBACRole(ctx, domain.RBACRole{
		ID:          id,
		Name:        "Test Role",
		Description: "PR-G3 integration test",
		Permissions: domain.PermissionMatrix{
			domain.ResourceInfrastructure: {View: true},
			domain.ResourcePipelines:      {View: true, Create: true},
		},
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if created.System {
		t.Errorf("created.System = true, want false for custom role")
	}

	for _, r := range domain.AllResources() {
		if _, ok := created.Permissions[r]; !ok {
			t.Errorf("created role missing resource %q (EnforceAuditInvariant should fill all 5)", r)
		}
	}

	updated, err := s.UpdateRBACRolePermissions(ctx, id, domain.PermissionMatrix{
		domain.ResourceInfrastructure: {View: true, Edit: true},
		domain.ResourcePipelines:      {View: true},
		domain.ResourceOrganization:   {View: true},
		domain.ResourceSecurity:       {View: true},
		domain.ResourceAudit:          {View: true},
	})
	if err != nil {
		t.Fatalf("update permissions: %v", err)
	}
	if !updated.Permissions[domain.ResourceInfrastructure].Edit {
		t.Errorf("permissions update not applied: %+v", updated.Permissions[domain.ResourceInfrastructure])
	}
	if updated.Permissions[domain.ResourceAudit].Create || updated.Permissions[domain.ResourceAudit].Edit || updated.Permissions[domain.ResourceAudit].Delete {
		t.Errorf("audit invariant not enforced after update: %+v", updated.Permissions[domain.ResourceAudit])
	}

	updated2, err := s.UpdateRBACRoleMetadata(ctx, id, "Renamed", "Description after rename")
	if err != nil {
		t.Fatalf("update metadata: %v", err)
	}
	if updated2.Name != "Renamed" || updated2.Description != "Description after rename" {
		t.Errorf("metadata update not applied: %+v", updated2)
	}

	if err := s.DeleteRBACRole(ctx, id); err != nil {
		t.Fatalf("delete role: %v", err)
	}
	if _, err := s.GetRBACRole(ctx, id); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("post-delete get err = %v, want ErrNotFound", err)
	}
}

func TestRBAC_SystemRoleImmutable(t *testing.T) {
	s, ctx := newTestRBACStore(t)

	if err := s.DeleteRBACRole(ctx, "manager"); !errors.Is(err, store.ErrSystemRoleImmutable) {
		t.Errorf("delete system role err = %v, want ErrSystemRoleImmutable", err)
	}

	if _, err := s.UpdateRBACRoleMetadata(ctx, "developer", "Hacked", "should not change"); !errors.Is(err, store.ErrSystemRoleImmutable) {
		t.Errorf("rename system role err = %v, want ErrSystemRoleImmutable", err)
	}

	if _, err := s.CreateRBACRole(ctx, domain.RBACRole{ID: "system_admin", Name: "x"}); !errors.Is(err, store.ErrConflict) {
		t.Errorf("create with system id err = %v, want ErrConflict", err)
	}
}

func TestRBAC_AuditInvariantOnCreate(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	id := fmt.Sprintf("custom-audit-%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = s.DeleteRBACRole(context.Background(), id) })

	role := domain.RBACRole{
		ID:   id,
		Name: "Audit attempt",
		Permissions: domain.PermissionMatrix{
			domain.ResourceAudit: {View: true, Create: true, Edit: true, Delete: true},
		},
	}
	created, err := s.CreateRBACRole(ctx, role)
	if err != nil {
		t.Fatalf("create with audit write should succeed (helper strips it): %v", err)
	}
	got := created.Permissions[domain.ResourceAudit]
	if got.Create || got.Edit || got.Delete {
		t.Errorf("audit invariant not enforced: %+v", got)
	}
}

func TestRBAC_DeleteCustomRoleInUse(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	roleID := fmt.Sprintf("custom-inuse-%d", time.Now().UnixNano())
	userID := fmt.Sprintf("u-rbac-test-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		bg := context.Background()
		_ = s.SetSubjectRole(bg, userID, "developer")
		_ = s.DeleteUser(bg, userID)
		_ = s.DeleteRBACRole(bg, roleID)
	})

	if _, err := s.CreateRBACRole(ctx, domain.RBACRole{
		ID:          roleID,
		Name:        "In-use role",
		Description: "for delete-in-use test",
		Permissions: domain.PermissionMatrix{domain.ResourceInfrastructure: {View: true}},
	}); err != nil {
		t.Fatalf("create custom role: %v", err)
	}

	if _, err := s.CreateUser(ctx, domain.CreateUserInput{
		UserID:      userID,
		Email:       userID + "@example.com",
		DisplayName: "RBAC Test User",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		JoinedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := s.SetSubjectRole(ctx, userID, roleID); err != nil {
		t.Fatalf("set subject role: %v", err)
	}

	if err := s.DeleteRBACRole(ctx, roleID); !errors.Is(err, store.ErrRoleInUse) {
		t.Errorf("delete in-use role err = %v, want ErrRoleInUse", err)
	}
}

func TestRBAC_SetSubjectRole_RoleNotFound(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	userID := fmt.Sprintf("u-rbac-nf-%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = s.DeleteUser(context.Background(), userID) })

	if _, err := s.CreateUser(ctx, domain.CreateUserInput{
		UserID:      userID,
		Email:       userID + "@example.com",
		DisplayName: "RBAC NF User",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		JoinedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := s.SetSubjectRole(ctx, userID, "custom-nope-not-real"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("set non-existent role err = %v, want ErrNotFound", err)
	}
}

func TestRBAC_GetSubjectRoles_SingleRoleMode(t *testing.T) {
	s, ctx := newTestRBACStore(t)
	roles, err := s.GetSubjectRoles(ctx, "u1")
	if err != nil {
		t.Skipf("seed user u1 not present (env-dependent): %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("GetSubjectRoles len = %d, want 1 (single role mode)", len(roles))
	}
}
