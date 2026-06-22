---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/audit-ops/api.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# audit-ops 도메인 API

- 문서 목적: `GET /api/v1/audit-logs` (API-18) + `POST /api/v1/internal/keycloak-events` (push SPI 수신) API 계약을 정의한다.
- 범위: API-18 (audit log 조회) + internal Keycloak event push endpoint. 각 도메인의 audit emit 동작 자체는 그 도메인의 api/architecture 에서 정의된다 (cross-domain catalog 는 [architecture.md §4](./architecture.md) 참조).
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/backend_api_contract.md` §9 (audit 부분) 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [ADR-0020 sub-carve E](../../adr/0020-account-user-management-boundary.md), [keycloak_operations.md §8.6](../../setup/keycloak_operations.md)

## 1. `GET /api/v1/audit-logs` (API-18)

command 및 조직/사용자 관리 변경에서 생성된 audit log를 최신순으로 조회한다.

### 1.1 권한

`audit:view` (rbac-permissions §8.2 의 매핑 정합).

### 1.2 Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | `1..100` |
| `offset` | `0` | pagination offset |
| `actor_login` | 없음 | actor login 필터 |
| `action` | 없음 | 예: `user.created`, `org_unit.members_replaced` |
| `target_type` | 없음 | 예: `user`, `org_unit`, `service`, `risk` |
| `target_id` | 없음 | 대상 식별자 |
| `command_id` | 없음 | command 기반 audit log 필터 |

### 1.3 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "audit_id": "audit_1f2a3b4c5d6e",
      "actor_login": "admin",
      "action": "user.created",
      "target_type": "user",
      "target_id": "u3",
      "payload": {
        "actor_source": "x-devhub-actor",
        "role": "developer",
        "status": "active"
      },
      "created_at": "2026-05-07T10:00:00Z"
    }
  ],
  "meta": {
    "limit": 50,
    "offset": 0,
    "count": 1
  }
}
```

조직/사용자 쓰기 API는 audit log 생성에 성공하면 응답 `meta.audit_log_id`를 포함할 수 있다.

## 2. `POST /api/v1/internal/keycloak-events` (push SPI 수신)

Keycloak event listener SPI(Java) 가 이벤트를 push 하는 internal endpoint.

### 2.1 인증

- 일반 OIDC 가 아닌 `X-Webhook-Secret` 상수 비교(fail-closed)로만 인증.
- v1 그룹 미들웨어(인증/RBAC) 밖에 등록 — onboardingGate 도 적용되지 않는다.

### 2.2 본문

Keycloak event JSON 그대로 (event listener SPI 페이로드 스펙 참조).

### 2.3 응답

- 200 OK on accepted (idempotent — 7-tuple SHA-256 이 source_event_id partial UNIQUE 로 dedupe).
- 401 on secret mismatch.

자세한 운영 SOP 는 [keycloak_operations.md §8.6](../../setup/keycloak_operations.md) 참조.

## 3. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §9 (Audit 부분, `GET /api/v1/audit-logs` API-18) 본문 그대로 이관. ID(API-18) 보존, 신규 발급/삭제 없음. internal Keycloak event push endpoint 는 ADR-0020 sub-carve E 의 결과로 본 도메인 API 로 정의. |
