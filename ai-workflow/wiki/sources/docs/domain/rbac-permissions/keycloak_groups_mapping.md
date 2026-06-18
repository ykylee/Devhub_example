---
title: keycloak_groups_mapping
type: source
tags: [domain, keycloak_groups_mapping.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/rbac-permissions/keycloak_groups_mapping.md]
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
related: [none]
status: draft
contradictions: [none]
---

# [DEPRECATED] Keycloak group → DevHub RBAC role 자동 매핑 (설계만, 구현 안 함)

> **2026-06-02: ADR-0026 결정으로 Keycloak role 정보를 완전히 무시하기로 함.**
> Keycloak의 realm role / group membership / `devhub_role` attribute 등 모든 role 정보는 DevHub RBAC의 source of truth로 사용하지 않는다.
> Role 변경은 DevHub Admin UI 또는 DevHub API를 통해서만 가능하다.
> 상세: [ADR-0026 Keycloak Role 무시 — DevHub 내부 Role 단독 사용](../adr/0026-keycloak-role-excluded-decision.md)

- 문서 목적: [ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 잔여 carve out 의 Keycloak `groups` claim → DevHub RBAC role 자동 매핑 design. **ADR-0026으로 이 접근법이 채택되지 않음. 본 문서는 설계 기록으로만 보존.**
- 범위: Keycloak group 운영을 통한 DevHub RBAC role 자동 할당 옵션 비교 + 권장 + 운영 SOP. Platform/Project Owner 위양 등 [ADR-0011 RBAC row-scoping](../adr/0011-rbac-row-scoping.md) 영역은 본 design 범위 밖.
- 대상 독자: 아키텍트, 운영자 (SRE / IdP), Backend / IdP 담당자, RBAC 정책 결정자.
- 상태: deprecated (2026-06-02, ADR-0026로 대체)
- 최종 수정일: 2026-06-02
- 결정 근거 sprint: `claude/work_260519-f`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0011 RBAC row-scoping](../adr/0011-rbac-row-scoping.md), [ADR-0002 RBAC policy edit API](../adr/0002-rbac-policy-edit-api.md), [keycloak_operations.md](../setup/keycloak_operations.md) §4, [backend-core/internal/auth/keycloak_verifier.go](../../backend-core/internal/auth/keycloak_verifier.go) (role extraction logic).

## 1. 컨텍스트 + 동기

### 1.1 현재 (옵션 A — realm role 직접 할당)

- [keycloak_operations.md §4](../setup/keycloak_operations.md) 가 정의: realm role 4종 (`developer` / `manager` / `team_manager` / `system_admin`) 직접 할당.
- §8.1 신규 user 생성 SOP: Keycloak admin 이 user 마다 Role Mapping 탭에서 realm role 1개 선택.
- backend `keycloak_verifier.go:260-285` 가 3단계 fallback 으로 role 추출:
  1. `claims["roles"]` (custom, 사용 안 함)
  2. `claims["realm_access"]["roles"]` (Keycloak 표준 realm role) ← **현재 source**
  3. `claims["resource_access"][client_id]["roles"]` (Keycloak client role, fallback)
- `groups` claim 은 현재 **활용 안 함**.

### 1.2 옵션 A 의 운영 부담

- **다중 user 일괄 처리 어려움** — 부서 단위 role 부여 시 1명씩 Role Mapping 진입 필요
- **HR sync 자연스럽지 못함** — HR 시스템의 부서/팀 → Keycloak group 매핑 자연스러우나, role 직접 할당으로는 group 우회
- **Audit trail 측면** — Keycloak admin event 의 `REALM_ROLE_MAPPING:CREATE/DELETE` 가 user 별 발생 (group event 는 더 압축적)
- **확장성** — 신규 role 추가 시 모든 user 의 매핑 재검토 burden

### 1.3 통합 가능 옵션

Keycloak group 기능 (Group / Composite Role / Group Membership Mapper) 을 활용하면 위 부담 해소 가능. 본 design 이 옵션 비교 + 권장 결정.

## 2. 통합 옵션 비교 (4종)

| 옵션 | 변경 범위 | backend 영향 | 운영 SOP 변화 | 권장 |
| --- | --- | --- | --- | --- |
| **A. realm role 직접 할당 (현행 유지)** | 없음 | 변경 없음 | §8.1 step 3 그대로 (user 별 role assign) | △ 부담 ↑ — 다중 user / HR sync 시점 약함 |
| **B. Keycloak group 의 composite realm role** | Realm Groups 4개 + 각 group 에 realm role assign | **변경 없음** — token 의 `realm_access.roles` 에 group 의 composite role 자동 포함 | §8.1 step 3 → user 를 group 에 추가만, role 자동 상속 | ⭐ **권장** — backend impact 0 + 운영 단순 + HR sync 자연 |
| **C. `groups` claim mapper + backend mapping** | Keycloak group 생성 + Group Membership token mapper 활성 + backend `keycloak_verifier.go` 에 groups → role mapping table 추가 | 변경 필요 (`extractRole` 4번째 step 추가) | §8.1 group 추가 + backend mapping table 동기화 burden | △ nested group hierarchy 활용 시 가치 — Phase 2 carve |
| **D. Composite Role (group 없이 realm role 위계만)** | 기존 realm role 4종에 composite assign (예: system_admin = team_manager + manager + developer composite) | 변경 없음 — `realm_access.roles` 에 composite role 자동 포함 | role 위계 표현 가능, group 미사용 | △ role 위계 표현 목적이지 group → role 매핑 목적과 다름. 보조 옵션. |

## 3. 옵션 B (권장) — Keycloak group composite realm role 상세

### 3.1 그룹 구조 (1:1 매핑)

| Keycloak Group | Composite Realm Role | DevHub `users.role` 값 |
| --- | --- | --- |
| `devhub-developers` | `developer` | `developer` |
| `devhub-managers` | `manager` | `manager` |
| `devhub-pmo-managers` | `team_manager` | `team_manager` |
| `devhub-system-admins` | `system_admin` | `system_admin` |

- group ↔ role 1:1 매핑 — 단순화 + 운영 명확성
- group naming 에 `devhub-` prefix — Keycloak 의 사내 다른 시스템 group 과 충돌 회피 (사내 Keycloak 공용 시)
- composite 설정: Group → 해당 realm role 1개를 Composite Role 로 추가 → group 멤버 token 에 자동 포함

### 3.2 Keycloak admin 설정 SOP

1. Realm `devhub` → Groups → Create Group 4회 (위 §3.1 표 의 4 group name)
2. 각 Group → Role Mappings 탭 → realm role 1개 assign (group ↔ role 매핑)
   - `devhub-developers` group → `developer` role
   - `devhub-managers` → `manager`
   - `devhub-pmo-managers` → `team_manager`
   - `devhub-system-admins` → `system_admin`
3. **Default Groups 미설정 권장** (codex review #9 정정) — `devhub-developers` 를 Default Groups 로 추가 시, 신규 manager / team_manager / system_admin user 도 자동 `devhub-developers` 가입 → token `realm_access.roles` 에 multiple role 포함. **sprint -q (PR #185, 2026-05-19)** 에서 backend `extractKeycloakRole` 가 `selectHighestPriorityRole` helper + `devhubRolePriority` map (system_admin 4 > team_manager 3 > manager 2 > developer 1) 으로 multi-role priority filter 구현 — order-dependency 해소. **그러나 Default Groups 미설정 권장은 여전히 유효** (정책 안전성 + 명시 group 가입 SOP 일관). priority filter 는 fallback 정합 — operator 가 잘못 multi-role 부여해도 token 의 highest priority role 이 사용됨.

### 3.3 backend 동작 (sprint -q multi-role priority 구현)

backend `internal/auth/keycloak_verifier.go` 의 role 추출 로직 — sprint -q (PR #185) 의 `selectHighestPriorityRole` filter 적용 후:

```go
if raw, ok := claims["realm_access"]; ok {
    if m, ok := raw.(map[string]any); ok {
        if roles := anyToStrings(m["roles"]); len(roles) > 0 {
            return roles[0]
        }
    }
}
```

- group composite role 은 Keycloak 이 token 발급 시 user 의 group membership 의 모든 composite role 을 `realm_access.roles` 에 자동 포함
- backend 는 `realm_access.roles[0]` 그대로 추출 → 정상 동작
- **변경 코드 0 — Phase 1 에서 backend PR 없음**

### 3.4 user 운영 SOP (§8.1 신규 user 생성 갱신)

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1 | Keycloak admin → Users → Add user | username + email + first/last name |
| 2 | user attribute | Attributes 탭 → `employee_id` = HRDB primary key |
| **3 (갱신)** | **Groups 탭** | **해당 group 에 join (1개 group 만 = role 1개)** |
| 4 | Credentials 탭 | 초기 password + "Temporary" ON |

§8.1 step 3 의 "role 1개 선택" 이 **"group 1개 가입"** 으로 단순화. role mapping 별도 진입 불필요.

### 3.5 HR sync 자연 통합 (carve)

옵션 B 의 부수 장점 — HR 시스템 → Keycloak group sync 가 자연:

- 사내 HR 시스템에 부서 / 직급 정보 → Keycloak group 으로 자동 sync (SCIM bridge 또는 LDAP federation)
- 부서 변경 → group membership 변경 → role 자동 갱신
- 단, SCIM bridge / LDAP federation 자체는 별도 carve out (사내 시스템 영향)

## 4. 옵션 C 의 future 가능성 (Phase 2 carve)

옵션 C (groups claim mapper + backend mapping) 는 다음 case 에서 가치:

- **nested group hierarchy** — 예: `devhub/department-A/team-1` 같은 path 기반 권한
- **다중 role 동시 보유** — 1 user 가 여러 group 에 속하고 각 group 이 다른 role 부여 (현재 backend 는 `roles[0]` 만 사용 — multi-role 미지원)
- **그룹 path 의 RBAC row-scoping 활용** — ADR-0011 의 row-scoping 에 group path 활용 (예: `dev_request.assignee_group`)

Phase 2 진입 시 옵션 C 로 확장 시:
- backend `keycloak_verifier.go` 의 `extractRole` 에 4단계 추가: `claims["groups"]` → mapping table → role
- token mapper SOP — Keycloak client `devhub-frontend` 에 Group Membership mapper 추가 (`Full group path` ON / `Add to access token` ON)
- 본 design 의 1:1 매핑은 보존 + nested path 만 추가 처리

## 5. ADR governance 결정

### 5.1 별도 ADR 발행 여부

옵션 B 는 **backend 변경 없음 + Keycloak 운영 단순화** 만으로 결정. ADR-0019 §5.3 carve resolved 만으로 충분 가능. 별도 ADR-0021 발행은 다음 조건 시:

- Phase 2 옵션 C 로 확장 시 multi-role 정책 (backend `users.role` 단일 string 제약 변경 필요 — ADR 가치 있음)
- 사내 정책으로 group hierarchy 변경 (예: 다중 부서 = 다중 role) 시 ADR 후보

**1차 결정**: ADR-0019 §5.3 (8) carve resolved 만으로 한정. 별도 ADR 발행 안 함. Phase 2 진입 시 재평가.

### 5.2 ADR-0011 row-scoping 와의 관계

- ADR-0011 의 row-scoping (Platform/Project Owner 위양) 은 본 design 범위 밖 — row-level 권한은 그대로
- 본 design = **realm role 부여 메커니즘** 만 결정 (직접 할당 → group 가입)
- enforceRowOwnership helper (ADR-0011 §4.2) 는 그대로 동작 — role 자체는 동일하게 추출됨

## 6. cutover 절차

### 6.1 Phase 1 (본 sprint) — design + Keycloak admin SOP

- ✅ 본 design 문서
- ADR-0019 §5.3 (8) carve resolved 마킹
- keycloak_operations.md §4.3 group section + §8.1 step 3 갱신

### 6.2 Phase 2 — staging 적용 (별도 sprint)

- staging Keycloak realm 에 group 4개 + composite role 설정 (위 §3.2 SOP)
- 적용 직후 자동 검증 — [`scripts/verify-keycloak-groups.sh`](../../scripts/verify-keycloak-groups.sh) 1회 실행 (read-only, idempotent). 4 항목 PASS 시 acceptance 충족, 1건 이상 FAIL 시 §3.2 SOP 단계 재진행. 상세 SOP: [keycloak_operations.md §4.4](../setup/keycloak_operations.md#44-group-setup-검증-자동화-scriptsverify-keycloak-groupssh).
- 1주 사용 후 token decode 검수 (`realm_access.roles` 에 정상 포함 확인)
- DevHub backend 변경 없으므로 PR 불필요

### 6.3 Phase 3 — prod 적용

- 사내 보안팀 동의 후 prod 적용
- 기존 user 의 role mapping → group 가입으로 migration (운영팀 일괄 작업)
- 적용 직후 자동 검증 — [`scripts/verify-keycloak-groups.sh`](../../scripts/verify-keycloak-groups.sh) prod 환경 변수로 1회 실행
- 1주 모니터링

### 6.4 Phase 4 (선택, carve) — 옵션 C 확장

- nested group hierarchy 요구 발생 시
- backend `keycloak_verifier.go` 의 `extractRole` 에 groups claim 4단계 추가
- 별도 ADR (별도 ADR 후보 (ADR-0021 은 Onboarding 으로 2026-05-21 발급됨 — Phase 2 진입 시점에 다음 번호 사용)) 발행

## 7. 보안 점검

### 7.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| user 가 여러 group 에 속한 경우 권한 충돌 | 옵션 B 의 1:1 매핑 + §3.1 표의 "1개 group 만" SOP. composite role 누적 시 backend 는 `realm_access.roles[0]` 만 추출 → 가장 먼저 매칭되는 role 이 정해짐 — Keycloak token 의 role 순서 정의 (carve — Keycloak 의 role include order 검수). 또는 backend 확장 시 가장 높은 권한 우선 처리 (Phase 2 carve). |
| group 비활성화 / 삭제 시 user role 회수 | Keycloak admin 의 group 삭제 → 모든 member 의 composite role 자동 회수. user 의 새 token 부터 즉시 반영. 기존 access_token 은 TTL 만료까지 유효 (운영 SOP — short access TTL 권장). |
| `devhub-system-admins` group 의 명시 보호 | Keycloak admin role mapping 권한을 별도 group (예: `keycloak-admins`) 에만 부여. devhub 운영자가 system_admin 권한 임의 부여 차단. |
| HR sync 의 잘못된 group assign | Phase 1 에서는 수동 group 관리. SCIM / LDAP federation 도입 시 별도 carve (사내 시스템 영향). |

### 7.2 audit_logs 영향

- Keycloak admin event `GROUP_MEMBERSHIP:CREATE/DELETE` 가 group 가입/탈퇴 audit
- [keycloak_event_audit_integration.md §4.2](./keycloak_event_audit_integration.md#42-admin-event) 의 admin event 매핑 표에 신규 row 추가 — `GROUP_MEMBERSHIP:CREATE` → `keycloak.user.group.{joined,left}` (carve — design 단계 확정 후 매핑 표 갱신)

## 8. 잔여 carve out / open question

- **(carve)** SCIM bridge / LDAP federation 자동 group sync — HR 시스템 → Keycloak group 자동 동기화. Phase 2 carve.
- **(carve)** 옵션 C 의 multi-role 정책 — backend `users.role` single string 변경 + RBAC policy 재설계. 별도 ADR 후보 (ADR-0021 은 Onboarding 으로 2026-05-21 발급됨 — Phase 2 진입 시점에 다음 번호 사용).
- **(carve)** Keycloak token 의 role include order — multi group composite 시 `realm_access.roles[0]` 가 어느 role 인지 검수 SOP.
- **(carve)** keycloak_event_audit_integration.md §4.2 admin event 매핑 표에 `GROUP_MEMBERSHIP:CREATE/DELETE` row 추가 — design + audit 통합 sprint 정합.
- **(closed by codex review #9)** Default Group 정책 — **미설정 권장** (§3.2 step 3). 사유: backend `extractKeycloakRole` 의 `realm_access.roles[0]` 사용으로 인한 multi-role order-dependency 위험.
- **(open)** Keycloak `User Profile Provider` 의 user attribute 필수 group 매핑 정책 (Phase 2 자동화 carve).

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — 9 section + 옵션 4종 비교 (A realm role 직접 / B group composite / C groups claim mapper / D composite role only) + 권장 B (backend 변경 없음 + 운영 단순) + group 1:1 매핑 표 + Keycloak admin SOP + user 운영 SOP 갱신 + HR sync 자연 통합 + Phase 1..4 cutover + ADR governance 결정 (1차 별도 ADR 발행 안 함) + 보안 점검 + carve 6 항목. | `claude/work_260519-f` |
