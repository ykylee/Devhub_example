# AGENTS.md (Pi Coding Agent Profile)

- 문서 목적: Pi Coding Agent가 본 저장소에서 작업을 수행할 때 따라야 할 핵심 지침과 워크플로우를 정의한다.
- 범위: 세션 시작 루틴, 작업 원칙, 상태 관리, 도구 및 언어 가이드
- 대상 독자: Pi Coding Agent, AI Agent 설계자
- 상태: stable
- 최종 수정일: 2026-04-29
- 관련 문서: `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, `ai-workflow/memory/work_backlog.md`

- **Mandate**: 본 저장소는 'Standard AI Workflow'를 따릅니다. 모든 행동은 아래 문서의 상태를 기준으로 결정하십시오.
- **Priority Docs**:
    1. `ai-workflow/memory/state.json` (현재 세션의 진실의 원천)
    2. `ai-workflow/memory/session_handoff.md` (이전 세션 인계 사항)
    3. `ai-workflow/memory/work_backlog.md` (작업 목록)

## 1. 세션 시작 루틴 (Mandatory)
세션이 시작되면 가장 먼저 `ai-workflow/memory/state.json`을 읽고 `current_focus`와 `next_documents`를 파악하십시오. 이후 `session_handoff.md`를 읽어 중단된 지점부터 작업을 재개하십시오.

## 2. 작업 원칙 (Research -> Strategy -> Execution)
- **Research**: `grep_search`와 `read_file`을 사용하여 현재 코드와 문서 상태를 객관적으로 확인하십시오.
- **Strategy**: 변경 계획을 세우고, 작업 전후에 어떤 문서를 갱신할지 결정하십시오.
- **Execution**: `edit`, `write`, `bash` 도구를 사용하여 변경을 수행하십시오.

## 3. 워크플로우 상태 관리
- 작업 상태가 변경되면 반드시 `ai-workflow/memory/backlog/`의 해당 날짜 문서를 업데이트하십시오.
- 세션 종료 전에는 `ai-workflow/memory/state.json`과 `session_handoff.md`를 갱신하여 다음 에이전트를 위한 맥락을 보존하십시오.

## 4. 도구 사용 가이드
- 복잡한 워크플로우 제어(상태 자동 갱신 등)가 필요할 때 `python3 ai-workflow/scripts/` 아래의 도구들을 활용할 수 있습니다.
- 모든 도구 호출 결과는 구조화된 JSON으로 처리하는 것을 선호합니다.

## 5. 언어 가이드
- 사용자에게 보고하거나 문서를 작성할 때는 한국어를 사용하십시오.
- 코드와 기술적 명칭은 원문을 유지하십시오.
