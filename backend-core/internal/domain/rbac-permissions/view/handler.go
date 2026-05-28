package view

import (
	"context"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)


type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}


type RBACConfig struct {
	RBACStore       RBACStore
	PermissionCache *PermissionCache
	AuthDevFallback bool
	AuditStore      AuditStore
}

type RBACHandler struct {
	cfg RBACConfig
}

func NewRBACHandler(cfg RBACConfig) *RBACHandler {
	return &RBACHandler{cfg: cfg}
}

func (h *RBACHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func addAuditMeta(resp gin.H, log domain.AuditLog) {
	if log.AuditID != "" {
		resp["audit_log_id"] = log.AuditID
	}
}
