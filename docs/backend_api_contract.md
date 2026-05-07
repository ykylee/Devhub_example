# Backend API Contract

- 문서 목적: 프론트엔드 별도 브랜치 개발을 위한 초기 백엔드 API 계약을 기록한다.
- 기준일: 2026-05-04
- 최종 수정: 2026-05-07 (Account/Auth 계약 추가)
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

`X-Devhub-Actor`는 1차 내부/개발용 actor 전달 경계다. 운영 환경에서는 클라이언트가 임의로 보낼 수 있는 header를 신뢰하지 않고, 인증된 세션 또는 JWT claim에서 actor login과 role을 도출해야 한다.

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
- JWT/session 기반 command actor verification
- `POST /api/v1/admin/service-actions`
- `GET /api/v1/realtime/ws`

예정 API는 도메인 정규화 테이블 설계 이후 확정한다.

## 11. 계정 및 인증 (Account & Auth)

> **재설계 예정 (2026-05-07, [ADR-0001](./adr/0001-idp-selection.md))**: 본 §11 의 7개 endpoint (`/api/v1/accounts/*`, `/api/v1/auth/*`) 는 자체 `accounts` 테이블 전제로 작성됐으나, **Ory Hydra + Kratos 도입 결정에 따라 재작성된다**. 아래 본문은 정책 invariant(1:1 매핑, password 평문 미노출, audit log 매핑) 만 historical 기록으로 보존하고, 신규 contract 는 ADR-0001 §9 구현 계획 단계에서 다음 형태로 분기되어 작성된다.
>
> - **외부 클라이언트용 (다른 앱이 DevHub IdP 를 사용)**: Hydra 표준 endpoint — `/.well-known/openid-configuration`, `/oauth2/auth`, `/oauth2/token`, `/oauth2/revoke`, `/oauth2/introspect`, JWKS endpoint. DevHub 자체가 신규 contract 를 정의하지 않고 OIDC 표준 path 를 그대로 노출.
> - **DevHub 시스템 관리자용 (계정 발급/회수/잠금 해제/강제 재설정)**: `/api/v1/admin/identities/*` — Kratos admin API 를 wrap 하는 Go Core endpoint. 실제 신규 contract 정의 위치.
> - **DevHub self-service (본인 비밀번호 변경 등)**: Next.js 가 Kratos public flow API 를 직접 호출. Go Core 통과 endpoint 신규 정의 없음.
>
> §11.1 ~ §11.10 의 본문은 **참고용 historical baseline** 이며, Phase 13 코드 시작 시점에 본 절을 위 3개 카테고리로 교체한다.

DevHub 자체 사용자 계정(Account) CRUD 와 인증 lifecycle 을 정의한다. 정책 기반은 [요구사항 정의서 2.5절](./requirements.md#25-사용자-계정-관리-user-account-management) 과 [architecture.md 6.2절](./architecture.md#62-사용자user--계정account-도메인-분리)을 따른다.

### 11.1 핵심 invariant

- 사용자 1명은 정확히 0개 또는 1개의 계정을 가진다. 동일 `user_id` 에 대한 두 번째 계정 생성 요청은 `409 Conflict` 로 거절한다.
- `login_id` 는 시스템 전역 unique 다. 충돌은 `409 Conflict` 로 응답한다.
- 비밀번호 평문은 어떤 응답/audit/log 에도 포함하지 않는다.

### 11.2 Account 응답 envelope

```json
{
  "id": 12,
  "user_id": "u3",
  "login_id": "samj",
  "status": "active",
  "password_algo": "bcrypt",
  "password_changed_at": "2026-05-07T10:00:00Z",
  "last_login_at": "2026-05-07T09:55:00Z",
  "failed_login_attempts": 0,
  "created_at": "2026-04-10T12:00:00Z",
  "updated_at": "2026-05-07T10:00:00Z"
}
```

`password_hash` 는 어떤 응답에도 포함하지 않는다.

### 11.3 `POST /api/v1/accounts`

신규 계정을 발급한다. **시스템 관리자 전용.** 지정된 `user_id` 에 이미 계정이 있으면 `409 Conflict` 로 응답한다.

#### 요청

```json
{
  "user_id": "u3",
  "login_id": "samj",
  "initial_password": "TempPass!2026",
  "force_change_on_first_login": true
}
```

`force_change_on_first_login` 이 `true` 이면 계정 상태를 `password_reset_required` 로 시작한다.

#### 응답

```json
{
  "status": "created",
  "data": { /* Account envelope */ }
}
```

### 11.4 `GET /api/v1/accounts/{user_id}`

사용자에 연결된 계정을 조회한다. 본인 또는 시스템 관리자만 호출 가능. 계정이 없으면 `404 Not Found`.

### 11.5 `PATCH /api/v1/accounts/{user_id}`

계정의 `login_id` 또는 `status` 변경. **시스템 관리자 전용.** `password` 는 본 endpoint 로 변경하지 않는다.

#### 요청

```json
{
  "login_id": "sam.jones",
  "status": "disabled"
}
```

### 11.6 `PUT /api/v1/accounts/{user_id}/password`

비밀번호 변경. 본인 호출 시 `current_password` 검증을 요구하고, 시스템 관리자가 강제 재설정할 때는 `current_password` 없이 `force=true` 와 `new_password` 만으로 처리하되 결과 상태를 `password_reset_required` 로 설정한다.

#### 요청 (본인)

```json
{
  "current_password": "OldPass!2026",
  "new_password": "NewPass!2027"
}
```

#### 요청 (관리자 강제 재설정)

```json
{
  "force": true,
  "new_password": "TempPass!2027"
}
```

#### 응답

```json
{
  "status": "ok",
  "data": {
    "user_id": "u3",
    "password_changed_at": "2026-05-07T11:00:00Z",
    "status": "password_reset_required"
  }
}
```

### 11.7 `DELETE /api/v1/accounts/{user_id}`

계정 회수. **시스템 관리자 전용.** 사용자 행은 유지되며 계정만 삭제된다. 활성 세션은 즉시 무효화한다.

### 11.8 `POST /api/v1/auth/login`

로그인. 비인증 endpoint.

#### 요청

```json
{
  "login_id": "samj",
  "password": "NewPass!2027"
}
```

#### 응답 (성공)

```json
{
  "status": "ok",
  "data": {
    "user_id": "u3",
    "login_id": "samj",
    "session_token": "...",
    "expires_at": "2026-05-08T11:00:00Z",
    "must_change_password": false
  }
}
```

`must_change_password=true` 이면 프론트는 비밀번호 변경 화면으로 강제 라우팅한다.

#### 응답 (실패)

| HTTP | status | 비고 |
| --- | --- | --- |
| 401 | `unauthenticated` | login_id 부재 또는 비밀번호 불일치 — 어느 쪽인지 응답 본문에서 구분하지 않는다 |
| 423 | `locked` | 계정 잠금 |
| 403 | `disabled` | 계정 회수 상태 |

### 11.9 `POST /api/v1/auth/logout`

세션 무효화. 인증된 호출.

### 11.10 Audit log 매핑

| API | action | target_type |
| --- | --- | --- |
| `POST /accounts` | `account.created` | `account` |
| `PATCH /accounts/{id}` (status=disabled) | `account.disabled` | `account` |
| `PUT /accounts/{id}/password` | `account.password_changed` | `account` |
| `POST /auth/login` 200 | `auth.login.succeeded` | `account` |
| `POST /auth/login` 401/423/403 | `auth.login.failed` | `account` 또는 `login_id` |
| 자동 lock | `account.locked` | `account` |
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
