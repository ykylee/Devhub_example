# ADR-0013: Dev Request 도메인의 RBAC row-scoping 정책

- 문서 목적: `dev_requests` resource 의 RBAC row-level scoping 정책을 명문화한다. ADR-0011 §4.2 `enforceRowOwnership` 패턴의 dev_requests resource 적용 사례이며, 사후 명문화 (handler wire-up 은 이미 sprint `claude/work_260515-i` 의 PR #124 에 도입됨) 다.
- 범위: `dev_requests` resource 의 read / write 권한이 (a) 전역 `system_admin`, (b) `pmo_manager`, (c) 행 단위 `assignee owner-self` 로 어떻게 분기되는지 결정한다. `dev_request_intake_tokens` resource (외부 수신 인증 관리) 는 별도 ADR-0014 (DREQ-Admin-UI sprint) 에서 다룬다.
- 대상 독자: Backend 개발자, RBAC 정책 stakeholder, AI agent, 추적성 리뷰어.
- 상태: accepted
- 작성일: 2026-05-15
- 결정일: 2026-05-15 (sprint `claude/work_260515-m`)
- 결정 근거 sprint: `claude/work_260515-m` — DREQ-RBAC-ADR + DREQ-Promote-Tx 1번째 carve out.
- 관련 문서: [ADR-0011 Application/Project Owner 위양](./0011-rbac-row-scoping.md) §4.2, [ADR-0012 DREQ 외부 수신 인증](./0012-dreq-external-intake-auth.md), [`docs/planning/development_request_concept.md`](../planning/development_request_concept.md) §6, [`docs/requirements.md`](../requirements.md) §5.5 (REQ-FR-DREQ-002, REQ-FR-DREQ-007, REQ-FR-DREQ-008), [추적성 매트릭스 §3 Dev Request 행](../traceability/report.md).

## 1. 컨텍스트

DREQ 도메인 1차 (sprint `claude/work_260515-f` 의 컨셉 ~ sprint `claude/work_260515-j` 의 frontend 1차) 가 완료되며 다음의 사실들이 굳어졌다:

- 외부 시스템이 보낸 의뢰 (intake) 는 ADR-0012 의 token 인증으로 처리. **이 path 는 RBAC 매트릭스를 거치지 않는다** (전용 middleware `requireIntakeToken`).
- 의뢰가 저장된 이후의 모든 후속 행위 (view / register / reject / reassign / close) 는 일반 actor 의 RBAC 매트릭스 + dev_requests resource 에 대한 권한 + row-level scope 의 3축으로 통제된다.
- handler `getDevRequest` / `registerDevRequest` / `rejectDevRequest` 3 endpoint 는 이미 sprint `claude/work_260515-i` (PR #124) 에서 `enforceRowOwnership(c, current.AssigneeUserID, string(domain.AppRolePMOManager))` 호출로 wire-up 완료. 본 ADR 는 이 wire-up 의 **정책 근거** 를 사후 명문화한다.
- `reassign` (PATCH) 와 `close` (DELETE) 는 system_admin only — handler 가 직접 role 검증 (route gate 와 분리, 자체 forbid). 이 carve 는 REQ-FR-DREQ-008 (`system_admin 만 close 가능`) 에 정합.

ADR-0011 §4.2 가 이미 일반 helper `enforceRowOwnership(c, ownerUserID, allowedRoles ...)` 와 `auth.row_denied` audit pattern 을 정의했으나, **`ownerUserID` 의 의미는 resource 별로 다르다** — applications/projects 는 `OwnerUserID` 컬럼, dev_requests 는 `AssigneeUserID` 컬럼. 이 resource-specific 의미 + 허용 role 화이트리스트를 본 ADR 가 명시한다.

## 2. 결정 동인

- **resource 별 owner 컬럼이 다르다**: applications/projects 는 `owner_user_id`, dev_requests 는 `assignee_user_id`. ADR-0011 helper 의 `ownerUserID` parameter 는 의미적으로 "row 의 책임자 = 본인 의뢰만 보일 수 있는 사람" 으로 일반화되었으므로 dev_requests 에서는 `AssigneeUserID` 를 넣는다.
- **pmo_manager 위양**: ADR-0011 §4.2 가 pmo_manager 의 row-level 위양을 일반 패턴으로 정의했다. dev_requests resource 도 동일 패턴 (행 단위 위양) 으로 일관시킨다 — 다른 resource 와 운영 모델을 통일.
- **assignee 본인 의뢰 view**: 개발자 / 매니저 dashboard 위젯 (`MyPendingDevRequestsWidget`) 이 assignee 본인 의뢰만 보이는 동작은 frontend UX 가 아닌 backend RBAC 가 강제해야 한다 (frontend bypass 방지).
- **외부 intake 와 RBAC 무관**: `POST /api/v1/dev-requests` (API-59) 는 RBAC 매트릭스 외 (intake token 인증). 이 carve 는 ADR-0012 가 이미 결정.
- **close/reassign 의 system_admin only**: 두 동작은 운영 결정 (의뢰 종료 / 담당자 변경) 으로 row 단위 위양 없이 단일 주체가 책임. handler 자체 검증 + REQ-FR-DREQ-008 정합.

## 3. 검토 옵션

본 ADR 는 ADR-0011 의 일반 결정 (옵션 C handler-level + 옵션 B 보존) 을 dev_requests resource 에 적용하는 사례이므로 옵션 비교는 ADR-0011 §3 을 따른다. 본 ADR 는 다음만 결정한다:

| 결정 항목 | 옵션 | 채택 |
| --- | --- | --- |
| owner 컬럼 | `assignee_user_id` vs `requester` | `assignee_user_id` (의뢰 처리 책임자) |
| 위양 role | `pmo_manager` only vs `pmo_manager + developer/manager` | `pmo_manager` only (의뢰 관리 운영 권한, 개별 개발자는 owner-self path) |
| reassign 권한 | row-level (현 assignee 의 자기 이양) vs system_admin only | system_admin only (운영 결정으로 단일화) |
| close 권한 | row-level vs system_admin only | system_admin only (REQ-FR-DREQ-008) |
| register/reject 권한 | row-level (owner-self + pmo_manager) | row-level (위 helper 패턴) |
| view 권한 | row-level (owner-self + pmo_manager) | row-level (위 helper 패턴) |

## 4. 결정

**dev_requests resource 의 RBAC 정책은 다음과 같다.**

### 4.1 RBAC 매트릭스 4축 (resource × action)

`dev_requests` resource 의 4 action 별 role 허용 (마이그레이션 `000024_rbac_dev_requests.sql` 기반):

| role | view | create | edit | delete |
| --- | --- | --- | --- | --- |
| `system_admin` | ✅ | ✅ | ✅ | ✅ |
| `pmo_manager` | ✅ | ❌ | ✅ | ❌ |
| `manager` | ✅ (route gate) | ❌ | ✅ (route gate) | ❌ |
| `developer` | ✅ (route gate) | ❌ | ✅ (route gate) | ❌ |

- `create`: intake path 는 token 인증으로 처리되므로 RBAC `create` 권한은 실질 미사용 (system_admin 단일).
- `delete`: DELETE `/api/v1/dev-requests/:id` (close). system_admin only.
- `view` / `edit`: 4 role 모두 route gate 는 통과하지만, **handler 가 row-scope 검증** 으로 본인 의뢰 외 접근을 차단.

### 4.2 row-level scope (`enforceRowOwnership` 적용)

ADR-0011 §4.2 helper 시그니처 `func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool` 를 그대로 사용한다.

dev_requests resource 의 caller 패턴 (이미 wire-up 완료):

```go
// getDevRequest, rejectDevRequest, registerDevRequest 가 동일 호출.
if !h.enforceRowOwnership(c, dr.AssigneeUserID, string(domain.AppRolePMOManager)) {
    return
}
```

- 첫 번째 인자 `ownerUserID = dr.AssigneeUserID` — dev_request 의 행 단위 책임자.
- 두 번째 가변 인자 `allowedRoles = [pmo_manager]` — system_admin 은 helper 가 자동 통과시키므로 명시 불요. developer/manager 는 owner-self path (=actor.login == AssigneeUserID) 로 통과.
- helper 의 allow 규칙:
  1. `actor.role == system_admin` → allow (audit 없음)
  2. `actor.role ∈ allowedRoles` (=pmo_manager) → allow (audit 없음)
  3. `actor.login == ownerUserID` (=AssigneeUserID, owner-self) → allow (audit 없음)
  4. 그 외 → deny → 403 `code=auth_row_denied` + audit `auth.row_denied` event with payload `{actor_role, owner_user_id, resource="dev_requests", action, denied_reason="owner_mismatch"}`

### 4.3 list endpoint 의 row-level filter

`listDevRequests` (API-60) 는 `enforceRowOwnership` 을 호출하지 않고 (단일 row 가 없으므로) 대신 **store query 의 `assignee_user_id` filter** 를 강제한다:

```go
if !devFallbackEnabled(c) && role != system_admin && role != pmo_manager {
    opts.AssigneeUserID = actor.login  // 본인 의뢰만 보임
}
```

이 흐름은 `MyPendingDevRequestsWidget` 등 frontend 의 본인 의뢰만 표시 정책의 backend enforcement 다.

### 4.4 reassign / close 의 system_admin only

PATCH `/api/v1/dev-requests/:id` (reassign) 와 DELETE `/api/v1/dev-requests/:id` (close) 는 row-level scope 없이 handler 가 직접 role 검증:

```go
if !devFallbackEnabled(c) && actorRole != string(domain.AppRoleSystemAdmin) {
    c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
        "status": "forbidden",
        "error":  "{reassign|close} requires system_admin",
        "code":   "auth_row_denied",
    })
    return
}
```

이 carve 는 REQ-FR-DREQ-008 (`system_admin 만 close 가능`) + 운영 결정 (담당자 변경은 단일 주체) 의 결합.

### 4.5 audit 정합

deny 시 audit event:
- action: `auth.row_denied`
- target_type: `dev_request`
- target_id: `dr.ID`
- payload: `{actor_role, owner_user_id (=assignee_user_id), resource: "dev_requests", action, denied_reason: "owner_mismatch"}`

allow 시 audit event 는 별도 (e.g., `dev_request.registered`, `dev_request.rejected`, `dev_request.reassigned`) 로 handler 가 emit.

## 5. 결과

- 본 ADR 는 sprint `claude/work_260515-i` (PR #124) 에서 도입된 row-scope wire-up 의 **사후 정책 근거** 다.
- handler 수정 0 — 기존 wire-up 이 본 ADR 정책에 이미 정합.
- 매트릭스 §3 Dev Request 행에 본 ADR §4 reference 가 추가 (sprint `claude/work_260515-m`).
- REQ-FR-DREQ-002 (`audit 보존`) + REQ-FR-DREQ-007 (`assignee 본인 의뢰만 보임`) + REQ-FR-DREQ-008 (`system_admin 만 close`) 의 RBAC 측면 정합 완성.

## 6. 후속 작업

- **(carve out, ADR-0014)** `dev_request_intake_tokens` resource 의 admin endpoint (발급 / revoke) 도입 시 별도 ADR 로 정책 결정 — sprint `DREQ-Admin-UI` 진입 시 작성.
- **(carve out, TC-DREQ-RBAC)** 본 ADR 정책의 e2e 테스트 (assignee owner-self / pmo_manager / system_admin / forbidden 4 케이스) — sprint `DREQ-E2E` 진입 시 발급.
- **(monitoring)** ADR-0011 §4.3 의 row_predicate 마이그레이션 진입 조건 (정책 row ≥10 또는 운영 변경 빈도 임계) 에 본 ADR 의 dev_requests 정책도 포함.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-15 | accepted — dev_requests resource 의 RBAC row-scoping 정책 명문화. ADR-0011 §4.2 helper 의 dev_requests resource 적용 사례 사후 명문화. handler wire-up 은 PR #124 (sprint `claude/work_260515-i`) 에서 이미 도입. reassign/close 의 system_admin only carve 도 본 ADR 가 명시. | sprint `claude/work_260515-m` (PR TBD) |
