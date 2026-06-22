---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/rbac-permissions/test_cases.md]
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# rbac-permissions 도메인 테스트 케이스

- 문서 목적: RBAC 정책/강제/동기화/로그아웃/추적성 요구사항을 검증하는 테스트 케이스를 정의한다.
- 범위: `REQ-RBAC-001..016`, `UC-RBAC-01..07`
- 대상 독자: backend/frontend 개발자, QA, 리뷰어
- 상태: active
- 최종 수정일: 2026-06-02
- 관련 문서: [requirements.md](./requirements.md), [api.md](./api.md), [architecture.md](./architecture.md), [system_usecases.md](../../planning/system_usecases.md), [traceability report](../../traceability/report.md)

## 1. 테스트 케이스 목록

| TC ID | 시나리오 | 검증 포인트 | 관련 REQ | 관련 UC |
| --- | --- | --- | --- | --- |
| `TC-RBAC-POLICY-01` | 정책 조회/수정 + invariant 검증 | 시스템 role immutable, `audit` 리소스 C/E/D 강제 false, 캐시 재적재 | REQ-RBAC-001,002,003,008,009 | UC-RBAC-01 |
| `TC-RBAC-ROUTE-01` | 라우트 권한 강제 | routePermissionTable 기반 허용/거부 | REQ-RBAC-004 | UC-RBAC-03 |
| `TC-RBAC-DENY-01` | 미매핑 라우트 차단 | deny-by-default + `auth.policy_unmapped` | REQ-RBAC-005,014 | UC-RBAC-03,05 |
| `TC-RBAC-ROW-WRITE-01` | row ownership write 검증 | 허용 규칙(관리자/허용 role/owner) + deny 시 `auth_row_denied` | REQ-RBAC-006 | UC-RBAC-04 |
| `TC-RBAC-ROLE-SYNC-01` | onboarding role sync | IdP role/group 이 `users.role`에 반영 | REQ-RBAC-007,010 | UC-RBAC-02 |
| `TC-RBAC-ROLE-SYNC-02` | event listener role sync | role 변경 이벤트 반영 + audit 기록 | REQ-RBAC-010 | UC-RBAC-02 |
| `TC-RBAC-ROLE-DRIFT-01` | role drift fail-closed | RBAC 보호 endpoint 403 + `auth.role_sync_required` | REQ-RBAC-010A,014 | UC-RBAC-02,05 |
| `TC-RBAC-LEGACY-01` | legacy role 금지 | `manager`/`team_manager` 신규 할당 거부, `team_manager` alias migration만 허용 | REQ-RBAC-011 | UC-RBAC-02 |
| `TC-RBAC-LOGOUT-01` | logout API 동작 | `POST /api/v1/auth/logout` 에서 세션 정리 + revoke(가능 환경) + audit | REQ-RBAC-012 | UC-RBAC-06 |
| `TC-RBAC-LOGOUT-02` | FE signout 연계 | `/devhub/auth/signout`이 logout API orchestration route로 동작 | REQ-RBAC-012A | UC-RBAC-06 |
| `TC-RBAC-ROW-READ-01` | List read scope 필터 | scope 밖 row 제외, 빈 목록 허용 | REQ-RBAC-013,015 | UC-RBAC-04 |
| `TC-RBAC-ROW-READ-02` | Get read scope 차단 | scope 밖 단건 조회 시 403 + `auth_row_denied` | REQ-RBAC-013,015 | UC-RBAC-04 |
| `TC-RBAC-CODE-01` | 거부 코드 표준화 | `auth.policy_unmapped`/`auth.row_denied`/`auth.role_sync_required`만 사용 | REQ-RBAC-014 | UC-RBAC-05 |
| `TC-RBAC-TRACE-01` | 우선순위/추적성 동기화 | RBAC REQ 변경 PR에서 traceability 동시 갱신 확인 | REQ-RBAC-016 | UC-RBAC-07 |

## 2. 실행 레벨 가이드

- `UT` 권장:
  - `TC-RBAC-POLICY-01`, `TC-RBAC-ROUTE-01`, `TC-RBAC-DENY-01`, `TC-RBAC-ROW-WRITE-01`, `TC-RBAC-CODE-01`
- `IT` 권장:
  - `TC-RBAC-ROLE-SYNC-01`, `TC-RBAC-ROLE-SYNC-02`, `TC-RBAC-ROLE-DRIFT-01`, `TC-RBAC-LEGACY-01`
- `E2E` 권장:
  - `TC-RBAC-LOGOUT-02`, `TC-RBAC-ROW-READ-01`, `TC-RBAC-ROW-READ-02`
- `Process/Review`:
  - `TC-RBAC-TRACE-01`

## 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-02 | 초안 작성 — RBAC 고도화 요구사항(REQ-RBAC-010A/012A/015/016 포함) 기준 TC-RBAC-POLICY/ROUTE/ROLE-SYNC/DRIFT/LOGOUT/ROW-READ/TRACE 카탈로그 신규 발급. |
