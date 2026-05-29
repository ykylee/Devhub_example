package view

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)

type fakeRepoIntegrationAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeRepoIntegrationAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_repo_id"
	f.created = append(f.created, log)
	return log, nil
}

type fakeRepoIntegrationAppStore struct{}

func (f *fakeRepoIntegrationAppStore) GetIntegrationProviderByID(_ context.Context, _ string) (domain.IntegrationProvider, error) {
	return domain.IntegrationProvider{}, nil
}
func (f *fakeRepoIntegrationAppStore) ListRepositoriesByProvider(_ context.Context, _ string) ([]domain.Repository, error) {
	return nil, nil
}
func (f *fakeRepoIntegrationAppStore) UpsertRepository(_ context.Context, _ domain.Repository) error {
	return nil
}

func TestNewRepositoryIntegrationHandler_NonNil(t *testing.T) {
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewRepositoryIntegrationHandler_ConfigPropagation(t *testing.T) {
	cfg := RepositoryIntegrationConfig{
		ApplicationStore: &fakeRepoIntegrationAppStore{},
		AuditStore:       &fakeRepoIntegrationAuditStore{},
	}
	h := NewRepositoryIntegrationHandler(cfg)
	if h.cfg.ApplicationStore == nil {
		t.Fatal("ApplicationStore not propagated")
	}
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRepoIntegrationAuditStore{}
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "repo.test", "repository", "repo-1", nil)
	if got.AuditID != "audit_repo_id" {
		t.Fatalf("audit stamp: %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d", len(store.created))
	}
	c0 := store.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "repo.test" {
		t.Fatalf("mapping = %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRepoIntegrationAuditStore{}
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{AuditStore: store})
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

	store := &fakeRepoIntegrationAuditStore{err: errors.New("db_down")}
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{AuditStore: store})
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
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	got, ok := h.ApplicationStoreOrUnavailable(c)
	if ok {
		t.Fatal("expected ok=false when store nil")
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
	store := &fakeRepoIntegrationAppStore{}
	h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: store})
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
	if rec.Code != 0 && rec.Code != 200 {
		t.Fatalf("must not write status, got %d", rec.Code)
	}
}

func TestNormalizeProviderSDKKey(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"gitea", "gitea"},
		{"GITEA", "gitea"},
		{"  gitea  ", "gitea"},
		{"my-gitea-internal", "gitea"},
		{"github", "github"},
		{"GitHub", "github"},
		{"github-enterprise", "github"},
		{"gitlab", "gitlab"},
		{"  GitLab  ", "gitlab"},
		{"", ""},
	}
	for _, c := range cases {
		got := normalizeProviderSDKKey(c.input)
		if got != c.want {
			t.Errorf("normalizeProviderSDKKey(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
