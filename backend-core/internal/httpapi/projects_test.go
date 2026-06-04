package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// Project CRUD handler tests (API-55..56, sprint claude/work_260514-c).

// 1) POST /repositories/:repository_id/projects — happy.
func TestCreateProject_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects",
		`{"key":"sprint-q3","name":"Q3 Sprint","owner_user_id":"u1","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key":"sprint-q3"`)) {
		t.Errorf("response should echo key: %s", rec.Body.String())
	}
}

func TestCreateApplicationProject_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms/app-1/projects",
		`{"repository_ids":[42],"key":"sprint-q4","name":"Q4 Sprint","owner_user_id":"u1","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"platform_id":"app-1"`)) {
		t.Errorf("response should include platform_id: %s", rec.Body.String())
	}
}

func TestCreateProjectStandalone_WithRepositoryCreatePayload(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/projects",
		`{"key":"standalone-a","name":"Standalone A","owner_user_id":"u1","visibility":"internal","status":"planning","repository_create_payload":{"key":"DEVHUB","slug":"team/devhub","scm_provider":"gitea"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"repository_id"`)) {
		t.Errorf("response should include repository_id: %s", rec.Body.String())
	}
}

// 2) POST /repositories/:repository_id/projects — invalid status → 400.
func TestCreateProject_InvalidStatus(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/42/projects",
		`{"key":"x","name":"X","owner_user_id":"u1","visibility":"internal","status":"unknown"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 3) POST /repositories/:repository_id/projects — duplicate key → 409.
func TestCreateProject_DuplicateKey(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
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
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "sprint-q3", Name: "X", RepositoryID: 42, Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/projects/"+p.ID,
		`{"key":"new-key"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"project_key_immutable"`)) {
		t.Errorf("expected project_key_immutable: %s", rec.Body.String())
	}
}

// status 전이 정책 자유화 (2026-05-28) — closed→planning, archived→planning 같은
// backward 전이도 운영자 임의로 가능. 이전 테스트는 거부 검증이었으므로 자유화 후
// expected behavior 로 갱신.
func TestUpdateProject_AnyStatusTransitionAllowed(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "X", RepositoryID: 42, Status: domain.PlatformStatusClosed,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/projects/"+p.ID,
		`{"status":"planning"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
}

// 6) DELETE /projects/:id — archive happy.
func TestArchiveProject_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "X", RepositoryID: 42, Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/projects/"+p.ID,
		`{"archived_reason":"sprint ended"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"archived"`) {
		t.Errorf("expected status=archived: %s", rec.Body.String())
	}
}

func TestProjectDeleteLifecycle(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "X", RepositoryID: 42, Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	// 1) Active 상태에서 ?hard=true 로 삭제 시도 -> 400 Bad Request
	rec1 := doJSON(t, router, http.MethodDelete, "/api/v1/projects/"+p.ID+"?hard=true", "")
	if rec1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for hard deletion of non-archived project, got=%d body=%s", rec1.Code, rec1.Body.String())
	}

	// 2) 일반 soft-delete (archive) -> 200 OK & status=archived
	rec2 := doJSON(t, router, http.MethodDelete, "/api/v1/projects/"+p.ID, `{"archived_reason":"sprint ended"}`)
	if rec2.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec2.Code, rec2.Body.String())
	}

	// 3) Archived 상태에서 ?hard=true 로 영구 삭제 -> 200 OK & status=deleted
	rec3 := doJSON(t, router, http.MethodDelete, "/api/v1/projects/"+p.ID+"?hard=true", "")
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 for hard deletion of archived project, got=%d body=%s", rec3.Code, rec3.Body.String())
	}
	if !strings.Contains(rec3.Body.String(), `"status":"deleted"`) {
		t.Errorf("expected status=deleted: %s", rec3.Body.String())
	}

	// 4) 삭제 후 조회 시 -> 404 NotFound
	rec4 := doJSON(t, router, http.MethodGet, "/api/v1/projects/"+p.ID, "")
	if rec4.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deletion, got=%d", rec4.Code)
	}
}

// 7) GET /projects/:id — not_found → 404.
func TestGetProject_NotFound(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 8) GET /repositories/:repository_id/projects — list with filter.
func TestListProjects_Filter(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	for _, status := range []domain.PlatformStatus{
		domain.PlatformStatusPlanning,
		domain.PlatformStatusActive,
		domain.PlatformStatusArchived,
	} {
		_, _ = platformStore.CreateProject(context.Background(), domain.Project{
			Key: "k-" + string(status[:4]), Name: "N", RepositoryID: 42, Status: status,
			Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
		})
	}
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/repositories/42/projects", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":2`) {
		t.Errorf("default list should exclude archived (total=2): %s", rec.Body.String())
	}
}

func TestProjectDashboard_Developer_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "Proj X", RepositoryID: 42, Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
		ProjectMembers: []domain.ProjectMember{
			{UserID: "u1", ProjectRole: "developer"},
		},
	})

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "u1",
		Subject: "user-u1",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		PlatformStore:       platformStore,
		BearerTokenVerifier: verifier,
		AuthDevFallback:     false,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+p.ID+"/dashboard?persona=developer", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"current_persona":"developer"`) {
		t.Errorf("expected developer_view: %s", rec.Body.String())
	}
}

func TestProjectDashboard_ProjectLeader_Forbidden(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "Proj X", RepositoryID: 42, Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
		ProjectMembers: []domain.ProjectMember{
			{UserID: "u2", ProjectRole: "developer"},
		},
	})

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "u2",
		Subject: "user-u2",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		PlatformStore:       platformStore,
		BearerTokenVerifier: verifier,
		AuthDevFallback:     false,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+p.ID+"/dashboard?persona=project_leader", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_row_denied"`) {
		t.Errorf("expected auth_row_denied: %s", rec.Body.String())
	}
}

func TestProjectDashboard_InvalidPersona(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	p, _ := platformStore.CreateProject(context.Background(), domain.Project{
		Key: "k1", Name: "Proj X", RepositoryID: 42, Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
		ProjectMembers: []domain.ProjectMember{
			{UserID: "u1", ProjectRole: "developer"},
		},
	})

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "u1",
		Subject: "user-u1",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		PlatformStore:       platformStore,
		BearerTokenVerifier: verifier,
		AuthDevFallback:     false,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+p.ID+"/dashboard?persona=invalid", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_persona"`) {
		t.Errorf("expected invalid_persona code: %s", rec.Body.String())
	}
}

