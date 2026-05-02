# Repository Assessment

- 문서 목적: 기존 프로젝트에 표준 AI 워크플로우를 도입하기 전에 현재 코드베이스와 문서 구조를 빠르게 진단한다.
- 범위: 저장소 구조, 추정 기술 스택, 문서 위치, 테스트 흔적, 초기 워크플로우 도입 포인트
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-04-28
- 관련 문서: `../ai-workflow/memory/PROJECT_PROFILE.md`, `../ai-workflow/memory/session_handoff.md`, `../ai-workflow/core/workflow_adoption_entrypoints.md`

## 1. 요약

- 분석 대상 프로젝트:
- `DevHub (Multi-Role Team Hub)`
- 분석 모드:
- `existing`
- 추정 기본 스택:
- `Go / Python / Next.js / PostgreSQL / Docker Compose`
- 감지된 스택 라벨:
- `backend-core: Go Gin`, `backend-ai: Python FastAPI`, `frontend: Next.js App Router`, `proto: gRPC/Protocol Buffers`, `database: PostgreSQL`

## 2. 저장소 구조 관찰

- 상위 디렉터리 항목:
- `.git, README.md, AGENTS.md, Makefile, docker-compose.yml, backend-core, backend-ai, frontend, proto, docs, ai-workflow`
- 소스 디렉터리 후보:
- `backend-core/`, `backend-ai/`, `frontend/`, `proto/`
- 문서 디렉터리 후보:
- `docs/`, `ai-workflow/memory/`
- 테스트 디렉터리 후보:
- 현재 명시적 테스트 디렉터리는 제한적이며, Go package test와 frontend lint를 우선 검증 기준으로 사용함.

## 3. 추정 명령

- 설치:
- `make init`
- 로컬 실행:
- `make run`
- 빠른 테스트:
- `cd backend-core && go test ./...`
- 격리 테스트:
- `pytest ai-workflow/tests/ && cd frontend && npm run lint`
- 실행 확인:
- `make build`

## 4. package script 및 경로 샘플

- package script 목록:
- `frontend/package.json` 기준 `dev`, `build`, `start`, `lint`
- 분석 중 확인한 경로 샘플:
- `README.md`, `Makefile`, `docker-compose.yml`, `backend-core/main.go`, `backend-ai/main.py`, `frontend/app/page.tsx`, `proto/analysis.proto`, `docs/architecture.md`, `docs/tech_stack.md`

## 5. 워크플로우 도입 초안

- 추천 문서 위키 홈:
- `docs/README.md`
- 추천 운영 문서 위치:
- `ai-workflow/memory/`
- 추천 backlog 위치:
- `ai-workflow/memory/backlog/`
- 추천 session handoff 위치:
- `ai-workflow/memory/session_handoff.md`

## 6. 자동 분석 기반 다음 작업

- `TASK-007`에서 Gitea Webhook 수신 API와 PostgreSQL 데이터 모델을 구현한다.
- gRPC proto 생성물과 Python gRPC 서버(`50051`) 구현 범위를 별도 작업으로 추적한다.
- Docker daemon, Node 의존성, protoc 설치 여부에 따라 `make init`, `make build`, frontend lint 검증을 순차 실행한다.

## 다음에 읽을 문서

- 프로젝트 프로파일: [./PROJECT_PROFILE.md](../ai-workflow/memory/PROJECT_PROFILE.md)
- 세션 인계 문서: [./session_handoff.md](../ai-workflow/memory/session_handoff.md)
- 도입 분기 가이드: [../ai-workflow/core/workflow_adoption_entrypoints.md](../ai-workflow/core/workflow_adoption_entrypoints.md)
