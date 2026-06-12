package view

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/application-lifecycle/routing"
	"github.com/gin-gonic/gin"
)

// fakeAutoRouter implements routing.AutoRouter for integration tests.
type fakeAutoRouter struct {
	routeFunc func(ctx context.Context, voc routing.VocRegistration) (routing.AutoRouteDecision, error)
}

func (f *fakeAutoRouter) Route(ctx context.Context, voc routing.VocRegistration) (routing.AutoRouteDecision, error) {
	if f.routeFunc != nil {
		return f.routeFunc(ctx, voc)
	}
	return routing.AutoRouteDecision{Matched: false, Reason: "no_match"}, nil
}

// fakeITVocStore extends fakeVocStore with RouteVoc support for integration tests.
type fakeITVocStore struct {
	fakeVocStore
	routeVocFunc func(ctx context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error)
}

func (f *fakeITVocStore) RouteVoc(ctx context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error) {
	if f.routeVocFunc != nil {
		return f.routeVocFunc(ctx, vocID, projectID, dr)
	}
	dr.ID = "dr-auto-" + vocID
	now := time.Now().UTC()
	return domain.DevRequestVoc{
		ID:           vocID,
		Status:       domain.DevRequestVocStatusRouted,
		ProjectID:    projectID,
		DevRequestID: dr.ID,
		RoutedAt:     &now,
	}, dr, nil
}

func newVocTestRouter(vs VocStore, ar routing.AutoRouter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewVocHandler(VocHandlerConfig{VocStore: vs, AutoRouter: ar})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("devhub_actor_login", "admin01")
		c.Next()
	})
	rg := r.Group("/api/v1")
	RegisterVocRoutes(rg, h)
	return r
}

func TestIntegration_VocAutoRouting_GiteaOK(t *testing.T) {
	ar := &fakeAutoRouter{
		routeFunc: func(_ context.Context, voc routing.VocRegistration) (routing.AutoRouteDecision, error) {
			return routing.AutoRouteDecision{
				Matched:    true,
				PlatformID: "plat-gitea-1",
				Reason:     "external_ref_pattern",
			}, nil
		},
	}
	vs := &fakeITVocStore{
		routeVocFunc: func(_ context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error) {
			now := time.Now().UTC()
			return domain.DevRequestVoc{
				ID:           vocID,
				Status:       domain.DevRequestVocStatusRouted,
				ExternalRef:  dr.ExternalRef,
				SourceSystem: dr.SourceSystem,
				ProjectID:    projectID,
				DevRequestID: "dr-auto-" + vocID,
				RoutedAt:     &now,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, domain.DevRequest{
				ID: "dr-auto-" + vocID,
			}, nil
		},
	}
	r := newVocTestRouter(vs, ar)

	body := `{"title":"Test Gitea Request","details":"Auto-routing test","source_system":"gitea"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dev-requests/GITEA-123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		AutoRouted  bool   `json:"auto_routed"`
		DevRequestID string `json:"dev_request_id"`
		PlatformID  string `json:"platform_id"`
		Reason      string `json:"reason"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != string(domain.DevRequestVocStatusRouted) {
		t.Fatalf("expected status routed, got %s", resp.Status)
	}
	if !resp.AutoRouted {
		t.Fatal("expected auto_routed=true")
	}
	if resp.DevRequestID == "" {
		t.Fatal("expected non-empty dev_request_id")
	}
	if resp.PlatformID != "plat-gitea-1" {
		t.Fatalf("expected platform_id plat-gitea-1, got %s", resp.PlatformID)
	}
	if resp.Reason != "external_ref_pattern" {
		t.Fatalf("expected reason external_ref_pattern, got %s", resp.Reason)
	}
}

func TestIntegration_VocAutoRouting_NoMatch(t *testing.T) {
	ar := &fakeAutoRouter{
		routeFunc: func(_ context.Context, voc routing.VocRegistration) (routing.AutoRouteDecision, error) {
			return routing.AutoRouteDecision{Matched: false, Reason: "no_match"}, nil
		},
	}
	vs := &fakeITVocStore{}
	r := newVocTestRouter(vs, ar)

	body := `{"title":"No Match Request","details":"Should not be auto-routed","source_system":"manual"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dev-requests/NOMATCH-001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		AutoRouted  bool   `json:"auto_routed,omitempty"`
		DevRequestID string `json:"dev_request_id,omitempty"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != string(domain.DevRequestVocStatusReceived) {
		t.Fatalf("expected status received, got %s", resp.Status)
	}
	if resp.AutoRouted {
		t.Fatal("expected auto_routed=false")
	}
	if resp.DevRequestID != "" {
		t.Fatalf("expected empty dev_request_id, got %s", resp.DevRequestID)
	}
}

func TestIntegration_VocAutoRouting_RouteErrorDegradation(t *testing.T) {
	// RouteVoc failure should not break the voc creation — graceful degradation.
	ar := &fakeAutoRouter{
		routeFunc: func(_ context.Context, voc routing.VocRegistration) (routing.AutoRouteDecision, error) {
			return routing.AutoRouteDecision{
				Matched:    true,
				PlatformID: "plat-gitea-1",
				Reason:     "external_ref_pattern",
			}, nil
		},
	}
	vs := &fakeITVocStore{
		routeVocFunc: func(_ context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error) {
			return domain.DevRequestVoc{}, domain.DevRequest{}, assertAnError
		},
	}
	r := newVocTestRouter(vs, ar)

	body := `{"title":"Route Error Test","details":"RouteVoc fails","source_system":"gitea"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dev-requests/GITEA-ERR", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 (graceful degradation), got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		AutoRouted bool   `json:"auto_routed,omitempty"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Status != string(domain.DevRequestVocStatusReceived) {
		t.Fatalf("expected status received on route failure, got %s", resp.Status)
	}
	if resp.AutoRouted {
		t.Fatal("expected auto_routed=false on route failure")
	}
}

var assertAnError = &assertError{}

type assertError struct{}

func (e *assertError) Error() string { return "assertion: route error" }
