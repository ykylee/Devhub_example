package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/store"
)

// fakeUserSyncOrgStore — UserSyncOrgStore 의 memory mock. PR #241 codex P1
// hotfix 회귀 test 용. GetUser / GetUserByIdPSubject / UpdateUser 에 각각
// error injection 가능 + 호출 count 추적.
type fakeUserSyncOrgStore struct {
	users           map[string]domain.AppUser // user_id → user
	idpIndex        map[string]string         // idp_subject → user_id
	getUserErr      error                     // GetUser 가 반환할 error (injection)
	getByIdPErr     error                     // GetUserByIdPSubject 가 반환할 error
	updateCalls     int
	updateInputs    []updateCall
}

type updateCall struct {
	userID string
	input  domain.UpdateUserInput
}

func newFakeUserSyncOrgStore() *fakeUserSyncOrgStore {
	return &fakeUserSyncOrgStore{
		users:    map[string]domain.AppUser{},
		idpIndex: map[string]string{},
	}
}

func (f *fakeUserSyncOrgStore) seedUser(u domain.AppUser) {
	f.users[u.UserID] = u
	if u.IdPSubject != "" {
		f.idpIndex[u.IdPSubject] = u.UserID
	}
}

func (f *fakeUserSyncOrgStore) GetUser(_ context.Context, userID string) (domain.AppUser, error) {
	if f.getUserErr != nil {
		return domain.AppUser{}, f.getUserErr
	}
	u, ok := f.users[userID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	return u, nil
}

func (f *fakeUserSyncOrgStore) GetUserByIdPSubject(_ context.Context, identityID string) (domain.AppUser, error) {
	if f.getByIdPErr != nil {
		return domain.AppUser{}, f.getByIdPErr
	}
	userID, ok := f.idpIndex[identityID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user idp_subject=%s: %w", identityID, store.ErrNotFound)
	}
	return f.users[userID], nil
}

func (f *fakeUserSyncOrgStore) UpdateUser(_ context.Context, userID string, input domain.UpdateUserInput) (domain.AppUser, error) {
	f.updateCalls++
	f.updateInputs = append(f.updateInputs, updateCall{userID: userID, input: input})
	u, ok := f.users[userID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	if input.Status != nil {
		u.Status = *input.Status
	}
	if input.Role != nil {
		u.Role = *input.Role
	}
	if input.Email != nil {
		u.Email = *input.Email
	}
	if input.DisplayName != nil {
		u.DisplayName = *input.DisplayName
	}
	f.users[userID] = u
	return u, nil
}

// fakeAdminClient — UserSyncAdminClient mock.
type fakeAdminClient struct {
	detailsByID map[string]httpapi.KeycloakUserDetails
	groupsByID  map[string][]httpapi.KeycloakGroup
	detailsErr  error
	groupsErr   error
}

func (f *fakeAdminClient) GetUserDetails(_ context.Context, identityID string) (httpapi.KeycloakUserDetails, error) {
	if f.detailsErr != nil {
		return httpapi.KeycloakUserDetails{}, f.detailsErr
	}
	d, ok := f.detailsByID[identityID]
	if !ok {
		return httpapi.KeycloakUserDetails{}, fmt.Errorf("not found")
	}
	return d, nil
}

func (f *fakeAdminClient) GetUserGroups(_ context.Context, identityID string) ([]httpapi.KeycloakGroup, error) {
	if f.groupsErr != nil {
		return nil, f.groupsErr
	}
	return f.groupsByID[identityID], nil
}

// TestMarkUserDeactivated_LooksUpByIdPSubject — PR #241 codex P1 hotfix 회귀.
// identity_id 로 호출하면 GetUserByIdPSubject 로 user_id lookup 후 UpdateUser.
func TestMarkUserDeactivated_LooksUpByIdPSubject(t *testing.T) {
	orgs := newFakeUserSyncOrgStore()
	orgs.seedUser(domain.AppUser{
		UserID:      "alice",
		IdPSubject:  "kc-uuid-alice",
		Status:      domain.UserStatusActive,
		Role:        domain.AppRoleDeveloper,
	})

	err := MarkUserDeactivated(context.Background(), orgs, "kc-uuid-alice")
	if err != nil {
		t.Fatalf("MarkUserDeactivated: %v", err)
	}
	if orgs.updateCalls != 1 {
		t.Fatalf("UpdateUser call count = %d; want 1", orgs.updateCalls)
	}
	call := orgs.updateInputs[0]
	if call.userID != "alice" {
		t.Errorf("UpdateUser userID = %q; want %q (idp_subject lookup 결과)", call.userID, "alice")
	}
	if call.input.Status == nil || *call.input.Status != domain.UserStatusDeactivated {
		t.Errorf("Status = %v; want deactivated (soft delete, ADR-0020 §5.3.1 P1-2)", call.input.Status)
	}
}

// TestMarkUserDeactivated_NotFoundNoop — DevHub row 없으면 noop (pre-onboarding
// 또는 idp_subject lazy backfill 미완 상태).
func TestMarkUserDeactivated_NotFoundNoop(t *testing.T) {
	orgs := newFakeUserSyncOrgStore()
	// seed 없이 호출 → idp_subject lookup 이 ErrNotFound
	err := MarkUserDeactivated(context.Background(), orgs, "kc-uuid-ghost")
	if err != nil {
		t.Errorf("ErrNotFound 은 noop 이어야 함, got: %v", err)
	}
	if orgs.updateCalls != 0 {
		t.Errorf("UpdateUser 호출 안 함, got %d", orgs.updateCalls)
	}
}

// TestMarkUserDeactivated_DBErrorPropagated — PR #241 codex P1 hotfix. DB 장애 등
// non-ErrNotFound error 는 propagate (이전엔 모든 err swallow → silent metric success).
func TestMarkUserDeactivated_DBErrorPropagated(t *testing.T) {
	orgs := newFakeUserSyncOrgStore()
	orgs.getByIdPErr = errors.New("simulated db outage")
	err := MarkUserDeactivated(context.Background(), orgs, "kc-uuid-anything")
	if err == nil {
		t.Fatal("DB error 가 propagate 되어야 함 (codex P1 응답), got nil")
	}
	if !errors.Is(err, orgs.getByIdPErr) && err.Error() == "" {
		t.Errorf("err 본문에 underlying cause 포함되어야: %v", err)
	}
	if orgs.updateCalls != 0 {
		t.Error("DB error 시 UpdateUser 호출 안 함")
	}
}

// TestSyncUserProfile_NotFoundNoop_ErrorPropagated — PR #241 codex P1 hotfix.
// GetUser 의 ErrNotFound 만 noop, 그 외 (DB 장애 등) propagate.
func TestSyncUserProfile_NotFoundNoop_ErrorPropagated(t *testing.T) {
	t.Run("ErrNotFound noop", func(t *testing.T) {
		orgs := newFakeUserSyncOrgStore()
		admin := &fakeAdminClient{
			detailsByID: map[string]httpapi.KeycloakUserDetails{
				"kc-id-1": {ID: "kc-id-1", Username: "alice", Email: "alice@example.com", Enabled: true},
			},
		}
		// alice DevHub row 없음 → GetUser ErrNotFound → noop
		if err := SyncUserProfile(context.Background(), admin, orgs, "kc-id-1"); err != nil {
			t.Errorf("ErrNotFound noop expected, got: %v", err)
		}
		if orgs.updateCalls != 0 {
			t.Errorf("UpdateUser 호출 안 함, got %d", orgs.updateCalls)
		}
	})
	t.Run("DB error propagated", func(t *testing.T) {
		orgs := newFakeUserSyncOrgStore()
		orgs.getUserErr = errors.New("simulated db outage")
		admin := &fakeAdminClient{
			detailsByID: map[string]httpapi.KeycloakUserDetails{
				"kc-id-1": {ID: "kc-id-1", Username: "alice"},
			},
		}
		if err := SyncUserProfile(context.Background(), admin, orgs, "kc-id-1"); err == nil {
			t.Fatal("DB error propagate 되어야 함")
		}
	})
}

// TestSyncUserProfile_UpdatesProfileFields — happy path. email/display_name/status
// 모두 정상 set.
func TestSyncUserProfile_UpdatesProfileFields(t *testing.T) {
	orgs := newFakeUserSyncOrgStore()
	orgs.seedUser(domain.AppUser{
		UserID: "alice", Email: "old@example.com", DisplayName: "Old Name",
		IdPSubject: "kc-id-1", Status: domain.UserStatusActive, Role: domain.AppRoleDeveloper,
	})
	admin := &fakeAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			"kc-id-1": {
				ID:        "kc-id-1",
				Username:  "alice",
				Email:     "new@example.com",
				FirstName: "Alice",
				LastName:  "Smith",
				Enabled:   true,
			},
		},
	}
	if err := SyncUserProfile(context.Background(), admin, orgs, "kc-id-1"); err != nil {
		t.Fatalf("SyncUserProfile: %v", err)
	}
	if orgs.updateCalls != 1 {
		t.Fatalf("UpdateUser count = %d; want 1", orgs.updateCalls)
	}
	in := orgs.updateInputs[0].input
	if in.Email == nil || *in.Email != "new@example.com" {
		t.Errorf("Email = %v; want new@example.com", in.Email)
	}
	if in.DisplayName == nil || *in.DisplayName != "Alice Smith" {
		t.Errorf("DisplayName = %v; want Alice Smith", in.DisplayName)
	}
	if in.Status == nil || *in.Status != domain.UserStatusActive {
		t.Errorf("Status = %v; want active", in.Status)
	}
}

// TestSyncUserMembership_PicksHighestPriorityRole — group composite role priority filter.
func TestSyncUserMembership_PicksHighestPriorityRole(t *testing.T) {
	orgs := newFakeUserSyncOrgStore()
	orgs.seedUser(domain.AppUser{
		UserID: "alice", IdPSubject: "kc-id-1",
		Status: domain.UserStatusActive, Role: domain.AppRoleDeveloper,
	})
	admin := &fakeAdminClient{
		detailsByID: map[string]httpapi.KeycloakUserDetails{
			"kc-id-1": {ID: "kc-id-1", Username: "alice"},
		},
		groupsByID: map[string][]httpapi.KeycloakGroup{
			"kc-id-1": {
				{Name: "devhub-developers"},
				{Name: "devhub-managers"}, // 더 우선
				{Name: "some-other-group"},
			},
		},
	}
	if err := SyncUserMembership(context.Background(), admin, orgs, "kc-id-1"); err != nil {
		t.Fatalf("SyncUserMembership: %v", err)
	}
	in := orgs.updateInputs[0].input
	if in.Role == nil || *in.Role != domain.AppRoleTeamManager {
		t.Errorf("Role = %v; want manager (priority filter)", in.Role)
	}
}

// TestParseIdentityIDFromResourcePath — admin event ResourcePath 파싱.
func TestParseIdentityIDFromResourcePath(t *testing.T) {
	cases := []struct {
		path, want string
	}{
		{"users/abc-uuid", "abc-uuid"},
		{"users/abc-uuid/role-mappings/realm", "abc-uuid"},
		{"groups/g1", ""}, // not users prefix
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := ParseIdentityIDFromResourcePath(tc.path); got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
