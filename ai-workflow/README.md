# Standard AI Workflow Kit

- 문서 목적: `Devhub Example` 저장소의 표준 AI 워크플로우 진입점을 안내한다.
- 범위: 현재 배포된 최소 workflow 문서 세트, 프로젝트 상태 문서, 온보딩 후속 작업
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-05-03

- 관련 문서: `ai-workflow/memory/PROJECT_PROFILE.md`, `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, `ai-workflow/memory/work_backlog.md`


## 1. 현재 배포 형태

- 배포 버전: `v0.4.1-beta`
- 도입 모드: `existing`
- 요약: workflow engine/source bundle(`scripts`, `skills`, `mcp`, `templates`, `tests`, `examples`)은 저장소에서 제외하고, 프로젝트 운영에 필요한 runtime 문서만 유지한다.


## 2. 현재 유지되는 파일

- [WORKFLOW_INDEX.md](./WORKFLOW_INDEX.md)
- [memory/PROJECT_PROFILE.md](./memory/PROJECT_PROFILE.md)
- [memory/state.json](./memory/state.json)
- [memory/session_handoff.md](./memory/session_handoff.md)
- [memory/work_backlog.md](./memory/work_backlog.md)
- [memory/backlog/2026-05-03.md](./memory/backlog/2026-05-03.md)
- [memory/project_status_assessment.md](./memory/project_status_assessment.md)


## 3. 코어 문서

- [core/global_workflow_standard.md](./core/global_workflow_standard.md)
- [core/workflow_skill_catalog.md](./core/workflow_skill_catalog.md)
- [core/workflow_adoption_entrypoints.md](./core/workflow_adoption_entrypoints.md)

## 4. 하네스/도구 오버레이

- 현재 저장소에는 하네스, MCP, skill 실행 스크립트가 포함되어 있지 않다.
- 필요한 경우 별도 배포된 workflow kit source bundle에서 도구를 설치하거나, core 문서를 기준으로 수동 절차를 수행한다.

## 5. 도입 직후 해야 할 일


1. `memory/PROJECT_PROFILE.md`의 TODO 항목과 기본 명령을 실제 프로젝트 값으로 채운다.
2. `memory/state.json`, `memory/session_handoff.md`, 최신 backlog가 현재 작업 기준과 맞는지 확인한다.
3. `memory/project_status_assessment.md`의 추정값을 실제 저장소 규칙과 대조해 수정한다.
4. 루트 `AGENTS.md`의 프로젝트 실행 기본값을 `PROJECT_PROFILE.md`와 맞춘다.
5. 이후 skill/MCP 도입 범위는 현재 저장소에 남아 있는 `core/` 문서와 별도 workflow kit 배포물을 기준으로 결정한다.


## 6. 언어와 컨텍스트 운영 원칙

- 사용자에게 직접 보이는 작업 보고, 상태 요약, handoff/backlog 갱신 문안은 기본적으로 한국어로 작성한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지한다.
- 내부 사고 과정과 중간 분류는 모델이 가장 효율적인 형태로 처리하고, 사용자에게는 필요한 결론만 짧게 전달한다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 불필요한 컨텍스트 누적을 줄인다.

## 7. 프로젝트 실제 문서 경로 설정값


- 문서 위키 홈: `README.md`, `docs/README.md`
- 운영 문서 위치: `ai-workflow/memory/`
- 백로그 위치: `ai-workflow/memory/backlog/`
- 세션 인계 문서 위치: `ai-workflow/memory/session_handoff.md`
- 환경 기록 위치: `ai-workflow/memory/environments/`

## 다음에 읽을 문서

- workflow 인덱스: [./WORKFLOW_INDEX.md](./WORKFLOW_INDEX.md)
- 프로젝트 프로파일: [./memory/PROJECT_PROFILE.md](./memory/PROJECT_PROFILE.md)
- 빠른 상태 요약: [./memory/state.json](./memory/state.json)
- 세션 인계 문서: [./memory/session_handoff.md](./memory/session_handoff.md)
- 작업 백로그 인덱스: [./memory/work_backlog.md](./memory/work_backlog.md)

