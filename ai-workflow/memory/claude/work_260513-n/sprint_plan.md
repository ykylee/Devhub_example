# Sprint Plan — claude/work_260513-n (M4 전 잔여 일괄)

- 문서 목적: M4 진입 직전 RM-M3 carve out 의 잔여 항목 일괄 처리.
- 진입 base: main HEAD `c7c2f35` (PR #99 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

| # | 항목 | 위치 | 규모 |
|---|------|------|------|
| A | ETL seed SQL + ADR-0008 §6 갱신 | `scripts/hrdb_etl_seed.sql` + ADR-0008 §6 | S |
| B | total_count MV migration 000011 | `migrations/000011_create_org_units_total_count_mv.{up,down}.sql` | S |
| C | ADR-0010 primary_dept 자동 판정 + 1차 알고리즘 | `docs/adr/0010-*.md` + `internal/domain/*.go` 또는 `internal/store/*.go` + 단위테스트 | M |
| E | frontend Sign Up e2e spec ts | `frontend/tests/e2e/signup.spec.ts` | S+ |
| F | 매트릭스 + sync + PR + 2-pass + merge | docs + ai-workflow | S |

## 2. carve out 명시

- **D (getHierarchy MV join 코드 변경)**: backend store 인터페이스 변경 + memory store 동기화 + 통합 테스트 환경 의존. 본 sprint 의 큰 영역으로 분리 — M4 진입 후 또는 별도 인프라 sprint.
- ETL daily cron 운영: scripts 의 PoC seed 만, 운영 cron entry 는 운영 sprint.
- primary_dept 알고리즘의 모든 edge case (파견 우선순위, joined_at 동률) carve.
- signup e2e happy path (Kratos identity 실제 생성 검증) carve — 본 sprint 는 form smoke + 403 negative 만.

## 3. 검증

- backend `go test ./...` PASS (primary_dept 단위테스트 포함)
- frontend `npm test` PASS
- e2e signup.spec.ts → CI 의 e2e 잡 SUCCESS
- migration 000011 SQL syntax check (직접 검토)
