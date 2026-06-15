---
type: concept
status: active
last_ingested_from: workflow-source/core/workflow_harness_distribution.md + workflow-source/harnesses/
related_pages: [concepts/agent-topology, topics/harness-distribution-model]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# Harness Distribution: 6-Harness Overlay Model

## TL;DR

| 하네스 | 타겟 도구 | 진입 파일 | config / 설정 | 에이전트 토폴로지 |
|---|---|---|---|---|
| **Codex** | Codex CLI | `AGENTS.md` | `.codex/config.toml.example` | in-doc main/worker (분리 약함) |
| **OpenCode** | OpenCode | `AGENTS.md` | `opencode.json` + `.opencode/agents/*` | orchestrator + 4 worker (worker/doc/code/validation) |
| **Gemini CLI** | Gemini CLI | `GEMINI.md` | (없음) | `invoke_agent` 로 main+worker 분리 |
| **Antigravity** | Antigravity | `ANTIGRAVITY.md` | (없음) | browser sub-agent 위임 |
| **MiniMax Code** | MiniMax Code | `AGENTS.md` + `MiniMax.md` | `MiniMax_config.example.json` + `.MiniMax/agents/*` | orchestrator + 4 worker (Codex/OpenCode 와 동일 4종) |
| **pi-dev** | Pi Coding Agent | `AGENTS.md` + `SYSTEM.md` | (없음) | 페르소나 기반, 별도 worker 디렉터리 없음 |

v0.6.3-beta 시점. 공통 정책 원문은 `workflow-source/core/` 와 `ai-workflow/memory/active/` 에 두고, 하네스별 파일은 **공통 문서를 가리키는 오버레이** 로만 유지한다 (중복 본문 금지).

## Harness Matrix

| 하네스 | 진입 파일 | 오버레이 패턴 | MCP transport (opt-in) | 등록 버전 |
|---|---|---|---|---|
| **codex** | `AGENTS.md` | root `AGENTS.md` + `.codex/config.toml.example` (snippet). 메인/워커 분리는 `AGENTS.md` 본문에 명시 | `.codex/mcp.toml` (read-only draft, `manual_review_only`) | v0.6.3-beta |
| **opencode** | `AGENTS.md` + `opencode.json` | `opencode.json` `instructions` → 공통 문서. `.opencode/agents/` 에 orchestrator + 4 worker (worker / doc / code / validation) | `mcp.opencode.json` (read-only draft, `manual_review_only`) | v0.6.3-beta |
| **gemini-cli** | `GEMINI.md` | 단일 진입점. `invoke_agent` 로 main+worker 분리. 시스템 프롬프트보다 우선 | `.gemini/mcp.json` (선택) | v0.6.3-beta |
| **antigravity** | `ANTIGRAVITY.md` | 단일 진입점. bootstrap 시 `ANTIGRAVITY.md` 외 파일 생성 안 함 | `antigravity.mcp.json` (선택) | v0.6.3-beta |
| **minimax-code** | `AGENTS.md` + `MiniMax.md` | Codex/OpenCode 와 동일한 4-worker 패턴을 `.MiniMax/agents/` 로 배포. `MiniMax_config.example.json` → `.minimax/config.json` 으로 복사해 사용 | `.MiniMax/mcp.json` (read-only draft bridge, 시크릿은 환경 변수 분리) | v0.6.3-beta |
| **pi-dev** | `AGENTS.md` + `SYSTEM.md` | 페르소나 + 운영 원칙 보강. `ai-workflow/memory/active/state.json` 을 strict source of truth 로 사용 | (없음) | v0.6.3-beta |

오버레이 원칙 (전 하네스 공통):
- 정책 본문 중복 금지. 공통 문서 경로만 참조.
- 하네스 전역 설정은 **기본값만** 제공, 프로젝트별 규칙은 local 문서가 우선.
- 새 하네스 추가는 `bootstrap_lib.harnesses.HARNESS_SPECS` 한 줄 + `bootstrap_workflow_kit.py` 의 `register_harness_builder` 한 줄로 끝난다 (legacy `HARNESS_DEFINITIONS` 는 v0.5.8 부터 deprecated).
- 에이전트 토폴로지 상세: [[concepts/agent-topology]].

## Bundle Structure

`dist/harnesses/<harness>/<version>/` 한 디렉터리가 하나의 배포 단위. 예: `dist/harnesses/opencode/v0.6.3-beta/`.

| 파일 / 디렉터리 | 역할 |
|---|---|
| `bundle/AGENTS.md` (또는 하네스별 진입점) | 하네스가 실제 읽는 루트 지침. 공통 문서로 연결 |
| `bundle/opencode.json` / `bundle/.codex/config.toml.example` / `bundle/.MiniMax/config.json` | 하네스별 설정 파일 |
| `bundle/.opencode/agents/` 또는 `bundle/.MiniMax/agents/` | orchestrator + worker 페르소나 (4종) |
| `bundle/.opencode/skills/standard-ai-workflow/SKILL.md` | project-local skill 진입점 |
| `bundle/ai-workflow/README.md` | runtime layer 진입 안내 |
| `bundle/ai-workflow/core/{global_workflow_standard, workflow_adoption_entrypoints, workflow_skill_catalog}.md` | 공통 runtime 정책 문서 (3종 minimal core) |
| `bundle/ai-workflow/memory/active/PROJECT_PROFILE.md` | 프로젝트 프로파일 |
| `bundle/ai-workflow/memory/active/state.json` | 세션 캐시 (state of truth 아님) |
| `bundle/ai-workflow/memory/active/session_handoff.md` | 세션 인계 |
| `bundle/ai-workflow/memory/active/work_backlog.md` + `backlog/YYYY-MM-DD.md` | 작업 백로그 |
| `manifest.json` | 머신 파싱 가능. `included_files`, `recommended_entrypoints`, `excluded_by_default`, `deferred_release_items` 기록 |
| `PACKAGE_CONTENTS.md` | 사람이 읽는 패키지 안내 (한글) |
| `APPLY_GUIDE.md` | 대상 저장소 적용 절차 (복사 경로 매핑, 첫 세션 읽기 순서) |
| `standard-ai-workflow-<harness>-<version>.zip` | 위 bundle + manifest + 문서를 묶은 배포 자산 |

`optimization_profile = agent_runtime_minimal` 이 기본. `excluded_by_default`: developer source docs / global snippet examples / draft MCP reference assets. opt-in: `--include-source-docs`, `--include-global-snippets`.

## Export Workflow

`workflow-source/scripts/export_harness_package.py` 가 단일 진입점. `workflow_kit.__version__` 을 기본 버전으로 stamp.

| 단계 | 동작 | 산출물 |
|---|---|---|
| 1. args parse | `--harness` (다중 가능, choices: `codex` / `opencode` / `gemini-cli` / `pi-dev` / `antigravity`), `--include-source-docs`, `--include-global-snippets` | `argparse.Namespace` |
| 2. workflow-source → runtime 변환 | `DocTransformer` 로 `workflow-source/core/` 의 원본을 runtime `ai-workflow/core/` 형식으로 가공 | 변환된 core 문서 3종 |
| 3. overlay 합성 | `bootstrap_workflow_kit.py` 의 harness builder 를 호출해 `AGENTS.md`, `opencode.json`, `.opencode/agents/*` 등 오버레이 생성 | 하네스별 overlay 파일 |
| 4. bundle 복사 | `bundle/` 아래에 runtime + overlay 를 모음. `workflow-source/scripts/bootstrap_lib/harnesses/__init__.py` 의 `HARNESS_SPECS` 단일 source of truth 사용 | `dist/harnesses/<harness>/<version>/bundle/` |
| 5. manifest 기록 | `package_name`, `package_version`, `release_focus`, `optimization_profile`, `recommended_entrypoints`, `excluded_by_default`, `deferred_release_items` | `manifest.json` |
| 6. 문서 emit | `PACKAGE_CONTENTS.md`, `APPLY_GUIDE.md` 자동 생성 (각 export 루트) | 한글 apply 안내 2종 |
| 7. zip 압축 | `standard-ai-workflow-<harness>-<version>.zip` | GitHub release asset 후보 |

packaging smoke 은 v0.5.8+ `tools/check_packaging.py` 가 자동 검증. `MiniMax Code` 는 `export_harness_package.py` 의 `SUPPORTED_HARNESSES` 에 아직 포함되지 않으므로 (소스 line 23 기준 5종만 등록), bootstrap 직접 호출로 동기화한다. (별도 pattern page 없음).

## Related Decisions

- [[concepts/agent-topology]] — orchestrator + 4 worker 페르소나 정의 (OpenCode / MiniMax Code 가 그대로 가져감)
- [[topics/harness-distribution-model]] — 6-harness 모델의 상위 topic 페이지
- [[entities/harness-overlay-codex]] — Codex 오버레이 entity (root `AGENTS.md` + `.codex/config.toml.example`)
- [[entities/harness-overlay-opencode]] — OpenCode 오버레이 entity (`AGENTS.md` + `opencode.json` + `.opencode/agents/*`)
- [[entities/harness-overlay-minimax-code]] — MiniMax Code 오버레이 entity (`MiniMax.md` + `.MiniMax/agents/*` + `MiniMax_config.example.json`)
- ADR-001 (3-Layer Source/Runtime/Project Docs) — 하네스 export 가 runtime layer 만 건드리는 근거

## References

- 본문 출처: [workflow-source/core/workflow_harness_distribution.md](../../../workflow-source/core/workflow_harness_distribution.md)
- 하네스 허브: [workflow-source/harnesses/README.md](../../../workflow-source/harnesses/README.md)
- per-harness overlay spec: `workflow-source/harnesses/{codex,opencode,gemini-cli,antigravity,minimax-code,pi-dev}/README.md`
- export 스크립트: [workflow-source/scripts/export_harness_package.py](../../../workflow-source/scripts/export_harness_package.py)
- 스크립트 허브: [workflow-source/scripts/README.md](../../../workflow-source/scripts/README.md) (`export_harness_package.py` 섹션)
- bootstrap 스크립트: [workflow-source/scripts/bootstrap_workflow_kit.py](../../../workflow-source/scripts/bootstrap_workflow_kit.py)
- 하네스 레지스트리 (single source of truth): [workflow-source/scripts/bootstrap_lib/harnesses/__init__.py](../../../workflow-source/scripts/bootstrap_lib/harnesses/__init__.py)
- 적용 도구: [workflow-source/scripts/apply_harness_update.py](../../../workflow-source/scripts/apply_harness_update.py)
- 샘플 bundle (Codex): [dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md)
- 샘플 bundle (OpenCode): [dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md)
- 샘플 manifest: [dist/harnesses/opencode/v0.6.3-beta/manifest.json](../../../dist/harnesses/opencode/v0.6.3-beta/manifest.json)
- 샘플 apply guide: [dist/harnesses/opencode/v0.6.3-beta/APPLY_GUIDE.md](../../../dist/harnesses/opencode/v0.6.3-beta/APPLY_GUIDE.md)
