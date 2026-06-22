---
type: pattern
status: active
last_ingested_from: ai-workflow/IMPORT_NOTES.md + vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py + vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py
related_pages: [concepts/devhub-overview, decisions/v0.7.17-import, topics/standard-ai-workflow-vendor]
created: 2026-06-15
updated: 2026-06-15
active_since: 2026-06-15
active_reason: "v0.7.17 wiki in-repo redirect (PR #600) + DevHub invariant 5/5 PASS"
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
---

# Wiki In-Repo Redirect Pattern (L1 pattern, in-repo)

## TL;DR

외부 vault (`~/wiki/`) 와 본 저장소 의 연결을 *완전 차단* 한 v0.7.17 의 5 file 의 in-repo path redirect 정공법. DevHub 의 wiki 운영 = `ai-workflow/wiki/` (L1) + `ai-workflow/wiki/sources/` (L2) + `ai-workflow/memory/log.md` (event log) + `ai-workflow/memory/active/` (active memory). 4-priority REPO_ROOT auto-detect (CLI flag > env var > git rev-parse > cwd fallback).

## 1. 핵심 결정 (5 file 의 in-repo path redirect)

| File | 이전 (`~/wiki/` 외부) | 이후 (in-repo) |
|---|---|---|
| `tools/refresh_wiki_memory.py` | `VAULT_ROOT = Path.home() / "wiki"`, `RAW_BASE = VAULT_ROOT / "raw" / "projects"`, `L2_BASE = VAULT_ROOT / "wiki" / "projects"` | `L1_BASE = REPO_ROOT / "ai-workflow"`, `L2_BASE = L1_BASE / "wiki" / "sources"` |
| `tools/emit_wiki_l2_body.py` | `VAULT_ROOT = Path.home() / "wiki"`, `RAW_MIRROR = VAULT_ROOT / "raw" / "projects"`, `L2_SOURCES = VAULT_ROOT / "wiki" / "projects"` | `REPO_ROOT = _detect_repo_root()` (git rev-parse), `RAW_MIRROR = L1_BASE / "wiki"`, `L2_SOURCES = L1_BASE / "wiki" / "sources"` |
| `tools/score_wiki_maintainability.py` | `L2_SOURCES = VAULT_ROOT / "wiki" / "projects"` | `L2_SOURCES = INREPO_WIKI / "sources"` |
| `tests/check_refresh_wiki_memory.py` | `VAULT_ROOT = Path.home() / "wiki"` | (제거, *active code* 에서) |
| `tests/check_wiki_drift.py` | `_raw_mtime` 가 `VAULT_ROOT / raw_path` (외부) | `_raw_mtime` 가 `REPO_ROOT / raw_path` (in-repo) + `RAW_MIRROR = INREPO_WIKI` |

## 2. REPO_ROOT 결정 정공법 (4-priority auto-detect)

1. **CLI flag** (`--repo-root`): refresh only
2. **env var** (`STANDARD_AI_WF_REPO`): refresh only
3. **git rev-parse** (`git rev-parse --show-toplevel`): 모든 도구 의 default
4. **fallback** (`Path.cwd().resolve()`): git 외부 환경 (CI container)

`_detect_repo_root()` (`emit_wiki_l2_body.py` L44-58):
- git rev-parse 성공 시 `Path(proc.stdout.strip()).resolve()`
- 실패 시 `Path.cwd().resolve()` (legacy fallback, stderr deprecation warning 1 회)

## 3. 신규 dir + .gitkeep (in-repo L2 emit target)

**`ai-workflow/wiki/sources/`** (L2 dense emit target, v0.7.17+ 신규):
- 위치: `ai-workflow/wiki/sources/.gitkeep`
- 역할: `tools/emit_wiki_l2_body.py --apply` 의 emit target. L1 raw mirror (`ai-workflow/wiki/concepts/`, `decisions/`, `entities/`, `patterns/`, `topics/`) 의 본문 발췌 + dense derived view 가 본 dir 에 emit 됨.
- Drift: L2 last_touched vs L1 mtime (7일 임계값, `check_wiki_drift.py` 의 smoke test)

## 4. 활성 memory 4 file (in-repo 1차 출처)

- `ai-workflow/memory/active/state.json` (v0.7.5+ 의 1차 출처)
- `ai-workflow/memory/active/work_backlog.md` (1차 출처)
- `ai-workflow/memory/active/session_handoff.md` (1차 출처)
- `ai-workflow/memory/active/PROJECT_PROFILE.md` (v0.7.5+ 의 1차 출처)

`refresh_wiki_memory.py` 의 RAW_FILES dict (4 file) — 모두 `L1_BASE / "memory" / ...` (in-repo).

## 5. 4 smoke (DevHub + vendor, 16/16 PASS)

### 5.1 vendor smoke (11/11)

`vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py` — 외부 vault reference 0 검증. 본 저장소 의 vendor 디렉터리 안 `vendor/standard_ai_workflow/ai-workflow/wiki/sources/.gitkeep` + `ai-workflow/memory/log.md` 의 fixture 0 회.

### 5.2 DevHub invariant (5/5)

`tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py` — 본 저장소 자체 의 5 file (`scripts/wiki-*.sh`, `docs/llm-wiki/*.md`, `ai-workflow/wiki/`, `ai-workflow/memory/log.md`, scripts 의 WIKI_SOURCES flat) 의 legacy reference 0 검증.

## 6. 갱신 SOP (vendor release)

- 새 vendor release 시 `git pull` + `cp -R ~/repos/standard_ai_workflow_minimax/workflow-source/. vendor/standard_ai_workflow/` (build/dist/pycache 제외, §1 의 filter)
- vendor smoke 11/11 + DevHub invariant 5/5 회귀 0 확인
- vendor 의 위임 출력/입력 스키마 + 위임 4 role (orchestrator/doc/code/validation) 의 *본 저장소 매핑* 갱신 (PR #601 의 minimax_code_workflow.md §5 의 cross-reference table)
- 메모리 anchor 갱신: MEMORY.md §12 ("Vendor side-by-side 격리 정공법") + §13 ("Vendor SSOT 동기화 정공법")

## 7. Cross-project 적용

- my_harness / server_manager 등 다른 project 의 *vendor import + wiki 운영 in-repo 전환* 시 동일 패턴.
- 5 file redirect + 4-priority REPO_ROOT + 4 smoke (2종) 의 5 box 가 *항상 동일*.
