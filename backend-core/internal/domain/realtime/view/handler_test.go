package view

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)

type fakeRealtimeAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeRealtimeAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_rt_id"
	f.created = append(f.created, log)
	return log, nil
}

type fakeRealtimePermCache struct{}

func (f *fakeRealtimePermCache) Allows(_ context.Context, _ string, _ domain.Resource, _ domain.Action) (bool, error) {
	return true, nil
}

func TestNewRealtimeHandler_NonNil(t *testing.T) {
	h := NewRealtimeHandler(RealtimeConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewRealtimeHandler_ConfigPropagation(t *testing.T) {
	cfg := RealtimeConfig{
		PermissionCache: &fakeRealtimePermCache{},
		AuditStore:      &fakeRealtimeAuditStore{},
		AuthDevFallback: true,
	}
	h := NewRealtimeHandler(cfg)
	if h.cfg.PermissionCache == nil {
		t.Fatal("PermissionCache not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
	if !h.cfg.AuthDevFallback {
		t.Fatal("AuthDevFallback not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRealtimeHandler(RealtimeConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRealtimeAuditStore{}
	h := NewRealtimeHandler(RealtimeConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "rt.test", "channel", "ch-1", nil)
	if got.AuditID != "audit_rt_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "rt.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRealtimeAuditStore{}
	h := NewRealtimeHandler(RealtimeConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)

	h.recordAuditBestEffort(c, "a", "t", "id", map[string]any{"existing": "value"})
	c0 := store.created[0]
	if c0.Payload["existing"] != "value" {
		t.Fatalf("existing lost: %+v", c0.Payload)
	}
	if _, ok := c0.Payload["actor_source"]; !ok {
		t.Fatal("actor_source must be augmented")
	}
}

func TestRecordAuditBestEffort_PersistFailureLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logBuf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(orig) })

	store := &fakeRealtimeAuditStore{err: errors.New("db_down")}
	h := NewRealtimeHandler(RealtimeConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set(httphelp.CtxKeyRequestID, "req_x")

	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatal("expected zero audit on err")
	}
	if !strings.Contains(logBuf.String(), "audit log persistence failed") {
		t.Fatalf("expected log, got %q", logBuf.String())
	}
}
