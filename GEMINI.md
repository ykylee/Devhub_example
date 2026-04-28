# GEMINI.md

- 문서 목적: Gemini CLI 가 이 저장소에서 먼저 읽어야 할 workflow 진입 규칙과 기본 작업 원칙을 제공한다.
- 범위: 세션 복원, workflow state docs 참조 순서, 사용자 보고 언어, 기본 실행/검증 명령
- 대상 독자: Gemini CLI, 저장소 관리자, workflow 설계자
- 상태: draft
- 최종 수정일: 2026-04-28
- 관련 문서: `ai-workflow/project/state.json`, `ai-workflow/project/session_handoff.md`, `ai-workflow/project/work_backlog.md`, `ai-workflow/project/project_workflow_profile.md`

## 목적

이 저장소에서는 표준 AI 워크플로우를 기준으로 작업한다. 세션 시작, backlog 갱신, 문서 동기화, 세션 종료는 `ai-workflow/` 아래 문서를 우선 기준으로 삼는다.

## 항상 먼저 읽을 문서

- `ai-workflow/project/state.json`
- `ai-workflow/project/session_handoff.md`
- `ai-workflow/project/work_backlog.md`
- `ai-workflow/project/project_workflow_profile.md`

`ai-workflow/` 는 세션 복원과 workflow 상태 관리용 메타 레이어다. 프로젝트 코드나 프로젝트 문서를 탐색할 때는 이 경로를 기본 탐색 범위에 넣지 말고, workflow 문서 자체를 갱신하거나 현재 세션 상태를 복원할 때만 예외적으로 참조한다.

## 작업 원칙

- 작업을 시작하기 전에 목적, 범위, 영향 문서를 짧게 정리한다.
- 작업 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로 관리한다.
- 검증하지 않은 결과는 완료로 확정하지 않는다.
- 세션 종료 전에는 `state.json`, `session_handoff.md`, 최신 backlog 를 갱신한다.

## 언어와 컨텍스트 원칙

- 사용자에게 직접 보이는 작업 보고, 상태 요약, 문서 갱신 문안은 기본적으로 한국어로 작성한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지한다.
- 내부 사고 과정과 임시 분류는 모델이 가장 효율적인 방식으로 처리하되, 사용자에게는 필요한 결론과 다음 행동만 짧게 전달한다.
- 장문의 중간 reasoning, 중복 요약, 불필요한 자기 설명을 피한다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 불필요한 컨텍스트 누적을 줄인다.

## 프로젝트 실행 기본값

- 설치: `python3 -m venv .venv && source .venv/bin/activate && pip install mcp pytest pytest-asyncio`
- 로컬 실행: `python3 ai-workflow/scripts/run_demo_workflow.py`
- 빠른 테스트: `pytest ai-workflow/tests/check_docs.py`
- 격리 테스트: `pytest ai-workflow/tests/`
- 실행 확인: `python3 ai-workflow/scripts/bootstrap_workflow_kit.py --help`

## 문서 작업 기준

- 문서 위키 홈: `README.md`
- 운영 문서 위치: `ai-workflow/project/`
- backlog 위치: `ai-workflow/project/backlog/`
- session handoff 위치: `ai-workflow/project/session_handoff.md`

## Gemini CLI 전용 메모

- Gemini CLI 는 프로젝트 루트의 `GEMINI.md` 를 읽으므로, 상세 정책은 본 문서에서 시작하고 세부 운영 기준은 `ai-workflow/` 문서를 참조한다.
- `GEMINI.md` 에 기재된 지침은 시스템 프롬프트보다 우선하는 강력한 지침으로 취급한다.
- 가능한 경우 메인 에이전트는 조정과 통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 서브 에이전트(`invoke_agent`)로 분리하는 패턴을 권장한다.
- 서브 에이전트에게는 책임 범위와 종료 조건을 명확히 넘기고, 메인 에이전트에는 핵심 사실과 결과만 다시 모은다.
- 기존 코드베이스 분석 결과를 반영한 초안이다. 추정 명령과 문서 경로는 실제 저장소 기준으로 수정할 수 있다.
