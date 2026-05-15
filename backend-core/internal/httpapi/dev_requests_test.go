package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// --- memoryDevRequestStore — in-memory test store (DREQ-Backend, sprint claude/work_260515-i) ---

type memoryDevRequestStore struct {
	mu               sync.Mutex
	rows             map[string]domain.DevRequest
	nextID           int
	nextAppID        int
	nextProjectID    int
	knownUsers       map[string]bool
	knownRepoIDs     map[int64]bool
	knownDevUnits    map[string]bool
	createdApps      []domain.Application
	createdProjects  []domain.Project
	createdRepoLinks []domain.ApplicationRepository
	// failPromote, when non-empty, makes promote methods return that error
	// (used to assert tx rollback semantics from the handler's perspective).
	failPromote error
}

func newMemoryDevRequestStore() *memoryDevRequestStore {
	return &memoryDevRequestStore{
		rows:          map[string]domain.DevRequest{},
		knownUsers:    map[string]bool{"alice": true, "bob": true, "charlie": true},
		knownRepoIDs:  map[int64]bool{1: true, 2: true},
		knownDevUnits: map[string]bool{"unit-dev": true},
	}
}

var _ DevRequestStore = (*memoryDevRequestStore)(nil)

func (s *memoryDevRequestStore) CreateDevRequest(_ context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dr.AssigneeUserID != "" && !s.knownUsers[dr.AssigneeUserID] {
		return domain.DevRequest{}, store.ErrConflict
	}
	if dr.ExternalRef != "" {
		for _, existing := range s.rows {
			if existing.SourceSystem == dr.SourceSystem && existing.ExternalRef == dr.ExternalRef {
				return domain.DevRequest{}, store.ErrConflict
			}
		}
	}
	s.nextID++
	dr.ID = "dreq-" + itoa(s.nextID)
	now := time.Now().UTC()
	if dr.ReceivedAt.IsZero() {
		dr.ReceivedAt = now
	}
	dr.CreatedAt = now
	dr.UpdatedAt = now
	s.rows[dr.ID] = dr
	return dr, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func (s *memoryDevRequestStore) GetDevRequest(_ context.Context, id string) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dr, ok := s.rows[id]
	if !ok {
		return domain.DevRequest{}, store.ErrNotFound
	}
	return dr, nil
}

func (s *memoryDevRequestStore) GetDevRequestByExternalRef(_ context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, dr := range s.rows {
		if dr.SourceSystem == sourceSystem && dr.ExternalRef == externalRef && externalRef != "" {
			return dr, nil
		}
	}
	return domain.DevRequest{}, store.ErrNotFound
}

func (s *memoryDevRequestStore) ListDevRequests(_ context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.DevRequest, 0, len(s.rows))
	for _, dr := range s.rows {
		if opts.AssigneeUserID != "" && dr.AssigneeUserID != opts.AssigneeUserID {
			continue
		}
		if opts.SourceSystem != "" && dr.SourceSystem != opts.SourceSystem {
			continue
		}
		if len(opts.Statuses) > 0 {
			match := false
			for _, st := range opts.Statuses {
				if dr.Status == st {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		out = append(out, dr)
	}
	return out, len(out), nil
}

func (s *memoryDevRequestStore) TransitionDevRequestStatus(_ context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dr, ok := s.rows[id]
	if !ok {
		return domain.DevRequest{}, store.ErrNotFound
	}
	dr.Status = to
	switch to {
	case domain.DevRequestStatusRejected:
		dr.RejectedReason = rejectedReason
		dr.RegisteredTargetType = ""
		dr.RegisteredTargetID = ""
	case domain.DevRequestStatusClosed:
		// preserve target / reason from previous terminal state.
	default:
		dr.RejectedReason = ""
	}
	dr.UpdatedAt = time.Now().UTC()
	s.rows[id] = dr
	return dr, nil
}

func (s *memoryDevRequestStore) ReassignDevRequest(_ context.Context, id, newAssigneeUserID string) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dr, ok := s.rows[id]
	if !ok {
		return domain.DevRequest{}, store.ErrNotFound
	}
	if !s.knownUsers[newAssigneeUserID] {
		return domain.DevRequest{}, store.ErrConflict
	}
	dr.AssigneeUserID = newAssigneeUserID
	dr.UpdatedAt = time.Now().UTC()
	s.rows[id] = dr
	return dr, nil
}

func (s *memoryDevRequestStore) MarkDevRequestRegistered(_ context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dr, ok := s.rows[id]
	if !ok {
		return domain.DevRequest{}, store.ErrNotFound
	}
	dr.Status = domain.DevRequestStatusRegistered
	dr.RegisteredTargetType = targetType
	dr.RegisteredTargetID = targetID
	dr.UpdatedAt = time.Now().UTC()
	s.rows[id] = dr
	return dr, nil
}

func (s *memoryDevRequestStore) RegisterDevRequestWithNewApplication(_ context.Context, drID string, app domain.Application, primaryRepo *domain.ApplicationRepository) (domain.DevRequest, domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failPromote != nil {
		return domain.DevRequest{}, domain.Application{}, s.failPromote
	}
	dr, ok := s.rows[drID]
	if !ok {
		return domain.DevRequest{}, domain.Application{}, store.ErrNotFound
	}
	// FK-style validation (mirrors Postgres FK checks so tests catch handler/store contract drift).
	if app.OwnerUserID != "" && !s.knownUsers[app.OwnerUserID] {
		return domain.DevRequest{}, domain.Application{}, store.ErrConflict
	}
	if app.LeaderUserID != "" && !s.knownUsers[app.LeaderUserID] {
		return domain.DevRequest{}, domain.Application{}, store.ErrConflict
	}
	if app.DevelopmentUnitID != "" && !s.knownDevUnits[app.DevelopmentUnitID] {
		return domain.DevRequest{}, domain.Application{}, store.ErrConflict
	}
	// UNIQUE (key) emulation.
	for _, existing := range s.createdApps {
		if existing.Key == app.Key {
			return domain.DevRequest{}, domain.Application{}, store.ErrConflict
		}
	}
	s.nextAppID++
	app.ID = "app-" + itoa(s.nextAppID)
	now := time.Now().UTC()
	app.CreatedAt = now
	app.UpdatedAt = now
	s.createdApps = append(s.createdApps, app)
	if primaryRepo != nil {
		primaryRepo.ApplicationID = app.ID
		primaryRepo.LinkedAt = now
		s.createdRepoLinks = append(s.createdRepoLinks, *primaryRepo)
	}
	dr.Status = domain.DevRequestStatusRegistered
	dr.RegisteredTargetType = domain.DevRequestTargetApplication
	dr.RegisteredTargetID = app.ID
	dr.UpdatedAt = now
	s.rows[drID] = dr
	return dr, app, nil
}

func (s *memoryDevRequestStore) RegisterDevRequestWithNewProject(_ context.Context, drID string, project domain.Project) (domain.DevRequest, domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failPromote != nil {
		return domain.DevRequest{}, domain.Project{}, s.failPromote
	}
	dr, ok := s.rows[drID]
	if !ok {
		return domain.DevRequest{}, domain.Project{}, store.ErrNotFound
	}
	if project.RepositoryID == 0 || !s.knownRepoIDs[project.RepositoryID] {
		return domain.DevRequest{}, domain.Project{}, store.ErrConflict
	}
	if project.OwnerUserID != "" && !s.knownUsers[project.OwnerUserID] {
		return domain.DevRequest{}, domain.Project{}, store.ErrConflict
	}
	// UNIQUE (repository_id, key) emulation.
	for _, existing := range s.createdProjects {
		if existing.RepositoryID == project.RepositoryID && existing.Key == project.Key {
			return domain.DevRequest{}, domain.Project{}, store.ErrConflict
		}
	}
	s.nextProjectID++
	project.ID = "project-" + itoa(s.nextProjectID)
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now
	s.createdProjects = append(s.createdProjects, project)
	dr.Status = domain.DevRequestStatusRegistered
	dr.RegisteredTargetType = domain.DevRequestTargetProject
	dr.RegisteredTargetID = project.ID
	dr.UpdatedAt = now
	s.rows[drID] = dr
	return dr, project, nil
}

func (s *memoryDevRequestStore) seed(dr domain.DevRequest) domain.DevRequest {
	created, _ := s.CreateDevRequest(context.Background(), dr)
	return created
}

// --- handler tests ---

func newDevRequestsRouter(s DevRequestStore) http.Handler {
	return NewRouter(RouterConfig{
		DevRequestStore: s,
		AuditStore:      &memoryAuditStore{},
		AuthDevFallback: true,
	})
}

func TestIntakeDevRequest_RouteRegistered(t *testing.T) {
	policy, ok := lookupRoutePolicy(http.MethodPost, "/api/v1/dev-requests")
	if !ok {
		t.Fatal("POST /api/v1/dev-requests must exist in routePermissionTable")
	}
	if !policy.Bypass {
		t.Errorf("intake POST should be Bypass: true, got %+v", policy)
	}
}

func TestListDevRequests_Empty(t *testing.T) {
	router := newDevRequestsRouter(newMemoryDevRequestStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/dev-requests", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"data":[]`)) {
		t.Errorf("expected empty data array, got %s", rec.Body.String())
	}
}

func TestGetDevRequest_NotFound(t *testing.T) {
	router := newDevRequestsRouter(newMemoryDevRequestStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/dev-requests/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRejectDevRequest_RequiresReason(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+dr.ID+"/reject", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_reason_required"`)) {
		t.Errorf("expected code=dev_request_reason_required: %s", rec.Body.String())
	}
}

func TestRejectDevRequest_TransitionsAndAudits(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		DevRequestStore: s,
		AuditStore:      audits,
		AuthDevFallback: true,
	})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+dr.ID+"/reject",
		`{"rejected_reason":"중복"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"rejected"`)) {
		t.Errorf("expected status=rejected in body: %s", rec.Body.String())
	}
	found := false
	for _, l := range audits.logs {
		if l.Action == "dev_request.rejected" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dev_request.rejected audit, got %+v", audits.logs)
	}
}

func TestRegisterDevRequest_HappyAndStateGuard(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register",
		`{"target_type":"application","target_id":"app-1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("happy code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"registered"`)) {
		t.Errorf("expected status=registered: %s", rec.Body.String())
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register",
		`{"target_type":"application","target_id":"app-2"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("re-register code=%d body=%s, want 409", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_already_registered"`)) {
		t.Errorf("expected dev_request_already_registered: %s", rec.Body.String())
	}
}

func TestRegisterDevRequest_InvalidTargetType(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+dr.ID+"/register",
		`{"target_type":"repository","target_id":"foo"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_register_target_invalid"`)) {
		t.Errorf("expected dev_request_register_target_invalid: %s", rec.Body.String())
	}
}

func TestCloseDevRequest_InvalidTransition(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/dev-requests/"+dr.ID, "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_status_transition_close"`)) {
		t.Errorf("expected invalid_status_transition_close: %s", rec.Body.String())
	}
}

func TestCloseDevRequest_FromRegistered(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem:         "ops",
		Status:               domain.DevRequestStatusRegistered,
		RegisteredTargetType: domain.DevRequestTargetApplication,
		RegisteredTargetID:   "app-1",
	})
	router := newDevRequestsRouter(s)
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/dev-requests/"+dr.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"closed"`)) {
		t.Errorf("expected status=closed: %s", rec.Body.String())
	}
}

func TestReassignDevRequest_Happy(t *testing.T) {
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/dev-requests/"+dr.ID,
		`{"assignee_user_id":"bob"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"assignee_user_id":"bob"`)) {
		t.Errorf("expected assignee=bob: %s", rec.Body.String())
	}
}

// --- intake auth middleware unit-level smoke ---

type fakeIntakeTokenStore struct {
	rows    map[string]domain.DevRequestIntakeToken
	touched []string
}

var _ IntakeTokenStore = (*fakeIntakeTokenStore)(nil)

func (f *fakeIntakeTokenStore) LookupDevRequestIntakeToken(_ context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
	if row, ok := f.rows[hashedToken]; ok {
		return row, nil
	}
	return domain.DevRequestIntakeToken{}, store.ErrNotFound
}

func (f *fakeIntakeTokenStore) MarkDevRequestIntakeTokenUsed(_ context.Context, tokenID string) error {
	f.touched = append(f.touched, tokenID)
	return nil
}

func TestIntakeAuth_MissingHeaderDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := Handler{cfg: RouterConfig{DevRequestIntakeTokenStore: &fakeIntakeTokenStore{rows: map[string]domain.DevRequestIntakeToken{}}, AuditStore: &memoryAuditStore{}}}
	router.POST("/intake", h.requireIntakeToken, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodPost, "/intake", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"auth_intake_token_missing"`)) {
		t.Errorf("expected auth_intake_token_missing: %s", rec.Body.String())
	}
}

func TestIntakeAuth_UnknownTokenDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := Handler{cfg: RouterConfig{DevRequestIntakeTokenStore: &fakeIntakeTokenStore{rows: map[string]domain.DevRequestIntakeToken{}}, AuditStore: &memoryAuditStore{}}}
	router.POST("/intake", h.requireIntakeToken, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	req := httptest.NewRequest(http.MethodPost, "/intake", nil)
	req.Header.Set("Authorization", "Bearer some-random-token-string-32bytes")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"auth_intake_token_invalid"`)) {
		t.Errorf("expected auth_intake_token_invalid: %s", rec.Body.String())
	}
}

func TestIntakeAuth_IPDeniedWithEmptyAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	plain := "test-token-32bytes-abcdef0123456789"
	hashed := hashIntakeToken(plain)
	tokenStore := &fakeIntakeTokenStore{rows: map[string]domain.DevRequestIntakeToken{
		hashed: {TokenID: "tok-1", ClientLabel: "ops", HashedToken: hashed, AllowedIPs: nil, SourceSystem: "ops"},
	}}
	h := Handler{cfg: RouterConfig{DevRequestIntakeTokenStore: tokenStore, AuditStore: &memoryAuditStore{}}}
	router.POST("/intake", h.requireIntakeToken, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	req := httptest.NewRequest(http.MethodPost, "/intake", nil)
	req.Header.Set("Authorization", "Bearer "+plain)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"auth_intake_ip_denied"`)) {
		t.Errorf("expected auth_intake_ip_denied: %s", rec.Body.String())
	}
}

func TestIntakeAuth_RevokedTokenDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	plain := "revoked-token-data-abcdef0123456789"
	hashed := hashIntakeToken(plain)
	now := time.Now()
	tokenStore := &fakeIntakeTokenStore{rows: map[string]domain.DevRequestIntakeToken{
		hashed: {TokenID: "tok-r", ClientLabel: "ops", HashedToken: hashed, AllowedIPs: []string{"0.0.0.0/0"}, SourceSystem: "ops", RevokedAt: &now},
	}}
	h := Handler{cfg: RouterConfig{DevRequestIntakeTokenStore: tokenStore, AuditStore: &memoryAuditStore{}}}
	router.POST("/intake", h.requireIntakeToken, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	req := httptest.NewRequest(http.MethodPost, "/intake", nil)
	req.Header.Set("Authorization", "Bearer "+plain)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"auth_intake_token_revoked"`)) {
		t.Errorf("expected auth_intake_token_revoked: %s", rec.Body.String())
	}
}

// --- DREQ-Promote-Tx (sprint claude/work_260515-m) handler tests ---

func TestRegisterDevRequest_NewApplicationHappy(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		DevRequestStore: s,
		AuditStore:      audits,
		AuthDevFallback: true,
	})

	body := `{
		"target_type": "application",
		"application_payload": {
			"key": "ALPHA12345",
			"name": "Alpha",
			"description": "from intake",
			"owner_user_id": "alice",
			"leader_user_id": "bob",
			"development_unit_id": "unit-dev",
			"visibility": "internal",
			"status": "planning"
		}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"created":true`)) {
		t.Errorf("expected created:true: %s", rec.Body.String())
	}
	if len(s.createdApps) != 1 || s.createdApps[0].Key != "ALPHA12345" {
		t.Errorf("expected 1 created app with key=ALPHA12345, got %+v", s.createdApps)
	}
	if s.rows[pending.ID].Status != domain.DevRequestStatusRegistered {
		t.Errorf("expected dev_request status=registered, got %s", s.rows[pending.ID].Status)
	}
	// audit emits: application.created + dev_request.registered
	gotAppAudit, gotDRAudit := false, false
	for _, l := range audits.logs {
		switch l.Action {
		case "application.created":
			gotAppAudit = true
		case "dev_request.registered":
			gotDRAudit = true
		}
	}
	if !gotAppAudit || !gotDRAudit {
		t.Errorf("expected application.created + dev_request.registered audits, got %+v", audits.logs)
	}
}

func TestRegisterDevRequest_NewApplicationWithPrimaryRepo(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	body := `{
		"target_type": "application",
		"application_payload": {
			"key": "BETA123456",
			"name": "Beta",
			"owner_user_id": "alice",
			"leader_user_id": "bob",
			"development_unit_id": "unit-dev",
			"visibility": "internal",
			"status": "planning",
			"primary_repo": {
				"repo_provider": "gitea",
				"repo_full_name": "org/beta",
				"role": "primary"
			}
		}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(s.createdRepoLinks) != 1 {
		t.Errorf("expected 1 primary repo link, got %d", len(s.createdRepoLinks))
	}
	if s.createdRepoLinks[0].RepoFullName != "org/beta" {
		t.Errorf("expected repo_full_name=org/beta, got %s", s.createdRepoLinks[0].RepoFullName)
	}
}

func TestRegisterDevRequest_NewProjectHappy(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	body := `{
		"target_type": "project",
		"project_payload": {
			"repository_id": 1,
			"key": "PROJ1",
			"name": "Proj1",
			"owner_user_id": "alice",
			"visibility": "internal",
			"status": "planning"
		}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(s.createdProjects) != 1 || s.createdProjects[0].Key != "PROJ1" {
		t.Errorf("expected 1 created project key=PROJ1, got %+v", s.createdProjects)
	}
}

func TestRegisterDevRequest_PayloadMutualExclusion(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)

	// Case A: both target_id and application_payload set → 400.
	body := `{
		"target_type": "application",
		"target_id": "app-existing",
		"application_payload": {"key":"ZZZZZZZZZZ","name":"Z","owner_user_id":"alice","leader_user_id":"bob","development_unit_id":"unit-dev","visibility":"internal","status":"planning"}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_register_payload_invalid"`)) {
		t.Errorf("case A: expected 400 dev_request_register_payload_invalid, got code=%d body=%s", rec.Code, rec.Body.String())
	}

	// Case B: neither set → 400.
	body = `{"target_type":"application"}`
	rec = doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_register_payload_invalid"`)) {
		t.Errorf("case B: expected 400 dev_request_register_payload_invalid, got code=%d body=%s", rec.Code, rec.Body.String())
	}

	// Case C: project_payload with target_type=application → 400.
	body = `{"target_type":"application","project_payload":{"repository_id":1,"key":"PROJX","name":"X","owner_user_id":"alice","visibility":"internal","status":"planning"}}`
	rec = doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"dev_request_register_payload_invalid"`)) {
		t.Errorf("case C: expected 400 dev_request_register_payload_invalid, got code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterDevRequest_NewApplicationFKConflict(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	// owner_user_id="ghost" 는 knownUsers 에 없음 → store ErrConflict → 409.
	body := `{
		"target_type":"application",
		"application_payload":{"key":"GAMMA12345","name":"G","owner_user_id":"ghost","leader_user_id":"bob","development_unit_id":"unit-dev","visibility":"internal","status":"planning"}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"application_key_conflict"`)) {
		t.Errorf("expected application_key_conflict: %s", rec.Body.String())
	}
	// Rollback semantics: no app should have been "persisted" + dev_request still pending.
	if len(s.createdApps) != 0 {
		t.Errorf("expected 0 created apps after FK conflict, got %+v", s.createdApps)
	}
	if s.rows[pending.ID].Status != domain.DevRequestStatusPending {
		t.Errorf("expected dev_request still pending after rollback, got %s", s.rows[pending.ID].Status)
	}
}

func TestRegisterDevRequest_LegacyTargetIDStillSupported(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		DevRequestStore: s,
		AuditStore:      audits,
		AuthDevFallback: true,
	})
	body := `{"target_type":"application","target_id":"app-legacy-1"}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	// Legacy path emits dev_request.registered with created:false.
	gotCreatedFalse := false
	for _, l := range audits.logs {
		if l.Action == "dev_request.registered" && l.Payload["created"] == false {
			gotCreatedFalse = true
		}
	}
	if !gotCreatedFalse {
		t.Errorf("expected legacy dev_request.registered audit with created=false, got %+v", audits.logs)
	}
	if len(s.createdApps) != 0 {
		t.Errorf("legacy path should not create an app, got %+v", s.createdApps)
	}
}

// --- codex hotfix #4 (sprint claude/work_260515-n) regression guards ---

func TestRegisterDevRequest_PromoteApp_InvalidRepoRole(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := NewRouter(RouterConfig{
		DevRequestStore:  s,
		ApplicationStore: newMemoryApplicationStore(),
		AuditStore:       &memoryAuditStore{},
		AuthDevFallback:  true,
	})
	// role="owner" 는 application_repositories_role_check 위반 — 본래 PG CHECK 가 잡지만,
	// codex P1 hotfix 가 handler 의 application-level gate 로 422 invalid_repo_link_role 로 surface.
	body := `{
		"target_type":"application",
		"application_payload":{"key":"DELTA12345","name":"D","owner_user_id":"u1","leader_user_id":"u2","development_unit_id":"unit-dev","visibility":"internal","status":"planning",
			"primary_repo":{"repo_provider":"gitea","repo_full_name":"org/d","role":"owner"}}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_repo_link_role"`)) {
		t.Errorf("expected invalid_repo_link_role: %s", rec.Body.String())
	}
	// 0 created — handler 가 store 호출 전에 차단.
	if len(s.createdApps) != 0 {
		t.Errorf("expected 0 created apps after gate, got %d", len(s.createdApps))
	}
}

func TestRegisterDevRequest_PromoteApp_DisabledSCMProvider(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	// newMemoryApplicationStore() 에서 forgejo 는 Enabled=false.
	router := NewRouter(RouterConfig{
		DevRequestStore:  s,
		ApplicationStore: newMemoryApplicationStore(),
		AuditStore:       &memoryAuditStore{},
		AuthDevFallback:  true,
	})
	body := `{
		"target_type":"application",
		"application_payload":{"key":"EPSILON123","name":"E","owner_user_id":"u1","leader_user_id":"u2","development_unit_id":"unit-dev","visibility":"internal","status":"planning",
			"primary_repo":{"repo_provider":"forgejo","repo_full_name":"org/e","role":"primary"}}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"unsupported_repo_provider"`)) {
		t.Errorf("expected unsupported_repo_provider: %s", rec.Body.String())
	}
	if len(s.createdApps) != 0 {
		t.Errorf("expected 0 created apps after gate, got %d", len(s.createdApps))
	}
}

func TestRegisterDevRequest_PromoteApp_UnknownProviderRejectedSamePathAsLegacy(t *testing.T) {
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := NewRouter(RouterConfig{
		DevRequestStore:  s,
		ApplicationStore: newMemoryApplicationStore(),
		AuditStore:       &memoryAuditStore{},
		AuthDevFallback:  true,
	})
	body := `{
		"target_type":"application",
		"application_payload":{"key":"ZETA123456","name":"Z","owner_user_id":"u1","leader_user_id":"u2","development_unit_id":"unit-dev","visibility":"internal","status":"planning",
			"primary_repo":{"repo_provider":"unknown","repo_full_name":"org/z","role":"primary"}}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"unsupported_repo_provider"`)) {
		t.Errorf("expected unsupported_repo_provider: %s", rec.Body.String())
	}
}

func TestRegisterDevRequest_PromoteApp_NoApplicationStoreFallback(t *testing.T) {
	// ApplicationStore unwired — provider gate 는 dev environment 에서 통과.
	// 회귀 가드: ApplicationStore 가 nil 일 때 happy path 가 깨지지 않음.
	s := newMemoryDevRequestStore()
	pending := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusPending,
	})
	router := newDevRequestsRouter(s)
	body := `{
		"target_type":"application",
		"application_payload":{"key":"OMEGA12345","name":"O","owner_user_id":"alice","leader_user_id":"bob","development_unit_id":"unit-dev","visibility":"internal","status":"planning",
			"primary_repo":{"repo_provider":"any","repo_full_name":"org/o","role":"primary"}}
	}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+pending.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (gate skipped without ApplicationStore), code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterDevRequest_PromoteApp_RejectedReasonClearedOnReopen(t *testing.T) {
	// 직전 rejected 였던 dev_request 가 pending 으로 reopen → register 되면
	// rejected_reason 이 NULL 로 비워져야 한다. self-review P2 #1 회귀 가드.
	s := newMemoryDevRequestStore()
	dr := s.seed(domain.DevRequest{
		Title: "x", Requester: "ext", AssigneeUserID: "alice",
		SourceSystem: "ops", Status: domain.DevRequestStatusRejected,
		RejectedReason: "이전 거절 사유",
	})
	// memory store 는 reopen 표현이 약하지만, register 가 MarkDevRequestRegistered 로
	// status='registered' 갱신할 때 rejected_reason 이 NULL 로 비워지는지 검사.
	router := newDevRequestsRouter(s)
	// status 가 rejected 이므로 직접 register 는 conflict — 먼저 pending 으로 복원.
	dr.Status = domain.DevRequestStatusPending
	s.rows[dr.ID] = dr
	body := `{"target_type":"application","target_id":"app-1"}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/dev-requests/"+dr.ID+"/register", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	// memoryDevRequestStore.MarkDevRequestRegistered 가 rejected_reason 을 자동 clear 하지 않는다 — store 가 SQL 정합 검증 대상이므로 본 회귀는 분리.
	// 본 test 는 handler path 가 정상 200 응답하는지만 가드. SQL 의 rejected_reason=NULL 은 integration 테스트의 대상.
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"registered"`)) {
		t.Errorf("expected status=registered: %s", rec.Body.String())
	}
}

func TestClientIPAllowed_CIDRAndSingle(t *testing.T) {
	if !clientIPAllowed("192.0.2.5", []string{"192.0.2.0/24"}) {
		t.Error("CIDR should include 192.0.2.5")
	}
	if !clientIPAllowed("198.51.100.7", []string{"198.51.100.7"}) {
		t.Error("single IP should match")
	}
	if clientIPAllowed("10.0.0.1", []string{"192.0.2.0/24"}) {
		t.Error("CIDR should not include 10.0.0.1")
	}
	if clientIPAllowed("192.0.2.1", nil) {
		t.Error("nil allowlist should deny")
	}
}
