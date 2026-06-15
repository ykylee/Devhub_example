# MiniMax Code Harness

- 문서 목적: MiniMax Code(미니맥스 코드) 환경에서 표준 AI 워크플로우를 운영하기 위한 진입점과 워커 분배 구조를 안내한다.
- 범위: 진입 파일, `.minimax/agents/` 워커 오버레이, `MiniMax_config.example.json` 적용, bootstrap 재실행
- 대상 독자: MiniMax Code 운영자, 멀티 에이전트 워크플로우 설계자
- 상태: beta
- 최종 수정일: 2026-06-05
- 관련 문서: `../../core/workflow_harness_distribution.md`, `../../core/workflow_agent_topology.md`

## 1. 진입 파일

MiniMax Code overlay는 다음 두 진입점을 생성한다.

- `AGENTS.md` (Codex/OpenCode와 공통): 워크플로우 규칙 요약
- `MiniMax.md` (MiniMax Code 전용): 메인 orchestrator 운영 원칙과 한국어 보고 규칙

`AGENTS.md` 와 `MiniMax.md` 가 가리키는 사실이 다르다면 `MiniMax.md` 가 우선하되, 두 문서를 동기화 상태로 유지한다.

Wiki 진입점: `ai-workflow/wiki/index.md` (R4 anchor 기반). AI agent query 시 먼저 로드.

## 2. 워커 오버레이 (`.minimax/agents/`)

| 파일 | 역할 | 참고 |
| --- | --- | --- |
| `workflow-orchestrator.md` | 메인 orchestrator 페르소나 (작업 분해, 워커 위임, handoff/state 동기화) | `../../../workflow-source/prompts/code_worker_prompt.md` 와 페어링 |
| `workflow-worker.md` | 워커 공통 운영 계약 (`WorkerTask` / `WorkerResponse` 형식) | `workflow_kit.common.schemas.worker.*` |
| `workflow-doc-worker.md` | 문서 정합성 / 메타데이터 / 카탈로그 동기화 | `workflow-source/skills/doc-sync` |
| `workflow-code-worker.md` | 코드 구현 / 정밀 리팩토링 | `workflow-source/skills/robust_patcher`, `code-index-update` |
| `workflow-validation-worker.md` | 테스트/스모크 실행 및 결과 기록 | `workflow-source/skills/validation-plan`, `workflow-source/tests/check_*.py` |

## 3. `MiniMax_config.example.json` 적용

1. 프로젝트 루트에서 `MiniMax_config.example.json` 을 `.minimax/config.json` 으로 복사한다.
   ```bash
   cp MiniMax_config.example.json .minimax/config.json
   ```
2. `project_name`, `agents.*.file` 경로, `mcp_servers.*.command` 등을 실제 환경에 맞게 채운다.
3. 사용자 인증 토큰 등 시크릿은 환경 변수 또는 별도 시크릿 매니저에서 주입한다 (절대 config.json 에 직접 두지 않는다).

## 4. bootstrap 재실행

`bootstrap_workflow_kit.py` 가 `MiniMax.md` 와 `.minimax/agents/*` 를 한 번에 재생성한다. 이미 적용한 사용자 편집은 `--force` 옵션으로 덮어쓴다.

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root . \
  --project-slug <slug> \
  --project-name "<name>" \
  --harness minimax-code \
  --adoption-mode existing \
  --copy-core-docs \
  --force
```

## 5. 첫 세션

MiniMax Code 세션에서 다음과 같이 요청해 워크플로우를 활성화한다.

> 프로젝트 루트의 `MiniMax.md` 와 `AGENTS.md` 를 읽고, `ai-workflow/memory/active/state.json` 을 기준으로 워크플로우 세션을 시작해줘.

세션 시작 직후 `state.json` 과 `session_handoff.md` 가 자동 갱신되며, 이후 모든 작업은 메인 orchestrator가 워커에 위임해 수행한다.
