package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// memoryOrganizationStore is an in-memory OrganizationStore used by handler
// tests. It only models the parts the handlers exercise.
type memoryOrganizationStore struct {
	users map[string]domain.AppUser
	units map[string]domain.OrgUnit
	// memberships keyed by unit_id then user_id => appointment role
	memberships map[string]map[string]domain.AppointmentRole
}

func newMemoryOrganizationStore() *memoryOrganizationStore {
	return &memoryOrganizationStore{
		users:       make(map[string]domain.AppUser),
		units:       make(map[string]domain.OrgUnit),
		memberships: make(map[string]map[string]domain.AppointmentRole),
	}
}

func (s *memoryOrganizationStore) ListUsers(_ context.Context, opts domain.UserListOptions) ([]domain.AppUser, int, error) {
	out := make([]domain.AppUser, 0, len(s.users))
	for _, user := range s.users {
		if opts.Role != "" && string(user.Role) != opts.Role {
			continue
		}
		if opts.Status != "" && string(user.Status) != opts.Status {
			continue
		}
		if opts.PrimaryUnitID != "" && user.PrimaryUnitID != opts.PrimaryUnitID {
			continue
		}
		out = append(out, user)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out, len(out), nil
}

func (s *memoryOrganizationStore) GetUser(_ context.Context, userID string) (domain.AppUser, error) {
	user, ok := s.users[userID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	return user, nil
}

func (s *memoryOrganizationStore) CreateUser(_ context.Context, input domain.CreateUserInput) (domain.AppUser, error) {
	if _, exists := s.users[input.UserID]; exists {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", input.UserID, store.ErrConflict)
	}
	for _, existing := range s.users {
		if existing.Email == input.Email {
			return domain.AppUser{}, fmt.Errorf("email %s: %w", input.Email, store.ErrConflict)
		}
	}
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	user := domain.AppUser{
		ID:            int64(len(s.users) + 1),
		UserID:        input.UserID,
		Email:         input.Email,
		DisplayName:   input.DisplayName,
		Role:          input.Role,
		Status:        input.Status,
		PrimaryUnitID: input.PrimaryUnitID,
		CurrentUnitID: input.CurrentUnitID,
		IsSeconded:    input.IsSeconded,
		JoinedAt:      input.JoinedAt,
		Appointments:  []domain.UnitAppointment{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.users[input.UserID] = user
	return user, nil
}

func (s *memoryOrganizationStore) UpdateUser(_ context.Context, userID string, input domain.UpdateUserInput) (domain.AppUser, error) {
	user, ok := s.users[userID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	if input.Email != nil {
		for otherID, existing := range s.users {
			if otherID != userID && existing.Email == *input.Email {
				return domain.AppUser{}, fmt.Errorf("email %s: %w", *input.Email, store.ErrConflict)
			}
		}
		user.Email = *input.Email
	}
	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.Status != nil {
		user.Status = *input.Status
	}
	if input.PrimaryUnitID != nil {
		// production *store.PostgresStore 의 FK constraint mirror — primary_unit_id
		// 가 organization_units(unit_id) 에 없으면 ErrNotFound. PATCH /me 의
		// `unit_not_found` 404 분기 (API-85 §16.5) 검증에 필수.
		if *input.PrimaryUnitID != "" {
			if _, ok := s.units[*input.PrimaryUnitID]; !ok {
				return domain.AppUser{}, fmt.Errorf("unit %s: %w", *input.PrimaryUnitID, store.ErrNotFound)
			}
		}
		user.PrimaryUnitID = *input.PrimaryUnitID
	}
	if input.CurrentUnitID != nil {
		user.CurrentUnitID = *input.CurrentUnitID
	}
	if input.IsSeconded != nil {
		user.IsSeconded = *input.IsSeconded
	}
	if input.JoinedAt != nil {
		user.JoinedAt = *input.JoinedAt
	}
	if input.ReviewStatus != nil {
		user.ReviewStatus = *input.ReviewStatus
	}
	user.UpdatedAt = time.Date(2026, 5, 7, 13, 0, 0, 0, time.UTC)
	s.users[userID] = user
	return user, nil
}

func (s *memoryOrganizationStore) DeleteUser(_ context.Context, userID string) error {
	if _, ok := s.users[userID]; !ok {
		return fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	delete(s.users, userID)
	for unitID, members := range s.memberships {
		delete(members, userID)
		if len(members) == 0 {
			delete(s.memberships, unitID)
		}
	}
	return nil
}

func (s *memoryOrganizationStore) SetIdPSubject(_ context.Context, userID, identityID string) error {
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	user.IdPSubject = identityID
	user.UpdatedAt = time.Date(2026, 5, 7, 13, 0, 0, 0, time.UTC)
	s.users[userID] = user
	return nil
}

// GetUserByIdPSubject — ADR-0020 sub-carve C (sprint -k, issue #212 codex P1
// hotfix). idp_subject 컬럼 lookup. 메모리 fake 는 O(n) 선형 탐색이지만
// test 용이라 acceptable.
func (s *memoryOrganizationStore) GetUserByIdPSubject(_ context.Context, identityID string) (domain.AppUser, error) {
	for _, user := range s.users {
		if user.IdPSubject == identityID {
			return user, nil
		}
	}
	return domain.AppUser{}, fmt.Errorf("user idp_subject=%s: %w", identityID, store.ErrNotFound)
}

func (s *memoryOrganizationStore) GetHierarchy(_ context.Context) (domain.Hierarchy, error) {
	units := make([]domain.OrgUnit, 0, len(s.units))
	for _, unit := range s.units {
		units = append(units, unit)
	}
	sort.Slice(units, func(i, j int) bool { return units[i].UnitID < units[j].UnitID })
	edges := make([]domain.OrgEdge, 0)
	for _, unit := range units {
		if unit.ParentUnitID != "" {
			edges = append(edges, domain.OrgEdge{SourceUnitID: unit.ParentUnitID, TargetUnitID: unit.UnitID})
		}
	}
	return domain.Hierarchy{Units: units, Edges: edges}, nil
}

func (s *memoryOrganizationStore) UpdateHierarchy(_ context.Context, hie domain.Hierarchy) error {
	next := make(map[string]domain.OrgUnit, len(hie.Units))
	for _, unit := range hie.Units {
		next[unit.UnitID] = unit
	}
	s.units = next
	return nil
}

func (s *memoryOrganizationStore) GetOrgUnit(_ context.Context, unitID string) (domain.OrgUnit, error) {
	unit, ok := s.units[unitID]
	if !ok {
		return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", unitID, store.ErrNotFound)
	}
	return unit, nil
}

func (s *memoryOrganizationStore) CreateOrgUnit(_ context.Context, input domain.CreateOrgUnitInput) (domain.OrgUnit, error) {
	if _, exists := s.units[input.UnitID]; exists {
		return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", input.UnitID, store.ErrConflict)
	}
	if input.ParentUnitID != "" {
		if _, ok := s.units[input.ParentUnitID]; !ok {
			return domain.OrgUnit{}, fmt.Errorf("parent unit %s: %w", input.ParentUnitID, store.ErrNotFound)
		}
	}
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	unit := domain.OrgUnit{
		ID:           int64(len(s.units) + 1),
		UnitID:       input.UnitID,
		ParentUnitID: input.ParentUnitID,
		UnitType:     input.UnitType,
		Label:        input.Label,
		LeaderUserID: input.LeaderUserID,
		PositionX:    input.PositionX,
		PositionY:    input.PositionY,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.units[input.UnitID] = unit
	return unit, nil
}

func (s *memoryOrganizationStore) UpdateOrgUnit(_ context.Context, unitID string, input domain.UpdateOrgUnitInput) (domain.OrgUnit, error) {
	unit, ok := s.units[unitID]
	if !ok {
		return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", unitID, store.ErrNotFound)
	}
	if input.ParentUnitID != nil {
		unit.ParentUnitID = *input.ParentUnitID
	}
	if input.UnitType != nil {
		unit.UnitType = *input.UnitType
	}
	if input.Label != nil {
		unit.Label = *input.Label
	}
	if input.LeaderUserID != nil {
		unit.LeaderUserID = *input.LeaderUserID
	}
	if input.PositionX != nil {
		unit.PositionX = *input.PositionX
	}
	if input.PositionY != nil {
		unit.PositionY = *input.PositionY
	}
	unit.UpdatedAt = time.Date(2026, 5, 7, 13, 0, 0, 0, time.UTC)
	s.units[unitID] = unit
	return unit, nil
}

func (s *memoryOrganizationStore) DeleteOrgUnit(_ context.Context, unitID string) error {
	if _, ok := s.units[unitID]; !ok {
		return fmt.Errorf("unit %s: %w", unitID, store.ErrNotFound)
	}
	delete(s.units, unitID)
	delete(s.memberships, unitID)
	return nil
}

func (s *memoryOrganizationStore) ListUnitMembers(_ context.Context, unitID string) ([]domain.AppUser, error) {
	members, ok := s.memberships[unitID]
	if !ok {
		return []domain.AppUser{}, nil
	}
	out := make([]domain.AppUser, 0, len(members))
	for userID := range members {
		if user, ok := s.users[userID]; ok {
			out = append(out, user)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out, nil
}

func (s *memoryOrganizationStore) ReplaceUnitMembers(_ context.Context, unitID string, userIDs []string) error {
	if _, ok := s.units[unitID]; !ok {
		return fmt.Errorf("unit %s: %w", unitID, store.ErrNotFound)
	}
	current := s.memberships[unitID]
	if current == nil {
		current = make(map[string]domain.AppointmentRole)
	}
	// Drop existing member appointments, keep leader appointments.
	for userID, role := range current {
		if role == domain.AppointmentRoleMember {
			delete(current, userID)
		}
	}
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if existing, ok := current[userID]; ok && existing == domain.AppointmentRoleLeader {
			continue
		}
		current[userID] = domain.AppointmentRoleMember
	}
	if len(current) == 0 {
		delete(s.memberships, unitID)
	} else {
		s.memberships[unitID] = current
	}
	return nil
}

// RM-ONBOARD-01 (ADR-0021 §3.3, API-83). Onboarding 제출 — row INSERT 또는
// UPDATE 단일 트랜잭션 (memory store 에서는 단순 map mutation).
func (s *memoryOrganizationStore) SubmitOnboarding(_ context.Context, input domain.OnboardingSubmitInput) (domain.AppUser, error) {
	if _, ok := s.units[input.PrimaryUnitID]; !ok {
		return domain.AppUser{}, fmt.Errorf("unit %s: %w", input.PrimaryUnitID, store.ErrNotFound)
	}
	now := time.Now().UTC()
	role := input.FallbackRole
	if role == "" {
		role = domain.AppRoleDeveloper
	}
	if existing, ok := s.users[input.UserID]; ok {
		if existing.OnboardingCompletedAt != nil {
			return domain.AppUser{}, fmt.Errorf("user %s already completed onboarding: %w", input.UserID, store.ErrConflict)
		}
		existing.DisplayName = input.DisplayName
		if input.Email != "" {
			existing.Email = input.Email
		}
		existing.PrimaryUnitID = input.PrimaryUnitID
		existing.CurrentUnitID = input.PrimaryUnitID
		if input.IdPSubject != "" {
			existing.IdPSubject = input.IdPSubject
		}
		completedAt := now
		existing.OnboardingCompletedAt = &completedAt
		existing.ReviewStatus = domain.ReviewStatusPendingReview
		existing.UpdatedAt = now
		s.users[input.UserID] = existing
		return existing, nil
	}
	completedAt := now
	user := domain.AppUser{
		ID:                    int64(len(s.users) + 1),
		UserID:                input.UserID,
		Email:                 input.Email,
		DisplayName:           input.DisplayName,
		Role:                  role,
		Status:                domain.UserStatusActive,
		Type:                  domain.UserTypeHuman,
		IdPSubject:            input.IdPSubject,
		PrimaryUnitID:         input.PrimaryUnitID,
		CurrentUnitID:         input.PrimaryUnitID,
		JoinedAt:              now,
		OnboardingCompletedAt: &completedAt,
		ReviewStatus:          domain.ReviewStatusPendingReview,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	s.users[input.UserID] = user
	return user, nil
}

// RM-ONBOARD-01 (API-86). admin 의 review confirm transition.
func (s *memoryOrganizationStore) ConfirmUserReview(_ context.Context, userID string) (domain.AppUser, error) {
	existing, ok := s.users[userID]
	if !ok {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, store.ErrNotFound)
	}
	if existing.OnboardingCompletedAt == nil || existing.ReviewStatus != domain.ReviewStatusPendingReview {
		return domain.AppUser{}, fmt.Errorf("user %s review confirm: %w", userID, store.ErrNotFound)
	}
	existing.ReviewStatus = domain.ReviewStatusReviewed
	existing.UpdatedAt = time.Now().UTC()
	s.users[userID] = existing
	return existing, nil
}

// RM-ONBOARD-01 (API-84). typeahead 검색 — label substring (case-insensitive).
func (s *memoryOrganizationStore) SearchOrgUnits(_ context.Context, q string, limit int) ([]domain.OrgUnit, error) {
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	needle := strings.ToLower(q)
	results := make([]domain.OrgUnit, 0, limit)
	for _, u := range s.units {
		if strings.Contains(strings.ToLower(u.Label), needle) {
			results = append(results, u)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

// helper to fire a request through a fresh router.
func newOrgTestRouter(orgStore OrganizationStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return testRouter(RouterConfig{OrganizationStore: orgStore})
}

func decodeJSON(t *testing.T, body []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode response: %v body=%s", err, string(body))
	}
}

func TestCreateUserInsertsAndRejectsDuplicate(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	router := newOrgTestRouter(storeMem)

	body := []byte(`{
		"user_id": "u100",
		"email": "u100@example.com",
		"display_name": "Hundred",
		"role": "developer",
		"status": "active",
		"primary_unit_id": "team-frontend",
		"joined_at": "2026-05-07"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create user: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			UserID      string `json:"user_id"`
			Role        string `json:"role"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Status != "created" || resp.Data.UserID != "u100" || resp.Data.Role != "developer" || resp.Data.DisplayName != "Hundred" {
		t.Fatalf("unexpected create response: %+v", resp)
	}

	// Duplicate user_id -> 409
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create: expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateUserRejectsInvalidRole(t *testing.T) {
	router := newOrgTestRouter(newMemoryOrganizationStore())
	body := []byte(`{"user_id":"u200","email":"u200@example.com","display_name":"Two","role":"chief","status":"active"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid role, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCreateUserAcceptsPMOManager covers PR #205 codex P1 (sprint -d): after
// removing /rbac/subjects/:id/roles, team_manager must remain assignable via
// POST /api/v1/users.
func TestCreateUserAcceptsPMOManager(t *testing.T) {
	router := newOrgTestRouter(newMemoryOrganizationStore())
	body := []byte(`{
		"user_id": "u-pmo-create",
		"email": "u-pmo-create@example.com",
		"display_name": "PMO Create",
		"role": "team_manager",
		"status": "active",
		"primary_unit_id": "team-frontend",
		"joined_at": "2026-05-20"
	}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user with team_manager: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Role string `json:"role"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Data.Role != "team_manager" {
		t.Errorf("expected role=team_manager, got %q", resp.Data.Role)
	}
}

func TestCreateUserRequiresOrganizationWrite(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	router := NewRouter(RouterConfig{
		OrganizationStore: storeMem,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "developer",
			Role:  "developer",
		}},
	})

	body := []byte(`{
		"user_id": "u201",
		"email": "u201@example.com",
		"display_name": "No Write",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer developer-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateUserAppliesPartialFields(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.users["u300"] = domain.AppUser{
		UserID:      "u300",
		Email:       "u300@example.com",
		DisplayName: "Three",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
	}
	router := newOrgTestRouter(storeMem)

	body := []byte(`{"display_name":"Three Updated","status":"deactivated"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/u300", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update user: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			UserID      string `json:"user_id"`
			DisplayName string `json:"display_name"`
			Status      string `json:"status"`
			Role        string `json:"role"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Data.UserID != "u300" || resp.Data.DisplayName != "Three Updated" || resp.Data.Status != "deactivated" || resp.Data.Role != "developer" {
		t.Fatalf("unexpected update response: %+v", resp.Data)
	}
}

// TestUpdateUserAcceptsPMOManagerRole covers PR #205 codex P1 (sprint -d):
// PATCH /api/v1/users/:id must accept team_manager after /rbac/subjects/:id/roles
// removal. Regression guard for the role allowlist expansion.
func TestUpdateUserAcceptsPMOManagerRole(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.users["u-pmo-update"] = domain.AppUser{
		UserID:      "u-pmo-update",
		Email:       "u-pmo-update@example.com",
		DisplayName: "PMO Update",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
	}
	router := newOrgTestRouter(storeMem)

	body := []byte(`{"role":"team_manager"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/users/u-pmo-update", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("update user role to team_manager: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Role string `json:"role"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Data.Role != "team_manager" {
		t.Errorf("expected role=team_manager, got %q", resp.Data.Role)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	router := newOrgTestRouter(newMemoryOrganizationStore())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/users/u999", bytes.NewReader([]byte(`{"display_name":"x"}`))))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUserRemovesRow(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.users["u400"] = domain.AppUser{UserID: "u400", Role: domain.AppRoleDeveloper, Status: domain.UserStatusActive}
	router := newOrgTestRouter(storeMem)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/users/u400", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete user: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if _, exists := storeMem.users["u400"]; exists {
		t.Fatalf("expected u400 to be removed from store")
	}

	// Second delete -> 404
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/users/u400", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("repeated delete: expected 404, got %d", rec.Code)
	}
}

func TestCreateOrgUnitInsertsAndRejectsBadType(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	router := newOrgTestRouter(storeMem)

	good := []byte(`{"unit_id":"unit-a","unit_type":"team","label":"Team A","position_x":10,"position_y":20}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/organization/units", bytes.NewReader(good)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create unit: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			UnitID    string `json:"unit_id"`
			UnitType  string `json:"unit_type"`
			Label     string `json:"label"`
			PositionX int    `json:"position_x"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Data.UnitID != "unit-a" || resp.Data.UnitType != "team" || resp.Data.Label != "Team A" || resp.Data.PositionX != 10 {
		t.Fatalf("unexpected unit response: %+v", resp.Data)
	}

	bad := []byte(`{"unit_id":"unit-b","unit_type":"squad","label":"Bad"}`)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/organization/units", bytes.NewReader(bad)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad unit_type, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateOrgUnitChangesLabel(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.units["unit-x"] = domain.OrgUnit{UnitID: "unit-x", UnitType: domain.UnitTypeTeam, Label: "Old"}
	router := newOrgTestRouter(storeMem)

	body := []byte(`{"label":"New Label","position_x":42}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/organization/units/unit-x", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("update unit: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if storeMem.units["unit-x"].Label != "New Label" || storeMem.units["unit-x"].PositionX != 42 {
		t.Fatalf("unit not updated: %+v", storeMem.units["unit-x"])
	}
}

func TestDeleteOrgUnitRemovesRow(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.units["unit-z"] = domain.OrgUnit{UnitID: "unit-z", UnitType: domain.UnitTypeTeam, Label: "Z"}
	router := newOrgTestRouter(storeMem)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/organization/units/unit-z", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete unit: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := storeMem.units["unit-z"]; ok {
		t.Fatalf("expected unit-z to be removed")
	}
}

func TestReplaceUnitMembersAppliesAndPreservesLeader(t *testing.T) {
	storeMem := newMemoryOrganizationStore()
	storeMem.units["team-x"] = domain.OrgUnit{UnitID: "team-x", UnitType: domain.UnitTypeTeam, Label: "X"}
	storeMem.users["u1"] = domain.AppUser{UserID: "u1", Role: domain.AppRoleSystemAdmin, Status: domain.UserStatusActive}
	storeMem.users["u2"] = domain.AppUser{UserID: "u2", Role: domain.AppRoleDeveloper, Status: domain.UserStatusActive}
	storeMem.users["u3"] = domain.AppUser{UserID: "u3", Role: domain.AppRoleDeveloper, Status: domain.UserStatusActive}
	storeMem.memberships["team-x"] = map[string]domain.AppointmentRole{
		"u1": domain.AppointmentRoleLeader,
		"u2": domain.AppointmentRoleMember,
	}

	router := newOrgTestRouter(storeMem)
	body := []byte(`{"user_ids":["u3","u3",""]}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/organization/units/team-x/members", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("replace members: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	final := storeMem.memberships["team-x"]
	if final["u1"] != domain.AppointmentRoleLeader {
		t.Fatalf("leader appointment lost: %+v", final)
	}
	if final["u3"] != domain.AppointmentRoleMember {
		t.Fatalf("u3 not added as member: %+v", final)
	}
	if _, ok := final["u2"]; ok {
		t.Fatalf("u2 should have been removed: %+v", final)
	}
}

func TestReplaceUnitMembersUnitNotFound(t *testing.T) {
	router := newOrgTestRouter(newMemoryOrganizationStore())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/v1/organization/units/nope/members", bytes.NewReader([]byte(`{"user_ids":["u1"]}`))))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetOrgUnitNotFoundUsesStoreError(t *testing.T) {
	router := newOrgTestRouter(newMemoryOrganizationStore())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/organization/units/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	// Sanity-check error wraps store.ErrNotFound semantically.
	_, err := newMemoryOrganizationStore().GetOrgUnit(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound from mock store, got %v", err)
	}
}

// TestUpdateOrgUnit_RejectsSelfParent_CycleDetected pins the §10.4.2
// cycle_detected branch added by RM-M3-03 (sprint claude/work_260513-m).
// Updating a unit so its parent_unit_id equals itself must return 422 with
// code=cycle_detected.
func TestUpdateOrgUnit_RejectsSelfParent_CycleDetected(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateOrgUnit(context.Background(), domain.CreateOrgUnitInput{
		UnitID:   "unit-A",
		UnitType: domain.UnitType("team"),
		Label:    "Team A",
	}); err != nil {
		t.Fatalf("seed CreateOrgUnit: %v", err)
	}
	router := testRouter(RouterConfig{OrganizationStore: orgs})

	self := "unit-A"
	body := []byte(`{"parent_unit_id":"unit-A"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organization/units/"+self, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cycle_detected") {
		t.Errorf("expected code=cycle_detected, got body=%q", rec.Body.String())
	}
}

// TestUpdateOrgUnit_RejectsAncestorParent_CycleDetected covers the indirect
// case: with hierarchy A -> B already in place, moving A so its parent is B
// would close the loop (B's ancestor chain reaches A). Must return 422.
func TestUpdateOrgUnit_RejectsAncestorParent_CycleDetected(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateOrgUnit(context.Background(), domain.CreateOrgUnitInput{
		UnitID: "unit-A", UnitType: domain.UnitType("team"), Label: "A",
	}); err != nil {
		t.Fatalf("seed A: %v", err)
	}
	if _, err := orgs.CreateOrgUnit(context.Background(), domain.CreateOrgUnitInput{
		UnitID: "unit-B", ParentUnitID: "unit-A", UnitType: domain.UnitType("team"), Label: "B",
	}); err != nil {
		t.Fatalf("seed B: %v", err)
	}
	router := testRouter(RouterConfig{OrganizationStore: orgs})

	body := []byte(`{"parent_unit_id":"unit-B"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organization/units/unit-A", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cycle_detected") {
		t.Errorf("expected code=cycle_detected, got body=%q", rec.Body.String())
	}
}
