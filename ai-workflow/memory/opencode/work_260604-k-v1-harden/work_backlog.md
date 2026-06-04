# Work Backlog — opencode/work_260604-k-v1-harden

- Branch: `opencode/work_260604-k-v1-harden`
- Agent: opencode (Sisyphus)
- Updated: 2026-06-04

## 0. Sprint 목표

v1.0 마무리 5종 병렬 작업 + Application→Platform 전면 리네임.

## 1. 작업 단위

| ID | 작업 | 상태 |
| --- | --- | --- |
| WK-01 | P0-1 ADR-0020 sub-carve B 완료 확인 | ✅ done |
| WK-02 | N-10 P1 follow-up TC 구현 확인 | ✅ done |
| WK-03 | N-5 CI guard 강화 (migration 순차/중복 검증 보강) | ✅ done |
| WK-04 | Governance 문서 sync (main state/handoff/backlog) | ✅ done |
| WK-05 | Traceability report 정합 (report.md 갱신) | ✅ done |
| WK-06 | N-2: Repository draft→publish UT 보강 | ✅ done |
| WK-07 | P1-6: Keycloak session revocation | ✅ done |
| WK-08 | N-3: SCM import/create E2E test spec | ✅ done |
| WK-09 | Application→Platform 전면 리네임 | ✅ done |
| WK-10 | Codex review 반영 + Conflict 해결 | ✅ done |
| WK-11 | 검증 + PR | ✅ done |

## 2. 사전 확인 결과

| 아이템 | 상태 | 근거 |
| --- | --- | --- |
| P0-1 (accounts 폐기) | ✅ 완료 | router.go에 accounts endpoint 없음, lazy_auto_create.go 삭제됨, frontend account.service.ts 없음 |
| N-10 (RBAC E2E TC) | ✅ 완료 | `rbac-data-scope.spec.ts` 4 TC 머지됨 (`cbd375b`) |
