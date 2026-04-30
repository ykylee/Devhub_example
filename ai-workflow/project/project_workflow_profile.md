# Project Workflow Profile

- 목적: 프로젝트 특화 규칙 정의
- 관련: [공통 표준](../core/global_workflow_standard.md)

## 1. 프로젝트 개요
- 프로젝트명: DevHub (Multi-Role Team Hub)
- 프로젝트 목적: 역할군별(개발자, 관리자 등) 맞춤형 UI를 제공하여 팀의 협업 효율을 극대화하며, 새로운 역할군의 추가가 용이한 확장성 있는 구조를 지향함.
- 주요 이해관계자: 개발자, PM, PMO, 향후 추가될 QA 및 유관 부서.

## 2. 문서 구조 (Path)
- 문서 위키 홈: docs/README.md
- 운영 문서 홈: ai-workflow/project/
- 백로그 위치: ai-workflow/project/backlog/
- 세션 인계 문서: ai-workflow/project/session_handoff.md
- 환경 기록 위치: ai-workflow/project/environments/

## 3. 기본 명령 (Commands)
- 설치: `make init`
- 로컬 실행: `make run`
- 빠른 테스트: `cd backend-core && go test ./...`
- 격리 테스트: `pytest ai-workflow/tests/ && cd frontend && npm run lint`
- 실행 확인: `make build`

## 4. 검증 포인트 (Validation)
- 코드 변경: 관련 서비스 테스트 및 `make build` 확인. Go 변경은 `cd backend-core && go test ./...` 실행.
- 문서 변경: workflow 상태 문서(`state.json`, `session_handoff.md`, 최신 backlog)와 링크 정합성 확인.
- UI 변경: `cd frontend && npm run lint` 및 브라우저 확인.
- 배포/운영: Docker Compose 구성과 환경 변수(`GITEA_URL`, `GITEA_TOKEN`) 확인 후 실행.

## 5. 예외 규칙 (Policy)
- 병합: 상태 문서 충돌 시 최신 backlog, `session_handoff.md`, `state.json` 순서로 대조해 동기화한다.
- 승인: 외부 시스템 연동, 인증/권한, 데이터 보존 정책 변경은 사용자 확인 후 진행한다.
- 제약: 전체 로컬 검증에는 `protoc`, Node 의존성, Docker daemon 실행이 필요하다.
- 기타: `ai-workflow/`는 workflow 메타 레이어로 취급하고, 프로젝트 코드 탐색 기본 범위에서는 제외한다.

## 다음에 읽을 문서
- [세션 인계 문서](./session_handoff.md)
- [작업 백로그](./work_backlog.md)
