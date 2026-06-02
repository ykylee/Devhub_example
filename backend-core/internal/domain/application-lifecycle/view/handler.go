package view

import (
	"context"
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}

type DevRequestStore interface {
	ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
}

// ApplicationStore — application-lifecycle 도메인 persistence 컨트랙트.
// issue #422 (sprint claude/work_260529-n) — 기존 interface 가 integration
// CRUD 13 메서드를 포함해 cross-domain bloat 상태였음. 본 sprint 에서 integration
// 메서드를 integration-registry/view 의 `IntegrationStore` 로 이관 후 본 interface
// 는 Application / Application-Repository / SCM Provider / Project / Repository
// 운영 지표 / Application 롤업 + SCM repository import 에만 한정.
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

func (h *ApplicationHandler) actorIdentity(c *gin.Context) (string, string) {
	loginVal, _ := c.Get("devhub_actor_login")
	roleVal, _ := c.Get("devhub_actor_role")
	login, _ := loginVal.(string)
	role, _ := roleVal.(string)
	return login, role
}

func (h *ApplicationHandler) actorCanReadProject(c *gin.Context, storeI ApplicationStore, project domain.Project) (bool, string, error) {
	if httphelp.DevFallbackEnabled(c) {
		return true, "", nil
	}

	login, role := h.actorIdentity(c)
	if role == string(domain.AppRoleSystemAdmin) || role == string(domain.AppRoleTeamManager) {
		return true, "", nil
	}
	if login == "" {
		return false, "not_project_member", nil
	}
	if project.OwnerUserID == login {
		return true, "", nil
	}

	if len(project.ProjectMembers) == 0 && project.ID != "" {
		loaded, err := storeI.GetProject(c.Request.Context(), project.ID)
		if err != nil {
			return false, "", err
		}
		project = loaded
	}
	for _, member := range project.ProjectMembers {
		if member.UserID == login {
			return true, "", nil
		}
	}
	return false, "not_project_member", nil
}

func (h *ApplicationHandler) actorCanReadApplication(c *gin.Context, storeI ApplicationStore, app domain.Application) (bool, string, error) {
	if httphelp.DevFallbackEnabled(c) {
		return true, "", nil
	}

	login, role := h.actorIdentity(c)
	if role == string(domain.AppRoleSystemAdmin) || role == string(domain.AppRoleTeamManager) {
		return true, "", nil
	}
	if login == "" {
		return false, "not_application_member", nil
	}
	if app.OwnerUserID == login || app.LeaderUserID == login {
		return true, "", nil
	}

	projects, _, err := storeI.ListProjects(c.Request.Context(), store.ProjectListOptions{
		ApplicationID:   app.ID,
		IncludeArchived: true,
		Limit:           5000,
	})
	if err != nil {
		return false, "", err
	}
	for _, project := range projects {
		allowed, _, err := h.actorCanReadProject(c, storeI, project)
		if err != nil {
			return false, "", err
		}
		if allowed {
			return true, "", nil
		}
	}
	return false, "not_application_member", nil
}

func (h *ApplicationHandler) denyRowRead(c *gin.Context, deniedReason string) {
	_, actorRole := h.actorIdentity(c)
	h.recordAuditBestEffort(c, "auth.row_denied", "route", c.FullPath(), map[string]any{
		"actor_role":    actorRole,
		"denied_reason": deniedReason,
	})
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status":        "forbidden",
		"error":         "row read scope denied",
		"code":          "auth_row_denied",
		"denied_reason": deniedReason,
	})
}
