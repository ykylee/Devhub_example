package service

import (
	"context"
	"log"
	"time"
)

// IntakeTokenStore 는 cron loop 가 의존하는 store 메서드의 minimal interface.
// httpapi.IntakeTokenStore 의 부분집합 — cron 책임만 노출 (테스트성 ↑).
type IntakeTokenStore interface {
	HardRevokeExpiredIntakeTokens(ctx context.Context, before time.Time) ([]string, error)
	CountExpiringSoonIntakeTokens(ctx context.Context, threshold time.Time) (int, error)
	CountStaleIntakeTokens(ctx context.Context, before time.Time) (int, error)
}

// AuditEmitter 는 cron 이 hard-revoke 한 token 마다 audit log 를 남기는 콜백.
// httpapi 의 recordAudit best-effort 와 동일 contract — 실패해도 cron 자체는 계속.
type AuditEmitter func(ctx context.Context, action, targetType, targetID string, payload map[string]any)

// IntakeTokenCronOptions 는 cron loop 동작 설정.
type IntakeTokenCronOptions struct {
	// Interval 은 tick 주기 (예: 10 * time.Minute).
	Interval time.Duration
	// ExpiringSoonThreshold 는 expires_at - NOW() <= 이 값인 활성 token 을
	// "expiring soon" 으로 카운트 (예: 24h). 0 이하 면 expiring_soon metric 을
	// 항상 0 으로 emit (운영자가 임계 미설정 명시).
	ExpiringSoonThreshold time.Duration
	// StaleThreshold 는 last_used_at <= NOW() - 이 값인 활성 token 을 "stale"
	// 로 카운트 (예: 30 * 24h). 0 이하 면 stale metric 은 0 으로 emit (disable).
	StaleThreshold time.Duration
	// AuditEmitter 가 nil 이면 audit 생략 (테스트 환경 / 운영자가 audit 분리).
	AuditEmitter AuditEmitter
	// Now 는 시간 주입 (테스트). nil 이면 time.Now() 사용.
	Now func() time.Time
}

// RunIntakeTokenCron 는 DREQ intake token cron loop 를 ctx cancel 까지 실행한다.
// ADR-0017 §6 carve (a)+(c)+(d) — sprint claude/work_260518-t.
//
// 매 tick:
//  1. HardRevokeExpiredIntakeTokens(now) → revoke 된 token ID list 반환.
//     각 ID 마다 audit emit + counter `devhub_intake_token_auto_revoked_total` increment.
//  2. CountExpiringSoonIntakeTokens(now + ExpiringSoonThreshold) → gauge
//     `devhub_intake_token_expiring_soon` 갱신.
//  3. CountStaleIntakeTokens(now - StaleThreshold) → gauge
//     `devhub_intake_token_stale` 갱신 (StaleThreshold == 0 이면 0 emit).
//
// HomeLab pull loop (adapters.RunHomeLabPullLoop) 패턴 정합 — single goroutine +
// ctx cancellation. tick error 는 log 만 + 다음 tick 계속 (blast radius 격리).
func RunIntakeTokenCron(ctx context.Context, store IntakeTokenStore, opts IntakeTokenCronOptions) error {
	if opts.Interval <= 0 {
		opts.Interval = 10 * time.Minute
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	InitMetrics()

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// 첫 tick 은 즉시 실행 (startup 직후 metric/cron 가시화 — 운영 대시보드의 cold start 단축).
	runIntakeTokenCronTick(ctx, store, opts, now())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t := <-ticker.C:
			runIntakeTokenCronTick(ctx, store, opts, t)
		}
	}
}

func runIntakeTokenCronTick(ctx context.Context, store IntakeTokenStore, opts IntakeTokenCronOptions, now time.Time) {
	// 1. hard-revoke 만료 token.
	revokedIDs, err := store.HardRevokeExpiredIntakeTokens(ctx, now)
	if err != nil {
		log.Printf("intake_token_cron: hard revoke failed: %v", err)
	} else if len(revokedIDs) > 0 {
		ObserveAutoRevoked(len(revokedIDs))
		if opts.AuditEmitter != nil {
			for _, tokenID := range revokedIDs {
				opts.AuditEmitter(ctx, "dev_request_intake_token.auto_revoked", "dev_request_intake_token", tokenID, map[string]any{
					"reason":     "expires_at_elapsed",
					"revoked_at": now.UTC().Format(time.RFC3339),
				})
			}
		}
	}

	// 2. expiring_soon gauge.
	if opts.ExpiringSoonThreshold > 0 {
		threshold := now.Add(opts.ExpiringSoonThreshold)
		count, cErr := store.CountExpiringSoonIntakeTokens(ctx, threshold)
		if cErr != nil {
			log.Printf("intake_token_cron: count expiring soon failed: %v", cErr)
		} else {
			ObserveExpiringSoon(count)
		}
	} else {
		ObserveExpiringSoon(0)
	}

	// 3. stale gauge.
	if opts.StaleThreshold > 0 {
		before := now.Add(-opts.StaleThreshold)
		count, cErr := store.CountStaleIntakeTokens(ctx, before)
		if cErr != nil {
			log.Printf("intake_token_cron: count stale failed: %v", cErr)
		} else {
			ObserveStale(count)
		}
	} else {
		// disabled — 0 emit 으로 운영 대시보드의 "no data" 와 명확 구분.
		ObserveStale(0)
	}
}
