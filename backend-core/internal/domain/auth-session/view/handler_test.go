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

type fakeAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_test_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewAuthHandler_ConfigPropagation(t *testing.T) {
	cfg := AuthConfig{OnboardingGateEnabled: true, AuthDevFallback: true}
	h := NewAuthHandler(cfg)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if !h.cfg.OnboardingGateEnabled || !h.cfg.AuthDevFallback {
		t.Fatal("config not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(AuthConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "action", "target_type", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero AuditLog, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAuditStore{}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "test.action", "user", "u-1", nil)
	if got.AuditID != "audit_test_id" {
		t.Fatalf("expected audit id stamped, got %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(store.created))
	}
	created := store.created[0]
	if created.ActorLogin != "alice" {
		t.Fatalf("actor_login = %q", created.ActorLogin)
	}
	if created.Action != "test.action" || created.TargetType != "user" || created.TargetID != "u-1" {
		t.Fatalf("action mapping = %+v", created)
	}
	src, _ := created.Payload["actor_source"].(string)
	if src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAuditStore{}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
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

	store := &fakeAuditStore{err: errors.New("db_down")}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
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
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true})
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
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: false})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if h.RequireOnboardingFlag(c) {
			t.Fatal("expected false")
		}
		if rec.Code != 404 {
			t.Fatalf("got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "onboarding_feature_disabled") {
			t.Fatalf("body=%q", rec.Body.String())
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
		t.Fatalf("basic field mapping wrong: %+v", got)
	}
	if got.Role != string(domain.AppRoleDeveloper) {
		t.Fatalf("role mapping wrong: %q", got.Role)
	}
	if got.PrimaryUnitID != "team-a" || got.CurrentUnitID != "team-b" {
		t.Fatalf("unit mapping wrong: %+v", got)
	}
	if !got.IsSeconded {
		t.Fatal("seconded flag")
	}
	if len(got.Appointments) != 2 {
		t.Fatalf("appointments count = %d", len(got.Appointments))
	}
	if got.Appointments[0].UnitID != "team-a" || got.Appointments[0].AppointmentRole != string(domain.AppointmentRoleLeader) {
		t.Fatalf("appt 0: %+v", got.Appointments[0])
	}
	if got.ReviewStatus != "approved" {
		t.Fatalf("review status: %q", got.ReviewStatus)
	}
}

func TestAppUserFromDomain_EmptyAppointments(t *testing.T) {
	user := domain.AppUser{UserID: "u-1"}
	got := appUserFromDomain(user)
	if got.Appointments == nil {
		t.Fatal("Appointments must not be nil (slice allocation)")
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
