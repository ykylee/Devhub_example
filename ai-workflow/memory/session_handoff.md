# 세션 인계 문서

- 목적: 세션 상태 복원용 요약
- 상태: in_progress (2026-05-04)
- 관련: [Project Profile](./PROJECT_PROFILE.md), [Work Backlog](./work_backlog.md), [백엔드 개발 로드맵](./backend_development_roadmap.md)

## 1. 현재 작업 요약

- 기준선: main의 `gemini/frontend_phase1` 머지 내용을 `codex/backend_init`에 반영 중. workflow 운영 경로는 `ai-workflow/project/`에서 `ai-workflow/memory/`로 이동한 main 기준을 따른다.
- 백엔드 기준선: Go Core raw Webhook 수집 기반, `GET /api/v1/events` 조회 API 초안, 홈랩 PostgreSQL `webhook_events` migration version 1 및 도메인 정규화 migration version 2 적용 검증 완료.
- 문서 기준선: 프론트엔드가 작성한 백엔드 요구사항의 상세 리뷰를 `docs/backend/requirements_review.md`에 추가했고, gRPC/WebSocket 경계와 관리자 command 계약 보완 필요 사항을 정리했다.
- 프론트 연동 기준선: frontend phase1 화면과 mock/service layer를 기준으로 `docs/backend/frontend_integration_requirements.md`를 추가했고, `docs/backend/requirements_review.md`의 P1/P2 finding까지 포함해 백엔드 로드맵을 REST snapshot API, command/audit, WebSocket event 우선순위로 갱신했다.
- 구현 기준선: static fallback 기반 프론트 snapshot API 1차(`dashboard/metrics`, `infra/*`, `ci-runs`, `risks/critical`)를 Go Core에 추가했고, 도메인 정규화 1차 migration, `backend-core/internal/normalize` fixture 기반 processor, Postgres domain upsert sink, webhook 저장 이후 processor 실행 경로, repository/issue/PR/CI run/risk DB-backed 조회 API, snapshot provider 교체 경계, runtime infra health provider, CI 실패 이벤트 기반 risk 생성 경로, command/audit migration, risk mitigation command 생성/조회 API를 추가했다. `cd backend-core && go test ./...` 및 홈랩 DB 통합 테스트 통과를 확인했다.
- 주 작업 축: Gitea 연동 구현 및 frontend_phase1 병합 반영.
- 핵심 문서: docs/requirements.md, docs/architecture.md, docs/tech_stack.md, docs/backend_api_contract.md, ai-workflow/memory/backend_development_roadmap.md, ai-workflow/memory/backlog/2026-05-04.md.

## 2. 작업 상태 (State)

- 진행 중 (In Progress): TASK-007 Gitea Webhook 수신부 및 데이터 모델링 구현, Phase 5 프론트 snapshot API data source 교체
- 대기 중 (Pending): service action command API, command status transition worker, Gitea Runner 세부 상태 adapter, metrics DB-backed provider, WebSocket publish, Hourly Pull reconciliation, Python AI gRPC 연동
- 차단됨 (Blocked): 없음
- 최근 완료 (Done): TASK-019 command/audit 최소 schema 및 risk mitigation command API 구현, TASK-018 risk DB-backed 조회 및 CI 실패 risk 생성 경로 구현, TASK-017 Runtime infra snapshot provider 구현, TASK-016 metrics/infra snapshot provider 경계 분리, TASK-015 DB-backed 도메인 조회 API 1차 구현, TASK-014 도메인 정규화 sink 및 webhook 처리 경로 연결, TASK-013 도메인 정규화 1차 기반 구현, TASK-012 프론트 snapshot API 1차 구현, TASK-011 프론트엔드 현재 구현 기반 백엔드 연동 요구사항 및 로드맵 갱신, TASK-010 프론트엔드 백엔드 요구사항 상세 리뷰 문서화, TASK-009 핵심 개발 문서 집중 리뷰, TASK-008 개발 문서 리뷰 결과 정리 및 수정, TASK-006 프로젝트 초기 스캐폴딩 완료

## 3. 잔여 작업 우선순위

### P1 (즉시 실행)
- service action command API 구현
- command status transition worker 또는 executor boundary 설계
- Gitea Runner 세부 상태 adapter 또는 Gitea REST client 연동 범위 확정
- metrics DB-backed provider 구현 범위 확정
- main에서 병합된 frontend service layer와 백엔드 API 계약 간 naming/response shape 정합성 확인

### P2 (차순위)
- WebSocket 실시간 이벤트 publish 구현
- gRPC 프로토콜 기반 Go-Python 연동 테스트

## 4. 환경 및 검증

- 검증 호스트: darwin / local
- 완료 검증: `cd backend-core && go test ./...`, `DEVHUB_TEST_DB_URL=... go test ./internal/store -run TestPostgresStoreProcessesNormalizedWebhookEvent -count=1 -v`, 홈랩 PostgreSQL migration version `3` 확인, 기존 `cd frontend && npm run lint` 통과 기록 유지.
- 환경 제약: Docker daemon 미실행으로 Docker Compose 기반 통합 검증은 미실행. 홈랩 PostgreSQL migration 검증은 완료. 전체 검증에는 `protoc`, Node 의존성 설치, Docker daemon 실행 또는 접근 가능한 PostgreSQL 환경이 필요함.

## 다음에 읽을 문서
- [아키텍처 설계서](../../docs/architecture.md)
- [기술 스택 문서](../../docs/tech_stack.md)
- [백엔드 API 계약](../../docs/backend_api_contract.md)
- [백엔드 개발 로드맵](./backend_development_roadmap.md)
- [홈랩 PostgreSQL 환경 기록](./environments/homelab-postgresql.md)
- [작업 백로그](./backlog/2026-05-04.md)
