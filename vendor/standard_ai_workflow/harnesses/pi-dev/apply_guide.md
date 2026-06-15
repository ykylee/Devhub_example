# Pi-dev Harness Apply Guide

- 문서 목적: `bootstrap_workflow_kit.py` 로 생성한 Pi Coding Agent 오버레이를 실제 프로젝트에 적용할 때의 단계별 절차와 주의 사항을 정리한다.
- 범위: 사전 점검, 파일 복사, 첫 세션 시작, 트러블슈팅
- 대상 독자: 표준 AI 워크플로우를 Pi Coding Agent로 처음 도입하는 개발자
- 상태: prototype
- 최종 수정일: 2026-06-09
- 관련 문서: `./README.md`, `./AGENTS.md`, `../../core/workflow_adoption_entrypoints.md`

## 1. 사전 점검

적용 전 다음을 확인한다.

- Python 3.11+ 설치 (`python3 --version`)
- Pi Coding Agent 접근 가능 (`pi.dev` 에서 확인)
- 저장소에 표준 워크플로우 패키지 (`ai-workflow/`) 가 bootstrap 되었거나, 도입 직전 단계

## 2. overlay 파일 적용

`bootstrap_workflow_kit.py --harness pi-dev` 실행 시 다음 파일이 생성된다.

```
<project_root>/
└── AGENTS.md
```

Pi Coding Agent는 Codex와 동일한 `AGENTS.md` 진입점을 사용한다. `AGENTS.md` 는 Pi가 프로젝트의 `state.json`과 `session_handoff.md`를 엄격하게 관리하도록 지침을 제공한다.

> **참고**: `SYSTEM.md` (에이전트 페르소나 및 운영 원칙 보강) 는 `README.md` 에 포함 파일로 명시되어 있으나, 현재 bootstrap 출력에 포함되지 않은 프로토타입 상태다. 필요하면 수동으로 작성한다.

## 3. 첫 세션 검증

```bash
# 1. 워크플로우 상태 점검
python3 -m workflow_kit.server.read_only_entrypoint --list-tools

# 2. Pi Coding Agent 세션 시작
# pi.dev 에서 다음 요청으로 시작:
# "AGENTS.md를 읽고 워크플로우 세션을 시작해줘"
```

세션이 정상 시작되면 다음을 확인한다.

- `ai-workflow/memory/active/state.json` 이 자동 갱신됨
- `session_handoff.md` 의 "다음 세션 시작 포인트" 가 한 문장 갱신됨
- Wiki 진입점: `ai-workflow/wiki/index.md` (R4 anchor 기반). AI agent query 시 먼저 로드.
- `AGENTS.md` 의 §1 세션 시작 루틴에 따라 `state.json` → `session_handoff.md` → `work_backlog.md` 순서로 읽기 루틴이 수행됨

## 4. 트러블슈팅

| 증상 | 원인 | 해결 |
| --- | --- | --- |
| `AGENTS.md` 가 로드되지 않음 | Pi가 프로젝트 루트를 인식하지 못함 | 프로젝트 루트에서 세션 재시작 |
| `state.json` 갱신 실패 | `PYTHONPATH` 가 `workflow-source` 를 가리키지 않음 | `export PYTHONPATH=workflow-source:$PYTHONPATH` |
| 한국어 보고가 영어로 나옴 | `AGENTS.md` §5 언어 가이드 미적용 | `AGENTS.md` 의 언어 설정 확인 후 세션 재시작 |
| 작업 상태가 backlog 에 반영되지 않음 | §3 워크플로우 상태 관리 루틴 누락 | `ai-workflow/memory/active/backlog/` 의 날짜 문서 직접 갱신 |

## 5. 다음 단계

- 첫 적용이 끝나면 `workflow-source/harnesses/pi-dev/README.md` 와 본 가이드를 함께 검토한다.
- 추가 에이전트 페르소나가 필요하면 `SYSTEM.md` 를 수동 작성하여 `AGENTS.md` 와 페어링한다.
- 운영 패턴이 안정되면 `pilot_adoption_record_template.md` 로 도입 기록을 남긴다.

## 6. 로컬 MCP 설치 (`--enable-mcp`)

Pi Coding Agent 의 MCP 연결은 `AGENTS.md` 의 도구 사용 가이드(§4)에 따라 `workflow-source/` 아래 도구를 직접 호출하는 방식이다.

### 6.1 자동 심기

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root <project_root> \
  --project-slug <slug> --project-name "<name>" \
  --harness pi-dev --adoption-mode existing --copy-core-docs \
  --enable-mcp
```

Pi-dev 하네스는 별도 MCP config 파일을 생성하지 않으며, `AGENTS.md` §4 에서 `ai-workflow/scripts/` 아래 도구를 직접 호출하는 패턴을 따른다.

### 6.2 Transport 선택

- 기본 `jsonrpc-bridge` (권장, 안정) — `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines`
- `stdio-sdk` (실험적) — `python3 -m workflow_kit.server.read_only_mcp_sdk --stdio-sdk`. `--mcp-bridge stdio-sdk` 로 전환.

자세한 가이드: [`../core/mcp_installation_by_harness.md`](../core/mcp_installation_by_harness.md)
