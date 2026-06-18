---
title: task_api
type: source
tags: [domain, task_api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/task_api.md]
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

# integration-registry — Task Item Ingestion API

- 문서 목적: 외부 ALM/SCM/Issue Tracker 작업 항목 수집 도메인의 API 계약을 정의한다.
- 범위: API-94..96. 일반 Integration Provider API 는 `api.md` (API-69..78, API-80, API-87..90) 참조. envelope/공통 enum 은 master `docs/backend_api_contract.md` §1–§2 또는 `docs/api/conventions.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §17 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./task_ingestion_concept.md), [task_requirements.md](./task_requirements.md), [task_architecture.md](./task_architecture.md), [provider api](./api.md), [master API](../../backend_api_contract.md)

## 개요

> 외부 ALM/SCM/Issue Tracker 시스템(Jira, GitHub Issues, GitLab, Linear)의 작업 항목을 Webhook(실시간) + Pull(주기 동기화) 혼합 방식으로 수집한다. 기존 `integration_providers`/`integration_bindings` 인프라를 확장한다. 아키텍처: [`./task_architecture.md`](./task_architecture.md), 컨셉: [`./task_ingestion_concept.md`](./task_ingestion_concept.md).

**API ID 범위**: API-94 ~ API-96

## 1. Integration Provider 확장 (task_tracker type)

기존 API-70/API-71(POST/PATCH `/api/v1/integration/providers`) 에 `provider_type = "task_tracker"` 지원을 추가한다. 아래 필드는 task_tracker type 일 때만 유효하다.

| 필드 | 타입 | 필수 | 설명 |
| --- | --- | --- | --- |
| `provider_type` | string | yes | `"task_tracker"` |
| `capabilities` | string[] | yes | `["task_item"]` 포함 |
| `webhook_secret` | string | no | Webhook 서명 검증용 secret. **write-only** — 응답 시 `webhook_secret_set: bool` 만 반환. api_token 과 동일 패턴. |
| `pull_interval_seconds` | int | no | Pull loop 주기 (기본 1800). |

**API-70 요청 예시** (task_tracker provider 생성):

```json
{
  "provider_key": "jira-dev",
  "provider_type": "task_tracker",
  "name": "Jira (Dev Team)",
  "capabilities": ["task_item"],
  "base_url": "https://jira.example.com",
  "api_token": "jt-xxxxxxxx",
  "webhook_secret": "whsec_yyyyyyyy",
  "pull_interval_seconds": 900
}
```

**API-70 응답** (webhook_secret write-only):

```json
{
  "status": "ok",
  "data": {
    "id": "uuid-...",
    "provider_key": "jira-dev",
    "provider_type": "task_tracker",
    "capabilities": ["task_item"],
    "base_url": "https://jira.example.com",
    "api_token_set": true,
    "webhook_secret_set": true,
    "pull_interval_seconds": 900
  }
}
```

**API-69** (list) / **API-80** (delete) 는 provider_type 무관 동일하게 동작한다.

## 2. 외부 Task Webhook 수신  *(API-94)*

- **endpoint**: `POST /api/v1/integration/providers/:provider_id/tasks/webhook`
- **인증**: Provider `webhook_secret` 과 `X-Webhook-Secret` 헤더 대조. Provider 가 `provider_type = "task_tracker"` 가 아니면 404.
- **멱등성**: `(provider_id, external_id)` UNIQUE 기준 upsert. 동일 external_id 의 중복 webhook 은 202 + 무시.
- **응답**: `202 Accepted` (즉시 반환. 처리 실패 시 audit + error log 는 비동기).

**요청 — 공통 포맷** (adapter 정규화 이후):

```json
{
  "event": "created",
  "external_id": "PRJ-123",
  "title": "Fix login timeout",
  "raw_status": "Open",
  "priority": "High",
  "assignee": "user@example.com",
  "reporter": "dev@example.com",
  "url": "https://jira.example.com/browse/PRJ-123",
  "labels": ["bug", "auth"],
  "description": "Users report login timeout after 30 seconds of inactivity."
}
```

| 필드 | 타입 | 필수 | 설명 |
| --- | --- | --- | --- |
| `event` | string | yes | `"created"` / `"updated"` / `"deleted"` |
| `external_id` | string | yes | 외부 시스템의 ticket/issue key |
| `title` | string | yes | 작업 제목 (200자 이내) |
| `raw_status` | string | yes | 외부 시스템 원본 status |
| `normalized_status` | string | no | DevHub 공통 enum (adapter 가 매핑 시) |
| `priority` | string | no | 우선순위 |
| `assignee` | string | no | 담당자 식별자 |
| `reporter` | string | no | 보고자 식별자 |
| `url` | string | no | 원본 링크 |
| `labels` | string[] | no | 태그 목록 |
| `description` | string | no | 상세 내용 (markdown) |

**응답 — 202**:

```json
{
  "status": "accepted",
  "data": {
    "webhook_seq": 1042,
    "external_id": "PRJ-123",
    "event": "created",
    "provider_id": "uuid-..."
  }
}
```

**에러**:

| code | HTTP | 설명 |
| --- | --- | --- |
| `provider_not_found` | 404 | provider_id 없음 |
| `provider_type_mismatch` | 404 | provider_type ≠ task_tracker |
| `webhook_secret_mismatch` | 401 | X-Webhook-Secret 불일치 |
| `invalid_webhook_payload` | 422 | event/external_id/title 누락 또는 형식 위반 |

**Audit**: `external_task.received` (created) / `external_task.updated` (updated) / `external_task.deleted` (deleted — soft-delete).

## 3. Task Item 목록 조회  *(API-95)*

- **endpoint**: `GET /api/v1/external-tasks`
- **인증**: OIDC + RBAC + scope binding (ARCH-TASK-06).

**Query parameters**:

| param | 타입 | 필수 | 설명 |
| --- | --- | --- | --- |
| `provider_id` | UUID | no | 특정 Provider 의 task 만 조회 |
| `raw_status` | string | no | 원본 status 필터 (예: "Open") |
| `normalized_status` | string | no | 공통 enum 필터 (예: "open") |
| `assignee` | string | no | 담당자 식별자 필터 |
| `labels` | string (comma-sep) | no | 레이블 OR 필터 (하나라도 포함) |
| `include_deleted` | boolean | no | true 시 soft-deleted 포함 (기본 false) |
| `page` | int | no | 페이지 번호 (기본 1) |
| `per_page` | int | no | 페이지 크기 (기본 20, 최대 100) |

**응답 — 200**:

```json
{
  "status": "ok",
  "data": [
    {
      "id": "uuid-...",
      "provider_id": "uuid-...",
      "external_id": "PRJ-123",
      "title": "Fix login timeout",
      "raw_status": "Open",
      "normalized_status": "open",
      "priority": "High",
      "assignee": "user@example.com",
      "reporter": "dev@example.com",
      "url": "https://jira.example.com/browse/PRJ-123",
      "labels": ["bug", "auth"],
      "fetched_at": "2026-05-28T10:00:00Z",
      "webhook_seq": 1042
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 42
  }
}
```

## 4. Task Item 단건 조회  *(API-96)*

- **endpoint**: `GET /api/v1/external-tasks/:task_id`
- **인증**: OIDC + RBAC + scope binding.

**응답 — 200**: 위 §3 단일 item shape + `raw_payload` (JSONB) 포함.

**에러**:

| code | HTTP | 설명 |
| --- | --- | --- |
| `external_task_not_found` | 404 | task_id 없음 |
| `external_task_forbidden` | 403 | scope 밖 접근 |

## 5. 공통 에러 코드 (Task Item 신규)

```
webhook_secret_mismatch      # 401, X-Webhook-Secret 불일치
provider_type_mismatch       # 404, provider_type ≠ task_tracker
invalid_webhook_payload      # 422, webhook 필수 필드 누락
external_task_not_found      # 404, task_id 미존재
external_task_forbidden      # 403, scope 밖 접근
```

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §17 (Task Item Ingestion) 본문 그대로 이관. ID(API-94..96) 보존, 신규 발급/삭제 없음. |
