package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// ExternalTaskStore defines the persistence contract for task item ingestion.
type ExternalTaskStore interface {
	UpsertExternalTaskItem(ctx context.Context, t domain.ExternalTaskItem) (domain.ExternalTaskItem, error)
	SoftDeleteExternalTaskItem(ctx context.Context, providerID, externalID string) error
	ListExternalTaskItems(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error)
	GetExternalTaskItemByID(ctx context.Context, id string) (domain.ExternalTaskItem, error)
	NextWebhookSeq(ctx context.Context) (int64, error)
	DetectWebhookSeqGaps(ctx context.Context, providerID string) (int64, error)
	UpdateProviderLastPulledAt(ctx context.Context, providerID string, pulledAt time.Time) error
	ListTaskTrackers(ctx context.Context) ([]domain.IntegrationProvider, error)
}

// webhookTaskEvent is the normalized webhook payload for task items.
type webhookTaskEvent struct {
	Event            string   `json:"event"`
	ExternalID       string   `json:"external_id"`
	Title            string   `json:"title"`
	RawStatus        string   `json:"raw_status"`
	NormalizedStatus string   `json:"normalized_status,omitempty"`
	Priority         string   `json:"priority,omitempty"`
	Assignee         string   `json:"assignee,omitempty"`
	Reporter         string   `json:"reporter,omitempty"`
	URL              string   `json:"url,omitempty"`
	Labels           []string `json:"labels,omitempty"`
	Description      string   `json:"description,omitempty"`
}

// externalTaskItemResponse builds the API response shape.
func externalTaskItemResponse(t domain.ExternalTaskItem) gin.H {
	resp := gin.H{
		"id":                t.ID,
		"provider_id":       t.ProviderID,
		"external_id":       t.ExternalID,
		"title":             t.Title,
		"raw_status":        t.RawStatus,
		"priority":          emptyAsNil(t.Priority),
		"assignee":          emptyAsNil(t.Assignee),
		"reporter":          emptyAsNil(t.Reporter),
		"url":               emptyAsNil(t.URL),
		"labels":            t.Labels,
		"webhook_seq":       t.WebhookSeq,
		"fetched_at":        t.FetchedAt.UTC().Format(time.RFC3339),
	}
	if t.NormalizedStatus != "" {
		resp["normalized_status"] = t.NormalizedStatus
	}
	return resp
}

// ReceiveExternalTaskWebhook handles POST /api/v1/integration/providers/:provider_id/tasks/webhook
func (h *IntegrationHandler) ReceiveExternalTaskWebhook(c *gin.Context) {
	providerID := c.Param("provider_id")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "provider_id is required"})
		return
	}

	storeI, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		return
	}

	taskStore, ok := h.ExternalTaskStoreOrUnavailable(c)
	if !ok {
		return
	}

	// 1. Provider lookup + type validation
	provider, err := storeI.GetIntegrationProviderByID(c.Request.Context(), providerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "provider_not_found"})
			return
		}
		httphelp.WriteServerError(c, err, "external_task.webhook.lookup")
		return
	}
	if provider.ProviderType != domain.IntegrationProviderTypeTaskTracker {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "provider_type_mismatch"})
		return
	}

	// 2. Webhook secret verification
	secret := c.GetHeader("X-Webhook-Secret")
	if provider.WebhookSecret != "" && secret == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "webhook_secret_mismatch"})
		return
	}
	if provider.WebhookSecret != "" && subtle.ConstantTimeCompare([]byte(secret), []byte(provider.WebhookSecret)) != 1 {
		_ = h.recordAuditBestEffort(c, "external_task.auth_failed", "external_task", "", map[string]any{
			"provider_id": providerID,
			"reason":      "webhook_secret_mismatch",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "webhook_secret_mismatch"})
		return
	}

	// 3. Bind webhook event
	var event webhookTaskEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "error": "invalid_webhook_payload"})
		return
	}
	if event.Event == "" || event.ExternalID == "" || event.Title == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "error": "invalid_webhook_payload"})
		return
	}
	if event.Event != "created" && event.Event != "updated" && event.Event != "deleted" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "error": "invalid_webhook_payload"})
		return
	}

	// 4. Acquire next webhook_seq (best-effort, failure allowed)
	seq, seqErr := taskStore.NextWebhookSeq(c.Request.Context())
	if seqErr != nil {
		seq = 0 // fallback — continue without seq
	}

	// 5. Process event
	now := time.Now().UTC()
	if event.Event == "deleted" {
		if err := taskStore.SoftDeleteExternalTaskItem(c.Request.Context(), providerID, event.ExternalID); err != nil && !errors.Is(err, store.ErrNotFound) {
			httphelp.WriteServerError(c, err, "external_task.webhook.delete")
			return
		}
		h.recordAuditBestEffort(c, "external_task.deleted", "external_task", "", map[string]any{
			"provider_id":  providerID,
			"external_id":  event.ExternalID,
			"webhook_seq":  seq,
		})
	} else {
		item := domain.ExternalTaskItem{
			ProviderID:       providerID,
			ExternalID:       event.ExternalID,
			Title:            event.Title,
			Description:      event.Description,
			RawStatus:        event.RawStatus,
			NormalizedStatus: event.NormalizedStatus,
			Priority:         event.Priority,
			Assignee:         event.Assignee,
			Reporter:         event.Reporter,
			URL:              event.URL,
			Labels:           event.Labels,
			WebhookSeq:       &seq,
			FetchedAt:        now,
		}
		_, err := taskStore.UpsertExternalTaskItem(c.Request.Context(), item)
		if err != nil {
			httphelp.WriteServerError(c, err, "external_task.webhook.upsert")
			return
		}
		auditAction := "external_task.received"
		if event.Event == "updated" {
			auditAction = "external_task.updated"
		}
		h.recordAuditBestEffort(c, auditAction, "external_task", "", map[string]any{
			"provider_id":  providerID,
			"external_id":  event.ExternalID,
			"webhook_seq":  seq,
		})
	}

	// 6. Response
	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
		"data": gin.H{
			"webhook_seq": seq,
			"external_id": event.ExternalID,
			"event":       event.Event,
			"provider_id": providerID,
		},
	})
}

// ListExternalTaskItems handles GET /api/v1/external-tasks
func (h *IntegrationHandler) ListExternalTaskItems(c *gin.Context) {
	taskStore, ok := h.ExternalTaskStoreOrUnavailable(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if offset > 0 {
		offset = (offset - 1) * limit
	}

	labelsStr := c.Query("labels")
	var labels []string
	if labelsStr != "" {
		labels = strings.Split(labelsStr, ",")
	}

	opts := store.ExternalTaskListOptions{
		ProviderID:       c.Query("provider_id"),
		RawStatus:        c.Query("raw_status"),
		NormalizedStatus: c.Query("normalized_status"),
		Assignee:         c.Query("assignee"),
		Labels:           labels,
		IncludeDeleted:   c.Query("include_deleted") == "true",
		Limit:            limit,
		Offset:           offset,
	}

	items, total, err := taskStore.ListExternalTaskItems(c.Request.Context(), opts)
	if err != nil {
		httphelp.WriteServerError(c, err, "external_task.list")
		return
	}

	data := make([]gin.H, 0, len(items))
	for _, item := range items {
		data = append(data, externalTaskItemResponse(item))
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"page":     page,
			"per_page": limit,
			"total":    total,
		},
	})
}

// GetExternalTaskItem handles GET /api/v1/external-tasks/:task_id
func (h *IntegrationHandler) GetExternalTaskItem(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "task_id is required"})
		return
	}

	taskStore, ok := h.ExternalTaskStoreOrUnavailable(c)
	if !ok {
		return
	}

	item, err := taskStore.GetExternalTaskItemByID(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "external_task_not_found"})
			return
		}
		httphelp.WriteServerError(c, err, "external_task.get")
		return
	}

	// TODO: scope binding check (ARCH-TASK-06) — compare actor's accessible providers

	resp := externalTaskItemResponse(item)
	resp["description"] = item.Description
	resp["raw_payload"] = item.RawPayload
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}

// ExternalTaskStoreOrUnavailable is a guard for ExternalTaskStore config.
func (h *IntegrationHandler) ExternalTaskStoreOrUnavailable(c *gin.Context) (ExternalTaskStore, bool) {
	if h.cfg.ExternalTaskStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "external task store is not configured",
		})
		return nil, false
	}
	return h.cfg.ExternalTaskStore, true
}
