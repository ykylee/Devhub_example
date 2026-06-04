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
func newAppHandlerForTest(t *testing.T) (*PlatformHandler, *fakeViewPlatformStore, *fakePlatformLifecycleAuditStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st := newFakeViewPlatformStore()
	st.seedProvider(domain.SCMProvider{ProviderKey: "gitea", DisplayName: "Gitea", Enabled: true, AdapterVersion: "v1"})
	st.seedProvider(domain.SCMProvider{ProviderKey: "github", DisplayName: "GitHub", Enabled: true, AdapterVersion: "v1"})
	st.seedProvider(domain.SCMProvider{ProviderKey: "forgejo", DisplayName: "Forgejo", Enabled: false, AdapterVersion: "v1"})
	audit := &fakePlatformLifecycleAuditStore{}
	h := NewPlatformHandler(PlatformConfig{
		PlatformStore: st,
		AuditStore:       audit,
		ProjectModel:     "hybrid",
	})
	return h, st, audit
}

// invokeJSON — gin engine 생성 + route 등록 + 요청 + recorder 반환.
// admin actor 로 컨텍스트를 세팅해 enforceRowOwnership 통과시킬 수 있도록 middleware 주입.
func invokeJSON(method, path, fullPath string, handler gin.HandlerFunc, body any, actorLogin, actorRole string) *httptest.ResponseRecorder {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if actorLogin == "" && actorRole == "" {
			c.Set("devhub_actor_login", "system")
			c.Set("devhub_actor_role", "system_admin")
			c.Next()
			return
		}
		if actorLogin != "" {
			c.Set("devhub_actor_login", actorLogin)
		}
		if actorRole != "" {
			c.Set("devhub_actor_role", actorRole)
		}
		c.Next()
	})
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
	h := NewPlatformHandler(PlatformConfig{})
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
	h := NewPlatformHandler(PlatformConfig{})
	body := map[string]any{"enabled": true}
	rec := invokeJSON("PATCH", "/scm-providers/gitea", "/scm-providers/:provider_key", h.UpdateSCMProvider, body, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListPlatforms ---

func TestListPlatforms_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{Key: "APP1", Name: "App 1", Status: domain.PlatformStatusActive})
	st.seedApp(domain.Platform{Key: "APP2", Name: "App 2", Status: domain.PlatformStatusPlanning})
	rec := invokeJSON("GET", "/platforms", "/platforms", h.ListPlatforms, nil, "", "")
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

func TestListPlatforms_StatusFilter(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{Key: "APP1", Name: "App 1", Status: domain.PlatformStatusActive})
	st.seedApp(domain.Platform{Key: "APP2", Name: "App 2", Status: domain.PlatformStatusPlanning})
	rec := invokeJSON("GET", "/platforms?status=active", "/platforms", h.ListPlatforms, nil, "", "")
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

func TestListPlatforms_InvalidStatus400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms?status=invalid", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListPlatforms_InvalidLimit400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms?limit=abc", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListPlatforms_LimitOutOfRange400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms?limit=999", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListPlatforms_InvalidOffset400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms?offset=-1", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListPlatforms_LimitOffsetOK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{Key: "APP1", Name: "App 1", Status: domain.PlatformStatusActive})
	rec := invokeJSON("GET", "/platforms?limit=10&offset=0", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListPlatforms_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errListPlatforms = errors.New("db")
	rec := invokeJSON("GET", "/platforms", "/platforms", h.ListPlatforms, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListPlatforms_FiltersToReadableMembership(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Name: "App 1", OwnerUserID: "owner-1", Status: domain.PlatformStatusActive})
	st.seedApp(domain.Platform{ID: "app-2", Key: "APP2", Name: "App 2", OwnerUserID: "owner-2", Status: domain.PlatformStatusActive})
	st.seedProject(domain.Project{
		ID:            "proj-1",
		PlatformID: "app-1",
		Key:           "PRJ1",
		OwnerUserID:   "owner-1",
		Status:        domain.PlatformStatusActive,
		ProjectMembers: []domain.ProjectMember{
			{ProjectID: "proj-1", UserID: "alice", ProjectRole: domain.ProjectMemberRoleContributor},
		},
	})

	rec := invokeJSON("GET", "/platforms", "/platforms", h.ListPlatforms, nil, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 visible application, got %d body=%s", len(data), rec.Body.String())
	}
	first, _ := data[0].(map[string]any)
	if first["id"] != "app-1" {
		t.Fatalf("expected app-1, got %+v", first)
	}
	meta, _ := resp["meta"].(map[string]any)
	if meta["raw_total"] != float64(2) || meta["total"] != float64(1) {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestGetPlatform_DeniesNonMemberRead(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Name: "App 1", OwnerUserID: "owner-1", Status: domain.PlatformStatusActive})

	rec := invokeJSON("GET", "/platforms/app-1", "/platforms/:platform_id", h.GetPlatform, nil, "alice", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "auth_row_denied") || !strings.Contains(rec.Body.String(), "not_platform_member") {
		t.Fatalf("body=%s", rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "auth.row_denied" {
		t.Fatalf("audit=%+v", audit.created)
	}
}

// --- CreatePlatform ---

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

func TestCreatePlatform_OK(t *testing.T) {
	h, _, audit := newAppHandlerForTest(t)
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, validCreateAppBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform.created" {
		t.Fatalf("audit not recorded: %+v", audit.created)
	}
}

func TestCreatePlatform_BadJSON400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	r := gin.New()
	r.POST("/platforms", h.CreatePlatform)
	req := httptest.NewRequest("POST", "/platforms", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_InvalidKey422(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["key"] = "with-dash"
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "invalid_platform_key") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestCreatePlatform_MissingName400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["name"] = ""
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_MissingOwner400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["owner_user_id"] = ""
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_MissingLeader400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["leader_user_id"] = ""
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_MissingDevUnit400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["development_unit_id"] = ""
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_InvalidVisibility400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["visibility"] = "weird"
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_InvalidStatus400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["status"] = "weird"
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_BadStartDate400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["start_date"] = "2026/01/01"
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_BadDueDate400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := validCreateAppBody()
	body["due_date"] = "not-a-date"
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatform_Conflict409(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{Key: "APP1", Name: "Existing", Status: domain.PlatformStatusActive})
	body := validCreateAppBody()
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, body, "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreatePlatform_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errCreatePlatform = errors.New("db")
	rec := invokeJSON("POST", "/platforms", "/platforms", h.CreatePlatform, validCreateAppBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- GetPlatform ---

func TestGetPlatform_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Name: "App 1", Status: domain.PlatformStatusActive})
	rec := invokeJSON("GET", "/platforms/app-1", "/platforms/:platform_id", h.GetPlatform, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPlatform_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/missing", "/platforms/:platform_id", h.GetPlatform, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetPlatform_LookupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.errGetPlatform = errors.New("db")
	rec := invokeJSON("GET", "/platforms/whatever", "/platforms/:platform_id", h.GetPlatform, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetPlatform_LinkListError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errListPlatformRepositories = errors.New("db")
	rec := invokeJSON("GET", "/platforms/app-1", "/platforms/:platform_id", h.GetPlatform, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ApplicationDashboard ---

func TestApplicationDashboard_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Name: "App", Status: domain.PlatformStatusActive, LeaderUserID: "leader"})
	rec := invokeJSON("GET", "/platforms/app-1/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	if data["platform_id"] != "app-1" {
		t.Fatalf("platform_id = %v", data["platform_id"])
	}
}

func TestApplicationDashboard_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/missing/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestApplicationDashboard_WithDevRequestStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := newFakeViewPlatformStore()
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	dreqStore := &fakeViewDevRequestStore{
		dreqs: []domain.DevRequest{
			{ID: "dr-1", Title: "T1", Status: domain.DevRequestStatusPending, RegisteredTargetType: domain.DevRequestTargetPlatform, RegisteredTargetID: "app-1"},
			{ID: "dr-2", Title: "T2", Status: domain.DevRequestStatusInReview, RegisteredTargetType: domain.DevRequestTargetPlatform, RegisteredTargetID: "other"},
		},
	}
	h := NewPlatformHandler(PlatformConfig{PlatformStore: st, DevRequestStore: dreqStore})
	rec := invokeJSON("GET", "/platforms/app-1/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "dr-1") {
		t.Fatalf("expected dr-1 included, body=%s", rec.Body.String())
	}
}

func TestApplicationDashboard_DevRequestStoreError_DataGap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := newFakeViewPlatformStore()
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	dreqStore := &fakeViewDevRequestStore{errList: errors.New("dreq_down")}
	h := NewPlatformHandler(PlatformConfig{PlatformStore: st, DevRequestStore: dreqStore})
	rec := invokeJSON("GET", "/platforms/app-1/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "linked_dev_requests:store_error") {
		t.Fatalf("expected data_gap, body=%s", rec.Body.String())
	}
}

func TestApplicationDashboard_RollupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errComputePlatformRollup = errors.New("rollup_down")
	rec := invokeJSON("GET", "/platforms/app-1/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApplicationDashboard_ListProjectsError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errListProjects = errors.New("db")
	rec := invokeJSON("GET", "/platforms/app-1/dashboard", "/platforms/:platform_id/dashboard", h.PlatformDashboard, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- UpdatePlatform ---

func TestUpdatePlatform_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "Renamed"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform.updated" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestUpdatePlatform_BadJSON400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	r := gin.New()
	r.PATCH("/platforms/:platform_id", h.UpdatePlatform)
	req := httptest.NewRequest("PATCH", "/platforms/app-1", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_KeyImmutable422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"key": "NEWKEY"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "platform_key_immutable") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestUpdatePlatform_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/platforms/missing", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_OwnerMismatch403(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdatePlatform_EmptyName400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"name": "   "}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_EmptyLeader400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"leader_user_id": "   "}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_EmptyDevUnit400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"development_unit_id": "   "}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_InvalidVisibility400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"visibility": "weird"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_InvalidStartDate400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"start_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_InvalidDueDate400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"due_date": "not-a-date"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_InvalidStatus400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"status": "weird"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdatePlatform_StatusTransitionOK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusPlanning, OwnerUserID: "alice"})
	body := map[string]any{
		"status":      "active",
		"hold_reason": "ignored",
	}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestUpdatePlatform_AllPatchFields(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
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
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdatePlatform_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	st.errUpdatePlatform = errors.New("db")
	body := map[string]any{"name": "X"}
	rec := invokeJSON("PATCH", "/platforms/app-1", "/platforms/:platform_id", h.UpdatePlatform, body, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ArchivePlatform ---

func TestArchivePlatform_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	body := map[string]any{"archived_reason": "decommissioned"}
	rec := invokeJSON("DELETE", "/platforms/app-1", "/platforms/:platform_id", h.ArchivePlatform, body, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform.archived" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchivePlatform_NotFound404(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("DELETE", "/platforms/missing", "/platforms/:platform_id", h.ArchivePlatform, nil, "alice", "developer")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchivePlatform_OwnerMismatch403(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/platforms/app-1", "/platforms/:platform_id", h.ArchivePlatform, nil, "bob", "developer")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchivePlatform_HardDeleteOK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusArchived, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/platforms/app-1?hard=true", "/platforms/:platform_id", h.ArchivePlatform, nil, "alice", "developer")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "deleted") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform.deleted" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestArchivePlatform_HardDeleteNotArchived400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	rec := invokeJSON("DELETE", "/platforms/app-1?hard=true", "/platforms/:platform_id", h.ArchivePlatform, nil, "alice", "developer")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestArchivePlatform_HardDeleteStoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusArchived, OwnerUserID: "alice"})
	st.errDeletePlatform = errors.New("db")
	rec := invokeJSON("DELETE", "/platforms/app-1?hard=true", "/platforms/:platform_id", h.ArchivePlatform, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestArchivePlatform_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive, OwnerUserID: "alice"})
	st.errArchivePlatform = errors.New("db")
	rec := invokeJSON("DELETE", "/platforms/app-1", "/platforms/:platform_id", h.ArchivePlatform, nil, "alice", "developer")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- ListPlatformRepositories ---

func TestListApplicationRepositories_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.seedLink(domain.PlatformRepository{
		PlatformID: "app-1",
		RepoProvider:  "gitea",
		RepoFullName:  "org/repo",
		Role:          domain.PlatformRepositoryRole("primary"),
		SyncStatus:    domain.SyncStatusActive,
		LinkSource:    "direct",
	})
	rec := invokeJSON("GET", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.ListPlatformRepositories, nil, "", "")
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
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errListPlatformRepositories = errors.New("db")
	rec := invokeJSON("GET", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.ListPlatformRepositories, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- CreatePlatformRepository ---

func validCreateAppRepoBody() map[string]any {
	return map[string]any{
		"repo_provider":    "gitea",
		"repo_full_name":   "org/repo",
		"role":             "primary",
		"external_repo_id": "12345",
	}
}

func TestCreatePlatformRepository_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform_repository.linked" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestCreatePlatformRepository_BadJSON400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	r := gin.New()
	r.POST("/platforms/:platform_id/repositories", h.CreatePlatformRepository)
	req := httptest.NewRequest("POST", "/platforms/app-1/repositories", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatformRepository_MissingProvider400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	body := validCreateAppRepoBody()
	body["repo_provider"] = "   "
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatformRepository_MissingFullName400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	body := validCreateAppRepoBody()
	body["repo_full_name"] = ""
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatformRepository_InvalidRole400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	body := validCreateAppRepoBody()
	body["role"] = "weird"
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, body, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatformRepository_UnsupportedProvider422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	body := validCreateAppRepoBody()
	body["repo_provider"] = "forgejo" // seeded as disabled
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, body, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported_repo_provider") {
		t.Fatalf("expected code, body=%s", rec.Body.String())
	}
}

func TestCreatePlatformRepository_Conflict409(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.seedLink(domain.PlatformRepository{PlatformID: "app-1", RepoProvider: "gitea", RepoFullName: "org/repo"})
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreatePlatformRepository_ProviderLookupError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errListSCMProviders = errors.New("db")
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreatePlatformRepository_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errCreatePlatformRepository = errors.New("db")
	rec := invokeJSON("POST", "/platforms/app-1/repositories", "/platforms/:platform_id/repositories", h.CreatePlatformRepository, validCreateAppRepoBody(), "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- DeletePlatformRepository ---

func TestDeletePlatformRepository_OK(t *testing.T) {
	h, st, audit := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.seedLink(domain.PlatformRepository{PlatformID: "app-1", RepoProvider: "gitea", RepoFullName: "org/repo"})
	r := gin.New()
	r.DELETE("/platforms/:platform_id/repositories/*repo_key", h.DeletePlatformRepository)
	req := httptest.NewRequest("DELETE", "/platforms/app-1/repositories/gitea:org/repo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audit.created) != 1 || audit.created[0].Action != "platform_repository.unlinked" {
		t.Fatalf("audit: %+v", audit.created)
	}
}

func TestDeletePlatformRepository_MalformedKey400(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	r := gin.New()
	r.DELETE("/platforms/:platform_id/repositories/*repo_key", h.DeletePlatformRepository)
	req := httptest.NewRequest("DELETE", "/platforms/app-1/repositories/noseparator", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeletePlatformRepository_NotFound404(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	r := gin.New()
	r.DELETE("/platforms/:platform_id/repositories/*repo_key", h.DeletePlatformRepository)
	req := httptest.NewRequest("DELETE", "/platforms/app-1/repositories/gitea:org/missing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeletePlatformRepository_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errDeletePlatformRepository = errors.New("db")
	r := gin.New()
	r.DELETE("/platforms/:platform_id/repositories/*repo_key", h.DeletePlatformRepository)
	req := httptest.NewRequest("DELETE", "/platforms/app-1/repositories/gitea:org/repo", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- PlatformRollup ---

func TestPlatformRollup_OK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	rec := invokeJSON("GET", "/platforms/app-1/rollup", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformRollup_InvalidWeightPolicy400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/app-1/rollup?weight_policy=weird", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPlatformRollup_BadCustomWeights400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/app-1/rollup?weight_policy=custom&custom_weights=not-json", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPlatformRollup_BadFromParam400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/app-1/rollup?from=not-rfc3339", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPlatformRollup_BadToParam400(t *testing.T) {
	h, _, _ := newAppHandlerForTest(t)
	rec := invokeJSON("GET", "/platforms/app-1/rollup?to=not-rfc3339", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPlatformRollup_FromToOK(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	rec := invokeJSON("GET", "/platforms/app-1/rollup?from=2026-01-01T00:00:00Z&to=2026-12-31T00:00:00Z", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformRollup_InvalidPolicyFromStore422(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	// custom policy + negative weight → store returns "invalid weight policy" error → handler maps to 422.
	rec := invokeJSON("GET", "/platforms/app-1/rollup?weight_policy=custom&custom_weights=%7B%22repo%22%3A-1.0%7D", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlatformRollup_StoreError500(t *testing.T) {
	h, st, _ := newAppHandlerForTest(t)
	st.seedApp(domain.Platform{ID: "app-1", Key: "APP1", Status: domain.PlatformStatusActive})
	st.errComputePlatformRollup = errors.New("rollup_down")
	rec := invokeJSON("GET", "/platforms/app-1/rollup", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPlatformRollup_StoreUnavailable503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlatformHandler(PlatformConfig{})
	rec := invokeJSON("GET", "/platforms/app-1/rollup", "/platforms/:platform_id/rollup", h.PlatformRollup, nil, "", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

// --- 누락된 부분: ensure compile uses context import ---
var _ = context.Background
