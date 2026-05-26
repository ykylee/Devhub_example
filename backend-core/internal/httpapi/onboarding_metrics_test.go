package httpapi

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TestInitOnboardingMetrics_Idempotent — initOnboardingMetrics 여러 번 호출해도
// register panic 없음 (sync.Once + AlreadyRegisteredError 가드).
func TestInitOnboardingMetrics_Idempotent(t *testing.T) {
	initOnboardingMetrics()
	initOnboardingMetrics()
	initOnboardingMetrics()
	// no panic → pass
}

// TestObserveOnboardingGateBlocked_IncrementsCounter — gate 차단 counter increment.
func TestObserveOnboardingGateBlocked_IncrementsCounter(t *testing.T) {
	initOnboardingMetrics()
	before := counterVecValueOnboarding(t, onboardingGateBlockedTotal, []string{"onboarding_required"})
	observeOnboardingGateBlocked("onboarding_required")
	observeOnboardingGateBlocked("onboarding_required")
	after := counterVecValueOnboarding(t, onboardingGateBlockedTotal, []string{"onboarding_required"})
	if delta := after - before; delta != 2 {
		t.Fatalf("gate_blocked delta = %v; want 2", delta)
	}
}

// TestObserveOnboardingSubmit_AllStatuses — 7 status label 각각 increment 분리 검증.
func TestObserveOnboardingSubmit_AllStatuses(t *testing.T) {
	initOnboardingMetrics()
	statuses := []string{"ok", "rejected", "conflict", "not_found", "server_error", "unavailable", "unauthenticated"}
	before := make(map[string]float64, len(statuses))
	for _, s := range statuses {
		before[s] = counterVecValueOnboarding(t, onboardingSubmitTotal, []string{s})
		observeOnboardingSubmit(s)
	}
	for _, s := range statuses {
		after := counterVecValueOnboarding(t, onboardingSubmitTotal, []string{s})
		if delta := after - before[s]; delta != 1 {
			t.Errorf("submit{status=%q} delta = %v; want 1", s, delta)
		}
	}
}

// TestObserveOnboardingSubmitDuration_RecordsHistogram — Histogram bucket count 증가 검증.
func TestObserveOnboardingSubmitDuration_RecordsHistogram(t *testing.T) {
	initOnboardingMetrics()
	before := histogramSampleCount(t, onboardingSubmitDurationSec)
	observeOnboardingSubmitDuration(0.05)  // 50ms
	observeOnboardingSubmitDuration(0.5)   // 500ms
	observeOnboardingSubmitDuration(2.0)   // 2s
	after := histogramSampleCount(t, onboardingSubmitDurationSec)
	if delta := after - before; delta != 3 {
		t.Fatalf("histogram sample count delta = %v; want 3", delta)
	}
}

// TestObserveOnboardingReviewConfirm_AllStatuses — confirm handler 7 status 분리 검증.
func TestObserveOnboardingReviewConfirm_AllStatuses(t *testing.T) {
	initOnboardingMetrics()
	statuses := []string{"ok", "rejected", "conflict", "not_found", "server_error", "unavailable", "bad_request"}
	before := make(map[string]float64, len(statuses))
	for _, s := range statuses {
		before[s] = counterVecValueOnboarding(t, onboardingReviewConfirmTotal, []string{s})
		observeOnboardingReviewConfirm(s)
	}
	for _, s := range statuses {
		after := counterVecValueOnboarding(t, onboardingReviewConfirmTotal, []string{s})
		if delta := after - before[s]; delta != 1 {
			t.Errorf("review_confirm{status=%q} delta = %v; want 1", s, delta)
		}
	}
}

// counterVecValueOnboarding — audit/metrics_test.go 의 counterVecValue 와 동일
// pattern (별도 package 라 helper 중복).
func counterVecValueOnboarding(t *testing.T, c *prometheus.CounterVec, labels []string) float64 {
	t.Helper()
	m, err := c.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues: %v", err)
	}
	pb := &dto.Metric{}
	if err := m.Write(pb); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if pb.Counter == nil {
		return 0
	}
	return pb.Counter.GetValue()
}

// histogramSampleCount — Histogram 의 누적 sample count 반환. observe N 회 시
// count 가 N 증가.
func histogramSampleCount(t *testing.T, h prometheus.Histogram) uint64 {
	t.Helper()
	pb := &dto.Metric{}
	if err := h.Write(pb); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if pb.Histogram == nil {
		return 0
	}
	return pb.Histogram.GetSampleCount()
}
