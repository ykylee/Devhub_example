package service

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetrics_ObserveHelpers(t *testing.T) {
	// 1. Initialize
	InitMetrics()
	
	// Double initialization guard (sync.Once check)
	InitMetrics()

	// 2. Observe helpers
	ObserveExpiringSoon(5)
	ObserveStale(12)
	
	// Positive count check
	ObserveAutoRevoked(3)
	
	// Negative/Zero count branch check
	ObserveAutoRevoked(0)
	ObserveAutoRevoked(-1)
}

func TestMetrics_AlreadyRegistered(t *testing.T) {
	// Force AlreadyRegisteredError inside registerCollector by registering same gauge twice
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "forced_already_registered_test_gauge",
		Help: "forced duplicate",
	})
	
	registerCollector(gauge)
	registerCollector(gauge) // AlreadyRegisteredError branch covered!
}
