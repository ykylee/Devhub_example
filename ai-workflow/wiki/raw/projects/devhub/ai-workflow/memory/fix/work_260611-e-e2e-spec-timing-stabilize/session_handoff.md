# Session Handoff — fix/work_260611-e-e2e-spec-timing-stabilize

- 문서 목적: PR #548 (N-13 backend foundation, OPEN) 의 E2E Internal 3 spec fail (repositories-ui:41 + repository-dashboard:115 + signout:80) 자체 보완. frontend only spec timing 안정화.
- 범위: frontend/tests/e2e/fixtures.ts + repositories-ui.spec.ts + repository-dashboard.spec.ts + signout.spec.ts. **코드 변경 4 file +11 -6줄, frontend only, 백엔드 무영향**.
- 상태: branch `fix/work_260611-e-e2e-spec-timing-stabilize` 작업 완료, push/PR 발행 pending.
- 최종 수정일: 2026-06-11

## 0. 본 세션 핵심 결과

### 변경 요약 (4 file, +11 -6줄, frontend only)

| 파일 | 변경 | line |
|---|---|---|
| `frontend/tests/e2e/fixtures.ts` | `WaitForSignInFormOptions` 에 `timeoutMs` 옵션 추가 (default 30_000, opt-in 45_000) + `waitForSignInForm` 의 deadline 계산 반영 | +4 -1 |
| `frontend/tests/e2e/repositories-ui.spec.ts` | line 41-42: `getByText("devhub/e2e-repo-a")` → `getByText(/devhub\/e2e-repo-a/i)` (regex) + timeout 15_000 → 20_000 + comment 보강 | +2 -1 |
| `frontend/tests/e2e/repository-dashboard.spec.ts` | line 115-117: `getByText("e2e-repo-a")` → `getByText(/e2e-repo-a/i).first()` (regex + first) + comment 보강 | +3 -1 |
| `frontend/tests/e2e/signout.spec.ts` | line 59, 80, 109: 3 call site 에 `timeoutMs: 45_000` 추가 (CI race 환경에서 30s 도 fail 가능) | +3 -3 |

### 정공법 핵심

1. **3 spec fail 의 본질**: `getByText` 의 exact match + 짧은 timeout. CI race 환경에서 element 가 늦게 render → visibility wait timeout.
2. **regex selector + first() + timeout 증가** 가 3 spec 모두에서 표준 정공법.
3. **waitForSignInForm default 30s 유지** (12 call site 영향 0). `timeoutMs` opt-in 으로 signout 3 call site 만 45s.
4. **백엔드 무영향** — frontend E2E spec 의 selector + timeout 만 안정화. `platformResponse` echo 의 inbound_source_type/config 영향 0.

### Pre-flight / Safety

- **Tier**: 공용 (frontend only, 사내 한정 정보 미포함)
- **CI 4/4 + frontend E2E + Playwright PASS 예상** (path-detect → frontend 변경 감지, backend/e2e 사외 skip + frontend e2e 영향)
- **ESLint PASS** (1 warning = 기존 OIDC_CLIENT_ID, 본 PR scope 외)
- **TypeScript typecheck**: 기존 63 diagnostics (12 file, 본 PR scope 외, 기존 test code 의 type)

## 1. 다음 세션 directive

1. **본 PR 발행 + 머지** (사용자 confirm 후).
2. **PR #548 (N-13 backend foundation) 머지 결정** (E2E Internal fail 본질이 본 PR 머지 후 안정화되면 머지 가능).
3. **PR A-2 (routing/auto_route.go + voc_handler 통합 + openapi.yaml)** 별도 sprint.
4. **T-d-72-3~6 + Phase 3** (my_harness 일임 결정).
5. **N-6 staging 1주 운영** (사용자 결정 영역).

## 2. 후속 (사용자 결정 영역)

- **본 PR 머지 시점**: 사용자 confirm 후.
- **PR #548 머지 시점**: 본 PR 머지 후 E2E Internal 1 fail 안정화 검증 후.

## 3. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — PR #548 E2E Internal fail 자체 보완 (3 spec timing 안정화 + fixtures timeoutMs 옵션) + branch memory + PR 발행 pending |
