---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/rbac-permissions/requirements.md]
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# rbac-permissions 도메인 요구사항

- 문서 목적: Role/Resource/Action 매트릭스 기반 다차원 접근 제어의 정책·운영 요구사항을 정의한다.
- 범위: ADR-0002 (RBAC policy edit API), ADR-0007 (RBAC enforcement), ADR-0011 (row-scoping), ADR-0020 결정 D (subject role API 폐기) 의 요구사항-측면 통합. 각 endpoint 의 실제 책임은 `api.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-06-02 (Two-Dimensional RBAC 정합화 + 테스트 발견사항 반영)
- 관련 문서: [도메인 README](./README.md), [architecture.md](./architecture.md), [api.md](./api.md), [keycloak_groups_mapping.md](./keycloak_groups_mapping.md), [master requirements](../../requirements.md), [ADR-0002](../../adr/0002-rbac-policy-edit-api.md), [ADR-0007](../../adr/0007-rbac-enforcement.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 1. 개요

DevHub 의 다차원 접근 제어는 다음 모델을 사용한다.

- **System Role**: 3종(`developer`, `team_manager`, `system_admin`).
- **Resource Role**: 4종(`project_member`, `project_leader`, `application_leader`, `org_head`) — row-level scope 계산에 사용.
- **Resource**: 11종(`infrastructure`, `pipelines`, `organization`, `security`, `audit`, `applications`, `platform_repositories`, `projects`, `scm_providers`, `dev_requests`, `dev_request_intake_tokens`).
- **Action**: 4종(`view`, `create`, `edit`, `delete`).
- **Row-scoping**: ADR-0011 `enforceRowOwnership` + read scope filter 로 row-level 권한 검사.

본 도메인의 REQ-FR ID 는 master `docs/traceability/report.md` §3 매트릭스 (REQ-FR-27 RBAC, REQ-FR-86 row-scoping 등) 에서 관리된다. 본 문서는 master `docs/requirements.md` 의 §2.4 (System Administrator — 권한 격리) + §5 의 다양한 도메인 절(REQ-FR-PROJ-009 row-scoping, REQ-FR-DREQ row-level, REQ-FR-APP-010 status transition 권한 등)에 분산된 RBAC 관련 요구사항을 도메인 관점에서 재집합한 sub-document 이다.

## 2. 정책 요구사항

### 2.1 Role + Resource + Action 모델 (ADR-0002)

- **REQ-RBAC-001 (MVP, 확정):** RBAC 정책은 DB-backed matrix 로 영속화하며 write API (PUT /api/v1/rbac/policies) 로 갱신 가능해야 한다.
- **REQ-RBAC-002 (MVP, 확정):** 각 (resource, action) 좌표는 boolean 이다. 시스템 정의 role 의 id/name/system flag 는 immutable. 사용자 정의 role 은 `custom-{slug}` prefix 강제.
- **REQ-RBAC-003 (MVP, 확정):** `audit` resource 의 `create`/`edit`/`delete` 는 **모든 role 에 대해 false 강제** (audit append-only invariant). store 가 PUT 또는 seed 시 invariant 검증으로 거부한다.

### 2.2 Enforcement 정책 (ADR-0007)

- **REQ-RBAC-004 (MVP, 확정):** `requirePermission` 미들웨어가 라우트 → (resource, action) 매핑 표(`routePermissionTable`)를 source-of-truth 로 enforcement.
- **REQ-RBAC-005 (MVP, 확정):** 매핑 누락 라우트는 **deny-by-default** (`403 Forbidden` + `auth.policy_unmapped` audit) — 신규 라우트 추가 시 매핑 표 갱신을 누락하면 런타임 거부로 알린다.

### 2.3 Row-scoping (ADR-0011)

- **REQ-RBAC-006 (MVP, 확정, REQ-FR-PROJ-009 부합):** Row-level 권한은 `enforceRowOwnership(c, ownerUserID, allowedRoles...)` helper 로 일관 enforce. allow 규칙: (1) `system_admin`, (2) `allowedRoles` 화이트리스트, (3) `actor.login == ownerUserID`. deny 시 `auth.row_denied` audit + `403` + `code=auth_row_denied`.

### 2.4 Subject role 동기화 (ADR-0020 결정 D)

- **REQ-RBAC-007 (MVP, 확정):** DevHub 가 직접 노출하던 `GET/PUT /api/v1/rbac/subjects/{subject_id}/roles` (API-30/API-31) 는 **폐기**. role assignment 는 Keycloak admin console 의 group membership 으로 처리하고, DevHub `users.role` 컬럼은 event listener (audit-ops 도메인) 가 자동 sync 한다.
- **REQ-RBAC-010 (P1, 신규):** Keycloak role/group 정보와 DevHub `users.role` 간 동기화는 onboarding + event listener 양 경로에서 보장해야 한다. drift 감지 시 `auth.role_sync_required` audit 이벤트를 남기고 운영 경고를 발생시킨다.
- **REQ-RBAC-010A (P1, 신규):** role drift 상태에서는 권한 판단을 **fail-closed** 로 처리해야 한다. 즉, RBAC 보호 엔드포인트는 `403` + `code=auth.role_sync_required` 를 반환하고, system_admin이 동기화를 복구하기 전까지 임시 허용(fail-open)을 금지한다.
- **REQ-RBAC-011 (P1, 신규):** 레거시 role(`manager`, `team_manager`)은 신규 부여를 금지하고 `team_manager` alias migration 만 허용한다.

### 2.5 Sign-out / Session 종료 정책

- **REQ-RBAC-012 (P1, 신규):** `POST /api/v1/auth/logout` endpoint 를 제공해야 하며, 최소 동작은 (1) 세션 쿠키 정리, (2) refresh token revoke(가능한 환경), (3) `auth.logout` audit 기록이다.
- **REQ-RBAC-012A (P1, 신규):** 프론트엔드 라우트 `/devhub/auth/signout` 는 직접 세션 종료 로직을 갖지 않고, 반드시 `POST /api/v1/auth/logout` 를 호출하는 orchestration route 로 동작해야 한다. 구형 signout 경로는 API 전환 완료 시 301/302 또는 내부 rewrite 로 정규화한다.

### 2.6 View 권한 + Row Scope 결합 강제

- **REQ-RBAC-013 (P1, 신규):** `applications:view`, `projects:view` 는 route-level 허용만으로 충분하지 않다. list/detail 모두 actor scope(`system role + resource role`) 기반 row filter를 적용해야 한다.
- **REQ-RBAC-014 (P1, 신규):** 권한 거부 에러 코드는 `auth.policy_unmapped`, `auth.row_denied`, `auth.role_sync_required` 3종으로 표준화한다.
- **REQ-RBAC-015 (P1, 신규):** Read scope enforcement 규칙:
  - `List*` 계열: scope 밖 리소스는 에러가 아니라 필터링하여 제외한다(빈 목록 허용).
  - `Get*` 계열: scope 밖 리소스 요청은 `403` + `code=auth_row_denied`.
  - 위 규칙은 `projects`, `applications`, `platform_repositories`, `dev_requests` read API에 동일 적용한다.

### 2.7 우선순위 정합성 규칙

- **REQ-RBAC-016 (P1, 신규):** RBAC 우선순위 충돌 시 source-of-truth 는 본 문서(`docs/domain/rbac-permissions/requirements.md`)와 master 요구사항(`docs/requirements.md`)의 최신 수정일 기준 합의본으로 한다. 도메인 요구사항(`platform-lifecycle` 등)에 남아 있는 P2 표기는 릴리즈 계획에서 P1로 상향될 수 있으며, 상향 시 traceability 매트릭스를 같은 PR에서 동기화해야 한다.

## 3. UI 및 캐시 운영

- **REQ-RBAC-008 (MVP):** PermissionCache(in-memory) 가 PUT/POST/DELETE 후 same-process reload. multi-instance cache 일관성은 운영 phase 진입 시 pub/sub 또는 polling 보강 (carve out).
- **REQ-RBAC-009 (MVP):** UI 노출 정책 — `audit.create/edit/delete` 는 readOnly 표기. 시스템 role id/name 도 readOnly.

## 4. RBAC 단계화 (master architecture §6.3 인용)

| 단계 | 범위 | 기준 |
| --- | --- | --- |
| Phase 1 | Webhook secret 검증, system admin role 분리, 관리자 작업 Audit Log | TASK-007 및 초기 시스템 관리자 기능 구현 기준 |
| Phase 2 | Keycloak 기반 OIDC 도입, DevHub OIDC client 전환, token 검증/actor 매핑/audit 경계 정착 | Keycloak/OIDC 운영 진입 및 backend Phase 13 완료 시점 ([ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0001](../../adr/0001-idp-selection.md) superseded) |
| Phase 3 | Gitea 사용자/조직/저장소 권한 동기화, Repository 하위 Project role 매핑 | Application-Repository-Project 매핑과 관리자 대시보드 확장 시점 |
| Phase 4 | Gitea SSO 연동 기반 통합 인증, 자체 계정과의 병행/대체 정책 결정 | 운영 환경 전환 전 별도 보안 검토 후 도입 |

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-02 | 리뷰 보완 반영: REQ-RBAC-010A(fail-closed role drift), REQ-RBAC-012A(FE `/devhub/auth/signout` → API logout 연계), REQ-RBAC-015(read scope list/get 규칙), REQ-RBAC-016(우선순위 source-of-truth) 추가. |
| 2026-06-02 | Two-Dimensional RBAC 정합화: System Role(`developer/team_manager/system_admin`) + Resource Role(`project_member/project_leader/application_leader/org_head`) 반영. 테스트 발견사항 기반 REQ-RBAC-010..014(역할 동기화, sign-out, view+row-scope 결합, 거부 코드 표준화) 추가. |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` 분산 RBAC 요구사항(§2.4 권한 격리 + §5 row-scoping/transition 권한 분산)을 본 도메인 sub-document 로 재집합. 본문은 master `docs/backend_api_contract.md` §12.0 (모델) + §12.1 (default matrix) + ADR-0002/0007/0011/0020 를 통합. ID는 REQ-RBAC-001..009 도메인 임시 발급(traceability matrix 와 별도 — Phase 4 재구성 시 정합). |
