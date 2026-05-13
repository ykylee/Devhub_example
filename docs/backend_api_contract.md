# Backend API Contract

- 문서 목적: 프론트엔드와 백엔드 사이의 API 계약 (endpoint, request/response schema, envelope, enum) 을 단일 source-of-truth 로 기록한다.
- 범위: REST endpoint (/api/v1/**) + WebSocket envelope + 공통 enum + RBAC 정책 API (§12). 도메인 결정 근거는 `docs/architecture.md` + `docs/adr/`, 운영 시드는 `docs/setup/test-server-deployment.md`.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent, 외부 API consumer, QA.
- 상태: accepted
- 기준일: 2026-05-04
- 최종 수정일: 2026-05-13 (메타 헤더 표준화, sprint `claude/work_260513-d`. 직전 본문 갱신 2026-05-08 — §12 RBAC 모델/라우트 매핑/audit, ADR-0002 채택 반영, M1 PR-G1)
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

### 10.1 본문 spec 부재 endpoint (sprint `claude/work_260513-i` 결정)

다음 endpoint set 은 §12.8.2 의 RBAC enforcement 매핑 표에만 존재하고 본문 spec 절이 부재. 본 sprint 가 ID 만 부여 — 본문 spec 작성은 후속 sprint 후보.

| API ID | endpoint set | IMPL | 위치 |
| --- | --- | --- | --- |
| `API-25` | `/api/v1/accounts/*` admin (POST /accounts, PUT /accounts/:id/password, PATCH /accounts/:id/status, DELETE /accounts/:id) | `IMPL-account-01..04` | `backend-core/internal/httpapi/accounts_admin.go` |
| `API-33` | `/api/v1/users` CRUD (GET 조회 + POST 생성 + PATCH 수정 + DELETE 삭제) | `IMPL-org-01` (users 차원) | `backend-core/internal/httpapi/organization.go` |
| `API-34` | `/api/v1/organization/*` (hierarchy + units + members) | `IMPL-org-02..04` | `backend-core/internal/httpapi/organization.go` |
| `API-36` | `command.status.updated` WebSocket event envelope (§8) | `IMPL-realtime-01` | `backend-core/internal/httpapi/realtime.go` |
| `API-37` | command lifecycle audit 매핑 (§11.6 의 command-target audit action) | `IMPL-audit-01..02` (cross-cut) | `backend-core/internal/httpapi/audit.go` + `internal/store/audit_logs.go` |

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
