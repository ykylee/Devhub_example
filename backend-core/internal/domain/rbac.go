package domain

import (
	"fmt"
	"regexp"
	"time"
)

// Resource is the RBAC resource axis defined by docs/backend_api_contract.md section 12.0.2.
type Resource string

const (
	ResourceInfrastructure          Resource = "infrastructure"
	ResourcePipelines               Resource = "pipelines"
	ResourceOrganization            Resource = "organization"
	ResourceSecurity                Resource = "security"
	ResourceAudit                   Resource = "audit"
	ResourceApplications            Resource = "applications"
	ResourceApplicationRepositories Resource = "application_repositories"
	ResourceProjects                Resource = "projects"
	ResourceSCMProviders            Resource = "scm_providers"
	// ResourceDevRequests — sprint claude/work_260515-i (ADR-0012 / ARCH-DREQ-04).
	ResourceDevRequests Resource = "dev_requests"
)

// Action is the RBAC action axis defined by docs/backend_api_contract.md section 12.0.3.
type Action string

const (
	ActionView   Action = "view"
	ActionCreate Action = "create"
	ActionEdit   Action = "edit"
	ActionDelete Action = "delete"
)

// ResourcePermissions is the per-resource 4-boolean flag set carried by a role's permission matrix.
type ResourcePermissions struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

// PermissionMatrix maps each Resource to its ResourcePermissions for a single role.
type PermissionMatrix map[Resource]ResourcePermissions

// RBACRole is the persisted view of an entry in rbac_policies.
type RBACRole struct {
	ID          string
	Name        string
	Description string
	System      bool
	Permissions PermissionMatrix
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AllResources lists the 10 canonical resources in display order. Resources 6~9
// (applications / application_repositories / projects / scm_providers) were added
// in sprint claude/work_260514-a (migration 000018, ADR-0011 §4.1). Resource 10
// (dev_requests) was added in sprint claude/work_260515-i (migration 000024,
// ARCH-DREQ-04 / ADR-0012).
func AllResources() []Resource {
	return []Resource{
		ResourceInfrastructure,
		ResourcePipelines,
		ResourceOrganization,
		ResourceSecurity,
		ResourceAudit,
		ResourceApplications,
		ResourceApplicationRepositories,
		ResourceProjects,
		ResourceSCMProviders,
		ResourceDevRequests,
	}
}

// AllActions lists the 4 canonical actions in CRUD order.
func AllActions() []Action {
	return []Action{ActionView, ActionCreate, ActionEdit, ActionDelete}
}

// SystemRoleIDs returns the immutable set of role ids that the seed migrations install.
// pmo_manager 는 migration 000021 (sprint claude/work_260515-d) 에서 도입.
func SystemRoleIDs() []string {
	return []string{
		string(AppRoleDeveloper),
		string(AppRoleManager),
		string(AppRoleSystemAdmin),
		string(AppRolePMOManager),
	}
}

// IsSystemRole reports whether id matches one of the seeded system role ids.
func IsSystemRole(id string) bool {
	for _, s := range SystemRoleIDs() {
		if id == s {
			return true
		}
	}
	return false
}

var customRoleIDPattern = regexp.MustCompile(`^custom-[a-z0-9][a-z0-9_-]{0,62}$`)

// ValidateRoleID enforces the role id contract: either a system id or a `custom-{slug}` value
// that the rbac_policies CHECK constraint will accept.
func ValidateRoleID(id string) error {
	if IsSystemRole(id) {
		return nil
	}
	if customRoleIDPattern.MatchString(id) {
		return nil
	}
	return fmt.Errorf("role id %q must be a system role or match the custom-{slug} pattern", id)
}

// EnforceAuditInvariant sets the audit resource's create/edit/delete flags to false and fills in
// any missing resource entries with all-false permissions. Implements the section 12.0.4 invariant
// so callers cannot grant write access to audit by mistake.
func EnforceAuditInvariant(p PermissionMatrix) PermissionMatrix {
	out := make(PermissionMatrix, len(AllResources()))
	for _, r := range AllResources() {
		out[r] = p[r]
	}
	audit := out[ResourceAudit]
	out[ResourceAudit] = ResourcePermissions{View: audit.View}
	return out
}

// Allows reports whether the matrix grants the (resource, action) coordinate.
func Allows(p PermissionMatrix, r Resource, a Action) bool {
	rp, ok := p[r]
	if !ok {
		return false
	}
	switch a {
	case ActionView:
		return rp.View
	case ActionCreate:
		return rp.Create
	case ActionEdit:
		return rp.Edit
	case ActionDelete:
		return rp.Delete
	default:
		return false
	}
}

// DefaultPermissionMatrix returns the section 12.1 default matrix for the given system role id.
// The second return value is false when id is not a system role; callers should fall back to the
// stored matrix in that case.
func DefaultPermissionMatrix(roleID string) (PermissionMatrix, bool) {
	switch roleID {
	case string(AppRoleDeveloper):
		return PermissionMatrix{
			ResourceInfrastructure:          {View: true},
			ResourcePipelines:               {View: true},
			ResourceOrganization:            {View: true},
			ResourceSecurity:                {View: true},
			ResourceAudit:                   {},
			ResourceApplications:            {},
			ResourceApplicationRepositories: {},
			ResourceProjects:                {},
			ResourceSCMProviders:            {},
			// dev_requests: route gate 는 view 만 통과. handler 가 row-level filter
			// (`assignee_user_id == actor.login`) 로 추가 제한. ARCH-DREQ-04.
			ResourceDevRequests: {View: true},
		}, true
	case string(AppRoleManager):
		return PermissionMatrix{
			ResourceInfrastructure:          {View: true},
			ResourcePipelines:               {View: true},
			ResourceOrganization:            {View: true},
			ResourceSecurity:                {View: true, Create: true},
			ResourceAudit:                   {View: true},
			ResourceApplications:            {},
			ResourceApplicationRepositories: {},
			ResourceProjects:                {},
			ResourceSCMProviders:            {},
			// dev_requests: developer 와 동일 — view 만, row-level filter 는 handler.
			ResourceDevRequests: {View: true},
		}, true
	case string(AppRoleSystemAdmin):
		return PermissionMatrix{
			ResourceInfrastructure:          {View: true, Create: true, Edit: true, Delete: true},
			ResourcePipelines:               {View: true, Create: true, Edit: true, Delete: true},
			ResourceOrganization:            {View: true, Create: true, Edit: true, Delete: true},
			ResourceSecurity:                {View: true, Create: true, Edit: true, Delete: true},
			ResourceAudit:                   {View: true},
			ResourceApplications:            {View: true, Create: true, Edit: true, Delete: true},
			ResourceApplicationRepositories: {View: true, Create: true, Edit: true, Delete: true},
			ResourceProjects:                {View: true, Create: true, Edit: true, Delete: true},
			ResourceSCMProviders:            {View: true, Create: true, Edit: true, Delete: true},
			ResourceDevRequests:             {View: true, Create: true, Edit: true, Delete: true},
		}, true
	case string(AppRolePMOManager):
		// REQ-FR-PROJ-010 정책 매핑 (sprint claude/work_260515-d):
		//   - applications: 수정만 (View+Edit). create/delete 는 system_admin 만.
		//   - application_repositories: view only (link/unlink 초기 비허용).
		//   - projects: 전체 CRUD (project.manage + project.member.manage 위양).
		//   - scm_providers: view only.
		//   - infrastructure / pipelines / organization / security / audit: view only.
		// row-level owner-self 위양은 enforceRowOwnership helper 가 별도 검증한다 (ADR-0011 §4.2).
		return PermissionMatrix{
			ResourceInfrastructure:          {View: true},
			ResourcePipelines:               {View: true},
			ResourceOrganization:            {View: true},
			ResourceSecurity:                {View: true},
			ResourceAudit:                   {View: true},
			ResourceApplications:            {View: true, Edit: true},
			ResourceApplicationRepositories: {View: true},
			ResourceProjects:                {View: true, Create: true, Edit: true, Delete: true},
			ResourceSCMProviders:            {View: true},
			// dev_requests: view + edit (promote/reject), 단 close/reassign 은 handler
			// 가 추가로 system_admin 검증 (REQ-FR-DREQ-007/008 + ARCH-DREQ-04).
			// create 는 외부 intake auth 경로라 RBAC 외 — false.
			ResourceDevRequests: {View: true, Edit: true},
		}, true
	default:
		return nil, false
	}
}

// SystemRoles returns the seeded role rows that the 000005_create_rbac_policies migration installs.
// The seed migration uses these descriptions verbatim; PR-G3 store seed and tests both consume this
// helper as the single source of truth.
func SystemRoles() []RBACRole {
	roles := make([]RBACRole, 0, 3)
	for _, id := range SystemRoleIDs() {
		matrix, _ := DefaultPermissionMatrix(id)
		roles = append(roles, RBACRole{
			ID:          id,
			Name:        systemRoleName(id),
			Description: systemRoleDescription(id),
			System:      true,
			Permissions: matrix,
		})
	}
	return roles
}

func systemRoleName(id string) string {
	switch id {
	case string(AppRoleDeveloper):
		return "Developer"
	case string(AppRoleManager):
		return "Manager"
	case string(AppRoleSystemAdmin):
		return "System Admin"
	case string(AppRolePMOManager):
		return "PMO Manager"
	default:
		return id
	}
}

func systemRoleDescription(id string) string {
	switch id {
	case string(AppRoleDeveloper):
		return "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한"
	case string(AppRoleManager):
		return "팀 운영, risk triage, 승인 전 command 생성 권한"
	case string(AppRoleSystemAdmin):
		return "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한"
	case string(AppRolePMOManager):
		return "Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지."
	default:
		return ""
	}
}
