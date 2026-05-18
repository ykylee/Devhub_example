# Prometheus + Alertmanager 운영 setup (HomeLab observability)

- 문서 목적: [ADR-0016 Prometheus alerts policy](../adr/0016-prometheus-alerts-policy.md) §6 carve out (1) 의 운영 자산 — Alertmanager 규칙 raw YAML 의 외부 git 이관 layout + 운영 SOP + raw YAML reference.
- 범위: Prometheus scrape (backend-core `/metrics`) + Alertmanager 규칙 배포 + 라우팅. Grafana dashboard JSON 모델은 [`grafana/homelab_dashboard.json`](./grafana/homelab_dashboard.json) 별도.
- 대상 독자: 운영자 (SRE), On-call.
- 상태: draft (1차)
- 최종 수정일: 2026-05-18
- 관련 문서: [ADR-0015 HomeLab pull](../adr/0015-homelab-adapter-pull-strategy.md), [ADR-0016 alerts policy](../adr/0016-prometheus-alerts-policy.md), [HomeLab agent token rotation SOP](./homelab_agent_token_rotation.md), [`docs/planning/prometheus_homelab_alerts.md`](../planning/prometheus_homelab_alerts.md) (planning 1차 초안).

## 1. 배경 + 책임 분리

DevHub backend-core 는 `/metrics` 엔드포인트에 5 핵심 HomeLab 메트릭을 노출 ([ADR-0015 §4.5](../adr/0015-homelab-adapter-pull-strategy.md#45-메트릭-노출-prometheus-정합)). Prometheus 가 scrape, Alertmanager 가 규칙을 평가해 알림 전송. 본 가이드는 **운영 자산 (Prometheus + Alertmanager 설정)** 의 git 이관 layout 과 SOP 결정.

- **DevHub 저장소 (`devhub`)**: 의도 + 임계 source-of-truth (본 ADR-0016 + 본 가이드 + raw YAML reference).
- **운영 자산 저장소 (별도)**: 실제 Prometheus `prometheus.yml` + Alertmanager `alertmanager.yml` + alert rules 가 deploy 되는 곳. **devhub git 외부**, vault 또는 사내 ops git.

CLAUDE.md 의 docker = env-specific 정책 정합 — 환경 특화 자산 (Prometheus host 주소, Slack/Email webhook secret, runbook URL 등) 은 devhub 저장소 외부.

## 2. 외부 git 이관 권장 layout

운영 자산 저장소 layout 예시:

```
ops-monitoring/                      # 별도 git repo (vault 또는 internal git)
├── prometheus/
│   ├── prometheus.yml               # scrape config (DevHub backend-core 포함)
│   └── rules/
│       ├── devhub-homelab-stage.yml # stage rules (ADR-0016 §4.2)
│       └── devhub-homelab-prod.yml  # prod rules (ADR-0016 §4.2)
├── alertmanager/
│   └── alertmanager.yml             # 라우팅 + receiver (Slack/Email/PagerDuty)
└── grafana/
    └── dashboards/
        └── homelab.json             # devhub/docs/setup/grafana/homelab_dashboard.json 의 import 사본
```

devhub 저장소의 본 가이드 + ADR-0016 + `docs/setup/grafana/homelab_dashboard.json` 가 정 (source-of-truth), 운영 자산 저장소는 deploy 대상. 두 저장소 간 sync 는 운영자 책임 (PR template 의 추적성 영향 섹션 참조).

## 3. Prometheus scrape 설정 (reference)

`ops-monitoring/prometheus/prometheus.yml` 의 scrape job 예시:

```yaml
scrape_configs:
  - job_name: 'devhub-backend-core'
    scrape_interval: 30s
    scrape_timeout: 10s
    static_configs:
      - targets:
          - 'devhub-backend-1.internal:8080'  # 운영 환경별 호스트
          - 'devhub-backend-2.internal:8080'  # 다중 인스턴스 시
        labels:
          environment: prod
          service: devhub-backend-core
    metrics_path: '/metrics'
```

stage 환경은 `environment: stage` label + 별도 target.

## 4. Alertmanager 규칙 raw YAML (reference)

ADR-0016 §4.2 의 baseline. **`ops-monitoring/prometheus/rules/devhub-homelab-prod.yml`** 사본:

```yaml
groups:
  - name: devhub-homelab-prod
    rules:
      - alert: DevhubHomeLabPullFailing
        expr: increase(devhub_homelab_pull_runs_total{result="error"}[10m]) >= 3
        for: 5m
        labels:
          severity: warning
          environment: prod
        annotations:
          summary: "HomeLab pull 반복 실패 (prod)"
          description: "10분 윈도우 내 pull error가 3회 이상 발생했습니다. backend-core 로그 + upstream 상태코드 분포 확인."
          runbook_url: "https://internal.example.com/runbooks/devhub-homelab-pull-failing"

      - alert: DevhubHomeLabPullNoRecentSuccess
        expr: (time() - devhub_homelab_last_success_unixtime) > 900
        for: 5m
        labels:
          severity: critical
          environment: prod
        annotations:
          summary: "HomeLab pull 성공 지연 (prod, 15m+)"
          description: "최근 15분 이상 pull 성공 기록이 없습니다. push 경로도 끊겼는지 확인."
          runbook_url: "https://internal.example.com/runbooks/devhub-homelab-no-recent-success"

      - alert: DevhubHomeLabDegradedProvidersDetected
        expr: devhub_homelab_degraded_providers > 0
        for: 10m
        labels:
          severity: warning
          environment: prod
        annotations:
          summary: "Degraded provider 감지 (prod)"
          description: "HomeLab snapshot 에 degraded provider 가 10분 이상 존재합니다. /admin/topology-v2 의 banner 확인."
```

**`ops-monitoring/prometheus/rules/devhub-homelab-stage.yml`** 사본 (ADR-0016 §4.2 stage 임계):

```yaml
groups:
  - name: devhub-homelab-stage
    rules:
      - alert: DevhubHomeLabPullFailing
        expr: increase(devhub_homelab_pull_runs_total{result="error"}[15m]) >= 5
        for: 10m
        labels:
          severity: warning
          environment: stage
        annotations:
          summary: "HomeLab pull 반복 실패 (stage)"
          description: "stage: 15분 윈도우 내 pull error 5회 이상. 1주 baseline 관찰 중."

      - alert: DevhubHomeLabPullNoRecentSuccess
        expr: (time() - devhub_homelab_last_success_unixtime) > 1800
        for: 10m
        labels:
          severity: critical
          environment: stage
        annotations:
          summary: "HomeLab pull 성공 지연 (stage, 30m+)"
          description: "stage: 최근 30분 이상 pull 성공 없음."

      - alert: DevhubHomeLabDegradedProvidersDetected
        expr: devhub_homelab_degraded_providers > 0
        for: 15m
        labels:
          severity: warning
          environment: stage
        annotations:
          summary: "Degraded provider 감지 (stage)"
          description: "stage: degraded provider 가 15분 이상 존재."
```

## 5. Alertmanager 라우팅 (reference)

`ops-monitoring/alertmanager/alertmanager.yml` 예시:

```yaml
route:
  receiver: 'default-slack'
  group_by: ['alertname', 'environment']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - matchers:
        - environment="prod"
        - severity="critical"
      receiver: 'pagerduty-oncall'
      continue: true
    - matchers:
        - environment="prod"
      receiver: 'devhub-ops-slack'
    - matchers:
        - environment="stage"
      receiver: 'devhub-stage-slack'
      group_interval: 15m
      repeat_interval: 12h

receivers:
  - name: 'default-slack'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_DEFAULT}'
        channel: '#devhub-alerts'
        send_resolved: true

  - name: 'devhub-ops-slack'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_OPS}'
        channel: '#devhub-prod-alerts'
        send_resolved: true
        title: '{{ .GroupLabels.alertname }} ({{ .GroupLabels.environment }})'
        text: |
          {{ range .Alerts }}*{{ .Labels.severity | toUpper }}* — {{ .Annotations.summary }}
          {{ .Annotations.description }}
          Runbook: {{ .Annotations.runbook_url }}
          {{ end }}

  - name: 'devhub-stage-slack'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_STAGE}'
        channel: '#devhub-stage-alerts'
        send_resolved: true

  - name: 'pagerduty-oncall'
    pagerduty_configs:
      - routing_key: '${PAGERDUTY_ROUTING_KEY}'
        severity: '{{ .CommonLabels.severity }}'
```

- `${SLACK_WEBHOOK_*}`, `${PAGERDUTY_ROUTING_KEY}` 는 vault 또는 환경 변수로 주입 — git 추적 금지.
- stage 알림은 group_interval 15m + repeat_interval 12h 로 노이즈 감소.

## 6. 운영 SOP

### 6.1 규칙 추가 / 변경

1. **devhub 저장소** — ADR-0016 §4.X 또는 본 가이드 본문 갱신 (의도 + 임계 source-of-truth).
2. **운영 자산 저장소** — `ops-monitoring/prometheus/rules/devhub-homelab-*.yml` 갱신 + PR.
3. **검증** — `promtool check rules` 로 syntax 검증 + `promtool test rules` 로 unit test (옵션).
4. **배포** — Prometheus reload (`SIGHUP` 또는 `POST /-/reload`).
5. **검증** — Alertmanager UI / Prometheus `/alerts` 에서 새 규칙 평가 확인.

### 6.2 신규 알림 발생 시

1. Alertmanager 알림 채널 (Slack/PagerDuty) 에서 발생 알림 확인.
2. **annotation 의 `runbook_url`** 따라 진단.
3. backend-core `/metrics` 직접 조회 + `/admin/topology-v2` 의 degraded banner 확인.
4. upstream HomeLab agent 의 health 확인 (token 만료 여부 → [agent token rotation SOP](./homelab_agent_token_rotation.md)).
5. 임시 진정화 필요 시 Alertmanager UI 의 silence 기능 (시간 제한 명시).

### 6.3 임계 튜닝 (stage → prod)

ADR-0016 §4.4 의 운영 정책 정합:
- **stage 1주 관찰** — false positive ratio < 10% 유지 시 prod 임계 적용 (severity / for / window 동일화).
- **튜닝 체크리스트**:
  - error/success ratio: `increase(error[30m]) / clamp_min(increase(success[30m]), 1)` 추이
  - pull latency p95 baseline 대비 2배 이상 증가 시 latency alert 도입 trigger (현재 carve out)
  - `NoRecentSuccess` 발생 시점의 upstream 상태코드 분포 (429 / 5xx)
- 튜닝 결과는 본 가이드 + ADR-0016 §7 변경 이력 row 로 명문화 (devhub PR + 운영 자산 PR).

### 6.4 multi-instance 운영

backend-core 가 2개 이상 인스턴스로 배포된 경우:
- `last_success` 기반 알림은 `max by(provider) (devhub_homelab_last_success_unixtime)` aggregation 사용.
- instance 별 metric 은 dashboard ([homelab_dashboard.json](./grafana/homelab_dashboard.json)) 에서만 노출 — 알림은 aggregation 결과 기준.

## 7. 잔여 carve out (ADR-0016 §6)

본 가이드는 ADR-0016 §6 의 (1)+(2) 항목 해소. 다음은 후속:

- **(3) pull latency p95 alert** — stage 1주 baseline 관찰 후 임계 결정. `histogram_quantile(0.95, ...)` 기반 alert rule + 본 가이드 §4 에 추가 예정.
- **(4) push 경로 (API-73 webhook) 알림** — webhook 수신 실패율 metric (`devhub_integration_webhook_*` 후보) 도입 후 별도 ADR + 본 가이드 §4 확장.
- **(5) stage→prod 임계 확정** — 1주 관찰 후 §4 의 stage/prod 표 통합 (또는 영구 분리 결정).

## 8. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-18 | draft 1차 — ADR-0016 §6 carve (1) 의 운영 자산 git 이관 layout + raw YAML reference + Alertmanager 라우팅 + 운영 SOP. | `claude/work_260518-s` |
