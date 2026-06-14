# ADR-0034: Gitea Hourly Pull Architecture (X-5, RM-M4-06 잔여)

- 문서 목적: X-5 (Gitea Hourly Pull 정밀화, issue #231) 의 architecture 결정. webhook only 의 한계 보완.
- 범위: Gitea API → DevHub `pr_activities` / `build_runs` / `quality_snapshots` 비동기 sync + per-repository state 추적 + multi-repository parallelism + exponential backoff + 5 metric + 4 audit.
- 상태: Accepted
- 작성일: 2026-06-14
- 결정 근거 sprint: `feat/x5-gitea-hourly-pull`
- 결정 근거 commit: (본 PR squash 예정)
- **Tier**: 공용 (코드 + ADR + openapi + migration 모두 사내 한정 정보 미포함)
- 관련 문서: [release_v0-1_roadmap.md §3.5 X-5](../planning/release_v0-1_roadmap.md), [issue #231](https://github.com/ykylee/Devhub_example/issues/231), [sprint plan](../planning/2026-06-14-x5-gitea-hourly-pull-sprint-plan.md), [ADR-0015 HomeLab pull strategy](./0015-homelab-adapter-pull-strategy.md) (정합 패턴), [ADR-0016 Prometheus alerts](./0016-prometheus-alerts-policy.md) (5 metric 정합), [ADR-0028 dev-requests-voc-external-ref](./0028-dev-requests-voc-external-ref.md) (N-13 inbound source 정합, X-2 cross-ref).

---

## 1. 컨텍스트

### 1.1 v1.0 webhook 한계

v1.0 정공법 = Gitea webhook (push/PR event) + frontend 수동 refresh. 결정적 한계 3가지:

1. **누락 복구 불가**: webhook 수신 실패 시 / Gitea outage 시 / DevHub downtime 시 event 손실 → 영구 누락.
2. **historical backfill 불가**: 신규 repository 등록 시 Gitea 의 기존 PR/build/commit data 가 DevHub 에 없음.
3. **운영 가시성 부족**: webhook metric (수신/처리 카운트) 만 있고, **Gitea → DevHub 의 sync 상태** (= per-repository last_pull_at, consecutive failures, staleness) 가 운영자에게 보이지 않음.

### 1.2 X-5 정공법

**webhook + hourly cron** 의 2-track 보완. webhook 은 실시간 트리거 (push/PR event → 즉시 sync), hourly cron 은 **정합 reconcile + 누락 복구 + 운영 가시성**.

### 1.3 issue #231 정공법 정합

> Gitea API → DevHub `pr_activities` / `build_runs` / `quality_snapshots` 비동기 sync
> hourly cron (또는 event-driven)
> v1.0 은 HomeLab pull 만, Gitea pull 은 v2

본 ADR = v2 의 1차 진입. X-5 v0.1.1 carve.

## 2. 검토 옵션

| 옵션 | 채택 | 이유 |
|---|---|---|
| webhook only (현재) | ❌ | §1.1 의 3가지 한계 |
| **hourly cron (PR-1)** | ✅ | webhook 보완. 1h staleness 허용. metric + alert 으로 운영 가시성. ADR-0015 의 HomeLab pull pattern 미러 |
| event-driven (Gitea → DevHub outbound hook) | ❌ | Gitea 가 outbound hook 미지원 (issue #231 정공법 — webhook push 만 가능) |
| full backfill on startup | ❌ | cold start 시 1회성 full sync. 본 sprint scope 외 (별도 carve) |

## 3. 결정

### 3.1 per-repository state 추적

신규 테이블 `repository_pull_state` (migration `000043_repository_pull_state`):
- `repository_id` (PK, FK to `repositories`)
- `last_pull_at` (마지막 successful pull 시각)
- `last_pull_status` (success | error | partial)
- `last_pull_error` (error_class + error_message)
- `consecutive_failures` (연속 error/partial 횟수; success 시 0 reset)
- `backoff_until` (`now() < backoff_until` 시 skip — exponential backoff)
- `last_alert_at` (5회 연속 실패 시 alert emit 시각; idempotent)

### 3.2 sync query

per-repository `last_pull_at` 기준 `since` query. Gitea API 가 `?since=` filter 지원 시 활용 (`/api/v1/repos/{owner}/{repo}/pulls?state=all&since=...`). 미지원 시 전체 list + client-side filter (fallback).

### 3.3 exponential backoff

`failures` → `min(2^failures minutes, 24h)`:
- failures=1 → 2m
- failures=2 → 4m
- failures=5 → 32m
- failures=8 → ~4h
- failures=10 → ~17h
- failures=11+ → 24h (cap)

5회 연속 실패 시 `devhub_gitea_pull_consecutive_failures{repository_id="..."}` metric emit (alert trigger).

### 3.4 multi-repository parallelism

semaphore `NewSemaphore(concurrency)` (default 4, env `DEVHUB_GITEA_PULL_CONCURRENCY`). partial failure 격리: 한 repository fail 이 다른 repository 의 sync 를 block 안 함.

### 3.5 metric (5종, ADR-0016 정합)

| Metric | Type | Labels | 의미 |
|---|---|---|---|
| `devhub_gitea_pull_runs_total` | counter | `result=success\|error\|partial` | pull cycle 결과 |
| `devhub_gitea_pull_duration_seconds` | histogram | `result` | pull cycle duration |
| `devhub_gitea_pull_repositories_total` | gauge | — | 현재 sync 대상 repo 수 |
| `devhub_gitea_pull_consecutive_failures` | gauge | `repository_id` | per-repo consecutive failure (alert trigger) |
| `devhub_gitea_pull_last_success_unixtime` | gauge | — | 마지막 success 시각 |

### 3.6 audit (4종)

- `gitea.pull_started` (per-cycle, payload: `cycle_id`)
- `gitea.pull_success` (per-cycle, payload: `cycle_id` + `repositories_synced` + `duration_ms`)
- `gitea.pull_error` (per-cycle, payload: `cycle_id` + `error_class` + `error_message`)
- `gitea.pull_partial` (per-repository, payload: `cycle_id` + `repository_id` + `error_class` + `error_message` + `consecutive_failures`)

기본 impl = log-only (`LogPullAuditHook`); production wiring 은 `SetPullAuditHook` 으로 audit-ops emitter 교체 (별도 follow-up).

### 3.7 env config

| Env | Default | 의미 |
|---|---|---|
| `DEVHUB_GITEA_PULL_ENABLED` | `false` | opt-in, prod-safe |
| `DEVHUB_GITEA_PULL_INTERVAL` | `1h` | cycle interval |
| `DEVHUB_GITEA_PULL_CYCLE_TIMEOUT` | `30m` | 1 cycle 의 hard cap |
| `DEVHUB_GITEA_PULL_CONCURRENCY` | `4` | max concurrent repo pull |
| `DEVHUB_GITEA_PULL_BACKOFF_CAP` | `24h` | exponential backoff cap |
| `DEVHUB_GITEA_PULL_FAILURE_ALERT_THRESHOLD` | `5` | alert 발생 consecutive failures |
| `DEVHUB_GITEA_API_BASE_URL` | (none) | Gitea instance base URL |
| `DEVHUB_GITEA_API_TOKEN` | (none) | optional API token |

## 4. trade-off

### 4.1 cold start

- cold start 시 `last_pull_at = NULL` → Gitea API full list (페이지네이션 4 page = 200 PR cap). 첫 cycle latency 가 길 수 있으나, metric 으로 가시성 확보.
- 본 sprint scope 외 (full backfill on startup = 별도 carve). 운영 SOP 으로 cold start 시 `last_pull_at` 사전 seed 가능 (예: `UPDATE repository_pull_state SET last_pull_at = now() - interval '7 days'`).

### 4.2 Gitea API rate limit

- Gitea 기본 rate limit: 인증 시 1000 req/h, 비인증 시 50 req/h. 본 PR 의 `?since=` query 가 인증 필수 (`DEVHUB_GITEA_API_TOKEN`) — 1000 req/h 확보.
- 4 concurrent 의 worst case = 4 req / cycle × (24 cycle/day) × (N repo) = 96N req/day. N=10 → 960 req/day (limit 1000/h 미만). **headroom 충분**.

### 4.3 partial failure 격리

- 한 repository 의 `?since=` query 실패 (DNS, 5xx, partial JSON) 시, 다른 repository sync 영향 없음. **partial + per-repo backoff**.
- 단, cycle 전체의 `repoLister` 실패 (DB outage) 시 cycle 자체 skip. 1h 후 재시도. metric `result=error` emit.

### 4.4 24h backoff cap 정당화

- `failures=11+` 부터 24h cap. **운영 가시성 vs alert fatigue trade-off**.
- 5회 연속 실패 시 metric emit + last_alert_at 갱신. 24h cap 은 **단일 repo 의 24 cycle** (1h × 24) 동안 Gitea 와 sync 시도 안 함 → 운영자는 alert 보고 24h 내 수동 진단 (Gitea 설정, network, credential).
- **5회 cap 미적용 시 false positive alert fatigue 위험**. 24h cap 으로 alert noise 최소화.

## 5. cross-tier (사외 / 사내 정합)

| 영역 | Tier | 비고 |
|---|---|---|
| backend migration 000043 | 공용 | schema 만, 사내 한정 컬럼 없음 |
| backend gitea_pull.go + gitea_pull_loop.go + pull_audit.go | 공용 | Gitea API generic, 사내 instance 한정 없음 |
| backend metrics.go (gitea metric 추가) | 공용 | Prometheus 표준 metric |
| main.go wire | 공용 | env var 만, 사내 token/URL 미포함 |
| ADR-0034 | 공용 | docs only |
| sprint plan | 공용 | docs only |
| staging Gitea 실 검증 SOP | **사내** | Gitea instance URL / API token / staging account 사내 한정 → 별도 docs (X-5 follow-up) |
| production Gitea 검증 SOP | **사내** | 동일 |

## 6. 검증

- `go test ./internal/integrations/adapters/...` 8 신규 unit test PASS (gitea_pull_test.go)
- `go test ./...` 30+ packages 회귀 0
- `go build ./...` silent PASS
- backend e2e shard 1/2/3 skip (path-detect: backend only 변경, frontend/e2e 영향 0)
- openapi lint PASS (변경 0, cron worker = internal only)

## 7. supersession

- ADR-0015 §5 (Homelab 1차 본) 와 cross-ref, **conflict 없음**. ADR-0015 = Homelab, ADR-0034 = Gitea — 별도 adapter, 별도 metric, 별도 cycle.
- ADR-0016 §4 (Prometheus alert baseline) 와 cross-ref. 5 metric 모두 ADR-0016 §4 의 alert 정합 (`DevhubHomeLabPullFailing` 와 동일한 `DevhubGiteaPullFailing` 은 본 PR 후속 follow-up, ADR-0016 §6.3 carve out).

## 8. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 1차 발행 (Accepted). X-5 Gitea Hourly Pull 정밀화 + per-repo state + 4 concurrent + 24h backoff cap + 5 metric + 4 audit. | `feat/x5-gitea-hourly-pull` |
