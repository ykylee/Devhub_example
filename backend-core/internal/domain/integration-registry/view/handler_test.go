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
		ScopeType:   domain.IntegrationScopeType("application"),
		ScopeID:     "app-1",
		ProviderID:  "prov-1",
		ExternalKey: "PROJ-1",
		Policy:      domain.IntegrationPolicy("default"),
		Enabled:     true,
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	resp := integrationBindingResponse(b)
	if resp["binding_id"] != "bind-1" || resp["scope_type"] != "application" || resp["scope_id"] != "app-1" {
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
