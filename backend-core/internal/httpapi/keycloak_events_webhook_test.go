package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	auditview "github.com/devhub/backend-core/internal/domain/audit-ops/view"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// keycloakWebhookRequest — auditview 의 cross-package test 접근용 alias.
// (re-exported via auditview.KeycloakWebhookRequest 가 type alias 이므로 같은 타입.)
type keycloakWebhookRequest = auditview.KeycloakWebhookRequest

type mockAuditStore struct {
	logs      []domain.AuditLog
	returnErr error
}

func (m *mockAuditStore) CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if m.returnErr != nil {
		return domain.AuditLog{}, m.returnErr
	}
	m.logs = append(m.logs, log)
	return log, nil
}

func (m *mockAuditStore) ListAuditLogs(ctx context.Context, opts store.ListAuditLogsOptions) ([]domain.AuditLog, error) {
	return m.logs, nil
}

// TestReceiveKeycloakEventWebhook_SecretNotConfigured covers PR #203 codex P1
// hotfix (sprint -d): the handler must fail-closed (503) when
// DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET is unset rather than disabling auth.
func TestReceiveKeycloakEventWebhook_SecretNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	store := &mockAuditStore{}
	handler := Handler{
		cfg: RouterConfig{
			KeycloakWebhookSecret: "", // not configured
			AuditStore:            store,
		},
	}

	r.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	reqBody := `{"id":"test-event-no-secret"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/internal/keycloak-events", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Even without any X-Webhook-Secret header the response must be 503
	// (handler fails closed rather than silently accepting).

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 (secret not configured fail-closed), got %d", w.Code)
	}
	if len(store.logs) != 0 {
		t.Errorf("Expected no audit logs emitted when secret unset, got %d", len(store.logs))
	}
}

func TestReceiveKeycloakEventWebhook_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	store := &mockAuditStore{}
	handler := Handler{
		cfg: RouterConfig{
			KeycloakWebhookSecret: "super-secret-token",
			AuditStore:            store,
		},
	}

	r.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	reqBody := `{"id":"test-event-1"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/internal/keycloak-events", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Miss or wrong secret
	c.Request.Header.Set("X-Webhook-Secret", "wrong-secret")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestReceiveKeycloakEventWebhook_UserEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	store := &mockAuditStore{}
	handler := Handler{
		cfg: RouterConfig{
			KeycloakWebhookSecret: "super-secret-token",
			AuditStore:            store,
		},
	}

	r.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	event := keycloakWebhookRequest{
		ID:        "event-u-123",
		Time:      1716120000000,
		Type:      "LOGIN",
		RealmID:   "devhub",
		ClientID:  "devhub-frontend",
		UserID:    "alice",
		IPAddress: "127.0.0.1",
		Details:   map[string]string{"sessionId": "sess-999"},
	}
	body, _ := json.Marshal(event)

	c.Request, _ = http.NewRequest("POST", "/api/v1/internal/keycloak-events", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Webhook-Secret", "super-secret-token")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	if len(store.logs) != 1 {
		t.Fatalf("Expected 1 audit log created, got %d", len(store.logs))
	}

	logEntry := store.logs[0]
	if logEntry.ActorLogin != "alice" {
		t.Errorf("Expected actor 'alice', got %q", logEntry.ActorLogin)
	}
	if logEntry.Action != "auth.login.success" {
		t.Errorf("Expected action 'auth.login.success', got %q", logEntry.Action)
	}
	if logEntry.TargetType != "auth" {
		t.Errorf("Expected targetType 'auth', got %q", logEntry.TargetType)
	}
	if logEntry.TargetID != "alice" {
		t.Errorf("Expected targetID 'alice', got %q", logEntry.TargetID)
	}
}

func TestReceiveKeycloakEventWebhook_AdminEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	store := &mockAuditStore{}
	handler := Handler{
		cfg: RouterConfig{
			KeycloakWebhookSecret: "super-secret-token",
			AuditStore:            store,
		},
	}

	r.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	event := keycloakWebhookRequest{
		ID:            "event-a-456",
		Time:          1716120000000,
		OperationType: "CREATE",
		ResourceType:  "USER",
		ResourcePath:  "users/bob-id-99",
		RealmID:       "devhub",
		Error:         "",
	}
	event.AuthDetails.UserID = "charlie"
	event.AuthDetails.ClientID = "devhub-backend"
	event.AuthDetails.IPAddress = "10.0.0.5"

	body, _ := json.Marshal(event)

	c.Request, _ = http.NewRequest("POST", "/api/v1/internal/keycloak-events", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Webhook-Secret", "super-secret-token")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	if len(store.logs) != 1 {
		t.Fatalf("Expected 1 audit log created, got %d", len(store.logs))
	}

	logEntry := store.logs[0]
	if logEntry.ActorLogin != "charlie" {
		t.Errorf("Expected actor 'charlie', got %q", logEntry.ActorLogin)
	}
	if logEntry.Action != "keycloak.user.created" {
		t.Errorf("Expected action 'keycloak.user.created', got %q", logEntry.Action)
	}
	if logEntry.TargetType != "user" {
		t.Errorf("Expected targetType 'user', got %q", logEntry.TargetType)
	}
	if logEntry.TargetID != "users/bob-id-99" {
		t.Errorf("Expected targetID 'users/bob-id-99', got %q", logEntry.TargetID)
	}
}

func TestReceiveKeycloakEventWebhook_DuplicateIgnore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	// Mock DB conflict return
	store := &mockAuditStore{
		returnErr: errors.New("duplicate key value violates unique constraint"),
	}
	handler := Handler{
		cfg: RouterConfig{
			KeycloakWebhookSecret: "super-secret-token",
			AuditStore:            store,
		},
	}

	r.POST("/api/v1/internal/keycloak-events", handler.receiveKeycloakEventWebhook)

	event := keycloakWebhookRequest{
		ID:        "event-u-123",
		Time:      1716120000000,
		Type:      "LOGIN",
		RealmID:   "devhub",
		ClientID:  "devhub-frontend",
		UserID:    "alice",
		IPAddress: "127.0.0.1",
		Details:   map[string]string{"sessionId": "sess-999"},
	}
	body, _ := json.Marshal(event)

	c.Request, _ = http.NewRequest("POST", "/api/v1/internal/keycloak-events", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Webhook-Secret", "super-secret-token")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on duplicate ignore, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ignored" || resp["reason"] != "duplicate" {
		t.Errorf("Expected response status ignored and reason duplicate, got %+v", resp)
	}
}
