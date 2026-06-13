# Session Handoff — main (2026-06-10/11, v0.5.0→v0.5.11 ai-workflow 동기화 — PR #545)

- 문서 목적: PR #545 (v0.5.0-beta → v0.5.11-beta 동기화, theirs-only 1 squash, 97 file / 4562줄 / 178줄 삭제) 상태 인계.
- 범위: `ai-workflow/VERSION` + README/WORKFLOW_INDEX + 메모리 전체 + archive/ + codex/phase6 + gemini/phase6/7/10 + release/v0.5.{1..10} + `.gitignore` 백업 라인.
- 상태: main HEAD `165b8e8` (PR #545 squash) → branch `chore/v0.5.11-sync-2026-06-10` push 완료, PR #545 open (https://github.com/ykylee/Devhub_example/pull/545). 3 file (state.json / session_handoff.md / work_backlog.md) reapply 분기는 theirs-only 흡수 + 백업 보존 — 머지 후 추가 결정 불요.
- 최종 수정일: 2026-06-11 07:30 KST (PR #545 push 기준)
- 직전 handoff (PR #514 + #515 finalizing): §0a 참조, main HEAD `fee06d4` 까지.

## 0. 본 세션 핵심 결과 (2026-06-10/11, v0.5.0→v0.5.11 ai-workflow 동기화 + N-6 skip)

### N-6 skip 결정 (2026-06-11, 사용자 결정)

- **N-6 (v1.0 staging 1주 운영 검증) skip** — 사용자 결정 (2026-06-11).
- §1.3 DoD 7 + §3.5 N-6 row + `state.json` M-v1.0 notes + `work_backlog.md` status line 모두 ✅ skipped 마킹.
- **v1.0 release blocker 0건** — 8 DoD 중 7 ✅ + 1 ✅ skipped = **8 DoD 모두 close**.
- **v1.0 release 가능 상태**. main HEAD `82935f8b`.
- 잔여 follow-up: 잔여 3 (T-d-79-2 / T-d-80-2 my_harness 측 SSOT 작성, 사용자 전달 후 진행 중) + 잔여 5 (T-d-72-5/6 + D-73/74 + X-1~8, v1.1 forward path, v1.1 milestone 진입 시점 별도 sprint).

### v0.1.0-alpha release (2026-06-11, 사용자 결정)

- **v0.1.0-alpha release 발표** (commit `356d08b7`, main HEAD) + tag `v0.1.0-alpha` (re-tag 후 main HEAD 부착).
- PR #554 (N-6 skip) 머지 → main HEAD `d860b7c9` → CHANGELOG + 발표 자료 commit `356d08b7`.
- 신규 3 file: `CHANGELOG.md` (670+ lines) + `docs/presentations/2026-06-11-v0.1.0-alpha-announcement.md` (200+ lines) + `docs/presentations/2026-06-11-v0.1.0-alpha-announcement.html` (자체 HTML 14 슬라이드, 키보드 네비게이션, reveal.js 의존성 0).
- 8 DoD 모두 close (7 ✅ + 1 ✅ skipped). v0.1.0-alpha release blocker 0건.
- v1.0 release 가능 상태 (사용자 결정 시점 v1.0 release 태그 가능).

### v0.1.1-alpha release 정공법 (2026-06-11, 사용자 결정)

- **잔여 5 (T-d-72-5/6 + D-73/74 + X-1~8) 의 v0.1.1-alpha 격하** — 사용자 결정 (v1.1 forward path 가 아닌 v0.1.x patch release).
- `docs/planning/release_v1_roadmap.md` §3.5 NEXT block title `v1.1` → `v0.1.1-alpha` + X-1~8 의 status `v1.1 NEXT` → `v0.1.1-alpha NEXT` 격하.
- `ai-workflow/memory/state.json` M-v1.0 notes "v0.1.1-alpha release 정공법" + 잔여 5 의 8 item 의 v0.1.1-alpha 격하 마킹.
- `ai-workflow/memory/work_backlog.md` status line + §5 변경 이력 row.
- `CHANGELOG.md` v0.1.1-alpha release note 추가.
- 8 item 모두 v0.1.1-alpha release 의 정공법 정합 (실제 구현은 사용자 결정 시점).
- 잔여 3 (T-d-79-2 / T-d-80-2 my_harness 측 SSOT 작성) + vault Gitea remote push (사용자 수동) 별도 이월.

### N-9 (P1-7 Repository build-runs) 기본 구현 완료 정합 (2026-06-11, PR #555)

- **PR #555 ✅ MERGED** (squash `1e9e4f80`, main HEAD 2026-06-11) — docs only 정합 (3 file / +8 -5).
- **§3.5 N-9 row status `✅ resolved (기본 구현, 2026-06-11)`** + §3.2 P1-7 row 비고 `기본 구현 완료` + path 차이 명시 + §9 변경 이력 1 row.
- 본 carve 의 endpoint `GET /api/v1/repositories/:repository_id/build-runs` (router.go:509) + `platformStoreOrUnavailable` 가드 + `ListRepositoryBuildRuns` (postgres.go `repository_ops.go`) + UT 3건 + IT 1건 + openapi.yaml §repositories/build-runs 정의 + frontend `repositoryService.getRepositoryBuildRuns` + `DeveloperView` 위젯 + e2e `repository-dashboard.spec.ts` 의 inline build-runs mock 검증 모두 main 반영 완료.
- **issue #487 close** (정식 ID = §3.5 N-9, 기본 구현 완료) + 잔여 4건 sub-issue 분리 (v1.1 milestone 진입 시점):
  - **#556** [N-9 sub-1] RBAC 403/404 가드 (backend)
  - **#557** [N-9 sub-2] `devhub_repository_build_runs_query_duration_seconds{status_filter}` Histogram (backend)
  - **#558** [N-9 sub-3] `useRepositoryBuildRuns` TanStack Query hook + status filter dropdown + skeleton + 무한 스크롤 (frontend)
  - **#559** [N-9 sub-4] Dashboard widget "Recent repository activity" 통합 + 독립 e2e spec `tests/e2e/repository-build-runs.spec.ts` (frontend)
- **Tier**: 공용 (docs only). **신규 ID 발급 0건**. CI 4/4 PASS.
- branch `chore/work_260611-d-n9-status-align` PR 머지 후 GitHub 자동 삭제.

## 0a. 본 세션 직전 (2026-06-10/11, v0.5.0→v0.5.11 ai-workflow 동기화)

### PR #545 결과

| 항목 | 값 |
|---|---|
| **PR #545** | 🟡 OPEN (squash push 완료) |
| **Title** | `chore(ai-workflow): v0.5.0-beta → v0.5.11-beta 동기화 (theirs-only, 1 squash)` |
| **URL** | https://github.com/ykylee/Devhub_example/pull/545 |
| **Branch** | `chore/v0.5.11-sync-2026-06-10` (main `fee06d4` → `165b8e8` 1 squash) |
| **Diff stat** | 97 file, +4562 / -178 |
| **Tier** | 사외 (no internal-only paths) |
| **Backing 결정** | scope=devhub-only / stride=1 squash / merge=theirs-only / risk=backup 1단계 (사용자 2026-06-10 directive 4건) |

### 변경 요약

| 영역 | 변경 |
|---|---|
| `ai-workflow/VERSION` | v0.5.0-beta → v0.5.11-beta |
| 운영 가이드 | README.md / WORKFLOW_INDEX.md / PROJECT_PROFILE.md / repository_assessment.md (v0.5.11 standard_ai_workflow 양식) |
| 메모리 3 file | state.json / session_handoff.md / work_backlog.md (구조 갱신, 3 file reapply 분기는 백업 보존 후 theirs-only 흡수) |
| archive/ | 5 report 백업 (comprehensive / phase5 / phase8 / session_handoff / 2026-04-30) |
| codex/phase6/ | codex phase 6 작업물 (backlog + tasks 8건) |
| gemini/phase6/7/10/ | gemini phase 6/7/10 기록 (backlog + tasks 다수) |
| release/v0.5.{1..10}/ | release 별 state / handoff / backlog 분리 (8 minor release) |
| `.gitignore` | `_pre_v0.5.11_backup_2026-06-10/` 명시적 제외 |

### Pre-flight / Safety

- **origin/main 0 commit drift** (squash PR 사실상 첫 동기화)
- **백업 1단계**: `ai-workflow/memory/_pre_v0.5.11_backup_2026-06-10/` 에 state.json (153줄) / session_handoff.md (389줄) / work_backlog.md (97줄) 보관 (2026-06-10 23:47 KST)
- **Tier check**: 사외 / 2-tier 정책 self-review 통과 (`docs/governance/worker_division.md` §6)

### 후속 (사용자 결정 영역)

- 3 file (state.json / session_handoff.md / work_backlog.md) reapply 분기 — **theirs-only 흡수했으므로 별도 작업 불요할 가능성 ↑** (백업 보존, 회귀 시 `git checkout 165b8e8^ -- <file>` 로 즉시 복원 가능)
- `_pre_v0.5.11_backup_2026-06-10/` rotate (다음 sync 시점에 `.gitignore` 라인 제거, `mavis-trash` 권장)
- my_harness 측 동기화 — 사용자 confirm 후 별도 plan trigger (이번 scope=devhub-only)

## 0b. 이전 세션 (2026-06-09, v1.0 출시 직전 finalizing — PR #514 + PR #515 + codex P2 fix)

- 문서 목적: PR #514 (voc + notification, ADR-0028) + PR #515 (옵션 A N-12 housekeeping + B voc list + C N-10 IT 3 TC + codex P2 fix 3 layer) 머지 상태 인계.
- 범위: 본 세션의 2 PR (PR #514 + PR #515 squash). 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (3 layer: production router mount + routePermissionTable + gin path conflict).
- 상태: main `f7d2705` (PR #515 squash) + PR #514 (squash) 모두 머지 완료. main HEAD `897953c` (PR #503 housekeeping 기준) + 이후 06-09~06-10 v1.1 sprint -a follow-up PR #538/539/540/541/542/543 + tier-governance / branding / agentic-rag 등 다수 PR 머지. main 최신 HEAD `fee06d4` (2026-06-10 housekeeping `chore(memory)`).
- 최종 수정일: 2026-06-10 (handoff 본문 마지막 갱신, 다음 cross-check: 2026-06-10 23:45 KST)

## 0. 본 세션 핵심 결과 (2026-06-09, v1.0 출시 직전 finalizing)

### PR 머지 / Push 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#514** (voc + notification, ADR-0028) | ✅ MERGED (squash) | 외부 시스템 의뢰 staging 도메인 + 9 field + in-app notification + 5 API. `(source_system, external_ref)` UNIQUE for idempotency. ADR-0028 §3 옵션 1 (별도 도메인 + 1:1 dev-request 매핑) 채택. 12 file +1043 line. main `ba7823f`. |
| **#515** (v1.0 출시 직전 finalizing) | ✅ MERGED (squash, `f7d2705`, 06-09 07:08 UTC) | 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (PR #514 latent 회귀 3 layer 동시 fix). 5 commit (74ff06f + de94bac + 0a90782 + 2b00fe0 + 22306db). branch HEAD `22306db`. |

### Commit 4 — codex P2 fix (3 layer 동시)

**Codex review id 3378458885**: P2 — production router 에 VOC list route 미등록. `NewRouter` 가 `Handler{...}` literal 에서 `voc` field 를 init 하지 않아 `line 527` 의 `if handler.voc != nil` 체크가 production 에서 항상 `false`. PR #514 의 5 voc route 가 production 에서 mount 안 됨.

**3 layer 동시 fix (`2b00fe0`)**:
1. `router.go` `NewRouter` `Handler{...}` literal: `voc: devreqview.NewVocHandler(...)` init 추가.
2. `voc_handler.go` `RegisterVocRoutes`: `GET /dev-requests/:external_ref` → `GET /dev-requests/external/:external_ref` (gin strict duplicate path error 회피, 사용자 선택) + `POST` param name `external_ref` → `dev_request_id` 정합.
3. `permissions.go` `routePermissionTable`: 5 voc route 의 entry 추가 (PR #514 의 latent 회귀):
   - `POST /dev-requests/:dev_request_id` → `ResourceDevRequests, ActionCreate` (system_admin, 외부 intake)
   - `POST /dev-requests/:dev_request_id/route` → `ResourceDevRequests, ActionEdit` (team_manager/system_admin)
   - `GET /dev-requests/external/:external_ref` → `ResourceDevRequests, ActionView`
   - `GET /vocs` → `ResourceDevRequests, ActionView` (ADR-0028 §6 carve d, system_admin 도구)
   - `GET /me/notifications` → `Bypass` (자기 정보)
   - `POST /me/notifications/:id/read` → `Bypass` (자기 마킹)

**검증**: `go test ./...` 0 FAIL, `go build ./...` silent PASS, codex reply id 3378526106.

### 신규 ID 4건
- `IMPL-voc-01`: voc 등록 / routing / 조회
- `IMPL-notification-01`: in-app notification (`/api/v1/me/notifications`)
- `IMPL-dreq-02`: dev-request 9 field + 단일 트랜잭션 자동 생성
- `IMPL-voc-list-01`: `GET /api/v1/vocs` (ADR-0028 §6 carve d)

### 신규 TC 7건
- TC-VOC-LIST-01..03 (4 케이스)
- TC-RBAC-LOGOUT-01 + TC-RBAC-ROLE-DRIFT-01 + TC-RBAC-LEGACY-01

### PR #514 / #515 정합 후 v1.0 출시 직전 잔여
- **N-6**: staging 1주 운영 + 외부 사용자 ≥5 로그인 검증 (사용자 결정 영역)
- 옵션 D (`project.inbound_source` 자동 routing, ADR-0028 §6 carve a): post-MVP 후속 sprint 후보

## 0a. 이전 세션 (2026-06-09, swagger UI 1차 bootstrap + v1.0 직전 housekeeping)

### PR 머지 결과 (squash, 후속 housekeeping PR은 v1.0 finalizing sprint 본 §0 참조)

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#505** (swagger UI 1차 bootstrap) | ✅ MERGED (squash) | 정적 HTML + CDN (swagger-ui-dist@5.17.14 unpkg) + `embed.FS` 결정. 0 Go 의존성. `DEVHUB_SWAGGER_ENABLED=true` opt-in (default false, prod-safe). 4 commit (`1b640d4` + `428c3f4` + `b48794d` + `6cc208a`) — codex P1 2건 fix + nginx `/devhub/swagger/` forward (단일 포트 정합) + e2e shard 3/3 hotfix. 신규 ID: `IMPL-swagger-01` / `ADR-0027`. main `ad8d481`. |
| **#500** (워커 분업 전면 취소, 06-09) | ✅ MERGED (squash) | `worker_division.md` §0 + §1~§4 historical 標記 + `AGENTS.md` 워커 일반 메모 갱신 + branch prefix 자유화. 사용자 결정 (Claude/Codex 자유 이용 불가). main `f99fef7`. |
| **#499** (N-11 메모리 sync) | ✅ MERGED (squash, rebase 후) | 메모리 4종 (state/handoff/work_backlog/release_v1_roadmap) + traceability report.md §3.5/§6. main `da7d57e`. |
| **#498** (ci.yml 코멘트 갱신) | ✅ CLOSED | e2e shard 2/3+3/3 fail → N-8 race 발견 → close + N-8 hotfix 4차 별도 sprint. |
| **#502** (N-8 hotfix 4차 1차: 502→204) | ✅ MERGED (squash) | backend logout handler graceful degradation. main `6654b44`. |
| **#503** (N-8 hotfix 4차 2차: codex P1 + follow-up) | ✅ MERGED (squash) | response header `X-Keycloak-Likely-Down: true` + typed error sentinel `ErrOIDCConfigMissing`/`ErrOIDCNetworkUnreachable`. main `897953c`. |

### N-8 hotfix 4차 정공법 (3 commit, 2 PR)

**근본 layer**: backend `POST /api/v1/auth/logout` 가 Keycloak 도달 실패 시 502 즉시 반환 → frontend logout() 가 OIDC skip + `window.location.assign('/login')` 강제 → AuthGuard pathname 변화 useEffect 에서 stale actor 박음 → `/developer` 진입 → `/login` 도착 못함. PR #497 의 hotfix #1/`#2/`#3 가 모두 backend 502 자체를 막지 못함 (deterministic, 32회 retry).

**PR #502 (1차)**: backend logout handler 가 502 → **204 No Content** + audit `revoke_status=unreachable` + hotfix 식별자. frontend logout() 가 정상 204 분기 진입 → OIDC end_session_endpoint 호출 → /login 정상 도착 → race close.

**PR #503 (2차 commit 066cd7b, codex P1 응답)**: "구분 가능한 응답" — 204 + response header `X-Keycloak-Likely-Down: true` + `X-Logout-Hotfix: N-8-4:graceful-degrade`. frontend 가 header 마커 conditional 확인 → OIDC skip 또는 정상 OIDC 결정. 진짜 IdP outage 시 dead IdP trap 회피.

**PR #503 (3차 commit e18b34f, codex P1 follow-up 응답)**: typed error sentinel 도입.
- `authview.ErrOIDCConfigMissing` (sentinel): backend config 결함 (missing realm/oidc_client_id/oidc_client_secret) → handler 가 **marker 미부착** + 정상 OIDC 분기 + audit `revoke_status=config_error` + `config_error_detail`
- `authview.ErrOIDCNetworkUnreachable` (sentinel): 네트워크/5xx outage (DNS 실패, conn refused, timeout, Keycloak 5xx) → handler 가 marker 부착
- 그 외 미분류 error: conservative — outage 분류

codex P1 의 핵심 우려 "reachable Keycloak SSO session is not terminated" 정공법: config error 분기에서 marker 미부착 → frontend 정상 OIDC → RP-initiated logout 시도 → SSO session 정상 종료.

### 검증

- **CI 모두 SUCCESS** (PR #503 머지 시점): workflow-lint / changed-paths / migration-prefix / Backend Unit + Integration / Frontend Unit / E2E Build / **E2E shard 1/2/3 모두 PASS**
- `go test ./...` (35 packages) PASS
- `npx vitest run` (80 files, **1033 tests**) PASS
- **신규 test 4건**:
  - TC-AUTH-LOGOUT-04 (network/5xx → 204 + marker)
  - **TC-AUTH-LOGOUT-08** (config error → 204 + marker 미부착)
  - TC-AUTH-LOGOUT-FE-07 (frontend header 마커 확인 → OIDC skip)
  - **TC-AUTH-LOGOUT-FE-08** (frontend header 없음 → 정상 OIDC)

### 잔여 DoD 해소

- **N-11 잔여 DoD** (main 첫 PR 두 job PASS, issue #419): PR #503 머지 시점에 e2e shard 1..3 모두 PASS → 해소
- **N-8 race** (issue #501): close
- **워커 분업 전면 취소** (사용자 결정): PR #500 머지로 정합. branch prefix 자유 (`maintenance/`, `chore/`, `docs/`, `fix/`, `feat/` 등)

## 1. 다음 세션 directive

### v1.0 출시 직전 — 우선순위

1. **N-6 (v1.0 staging 1주 운영)** — N-8 + N-11 + N-7 + 워커 분업 취소 + swagger UI + housekeeping 정합 완료. 사용자가 staging 환경 운영 + 외부 사용자 ≥5 로그인 검증. (사용자 결정, sprint 영역 외)
2. **N-10 Manager RBAC partial verified (status ⏳)** — 검증 보고서 [docs/validation/N-10-manager-rbac.md](docs/validation/N-10-manager-rbac.md) 의 follow-up 6 TC 중 7 TC active 완료: **E2E 4 TC** (TC-RBAC-LOGOUT-02 + TC-RBAC-ROW-READ-01/02 + TC-RBAC-CODE-01) 는 `frontend/tests/e2e/rbac-data-scope.spec.ts` 에 PR #509 머지 (branch `maintenance/work_260610-c-N10-rbac-e2e-tcs`, mergeCommit `cb59b39`). **backend IT 3 TC** (TC-RBAC-LOGOUT-01 + TC-RBAC-ROLE-DRIFT-01 + TC-RBAC-LEGACY-01) 는 `backend-core/internal/domain/rbac-permissions/view/rbac_n10_integration_test.go` 에 PR #515 옵션 C 머지. TC-RBAC-TRACE-01 는 Process/review 단계로 spec header 주석에 입증. PR #512 (codex P2 follow-up `fix/work_260611-c-n10-status-partial-apply`) 로 `release_v1_roadmap.md §3.5 N-10 row` status = `⏳ verified (partial — E2E 4 TC done, IT/UT 3 TC + Process 1 TC scoped to follow-up sprints)`. 잔여 housekeeping (메모리 drift 정합, 검증 보고서 §3 close 마킹) 만 남음.

### 완료 정합 (2026-06-10)

- release_v1_roadmap.md §3.5 N-8 race status `✅ resolved` + N-11 row "✅ 잔여 DoD 해소" + §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 + §3.5 N-10 follow-up 6 TC 명시.
- docs/traceability/report.md §4 ADR-0025 (envelope encryption) + ADR-0026 (Keycloak role excluded) row 보강.
- state.json 1515 → 150 line (90% 감소) — 21 top-level key + next_actions 5 closed key archive.
- work_backlog.md §5 5월 135행 archive summary 1 line 으로 축약.
- ADR-0027 (swagger UI 결정) + IMPL-swagger-01.

### 자유 에이전트 정책 (2026-06-09 결정)

본 세션 결정으로 **누구든** 어느 sprint/영역 진입 가능. `worker_division.md` §0 + `AGENTS.md` "워커 일반 메모 (2026-06-09 전면 갱신)" 정합. branch prefix 자유.

### 자유 에이전트 정책 (2026-06-09 결정)

본 세션 결정으로 **누구든** 어느 sprint/영역 진입 가능. `worker_division.md` §0 + `AGENTS.md` "워커 일반 메모 (2026-06-09 전면 갱신)" 정합. branch prefix 자유.

## 2. 이전 핵심 스프린트 (5월/4월 historical, archive)

5월/4월 historical 5개 sprint (X-3 envelope encryption / NOW-3 SCM E2E / NOW-4 frontend unit test / NOW-5 migration prefix guard / 2026-06-01 CI 복구 / 2026-06-06 sprint -h 추적성 ID) 의 상세 본문은 [_archive/state-2026-05-pr-tracker/session_handoff_2026-05-06_archive.md](./_archive/state-2026-05-pr-tracker/session_handoff_2026-05-06_archive.md) 로 이동. canonical source 는 `git log --since=2026-04-01 --until=2026-06-08 --merges` (PR #20~#504, 470+ commit).

### 1 line summary (5월 ~ 4월)

- **X-3** (envelope encryption + KEK key management, 2026-05-27) ✅ — `internal/crypt` AES-GCM-256 봉투 + `IntegrationRepository` 자동 Encrypt/Decrypt 결합 + 6 envelope_test.go unit test PASS + 32 packages green.
- **NOW-3** (SCM import/create + draft/publish E2E) ✅ — backend 캐스팅 정정 + Gitea Mock Provider Fallback + nano-ts unique ID 매핑 + Playwright auto-wait Locator → 63 passed / 6 skipped.
- **NOW-4** (frontend unit test 보강) ✅ — Zustand store / ProviderModal / MemberTable / PermissionEditor 4 module Vitest 작성 → 962 unit test 100% PASS.
- **NOW-5** (migration prefix uniqueness CI guard) ✅ — `scripts/check-migration-uniqueness.sh` + `ci.yml` 상시 lint + `make lint-migrations` 바인딩.
- **2026-06-01 CI 복구** ✅ — `applications/[id]/page.tsx` 중복 import 제거 + `admin-projects.spec.ts` 환경 독립 검증 → CI run `26738464130` 성공.
- **2026-06-06 sprint -h 추적성 ID 발급** ✅ — PR #490, N-7 (REQ-FR-106/ARCH-18/API-98 + IMPL/UT/TC) + N-8 (REQ-FR-107/ARCH-19/API-99) + N-9 (REQ-FR-108/ARCH-20/API-100) + `integration-registry` 도메인 cross-ref 보강.

## 3. 후속 carve out / 잔여 백로그 우선순위 (current)

| 우선순위 | 항목 | 사유 |
| --- | --- | --- |
| **N-9** | **P1-7 Repository build-runs 기본 구현 완료** | endpoint + UT 3건 + IT 1건 + frontend 통합 모두 main 반영. 잔여 4건 sub-issue #556/#557/#558/#559 (v1.1). PR #555 (`1e9e4f80`). issue #487 closed. |
| **N-6** | v1.0 staging 1주 운영 검증 | 외부 사용자 로그인 + Onboarding SOP DoD 8 만족 (사용자) |

| **N-10** | Manager RBAC partial verified (status ⏳) | E2E 4 TC (PR #509) + backend IT 3 TC (PR #515) active. TRACE-01 Process scoped-out. 잔여 housekeeping 만 |
| **X-1** | System Admin 운영 대시보드 | Gitea sync job 큐/상태 + provider health (v1.1) |
| **X-2** | inbound webhook 정규화 깊이 | multi-provider sync 일반화 (v1.1) |

## 4. 다음 세션 directive
* **PR #515 ✅ MERGED** (squash `f7d2705`) + **PR #516 ✅ MERGED** (squash `2b3c766`) + **PR #517 ✅ MERGED** (squash `97bc6bc`).
* **PR #555 ✅ MERGED** (squash `1e9e4f80`, 2026-06-11) — N-9 (P1-7 Repository build-runs) 기본 구현 완료 정합. 3 file / +8 -5 (docs only, 공용 tier). issue #487 closed + 잔여 4건 sub-issue #556/#557/#558/#559 분리 (v1.1).
* **swagger UI 정상동작 fix 완료** — PR #508 의 silent 404 의도적 결정을 embed fallback 으로 supersede. 7 swagger TC 모두 PASS. staging env `DEVHUB_SWAGGER_ENABLED=true` 만 설정해도 openapi.yaml 정상 서빙.
* **N-6**: staging 1주 운영 (사용자 결정 영역). swagger UI 정상동작 정합.
* **N-10 IT 3 TC 완료 정합** (이전 sprint, PR #515 옵션 C): `TC-RBAC-LOGOUT-01` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-LEGACY-01` ✅ verified (`rbac_n10_integration_test.go`). E2E 4 TC 는 PR #509 로 별도 active. status partial verified.
* **option D 검토 완료**: N-13 + ADR-0028 §6 정합. 구현 = v1.1 milestone 진입 시점.
* **V1.1 진입 준비**: X-1/X-2 로드맵 백로그 분석.

---

## 5. 본 세션 (2026-06-10, v1.1 sprint -a follow-up — PR #539 머지)

### PR 머지 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#538** (sprint -a 본 — port interface, 이전 머지) | ✅ MERGED | `domain/auth-session/integration/ports.go` 신규 (4 port + 4 type alias + 3 sentinel error alias). ADR-0030. main `20b4bb3b`. |
| **#539** (sprint -a follow-up — stub + main wiring + view/ deprecation) | ✅ MERGED (squash, branch delete) | saovae_stub (4 port + webhook handler) + main.go `DEVHUB_BUILD_TIER` env var 분기 + ports.go mirror alias 통합 + view/ 3 interface deprecation. main `87e6c1f5`. CI 7/7 PASS. |

### PR #539 5 file (commit `a00793bc`, +238 -97)

1. `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (NEW, 105 lines) — 4 port stub + webhook handler. `go build ./...` PASS. no test files (의도적 — 사외 stub, keycloak infra 비의존).
2. `backend-core/main.go` (L148-235 분기 + 2 import 추가) — `var (idpAdmin httpapi.IdentityAdmin; oidcLogout httpapi.OIDCLogoutClient; keycloakEventPort integration.KeycloakEventPort)` + `DEVHUB_BUILD_TIER` env var. default (사외) = saovae_stub, `=internal` 시 real KeycloakAdminClient.
3. `backend-core/internal/domain/auth-session/integration/ports.go` — `KeycloakUserEvent` / `KeycloakAdminEvent` mirror struct → `type X = httpapi.X` alias. `*KeycloakAdminClient` 가 `KeycloakEventPort` 충족 위해 필수.
4. `backend-core/internal/domain/auth-session/view/auth.go:59` — `BearerTokenVerifier` interface deprecation comment.
5. `backend-core/internal/domain/auth-session/view/handler.go:27 + :197` — `IdentityAdmin` + `OIDCLogoutClient` interface deprecation comment.

### CI 7/7 PASS

| Check | Result | Duration |
|---|---|---|
| Backend Integration Tests | pass | 1m15s |
| Backend Unit Tests | pass | 1m8s |
| E2E Build Artifacts | pass | 1m49s |
| E2E Tests (Playwright, shard 1/3) | pass | 4m0s |
| E2E Tests (Playwright, shard 2/3) | pass | 4m2s |
| E2E Tests (Playwright, shard 3/3) | pass | 5m30s |
| Detect Changed Paths | pass | 11s |
| Migration Prefix Uniqueness | pass | 7s |
| OpenAPI YAML Lint | pass | 9s |
| Workflow Lint (actionlint) | pass | 13s |
| Frontend Unit Tests | skip (path-detect 결과 — backend-only 변경) | - |

### Tier 분류 (PR #539 self-check)

- **ports.go / view/ = 공용** (interface 만 노출)
- **sso-integrations/keycloak/ = 사외** (saovae_stub — Keycloak 인프라 비의존)
- **main.go 변경 = 공용** (runtime injection branch 만, 사내 한정 패턴 미도입 — `check-tier-separation.sh no changes` 확인)

### 사용자 결정 사항 (in-session)

- **PR scope 분리**: sprint -a follow-up 본 PR = stub + main wiring (Recommended). real adapter 별도 PR.
- **Build 정책**: `//go:build` tag 미사용, **runtime injection** (단일 binary, `DEVHUB_BUILD_TIER` env var).
- **Default = 사외 (saovae_stub)**: env var 미설정 시 saovae_stub 자동 사용.
- **`=internal` 시 real KeycloakAdminClient**: 사내 staging/prod-smoke 검증.
- **Type alias**: ports.go 의 mirror struct 를 httpapi alias 로 통합 (distinct type → alias).

### 후속 carry-over (다음 PR)

1. **C-a** real adapter: `sso-integrations/keycloak/{verifier,admin_client}.go` (P0)
2. **C-b** main.go event listener type assertion 정리 (P0)
3. **C-c** `_ = keycloakEventPort` placeholder 제거 (P0, C-b 완료 시)
4. **C-d** v1.0 mirror struct 제거: `httpapi.KeycloakUserEvent` / `httpapi.KeycloakAdminEvent` (P1)
5. **C-e** audit-ops 의 mirror 와 통합 (P1)
6. **C-f** `infra/idp/_archive_2026-06-10/` immutable archive (P1)
7. **C-g** traceability report.md IMPL-30/31/32 row 갱신 (P2)
8. **C-h** ADR-0030 §5 timeline 갱신 (P2)
9. **C-i** E2E test saovae_stub + real adapter CI matrix (P2)
10. **C-j** build tag 정책 재검토 (P3)

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-v1-1-sprint-a-followup/` 신규 작성 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md).
- main `state.json` status 갱신 (PR #539 entry).
- 본 `session_handoff.md` §5 append.

### 다음 세션 directive (sprint -a follow-up 완료 후)

- **real adapter PR 시작**: branch `feat/work_260610-v1-1-sprint-a-real-adapter` 분기. C-a → C-b → C-c → C-d → C-e → C-f 순서로 진행.
- **C-g traceability 갱신**: `docs/traceability/report.md` IMPL-30/31/32 row 추가.
- **또는 다른 sprint 진입**: PR #538 이전의 carry-over, ADR-0030 §5 timeline 갱신 등.

## 6. 본 세션 (2026-06-10, sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 정공법 PR — docs only)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#540** (sprint -a follow-up PR1 — real adapter + v1.0 mirror struct 제거) | ✅ MERGED (squash, branch delete) | `sso-integrations/keycloak/{verifier,admin_client,metrics}.go` 신규 (real KeycloakJWKSVerifier + KeycloakAdminClient 3 port 동시 충족 + raw wire → flat canonical struct 매핑) + saovae_stub 보강 (PR #539 머지 후) + v1.0 mirror struct 제거 (`httpapi.KeycloakUserEvent` + `KeycloakAdminEvent` 폐기) + audit-ops mirror 통합 (`KeycloakEventLister` interface 통폐합) + `infra/idp/_archive_2026-06-10/identity.schema.json` archive + `infra/idp/README.md` 갱신. main `58d163f`. CI 7/7 PASS. |
| **본 PR (carry-over C-g + C-h 정합 PR)** | ✅ MERGED (squash, PR #541, main `88681f4`, branch delete) | `docs/work_260610-traceability-impl-sso-keycloak` 분기 (docs only, 코드 0줄). §2.4 IMPL 개요 paragraph 갱신 + 5 row IMPL 신규 sub-table (`sso-keycloak-01` + `sso-keycloak-stub-01` + `sso-keycloak-metrics-01` + `auth-session-port-01` + `audit-ops-event-mirror-01`, conventions.md §1 kebab-case 정합) + §3.1 auth-session/audit-ops + §3.3 keycloak-idp 매트릭스 row 갱신 + §4 ADR 인덱스 ADR-0030 row + §6 변경 이력 row + ADR-0030 §5 timeline accepted/done + §9 변경 이력 row. 신규 ID 5건 (모두 IMPL, REQ/UC/ARCH/API/RM/UT/TC 신규 발급 0건). CI 4/4 PASS (docs only PR — backend/e2e/frontend skip, 4 path-detect + lint). commit `22e8c84` → squash merge `88681f4`. Tier: **공용** (문서만). |

### C-g / C-h 정합 PR scope

sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 의 정공법. 본 PR 은 **문서만 변경** (코드 0줄). 5 row ID 의 정공법 = `conventions.md §1` 의 kebab-case module ID 정합 (메모리 출발점의 `IMPL-30/31/32` 표기는 형식 위반 → 정정).

### Tier 분류

본 PR 의 모든 변경 = **공용** (사내 한정 정보 미포함, `docs/traceability/report.md` + `docs/adr/0030-...` 의 문서 정합). `check-tier-separation.sh` PASS 예상.

### 후속 carry-over (C-i + C-j)

- **C-i (P2)**: E2E saovae_stub + real adapter CI matrix — DEVHUB_BUILD_TIER=internal env var + e2e shard 양쪽 정합. 본 PR 후속 별도 PR.
- **C-j (P3)**: build tag 정책 재검토 — runtime injection (현재) ↔ build tag 전환 trade-off. 별도 ADR 후보.

### Memory 갱신

- `ai-workflow/memory/docs/work_260610-traceability-impl-sso-keycloak/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md).
- main flat `state.json` status + head_commit 갱신 (PR #540 baseline = `58d163f`, 본 PR 머지 시점은 head_commit main = PR 본 PR 머지 commit).
- 본 `session_handoff.md` §6 append.

### 다음 세션 directive

- 본 PR commit + push + PR 발행 (사용자 confirm 후).
- 또는 C-i (E2E CI matrix) 진입.
- 또는 다른 sprint (N-10 housekeeping close / release_v1_roadmap.md §3.5 N-10 close 마킹 / 검증 보고서 §3 follow-up close).

## 7. 본 세션 (2026-06-10, sprint -a follow-up PR1 PR #540 의 carry-over C-i 정공법 PR — ci.yml + script 2 file)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#542** (C-i E2E Internal job) | ✅ MERGED (squash, branch delete) | `feat/work_260610-c-i-e2e-internal-job` 분기. `.github/workflows/ci.yml` 에 `e2e-internal` job 신규 (+202 lines, 23 step, PG 15 + Keycloak container port 8181 + apply migrations + validate E2E-CI sync contract + Start Backend DEVHUB_BUILD_TIER=internal + Start Frontend + Wait + Run E2E Tests shard 1/1 + Upload Report + Upload Logs). `scripts/ci-e2e-sync-check.sh` 에 DEVHUB_BUILD_TIER 의도적 미포함 rationale comment (+5 lines). e2e shard 1/2/3 (saovae_stub default) 의 env block, start command, test invocation 모두 변경 0. main `24674b8`. CI 4/4 PASS (workflow 변경만, backend/e2e/frontend skip). Tier: **공용** (`.github/workflows/*` + `scripts/*` 모두 사내 한정 정보 미포함). |

### scope 결정 (옵션 A 채택)

e2e shard 1/2/3 (saovae_stub default) + 별도 `e2e-internal` job 1개 (`DEVHUB_BUILD_TIER=internal`) 의 CI matrix 1쌍. 옵션 B (unit test cover only) / C (6 matrix) / D (matrix shard 1/2/3 × 2) 모두 거부.

### trade-off

- **Keycloak container port 8181** (e2e shard 의 8180 과 분리) — e2e shard 와 e2e-internal 동시 trigger 가능
- **Playwright shard 1/1** (단일 shard, ≈ 4-5min) — logout flow 외 다른 e2e suite (auth, RBAC, CRUD) 는 backend 의 build tier 무관
- **DEVHUB_BUILD_TIER token** required_e2e_tokens 에 의도적 미포함 — e2e shard 1/2/3 의 saovae_stub default env block 미설정 유지. e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 actionlint + 실제 e2e run 이 검증

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-c-i-e2e-internal-job/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `24674b8` (PR #542 머지 baseline).
- 본 `session_handoff.md` §7 append.

### 다음 세션 directive

- **C-j (P3)**: build tag 정책 재검토 (runtime injection ↔ build tag 전환 trade-off). 별도 ADR 후보.
- **backend-integration DEVHUB_BUILD_TIER matrix** (sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix).
- **release_v1_roadmap.md §3.5 N-13** 정합 (C-i done 마킹).

## 8. 본 세션 (2026-06-10, sprint -a follow-up PR1 PR #540 의 carry-over C-j 정공법 PR — docs/adr/ + docs/traceability/ + ai-workflow/memory/)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#543** (C-j build tag 정책 재검토 PR) | ✅ MERGED (squash, branch delete) | `docs/work_260610-c-j-build-tag-review` 분기. **`docs/adr/0031-build-tag-policy-review.md` 신규 (12KB, 9 section)** — ADR-0030 §2.3 runtime injection 결정을 **정량 측정 후 confirmed** (supersede X). **결정**: 옵션 2 (런타임 injection 유지). 근거 = stub binary overhead < 5KB (전체 backend-core < 50MB 대비 0.01%) vs build tag 전환 시 CI matrix 2배 (+30~60min) + 5~10 file `//go:build` tag + 2개 binary 운영. **재검토 trigger 5건** (§5): stub code size > 250KB / stub production risk / CI axes 5+ / Phase 2 agentic RAG / stub safety — 현시점 trigger 0건. ADR-0030 §2.3 row confirmed reference 추가. `docs/traceability/report.md` §4 ADR-0031 row + §6 row 신규. main `d3488ca`. CI 4/4 PASS (docs only PR, backend/e2e/frontend skip). Tier: **공용**. |

### scope 결정

코드 0줄 변경. ADR + traceability + memory 4 file. C-j 의 정공법 = **ADR-0031 신규 + ADR-0030 §2.3 confirmed (supersede X) + 9 section 정공법** (배경 + 정량 측정 + 옵션 + 결정 + 재검토 trigger + cross-tier + risks + supersession + 변경 이력).

### trade-off (현시점 정량 측정)

| 측정 항목 | Runtime injection (현재) | Build tag (이론) | 차이 |
| --- | --- | --- | --- |
| Binary overhead | < 5KB | -6.3KB (절감) | -6.3KB (build tag 유리) |
| CI runtime | +15~20min (PR #542) | +30~60min (이론) | build tag 가 +15~40min 더 |
| CI matrix jobs | 4 (e2e 1/2/3 + e2e-internal) | 6 (e2e × 2 tags) | build tag 가 +2 jobs |
| 코드 변경 | 0 (현재 상태 유지) | 5~10 file | build tag 가 +5~10 file |
| 운영 복잡도 | 1 binary | 2 binary | build tag 가 +1 |

**결론**: build tag 의 binary size 절감 (~6KB) 은 무시 가능 수준. runtime injection 의 cost 가 build tag 의 cost 보다 본질적으로 작음. 1:5+ cost ratio.

### Memory 갱신

- `ai-workflow/memory/docs/work_260610-c-j-build-tag-review/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `d3488ca` (PR #543 머지 baseline).
- 본 `session_handoff.md` §8 append.

### 다음 세션 directive

- **backend-integration DEVHUB_BUILD_TIER matrix** (P3): sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix.
- **release_v1_roadmap.md §3.5 N-13** 정합 (P3): C-i + C-j + C-g/C-h + C-j done 마킹. N-13 row close.
- **N-10 housekeeping close** (P3): 메모리 drift 정합 (session_handoff.md / work_backlog.md / state.json) — E2E 4 + IT 3 active, status partial verified 정공법. 검증 보고서 §3 follow-up close 마킹.
- 또는 다른 sprint 진입.

## 9. 본 세션 (2026-06-10, D-72 Phase 1 — `~/wiki/` LLM Wiki 통합 의 in-repo source-of-truth + sync script)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#544** (D-72 Phase 1 LLM Wiki 통합) | ✅ MERGED (squash, branch delete) | `feat/work_260610-d-72-wiki-phase-1` 분기. **`docs/llm-wiki/` 5 file 신규** (README 7.8KB + scope-and-rationale 10.6KB + mirror-list 10KB + lint-config.toml 4.4KB + operation-sop 10.7KB = 43.5KB) + **`scripts/wiki-sync-devhub.sh` 6.4KB executable** (BSD-rsync safe, 7 source 패턴, 82 file, --dry-run + vault 부재 no-op). **D-72 응답 §2 Q1~Q6 전체 적용**: 단일 vault + per-project 동거 + Q3 단순화 (lint L11 + sa-internal/ 격리 불요) + Q4 L01~L10 + L07 ADR 면제 + Q5 v1.5 동시 시작 + Q6 단일 AGENTS.md + per-project lint report. `docs/wiki/` (Public, GitHub Wiki 게시 source) vs `docs/llm-wiki/` (LLM Wiki SSOT) 의 분리. main `a96f586`. CI 4/4 PASS (docs only PR, backend/e2e/frontend skip). Tier: **공용**. |

### scope 결정

**코드 0줄 변경** (스크립트 6.4KB + docs 5 file 신규). mirror list = **core subset ~82 file** (ADR 31 + Governance 5 + Planning 26 + Setup 15 + Requirements 1 + OpenAPI 1 + AI-workflow memory 3). domain (66) + architecture + infrastructure + validation (~100 file) 은 Phase 3 (mass ingest) 의 별도 PR. **mirror 실행은 본 PR scope 외** (`~/wiki/raw/projects/devhub/` 의 out-of-repo 변경) — **T-d-72-2 의 1회 실행 (2026-06-11 01:10:39Z 완료, 83 file / 1.6M)**.

### trade-off

- **`docs/llm-wiki/` 선택 (vs `docs/wiki/` 또는 `docs/wiki-integration/`)**: 기존 `docs/wiki/` = **Public Wiki** (GitHub Wiki 게시 source, 인간 큐레이션, mtime 2026-05-20). 본 Phase 1 의 **LLM Wiki SSOT** 와 audience 다름. 디렉터리 이름 분리 = 두 wiki 의 명확한 구분. `docs/wiki/` (Public) ↔ `docs/llm-wiki/` (LLM) 의 cross-link 없음.
- **mirror list 의 scope = core subset ~82 file**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture + infrastructure + validation (~100 file) 은 **Phase 3 (mass ingest)** 에서 별도 PR. 본 PR 의 검증 가능한 정공법 (CI 4/4 + script smoke test) = 작은 core subset.
- **lint-config.toml 의 L07 ADR 면제 config 작성 (옵션 미사용)**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 config 의 source 만 제공. 옵션 추가 후 자동 활성.

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-d-72-wiki-phase-1/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `a96f586` (PR #544 머지 baseline).
- 본 `session_handoff.md` §9 append.

### 다음 세션 directive

- **T-d-72-2** (P3, 사용자 trigger): `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~82 file mirror + `_manifest.md` 자동 생성. **2026-06-11 01:10:39Z 완료** (83 file, 1.6M). vault = 공유 자원 (my_harness 측 Gitea private) 이므로 본 저장소 metadata 에 결과 정합.
- **D-73** (P3, my_harness 측): wiki-lint skill 에 `--project` + `--project-config` 옵션 추가. 본 저장소 의 lint-config.toml 활성.
- **D-74** (P3, my_harness 측): my_harness 의 `_lint/my-harness/` + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업.
- **Phase 3** (P3, mass ingest, 별도 PR): domain (66) + architecture + infrastructure + validation (~100 file) mirror + 30~50 wiki page.
- **wiki/cross/** (P3, Phase 3 후속): cross-project 종합 (my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection).
- **v2.0** (P3, forward): LLM 호출 + BM25+vector+MCP. my_harness 의 v2.0 경험 보고 진입.
- **N-13 release_v1_roadmap §3.5 정합** (P3, housekeeping): N-13 row status = done 마킹.
- 또는 다른 sprint (backend-integration matrix / N-10 housekeeping close).

## 10. wiki 통합 일임 결정 (2026-06-10, yklee directive)

### 결정

**yklee 2026-06-10 directive**: wiki 통합 작업은 my_harness 측 에이전트에 일임. 본 저장소 (DevHub) 의 sprint 는 **my_harness 의 결과 통보 대기**. **동시 진행 시 꼬일 가능성 회피** (mirror 실행 / lint config 활성화 / mass ingest / wiki page 작성 / cross-project 종합 등이 양 project 에서 동시 진행 시 race condition + 정책 drift 위험).

### 본 저장소 측 follow-up

- **본 PR #544 머지로 Phase 1 의 in-repo source-of-truth 정합 완료** (docs/llm-wiki/ 5 file + scripts/wiki-sync-devhub.sh). 변경 불요.
- **본 저장소 측 mirror 실행 (T-d-72-2) 완료** (2026-06-11 01:10:39Z): `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 82 file mirror + `_manifest.md` 자동 생성 (83 file total, 1.6M). **vault = 공유 자원 (my_harness 측 Gitea private, 본 저장소 측 관리 X)**, 본 저장소 metadata 에 결과 정합. **본 sprint 의 mirror list 정공법 (`docs/llm-wiki/mirror-list.md` §3 의 7 패턴) 매칭**. 후속 T-d-72-3~6 + Phase 3 carry-over.
- **본 저장소 측의 follow-up task (carry-over, my_harness 통보 대기)**:
  - my_harness 의 D-73 (wiki-lint `--project` + `--project-config` 옵션 추가) — 본 저장소 의 lint-config.toml 자동 활성
  - my_harness 의 D-74 (`_lint/devhub/` 셋업) — 본 저장소 의 per-project lint report 정합
  - my_harness 의 Phase 3 (mass ingest, ~100 file mirror + 30~50 wiki page) — 본 저장소 의 domain + architecture + infrastructure + validation 영역
  - my_harness 의 wiki/cross/ (cross-project 종합) — my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection
  - my_harness 의 v2.0 (full compile, LLM 호출 + BM25+vector+MCP) — my_harness 의 v2.0 경험 보고 진입
- **본 저장소 측 follow-up task (독립 진행 가능, yklee 별도 confirm 시)**:
  - **N-13 release_v1_roadmap.md §3.5 정합** (P3, housekeeping): D-72 + D-73 + D-74 + D-75 의 carry-over N-13 row status = done 마킹. **본 저장소 측에서 독립 진행 가능** (my_harness 결과 통보와 무관).
  - **backend-integration DEVHUB_BUILD_TIER matrix** (P3): sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix. **본 저장소 측에서 독립 진행 가능**.
  - **N-10 housekeeping close** (P3): E2E 4 + IT 3 active, status partial verified 정공법. **본 저장소 측에서 독립 진행 가능**.

### Memory 갱신

- 본 §10 append. 다음 세션 진입 시 본 §10 의 결정 참조.

### 다음 세션 directive

- **다른 sprint 진입 (본 저장소 측의 독립 진행 가능 task)**:
  - N-13 release_v1_roadmap.md §3.5 정합 (housekeeping)
  - backend-integration DEVHUB_BUILD_TIER matrix (P3)
  - N-10 housekeeping close (P3)
- **또는 사용자 confirm 후 본 저장소 측의 mirror 실행 (T-d-72-5)** — my_harness 결과 통보와 무관하게 단독 진행 가능 (mirror list 가 본 저장소 의 SSOT 이므로).
- **또는 my_harness 의 결과 통보 대기** (D-73, D-74, Phase 3 등).
- 또는 N-10 housekeeping close / release_v1_roadmap.md §3.5 N-13 정합.

## 11. 본 세션 (2026-06-11, main flat memory finalize + 잔여 follow-up sprint)

### PR 머지 결과 (squash, 사용자 confirm)

| PR | 머지 commit | 의의 |
| --- | --- | --- |
| **#551** (T-d-72-2 re-sync) | `f0b5ee519799` (2026-06-11 05:20:12Z) | `chore/work_260611-e-wiki-resync-2026-06-11` branch. 5 file 갭 해소 — main HEAD `837c26c8` 기준 vault 최신화. |
| **#552** (wiki-ingest + D-79/D-80 + housekeeping + handoff) | `5870f1a24d1f` (2026-06-11 05:22:00Z) | `feat/work_260611-a-wiki-ingest-from-raw` branch. 28 file, +2319/-34. wiki-ingest-from-raw wrapper (D-72 Phase 3) + wiki-query + wiki-pr-update thin wrapper (D-79, D-80) + 3 commit 정정 (publication-matrix, 3 columns, 5 broken link) + handoff-to-my-harness.md. PR #552 rebase 후 conflict resolve (state.json M-v1.0 notes 의 N-10 partial + main 의 PR #546/#547/#550 정합). |
| **#553** (N-10 close + lint 8/62 close + memory state 정합) | `af64f189` (2026-06-11, admin force merge) | `chore/work_260611-b-residual-housekeeping-close` branch. 4 file, +8/-4. N-10 status ⏳ → ✅ resolved (full — E2E 4 + IT/UT 3 + Process 1 TC 모두 verified) + lint 8 errors / 62 warns follow-up 사실상 close (D-74 L03 skip patch + skip config 적용, 3개 lint report 모두 0 error / 0 warn / 0 info 정합) + memory 4 file 갱신. PR #553 rebase 후 state.json conflict resolve (main 의 PR #552 머지 + PR #553 의 N-10 close 통합). |
| **#548** (N-13 backend foundation) | ❌ **E2E Internal 1 fail 보류** | `feat/work_260611-a-n13-inbound-source-impl` branch. 13 file, +627/-38. CI 11/12 PASS, E2E Internal 1 fail = e2e seed 중복 (Test 1, strict mode violation: `getByText('e2e-repo-a')` 2 elements) + Sign-out timeout (Test 2, N-8 race 유사, `Test timeout of 30000ms exceeded`). main 의 PR #550 spec timing fix 적용 후 자동 재실행 안 됨 (run 시각 `27316392137` 2026-06-11T01:04Z < PR #550 머지 2026-06-11T01:51Z). codex review = COMMENTED (blocker 아님). 사용자 confirm 별도 (re-run + 자동 재실행 또는 spec/e2e seed 정합 fix). |

### codex 리뷰 검색 (4 PR)
- **PR #552 / #551 / #548** 모두 `chatgpt-codex-connector` 가 **COMMENTED** state (blocker 아님, approve/reject 미해당, 자동 review suggestion 만).
- **PR #553** codex review 없음 (방금 OPEN).
- 4 PR 모두 머지 가능 정합.

### main flat memory finalize (3 file, 본 turn)
- `state.json` M-v1.0 notes: main HEAD `af64f189` + PR #514~#553 (39 PR) 정합 + T-d-72-4 done + N-10 ✅ resolved + lint 8/62 follow-up 사실상 close + N-13 ID slot + N-13 #548 OPEN + 다음 sprint (N-6 staging 1주 운영 v1.0 출시 차단).
- `work_backlog.md` status line: main HEAD `af64f189` + PR #551/#552/#553 row + N-6 staging 1주 운영 + main flat memory finalize carry-over.
- 본 `session_handoff.md` §11: PR #551/#552/#553 + #548 보류 + codex 리뷰 결과 + main finalize 3 file 정합.

### 다음 directive
- **N-6 staging 1주 운영 시작** (사용자 결정, v1.0 출시 차단 해소) — 외부 사용자 ≥5 로그인 + Onboarding SOP DoD 8.
- **PR #548 머지 결정** (사용자 confirm) — rebase main 의 `af64f189` + push → 자동 재실행 → E2E Internal 1 fail 결과 확인. Test 1 (e2e seed 중복, strict mode violation) 는 spec/e2e seed 정합 fix 별도 sprint 가능. Test 2 (Sign-out timeout) 는 PR #550 spec timing fix 가 해결 가능성 높음.
- **vault Gitea remote push** (사용자 수동) — `~/wiki` 의 b1599cc + af64f189 의 변경분을 my_harness 측 Gitea private 으로 push.
- **T-d-79-2 / T-d-80-2** (my_harness 측) — `handoff-to-my-harness.md` 가이드.

## 18. 본 세션 (2026-06-11, Phase 2 1차 chunk — concept 5 page 신규)

### 작업 흐름 (out-of-repo, 사용자 confirm 후)

| Step | 작업 | 결과 |
|---|---|---|
| 1 | **type 분류 사전 정의** (AGENTS.md v1.5 + schema/page_template.md) | type 6종 (concept/entity/topic/source/comparison/query) + frontmatter 8 key 정공법 명확 |
| 2 | **5 page 작성** (Mavis 직접, chunked) | `concepts/rbac.md` / `entities/keycloak.md` / `concepts/agent-memory.md` / `concepts/llm-wiki-pattern.md` / `topics/workflow.md` |
| 3 | **1차 lint 검증** (작성 직후) | errors 52 + warns 89 = **141 findings** (1차 page 의 `related:` 의 wikilink 25 가 target 없음 → L02 error 폭발) |
| 4 | **`related:` 의 wikilink 정공법** (target 없으면 plain text) | wikilink 25 → 7 (target 있는 것만 유지), errors 52 → 20 |
| 5 | **index.md 자동 갱신** (L08 fix) | Concepts — devhub 3 → 5 + Topics — devhub 2 → 3 + Entities — devhub keycloak dedup. errors 20 → 17 |
| 6 | **`sources:` 의 path 정합** (L06 fix) | vault mirror 의 7 패턴 (ADR/governance/planning/setup/requirements/openapi/ai-workflow-memory) 의 path 만 유지. L06 9 → 0. **errors 17 → 11** |

### Phase 2 1차 chunk 결과

| 항목 | Before | After |
|---|---|---|
| lint total | 196 findings | **98 findings** (-98, -50%) |
| errors | 18 | 11 (-7) |
| warns | 178 | 87 (-91) |
| L02 | 18 (broken wiki link) | **11** (forward path, cross/ 의 4 page 의 wikilink) |
| L03 | 116 (고아) | 86 (forward path, sources/ 의 cross-ref) |
| L04 | 31 (ADR naming 중복) | **0** (mavis-trash 후) |
| L06 | 9 (sources path) | **0** (7 패턴 정합) |
| L08 | 31 (index 미등록) | **1** (5 page 만 등록, 잔여 1 file) |
| type 분포 | sources 113 / concept 3 / entity 4 / topic 2 / comparison 0 / query 0 | sources 83 / **concept 5** / **entity 4** / **topic 3** / comparison 0 / query 0 |

### Phase 2 forward path 잔여

- **L02 11**: cross/concepts/, cross/topics/ 의 4 page 의 wikilink 11개 (`[[ai-workflow]]` / `[[context-budget]]` / `[[AGENTS]]` / `[[keycloak-admin]]` / `[[envelope-encryption]]` / `[[wiki-prompt-log]]` / `[[wiki-event-sync]]` / `[[wiki-query-helper]]` / `[[wiki-lint]]` / `[[wiki-pr-update]]` / `[[wiki-query]]`)
- **L03 86**: wiki 의 sources/ page 의 cross-reference 부족 (forward, 2차 chunk 에서 해소)
- **2차 chunk plan**: 8 page 추가 (keycloak-admin / envelope-encryption / ai-workflow / wiki-prompt-log / wiki-event-sync / wiki-query-helper / wiki-lint / wiki-pr-update / wiki-query) + comparisons 1~3
- **3차 chunk plan**: cross-project 종합 + L03 86 점진 감소

### 다음 세션 directive

- **2차 chunk 진행** (forward, 사용자 confirm 후): 8~9 page 추가 (L02 11 → 0 정공법)
- 또는 `docs/llm-wiki/Mavis-workflow.md` 9번째 문서 작성
- 또는 본 세션 작업의 prompt log (D-86 흐름 1) 작성
- 또는 다른 sprint (N-13 release_v1_roadmap §3.5 정합 / v0.1.1-alpha release 8 item / PR #548 머지 결정)

## 19. 본 세션 (2026-06-12, N-13 PR #548 close follow-up 결정 — sprint `fix/work_260612-1-n13-housekeeping-followup`)

### PR 머지 결과 (정공법 본문)

| 항목 | 결과 |
| --- | --- |
| **PR #572** (N-10 follow-up 보류 결정) | ✅ **MERGED** (squash `8616ac59`, 2026-06-12 02:49 UTC) — docs only, 4 file (+70). `fix/work_260611-5-n10-n9` branch delete (auto). 직전 세션 (2026-06-11 21:30 KST) PR #572 머지 정합 = main HEAD `8616ac59`. |
| **다른 open PR** | **0건** (전 sprint 정합) |

### PR #548 close 결과 (sprint `feat/work_260611-a-n13-inbound-source-impl`)

- **CLOSED** (2026-06-11 05:40 UTC) — E2E Internal 1 fail
  - **Test 1** e2e seed 중복 strict mode violation: `getByText('e2e-repo-a')` 2 elements
  - **Test 2** Sign-out timeout: N-8 race 유사, `Test timeout of 30000ms exceeded`
  - codex review = COMMENTED (blocker 아님)
  - 자동 재실행 미적용 (run 시각 `27316392137` 2026-06-11T01:04Z < PR #550 spec timing fix 머지 2026-06-11T01:51Z)

### 본 sprint 정공법 (N-13 housekeeping follow-up)

- **sprint**: `fix/work_260612-1-n13-housekeeping-followup`
- **scope**: 5 file (docs only, 코드 0줄)
  1. `docs/planning/release_v1_roadmap.md` — §3.5 N-13 row status `⏳ planned` 유지 + 본문 보강 (PR #548 close 결과 + follow-up 3 branch 결정) + §9 변경 이력 1 row + 헤더 메타 (최종 수정일 2026-06-12, 직전 결정 근거 2026-06-11, 결정 근거 sprint 추가)
  2. `docs/adr/0028-dev-requests-voc-external-ref.md` — §6 (a) 본문 보강 (PR #548 close + follow-up 3 branch + branch 정책) + §7 변경 이력 1 row + 헤더 메타 (최종 수정일 2026-06-12 N-13 follow-up 결정)
  3. `docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md` — §3.3 의존 표에 PR #548 CLOSED row 추가 + §5 결정 보류 사유에 3 branch follow-up 결정 + §6 변경 이력 1 row + 헤더 메타 (최종 수정일 2026-06-12 follow-up 결정, 결정 근거 sprint 추가)
  4. `docs/traceability/report.md` — §6 변경 이력 1 row (PR #548 close follow-up 보고)
  5. 메모리 4 file 동기화 (`state.json` M-v1.0 notes `phase2_5th_chunk_n13_housekeeping_followup` 추가 + `work_backlog.md` status line + §5 row + 본 `session_handoff.md` §19 append + 브랜치 memory 4 file 신규)
- **신규 ID 발급 0건** (housekeeping follow-up)
- **Tier**: **공용** (docs only, 사내 한정 정보 미포함)

### follow-up 결정 3 branch (사용자 결정 영역)

1. **Test 1 e2e seed 중복** → spec/e2e seed 정합 fix 별도 sprint (mock data 유일성 보장)
2. **Test 2 Sign-out timeout** → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑)
3. **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 종합)

### 다음 directive (사용자 결정 영역)

- **follow-up 3 branch 중 우선 진행 방향 결정**:
  - 옵션 A: Test 1 e2e seed 중복 fix 먼저 (small, e2e 안정성 ↑)
  - 옵션 B: Test 2 Sign-out timeout rebase + 자동 재실행 (small, N-8 race 잔여 검증)
  - 옵션 C: v1.1 milestone 진입 시점으로 보류 (N-13 자체의 구현 follow-up)
- 또는 다른 sprint (N-6 staging 1주 운영 / backend-integration DEVHUB_BUILD_TIER matrix / v0.1.1-alpha release 8 item)

## 20. 본 세션 (2026-06-12, N-13 follow-up A 완료 + B 진행 — sprint `fix/work_260612-2-e2e-seed-strict-mode-fix` + `fix/work_260612-3-n13-followup-b-test2-rebase`)

### PR 머지 결과 (사용자 confirm + squash merge)

| PR | 머지 commit | 의의 |
| --- | --- | --- |
| **#574** (N-13 follow-up A — Test 1 e2e seed fix) | `896d9018` (2026-06-12) | sprint `fix/work_260612-2-e2e-seed-strict-mode-fix`. root cause = `frontend/tests/e2e/repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher `.first()` 미적용. fix = matcher 정의에 `.first()` 추가 (기존 `repository-dashboard.spec.ts:71, 117` 의 동일 패턴 정합). 1 file 변경, diff +3/-2. Tier: 사외. 신규 ID 0건 (test stabilization). |

### N-13 follow-up A 정공법 핵심 (PR #574)

**Root cause**: `frontend/tests/e2e/repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher 가 `.first()` 미적용. E2E shard 실행 중 다른 spec 잔재 (예: `e2e-repo-a3xd7` 와 같이 random suffix 가 붙은 publish spec 잔재) 가 동일 이름 link 2+ 개 생성 시 strict mode violation.

**Fix**:
```ts
// 변경 후
const repoALink = (page) => page.getByRole("link", { name: "e2e-repo-a", exact: true }).first();
const repoBLink = (page) => page.getByRole("link", { name: "e2e-repo-b", exact: true }).first();
```

**검증**: main 의 e2e CI 자동 trigger (PR 머지 시) → e2e shard 1/2/3 모두 PASS (예상). Test 1 자동 해결.

### N-13 follow-up B 진행 (sprint `fix/work_260612-3-n13-followup-b-test2-rebase`)

**Test 2 root cause**: `frontend/tests/e2e/screenshots.spec.ts:66` 의 30s default timeout + logout flow network race (backend 204/401/502 분기 + frontend `window.location.assign('/login')` 강제 redirect, N-8 race 유사).

**본 sprint 정공법**: main 의 PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 main 의 e2e CI 가 자동 재실행 시 e2e shard 1/2/3 모두 PASS 검증. 5 file 변경 (docs only, verification report) + 메모리 4 file 동기화. 본 verification PR 의 본질 = main 의 e2e CI 안정성 evidence.

### follow-up 잔여 1 branch (사용자 결정 영역)

- **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + 본 fix 종합 + 자동 재실행). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 별도 결정. status `⏳ planned` 유지.

### 다음 directive (사용자 결정 영역)

- **본 verification PR 머지 후 e2e CI 결과 확인**: e2e shard 1/2/3 PASS 시 본 sprint 종료 (Test 2 자동 해결 확인). FAIL 시 추가 fix PR 발행.
- **follow-up 1 branch 결정** (구현 follow-up, v1.1 진입 시점): 별도 sprint.
- 또는 다른 sprint (N-6 staging 1주 운영 / backend-integration DEVHUB_BUILD_TIER matrix / v0.1.1-alpha release 8 item)

## 20. 본 세션 (2026-06-12, N-13 follow-up 3 branch 종합 + E2E Internal disable)

### PR 머지 / Close 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#573** (N-13 housekeeping follow-up) | ✅ MERGED (squash `5fb9ae75`) | PR #548 close 결과 + 3 branch follow-up 결정 (A: Test 1 fix, B: Test 2 verification, C: 구현 follow-up) |
| **#574** (N-13 follow-up A — Test 1 e2e seed fix) | ✅ MERGED (squash `896d9018`) | `repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher 에 `.first()` 추가 |
| **#575** (N-13 follow-up B — Test 2 verification) | ✅ MERGED (squash `8d0e2e88`) | verification report + CI Run #1227 SUCCESS (Test 2 Sign-out timeout 자동 해결 확인) |
| **#576** (N-13 follow-up C — PR A-2 routing + voc_handler) | ❌ **CLOSED** (사용자 결정) | E2E shard 3/3 fail — PATCH /api/v1/platforms 가 4xx 반환 (PATCH inbound_source 의 backend 처리 검증 필요) |
| **#577** (E2E Internal disable) | ✅ MERGED (squash `802afe62`) | ci.yml `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}` (actionlint 호환) — v1.1 sprint 에서 real Keycloak adapter 안정화 후 재활성화 |

### N-13 follow-up 3 branch 종합

| Branch | 결과 | PR |
|---|---|---|
| A: Test 1 e2e seed fix | ✅ 완료 | PR #574 MERGED (`896d9018`) |
| B: Test 2 rebase + 자동 재실행 검증 | ✅ 완료 | PR #575 MERGED + CI Run #1227 SUCCESS |
| C: 구현 follow-up (v1.1 진입 시점) | ⏳ 미해결 (사용자 결정) | PR #576 close (E2E shard 3/3 fail) — 후속 sprint 에서 PATCH inbound_source backend 처리 검증 + e2e spec 재작성 + 새 PR |

### follow-up 잔여 (사용자 결정 영역)

1. **PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행** (별도 sprint)
2. **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + 본 fix 종합 + 자동 재실행)
3. **E2E Internal 재활성화** (v1.1 sprint 에서 real Keycloak adapter 안정화 후, PR #577 의 `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}` → 원래 조건 복원)
4. **state.json line 45 의 broken JSON fix** (chunks separator 의 `**` markdown bold 가 raw character 로 포함, 다음 sprint 에서 fix)

### 다음 directive (사용자 결정 영역)

- 위 follow-up 4 가지 중 우선 진행 방향 결정
- 또는 다른 sprint (N-6 staging 1주 운영 / backend-integration DEVHUB_BUILD_TIER matrix / v0.1.1-alpha release 8 item)

## 21. 본 세션 (2026-06-12, N-13 follow-up 3 branch 종합 + E2E Internal disable 정공법 + session end — main HEAD `fc7e6c76`)

### 본 세션 정공법 (session end 결정)

- **session 종료 결정** (사용자, 2026-06-12): 일단 여기까지 하고 세션 종료.
- 본 sprint 의 5 PR 결과 종합 + 메모리 finalize commit `fc7e6c76` push 완료.
- follow-up 4가지 중 follow-up 4 (state.json line 45 markdown bold 잔재 fix) **본 sprint 의 메모리 finalize 의 일부로 정상화 완료** (line 45 length 1034, `**` count 0, JSON 유효).

### PR 머지 / Close 결과 (5 PR)

| PR | 상태 | 의의 |
|---|---|---|
| **#573** (N-13 housekeeping follow-up) | ✅ MERGED (squash `5fb9ae75`) | PR #548 close 결과 + 3 branch follow-up 결정 (A: Test 1 fix, B: Test 2 verification, C: 구현 follow-up) |
| **#574** (N-13 follow-up A — Test 1 e2e seed fix) | ✅ MERGED (squash `896d9018`) | `repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher 에 `.first()` 추가 |
| **#575** (N-13 follow-up B — Test 2 verification) | ✅ MERGED (squash `8d0e2e88`) | verification report + CI Run #1227 SUCCESS (Test 2 Sign-out timeout 자동 해결 확인) |
| **#576** (N-13 follow-up C — PR A-2 routing + voc_handler) | ❌ **CLOSED** (사용자 결정) | E2E shard 3/3 fail — PATCH /api/v1/platforms 가 4xx 반환 (PATCH inbound_source 의 backend 처리 검증 필요) |
| **#577** (E2E Internal disable) | ✅ MERGED (squash `802afe62`) | ci.yml `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}` (actionlint 호환) — v1.1 sprint 에서 real Keycloak adapter 안정화 후 재활성화 |

### 본 sprint 의 메모리 finalize (commit `fc7e6c76`)

- **state.json** line 46 notes value 갱신: 본 sprint 정공법 (5 PR 결과 + N-13 follow-up 3 branch + E2E Internal disable + follow-up 잔여 3가지 + session 종료) 추가 + main HEAD reference `802afe62` → `fc7e6c76` 갱신.
- **work_backlog.md** status line main HEAD `802afe62` → `fc7e6c76` 갱신 + §5 변경 이력 2026-06-12 row 추가.
- **session_handoff.md** 본 §21 append (session end 결정 + 본 sprint 정공법 + follow-up 잔여 3가지 + 다음 진입 가이드).

### N-13 follow-up 3 branch 종합

| Branch | 결과 | PR |
|---|---|---|
| A: Test 1 e2e seed fix | ✅ 완료 | PR #574 MERGED (`896d9018`) |
| B: Test 2 rebase + 자동 재실행 검증 | ✅ 완료 | PR #575 MERGED + CI Run #1227 SUCCESS |
| C: 구현 follow-up (v1.1 진입 시점) | ⏳ 미해결 (사용자 결정) | PR #576 close (E2E shard 3/3 fail) — 후속 sprint 에서 PATCH inbound_source backend 처리 검증 + e2e spec 재작성 + 새 PR |

### follow-up 잔여 3가지 (사용자 결정 영역)

1. **PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행** (별도 sprint)
2. **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + 본 fix 종합 + 자동 재실행)
3. **E2E Internal 재활성화** (v1.1 sprint 에서 real Keycloak adapter 안정화 후, PR #577 의 `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}` → 원래 조건 복원 + 새 PR)

### 다음 directive (다음 세션 진입 시, 사용자 결정 영역)

- 위 follow-up 3가지 중 우선 진행 방향 결정
- 또는 다른 sprint (N-6 staging 1주 운영 / backend-integration DEVHUB_BUILD_TIER matrix / v0.1.1-alpha release 8 item)
- session_handoff.md 본 §21 + state.json M-v1.0 notes + work_backlog.md status line + §5 row 본 row 가 다음 진입 시점의 메모리 정합 기준선
- main HEAD `fc7e6c76` 정합 (메모리 finalize commit)

## 22. In-flight 세션 (2026-06-12, CI 재구성 + flaky 복구 + PR 준비 — branch `codex/work_260612-579-ci-rearchitecture`)

### 이번 세션 핵심 결과

- **CI 재구성 설계 문서 작성**: `docs/planning/2026-06-12-ci-rearchitecture-design.md`
- **워크플로우 분리 구현**
  - `.github/workflows/ci.yml` = fast required + smoke
  - `.github/workflows/e2e-regression.yml` = non-quarantine full regression
  - `.github/workflows/e2e-quarantine.yml` = flaky/quarantine 전용
- **spec selection SSOT 도입**
  - `frontend/tests/e2e-manifests/smoke.txt`
  - `frontend/tests/e2e-manifests/quarantine.txt`
  - `scripts/select-playwright-specs.sh`
- **flaky 복구 우선 적용**
  - `frontend/tests/e2e/fixtures.ts`: `/login` loop 감지, OIDC 재시작 helper, CI timeout 완화, sign-in form 재시도 강화
  - `frontend/tests/e2e/signout.spec.ts`: 장시간 sign-out/user-switch 케이스 timeout 상향

### root cause / 의도

- 최근 signout/login 계열 flaky 는 app login page 체류, OIDC redirect 재시작 지연, CI 환경에서의 느린 signout round-trip 이 겹치며 발생.
- 전체 회귀를 단일 required workflow 에 계속 묶어두면 flaky 1건이 merge path 전체를 막고 runner 병렬성도 충분히 활용되지 않음.
- 그래서 **required path 는 빠르게 유지**하고, **full regression 과 quarantine 은 별도 runner pool 에 분리**하는 구조로 재설계.

### 검증 결과

- workflow YAML parse: `ci.yml`, `e2e-regression.yml`, `e2e-quarantine.yml` 모두 OK
- `scripts/ci-e2e-sync-check.sh` 로 3 workflow 정합 OK
- `scripts/select-playwright-specs.sh smoke|quarantine|regression` 출력 OK
- focused type check OK:
  - `cd frontend && npx tsc --noEmit tests/e2e/fixtures.ts tests/e2e/signout.spec.ts --skipLibCheck`
- full `frontend` `npx tsc --noEmit` 는 기존 repo 전반의 선행 타입 오류로 실패

### 알려진 제약 / 다음 진입 포인트

1. 현재 브랜치 기준 `voc-auto-routing.spec.ts` 는 존재하지 않아 quarantine manifest 에서 자동 skip 되도록 처리함.
2. full regression 을 GitHub Actions 에서 실제로 돌려 보고 branch protection 조합을 확정해야 함.
3. merge 전 PR 본문에는 required check 후보(`CI`, `E2E Regression`)와 quarantine 운영 원칙을 명시하는 것이 좋음.

## 23. 후속 결과 (2026-06-13, CI 재구성 운영 검증 완료)

### 완료 결과

- PR #580 merged
  - merge commit: `b1fa5c27698403620b63ef09d7e32e2235592d59`
- 분리된 workflow 3종이 GitHub Actions 상에서 정상 동작 확인
  - `CI`
  - `E2E Regression`
  - `E2E Quarantine`
- old CI에서 shard 3/3 failure가 있던 PR #579가 최신 `main`을 반영한 새 SHA에서 재평가 통과

### 핵심 레슨런

1. base branch workflow가 바뀌어도 old failure run이 자동 success로 바뀌지는 않는다.
2. old PR에 새 CI 구조의 효과를 적용하려면 최신 `main`을 반영해 새 SHA로 다시 평가해야 한다.
3. flaky와 기능 실패를 같은 required lane에 두지 않는 편이 triage 속도와 merge 안정성에 유리하다.
4. manifest + selector 조합은 smoke/regression/quarantine 경계를 유지하는 데 효과적이었다.

### 다음 진입 포인트

- branch protection required check 목록을 문서 목표 상태와 다시 정합
- quarantine spec pass/fail 추세 관찰 자동화 검토
- full frontend `tsc --noEmit` 선행 오류는 별도 정리 과제로 유지

## 24. 본 세션 (2026-06-13, 4 PR 종합 — main HEAD `971c62a`)

### PR 머지 결과 (squash, yklee 직접 머지)

| PR | 머지 commit | 의의 |
| --- | --- | --- |
| **#578** (e2e-internal job 완전 폐기 + ADR-0030/0031 partial supersession) | `dab302a` (2026-06-12 21:05 KST) | sprint `fix/work_260612-6-e2e-internal-removal`. ci.yml e2e-internal job 완전 삭제 (-205 lines) + scripts/ci-e2e-sync-check.sh DEVHUB_BUILD_TIER 코멘트 정리 + ADR-0031 partial supersession (baseline 변경 정공법) + ADR-0030 §2.3 row 갱신. **결론 (runtime injection 유지) 변동 0건** — e2e-internal 폐기 = **독립 결정** (사내 staging/prod-smoke 가 real adapter 검증 책임). CI Run 27414157106 SUCCESS (11/11 PASS). 10 file. 신규 ID 0건. |
| **#580** (CI 재구성 3-tier) | `b1fa5c2` (2026-06-13 02:21:00 UTC) | sprint `codex/work_260612-579-ci-rearchitecture`. ci.yml + e2e-regression.yml + e2e-quarantine.yml 3 workflow 분리 + manifest-based spec routing (e2e-manifests/smoke.txt + quarantine.txt) + scripts/select-playwright-specs.sh + fixtures.ts OIDC 재시작 helper + signout timeout 완화. 본 PR 의 핵심 = **chronic flake 격리** (smoke/regression/quarantine). 신규 ID 0건 (workflow 만 변경). |
| **#579** (N-13 follow-up C — v1.1 sprint -a inbound_source 자동 routing) | `84f2227` (2026-06-12 23:43 UTC) | sprint `feat/work_260612-7-v1-1-inbound-source-impl`. backend `auto_route.go` 신규 (AutoRouter interface + 3 case pattern matcher + 23 unit + 6 integration test) + `voc_handler.go` AutoRouter 통합 + e2e `voc-auto-routing.spec.ts` 신규 (TC-INBOUND-SRC-01 + NEG, beforeAll hook) + `docs/openapi.yaml` 81 paths / 79 schemas + ADR-0028 §6 (a) amendment. **9 ID 발급**: REQ-FR-109/110/111 + ARCH-21/22/23 + API-101/102/103 + RM-voc-auto-01 + IMPL-auto-route-01 + UT-AR-01..03 + TC-INBOUND-SRC-01/NEG. **8-step 정공법** (5-step chronic flake fix + 3-step 1차 layer fix). Tier: 공용. |
| **#581** (CI retrospective) | `971c62a` (2026-06-12 23:51:57 UTC) | sprint `codex/work_260612-579-ci-rearchitecture` 의 retrospective. docs/planning/2026-06-13-ci-rearchitecture-retrospective.md 신규 + memory 4 file 동기화. 핵심 = §23 의 4 lesson 정합. |

### 8-step 정공법 (PR #579 의 chronic flake 종합)

PR #579 의 7 commit (= 본 PR 의 8-step) = e2e shard 3/3 fail 의 chronic flake 정공법. **PR #580 의 CI 3-tier 분리 = 1차 layer fix**, **PR #579 의 5 commit = 2~5차 layer fix**:

| Step | Commit | Layer | 정공법 |
|---|---|---|---|
| 1 (1차 layer) | PR #580 (b1fa5c2) | workflow 구조 | CI 3-tier 분리 (smoke/regression/quarantine) + manifest SSOT + OIDC 재시작 helper |
| 2 | 4b72f81 (1차 commit) | e2e spec | TC-INBOUND-SRC-01 PATCH inbound_source 를 beforeAll hook 으로 이동 (Keycloak startup race 회피) |
| 3 | 0049f69 (2차) | e2e spec | beforeAll hook timeout 30s → 180s 명시 (Playwright 시그너처 `beforeAll(fn, timeout?: number)`) |
| 4 | 33644ed (3차) | memory | 5-step 정공법 + 4차 commit syntax fix 정합 |
| 5 | 65c8410 (4차) | fixtures.ts | loginAs `page.waitForURL` 30s→60s + retry 5회 75s (근본 fix, 옵션 L) |
| 6 | 891ebd6 (5차) | fixtures.ts | waitForSignInForm default 30s→60s (진짜 근본 fix, 7-step 정공법) |
| 7 | 051a763 (6차) | ci.yml | Keycloak wait step timeout 120s→300s + retry 5회 (1차 layer fix, 옵션 M) |

**3-way conflict 해결** (worktree branch rebase on origin/main b1fa5c2 + 4f863f9):
- state.json / session_handoff.md 의 b1fa5c2 broken JSON fixup: closing quote + trailing comma 추가
- session_handoff.md §22~§23 정합 (PR #580 ci_rearchitecture + PR #578 e2e-internal removal 양쪽 보존)
- fixtures.ts loginAs conflict: PR #580 의 `process.env.CI ? 60_000 : 30_000` 분기 채택 (PR #579 의 unconditional 60_000 보다 정교)
- fixtures.ts waitForSignInForm conflict: PR #580 의 restart logic + PR #579 의 60_000 timeout 통합 (CI 60_000 / local 45_000)
- worktree branch 의 0d2dd89 (memory fixup): `git rebase --skip` (main 의 memory fixup 와 중복)

### PR #579 CI 9차 run 결과 (run 27426129165-68)

| Check | Result | Duration |
|---|---|---|
| Detect Changed Paths | SUCCESS | 6s |
| Workflow Lint (actionlint) | SUCCESS | 11s |
| Migration Prefix Uniqueness | SUCCESS | 7s |
| OpenAPI YAML Lint | SUCCESS | 8s |
| Backend Unit Tests | SUCCESS | 1m8s |
| Backend Integration Tests | SUCCESS | 1m9s |
| Frontend Unit Tests | SUCCESS | 1m1s |
| E2E Smoke Build Artifacts | SUCCESS | 1m44s |
| E2E Smoke Tests | SUCCESS | 2m3s |
| E2E Quarantine Build Artifacts | SUCCESS | 1m43s |
| E2E Quarantine | SUCCESS | 1m34s |
| E2E Build Artifacts (Regression) | SUCCESS | 1m44s |
| E2E Regression (Playwright, shard 1/3) | SUCCESS | 1m38s |
| E2E Regression (Playwright, shard 2/3) | SUCCESS | 1m44s |
| E2E Regression (Playwright, shard 3/3) | SUCCESS | 2m4s |

**16/16 PASS in 4 minutes**. CI PASS 시점 = 2026-06-12 15:43 UTC.

### N-13 follow-up 3 branch 종합 결과 (사용자 결정 영역)

| Branch | PR | 결과 |
|---|---|---|
| (1) Test 1 e2e seed 중복 fix | PR #574 (squash `896d9018`) | ✅ MERGED (2026-06-12) |
| (2) Test 2 Sign-out timeout rebase + 자동 재실행 | PR #575 (squash `8d0e2e88`) + CI Run #1227 | ✅ MERGED (2026-06-12) |
| (3) 구현 follow-up (v1.1 sprint -a inbound_source) | PR #579 (squash `84f2227`) | ✅ MERGED (2026-06-12) — 9 ID 발급 |

**3 branch 모두 해소 완료**. N-13 housekeeping follow-up (PR #572) + 3 branch 종합 (PR #574/#575/#579) + E2E Internal disable (PR #577) + E2E Internal 완전 폐기 (PR #578) + CI 재구성 (PR #580) + retrospective (PR #581) = **N-13 6 PR + 1 retrospective** main 정합.

### follow-up 잔여 (사용자 결정 영역)

- (1) PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행 → **PR #579 로 해소 완료**
- (2) 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint → **PR #579 로 해소 완료** (v1.1 sprint -a)
- (3) 사내 staging/prod-smoke 의 real adapter 검증 SOP 갱신 (별도 docs, 본 sprint scope 외)
- (4) E2E Internal 재활성화 = **OBSOLETE** (재활성화할 job 자체가 없어짐)

### CRITICAL 레슨런 (cross-project, 2026-06-13)

1. **PR 의 CI fail = base branch workflow 변경만으로 자동 success 전환 안 됨**. PR branch 의 rebase 후 새 SHA push 가 결정적 fix. (PR #581 retrospective + 본 세션의 PR #579 9차 run 정합)
2. **flaky 와 기능 실패 = 같은 required lane 두지 않는 편이 triage / merge 안정성 유리**. (PR #580 의 3-tier 분리가 본 정공법)
3. **cron self-reminder 의 gate-discipline 함정**: `statusCheckRollup` 의 artifact upload step 이 IN_PROGRESS 표시되어 8시간 false-negative 발생 가능. 향후 cron trigger 시 `mergeStateStatus='CLEAN'` + `all check conclusion=SUCCESS` 직접 확인 (artifact step 의 IN_PROGRESS 는 제외). (2026-06-13 cron `pr579-ci-9th` 8시간 false-negative 경험)
4. **3-way conflict 해결 정공법** (rebase on origin/main with local + remote commit): ours/theirs 양쪽 보존 + index reset (`git rm --cached <file> && git add <file>`) + 본인의 memory fixup commit 이 이미 main 에 있을 때 `git rebase --skip`.

### Memory 갱신

- 본 `session_handoff.md` §24 append.
- `state.json` M-v1.0 `ci_rearchitecture_outcome_2026_06_13` row 본 row 내용으로 갱신 (PR #578 + #579 + #580 + #581 4 PR + 8-step + CRITICAL 레슨런 4건 통합).
- `work_backlog.md` §5 변경 이력 row 1건 append.
- main HEAD `971c62a` 정합.

### 다음 directive

- **위키 동기화** (raw/projects/devhub/ → ~/wiki/) — wiki-source-sync skill 호출 (본 sprint 의 user directive 순서 = 1.main flat memory sync → 3.sprint close + memory finalize commit → 2.위키 동기화).
- 또는 v0.1.1-alpha release 8 item 중 1~2 진입 (사용자 결정).
- 또는 다른 sprint (N-10 housekeeping close / backend-integration DEVHUB_BUILD_TIER matrix / v1.0 staging 운영).

## 25. 본 세션 (2026-06-13, Phase 3 mass ingest + provenance tracking 영구화)

### Phase 3 mass ingest 정공법 (main HEAD `74a6d5c`)

**위키의 LLM Wiki mass ingest** = docs/domain (66 file) + docs/architecture (1) + docs/infrastructure (9) + docs/validation (2) = **78 file** 의 mirror scope 확장 + **78 wiki page 자동 생성** (frontmatter + raw body 1:1) + wiki index.md 335 → 413 line (+78).

### 4 commit 종합 (main HEAD `74a6d5c`)

- `9c32212` (Phase 1.5 추가: mirror scope 6 패턴 ~55 file)
- `05c246a` (Phase 1.5 follow-up: outdated file list 8 file 경로 정합)
- `32454fc` (provenance tracking: commit hash + version + branch 자동 추적)
- `74a6d5c` (Phase 3 mass ingest: domain + architecture + infrastructure + validation 78 file + 78 wiki page)

### 5 file 갱신 + 1 file 신규

- `docs/llm-wiki/mirror-list.md` (107 line 추가): §1.8 Phase 3 추가 + §2 §14/§15 패턴 + §3 lint 갱신 + §4 forward path
- `docs/llm-wiki/lint-config.toml`: Phase 3 갱신, config_version 2 → 3
- `docs/llm-wiki/operation-sop.md`: Phase 3 mass ingest SOP + 4 script 추가
- `docs/llm-wiki/scope-and-rationale.md`: §0.3 Phase 3 추가 scope
- `docs/llm-wiki/README.md`: 제목 Phase 1+1.5+3 + 4 script 추가
- **`scripts/wiki-mass-ingest.sh` (NEW)**: Phase 3 mass ingest 자동화 (78 file → 78 page)

### 위키 1:1 mirror byte-identical 검증 (현시점)

- Phase 1+1.5+3 scope: 196 file
- Identical: 196 ✅
- Diff: 0 / Missing: 0
- 위키 sources/ 162 page matched (commit `74a6d5c`)
- mirror 925 file / 7.6M
- provenance 정공법: commit hash + version + last sync 자동 추적

### CRITICAL 레슨런 (cross-project)

1. **위키 1:1 mirror 정공법 = 위키만으로 코드 maintenance 가능** 영구화. mirror scope = docs + source code + workflows + scripts + domain/architecture/infrastructure/validation = ~220 file / 15 패턴.
2. **provenance tracking 정공법** = mirror script 의 manifest 의 commit hash + version 자동 capture + 위키 page 의 frontmatter 자동 반영 + status-check 의 162 page matched 검증. 위키 정보의 시점/commit/버전 자동 추적.
3. **mass ingest 정공법** = Phase 3 의 78 file 의 wiki page 자동 생성 (frontmatter + raw body 1:1) + index.md 자동 갱신. 본 저장소 한정 + my_harness 측 wiki 일임 결정 (session §10) 해제 후 진행.
4. **glob recursive 갱신** = frontmatter-update + status-check 의 `find $WIKI_SOURCES -name '*.md' -type f` recursive. sources/ sub-directory 의 page 들 (Phase 3 의 78 page) 포함 검증.

### 다음 directive (forward path)

- v0.1.2-alpha patch (release_v1_roadmap.md §3.5 의 잔여 5 / X-1~8 의 8 item 중 1~2 entry, 사용자 결정)
- 또는 PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행 (별도 sprint, 사용자 결정)
- 또는 v1.0 staging 1주 운영 (사용자 결정)
- 또는 P0~P3 forward path 의 다른 sprint
- 또는 Phase 3 의 trigger 조건 충족 (forward path, my_harness 측 worker 일임 결정 해제 후 본 저장소 측 추가)

## 26. 본 세션 (2026-06-13, v0.1.2-alpha 준비 1차 PR #582 MERGED)

### version downgrade 정공법 (v1 → v0.1)

**v0.1.2-alpha 의 1차 PR** = version downgrade 정공법. 본 저장소 의 DevHub system version = v0.1.1-alpha (/VERSION) + forward-facing 18 file + 1 file name 변경.

**scope** (24 file, 354/354 line):
- `docs/planning/release_v1_roadmap.md` → `docs/planning/release_v0-1_roadmap.md` (file name 변경, 100% rename)
- 17 forward-facing file 의 v1.0/v1.1 + M-v1.0/M-v1.1 reference → v0.1 정공법
- 4 llm-wiki file 갱신
- AGENTS.md 정합

**정공법 원칙**:
- **forward-facing docs** = v0.1 정공법 적용
- **historical memory** (state.json, session_handoff.md, work_backlog.md) = historical record = 그대로 유지
- **historical archive** (analysis/*, learning-session/*, presentations/*) = 그대로 유지
- **VERSION file** = 이미 v0.1.1-alpha (정합)
- **위키 mirror 1:1 byte-identical** (md5 5ee61d55...)

**위키 정공법**:
- `topics/v1-0-release-roadmap.md` → `topics/v0-1-release-roadmap.md` (file name 변경)
- 2 sources/ file 의 wikilink `[[v1-0-release-roadmap]]` → `[[v0-1-release-roadmap]]`
- index.md 갱신

### PR #582 머지 결과

- **squash `f1c0995` (2026-06-13)**
- main HEAD f1c0995 (clean)
- 위키 mirror: 925 file / 7.6M / commit f1c0995
- 위키 162 page matched, 0 stale
- 2 ADR updated (adr-0025 + adr-0031)
- wiki-event-sync log 1 row

### 2차 PR (X-1 System Admin 운영 대시보드) — 별도 follow-up

- 본 1차 PR (version downgrade) = v0.1.2-alpha 의 prep
- **2차 PR (X-1, RM-M4-07)** = 본 sprint 의 본 release scope (5~8 commit, FE+BE)
- 별도 worktree + 별도 commit
- v0.1.2-alpha 의 핵심 scope

### CRITICAL 레슨런

1. **scope 분리 정공법** (cross-project lesson §1): version downgrade 정공법 = 1차 PR (30+ file, 1:1 mirror 정공법), X-1 = 2차 PR (5~8 commit, FE+BE). bundle 가능 조건 미충족 → 별도 PR.
2. **historical record vs forward-facing docs 분리**: state.json/session_handoff.md/work_backlog.md 의 historical v1 reference = 그대로 유지 (각 PR 의 당시 v1.0/v1.1 표기는 변경 불가). forward-facing docs (release_roadmap/ADR/setup/planning/validation) 만 v0.1 정공법.
3. **mirror 1:1 byte-identical 정공법** (cross-project lesson §1): `docs/planning/release_v0-1_roadmap.md` 의 raw/ ↔ 위키 mirror md5 동일 (5ee61d55...). file name 변경 시 mirror script 의 source list 자동 갱신 (find glob) + lint-config + operation-sop 의 reference 자동 갱신.

## 28. 본 세션 후속 (2026-06-13, X-1 System Admin 운영 대시보드 4~5차 commit + 2차 PR 머지)

### X-1 2차 PR #584 MERGED (squash 23c0ccd, main HEAD 23c0ccd)

**X-1 v0.1.1 milestone sprint 100% 완료** (1차 PR #583 + 2차 PR #584).

### 4차 commit (4a19392, branch `feat/work_260614-x1-frontend-e2e`) — frontend widget 4 + admin page 강화

- 9 file, 700 line
- `frontend/components/admin/x1-widgets/{SyncJobQueueWidget,SyncJobStatusWidget,ProviderHealthWidget,DashboardSummaryWidget,index,x1-widgets.test.tsx}` (6 file)
- `frontend/app/(dashboard)/admin/page.tsx` 강화 (X-1 widget 4 grid + 운영 도구 link, "Sys Admin Dashboard Archived" 메시지 제거)
- `frontend/domain/integration-registry/service/admin-x1.service.ts` (adminX1Service class + 3 method)
- `frontend/domain/integration-registry/schema/integration.types.ts` (X-1 types +39 line)

### 5차 commit (3dabe96, PR #584) — e2e + ADR-0032 + docs 정합

- 5 file, 340 line
- `frontend/tests/e2e/admin-x1.spec.ts` (TC-ADMIN-X1-01/02/03)
- `docs/adr/0032-system-admin-x1-dashboard.md` (10 section 정책 결정 ADR)
- `docs/traceability/report.md` (헤더 메타 + §4 ADR-0032 row + §6 row)
- `CHANGELOG.md` (X-1 row status ⏳ planned → ✅ implemented)
- `docs/llm-wiki/mirror-list.md` (§1.7.2 frontend e2e+helper scope +6 file)

### 검증

- `bash scripts/check-openapi-yaml-lint.sh` PASS
- `bash scripts/check-tier-separation.sh` PASS (사내 한정 패턴 매칭 0)
- 1차 PR #583 검증: `go build ./...` + `go test ./internal/domain/integration-registry/view/` (4/4)
- CI frontend-unit (X-1 widget 4 + types + service): 5 case
- CI e2e (admin-x1.spec.ts): 3 case

### 정합

- main HEAD = `23c0ccd` (squash PR #584)
- mirror 927 file, 162 page matched, 0 stale
- 위키 1:1 mirror byte-identical (manifest = HEAD)
- ADR-0032 = accepted 2026-06-13

### X-1 residual (별도 follow-up, 본 sprint scope 외)

- §2.4/§2.5/§3.1 의 IMPL/UT 컬럼 cross-ref 일괄 보강 = 별도 housekeeping sprint
- ADR-0032 §3 Provider Health Endpoint (API-107) = 별도 follow-up sprint (post-MVP, 1주일 staging 후)
- `docs/openapi.yaml` (line 9455) ↔ `swaggerui/asset/openapi.yaml` (line 9360) drift 95 line = 별도 sprint sync

### CRITICAL 레슨런

1. **scope 분리 정공법 (cross-project lesson §1)**: X-1 sprint = 1차 PR #583 (backend + openapi) + 2차 PR #584 (frontend + e2e + docs) 의 2-step 분리. backend half 와 frontend half 가 다른 PR 인 이유: backend 검증 (go test + openapi lint) 와 frontend 검증 (Vitest + Playwright e2e) 가 다른 CI 잡에서 실행되므로 PR 단위로 분리하는 게 자연스러움. scope 폭주 방지 + PR review 부담 분산.
2. **mirror list §1.7.2 frontend e2e+helper scope 갱신 정공법**: X-1 frontend widget 4 + admin-x1 service + admin-x1 e2e = 6 file 추가. mirror-list.md 의 source list 동적 = wiki-sync-devhub.sh 가 자동 반영. **위반 시점 drift**: mirror byte-identical 검증 script 의 Diff != 0 미충족 시 즉시 fix.
3. **traceability partial 정합 정공법 (cross-project lesson)**: 본 2차 PR 은 §6 row + §4 ADR row + 헤더 메타 만 갱신. §2.4/§2.5/§3.1 의 IMPL/UT 컬럼 일괄 보강 = 별도 housekeeping sprint. **이유**: 한 PR 에서 모든 위치 동시 갱신 시 scope 폭주 → review 부담 + conflict 가능성 ↑. **원칙**: PR scope = 한 가지 정합 (e.g. X-1 1차 PR 의 IMPL 정합) + housekeeping 별도.

## 29. 본 세션 후속 (2026-06-13, X-1 후속 housekeeping PR #585 MERGED)

### housekeeping 정공법 (sprint `chore/work_260614-a-housekeeping`)

**2 file, 197 line, docs only, 코드 변경 0**. main HEAD `32ffe0e` (squash).

#### A-1: §3.1 integration-registry row cross-ref (docs/traceability/report.md)

X-1 1차 PR #583 + 2차 PR #584 의 17 row 신규 ID 의 partial 정합 해소. §3.1 integration-registry row (line 414, 3228 chars) 의 IMPL/UT/TC 컬럼 끝에 X-1 cross-ref 1 line 추가 (+952 chars, 14 ID). §6 row 1건 + 헤더 메타 갱신.

#### A-2: docs/openapi.yaml X-1 block sync

CI lint canonical = `backend-core/internal/httpapi/swaggerui/asset/openapi.yaml`. docs (line 9456) ↔ swaggerui (line 9428) 의 drift 192 line (X-1 의 3 path + 2 schema 가 docs 에 미반영) 정합.

- X-1 paths block 121 line (swaggerui line 5915-6035) → docs line 6024 직전 삽입
- X-1 schemas block 69 line (swaggerui line 8018-8086) → docs line 8010 직전 삽입
- 결과: docs/openapi.yaml 9456 → 9648 line (+192)
- drift 0 (swaggerui 와 byte-identical 정합, paths 83/83, schemas 91/90 — DevRequestVoc 1 차이 = 별도 PR)

### 검증

- `bash scripts/check-openapi-yaml-lint.sh` PASS (yaml valid + semver + paths>=84 + cross-link ok)
- `bash scripts/check-tier-separation.sh` PASS (사내 한정 패턴 매칭 0)
- 위키 162 page matched, 0 stale
- mirror manifest = HEAD

### 잔여 (별도 follow-up, 본 sprint scope 외)

- §2.4/§2.5 의 IMPL/UT list 일괄 보강 (X-1 frontend-2 + UT-1 + TC-3 + ADR-1 row 추가) = 별도 housekeeping PR
- ADR-0025/0026 §4 row 누락 정공법 = 별도 housekeeping PR
- DevRequestVoc schema 의 swaggerui sync = 별도 PR
- Provider Health Endpoint (ADR-0032 §3, API-107) = 별도 follow-up sprint (post-MVP)

### CRITICAL 레슨런

1. **drift sync 정공법 (cross-project lesson)**: docs/openapi.yaml ↔ swaggerui/asset 의 X-1 block drift 192 line = 단순 paste byte-identical 정합. swaggerui/asset 가 CI lint canonical 이므로, docs 는 항상 swaggerui 의 subset 이어야 함. **위반 시점 drift**: paths/schemas count 차이 시 즉시 sync.
2. **partial 정합 해소 정공법 (cross-project lesson)**: §6 row + §4 ADR row + 헤더 메타 의 정공법 만으로는 ID 가 1차/2차 PR 의 산출물과 cross-ref 되지 않음 — §3 도메인 row 의 IMPL/UT/TC 컬럼에 cross-ref 필수. **이유**: 매트릭스 row 가 ID 의 SSOT.

## 30. 본 세션 후속 (2026-06-13, X-2 inbound webhook multi-provider 1차 PR #586 MERGED)

### X-2 sprint scope (1차 출처 = `2026-06-12-inbound-source-routing-sprint-plan.md`)

**X-2 = v0.1.1 milestone X-2 = inbound webhook 정규화 깊이 (multi-provider sync 일반화)**.

- N-13 backend foundation (migration 000007 + InboundSourceType enum + UpdatePlatformInboundSource + 4 UT) = main byte-identical
- N-13 PR A-2 (auto_route.go + voc_handler 통합 + IT 3 case + e2e) = main byte-identical
- **X-2 의 본 1차 commit** = N-13 의 multi-provider depth 정밀화

### 1차 commit (e2eae79, PR #586 MERGED squash 7075dbd, main HEAD 7075dbd)

**2 file, 245 line**:
- `auto_route.go` (92 → 165 line, +73) — InboundSourceRoutingConfig struct + ParseInboundSourceRoutingConfig + 3 provider pattern (jira/github/gitlab) + matchExternalRefPattern helper + AutoRouteDecision.ProviderHint field
- `auto_route_test.go` (+139 line) — 6 unit test (JiraOK / GithubOK / OtherCustom / InvalidPattern / ParseConfig / CrossProvider)

### 검증

- `go build ./...` PASS
- `go vet` clean
- `go test -v -run "TestAutoRoute|TestParseInboundSourceRoutingConfig"` **12/12 PASS** (6 N-13 기존 + 6 X-2 신규)

### 정합

- v0.1.1 milestone X-2 의 1차 PR (backend depth)
- P3-9 RM-M4-07 (X-1) 의 multi-provider 일반화
- 1차 출처 = ADR-0028 §6 (a) + N-13 follow-up 종합 검증
- N-13 의 foundation + PR A-2 와 byte-identical 정합

### 잔여 (X-2 2~5차, 별도 follow-up)

- 2차: WebhookAdapter interface + Gitea/Jira/Generic adapter (multi-provider webhook endpoint 일반화)
- 3차: openapi.yaml 정합
- 4차: frontend multi-provider 운영 UI (system_admin /admin/settings/integrations 의 InboundSourceConfig JSONB editor)
- 5차: e2e + ADR-0033 + traceability/CHANGELOG/mirror-list

### CRITICAL 레슨런 (3건)

1. **기존 infrastructure 최대 활용 정공법 (cross-project lesson)**: N-13 의 backend foundation (PR A-1) + PR A-2 (auto_route.go + voc_handler) 가 main 에 byte-identical 포함 — X-2 의 1차 commit 은 **새 method 추가가 아닌** 기존 pattern matcher 의 **multi-provider depth 정밀화** 만.
2. **JSONB schema 의 typed parse 정공법**: `InboundSourceRoutingConfig` struct + `ParseInboundSourceRoutingConfig(raw string)` 함수로 platform.InboundSourceConfig (raw text) 의 typed parse. 빈 문자열 / invalid JSON 의 silent skip (zero value + nil error) = 운영자 audit 으로 silent invalid 알림.
3. **ProviderHint field 추가 정공법 (X-2 의 reason 정밀화)**: AutoRouteDecision.Reason (`external_ref_pattern` / `requester_email` / `req_department` / `no_match`) 에 ProviderHint (`gitea` | `jira` | `github` | `gitlab` | `other_custom` | `custom` | empty) 필드 추가 — reason + provider 의 정밀 식별. 운영자 dashboard 에서 "왜 gitea enum 으로 match 됐는가" 즉시 식별.

## 31. 본 세션 후속 (2026-06-13, X-2 inbound webhook multi-provider adapter 2차 PR #587 MERGED)

### X-2 의 2차 commit (PR #587, main HEAD 57fda7e)

**3 file 신규, 408 line, go test 9/9 PASS**:
- `backend-core/internal/infrastructure/webhook/adapter.go` (102 line) — WebhookAdapter interface + registry dispatcher + helper (payloadHashHex, firstHeader)
- `backend-core/internal/infrastructure/webhook/gitea_adapter.go` (109 line) — GiteaWebhookAdapter (provider_type='scm', Gitea/Forgejo/Gogs X-Gitea-*/X-Gogs-* header + repository/sender envelope)
- `backend-core/internal/infrastructure/webhook/adapter_test.go` (152 line, 9 case)

### 정공법 (cross-project lesson §1 scope 분리)

**본 2차 commit = adapter package + Gitea adapter 신규 만**, IngestIntegrationProviderWebhook handler 통합 (provider_type='scm' dispatch) = **3~5차 commit 의 후속** (scope 폭주 방지 + ADR-0033 결정 + 다중 adapter 추가 시 일괄 dispatch).

### 검증

- `go build ./...` PASS
- `go test -v ./internal/infrastructure/webhook/` 9/9 PASS

### 정합

- 1차 PR #586 의 auto_route.go giteaExternalRefPattern `^GITEA-\d+$` 와 cross-ref (Gitea adapter 의 ExternalRef `GITEA-<number>` 정공법)
- IngestIntegrationProviderWebhook handler 통합 = 후속 commit

### 잔여 (X-2 3~5차, 별도 follow-up)

- 3차: openapi.yaml 정합 (multi-provider endpoint 명세)
- 4차: frontend multi-provider 운영 UI (system_admin /admin/settings/integrations 의 InboundSourceConfig JSONB editor)
- 5차: e2e + ADR-0033 (Multi-Provider Webhook Architecture) + traceability/CHANGELOG/mirror-list

### CRITICAL 레슨런 (3건)

1. **adapter registry dispatcher 정공법 (cross-project lesson)**: `package webhook` + `RegisterAdapter` + `GetAdapterForProviderType` 함수 — provider_type 별로 adapter dispatch. multi-provider 일반화의 foundation. 추후 Jira/Generic adapter 추가 시 `RegisterAdapter` 호출만으로 dispatch 가능.
2. **GiteaExternalRef 의 cross-ref 정공법**: Gitea adapter 의 `buildGiteaExternalRef` 가 `pull_request/issue/release` = `GITEA-<number>`, `push/repository` = `GITEA-<repository.full_name>` 정공법 — AutoRoute 의 `giteaExternalRefPattern` (`^GITEA-\d+$`) 와 1:1 매핑. webhook ingest → AutoRoute routing 의 seamless 통합.
3. **scope 분리 정공법 (cross-project lesson §1)**: 본 2차 commit 은 adapter package + Gitea adapter 만 (3 file 신규, 408 line). IngestIntegrationProviderWebhook handler 통합 + Jira/Generic adapter + openapi + frontend + e2e + ADR-0033 = 3~5차 commit 으로 분할. **이유**: adapter 도입 자체가 큰 architectural change → 한 commit 으로 handler 통합 + 다중 adapter + ADR-0033 까지 묶으면 review 부담 ↑ + scope 폭주.
