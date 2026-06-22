---
type: pattern
status: active
last_ingested_from: ai-workflow/wiki/patterns/in-repo-redirect.md
related_pages: [sources/in-repo-redirect]
created: 2026-06-15
updated: 2026-06-15
last_touched: 2026-06-22T06:03:34Z
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
mirror_dirty: |
---

# Wiki In-Repo Redirect Pattern (L2 dense, in-repo)

> L1 SSOT: `ai-workflow/wiki/patterns/in-repo-redirect.md`
> 본 L2 derived view 는 in-repo retrieval 용 압축 요약.

## TL;DR

외부 vault `~/wiki/` 와 본 저장소 연결 *완전 차단*. DevHub wiki 운영 = `ai-workflow/wiki/` (L1) + `ai-workflow/wiki/sources/` (L2) + `ai-workflow/memory/log.md` (event log) + `ai-workflow/memory/active/` (active memory). 4-priority REPO_ROOT auto-detect (CLI flag > env var > git rev-parse > cwd fallback).

## 1. 5 file 의 in-repo path redirect

| File | 이후 (in-repo) |
|---|---|
| `refresh_wiki_memory.py` | `L1_BASE = REPO_ROOT / "ai-workflow"`, `L2_BASE = L1_BASE / "wiki" / "sources"` |
| `emit_wiki_l2_body.py` | `REPO_ROOT = _detect_repo_root()` (git rev-parse), `RAW_MIRROR = L1_BASE / "wiki"`, `L2_SOURCES = L1_BASE / "wiki" / "sources"` |
| `score_wiki_maintainability.py` | `L2_SOURCES = INREPO_WIKI / "sources"` |
| `check_refresh_wiki_memory.py` | VAULT_ROOT active code 제거 |
| `check_wiki_drift.py` | `_raw_mtime` = `REPO_ROOT / raw_path` (in-repo) + `RAW_MIRROR = INREPO_WIKI` |

## 2. REPO_ROOT 4-priority

1. CLI flag `--repo-root` (refresh only)
2. env var `STANDARD_AI_WF_REPO` (refresh only)
3. `git rev-parse --show-toplevel` (모든 도구 default)
4. `Path.cwd().resolve()` (legacy fallback, stderr deprecation 1회)

## 3. 4 smoke (2종 16/16)

- **vendor smoke (11/11)**: `vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py` — 외부 vault reference 0
- **DevHub invariant (5/5)**: `tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py` — 본 저장소 의 5 file 의 legacy reference 0

## 4. Cross-project

- my_harness / server_manager 등 다른 project 의 *vendor import + wiki 운영 in-repo 전환* 시 동일 패턴.
- 5 file redirect + 4-priority REPO_ROOT + 4 smoke (2종) 의 5 box 가 *항상 동일*.
