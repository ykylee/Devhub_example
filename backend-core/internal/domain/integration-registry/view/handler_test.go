package view

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type fakeIntegrationAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeIntegrationAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_int_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewIntegrationHandler_NonNil(t *testing.T) {
	h := NewIntegrationHandler(IntegrationConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewIntegrationHandler_ConfigPropagation(t *testing.T) {
	cfg := IntegrationConfig{AuditStore: &fakeIntegrationAuditStore{}}
	h := NewIntegrationHandler(cfg)
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewIntegrationHandler(IntegrationConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeIntegrationAuditStore{}
	h := NewIntegrationHandler(IntegrationConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "int.test", "integration", "int-1", nil)
	if got.AuditID != "audit_int_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "int.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeIntegrationAuditStore{}
	h := NewIntegrationHandler(IntegrationConfig{AuditStore: store})
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

	store := &fakeIntegrationAuditStore{err: errors.New("db_down")}
	h := NewIntegrationHandler(IntegrationConfig{AuditStore: store})
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

func TestIntegrationStoreOrUnavailable_NilReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewIntegrationHandler(IntegrationConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	got, ok := h.IntegrationStoreOrUnavailable(c)
	if ok {
		t.Fatal("expected ok=false")
	}
	if got != nil {
		t.Fatal("expected nil store ref")
	}
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "integration store is not configured") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestActorLogin_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("devhub_actor_login", "alice")
	if got := actorLogin(c); got != "alice" {
		t.Fatalf("got %q", got)
	}
}

func TestActorLogin_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := actorLogin(c); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestActorLogin_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("devhub_actor_login", 123)
	if got := actorLogin(c); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestFirstHeader_PickFirstNonEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x", nil)
	c.Request.Header.Set("X-A", "a-val")
	c.Request.Header.Set("X-B", "b-val")
	if got := firstHeader(c, "X-A", "X-B"); got != "a-val" {
		t.Fatalf("got %q", got)
	}
}

func TestFirstHeader_FallsThroughEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x", nil)
	c.Request.Header.Set("X-B", "b-val")
	if got := firstHeader(c, "X-A", "X-B"); got != "b-val" {
		t.Fatalf("got %q", got)
	}
}

func TestFirstHeader_NoneSetReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x", nil)
	if got := firstHeader(c, "X-A", "X-B"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestIntegrationProviderResponse_WriteOnlyFields(t *testing.T) {
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	p := domain.IntegrationProvider{
		ID:                  "prov-1",
		ProviderKey:         "gitea",
		ProviderType:        domain.IntegrationProviderType("scm"),
		DisplayName:         "Gitea",
		Enabled:             true,
		AuthMode:            domain.IntegrationAuthMode("token"),
		CredentialsRef:      "hmac_sha256:secret",
		Capabilities:        []string{"pull", "sync"},
		SyncStatus:          "ok",
		BaseURL:             "https://gitea.example.com",
		APIToken:            "super-secret",
		AuthUsername:        "alice",
		AuthSecret:          "secret-x",
		WebhookSecret:       "wh-secret",
		PullIntervalSeconds: 60,
		CreatedAt:           created,
		UpdatedAt:           updated,
	}
	resp := integrationProviderResponse(p)
	if resp["provider_id"] != "prov-1" || resp["provider_key"] != "gitea" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["api_token_set"] != true {
		t.Fatalf("api_token_set should be true: %v", resp["api_token_set"])
	}
	if _, ok := resp["api_token"]; ok {
		t.Fatal("raw api_token must not be exposed")
	}
	if resp["auth_secret_set"] != true {
		t.Fatalf("auth_secret_set: %v", resp["auth_secret_set"])
	}
	if _, ok := resp["auth_secret"]; ok {
		t.Fatal("raw auth_secret must not be exposed")
	}
	if resp["webhook_secret_set"] != true {
		t.Fatalf("webhook_secret_set: %v", resp["webhook_secret_set"])
	}
	if _, ok := resp["webhook_secret"]; ok {
		t.Fatal("raw webhook_secret must not be exposed")
	}
}

func TestIntegrationProviderResponse_EmptySecretsBoolFalse(t *testing.T) {
	p := domain.IntegrationProvider{ID: "prov-1"}
	resp := integrationProviderResponse(p)
	if resp["api_token_set"] != false || resp["auth_secret_set"] != false || resp["webhook_secret_set"] != false {
		t.Fatalf("empty secrets should give false flags: %+v", resp)
	}
}

func TestIntegrationBindingResponse_FieldMapping(t *testing.T) {
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	b := domain.IntegrationBinding{
		ID:          "bind-1",
		ScopeType:   domain.IntegrationScopeType("platform"),
		ScopeID:     "app-1",
		ProviderID:  "prov-1",
		ExternalKey: "PROJ-1",
		Policy:      domain.IntegrationPolicy("default"),
		Enabled:     true,
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	resp := integrationBindingResponse(b)
	if resp["binding_id"] != "bind-1" || resp["scope_type"] != "platform" || resp["scope_id"] != "app-1" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["provider_id"] != "prov-1" || resp["external_key"] != "PROJ-1" {
		t.Fatalf("provider/key: %+v", resp)
	}
	if resp["policy"] != "default" || resp["enabled"] != true {
		t.Fatalf("policy/enabled: %+v", resp)
	}
}

func TestNullableRFC3339(t *testing.T) {
	if got := nullableRFC3339(nil); got != nil {
		t.Fatalf("nil -> nil, got %v", got)
	}
	tm := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	if got := nullableRFC3339(&tm); got != "2026-05-29T12:00:00Z" {
		t.Fatalf("got %v", got)
	}
}

func TestEmptyAsNil(t *testing.T) {
	if got := emptyAsNil(""); got != nil {
		t.Fatalf("empty -> nil, got %v", got)
	}
	if got := emptyAsNil("   "); got != nil {
		t.Fatalf("whitespace -> nil, got %v", got)
	}
	if got := emptyAsNil("x"); got != "x" {
		t.Fatalf("got %v", got)
	}
}

func TestParseOptionalBool_Empty(t *testing.T) {
	v, ok := parseOptionalBool("")
	if !ok {
		t.Fatal("empty should be ok")
	}
	if v != nil {
		t.Fatalf("empty should give nil ptr, got %v", v)
	}
}

func TestParseOptionalBool_True(t *testing.T) {
	v, ok := parseOptionalBool("true")
	if !ok {
		t.Fatal("true should be ok")
	}
	if v == nil || *v != true {
		t.Fatalf("got %v", v)
	}
}

func TestParseOptionalBool_False(t *testing.T) {
	v, ok := parseOptionalBool("false")
	if !ok {
		t.Fatal("false should be ok")
	}
	if v == nil || *v != false {
		t.Fatalf("got %v", v)
	}
}

func TestParseOptionalBool_Invalid(t *testing.T) {
	v, ok := parseOptionalBool("not-bool")
	if ok {
		t.Fatal("invalid should not be ok")
	}
	if v != nil {
		t.Fatalf("invalid -> nil ptr, got %v", v)
	}
}

func TestNormalizeProviderSDKKey_Variants(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"gitea":      "gitea",
		"GITEA":      "gitea",
		"  gitea  ":  "gitea",
		"gitea-internal": "gitea",
		"github-enterprise": "github",
		"jira-cloud-x": "jira",
	}
	for in, want := range cases {
		if got := normalizeProviderSDKKey(in); got != want {
			t.Errorf("normalizeProviderSDKKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidBaseURL_Empty(t *testing.T) {
	if !validBaseURL("") {
		t.Fatal("empty must be allowed")
	}
	if !validBaseURL("  ") {
		t.Fatal("whitespace-only must be allowed")
	}
}

func TestValidBaseURL_ValidHTTP(t *testing.T) {
	if !validBaseURL("http://example.com") {
		t.Fatal("http URL must be valid")
	}
	if !validBaseURL("https://example.com") {
		t.Fatal("https URL must be valid")
	}
	if !validBaseURL("https://example.com:443/path") {
		t.Fatal("URL with port/path must be valid")
	}
}

func TestValidBaseURL_InvalidScheme(t *testing.T) {
	if validBaseURL("ftp://example.com") {
		t.Fatal("ftp must be rejected")
	}
}

func TestValidBaseURL_HostMissing(t *testing.T) {
	if validBaseURL("https://") {
		t.Fatal("scheme-only must be rejected (codex PR #352 P2)")
	}
}

func TestValidIntegrationProviderTypes(t *testing.T) {
	for _, k := range []string{"alm", "scm", "ci_cd", "doc", "infra", "task_tracker"} {
		if !validIntegrationProviderTypes[k] {
			t.Errorf("expected %q valid", k)
		}
	}
	if validIntegrationProviderTypes["unknown"] {
		t.Error("unknown must be invalid")
	}
}

func TestValidIntegrationAuthModes(t *testing.T) {
	for _, k := range []string{"token", "basic", "oauth2", "app_password", "agent"} {
		if !validIntegrationAuthModes[k] {
			t.Errorf("expected %q valid", k)
		}
	}
	if validIntegrationAuthModes["weird"] {
		t.Error("weird must be invalid")
	}
}

func TestExternalTaskItemResponse_FieldMapping(t *testing.T) {
	fetched := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	seq := int64(1)
	item := domain.ExternalTaskItem{
		ID:               "task-1",
		ProviderID:       "prov-1",
		ExternalID:       "EXT-1",
		Title:            "Title",
		RawStatus:        "Open",
		NormalizedStatus: "open",
		Priority:         "High",
		Assignee:         "alice",
		Reporter:         "bob",
		URL:              "https://example.com/EXT-1",
		Labels:           []string{"bug"},
		WebhookSeq:       &seq,
		FetchedAt:        fetched,
	}
	resp := externalTaskItemResponse(item)
	if resp["id"] != "task-1" || resp["external_id"] != "EXT-1" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["normalized_status"] != "open" {
		t.Fatalf("normalized_status: %v", resp["normalized_status"])
	}
	if resp["priority"] != "High" {
		t.Fatalf("priority: %v", resp["priority"])
	}
}

func TestExternalTaskItemResponse_NormalizedStatusOmittedIfEmpty(t *testing.T) {
	item := domain.ExternalTaskItem{ID: "x", FetchedAt: time.Now().UTC()}
	resp := externalTaskItemResponse(item)
	if _, ok := resp["normalized_status"]; ok {
		t.Fatalf("normalized_status should be omitted when empty: %+v", resp)
	}
}

func TestVerifyIntegrationWebhookSignature_EmptySignature(t *testing.T) {
	p := domain.IntegrationProvider{CredentialsRef: "secret"}
	if verifyIntegrationWebhookSignature(p, []byte("body"), "") {
		t.Fatal("empty signature must reject")
	}
}

func TestVerifyIntegrationWebhookSignature_EmptyCredentials(t *testing.T) {
	p := domain.IntegrationProvider{}
	if verifyIntegrationWebhookSignature(p, []byte("body"), "sig") {
		t.Fatal("empty creds must reject")
	}
}

func TestVerifyIntegrationWebhookSignature_SharedTokenMatch(t *testing.T) {
	p := domain.IntegrationProvider{CredentialsRef: "shared-token-x"}
	if !verifyIntegrationWebhookSignature(p, []byte("body"), "shared-token-x") {
		t.Fatal("matching shared token must pass")
	}
}

func TestVerifyIntegrationWebhookSignature_SharedTokenMismatch(t *testing.T) {
	p := domain.IntegrationProvider{CredentialsRef: "shared-token-x"}
	if verifyIntegrationWebhookSignature(p, []byte("body"), "wrong-token") {
		t.Fatal("non-matching token must reject")
	}
}

func TestVerifyIntegrationWebhookSignature_BearerTokenMatch(t *testing.T) {
	p := domain.IntegrationProvider{CredentialsRef: "shared-token-x"}
	if !verifyIntegrationWebhookSignature(p, []byte("body"), "Bearer shared-token-x") {
		t.Fatal("Bearer-prefixed match must pass")
	}
}

func TestVerifyProviderSDKWebhookSignature_EmptySecretReturnsFalse(t *testing.T) {
	p := domain.IntegrationProvider{ProviderKey: "gitea"}
	if verifyProviderSDKWebhookSignature(p, []byte("body"), "sig", "") {
		t.Fatal("empty secret must reject")
	}
}

func TestVerifyProviderSDKWebhookSignature_UnknownVendorReturnsFalse(t *testing.T) {
	p := domain.IntegrationProvider{ProviderKey: "weird-vendor", CredentialsRef: "x"}
	if verifyProviderSDKWebhookSignature(p, []byte("body"), "sig", "") {
		t.Fatal("unknown vendor must reject")
	}
}

func TestIntegrationResponse_FieldMapping(t *testing.T) {
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	pi := domain.ProjectIntegration{
		ID:              "int-1",
		Scope:           domain.IntegrationScope("project"),
		ProjectID:       "proj-1",
		IntegrationType: domain.IntegrationType("jira"),
		ExternalKey:     "JIRA-1",
		URL:             "https://gitea.example.com",
		CreatedAt:       created,
		UpdatedAt:       created,
	}
	resp := integrationResponse(pi)
	if resp["id"] != "int-1" || resp["project_id"] != "proj-1" {
		t.Fatalf("basic: %+v", resp)
	}
	if resp["scope"] != "project" || resp["integration_type"] != "jira" {
		t.Fatalf("scope/type: %+v", resp)
	}
	if resp["external_key"] != "JIRA-1" {
		t.Fatalf("external_key: %v", resp["external_key"])
	}
}

// --- fakeIntegrationStore for Handler tests -------------------------

type fakeIntegrationStore struct {
	IntegrationStore // Embed to automatically satisfy all interface methods

	integrations map[string]domain.ProjectIntegration
	errList      error
	errGet       error
	errCreate    error
	errUpdate    error
	errDelete    error

	// Dynamic mock functions
	listProvidersFunc       func(ctx context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error)
	getProviderByIDFunc     func(ctx context.Context, id string) (domain.IntegrationProvider, error)
	getProviderByKeyFunc    func(ctx context.Context, key string) (domain.IntegrationProvider, error)
	createProviderFunc      func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error)
	updateProviderFunc      func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error)
	deleteProviderFunc      func(ctx context.Context, id string) error
	createSyncJobFunc       func(ctx context.Context, providerID string, triggerBy string) (string, error)
	listBindingsFunc        func(ctx context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error)
	getBindingByIDFunc      func(ctx context.Context, id string) (domain.IntegrationBinding, error)
	createBindingFunc       func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error)
	updateBindingFunc       func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error)
	deleteBindingFunc       func(ctx context.Context, id string) error
}

func (s *fakeIntegrationStore) ListIntegrations(_ context.Context, _ store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error) {
	if s.errList != nil {
		return nil, 0, s.errList
	}
	out := []domain.ProjectIntegration{}
	for _, v := range s.integrations {
		out = append(out, v)
	}
	return out, len(out), nil
}

func (s *fakeIntegrationStore) GetIntegration(_ context.Context, id string) (domain.ProjectIntegration, error) {
	if s.errGet != nil {
		return domain.ProjectIntegration{}, s.errGet
	}
	if v, ok := s.integrations[id]; ok {
		return v, nil
	}
	return domain.ProjectIntegration{}, store.ErrNotFound
}

func (s *fakeIntegrationStore) CreateIntegration(_ context.Context, p domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	if s.errCreate != nil {
		return domain.ProjectIntegration{}, s.errCreate
	}
	if p.ID == "" {
		p.ID = "int-" + p.ExternalKey
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	if s.integrations == nil {
		s.integrations = make(map[string]domain.ProjectIntegration)
	}
	s.integrations[p.ID] = p
	return p, nil
}

func (s *fakeIntegrationStore) UpdateIntegration(_ context.Context, p domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	if s.errUpdate != nil {
		return domain.ProjectIntegration{}, s.errUpdate
	}
	if _, ok := s.integrations[p.ID]; !ok {
		return domain.ProjectIntegration{}, store.ErrNotFound
	}
	p.UpdatedAt = time.Now()
	s.integrations[p.ID] = p
	return p, nil
}

func (s *fakeIntegrationStore) DeleteIntegration(_ context.Context, id string) error {
	if s.errDelete != nil {
		return s.errDelete
	}
	if _, ok := s.integrations[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrations, id)
	return nil
}

func (s *fakeIntegrationStore) ListIntegrationProviders(ctx context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
	if s.listProvidersFunc != nil {
		return s.listProvidersFunc(ctx, opts)
	}
	return nil, 0, nil
}

func (s *fakeIntegrationStore) GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error) {
	if s.getProviderByIDFunc != nil {
		return s.getProviderByIDFunc(ctx, id)
	}
	return domain.IntegrationProvider{}, store.ErrNotFound
}

func (s *fakeIntegrationStore) GetIntegrationProviderByKey(ctx context.Context, key string) (domain.IntegrationProvider, error) {
	if s.getProviderByKeyFunc != nil {
		return s.getProviderByKeyFunc(ctx, key)
	}
	return domain.IntegrationProvider{}, store.ErrNotFound
}

func (s *fakeIntegrationStore) CreateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	if s.createProviderFunc != nil {
		return s.createProviderFunc(ctx, p)
	}
	return p, nil
}

func (s *fakeIntegrationStore) UpdateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	if s.updateProviderFunc != nil {
		return s.updateProviderFunc(ctx, p)
	}
	return p, nil
}

func (s *fakeIntegrationStore) DeleteIntegrationProvider(ctx context.Context, id string) error {
	if s.deleteProviderFunc != nil {
		return s.deleteProviderFunc(ctx, id)
	}
	return nil
}

func (s *fakeIntegrationStore) CreateIntegrationSyncJob(ctx context.Context, providerID string, triggerBy string) (string, error) {
	if s.createSyncJobFunc != nil {
		return s.createSyncJobFunc(ctx, providerID, triggerBy)
	}
	return "job-123", nil
}

func (s *fakeIntegrationStore) ListIntegrationBindings(ctx context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
	if s.listBindingsFunc != nil {
		return s.listBindingsFunc(ctx, opts)
	}
	return nil, 0, nil
}

func (s *fakeIntegrationStore) GetIntegrationBindingByID(ctx context.Context, id string) (domain.IntegrationBinding, error) {
	if s.getBindingByIDFunc != nil {
		return s.getBindingByIDFunc(ctx, id)
	}
	return domain.IntegrationBinding{}, store.ErrNotFound
}

func (s *fakeIntegrationStore) CreateIntegrationBinding(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	if s.createBindingFunc != nil {
		return s.createBindingFunc(ctx, p)
	}
	return p, nil
}

func (s *fakeIntegrationStore) UpdateIntegrationBinding(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	if s.updateBindingFunc != nil {
		return s.updateBindingFunc(ctx, p)
	}
	return p, nil
}

func (s *fakeIntegrationStore) DeleteIntegrationBinding(ctx context.Context, id string) error {
	if s.deleteBindingFunc != nil {
		return s.deleteBindingFunc(ctx, id)
	}
	return nil
}

// --- fakeWebhookEventStore for Handler tests -------------------------

type fakeWebhookEventStore struct {
	WebhookEventStore
	saveWebhookEventFunc func(ctx context.Context, ev store.WebhookEvent) (int64, error)
}

func (f *fakeWebhookEventStore) SaveWebhookEvent(ctx context.Context, ev store.WebhookEvent) (int64, error) {
	if f.saveWebhookEventFunc != nil {
		return f.saveWebhookEventFunc(ctx, ev)
	}
	return 1, nil
}

// --- fakeExternalTaskStore for Handler tests -------------------------

type fakeExternalTaskStore struct {
	ExternalTaskStore

	upsertFunc       func(ctx context.Context, t domain.ExternalTaskItem) (domain.ExternalTaskItem, error)
	deleteFunc       func(ctx context.Context, providerID, externalID string) error
	listFunc         func(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error)
	getFunc          func(ctx context.Context, id string) (domain.ExternalTaskItem, error)
	nextSeqFunc      func(ctx context.Context) (int64, error)
	detectGapsFunc   func(ctx context.Context, providerID string) (int64, error)
	updatePulledFunc func(ctx context.Context, providerID string, pulledAt time.Time) error
	listTrackersFunc func(ctx context.Context) ([]domain.IntegrationProvider, error)
}

func (f *fakeExternalTaskStore) UpsertExternalTaskItem(ctx context.Context, t domain.ExternalTaskItem) (domain.ExternalTaskItem, error) {
	if f.upsertFunc != nil {
		return f.upsertFunc(ctx, t)
	}
	return t, nil
}

func (f *fakeExternalTaskStore) SoftDeleteExternalTaskItem(ctx context.Context, providerID, externalID string) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, providerID, externalID)
	}
	return nil
}

func (f *fakeExternalTaskStore) ListExternalTaskItems(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, opts)
	}
	return nil, 0, nil
}

func (f *fakeExternalTaskStore) GetExternalTaskItemByID(ctx context.Context, id string) (domain.ExternalTaskItem, error) {
	if f.getFunc != nil {
		return f.getFunc(ctx, id)
	}
	return domain.ExternalTaskItem{}, store.ErrNotFound
}

func (f *fakeExternalTaskStore) NextWebhookSeq(ctx context.Context) (int64, error) {
	if f.nextSeqFunc != nil {
		return f.nextSeqFunc(ctx)
	}
	return 1, nil
}

func (f *fakeExternalTaskStore) DetectWebhookSeqGaps(ctx context.Context, providerID string) (int64, error) {
	if f.detectGapsFunc != nil {
		return f.detectGapsFunc(ctx, providerID)
	}
	return 0, nil
}

func (f *fakeExternalTaskStore) UpdateProviderLastPulledAt(ctx context.Context, providerID string, pulledAt time.Time) error {
	if f.updatePulledFunc != nil {
		return f.updatePulledFunc(ctx, providerID, pulledAt)
	}
	return nil
}

func (f *fakeExternalTaskStore) ListTaskTrackers(ctx context.Context) ([]domain.IntegrationProvider, error) {
	if f.listTrackersFunc != nil {
		return f.listTrackersFunc(ctx)
	}
	return nil, nil
}

// --- Integration Handler CRUD Tests ----------------------------------

func TestListIntegrations_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("scope validation failure", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/integrations?scope=invalid", nil)

		h.ListIntegrations(c)

		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("integration_type validation failure", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/integrations?integration_type=invalid", nil)

		h.ListIntegrations(c)

		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("limit/offset validation failure", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/integrations?limit=200", nil)
		h.ListIntegrations(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 limit, got %d", rec.Code)
		}

		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("GET", "/integrations?offset=-5", nil)
		h.ListIntegrations(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400 offset, got %d", rec2.Code)
		}
	})

	t.Run("store general error returns 500", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errList: errors.New("db lost")}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/integrations", nil)

		h.ListIntegrations(c)

		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("success returns 200 with data", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {
					ID:              "int-1",
					Scope:           "platform",
					PlatformID:   "app-1",
					IntegrationType: "jira",
					ExternalKey:     "JIRA-1",
					URL:             "https://example.com",
					Policy:          "summary_only",
				},
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/integrations", nil)

		h.ListIntegrations(c)

		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "JIRA-1") {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})
}

func TestCreateIntegration_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid json payload aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader("{bad-json"))

		h.CreateIntegration(c)

		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("invalid scope type aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{"scope":"invalid"}`))

		h.CreateIntegration(c)

		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing scope application target aborts 422", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"platform",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))

		h.CreateIntegration(c)

		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("store ErrConflict aborts 409", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errCreate: store.ErrConflict}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"platform",
			"platform_id":"app-1",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))

		h.CreateIntegration(c)

		if rec.Code != 409 {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("successful creation project scope records audit", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{}
		audit := &fakeIntegrationAuditStore{}
		h := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			AuditStore:       audit,
		})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))

		h.CreateIntegration(c)

		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("expected 1 audit log, got %d", len(audit.created))
		}
		c0 := audit.created[0]
		if c0.Action != "integration.created" || c0.TargetID != "int-J-1" {
			t.Fatalf("wrong audit mapping: %+v", c0)
		}
	})

	t.Run("invalid integration type aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"invalid",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))
		h.CreateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("invalid policy aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"jira",
			"policy":"invalid",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))
		h.CreateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing external_key aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"",
			"url":"https://example.com"
		}`))
		h.CreateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing url aborts 400", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":""
		}`))
		h.CreateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing project_id for project scope aborts 422", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: &fakeIntegrationStore{}})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))
		h.CreateIntegration(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("store general error on CreateIntegration returns 500", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errCreate: errors.New("db lost")}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/integrations", strings.NewReader(`{
			"scope":"project",
			"project_id":"proj-1",
			"integration_type":"jira",
			"policy":"summary_only",
			"external_key":"J-1",
			"url":"https://example.com"
		}`))
		h.CreateIntegration(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestUpdateIntegration_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("store ErrNotFound on lookup aborts 404", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errGet: store.ErrNotFound}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com"}`))

		h.UpdateIntegration(c)

		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("store general error on lookup aborts 500", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errGet: errors.New("db lost")}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com"}`))

		h.UpdateIntegration(c)

		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("successful update triggers audit", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {
					ID:              "int-1",
					Scope:           "platform",
					PlatformID:   "app-1",
					IntegrationType: "jira",
					ExternalKey:     "J-1",
					URL:             "https://old.com",
					Policy:          "summary_only",
				},
			},
		}
		audit := &fakeIntegrationAuditStore{}
		h := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			AuditStore:       audit,
		})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com","policy":"execution_system"}`))

		h.UpdateIntegration(c)

		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("expected 1 audit log, got %d", len(audit.created))
		}
		c0 := audit.created[0]
		if c0.Action != "integration.updated" || c0.TargetID != "int-1" {
			t.Fatalf("wrong audit mapping: %+v", c0)
		}
	})

	t.Run("empty external_key on update aborts 400", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"external_key":""}`))
		h.UpdateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 external_key empty, got %d", rec.Code)
		}
	})

	t.Run("empty url on update aborts 400", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":""}`))
		h.UpdateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 url empty, got %d", rec.Code)
		}
	})

	t.Run("invalid policy on update aborts 400", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"policy":"invalid"}`))
		h.UpdateIntegration(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 policy invalid, got %d", rec.Code)
		}
	})

	t.Run("store update ErrNotFound aborts 404", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
			errUpdate: store.ErrNotFound,
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com"}`))
		h.UpdateIntegration(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 update not found, got %d", rec.Code)
		}
	})

	t.Run("store update ErrConflict aborts 409", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
			errUpdate: store.ErrConflict,
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com"}`))
		h.UpdateIntegration(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409 update conflict, got %d", rec.Code)
		}
	})

	t.Run("store general update error aborts 500", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
			errUpdate: errors.New("db lost"),
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("PATCH", "/integrations/int-1", strings.NewReader(`{"url":"https://new.com"}`))
		h.UpdateIntegration(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 update error, got %d", rec.Code)
		}
	})
}

func TestDeleteIntegration_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("store ErrNotFound aborts 404", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errDelete: store.ErrNotFound}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("DELETE", "/integrations/int-1", nil)

		h.DeleteIntegration(c)

		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("successful deletion records audit", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			integrations: map[string]domain.ProjectIntegration{
				"int-1": {ID: "int-1"},
			},
		}
		audit := &fakeIntegrationAuditStore{}
		h := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			AuditStore:       audit,
		})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("DELETE", "/integrations/int-1", nil)

		h.DeleteIntegration(c)

		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if len(audit.created) != 1 {
			t.Fatalf("expected 1 audit log, got %d", len(audit.created))
		}
		c0 := audit.created[0]
		if c0.Action != "integration.deleted" || c0.TargetID != "int-1" {
			t.Fatalf("wrong audit mapping: %+v", c0)
		}
	})

	t.Run("store general error aborts 500", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{errDelete: errors.New("db lost")}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "integration_id", Value: "int-1"}}
		c.Request = httptest.NewRequest("DELETE", "/integrations/int-1", nil)
		h.DeleteIntegration(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 delete error, got %d", rec.Code)
		}
	})
}

func TestIntegrationProviders_Handlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ListIntegrationProviders - validation failures and success", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			listProvidersFunc: func(ctx context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
				if opts.ProviderType == "scm" {
					return []domain.IntegrationProvider{
						{ID: "prov-1", ProviderKey: "gitea", ProviderType: "scm", DisplayName: "Gitea SCM"},
					}, 1, nil
				}
				return []domain.IntegrationProvider{
					{ID: "prov-1", ProviderKey: "gitea", ProviderType: "scm", DisplayName: "Gitea SCM"},
					{ID: "prov-2", ProviderKey: "jira", ProviderType: "task_tracker", DisplayName: "Jira Tracker"},
				}, 2, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})

		// 1. invalid provider type
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers?provider_type=invalid", nil)
		h.ListIntegrationProviders(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. invalid enabled bool
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers?enabled=not-bool", nil)
		h.ListIntegrationProviders(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 3. invalid limit
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers?limit=200", nil)
		h.ListIntegrationProviders(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 4. invalid offset
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers?offset=-5", nil)
		h.ListIntegrationProviders(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 5. Success filtering provider_type=scm
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers?provider_type=scm", nil)
		h.ListIntegrationProviders(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "Gitea SCM") || strings.Contains(rec.Body.String(), "Jira Tracker") {
			t.Fatalf("body mismatch: %s", rec.Body.String())
		}

		// 6. general store error
		storeErr := &fakeIntegrationStore{
			listProvidersFunc: func(ctx context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
				return nil, 0, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/providers", nil)
		hErr.ListIntegrationProviders(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("CreateIntegrationProvider - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			createProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				p.ID = "prov-created-id"
				return p, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. JSON parse failure
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader("{bad-json"))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. Missing provider_key
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 3. Invalid provider_type
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"invalid","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 4. Invalid auth_mode
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"invalid","credentials_ref":"secret"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 5. Missing display_name
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","auth_mode":"token","credentials_ref":"secret"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 DisplayName, got %d", rec.Code)
		}

		// 6. Missing credentials_ref
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 credentials_ref, got %d", rec.Code)
		}

		// 7. Invalid base_url
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret","base_url":"ftp://not-http"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 base_url, got %d", rec.Code)
		}

		// 8. Invalid auth_token_url
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret","auth_token_url":"ftp://not-http"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 auth_token_url, got %d", rec.Code)
		}

		// 9. Store Conflict
		storeConflict := &fakeIntegrationStore{
			createProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, store.ErrConflict
			},
		}
		hConflict := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeConflict})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret"}`))
		hConflict.CreateIntegrationProvider(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409, got %d", rec.Code)
		}

		// 10. Store general error
		storeErr := &fakeIntegrationStore{
			createProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret"}`))
		hErr.CreateIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 11. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/providers", strings.NewReader(`{"provider_key":"gitea","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret"}`))
		h.CreateIntegrationProvider(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("UpdateIntegrationProvider - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", DisplayName: "Old Name", CredentialsRef: "old-secret"}, nil
			},
			updateProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return p, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. JSON parsing failure
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader("{bad-json"))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. Provider not found
		storeNotFound := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
		}
		hNotFound := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeNotFound})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":"New"}`))
		hNotFound.UpdateIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		// 3. Store error on Get
		storeErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":"New"}`))
		hErr.UpdateIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 4. empty display_name
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":""}`))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 DisplayName empty, got %d", rec.Code)
		}

		// 5. empty credentials_ref
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"credentials_ref":""}`))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 credentials_ref empty, got %d", rec.Code)
		}

		// 6. invalid base_url
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"base_url":"ftp://not-http"}`))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 base_url, got %d", rec.Code)
		}

		// 7. invalid auth_token_url
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"auth_token_url":"ftp://not-http"}`))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 auth_token_url, got %d", rec.Code)
		}

		// 8. store ErrNotFound on Update
		storeUpdateNotFound := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id}, nil
			},
			updateProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
		}
		hUpdateNotFound := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeUpdateNotFound})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":"New"}`))
		hUpdateNotFound.UpdateIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 update, got %d", rec.Code)
		}

		// 9. store general error on Update
		storeUpdateErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id}, nil
			},
			updateProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hUpdateErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeUpdateErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":"New"}`))
		hUpdateErr.UpdateIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 update, got %d", rec.Code)
		}

		// 10. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-1"}}
		c.Request = httptest.NewRequest("PATCH", "/providers/prov-1", strings.NewReader(`{"display_name":"New Name","enabled":false}`))
		h.UpdateIntegrationProvider(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("TestIntegrationConnection - validations and reachability mock", func(t *testing.T) {
		h := NewIntegrationHandler(IntegrationConfig{})

		// 1. JSON parsing failure
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test-connection", strings.NewReader("{bad-json"))
		h.TestIntegrationConnection(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. base_url empty
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test-connection", strings.NewReader(`{"base_url":""}`))
		h.TestIntegrationConnection(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 base_url empty, got %d", rec.Code)
		}

		// 3. invalid base_url
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test-connection", strings.NewReader(`{"base_url":"ftp://not-http"}`))
		h.TestIntegrationConnection(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 base_url invalid, got %d", rec.Code)
		}

		// 4. unreachable mock
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test-connection", strings.NewReader(`{"base_url":"http://127.0.0.1:9999/unreachable"}`))
		h.TestIntegrationConnection(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200 on unreachable, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"reachable":false`) {
			t.Fatalf("expected reachable=false, got %s", rec.Body.String())
		}

		// 5. reachable mock using httptest server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/test-connection", strings.NewReader(`{"base_url":"`+ts.URL+`"}`))
		h.TestIntegrationConnection(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200 reachable, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"reachable":true`) {
			t.Fatalf("expected reachable=true, got %s", rec.Body.String())
		}
	})

	t.Run("SyncIntegrationProvider - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				if id == "prov-scm" {
					return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", Enabled: true, Capabilities: []string{"pull", "sync"}}, nil
				}
				if id == "prov-disabled" {
					return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", Enabled: false}, nil
				}
				if id == "prov-non-scm" {
					return domain.IntegrationProvider{ID: id, ProviderKey: "jira", ProviderType: "task_tracker", Enabled: true}, nil
				}
				if id == "prov-no-caps" {
					return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", Enabled: true, Capabilities: []string{"none"}}, nil
				}
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
			createSyncJobFunc: func(ctx context.Context, providerID string, triggerBy string) (string, error) {
				if providerID == "prov-scm" {
					return "job-scm-id", nil
				}
				return "", store.ErrNotFound
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. Provider not found
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-missing"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-missing/sync", nil)
		h.SyncIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 provider missing, got %d", rec.Code)
		}

		// 2. Store lookup error
		storeErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-scm"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-scm/sync", nil)
		hErr.SyncIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 lookup, got %d", rec.Code)
		}

		// 3. Provider disabled
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-disabled"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-disabled/sync", nil)
		h.SyncIntegrationProvider(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409 provider disabled, got %d", rec.Code)
		}

		// 4. Provider type not SCM
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-non-scm"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-non-scm/sync", nil)
		h.SyncIntegrationProvider(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 non-SCM, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 5. Provider missing pull/sync capability
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-no-caps"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-no-caps/sync", nil)
		h.SyncIntegrationProvider(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 missing capabilities, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 6. create job not found
		storeJobNotFound := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", Enabled: true, Capabilities: []string{"pull", "sync"}}, nil
			},
			createSyncJobFunc: func(ctx context.Context, providerID string, triggerBy string) (string, error) {
				return "", store.ErrNotFound
			},
		}
		hJobNotFound := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeJobNotFound})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-scm"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-scm/sync", nil)
		hJobNotFound.SyncIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 job provider missing, got %d", rec.Code)
		}

		// 7. create job general error
		storeJobErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", Enabled: true, Capabilities: []string{"pull", "sync"}}, nil
			},
			createSyncJobFunc: func(ctx context.Context, providerID string, triggerBy string) (string, error) {
				return "", errors.New("db lost")
			},
		}
		hJobErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeJobErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-scm"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-scm/sync", nil)
		hJobErr.SyncIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 job error, got %d", rec.Code)
		}

		// 8. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-scm"}}
		c.Request = httptest.NewRequest("POST", "/providers/prov-scm/sync", nil)
		h.SyncIntegrationProvider(c)
		if rec.Code != 202 {
			t.Fatalf("expected 202 success, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("DeleteIntegrationProvider - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				if id == "prov-exists" {
					return domain.IntegrationProvider{ID: id, ProviderKey: "gitea", ProviderType: "scm", DisplayName: "Mock Gitea"}, nil
				}
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
			deleteProviderFunc: func(ctx context.Context, id string) error {
				if id == "prov-exists" {
					return nil
				}
				if id == "prov-conflict" {
					return store.ErrConflict
				}
				return store.ErrNotFound
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. Missing provider_id
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: ""}}
		c.Request = httptest.NewRequest("DELETE", "/providers/", nil)
		h.DeleteIntegrationProvider(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 empty provider_id, got %d", rec.Code)
		}

		// 2. Provider not found
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-missing"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-missing", nil)
		h.DeleteIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		// 3. store lookup error
		storeErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-exists", nil)
		hErr.DeleteIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 lookup, got %d", rec.Code)
		}

		// 4. store conflict (provider has active bindings)
		storeConflict := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id}, nil
			},
			deleteProviderFunc: func(ctx context.Context, id string) error {
				return store.ErrConflict
			},
		}
		hConflict := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeConflict})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-conflict"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-conflict", nil)
		hConflict.DeleteIntegrationProvider(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409 conflict, got %d", rec.Code)
		}

		// 5. delete Lookup race not found
		storeRace := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id}, nil
			},
			deleteProviderFunc: func(ctx context.Context, id string) error {
				return store.ErrNotFound
			},
		}
		hRace := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeRace})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-race"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-race", nil)
		hRace.DeleteIntegrationProvider(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 race deletion, got %d", rec.Code)
		}

		// 6. delete general error
		storeDeleteErr := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id}, nil
			},
			deleteProviderFunc: func(ctx context.Context, id string) error {
				return errors.New("db lost")
			},
		}
		hDeleteErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeDeleteErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-exists", nil)
		hDeleteErr.DeleteIntegrationProvider(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 delete, got %d", rec.Code)
		}

		// 7. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/providers/prov-exists", nil)
		h.DeleteIntegrationProvider(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200 success, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("IngestIntegrationProviderWebhook - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getProviderByKeyFunc: func(ctx context.Context, key string) (domain.IntegrationProvider, error) {
				if key == "gitea" {
					return domain.IntegrationProvider{ID: "prov-1", ProviderKey: "gitea", Enabled: true, CredentialsRef: "shared-token-x"}, nil
				}
				if key == "disabled" {
					return domain.IntegrationProvider{ID: "prov-2", ProviderKey: "disabled", Enabled: false}, nil
				}
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
			updateProviderFunc: func(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
				return p, nil
			},
		}
		evStore := &fakeWebhookEventStore{}
		h := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			EventStore:       evStore,
			AuditStore:       audit,
		})

		// 1. Provider not found
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "missing"}}
		c.Request = httptest.NewRequest("POST", "/providers/missing/webhook", nil)
		h.IngestIntegrationProviderWebhook(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		// 2. Provider lookup error
		storeErr := &fakeIntegrationStore{
			getProviderByKeyFunc: func(ctx context.Context, key string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "gitea"}}
		c.Request = httptest.NewRequest("POST", "/providers/gitea/webhook", nil)
		hErr.IngestIntegrationProviderWebhook(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 3. Provider disabled
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "disabled"}}
		c.Request = httptest.NewRequest("POST", "/providers/disabled/webhook", nil)
		h.IngestIntegrationProviderWebhook(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409, got %d", rec.Code)
		}

		// 4. Invalid Signature
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "gitea"}}
		c.Request = httptest.NewRequest("POST", "/providers/gitea/webhook", strings.NewReader("webhook payload body"))
		c.Request.Header.Set("X-Integration-Signature", "wrong-signature")
		h.IngestIntegrationProviderWebhook(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 5. Duplicate delivery
		evStoreDup := &fakeWebhookEventStore{
			saveWebhookEventFunc: func(ctx context.Context, ev store.WebhookEvent) (int64, error) {
				return 0, store.ErrDuplicateEvent
			},
		}
		hDup := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			EventStore:       evStoreDup,
		})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "gitea"}}
		c.Request = httptest.NewRequest("POST", "/providers/gitea/webhook", strings.NewReader("webhook payload body"))
		c.Request.Header.Set("X-Integration-Signature", "shared-token-x")
		hDup.IngestIntegrationProviderWebhook(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409 duplicate, got %d", rec.Code)
		}

		// 6. Save event general error
		evStoreErr := &fakeWebhookEventStore{
			saveWebhookEventFunc: func(ctx context.Context, ev store.WebhookEvent) (int64, error) {
				return 0, errors.New("db lost")
			},
		}
		hSaveErr := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore: storeVal,
			EventStore:       evStoreErr,
		})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "gitea"}}
		c.Request = httptest.NewRequest("POST", "/providers/gitea/webhook", strings.NewReader("webhook payload body"))
		c.Request.Header.Set("X-Integration-Signature", "shared-token-x")
		hSaveErr.IngestIntegrationProviderWebhook(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 save, got %d", rec.Code)
		}

		// 7. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "gitea"}}
		c.Request = httptest.NewRequest("POST", "/providers/gitea/webhook", strings.NewReader("webhook payload body"))
		c.Request.Header.Set("X-Integration-Signature", "shared-token-x")
		h.IngestIntegrationProviderWebhook(c)
		if rec.Code != 202 {
			t.Fatalf("expected 202 success, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestIntegrationBindings_Handlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ListIntegrationBindings - validations and success", func(t *testing.T) {
		storeVal := &fakeIntegrationStore{
			listBindingsFunc: func(ctx context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
				return []domain.IntegrationBinding{
					{ID: "bind-1", ScopeType: "platform", ScopeID: "app-1", ProviderID: "prov-1", ExternalKey: "KEY-1", Policy: "summary_only", Enabled: true},
				}, 1, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal})

		// 1. invalid scope_type
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings?scope_type=invalid", nil)
		h.ListIntegrationBindings(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. invalid provider_type
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings?provider_type=invalid", nil)
		h.ListIntegrationBindings(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 3. invalid enabled
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings?enabled=not-bool", nil)
		h.ListIntegrationBindings(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 4. invalid limit/offset
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings?limit=200", nil)
		h.ListIntegrationBindings(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 limit, got %d", rec.Code)
		}

		// 5. store error
		storeErr := &fakeIntegrationStore{
			listBindingsFunc: func(ctx context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
				return nil, 0, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings", nil)
		hErr.ListIntegrationBindings(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 6. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/bindings?scope_type=platform", nil)
		h.ListIntegrationBindings(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("CreateIntegrationBinding - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			createBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				p.ID = "bind-created-id"
				return p, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. JSON parsing failure
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader("{bad-json"))
		h.CreateIntegrationBinding(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. invalid scope_type
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"invalid","scope_id":"app-1","provider_id":"prov-1","external_key":"KEY-1","policy":"summary_only"}`))
		h.CreateIntegrationBinding(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 scope_type, got %d", rec.Code)
		}

		// 3. missing scope_id/provider_id/external_key
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"platform","provider_id":"prov-1","external_key":"KEY-1","policy":"summary_only"}`))
		h.CreateIntegrationBinding(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 missing scope_id, got %d", rec.Code)
		}

		// 4. unsupported policy
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"platform","scope_id":"app-1","provider_id":"prov-1","external_key":"KEY-1","policy":"invalid_policy"}`))
		h.CreateIntegrationBinding(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 unsupported policy, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 5. store conflict
		storeConflict := &fakeIntegrationStore{
			createBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, store.ErrConflict
			},
		}
		hConflict := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeConflict})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"platform","scope_id":"app-1","provider_id":"prov-1","external_key":"KEY-1","policy":"summary_only"}`))
		hConflict.CreateIntegrationBinding(c)
		if rec.Code != 409 {
			t.Fatalf("expected 409 conflict, got %d", rec.Code)
		}

		// 6. store general error
		storeErr := &fakeIntegrationStore{
			createBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"platform","scope_id":"app-1","provider_id":"prov-1","external_key":"KEY-1","policy":"summary_only"}`))
		hErr.CreateIntegrationBinding(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 7. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/bindings", strings.NewReader(`{"scope_type":"platform","scope_id":"app-1","provider_id":"prov-1","external_key":"KEY-1","policy":"summary_only"}`))
		h.CreateIntegrationBinding(c)
		if rec.Code != 201 {
			t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("UpdateIntegrationBinding - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id, ScopeType: "platform", ScopeID: "app-1", ProviderID: "prov-1", ExternalKey: "KEY-1", Policy: "summary_only"}, nil
			},
			updateBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				return p, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. JSON parsing failure
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader("{bad-json"))
		h.UpdateIntegrationBinding(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. Binding not found
		storeNotFound := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, store.ErrNotFound
			},
		}
		hNotFound := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeNotFound})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"summary_only"}`))
		hNotFound.UpdateIntegrationBinding(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		// 3. store get error
		storeErr := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"summary_only"}`))
		hErr.UpdateIntegrationBinding(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 lookup, got %d", rec.Code)
		}

		// 4. unsupported policy
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"invalid"}`))
		h.UpdateIntegrationBinding(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 policy, got %d", rec.Code)
		}

		// 5. store update not found
		storeUpdateNotFound := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id}, nil
			},
			updateBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, store.ErrNotFound
			},
		}
		hUpdateNotFound := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeUpdateNotFound})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"summary_only"}`))
		hUpdateNotFound.UpdateIntegrationBinding(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 update, got %d", rec.Code)
		}

		// 6. store update general error
		storeUpdateErr := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id}, nil
			},
			updateBindingFunc: func(ctx context.Context, p domain.IntegrationBinding) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{}, errors.New("db lost")
			},
		}
		hUpdateErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeUpdateErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"summary_only"}`))
		hUpdateErr.UpdateIntegrationBinding(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500 update, got %d", rec.Code)
		}

		// 7. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-1"}}
		c.Request = httptest.NewRequest("PATCH", "/bindings/bind-1", strings.NewReader(`{"policy":"execution_system","enabled":false}`))
		h.UpdateIntegrationBinding(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})

	t.Run("DeleteIntegrationBinding - validations, errors, success", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id}, nil
			},
			deleteBindingFunc: func(ctx context.Context, id string) error {
				if id == "bind-exists" {
					return nil
				}
				return store.ErrNotFound
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeVal, AuditStore: audit})

		// 1. Binding not found
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-missing"}}
		c.Request = httptest.NewRequest("DELETE", "/bindings/bind-missing", nil)
		h.DeleteIntegrationBinding(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		// 2. store delete error
		storeErr := &fakeIntegrationStore{
			deleteBindingFunc: func(ctx context.Context, id string) error {
				return errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/bindings/bind-exists", nil)
		hErr.DeleteIntegrationBinding(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 3. store delete not found (race)
		storeRace := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id}, nil
			},
			deleteBindingFunc: func(ctx context.Context, id string) error {
				return store.ErrNotFound
			},
		}
		hRace := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeRace})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-race"}}
		c.Request = httptest.NewRequest("DELETE", "/bindings/bind-race", nil)
		hRace.DeleteIntegrationBinding(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 race, got %d", rec.Code)
		}

		// 4. store delete general error
		storeDeleteErr := &fakeIntegrationStore{
			getBindingByIDFunc: func(ctx context.Context, id string) (domain.IntegrationBinding, error) {
				return domain.IntegrationBinding{ID: id}, nil
			},
			deleteBindingFunc: func(ctx context.Context, id string) error {
				return errors.New("db lost")
			},
		}
		hDeleteErr := NewIntegrationHandler(IntegrationConfig{IntegrationStore: storeDeleteErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/bindings/bind-exists", nil)
		hDeleteErr.DeleteIntegrationBinding(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 5. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "binding_id", Value: "bind-exists"}}
		c.Request = httptest.NewRequest("DELETE", "/bindings/bind-exists", nil)
		h.DeleteIntegrationBinding(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(audit.created) != 1 {
			t.Fatalf("audit logs = %d", len(audit.created))
		}
	})
}

func TestExternalTasks_Handlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ReceiveExternalTaskWebhook - validation and ingestion flow", func(t *testing.T) {
		audit := &fakeIntegrationAuditStore{}
		storeVal := &fakeIntegrationStore{
			getProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				if id == "prov-tracker" {
					return domain.IntegrationProvider{ID: id, ProviderType: "task_tracker", WebhookSecret: "secret-token"}, nil
				}
				if id == "prov-non-tracker" {
					return domain.IntegrationProvider{ID: id, ProviderType: "scm"}, nil
				}
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
		}
		taskStore := &fakeExternalTaskStore{
			nextSeqFunc: func(ctx context.Context) (int64, error) {
				return 42, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{
			IntegrationStore:  storeVal,
			ExternalTaskStore: taskStore,
			AuditStore:        audit,
		})

		// 1. provider_id empty
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: ""}}
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		// 2. provider unavailable
		hNoStores := NewIntegrationHandler(IntegrationConfig{})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		hNoStores.ReceiveExternalTaskWebhook(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}

		// 3. Provider not found
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-missing"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", nil)
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 4. Provider type mismatch
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-non-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", nil)
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 mismatch, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		// 5. Webhook secret mismatch (header empty)
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", nil)
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401 empty secret, got %d", rec.Code)
		}

		// 6. Webhook secret mismatch (header invalid)
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", nil)
		c.Request.Header.Set("X-Webhook-Secret", "wrong-secret")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 401 {
			t.Fatalf("expected 401 wrong secret, got %d", rec.Code)
		}
		if len(audit.created) != 1 || audit.created[0].Action != "external_task.auth_failed" {
			t.Fatalf("audit logic = %+v", audit.created)
		}

		// reset audit
		audit.created = nil

		// 7. invalid webhook payload JSON
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", strings.NewReader("{bad-json"))
		c.Request.Header.Set("X-Webhook-Secret", "secret-token")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 JSON, got %d", rec.Code)
		}

		// 8. invalid payload (missing event/external_id/title)
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", strings.NewReader(`{"event":"","external_id":"EXT-1","title":"T"}`))
		c.Request.Header.Set("X-Webhook-Secret", "secret-token")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 invalid payload, got %d", rec.Code)
		}

		// 9. unsupported event value
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", strings.NewReader(`{"event":"invalid_event","external_id":"EXT-1","title":"T"}`))
		c.Request.Header.Set("X-Webhook-Secret", "secret-token")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422 invalid event, got %d", rec.Code)
		}

		// 10. Process deleted event success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", strings.NewReader(`{"event":"deleted","external_id":"EXT-1","title":"T"}`))
		c.Request.Header.Set("X-Webhook-Secret", "secret-token")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 202 {
			t.Fatalf("expected 202 deletion, got %d", rec.Code)
		}
		if len(audit.created) != 1 || audit.created[0].Action != "external_task.deleted" {
			t.Fatalf("audit deleted mismatch: %+v", audit.created)
		}

		audit.created = nil

		// 11. Process upsert success (event: updated)
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "provider_id", Value: "prov-tracker"}}
		c.Request = httptest.NewRequest("POST", "/tasks/webhook", strings.NewReader(`{"event":"updated","external_id":"EXT-1","title":"Updated Title"}`))
		c.Request.Header.Set("X-Webhook-Secret", "secret-token")
		h.ReceiveExternalTaskWebhook(c)
		if rec.Code != 202 {
			t.Fatalf("expected 202 update, got %d", rec.Code)
		}
		if len(audit.created) != 1 || audit.created[0].Action != "external_task.updated" {
			t.Fatalf("audit updated mismatch: %+v", audit.created)
		}
	})

	t.Run("ListExternalTaskItems - validations and success", func(t *testing.T) {
		taskStore := &fakeExternalTaskStore{
			listFunc: func(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error) {
				return []domain.ExternalTaskItem{
					{ID: "task-1", ProviderID: "prov-1", ExternalID: "EXT-1", Title: "Task 1", RawStatus: "Open", FetchedAt: time.Now()},
				}, 1, nil
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{ExternalTaskStore: taskStore})

		// 1. store list error
		storeErr := &fakeExternalTaskStore{
			listFunc: func(ctx context.Context, opts store.ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error) {
				return nil, 0, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{ExternalTaskStore: storeErr})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/external-tasks", nil)
		hErr.ListExternalTaskItems(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 2. Success with paging & label parsing
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/external-tasks?labels=bug,critical&per_page=10&page=2", nil)
		h.ListExternalTaskItems(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GetExternalTaskItem - validations and success", func(t *testing.T) {
		taskStore := &fakeExternalTaskStore{
			getFunc: func(ctx context.Context, id string) (domain.ExternalTaskItem, error) {
				if id == "task-1" {
					return domain.ExternalTaskItem{ID: id, ProviderID: "prov-1", ExternalID: "EXT-1", Title: "Task 1", RawStatus: "Open", FetchedAt: time.Now()}, nil
				}
				return domain.ExternalTaskItem{}, store.ErrNotFound
			},
		}
		h := NewIntegrationHandler(IntegrationConfig{ExternalTaskStore: taskStore})

		// 1. task_id empty
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "task_id", Value: ""}}
		c.Request = httptest.NewRequest("GET", "/external-tasks/", nil)
		h.GetExternalTaskItem(c)
		if rec.Code != 400 {
			t.Fatalf("expected 400 empty task_id, got %d", rec.Code)
		}

		// 2. Task not found
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "task_id", Value: "task-missing"}}
		c.Request = httptest.NewRequest("GET", "/external-tasks/task-missing", nil)
		h.GetExternalTaskItem(c)
		if rec.Code != 404 {
			t.Fatalf("expected 404 task missing, got %d", rec.Code)
		}

		// 3. store error
		storeErr := &fakeExternalTaskStore{
			getFunc: func(ctx context.Context, id string) (domain.ExternalTaskItem, error) {
				return domain.ExternalTaskItem{}, errors.New("db lost")
			},
		}
		hErr := NewIntegrationHandler(IntegrationConfig{ExternalTaskStore: storeErr})
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "task_id", Value: "task-1"}}
		c.Request = httptest.NewRequest("GET", "/external-tasks/task-1", nil)
		hErr.GetExternalTaskItem(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		// 4. Success
		rec = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(rec)
		c.Params = []gin.Param{{Key: "task_id", Value: "task-1"}}
		c.Request = httptest.NewRequest("GET", "/external-tasks/task-1", nil)
		h.GetExternalTaskItem(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}



// ---------------------------------------------------------------------------
// validBaseURL edge cases (pure function, integration_registry.go)
// ---------------------------------------------------------------------------

func TestValidBaseURL_SchemeOnly(t *testing.T) {
	if validBaseURL("https://") {
		t.Fatal("expected false for scheme-only URL (no host)")
	}
	if validBaseURL("http://") {
		t.Fatal("expected false for http://")
	}
}

func TestValidBaseURL_Invalid(t *testing.T) {
	if validBaseURL("://invalid") {
		t.Fatal("expected false for malformed URL")
	}
}

// ---------------------------------------------------------------------------
// ExternalTaskStoreOrUnavailable — store available path
// ---------------------------------------------------------------------------

func TestExternalTaskStoreOrUnavailable_StoreAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewIntegrationHandler(IntegrationConfig{
		ExternalTaskStore: &fakeExternalTaskStore{},
	})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)

	storeI, ok := h.ExternalTaskStoreOrUnavailable(c)
	if !ok {
		t.Fatal("expected ok=true when store is configured")
	}
	if storeI == nil {
		t.Fatal("expected non-nil store")
	}
}
