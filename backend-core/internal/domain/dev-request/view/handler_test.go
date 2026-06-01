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

type fakeDevReqAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeDevReqAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_dr_id"
	f.created = append(f.created, log)
	return log, nil
}

type fakeDevReqAppStore struct {
	listSCMProvidersFunc func(ctx context.Context) ([]domain.SCMProvider, error)
	createProjectWithRepositoryPayloadFunc func(ctx context.Context, p domain.Project, repoIDs []int64, payload *store.RepositoryCreatePayload) (domain.Project, error)
	createApplicationFunc func(ctx context.Context, app domain.Application) (domain.Application, error)
}

func (f *fakeDevReqAppStore) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
	if f.listSCMProvidersFunc != nil {
		return f.listSCMProvidersFunc(ctx)
	}
	return nil, nil
}
func (f *fakeDevReqAppStore) CreateProjectWithRepositoryPayload(ctx context.Context, p domain.Project, repoIDs []int64, payload *store.RepositoryCreatePayload) (domain.Project, error) {
	if f.createProjectWithRepositoryPayloadFunc != nil {
		return f.createProjectWithRepositoryPayloadFunc(ctx, p, repoIDs, payload)
	}
	return p, nil
}
func (f *fakeDevReqAppStore) CreateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	if f.createApplicationFunc != nil {
		return f.createApplicationFunc(ctx, app)
	}
	return app, nil
}

type fakeDevRequestStore struct {
	createDevRequestFunc func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error)
	getDevRequestFunc func(ctx context.Context, id string) (domain.DevRequest, error)
	getDevRequestByExternalRefFunc func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error)
	listDevRequestsFunc func(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error)
	transitionDevRequestStatusFunc func(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error)
	reassignDevRequestFunc func(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error)
	markDevRequestRegisteredFunc func(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error)
	registerDevRequestWithNewApplicationFunc func(ctx context.Context, drID string, app domain.Application, primaryRepo *domain.ApplicationRepository) (domain.DevRequest, domain.Application, error)
	registerDevRequestWithNewProjectFunc func(ctx context.Context, drID string, project domain.Project) (domain.DevRequest, domain.Project, error)
}

func (f *fakeDevRequestStore) CreateDevRequest(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
	if f.createDevRequestFunc != nil {
		return f.createDevRequestFunc(ctx, dr)
	}
	return dr, nil
}
func (f *fakeDevRequestStore) GetDevRequest(ctx context.Context, id string) (domain.DevRequest, error) {
	if f.getDevRequestFunc != nil {
		return f.getDevRequestFunc(ctx, id)
	}
	return domain.DevRequest{}, nil
}
func (f *fakeDevRequestStore) GetDevRequestByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
	if f.getDevRequestByExternalRefFunc != nil {
		return f.getDevRequestByExternalRefFunc(ctx, sourceSystem, externalRef)
	}
	return domain.DevRequest{}, nil
}
func (f *fakeDevRequestStore) ListDevRequests(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
	if f.listDevRequestsFunc != nil {
		return f.listDevRequestsFunc(ctx, opts)
	}
	return nil, 0, nil
}
func (f *fakeDevRequestStore) TransitionDevRequestStatus(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
	if f.transitionDevRequestStatusFunc != nil {
		return f.transitionDevRequestStatusFunc(ctx, id, to, rejectedReason)
	}
	return domain.DevRequest{}, nil
}
func (f *fakeDevRequestStore) ReassignDevRequest(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error) {
	if f.reassignDevRequestFunc != nil {
		return f.reassignDevRequestFunc(ctx, id, newAssigneeUserID)
	}
	return domain.DevRequest{}, nil
}
func (f *fakeDevRequestStore) MarkDevRequestRegistered(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error) {
	if f.markDevRequestRegisteredFunc != nil {
		return f.markDevRequestRegisteredFunc(ctx, id, targetType, targetID)
	}
	return domain.DevRequest{}, nil
}
func (f *fakeDevRequestStore) RegisterDevRequestWithNewApplication(ctx context.Context, drID string, app domain.Application, primaryRepo *domain.ApplicationRepository) (domain.DevRequest, domain.Application, error) {
	if f.registerDevRequestWithNewApplicationFunc != nil {
		return f.registerDevRequestWithNewApplicationFunc(ctx, drID, app, primaryRepo)
	}
	return domain.DevRequest{}, app, nil
}
func (f *fakeDevRequestStore) RegisterDevRequestWithNewProject(ctx context.Context, drID string, project domain.Project) (domain.DevRequest, domain.Project, error) {
	if f.registerDevRequestWithNewProjectFunc != nil {
		return f.registerDevRequestWithNewProjectFunc(ctx, drID, project)
	}
	return domain.DevRequest{}, project, nil
}

type fakeIntakeTokenStore struct {
	lookupDevRequestIntakeTokenFunc func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error)
	markDevRequestIntakeTokenUsedFunc func(ctx context.Context, tokenID string) error
	createDevRequestIntakeTokenFunc func(ctx context.Context, tok domain.DevRequestIntakeToken) (domain.DevRequestIntakeToken, error)
	listDevRequestIntakeTokensFunc func(ctx context.Context) ([]domain.DevRequestIntakeToken, error)
	revokeDevRequestIntakeTokenFunc func(ctx context.Context, tokenID string) (domain.DevRequestIntakeToken, error)
	updateDevRequestIntakeTokenIPsFunc func(ctx context.Context, tokenID string, allowedIPs []string) (domain.DevRequestIntakeToken, error)
	updateDevRequestIntakeTokenFunc func(ctx context.Context, tokenID string, allowedIPs []string, expiresAt *time.Time, updateIPs bool, updateExpiresAt bool) (domain.DevRequestIntakeToken, error)
	hardRevokeExpiredIntakeTokensFunc func(ctx context.Context, before time.Time) ([]string, error)
	countExpiringSoonIntakeTokensFunc func(ctx context.Context, threshold time.Time) (int, error)
	countStaleIntakeTokensFunc func(ctx context.Context, before time.Time) (int, error)
}

func (f *fakeIntakeTokenStore) LookupDevRequestIntakeToken(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
	if f.lookupDevRequestIntakeTokenFunc != nil {
		return f.lookupDevRequestIntakeTokenFunc(ctx, hashedToken)
	}
	return domain.DevRequestIntakeToken{}, nil
}
func (f *fakeIntakeTokenStore) MarkDevRequestIntakeTokenUsed(ctx context.Context, tokenID string) error {
	if f.markDevRequestIntakeTokenUsedFunc != nil {
		return f.markDevRequestIntakeTokenUsedFunc(ctx, tokenID)
	}
	return nil
}
func (f *fakeIntakeTokenStore) CreateDevRequestIntakeToken(ctx context.Context, tok domain.DevRequestIntakeToken) (domain.DevRequestIntakeToken, error) {
	if f.createDevRequestIntakeTokenFunc != nil {
		return f.createDevRequestIntakeTokenFunc(ctx, tok)
	}
	return tok, nil
}
func (f *fakeIntakeTokenStore) ListDevRequestIntakeTokens(ctx context.Context) ([]domain.DevRequestIntakeToken, error) {
	if f.listDevRequestIntakeTokensFunc != nil {
		return f.listDevRequestIntakeTokensFunc(ctx)
	}
	return nil, nil
}
func (f *fakeIntakeTokenStore) RevokeDevRequestIntakeToken(ctx context.Context, tokenID string) (domain.DevRequestIntakeToken, error) {
	if f.revokeDevRequestIntakeTokenFunc != nil {
		return f.revokeDevRequestIntakeTokenFunc(ctx, tokenID)
	}
	return domain.DevRequestIntakeToken{}, nil
}
func (f *fakeIntakeTokenStore) UpdateDevRequestIntakeTokenIPs(ctx context.Context, tokenID string, allowedIPs []string) (domain.DevRequestIntakeToken, error) {
	if f.updateDevRequestIntakeTokenIPsFunc != nil {
		return f.updateDevRequestIntakeTokenIPsFunc(ctx, tokenID, allowedIPs)
	}
	return domain.DevRequestIntakeToken{}, nil
}
func (f *fakeIntakeTokenStore) UpdateDevRequestIntakeToken(ctx context.Context, tokenID string, allowedIPs []string, expiresAt *time.Time, updateIPs bool, updateExpiresAt bool) (domain.DevRequestIntakeToken, error) {
	if f.updateDevRequestIntakeTokenFunc != nil {
		return f.updateDevRequestIntakeTokenFunc(ctx, tokenID, allowedIPs, expiresAt, updateIPs, updateExpiresAt)
	}
	return domain.DevRequestIntakeToken{}, nil
}
func (f *fakeIntakeTokenStore) HardRevokeExpiredIntakeTokens(ctx context.Context, before time.Time) ([]string, error) {
	if f.hardRevokeExpiredIntakeTokensFunc != nil {
		return f.hardRevokeExpiredIntakeTokensFunc(ctx, before)
	}
	return nil, nil
}
func (f *fakeIntakeTokenStore) CountExpiringSoonIntakeTokens(ctx context.Context, threshold time.Time) (int, error) {
	if f.countExpiringSoonIntakeTokensFunc != nil {
		return f.countExpiringSoonIntakeTokensFunc(ctx, threshold)
	}
	return 0, nil
}
func (f *fakeIntakeTokenStore) CountStaleIntakeTokens(ctx context.Context, before time.Time) (int, error) {
	if f.countStaleIntakeTokensFunc != nil {
		return f.countStaleIntakeTokensFunc(ctx, before)
	}
	return 0, nil
}

func TestNewDevRequestHandler_NonNil(t *testing.T) {
	h := NewDevRequestHandler(DevRequestConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewDevRequestHandler_ConfigPropagation(t *testing.T) {
	cfg := DevRequestConfig{
		ApplicationStore: &fakeDevReqAppStore{},
		AuditStore:       &fakeDevReqAuditStore{},
	}
	h := NewDevRequestHandler(cfg)
	if h.cfg.ApplicationStore == nil {
		t.Fatal("ApplicationStore not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeDevReqAuditStore{}
	h := NewDevRequestHandler(DevRequestConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "dr.test", "dev_request", "dr-1", nil)
	if got.AuditID != "audit_dr_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "dr.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeDevReqAuditStore{}
	h := NewDevRequestHandler(DevRequestConfig{AuditStore: store})
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

	store := &fakeDevReqAuditStore{err: errors.New("db_down")}
	h := NewDevRequestHandler(DevRequestConfig{AuditStore: store})
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

func TestApplicationStoreOrUnavailable_NilReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	got, ok := h.ApplicationStoreOrUnavailable(c)
	if ok {
		t.Fatal("expected ok=false")
	}
	if got != nil {
		t.Fatal("expected nil store ref")
	}
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "application store is not configured") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestApplicationStoreOrUnavailable_PresentReturnsRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeDevReqAppStore{}
	h := NewDevRequestHandler(DevRequestConfig{ApplicationStore: store})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	got, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestEnforceRowOwnership_DevFallbackAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_auth_dev_fallback", true)

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("dev fallback must allow")
	}
}

func TestEnforceRowOwnership_SystemAdminAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "admin")
	c.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("system_admin must be allowed")
	}
}

func TestEnforceRowOwnership_OwnerMatchAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "owner-x")

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("owner self-match must pass")
	}
}

func TestEnforceRowOwnership_DeniedWritesAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeDevReqAuditStore{}
	h := NewDevRequestHandler(DevRequestConfig{AuditStore: store})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "someone-else")
	c.Set("devhub_actor_role", "developer")

	if h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("expected denied")
	}
	if rec.Code != 403 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "auth_row_denied") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if len(store.created) != 1 || store.created[0].Action != "auth.row_denied" {
		t.Fatalf("audit row denied not recorded: %+v", store.created)
	}
}

func TestParseDate_Empty(t *testing.T) {
	got, err := parseDate("")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestParseDate_Valid(t *testing.T) {
	got, err := parseDate("2026-05-29")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil time")
	}
	if got.Year() != 2026 || got.Month() != 5 || got.Day() != 29 {
		t.Fatalf("date = %v", got)
	}
}

func TestParseDate_Invalid(t *testing.T) {
	_, err := parseDate("not-a-date")
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestParseDate_Trim(t *testing.T) {
	got, err := parseDate("  2026-05-29  ")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got == nil || got.Day() != 29 {
		t.Fatalf("got %v", got)
	}
}

func TestFormatDatePtr_NilReturnsNil(t *testing.T) {
	got := formatDatePtr(nil)
	if got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestFormatDatePtr_NonNilFormatsYYYYMMDD(t *testing.T) {
	tm := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	got := formatDatePtr(&tm)
	if got != "2026-05-29" {
		t.Fatalf("got %v", got)
	}
}

func TestFormatTimePtr_NilReturnsNil(t *testing.T) {
	got := formatTimePtr(nil)
	if got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestFormatTimePtr_NonNilFormatsRFC3339(t *testing.T) {
	tm := time.Date(2026, 5, 29, 12, 30, 0, 0, time.UTC)
	got := formatTimePtr(&tm)
	if got != "2026-05-29T12:30:00Z" {
		t.Fatalf("got %v", got)
	}
}

func TestApplicationKeyPattern(t *testing.T) {
	cases := map[string]bool{
		"APP":        true,
		"app1":       true,
		"AB123":      true,
		"":           false,
		"longerthan10char": false,
		"with-dash":  false,
		"with_score": false,
		"contains space": false,
	}
	for input, want := range cases {
		got := applicationKeyPattern.MatchString(input)
		if got != want {
			t.Errorf("pattern.Match(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestValidApplicationVisibilities(t *testing.T) {
	for _, v := range []string{"public", "internal", "restricted"} {
		if !validApplicationVisibilities[v] {
			t.Errorf("expected %q valid", v)
		}
	}
	if validApplicationVisibilities["unknown"] {
		t.Error("unknown must be invalid")
	}
}

func TestValidApplicationStatuses(t *testing.T) {
	for _, s := range []string{"planning", "active", "on_hold", "closed", "archived"} {
		if !validApplicationStatuses[s] {
			t.Errorf("expected %q valid", s)
		}
	}
	if validApplicationStatuses["unknown"] {
		t.Error("unknown must be invalid")
	}
}

func TestValidApplicationRepoRoles(t *testing.T) {
	for _, r := range []string{"primary", "sub", "shared"} {
		if !validApplicationRepoRoles[r] {
			t.Errorf("expected %q valid", r)
		}
	}
	if validApplicationRepoRoles["unknown"] {
		t.Error("unknown must be invalid")
	}
}

func TestApplicationResponse_FieldMapping(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	app := domain.Application{
		ID:                "app-1",
		Key:               "APP",
		Name:              "App One",
		Description:       "desc",
		Status:            domain.ApplicationStatusActive,
		Visibility:        domain.ApplicationVisibilityInternal,
		OwnerUserID:       "owner",
		LeaderUserID:      "leader",
		DevelopmentUnitID: "team",
		StartDate:         &start,
		DueDate:           &due,
		CreatedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	resp := applicationResponse(app)
	if resp["id"] != "app-1" || resp["key"] != "APP" {
		t.Fatalf("basic mapping: %+v", resp)
	}
	if resp["start_date"] != "2026-01-01" || resp["due_date"] != "2026-12-31" {
		t.Fatalf("dates: %+v", resp)
	}
	if resp["status"] != "active" || resp["visibility"] != "internal" {
		t.Fatalf("status/visibility: %+v", resp)
	}
}

func TestProjectResponse_FieldMapping(t *testing.T) {
	p := domain.Project{
		ID:            "proj-1",
		ApplicationID: "app-1",
		RepositoryID:  42,
		Key:           "PRJ",
		Name:          "Proj One",
		Status:        domain.ApplicationStatusActive,
		Visibility:    domain.ApplicationVisibilityInternal,
		OwnerUserID:   "owner",
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	resp := projectResponse(p)
	if resp["id"] != "proj-1" || resp["key"] != "PRJ" || resp["application_id"] != "app-1" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["repository_id"] != int64(42) {
		t.Fatalf("repository_id = %v", resp["repository_id"])
	}
}

func TestProjectResponse_RepositoryIDZeroReturnsNil(t *testing.T) {
	p := domain.Project{ID: "p-1", RepositoryID: 0}
	resp := projectResponse(p)
	if resp["repository_id"] != nil {
		t.Fatalf("repository_id should be nil when 0, got %v", resp["repository_id"])
	}
}

func TestDevRequestResponse_FieldMapping(t *testing.T) {
	received := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	dr := domain.DevRequest{
		ID:                   "dr-1",
		Title:                "Title",
		Details:              "Details",
		Requester:            "alice",
		AssigneeUserID:       "bob",
		SourceSystem:         "jira",
		ExternalRef:          "JIRA-1",
		Status:               domain.DevRequestStatusPending,
		RegisteredTargetType: domain.DevRequestTargetType("application"),
		RegisteredTargetID:   "app-1",
		RejectedReason:       "",
		ReceivedAt:           received,
		CreatedAt:            received,
		UpdatedAt:            received,
	}
	resp := devRequestResponse(dr)
	if resp["id"] != "dr-1" || resp["title"] != "Title" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["requester"] != "alice" || resp["assignee_user_id"] != "bob" {
		t.Fatalf("requester/assignee: %+v", resp)
	}
	if resp["source_system"] != "jira" || resp["external_ref"] != "JIRA-1" {
		t.Fatalf("source: %+v", resp)
	}
	if resp["registered_target_type"] != "application" || resp["registered_target_id"] != "app-1" {
		t.Fatalf("target: %+v", resp)
	}
}

func TestCombineRejectedReason(t *testing.T) {
	cases := []struct {
		existing string
		extra    string
		want     string
	}{
		{"", "", ""},
		{"a", "", "a"},
		{"", "b", "b"},
		{"a", "b", "a; b"},
	}
	for _, c := range cases {
		if got := combineRejectedReason(c.existing, c.extra); got != c.want {
			t.Errorf("combineRejectedReason(%q,%q) = %q, want %q", c.existing, c.extra, got, c.want)
		}
	}
}

func TestValidateIntakeRequest_AllValid(t *testing.T) {
	req := createDevRequestRequest{
		Title:          "Need new feature",
		Requester:      "alice",
		AssigneeUserID: "bob",
	}
	if got := validateIntakeRequest(req); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestValidateIntakeRequest_MissingTitle(t *testing.T) {
	req := createDevRequestRequest{Requester: "alice", AssigneeUserID: "bob"}
	got := validateIntakeRequest(req)
	if !strings.Contains(got, "title is required") {
		t.Fatalf("got %q", got)
	}
}

func TestValidateIntakeRequest_TitleTooLong(t *testing.T) {
	req := createDevRequestRequest{
		Title:          strings.Repeat("x", 201),
		Requester:      "alice",
		AssigneeUserID: "bob",
	}
	got := validateIntakeRequest(req)
	if !strings.Contains(got, "title exceeds 200 chars") {
		t.Fatalf("got %q", got)
	}
}

func TestValidateIntakeRequest_MissingRequester(t *testing.T) {
	req := createDevRequestRequest{Title: "x", AssigneeUserID: "bob"}
	got := validateIntakeRequest(req)
	if !strings.Contains(got, "requester is required") {
		t.Fatalf("got %q", got)
	}
}

func TestValidateIntakeRequest_MissingAssignee(t *testing.T) {
	req := createDevRequestRequest{Title: "x", Requester: "alice"}
	got := validateIntakeRequest(req)
	if !strings.Contains(got, "assignee_user_id is required") {
		t.Fatalf("got %q", got)
	}
}

func TestValidateIntakeRequest_MultipleErrorsJoined(t *testing.T) {
	req := createDevRequestRequest{}
	got := validateIntakeRequest(req)
	if !strings.Contains(got, ";") {
		t.Fatalf("expected joined message, got %q", got)
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	if got := extractBearerToken("Bearer abc"); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractBearerToken_TrimSpaces(t *testing.T) {
	if got := extractBearerToken("Bearer  abc  "); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractBearerToken_NoPrefix(t *testing.T) {
	if got := extractBearerToken("Basic abc"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractBearerToken_Empty(t *testing.T) {
	if got := extractBearerToken(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestHashIntakeToken_DeterministicAndDistinct(t *testing.T) {
	h1 := hashIntakeToken("plain")
	h2 := hashIntakeToken("plain")
	if h1 != h2 {
		t.Fatal("must be deterministic")
	}
	if hashIntakeToken("different") == h1 {
		t.Fatal("must distinguish different inputs")
	}
	// sha256 hex = 64 chars
	if len(h1) != 64 {
		t.Fatalf("len = %d", len(h1))
	}
}

func TestTokenPrefix4(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"a":    "a",
		"abcd": "abcd",
		"abcde": "abcd",
		"abcdef": "abcd",
	}
	for in, want := range cases {
		if got := tokenPrefix4(in); got != want {
			t.Errorf("tokenPrefix4(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClientIPAllowed_EmptyCIDRsDeny(t *testing.T) {
	if clientIPAllowed("10.0.0.1", nil) {
		t.Fatal("empty cidrs must deny")
	}
	if clientIPAllowed("10.0.0.1", []string{}) {
		t.Fatal("empty slice must deny")
	}
}

func TestClientIPAllowed_InvalidIPReturnsFalse(t *testing.T) {
	if clientIPAllowed("not-an-ip", []string{"10.0.0.0/8"}) {
		t.Fatal("invalid IP must deny")
	}
}

func TestClientIPAllowed_SingleIPMatch(t *testing.T) {
	if !clientIPAllowed("192.0.2.1", []string{"192.0.2.1"}) {
		t.Fatal("exact single-IP match must allow")
	}
	if clientIPAllowed("192.0.2.2", []string{"192.0.2.1"}) {
		t.Fatal("non-match single IP must deny")
	}
}

func TestClientIPAllowed_CIDRMatch(t *testing.T) {
	if !clientIPAllowed("192.0.2.5", []string{"192.0.2.0/24"}) {
		t.Fatal("in-range CIDR must allow")
	}
	if clientIPAllowed("198.51.100.1", []string{"192.0.2.0/24"}) {
		t.Fatal("out-of-range must deny")
	}
}

func TestClientIPAllowed_MultipleCIDRs(t *testing.T) {
	cidrs := []string{"10.0.0.0/8", "192.0.2.0/24"}
	if !clientIPAllowed("10.1.2.3", cidrs) {
		t.Fatal("first range match")
	}
	if !clientIPAllowed("192.0.2.42", cidrs) {
		t.Fatal("second range match")
	}
	if clientIPAllowed("172.16.0.1", cidrs) {
		t.Fatal("not in any range")
	}
}

func TestClientIPAllowed_InvalidCIDRSkipped(t *testing.T) {
	if clientIPAllowed("10.0.0.1", []string{"not-a-cidr/xx"}) {
		t.Fatal("invalid CIDR must not match")
	}
	if !clientIPAllowed("10.0.0.1", []string{"not-a-cidr/xx", "10.0.0.0/8"}) {
		t.Fatal("invalid CIDR skipped, valid match must allow")
	}
}

func TestRequireIntakeToken_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil token store returns 503", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		h.RequireIntakeToken(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("missing bearer token returns 401", func(t *testing.T) {
		storeI := &fakeIntakeTokenStore{}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		h.RequireIntakeToken(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("token not found in store returns 401", func(t *testing.T) {
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer invalidtoken")
		h.RequireIntakeToken(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("token lookup general error returns 500", func(t *testing.T) {
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{}, errors.New("db fail")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer sometoken")
		h.RequireIntakeToken(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("token revoked returns 401", func(t *testing.T) {
		revTime := time.Now().Add(-1 * time.Hour)
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{
					TokenID:      "t-1",
					ClientLabel:  "jira",
					SourceSystem: "jira",
					RevokedAt:    &revTime,
				}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer sometoken")
		h.RequireIntakeToken(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "token is revoked") {
			t.Fatalf("expected revoked body, got %s", rec.Body.String())
		}
	})

	t.Run("token expired returns 401", func(t *testing.T) {
		expTime := time.Now().Add(-1 * time.Hour)
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{
					TokenID:      "t-1",
					ClientLabel:  "jira",
					SourceSystem: "jira",
					ExpiresAt:    &expTime,
				}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer sometoken")
		h.RequireIntakeToken(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "token is expired") {
			t.Fatalf("expected expired body, got %s", rec.Body.String())
		}
	})

	t.Run("IP not allowed returns 401", func(t *testing.T) {
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{
					TokenID:      "t-1",
					ClientLabel:  "jira",
					SourceSystem: "jira",
					AllowedIPs:   []string{"192.168.1.1"},
				}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test", nil)


		c.Request.Header.Set("Authorization", "Bearer sometoken")
		c.Request.RemoteAddr = "10.0.0.1:1234"
		h.RequireIntakeToken(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "client IP is not in allowlist") {
			t.Fatalf("expected IP error, got %s", rec.Body.String())
		}
	})

	t.Run("success next called and context populated", func(t *testing.T) {
		var touchedToken string
		storeI := &fakeIntakeTokenStore{
			lookupDevRequestIntakeTokenFunc: func(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
				return domain.DevRequestIntakeToken{
					TokenID:      "t-1",
					ClientLabel:  "jira",
					SourceSystem: "jira",
					AllowedIPs:   []string{"10.0.0.1"},
				}, nil
			},
			markDevRequestIntakeTokenUsedFunc: func(ctx context.Context, tokenID string) error {
				touchedToken = tokenID
				return nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec := httptest.NewRecorder()
		r := gin.New()
		var nextCalled bool
		var capturedSystem string
		var capturedToken string
		r.POST("/test", h.RequireIntakeToken, func(gc *gin.Context) {
			nextCalled = true
			capturedSystem = gc.GetString(ctxKeyDREQSourceSystem)
			capturedToken = gc.GetString(ctxKeyDREQTokenID)
			gc.String(200, "ok")
		})

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer sometoken")
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if !nextCalled {
			t.Fatal("expected next handler to be called")
		}
		if touchedToken != "t-1" {
			t.Fatalf("expected token touch t-1, got %s", touchedToken)
		}
		if capturedSystem != "jira" || capturedToken != "t-1" {
			t.Fatalf("context values are missing or mismatched: system=%s token=%s", capturedSystem, capturedToken)
		}
	})
}

func TestIntakeDevRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil store returns 503", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/intake", nil)
		h.IntakeDevRequest(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("missing source system context returns 401", func(t *testing.T) {
		storeI := &fakeDevRequestStore{}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/intake", nil)
		h.IntakeDevRequest(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("invalid json returns 400", func(t *testing.T) {
		storeI := &fakeDevRequestStore{}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader("invalid-json"))
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("idempotence hit returns 200", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: "dr-1", Title: "existing", SourceSystem: sourceSystem, ExternalRef: externalRef}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("idempotence store lookup error returns 500", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{}, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("invalid intake fields saves rejected status", func(t *testing.T) {
		var createdRow domain.DevRequest
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				dr.ID = "dr-rejected"
				createdRow = dr
				return dr, nil
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "", "external_ref": "REF-1"}` // missing requester, assignee, title
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		if createdRow.Status != domain.DevRequestStatusRejected {
			t.Fatalf("expected rejected status, got %s", createdRow.Status)
		}
		if !strings.Contains(createdRow.RejectedReason, "title is required") {
			t.Fatalf("expected rejected reason about title, got %s", createdRow.RejectedReason)
		}
	})

	t.Run("creation conflict on duplicate external_ref triggers idempotence lookup race retry", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				return domain.DevRequest{}, store.ErrConflict
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: "dr-race", Title: "existing", SourceSystem: sourceSystem, ExternalRef: externalRef}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("creation conflict fallback assignee error successfully saves fallback rejected row", func(t *testing.T) {
		var createdRow domain.DevRequest
		callCount := 0
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				callCount++
				if callCount == 1 {
					// 1st call conflict
					return domain.DevRequest{}, store.ErrConflict
				}
				// 2nd call success
				dr.ID = "dr-fallback"
				createdRow = dr
				return dr, nil
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				// second lookup after conflict also fails with NotFound to simulate conflict on assignee FK
				return domain.DevRequest{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "requester": "alice", "assignee_user_id": "bob_non_existent", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		if createdRow.Status != domain.DevRequestStatusRejected || createdRow.AssigneeUserID != "" {
			t.Fatalf("expected rejected status & empty assignee: status=%s assignee=%s", createdRow.Status, createdRow.AssigneeUserID)
		}
	})

	t.Run("creation conflict fallback fails returns 500", func(t *testing.T) {
		callCount := 0
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, store.ErrConflict
				}
				return domain.DevRequest{}, errors.New("db error")
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "requester": "alice", "assignee_user_id": "bob", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("creation general error returns 500", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				return domain.DevRequest{}, errors.New("db error")
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "requester": "alice", "assignee_user_id": "bob", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("success creation returns 201", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			createDevRequestFunc: func(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
				dr.ID = "dr-new"
				return dr, nil
			},
			getDevRequestByExternalRefFunc: func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
				return domain.DevRequest{}, store.ErrNotFound
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		body := `{"title": "new", "requester": "alice", "assignee_user_id": "bob", "external_ref": "REF-1"}`
		c.Request = httptest.NewRequest("POST", "/intake", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(ctxKeyDREQSourceSystem, "jira")
		h.IntakeDevRequest(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})
}

func TestDevRequestsCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ListDevRequests - nil store returns 503", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/dev-requests", nil)
		h.ListDevRequests(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("ListDevRequests - default limits and status query parsing", func(t *testing.T) {
		var capturedOpts store.DevRequestListOptions
		storeI := &fakeDevRequestStore{
			listDevRequestsFunc: func(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
				capturedOpts = opts
				return []domain.DevRequest{{ID: "dr-1"}}, 1, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/dev-requests?limit=150&offset=10&status=pending,in_review", nil)
		c.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.ListDevRequests(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedOpts.Limit != 100 { // clamped to 100
			t.Fatalf("expected limit 100, got %d", capturedOpts.Limit)
		}
		if capturedOpts.Offset != 10 {
			t.Fatalf("expected offset 10, got %d", capturedOpts.Offset)
		}
		if len(capturedOpts.Statuses) != 2 || capturedOpts.Statuses[0] != "pending" {
			t.Fatalf("statuses parsing issue: %v", capturedOpts.Statuses)
		}
	})

	t.Run("ListDevRequests - limit less than 1 and row-level filter for non-elevated user", func(t *testing.T) {
		var capturedOpts store.DevRequestListOptions
		storeI := &fakeDevRequestStore{
			listDevRequestsFunc: func(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
				capturedOpts = opts
				return []domain.DevRequest{}, 0, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/dev-requests?limit=0", nil)
		c.Set("devhub_actor_login", "alice")
		c.Set("devhub_actor_role", "developer")
		h.ListDevRequests(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if capturedOpts.Limit != 1 { // clamped to 1
			t.Fatalf("expected limit 1, got %d", capturedOpts.Limit)
		}
		if capturedOpts.AssigneeUserID != "alice" {
			t.Fatalf("expected row-level filter for alice, got %s", capturedOpts.AssigneeUserID)
		}
	})

	t.Run("ListDevRequests - store list error returns 500", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			listDevRequestsFunc: func(ctx context.Context, opts store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
				return nil, 0, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/dev-requests", nil)
		h.ListDevRequests(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("GetDevRequest - not found and store error", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-notfound" {
					return domain.DevRequest{}, store.ErrNotFound
				}
				return domain.DevRequest{}, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 404 Case
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-notfound"}}
		c1.Request = httptest.NewRequest("GET", "/dev-requests/dr-notfound", nil)
		h.GetDevRequest(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// 500 Case
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-dbfail"}}
		c2.Request = httptest.NewRequest("GET", "/dev-requests/dr-dbfail", nil)
		h.GetDevRequest(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}
	})

	t.Run("GetDevRequest - enforceRowOwnership mismatch", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: "dr-1", AssigneeUserID: "bob"}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c.Request = httptest.NewRequest("GET", "/dev-requests/dr-1", nil)
		c.Set("devhub_actor_login", "alice") // mismatch assignee
		c.Set("devhub_actor_role", "developer")
		h.GetDevRequest(c)
		if rec.Code != 403 {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("GetDevRequest - success", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: "dr-1", AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c.Request = httptest.NewRequest("GET", "/dev-requests/dr-1", nil)
		c.Set("devhub_actor_login", "bob")
		c.Set("devhub_actor_role", "developer")
		h.GetDevRequest(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("RejectDevRequest - empty reason & bind error", func(t *testing.T) {
		storeI := &fakeDevRequestStore{}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// Bind error
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("POST", "/dev-requests/dr-1/reject", strings.NewReader("invalid-json"))
		h.RejectDevRequest(c1)
		if rec1.Code != 400 {
			t.Fatalf("expected 400, got %d", rec1.Code)
		}

		// Empty reason
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/dev-requests/dr-1/reject", strings.NewReader(`{"rejected_reason": ""}`))
		h.RejectDevRequest(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}
	})

	t.Run("RejectDevRequest - lookup not found & general db error", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-notfound" {
					return domain.DevRequest{}, store.ErrNotFound
				}
				return domain.DevRequest{}, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})



		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-notfound"}}
		c1.Request = httptest.NewRequest("POST", "/dev-requests/dr-notfound/reject", strings.NewReader(`{"rejected_reason": "fail"}`))
		h.RejectDevRequest(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// 500
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-dbfail"}}
		c2.Request = httptest.NewRequest("POST", "/dev-requests/dr-dbfail/reject", strings.NewReader(`{"rejected_reason": "fail"}`))
		h.RejectDevRequest(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}
	})

	t.Run("RejectDevRequest - transition invalid or transition store error", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-invalid-state" {
					// status closed is invalid to transition to rejected
					return domain.DevRequest{ID: id, Status: domain.DevRequestStatusClosed, AssigneeUserID: "bob"}, nil
				}
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusPending, AssigneeUserID: "bob"}, nil
			},
			transitionDevRequestStatusFunc: func(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
				return domain.DevRequest{}, errors.New("transition db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// Invalid transition (409)
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-invalid-state"}}
		c1.Request = httptest.NewRequest("POST", "/dev-requests/dr-invalid-state/reject", strings.NewReader(`{"rejected_reason": "fail"}`))
		c1.Set("devhub_actor_login", "bob")
		c1.Set("devhub_actor_role", "developer")
		h.RejectDevRequest(c1)
		if rec1.Code != 409 {
			t.Fatalf("expected 409, got %d", rec1.Code)
		}

		// Transition store error (500)
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-normal"}}
		c2.Request = httptest.NewRequest("POST", "/dev-requests/dr-normal/reject", strings.NewReader(`{"rejected_reason": "fail"}`))
		c2.Set("devhub_actor_login", "bob")
		c2.Set("devhub_actor_role", "developer")
		h.RejectDevRequest(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}
	})

	t.Run("RejectDevRequest - success", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusPending, AssigneeUserID: "bob"}, nil
			},
			transitionDevRequestStatusFunc: func(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusRejected, RejectedReason: rejectedReason}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c.Request = httptest.NewRequest("POST", "/dev-requests/dr-1/reject", strings.NewReader(`{"rejected_reason": "real rejected"}`))
		c.Set("devhub_actor_login", "bob")
		c.Set("devhub_actor_role", "developer")
		h.RejectDevRequest(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("PatchDevRequest - system_admin only gate", func(t *testing.T) {
		storeI := &fakeDevRequestStore{}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader(`{"assignee_user_id": "new-guy"}`))
		c.Set("devhub_actor_role", "pmo_manager") // not system_admin, dev fallback disabled
		h.PatchDevRequest(c)
		if rec.Code != 403 {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("PatchDevRequest - bind error, missing assignee & lookup error", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{}, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// Bind error
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader("invalid-json"))
		c1.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c1)
		if rec1.Code != 400 {
			t.Fatalf("expected 400, got %d", rec1.Code)
		}

		// Missing assignee
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader(`{"assignee_user_id": ""}`))
		c2.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// Lookup error
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "dev_request_id", Value: "dr-dbfail"}}
		c3.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-dbfail", strings.NewReader(`{"assignee_user_id": "alice"}`))
		c3.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}
	})

	t.Run("PatchDevRequest - not found & reassing error conflict / general error", func(t *testing.T) {
		callCount := 0
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-notfound" {
					return domain.DevRequest{}, store.ErrNotFound
				}
				return domain.DevRequest{ID: id, AssigneeUserID: "bob"}, nil
			},
			reassignDevRequestFunc: func(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, store.ErrConflict // conflict: assignee not exist
				}
				return domain.DevRequest{}, errors.New("reassign db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 404 Case
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-notfound"}}
		c1.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-notfound", strings.NewReader(`{"assignee_user_id": "alice"}`))
		c1.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// Reassign conflict (409)
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c2.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader(`{"assignee_user_id": "alice"}`))
		c2.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c2)
		if rec2.Code != 409 {
			t.Fatalf("expected 409, got %d", rec2.Code)
		}

		// Reassign error (500)
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c3.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader(`{"assignee_user_id": "alice"}`))
		c3.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}
	})

	t.Run("PatchDevRequest - success reassign", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob"}, nil
			},
			reassignDevRequestFunc: func(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: newAssigneeUserID}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c.Request = httptest.NewRequest("PATCH", "/dev-requests/dr-1", strings.NewReader(`{"assignee_user_id": "alice"}`))
		c.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.PatchDevRequest(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("CloseDevRequest - admin validation and invalid transitions", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-notfound" {
					return domain.DevRequest{}, store.ErrNotFound
				}
				if id == "dr-dbfail" {
					return domain.DevRequest{}, errors.New("db error")
				}
				// status pending is invalid for close (only registered/rejected allowed)
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusPending}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// PMO manager role disallowed
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-1", nil)
		c1.Set("devhub_actor_role", "pmo_manager")
		h.CloseDevRequest(c1)
		if rec1.Code != 403 {
			t.Fatalf("expected 403, got %d", rec1.Code)
		}

		// 404
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-notfound"}}
		c2.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-notfound", nil)
		c2.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.CloseDevRequest(c2)
		if rec2.Code != 404 {
			t.Fatalf("expected 404, got %d", rec2.Code)
		}

		// 500
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "dev_request_id", Value: "dr-dbfail"}}
		c3.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-dbfail", nil)
		c3.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.CloseDevRequest(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}

		// Invalid transition (422)
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "dev_request_id", Value: "dr-pending"}}
		c4.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-pending", nil)
		c4.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.CloseDevRequest(c4)
		if rec4.Code != 422 {
			t.Fatalf("expected 422, got %d", rec4.Code)
		}
	})

	t.Run("CloseDevRequest - transition error vs success close", func(t *testing.T) {
		callCount := 0
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusRejected}, nil
			},
			transitionDevRequestStatusFunc: func(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, errors.New("transition error")
				}
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusClosed}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// Transition error (500)
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c1.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-1", nil)
		c1.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.CloseDevRequest(c1)
		if rec1.Code != 500 {
			t.Fatalf("expected 500, got %d", rec1.Code)
		}

		// Success Close
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c2.Request = httptest.NewRequest("DELETE", "/dev-requests/dr-1", nil)
		c2.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		h.CloseDevRequest(c2)
		if rec2.Code != 200 {
			t.Fatalf("expected 200, got %d", rec2.Code)
		}
	})
}

func TestRegisterDevRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("RegisterDevRequest - basic validation & mutual exclusion & mismatch", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("POST", "/register", nil)
		h.RegisterDevRequest(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeDevRequestStore{}
		h = NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 2. bind error -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader("invalid-json"))
		h.RegisterDevRequest(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. invalid target_type -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "invalid"}`))
		h.RegisterDevRequest(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		// 4. mutual exclusion fail (payloadCount != 1)
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1", "application_payload": {}}`))
		h.RegisterDevRequest(c4)
		if rec4.Code != 400 {
			t.Fatalf("expected 400, got %d", rec4.Code)
		}

		// 5. mismatch target_type and payload
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "project", "application_payload": {}}`))
		h.RegisterDevRequest(c5)
		if rec5.Code != 400 {
			t.Fatalf("expected 400, got %d", rec5.Code)
		}
	})

	t.Run("RegisterDevRequest - dev request lookup cases", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-notfound" {
					return domain.DevRequest{}, store.ErrNotFound
				}
				return domain.DevRequest{}, errors.New("db error")
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 1. not found -> 404
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-notfound"}}
		c1.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		h.RegisterDevRequest(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// 2. lookup error -> 500
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-dbfail"}}
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		h.RegisterDevRequest(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}
	})

	t.Run("RegisterDevRequest - ownership gate & invalid devrequest status", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				if id == "dr-ownership-fail" {
					return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
				}
				// status closed is invalid to register
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusClosed}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 1. ownership mismatch -> 403
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-ownership-fail"}}
		c1.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		c1.Set("devhub_actor_login", "alice")
		c1.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c1)
		if rec1.Code != 403 {
			t.Fatalf("expected 403, got %d", rec1.Code)
		}

		// 2. invalid status closed -> 409 conflict
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-closed"}}
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		c2.Set("devhub_actor_login", "bob")
		c2.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c2)
		if rec2.Code != 409 {
			t.Fatalf("expected 409, got %d", rec2.Code)
		}
	})

	t.Run("RegisterDevRequest - Legacy Path", func(t *testing.T) {
		callCount := 0
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
			markDevRequestRegisteredFunc: func(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, errors.New("db error")
				}
				return domain.DevRequest{ID: id, Status: domain.DevRequestStatusRegistered}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 1. mark registered store error -> 500
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c1.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		c1.Set("devhub_actor_login", "bob")
		c1.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c1)
		if rec1.Code != 500 {
			t.Fatalf("expected 500, got %d", rec1.Code)
		}

		// 2. success -> 200
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader(`{"target_type": "application", "target_id": "app-1"}`))
		c2.Set("devhub_actor_login", "bob")
		c2.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c2)
		if rec2.Code != 200 {
			t.Fatalf("expected 200, got %d", rec2.Code)
		}
	})

	t.Run("RegisterDevRequest - Promote Application Validation Errors", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		cases := []struct {
			name    string
			payload string
			code    int
		}{
			{"invalid app key pattern", `{"target_type": "application", "application_payload": {"key": "INVALID-KEY-12345"}}`, 422},
			{"missing app name", `{"target_type": "application", "application_payload": {"key": "APP", "name": ""}}`, 400},
			{"missing owner", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": ""}}`, 400},
			{"missing leader", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": ""}}`, 400},
			{"missing development unit", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": ""}}`, 400},
			{"invalid visibility", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U", "visibility": "invalid"}}`, 400},
			{"invalid status", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U", "visibility": "public", "status": "invalid"}}`, 400},
			{"invalid start date format", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U", "visibility": "public", "status": "active", "start_date": "not-a-date"}}`, 400},
			{"invalid due date format", `{"target_type": "application", "application_payload": {"key": "APP", "name": "A", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U", "visibility": "public", "status": "active", "start_date": "2026-05-30", "due_date": "not-a-date"}}`, 400},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
				c.Request = httptest.NewRequest("POST", "/register", strings.NewReader(tc.payload))
				c.Set("devhub_actor_login", "bob")
				c.Set("devhub_actor_role", "developer")
				h.RegisterDevRequest(c)
				if rec.Code != tc.code {
					t.Fatalf("expected status %d, got %d", tc.code, rec.Code)
				}
			})
		}
	})

	t.Run("RegisterDevRequest - Promote Application Primary Repo Validation Errors", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
		}
		appStore := &fakeDevReqAppStore{
			listSCMProvidersFunc: func(ctx context.Context) ([]domain.SCMProvider, error) {
				return []domain.SCMProvider{
					{ProviderKey: "github", Enabled: true},
					{ProviderKey: "gitea", Enabled: false},
				}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{
			DevRequestStore:  storeI,
			ApplicationStore: appStore,
		})

		basePayload := `{"target_type": "application", "application_payload": {
			"key": "APP", "name": "App", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U",
			"visibility": "public", "status": "active",
			"primary_repo": %s
		}}`

		cases := []struct {
			name        string
			repoPayload string
			code        int
		}{
			{"missing repo_provider", `{"repo_provider": "", "repo_full_name": "org/repo"}`, 400},
			{"missing repo_full_name", `{"repo_provider": "github", "repo_full_name": ""}`, 400},
			{"invalid role", `{"repo_provider": "github", "repo_full_name": "org/repo", "role": "invalid"}`, 422},
			{"unsupported repo provider (disabled)", `{"repo_provider": "gitea", "repo_full_name": "org/repo", "role": "primary"}`, 422},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
				c.Request = httptest.NewRequest("POST", "/register", strings.NewReader(strings.Replace(basePayload, "%s", tc.repoPayload, 1)))
				c.Set("devhub_actor_login", "bob")
				c.Set("devhub_actor_role", "developer")
				h.RegisterDevRequest(c)
				if rec.Code != tc.code {
					t.Fatalf("expected status %d, got %d. Body: %s", tc.code, rec.Code, rec.Body.String())
				}
			})
		}
	})

	t.Run("RegisterDevRequest - Promote Application Store Response Mappings", func(t *testing.T) {
		callCount := 0
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
			registerDevRequestWithNewApplicationFunc: func(ctx context.Context, drID string, app domain.Application, primaryRepo *domain.ApplicationRepository) (domain.DevRequest, domain.Application, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, domain.Application{}, store.ErrConflict
				}
				if callCount == 2 {
					return domain.DevRequest{}, domain.Application{}, store.ErrNotFound
				}
				if callCount == 3 {
					return domain.DevRequest{}, domain.Application{}, errors.New("db error")
				}
				return domain.DevRequest{ID: drID, Status: domain.DevRequestStatusRegistered}, app, nil
			},
		}
		appStore := &fakeDevReqAppStore{}
		h := NewDevRequestHandler(DevRequestConfig{
			DevRequestStore:  storeI,
			ApplicationStore: appStore,
		})

		payload := `{"target_type": "application", "application_payload": {
			"key": "APP", "name": "App", "owner_user_id": "O", "leader_user_id": "L", "development_unit_id": "U",
			"visibility": "public", "status": "active"
		}}`

		// 1. Conflict -> 409
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c1.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload))
		c1.Set("devhub_actor_login", "bob")
		c1.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c1)
		if rec1.Code != 409 {
			t.Fatalf("expected 409, got %d", rec1.Code)
		}

		// 2. Not Found -> 404
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload))
		c2.Set("devhub_actor_login", "bob")
		c2.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c2)
		if rec2.Code != 404 {
			t.Fatalf("expected 404, got %d", rec2.Code)
		}

		// 3. General DB Error -> 500
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c3.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload))
		c3.Set("devhub_actor_login", "bob")
		c3.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}

		// 4. Success Promote Application -> 200
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c4.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload))
		c4.Set("devhub_actor_login", "bob")
		c4.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c4)
		if rec4.Code != 200 {
			t.Fatalf("expected 200, got %d", rec4.Code)
		}
	})

	t.Run("RegisterDevRequest - Promote Project Validation & Store Response Mappings", func(t *testing.T) {
		storeI := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI})

		// 1. missing repository_id -> 400
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		payload1 := `{"target_type": "project", "project_payload": {"repository_id": 0}}`
		c1.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload1))
		c1.Set("devhub_actor_login", "bob")
		c1.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c1)
		if rec1.Code != 400 {
			t.Fatalf("expected 400, got %d", rec1.Code)
		}

		// 2. missing key -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		payload2 := `{"target_type": "project", "project_payload": {"repository_id": 42, "key": ""}}`
		c2.Request = httptest.NewRequest("POST", "/register", strings.NewReader(payload2))
		c2.Set("devhub_actor_login", "bob")
		c2.Set("devhub_actor_role", "developer")
		h.RegisterDevRequest(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// Store response mappings
		callCount := 0
		storeI2 := &fakeDevRequestStore{
			getDevRequestFunc: func(ctx context.Context, id string) (domain.DevRequest, error) {
				return domain.DevRequest{ID: id, AssigneeUserID: "bob", Status: domain.DevRequestStatusPending}, nil
			},
			registerDevRequestWithNewProjectFunc: func(ctx context.Context, drID string, project domain.Project) (domain.DevRequest, domain.Project, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequest{}, domain.Project{}, store.ErrConflict
				}
				if callCount == 2 {
					return domain.DevRequest{}, domain.Project{}, store.ErrNotFound
				}
				if callCount == 3 {
					return domain.DevRequest{}, domain.Project{}, errors.New("db error")
				}
				return domain.DevRequest{ID: drID, Status: domain.DevRequestStatusRegistered}, project, nil
			},
		}
		h2 := NewDevRequestHandler(DevRequestConfig{DevRequestStore: storeI2})
		validPayload := `{"target_type": "project", "project_payload": {
			"repository_id": 42, "key": "PRJ", "name": "Proj", "owner_user_id": "O",
			"visibility": "public", "status": "active"
		}}`

		// 3. Conflict -> 409
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c3.Request = httptest.NewRequest("POST", "/register", strings.NewReader(validPayload))
		c3.Set("devhub_actor_login", "bob")
		c3.Set("devhub_actor_role", "developer")
		h2.RegisterDevRequest(c3)
		if rec3.Code != 409 {
			t.Fatalf("expected 409, got %d", rec3.Code)
		}

		// 4. Not Found -> 404
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c4.Request = httptest.NewRequest("POST", "/register", strings.NewReader(validPayload))
		c4.Set("devhub_actor_login", "bob")
		c4.Set("devhub_actor_role", "developer")
		h2.RegisterDevRequest(c4)
		if rec4.Code != 404 {
			t.Fatalf("expected 404, got %d", rec4.Code)
		}

		// 5. DB Error -> 500
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c5.Request = httptest.NewRequest("POST", "/register", strings.NewReader(validPayload))
		c5.Set("devhub_actor_login", "bob")
		c5.Set("devhub_actor_role", "developer")
		h2.RegisterDevRequest(c5)
		if rec5.Code != 500 {
			t.Fatalf("expected 500, got %d", rec5.Code)
		}

		// 6. Success Promote Project -> 200
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Params = gin.Params{{Key: "dev_request_id", Value: "dr-1"}}
		c6.Request = httptest.NewRequest("POST", "/register", strings.NewReader(validPayload))
		c6.Set("devhub_actor_login", "bob")
		c6.Set("devhub_actor_role", "developer")
		h2.RegisterDevRequest(c6)
		if rec6.Code != 200 {
			t.Fatalf("expected 200, got %d", rec6.Code)
		}
	})
}

func TestIntakeTokenAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CreateDevRequestIntakeToken - validation & error mappings", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("POST", "/tokens", nil)
		h.CreateDevRequestIntakeToken(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeIntakeTokenStore{}
		h = NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})

		// 2. bind error -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader("invalid-json"))
		h.CreateDevRequestIntakeToken(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. missing label/system -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(`{"client_label": "", "source_system": "jira"}`))
		h.CreateDevRequestIntakeToken(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(`{"client_label": "L", "source_system": ""}`))
		h.CreateDevRequestIntakeToken(c4)
		if rec4.Code != 400 {
			t.Fatalf("expected 400, got %d", rec4.Code)
		}

		// 4. IP validation failures
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(`{"client_label": "L", "source_system": "S", "allowed_ips": ["invalid-ip"]}`))
		h.CreateDevRequestIntakeToken(c5)
		if rec5.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec5.Code, rec5.Body.String())
		}

		rec5b := httptest.NewRecorder()
		c5b, _ := gin.CreateTestContext(rec5b)
		c5b.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(`{"client_label": "L", "source_system": "S", "allowed_ips": ["192.168.1.0/99"]}`))
		h.CreateDevRequestIntakeToken(c5b)
		if rec5b.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec5b.Code, rec5b.Body.String())
		}

		// 5. invalid expires_at RFC3339 -> 400
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(`{"client_label": "L", "source_system": "S", "allowed_ips": ["10.0.0.1"], "expires_at": "not-rfc3339"}`))
		h.CreateDevRequestIntakeToken(c6)
		if rec6.Code != 400 {
			t.Fatalf("expected 400, got %d", rec6.Code)
		}
	})

	t.Run("CreateDevRequestIntakeToken - store response mappings & success", func(t *testing.T) {
		callCount := 0
		storeI := &fakeIntakeTokenStore{
			createDevRequestIntakeTokenFunc: func(ctx context.Context, tok domain.DevRequestIntakeToken) (domain.DevRequestIntakeToken, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequestIntakeToken{}, store.ErrConflict
				}
				if callCount == 2 {
					return domain.DevRequestIntakeToken{}, errors.New("db error")
				}
				tok.TokenID = "t-new"
				return tok, nil
			},
		}
		h := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		payload := `{"client_label": "L", "source_system": "S", "allowed_ips": ["10.0.0.1"], "expires_at": "2026-05-30T12:00:00Z"}`

		// 1. Conflict -> 409
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(payload))
		c1.Set("devhub_actor_login", "admin")
		h.CreateDevRequestIntakeToken(c1)
		if rec1.Code != 409 {
			t.Fatalf("expected 409, got %d", rec1.Code)
		}

		// 2. general db error -> 500
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(payload))
		h.CreateDevRequestIntakeToken(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}

		// 3. success -> 201
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/tokens", strings.NewReader(payload))
		h.CreateDevRequestIntakeToken(c3)
		if rec3.Code != 201 {
			t.Fatalf("expected 201, got %d", rec3.Code)
		}
	})

	t.Run("ListDevRequestIntakeTokens - nil store & store error & success", func(t *testing.T) {
		// 1. nil store -> 503
		h := NewDevRequestHandler(DevRequestConfig{})
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("GET", "/tokens", nil)
		h.ListDevRequestIntakeTokens(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		// 2. store error -> 500
		storeI := &fakeIntakeTokenStore{
			listDevRequestIntakeTokensFunc: func(ctx context.Context) ([]domain.DevRequestIntakeToken, error) {
				return nil, errors.New("db error")
			},
		}
		h = NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("GET", "/tokens", nil)
		h.ListDevRequestIntakeTokens(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}

		// 3. success -> 200
		storeI2 := &fakeIntakeTokenStore{
			listDevRequestIntakeTokensFunc: func(ctx context.Context) ([]domain.DevRequestIntakeToken, error) {
				return []domain.DevRequestIntakeToken{{TokenID: "t-1"}}, nil
			},
		}
		h2 := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI2})
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("GET", "/tokens", nil)
		h2.ListDevRequestIntakeTokens(c3)
		if rec3.Code != 200 {
			t.Fatalf("expected 200, got %d", rec3.Code)
		}
	})

	t.Run("RevokeDevRequestIntakeToken - validation & store responses", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})

		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("DELETE", "/tokens/t-1", nil)
		c1.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		h.RevokeDevRequestIntakeToken(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		callCount := 0
		storeI := &fakeIntakeTokenStore{
			revokeDevRequestIntakeTokenFunc: func(ctx context.Context, tokenID string) (domain.DevRequestIntakeToken, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequestIntakeToken{}, store.ErrNotFound
				}
				if callCount == 2 {
					return domain.DevRequestIntakeToken{}, errors.New("db error")
				}
				return domain.DevRequestIntakeToken{TokenID: tokenID}, nil
			},
		}
		h = NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})

		// 2. not found -> 404
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("DELETE", "/tokens/t-notfound", nil)
		c2.Params = gin.Params{{Key: "token_id", Value: "t-notfound"}}
		h.RevokeDevRequestIntakeToken(c2)
		if rec2.Code != 404 {
			t.Fatalf("expected 404, got %d", rec2.Code)
		}

		// 3. db error -> 500
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("DELETE", "/tokens/t-dbfail", nil)
		c3.Params = gin.Params{{Key: "token_id", Value: "t-dbfail"}}
		h.RevokeDevRequestIntakeToken(c3)
		if rec3.Code != 500 {
			t.Fatalf("expected 500, got %d", rec3.Code)
		}

		// 4. success -> 200
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Request = httptest.NewRequest("DELETE", "/tokens/t-1", nil)
		c4.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		h.RevokeDevRequestIntakeToken(c4)
		if rec4.Code != 200 {
			t.Fatalf("expected 200, got %d", rec4.Code)
		}
	})

	t.Run("UpdateDevRequestIntakeTokenIPs - validation & store responses", func(t *testing.T) {
		h := NewDevRequestHandler(DevRequestConfig{})

		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Request = httptest.NewRequest("PATCH", "/tokens/t-1", nil)
		c1.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		h.UpdateDevRequestIntakeTokenIPs(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeIntakeTokenStore{}
		h = NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI})

		// 2. empty token_id -> 400 is not possible via Param binding but let's test general 400 JSON bind error
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c2.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader("invalid-json"))
		h.UpdateDevRequestIntakeTokenIPs(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. neither ips nor expires provided -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c3.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{}`))
		h.UpdateDevRequestIntakeTokenIPs(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec3.Code, rec3.Body.String())
		}

		// 4. allowed_ips format error -> 400
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c4.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{"allowed_ips": "not-an-array"}`))
		h.UpdateDevRequestIntakeTokenIPs(c4)
		if rec4.Code != 400 {
			t.Fatalf("expected 400, got %d", rec4.Code)
		}

		rec4b := httptest.NewRecorder()
		c4b, _ := gin.CreateTestContext(rec4b)
		c4b.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c4b.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{"allowed_ips": [1234]}`))
		h.UpdateDevRequestIntakeTokenIPs(c4b)
		if rec4b.Code != 400 {
			t.Fatalf("expected 400, got %d", rec4b.Code)
		}

		// 5. invalid ip entry in array -> 400
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c5.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{"allowed_ips": ["invalid-ip"]}`))
		h.UpdateDevRequestIntakeTokenIPs(c5)
		if rec5.Code != 400 {
			t.Fatalf("expected 400, got %d", rec5.Code)
		}

		// 6. expires_at format error (not string) & invalid RFC3339 -> 400
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c6.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{"expires_at": 12345}`))
		h.UpdateDevRequestIntakeTokenIPs(c6)
		if rec6.Code != 400 {
			t.Fatalf("expected 400, got %d", rec6.Code)
		}

		rec6b := httptest.NewRecorder()
		c6b, _ := gin.CreateTestContext(rec6b)
		c6b.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c6b.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(`{"expires_at": "not-rfc3339"}`))
		h.UpdateDevRequestIntakeTokenIPs(c6b)
		if rec6b.Code != 400 {
			t.Fatalf("expected 400, got %d", rec6b.Code)
		}

		// Store response mappings
		callCount := 0
		storeI2 := &fakeIntakeTokenStore{
			updateDevRequestIntakeTokenFunc: func(ctx context.Context, tokenID string, allowedIPs []string, expiresAt *time.Time, updateIPs, updateExpiresAt bool) (domain.DevRequestIntakeToken, error) {
				callCount++
				if callCount == 1 {
					return domain.DevRequestIntakeToken{}, store.ErrNotFound
				}
				if callCount == 2 {
					return domain.DevRequestIntakeToken{}, store.ErrConflict
				}
				if callCount == 3 {
					return domain.DevRequestIntakeToken{}, errors.New("db error")
				}
				return domain.DevRequestIntakeToken{TokenID: tokenID, AllowedIPs: allowedIPs, ExpiresAt: expiresAt}, nil
			},
		}
		h2 := NewDevRequestHandler(DevRequestConfig{DevRequestIntakeTokenStore: storeI2})
		validPayload := `{"allowed_ips": ["10.0.0.1"], "expires_at": "2026-05-30T12:00:00Z"}`

		// 7. Not Found -> 404
		rec7 := httptest.NewRecorder()
		c7, _ := gin.CreateTestContext(rec7)
		c7.Params = gin.Params{{Key: "token_id", Value: "t-notfound"}}
		c7.Request = httptest.NewRequest("PATCH", "/tokens/t-notfound", strings.NewReader(validPayload))
		h2.UpdateDevRequestIntakeTokenIPs(c7)
		if rec7.Code != 404 {
			t.Fatalf("expected 404, got %d", rec7.Code)
		}

		// 8. Conflict -> 409
		rec8 := httptest.NewRecorder()
		c8, _ := gin.CreateTestContext(rec8)
		c8.Params = gin.Params{{Key: "token_id", Value: "t-revoked"}}
		c8.Request = httptest.NewRequest("PATCH", "/tokens/t-revoked", strings.NewReader(validPayload))
		h2.UpdateDevRequestIntakeTokenIPs(c8)
		if rec8.Code != 409 {
			t.Fatalf("expected 409, got %d", rec8.Code)
		}

		// 9. DB Error -> 500
		rec9 := httptest.NewRecorder()
		c9, _ := gin.CreateTestContext(rec9)
		c9.Params = gin.Params{{Key: "token_id", Value: "t-dbfail"}}
		c9.Request = httptest.NewRequest("PATCH", "/tokens/t-dbfail", strings.NewReader(validPayload))
		h2.UpdateDevRequestIntakeTokenIPs(c9)
		if rec9.Code != 500 {
			t.Fatalf("expected 500, got %d", rec9.Code)
		}

		// 10. Success -> 200
		rec10 := httptest.NewRecorder()
		c10, _ := gin.CreateTestContext(rec10)
		c10.Params = gin.Params{{Key: "token_id", Value: "t-1"}}
		c10.Request = httptest.NewRequest("PATCH", "/tokens/t-1", strings.NewReader(validPayload))
		h2.UpdateDevRequestIntakeTokenIPs(c10)
		if rec10.Code != 200 {
			t.Fatalf("expected 200, got %d", rec10.Code)
		}
	})
}

// ---------------------------------------------------------------------------
// validateAllowedIPs tests (pure function, dev_request_intake_tokens_admin.go)
// ---------------------------------------------------------------------------

func TestValidateAllowedIPs_ValidSingleIP(t *testing.T) {
	got, errMsg := validateAllowedIPs([]string{"192.168.1.1"})
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if len(got) != 1 || got[0] != "192.168.1.1" {
		t.Fatalf("expected [192.168.1.1], got %v", got)
	}
}

func TestValidateAllowedIPs_ValidCIDR(t *testing.T) {
	got, errMsg := validateAllowedIPs([]string{"10.0.0.0/8"})
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Fatalf("expected [10.0.0.0/8], got %v", got)
	}
}

func TestValidateAllowedIPs_InvalidIP(t *testing.T) {
	_, errMsg := validateAllowedIPs([]string{"not-an-ip"})
	if errMsg == "" || !strings.Contains(errMsg, "not a valid IP") {
		t.Fatalf("expected IP validation error, got %q", errMsg)
	}
}

func TestValidateAllowedIPs_InvalidCIDR(t *testing.T) {
	_, errMsg := validateAllowedIPs([]string{"10.0.0.0/33"})
	if errMsg == "" || !strings.Contains(errMsg, "not a valid CIDR") {
		t.Fatalf("expected CIDR validation error, got %q", errMsg)
	}
}

func TestValidateAllowedIPs_DedupPreservesOrder(t *testing.T) {
	got, errMsg := validateAllowedIPs([]string{"10.0.0.1", "10.0.0.2", "10.0.0.1", "10.0.0.3"})
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if len(got) != 3 || got[0] != "10.0.0.1" || got[1] != "10.0.0.2" || got[2] != "10.0.0.3" {
		t.Fatalf("expected deduped order-preserved list, got %v", got)
	}
}

func TestValidateAllowedIPs_RejectsEmpty(t *testing.T) {
	_, errMsg := validateAllowedIPs(nil)
	if errMsg == "" || !strings.Contains(errMsg, "at least one") {
		t.Fatalf("expected 'at least one' error, got %q", errMsg)
	}
	_, errMsg = validateAllowedIPs([]string{})
	if errMsg == "" || !strings.Contains(errMsg, "at least one") {
		t.Fatalf("expected 'at least one' error for empty list, got %q", errMsg)
	}
}

func TestValidateAllowedIPs_SkipsEmptyEntries(t *testing.T) {
	got, errMsg := validateAllowedIPs([]string{"10.0.0.1", "", "  ", "10.0.0.2"})
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries (empty skipped), got %v", got)
	}
}

// ---------------------------------------------------------------------------
// RevokeDevRequestIntakeToken edge — empty token_id
// ---------------------------------------------------------------------------

func TestRevokeDevRequestIntakeToken_EmptyTokenID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDevRequestHandler(DevRequestConfig{
		DevRequestIntakeTokenStore: &fakeIntakeTokenStore{},
	})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("DELETE", "/tokens/", nil)
	// Manually set the param as empty to exercise the validation branch
	c.Params = gin.Params{{Key: "token_id", Value: ""}}

	h.RevokeDevRequestIntakeToken(c)
	if rec.Code != 400 {
		t.Fatalf("expected 400 for empty token_id, got %d: %s", rec.Code, rec.Body.String())
	}
}
