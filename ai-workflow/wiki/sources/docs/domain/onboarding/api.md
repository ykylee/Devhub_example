---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/onboarding/api.md]
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

# onboarding 도메인 API

- 문서 목적: onboarding 도메인 (`/api/v1/me/onboarding`, `/api/v1/organizations/search`, `/api/v1/admin/users/:id/review`) 의 API 계약을 정의한다.
- 범위: API-32 (확장), API-33 (확장), API-83..86. envelope/공통 enum 은 master `docs/backend_api_contract.md` §1–§2 또는 `docs/api/conventions.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §16 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [컨셉](./concept.md), [master API](../../backend_api_contract.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md), [ADR-0021](../../adr/0021-onboarding-self-service-unit-selection.md)

## 개요

Keycloak 인증 통과 + DevHub 프로필 미완료 사용자의 self-service 초기 등록 흐름의 endpoint spec. 컨셉: [`./concept.md`](./concept.md). 요구사항: [REQ-FR-ONBOARD-001..012, REQ-NFR-ONBOARD-001..008](./requirements.md). Usecase: [`UC-ONBOARD-01..11`](../../planning/system_usecases.md). 아키텍처: [`./architecture.md`](./architecture.md) (ARCH-ONBOARD-01..06).

본 문서의 모든 endpoint 는 sprint `claude/onboarding-arch-2026-05-21` 에서 *spec only* (planned) 로 staged. backend 구현은 carve out — `IMPL-onboarding-*` sprint 에서 활성화.

## 1. API ID 인덱스

| API ID | 본문 위치 | 항목 | 상태 |
| --- | --- | --- | --- |
| `API-32` | §2 | `GET /api/v1/me` (응답 shape 확장 — `onboarding_required` flag) | **기존 endpoint 응답 확장** (sprint `claude/onboarding-arch-2026-05-21`) |
| `API-83` | §3 | `POST /api/v1/me/onboarding` (제출) | spec staged |
| `API-84` | §4 | `GET /api/v1/organizations/search` (typeahead) | spec staged |
| `API-85` | §5 | `PATCH /api/v1/me` (self-service primary_unit_id 변경) | spec staged |
| `API-33` | §6 | `POST /api/v1/users` (admin 사전 등록 — 동작 명시) | **기존 endpoint 동작 명시** (sprint `claude/onboarding-arch-2026-05-21`) |
| `API-86` | §7 | `POST /api/v1/admin/users/:user_id/review` (review_status transition) | spec staged |

## 2. GET `/api/v1/me` — 응답 확장 (API-32)

기존 `GET /api/v1/me` 의 응답 shape 에 다음 필드를 신규 추가한다:

```json
{
  "status": "ok",
  "data": {
    "user_id": "alice",
    "email": "alice@example.com",
    "display_name": "Alice Kim",
    "primary_unit_id": "team-platform",
    "current_unit_id": "team-platform",
    "role": "developer",
    "onboarding_required": true,
    "onboarding_completed_at": null,
    "review_status": null
  }
}
```

- **`onboarding_required`** (boolean, 신규 — 항상 응답) — `true` 이면 사용자는 미완료 상태 (UC-ONBOARD-01).
  - `true` 조건: DB row 미존재 OR `onboarding_completed_at IS NULL`.
  - `false` 조건: row 존재 + `onboarding_completed_at IS NOT NULL` (`review_status` 가 `pending_review` 또는 `reviewed` 어느 것이든).
- **`onboarding_completed_at`** (timestamp, nullable, 신규) — 완료 시점 (UTC ISO8601). 미완료 시 `null`.
- **`review_status`** (string, nullable, 신규) — `null | "pending_review" | "reviewed"`. 미완료 시 `null`.
- **token-only actor** (DB row 미존재) 의 응답 — `display_name`/`email`/`role` 은 Keycloak token claim 에서 추출, `primary_unit_id`/`current_unit_id` 는 `null`, `onboarding_required: true`.

## 3. Onboarding 제출 — `POST /api/v1/me/onboarding` (API-83)

- **인증**: OIDC (token-only actor 도 호출 가능 — DB row 가 없어도 미완료 사용자가 제출할 수 있어야 함).
- **gating**: `onboardingGate` allowlist 포함 (REQ-FR-ONBOARD-009).
- **요청 (JSON)**:

```json
{
  "display_name": "Alice Kim",
  "primary_unit_id": "team-platform"
}
```

- **요청 필드 제약**:
  - `display_name`: 필수, 1~100자.
  - `primary_unit_id`: 필수, `organization_units(unit_id)` FK.
  - **role 필드는 받지 않는다** — payload 에 `role` 이 포함되어도 무시 (REQ-FR-ONBOARD-002, REQ-FR-ONBOARD-008). Keycloak claim 매핑 + fallback `developer` 로만 결정.
- **응답 — 201 Created** (단일 트랜잭션 성공 — `users` row **INSERT (DB 미등록 사용자) 또는 UPDATE (관리자 사전 등록된 미완료 사용자)** + `onboarding_completed_at=NOW()` + `review_status=pending_review` + audit emit; REQ-FR-ONBOARD-003 의 "row INSERT 또는 UPDATE" + UC-ONBOARD-08 정합. POST /dev-requests / POST /users 패턴과 일관):

```json
{
  "status": "ok",
  "data": {
    "user_id": "alice",
    "email": "alice@example.com",
    "display_name": "Alice Kim",
    "primary_unit_id": "team-platform",
    "role": "developer",
    "onboarding_required": false,
    "onboarding_completed_at": "2026-05-21T08:30:00Z",
    "review_status": "pending_review"
  }
}
```

- **에러 — 422** `invalid_payload` (필드 누락/길이 위반).
- **에러 — 404** `unit_not_found` (`primary_unit_id` 가 organization_units 에 없음).
- **에러 — 409** `onboarding_already_completed` (이미 `onboarding_completed_at IS NOT NULL` 인 사용자가 중복 호출 — self-service 소속 변경은 `PATCH /me` 사용).
- **Audit**: `account.onboarding_completed` emit (ARCH-ONBOARD-06).

## 4. 조직 검색 — `GET /api/v1/organizations/search` (API-84)

- **인증**: OIDC (token-only actor 도 호출 가능 — onboarding 화면에서 조직 검색 필요).
- **gating**: `onboardingGate` allowlist 포함.
- **쿼리**:
  - `q` (string, 필수, 2자 이상 — 1자 이하면 422).
  - `limit` (int, 선택, 기본 20, 최대 20).
- **권한 가드 없음** — 모든 사용자에게 모든 조직 후보 노출 (REQ-FR-ONBOARD-004).
- **응답 — 200 OK**:

```json
{
  "status": "ok",
  "data": [
    { "unit_id": "team-platform", "name": "AI/플랫폼팀" },
    { "unit_id": "team-platform-infra", "name": "AI/플랫폼/인프라" }
  ],
  "meta": { "limit": 20, "total_matched": 17 }
}
```

- **응답 필드 제약** — `unit_id` + `name` 만 노출 (조직명만 표시, REQ-FR-ONBOARD-004). 계층/parent 정보 미포함 — 트리 picker 는 기존 `/api/v1/organization/hierarchy` 응답 재사용.
- **검색 매칭** — `name` 의 case-insensitive substring 매치를 기본. 정확한 매칭 알고리즘 (prefix vs substring vs full-text) 은 IMPL carve 에서 확정.
- **에러 — 422** `invalid_query_params` (q 미지정 / q 1자 이하 / limit > 20).

## 5. Self-service 프로필 변경 — `PATCH /api/v1/me` (API-85)

- **인증**: OIDC + 본인 row 만.
- **gating**: 본 endpoint 는 `onboardingGate` allowlist **외** — 즉 미완료 사용자는 호출 불가. 미완료 사용자는 `POST /me/onboarding` 으로 첫 제출.
- **요청 (JSON)** — 일부 필드 update (PATCH):

```json
{
  "display_name": "Alice K.",
  "primary_unit_id": "team-data"
}
```

- **요청 필드 제약**:
  - `display_name`: 선택, 변경 시 1~100자.
  - `primary_unit_id`: 선택, 변경 시 `organization_units(unit_id)` FK.
  - 두 필드 모두 미포함 시 422 `invalid_payload`.
  - **role 필드는 받지 않는다** (PATCH 통한 권한 변경 차단 — REQ-FR-ONBOARD-002 정합).
- **응답 — 200 OK**:

```json
{
  "status": "ok",
  "data": {
    "user_id": "alice",
    "display_name": "Alice K.",
    "primary_unit_id": "team-data",
    "review_status": "pending_review"
  }
}
```

- **부수 효과** — `primary_unit_id` 가 변경되면 `review_status` 가 자동으로 `pending_review` 로 되돌려진다 (REQ-FR-ONBOARD-007, UC-ONBOARD-07). `display_name` 만 변경 시 `review_status` 영향 없음.
- **에러 — 404** `unit_not_found`.
- **에러 — 422** `invalid_payload`.
- **Audit**:
  - `display_name` 변경: 기존 user.profile.updated 패턴 (별도 audit, 본 PR scope 외).
  - `primary_unit_id` 변경: `account.unit_changed` emit (ARCH-ONBOARD-06).

## 6. 관리자 사전 등록 — `POST /api/v1/users` (API-33 확장)

기존 `POST /api/v1/users` (master `docs/backend_api_contract.md` §10.3, API-33) 의 의미를 명시. ADR-0020 sub-carve E (PR #244) 머지 후 password 분기가 제거된 상태이므로 본 endpoint 는 **admin 사전 등록** 의미로 정착.

- **인증**: OIDC + RBAC `users:create` (system_admin).
- **요청 (JSON)** — 허용 필드:

```json
{
  "user_id": "bob",
  "email": "bob@example.com",
  "display_name": "Bob Lee",
  "primary_unit_id": "team-search"
}
```

- **요청 필드 제약**:
  - `user_id`: 필수, system-wide unique.
  - `email`: 필수.
  - `display_name`: 선택, 미입력 시 `user_id` 로 fallback.
  - `primary_unit_id`: 선택, 미입력 시 `null` (사용자가 onboarding 화면에서 입력).
  - **role 필드는 받지 않는다** — onboarding 동일 정책. Keycloak claim 매핑 + fallback `developer`.
- **응답 — 201 Created**:

```json
{
  "status": "ok",
  "data": {
    "user_id": "bob",
    "email": "bob@example.com",
    "display_name": "Bob Lee",
    "primary_unit_id": "team-search",
    "role": "developer",
    "onboarding_completed_at": null,
    "review_status": null
  }
}
```

- **부수 효과** — `onboarding_completed_at` 은 `null` 로 시작. 사전 등록된 사용자도 첫 로그인 시 onboarding 화면에서 정보 확인/수정 후 제출해야 완료 처리 (REQ-FR-ONBOARD-008).
- **에러 — 409** `user_id_conflict`.
- **에러 — 404** `unit_not_found` (`primary_unit_id` 제공 시).
- **에러 — 422** `invalid_payload`.
- **Audit**: `user.created` (기존 패턴 유지).

## 7. 관리자 검토 confirm — `POST /api/v1/admin/users/:user_id/review` (API-86)

- **인증**: OIDC + RBAC `users:edit` (system_admin).
- **path**: `:user_id` = 검토할 사용자.
- **요청 (JSON)** — body 없음 (transition 만):

```json
{}
```

- **응답 — 200 OK**:

```json
{
  "status": "ok",
  "data": {
    "user_id": "alice",
    "review_status": "reviewed",
    "reviewed_at": "2026-05-21T09:15:00Z",
    "reviewed_by": "admin-1"
  }
}
```

- **에러 — 404** `user_not_found`.
- **에러 — 409** `review_already_confirmed` (이미 `review_status='reviewed'`).
- **에러 — 422** `onboarding_not_completed` (사용자가 아직 `onboarding_completed_at IS NULL` — 검토 대상 아님).
- **Audit**: `account.review_confirmed` emit (ARCH-ONBOARD-06).
- **`reviewed_at` / `reviewed_by` source-of-truth**:
  - **권장 (default)**: audit_logs join — `account.review_confirmed` event 의 `created_at` (→ `reviewed_at`) + `actor` (→ `reviewed_by`). 추가 schema 없음, audit 가 single source-of-truth.
  - **대안**: `users` 테이블에 `reviewed_at timestamptz NULLABLE` + `reviewed_by text NULLABLE` 컬럼 신규. read latency 우월하나 schema overhead.
  - 응답에는 **항상 노출** (위 sample 처럼). 어느 source 든 동일 응답 shape 보장. 최종 결정은 IMPL carve 에서 (default = audit_logs join 권장).

## 8. 공통 에러 코드 (Onboarding 신규)

```
onboarding_required          # 403, gating 차단
onboarding_already_completed # 409, 중복 제출
unit_not_found               # 404, primary_unit_id FK 위반
review_already_confirmed     # 409, 이미 reviewed
onboarding_not_completed     # 422, review 대상 아님
```

기존 공통 에러 (`invalid_payload` / `invalid_query_params` / `user_id_conflict` 등) 는 master `docs/backend_api_contract.md` §1 의 공통 에러 코드 카탈로그 재사용.

## 9. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §16 (전체) 을 도메인 sub-document 로 이관. ID(API-32/33 확장, API-83..86) 보존, 신규 발급/삭제 없음. |
