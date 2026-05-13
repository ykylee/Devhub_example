package httpapi

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
	"github.com/devhub/backend-core/internal/hrdb"
)

// signupFakeHRDB satisfies router.HRDBClient. Returns (email, userID, dept, err)
// for the seeded triple; otherwise hrdb.ErrPersonNotFound.
type signupFakeHRDB struct {
	hits  int
	miss  bool // when true, always return ErrPersonNotFound
}

func (f *signupFakeHRDB) Lookup(_ context.Context, systemID, employeeID, name string) (string, string, string, error) {
	f.hits++
	if f.miss {
		return "", "", "", hrdb.ErrPersonNotFound
	}
	if systemID == "yklee" && employeeID == "1001" && strings.EqualFold(name, "YK Lee") {
		return "yklee@example.com", "yklee", "Engineering", nil
	}
	return "", "", "", hrdb.ErrPersonNotFound
}

// fakeKratosAdminForSignup is a narrow Kratos admin double for signup paths.
// signup only needs CreateIdentity; the rest of the interface returns
// not-applicable errors so a misuse surfaces loudly.
type fakeKratosAdminForSignup struct {
	createCalls    int
	failOnCreate   bool
	createdUserID  string
	createdEmail   string
	createdName    string
	createdKratos  string
}

func (f *fakeKratosAdminForSignup) CreateIdentity(_ context.Context, email, name, userID, _ string) (string, error) {
	f.createCalls++
	if f.failOnCreate {
		return "", errors.New("kratos boom")
	}
	f.createdUserID = userID
	f.createdEmail = email
	f.createdName = name
	f.createdKratos = "kratos-id-" + userID
	return f.createdKratos, nil
}

func (f *fakeKratosAdminForSignup) FindIdentityByUserID(_ context.Context, _ string) (string, error) {
	return "", errors.New("not used in signup tests")
}

func (f *fakeKratosAdminForSignup) UpdateIdentityPassword(_ context.Context, _, _ string) error {
	return errors.New("not used in signup tests")
}

func (f *fakeKratosAdminForSignup) SetIdentityState(_ context.Context, _ string, _ bool) error {
	return errors.New("not used in signup tests")
}

func (f *fakeKratosAdminForSignup) DeleteIdentity(_ context.Context, _ string) error {
	return errors.New("not used in signup tests")
}

func signupBody() []byte {
	return []byte(`{
		"name": "YK Lee",
		"system_id": "yklee",
		"employee_id": "1001",
		"password": "ChangeMe-12345!"
	}`)
}

func TestSignUp_HappyPath_CreatesIdentityUserAndAudit(t *testing.T) {
	// RM-M3-01: hrdb hit → Kratos identity → DevHub user → audit emit
	// (account.signup.requested). Asserts the canonical happy path so
	// future regressions of any step (audit drop, Kratos call elision,
	// payload shape change) surface.
	hr := &signupFakeHRDB{}
	kratos := &fakeKratosAdminForSignup{}
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}

	router := testRouter(RouterConfig{
		HRDB:              hr,
		KratosAdmin:       kratos,
		OrganizationStore: orgs,
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hr.hits != 1 {
		t.Errorf("expected exactly one HRDB lookup, got %d", hr.hits)
	}
	if kratos.createCalls != 1 || kratos.createdUserID != "yklee" {
		t.Errorf("expected one CreateIdentity for yklee, got calls=%d userID=%q", kratos.createCalls, kratos.createdUserID)
	}

	var response struct {
		Status string `json:"status"`
		Data   struct {
			UserID     string `json:"user_id"`
			KratosID   string `json:"kratos_id"`
			Department string `json:"department"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "created" || response.Data.UserID != "yklee" || response.Data.Department != "Engineering" || response.Data.KratosID == "" {
		t.Fatalf("unexpected response data: %+v", response.Data)
	}

	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log (account.signup.requested), got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.Action != "account.signup.requested" || log.TargetType != "user" || log.TargetID != "yklee" {
		t.Errorf("unexpected audit row: action=%q target=%s/%s", log.Action, log.TargetType, log.TargetID)
	}
	if log.Payload["kratos_id"] == nil || log.Payload["email"] != "yklee@example.com" || log.Payload["department"] != "Engineering" || log.Payload["system_id"] != "yklee" {
		t.Errorf("unexpected audit payload: %+v", log.Payload)
	}
}

func TestSignUp_HRDBMissReturns403(t *testing.T) {
	hr := &signupFakeHRDB{miss: true}
	kratos := &fakeKratosAdminForSignup{}
	audits := &memoryAuditStore{}

	router := testRouter(RouterConfig{
		HRDB:              hr,
		KratosAdmin:       kratos,
		OrganizationStore: newMemoryOrganizationStore(),
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "hr_lookup_failed") {
		t.Errorf("expected body code=hr_lookup_failed, got %q", rec.Body.String())
	}
	if kratos.createCalls != 0 {
		t.Errorf("Kratos CreateIdentity must not run when HRDB lookup fails, got calls=%d", kratos.createCalls)
	}
	if len(audits.logs) != 0 {
		t.Errorf("no audit row when signup rejected upstream, got %d", len(audits.logs))
	}
}

func TestSignUp_KratosCreateFailureReturns500(t *testing.T) {
	hr := &signupFakeHRDB{}
	kratos := &fakeKratosAdminForSignup{failOnCreate: true}
	audits := &memoryAuditStore{}

	router := testRouter(RouterConfig{
		HRDB:              hr,
		KratosAdmin:       kratos,
		OrganizationStore: newMemoryOrganizationStore(),
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}
	if kratos.createCalls != 1 {
		t.Errorf("expected CreateIdentity attempt, got calls=%d", kratos.createCalls)
	}
	if len(audits.logs) != 0 {
		t.Errorf("no audit row when Kratos identity creation failed, got %d", len(audits.logs))
	}
}

func TestSignUp_DevHubUserConflictEmitsPartialFailureAudit(t *testing.T) {
	// RM-M3-01 carve: when CreateUser conflicts (user_id already exists),
	// the Kratos identity has already been created — leaving an orphan
	// pair. Until the next sprint adds rollback/retry, we emit
	// account.signup.partial_failure so the operator dashboard can find
	// the discrepancy.
	hr := &signupFakeHRDB{}
	kratos := &fakeKratosAdminForSignup{}
	orgs := newMemoryOrganizationStore()
	// Pre-seed an existing user_id so the second CreateUser (from the
	// signup handler) hits ErrConflict.
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "yklee",
		Email:       "yklee@example.com",
		DisplayName: "YK Lee (seeded)",
		Role:        domain.AppRole("developer"),
		Status:      domain.UserStatus("active"),
	}); err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	audits := &memoryAuditStore{}

	router := testRouter(RouterConfig{
		HRDB:              hr,
		KratosAdmin:       kratos,
		OrganizationStore: orgs,
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 (Kratos id still created), got %d body=%s", rec.Code, rec.Body.String())
	}
	if kratos.createCalls != 1 {
		t.Errorf("expected one CreateIdentity, got calls=%d", kratos.createCalls)
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log (partial_failure), got %d", len(audits.logs))
	}
	if audits.logs[0].Action != "account.signup.partial_failure" {
		t.Errorf("expected action=account.signup.partial_failure, got %q", audits.logs[0].Action)
	}
	if audits.logs[0].Payload["reason"] != "devhub_user_create_failed" {
		t.Errorf("expected reason=devhub_user_create_failed, got %v", audits.logs[0].Payload["reason"])
	}
}
