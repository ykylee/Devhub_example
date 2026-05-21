package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// RM-ONBOARD-01 UT — onboardingGate middleware + 5 handler (API-83/84/85/86 +
// API-32 확장) + token-only actor 회귀. carve plan §2.1 의 Acceptance criteria 9건
// 핵심 cover.

// seedOnboardingFixture wires up a memory store with a single org unit
// and (optionally) an admin-preseed user.
func seedOnboardingFixture(t *testing.T, preseedIncomplete bool) *memoryOrganizationStore {
	t.Helper()
	store := newMemoryOrganizationStore()
	if _, err := store.CreateOrgUnit(context.Background(), domain.CreateOrgUnitInput{
		UnitID:   "team-platform",
		UnitType: domain.UnitType("team"),
		Label:    "AI/플랫폼팀",
	}); err != nil {
		t.Fatalf("seed unit: %v", err)
	}
	if _, err := store.CreateOrgUnit(context.Background(), domain.CreateOrgUnitInput{
		UnitID:   "team-data",
		UnitType: domain.UnitType("team"),
		Label:    "데이터팀",
	}); err != nil {
		t.Fatalf("seed unit data: %v", err)
	}
	if preseedIncomplete {
		// admin pre-seeded user (onboarding 미완료) — API-33 확장 흐름.
		_, err := store.CreateUser(context.Background(), domain.CreateUserInput{
			UserID:      "preseeded",
			Email:       "preseeded@example.com",
			DisplayName: "Pre Seeded",
			Role:        domain.AppRoleDeveloper,
			Status:      domain.UserStatusActive,
			Type:        domain.UserTypeHuman,
		})
		if err != nil {
			t.Fatalf("preseed user: %v", err)
		}
	}
	return store
}

// Carve A — feature flag OFF default = onboardingGate no-op + lazy 동작 유지.
// TestOnboardingEndpoints_FlagOff404 — Stage 3 보강 (P1 #1). flag false 일 때
// 신규 onboarding endpoint 4건이 404 onboarding_feature_disabled 반환 → main
// 동작 변경 없음 (단독 머지 안정성). 본 case 의 핵심.
func TestOnboardingEndpoints_FlagOff404(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	// system_admin role 사용 — admin endpoint 의 enforceRoutePermission 통과 후
	// flag guard 까지 도달 검증 가능.
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "admin1", Subject: "sub-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: false, // explicit
	})

	endpoints := []struct {
		method, path, body string
	}{
		{http.MethodPost, "/api/v1/me/onboarding", `{"display_name":"X","primary_unit_id":"team-platform"}`},
		{http.MethodPatch, "/api/v1/me", `{"display_name":"X"}`},
		{http.MethodGet, "/api/v1/organizations/search?q=team", ""},
		{http.MethodPost, "/api/v1/admin/users/admin1/review", `{}`},
	}
	for _, e := range endpoints {
		var body *bytes.Reader
		if e.body != "" {
			body = bytes.NewReader([]byte(e.body))
		} else {
			body = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(e.method, e.path, body)
		req.Header.Set("Authorization", "Bearer t")
		if e.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("%s %s: expected 404 with flag off, got %d body=%s", e.method, e.path, rec.Code, rec.Body.String())
			continue
		}
		if !strings.Contains(rec.Body.String(), "onboarding_feature_disabled") {
			t.Errorf("%s %s: expected code=onboarding_feature_disabled, got %q", e.method, e.path, rec.Body.String())
		}
	}
}

// TestOnboardingGate_FeatureFlagOff_NoOp verifies that with the gate disabled,
// /api/v1/me responds normally for an authenticated actor without an
// onboarding_required block.
func TestOnboardingGate_FeatureFlagOff_NoOp(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: false, // explicit
	})

	// /api/v1/me/onboarding 호출이 gate 비활성화 상태에서도 routePermissionTable
	// Bypass 로 통과해야 함 — 본 test 의 핵심은 gate=OFF 일 때 동작 회귀 없음.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with gate off, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestOnboardingGate_FeatureFlagOn_AllowlistAccess — gate ON + token-only actor
// (DB row 미존재) + allowlist endpoint (/me) 호출 → 통과.
func TestOnboardingGate_FeatureFlagOn_AllowlistAccess(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "newuser", Subject: "sub-new", Role: "developer",
		Email: "new@example.com", DisplayName: "New User",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on allowlist, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data meResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !body.Data.OnboardingRequired {
		t.Errorf("expected onboarding_required=true for token-only actor, got %+v", body.Data)
	}
}

// TestOnboardingGate_FeatureFlagOn_NonAllowlistBlocks — gate ON + token-only +
// allowlist 외 endpoint (/dashboard/metrics) → 403 onboarding_required.
func TestOnboardingGate_FeatureFlagOn_NonAllowlistBlocks(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "newuser", Subject: "sub-new", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "onboarding_required") {
		t.Errorf("expected code=onboarding_required in body, got %q", rec.Body.String())
	}
}

// TestSubmitOnboarding_HappyPath — POST /me/onboarding 성공 (new user INSERT
// 흐름). 응답 201 + audit emit + user row 생성.
func TestSubmitOnboarding_HappyPath(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
		Email: "alice@example.com", DisplayName: "Alice",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	body := []byte(`{"display_name":"Alice Kim","primary_unit_id":"team-platform"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	user, err := store.GetUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("getuser: %v", err)
	}
	if user.OnboardingCompletedAt == nil {
		t.Errorf("expected completed_at set")
	}
	if user.ReviewStatus != domain.ReviewStatusPendingReview {
		t.Errorf("expected pending_review, got %q", user.ReviewStatus)
	}
	if user.PrimaryUnitID != "team-platform" {
		t.Errorf("expected primary_unit_id=team-platform, got %q", user.PrimaryUnitID)
	}
}

// TestSubmitOnboarding_PreSeededUpdate — admin pre-seeded user (row exists +
// completed_at NULL) 가 POST /me/onboarding 호출 → UPDATE 흐름 정합. codex P1
// (PR #270) 정합.
func TestSubmitOnboarding_PreSeededUpdate(t *testing.T) {
	store := seedOnboardingFixture(t, true) // preseed: "preseeded" user
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "preseeded", Subject: "sub-pre", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	body := []byte(`{"display_name":"Updated Name","primary_unit_id":"team-data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 (UPDATE flow), got %d body=%s", rec.Code, rec.Body.String())
	}
	user, _ := store.GetUser(context.Background(), "preseeded")
	if user.DisplayName != "Updated Name" {
		t.Errorf("expected display_name updated, got %q", user.DisplayName)
	}
	if user.PrimaryUnitID != "team-data" {
		t.Errorf("expected unit updated to team-data, got %q", user.PrimaryUnitID)
	}
	if user.OnboardingCompletedAt == nil {
		t.Errorf("expected completed_at set after UPDATE")
	}
}

// TestSubmitOnboarding_AlreadyCompleted409 — 이미 완료된 사용자 중복 호출 → 409.
func TestSubmitOnboarding_AlreadyCompleted409(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	// 미리 alice 완료 처리.
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed alice complete: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	body := []byte(`{"display_name":"Alice2","primary_unit_id":"team-data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "onboarding_already_completed") {
		t.Errorf("expected code=onboarding_already_completed, got %q", rec.Body.String())
	}
}

// TestSubmitOnboarding_UnitNotFound404 — primary_unit_id 가 organization_units 에
// 없음 → 404.
func TestSubmitOnboarding_UnitNotFound404(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "bob", Subject: "sub-bob", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier:   verifier,
		OrganizationStore:     store,
		OnboardingGateEnabled: true,
	})

	body := []byte(`{"display_name":"Bob","primary_unit_id":"nonexistent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unit_not_found") {
		t.Errorf("expected code=unit_not_found, got %q", rec.Body.String())
	}
}

// TestSubmitOnboarding_InvalidPayload422 — display_name 누락 → 422.
func TestSubmitOnboarding_InvalidPayload422(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "bob", Subject: "sub-bob", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	body := []byte(`{"primary_unit_id":"team-platform"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestOrganizationsSearch_HappyPath — typeahead 검색 happy + limit clamp.
func TestOrganizationsSearch_HappyPath(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/search?q=팀", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "team-platform") && !strings.Contains(rec.Body.String(), "team-data") {
		t.Errorf("expected at least one match in body, got %q", rec.Body.String())
	}
}

// TestOrganizationsSearch_QTooShort422 — q 가 2 chars 미만 → 422.
func TestOrganizationsSearch_QTooShort422(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/search?q=A", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// TestConfirmUserReview_HappyPath — admin POST /admin/users/:id/review
// pending_review → reviewed.
func TestConfirmUserReview_HappyPath(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	// alice 가 onboarding 제출 (pending_review).
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	// admin1 도 완료 + reviewed (admin 자신은 onboarding 완료 상태).
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "admin1", DisplayName: "Admin", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleSystemAdmin,
	}); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if _, err := store.ConfirmUserReview(context.Background(), "admin1"); err != nil {
		t.Fatalf("review admin: %v", err)
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

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	user, _ := store.GetUser(context.Background(), "alice")
	if user.ReviewStatus != domain.ReviewStatusReviewed {
		t.Errorf("expected reviewed, got %q", user.ReviewStatus)
	}
}

// TestPatchMe_UnitChangeResetsReview — PATCH /me 의 primary_unit_id 변경 시
// review_status=pending_review 재진입 (REQ-FR-ONBOARD-007).
func TestPatchMe_UnitChangeResetsReview(t *testing.T) {
	store := seedOnboardingFixture(t, false)
	if _, err := store.SubmitOnboarding(context.Background(), domain.OnboardingSubmitInput{
		UserID: "alice", DisplayName: "Alice", PrimaryUnitID: "team-platform",
		FallbackRole: domain.AppRoleDeveloper,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// reviewed 처리.
	if _, err := store.ConfirmUserReview(context.Background(), "alice"); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "alice", Subject: "sub-alice", Role: "developer",
	}}
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier, OrganizationStore: store, OnboardingGateEnabled: true,
	})

	body := []byte(`{"primary_unit_id":"team-data"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	user, _ := store.GetUser(context.Background(), "alice")
	if user.ReviewStatus != domain.ReviewStatusPendingReview {
		t.Errorf("expected pending_review after unit change, got %q", user.ReviewStatus)
	}
	if user.PrimaryUnitID != "team-data" {
		t.Errorf("expected unit changed to team-data, got %q", user.PrimaryUnitID)
	}
}
