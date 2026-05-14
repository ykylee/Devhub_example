package httpapi

import (
	"bytes"
	"net/http"
	"testing"
)

// Integration CRUD handler tests (API-58, sprint claude/work_260514-c).

// 1) POST /integrations — happy (scope=application).
func TestCreateIntegration_HappyApplicationScope(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integrations",
		`{"scope":"application","application_id":"app-1","integration_type":"jira","external_key":"PROJ","url":"https://example.atlassian.net","policy":"summary_only"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"scope":"application"`)) {
		t.Errorf("expected scope=application: %s", rec.Body.String())
	}
}

// 2) POST /integrations — scope=application 인데 application_id 없음 → 422.
func TestCreateIntegration_ScopeTargetMismatch(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integrations",
		`{"scope":"application","integration_type":"jira","external_key":"PROJ","url":"https://x","policy":"summary_only"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"scope_target_mismatch"`)) {
		t.Errorf("expected scope_target_mismatch: %s", rec.Body.String())
	}
}

// 3) POST /integrations — invalid policy → 400.
func TestCreateIntegration_InvalidPolicy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integrations",
		`{"scope":"application","application_id":"app-1","integration_type":"jira","external_key":"X","url":"https://x","policy":"forbidden"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 4) POST /integrations — duplicate (same scope+target+type+key) → 409.
func TestCreateIntegration_Duplicate(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	body := `{"scope":"application","application_id":"app-1","integration_type":"jira","external_key":"PROJ","url":"https://x","policy":"summary_only"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/integrations", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integrations", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 5) GET /integrations — empty list.
func TestListIntegrations_Empty(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/integrations", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"data":[]`)) {
		t.Errorf("expected empty data: %s", rec.Body.String())
	}
}

// 6) DELETE /integrations/:id — not_found → 404.
func TestDeleteIntegration_NotFound(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integrations/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 7) PR #107 codex review P2 회귀 guard — PATCH external_key 가 다른 integration
// 의 키와 충돌하면 409 integration_conflict.
func TestUpdateIntegration_UniqueConflict(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	// 2 개의 integration 생성 (same scope/target/type, 다른 external_key)
	first := doJSON(t, router, http.MethodPost, "/api/v1/integrations",
		`{"scope":"application","application_id":"app-1","integration_type":"jira","external_key":"PROJ-A","url":"https://x","policy":"summary_only"}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed 1 failed: %s", first.Body.String())
	}
	second := doJSON(t, router, http.MethodPost, "/api/v1/integrations",
		`{"scope":"application","application_id":"app-1","integration_type":"jira","external_key":"PROJ-B","url":"https://x","policy":"summary_only"}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("seed 2 failed: %s", second.Body.String())
	}
	// second.id 를 추출 — response 의 "id" 필드.
	// 단순화: memory store 의 ID 컨벤션이 "int-{external_key}" 라 알려져 있음.
	secondID := "int-PROJ-B"
	// PROJ-B → PROJ-A 로 변경 시 first 와 충돌 → 409.
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/integrations/"+secondID,
		`{"external_key":"PROJ-A"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s (expected 409 integration_conflict)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"integration_conflict"`)) {
		t.Errorf("expected integration_conflict: %s", rec.Body.String())
	}
}
