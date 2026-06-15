# Project Workflow Profile

- 문서 목적: Standard AI Workflow 저장소 자체의 운영 규칙과 검증 기준을 정의한다.
- 범위: 저장소 개요, 문서 구조, 기본 명령, 검증 포인트, 예외 규칙
- 대상 독자: 저장소 maintainer, 멀티 에이전트 운영자, AI agent
- 상태: stable
- 최종 수정일: 2026-06-07
- 관련 문서: [공통 표준](../../../../workflow-source/core/global_workflow_standard.md), [Maturity Matrix](../../../../workflow-source/core/maturity_matrix.json), [Orchestrator ↔ Sub-agent Contract v1](../../../../workflow-source/core/orchestrator_subagent_contract_v1.md)

## 1. 프로젝트 개요

- 프로젝트명: **Standard AI Workflow**
- 프로젝트 슬러그: `standard-ai-workflow`
- 프로젝트 목적: 여러 프로젝트에서 공통으로 사용할 수 있는 표준 AI 협업 워크플로우 문서와 템플릿, skill/MCP/agent 구현 기준을 독립 프로젝트 형태로 제공한다.
- 주요 이해관계자: 저장소 maintainer (`ykylee`), 워크플로우 도입 검토자, 멀티 에이전트 운영자
- 현재 베이스라인: **v0.5.6-beta** (in progress; main 은 v0.5.5 까지 stable, v0.5.6 에서 P0 enforcement — §5 validator + §6.1 delegator)

## 2. 문서 구조 (Path)

- 문서 위키 홈: [`docs/README.md`](../../../../docs/README.md)
- 운영 문서 홈: `ai-workflow/memory/`
- 백로그 위치: `ai-workflow/memory/backlog/`
- 세션 인계 문서: `ai-workflow/memory/session_handoff.md`
- 환경 기록 위치: `ai-workflow/memory/environments/`

## 3. 운영 / 검증 명령

- **venv 셋업 (최초 1회)**: `python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt -r requirements-dev.txt`
- bootstrap dry-run: `python3 workflow-source/scripts/bootstrap_workflow_kit.py --target-root . --project-slug demo --project-name "Demo" --harness codex --dry-run`
- bootstrap 실제 실행: `python3 workflow-source/scripts/bootstrap_workflow_kit.py --target-root . --project-slug demo --project-name "Demo" --harness codex --copy-core-docs`
- 회귀: `python workflow-source/tests/check_bootstrap.py`
- 회귀 (MCP): `python workflow-source/tests/check_bootstrap_mcp_roundtrip.py`
- 문서 링크 검증: `python workflow-source/tests/check_docs.py`
- contract v1:
  - `python workflow-source/tests/check_contract_v1_roundtrip.py`
  - `python workflow-source/tests/check_contract_v1_role_mapping.py`
  - `python workflow-source/tests/check_contract_v1_direct_only.py`
  - `python workflow-source/tests/check_contract_v1_output_validator.py` (v0.5.6 신규)
  - `python workflow-source/tests/check_contract_v1_delegator.py` (v0.5.6 신규)
- pilot phase11 (v0.5.5): `python workflow-source/tests/check_pilot_phase11_contract_v1.py`
- 워크플로우 linter: `PYTHONPATH=workflow-source python workflow-source/skills/workflow-linter/scripts/run_workflow_linter.py --project-profile-path docs/PROJECT_PROFILE.md --state-json-path ai-workflow/memory/release/v0.5.6/state.json --session-handoff-path ai-workflow/memory/release/v0.5.6/session_handoff.md --latest-backlog-path ai-workflow/memory/release/v0.5.6/backlog/2026-06-07.md`

## 4. 예외 규칙

- 새 하네스 overlay 추가 시 `bootstrap_lib/harnesses/` 아래에 renderer + `HARNESS_SPECS` 등록 + `__main__.py` 의 `register_harness_builder` 호출이 모두 세트로 들어가야 한다.
- `bootstrap_workflow_kit.py` 는 thin re-export entry. 직접 import 하지 말고 `bootstrap_lib.*` 사용.
- `--enable-mcp` 의 MCP config 출력 경로는 하네스별 dot-dir 컨벤션(`.codex/`, `.gemini/`, `.MiniMax/`, `.antigravity/`)을 따른다.
- `requirements.txt` / `package.json` 자동 갱신은 `--update-deps` 옵션일 때만.
- 승인: `core/` 문서 (특히 `global_workflow_standard.md`, `maturity_matrix.json`, `orchestrator_subagent_contract_v1.md`) 변경 시 자체 self-review 필수
- (v0.5.4 부터) orchestrator는 sub-agent 위임 가능한 작업을 직접 도구 호출로 처리하지 않는다 (제6.1장). 자세한 contract: `../workflow-source/core/orchestrator_subagent_contract_v1.md`.
- (v0.5.6 신규) sub-agent 출력은 `workflow_kit.contract_v1.output_validator.validate_output()` 으로 자동 검증 후 orchestrator 에 보고. 검증 실패 시 §7.4 의 "출력 스키마 위반" 정책 적용.

## 5. 다음에 읽을 문서

- [Maturity Matrix](../../../../workflow-source/core/maturity_matrix.json)
- [Orchestrator ↔ Sub-agent Contract v1](../../../../workflow-source/core/orchestrator_subagent_contract_v1.md)
- [세션 인계 문서](./session_handoff.md)
- [작업 백로그 인덱스](../../work_backlog.md)
- [Pilot result v0.5.5](../../../../workflow-source/examples/pilot_phase11_devhub_contract_v1.md)
- [릴리스 노트 v0.5.6](../../../../workflow-source/releases/Beta-v0.5.6.md)
