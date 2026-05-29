package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeCronStore — in-memory IntakeTokenStore (cron 책임만, sprint -t).
type fakeCronStore struct {
	mu                  sync.Mutex
	revokeInput         time.Time
	revokeReturn        []string
	revokeErr           error
	expiringInput       time.Time
	expiringReturn      int
	expiringErr         error
	staleInput          time.Time
	staleReturn         int
	staleErr            error
	revokeCalls         int
	expiringCalls       int
	staleCalls          int
}

func (f *fakeCronStore) HardRevokeExpiredIntakeTokens(_ context.Context, before time.Time) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revokeInput = before
	f.revokeCalls++
	if f.revokeErr != nil {
		return nil, f.revokeErr
	}
	return append([]string(nil), f.revokeReturn...), nil
}

func (f *fakeCronStore) CountExpiringSoonIntakeTokens(_ context.Context, threshold time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expiringInput = threshold
	f.expiringCalls++
	if f.expiringErr != nil {
		return 0, f.expiringErr
	}
	return f.expiringReturn, nil
}

func (f *fakeCronStore) CountStaleIntakeTokens(_ context.Context, before time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.staleInput = before
	f.staleCalls++
	if f.staleErr != nil {
		return 0, f.staleErr
	}
	return f.staleReturn, nil
}

// runTickOnce — RunIntakeTokenCron 의 ctx 즉시 cancel 후 첫 tick 동작 검증
// (ticker 가 아직 발화 안 한 시점). 본 패턴이 sprint -t 의 cron 1 회 실행
// (startup 직후 즉시 emit) 동작 검증의 핵심.
func runTickOnce(t *testing.T, store *fakeCronStore, opts IntakeTokenCronOptions) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // ctx 즉시 done — runTickOnce 직후 첫 tick 만 실행되고 종료.
	if err := RunIntakeTokenCron(ctx, store, opts); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("RunIntakeTokenCron error: %v", err)
	}
}

func TestIntakeTokenCron_FirstTickInvokesAllThreeQueries(t *testing.T) {
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	store := &fakeCronStore{
		revokeReturn:   []string{"tok-A", "tok-B"},
		expiringReturn: 3,
		staleReturn:    7,
	}
	auditCalls := 0
	var auditPayloads []map[string]any
	emitter := func(_ context.Context, action, _ /*targetType*/, _ /*targetID*/ string, payload map[string]any) {
		auditCalls++
		auditPayloads = append(auditPayloads, payload)
		if action != "dev_request_intake_token.auto_revoked" {
			t.Errorf("unexpected audit action: %q", action)
		}
	}
	opts := IntakeTokenCronOptions{
		Interval:              time.Hour,
		ExpiringSoonThreshold: 24 * time.Hour,
		StaleThreshold:        30 * 24 * time.Hour,
		AuditEmitter:          emitter,
		Now:                   func() time.Time { return now },
	}
	runTickOnce(t, store, opts)

	if store.revokeCalls != 1 {
		t.Errorf("revokeCalls=%d want 1", store.revokeCalls)
	}
	if !store.revokeInput.Equal(now) {
		t.Errorf("revokeInput=%v want %v", store.revokeInput, now)
	}
	if store.expiringCalls != 1 {
		t.Errorf("expiringCalls=%d want 1", store.expiringCalls)
	}
	expectedExpiring := now.Add(24 * time.Hour)
	if !store.expiringInput.Equal(expectedExpiring) {
		t.Errorf("expiringInput=%v want %v", store.expiringInput, expectedExpiring)
	}
	if store.staleCalls != 1 {
		t.Errorf("staleCalls=%d want 1", store.staleCalls)
	}
	expectedStale := now.Add(-30 * 24 * time.Hour)
	if !store.staleInput.Equal(expectedStale) {
		t.Errorf("staleInput=%v want %v", store.staleInput, expectedStale)
	}
	if auditCalls != 2 {
		t.Errorf("auditCalls=%d want 2 (revoke 2 tokens)", auditCalls)
	}
	if len(auditPayloads) != 2 {
		t.Fatalf("auditPayloads len=%d want 2", len(auditPayloads))
	}
	for _, p := range auditPayloads {
		if reason := p["reason"]; reason != "expires_at_elapsed" {
			t.Errorf("audit payload reason=%v want expires_at_elapsed", reason)
		}
	}
}

func TestIntakeTokenCron_StaleDisabledWhenThresholdZero(t *testing.T) {
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	store := &fakeCronStore{
		staleReturn: 99, // 만일 호출되면 99 — but stale disabled 이면 호출 자체 안 함.
	}
	opts := IntakeTokenCronOptions{
		Interval:              time.Hour,
		ExpiringSoonThreshold: 24 * time.Hour,
		StaleThreshold:        0,
		Now:                   func() time.Time { return now },
	}
	runTickOnce(t, store, opts)

	if store.staleCalls != 0 {
		t.Errorf("stale disabled — staleCalls=%d want 0", store.staleCalls)
	}
}

func TestIntakeTokenCron_ExpiringDisabledWhenThresholdZero(t *testing.T) {
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	store := &fakeCronStore{}
	opts := IntakeTokenCronOptions{
		Interval:              time.Hour,
		ExpiringSoonThreshold: 0,
		StaleThreshold:        24 * time.Hour,
		Now:                   func() time.Time { return now },
	}
	runTickOnce(t, store, opts)

	if store.expiringCalls != 0 {
		t.Errorf("expiring disabled — expiringCalls=%d want 0", store.expiringCalls)
	}
}

func TestIntakeTokenCron_StoreErrorsAreLogged_NotPanicked(t *testing.T) {
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	store := &fakeCronStore{
		revokeErr:    errors.New("revoke boom"),
		expiringErr:  errors.New("expiring boom"),
		staleErr:     errors.New("stale boom"),
	}
	opts := IntakeTokenCronOptions{
		Interval:              time.Hour,
		ExpiringSoonThreshold: 24 * time.Hour,
		StaleThreshold:        30 * 24 * time.Hour,
		Now:                   func() time.Time { return now },
	}
	runTickOnce(t, store, opts)

	// 본 test 의 점검 포인트: store 에러가 panic 없이 swallow 되고 cron 이 정상 종료.
	// (3 호출 모두 1회 일어남)
	if store.revokeCalls != 1 || store.expiringCalls != 1 || store.staleCalls != 1 {
		t.Errorf("calls revoke=%d expiring=%d stale=%d want 1/1/1",
			store.revokeCalls, store.expiringCalls, store.staleCalls)
	}
}

func TestIntakeTokenCron_NoRevokesNoAudit(t *testing.T) {
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	store := &fakeCronStore{
		revokeReturn:   []string{}, // 만료 없음
		expiringReturn: 0,
		staleReturn:    0,
	}
	auditCalls := 0
	emitter := func(_ context.Context, _, _, _ string, _ map[string]any) {
		auditCalls++
	}
	opts := IntakeTokenCronOptions{
		Interval:              time.Hour,
		ExpiringSoonThreshold: 24 * time.Hour,
		StaleThreshold:        30 * 24 * time.Hour,
		AuditEmitter:          emitter,
		Now:                   func() time.Time { return now },
	}
	runTickOnce(t, store, opts)

	if auditCalls != 0 {
		t.Errorf("no revokes → auditCalls=%d want 0", auditCalls)
	}
}

func TestIntakeTokenCron_TickerLoop(t *testing.T) {
	store := &fakeCronStore{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	opts := IntakeTokenCronOptions{
		Interval:              2 * time.Millisecond,
		ExpiringSoonThreshold: 24 * time.Hour,
		StaleThreshold:        30 * 24 * time.Hour,
		Now:                   time.Now,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- RunIntakeTokenCron(ctx, store, opts)
	}()

	err := <-errChan
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("expected deadline exceeded or canceled error, got %v", err)
	}

	store.mu.Lock()
	calls := store.revokeCalls
	store.mu.Unlock()

	if calls < 2 {
		t.Errorf("expected multiple ticks, got %d", calls)
	}
}

func TestIntakeTokenCron_DefaultOptions(t *testing.T) {
	store := &fakeCronStore{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 cancel

	opts := IntakeTokenCronOptions{
		Interval:              0, // default Interval 테스트
		ExpiringSoonThreshold: 0,
		StaleThreshold:        0,
		Now:                   nil, // default time.Now 테스트
	}

	err := RunIntakeTokenCron(ctx, store, opts)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected canceled error, got %v", err)
	}
}
