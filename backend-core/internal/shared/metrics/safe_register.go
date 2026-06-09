package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterCollectorSafe 는 prometheus.MustRegister 를 best-effort 로 호출.
// 중복 등록 시 panic 하지만 recover() 로 무시. handler 가 metric 때문에
// 깨지면 안 됨.
//
// internal/shared/metrics/ 위치 — domain view + httpapi + integrations 양쪽
// 에서 import 가능. (ADR-0029 §6 (g) — api_key.go 신규와 정합)
func RegisterCollectorSafe(c prometheus.Collector) {
	defer func() { _ = recover() }()
	prometheus.MustRegister(c)
}
