# Antigravity Harness Apply Guide

- 문서 목적: Antigravity 하네스에서 표준 AI 워크플로우를 운영할 때의 단계별 절차와 MCP 설치 방법을 정리한다.
- 범위: 사전 점검, overlay 파일 적용, 설정 적용, 로컬 MCP 설치, 트러블슈팅
- 대상 독자: 표준 AI 워크플로우를 Antigravity 로 처음 도입하는 개발자
- 상태: beta
- 최종 수정일: 2026-06-05
- 관련 문서: [`./README.md`](./README.md), [`./overlay_spec.md`](./overlay_spec.md), [`../../core/workflow_adoption_entrypoints.md`](../../core/workflow_adoption_entrypoints.md), [`../../core/mcp_installation_by_harness.md`](../../core/mcp_installation_by_harness.md)

## 1. 사전 점검

- Python 3.11+
- Antigravity CLI 설치 확인
- 저장소에 표준 워크플로우 패키지(`ai-workflow/`) 가 bootstrap 되었거나, 도입 직전 단계

## 2. overlay 파일 적용

`bootstrap_workflow_kit.py --harness antigravity` 실행 시 다음 파일이 생성된다.

```
<project_root>/
├── ANTIGRAVITY.md
└── .antigravity/
    └── mcp.json   # --enable-mcp 사용 시
```

## 3. 설정 적용

Antigravity 의 글로벌 설정 위치는 하네스 문서를 확인하되, 일반적으로 `~/.antigravity/config.json` 에 `mcpServers` 키로 MCP 를 등록한다. JSON 스키마는 Gemini CLI 와 호환 (`command`, `args`, `env`, `trust`, `includeTools`).

## 4. 로컬 MCP 설치 (`--enable-mcp`)

### 4.1 자동 심기

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root <project_root> \
  --project-slug <slug> --project-name "<name>" \
  --harness antigravity --adoption-mode existing --copy-core-docs \
  --enable-mcp
```

`<root>/.antigravity/mcp.json` 생성. 글로벌에 등록:

```bash
mkdir -p ~/.antigravity
cp <project_root>/.antigravity/mcp.json ~/.antigravity/mcp.json
# 또는
jq -s '.[0].mcpServers * .[1].mcpServers' \
   ~/.antigravity/mcp.json <project_root>/.antigravity/mcp.json \
   > ~/.antigravity/mcp.json.new && mv ~/.antigravity/mcp.json.new ~/.antigravity/mcp.json
```

### 4.2 전역에 적용

`~/.antigravity/config.json` 의 `mcpServers` 블록에 `standardAiWorkflowReadOnly` 항목을 옮긴다. 예:

```json
{
  "mcpServers": {
    "standardAiWorkflowReadOnly": {
      "type": "stdio",
      "command": "python3",
      "args": ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"],
      "env": {
        "PYTHONPATH": "/ABSOLUTE/PATH/TO/standard_ai_workflow/workflow-source",
        "STANDARD_AI_WORKFLOW_ROOT": "/ABSOLUTE/PATH/TO/<project_root>"
      }
    }
  }
}
```

자세한 가이드: [`../../core/mcp_installation_by_harness.md`](../../core/mcp_installation_by_harness.md), 예시: [`../../examples/mcp_config_examples/antigravity-mcp.json`](../../examples/mcp_config_examples/antigravity-mcp.json)

## 5. 트러블슈팅

| 증상 | 원인 | 해결 |
| --- | --- | --- |
| `ANTIGRAVITY.md` 가 로드되지 않음 | Antigravity 의 신뢰 경로 밖 | 프로젝트 루트에 있는지 확인 |
| `mcpServers` 가 연결되지 않음 | `PYTHONPATH` 가 `workflow-source` 를 가리키지 않음 | 절대 경로로 보정 |
| 한국어 보고가 영어로 나옴 | 시스템 프롬프트 설정 변경 | Antigravity 의 `~/.antigravity/profile.json` 에 `language: "ko-KR"` 추가 |

## 6. 다음 단계

- 첫 적용이 끝나면 `workflow-source/harnesses/antigravity/README.md` 와 본 가이드를 함께 검토한다.
- 운영 패턴이 안정되면 `pilot_adoption_record_template.md` 로 도입 기록을 남긴다.
