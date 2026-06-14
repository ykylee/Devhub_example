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

	giteaPullRunsTotal        *prometheus.CounterVec
	giteaPullDuration         *prometheus.HistogramVec
	giteaPullRepositoriesGaug prometheus.Gauge
	giteaPullConsecFailures   *prometheus.GaugeVec
	giteaPullLastSuccess      prometheus.Gauge
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
		giteaPullRunsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "devhub_gitea_pull_runs_total",
				Help: "Total number of Gitea pull cycles by result.",
			},
			[]string{"result"},
		)
		giteaPullDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "devhub_gitea_pull_duration_seconds",
				Help:    "Duration of Gitea pull cycles by result.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"result"},
		)
		giteaPullRepositoriesGaug = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_gitea_pull_repositories_total",
				Help: "Number of repositories due in the current Gitea pull cycle.",
			},
		)
		giteaPullConsecFailures = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "devhub_gitea_pull_consecutive_failures",
				Help: "Consecutive failure count per repository (alert trigger when >= threshold).",
			},
			[]string{"repository_id"},
		)
		giteaPullLastSuccess = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "devhub_gitea_pull_last_success_unixtime",
				Help: "Unix time of the latest successful Gitea pull cycle.",
			},
		)
		registerCollector(homeLabPullRunsTotal)
		registerCollector(homeLabPullDuration)
		registerCollector(homeLabServicesGauge)
		registerCollector(homeLabDegradedGauge)
		registerCollector(homeLabLastSuccess)
		registerCollector(giteaPullRunsTotal)
		registerCollector(giteaPullDuration)
		registerCollector(giteaPullRepositoriesGaug)
		registerCollector(giteaPullConsecFailures)
		
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
				Help: "Total number of SCM create failures, partitioned by error class and SCM provider.",
			},
			[]string{"error_class", "scm_provider"},
		)
	registerCollector(giteaPullLastSuccess)

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

// Gitea pull metric helpers (X-5 ADR-0034 §1.4).

func observeGiteaPull(result string, duration time.Duration) {
	initHomeLabMetrics()
	giteaPullRunsTotal.WithLabelValues(result).Inc()
	giteaPullDuration.WithLabelValues(result).Observe(duration.Seconds())
}

func observeGiteaPullRepositoriesTotal(count int) {
	initHomeLabMetrics()
	giteaPullRepositoriesGaug.Set(float64(count))
}

func observeGiteaPullAlertTriggered(repositoryID string, consecutiveFailures int) {
	initHomeLabMetrics()
	giteaPullConsecFailures.WithLabelValues(repositoryID).Set(float64(consecutiveFailures))
}

func observeGiteaPullLastSuccess(now time.Time) {
	initHomeLabMetrics()
	giteaPullLastSuccess.Set(float64(now.UTC().Unix()))
}

// X-4 SCM create metric helpers (ADR-0035 §3.3).
var (
	scmCreateRunsTotal *prometheus.CounterVec
	scmCreateDuration  *prometheus.HistogramVec
	scmCreateReposGaug prometheus.Gauge
	scmCreateFailures  *prometheus.CounterVec
)

func observeSCMSuccess(duration time.Duration) {
	initHomeLabMetrics()
	scmCreateRunsTotal.WithLabelValues("success").Inc()
	scmCreateDuration.WithLabelValues("success").Observe(duration.Seconds())
	scmCreateReposGaug.Inc()
}

func observeSCMError(errorClass string, scmProvider string, duration time.Duration) {
	initHomeLabMetrics()
	// errorClass label: validation | permission | not_found | rate_limit | server | network | config | unknown
	classLabel := "unknown"
	switch errorClass {
	case "validation", "permission", "not_found", "rate_limit", "server", "network", "config":
		classLabel = errorClass
	}
	// scm_provider label: gitea | github | gitlab | unknown — sourced from SCMCreateRequest.SCMProvider
	// so dashboards/alerts grouping by provider see real provider values, not error classes.
	// codex review #6 (P2).
	providerLabel := "unknown"
	if scmProvider != "" {
		providerLabel = scmProvider
	}
	scmCreateRunsTotal.WithLabelValues("error").Inc()
	scmCreateDuration.WithLabelValues("error").Observe(duration.Seconds())
	scmCreateFailures.WithLabelValues(classLabel, providerLabel).Inc()
}
