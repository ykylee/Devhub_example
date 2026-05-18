// Package devrequest 는 DREQ 도메인의 Prometheus metric 을 노출한다.
// ADR-0017 §6 carve (c)+(d) — sprint claude/work_260518-t.
//
// HomeLab adapters/metrics.go 패턴 정합: sync.Once init + registerCollector +
// observe helpers. metric 은 backend-core `/metrics` 엔드포인트 (router.go 의
// promhttp.Handler) 가 자동 노출.
package devrequest

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsOnce sync.Once

	intakeTokenExpiringSoon prometheus.Gauge
	intakeTokenStale        prometheus.Gauge
	intakeTokenAutoRevoked  prometheus.Counter
)

// InitMetrics 는 DREQ metric 을 한 번만 등록한다. caller (main.go) 가 cron loop
// 기동 시 1회 호출. main.go 가 metric 초기화 전에 emit 헬퍼를 호출해도 init 이
// lazy 하게 작동하지만, 명시 호출이 startup log 가독성에 도움.
func InitMetrics() {
	metricsOnce.Do(func() {
		intakeTokenExpiringSoon = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_intake_token_expiring_soon",
				Help: "Number of DREQ intake tokens expiring within the configured threshold (default 24h). ADR-0017 §6 (c).",
			},
		)
		intakeTokenStale = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_intake_token_stale",
				Help: "Number of DREQ intake tokens that have not been used within the configured threshold (default 30d). ADR-0017 §6 (d).",
			},
		)
		intakeTokenAutoRevoked = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "devhub_intake_token_auto_revoked_total",
				Help: "Total number of DREQ intake tokens auto-revoked by the cron loop (expires_at <= now). ADR-0017 §6 (a).",
			},
		)
		registerCollector(intakeTokenExpiringSoon)
		registerCollector(intakeTokenStale)
		registerCollector(intakeTokenAutoRevoked)
	})
}

func registerCollector(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

// ObserveExpiringSoon 은 expires_at 임박 token 의 현재 count 를 gauge 로 emit.
// cron tick 마다 호출.
func ObserveExpiringSoon(count int) {
	InitMetrics()
	intakeTokenExpiringSoon.Set(float64(count))
}

// ObserveStale 은 last_used_at 노후 token 의 현재 count 를 gauge 로 emit. cron
// tick 마다 호출. stale threshold == 0 (disabled) 인 경우 caller 가 본 헬퍼를
// 호출하지 않거나 0 을 전달 — 본 헬퍼는 항상 set (감시 dashboard 의 panel 이
// 0 vs "no data" 구분에 의존하지 않게).
func ObserveStale(count int) {
	InitMetrics()
	intakeTokenStale.Set(float64(count))
}

// ObserveAutoRevoked 는 cron 이 hard-revoke 한 token 개수만큼 counter increment.
// revokedCount 는 한 tick 에서 revoke 된 row 의 누계.
func ObserveAutoRevoked(revokedCount int) {
	InitMetrics()
	if revokedCount <= 0 {
		return
	}
	intakeTokenAutoRevoked.Add(float64(revokedCount))
}
