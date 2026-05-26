package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

// ApplicationStore is the persistence contract for the Application / Repository / Project
// 도메인 (API-41..50). Implemented by *store.PostgresStore. Sprint claude/work_260514-b
// (Application Design 2차) 가 stub → body 로 교체.
type ApplicationStore interface {
	// Applications
	ListApplications(context.Context, store.ApplicationListOptions) ([]domain.Application, int, error)
	GetApplication(context.Context, string) (domain.Application, error)
	GetApplicationByKey(context.Context, string) (domain.Application, error)
	CreateApplication(context.Context, domain.Application) (domain.Application, error)
	UpdateApplication(context.Context, domain.Application) (domain.Application, error)
	ArchiveApplication(context.Context, string, string) (domain.Application, error)
	CountActiveApplicationRepositories(context.Context, string) (int, error)

	// Application-Repository link
	ListApplicationRepositories(context.Context, string) ([]domain.ApplicationRepository, error)
	CreateApplicationRepository(context.Context, domain.ApplicationRepository) (domain.ApplicationRepository, error)
	DeleteApplicationRepository(context.Context, store.ApplicationRepositoryLinkKey) error
	UpdateApplicationRepositorySync(context.Context, store.ApplicationRepositoryLinkKey, domain.ApplicationRepositorySyncStatus, domain.SyncErrorCode) error

	// SCM Provider catalog
	ListSCMProviders(context.Context) ([]domain.SCMProvider, error)
	UpdateSCMProvider(context.Context, domain.SCMProvider) (domain.SCMProvider, error)

	// Projects
	ListProjects(context.Context, store.ProjectListOptions) ([]domain.Project, int, error)
	GetProject(context.Context, string) (domain.Project, error)
	CreateProject(context.Context, domain.Project) (domain.Project, error)
	UpdateProject(context.Context, domain.Project) (domain.Project, error)
	ArchiveProject(context.Context, string, string) (domain.Project, error)
	ListProjectRepositories(context.Context, string) ([]domain.ProjectRepository, error)
	CreateProjectRepository(context.Context, domain.ProjectRepository) (domain.ProjectRepository, error)
	DeleteProjectRepository(context.Context, string, int64) error
	CreateProjectWithRepositories(context.Context, domain.Project, []int64) (domain.Project, error)

	// Repository 운영 지표 (API-51..54, sprint claude/work_260514-c)
	ListRepositoryActivity(context.Context, int64, store.RepositoryActivityOptions) (domain.RepositoryActivity, error)
	ListRepositoryPullRequests(context.Context, int64, store.PRActivityListOptions) ([]domain.PRActivity, int, error)
	ListRepositoryBuildRuns(context.Context, int64, store.BuildRunListOptions) ([]domain.BuildRun, int, error)
	ListRepositoryQualitySnapshots(context.Context, int64, store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error)

	// Application 롤업 (API-57)
	ComputeApplicationRollup(context.Context, string, domain.ApplicationRollupOptions) (domain.ApplicationRollup, error)
	CountApplicationCriticalWarnings(context.Context, string) (int, error)

	// Integration CRUD (API-58)
	ListIntegrations(context.Context, store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error)
	GetIntegration(context.Context, string) (domain.ProjectIntegration, error)
	CreateIntegration(context.Context, domain.ProjectIntegration) (domain.ProjectIntegration, error)
	UpdateIntegration(context.Context, domain.ProjectIntegration) (domain.ProjectIntegration, error)
	DeleteIntegration(context.Context, string) error
	// Integration registry/binding (API-69..75)
	ListIntegrationProviders(context.Context, store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error)
	GetIntegrationProviderByID(context.Context, string) (domain.IntegrationProvider, error)
	GetIntegrationProviderByKey(context.Context, string) (domain.IntegrationProvider, error)
	CreateIntegrationProvider(context.Context, domain.IntegrationProvider) (domain.IntegrationProvider, error)
	UpdateIntegrationProvider(context.Context, domain.IntegrationProvider) (domain.IntegrationProvider, error)
	DeleteIntegrationProvider(context.Context, string) error
	CreateIntegrationSyncJob(context.Context, string, string) (string, error)
	ListIntegrationBindings(context.Context, store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error)
	GetIntegrationBindingByID(context.Context, string) (domain.IntegrationBinding, error)
	CreateIntegrationBinding(context.Context, domain.IntegrationBinding) (domain.IntegrationBinding, error)
	UpdateIntegrationBinding(context.Context, domain.IntegrationBinding) (domain.IntegrationBinding, error)
	DeleteIntegrationBinding(context.Context, string) error
}

// IdentityAdmin — ADR-0020 sub-carve E (sprint -n) — Keycloak admin = 별도
// 운영팀 (PoLP). write methods 제거 (CreateIdentity / UpdateIdentityPassword /
// SetIdentityState / DeleteIdentity). service account 는 view-users +
// view-events 만 요구.
type IdentityAdmin interface {
	FindIdentityByUserID(ctx context.Context, userID string) (string, error)
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
	ApplicationStore      ApplicationStore
	// DevRequestStore + DevRequestIntakeTokenStore — DREQ 도메인 (ADR-0012, sprint claude/work_260515-i).
	DevRequestStore            DevRequestStore
	DevRequestIntakeTokenStore IntakeTokenStore
	RBACStore                  RBACStore
	PermissionCache            *PermissionCache
	IdentityAdmin              IdentityAdmin
	IdPProvider                string
	HRDB                       HRDBClient
	SnapshotProvider           SnapshotProvider
	RealtimeHub                *RealtimeHub
	// RealtimeTickets — ADR-0024 §3.2 ticket pattern. nil 이면 ticket endpoint
	// 가 503 unavailable + WS auth 는 access_token query fallback 사용.
	RealtimeTickets *RealtimeTicketStore
	// AuthDevFallback toggles dev-only authentication fallbacks: empty Authorization passes through authenticateActor and requireMinRole. Actor identity always resolves to "system" without a verifier. Default false: production-safe.
	AuthDevFallback bool
	// OnboardingGateEnabled — RM-ONBOARD-01 (ADR-0021 §3.3, ARCH-ONBOARD-03).
	// Feature flag (env `DEVHUB_ONBOARDING_GATE_ENABLED`) default **true** since
	// the 2026-05-21 lazy 폐기 sprint (issue #284). authenticateActor 의
	// token-only actor 처리는 항상 활성 — 본 flag 는 onboardingGate middleware
	// 의 차단 동작에만 영향.
	// - true (default): onboardingGate 가 미완료 사용자의 allowlist 외 endpoint
	//   호출 시 403 onboarding_required 차단.
	// - false (rollback): onboardingGate no-op (모든 endpoint 통과). token-only
	//   actor 처리는 여전히 활성. 운영 사고 시 빠른 mitigation 경로.
	OnboardingGateEnabled bool
	// ProjectModel toggles project-management route mode: legacy|hybrid|v2.
	ProjectModel string
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	// SetTrustedProxies(nil) makes gin.Context.ClientIP() return the raw
	// RemoteAddr without honouring X-Forwarded-For / X-Real-IP. PR-D
	// (audit_logs.source_ip) relies on ClientIP being attribution-grade —
	// trusting client-supplied forwarding headers would let any external
	// caller forge the audit row's IP. Operators that legitimately sit
	// behind a reverse proxy opt in via DEVHUB_TRUSTED_PROXIES (PR-D
	// follow-up, work_260512-i):
	//   - empty / "none"  → SetTrustedProxies(nil) (default, attribution-grade)
	//   - "*"             → trust everything (testing only, audit forgery risk)
	//   - "10.0.0.0/8,192.168.1.5" → trust the listed CIDRs/IPs
	//
	// gin's parseTrustedProxies stops at the first invalid entry and returns
	// a partial trust set + the parse error (work_260512-j). Silent partial
	// trust silently diverges from operator intent (e.g. a typo'd CIDR drops
	// every later entry), so we fall back to attribution-grade default + log
	// when the env contains an invalid token. Listed entries earlier than the
	// invalid one would already be partially applied; resetting back to nil
	// ensures a uniform behaviour rather than an unpredictable mix.
	if err := router.SetTrustedProxies(trustedProxiesFromEnv()); err != nil {
		log.Printf("[trusted-proxies] DEVHUB_TRUSTED_PROXIES contains an invalid entry (%v); falling back to attribution-grade default (SetTrustedProxies(nil))", err)
		_ = router.SetTrustedProxies(nil)
	}

	if cfg.PermissionCache == nil {
		cfg.PermissionCache = NewPermissionCache(cfg.RBACStore)
	}

	handler := Handler{cfg: cfg}
	router.GET("/health", handler.health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Keycloak Event Listener SPI Webhook (unauthenticated, X-Webhook-Secret check)
	router.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	v1 := router.Group("/api/v1")
	v1.Use(handler.requireRequestID)
	v1.Use(handler.authenticateActor)
	// RM-ONBOARD-01 (ADR-0021 §3.3) — onboardingGate middleware.
	// Feature flag default OFF (no-op) — Carve A 단독 머지 후 main 안정성.
	// Flag ON 시 미완료 사용자의 allowlist 외 endpoint 호출 시 403.
	v1.Use(handler.onboardingGate)
	v1.Use(handler.enforceRoutePermission)
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
	v1.GET("/issues", handler.issues)
	v1.GET("/pull-requests", handler.pullRequests)
	v1.GET("/ci-runs", handler.ciRuns)
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
	// Application/Repository/Project 관리 API (API-41..50, sprint claude/work_260514-a).
	// Handler bodies are 501 stubs; store body 는 후속 sprint carve out.
	v1.GET("/scm/providers", handler.listSCMProviders)
	v1.PATCH("/scm/providers/:provider_key", handler.updateSCMProvider)
	v1.GET("/applications", handler.listApplications)
	v1.POST("/applications", handler.createApplication)
	v1.GET("/applications/:application_id", handler.getApplication)
	v1.PATCH("/applications/:application_id", handler.updateApplication)
	v1.DELETE("/applications/:application_id", handler.archiveApplication)
	v1.GET("/applications/:application_id/repositories", handler.listApplicationRepositories)
	v1.POST("/applications/:application_id/repositories", handler.createApplicationRepository)
	// :repo_key 가 'provider:org/repo' 컨벤션이라 path 에 `/` 포함. gin 의 catch-all
	// `*repo_key` 사용 — 핸들러는 leading `/` 를 strip 한 뒤 콜론으로 분리.
	v1.DELETE("/applications/:application_id/repositories/*repo_key", handler.deleteApplicationRepository)
	// API-51..54 Repository 운영 지표 (sprint claude/work_260514-c)
	v1.GET("/repositories/:repository_id/activity", handler.repositoryActivity)
	v1.GET("/repositories/:repository_id/pull-requests", handler.repositoryPullRequests)
	v1.GET("/repositories/:repository_id/build-runs", handler.repositoryBuildRuns)
	v1.GET("/repositories/:repository_id/quality-snapshots", handler.repositoryQualitySnapshots)
	// API-55..56 Project CRUD (sprint claude/work_260514-c)
	v1.GET("/repositories/:repository_id/projects", handler.listProjects)
	v1.POST("/repositories/:repository_id/projects", handler.createProject)
	v1.GET("/applications/:application_id/projects", handler.listApplicationProjects)
	v1.POST("/applications/:application_id/projects", handler.createApplicationProject)
	v1.GET("/projects/:project_id", handler.getProject)
	v1.PATCH("/projects/:project_id", handler.updateProject)
	v1.DELETE("/projects/:project_id", handler.archiveProject)
	v1.GET("/projects/:project_id/repositories", handler.listProjectRepositories)
	v1.POST("/projects/:project_id/repositories", handler.createProjectRepository)
	v1.DELETE("/projects/:project_id/repositories/:repository_id", handler.deleteProjectRepository)
	// API-57 Application 롤업 (sprint claude/work_260514-c)
	v1.GET("/applications/:application_id/rollup", handler.applicationRollup)
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
	v1.POST("/integration/providers/:provider_id/webhook", handler.ingestIntegrationProviderWebhook)
	v1.GET("/integration/bindings", handler.listIntegrationBindings)
	v1.POST("/integration/bindings", handler.createIntegrationBinding)
	v1.PATCH("/integration/bindings/:binding_id", handler.updateIntegrationBinding)
	v1.DELETE("/integration/bindings/:binding_id", handler.deleteIntegrationBinding)

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
	cfg RouterConfig
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
