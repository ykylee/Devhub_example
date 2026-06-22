---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/dev-request/architecture.md]
git_commit: 71c0d2cd
git_branch: chore/260622-wiki-drift-cleanup
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:47:55Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# dev-request 도메인 아키텍처 (DREQ)

- 문서 목적: DREQ 도메인의 컴포넌트·상태머신·인증경계·RBAC·데이터 모델·audit catalog 를 정의한다.
- 범위: ARCH-DREQ-01..06. cross-cutting 3대 레이어 / 호출 규칙 / OIDC + RBAC 일반 정책은 master `docs/architecture.md` §1–§4, §6 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/architecture.md` §7 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./concept.md), [requirements.md](./requirements.md), [api.md](./api.md), [master architecture](../../architecture.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0012](../../adr/0012-dreq-external-intake-auth.md)

## 개요

외부 시스템에서 들어오는 개발 의뢰를 수신 → 담당자 검토 → application/project 등록(promote) 까지 처리하는 도메인. 컨셉 문서: [`./concept.md`](./concept.md). 요구사항: [`./requirements.md`](./requirements.md). Usecase: [`UC-DREQ-01..10`](../../planning/system_usecases.md).

## 1. 컴포넌트 (ARCH-DREQ-01)

```
┌──────────────────┐                       ┌──────────────────────────────────────┐
│  External System │ ──── POST /api/v1 ─▶  │  Go Core: dev_requests handler       │
│ (ops portal /    │   /dev-requests       │  ├── auth: 외부 수신용 별도 정책      │
│  ITSM / Jira /   │                       │  │   (REQ-NFR-DREQ-001, ADR 후보)     │
│  사내 워크플로우)│                       │  ├── validate: 필수 필드 + assignee   │
└──────────────────┘                       │  │   존재 / (source_system,           │
                                           │  │   external_ref) idempotency        │
                                           │  ├── store: dev_requests (Postgres)   │
                                           │  └── audit: dev_request.received      │
                                           └────────────┬─────────────────────────┘
                                                        │
                                                        ▼
                                           ┌──────────────────────────────────────┐
                                           │  Frontend: 담당자 dashboard          │
                                           │  + /admin/settings/dev-requests       │
                                           │  └── Promote-to-Platform/Project  │
                                           │     (단일 트랜잭션 — REQ-FR-DREQ-005) │
                                           └──────────────────────────────────────┘
                                                        │
                                                        ▼
                                           ┌──────────────────────────────────────┐
                                           │  Application / Project 도메인        │
                                           │  (DREQ.registered_target_id 로 매핑)  │
                                           └──────────────────────────────────────┘
```

## 2. 상태 머신 (ARCH-DREQ-02)

[컨셉 §2.3](./concept.md) 의 6-상태 머신 (`received → pending → in_review → registered | rejected | closed`). 모든 전이는 `dev_request.*` audit action 으로 기록.

## 3. 외부 수신 인증 경계 (ARCH-DREQ-03)

- 외부 수신 endpoint (`POST /api/v1/dev-requests`) 는 일반 사용자 OIDC 흐름이 아닌 **별도 인증 middleware (`requireIntakeToken`)** 를 사용. **[ADR-0012](../../adr/0012-dreq-external-intake-auth.md)** 가 옵션 A (API 토큰 + IP allowlist) 를 채택. 옵션 B (HMAC) / C (OAuth client_credentials) 는 후속 단계 마이그레이션 경로.
- 검증 흐름 (ADR-0012 §4.1.2):
  - 외부 호출은 `Authorization: Bearer <plain-token>` 헤더로 도착.
  - middleware 가 `SHA-256(plain-token)` 으로 `dev_request_intake_tokens.hashed_token` lookup.
  - 매칭 없음 또는 `revoked_at IS NOT NULL` → 401.
  - caller IP 가 row 의 `allowed_ips` CIDR 범위 밖 → 401.
  - 검증 성공 시 `source_system` 컨텍스트 주입 + `last_used_at` 갱신 + audit `dev_request.intake_auth_succeeded` emit.
- 본 endpoint 는 `routePermissionTable` 의 `Bypass: true` 또는 별도 `IntakeAuth: true` 플래그로 일반 OIDC enforce 를 건너뛴다.
- 인증 성공 시 `source_system` 은 토큰의 매핑 값에서 자동 채움 (request body 의 self-claim 은 신뢰하지 않음 — spoofing 방지).
- 그 외 endpoint (GET 목록 / 상세 / Promote / Reject / Reassign / Close) 는 일반 OIDC + RBAC + 본 sprint 의 `enforceRowOwnership` 패턴([ADR-0011 §4.2](../../adr/0011-rbac-row-scoping.md))으로 보호. 담당자 본인 의뢰 또는 system_admin / team_manager 만 가능.

## 4. RBAC 자원 (ARCH-DREQ-04)

- 신규 resource `dev_requests` 를 RBAC matrix 에 추가.
- 1차 정책 (MVP):
  - `system_admin`: view + create(외부 수신 server-side, frontend 에서는 미노출) + edit + delete
  - `team_manager`: view + edit (담당자 재할당은 제외 — system_admin 만)
  - `manager` / `developer`: view (본인 의뢰만, row-level `actor.login == assignee_user_id`)
- 정책 매핑 표는 backend 구현 sprint 의 migration (`000022_dev_requests` 또는 `000023_rbac_dev_request_resource`) 에서 확정.

## 5. 데이터 모델 (ARCH-DREQ-05)

```text
dev_requests
  id                      uuid       PK
  title                   text       NOT NULL
  details                 text
  requester               text       NOT NULL
  assignee_user_id        text       NOT NULL  REFERENCES users(user_id) ON DELETE RESTRICT
  source_system           text       NOT NULL
  external_ref            text       NULLABLE  -- (source_system, external_ref) UNIQUE
  status                  text       NOT NULL  CHECK in (received, pending, in_review, registered, rejected, closed)
  registered_target_type  text                 CHECK in (application, project) WHEN status='registered'
  registered_target_id    text                 NULLABLE
  rejected_reason         text                 NOT NULL WHEN status='rejected'
  received_at             timestamptz NOT NULL
  created_at, updated_at  timestamptz NOT NULL DEFAULT NOW()

  CONSTRAINT dev_requests_idempotency_uniq
    UNIQUE (source_system, external_ref)
    WHERE external_ref IS NOT NULL;
  CONSTRAINT dev_requests_registered_target_consistency
    CHECK ( (status = 'registered') = (registered_target_type IS NOT NULL AND registered_target_id IS NOT NULL) );
  CONSTRAINT dev_requests_rejected_reason_required
    CHECK ( (status = 'rejected') = (rejected_reason IS NOT NULL) );
```

application / project 의 `origin_dreq_id` 역참조 컬럼 도입 여부는 REQ-FR-DREQ-009 의 ADR 후속에서 결정.

### 외부 수신 토큰 테이블 (ADR-0012 §4.1.1)

```text
dev_request_intake_tokens
  token_id        uuid       PK
  client_label    text       NOT NULL  -- 운영용 식별자 (예: "ops_portal")
  hashed_token    text       NOT NULL  UNIQUE  -- SHA-256 hex of plain token
  allowed_ips     jsonb      NOT NULL  -- CIDR 배열
  source_system   text       NOT NULL  -- token 매핑되는 source_system 값
  created_at      timestamptz NOT NULL DEFAULT NOW()
  created_by      text       NOT NULL  REFERENCES users(user_id)
  last_used_at    timestamptz NULLABLE
  revoked_at      timestamptz NULLABLE
```

plain token 은 발급 직후 1회만 admin 에게 노출하고 어디에도 저장하지 않는다 (IdP admin password issuance 패턴, [accounts_admin](../../backend/) 참조).

## 6. Audit action 카탈로그 (ARCH-DREQ-06)

| action | target_type | 비고 |
| --- | --- | --- |
| `dev_request.received` | `dev_request` | 외부 수신, payload 에 source_system / external_ref / assignee |
| `dev_request.registered` | `dev_request` | promote 시점, payload 에 registered_target_type/id |
| `dev_request.rejected` | `dev_request` | rejected_reason 포함 |
| `dev_request.reassigned` | `dev_request` | from / to assignee |
| `dev_request.reopened` | `dev_request` | rejected → pending |
| `dev_request.closed` | `dev_request` | registered/rejected → closed |
| `dev_request.intake_auth_succeeded` | `dev_request_intake_token` | ADR-0012 §4.1.6 — payload `{token_id, client_label, source_ip}`. token plain 값은 절대 기록 안 함. |
| `dev_request.intake_auth_failed` | `dev_request_intake_token` 또는 `route` | ADR-0012 §4.1.6 — payload `{reason, source_ip, header_present, token_prefix_4chars}`. token full 값은 절대 기록 안 함. |
| `auth.row_denied` | `route` | enforceRowOwnership 패턴, 본 도메인 row 거절 |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §7 (DREQ 도메인 본문) 을 도메인 sub-document 로 이관. ID(ARCH-DREQ-01..06) 보존, 신규 발급/삭제 없음. |
