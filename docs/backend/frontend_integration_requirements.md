# 프론트엔드 현재 구현 기반 백엔드 연동 요구사항

- 문서 목적: 현재 프론트엔드 구현을 기준으로 백엔드가 프론트엔드에 전달해야 할 계약과 백엔드 개발 필요 항목을 도출한다.
- 기준일: 2026-05-02
- 상태: draft
- 관련 문서: `docs/backend/requirements.md`, `docs/backend/requirements_review.md`, `docs/backend_api_contract.md`, `ai-workflow/memory/backend_development_roadmap.md`
- 확인 범위: `frontend/app/(dashboard)/*`, `frontend/lib/services/*`, `frontend/lib/mockData.ts`, `frontend/lib/store.ts`, `frontend/components/layout/*`

## 1. 현재 프론트엔드 구현 요약

프론트엔드는 Next.js App Router 기반의 역할별 대시보드 3종을 구현하고 있다.

- Developer: active task stream, deployment pipeline, deep focus mode, AI Gardener suggestion, infrastructure mini status.
- Manager: KPI cards, critical risk list, risk mitigation modal, resource load, decision audit, weekly report trigger.
- System Admin: infrastructure topology graph, node detail modal, service action buttons, security audit/config 진입 버튼.

현재 실제 백엔드 호출은 없고, `frontend/lib/services/infra.service.ts`, `frontend/lib/services/risk.service.ts`, `frontend/lib/mockData.ts`의 mock 데이터가 화면 요구사항의 사실상 기준이다. 따라서 백엔드 로드맵은 raw webhook 수집 이후 곧바로 프론트 교체 가능한 REST snapshot API와 WebSocket event envelope를 확정해야 한다.

## 2. 백엔드가 프론트엔드에 전달해야 할 요구사항

### 2.1 통신 경계

- 브라우저 프론트엔드는 gRPC를 직접 사용하지 않는다.
- Backend ↔ Frontend는 REST + WebSocket을 사용한다.
- Go Core ↔ Python AI 내부 분석만 gRPC를 사용한다.
- 프론트 문구와 코드 주석에서 `integrated gRPC telemetry stream`, `gRPC or REST call`처럼 브라우저 직접 gRPC로 오해될 수 있는 표현은 WebSocket/REST 기준으로 정리한다.

### 2.2 API base URL과 fetch adapter

- 프론트는 `NEXT_PUBLIC_API_URL`을 기준으로 Go Core에 연결한다.
- mock service를 실제 API로 바꿀 때 response envelope는 `status`, `data`, `meta`를 따른다.
- 표시 문자열(`12%`, `1.2GB`, `4.2d`)은 가능하면 프론트가 포맷팅하고, 백엔드는 계산 가능한 원시 값(`cpu_percent`, `memory_bytes`, `duration_seconds`)을 제공한다.

### 2.3 역할 wire format

프론트 UI 표시명은 현재 `Developer`, `Manager`, `System Admin`이다. API wire format은 다음 값으로 고정하는 것을 권장한다.

```text
developer
manager
system_admin
```

프론트는 UI label과 API role 값을 분리해야 한다.

### 2.4 실시간 연결 방식

- 화면 진입 시 REST snapshot을 먼저 조회한다.
- 이후 `GET /api/v1/realtime/ws` WebSocket으로 변경 이벤트를 받는다.
- RBAC enabled 환경에서는 WebSocket 연결 시 `types` query로 필요한 event type을 명시한다. 예: `/api/v1/realtime/ws?types=command.status.updated`.
- WebSocket 메시지는 공통 envelope를 사용한다.
- 프론트는 알 수 없는 `type`을 무시하고, `schema_version`이 맞지 않는 메시지는 로깅만 한다.

권장 envelope:

```json
{
  "schema_version": "1",
  "type": "risk.critical.created",
  "event_id": "evt-001",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {}
}
```

### 2.5 명령성 액션 처리

- 서비스 재시작, 롤백, 스케일, 일정 조정은 즉시 boolean 성공으로 처리하지 않는다.
- 백엔드는 `command_id`를 반환하고, 프론트는 command 상태 조회 또는 WebSocket command event로 완료 여부를 반영한다.
- 프론트는 위험 명령 실행 전 `dry_run` 또는 confirmation UI를 둘 수 있어야 한다.

## 3. 백엔드 개발 필요 항목

### 3.1 공통 사용자/역할/알림

필요 API:

```text
GET /api/v1/me
GET /api/v1/rbac/policy
GET /api/v1/notifications
POST /api/v1/notifications/clear
```

필요 데이터:
- user id, login, display name, role, allowed roles
- RBAC roles/resources/permissions/matrix
- unread notification count
- focus mode 상태는 초기에는 프론트 local state로 유지 가능하나, 장기적으로 사용자 설정 API로 이동 검토

2026-05-07 기준 `GET /api/v1/me`가 추가됐다. 프론트는 인증 초기화 시 이 API로 DevHub user profile, `allowed_roles`, `effective_permissions`를 조회할 수 있다. 권한 source of truth는 token role claim보다 DevHub `users.role`이다.

2026-05-07 기준 `GET /api/v1/rbac/policy`와 `PUT /api/v1/rbac/policy`가 추가됐다. Organization > Permissions 화면은 조회 API를 우선 호출하고, 실패 시 기존 default matrix로 fallback한다. 교체 API는 전체 matrix replace 방식이며 `reason`, audit log, `system_config: admin` 권한을 요구한다. 프론트 편집 UI는 confirmation UX가 정리된 뒤 활성화한다.

### 3.2 역할별 KPI metric

프론트 요구:
- Developer: active tasks, build success, code review pending
- Manager: completion, team velocity, open risks, average cycle time
- System Admin: availability, active runners, AI engine load, storage

필요 API:

```text
GET /api/v1/dashboard/metrics?role=developer
GET /api/v1/dashboard/metrics?role=manager
GET /api/v1/dashboard/metrics?role=system_admin
```

백엔드 메모:
- 초기에는 Gitea webhook raw event 기반 집계 + static fallback으로 시작할 수 있다.
- `value`/`trend`는 프론트 호환을 위해 문자열을 제공하되, 장기적으로 `numeric_value`, `unit`, `trend_direction`을 함께 제공한다.

### 3.3 Developer dashboard

필요 API:

```text
GET /api/v1/developer/active-stream
GET /api/v1/ci-runs?scope=mine
GET /api/v1/ci-runs/{ci_run_id}/logs
GET /api/v1/gardener/suggestions
POST /api/v1/gardener/suggestions/{suggestion_id}/adopt
```

필요 백엔드 개발:
- Gitea issue/PR/commit/action 이벤트 정규화
- 사용자-저장소-작업 매핑
- CI run 상태와 로그 summary 저장 모델
- AI Gardener suggestion은 Python AI 연동 전까지 rule-based 또는 mock API로 시작 가능

### 3.4 Manager dashboard

필요 API:

```text
GET /api/v1/risks/critical
GET /api/v1/team/load
GET /api/v1/decisions
POST /api/v1/risks/{risk_id}/mitigations
POST /api/v1/reports/weekly
```

필요 백엔드 개발:
- risk table과 risk status lifecycle
- risk owner, impact, status, suggested action 모델
- team load 산출 기준
- decision audit 조회 모델
- weekly report 생성 command 모델
- `ApplyRiskMitigation`은 command 생성으로 처리하고 audit log를 남긴다.

### 3.5 System Admin dashboard

필요 API:

```text
GET /api/v1/infra/nodes
GET /api/v1/infra/edges
GET /api/v1/infra/topology
POST /api/v1/admin/service-actions
GET /api/v1/admin/service-actions/{command_id}
GET /api/v1/audit-logs?scope=system
GET /api/v1/admin/config
```

필요 백엔드 개발:
- service node/edge 모델
- runner/Gitea/PostgreSQL/Go Core/Python AI health adapter
- system admin allowlist 또는 seed admin
- service action command table
- audit log table
- 2026-05-06 기준 프론트 `infraService.controlService()`는 `POST /api/v1/admin/service-actions`를 호출해 dry-run service action command를 생성한다. 백엔드는 승인 불필요 dry-run command를 `running` 이후 `succeeded`로 자동 전이하고 `command.status.updated` 이벤트를 발행한다. 프론트 후속 작업은 이 이벤트를 toast/상태 UI에 반영하는 것이다.
- 2026-05-07 기준 `GET /api/v1/audit-logs`가 추가됐고, 조직/사용자 CRUD 및 멤버 교체는 audit log를 남긴다. `X-Devhub-Actor` 사용 시 deprecation 응답 헤더를 내려 Phase 13 token actor 전환 경로를 노출한다.
- 2026-05-07 기준 RBAC enforcement가 일부 쓰기 API에 적용됐다: service action dry-run은 `commands: write`, live/force service action은 `commands: admin`, risk mitigation은 `risks: write`, audit 조회는 `system_config: read`, 조직/사용자 쓰기는 `organization: write`가 필요하다.
- 2026-05-07 기준 WebSocket 연결은 `types` query 기반 subscription filtering과 event별 RBAC read permission check를 수행한다. 프론트 `RealtimeService`는 우선 `command.status.updated`를 구독하도록 준비됐다.

### 3.6 WebSocket event

필요 event type:

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

필요 백엔드 개발:
- WebSocket hub
- role 기반 subscription filtering
- event id, timestamp, schema version, reconnect 기준
  - raw webhook 처리 후 domain event publish 경로

2026-05-07 기준 `command.status.updated`, `risk.*`, `ci.*`, `infra.*`, `notification.created` event type에 대한 권한 매핑과 subscription allowlist 1차 구현이 추가됐다. 마지막 event replay와 project/resource 세부 scope 필터는 후속 범위다.

### 3.7 Organization Management

필요 API:

```text
GET /api/v1/organizations/hierarchy
GET /api/v1/organizations/{unit_id}/members
PUT /api/v1/organizations/{unit_id}/members
```

필요 백엔드 개발:
- Division, Team, Group, Part 등 계층적 조직(Org Unit) 도메인 모델 설계
- `unit_members` 등 조직-구성원 N:M 매핑 테이블 및 구성원 할당(allocation) 상태 영속화 로직 구현
- 부모-자식 관계에 따른 `direct_count` 및 하위 조직을 모두 합산한 `total_count` 집계 최적화 (Materialized View 또는 Trigger 등 고려)
- 구성원 추가/제거 시 `command`/`audit` 및 WebSocket 이벤트를 통한 실시간 갱신 지원

### 3.8 사용자 계정 / 로그인 관리

> **재설계 예정 (2026-05-07, [ADR-0001](../adr/0001-idp-selection.md))**: 본 §3.8 의 7개 endpoint 호출은 자체 `accounts` 전제다. **Ory Hydra + Kratos 도입 결정**에 따라 프론트 흐름이 다음으로 바뀐다 — (a) 로그인 화면은 Hydra OIDC Authorization Code + PKCE 흐름을 시작하고 자격 검증은 Kratos public flow API 를 직접 호출, (b) 본인 비밀번호 변경은 Kratos self-service flow 를 호출, (c) 시스템 관리자의 계정 발급/회수/잠금 해제/강제 재설정은 신규 `/api/v1/admin/identities/*` (Kratos admin API wrapper) 를 호출. 아래 7개 endpoint 표는 historical baseline 이며 Phase 13 시작 시 교체된다.

DevHub 자체 사용자 계정(Account) 1:1 컨셉 도입에 따라 추가되는 프론트 ↔ 백엔드 연동 항목.

필요 API (계약은 `backend_api_contract.md §11`):

```text
POST   /api/v1/accounts
GET    /api/v1/accounts/{user_id}
PATCH  /api/v1/accounts/{user_id}
PUT    /api/v1/accounts/{user_id}/password
DELETE /api/v1/accounts/{user_id}
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
```

프론트 화면 영향:
- 시스템 관리자 화면 — 사용자 row 옆에 "계정 발급/회수/잠금 해제/비밀번호 강제 재설정" action.
- 모든 사용자 — "내 계정" 화면에서 로그인 ID 표시(읽기 전용)와 비밀번호 변경.
- 로그인 화면 — `login_id` + `password` 만 입력. 응답 `must_change_password=true` 면 비밀번호 변경 화면으로 강제 라우팅.
- 계정이 회수(`disabled`)되거나 잠긴(`locked`) 상태에서 사용자가 시스템에 접근 시 명시적 안내 + 시스템 관리자 연락 안내.

데이터 표시 메모:
- `password`, `password_hash`, `initial_password` 는 어떤 화면에도 표시/저장하지 않는다.
- 시스템 관리자가 강제 재설정한 임시 비밀번호는 1회 표시 후 별도 저장 없이 시스템 관리자가 사용자에게 전달하는 흐름으로 한다.

## 4. 프론트엔드에 요청할 정리 사항

- `frontend/lib/services/types.ts`의 UI 표시 타입과 API wire 타입을 분리한다.
- `UserRole` 표시명은 유지하되 API 요청에는 `developer`, `manager`, `system_admin`을 사용한다.
- `BuildLog` mock 타입과 `mockBuildLogs` 실제 shape가 다르므로 하나로 통일한다.
- `cpu`, `memory`, `duration`, `time` 같은 표시 문자열은 백엔드 원시 값에서 프론트가 포맷팅하는 방향으로 조정한다.
- `Manager` 화면의 `integrated gRPC telemetry stream` 문구는 WebSocket telemetry stream으로 바꾼다.
- 명령성 액션은 boolean 반환 대신 `command_id`, `command_status`를 받는 흐름으로 service adapter를 설계한다.
- 검색, notification, settings, weekly report, security audit, config 버튼은 아직 mock/placeholder이므로 실제 API 연결 우선순위를 프론트에서 명시한다.

## 5. 백엔드 구현 우선순위

### P1

- 프론트 교체 가능한 REST snapshot API 계약 확정: metrics, risks, ci-runs, infra topology.
- `docs/backend_api_contract.md`에 WebSocket envelope와 role wire format 반영.
- repository/issue/PR/ci_run 정규화 테이블 설계.
- risk/command/audit log 최소 schema 설계.

### P2

- WebSocket hub와 주요 event type 구현.
- system admin service action command API 구현.
- Manager risk mitigation command API 구현.
- Gitea Actions fixture 기반 CI run/log 정규화 테스트 작성.

### P3

- Python AI gRPC 연동으로 build log summary와 risk detection 고도화.
- weekly report 생성, AI Gardener suggestion, team load 산출 모델 구현.
- notification/focus mode/user settings 영속화.

## 6. Phase 2 & 3 통합 개발 피드백 (2026-05-04)

프론트엔드 Phase 2(REST API) 및 Phase 3(WebSocket) 통합 개발 과정에서 도출된 추가 요구사항 및 정정 사항이다.

### 6.1 실시간 로그 스트리밍 세부 계약
- **현황**: `AdminDashboard` 및 `DeveloperDashboard`에 로그 스트림 UI는 준비되었으나 백엔드 스트리밍 규격 미정.
- **요청**: `ci.log.appended` 이벤트 페이로드에 `ansi_text`, `line_number`, `stream_type(stdout/stderr)` 포함 여부 확정 필요.

### 6.2 사용자 상태 및 알림 실체화
- **현황**: 헤더의 알림(Notification) 및 사용자 프로필 정보가 현재 로컬 하드코딩 상태임.
- **요청**: `GET /api/v1/me`를 통한 사용자 역할(Role) 검증과 `notification.created` 이벤트를 통한 실시간 알림 카운트 갱신 연동 필요.

### 6.3 매니저용 분석 데이터 집계
- **현황**: `Team Load Balancing` 및 `Utilization Velocity`가 현재 정적 배열로 처리됨.
- **요청**: 
    - `GET /api/v1/team/load`: 팀원별 작업 부하(Active PR/Issue 수) 기반 지수 산출 API.
    - `GET /api/v1/dashboard/velocity`: gRPC 텔레메트리 기반 시계열 데이터 REST 조회 API.

### 6.4 인프라 토폴로지 레이아웃 최적화
- **현황**: 프론트에서 노드 좌표를 단순 인덱스로 계산하여 노드가 많아질 경우 겹침 가능성 있음.
- **요청**: `GET /api/v1/infra/topology` 응답 시 노드의 그룹핑 정보(tier, region 등)를 추가하여 프론트가 지능적으로 배치할 수 있도록 정보 보강 요청.

### 6.5 Idempotency Key 및 필터링 정책
- **요청**: 프론트에서 생성하는 `mitigation-{riskId}-{timestamp}` 키에 대한 백엔드 중복 체크 로직과, 사용자 역할(Role)에 따른 WebSocket 이벤트 구독 필터링(Subscription Filtering) 로직을 Phase 8의 핵심 요건으로 포함한다.
- **진행**: 2026-05-07 기준 idempotency replay와 역할 기반 WebSocket subscription filtering 1차 구현은 완료됐다. 남은 범위는 replay, resource/project scope 필터, 추가 domain event publish다.
