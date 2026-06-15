---
type: concept
status: active
last_ingested_from: docs/PROJECT_PROFILE.md + workflow-source/core/mcp_installation_by_harness.md
related_pages: [concepts/orchestrator-subagent-pattern, patterns/harness-overlay]
created: 2026-06-12
updated: 2026-06-12
---

# MCP Transport

- 문서 목적: standard_ai_workflow 가 하네스에 심는 read-only MCP 서버의 transport 두 종류 (jsonrpc-bridge, stdio-sdk) 의 차이, 결정, 실패 모드, wiki layer 와의 상호작용을 정리한다.
- 범위: transport 비교, key decisions, failure mode, read-only access pattern
- 관련 결정: ADR-004 (wiki layer 도입, P1 review 대상)
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 권장 transport | `jsonrpc-bridge` (default, stable, v0.5.0+) |
| 2 | 실험 transport | `stdio-sdk` (정식 MCP Python SDK, `Connection closed` 회귀 있음) |
| 3 | 엔트리포인트 (jsonrpc) | `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines` |
| 4 | 엔트리포인트 (sdk) | `python3 -m workflow_kit.server.read_only_mcp_sdk --stdio-sdk` |
| 5 | wiki layer 접근 | MCP server read-only 도구 (`latest_backlog`, `check_doc_metadata`, `check_doc_links`, `suggest_impacted_docs`) 로만 |
| 6 | wiki 가 MCP 를 호출 | ❌ 안 함. wiki 는 파일 직접 read/write (R1) |

## §2 두 transport 비교  {#s2-transport-comparison}

`bootstrap_workflow_kit.py` 의 `--mcp-bridge` 플래그로 선택한다. 기본값은 `jsonrpc-bridge`.

| 이름 | 엔트리포인트 | 상태 | 권장 시나리오 |
|---|---|---|---|
| `jsonrpc-bridge` | `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines` | stable (v0.5.0) | 일반 도입. 항상 동작. `transport_ready=false` 지만 `tools/list` / `tools/call` round-trip 정상 |
| `stdio-sdk` | `python3 -m workflow_kit.server.read_only_mcp_sdk --stdio-sdk` | experimental | 정식 MCP SDK 호환이 필요한 경우. 현재 `check_read_only_mcp_sdk_stdio.py` 가 `Connection closed` 로 fail |

권장: 처음 도입 시 `jsonrpc-bridge` 로 시작. SDK 호환이 꼭 필요하면 별도 TASK 로 추적하면서 점진 전환.

## §3 공통 환경 변수  {#s3-env}

| 변수 | 용도 | 글로벌 심을 때 |
|---|---|---|
| `PYTHONPATH` | `workflow_kit` 모듈 경로. bootstrap 이 자동 설정 (`workflow-source` 기준 상대 경로) | 절대 경로 권장 |
| `STANDARD_AI_WORKFLOW_ROOT` | `ai-workflow/`, `workflow-source/` 위치 hint | 절대 경로 권장 |

글로벌로 심을 때 `PYTHONPATH` 가 상대 경로면 하네스가 다른 cwd 로 띄울 때 `ModuleNotFoundError: workflow_kit` 발생. 절대 경로로 교체.

## §4 Key Decisions  {#s4-key-decisions}

| ID | 결정 | 근거 |
|---|---|---|
| **KD-1** | default transport = `jsonrpc-bridge` | round-trip 안정성, `tools/list` + `tools/call` 모두 동작 |
| **KD-2** | `stdio-sdk` 는 experimental 로 둠 | `mcp 1.27.0` 의 `CallToolResult(structuredContent=...)` API 와 SDK 시그니처 불일치로 `Connection closed` |
| **KD-3** | wiki layer 는 MCP 를 호출하지 않음 | R1 (파일 직접 read/write), A1 (page atomicity). wiki 는 `ai-workflow/wiki/` 안에서만 동작 |
| **KD-4** | read-only bundle 만 노출 | 위키는 read-only 접근. `latest_backlog`, `check_doc_metadata`, `check_doc_links`, `suggest_impacted_docs` 4종 |

### §4.1 transport 선택 의사결정  {#s4-1-transport-decision}

```bash
# default (jsonrpc-bridge)
python3 -m bootstrap_lib --target-root <project_root> \
  --project-slug <slug> --harness opencode \
  --enable-mcp

# 실험적 stdio-sdk
python3 -m bootstrap_lib --target-root <project_root> \
  --project-slug <slug> --harness opencode \
  --enable-mcp --mcp-bridge stdio-sdk   # 회귀 감수
```

권장 흐름: `jsonrpc-bridge` 도입 → `stdio-sdk` 필요 별도 TASK 추적 → 회귀 fix 후 전환. ADR-004 의 wiki layer 도입과 무관하게 transport 선택은 별도 결정으로 유지.

## §5 Failure Modes  {#s5-failure-modes}

| ID | 증상 | 원인 | 해결 |
|---|---|---|---|
| **F-1** | `ModuleNotFoundError: workflow_kit` | `PYTHONPATH` 잘못 | 글로벌 설정으로 옮길 때 절대 경로로 교체 |
| **F-2** | `Connection closed` (stdio-sdk 한정) | `mcp 1.27.0` 의 `CallToolResult(structuredContent=...)` API 불일치 | `--mcp-bridge jsonrpc-bridge` 로 fallback. 정식 회귀는 후속 TASK |
| **F-3** | `tools/list` 비어 있음 | bridge 가 `initialize` 호출 없이 `tools/list` 만 받음 | jsonrpc-bridge stdio 모드는 첫 줄에서 session 생성. 하네스가 `initialize` 후 곧장 `tools/list` 보내는지 로그 확인 |
| **F-4** | 한국어 `description` 깨짐 | 하네스가 latin-1 로 읽음 | 하네스 로그 인코딩을 UTF-8 로 강제. Codex/OpenCode 기본 UTF-8. Gemini CLI 는 `--encoding utf-8` |
| **F-5** | `transport_ready=false` 표시 | 정상. read-only bundle 이 draft 단계 | 도구 호출 (`tools/call`) 이 정상 동작하면 OK. `_meta.transport_ready` 가 false 여도 실제 도구 사용 가능 |

### §5.1 비대화형 picker (v0.5.8+)  {#s5-1-picker}

v0.5.8 부터 `--harness` 미지정 + TTY 환경이면 interactive picker 자동 실행. 비대화형 (CI, 파이프라인) 에서 `--harness` 미지정 시 silent 0 overlay 없이 `SystemExit(1)` + 6개 harness 목록 제시 후 fail-fast.

```bash
# 비대화형 권장 호출
python3 -m bootstrap_lib --target-root "$REPO" \
  --project-slug "$SLUG" \
  --harness opencode \
  --no-interactive \
  --adoption-mode existing
```

## §6 Wiki Layer 와의 상호작용  {#s6-wiki-interaction}

wiki layer (`ai-workflow/wiki/`) 는 MCP server 를 호출하지 않는다. 두 layer 는 다른 역할이다.

| Layer | 위치 | read | write | 트리거 |
|---|---|---|---|---|
| **MCP server (read-only)** | 외부 프로세스 (각 하네스가 spawn) | workflow_kit 도구 4종 | ❌ | 하네스 tool call |
| **wiki layer** | `ai-workflow/wiki/` | 파일 직접 read | 페이지 atomic write (R2) | `wiki-ingest` (P2 부터) |

### §6.1 read-only access pattern  {#s6-1-readonly}

MCP server 가 노출하는 read-only 도구 4종은 wiki 운영자가 ingest / lint 시 참고용으로만 사용한다.

| 도구 | 용도 | wiki 단계 |
|---|---|---|
| `latest_backlog` | 최근 작업 백로그 조회 | ingest §2 단계에서 후보 페이지 식별 보조 |
| `check_doc_metadata` | 문서 frontmatter / 메타 검증 | lint §4 (stale, missing) 검사 |
| `check_doc_links` | 문서 내부 link 검증 | lint §4 (broken backlinks) 검사 |
| `suggest_impacted_docs` | 변경 영향 받는 문서 후보 | ingest §3 (multi-page update) 후보 선정 |

wiki 자체는 위 도구를 직접 호출하지 않고, orchestrator 또는 `wiki-ingest` skill 이 호출한 결과를 받아 페이지를 갱신한다. 이 분리 정책은 ADR-004 의 wiki layer 분리 정신 (R1, "Runtime layer 안에 흡수, Project docs 침범 안 함") 과 일치한다.

### §6.2 wiki 가 MCP 를 호출하지 않는 이유  {#s6-2-no-direct-call}

1. R1 (Wiki Location): wiki 는 `ai-workflow/wiki/` 안에서만 read/write. 외부 프로세스 호출 = 경계 침범
2. A1 (페이지 분할 동시 작업 회피): MCP 호출 결과로 페이지 갱신 = sub-agent + MCP 결합 = 추적 어려움
3. ADR-004 의 layer 분리 정신: wiki 는 Runtime layer 의 일종이지, MCP transport 의 consumer 가 아님

## §7 다음에 읽을 문서  {#s7-next}

- 분산 규칙: [../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md](../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md)
- 운영 헌법: [../SCHEMA.md](../SCHEMA.md)
- 관련 개념: [./orchestrator-subagent-pattern.md](./orchestrator-subagent-pattern.md)
- 원본 소스: [../../workflow-source/core/mcp_installation_by_harness.md](../../workflow-source/core/mcp_installation_by_harness.md)

## §8 Revision Log  {#s8-revision-log}

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-06-12 | 0.1.0 | 초안. 두 transport 비교, KD-1~KD-4, F-1~F-5, wiki read-only access pattern. P1 bootstrap 의 일부 | Sisyphus (orchestrator) |
