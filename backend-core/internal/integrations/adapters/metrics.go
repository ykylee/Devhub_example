package adapters

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsOnce sync.Once

	homeLabPullRunsTotal *prometheus.CounterVec
	homeLabPullDuration  prometheus.Histogram
	homeLabServicesGauge prometheus.Gauge
	homeLabDegradedGauge prometheus.Gauge
	homeLabLastSuccess   prometheus.Gauge
)

func initHomeLabMetrics() {
	metricsOnce.Do(func() {
		homeLabPullRunsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_homelab_pull_runs_total",
				Help: "Total number of HomeLab pull runs by result.",
			},
			[]string{"result"},
		)
		homeLabPullDuration = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "devhub_homelab_pull_duration_seconds",
				Help:    "Duration of HomeLab pull-and-ingest runs.",
				Buckets: prometheus.DefBuckets,
			},
		)
		homeLabServicesGauge = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_homelab_snapshot_services",
				Help: "Number of services in the latest HomeLab snapshot.",
			},
		)
		homeLabDegradedGauge = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_homelab_degraded_providers",
				Help: "Number of degraded providers in the latest HomeLab snapshot.",
			},
		)
		homeLabLastSuccess = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_homelab_last_success_unixtime",
				Help: "Unix time of the latest successful HomeLab pull-and-ingest run.",
			},
		)
		scmCreateRunsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_scm_create_runs_total",
				Help: "Total number of SCM create runs by result.",
			},
			[]string{"result"},
		)
		scmCreateDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "devhub_scm_create_duration_seconds",
				Help:    "Duration of SCM create runs by result.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"result"},
		)
		scmCreateReposGaug = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_scm_create_repos_total",
				Help: "Number of repositories currently in success state.",
			},
		)
		scmCreateFailures = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_scm_create_failures_total",
				Help: "Total number of SCM create failures by provider.",
			},
			[]string{"scm_provider"},
		)
		registerCollector(homeLabPullRunsTotal)
		registerCollector(homeLabPullDuration)
		registerCollector(homeLabServicesGauge)
		registerCollector(homeLabDegradedGauge)
		registerCollector(homeLabLastSuccess)
		registerCollector(scmCreateRunsTotal)
		registerCollector(scmCreateDuration)
		registerCollector(scmCreateReposGaug)
		registerCollector(scmCreateFailures)
	})
}

func registerCollector(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
	}
}

func observeHomeLabPull(result string, duration time.Duration) {
	initHomeLabMetrics()
	homeLabPullRunsTotal.WithLabelValues(result).Inc()
	homeLabPullDuration.Observe(duration.Seconds())
}

func observeHomeLabSnapshot(snapshot HomeLabSnapshot) {
	initHomeLabMetrics()
	serviceCount := 0
	if len(snapshot.ServicesJSON) > 0 {
		var services []json.RawMessage
		if err := json.Unmarshal(snapshot.ServicesJSON, &services); err == nil {
			serviceCount = len(services)
		}
	}
	homeLabServicesGauge.Set(float64(serviceCount))
	homeLabDegradedGauge.Set(float64(len(snapshot.DegradedProviders)))
}

func observeHomeLabPullSuccess(now time.Time) {
	initHomeLabMetrics()
	homeLabLastSuccess.Set(float64(now.UTC().Unix()))
}

// X-5 Gitea pull metric helpers (ADR-0034 §3.5).
var (
	giteaPullRunsTotal        *prometheus.CounterVec
	giteaPullDuration         *prometheus.HistogramVec
	giteaPullRepositoriesGaug prometheus.Gauge
	giteaPullConsecFailures   *prometheus.GaugeVec
	giteaPullLastSuccess      prometheus.Gauge

	// X-4 SCM create metric (ADR-0035 §3.3)
	scmCreateRunsTotal *prometheus.CounterVec
	scmCreateDuration  *prometheus.HistogramVec
	scmCreateReposGaug prometheus.Gauge
	scmCreateFailures  *prometheus.CounterVec
)

func observeGiteaPull(result string, duration time.Duration) {
	initHomeLabMetrics()
	_ = result
	_ = duration
	// Placeholder: Gitea metric registration will be added by X-5 PR (#592) when merged.
}

func observeGiteaPullRepositoriesTotal(count int) {
	initHomeLabMetrics()
	_ = count
}

func observeGiteaPullAlertTriggered(repositoryID string, consecutiveFailures int) {
	initHomeLabMetrics()
	_ = repositoryID
	_ = consecutiveFailures
}

func observeGiteaPullLastSuccess(now time.Time) {
	initHomeLabMetrics()
	_ = now
}

// X-4 SCM create metric helpers (ADR-0035 §3.3).
// Init merged into initHomeLabMetrics (single sync.Once for thread-safety).

func observeSCMSuccess(duration time.Duration) {
	initHomeLabMetrics()
	scmCreateRunsTotal.WithLabelValues("success").Inc()
	scmCreateDuration.WithLabelValues("success").Observe(duration.Seconds())
	scmCreateReposGaug.Inc()
}

func observeSCMError(errorClass string, duration time.Duration) {
	initHomeLabMetrics()
	label := "unknown"
	switch errorClass {
	case "validation", "permission", "not_found", "rate_limit", "server", "network", "config":
		label = errorClass
	}
	scmCreateRunsTotal.WithLabelValues("error").Inc()
	scmCreateDuration.WithLabelValues("error").Observe(duration.Seconds())
	scmCreateFailures.WithLabelValues(label).Inc()
}
