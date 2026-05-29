# ADR-0016: Prometheus 알림 규칙 정책 (HomeLab observability)

- 문서 목적: HomeLab pull/push 경로의 Prometheus 알림 규칙과 임계값 정책 결정. PR #139 (`codex/next-step-20260516`) 으로 backend `/metrics` 엔드포인트 + 5 핵심 메트릭이 활성화된 사실의 사후 명문화 + 알림 규칙 표준화.
- 범위: Alertmanager 규칙 (alert name, expr, for, severity) + stage/prod 환경별 임계값 분리 + Dashboard panel 권장. [ADR-0015 HomeLab adapter pull strategy](./0015-homelab-adapter-pull-strategy.md) 의 메트릭 명세 위에 부착.
- 대상 독자: 운영자 (SRE), Backend 개발자, On-call.
- 상태: accepted
- 작성일: 2026-05-18
- 결정일: 2026-05-17 (실 초안), 2026-05-18 사후 명문화 (sprint `claude/work_260518-c`)
- 결정 근거 sprint: `codex/next-step-20260516` (PR #139, sha `e2a76fb`).
- 관련 문서: [ADR-0015 HomeLab pull strategy](./0015-homelab-adapter-pull-strategy.md), [`docs/domain/integration-registry/prometheus_homelab_alerts.md`](../domain/integration-registry/prometheus_homelab_alerts.md) (1차 초안 — 본 ADR 이 source-of-truth 로 승격), [추적성 매트릭스 §3 External Integration 행 + §4 ADR](../traceability/report.md).

## 1. 컨텍스트

ADR-0015 가 명시한 5 핵심 메트릭은 backend-core `/metrics` 엔드포인트에서 Prometheus scrape 됨:

- `devhub_homelab_pull_runs_total{result="success|error"}` (counter)
- `devhub_homelab_pull_duration_seconds` (histogram)
- `devhub_homelab_snapshot_services` (gauge)
- `devhub_homelab_degraded_providers` (gauge)
- `devhub_homelab_last_success_unixtime` (gauge)

본 ADR 은 그 위에 **Alertmanager 규칙 3종** 과 **환경별 임계값** 을 결정한다.

## 2. 결정 동인

- **false positive 최소화**: pull 1회 실패는 network jitter 가능성이 높음. 연속/누적 신호로만 알림.
- **장애 조기 감지**: `last_success` staleness 가 길어지면 push 경로마저 끊긴 상태 — production 에서는 15분 임계가 한계.
- **stage 1주 관찰 후 prod 임계값 확정**: 운영 데이터 없는 초기 임계는 conservative + 1주 후 튜닝.
- **degraded provider 신호 분리**: pull 자체는 성공하지만 snapshot 내 provider 가 degraded — 운영자가 별도 인지해야 하는 시나리오.
- **multi-instance 운영 정합**: `last_success` 기반 알림은 backend-core 가 다중 인스턴스 배포 시 `max by(instance)` aggregation 필요. 본 ADR 은 단일 인스턴스 baseline 임계만 결정.

## 3. 검토 옵션

### 3.1 알림 트리거 신호

| 옵션 | 채택 |
| --- | --- |
| pull 1회 실패 → 즉시 알림 | ❌ (false positive 과다) |
| **누적 error N회 + duration window** | ✅ |
| **last_success staleness** | ✅ |
| **degraded provider count > 0** | ✅ |
| pull latency p95 > threshold | ❌ (1차 carve out — baseline 데이터 부재) |

### 3.2 환경별 임계값 분리

| 옵션 | 채택 |
| --- | --- |
| 단일 임계 (stage = prod) | ❌ (stage 의 reliability 가 prod 보다 낮음) |
| **stage / prod 임계 분리** | ✅ (stage 더 너그러운 임계, prod 더 엄격) |
| 환경 외에 시간대별 (peak/off-peak) | ❌ (1차는 시간대 무관) |

### 3.3 dashboard 노출 panel

5 panel 채택 (decision §4.3) — 모든 panel 은 PromQL 기반, Grafana JSON 모델은 본 ADR 외부 (carve out).

## 4. 결정

### 4.1 알림 규칙 3종

`devhub-homelab` Prometheus group 의 3 alert. 본 ADR 의 baseline (production 임계) 은 다음 표:

| Alert | expr | for | severity | 의미 |
| --- | --- | --- | --- | --- |
| `DevhubHomeLabPullFailing` | `increase(devhub_homelab_pull_runs_total{result="error"}[10m]) >= 3` | 5m | warning | 10분 윈도우 내 pull error 3회 이상 |
| `DevhubHomeLabPullNoRecentSuccess` | `(time() - devhub_homelab_last_success_unixtime) > 900` | 5m | critical | 최근 15분 내 pull 성공 기록 없음 |
| `DevhubHomeLabDegradedProvidersDetected` | `devhub_homelab_degraded_providers > 0` | 10m | warning | snapshot 내 degraded provider 가 10분 이상 유지 |

규칙 raw YAML 은 [`docs/domain/integration-registry/prometheus_homelab_alerts.md`](../domain/integration-registry/prometheus_homelab_alerts.md) §2 (planning doc) 또는 본 ADR §6 의 carve out 항목 참조.

### 4.2 환경별 임계값

stage 환경은 reliability 가 낮으므로 더 너그러운 임계 적용. stage 1주 관찰 후 prod 임계 확정 — 본 ADR 의 baseline 은 production 기준.

| Alert | stage | prod |
| --- | --- | --- |
| `DevhubHomeLabPullFailing` | `increase(error[15m]) >= 5`, `for: 10m` | `increase(error[10m]) >= 3`, `for: 5m` |
| `DevhubHomeLabPullNoRecentSuccess` | `> 1800` (30분), `for: 10m` | `> 900` (15분), `for: 5m` |
| `DevhubHomeLabDegradedProvidersDetected` | `> 0`, `for: 15m` | `> 0`, `for: 10m` |

label `environment: stage|prod` 로 Alertmanager 라우팅 분리.

### 4.3 Dashboard panel 5종 (Grafana 권장)

1. **Pull Success/Error (10m)** — `increase(devhub_homelab_pull_runs_total{result="success"}[10m])` + `increase(...{result="error"}[10m])`.
2. **Pull Latency p95** — `histogram_quantile(0.95, sum(rate(devhub_homelab_pull_duration_seconds_bucket[5m])) by (le))`.
3. **Last Success Age (sec)** — `time() - devhub_homelab_last_success_unixtime`.
4. **Degraded Providers** — `devhub_homelab_degraded_providers`.
5. **Observed Services** — `devhub_homelab_snapshot_services`.

### 4.4 운영 정책

- **stage 1주 관찰** — stage 환경 알림이 1주간 false positive ratio < 10% 유지되면 prod 임계 적용. 초과 시 임계값 (count / window / for) 1단계 완화.
- **multi-instance aggregation** — backend-core 2개 이상 배포 시 `last_success` 는 `max by(provider) (devhub_homelab_last_success_unixtime)` 로 집계, instance 별 metric 은 dashboard 에서만 노출.
- **file vs HTTP exclusive** — 한 인스턴스에서 두 source 동시 활성화 금지 (ADR-0015 §4.1). 알림은 mode 무관 동일 임계 적용.
- **튜닝 체크리스트** — (a) error/success ratio `increase(error[30m]) / clamp_min(increase(success[30m]), 1)` 추이, (b) pull latency p95 baseline 대비 2배 이상 증가 시 (1차 carve out 의 latency alert 도입 trigger), (c) `NoRecentSuccess` 발생 시점의 upstream 상태 코드 분포 (429 / 5xx).

## 5. 결과

본 ADR 은 운영 정책 결정이므로 코드 변경 없음. backend 측 구현은 [ADR-0015](./0015-homelab-adapter-pull-strategy.md) 의 §5 가 다룸.

- Alertmanager 규칙 YAML — `docs/domain/integration-registry/prometheus_homelab_alerts.md` §2~§2.2 의 stage/prod baseline 이 본 ADR 의 결정과 정합. **운영 자산 setup 가이드는 [`docs/setup/prometheus_alertmanager_setup.md`](../setup/prometheus_alertmanager_setup.md) — 외부 git 이관 layout + raw YAML reference + Alertmanager 라우팅 + 운영 SOP** (sprint `claude/work_260518-s`, §6 carve (1) resolved).
- Dashboard JSON 모델 — **[`docs/setup/grafana/homelab_dashboard.json`](../setup/grafana/homelab_dashboard.json) 신규** (sprint `claude/work_260518-s`, §6 carve (2) resolved). Grafana schemaVersion 38 + 5 panel + Prometheus datasource + environment template variable. 운영팀이 Grafana 에 import 후 ops 저장소의 사본으로 deploy.

## 6. 후속 작업

- **✅ resolved (sprint `claude/work_260518-s`, 본 PR)** Alertmanager 규칙 raw YAML 외부 git 이관 + 운영 SOP — [`docs/setup/prometheus_alertmanager_setup.md`](../setup/prometheus_alertmanager_setup.md) 신규. 외부 git layout (ops-monitoring 저장소) + Prometheus scrape + stage/prod raw YAML reference + Alertmanager 라우팅 (Slack/PagerDuty receiver) + 운영 SOP (rule 추가/변경/검증/배포 + 신규 알림 진단 + 임계 튜닝 + multi-instance aggregation).
- **✅ resolved (sprint `claude/work_260518-s`, 본 PR)** Grafana dashboard JSON 모델 — [`docs/setup/grafana/homelab_dashboard.json`](../setup/grafana/homelab_dashboard.json) 신규. schemaVersion 38 + 5 panel (§4.3 PromQL 정합) + Prometheus datasource + environment template variable + ADR-0015/0016/setup 가이드 link.
- **(carve)** pull latency p95 alert — baseline 1주 관찰 후 임계 결정 + 본 ADR 후속 갱신. Grafana JSON 의 Panel 2 (Pull Latency p95) 가 baseline 관찰용으로 이미 active — 임계 결정만 carve out.
- **(carve)** push 경로 (`API-73`) 의 알림 — webhook 수신 실패율 metric 도입 후 별도 ADR.
- **(carve)** stage → prod 임계값 확정 — 1주 관찰 후 본 §7 변경 이력에 row 추가. 본 가이드 §6.3 의 튜닝 체크리스트 따라 진행.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-17 | 1차 초안 (`docs/domain/integration-registry/prometheus_homelab_alerts.md`) — 3 alert + stage/prod 임계값 분리. | PR #139 활성화 시점 |
| 2026-05-18 | accepted — ADR 형식으로 사후 명문화. baseline = production 임계, stage 1주 관찰 후 prod 확정. multi-instance aggregation = `max by(provider)`. latency p95 alert 는 carve out. | sprint `claude/work_260518-c` |
| 2026-05-18 | §6 carve out (1)+(2) **resolved** — (1) Alertmanager 운영 자산 setup 가이드 `docs/setup/prometheus_alertmanager_setup.md` 신규 (외부 git 이관 layout + raw YAML reference + 라우팅 예시 + 운영 SOP). (2) Grafana dashboard JSON 모델 `docs/setup/grafana/homelab_dashboard.json` 신규 (5 panel + Prometheus datasource + environment template variable). §5 결과에 두 신규 자산 link 추가. (3)/(4)/(5) carve out 유지 — baseline 1주 관찰 / push 경로 metric / stage→prod 확정 모두 사전 조건 부재. | sprint `claude/work_260518-s` (PR #160) |
| 2026-05-18 | codex hotfix #8 P1 #2 + P2 #6 — `docs/setup/prometheus_alertmanager_setup.md` §4 prod/stage rule expr 에 `{environment="prod\|stage"}` matcher 명시 (P1: 단일 Prometheus 가 stage+prod scrape 시 cross-env contamination 회피 — 이전 expr 은 metric 양쪽 평가 후 label 만 rewrite → 잘못된 receiver routing) + `DevhubHomeLabPullNoRecentSuccess` 와 `DegradedProvidersDetected` 의 expr 를 `max by(provider)` aggregation 으로 (P2: multi-instance 시 instance 별 series 가 single-instance staleness 로도 알림하는 회귀 회피 — §6.4 의 운영 지침 정합). §8 DREQ rule 도 동일 패턴 (`environment` matcher + `max`/`sum` aggregation). | sprint `claude/work_260518-w` (본 PR) |
