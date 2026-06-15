# MiniMax Code Harness Apply Guide

- 문서 목적: `bootstrap_workflow_kit.py` 로 생성한 MiniMax Code 오버레이를 실제 프로젝트에 적용할 때의 단계별 절차와 주의 사항을 정리한다.
- 범위: 사전 점검, 파일 복사, 설정 적용, 첫 세션 시작, 트러블슈팅
- 대상 독자: 표준 AI 워크플로우를 MiniMax Code로 처음 도입하는 개발자
- 상태: beta
- 최종 수정일: 2026-06-05
- 관련 문서: `./README.md`, `../../../core/workflow_adoption_entrypoints.md`

## 1. 사전 점검

적용 전 다음을 확인한다.

- Python 3.11+ 설치 (`python3 --version`)
- MiniMax Code CLI 설치 (`MiniMax --version`)
- 저장소에 표준 워크플로우 패키지 (`ai-workflow/`) 가 bootstrap 되었거나, 도입 직전 단계

## 2. overlay 파일 적용

`bootstrap_workflow_kit.py --harness minimax-code` 실행 시 다음 파일이 생성된다.

```
<project_root>/
├── AGENTS.md
├── MiniMax.md
├── MiniMax_config.example.json
└── .minimax/
    └── agents/
        ├── workflow-orchestrator.md
        ├── workflow-worker.md
        ├── workflow-doc-worker.md
        ├── workflow-code-worker.md
        └── workflow-validation-worker.md
```

## 3. 설정 적용

```bash
# 1. config.json 초기화
mkdir -p .minimax
cp MiniMax_config.example.json .minimax/config.json

# 2. 환경별 값 채우기
$EDITOR .minimax/config.json
# - project_name, agents[*].file 경로 확인
# - mcp_servers[*].command 가 workflow-source 위치를 가리키는지 확인
# - 시크릿은 환경 변수로 분리 (.env, vault 등)
```

## 4. 첫 세션 검증

```bash
# 1. 워크플로우 상태 점검
python3 -m workflow_kit.server.read_only_entrypoint --list-tools

# 2. MiniMax Code 세션 시작
MiniMax chat "AGENTS.md와 MiniMax.md를 읽고 워크플로우 세션을 시작해줘"
```

세션이 정상 시작되면 다음을 확인한다.

- `ai-workflow/memory/active/state.json` 이 자동 갱신됨
- `session_handoff.md` 의 "다음 세션 시작 포인트" 가 한 문장 갱신됨
- `MiniMax.md` 가 시스템 프롬프트 일부로 로드되어 우선 규칙으로 작동
- Wiki 진입점: `ai-workflow/wiki/index.md` (R4 anchor 기반). AI agent query 시 먼저 로드.

## 5. 트러블슈팅

| 증상 | 원인 | 해결 |
| --- | --- | --- |
| `MiniMax.md` 가 로드되지 않음 | `.minimax/` 디렉터리가 MiniMax Code의 신뢰 경로 밖 | 프로젝트 루트에 있는지 확인 후 세션 재시작 |
| `mcp_servers` 가 연결되지 않음 | `PYTHONPATH` 가 `workflow-source` 를 가리키지 않음 | `MiniMax_config.example.json` 의 `mcp_servers.standard-ai-workflow-readonly.env.PYTHONPATH` 확인 |
| 워커 호출 시 무한 대기 | orchestrator 가 `output_files` 를 명시하지 않음 | `WorkerTask.constraints` 에 "변경 범위는 output_files 한정" 명시 |
| 한국어 보고가 영어로 나옴 | `language: "ko-KR"` 가 config 에 없음 | `MiniMax_config.example.json` 의 `language` 키 추가 |

## 6. 다음 단계

- 첫 적용이 끝나면 `workflow-source/harnesses/minimax-code/README.md` 와 본 가이드를 함께 검토한다.
- 추가 워커 페르소나가 필요하면 `MiniMax.md` 의 "오케스트레이터 / 워커 운영 원칙" 섹션을 갱신하고 `.minimax/agents/` 에 새 파일을 추가한다.
- 운영 패턴이 안정되면 `pilot_adoption_record_template.md` 로 도입 기록을 남긴다.

## 7. 로컬 MCP 설치 (`--enable-mcp`)

MiniMax Code 의 MCP 연결은 `mcp_servers` 키다.

### 7.1 자동 심기

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root <project_root> \
  --project-slug <slug> --project-name "<name>" \
  --harness minimax-code --adoption-mode existing --copy-core-docs \
  --enable-mcp
```

`<root>/.minimax/mcp.json` 생성. 글로벌에 등록하려면 symlink:

```bash
mkdir -p ~/.minimax
ln -sf <project_root>/.minimax/mcp.json ~/.minimax/mcp.json
```

### 7.2 전역에 적용

`~/.minimax/config.json` 의 `mcp_servers` 블록에 bootstrap 출력의 `standardAiWorkflowReadOnly` 항목을 옮긴다. `PYTHONPATH` 와 `STANDARD_AI_WORKFLOW_ROOT` 는 절대 경로로 보정.

### 7.3 Transport 선택

- 기본 `jsonrpc-bridge` (권장, 안정) — `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines`
- `stdio-sdk` (실험적) — `python3 -m workflow_kit.server.read_only_mcp_sdk --stdio-sdk`. `--mcp-bridge stdio-sdk` 로 전환. `check_read_only_mcp_sdk_stdio.py` 가 그린이 되면 정식 권장으로 승격.

자세한 가이드: [`../../core/mcp_installation_by_harness.md`](../../core/mcp_installation_by_harness.md), 예시: [`../../examples/mcp_config_examples/minimax-code-mcp.json`](../../examples/mcp_config_examples/minimax-code-mcp.json), stdio-sdk variant: [`../../examples/mcp_config_examples/minimax-code-mcp.stdio-sdk.json`](../../examples/mcp_config_examples/minimax-code-mcp.stdio-sdk.json)
