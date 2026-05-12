package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

type memoryAuditStore struct {
	logs []domain.AuditLog
}

func (s *memoryAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if log.AuditID == "" {
		log.AuditID = "audit-test"
		if len(s.logs) > 0 {
			log.AuditID = log.AuditID + "-" + string(rune('0'+len(s.logs)))
		}
	}
	if log.ActorLogin == "" {
		log.ActorLogin = "system"
	}
	if log.Payload == nil {
		log.Payload = map[string]any{}
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Date(2026, 5, 7, 12, len(s.logs), 0, 0, time.UTC)
	}
	s.logs = append(s.logs, log)
	return log, nil
}

func (s *memoryAuditStore) ListAuditLogs(_ context.Context, opts store.ListAuditLogsOptions) ([]domain.AuditLog, error) {
	out := make([]domain.AuditLog, 0, len(s.logs))
	for _, log := range s.logs {
		if opts.ActorLogin != "" && log.ActorLogin != opts.ActorLogin {
			continue
		}
		if opts.Action != "" && log.Action != opts.Action {
			continue
		}
		if opts.TargetType != "" && log.TargetType != opts.TargetType {
			continue
		}
		if opts.TargetID != "" && log.TargetID != opts.TargetID {
			continue
		}
		if opts.CommandID != "" && log.CommandID != opts.CommandID {
			continue
		}
		out = append(out, log)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func TestListAuditLogsFiltersByTarget(t *testing.T) {
	audits := &memoryAuditStore{}
	_, _ = audits.CreateAuditLog(context.Background(), domain.AuditLog{
		ActorLogin: "admin",
		Action:     "user.created",
		TargetType: "user",
		TargetID:   "u1",
	})
	_, _ = audits.CreateAuditLog(context.Background(), domain.AuditLog{
		ActorLogin: "admin",
		Action:     "org_unit.created",
		TargetType: "org_unit",
		TargetID:   "team-a",
	})
	router := testRouter(RouterConfig{AuditStore: audits})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?target_type=user", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []struct {
			Action     string `json:"action"`
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
		} `json:"data"`
		Meta struct {
			Count int `json:"count"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Meta.Count != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected one audit log, got count=%d len=%d", resp.Meta.Count, len(resp.Data))
	}
	if resp.Data[0].Action != "user.created" || resp.Data[0].TargetType != "user" || resp.Data[0].TargetID != "u1" {
		t.Fatalf("unexpected audit log: %+v", resp.Data[0])
	}
}

func TestCreateUserWritesAuditLogWithSystemFallback(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := testRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{
		"user_id": "u-audit",
		"email": "audit@example.com",
		"display_name": "Audit User",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	// X-Devhub-Actor is intentionally still sent; SEC-4 removal must ignore it and the response must not include any deprecation header.
	req.Header.Set("X-Devhub-Actor", "admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Devhub-Actor-Deprecated"); got != "" {
		t.Fatalf("X-Devhub-Actor-Deprecated header must not be set after SEC-4 removal, got %q", got)
	}
	if got := rec.Header().Get("Warning"); got != "" {
		t.Fatalf("Warning header must not be set after SEC-4 removal, got %q", got)
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log, got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.ActorLogin != "system" || log.Action != "user.created" || log.TargetType != "user" || log.TargetID != "u-audit" {
		t.Fatalf("unexpected audit log: %+v", log)
	}
	if log.Payload["actor_source"] != "system_fallback" {
		t.Fatalf("expected actor_source=system_fallback payload, got %+v", log.Payload)
	}
}

// T-M1-04: dev-fallback path tags audit rows as system + every request gets a
// request_id auto-stamped on both the response header (X-Request-ID) and the
// persisted audit row. This is the integration check that closes DoD #3 of
// the original M1 spec for PR-D.
func TestAuditEnrichment_RequestIDMatchesResponseHeaderAndAuditRow(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := testRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{"user_id":"u-trace","email":"trace@x","display_name":"T","role":"developer","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.42:55001"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	headerID := rec.Header().Get("X-Request-ID")
	if headerID == "" {
		t.Fatal("X-Request-ID response header missing — middleware not installed")
	}
	if !startsWithReqPrefix(headerID) {
		t.Errorf("X-Request-ID = %q, want req_ prefix per DEC-3=B", headerID)
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit row, got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.RequestID != headerID {
		t.Errorf("audit RequestID=%q, want match for response header %q", log.RequestID, headerID)
	}
	if log.SourceType != domain.AuditSourceSystem {
		t.Errorf("dev-fallback path should tag SourceType=system, got %q", log.SourceType)
	}
	if log.SourceIP == "" {
		t.Error("audit SourceIP empty; want client IP from gin.Context.ClientIP")
	}
}

// T-M1-04: server-supplied requests that already carry X-Request-ID get the
// existing id honoured rather than overwritten — preserves caller-supplied
// trace ids when DevHub is one hop in a longer chain.
func TestAuditEnrichment_PreservesInboundRequestIDHeader(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := testRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{"user_id":"u-trace2","email":"trace2@x","display_name":"T","role":"developer","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("X-Request-ID", "req_caller_supplied_token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Request-ID"); got != "req_caller_supplied_token" {
		t.Errorf("X-Request-ID = %q, want passthrough", got)
	}
	if len(audits.logs) != 1 || audits.logs[0].RequestID != "req_caller_supplied_token" {
		t.Errorf("audit RequestID = %+v, want passthrough", audits.logs)
	}
}

func startsWithReqPrefix(s string) bool {
	return len(s) > 4 && s[:4] == "req_"
}

// T-M1-04 fix-up (Codex P1, PR #57): NewRouter calls SetTrustedProxies(nil)
// so client-supplied X-Forwarded-For cannot forge audit_logs.source_ip. The
// stored IP must equal the actual peer address (RemoteAddr), regardless of
// what the request advertises in forwarding headers.
func TestAuditEnrichment_SourceIPIgnoresForwardedHeader(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := testRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{"user_id":"u-ip","email":"ip@x","display_name":"I","role":"developer","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.42:55001"
	req.Header.Set("X-Forwarded-For", "1.2.3.4") // attacker-supplied; must be ignored
	req.Header.Set("X-Real-IP", "5.6.7.8")       // attacker-supplied; must be ignored
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit row, got %d", len(audits.logs))
	}
	got := audits.logs[0].SourceIP
	if got == "1.2.3.4" || got == "5.6.7.8" {
		t.Fatalf("audit SourceIP=%q honoured forged forwarding header — SetTrustedProxies(nil) regression", got)
	}
	if got != "203.0.113.42" {
		t.Errorf("audit SourceIP=%q, want 203.0.113.42 (RemoteAddr peer)", got)
	}
}

// PR-D follow-up (work_260512-i): when operators legitimately sit behind a
// reverse proxy and set DEVHUB_TRUSTED_PROXIES, X-Forwarded-For from the
// trusted hop should be honoured so audit_logs.source_ip is the real client
// IP, not the proxy. This pins the env-driven opt-in path.
func TestAuditEnrichment_SourceIPHonoursTrustedProxy(t *testing.T) {
	t.Setenv("DEVHUB_TRUSTED_PROXIES", "203.0.113.42")
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := testRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{"user_id":"u-tp","email":"tp@x","display_name":"T","role":"developer","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.42:55001"
	req.Header.Set("X-Forwarded-For", "198.51.100.7")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit row, got %d", len(audits.logs))
	}
	if got := audits.logs[0].SourceIP; got != "198.51.100.7" {
		t.Errorf("audit SourceIP=%q, want 198.51.100.7 (real client behind trusted proxy)", got)
	}
}

// T-M1-04: bearer-verified request → source_type=oidc. The dev-fallback /
// system split is already covered by
// TestAuditEnrichment_RequestIDMatchesResponseHeaderAndAuditRow above; this
// test pins the OIDC branch (the load-bearing production path).
func TestAuditEnrichment_BearerActorTagsOIDCSource(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "admin", Subject: "user-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	})

	body := []byte(`{"user_id":"u-oidc","email":"o@x","display_name":"O","role":"developer","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit row, got %d", len(audits.logs))
	}
	if got := audits.logs[0].SourceType; got != domain.AuditSourceOIDC {
		t.Errorf("SourceType = %q, want %q", got, domain.AuditSourceOIDC)
	}
}
