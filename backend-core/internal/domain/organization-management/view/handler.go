package view

import (
	"context"
	"net/http"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)


type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}


type HRDBClient interface {
	Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error)
}

type OrganizationConfig struct {
	OrganizationStore     OrganizationStore
	HRDB                  HRDBClient
	AuditStore            AuditStore
	OnboardingGateEnabled bool
}

type OrganizationHandler struct {
	cfg OrganizationConfig
}

func NewOrganizationHandler(cfg OrganizationConfig) *OrganizationHandler {
	return &OrganizationHandler{cfg: cfg}
}

func (h *OrganizationHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *OrganizationHandler) RequireOnboardingFlag(c *gin.Context) bool {
	if h.cfg.OnboardingGateEnabled {
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{
		"status": "not_found",
		"code":   "onboarding_feature_disabled",
		"error":  "onboarding endpoints are disabled (DEVHUB_ONBOARDING_GATE_ENABLED=false)",
	})
	return false
}
