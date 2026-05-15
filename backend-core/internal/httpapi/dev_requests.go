package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// DevRequestStore는 dev_requests row CRUD interface. ARCH-DREQ-05 / API §14.
// RegisterDevRequestWithNew* 두 메서드는 신규 application/project 생성 + dev_request
// 상태 갱신을 단일 트랜잭션으로 수행한다 (REQ-FR-DREQ-005, ADR-0013 §5, sprint
// claude/work_260515-m). 기존 MarkDevRequestRegistered 는 legacy "기존 target_id
// 매핑" path 에서 그대로 사용한다.
type DevRequestStore interface {
	CreateDevRequest(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error)
	GetDevRequest(ctx context.Context, id string) (domain.DevRequest, error)
	GetDevRequestByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error)
	ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
	TransitionDevRequestStatus(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error)
	ReassignDevRequest(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error)
	MarkDevRequestRegistered(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error)
	RegisterDevRequestWithNewApplication(ctx context.Context, drID string, app domain.Application, primaryRepo *domain.ApplicationRepository) (domain.DevRequest, domain.Application, error)
	RegisterDevRequestWithNewProject(ctx context.Context, drID string, project domain.Project) (domain.DevRequest, domain.Project, error)
}

func (h *Handler) devRequestStoreOrUnavailable(c *gin.Context) (DevRequestStore, bool) {
	if h.cfg.DevRequestStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request store is not configured",
		})
		return nil, false
	}
	return h.cfg.DevRequestStore, true
}

func devRequestResponse(dr domain.DevRequest) gin.H {
	resp := gin.H{
		"id":                     dr.ID,
		"title":                  dr.Title,
		"details":                dr.Details,
		"requester":              dr.Requester,
		"assignee_user_id":       dr.AssigneeUserID,
		"source_system":          dr.SourceSystem,
		"external_ref":           dr.ExternalRef,
		"status":                 string(dr.Status),
		"registered_target_type": string(dr.RegisteredTargetType),
		"registered_target_id":   dr.RegisteredTargetID,
		"rejected_reason":        dr.RejectedReason,
		"received_at":            dr.ReceivedAt.UTC().Format(time.RFC3339),
		"created_at":             dr.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             dr.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return resp
}

// --- API-59: POST /api/v1/dev-requests (외부 수신) ---

type createDevRequestRequest struct {
	Title          string `json:"title"`
	Details        string `json:"details"`
	Requester      string `json:"requester"`
	AssigneeUserID string `json:"assignee_user_id"`
	ExternalRef    string `json:"external_ref"`
	// source_system 은 body 에서 무시 (ADR-0012 §4.1.2 spoofing 방지). intake token 매핑값 사용.
}

func (h *Handler) intakeDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}

	// intake auth middleware 가 컨텍스트에 source_system 을 set 해야 한다.
	sourceSystemVal, ok := c.Get(ctxKeyDREQSourceSystem)
	if !ok {
		// middleware 가 통과하지 않은 이상한 경로 — 안전 deny.
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "intake auth context missing",
			"code":   "auth_intake_token_missing",
		})
		return
	}
	sourceSystem, _ := sourceSystemVal.(string)
	if sourceSystem == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "intake source_system is empty",
			"code":   "auth_intake_token_invalid",
		})
		return
	}

	var req createDevRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	// idempotency: (source_system, external_ref) 매칭 시 기존 row 반환.
	if req.ExternalRef != "" {
		existing, err := storeI.GetDevRequestByExternalRef(c.Request.Context(), sourceSystem, req.ExternalRef)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(existing)})
			return
		}
		if !errors.Is(err, store.ErrNotFound) {
			writeServerError(c, err, "dev_request.intake.idempotency_lookup")
			return
		}
	}

	// 검증: 필수 필드 + assignee 존재. 실패 시 rejected (invalid_intake) 로 저장 (audit 보존).
	rejectedReason := validateIntakeRequest(req)
	status := domain.DevRequestStatusPending
	if rejectedReason != "" {
		status = domain.DevRequestStatusRejected
	}

	dr := domain.DevRequest{
		Title:          strings.TrimSpace(req.Title),
		Details:        req.Details,
		Requester:      strings.TrimSpace(req.Requester),
		AssigneeUserID: strings.TrimSpace(req.AssigneeUserID),
		SourceSystem:   sourceSystem,
		ExternalRef:    req.ExternalRef,
		Status:         status,
		RejectedReason: rejectedReason,
		ReceivedAt:     time.Now().UTC(),
	}
	created, err := storeI.CreateDevRequest(c.Request.Context(), dr)
	if errors.Is(err, store.ErrConflict) {
		// race: idempotency 검사와 insert 사이 다른 worker 가 insert. 다시 lookup.
		if req.ExternalRef != "" {
			if existing, lookupErr := storeI.GetDevRequestByExternalRef(c.Request.Context(), sourceSystem, req.ExternalRef); lookupErr == nil {
				c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(existing)})
				return
			}
		}
		// assignee FK violation (assignee_user_id 가 존재하지 않는 user) 케이스 —
		// REQ-FR-DREQ-002 의 'drop 안 함' 정책 위반 방지 (codex PR #124 review P1).
		// migration 000025 가 assignee_user_id 를 NULLABLE 로 alter 했으므로
		// assignee 를 NULL 로 두고 rejected (invalid_intake) row 를 재저장 → audit 보존.
		dr.Status = domain.DevRequestStatusRejected
		dr.AssigneeUserID = ""
		dr.RejectedReason = combineRejectedReason(rejectedReason, "assignee_user_id not found")
		created, err = storeI.CreateDevRequest(c.Request.Context(), dr)
		if err != nil {
			writeServerError(c, err, "dev_request.intake.create_rejected_fallback")
			return
		}
		h.recordAuditBestEffort(c, "dev_request.received", "dev_request", created.ID, map[string]any{
			"source_system":   sourceSystem,
			"external_ref":    created.ExternalRef,
			"status":          string(created.Status),
			"assignee_user_id": "",
		})
		h.recordAuditBestEffort(c, "dev_request.rejected", "dev_request", created.ID, map[string]any{
			"denied_reason":   "invalid_intake",
			"rejected_reason": dr.RejectedReason,
		})
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": devRequestResponse(created)})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.intake.create")
		return
	}

	h.recordAuditBestEffort(c, "dev_request.received", "dev_request", created.ID, map[string]any{
		"source_system":   sourceSystem,
		"external_ref":    created.ExternalRef,
		"assignee_user_id": created.AssigneeUserID,
		"status":          string(created.Status),
	})
	if rejectedReason != "" {
		h.recordAuditBestEffort(c, "dev_request.rejected", "dev_request", created.ID, map[string]any{
			"denied_reason":   "invalid_intake",
			"rejected_reason": rejectedReason,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": devRequestResponse(created)})
}

func combineRejectedReason(existing, extra string) string {
	if existing == "" {
		return extra
	}
	if extra == "" {
		return existing
	}
	return existing + "; " + extra
}

func validateIntakeRequest(req createDevRequestRequest) string {
	var problems []string
	if strings.TrimSpace(req.Title) == "" {
		problems = append(problems, "title is required")
	}
	if len(req.Title) > 200 {
		problems = append(problems, "title exceeds 200 chars")
	}
	if strings.TrimSpace(req.Requester) == "" {
		problems = append(problems, "requester is required")
	}
	if strings.TrimSpace(req.AssigneeUserID) == "" {
		problems = append(problems, "assignee_user_id is required")
	}
	if len(problems) == 0 {
		return ""
	}
	return strings.Join(problems, "; ")
}

// --- API-60: GET /api/v1/dev-requests (목록) ---

func (h *Handler) listDevRequests(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}

	// limit/offset 파싱 (codex PR #124 review P2). 기본 50, 최대 100 clamp.
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	offset := 0
	if raw := c.Query("offset"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			offset = n
		}
	}
	opts := store.DevRequestListOptions{
		AssigneeUserID: c.Query("assignee_user_id"),
		SourceSystem:   c.Query("source_system"),
		Limit:          limit,
		Offset:         offset,
	}
	if raw := c.Query("status"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			opts.Statuses = append(opts.Statuses, domain.DevRequestStatus(s))
		}
	}

	// row-level filter: system_admin / pmo_manager 외에는 본인 의뢰만 (ARCH-DREQ-04).
	actorLogin, _ := c.Get("devhub_actor_login")
	actorRole, _ := c.Get("devhub_actor_role")
	login, _ := actorLogin.(string)
	role, _ := actorRole.(string)
	if !devFallbackEnabled(c) && role != string(domain.AppRoleSystemAdmin) && role != string(domain.AppRolePMOManager) {
		opts.AssigneeUserID = login
	}

	drs, total, err := storeI.ListDevRequests(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "dev_request.list")
		return
	}
	resp := make([]gin.H, 0, len(drs))
	for _, dr := range drs {
		resp = append(resp, devRequestResponse(dr))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta": gin.H{
			"total":  total,
			"limit":  opts.Limit,
			"offset": opts.Offset,
		},
	})
}

// --- API-61: GET /api/v1/dev-requests/:id (상세) ---

func (h *Handler) getDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("dev_request_id")
	dr, err := storeI.GetDevRequest(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.get")
		return
	}
	if !h.enforceRowOwnership(c, dr.AssigneeUserID, string(domain.AppRolePMOManager)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(dr)})
}

// --- API-62: POST /api/v1/dev-requests/:id/register (Promote) ---

// registerDevRequestRequest accepts one of two mutually-exclusive payload shapes:
//
//  1. legacy "기존 target_id 매핑" — caller supplies `target_id` referring to an
//     already-existing Application/Project row. handler calls
//     MarkDevRequestRegistered (no transaction).
//  2. new "promote + create" — caller supplies `application_payload` or
//     `project_payload` (matching the field's target_type). handler calls the
//     transactional store method which creates the target and flips the
//     dev_request row to status='registered' atomically (REQ-FR-DREQ-005,
//     ADR-0013 §5).
//
// 둘 다 채우거나 둘 다 비우면 400. application_payload + target_type='project'
// (또는 그 반대) 와 같은 mismatch 도 400.
type registerDevRequestRequest struct {
	TargetType         string                          `json:"target_type"` // "application" | "project"
	TargetID           string                          `json:"target_id,omitempty"`
	ApplicationPayload *registerApplicationPayload     `json:"application_payload,omitempty"`
	ProjectPayload     *registerProjectPayload         `json:"project_payload,omitempty"`
}

type registerApplicationPayload struct {
	Key               string                          `json:"key"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	OwnerUserID       string                          `json:"owner_user_id"`
	LeaderUserID      string                          `json:"leader_user_id"`
	DevelopmentUnitID string                          `json:"development_unit_id"`
	StartDate         string                          `json:"start_date"`
	DueDate           string                          `json:"due_date"`
	Visibility        string                          `json:"visibility"`
	Status            string                          `json:"status"`
	PrimaryRepo       *registerPrimaryRepoPayload     `json:"primary_repo,omitempty"`
}

type registerPrimaryRepoPayload struct {
	RepoProvider   string `json:"repo_provider"`
	RepoFullName   string `json:"repo_full_name"`
	ExternalRepoID string `json:"external_repo_id"`
	Role           string `json:"role"`
}

type registerProjectPayload struct {
	ApplicationID string `json:"application_id"` // optional
	RepositoryID  int64  `json:"repository_id"`  // required FK
	Key           string `json:"key"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	OwnerUserID   string `json:"owner_user_id"`
	StartDate     string `json:"start_date"`
	DueDate       string `json:"due_date"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
}

func (h *Handler) registerDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("dev_request_id")
	var req registerDevRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	targetType := domain.DevRequestTargetType(req.TargetType)
	if targetType != domain.DevRequestTargetApplication && targetType != domain.DevRequestTargetProject {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "target_type must be 'application' or 'project'",
			"code":   "dev_request_register_target_invalid",
		})
		return
	}

	// Mutual exclusion: exactly one of target_id / application_payload / project_payload.
	hasLegacyID := strings.TrimSpace(req.TargetID) != ""
	hasAppPayload := req.ApplicationPayload != nil
	hasProjectPayload := req.ProjectPayload != nil
	payloadCount := 0
	if hasLegacyID {
		payloadCount++
	}
	if hasAppPayload {
		payloadCount++
	}
	if hasProjectPayload {
		payloadCount++
	}
	if payloadCount != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "exactly one of target_id / application_payload / project_payload is required",
			"code":   "dev_request_register_payload_invalid",
		})
		return
	}
	if hasAppPayload && targetType != domain.DevRequestTargetApplication {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "application_payload requires target_type='application'",
			"code":   "dev_request_register_payload_invalid",
		})
		return
	}
	if hasProjectPayload && targetType != domain.DevRequestTargetProject {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "project_payload requires target_type='project'",
			"code":   "dev_request_register_payload_invalid",
		})
		return
	}

	current, err := storeI.GetDevRequest(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.register.lookup")
		return
	}
	if !h.enforceRowOwnership(c, current.AssigneeUserID, string(domain.AppRolePMOManager)) {
		return
	}
	if current.Status != domain.DevRequestStatusPending && current.Status != domain.DevRequestStatusInReview {
		c.JSON(http.StatusConflict, gin.H{
			"status": "rejected",
			"error":  "dev_request is already registered/rejected/closed",
			"code":   "dev_request_already_registered",
		})
		return
	}

	switch {
	case hasLegacyID:
		h.registerDevRequestLegacy(c, storeI, id, targetType, req.TargetID)
	case hasAppPayload:
		h.registerDevRequestWithNewApplication(c, storeI, id, req.ApplicationPayload)
	case hasProjectPayload:
		h.registerDevRequestWithNewProject(c, storeI, id, req.ProjectPayload)
	}
}

// registerDevRequestLegacy is the pre-Promote-Tx path: caller maps the dev_request to
// an already-existing Application/Project. Single UPDATE on dev_requests only.
func (h *Handler) registerDevRequestLegacy(c *gin.Context, storeI DevRequestStore, drID string, targetType domain.DevRequestTargetType, targetID string) {
	updated, err := storeI.MarkDevRequestRegistered(c.Request.Context(), drID, targetType, targetID)
	if err != nil {
		writeServerError(c, err, "dev_request.register.mark")
		return
	}
	h.recordAuditBestEffort(c, "dev_request.registered", "dev_request", drID, map[string]any{
		"target_type": string(targetType),
		"target_id":   targetID,
		"created":     false,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"dev_request": devRequestResponse(updated),
			"registered_target": gin.H{
				"target_type": string(targetType),
				"target_id":   targetID,
				"created":     false,
			},
		},
	})
}

// registerDevRequestWithNewApplication invokes the transactional store method that
// (1) inserts a new Application, (2) optionally links one primary repository, and
// (3) flips the dev_request row to status='registered' — all atomically.
func (h *Handler) registerDevRequestWithNewApplication(c *gin.Context, storeI DevRequestStore, drID string, p *registerApplicationPayload) {
	if !applicationKeyPattern.MatchString(p.Key) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "application_payload.key must match ^[A-Za-z0-9]{10}$",
			"code":   "invalid_application_key",
		})
		return
	}
	if strings.TrimSpace(p.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.name is required"})
		return
	}
	if strings.TrimSpace(p.OwnerUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.owner_user_id is required"})
		return
	}
	if strings.TrimSpace(p.LeaderUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.leader_user_id is required"})
		return
	}
	if strings.TrimSpace(p.DevelopmentUnitID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.development_unit_id is required"})
		return
	}
	if !validApplicationVisibilities[p.Visibility] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.visibility must be one of public/internal/restricted"})
		return
	}
	if !validApplicationStatuses[p.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.status must be one of planning/active/on_hold/closed/archived"})
		return
	}
	startDate, err := parseDate(p.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.start_date must be YYYY-MM-DD"})
		return
	}
	dueDate, err := parseDate(p.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_payload.due_date must be YYYY-MM-DD"})
		return
	}
	app := domain.Application{
		Key:               p.Key,
		Name:              p.Name,
		Description:       p.Description,
		Status:            domain.ApplicationStatus(p.Status),
		Visibility:        domain.ApplicationVisibility(p.Visibility),
		OwnerUserID:       p.OwnerUserID,
		LeaderUserID:      p.LeaderUserID,
		DevelopmentUnitID: p.DevelopmentUnitID,
		StartDate:         startDate,
		DueDate:           dueDate,
	}
	var primaryRepo *domain.ApplicationRepository
	if p.PrimaryRepo != nil {
		if strings.TrimSpace(p.PrimaryRepo.RepoProvider) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "primary_repo.repo_provider is required"})
			return
		}
		if strings.TrimSpace(p.PrimaryRepo.RepoFullName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "primary_repo.repo_full_name is required"})
			return
		}
		role := p.PrimaryRepo.Role
		if strings.TrimSpace(role) == "" {
			role = string(domain.ApplicationRepositoryRolePrimary)
		}
		// codex hotfix #4 P1 (sprint claude/work_260515-n) — primary_repo.role
		// 가 application_repositories_role_check 의 허용 값 (primary/sub/shared)
		// 외이면 store insert 가 PG CHECK 위반으로 500 을 일으킨다. legacy
		// createApplicationRepository handler 와 동일한 application-level gate 를
		// 적용해 422 invalid_repo_link_role 로 surface.
		if !validApplicationRepoRoles[role] {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "rejected",
				"error":  "primary_repo.role must be one of primary/sub/shared",
				"code":   "invalid_repo_link_role",
			})
			return
		}
		// codex hotfix #4 P2 (sprint claude/work_260515-n) — legacy
		// createApplicationRepository 는 ListSCMProviders 로 enabled 여부를
		// 검증해 unsupported_repo_provider 422 를 반환한다. promote path 가 이
		// gate 를 우회하면 disabled provider 도 application_repositories 행을
		// 만들어 정책이 깨진다. ApplicationStore 가 wire 안 됐을 땐 dev 환경의
		// in-memory store 케이스이므로 통과시킨다 (devRequests test 케이스는
		// ApplicationStore 없이 동작 — 통합 검증은 production wire 에서).
		if h.cfg.ApplicationStore != nil {
			providers, err := h.cfg.ApplicationStore.ListSCMProviders(c.Request.Context())
			if err != nil {
				writeServerError(c, err, "dev_request.register.promote_application.lookup_provider")
				return
			}
			enabled := false
			for _, prov := range providers {
				if prov.ProviderKey == p.PrimaryRepo.RepoProvider && prov.Enabled {
					enabled = true
					break
				}
			}
			if !enabled {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"status": "rejected",
					"error":  "primary_repo.repo_provider is not registered or disabled",
					"code":   "unsupported_repo_provider",
				})
				return
			}
		}
		primaryRepo = &domain.ApplicationRepository{
			RepoProvider:   p.PrimaryRepo.RepoProvider,
			RepoFullName:   p.PrimaryRepo.RepoFullName,
			ExternalRepoID: p.PrimaryRepo.ExternalRepoID,
			Role:           domain.ApplicationRepositoryRole(role),
			SyncStatus:     domain.SyncStatusRequested,
		}
	}

	updatedDR, createdApp, err := storeI.RegisterDevRequestWithNewApplication(c.Request.Context(), drID, app, primaryRepo)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "application key conflict or referenced owner/leader/development_unit not found",
			"code":   "application_key_conflict",
		})
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.register.promote_application")
		return
	}
	h.recordAuditBestEffort(c, "application.created", "application", createdApp.ID, map[string]any{
		"key":              createdApp.Key,
		"status":           string(createdApp.Status),
		"via_dev_request":  drID,
	})
	h.recordAuditBestEffort(c, "dev_request.registered", "dev_request", drID, map[string]any{
		"target_type": string(domain.DevRequestTargetApplication),
		"target_id":   createdApp.ID,
		"created":     true,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"dev_request": devRequestResponse(updatedDR),
			"registered_target": gin.H{
				"target_type": string(domain.DevRequestTargetApplication),
				"target_id":   createdApp.ID,
				"created":     true,
				"application": applicationResponse(createdApp),
			},
		},
	})
}

// registerDevRequestWithNewProject invokes the transactional store method that
// (1) inserts a new Project and (2) flips the dev_request row to status='registered'
// — all atomically.
func (h *Handler) registerDevRequestWithNewProject(c *gin.Context, storeI DevRequestStore, drID string, p *registerProjectPayload) {
	if p.RepositoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.repository_id is required"})
		return
	}
	if strings.TrimSpace(p.Key) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.key is required"})
		return
	}
	if strings.TrimSpace(p.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.name is required"})
		return
	}
	if strings.TrimSpace(p.OwnerUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.owner_user_id is required"})
		return
	}
	if !validApplicationVisibilities[p.Visibility] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.visibility must be one of public/internal/restricted"})
		return
	}
	if !validApplicationStatuses[p.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.status must be one of planning/active/on_hold/closed/archived"})
		return
	}
	startDate, err := parseDate(p.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.start_date must be YYYY-MM-DD"})
		return
	}
	dueDate, err := parseDate(p.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_payload.due_date must be YYYY-MM-DD"})
		return
	}
	project := domain.Project{
		ApplicationID: p.ApplicationID,
		RepositoryID:  p.RepositoryID,
		Key:           p.Key,
		Name:          p.Name,
		Description:   p.Description,
		Status:        domain.ApplicationStatus(p.Status),
		Visibility:    domain.ApplicationVisibility(p.Visibility),
		OwnerUserID:   p.OwnerUserID,
		StartDate:     startDate,
		DueDate:       dueDate,
	}
	updatedDR, createdProject, err := storeI.RegisterDevRequestWithNewProject(c.Request.Context(), drID, project)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "project key conflict or referenced application/repository not found",
			"code":   "project_key_conflict",
		})
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.register.promote_project")
		return
	}
	h.recordAuditBestEffort(c, "project.created", "project", createdProject.ID, map[string]any{
		"key":              createdProject.Key,
		"repository_id":    createdProject.RepositoryID,
		"status":           string(createdProject.Status),
		"via_dev_request":  drID,
	})
	h.recordAuditBestEffort(c, "dev_request.registered", "dev_request", drID, map[string]any{
		"target_type": string(domain.DevRequestTargetProject),
		"target_id":   createdProject.ID,
		"created":     true,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"dev_request": devRequestResponse(updatedDR),
			"registered_target": gin.H{
				"target_type": string(domain.DevRequestTargetProject),
				"target_id":   createdProject.ID,
				"created":     true,
				"project":     projectResponse(createdProject),
			},
		},
	})
}

// --- API-63: POST /api/v1/dev-requests/:id/reject ---

type rejectDevRequestRequest struct {
	RejectedReason string `json:"rejected_reason"`
}

func (h *Handler) rejectDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("dev_request_id")
	var req rejectDevRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.RejectedReason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "rejected_reason is required",
			"code":   "dev_request_reason_required",
		})
		return
	}
	current, err := storeI.GetDevRequest(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.reject.lookup")
		return
	}
	if !h.enforceRowOwnership(c, current.AssigneeUserID, string(domain.AppRolePMOManager)) {
		return
	}
	if !domain.IsValidDevRequestTransition(current.Status, domain.DevRequestStatusRejected) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "rejected",
			"error":  "invalid status transition to rejected",
			"code":   "invalid_status_transition",
			"from":   string(current.Status),
			"to":     "rejected",
		})
		return
	}
	updated, err := storeI.TransitionDevRequestStatus(c.Request.Context(), id, domain.DevRequestStatusRejected, req.RejectedReason)
	if err != nil {
		writeServerError(c, err, "dev_request.reject.transition")
		return
	}
	h.recordAuditBestEffort(c, "dev_request.rejected", "dev_request", id, map[string]any{
		"rejected_reason": req.RejectedReason,
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(updated)})
}

// --- API-64: PATCH /api/v1/dev-requests/:id (Reassign) ---

type patchDevRequestRequest struct {
	AssigneeUserID *string `json:"assignee_user_id"`
}

func (h *Handler) patchDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}
	// system_admin only — handler 가 추가 검증 (route gate 의 edit 는 pmo_manager 도 통과하므로).
	actorRoleVal, _ := c.Get("devhub_actor_role")
	actorRole, _ := actorRoleVal.(string)
	if !devFallbackEnabled(c) && actorRole != string(domain.AppRoleSystemAdmin) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "reassign requires system_admin",
			"code":   "auth_row_denied",
		})
		return
	}

	id := c.Param("dev_request_id")
	var req patchDevRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.AssigneeUserID == nil || strings.TrimSpace(*req.AssigneeUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "assignee_user_id is required",
		})
		return
	}

	current, err := storeI.GetDevRequest(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.reassign.lookup")
		return
	}

	updated, err := storeI.ReassignDevRequest(c.Request.Context(), id, *req.AssigneeUserID)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "rejected",
			"error":  "assignee_user_id does not exist",
			"code":   "dev_request_assignee_not_found",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.reassign.update")
		return
	}
	h.recordAuditBestEffort(c, "dev_request.reassigned", "dev_request", id, map[string]any{
		"from_assignee": current.AssigneeUserID,
		"to_assignee":   updated.AssigneeUserID,
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(updated)})
}

// --- API-65: DELETE /api/v1/dev-requests/:id (Close) ---

func (h *Handler) closeDevRequest(c *gin.Context) {
	storeI, ok := h.devRequestStoreOrUnavailable(c)
	if !ok {
		return
	}
	// system_admin only — REQ-FR-DREQ-008 + ARCH-DREQ-04 (codex PR #121 review P1).
	actorRoleVal, _ := c.Get("devhub_actor_role")
	actorRole, _ := actorRoleVal.(string)
	if !devFallbackEnabled(c) && actorRole != string(domain.AppRoleSystemAdmin) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "close requires system_admin",
			"code":   "auth_row_denied",
		})
		return
	}

	id := c.Param("dev_request_id")
	current, err := storeI.GetDevRequest(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "dev_request not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.close.lookup")
		return
	}
	if !domain.IsValidDevRequestTransition(current.Status, domain.DevRequestStatusClosed) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "close is only allowed from registered/rejected status",
			"code":   "invalid_status_transition_close",
			"from":   string(current.Status),
		})
		return
	}
	updated, err := storeI.TransitionDevRequestStatus(c.Request.Context(), id, domain.DevRequestStatusClosed, "")
	if err != nil {
		writeServerError(c, err, "dev_request.close.transition")
		return
	}
	h.recordAuditBestEffort(c, "dev_request.closed", "dev_request", id, map[string]any{
		"from_status": string(current.Status),
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": devRequestResponse(updated)})
}
