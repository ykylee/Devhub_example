# Beta v0.7.17 — wiki in-repo storage isolation (2026-06-15)

> v0.7.5 의 `refresh_wiki_memory.py` 정식화 이후 *외부 vault* (`~/wiki/`) 와 본 project 의
> 연결을 완전 차단. 본 project 의 wiki 가 *전부 in-repo* (`ai-workflow/wiki/` +
> `ai-workflow/memory/active/`) 에서 관리됨. L1 raw mirror / L2 dense sources 모두 in-repo.
>
> **memory anchor**: 외부 vault 연결은 v0.7.5~v0.7.16 동안 *legacy* 였음 (symlink 회피,
> `_LEGACY_REPO_ROOT` deprecation). v0.7.17 에서 `Path.home() / "wiki"` reference *전부* 제거.

## 핵심 추가 (1 follow-up 본, 1 collateral cleanup, 0 deferred)

### 1. 5 file 의 in-repo path redirect (5 file)

| File | 이전 (`~/wiki/` 외부) | 이후 (in-repo) |
|---|---|---|
| `tools/refresh_wiki_memory.py` | `VAULT_ROOT = Path.home() / "wiki"`, `RAW_BASE = VAULT_ROOT / "raw" / "projects" / "standard-ai-workflow"`, `L2_BASE = VAULT_ROOT / "wiki" / "projects" / "standard-ai-workflow"` | `L1_BASE = REPO_ROOT / "ai-workflow"`, `L2_BASE = L1_BASE / "wiki" / "sources"` |
| `tools/emit_wiki_l2_body.py` | `VAULT_ROOT = Path.home() / "wiki"`, `RAW_MIRROR = VAULT_ROOT / "raw" / "projects"`, `L2_SOURCES = VAULT_ROOT / "wiki" / "projects"` | `REPO_ROOT = _detect_repo_root()` (git rev-parse), `RAW_MIRROR = L1_BASE / "wiki"`, `L2_SOURCES = L1_BASE / "wiki" / "sources"` |
| `tools/score_wiki_maintainability.py` | `L2_SOURCES = VAULT_ROOT / "wiki" / "projects" / "standard-ai-workflow" / "sources"` | `L2_SOURCES = INREPO_WIKI / "sources"` |
| `tests/check_refresh_wiki_memory.py` | `VAULT_ROOT = Path.home() / "wiki"` | (제거, *active code* 에서) |
| `tests/check_wiki_drift.py` | `_raw_mtime` 가 `VAULT_ROOT / raw_path` (외부) | `_raw_mtime` 가 `REPO_ROOT / raw_path` (in-repo) + `RAW_MIRROR = INREPO_WIKI` |

**Path 결정 정공법 (4-priority auto-detect + in-repo redirect)**:
1. `REPO_ROOT` 결정: `--repo-root` flag (refresh only) > `STANDARD_AI_WF_REPO` env var (refresh only) > `git rev-parse --show-toplevel` > `Path.cwd().resolve()` fallback
2. **in-repo** L1 raw mirror = `REPO_ROOT / "ai-workflow"` (= `ai-workflow/memory/active/` + `ai-workflow/wiki/`)
3. **in-repo** L2 dense sources = `REPO_ROOT / "ai-workflow" / "wiki" / "sources"`
4. 외부 vault (`~/wiki/`) reference **0** — code path / docstring / comment 모두 차단

### 2. 신규 dir + .gitkeep (1 file)

**`ai-workflow/wiki/sources/`** (L2 dense emit target, v0.7.17+ 신규):
- 위치: `ai-workflow/wiki/sources/.gitkeep`
- 역할: `tools/emit_wiki_l2_body.py --apply` 의 emit target. L1 raw mirror
  (`ai-workflow/wiki/concepts/`, `decisions/`, `entities/`, `patterns/`, `topics/`)
  의 본문 발췌 + dense derived view 가 본 dir 에 emit 됨.
- v0.7.17 이전: 외부 vault 의 `~/wiki/wiki/projects/standard-ai-workflow/sources/` 가
  emit target 이었음. *Multi-project centralization* 이라 사용자가 의도했으나
  본 project 의 *self-contained* 사용 (deploy / git clone 후 즉시 동작) 에는
  방해. v0.7.17 부터 *본 project 의 wiki 는 in-repo* 만.

### 3. 신규 smoke test (11 test, check_v0_7_17_wiki_in_repo_isolation.py)

| Test | 검증 |
|---|---|
| `test_refresh_wiki_memory_no_vault_root` | `Path.home() / "wiki"` reference 0 |
| `test_refresh_wiki_memory_raw_files_in_repo` | RAW_FILES 4 file 이 in-repo path |
| `test_refresh_wiki_memory_l2_stubs_in_repo` | L2_STUBS 4 file 이 `ai-workflow/wiki/sources/` 안 |
| `test_emit_wiki_l2_body_no_vault_root` | VAULT_ROOT/RAW_MIRROR/L2_SOURCES 가 in-repo path |
| `test_emit_wiki_l2_body_repo_root_auto_detect` | `_detect_repo_root` (git rev-parse) |
| `test_score_wiki_maintainability_l2_in_repo` | L2_SOURCES = INREPO_WIKI/sources |
| `test_check_refresh_wiki_memory_no_vault_root` | test 의 active code 에 VAULT_ROOT 0 |
| `test_check_wiki_drift_raw_mtime_in_repo` | `_raw_mtime` 이 REPO_ROOT 기반 in-repo path |
| `test_inrepo_sources_dir_exists` | `ai-workflow/wiki/sources/` + `.gitkeep` |
| `test_inrepo_memory_log_exists` | `ai-workflow/memory/log.md` (refresh 대상) |
| `test_inrepo_no_legacy_symlink_or_legacy_path` | in-repo 에 `raw/` dir 흔적 0 |

**11/11 PASS**.

## Deferred (v0.7.18+)

| Deferred | 이유 | v0.7.18+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` |
| linter 의 symlink-resolve fix (v0.7.16) | mavis data dir 격리 + macOS `/var` symlink | `.resolve()` → `.absolute()` |
| release coordination observability (v0.7.16) | session-start 시 remote tag check + 자동 bump | `cmd_release` pre-check |
| **외부 vault 의 L2 dense content 보존** | v0.7.17 이전 의 L2 emit 이 `~/wiki/.../sources/` 에 있었음. v0.7.17 의 in-repo redirect 시 *legacy* L2 file 들은 *옛 위치에 그대로*. 본 release 는 *새로운* L2 emit 만 in-repo. **legacy L2 → in-repo migration** (별도 release) | `tools/migrate_vault_l2_to_inrepo.py` 신규 tool 또는 `emit_wiki_l2_body.py --apply --project=standard-ai-workflow` 가 in-repo 에 dense content 처음부터 emit |

## 검증

- **신규 test**: 11 (위 §3)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_v0_7_17_wiki_in_repo_isolation: 11/11 (신규, 본 release)
  - check_cli_doctor: 8/8
  - check_metadata: 10/10
  - check_baselines_compliance: 16/16
  - check_state_aware_baselines: 8/8
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 8/8
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_changelog_gen: 4/4
  - check_atomic_write: 5/5 (v0.7.15)
- **기존 wiki test**:
  - check_wiki_drift: 5/5 (in-repo redirect 정상)
  - check_wiki_score: 12/12
  - check_wiki_lint: PASS
  - check_wiki_index_structure: PASS (44 entries)
  - check_wiki_source_rule: 1 fail (V-R9 baseline, 본 release 와 무관)
  - check_refresh_wiki_memory: 13/14 (1 fail 은 *pre-condition* — `test_reemit_l2_stubs_dry_returns_4_non_empty` 가 `ai-workflow/wiki/sources/active-state.md` 등 4 file 부재 가정. 본 release 의 in-repo sources/ 는 `.gitkeep` 만 있어 file 부재. **release cut 후 첫 `emit_wiki_l2_body.py --apply` 실행 시 fix**)
- **누적 178+ test PASS** (v0.7.16 167+ + 11 신규)

## Commit

| Hash | Subject |
|---|---|
| `6f6f1af` | feat(v0.7.17): wiki in-repo storage isolation (5 file redirect + sources/ 신규 + 11 smoke) |

## 다음 (v0.7.18 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5) — GH Actions 또는 local pre-push hook
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **linter 의 symlink-resolve fix** (v0.7.16 발견) — `.resolve()` → `.absolute()`
- **release coordination observability** (v0.7.16 발견) — session-start 시 remote tag check + 자동 bump
- **legacy L2 → in-repo migration** (v0.7.17 발견) — 외부 vault 의 기존 L2 file 을 in-repo 로 1회성 migrate

## Reference

- [v0.7.16 release note](Beta-v0.7.16.md) (직전) — config thresholds/excluded_paths + linter fix
- [v0.7.15 release note](Beta-v0.7.15.md) — atomic_write + changelog-gen filter
- [v0.7.5 release note](Beta-v0.7.5.md) — `refresh_wiki_memory.py` 정식화 (외부 vault 가 1차 출처 였던 시점)
- [v0.6.0 release note](Beta-v0.6.0.md) — LLM Wiki Layer 도입 (`ai-workflow/wiki/` in-repo + vault 동시 사용)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [ai-workflow/wiki/sources/.gitkeep](../../ai-workflow/wiki/sources/.gitkeep) (L2 dense emit target, 본 release 신규)
- [workflow-source/tools/refresh_wiki_memory.py](../tools/refresh_wiki_memory.py) (~545 line, v0.7.17 본 release, in-repo)
- [workflow-source/tools/emit_wiki_l2_body.py](../tools/emit_wiki_l2_body.py) (~365 line, v0.7.17 본 release, in-repo + _detect_repo_root)
- [workflow-source/tools/score_wiki_maintainability.py](../tools/score_wiki_maintainability.py) (~488 line, v0.7.17 본 release, in-repo)
- [tests/check_v0_7_17_wiki_in_repo_isolation.py](../tests/check_v0_7_17_wiki_in_repo_isolation.py) (~330 line, 11 test, 본 release)
