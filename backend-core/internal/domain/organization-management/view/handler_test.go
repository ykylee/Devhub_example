package view

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type fakeOrgAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeOrgAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_org_id"
	f.created = append(f.created, log)
	return log, nil
}

type fakeHRDB struct{}

func (f *fakeHRDB) Lookup(_ context.Context, systemID, employeeID, name string) (string, string, string, error) {
	return systemID, employeeID, name, nil
}

func TestNewOrganizationHandler_NonNil(t *testing.T) {
	h := NewOrganizationHandler(OrganizationConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewOrganizationHandler_ConfigPropagation(t *testing.T) {
	cfg := OrganizationConfig{
		HRDB:                  &fakeHRDB{},
		AuditStore:            &fakeOrgAuditStore{},
		OnboardingGateEnabled: true,
	}
	h := NewOrganizationHandler(cfg)
	if h.cfg.HRDB == nil {
		t.Fatal("HRDB not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
	if !h.cfg.OnboardingGateEnabled {
		t.Fatal("OnboardingGateEnabled not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeOrgAuditStore{}
	h := NewOrganizationHandler(OrganizationConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "org.test", "unit", "u-1", nil)
	if got.AuditID != "audit_org_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "org.test" || c0.TargetType != "unit" || c0.TargetID != "u-1" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeOrgAuditStore{}
	h := NewOrganizationHandler(OrganizationConfig{AuditStore: store})
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

	store := &fakeOrgAuditStore{err: errors.New("db_down")}
	h := NewOrganizationHandler(OrganizationConfig{AuditStore: store})
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
		h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if !h.RequireOnboardingFlag(c) {
			t.Fatal("expected true")
		}
		if rec.Code == 404 {
			t.Fatal("must not write 404 when enabled")
		}
	})

	t.Run("disabled returns false + 404 body", func(t *testing.T) {
		h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: false})
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

func TestOrgUnitFromDomain_FullMapping(t *testing.T) {
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(24 * time.Hour)
	unit := domain.OrgUnit{
		ID:           1,
		UnitID:       "team-a",
		ParentUnitID: "root",
		UnitType:     domain.UnitType("team"),
		Label:        "Team A",
		LeaderUserID: "leader-1",
		PositionX:    100,
		PositionY:    200,
		DirectCount:  5,
		TotalCount:   10,
		CreatedAt:    created,
		UpdatedAt:    updated,
	}
	got := orgUnitFromDomain(unit)
	if got.UnitID != "team-a" || got.ParentUnitID != "root" || got.Label != "Team A" {
		t.Fatalf("basic: %+v", got)
	}
	if got.UnitType != "team" {
		t.Fatalf("unit_type = %q", got.UnitType)
	}
	if got.PositionX != 100 || got.PositionY != 200 {
		t.Fatalf("position: %+v", got)
	}
	if got.DirectCount != 5 || got.TotalCount != 10 {
		t.Fatalf("counts: %+v", got)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("ts: %+v", got)
	}
}

func TestParseJoinedAt_EmptyReturnsZero(t *testing.T) {
	got, err := parseJoinedAt("")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !got.IsZero() {
		t.Fatalf("expected zero, got %v", got)
	}
}

func TestParseJoinedAt_RFC3339(t *testing.T) {
	got, err := parseJoinedAt("2026-05-29T10:00:00Z")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.Year() != 2026 || got.Hour() != 10 {
		t.Fatalf("got %v", got)
	}
}

func TestParseJoinedAt_DateOnly(t *testing.T) {
	got, err := parseJoinedAt("2026-05-29")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.Year() != 2026 || got.Day() != 29 {
		t.Fatalf("got %v", got)
	}
}

func TestParseJoinedAt_InvalidReturnsErr(t *testing.T) {
	_, err := parseJoinedAt("not-a-date")
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestWriteStoreError_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	writeStoreError(c, store.ErrNotFound, "op")
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "not_found") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestWriteStoreError_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	writeStoreError(c, store.ErrConflict, "op")
	if rec.Code != 409 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "conflict") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestWriteStoreError_DefaultServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	writeStoreError(c, errors.New("boom"), "op")
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOrgAddAuditMeta_WithIDStampsMeta(t *testing.T) {
	resp := gin.H{}
	addAuditMeta(resp, domain.AuditLog{AuditID: "audit_x"})
	meta, _ := resp["meta"].(gin.H)
	if meta == nil || meta["audit_log_id"] != "audit_x" {
		t.Fatalf("meta = %+v", resp["meta"])
	}
}

func TestOrgAddAuditMeta_PreservesExistingMeta(t *testing.T) {
	resp := gin.H{"meta": gin.H{"x": 1}}
	addAuditMeta(resp, domain.AuditLog{AuditID: "audit_x"})
	meta, _ := resp["meta"].(gin.H)
	if meta["x"] != 1 {
		t.Fatalf("existing meta key lost: %+v", meta)
	}
	if meta["audit_log_id"] != "audit_x" {
		t.Fatalf("audit_log_id missing: %+v", meta)
	}
}

func TestOrgAddAuditMeta_EmptyIDNoop(t *testing.T) {
	resp := gin.H{}
	addAuditMeta(resp, domain.AuditLog{})
	if _, ok := resp["meta"]; ok {
		t.Fatal("must not add meta when id empty")
	}
}

func TestUserUpdateAuditPayload_AllFields(t *testing.T) {
	email := "alice@example.com"
	display := "Alice"
	role := domain.AppRoleDeveloper
	status := domain.UserStatusActive
	primary := "team-a"
	current := "team-b"
	seconded := true
	joined := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	input := domain.UpdateUserInput{
		Email:         &email,
		DisplayName:   &display,
		Role:          &role,
		Status:        &status,
		PrimaryUnitID: &primary,
		CurrentUnitID: &current,
		IsSeconded:    &seconded,
		JoinedAt:      &joined,
	}
	got := userUpdateAuditPayload(input)
	if got["email"] != email || got["display_name"] != display {
		t.Fatalf("email/display: %+v", got)
	}
	if got["role"] != string(role) || got["status"] != string(status) {
		t.Fatalf("role/status: %+v", got)
	}
	if got["primary_unit_id"] != primary || got["current_unit_id"] != current {
		t.Fatalf("unit ids: %+v", got)
	}
	if got["is_seconded"] != seconded {
		t.Fatalf("is_seconded: %+v", got)
	}
	if _, ok := got["joined_at"].(string); !ok {
		t.Fatalf("joined_at type: %T", got["joined_at"])
	}
}

func TestUserUpdateAuditPayload_NilFieldsOmitted(t *testing.T) {
	input := domain.UpdateUserInput{}
	got := userUpdateAuditPayload(input)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}

func TestOrgUnitUpdateAuditPayload_Fields(t *testing.T) {
	parent := "root"
	unitType := domain.UnitType("team")
	label := "Team A"
	leader := "leader-1"
	x := 100
	y := 200
	input := domain.UpdateOrgUnitInput{
		ParentUnitID: &parent,
		UnitType:     &unitType,
		Label:        &label,
		LeaderUserID: &leader,
		PositionX:    &x,
		PositionY:    &y,
	}
	got := orgUnitUpdateAuditPayload(input)
	if got["parent_unit_id"] != parent || got["unit_type"] != string(unitType) {
		t.Fatalf("basic: %+v", got)
	}
	if got["label"] != label || got["leader_user_id"] != leader {
		t.Fatalf("label/leader: %+v", got)
	}
	if got["position_x"] != x || got["position_y"] != y {
		t.Fatalf("position: %+v", got)
	}
}

func TestOrgUnitUpdateAuditPayload_NilFieldsOmitted(t *testing.T) {
	got := orgUnitUpdateAuditPayload(domain.UpdateOrgUnitInput{})
	if len(got) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}
