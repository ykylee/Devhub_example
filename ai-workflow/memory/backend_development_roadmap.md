# 백엔드 개발 로드맵

- 문서 목적: DevHub 백엔드 구현 범위, 순서, 진척 상태를 추적한다.
- 범위: 인프라, 리스크, 웹훅, API 로드맵
- 대상 독자: 백엔드 개발자, AI Agent
- 상태: in_progress
- 최종 수정일: 2026-05-07
- 관련 문서: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`, `docs/backend_api_contract.md`, `docs/backend/frontend_integration_requirements.md`, `docs/backend/requirements_review.md`, `ai-workflow/memory/backlog/2026-05-04.md`
- 현재 브랜치: `codex/backend_init`
- 프론트엔드 전제: frontend phase1 화면과 mock service layer가 병합된 상태이므로, 백엔드는 raw webhook 수집 이후 프론트 교체 가능한 REST snapshot API와 WebSocket 이벤트 계약을 우선 안정화한다.

## 1. 개발 원칙

- Go Core는 Gitea Webhook/API 연동, 권한 관리, 데이터 저장, 시스템 관리자 기능의 중심 서비스로 둔다.
- Python AI는 초기 단계에서 PostgreSQL에 직접 접근하지 않고, Go Core가 필터링한 입력을 gRPC로 전달받는다.
- 모든 Gitea Webhook 이벤트는 raw event로 먼저 저장하고, 정규화/분석/실시간 publish는 후속 단계에서 확장한다.
- 프론트엔드 병렬 개발을 위해 REST API 응답 형태와 WebSocket 메시지 타입을 변경 가능성이 낮은 계약으로 관리한다.
- 프론트엔드 작성 요구사항 리뷰(`docs/backend/requirements_review.md`)의 P1/P2 finding은 로드맵 phase 완료 조건에 포함한다.
- 프론트 대상 실시간 API는 gRPC stream이 아니라 REST snapshot + WebSocket event로만 계약한다.
- 프론트엔드 UI 표시명과 API wire format은 분리한다. 역할 값은 API에서 `developer`, `manager`, `system_admin`을 기본으로 한다.
- 명령성 액션은 boolean 결과가 아니라 `command_id`와 command status lifecycle로 관리하고 audit log를 남긴다.
- 검증하지 않은 단계는 `done`으로 전환하지 않는다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 산출물 | 검증 기준 |
| --- | --- | --- | --- | --- |
| Phase 1 | done | Go Core 기반 구조 정리 | `internal/config`, `internal/httpapi`, `internal/gitea`, `internal/store` 분리 | `cd backend-core && go test ./...` |
| Phase 2 | done | PostgreSQL 초기 스키마 | `webhook_events` migration | migration 적용 검증 |
| Phase 3 | done | Gitea Webhook raw 수신부 | `POST /api/v1/integrations/gitea/webhooks`, signature 검증, dedupe 처리 | handler 단위 테스트 |
| Phase 4 | in_progress | 프론트 연동 계약 안정화 | `GET /api/v1/events`, role wire format, REST snapshot/WebSocket envelope 계약, frontend integration requirements, requirements review finding 반영 | API 계약 문서화, frontend mock shape 대조, gRPC 직접 노출 없음 확인 |
| Phase 5 | in_progress | 프론트 snapshot API 1차 | metrics, infra topology, ci-runs/logs, critical risks 조회 API 초안, `ServiceNode`/`CiRun`/`Risk` 공통 필드, DB-backed domain 조회 | handler 테스트 및 응답 예시 문서화 |
| Phase 6 | done | 도메인 정규화 1차 | repository/user/issue/pull_request/ci_run/risk 기초 테이블 및 정규화 로직, 식별자/timestamp/pagination 기준 | fixture 기반 정규화 테스트 및 홈랩 DB 통합 테스트 |
| Phase 7 | in_progress | command/audit 기반 액션 API | service action, risk mitigation, weekly report command, command status, idempotency, audit log | 권한/audit/idempotency 테스트 |
| Phase 8 | planned | WebSocket 실시간 채널 | `/api/v1/realtime/ws`, infra/ci/risk/command/notification event publish | WebSocket 연결/필터링/메시지 테스트 |
| Phase 9 | planned | Python AI gRPC 연결 | Go gRPC client, Python `AnalysisService` server, build log summary/risk detection 연동 | gRPC 통합 테스트 |
| Phase 10 | planned | Hourly Pull Reconciliation | Gitea REST client, 누락 이벤트 보정 worker | dry-run 및 idempotency 테스트 |
| Phase 11 | planned | 시스템 관리자 기능 고도화 | Runner/서버 상태 adapter, config 조회, allowlist/seed admin | 권한/audit/health adapter 테스트 |
| Phase 12 | in_progress | 사용자 및 조직/멤버 관리 API | User/Org/Team 도메인 확장, 계층형 조직망 구조, 구성원 할당(Allocation), 관리자 전용 관리 API | 도메인 통합 테스트 및 권한 검증 |
| Phase 13 | planned | 사용자 계정(Account) 및 인증 1차 | `accounts` 테이블, User-Account 1:1 invariant, login/logout, 비밀번호 해시(bcrypt/argon2id), 본인/관리자 비밀번호 변경, 계정 상태 lifecycle, 로그인 audit log | DB invariant 테스트, 핸들러 단위 테스트, 비밀번호 해시 round-trip 테스트, audit log 기록 검증 |

## 3. 현재 완료 범위

- Go Core 서버 초기화 구조를 분리했다.
- `DB_URL`이 설정된 경우 PostgreSQL 연결과 `/health` DB 상태 확인을 수행한다.
- `GITEA_WEBHOOK_SECRET` 기반 HMAC-SHA256 signature 검증을 추가했다.
- Gitea event type, delivery id, repository, sender metadata를 추출한다.
- delivery id가 있으면 delivery id를 dedupe key로 사용하고, 없으면 event type과 payload hash를 조합한다.
- `webhook_events` raw event 저장 migration을 추가했다.
- signature 검증, webhook 저장, invalid signature reject, duplicate 처리 단위 테스트를 추가했다.
- `GET /api/v1/events` raw event 조회 API와 초기 API 계약 문서를 추가했다.
- validated webhook event 저장 시 `validated_at`을 함께 기록하도록 정리했다.
- 홈랩 PostgreSQL `devhub` DB에 `webhook_events` migration version 1 적용을 검증했다.
- frontend phase1 화면과 mock service layer를 검토해 `docs/backend/frontend_integration_requirements.md`로 백엔드 연동 요구사항을 정리했다.
- 프론트엔드 작성 백엔드 요구사항을 `docs/backend/requirements_review.md`에서 상세 리뷰했고, gRPC/WebSocket 경계, command/audit, role enum, 데이터 모델 부족분을 로드맵 고려 사항으로 편입했다.
- 프론트 직접 gRPC 사용 오해를 막기 위해 REST snapshot + WebSocket 갱신을 프론트 연동 기본 방향으로 명시했다.
- static fallback 기반 프론트 snapshot API 1차를 추가했다: `GET /api/v1/dashboard/metrics`, `GET /api/v1/infra/nodes`, `GET /api/v1/infra/edges`, `GET /api/v1/infra/topology`, `GET /api/v1/ci-runs`, `GET /api/v1/ci-runs/{ci_run_id}/logs`, `GET /api/v1/risks/critical`.
- `docs/backend_api_contract.md`에 role wire format, WebSocket envelope, command/audit lifecycle 초안을 보강했다.
- 도메인 정규화 1차 migration을 추가했다: `gitea_users`, `repositories`, `issues`, `pull_requests`, `ci_runs`, `risks`.
- `backend-core/internal/normalize`에 Gitea `issues`, `pull_request`, `action_run/workflow_run`, `push` 이벤트를 최소 도메인 change set으로 해석하는 processor와 fixture 테스트를 추가했다.
- 홈랩 PostgreSQL `devhub` DB에 migration version 2 적용을 검증했다.
- `backend-core/internal/domain`을 추가해 정규화 change set과 sink interface를 분리했다.
- `backend-core/internal/store`에 repository/user/issue/pull_request/ci_run upsert와 `webhook_events` 상태 전이(`processed`, `failed`, `ignored`)를 구현했다.
- `POST /api/v1/integrations/gitea/webhooks` 저장 성공 이후 optional `normalize.Processor`를 실행하도록 연결했다.
- 홈랩 PostgreSQL 통합 테스트로 raw webhook event 저장, domain upsert, `processed` 상태 전이를 검증했다.
- `GET /api/v1/repositories`, `GET /api/v1/issues`, `GET /api/v1/pull-requests` DB-backed 조회 API를 추가했다.
- `GET /api/v1/ci-runs`는 `DomainStore`가 있고 DB 결과가 있으면 DB-backed 응답을 우선 사용하고, 없으면 static fallback을 유지하도록 정리했다.
- 홈랩 PostgreSQL 통합 테스트로 repository/issue/ci_run list query를 검증했다.
- metrics/infra/risk/CI log snapshot handler가 `SnapshotProvider` 경계를 통해 data source를 읽도록 분리했고, 기본 `StaticSnapshotProvider`를 추가했다.
- dashboard metrics, infra nodes/edges/topology, risk snapshot 응답 meta에 provider `source`를 포함하도록 정리했다.
- `RuntimeSnapshotProvider`를 추가해 `DB_URL`, `GITEA_URL`, `BACKEND_AI_URL` 기반 health check로 infra node/edge status를 보강했다.
- Docker Compose backend-core 환경에 `BACKEND_AI_URL=http://backend-ai:8000`을 추가했다.
- `GET /api/v1/risks` DB-backed 조회 API를 추가했고, `GET /api/v1/risks/critical`은 action-required high risk가 있으면 DB 응답을 우선 사용한다.
- Gitea `action_run/workflow_run` 실패 이벤트에서 `ci_failure:{ci_run_id}` risk를 생성해 `risks` 테이블에 upsert하도록 정규화 경로를 확장했다.
- command/audit migration을 추가했다: `commands`, `audit_logs`.
- `POST /api/v1/risks/{risk_id}/mitigations`는 risk 상태를 즉시 바꾸지 않고 `pending` command와 audit log를 생성하며, `idempotency_key` 재시도는 기존 command를 반환한다.
- `GET /api/v1/commands/{command_id}`로 command 상태를 조회할 수 있다.
- 홈랩 PostgreSQL `devhub` DB에 migration version 3 적용을 검증했다.

## 4. 다음 작업 큐

### P1

- metrics용 DB-backed provider 구현 범위를 확정한다.
- Gitea Runner 세부 상태 adapter 또는 Gitea REST client 연동 범위를 확정한다.
- service action command API와 command status transition worker 범위를 확정한다.

### P2

- Commit/push 이벤트의 commit 단위 정규화 필요성을 검토하고 별도 `commits` 테이블 도입 여부를 결정한다.
- `ServiceNode`, `CiRun`, `Risk`, `Command` 모델에 프론트 mock보다 부족한 식별자, owner, timestamp, status, pagination/filtering 기준을 추가한다.
- WebSocket 메시지 타입(`infra.node.updated`, `ci.run.updated`, `risk.updated`, `command.status.updated`, `notification.created`) 초안을 작성한다.
- system admin allowlist 또는 seed admin 기준을 API 계약과 함께 확정한다.

### P3

- Python AI gRPC 서버 구현과 Go Core client 연결을 시작한다.
- Hourly Pull reconciliation worker와 Gitea REST client를 설계한다.
- weekly report, AI Gardener suggestion, team load 산출 모델을 후속 기능으로 구체화한다.

### P1 — Phase 13 (계정/인증 1차)

- `accounts` migration 추가 (`user_id` UNIQUE FK, `login_id` UNIQUE, `password_hash`, `password_algo`, `status`, `failed_login_attempts`, `last_login_at`, `password_changed_at`).
- domain/store layer: 1:1 invariant 보장하는 `CreateAccount`, `GetAccountByUser`, `UpdateAccount`, `ChangePassword`, `DeleteAccount`. 비밀번호 해시는 bcrypt cost ≥ 12 또는 argon2id 중 한 가지 선택.
- HTTP handler: `backend_api_contract.md §11` 의 7개 endpoint. 비밀번호는 요청 본문에서 파싱 즉시 해시로 변환하고 평문 변수 수명을 최소화. 응답에 평문/해시 미포함.
- audit log: `account.created`, `account.disabled`, `account.password_changed`, `account.locked`, `auth.login.succeeded`, `auth.login.failed` 6종 기록.
- 세션/JWT: 1차 구현은 server-side session 또는 short-lived JWT 중 한 가지 선택 → 결정 결과를 `architecture.md` 6.2.3 에 기록.
- 핸들러 테스트: in-memory account store mock + bcrypt round-trip + 1:1 conflict + 잠금 임계치 테스트.

## 5. Blocked 항목

- 현재 백엔드 로드맵 진행을 막는 blocked 항목 없음.
- 참고: Docker daemon socket(`/Users/yklee/.colima/default/docker.sock`) 연결 실패는 남아 있으나, 홈랩 PostgreSQL로 migration 검증을 완료해 Phase 2 차단은 해제됨.

## 6. 진척 관리 방식

- 이 문서의 Phase 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로만 관리한다.
- 코드 변경이 포함된 Phase는 테스트 또는 실행 검증 결과를 남긴 뒤 `done`으로 전환한다.
- 세션 종료 전 `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, 최신 backlog에서 이 문서와 현재 Phase를 함께 갱신한다.
