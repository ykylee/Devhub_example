package view

import (
	"context"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

func TestPermissionCache_NilStore_FallbackToSystemRoles(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	// admin should have access to everything
	allowed, err := cache.Allows(ctx, "system_admin", domain.ResourceApplications, domain.ActionDelete)
	if err != nil {
		t.Fatalf("admin Allows: %v", err)
	}
	if !allowed {
		t.Error("admin should be allowed to delete applications")
	}
}

func TestPermissionCache_NilStore_DeveloperLimited(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	// developer can view infrastructure
	allowed, err := cache.Allows(ctx, "developer", domain.ResourceInfrastructure, domain.ActionView)
	if err != nil {
		t.Fatalf("developer Allows(view infra): %v", err)
	}
	if !allowed {
		t.Error("developer should be allowed to view infrastructure")
	}

	// developer cannot delete applications
	allowed, err = cache.Allows(ctx, "developer", domain.ResourceApplications, domain.ActionDelete)
	if err != nil {
		t.Fatalf("developer Allows(delete app): %v", err)
	}
	if allowed {
		t.Error("developer should NOT be allowed to delete applications")
	}

	// developer cannot view applications at all
	allowed, err = cache.Allows(ctx, "developer", domain.ResourceApplications, domain.ActionView)
	if err != nil {
		t.Fatalf("developer Allows(view app): %v", err)
	}
	if allowed {
		t.Error("developer should NOT be allowed to view applications")
	}
}

func TestPermissionCache_NilStore_ManagerCanEdit(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	// manager can view organization
	allowed, err := cache.Allows(ctx, "manager", domain.ResourceOrganization, domain.ActionView)
	if err != nil {
		t.Fatalf("manager Allows(view org): %v", err)
	}
	if !allowed {
		t.Error("manager should be allowed to view organization")
	}

	// manager cannot edit organization
	allowed, err = cache.Allows(ctx, "manager", domain.ResourceOrganization, domain.ActionEdit)
	if err != nil {
		t.Fatalf("manager Allows(edit org): %v", err)
	}
	if allowed {
		t.Error("manager should NOT be allowed to edit organization")
	}

	// manager can create security resources (risks)
	allowed, err = cache.Allows(ctx, "manager", domain.ResourceSecurity, domain.ActionCreate)
	if err != nil {
		t.Fatalf("manager Allows(create security): %v", err)
	}
	if !allowed {
		t.Error("manager should be allowed to create security resources")
	}
}

func TestPermissionCache_UnknownRole_ReturnsFalse(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	allowed, err := cache.Allows(ctx, "nonexistent_role", domain.ResourceApplications, domain.ActionView)
	if err != nil {
		t.Fatalf("unknown role Allows: %v", err)
	}
	if allowed {
		t.Error("unknown role should not be allowed anything")
	}
}

func TestPermissionCache_Invalidate_ReloadsRoles(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	// First call loads roles
	allowed, err := cache.Allows(ctx, "system_admin", domain.ResourceApplications, domain.ActionView)
	if err != nil {
		t.Fatalf("first Allows: %v", err)
	}
	if !allowed {
		t.Error("system_admin should be allowed before invalidation")
	}

	cache.Invalidate()

	allowed, err = cache.Allows(ctx, "system_admin", domain.ResourceApplications, domain.ActionView)
	if err != nil {
		t.Fatalf("reload Allows: %v", err)
	}
	if !allowed {
		t.Error("system_admin should still be allowed after reload")
	}
}

func TestPermissionCache_SystemAdminHasFullAccess(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	for _, role := range domain.SystemRoles() {
		if role.ID != "system_admin" {
			continue
		}
		for resource := range role.Permissions {
			allowed, err := cache.Allows(ctx, role.ID, resource, domain.ActionView)
			if err != nil {
				t.Fatalf("role %q resource %q Allows(view): %v", role.ID, resource, err)
			}
			if !allowed {
				t.Errorf("system_admin should have view access to %q", resource)
			}
			// system_admin can also delete everything (except audit)
			if resource != domain.ResourceAudit {
				deleteAllowed, err := cache.Allows(ctx, role.ID, resource, domain.ActionDelete)
				if err != nil {
					t.Fatalf("role %q resource %q Allows(delete): %v", role.ID, resource, err)
				}
				if !deleteAllowed {
					t.Errorf("system_admin should have delete access to %q", resource)
				}
			}
		}
	}
}

func TestPermissionCache_NilStore_ConcurrentSafe(t *testing.T) {
	cache := NewPermissionCache(nil)
	ctx := context.Background()

	// Run concurrent reads to catch data races
	const goroutines = 10
	done := make(chan struct{}, goroutines)
	for range goroutines {
		go func() {
			_, _ = cache.Allows(ctx, "system_admin", domain.ResourceApplications, domain.ActionView)
			done <- struct{}{}
		}()
	}
	for range goroutines {
		<-done
	}
}
