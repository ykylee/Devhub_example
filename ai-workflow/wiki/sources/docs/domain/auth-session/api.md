---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/auth-session/api.md]
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
related: [none]
status: draft
contradictions: [none]
---

# auth-session 도메인 API

- 문서 목적: Keycloak OIDC 단일 IdP 기반 인증 API 계약 (Bearer token 경계, `/api/v1/me`, `/api/v1/account/password`) 을 정의한다.
- 범위: API-19 (Bearer token 경계), API-32 (`GET /me` — 기본 응답), API-35 (self-service 비밀번호 변경, polled 폐기 권고). `/api/v1/me` 의 onboarding_required 확장은 `docs/domain/onboarding/api.md` 참조. RBAC API 는 `docs/domain/rbac-permissions/api.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/backend_api_contract.md` §11 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [ADR-0001 superseded](../../adr/0001-idp-selection.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 개요

DevHub 인증 경계는 Keycloak 기반 OIDC 표준 흐름을 사용한다. Go Core는 토큰 검증과 actor 매핑, 권한/감사 정책 enforcement를 담당하며, 자체 `/api/v1/auth/*` 프록시 API는 제공하지 않는다.

정책 기준은 [ADR-0019 Keycloak 단일화 (현재 결정)](../../adr/0019-keycloak-only-idp.md), [ADR-0001 IdP selection (Hydra+Kratos, superseded)](../../adr/0001-idp-selection.md), [architecture.md](./architecture.md) 를 따른다.

## 1. 소유권 분리

| 영역 | Source of truth | DevHub 역할 |
| --- | --- | --- |
| 조직/사용자 master data | Go Core `users`, `org_units`, `unit_memberships` | 사용자/조직 CRUD, 권한/소속 조회, audit |
| credential, recovery, session | Keycloak | identity, password, recovery, session lifecycle |
| OAuth2/OIDC token | Keycloak | authorization code, token, JWKS/introspection |
| frontend session UX | Next.js + OIDC client | 로그인/로그아웃/세션 갱신 UI orchestration |

`users.user_id`는 OIDC `sub`와 안정적으로 매핑되는 내부 식별자로 운영한다. credential secret과 토큰 원문은 DevHub API 응답 및 audit payload에 포함하지 않는다.

## 2. OIDC 표준 endpoint

다른 앱과 DevHub frontend는 IdP의 OIDC 표준 endpoint를 사용한다. Go Core는 아래 endpoint를 재정의하지 않는다.

| endpoint | 용도 |
| --- | --- |
| `/.well-known/openid-configuration` | issuer, authorization/token/JWKS endpoint discovery |
| `.../protocol/openid-connect/auth` | authorization code flow 시작 |
| `.../protocol/openid-connect/token` | code/token 교환 |
| `.../protocol/openid-connect/logout` | RP-initiated logout |
| `.../protocol/openid-connect/userinfo` | 사용자 claim 조회 |
| `.../protocol/openid-connect/certs` (또는 discovery `jwks_uri`) | JWT signature 검증 |

운영 환경에서는 issuer, audience, JWKS URI, clock-skew 허용치를 환경별 설정으로 주입한다.

## 3. Go Core Bearer token 경계 (API-19)

Go Core `/api/v1/*` 라우터는 `Authorization: Bearer <token>`을 받으면 configured verifier에 위임한다.

- verifier 성공 시 `subject`, `login`, `role` claim을 내부 request context에 저장하고 command/audit actor로 사용한다.
- verifier 실패 시 `401 unauthenticated`를 반환한다.
- `X-Devhub-Actor` fallback 헤더는 [ADR-0004](../../adr/0004-x-devhub-actor-removal.md) (2026-05-13)로 폐기됐고, prod 코드는 해당 헤더를 처리하지 않는다.

## 4. DevHub 계정/권한 관련 API 범위

현재 DevHub API에서 인증 연계로 유지하는 endpoint는 다음 범위다.

| method/path | 목적 | audit action |
| --- | --- | --- |
| `GET /api/v1/me` | OIDC subject 기준 DevHub user profile/role/org context 조회 | 없음 |
| `POST /api/v1/account/password` | **본인 비밀번호 변경** (OIDC 인증 하 self-service) | `account.password_self_change` |
| ~~`POST /api/v1/accounts`~~ | ~~시스템 관리자 계정 발급~~ (**폐기**) | ~~`account.created`~~ |
| ~~`PUT /api/v1/accounts/{user_id}/password`~~ | ~~시스템 관리자 강제 비밀번호 재설정~~ (**폐기**) | ~~`account.password_reset`~~ |
| ~~`DELETE /api/v1/accounts/{user_id}`~~ | ~~시스템 관리자 계정 회수/비활성화~~ (**폐기**) | ~~`account.disabled`~~ |

`/api/v1/auth/login`, `/api/v1/auth/logout`, `/api/v1/auth/token`, `/api/v1/auth/signup`, `/api/v1/auth/consent` 는 제거된 legacy endpoint다.

> `/api/v1/me` 의 onboarding_required + onboarding_completed_at + review_status 확장(API-32 응답 확장)은 `docs/domain/onboarding/api.md` §2 참조.

## 5. 비밀번호 변경 (`POST /api/v1/account/password`, API-35)

> **상태(2026-05-19 이후)**: ADR-0019 단일화 + sprint -ad 의 self-account 폐기 작업으로 본 endpoint 는 dead path (`h.cfg.KratosLogin == nil` 503) 다. spec 은 historical 보존 — 사용자의 비밀번호 변경은 Keycloak Account Console 로 redirect 한다.

self-service 비밀번호 변경 API. 호출자는 자신의 OIDC access token (`Authorization: Bearer ...`)으로 인증한다.

요청 body:

```json
{ "current_password": "OldPass-1!", "new_password": "NewPass-2!" }
```

응답 200:

```json
{ "status": "ok", "data": { "user_id": "alice" } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | `validation` | 신규 비밀번호 정책 위반 |
| 400 | (없음) | body parse 실패 또는 입력값 불량 |
| 401 | `current_password_invalid` | 현재 비밀번호 불일치 |
| 401 | `reauth_required` | 세션 만료/재인증 필요 |
| 500 | (없음) | IdP 호출 실패 또는 내부 invariant 위반 |

## 6. Audit log 매핑

| event | action | target_type |
| --- | --- | --- |
| 계정 발급 | `account.created` | `account` |
| 계정 회수/비활성화 | `account.disabled` | `account` |
| 관리자 비밀번호 재설정 | `account.password_reset` | `account` |
| 본인 비밀번호 변경 성공 | `account.password_self_change` | `user` |
| 본인 비밀번호 변경 실패 (현재 비번 오류) | `account.password_self_change.invalid_current` | `user` |
| token 기반 command 생성 | command별 action | command target |
| RBAC 권한 거부 | `auth.role_denied` | `route` |

비밀번호 평문/해시, recovery token, 세션 시크릿, access token 원문은 audit payload에 저장하지 않는다.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §11 (계정 및 인증 본문) 을 도메인 sub-document 로 이관. ID(API-19, API-32 기본, API-35) 보존, 신규 발급/삭제 없음. onboarding-specific 응답 확장은 onboarding api 로 분리. |
