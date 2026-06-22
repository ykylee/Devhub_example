---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/auth-session/requirements.md]
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

# auth-session 도메인 요구사항

- 문서 목적: Keycloak OIDC 단일 IdP 기반 사용자 계정/세션 운영의 기능·비기능·정책 요구사항을 정의한다.
- 범위: 사용자(User) ↔ 계정(Account) 도메인 분리, OIDC 흐름, 비밀번호 정책의 historical 보존, 운영 책임 분리. 사용자 마스터/조직 마스터는 `docs/domain/organization-management/requirements.md` 참조. RBAC 정책은 `docs/domain/rbac-permissions/requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §2.5 본문 이관)
- 관련 문서: [도메인 README](./README.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0001 superseded](../../adr/0001-idp-selection.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md), [ADR-0024](../../adr/0024-websocket-auth-query-token.md)

## 1. 개요

DevHub 사용자(person)와 인증 자격(credential)을 분리해 관리한다. 인증은 Keycloak 기반 OIDC 표준 흐름을 1차 수단으로 사용한다.

> **구현 방식 (2026-05-07 [ADR-0001](../../adr/0001-idp-selection.md) → 2026-05-19 [ADR-0019](../../adr/0019-keycloak-only-idp.md) supersede)**: 본 절의 정책 invariant(1:1 매핑, 비밀번호 평문 미보관, 자동 lock, audit log 대상) 는 그대로 유지하되, 구현은 **자체 `accounts` 테이블이 아닌 Keycloak (단일 IdP)** 가 책임진다. 신규 요구 — DevHub 의 계정 서비스를 다른 앱에도 OIDC IdP 로 제공 — 를 충족하기 위한 결정이다. ADR-0001 (Hydra+Kratos) 의 원본 결정은 PR #167 (2026-05-18) 로 Keycloak 단일화 전환됐고 ADR-0019 가 사후 명문화. 정책 변경 없음.

> **갱신(2026-05-27) — 자체 계정/비밀번호 흐름 = historical**: Keycloak 단일 IdP 전환([ADR-0019](../../adr/0019-keycloak-only-idp.md) / [ADR-0020](../../adr/0020-account-user-management-boundary.md) / [ADR-0021](../../adr/0021-onboarding-self-service-unit-selection.md)) 이 코드까지 완결되면서, 본 절의 **자체 credential store 시절 표현**(아래 "비밀번호 정책"의 해시 저장, "로그인 ID 정책"의 형식 강제, "계정 상태(Account Status)" 4종 직접 관리, "데이터 주권 메모"의 비밀번호 self-service 변경)은 **historical** 로 보존만 한다 — 코드에서는 폐기됐다. 현행:
> - DevHub 는 비밀번호/credential 을 일절 저장·검증하지 않는다. 자체 password form 은 폐기됐고(`/api/v1/account/password` dead path 제거, ADR-0019 sprint -ad), 사용자의 비밀번호 변경은 **Keycloak Account Console 로 redirect** 한다.
> - 로그인 ID/비밀번호 형식·해시·잠금(`active`/`disabled`/`locked`/`password_reset_required`)·강제 재설정·세션 만료는 **Keycloak** 이 단일 책임진다 (`users.idp_subject` 만 DevHub 에 보관, `kratos_identity_id` → `idp_subject` rename, migration 000030).
> - DevHub `users` row 는 프로필/권한 메타데이터 + onboarding/검토 상태(onboarding 도메인)만 관리하며, 사용자 관리 책임 경계는 [ADR-0020](../../adr/0020-account-user-management-boundary.md) 가 확정한다. 정책 invariant(1:1 매핑, 평문 미보관, audit 대상)는 위 2026-05-19 배너대로 그대로 유지된다.

## 2. 핵심 요구사항

- **핵심 니즈:** 식별 가능한 사람 단위 권한 관리, 분실/유출 시 빠른 회수, 감사 가능한 비밀번호 변경 기록.
- **용어 분리:**
    - **사용자(User):** 조직에 소속된 사람. 표시명, 이메일, 직책, 소속 조직 단위(Org Unit), 역할(Role)로 구성.
    - **계정(Account):** 사용자가 DevHub에 로그인하기 위한 인증 자격. credential 원천은 IdP(Keycloak)이며 DevHub는 사용자/권한 메타데이터를 관리.
- **주요 기능 (확정) — DevHub 정책/데이터 관점:**
    - [x] **1:1 매핑 강제:** 사용자 1명은 정확히 0개 또는 1개의 계정을 가질 수 있다. 계정은 반드시 1명의 사용자에 귀속된다. 동일 사용자에 복수 계정을 만들 수 없으며, 1개의 계정이 복수 사용자를 대표할 수 없다.
    - [x] **로그인 ID 정책:** 로그인 ID는 시스템 전역에서 unique 하다. 형식 정책(허용 문자, 길이)은 IdP 정책 표로 관리한다. 사용자 식별자(`user_id`)와는 별도이며, 로그인 ID 변경이 사용자 식별자를 바꾸지 않는다.
    - [x] **비밀번호 정책:** 비밀번호는 평문으로 저장하지 않는다. 단방향 해시(예: bcrypt cost ≥ 12 또는 argon2id)만 저장한다. 사용자 self-service 비밀번호 변경은 허용하며, 관리자 강제 재설정 정책은 IdP 운영 정책을 따른다.
    - [x] **계정 상태(Account Status):** `active`, `disabled`, `locked`, `password_reset_required` 의 4종 상태를 가진다. `locked`는 정책상 자동 잠금(예: 연속 실패) 또는 수동 잠금이며, `disabled`는 계정 회수 상태다.
    - [x] **감사 로그 대상:** 비밀번호 변경(본인), 로그인 성공/실패, 계정 상태 전이(회수/잠금/해제)는 Audit Log 기록 대상이다. 비밀번호 자체나 해시는 Audit Log에 기록하지 않는다.
- **운영 책임 분리 (확정):**
    - [x] **외부 IdP 책임:** 계정 발급/회수, 관리자 강제 비밀번호 재설정, 세션 강제 만료는 Keycloak Admin Console 또는 HRDB ETL 경로에서 수행한다.
    - [x] **DevHub 책임:** 사용자 프로필/권한 메타데이터(`users`, `org_units`) 관리와 감사 추적을 담당한다.
- **주요 기능 (후보):**
    - [ ] 비밀번호 만료 정책(주기 강제 변경) — 운영 단계에서 정책 결정.
    - [ ] 다단계 인증(MFA/2FA) — 초기 단계는 미도입.
    - [ ] Gitea SSO 연동으로 자체 계정 발급 없이 통합 인증 — `architecture.md` 의 RBAC 단계화에서 후속 phase로 관리.
- **데이터 주권 메모:** 사용자 자신은 자신의 계정 정보(로그인 ID, 비밀번호)를 변경할 수 있다. 계정 발급/회수와 강제 재설정은 DevHub API가 아니라 외부 IdP 운영 절차를 따른다.

## 3. 데이터 및 권한 운영 기준 (사용자 계정 row, master §4.1 발췌)

| 데이터 분류 | 원천 | 수정 권한 | 조회 권한 | 보존 기준 | 알림 기준 |
| --- | --- | --- | --- | --- | --- |
| 사용자 계정/자격 | Keycloak + DevHub | 본인은 자기 비밀번호 변경, 계정 발급/회수/강제 재설정은 IdP 운영 절차, DevHub는 사용자 메타데이터/감사 추적 관리 | 본인(자기 자신), 시스템 관리자(전체) | 계정 활성 중 유지, 회수 후 90일 보존 후 삭제 | 비밀번호 변경/잠금/회수 시 본인 + 시스템 관리자 알림 |

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §2.5 (사용자 계정 관리) 본문 + §4.1 의 "사용자 계정/자격" 행을 도메인 sub-document 로 이관. 본문 보존, 신규 발급/삭제 없음. |
