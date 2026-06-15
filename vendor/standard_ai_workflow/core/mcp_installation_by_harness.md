# MCP Installation by Harness

- 문서 목적: 각 하네스 (Codex / OpenCode / Gemini CLI / Antigravity / MiniMax Code) 에 표준 AI 워크플로우의 로컬 MCP 서버를 심고 동작시키는 방법을 안내한다.
- 범위: 하네스별 MCP config 위치 / 스키마 / 자동 심기 (`--enable-mcp`) / 수동 적용 / 트러블슈팅
- 대상 독자: 워크플로우 도입자, AI agent 운영자, 멀티 에이전트 setup 담당자
- 상태: beta
- 최종 수정일: 2026-06-09
- 관련 문서: [./workflow_harness_distribution.md](./workflow_harness_distribution.md), [./read_only_mcp_transport_promotion.md](./read_only_mcp_transport_promotion.md), [../core/workflow_global_injection_policy.md](./workflow_global_injection_policy.md), [`../../scripts/bootstrap_workflow_kit.py`](../../scripts/bootstrap_workflow_kit.py), [../harnesses/*/apply_guide.md](../harnesses/)

## 1. 두 가지 MCP transport

저장소는 두 가지 MCP transport 를 같이 제공하며, bootstrap 시 `--mcp-bridge` 플래그로 선택한다.

| 이름 | 엔트리포인트 | 상태 | 권장 시나리오 |
| --- | --- | --- | --- |
| `jsonrpc-bridge` (default) | `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines` | stable (v0.5.0) | 일반 도입. 항상 동작. `transport_ready=false` 이지만 `tools/list` / `tools/call` 은 round-trip 동작 |
| `stdio-sdk` | `python3 -m workflow_kit.server.read_only_mcp_sdk --stdio-sdk` | experimental | 정식 MCP SDK 호환이 필요한 경우. **현재 `check_read_only_mcp_sdk_stdio.py` 가 `Connection closed` 로 fail** (TASK 후속) |

권장: 처음 도입 시 `jsonrpc-bridge` 로 시작. SDK 호환이 꼭 필요하면 별도 TASK 로 추적하면서 점진 전환.

## 2. 공통 환경 변수

두 transport 모두 다음 환경 변수를 받는다.

- `PYTHONPATH`: `workflow_kit` 모듈이 위치한 경로. bootstrap 이 자동 설정 (`workflow-source` 기준 상대 경로). 글로벌로 심을 땐 절대 경로 권장.
- `STANDARD_AI_WORKFLOW_ROOT`: `workflow_kit` 가 `ai-workflow/`, `workflow-source/` 같은 디렉터리 위치를 찾을 때 사용하는 hint. 절대 경로 권장.

## 3. Interactive `--harness` Picker (v0.5.8)

v0.5.8 부터 bootstrap 시 `--harness` 를 명시하지 않으면 TTY 환경에서 interactive picker 가 자동 실행된다.

```bash
python3 -m bootstrap_lib --target-root <project_root> ...
# --harness 미지정 → TTY picker 로 대화형 선택
```

선택 가능한 하네스: `codex`, `opencode`, `gemini-cli`, `antigravity`, `minimax-code`, `pi-dev`

비대화형(non-TTY) 환경에서는 `--harness` 를 명시적으로 지정하지 않으면 오류가 발생한다. CI/CD 파이프라인에서는 반드시 `--harness` 플래그를 명시해야 한다.

비대화형 정책 (v0.5.11 §6 보강): `--no-interactive` 플래그 또는 non-TTY stdin 에서 `--harness` 미지정 시 `bootstrap_lib` 은 silent 0 overlay 생성을 하지 않고 명확한 `SystemExit(1)` + 6개 harness 목록 (`codex, opencode, gemini-cli, antigravity, minimax-code, pi-dev`) 제시 후 fail-fast 한다. 이 동작은 v0.5.8 부터 변경 없이 유지되지만, 비대화형 환경(CI, 파이프라인, 자동 에이전트)에서 명시적 contract 으로 문서화.

CI / 스크립트 환경 권장 호출 (silent 0 overlay 방지):

```bash
python3 -m bootstrap_lib --target-root "$REPO" \
  --project-slug "$SLUG" \
  --harness opencode \
  --no-interactive \
  --adoption-mode existing
```

picker 가 선택한 하네스에 따라 `--enable-mcp` 와 결합 시 해당 하네스용 MCP config 도 자동 emit 된다.

## 4. 하네스별 MCP config 위치

| 하네스 | 글로벌 설정 위치 | 프로젝트 로컬 위치 (bootstrap 출력) | 스키마 |
| --- | --- | --- | --- |
| **Codex** | `~/.codex/config.toml` (`[mcp_servers.<alias>]` 섹션) | `<root>/.codex/mcp.toml` | TOML |
| **OpenCode** | `opencode.json` 의 `"mcp": { ... }` 키 | `<root>/mcp.opencode.json` | JSON (top-level `mcp` 키) |
| **Gemini CLI** | `~/.gemini/settings.json` 의 `"mcpServers": { ... }` | `<root>/.gemini/mcp.json` | JSON (`mcpServers` 키) |
| **Antigravity** | `~/.MiniMax/antigravity.json` (가정, 하네스별 확인 필요) | `<root>/.antigravity/mcp.json` | JSON (`mcpServers` 키) |
| **MiniMax Code** | `~/.MiniMax/mcp.json` 또는 `~/.MiniMax/config.json` 의 `mcp_servers` | `<root>/.MiniMax/mcp.json` | JSON (`mcp_servers` 키) |

## 5. 자동 심기 (`bootstrap --enable-mcp`)

`bootstrap_workflow_kit.py` 가 한 번에 해결한다.

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root <project_root> \
  --project-slug <slug> \
  --project-name "<name>" \
  --harness codex --harness opencode --harness gemini-cli \
  --harness antigravity --harness minimax-code \
  --adoption-mode existing \
  --copy-core-docs \
  --enable-mcp \
  --mcp-bridge jsonrpc-bridge  # or stdio-sdk
```

효과:
- 각 하네스별 MCP config 파일을 `<root>/` 아래에 emit (위 표의 "프로젝트 로컬 위치" 열)
- 이미 존재하면 `--force` 로 덮어쓰기
- manifest 의 `generated_harness_files` 에 `*_mcp_config` 키로 경로 기록

## 6. 수동 적용 (글로벌)

자동 심기는 프로젝트 로컬 파일만 만든다. 글로벌 (사용자 홈) 설정에 등록하려면 bootstrap 후 다음 한 줄을 각 글로벌 config 에 병합한다.

### 6.1 Codex (`~/.codex/config.toml`)

```toml
[mcp_servers.standardAiWorkflowReadOnly]
command = "python3"
args = ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"]
PYTHONPATH = "/ABSOLUTE/PATH/TO/standard_ai_workflow/workflow-source"
STANDARD_AI_WORKFLOW_ROOT = "/ABSOLUTE/PATH/TO/<project_root>"

[mcp_servers.standardAiWorkflowReadOnly.tool_descriptions]
workflow_kit.read_only = "Read-only MCP tools (latest_backlog, check_doc_metadata, ...) for the Standard AI Workflow kit."
```

### 6.2 OpenCode (글로벌 `opencode.json`)

```json
{
  "mcp": {
    "standardAiWorkflowReadOnly": {
      "type": "local",
      "command": "python3",
      "args": ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"],
      "env": {
        "PYTHONPATH": "/ABSOLUTE/PATH/TO/standard_ai_workflow/workflow-source",
        "STANDARD_AI_WORKFLOW_ROOT": "/ABSOLUTE/PATH/TO/<project_root>"
      },
      "timeout": 30000
    }
  }
}
```

### 6.3 Gemini CLI (`~/.gemini/settings.json`)

```json
{
  "mcpServers": {
    "standardAiWorkflowReadOnly": {
      "command": "python3",
      "args": ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"],
      "env": {
        "PYTHONPATH": "/ABSOLUTE/PATH/TO/standard_ai_workflow/workflow-source",
        "STANDARD_AI_WORKFLOW_ROOT": "/ABSOLUTE/PATH/TO/<project_root>"
      },
      "trust": true,
      "includeTools": [
        "latest_backlog",
        "check_doc_metadata",
        "check_doc_links",
        "suggest_impacted_docs"
      ]
    }
  }
}
```

### 6.4 Antigravity

Antigravity 의 정확한 글로벌 설정 경로는 하네스 문서를 참조. 일반적으로 `~/.antigravity/config.json` 에 `mcpServers` 키로 등록 (Gemini CLI 와 동일 스키마). bootstrap 이 emit 한 `.antigravity/mcp.json` 의 `mcpServers` 블록을 그대로 복사.

### 6.5 MiniMax Code (`~/.MiniMax/mcp.json` 또는 `~/.MiniMax/config.json`)

`mcp_servers` 키 사용. bootstrap 출력 (`<root>/.MiniMax/mcp.json`) 을 그대로 `~/.MiniMax/mcp.json` 으로 심거나 symlink.

```bash
mkdir -p ~/.MiniMax
ln -sf /ABSOLUTE/PATH/TO/<project_root>/.MiniMax/mcp.json ~/.MiniMax/mcp.json
```

또는 전역 config 가 `~/.MiniMax/config.json` 인 경우, 그 파일의 `mcp_servers` 블록에 bootstrap 출력의 `standardAiWorkflowReadOnly` 항목을 복사.

## 7. 트러블슈팅

### 7.1 `PYTHONPATH` 가 잘못 잡혀 `ModuleNotFoundError: workflow_kit`

- bootstrap 이 emit 한 env 의 `PYTHONPATH` 가 상대 경로(`workflow-source`)인 경우, 하네스가 다른 cwd 로 띄우면 실패.
- **해결**: 글로벌 설정으로 옮길 때 절대 경로로 바꾼다. 예: `PYTHONPATH = "/Users/yklee/repos/standard_ai_workflow/workflow-source"`

### 7.2 `Connection closed` (stdio-sdk 만)

- `mcp 1.27.0` 의 `CallToolResult(structuredContent=...)` API 와 SDK 시그니처 불일치.
- **해결 (임시)**: `--mcp-bridge jsonrpc-bridge` 로 fallback. 정식 SDK 회귀는 후속 TASK.

### 7.3 `tools/list` 가 비어 있음

- bridge 가 `initialize` 호출 없이 `tools/list` 만 들어오는 경우.
- **해결**: jsonrpc-bridge 의 stdio 모드는 첫 줄에서 session 을 만들고 그 이후로만 응답한다. 하네스가 `initialize` 후 곧장 `tools/list` 를 보내는지 로그로 확인.

### 7.4 한국어 `description` 이 깨져 보임

- `ensure_ascii=False` 로 dump 했으므로 UTF-8 그대로 전송. 하네스가 latin-1 로 읽으면 깨진다.
- **해결**: 하네스 로그 인코딩을 UTF-8 로 강제. Codex/OpenCode 는 기본 UTF-8. Gemini CLI 는 `--encoding utf-8` 플래그 필요할 수 있음.

### 7.5 `transport_ready=false` 표시

- 정상. read-only bundle 이 draft 단계라 명시적으로 false 로 둠.
- **해결**: 도구 호출 (`tools/call`) 이 정상 동작하면 OK. `tools/list` 의 `_meta.transport_ready` 가 false 여도 실제 도구는 사용 가능.

## 8. 다음 단계 / 후속 TASK

- [ ] `check_read_only_mcp_sdk_stdio.py` 의 `Connection closed` 원인 추적 + 수정 (TASK-V051-004 후보)
- [ ] 5개 하네스별 실제 smoke 테스트 (TASK-V051-005 후보)
- [ ] `workflow_kit` 정식 패키지 배포 후 `pip install standard-ai-workflow` 한 줄로 import 가능하게 (TASK-V051-006 후보)

## 다음에 읽을 문서

- [harnesses/codex/apply_guide.md](../harnesses/codex/apply_guide.md)
- [harnesses/opencode/apply_guide.md](../harnesses/opencode/apply_guide.md)
- [harnesses/gemini-cli/apply_guide.md](../harnesses/gemini-cli/apply_guide.md)
- [harnesses/antigravity/apply_guide.md](../harnesses/antigravity/apply_guide.md)
- [harnesses/minimax-code/apply_guide.md](../harnesses/minimax-code/apply_guide.md)
- [examples/mcp_config_examples/](../examples/mcp_config_examples/)
