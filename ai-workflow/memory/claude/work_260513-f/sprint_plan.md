# Sprint Plan — claude/work_260513-f (B 묶음: RBAC IMPL 정밀 매핑 + API §12 본문 ID 노출)

- 문서 목적: B1 (본문 ID 노출, RBAC 도메인 1차) + B3 (RBAC API §12 IMPL 정밀 매핑) 묶음의 작업 계획.
- 범위: docs only. `backend_api_contract.md` §12 endpoint 헤더 + `docs/traceability/report.md` 갱신. 코드 변경 0.
- 진입 base: main HEAD `ae8b459` (PR #91 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

### B3 + B1 — RBAC 도메인 1차

| 항목 | 위치 | 결과 |
| --- | --- | --- |
| §12 endpoint 헤더 (API-XX) | `docs/backend_api_contract.md` §12.2~§12.10 | 9 헤더에 `(API-26)` ~ `(API-40)` 노출 |
| API-XX → endpoint 매핑 표 | `docs/traceability/report.md` §2.2 끝 | RBAC §12 의 9 항목 명시 표 |
| IMPL-rbac-XX 정의 | `docs/traceability/report.md` §2.4 | rbac-01 handler / 02 store / 03 enforcement / 04 cache |
| §3 RBAC 행 IMPL 컬럼 정밀화 | `docs/traceability/report.md` §3 | endpoint-IMPL 1:1 매핑 |
| §5.2 closed | `docs/traceability/report.md` §5.2 | "RBAC API §12 IMPL 정밀 매핑" → closed |
| §6 변경 이력 | `docs/traceability/report.md` §6 | 본 sprint 한 줄 |

## 2. ID 매핑 (본 sprint 결정)

### API-XX (RBAC §12)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| API-26 | §12.2 | GET /api/v1/rbac/policies |
| API-27 | §12.3 | PUT /api/v1/rbac/policies |
| API-28 | §12.4 | POST /api/v1/rbac/policies (사용자 정의 role 생성) |
| API-29 | §12.5 | DELETE /api/v1/rbac/policies/:role_id (사용자 정의 role 삭제) |
| API-30 | §12.6 | GET /api/v1/rbac/subjects/:subject_id/roles |
| API-31 | §12.7 | PUT /api/v1/rbac/subjects/:subject_id/roles |
| API-38 | §12.8 | 라우트 → (resource, action) 매핑 표 (정책) |
| API-39 | §12.9 | 매핑 누락 정책 (deny-by-default) |
| API-40 | §12.10 | Cache 와 무효화 정책 |

### IMPL-rbac-XX

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| IMPL-rbac-01 | `backend-core/internal/httpapi/rbac.go` | §12.2~§12.7 의 6 endpoint handler + §6 legacy gone (`listRBACPolicies`, `createRBACPolicy`, `updateRBACPolicies`, `deleteRBACPolicy`, `getSubjectRoles`, `setSubjectRoles`, `getRBACPolicyLegacyGone`) |
| IMPL-rbac-02 | `backend-core/internal/store/postgres_rbac.go` | RBAC role + subject-role assignment persistence (`ListRBACRoles`, `GetRBACRole`, `CreateRBACRole`, `UpdateRBACRolePermissions`, `UpdateRBACRoleMetadata`, `DeleteRBACRole`, `GetSubjectRoles`, `SetSubjectRole`) |
| IMPL-rbac-03 | `backend-core/internal/httpapi/permissions.go::routePermissionTable` + `enforceRoutePermission` | §12.8 라우트 매핑 source-of-truth + 미들웨어 enforcement |
| IMPL-rbac-04 | `backend-core/internal/httpapi/permissions.go::PermissionCache` | §12.10 in-memory matrix cache + Invalidate |

## 3. 미진입 / 다음 sprint 후보

본 sprint 의 직속 후속:
- B1 의 다른 도메인 (auth / account / org / command / audit / infra) 본문 ID 노출 — 점진 진행
- B2 — deprecated 문서 식별 + 마킹 (§8 우선순위 4)
- B4 — X-Devhub-Actor 폐기 ADR
- C1 — frontend 컴포넌트 Vitest
- C2 — E2E 신규 TC
- D5 — actionlint
- M3 / M4 진입
