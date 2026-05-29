# Backend API Contract (Master Index)

- 문서 목적: 프론트엔드와 백엔드 사이의 API 계약 (envelope, 공통 enum, cross-cutting endpoint, 도메인별 API 진입) 의 master index 를 제공한다.
- 범위: §1-§2 envelope/enum cross-cutting (분리본 `docs/api/conventions.md` 와 동기) + §3 Health + §4-§5 Gitea Webhook 수신/조회 + §6 프론트 Snapshot API (legacy/cross-cutting) + §7 도메인 정규화 조회 API (legacy/cross-cutting) + §8 Realtime WebSocket envelope (cross-cutting, 본문은 realtime 도메인) + §9 Command/Audit 계약 (cross-cutting + audit) + §10 ID 노출 표 + §11 도메인별 API link 표.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent, 외부 API consumer, QA.
- 상태: accepted
- 기준일: 2026-05-04
- 최종 수정일: 2026-05-29 (Phase 3 split — 도메인별 본문 §11~§17 sub-document 로 이관, 본 문서는 master index 로 전환)
- 관련 문서: [공통 규약 (envelope/enum)](./api/conventions.md), [아키텍처 (master index)](./architecture.md), [기술 스택](./tech_stack.md), [프론트 연동 요구사항](./backend/frontend_integration_requirements.md), [백엔드 요구사항 리뷰](./backend/requirements_review.md), [ADR-0002 RBAC](./adr/0002-rbac-policy-edit-api.md), [백엔드 로드맵](../docs/backend_development_roadmap.md), [추적성 매트릭스](./traceability/report.md).

## 1. 공통 응답 원칙

- 성공 응답은 `status`, `data`, `meta`를 기본 envelope로 사용한다.
- 단일 command성 endpoint는 `status`와 생성/처리 결과 key를 함께 반환할 수 있다.
- 실패 응답은 `status`, `error`를 반환한다.
- 시간 값은 ISO 8601/RFC3339 형식의 UTC timestamp를 사용한다.
- API role wire format은 `developer`, `manager`, `system_admin`을 사용하고 UI 표시명과 분리한다.

> 본 절 + §2 의 단일 동기 본은 [`./api/conventions.md`](./api/conventions.md) 다. 신규 cross-cutting 결정은 conventions.md 에 먼저 작성.

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

> Gitea-specific architecture / sync worker / webhook 헤더 alias 는 [integration-registry 도메인 architecture §7](./domain/integration-registry/architecture.md) 참조.

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

## 6. 프론트 Snapshot API 1차 (cross-cutting)

프론트 mock service 교체를 위한 snapshot API다. 응답 shape는 유지하고, backing data source는 `SnapshotProvider` 경계 뒤에 둔다. 기본 구성은 runtime provider가 infra 상태를 health check로 보강하고, 나머지 snapshot은 static fallback provider에 위임한다.

### `GET /api/v1/rbac/policy` (legacy, deprecated)

> **Deprecated (M1 PR-G1)** — ADR-0002 채택으로 RBAC 모델이 *per-resource 4-boolean* 으로 통일됐다. 본 endpoint 의 1차원 (`none|read|write|admin`) 응답은 호환성 유지용으로만 남기며, 신규 통합은 [rbac-permissions api §3](./domain/rbac-permissions/api.md) `GET /api/v1/rbac/policies` (복수형) 를 사용한다. M1 PR-G4 머지와 함께 본 endpoint 는 410 Gone 으로 회수될 예정이다.

프론트 Organization > Permissions 화면이 사용할 *legacy* RBAC policy를 조회한다. 응답은 ADR-0002 이전의 1차원 모델을 보존한다. (자세한 응답 형식은 master 의 이전 revision 참조 — 본 도메인 split 후 신규 호출은 권장하지 않음.)

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

인프라 node와 edge를 한 번에 조회한다. (Topology v2 — `GET /api/v1/infra/topology/v2` API-78 — 는 [integration-registry api §4.3](./domain/integration-registry/api.md) 참조.)

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

## 7. 도메인 정규화 조회 API 1차 (cross-cutting)

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
- `status` — `active` / `draft`
- `provider_id`, `provider_key` — 연동 SCM provider (FK + derived key)
- `publish_requested_at`, `published_at` — draft → publish lifecycle (#368, repository-integration 도메인 참조)
- `updated_at`
- **linked classification (Task B, 2026-05-28)**:
  - `linked_applications_count` — `application_repositories` 의 직접 link 수
  - `linked_projects_count` — `project_repositories` 의 매핑 수
  - 합산 = 0 이면 "unlinked" (외부 SCM mirror 만 존재, orphan), > 0 이면 "linked" (시스템 application/project 와 연결됨)
  - `GET /api/v1/repositories/{id}` (repository-integration api API-53 의 detail) 응답에도 동일 필드 포함

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

## 8. Realtime WebSocket envelope (cross-cutting)

WebSocket envelope (`schema_version`, `type`, `event_id`, `occurred_at`, `data`) 와 초기 event type 카탈로그 본문은 [realtime 도메인 api §2](./domain/realtime/api.md) (API-14) + [§3](./domain/realtime/api.md) (API-97 ticket) 으로 이관됐다.

## 9. Command/Audit 계약 초안

서비스 제어와 리스크 완화 같은 명령성 액션은 즉시 boolean 성공으로 처리하지 않는다. 백엔드는 command를 생성하고 `202 Accepted`로 `command_id`, `command_status`, `audit_log_id`를 반환한다. 실행 결과는 `GET /api/v1/commands/{command_id}` 또는 `command.status.updated` WebSocket event로 추적한다.

### `POST /api/v1/admin/service-actions` (API-15)

System Admin dashboard의 서비스 제어 요청을 command lifecycle로 생성한다. `dry_run` 기본값은 `true`이며, `dry_run=false` 또는 `force=true` 요청은 승인 API가 확인할 수 있도록 `requires_approval=true`로 기록한다. 승인 불필요 dry-run command는 백엔드 worker가 `running` 이후 `succeeded`로 자동 전이하고 `command.status.updated` WebSocket event를 publish한다. 승인된 live service action은 worker가 `FOR UPDATE SKIP LOCKED` 기반 claim으로 `running` 전이한 뒤 executor adapter 후보로 처리한다. 중복 요청 방지를 위해 `idempotency_key`를 지원하며, 같은 key가 다시 들어오면 기존 command를 반환한다.

#### Header

본 endpoint 의 actor 는 Bearer token 검증 결과 (`AuthenticatedActor` context) 또는 인증된 session 에서 도출한다. 과거의 `X-Devhub-Actor` fallback 헤더는 [ADR-0004](./adr/0004-x-devhub-actor-removal.md) (2026-05-13) 로 폐기됐다 — prod 코드는 무시하고 회귀 방지 negative 테스트만 유지.

(요청/응답 예시는 master 의 이전 revision 참조 — split 후에도 본문 보존, 변경 없음.)

### `POST /api/v1/risks/{risk_id}/mitigations` (API-16)

Manager dashboard의 리스크 완화 요청도 동일한 command lifecycle을 따른다. 1차 구현은 risk 상태를 즉시 변경하지 않고 `pending` command와 audit log를 생성한다.

### `GET /api/v1/commands/{command_id}` (API-17)

command의 현재 상태, actor, target, 요청 사유, dry-run 여부, approval 상태, 생성/갱신 시각을 반환한다.

### `GET /api/v1/audit-logs` (API-18) → audit-ops 도메인

본문은 [audit-ops api §1](./domain/audit-ops/api.md) 로 이관됐다. cross-cutting 진입점은 본 master 의 §11 link 표.

## 10. 구현된 / 예정 API 보충

- `GET /api/v1/me` (API-32) — 구현됨 (`backend-core/internal/httpapi/me.go`). authenticated actor (login / subject / role / actor_source) 반환. onboarding 확장 응답은 [onboarding api §2](./domain/onboarding/api.md) 로 이관.
- Keycloak JWKS 기반 Bearer token verification ([ADR-0019](./adr/0019-keycloak-only-idp.md) §4) — [auth-session api §3](./domain/auth-session/api.md) 참조.
- WebSocket 인증, 구독 필터, 마지막 event replay — ADR-0024 채택으로 ticket 패턴 완료. [realtime api](./domain/realtime/api.md) 참조.

### 10.1 ID 노출 표 (sprint `claude/work_260513-i` 결정 + sprint `claude/work_260513-j` 본문 spec 신설)

| API ID | 본문 위치 | endpoint set | IMPL |
| --- | --- | --- | --- |
| ~~`API-25`~~ | ~~§10.2~~ | ~~`/api/v1/accounts/*` admin (POST + PUT password + PATCH + DELETE)~~ | **폐기 (sprint -i, ADR-0020 sub-carve B, 2026-05-20)** |
| `API-33` | [organization-management api §1](./domain/organization-management/api.md) | `/api/v1/users` CRUD (5 endpoint) | `IMPL-org-01` |
| `API-34` | [organization-management api §2](./domain/organization-management/api.md) | `/api/v1/organization/*` (hierarchy + units 5 endpoint + unit members 2 endpoint) | `IMPL-org-02..04` |
| `API-36` | [realtime api §2](./domain/realtime/api.md) | `command.status.updated` WebSocket event envelope | `IMPL-realtime-01` |
| `API-37` | [auth-session api §6](./domain/auth-session/api.md) + [audit-ops](./domain/audit-ops/api.md) | command lifecycle audit 매핑 (cross-cut audit + command 도메인) | `IMPL-audit-01..02` |

### 10.2 ~~`POST /api/v1/accounts` / `PUT .../password` / `PATCH .../accounts/{user_id}` / `DELETE .../accounts/{user_id}` (API-25)~~ — 폐기

**상태**: 폐기 (sprint `claude/work_260520-i-209-accounts-deprecation`, ADR-0020 sub-carve B Commit 2/3, 2026-05-20). historical body 는 master 의 이전 revision 또는 `docs/domain/auth-session/api.md` 의 폐기 표 참조.

## 11. 도메인별 API (sub-document link 표)

본 절은 도메인별 API sub-document 의 진입점이다. ID 본문 (API-*) 은 각 sub-document 가 source-of-truth.

| 도메인 | API | 본문 ID 범위 |
| --- | --- | --- |
| auth-session | [`./domain/auth-session/api.md`](./domain/auth-session/api.md) | API-19, API-32 (기본), API-35, API-37(audit 매핑) |
| audit-ops | [`./domain/audit-ops/api.md`](./domain/audit-ops/api.md) | API-18 + internal Keycloak event push endpoint |
| rbac-permissions | [`./domain/rbac-permissions/api.md`](./domain/rbac-permissions/api.md) | API-26..29, ~~API-30/31 폐기~~, API-38..40 |
| organization-management | [`./domain/organization-management/api.md`](./domain/organization-management/api.md) | API-33, API-34 (+ subpaths) |
| onboarding | [`./domain/onboarding/api.md`](./domain/onboarding/api.md) | API-32 확장, API-33 확장, API-83..86 |
| application-lifecycle | [`./domain/application-lifecycle/api.md`](./domain/application-lifecycle/api.md) | API-41..50, 55, 56, 56A, 56B, 57, 58, 93 |
| repository-integration | [`./domain/repository-integration/api.md`](./domain/repository-integration/api.md) | API-51..54, API-91, API-92 |
| dev-request | [`./domain/dev-request/api.md`](./domain/dev-request/api.md) | API-59..68, API-79 |
| integration-registry | [`./domain/integration-registry/api.md`](./domain/integration-registry/api.md) + [`task_api.md`](./domain/integration-registry/task_api.md) | API-69..78, API-80, API-87..90, API-94..96 |
| realtime | [`./domain/realtime/api.md`](./domain/realtime/api.md) | API-14, API-97 |

## 12. 변경 이력 (요약)

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | **Phase 3 split** — 도메인별 본문(§11~§17 — auth/RBAC/application/DREQ/integration/onboarding/task) 을 10 도메인 sub-document 의 `api.md` (+ Task 전용 `task_api.md`) 로 이관. §1+§2 의 cross-cutting envelope/enum 은 신규 `docs/api/conventions.md` 와 동기. §3 Health / §4-5 Gitea Webhook / §6 프론트 Snapshot API / §7 도메인 조회 API / §10 ID 노출 표 / §11 도메인 link 표 만 master 에 유지. ID 보존 (API-01..96), 신규 발급/삭제 없음. |
| 2026-05-28 | (split 이전) §17 Task Item Ingestion 신규 — task_api.md 로 이관. |
