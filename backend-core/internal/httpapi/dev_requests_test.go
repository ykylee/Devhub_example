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
	mu         sync.Mutex
	rows       map[string]domain.DevRequest
	nextID     int
	knownUsers map[string]bool
}

func newMemoryDevRequestStore() *memoryDevRequestStore {
	return &memoryDevRequestStore{
		rows:       map[string]domain.DevRequest{},
		knownUsers: map[string]bool{"alice": true, "bob": true, "charlie": true},
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
