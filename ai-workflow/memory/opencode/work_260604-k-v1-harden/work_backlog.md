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
| WK-12 | PR #476 CI 실패 수정 (rename 잔여 SQL/test cleanup + E2E locator) | ✅ done |

## 2. 사전 확인 결과

| 아이템 | 상태 | 근거 |
| --- | --- | --- |
| P0-1 (accounts 폐기) | ✅ 완료 | router.go에 accounts endpoint 없음, lazy_auto_create.go 삭제됨, frontend account.service.ts 없음 |
| N-10 (RBAC E2E TC) | ✅ 완료 | `rbac-data-scope.spec.ts` 4 TC 머지됨 (`cbd375b`) |

## 3. PR #476 CI 수정 내역 (WK-12)

### 3.1 CI 실패 진단

PR #476 의 3개 job 실패 (run 26950775348):
- **Backend Integration Tests**: `internal/store/integrations_integration_test.go:15: cleanup static tables: ERROR: relation "applications" does not exist (SQLSTATE 42P01)` → fixture panic → 10분 timeout
- **E2E shard 1/2**: catalog/projects/repository test 다수 실패 (testid 미스매치, 행 미발견, 500 에러)
- **E2E shard 2/2**: SCM flow strict mode violation, repository row 미발견

### 3.2 근본 원인

Application→Platform rename (migration 000048) 시 4종 SQL이 누락:

| 파일 | 라인 | 문제 |
| --- | --- | --- |
| `backend-core/internal/store/integrations.go` | 30, 62, 75, 116 | `project_integrations.application_id` 컬럼 참조 (→ `platform_id`) |
| `backend-core/internal/store/repository_ops.go` | 287, 289 | `application_repositories` 테이블 + `ar.application_id` |
| `backend-core/internal/store/postgres.go` | 1423 | `application_repositories` 테이블 (ListRepositories) |

영향: `/api/v1/repositories` → 500 → catalog/projects/repository E2E 모두 연쇄 실패. Integrations endpoint들도 500.

E2E의 strict mode violation은 신규 TC-REPO-SCM-IMPORT-01 의 `/import/i` regex가 displayName "E2E SCM Import …" 가 5개 버튼명 모두에 들어가서 매칭.

### 3.3 적용 수정

| # | 파일 | 변경 |
| --- | --- | --- |
| 1 | `internal/store/integrations.go` | `application_id` → `platform_id` (4 곳) |
| 2 | `internal/store/repository_ops.go` | `application_repositories` → `platform_repositories` + `ar.platform_id` |
| 3 | `internal/store/postgres.go` | `application_repositories` → `platform_repositories` |
| 4 | `internal/store/integration_test_helpers_test.go` | TRUNCATE `applications` → `platforms` |
| 5 | `internal/domain/application-lifecycle/repository/applications_integration_test.go` | TRUNCATE 1 + INSERT/DELETE 4 |
| 6 | `internal/domain/dev-request/repository/dev_requests_integration_test.go` | DELETE FROM `applications` → `platforms` |
| 7 | `frontend/tests/e2e/repositories-scm-flow.spec.ts` | `name: /import/i` → `name: /^import/i` (2 곳) |

### 3.4 검증

- `go build ./...` clean
- `go test ./...` 전체 유닛 테스트 PASS
- `go test -run 'TestIntegration_' ./internal/store/...` PASS (CI 가 실행하는 scope)
- `npx tsc --noEmit` 변경 파일 에러 없음 (기존 `page.test.tsx` / `OrgTree.test.tsx` pre-existing TS 에러는 무관)

### 3.5 CI scope 외 (별도 후속)

`./internal/domain/application-lifecycle/repository/...` 통합 테스트에서 pre-existing 버그 2건 노출:
- `TestIntegration_DeleteRepository:911` — repository_status 가 publish_requested 시에도 'draft' 유지 (MarkRepositoryDraftPublishRequested 가 status 컬럼 미갱신)
- `TestIntegration_DeleteRepository:929` — "del-fk-app" 가 `platforms_key_format` (^[A-Za-z0-9]{1,10}$) 위반 (hyphen 미허용)

→ rename 자체와 무관. CI 는 `./internal/store/...` 만 실행하므로 PR 머지 차단 안 함. 별도 sprint 에서 status state-machine + key format 정리 필요.
