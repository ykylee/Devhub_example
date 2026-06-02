package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/audit-ops/service"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/store"
)

// audit-ops/service 의 user_sync.go 3 hot path (SyncUserProfile,
// SyncUserMembership, MarkUserDeactivated) 의 실 DB integration test.
//
// 직전 단위테스트 (user_sync_test.go) 는 fake UserSyncOrgStore 로 cover —
// production *store.PostgresStore 와 fake 의 signature 차이 (예: UpdateUser
// 가 silent skip 하는 field, idp_subject 컬럼 lookup 의 실 SQL) 가 silent
// 회귀 가능. 본 integration test 가 wire 정합성 catch-net.
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip — 기존 패턴.
//
// fixture 정책:
//   - TRUNCATE 회피 (cross-pkg race 위험), prefix 기반 DELETE 만.
//   - test user_id 는 `audit_sync_test_` prefix 로 unique 보장.
//   - primary_unit_id 는 migration 000004 seed 의 `dept-eng` (이미 존재) 사용.

const userSyncTestPrefix = "audit_sync_test_"

// fakeUserSyncAdminClient — production 의 외부 Keycloak admin API call 을
// 우회하는 in-process fake. user_sync_test.go 의 fakeAdminClient 와 본질
// 동일하지만 _test 패키지 분리 (`service_test`) 라 재선언.
type fakeUserSyncAdminClient struct {
	detailsByID map[string]httpapi.KeycloakUserDetails
	groupsByID  map[string][]httpapi.KeycloakGroup
	detailsErr  error
	groupsErr   error
}

func (f *fakeUserSyncAdminClient) GetUserDetails(_ context.Context, identityID string) (httpapi.KeycloakUserDetails, error) {
	if f.detailsErr != nil {
		return httpapi.KeycloakUserDetails{}, f.detailsErr
	}
	d, ok := f.detailsByID[identityID]
	if !ok {
		return httpapi.KeycloakUserDetails{}, fmt.Errorf("admin: user %s not found", identityID)
	}
	return d, nil
}

func (f *fakeUserSyncAdminClient) GetUserGroups(_ context.Context, identityID string) ([]httpapi.KeycloakGroup, error) {
	if f.groupsErr != nil {
		return nil, f.groupsErr
	}
	return f.groupsByID[identityID], nil
}

func setupUserSyncIntegrationTest(t *testing.T) (*store.PostgresStore, context.Context, func()) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	// pre-test purge — prior test run residue.
	purge := func() {
		_, _ = pgStore.Pool().Exec(context.Background(),
			`DELETE FROM users WHERE user_id LIKE $1`, userSyncTestPrefix+"%")
	}
	purge()
	return pgStore, ctx, func() {
		purge()
		pgStore.Close()
	}
}

// seedTestUser INSERT 후 idp_subject 컬럼 UPDATE (CreateUser 는 input 에 IdPSubject
// 가 없어 raw SQL 로 후속 설정).
func seedTestUser(t *testing.T, ctx context.Context, s *store.PostgresStore, userID, email, displayName, idpSubject string) domain.AppUser {
	t.Helper()
	user, err := s.CreateUser(ctx, domain.CreateUserInput{
		UserID:        userID,
		Email:         email,
		DisplayName:   displayName,
		Role:          domain.AppRoleDeveloper,
		Status:        domain.UserStatusActive,
		Type:          domain.UserTypeHuman,
		PrimaryUnitID: "dept-eng",
		JoinedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed user %s: %v", userID, err)
	}
	if idpSubject != "" {
		if _, err := s.Pool().Exec(ctx, `UPDATE users SET idp_subject = $1 WHERE user_id = $2`, idpSubject, userID); err != nil {
			t.Fatalf("set idp_subject for %s: %v", userID, err)
		}
	}
	return user
}

// --- SyncUserProfile ---

func TestIntegration_SyncUserProfile_UpdatesEmailDisplayStatus(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	userID := userSyncTestPrefix + "profile_happy"
	identityID := "kc-uuid-profile-happy"
	seedTestUser(t, ctx, s, userID, "old@test.example.com", "Old Name", "")

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {
				Username:  userID,
				Email:     "new@test.example.com",
				FirstName: "New",
				LastName:  "Person",
				Enabled:   true,
			},
		},
	}

	if err := service.SyncUserProfile(ctx, admin, s, identityID); err != nil {
		t.Fatalf("SyncUserProfile: %v", err)
	}

	got, err := s.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("post-sync GetUser: %v", err)
	}
	if got.Email != "new@test.example.com" {
		t.Errorf("email = %q, want new@test.example.com", got.Email)
	}
	if got.DisplayName != "New Person" {
		t.Errorf("display_name = %q, want 'New Person'", got.DisplayName)
	}
	if got.Status != domain.UserStatusActive {
		t.Errorf("status = %q, want active", got.Status)
	}
}

func TestIntegration_SyncUserProfile_DisabledFlagMarksDeactivated(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	userID := userSyncTestPrefix + "profile_disabled"
	identityID := "kc-uuid-profile-disabled"
	seedTestUser(t, ctx, s, userID, "u@test.example.com", "User", "")

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {
				Username:  userID,
				Email:     "u@test.example.com",
				FirstName: "User",
				Enabled:   false,
			},
		},
	}

	if err := service.SyncUserProfile(ctx, admin, s, identityID); err != nil {
		t.Fatalf("SyncUserProfile: %v", err)
	}

	got, _ := s.GetUser(ctx, userID)
	if got.Status != domain.UserStatusDeactivated {
		t.Errorf("disabled Keycloak user → DevHub status = %q, want deactivated", got.Status)
	}
}

func TestIntegration_SyncUserProfile_MissingDevHubRowNoop(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	identityID := "kc-uuid-no-devhub-row"
	missingUserID := userSyncTestPrefix + "noop_target"

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {Username: missingUserID, Email: "x@test.example.com", Enabled: true},
		},
	}

	// pre-onboarding 시점이면 DevHub row 가 없어도 정상 (lazy auto-create scope).
	if err := service.SyncUserProfile(ctx, admin, s, identityID); err != nil {
		t.Errorf("missing DevHub row should be noop, got err: %v", err)
	}
	// Row 가 실제로 미생성됨을 확인 (production 이 silent INSERT 회귀 가드).
	if _, err := s.GetUser(ctx, missingUserID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected user still ErrNotFound (noop must not create), got %v", err)
	}
}

func TestIntegration_SyncUserProfile_EmptyKeycloakUsernameErr(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	identityID := "kc-uuid-empty-username"
	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {Username: "  ", Email: "x@test.example.com", Enabled: true},
		},
	}

	err := service.SyncUserProfile(ctx, admin, s, identityID)
	if err == nil {
		t.Fatal("expected error on empty Keycloak username, got nil")
	}
}

// --- SyncUserMembership ---

func TestIntegration_SyncUserMembership_AssignsHighestPriorityRole(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	userID := userSyncTestPrefix + "membership_priority"
	identityID := "kc-uuid-membership-priority"
	seedTestUser(t, ctx, s, userID, "m@test.example.com", "M Person", "")

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {Username: userID, Enabled: true, FirstName: "M"},
		},
		groupsByID: map[string][]httpapi.KeycloakGroup{
			identityID: {
				{Name: "devhub-developers"},
				{Name: "devhub-managers"},
				{Name: "irrelevant-group"},
			},
		},
	}

	if err := service.SyncUserMembership(ctx, admin, s, identityID); err != nil {
		t.Fatalf("SyncUserMembership: %v", err)
	}

	got, _ := s.GetUser(ctx, userID)
	if got.Role != domain.AppRoleTeamManager {
		t.Errorf("role = %q, want manager (highest priority of [developer, manager])", got.Role)
	}
}

func TestIntegration_SyncUserMembership_DefaultDeveloperOnNoMatch(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	userID := userSyncTestPrefix + "membership_default"
	identityID := "kc-uuid-membership-default"
	seedTestUser(t, ctx, s, userID, "d@test.example.com", "D", "")

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {Username: userID, Enabled: true},
		},
		groupsByID: map[string][]httpapi.KeycloakGroup{
			identityID: {{Name: "non-devhub-group"}},
		},
	}

	if err := service.SyncUserMembership(ctx, admin, s, identityID); err != nil {
		t.Fatalf("SyncUserMembership: %v", err)
	}

	got, _ := s.GetUser(ctx, userID)
	if got.Role != domain.AppRoleDeveloper {
		t.Errorf("default role = %q, want developer (no DevHub-prefixed group match)", got.Role)
	}
}

func TestIntegration_SyncUserMembership_MissingDevHubRowNoop(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	identityID := "kc-uuid-membership-noop"
	missingUserID := userSyncTestPrefix + "membership_noop_target"

	admin := &fakeUserSyncAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			identityID: {Username: missingUserID, Enabled: true},
		},
		groupsByID: map[string][]httpapi.KeycloakGroup{
			identityID: {{Name: "devhub-developers"}},
		},
	}

	if err := service.SyncUserMembership(ctx, admin, s, identityID); err != nil {
		t.Errorf("missing DevHub row should be noop, got: %v", err)
	}
}

// --- MarkUserDeactivated ---

// PR #241 codex P1 hotfix 회귀 가드 — identity_id 로 GetUserByIdPSubject lookup
// 후 UpdateUser status=deactivated. fake store 의 idpIndex 와 실 DB 의
// idp_subject 컬럼 UNIQUE index 가 의미적 동일한지 검증.
func TestIntegration_MarkUserDeactivated_SoftDeleteHappy(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	userID := userSyncTestPrefix + "delete_happy"
	identityID := "kc-uuid-delete-happy"
	seedTestUser(t, ctx, s, userID, "d@test.example.com", "D", identityID)

	if err := service.MarkUserDeactivated(ctx, s, identityID); err != nil {
		t.Fatalf("MarkUserDeactivated: %v", err)
	}

	got, _ := s.GetUser(ctx, userID)
	if got.Status != domain.UserStatusDeactivated {
		t.Errorf("status = %q, want deactivated (soft delete)", got.Status)
	}
	if got.UserID != userID {
		t.Errorf("user_id changed? got %q, want %q (row 보존 ADR-0020 §5.3.1)", got.UserID, userID)
	}
}

// idp_subject 없는 case — pre-onboarding 또는 backfill 미완 → noop.
// production 이 GetUserByIdPSubject 가 ErrNotFound 일 때 silent noop.
func TestIntegration_MarkUserDeactivated_MissingIdPSubjectNoop(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	identityID := "kc-uuid-no-mapping"
	if err := service.MarkUserDeactivated(ctx, s, identityID); err != nil {
		t.Errorf("missing idp_subject mapping should be noop, got: %v", err)
	}
}

func TestIntegration_MarkUserDeactivated_EmptyIdentityErr(t *testing.T) {
	s, ctx, teardown := setupUserSyncIntegrationTest(t)
	defer teardown()

	if err := service.MarkUserDeactivated(ctx, s, "   "); err == nil {
		t.Fatal("expected error on empty identity_id, got nil")
	}
}
