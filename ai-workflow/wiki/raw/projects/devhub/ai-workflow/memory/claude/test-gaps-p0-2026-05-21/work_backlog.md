# Work Backlog — sprint `claude/test-gaps-p0-2026-05-21`

- 문서 목적: 본 sprint 의 in-scope / 잔여 carve 분류.
- 상태: 작업 완료
- 최종 수정일: 2026-05-21

## 1. In-scope (완료)

- ✅ P0-2 JWKS metric assertion (5 test 보강 + helper 2건)
- ✅ DREQ IP allowlist UT 보강 (table test 12 + middleware-level 1)
- ✅ DREQ revoke cancel E2E (TC-DREQ-ADMIN-TOKEN-REVOKE-CANCEL-01)

## 2. False-positive 정정

- P0-1 (DREQ intake admin E2E): 직전 sprint 의 gap 평가 잘못 — dev-requests.spec.ts mega test 가 8 TC cover. subagent 보고 spot-check 미흡.

## 3. 다음 sprint carve

| # | 영역 | P | 비고 |
| --- | --- | --- | --- |
| 1 | Keycloak 실 OIDC e2e flow | P1 | login → callback → /me → dashboard |
| 2 | Integration bindings PATCH/DELETE backend handler 테스트 | P1 | integration_registry_test.go 보강 |
| 3 | Single-port nginx e2e (ADR-0018) | P1 | 환경 의존 — 통합 환경 검토 |
| 4 | frontend service unit (auth/api-client/websocket) | P2 | Vitest |
| 5 | backend-ai pytest | P3 | gRPC server 구현 동반 |
| 6 | dashboard page snapshot | P3 | @testing-library/react |
| 7 | pre-existing ESLint 4 errors refactor | P2 | gemini PRs 의 useEffect 패턴 + any types |
| 8 | main flat memory housekeeping | P2 | PR #261 + 본 sprint 머지 후 |
