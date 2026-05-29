# auth-session 도메인

- 문서 목적: `auth-session` 도메인의 SDLC 진입점 — REQ / ARCH / API / TC 문서 link + 4 계층 모듈 매핑 + 관련 ADR.
- 범위: Keycloak OIDC 연동을 통한 브라우저 토큰 라이프사이클 및 로그인 플로우 통제.
- 대상 독자: 도메인 contributor, 후속 reviewer, AI agent.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.1](../../governance/code-taxonomy.md), [traceability/report.md §3](../../traceability/report.md)

## 1. 도메인 정의

> Keycloak OIDC 연동을 통한 브라우저 토큰 라이프사이클 및 로그인 플로우를 통제한다. ([code-taxonomy.md §2.1.1](../../governance/code-taxonomy.md))

이 도메인은 외부 IdP (Keycloak) 가 자격 증명을 소유하므로 backend repository 계층은 없고, view + service 계층이 토큰 발급/갱신/만료 흐름과 PKCE/OIDC discovery 를 담당한다.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/auth-session/view/` (`auth.go`, `me.go`, `identity_resolver.go`, `handler.go`) | `frontend/app/login/`, `frontend/app/auth/{callback,logout}/`, `frontend/shared/ui-foundation/layout/AuthGuard.tsx` |
| service | (Keycloak verifier 만) | `frontend/domain/auth-session/service/` (`auth.service.ts`, `refresh-scheduler.ts`, `session-death.ts`, `pkce.ts`, `refresh.ts`, `role-routing.ts`, `token-store.ts`) |
| repository | — (Keycloak IdP 가 소유) | — |
| schema | 토큰 Claims 구조체 (OIDC/PKCE 데이터 교환 모델) | (frontend 내장) |

의존 도메인: [rbac-permissions](../rbac-permissions/), [audit-ops](../audit-ops/)

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 이관 — `docs/domain/auth-session/test_cases.md`) |
| Concept | `./account_redesign.md` | planned (Phase 2 — `docs/domain/auth-session/account_redesign.md`) |

## 4. 관련 ADR

- ADR-0006 (auth gate baseline)
- ADR-0019 (Keycloak 단일화 — supersede ADR-0001)
- ADR-0020 (Keycloak event listener)
- ADR-0024 (WS `?access_token=` query + ticket 패턴)

## 5. cross-cutting 참조

- `docs/architecture.md` §11~12 (auth + signup, Phase 3 split 대상)
- `docs/backend_api_contract.md` §11 (Auth API)
- `docs/governance/code-taxonomy.md` §2.1.1
- `docs/traceability/report.md` §3 매트릭스 row (Phase 4 재구성 대상)

## 6. E2E spec

- `frontend/tests/e2e/auth.spec.ts`
- `frontend/tests/e2e/signout.spec.ts`
