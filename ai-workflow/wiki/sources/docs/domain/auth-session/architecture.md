---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/auth-session/architecture.md]
git_commit: e91115f0
git_branch: chore/260622-wiki-drift-cleanup-2
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T04:24:49Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# auth-session 도메인 아키텍처

- 문서 목적: Keycloak OIDC 단일화 기반 인증·세션 운영의 데이터 모델·인증 흐름·운영 경계를 정의한다.
- 범위: master `docs/architecture.md` §6.2 (User ↔ Account 분리) + §6.5.1 (Keycloak 버전 pin) + §6.5.2 (event listener 동기화) 의 도메인-local 본문 인용. cross-cutting §6.1/§6.3 (RBAC 단계화)/§6.4 (audit) 일반 정책은 master 에서 유지.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/architecture.md` §6 본문 부분 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [api.md](./api.md), [master architecture](../../architecture.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0022 KC 25.0 pin](../../adr/0022-keycloak-version-pin-25-0.md), [ADR-0023 KC 26.0 pin](../../adr/0023-keycloak-version-pin-26-0.md), [ADR-0024 WS ticket](../../adr/0024-websocket-auth-query-token.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 1. 사용자(User) ↔ 계정(Account) 도메인 분리

DevHub는 사람 단위 식별(User)과 인증 자격(Account)을 분리해 관리합니다. 자세한 정책은 [requirements.md](./requirements.md) 를 참조합니다. 본 문서는 그 정책을 만족하기 위한 데이터 모델과 인증 흐름만 정의합니다.

### 1.1 데이터 모델

```text
users (이미 존재)
  user_id        text  PK
  email          text  unique
  display_name   text
  role           text  CHECK in (developer, team_manager, system_admin)
  status         text  CHECK in (active, pending, deactivated)
  idp_subject    text  unique      -- OIDC subject 매핑
  primary_unit_id, current_unit_id, is_seconded, joined_at, ...
```

인증 credential(비밀번호/세션/복구)은 Keycloak 이 소유하고, DevHub는 사용자/조직 메타데이터와 권한 모델을 소유합니다.

### 1.2 비밀번호 처리 원칙

- 비밀번호 평문은 어떤 경로로도 저장/로깅하지 않습니다. 핸들러 진입 직후 즉시 해시로 변환하고 평문 변수의 수명은 최소화합니다.
- 해시 알고리즘은 bcrypt(cost ≥ 12) 또는 argon2id 중 하나를 선택하며, 선택 결과를 `password_algo` 컬럼에 저장해 향후 알고리즘 회전을 가능하게 합니다.
- 비밀번호 강도는 운영 정책으로 별도 정의하되, 최소 길이/금지 패턴 검사는 핸들러 입력 검증 단계에서 수행합니다.
- 강제 재설정(시스템 관리자) 후 다음 로그인은 비밀번호 변경을 강제하기 위해 계정 상태를 `password_reset_required` 로 설정합니다.

> **2026-05-19 이후**: ADR-0019 단일화로 비밀번호 처리는 Keycloak 책임. 위 원칙은 historical 보존이며 코드 경로는 Keycloak Account Console redirect 로 통합됨.

### 1.3 인증 흐름 (1차)

> **결정 (2026-05-07 [ADR-0001](../../adr/0001-idp-selection.md) → 2026-05-19 [ADR-0019](../../adr/0019-keycloak-only-idp.md) 으로 supersede)**: DevHub 인증은 **Keycloak OIDC** 표준 흐름으로 통일한다. `users` 는 사람·조직 master 로 유지하고, credential·session lifecycle 은 IdP가 소유한다. ADR-0001 의 Hydra+Kratos 원본 결정은 PR #167 (2026-05-18) 로 Keycloak 단일화로 전환됐고 ADR-0019 가 결정 사후 명문화. ADR-0001 본문은 historical context 로 immutable 보존.

흐름 (사용자가 DevHub Next.js 에서 로그인하는 first-party 케이스 기준):

1. 브라우저가 DevHub Next.js `/login` 에 진입하면 Next.js 는 Keycloak authorization endpoint로 Authorization Code + PKCE 흐름을 시작합니다.
2. 인증 성공 후 callback 에서 token endpoint 호출로 ID Token + Access Token (+ 필요 시 Refresh Token)을 발급받습니다.
3. Go Core 는 인입 요청의 Bearer token 을 issuer/JWKS 기준으로 검증하고, `sub` claim 을 DevHub actor(`users.idp_subject`)와 매핑합니다.
4. `X-Devhub-Actor` fallback 헤더는 [ADR-0004](../../adr/0004-x-devhub-actor-removal.md) (2026-05-13) 기준 폐기되어 prod 코드에서 처리하지 않습니다.
5. 다른 앱도 동일 IdP에 OIDC client 로 등록해 표준 흐름을 사용합니다.

## 2. Keycloak 버전 pin + SPI event listener (master §6.5.1, §6.5.2)

본 절은 §1 의 보안 baseline 위에, 2026-05-21 이후 코드에 반영된 인증 운영 사실을 추가 명문화한다. 기존 §1.3 OIDC 흐름은 변경 없이 유지된다.

### 2.1 Keycloak 버전 pin

- DevHub 는 IdP 를 Keycloak 단일화([ADR-0019](../../adr/0019-keycloak-only-idp.md))하면서 운영 환경의 Keycloak 컨테이너 버전을 명시 pin 한다: [ADR-0022](../../adr/0022-keycloak-version-pin-25-0.md)(25.0) → [ADR-0023](../../adr/0023-keycloak-version-pin-26-0.md)(26.0).
- 버전 변경 시 admin bootstrap env 가 silent fail 하지 않도록, 26+ 표준(`KC_BOOTSTRAP_ADMIN_USERNAME/PASSWORD`)과 25.x legacy(`KEYCLOAK_ADMIN/KEYCLOAK_ADMIN_PASSWORD`)를 양쪽 동시 주입하는 것을 운영 기준으로 한다(E2E Keycloak realm bootstrap 정합).
- JWKS 검증기(`internal/auth`)는 issuer/audience(ClientID) validation + RS256/384/512 만 허용하며, key rotation 직후 kid mismatch 시 1회 forced refetch + retry, Keycloak unreachable 시 `stale-while-error` fallback(default 24h cutoff)으로 DevHub uptime 을 보장한다. stale 사용 중에는 revoked key 보호가 제한적으로 깨질 수 있어 rotation 직후 운영 SOP(강제 재시작 / cache flush)는 별도 carve 다.

### 2.2 Keycloak event → audit_logs 동기화 (SPI + polling)

Keycloak 에서 발생한 사용자/관리자 이벤트(로그인, group/role 변경, 계정 enable/disable, USER:DELETE 등)는 두 경로로 DevHub `audit_logs` + `users` 동기화에 반영된다.

- **Push (SPI)**: Keycloak event listener SPI(Java, `infra/idp/keycloak-event-listener-spi/`)가 이벤트를 `POST /api/v1/internal/keycloak-events` 로 전송한다. 이 endpoint 는 일반 OIDC 가 아닌 `X-Webhook-Secret` 상수 비교(fail-closed)로만 인증하며 v1 그룹 미들웨어(인증/RBAC) 밖에 등록된다([ADR-0020](../../adr/0020-account-user-management-boundary.md) §5.6 push 경로).
- **Poll (cron)**: `internal/audit` 의 Keycloak event 폴러가 Admin REST(`/admin/realms/{realm}/events` + `/admin-events`)를 기본 30s 주기로 polling 해 cursor(`event_cursors`, migration 000031) 이후 이벤트를 audit 으로 emit + `users` profile/membership/status sync(ADR-0020 sub-carve C)한다.
- **dedup**: push 와 poll 이 동시 존재할 수 있으므로(SPI push 단일화는 미전환 부채), distinguishing 7-tuple SHA-256 을 `audit_logs.source_event_id`(`source_type=keycloak_event`, partial UNIQUE migration 000032)에 기록해 at-least-once 중복을 흡수한다.
- audit source_type 카탈로그: `oidc | webhook | keycloak_event | system`(legacy `kratos` enum 은 historical row decode 용으로만 보존, ADR-0001 superseded).

## 3. WebSocket ticket 인증 경계 (master §6.5.3 cross-link)

본 도메인은 토큰/세션 운영 책임을 가지나 WebSocket ticket 패턴([ADR-0024](../../adr/0024-websocket-auth-query-token.md)) 자체는 realtime 도메인의 인증 경계로 분리된다. WS subscribe 이후 event-type 별 RBAC 재검사 정책 등은 `docs/domain/realtime/architecture.md` 및 master `docs/architecture.md` §6.5.3 참조.

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §6.2 (User/Account 분리) + §6.5.1/§6.5.2 (Keycloak pin + event sync) 본문을 도메인 sub-document 로 이관. cross-cutting §6.1 (Webhook secret/관리자 접근/audit 일반)/§6.3 (RBAC 단계화)/§6.4 (audit 최소 필드)/§6.5.3 (WS ticket)는 master 또는 다른 도메인에 유지. |
