<!--
PR # — N-8 (sprint -i) frontend logout status 분기.
backend PR #495/#496 머지 후속. issue #488 spec §"Frontend (Gemini — 후속, sprint -i)".
-->

## 카테고리 · 모듈

- 카테고리: `auth-session` / frontend
- 모듈: `frontend/domain/auth-session/service/auth.service.ts`, `frontend/domain/auth-session/service/auth.service.test.ts`, `frontend/tests/e2e/auth.spec.ts`, `docs/traceability/report.md`

## 요약

backend `POST /api/v1/auth/logout` 의 204/401/502 응답을 frontend logout 흐름에서 정확히 분기. 502 (Keycloak unreachable) 시 OIDC 단계 건너뛰고 `useStore.addToast(error)` + 강제 `/login`. 5xx/network 시 warning toast + OIDC redirect. 204/401 (idempotent) 시 OIDC `end_session_endpoint` 로 RP-initiated logout. token store / actor cleanup 은 backend 응답 "이후" 로 이동 (Bearer 헤더 보호).

## 변경 상세

### Frontend (`frontend/domain/auth-session/service/auth.service.ts`)

- `logout()` 함수 분리:
  - `postBackendLogout(accessToken, refreshToken, idToken)` — backend POST 호출, status → `ok` / `expired` / `unreachable` / `error` 4 outcome.
  - `redirectToOIDCEndSession(discovery, runtimeConfig, idToken)` — OIDC end_session_endpoint URL 빌드 + `window.location.assign`.
- status 분기:
  - **204** (ok) / **401** (expired, idempotent) → OIDC RP-initiated logout
  - **502** (unreachable) → `addToast(error, "Sign-out service is temporarily unreachable. Your session has been cleared locally.")` + 강제 `/login` redirect (OIDC 단계 skip)
  - 기타 4xx/5xx / network throw → `addToast(warning, "Sign-out encountered a non-fatal error. Continuing local cleanup.")` + OIDC redirect
- cleanup 순서 변경: 기존엔 fetch 후 즉시 `tokenStore.clear()` + `clearActor()` — 변경 후엔 backend 응답 받은 "이후" 로 이동. backend 가 Bearer 헤더로 access token 을 검증하는 동안 frontend 가 먼저 토큰을 비우면 middleware 가 401 로 단락시키므로, 응답 받은 뒤 cleanup.

### Frontend vitest (`frontend/domain/auth-session/service/auth.service.test.ts`)

- `useStoreAddToast` mock 추가 + `useStoreGetState` 가 `addToast` 노출.
- 기존 4 test: backend status `200` → `204` (spec 정합) 으로 갱신.
- 신규 6 test (**UT-frontend-auth-05**):
  - FE-01: 204 → token cleanup + OIDC redirect (id_token_hint 포함)
  - FE-02: 401 (idempotent) → 동일 cleanup + OIDC redirect
  - FE-03: 502 → `addToast(error)` + `/login` 강제 redirect (OIDC 단계 호출 안 됨 검증)
  - FE-04: 503 (other 5xx) → `addToast(warning)` + OIDC redirect
  - FE-05: network throw → `addToast(warning)` + OIDC redirect + cleanup
  - FE-06: access token 없을 때 Bearer header 미포함 + 정상 흐름

### E2E (`frontend/tests/e2e/auth.spec.ts`)

- 신규 `TC-E2E-LOGOUT-01`: developer 시드 login → `/auth/signout` → backend POST 204 + OIDC `end_session_endpoint` 호출 (post_logout_redirect_uri=`/login` 검증) → `/login` 도달.

### Traceability (`docs/traceability/report.md`)

- §2.4 (IMPL): `frontend-auth-07` + `frontend-service-auth-02` 추가, frontend IMPL count 35 → 37.
- §2.5 (UT): `UT-frontend-auth-05` 추가.
- §2.6 (E2E): `TC-E2E-LOGOUT-01` 추가.
- §3.1 (auth-session row 375): IMPL/UT/TC cross-ref 갱신.
- §4 변경이력 (487줄 다음): sprint -i 진행 log 1 row 추가 (날짜 역순 유지).

## 추적성 영향

- 추가:
  - `IMPL-frontend-auth-07` — frontend logout status 분기 + cleanup sequence
  - `IMPL-frontend-service-auth-02` — `postBackendLogout` + `redirectToOIDCEndSession` private method 분리
  - `UT-frontend-auth-05` — vitest 6건 (FE-01..06)
  - `TC-E2E-LOGOUT-01` — playwright logout full roundtrip
- 갱신:
  - §2.4 Frontend IMPL count 35 → 37 (frontend-auth-07 + frontend-service-auth-02)
  - §2.5 Frontend Vitest list (UT-frontend-auth-05 추가)
  - §2.6 §3 auth-session row TC list (TC-E2E-LOGOUT-01 추가)
  - §3.1 auth-session row 375 cross-ref (frontend-auth-07 + frontend-service-auth-02 + UT-frontend-auth-05 + TC-E2E-LOGOUT-01)
  - §4 변경이력 — sprint -i row 추가 (날짜 2026-06-08, 본 PR)
- Deprecate: 없음
- 매트릭스: `docs/traceability/report.md` §2.4 / §2.5 / §2.6 / §3.1 / §4 5섹션 cross-ref 갱신

## 검증

- frontend `npx vitest run` → 80 test files / 1030 tests PASS (기존 1024 + 신규 6)
- frontend `auth.service.test.ts` 단독 → 32 tests PASS
- (CI 환경) frontend type check + lint + E2E shard 1/2/3 + frontend unit test
- backend 영향 0 (singleton private method 추가만)
