# Work Backlog — docs/x6-keycloak-groups-public

- 문서 목적: X-6 (Keycloak group staging-prod 적용, issue #214) 의 사외 PR work backlog
- 범위: ADR-0036 + sprint plan + memory 4 file (Tier: 공용)
- 상태: 작업 완료, push + PR 발행 pending
- 최종 수정일: 2026-06-14

## 0. 현재 status

**main HEAD**: `fc2d2a0` (PR #590 X-2 100% 완료 시점)
**branch**: `docs/x6-keycloak-groups-public` (worktree `.worktrees/x6-keycloak-groups-public/`)
**sprint scope**: X-6 (Keycloak group staging-prod 적용, issue #214) — 사외 PR (공용) + 사내 PR (사내)

## 1. 본 sprint 진행 상황

| Step | 작업 | 결과 |
|---|---|---|
| 1 | worktree 생성 (`docs/x6-keycloak-groups-public`) | ✅ HEAD 기준 |
| 2 | ADR-0036 (X-6 Architecture, 9 section) | ✅ |
| 3 | sprint plan (2-PR 분할) | ✅ |
| 4 | 메모리 4 file | ✅ |
| 5 | commit + push + PR 발행 | ⏳ pending |

## 2. 잔여 follow-up (사외 PR 외)

| Step | 작업 | 주체 |
|---|---|---|
| 6 | 사내 PR (`infra/x6-keycloak-groups-internal`) — `scripts/setup-keycloak-groups.sh` + runbook | 사용자 (사내 Gitea push) |
| 7 | staging Keycloak admin console 적용 | 사용자 (수동) |
| 8 | prod Keycloak admin console 적용 | 사용자 (staging 1주 후) |
| 9 | 운영자 수동 retry API | follow-up sprint |
| 10 | e2e spec | follow-up sprint |

## 3. 다음 sprint

| ID | 아이템 | 영역 | effort |
|---|---|---|---|
| X-8 | Keycloak SPI realm events push (P2-6/P3-5) | BE/사내 | 2~3 ses |
| X-4 Phase 2 | handler post-commit wire + openapi + main.go | BE | 1.1 ses |

## 4. 변경 이력

| 일자 | 변경 | 비고 |
|---|---|---|
| 2026-06-14 | X-6 사외 PR (ADR-0036 + sprint plan + memory) 신규 (Tier: 공용) | 본 sprint |
