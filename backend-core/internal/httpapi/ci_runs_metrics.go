package httpapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CI Run ingest metrics — sprint mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4).
//
// devhub_ci_runs_total{status}  Counter  — issue #486 spec: every accepted
//     POST /api/v1/ci-runs increments this Counter with the status label
//     (queued / running / success / failed / cancelled / skipped / unknown).
//
// devhub_ci_run_ingest_duration_seconds  Histogram — webhook → row insert latency.
//     Spec: webhook → row insert end-to-end. 본 sprint 의 handler 자체 latency
//     (row insert time) 가 metric 의 upper bound. webhook fan-out 은 별도
//     sprint 에서 구현 시 동일 Histogram 사용.
//
// 두 metric 모두 best-effort: prometheus.Register 실패 시 (이미 등록됨 등)
// 패닉 안 하고 무시 — handler 가 metric 때문에 깨지면 안 됨.

var (
	devhubCIRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "devhub_ci_runs_total",
			Help: "Total number of CI runs ingested via POST /api/v1/ci-runs, labeled by status.",
		},
		[]string{"status"},
	)

	devhubCIRunIngestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "devhub_ci_run_ingest_duration_seconds",
			Help:    "CI run ingest duration (handler time, seconds) for POST /api/v1/ci-runs.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	// prometheus.Register 는 중복 등록 시 panic. 동일 binary 에서 한 번만 호출.
	// best-effort: register 실패 시 (이미 등록됨) 무시.
	registerCollectorSafe(devhubCIRunsTotal)
	registerCollectorSafe(devhubCIRunIngestDuration)
}

func registerCollectorSafe(c prometheus.Collector) {
	defer func() { _ = recover() }()
	prometheus.MustRegister(c)
}
