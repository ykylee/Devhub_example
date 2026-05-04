package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type eventResponse struct {
	ID             int64           `json:"id"`
	EventType      string          `json:"event_type"`
	DeliveryID     string          `json:"delivery_id,omitempty"`
	DedupeKey      string          `json:"dedupe_key"`
	RepositoryID   *int64          `json:"repository_id,omitempty"`
	RepositoryName string          `json:"repository_name,omitempty"`
	SenderLogin    string          `json:"sender_login,omitempty"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"`
	ReceivedAt     time.Time       `json:"received_at"`
	ValidatedAt    *time.Time      `json:"validated_at,omitempty"`
}

func (h Handler) listWebhookEvents(c *gin.Context) {
	if h.cfg.EventStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "event store is not configured",
		})
		return
	}

	limit, err := parseBoundedInt(c.DefaultQuery("limit", "50"), 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "limit must be an integer between 1 and 100",
		})
		return
	}
	offset, err := parseBoundedInt(c.DefaultQuery("offset", "0"), 0, 100000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "offset must be a non-negative integer",
		})
		return
	}

	events, err := h.cfg.EventStore.ListWebhookEvents(c.Request.Context(), store.ListWebhookEventsOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	data := make([]eventResponse, 0, len(events))
	for _, event := range events {
		data = append(data, eventResponse{
			ID:             event.ID,
			EventType:      event.EventType,
			DeliveryID:     event.DeliveryID,
			DedupeKey:      event.DedupeKey,
			RepositoryID:   event.RepositoryID,
			RepositoryName: event.RepositoryName,
			SenderLogin:    event.SenderLogin,
			Payload:        json.RawMessage(event.Payload),
			Status:         event.Status,
			ReceivedAt:     event.ReceivedAt,
			ValidatedAt:    event.ValidatedAt,
		})
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

func parseBoundedInt(value string, minValue, maxValue int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < minValue || parsed > maxValue {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}
