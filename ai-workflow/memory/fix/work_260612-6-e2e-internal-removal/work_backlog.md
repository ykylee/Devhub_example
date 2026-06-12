# Work Backlog — fix/work_260612-6-e2e-internal-removal (2026-06-12, e2e-internal job 폐기)

- sprint branch: fix/work_260612-6-e2e-internal-removal
- based on: main @ 379e894 (PR #577 E2E Internal disable + 후속 memory finalize)
- status: in_progress (PR 발행 대기)
- 최종 수정일: 2026-06-12

## 본 sprint 결정 (사용자 2026-06-12)

> "e2e-internal은 어차피 사내 환경용 셋팅이면 github action으로 체크할건 아니야 그냥 없애줘."

**결론**:
1. e2e-internal job 자체 폐기 (ci.yml 에서 완전 삭제)
2. ADR-0030 §2.3 runtime injection 결정 = 유지 (e2e-internal 폐기와 독립)
3. ADR-0031 partial supersession (baseline 변경 정공법)
4. 사내 staging/prod-smoke 가 real adapter 검증 책임

## Tasks

| ID | Task | Status | Priority | 비고 |
|---|---|---|---|---|
| T-1 | ci.yml e2e-internal job 삭제 | done | P0 | line 554-756 (-205 lines). YAML parse PASS |
| T-2 | ci-e2e-sync-check.sh DEVHUB_BUILD_TIER 코멘트 정리 | done | P0 | 5 lines 정리 |
| T-3 | ADR-0031 partial supersession (baseline 변경 정공법) | done | P0 | §1.2 / §2.x / §3.x / §4 / §5 / §6.1 / §7.1 / §8 baseline 갱신. 결론 변동 0건 |
| T-4 | ADR-0030 §2.3 + §9 reference 갱신 | done | P1 | partial supersession 정공법 reference |
| T-5 | traceability report.md §6 row 추가 | done | P1 | 1 row 추가 |
| T-6 | 메모리 4 file (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-12.md) | done | P1 | 본 sprint 정공법 보고 |
| T-7 | worktree commit + push + gh pr create --body-file | pending | P0 | 사용자 confirm 대기 |
| T-8 | main flat memory 4종 sync (post-merge) | planned | P1 | 본 PR 머지 후 자동 sync |

## Follow-up (사용자 결정 영역)

- e2e-internal 폐기 결정의 사내 staging/prod-smoke 운영 SOP 갱신 (별도 docs, 본 sprint scope 외)
- ADR-0030 / ADR-0031 외 다른 doc 의 e2e-internal reference 정합 (본 sprint 1차 5 file, 잔여 없음 — N-13 follow-up 의 "E2E Internal 1 fail" historical record 는 별개)
- v1.1 milestone 진입 시점: PR #548 의 구현 follow-up (e2e-internal 없이 검증 정합)
