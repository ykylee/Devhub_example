---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/dev-request/api.md]
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

# dev-request 도메인 API (DREQ)

- 문서 목적: DREQ 도메인 (`/api/v1/dev-requests/*`, `/api/v1/dev-request-tokens/*`) 의 API 계약을 정의한다.
- 범위: API-59..68 + API-79. envelope / 공통 enum / 공통 에러는 master `docs/backend_api_contract.md` §1–§2 또는 `docs/api/conventions.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §14 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [컨셉](./concept.md), [master API](../../backend_api_contract.md), [ADR-0012](../../adr/0012-dreq-external-intake-auth.md), [ADR-0013](../../adr/0013-dreq-rbac-row-scoping.md), [ADR-0014](../../adr/0014-application-project-lifecycle.md).

## 개요

외부 시스템에서 들어오는 개발 의뢰 (Dev Request, DREQ) 의 수신/조회/등록(promote)/거절/재할당/닫기 API. 컨셉: [`./concept.md`](./concept.md). 요구사항: [REQ-FR-DREQ-001..013, REQ-NFR-DREQ-001..006](./requirements.md). Usecase: [`UC-DREQ-01..10`](../../planning/system_usecases.md). 아키텍처: [`./architecture.md`](./architecture.md) (ARCH-DREQ-01..06).

본 문서의 모든 endpoint 는 sprint `claude/work_260515-f` 에서 *spec only* (planned) 로 시작했고, backend 구현은 carve out — DREQ-AuthADR 머지 후 DREQ-Backend sprint 에서 활성화됐다.

## 1. 외부 수신 — `POST /api/v1/dev-requests`  *(API-59)*

- **인증**: 별도 middleware `requireIntakeToken`. [ADR-0012](../../adr/0012-dreq-external-intake-auth.md) 가 옵션 A (API 토큰 + IP allowlist) 채택.
  - `Authorization: Bearer <plain-token>` 헤더 필수.
  - middleware 가 `SHA-256(token)` 으로 `dev_request_intake_tokens` lookup + `allowed_ips` CIDR 검증 + `revoked_at IS NULL` 확인.
  - 실패 시 401 (`auth_intake_token_invalid` / `auth_intake_ip_denied` / `auth_intake_token_revoked`) + audit `dev_request.intake_auth_failed`.
  - 성공 시 audit `dev_request.intake_auth_succeeded` + `last_used_at` 갱신.
- `source_system` 은 토큰의 매핑 값에서 자동 채움 — body 의 self-claim 은 무시 (ADR-0012 §4.1.2 spoofing 방지).
- **요청 (JSON)**:

```json
{
  "title": "Backend 검색 성능 개선",
  "details": "p95 응답시간 2s 이상 발생, 재현 시나리오 첨부.",
  "requester": "ops_portal/user/jane",
  "assignee_user_id": "charlie",
  "external_ref": "OPS-2026-00482"
}
```

- **응답 — 201 Created** (정상 수신, `pending`):

```json
{
  "status": "ok",
  "data": {
    "id": "11111111-2222-3333-4444-555555555555",
    "title": "Backend 검색 성능 개선",
    "details": "...",
    "requester": "ops_portal/user/jane",
    "assignee_user_id": "charlie",
    "source_system": "ops_portal",
    "external_ref": "OPS-2026-00482",
    "status": "pending",
    "registered_target_type": null,
    "registered_target_id": null,
    "rejected_reason": null,
    "received_at": "2026-05-15T04:55:00Z",
    "created_at": "2026-05-15T04:55:00Z",
    "updated_at": "2026-05-15T04:55:00Z"
  }
}
```

- **응답 — 200 OK** (idempotent 재수신, `(source_system, external_ref)` 매칭): 동일 `data` 반환 + `"status":"ok"`.
- **응답 — 201 Created with status=rejected** (검증 실패): assignee 미존재 / 필수 필드 누락 시 `pending` 대신 `rejected (reason: invalid_intake)` 로 저장 — REQ-FR-DREQ-002. audit 보존 목적이며 절대 drop 하지 않는다.
- **응답 — 401** 인증 실패. **400** body schema 위반.
- **Audit**: `dev_request.received` (정상) 또는 `dev_request.received + dev_request.rejected` (검증 실패) emit.

## 2. 목록 — `GET /api/v1/dev-requests`  *(API-60)*

- **인증**: OIDC + RBAC `dev_requests:view`.
- **권한**: `system_admin` / `team_manager` 는 전체. 그 외 role 은 `assignee_user_id == actor.login` 의 row 만 (route-level RBAC + handler 단의 server-side filter).
- **쿼리**: `status` (콤마 다중) / `source_system` / `assignee_user_id` (system_admin 만 의미) / `limit` (기본 50, 최대 100) / `offset`.
- **응답 — 200**:

```json
{
  "status": "ok",
  "data": [ { /* dev_request shape */ } ],
  "meta": { "total": 17, "limit": 50, "offset": 0 }
}
```

- **에러 422** `invalid_query_params`.

## 3. 상세 — `GET /api/v1/dev-requests/:id`  *(API-61)*

- **인증**: OIDC + RBAC `dev_requests:view` + row-level (system_admin / team_manager / assignee 본인).
- **응답 — 200** `{ "status":"ok", "data": <dev_request> }`. **404** not found. **403** `auth_row_denied` (audit `auth.row_denied`).

## 4. Promote (등록) — `POST /api/v1/dev-requests/:id/register`  *(API-62)*

- **인증**: OIDC + RBAC `dev_requests:edit` + row-level (system_admin / team_manager / assignee 본인).
- **요청 schema (mutual exclusion)**: 다음 셋 중 정확히 하나만 채워야 한다. 둘 이상 채우거나 모두 비우면 `400 dev_request_register_payload_invalid`. (sprint `claude/work_260515-m` 도입)
  1. `target_id` (legacy 매핑) — 이미 존재하는 application/project id 로 dev_request 를 묶기만 한다. dev_requests row 만 UPDATE 한다 (단일 row, 트랜잭션 불요).
  2. `application_payload` (target_type=application 필수) — 새 Application 을 생성하고 dev_request 를 registered 로 갱신한다. **단일 Postgres 트랜잭션** (REQ-FR-DREQ-005, ADR-0013 §5). `primary_repo` 필드는 optional 이며 함께 platform_repositories 행 1개를 추가한다.
  3. `project_payload` (target_type=project 필수) — 새 Project 를 생성하고 dev_request 를 registered 로 갱신한다. 단일 Postgres 트랜잭션.

- **요청 (JSON, legacy 매핑)**:

```json
{ "target_type": "application", "target_id": "5e1c..." }
```

- **요청 (JSON, 신규 application 생성)**:

```json
{
  "target_type": "application",
  "application_payload": {
    "key": "PLATFORM26",
    "name": "Platform 2026",
    "description": "DREQ intake 로 생성",
    "owner_user_id": "charlie",
    "leader_user_id": "charlie",
    "development_unit_id": "dept-eng",
    "visibility": "internal",
    "status": "planning",
    "primary_repo": {
      "repo_provider": "gitea",
      "repo_full_name": "org/platform-2026",
      "external_repo_id": "",
      "role": "primary"
    }
  }
}
```

- **요청 (JSON, 신규 project 생성)**:

```json
{
  "target_type": "project",
  "project_payload": {
    "platform_id": "",
    "repository_id": 42,
    "key": "PROJ1",
    "name": "Proj1",
    "owner_user_id": "alice",
    "visibility": "internal",
    "status": "planning"
  }
}
```

- **단일 트랜잭션 효과** (REQ-FR-DREQ-005, ADR-0013 §5, sprint `claude/work_260515-m`): payload 분기에서 (a) 신규 target entity 생성 (+ optional primary_repo link), (b) `dev_requests.status='registered'`, (c) `registered_target_type/id` 갱신이 모두 한 Postgres tx 안에서 일어난다. 부분 실패 시 모두 롤백. audit emit 은 tx commit 이후: `application.created` (또는 `project.created`) + `dev_request.registered` (`payload.created=true`).
- **응답 — 200** (신규 생성 path): registered_target 의 `created=true` 와 함께 생성된 entity body 포함.

```json
{
  "status": "ok",
  "data": {
    "dev_request": { "status": "registered", "registered_target_type": "application", "registered_target_id": "...", ... },
    "registered_target": {
      "target_type": "application",
      "target_id": "...",
      "created": true,
      "application": { /* application response shape */ }
    }
  }
}
```

- **응답 — 200** (legacy target_id path): registered_target 의 `created=false`, entity body 미포함.
- **에러 400** `dev_request_register_target_invalid` (target_type 이 application/project 외) / `dev_request_register_payload_invalid` (payload mutual exclusion 위반). **422** `invalid_application_key` (application_payload.key 정규식 위반) / `invalid_repo_link_role` (primary_repo.role 이 primary/sub/shared 외 — codex hotfix #4, sprint `claude/work_260515-n`) / `unsupported_repo_provider` (primary_repo.repo_provider 가 SCM 카탈로그에 없거나 disabled — codex hotfix #4). **409** `dev_request_already_registered` (status 가 이미 registered/rejected/closed) / `application_key_conflict` / `project_key_conflict` (신규 생성 path 에서 FK 또는 UNIQUE 또는 CHECK 위반 → tx 롤백). **403** `auth_row_denied`.

## 5. 거절 — `POST /api/v1/dev-requests/:id/reject`  *(API-63)*

- **인증**: OIDC + RBAC + row-level (system_admin / team_manager / assignee 본인).
- **요청**: `{ "rejected_reason": "중복 의뢰 (OPS-2026-00481 과 동일)" }` — `rejected_reason` 필수.
- **응답 — 200** `{ "status":"ok", "data": <dev_request with status=rejected> }`. **400** reason 누락. **409** 이미 registered/closed/rejected.

## 6. 재할당 — `PATCH /api/v1/dev-requests/:id`  *(API-64)*

- **인증**: OIDC + RBAC `dev_requests:edit` + **system_admin 만** (담당자 변경은 row owner 가 self-change 불가하도록 RBAC 으로 강제).
- **요청**: `{ "assignee_user_id": "alice" }` — 1차에서는 assignee 만 변경 가능. title/details 등 본문 수정은 carve out.
- **응답 — 200** `{ "status":"ok", "data": <dev_request> }`. audit `dev_request.reassigned` + payload `{from_assignee, to_assignee}`.

## 7. 닫기 — `DELETE /api/v1/dev-requests/:id`  *(API-65)*

- **인증**: OIDC + RBAC `dev_requests:delete` — **system_admin 만**. (REQ-FR-DREQ-008 + ARCH-DREQ-04 의 team_manager 매트릭스가 delete 권한을 부여하지 않음과 정합. codex PR #121 review P1, sprint `claude/work_260515-h` 반영.)
- **전이**: `registered` 또는 `rejected` → `closed`. `pending` / `in_review` 는 거부 (먼저 reject 후 close).
- **응답 — 200** `{ "status":"ok", "data": <dev_request with status=closed> }`. **422** `invalid_status_transition_close` (pending/in_review 에서 시도). audit `dev_request.closed`.

## 8. 에러 코드 카탈로그 (DREQ 신규)

```
dev_request_already_registered
dev_request_invalid_intake
dev_request_idempotency_conflict
dev_request_register_target_invalid
dev_request_register_target_mismatch
dev_request_register_payload_invalid          # promote: target_id / application_payload / project_payload mutual exclusion (sprint m)
dev_request_assignee_not_found
dev_request_reason_required
invalid_status_transition_close
invalid_application_key                        # promote application_payload.key 정규식 (재사용)
invalid_repo_link_role                         # codex hotfix #4 (sprint n): primary_repo.role 의 application-level gate
unsupported_repo_provider                      # codex hotfix #4 (sprint n): primary_repo.repo_provider 의 SCM enablement gate (legacy 재사용)
application_key_conflict                       # promote application 신규 생성 시 FK/UNIQUE/CHECK 위반
project_key_conflict                           # promote project 신규 생성 시 FK/UNIQUE 위반
auth_intake_token_invalid
auth_intake_token_revoked
auth_intake_ip_denied
auth_intake_token_missing
invalid_allowed_ips                            # sprint o (ADR-0014): intake token admin 발급의 allowed_ips 빈 배열/CIDR 오류
intake_token_collision                         # sprint o (ADR-0014): hashed_token UNIQUE 위반 (사실상 발생 불가)
```

## 9. API ID 인덱스 (sprint `claude/work_260515-f`, intake token admin sprint `o`)

| API ID | endpoint |
| --- | --- |
| API-59 | `POST /api/v1/dev-requests` (외부 수신) |
| API-60 | `GET /api/v1/dev-requests` (목록) |
| API-61 | `GET /api/v1/dev-requests/:id` (상세) |
| API-62 | `POST /api/v1/dev-requests/:id/register` (Promote) |
| API-63 | `POST /api/v1/dev-requests/:id/reject` |
| API-64 | `PATCH /api/v1/dev-requests/:id` (Reassign) |
| API-65 | `DELETE /api/v1/dev-requests/:id` (Close) |
| API-66 | `POST /api/v1/dev-request-tokens` (intake token 발급, sprint `o` / ADR-0014) |
| API-67 | `GET /api/v1/dev-request-tokens` (intake token 목록) |
| API-68 | `DELETE /api/v1/dev-request-tokens/:token_id` (intake token revoke) |
| API-79 | `PATCH /api/v1/dev-request-tokens/:token_id` (intake token IP mutation) |

## 10. Intake Token Admin (API-66..68, sprint `claude/work_260515-o` / ADR-0014)

`dev_request_intake_tokens` resource 의 system_admin 일임 endpoint. plain token 은 발급 응답에 1회만 노출, server 는 SHA-256(plain) hex 만 보관. accounts_admin temp_password 패턴과 정합.

### 10.1 API-66 `POST /api/v1/dev-request-tokens` — 발급

- **인증**: OIDC + RBAC `dev_request_intake_tokens:create` (system_admin only).
- **요청**: `{ "client_label": "ops_portal", "source_system": "ops", "allowed_ips": ["10.0.0.0/24", "192.0.2.7"] }`. 모두 필수. `allowed_ips` 빈 배열 거절 (`invalid_allowed_ips`).
- **처리**: server 가 32-byte base64url plain token 생성 → SHA-256 hex 저장. `created_by` = actor.login.
- **응답 — 201**: `plain_token` 1회 노출 + token_id / client_label / source_system / allowed_ips / created_at / created_by / last_used_at / revoked_at. **`hashed_token` 미노출**.
- **audit**: `dev_request_intake_token.issued` (plain/hashed 모두 미포함).
- **에러**: 400 `invalid_allowed_ips` / 400 missing client_label or source_system / 409 `intake_token_collision`.

### 10.2 API-67 `GET /api/v1/dev-request-tokens` — 목록

- **인증**: OIDC + RBAC `dev_request_intake_tokens:view` (system_admin only).
- **응답 — 200**: `{ "data": [{...}], "meta": {"total": N} }`. revoked 행 포함, `created_at DESC`. **`plain_token` / `hashed_token` 모두 미노출**.

### 10.3 API-68 `DELETE /api/v1/dev-request-tokens/:token_id` — revoke

- **인증**: OIDC + RBAC `dev_request_intake_tokens:delete` (system_admin only).
- **처리**: `revoked_at = COALESCE(revoked_at, NOW())` — idempotent.
- **응답 — 200**: 갱신된 row (plain_token 미포함). **404** `not_found` (token_id 미존재).
- **audit**: `dev_request_intake_token.revoked`.

### 10.4 API-79 `PATCH /api/v1/dev-request-tokens/:token_id` — IP mutation

- **인증**: OIDC + RBAC `dev_request_intake_tokens:edit` (system_admin only).
- **요청**: `{ "allowed_ips": ["10.0.0.1/32"] }`. 필수.
- **처리**: allowed_ips 갱신.
- **응답 — 200**: 갱신된 row.
- **audit**: `dev_request_intake_token.updated`.

## 11. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §14 (전체) 를 도메인 sub-document 로 이관. ID(API-59..68, API-79) 보존, 신규 발급/삭제 없음. |
