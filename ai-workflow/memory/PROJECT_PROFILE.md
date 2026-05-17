# Project Profile (Workflow Memory)

- 문서 목적: DevHub 프로젝트의 공통 운영 기준, 실행/검증 명령, 문서 참조 경로를 정의한다.
- 범위: 프로젝트 개요, 문서 구조, 기본 명령, 검증 포인트, 예외 규칙
- 대상 독자: 개발자, 운영자, AI 에이전트, 온보딩 담당자
- 상태: active
- 최종 수정일: 2026-05-16
- 관련 문서: [공통 표준](../core/global_workflow_standard.md)

## 1. 프로젝트 개요

- 프로젝트명: DevHub Example
- 프로젝트 목적: 역할별 기본 진입 우선순위 대시보드와 AI 분석 도구를 포함한 통합 관리 플랫폼.
- 주요 도메인:
  - UI/UX: Glassmorphism, Dark/Light Mode, Dashboard Widgets
  - RBAC: Fine-grained Resource-Action matrix (11 resources)
  - DREQ: Development Request Intake with Auth Tokens & IP Filtering
  - PMO: Application/Project management delegation (pmo_manager role)
- 주요 이해관계자:
  - Developers (DREQ assignee, Repository/CI/Risk view)
  - Managers (Risk triage, Team load balancing, DREQ oversight)
  - PMO Managers (Application/Project lifecycle management)
  - System Admins (System control, RBAC policy, DREQ Token admin)

## 2. 문서 구조 (Path)

- 문서 위키 홈: README.md, docs/README.md
- 운영 문서 홈: ai-workflow/memory/<agent>/<branch>/
- 현재 작업 운영 문서:
  - Gemini (UI/RBAC): ai-workflow/memory/gemini/phase6/
  - Codex (Backend/Action): ai-workflow/memory/codex/service-action-command/
- 백로그 위치: ai-workflow/memory/<agent>/<branch>/backlog/
- 세션 인계 문서: ai-workflow/memory/<agent>/<branch>/session_handoff.md
- flat memory 위치: legacy fallback 및 공용 색인 전용
- 환경 기록 위치: ai-workflow/memory/environments/

## 3. 기본 명령 (Commands)

- 설치: `make setup` (Go, Python, NPM 의존성 설치)
- 로컬 실행: `make run` (docker-compose 기반 전체 실행) 또는 `cd frontend && npm run dev` (frontend 개별 실행)
- 빠른 테스트: `cd backend-core && go test ./...`
- 격리 테스트: `cd frontend && npm run lint`
- 실행 확인: `make build`

## 4. 검증 포인트 (Validation)

- 코드 변경: PR 생성 전 로컬 테스트 통과 필수, Protobuf 변경 시 `make proto` 실행 필수
- 문서 변경: `PYTHONPATH=ai-workflow <bundled-python> ai-workflow/tests/check_docs.py` 또는 동등한 문서 검증 통과, 상대 경로 정합성 확인
- UI 변경: 브라우저 도구를 이용한 다크모드 및 Glassmorphism 레이아웃 깨짐 확인
- 배포/운영: `docker-compose build` 성공 여부 확인

## 5. 예외 규칙 (Policy)

- 병합: 브랜치별 워크플로우 상태 문서(`state.json`) 충돌 시 해당 브랜치의 최신 백로그 내용을 우선함
- 승인: `proto/` 디렉토리 변경 시 백엔드/프론트엔드 담당자 동시 승인 권장
- 제약: 로컬 개발 시 Docker Desktop 또는 호환되는 컨테이너 환경 필요
- 기타: Next.js frontend는 `app` 디렉토리 구조(App Router)를 따름

## 다음에 읽을 문서

- [Gemini 세션 인계 문서](./gemini/phase6/session_handoff.md)
- [Codex 세션 인계 문서](./codex/service-action-command/session_handoff.md)
