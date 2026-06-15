# HomeLab Adapter Pull Source Strategy

- 문서 목적: HomeLab adapter의 pull source 구현체 도입 단계를 합의 가능한 최소 단위로 정의한다.
- 상태: accepted
- 최종 수정일: 2026-05-20
- 관련 코드: `backend-core/internal/integrations/adapters/{contract.go,homelab.go,homelab_http_puller.go}`

## 1. 현재 baseline

- `HomeLabPuller` contract는 도입되어 있음.
- `HomeLabAdapter.PullAndIngest` 경로는 구현/테스트 완료.
- pull source 구현체(file/http)는 도입됨.

## 2. 구현 우선순위

1. File puller (P0)
- 용도: 로컬/CI에서 fixture JSON으로 pull-path 회귀 검증.
- 입력: `DEVHUB_HOMELAB_PULL_FILE` (json file path).
- 출력: `HomeLabRawSnapshot`.
- 상태: implemented (`internal/integrations/adapters/homelab_file_puller.go`).

2. HTTP puller (P1)
- 용도: HomeLab agent endpoint에서 snapshot pull.
- 입력: `DEVHUB_HOMELAB_PULL_URL`, `DEVHUB_HOMELAB_PULL_TOKEN`.
- 정책: timeout/retry/backoff는 최소값부터 시작.
- 상태: implemented (`internal/integrations/adapters/homelab_http_puller.go`).

3. Scheduler integration (P1)
- 용도: 주기적 pull + ingest 트리거.
- 경계: command worker와 분리된 lightweight goroutine or dedicated worker.

## 2.1 실행 방식 결정 (2026-05-16)

- 1차 결정: **lightweight goroutine + feature flag**.
- 제어 변수:
  - `DEVHUB_HOMELAB_PULL_ENABLED`
  - `DEVHUB_HOMELAB_PULL_INTERVAL`
  - `DEVHUB_HOMELAB_PULL_FILE` (file mode 우선)
  - `DEVHUB_HOMELAB_PULL_URL` + `DEVHUB_HOMELAB_PULL_TOKEN` (HTTP mode)
  - `DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX`
  - `DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF`
- 이유:
  - command worker 책임과 분리해 blast radius를 최소화.
  - pull 실패가 command lifecycle에 영향 주지 않도록 격리.
  - 로컬/CI에서 file puller 기반으로 재현성 높은 회귀 테스트 가능.

## 3. 수용 기준

- File puller: malformed JSON/필수 필드 누락 시 `ErrInvalidHomeLabSnapshot` 경로로 실패.
- HTTP puller: 2xx만 성공 처리, 4xx/5xx는 degraded candidate 이벤트로 보고.
- Pull-and-ingest 실행 후 `infra_service_snapshots` row 증가 + API-76/78 조회 반영.

## 4. 리스크와 완화

- pull/push 동시 운영 충돌: `snapshot_at` 기준 최신 우선 + source tag 기록 필요.
- 대용량 payload: size limit + streaming decode 후속 필요.
- auth drift: agent token rotation 정책 문서화 필요.

## 5. Prometheus 도입 검토 (결론)

- 결론: **적합함**. 외부 서비스/서버 상태 관리 목적에 맞고, pull/push 혼합 구조의 관측성 확보에 유리.
- 1차 권장 메트릭:
  - `devhub_homelab_pull_runs_total{result=\"success|error\"}`
  - `devhub_homelab_pull_duration_seconds` (histogram)
  - `devhub_homelab_snapshot_services` (gauge)
  - `devhub_homelab_degraded_providers` (gauge)
  - `devhub_homelab_last_success_unixtime` (gauge)
- 도입 순서:
  1. backend-core `/metrics` endpoint 노출 (Prometheus scrape target).
  2. pull loop + ingest handler에 success/error/latency 계측 삽입.
  3. 알림 규칙(Alertmanager): pull 실패 연속, degraded provider 수 증가, last_success staleness.
- 상태:
  - 1, 2번 완료.
  - 3번은 `docs/domain/integration-registry/prometheus_homelab_alerts.md` 초안에 정리.
