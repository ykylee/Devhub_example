# 프론트엔드 작성 백엔드 요구사항 상세 리뷰

- 문서 목적: 프론트엔드 개발 중 작성된 백엔드 요구사항을 현재 확정 아키텍처, API 계약, 프론트 mock/service 타입 기준으로 검토한다.
- 리뷰 대상: `docs/backend/requirements.md`
- 기준일: 2026-05-02
- 상태: draft
- 관련 문서: `docs/architecture.md`, `docs/tech_stack.md`, `docs/backend/frontend_integration_requirements.md`, `docs/backend_api_contract.md`, `docs/requirements.md`, `ai-workflow/memory/backend_development_roadmap.md`

## 1. 검토 결론

프론트엔드 요구사항은 역할별 대시보드가 필요로 하는 기능 방향을 잘 드러낸다. 다만 현재 문서의 API 표기는 백엔드 구현 계약으로 사용하기에는 아직 위험하다.

가장 큰 문제는 프론트엔드와 백엔드 간 실시간 통신을 `gRPC stream`으로 적어둔 점이다. 현재 확정 아키텍처에서 gRPC는 Go Core와 Python AI 사이의 내부 분석 통신이며, 브라우저 프론트엔드는 Go Core의 REST API와 WebSocket을 사용한다. 따라서 `InfrastructureEvents`, `BuildLogs`, `CriticalAlerts`는 직접 gRPC API가 아니라 REST 초기 조회 endpoint와 WebSocket 메시지 타입으로 재정의해야 한다.

관리자 제어 명령도 단순 RPC 호출처럼 정의되어 있지만, 실제 구현에서는 권한, 감사 로그, idempotency, dry-run, 승인, 비동기 상태 추적이 필요하다. 특히 `ROLLBACK`, `SCALE`, `SCHEDULE_ADJUST`, `RESTART`, `STOP`은 서비스 상태나 일정에 영향을 주므로 `202 Accepted` 기반 command 생성과 command 상태 조회 모델을 우선 검토해야 한다.

## 2. 리뷰 기준

- Backend ↔ Frontend 기본 계약은 REST + WebSocket으로 둔다.
- Go Core ↔ Python AI 내부 분석 요청/응답은 gRPC로 둔다.
- 모든 외부 API 호출, Gitea 이벤트 처리, 권한 판단, 시스템 제어는 Go Core를 통과한다.
- 시스템 관리자 기능은 일반 관리자/PM 권한과 분리한다.
- 제어성 명령은 audit log 기록 대상이며, 재시도와 중복 제출에 안전해야 한다.
- 프론트 mock/service 타입을 실제 API 응답 모델로 교체할 수 있어야 한다.

## 3. 상세 리뷰

### 3.1 실시간 API가 gRPC로 정의됨

대상:
- `stream InfrastructureEvents(Empty) returns (stream InfraState)`
- `stream BuildLogs(BuildRequest) returns (stream LogLine)`
- `stream CriticalAlerts(UserIdentity) returns (stream AlertEvent)`

문제:
- 브라우저 프론트엔드가 Go Core 또는 Python AI의 gRPC stream에 직접 붙는 구조처럼 읽힌다.
- `docs/architecture.md`는 Backend ↔ Frontend 실시간 통신을 WebSocket으로 확정하고 있다.
- `proto/analysis.proto`의 현재 범위는 `AnalyzeBuildLog`, `DetectRisks` 같은 내부 분석 요청이며, 프론트 UI 이벤트 스트림 계약과 성격이 다르다.

권장 수정:
- REST는 초기 화면 진입 시 snapshot 조회에 사용한다.
- WebSocket은 snapshot 이후 변경 이벤트를 전달한다.
- gRPC는 Go Core가 Python AI에 분석 요청을 보낼 때만 사용한다.

권장 endpoint 초안:

```text
GET /api/v1/infra/nodes
GET /api/v1/infra/edges
GET /api/v1/ci-runs
GET /api/v1/ci-runs/{ci_run_id}/logs
GET /api/v1/risks/critical
GET /api/v1/realtime/ws
```

권장 WebSocket 메시지 타입:

```json
{
  "type": "infra.node.updated",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {
    "node_id": "backend-core",
    "status": "stable",
    "cpu_percent": 12.4,
    "memory_bytes": 1288490188,
    "active_instances": 1
  }
}
```

```json
{
  "type": "ci.log.appended",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {
    "ci_run_id": "run-101",
    "repository_name": "devhub-core",
    "step_name": "test",
    "level": "error",
    "message": "go test failed"
  }
}
```

```json
{
  "type": "risk.critical.created",
  "occurred_at": "2026-05-02T10:00:00Z",
  "data": {
    "risk_id": "risk-001",
    "severity": "high",
    "title": "Frontend CI Pipeline Delay",
    "description": "Average build time increased by 45% in last 24h.",
    "suggested_action_id": "scale-runners"
  }
}
```

### 3.2 관리자 명령 계약이 부족함

대상:
- `ControlService(ServiceActionRequest) returns (ActionResponse)`
- `ApplyRiskMitigation(MitigationRequest) returns (ActionResponse)`

문제:
- actor, role, request id, idempotency key, dry-run 여부가 없다.
- command 실행 결과가 즉시 성공인지, 비동기 접수인지 구분되지 않는다.
- 위험 명령에 대한 audit log와 승인 절차가 빠져 있다.
- `Manager`와 `System Admin` 권한 경계가 API에 드러나지 않는다.

권장 수정:
- 명령은 REST command endpoint로 생성한다.
- 응답은 `command_id`와 `status`를 반환한다.
- 실행 완료 여부는 command status 조회 또는 WebSocket 이벤트로 추적한다.
- 모든 명령 요청에는 idempotency key를 허용한다.
- 위험 명령은 `dry_run`, `requires_approval`, `approval_status`를 포함한다.

권장 endpoint 초안:

```text
POST /api/v1/admin/service-actions
GET /api/v1/admin/service-actions/{command_id}
POST /api/v1/risks/{risk_id}/mitigations
GET /api/v1/commands/{command_id}
```

권장 요청 예시:

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

권장 응답 예시:

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

### 3.3 역할 enum이 권한 정책과 맞지 않음

대상:
- `UserRole: enum { DEVELOPER, MANAGER, ADMIN }`

문제:
- 프론트 타입은 `Developer`, `Manager`, `System Admin`을 사용한다.
- 제품 요구사항은 일반 관리자/PM과 시스템 관리자를 분리한다.
- `ADMIN` 하나만 두면 프로젝트 관리 권한과 인프라 제어 권한이 섞일 수 있다.

권장 수정:
- API wire format은 소문자 snake case로 고정한다.
- 최소 역할은 `developer`, `manager`, `system_admin`으로 둔다.
- 향후 QA 등 확장 역할은 별도 enum 확장 또는 role table로 관리한다.

권장 enum:

```text
UserRole = developer | manager | system_admin
ImpactLevel = low | medium | high | critical
ServiceStatus = stable | warning | degraded | down
CommandStatus = pending | running | succeeded | failed | rejected | cancelled
```

### 3.4 데이터 모델이 프론트 mock보다 좁음

문제:
- `NodeState`에 `node_id`, `label`, `kind`, `region`, `updated_at`이 없다.
- `BuildLogs`에 build/run 식별자, repository, branch, commit SHA, duration, status가 없다.
- `Risk`에 `risk_id`, owner, status, created_at, detected_by, suggested actions가 없다.
- 목록 조회용 pagination, filtering, initial snapshot 기준이 없다.

프론트 mock/service 기준 최소 모델:

```text
ServiceNode
- id
- label
- status
- cpu_percent
- memory_bytes
- active_instances
- kind
- region
- updated_at

CiRun
- id
- repository_name
- branch
- commit_sha
- status
- duration_seconds
- started_at
- finished_at

Risk
- id
- title
- reason
- impact
- status
- owner_login
- suggested_actions
- created_at
- updated_at
```

권장 수정:
- REST 목록 조회 모델을 먼저 확정한 뒤 WebSocket 이벤트의 `data` payload가 같은 모델의 부분 갱신인지, 별도 event model인지 정한다.
- 날짜는 RFC3339 UTC timestamp로 통일한다.
- 표시용 문자열과 계산 가능한 원시 값을 분리한다. 예: `cpu_percent`는 number, UI의 `12%`는 프론트에서 포맷팅.

## 4. 우선순위별 후속 작업

### P1

- `docs/backend/requirements.md`의 실시간 API 섹션을 REST + WebSocket 기준으로 재작성한다.
- 관리자 명령 API에 command lifecycle, audit log, idempotency, 권한 조건을 추가한다.
- `docs/backend_api_contract.md`에 `GET /api/v1/realtime/ws` 메시지 envelope 초안을 추가한다.
- `developer`, `manager`, `system_admin` role wire format을 확정한다.

### P2

- `GET /api/v1/infra/nodes`, `GET /api/v1/ci-runs`, `GET /api/v1/risks/critical` 응답 예시를 추가한다.
- `ServiceNode`, `CiRun`, `Risk`, `Command` 공통 필드와 timestamp 정책을 문서화한다.
- 프론트 `frontend/lib/services/types.ts`와 백엔드 API 계약의 naming 차이를 정리한다.

### P3

- Go Core ↔ Python AI gRPC proto는 프론트 이벤트 스트림과 분리해 `analysis.proto` 또는 후속 proto에서 확장한다.
- WebSocket 인증, 재연결, 마지막 이벤트 재수신 정책을 별도 운영 항목으로 검토한다.
- 리스크 완화 조치 중 `SCHEDULE_ADJUST`는 실제 일정 시스템 또는 알림 시스템 연동 전까지 mock command 또는 draft proposal로 제한한다.

## 5. 구현 판단 메모

- 현재 백엔드 구현은 `POST /api/v1/integrations/gitea/webhooks`와 `GET /api/v1/events`까지 존재한다.
- `GET /api/v1/events`는 raw webhook event 조회용이므로, 프론트 대시보드의 최종 도메인 API를 대체하지 않는다.
- 다음 백엔드 작업은 raw event를 repository, issue, pull request, ci run 등 도메인 테이블로 정규화하는 설계와 함께 진행하는 것이 좋다.
- 프론트 요구사항 문서는 제품/화면 요구를 남기는 문서로 두고, 실제 구현 계약은 `docs/backend_api_contract.md`에 반영하는 편이 안전하다.
