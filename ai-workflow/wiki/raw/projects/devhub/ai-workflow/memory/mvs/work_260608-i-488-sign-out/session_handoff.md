# Session Handoff — mvs/work_260608-i-488-sign-out (N-8 / P1-6)

- 문서 목적: 본 sprint 의 진입 상태 + 산출물 + 다음 sprint 진입점을 인계한다.
- 범위: `POST /api/v1/auth/logout` 신규 endpoint (access token 폐기 + session 종료).
- 대상 독자: 후속 sprint 진입자 (Claude 정식 인계 시), 리뷰어 (Codex 외부 리뷰).
- 상태: in_progress (scaffold 단계)
- 최종 수정일: 2026-06-08
- 관련 문서: [`work_backlog.md`](./work_backlog.md), [`state.json`](./state.json), [Issue #488](https://github.com/ykylee/Devhub_example/issues/488), [release_v1_roadmap.md §3.2 P1-6 + §3.5 N-8](../../../../docs/planning/release_v1_roadmap.md).
- 브랜치: `mvs/work_260608-i-488-sign-out` (base `origin/main` @ `7609000`, PR #494 merge commit).

## 1. 본 sprint 작업 목표

v1.0 release 안정성 carve (P1-6) 해소. 세션 관리 기본. OIDC 로그인 흐름의 정상 종료 보장.

### Background
- Keycloak OIDC 로그인 + Sign In / token refresh 는 main 에서 동작 ✓
- 그러나 **Sign Out 경로 부재** — frontend 가 token 폐기/세션 종료 호출할 endpoint 없음
- 통합 테스트 BUG-03 (2026-06-01) 시나리오: user 가 Sign Out 클릭해도 backend 가 session 을 정리하지 않음 → 다음 로그인 시 stale session 누적

### In-scope (본 sprint)
- `POST /api/v1/auth/logout` handler (`internal/httpapi/auth_logout.go` 또는 기존 `auth*.go` 확장)
- Request body: (없음) — access token 은 Authorization header
- 동작:
  1. Bearer token verify (Keycloak OIDC) — `BearerTokenVerifier.Verify(ctx, token)`
  2. Token 폐기: Keycloak admin client 의 logout endpoint 호출 OR token revocation endpoint
  3. Session 종료: `IdentityAdmin.LogoutUserSession(ctx, identityID)` 호출
  4. Audit emit: `auth.logout` (actor, session_id, identity_id)
  5. Response 200: `{ "status": "logged_out" }` (idempotent — 이미 만료된 token 이어도 200)
- Audit emit + metric (선택)
- Idempotency: 같은 token 으로 두번 호출해도 두번 다 200 (race-safe)
- UT 4-5건: TC-AUTH-LOGOUT-01..05 (happy / no token / expired token / unknown identity / concurrent)

### Out-of-scope (본 sprint)
- refresh token rotate — issue #488 spec 옵션. 사용자 결정 후 별도 작업 가능
- Frontend (Sign Out 버튼 UI) — Gemini 후속
- Session 영구화 (DB 저장) — ADR-0020 sub-carve D 의 JWKS stale-while-error 와 별개

## 2. 산출물 (예정)

- `internal/httpapi/auth_logout.go` — handler
- `internal/httpapi/auth_logout_test.go` — UT 4-5건
- `internal/httpapi/router.go` — `v1.POST("/auth/logout", handler.signOut)` 1 line 추가
- `docs/traceability/report.md` — §9 row 추가
- PR 1건 — 본 branch → main

## 3. 다음 sprint 진입 안내

본 sprint 머지 후:
1. **P1-2 sub-carve D** (JWKS stale-while-error expiry) — Claude sprint-i 동시
2. **P0-2 UI polish 마무리** — Gemini 주도
3. **N-9 frontend widget** (repository build-runs) — Gemini
4. **N-11 CI e2e+backend-integration job 복원** (#419) — Codex + 사용자

## 4. 인계 노트 (Claude 정식 인계 시)

- 본 session 은 Mavis (MiniMax-Code) 가 사용자 직접 진입으로 시작. Claude 정식 인계 시 본 scaffold + state.json 의 todo_done / todo 부터 이어가면 됨.
- `docs/governance/worker_division.md` §1.4 OpenCode Lane 정의에 따라 Mavis 가 직접 backend handler 작성은 Lane 2 (cross-cutting validation) 영역 외. **Lane 1 (workflow curation) 진입** 으로 분류. Claude 정식 진입 시 Lane 2/3 으로 재분류 필요.
- Keycloak admin client (`keycloak_admin_client.go`) 의 `LogoutUserSession` 가 본 sprint 의 핵심 의존성. 사용 가능 여부 + 시그니처 본 sprint 진입 시 확인 필수.
- OIDC token revocation endpoint vs Keycloak admin logout endpoint 의 선택지가 있음 (RFC 7009 vs Keycloak-specific). Keycloak 은 양쪽 다 지원. 본 sprint 는 Keycloak-specific 사용 (admin endpoint) — frontend 가 token 폐기 만 필요한 경우 RFC 7009 가 가벼움. 사용자 결정 필요.
