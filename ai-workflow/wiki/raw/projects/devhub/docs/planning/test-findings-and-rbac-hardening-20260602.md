# 통합 테스트 발견사항 + 권한 관리 요구사항 고도화안 (2026-06-02)

- 문서 목적: `integrated_test_report_20260601.md`에서 확인된 버그/이슈를 우선순위로 재정리하고, 신규 Two-Dimensional RBAC 요구사항의 보완 포인트를 정의한다.
- 범위: 인증/권한(RBAC), 세션 관리, SCM/CI 연동 중 v0.1.0 릴리즈 영향 항목.
- 상태: active
- 작성일: 2026-06-02
- 입력 문서:
  - `docs/planning/integrated_test_report_20260601.md`
  - `docs/planning/role-access-concept.md`
  - `docs/domain/platform-lifecycle/requirements.md` §5 (REQ-FR-ROLE-001..016)
  - `docs/domain/rbac-permissions/requirements.md`

## 1. 테스트 발견사항 정리 (우선순위)

### P0 (v0.1.0 차단)

1. **ISSUE-05: CI Run 생성 API 부재**
- 현상: CI Run 생성이 API로 불가능하여 DB 직접 INSERT 필요.
- 영향: CI/CD 운영 플로우 실사용 불가.
- 조치: `POST /api/v0-1/ci-runs` 구현 + status enum validation + traceability ID 발급.

### P1 (v0.1.0 필수 권장)

1. **BUG-02: Keycloak `devhub_role` 동기화 누락**
- 현상: Keycloak role 속성이 DevHub `users.role`/JWT에 반영되지 않음.
- 영향: RBAC 결정과 IdP role이 불일치.
- 조치: onboarding/event listener 양쪽에서 role sync 보장 + drift 감지 로깅.

2. **BUG-03: Sign-out endpoint 미구현**
- 현상: `/devhub/auth/signout` 404.
- 영향: 세션 종료 UX/보안 경계 불완전.
- 조치: `POST /api/v0-1/auth/logout` + 프론트 signout route 연계 + refresh token revoke 정책 명시.

3. **ISSUE-04: Repository build-runs endpoint 미구현/불일치**
- 현상: CI run 데이터 존재해도 repo endpoint 빈 배열.
- 영향: repo 단위 운영 가시성 저하.
- 조치: source-of-truth를 `ci_runs`로 통일.

4. **NEW-P1D: developer `applications:view` 의도 검증**
- 현상: 테스트 보고서에서 권한 매트릭스와 요구사항 간 불일치 가능성 지적.
- 영향: 최소 가시성 정책 해석 차이.
- 조치: `REQ-FR-ROLE-001/002` 기준으로 developer 기본 조회 권한 확정 및 테스트 고정.

### P2~P3 (v0.1.1+ 또는 운영 개선)

1. **BUG-06: Issue/PR 증분 sync 미반영**
- 단기: updated_at pull sync.
- 중기: webhook event 기반 반영.

2. **BUG-04/05/07**
- 기능 차단도 낮음. 운영 noise/호환성/계정 생성 프로시저 보완으로 관리.

## 2. 권한 관리 요구사항 갭 분석

1. `docs/domain/platform-lifecycle/requirements.md`는 이미 Two-Dimensional RBAC(REQ-FR-ROLE-001..016)로 정합.
2. `docs/domain/rbac-permissions/requirements.md`는 아직 구 모델(`manager`, `team_manager`) 중심 표현이 남아 있어 정책 문서 간 충돌 위험.
3. 테스트 결과 기준 필수 보완 요구:
- IdP role sync 보장(requirement level)
- logout API 표준화
- route-level matrix와 row-scope 정책의 테스트 케이스 연결 강화

## 3. 고도화 요구사항 (추가/명확화)

### RBAC-H-001 (P1)
- DevHub system role은 `developer`, `team_manager`, `system_admin` 3종으로 통일하고, `manager`/`team_manager`는 마이그레이션 alias로만 허용한다.

### RBAC-H-002 (P1)
- role source-of-truth는 Keycloak group/attribute이며, DevHub `users.role`은 동기화 캐시로 정의한다.
- 인증 시점마다 token claim ↔ DB role drift 검사 정책을 둔다.

### RBAC-H-003 (P1)
- 로그아웃은 API contract로 명시한다.
- 최소 요구: token revoke(가능 시), 세션 쿠키 정리, audit event 기록.

### RBAC-H-004 (P1)
- `applications:view`, `projects:view`는 route-level 허용 + row-scope 필터를 반드시 함께 enforce한다.
- list/detail 모두 동일한 scope merge 규칙(system role + resource role) 적용.

### RBAC-H-005 (P2)
- 권한 거부 응답 코드 표준화: `auth.policy_unmapped`, `auth.row_denied`, `auth.role_sync_required`.

## 4. 즉시 실행 항목 (권장)

1. `rbac-permissions/requirements.md`를 Two-Dimensional RBAC 기준으로 정합화.
2. BUG-02/BUG-03/ISSUE-04/ISSUE-05를 v0.1.0 sprint backlog에 P0/P1로 재배치.
3. REQ-FR-ROLE-001..016 ↔ E2E TC 매핑표를 `docs/traceability/report.md`에 보강.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-02 | 통합 테스트 발견사항 우선순위 재정리 + RBAC 고도화 요구사항(RBAC-H-001..005) 정의 |
