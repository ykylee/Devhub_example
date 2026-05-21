package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// RM-ONBOARD-04 (Carve D) — Carve A 의 onboarding_test.go 가 happy path + 핵심
// edge 13건을 cover. 본 파일은 잔여 edge case 8건 (PATCH /me, ConfirmUserReview,
// SubmitOnboarding role-ignore, SearchOrgs limit) 을 추가한다.

// TestPatchMe_HappyPath — display_name 만 변경 시 review_status 영향 없음
// (REQ-FR-ONBOARD-007 negative — primary_unit_id 변경이 아니면 reset 안 함).
func TestPatchMe_HappyPath_DisplayNameOnly(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := store.ConfirmUserReview(context.Background(), "alice"); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", bytes.NewReader([]byte(`{"display_name":"Alice K."}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	user, _ := store.GetUser(context.Background(), "alice")
	if user.DisplayName != "Alice K." {
		t.Errorf("expected display_name updated, got %q", user.DisplayName)
	}
	if user.ReviewStatus != domain.ReviewStatusReviewed {
		t.Errorf("expected review_status unchanged (reviewed), got %q", user.ReviewStatus)
	}
}

// TestPatchMe_UnitNotFound404 — primary_unit_id 가 organization_units 에 없음.
func TestPatchMe_UnitNotFound404(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", bytes.NewReader([]byte(`{"primary_unit_id":"team-ghost"}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestPatchMe_InvalidPayload422 — display_name, primary_unit_id 둘 다 미포함.
func TestPatchMe_InvalidPayload422_NoFields(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestConfirmUserReview_UserNotFound404 — 존재하지 않는 user.
func TestConfirmUserReview_UserNotFound404(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	// admin 만 seed.
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "admin1", DisplayName: "Admin", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleSystemAdmin,
	}); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "admin1", Subject: "sub-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/ghost/review", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestConfirmUserReview_AlreadyConfirmed409 — review_status='reviewed' 인 사용자
// 중복 confirm.
func TestConfirmUserReview_AlreadyConfirmed409(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	if _, err := store.ConfirmUserReview(context.Background(), "alice"); err != nil {
		t.Fatalf("review alice: %v", err)
	}
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "admin1", DisplayName: "Admin", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleSystemAdmin,
	}); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "admin1", Subject: "sub-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/alice/review", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestConfirmUserReview_OnboardingNotCompleted422 — onboarding_completed_at IS
// NULL 사용자가 검토 대상 아님 (admin pre-seed 사용자가 제출 전).
func TestConfirmUserReview_OnboardingNotCompleted422(t *testing.T) {
	store := seedOnboardingFixture(t, true) // pre-seed alice (incomplete)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "admin1", DisplayName: "Admin", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleSystemAdmin,
	}); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "admin1", Subject: "sub-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/preseeded/review", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestSubmitOnboarding_RoleIgnored — payload 의 role 필드는 무시 (REQ-FR-ONBOARD-002).
// fallback role 또는 Keycloak claim 만 사용.
func TestSubmitOnboarding_RoleIgnored(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	body := []byte(`{"display_name":"Alice","primary_unit_id":"team-platform","role":"system_admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	user, _ := store.GetUser(context.Background(), "alice")
	if user.Role == domain.AppRoleSystemAdmin {
		t.Errorf("payload role must be ignored, but got %q", user.Role)
	}
}

// TestOrganizationsSearch_LimitOverMax422 — limit > 20 → 422 invalid_query_params.
func TestOrganizationsSearch_LimitOverMax422(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/search?q=team&limit=21", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}
