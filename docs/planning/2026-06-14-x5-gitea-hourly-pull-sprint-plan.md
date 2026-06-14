# X-5 Gitea Hourly Pull 정밀화 — Sprint Plan (2026-06-14)

- 문서 목적: X-5 (Gitea Hourly Pull 정밀화, RM-M4-06 잔여, #231) 의 sprint plan.
- 범위: Gitea API → DevHub `pr_activities` / `build_runs` / `quality_snapshots` 비동기 sync + per-repository last_pull_at 추적 + 5 metric + 4 audit + ADR-0034 + openapi + UT.
- 상태: draft (X-3 PR #591 머지 후 본 sprint 진입)
- 결정 근거 sprint: `feat/x5-gitea-hourly-pull`
- 관련 문서: [release_v0-1_roadmap.md §3.5 X-5](../release_v0-1_roadmap.md), [issue #231](https://github.com/ykylee/Devhub_example/issues/231), [ADR-0015 HomeLab pull strategy](../adr/0015-homelab-adapter-pull-strategy.md) (정합 패턴), [ADR-0016 Prometheus alerts](../adr/0016-prometheus-alerts-policy.md) (5 metric), [Homelab pull loop 패턴](../../backend-core/internal/integrations/adapters/homelab_pull_loop.go).

## 0. 컨텍스트

### 0.1 v1.0 현재 상태

- **Homelab pull**: `internal/integrations/adapters/homelab_pull_loop.go` 가 30초 interval 로 운영 중. 1 adapter = 1 file pull. 5 metric (`devhub_homelab_pull_runs_total`, `_duration_seconds`, `_snapshot_services`, `_degraded_providers`, `_last_success_unixtime`).
- **Gitea integration**: webhook (push/PR) 은 active (sprint -h 의 N-7) — `gitea_webhook.go` 가 API-72 webhook 수신. 하지만 **Gitea → DevHub 비동기 pull (역방향, 외부 SCM pull)** 은 **미구현** (issue #231 정공법).
- **v1.0 정공법 = webhook + frontend 수동 refresh**. v1.1 = webhook + hourly pull reconcile. 본 X-5 = v1.1 의 cron worker 신규 + 정밀화.

### 0.2 issue #231 작업 범위 (정합)

> Gitea API → DevHub `pr_activities` / `build_runs` / `quality_snapshots` 비동기 sync
> hourly cron (또는 event-driven)
> v1.0 은 HomeLab pull 만, Gitea pull 은 v2

본 sprint = v2 의 1차 진입 (X-5 v0.1.1 carve).

## 1. 결정

### 1.1 sync 방식

| 옵션 | 채택 | 이유 |
|---|---|---|
| webhook only (현재) | ❌ | Gitea push/PR event 만 잡힘. batch reconcile / historical backfill / 누락 복구 불가 |
| **hourly cron (PR-1)** | ✅ | webhook 의 보완. 1h staleness 허용. metric + alert 으로 운영 가시성 |
| event-driven (Gitea → DevHub outbound hook) | ❌ | Gitea outbound hook 미지원 (issue #231 정공법 — webhook push 만 가능) |
| full backfill on startup | ❌ | cold start 시 1회성 full sync. 본 sprint scope 외 (별도 carve) |

### 1.2 per-repository state 추적

- 신규 테이블 `repository_pull_state` (repository_id PK + last_pull_at timestamptz + last_pull_status text + last_pull_error text + consecutive_failures int + backoff_until timestamptz)
- pull query = Gitea API `?since=last_pull_at` (per-repo, Gitea 가 지원 시 — `updated_after` filter)
- 1회 실패 시 `consecutive_failures++` + `backoff_until = now() + min(2^failures, 24h)` (exponential backoff, 24h cap)
- 5회 연속 실패 시 alert (ADR-0016 §6.1 의 DevhubHomeLabPullFailing 와 동일 패턴, Gitea 별도 metric)
- 1회 성공 시 `consecutive_failures = 0` + `backoff_until = NULL`

### 1.3 multi-repository parallelism

- semaphore (max 4 concurrent Gitea API call, configurable via `DEVHUB_GITEA_PULL_CONCURRENCY`)
- partial failure 격리: 한 repository fail 이 다른 repository 의 sync 를 block 안 함
- 1 cycle = 모든 repo 의 last_pull_at 기준 since query, 4 concurrent

### 1.4 metric (5종, ADR-0016 정합)

| Metric | Type | Labels | 의미 |
|---|---|---|---|
| `devhub_gitea_pull_runs_total` | counter | `result=success|error|partial` | pull cycle 결과 |
| `devhub_gitea_pull_duration_seconds` | histogram | `result` | pull cycle duration |
| `devhub_gitea_pull_repositories_total` | gauge | — | 현재 sync 대상 repo 수 |
| `devhub_gitea_pull_consecutive_failures` | gauge | `repository_id` | per-repo consecutive failure |
| `devhub_gitea_pull_last_success_unixtime` | gauge | — | 마지막 success 시각 |

### 1.5 audit (4종)

- `gitea.pull_started` (per-cycle, payload: `cycle_id` + `repository_count`)
- `gitea.pull_success` (per-cycle, payload: `cycle_id` + `repositories_synced` + `duration_ms`)
- `gitea.pull_error` (per-cycle, payload: `cycle_id` + `error_class` + `error_message`)
- `gitea.pull_partial` (per-repository, payload: `repository_id` + `error_class` + `error_message` + `consecutive_failures`)

### 1.6 env config

| Env | Default | 의미 |
|---|---|---|
| `DEVHUB_GITEA_PULL_ENABLED` | `false` | opt-in, prod-safe |
| `DEVHUB_GITEA_PULL_INTERVAL` | `1h` | cycle interval |
| `DEVHUB_GITEA_PULL_CYCLE_TIMEOUT` | `30m` | 1 cycle 의 hard cap |
| `DEVHUB_GITEA_PULL_CONCURRENCY` | `4` | max concurrent repo pull |
| `DEVHUB_GITEA_PULL_BACKOFF_CAP` | `24h` | exponential backoff cap |
| `DEVHUB_GITEA_PULL_FAILURE_ALERT_THRESHOLD` | `5` | alert 발생 consecutive failures |

## 2. 변경 범위 (PR 1개, 코드 ~1200 line)

### 2.1 backend (6 file)

1. `backend-core/migrations/000046_repository_pull_state.up.sql` (NEW, ~30 line)
   - `repository_pull_state` (repository_id uuid PK + last_pull_at timestamptz + last_pull_status text + last_pull_error text + consecutive_failures int default 0 + backoff_until timestamptz + updated_at timestamptz default now())

2. `backend-core/internal/integrations/adapters/gitea_pull.go` (NEW, ~250 line)
   - `GiteaPullAdapter` interface (`PullAndIngestSince(ctx, repositoryID, since time.Time) error`)
   - `GiteaClient` struct (Gitea API client — list pull requests since X, list builds, list commits)
   - `GiteaPullAdapter` 구현: since 기반 pr_activities upsert + build_runs upsert + quality_snapshots upsert
   - 5 metric emission (homelab 동일 패턴)

3. `backend-core/internal/integrations/adapters/gitea_pull_loop.go` (NEW, ~80 line)
   - `RunGiteaPullLoop(ctx, adapter, interval, onError, cycleTimeout, concurrency, backoffCap) error`
   - `runGiteaPullOnce` per-cycle (전체 repo 목록 fetch + per-repo goroutine + semaphore + partial failure 격리)
   - 4 audit emission
   - homelab 패턴 미러

4. `backend-core/internal/shared/config/config.go` (MODIFY, +20 line)
   - `GiteaPullEnabled` + `GiteaPullInterval` + `GiteaPullCycleTimeout` + `GiteaPullConcurrency` + `GiteaPullBackoffCap` + `GiteaPullFailureAlertThreshold` 필드 + env wire

5. `backend-core/main.go` (MODIFY, +30 line)
   - `DEVHUB_GITEA_PULL_ENABLED` true 시 `go RunGiteaPullLoop(...)` 시작
   - graceful shutdown 시 ctx cancel

6. `backend-core/internal/integrations/adapters/gitea_pull_test.go` (NEW, ~400 line)
   - `TestGiteaPullAdapter_PullAndIngestSince_Success` (mock Gitea server + 3 PR/2 build/1 commit → upsert 검증)
   - `TestGiteaPullAdapter_PullAndIngestSince_GiteaUnreachable` (500 + retry → error + backoff)
   - `TestGiteaPullAdapter_PullAndIngestSince_PartialResponse` (2 PR/3 PR 만 응답 → 1 missing → audit partial)
   - `TestGiteaPullLoop_FirstCycle` (전체 repo 1 cycle)
   - `TestGiteaPullLoop_PartialFailureIsolation` (2 repo 중 1 fail → 1 success)
   - `TestGiteaPullLoop_Backoff` (5회 연속 fail → backoff 24h cap)
   - `TestGiteaPullLoop_AlertThreshold` (5회 fail → metric emit)
   - `TestGiteaPullLoop_Shutdown` (ctx cancel)

### 2.2 docs (3 file)

1. `docs/adr/0034-gitea-hourly-pull-architecture.md` (NEW, ~12KB)
   - §1 상태: Accepted
   - §2 컨텍스트: webhook only 정공법의 한계
   - §3 검토 옵션: 4가지 (webhook only / hourly / event-driven / backfill)
   - §4 결정: hourly + per-repo state + 4 concurrent + 24h backoff cap
   - §5 trade-off: cold start / Gitea API rate limit / partial failure / 24h cap 정당화
   - §6 검증: 8 unit test PASS
   - §7 cross-tier: 사외 metric + audit, 사내 staging Gitea 검증 SOP
   - §8 supersession: ADR-0015 §5 (Homelab 1차 본) 와 cross-ref, conflict 없음
   - §9 변경 이력

2. `docs/traceability/report.md` (MODIFY, +1 row)
   - §6 본 row 추가

3. `docs/openapi.yaml` (MODIFY, +0 line, 변경 0)
   - 본 sprint = backend only. openapi 영향 0 (cron worker 신규, 외부 endpoint 미추가). 다만 §6.3 (operational) 의 webhook pattern 미변경 명시.

### 2.3 메모리 (4 file)

1. `ai-workflow/memory/feat/x5-gitea-hourly-pull/state.json` (NEW)
2. `ai-workflow/memory/feat/x5-gitea-hourly-pull/session_handoff.md` (NEW)
3. `ai-workflow/memory/feat/x5-gitea-hourly-pull/work_backlog.md` (NEW)
4. `ai-workflow/memory/feat/x5-gitea-hourly-pull/backlog/2026-06-14.md` (NEW)

## 3. 신규 ID 발급 (8 row)

- `REQ-FR-GITEA-PULL-01` (Gitea hourly pull 정공법)
- `ARCH-GITEA-PULL-01` (per-repo state + 4 concurrent + 24h backoff 결정)
- `API-109` (Gitea API client wrapper, 5 method: list PRs since / list builds / list commits / get commit / get build)
- `RM-GITEA-PULL-01` (cron worker 운영 정합)
- `IMPL-GITEA-PULL-01` (GiteaPullAdapter + GiteaClient + RunGiteaPullLoop)
- `IMPL-GITEA-PULL-STATE-01` (repository_pull_state migration + repository)
- `UT-GITEA-PULL-01` (8 unit test)
- `TC-GITEA-PULL-01` (e2e: staging Gitea 1h sync, 후속 sprint)

## 4. Tier

- **공용** (backend + docs + ADR + memory, 사내 한정 정보 미포함)
- 사내 검증 SOP = 별도 follow-up (Gitea instance URL / API token / staging Gitea test account 등)

## 5. 검증

- `go test ./internal/integrations/adapters/...` 8 신규 unit test PASS
- `go test ./...` 30+ packages 회귀 0
- `go build ./...` silent PASS
- openapi lint PASS (변경 0)
- backend e2e shard 1/2/3 skip (path-detect: backend only 변경, frontend/e2e 영향 0)

## 6. 잔여 (follow-up)

- **frontend widget**: Gitea pull 상태 표시 (system admin 대시보드, X-1 1차 PR #583/#584 의 admin dashboard 와 통합)
- **staging Gitea 실 검증 SOP**: 1h cycle 동작 확인
- **e2e spec**: staging Gitea mock + 1h cycle (본 sprint scope 외)
- **PR/commit status mapping**: Gitea commit SHA → DevHub `quality_snapshots` 매핑 정밀화

## 7. 다음 세션 directive

- 본 sprint 진입 시 X-3 PR #591 머지 확인 + main 정합
- 본 sprint = PR 1개 (backend + ADR + memory, ~1200 line)
- PR 머지 후 main flat memory 3 file finalize + 위키 mirror 갱신
- X-4 / X-6 / X-8 (사내 2건) follow-up 결정

## 8. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 본 sprint plan 초안 (X-5 Gitea Hourly Pull 정밀화) | `feat/x5-gitea-hourly-pull` |
