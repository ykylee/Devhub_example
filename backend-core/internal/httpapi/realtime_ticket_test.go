package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	realtimeview "github.com/devhub/backend-core/internal/domain/realtime/view"
	"github.com/devhub/backend-core/internal/store"
)

// ADR-0024 §3.2 ticket pattern — router-level integration test. 본 파일은 NewRouter
// 의 ticket store 분기 (fault → 503 / miss → 401 / access_token query 무시 → 401) 만
// 검증. pure store unit test 는 internal/domain/realtime/view/realtime_ticket_test.go.

// faultyTicketStore — Consume 가 항상 store fault 를 반환. realtimeview.RealtimeTicketStore
// interface (Issue/Consume) 구현.
type faultyTicketStore struct{}

func (faultyTicketStore) Issue(_ context.Context, _, _ string, _ domain.AuditSourceType) (string, error) {
	return "", nil
}
func (faultyTicketStore) Consume(_ context.Context, _ string) (store.RealtimeTicket, bool, error) {
	return store.RealtimeTicket{}, false, errors.New("postgres connection lost")
}

// codex review PR #344 — auth flow 가 ticket store fault (5xx) 와 invalid ticket
// (401) 를 구분해야 한다. store fault 시 valid ticket 이 401 로 거부되거나 infra
// 신호가 사라지면 안 됨.
func TestRealtimeWS_TicketStoreFault_Returns503(t *testing.T) {
	r := NewRouter(RouterConfig{
		RealtimeHub:     realtimeview.NewRealtimeHub(),
		RealtimeTickets: faultyTicketStore{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws?ticket=maybe-valid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("ticket store fault should return 503, got %d", rec.Code)
	}
}

// 진짜 miss (unknown/expired) 는 401 — store fault 와 구분.
func TestRealtimeWS_TicketMiss_Returns401(t *testing.T) {
	r := NewRouter(RouterConfig{
		RealtimeHub:     realtimeview.NewRealtimeHub(),
		RealtimeTickets: realtimeview.NewInMemoryRealtimeTicketStore(), // 비어 있음 → consume miss
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws?ticket=unknown", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown ticket should return 401, got %d", rec.Code)
	}
}

// ADR-0024 §6 carve 5 (ticket-only 컷오버) — 레거시 `?access_token=` query 는 더
// 이상 honor 되지 않는다. token 을 무조건 수락하는 verifier 가 붙어 있어도 WS
// 경로의 access_token query 는 무시되어 401 이어야 한다 (회귀 가드).
func TestRealtimeWS_AccessTokenQuery_NoLongerHonored(t *testing.T) {
	r := NewRouter(RouterConfig{
		RealtimeHub:         realtimeview.NewRealtimeHub(),
		RealtimeTickets:     realtimeview.NewInMemoryRealtimeTicketStore(),
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{Login: "x", Role: "developer"}},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws?access_token=would-be-valid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("access_token query must no longer be honored (ticket-only), expected 401, got %d", rec.Code)
	}
}
