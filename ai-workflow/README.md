# Standard AI Workflow Kit

- 문서 목적: `Devhub Example Codex` 저장소에 표준 AI 워크플로우 기본 문서 세트를 도입하고 현재 운영 문서 위치를 안내한다.
- 범위: 공통 코어 문서 위치, 프로젝트 상태 문서 세트, 도입 모드별 후속 작업
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-05-07
- 관련 문서: `ai-workflow/MEMORY_GOVERNANCE.md`, `ai-workflow/memory/<agent>/<branch>/state.json`, `ai-workflow/memory/PROJECT_PROFILE.md`

## 1. 현재 배포 형태

- 선택한 도입 모드: `existing`
- 요약: 기존 프로젝트 분석 결과를 반영한 문서 초안과 평가 문서를 생성했고, 런타임 운영 문서 세트는 `ai-workflow/memory/<agent>/<branch>/` 아래에서 브랜치별로 관리한다.

## 2. 현재 유지되는 파일

- 공용 프로파일: [PROJECT_PROFILE.md](./memory/PROJECT_PROFILE.md)
- 공용 평가: [repository_assessment.md](./memory/repository_assessment.md)
- 현재 Codex 브랜치 state: [codex/service-action-command/state.json](./memory/codex/service-action-command/state.json)
- 현재 Codex 브랜치 handoff: [codex/service-action-command/session_handoff.md](./memory/codex/service-action-command/session_handoff.md)
- 현재 Codex 브랜치 backlog: [codex/service-action-command/work_backlog.md](./memory/codex/service-action-command/work_backlog.md)
- Claude 브랜치 memory: [claude/](./memory/claude/)
- flat `memory/state.json`, `memory/session_handoff.md`, `memory/work_backlog.md`, `memory/backlog/`는 legacy fallback 및 공용 색인 전용

## 3. 코어 문서

- core 문서는 `--copy-core-docs` 옵션을 사용하면 함께 복사할 수 있다.

## 4. 하네스 오버레이

- `codex` 하네스용 오버레이 파일 생성

## 5. 도입 직후 해야 할 일

1. `memory/PROJECT_PROFILE.md`의 TODO 항목과 기본 명령을 실제 프로젝트 값으로 채운다.
2. 현재 브랜치별 `state.json`, `session_handoff.md`, 최신 backlog가 현재 작업 기준과 맞는지 확인한다.
3. `memory/repository_assessment.md`의 추정값을 실제 저장소 규칙과 대조해 수정한다.
4. 루트 `AGENTS.md`의 프로젝트 실행 기본값을 `PROJECT_PROFILE.md`와 맞춘다.
5. 이후 skill/MCP 도입 범위는 현재 저장소에 남아 있는 `core/` 문서와 별도 workflow kit 배포물을 기준으로 결정한다.

## 6. 언어와 컨텍스트 운영 원칙

- 사용자에게 직접 보이는 작업 보고, 상태 요약, handoff/backlog 갱신 문안은 기본적으로 한국어로 작성한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지한다.
- 내부 사고 과정과 중간 분류는 모델이 가장 효율적인 형태로 처리하고, 사용자에게는 필요한 결론만 짧게 전달한다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 불필요한 컨텍스트 누적을 줄인다.

## 7. 프로젝트 실제 문서 경로 설정값

- 문서 위키 홈: `README.md`, `docs/README.md`
- 운영 문서 위치: `ai-workflow/memory/<agent>/<branch>/`
- 백로그 위치: `ai-workflow/memory/<agent>/<branch>/backlog/`
- 세션 인계 문서 위치: `ai-workflow/memory/<agent>/<branch>/session_handoff.md`
- flat memory 위치: legacy fallback 및 공용 색인 전용
- 환경 기록 위치: `ai-workflow/memory/environments/`

## 다음에 읽을 문서

- 브랜치별 memory 규칙: [./MEMORY_GOVERNANCE.md](./MEMORY_GOVERNANCE.md)
- 프로젝트 프로파일: [./memory/PROJECT_PROFILE.md](./memory/PROJECT_PROFILE.md)
- 현재 Codex 브랜치 상태: [./memory/codex/service-action-command/state.json](./memory/codex/service-action-command/state.json)
- 현재 Codex 브랜치 인계: [./memory/codex/service-action-command/session_handoff.md](./memory/codex/service-action-command/session_handoff.md)
- 현재 Codex 브랜치 백로그: [./memory/codex/service-action-command/work_backlog.md](./memory/codex/service-action-command/work_backlog.md)
