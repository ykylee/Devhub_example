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


type ApplicationStore interface {
	ListIntegrations(ctx context.Context, opts store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error)
	GetIntegration(ctx context.Context, id string) (domain.ProjectIntegration, error)
	CreateIntegration(ctx context.Context, p domain.ProjectIntegration) (domain.ProjectIntegration, error)
	UpdateIntegration(ctx context.Context, p domain.ProjectIntegration) (domain.ProjectIntegration, error)
	DeleteIntegration(ctx context.Context, id string) error
	ListIntegrationProviders(ctx context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error)
	GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error)
	GetIntegrationProviderByKey(ctx context.Context, key string) (domain.IntegrationProvider, error)
	CreateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error)
	UpdateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error)
	DeleteIntegrationProvider(ctx context.Context, id string) error
	CreateIntegrationSyncJob(ctx context.Context, providerID string, triggerBy string) (string, error)
	ListIntegrationBindings(ctx context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error)
	GetIntegrationBindingByID(ctx context.Context, id string) (domain.IntegrationBinding, error)
	CreateIntegrationBinding(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error)
	UpdateIntegrationBinding(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error)
	DeleteIntegrationBinding(ctx context.Context, id string) error
}

type WebhookEventStore interface {
	SaveWebhookEvent(ctx context.Context, ev store.WebhookEvent) (int64, error)
	ListWebhookEvents(ctx context.Context, opts store.ListWebhookEventsOptions) ([]store.WebhookEvent, error)
}

type WebhookEventProcessor interface {
	Process(ctx context.Context, ev store.WebhookEvent) error
}

type IntegrationConfig struct {
	ApplicationStore  ApplicationStore
	EventStore        WebhookEventStore
	EventProcessor    WebhookEventProcessor
	ExternalTaskStore ExternalTaskStore
	AuditStore        AuditStore
}

type IntegrationHandler struct {
	cfg IntegrationConfig
}

func NewIntegrationHandler(cfg IntegrationConfig) *IntegrationHandler {
	return &IntegrationHandler{cfg: cfg}
}

func (h *IntegrationHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *IntegrationHandler) ApplicationStoreOrUnavailable(c *gin.Context) (ApplicationStore, bool) {
	if h.cfg.ApplicationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "application store is not configured"})
		return nil, false
	}
	return h.cfg.ApplicationStore, true
}

// providerHasCapability reports whether the provider declares any of the given
// capabilities (OR semantics — same as the repository-integration view helper
// and the main-HEAD baseline before the gemini code-cleanup split).
func providerHasCapability(p domain.IntegrationProvider, caps ...string) bool {
	for _, have := range p.Capabilities {
		for _, want := range caps {
			if have == want {
				return true
			}
		}
	}
	return false
}

func actorLogin(c *gin.Context) string {
	val, _ := c.Get("devhub_actor_login")
	s, _ := val.(string)
	return s
}

func firstHeader(c *gin.Context, names ...string) string {
	for _, name := range names {
		if v := c.GetHeader(name); v != "" {
			return v
		}
	}
	return ""
}
