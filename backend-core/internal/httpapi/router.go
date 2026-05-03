package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type WebhookEventStore interface {
	SaveWebhookEvent(context.Context, store.WebhookEvent) (int64, error)
	ListWebhookEvents(context.Context, store.ListWebhookEventsOptions) ([]store.WebhookEvent, error)
}

type WebhookEventProcessor interface {
	Process(context.Context, store.WebhookEvent) error
}

type HealthStore interface {
	Ping(context.Context) error
}

type RouterConfig struct {
	WebhookSecret  string
	EventStore     WebhookEventStore
	EventProcessor WebhookEventProcessor
	HealthStore    HealthStore
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	handler := Handler{cfg: cfg}
	router.GET("/health", handler.health)

	v1 := router.Group("/api/v1")
	v1.GET("/dashboard/metrics", handler.dashboardMetrics)
	v1.GET("/events", handler.listWebhookEvents)
	v1.GET("/infra/edges", handler.infraEdges)
	v1.GET("/infra/nodes", handler.infraNodes)
	v1.GET("/infra/topology", handler.infraTopology)
	v1.GET("/ci-runs", handler.ciRuns)
	v1.GET("/ci-runs/:ci_run_id/logs", handler.ciRunLogs)
	v1.GET("/risks/critical", handler.criticalRisks)
	v1.POST("/integrations/gitea/webhooks", handler.receiveGiteaWebhook)

	return router
}

type Handler struct {
	cfg RouterConfig
}

func (h Handler) health(c *gin.Context) {
	dbStatus := "disabled"
	if h.cfg.HealthStore != nil {
		dbStatus = "ok"
		if err := h.cfg.HealthStore.Ping(c.Request.Context()); err != nil {
			dbStatus = "error"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "degraded",
				"service": "backend-core",
				"db":      dbStatus,
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "backend-core",
		"db":      dbStatus,
	})
}

func statusFromStoreError(err error) (int, string) {
	if errors.Is(err, store.ErrDuplicateEvent) {
		return http.StatusOK, "duplicate"
	}
	return http.StatusInternalServerError, "failed"
}
