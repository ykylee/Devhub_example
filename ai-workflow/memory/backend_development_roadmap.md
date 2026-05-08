# 백엔드 개발 로드맵

- 문서 목적: DevHub 백엔드 구현 범위, 순서, 진척 상태를 추적한다.
- 범위: backend-core phase 로드맵, 완료 범위, 다음 작업 큐, 차단 항목
- 대상 독자: 백엔드 개발자, 프론트엔드 연동 담당자, AI agent
- 기준일: 2026-05-07
- 상태: in_progress
- 최종 수정일: 2026-05-07
- 관련 문서: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`, `docs/backend_api_contract.md`, `docs/backend/frontend_integration_requirements.md`, `docs/backend/requirements_review.md`, `docs/adr/0001-idp-selection.md`, `ai-workflow/memory/codex/service-action-command/session_handoff.md`
- 현재 브랜치: `codex/service-action-command`
- 현재 기준선: `origin/main@7316ddd` 반영 완료. Phase 12 조직/사용자 CRUD, Phase 13 Ory Hydra/Kratos PoC scaffold, TASK-020~023 command/realtime worker가 같은 브랜치에 공존한다.

## 1. 개발 원칙

- Go Core는 Gitea Webhook/API 연동, 권한 관리, 데이터 저장, 시스템 관리자 기능의 중심 서비스로 둔다.
- Python AI는 초기 단계에서 PostgreSQL에 직접 접근하지 않고, Go Core가 필터링한 입력을 gRPC로 전달받는다.
- 프론트 대상 실시간 API는 gRPC stream이 아니라 REST snapshot + WebSocket event로만 계약한다.
- 프론트엔드 UI 표시명과 API wire format은 분리한다. 역할 값은 API에서 `developer`, `manager`, `system_admin`을 기본으로 한다.
- 명령성 액션은 boolean 결과가 아니라 `command_id`와 command status lifecycle로 관리하고 audit log를 남긴다.
- 운영 actor는 최종적으로 `X-Devhub-Actor` header가 아니라 Hydra/Kratos 기반 세션 또는 JWT claim에서 도출한다.
- 조직/사용자 도메인의 master data는 DevHub `users`/`org_units`가 담당하고, credential/session master는 Kratos가 담당한다.
- 검증하지 않은 단계는 `done`으로 전환하지 않는다.
- 세션 상태 문서는 브랜치별 memory 경로(`ai-workflow/memory/<agent>/<branch>/`)를 source of truth로 사용한다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 산출물 | 다음 판정 기준 |
| --- | --- | --- | --- | --- |
| Phase 1 | done | Go Core 기반 구조 정리 | `internal/config`, `internal/httpapi`, `internal/gitea`, `internal/store` 분리 | `cd backend-core && go test ./...` |
| Phase 2 | done | PostgreSQL 초기 스키마 | `webhook_events` migration | migration 적용 검증 |
| Phase 3 | done | Gitea Webhook raw 수신부 | `POST /api/v1/integrations/gitea/webhooks`, signature 검증, dedupe 처리 | handler 단위 테스트 |
| Phase 4 | done | 프론트 연동 계약 안정화 1차 | role wire format, REST snapshot/WebSocket envelope, integration requirements | API 계약 문서화 및 smoke/lint 통과 |
| Phase 5 | done | 프론트 snapshot API 1차 | metrics, infra topology, ci-runs/logs, risk 조회 API, runtime snapshot provider | handler 테스트 및 fallback 동작 확인 |
| Phase 6 | done | 도메인 정규화 1차 | repository/user/issue/pull_request/ci_run/risk 기초 테이블 및 normalize processor | fixture 및 store 테스트 |
| Phase 7 | in_progress | command/audit 기반 액션 API | service action, risk mitigation, command status, idempotency, audit log | approval/executor boundary, audit 조회 API, actor 검증 |
| Phase 8 | in_progress | WebSocket 실시간 채널 | `/api/v1/realtime/ws`, `command.status.updated` publish, `types` subscription/RBAC filter | replay, infra/ci/risk event publish, resource scope filter |
| Phase 9 | planned | Python AI gRPC 연결 | Go gRPC client, Python `AnalysisService`, build log summary/risk detection | gRPC 통합 테스트 |
| Phase 10 | planned | Hourly Pull Reconciliation | Gitea REST client, 누락 이벤트 보정 worker | dry-run 및 idempotency 테스트 |
| Phase 11 | planned | 시스템 관리자 기능 고도화 | Runner/server adapter, config 조회, allowlist/seed admin | 권한/audit/health adapter 테스트 |
| Phase 12 | done | 조직/사용자 관리 API | `users`, `org_units`, appointments, hierarchy, unit members CRUD | handler/store 테스트 및 프론트 연동 |
| Phase 13 | in_progress | Ory Hydra/Kratos IdP 도입 | ADR-0001, IdP PoC scaffold, schema/config/client 등록 | Go Core token middleware, identity admin wrapper, audit mapping |

## 3. 현재 완료 범위

- Gitea webhook raw 저장, signature 검증, dedupe 처리, event 조회 API를 구현했다.
- repository/user/issue/pull_request/ci_run/risk 정규화 테이블과 normalize processor를 구현했다.
- DB-backed domain 조회 API를 구현했다: repositories, issues, pull-requests, ci-runs, risks.
- snapshot handler를 `SnapshotProvider` 경계로 분리하고 runtime/static fallback을 제공한다.
- command/audit migration을 추가했다: `commands`, `audit_logs`.
- `POST /api/v1/admin/service-actions`, `POST /api/v1/risks/{risk_id}/mitigations`, `GET /api/v1/commands/{command_id}`를 구현했다.
- idempotency replay와 command 조회 테스트를 추가했다.
- 승인 불필요 dry-run command worker를 추가해 `pending -> running -> succeeded`로 자동 전이한다.
- `/api/v1/realtime/ws`와 in-process `RealtimeHub`를 추가했고 `command.status.updated`를 publish한다.
- WebSocket `types` query 기반 subscription filtering과 event type별 RBAC read permission check 1차 구현을 추가했다.
- Phase 12 조직/사용자 CRUD API를 구현했다: users CRUD, org unit CRUD, hierarchy, unit members replace/list.
- `GET /api/v1/audit-logs`를 추가했고 조직/사용자 CRUD와 멤버 교체에 audit log 생성을 연결했다.
- `X-Devhub-Actor` 사용 시 deprecation 응답 헤더를 추가해 Phase 13 token actor 전환 경로를 노출했다.
- `docs/backend_api_contract.md` §11을 Hydra/Kratos 기준으로 재작성하고 Go Core Bearer token verifier 경계를 추가했다.
- `GET /api/v1/rbac/policy`를 추가하고 프론트 Permissions 화면이 backend policy를 조회하도록 준비했다.
- RBAC policy version table과 `PUT /api/v1/rbac/policy`를 추가해 전체 matrix 교체와 audit log 기록 경계를 만들었다.
- `PUT /api/v1/rbac/policy`에 `system_config: admin` RBAC enforcement를 적용했다.
- `GET /api/v1/me`를 추가해 인증 actor를 DevHub `users`와 매핑하고 effective permissions를 반환한다.
- service action, risk mitigation, audit 조회, 조직/사용자 쓰기 API에 RBAC enforcement를 적용했다.
- Phase 13 Ory Hydra/Kratos PoC scaffold가 main에 반영됐다: `infra/idp/`, schema/config/client 등록 관련 파일.
- 브랜치별 memory 구조를 적용해 현재 브랜치 상태 문서는 `ai-workflow/memory/codex/service-action-command/` 아래에서 관리한다.

## 4. 재검토 결과

### 방향성 충돌 없음

- Phase 12 조직/사용자 관리와 현재 command/realtime 작업은 충돌하지 않는다. 둘 다 `backend-core` 라우터와 Postgres store에 공존 가능하다.
- Phase 13 IdP 도입은 현재 command actor 처리 방식과 직접 연결된다. 기존 `X-Devhub-Actor`는 개발용 임시 경계로 유지하고, production 경계는 JWT/session claim 기반으로 전환해야 한다.
- service action command의 dry-run 자동 성공 전이는 “실제 executor 도입 전 안전한 시뮬레이션”으로 유지 가능하다.

### 조정이 필요한 전제

- WebSocket은 endpoint, command event, `types` subscription/RBAC filter 1차까지 구현됐다. Phase 8은 done이 아니라 “command event와 역할 필터 1차 완료, replay와 resource scope filter 미완료” 상태다.
- Command/Audit는 command 생성, 상태 전이, audit 조회, approval boundary 1차까지 구현됐다. 실제 executor adapter는 아직 없으므로 Phase 7은 계속 in_progress다.
- Phase 13 계정/인증 계약은 자체 accounts table 전제를 폐기하고 Hydra/Kratos 기준으로 재작성했다. 실제 JWKS/introspection verifier와 admin identity wrapper 구현은 남아 있다.
- Docker 기반 실행 전제와 Phase 13 native binary 운영 전제가 문서마다 섞여 있다. 로컬 검증 명령은 당분간 Go/NPM/native PostgreSQL 중심으로 유지하고, Docker Compose는 호환 폴백으로 낮춘다.

## 5. 우선순위 계획

### P0: 통합 안정화

- `docs/backend_api_contract.md` §11 Hydra/Kratos 재작성은 완료했다.
- Go Core actor 추출 경계는 1차 완료했다: `X-Devhub-Actor` fallback deprecation, Bearer token verifier interface, 검증된 actor context 연결, `/me` user-role lookup까지 구현했다. 다음은 Hydra JWKS/introspection verifier다.
- 인증 actor가 DevHub user에 매핑되지 않거나 비활성 상태인 경우 role fallback으로 우회하지 않도록 RBAC actor 경계를 강화한다.
- RBAC policy 조회/교체 API, persistence, `/me`, 주요 쓰기 API enforcement는 1차 완료했다. 다음은 편집 UI 활성화 전 confirmation UX와 approval 필요 여부 확정이다.
- WebSocket 인증/구독 필터 1차 구현과 publish lock 개선은 완료했다. 다음은 replay와 resource/project scope filter다.
- audit log 조회 API와 조직/사용자 CRUD audit 연결은 1차 완료됐다. 후속으로 auth actor와 source_ip/request_id 보강이 필요하다.

### P1: Admin Action 실행 경계

- service action approval model 1차를 구현했다: `requires_approval=true` pending command를 승인하면 executor 후보, 거절하면 `rejected`로 종료한다.
- 승인된 live service action만 조회하는 query와 `ServiceActionExecutor` adapter interface/worker 경계를 추가했다.
- `SERVICE_ACTION_EXECUTOR_MODE=simulation`에서만 켜지는 simulation executor를 추가했다. service/action allowlist를 모두 통과한 command만 외부 side effect 없이 성공 처리한다. 운영용 실제 side-effect adapter는 후속 범위다.
- Gitea Runner/server 상태 adapter 범위를 확정한다.
- live command는 기본 거절 또는 approval required로 유지하고, dry-run과 실제 side effect 경계를 테스트로 고정한다.

### P2: Realtime 확장

- `command.status.updated`를 프론트 toast/status UI와 연결한 뒤 event payload 안정성을 확인한다.
- `infra.node.updated`, `ci.run.updated`, `risk.updated`, `notification.created` publish 경계를 구현한다.
- WebSocket hub publish 경로는 client snapshot 후 hub lock 밖에서 write하고 실패 client를 제거하도록 1차 개선했다. 장기적으로 backpressure가 필요하면 client별 bounded send queue를 추가한다.
- WebSocket reconnect/replay 전략을 정한다.
- resource/project scope 기반 subscription filtering을 구현한다.

### P3: Gitea REST 및 AI

- Gitea REST client와 hourly reconciliation worker를 설계한다.
- commit 단위 정규화 필요성을 검토하고 `commits` 테이블 도입 여부를 결정한다.
- Python AI gRPC 서버와 Go Core client 연결을 시작한다.
- AI Gardener suggestion 입력/출력 모델을 command/audit 및 risk 모델과 연결한다.

## 6. 다음 작업 큐

- [x] API 계약 §11 Hydra/Kratos 재작성
- [x] Bearer token 검증 middleware 설계 및 최소 구현
- [x] RBAC policy 조회 API 및 프론트 Permissions 연동 준비
- [x] RBAC policy persistence/edit API와 audit 경계
- [x] RBAC policy edit enforcement (`system_config: admin`)
- [x] `GET /api/v1/me` 및 DevHub user-role lookup
- [x] service action/risk/audit/organization RBAC enforcement
- [x] 인증 actor 미매핑/비활성 시 role fallback 우회 차단
- [x] `X-Devhub-Actor` deprecation warning 경로 추가
- [x] audit log 조회 API와 organization CRUD audit 연결
- [x] WebSocket 인증/구독 필터 1차 구현
- [x] WebSocket publish lock 개선
- [x] service action approval/reject API 및 audit boundary
- [x] approved live service action query 및 executor adapter boundary
- [x] simulation service action executor 및 명시적 main 주입 설정
- [ ] WebSocket replay 및 resource scope filter 설계
- [ ] service action 운영 executor adapter 구현 범위 확정
- [ ] Gitea Runner adapter 범위 확정
- [ ] AI Gardener suggestion API/UI 연결 범위 확정

## 7. Blocked 항목

- 현재 백엔드 코드 진행을 막는 hard blocker는 없다.
- Phase 13 실제 round-trip 검증은 Hydra/Kratos native binary, PostgreSQL schema, frontend auth route 준비가 필요하다.
- 외부 네트워크/사내 SSL inspection 환경에서는 Go module, npm package, font 다운로드가 막힐 수 있으므로 mirror 또는 사내 CA 설정을 사용한다.

## 8. 진척 관리 방식

- 이 문서의 Phase 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로만 관리한다.
- 코드 변경이 포함된 Phase는 테스트 또는 실행 검증 결과를 남긴 뒤 `done`으로 전환한다.
- 세션 종료 전 현재 브랜치별 `ai-workflow/memory/<agent>/<branch>/state.json`, `session_handoff.md`, 최신 backlog에서 이 문서와 현재 Phase를 함께 갱신한다.
