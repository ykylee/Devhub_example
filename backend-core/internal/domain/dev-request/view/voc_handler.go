package view

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/dev-request/repository"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)

// VocStore는 voc 관련 store interface. ADR-0028 §3.
type VocStore interface {
	CreateVoc(ctx context.Context, v domain.DevRequestVoc) (domain.DevRequestVoc, error)
	GetVocByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequestVoc, bool, error)
	GetVocByID(ctx context.Context, id string) (domain.DevRequestVoc, error)
	RouteVoc(ctx context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error)
	ListVocs(ctx context.Context, status, assigneeUserID string, limit, offset int) ([]domain.DevRequestVoc, error)
}

// NotificationStore는 in-app notification store interface. ADR-0028 §3.
type NotificationStore interface {
	InsertNotification(ctx context.Context, n domain.UserNotification) (domain.UserNotification, error)
	ListUnreadByUser(ctx context.Context, userID string, limit int) ([]domain.UserNotification, error)
	MarkRead(ctx context.Context, id, userID string) error
}

// VocHandlerConfig는 voc + notification handler 의 의존성.
type VocHandlerConfig struct {
	VocStore          VocStore
	NotificationStore NotificationStore
	AuditStore        AuditStore
}

type VocHandler struct {
	cfg VocHandlerConfig
}

func NewVocHandler(cfg VocHandlerConfig) *VocHandler {
	return &VocHandler{cfg: cfg}
}

var externalRefPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// RegisterVocRoutes는 /api/v1/dev-requests/{external_ref} + /me/notifications + /vocs 라우트 등록.
func RegisterVocRoutes(rg *gin.RouterGroup, h *VocHandler) {
	rg.POST("/dev-requests/:external_ref", h.createOrGetVoc)
	rg.POST("/dev-requests/:external_ref/route", h.routeVoc)
	rg.GET("/dev-requests/:external_ref", h.getVoc)

	rg.GET("/me/notifications", h.listMyNotifications)
	rg.POST("/me/notifications/:id/read", h.markMyNotificationRead)

	rg.GET("/vocs", h.listVocs)
}

// createVocRequest는 POST body 의 9 field. ADR-0028 §3 결정.
// external_ref 와 source_system 은 path param + query param 으로 받음 (body 외부).
type createVocRequest struct {
	Title         string     `json:"title"`
	Details       string     `json:"details"`
	Requester     string     `json:"requester"`
	ReqDepartment string     `json:"req_department"`
	Assignee      string     `json:"assignee"`
	DevDepartment string     `json:"dev_department"`
	RequestDate   *time.Time `json:"request_date,omitempty"`
	DevSchedule   string     `json:"dev_schedule"`
	SourceSystem  string     `json:"source_system"`
}

// validateExternalRefAbsent은 createVocRequest 의 body 에 external_ref 가 없음을 검증.
// external_ref 는 path param 으로 받음 (idempotency key) — body 에 동시 지정 시 reject.
func (r createVocRequest) validateExternalRefAbsent() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

// createOrGetVoc는 POST /api/v1/dev-requests/{external_ref}. ADR-0028 §3.
//
// idempotent: 동일 (source_system, external_ref) 면 기존 voc 200 반환.
// source_system 의 default 는 "manual" (frontend 직접 등록).
func (h *VocHandler) createOrGetVoc(c *gin.Context) {
	externalRef := strings.TrimSpace(c.Param("external_ref"))
	if !externalRefPattern.MatchString(externalRef) {
		c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", "external_ref is required and must match ^[A-Za-z0-9._-]{1,128}$"))
		return
	}

	var req createVocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", err.Error()))
		return
	}
	// external_ref 는 path param (idempotency key) → body 에 없음. source_system 은 body 또는 query.
	if req.SourceSystem == "" {
		req.SourceSystem = "manual"
	}

	if err := req.validateExternalRefAbsent(); err != nil {
		c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", err.Error()))
		return
	}

	if h.cfg.VocStore == nil {
		c.JSON(http.StatusServiceUnavailable, httphelp.EnvelopeErrorResponse("voc_store_unavailable", "voc store is not configured"))
		return
	}

	// 1) idempotency 사전 SELECT
	existing, found, err := h.cfg.VocStore.GetVocByExternalRef(c.Request.Context(), req.SourceSystem, externalRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	if found {
		c.JSON(http.StatusOK, vocResponse(existing))
		h.recordAuditBestEffort(c, "dev_request_voc.get_idempotent", "dev_request_voc", existing.ID, map[string]any{
			"source_system": req.SourceSystem,
			"external_ref":  externalRef,
		})
		return
	}

	// 2) INSERT
	v := domain.DevRequestVoc{
		ExternalRef:    externalRef,
		SourceSystem:   req.SourceSystem,
		Title:          strings.TrimSpace(req.Title),
		Details:        req.Details,
		Requester:      req.Requester,
		ReqDepartment:  req.ReqDepartment,
		AssigneeUserID: req.Assignee,
		DevDepartment:  req.DevDepartment,
		RequestDate:    req.RequestDate,
		DevSchedule:    req.DevSchedule,
		Status:         domain.DevRequestVocStatusReceived,
	}
	created, err := h.cfg.VocStore.CreateVoc(c.Request.Context(), v)
	if err != nil {
		if errors.Is(err, repository.ErrVocExternalRefConflict) {
			// race: 다른 request 가 먼저 INSERT. 재조회.
			existing, found, _ := h.cfg.VocStore.GetVocByExternalRef(c.Request.Context(), req.SourceSystem, externalRef)
			if found {
				c.JSON(http.StatusOK, vocResponse(existing))
				return
			}
		}
		c.JSON(http.StatusInternalServerError, httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}

	// 3) assignee 에게 in-app notification (ADR-0028 §3)
	if created.AssigneeUserID != "" {
		h.emitDevVocNotification(c.Request.Context(), created)
	}

	h.recordAuditBestEffort(c, "dev_request_voc.create", "dev_request_voc", created.ID, map[string]any{
		"source_system":      created.SourceSystem,
		"external_ref":       created.ExternalRef,
		"assignee_user_id":   created.AssigneeUserID,
	})

	c.JSON(http.StatusCreated, vocResponse(created))
}

// routeVocRequest는 POST /api/v1/dev-requests/{external_ref}/route body.
type routeVocRequest struct {
	ProjectID string `json:"project_id"`
}

func (h *VocHandler) routeVoc(c *gin.Context) {
	externalRef := strings.TrimSpace(c.Param("external_ref"))
	var req routeVocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, 	httphelp.EnvelopeErrorResponse("validation_failed", err.Error()))
		return
	}
	if strings.TrimSpace(req.ProjectID) == "" {
		c.JSON(http.StatusBadRequest, 	httphelp.EnvelopeErrorResponse("validation_failed", "project_id is required"))
		return
	}

	if h.cfg.VocStore == nil {
		c.JSON(http.StatusServiceUnavailable, 	httphelp.EnvelopeErrorResponse("voc_store_unavailable", "voc store is not configured"))
		return
	}

	// source_system = "manual" 가정 (frontend routing). 향후 ADR-0028 §3 (b) 정합.
	existing, found, err := h.cfg.VocStore.GetVocByExternalRef(c.Request.Context(), "manual", externalRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 	httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, 	httphelp.EnvelopeErrorResponse("voc_not_found", "dev_request_voc not found for source_system=manual external_ref="+externalRef))
		return
	}
	if existing.Status != domain.DevRequestVocStatusReceived {
		c.JSON(http.StatusConflict, 	httphelp.EnvelopeErrorResponse("voc_already_routed", "voc is already routed or closed; status="+string(existing.Status)))
		return
	}

	// dev-request 필드 (voc 의 9 field 복사 + project_id 결정)
	dr := domain.DevRequest{
		Title:                existing.Title,
		Details:              existing.Details,
		Requester:            existing.Requester,
		ReqDepartment:        existing.ReqDepartment,
		AssigneeUserID:       existing.AssigneeUserID,
		DevDepartment:        existing.DevDepartment,
		RequestDate:          existing.RequestDate,
		DevSchedule:          existing.DevSchedule,
		SourceSystem:         existing.SourceSystem,
		ExternalRef:          existing.ExternalRef,
		Status:               domain.DevRequestStatusPending,
		RegisteredTargetType: domain.DevRequestTargetPlatform, // routing 시 자동 결정: project_id 가 platform
		RegisteredTargetID:   req.ProjectID,
		ReceivedAt:           time.Now().UTC(),
	}
	voc, devReq, err := h.cfg.VocStore.RouteVoc(c.Request.Context(), existing.ID, req.ProjectID, dr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 	httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}

	// assignee 에게 dev-request 등록 알림
	if devReq.AssigneeUserID != "" {
		h.emitDevRequestNotification(c.Request.Context(), devReq)
	}

	h.recordAuditBestEffort(c, "dev_request_voc.route", "dev_request_voc", voc.ID, map[string]any{
		"project_id":     req.ProjectID,
		"dev_request_id": devReq.ID,
	})

	c.JSON(http.StatusOK, gin.H{
		"voc":          vocResponse(voc),
		"dev_request": devRequestResponse(devReq),
	})
}

func (h *VocHandler) getVoc(c *gin.Context) {
	externalRef := strings.TrimSpace(c.Param("external_ref"))
	existing, found, err := h.cfg.VocStore.GetVocByExternalRef(c.Request.Context(), "manual", externalRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 	httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, 	httphelp.EnvelopeErrorResponse("voc_not_found", "dev_request_voc not found"))
		return
	}
	c.JSON(http.StatusOK, vocResponse(existing))
}

func (h *VocHandler) listMyNotifications(c *gin.Context) {
	actor := httphelp.RequestActor(c)
	if actor.Login == "" {
		c.JSON(http.StatusUnauthorized, 	httphelp.EnvelopeErrorResponse("auth_failed", "user not authenticated"))
		return
	}
	if h.cfg.NotificationStore == nil {
		c.JSON(http.StatusServiceUnavailable, 	httphelp.EnvelopeErrorResponse("notification_store_unavailable", "notification store is not configured"))
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if _, err := parseLimit(l, &limit); err != nil {
			c.JSON(http.StatusBadRequest, 	httphelp.EnvelopeErrorResponse("validation_failed", err.Error()))
			return
		}
	}
	notes, err := h.cfg.NotificationStore.ListUnreadByUser(c.Request.Context(), actor.Login, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, 	httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  notes,
		"count": len(notes),
	})
}

// listVocs는 system_admin 도구 — status / assignee 필터 + pagination. ADR-0028 §6 carve (d).
func (h *VocHandler) listVocs(c *gin.Context) {
	actor := httphelp.RequestActor(c)
	if actor.Login == "" {
		c.JSON(http.StatusUnauthorized, httphelp.EnvelopeErrorResponse("auth_failed", "user not authenticated"))
		return
	}
	if h.cfg.VocStore == nil {
		c.JSON(http.StatusServiceUnavailable, httphelp.EnvelopeErrorResponse("voc_store_unavailable", "voc store is not configured"))
		return
	}
	statusFilter := strings.TrimSpace(c.Query("status"))
	if statusFilter != "" {
		switch domain.DevRequestVocStatus(statusFilter) {
		case domain.DevRequestVocStatusReceived, domain.DevRequestVocStatusRouted, domain.DevRequestVocStatusClosed:
		default:
			c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", "status must be one of received/routed/closed"))
			return
		}
	}
	assigneeFilter := strings.TrimSpace(c.Query("assignee"))

	limit := 50
	if l := c.Query("limit"); l != "" {
		if _, err := parseLimit(l, &limit); err != nil {
			c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", err.Error()))
			return
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		var n int
		for _, ch := range o {
			if ch < '0' || ch > '9' {
				c.JSON(http.StatusBadRequest, httphelp.EnvelopeErrorResponse("validation_failed", "offset must be a non-negative integer"))
				return
			}
			n = n*10 + int(ch-'0')
		}
		offset = n
	}

	vocs, err := h.cfg.VocStore.ListVocs(c.Request.Context(), statusFilter, assigneeFilter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	items := make([]gin.H, 0, len(vocs))
	for _, v := range vocs {
		items = append(items, vocResponse(v))
	}
	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"count":  len(items),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *VocHandler) markMyNotificationRead(c *gin.Context) {
	actor := httphelp.RequestActor(c)
	if actor.Login == "" {
		c.JSON(http.StatusUnauthorized, 	httphelp.EnvelopeErrorResponse("auth_failed", "user not authenticated"))
		return
	}
	if h.cfg.NotificationStore == nil {
		c.JSON(http.StatusServiceUnavailable, 	httphelp.EnvelopeErrorResponse("notification_store_unavailable", "notification store is not configured"))
		return
	}
	id := c.Param("id")
	if err := h.cfg.NotificationStore.MarkRead(c.Request.Context(), id, actor.Login); err != nil {
		if errors.Is(err, ErrNotificationNotFoundSentinel) {
			c.JSON(http.StatusNotFound, 	httphelp.EnvelopeErrorResponse("notification_not_found", "notification not found or not owned"))
			return
		}
		c.JSON(http.StatusInternalServerError, 	httphelp.EnvelopeErrorResponse("internal_error", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}


// ErrNotificationNotFound mirrors ErrNotificationNotFoundSentinel so voc_handler
// does not need to import the user-notification repository package.
var ErrNotificationNotFoundSentinel = errors.New("user_notification: not found or not owned")
// emitDevVocNotification은 voc 등록 시 assignee 에게 in-app notification 발송.
func (h *VocHandler) emitDevVocNotification(ctx context.Context, v domain.DevRequestVoc) {
	if h.cfg.NotificationStore == nil {
		return
	}
	n := domain.UserNotification{
		UserID: v.AssigneeUserID,
		Kind:   domain.UserNotificationKindDevVoc,
		RefID:  v.ID,
		Title:  "새 의뢰 도착: " + v.Title,
		Body:   "external_ref=" + v.ExternalRef + " 의 의뢰가 도착했습니다. project_id 가 미정인 상태로 알림이 발송되었습니다.",
	}
	_, _ = h.cfg.NotificationStore.InsertNotification(ctx, n)
}

// emitDevRequestNotification은 dev-request 자동 생성 시 assignee 에게 in-app notification.
func (h *VocHandler) emitDevRequestNotification(ctx context.Context, dr domain.DevRequest) {
	if h.cfg.NotificationStore == nil {
		return
	}
	n := domain.UserNotification{
		UserID: dr.AssigneeUserID,
		Kind:   domain.UserNotificationKindDevRequest,
		RefID:  dr.ID,
		Title:  "의뢰 라우팅 완료: " + dr.Title,
		Body:   "external_ref=" + dr.ExternalRef + " 의 의뢰가 project 에 매칭되었습니다.",
	}
	_, _ = h.cfg.NotificationStore.InsertNotification(ctx, n)
}

// recordAuditBestEffort는 audit 실패가 핵심 흐름을 막지 않도록 best-effort.
func (h *VocHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) {
	if h.cfg.AuditStore == nil {
		return
	}
	actor := httphelp.RequestActor(c)
	if payload == nil {
		payload = map[string]any{}
	}
	log := domain.AuditLog{
		ActorLogin:    actor.Login,
		SourceIP:       c.ClientIP(),
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		Payload:        payload,
		SourceType:     "api",
	}
	_, _ = h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), log)
}

func vocResponse(v domain.DevRequestVoc) gin.H {
	return gin.H{
		"id":              v.ID,
		"external_ref":    v.ExternalRef,
		"source_system":   v.SourceSystem,
		"title":           v.Title,
		"details":         v.Details,
		"requester":       v.Requester,
		"req_department":  v.ReqDepartment,
		"assignee":        v.AssigneeUserID,
		"dev_department":  v.DevDepartment,
		"request_date":    v.RequestDate,
		"dev_schedule":    v.DevSchedule,
		"status":          v.Status,
		"project_id":      v.ProjectID,
		"dev_request_id":  v.DevRequestID,
		"routed_at":       v.RoutedAt,
		"created_at":      v.CreatedAt,
		"updated_at":      v.UpdatedAt,
	}
}

func parseLimit(raw string, dst *int) (int, error) {
	// simple integer parse
	var n int
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, errors.New("limit must be a positive integer")
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return 0, errors.New("limit must be a positive integer")
	}
	if n > 200 {
		n = 200
	}
	*dst = n
	return n, nil
}
