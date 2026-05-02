package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devhub/backend-core/internal/gitea"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

const maxWebhookBodyBytes = 5 << 20

type giteaWebhookEnvelope struct {
	Repository *struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"repository"`
	Sender *struct {
		Login    string `json:"login"`
		UserName string `json:"username"`
	} `json:"sender"`
}

func (h Handler) receiveGiteaWebhook(c *gin.Context) {
	if h.cfg.EventStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "webhook store is not configured",
		})
		return
	}
	if h.cfg.WebhookSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "GITEA_WEBHOOK_SECRET is not configured",
		})
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBodyBytes))
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"status": "rejected",
			"error":  "webhook payload is too large",
		})
		return
	}

	signature := firstHeader(c, "X-Gitea-Signature", "X-Gogs-Signature")
	if !gitea.VerifySignature(body, h.cfg.WebhookSecret, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "rejected",
			"error":  "invalid webhook signature",
		})
		return
	}

	eventType := firstHeader(c, "X-Gitea-Event", "X-Gogs-Event")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "missing Gitea event type",
		})
		return
	}

	deliveryID := firstHeader(c, "X-Gitea-Delivery", "X-Gogs-Delivery")
	payloadHash := gitea.PayloadHash(body)
	dedupeKey := deliveryID
	if dedupeKey == "" {
		dedupeKey = fmt.Sprintf("%s:%s", eventType, payloadHash)
	}

	envelope, err := parseEnvelope(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "invalid JSON payload",
		})
		return
	}

	now := time.Now().UTC()
	record := store.WebhookEvent{
		EventType:      eventType,
		DeliveryID:     deliveryID,
		DedupeKey:      dedupeKey,
		RepositoryID:   envelope.repositoryID(),
		RepositoryName: envelope.repositoryName(),
		SenderLogin:    envelope.senderLogin(),
		Payload:        body,
		Status:         "validated",
		ReceivedAt:     now,
		ValidatedAt:    &now,
	}

	id, err := h.cfg.EventStore.SaveWebhookEvent(c.Request.Context(), record)
	if err != nil {
		statusCode, status := statusFromStoreError(err)
		response := gin.H{"status": status}
		if status == "failed" {
			response["error"] = err.Error()
		}
		c.JSON(statusCode, response)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":     "accepted",
		"event_id":   id,
		"event_type": eventType,
		"duplicate":  false,
	})
}

func parseEnvelope(body []byte) (giteaWebhookEnvelope, error) {
	var envelope giteaWebhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return giteaWebhookEnvelope{}, err
	}
	return envelope, nil
}

func (e giteaWebhookEnvelope) repositoryID() *int64 {
	if e.Repository == nil || e.Repository.ID == 0 {
		return nil
	}
	return &e.Repository.ID
}

func (e giteaWebhookEnvelope) repositoryName() string {
	if e.Repository == nil {
		return ""
	}
	if e.Repository.FullName != "" {
		return e.Repository.FullName
	}
	return e.Repository.Name
}

func (e giteaWebhookEnvelope) senderLogin() string {
	if e.Sender == nil {
		return ""
	}
	if e.Sender.Login != "" {
		return e.Sender.Login
	}
	return e.Sender.UserName
}

func firstHeader(c *gin.Context, names ...string) string {
	for _, name := range names {
		if value := c.GetHeader(name); value != "" {
			return value
		}
	}
	return ""
}
