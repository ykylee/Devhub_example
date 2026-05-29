# audit-ops 도메인 아키텍처

- 문서 목적: audit log 발행 파이프라인, Keycloak event listener (SPI push + poll cron), Prometheus 메트릭 아키텍처를 정의한다.
- 범위: master `docs/architecture.md` §6.4 (audit 최소 필드) + §6.5.2 (Keycloak event → audit_logs 동기화) 의 도메인-local 본문 인용 + ADR-0020 sub-carve E 통합.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [api.md](./api.md), [keycloak_operations.md §8.6](../../setup/keycloak_operations.md), [master architecture](../../architecture.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md)

## 1. 컴포넌트 (ARCH-AUDIT-01)

```
                ┌─────────────────────────────────┐
   각 도메인 ──▶│  Audit Emitter (in-process)     │
                │  (mutation handler 가 호출)      │
                └────────────────┬────────────────┘
                                 │
                                 ▼
                ┌─────────────────────────────────┐
                │  audit_logs (Postgres, 000003)  │
                │  + source_event_id (000032,     │
                │    partial UNIQUE)               │
                └─────────────────────────────────┘
                                 ▲
                                 │
                                 │ (event sync)
                                 │
        ┌────────────────────────┴─────────────────────────┐
        │                                                  │
        ▼                                                  ▼
┌─────────────────────────────┐         ┌───────────────────────────────┐
│  Push (SPI, Java)           │         │  Poll (cron, Go internal/audit)│
│  Keycloak event listener    │         │  Admin REST /admin/realms/    │
│  → POST /api/v1/internal/   │         │    {realm}/events + /admin-   │
│      keycloak-events         │         │    events (30s 기본)           │
│  + X-Webhook-Secret 인증    │         │  + event_cursors (000031)     │
└─────────────────────────────┘         └───────────────────────────────┘
                │                                                  │
                └────────────────────┬─────────────────────────────┘
                                     │
                                     ▼
                           ┌────────────────────────┐
                           │  user_sync (sub-carve C)│
                           │  → users.role / status /│
                           │    primary_unit_id 갱신  │
                           └────────────────────────┘
```

## 2. Audit log 최소 필드 (ARCH-AUDIT-02, master §6.4 인용)

Audit Log 는 최소한 `actor_id`, `actor_role`, `action`, `target_type`, `target_id`, `request_id`, `source_ip`, `result`, `reason`, `created_at` 을 기록한다. Webhook 처리 계열 작업은 `actor_id` 대신 `gitea_delivery_id` 또는 dedupe key 를 함께 남겨 재처리 경로를 추적한다.

비밀번호 평문, 해시, 임시 비밀번호는 어떤 audit 필드에도 기록하지 않는다.

## 3. Keycloak event 동기화 (ARCH-AUDIT-03, master §6.5.2 인용)

Keycloak 에서 발생한 사용자/관리자 이벤트(로그인, group/role 변경, 계정 enable/disable, USER:DELETE 등)는 두 경로로 DevHub `audit_logs` + `users` 동기화에 반영된다.

- **Push (SPI)**: Keycloak event listener SPI(Java, `infra/idp/keycloak-event-listener-spi/`)가 이벤트를 `POST /api/v1/internal/keycloak-events` 로 전송한다. 이 endpoint 는 일반 OIDC 가 아닌 `X-Webhook-Secret` 상수 비교(fail-closed)로만 인증하며 v1 그룹 미들웨어(인증/RBAC) 밖에 등록된다([ADR-0020](../../adr/0020-account-user-management-boundary.md) §5.6 push 경로).
- **Poll (cron)**: `internal/audit` 의 Keycloak event 폴러가 Admin REST(`/admin/realms/{realm}/events` + `/admin-events`)를 기본 30s 주기로 polling 해 cursor(`event_cursors`, migration 000031) 이후 이벤트를 audit 으로 emit + `users` profile/membership/status sync(ADR-0020 sub-carve C)한다.
- **dedup**: push 와 poll 이 동시 존재할 수 있으므로(SPI push 단일화는 미전환 부채), distinguishing 7-tuple SHA-256 을 `audit_logs.source_event_id`(`source_type=keycloak_event`, partial UNIQUE migration 000032)에 기록해 at-least-once 중복을 흡수한다.
- audit source_type 카탈로그: `oidc | webhook | keycloak_event | system`(legacy `kratos` enum 은 historical row decode 용으로만 보존, ADR-0001 superseded).

## 4. Audit action 카탈로그 (cross-domain)

본 도메인은 audit action 발행 메커니즘만 소유하고, 각 action 의 정의는 다음 도메인 sub-documents 의 catalog 에 의해 분산 소유된다.

- `account.*` / `auth.*` — [auth-session/api.md §6](../auth-session/api.md), [auth-session/architecture.md](../auth-session/architecture.md)
- `account.onboarding_completed` / `account.review_confirmed` / `account.unit_changed` — [onboarding/architecture.md §6](../onboarding/architecture.md)
- `dev_request.*` / `dev_request_intake_token.*` — [dev-request/architecture.md §6](../dev-request/architecture.md)
- `application.*` / `project.*` / `application_repository.*` / `application.weight_policy_updated` — [application-lifecycle/architecture.md §6](../application-lifecycle/architecture.md)
- `external_task.*` — [integration-registry/task_architecture.md §7](../integration-registry/task_architecture.md)
- `integration.*` / `infra.node.*` / `infra.service.*` — [integration-registry/architecture.md](../integration-registry/architecture.md)
- `user.*` / `org_unit.*` / `organization.hierarchy_updated` — [organization-management/api.md](../organization-management/api.md)
- `rbac.policy.*` / `auth.role_denied` / `auth.row_denied` / `auth.policy_unmapped` — [rbac-permissions/architecture.md](../rbac-permissions/architecture.md)

## 5. Prometheus 메트릭 (ARCH-AUDIT-04)

ADR-0020 sub-carve E + sprint -y (PR #200~#202) 의 결과로 Keycloak event listener 의 운영 가시성을 위해 다음 3종 metric 을 발행한다.

- `keycloak_event_puller_iterations_total{result=success|error}` — poll 주기 실행 카운터.
- `keycloak_event_puller_events_total{kind=user|admin}` — emit 된 event 카테고리 카운터.
- `keycloak_event_puller_lag_seconds` — 마지막 cursor 와 현재 시각의 lag (gauge).

상세 운영 SOP 는 [keycloak_operations.md §8.6](../../setup/keycloak_operations.md) 참조.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §6.4 (audit 최소 필드) + §6.5.2 (Keycloak event listener) 를 도메인 sub-document 로 재집합. ID는 ARCH-AUDIT-01..04 도메인 임시 발급(master 의 audit 전용 ARCH ID 없음). |
