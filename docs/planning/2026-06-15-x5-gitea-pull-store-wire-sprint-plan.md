# X-5 Gitea Hourly Pull 정밀화 — production wire follow-up (2026-06-15)

- 문서 목적: X-5 (Gitea Hourly Pull 정밀화, RM-M4-06 잔여, #231) 의 production wire follow-up sprint plan.
- 범위: `RepositoryPullStore` 9 method + `ListGiteaPullTargets` + main.go wire 교체 + adapter minor fix (state → event_type) + integration test + 정합.
- 상태: draft (X-5 PR #592 + a49a5660 fix 머지 후 본 follow-up 진입)
- 결정 근거 sprint: `feat/x5-gitea-pull-store-wire`
- 관련 문서: [release_v0-1_roadmap.md §3.5 X-5](../release_v0-1_roadmap.md), [issue #231](https://github.com/ykylee/Devhub_example/issues/231), [ADR-0034](../adr/0034-gitea-hourly-pull-architecture.md) (1차 결정), [X-5 sprint plan](./2026-06-14-x5-gitea-hourly-pull-sprint-plan.md) (1차), [ADR-0015 Homelab pull pattern](../adr/0015-homelab-adapter-pull-strategy.md) (정합).

## 0. 컨텍스트

### 0.1 X-5 1차 PR (#592 + a49a5660 fix) 정공법

- `RepositoryPullStore` interface 9 method 정의 (`backend-core/internal/integrations/adapters/gitea_pull.go:157`).
- `migration 000043_repository_pull_state` 신규 (per-repo state table + 2 partial index + trigger).
- main.go 의 `adapter.Store = nil` + `repoLister` placeholder → **운영 시 cycle fail-fast**. 본 follow-up = production wire.

### 0.2 본 follow-up 의 본질

`X-5 sprint plan` §6.1 follow-up 1번째: **production RepositoryPullStore wire**. 본 follow-up 후 `DEVHUB_GITEA_PULL_ENABLED=true` 가 production 에서 정상 동작 (cycle 마다 repositories 테이블의 gitea repo 목록 fetch → pr_activities / build_runs / quality_snapshots / repository_pull_state 갱신).

## 1. 결정

### 1.1 PostgresStore method 위치

| Method | 위치 |
|---|---|
| `UpsertPullActivity` / `UpsertBuildRun` / `UpsertQualitySnapshot` (3 method) | 신규 file `backend-core/internal/store/repository_pull_ingest.go` |
| `UpdatePullState` / `IncrementConsecutiveFailures` / `ResetConsecutiveFailures` / `SetBackoff` / `BackoffUntil` / `LastPullAt` (6 method) | 신규 file `backend-core/internal/store/repository_pull_state.go` |
| `ListGiteaPullTargets` (1 method) | 신규 file `backend-core/internal/store/repository_pull_targets.go` |

**이유**: 3 file 분리 — `ingest` (pr/build/quality upsert, X-5 1차 PR 의 pr_activities/build_runs/quality_snapshots 1:1 매핑) vs `state` (repository_pull_state CRUD) vs `targets` (repositories table read for cycle lister). cross-project lesson §1 의 "기존 pattern 준수" 정공법.

### 1.2 state → event_type 매핑 정공법

`pr_activities.event_type` enum (migration 000001 L411): `opened/reviewed/commented/closed/merged/reopened/updated`. Gitea PR API response = `state` (open/closed) + `merged` (bool) + `merged_at` (nullable).

**Decision**:
- `state="open"` → `event_type="opened"`
- `state="closed" && merged=true` → `event_type="merged"`
- `state="closed" && merged=false` → `event_type="closed"`
- 그 외 (all, unknown) → `event_type="updated"` (defensive fallback, CHECK constraint 통과)

**책임**: adapter (`gitea_pull.go:241` 의 `pr.State` 전달) 가 event_type 결정 후 `UpsertPullActivity` 의 5번째 인자로 전달. **store 는 CHECK constraint 통과만 보장**. **interface 변경 없음** (의미는 adapter 가 결정).

adapter minor fix: `GiteaPullRequest` struct 에 `Merged bool` field 추가, `GiteaClient.ListPullRequestsSince` 가 Gitea API response 의 `merged` boolean 파싱. `merged_at` 으로 merged 검증 보조.

### 1.3 `b.Event` (Gitea build event) 매핑 정공법

Gitea Actions API response = `event` (push/pull_request/schedule) + `status` (running/success/failed/cancelled/skipped) + `conclusion` (success/failure/cancelled/skipped/neutral) + `commit_sha` + `created_at` + `updated_at` + `branch` (nullable).

`build_runs` schema (migration 000001 L95-108): `branch text NOT NULL, commit_sha text NOT NULL, status text, duration_seconds int, started_at, finished_at, created_at`. `event`/`conclusion`/`updated_at` 는 **별도 column 없음**.

**Decision**:
- `branch`: `b.Event` 가 비어있지 않으면 `b.Event` 그대로 (e.g. "pull_request"), 비어있으면 `repositories.default_branch` fallback.
- `commit_sha`: `b.CommitSHA` 1:1.
- `status`: `b.Status` 1:1 (running/success/failed/cancelled/skipped → migration CHECK constraint 통과).
- `duration_seconds`: `b.UpdatedAt - b.CreatedAt` 차이 (초, int) 추정. 0 또는 음수면 0.
- `started_at`: `b.CreatedAt`.
- `finished_at`: `b.UpdatedAt`.
- `conclusion`: **무시** (status 가 같은 정보 보유). 단 metric payload 차원에서 향후 활용 가능성.

`UpsertBuildRun` 의 7번째 인자 `conclusion` 은 **adapter 가 store 에 전달 시점에서 status 와 일치** 보장. store 의 책임은 column 매핑만.

### 1.4 `UpsertQualitySnapshot` 매핑 정공법

`quality_snapshots` schema (migration 000001 L505-517): `tool text NOT NULL, ref_name text NOT NULL, commit_sha text, score numeric(6,2), gate_passed boolean, metric_payload jsonb, measured_at timestamptz, created_at`.

**Decision**:
- `tool = 'gitea-build'` (X-5 1차 PR 의 본문 의도).
- `ref_name = commitSHA[:40]` (git SHA 40자).
- `commit_sha = commitSHA`.
- `score = NULL` (Gitea API response 에 score field 없음).
- `gate_passed = NULL`.
- `metric_payload = '{}'::jsonb` (defensive).
- `measured_at = recordedAt` (adapter 의 `b.UpdatedAt`).

### 1.5 ListGiteaPullTargets query

```sql
SELECT r.id, r.owner_login, r.name, r.gitea_repository_id
FROM repositories r
JOIN integration_providers p ON r.provider_id = p.provider_id
WHERE p.provider_type = 'scm'
  AND p.provider_key = 'gitea'  -- Gitea 한정
  AND r.repository_status = 'active'
  AND r.gitea_repository_id IS NOT NULL
  AND (s.backoff_until IS NULL OR s.backoff_until < now())
  AND r.deleted_at IS NULL
ORDER BY r.id ASC
```

**LEFT JOIN repository_pull_state s ON r.id = s.repository_id** — state table 부재해도 (cold start) repo 목록에 포함 (backoff_until IS NULL 분기로 skip 안 됨).

**책임**: lister 가 `backoff` check 를 자체적으로. loop 의 per-repo 호출에서 adapter 가 다시 `BackoffUntil` 호출하여 double-check (race 방지). 단, race 가능성 낮음 (cycle 1h, lister 1회 cycle 시점 단일).

### 1.6 main.go wire

```go
// production wire: store + lister 모두 PostgresStore.
adapter := &adapters.GiteaPullAdapter{
    Client:         giteaClient,
    Store:          pgStore,  // 변경: nil → pgStore
    MaxItemsPerCall: 200,
}
repoLister := func(ctx context.Context) ([]adapters.RepositoryTarget, error) {
    // 변경: nil → pgStore.ListGiteaPullTargets(ctx)
    return pgStore.ListGiteaPullTargets(ctx)
}
```

`pgStore` 변수 — main.go 내 다른 cron loop (DREQ token cron 등) 의 store pattern 동일.

### 1.7 Tier

- **공용** (backend only, 사내 한정 정보 미포함).
- 사내 검증 SOP = 별도 follow-up.

## 2. 변경 범위 (PR 1개, ~600 line)

### 2.1 backend (5 file)

1. `backend-core/internal/store/repository_pull_ingest.go` (NEW, ~180 line)
   - `UpsertPullActivity` (pr_activities UPSERT)
   - `UpsertBuildRun` (build_runs UPSERT)
   - `UpsertQualitySnapshot` (quality_snapshots UPSERT)

2. `backend-core/internal/store/repository_pull_state.go` (NEW, ~200 line)
   - 6 method (UpdatePullState / IncrementConsecutiveFailures / ResetConsecutiveFailures / SetBackoff / BackoffUntil / LastPullAt).

3. `backend-core/internal/store/repository_pull_targets.go` (NEW, ~80 line)
   - `ListGiteaPullTargets` — store 자체 type 반환 + main.go 에서 adapter type 매핑.

4. `backend-core/main.go` (MODIFY, +15 line)
   - L429 코멘트 제거 + L435 `Store: nil` → `Store: pgStore` + L438-441 repoLister closure 교체.

5. `backend-core/internal/integrations/adapters/gitea_pull.go` (MODIFY, +20 line)
   - `GiteaPullRequest.Merged` field 추가 + adapter state → event_type 매핑 결정.

### 2.2 migration (1 file, 신규)

6. `backend-core/migrations/000045_quality_snapshots_ref_name_unique.up.sql` (NEW, ~10 line)
   - partial unique index (tool='gitea-build' 한정).

### 2.3 test (2 file)

7. `backend-core/internal/store/repository_pull_ops_integration_test.go` (NEW, ~200 line)
   - 1 integration test (`TestIntegration_RepositoryPullState_AllMethods` + ListGiteaPullTargets filter 검증).

8. `backend-core/internal/integrations/adapters/gitea_pull_test.go` (MODIFY, +30 line)
   - `TestAdapter_StateToEventType_*` 3 test.

### 2.4 docs (3 file)

9. `docs/adr/0034-gitea-hourly-pull-architecture.md` (MODIFY, +30 line)
   - §8 변경 이력 + §5 cross-tier 표.

10. `docs/planning/2026-06-15-x5-gitea-pull-store-wire-sprint-plan.md` (NEW, 본 문서)

11. `docs/traceability/report.md` (MODIFY, +1 row)
    - §6 본 row 추가.

12. `CHANGELOG.md` (MODIFY, +1 row)
    - v0.1.1-alpha X-5 follow-up status 갱신.

### 2.5 메모리 (4 file)

13-16. `ai-workflow/memory/feat/x5-gitea-pull-store-wire/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-15.md}` (NEW)

## 3. 신규 ID 발급 (4 row)

- `IMPL-GITEA-PULL-STORE-01` (RepositoryPullStore 9 method + ListGiteaPullTargets)
- `IMPL-GITEA-PULL-INGEST-01` (state → event_type 매핑 adapter fix)
- `UT-GITEA-PULL-STORE-01` (3 adapter test)
- `TC-GITEA-PULL-STORE-01` (1 integration test)

기존 X-5 1차 PR 발급 ID (8 row) 의 follow-up 정합.

## 4. 검증

- `go test ./internal/integrations/adapters/...` 9 unit test PASS (기존 8 + adapter state→event_type 3 신규)
- `go test ./internal/store/...` integration test 1건 (DEVHUB_TEST_DB_URL 설정 시) PASS
- `go test ./...` 30+ packages 회귀 0
- `go build ./...` silent PASS
- openapi lint PASS (변경 0)
- backend e2e shard 1/2/3 skip (path-detect: backend only)
- CI 4/4 PASS 예상

## 5. 잔여 (follow-up)

- **staging Gitea 실 검증 SOP** (사내 follow-up) — production wire 후 1h cycle 실 검증.
- **e2e spec (staging Gitea mock + 1h cycle)** (사내 follow-up).
- **frontend widget (Gitea pull 상태)** (X-1 admin dashboard 와 통합, 사외 가능).
- **PR/commit status mapping 정밀화** (Gitea commit SHA → DevHub quality_snapshots 매핑 정밀화, 본 follow-up 의 1차 매핑 + 후속).
- **repository_pull_state.last_alert_at** (5회 연속 실패 시 alert emit 시각, 본 follow-up scope 외, ADR-0034 §3.1 의 향후 follow-up).

## 6. 다음 세션 directive

- 본 sprint 진입 시 main HEAD rebase.
- PR 1개 (commit N개, 600 line cap 이내).
- PR 머지 후 main flat memory 3 file finalize + 위키 mirror 갱신.

## 7. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-15 | 본 sprint plan 초안 (X-5 Gitea Hourly Pull production wire follow-up) | `feat/x5-gitea-pull-store-wire` |
