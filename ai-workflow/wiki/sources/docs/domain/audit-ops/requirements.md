---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/audit-ops/requirements.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# audit-ops 도메인 요구사항

- 문서 목적: 사용자 및 시스템 변경의 audit log 발행, Keycloak event listener (push SPI + poll cron) 동기화, Prometheus 메트릭 발행의 요구사항을 정의한다.
- 범위: master `docs/requirements.md` 의 §2.5 audit 정책 (`account.*` / `auth.*` event), §5.5 DREQ audit, §5.7 onboarding audit, §6.4 일반 audit 최소 필드 등 분산 요구사항을 도메인 관점에서 재집합. 각 도메인 sub-document 의 audit catalog 가 source-of-truth.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [keycloak_operations.md §8.6](../../setup/keycloak_operations.md), [ADR-0020 sub-carve E](../../adr/0020-account-user-management-boundary.md)

## 1. 개요

본 도메인은 다음 책임을 가진다.

1. **DevHub 자체 변경의 audit emit** — 각 도메인의 mutation handler 가 발행하는 `<domain>.<action>` event 를 `audit_logs` 에 영속화.
2. **Keycloak event 동기화** — Keycloak SPI push + Admin REST poll 양 경로로 사용자/관리자 이벤트(로그인, group/role 변경, 계정 enable/disable 등)를 `audit_logs` + `users` 로 sync.
3. **Audit log 조회 API** — `GET /api/v1/audit-logs` (API-18).
4. **Prometheus metric** — onboarding/audit 관련 운영 메트릭 발행.

## 2. 기능 요구사항

### 2.1 Audit log 영속화

- **REQ-AUDIT-001 (MVP, 확정):** 모든 mutation endpoint 는 audit log 를 emit 한다. Audit log 의 최소 필드는 `actor_id`, `actor_role`, `action`, `target_type`, `target_id`, `request_id`, `source_ip`, `result`, `reason`, `created_at` (master architecture §6.4).
- **REQ-AUDIT-002 (MVP, 확정, append-only invariant):** Audit log 는 사용자 API 로 직접 mutation 하지 않는다 (rbac-permissions §1.4 의 `audit.create/edit/delete=false` 매트릭스 invariant 정합).
- **REQ-AUDIT-003 (MVP, 확정, 비밀 미보관):** 비밀번호 평문/해시, recovery token, 세션 시크릿, access token 원문은 audit payload 에 절대 저장하지 않는다.

### 2.2 Keycloak event 동기화 (ADR-0020 sub-carve E)

- **REQ-AUDIT-004 (MVP, 확정):** Keycloak SPI(Java) 가 이벤트를 `POST /api/v1/internal/keycloak-events` 로 push. endpoint 는 일반 OIDC 가 아닌 `X-Webhook-Secret` 상수 비교(fail-closed) 로만 인증.
- **REQ-AUDIT-005 (MVP, 확정):** 별도 poll cron (`internal/audit`) 이 Admin REST(`/admin/realms/{realm}/events` + `/admin-events`) 를 기본 30s 주기로 polling, cursor(`event_cursors`, migration 000031) 이후 이벤트를 audit + `users` sync.
- **REQ-AUDIT-006 (MVP, 확정, dedup):** push + poll 이 at-least-once 중복을 발생시킬 수 있으므로 distinguishing 7-tuple SHA-256 을 `audit_logs.source_event_id`(`source_type=keycloak_event`, partial UNIQUE migration 000032) 에 기록해 흡수한다.

### 2.3 메트릭

- **REQ-AUDIT-007 (MVP):** Keycloak event listener 의 운영 가시성을 위해 Prometheus 3종 metric 을 발행한다 — 운영 SOP 는 [`keycloak_operations.md §8.6`](../../setup/keycloak_operations.md) 참조.

## 3. 비기능 / 운영 요구사항

- **REQ-NFR-AUDIT-001:** Audit log 보존 기간은 master `docs/requirements.md` §4.1 의 표 — 운영 로그 1개월 + 보안/운영 이벤트는 즉시 알림.
- **REQ-NFR-AUDIT-002:** Audit payload 는 mutation 의 before/after diff 를 포함하되 secret/credential 은 마스킹.

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` 분산 audit 요구사항(§2.5/§4.1/§5.5/§5.7) + master `docs/architecture.md` §6.4 (audit 최소 필드) 를 도메인 sub-document 로 재집합. ID는 REQ-AUDIT-001..007 + REQ-NFR-AUDIT-001/002 도메인 임시 발급(traceability matrix 와 별도 — Phase 4 재구성 시 정합). |
