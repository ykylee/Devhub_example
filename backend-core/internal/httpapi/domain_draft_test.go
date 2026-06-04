package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateRepositoryDraft_Success(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories",
		`{"key":"my-repo","slug":"team/my-repo"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"status":"draft"`) {
		t.Errorf("expected status=draft: %s", body)
	}
	if !strings.Contains(body, `"full_name":"team/my-repo"`) {
		t.Errorf("expected full_name=team/my-repo: %s", body)
	}
	if !strings.Contains(body, `"name":"my-repo"`) {
		t.Errorf("expected name=my-repo: %s", body)
	}
}

func TestCreateRepositoryDraft_ValidationError(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories",
		`{"key":"","slug":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s (expected 400)", rec.Code, rec.Body.String())
	}
}

func TestUpdateRepositoryDraft_Success(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	draft, _ := platformStore.CreateRepositoryDraft(context.Background(), "repo-a", "team/repo-a", "")
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/repositories/"+strconv.FormatInt(draft.ID, 10),
		`{"key":"repo-a-v2","slug":"team/repo-a-v2"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"name":"repo-a-v2"`) {
		t.Errorf("expected updated name: %s", body)
	}
	if !strings.Contains(body, `"full_name":"team/repo-a-v2"`) {
		t.Errorf("expected updated full_name: %s", body)
	}
}

func TestUpdateRepositoryDraft_NotFound(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/repositories/99999",
		`{"key":"x"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s (expected 404)", rec.Code, rec.Body.String())
	}
}

func TestUpdateRepositoryDraft_NotDraft(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	draft, _ := platformStore.CreateRepositoryDraft(context.Background(), "repo-n", "team/repo-n", "")
	platformStore.mu.Lock()
	repo := platformStore.draftRepos[draft.ID]
	repo.Status = "active"
	platformStore.draftRepos[draft.ID] = repo
	platformStore.mu.Unlock()

	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/repositories/"+strconv.FormatInt(draft.ID, 10),
		`{"key":"should-fail"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s (expected 404, update on non-draft repo)", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepository_Success(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	draft, _ := platformStore.CreateRepositoryDraft(context.Background(), "repo-d", "team/repo-d", "")
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/repositories/"+strconv.FormatInt(draft.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	_, err := platformStore.GetRepositoryByID(context.Background(), draft.ID)
	if err == nil {
		t.Errorf("expected ErrNotFound after delete")
	}
}

func TestDeleteRepository_NotFound(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/repositories/99999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s (expected 404)", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepository_FKConflict(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	draft, _ := platformStore.CreateRepositoryDraft(context.Background(), "repo-fk", "team/repo-fk", "")
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "LINKAPP99", Name: "LinkApp", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo-fk",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/repositories/"+strconv.FormatInt(draft.ID, 10), "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s (expected 409 FK conflict)", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"code":"repository_has_links"`) {
		t.Errorf("expected repository_has_links code: %s", body)
	}
}

func TestRequestRepositoryPublish_Success(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	createRec := doJSON(t, router, http.MethodPost, "/api/v1/repositories",
		`{"key":"repo-pub","slug":"team/repo-pub","provider_key":"gitea"}`)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("seed create failed: status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	repoID := extractJSONInt(createRec.Body.String(), `"id":`)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/"+strconv.FormatInt(repoID, 10)+"/publish", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("expected publish success: %s", body)
	}
}

func TestRequestRepositoryPublish_NotFound(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/99999/publish", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s (expected 404)", rec.Code, rec.Body.String())
	}
}

func TestRequestRepositoryPublish_AlreadyPublished(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	draft, _ := platformStore.CreateRepositoryDraft(context.Background(), "repo-done", "team/repo-done", "")
	platformStore.mu.Lock()
	repo := platformStore.draftRepos[draft.ID]
	repo.Status = "active"
	platformStore.draftRepos[draft.ID] = repo
	platformStore.mu.Unlock()

	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodPost, "/api/v1/repositories/"+strconv.FormatInt(draft.ID, 10)+"/publish", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s (expected 409 for already-published)", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "only draft repository can be published") {
		t.Errorf("expected 'only draft repository can be published' error: %s", body)
	}
}

// extractJSONInt extracts an integer value following a JSON key prefix from a body string.
func extractJSONInt(body, key string) int64 {
	idx := strings.Index(body, key)
	if idx < 0 {
		return 0
	}
	start := idx + len(key)
	end := start
	for end < len(body) && body[end] >= '0' && body[end] <= '9' {
		end++
	}
	v, _ := strconv.ParseInt(body[start:end], 10, 64)
	return v
}
