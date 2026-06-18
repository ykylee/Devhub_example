---
title: task_requirements
type: source
tags: [domain, task_requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/task_requirements.md]
git_commit: 01f1969c
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T07:11:15Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# integration-registry — Task Item Ingestion 요구사항

- 문서 목적: 외부 ALM/SCM/Issue Tracker 작업 항목(task/issue/ticket) 수집 도메인의 기능·비기능 요구사항을 정의한다.
- 범위: REQ-FR-TASK-001..010 / REQ-NFR-TASK-001..004. 일반 Provider 등록/카탈로그/연결 정책은 `requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §5.10 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./task_ingestion_concept.md), [task_architecture.md](./task_architecture.md), [task_api.md](./task_api.md), [provider requirements](./requirements.md), [master requirements](../../requirements.md)

## 1. 개요

외부 ALM/SCM/Issue Tracker 시스템(Jira, GitHub Issues, GitLab, Linear 등)의 작업 항목(task/issue/ticket)을 DevHub 에서 수집해 통합 조회/추적하는 도메인.

**핵심 설계 결정** ([컨셉 문서](./task_ingestion_concept.md) §9):

- Webhook 인증: Provider 별도 `webhook_secret` 필드 사용 (api_token 과 분리)
- Task status: `raw_status`(원본 string) + `normalized_status`(공통 enum, 초기 NULL 허용) 병행 저장
- Webhook payload 정규화: 시스템별 adapter 가 담당 (Webhook/Pull 공통 경로)
- Webhook 유실 탐지: monotonic sequence(seq) 발급 + Pull 이 gap 탐지 → 보강 sync
- Pull adapter: 범용 REST adapter 우선 구현 (interface 검증 후 시스템별 adapter 확장)

**요구사항 ID 범위**: REQ-FR-TASK-001..010 / REQ-NFR-TASK-001..004

## 2. 기능 요구사항 (REQ-FR-TASK)

### Provider 관리 (기존 `integration_providers` 확장)

- **REQ-FR-TASK-001 (P0, 확정):** 외부 시스템을 `integration_providers` 에 `provider_type = "task_tracker"`로 등록할 수 있어야 한다. 등록 시 `base_url`, `api_token`(Pull 용), `webhook_secret`(Webhook 용) 을 설정할 수 있어야 한다.
- **REQ-FR-TASK-002 (P0, 확정):** Provider 의 `capabilities` 에 `task_item` 플래그를 추가해 해당 Provider 가 Task Item 수집을 지원함을 명시할 수 있어야 한다.

### Webhook 수신 (실시간)

- **REQ-FR-TASK-003 (P0, 확정):** 외부 시스템이 `POST /api/v1/integration/providers/:provider_id/tasks/webhook` 으로 작업 항목 이벤트를 전송할 수 있어야 한다. Webhook payload 는 `event`(created/updated/deleted), `external_id`, `title`, `raw_status`, `priority`, `assignee`, `reporter`, `url`, `labels`, `description` 을 포함한 공통 포맷을 사용한다.
- **REQ-FR-TASK-004 (P0, 확정):** Webhook 수신 시 `X-Webhook-Secret` 헤더 값을 Provider 의 `webhook_secret` 과 대조하여 인증해야 한다. secret 불일치 시 `401 Unauthorized`를 반환한다.
- **REQ-FR-TASK-005 (P0, 확정):** Webhook 수신 성공 시 `external_task_items` 테이블에 upsert(created/updated) 또는 soft-delete(deleted)를 수행하고 `202 Accepted`를 즉시 반환해야 한다. webhook 수신마다 monotonic sequence 번호(`webhook_seq`)를 발급해야 한다.
- **REQ-FR-TASK-006 (P0, 확정):** Webhook 수신 시 원본 payload 전체를 `raw_payload JSONB` 에 보존해야 한다. adapter 가 외부 시스템별 포맷을 공통 포맷으로 정규화하며, 동일한 정규화 경로를 Pull 에서도 재사용해야 한다.

### Pull 동기화 (주기적)

- **REQ-FR-TASK-007 (P1, 확정):** `TaskItemPuller` 인터페이스를 정의하고, Provider 설정의 `pull_interval_seconds`(기본 1800s) 간격으로 외부 시스템 API 를 호출해 작업 항목을 수집하는 Pull loop 를 구현해야 한다.
- **REQ-FR-TASK-008 (P1, 확정):** Pull 동기화는 `last_pulled_at` timestamp 기준 증분 조회를 기본으로 하며, 첫 실행 시 전수 수집(초기 전체 동기화)을 수행해야 한다. 페이지네이션을 지원해야 한다.
- **REQ-FR-TASK-009 (P1, 확정):** Pull loop 는 webhook_seq gap 을 탐지하여 유실된 webhook 이력을 발견할 수 있어야 한다. gap 발견 시 보강(recovery) pull 을 트리거해야 한다.

### 조회 API

- **REQ-FR-TASK-010 (P0, 확정):** `GET /api/v1/external-tasks` 로 수집된 task item 목록을 조회할 수 있어야 한다. `provider_id`, `raw_status`, `normalized_status`, `assignee`, `labels` 필터를 지원해야 하며, `integration_bindings` 를 통한 scope 기반 접근 제어를 적용해야 한다.

## 3. 비기능 / 운영 요구사항 (REQ-NFR-TASK)

- **REQ-NFR-TASK-001 (P0):** 모든 Webhook 수신(성공/실패) 및 Pull 동기화(시작/완료/실패) 이벤트는 audit log 로 기록되어야 한다.
- **REQ-NFR-TASK-002 (P0):** 특정 Provider 의 Webhook/Pull 장애가 다른 Provider 의 수집이나 전체 API 에 영향을 주지 않도록 Provider 단위 장애 격리가 보장되어야 한다.
- **REQ-NFR-TASK-003 (P1):** Pull 동기화 성능 메트릭(Prometheus: `task_item_pull_duration_seconds`, `task_item_pull_total`, `task_item_pull_errors_total`)이 노출되어야 한다.
- **REQ-NFR-TASK-004 (P1):** Webhook secret 은 `integration_providers` 테이블에 저장될 때 write-only(읽기 응답에서 마스킹 또는 미포함) 처리되어야 한다.

## 4. 범위 경계 (Out of Scope)

- DevHub → 외부 시스템 write-back (상태 변경, assign 변경 등). 원천 불변 원칙에 따라 v2 범위.
- 실시간 WebSocket 푸시 (task updated event). 초기엔 polling/refresh 로 충분.
- 외부 시스템 식별자 → DevHub user_id 자동 매핑. 후속 도메인.
- Task item 간 dependency / linkage (epic-link, block-by 등). MVP 이후.
- AI 기반 task 분류/추천. v2 범위.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.10 (Task Ingestion) 본문 그대로 이관. ID(REQ-FR-TASK-001..010, REQ-NFR-TASK-001..004) 보존, 신규 발급/삭제 없음. |
