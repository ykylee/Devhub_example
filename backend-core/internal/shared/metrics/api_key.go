package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// API key authentication metrics — ADR-0029 §6 (g) P2 + SOP
// [`docs/setup/api_key_rotation.md` §6.1].
//
// 3 metric:
//   - DevhubAPIKeyAuthTotal{result}  Counter
//       result = "success" | "denied"
//       success: API key 인증 통과 (auth.api_key_authenticated emit 매칭)
//       denied: API key 인증 거부 (auth.api_key_denied emit 매칭,
//               §6 (a) admin gate 거부 OR invalid key 양쪽 포함)
//   - DevhubAPIKeyActorTotal  Gauge
//       unique actor count. 정상 = 1 (api-key 단일). SOP §6.1 alert
//       `DevhubAPIKeyActorTotal > 5` 와 정합 (multi-key 의도 외).
//   - DevhubAPIKeyRotationDueAt  Gauge
//       다음 rotation due date 까지 unix timestamp. SOP §6.2 alert D-7
//       post 정합. 운영자 수동 set (rotation SOP §4.2 step 6 와 정합) —
//       본 sprint 에서는 default 0 (미설정) + cron refresh carve.
//
// 모두 best-effort: prometheus.Register 실패 시 (이미 등록됨 등) 패닉 안
// 하고 무시 — auth middleware 가 metric 때문에 깨지면 안 됨.
//
// internal/shared/metrics/ 위치 — domain view package (rbac-permissions/view) 와
// httpapi package 양쪽에서 import 가능. architecture.md §taxonomy 의 shared
// layer 정합 (config, httphelp, integrationcaps 와 동급).

var (
	DevhubAPIKeyAuthTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "devhub_api_key_auth_total",
			Help: "Total number of API key authentication attempts, labeled by result (success|denied).",
		},
		[]string{"result"},
	)

	DevhubAPIKeyActorTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "devhub_api_key_actor_total",
			Help: "Unique actor count for API key authenticated sessions. 정상 = 1 (api-key 단일). > 5 = multi-key 의도 외 / leak 의심 (SOP §6.2).",
		},
	)

	DevhubAPIKeyRotationDueAt = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "devhub_api_key_rotation_due_at_seconds",
			Help: "Next API key rotation due date as Unix timestamp. 운영자 수동 set (rotation SOP §4.2 step 6). 0 = 미설정. SOP §6.2 alert D-7 post 정합.",
		},
	)
)

func init() {
	RegisterCollectorSafe(DevhubAPIKeyAuthTotal)
	RegisterCollectorSafe(DevhubAPIKeyActorTotal)
	RegisterCollectorSafe(DevhubAPIKeyRotationDueAt)
}
