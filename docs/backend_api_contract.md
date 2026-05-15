# Backend API Contract

- 문서 목적: 프론트엔드와 백엔드 사이의 API 계약 (endpoint, request/response schema, envelope, enum) 을 단일 source-of-truth 로 기록한다.
- 범위: REST endpoint (/api/v1/**) + WebSocket envelope + 공통 enum + RBAC 정책 API (§12). 도메인 결정 근거는 `docs/architecture.md` + `docs/adr/`, 운영 시드는 `docs/setup/test-server-deployment.md`.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent, 외부 API consumer, QA.
- 상태: accepted
- 기준일: 2026-05-04
- 최종 수정일: 2026-05-15 (외부 시스템 연동 API 초안 §15 추가, API-69..78)
- 관련 문서: [아키텍처](./architecture.md), [기술 스택](./tech_stack.md), [프론트 연동 요구사항](./backend/frontend_integration_requirements.md), [백엔드 요구사항 리뷰](./backend/requirements_review.md), [ADR-0002 RBAC](./adr/0002-rbac-policy-edit-api.md), [백엔드 로드맵](../ai-workflow/memory/backend_development_roadmap.md), [추적성 매트릭스](./traceability/report.md).

## 1. 공통 응답 원칙

- 성공 응답은 `status`, `data`, `meta`를 기본 envelope로 사용한다.
- 단일 command성 endpoint는 `status`와 생성/처리 결과 key를 함께 반환할 수 있다.
- 실패 응답은 `status`, `error`를 반환한다.
- 시간 값은 ISO 8601/RFC3339 형식의 UTC timestamp를 사용한다.
- API role wire format은 `developer`, `manager`, `system_admin`을 사용하고 UI 표시명과 분리한다.

## 2. 공통 enum 및 상태 값

### Role wire format

```text
developer
manager
system_admin
```

### 공통 상태 값

```text
ServiceStatus = stable | warning | degraded | down
RiskImpact = low | medium | high | critical
RiskStatus = detected | investigation | action_required | mitigated | dismissed
CommandStatus = pending | running | succeeded | failed | rejected | cancelled
WebhookEventStatus = received | validated | processed | failed | ignored
AccountStatus = active | disabled | locked | password_reset_required
```

Webhook event는 signature 검증과 raw 저장이 끝나면 `validated`가 되며, 정규화가 성공하면 `processed`, 재처리 가능한 오류는 `failed`, 지원하지 않거나 처리 대상이 아닌 이벤트는 `ignored`로 전환한다.

## 3. Health

### `GET /health` (API-01)

Go Core 상태를 확인한다.

#### 응답 예시

```json
{
  "status": "ok",
  "service": "backend-core",
  "db": "ok"
}
```

`DB_URL`이 설정되지 않은 로컬 실행에서는 `db`가 `disabled`일 수 있다.

## 4. Gitea Webhook 수신

### `POST /api/v1/integrations/gitea/webhooks` (API-02)

Gitea Webhook payload를 수신해 signature를 검증하고 raw event로 저장한다.

#### 필수 Header

- `X-Gitea-Signature`: `GITEA_WEBHOOK_SECRET` 기반 HMAC-SHA256 signature. `sha256=` prefix 유무를 모두 허용한다.
- `X-Gitea-Event`: Gitea event type.

#### 선택 Header

- `X-Gitea-Delivery`: delivery id. 있으면 dedupe key로 사용한다.

#### 성공 응답

```json
{
  "status": "accepted",
  "event_id": 1,
  "event_type": "pull_request",
  "duplicate": false
}
```

#### 중복 응답

중복 이벤트는 저장하지 않고 성공 계열 응답으로 처리한다.

```json
{
  "status": "duplicate"
}
```

## 5. Webhook Event 조회

### `GET /api/v1/events` (API-04)

저장된 raw webhook event 목록을 최신순으로 조회한다. 프론트엔드 초기 개발에서는 이 endpoint를 이벤트 피드와 연동 상태 확인에 사용할 수 있다.

#### Query

| 이름 | 기본값 | 범위 | 설명 |
| --- | --- | --- | --- |
| `limit` | `50` | `1..100` | 한 번에 조회할 이벤트 수 |
| `offset` | `0` | `0..100000` | pagination offset |

#### 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": 7,
      "event_type": "push",
      "delivery_id": "delivery-7",
      "dedupe_key": "delivery-7",
      "repository_id": 42,
      "repository_name": "acme/api",
      "sender_login": "yklee",
      "payload": {
        "ref": "refs/heads/main"
      },
      "status": "validated",
      "received_at": "2026-05-02T10:00:00Z",
      "validated_at": "2026-05-02T10:00:00Z"
    }
  ],
  "meta": {
    "limit": 50,
    "offset": 0,
    "count": 1
  }
}
```

## 6. 프론트 Snapshot API 1차

프론트 mock service 교체를 위한 snapshot API다. 응답 shape는 유지하고, backing data source는 `SnapshotProvider` 경계 뒤에 둔다. 기본 구성은 runtime provider가 infra 상태를 health check로 보강하고, 나머지 snapshot은 static fallback provider에 위임한다.

### `GET /api/v1/rbac/policy` (legacy, deprecated)

> **Deprecated (M1 PR-G1)** — ADR-0002 채택으로 RBAC 모델이 *per-resource 4-boolean* 으로 통일됐다. 본 endpoint 의 1차원 (`none|read|write|admin`) 응답은 호환성 유지용으로만 남기며, 신규 통합은 §12.1 `GET /api/v1/rbac/policies` (복수형) 를 사용한다. M1 PR-G4 머지와 함께 본 endpoint 는 410 Gone 으로 회수될 예정이다.

프론트 Organization > Permissions 화면이 사용할 *legacy* RBAC policy를 조회한다. 응답은 ADR-0002 이전의 1차원 모델을 보존한다.

#### 응답 예시

```json
{
  "status": "ok",
  "data": {
    "roles": [
      {
        "role": "developer",
        "label": "Developer",
        "description": "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한"
      }
    ],
    "resources": [
      {
        "resource": "repositories",
        "label": "Repositories",
        "description": "repository, issue, pull request metadata"
      }
    ],
    "permissions": [
      {
        "permission": "read",
        "label": "Read",
        "rank": 10,
        "description": "조회 가능"
      }
    ],
    "matrix": {
      "developer": {
        "repositories": "read",
        "commands": "none"
      },
      "manager": {
        "repositories": "write",
        "commands": "write"
      },
      "system_admin": {
        "repositories": "admin",
        "commands": "admin"
      }
    }
  },
  "meta": {
    "policy_version": "2026-05-07.default",
    "source": "static_default_policy",
    "editable": false
  }
}
```

### `GET /api/v1/dashboard/metrics` (API-05)

역할별 KPI metric 목록을 조회한다.

#### Query

| 이름 | 기본값 | 범위 | 설명 |
| --- | --- | --- | --- |
| `role` | `developer` | `developer`, `manager`, `system_admin` | 조회할 역할 |

#### 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": "build_success",
      "label": "Build Success",
      "value": "98%",
      "trend": "+2%",
      "trend_direction": "up",
      "numeric_value": 98,
      "unit": "percent"
    }
  ],
  "meta": {
    "role": "developer",
    "count": 3,
    "source": "static"
  }
}
```

### `GET /api/v1/infra/nodes` (API-06)

인프라 topology node 목록을 조회한다. CPU, memory, duration 계열 값은 프론트가 표시 문자열로 포맷팅할 수 있도록 원시 값을 우선 제공한다.
응답 `meta.source`는 snapshot provider 출처를 나타낸다.
runtime provider는 `DB_URL`, `GITEA_URL`, `BACKEND_AI_URL` 설정을 기준으로 `postgres`, `gitea`, `backend-ai` node 상태를 `stable`, `warning`, `down` 중 하나로 갱신한다.

### `GET /api/v1/infra/edges` (API-07)

인프라 topology edge 목록을 조회한다.
응답 `meta.source`는 snapshot provider 출처를 나타낸다.

### `GET /api/v1/infra/topology` (API-07, composite)

인프라 node와 edge를 한 번에 조회한다.

#### 응답 예시

```json
{
  "status": "ok",
  "data": {
    "nodes": [
      {
        "id": "backend-core",
        "label": "Go Core Service",
        "kind": "service",
        "status": "stable",
        "region": "asia-01",
        "cpu_percent": 12.4,
        "memory_bytes": 1288490189,
        "active_instances": 1,
        "updated_at": "2026-05-02T10:00:00Z"
      }
    ],
    "edges": [
      {
        "id": "gitea-backend-core",
        "source_id": "gitea",
        "target_id": "backend-core",
        "label": "WEBHOOK",
        "status": "stable",
        "latency_ms": 28.5,
        "throughput_rps": 2.4,
        "updated_at": "2026-05-02T10:00:00Z"
      }
    ]
  },
  "meta": {
    "node_count": 4,
    "edge_count": 3,
    "source": "static"
  }
}
```

## 7. 도메인 조회 API 1차

도메인 정규화 테이블 기반 조회 API다. 공통 query는 `limit`, `offset`, `repository_name`을 사용하며, 목록 응답은 `status`, `data`, `meta` envelope를 따른다.

### `GET /api/v1/repositories` (API-08)

정규화된 Gitea repository 목록을 조회한다.

#### Query

| 이름 | 기본값 | 범위 | 설명 |
| --- | --- | --- | --- |
| `limit` | `50` | `1..100` | 조회할 항목 수 |
| `offset` | `0` | `0..100000` | pagination offset |

#### 응답 필드

- `id`
- `gitea_repository_id`
- `full_name`
- `owner_login`
- `name`
- `clone_url`
- `html_url`
- `default_branch`
- `private`
- `updated_at`

### `GET /api/v1/issues` (API-09)

정규화된 issue 목록을 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | 조회할 항목 수 |
| `offset` | `0` | pagination offset |
| `repository_name` | 없음 | 특정 repository full name으로 필터링 |
| `state` | 없음 | `open`, `closed` 필터링 |

### `GET /api/v1/pull-requests` (API-10)

정규화된 pull request 목록을 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | 조회할 항목 수 |
| `offset` | `0` | pagination offset |
| `repository_name` | 없음 | 특정 repository full name으로 필터링 |
| `state` | 없음 | `open`, `closed`, `merged` 필터링 |

### `GET /api/v1/ci-runs` (API-11)

CI run snapshot 목록을 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `scope` | `all` | 초기 구현에서는 응답 필터링 없이 meta에만 반영한다. 후속 구현에서 `mine` 등으로 필터링한다. |
| `limit` | `50` | DB-backed 조회 시 조회할 항목 수 |
| `offset` | `0` | DB-backed 조회 시 pagination offset |
| `repository_name` | 없음 | DB-backed 조회 시 repository full name으로 필터링 |
| `status` | 없음 | DB-backed 조회 시 CI status로 필터링 |

`DB_URL`이 설정되고 정규화된 CI run 데이터가 있으면 PostgreSQL 기반 응답을 우선 사용한다. DB 데이터가 없거나 `DomainStore`가 설정되지 않은 경우 static fallback snapshot을 반환할 수 있다. 응답 `meta.source`는 `db` 또는 `static`이다.

### `GET /api/v1/ci-runs/{ci_run_id}/logs` (API-12)

CI run 로그 라인을 조회한다.

### `GET /api/v1/risks` (API-13)

정규화/분석 결과로 저장된 risk 목록을 조회한다. CI 실패 이벤트는 1차 구현에서 `ci_failure:{ci_run_id}` risk key로 `action_required` risk를 생성할 수 있다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | 조회할 항목 수 |
| `offset` | `0` | pagination offset |
| `status` | 없음 | `detected`, `investigation`, `action_required`, `mitigated`, `dismissed` 필터링 |
| `impact` | 없음 | `low`, `medium`, `high`, `critical` 필터링 |

#### 응답 필드

- `id`: risk key. 예: `ci_failure:502`
- `title`
- `reason`
- `impact`
- `status`
- `owner_login`
- `suggested_actions`
- `created_at`
- `updated_at`

### `GET /api/v1/risks/critical` (API-13, critical filter)

Manager dashboard의 critical risk 목록을 조회한다.
`DB_URL`이 설정되고 `action_required` + `high` risk 데이터가 있으면 PostgreSQL 기반 응답을 우선 사용한다. DB 데이터가 없거나 `DomainStore`가 설정되지 않은 경우 snapshot provider fallback을 반환할 수 있다. 응답 `meta.source`는 `db`, `runtime`, `static` 중 하나다.

#### 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": "risk-001",
      "title": "Gitea Migration Blocked",
      "reason": "Access token expiration and scope mismatch detected in logs.",
      "impact": "high",
      "status": "action_required",
      "owner_login": "alex.k",
      "suggested_actions": ["rotate_token", "verify_scopes"],
      "created_at": "2026-05-01T10:00:00Z",
      "updated_at": "2026-05-02T10:00:00Z"
    }
  ],
  "meta": {
    "count": 2,
    "source": "db"
  }
}
```

## 8. Realtime WebSocket 계약

### `GET /api/v1/realtime/ws` (API-14)

REST snapshot 조회 이후 변경 이벤트를 수신하는 WebSocket endpoint다. 브라우저 프론트엔드는 gRPC stream에 직접 연결하지 않는다.

RBAC enabled 환경에서는 `types` query로 필요한 event type을 명시한다. `actor`/`role` query 는 `DEVHUB_AUTH_DEV_FALLBACK=true`인 개발 환경에서만 허용한다. 운영 환경에서는 Bearer token 또는 session 기반 actor context를 사용해야 한다 ([ADR-0004](./adr/0004-x-devhub-actor-removal.md) 로 `X-Devhub-Actor` fallback 은 폐기됐다).

#### 메시지 envelope

```json
{
  "schema_version": "1",
  "type": "ci.run.updated",
  "event_id": "evt_20260502100000.000000000",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {
    "id": "101",
    "repository_name": "devhub-core",
    "status": "success"
  }
}
```

#### 초기 event type

```text
infra.node.updated
infra.edge.updated
ci.run.updated
ci.log.appended
risk.critical.created
risk.updated
command.status.updated
notification.created
```

프론트는 알 수 없는 `type`을 무시하고, 지원하지 않는 `schema_version`은 사용자 화면을 깨뜨리지 않는 방식으로 로깅한다.

#### 구현 상태

- 2026-05-06 기준 `/api/v1/realtime/ws` endpoint와 in-process WebSocket hub가 구현되어 있다.
- 현재 publish 대상은 dry-run command worker가 발생시키는 `command.status.updated`다.
- 인증, 구독 필터, 마지막 event replay는 아직 후속 범위다.

#### `command.status.updated` data 예시

```json
{
  "command_id": "cmd_1f2a3b4c5d6e",
  "command_type": "service_action",
  "target_type": "service",
  "target_id": "runner-asia-01",
  "action_type": "restart",
  "status": "succeeded",
  "actor_login": "yklee",
  "result_payload": {
    "executor": "dry_run",
    "message": "Dry-run command accepted without external side effects."
  },
  "updated_at": "2026-05-06T10:00:00Z"
}
```

## 9. Command/Audit 계약 초안

서비스 제어와 리스크 완화 같은 명령성 액션은 즉시 boolean 성공으로 처리하지 않는다. 백엔드는 command를 생성하고 `202 Accepted`로 `command_id`, `command_status`, `audit_log_id`를 반환한다. 실행 결과는 `GET /api/v1/commands/{command_id}` 또는 `command.status.updated` WebSocket event로 추적한다.

### `POST /api/v1/admin/service-actions` (API-15)

System Admin dashboard의 서비스 제어 요청을 command lifecycle로 생성한다. `dry_run` 기본값은 `true`이며, `dry_run=false` 또는 `force=true` 요청은 승인 API가 확인할 수 있도록 `requires_approval=true`로 기록한다. 승인 불필요 dry-run command는 백엔드 worker가 `running` 이후 `succeeded`로 자동 전이하고 `command.status.updated` WebSocket event를 publish한다. 승인된 live service action은 worker가 `FOR UPDATE SKIP LOCKED` 기반 claim으로 `running` 전이한 뒤 executor adapter 후보로 처리한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

본 endpoint 의 actor 는 Bearer token 검증 결과 (`AuthenticatedActor` context) 또는 인증된 session 에서 도출한다. 과거의 `X-Devhub-Actor` fallback 헤더는 [ADR-0004](./adr/0004-x-devhub-actor-removal.md) (2026-05-13) 로 폐기됐다 — prod 코드는 무시하고 회귀 방지 negative 테스트만 유지.

#### 요청 예시

```json
{
  "service_id": "runner-asia-01",
  "action_type": "restart",
  "force": false,
  "dry_run": true,
  "reason": "Runner queue is blocked",
  "idempotency_key": "01HX-service-restart-runner-asia-01"
}
```

#### 응답 예시

```json
{
  "status": "accepted",
  "data": {
    "command_id": "cmd-001",
    "command_status": "pending",
    "requires_approval": false,
    "audit_log_id": "audit-001",
    "idempotent_replay": false,
    "created_at": "2026-05-04T10:00:00Z"
  }
}
```

### `POST /api/v1/risks/{risk_id}/mitigations` (API-16)

Manager dashboard의 리스크 완화 요청도 동일한 command lifecycle을 따른다. 1차 구현은 risk 상태를 즉시 변경하지 않고 `pending` command와 audit log를 생성한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

본 endpoint 의 actor 는 Bearer token 검증 결과 (`AuthenticatedActor` context) 또는 인증된 session 에서 도출한다. 과거의 `X-Devhub-Actor` fallback 헤더는 [ADR-0004](./adr/0004-x-devhub-actor-removal.md) (2026-05-13) 로 폐기됐다 — prod 코드는 무시하고 회귀 방지 negative 테스트만 유지.

#### 요청 예시

```json
{
  "action_type": "rerun_ci",
  "dry_run": true,
  "reason": "CI failure blocks release",
  "idempotency_key": "risk-502-rerun-ci",
  "metadata": {
    "ci_run_id": "502"
  }
}
```

#### 응답 예시

```json
{
  "status": "accepted",
  "data": {
    "command_id": "cmd_1f2a3b4c5d6e",
    "command_status": "pending",
    "requires_approval": false,
    "audit_log_id": "audit_1f2a3b4c5d6e",
    "idempotent_replay": false,
    "created_at": "2026-05-04T10:00:00Z"
  }
}
```

### `GET /api/v1/commands/{command_id}` (API-17)

command의 현재 상태, actor, target, 요청 사유, dry-run 여부, approval 상태, 생성/갱신 시각을 반환한다.

#### 응답 예시

```json
{
  "status": "ok",
  "data": {
    "command_id": "cmd_1f2a3b4c5d6e",
    "command_type": "risk_mitigation",
    "target_type": "risk",
    "target_id": "ci_failure:502",
    "action_type": "rerun_ci",
    "command_status": "pending",
    "actor_login": "yklee",
    "reason": "CI failure blocks release",
    "dry_run": true,
    "requires_approval": false,
    "request_payload": {
      "action_type": "rerun_ci",
      "dry_run": true,
      "reason": "CI failure blocks release"
    },
    "result_payload": {},
    "created_at": "2026-05-04T10:00:00Z",
    "updated_at": "2026-05-04T10:00:00Z"
  }
}
```

### `GET /api/v1/audit-logs` (API-18)

command 및 조직/사용자 관리 변경에서 생성된 audit log를 최신순으로 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | `1..100` |
| `offset` | `0` | pagination offset |
| `actor_login` | 없음 | actor login 필터 |
| `action` | 없음 | 예: `user.created`, `org_unit.members_replaced` |
| `target_type` | 없음 | 예: `user`, `org_unit`, `service`, `risk` |
| `target_id` | 없음 | 대상 식별자 |
| `command_id` | 없음 | command 기반 audit log 필터 |

#### 응답 예시

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

## 10. 구현된 / 예정 API 보충

- `GET /api/v1/me` (API-32) — 구현됨 (`backend-core/internal/httpapi/me.go`). authenticated actor (login / subject / role / actor_source) 반환. 별도 spec 절은 후속 sprint 의 본문 작성 후보 (현재는 본 §10 행이 1차 노출).
- Hydra/Kratos JWKS 또는 introspection 기반 Bearer token verification (예정 — ADR-0001 §9 후속 step).
- WebSocket 인증, 구독 필터, 마지막 event replay (M3 진입 시).

도메인 정규화 테이블 설계 이후 별도 spec 절로 분리한다.

### 10.1 ID 노출 표 (sprint `claude/work_260513-i` 결정 + sprint `claude/work_260513-j` 본문 spec 신설)

| API ID | 본문 위치 | endpoint set | IMPL |
| --- | --- | --- | --- |
| `API-25` | §10.2 | `/api/v1/accounts/*` admin (POST + PUT password + PATCH + DELETE) | `IMPL-account-01..04` (`accounts_admin.go`) |
| `API-33` | §10.3 | `/api/v1/users` CRUD (5 endpoint) | `IMPL-org-01` (`organization.go::users handlers`) |
| `API-34` | §10.4 | `/api/v1/organization/*` (hierarchy + units 5 endpoint + unit members 2 endpoint) | `IMPL-org-02..04` (`organization.go::org handlers`) |
| `API-36` | §8 envelope | `command.status.updated` WebSocket event envelope | `IMPL-realtime-01` (`realtime.go`) |
| `API-37` | §11.6 | command lifecycle audit 매핑 (cross-cut audit + command 도메인) | `IMPL-audit-01..02` (`audit.go` + `store/audit_logs.go`) |

### 10.2 `POST /api/v1/accounts` / `PUT /api/v1/accounts/{user_id}/password` / `PATCH /api/v1/accounts/{user_id}` / `DELETE /api/v1/accounts/{user_id}` (API-25)

System Admin 이 다른 사용자의 identity 를 발급/회수/잠금 해제할 때 사용하는 admin endpoint set. self-service 비밀번호 변경 (API-35 `POST /api/v1/account/password`) 와 구별된다.

| method/path | 용도 | audit action |
| --- | --- | --- |
| `POST /api/v1/accounts` | DevHub user + Kratos identity 동시 생성 (temp password) | `account.created` |
| `PUT /api/v1/accounts/{user_id}/password` | admin 비밀번호 재설정 (Kratos admin recovery / direct update) | `account.password_reset` |
| `PATCH /api/v1/accounts/{user_id}` | identity 상태 변경 (disable / enable, traits 조정) | `account.status_changed` |
| `DELETE /api/v1/accounts/{user_id}` | identity 회수 + DevHub user soft-delete | `account.deleted` |

#### 권한

- `security:create` (POST), `security:edit` (PUT password / PATCH), `organization:delete` (DELETE) — default policy 상 `system_admin` 단독.
- Bearer token 으로 actor 인증. legacy `X-Devhub-Actor` 헤더는 [ADR-0006](./adr/0006-x-devhub-actor-reject-inbound.md) 로 명시 거부.

#### `POST /api/v1/accounts` 요청 예시

```json
{
  "user_id": "u-new-2026",
  "email": "alice@example.com",
  "display_name": "Alice Lee",
  "role": "developer",
  "system_id": "frontend",
  "temp_password": "TempPass-1!"
}
```

응답 `201 Created`:

```json
{
  "status": "created",
  "data": {
    "user_id": "u-new-2026",
    "kratos_identity_id": "01H...",
    "temp_password_expires_at": "2026-05-20T00:00:00Z"
  },
  "meta": { "audit_log_id": "aud-001" }
}
```

에러 매트릭스 요약:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패, 필드 누락, role 값 invalid |
| 400 | `x_devhub_actor_removed` | inbound `X-Devhub-Actor` 헤더 발견 (ADR-0006) |
| 401 | `unauthenticated` | Bearer token 부재 또는 검증 실패 |
| 403 | `forbidden` | RBAC `security:create` 미보유 |
| 409 | `conflict` | `user_id` 또는 email 중복 |
| 500 | (없음) | Kratos identity 생성 실패 또는 DevHub user 저장 실패 |

> 자세한 schema (모든 endpoint 의 request/response/error 매트릭스) 는 후속 sprint 의 spec 정리 후보. 본 절은 1차 노출.

### 10.3 `/api/v1/users` CRUD (API-33)

조직 master data 의 user 차원 endpoint set.

| method/path | 용도 | audit action |
| --- | --- | --- |
| `GET /api/v1/users` | 사용자 목록 조회 (query: `unit_id`, `role`, `status`, `q`) | (조회, audit 미작성) |
| `POST /api/v1/users` | DevHub user master data 생성 (identity 발급은 §10.2 `POST /api/v1/accounts` 가 담당) | `user.created` |
| `GET /api/v1/users/{user_id}` | 개별 사용자 조회 | (조회, audit 미작성) |
| `PATCH /api/v1/users/{user_id}` | user 정보 수정 (display_name, role, status, kratos_identity_id 등) | `user.updated` |
| `DELETE /api/v1/users/{user_id}` | DevHub user soft-delete (identity 는 별도 §10.2 가 담당) | `user.deleted` |

#### 권한

- `organization:view` (GET), `organization:create` (POST), `organization:edit` (PATCH), `organization:delete` (DELETE).

#### 응답 envelope

```json
{
  "status": "ok",
  "data": { "user_id": "...", "email": "...", "display_name": "...", "role": "...", "status": "active", "system_id": "...", "kratos_identity_id": "..." },
  "meta": { "audit_log_id": "..." }
}
```

`audit_log_id` 는 mutation endpoint (POST/PATCH/DELETE) 응답에만 포함된다.

### 10.4 `/api/v1/organization/*` (API-34)

조직 계층 (hierarchy + units + unit members).

| method/path | 용도 | audit action |
| --- | --- | --- |
| `GET /api/v1/organization/hierarchy` | 전체 조직 계층 트리 조회 (parent-child 그래프) | (조회) |
| `PUT /api/v1/organization/hierarchy` | hierarchy bulk replace (parent/order 일괄 갱신) | `organization.hierarchy_updated` |
| `POST /api/v1/organization/units` | 새 organizational unit (부서) 생성 | `org_unit.created` |
| `GET /api/v1/organization/units/{unit_id}` | unit 단건 조회 (members 포함) | (조회) |
| `PATCH /api/v1/organization/units/{unit_id}` | unit 정보 수정 (name, type, parent, leader_id 등) | `org_unit.updated` |
| `DELETE /api/v1/organization/units/{unit_id}` | unit 삭제 (cascade 또는 422 — 데이터 정합성 정책은 `docs/organizational_hierarchy_spec.md` 참조) | `org_unit.deleted` |
| `GET /api/v1/organization/units/{unit_id}/members` | unit 의 members 조회 | (조회) |
| `PUT /api/v1/organization/units/{unit_id}/members` | unit members bulk replace | `org_unit.members_replaced` |

#### 권한

- `organization:view` (GET), `organization:create` (POST units), `organization:edit` (PUT/PATCH), `organization:delete` (DELETE).

#### 응답 envelope

```json
{
  "status": "ok",
  "data": { "unit_id": "...", "name": "...", "type": "team", "parent_id": "...", "leader_id": "...", "members": [...] },
  "meta": { "audit_log_id": "..." }
}
```

> 자세한 schema 는 `docs/organizational_hierarchy_spec.md` + `docs/org_chart_ux_spec.md` 참조. 본 절은 §12.8.2 의 RBAC enforcement 매핑과 cross-link 1차. 하위 mutation endpoint 의 1차 schema 보강은 §10.4.1~§10.4.4 (sprint `claude/work_260513-l`, RM-M3-03).

#### 10.4.1 `POST /api/v1/organization/units`

새 organizational unit 을 생성한다.

요청 body:

```json
{
  "unit_id": "team-frontend",
  "parent_unit_id": "dept-engineering",
  "unit_type": "team",
  "label": "Frontend Team",
  "leader_user_id": "yklee",
  "position_x": 200,
  "position_y": 120
}
```

- `unit_id`: 필수, 신규 식별자. 충돌 시 409.
- `parent_unit_id`: optional. root 단위는 빈 문자열.
- `unit_type`: 자유 텍스트 (`team`, `department`, `org` 등). 정합 규칙은 후속 spec.
- `leader_user_id`: optional. 비어 있으면 leader 미배정.
- `position_x`, `position_y`: 조직도 캔버스 좌표. UI drag 위치 영속화에 사용.

응답 (`201 Created`):

```json
{ "status": "created", "data": { /* orgUnitResponse */ }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패, `unit_id` 누락 |
| 400 | `x_devhub_actor_removed` | inbound `X-Devhub-Actor` 헤더 ([ADR-0006](./adr/0006-x-devhub-actor-reject-inbound.md)) |
| 401 | `unauthenticated` | Bearer token 부재 / 실패 |
| 403 | `forbidden` | RBAC `organization:create` 미보유 |
| 404 | `parent_not_found` | `parent_unit_id` 가 존재하지 않음 |
| 409 | `conflict` | `unit_id` 중복 |
| 500 | (없음) | 저장 실패 |

#### 10.4.2 `PATCH /api/v1/organization/units/{unit_id}`

unit 의 일부 필드를 갱신. 모든 필드 optional (pointer wire) — 명시된 필드만 변경.

요청 body:

```json
{
  "parent_unit_id": "dept-platform",
  "label": "Frontend Platform Team",
  "leader_user_id": "akim"
}
```

응답 (`200 OK`):

```json
{ "status": "ok", "data": { /* updated orgUnitResponse */ }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패 |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | (위와 동일) |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 409 | `cycle_detected` | `parent_unit_id` 변경이 cycle 을 유발 (구현은 carve out, 본 spec 은 의도 기록) |

> `parent_unit_id` 변경 시 cycle 방지 검증은 carve out — `docs/organizational_hierarchy_spec.md` §3 의 결정 항목. 본 sprint 는 spec 의도만 노출.

#### 10.4.3 `DELETE /api/v1/organization/units/{unit_id}`

unit 삭제. cascade 정책 (자식 unit / member 처리) 은 `docs/organizational_hierarchy_spec.md` 참조.

응답 (`200 OK`):

```json
{ "status": "deleted", "data": { "unit_id": "team-frontend" }, "meta": { "audit_log_id": "..." } }
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 422 | `has_children` | 자식 unit 이 존재 (cascade 미지원 정책 시) |
| 422 | `has_members` | members 가 비어 있지 않음 (cascade 미지원 정책 시) |

#### 10.4.4 `PUT /api/v1/organization/units/{unit_id}/members`

unit 의 member 목록을 bulk replace. 누락된 user 는 unit 에서 제거, 신규 user 는 추가.

요청 body:

```json
{
  "user_ids": ["yklee", "akim", "sjones"]
}
```

응답 (`200 OK`):

```json
{
  "status": "ok",
  "data": {
    "unit_id": "team-frontend",
    "members": [ /* user array */ ]
  },
  "meta": { "audit_log_id": "..." }
}
```

에러 매트릭스:

| status | code | 의미 |
| --- | --- | --- |
| 400 | (없음) | body parse 실패 |
| 401 / 403 / 400 (`x_devhub_actor_removed`) | (auth/RBAC 공통) | |
| 404 | `not_found` | `unit_id` 가 존재하지 않음 |
| 422 | `unknown_user_ids` | `user_ids` 중 DevHub `users` 에 없는 항목 — 응답 detail 에 목록 포함 |

> primary_dept 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수) 은 본 endpoint 의 후속 결정 — `docs/backend_requirements_org_hierarchy.md` §1·2 의 미해결 항목. 본 sprint 는 spec 의도만 노출.

## 11. 계정 및 인증 (Hydra/Kratos)

DevHub는 자체 `/api/v1/accounts/*`, `/api/v1/auth/*` 인증 API를 만들지 않는다. 인증과 credential/session lifecycle은 Ory Hydra/Kratos가 소유하고, Go Core는 검증된 token claim에서 actor를 도출한다.

정책 기준은 [ADR-0001](./adr/0001-idp-selection.md), [architecture.md 6.2절](./architecture.md#62-사용자user--계정account-도메인-분리)을 따른다.

### 11.1 소유권 분리

| 영역 | Source of truth | DevHub 역할 |
| --- | --- | --- |
| 조직/사용자 master data | Go Core `users`, `org_units`, `unit_memberships` | 사용자/조직 CRUD, 권한/소속 조회, audit |
| credential, recovery, session | Kratos | identity, password, recovery, verification flow |
| OAuth2/OIDC token | Hydra | authorization code, token, JWKS, introspection |
| frontend session UX | Next.js + Kratos public flow | 로그인/로그아웃/비밀번호 변경 flow orchestration |

`users.user_id`는 Kratos identity 또는 Hydra ID token의 안정적인 subject와 1:1로 매핑한다. email/display name/status 같은 업무 속성은 DevHub `users`가 제공하고, credential secret은 DevHub API 응답과 audit payload에 포함하지 않는다.

### 11.2 Hydra 표준 endpoint

다른 앱과 DevHub frontend는 Hydra 표준 endpoint를 사용한다. Go Core는 아래 endpoint를 재정의하지 않는다.

| endpoint | 용도 |
| --- | --- |
| `/.well-known/openid-configuration` | issuer, authorization/token/JWKS endpoint discovery |
| `/oauth2/auth` | authorization code flow 시작 |
| `/oauth2/token` | code/token 교환 |
| `/oauth2/revoke` | refresh/access token revoke |
| `/oauth2/introspect` | opaque token 또는 서버 간 token introspection |
| `/.well-known/jwks.json` 또는 discovery의 `jwks_uri` | JWT access token signature 검증 |

로컬 PoC 기본 issuer는 `infra/idp/hydra.yaml` 기준 `http://localhost:4444/`다. 운영 환경에서는 issuer, audience, JWKS URI를 환경별 설정으로 주입한다.

### 11.3 Go Core Bearer token 경계 (API-19)

Go Core `/api/v1/*` 라우터는 `Authorization: Bearer <token>`을 받으면 configured verifier에 위임한다.

- verifier가 성공하면 `subject`, `login`, `role` claim을 내부 request context에 저장하고 command/audit actor로 사용한다.
- verifier가 실패하면 `401 unauthenticated`를 반환한다.
- verifier가 설정되지 않은 개발 환경에서는 Bearer token을 actor로 신뢰하지 않고 `X-Devhub-Auth: bearer_unverified`만 응답한다.
- `X-Devhub-Actor` fallback 헤더는 [ADR-0004](./adr/0004-x-devhub-actor-removal.md) (2026-05-13) 로 폐기됐다 — prod 코드는 어떤 분기에서도 처리하지 않고 회귀 방지 negative 테스트만 유지.

현재 구현된 경계는 verifier interface와 actor context 연결까지다. Hydra JWKS 또는 introspection 기반 실제 verifier는 후속 작업에서 연결한다.

### 11.4 DevHub admin identity wrapper 예정 API

시스템 관리자가 identity 발급/회수/복구를 수행할 때는 Go Core가 Kratos admin API를 감싸는 `/api/v1/admin/identities/*` endpoint를 제공한다. 이 endpoint들은 DevHub 권한, audit log, 조직/사용자 상태 검증을 추가하는 thin wrapper다.

| method/path | 목적 | audit action |
| --- | --- | --- |
| `GET /api/v1/admin/identities` | `user_id`, `email`, `identity_id` 기준 identity 조회 | 없음 |
| `POST /api/v1/admin/identities` | DevHub user에 연결되는 Kratos identity 생성 | `identity.created` |
| `PATCH /api/v1/admin/identities/{identity_id}` | trait/status/metadata 조정 | `identity.updated` |
| `POST /api/v1/admin/identities/{identity_id}/recovery-link` | 관리자 주도 recovery link 발급 | `identity.recovery_link_created` |
| `DELETE /api/v1/admin/identities/{identity_id}` | identity 회수 또는 비활성화 | `identity.disabled` |

요청/응답 schema는 Kratos admin payload를 그대로 노출하지 않고 DevHub user 매핑과 audit metadata를 포함하는 envelope로 확정한다.

### 11.5 Self-service flow

frontend 는 OIDC code flow 진입과 권한 정합을 backend proxy 로 운반한다. 직접 Kratos public flow 만 사용하지 않고, DevHub `/api/v1/auth/*` + `/api/v1/account/*` 가 Kratos/Hydra 와의 server-to-server 통신을 담당한다 (PR-LOGIN-1/2, PR-L3/L4).

| API | method/path | 용도 | audit action |
| --- | --- | --- | --- |
| `API-20` | `POST /api/v1/auth/login` | login_challenge + identifier/password → Kratos api-mode login + Hydra accept | `auth.login.succeeded` / `auth.login.failed` |
| `API-21` | `POST /api/v1/auth/logout` | logout_challenge + refresh_token → Hydra revoke + accept | `auth.logout.succeeded` |
| `API-22` | `POST /api/v1/auth/token` | authorization_code → Hydra `/oauth2/token` 교환 | (passthrough) |
| `API-23` | `POST /api/v1/auth/signup` | HRDB lookup + Kratos identity 생성 | `account.signup.requested` |
| `API-24` | `GET /api/v1/auth/consent` | Hydra consent flow auto-accept | (passthrough) |
| `API-35` | `POST /api/v1/account/password` | **본인 비밀번호 변경** — current_password 검증 + Kratos settings flow proxy | `account.password_self_change` |

#### 11.5.2 `POST /api/v1/auth/signup` (API-23, RM-M3-01)

인사 DB 조회 → Kratos identity 생성 → DevHub user 생성 → audit log. 인증 없이 진입 가능한 self-service flow (`publicAPIPaths` 등재). 대상은 인사 DB 에 존재하는 인원에 한정 — `system_id` + `employee_id` + `name` 3 키 모두 일치해야 가입 가능 (대소문자 무시).

##### 요청 body

```json
{
  "name": "YK Lee",
  "system_id": "yklee",
  "employee_id": "1001",
  "password": "ChangeMe-12345!"
}
```

- `name`: 사용자 표시 이름. HR DB 의 `name` 과 case-insensitive 매칭.
- `system_id`: 사내 ID (계정명 prefix). HR DB 의 `system_id` 와 case-insensitive 매칭.
- `employee_id`: 사번. HR DB 의 `employee_id` 와 정확 매칭.
- `password`: 초기 비밀번호. Kratos identity 생성에 사용. 길이/복잡성은 Kratos schema 가 enforce (1차 PoC = 12자 이상 권장).

모든 필드 필수.

##### 정상 응답 (`201 Created`)

```json
{
  "status": "created",
  "data": {
    "user_id": "yklee",
    "kratos_id": "01H...",
    "department": "Engineering",
    "message": "Account created successfully. You can now sign in."
  }
}
```

- `user_id` = HR DB 의 `system_id` (DevHub `users.user_id` 와 1:1).
- `kratos_id` = Kratos identity UUID (`metadata_public.user_id` 와 매핑).
- `department` = HR DB 의 `department_name` (조직 단위 자동 배정은 carve out — 본 sprint 는 표시용).

##### 에러 매트릭스

| status | code | 의미 | 조건 |
| --- | --- | --- | --- |
| 400 | `invalid_payload` | body parse 실패 또는 필드 누락 | 필드 누락 |
| 400 | `x_devhub_actor_removed` | inbound `X-Devhub-Actor` 헤더 발견 ([ADR-0006](./adr/0006-x-devhub-actor-reject-inbound.md)) | legacy 헤더 사용 |
| 403 | `hr_lookup_failed` | HR DB 조회 실패 | 3 키 mismatch 또는 미등록 인원 |
| 500 | (없음) | Kratos identity 생성 실패 | `KratosAdmin.CreateIdentity` 에러 |
| 503 | (없음) | `HRDB` / `KratosAdmin` 미주입 | 운영 환경 누락 |

##### Audit 매핑

성공 시 두 가지 action 중 하나 emit (§11.6):

| action | target | payload | 조건 |
| --- | --- | --- | --- |
| `account.signup.requested` | `user` / `user_id` | `kratos_id`, `email`, `department`, `system_id` | DevHub user 생성 성공 (정상 path) |
| `account.signup.partial_failure` | `user` / `user_id` | `reason=devhub_user_create_failed`, `kratos_id`, `email`, `department`, `error_class` | Kratos identity 는 생성됐으나 DevHub user 생성 실패 (충돌 등). 운영자 reconciliation 대상. |

HR DB miss / Kratos 실패 시 audit row 작성 없음 — Kratos identity 가 만들어지기 전이라 reconciliation 대상이 아님.

##### IMPL / 후속

- `IMPL-auth-06` (`internal/httpapi/auth_signup.go::authSignUp`) + `IMPL-hrdb-01` (`internal/hrdb/mock.go::MockClient`, PoC).
- production HR DB 어댑터 결정: [ADR-0008](./adr/0008-hrdb-production-adapter.md) (PostgreSQL `hrdb` schema 채택, 실 구현 carve out).
- 부분 실패 (`account.signup.partial_failure`) 의 자동 rollback / retry: 후속 sprint carve.

#### 11.5.1 `POST /api/v1/account/password` (PR-L4) (API-35)

self-service 비밀번호 변경 proxy. 호출자는 자신의 OIDC access token (`Authorization: Bearer …`) 으로 인증한다. 다른 사용자의 비밀번호 재설정은 `PUT /api/v1/accounts/{user_id}/password` (PR-S3) 가 담당한다.

- 인증: required Bearer. system fallback 은 거절 (`401 reauth_required`).
- 권한: RBAC 매트릭스 bypass (§12.8.1 self-info 패턴). 본 endpoint 는 caller 본인의 identity 만 mutate.
- 흐름: caller email lookup → Kratos api-mode login(current_password) → 새 session_token 으로 settings flow → password 변경. 새 session_token 은 `KratosSessionCache` 에 user_id 기준으로 저장 (DEC-D=α, 단일 instance PoC).

요청 body:

```json
{ "current_password": "OldPass-1!", "new_password": "NewPass-2!" }
```

응답 200:

```json
{ "status": "ok", "data": { "user_id": "alice" } }
```

에러 매트릭스:

| status | code | 의미 | frontend 처리 |
| --- | --- | --- | --- |
| 400 | `validation` | Kratos 가 new_password 거절 (길이, breach, complexity). `error` 에 사유 | `SettingsFlowError(VALIDATION)` 로 inline 표시 |
| 400 | (없음) | `current_password == new_password` 또는 body parse 실패 | 폼 검증 메시지 |
| 401 | `current_password_invalid` | Kratos 가 current_password 거절 | `SettingsFlowError(CURRENT_PASSWORD_INVALID)` — 현재 비밀번호 입력 강조 |
| 401 | `reauth_required` | session_token 만료/거절, actor 누락, DevHub users miss | `SettingsFlowError(REAUTH_REQUIRED)` → `/login` |
| 410 | `flow_expired` | settings flow lifespan 경과 (생성 직후라 거의 발생 안 함) | `SettingsFlowError(FLOW_INIT_FAILED)` 재시도 안내 |
| 500 | (없음) | Kratos 호출 실패 또는 invariant 위반 | `SettingsFlowError(SUBMIT_FAILED)` |
| 503 | (없음) | `KratosLogin` / `OrganizationStore` 미주입 환경 | `SettingsFlowError(SUBMIT_FAILED)` |

frontend 는 인증 완료 후 Hydra/Kratos 세션 또는 token 에서 얻은 subject 를 기준으로 `GET /api/v1/me` 를 호출해 DevHub user profile, role, organization context 를 조회한다.

### 11.6 Audit log 매핑

| event | action | target_type |
| --- | --- | --- |
| admin identity 생성 | `identity.created` | `identity` |
| admin identity 비활성화 | `identity.disabled` | `identity` |
| admin recovery link 생성 | `identity.recovery_link_created` | `identity` |
| Kratos login 성공 webhook | `auth.login.succeeded` | `identity` |
| Kratos login 실패 webhook | `auth.login.failed` | `identity` 또는 `login_id` |
| 본인 비밀번호 변경 성공 (`POST /api/v1/account/password`) | `account.password_self_change` | `user` |
| 본인 비밀번호 변경 — current 비번 실패 | `account.password_self_change.invalid_current` | `user` |
| 본인 비밀번호 변경 — DevHub users 미존재 | `account.password_self_change.no_user` | `user` |
| token 기반 command 생성 | command별 action | `service`, `risk` 등 command target |
| RBAC role 가드 거부 | `auth.role_denied` | `route` |
| RBAC 매핑 누락 거부 (deny-by-default) | `auth.policy_unmapped` | `route` |
| RBAC policy 매트릭스 갱신 (PUT /api/v1/rbac/policies) | `rbac.policy.updated` | `rbac_role` |
| Subject role 할당 갱신 (PUT /api/v1/rbac/subjects/:id/roles) | `rbac.role.assigned` | `user` |

Kratos/Hydra event를 audit log에 반영할 때 password, recovery token, session secret, access token 전문은 저장하지 않는다. RBAC audit payload 에는 이전 role/permission 매트릭스의 `before`/`after` 다이프, 변경 actor, request_id 를 포함한다 (M1 PR-D 의 audit actor 보강과 정합).

## 12. 권한 관리 (RBAC Policy Management)

ADR-0002 채택 (2026-05-08) 으로 *DB-backed RBAC matrix + write API + per-resource 4-boolean 모델* 을 source-of-truth 로 한다. 본 §12 의 spec 은 M1 PR-G2~G6 의 구현 대상이다.

### 12.0 모델

#### 12.0.1 Role

- 시스템 정의 role 3종 (immutable id, name 변경 불가): `developer`, `manager`, `system_admin`.
- 사용자 정의 role: `system_admin` 만 생성 가능. id 는 `custom-{slug}` 패턴 권장. 시스템 role 의 권한을 *상회* 하는 매트릭스 설정 가능 (단 본 단계에서는 enforcement 가 *최상위 1개 role 단일 평가* 라 multi-role 합산 정책은 §6 미해결).
- 응답 wire 형식: `id` (snake_case 또는 `custom-*`), `name` (display, 자유 문자열), `description`, `permissions`.

#### 12.0.2 Resource (5종)

| resource | 의미 |
| --- | --- |
| `infrastructure` | 인프라 토폴로지, 노드/엣지, dashboard 메트릭, system action command, command lifecycle |
| `pipelines` | repository, issue, pull request, CI run, CI log |
| `organization` | users, org units, hierarchy, unit members, subject role assignment |
| `security` | risks, risk mitigation command, RBAC policy 자체 (조회·편집) |
| `audit` | audit-logs 조회 (audit 생성은 시스템 전용) |

> 본 5종은 ADR-0002 §4.2 채택. 신규 자원이 추가되면 본 contract 갱신 + 매핑 표 (§12.6) 갱신.

#### 12.0.3 Action (4축)

| action | 의미 |
| --- | --- |
| `view` | GET / 조회 |
| `create` | POST / 생성 또는 명령 발행 |
| `edit` | PATCH/PUT / 수정 또는 멤버 갱신 |
| `delete` | DELETE / 삭제 또는 회수 |

각 (resource, action) 좌표는 boolean. 예: `{ "security": { "view": true, "create": true, "edit": false, "delete": false } }` 는 risk 조회·mitigation 발행은 가능하지만 risk 수정/삭제는 불가.

#### 12.0.4 Audit append-only invariant

`audit` resource 의 `create`, `edit`, `delete` 는 **모든 role 에 대해 false 를 강제** (UI 노출도 readOnly). audit 항목은 시스템 코드가 작성하며 사용자 API 로 직접 mutation 하지 않는다. seed 또는 PUT 으로 true 를 설정하려 해도 store 가 invariant 검증으로 거부한다 (PR-G3 도메인 규칙).

### 12.1 시스템 default policy 매트릭스

시스템 role 3종의 *기본* 매트릭스는 다음과 같으며, M0 sprint 의 `requireMinRole` enforcement 와 *완전히 호환* 된다 (PR-G5 의 `requirePermission` 마이그레이션 시 거동 보존). seed 시 store 에 시스템 role row 로 영속화 (PR-G3).

| role \ resource | infrastructure | pipelines | organization | security | audit |
| --- | --- | --- | --- | --- | --- |
| `developer` | view | view | view | view | — |
| `manager` | view | view | view | view, create | view |
| `system_admin` | view, create, edit, delete | view, create, edit, delete | view, create, edit, delete | view, create, edit, delete | view |

> `audit` 의 create/edit/delete 는 §12.0.4 invariant 로 모든 role 에서 false. system_admin 도 view 만 true.

### 12.2 `GET /api/v1/rbac/policies` (API-26)

시스템 정의 + 사용자 정의 role 전체와 각 role 의 4축 매트릭스를 조회. 매핑 누락 resource (시스템 자원 추가됐는데 role 매트릭스가 미갱신) 는 응답에 `view=false, create=false, edit=false, delete=false` 로 채워 반환.

#### 권한

`security:view` (모든 시스템 role 이 보유 — 자기 권한 가시성).

#### 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": "developer",
      "name": "Developer",
      "description": "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true,  "create": false, "edit": false, "delete": false },
        "pipelines":      { "view": true,  "create": false, "edit": false, "delete": false },
        "organization":   { "view": true,  "create": false, "edit": false, "delete": false },
        "security":       { "view": true,  "create": false, "edit": false, "delete": false },
        "audit":          { "view": false, "create": false, "edit": false, "delete": false }
      }
    },
    {
      "id": "manager",
      "name": "Manager",
      "description": "팀 운영, risk triage, 승인 전 command 생성 권한",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true, "create": false, "edit": false, "delete": false },
        "pipelines":      { "view": true, "create": false, "edit": false, "delete": false },
        "organization":   { "view": true, "create": false, "edit": false, "delete": false },
        "security":       { "view": true, "create": true,  "edit": false, "delete": false },
        "audit":          { "view": true, "create": false, "edit": false, "delete": false }
      }
    },
    {
      "id": "system_admin",
      "name": "System Admin",
      "description": "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한",
      "system": true,
      "permissions": {
        "infrastructure": { "view": true, "create": true, "edit": true, "delete": true },
        "pipelines":      { "view": true, "create": true, "edit": true, "delete": true },
        "organization":   { "view": true, "create": true, "edit": true, "delete": true },
        "security":       { "view": true, "create": true, "edit": true, "delete": true },
        "audit":          { "view": true, "create": false, "edit": false, "delete": false }
      }
    }
  ],
  "meta": {
    "policy_version": "2026-05-08.adr-0002.v1",
    "source": "rbac_policies_store",
    "editable": true,
    "system_roles": ["developer", "manager", "system_admin"]
  }
}
```

### 12.3 `PUT /api/v1/rbac/policies` (API-27)

전체 role 또는 특정 role 의 매트릭스를 갱신한다. 시스템 role 의 *id, name, system flag* 는 변경 불가 (store invariant). 시스템 role 의 *permissions* 만 변경 가능.

#### 권한

`security:edit` (default policy 상 `system_admin` 단독).

#### 요청 예시

```json
{
  "roles": [
    {
      "id": "manager",
      "permissions": {
        "pipelines": { "view": true, "create": true, "edit": false, "delete": false }
      }
    }
  ]
}
```

- 부분 갱신 — 응답에서는 partial diff 가 적용된 *전체 매트릭스* 를 반환 (§12.2 응답과 동일 shape).
- 매트릭스 갱신 시 `audit` resource 의 view 외 다른 action 을 true 로 설정하면 422 + `audit_invariant_violation` 거부 (§12.0.4).
- `auth.policy_unmapped` audit (§12.7) 와 `rbac.policy.updated` audit 가 동일 트랜잭션에 기록.

#### 응답

200 + 전체 role 응답 (§12.2). audit log 에 `rbac.policy.updated` 항목 1건 (target_type=`rbac_role`, target_id=role id). payload 에 `before`/`after` 매트릭스 diff.

### 12.4 `POST /api/v1/rbac/policies` (사용자 정의 role 생성) (API-28)

사용자 정의 role 신규 생성. id 는 `custom-{slug}` 검증 (필수 prefix), name 은 자유 문자열, permissions 는 §12.0.3 의 4축 boolean 매트릭스.

#### 권한

`security:edit` (system_admin).

### 12.5 `DELETE /api/v1/rbac/policies/:role_id` (사용자 정의 role 삭제) (API-29)

사용자 정의 role 만 삭제 가능. 시스템 role (`developer`, `manager`, `system_admin`) 은 422 + `system_role_not_deletable`. 삭제 직전 해당 role 이 할당된 subject 가 있으면 store 가 cascade 또는 422 거부 — *cascade 거부* 채택 (subject 가 있으면 422 + `role_in_use`). 호출자가 `PUT /api/v1/rbac/subjects/:id/roles` 로 먼저 재할당한 뒤 삭제.

#### 권한

`security:edit` (system_admin).

### 12.6 `GET /api/v1/rbac/subjects/{subject_id}/roles` (API-30)

특정 사용자(Subject)에게 할당된 role 목록을 조회. 1차에서는 *single role* 만 보장 (현 backend `users.role` 단일 필드). 응답 array 의 길이는 0~1.

#### 권한

`organization:view` (모든 시스템 role 이 보유 — 사용자 정보 일부).

#### 응답 예시

```json
{
  "status": "ok",
  "data": ["manager"],
  "meta": {
    "subject_id": "u123",
    "single_role_mode": true
  }
}
```

### 12.7 `PUT /api/v1/rbac/subjects/{subject_id}/roles` (API-31)

특정 사용자에게 role 을 할당. 1차에서는 single role 만 허용 — 요청 body 의 `roles` 배열 길이는 정확히 1 (그 외는 422 + `single_role_required`). 다중 role + rank 합산은 §6 미해결로 후속 phase.

#### 권한

`organization:edit` (default policy 상 `system_admin` 단독).

#### 요청 예시

```json
{
  "roles": ["manager"]
}
```

#### 응답

200 + `{ "status": "ok", "data": ["manager"], "meta": { "subject_id": "u123" } }`. audit log 에 `rbac.role.assigned` (target_type=`user`, target_id=subject id, payload=`{before, after}`).

### 12.8 라우트 → (resource, action) 매핑 표 (API-38)

`requirePermission` 미들웨어 (PR-G5) 가 본 표를 source-of-truth 로 enforcement 한다. 표에 *없는* 보호 라우트는 §12.9 의 deny-by-default 정책에 따라 거부된다.

#### 12.8.1 매핑이 *불필요* 한 라우트 (별도 정책)

| method/path | 정책 |
| --- | --- |
| `GET /health` | public, 인증 미부착 |
| `POST /api/v1/integrations/gitea/webhooks` | HMAC 시그니처 검증 (M0 SEC-2 화이트리스트), RBAC 미적용 |
| `GET /api/v1/me` | 인증된 모든 사용자 (자기 정보) |
| `GET /api/v1/realtime/ws` | 인증된 모든 사용자, 권한 필터링은 메시지 수준 (M3 publish 분류 의존) |

#### 12.8.2 (resource, action) 매핑

| method/path | resource | action |
| --- | --- | --- |
| `GET /api/v1/dashboard/metrics` | infrastructure | view |
| `GET /api/v1/events` | infrastructure | view |
| `GET /api/v1/infra/edges` | infrastructure | view |
| `GET /api/v1/infra/nodes` | infrastructure | view |
| `GET /api/v1/infra/topology` | infrastructure | view |
| `GET /api/v1/repositories` | pipelines | view |
| `GET /api/v1/issues` | pipelines | view |
| `GET /api/v1/pull-requests` | pipelines | view |
| `GET /api/v1/ci-runs` | pipelines | view |
| `GET /api/v1/ci-runs/:ci_run_id/logs` | pipelines | view |
| `GET /api/v1/risks` | security | view |
| `GET /api/v1/risks/critical` | security | view |
| `POST /api/v1/risks/:risk_id/mitigations` | security | create |
| `GET /api/v1/audit-logs` | audit | view |
| `GET /api/v1/rbac/policy` *(legacy, §6)* | security | view |
| `GET /api/v1/rbac/policies` | security | view |
| `POST /api/v1/rbac/policies` | security | edit |
| `PUT /api/v1/rbac/policies` | security | edit |
| `DELETE /api/v1/rbac/policies/:role_id` | security | edit |
| `GET /api/v1/rbac/subjects/:subject_id/roles` | organization | view |
| `PUT /api/v1/rbac/subjects/:subject_id/roles` | organization | edit |
| `POST /api/v1/admin/service-actions` | infrastructure | create |
| `GET /api/v1/commands/:command_id` | infrastructure | view |
| `GET /api/v1/users` | organization | view |
| `POST /api/v1/users` | organization | create |
| `GET /api/v1/users/:user_id` | organization | view |
| `PATCH /api/v1/users/:user_id` | organization | edit |
| `DELETE /api/v1/users/:user_id` | organization | delete |
| `GET /api/v1/organization/hierarchy` | organization | view |
| `GET /api/v1/organization/units/:unit_id` | organization | view |
| `POST /api/v1/organization/units` | organization | create |
| `PATCH /api/v1/organization/units/:unit_id` | organization | edit |
| `DELETE /api/v1/organization/units/:unit_id` | organization | delete |
| `GET /api/v1/organization/units/:unit_id/members` | organization | view |
| `PUT /api/v1/organization/units/:unit_id/members` | organization | edit |

> 신규 v1 라우트가 추가되면 본 표에 행 추가가 *필수*. 누락 시 §12.9 deny-by-default 가 발동해 모든 사용자 거부 + audit 알림.

### 12.9 매핑 누락 정책 (deny-by-default) (API-39)

`requirePermission` 미들웨어가 라우트 처리 시점에 (resource, action) 매핑을 찾지 못하면:

1. 응답: `403 Forbidden` + `{"status":"forbidden","error":"route is not mapped to an RBAC permission"}`.
2. audit: `auth.policy_unmapped`, target_type=`route`, target_id=`c.FullPath()`, payload=`{"actor_role": "...", "method": "..."}`.
3. 운영 알림: 본 audit action 은 별도 monitoring (sprint M3 publish 확장 대상) 으로 *모든 발생을 즉시 인지* 가능하게 한다.

본 정책은 라우트 추가 시점의 매핑 표 갱신 누락을 *런타임 거부* 로 강제하기 위한 안전장치다.

### 12.10 Cache 와 무효화 (API-40)

- store 적중 비용 회피를 위해 `requirePermission` 은 in-memory matrix cache (per process) 를 유지한다.
- `PUT/POST/DELETE /api/v1/rbac/policies` 또는 `PUT /api/v1/rbac/subjects/.../roles` 머지 시 동일 프로세스 내 cache reload.
- 다중 인스턴스 환경의 cache 일관성은 §6 미해결 — 운영 phase 진입 시 pub/sub 또는 polling 으로 보강.

## 13. Application/Repository/Project 관리 API (혼합 — scaffolded + planned)

본 섹션은 `Application > Repository > Project` 관리 기능의 API 계약. `API-41 ~ API-50` 은 sprint `claude/work_260514-a` 에서 **scaffolded** (gin 라우트 + RBAC matrix + handler stub) 단계까지 도달. store body 와 응답 body 는 후속 sprint 의 carve out. `API-51 ~ API-58` 은 **planned** (route 미등록, 본 §13.4~§13.7 의 endpoint 설명만 유지).

### 13.0 §13 API ID 인덱스

| API ID | endpoint | 본문 위치 | 상태 |
| --- | --- | --- | --- |
| `API-41` | `GET /api/v1/scm/providers` | §13.1.1 | activated (sprint claude/work_260514-b) |
| `API-42` | `PATCH /api/v1/scm/providers/{provider_key}` | §13.1.1 | activated |
| `API-43` | `GET /api/v1/applications` | §13.2 | activated |
| `API-44` | `POST /api/v1/applications` | §13.2 | activated |
| `API-45` | `GET /api/v1/applications/{application_id}` | §13.2 | activated |
| `API-46` | `PATCH /api/v1/applications/{application_id}` | §13.2 | activated |
| `API-47` | `DELETE /api/v1/applications/{application_id}` (archive) | §13.2 | activated |
| `API-48` | `GET /api/v1/applications/{application_id}/repositories` | §13.3 | activated |
| `API-49` | `POST /api/v1/applications/{application_id}/repositories` | §13.3 | activated |
| `API-50` | `DELETE /api/v1/applications/{application_id}/repositories/{repo_key}` | §13.3 | activated (path: gin catch-all, `provider:org/repo` 콜론 컨벤션) |
| `API-51` | `GET /api/v1/repositories/{repository_id}/activity` | §13.4 | activated (sprint claude/work_260514-c) |
| `API-52` | `GET /api/v1/repositories/{repository_id}/pull-requests` | §13.4 | activated |
| `API-53` | `GET /api/v1/repositories/{repository_id}/build-runs` | §13.4 | activated |
| `API-54` | `GET /api/v1/repositories/{repository_id}/quality-snapshots` | §13.4 | activated |
| `API-55` | `GET /api/v1/repositories/{repository_id}/projects` + `POST` | §13.5 | activated |
| `API-56` | `GET /api/v1/projects/{project_id}` + `PATCH` + `DELETE` | §13.5 | activated |
| `API-57` | `GET /api/v1/applications/{application_id}/rollup` | §13.6 | activated (concept §13.4 normalize 실 구현 + critical 가드 흡수) |
| `API-58` | `GET /api/v1/integrations` + CRUD | §13.7 | activated (scope polymorphism application/project) |

**activated 단계 정의 (sprint claude/work_260514-b)**: gin v1 group route + RBAC matrix + handler body + store body + 요청 validation + 상태 전이 가드 + audit emit. RBAC 매트릭스에서 system_admin 만 4 신규 resource (`applications` / `application_repositories` / `projects` / `scm_providers`) 의 모든 axis true (migration 000018, ADR-0011 §4.1).

**가드 1차 (concept §13.2.1 의 부분 흡수)**:
- `planning → active`: 활성 (sync_status=active) Repository ≥1
- `active → on_hold`: `hold_reason` 필수
- `on_hold → active`: `resume_reason` 필수
- `* → archived`: `archived_reason` 필수
- `closed → planning` 같은 invalid transition: `422 invalid_status_transition`
- 미흡수 (carve out): `active → closed` 의 critical 롤업 0건 검증 (롤업 store 의존, 후속 sprint)

**audit 발급 (`application.*` / `application_repository.*` / `scm_provider.*` namespace)**:
- `application.create` / `application.update` / `application.archive`
- `application_repository.link` / `application_repository.unlink`
- `scm_provider.update`
- read endpoint (list / get) 는 audit 발급하지 않음 (운영 노이즈 회피).

### 13.1 공통 규칙

- 쓰기 권한: 기본 `system_admin` 전용.
- `pmo_manager`는 정책 확정 전 `disabled`, 쓰기 요청은 `403` + `error=role_not_enabled`.
- archive는 soft-delete로 처리한다 (`archived_at` 기록 + 기본 조회 목록 제외, `include_archived=true` 토글로 노출).
- `Application.key`는 시스템 전역 unique 식별자다.
- `Application.key`는 **immutable** — 발급 후 변경 불가. PATCH 본문에 `key` 가 포함되면 `422 application_key_immutable` 로 거절한다.
- `Application.key` 입력 정책: 영문/숫자 10자 (`^[A-Za-z0-9]{10}$`).
- DB 컬럼 길이는 정책 변경 대비 여유 길이로 유지하고, 길이/패턴 제한은 애플리케이션 검증에서 강제한다.
- provider 라우팅은 `repo_provider`를 기준으로 동작하며, backend는 내부 `SCM Adapter Registry`에서 provider별 어댑터를 선택한다.
- 미등록 provider 요청은 `422 unsupported_repo_provider`로 거절한다.
- **상태 전이 가드 표 SoT (Single Source of Truth)**: `docs/planning/project_management_concept.md` §13.2.1 — 권한/검증/실패 코드 매트릭스. 본 §13.2 PATCH 의 규칙은 그 요약이다.
- **visibility 별 데이터 공개 범위 (초안)**:
  - `public`: 메타(key/name/owner) + 진행 요약 (집계만). 멤버 목록/원본 지표는 비공개.
  - `internal`: 조직 내 사용자에게 메타 + 멤버 목록 + 롤업 요약. 원본 PR/Build 본문은 RBAC 별도.
  - `restricted`: 멤버 본인 + system_admin 만 조회. 외부 노출 없음.

### 13.1.1 SCM Provider Catalog (planned)

#### `GET /api/v1/scm/providers` (`API-41 (planned)`)

- 설명: 사용 가능한 SCM provider 목록 조회 (`enabled`, `display_name`, `adapter_version` 포함).
- `adapter_version`: provider 어댑터 모듈의 semver 문자열 (예: `1.4.0`). 갱신 주체는 어댑터 배포 파이프라인 (배포 후 마이그레이션/관리 API 로 등록). 운영 중 임의 수정 금지.

#### `PATCH /api/v1/scm/providers/{provider_key}` (`API-42 (planned)`)

- 설명: provider 활성/비활성 및 운영 설정 변경 (system_admin 전용).
- 허용 필드: `enabled`, `display_name`. `adapter_version` 은 배포 파이프라인 외 수정 불가.

### 13.2 Application

#### `GET /api/v1/applications` (`API-43 (planned)`)

- 설명: Application 목록 조회.
- Query:
  - `status` (optional): `planning|active|on_hold|closed|archived`
  - `include_archived` (optional, default `false`)
  - `q` (optional): `key`, `name` 검색

#### `POST /api/v1/applications` (`API-44 (planned)`)

- 설명: Application 신규 생성.
- 요청 body 필드:
  - `key` (required): `^[A-Za-z0-9]{10}$`
  - `name` (required)
  - `description` (optional)
  - `owner_user_id` (required)
  - `start_date`, `due_date` (optional)
  - `visibility` (required): `public|internal|restricted`
  - `status` (required): `planning|active|on_hold|closed|archived`
- 실패:
  - `409 application_key_conflict`
  - `422 invalid_application_key`

요청 예시:

```json
{
  "key": "A1B2C3D4E5",
  "name": "DevHub Platform 2026",
  "description": "Cross-repo delivery governance",
  "owner_user_id": "u-1001",
  "start_date": "2026-01-01",
  "due_date": "2026-12-31",
  "visibility": "internal",
  "status": "planning"
}
```

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "id": "1a2b3c4d-1111-2222-3333-444455556666",
    "key": "A1B2C3D4E5",
    "name": "DevHub Platform 2026",
    "owner_user_id": "u-1001",
    "visibility": "internal",
    "status": "planning",
    "created_at": "2026-05-14T01:00:00Z",
    "updated_at": "2026-05-14T01:00:00Z"
  }
}
```

#### `GET /api/v1/applications/{application_id}` (`API-45 (planned)`)

- 설명: Application 상세 조회.
- 응답 포함:
  - Application 메타
  - 연결 Repository 목록
  - 하위 Project 롤업 요약

#### `PATCH /api/v1/applications/{application_id}` (`API-46 (planned)`)

- 설명: Application 메타/상태 수정.
- 허용 필드: `name`, `description`, `owner_user_id`, `start_date`, `due_date`, `visibility`, `status` (+ 전이별 보조 필드 `hold_reason`/`resume_reason`/`archived_reason`).
- 금지 필드: `key` (immutable — 요청 body 에 포함 시 `422 application_key_immutable`).
- 상태 전이 규칙:
  - `planning -> active|on_hold|archived`
  - `active -> on_hold|closed|archived`
  - `on_hold -> active|closed|archived`
  - `closed -> archived`
  - `archived -> *` 기본 비허용 (`422 invalid_status_transition`)
- 전이 가드:
  - `planning -> active`: `active` 상태 Repository 연결 1개 이상 필요
  - `active -> on_hold`: `hold_reason` 필수
  - `on_hold -> active`: `resume_reason` 필수
  - `active -> closed`: 롤업 `critical` 0건 필요
  - `* -> archived`: `archived_reason` 필수
- 가드 실패 에러:
  - `422 application_activation_precondition_failed`
  - `422 application_close_precondition_failed`
  - `422 invalid_status_transition_payload`
  - `422 application_key_immutable`
- 가드 표 SoT: [`project_management_concept.md` §13.2.1](../planning/project_management_concept.md) (권한/가드/실패 코드 매트릭스).
- **active → closed 가드 — `claude/work_260514-c` 에서 흡수 완료**: concept §13.2.1 의 "active → closed: 롤업 `critical` 0건" 가드를 본 sprint 가 활성화. `CountApplicationCriticalWarnings` store 메서드 (1차 정의: gate_passed=false 합산 + build success rate <50% 시 +1) 를 handler updateApplication 의 active→closed 분기에서 호출, count > 0 면 `422 application_close_precondition_failed` (응답에 `critical_warning_count` 포함). 가드 임계치 외부화는 후속 (concept §13.2.1 운영 메모).
- **"활성 Repository" 정의**: concept §13.3 의 lifecycle 표에서 명시 — `sync_status='active'` 만 활성. `degraded` 는 1차 정책에서 활성 제외.

요청 예시:

```json
{
  "status": "on_hold",
  "hold_reason": "External dependency delay"
}
```

#### `DELETE /api/v1/applications/{application_id}` (`API-47 (planned)`)

- 설명: **archive 전용 (soft-delete)** — hard delete 가 아님.
- 동작: `status=archived`, `archived_at` 기록. 연결 Repository/Project 는 유지되며 `include_archived=true` 토글로만 노출.
- hard delete (영구 삭제) 는 별도 endpoint 로 후속 sprint 에서 정의 — concept §10 미해결 항목 "영구 삭제 정책" 참조.

### 13.3 Application-Repository 연결

#### `GET /api/v1/applications/{application_id}/repositories` (`API-48 (planned)`)

- 설명: Application에 연결된 Repository 조회.

#### `POST /api/v1/applications/{application_id}/repositories` (`API-49 (planned)`)

- 설명: Repository 연결 생성.
- 요청 body 필드:
  - `repo_provider` (required): `bitbucket|gitea|forgejo|github|...` (동등 지원, 특정 provider 우선순위 없음)
  - `repo_full_name` (required): `org/repo`
  - `role` (required): `primary|sub|shared`
- 실패:
  - `409 repository_link_conflict`
  - `422 invalid_repository_reference`
  - `422 unsupported_repo_provider`
  - `503 provider_unreachable`

- 연결 lifecycle:
  - 초기: `requested`
  - 검증중: `verifying`
  - 정상: `active`
  - 부분장애: `degraded` (`sync_error_code` 기록)
  - 해제: `disconnected`
- `sync_error_code` 표준 (link 단위 최신 1건 캐시):
  - `provider_unreachable` (retryable=true)
  - `auth_invalid` (retryable=false)
  - `permission_denied` (retryable=false)
  - `rate_limited` (retryable=true)
  - `webhook_signature_invalid` (retryable=false)
  - `payload_schema_mismatch` (retryable=false)
  - `resource_not_found` (retryable=false)
  - `internal_adapter_error` (retryable=true)
- **scope**: 본 `sync_error_code` 는 **link (application_id, repo_provider, repo_full_name) 단위의 최신 에러 1건만 캐시**한다. 개별 webhook event/payload 단위 상세 에러는 `webhook_events` (현행) 또는 후속 `adapter_event_logs` 테이블에 보관 (예: 동일 link 에서 1시간 동안 발생한 N건의 `webhook_signature_invalid` 는 link.sync_error_code 에 최신 1건 + 카운트는 별도 테이블).

요청 예시:

```json
{
  "repo_provider": "bitbucket",
  "repo_full_name": "team/devhub-core",
  "role": "primary"
}
```

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "application_id": "1a2b3c4d-1111-2222-3333-444455556666",
    "repo_provider": "bitbucket",
    "repo_full_name": "team/devhub-core",
    "role": "primary",
    "sync_status": "verifying",
    "sync_error_code": null,
    "sync_error_retryable": null,
    "linked_at": "2026-05-14T01:10:00Z"
  }
}
```

#### `DELETE /api/v1/applications/{application_id}/repositories/{repo_key}` (`API-50 (planned)`)

- 설명: Application-Repository 연결 해제.

### 13.4 Repository 운영 지표 조회

#### `GET /api/v1/repositories/{repository_id}/activity` (`API-51 (planned)`)

- 설명: commit/contributor/작업 추이 조회.

#### `GET /api/v1/repositories/{repository_id}/pull-requests` (`API-52 (planned)`)

- 설명: PR 상태/타임라인/활동 지표 조회.

#### `GET /api/v1/repositories/{repository_id}/build-runs` (`API-53 (planned)`)

- 설명: 빌드 실행 이력/상태/소요 시간 조회.

#### `GET /api/v1/repositories/{repository_id}/quality-snapshots` (`API-54 (planned)`)

- 설명: 정적분석/품질 점수/게이트 결과 조회.

### 13.5 Repository 하위 Project

#### `GET /api/v1/repositories/{repository_id}/projects` (`API-55 (planned)`)

- 설명: Repository 하위 Project 목록 조회.
- Query:
  - `status` (optional)
  - `include_archived` (optional, default `false`)

#### `POST /api/v1/repositories/{repository_id}/projects` (`API-55 (planned)`)

- 설명: Project 생성.
- 요청 body 필드:
  - `key` (required)
  - `name` (required)
  - `description` (optional)
  - `owner_user_id` (required)
  - `start_date`, `due_date` (optional)
  - `visibility` (required)
  - `status` (required)
- 제약:
  - `UNIQUE (repository_id, key)`

#### `GET /api/v1/projects/{project_id}` (`API-56 (planned)`)

- 설명: Project 상세 조회.
- 응답 포함:
  - Project 메타
  - 멤버/owner
  - 상/하위 마일스톤 매핑 요약
  - Integration 상태 요약

#### `PATCH /api/v1/projects/{project_id}` (`API-56 (planned)`)

- 설명: Project 메타/상태 수정.

#### `DELETE /api/v1/projects/{project_id}` (`API-56 (planned)`)

- 설명: Project archive (soft-delete) — Application archive 와 동일 규칙.

### 13.6 Application 롤업

#### `GET /api/v1/applications/{application_id}/rollup` (`API-57 (planned)`)

- 설명: Repository 단위 운영 지표를 Application 레벨로 집계해 조회.
- Query:
  - `weight_policy` (optional, default `equal`): `equal|repo_role|custom`
- 최소 집계 항목:
  - PR 상태 분포/open 지속시간
  - 빌드 성공률/평균 소요시간
  - 품질 점수 평균/게이트 실패 건수
- `meta` 필드(필수):
  - `period`: 집계 기간
  - `filters`: 적용 필터
  - `weight_policy`: 가중치 정책
  - `applied_weights`: repo별 최종 적용 가중치 맵
  - `fallbacks`: 가중치 누락/정책 불일치 시 적용된 fallback 목록
  - `data_gaps`: 누락/장애 provider 또는 repository 목록
- 검증:
  - `weight_policy=custom` 인 경우 `custom_weights`의 합은 1.0(±허용오차)이어야 한다. **허용오차 기본값 = ±0.001** (1e-3) — 합이 [0.999, 1.001] 범위면 통과.
  - 음수 가중치는 허용하지 않는다 (`422 invalid_weight_policy`).
- **Normalize 규칙 (weight_policy 별)**:
  - `equal`: 모든 연결 Repository 가 `1/N` 가중치. 0개면 `weight_policy=equal` 이라도 가중치 맵은 빈 객체, 결과는 `data_gap`.
  - `repo_role`: 기본 카탈로그 `primary=0.6 / sub=0.3 / shared=0.1`. 단일 카테고리 내 다중 repo 가 있으면 카테고리 가중치를 균등 분할 (예: primary 2개면 각 0.3). 카테고리가 0개면 해당 가중치는 다른 카테고리에 비례 재분배 후 정규화.
  - `custom`: 명시되지 않은 repo 는 `equal` 카테고리 가중치로 fallback 후 합 정규화. `fallbacks` 메타에 `reason="custom_weight_missing"` 으로 기록.

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "pull_request_distribution": {
      "open": 24,
      "draft": 5,
      "merged": 132,
      "closed": 11
    },
    "build_success_rate": 0.94,
    "build_avg_duration_seconds": 412,
    "quality_score": 83.4,
    "quality_gate_failed_count": 2
  },
  "meta": {
    "period": {
      "from": "2026-05-01T00:00:00Z",
      "to": "2026-05-14T00:00:00Z"
    },
    "filters": {
      "repository_roles": ["primary", "sub", "shared"]
    },
    "weight_policy": "repo_role",
    "applied_weights": {
      "team/devhub-core": 0.6,
      "team/devhub-web": 0.3,
      "shared/devhub-lib": 0.1
    },
    "fallbacks": [],
    "data_gaps": [
      {
        "repo_full_name": "shared/devhub-lib",
        "provider": "forgejo",
        "reason": "provider_unreachable"
      }
    ]
  }
}
```

### 13.7 Integration

#### `GET /api/v1/integrations` (`API-58 (planned)`)

- 설명: Application/Project 연계 통합 설정 조회.

#### `POST /api/v1/integrations` (`API-58 (planned)`)

- 설명: Jira/Confluence 연계 생성.
- 요청 body 필드:
  - `scope`: `application|project`
  - `integration_type`: `jira|confluence`
  - `external_key`, `url`
  - `policy`: `summary_only|execution_system`

#### `PATCH /api/v1/integrations/{integration_id}` (`API-58 (planned)`)

- 설명: 연계 정책/키 수정.

#### `DELETE /api/v1/integrations/{integration_id}` (`API-58 (planned)`)

- 설명: 연계 해제.

> **Jira 정책 cross-cut 메모 (`REQ-FR-PROJ-005` 후속)**: REQ-FR-PROJ-005 는 "Repository Jira 가 실행 SoT" 라는 하이브리드 정책을 명시한다. 그러나 `repo_provider` 가 `bitbucket|gitea|forgejo` 인 경우의 Jira 매핑 (= 비-Jira SCM 의 실행 이슈를 어떻게 Jira project 와 묶는가) 은 본 sprint 에서 결정되지 않음. concept §10 미해결 항목으로 이관, Integration sprint 에서 결정.

### 13.8 공통 에러 코드 (초안)

```text
role_not_enabled
application_key_conflict
invalid_application_key
application_key_immutable
application_activation_precondition_failed
application_close_precondition_failed
invalid_status_transition_payload
invalid_status_transition
repository_link_conflict
invalid_repository_reference
unsupported_repo_provider
provider_unreachable
webhook_signature_invalid
invalid_weight_policy
project_key_conflict
integration_policy_violation
```

## 14. 개발 의뢰 (Dev Request, DREQ) API

외부 시스템에서 들어오는 개발 의뢰 (Dev Request, DREQ) 의 수신/조회/등록(promote)/거절/재할당/닫기 API. 컨셉: [`docs/planning/development_request_concept.md`](./planning/development_request_concept.md). 요구사항: [§5.5 (REQ-FR-DREQ-001..011, REQ-NFR-DREQ-001..006)](./requirements.md). Usecase: [`UC-DREQ-01..10`](./planning/system_usecases.md). 아키텍처: [`docs/architecture.md §7`](./architecture.md) (ARCH-DREQ-01..06).

본 §의 모든 endpoint 는 sprint `claude/work_260515-f` 에서 *spec only* (planned). backend 구현은 carve out — DREQ-AuthADR 머지 후 DREQ-Backend sprint 에서 활성화.

### 14.1 외부 수신 — `POST /api/v1/dev-requests`  *(API-59)*

- **인증**: 별도 middleware `requireIntakeToken`. [ADR-0012](./adr/0012-dreq-external-intake-auth.md) 가 옵션 A (API 토큰 + IP allowlist) 채택.
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

### 14.2 목록 — `GET /api/v1/dev-requests`  *(API-60)*

- **인증**: OIDC + RBAC `dev_requests:view`.
- **권한**: `system_admin` / `pmo_manager` 는 전체. 그 외 role 은 `assignee_user_id == actor.login` 의 row 만 (route-level RBAC + handler 단의 server-side filter).
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

### 14.3 상세 — `GET /api/v1/dev-requests/:id`  *(API-61)*

- **인증**: OIDC + RBAC `dev_requests:view` + row-level (system_admin / pmo_manager / assignee 본인).
- **응답 — 200** `{ "status":"ok", "data": <dev_request> }`. **404** not found. **403** `auth_row_denied` (audit `auth.row_denied`).

### 14.4 Promote (등록) — `POST /api/v1/dev-requests/:id/register`  *(API-62)*

- **인증**: OIDC + RBAC `dev_requests:edit` + row-level (system_admin / pmo_manager / assignee 본인).
- **요청 schema (mutual exclusion)**: 다음 셋 중 정확히 하나만 채워야 한다. 둘 이상 채우거나 모두 비우면 `400 dev_request_register_payload_invalid`. (sprint `claude/work_260515-m` 도입)
  1. `target_id` (legacy 매핑) — 이미 존재하는 application/project id 로 dev_request 를 묶기만 한다. dev_requests row 만 UPDATE 한다 (단일 row, 트랜잭션 불요).
  2. `application_payload` (target_type=application 필수) — 새 Application 을 생성하고 dev_request 를 registered 로 갱신한다. **단일 Postgres 트랜잭션** (REQ-FR-DREQ-005, ADR-0013 §5). `primary_repo` 필드는 optional 이며 함께 application_repositories 행 1개를 추가한다.
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
    "application_id": "",
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

### 14.5 거절 — `POST /api/v1/dev-requests/:id/reject`  *(API-63)*

- **인증**: OIDC + RBAC + row-level (system_admin / pmo_manager / assignee 본인).
- **요청**: `{ "rejected_reason": "중복 의뢰 (OPS-2026-00481 과 동일)" }` — `rejected_reason` 필수.
- **응답 — 200** `{ "status":"ok", "data": <dev_request with status=rejected> }`. **400** reason 누락. **409** 이미 registered/closed/rejected.

### 14.6 재할당 — `PATCH /api/v1/dev-requests/:id`  *(API-64)*

- **인증**: OIDC + RBAC `dev_requests:edit` + **system_admin 만** (담당자 변경은 row owner 가 self-change 불가하도록 RBAC 으로 강제).
- **요청**: `{ "assignee_user_id": "alice" }` — 1차에서는 assignee 만 변경 가능. title/details 등 본문 수정은 carve out.
- **응답 — 200** `{ "status":"ok", "data": <dev_request> }`. audit `dev_request.reassigned` + payload `{from_assignee, to_assignee}`.

### 14.7 닫기 — `DELETE /api/v1/dev-requests/:id`  *(API-65)*

- **인증**: OIDC + RBAC `dev_requests:delete` — **system_admin 만**. (REQ-FR-DREQ-008 + ARCH-DREQ-04 의 pmo_manager 매트릭스가 delete 권한을 부여하지 않음과 정합. codex PR #121 review P1, sprint `claude/work_260515-h` 반영.)
- **전이**: `registered` 또는 `rejected` → `closed`. `pending` / `in_review` 는 거부 (먼저 reject 후 close).
- **응답 — 200** `{ "status":"ok", "data": <dev_request with status=closed> }`. **422** `invalid_status_transition_close` (pending/in_review 에서 시도). audit `dev_request.closed`.

### 14.8 에러 코드 카탈로그 (DREQ 신규)

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

### 14.9 API ID 인덱스 (sprint `claude/work_260515-f`, intake token admin sprint `o`)

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

### 14.10 Intake Token Admin (API-66..68, sprint `claude/work_260515-o` / ADR-0014)

`dev_request_intake_tokens` resource 의 system_admin 일임 endpoint. plain token 은 발급 응답에 1회만 노출, server 는 SHA-256(plain) hex 만 보관. accounts_admin temp_password 패턴과 정합.

#### API-66 `POST /api/v1/dev-request-tokens` — 발급

- **인증**: OIDC + RBAC `dev_request_intake_tokens:create` (system_admin only).
- **요청**: `{ "client_label": "ops_portal", "source_system": "ops", "allowed_ips": ["10.0.0.0/24", "192.0.2.7"] }`. 모두 필수. `allowed_ips` 빈 배열 거절 (`invalid_allowed_ips`).
- **처리**: server 가 32-byte base64url plain token 생성 → SHA-256 hex 저장. `created_by` = actor.login.
- **응답 — 201**: `plain_token` 1회 노출 + token_id / client_label / source_system / allowed_ips / created_at / created_by / last_used_at / revoked_at. **`hashed_token` 미노출**.
- **audit**: `dev_request_intake_token.issued` (plain/hashed 모두 미포함).
- **에러**: 400 `invalid_allowed_ips` / 400 missing client_label or source_system / 409 `intake_token_collision`.

#### API-67 `GET /api/v1/dev-request-tokens` — 목록

- **인증**: OIDC + RBAC `dev_request_intake_tokens:view` (system_admin only).
- **응답 — 200**: `{ "data": [{...}], "meta": {"total": N} }`. revoked 행 포함, `created_at DESC`. **`plain_token` / `hashed_token` 모두 미노출**.

#### API-68 `DELETE /api/v1/dev-request-tokens/:token_id` — revoke

- **인증**: OIDC + RBAC `dev_request_intake_tokens:delete` (system_admin only).
- **처리**: `revoked_at = COALESCE(revoked_at, NOW())` — idempotent.
- **응답 — 200**: 갱신된 row (plain_token 미포함). **404** `not_found` (token_id 미존재).
- **audit**: `dev_request_intake_token.revoked`.

## 15. 외부 시스템 연동 (Integration) API 초안

본 섹션은 [`docs/planning/external_system_integration_concept.md`](./planning/external_system_integration_concept.md) 및 [`docs/requirements.md §5.6`](./requirements.md) 의 1차 API 계약 초안이다. ID는 임시 발급(`API-69..78`)이며 상세 응답 스키마는 설계 sprint 에서 확정한다.

### 15.1 API ID 인덱스 (draft)

| API ID | endpoint | 목적 |
| --- | --- | --- |
| API-69 | `GET /api/v1/integration/providers` | Provider catalog 조회 |
| API-70 | `POST /api/v1/integration/providers` | Provider 등록 |
| API-71 | `PATCH /api/v1/integration/providers/{provider_id}` | Provider 수정/활성화/비활성화 |
| API-72 | `POST /api/v1/integration/providers/{provider_id}/sync` | Provider 수동 재동기화 트리거 |
| API-73 | `POST /api/v1/integration/providers/{provider_key}/webhook` | Provider webhook ingest |
| API-74 | `GET /api/v1/integration/bindings` | scope별 Integration binding 조회 |
| API-75 | `POST /api/v1/integration/bindings` | scope별 Integration binding 생성 |
| API-76 | `GET /api/v1/infra/services` | 홈랩 서비스 인벤토리 조회 |
| API-77 | `POST /api/v1/infra/services/snapshot` | 홈랩 서비스 상태 스냅샷 수집 ingest |
| API-78 | `GET /api/v1/infra/topology/v2` | 노드+서비스+의존성 통합 토폴로지 조회 |

### 15.2 Provider Catalog

#### API-69 `GET /api/v1/integration/providers`

- **인증**: OIDC + RBAC (`infrastructure:view` 또는 `pipelines:view`).
- **응답 — 200**: provider 목록 + `provider_type`, `enabled`, `capabilities`, `sync_status`, `last_sync_at`, `last_error_code`.

#### API-70 `POST /api/v1/integration/providers`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `provider_key`, `provider_type`, `display_name`, `auth_mode`, `credentials_ref`, `capabilities`, `scope`.
- **응답 — 201**: 생성된 provider.
- **에러**: 409 `integration_provider_conflict`, 400 `invalid_provider_type`.

요청 예시:

```json
{
  "provider_key": "jira-main",
  "provider_type": "alm",
  "display_name": "Jira Cloud (Main)",
  "auth_mode": "oauth2",
  "credentials_ref": "secret://integrations/jira-main",
  "capabilities": ["issue.read", "epic.read", "issue.link"],
  "scope": {
    "scope_type": "project",
    "scope_id": "PRJ-001"
  }
}
```

응답 예시:

```json
{
  "status": "created",
  "data": {
    "provider_id": "8f8cdb8d-c690-458f-a243-a8b8b67f9a4d",
    "provider_key": "jira-main",
    "provider_type": "alm",
    "display_name": "Jira Cloud (Main)",
    "enabled": true,
    "auth_mode": "oauth2",
    "capabilities": ["issue.read", "epic.read", "issue.link"],
    "sync_status": "requested",
    "last_sync_at": null,
    "last_error_code": null,
    "created_at": "2026-05-15T14:00:00Z",
    "updated_at": "2026-05-15T14:00:00Z"
  }
}
```

#### API-71 `PATCH /api/v1/integration/providers/{provider_id}`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `enabled`, `display_name`, `capabilities`, `credentials_ref` 일부 수정.
- **응답 — 200**: 수정된 provider.

#### API-72 `POST /api/v1/integration/providers/{provider_id}/sync`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **설명**: provider 단위 수동 reconciliation job enqueue.
- **응답 — 202**: `{status:"accepted", job_id:"..."}`.

### 15.3 Ingest / Binding

#### API-73 `POST /api/v1/integration/providers/{provider_key}/webhook`

- **인증**: provider별 webhook 인증(header signature/token); OIDC 미적용.
- **설명**: raw event 저장 + 검증 + normalize enqueue.
- **응답**: 202 accepted / 401 invalid signature / 409 duplicate delivery.
- **검증 확장성**:
  - `Adapter Router` 가 provider별 verifier 전략을 선택해 검증한다.
  - verifier contract 예: `Verify(headers, body) -> (ok, reason)` (`hmac_sha256`, `shared_token`, `provider_sdk` 등).
- **헤더(권장 공통)**:
  - `X-Integration-Delivery`: 외부 전송 고유 ID (없으면 payload hash로 보조 dedupe)
  - `X-Integration-Event`: 이벤트 타입
  - `X-Integration-Signature`: provider 정책 기반 서명값

#### API-74 `GET /api/v1/integration/bindings`

- **인증**: OIDC + RBAC view.
- **쿼리**: `scope_type`, `scope_id`, `provider_type`, `enabled`, `limit`, `offset`.
- **응답 — 200**: binding 목록 + pagination meta.

#### API-75 `POST /api/v1/integration/bindings`

- **인증**: OIDC + RBAC edit (system_admin only).
- **요청**: `scope_type` (`application|project`), `scope_id`, `provider_id`, `external_key`, `policy`.
- **응답 — 201**: 생성 binding.
- **에러**: 409 `integration_binding_conflict`, 422 `integration_policy_violation`.

요청 예시:

```json
{
  "scope_type": "application",
  "scope_id": "APP-001",
  "provider_id": "8f8cdb8d-c690-458f-a243-a8b8b67f9a4d",
  "external_key": "PROJ",
  "policy": "execution_system"
}
```

### 15.4 HomeLab Infra

#### API-76 `GET /api/v1/infra/services`

- **인증**: OIDC + RBAC `infrastructure:view`.
- **응답 — 200**: 서비스 인벤토리(`service_id`, `node_id`, `name`, `version`, `port`, `health_status`, `observed_at`).

#### API-77 `POST /api/v1/infra/services/snapshot`

- **인증**: Agent 토큰 기반 ingest 인증 (OIDC 미적용).
- **요청**: 노드/서비스 상태 스냅샷 배열.
- **응답 — 202**: 수집 accepted + ingest_id.

요청 예시:

```json
{
  "agent_id": "homelab-agent-a",
  "snapshot_at": "2026-05-15T14:10:00Z",
  "trace_id": "trc_01jv7w2mm4m7",
  "nodes": [
    {
      "node_id": "node-nas-01",
      "hostname": "nas-01.local",
      "ip_address": "192.168.0.20",
      "environment": "homelab",
      "status": "stable",
      "metrics": { "cpu_percent": 21.3, "mem_percent": 63.1, "disk_percent": 57.2 },
      "observed_at": "2026-05-15T14:09:58Z"
    }
  ],
  "services": [
    {
      "service_id": "svc-jenkins",
      "node_id": "node-nas-01",
      "name": "jenkins",
      "version": "2.504.1",
      "port": 8080,
      "health_status": "healthy",
      "metadata": { "runtime": "docker", "compose_project": "ci-stack" },
      "observed_at": "2026-05-15T14:09:59Z"
    }
  ]
}
```

#### API-78 `GET /api/v1/infra/topology/v2`

- **인증**: OIDC + RBAC `infrastructure:view`.
- **응답 — 200**: `nodes`, `edges`, `services`, `meta`(`snapshot_at`, `degraded_providers`).

### 15.5 공통 에러 코드 (초안)

```
integration_provider_conflict
integration_provider_not_found
integration_provider_disabled
integration_binding_conflict
integration_binding_not_found
integration_policy_violation
integration_webhook_signature_invalid
integration_event_duplicate
integration_sync_job_rejected
infra_snapshot_invalid
infra_agent_unauthorized
```

### 15.6 값 제약 (draft)

- `provider_type`: `alm | scm | ci_cd | doc | infra`
- `sync_status`: `requested | verifying | active | degraded | disconnected`
- `binding.policy`: `summary_only | execution_system | bidirectional_candidate`
- `infra.health_status`: `healthy | degraded | down`
