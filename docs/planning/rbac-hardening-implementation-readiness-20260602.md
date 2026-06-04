# RBAC 고도화 구현 준비 문서 (2026-06-02)

- 문서 목적: 확정된 RBAC 요구사항/유스케이스/테스트케이스를 구현 태스크로 전개한다.
- 범위: REQ-RBAC-010A/012A/013/014/015/016 중심 + 기존 REQ-RBAC-001..009 정합.
- 상태: active
- 최종 수정일: 2026-06-02
- 입력 SoT:
  - `docs/domain/rbac-permissions/requirements.md`
  - `docs/domain/rbac-permissions/architecture.md`
  - `docs/domain/rbac-permissions/api.md`
  - `docs/planning/system_usecases.md` (UC-RBAC-01..07)
  - `docs/domain/rbac-permissions/test_cases.md`
  - `docs/traceability/report.md`

## 1. 구현 준비 점검 결과

1. REQ-UC-TC 체인
- REQ-RBAC-001..016 ↔ UC-RBAC-01..07 ↔ TC-RBAC-* 매핑 완료.

2. 주요 정합 상태
- RBAC 역할 모델 문서는 `developer/team_manager/system_admin` 기준으로 정합됨.
- read scope 결합 규칙(List filter / Get 403), role drift fail-closed, signout 연계 정책 문서화 완료.

3. 주의사항
- `docs/domain/rbac-permissions/keycloak_groups_mapping.md` 및 일부 과거 이력 문서는 legacy role(`manager/team_manager`) 표기를 포함한다.
- 본 구현은 **requirements/architecture/api 최신 문서를 SoT**로 따르고, legacy 표기는 migration context로만 해석한다.

## 2. 구현 단위 (Execution Units)

### EU-1. Role Drift Fail-Closed
- 목표: drift 감지 시 RBAC 보호 엔드포인트를 `403 + auth.role_sync_required`로 일관 거부.
- 관련 REQ/UC/TC:
  - REQ-RBAC-010A, REQ-RBAC-014
  - UC-RBAC-02, UC-RBAC-05
  - TC-RBAC-ROLE-DRIFT-01, TC-RBAC-CODE-01
- 예상 변경 파일:
  - `backend-core/internal/httpapi/auth.go`
  - `backend-core/internal/httpapi/permissions.go`
  - `backend-core/internal/audit/*` (event code 연계)

### EU-2. Read Scope Enforcement (List/Get)
- 목표: route-level 통과 후 read scope를 결합해 List는 필터, Get은 403 처리.
- 관련 REQ/UC/TC:
  - REQ-RBAC-013, REQ-RBAC-015
  - UC-RBAC-04
  - TC-RBAC-ROW-READ-01, TC-RBAC-ROW-READ-02
- 예상 변경 파일:
  - `backend-core/internal/httpapi/platforms.go`
  - `backend-core/internal/httpapi/projects.go`
  - `backend-core/internal/store/*` (list/get query scope 주입)

### EU-3. 거부 코드 표준화
- 목표: `auth.policy_unmapped`, `auth.row_denied`, `auth.role_sync_required` 외 거부 코드 제거.
- 관련 REQ/UC/TC:
  - REQ-RBAC-014
  - UC-RBAC-05
  - TC-RBAC-CODE-01, TC-RBAC-DENY-01
- 예상 변경 파일:
  - `backend-core/internal/httpapi/permissions.go`
  - `backend-core/internal/httpapi/*` (권한 거부 응답 지점)

### EU-4. FE Signout → API Logout 연계
- 목표: `/devhub/auth/signout`가 `POST /api/v1/auth/logout` orchestration 경로로 동작.
- 관련 REQ/UC/TC:
  - REQ-RBAC-012, REQ-RBAC-012A
  - UC-RBAC-06
  - TC-RBAC-LOGOUT-01, TC-RBAC-LOGOUT-02
- 예상 변경 파일:
  - `frontend/lib/services/auth.service.ts`
  - `frontend/app/**/signout*` 또는 auth route handler
  - `backend-core/internal/httpapi/*logout*`

### EU-5. 추적성/문서 동기화 가드
- 목표: RBAC 변경 PR마다 traceability row 자동/수동 체크.
- 관련 REQ/UC/TC:
  - REQ-RBAC-016
  - UC-RBAC-07
  - TC-RBAC-TRACE-01
- 예상 변경 파일:
  - `docs/traceability/report.md`
  - `.github/pull_request_template.md` (필요 시)

## 3. 구현 순서 권장

1. EU-3 (거부 코드 표준화)
2. EU-1 (drift fail-closed)
3. EU-2 (read scope 결합)
4. EU-4 (logout 연계)
5. EU-5 (추적성 최종 동기화)

이 순서는 공통 에러 경계/보안 경계를 먼저 고정해 회귀 범위를 줄인다.

## 4. 완료 기준 (Definition of Done)

1. 기능 DoD
- REQ-RBAC-010A/012A/013/014/015/016에 대응하는 코드 경로가 구현됨.

2. 테스트 DoD
- 최소 통과: TC-RBAC-ROLE-DRIFT-01, TC-RBAC-ROW-READ-01/02, TC-RBAC-LOGOUT-02, TC-RBAC-CODE-01.
- 회귀 없음: 기존 TC-RBAC-SUB-01, TC-RBAC-MGR-01, TC-PERMISSIONS-SMOKE-01.

3. 추적성 DoD
- `docs/traceability/report.md`에 REQ/UC/TC/IMPL 영향 동기화.
- PR 본문 “추적성 영향” 섹션 작성 완료.

## 5. 리스크 및 대응

1. 리스크: legacy role 잔여 문서/코드로 인한 정책 혼선
- 대응: 구현 기준은 `requirements/architecture/api` 최신본 우선.

2. 리스크: read scope 적용 시 성능 저하
- 대응: store query plan 확인 + index 점검 + pagination 회귀 테스트.

3. 리스크: logout 경로 변경으로 FE 라우팅 회귀
- 대응: E2E `TC-RBAC-LOGOUT-02` 우선 실행.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-02 | 구현 착수 전 정합 점검 결과와 실행 단위(EU-1..5), DoD, 리스크 대응을 정리한 준비 문서 신규 작성. |
