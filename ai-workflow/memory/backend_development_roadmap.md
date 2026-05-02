# 백엔드 개발 로드맵

- 문서 목적: DevHub 백엔드 구현 범위, 순서, 진척 상태를 추적한다.
- 기준일: 2026-05-02
- 상태: in_progress
- 관련 문서: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`, `docs/backend_api_contract.md`, `ai-workflow/memory/backlog/2026-05-02.md`
- 현재 브랜치: `codex/backend_init`
- 프론트엔드 전제: 프론트엔드는 별도 브랜치에서 개발 후 병합 예정이므로, 백엔드는 API 계약과 실시간 이벤트 계약을 먼저 안정화한다.

## 1. 개발 원칙

- Go Core는 Gitea Webhook/API 연동, 권한 관리, 데이터 저장, 시스템 관리자 기능의 중심 서비스로 둔다.
- Python AI는 초기 단계에서 PostgreSQL에 직접 접근하지 않고, Go Core가 필터링한 입력을 gRPC로 전달받는다.
- 모든 Gitea Webhook 이벤트는 raw event로 먼저 저장하고, 정규화/분석/실시간 publish는 후속 단계에서 확장한다.
- 프론트엔드 병렬 개발을 위해 REST API 응답 형태와 WebSocket 메시지 타입을 변경 가능성이 낮은 계약으로 관리한다.
- 검증하지 않은 단계는 `done`으로 전환하지 않는다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 산출물 | 검증 기준 |
| --- | --- | --- | --- | --- |
| Phase 1 | done | Go Core 기반 구조 정리 | `internal/config`, `internal/httpapi`, `internal/gitea`, `internal/store` 분리 | `cd backend-core && go test ./...` |
| Phase 2 | done | PostgreSQL 초기 스키마 | `webhook_events` migration | migration 적용 검증 |
| Phase 3 | done | Gitea Webhook raw 수신부 | `POST /api/v1/integrations/gitea/webhooks`, signature 검증, dedupe 처리 | handler 단위 테스트 |
| Phase 4 | in_progress | 프론트 연동용 조회 API | `GET /api/v1/events`, repository/issue/PR 조회 API 초안 | API 테스트 및 응답 계약 문서화 |
| Phase 5 | planned | 도메인 정규화 1차 | repository/user/issue/pull_request/ci_run 테이블 및 정규화 로직 | fixture 기반 정규화 테스트 |
| Phase 6 | planned | WebSocket 실시간 채널 | `/api/v1/realtime/ws`, 이벤트 publish 메시지 | WebSocket 연결/메시지 테스트 |
| Phase 7 | planned | Python AI gRPC 연결 | Go gRPC client, Python `AnalysisService` server | gRPC 통합 테스트 |
| Phase 8 | planned | Hourly Pull Reconciliation | Gitea REST client, 누락 이벤트 보정 worker | dry-run 및 idempotency 테스트 |
| Phase 9 | planned | 시스템 관리자 기능 | Runner/서버 상태 조회, 제한된 제어 API, audit log | 권한/audit 테스트 |

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

## 4. 다음 작업 큐

### P1

- repository/issue/PR 조회 API 초안을 도메인 정규화 설계와 함께 확정한다.
- `webhook_events` 상태 전이 기준(`validated` 이후 `processed`, `failed`, `ignored`)을 정리한다.

### P2

- Issue, Pull Request, Commit, Actions 이벤트 fixture를 수집해 정규화 테스트 기반을 만든다.
- repository/user/issue/pull_request/ci_run 초기 테이블을 설계한다.
- WebSocket 메시지 타입 초안을 작성한다.

### P3

- Python AI gRPC 서버 구현과 Go Core client 연결을 시작한다.
- Hourly Pull reconciliation worker와 Gitea REST client를 설계한다.
- 시스템 관리자 기능의 allowlist 또는 seed admin 기준을 확정한다.

## 5. Blocked 항목

- 현재 백엔드 로드맵 진행을 막는 blocked 항목 없음.
- 참고: Docker daemon socket(`/Users/yklee/.colima/default/docker.sock`) 연결 실패는 남아 있으나, 홈랩 PostgreSQL로 migration 검증을 완료해 Phase 2 차단은 해제됨.

## 6. 진척 관리 방식

- 이 문서의 Phase 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로만 관리한다.
- 코드 변경이 포함된 Phase는 테스트 또는 실행 검증 결과를 남긴 뒤 `done`으로 전환한다.
- 세션 종료 전 `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, 최신 backlog에서 이 문서와 현재 Phase를 함께 갱신한다.
