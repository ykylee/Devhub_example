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
	"github.com/gin-gonic/gin"
)

type fakePlatformLifecycleAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakePlatformLifecycleAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_app_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewPlatformHandler_NonNil(t *testing.T) {
	h := NewPlatformHandler(PlatformConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewPlatformHandler_ConfigPropagation(t *testing.T) {
	cfg := PlatformConfig{
		ProjectModel: "single",
		AuditStore:   &fakePlatformLifecycleAuditStore{},
	}
	h := NewPlatformHandler(cfg)
	if h.cfg.ProjectModel != "single" {
		t.Fatalf("ProjectModel = %q", h.cfg.ProjectModel)
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakePlatformLifecycleAuditStore{}
	h := NewPlatformHandler(PlatformConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "app.test", "platform", "app-1", nil)
	if got.AuditID != "audit_app_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "app.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakePlatformLifecycleAuditStore{}
	h := NewPlatformHandler(PlatformConfig{AuditStore: store})
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

	store := &fakePlatformLifecycleAuditStore{err: errors.New("db_down")}
	h := NewPlatformHandler(PlatformConfig{AuditStore: store})
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

func TestEnforceRowOwnership_DevFallbackAlwaysAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_auth_dev_fallback", true)

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("dev fallback must allow")
	}
}

func TestEnforceRowOwnership_SystemAdminAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "admin")
	c.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("system_admin must be allowed")
	}
}

func TestEnforceRowOwnership_AllowedRoleAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "leader-1")
	c.Set("devhub_actor_role", "team_lead")

	if !h.enforceRowOwnership(c, "owner-x", "team_lead") {
		t.Fatal("explicitly allowed role must pass")
	}
}

func TestEnforceRowOwnership_OwnerLoginMatchAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "owner-x")
	c.Set("devhub_actor_role", "developer")

	if !h.enforceRowOwnership(c, "owner-x") {
		t.Fatal("owner self-match must pass")
	}
}

func TestEnforceRowOwnership_DeniedNon403Body(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakePlatformLifecycleAuditStore{}
	h := NewPlatformHandler(PlatformConfig{AuditStore: store})
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

func TestApplicationResponse_FieldMapping(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	archived := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	app := domain.Platform{
		ID:                "app-1",
		Key:               "APP",
		Name:              "App One",
		Description:       "desc",
		Status:            domain.PlatformStatusActive,
		Visibility:        domain.PlatformVisibilityInternal,
		OwnerUserID:       "owner",
		LeaderUserID:      "leader",
		DevelopmentUnitID: "team",
		StartDate:         &start,
		DueDate:           &due,
		ArchivedAt:        &archived,
		CreatedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	resp := platformResponse(app)
	if resp["id"] != "app-1" || resp["key"] != "APP" || resp["name"] != "App One" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["status"] != "active" || resp["visibility"] != "internal" {
		t.Fatalf("status/visibility: %+v", resp)
	}
	if resp["start_date"] != "2026-01-01" || resp["due_date"] != "2026-12-31" {
		t.Fatalf("dates: %+v", resp)
	}
	if resp["archived_at"] == nil {
		t.Fatal("archived_at should be set")
	}
}

func TestApplicationRepositoryResponse_FieldMapping(t *testing.T) {
	linked := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	link := domain.PlatformRepository{
		PlatformID:      "app-1",
		RepoProvider:       "gitea",
		RepoFullName:       "org/repo",
		ExternalRepoID:     "12345",
		Role:               domain.PlatformRepositoryRole("primary"),
		SyncStatus:         domain.PlatformRepositorySyncStatus("ok"),
		SyncErrorCode:      domain.SyncErrorCode(""),
		LinkedAt:           linked,
		LinkSource:         "direct",
	}
	resp := platformRepositoryResponse(link)
	if resp["platform_id"] != "app-1" || resp["repo_provider"] != "gitea" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["role"] != "primary" || resp["sync_status"] != "ok" {
		t.Fatalf("role/sync: %+v", resp)
	}
	if resp["link_source"] != "direct" {
		t.Fatalf("link_source = %v", resp["link_source"])
	}
}

func TestLinkSourceOrDefault(t *testing.T) {
	cases := map[string]string{
		"direct":      "direct",
		"via_project": "via_project",
		"":            "direct",
		"unknown":     "direct",
		"legacy":      "direct",
	}
	for in, want := range cases {
		if got := linkSourceOrDefault(in); got != want {
			t.Errorf("linkSourceOrDefault(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSCMProviderResponse_NonGiteaNoCreds(t *testing.T) {
	p := domain.SCMProvider{
		ProviderKey:    "github",
		DisplayName:    "GitHub",
		Enabled:        true,
		AdapterVersion: "v1",
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	resp := scmProviderResponse(p)
	if resp["provider_key"] != "github" {
		t.Fatalf("provider_key = %v", resp["provider_key"])
	}
	if resp["has_credentials"] != false {
		t.Fatalf("non-gitea has_credentials must be false, got %v", resp["has_credentials"])
	}
	if resp["enabled"] != true {
		t.Fatal("enabled must propagate")
	}
}

func TestSCMProviderResponse_GiteaEnvFlag(t *testing.T) {
	t.Setenv("GITEA_URL", "https://gitea.example.com")
	t.Setenv("GITEA_TOKEN", "x")
	p := domain.SCMProvider{ProviderKey: "gitea"}
	resp := scmProviderResponse(p)
	if resp["has_credentials"] != true {
		t.Fatalf("gitea with both envs has_credentials must be true, got %v", resp["has_credentials"])
	}
}

func TestSCMProviderResponse_GiteaMissingEnv(t *testing.T) {
	t.Setenv("GITEA_URL", "")
	t.Setenv("GITEA_TOKEN", "")
	p := domain.SCMProvider{ProviderKey: "gitea"}
	resp := scmProviderResponse(p)
	if resp["has_credentials"] != false {
		t.Fatalf("gitea missing env has_credentials must be false, got %v", resp["has_credentials"])
	}
}

func TestFormatDatePtr(t *testing.T) {
	if got := formatDatePtr(nil); got != nil {
		t.Fatalf("nil -> nil, got %v", got)
	}
	tm := time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)
	if got := formatDatePtr(&tm); got != "2026-05-29" {
		t.Fatalf("got %v", got)
	}
}

func TestFormatTimePtr(t *testing.T) {
	if got := formatTimePtr(nil); got != nil {
		t.Fatalf("nil -> nil, got %v", got)
	}
	tm := time.Date(2026, 5, 29, 12, 30, 0, 0, time.UTC)
	if got := formatTimePtr(&tm); got != "2026-05-29T12:30:00Z" {
		t.Fatalf("got %v", got)
	}
}

func TestParseDate_Helpers(t *testing.T) {
	cases := map[string]struct {
		isErr  bool
		isNil  bool
		expect string
	}{
		"":            {false, true, ""},
		"   ":         {false, true, ""},
		"2026-05-29":  {false, false, "2026-05-29"},
		"  2026-05-29  ": {false, false, "2026-05-29"},
		"not-a-date":  {true, false, ""},
	}
	for in, want := range cases {
		got, err := parseDate(in)
		if want.isErr && err == nil {
			t.Errorf("parseDate(%q) expected err", in)
		} else if !want.isErr && err != nil {
			t.Errorf("parseDate(%q) unexpected err: %v", in, err)
		}
		if want.isNil && got != nil {
			t.Errorf("parseDate(%q) expected nil ptr, got %v", in, got)
		}
		if !want.isNil && got != nil && got.Format("2006-01-02") != want.expect {
			t.Errorf("parseDate(%q) = %v, want %s", in, got, want.expect)
		}
	}
}

func TestValidApplicationStatuses(t *testing.T) {
	for _, s := range []string{"planning", "active", "on_hold", "closed", "archived"} {
		if !validPlatformStatuses[s] {
			t.Errorf("expected %q valid", s)
		}
	}
	if validPlatformStatuses["unknown"] {
		t.Error("unknown must be invalid")
	}
}

func TestValidApplicationVisibilities(t *testing.T) {
	for _, v := range []string{"public", "internal", "restricted"} {
		if !validPlatformVisibilities[v] {
			t.Errorf("expected %q valid", v)
		}
	}
	if validPlatformVisibilities["x"] {
		t.Error("x must be invalid")
	}
}

func TestValidApplicationRepoRoles(t *testing.T) {
	for _, r := range []string{"primary", "sub", "shared"} {
		if !validPlatformRepoRoles[r] {
			t.Errorf("expected %q valid", r)
		}
	}
	if validPlatformRepoRoles["x"] {
		t.Error("x must be invalid")
	}
}

func TestProjectResponse_FieldMapping(t *testing.T) {
	p := domain.Project{
		ID:            "proj-1",
		PlatformID: "app-1",
		RepositoryID:  42,
		Key:           "PRJ",
		Name:          "Proj One",
		Status:        domain.PlatformStatusActive,
		Visibility:    domain.PlatformVisibilityInternal,
		OwnerUserID:   "owner",
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	resp := projectResponse(p)
	if resp["id"] != "proj-1" || resp["key"] != "PRJ" || resp["platform_id"] != "app-1" {
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

func TestProjectRepositoryResponse(t *testing.T) {
	linked := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	link := domain.ProjectRepository{
		ProjectID:    "proj-1",
		RepositoryID: 42,
		Role:         "primary",
		LinkedAt:     linked,
	}
	resp := projectRepositoryResponse(link)
	if resp["project_id"] != "proj-1" || resp["repository_id"] != int64(42) {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["role"] != "primary" {
		t.Fatalf("role = %v", resp["role"])
	}
}

func TestProjectModel_DefaultHybrid(t *testing.T) {
	h := NewPlatformHandler(PlatformConfig{})
	if got := h.projectModel(); got != "hybrid" {
		t.Fatalf("default = %q", got)
	}
}

func TestProjectModel_KnownValues(t *testing.T) {
	cases := []string{"legacy", "v2", "hybrid"}
	for _, mode := range cases {
		h := NewPlatformHandler(PlatformConfig{ProjectModel: mode})
		if got := h.projectModel(); got != mode {
			t.Errorf("ProjectModel=%q got %q", mode, got)
		}
	}
}

func TestProjectModel_UpperCaseNormalized(t *testing.T) {
	h := NewPlatformHandler(PlatformConfig{ProjectModel: "LEGACY"})
	if got := h.projectModel(); got != "legacy" {
		t.Fatalf("got %q", got)
	}
}

func TestProjectModel_UnknownFallsBackToHybrid(t *testing.T) {
	h := NewPlatformHandler(PlatformConfig{ProjectModel: "weird"})
	if got := h.projectModel(); got != "hybrid" {
		t.Fatalf("got %q", got)
	}
}

func TestAllowLegacyProjectRoutes(t *testing.T) {
	cases := map[string]bool{
		"legacy": true, "hybrid": true, "v2": false, "weird": true,
	}
	for mode, want := range cases {
		h := NewPlatformHandler(PlatformConfig{ProjectModel: mode})
		if got := h.allowLegacyProjectRoutes(); got != want {
			t.Errorf("mode=%q got %v want %v", mode, got, want)
		}
	}
}

func TestAllowV2ProjectRoutes(t *testing.T) {
	cases := map[string]bool{
		"legacy": false, "hybrid": true, "v2": true, "weird": true,
	}
	for mode, want := range cases {
		h := NewPlatformHandler(PlatformConfig{ProjectModel: mode})
		if got := h.allowV2ProjectRoutes(); got != want {
			t.Errorf("mode=%q got %v want %v", mode, got, want)
		}
	}
}

func TestPlatformStoreOrUnavailable_NilReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)
	got, ok := h.PlatformStoreOrUnavailable(c)
	if ok {
		t.Fatal("expected ok=false")
	}
	if got != nil {
		t.Fatal("expected nil store ref")
	}
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationKeyPattern(t *testing.T) {
	cases := map[string]bool{
		"APP":              true,
		"app1":             true,
		"AB123":             true,
		"":                 false,
		"longerthan10char": false,
		"with-dash":        false,
		"with_score":       false,
		"contains space":   false,
	}
	for input, want := range cases {
		got := platformKeyPattern.MatchString(input)
		if got != want {
			t.Errorf("platformKeyPattern(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("hello world", "world") {
		t.Fatal("should find world")
	}
	if !containsAny("foo bar baz", "qux", "baz") {
		t.Fatal("should find baz")
	}
	if containsAny("hello", "world", "foo") {
		t.Fatal("should not find either")
	}
	if containsAny("", "x") {
		t.Fatal("empty string contains nothing")
	}
	if !containsAny("ab", "") {
		// empty substring matches everything
		t.Fatal("empty substring should match")
	}
}
