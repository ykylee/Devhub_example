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
		registerCollector(homeLabPullRunsTotal)
		registerCollector(homeLabPullDuration)
		registerCollector(homeLabServicesGauge)
		registerCollector(homeLabDegradedGauge)
		registerCollector(homeLabLastSuccess)
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
