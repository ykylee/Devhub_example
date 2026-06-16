package view

import (
	"context"
	"net/http"
	"time"

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

// PlatformStore — application-lifecycle 도메인 persistence 컨트랙트.
// issue #422 (sprint claude/work_260529-n) — 기존 interface 가 integration
// CRUD 13 메서드를 포함해 cross-domain bloat 상태였음. 본 sprint 에서 integration
// 메서드를 integration-registry/view 의 `IntegrationStore` 로 이관 후 본 interface
// 는 Application / Application-Repository / SCM Provider / Project / Repository
// 운영 지표 / Application 롤업 + SCM repository import 에만 한정.
type PlatformStore interface {
	// Applications
	ListPlatforms(context.Context, store.PlatformListOptions) ([]domain.Platform, int, error)
	GetPlatform(context.Context, string) (domain.Platform, error)
	GetPlatformByKey(context.Context, string) (domain.Platform, error)
	CreatePlatform(context.Context, domain.Platform) (domain.Platform, error)
	UpdatePlatform(context.Context, domain.Platform) (domain.Platform, error)
	ArchivePlatform(context.Context, string, string) (domain.Platform, error)
	DeletePlatform(context.Context, string) error
	CountActivePlatformRepositories(context.Context, string) (int, error)
	// N-13 (ADR-0028 §6 a) — platforms.inbound_source 자동 routing 의 sub-resource API
	UpdatePlatformInboundSource(context.Context, string, string, string) (domain.Platform, error)
	ListEnabledInboundSourcePlatforms(context.Context) ([]domain.Platform, error)

	// Application-Repository link
	ListPlatformRepositories(context.Context, string) ([]domain.PlatformRepository, error)
	CreatePlatformRepository(context.Context, domain.PlatformRepository) (domain.PlatformRepository, error)
	DeletePlatformRepository(context.Context, store.PlatformRepositoryLinkKey) error
	UpdatePlatformRepositorySync(context.Context, store.PlatformRepositoryLinkKey, domain.PlatformRepositorySyncStatus, domain.SyncErrorCode) error

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
	UpdateProjectRepositoryWeight(context.Context, string, int64, float64) error
	CreateProjectWithRepositoryPayload(context.Context, domain.Project, []int64, *store.RepositoryCreatePayload) (domain.Project, error)

	// SCM repository import
	UpsertRepository(context.Context, domain.Repository) error
	ListRepositoriesByProvider(context.Context, string) ([]domain.Repository, error)
	GetRepositoryByID(context.Context, int64) (domain.Repository, error)

	// Repository 운영 지표
	ListRepositoryActivity(context.Context, int64, store.RepositoryActivityOptions) (domain.RepositoryActivity, error)
	ListRepositoryPullRequests(context.Context, int64, store.PRActivityListOptions) ([]domain.PRActivity, int, error)
	ListRepositoryBuildRuns(context.Context, int64, store.BuildRunListOptions) ([]domain.BuildRun, int, error)
	ListRepositoryQualitySnapshots(context.Context, int64, store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error)
	// CountOpenAndMergedPRs — Sprint A (kpi-tests-per-domain-scope.md §6.1) 의
	// repository KPI 종합 정공법. pr_activities 의 (event_type='opened'|'merged') 별
	// distinct number count. state="closed"+merged_at IS NOT NULL 도 "merged" 로 합산.
	CountOpenAndMergedPRs(context.Context, int64, time.Time, time.Time) (int, int, error)

	// Project 가중치 rollup — Sprint B (kpi-tests-per-domain-scope.md §6.2) 의
	// project KPI 종합 정공법. project 의 N개 linked repository 의 raw metric 을
	// contribution_weight 로 가중평균.
	ComputeProjectWeightedKPI(context.Context, string, store.RepositoryActivityOptions) (domain.ProjectWeightedKPI, error)
	CountProjectOpenAndMergedPRs(context.Context, string, time.Time, time.Time) (int, int, error)

	// Project 가중치 적용 test results — Sprint B-Tests
	// (kpi-tests-per-domain-scope.md §6.2 follow-up). N개 linked repository 의
	// build_runs status 종합 + 가중치 pass rate + multi-repo recent. handler
	// 측에서 JSON 직렬화 시 window/limit 정규화.
	ListProjectTestResults(context.Context, string, store.BuildRunListOptions) (domain.ProjectWeightedTestResults, int, error)

	// Application 롤업
	ComputePlatformRollup(context.Context, string, domain.PlatformRollupOptions) (domain.PlatformRollup, error)
	CountPlatformCriticalWarnings(context.Context, string) (int, error)
}

type PlatformConfig struct {
	PlatformStore PlatformStore
	DevRequestStore  DevRequestStore
	ProjectModel     string
	AuditStore       AuditStore
}

type PlatformHandler struct {
	cfg PlatformConfig
}

func NewPlatformHandler(cfg PlatformConfig) *PlatformHandler {
	return &PlatformHandler{cfg: cfg}
}

func (h *PlatformHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *PlatformHandler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
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

func (h *PlatformHandler) actorIdentity(c *gin.Context) (string, string) {
	loginVal, _ := c.Get("devhub_actor_login")
	roleVal, _ := c.Get("devhub_actor_role")
	login, _ := loginVal.(string)
	role, _ := roleVal.(string)
	return login, role
}

func (h *PlatformHandler) actorCanReadProject(c *gin.Context, storeI PlatformStore, project domain.Project) (bool, string, error) {
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

func (h *PlatformHandler) actorCanReadPlatform(c *gin.Context, storeI PlatformStore, app domain.Platform) (bool, string, error) {
	login, role := h.actorIdentity(c)
	if role == string(domain.AppRoleSystemAdmin) || role == string(domain.AppRoleTeamManager) {
		return true, "", nil
	}
	if login == "" {
		return false, "not_platform_member", nil
	}
	if app.OwnerUserID == login || app.LeaderUserID == login {
		return true, "", nil
	}

	projOpts := store.ProjectListOptions{
		PlatformID:   app.ID,
		IncludeArchived: true,
		Limit:           5000,
	}
	if loginVal, ok := c.Get("devhub_actor_login"); ok {
		if login, ok := loginVal.(string); ok {
			projOpts.ActorLogin = login
		}
	}
	if roleVal, ok := c.Get("devhub_actor_role"); ok {
		if role, ok := roleVal.(string); ok {
			projOpts.ActorRole = role
		}
	}
	if idsVal, ok := c.Get("devhub_actor_org_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			projOpts.OrgUnitIDs = ids
		}
	}
	if idsVal, ok := c.Get("devhub_actor_primary_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			projOpts.PrimaryUnitIDs = ids
		}
	}
	projects, _, err := storeI.ListProjects(c.Request.Context(), projOpts)
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
	return false, "not_platform_member", nil
}

func (h *PlatformHandler) denyRowRead(c *gin.Context, deniedReason string) {
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
