package view

import (
	"context"
	"log"
	"time"
)

// Onboarding pending_review count Gauge cron worker — SOP §8 잔여 carve P3 진입.
// audit/keycloak_event_puller.go 의 RunKeycloakEventPuller 패턴 정합 (single
// goroutine + time.Ticker + ctx cancellation).
//
// 사유: handler 호출 마다 SELECT COUNT 호출은 read 부담. cron refresh (default
// 60s interval) 로 update — alert SLA (15~30분 검토 latency) 보다 충분히 짧음.

// OnboardingPendingReviewCounter — RunOnboardingPendingReviewGauge 가 의존하는
// store interface 부분집합. circular import + 작은 interface 정합 — main.go 가
// *store.PostgresStore (CountPendingReview 구현) 를 그대로 전달.
type OnboardingPendingReviewCounter interface {
	CountPendingReview(ctx context.Context) (int, error)
}

// OnboardingPendingGaugeOptions — cron loop 설정.
type OnboardingPendingGaugeOptions struct {
	// Interval — tick 주기. default 60s.
	Interval time.Duration
}

// RunOnboardingPendingReviewGauge — pending_review Gauge refresh cron worker.
// ctx cancel 까지 실행. main.go 에서 `go RunOnboardingPendingReviewGauge(...)` 패턴.
//
// 초기 1차 tick 즉시 실행 (운영자가 startup 직후 Gauge 값 즉시 확인 — audit
// puller 패턴 정합). error 시 stderr log + 다음 tick 재시도 (Gauge 값 stale 그대로 유지).
func RunOnboardingPendingReviewGauge(
	ctx context.Context,
	counter OnboardingPendingReviewCounter,
	opts OnboardingPendingGaugeOptions,
) error {
	if opts.Interval <= 0 {
		opts.Interval = 60 * time.Second
	}
	initOnboardingMetrics()

	log.Printf("[onboarding-pending-gauge] starting (interval=%s)", opts.Interval)

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// 초기 1차 tick 즉시 (startup 직후 Gauge 즉시 가시화).
	if err := refreshOnce(ctx, counter); err != nil {
		log.Printf("[onboarding-pending-gauge] initial tick: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[onboarding-pending-gauge] stopping: %v", ctx.Err())
			return ctx.Err()
		case <-ticker.C:
			if err := refreshOnce(ctx, counter); err != nil {
				log.Printf("[onboarding-pending-gauge] tick: %v", err)
			}
		}
	}
}

// refreshOnce — 1회 SELECT COUNT → Gauge set. error 시 caller (RunOnboardingPending
// ReviewGauge) 가 log + Gauge stale 보존.
func refreshOnce(ctx context.Context, counter OnboardingPendingReviewCounter) error {
	count, err := counter.CountPendingReview(ctx)
	if err != nil {
		return err
	}
	SetOnboardingPendingReviewCount(count)
	return nil
}
