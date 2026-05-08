package domain

import "testing"

func TestDefaultPermissionMatrix_PreservesM0Enforcement(t *testing.T) {
	cases := []struct {
		name        string
		role        AppRole
		resource    Resource
		action      Action
		wantAllowed bool
		comment     string
	}{
		{"developer view risks", AppRoleDeveloper, ResourceSecurity, ActionView, true, "GET /risks accessible to all"},
		{"developer cannot create mitigation", AppRoleDeveloper, ResourceSecurity, ActionCreate, false, "POST /risks/:id/mitigations gated to manager+"},
		{"developer cannot view audit", AppRoleDeveloper, ResourceAudit, ActionView, false, "GET /audit-logs gated to manager+"},
		{"manager creates mitigation", AppRoleManager, ResourceSecurity, ActionCreate, true, "POST /risks/:id/mitigations"},
		{"manager views audit", AppRoleManager, ResourceAudit, ActionView, true, "GET /audit-logs"},
		{"manager cannot edit users", AppRoleManager, ResourceOrganization, ActionEdit, false, "PATCH /users/:id is system_admin only"},
		{"manager cannot create service-action", AppRoleManager, ResourceInfrastructure, ActionCreate, false, "POST /admin/service-actions is system_admin only"},
		{"system_admin all org mutations", AppRoleSystemAdmin, ResourceOrganization, ActionDelete, true, "DELETE /users/:id"},
		{"system_admin service action", AppRoleSystemAdmin, ResourceInfrastructure, ActionCreate, true, "POST /admin/service-actions"},
		{"system_admin cannot create audit", AppRoleSystemAdmin, ResourceAudit, ActionCreate, false, "audit is append-only by system code"},
		{"system_admin cannot edit audit", AppRoleSystemAdmin, ResourceAudit, ActionEdit, false, "audit invariant"},
		{"system_admin cannot delete audit", AppRoleSystemAdmin, ResourceAudit, ActionDelete, false, "audit invariant"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matrix, ok := DefaultPermissionMatrix(string(tc.role))
			if !ok {
				t.Fatalf("DefaultPermissionMatrix(%q) = ok=false, want true", tc.role)
			}
			if got := Allows(matrix, tc.resource, tc.action); got != tc.wantAllowed {
				t.Errorf("Allows(%s, %s, %s) = %v, want %v (%s)", tc.role, tc.resource, tc.action, got, tc.wantAllowed, tc.comment)
			}
		})
	}
}

func TestDefaultPermissionMatrix_UnknownRole(t *testing.T) {
	if _, ok := DefaultPermissionMatrix("custom-foo"); ok {
		t.Errorf("DefaultPermissionMatrix(custom-foo) = ok=true, want false (custom roles have no default)")
	}
}

func TestDefaultPermissionMatrix_AllRolesCoverAllResources(t *testing.T) {
	for _, id := range SystemRoleIDs() {
		matrix, _ := DefaultPermissionMatrix(id)
		for _, r := range AllResources() {
			if _, ok := matrix[r]; !ok {
				t.Errorf("role %q is missing resource %q in default matrix", id, r)
			}
		}
	}
}

func TestEnforceAuditInvariant_StripsAuditWriteFlags(t *testing.T) {
	dirty := PermissionMatrix{
		ResourceAudit: {View: true, Create: true, Edit: true, Delete: true},
	}
	clean := EnforceAuditInvariant(dirty)
	got := clean[ResourceAudit]
	if !got.View {
		t.Errorf("audit view should be preserved, got %+v", got)
	}
	if got.Create || got.Edit || got.Delete {
		t.Errorf("audit invariant violated: %+v", got)
	}
}

func TestEnforceAuditInvariant_FillsMissingResources(t *testing.T) {
	clean := EnforceAuditInvariant(PermissionMatrix{})
	for _, r := range AllResources() {
		if _, ok := clean[r]; !ok {
			t.Errorf("EnforceAuditInvariant did not fill missing resource %q", r)
		}
	}
}

func TestSystemRoles_ReturnsThreeWithMatchingDefaults(t *testing.T) {
	roles := SystemRoles()
	if len(roles) != 3 {
		t.Fatalf("SystemRoles() len = %d, want 3", len(roles))
	}
	for _, role := range roles {
		if !role.System {
			t.Errorf("role %q System flag = false, want true", role.ID)
		}
		want, _ := DefaultPermissionMatrix(role.ID)
		for _, r := range AllResources() {
			if role.Permissions[r] != want[r] {
				t.Errorf("role %q resource %q permissions = %+v, want %+v", role.ID, r, role.Permissions[r], want[r])
			}
		}
		if role.Name == "" || role.Description == "" {
			t.Errorf("role %q has empty Name or Description", role.ID)
		}
	}
}

func TestValidateRoleID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"developer", true},
		{"manager", true},
		{"system_admin", true},
		{"custom-foo", true},
		{"custom-foo-bar_2", true},
		{"custom-", false},
		{"custom-Foo", false},
		{"random", false},
		{"", false},
		{"-developer", false},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			err := ValidateRoleID(tc.id)
			got := err == nil
			if got != tc.want {
				t.Errorf("ValidateRoleID(%q) ok=%v, want %v (err=%v)", tc.id, got, tc.want, err)
			}
		})
	}
}

func TestIsSystemRole(t *testing.T) {
	for _, id := range SystemRoleIDs() {
		if !IsSystemRole(id) {
			t.Errorf("IsSystemRole(%q) = false, want true", id)
		}
	}
	for _, id := range []string{"custom-foo", "", "Manager"} {
		if IsSystemRole(id) {
			t.Errorf("IsSystemRole(%q) = true, want false", id)
		}
	}
}
