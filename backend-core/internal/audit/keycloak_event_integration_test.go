package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/store"
)

// TestEndToEnd_HTTPAPIAdapter_ToAuditEmitter — sprint -v PR-C integration test.
// HTTPAPIEventLister 어댑터 → audit.KeycloakEventLister → pullOnce → AuditEmitter
// 흐름을 1개 process 안에서 검증.
//
// scenario: tick 1 에서 user event 1건 (LOGIN) + admin event 1건 (USER:CREATE) 발생.
// tick 2 에서 동일 boundary event 가 dateFrom inclusive 로 반복 등장 — dedup hash 로 skip.
// tick 3 에서 신규 user event 1건 발생 — emit.
// 최종 audit emit = 3건 (tick1 user + tick1 admin + tick3 user).
func TestEndToEnd_HTTPAPIAdapter_ToAuditEmitter(t *testing.T) {
	t0 := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(10 * time.Second)
	t2 := t0.Add(20 * time.Second)
	t3 := t0.Add(30 * time.Second)

	lister := &mockHTTPAPILister{
		userEventsByTick: [][]HTTPAPIUserEvent{
			// tick 1
			{{Time: t1.UnixMilli(), Type: "LOGIN", UserID: "alice", IPAddr: "10.0.0.1"}},
			// tick 2 — dateFrom inclusive 라 t1 의 LOGIN 이 다시 등장. hash dedup 으로 skip.
			{{Time: t1.UnixMilli(), Type: "LOGIN", UserID: "alice", IPAddr: "10.0.0.1"}},
			// tick 3 — 신규 event
			{
				{Time: t1.UnixMilli(), Type: "LOGIN", UserID: "alice", IPAddr: "10.0.0.1"},  // boundary 잔재 → dedup skip
				{Time: t3.UnixMilli(), Type: "LOGOUT", UserID: "alice", IPAddr: "10.0.0.1"}, // 신규 → emit
			},
		},
		adminEventsByTick: [][]HTTPAPIAdminEvent{
			// tick 1
			{
				{
					Time:          t2.UnixMilli(),
					OperationType: "CREATE",
					ResourceType:  "USER",
					ResourcePath:  "users/abc-1",
					AuthUserID:    "admin-1",
					AuthIPAddr:    "10.0.0.99",
				},
			},
			// tick 2 — boundary 잔재
			{
				{
					Time:          t2.UnixMilli(),
					OperationType: "CREATE",
					ResourceType:  "USER",
					ResourcePath:  "users/abc-1",
					AuthUserID:    "admin-1",
					AuthIPAddr:    "10.0.0.99",
				},
			},
			// tick 3 — 신규 admin event 없음
			nil,
		},
	}

	cursorStore := newFakeCursorStore()
	// 사전 cursor 셋업 — t0 (스타트업 직후 first-run init 시뮬레이션)
	_ = cursorStore.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   userEventsCursor,
		LastEventAt: t0,
	})
	_ = cursorStore.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   adminEventsCursor,
		LastEventAt: t0,
	})

	var emitted []emittedRecord
	var emittedMu sync.Mutex
	emitter := AuditEmitter(func(_ context.Context, action, targetType, targetID string, payload map[string]any) {
		emittedMu.Lock()
		defer emittedMu.Unlock()
		emitted = append(emitted, emittedRecord{action: action, targetType: targetType, targetID: targetID, payload: payload})
	})

	adapter := NewHTTPAPIEventListerAdapter(lister)
	opts := KeycloakEventPullerOptions{
		AuditEmitter:       emitter,
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}

	ctx := context.Background()
	// tick 1
	if err := pullOnce(ctx, adapter, cursorStore, opts, func() time.Time { return t1.Add(1 * time.Second) }); err != nil {
		t.Fatalf("tick 1: %v", err)
	}
	// tick 2 — dateFrom inclusive 로 boundary 재등장. emit 추가 없어야 함.
	if err := pullOnce(ctx, adapter, cursorStore, opts, func() time.Time { return t2.Add(1 * time.Second) }); err != nil {
		t.Fatalf("tick 2: %v", err)
	}
	// tick 3 — 신규 user event 1건 emit.
	if err := pullOnce(ctx, adapter, cursorStore, opts, func() time.Time { return t3.Add(1 * time.Second) }); err != nil {
		t.Fatalf("tick 3: %v", err)
	}

	if len(emitted) != 3 {
		t.Fatalf("emitted = %d events; want 3 (tick1 user LOGIN + tick1 admin USER:CREATE + tick3 user LOGOUT)", len(emitted))
	}
	if emitted[0].action != "auth.login.success" || emitted[0].targetID != "alice" {
		t.Fatalf("emitted[0] = %+v; want auth.login.success alice", emitted[0])
	}
	if emitted[1].action != "keycloak.user.created" || emitted[1].targetID != "users/abc-1" {
		t.Fatalf("emitted[1] = %+v; want keycloak.user.created users/abc-1", emitted[1])
	}
	if emitted[2].action != "auth.logout.success" || emitted[2].targetID != "alice" {
		t.Fatalf("emitted[2] = %+v; want auth.logout.success alice", emitted[2])
	}

	// 어댑터 payload 가 ip_address / user_id / auth_user_id 등 payload 키 포함 확인.
	if ip, ok := emitted[0].payload["ip_address"].(string); !ok || ip != "10.0.0.1" {
		t.Fatalf("emitted[0].payload.ip_address = %v; want 10.0.0.1", emitted[0].payload["ip_address"])
	}
	if uid, ok := emitted[1].payload["auth_user_id"].(string); !ok || uid != "admin-1" {
		t.Fatalf("emitted[1].payload.auth_user_id = %v; want admin-1", emitted[1].payload["auth_user_id"])
	}

	// final cursor 가 t3 (user) / t2 (admin) 으로 advance 검증.
	userCursor, _ := cursorStore.GetEventCursor(ctx, userEventsCursor)
	if !userCursor.LastEventAt.Equal(t3) {
		t.Fatalf("user cursor = %v; want t3 (%v)", userCursor.LastEventAt, t3)
	}
	adminCursor, _ := cursorStore.GetEventCursor(ctx, adminEventsCursor)
	if !adminCursor.LastEventAt.Equal(t2) {
		t.Fatalf("admin cursor = %v; want t2 (%v)", adminCursor.LastEventAt, t2)
	}
}

type emittedRecord struct {
	action     string
	targetType string
	targetID   string
	payload    map[string]any
}

type mockHTTPAPILister struct {
	mu                sync.Mutex
	userEventsByTick  [][]HTTPAPIUserEvent
	adminEventsByTick [][]HTTPAPIAdminEvent
	userCallCount     int
	adminCallCount    int
}

func (m *mockHTTPAPILister) ListUserEvents(_ context.Context, _ time.Time, _ int) ([]HTTPAPIUserEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := m.userCallCount
	m.userCallCount++
	if idx >= len(m.userEventsByTick) {
		return nil, nil
	}
	return m.userEventsByTick[idx], nil
}

func (m *mockHTTPAPILister) ListAdminEvents(_ context.Context, _ time.Time, _ int) ([]HTTPAPIAdminEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := m.adminCallCount
	m.adminCallCount++
	if idx >= len(m.adminEventsByTick) {
		return nil, nil
	}
	return m.adminEventsByTick[idx], nil
}
