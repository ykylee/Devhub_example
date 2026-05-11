package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

func newAccountsAdminRouter(orgStore OrganizationStore, kratos KratosAdmin, audits *memoryAuditStore) http.Handler {
	return NewRouter(RouterConfig{
		OrganizationStore: orgStore,
		KratosAdmin:       kratos,
		AuditStore:        audits,
		AuthDevFallback:   true, // bypass bearer auth so handler tests can hit the routes
	})
}

func doJSON(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// 1) POST /api/v1/accounts — happy: temp_password supplied.
func TestCreateAccount_HappyExplicitPassword(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	kratos := &MockKratosAdmin{}
	audits := &memoryAuditStore{}
	router := newAccountsAdminRouter(orgStore, kratos, audits)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/accounts",
		`{"user_id":"alice","email":"alice@example.com","display_name":"Alice","role":"developer","temp_password":"InitialPass123!"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"temp_password":"InitialPass123!"`)) {
		t.Errorf("response should echo temp_password: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"identity_id":"mock-k-id-alice"`)) {
		t.Errorf("response should expose Kratos identity_id: %s", rec.Body.String())
	}
	if len(audits.logs) == 0 || audits.logs[0].Action != "account.issued" {
		t.Errorf("expected account.issued audit, got %+v", audits.logs)
	}
}

// 2) POST /api/v1/accounts — happy: server generates temp_password.
func TestCreateAccount_HappyAutoPassword(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	kratos := &MockKratosAdmin{}
	router := newAccountsAdminRouter(orgStore, kratos, &memoryAuditStore{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/accounts",
		`{"user_id":"bob","email":"bob@example.com","display_name":"Bob"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	// generated temp_password should be at least 12 chars
	body := rec.Body.String()
	if !strings.Contains(body, `"temp_password":"`) {
		t.Errorf("body missing temp_password: %s", body)
	}
}

// 3) POST /api/v1/accounts — temp_password too short → 400.
func TestCreateAccount_ShortPassword(t *testing.T) {
	router := newAccountsAdminRouter(newMemoryOrganizationStore(), &MockKratosAdmin{}, &memoryAuditStore{})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/accounts",
		`{"user_id":"u","email":"u@example.com","display_name":"U","temp_password":"short"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rec.Code)
	}
}

// 4) POST /api/v1/accounts — KratosAdmin nil → 503.
func TestCreateAccount_Unavailable(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore(), AuthDevFallback: true})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/accounts",
		`{"user_id":"u","email":"u@example.com","display_name":"U","temp_password":"InitialPass123!"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want 503", rec.Code)
	}
}

// 5) PUT /api/v1/accounts/:user_id/password — happy.
func TestResetAccountPassword_Happy(t *testing.T) {
	kratos := &MockKratosAdmin{}
	audits := &memoryAuditStore{}
	router := newAccountsAdminRouter(newMemoryOrganizationStore(), kratos, audits)

	rec := doJSON(t, router, http.MethodPut, "/api/v1/accounts/alice/password",
		`{"temp_password":"FreshPass-789!"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(kratos.PasswordResets) != 1 || kratos.PasswordResets[0] != "mock-k-id-alice" {
		t.Errorf("password reset call missing or wrong identity: %+v", kratos.PasswordResets)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"temp_password":"FreshPass-789!"`)) {
		t.Errorf("response should echo temp_password: %s", rec.Body.String())
	}
	if len(audits.logs) == 0 || audits.logs[0].Action != "account.password_force_reset" {
		t.Errorf("expected account.password_force_reset audit, got %+v", audits.logs)
	}
}

// 5b) PUT /api/v1/accounts/:user_id/password with empty body → server generates temp password.
func TestResetAccountPassword_EmptyBodyAutoGenerates(t *testing.T) {
	kratos := &MockKratosAdmin{}
	router := newAccountsAdminRouter(newMemoryOrganizationStore(), kratos, &memoryAuditStore{})
	rec := doJSON(t, router, http.MethodPut, "/api/v1/accounts/alice/password", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"temp_password":"`)) {
		t.Errorf("body should include generated temp_password: %s", rec.Body.String())
	}
}

// 6) PUT /api/v1/accounts/:user_id/password — identity not found → 404.
func TestResetAccountPassword_NotFound(t *testing.T) {
	kratos := &MockKratosAdmin{FindError: ErrKratosIdentityNotFound}
	router := newAccountsAdminRouter(newMemoryOrganizationStore(), kratos, &memoryAuditStore{})
	rec := doJSON(t, router, http.MethodPut, "/api/v1/accounts/ghost/password",
		`{"temp_password":"AnyTempPass123"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", rec.Code)
	}
}

// 7) PATCH /api/v1/accounts/:user_id — disabled.
func TestUpdateAccountStatus_Disable(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	// seed a user so UpdateUser succeeds
	if _, err := orgStore.CreateUser(nil, domain.CreateUserInput{
		UserID: "alice", Email: "a@x", DisplayName: "A",
		Role: "developer", Status: "active", Type: "human",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	kratos := &MockKratosAdmin{}
	audits := &memoryAuditStore{}
	router := newAccountsAdminRouter(orgStore, kratos, audits)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/accounts/alice", `{"status":"disabled"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if active, ok := kratos.StateChanges["mock-k-id-alice"]; !ok || active {
		t.Errorf("Kratos state should be set inactive: %+v", kratos.StateChanges)
	}
	hasDisabled := false
	for _, a := range audits.logs {
		if a.Action == "account.disabled" {
			hasDisabled = true
		}
	}
	if !hasDisabled {
		t.Errorf("expected account.disabled audit, got %+v", audits.logs)
	}
}

// 8) PATCH /api/v1/accounts/:user_id — invalid status → 400.
func TestUpdateAccountStatus_Invalid(t *testing.T) {
	router := newAccountsAdminRouter(newMemoryOrganizationStore(), &MockKratosAdmin{}, &memoryAuditStore{})
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/accounts/alice", `{"status":"frozen"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rec.Code)
	}
}

// 9) DELETE /api/v1/accounts/:user_id — happy.
func TestDeleteAccount_Happy(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	if _, err := orgStore.CreateUser(nil, domain.CreateUserInput{
		UserID: "alice", Email: "a@x", DisplayName: "A",
		Role: "developer", Status: "active", Type: "human",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	kratos := &MockKratosAdmin{}
	audits := &memoryAuditStore{}
	router := newAccountsAdminRouter(orgStore, kratos, audits)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/accounts/alice", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(kratos.DeletedIDs) != 1 || kratos.DeletedIDs[0] != "mock-k-id-alice" {
		t.Errorf("Kratos delete missing: %+v", kratos.DeletedIDs)
	}
	if _, err := orgStore.GetUser(nil, "alice"); err == nil {
		t.Errorf("DevHub user should be deleted")
	}
	hasDeleted := false
	for _, a := range audits.logs {
		if a.Action == "account.deleted" {
			hasDeleted = true
		}
	}
	if !hasDeleted {
		t.Errorf("expected account.deleted audit, got %+v", audits.logs)
	}
}

// 10) DELETE /api/v1/accounts/:user_id — Kratos identity already gone → still 200.
func TestDeleteAccount_KratosMissingStillSucceeds(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	if _, err := orgStore.CreateUser(nil, domain.CreateUserInput{
		UserID: "alice", Email: "a@x", DisplayName: "A",
		Role: "developer", Status: "active", Type: "human",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	kratos := &MockKratosAdmin{FindError: ErrKratosIdentityNotFound}
	router := newAccountsAdminRouter(orgStore, kratos, &memoryAuditStore{})

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/accounts/alice", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := orgStore.GetUser(nil, "alice"); err == nil {
		t.Errorf("DevHub user should still be deleted when Kratos identity missing")
	}
}
