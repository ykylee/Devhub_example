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
| T-7 | e2e voc-auto-routing.spec.ts (TC-INBOUND-SRC-01, seed platform 사용) | pending | P0 | 157 lines. PR #574 + #575 + #578 영향 없음 |
| T-8 | ADR-0028 §6 (a) amendment (구현 정합) | pending | P1 | +20/-10 lines |
| T-9 | release_v1_roadmap.md §3.5 N-13 row + §4.2 v1.1 milestone + §9 | pending | P1 | +15/-3 lines |
| T-10 | traceability report.md 9 ID row status `planned` → `implemented` | pending | P1 | cell fill 만 |
| T-11 | `go build ./...` + `go test ./...` (변경 package + 전체 회귀) | pending | P0 | sprint 정합 검증 |
| T-12 | `bash scripts/check-tier-separation.sh` PASS | pending | P1 | tier-separation lint |
| T-13 | `bash scripts/check-openapi-yaml-lint.sh` PASS | pending | P1 | openapi lint |
| T-14 | worktree commit + push + gh pr create --body-file | pending | P0 | 사용자 confirm 대기 |
| T-15 | main flat memory 3 file sync (post-merge) | planned | P1 | 본 PR 머지 후 자동 sync |

## Follow-up (사용자 결정 영역, 본 sprint 후속)

- v1.1 sprint -b (gitea + ci port) 진입 결정
- v1.0 staging 1주 운영 (N-6) 시작 (사용자 결정)
- PR #548 의 신규 branch 이름 (본 sprint 가 정식 1차 구현)
- v0.1.1-alpha release 의 8 item (T-d-72-5/6 + D-73/74 + X-1~8) 진행 방향
- N-10 housekeeping close 정공법
- backend-integration DEVHUB_BUILD_TIER matrix (obsolete — e2e-internal 폐기 결정 정합)
