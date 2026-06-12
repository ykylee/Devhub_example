# Work Backlog — feat/work_260612-7-v1-1-inbound-source-impl (2026-06-12, N-13 follow-up 종합 v1.1 sprint -a)

- sprint branch: feat/work_260612-7-v1-1-inbound-source-impl
- based on: main @ 0d2dd89 (PR #578 e2e-internal job 폐기 + 메모리 finalize)
- status: in_progress
- 최종 수정일: 2026-06-12

## 본 sprint 결정

> N-13 follow-up 3 branch 종합 (A: e2e seed fix [PR #574], B: signout timeout fix [PR #575], C: 구현 follow-up = 본 sprint). PR #548 (1차 구현 CLOSED) 의 rebase + PR A-1 (backend foundation) + PR A-2 (routing + voc_handler + openapi + e2e) 종합.

## Tasks

| ID | Task | Status | Priority | 비고 |
|---|---|---|---|---|
| T-1 | sprint memory 4 file (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-12.md) | done | P0 | branch memory 신규 |
| T-2 | routing/auto_route.go (3 case pattern matcher + graceful degradation) | pending | P0 | 126 lines. sprint plan v2 §4.4 |
| T-3 | routing/auto_route_test.go (6 UT: ExternalRef / Requester / ReqDepartment / NoMatch / MultiplePlatforms / EmptyPlatforms) | pending | P0 | 166 lines |
| T-4 | voc_handler.go 통합 (createOrGetVoc 에 AutoRouter.Route() + RouteVoc() + auto_routed 응답) | pending | P0 | +114/-42 lines |
| T-5 | voc_handler_integration_test.go (3 IT: GiteaOK / NoMatch / RouteErrorDegradation) | pending | P0 | 221 lines |
| T-6 | openapi.yaml 정합 (PATCH /platforms inbound_source + POST /dev-requests/{id} auto_routed + DevRequestVoc schema) | pending | P0 | +216 lines |
| T-7 | e2e voc-auto-routing.spec.ts (TC-INBOUND-SRC-01, seed platform 사용) | done (1차) → done (2차) → done (3차) → done (4차, syntax fix) | P0 | 1차 73 → 2차 110 → 3차 112 → **4차 112 (1 line syntax fix)** |
| T-7b | **PR #579 2차 commit** — e2e beforeAll fix + 메모리 §6 append | done (2차) | P0 | 옵션 B 정공법 적용. **근본 layer 부족 분석: beforeAll hook timeout 30000ms default 가 loginAs + retry 합쳐서 60s+ 필요** |
| T-7c | **PR #579 3차 commit** — beforeAll timeout 180s 명시 (시도) | done (3차) | P0 | `test.beforeAll(async () => {...}, { timeout: 180_000 })` option 추가. **시그너처 오류** — Playwright 시그너처는 (fn, timeout?: number) |
| T-7d | **PR #579 4차 commit** — beforeAll timeout **number syntax** fix (시도) | done (4차) | P0 | 2번째 인자 `{ timeout: 180_000 }` (object) → `180_000` (number) 1 line fix. **시도였음** — loginAs 의 internal timeout 30s 가 먼저 fail, beforeAll timeout 180s 적용 안 됨. 5회 fail |
| T-7e | **PR #579 5차 commit** — loginAs timeout 60s + retry 5회 75s (근본 fix, 옵션 L) | done (5차) | P0 | `fixtures.ts::loginAs` 의 `page.waitForURL` timeout 30_000 → 60_000 (1 line) + `voc-auto-routing.spec.ts` 의 delays = [0, 5s, 10s, 15s] → [0, 5s, 10s, 15s, 20s, 25s] (retry 5회, 6 attempts). **6-step 종합 정공법** — loginAs internal 60s + retry 75s + beforeAll timeout 180s 정합 |
| T-8 | ADR-0028 §6 (a) amendment (구현 정합) | done | P1 | +20/-10 lines |
| T-9 | release_v1_roadmap.md §3.5 N-13 row + §4.2 v1.1 milestone + §9 | pending | P1 | +15/-3 lines |
| T-10 | traceability report.md 9 ID row status `planned` → `implemented` | pending | P1 | cell fill 만 |
| T-11 | `go build ./...` + `go test ./...` (변경 package + 전체 회귀) | pending | P0 | sprint 정합 검증 |
| T-12 | `bash scripts/check-tier-separation.sh` PASS | pending | P1 | tier-separation lint |
| T-13 | `bash scripts/check-openapi-yaml-lint.sh` PASS | pending | P1 | openapi lint |
| T-14 | worktree commit + push + gh pr create --body-file | pending | P0 | 사용자 confirm 대기 |
| T-15 | main flat memory 3 file sync (post-merge) | planned | P1 | 본 PR 머지 후 자동 sync |

## 1차 commit (PR #579, 2026-06-12 21:25 KST)

- sprint plan v2 정공법 그대로 — 5 file 신규 + 2 file 수정 = 10 file 변경 (코드 + ADR + openapi + e2e)
- CI 8/8 PASS (backend / frontend / lint), e2e shard 1/2 PASS
- **e2e shard 3/3 fail 2건** (TC-INBOUND-SRC-01 + NEG) — shard 3/3 의 Keycloak startup race (curl: (56) Recv failure × 9)
- codex review = COMMENTED (blocker 없음, 자동 review 만)
- 사용자 결정: **옵션 B (beforeAll hook fix) 진행** (재실행 ❌)

## 2차 commit (PR #579, 2026-06-12 21:42 KST) — 옵션 B beforeAll fix

- `frontend/tests/e2e/voc-auto-routing.spec.ts` (73 → 110 lines) — `test.beforeAll(async ({ browser }) => {...})` 신규 + retry 3 회 with backoff (5s + 10s + 15s, 최대 4 attempts)
- 2 test case 의 PATCH 단계 제거 — 검증만 (beforeAll 에서 1 회 PATCH 로 통합)
- beforeAll 의 context.close() 명시 — leak 방지. throw 시 last error 메시지 + 4 attempts 명시
- 메모리 §6 append + backlog T-7b row 추가
- 정공법 정합: N-13 follow-up 3 branch 결정 (PR #573) 의 종합 정공법 정합 (PR #574 + PR #575 + 본 sprint 1차 구현 + 본 sprint 2차 Keycloak race fix)
- **근본 layer**: PR #548 (1차) → PR #574/575/576 (fix 1차 fail 2건) → PR #579 1차 (구현 + 1차 fail 2건 fix 의 종합) → PR #579 2차 (shard 3/3 Keycloak race fix). 4-step 종합 정공법.

## Follow-up (사용자 결정 영역, 본 sprint 후속)

- v1.1 sprint -b (gitea + ci port) 진입 결정
- v1.0 staging 1주 운영 (N-6) 시작 (사용자 결정)
- PR #548 의 신규 branch 이름 (본 sprint 가 정식 1차 구현)
- v0.1.1-alpha release 의 8 item (T-d-72-5/6 + D-73/74 + X-1~8) 진행 방향
- N-10 housekeeping close 정공법
- backend-integration DEVHUB_BUILD_TIER matrix (obsolete — e2e-internal 폐기 결정 정합)
