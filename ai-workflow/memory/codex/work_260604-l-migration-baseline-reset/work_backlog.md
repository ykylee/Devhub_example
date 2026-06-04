# Work Backlog — codex/work_260604-l-migration-baseline-reset

- 문서 목적: 현재 브랜치의 workflow 작업 항목과 상태를 관리한다.
- 범위: migration baseline reset
- 대상 독자: 구현 담당자, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-06-04

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| MIG-RESET-01 | baseline reset 계획 문서화 | done | `docs/planning/migration_baseline_reset_plan_2026-06-04.md` 작성 |
| MIG-RESET-02 | 현재 migration 최종 상태 검증 | done | `000001..000047` empty DB replay 성공 확인 |
| MIG-RESET-03 | active baseline chain 재구성 | done | schema + rbac seed + bootstrap catalog seed 3단계 |
| MIG-RESET-04 | legacy migration archive 보존 | done | `backend-core/migrations-legacy-20260604/` 이동 |
| MIG-RESET-05 | migration/문서 정합 검증 | done | migrate up/down, lint, backend integration/unit PASS |
| MIG-RESET-06 | memory 갱신 + PR 생성 | in_progress | branch memory 작성 완료, 커밋/PR 대기 |
