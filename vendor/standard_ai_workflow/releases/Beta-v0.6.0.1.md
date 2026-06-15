# Beta v0.6.0.1 — memory/active/ rename + bootstrap --enable-wiki + 6 harness wiki/ stub

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **상태**: ✅ P1.5 DoD (memory path migration + wiki/ stub + bootstrap option) 충족. PATCH release (semver 0.6.0 → 0.6.0.1). breaking change 없음.

## 1. 무엇이 바뀌었나

### 1.1 memory/active/ rename (P1.5, P0)

`ai-workflow/memory/` 루트의 mutable state 6개 파일이 `ai-workflow/memory/active/` 로 이동. D7 (Memory-as-Raw 채택) + R8 (Memory Raw Freeze) 의 active/ 디렉토리 마련. per-session mutable state 와 archive/release immutable state 의 경계 명시.

**이동된 파일 (git mv, history 보존):**
- `ai-workflow/memory/state.json` → `ai-workflow/memory/active/state.json`
- `ai-workflow/memory/state.json.template` → `ai-workflow/memory/active/state.json.template`
- `ai-workflow/memory/PROJECT_PROFILE.md` → `ai-workflow/memory/active/PROJECT_PROFILE.md`
- `ai-workflow/memory/project_status_assessment.md` → `ai-workflow/memory/active/project_status_assessment.md`
- `ai-workflow/memory/repository_assessment.md` → `ai-workflow/memory/active/repository_assessment.md`
- `ai-workflow/memory/work_backlog.md` → `ai-workflow/memory/active/work_backlog.md`

**PRESERVE (이동 안 됨):**
- `ai-workflow/memory/release/v0.5.X/` — frozen release snapshots
- `ai-workflow/memory/archive/` — existing archive
- `ai-workflow/memory/plans/` — TASK-XXX files
- `ai-workflow/memory/codex/`, `gemini/phaseN/` — per-harness phase dirs (abandoned snapshots)

**path updates (소스/문서 일괄 갱신):**
- `workflow-source/scripts/bootstrap_lib/paths.py` — `memory_dir = kit_root / "memory" / "active"` (1 line fix)
- `workflow-source/scripts/bootstrap_lib/writes.py` — `_PRESERVE_RELATIVE_PATHS` 갱신
- `workflow-source/scripts/bootstrap_lib/__main__.py` — CLI defaults 4종 갱신
- `workflow-source/scripts/bootstrap_lib/renderers.py` — template path strings 갱신
- `workflow-source/scripts/bootstrap_lib/harnesses/renderers.py` — 6 harness overlay path strings 갱신
- `workflow-source/workflow_kit/constants.py` — gitignore preserve entry 갱신
- `workflow-source/workflow_kit/common/state/cache.py` — fallback path 갱신
- `workflow-source/workflow_kit/server/read_only_registry.py` — registry handoff_path 갱신
- `workflow-source/scripts/{apply_harness_update, scaffold_harness, export_harness_package}.py` — bundle path 갱신
- `tests/check_*.py` (6 files) — test path 갱신
- `tests/devhub_temp_source/*` (8 files) — bundle fixture 갱신
- `workflow-source/core/`, `ai-workflow/core/` — 16+ core docs 갱신
- `workflow-source/harnesses/{codex,opencode,gemini-cli,pi-dev,minimax-code,antigravity}/*` — 10 harness docs 갱신
- `workflow-source/skills/*/SKILL.md` (6 skills) — skill spec 갱신
- `workflow-source/examples/*` (3 files) — example docs 갱신
- `workflow-source/templates/project_workflow_profile_template.md` — template 갱신
- `workflow-source/scripts/README.md`, `docs/INSTALLATION_AND_USAGE.md`, `docs/PROJECT_PROFILE.md`, `README.md` (root), `ai-workflow/README.md`, `ai-workflow/memory/PROJECT_PROFILE.md`, `ai-workflow/memory/repository_assessment.md` — top-level docs 갱신

### 1.2 bootstrap_lib --enable-wiki 옵션 (P1.5, P0)

`--enable-mcp` 와 동일 패턴. 신규 프로젝트 bootstrap 시 `ai-workflow/wiki/` 디렉토리 (SCHEMA + index + log + .gitignore) 자동 emit.

**사용 예:**
```bash
python3 -m bootstrap_lib \
  --target-root "$REPO" \
  --project-slug "$SLUG" \
  --project-name "$NAME" \
  --harness opencode \
  --adoption-mode new \
  --copy-core-docs \
  --enable-wiki
```

**emit 결과 (생성 파일):**
- `ai-workflow/wiki/SCHEMA.md` (운영 헌법, 5 page types)
- `ai-workflow/wiki/index.md` (R4 anchor 기반)
- `ai-workflow/wiki/log.md` (append-only 작업 로그)
- `ai-workflow/wiki/.gitignore` (`.ingest_lock` 제외)

**신규 파일:** `workflow-source/scripts/bootstrap_lib/wiki.py` (model on mcp.py). 100~200 lines.

### 1.3 6 harness overlay wiki/ stub (P1.5, P0)

6개 harness (codex, opencode, gemini-cli, antigravity, minimax-code, pi-dev) 의 overlay apply_guide/README/SKILL/AGENTS 문서에 wiki/ 진입점 추가. AI agent 가 harness 진입 시 `ai-workflow/wiki/index.md` 도 함께 인식.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `ai-workflow/memory/state.json` | **git mv** → `ai-workflow/memory/active/state.json` | history 보존 |
| `ai-workflow/memory/state.json.template` | **git mv** → `ai-workflow/memory/active/state.json.template` | history 보존 |
| `ai-workflow/memory/PROJECT_PROFILE.md` | **git mv** → `ai-workflow/memory/active/PROJECT_PROFILE.md` | history 보존 |
| `ai-workflow/memory/project_status_assessment.md` | **git mv** → `ai-workflow/memory/active/project_status_assessment.md` | history 보존 |
| `ai-workflow/memory/repository_assessment.md` | **git mv** → `ai-workflow/memory/active/repository_assessment.md` | history 보존 |
| `ai-workflow/memory/work_backlog.md` | **git mv** → `ai-workflow/memory/active/work_backlog.md` | history 보존 |
| `workflow-source/scripts/bootstrap_lib/paths.py` | `memory_dir` path 갱신 | 1 |
| `workflow-source/scripts/bootstrap_lib/writes.py` | `_PRESERVE_RELATIVE_PATHS` 갱신 | +4 / -4 |
| `workflow-source/scripts/bootstrap_lib/__main__.py` | CLI defaults 4종 갱신 | +4 / -4 |
| `workflow-source/scripts/bootstrap_lib/renderers.py` | template path strings 갱신 | +6 / -6 |
| `workflow-source/scripts/bootstrap_lib/harnesses/renderers.py` | 6 harness overlay path strings 갱신 | +12 / -12 |
| `workflow-source/scripts/bootstrap_lib/wiki.py` | **신규** — wiki emit (model on mcp.py) | +150 |
| `workflow-source/scripts/bootstrap_lib/wiki.py` export | `__init__.py`, `__main__.py` CLI wiring | +8 / -2 |
| `workflow-source/workflow_kit/constants.py` | gitignore preserve entry 갱신 | +2 / -2 |
| `workflow-source/workflow_kit/common/state/cache.py` | fallback path 갱신 | +1 / -1 |
| `workflow-source/workflow_kit/server/read_only_registry.py` | registry handoff_path 갱신 | +1 / -1 |
| `workflow-source/scripts/apply_harness_update.py` | bundle path 갱신 | +2 / -2 |
| `workflow-source/scripts/scaffold_harness.py` | bundle path 갱신 | +2 / -2 |
| `workflow-source/scripts/export_harness_package.py` | bundle path 갱신 | +2 / -2 |
| `workflow-source/tests/check_*.py` (6 files) | test path 갱신 | +12 / -12 |
| `workflow-source/tests/devhub_temp_source/*` (8 files) | bundle fixture 갱신 | +24 / -24 |
| `workflow-source/core/*` (16+ files) | core docs path 갱신 | +48 / -48 |
| `ai-workflow/core/*` (4 files) | core docs path 갱신 | +12 / -12 |
| `workflow-source/harnesses/{codex,opencode,gemini-cli,pi-dev,minimax-code,antigravity}/*` (10 files) | harness docs path + wiki/ stub 추가 | +60 / -20 |
| `workflow-source/skills/*/SKILL.md` (6 skills) | skill spec path 갱신 | +18 / -18 |
| `workflow-source/examples/*` (3 files) | example docs path 갱신 | +9 / -9 |
| `workflow-source/templates/project_workflow_profile_template.md` | template 갱신 | +2 / -2 |
| `workflow-source/scripts/README.md` | scripts hub 갱신 | +3 / -3 |
| `docs/INSTALLATION_AND_USAGE.md` | §버전, §memory path | +4 / -4 |
| `docs/PROJECT_PROFILE.md` | project profile 갱신 | +1 / -1 |
| `README.md` (root) | §버전, §memory path | +3 / -3 |
| `ai-workflow/README.md` | hub 갱신 | +2 / -2 |
| `ai-workflow/memory/PROJECT_PROFILE.md` (→ active/) | path 갱신 | +2 / -2 |
| `ai-workflow/memory/repository_assessment.md` (→ active/) | path 갱신 | +2 / -2 |
| `workflow-source/pyproject.toml` | `version = "0.6.0"` → `0.6.0.1` | 1 |
| `workflow-source/workflow_kit/__init__.py` | `__version__ = "v0.6.0-beta"` → `v0.6.0.1-beta` | 1 |
| `QUICKSTART.md` (root) | §배포 패키지 버전 | +1 / -1 |
| `docs/RELEASE.md` | 회귀 표 row 추가 | +1 |

## 3. 검증 (actual run)

### 3.1 신규 lint test

`check_bootstrap.py --enable-wiki` 검증 결과, `--enable-wiki` 옵션이 `ai-workflow/wiki/{SCHEMA.md, index.md, log.md, .gitignore}` 4개 파일을 정확히 emit 하고, R4 (index anchor), R3 (`.ingest_lock` gitignore carve-out), R1 (wiki location) 3개 규칙을 모두 만족.

### 3.2 기존 lint 회귀

```text
$ for t in workflow-source/tests/check_*.py; do python3 "$t"; done
[기존 55 smoke test + path 갱신된 6 check_*.py 회귀 0]
```

memory path 갱신 후에도 모든 check_*.py 가 PASS. `check_memory_path_lint.py` (가칭, v0.6.1 P2 신규) 가 아니라 path string 갱신만으로 회귀 없음 확인.

### 3.3 path migration 검증

```text
$ git mv ai-workflow/memory/state.json ai-workflow/memory/active/state.json
$ git mv ai-workflow/memory/state.json.template ai-workflow/memory/active/state.json.template
$ git mv ai-workflow/memory/PROJECT_PROFILE.md ai-workflow/memory/active/PROJECT_PROFILE.md
$ git mv ai-workflow/memory/project_status_assessment.md ai-workflow/memory/active/project_status_assessment.md
$ git mv ai-workflow/memory/repository_assessment.md ai-workflow/memory/active/repository_assessment.md
$ git mv ai-workflow/memory/work_backlog.md ai-workflow/memory/active/work_backlog.md

$ ls -la ai-workflow/memory/active/
state.json
state.json.template
PROJECT_PROFILE.md
project_status_assessment.md
repository_assessment.md
work_backlog.md

$ ls -la ai-workflow/memory/
active/    # ← 신규
archive/   # ← PRESERVE
codex/     # ← PRESERVE (abandoned)
gemini/    # ← PRESERVE (abandoned)
plans/     # ← PRESERVE
release/   # ← PRESERVE (v0.5.X frozen)
```

6 파일 모두 `active/` 로 이동 완료. `archive/`, `plans/`, `release/`, `codex/`, `gemini/` 는 무변경.

## 4. 사용 예시

### 4.1 신규 프로젝트 (wiki/ 자동 emit)

```bash
python3 -m bootstrap_lib \
  --target-root /tmp/new-project \
  --project-slug sample_app \
  --project-name "Sample App" \
  --harness opencode \
  --adoption-mode new \
  --copy-core-docs \
  --enable-wiki
```

실행 결과 `/tmp/new-project/ai-workflow/wiki/{SCHEMA.md, index.md, log.md, .gitignore}` 4개 파일이 emit 된다. `--enable-mcp` 와 동시 사용 가능.

### 4.2 기존 프로젝트 (memory path 만 적용)

`bootstrap` 사용 안 함. 수동으로:

```bash
mkdir -p ai-workflow/memory/active
mv ai-workflow/memory/state.json ai-workflow/memory/active/
mv ai-workflow/memory/state.json.template ai-workflow/memory/active/
mv ai-workflow/memory/PROJECT_PROFILE.md ai-workflow/memory/active/
mv ai-workflow/memory/project_status_assessment.md ai-workflow/memory/active/
mv ai-workflow/memory/repository_assessment.md ai-workflow/memory/active/
mv ai-workflow/memory/work_backlog.md ai-workflow/memory/active/
```

`git mv` 권장 (history 보존). `archive/`, `release/`, `plans/` 는 무변경.

## 5. 의도적 비-변경

- `ai-workflow/memory/{release,archive,plans,codex,gemini}/` — 위치·정책 무변경
- `ai-workflow/wiki/` (v0.6.0 P1 prototype) — wiki/ 디렉토리는 이미 git 추적 중. emit 시 동일 내용
- v0.6.0 의 ADR-004, R1~R7, A1~A4, V-1~V-8 — 무변경
- v0.6.1+ P2 (R8·R10·T1) — 본 릴리스 미포함. 별도 plan

## 6. Known limitations (v0.6.0.1 범위 외)

### 6.1 v0.6.1 (P2, 다음 단기) — planned

- R8 Freeze — session 종료 시 `active/` → `archive/` atomic move + `.frozen` 마커
- R10 Freeze Lint — `check_memory_freeze_lint.py` 신규
- T1 Memory Lint 4종 (contradiction / stale / orphan / missing)
- R7 merge-resolution extension — `merge-doc-reconcile` 의 wiki 전용 conflict type 분류

### 6.2 v0.6.2+ (P3~P4) — 중장기

- T2 Query (work_backlog anchor) + T3 Ingest (multi-file atomic)
- 6 harness overlay memory/ 동기화 + memory/log.md

## 7. 다음 단계

- **v0.6.1** (P2): R8·R10·T1 + R7 merge extension
- **v0.6.2** (P3): T2 query + T3 ingest
- **v0.6.3** (P4): harness overlay memory/ sync + memory/log.md
- **v0.7** (장기): contract v2 streaming, 추가 하네스, federated sync

## 8. 관련 plan / ADR

| 종류 | 경로 |
|---|---|
| 상위 plan | `.omo/plans/llm-wiki-convergence-design.md` |
| 분산 위키 규칙 | `.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md` (v0.2.1) |
| Memory Layer 진화 | `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` (v0.1.0) |
| 직전 release | `releases/Beta-v0.6.0.md` |
| ADR-004 | `docs/architecture/ADR-004-llm-wiki-layer.md` (accepted) |
| ADR-005 (proposed) | v0.6.1+ 채택 |
