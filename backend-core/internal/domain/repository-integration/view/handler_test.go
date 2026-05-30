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

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
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

type fakeRepoIntegrationAppStore struct {
	getIntegrationProviderByIDFunc func(ctx context.Context, id string) (domain.IntegrationProvider, error)
	listRepositoriesByProviderFunc func(ctx context.Context, providerID string) ([]domain.Repository, error)
	upsertRepositoryFunc func(ctx context.Context, r domain.Repository) error
}

func (f *fakeRepoIntegrationAppStore) GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error) {
	if f.getIntegrationProviderByIDFunc != nil {
		return f.getIntegrationProviderByIDFunc(ctx, id)
	}
	return domain.IntegrationProvider{}, nil
}
func (f *fakeRepoIntegrationAppStore) ListRepositoriesByProvider(ctx context.Context, providerID string) ([]domain.Repository, error) {
	if f.listRepositoriesByProviderFunc != nil {
		return f.listRepositoriesByProviderFunc(ctx, providerID)
	}
	return nil, nil
}
func (f *fakeRepoIntegrationAppStore) UpsertRepository(ctx context.Context, r domain.Repository) error {
	if f.upsertRepositoryFunc != nil {
		return f.upsertRepositoryFunc(ctx, r)
	}
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

func TestRepositoryIntegration_API(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ApplicationStoreOrUnavailable - nil store returns 503", func(t *testing.T) {
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		h.ListSCMRepositories(c)
		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("ListSCMRepositories - SCM Provider lookup exception cases", func(t *testing.T) {
		// 1. Not Found (404)
		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, store.ErrNotFound
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "provider_id", Value: "p-notfound"}}
		c1.Request = httptest.NewRequest("GET", "/repositories", nil)
		h.ListSCMRepositories(c1)
		if rec1.Code != 404 {
			t.Fatalf("expected 404, got %d", rec1.Code)
		}

		// 2. lookup DB Error (500)
		storeI2 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{}, errors.New("db error")
			},
		}
		h2 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI2})
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "provider_id", Value: "p-dbfail"}}
		c2.Request = httptest.NewRequest("GET", "/repositories", nil)
		h2.ListSCMRepositories(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}

		// 3. disabled provider (409)
		storeI3 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id, Enabled: false}, nil
			},
		}
		h3 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI3})
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{{Key: "provider_id", Value: "p-disabled"}}
		c3.Request = httptest.NewRequest("GET", "/repositories", nil)
		h3.ListSCMRepositories(c3)
		if rec3.Code != 409 {
			t.Fatalf("expected 409, got %d", rec3.Code)
		}

		// 4. provider type != scm (422)
		storeI4 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{ID: id, Enabled: true, ProviderType: domain.IntegrationProviderTypeCICD}, nil
			},
		}
		h4 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI4})
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "provider_id", Value: "p-cict"}}
		c4.Request = httptest.NewRequest("GET", "/repositories", nil)
		h4.ListSCMRepositories(c4)
		if rec4.Code != 422 {
			t.Fatalf("expected 422, got %d", rec4.Code)
		}

		// 5. provider capability 'pull' not enabled (422)
		storeI5 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:           id,
					Enabled:      true,
					ProviderType: domain.IntegrationProviderTypeSCM,
					Capabilities: []string{"push"}, // pull missing
				}, nil
			},
		}
		h5 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI5})
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Params = gin.Params{{Key: "provider_id", Value: "p-pushonly"}}
		c5.Request = httptest.NewRequest("GET", "/repositories", nil)
		h5.ListSCMRepositories(c5)
		if rec5.Code != 422 {
			t.Fatalf("expected 422, got %d", rec5.Code)
		}

		// 6. provider not gitea compatible (422)
		storeI6 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					CredentialsRef: "provider_sdk:github:token", // not compatible
				}, nil
			},
		}
		h6 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI6})
		rec6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(rec6)
		c6.Params = gin.Params{{Key: "provider_id", Value: "p-github"}}
		c6.Request = httptest.NewRequest("GET", "/repositories", nil)
		h6.ListSCMRepositories(c6)
		if rec6.Code != 422 {
			t.Fatalf("expected 422, got %d", rec6.Code)
		}
	})

	t.Run("ListSCMRepositories - SCM Provider client config & auth exception cases", func(t *testing.T) {
		// 1. base_url missing (422)
		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:           id,
					Enabled:      true,
					ProviderType: domain.IntegrationProviderTypeSCM,
					Capabilities: []string{"pull"},
					BaseURL:      "",
				}, nil
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c1.Request = httptest.NewRequest("GET", "/repositories", nil)
		h.ListSCMRepositories(c1)
		if rec1.Code != 422 {
			t.Fatalf("expected 422, got %d", rec1.Code)
		}

		// 2. SCM auth failed (502)
		storeI2 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        "http://invalid-url-domain-nonexist",
					CredentialsRef: "provider_sdk:gitea:invalid-auth-prefix", // invalid format triggers auth fail
					APIToken:       "xyz",
				}, nil
			},
		}
		h2 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI2})
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c2.Request = httptest.NewRequest("GET", "/repositories", nil)
		h2.ListSCMRepositories(c2)
		if rec2.Code != 502 {
			t.Fatalf("expected 502, got %d", rec2.Code)
		}
	})

	t.Run("ListSCMRepositories - SCM API unreachable or DB local list error", func(t *testing.T) {
		// Mock local httptest Server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/user/repos" {
				w.WriteHeader(http.StatusBadGateway) // unreachable or server error
				return
			}
		}))
		defer server.Close()

		// 1. remote API bad gateway (502)
		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c1.Request = httptest.NewRequest("GET", "/repositories", nil)
		h.ListSCMRepositories(c1)
		if rec1.Code != 502 {
			t.Fatalf("expected 502, got %d. Body: %s", rec1.Code, rec1.Body.String())
		}

		// Mock standard success HTTP API server for local DB error test
		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/user/repos" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[{"id":1, "name":"r1", "full_name":"org/r1", "clone_url":"http", "html_url":"http", "default_branch":"main", "private":false}]`))
				return
			}
		}))
		defer server2.Close()

		// 2. db local list error (500)
		storeI2 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server2.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
			listRepositoriesByProviderFunc: func(ctx context.Context, providerID string) ([]domain.Repository, error) {
				return nil, errors.New("db list error")
			},
		}
		h2 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI2})
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c2.Request = httptest.NewRequest("GET", "/repositories", nil)
		h2.ListSCMRepositories(c2)
		if rec2.Code != 500 {
			t.Fatalf("expected 500, got %d", rec2.Code)
		}
	})

	t.Run("ListSCMRepositories - Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/user/repos" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[
					{"id":1, "name":"r1", "full_name":"org/r1", "clone_url":"http1", "html_url":"html1", "default_branch":"main", "private":false},
					{"id":2, "name":"r2", "full_name":"org/r2", "clone_url":"http2", "html_url":"html2", "default_branch":"main", "private":true}
				]`))
				return
			}
		}))
		defer server.Close()

		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
			listRepositoriesByProviderFunc: func(ctx context.Context, providerID string) ([]domain.Repository, error) {
				// r1 is already locally imported
				return []domain.Repository{{FullName: "org/r1"}}, nil
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c.Request = httptest.NewRequest("GET", "/repositories", nil)
		h.ListSCMRepositories(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"full_name":"org/r1"`) || !strings.Contains(rec.Body.String(), `"imported":true`) {
			t.Fatalf("mismatch response body mapping: %s", rec.Body.String())
		}
	})

	t.Run("ImportSCMRepositories - validation and store errors", func(t *testing.T) {
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		h.ImportSCMRepositories(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeRepoIntegrationAppStore{}
		h = NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})

		// 2. bind error -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/import", strings.NewReader("invalid-json"))
		h.ImportSCMRepositories(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. no selection -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/import", strings.NewReader(`{"full_names": []}`))
		h.ImportSCMRepositories(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		// Mock server to test unreachable remote scm inside import
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		storeI2 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
		}
		h2 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI2})
		
		// 4. remote list repos unreachable -> 502
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c4.Request = httptest.NewRequest("POST", "/import", strings.NewReader(`{"full_names": ["org/r1"]}`))
		h2.ImportSCMRepositories(c4)
		if rec4.Code != 502 {
			t.Fatalf("expected 502, got %d", rec4.Code)
		}
	})

	t.Run("ImportSCMRepositories - upsert db error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/user/repos" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[{"id":1, "name":"r1", "full_name":"org/r1", "clone_url":"http1", "html_url":"html1", "default_branch":"main", "private":false}]`))
				return
			}
		}))
		defer server.Close()

		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
			upsertRepositoryFunc: func(ctx context.Context, r domain.Repository) error {
				return errors.New("db upsert error")
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c.Request = httptest.NewRequest("POST", "/import", strings.NewReader(`{"full_names": ["org/r1"]}`))
		h.ImportSCMRepositories(c)
		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ImportSCMRepositories - Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/user/repos" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`[
					{"id":1, "name":"r1", "full_name":"org/r1", "clone_url":"http1", "html_url":"html1", "default_branch":"main", "private":false},
					{"id":2, "name":"r2", "full_name":"org/r2", "clone_url":"http2", "html_url":"html2", "default_branch":"main", "private":true}
				]`))
				return
			}
		}))
		defer server.Close()

		var upserted []domain.Repository
		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"pull"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
			upsertRepositoryFunc: func(ctx context.Context, r domain.Repository) error {
				upserted = append(upserted, r)
				return nil
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		// org/r1 is found, org/r3 is not found
		c.Request = httptest.NewRequest("POST", "/import", strings.NewReader(`{"full_names": ["org/r1", "org/r3"]}`))
		h.ImportSCMRepositories(c)
		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if len(upserted) != 1 || upserted[0].FullName != "org/r1" {
			t.Fatalf("expected upserted org/r1, got %v", upserted)
		}
		if !strings.Contains(rec.Body.String(), `"not_found":["org/r3"]`) {
			t.Fatalf("expected not_found list containing org/r3, got %s", rec.Body.String())
		}
	})

	t.Run("CreateSCMRepository - validation & client errors", func(t *testing.T) {
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{})
		
		// 1. nil store -> 503
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		h.CreateSCMRepository(c1)
		if rec1.Code != 503 {
			t.Fatalf("expected 503, got %d", rec1.Code)
		}

		storeI := &fakeRepoIntegrationAppStore{}
		h = NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})

		// 2. bind error -> 400
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("POST", "/create", strings.NewReader("invalid-json"))
		h.CreateSCMRepository(c2)
		if rec2.Code != 400 {
			t.Fatalf("expected 400, got %d", rec2.Code)
		}

		// 3. empty repo name -> 400
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Request = httptest.NewRequest("POST", "/create", strings.NewReader(`{"name": ""}`))
		h.CreateSCMRepository(c3)
		if rec3.Code != 400 {
			t.Fatalf("expected 400, got %d", rec3.Code)
		}

		// 4. provider capability 'push' missing -> 422
		storeI4 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:           id,
					Enabled:      true,
					ProviderType: domain.IntegrationProviderTypeSCM,
					Capabilities: []string{"pull"}, // push missing
				}, nil
			},
		}
		h4 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI4})
		rec4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(rec4)
		c4.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c4.Request = httptest.NewRequest("POST", "/create", strings.NewReader(`{"name": "new-repo"}`))
		h4.CreateSCMRepository(c4)
		if rec4.Code != 422 {
			t.Fatalf("expected 422, got %d", rec4.Code)
		}

		// Mock server to test unreachable gitea on create
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		storeI5 := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"push"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
		}
		h5 := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI5})
		
		// 5. client create failed -> 502
		rec5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(rec5)
		c5.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c5.Request = httptest.NewRequest("POST", "/create", strings.NewReader(`{"name": "new-repo"}`))
		h5.CreateSCMRepository(c5)
		if rec5.Code != 502 {
			t.Fatalf("expected 502, got %d. Body: %s", rec5.Code, rec5.Body.String())
		}
	})

	t.Run("CreateSCMRepository - persist db error vs success creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id": 42, "name": "new-repo", "full_name": "owner/new-repo", "clone_url": "http1", "html_url": "html1", "default_branch": "main", "private": false}`))
		}))
		defer server.Close()

		callCount := 0
		var upserted domain.Repository
		storeI := &fakeRepoIntegrationAppStore{
			getIntegrationProviderByIDFunc: func(ctx context.Context, id string) (domain.IntegrationProvider, error) {
				return domain.IntegrationProvider{
					ID:             id,
					Enabled:        true,
					ProviderType:   domain.IntegrationProviderTypeSCM,
					Capabilities:   []string{"push"},
					BaseURL:        server.URL,
					CredentialsRef: "token:xyz",
					APIToken:       "xyz",
				}, nil
			},
			upsertRepositoryFunc: func(ctx context.Context, r domain.Repository) error {
				callCount++
				if callCount == 1 {
					return errors.New("db error")
				}
				upserted = r
				return nil
			},
		}
		h := NewRepositoryIntegrationHandler(RepositoryIntegrationConfig{ApplicationStore: storeI})

		// 1. db upsert fail -> 500
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c1.Request = httptest.NewRequest("POST", "/create", strings.NewReader(`{"name": "new-repo", "description": "desc"}`))
		h.CreateSCMRepository(c1)
		if rec1.Code != 500 {
			t.Fatalf("expected 500, got %d", rec1.Code)
		}

		// 2. success creation -> 201
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{{Key: "provider_id", Value: "p-1"}}
		c2.Request = httptest.NewRequest("POST", "/create", strings.NewReader(`{"name": "new-repo", "description": "desc"}`))
		h.CreateSCMRepository(c2)
		if rec2.Code != 201 {
			t.Fatalf("expected 201, got %d. Body: %s", rec2.Code, rec2.Body.String())
		}
		if upserted.GiteaID != 42 || upserted.FullName != "owner/new-repo" || upserted.Description != "desc" {
			t.Fatalf("mismatch upserted values: %+v", upserted)
		}
	})
}
