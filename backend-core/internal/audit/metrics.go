// Package audit 의 Prometheus metric — Keycloak event polling 의 진행/지연 가시성.
// ADR-0019 §5.3 (9) audit event listener Phase 2 PR-C (sprint claude/work_260519-v).
//
// devrequest/metrics.go + integrations/adapters/metrics.go 패턴 정합 — sync.Once init +
// registerCollector + observe helpers. main.go 가 cron loop 기동 시 1회 InitMetrics 호출.
package audit

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsOnce sync.Once

	eventsProcessedTotal *prometheus.CounterVec
	cursorLagSeconds     *prometheus.GaugeVec
	pullErrorsTotal      *prometheus.CounterVec
)

// InitMetrics 는 Keycloak event listener metric 을 한 번만 등록한다. caller (main.go)
// 가 cron loop 기동 시 1회 호출. lazy init 으로도 작동하지만 명시 호출이 startup log
// 가독성에 도움.
func InitMetrics() {
	metricsOnce.Do(func() {
		eventsProcessedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_keycloak_events_processed_total",
				Help: "Total Keycloak events emitted to audit_logs by event-listener cron, by kind (user/admin) and action.",
			},
			[]string{"kind", "action"},
		)
		cursorLagSeconds = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "devhub_keycloak_event_cursor_lag_seconds",
				Help: "Lag (now - cursor.LastEventAt) in seconds for each Keycloak event cursor key (keycloak.events / keycloak.events.admin).",
			},
			[]string{"cursor_key"},
		)
		pullErrorsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_keycloak_event_pull_errors_total",
				Help: "Total Keycloak event-listener cron pull errors, by kind (user/admin).",
			},
			[]string{"kind"},
		)
		registerCollector(eventsProcessedTotal)
		registerCollector(cursorLagSeconds)
		registerCollector(pullErrorsTotal)
	})
}

func registerCollector(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

// ObserveEventProcessed 는 audit emit 1건 직후 호출 (kind = "user" 또는 "admin").
// action label 은 normalizeMetricAction 으로 bounded — unknown fallback 의 unbounded
// TYPE suffix 가 Prometheus label cardinality 를 폭증시키지 않도록 unified label 사용.
func ObserveEventProcessed(kind, action string) {
	InitMetrics()
	eventsProcessedTotal.WithLabelValues(kind, normalizeMetricAction(action)).Inc()
}

// normalizeMetricAction — metric label cardinality 를 bounded 로 유지. 매핑 표가
// 만든 known action prefix (`auth.*` / `keycloak.user.*` / `keycloak.client.*` /
// `keycloak.realm.*`) 는 그대로 사용. unknown fallback (`keycloak.event.unknown:*` 또는
// `keycloak.admin.*` 미매핑 조합) 은 "unknown" 으로 unified — audit_logs.action 의
// unique 값은 그대로 보존, metric 만 normalize.
func normalizeMetricAction(action string) string {
	if strings.HasPrefix(action, "keycloak.event.unknown:") {
		return "keycloak.event.unknown"
	}
	return action
}

// ObserveCursorLag 는 매 poll tick 직후 호출. lagSeconds = now - cursor.LastEventAt.
// 운영 dashboard 가 tick interval 보다 lag 가 크게 누적되면 알람.
func ObserveCursorLag(cursorKey string, lagSeconds float64) {
	InitMetrics()
	cursorLagSeconds.WithLabelValues(cursorKey).Set(lagSeconds)
}

// ObservePullError 는 tick 중 pullUserEvents / pullAdminEvents error 시 호출.
func ObservePullError(kind string) {
	InitMetrics()
	pullErrorsTotal.WithLabelValues(kind).Inc()
}
