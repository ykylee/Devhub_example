# Beta v0.7.14 — version-bump auto-sync + changelog-gen subcommand (2026-06-15)

> v0.7.13 release 의 follow-up 2종 본 구현. (1) `cmd_version_bump` 의 `__init__.py` 자동 sync (manual sync 누락 패턴 해소), (2) `cmd_changelog_gen` subcommand 추가 (multi-release Keep-a-Changelog 형식).

## 핵심 추가 (2 follow-up 본, 3 deferred)

### 1. tools/release_pipeline.py cmd_version_bump auto-sync __version__

v0.7.11 → v0.7.12 → v0.7.13 release 의 3 release 동안 `workflow_kit/__init__.py` 의 `__version__` sync 가 manual 로 진행. release 시마다 `v0.7.2-beta` 가 정체되어 GH release 의 `__init__.py` 가 stale value 노출. 본 release 에서 자동화.

**구현**:
- `WORKFLOW_KIT_INIT = REPO_ROOT / "workflow_kit" / "__init__.py"` constant 추가
- `read_workflow_kit_version() -> str`: `__version__ = "..."` parse
- `write_workflow_kit_version(new_version, *, suffix="-beta") -> str`: `re.subn` 으로 `__version__ = "v\d+\.\d+\.\d+(?:-suffix)?"` 매치 + 교체, suffix 보존
- `cmd_version_bump` 가 `write_version(new)` + `write_workflow_kit_version(new, suffix="-beta")` 둘 다 호출 (default)
- result dict 에 `previous_pyproject` / `current_pyproject` / `previous_workflow_kit` / `current_workflow_kit` 4 key 추가 (CLI/text/JSON 모두)
- argparse `--no-init` flag: `__init__.py` sync skip (CI / override 시나리오)
- argparse `--dry-run` + `--json` flag 추가 (기존 v0.7.9 의 partial flag 추가)

**호출**:
```bash
# default (pyproject + __init__.py auto-sync)
python3 tools/release_pipeline.py version-bump --to=0.7.14 --apply

# dry-run (file 변경 없음, plan 만)
python3 tools/release_pipeline.py version-bump --to=0.7.14 --dry-run --json

# skip __init__.py sync (CI / override)
python3 tools/release_pipeline.py version-bump --to=0.7.14 --no-init --apply
```

**Smoke test (4 test 신규, check_release_pipeline_version_auto_sync.py)**:
- `test_version_bump_argparse`: `--no-init` / `--dry-run` / `--json` argparse error 없음
- `test_version_bump_dry_run_no_change`: dry-run mode 에서 pyproject + `__init__.py` 모두 변경 없음
- `test_version_bump_apply_sync_both`: apply mode 에서 pyproject + `__init__.py` 둘 다 갱신 + restore
- `test_version_bump_no_init_skips`: `--no-init` 시 pyproject 만 갱신 + `__init__.py` 보존

**기존 test 호환**:
- `check_release_pipeline.py` 의 3 test (test_version_bump_patch_dry_run / test_version_bump_to_explicit / test_version_bump_apply_and_restore) update: `result["next"]` / `result["previous"]` / `result["current"]` → `result["next_pyproject"]` / `result["previous_pyproject"]` / `result["current_pyproject"]` + `no_init=False` Args field 추가
- `check_release_pipeline_version_flag.py` 의 1 test (test_version_default_pyproject) update: hardcoded `v0.7.11-beta` → 동적 `f"v{current}-beta"` (release 시마다 갱신)

### 2. tools/release_pipeline.py cmd_changelog_gen subcommand (Phase 4)

v0.7.13 의 Deferred 표 항목 본 구현. multi-release git log → `CHANGELOG.md` (Keep-a-Changelog 1.1.0 형식). manual 작성의 한계 해소.

**구현**:
- `RELEASE_RE = re.compile(r"\(v(\d+\.\d+(?:\.\d+)?)\)")` (subject 의 `(vX.Y.Z)` 추출)
- `SECTION_PREFIXES` dict: `feat→Added`, `fix→Fixed`, `docs/refactor/perf/chore/test/build/ci→Changed` (Keep-a-Changelog 표준 6 종)
- `collect_commits_all_time() -> list[dict]`: `git log --all` 의 commit + `version` field (RELEASE_RE match 결과, "unreleased" if no match)
- `categorize_by_section(subject) -> str`: subject prefix 추출 → section mapping
- `draft_changelog(commits, unreleased_label="Unreleased") -> str`: Keep-a-Changelog 형식 본문 (latest first, version 별 section, section 별 30 commit)
- `cmd_changelog_gen(args) -> dict`: dry-run (preview) / apply (file write)
- argparse `--output` (default: `workflow-source/CHANGELOG.md`) + `--unreleased-label` (default: "Unreleased") + `--dry-run` / `--apply` / `--json`

**호출**:
```bash
# dry-run: 변경 plan preview (276 commit, 34 version, preview_first_500)
python3 tools/release_pipeline.py changelog-gen --dry-run --json

# apply: workflow-source/CHANGELOG.md 생성
python3 tools/release_pipeline.py changelog-gen --apply

# custom output
python3 tools/release_pipeline.py changelog-gen --output=/tmp/CHANGELOG.md --apply
```

**Smoke test (4 test 신규, check_release_pipeline_changelog_gen.py)**:
- `test_changelog_gen_argparse`: `--output` / `--unreleased-label` / `--dry-run` / `--json` argparse error 없음
- `test_changelog_gen_dry_run`: dry-run mode 에서 file 변경 없음 + result 정합
- `test_changelog_gen_apply`: apply mode 에서 CHANGELOG.md 작성 + Keep-a-Changelog 형식 검증 (# Changelog / ## [version] / ### Section / commit hash)
- `test_changelog_gen_section_categorization`: `categorize_by_section` 의 prefix → section mapping (feat/fix/docs/chore/refactor/perf/test)

**첫 적용** (본 release 시점):
- 276 commit, 34 version (v0.5.0 ~ v0.7.14 + Unreleased)
- 16,130 bytes CHANGELOG.md

## Deferred (v0.7.15+ Phase 5+)

| Deferred | 이유 | v0.7.15+ follow-up |
|---|---|---|
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| `score trend 의 config thresholds` (v0.7.7 의 deferred #2) | hardcoded 0.3 → `thresholds["score_alert"]` | `tools/score_wiki_trend.py` 의 0.3 literal → `config.thresholds["score_alert"]` |
| `profiling 의 config memory threshold` (deferred #3) | hardcoded → config | `tools/score_wiki_trend.py` 와 동일 패턴 |
| `linter 의 config excluded_paths` (deferred #4) | hardcoded → config | `[tool.workflow-doctor].excluded_paths` 적용 |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper 또는 subcommand 통합 |
| changelog-gen 의 `--from-tag` / `--to-tag` | 현재는 `--all` (전체). from/to tag 지정 시 filter | 다음 release 시: `--from-tag=v0.7.0 --to-tag=HEAD` |
| `cmd_release` 의 `--notes-template` | GH release notes 의 custom template | release note 를 `Beta-v0.7.14.md` 외 custom format 가능 |

## 검증

- **신규 test**: 8 (version-bump auto-sync 4 + changelog-gen 4) + Phase 1~3 회귀
- **회귀 test**: 0 (5 check / 35+ test PASS)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 8/8
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_version_flag: 3/3
  - check_release_pipeline_version_auto_sync: 4/4 (신규)
  - check_release_pipeline_changelog_gen: 4/4 (신규)
- **누적 150+ test PASS** (v0.7.13 142+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `23eb7fd` | feat(v0.7.14): cmd_version_bump auto-sync workflow_kit/__init__.py + cmd_changelog_gen subcommand + 8 smoke |
| `63ab483` | chore(v0.7.14): version bump 0.7.13 → 0.7.14 (auto-sync verified) + Beta-v0.7.14.md + CHANGELOG.md + state/work_backlog sync |
| `a01c7b4` | fix(v0.7.14): Beta-v0.7.14.md commit table TBD → 23eb7fd + 63ab483 |

## 다음 (v0.7.15 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 4 잔여) — GH Actions 또는 local pre-push hook
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **changelog-gen filter** (`--from-tag` / `--to-tag`) — multi-release 누적 시 range scan
- **`cmd_release` 의 `--notes-template`** — GH release notes 의 custom template


## Reference

- [v0.7.13 release note](Beta-v0.7.13.md) (직전) — cmd_release --version flag
- [v0.7.12 release note](Beta-v0.7.12.md) — REPO_ROOT auto-detect + v0.7.5~v0.7.10 backfill
- [v0.7.11 release note](Beta-v0.7.11.md) — Phase 3 (dist subcommand) + state sync + version sync
- [Keep-a-Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) (본 release 의 CHANGELOG.md 형식 reference)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (~990 line, v0.7.14 본 release, 8 subcommand)
- [CHANGELOG.md](../CHANGELOG.md) (v0.7.14 본 release 첫 생성, 276 commit, 34 version)
- [tests/check_release_pipeline_version_auto_sync.py](../tests/check_release_pipeline_version_auto_sync.py) (4 test, v0.7.14)
- [tests/check_release_pipeline_changelog_gen.py](../tests/check_release_pipeline_changelog_gen.py) (4 test, v0.7.14)
