package view

import (
	"context"
	"fmt"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/shared/metrics"
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
	{http.MethodGet, "/api/v1/me"}:           {Bypass: true},
	{http.MethodPost, "/api/v1/auth/logout"}: {Bypass: true},
	// RM-ONBOARD-01 (ADR-0021 §3.5, §16.3..16.5). onboardingGate middleware
	// 가 미완료 사용자의 본 endpoint 외 endpoint 호출 시 403 처리.
	{http.MethodPatch, "/api/v1/me"}:                 {Bypass: true},
	{http.MethodPost, "/api/v1/me/onboarding"}:       {Bypass: true},
	{http.MethodGet, "/api/v1/organizations/search"}: {Bypass: true},
	// RM-ONBOARD-01 (API-86) — admin 의 review confirm. system_admin 일임 —
	// rbac matrix 의 organization:edit 정합 (user/organization management).
	{http.MethodPost, "/api/v1/admin/users/:user_id/review"}: {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/realtime/ws"}:                  {Bypass: true},
	// ADR-0024 §3.2 ticket pattern — authenticateActor 가 Bearer 검증 후 발급.
	// RBAC bypass (인증된 actor 면 누구나 short-lived ticket 발급 권한).
	{http.MethodPost, "/api/v1/realtime/ticket"}:                            {Bypass: true},
	{http.MethodPost, "/api/v1/integrations/gitea/webhooks"}:                {Bypass: true},
	{http.MethodPost, "/api/v1/integration/providers/:provider_id/webhook"}: {Bypass: true},
	{http.MethodPost, "/api/v1/infra/services/snapshot"}:                    {Bypass: true},
	// Keycloak Event Listener SPI webhook (PR #203). Bypass — registered outside
	// the v1 group (router.go) so authenticateActor + enforceRoutePermission do
	// not run; secret verification via X-Webhook-Secret header is fail-closed in
	// receiveKeycloakEventWebhook (sprint -d Stage 3 codex P1 hotfix). Entry
	// kept here so TestRoutePermissionTable_CoversAllProtectedV1Routes recognizes
	// this path as authorized.
	{http.MethodPost, "/api/v1/internal/keycloak-events"}: {Bypass: true},

	// infrastructure
	{http.MethodGet, "/api/v1/dashboard/metrics"}:             {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/events"}:                        {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/edges"}:                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/nodes"}:                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/services"}:                {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/topology"}:                {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/infra/topology/v2"}:             {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/admin/service-actions"}:        {Resource: domain.ResourceInfrastructure, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/commands/:command_id"}:          {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/commands/:command_id/approve"}: {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodPost, "/api/v1/commands/:command_id/reject"}:  {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},

	// pipelines
	{http.MethodGet, "/api/v1/repositories"}:                         {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/repositories"}:                        {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionCreate},
	{http.MethodPatch, "/api/v1/repositories/:repository_id"}:         {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/repositories/:repository_id"}:        {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionDelete},
	{http.MethodPost, "/api/v1/repositories/:repository_id/publish"}: {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/issues"}:                               {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/pull-requests"}:                        {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/ci-runs"}:                              {Resource: domain.ResourcePipelines, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/ci-runs"}:                             {Resource: domain.ResourcePipelines, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/ci-runs/:ci_run_id/logs"}:              {Resource: domain.ResourcePipelines, Action: domain.ActionView},

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

	// ADR-0020 sub-carve B (sprint -i, issue #209): /api/v1/accounts/* 4 entry
	// 제거. user 생성/비밀번호/상태/삭제는 Keycloak admin console 또는 HRDB ETL
	// push 책임. lazy auto-create 가 authenticateActor 에서 자동 처리 (§5.2).
	// 이전: POST /accounts (security:create), PUT /accounts/:user_id/password
	// (security:edit), PATCH /accounts/:user_id (organization:edit), DELETE
	// /accounts/:user_id (organization:delete).

	// organization — units
	{http.MethodGet, "/api/v1/organization/hierarchy"}:              {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPut, "/api/v1/organization/hierarchy"}:              {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/organization/units/:unit_id"}:         {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/organization/units"}:                 {Resource: domain.ResourceOrganization, Action: domain.ActionCreate},
	{http.MethodPatch, "/api/v1/organization/units/:unit_id"}:       {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/organization/units/:unit_id"}:      {Resource: domain.ResourceOrganization, Action: domain.ActionDelete},
	{http.MethodGet, "/api/v1/organization/units/:unit_id/members"}: {Resource: domain.ResourceOrganization, Action: domain.ActionView},
	{http.MethodPut, "/api/v1/organization/units/:unit_id/members"}: {Resource: domain.ResourceOrganization, Action: domain.ActionEdit},

	// organization — HR lookup. ADR-0020 (sprint -d): rbac/subjects/:subject_id/roles
	// 2 entry 제거 — backend-only dead-end (frontend UI 미구현). users.role 직접 write
	// 였고 Keycloak group composite 가 실 권한 source. PATCH /api/v1/users/:id 와 중복.
	{http.MethodGet, "/api/v1/hr/lookup"}: {Resource: domain.ResourceOrganization, Action: domain.ActionView},

	// SCM Provider catalog (API-41..42, sprint claude/work_260514-a, ADR-0011 §4.1 1차).
	{http.MethodGet, "/api/v1/scm/providers"}:                 {Resource: domain.ResourceSCMProviders, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/scm/providers/:provider_key"}: {Resource: domain.ResourceSCMProviders, Action: domain.ActionEdit},

	// Applications (API-43..47, sprint claude/work_260514-a).
	{http.MethodGet, "/api/v1/platforms"}:                    {Resource: domain.ResourcePlatforms, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/platforms"}:                   {Resource: domain.ResourcePlatforms, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/platforms/:platform_id"}:    {Resource: domain.ResourcePlatforms, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/platforms/:platform_id"}:  {Resource: domain.ResourcePlatforms, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/platforms/:platform_id"}: {Resource: domain.ResourcePlatforms, Action: domain.ActionDelete},

	// Application-Repository link (API-48..50, sprint claude/work_260514-a).
	{http.MethodGet, "/api/v1/platforms/:platform_id/repositories"}:              {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/platforms/:platform_id/repositories"}:             {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionCreate},
	{http.MethodDelete, "/api/v1/platforms/:platform_id/repositories/*repo_key"}: {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionDelete},

	// Repository 운영 지표 (API-51..54, sprint claude/work_260514-c). 본 endpoint 들은
	// Application 의 연결 Repository 메트릭이므로 application_repositories:view 로 매핑.
	{http.MethodGet, "/api/v1/repositories/:repository_id/activity"}:          {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/pull-requests"}:     {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/build-runs"}:        {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/quality-snapshots"}: {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	// Sprint A — kpi-tests-per-domain-scope.md §6.1 (Repository sub-section).
	// 신규 2 endpoint 의 RBAC 정합. Repository view 권한 보유 actor 만 접근 가능.
	{http.MethodGet, "/api/v1/repositories/:repository_id/kpi"}:           {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/repositories/:repository_id/test-results"}: {Resource: domain.ResourcePlatformRepositories, Action: domain.ActionView},
	// Project KPI 가중치 rollup (Sprint B — kpi-tests-per-domain-scope.md §6.2).
	// projectTestResults 는 follow-up PR (Sprint B-Tests) 에서 추가. deny-by-default
	// (section 12.9) 회귀 가드 정합: 신규 route 의 row 미등록 시 authenticated 요청도
	// 403 + auth.policy_unmapped reject.
	{http.MethodGet, "/api/v1/projects/:project_id/kpi"}: {Resource: domain.ResourceProjects, Action: domain.ActionView},
	// Project Tests 가중치 종합 (Sprint B-Tests — kpi-tests-per-domain-scope.md
	// §6.2 follow-up). projectKPI 와 동일 Resource/Action (projects:view). 신규
	// route 의 row 미등록 시 deny-by-default 회귀 가드 정합.
	{http.MethodGet, "/api/v1/projects/:project_id/test-results"}: {Resource: domain.ResourceProjects, Action: domain.ActionView},
	// Platform sub-project rollup (Sprint C — kpi-tests-per-domain-scope.md §6.3).
	// platformKPI + platformTestResults 동일 Resource/Action (platforms:view,
	// 기존 /platforms/:id/rollup 와 정합). 신규 route 의 row 미등록 시 deny-by-default
	// 회귀 가드 정합.
	{http.MethodGet, "/api/v1/platforms/:platform_id/kpi"}:           {Resource: domain.ResourcePlatforms, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/platforms/:platform_id/test-results"}: {Resource: domain.ResourcePlatforms, Action: domain.ActionView},

	// Project CRUD (API-55..56, sprint claude/work_260514-c).
	{http.MethodGet, "/api/v1/repositories/:repository_id/projects"}:                {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/repositories/:repository_id/projects"}:               {Resource: domain.ResourceProjects, Action: domain.ActionCreate},
	{http.MethodPost, "/api/v1/projects"}:                                           {Resource: domain.ResourceProjects, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/platforms/:platform_id/projects"}:               {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/platforms/:platform_id/projects"}:              {Resource: domain.ResourceProjects, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/projects/standalone"}:                                 {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/projects/:project_id"}:                                {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/projects/:project_id"}:                              {Resource: domain.ResourceProjects, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/projects/:project_id"}:                             {Resource: domain.ResourceProjects, Action: domain.ActionDelete},
	{http.MethodGet, "/api/v1/projects/:project_id/repositories"}:                   {Resource: domain.ResourceProjects, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/projects/:project_id/repositories"}:                  {Resource: domain.ResourceProjects, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/projects/:project_id/repositories/:repository_id"}: {Resource: domain.ResourceProjects, Action: domain.ActionDelete},
	{http.MethodPatch, "/api/v1/projects/:project_id/repositories/:repository_id"}:  {Resource: domain.ResourceProjects, Action: domain.ActionEdit},

	// Application 롤업 (API-57, sprint claude/work_260514-c) — applications:view 매핑.
	{http.MethodGet, "/api/v1/platforms/:platform_id/rollup"}:    {Resource: domain.ResourcePlatforms, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/platforms/:platform_id/dashboard"}: {Resource: domain.ResourcePlatforms, Action: domain.ActionView},

	// Integration CRUD (API-58, sprint claude/work_260514-c) — applications:edit cross-cut
	// (관리 행위라 admin 일임).
	{http.MethodGet, "/api/v1/integrations"}:                                            {Resource: domain.ResourcePlatforms, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/integrations"}:                                           {Resource: domain.ResourcePlatforms, Action: domain.ActionEdit},
	{http.MethodPatch, "/api/v1/integrations/:integration_id"}:                          {Resource: domain.ResourcePlatforms, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/integrations/:integration_id"}:                         {Resource: domain.ResourcePlatforms, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/integration/providers"}:                                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/integration/providers"}:                                  {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodPatch, "/api/v1/integration/providers/:provider_id"}:                    {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/integration/providers/:provider_id"}:                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionDelete},
	{http.MethodPost, "/api/v1/integration/providers/:provider_id/sync"}:                {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/integration/providers/:provider_id/scm-repositories"}:     {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/integration/providers/:provider_id/import-repositories"}: {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodPost, "/api/v1/integration/providers/:provider_id/create-repository"}:   {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodPost, "/api/v1/integration/test-connection"}:                            {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/integration/bindings"}:                                    {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/integration/bindings"}:                                   {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/admin/integrations/sync-jobs"}:                            {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/admin/integrations/sync-jobs/:jobID"}:                     {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/admin/integrations/summary"}:                              {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	// PR #251 P2-4 sub-carve — Bindings UI 강화. API-81 PATCH + API-82 DELETE.
	{http.MethodPatch, "/api/v1/integration/bindings/:binding_id"}:  {Resource: domain.ResourceInfrastructure, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/integration/bindings/:binding_id"}: {Resource: domain.ResourceInfrastructure, Action: domain.ActionDelete},

	// Dev Request (DREQ) API-59..65 (sprint claude/work_260515-i, ADR-0012).
	// POST /api/v1/dev-requests 는 별도 intake group 으로 등록되어 v1 의
	// authenticateActor + enforceRoutePermission 가 적용되지 않으나, router.Routes()
	// 가 모든 라우트를 반환하므로 routePermissionTable 검증 테스트 통과 위해
	// Bypass: true 로 등록한다 (intake 인증은 requireIntakeToken 미들웨어가 책임).
	{http.MethodPost, "/api/v1/dev-requests"}:                          {Bypass: true},
	{http.MethodGet, "/api/v1/dev-requests"}:                           {Resource: domain.ResourceDevRequests, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/dev-requests/:dev_request_id"}:           {Resource: domain.ResourceDevRequests, Action: domain.ActionView},
	{http.MethodPost, "/api/v1/dev-requests/:dev_request_id/register"}: {Resource: domain.ResourceDevRequests, Action: domain.ActionEdit},
	{http.MethodPost, "/api/v1/dev-requests/:dev_request_id/reject"}:   {Resource: domain.ResourceDevRequests, Action: domain.ActionEdit},
	{http.MethodPatch, "/api/v1/dev-requests/:dev_request_id"}:         {Resource: domain.ResourceDevRequests, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/dev-requests/:dev_request_id"}:        {Resource: domain.ResourceDevRequests, Action: domain.ActionDelete},
	// ADR-0028 §3: voc (voice of customer) 도메인 — external_ref 기반 dev-request staging.
	// POST /dev-requests/:dev_request_id 는 외부 intake 의 path (사용자 명시) — handler
	// 는 createOrGetVoc 으로 dispatch. system_admin 일임 (external system source).
	{http.MethodPost, "/api/v1/dev-requests/:dev_request_id"}:          {Resource: domain.ResourceDevRequests, Action: domain.ActionCreate},
	{http.MethodPost, "/api/v1/dev-requests/:dev_request_id/route"}:   {Resource: domain.ResourceDevRequests, Action: domain.ActionEdit},
	{http.MethodGet, "/api/v1/dev-requests/external/:external_ref"}:   {Resource: domain.ResourceDevRequests, Action: domain.ActionView},
	// ADR-0028 §6 carve (d): voc list — system_admin 도구, N-6 staging 운영 SOP 정합.
	{http.MethodGet, "/api/v1/vocs"}: {Resource: domain.ResourceDevRequests, Action: domain.ActionView},
	// ADR-0028 §3: in-app notification — 자기 자신 조회/마킹. Bypass (R-9 §3 자기 정보).
	{http.MethodGet, "/api/v1/me/notifications"}:           {Bypass: true},
	{http.MethodPost, "/api/v1/me/notifications/:id/read"}:  {Bypass: true},

	// DREQ intake token admin (sprint claude/work_260515-o, ADR-0014). system_admin 일임.
	{http.MethodPost, "/api/v1/dev-request-tokens"}:             {Resource: domain.ResourceDevRequestIntakeTokens, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/dev-request-tokens"}:              {Resource: domain.ResourceDevRequestIntakeTokens, Action: domain.ActionView},
	{http.MethodDelete, "/api/v1/dev-request-tokens/:token_id"}: {Resource: domain.ResourceDevRequestIntakeTokens, Action: domain.ActionDelete},
	{http.MethodPatch, "/api/v1/dev-request-tokens/:token_id"}:  {Resource: domain.ResourceDevRequestIntakeTokens, Action: domain.ActionEdit},

	// Task Item Ingestion (sprint deepseek/work_260528-a-task-item-ingestion, API-94..96).
	// Webhook 수신 — Bypass (webhook_secret 인증, authenticateActor + RBAC 불필요).
	{http.MethodPost, "/api/v1/integration/providers/:provider_id/tasks/webhook"}: {Bypass: true},
	// Task item 조회 — infrastructure:view (provider 관리자 권한).
	{http.MethodGet, "/api/v1/external-tasks"}:          {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},
	{http.MethodGet, "/api/v1/external-tasks/:task_id"}: {Resource: domain.ResourceInfrastructure, Action: domain.ActionView},

	// ADR-0029 §6 (f) P3 — multi-key 관리 (sprint feat/work_260609-k-api-key-management).
	// system_admin 일임 (api_keys resource). raw key 1회 응답 (POST) + list/revoke/update
	// 는 key_prefix 만 노출 (보안). 자세한 동작: [`docs/planning/api-key-management-sprint-plan.md` §3.3].
	{http.MethodPost, "/api/v1/admin/api-keys"}:             {Resource: domain.ResourceAPIKeys, Action: domain.ActionCreate},
	{http.MethodGet, "/api/v1/admin/api-keys"}:              {Resource: domain.ResourceAPIKeys, Action: domain.ActionView},
	{http.MethodPatch, "/api/v1/admin/api-keys/:api_key_id"}:  {Resource: domain.ResourceAPIKeys, Action: domain.ActionEdit},
	{http.MethodDelete, "/api/v1/admin/api-keys/:api_key_id"}: {Resource: domain.ResourceAPIKeys, Action: domain.ActionDelete},
}

// lookupRoutePolicy is exported for tests to assert the table contents without
// reaching into the unexported map.
func lookupRoutePolicy(method, path string) (routePolicy, bool) {
	policy, ok := routePermissionTable[routeKey{Method: method, Path: path}]
	return policy, ok
}

// RoutePolicy / LookupRoutePolicy — cross-package test 접근용 export shim.
// docs/governance/code-taxonomy.md 의 도메인 분리 이후 httpapi/_test 가 routePermissionTable
// 정합을 검증해야 하므로 export 표면을 최소 노출. production code 가 이 alias 를
// 직접 참조하지 않도록 유지.
type RoutePolicy = routePolicy

func LookupRoutePolicy(method, path string) (RoutePolicy, bool) {
	return lookupRoutePolicy(method, path)
}

// enforceRowOwnership는 ADR-0011 §4.2 의 row-level 위양 진입점. caller 가
// ownerUserID 의 row 에 대해 쓰기 권한을 가지는지를 다음 규칙으로 결정한다:
//
//  1. actor.role == "system_admin"  (전역 일임 — 항상 통과)
//  2. actor.role ∈ allowedRoles      (예: "team_manager" — 화이트리스트)
//  3. actor.login == ownerUserID    (owner-self)
//
// 한 가지라도 만족하면 true 반환. 만족 못 하면 audit `auth.row_denied` 를
// emit 하고 403 으로 abort 한 뒤 false 반환. caller 는 단순히
// `if !h.enforceRowOwnership(c, app.OwnerUserID, "team_manager") { return }`
// 패턴으로 사용한다.
//
// audit payload (ADR-0011 §6):
//
//	{
//	  "actor_role":    <actor role>,
//	  "owner_user_id": <ownerUserID>,
//	  "resource":      <route policy resource>,
//	  "action":        <route policy action>,
//	  "denied_reason": "owner_mismatch"
//	}
//
// resource/action 은 enforceRoutePermission 이 이미 통과시킨 route 의 매핑에서
// 추출 — 미매핑 route 라면 ""로 채운다 (audit consumer 가 N/A 로 해석).
//
// EnforceRowOwnership — enforceRowOwnership 의 cross-package test 접근용 export
// wrapper. 동작/규약 동일.
func (h *RBACHandler) EnforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
	return h.enforceRowOwnership(c, ownerUserID, allowedRoles...)
}

// ownerUserID 가 "" 이면 owner-self 규칙은 비활성화 (system_admin / allowedRoles
// 만 통과). 잘못된 데이터(미설정 owner) 가 우연히 익명에게 허용되는 일을 방지.
func (h *RBACHandler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
	// dev fallback (AuthDevFallback=true) 환경은 actor 가 컨텍스트에 주입되지
	// 않으므로 enforceRoutePermission 과 동일하게 bypass — 그러지 않으면 핸들러
	// 단위 테스트가 모두 403 으로 깨진다. 운영에서는 devFallbackEnabled=false 라
	// 정상 평가 흐름을 탄다.
	if httphelp.DevFallbackEnabled(c) {
		return true
	}

	loginVal, _ := c.Get("devhub_actor_login")
	roleVal, _ := c.Get("devhub_actor_role")
	actorLogin, _ := loginVal.(string)
	actorRole, _ := roleVal.(string)

	if actorRole == string(domain.AppRoleSystemAdmin) {
		return true
	}
	for _, allowed := range allowedRoles {
		if actorRole == allowed {
			return true
		}
	}
	if ownerUserID != "" && actorLogin == ownerUserID {
		return true
	}

	policy, _ := lookupRoutePolicy(c.Request.Method, c.FullPath())
	h.recordAuditBestEffort(c, "auth.row_denied", string(policy.Resource), c.FullPath(), map[string]any{
		"actor_role":    actorRole,
		"owner_user_id": ownerUserID,
		"resource":      string(policy.Resource),
		"action":        string(policy.Action),
		"denied_reason": "owner_mismatch",
	})
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status": "forbidden",
		"error":  "owner_mismatch — row write requires owner or elevated role",
		"code":   "auth_row_denied",
	})
	return false
}

// enforceRoutePermission is the v1 group middleware that resolves the request
// route against routePermissionTable, looks up the actor's role in the
// PermissionCache, and applies section 12.9 deny-by-default for unmapped
// routes.
func (h *RBACHandler) EnforceRoutePermission(c *gin.Context) {
	if httphelp.DevFallbackEnabled(c) {
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

	// ADR-0029 §6 (a) P0 — API key caller 의 admin endpoint RBAC 가드. ADR-0029
	// API key 는 두 경로 — §3.4 정적 키 (DEVHUB_API_KEY, source="api_key") 와
	// PR #528 §6 (f) DB multi-key (source="api_key_db"). 두 경로 모두
	// §2.2 옵션 B (공개 read-only 만 API key 허용, admin API 는 Keycloak 강제)
	// 동일 SOP 적용. authenticateActor 가 위 두 source 중 하나를 set 한 경우,
	// policy.Action 이 View 가 아닌 (mutation) 경로만 차단 — read-only
	// (infrastructure:view 등) 는 그대로 통과. cache.Allows 호출 직전에
	// 위치하여 role 매트릭스 가드와 독립적으로 enforce. 운영 SOP
	// (DEVHUB_API_KEY 는 staging/dev 전용) 와 백엔드 RBAC 가드의 2중 방어.
	source, _ := c.Get("devhub_auth_source")
	isAPIKeyCaller := source == "api_key" || source == "api_key_db"
	if isAPIKeyCaller && policy.Action != domain.ActionView {
		h.recordAuditBestEffort(c, "auth.api_key_denied", "route", c.FullPath(), map[string]any{
			"actor_role":  c.GetString("devhub_actor_role"),
			"auth_source": source,
			"resource":    string(policy.Resource),
			"action":      string(policy.Action),
			"reason":      "admin_gate_mutation",
			"method":      c.Request.Method,
			"client_ip":   c.ClientIP(),
			"request_id":  httphelp.RequestIDFrom(c),
		})
		// SOP §6.1 metric 정합 — auth denied counter (admin gate).
		metrics.DevhubAPIKeyAuthTotal.WithLabelValues("denied").Inc()
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "api key caller is restricted to read-only endpoints; admin endpoints require Keycloak authentication (ADR-0029 §6 (a))",
			"code":   "auth_api_key_denied",
		})
		return
	}

	if driftValue, _ := c.Get("devhub_role_sync_required"); driftValue == true {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "role synchronization required",
			"code":   "auth.role_sync_required",
		})
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
		log.Printf("server error: op=auth.permission_check request_id=%s err=%v", httphelp.RequestIDFrom(c), err)
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
