package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	appview "github.com/devhub/backend-core/internal/domain/application-lifecycle/view"
	auditview "github.com/devhub/backend-core/internal/domain/audit-ops/view"
	authview "github.com/devhub/backend-core/internal/domain/auth-session/view"
	devreqview "github.com/devhub/backend-core/internal/domain/dev-request/view"
	integview "github.com/devhub/backend-core/internal/domain/integration-registry/view"
	onboardview "github.com/devhub/backend-core/internal/domain/onboarding/view"
	orgview "github.com/devhub/backend-core/internal/domain/organization-management/view"
	rbacview "github.com/devhub/backend-core/internal/domain/rbac-permissions/view"
	realtimeview "github.com/devhub/backend-core/internal/domain/realtime/view"
	repoview "github.com/devhub/backend-core/internal/domain/repository-integration/view"
	ci "github.com/devhub/backend-core/internal/infrastructure/ci"
	gitea "github.com/devhub/backend-core/internal/infrastructure/gitea"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AuthenticatedActor = authview.AuthenticatedActor
type BearerTokenVerifier = authview.BearerTokenVerifier
type OrganizationStore = orgview.OrganizationStore

// ErrInvalidBearerToken — auth-session/view 의 동명 var 를 httpapi 패키지에서
// 호환 노출 (테스트 + 외부 호출자 정합). Go 의 var 는 직접 type-alias 불가하므로
// re-binding 으로 노출.
var ErrInvalidBearerToken = authview.ErrInvalidBearerToken

// PermissionCache / NewPermissionCache — rbac-permissions/view 호환 노출. 기존
// httpapi 테스트가 직접 호출하므로 alias + thin wrapper 유지.
type PermissionCache = rbacview.PermissionCache

func NewPermissionCache(store RBACStore) *PermissionCache {
	return rbacview.NewPermissionCache(store)
}

type WebhookEventStore interface {
	SaveWebhookEvent(context.Context, store.WebhookEvent) (int64, error)
	ListWebhookEvents(context.Context, store.ListWebhookEventsOptions) ([]store.WebhookEvent, error)
}

type WebhookEventProcessor interface {
	Process(context.Context, store.WebhookEvent) error
}

type HealthStore interface {
	Ping(context.Context) error
}

type DomainStore interface {
	ListRepositories(context.Context, domain.ListOptions) ([]domain.Repository, error)
	ListIssues(context.Context, domain.ListOptions) ([]domain.Issue, error)
	ListPullRequests(context.Context, domain.ListOptions) ([]domain.PullRequest, error)
	ListCIRuns(context.Context, domain.ListOptions) ([]domain.CIRun, error)
	ListRisks(context.Context, domain.ListOptions) ([]domain.Risk, error)
}

type CommandStore interface {
	CreateRiskMitigationCommand(context.Context, domain.RiskMitigationCommandRequest) (domain.Command, domain.AuditLog, bool, error)
	CreateServiceActionCommand(context.Context, domain.ServiceActionCommandRequest) (domain.Command, domain.AuditLog, bool, error)
	GetCommand(context.Context, string) (domain.Command, error)
	ApproveCommand(context.Context, domain.CommandApprovalRequest) (domain.Command, domain.AuditLog, error)
	RejectCommand(context.Context, domain.CommandApprovalRequest) (domain.Command, domain.AuditLog, error)
}

type AuditStore interface {
	CreateAuditLog(context.Context, domain.AuditLog) (domain.AuditLog, error)
	ListAuditLogs(context.Context, store.ListAuditLogsOptions) ([]domain.AuditLog, error)
}

// PlatformStore — application-lifecycle 도메인 persistence 컨트랙트
// (API-41..57). Implemented by *store.PostgresStore. Sprint claude/work_260514-b
// (Application Design 2차) 가 stub → body 로 교체. issue #421/#422 (sprint
// claude/work_260529-n) 에서 integration CRUD 13 메서드를 `IntegrationStore`
// 로 분리해 cross-domain bloat 정정. PlatformStore = app-lifecycle alias,
// IntegrationStore = integration-registry alias.
type PlatformStore = appview.PlatformStore

// IntegrationStore — integration-registry 도메인 persistence 컨트랙트
// (API-58, API-69..75). issue #421/#422 (sprint claude/work_260529-n) 에서
// 기존 PlatformStore 의 integration CRUD 13 메서드를 본 interface 로 이관.
type IntegrationStore = integview.IntegrationStore

// IdentityAdmin — ADR-0020 sub-carve E (sprint -n) — Keycloak admin = 별도
// 운영팀 (PoLP). write methods 제거 (CreateIdentity / UpdateIdentityPassword /
// SetIdentityState / DeleteIdentity). service account 는 view-users +
// view-events 만 요구.
type IdentityAdmin interface {
	FindIdentityByUserID(ctx context.Context, userID string) (string, error)
	LogoutUserSession(ctx context.Context, identityID string) error
}

type DevRequestStore interface {
	CreateDevRequest(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error)
	GetDevRequest(ctx context.Context, id string) (domain.DevRequest, error)
	GetDevRequestByExternalRef(ctx context.Context, system, ref string) (domain.DevRequest, error)
	ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
	TransitionDevRequestStatus(ctx context.Context, id string, status domain.DevRequestStatus, reason string) (domain.DevRequest, error)
	ReassignDevRequest(ctx context.Context, id string, assignee string) (domain.DevRequest, error)
	MarkDevRequestRegistered(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error)
	RegisterDevRequestWithNewPlatform(ctx context.Context, id string, app domain.Platform, repo *domain.PlatformRepository) (domain.DevRequest, domain.Platform, error)
	RegisterDevRequestWithNewProject(ctx context.Context, id string, proj domain.Project) (domain.DevRequest, domain.Project, error)
}

type IntakeTokenStore = devreqview.IntakeTokenStore

type RBACStore = rbacview.RBACStore

type ExternalTaskStore interface {
	UpsertExternalTaskItem(ctx context.Context, t domain.ExternalTaskItem) (domain.ExternalTaskItem, error)
	SoftDeleteExternalTaskItem(ctx context.Context, providerID, externalID string) error
	ListExternalTaskItems(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error)
	GetExternalTaskItemByID(ctx context.Context, id string) (domain.ExternalTaskItem, error)
	NextWebhookSeq(ctx context.Context) (int64, error)
	DetectWebhookSeqGaps(ctx context.Context, providerID string) (int64, error)
	UpdateProviderLastPulledAt(ctx context.Context, providerID string, pulledAt time.Time) error
	ListTaskTrackers(ctx context.Context) ([]domain.IntegrationProvider, error)
}

type HRDBClient interface {
	Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error) // simplified for now: returns email, userID, dept
}

type RouterConfig struct {
	WebhookSecret         string
	KeycloakWebhookSecret string
	InfraAgentToken       string
	HomeLabProviderKey    string
	HomeLabDegradedRaw    string
	EventStore            WebhookEventStore
	EventProcessor        WebhookEventProcessor
	HealthStore           HealthStore
	DomainStore           DomainStore
	CommandStore          CommandStore
	AuditStore            AuditStore
	BearerTokenVerifier   BearerTokenVerifier
	OrganizationStore     OrganizationStore
	PlatformStore      PlatformStore
	// IntegrationStore — integration-registry 도메인 (API-58, API-69..75). issue
	// #421/#422 (sprint claude/work_260529-n) 에서 PlatformStore 와 분리. nil
	// 이면 기존 fallback (PlatformStore 가 IntegrationStore 도 구현하는 경우)
	// 으로 동작 — main.go 가 동일 *PostgresStore 를 양쪽에 주입하던 legacy 호환.
	IntegrationStore IntegrationStore
	// DevRequestStore + DevRequestIntakeTokenStore — DREQ 도메인 (ADR-0012, sprint claude/work_260515-i).
	DevRequestStore            DevRequestStore
	DevRequestIntakeTokenStore IntakeTokenStore
	RBACStore                  RBACStore
	PermissionCache            *rbacview.PermissionCache
	ExternalTaskStore          ExternalTaskStore
	IdentityAdmin              IdentityAdmin
	IdPProvider                string
	HRDB                       HRDBClient
	SnapshotProvider           SnapshotProvider
	RealtimeHub                *realtimeview.RealtimeHub
	// RealtimeTickets — ADR-0024 §3.2 ticket pattern.
	RealtimeTickets realtimeview.RealtimeTicketStore
	// AuthDevFallback toggles dev-only authentication fallbacks: empty Authorization passes through authenticateActor and requireMinRole. Actor identity always resolves to "system" without a verifier. Default false: production-safe.
	AuthDevFallback bool
	// OnboardingGateEnabled — RM-ONBOARD-01 (ADR-0021 §3.3, ARCH-ONBOARD-03).
	OnboardingGateEnabled bool
	// ProjectModel toggles project-management route mode: legacy|hybrid|v2.
	ProjectModel string
	CIAdapter    ci.Adapter
}

// resolveIntegrationStore — issue #421/#422 (sprint claude/work_260529-n) 의
// fallback. RouterConfig.IntegrationStore 가 명시되지 않으면 PlatformStore
// 가 IntegrationStore 도 구현하는지 type-assertion 으로 확인 후 사용. 양쪽 모두
// 미설정/미충족 이면 nil 반환 → IntegrationHandler 가 503 unavailable 응답.
func resolveIntegrationStore(cfg RouterConfig) IntegrationStore {
	if cfg.IntegrationStore != nil {
		return cfg.IntegrationStore
	}
	if cfg.PlatformStore == nil {
		return nil
	}
	if is, ok := any(cfg.PlatformStore).(IntegrationStore); ok {
		return is
	}
	return nil
}

// repoIntegrationStoreAdapter — repository-integration/view 의 store interface
// (PlatformStore 명) 는 GetIntegrationProviderByID + ListRepositoriesByProvider
// + UpsertRepository 3 메서드만 요구한다. issue #421/#422 (sprint
// claude/work_260529-n) 분리 후 application-lifecycle 의 PlatformStore 가
// integration 메서드를 잃었으므로, 두 store 를 합쳐 repository-integration
// PlatformStore 컨트랙트를 만족하는 어댑터를 제공한다.
type repoIntegrationStoreAdapter struct {
	platformStore   PlatformStore
	integStore IntegrationStore
}

func (a repoIntegrationStoreAdapter) GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error) {
	return a.integStore.GetIntegrationProviderByID(ctx, id)
}

func (a repoIntegrationStoreAdapter) ListRepositoriesByProvider(ctx context.Context, providerID string) ([]domain.Repository, error) {
	return a.platformStore.ListRepositoriesByProvider(ctx, providerID)
}

func (a repoIntegrationStoreAdapter) UpsertRepository(ctx context.Context, r domain.Repository) error {
	return a.platformStore.UpsertRepository(ctx, r)
}

// resolveRepoIntegrationStore — repository-integration/view 에 주입할 어댑터를
// 구성. PlatformStore 또는 IntegrationStore 가 nil 이면 nil 반환 →
// RepositoryIntegrationHandler 가 503 응답.
func resolveRepoIntegrationStore(cfg RouterConfig) repoview.PlatformStore {
	integStore := resolveIntegrationStore(cfg)
	if cfg.PlatformStore == nil || integStore == nil {
		return nil
	}
	return repoIntegrationStoreAdapter{platformStore: cfg.PlatformStore, integStore: integStore}
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	if err := router.SetTrustedProxies(trustedProxiesFromEnv()); err != nil {
		log.Printf("[trusted-proxies] DEVHUB_TRUSTED_PROXIES contains an invalid entry (%v); falling back to attribution-grade default (SetTrustedProxies(nil))", err)
		_ = router.SetTrustedProxies(nil)
	}

	if cfg.PermissionCache == nil {
		cfg.PermissionCache = rbacview.NewPermissionCache(cfg.RBACStore)
	}

	handler := Handler{
		cfg: cfg,
		auth: authview.NewAuthHandler(authview.AuthConfig{
			AuthDevFallback:       cfg.AuthDevFallback,
			RealtimeTickets:       cfg.RealtimeTickets,
			BearerTokenVerifier:   cfg.BearerTokenVerifier,
			OrganizationStore:     cfg.OrganizationStore,
			IdentityAdmin:         cfg.IdentityAdmin,
			AuditStore:            cfg.AuditStore,
			OnboardingGateEnabled: cfg.OnboardingGateEnabled,
		}),
		audit: auditview.NewAuditHandler(auditview.AuditConfig{
			AuditStore:            cfg.AuditStore,
			KeycloakWebhookSecret: cfg.KeycloakWebhookSecret,
		}),
		rbac: rbacview.NewRBACHandler(rbacview.RBACConfig{
			RBACStore:       cfg.RBACStore,
			PermissionCache: cfg.PermissionCache,
			AuthDevFallback: cfg.AuthDevFallback,
			AuditStore:      cfg.AuditStore,
		}),
		org: orgview.NewOrganizationHandler(orgview.OrganizationConfig{
			OrganizationStore:     cfg.OrganizationStore,
			HRDB:                  cfg.HRDB,
			AuditStore:            cfg.AuditStore,
			OnboardingGateEnabled: cfg.OnboardingGateEnabled,
		}),
		app: appview.NewPlatformHandler(appview.PlatformConfig{
			PlatformStore: cfg.PlatformStore,
			DevRequestStore:  cfg.DevRequestStore,
			ProjectModel:     cfg.ProjectModel,
			AuditStore:       cfg.AuditStore,
		}),
		devreq: devreqview.NewDevRequestHandler(devreqview.DevRequestConfig{
			DevRequestStore:            cfg.DevRequestStore,
			DevRequestIntakeTokenStore: cfg.DevRequestIntakeTokenStore,
			PlatformStore:           cfg.PlatformStore,
			AuditStore:                 cfg.AuditStore,
		}),
		integ: integview.NewIntegrationHandler(integview.IntegrationConfig{
			IntegrationStore:  resolveIntegrationStore(cfg),
			EventStore:        cfg.EventStore,
			EventProcessor:    cfg.EventProcessor,
			ExternalTaskStore: cfg.ExternalTaskStore,
			AuditStore:        cfg.AuditStore,
		}),
		realtime: realtimeview.NewRealtimeHandler(realtimeview.RealtimeConfig{
			RealtimeHub:     cfg.RealtimeHub,
			RealtimeTickets: cfg.RealtimeTickets,
			PermissionCache: cfg.PermissionCache,
			AuditStore:      cfg.AuditStore,
			AuthDevFallback: cfg.AuthDevFallback,
		}),
		repo: repoview.NewRepositoryIntegrationHandler(repoview.RepositoryIntegrationConfig{
			PlatformStore: resolveRepoIntegrationStore(cfg),
			AuditStore:       cfg.AuditStore,
		}),
		onboard: onboardview.NewOnboardingHandler(onboardview.OnboardingConfig{
			OrganizationStore:     cfg.OrganizationStore,
			OnboardingGateEnabled: cfg.OnboardingGateEnabled,
			AuditStore:            cfg.AuditStore,
		}),
		ciAdapter: cfg.CIAdapter,
	}
	router.GET("/health", handler.health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Keycloak Event Listener SPI Webhook (unauthenticated, X-Webhook-Secret check)
	router.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	v1 := router.Group("/api/v1")
	v1.Use(handler.requireRequestID)
	v1.Use(handler.authenticateActor)
	// RM-ONBOARD-01 (ADR-0021 §3.3) — onboardingGate middleware.
	// Feature flag default ON (PR #290, lazy_auto_create 폐기 후) — 미완료 사용자가
	// allowlist 외 endpoint 호출 시 403. DEVHUB_ONBOARDING_GATE_ENABLED=0 으로 rollback(no-op).
	v1.Use(handler.onboardingGate)
	v1.Use(handler.enforceRoutePermission)
	v1.POST("/auth/logout", handler.logout)
	v1.GET("/me", handler.getMe)
	// RM-ONBOARD-01 — API-83/84/85/86 onboarding endpoints.
	v1.PATCH("/me", handler.patchMe)
	v1.POST("/me/onboarding", handler.submitOnboarding)
	v1.GET("/organizations/search", handler.searchOrganizations)
	v1.POST("/admin/users/:user_id/review", handler.confirmUserReview)
	v1.GET("/dashboard/metrics", handler.dashboardMetrics)
	v1.GET("/events", handler.listWebhookEvents)
	v1.GET("/infra/edges", handler.infraEdges)
	v1.GET("/infra/nodes", handler.infraNodes)
	v1.GET("/infra/topology", handler.infraTopology)
	v1.GET("/infra/services", handler.listInfraServices)
	v1.POST("/infra/services/snapshot", handler.ingestInfraServicesSnapshot)
	v1.GET("/infra/topology/v2", handler.infraTopologyV2)
	v1.GET("/repositories", handler.repositories)
	v1.POST("/repositories", handler.createRepositoryDraft)
	v1.PATCH("/repositories/:repository_id", handler.updateRepositoryDraft)
	v1.DELETE("/repositories/:repository_id", handler.deleteRepository)
	v1.POST("/repositories/:repository_id/publish", handler.requestRepositoryPublish)
	v1.GET("/issues", handler.issues)
	v1.GET("/pull-requests", handler.pullRequests)
	v1.GET("/ci-runs", handler.ciRuns)
	v1.POST("/ci-runs", handler.createCIRun)
	v1.GET("/ci-runs/:ci_run_id/logs", handler.ciRunLogs)
	v1.GET("/risks", handler.risks)
	v1.GET("/risks/critical", handler.criticalRisks)
	v1.GET("/audit-logs", handler.listAuditLogs)
	v1.GET("/rbac/policy", handler.getRBACPolicyLegacyGone)
	v1.GET("/rbac/policies", handler.listRBACPolicies)
	v1.POST("/rbac/policies", handler.createRBACPolicy)
	v1.PUT("/rbac/policies", handler.updateRBACPolicies)
	v1.DELETE("/rbac/policies/:role_id", handler.deleteRBACPolicy)
	v1.POST("/admin/service-actions", handler.createServiceAction)
	v1.POST("/risks/:risk_id/mitigations", handler.createRiskMitigation)
	v1.GET("/commands/:command_id", handler.getCommand)
	v1.POST("/commands/:command_id/approve", handler.approveCommand)
	v1.POST("/commands/:command_id/reject", handler.rejectCommand)
	v1.GET("/users", handler.listUsers)
	v1.POST("/users", handler.createUser)
	v1.GET("/users/:user_id", handler.getUser)
	v1.PATCH("/users/:user_id", handler.updateUser)
	v1.DELETE("/users/:user_id", handler.deleteUser)
	// ADR-0020 sub-carve B (sprint -i, issue #209): /api/v1/accounts/* 4 endpoint
	// 폐기. user 생성/비밀번호/상태/삭제는 Keycloak admin console 또는 HRDB ETL
	// push 책임 (사내 IdP 팀). frontend admin actions cleanup 은 Gemini 별도 sprint.
	// lazy auto-create 가 첫 진입 시 DevHub `users` row 자동 생성 (authenticateActor,
	// ADR-0020 §5.2).
	v1.GET("/organization/hierarchy", handler.getHierarchy)
	v1.PUT("/organization/hierarchy", handler.updateHierarchy)
	v1.POST("/organization/units", handler.createOrgUnit)
	v1.GET("/organization/units/:unit_id", handler.getOrgUnit)
	v1.PATCH("/organization/units/:unit_id", handler.updateOrgUnit)
	v1.DELETE("/organization/units/:unit_id", handler.deleteOrgUnit)
	v1.GET("/organization/units/:unit_id/members", handler.listUnitMembers)
	v1.PUT("/organization/units/:unit_id/members", handler.replaceUnitMembers)
	// Application/Repository/Project 관리 API (API-41..58, sprint claude/work_260514-a~b).
	// Handler/store body 구현 완료 (activated) — draft/publish (API-91/92) + SCM 양방향 (API-88..90) 포함.
	v1.GET("/scm/providers", handler.listSCMProviders)
	v1.PATCH("/scm/providers/:provider_key", handler.updateSCMProvider)
	v1.GET("/platforms", handler.listPlatforms)
	v1.POST("/platforms", handler.createPlatform)
	v1.GET("/platforms/:platform_id", handler.getPlatform)
	v1.PATCH("/platforms/:platform_id", handler.updatePlatform)
	v1.DELETE("/platforms/:platform_id", handler.archivePlatform)
	v1.GET("/platforms/:platform_id/repositories", handler.listPlatformRepositories)
	v1.POST("/platforms/:platform_id/repositories", handler.createPlatformRepository)
	// :repo_key 가 'provider:org/repo' 컨벤션이라 path 에 `/` 포함. gin 의 catch-all
	// `*repo_key` 사용 — 핸들러는 leading `/` 를 strip 한 뒤 콜론으로 분리.
	v1.DELETE("/platforms/:platform_id/repositories/*repo_key", handler.deletePlatformRepository)
	// API-51..54 Repository 운영 지표 (sprint claude/work_260514-c)
	v1.GET("/repositories/:repository_id/activity", handler.repositoryActivity)
	v1.GET("/repositories/:repository_id/pull-requests", handler.repositoryPullRequests)
	v1.GET("/repositories/:repository_id/build-runs", handler.repositoryBuildRuns)
	v1.GET("/repositories/:repository_id/quality-snapshots", handler.repositoryQualitySnapshots)
	// API-55..56 Project CRUD (sprint claude/work_260514-c)
	v1.GET("/repositories/:repository_id/projects", handler.listProjects)
	v1.POST("/repositories/:repository_id/projects", handler.createProject)
	v1.POST("/projects", handler.createProjectStandalone)
	// /projects/standalone 은 /projects/:project_id 보다 먼저 정의해야 gin 이 ID 로 안 잡음.
	v1.GET("/projects/standalone", handler.listStandaloneProjects)
	v1.GET("/platforms/:platform_id/projects", handler.listPlatformProjects)
	v1.POST("/platforms/:platform_id/projects", handler.createPlatformProject)
	v1.GET("/projects/:project_id", handler.getProject)
	v1.PATCH("/projects/:project_id", handler.updateProject)
	v1.DELETE("/projects/:project_id", handler.archiveProject)
	v1.GET("/projects/:project_id/repositories", handler.listProjectRepositories)
	v1.POST("/projects/:project_id/repositories", handler.createProjectRepository)
	v1.DELETE("/projects/:project_id/repositories/:repository_id", handler.deleteProjectRepository)
	// API-57 Application 롤업 (sprint claude/work_260514-c)
	v1.GET("/platforms/:platform_id/rollup", handler.platformRollup)
	v1.GET("/platforms/:platform_id/dashboard", handler.platformDashboard)
	// API-58 Integration CRUD (sprint claude/work_260514-c)
	v1.GET("/integrations", handler.listIntegrations)
	v1.POST("/integrations", handler.createIntegration)
	v1.PATCH("/integrations/:integration_id", handler.updateIntegration)
	v1.DELETE("/integrations/:integration_id", handler.deleteIntegration)
	v1.POST("/integrations/gitea/webhooks", handler.receiveGiteaWebhook)
	v1.GET("/integration/providers", handler.listIntegrationProviders)
	v1.POST("/integration/providers", handler.createIntegrationProvider)
	v1.PATCH("/integration/providers/:provider_id", handler.updateIntegrationProvider)
	v1.DELETE("/integration/providers/:provider_id", handler.deleteIntegrationProvider)
	v1.POST("/integration/providers/:provider_id/sync", handler.syncIntegrationProvider)
	v1.GET("/integration/providers/:provider_id/scm-repositories", handler.listSCMRepositories)
	v1.POST("/integration/providers/:provider_id/import-repositories", handler.importSCMRepositories)
	v1.POST("/integration/providers/:provider_id/create-repository", handler.createSCMRepository)
	v1.POST("/integration/providers/:provider_id/webhook", handler.ingestIntegrationProviderWebhook)
	v1.POST("/integration/test-connection", handler.testIntegrationConnection)
	v1.GET("/integration/bindings", handler.listIntegrationBindings)
	v1.POST("/integration/bindings", handler.createIntegrationBinding)
	v1.PATCH("/integration/bindings/:binding_id", handler.updateIntegrationBinding)
	v1.DELETE("/integration/bindings/:binding_id", handler.deleteIntegrationBinding)

	// Task Item Ingestion (API-94..96) — sprint deepseek/work_260528-a-task-item-ingestion.
	v1.POST("/integration/providers/:provider_id/tasks/webhook", handler.receiveExternalTaskWebhook)
	v1.GET("/external-tasks", handler.listExternalTaskItems)
	v1.GET("/external-tasks/:task_id", handler.getExternalTaskItem)

	// DREQ 도메인 API-59..65 (sprint claude/work_260515-i, ADR-0012).
	// 외부 수신 POST 는 별도 intake group 에서 requireIntakeToken 미들웨어를 사용.
	// 나머지 6 endpoint 는 일반 OIDC + RBAC + enforceRowOwnership 로 보호.
	intakeGroup := router.Group("/api/v1")
	intakeGroup.Use(handler.requireIntakeToken)
	intakeGroup.POST("/dev-requests", handler.intakeDevRequest)
	v1.GET("/dev-requests", handler.listDevRequests)
	v1.GET("/dev-requests/:dev_request_id", handler.getDevRequest)
	v1.POST("/dev-requests/:dev_request_id/register", handler.registerDevRequest)
	v1.POST("/dev-requests/:dev_request_id/reject", handler.rejectDevRequest)
	v1.PATCH("/dev-requests/:dev_request_id", handler.patchDevRequest)
	v1.DELETE("/dev-requests/:dev_request_id", handler.closeDevRequest)

	// DREQ intake token admin (sprint claude/work_260515-o, ADR-0014).
	// system_admin 일임 — routePermissionTable 의 dev_request_intake_tokens resource gate.
	v1.POST("/dev-request-tokens", handler.createDevRequestIntakeToken)
	v1.GET("/dev-request-tokens", handler.listDevRequestIntakeTokens)
	v1.DELETE("/dev-request-tokens/:token_id", handler.revokeDevRequestIntakeToken)
	v1.PATCH("/dev-request-tokens/:token_id", handler.updateDevRequestIntakeTokenIPs)

	v1.GET("/hr/lookup", handler.hrLookup)
	v1.POST("/realtime/ticket", handler.issueRealtimeTicket)
	if cfg.RealtimeHub != nil {
		v1.GET("/realtime/ws", handler.handleRealtimeWebSocket)
	}

	return router
}

// trustedProxiesFromEnv returns the SetTrustedProxies argument derived from
// DEVHUB_TRUSTED_PROXIES. Empty / "none" → nil (raw RemoteAddr, the default
// PR-D contract); "*" → trust all (testing); otherwise comma-separated
// CIDRs/IPs. Whitespace around entries is trimmed; empty entries are dropped.
//
// Returning nil from "none" reads as "explicit opt-out from any forwarding
// header" and matches the silent default; the alias keeps the env contract
// expressive ("set to none" vs "leave unset" both work).
func trustedProxiesFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("DEVHUB_TRUSTED_PROXIES"))
	if raw == "" || strings.EqualFold(raw, "none") {
		return nil
	}
	if raw == "*" {
		return []string{"0.0.0.0/0", "::/0"}
	}
	parts := strings.Split(raw, ",")
	proxies := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			proxies = append(proxies, trimmed)
		}
	}
	if len(proxies) == 0 {
		return nil
	}
	return proxies
}

type Handler struct {
	cfg      RouterConfig
	auth     *authview.AuthHandler
	audit    *auditview.AuditHandler
	rbac     *rbacview.RBACHandler
	org      *orgview.OrganizationHandler
	app      *appview.PlatformHandler
	devreq   *devreqview.DevRequestHandler
	integ    *integview.IntegrationHandler
	realtime *realtimeview.RealtimeHandler
	repo     *repoview.RepositoryIntegrationHandler
	onboard  *onboardview.OnboardingHandler
	ciAdapter ci.Adapter
}

// resolveIdPSubject — test compatibility shim. Handler 직접 cfg 로 AuthHandler
// 인스턴스 만들어 view 패키지의 동명 로직을 위임. production 코드의 router.go
// 가 NewAuthHandler 로 만든 h.auth 와 동일 cfg fan-out (OrganizationStore +
// IdentityAdmin 만 사용) 이므로 test 호출 결과 동일.
func (h Handler) resolveIdPSubject(ctx context.Context, userID string) (string, error) {
	auth := authview.NewAuthHandler(authview.AuthConfig{
		OrganizationStore: h.cfg.OrganizationStore,
		IdentityAdmin:     h.cfg.IdentityAdmin,
	})
	return auth.ResolveIdPSubject(ctx, userID)
}

// enforceRowOwnership — test compatibility shim. RBACHandler 가 갖고 있는 동명
// helper 로 위임. test 가 Handler{cfg: RouterConfig{...}} 형태로 직접 호출.
func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
	rh := rbacview.NewRBACHandler(rbacview.RBACConfig{AuditStore: h.cfg.AuditStore})
	return rh.EnforceRowOwnership(c, ownerUserID, allowedRoles...)
}

func (h Handler) snapshotProvider() SnapshotProvider {
	if h.cfg.SnapshotProvider != nil {
		return h.cfg.SnapshotProvider
	}
	return StaticSnapshotProvider{}
}

func (h Handler) health(c *gin.Context) {
	dbStatus := "disabled"
	if h.cfg.HealthStore != nil {
		dbStatus = "ok"
		if err := h.cfg.HealthStore.Ping(c.Request.Context()); err != nil {
			dbStatus = "error"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "degraded",
				"service": "backend-core",
				"db":      dbStatus,
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "backend-core",
		"db":      dbStatus,
	})
}

func statusFromStoreError(err error) (int, string) {
	if errors.Is(err, store.ErrDuplicateEvent) {
		return http.StatusOK, "duplicate"
	}
	return http.StatusInternalServerError, "failed"
}

func (h Handler) scmProviderClient(c *gin.Context, provider domain.IntegrationProvider) (*gitea.Client, bool) {
	if strings.TrimSpace(provider.BaseURL) == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider base_url is not set", "code": "integration_base_url_missing"})
		return nil, false
	}
	client, err := gitea.NewClientForAuth(c.Request.Context(), provider.BaseURL, provider.ResolveOutboundAuth())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "SCM authentication failed: " + err.Error(), "code": "integration_scm_auth_failed"})
		return nil, false
	}
	if client == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider outbound credentials are not configured", "code": "integration_outbound_credentials_missing"})
		return nil, false
	}
	return client, true
}

var giteaCompatibleVendors = map[string]bool{"gitea": true, "forgejo": true, "gogs": true}

func isGiteaCompatibleProvider(p domain.IntegrationProvider) bool {
	const prefix = "provider_sdk:"
	ref := strings.TrimSpace(p.CredentialsRef)
	if !strings.HasPrefix(ref, prefix) {
		return true
	}
	parts := strings.SplitN(strings.TrimPrefix(ref, prefix), ":", 2)
	vendor := normalizeProviderSDKKey(parts[0])
	if vendor == "" {
		return true
	}
	return giteaCompatibleVendors[vendor]
}

func scmRepoOwnerLogin(fullName string) string {
	if i := strings.Index(fullName, "/"); i > 0 {
		return fullName[:i]
	}
	return ""
}

func normalizeProviderSDKKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func (h Handler) platformStoreOrUnavailable(c *gin.Context) (PlatformStore, bool) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "platform store is not configured"})
		return nil, false
	}
	return h.cfg.PlatformStore, true
}

func formatDatePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format("2006-01-02")
}

func formatTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func (h Handler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
	return h.audit.RecordAuditBestEffort(c, action, targetType, targetID, payload)
}

func (h Handler) requireOnboardingFlag(c *gin.Context) bool {
	return h.onboard.RequireOnboardingFlag(c)
}

func addAuditMeta(resp gin.H, log domain.AuditLog) {
	if log.AuditID != "" {
		resp["audit_log_id"] = log.AuditID
	}
}

// ==========================================
// Domain View Delegation Wrapper Methods
// ==========================================

// ensure — test 가 Handler{cfg: RouterConfig{...}} 우회 경로로 sub-handler 들이
// 모두 nil 인 상태에서 method 를 호출했을 때, cfg 로부터 즉석에서 sub-handler 들을
// build 해 nil deref 회피. production path 는 NewRouter 가 모두 init → 본 함수의
// nil 분기 미진입 (build cost = 0). value receiver 라 build 결과는 caller frame
// 한정.
func (h Handler) ensure() Handler {
	if h.auth == nil {
		h.auth = authview.NewAuthHandler(authview.AuthConfig{
			AuthDevFallback:       h.cfg.AuthDevFallback,
			RealtimeTickets:       h.cfg.RealtimeTickets,
			BearerTokenVerifier:   h.cfg.BearerTokenVerifier,
			OrganizationStore:     h.cfg.OrganizationStore,
			IdentityAdmin:         h.cfg.IdentityAdmin,
			AuditStore:            h.cfg.AuditStore,
			OnboardingGateEnabled: h.cfg.OnboardingGateEnabled,
		})
	}
	if h.audit == nil {
		h.audit = auditview.NewAuditHandler(auditview.AuditConfig{
			AuditStore:            h.cfg.AuditStore,
			KeycloakWebhookSecret: h.cfg.KeycloakWebhookSecret,
		})
	}
	if h.rbac == nil {
		pc := h.cfg.PermissionCache
		if pc == nil {
			pc = rbacview.NewPermissionCache(h.cfg.RBACStore)
		}
		h.rbac = rbacview.NewRBACHandler(rbacview.RBACConfig{
			RBACStore:       h.cfg.RBACStore,
			PermissionCache: pc,
			AuthDevFallback: h.cfg.AuthDevFallback,
			AuditStore:      h.cfg.AuditStore,
		})
	}
	if h.org == nil {
		h.org = orgview.NewOrganizationHandler(orgview.OrganizationConfig{
			OrganizationStore:     h.cfg.OrganizationStore,
			HRDB:                  h.cfg.HRDB,
			AuditStore:            h.cfg.AuditStore,
			OnboardingGateEnabled: h.cfg.OnboardingGateEnabled,
		})
	}
	if h.app == nil {
		h.app = appview.NewPlatformHandler(appview.PlatformConfig{
			PlatformStore: h.cfg.PlatformStore,
			DevRequestStore:  h.cfg.DevRequestStore,
			ProjectModel:     h.cfg.ProjectModel,
			AuditStore:       h.cfg.AuditStore,
		})
	}
	if h.devreq == nil {
		h.devreq = devreqview.NewDevRequestHandler(devreqview.DevRequestConfig{
			DevRequestStore:            h.cfg.DevRequestStore,
			DevRequestIntakeTokenStore: h.cfg.DevRequestIntakeTokenStore,
			PlatformStore:           h.cfg.PlatformStore,
			AuditStore:                 h.cfg.AuditStore,
		})
	}
	if h.integ == nil {
		h.integ = integview.NewIntegrationHandler(integview.IntegrationConfig{
			IntegrationStore:  resolveIntegrationStore(h.cfg),
			EventStore:        h.cfg.EventStore,
			EventProcessor:    h.cfg.EventProcessor,
			ExternalTaskStore: h.cfg.ExternalTaskStore,
			AuditStore:        h.cfg.AuditStore,
		})
	}
	if h.realtime == nil {
		h.realtime = realtimeview.NewRealtimeHandler(realtimeview.RealtimeConfig{
			RealtimeHub:     h.cfg.RealtimeHub,
			RealtimeTickets: h.cfg.RealtimeTickets,
			PermissionCache: h.cfg.PermissionCache,
			AuditStore:      h.cfg.AuditStore,
			AuthDevFallback: h.cfg.AuthDevFallback,
		})
	}
	if h.repo == nil {
		h.repo = repoview.NewRepositoryIntegrationHandler(repoview.RepositoryIntegrationConfig{
			PlatformStore: resolveRepoIntegrationStore(h.cfg),
			AuditStore:       h.cfg.AuditStore,
		})
	}
	if h.onboard == nil {
		h.onboard = onboardview.NewOnboardingHandler(onboardview.OnboardingConfig{
			OrganizationStore:     h.cfg.OrganizationStore,
			OnboardingGateEnabled: h.cfg.OnboardingGateEnabled,
			AuditStore:            h.cfg.AuditStore,
		})
	}
	return h
}

// Auth-Session
func (h Handler) authenticateActor(c *gin.Context) {
	h = h.ensure()
	h.auth.AuthenticateActor(c)
}
func (h Handler) getMe(c *gin.Context) {
	h = h.ensure()
	h.auth.GetMe(c)
}
func (h Handler) patchMe(c *gin.Context) {
	h = h.ensure()
	h.auth.PatchMe(c)
}
func (h Handler) logout(c *gin.Context) {
	h = h.ensure()
	h.auth.Logout(c)
}

// Onboarding
func (h Handler) onboardingGate(c *gin.Context) {
	h = h.ensure()
	h.onboard.OnboardingGate(c)
}
func (h Handler) submitOnboarding(c *gin.Context) {
	h = h.ensure()
	h.onboard.SubmitOnboarding(c)
}

// Organization Management
func (h Handler) searchOrganizations(c *gin.Context) {
	h = h.ensure()
	h.org.SearchOrganizations(c)
}

func (h Handler) listUsers(c *gin.Context) {
	h = h.ensure()
	h.org.ListUsers(c)
}
func (h Handler) createUser(c *gin.Context) {
	h = h.ensure()
	h.org.CreateUser(c)
}
func (h Handler) getUser(c *gin.Context) {
	h = h.ensure()
	h.org.GetUser(c)
}
func (h Handler) updateUser(c *gin.Context) {
	h = h.ensure()
	h.org.UpdateUser(c)
}
func (h Handler) deleteUser(c *gin.Context) {
	h = h.ensure()
	h.org.DeleteUser(c)
}
func (h Handler) getHierarchy(c *gin.Context) {
	h = h.ensure()
	h.org.GetHierarchy(c)
}
func (h Handler) updateHierarchy(c *gin.Context) {
	h = h.ensure()
	h.org.UpdateHierarchy(c)
}
func (h Handler) createOrgUnit(c *gin.Context) {
	h = h.ensure()
	h.org.CreateOrgUnit(c)
}
func (h Handler) getOrgUnit(c *gin.Context) {
	h = h.ensure()
	h.org.GetOrgUnit(c)
}
func (h Handler) updateOrgUnit(c *gin.Context) {
	h = h.ensure()
	h.org.UpdateOrgUnit(c)
}
func (h Handler) deleteOrgUnit(c *gin.Context) {
	h = h.ensure()
	h.org.DeleteOrgUnit(c)
}
func (h Handler) listUnitMembers(c *gin.Context) {
	h = h.ensure()
	h.org.ListUnitMembers(c)
}
func (h Handler) replaceUnitMembers(c *gin.Context) {
	h = h.ensure()
	h.org.ReplaceUnitMembers(c)
}
func (h Handler) hrLookup(c *gin.Context) {
	h = h.ensure()
	h.org.HrLookup(c)
}

// RBAC Permissions
func (h Handler) enforceRoutePermission(c *gin.Context) {
	h = h.ensure()
	h.rbac.EnforceRoutePermission(c)
}
func (h Handler) getRBACPolicyLegacyGone(c *gin.Context) {
	h = h.ensure()
	h.rbac.GetRBACPolicyLegacyGone(c)
}
func (h Handler) listRBACPolicies(c *gin.Context) {
	h = h.ensure()
	h.rbac.ListRBACPolicies(c)
}
func (h Handler) createRBACPolicy(c *gin.Context) {
	h = h.ensure()
	h.rbac.CreateRBACPolicy(c)
}
func (h Handler) updateRBACPolicies(c *gin.Context) {
	h = h.ensure()
	h.rbac.UpdateRBACPolicies(c)
}
func (h Handler) deleteRBACPolicy(c *gin.Context) {
	h = h.ensure()
	h.rbac.DeleteRBACPolicy(c)
}

// Audit Ops
func (h Handler) receiveKeycloakEventWebhook(c *gin.Context) {
	h = h.ensure()
	h.audit.ReceiveKeycloakEventWebhook(c)
}
func (h Handler) listAuditLogs(c *gin.Context) {
	h = h.ensure()
	h.audit.ListAuditLogs(c)
}

// Application Lifecycle
func (h Handler) listSCMProviders(c *gin.Context) {
	h = h.ensure()
	h.app.ListSCMProviders(c)
}
func (h Handler) updateSCMProvider(c *gin.Context) {
	h = h.ensure()
	h.app.UpdateSCMProvider(c)
}
func (h Handler) listPlatforms(c *gin.Context) {
	h = h.ensure()
	h.app.ListPlatforms(c)
}
func (h Handler) createPlatform(c *gin.Context) {
	h = h.ensure()
	h.app.CreatePlatform(c)
}
func (h Handler) getPlatform(c *gin.Context) {
	h = h.ensure()
	h.app.GetPlatform(c)
}
func (h Handler) updatePlatform(c *gin.Context) {
	h = h.ensure()
	h.app.UpdatePlatform(c)
}
func (h Handler) archivePlatform(c *gin.Context) {
	h = h.ensure()
	h.app.ArchivePlatform(c)
}
func (h Handler) listPlatformRepositories(c *gin.Context) {
	h = h.ensure()
	h.app.ListPlatformRepositories(c)
}
func (h Handler) createPlatformRepository(c *gin.Context) {
	h = h.ensure()
	h.app.CreatePlatformRepository(c)
}
func (h Handler) deletePlatformRepository(c *gin.Context) {
	h = h.ensure()
	h.app.DeletePlatformRepository(c)
}

func (h Handler) listProjects(c *gin.Context) {
	h = h.ensure()
	h.app.ListProjects(c)
}
func (h Handler) createProject(c *gin.Context) {
	h = h.ensure()
	h.app.CreateProject(c)
}
func (h Handler) createProjectStandalone(c *gin.Context) {
	h = h.ensure()
	h.app.CreateProjectStandalone(c)
}
func (h Handler) listStandaloneProjects(c *gin.Context) {
	h = h.ensure()
	h.app.ListStandaloneProjects(c)
}
func (h Handler) listPlatformProjects(c *gin.Context) {
	h = h.ensure()
	h.app.ListPlatformProjects(c)
}
func (h Handler) createPlatformProject(c *gin.Context) {
	h = h.ensure()
	h.app.CreatePlatformProject(c)
}
func (h Handler) getProject(c *gin.Context) {
	h = h.ensure()
	h.app.GetProject(c)
}
func (h Handler) updateProject(c *gin.Context) {
	h = h.ensure()
	h.app.UpdateProject(c)
}
func (h Handler) archiveProject(c *gin.Context) {
	h = h.ensure()
	h.app.ArchiveProject(c)
}
func (h Handler) listProjectRepositories(c *gin.Context) {
	h = h.ensure()
	h.app.ListProjectRepositories(c)
}
func (h Handler) createProjectRepository(c *gin.Context) {
	h = h.ensure()
	h.app.CreateProjectRepository(c)
}
func (h Handler) deleteProjectRepository(c *gin.Context) {
	h = h.ensure()
	h.app.DeleteProjectRepository(c)
}
func (h Handler) platformRollup(c *gin.Context) {
	h = h.ensure()
	h.app.PlatformRollup(c)
}
func (h Handler) platformDashboard(c *gin.Context) {
	h = h.ensure()
	h.app.PlatformDashboard(c)
}
func (h Handler) listIntegrations(c *gin.Context) {
	h = h.ensure()
	h.integ.ListIntegrations(c)
}
func (h Handler) createIntegration(c *gin.Context) {
	h = h.ensure()
	h.integ.CreateIntegration(c)
}
func (h Handler) updateIntegration(c *gin.Context) {
	h = h.ensure()
	h.integ.UpdateIntegration(c)
}
func (h Handler) deleteIntegration(c *gin.Context) {
	h = h.ensure()
	h.integ.DeleteIntegration(c)
}

// DevRequest
func (h Handler) requireIntakeToken(c *gin.Context) {
	h = h.ensure()
	h.devreq.RequireIntakeToken(c)
}
func (h Handler) intakeDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.IntakeDevRequest(c)
}
func (h Handler) listDevRequests(c *gin.Context) {
	h = h.ensure()
	h.devreq.ListDevRequests(c)
}
func (h Handler) getDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.GetDevRequest(c)
}
func (h Handler) registerDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.RegisterDevRequest(c)
}
func (h Handler) rejectDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.RejectDevRequest(c)
}
func (h Handler) patchDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.PatchDevRequest(c)
}
func (h Handler) closeDevRequest(c *gin.Context) {
	h = h.ensure()
	h.devreq.CloseDevRequest(c)
}
func (h Handler) createDevRequestIntakeToken(c *gin.Context) {
	h = h.ensure()
	h.devreq.CreateDevRequestIntakeToken(c)
}
func (h Handler) listDevRequestIntakeTokens(c *gin.Context) {
	h = h.ensure()
	h.devreq.ListDevRequestIntakeTokens(c)
}
func (h Handler) revokeDevRequestIntakeToken(c *gin.Context) {
	h = h.ensure()
	h.devreq.RevokeDevRequestIntakeToken(c)
}
func (h Handler) updateDevRequestIntakeTokenIPs(c *gin.Context) {
	h = h.ensure()
	h.devreq.UpdateDevRequestIntakeTokenIPs(c)
}

// Integration Providers & Tasks
func (h Handler) listIntegrationProviders(c *gin.Context) {
	h = h.ensure()
	h.integ.ListIntegrationProviders(c)
}
func (h Handler) createIntegrationProvider(c *gin.Context) {
	h = h.ensure()
	h.integ.CreateIntegrationProvider(c)
}
func (h Handler) updateIntegrationProvider(c *gin.Context) {
	h = h.ensure()
	h.integ.UpdateIntegrationProvider(c)
}
func (h Handler) deleteIntegrationProvider(c *gin.Context) {
	h = h.ensure()
	h.integ.DeleteIntegrationProvider(c)
}
func (h Handler) syncIntegrationProvider(c *gin.Context) {
	h = h.ensure()
	h.integ.SyncIntegrationProvider(c)
}
func (h Handler) receiveExternalTaskWebhook(c *gin.Context) {
	h = h.ensure()
	h.integ.ReceiveExternalTaskWebhook(c)
}
func (h Handler) listExternalTaskItems(c *gin.Context) {
	h = h.ensure()
	h.integ.ListExternalTaskItems(c)
}
func (h Handler) getExternalTaskItem(c *gin.Context) {
	h = h.ensure()
	h.integ.GetExternalTaskItem(c)
}

// SCM Repositories
func (h Handler) listSCMRepositories(c *gin.Context) {
	h = h.ensure()
	h.repo.ListSCMRepositories(c)
}
func (h Handler) importSCMRepositories(c *gin.Context) {
	h = h.ensure()
	h.repo.ImportSCMRepositories(c)
}
func (h Handler) createSCMRepository(c *gin.Context) {
	h = h.ensure()
	h.repo.CreateSCMRepository(c)
}
func (h Handler) ingestIntegrationProviderWebhook(c *gin.Context) {
	h = h.ensure()
	h.integ.IngestIntegrationProviderWebhook(c)
}
func (h Handler) testIntegrationConnection(c *gin.Context) {
	h = h.ensure()
	h.integ.TestIntegrationConnection(c)
}
func (h Handler) listIntegrationBindings(c *gin.Context) {
	h = h.ensure()
	h.integ.ListIntegrationBindings(c)
}
func (h Handler) createIntegrationBinding(c *gin.Context) {
	h = h.ensure()
	h.integ.CreateIntegrationBinding(c)
}
func (h Handler) updateIntegrationBinding(c *gin.Context) {
	h = h.ensure()
	h.integ.UpdateIntegrationBinding(c)
}
func (h Handler) deleteIntegrationBinding(c *gin.Context) {
	h = h.ensure()
	h.integ.DeleteIntegrationBinding(c)
}

// Realtime
func (h Handler) issueRealtimeTicket(c *gin.Context) {
	h = h.ensure()
	h.realtime.IssueRealtimeTicket(c)
}
func (h Handler) handleRealtimeWebSocket(c *gin.Context) {
	h = h.ensure()
	h.realtime.HandleRealtimeWebSocket(c)
}
