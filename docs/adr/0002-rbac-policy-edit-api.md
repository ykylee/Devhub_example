# ADR-0002: RBAC Policy 편집 API — 도입 vs static-default 유지 결정

- 문서 목적: M1 sprint DoD #6 — RBAC policy 편집 API 도입 여부와 그 근거를 결정한다.
- 범위: `defaultRBACPolicy()` 의 노출/편집 API, 프론트엔드 `PermissionEditor` 의 동작 모드, runtime 권한 enforcement 와의 관계.
- 대상 독자: Backend·Frontend 개발자, 프로젝트 리드, 후속 RBAC phase 의사결정자.
- 상태: **accepted — Option A (DB-backed RBAC matrix + write API + enforcement 재설계)**
- 결정일: 2026-05-08
- 관련 문서: [통합 로드맵 §3.2 M1 DoD #6](../development_roadmap.md#32-m1--핵심-기능-contract-정합성), [API 계약 §12](../backend_api_contract.md), [ADR-0001](./0001-idp-selection.md), [`backend-core/internal/httpapi/rbac.go`](../../backend-core/internal/httpapi/rbac.go), [`backend-core/internal/httpapi/authz.go`](../../backend-core/internal/httpapi/authz.go), [`frontend/components/organization/PermissionEditor.tsx`](../../frontend/components/organization/PermissionEditor.tsx), [`frontend/lib/services/rbac.types.ts`](../../frontend/lib/services/rbac.types.ts)

---

## 1. 컨텍스트

M0 sprint 가 인증/권한 게이트를 정상화하면서 `requireMinRole` 미들웨어 + 5 라우트 가드가 도입됐다(M0 PR-B). 동시에 다음 두 가지가 존재한다.

1. **Backend** `httpapi/rbac.go` — `GET /api/v1/rbac/policy` 가 `defaultRBACPolicy()` 의 read-only static 응답을 반환. 응답 meta 에 `editable: false`, `source: "static_default_policy"` 명시.
2. **Frontend** `components/organization/PermissionEditor.tsx` + `lib/services/rbac.types.ts` — UI 컴포넌트는 존재하나 *로컬 state* (`defaultRoles` const) 를 직접 편집하는 형태. 백엔드 통신 service (`rbac.service.ts`) 는 작성돼 있지만 호출 path 가 backend 미구현 (`/api/v1/rbac/policies` 복수형, 404).

통합 로드맵 §3.2 M1 DoD #6 은 이 둘 사이의 결정을 요구한다 — *write API 를 도입할 것인지, static 으로 유지할 것인지*.

### 1.1 결정의 핵심 — 모델 분기

코드를 들여다보면 backend 와 frontend/contract §12 의 RBAC 모델이 *완전히 다르다*:

| 측면 | Backend `defaultRBACPolicy()` | Frontend `defaultRoles` + API §12 spec |
| --- | --- | --- |
| Roles | `developer`, `manager`, `system_admin` (snake_case wire) | `developer`, `manager`, `sysadmin` (camelCase id, `sysadmin` 다름) |
| Resources | `repositories`, `ci_runs`, `risks`, `commands`, `organization`, `system_config` (6) | `infrastructure`, `pipelines`, `organization`, `security`, `audit` (5) — 이름·의미 모두 다름 |
| Permission shape | 단일 레벨 (`none|read|write|admin`, rank 0/10/20/30) | per-resource 4-boolean (`{view, create, edit, delete}`) |
| 의도 | role-resource 매트릭스 1차원 | resource 별 CRUD 4축 |
| 현재 호출자 | `getRBACPolicy` (display 전용) | `PermissionEditor` (로컬 state, 미연결) |

`backend_api_contract.md §12` 는 frontend 모델을 그대로 spec 화한 *aspirational* 문서이며 백엔드는 이를 구현하지 않았다. 두 모델이 서로 다른 영역을 표현하고 있어, 단순 "API 추가" 가 아니라 *모델 통일 결정* 이 선행돼야 한다.

### 1.2 결정의 핵심 — Runtime enforcement 와의 단절

가장 중요한 사실은 **`requireMinRole` 이 RBAC 매트릭스를 런타임에 참조하지 않는다** 는 것이다 (`authz.go`):

```go
func roleRank(role string) int {
    switch role {
    case string(domain.AppRoleSystemAdmin): return 30
    case string(domain.AppRoleManager):     return 20
    case string(domain.AppRoleDeveloper):   return 10
    default:                                return 0
    }
}
```

라우트별 가드는 `router.go` 에서 `h.requireMinRole(domain.AppRoleManager)` 식으로 *정적 코드* 로 부착되며, `defaultRBACPolicy()` 의 매트릭스나 DB 의 어떤 데이터도 enforcement 결정에 들어오지 않는다.

따라서 매트릭스 편집 API 를 추가해 DB 에 저장하더라도, *실제 권한 체크는 코드 변경 + 재배포가 있어야만 변한다*. 매트릭스 write API 는 사용자에게 "권한이 바뀐다" 는 잘못된 기대를 줄 위험이 있다 (security theatre).

---

## 2. 결정 동인 (Decision drivers)

1. **사용자 기대 일치**: UI 가 "Save" 한 권한 변경이 실제 enforcement 와 *일치* 하거나, 일치하지 않는다는 것이 *명확* 해야 한다.
2. **개발 비용**: M1 의 다른 DoD (#1·#2·#3·#7·#8) 가 더 큰 사용자 가치 — write API 에 sprint capacity 를 할애할지 여부.
3. **모델 통일 비용**: backend(1차원) ↔ frontend(2차원 boolean) 의 모델 차이를 좁히는 데 필요한 변환 매핑/마이그레이션 비용.
4. **M2 Phase 6/6.1 흐름**: M2 의 "PermissionEditor 복구 + RBAC Guard 실체화" 와의 의존 관계 — 본 ADR 결정이 M2 작업 범위를 좌우.
5. **운영 audit 요구**: 권한 변경이 audit log 에 남아야 하는지, 남아도 enforcement 가 따라가지 않으면 의미가 있는지.
6. **보안 theatre 회피**: 편집 가능한 UI 가 enforcement 와 단절돼 있으면 "내가 막은 줄 알았다" 식의 보안 사고 risk.

---

## 3. 검토한 옵션

### 3.1 Option A — Persist & enforce (full write path)

DB-backed RBAC matrix + `requireMinRole` 가 매트릭스를 runtime 에 참조 + admin write API + audit + cache + invalidation.

#### 변경 범위

- `backend-core/internal/store/postgres_rbac.go` — `rbac_policies` 테이블 + 마이그레이션.
- `backend-core/internal/httpapi/rbac.go` — `PUT /api/v1/rbac/policies` 추가, audit (`rbac.policy.updated`).
- `backend-core/internal/httpapi/authz.go` — `requireMinRole` → `requirePermission(resource, action)` 로 *재설계*. 라우트별 (resource, action) 매핑 테이블 추가.
- 모델 통일 — backend 의 `none|read|write|admin` 1차원과 frontend 의 `{view, create, edit, delete}` 4축 중 하나로 결정 + 변환.
- 캐시 — DB 적중 비용을 피하기 위해 in-memory matrix cache + invalidation.
- `frontend/lib/services/rbac.service.ts` — 이미 작성된 호출 path 가 동작.

#### 장점

- 사용자 기대와 enforcement 가 일치 (편집 = 즉시 enforcement).
- API §12 spec 이 실현됨.
- M2 Phase 6.1 의 RBAC Guard 실체화가 자연스럽게 흡수됨.

#### 단점

- **규모 큼** — M1 sprint capacity 를 상당히 소모 (~600~900 LoC backend + 마이그레이션 + frontend 통합 + 테스트).
- 모델 통일 결정이 ADR 두 번째로 필요 (1차원 vs 4축) — 그 자체가 별도 의사결정 부담.
- 라우트별 (resource, action) 매핑 누락 시 *기본 거부* vs *기본 허용* 판단이 또 다른 부담.
- 캐시 무효화 + RBAC 정책 변경의 가시성 (다른 인스턴스로 전파) 가 운영 문제로 등장.

### 3.2 Option B — Static-default 유지 + 모델 정렬 (read-only viewer) — **권장**

`defaultRBACPolicy()` 를 코드의 source-of-truth 로 확정. PermissionEditor 를 *read-only viewer* 로 재설계. frontend 모델을 backend 모델로 정렬. 편집은 코드 PR 로만.

#### 변경 범위

- `backend-core/internal/httpapi/rbac.go` — 변경 없음. `editable: false` 그대로.
- `backend_api_contract.md §12` — *현재 spec 을 invariant 박스로 라벨링* 하고, "현 단계에서는 GET /api/v1/rbac/policy 만 제공. PUT/subjects 는 후속 phase 의 결정에 따라 도입" 으로 수정. 12.2~12.4 제거 또는 *후보 backlog* 로 강등.
- `frontend/lib/services/rbac.types.ts` — `defaultRoles` 를 backend `defaultRBACPolicy()` 와 동형으로 재구성. 4-boolean → 1차원 (`none|read|write|admin`) 으로 합치거나, view-only 라면 backend 응답을 직접 표시.
- `frontend/lib/services/rbac.service.ts` — `getPolicy()` 만 유지, `getPolicies/updatePolicies/getSubjectRoles/updateSubjectRoles` 제거 (현 단계에서 backend 미구현이므로 미사용 코드 정리).
- `frontend/components/organization/PermissionEditor.tsx` — Create/Delete/Edit 버튼 비활성화 또는 제거. read-only 표시 모드. UI 라벨에 "Defined in code — see ADR-0002" 노트.
- `frontend/app/(dashboard)/organization/page.tsx` — `roles, setRoles` → `roles` (read-only) 로 단순화.
- M2 Phase 6.1 (RBAC Guard 실체화) 의 범위 갱신 — *enforcement* 작업으로 한정. policy CRUD 작업은 본 ADR 의 후속 재검토 트리거 (§5) 발생 시까지 보류.

#### 장점

- 사용자 기대 = 현실 (UI 가 enforcement 와 일치 — "변경하려면 엔지니어링 PR").
- M1 sprint capacity 보존 — 다른 DoD 에 집중.
- 보안 theatre 회피.
- 모델 정렬은 frontend → backend 단방향이라 단순 (1 PR 분량).
- 운영 부담 0 (DB 마이그레이션·캐시 없음).

#### 단점

- "관리자가 UI 에서 RBAC 을 편집하는" 미래 요구가 들어오면 본 ADR 을 다시 열어 Option A 로 전환해야 함.
- API §12 spec 의 일부가 "후속 phase" 로 강등돼 spec 의 권위가 약해짐 — *충돌 해소 표* 로 명시 필요.
- frontend `PermissionEditor` 의 "Create Custom Role" 같은 기능이 사라짐 (실제로는 enforcement 미연결이라 작동하지 않던 기능).

### 3.3 Option C — Display-only persistence (compromise)

DB-backed policy 표시 매트릭스를 두되, `requireMinRole` 은 계속 코드 enforce. 즉 *편집 가능한 표시 매트릭스* 는 DB, *실 권한* 은 코드.

#### 단점

- 사용자가 UI 편집을 enforcement 변경으로 오해할 위험 가장 큼 (Option A 의 단점 대부분 + Option B 의 장점 대부분 상실).
- "왜 admin 이 권한 바꿨는데 안 막혀요?" 식의 보안 사고/지원 부담.
- **명시적으로 비추천**.

### 3.4 옵션 요약 비교

| 측면 | A. Persist & enforce | B. Static + 정렬 (권장) | C. Display-only persistence |
| --- | --- | --- | --- |
| Sprint 부담 (LoC) | ~600~900 + 마이그레이션 | ~150 frontend 정리 + 문서 | ~300 backend + 마이그레이션 |
| 사용자 기대 일치 | 일치 | 일치 (read-only 명시) | **불일치 (theatre)** |
| Enforcement 작동 | DB 매트릭스 | 코드 라우트 가드 (현행) | 코드 라우트 가드 (현행) |
| API §12 spec | 실현 | 일부 후속 phase 로 강등 | 실현되나 enforce 안 함 |
| 모델 통일 결정 | 별도 ADR 필요 | frontend→backend 정렬 1 PR | 별도 ADR 필요 |
| 후속 phase 영향 | M2 Phase 6.1 흡수 | M2 Phase 6.1 = enforcement 만 | M2 Phase 6.1 흡수 |
| 운영 부담 | 캐시·전파·마이그레이션 | 0 | 캐시·전파·마이그레이션 |

---

## 4. 결정 (Accepted)

### 4.1 채택 옵션 — **Option A (Persist & enforce — DB-backed RBAC matrix + write API + enforcement 재설계)**

> 본 결정은 사용자 결정 (2026-05-08, ADR-0002 채택 단계) 으로 확정.

#### 근거 요약

1. **enforcement-편집 일치를 우선** — UI 가 "Save" 한 권한 변경이 실제로 적용돼야 한다는 product 요구가 우선시됨. 매트릭스 편집 ↔ enforcement 의 단절(Option C 의 theatre, Option B 의 read-only) 을 허용하지 않는다.
2. **API §12 spec 의 실현** — `backend_api_contract.md §12` 의 spec 이 그대로 살아남고, frontend `PermissionEditor` + `rbacService` 가 *작동* 하는 형태로 통합된다.
3. **M2 Phase 6.1 흡수** — "정책 CRUD 연동" + "RBAC Guard 실체화" 가 본 옵션의 자연스러운 산출물 — M2 Phase 6.1 의 범위가 본 ADR 의 후속 PR 들로 흡수됨.
4. **모델 통일 결정 동반** — Option A 채택은 backend(1차원)/frontend(4축) 모델 통일 결정을 함께 요구한다. §4.2 에서 이 결정도 함께 확정한다.

### 4.2 동반 결정 — RBAC 모델 통일

Option A 의 구현은 모델 통일 결정 없이 진행 불가능하므로 본 ADR 에서 함께 확정한다.

#### 채택 모델 — **per-resource CRUD 4축** (frontend / API §12 spec 모델)

| 구성 | 채택 표현 |
| --- | --- |
| Roles | 시스템 정의 3종 (`developer`, `manager`, `system_admin`) + 사용자 정의 가능 (Option A 의 자연스러운 결과). wire 형식은 backend 의 snake_case 유지. |
| Resources | API §12 의 frontend 어휘 그대로 (`infrastructure`, `pipelines`, `organization`, `security`, `audit`) — *단 backend 의 라우트 클래스와 매핑 표를 §4.3 에 정의*. |
| Permission shape | per-resource 4-boolean (`{view, create, edit, delete}`). |
| Wire 형식 | `{ id, name, description, permissions: { resource: { view, create, edit, delete } } }` — API §12.1 응답 그대로. |

#### 근거

- frontend `PermissionEditor` UI 와 API §12 spec 이 *기존부터* 4축 모델을 전제로 작성돼 있어, 4축 채택은 frontend 변경 0 에 가깝다.
- backend 1차원 (`none|read|write|admin`) 은 표현력이 낮아, 향후 *"이 role 은 read 는 되는데 delete 는 안 됨"* 식의 요구를 표현하지 못한다.
- backend 의 기존 라우트 가드(`requireMinRole`) 는 `requirePermission(resource, action)` 으로 *재설계* 되며, 이는 §4.3 의 라우트-(resource, action) 매핑 표가 source-of-truth.

#### 정책 evaluation 기본값

매핑 누락 시 — **deny by default**. 라우트가 (resource, action) 매핑에 없으면 거부하고 audit 에 `auth.policy_unmapped` 기록.

### 4.3 결정에 따른 후속 작업 (PR 분할)

본 ADR PR 은 *결정 문서만* 머지한다. 코드 변경은 후속 PR 들로 분리.

| PR (예정) | Track | 작업 | 의존 | 마일스톤 |
| --- | --- | --- | --- | --- |
| **M1-PR-G1** | B·X | `backend_api_contract.md §12` 갱신 — 4축 모델 + 라우트-(resource, action) 매핑 표 + deny-by-default 명시 | ADR-0002 | M1 |
| **M1-PR-G2** | B | `domain/rbac.go` (신규) — Permission/Resource/Action 도메인 + 4-boolean PermissionState + Role aggregate. 마이그레이션 `rbac_policies` 테이블 (system + custom roles, 정책 row JSON 또는 정규화). | G1 | M1 |
| **M1-PR-G3** | B | `store/postgres_rbac.go` — RBAC store CRUD + 시스템 role seed. | G2 | M1 |
| **M1-PR-G4** | B | `httpapi/rbac.go` 확장 — `GET/PUT /api/v1/rbac/policies`, `GET/PUT /api/v1/rbac/subjects/:id/roles`. write audit (`rbac.policy.updated`, `rbac.role.assigned`). | G3 | M1 |
| **M1-PR-G5** | B | `httpapi/authz.go` 재설계 — `requireMinRole` → `requirePermission(resource, action)`. 라우트-매핑 테이블. router 갱신. in-memory cache + invalidation hook. | G4 | M1 |
| **M1-PR-G6** | F | `rbac.types.ts` 정리 + `rbac.service.ts` 의 미연결 path 들이 실제 동작하도록 검증 (이미 작성됨). `PermissionEditor` 가 backend 데이터를 받도록 `organization/page.tsx` 수정. | G4 | M1 또는 M2 Phase 6.1 흡수 |
| Docs | X | 통합 로드맵 §4.2 "Phase 6.1 — 정책 CRUD 연동, RBAC Guard 실체화" 의 *완료 정의* 를 본 ADR 의 G1-G6 PR 들의 머지로 정의. | G1 | M1 (PR-G1 흡수) |

#### M1 sprint backlog 영향

- T-M1-06 (RBAC 결정) → ADR-0002 머지로 종결.
- 신규 task **T-M1-09 ~ T-M1-14** (G1-G6) 가 M1 backlog 에 추가됨 — 별도 backlog 갱신 PR 또는 PR #20 후속 commit 으로 반영.
- M1 sprint capacity 영향: 6 PR 추가 (~600~900 LoC backend + 마이그레이션 + frontend 통합 + 테스트). PR-B (envelope/types/WS) 와 동시 진행 시 M1 종료 일정이 1 sprint 가량 연장될 가능성 있음 — sprint planning 재조정 필요.

---

## 5. 대안의 재검토 트리거

다음 사건 발생 시 본 ADR 을 재검토한다.

- **자원 owner / scope 기반 enforcement** 요구 추가 (예: "이 risk 의 owner 만 수정") — 4-boolean (resource, action) 모델로 표현 불가능. Resource Attribute / ABAC 으로 확장 검토.
- 캐시 일관성 문제 — 다중 인스턴스 환경에서 정책 변경 전파가 운영 부담으로 부각될 경우, pub/sub 기반 invalidation 또는 versioning 강화.
- 정책 변경 빈도가 매우 낮아 (월 1회 미만) DB 매트릭스 + 캐시 인프라가 *과설계* 로 판명될 경우, Option B (read-only viewer) 로 *축소* 재검토.
- DevHub 가 외부 앱의 RBAC 마스터 역할을 수행해야 한다는 요구 (ADR-0001 IdP 확장과 유사) — Keto 등 별도 권한 서비스 도입 재검토.

---

## 6. 미해결 / 후속 결정 항목

본 ADR 채택 후에도 별도 spec 또는 PR 의 implementation note 로 풀어야 할 항목:

1. **라우트-(resource, action) 매핑 표** — `httpapi/router.go` 의 보호 라우트 (현 5종 + 후속 추가) 를 §12 의 5 resources × 4 actions 매트릭스 좌표로 1:1 매핑. PR-G1 의 산출물.
2. **Subject-role 단/다 할당** — 현재 backend `users.role` 은 단일 필드. §12.3·12.4 의 "roles" (배열) 를 단순 *최상위 role 1개* 로 좁힐지, 다중 role 할당 + rank 합산으로 확장할지. 1차에서는 *단일 role 유지* 권장 — 후속 phase 에서 확장.
3. **시스템 role 의 편집 가능 범위** — `system_admin` 의 권한 매트릭스를 UI 에서 변경 가능하게 둘지, 코드 invariant 로 잠글지. UI 는 이미 `sysadmin` 삭제 차단 (`PermissionEditor.tsx:32`) — backend 도 동일하게 `system_admin` row 의 mutation 거부 권장. PR-G4 결정.
4. **Cache 전략** — 단일 인스턴스 운영 가정 시 in-memory + 변경 시 reload 로 충분. 다중 인스턴스 진입 시 pub/sub 또는 polling 기반 invalidation 필요 — 운영 phase 에서 결정.
5. **Test coverage** — 라우트 가드 매트릭스 테스트 (resource × action × role 16+ 케이스), policy CRUD 테스트, deny-by-default 회귀 테스트.

---

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-08 | 초판 작성, Option B 권장 (proposed) | M1 sprint DoD #6 결정 단계, claude/m1-adr-0002-rbac 브랜치 |
| 2026-05-08 | **Option A 채택 (accepted)** + 모델 통일 결정 (per-resource 4-boolean) + PR-G1~G6 분할 + deny-by-default 정책 명시 | 사용자 결정 (M1 sprint DoD #6) |
