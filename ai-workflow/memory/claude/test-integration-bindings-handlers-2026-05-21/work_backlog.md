# Work Backlog — sprint `claude/test-integration-bindings-handlers-2026-05-21`

- 문서 목적: 본 sprint 의 in-scope / 잔여 carve.
- 상태: 작업 완료
- 최종 수정일: 2026-05-21

## 1. In-scope (완료)

- ✅ memoryApplicationStore.UpdateIntegrationBinding fake 보강 (FK + 4-tuple unique + Policy/Enabled 갱신)
- ✅ PATCH 5 test (Happy / NotFound / InvalidPolicy / ConflictDuplicate / ForbiddenForDeveloperRole)
- ✅ DELETE 3 test (Happy / NotFound / ForbiddenForDeveloperRole)
- ✅ seedBindingFixture helper

## 2. 다음 sprint carve

| # | 영역 | P | 비고 |
| --- | --- | --- | --- |
| 1 | Keycloak 실 OIDC e2e flow | P1 | login → callback → /me → dashboard |
| 2 | Single-port nginx e2e (ADR-0018) | P1 | 환경 의존 |
| 3 | frontend service unit (auth/api-client/websocket) | P2 | Vitest |
| 4 | main flat memory housekeeping | P2 | PR #261, #262, 본 PR + #263 머지 후 |
| 5 | pre-existing ESLint 4 errors refactor | P2 | gemini PRs useEffect/any |
| 6 | backend-ai pytest | P3 | gRPC 구현 동반 |
| 7 | dashboard page snapshot | P3 | @testing-library/react |
