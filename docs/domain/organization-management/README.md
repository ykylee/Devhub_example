# organization-management 도메인

- 문서 목적: `organization-management` 도메인의 SDLC 진입점.
- 범위: 인사 조직 트리, 직무 Appointments, 사용자 마스터 프로필의 변경 및 조회.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.4](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> 인사 조직 트리, 직무 Appointments, 사용자 마스터 프로필의 변경 및 조회를 제어한다. ([code-taxonomy.md §2.1.4](../../governance/code-taxonomy.md))

순환 의존성 검사 + 임원 할당 규칙 + HRDB 연동 + 사용자 lookup 을 통한 조직/사용자 마스터 운영.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/organization-management/view/` (`organization.go`, `organizations_search.go`, `hr_lookup.go`, `handler.go`) | `frontend/app/admin/settings/{organization,users}/`, `frontend/domain/organization-management/view/MemberManagementModal.tsx`, `MemberTable.tsx`, `OrgNode.tsx`, `OrgUnitGrid.tsx`, `UserCreationModal.tsx` |
| service | 조직 노드 유효성 (순환 의존성), 임원 할당 규칙 | `frontend/domain/organization-management/service/identity.service.ts` |
| repository | `backend-core/internal/domain/organization-management/repository/users_units.go` | — |
| schema | `domain/primary_unit.go`, `domain/user.go`, DB: `users`/`org_units`/`unit_appointments` (000004/000019), `org_units_total_count_mv` (000011) | (frontend 내장) |

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | `./requirements.md` | planned (Phase 3, 기존 `backend_requirements_org_hierarchy.md` 통합) |
| ARCH | `./architecture.md` | planned (Phase 3) |
| API | `./api.md` | planned (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 — `docs/tests/test_cases_m3_organization.md`) |
| Spec | `./org_chart_ux_spec.md` (기존) | active |
| Spec | `./organizational_hierarchy_spec.md` (기존) | active |
| Spec | `./backend_requirements_org_hierarchy.md` (기존) | active |

## 4. 관련 ADR

- ADR-0008 (organization model)
- ADR-0009 (organization sync)
- ADR-0010 (HRDB integration)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §13 (org / user API)
- `docs/architecture.md` §15~17 (organization architecture)

## 6. E2E spec

- `frontend/tests/e2e/admin-org-crud.spec.ts`
- `frontend/tests/e2e/admin-users-crud.spec.ts`
