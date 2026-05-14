package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// Project CRUD handler tests (API-55..56, sprint claude/work_260514-c).

// 1) POST /repositories/:repository_id/projects — happy.
func TestCreateProject_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects",
		`{"key":"sprint-q3","name":"Q3 Sprint","owner_user_id":"u1","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key":"sprint-q3"`)) {
		t.Errorf("response should echo key: %s", rec.Body.String())
	}
}

// 2) POST /repositories/:repository_id/projects — invalid status → 400.
func TestCreateProject_InvalidStatus(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects",
		`{"key":"x","name":"X","owner_user_id":"u1","visibility":"internal","status":"unknown"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 3) POST /repositories/:repository_id/projects — duplicate key → 409.
func TestCreateProject_DuplicateKey(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	body := `{"key":"sprint-q3","name":"Q3 Sprint","owner_user_id":"u1","visibility":"internal","status":"planning"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"project_key_conflict"`)) {
		t.Errorf("expected project_key_conflict: %s", rec.Body.String())
	}
}

// 4) PATCH /projects/:id — immutable key 거부.
func TestUpdateProject_ImmutableKey(t *testing.T) {
	appStore := newMemoryApplicationStore()
	p, _ := appStore.CreateProject(context.Background(), domain.Project{
		Key: "sprint-q3", Name: "X", RepositoryID: 42, Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/projects/"+p.ID,
		`{"key":"new-key"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"project_key_immutable"`)) {
		t.Errorf("expected project_key_immutable: %s", rec.Body.String())
	}
}

// 5) PATCH /projects/:id — invalid status transition (archived → planning) → 422.
func TestUpdateProject_InvalidStatusTransition(t *testing.T) {
	appStore := newMemoryApplicationStore()
	p, _ := appStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "X", RepositoryID: 42, Status: domain.ApplicationStatusClosed,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/projects/"+p.ID,
		`{"status":"planning"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 6) DELETE /projects/:id — archive happy.
func TestArchiveProject_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	p, _ := appStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "X", RepositoryID: 42, Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/projects/"+p.ID,
		`{"archived_reason":"sprint ended"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"archived"`) {
		t.Errorf("expected status=archived: %s", rec.Body.String())
	}
}

// 7) GET /projects/:id — not_found → 404.
func TestGetProject_NotFound(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 8) GET /repositories/:repository_id/projects — list with filter.
func TestListProjects_Filter(t *testing.T) {
	appStore := newMemoryApplicationStore()
	for _, status := range []domain.ApplicationStatus{
		domain.ApplicationStatusPlanning,
		domain.ApplicationStatusActive,
		domain.ApplicationStatusArchived,
	} {
		_, _ = appStore.CreateProject(context.Background(), domain.Project{
			Key: "k-" + string(status[:4]), Name: "N", RepositoryID: 42, Status: status,
			Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
		})
	}
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/repositories/42/projects", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":2`) {
		t.Errorf("default list should exclude archived (total=2): %s", rec.Body.String())
	}
}
