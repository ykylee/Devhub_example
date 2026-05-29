package view

import (
	"context"
	"net/http"
	"strings"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)


type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}


type ApplicationStore interface {
	GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error)
	ListRepositoriesByProvider(ctx context.Context, providerID string) ([]domain.Repository, error)
	UpsertRepository(ctx context.Context, r domain.Repository) error
}

type RepositoryIntegrationConfig struct {
	ApplicationStore ApplicationStore
	AuditStore       AuditStore
}

type RepositoryIntegrationHandler struct {
	cfg RepositoryIntegrationConfig
}

func NewRepositoryIntegrationHandler(cfg RepositoryIntegrationConfig) *RepositoryIntegrationHandler {
	return &RepositoryIntegrationHandler{cfg: cfg}
}

func (h *RepositoryIntegrationHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *RepositoryIntegrationHandler) ApplicationStoreOrUnavailable(c *gin.Context) (ApplicationStore, bool) {
	if h.cfg.ApplicationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "application store is not configured"})
		return nil, false
	}
	return h.cfg.ApplicationStore, true
}

func normalizeProviderSDKKey(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if strings.Contains(v, "gitea") {
		return "gitea"
	}
	if strings.Contains(v, "github") {
		return "github"
	}
	return v
}
