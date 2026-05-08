# Backend API Contract

- 문서 목적: 프론트엔드 별도 브랜치 개발을 위한 초기 백엔드 API 계약을 기록한다.
- 기준일: 2026-05-04
- 최종 수정: 2026-05-07 (RBAC policy API 및 Account/Auth 계약 추가)
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
AccountStatus = active | disabled | locked | password_reset_required
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

### `GET /api/v1/rbac/policy`

프론트 Organization > Permissions 화면이 사용할 기본 RBAC policy를 조회한다. 현재는 읽기 전용 static default policy이며, 편집/저장 API는 권한 audit와 approval 경계 확정 뒤 추가한다.

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

### `POST /api/v1/admin/service-actions`

System Admin dashboard의 서비스 제어 요청을 command lifecycle로 생성한다. 1차 구현은 실제 executor를 호출하지 않고 `pending` command와 audit log를 기록한다. `dry_run` 기본값은 `true`이며, `dry_run=false` 또는 `force=true` 요청은 후속 승인/실행 worker가 확인할 수 있도록 `requires_approval=true`로 기록한다. 승인 불필요 dry-run command는 백엔드 worker가 `running` 이후 `succeeded`로 자동 전이하고 `command.status.updated` WebSocket event를 publish한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `X-Devhub-Actor` | `system` | command/audit actor login |

`X-Devhub-Actor`는 1차 내부/개발용 actor 전달 경계다. 운영 환경에서는 클라이언트가 임의로 보낼 수 있는 header를 신뢰하지 않고, 인증된 세션 또는 JWT claim에서 actor login과 role을 도출해야 한다.
2026-05-07 이후 이 header를 사용한 응답은 `X-Devhub-Actor-Deprecated: true`와 `Warning` header를 함께 내려 deprecation 경로를 노출한다.

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

### `POST /api/v1/risks/{risk_id}/mitigations`

Manager dashboard의 리스크 완화 요청도 동일한 command lifecycle을 따른다. 1차 구현은 risk 상태를 즉시 변경하지 않고 `pending` command와 audit log를 생성한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

| 이름 | 기본값 | 설명 |
| --- | --- | --- |
| `X-Devhub-Actor` | `system` | command/audit actor login |

`X-Devhub-Actor`는 1차 내부/개발용 actor 전달 경계다. 운영 환경에서는 클라이언트가 임의로 보낼 수 있는 header를 신뢰하지 않고, 인증된 세션 또는 JWT claim에서 actor login과 role을 도출해야 한다.
2026-05-07 이후 이 header를 사용한 응답은 `X-Devhub-Actor-Deprecated: true`와 `Warning` header를 함께 내려 deprecation 경로를 노출한다.

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

### `GET /api/v1/audit-logs`

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

## 10. 예정 API

- `GET /api/v1/me`
- Hydra/Kratos JWKS 또는 introspection 기반 Bearer token verification
- WebSocket 인증, 구독 필터, 마지막 event replay

예정 API는 도메인 정규화 테이블 설계 이후 확정한다.

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

### 11.3 Go Core Bearer token 경계

Go Core `/api/v1/*` 라우터는 `Authorization: Bearer <token>`을 받으면 configured verifier에 위임한다.

- verifier가 성공하면 `subject`, `login`, `role` claim을 내부 request context에 저장하고 command/audit actor로 사용한다.
- verifier가 실패하면 `401 unauthenticated`를 반환한다.
- verifier가 설정되지 않은 개발 환경에서는 Bearer token을 actor로 신뢰하지 않고 `X-Devhub-Auth: bearer_unverified`만 응답한다.
- `X-Devhub-Actor`는 개발 fallback으로만 유지되며, 사용 시 `X-Devhub-Actor-Deprecated: true`와 `Warning` 헤더가 내려간다.

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

본인 로그인, 로그아웃, 비밀번호 변경, recovery, verification은 Next.js가 Kratos public flow API를 호출한다. Go Core에는 별도의 `/api/v1/auth/login`, `/api/v1/auth/logout`, `/api/v1/accounts/{id}/password` endpoint를 두지 않는다.

frontend는 인증 완료 후 Hydra/Kratos 세션 또는 token에서 얻은 subject를 기준으로 `GET /api/v1/me`를 호출해 DevHub user profile, role, organization context를 조회한다.

### 11.6 Audit log 매핑

| event | action | target_type |
| --- | --- | --- |
| admin identity 생성 | `identity.created` | `identity` |
| admin identity 비활성화 | `identity.disabled` | `identity` |
| admin recovery link 생성 | `identity.recovery_link_created` | `identity` |
| Kratos login 성공 webhook | `auth.login.succeeded` | `identity` |
| Kratos login 실패 webhook | `auth.login.failed` | `identity` 또는 `login_id` |
| token 기반 command 생성 | command별 action | `service`, `risk` 등 command target |

Kratos/Hydra event를 audit log에 반영할 때 password, recovery token, session secret, access token 전문은 저장하지 않는다.

## 12. 권한 관리 (RBAC Policy Management)

프론트엔드 `PermissionEditor`에서 설정하는 역할별 세부 권한 매트릭스를 관리한다.

### 12.1 `GET /api/v1/rbac/policies`

시스템에 정의된 전체 역할(Role)과 각 역할별 권한 매트릭스를 조회한다.

#### 응답 예시

```json
{
  "status": "ok",
  "data": [
    {
      "id": "sysadmin",
      "name": "System Admin",
      "description": "Full access to all resources",
      "permissions": {
        "infrastructure": { "view": true, "create": true, "edit": true, "delete": true },
        "pipelines": { "view": true, "create": true, "edit": true, "delete": true }
      }
    },
    {
      "id": "developer",
      "name": "Developer",
      "description": "Dev access",
      "permissions": {
        "infrastructure": { "view": true, "create": false, "edit": false, "delete": false },
        "pipelines": { "view": true, "create": true, "edit": false, "delete": false }
      }
    }
  ]
}
```

### 12.2 `PUT /api/v1/rbac/policies`

전체 역할 정책을 일괄 업데이트하거나 특정 역할을 업데이트한다. (구현 시점에 따라 Patch 또는 개별 PUT으로 분화 가능)

#### 요청 예시

```json
{
  "roles": [
    {
      "id": "developer",
      "permissions": {
        "infrastructure": { "view": true, "create": true, "edit": false, "delete": false }
      }
    }
  ]
}
```

### 12.3 `GET /api/v1/rbac/subjects/{subject_id}/roles`

특정 사용자(Subject)에게 할당된 역할 목록을 조회한다.

### 12.4 `PUT /api/v1/rbac/subjects/{subject_id}/roles`

특정 사용자에게 역할을 할당하거나 수정한다.

#### 요청 예시

```json
{
  "roles": ["developer", "project_manager"]
}
```

