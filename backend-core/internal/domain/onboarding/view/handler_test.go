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

type fakeOnboardingAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeOnboardingAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_ob_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewOnboardingHandler_NonNil(t *testing.T) {
	h := NewOnboardingHandler(OnboardingConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewOnboardingHandler_ConfigPropagation(t *testing.T) {
	cfg := OnboardingConfig{
		OnboardingGateEnabled: true,
		AuditStore:            &fakeOnboardingAuditStore{},
	}
	h := NewOnboardingHandler(cfg)
	if !h.cfg.OnboardingGateEnabled {
		t.Fatal("OnboardingGateEnabled not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOnboardingHandler(OnboardingConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeOnboardingAuditStore{}
	h := NewOnboardingHandler(OnboardingConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "ob.test", "user", "u-1", nil)
	if got.AuditID != "audit_ob_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "ob.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeOnboardingAuditStore{}
	h := NewOnboardingHandler(OnboardingConfig{AuditStore: store})
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

	store := &fakeOnboardingAuditStore{err: errors.New("db_down")}
	h := NewOnboardingHandler(OnboardingConfig{AuditStore: store})
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

func TestRequireOnboardingFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("enabled returns true and no body", func(t *testing.T) {
		h := NewOnboardingHandler(OnboardingConfig{OnboardingGateEnabled: true})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if !h.RequireOnboardingFlag(c) {
			t.Fatal("expected true")
		}
		if rec.Code == 404 {
			t.Fatal("must not write 404")
		}
	})

	t.Run("disabled returns false + 404 body", func(t *testing.T) {
		h := NewOnboardingHandler(OnboardingConfig{OnboardingGateEnabled: false})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if h.RequireOnboardingFlag(c) {
			t.Fatal("expected false")
		}
		if rec.Code != 404 {
			t.Fatalf("status = %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "onboarding_feature_disabled") {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})
}

func TestAppUserFromDomain_FullMapping(t *testing.T) {
	user := domain.AppUser{
		ID:            42,
		UserID:        "u-1",
		Email:         "alice@example.com",
		DisplayName:   "Alice",
		Role:          domain.AppRoleDeveloper,
		Status:        domain.UserStatusActive,
		PrimaryUnitID: "team-a",
		CurrentUnitID: "team-b",
		IsSeconded:    true,
		Appointments: []domain.UnitAppointment{
			{UnitID: "team-a", AppointmentRole: domain.AppointmentRoleLeader},
			{UnitID: "team-b", AppointmentRole: domain.AppointmentRoleMember},
		},
		ReviewStatus: "approved",
	}

	got := appUserFromDomain(user)
	if got.UserID != "u-1" || got.Email != "alice@example.com" || got.DisplayName != "Alice" {
		t.Fatalf("basic mapping: %+v", got)
	}
	if got.Role != string(domain.AppRoleDeveloper) {
		t.Fatalf("role: %q", got.Role)
	}
	if got.PrimaryUnitID != "team-a" || got.CurrentUnitID != "team-b" {
		t.Fatalf("unit: %+v", got)
	}
	if !got.IsSeconded {
		t.Fatal("seconded flag")
	}
	if len(got.Appointments) != 2 {
		t.Fatalf("appts = %d", len(got.Appointments))
	}
	if got.Appointments[0].UnitID != "team-a" || got.Appointments[0].AppointmentRole != string(domain.AppointmentRoleLeader) {
		t.Fatalf("appt 0: %+v", got.Appointments[0])
	}
	if got.ReviewStatus != "approved" {
		t.Fatalf("review_status: %q", got.ReviewStatus)
	}
}

func TestAppUserFromDomain_EmptyAppointments(t *testing.T) {
	user := domain.AppUser{UserID: "u-1"}
	got := appUserFromDomain(user)
	if got.Appointments == nil {
		t.Fatal("Appointments must not be nil")
	}
	if len(got.Appointments) != 0 {
		t.Fatalf("len = %d", len(got.Appointments))
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
