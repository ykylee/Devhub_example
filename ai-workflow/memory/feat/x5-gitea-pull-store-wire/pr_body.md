## 카테고리 · 모듈

- 카테고리: `integration-registry` (Gitea SCM pull 정밀화, RM-M4-06)
- 모듈: `internal/store/{repository_pull_ingest,repository_pull_state,repository_pull_targets}` + `internal/integrations/adapters/gitea_pull` + `main` + `migrations/000045`

## Tier

- [x] 사외 (GitHub main)
- [ ] 사내 (사내 SCM)
- [ ] 공용 (양쪽 동기화)

## 사내 한정 정보 self-check (사외 PR 인 경우)

- [x] 사내 env var 미포함 (DEVHUB_GITEA_PULL_* env var 만, 사내 한정 token/URL 미포함)
- [x] 사내 호스트명/IP 미포함
- [x] 사내 한정 경로 변경 없음
- [x] 사내 한정 env 파일 변경 없음

## 요약

X-5 (Gitea Hourly Pull 정밀화, RM-M4-06 잔여, issue #231) 의 1차 PR (`feat/x5-gitea-hourly-pull` PR #592 + `a49a5660` fix) 의 `adapter.Store = nil` + `repoLister` placeholder → production wire 교체. 운영 환경 (pgStore != nil) 에서 `DEVHUB_GITEA_PULL_ENABLED=true` 시 정상 sync (pr_activities / build_runs / quality_snapshots / repository_pull_state 갱신). 단일 PR, ~1311 line.

## 변경 상세

### backend (5 file, +490 line)

1. `backend-core/internal/store/repository_pull_ingest.go` (NEW, ~180 line)
   - `UpsertPullActivity` (pr_activities ON CONFLICT upsert, migration 000001 L402 정합)
   - `UpsertBuildRun` (build_runs ON CONFLICT upsert, migration 000001 L95 정합)
   - `UpsertQualitySnapshot` (quality_snapshots ON CONFLICT partial upsert, migration 000045 정합)
   - repositoryID string → bigint parse (RepositoryPullStore interface signature 정합)

2. `backend-core/internal/store/repository_pull_state.go` (NEW, ~200 line)
   - 6 method: `UpdatePullState` / `IncrementConsecutiveFailures` / `ResetConsecutiveFailures` / `SetBackoff` / `BackoffUntil` / `LastPullAt`
   - repository_pull_state CRUD (migration 000043), cold start 자동 upsert
   - pgx.ErrNoRows → zero time fallback (defensive)

3. `backend-core/internal/store/repository_pull_targets.go` (NEW, ~80 line)
   - `ListGiteaPullTargets` + `GiteaPullTarget` (store-local type, store→adapters import cycle 회피)
   - repositories + integration_providers + repository_pull_state LEFT JOIN
   - filter: provider_type='scm' + provider_key='gitea' + status='active' + gitea_repository_id IS NOT NULL + backoff_until < now()
   - LEFT JOIN 으로 cold start (state row 부재) 시 repo 포함

4. `backend-core/internal/integrations/adapters/gitea_pull.go` (MODIFY, +20 line)
   - `GiteaPullRequest.Merged bool` + `MergedAt *time.Time` field 추가 (Gitea API response 의 `merged` boolean 파싱)
   - `stateToEventType(state, merged) string` helper: state="open" → "opened", state="closed" + merged=true → "merged", state="closed" + merged=false → "closed", else "updated" (defensive)
   - `UpsertPullActivity` 호출 시 `pr.State` 직접 전달 → `stateToEventType(pr.State, pr.Merged)` 로 교체
   - **adapter 책임** 명시: event_type 결정. store 는 CHECK constraint 통과만 보장. **interface 변경 없음**.

5. `backend-core/main.go` (MODIFY, +15 line)
   - `pgStore != nil` 가드 유지 (sqlite/in-memory 환경 fail-fast 정공법)
   - `repoLister` closure 가 `pgStore.ListGiteaPullTargets(ctx)` 호출 + `adapters.RepositoryTarget` 매핑
   - `strconv.FormatInt` 로 `ID`/`ExternalID` (bigint → string) 변환
   - 기존 `onCycle` / `onError` / `go func() { RunGiteaPullLoop(...) }()` 보존

### migration (1 file, 신규)

6. `backend-core/migrations/000045_quality_snapshots_ref_name_unique.{up,down}.sql` (NEW, ~10 line)
   - `CREATE UNIQUE INDEX IF NOT EXISTS quality_snapshots_repo_ref_unique ON public.quality_snapshots (repository_id, ref_name) WHERE tool = 'gitea-build';`
   - partial UNIQUE INDEX (tool 한정) — 다른 tool (e.g. sonarqube) 의 동일 ref_name 허용
   - `UpsertQualitySnapshot` 의 `ON CONFLICT (repository_id, ref_name) WHERE tool = 'gitea-build' DO UPDATE SET ...` 정합

### test (2 file, +230 line)

7. `backend-core/internal/integrations/adapters/gitea_pull_test.go` (MODIFY, +30 line)
   - `TestStateToEventType_Open` / `_ClosedMerged` / `_ClosedNotMerged` / `_UnknownFallback` 4 case

8. `backend-core/internal/store/repository_pull_ops_integration_test.go` (NEW, ~200 line)
   - `TestIntegration_RepositoryPullState_AllMethods` — 6 method 모두 검증 (cold start + Update + Increment 3회 + SetBackoff + Update error + Reset + Increment after reset = 1)
   - `TestIntegration_RepositoryPullIngest_UpsertPRBuildQuality` — 3 method idempotent (pr_activities 1 row + build_runs 1 row + quality_snapshots 1 row)
   - `TestIntegration_ListGiteaPullTargets_Filter` — backoff filtering (test repo appears / disappears after SetBackoff)
   - DEVHUB_TEST_DB_URL 미설정 시 t.Skip (CI backend-unit job)
   - unique repository seed (suffix=UnixNano) — 다른 test 와 격리

### docs (4 file, +60 line)

9. `docs/adr/0034-gitea-hourly-pull-architecture.md` (MODIFY, +30 line)
   - §5.1 production wire follow-up cross-tier 표 7 row (PostgresStore 3 file + migration + adapter + main.go)
   - §6.2 검증 (X-5 follow-up 검증)
   - §8 변경 이력 row (2026-06-15)

10. `docs/planning/2026-06-15-x5-gitea-pull-store-wire-sprint-plan.md` (NEW, ~10KB)
    - §0 컨텍스트 + §1 결정 (file 위치 + state→event_type 매핑 + b.Event + UpsertQualitySnapshot + ListGiteaPullTargets query + main.go wire) + §2 변경 범위 + §3 신규 ID + §4 검증 + §5 잔여 + §6 다음 세션 directive

11. `docs/traceability/report.md` (MODIFY, +1 row) — §6 X-5 follow-up row

12. `docs/llm-wiki/mirror-list.md` (MODIFY) — §1.7.1 의 4 file 추가 (`repository_pull_ingest.go` + `repository_pull_state.go` + `repository_pull_targets.go` + `gitea_pull.go` 갱신) + count 15 → 19 file + §1.7 의 55 → 59 file 갱신

13. `CHANGELOG.md` (MODIFY) — X-5 status `⏳ planned (v0.1.1-alpha)` → `✅ implemented (production wire, 2026-06-15 sprint feat/x5-gitea-pull-store-wire)`

### 메모리 (4 file, NEW)

14-17. `ai-workflow/memory/feat/x5-gitea-pull-store-wire/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-15.md}`

## 추적성 영향

- 추가: `IMPL-GITEA-PULL-STORE-01`, `IMPL-GITEA-PULL-INGEST-01`, `UT-GITEA-PULL-STORE-01`, `TC-GITEA-PULL-STORE-01` (4 row 신규)
- 갱신: `docs/traceability/report.md` §6 (본 row 추가), `docs/adr/0034-...` §5.1 + §6.2 + §8, `CHANGELOG.md` X-5 status
- Deprecate: 없음
- 매트릭스: `docs/traceability/report.md` §6 본 row + `docs/llm-wiki/mirror-list.md` §1.7.1 의 4 file 추가

## 테스트

- [x] 로컬 backend `go build ./...` — silent PASS
- [x] 로컬 backend `go test -count=1 ./internal/integrations/adapters/...` — 12 unit test PASS (기존 8 + 신규 4 stateToEventType)
- [x] 로컬 backend `go test -count=1 ./...` — 30+ packages 회귀 0 (단, httpapi 의 routePermissionTable pre-existing 3 FAIL 은 X-1 잔여, 본 follow-up scope 외)
- [x] `bash scripts/check-openapi-yaml-lint.sh` — PASS (변경 0)
- [ ] CI: backend-unit / backend-integration (DEVHUB_TEST_DB_URL 설정) / e2e skip (path-detect: backend only)
- [ ] 수동 검증: production 환경 (DEVHUB_GITEA_PULL_ENABLED=true + pgStore != nil) 에서 1h cycle 동작 — **사내 staging Gitea SOP 별도 follow-up**

## 관련 issue / ADR

- issue: [#231](https://github.com/ykylee/Devhub_example/issues/231) — Gitea API → DevHub pr_activities / build_runs / quality_snapshots 비동기 sync (v2 → v0.1.1 진입)
- ADR: [0034-gitea-hourly-pull-architecture.md](docs/adr/0034-gitea-hourly-pull-architecture.md) (X-5 1차 + 본 follow-up §5.1)
- sprint plan: [2026-06-15-x5-gitea-pull-store-wire-sprint-plan.md](docs/planning/2026-06-15-x5-gitea-pull-store-wire-sprint-plan.md)
- X-5 1차 PR: PR #592 + `a49a5660` fix (`feat/x5-gitea-hourly-pull`)

## 잔여 follow-up (본 PR scope 외, 사용자 결정 영역)

- **staging Gitea 실 검증 SOP** (사내) — production wire 후 1h cycle 실 검증.
- **e2e spec (staging Gitea mock + 1h cycle)** (사내).
- **frontend widget (Gitea pull 상태, X-1 admin dashboard 통합)** (사외 가능).
- **PR/commit status mapping 정밀화** (Gitea commit SHA → DevHub quality_snapshots 매핑 정밀화, 본 follow-up 의 1차 매핑 + 후속).
- **repository_pull_state.last_alert_at** (5회 연속 실패 시 alert emit 시각, ADR-0034 §3.1 의 향후 follow-up).
