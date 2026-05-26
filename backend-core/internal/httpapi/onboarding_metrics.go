// Onboarding 도메인 Prometheus metric — Onboarding 운영 SOP §8 잔여 carve P2.
// audit/metrics.go 패턴 정합 (sync.Once init + registerCollector + Observe helpers).
//
// SOP §4.3 SQL 기반 monitoring 을 metric 으로 자산화 — staging 1주 monitoring
// (SOP §7 DoD) 진입 시 SQL 대안 + Prometheus dashboard 둘 다 사용 가능.
//
// 4 metric:
//   1. devhub_onboarding_gate_blocked_total{reason}        Counter — onboardingGate 403 차단
//   2. devhub_onboarding_submit_total{status}              Counter — POST /me/onboarding 결과
//   3. devhub_onboarding_submit_duration_seconds           Histogram — submit 처리 latency
//   4. devhub_onboarding_review_confirm_total{status}      Counter — POST /admin/users/:id/review 결과
//
// label cardinality bounded:
//   reason ∈ {"onboarding_required"} — 향후 "pending_review_block" 등 확장 carve
//   status ∈ {"ok", "rejected", "conflict", "not_found", "server_error", "unavailable", "unauthenticated"}
//
// pending_review Gauge (SOP §8 잔여) 는 별도 carve — DB SELECT COUNT 부담 + cron
// refresh 패턴 결정 후. 본 sprint 는 Counter / Histogram 4종만.

package httpapi

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	onboardingMetricsOnce sync.Once

	onboardingGateBlockedTotal    *prometheus.CounterVec
	onboardingSubmitTotal         *prometheus.CounterVec
	onboardingSubmitDurationSec   prometheus.Histogram
	onboardingReviewConfirmTotal  *prometheus.CounterVec
)

// initOnboardingMetrics 는 4 metric 을 한 번만 등록. Observe* helpers 가 lazy
// 호출하므로 main.go 의 명시 init 불필요 (audit/metrics.go 와 차이).
func initOnboardingMetrics() {
	onboardingMetricsOnce.Do(func() {
		onboardingGateBlockedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_onboarding_gate_blocked_total",
				Help: "Onboarding gate 가 미완료 actor 를 403 차단한 횟수. reason ∈ {onboarding_required}. ADR-0021 §3.3 / Onboarding SOP §8.",
			},
			[]string{"reason"},
		)
		onboardingSubmitTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_onboarding_submit_total",
				Help: "POST /api/v1/me/onboarding 호출 결과 카운터. status ∈ {ok, rejected, conflict, not_found, server_error, unavailable, unauthenticated}. ADR-0021 / Onboarding SOP §8.",
			},
			[]string{"status"},
		)
		onboardingSubmitDurationSec = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "devhub_onboarding_submit_duration_seconds",
				Help:    "POST /api/v1/me/onboarding 처리 latency (s). status 무관 (handler 전체 측정). Onboarding SOP §8.",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms ~ 10.24s
			},
		)
		onboardingReviewConfirmTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_onboarding_review_confirm_total",
				Help: "POST /api/v1/admin/users/:user_id/review 호출 결과 카운터. status ∈ {ok, rejected, conflict, not_found, server_error, unavailable, bad_request}. ADR-0021 / Onboarding SOP §8.",
			},
			[]string{"status"},
		)
		registerOnboardingCollector(onboardingGateBlockedTotal)
		registerOnboardingCollector(onboardingSubmitTotal)
		registerOnboardingCollector(onboardingSubmitDurationSec)
		registerOnboardingCollector(onboardingReviewConfirmTotal)
	})
}

func registerOnboardingCollector(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

// observeOnboardingGateBlocked — onboardingGate middleware 가 403 차단 직전 호출.
func observeOnboardingGateBlocked(reason string) {
	initOnboardingMetrics()
	onboardingGateBlockedTotal.WithLabelValues(reason).Inc()
}

// observeOnboardingSubmit — submit handler 의 모든 분기 (ok / 각 error code) 에서 호출.
func observeOnboardingSubmit(status string) {
	initOnboardingMetrics()
	onboardingSubmitTotal.WithLabelValues(status).Inc()
}

// observeOnboardingSubmitDuration — submit handler 진입 시점 (defer) 측정한 elapsed seconds.
func observeOnboardingSubmitDuration(seconds float64) {
	initOnboardingMetrics()
	onboardingSubmitDurationSec.Observe(seconds)
}

// observeOnboardingReviewConfirm — confirm handler 의 모든 분기에서 호출.
func observeOnboardingReviewConfirm(status string) {
	initOnboardingMetrics()
	onboardingReviewConfirmTotal.WithLabelValues(status).Inc()
}
