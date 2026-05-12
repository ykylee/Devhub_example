package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type auditLogResponse struct {
	AuditID    string         `json:"audit_id"`
	ActorLogin string         `json:"actor_login"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	CommandID  string         `json:"command_id,omitempty"`
	Payload    map[string]any `json:"payload"`
	SourceIP   string         `json:"source_ip,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	SourceType string         `json:"source_type,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (h Handler) listAuditLogs(c *gin.Context) {
	if h.cfg.AuditStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "audit store is not configured",
		})
		return
	}
	limit, err := parseBoundedInt(c.DefaultQuery("limit", "50"), 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be an integer between 1 and 100"})
		return
	}
	offset, err := parseBoundedInt(c.DefaultQuery("offset", "0"), 0, 100000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be a non-negative integer"})
		return
	}

	opts := store.ListAuditLogsOptions{
		Limit:      limit,
		Offset:     offset,
		ActorLogin: strings.TrimSpace(c.Query("actor_login")),
		Action:     strings.TrimSpace(c.Query("action")),
		TargetType: strings.TrimSpace(c.Query("target_type")),
		TargetID:   strings.TrimSpace(c.Query("target_id")),
		CommandID:  strings.TrimSpace(c.Query("command_id")),
	}
	logs, err := h.cfg.AuditStore.ListAuditLogs(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "audit.list_logs")
		return
	}

	data := make([]auditLogResponse, 0, len(logs))
	for _, log := range logs {
		data = append(data, auditLogFromDomain(log))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(data),
		},
	})
}

func (h Handler) recordAudit(c *gin.Context, action, targetType, targetID string, payload map[string]any) (domain.AuditLog, error) {
	if h.cfg.AuditStore == nil {
		return domain.AuditLog{}, nil
	}
	actor := requestActor(c)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["actor_source"] = actor.Source
	// T-M1-04: stamp request-scoped operator-actor context (request_id /
	// source_ip / source_type). The middleware (requireRequestID) puts
	// request_id on the gin context; authenticateActor classifies the
	// source_type. clientIPFrom proxies gin.Context.ClientIP for nil-safety
	// in test fakes that pass a bare context.
	return h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), domain.AuditLog{
		ActorLogin: actor.Login,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Payload:    payload,
		SourceIP:   clientIPFrom(c),
		RequestID:  requestIDFrom(c),
		SourceType: sourceTypeFrom(c),
	})
}

// recordAuditBestEffort logs and swallows audit errors so callers (which have already committed the main mutation) do not surface a 500 that triggers duplicate retries.
func (h Handler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
	auditLog, err := h.recordAudit(c, action, targetType, targetID, payload)
	if err != nil {
		logRequest(c, "audit log persistence failed: action=%s target=%s/%s err=%v", action, targetType, targetID, err)
	}
	return auditLog
}

func auditLogFromDomain(log domain.AuditLog) auditLogResponse {
	payload := log.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	return auditLogResponse{
		AuditID:    log.AuditID,
		ActorLogin: log.ActorLogin,
		Action:     log.Action,
		TargetType: log.TargetType,
		TargetID:   log.TargetID,
		CommandID:  log.CommandID,
		Payload:    payload,
		SourceIP:   log.SourceIP,
		RequestID:  log.RequestID,
		SourceType: string(log.SourceType),
		CreatedAt:  log.CreatedAt,
	}
}
