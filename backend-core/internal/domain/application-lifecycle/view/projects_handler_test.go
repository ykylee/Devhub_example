package view

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// 공통 fixture — hybrid ProjectModel (legacy + v2 둘 다 허용).

func newProjectHandlerForTest(t *testing.T) (*ApplicationHandler, *fakeViewApplicationStore, *fakeAppLifecycleAuditStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st := newFakeViewApplicationStore()
	audit := &fakeAppLifecycleAuditStore{}
	h := NewApplicationHandler(ApplicationConfig{
		ApplicationStore: st,
		AuditStore:       audit,
		ProjectModel:     "hybrid",
	})
	return h, st, audit
}

// --- ListProjects (legacy, repository-centric) ---

func TestListProjects_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", RepositoryID: 42, Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/repositories/42/projects", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1, got %d", len(data))
	}
}

func TestListProjects_LegacyDisabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{
		ApplicationStore: newFakeViewApplicationStore(),
		ProjectModel:     "v2",
	})
	rec := invokeJSON("GET", "/repositories/42/projects", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("GET", "/repositories/42/projects", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_BadRepoID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/repositories/notanint/projects", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_InvalidStatus400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/repositories/42/projects?status=weird", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_InvalidLimit400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/repositories/42/projects?limit=999", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_InvalidOffset400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/repositories/42/projects?offset=-1", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_LimitOffsetOK(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/repositories/42/projects?limit=10&offset=0", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjects_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errListProjects = errors.New("db")
	rec := invokeJSON("GET", "/repositories/42/projects", "/repositories/:repository_id/projects", h.ListProjects, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateProject (legacy) ---

func validCreateProjectBody() map[string]any {
	return map[string]any{
		"key":           "PRJ1",
		"name":          "Project 1",
		"description":   "desc",
		"owner_user_id": "alice",
		"visibility":    "internal",
		"status":        "planning",
		"start_date":    "2026-01-01",
		"due_date":      "2026-12-31",
	}
}

func TestCreateProject_OK(t *testing.T) {
	h, _, audit := newProjectHandlerForTest(t)
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.created" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestCreateProject_LegacyDisabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{
		ApplicationStore: newFakeViewApplicationStore(),
		ProjectModel:     "v2",
	})
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_BadRepoID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("POST", "/repositories/notanint/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_BadJSON400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.POST("/repositories/:repository_id/projects", h.CreateProject)
	req := httptest.NewRequest("POST", "/repositories/42/projects", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_MissingKey400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["key"] = ""
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_MissingName400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["name"] = ""
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_MissingOwner400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["owner_user_id"] = ""
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_InvalidVisibility400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["visibility"] = "weird"
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_InvalidStatus400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["status"] = "weird"
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_BadStartDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["start_date"] = "not-a-date"
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_BadDueDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["due_date"] = "not-a-date"
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProject_Conflict409(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", RepositoryID: 42, Status: domain.ApplicationStatusActive})
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProject_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errCreateProject = errors.New("db")
	rec := invokeJSON("POST", "/repositories/42/projects", "/repositories/:repository_id/projects", h.CreateProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateProjectStandalone (v2) ---

func TestCreateProjectStandalone_OK(t *testing.T) {
	h, _, audit := newProjectHandlerForTest(t)
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.created" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestCreateProjectStandalone_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{
		ApplicationStore: newFakeViewApplicationStore(),
		ProjectModel:     "legacy",
	})
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_BadJSON400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.POST("/projects", h.CreateProjectStandalone)
	req := httptest.NewRequest("POST", "/projects", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_MissingRequired400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["key"] = ""
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_InvalidVisibility400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["visibility"] = "weird"
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_BadStartDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["start_date"] = "not-a-date"
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_BadDueDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["due_date"] = "not-a-date"
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_RepoPayloadMissingFields400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_create_payload"] = map[string]any{"key": "REPO", "slug": "", "scm_provider": "gitea"}
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectStandalone_WithRepoPayloadOK(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_create_payload"] = map[string]any{"key": "REPO", "slug": "team/devhub-core", "scm_provider": "gitea"}
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectStandalone_RepositoryIDs(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_ids"] = []int64{100, 200}
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectStandalone_Conflict409(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errCreateProjectWithRepoPayload = nil
	st.errCreateProject = nil
	// 같은 key + repo_id 0 으로 두 번 — fake 가 conflict 반환.
	body := validCreateProjectBody()
	_ = invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, body, "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectStandalone_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errCreateProjectWithRepoPayload = errors.New("db")
	rec := invokeJSON("POST", "/projects", "/projects", h.CreateProjectStandalone, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListApplicationProjects ---

func TestListApplicationProjects_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", ApplicationID: "app-1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/applications/app-1/projects", "/applications/:application_id/projects", h.ListApplicationProjects, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListApplicationProjects_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: newFakeViewApplicationStore(), ProjectModel: "legacy"})
	rec := invokeJSON("GET", "/applications/app-1/projects", "/applications/:application_id/projects", h.ListApplicationProjects, nil, "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplicationProjects_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("GET", "/applications/app-1/projects", "/applications/:application_id/projects", h.ListApplicationProjects, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplicationProjects_MissingAppID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.GET("/applications/:application_id/projects", h.ListApplicationProjects)
	req := httptest.NewRequest("GET", "/applications/%20/projects", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListApplicationProjects_InvalidStatus400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/app-1/projects?status=weird", "/applications/:application_id/projects", h.ListApplicationProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplicationProjects_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errListProjects = errors.New("db")
	rec := invokeJSON("GET", "/applications/app-1/projects", "/applications/:application_id/projects", h.ListApplicationProjects, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListStandaloneProjects ---

func TestListStandaloneProjects_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive}) // standalone (ApplicationID="")
	rec := invokeJSON("GET", "/projects/standalone", "/projects/standalone", h.ListStandaloneProjects, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListStandaloneProjects_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	rec := invokeJSON("GET", "/projects/standalone", "/projects/standalone", h.ListStandaloneProjects, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListStandaloneProjects_InvalidStatus400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/projects/standalone?status=weird", "/projects/standalone", h.ListStandaloneProjects, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListStandaloneProjects_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errListProjects = errors.New("db")
	rec := invokeJSON("GET", "/projects/standalone", "/projects/standalone", h.ListStandaloneProjects, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateApplicationProject ---

func TestCreateApplicationProject_OK(t *testing.T) {
	h, _, audit := newProjectHandlerForTest(t)
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.created" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestCreateApplicationProject_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: newFakeViewApplicationStore(), ProjectModel: "legacy"})
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_MissingAppID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.POST("/applications/:application_id/projects", h.CreateApplicationProject)
	req := httptest.NewRequest("POST", "/applications/%20/projects", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationProject_BadJSON400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.POST("/applications/:application_id/projects", h.CreateApplicationProject)
	req := httptest.NewRequest("POST", "/applications/app-1/projects", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_MissingFields400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["key"] = ""
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_InvalidVisibility400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["visibility"] = "weird"
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_BadStartDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["start_date"] = "not-a-date"
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_BadDueDate400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["due_date"] = "not-a-date"
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_RepoPayloadMissingFields400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_create_payload"] = map[string]any{"key": "", "slug": "team/x", "scm_provider": "gitea"}
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationProject_WithRepoPayloadOK(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_create_payload"] = map[string]any{"key": "REPO", "slug": "team/devhub-core", "scm_provider": "gitea"}
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationProject_RepositoryIDs(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	body["repository_ids"] = []int64{100, 200}
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationProject_Conflict409(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := validCreateProjectBody()
	_ = invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, body, "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationProject_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errCreateProjectWithRepoPayload = errors.New("db")
	rec := invokeJSON("POST", "/applications/app-1/projects", "/applications/:application_id/projects", h.CreateApplicationProject, validCreateProjectBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListProjectRepositories ---

func TestListProjectRepositories_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	st.seedProjectRepo(domain.ProjectRepository{ProjectID: "p-1", RepositoryID: 42, Role: "primary"})
	rec := invokeJSON("GET", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.ListProjectRepositories, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListProjectRepositories_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: newFakeViewApplicationStore(), ProjectModel: "legacy"})
	rec := invokeJSON("GET", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.ListProjectRepositories, nil, "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjectRepositories_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("GET", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.ListProjectRepositories, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListProjectRepositories_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errListProjectRepositories = errors.New("db")
	rec := invokeJSON("GET", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.ListProjectRepositories, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateProjectRepository ---

func TestCreateProjectRepository_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	body := map[string]any{"repository_id": 42, "role": "linked"}
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectRepository_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: newFakeViewApplicationStore(), ProjectModel: "legacy"})
	body := map[string]any{"repository_id": 42, "role": "linked"}
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectRepository_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	body := map[string]any{"repository_id": 42, "role": "linked"}
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectRepository_BadJSON400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	r := gin.New()
	r.POST("/projects/:project_id/repositories", h.CreateProjectRepository)
	req := httptest.NewRequest("POST", "/projects/p-1/repositories", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectRepository_MissingRepoID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := map[string]any{"role": "linked"} // repository_id 누락 → 0
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateProjectRepository_DefaultRole(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	body := map[string]any{"repository_id": 42} // role 없음 → "linked"
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectRepository_Conflict409(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	st.seedProjectRepo(domain.ProjectRepository{ProjectID: "p-1", RepositoryID: 42, Role: "primary"})
	body := map[string]any{"repository_id": 42, "role": "linked"}
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateProjectRepository_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errCreateProjectRepository = errors.New("db")
	body := map[string]any{"repository_id": 42}
	rec := invokeJSON("POST", "/projects/p-1/repositories", "/projects/:project_id/repositories", h.CreateProjectRepository, body, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- DeleteProjectRepository ---

func TestDeleteProjectRepository_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	st.seedProjectRepo(domain.ProjectRepository{ProjectID: "p-1", RepositoryID: 42, Role: "primary"})
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/42", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteProjectRepository_V2Disabled410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: newFakeViewApplicationStore(), ProjectModel: "legacy"})
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/42", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteProjectRepository_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{ProjectModel: "hybrid"})
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/42", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteProjectRepository_BadRepoID400(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/notanint", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteProjectRepository_NotFound404(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/42", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteProjectRepository_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errDeleteProjectRepository = errors.New("db")
	rec := invokeJSON("DELETE", "/projects/p-1/repositories/42", "/projects/:project_id/repositories/:repository_id", h.DeleteProjectRepository, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- GetProject ---

func TestGetProject_OK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/projects/p-1", "/projects/:project_id", h.GetProject, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetProject_NotFound404(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("GET", "/projects/missing", "/projects/:project_id", h.GetProject, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetProject_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errGetProject = errors.New("db")
	rec := invokeJSON("GET", "/projects/p-1", "/projects/:project_id", h.GetProject, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetProject_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	rec := invokeJSON("GET", "/projects/p-1", "/projects/:project_id", h.GetProject, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- UpdateProject ---

func TestUpdateProject_OK(t *testing.T) {
	h, st, audit := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "Renamed"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.updated" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestUpdateProject_BadJSON400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	r := gin.New()
	r.PATCH("/projects/:project_id", h.UpdateProject)
	req := httptest.NewRequest("PATCH", "/projects/p-1", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_NotFound404(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/projects/missing", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_OwnerMismatch403(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_KeyImmutable422(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"key": "DIFFKEY"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "project_key_immutable") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestUpdateProject_KeySameOK(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"key": "PRJ1"} // same → ok
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_EmptyName400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "  "}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_ApplicationIDInvalidUUID422(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"application_id": "not-a-uuid"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "application_id_invalid") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestUpdateProject_ApplicationIDNotFound422(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"application_id": "11111111-2222-3333-4444-555555555555"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_ApplicationIDValidExisting(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	appUUID := "11111111-2222-3333-4444-555555555555"
	st.seedApp(domain.Application{ID: appUUID, Key: "APP1", Status: domain.ApplicationStatusActive})
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"application_id": appUUID}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_ApplicationIDDetachEmpty(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice", ApplicationID: "old"})
	body := map[string]any{"application_id": ""}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_ApplicationIDLookupError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	st.errGetApplication = errors.New("db")
	body := map[string]any{"application_id": "11111111-2222-3333-4444-555555555555"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_InvalidVisibility400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"visibility": "weird"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_BadStartDate400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"start_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_BadDueDate400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"due_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_InvalidStatus400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"status": "weird"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_StatusTransition(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusPlanning, OwnerUserID: "alice"})
	body := map[string]any{"status": "active", "resume_reason": "back"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_AllPatchFields(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{
		"name":            "Renamed",
		"description":     "newdesc",
		"owner_user_id":   "carol",
		"visibility":      "public",
		"start_date":      "2026-02-01",
		"due_date":        "2026-11-30",
		"hold_reason":     "hold",
		"archived_reason": "ignored",
	}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProject_LookupError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errGetProject = errors.New("db")
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateProject_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	st.errUpdateProject = errors.New("db")
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/projects/p-1", "/projects/:project_id", h.UpdateProject, body, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ArchiveProject ---

func TestArchiveProject_OK(t *testing.T) {
	h, st, audit := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"archived_reason": "done"}
	rec := invokeJSON("DELETE", "/projects/p-1", "/projects/:project_id", h.ArchiveProject, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.archived" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchiveProject_NotFound404(t *testing.T) {
	h, _, _ := newProjectHandlerForTest(t)
	rec := invokeJSON("DELETE", "/projects/missing", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_OwnerMismatch403(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/projects/p-1", "/projects/:project_id", h.ArchiveProject, nil, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_HardDeleteOK(t *testing.T) {
	h, st, audit := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusArchived, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/projects/p-1?hard=true", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "deleted") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "project.deleted" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchiveProject_HardDeleteNotArchived400(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/projects/p-1?hard=true", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_HardDeleteStoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusArchived, OwnerUserID: "alice"})
	st.errDeleteProject = errors.New("db")
	rec := invokeJSON("DELETE", "/projects/p-1?hard=true", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_StoreError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.seedProject(domain.Project{ID: "p-1", Key: "PRJ1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	st.errArchiveProject = errors.New("db")
	rec := invokeJSON("DELETE", "/projects/p-1", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_LookupError500(t *testing.T) {
	h, st, _ := newProjectHandlerForTest(t)
	st.errGetProject = errors.New("db")
	rec := invokeJSON("DELETE", "/projects/p-1", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveProject_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	rec := invokeJSON("DELETE", "/projects/p-1", "/projects/:project_id", h.ArchiveProject, nil, "alice", "developer")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}
