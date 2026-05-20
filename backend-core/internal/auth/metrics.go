// Package auth 의 Prometheus metric — JWKS stale-while-error 운영 가시성.
// ADR-0020 sub-carve D (sprint -l, issue #213). 본 sprint 가 keycloak_verifier
// 의 cache fetch 실패 시 stale cache fallback 흐름을 추가하면서 운영 dashboard
// 가 fallback 활성 여부 + stale_age 분포 추적 가능하도록 metric 등록.
//
// audit/metrics.go + devrequest/metrics.go + integrations/adapters/metrics.go
// 패턴 정합 — sync.Once init + registerCollector + observe helpers.
package auth

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	jwksMetricsOnce sync.Once

	jwksStaleWhileErrorTotal *prometheus.CounterVec
	jwksStaleAgeSeconds      prometheus.Histogram
)

// initJWKSMetrics — keycloak_verifier 가 stale fallback path 진입 직전 호출.
// sync.Once 로 중복 등록 회피.
func initJWKSMetrics() {
	jwksMetricsOnce.Do(func() {
		jwksStaleWhileErrorTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_jwks_stale_while_error_total",
				Help: "JWKS cache fetch 실패 후 stale-while-error fallback 사용 횟수. result ∈ {ok, fail} (ok=stale 사용 성공, fail=stale 도 없거나 만료). ADR-0020 §5.6.",
			},
			[]string{"result"},
		)
		jwksStaleAgeSeconds = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "devhub_jwks_stale_age_seconds",
				Help:    "stale-while-error fallback 사용 시 cache 의 age (s). cachedAt 이후 경과 시간. 운영 dashboard 가 MaxStaleDuration 임계 접근 식별. ADR-0020 §5.6.",
				Buckets: prometheus.ExponentialBuckets(60, 2, 12), // 1m ~ 4096m (~68h)
			},
		)
		registerJWKSCollector(jwksStaleWhileErrorTotal)
		registerJWKSCollector(jwksStaleAgeSeconds)
	})
}

func registerJWKSCollector(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

// observeJWKSStaleWhileError 는 fetchJWKS 가 network fail 후 stale fallback
// path 에 들어갈 때마다 1회 호출. result ∈ {ok, fail}.
func observeJWKSStaleWhileError(result string) {
	initJWKSMetrics()
	jwksStaleWhileErrorTotal.WithLabelValues(result).Inc()
}

// observeJWKSStaleAge 는 result=ok 분기에서만 호출 — stale cache 의 age 분포
// 추적. fail 분기는 stale 자체가 없으므로 age 미정의.
func observeJWKSStaleAge(seconds float64) {
	initJWKSMetrics()
	jwksStaleAgeSeconds.Observe(seconds)
}
