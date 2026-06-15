---
type: topic
status: active
last_ingested_from: workflow-source/core/workflow_harness_distribution.md + dist/harnesses/*/v0.6.3-beta/
related_pages: [concepts/harness-distribution, concepts/agent-topology, entities/standard-ai-workflow, entities/harness-overlay-codex, entities/harness-overlay-opencode, entities/harness-overlay-gemini-cli, entities/harness-overlay-antigravity, entities/harness-overlay-minimax-code, entities/harness-overlay-pi-dev]
created: 2026-06-12
updated: 2026-06-12
---

# Harness Distribution Model: 6 × {code, doc, manifest} × version

## TL;DR

The Standard AI Workflow distribution model ships **6 harness overlays** as **per-harness, per-version bundles** in `dist/harnesses/<harness>/<version>/`, each emitting one runtime bundle, one `manifest.json`, one `PACKAGE_CONTENTS.md`, one `APPLY_GUIDE.md`, and one GitHub release zip — see [[concepts/harness-distribution]] for the per-harness overlay rules and [[entities/standard-ai-workflow]] for the parent project entity.

| 하네스 | 패키지명 | dist 산출물 | 비고 |
|---|---|---|---|
| codex | `standard-ai-workflow-codex` | `dist/harnesses/codex/v0.6.3-beta/` | zip 자산 |
| opencode | `standard-ai-workflow-opencode` | `dist/harnesses/opencode/v0.6.3-beta/` | zip 자산 |
| gemini-cli | `standard-ai-workflow-gemini-cli` | `dist/harnesses/gemini-cli/v0.6.3-beta/` | zip 자산 |
| antigravity | `standard-ai-workflow-antigravity` | `dist/harnesses/antigravity/v0.6.3-beta/` | zip 자산 |
| minimax-code | `standard-ai-workflow-minimax-code` | (not produced — bootstrap 직접 동기화) | [[entities/harness-overlay-minimax-code]] |
| pi-dev | `standard-ai-workflow-pi-dev` | `dist/harnesses/pi-dev/v0.6.3-beta/` | zip 자산 |

## Harness Matrix

6개 하네스의 진입점·토폴로지·MCP transport·버전·산출물 상태. 상세 페르소나는 [[concepts/agent-topology]].

| 하네스 | entry file | agent topology | MCP transport (opt-in) | version | dist status |
|---|---|---|---|---|---|
| **codex** | `AGENTS.md` | task-only orchestrator + in-doc bounded worker ([[entities/harness-overlay-codex]]) | `.codex/mcp.toml` (read-only draft, `manual_review_only`) | v0.6.3-beta | shipped (zip) |
| **opencode** | `AGENTS.md` + `opencode.json` | orchestrator + 4 worker (worker / doc / code / validation) ([[entities/harness-overlay-opencode]]) | `mcp.opencode.json` (read-only draft, `manual_review_only`) | v0.6.3-beta | shipped (zip) |
| **gemini-cli** | `GEMINI.md` | main+worker split via `invoke_agent` ([[entities/harness-overlay-gemini-cli]]) | `.gemini/mcp.json` (선택) | v0.6.3-beta | shipped (zip) |
| **antigravity** | `ANTIGRAVITY.md` | browser sub-agent 위임 ([[entities/harness-overlay-antigravity]]) | `antigravity.mcp.json` (선택) | v0.6.3-beta | shipped (zip) |
| **minimax-code** | `AGENTS.md` + `MiniMax.md` | orchestrator + 4 worker (Codex/OpenCode 와 동일 4종) ([[entities/harness-overlay-minimax-code]]) | `.MiniMax/mcp.json` (read-only draft bridge) | v0.6.3-beta | **not produced** (bootstrap 직접) |
| **pi-dev** | `AGENTS.md` + `SYSTEM.md` | persona 기반, 별도 worker 디렉터리 없음 ([[entities/harness-overlay-pi-dev]]) | (없음) | v0.6.3-beta | shipped (zip) |

공통 제약: 정책 본문 중복 금지 (오버레이는 경로만 참조) / 하네스 전역 설정은 기본값만 / `state.json` 은 캐시 (source-of-truth 아님).

### Topology tiers

| tier | 하네스 | worker 분리 방식 | sub-agent 디렉터리 |
|---|---|---|---|
| 4-worker fan-out | opencode, minimax-code | orchestrator → doc/code/validation | `.opencode/agents/`, `.MiniMax/agents/` |
| in-doc bounded | codex | 본문 내 main/worker 섹션 분리 | 없음 |
| runtime primitive | gemini-cli, antigravity | `invoke_agent` / browser sub-agent | runtime 호출 |
| persona only | pi-dev | SYSTEM.md 페르소나 | 없음 |

## Bundle Anatomy

`dist/harnesses/<harness>/<version>/` 한 디렉터리가 단일 배포 단위. 3개 레이어 × 코드·문서·매니페스트로 직교.

| 파일 / 디렉터리 | 종류 | 역할 |
|---|---|---|
| `bundle/AGENTS.md` (또는 하네스 진입점) | code | 하네스가 실제 읽는 루트 지침. 공통 문서 경로만 참조 |
| `bundle/opencode.json` / `bundle/.codex/config.toml.example` / `bundle/.MiniMax/config.json` | code | 하네스별 설정 파일 (snippet 또는 example) |
| `bundle/.opencode/agents/` 또는 `bundle/.MiniMax/agents/` | code | orchestrator + 4 worker 페르소나 (worker / doc / code / validation) |
| `bundle/.opencode/skills/standard-ai-workflow/SKILL.md` | code | project-local skill 진입점 |
| `bundle/ai-workflow/README.md` | doc | runtime layer 진입 안내 ([[entities/ai-workflow-runtime]]) |
| `bundle/ai-workflow/core/{global_workflow_standard, workflow_adoption_entrypoints, workflow_skill_catalog}.md` | doc | 공통 runtime 정책 문서 3종 (minimal core) |
| `bundle/ai-workflow/memory/active/PROJECT_PROFILE.md` | doc | 프로젝트 프로파일 |
| `bundle/ai-workflow/memory/active/state.json` | doc | 세션 캐시 (state of truth 아님) |
| `bundle/ai-workflow/memory/active/session_handoff.md` | doc | 세션 인계 |
| `bundle/ai-workflow/memory/active/work_backlog.md` + `backlog/YYYY-MM-DD.md` | doc | 작업 백로그 |
| `manifest.json` | manifest | 머신 파싱 가능. `package_name`, `package_version`, `release_focus`, `optimization_profile`, `recommended_entrypoints`, `excluded_by_default`, `deferred_release_items` |
| `PACKAGE_CONTENTS.md` | manifest | 사람이 읽는 한글 패키지 안내 |
| `APPLY_GUIDE.md` | manifest | 대상 저장소 적용 절차 (복사 경로 매핑, 첫 세션 읽기 순서) |
| `standard-ai-workflow-<harness>-<version>.zip` | manifest | 위 bundle + manifest + 문서를 묶은 GitHub release 자산 |

기본 `optimization_profile = agent_runtime_minimal`. `excluded_by_default`: developer source docs / global snippet examples / draft MCP reference assets. opt-in 플래그: `--include-source-docs`, `--include-global-snippets`. Codex/OpenCode 실증 자료: [dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md), [dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md).

### Code × Doc × Manifest cross-cut

| axis | 책임 | harness 의존성 |
|---|---|---|
| code | `AGENTS.md` / `opencode.json` / `.opencode/agents/*` — 하네스 runtime 이 실행 시 직접 소비 | 하네스별 상이 (full, snippet, example 3종) |
| doc | `ai-workflow/{README.md, core/*, memory/active/*}` — 공통 runtime 정책 + 세션 상태 | 6 하네스 공통 (3종 core + 5종 memory) |
| manifest | `manifest.json` + `PACKAGE_CONTENTS.md` + `APPLY_GUIDE.md` + `.zip` — 머신/사람/릴리스 자산 | 6 하네스 공통 (필드/문서 동일, 값만 다름) |

## Version Cadence

| 측면 | 값 |
|---|---|
| 포맷 | `v0.X.Y-beta` (예: `v0.6.3-beta`) |
| suffix 규칙 | pre-release `-beta` 필수 |
| stamp 출처 | `workflow_kit.__version__` (CLI override 가능) |
| 1 release = 6 stamp | 동일 버전이 6개 하네스 디렉터리에 동시 박힘 (`minimax-code` 는 bootstrap 직접) |
| zip 이름 | `standard-ai-workflow-<harness>-<version>.zip` |
| GitHub release | harness 별 개별 asset |

### Distribution gates

`tools/check_packaging.py` (v0.5.8+ 자동). export 직후 5가지를 검증한다.

| # | gate | 실패 시 영향 |
|---|---|---|
| 1 | `manifest.json` schema 통과 | zip 폐기, release 차단 |
| 2 | `recommended_entrypoints` 실제 bundle 안에 존재 | 누락 파일 자동 추적 |
| 3 | `excluded_by_default` 항목이 bundle 안에 없음 | 컨텍스트 비대화 경고 |
| 4 | `PACKAGE_CONTENTS.md` / `APPLY_GUIDE.md` 가 export 루트에 동시 emit | 사용자 적용 절차 누락 경고 |
| 5 | zip 내부 top-level 가 `<harness>` 디렉터리 1개 | 압축 해제 경로 충돌 방지 |

`deferred_release_items` (예: `official_mcp_server_default_adoption`, `harness_mcp_activation`) 은 manifest 에 기록만 하고 bundle 에는 미포함. 이번 release 의 release_focus 는 `workflow_skill_onboarding` 이고, MCP 는 opt-in layer 로만 남는다 ([[entities/mcp-read-only-bundle]]).

## Export Pipeline

단일 진입점: `workflow-source/scripts/export_harness_package.py`. 7단계 직선 파이프라인.

| # | 단계 | 동작 | 산출물 |
|---|---|---|---|
| 1 | args parse | `--harness` (다중 가능, choices: `codex` / `opencode` / `gemini-cli` / `pi-dev` / `antigravity`), `--include-source-docs`, `--include-global-snippets` | `argparse.Namespace` |
| 2 | workflow-source → runtime 변환 | `DocTransformer` 로 `workflow-source/core/` 원본을 runtime `ai-workflow/core/` 형식으로 가공 | 변환된 core 문서 3종 |
| 3 | overlay 합성 | `bootstrap_workflow_kit.py` 의 harness builder 호출 → `AGENTS.md`, `opencode.json`, `.opencode/agents/*` 등 overlay 생성 | 하네스별 overlay 파일 |
| 4 | bundle 복사 | `bundle/` 아래 runtime + overlay 모음. `bootstrap_lib/harnesses/__init__.py` 의 `HARNESS_SPECS` 단일 source-of-truth 사용 | `dist/harnesses/<harness>/<version>/bundle/` |
| 5 | manifest 기록 | `package_name`, `package_version`, `release_focus`, `optimization_profile`, `recommended_entrypoints`, `excluded_by_default`, `deferred_release_items` | `manifest.json` |
| 6 | 문서 emit | `PACKAGE_CONTENTS.md`, `APPLY_GUIDE.md` 자동 생성 (한글 apply 안내 2종) | `dist/harnesses/<harness>/<version>/` 루트 |
| 7 | zip 압축 | `standard-ai-workflow-<harness>-<version>.zip` 생성 | GitHub release asset 후보 |

`export_harness_package.py` 의 `SUPPORTED_HARNESSES` 는 5종만 등록 (`minimax-code` 미포함). `minimax-code` 는 `bootstrap_workflow_kit.py` 의 `minimax-code` builder 를 직접 호출해 동기화 → 별도 dist 디렉터리 미생성. bootstrap 으로 직접 등록 (1줄 + 1줄).

### Post-export verification

export 직후 5개 sanity check (수동 또는 CI hook).

| check | 명령 | pass 기준 |
|---|---|---|
| 1 | `python3 tools/check_packaging.py --harness <h> --version <v>` | exit 0 |
| 2 | `unzip -l dist/harnesses/<h>/<v>/standard-ai-workflow-<h>-<v>.zip` | top-level 단일 `<h>/` 디렉터리 |
| 3 | `jq -r '.recommended_entrypoints[]' dist/harnesses/<h>/<v>/manifest.json` | 모든 entrypoint 가 zip 안에 존재 |
| 4 | `diff <(cat manifest.json | jq -S) <(cat dist/.../manifest.json | jq -S)` | 동일 (sha 검증) |
| 5 | `python3 -c "import json; json.load(open('dist/.../manifest.json'))"` | JSON schema 통과 |

## Related

- [[concepts/harness-distribution]] — 하네스별 overlay 원칙, bundle 구조, 7-step export 상세
- [[concepts/agent-topology]] — orchestrator + 4 worker 페르소나 정의
- [[entities/standard-ai-workflow]] — 본 프로젝트의 parent entity
- [[entities/harness-overlay-codex]] — Codex 오버레이 (`AGENTS.md` + `.codex/config.toml.example`)
- [[entities/harness-overlay-opencode]] — OpenCode 오버레이 (`AGENTS.md` + `opencode.json` + `.opencode/agents/*`)
- [[entities/harness-overlay-gemini-cli]] — Gemini CLI 오버레이 (`GEMINI.md`)
- [[entities/harness-overlay-antigravity]] — Antigravity 오버레이 (`ANTIGRAVITY.md`)
- [[entities/harness-overlay-minimax-code]] — MiniMax Code 오버레이 (`MiniMax.md` + `.MiniMax/agents/*`, dist 미생성)
- [[entities/harness-overlay-pi-dev]] — pi-dev 오버레이 (`AGENTS.md` + `SYSTEM.md`, persona 기반)
- [[entities/mcp-read-only-bundle]] — read-only MCP transport 공통 layer (5종 config)
- [[entities/ai-workflow-runtime]] — bundle 안 `ai-workflow/` runtime layer
- [[entities/workflow-kit]] — `workflow_kit.__version__` stamp 출처

## Source Path

- 본문: [workflow-source/core/workflow_harness_distribution.md](../../../workflow-source/core/workflow_harness_distribution.md)
- export 스크립트: [workflow-source/scripts/export_harness_package.py](../../../workflow-source/scripts/export_harness_package.py)
- bootstrap builder: [workflow-source/scripts/bootstrap_workflow_kit.py](../../../workflow-source/scripts/bootstrap_workflow_kit.py)
- harness 레지스트리: [workflow-source/scripts/bootstrap_lib/harnesses/__init__.py](../../../workflow-source/scripts/bootstrap_lib/harnesses/__init__.py)
- packaging smoke: [tools/check_packaging.py](../../../tools/check_packaging.py)
- sample bundle: [dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/codex/v0.6.3-beta/PACKAGE_CONTENTS.md) / [dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md](../../../dist/harnesses/opencode/v0.6.3-beta/PACKAGE_CONTENTS.md)
