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

type fakeDevReqAppStore struct{}

func (f *fakeDevReqAppStore) ListSCMProviders(_ context.Context) ([]domain.SCMProvider, error) {
	return nil, nil
}
func (f *fakeDevReqAppStore) CreateProjectWithRepositoryPayload(_ context.Context, p domain.Project, _ []int64, _ *store.RepositoryCreatePayload) (domain.Project, error) {
	return p, nil
}
func (f *fakeDevReqAppStore) CreateApplication(_ context.Context, app domain.Application) (domain.Application, error) {
	return app, nil
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
