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
type DevRequestStore interface {
	CreateDevRequest(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error)
	GetDevRequest(ctx context.Context, id string) (domain.DevRequest, error)
	GetDevRequestByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error)
	ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
	TransitionDevRequestStatus(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error)
	ReassignDevRequest(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error)
	MarkDevRequestRegistered(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error)
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

type registerDevRequestRequest struct {
	TargetType string `json:"target_type"` // "application" | "project"
	// target_payload 는 application/project handler 의 요청 schema 와 동일.
	// 본 sprint 는 단순 매핑으로 구현 — target_id 를 직접 받는 경량 형태 채택.
	// 이상적으로는 target 신규 생성을 트랜잭션으로 묶지만 (REQ-FR-DREQ-005), 본
	// sprint 는 ApplicationStore / DevRequestStore 두 인터페이스 사이 트랜잭션
	// 분리 한계로 *기존* application/project id 를 매핑하는 흐름으로 1차 구현.
	// 신규 생성 + 단일 트랜잭션은 carve out (DREQ-Backend follow-up).
	TargetID string `json:"target_id"`
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
	if strings.TrimSpace(req.TargetID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "target_id is required",
			"code":   "dev_request_register_target_invalid",
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

	updated, err := storeI.MarkDevRequestRegistered(c.Request.Context(), id, targetType, req.TargetID)
	if err != nil {
		writeServerError(c, err, "dev_request.register.mark")
		return
	}
	h.recordAuditBestEffort(c, "dev_request.registered", "dev_request", id, map[string]any{
		"target_type": string(targetType),
		"target_id":   req.TargetID,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"dev_request": devRequestResponse(updated),
			"registered_target": gin.H{
				"target_type": string(targetType),
				"target_id":   req.TargetID,
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
