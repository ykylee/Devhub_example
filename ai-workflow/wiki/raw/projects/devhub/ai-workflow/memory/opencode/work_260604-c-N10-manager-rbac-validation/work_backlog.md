# Work Backlog — opencode/work_260604-c-N10-manager-rbac-validation

- Branch: `opencode/work_260604-c-N10-manager-rbac-validation`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Source of truth: 본 파일

## 0. 본 sprint 의 목표

release_v1_roadmap §3.5 N-10 "Manager role RBAC 검증" — PR #462 + PR #461 의 RBAC 관련 변경 직후 회귀 검증. 발견 결함은 issue + fix proposal 로 처리 (본 sprint 는 fix scope out).

## 1. 작업 단위 분해

| ID | 작업 | 의존 | 상태 |
| --- | --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | — | done |
| WB-02 | manager RBAC 코드/테스트/시드 상태 탐색 | WB-01 | in_progress |
| WB-03 | 검증 대상 식별 + 검증 시나리오 설계 | WB-02 | planned |
| WB-04 | 검증 실행 (Keycloak seed + backend UT + E2E) | WB-03 | planned |
| WB-05 | 결과 문서화 (검증 보고서) | WB-04 | planned |
| WB-06 | 발견 결함 issue + fix proposal (선택) | WB-05 | planned |
| WB-07 | 커밋 + push + PR + state/handoff final | WB-05/06 | planned |

## 2. 검증 대상 (1차 식별)

| # | 대상 | 확인 방법 | 우선순위 |
| --- | --- | --- | --- |
| V-01 | Keycloak realm `team_manager` role 존재 | Keycloak admin API 또는 realm.json | P0 |
| V-02 | Keycloak realm `org_head` role 존재 | Keycloak admin API 또는 realm.json | P0 |
| V-03 | E2E seed `mgr-user-b` 가 `team_manager` realm role 보유 | E2E fixture + global-setup 확인 | P0 |
| V-04 | `ListProjects` row filter (team_manager subtree scope) | backend `go test ./internal/store/...` 또는 curl | P0 |
| V-05 | `ListApplications` row filter (team_manager subtree scope) | 동상 | P0 |
| V-06 | E2E `TC-RBAC-ROW-READ-01/02` (PR #461 추가) PASS | `npm run e2e` | P0 |
| V-07 | `role-access-concept.md` 가 3-role 모델 (system_admin / team_manager / user) 기준 정합 | docs/adr + role-access-concept.md vs code | P1 |
| V-08 | legacy `manager` alias 가 완전히 제거됨 (deprecated migration 000047 일관) | `grep -r "manager" backend-core/ --include="*.go"` | P1 |
| V-09 | `normalizeSystemRoleAlias` 가 legacy manager → team_manager 정규화 | backend UT | P1 |
| V-10 | team_manager 의 subtree scope 에서 privilege escalation 차단 (PR #462 review 반영) | backend UT `team_manager_test.go` | P1 |

## 3. 검증 도구 / 절차

- **Keycloak**: `scripts/setup-keycloak.sh` + realm.json export, 또는 staging `curl`
- **Backend UT**: `cd backend-core && go test ./internal/...` (전체 또는 rbac/store subset)
- **Backend 통합**: `go test -tags integration ./...` (DEVHUB_TEST_DB_URL)
- **E2E**: `cd frontend && npm run e2e` 또는 `npm run e2e -- tests/e2e/rbac-routes.spec.ts`
- **문서**: `role-access-concept.md` + `docs/adr/0020*` + `docs/adr/0026*` + `docs/adr/0002*` cross-ref

## 4. 검증 기준 (DoD)

- [x] 10개 검증 대상 모두 결과 명시 (PASS / FAIL / N/A / PARTIAL + 사유)
- [x] 발견 결함 식별 (P1 1건 + P3 3건) — P1 은 follow-up 으로 후속 sprint 인계, 별도 GitHub issue 발행은 본 sprint scope out
- [x] 검증 보고서 작성 — `docs/validation/N-10-manager-rbac.md` (156 lines)
- [x] E2E 시드 (`bob`) 가 DB `team_manager` role 과 정합 (`mgr-user-b` 비공식 명칭 정정)
- [x] PR 머지 후 v1.0 D-11 안에서 N-10 ✅ verified (with P1 follow-up) 마킹

## 5. carry-over (sprint 종료 후)

- 발견 결함 fix → 별도 sprint (Claude 또는 OpenCode Lane 1)
- N-6 staging 운영 (사내 동반)
- v1.0 D-11 안 N-10 ✅ 후 v1.0 릴리즈 DoD 추적
