package service

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TestInitMetrics_Idempotent — InitMetrics 가 여러 번 호출돼도 register panic 없음.
// sync.Once + AlreadyRegisteredError 가드 둘 다 검증.
func TestInitMetrics_Idempotent(t *testing.T) {
	InitMetrics()
	InitMetrics() // second call no-op
	InitMetrics() // third call no-op
	// no panic → pass
}

// TestObserveEventProcessed_IncrementsCounter — Counter Vec 의 (kind, action) label
// 별 increment 검증.
func TestObserveEventProcessed_IncrementsCounter(t *testing.T) {
	InitMetrics()
	before := counterVecValue(t, eventsProcessedTotal, []string{"user", "auth.login.success"})
	ObserveEventProcessed("user", "auth.login.success")
	ObserveEventProcessed("user", "auth.login.success")
	after := counterVecValue(t, eventsProcessedTotal, []string{"user", "auth.login.success"})
	if delta := after - before; delta != 2 {
		t.Fatalf("counter delta = %v; want 2", delta)
	}
}

// TestObserveCursorLag_SetsGauge — GaugeVec 의 cursor_key 별 set 동작 검증.
func TestObserveCursorLag_SetsGauge(t *testing.T) {
	InitMetrics()
	ObserveCursorLag("keycloak.events", 12.5)
	got := gaugeVecValue(t, cursorLagSeconds, []string{"keycloak.events"})
	if got != 12.5 {
		t.Fatalf("gauge = %v; want 12.5", got)
	}
	// 두 번째 호출 — set 은 누적이 아니라 덮어쓰기.
	ObserveCursorLag("keycloak.events", 3.0)
	got = gaugeVecValue(t, cursorLagSeconds, []string{"keycloak.events"})
	if got != 3.0 {
		t.Fatalf("gauge = %v; want 3.0 (overwrite, not add)", got)
	}
}

// TestObserveEventProcessed_UnknownActionNormalized — unknown fallback action 의
// metric label 이 "keycloak.event.unknown" 단일 값으로 normalize 되어 cardinality
// unbounded 폭증 회피 (Stage 3 self-review 보강 — P2 #1).
// 서로 다른 suffix 의 unknown action 2건이 single counter 로 collapse 되는지 검증.
func TestObserveEventProcessed_UnknownActionNormalized(t *testing.T) {
	InitMetrics()
	before := counterVecValue(t, eventsProcessedTotal, []string{"user", "keycloak.event.unknown"})
	ObserveEventProcessed("user", "keycloak.event.unknown:FOO_BAR_BAZ")
	ObserveEventProcessed("user", "keycloak.event.unknown:DIFFERENT_ONE")
	after := counterVecValue(t, eventsProcessedTotal, []string{"user", "keycloak.event.unknown"})
	if delta := after - before; delta != 2 {
		t.Fatalf("unified unknown counter delta = %v; want 2 (both unknown:* must collapse to single label)", delta)
	}
}

// TestNormalizeMetricAction — bounded label normalize 규칙 검증.
func TestNormalizeMetricAction(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"auth.login.success", "auth.login.success"},
		{"keycloak.user.created", "keycloak.user.created"},
		{"keycloak.event.unknown:FOO", "keycloak.event.unknown"},
		{"keycloak.event.unknown:BAR_BAZ", "keycloak.event.unknown"},
		// admin fallback 은 bounded 매핑 표의 ResourceType 한정이라 cardinality 제한 — 그대로 둠.
		{"keycloak.admin.group:create", "keycloak.admin.group:create"},
	}
	for _, tc := range cases {
		got := normalizeMetricAction(tc.in)
		if got != tc.want {
			t.Errorf("normalizeMetricAction(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

// TestObservePullError_IncrementsCounter — pull error counter 검증.
func TestObservePullError_IncrementsCounter(t *testing.T) {
	InitMetrics()
	before := counterVecValue(t, pullErrorsTotal, []string{"user"})
	ObservePullError("user")
	ObservePullError("admin")
	ObservePullError("user")
	afterUser := counterVecValue(t, pullErrorsTotal, []string{"user"})
	afterAdmin := counterVecValue(t, pullErrorsTotal, []string{"admin"})
	if delta := afterUser - before; delta != 2 {
		t.Fatalf("user delta = %v; want 2", delta)
	}
	if afterAdmin < 1 {
		t.Fatalf("admin = %v; want >= 1", afterAdmin)
	}
}

// counterVecValue 는 prometheus CounterVec 의 특정 label 조합 값을 dto 로 읽는다.
// 다른 테스트가 같은 label 을 건드릴 수 있으므로 절대값이 아닌 delta 검증에만 사용.
func counterVecValue(t *testing.T, c *prometheus.CounterVec, labels []string) float64 {
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

func gaugeVecValue(t *testing.T, g *prometheus.GaugeVec, labels []string) float64 {
	t.Helper()
	m, err := g.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues: %v", err)
	}
	pb := &dto.Metric{}
	if err := m.Write(pb); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if pb.Gauge == nil {
		return 0
	}
	return pb.Gauge.GetValue()
}
