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

type fakeRBACAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeRBACAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_rbac_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewRBACHandler_NonNil(t *testing.T) {
	h := NewRBACHandler(RBACConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewRBACHandler_ConfigPropagation(t *testing.T) {
	cfg := RBACConfig{AuthDevFallback: true, AuditStore: &fakeRBACAuditStore{}}
	h := NewRBACHandler(cfg)
	if !h.cfg.AuthDevFallback {
		t.Fatal("AuthDevFallback not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRBACHandler(RBACConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "action", "type", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRBACAuditStore{}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "rbac.test", "role", "role-1", nil)
	if got.AuditID != "audit_rbac_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "rbac.test" || c0.TargetType != "role" || c0.TargetID != "role-1" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRBACAuditStore{}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)

	h.recordAuditBestEffort(c, "a", "t", "id", map[string]any{"existing": "value"})
	created := store.created[0]
	if created.Payload["existing"] != "value" {
		t.Fatalf("existing payload key lost: %+v", created.Payload)
	}
	if _, ok := created.Payload["actor_source"]; !ok {
		t.Fatal("actor_source must be augmented")
	}
}

func TestRecordAuditBestEffort_PersistFailureLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logBuf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(orig) })

	store := &fakeRBACAuditStore{err: errors.New("db_down")}
	h := NewRBACHandler(RBACConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set(httphelp.CtxKeyRequestID, "req_x")

	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatal("expected zero audit on err")
	}
	if !strings.Contains(logBuf.String(), "audit log persistence failed") {
		t.Fatalf("expected log msg, got %q", logBuf.String())
	}
}

func TestAddAuditMeta(t *testing.T) {
	t.Run("with audit id stamps key", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{AuditID: "audit_x"})
		if resp["audit_log_id"] != "audit_x" {
			t.Fatalf("got %v", resp["audit_log_id"])
		}
	})
	t.Run("without audit id no-op", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{})
		if _, ok := resp["audit_log_id"]; ok {
			t.Fatal("must not stamp empty id")
		}
	})
}
