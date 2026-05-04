# Backend API Contract

- 문서 목적: 프론트엔드 별도 브랜치 개발을 위한 초기 백엔드 API 계약을 기록한다.
- 기준일: 2026-05-04
- 상태: draft
- 관련 문서: `docs/architecture.md`, `docs/tech_stack.md`, `docs/backend/frontend_integration_requirements.md`, `docs/backend/requirements_review.md`, `ai-workflow/memory/backend_development_roadmap.md`

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
```

Webhook event는 signature 검증과 raw 저장이 끝나면 `validated`가 되며, 정규화가 성공하면 `processed`, 재처리 가능한 오류는 `failed`, 지원하지 않거나 처리 대상이 아닌 이벤트는 `ignored`로 전환한다.

## 3. Health

### `GET /health`

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

### `POST /api/v1/integrations/gitea/webhooks`

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

### `GET /api/v1/events`

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

### `GET /api/v1/dashboard/metrics`

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

### `GET /api/v1/infra/nodes`

인프라 topology node 목록을 조회한다. CPU, memory, duration 계열 값은 프론트가 표시 문자열로 포맷팅할 수 있도록 원시 값을 우선 제공한다.
응답 `meta.source`는 snapshot provider 출처를 나타낸다.
runtime provider는 `DB_URL`, `GITEA_URL`, `BACKEND_AI_URL` 설정을 기준으로 `postgres`, `gitea`, `backend-ai` node 상태를 `stable`, `warning`, `down` 중 하나로 갱신한다.

### `GET /api/v1/infra/edges`

인프라 topology edge 목록을 조회한다.
응답 `meta.source`는 snapshot provider 출처를 나타낸다.

### `GET /api/v1/infra/topology`

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

### `GET /api/v1/repositories`

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

### `GET /api/v1/issues`

정규화된 issue 목록을 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | 조회할 항목 수 |
| `offset` | `0` | pagination offset |
| `repository_name` | 없음 | 특정 repository full name으로 필터링 |
| `state` | 없음 | `open`, `closed` 필터링 |

### `GET /api/v1/pull-requests`

정규화된 pull request 목록을 조회한다.

#### Query

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `limit` | `50` | 조회할 항목 수 |
| `offset` | `0` | pagination offset |
| `repository_name` | 없음 | 특정 repository full name으로 필터링 |
| `state` | 없음 | `open`, `closed`, `merged` 필터링 |

### `GET /api/v1/ci-runs`

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

### `GET /api/v1/ci-runs/{ci_run_id}/logs`

CI run 로그 라인을 조회한다.

### `GET /api/v1/risks`

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

### `GET /api/v1/risks/critical`

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

### `GET /api/v1/realtime/ws`

REST snapshot 조회 이후 변경 이벤트를 수신하는 WebSocket endpoint다. 브라우저 프론트엔드는 gRPC stream에 직접 연결하지 않는다.

#### 메시지 envelope

```json
{
  "schema_version": "1",
  "type": "ci.run.updated",
  "event_id": "evt-001",
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

## 9. Command/Audit 계약 초안

서비스 제어와 리스크 완화 같은 명령성 액션은 즉시 boolean 성공으로 처리하지 않는다. 백엔드는 command를 생성하고 `202 Accepted`로 `command_id`, `command_status`, `audit_log_id`를 반환한다. 실행 결과는 `GET /api/v1/commands/{command_id}` 또는 `command.status.updated` WebSocket event로 추적한다.

### `POST /api/v1/admin/service-actions`

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
    "audit_log_id": "audit-001"
  }
}
```

### `POST /api/v1/risks/{risk_id}/mitigations`

Manager dashboard의 리스크 완화 요청도 동일한 command lifecycle을 따른다. 1차 구현은 risk 상태를 즉시 변경하지 않고 `pending` command와 audit log를 생성한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `X-Devhub-Actor` | `system` | command/audit actor login |

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

### `GET /api/v1/commands/{command_id}`

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

## 10. 예정 API

- `GET /api/v1/me`
- `POST /api/v1/admin/service-actions`
- `GET /api/v1/realtime/ws`

예정 API는 도메인 정규화 테이블 설계 이후 확정한다.
