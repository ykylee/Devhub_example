package audit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/store"
)

// fakeKeycloakEventLister — ListUserEvents / ListAdminEvents 의 in-memory mock.
type fakeKeycloakEventLister struct {
	userEvents  []KeycloakUserEvent
	adminEvents []KeycloakAdminEvent
	userCalls   int
	adminCalls  int
}

func (f *fakeKeycloakEventLister) ListUserEvents(_ context.Context, _ time.Time, _ int) ([]KeycloakUserEvent, error) {
	f.userCalls++
	return f.userEvents, nil
}

func (f *fakeKeycloakEventLister) ListAdminEvents(_ context.Context, _ time.Time, _ int) ([]KeycloakAdminEvent, error) {
	f.adminCalls++
	return f.adminEvents, nil
}

// fakeCursorStore — store.EventCursorStore 의 in-memory mock.
// GetEventCursor 가 row 없을 때 "not found" 포함 error 를 반환해야 loadCursor 가
// initialization branch 로 분기. isNotFound (puller.go) 가 string match 사용.
type fakeCursorStore struct {
	mu      sync.Mutex
	cursors map[string]store.EventCursor
}

func newFakeCursorStore() *fakeCursorStore {
	return &fakeCursorStore{cursors: make(map[string]store.EventCursor)}
}

func (s *fakeCursorStore) GetEventCursor(_ context.Context, cursorKey string) (store.EventCursor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cursors[cursorKey]
	if !ok {
		return store.EventCursor{}, fmt.Errorf("cursor %s not found", cursorKey)
	}
	return c, nil
}

func (s *fakeCursorStore) UpsertEventCursor(_ context.Context, cursor store.EventCursor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursors[cursor.CursorKey] = cursor
	return nil
}

func TestMapUserEventToAudit(t *testing.T) {
	cases := []struct {
		typ, action, targetType string
	}{
		{"LOGIN", "auth.login.success", "auth"},
		{"LOGIN_ERROR", "auth.login.failed", "auth"},
		{"LOGOUT", "auth.logout.success", "auth"},
		{"REGISTER", "auth.signup.success", "user"},
		{"UPDATE_PASSWORD", "auth.password.changed", "user"},
		{"REMOVE_TOTP", "auth.mfa.totp_removed", "user"},
		{"VERIFY_EMAIL", "auth.email.verified", "user"},
	}
	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			action, tt, _ := mapUserEventToAudit(KeycloakUserEvent{Type: tc.typ, UserID: "u1"})
			if action != tc.action {
				t.Fatalf("action = %q; want %q", action, tc.action)
			}
			if tt != tc.targetType {
				t.Fatalf("targetType = %q; want %q", tt, tc.targetType)
			}
		})
	}
}

func TestMapUserEventToAudit_UnknownTypeFallback(t *testing.T) {
	action, _, _ := mapUserEventToAudit(KeycloakUserEvent{Type: "CUSTOM_TYPE", UserID: "u1"})
	if action != "keycloak.event.unknown:CUSTOM_TYPE" {
		t.Fatalf("unknown fallback = %q", action)
	}
}

func TestMapAdminEventToAudit(t *testing.T) {
	cases := []struct {
		resource, operation, action, targetType string
	}{
		{"USER", "CREATE", "keycloak.user.created", "user"},
		{"USER", "DELETE", "keycloak.user.deleted", "user"},
		{"REALM_ROLE_MAPPING", "CREATE", "keycloak.user.role.granted", "user"},
		{"REALM_ROLE_MAPPING", "DELETE", "keycloak.user.role.revoked", "user"},
		{"CLIENT", "UPDATE", "keycloak.client.updated", "client"},
		{"REALM", "UPDATE", "keycloak.realm.updated", "realm"},
	}
	for _, tc := range cases {
		t.Run(tc.resource+":"+tc.operation, func(t *testing.T) {
			action, tt, _ := mapAdminEventToAudit(KeycloakAdminEvent{
				ResourceType:  tc.resource,
				OperationType: tc.operation,
				ResourcePath:  "users/u1",
			})
			if action != tc.action {
				t.Fatalf("action = %q; want %q", action, tc.action)
			}
			if tt != tc.targetType {
				t.Fatalf("targetType = %q; want %q", tt, tc.targetType)
			}
		})
	}
}

func TestMapAdminEventToAudit_UnknownKeyFallback(t *testing.T) {
	action, _, _ := mapAdminEventToAudit(KeycloakAdminEvent{
		ResourceType:  "GROUP",
		OperationType: "CREATE",
		ResourcePath:  "groups/g1",
	})
	if action != "keycloak.admin.group:create" {
		t.Fatalf("unknown fallback = %q", action)
	}
}

// TestPullUserEvents_SkipsConfiguredEventTypes — REFRESH_TOKEN 은 default skip.
// LOGIN 만 emit.
func TestPullUserEvents_SkipsConfiguredEventTypes(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	lister := &fakeKeycloakEventLister{
		userEvents: []KeycloakUserEvent{
			{Time: now.Add(1 * time.Second).UnixMilli(), Type: "REFRESH_TOKEN", UserID: "u1"},
			{Time: now.Add(2 * time.Second).UnixMilli(), Type: "LOGIN", UserID: "u1"},
		},
	}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   userEventsCursor,
		LastEventAt: now,
	})

	var emitted []string
	emitter := AuditEmitter(func(_ context.Context, action, _, _, _ string, _ map[string]any) {
		emitted = append(emitted, action)
	})

	opts := KeycloakEventPullerOptions{
		AuditEmitter:       emitter,
		Now:                func() time.Time { return now },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}

	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted = %v; want 1 (LOGIN), REFRESH_TOKEN must be skipped", emitted)
	}
	if emitted[0] != "auth.login.success" {
		t.Fatalf("emitted[0] = %q; want auth.login.success", emitted[0])
	}
}

// TestPullUserEvents_AdvancesCursor — pull 후 cursor 가 최신 event timestamp 로 갱신.
func TestPullUserEvents_AdvancesCursor(t *testing.T) {
	start := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	later := start.Add(10 * time.Second)
	lister := &fakeKeycloakEventLister{
		userEvents: []KeycloakUserEvent{
			{Time: later.UnixMilli(), Type: "LOGIN", UserID: "u1"},
		},
	}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   userEventsCursor,
		LastEventAt: start,
	})

	opts := KeycloakEventPullerOptions{
		AuditEmitter:       AuditEmitter(func(_ context.Context, _, _, _, _ string, _ map[string]any) {}),
		Now:                func() time.Time { return start },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}
	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	got, err := cursors.GetEventCursor(context.Background(), userEventsCursor)
	if err != nil {
		t.Fatalf("GetEventCursor: %v", err)
	}
	if !got.LastEventAt.Equal(later) {
		t.Fatalf("LastEventAt = %v; want %v", got.LastEventAt, later)
	}
}

// TestPullUserEvents_NoEvents_DoesNotUpsertCursor — 빈 결과 시 cursor 유지.
func TestPullUserEvents_NoEvents_DoesNotUpsertCursor(t *testing.T) {
	start := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	lister := &fakeKeycloakEventLister{userEvents: nil}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   userEventsCursor,
		LastEventAt: start,
	})
	opts := KeycloakEventPullerOptions{
		Now:                func() time.Time { return start },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}
	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	got, _ := cursors.GetEventCursor(context.Background(), userEventsCursor)
	if !got.LastEventAt.Equal(start) {
		t.Fatalf("cursor changed unexpectedly: %v (want unchanged %v)", got.LastEventAt, start)
	}
}

// TestPullUserEvents_FiltersAlreadyProcessed — cursor 보다 strictly before event 는 skip.
// boundary (동일 ms) event 는 cursor.LastEventHash 와 hash 같으면 skip (이미 처리한 그 event),
// hash 다르면 emit (boundary 새 event). after cursor 는 무조건 emit.
func TestPullUserEvents_FiltersAlreadyProcessed(t *testing.T) {
	cursor := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	// boundary event = cursor 시각에 이미 처리한 event (hash 도 사전 기록됨).
	processedBoundary := KeycloakUserEvent{Time: cursor.UnixMilli(), Type: "LOGIN", UserID: "u2-boundary"}
	processedHash := hashUserEvent(processedBoundary)

	lister := &fakeKeycloakEventLister{
		userEvents: []KeycloakUserEvent{
			{Time: cursor.Add(-1 * time.Second).UnixMilli(), Type: "LOGIN", UserID: "u1-before"}, // strictly before cursor → skip
			processedBoundary, // equal cursor + same hash → skip (이미 처리한 그 event)
			{Time: cursor.UnixMilli(), Type: "LOGIN", UserID: "u2b-new-boundary"},              // equal cursor + 다른 hash → emit
			{Time: cursor.Add(1 * time.Second).UnixMilli(), Type: "LOGIN", UserID: "u3-after"}, // after cursor → emit
		},
	}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:     userEventsCursor,
		LastEventAt:   cursor,
		LastEventHash: processedHash,
	})

	var emitted []string
	opts := KeycloakEventPullerOptions{
		AuditEmitter: AuditEmitter(func(_ context.Context, _, _, targetID, _ string, _ map[string]any) {
			emitted = append(emitted, targetID)
		}),
		Now:                func() time.Time { return cursor },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}
	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	if len(emitted) != 2 || emitted[0] != "u2b-new-boundary" || emitted[1] != "u3-after" {
		t.Fatalf("emitted = %v; want [u2b-new-boundary, u3-after] (before + processed-boundary skipped)", emitted)
	}
}

// TestPullUserEvents_BoundarySameHash_Skipped — design §3.3 의 명시적 hash dedup 검증.
// dateFrom inclusive 특성으로 직전 처리한 boundary event 가 반복 등장해도 skip.
func TestPullUserEvents_BoundarySameHash_Skipped(t *testing.T) {
	cursor := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	boundary := KeycloakUserEvent{Time: cursor.UnixMilli(), Type: "LOGIN", UserID: "u-dup"}
	h := hashUserEvent(boundary)

	lister := &fakeKeycloakEventLister{userEvents: []KeycloakUserEvent{boundary}}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:     userEventsCursor,
		LastEventAt:   cursor,
		LastEventHash: h,
	})

	var emitted int
	opts := KeycloakEventPullerOptions{
		AuditEmitter: AuditEmitter(func(_ context.Context, _, _, _, _ string, _ map[string]any) {
			emitted++
		}),
		Now:                func() time.Time { return cursor },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}
	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	if emitted != 0 {
		t.Fatalf("emitted = %d; want 0 (boundary same-hash must be deduped per design §3.3)", emitted)
	}
}

// TestPullAdminEvents_AdvancesCursor — admin event 도 cursor 갱신.
func TestPullAdminEvents_AdvancesCursor(t *testing.T) {
	start := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	later := start.Add(5 * time.Second)
	lister := &fakeKeycloakEventLister{
		adminEvents: []KeycloakAdminEvent{
			{Time: later.UnixMilli(), ResourceType: "USER", OperationType: "CREATE", ResourcePath: "users/u1"},
		},
	}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   adminEventsCursor,
		LastEventAt: start,
	})

	var emitted []string
	opts := KeycloakEventPullerOptions{
		AuditEmitter: AuditEmitter(func(_ context.Context, action, _, _, _ string, _ map[string]any) {
			emitted = append(emitted, action)
		}),
		Now:       func() time.Time { return start },
		MaxEvents: 500,
	}
	if err := pullAdminEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullAdminEvents: %v", err)
	}
	if len(emitted) != 1 || emitted[0] != "keycloak.user.created" {
		t.Fatalf("emitted = %v; want [keycloak.user.created]", emitted)
	}
	got, _ := cursors.GetEventCursor(context.Background(), adminEventsCursor)
	if !got.LastEventAt.Equal(later) {
		t.Fatalf("LastEventAt = %v; want %v", got.LastEventAt, later)
	}
}

// TestLoadCursor_NotFound_InitializesToNow — 첫 run 시 cursor 없으면 now 로 초기화.
// design §3.3 — 과거 event 모두 폭격 회피.
func TestLoadCursor_NotFound_InitializesToNow(t *testing.T) {
	cursors := newFakeCursorStore()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	c, err := loadCursor(context.Background(), cursors, "missing-key", func() time.Time { return now })
	if err != nil {
		t.Fatalf("loadCursor first-run should succeed, got: %v", err)
	}
	if !c.LastEventAt.Equal(now) {
		t.Fatalf("LastEventAt = %v; want now (%v)", c.LastEventAt, now)
	}
	if c.CursorKey != "missing-key" {
		t.Fatalf("CursorKey = %q; want missing-key", c.CursorKey)
	}
}

// TestLoadCursor_PropagatesNonNotFoundError — store 의 다른 error 는 그대로 전파.
func TestLoadCursor_PropagatesNonNotFoundError(t *testing.T) {
	cursors := &cursorStoreErr{err: errors.New("connection lost")}
	_, err := loadCursor(context.Background(), cursors, "any-key", time.Now)
	if err == nil {
		t.Fatal("expected error to propagate")
	}
}

type cursorStoreErr struct{ err error }

func (s *cursorStoreErr) GetEventCursor(_ context.Context, _ string) (store.EventCursor, error) {
	return store.EventCursor{}, s.err
}
func (s *cursorStoreErr) UpsertEventCursor(_ context.Context, _ store.EventCursor) error {
	return s.err
}

// TestLoadCursor_NotFound_PersistsSeed — sprint -y codex hotfix #10 P1-B 정정 검증.
// row 미존재 시 in-memory init 만 하면 첫 poll 이 빈 결과 시 영구화 안 됨 → 다음 tick
// now() 재초기화 → tick 사이 event 미수신. 정정 후 loadCursor 가 즉시 UPSERT 호출.
func TestLoadCursor_NotFound_PersistsSeed(t *testing.T) {
	cursors := newFakeCursorStore()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	_, err := loadCursor(context.Background(), cursors, "keycloak.events", func() time.Time { return now })
	if err != nil {
		t.Fatalf("loadCursor first-run: %v", err)
	}
	// fakeCursorStore.GetEventCursor 가 row 영구화되었는지 검증.
	persisted, err := cursors.GetEventCursor(context.Background(), "keycloak.events")
	if err != nil {
		t.Fatalf("GetEventCursor after seed: %v", err)
	}
	if !persisted.LastEventAt.Equal(now) {
		t.Fatalf("persisted.LastEventAt = %v; want now (%v)", persisted.LastEventAt, now)
	}
}

// TestPullUserEvents_SkipOnlyPage_AdvancesCursor — sprint -y codex hotfix #10 P1-A 정정 검증.
// 페이지가 default skip type (REFRESH_TOKEN 등) 만 들어와도 cursor 가 advance 되어야
// 다음 tick 이 동일 페이지 무한 재pull 하지 않음.
func TestPullUserEvents_SkipOnlyPage_AdvancesCursor(t *testing.T) {
	start := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	later := start.Add(5 * time.Second)
	lister := &fakeKeycloakEventLister{
		userEvents: []KeycloakUserEvent{
			{Time: later.UnixMilli(), Type: "REFRESH_TOKEN", UserID: "u-noisy"},
		},
	}
	cursors := newFakeCursorStore()
	_ = cursors.UpsertEventCursor(context.Background(), store.EventCursor{
		CursorKey:   userEventsCursor,
		LastEventAt: start,
	})

	var emitted int
	opts := KeycloakEventPullerOptions{
		AuditEmitter: AuditEmitter(func(_ context.Context, _, _, _, _ string, _ map[string]any) {
			emitted++
		}),
		Now:                func() time.Time { return later.Add(1 * time.Second) },
		SkipUserEventTypes: defaultSkipUserEventTypes(),
		MaxEvents:          500,
	}
	if err := pullUserEvents(context.Background(), lister, cursors, opts, opts.Now); err != nil {
		t.Fatalf("pullUserEvents: %v", err)
	}
	if emitted != 0 {
		t.Fatalf("emitted = %d; want 0 (REFRESH_TOKEN 만 들어왔으므로 audit 없음)", emitted)
	}
	// 핵심 검증 — cursor 가 later 로 advance.
	got, err := cursors.GetEventCursor(context.Background(), userEventsCursor)
	if err != nil {
		t.Fatalf("GetEventCursor: %v", err)
	}
	if !got.LastEventAt.Equal(later) {
		t.Fatalf("LastEventAt = %v; want %v (skip type 만 들어와도 cursor advance 필요)", got.LastEventAt, later)
	}
}

// TestHashUserEvent_DistinguishesByClientID — sprint -y codex hotfix #10 P2-D 정정 검증.
// 같은 (time, type, userID, ipAddr) 의 distinct event 가 client/realm/sessionId 다를 때
// 다른 hash 를 생성해야 store-level dedup (UNIQUE INDEX) 이 burst 동시 ms event 를
// audit_logs 에서 잃지 않음.
func TestHashUserEvent_DistinguishesByClientID(t *testing.T) {
	base := KeycloakUserEvent{
		Time:    1747641600000,
		Type:    "LOGIN",
		UserID:  "alice",
		IPAddr:  "10.0.0.1",
		RealmID: "devhub",
	}
	clientA := base
	clientA.ClientID = "devhub-frontend"
	clientB := base
	clientB.ClientID = "devhub-other-client"
	clientASession := base
	clientASession.ClientID = "devhub-frontend"
	clientASession.Details = map[string]string{"sessionId": "sess-A"}
	clientASession2 := base
	clientASession2.ClientID = "devhub-frontend"
	clientASession2.Details = map[string]string{"sessionId": "sess-B"}

	h1 := hashUserEvent(clientA)
	h2 := hashUserEvent(clientB)
	if h1 == h2 {
		t.Fatalf("hash collision: same hash for distinct client_id (%q vs %q): %s", clientA.ClientID, clientB.ClientID, h1)
	}
	h3 := hashUserEvent(clientASession)
	h4 := hashUserEvent(clientASession2)
	if h3 == h4 {
		t.Fatalf("hash collision: same hash for distinct sessionId: %s", h3)
	}
}

// TestHashAdminEvent_DistinguishesByAuthClient — admin event 도 동일 검증.
func TestHashAdminEvent_DistinguishesByAuthClient(t *testing.T) {
	base := KeycloakAdminEvent{
		Time:          1747641600000,
		ResourceType:  "USER",
		OperationType: "UPDATE",
		ResourcePath:  "users/u1",
		AuthUserID:    "admin-x",
		RealmID:       "devhub",
	}
	a := base
	a.AuthClientID = "client-A"
	a.AuthIPAddr = "10.0.0.1"
	b := base
	b.AuthClientID = "client-B"
	b.AuthIPAddr = "10.0.0.1"
	if hashAdminEvent(a) == hashAdminEvent(b) {
		t.Fatalf("admin hash collision on distinct AuthClientID")
	}
}
