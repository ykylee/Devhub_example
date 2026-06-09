package view

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type fakeVocStore struct {
	listVocsFunc        func(ctx context.Context, status, assigneeUserID string, limit, offset int) ([]domain.DevRequestVoc, error)
	getVocByExternalRef func(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequestVoc, bool, error)
}

func (f *fakeVocStore) CreateVoc(_ context.Context, v domain.DevRequestVoc) (domain.DevRequestVoc, error) {
	v.ID = "voc_test"
	return v, nil
}
func (f *fakeVocStore) GetVocByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequestVoc, bool, error) {
	if f.getVocByExternalRef != nil {
		return f.getVocByExternalRef(ctx, sourceSystem, externalRef)
	}
	return domain.DevRequestVoc{}, false, nil
}
func (f *fakeVocStore) GetVocByID(_ context.Context, id string) (domain.DevRequestVoc, error) {
	return domain.DevRequestVoc{ID: id}, nil
}
func (f *fakeVocStore) RouteVoc(_ context.Context, _ string, _ string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error) {
	dr.ID = "dr_test"
	return domain.DevRequestVoc{ID: "voc_test"}, dr, nil
}
func (f *fakeVocStore) ListVocs(ctx context.Context, status, assigneeUserID string, limit, offset int) ([]domain.DevRequestVoc, error) {
	if f.listVocsFunc != nil {
		return f.listVocsFunc(ctx, status, assigneeUserID, limit, offset)
	}
	return nil, nil
}

func newListVocsRouter(vs VocStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewVocHandler(VocHandlerConfig{VocStore: vs})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("devhub_actor_login", "admin01")
		c.Next()
	})
	rg := r.Group("/api/v1")
	RegisterVocRoutes(rg, h)
	return r
}

func TestListVocsFilters(t *testing.T) {
	called := struct {
		status  string
		assign  string
		limit   int
		offset  int
	}{}
	vs := &fakeVocStore{listVocsFunc: func(_ context.Context, status, assign string, limit, offset int) ([]domain.DevRequestVoc, error) {
		called.status = status
		called.assign = assign
		called.limit = limit
		called.offset = offset
		now := time.Now().UTC()
		return []domain.DevRequestVoc{
			{ID: "voc-1", Status: domain.DevRequestVocStatusReceived, AssigneeUserID: "alice", CreatedAt: now},
			{ID: "voc-2", Status: domain.DevRequestVocStatusReceived, AssigneeUserID: "alice", CreatedAt: now},
		}, nil
	}}
	r := newListVocsRouter(vs)

	t.Run("status_received_filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vocs?status=received&assignee=alice&limit=10&offset=20", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
		if called.status != "received" || called.assign != "alice" || called.limit != 10 || called.offset != 20 {
			t.Fatalf("filter not propagated: %+v", called)
		}
		var body struct {
			Count int `json:"count"`
			Data  []map[string]any
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Count != 2 || len(body.Data) != 2 {
			t.Fatalf("expected 2 items, got count=%d", body.Count)
		}
		if body.Data[0]["id"] != "voc-1" {
			t.Fatalf("expected voc-1, got %v", body.Data[0]["id"])
		}
	})

	t.Run("invalid_status_returns_400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vocs?status=pending", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid_limit_returns_400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/vocs?limit=abc", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestListVocsStoreUnavailable(t *testing.T) {
	r := newListVocsRouter(nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vocs", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", w.Code, w.Body.String())
	}
}
