package view

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// newAppHandlerForTest — 공통 fixture: handler + store + audit fake 생성.
// gin TestMode + 기본 SCM provider (gitea/github) 시드.
func newAppHandlerForTest(t *testing.T) (*ApplicationHandler, *fakeViewApplicationStore, *fakeAppLifecycleAuditStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st := newFakeViewApplicationStore()
	st.seedProvider(domain.SCMProvider{ProviderKey: "gitea", DisplayName: "Gitea", Enabled: true, AdapterVersion: "v1"})
	st.seedProvider(domain.SCMProvider{ProviderKey: "github", DisplayName: "GitHub", Enabled: true, AdapterVersion: "v1"})
	st.seedProvider(domain.SCMProvider{ProviderKey: "forgejo", DisplayName: "Forgejo", Enabled: false, AdapterVersion: "v1"})
	audit := &fakeAppLifecycleAuditStore{}
	h := NewApplicationHandler(ApplicationConfig{
		ApplicationStore: st,
		AuditStore:       audit,
		ProjectModel:     "hybrid",
	})
	return h, st, audit
}

// invokeJSON — gin engine 생성 + route 등록 + 요청 + recorder 반환.
// admin actor 로 컨텍스트를 세팅해 enforceRowOwnership 통과시킬 수 있도록 middleware 주입.
func invokeJSON(method, path, fullPath string, handler gin.HandlerFunc, body any, actorLogin, actorRole string) *httptest.ResponseRecorder {
	r := gin.New()
	if actorLogin != "" || actorRole != "" {
		r.Use(func(c *gin.Context) {
			if actorLogin != "" {
				c.Set("devhub_actor_login", actorLogin)
			}
			if actorRole != "" {
				c.Set("devhub_actor_role", actorRole)
			}
			c.Next()
		})
	}
	r.Handle(method, fullPath, handler)
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- ListSCMProviders ---

func TestListSCMProviders_OK(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/scm-providers", "/scm-providers", h.ListSCMProviders, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].([]any)
	if len(data) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(data))
	}
}

func TestListSCMProviders_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	rec := invokeJSON("GET", "/scm-providers", "/scm-providers", h.ListSCMProviders, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListSCMProviders_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errListSCMProviders = errors.New("db_down")
	rec := invokeJSON("GET", "/scm-providers", "/scm-providers", h.ListSCMProviders, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- UpdateSCMProvider ---

func TestUpdateSCMProvider_OK(t *testing.T) {
	h, _, audit := newAppHandlerForTest(t)
	body := map[string]any{"enabled": false, "display_name": "Gitea (paused)"}
	rec := invokeJSON("PATCH", "/scm-providers/gitea", "/scm-providers/:provider_key", h.UpdateSCMProvider, body, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "scm_provider.updated" {
		t.Fatalf("audit not recorded: %+v", audit.created)
	}
}

func TestUpdateSCMProvider_AdapterVersionRejected422(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := map[string]any{"adapter_version": "v999"}
	rec := invokeJSON("PATCH", "/scm-providers/gitea", "/scm-providers/:provider_key", h.UpdateSCMProvider, body, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "adapter_version_immutable") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestUpdateSCMProvider_BadJSON400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	r := gin.New()
	r.PATCH("/scm-providers/:provider_key", h.UpdateSCMProvider)
	req := httptest.NewRequest("PATCH", "/scm-providers/gitea", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateSCMProvider_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := map[string]any{"enabled": true}
	rec := invokeJSON("PATCH", "/scm-providers/unknown", "/scm-providers/:provider_key", h.UpdateSCMProvider, body, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateSCMProvider_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	body := map[string]any{"enabled": true}
	rec := invokeJSON("PATCH", "/scm-providers/gitea", "/scm-providers/:provider_key", h.UpdateSCMProvider, body, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListApplications ---

func TestListApplications_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{Key: "APP1", Name: "App 1", Status: domain.ApplicationStatusActive})
	st.seedApp(domain.Application{Key: "APP2", Name: "App 2", Status: domain.ApplicationStatusPlanning})
	rec := invokeJSON("GET", "/applications", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected 2, got %d", len(data))
	}
}

func TestListApplications_StatusFilter(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{Key: "APP1", Name: "App 1", Status: domain.ApplicationStatusActive})
	st.seedApp(domain.Application{Key: "APP2", Name: "App 2", Status: domain.ApplicationStatusPlanning})
	rec := invokeJSON("GET", "/applications?status=active", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1, got %d", len(data))
	}
}

func TestListApplications_InvalidStatus400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications?status=invalid", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplications_InvalidLimit400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications?limit=abc", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplications_LimitOutOfRange400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications?limit=999", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplications_InvalidOffset400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications?offset=-1", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListApplications_LimitOffsetOK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{Key: "APP1", Name: "App 1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/applications?limit=10&offset=0", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListApplications_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errListApplications = errors.New("db")
	rec := invokeJSON("GET", "/applications", "/applications", h.ListApplications, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateApplication ---

func validCreateAppBody() map[string]any {
	return map[string]any{
		"key":                 "APP1",
		"name":                "App 1",
		"description":         "desc",
		"owner_user_id":       "alice",
		"leader_user_id":      "bob",
		"development_unit_id": "team-a",
		"visibility":          "internal",
		"status":              "planning",
		"start_date":          "2026-01-01",
		"due_date":            "2026-12-31",
	}
}

func TestCreateApplication_OK(t *testing.T) {
	h, _, audit := newAppHandlerForTest(t)
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, validCreateAppBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application.created" {
		t.Fatalf("audit not recorded: %+v", audit.created)
	}
}

func TestCreateApplication_BadJSON400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	r := gin.New()
	r.POST("/applications", h.CreateApplication)
	req := httptest.NewRequest("POST", "/applications", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_InvalidKey422(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["key"] = "with-dash"
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "invalid_application_key") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestCreateApplication_MissingName400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["name"] = ""
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_MissingOwner400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["owner_user_id"] = ""
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_MissingLeader400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["leader_user_id"] = ""
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_MissingDevUnit400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["development_unit_id"] = ""
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_InvalidVisibility400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["visibility"] = "weird"
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_InvalidStatus400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["status"] = "weird"
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_BadStartDate400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["start_date"] = "2026/01/01"
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_BadDueDate400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["due_date"] = "not-a-date"
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplication_Conflict409(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{Key: "APP1", Name: "Existing", Status: domain.ApplicationStatusActive})
	body := validCreateAppBody()
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, body, "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplication_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errCreateApplication = errors.New("db")
	rec := invokeJSON("POST", "/applications", "/applications", h.CreateApplication, validCreateAppBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- GetApplication ---

func TestGetApplication_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Name: "App 1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/applications/app-1", "/applications/:application_id", h.GetApplication, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetApplication_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/missing", "/applications/:application_id", h.GetApplication, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetApplication_LookupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errGetApplication = errors.New("db")
	rec := invokeJSON("GET", "/applications/whatever", "/applications/:application_id", h.GetApplication, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetApplication_LinkListError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errListApplicationRepositories = errors.New("db")
	rec := invokeJSON("GET", "/applications/app-1", "/applications/:application_id", h.GetApplication, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ApplicationDashboard ---

func TestApplicationDashboard_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Name: "App", Status: domain.ApplicationStatusActive, LeaderUserID: "leader"})
	rec := invokeJSON("GET", "/applications/app-1/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	if data["application_id"] != "app-1" {
		t.Fatalf("application_id = %v", data["application_id"])
	}
}

func TestApplicationDashboard_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/missing/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationDashboard_WithDevRequestStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := newFakeViewApplicationStore()
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	dreqStore := &fakeViewDevRequestStore{
		dreqs: []domain.DevRequest{
			{ID: "dr-1", Title: "T1", Status: domain.DevRequestStatusPending, RegisteredTargetType: domain.DevRequestTargetApplication, RegisteredTargetID: "app-1"},
			{ID: "dr-2", Title: "T2", Status: domain.DevRequestStatusInReview, RegisteredTargetType: domain.DevRequestTargetApplication, RegisteredTargetID: "other"},
		},
	}
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: st, DevRequestStore: dreqStore})
	rec := invokeJSON("GET", "/applications/app-1/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "dr-1") {
		t.Fatalf("expected dr-1 included, body=%s", rec.Body.String())
	}
}

func TestApplicationDashboard_DevRequestStoreError_DataGap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := newFakeViewApplicationStore()
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	dreqStore := &fakeViewDevRequestStore{errList: errors.New("dreq_down")}
	h := NewApplicationHandler(ApplicationConfig{ApplicationStore: st, DevRequestStore: dreqStore})
	rec := invokeJSON("GET", "/applications/app-1/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "linked_dev_requests:store_error") {
		t.Fatalf("expected data_gap, body=%s", rec.Body.String())
	}
}

func TestApplicationDashboard_RollupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errComputeApplicationRollup = errors.New("rollup_down")
	rec := invokeJSON("GET", "/applications/app-1/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApplicationDashboard_ListProjectsError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errListProjects = errors.New("db")
	rec := invokeJSON("GET", "/applications/app-1/dashboard", "/applications/:application_id/dashboard", h.ApplicationDashboard, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- UpdateApplication ---

func TestUpdateApplication_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "Renamed"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application.updated" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestUpdateApplication_BadJSON400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	r := gin.New()
	r.PATCH("/applications/:application_id", h.UpdateApplication)
	req := httptest.NewRequest("PATCH", "/applications/app-1", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_KeyImmutable422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"key": "NEWKEY"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "application_key_immutable") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestUpdateApplication_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/applications/missing", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_OwnerMismatch403(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateApplication_EmptyName400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "   "}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_EmptyLeader400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"leader_user_id": "   "}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_EmptyDevUnit400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"development_unit_id": "   "}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_InvalidVisibility400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"visibility": "weird"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_InvalidStartDate400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"start_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_InvalidDueDate400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"due_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_InvalidStatus400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"status": "weird"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateApplication_StatusTransitionOK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusPlanning, OwnerUserID: "alice"})
	body := map[string]any{
		"status":      "active",
		"hold_reason": "ignored",
	}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestUpdateApplication_AllPatchFields(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{
		"name":                "Renamed",
		"description":         "newdesc",
		"owner_user_id":       "carol",
		"leader_user_id":      "leader2",
		"development_unit_id": "team-b",
		"visibility":          "public",
		"start_date":          "2026-02-01",
		"due_date":            "2026-11-30",
		"resume_reason":       "back from hold",
		"archived_reason":     "ignored",
	}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateApplication_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	st.errUpdateApplication = errors.New("db")
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/applications/app-1", "/applications/:application_id", h.UpdateApplication, body, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ArchiveApplication ---

func TestArchiveApplication_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"archived_reason": "decommissioned"}
	rec := invokeJSON("DELETE", "/applications/app-1", "/applications/:application_id", h.ArchiveApplication, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application.archived" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchiveApplication_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("DELETE", "/applications/missing", "/applications/:application_id", h.ArchiveApplication, nil, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveApplication_OwnerMismatch403(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/applications/app-1", "/applications/:application_id", h.ArchiveApplication, nil, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveApplication_HardDeleteOK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusArchived, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/applications/app-1?hard=true", "/applications/:application_id", h.ArchiveApplication, nil, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "deleted") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application.deleted" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchiveApplication_HardDeleteNotArchived400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/applications/app-1?hard=true", "/applications/:application_id", h.ArchiveApplication, nil, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestArchiveApplication_HardDeleteStoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusArchived, OwnerUserID: "alice"})
	st.errDeleteApplication = errors.New("db")
	rec := invokeJSON("DELETE", "/applications/app-1?hard=true", "/applications/:application_id", h.ArchiveApplication, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchiveApplication_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive, OwnerUserID: "alice"})
	st.errArchiveApplication = errors.New("db")
	rec := invokeJSON("DELETE", "/applications/app-1", "/applications/:application_id", h.ArchiveApplication, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListApplicationRepositories ---

func TestListApplicationRepositories_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.seedLink(domain.ApplicationRepository{
		ApplicationID: "app-1",
		RepoProvider:  "gitea",
		RepoFullName:  "org/repo",
		Role:          domain.ApplicationRepositoryRole("primary"),
		SyncStatus:    domain.SyncStatusActive,
		LinkSource:    "direct",
	})
	rec := invokeJSON("GET", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.ListApplicationRepositories, nil, "", "")
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

func TestListApplicationRepositories_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errListApplicationRepositories = errors.New("db")
	rec := invokeJSON("GET", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.ListApplicationRepositories, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreateApplicationRepository ---

func validCreateAppRepoBody() map[string]any {
	return map[string]any{
		"repo_provider":    "gitea",
		"repo_full_name":   "org/repo",
		"role":             "primary",
		"external_repo_id": "12345",
	}
}

func TestCreateApplicationRepository_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application_repository.linked" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestCreateApplicationRepository_BadJSON400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	r := gin.New()
	r.POST("/applications/:application_id/repositories", h.CreateApplicationRepository)
	req := httptest.NewRequest("POST", "/applications/app-1/repositories", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationRepository_MissingProvider400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	body := validCreateAppRepoBody()
	body["repo_provider"] = "   "
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationRepository_MissingFullName400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	body := validCreateAppRepoBody()
	body["repo_full_name"] = ""
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationRepository_InvalidRole400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	body := validCreateAppRepoBody()
	body["role"] = "weird"
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationRepository_UnsupportedProvider422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	body := validCreateAppRepoBody()
	body["repo_provider"] = "forgejo" // seeded as disabled
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, body, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported_repo_provider") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestCreateApplicationRepository_Conflict409(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.seedLink(domain.ApplicationRepository{ApplicationID: "app-1", RepoProvider: "gitea", RepoFullName: "org/repo"})
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationRepository_ProviderLookupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errListSCMProviders = errors.New("db")
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateApplicationRepository_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errCreateApplicationRepository = errors.New("db")
	rec := invokeJSON("POST", "/applications/app-1/repositories", "/applications/:application_id/repositories", h.CreateApplicationRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- DeleteApplicationRepository ---

func TestDeleteApplicationRepository_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.seedLink(domain.ApplicationRepository{ApplicationID: "app-1", RepoProvider: "gitea", RepoFullName: "org/repo"})
	r := gin.New()
	r.DELETE("/applications/:application_id/repositories/*repo_key", h.DeleteApplicationRepository)
	req := httptest.NewRequest("DELETE", "/applications/app-1/repositories/gitea:org/repo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "application_repository.unlinked" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestDeleteApplicationRepository_MalformedKey400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	r := gin.New()
	r.DELETE("/applications/:application_id/repositories/*repo_key", h.DeleteApplicationRepository)
	req := httptest.NewRequest("DELETE", "/applications/app-1/repositories/noseparator", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteApplicationRepository_NotFound404(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	r := gin.New()
	r.DELETE("/applications/:application_id/repositories/*repo_key", h.DeleteApplicationRepository)
	req := httptest.NewRequest("DELETE", "/applications/app-1/repositories/gitea:org/missing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteApplicationRepository_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	st.errDeleteApplicationRepository = errors.New("db")
	r := gin.New()
	r.DELETE("/applications/:application_id/repositories/*repo_key", h.DeleteApplicationRepository)
	req := httptest.NewRequest("DELETE", "/applications/app-1/repositories/gitea:org/repo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ApplicationRollup ---

func TestApplicationRollup_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/applications/app-1/rollup", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApplicationRollup_InvalidWeightPolicy400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/app-1/rollup?weight_policy=weird", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationRollup_BadCustomWeights400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/app-1/rollup?weight_policy=custom&custom_weights=not-json", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationRollup_BadFromParam400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/app-1/rollup?from=not-rfc3339", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationRollup_BadToParam400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/applications/app-1/rollup?to=not-rfc3339", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationRollup_FromToOK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	rec := invokeJSON("GET", "/applications/app-1/rollup?from=2026-01-01T00:00:00Z&to=2026-12-31T00:00:00Z", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApplicationRollup_InvalidPolicyFromStore422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Application{ID: "app-1", Key: "APP1", Status: domain.ApplicationStatusActive})
	// custom policy + negative weight → store returns "invalid weight policy" error → handler maps to 422.
	rec := invokeJSON("GET", "/applications/app-1/rollup?weight_policy=custom&custom_weights=%7B%22repo%22%3A-1.0%7D", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApplicationRollup_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errComputeApplicationRollup = errors.New("rollup_down")
	rec := invokeJSON("GET", "/applications/app-1/rollup", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationRollup_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewApplicationHandler(ApplicationConfig{})
	rec := invokeJSON("GET", "/applications/app-1/rollup", "/applications/:application_id/rollup", h.ApplicationRollup, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- 누락된 부분: ensure compile uses context import ---
var _ = context.Background
