# Prometheus Alerts/Dashboard Draft for HomeLab

- 문서 목적: HomeLab pull/push 관측 지표 기반의 운영 알림과 대시보드 초안을 제공한다.
- 상태: accepted
- 최종 수정일: 2026-05-20
- 관련 코드: `backend-core/internal/integrations/adapters/metrics.go`, `backend-core/internal/httpapi/router.go`

## 1. 전제

- scrape target: backend-core `/metrics`
- 핵심 메트릭:
  - `devhub_homelab_pull_runs_total{result="success|error"}`
  - `devhub_homelab_pull_duration_seconds`
  - `devhub_homelab_snapshot_services`
  - `devhub_homelab_degraded_providers`
  - `devhub_homelab_last_success_unixtime`

## 2. Alert Rule Draft

```yaml
groups:
  - name: devhub-homelab
    rules:
      - alert: DevhubHomeLabPullFailing
        expr: increase(devhub_homelab_pull_runs_total{result="error"}[10m]) >= 3
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "HomeLab pull 반복 실패"
          description: "10분 동안 pull error가 3회 이상 발생했습니다."

      - alert: DevhubHomeLabPullNoRecentSuccess
        expr: (time() - devhub_homelab_last_success_unixtime) > 900
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "HomeLab pull 성공 지연"
          description: "최근 15분 이상 pull 성공 기록이 없습니다."

      - alert: DevhubHomeLabDegradedProvidersDetected
        expr: devhub_homelab_degraded_providers > 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Degraded provider 감지"
          description: "HomeLab snapshot에 degraded provider가 10분 이상 존재합니다."
```

## 2.1 Stage 권장 임계값 (초기값)

- 목적: false positive를 줄이고 pull 경로 안정화 추세를 먼저 관찰.
- 기간: 최소 7일 관찰 후 조정.

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
      - alert: DevhubHomeLabPullNoRecentSuccess
        expr: (time() - devhub_homelab_last_success_unixtime) > 1800
        for: 10m
        labels:
          severity: critical
          environment: stage
      - alert: DevhubHomeLabDegradedProvidersDetected
        expr: devhub_homelab_degraded_providers > 0
        for: 15m
        labels:
          severity: warning
          environment: stage
```

## 2.2 Production 권장 임계값 (초기값)

- 목적: 장애 감지 시간을 단축하되 노이즈는 Alertmanager 라우팅으로 제어.

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
      - alert: DevhubHomeLabPullNoRecentSuccess
        expr: (time() - devhub_homelab_last_success_unixtime) > 900
        for: 5m
        labels:
          severity: critical
          environment: prod
      - alert: DevhubHomeLabDegradedProvidersDetected
        expr: devhub_homelab_degraded_providers > 0
        for: 10m
        labels:
          severity: warning
          environment: prod
```

## 3. Dashboard Panel Draft

- `Pull Success/Error (10m)`:
  - `increase(devhub_homelab_pull_runs_total{result="success"}[10m])`
  - `increase(devhub_homelab_pull_runs_total{result="error"}[10m])`
- `Pull Latency p95`:
  - `histogram_quantile(0.95, sum(rate(devhub_homelab_pull_duration_seconds_bucket[5m])) by (le))`
- `Last Success Age (sec)`:
  - `time() - devhub_homelab_last_success_unixtime`
- `Degraded Providers`:
  - `devhub_homelab_degraded_providers`
- `Observed Services`:
  - `devhub_homelab_snapshot_services`

## 4. 운영 메모

- file puller와 HTTP puller를 동시에 켜지 않고, 한 인스턴스에서 하나의 source mode를 사용한다.
- noise 방지를 위해 `PullFailing` 임계값(횟수/윈도우)은 stage에서 1주 이상 관찰 후 조정한다.
- multi-instance 배포 시 `last_success` 기반 알림은 집계 전략(max by instance 등)을 운영 정책에 맞춰 조정한다.

## 5. 튜닝 체크리스트

- `error/success ratio`: `increase(error[30m]) / clamp_min(increase(success[30m]), 1)` 추이 확인
- pull latency p95가 baseline 대비 2배 이상 증가하는지 확인
- `NoRecentSuccess` 발생 시점의 upstream 상태코드 분포(429/5xx) 확인
- stage에서 7일간 경보 발생 횟수/오탐 비율을 집계 후 prod 임계값 확정
