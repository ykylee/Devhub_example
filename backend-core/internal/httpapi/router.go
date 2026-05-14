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

// ApplicationStore defines the persistence interface for the application domain.
type ApplicationStore interface {
	ListApplications(context.Context, store.ApplicationListOptions) ([]domain.Application, int, error)
	GetApplication(context.Context, string) (domain.Application, error)
	CreateApplication(context.Context, domain.Application) (domain.Application, error)
	UpdateApplication(context.Context, domain.Application) (domain.Application, error)
	ArchiveApplication(context.Context, string, string) (domain.Application, error)
	ListApplicationRepositories(context.Context, string) ([]domain.ApplicationRepository, error)
	CreateApplicationRepository(context.Context, domain.ApplicationRepository) (domain.ApplicationRepository, error)
	DeleteApplicationRepository(context.Context, store.ApplicationRepositoryLinkKey) error
	ListSCMProviders(context.Context) ([]domain.SCMProvider, error)
	UpdateSCMProvider(context.Context, domain.SCMProvider) (domain.SCMProvider, error)
	ListProjects(context.Context, store.ProjectListOptions) ([]domain.Project, int, error)
	GetProject(context.Context, string) (domain.Project, error)
	CreateProject(context.Context, domain.Project) (domain.Project, error)
	UpdateProject(context.Context, domain.Project) (domain.Project, error)
	ArchiveProject(context.Context, string, string) (domain.Project, error)
}

type KratosAdmin interface {
	CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error)
	FindIdentityByUserID(ctx context.Context, userID string) (string, error)
	UpdateIdentityPassword(ctx context.Context, identityID, password string) error
	SetIdentityState(ctx context.Context, identityID string, active bool) error
	DeleteIdentity(ctx context.Context, identityID string) error
}

type HRDBClient interface {
	Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error) // simplified for now: returns email, userID, dept
}

type RouterConfig struct {
	WebhookSecret       string
	KratosWebhookToken  string
	EventStore          WebhookEventStore
	EventProcessor      WebhookEventProcessor
	HealthStore         HealthStore
	DomainStore         DomainStore
	CommandStore        CommandStore
	AuditStore          AuditStore
	ApplicationStore    ApplicationStore
	BearerTokenVerifier BearerTokenVerifier
	OrganizationStore   OrganizationStore
	RBACStore           RBACStore
	PermissionCache     *PermissionCache
	KratosLogin         KratosLoginClient
	HydraAdmin          HydraLoginAdmin
	HydraLogout         HydraLogoutAdmin
	HydraToken          HydraTokenExchanger
	HydraRevoker        HydraTokenRevoker
	KratosAdmin         KratosAdmin
	HRDB                HRDBClient
	SnapshotProvider    SnapshotProvider
	RealtimeHub         *RealtimeHub
	KratosSessionCache *KratosSessionCache
	AuthDevFallback bool
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	if err := router.SetTrustedProxies(trustedProxiesFromEnv()); err != nil {
		log.Printf("[trusted-proxies] DEVHUB_TRUSTED_PROXIES contains an invalid entry (%v); falling back to attribution-grade default (SetTrustedProxies(nil))", err)
		_ = router.SetTrustedProxies(nil)
	}

	if cfg.PermissionCache == nil {
		cfg.PermissionCache = NewPermissionCache(cfg.RBACStore)
	}
	if cfg.KratosSessionCache == nil {
		cfg.KratosSessionCache = NewKratosSessionCache()
	}

	handler := Handler{cfg: cfg}
	router.GET("/health", handler.health)

	v1 := router.Group("/api/v1")
	v1.Use(handler.requireRequestID)
	v1.Use(handler.authenticateActor)
	v1.Use(handler.enforceRoutePermission)
	v1.GET("/me", handler.getMe)
	v1.GET("/dashboard/metrics", handler.dashboardMetrics)
	v1.GET("/events", handler.listWebhookEvents)
	v1.GET("/infra/edges", handler.infraEdges)
	v1.GET("/infra/nodes", handler.infraNodes)
	v1.GET("/infra/topology", handler.infraTopology)
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
	v1.GET("/rbac/subjects/:subject_id/roles", handler.getSubjectRoles)
	v1.PUT("/rbac/subjects/:subject_id/roles", handler.setSubjectRoles)
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
	v1.POST("/accounts", handler.createAccount)
	v1.PUT("/accounts/:user_id/password", handler.resetAccountPassword)
	v1.PATCH("/accounts/:user_id", handler.updateAccountStatus)
	v1.DELETE("/accounts/:user_id", handler.deleteAccount)
	v1.POST("/account/password", handler.updateMyPassword)
	v1.GET("/organization/hierarchy", handler.getHierarchy)
	v1.PUT("/organization/hierarchy", handler.updateHierarchy)
	v1.POST("/organization/units", handler.createOrgUnit)
	v1.GET("/organization/units/:unit_id", handler.getOrgUnit)
	v1.PATCH("/organization/units/:unit_id", handler.updateOrgUnit)
	v1.DELETE("/organization/units/:unit_id", handler.deleteOrgUnit)
	v1.GET("/organization/units/:unit_id/members", handler.listUnitMembers)
	v1.PUT("/organization/units/:unit_id/members", handler.replaceUnitMembers)
	v1.GET("/scm/providers", handler.listSCMProviders)
	v1.PATCH("/scm/providers/:provider_key", handler.updateSCMProvider)
	v1.GET("/applications", handler.listApplications)
	v1.POST("/applications", handler.createApplication)
	v1.GET("/applications/:application_id", handler.getApplication)
	v1.PATCH("/applications/:application_id", handler.updateApplication)
	v1.DELETE("/applications/:application_id", handler.archiveApplication)
	v1.GET("/applications/:application_id/repositories", handler.listApplicationRepositories)
	v1.POST("/applications/:application_id/repositories", handler.createApplicationRepository)
	v1.DELETE("/applications/:application_id/repositories/:repo_key", handler.deleteApplicationRepository)

	// Projects (API-55..59)
	v1.GET("/projects", handler.listProjects)
	v1.POST("/repositories/:repository_id/projects", handler.createProject)
	v1.GET("/projects/:project_id", handler.getProject)
	v1.PATCH("/projects/:project_id", handler.updateProject)
	v1.DELETE("/projects/:project_id", handler.archiveProject)

	v1.POST("/integrations/gitea/webhooks", handler.receiveGiteaWebhook)
	v1.POST("/integrations/kratos/hook/settings/password/after", handler.kratosWebhookSettingsPasswordAfter)
	v1.GET("/hr/lookup", handler.hrLookup)
	v1.POST("/auth/signup", handler.authSignUp)
	v1.POST("/auth/login", handler.authLogin)
	v1.POST("/auth/logout", handler.authLogout)
	v1.POST("/auth/token", handler.authToken)
	v1.GET("/auth/consent", handler.authConsent)
	if cfg.RealtimeHub != nil {
		v1.GET("/realtime/ws", handler.handleRealtimeWebSocket)
	}

	return router
}

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
