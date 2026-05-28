package view

import (
	"context"
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)

type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}

type DevRequestStore interface {
	ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
}

type ApplicationStore interface {
	// Applications
	ListApplications(context.Context, store.ApplicationListOptions) ([]domain.Application, int, error)
	GetApplication(context.Context, string) (domain.Application, error)
	GetApplicationByKey(context.Context, string) (domain.Application, error)
	CreateApplication(context.Context, domain.Application) (domain.Application, error)
	UpdateApplication(context.Context, domain.Application) (domain.Application, error)
	ArchiveApplication(context.Context, string, string) (domain.Application, error)
	DeleteApplication(context.Context, string) error
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
	DeleteProject(context.Context, string) error
	ListProjectRepositories(context.Context, string) ([]domain.ProjectRepository, error)
	CreateProjectRepository(context.Context, domain.ProjectRepository) (domain.ProjectRepository, error)
	DeleteProjectRepository(context.Context, string, int64) error
	CreateProjectWithRepositoryPayload(context.Context, domain.Project, []int64, *store.RepositoryCreatePayload) (domain.Project, error)

	// SCM repository import
	UpsertRepository(context.Context, domain.Repository) error
	ListRepositoriesByProvider(context.Context, string) ([]domain.Repository, error)

	// Repository 운영 지표
	ListRepositoryActivity(context.Context, int64, store.RepositoryActivityOptions) (domain.RepositoryActivity, error)
	ListRepositoryPullRequests(context.Context, int64, store.PRActivityListOptions) ([]domain.PRActivity, int, error)
	ListRepositoryBuildRuns(context.Context, int64, store.BuildRunListOptions) ([]domain.BuildRun, int, error)
	ListRepositoryQualitySnapshots(context.Context, int64, store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error)

	// Application 롤업
	ComputeApplicationRollup(context.Context, string, domain.ApplicationRollupOptions) (domain.ApplicationRollup, error)
	CountApplicationCriticalWarnings(context.Context, string) (int, error)

	// Integration CRUD
	ListIntegrations(context.Context, store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error)
	GetIntegration(context.Context, string) (domain.ProjectIntegration, error)
	CreateIntegration(context.Context, domain.ProjectIntegration) (domain.ProjectIntegration, error)
	UpdateIntegration(context.Context, domain.ProjectIntegration) (domain.ProjectIntegration, error)
	DeleteIntegration(context.Context, string) error
	// Integration registry/binding
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

type ApplicationConfig struct {
	ApplicationStore ApplicationStore
	DevRequestStore  DevRequestStore
	ProjectModel     string
	AuditStore       AuditStore
}

type ApplicationHandler struct {
	cfg ApplicationConfig
}

func NewApplicationHandler(cfg ApplicationConfig) *ApplicationHandler {
	return &ApplicationHandler{cfg: cfg}
}

func (h *ApplicationHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
	if h.cfg.AuditStore == nil {
		return domain.AuditLog{}
	}
	actor := httphelp.RequestActor(c)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["actor_source"] = actor.Source
	logRow, err := h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), domain.AuditLog{
		ActorLogin: actor.Login,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Payload:    payload,
		SourceIP:   httphelp.ClientIPFrom(c),
		RequestID:  httphelp.RequestIDFrom(c),
		SourceType: httphelp.SourceTypeFrom(c),
	})
	if err != nil {
		httphelp.LogRequest(c, "audit log persistence failed: action=%s target=%s/%s err=%v", action, targetType, targetID, err)
	}
	return logRow
}

func (h *ApplicationHandler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
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

	h.recordAuditBestEffort(c, "auth.row_denied", "route", c.FullPath(), map[string]any{
		"actor_role":    actorRole,
		"owner_user_id": ownerUserID,
		"denied_reason": "owner_mismatch",
	})
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status": "forbidden",
		"error":  "owner_mismatch — row write requires owner or elevated role",
		"code":   "auth_row_denied",
	})
	return false
}
