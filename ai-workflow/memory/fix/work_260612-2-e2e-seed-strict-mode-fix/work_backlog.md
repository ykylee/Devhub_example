# Work Backlog — fix/work_260612-2-e2e-seed-strict-mode-fix (N-13 follow-up A)

- 문서 목적: N-13 follow-up A (Test 1 e2e seed 중복 strict mode violation fix) 의 작업 백로그.
- 범위: 1 file (frontend e2e spec) + 메모리 4 file 동기화 + PR 발행 + 머지 + main flat memory finalize.
- 대상 독자: 본 sprint 작업자 + 후속 PR 머지 후 세션 진입자.
- 상태: **2026-06-12 — 파일 변경 작업 완료, PR 발행 전 (사용자 confirm 대기)**
- 최종 수정일: 2026-06-12

## 1. 마일스톤

| 마일스톤 | 상태 | 비고 |
| --- | --- | --- |
| **M-v1.0** | in_progress | 본 sprint = N-13 follow-up A (Test 1 fix). Test 2 + 구현 follow-up = 후속 별도 sprint |

## 2. 작업 항목

| # | 항목 | 상태 | 비고 |
| --- | --- | --- | --- |
| 1 | PR #573 머지 후 main flat memory finalize (state.json main HEAD `5fb9ae75` 갱신) | ✅ done | 직전 sprint 마무리 |
| 2 | Test 1 e2e seed 중복 strict mode violation 분석 (5 spec file + global-setup.ts 검토) | ✅ done | root cause = `repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher `.first()` 미적용 |
| 3 | 새 브랜치 생성 | ✅ done | `fix/work_260612-2-e2e-seed-strict-mode-fix` |
| 4 | `repositories-ui.spec.ts:5, 7` `.first()` 추가 + 코멘트 1 line | ✅ done | 1 file, +3/-2 |
| 5 | 브랜치 memory 4 file 신규 | ✅ done | state.json + session_handoff.md + work_backlog.md + backlog/2026-06-12.md |
| 6 | **PR 발행** (사용자 confirm 후) | ⏳ pending | branch push + PR 발행 + squash merge |
| 7 | **main flat memory finalize** (머지 직후) | ⏳ pending | main HEAD `5fb9ae75` → 신규 squash commit 갱신 |

## 3. follow-up 잔여 (사용자 결정 영역)

| Branch | 대상 | 결정 |
| --- | --- | --- |
| 1 | **Test 2 Sign-out timeout** (N-8 race 유사) | main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑) |
| 2 | **구현 follow-up** | v1.1 milestone 진입 시점 별도 sprint (rebase main + PR #550 fix + 본 fix 종합 + 자동 재실행) |

## 4. 직전 sprint (`fix/work_260612-1-n13-housekeeping-followup`)

- PR #573 ✅ MERGED (2026-06-12, squash `5fb9ae75`) — docs only, 11 file +249/-9
- N-13 housekeeping follow-up (PR #548 close 결과 정공법 + 3 branch follow-up 결정)
- 본 sprint 의 직전 정공법

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-12 | 1차 작성 — N-13 follow-up A (Test 1 e2e seed 중복 strict mode violation fix). 1 file 변경 (frontend/tests/e2e/repositories-ui.spec.ts). |
