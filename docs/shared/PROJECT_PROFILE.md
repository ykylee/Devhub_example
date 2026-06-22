# Project Workflow Profile

- 문서 목적: 프로젝트 특화 규칙과 실행/검증 기준을 정의한다.
- 범위: 프로젝트 개요, 문서 구조, 기본 명령, 검증 포인트, 예외 규칙
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: accepted
- 최종 수정일: 2026-05-20
- 관련 문서: [공통 표준](../ai-workflow/core/global_workflow_standard.md)

이 문서는 `ai-workflow/memory/PROJECT_PROFILE.md`와 같은 운영 기준을 docs 트리에서 참조할 수 있도록 유지하는 mirror 문서다.

## 1. 프로젝트 개요

- 프로젝트명: DevHub Example
- 프로젝트 목적: 역할별 기본 진입 우선순위 기반 대시보드와 AI 기반 분석 도구를 포함한 통합 관리 플랫폼.
- 주요 이해관계자:
  - Developers (개발 대시보드 기본 진입 및 API 사용)
  - Managers (관리 대시보드 기본 진입 기반 프로젝트/리스크 관리)
  - System Admins (시스템 대시보드/시스템 설정 접근, 인프라 및 권한 제어)

## 2. 문서 구조 (Path)

- 문서 위키 홈: `README.md`, `docs/README.md`
- 운영 문서 위치 (브랜치별 분리): `ai-workflow/memory/<agent>/<branch>/`
- source-of-truth 결정 규칙 (CLAUDE.md 정합):
  - **sprint 브랜치 작업 시** → `ai-workflow/memory/<agent>/<branch>/{state.json,session_handoff.md,work_backlog.md}` 가 활성 source-of-truth
  - **main 브랜치 작업 시** → flat 경로 (`ai-workflow/memory/state.json` 등) 가 활성 source-of-truth
  - flat 경로는 main HEAD 동기화 + sprint 브랜치 디렉터리 없을 때 fallback. 두 위치 모두 존재하면 브랜치 디렉터리가 우선
- 백로그 위치: `ai-workflow/memory/backlog/` (main) 또는 `ai-workflow/memory/<agent>/<branch>/backlog/` (sprint)
- 세션 인계 문서: `ai-workflow/memory/session_handoff.md` (main) 또는 `<agent>/<branch>/session_handoff.md` (sprint)
- 환경 기록 위치: `ai-workflow/memory/environments/`

## 3. 기본 명령 (Commands)

- 설치: `make setup` (Go, Python, NPM 의존성 일괄 설치 — docker 비의존)
- 로컬 실행 (native default): 모드별 절차는 [`docs/setup/environment-setup.md`](./setup/environment-setup.md) 참조 — backend-core (`go run .`), frontend (`npm run dev`). **2026-06-22 M-v0.2.2 부터 backend-ai 제외** (PR #663).
- 빠른 테스트: `cd backend-core && go test ./...`, `cd frontend && npm run test`
- 격리 검증: `cd backend-core && go vet ./...`, `cd frontend && npm run lint`
- 빌드: 모드별 — `make build` 는 환경별 절차 안내만 출력하므로, native 는 `(cd backend-core && go build ./...) && (cd frontend && npm run build)` 직접 호출

## 4. 검증 포인트 (Validation)

- 코드 변경: PR 생성 전 로컬 테스트 통과 필수, Protobuf 변경 시 `make proto` 실행 필수
- 문서 변경: `ai-workflow/tests/check_docs.py` 또는 동등한 문서 검증 통과, 상대 경로 정합성 확인
- UI 변경: native dev 서버에서 브라우저 검증 (다크 / 라이트 모드, 반응형 layout, semantic theme 회귀)
- 배포 검증: native 모드는 헬스 엔드포인트 (`curl http://localhost:8080/health`, `:8000/health`); docker 모드는 사용자 로컬 자산 위에서 `docker-compose ps`

## 5. 예외 규칙 (Policy)

- 병합: 브랜치별 워크플로우 상태 문서(`state.json`) 충돌 시 해당 브랜치의 최신 백로그 내용을 우선함
- 승인: `proto/` 디렉토리 변경 시 백엔드/프론트엔드 담당자 동시 승인 권장
- 제약: 로컬 개발은 native (no-docker) default — 컨테이너 자산은 환경별로 git 추적 외부에서 관리하며, 본 저장소에는 환경 구성 가이드(`docs/setup/environment-setup.md`)만 둔다 (`.gitignore` 의 `DEV ENVIRONMENT` 섹션 참조). docker 사용 자체는 환경 정책에 따라 선택 가능.
- 인증: Keycloak OIDC 단일 IdP ([ADR-0019](./adr/0019-keycloak-only-idp.md)). 자체 `/api/v1/auth/*` proxy 와 Hydra+Kratos 흐름 ([ADR-0001](./adr/0001-idp-selection.md), superseded) 은 historical reference.
- 기타: Next.js frontend는 `app` 디렉토리 구조(App Router)를 따름

## 다음에 읽을 문서

- [세션 인계 (main)](../ai-workflow/memory/session_handoff.md)
- [작업 백로그 (main)](../ai-workflow/memory/work_backlog.md)
- [환경 구성 가이드](./setup/environment-setup.md)
