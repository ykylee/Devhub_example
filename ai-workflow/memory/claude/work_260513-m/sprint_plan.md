# Sprint Plan — claude/work_260513-m (M3 후속 1-4 일괄)

- 문서 목적: M3 후속 carve out 1-4 항목 일괄 진행.
- 진입 base: main HEAD `b1268ce` (PR #98 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

| # | 항목 | 위치 | 규모 |
|---|------|------|------|
| 1 | ADR-0008 PostgresClient + hrdb.persons migration | `internal/hrdb/postgres.go` + `migrations/000010_*` | M |
| 2 | parent_id cycle 검증 + primary_dept 자동 판정 1차 | `internal/httpapi/organization.go` + 단위테스트 | M |
| 3 | 파견/겸임 1:N + total_count MV (1차 schema) | `migrations/000011_*` + store 1차 | M |
| 4 | frontend Sign Up form | `frontend/app/(public)/signup/page.tsx` + auth.service.ts | M |
| 5 | 매트릭스 + sync + PR + 2-pass + merge | docs/traceability/report.md + ai-workflow/memory/ | S |

## 2. 1차 구현 범위 vs carve

각 항목은 1차 구현 + 단위테스트 1-2 case + carve out 명시 패턴.

### 1 PostgresClient
- migration up/down + table + indexes + COALESCE email fallback
- postgres.go::PostgresClient (Lookup 메서드 1개)
- 단위테스트 2 case (hit + miss)
- ETL / sync cron carve

### 2 cycle 검증 + primary_dept
- updateOrgUnit::cycleDetected 함수 (BFS or DFS)
- 422 cycle_detected 응답
- primary_dept 자동 판정: 단순 규칙 = leader 인 unit 우선, 동급 시 자식 노드 수 많은 unit
- 단위테스트 2 case (cycle / non-cycle)
- 복잡 edge case + 파견 우선순위 통합 검증 carve

### 3 파견/겸임 1:N + total_count MV
- migration: unit_secondary_memberships 테이블 + total_count MV
- store interface 추가 (AddSecondaryMember / RemoveSecondaryMember)
- 실 endpoint + cron 갱신 + RBAC 정책 carve

### 4 frontend Sign Up form
- /signup 페이지 (publicRoutes 등록 carve)
- form (name + system_id + employee_id + password) + auth.service.ts::signup() helper
- API client → POST /api/v1/auth/signup
- 단위테스트 1 (auth.service.signup happy path mock)

## 3. 검증

- backend `go test ./...` PASS (1, 2, 3 의 신규 테스트 포함)
- frontend `npm test` PASS (4 신규 1 test)
- CI 4 잡 SUCCESS
- migration up/down SQL syntax check (golang-migrate dry-run 또는 직접 검토)

## 4. carve out (다음 sprint)

- migration apply + 실 운영 schema 적용
- PostgresClient ETL cron
- cycle 검증 모든 edge case (deep parent chain, race)
- primary_dept 자동 판정 의 파견 우선순위 통합
- secondary memberships endpoint + RBAC 정책
- frontend Sign Up form 의 e2e spec ts + 더 세밀한 validation
- M4 sprint plan 진입
