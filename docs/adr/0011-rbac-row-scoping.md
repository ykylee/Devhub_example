# ADR-0011: Application/Project Owner 위양과 RBAC row-level scoping

- 문서 목적: Application/Repository/Project 도메인의 Owner / Member 가 자신이 소속된 리소스에 대해서만 쓰기 권한을 행사하도록 하는 **row-level RBAC scoping** 정책 결정. 단계적(phased) 도입 방식으로 1차는 handler/service 코드 검증, 2차/3차는 매트릭스 확장 옵션을 보존한다.
- 범위: `applications`, `application_repositories`, `projects`, `project_members` 의 read/write 권한이 (a) 전역 `system_admin`, (b) 후속 활성화될 `pmo_manager`, (c) 리소스의 owner / member 로 어떻게 분기되는지 결정한다. application/project 외 모듈 (auth/account/org/RBAC policy 자체) 은 본 ADR 의 범위 밖이다.
- 대상 독자: Backend 개발자, RBAC 정책 stakeholder, AI agent, 추적성 리뷰어.
- 상태: accepted
- 작성일: 2026-05-14
- 결정일: 2026-05-14 (sprint `claude/work_260514-a`)
- 결정 근거 sprint: `claude/work_260514-a` — Application Design 1차 + ADR-0011 평가/결정 (mixed sprint).
- 관련 문서: [`docs/planning/project_management_concept.md`](../planning/project_management_concept.md) §5.3, §10, [`docs/requirements.md`](../requirements.md) §5.4.1 (REQ-FR-PROJ-000, REQ-FR-PROJ-009, REQ-FR-PROJ-010), [ADR-0002](./0002-rbac-policy-edit-api.md), [ADR-0007](./0007-rbac-cache-multi-instance.md), [추적성 매트릭스 §2.2 RBAC API + §3 Application/Project 행](../traceability/report.md).

## 1. 컨텍스트

`docs/planning/project_management_concept.md` §5.3 는 Owner 권한 위양을 3단계로 정의한다:

- **1차 (MVP)**: 모든 쓰기 작업은 `system_admin` 단일 주체.
- **2차**: `owner` 에게 제한된 메타/멤버 수정 위양.
- **3차**: RBAC row-scoping — 부서/리소스 단위 위양 확장.

3차 단계의 enforcement 가 본 ADR 의 결정 대상이다. 현재 RBAC 모델 (ADR-0002, ADR-0007) 은 `(role, resource, action)` 의 4축 boolean 매트릭스이며, **row-level 조건 (= 이 row 의 `owner_user_id` 가 caller 인가?) 은 없다**. 즉 현재 매트릭스 그대로는 *owner-self 만 PATCH 가능* 같은 표현이 불가능하다.

REQ-FR-PROJ-009 와 REQ-FR-PROJ-010 (pmo_manager 활성 후 권한 범위) 가 본 ADR 의 결과를 의존하고 있어, 결정 전까지는 `system_admin` 일임 정책 (REQ-FR-PROJ-000) 이 강제된다.

## 2. 결정 동인

- **deny-by-default 유지**: 모든 row-level scoping 결정은 explicit allow 만 추가하고 deny 는 그대로.
- **현행 RBAC 매트릭스와의 호환성**: 4축 boolean 의 단일 SoT 를 깨지 않거나, 깬다면 마이그레이션 경로가 명확해야 한다.
- **enforcement 위치**: handler / service / store 중 어느 계층에서 row 조건을 검증할지. SQL `WHERE owner_user_id=?` 형태로 store 에 push down 할지, application code 에서 post-filter 할지.
- **감사 (audit)**: row-level 거부도 deny 이벤트로 audit log 에 남길지. `auth.role_denied` (현행) 와 정합해야 한다.
- **운영 디버깅**: deny 사유가 즉시 보여야 함 — "왜 차단됐는지" 가 audit log 한 줄로 설명 가능해야 한다.
- **정책 변동성**: Application Owner 위양 정책은 MVP→2차→3차 의 3단계로 미리 정의됨. 단계 내 변동성은 낮음.

## 3. 검토 옵션 평가

### 3.1 비교 표

| 옵션 | 요약 | 장점 | 단점 | DevHub 정합 | 마이그레이션 영향 | 운영 디버깅 | 평가 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **A. Casbin / ABAC** | 별도 정책 엔진 도입, row 조건을 정책 표현(REGO/CASBIN-MODEL)으로 기술 | 표현력 ↑, 정책 hot-reload, 외부 정책 변경 가능 | 학습 비용, 또 하나의 SoT (ADR-0002 의 4축 매트릭스와 분리), Permission UI 별도 필요, casbin-go 의존 추가 | **낮음** — ADR-0002 + ADR-0007 PermissionCache 와 정합 깨짐, 둘 중 하나로 통합 필요 | 정책 데이터 신규 + cache invalidation 채널 재설계 | 보통 (casbin 자체 로깅) | ❌ |
| **B. RBAC 매트릭스 확장 (`row_predicate`)** | 매트릭스 row 에 `row_predicate` jsonb 컬럼 추가 (예: `{"owner_user_id": "$caller"}`) + 평가 엔진 | 단일 SoT 유지, 기존 RBAC 정책 UI 의 자연스러운 확장, ADR-0007 cache 와 정합 가능 | 평가 엔진 (DSL) 직접 작성 필요, 표현력 한계 (단순 equality + AND/OR 정도), `$caller` / `$caller.org_unit` 같은 컨텍스트 변수 정의 비용 | **높음** — ADR-0002 의 자연스러운 확장 | rbac_policies 컬럼 추가 + IMPL-rbac-03 enforce 갱신 + PermissionCache invalidation | 좋음 (RBAC 정책 row 와 audit 정합) | 🟡 (3차에서 채택 후보) |
| **C. handler/service 코드 검증** | RBAC 4축 매트릭스는 그대로 두고, handler/service 가 row 조건을 Go 코드로 검증 | 도입 비용 0, 정책 표현 유연 (복잡 조건도 가능), 기존 audit 패턴 그대로 사용 | 정책 분산 (handler 마다 코드), 누락 위험, 정책 변경 = 코드 변경 (관리자 UI 로 못 바꿈) | **높음** — 1차 단계 (system_admin 일임 → owner 위양) 의 단순 조건에 최적 | 0 (코드 변경만) | **매우 좋음** (deny 이벤트가 audit row 한 줄로 명확) | ✅ (1차 채택) |
| **D. PostgreSQL RLS** | PG `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` + `CREATE POLICY` 로 DB layer enforcement, caller 전파는 `SET LOCAL devhub.current_user_id` | DB 가 enforce → 코드 누락 시에도 차단 (defense-in-depth), 효율적 | 정책이 DB 마이그레이션과 결합 (변경 = 마이그레이션 + 재배포), audit log 통합 어려움 (deny 가 SQL 응답 0 row 로 표현됨 — "왜" 가 보이지 않음), connection-level variable 전파 비용, ORM/raw SQL 모두에서 일관 적용 필요 | **낮음** — audit 와 디버깅 정합성 깨짐 | 매 테이블 RLS enable + policy 정의 (마이그레이션 비용 큼) | 나쁨 (deny 가 보이지 않음) | ❌ |

### 3.2 단계적 도입 관점

- **1차 (MVP, 본 sprint 부터 적용)**: row-scoping 코드 경로 자체가 0 — system_admin 일임 (REQ-FR-PROJ-000). 어떤 옵션을 선택해도 1차에서는 dead code.
- **2차 (Owner 위양)**: Application owner 가 자신이 owner 인 리소스의 메타/멤버 수정 가능. 조건이 단순(`actor.UserID == app.OwnerUserID`). **옵션 C 의 handler-level check 1줄** 로 충분.
- **3차 (부서 매핑, RBAC row-scoping)**: `actor.OrgUnitID ∈ ancestors(app.OwnerOrgUnitID)` 같은 조건이 등장. C 의 handler 코드로 표현 가능하나, 정책이 빠르게 늘어나면 분산 위험 ↑. 이 시점에 **옵션 B (row_predicate) 로 마이그레이션** 검토.

## 4. 결정

**옵션 C (handler/service 코드 검증) 를 1차 채택**한다. 후속 단계에서 **옵션 B (row_predicate) 로의 확장 옵션을 보존**한다.

### 4.1 1차 (MVP 진입 시) — 본 sprint
- RBAC 매트릭스는 ADR-0002 의 4축 boolean 그대로 유지. `applications` / `application_repositories` / `projects` / `scm_providers` 4개 신규 resource 를 매트릭스에 추가, **모든 4축은 `system_admin` 만 true**, 나머지 role 은 false.
- handler/service 코드는 row 조건 검증을 하지 않음 — system_admin 일임 정책 (REQ-FR-PROJ-000) 이 enforce.
- `pmo_manager` 활성화 전 요청은 `403 role_not_enabled` (REQ-FR-PROJ-000 의 기존 정책).

### 4.2 2차 (Owner 위양) — 후속 sprint
- `pmo_manager` 활성화 + Application owner 의 메타/멤버 위양 시:
  - handler 가 `actor.HasRole(systemAdmin) || actor.HasRole(pmoManager) || actor.UserID == application.OwnerUserID` 형태로 직접 검증.
  - deny 시 audit log 에 `auth.row_denied` action + `target_type=application` + `target_id` + `denied_reason=owner_mismatch` 기록.
  - REQ-FR-PROJ-009 (`Owner 위양`) 가 이 단계에서 활성화.
- 검증 helper 는 `internal/httpapi/permissions.go` 의 `enforceRowOwnership(actor, ownerUserID, allowedRoles)` 형태로 통일 — 분산을 한 helper 로 모아 누락 위험 완화.

### 4.3 3차 (부서 매핑 / row_predicate 마이그레이션) — 후속
- 정책 표현이 `actor.OrgUnitID ∈ ancestors(app.OwnerOrgUnitID)` 이상으로 복잡해지거나, 정책 변경이 운영 시점에 빈번해지면:
  - **옵션 B (row_predicate)** 채택. `rbac_policies` 에 `row_predicate jsonb` 컬럼 추가, 평가 엔진을 `internal/httpapi/permissions.go` 에 구현.
  - 마이그레이션: 기존 4축 매트릭스 row 의 `row_predicate = NULL` 이면 unconditional, NOT NULL 이면 평가 엔진 호출.
  - 본 ADR 의 4.2 helper 는 평가 엔진의 sugar 로 유지하여 backward compat.
- 이 단계 진입 조건: (a) row-condition 정책이 ≥10 개 또는 (b) 정책 변경이 코드 배포 없이 필요해진 운영 시점. ADR 갱신 (or 신규 ADR-00XX) 으로 결정 재확인.

### 4.4 옵션 A (Casbin) 와 D (PG RLS) 거부 사유

- **A (Casbin)**: ADR-0002 + ADR-0007 의 단일 SoT (RBAC matrix + PermissionCache LISTEN/NOTIFY) 를 깨지 않고는 도입 불가. 거부 비용보다 효익 작음. 향후 외부 정책 hot-reload 요구가 강하게 등장하면 재평가.
- **D (PG RLS)**: deny 가 빈 응답으로 표현되어 audit/디버깅 정합 깨짐. defense-in-depth 가 필요한 *민감 데이터* (예: HRDB persons) 에 한정해 별도 ADR 로 평가 가능하나, Application/Project 도메인은 운영 가시성이 우선.

## 5. 결과

- **sprint `claude/work_260514-a`** 가 1차 채택 (4.1) 의 RBAC 매트릭스 확장 (`applications` / `application_repositories` / `projects` / `scm_providers` 4 resource × 4 axis = 16 cell × N role) 을 수행.
- handler/service 코드는 row 조건 검증을 도입하지 않음 (1차 단계에서는 dead path).
- **sprint `claude/work_260515-c`** 가 §4.2 의 `enforceRowOwnership` helper + `auth.row_denied` audit pattern 을 `backend-core/internal/httpapi/permissions.go` 에 도입. **REQ-FR-PROJ-009 활성화 조건 충족.** handler 단위 호출은 별도 sprint (pmo_manager seed 결정 시점).
- 매트릭스 §2.2 RBAC 인덱스에 신규 4 resource 추가 + §3 Application/Project row 의 IMPL 컬럼에 본 ADR §4.1 reference 추가.

## 6. 후속 작업

- **(1차, 본 sprint)** `applications` / `application_repositories` / `projects` / `scm_providers` resource 의 RBAC matrix seed 마이그레이션 작성. Frontend `PermissionMatrix` 의 `resources` 배열과 `rbac.types.ts` 의 `defaultRoles` 도 9 resource 로 확장 (self-review B1 보강).
- **(2차, 후속 sprint)** `enforceRowOwnership` helper + audit `auth.row_denied` action 도입 + REQ-FR-PROJ-009 활성화.
  - **시그니처 후보**: `func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) (allowed bool)` — `*gin.Context` 에서 actor 추출 + audit emit 책임을 helper 가 가짐 (`auth.row_denied` 자동 기록). `allowedRoles` 가 비어 있으면 `[system_admin]` fallback. caller 는 단순히 `if !h.enforceRowOwnership(c, app.OwnerUserID, "pmo_manager"); return`.
  - **audit payload**: `{actor_role, owner_user_id, resource, action, denied_reason: "owner_mismatch"}`.
- **(3차, 후속)** row_predicate 마이그레이션 진입 조건 모니터링 — 정책 row 수 + 운영 변경 빈도 임계 도달 시 옵션 B 채택 결정.
- **(carve out, API-41~50 stub audit emit)** — 본 sprint 의 handler stub 은 501 응답에 audit 미기록. 후속 sprint 에서 handler body 도입 시 `application.{list,get,create,update,archive}.requested` 같은 audit action 도 함께 발급 (자세한 사항은 sprint state.json `carve_out`).

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-14 | placeholder 초안 (proposed) — concept.md 가 4회 forward reference 하던 ADR-0011 의 anchor. 결정 carve out. | PR #104 보강 commit |
| 2026-05-14 | accepted — 옵션 C (handler/service 코드 검증) 1차 채택, 옵션 B (row_predicate) 단계적 확장 옵션 보존. 옵션 A/D 거부. 1차 매트릭스 seed (4 resource × system_admin) 본 sprint 처리. | sprint `claude/work_260514-a` |
| 2026-05-15 | §4.2 `enforceRowOwnership(c, ownerUserID, allowedRoles...) bool` helper + `auth.row_denied` audit pattern 을 `internal/httpapi/permissions.go` 에 도입. allow 규칙: (1) system_admin, (2) allowedRoles 화이트리스트, (3) actor.login == ownerUserID. deny 시 403 + `code=auth_row_denied` envelope + audit payload `{actor_role, owner_user_id, resource, action, denied_reason="owner_mismatch"}`. 단위테스트 6건. REQ-FR-PROJ-009 활성화. handler 호출은 후속 sprint. | sprint `claude/work_260515-c` |
