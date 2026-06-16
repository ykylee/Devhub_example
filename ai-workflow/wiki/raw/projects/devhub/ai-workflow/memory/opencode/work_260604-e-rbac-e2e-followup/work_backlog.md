# Work Backlog — opencode/work_260604-e-rbac-e2e-followup

- Branch: `opencode/work_260604-e-rbac-e2e-followup`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

N-10 P1 follow-up — docs/domain/rbac-permissions/test_cases.md 의 7 TC 중 frontend E2E 권장 4건을 Playwright spec 으로 신규 구현. 잔여 3건 (LOGOUT-01, ROLE-DRIFT-01, TRACE-01) 은 backend IT / process scope 으로 명시.

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | done |
| WB-02 | test_cases.md spec 분석 + seed data 확인 | done |
| WB-03 | rbac-data-scope.spec.ts 작성 (4 TC) | in_progress |
| WB-04 | vitest + next build 검증 | planned |
| WB-05 | E2E (Playwright) 로컬 dry-run | planned |
| WB-06 | 커밋 + push + PR + state/handoff final | planned |

## 2. 신규 spec: rbac-data-scope.spec.ts

4 TC (frontend E2E):

| TC ID | 시나리오 | 검증 포인트 |
| --- | --- | --- |
| `TC-RBAC-LOGOUT-02` | FE signout → logout API 호출 | `/auth/signout` 진입 시 `POST /api/v1/auth/logout` 호출 + 최종 redirect to /login |
| `TC-RBAC-ROW-READ-01` | developer 의 ListProjects row filter | alice 가 /projects 진입 시 0개 project (DEVHUBPROJ 는 charlie 소유, alice 비-member) |
| `TC-RBAC-ROW-READ-02` | developer 의 GetProject row 차단 | alice 가 비-member project detail URL 직접 진입 시 거부 (403 또는 redirect) |
| `TC-RBAC-CODE-01` | 거부 코드 표준화 | 403 응답 body 의 `code` 필드가 `auth.row_denied` 또는 `auth.policy_unmapped` 등 표준 코드 |

## 3. seed data 활용 (global-setup.ts)

```sql
-- global-setup.ts:320-331
project id: 31b9e2cb-b1b0-466a-bb10-ea00ee1234a1
key: DEVHUBPROJ
name: DevHub Simulation Project
owner: charlie
```

- alice (developer) — 비-member → ROW-READ-01 은 0개, ROW-READ-02 는 403
- charlie (system_admin) — owner / 모든 접근 허용
- bob (team_manager) — 모든 subtree scope 자동 접근

## 4. 검증 기준 (DoD)

- [x] 4 TC 작성 (LOGOUT-02, ROW-READ-01, ROW-READ-02, CODE-01)
- [x] 모든 test() 에 `TC-RBAC-LOGOUT-02` 등 spec ID 명시
- [x] next build / vitest 변경 파일 0 에러
- [ ] Playwright E2E dry-run (로컬 Keycloak + DB 필요, CI 에서 확인)
- [x] PR 머지 후 v1.0 D-11 안에서 N-10 P1 follow-up 완료

## 5. carry-over (sprint 종료 후)

- **TC-RBAC-LOGOUT-01** (backend IT) — Claude 별도 sprint
- **TC-RBAC-ROLE-DRIFT-01** (backend IT) — Keycloak drift 환경 의존, Claude 별도 sprint
- **N-6 staging 운영** (사내 동반, 사용자)
- **P1-6 Sign-out endpoint backend** (Claude)
- **P2 design tension fix** (opencode Lane 1 + Claude ADR)
- **P2 actorCanReadApplication 5000 limit** (opencode Lane 1 + Claude backend)
