# Phase 6.1 RBAC API 통합 및 M2 인증 콜백 최종화

- 문서 목적: Phase 6.1 RBAC API 통합과 M2 인증 고도화 구현 계획을 정리한다.
- 범위: RBAC 서비스 정비, 인증 콜백 검증, Account 서비스 실데이터 전환
- 대상 독자: 프론트엔드 개발자, 후속 에이전트, 리뷰어
- 상태: draft
- 최종 수정일: 2026-05-10
- 관련 문서: [세션 인계](./session_handoff.md), [작업 백로그](./work_backlog.md), [당일 backlog](./backlog/2026-05-10.md)

이 계획은 프론트엔드 Phase 6.1의 RBAC 실데이터 연동과 M2 단계의 인증 콜백 플로우 및 계정 관리 기능의 최종 통합을 목표로 합니다.

## User Review Required

> [!IMPORTANT]
> `AccountService`의 관리자 기능(계정 발급, 비밀번호 강제 초기화 등)은 현재 백엔드에 전용 엔드포인트가 일부 누락되어 있을 수 있습니다. 기존 `PATCH /api/v1/users/:user_id`를 활용하거나 필요한 경우 백엔드 보강이 병행되어야 합니다.

## Proposed Changes

### 1. RBAC 서비스 정비 (Phase 6.1)

`rbac.service.ts`에서 ADR-0002 도입 이전의 레거시 메서드를 제거하고, `apiClient`를 사용한 표준 CRUD 플로우를 확정합니다.

#### [MODIFY] [rbac.service.ts](../../../../frontend/lib/services/rbac.service.ts)
- `replacePolicy` 메서드 제거 (레거시/미사용).
- `updatePolicies`가 반환하는 전체 매트릭스 데이터를 UI 상태에 올바르게 동기화하는지 재확인.

---

### 2. 인증 콜백 및 세션 처리 (M2)

`/auth/callback` 라우트에서 수행하는 토큰 교환과 사용자 정보 조회가 실제 백엔드 API와 완벽히 연동되도록 보장합니다.

#### [VERIFY] [auth/callback/page.tsx](../../../../frontend/app/auth/callback/page.tsx)
- `authService.exchangeCode` -> `POST /api/v1/auth/token` (Hydra Proxy) 호출 확인.
- `authService.resolveIdentity` -> `GET /api/v1/me` 호출 및 `useStore` 반영 확인.

---

### 3. Account 서비스 실데이터 연동 (M2 Integration)

현재 Mock으로 구현된 `accountService`를 실제 API 호출로 전환합니다.

#### [MODIFY] [account.service.ts](../../../../frontend/lib/services/account.service.ts)
- `apiClient`를 사용하여 백엔드 연동.
- `issueAccount`: `POST /api/v1/users` (password 포함) 호출.
- `disableAccount`: `PATCH /api/v1/users/:user_id` (status: "deactivated") 호출.
- `forceResetPassword`: (백엔드 지원 확인 필요, 우선 `PATCH`로 비밀번호 업데이트 시도).

---

## Verification Plan

### Automated Tests
- `pytest ai-workflow/tests/check_docs.py`를 실행하여 문서 무결성 확인.
- RBAC 정책 저장 후 `GET /api/v1/rbac/policies`를 통해 DB 반영 여부 확인.

### Manual Verification
- **RBAC**: Permission Editor에서 권한 변경 후 'Save' 클릭 시 성공 토스트 확인 및 새로고침 후 유지 확인.
- **Auth**: 로그인 시뮬레이션 후 `/auth/callback`을 거쳐 `/developer` 대시보드로 정상 이동하는지 확인.
- **Admin**: Member Table에서 'Revoke Account' 수행 시 상태가 'deactivated'로 변경되는지 확인.
