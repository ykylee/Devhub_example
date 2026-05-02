package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type memoryEventStore struct {
	events  []store.WebhookEvent
	err     error
	listErr error
}

func (s *memoryEventStore) SaveWebhookEvent(_ context.Context, event store.WebhookEvent) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.events = append(s.events, event)
	return int64(len(s.events)), nil
}

func (s *memoryEventStore) ListWebhookEvents(_ context.Context, opts store.ListWebhookEventsOptions) ([]store.WebhookEvent, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	limit := opts.Limit
	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}
	if opts.Offset >= len(s.events) {
		return []store.WebhookEvent{}, nil
	}
	end := opts.Offset + limit
	if end > len(s.events) {
		end = len(s.events)
	}
	return s.events[opts.Offset:end], nil
}

func TestReceiveGiteaWebhookStoresValidatedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "devhub-secret"
	body := []byte(`{"repository":{"id":42,"full_name":"acme/api"},"sender":{"login":"yklee"}}`)
	eventStore := &memoryEventStore{}
	router := NewRouter(RouterConfig{
		WebhookSecret: secret,
		EventStore:    eventStore,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/gitea/webhooks", bytes.NewReader(body))
	req.Header.Set("X-Gitea-Signature", signBody(body, secret))
	req.Header.Set("X-Gitea-Event", "pull_request")
	req.Header.Set("X-Gitea-Delivery", "delivery-1")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d: %s", http.StatusAccepted, rec.Code, rec.Body.String())
	}
	if len(eventStore.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(eventStore.events))
	}
	event := eventStore.events[0]
	if event.EventType != "pull_request" {
		t.Fatalf("expected event type pull_request, got %q", event.EventType)
	}
	if event.DedupeKey != "delivery-1" {
		t.Fatalf("expected delivery id dedupe key, got %q", event.DedupeKey)
	}
	if event.RepositoryName != "acme/api" {
		t.Fatalf("expected repository name, got %q", event.RepositoryName)
	}
	if event.SenderLogin != "yklee" {
		t.Fatalf("expected sender login, got %q", event.SenderLogin)
	}
	if event.ValidatedAt == nil {
		t.Fatal("expected validated_at to be set")
	}
}

func TestReceiveGiteaWebhookRejectsInvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eventStore := &memoryEventStore{}
	router := NewRouter(RouterConfig{
		WebhookSecret: "devhub-secret",
		EventStore:    eventStore,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/gitea/webhooks", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Gitea-Signature", "bad")
	req.Header.Set("X-Gitea-Event", "push")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if len(eventStore.events) != 0 {
		t.Fatalf("expected no stored events, got %d", len(eventStore.events))
	}
}

func TestListWebhookEventsReturnsStableEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	repositoryID := int64(42)
	eventStore := &memoryEventStore{
		events: []store.WebhookEvent{
			{
				ID:             7,
				EventType:      "push",
				DeliveryID:     "delivery-7",
				DedupeKey:      "delivery-7",
				RepositoryID:   &repositoryID,
				RepositoryName: "acme/api",
				SenderLogin:    "yklee",
				Payload:        []byte(`{"ref":"refs/heads/main"}`),
				Status:         "validated",
				ReceivedAt:     now,
				ValidatedAt:    &now,
			},
		},
	}
	router := NewRouter(RouterConfig{EventStore: eventStore})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Status string `json:"status"`
		Data   []struct {
			ID             int64           `json:"id"`
			EventType      string          `json:"event_type"`
			DeliveryID     string          `json:"delivery_id"`
			RepositoryName string          `json:"repository_name"`
			Payload        json.RawMessage `json:"payload"`
			Status         string          `json:"status"`
		} `json:"data"`
		Meta struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Count  int `json:"count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("expected ok status, got %q", response.Status)
	}
	if len(response.Data) != 1 {
		t.Fatalf("expected 1 event, got %d", len(response.Data))
	}
	if response.Data[0].EventType != "push" || response.Data[0].RepositoryName != "acme/api" {
		t.Fatalf("unexpected event response: %+v", response.Data[0])
	}
	if response.Meta.Limit != 10 || response.Meta.Offset != 0 || response.Meta.Count != 1 {
		t.Fatalf("unexpected meta: %+v", response.Meta)
	}
}

func TestListWebhookEventsRejectsBadLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{EventStore: &memoryEventStore{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?limit=101", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestReceiveGiteaWebhookHandlesDuplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "devhub-secret"
	body := []byte(`{"repository":{"name":"api"}}`)
	eventStore := &memoryEventStore{err: store.ErrDuplicateEvent}
	router := NewRouter(RouterConfig{
		WebhookSecret: secret,
		EventStore:    eventStore,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/gitea/webhooks", bytes.NewReader(body))
	req.Header.Set("X-Gitea-Signature", signBody(body, secret))
	req.Header.Set("X-Gitea-Event", "push")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["status"] != "duplicate" {
		t.Fatalf("expected duplicate status, got %q", response["status"])
	}
}

func signBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
