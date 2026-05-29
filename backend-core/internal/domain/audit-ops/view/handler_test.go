package view

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type fakeAuditOpsStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeAuditOpsStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_test_id"
	f.created = append(f.created, log)
	return log, nil
}

func (f *fakeAuditOpsStore) ListAuditLogs(_ context.Context, _ store.ListAuditLogsOptions) ([]domain.AuditLog, error) {
	return nil, nil
}

func TestNewAuditHandler_NonNil(t *testing.T) {
	h := NewAuditHandler(AuditConfig{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewAuditHandler_ConfigPropagation(t *testing.T) {
	store := &fakeAuditOpsStore{}
	cfg := AuditConfig{AuditStore: store, KeycloakWebhookSecret: "secret-x"}
	h := NewAuditHandler(cfg)
	if h.cfg.AuditStore == nil {
		t.Fatal("AuditStore not propagated")
	}
	if h.cfg.KeycloakWebhookSecret != "secret-x" {
		t.Fatalf("KeycloakWebhookSecret = %q", h.cfg.KeycloakWebhookSecret)
	}
}

func TestNewAuditHandler_EmptyConfig(t *testing.T) {
	h := NewAuditHandler(AuditConfig{})
	if h.cfg.AuditStore != nil {
		t.Fatal("AuditStore should be nil by default")
	}
	if h.cfg.KeycloakWebhookSecret != "" {
		t.Fatalf("KeycloakWebhookSecret should be empty, got %q", h.cfg.KeycloakWebhookSecret)
	}
}

func TestAuditLogFromDomain_FullMapping(t *testing.T) {
	created := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	log := domain.AuditLog{
		AuditID:    "audit_1",
		ActorLogin: "alice",
		Action:     "user.created",
		TargetType: "user",
		TargetID:   "u-1",
		CommandID:  "cmd_1",
		Payload:    map[string]any{"k": "v"},
		SourceIP:   "10.0.0.1",
		RequestID:  "req_x",
		SourceType: domain.AuditSourceType("api"),
		CreatedAt:  created,
	}
	got := auditLogFromDomain(log)
	if got.AuditID != "audit_1" || got.ActorLogin != "alice" || got.Action != "user.created" {
		t.Fatalf("basic mapping wrong: %+v", got)
	}
	if got.TargetType != "user" || got.TargetID != "u-1" || got.CommandID != "cmd_1" {
		t.Fatalf("target mapping wrong: %+v", got)
	}
	if got.Payload["k"] != "v" {
		t.Fatalf("payload not preserved: %+v", got.Payload)
	}
	if got.SourceIP != "10.0.0.1" || got.RequestID != "req_x" || got.SourceType != "api" {
		t.Fatalf("source mapping wrong: %+v", got)
	}
	if !got.CreatedAt.Equal(created) {
		t.Fatalf("created_at: got %v", got.CreatedAt)
	}
}

func TestAuditLogFromDomain_NilPayloadDefaultsToEmptyMap(t *testing.T) {
	log := domain.AuditLog{AuditID: "x"}
	got := auditLogFromDomain(log)
	if got.Payload == nil {
		t.Fatal("payload must not be nil")
	}
	if len(got.Payload) != 0 {
		t.Fatalf("payload should be empty: %+v", got.Payload)
	}
}

func TestRecordAudit_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(AuditConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got, err := h.RecordAudit(c, "act", "tt", "tid", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.AuditID != "" {
		t.Fatalf("expected zero, got %+v", got)
	}
}

func TestRecordAudit_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := &fakeAuditOpsStore{}
	h := NewAuditHandler(AuditConfig{AuditStore: st})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got, err := h.RecordAudit(c, "test.action", "user", "u-1", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.AuditID != "audit_test_id" {
		t.Fatalf("expected stamp, got %+v", got)
	}
	if len(st.created) != 1 {
		t.Fatalf("created count = %d", len(st.created))
	}
	c0 := st.created[0]
	if c0.ActorLogin != "alice" || c0.Action != "test.action" {
		t.Fatalf("mapping: %+v", c0)
	}
	if src, _ := c0.Payload["actor_source"].(string); src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_SwallowsErr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(AuditConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.RecordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero on nil store: %+v", got)
	}
}

func TestMapUserEventToAudit_KnownTypes(t *testing.T) {
	cases := map[string]struct {
		action     string
		targetType string
	}{
		"LOGIN":                          {"auth.login.success", "auth"},
		"LOGIN_ERROR":                    {"auth.login.failed", "auth"},
		"LOGOUT":                         {"auth.logout.success", "auth"},
		"LOGOUT_ERROR":                   {"auth.logout.failed", "auth"},
		"REGISTER":                       {"auth.signup.success", "user"},
		"REGISTER_ERROR":                 {"auth.signup.failed", "user"},
		"UPDATE_PASSWORD":                {"auth.password.changed", "user"},
		"UPDATE_PASSWORD_ERROR":          {"auth.password.change_failed", "user"},
		"SEND_RESET_PASSWORD":            {"auth.password.reset_requested", "user"},
		"RESET_PASSWORD":                 {"auth.password.reset_success", "user"},
		"IDENTITY_PROVIDER_LINK_ACCOUNT": {"auth.idp.linked", "user"},
		"IDENTITY_PROVIDER_FIRST_LOGIN":  {"auth.idp.first_login", "user"},
		"VERIFY_EMAIL":                   {"auth.email.verified", "user"},
		"REMOVE_TOTP":                    {"auth.mfa.totp_removed", "user"},
		"UPDATE_TOTP":                    {"auth.mfa.totp_updated", "user"},
	}
	for typ, want := range cases {
		ev := keycloakWebhookRequest{Type: typ, UserID: "u-1"}
		action, targetType, targetID := mapUserEventToAudit(ev)
		if action != want.action {
			t.Errorf("type=%s action got %q want %q", typ, action, want.action)
		}
		if targetType != want.targetType {
			t.Errorf("type=%s targetType got %q want %q", typ, targetType, want.targetType)
		}
		if targetID != "u-1" {
			t.Errorf("type=%s targetID = %q", typ, targetID)
		}
	}
}

func TestMapUserEventToAudit_UnknownTypeFallback(t *testing.T) {
	ev := keycloakWebhookRequest{Type: "WEIRD_EVENT", UserID: "u-1"}
	action, tt, tid := mapUserEventToAudit(ev)
	if action != "keycloak.event.unknown:WEIRD_EVENT" {
		t.Fatalf("action = %q", action)
	}
	if tt != "auth" || tid != "u-1" {
		t.Fatalf("tt=%q tid=%q", tt, tid)
	}
}

func TestMapAdminEventToAudit_KnownTypes(t *testing.T) {
	cases := []struct {
		resType    string
		opType     string
		wantAction string
		wantTarget string
	}{
		{"USER", "CREATE", "keycloak.user.created", "user"},
		{"USER", "UPDATE", "keycloak.user.updated", "user"},
		{"USER", "DELETE", "keycloak.user.deleted", "user"},
		{"USER", "ACTION", "keycloak.user.action", "user"},
		{"REALM_ROLE_MAPPING", "CREATE", "keycloak.user.role.granted", "user"},
		{"REALM_ROLE_MAPPING", "DELETE", "keycloak.user.role.revoked", "user"},
		{"CLIENT", "UPDATE", "keycloak.client.updated", "client"},
		{"REALM", "UPDATE", "keycloak.realm.updated", "realm"},
	}
	for _, c := range cases {
		ev := keycloakWebhookRequest{ResourceType: c.resType, OperationType: c.opType, ResourcePath: "rp"}
		action, tt, tid := mapAdminEventToAudit(ev)
		if action != c.wantAction {
			t.Errorf("%s:%s action = %q, want %q", c.resType, c.opType, action, c.wantAction)
		}
		if tt != c.wantTarget {
			t.Errorf("%s:%s targetType = %q, want %q", c.resType, c.opType, tt, c.wantTarget)
		}
		if tid != "rp" {
			t.Errorf("%s:%s targetID = %q", c.resType, c.opType, tid)
		}
	}
}

func TestMapAdminEventToAudit_UnknownFallback(t *testing.T) {
	ev := keycloakWebhookRequest{ResourceType: "GROUP", OperationType: "CREATE", ResourcePath: "/g/x"}
	action, tt, tid := mapAdminEventToAudit(ev)
	if action != "keycloak.admin.group:create" {
		t.Fatalf("action = %q", action)
	}
	if tt != "realm" || tid != "/g/x" {
		t.Fatalf("tt=%q tid=%q", tt, tid)
	}
}

func TestHashUserEvent_DeterministicAndDistinct(t *testing.T) {
	ev1 := keycloakWebhookRequest{Time: 1, Type: "LOGIN", UserID: "u-1", Details: map[string]string{"sessionId": "s1"}}
	h1 := hashUserEvent(ev1)
	h1Repeat := hashUserEvent(ev1)
	if h1 != h1Repeat {
		t.Fatal("hash must be deterministic")
	}
	ev2 := ev1
	ev2.UserID = "u-2"
	if hashUserEvent(ev2) == h1 {
		t.Fatal("different userID must produce different hash")
	}
}

func TestHashAdminEvent_DeterministicAndDistinct(t *testing.T) {
	ev1 := keycloakWebhookRequest{Time: 1, ResourceType: "USER", OperationType: "CREATE", ResourcePath: "/u/1"}
	h1 := hashAdminEvent(ev1)
	if hashAdminEvent(ev1) != h1 {
		t.Fatal("hash must be deterministic")
	}
	ev2 := ev1
	ev2.OperationType = "DELETE"
	if hashAdminEvent(ev2) == h1 {
		t.Fatal("different op must produce different hash")
	}
}

func TestUserEventPayload_BasicAndError(t *testing.T) {
	ev := keycloakWebhookRequest{
		Type: "LOGIN", ClientID: "cli", RealmID: "realm-x", UserID: "u-1",
		IPAddress: "10.0.0.1", Details: map[string]string{"sessionId": "s1"},
	}
	p := userEventPayload(ev)
	if p["keycloak_event_type"] != "LOGIN" || p["client_id"] != "cli" || p["realm_id"] != "realm-x" {
		t.Fatalf("basic mapping: %+v", p)
	}
	if p["user_id"] != "u-1" || p["ip_address"] != "10.0.0.1" {
		t.Fatalf("user/ip: %+v", p)
	}
	if p["session_id"] != "s1" {
		t.Fatalf("session_id: %+v", p)
	}
	if _, ok := p["error"]; ok {
		t.Fatal("error must be absent when ev.Error empty")
	}

	ev.Error = "oops"
	p2 := userEventPayload(ev)
	if p2["error"] != "oops" {
		t.Fatalf("error not propagated: %+v", p2)
	}
}

func TestAdminEventPayload_BasicAndError(t *testing.T) {
	ev := keycloakWebhookRequest{
		ResourceType: "USER", OperationType: "CREATE", ResourcePath: "/u/1", RealmID: "realm",
	}
	ev.AuthDetails.UserID = "admin-u"
	ev.AuthDetails.ClientID = "admin-cli"
	ev.AuthDetails.IPAddress = "10.0.0.2"
	p := adminEventPayload(ev)
	if p["resource_type"] != "USER" || p["operation_type"] != "CREATE" {
		t.Fatalf("basic: %+v", p)
	}
	if p["auth_user_id"] != "admin-u" || p["auth_client_id"] != "admin-cli" || p["ip_address"] != "10.0.0.2" {
		t.Fatalf("auth: %+v", p)
	}
	if _, ok := p["error"]; ok {
		t.Fatal("error must be absent")
	}
	ev.Error = "boom"
	if adminEventPayload(ev)["error"] != "boom" {
		t.Fatal("error not propagated")
	}
}
