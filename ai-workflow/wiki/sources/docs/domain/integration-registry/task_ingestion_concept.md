---
title: task_ingestion_concept
type: source
tags: [domain, task_ingestion_concept.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/task_ingestion_concept.md]
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

# Task Item Ingestion 도메인 컨셉 — Webhook + Pull 혼합

- 문서 목적: DevHub 에서 외부 시스템(Jira, GitHub Issues, GitLab, Linear 등)의 작업 항목(task/issue/ticket)을 수집해 통합 조회/추적할 수 있는 **Task Item Ingestion** 도메인의 컨셉을 정의한다. Webhook 실시간 수신 + Pull 주기 동기화의 혼합(Hybrid) 접근을 기준으로 삼는다.
- 범위: 컨셉 단계. 도메인 정의, hybrid flow 설계, 기존 External Integration 시스템과의 관계, 데이터 모델 초안, MVP scope, out-of-scope, 미해결 항목, 후속 단계 hook.
- 대상 독자: 프로젝트 리드, Backend/Frontend 담당자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-28
- 결정 근거 sprint: `deepseek/work_260528-a-task-item-ingestion`
- 관련 문서: [external_system_integration_concept.md](./external_system_integration_concept.md), [development_request_concept.md](./development_request_concept.md) (DREQ intake 패턴), [homelab_adapter_pull_strategy.md](./homelab_adapter_pull_strategy.md) (pull adapter 패턴), [release_v1_roadmap.md](./release_v1_roadmap.md), [system_usecases.md](./system_usecases.md), [system_erd.md](./system_erd.md)

---

## 1. 컨셉 정리 배경

- DevHub 는 현재 Application / Repository / Project 3-tier 운영 모델과 DREQ (개발 의뢰) 도메인을 통해 **운영 요청의 수신→검토→등록** 흐름을 확보했다.
- External Integration 시스템을 통해 HomeLab 인프라 상태(Pull)와 DREQ webhook 수신(Intake Token) 패턴을 이미 운영 중이다.
- **다음缺口**: 팀이 실제로 사용하는 ALM/SCM 시스템(Jira, GitHub Issues, GitLab, Linear 등)에서 발생하는 **일상 작업 항목(issues, tickets, tasks)** 을 DevHub 에서 통합 조회/추적할 수단이 없다.
  - 현재는 각 시스템에 개별 접속해야 작업 현황을 알 수 있음
  - 프로젝트/애플리케이션 단위로 연결된 외부 작업 항목을 DevHub 대시보드에서 함께 볼 수 없음
  - 작업 상태 변경 알림을 DevHub audit 로그와 연계할 수 없음

### 1.1 기존 패턴과의 관계

| 패턴 | Task Item Ingestion 과의 관계 |
|------|-------------------------------|
| **DREQ Intake Token** | 유사하나 DREQ 는 "아직 등록되지 않은 작업 의뢰"가 대상. Task Item 은 "이미 외부 시스템에 존재하는 작업 항목"을 미러링. |
| **HomeLab Pull Adapter** | Pull 방식의 adapter 계약/인프라 재사용 가능. 단, HomeLab 은 infra snapshot 전체 교체인 반면, Task Item 은 개별 항목 단위 upsert 필요. |
| **Integration Provider/Binding** | Provider 등록 + scope 연결 모델 재사용. `capabilities` 에 `task_item` 추가로 기존 인프라 확장. |

---

## 2. 도메인 정의

### 2.1 entity: ExternalTaskItem

외부 ALM/SCM/Issue Tracker 시스템에 존재하는 1건의 작업 항목을 DevHub에 미러링한 read-only 복제본.

| 속성 | 의미 |
|------|------|
| **원천 불변** | 외부 시스템이 SoT. DevHub 에서 직접 수정 불가 (status 변경, assign 변경 등 write-action 은 별도 제어 채널). |
| **정규화** | 시스템별 상이한 필드를 DevHub 공통 스키마로 매핑. 원본 필드는 `raw_payload JSONB` 로 보존. |
| **추적성** | task item 의 생성/업데이트/삭제 이벤트는 audit log 로 기록. |

### 2.2 행위자

| 행위자 | 역할 |
|--------|------|
| **system_admin** | Provider 등록/수정/삭제. 전체 task item 조회. |
| **manager / developer** | 자신의 프로젝트/애플리케이션에 binding 된 provider 의 task item 조회. |
| **외부 시스템** | Webhook POST 발신. Pull API 응답. |

---

## 3. Hybrid Flow: Webhook + Pull

### 3.1 전체 흐름

```
                      ┌─────────────────────────┐
                      │   External System        │
                      │  (Jira / GitHub / GitLab)│
                      └──────┬──────────┬───────┘
                             │          │
                 Webhook     │          │  Pull (REST API)
                 (실시간)     │          │  (주기/초기 전체)
                             │          │
                             ▼          ▼
               ┌──────────────────────────────┐
               │       DevHub Server           │
               │                               │
               │  POST /.../webhook            │
               │    → validate provider auth    │
               │    → parse event               │
               │    → upsert task_item          │
               │    → emit audit event          │
               │                               │
               │  Pull Loop (ticker)            │
               │    → fetch external API        │
               │    → diff + upsert batch       │
               │    → emit metrics              │
               │                               │
               │  GET /api/v1/external-tasks    │
               │    → list (scope/provider/tag) │
               └──────────────────────────────┘
```

### 3.2 Webhook (실시간)

외부 시스템의 이벤트(`issues.created`, `issues.updated`, `issues.deleted`)를 DevHub가 수신.

**Endpoint**: `POST /api/v1/integration/providers/:provider_id/tasks/webhook`

**인증 방식**: Provider 별도 `webhook_secret` 필드 사용
- 외부 시스템이 `X-Webhook-Secret` 헤더에 서명을 담아 전송
- DevHub 가 provider 의 `webhook_secret` 과 대조 검증
- provider `api_token` 과 분리되어 webhook/Pull 인증 독립적 운영
- secret 은 admin 이 Provider 등록/수정 시 설정 (PATCH 가능)

**Payload (정규화)**: 외부 시스템별 adapter 가 원본 payload 를 공통 포맷으로 정규화. adapter 는 Webhook 과 Pull 에서 동일한 정규화 경로 사용.

```json
{
  "event": "created" | "updated" | "deleted",
  "external_id": "PRJ-123",
  "title": "Fix login timeout",
  "status": "open",
  "priority": "high",
  "assignee": "user@example.com",
  "reporter": "dev@example.com",
  "url": "https://jira.example.com/browse/PRJ-123",
  "labels": ["bug", "auth"],
  "description": "Users report login timeout after 30s...",
  "occurred_at": "2026-05-28T10:00:00Z"
}
```

**처리**:
1. Provider 조회 + webhook secret 검증
2. `event` 타입에 따라 upsert 또는 soft-delete
3. `raw_payload JSONB` 에 원본 payload 보존
4. audit event emit: `external_task.received` / `external_task.updated` / `external_task.deleted`
5. 응답: `202 Accepted` (즉시 반환, 처리 실패 시 별도 error log)

### 3.3 Pull (전체/증분 동기화)

Pull adapter 가 주기적으로 외부 시스템 API 를 호출해 전체/증분 데이터를 수집.

**Adapter 계약**:

```go
type TaskItemPuller interface {
    // FetchTaskItems pulls task items updated since the given timestamp.
    // Returns mapped ExternalTaskItem slice + optional cursor for pagination.
    FetchTaskItems(ctx context.Context, since time.Time, cursor string) ([]ExternalTaskItem, string, error)
}
```

**Pull Loop**:
1. Provider 의 `last_pulled_at` 기준으로 since 시간 계산
2. 첫 실행 (초기 전체 동기화) 시 since = epoch → 전수 수집
3. 페이지네이션 → 각 page upsert
4. `last_pulled_at` 갱신
5. Prometheus metrics: `task_item_pull_duration_seconds`, `task_item_pull_total`
6. webhook 으로 이미 수신된 항목은 skip (external_id + source_system unique)

**Pull schedule**: Provider 설정 기반
- `cron` 표현식 (예: `*/15 * * * *`)
- 또는 고정 interval (15m, 1h, 6h)
- 기본값: 30분 간격

### 3.4 Hybrid 정합성 보장

| 시나리오 | 처리 |
|---------|------|
| **Webhook 먼저, Pull 이후** | webhook 으로 upsert → Pull 시 `external_id` 로 중복 체크 → skip |
| **Pull 먼저, Webhook 이후** | Pull 로 전체 수집 → webhook 이 delta upsert |
| **Webhook 유실** | webhook_seq gap 탐지 → 보강 pull 트리거. Pull 이 유실분 발견 + upsert (최대 30분 지연) |
| **Webhook 과 Pull 동시 업데이트** | last-mile wins (두 경로 모두 동일 upsert 로직) |

---

## 4. 기존 인프라와의 통합

### 4.1 Integration Provider 확장

기존 `integration_providers` 테이블 활용:

| 필드 | 활용 |
|------|------|
| `provider_key` | 고유 식별자 (예: `jira`, `github-issues`) |
| `provider_type` | 신규 type `"task_tracker"` 추가 (기존: alm/scm/ci_cd/doc/infra) |
| `capabilities` | `["task_item"]` 추가 |
| `base_url` | 외부 시스템 API endpoint |
| `api_token` | Pull API 호출용 인증 |
| `webhook_secret` | **신규 필드** — webhook payload 서명 검증용 |

### 4.2 Integration Binding 확장

기존 `integration_bindings` 활용 — task item 조회 scope 제어:

- Application 에 binding: 해당 application 과 관련된 task item 만 조회
- Project 에 binding: project 소속 application 들의 task item 집합 조회
- Provider 에 직접: 전체 task item 조회 (system_admin 전용)

### 4.3 Task Item Store 인터페이스

```go
type ExternalTaskStore interface {
    UpsertTaskItem(ctx, providerID, externalID string, item *ExternalTaskItem) error
    SoftDeleteTaskItem(ctx, providerID, externalID string) error
    ListTaskItems(ctx, filter TaskItemFilter) ([]ExternalTaskItem, error)
    GetTaskItem(ctx, providerID, externalID string) (*ExternalTaskItem, error)
}
```

---

## 5. 데이터 모델 (1차)

### 5.1 테이블: `external_task_items`

```sql
CREATE TABLE external_task_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id    UUID NOT NULL REFERENCES integration_providers(id) ON DELETE CASCADE,
    external_id    TEXT NOT NULL,          -- 외부 시스템의 ticket/issue key (e.g. "PRJ-123", "1234")
    source_system  TEXT NOT NULL,          -- provider_key 와 동일 (조회 최적화용 denormalization)

    title          TEXT NOT NULL,
    description    TEXT,                   -- markdown 허용
    raw_status     TEXT NOT NULL,          -- 외부 시스템 원본 status (e.g. "Open", "In Progress", "Closed")
    normalized_status TEXT,                -- DevHub 공통 enum 매핑 (e.g. "open", "in_progress", "resolved", "closed"). 초기엔 NULL 허용.
    priority       TEXT,                   -- e.g. "highest", "high", "medium", "low", "lowest"
    assignee       TEXT,                   -- 외부 시스템 assignee 식별자 (email 또는 login)
    reporter       TEXT,                   -- 외부 시스템 reporter 식별자
    url            TEXT,                   -- 외부 시스템 원본 링크
    labels         TEXT[],                 -- 태그/레이블 목록

    raw_payload    JSONB,                  -- 원본 webhook/Pull 응답 전체 보존

    fetched_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- 마지막 수집 시각
    deleted_at     TIMESTAMPTZ,            -- soft delete (webhook deleted event 수신 시)

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    webhook_seq    BIGINT,                 -- webhook 수신 monotonic sequence (gap 탐지용, Pull 미수신분 감지)

    UNIQUE (provider_id, external_id)      -- 중복 방지 + upsert key
    UNIQUE (provider_id, webhook_seq)      -- seq uniqueness (nullable: partial index WHERE webhook_seq IS NOT NULL)
);
```

### 5.2 Provider 확장 필드

기존 `integration_providers` 테이블에 Webhook 전용 필드 추가 (별도 migration):

```sql
ALTER TABLE integration_providers ADD COLUMN IF NOT EXISTS webhook_secret TEXT;
ALTER TABLE integration_providers ADD COLUMN IF NOT EXISTS pull_interval_seconds INTEGER NOT NULL DEFAULT 1800;
ALTER TABLE integration_providers ADD COLUMN IF NOT EXISTS last_pulled_at TIMESTAMPTZ;
```

---

## 6. API 개요 (1차)

| Method | Path | 설명 |
|--------|------|------|
| `POST` | `/api/v1/integration/providers/:provider_id/tasks/webhook` | webhook 수신 (no auth — webhook secret 검증) |
| `POST` | `/api/v1/integration/providers/:provider_id/sync` | 즉시 Pull 트리거 (기존 API-72 확장) |
| `GET` | `/api/v1/external-tasks` | task item 목록 조회 (scope/binding/provider/status 필터) |
| `GET` | `/api/v1/external-tasks/:task_id` | 단건 조회 |

---

## 7. MVP Scope

| Priority | Carve | 설명 |
|----------|-------|------|
| **P0** | Webhook 수신 + upsert | `POST .../tasks/webhook` 수신 → DB upsert + audit |
| **P0** | GET 외부 시스템별 task 목록 | provider_id + status 필터 기반 리스트 |
| **P1** | Pull adapter base + 주기 loop | `TaskItemPuller` 인터페이스 + 기본 loop |
| **P1** | Provider webhook_secret 필드 | migration + store + webhook 검증 |
| **P1** | Provider `capabilities` 에 `task_item` | 기존 Integration Registry 확장 |
| **P2** | 범용 REST Pull adapter 구현 | `TaskItemPuller` interface + REST API 기반 reference adapter (예: response → 공통 포맷 변환) |
| **P2** | Binding 연동 조회 | application/project binding 기반 task filter |
| **P2** | Webhook secret CRUD (admin UI) | Provider 설정 화면에 webhook secret 필드 |
| **P3** | GitHub/GitLab/Jira adapter | 추가 외부 시스템 adapter |
| **P3** | Pull 증분 동기화 최적화 | since timestamp / cursor 기반 |
| **P3** | Task status aggregation | dashboard 위젯 (open/in_progress/closed count) |

---

## 8. Out of Scope (1차)

| 항목 | 이유 |
|------|------|
| **DevHub → 외부 시스템 write** | 원천 불변 원칙. write-action 은 별도 ADR 필요. |
| **실시간 WebSocket 푸시 (task updated event)** | v1.0 에서는 polling/refresh 로 충분 |
| **자동 assignee/user 매핑** | 외부 시스템 식별자 → DevHub user_id 매핑은 후속 |
| **Task item 간 dependency / linkage** | epic-link, block-by 등 고급 관계는 MVP 이후 |
| **AI 기반 task 분류/추천** | v2 영역 |
| **템플릿 기반 task 생성** | write-action 범주. 후속. |

---

## 9. 결정사항 (Resolved)

| # | 질문 | 결정 | 근거 |
|---|------|------|------|
| Q1 | Webhook 인증: Provider api_token 재사용 vs 별도 webhook_secret? | **별도 `webhook_secret` 필드** | webhook/Pull 인증 분리. Provider api_token 과 독립적 운영 |
| Q2 | Pull adapter 구현체 첫 대상? | **범용 REST adapter 우선** | interface + reference implementation 먼저 검증. 특정 시스템 adapter 는 P3 |
| Q3 | Task status: 원본 string vs 공통 enum? | **둘 다: `raw_status` + `normalized_status`** | 원본 보존 + 공통 enum 병행. normalized 초기 NULL 허용 |
| Q4 | Webhook payload 정규화 담당? | **시스템별 adapter** | adapter 가 Webhook/Pull 공통 정규화 경로 책임. 공통 계층에 대한 추가 복잡도 불필요 |
| Q5 | Webhook 유실 탐지 필요? | **SEQ 추적: webhook 수신 monotonic seq** | Pull 이 seq gap 탐지 → 보강 pull 트리거. 정합성 보강 |

---

## 10. 후속 단계 Hook

| 단계 | 산출물 | 예상 sprint |
|------|--------|-------------|
| **REQ** | 요구사항 문서 (`docs/requirements.md` §Task Item 행 추가) | 다음 sprint |
| **UC** | 유스케이스 (`docs/planning/system_usecases.md` UC-TASK-*) | REQ 직후 |
| **ARCH/API** | 아키텍처 결정 + API contract | UC 직후 |
| **IMPL** | Backend handler + store + adapter + migration | 본 문서 검토 후 |
| **IMPL** | Frontend list page + filter | Backend API 완료 후 |
| **TC** | E2E test (webhook ingest → list → verify) | FE/BE 완료 후 |
