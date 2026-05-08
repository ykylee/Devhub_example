package httpapi

import (
	"context"
	"errors"
	"net/http"

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
}

type AuditStore interface {
	CreateAuditLog(context.Context, domain.AuditLog) (domain.AuditLog, error)
	ListAuditLogs(context.Context, store.ListAuditLogsOptions) ([]domain.AuditLog, error)
}

type RouterConfig struct {
	WebhookSecret       string
	EventStore          WebhookEventStore
	EventProcessor      WebhookEventProcessor
	HealthStore         HealthStore
	DomainStore         DomainStore
	CommandStore        CommandStore
	AuditStore          AuditStore
	BearerTokenVerifier BearerTokenVerifier
	OrganizationStore   OrganizationStore
	RBACStore           RBACStore
	SnapshotProvider    SnapshotProvider
	RealtimeHub         *RealtimeHub
	// AuthDevFallback toggles dev-only authentication fallbacks: empty Authorization passes through authenticateActor and requireMinRole. Actor identity always resolves to "system" without a verifier. Default false: production-safe.
	AuthDevFallback bool
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	handler := Handler{cfg: cfg}
	router.GET("/health", handler.health)

	v1 := router.Group("/api/v1")
	v1.Use(handler.authenticateActor)
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
	v1.GET("/audit-logs", handler.requireMinRole(domain.AppRoleManager), handler.listAuditLogs)
	v1.GET("/rbac/policy", handler.getRBACPolicyLegacyGone)
	v1.GET("/rbac/policies", handler.listRBACPolicies)
	v1.POST("/rbac/policies", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.createRBACPolicy)
	v1.PUT("/rbac/policies", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.updateRBACPolicies)
	v1.DELETE("/rbac/policies/:role_id", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.deleteRBACPolicy)
	v1.GET("/rbac/subjects/:subject_id/roles", handler.getSubjectRoles)
	v1.PUT("/rbac/subjects/:subject_id/roles", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.setSubjectRoles)
	v1.POST("/admin/service-actions", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.createServiceAction)
	v1.POST("/risks/:risk_id/mitigations", handler.requireMinRole(domain.AppRoleManager), handler.createRiskMitigation)
	v1.GET("/commands/:command_id", handler.getCommand)
	v1.GET("/users", handler.listUsers)
	v1.POST("/users", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.createUser)
	v1.GET("/users/:user_id", handler.getUser)
	v1.PATCH("/users/:user_id", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.updateUser)
	v1.DELETE("/users/:user_id", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.deleteUser)
	v1.GET("/organization/hierarchy", handler.getHierarchy)
	v1.POST("/organization/units", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.createOrgUnit)
	v1.GET("/organization/units/:unit_id", handler.getOrgUnit)
	v1.PATCH("/organization/units/:unit_id", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.updateOrgUnit)
	v1.DELETE("/organization/units/:unit_id", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.deleteOrgUnit)
	v1.GET("/organization/units/:unit_id/members", handler.listUnitMembers)
	v1.PUT("/organization/units/:unit_id/members", handler.requireMinRole(domain.AppRoleSystemAdmin), handler.replaceUnitMembers)
	v1.POST("/integrations/gitea/webhooks", handler.receiveGiteaWebhook)
	if cfg.RealtimeHub != nil {
		v1.GET("/realtime/ws", cfg.RealtimeHub.HandleWebSocket)
	}

	return router
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
