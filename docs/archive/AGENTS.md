# AGENTS.md

> ⚠ **Deprecated** (2026-05-13, sprint `claude/work_260513-i`). 본 문서는 archive 위치로 이동된 초안. workflow 진입 규칙은 `CLAUDE.md` (Claude Code 진입점) + `ai-workflow/memory/` (브랜치별 운영 메모리) 가 source-of-truth. 본 문서는 참조용으로만 유지된다.

- 문서 목적: Codex 가 이 저장소에서 먼저 읽어야 할 workflow 진입 규칙과 기본 작업 원칙을 제공한다.
- 범위: 세션 복원, workflow state docs 참조 순서, 사용자 보고 언어, 기본 실행/검증 명령
- 대상 독자: Codex, 저장소 관리자, workflow 설계자
- 상태: deprecated
- 최종 수정일: 2026-05-13 (deprecated 마킹)
- 관련 문서: `ai-workflow/WORKFLOW_INDEX.md`, `ai-workflow/memory/codex/phase6/state.json`, `ai-workflow/memory/codex/phase6/session_handoff.md`, `ai-workflow/memory/work_backlog.md`, `docs/PROJECT_PROFILE.md`

## 목적

이 저장소에서는 표준 AI 워크플로우를 기준으로 작업한다. 세션 시작, backlog 갱신, 문서 동기화, 세션 종료는 `ai-workflow/` 아래 문서를 우선 기준으로 삼는다.

## 항상 먼저 읽을 문서 (운영 전용)

- `ai-workflow/WORKFLOW_INDEX.md`
- `ai-workflow/memory/codex/phase6/state.json` (브랜치별 상태 캐시)
- `ai-workflow/memory/codex/phase6/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`

## 프로젝트 분석 및 온보딩 원칙

- **격리**: `ai-workflow/`는 세션 복원과 워크플로우 상태 관리용 메타 레이어입니다. **프로젝트 코드나 프로젝트 아키텍처를 분석할 때 이 경로를 참조하거나 학습 데이터로 사용해서는 안 됩니다.**
- **소스-오브-트루스**: 프로젝트 이해를 위한 모든 정보는 `docs/` 및 소스 코드에서만 가져옵니다. 에이전트 온보딩 시 **[Project Profile](docs/PROJECT_PROFILE.md)**과 **[Document Index](docs/DOCUMENT_INDEX.md)**를 시작점으로 사용하십시오.

## 작업 원칙

- 작업을 시작하기 전에 목적, 범위, 영향 문서를 짧게 정리한다.
- 작업 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로 관리한다.
- 검증하지 않은 결과는 완료로 확정하지 않는다.
- 세션 종료 전에는 브랜치별 `state.json`, `session_handoff.md`, 최신 backlog 를 갱신한다.

## 언어와 컨텍스트 원칙

- 사용자에게 직접 보이는 작업 보고, 상태 요약, 문서 갱신 문안은 기본적으로 한국어로 작성한다.
- **AI 내부 운영 문서**: `ai-workflow/memory/` 하위 문서는 AI 최적화 가이드([MEMORY_GOVERNANCE.md](./workflow-source/MEMORY_GOVERNANCE.md))를 준수하여 작성한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지한다.
- 내부 사고 과정과 임시 분류는 모델이 가장 효율적인 방식으로 처리하되, 사용자에게는 필요한 결론과 다음 행동만 짧게 전달한다.
- 장문의 중간 reasoning, 중복 요약, 불필요한 자기 설명을 피한다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 불필요한 컨텍스트 누적을 줄인다.

## 프로젝트 실행 기본값

- 설치: `python3 -m pip install -r requirements-dev.txt`
- 로컬 실행: `python3 workflow-source/scripts/bootstrap_workflow_kit.py --help`
- 빠른 테스트: `python3 workflow-source/tests/check_docs.py`, `python3 workflow-source/tests/check_output_samples.py`
- 격리 테스트: `python3 workflow-source/tests/check_demo_workflow.py`, `python3 workflow-source/tests/check_existing_project_onboarding.py`
- 실행 확인: `python3 workflow-source/scripts/run_demo_workflow.py --example-project acme_delivery_platform`

## 문서 작업 기준

- 문서 위키 홈: `docs/README.md`
- 운영 문서 위치: `ai-workflow/memory/codex/phase6/`
- backlog 위치: `ai-workflow/memory/codex/phase6/backlog/tasks/`
- session handoff 위치: `ai-workflow/memory/codex/phase6/session_handoff.md`

## Codex 전용 메모

- Codex 는 프로젝트 루트의 `AGENTS.md` 를 읽으므로, 상세 정책은 본 문서에서 시작하고 세부 운영 기준은 `ai-workflow/` 문서를 참조한다.
- OpenAI 관련 질문이 나오면 OpenAI 문서 MCP 를 우선 사용하는 구성을 권장한다.
- 가능한 경우 메인 에이전트는 조정과 통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 worker 성격의 서브 에이전트로 분리하는 패턴을 권장한다.
- worker 에게는 책임 파일과 종료 조건을 명확히 넘기고, 메인 에이전트에는 핵심 사실과 결과만 다시 모은다.
- `main`/`small` 모델을 함께 운영한다면, 메인 에이전트는 난도 높은 판단과 통합에, worker 는 bounded scope 탐색/초안/검증에 우선 배치하는 편이 효율적이다.
- 기존 코드베이스 분석 결과를 반영한 초안이다. 추정 명령과 문서 경로는 실제 저장소 기준으로 수정할 수 있다.
- 이 저장소는 self-dogfood 중이므로 `ai-workflow/memory/*` 는 workflow state docs, `docs/` 및 소스 코드는 실제 project docs/code 로 구분해 다룬다.
