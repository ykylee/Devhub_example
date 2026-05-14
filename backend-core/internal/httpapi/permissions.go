package httpapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// PermissionCache holds the rbac_policies matrix in memory so the per-request
// enforceRoutePermission middleware can answer (role, resource, action) lookups
// without touching the database. RBAC policy mutations call Invalidate so the
// next request reloads.
//
// When constructed with a nil store the cache falls back to domain.SystemRoles()
// so tests and dev environments without an RBAC table still enforce the section
// 12.1 default matrix.
type PermissionCache struct {
	mu     sync.RWMutex
	roles  map[string]domain.PermissionMatrix
	store  RBACStore
	loaded bool
}

// NewPermissionCache returns a cache backed by the given store. Pass nil for
// dev/test environments without an rbac_policies table to fall back to the
// section 12.1 default matrix.
func NewPermissionCache(store RBACStore) *PermissionCache {
	return &PermissionCache{store: store}
}

// Allows reports whether the given role grants (resource, action). A role that
// does not exist in the cache (or in the store) yields (false, nil) — deny.
func (p *PermissionCache) Allows(ctx context.Context, role string, r domain.Resource, a domain.Action) (bool, error) {
	p.mu.RLock()
	loaded := p.loaded
	if loaded {
		matrix, ok := p.roles[role]
		p.mu.RUnlock()
		if !ok {
			return false, nil
		}
		return domain.Allows(matrix, r, a), nil
	}
	p.mu.RUnlock()

	if err := p.load(ctx); err != nil {
		return false, err
	}
	return p.Allows(ctx, role, r, a)
}

// Invalidate clears the cached snapshot so the next Allows call reloads.
// Called by the rbac.go mutation handlers after successful policy changes.
func (p *PermissionCache) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loaded = false
	p.roles = nil
}

func (p *PermissionCache) load(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.loaded {
		return nil
	}

	if p.store == nil {
		p.roles = make(map[string]domain.PermissionMatrix, 3)
		for _, role := range domain.SystemRoles() {
			p.roles[role.ID] = copyMatrix(role.Permissions)
		}
		p.loaded = true
		return nil
	}

	roles, err := p.store.ListRBACRoles(ctx)
	if err != nil {
		return fmt.Errorf("permission cache load: %w", err)
	}
	p.roles = make(map[string]domain.PermissionMatrix, len(roles))
	for _, role := range roles {
		p.roles[role.ID] = copyMatrix(role.Permissions)
	}
	p.loaded = true
	return nil
}

// copyMatrix returns a defensive shallow copy. PermissionMatrix is map[Resource]
// ResourcePermissions where ResourcePermissions is a value type with no nested
// reference fields, so a shallow copy is enough to isolate the cache from
// later mutations of the source matrix.
func copyMatrix(m domain.PermissionMatrix) domain.PermissionMatrix {
	out := make(domain.PermissionMatrix, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// routePolicy describes how the enforcement middleware treats a single route.
// Bypass=true means the request is authenticated but no RBAC matrix lookup is
// performed (e.g., self-info, websocket, gitea webhook with HMAC).
type routePolicy struct {
	Resource domain.Resource
	Action   domain.Action
	Bypass   bool
}

type routeKey struct {
	Method string
	Path   string
}

// routePermissionTable encodes docs/backend_api_contract.md section 12.8 as a
// runtime lookup. Adding a new v1 route without a matching entry triggers
// section 12.9 deny-by-default — the route is registered but every request is
// rejected with a 403 + auth.policy_unmapped audit, so omissions are visible
// instead of silent.
var routePermissionTable = map[routeKey]routePolicy{
	// Bypass — section 12.8.1 (auth-only, no matrix lookup)
	{http.MethodGet, "/api/v1/me"}:                           {Bypass: true},
	{http.MethodGet, "/api/v1/realtime/ws"}:                  {Bypass: true},
	{http.MethodPost, "/api/v1/integrations/gitea/webhooks"}:                            {Bypass: true},
	{http.MethodPost, "/api/v1/integrations/kratos/hook/settings/password/after"}:       {Bypass: true},
	// Self-service password change (L4-D, work_26_05_11-e). RBAC matrix is
	// not the right tool here — every authenticated user can change their
	// own password; admin-driven resets go through /accounts/:user_id/password
	// which is already mapped to security:edit.
	{http.MethodPost, "/api/v1/account/password"}: {Bypass: true},
	// Auth proxy endpoints run before the user has a token; Hydra's
	// challenge tokens (single-use, lifespan-bound) protect them.
	{http.MethodPost, "/api/v1/auth/login"}:  {Bypass: true},
	{http.MethodPost, "/api/v1/auth/logout"}: {Bypass: true},
	{http.MethodPost, "/api/v1/auth/token"}:  {Bypass: true},
	{http.MethodPost, "/api/v1/auth/signup"}: {Bypass: true},
	{http.MethodGet, "/api/v1/auth/consent"}: {Bypass: true},

	// infrastructure
	{http.MethodGet, "/api/v1/dashboard/metrics"}:             {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/events"}:                        {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/edges"}:                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/nodes"}:                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/topology"}:                {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/admin/service-actions"}:        {Resource: domain.ResourceInfrastructure, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/commands/:command_id"}:          {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/commands/:command_id/approve"}: {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodPost, "/api/v1/commands/:command_id/reject"}:  {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},

	// pipelines
	{http.MethodGet, "/api/v1/repositories"}:            {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/issues"}:                  {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/pull-requests"}:           {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/ci-runs"}:                 {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/ci-runs/:ci_run_id/logs"}: {Resource: domain.ResourcePipelines, Action: domain.ActionView},

	// security
	{http.MethodGet, "/api/v1/risks"}:                       {Resource: domain.ResourceSecurity, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/risks/critical"}:              {Resource: domain.ResourceSecurity, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/risks/:risk_id/mitigations"}: {Resource: domain.ResourceSecurity, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/rbac/policy"}:                 {Resource: domain.ResourceSecurity, Action: domain.ActionView}, // legacy 410
	{http.MethodGet, "/api/v1/rbac/policies"}:               {Resource: domain.ResourceSecurity, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/rbac/policies"}:              {Resource: domain.ResourceSecurity, Action: domain.ActionEdit},
	{http.MethodPut, "/api/v1/rbac/policies"}:               {Resource: domain.ResourceSecurity, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/rbac/policies/:role_id"}:   {Resource: domain.ResourceSecurity, Action: domain.ActionEdit},

	// audit
	{http.MethodGet, "/api/v1/audit-logs"}: {Resource: domain.ResourceAudit, Action: domain.ActionView},

	// organization — users
	{http.MethodGet, "/api/v1/users"}:             {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/users"}:            {Resource: domain.ResourceOrganization, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/users/:user_id"}:    {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/users/:user_id"}:  {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/users/:user_id"}: {Resource: domain.ResourceOrganization, Action: domain.ActionDelete},

	// account admin (PR-S3) — credential issuance + force reset + state toggle + revoke.
	// Credential mutations are a security-grade resource; route the auth-y ones to
	// security:edit so only system_admin (per default matrix) can hit them.
	{http.MethodPost, "/api/v1/accounts"}:                     {Resource: domain.ResourceSecurity, Action: domain.ActionCreate},
	{http.MethodPut, "/api/v1/accounts/:user_id/password"}:    {Resource: domain.ResourceSecurity, Action: domain.ActionEdit},
	{http.MethodPatch, "/api/v1/accounts/:user_id"}:           {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/accounts/:user_id"}:          {Resource: domain.ResourceOrganization, Action: domain.ActionDelete},

	// organization — units
	{http.MethodGet, "/api/v1/organization/hierarchy"}:              {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPut, "/api/v1/organization/hierarchy"}:              {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/organization/units/:unit_id"}:         {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/organization/units"}:                 {Resource: domain.ResourceOrganization, Action: domain.ActionCreate},
	{http.MethodPatch, "/api/v1/organization/units/:unit_id"}:       {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/organization/units/:unit_id"}:      {Resource: domain.ResourceOrganization, Action: domain.ActionDelete},
	{http.MethodGet, "/api/v1/organization/units/:unit_id/members"}: {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPut, "/api/v1/organization/units/:unit_id/members"}: {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},

	// organization — RBAC subject role assignment
	{http.MethodGet, "/api/v1/rbac/subjects/:subject_id/roles"}: {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPut, "/api/v1/rbac/subjects/:subject_id/roles"}: {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/hr/lookup"}:                        {Resource: domain.ResourceOrganization, Action: domain.ActionView},

	// SCM Provider catalog (API-41..42, sprint claude/work_260514-a, ADR-0011 §4.1 1차).
	{http.MethodGet, "/api/v1/scm/providers"}:                  {Resource: domain.ResourceSCMProviders, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/scm/providers/:provider_key"}:  {Resource: domain.ResourceSCMProviders, Action: domain.ActionEdit},

	// Applications (API-43..47, sprint claude/work_260514-a).
	{http.MethodGet, "/api/v1/applications"}:                          {Resource: domain.ResourceApplications, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/applications"}:                         {Resource: domain.ResourceApplications, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/applications/:application_id"}:          {Resource: domain.ResourceApplications, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/applications/:application_id"}:        {Resource: domain.ResourceApplications, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/applications/:application_id"}:       {Resource: domain.ResourceApplications, Action: domain.ActionDelete},

	// Application-Repository link (API-48..50, sprint claude/work_260514-a).
	{http.MethodGet, "/api/v1/applications/:application_id/repositories"}:                       {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/applications/:application_id/repositories"}:                      {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionCreate},
	{http.MethodDelete, "/api/v1/applications/:application_id/repositories/*repo_key"}:          {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionDelete},

	// Repository 운영 지표 (API-51..54, sprint claude/work_260514-c). 본 endpoint 들은
	// Application 의 연결 Repository 메트릭이므로 application_repositories:view 로 매핑.
	{http.MethodGet, "/api/v1/repositories/:repository_id/activity"}:           {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/pull-requests"}:      {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/build-runs"}:         {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/quality-snapshots"}:  {Resource: domain.ResourceApplicationRepositories, Action: domain.ActionView},

	// Project CRUD (API-55..56, sprint claude/work_260514-c).
	{http.MethodGet, "/api/v1/repositories/:repository_id/projects"}:  {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/repositories/:repository_id/projects"}: {Resource: domain.ResourceProjects, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/projects/:project_id"}:                  {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/projects/:project_id"}:                {Resource: domain.ResourceProjects, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/projects/:project_id"}:               {Resource: domain.ResourceProjects, Action: domain.ActionDelete},

	// Application 롤업 (API-57, sprint claude/work_260514-c) — applications:view 매핑.
	{http.MethodGet, "/api/v1/applications/:application_id/rollup"}: {Resource: domain.ResourceApplications, Action: domain.ActionView},

	// Integration CRUD (API-58, sprint claude/work_260514-c) — applications:edit cross-cut
	// (관리 행위라 admin 일임).
	{http.MethodGet, "/api/v1/integrations"}:                    {Resource: domain.ResourceApplications, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/integrations"}:                   {Resource: domain.ResourceApplications, Action: domain.ActionEdit},
	{http.MethodPatch, "/api/v1/integrations/:integration_id"}:  {Resource: domain.ResourceApplications, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/integrations/:integration_id"}: {Resource: domain.ResourceApplications, Action: domain.ActionEdit},
}

// lookupRoutePolicy is exported for tests to assert the table contents without
// reaching into the unexported map.
func lookupRoutePolicy(method, path string) (routePolicy, bool) {
	policy, ok := routePermissionTable[routeKey{Method: method, Path: path}]
	return policy, ok
}

// enforceRoutePermission is the v1 group middleware that resolves the request
// route against routePermissionTable, looks up the actor's role in the
// PermissionCache, and applies section 12.9 deny-by-default for unmapped
// routes.
func (h Handler) enforceRoutePermission(c *gin.Context) {
	if devFallbackEnabled(c) {
		c.Next()
		return
	}

	policy, mapped := lookupRoutePolicy(c.Request.Method, c.FullPath())
	if !mapped {
		h.recordAuditBestEffort(c, "auth.policy_unmapped", "route", c.FullPath(), map[string]any{
			"method": c.Request.Method,
		})
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "route is not mapped to an RBAC permission",
			"code":   "auth_policy_unmapped",
		})
		return
	}
	if policy.Bypass {
		c.Next()
		return
	}

	actorValue, _ := c.Get("devhub_actor_role")
	actorRole, _ := actorValue.(string)

	cache := h.cfg.PermissionCache
	if cache == nil {
		// NewRouter installs a default cache, so this branch should only fire
		// when a caller bypasses NewRouter and constructs Handler directly.
		// Falling back to a fresh default cache here keeps behavior consistent
		// with the section 12.1 default matrix.
		cache = NewPermissionCache(nil)
	}

	allowed, err := cache.Allows(c.Request.Context(), actorRole, policy.Resource, policy.Action)
	if err != nil {
		log.Printf("server error: op=auth.permission_check request_id=%s err=%v", requestIDFrom(c), err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"error":  "internal error",
		})
		return
	}
	if !allowed {
		h.recordAuditBestEffort(c, "auth.role_denied", "route", c.FullPath(), map[string]any{
			"actor_role": actorRole,
			"resource":   string(policy.Resource),
			"action":     string(policy.Action),
		})
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  fmt.Sprintf("role %q lacks %s:%s permission", actorRole, policy.Resource, policy.Action),
		})
		return
	}
	c.Next()
}
