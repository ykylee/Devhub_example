# Work Backlog — fix/work_260611-e-e2e-spec-timing-stabilize

- 문서 목적: PR #548 (N-13 backend foundation) 의 E2E Internal 3 spec fail 자체 보완 sprint 의 백로그.
- 범위: frontend/tests/e2e/ 4 file. **백엔드 변경 0줄**. **신규 ID 0 row**.
- 상태: in_progress (PR 발행 pending)
- 최종 수정일: 2026-06-11

## 1. 태스크 (sprint)

- [x] WB-01: branch 생성 (`fix/work_260611-e-e2e-spec-timing-stabilize`)
- [x] WB-02: fixtures.ts: WaitForSignInFormOptions timeoutMs 옵션 추가
- [x] WB-03: repositories-ui.spec.ts:41 timing 안정화 (regex + 20s)
- [x] WB-04: repository-dashboard.spec.ts:115 timing 안정화 (regex + first + 20s)
- [x] WB-05: signout.spec.ts 3 call site timeoutMs 45_000
- [x] WB-06: ESLint PASS + TypeScript typecheck (기존 63 diagnostics 본 PR scope 외)
- [x] WB-07: branch memory directory (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-11.md)
- [ ] WB-08: PR 발행 (push + gh pr create)
- [ ] WB-09: PR 머지 (사용자 confirm 후)
- [ ] WB-10: PR #548 (N-13 backend foundation) 머지 결정 (E2E Internal 1 fail 안정화 검증 후)

## 2. 잔여 (별도 sprint)

- PR A-2 (routing/auto_route.go + voc_handler 통합 + openapi.yaml)
- T-d-72-3~6 + Phase 3 (my_harness 일임)
- N-6 staging 1주 운영 (사용자 결정)

## 3. 관련 PR

- **PR #548** (이전 sprint, OPEN, `feat/work_260611-a-n13-inbound-source-impl`): N-13 backend foundation — 본 PR 의 보완 대상
- **본 PR (pending)**: E2E spec timing 안정화 (frontend only)

## 4. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | sprint 시작 + 4 file 변경 + ESLint PASS + branch memory + PR 발행 pending |
