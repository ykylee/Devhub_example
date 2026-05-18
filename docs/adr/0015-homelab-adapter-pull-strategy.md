# ADR-0015: HomeLab adapter pull source 전략

- 문서 목적: HomeLab adapter 의 pull source 구현체 (file / HTTP) 와 scheduler 통합 방식 결정. PR #139 (`codex/next-step-20260516`) 으로 backend 1차가 활성화된 사실의 사후 명문화.
- 범위: `backend-core/internal/integrations/adapters/{contract,homelab,homelab_file_puller,homelab_http_puller,homelab_pull_loop,metrics}.go` 의 설계 결정. 알림 정책은 [ADR-0016](./0016-prometheus-alerts-policy.md) 가 분기. push 경로 (`POST /api/v1/integration/providers/{provider_key}/webhook`) 의 인증 정책은 본 ADR 범위 밖.
- 대상 독자: Backend 개발자, 운영자 (HomeLab agent), External Integration 도메인 stakeholder.
- 상태: accepted
- 작성일: 2026-05-18
- 결정일: 2026-05-16 (실 결정), 2026-05-18 사후 명문화 (sprint `claude/work_260518-c`)
- 결정 근거 sprint: `codex/next-step-20260516` (PR #139, sha `e2a76fb`) — External Integration backend 1차 활성화.
- 관련 문서: [`docs/planning/homelab_adapter_pull_strategy.md`](../planning/homelab_adapter_pull_strategy.md) (1차 결정 초안 — 본 ADR 이 source-of-truth 로 승격), [`docs/planning/external_system_integration_concept.md`](../planning/external_system_integration_concept.md), [ADR-0016 Prometheus alerts policy](./0016-prometheus-alerts-policy.md), [추적성 매트릭스 §3 External Integration 행 + §4 ADR](../traceability/report.md).

## 1. 컨텍스트

External Integration 도메인 (PR #135 concept staged + PR #139 backend 1차) 은 HomeLab agent 로부터 infra service snapshot 을 수신해 `infra_service_snapshots` 테이블에 영속화하고 `/api/v1/infra/*` 조회 경로 (API-76 / 77 / 78) 가 hydrate 한다. 수신 채널은 두 축:

1. **Push** (`POST /api/v1/integration/providers/{provider_key}/webhook`, API-73) — HomeLab agent 가 backend 로 직접 송부.
2. **Pull** — backend 가 HomeLab agent 의 snapshot endpoint 또는 로컬 fixture file 을 polling 해 수집.

본 ADR 은 **pull 경로** 의 구현 결정을 다룬다. 두 가지 의 source mode (file / HTTP) 와 그 실행 컨테이너 (lightweight goroutine vs dedicated worker) 의 트레이드오프를 평가.

## 2. 결정 동인

- **blast radius 최소화**: pull 실패가 command worker lifecycle, Hydra/Kratos session, audit emit 등 다른 책임에 영향 주지 않아야 한다.
- **로컬/CI 재현성**: 외부 HomeLab agent 없이 fixture JSON 으로 pull-and-ingest 경로 회귀 검증 가능해야 한다 (CI 환경 = no external network 정합 — ADR-0003 / CLAUDE.md).
- **운영 가시성**: pull success/error/latency 가 Prometheus 로 노출되어 [ADR-0016](./0016-prometheus-alerts-policy.md) 의 알림 규칙이 부착될 수 있어야 한다.
- **feature flag 1차 도입**: 운영 환경에서 점진적 활성화. on/off 토글이 인프라 변경 없이 가능해야 한다.

## 3. 검토 옵션

### 3.1 Source mode

| 옵션 | 설명 | 채택 |
| --- | --- | --- |
| A. File puller only | 로컬 fixture JSON 만 지원. HTTP 는 후속 | ❌ (HTTP 실 운영 필요) |
| B. HTTP puller only | HomeLab agent endpoint 직접 호출 | ❌ (CI/로컬 회귀 부재) |
| C. **File + HTTP 둘 다 (mutually exclusive per instance)** | 한 인스턴스에서 하나만 활성, env 로 선택 | ✅ |
| D. File + HTTP 동시 활성 | 두 source 동시 ingest | ❌ (snapshot_at 충돌 + audit 복잡도) |

### 3.2 실행 컨테이너

| 옵션 | 설명 | 채택 |
| --- | --- | --- |
| A. command worker 재사용 | 기존 `infra_commands` lifecycle 위에 pull task 추가 | ❌ (blast radius — command 실패와 pull 실패 분리 안 됨) |
| B. **lightweight goroutine + feature flag** | backend-core `main.go` 가 단일 `pullLoop` goroutine 기동, env 기반 | ✅ |
| C. dedicated OS worker (별도 binary) | 독립 binary + systemd | ❌ (no-docker 정책 호환은 가능하지만 도입 부담 ↑, M4 carve out 가능) |
| D. cron job (외부 OS cron) | 외부 cron 이 admin endpoint POST | ❌ (운영 environment 의존 ↑, OS 시간 drift) |

### 3.3 Retry / backoff 전략

| 옵션 | 채택 |
| --- | --- |
| no retry (single try, error → metric) | ❌ (network jitter 빈도) |
| **exponential backoff + max retry (env-controlled)** | ✅ `DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX` + `_BACKOFF` |
| infinite retry until success | ❌ (degraded provider 와 구분 어려움) |

## 4. 결정

### 4.1 Source mode — file + HTTP, mutually exclusive (옵션 3.1.C)

같은 backend 인스턴스에서 두 source 를 동시에 켜지 않는다. env 가 둘 다 set 이면 file mode 우선 + warning log.

env 우선순위:
1. `DEVHUB_HOMELAB_PULL_FILE` set 이면 file mode.
2. else `DEVHUB_HOMELAB_PULL_URL` set 이면 HTTP mode.
3. 모두 unset 이면 pull loop 시작 안 함 (push 만 동작).

### 4.2 실행 컨테이너 — lightweight goroutine + feature flag (옵션 3.2.B)

```
backend-core/main.go::main()
└── if DEVHUB_HOMELAB_PULL_ENABLED == "true":
    └── go adapters.RunHomeLabPullLoop(ctx, store, metrics, opts)
        └── ticker (DEVHUB_HOMELAB_PULL_INTERVAL, default 60s)
            └── puller.Pull(ctx) → adapter.Normalize() → store.IngestSnapshot()
```

- 단일 goroutine, ctx cancellation 으로 graceful shutdown.
- command worker / Hydra session / audit emit 와 격리.
- pull 실패는 metric + log, command lifecycle 영향 없음.

### 4.3 환경 변수 contract

| 변수 | 기본 | 의미 |
| --- | --- | --- |
| `DEVHUB_HOMELAB_PULL_ENABLED` | `"false"` | `"true"` 일 때만 loop 시작 |
| `DEVHUB_HOMELAB_PULL_INTERVAL` | `60s` | tick 주기 (`time.ParseDuration` 형식) |
| `DEVHUB_HOMELAB_PULL_FILE` | unset | file mode: JSON fixture 절대 경로 |
| `DEVHUB_HOMELAB_PULL_URL` | unset | HTTP mode: HomeLab agent snapshot endpoint |
| `DEVHUB_HOMELAB_PULL_TOKEN` | unset | HTTP mode: Bearer token (caller → HomeLab agent 인증) |
| `DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX` | `3` | exponential backoff 최대 시도 횟수 |
| `DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF` | `2s` | 최초 backoff 간격 (지수 증가) |

### 4.4 수용 기준 (PR #139 활성화 기준)

- File puller — malformed JSON / 필수 필드 누락 시 `ErrInvalidHomeLabSnapshot` 경로로 실패. snapshot 미반영.
- HTTP puller — 2xx 만 성공 처리. 4xx/5xx 는 degraded provider 후보 이벤트 + error metric +1.
- Pull-and-ingest 후 `infra_service_snapshots` row count 증가 + API-76/78 응답에 반영.

### 4.5 메트릭 노출 (Prometheus 정합)

본 ADR 은 메트릭 명세만 명시. 알림 규칙은 [ADR-0016](./0016-prometheus-alerts-policy.md) 가 다룬다.

- `devhub_homelab_pull_runs_total{result="success|error"}` (counter)
- `devhub_homelab_pull_duration_seconds` (histogram)
- `devhub_homelab_snapshot_services` (gauge)
- `devhub_homelab_degraded_providers` (gauge)
- `devhub_homelab_last_success_unixtime` (gauge)

5 메트릭 모두 `backend-core/internal/integrations/adapters/metrics.go` 의 Prometheus client 로 export, `/metrics` endpoint 에서 scrape.

## 5. 결과

- `backend-core/internal/integrations/adapters/contract.go` — `HomeLabPuller` interface.
- `backend-core/internal/integrations/adapters/homelab.go` — adapter normalize + health policy + ingest 경로.
- `backend-core/internal/integrations/adapters/homelab_file_puller.go` — file mode 구현.
- `backend-core/internal/integrations/adapters/homelab_http_puller.go` — HTTP mode 구현 + retry/backoff.
- `backend-core/internal/integrations/adapters/homelab_pull_loop.go` — `RunHomeLabPullLoop` scheduler.
- `backend-core/internal/integrations/adapters/metrics.go` — Prometheus collector.
- `backend-core/internal/store/infra_snapshots.go` + migration 000029 — snapshot 영속화.
- `backend-core/main.go` — env 기반 pull loop wire.
- unit test — `*_test.go` (file_puller / http_puller / adapter normalize / pull loop).
- IT 자동화 — `docs/tests/test_cases_m4_integration.md` + `reports/report_20260516_m4_integration.md`.

## 6. 후속 작업

- **(carve)** size limit + streaming decode — 대용량 payload 보호. 본 1차는 in-memory full decode.
- **(carve)** agent token rotation 정책 — HomeLab agent 의 `DEVHUB_HOMELAB_PULL_TOKEN` 갱신 절차 문서화 (운영 SOP).
- **(carve)** dedicated worker binary (옵션 3.2.C) — M4 운영 진입 시 트래픽 패턴 기반 재평가.
- **(carve)** push/pull 경로 동시 운영 시 `snapshot_at` 기준 + `source` tag 의 dedup 정책 — 현재는 file/HTTP exclusive 만 다룸. push 와의 정합은 별도 ADR 후보.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-16 | 1차 결정 (`docs/planning/homelab_adapter_pull_strategy.md` §2.1) — lightweight goroutine + feature flag. | PR #139 활성화 |
| 2026-05-18 | accepted — ADR 형식으로 사후 명문화. source mode = file + HTTP mutually exclusive, 실행 컨테이너 = lightweight goroutine + feature flag, retry = exponential backoff env-controlled. 메트릭 5종 명시 (알림 규칙은 ADR-0016 분기). | sprint `claude/work_260518-c` |
