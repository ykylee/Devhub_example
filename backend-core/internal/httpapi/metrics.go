package httpapi

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	buildRunsDuration *prometheus.HistogramVec
)

func initBuildRunsMetrics() {
	buildRunsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "devhub_repository_build_runs_query_duration_seconds",
			Help:    "Query duration of ListRepositoryBuildRuns by status filter.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status_filter"},
	)
	if err := prometheus.Register(buildRunsDuration); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

func observeBuildRunsQueryDuration(statusFilter string, duration time.Duration) {
	initBuildRunsMetrics()
	buildRunsDuration.WithLabelValues(statusFilter).Observe(duration.Seconds())
}
