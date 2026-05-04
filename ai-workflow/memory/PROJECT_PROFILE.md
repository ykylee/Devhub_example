# Project Workflow Profile

- 문서 목적: 프로젝트 특화 규칙과 실행/검증 기준을 정의한다.
- 범위: 프로젝트 개요, 문서 구조, 기본 명령, 검증 포인트, 예외 규칙
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-05-03
- 관련 문서: [공통 표준](../core/global_workflow_standard.md)

## 1. 프로젝트 개요
- 프로젝트명: DevHub (Multi-Role Team Hub)
- 프로젝트 목적: 역할별 개발 허브를 구축해 Developer, Manager, System Admin이 Gitea 이벤트, CI 상태, 리스크, 인프라 상태를 한 화면에서 확인하고 조치할 수 있게 한다.
- 주요 이해관계자: 개발자, 매니저, 시스템 관리자, 저장소 운영자, AI agent 협업자

## 2. 문서 구조 (Path)
- 문서 위키 홈: README.md
- 운영 문서 홈: ai-workflow/memory/
- 백로그 위치: ai-workflow/memory/backlog/
- 세션 인계 문서: ai-workflow/memory/session_handoff.md
- 환경 기록 위치: ai-workflow/memory/environments/

## 3. 기본 명령 (Commands)
- 설치: `make setup`
- 로컬 실행: `cd backend-core && go run .`
- 빠른 테스트: `cd backend-core && go test ./...`
- 격리 테스트: `cd frontend && npm run lint`
- 실행 확인: `make build`

## 4. 검증 포인트 (Validation)
- 코드 변경: 영향 범위에 맞는 Go/프론트 테스트를 실행하고, 실행하지 못한 검증은 사유를 handoff/backlog에 남긴다.
- 문서 변경: 링크 경로와 `state.json` JSON 유효성을 확인한다.
- UI 변경: 로컬 dev server 또는 브라우저 확인으로 주요 화면을 검증한다.
- 배포/운영: DB migration은 대상 환경과 migration version을 기록하고, Docker 기반 검증은 Docker daemon 접근 가능 여부를 함께 남긴다.

## 5. 예외 규칙 (Policy)
- 병합: 상태 문서 충돌 시 `state.json`, `session_handoff.md`, 최신 backlog, 로드맵 순서로 현재 세션 기준선을 재확정한다.
- 승인: 위험한 인프라 제어, credential 변경, destructive git 작업은 사용자 확인 후 진행한다.
- 제약: Docker daemon, protoc, Node 의존성, 홈랩 PostgreSQL 접근 가능 여부가 전체 검증 범위를 좌우한다.
- 기타: 사용자 보고와 workflow 문서 갱신 문안은 기본적으로 한국어로 작성한다.

## 다음에 읽을 문서
- [세션 인계 문서](./session_handoff.md)
- [작업 백로그](./work_backlog.md)
