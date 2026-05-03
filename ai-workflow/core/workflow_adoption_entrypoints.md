# Workflow Adoption Entry Points
- 문서 목적: 표준 AI 워크플로우를 도입할 때 `신규 프로젝트` 와 `작업 중인 프로젝트` 두 가지 진입 경로를 구분해 시작 절차를 정리한다.
- 범위: 도입 모드별 목표, 추천 시작 순서, 자동화 가능 범위, 주의점
- 대상 독자: 저장소 관리자, 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-04-23
- 관련 문서: `./global_workflow_standard.md`, `../WORKFLOW_INDEX.md`, `../README.md`, `../memory/project_status_assessment.md`
- 상태 문서/프로젝트 문서 경계: `ai-workflow/memory/`는 workflow 상태, `README.md`와 `docs/`는 프로젝트 문서로 구분한다.
## 1. 도입 경로 개요
1. 신규 프로젝트
2. 작업 중인 프로젝트
## 2. 신규 프로젝트 도입
### 목표
- 공통 문서 구조를 빠르게 깔고, 프로젝트 특화 규칙만 얇게 채운다.
- 문서 경로, 기본 명령, 검증 규칙을 초기부터 표준 포맷으로 맞춘다.
### 추천 시작 순서
1. 별도 workflow kit source bundle에서 기본 문서 세트를 생성하거나, 현재 저장소의 `ai-workflow/memory/` 구조를 복사해 시작한다.
2. `PROJECT_PROFILE.md` 에 프로젝트 목적, 문서 구조, 명령, 검증 규칙을 채운다.
3. `session_handoff.md` 와 날짜별 backlog 에 첫 작업 기준선을 적는다.
4. 자동 도구가 있으면 `state.json` 을 생성하고, 없으면 `session_handoff.md`와 backlog 기준으로 수동 갱신한다.
4. 이후 skill/MCP 도입 범위를 정한다.
### 자동화 포인트
- 문서 세트 생성
- core 문서 복사
- 첫 backlog 파일 생성
### 주의점
- 신규 프로젝트는 자동 추정할 기존 코드베이스가 없으므로, profile 문서의 TODO 를 반드시 사람이 채워야 한다.
- 운영 절차가 아직 없는 상태라면 profile 문서의 예외 규칙 섹션부터 먼저 정리하는 편이 좋다.
## 3. 작업 중인 프로젝트 도입
### 목표
- 기존 코드베이스와 문서 구조를 빠르게 읽고, 워크플로우용 문서 초안을 자동 생성한다.
- 현재 저장소 현실에 맞는 도입 경로를 제시해 "빈 템플릿" 상태를 줄인다.
### 추천 시작 순서
1. 별도 workflow kit source bundle의 existing onboarding 도구를 실행하거나, 현재 저장소 구조를 수동 분석한다.
2. `memory/project_status_assessment.md` 에 적힌 추정 스택, 명령, 문서 위치를 실제 운영 규칙과 대조한다.
3. `PROJECT_PROFILE.md` 에 자동 추정값을 확정 또는 수정한다.
4. `session_handoff.md` 와 backlog 에 현재 진행 중인 실제 작업과 리스크를 반영한다.
5. 자동 도구가 있으면 `state.json` 을 갱신하고, 없으면 JSON 유효성을 확인하며 수동 갱신한다.
6. 이후 문서 동기화, 세션 시작, backlog 갱신 흐름을 단계적으로 도입한다.
### bootstrap 직후 후속 루틴

현재 저장소에는 onboarding script가 포함되어 있지 않다. 자동 도구가 없는 경우 아래 순서를 수동 수행한다.

1. latest backlog 식별
2. session-start 기준선 복원
3. validation-plan 으로 초기 검증 수준 정리
4. code-index-update 로 README/허브/index 재확인 후보 정리

세부 자동화 계약은 별도 workflow kit source bundle에서 관리한다. 현재 저장소에서는 `WORKFLOW_INDEX.md`와 `memory/` 문서를 기준으로 수동 대체한다.
### 자동화 포인트
- 상위 디렉터리 구조 스캔
- 기술 스택 후보 감지
- docs/tests/source 디렉터리 감지
- 설치, 실행, 테스트 명령 추정
- project status assessment 문서 생성
- handoff/backlog 초기 항목 자동 작성
### 주의점
- 자동 추정한 명령은 편의용 초안일 뿐이며, 실제 CI/CD 또는 운영 절차와 다를 수 있다.
- 기존 문서 체계가 이미 있다면 별도 워크플로우 문서를 둘지, 기존 문서 위치에 흡수할지 먼저 결정해야 한다.
- 리뷰 규칙, 배포 승인 규칙, 환경 제약은 자동 추정하기 어렵기 때문에 반드시 사람이 보강해야 한다.
- 실제 파일럿 적용 대상은 프로젝트 관리자가 별도 체크리스트로 먼저 추리는 편이 좋다.
## 4. 어떤 경로를 고를지 판단 기준
- 아직 저장소 골격만 있는 경우:
- `new`
- 이미 코드와 테스트, 문서가 존재하는 경우:
- `existing`
- 코드가 조금 있지만 운영 규칙이 거의 없는 경우:
- `existing` 으로 시작하되, 신규 프로젝트처럼 profile 문서를 적극 보완한다.
## 5. 권장 출력물
- 신규 프로젝트:
- `PROJECT_PROFILE.md`, `session_handoff.md`, `work_backlog.md`, 날짜별 backlog
- 작업 중인 프로젝트:
- 위 문서 세트 + `project_status_assessment.md`
## 6. 이번 릴리즈 기준 권장 도입 묶음
### 6.1 첫 세션 기본 묶음
1. existing project onboarding 도구 또는 수동 repository assessment
2. `session-start`
3. `validation-plan`
4. `code-index-update`
### 6.2 작업 등록/문서 정렬 묶음
1. `backlog-update`
2. `doc-sync`
3. 필요 시 `merge-doc-reconcile`
4. `state.json` 자동 생성 또는 수동 갱신
### 6.3 다음 릴리즈로 미루는 항목
- 하네스 기본 연결 경로를 MCP server 로 승격하는 작업
- read-only MCP draft bridge 를 기본 운영 경로로 바꾸는 작업
- MCP capability 확장과 정식 client 상호운용 범위 확대
## 7. 현재 저장소 배포 메모 (2026-05-03)

- 현재 저장소에는 workflow source bundle이 아니라 runtime 문서 세트만 남긴다.
- `scripts/`, `skills/`, `mcp`, `templates`, `tests`, `examples`는 별도 배포물 또는 외부 도구로 취급한다.
- 온보딩은 `WORKFLOW_INDEX.md`에서 시작하고, 자동화가 없을 때는 core 문서를 수동 절차로 사용한다.

## 다음에 읽을 문서
- 공통 표준: [./global_workflow_standard.md](./global_workflow_standard.md)
- workflow 인덱스: [../WORKFLOW_INDEX.md](../WORKFLOW_INDEX.md)
- 프로젝트 상태 진단: [../memory/project_status_assessment.md](../memory/project_status_assessment.md)
