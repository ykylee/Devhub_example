# Prototype-v2 Pre-release

- 문서 목적: `prototype-v2` pre-release 의 주요 변경점, 배포 산출물, 적용 대상, 검증 결과를 기록한다.
- 범위: state cache 도입, orchestrator/worker 운영 기본값 정렬, 하네스 패키지 배포 산출물
- 대상 독자: 저장소 관리자, 배포 담당자, 하네스 통합 담당자, 파일럿 적용자
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../README.md`, `../core/workflow_kit_roadmap.md`, `../core/workflow_release_spec.md`, `../scripts/export_harness_package.py`, `../scripts/generate_workflow_state.py`

## 1. 릴리즈 성격

- 버전: `prototype-v2`
- 채널: `pre-release`
- 릴리즈 초점: 빠른 세션 복원용 `state.json` 캐시 추가와 task-only orchestrator 운영 기본값 정착
- 제외 범위: official MCP SDK server 기본 활성화, MCP runtime 자동 연결, multi-project 파일럿 일반화

## 2. 핵심 변경

### workflow 운영 축

- orchestrator 는 직접 도구 호출보다 task delegation 과 결과 통합에 집중한다는 원칙을 core workflow 문서에 기본값으로 올렸다.
- worker 는 bounded scope 안에서 실제 탐색/수정/검증을 수행하고, low-risk 작업에서는 `ask` 를 최소화하는 운영값을 기본으로 정리했다.
- bootstrap, 하네스 가이드, 예시 문서, 문서 smoke 기준을 같은 방향으로 다시 맞췄다.

### state cache 축

- `ai-workflow/memory/state.json` 을 기본 생성물에 추가했다.
- `PROJECT_PROFILE.md`, `session_handoff.md`, `work_backlog.md`, 최신 날짜 backlog 를 읽어 빠른 세션 복원용 최소 상태를 추출한다.
- `generate_workflow_state.py` 를 추가해 handoff/backlog 갱신 뒤 state cache 를 재생성할 수 있게 했다.

### 배포 패키징 축

- 하네스 export minimal runtime 패키지에 `bundle/ai-workflow/memory/state.json` 을 포함하도록 바꿨다.
- Codex/OpenCode instruction chain 과 package apply guide 가 `state.json` 을 먼저 읽는 흐름을 기본으로 따르도록 정리했다.

## 3. 배포 산출물

### Codex

- package root: `dist/harnesses/codex/prototype-v2/`
- zip: `dist/harnesses/codex/prototype-v2/standard-ai-workflow-codex-prototype-v2.zip`
- entrypoints:
- `bundle/AGENTS.md`
- `bundle/ai-workflow/memory/state.json`
- `bundle/ai-workflow/memory/session_handoff.md`
- `bundle/ai-workflow/memory/work_backlog.md`
- `bundle/ai-workflow/memory/PROJECT_PROFILE.md`

### OpenCode

- package root: `dist/harnesses/opencode/prototype-v2/`
- zip: `dist/harnesses/opencode/prototype-v2/standard-ai-workflow-opencode-prototype-v2.zip`
- entrypoints:
- `bundle/AGENTS.md`
- `bundle/opencode.json`
- `bundle/.opencode/skills/standard-ai-workflow/SKILL.md`
- `bundle/.opencode/agents/workflow-orchestrator.md`
- `bundle/ai-workflow/memory/state.json`

## 4. 검증

- `python3 tests/check_generate_workflow_state.py`
- `python3 tests/check_bootstrap.py`
- `python3 tests/check_export_harness_package.py`
- `python3 tests/check_output_samples.py`
- `python3 tests/check_read_only_transport_descriptors.py`
- `python3 tests/check_read_only_jsonrpc_fixtures.py`

## 5. 다음 릴리즈 제안

1. handoff/backlog 갱신 스크립트가 실제 파일을 수정하는 경로에 `state.json` 자동 재생성을 더 직접적으로 연결한다.
2. 파일럿 저장소 1~2건에서 `state.json` 이 실제 세션 복원 비용을 얼마나 줄이는지 기록한다.
3. MCP 승격 전, state cache 와 read-only bundle 의 경계를 더 명확히 문서화한다.
