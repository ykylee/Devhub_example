# M1 PR Review Actions (PR #20–#27)

- 문서 목적: M1 sprint 의 8개 PR 에 대한 자체 리뷰 + Codex bot 자동 리뷰 종합. 머지 시점에 *해소된* 항목과 *후속 개발로 넘긴* 항목을 분리해 다음 세션 진입자가 즉시 인지하도록 한다.
- 범위: PR #20 (M1 PR-A SEC-5) ~ PR #27 (M1 PR-G6 frontend RBAC integration)
- 상태: in_progress (머지 작업과 동시 갱신)
- 최종 수정일: 2026-05-08
- 관련 문서: [M1 sprint backlog](./claude/m1-sprint-plan/backlog/2026-05-08.md), [ADR-0002](../../docs/adr/0002-rbac-policy-edit-api.md), [API contract §12](../../docs/backend_api_contract.md), [PR-12 review-actions 패턴](./PR-12-review-actions.md)

---

## 1. 출처별 리뷰 요약

| PR | 제목 | 자체 리뷰 | Codex bot |
| --- | --- | --- | --- |
| #20 | feat(http): SEC-5 mask 5xx + sprint plan | OK | (미리뷰) |
| #21 | docs(adr): ADR-0002 RBAC policy edit API | OK | (미리뷰) |
| #22 | docs(api): RBAC §12 rewrite + route mapping | OK | (미리뷰) |
| #23 | feat(rbac): domain + rbac_policies migration | OK | P2 — `is_system` ↔ `role_id` CHECK 누락 |
| #24 | feat(rbac): postgres store + users.role FK | fix amend 권장 | P1 — DeleteRBACRole race → 500 누출 |
| #25 | feat(rbac): RBAC handlers | fix amend 권장 | P1 — main.go 미연결, P1 — bulk PUT non-atomic |
| #26 | feat(rbac): permission cache + enforcement | coordination | P1 — bulk failure 시 cache stale |
| #27 | feat(frontend): RBAC PermissionEditor ↔ backend | fix amend 권장 | P2 — `rolesBaseline` shared ref → Save 무반응 |

---

## 2. 즉시 수정 (해당 PR 안에 amend) — `merge-fixed`

### M1-FIX-A — PR #24 DeleteRBACRole FK race

- **위치**: `backend-core/internal/store/postgres_rbac.go:218`
- **결함**: `DeleteRBACRole` 의 COUNT 검사 후 DELETE 실행 사이에 다른 tx 가 동일 role 을 user 에게 할당하면, DELETE 가 FK violation (23503) 으로 실패하고 raw error 가 ErrRoleInUse 로 *매핑되지 않음* → 응답 500 (계약 위반).
- **조치**: DELETE 분기에 `isForeignKeyViolation(err)` 체크 추가, true 면 `ErrRoleInUse` 로 wrap.
- **출처**: Codex bot P1 (PR #24 inline comment)
- **상태**: PR #24 amend
- **위험**: 단일 인스턴스 dev 환경에서는 race 거의 발생 안 함. 운영 진입 전 fix 필요.

### M1-FIX-B — PR #25 main.go RBACStore wiring

- **위치**: `backend-core/main.go` 의 `RouterConfig` 초기화
- **결함**: PR #25 가 `RouterConfig.RBACStore` 의존을 추가했으나 bootstrap 이 이 필드를 설정하지 않아, prod 의 `/api/v1/rbac/policies*` 와 `/api/v1/rbac/subjects/*/roles` 가 *항상* 503 unavailable.
- **조치**: bootstrap 에서 기존 postgres store 를 `RBACStore` 로도 주입.
- **출처**: Codex bot P1 (PR #25 inline comment)
- **상태**: PR #25 amend
- **위험**: 적용 안 하면 RBAC endpoint 자체가 실행 불가.

### M1-FIX-C — PR #25 bulk PUT atomicity (validate-all-then-apply)

- **위치**: `backend-core/internal/httpapi/rbac.go` `updateRBACPolicies` 핸들러
- **결함**: 입력 role 배열을 loop 안에서 *즉시* store 에 쓰며, 두 번째 entry 의 검증 실패 시 첫 entry 는 이미 commit. 부분 적용 + 에러 응답 = 일관성 없는 상태.
- **조치**: 두 단계로 분리 — (1) 모든 entry 의 lookup + 검증 (system-role metadata 변경 거부, audit invariant 등) 을 먼저 수행, 실패 시 *어느 store call 도 하지 않고* 4xx 반환. (2) 모두 통과 시에만 store update 실행.
- **출처**: Codex bot P1 (PR #25 inline comment) + 본 amend 가 PR #26 의 P1 (cache stale) 도 자동 해소.
- **상태**: PR #25 amend
- **부수 효과**: PR #26 의 cache invalidate 호출 위치 변경 불필요 (atomic 보장 후 loop 끝 invalidate 가 안전).

### M1-FIX-D — PR #27 rolesBaseline deep clone

- **위치**: `frontend/app/(dashboard)/organization/page.tsx:27` 의 `rolesBaseline` useState 초기값
- **결함**: `useState<Role[]>(defaultRoles)` 두 번 호출이 같은 객체 참조 → PermissionMatrix 토글이 baseline 까지 mutate → `isRolesDirty` 가 false 유지 → Save 버튼 영구 disable. backend listPolicies 가 실패한 fallback 경로에서 사용자가 매트릭스 편집 불가.
- **조치**: `useState<Role[]>(() => cloneRoles(defaultRoles))` 로 lazy initializer + deep copy.
- **출처**: Codex bot P2 (PR #27 inline comment) — *사용자 가시 결함* 이라 P2 라도 즉시 fix.
- **상태**: PR #27 amend

---

## 3. 다음 개발로 넘김 — `defer`

### M1-DEFER-A — PR #23 `is_system` ↔ `role_id` 일관성 CHECK (P2 방어선)

- **위치**: `backend-core/migrations/000005_create_rbac_policies.up.sql`
- **결함**: 스키마가 `role_id` 형식과 `is_system` 의 *상호 일관성* 을 강제하지 않아, 직접 SQL/도구로 `role_id='custom-foo', is_system=TRUE` 같은 모순 row 를 삽입할 수 있음. Go store 코드 경로는 항상 일관된 값을 쓰지만, 미래 DB 직접 조작에 대한 방어선 부재.
- **권장 조치 (별도 hygiene migration)**:
  ```sql
  ALTER TABLE rbac_policies
      ADD CONSTRAINT rbac_policies_is_system_consistency CHECK (
          (is_system = TRUE  AND role_id IN ('developer', 'manager', 'system_admin'))
          OR
          (is_system = FALSE AND role_id LIKE 'custom-%')
      );
  ```
- **마일스톤**: M1 잔여 또는 M2 hygiene PR
- **위험**: 현 코드 경로에서 발생 0건. 직접 SQL 운영 진입 시 fix 필요.

### M1-DEFER-B — `roleMeetsMin` / `roleRank` / `requireMinRole` deadcode 정리

- **위치**: `backend-core/internal/httpapi/authz.go`
- **상태**: PR #26 가 enforcement 를 `enforceRoutePermission` 으로 일원화 후, 이 헬퍼들은 *사용처 0건* (test 의 `TestRoleMeetsMin` 만 함수 자체를 검증). cleanup 후보.
- **권장 조치**: 별도 cleanup PR 에서 `requireMinRole`/`roleMeetsMin`/`roleRank` 와 그 단위 테스트 제거.
- **마일스톤**: M1 종료 시 또는 M2 hygiene
- **위험**: 0 (사용 안 하는 코드).

### M1-DEFER-C — `writeRBACServerError` → `writeServerError` 통합

- **위치**: `backend-core/internal/httpapi/rbac.go` 의 임시 helper
- **상태**: PR #25 가 PR-A (#20) 의 `writeServerError` 와 충돌 회피로 추가한 stand-in. PR #20 머지 후 main 에는 `writeServerError` 가 존재하므로 통합 가능.
- **권장 조치**: PR #20·#25 모두 머지된 main 위에서 `writeRBACServerError` 호출을 일괄 `writeServerError` 로 sed-replace + 임시 helper 제거.
- **마일스톤**: PR #25 머지 직후 follow-up 1줄 PR
- **위험**: 0 (이름만 변경).

### M1-DEFER-D — DeleteRBACRole row-lock (강한 race 보장)

- **상태**: M1-FIX-A (FK violation → ErrRoleInUse 매핑) 가 단일 인스턴스 race 를 충분히 처리. 다중 인스턴스 또는 *동시 SetSubjectRole + DeleteRBACRole* 빈도가 높은 운영 환경에서는 `SELECT FOR UPDATE` 도입 검토.
- **마일스톤**: M3 운영 진입 시
- **위험**: 단일 인스턴스 운영에서는 발생 거의 0.

### M1-DEFER-E — PermissionCache 다중 인스턴스 일관성

- **위치**: `backend-core/internal/httpapi/permissions.go` `PermissionCache`
- **상태**: ADR-0002 §6 미해결로 명시됨. 단일 프로세스 in-memory cache + 변경 시 reload. 다중 인스턴스 시 한 노드의 RBAC 변경이 다른 노드에 즉시 전파되지 않음.
- **권장 조치**: pub/sub (Redis 또는 PG LISTEN/NOTIFY) 또는 polling 기반 invalidation. 운영 phase 진입 시 결정.
- **마일스톤**: M3 또는 M4
- **위험**: 단일 인스턴스 dev 에서는 0.

### M1-DEFER-F — API contract §12.4 / §12.5 응답 예시 추가

- **위치**: `docs/backend_api_contract.md` §12.4 (POST custom role), §12.5 (DELETE)
- **상태**: 다른 endpoint 응답 예시는 있으나 12.4/12.5 는 spec 만 있고 응답 JSON 예시 없음. 통합자 가독성 저하.
- **권장 조치**: 응답 예시 + audit_log_id meta 표기 추가.
- **마일스톤**: M2 hygiene
- **위험**: 0 (문서).

### M1-DEFER-G — `MemberTable` 의 role display 어휘 회귀 검증

- **위치**: `frontend/components/organization/MemberTable.tsx:107-118` role 드롭다운
- **상태**: PR #27 의 `defaultRoles.id` 가 `'sysadmin'` → `'system_admin'` 변경 후 MemberTable 의 `member.role === 'System Admin'` 비교는 *표시명* 비교라 wire id 변경 영향 없음. 다만 사용자 환경에서 role 변경이 정상 작동하는지 회귀 확인 필요.
- **권장 조치**: 사용자 환경 e2e 검증 시나리오에 포함.
- **마일스톤**: PR #27 머지 후 사용자 환경 검증
- **위험**: 0 (회귀 가능성 낮음, 검증만 권장).

---

## 4. 머지 순서

P1 fix amend 가 필요한 PR 두 건 (#24, #25) 을 amend 하고, P2 사용자 가시 결함 하나 (#27) 도 amend. 나머지는 즉시 머지 가능.

| 순서 | PR | 작업 |
| --- | --- | --- |
| 1 | #20 SEC-5 | 즉시 머지 |
| 2 | #21 ADR-0002 | 즉시 머지 |
| 3 | #22 contract §12 | 즉시 머지 |
| 4 | #23 domain | 즉시 머지 (M1-DEFER-A 후속) |
| 5 | #24 store + FK | M1-FIX-A amend → base main 변경 → 머지 |
| 6 | #25 handlers | M1-FIX-B + M1-FIX-C amend → base main 변경 → 머지 |
| 7 | #26 enforcement | base main 변경 → 머지 (PR-G4 atomic 후 stale 자동 해소) |
| 8 | #27 frontend | M1-FIX-D amend → 머지 |

본 리뷰 actions 문서 자체도 별도 PR 로 머지 (본 PR).

---

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-08 | 초판 작성 — PR #20~#27 종합 리뷰 결과, FIX A~D + DEFER A~G 정리 |
