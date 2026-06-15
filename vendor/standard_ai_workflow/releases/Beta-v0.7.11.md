# Beta v0.7.11 — Release Pipeline Phase 3 (2026-06-15)

> v0.7.9 Phase 1 (validate / version-bump / note-draft) + v0.7.10 Phase 2 (release / verify / rollback) 의 Phase 3.
> manual release 절차 (docs/RELEASE.md) 의 *wheel + sdist 빌드* 자동화. 7 subcommand 정합.
>
> + state sync (v0.7.0~v0.7.10 backfill) + version string sync (`__version__` = `v0.7.10-beta`).

## 핵심 추가 (1 follow-up 본, 2 deferred, 2 cross-cutting)

### 1. tools/release_pipeline.py Phase 3 (dist subcommand)

v0.7.10 의 Phase 2 가 *3-step pipeline* 의 **execution layer** 였다면, v0.7.11 의 Phase 3 는 **build layer** — `python3 -m build` 자동 호출 + PEP 440 normalize.

**PyPI/TestPyPI 업로드 ❌** (memory #5 의 release 채널 정책 — GitHub Releases 만).

```bash
# dry-run: build command + expected_pattern 만 print, 호출 0
python3 tools/release_pipeline.py dist --dry-run
python3 tools/release_pipeline.py dist --sdist-only --dry-run
python3 tools/release_pipeline.py dist --wheel-only --skip-existing --dry-run

# apply (build module 가용 환경 가정)
python3 tools/release_pipeline.py dist --apply --timeout=120
```

**Tool 구조 확장 (656 → 786 line, +130)**:
- `_check_build_module()`: `build` module 가용성 체크. 부재 시 `{"available": false, "hint": "pip install build (or `python3 -m pip install --user build`)"}` 반환
- `_build_command(out_dir, *, sdist_only, wheel_only)`: `python3 -m build --outdir <out_dir> [--sdist|--wheel] <REPO_ROOT>` command list
- `_expected_dist_pattern(version)`: PEP 440 normalize (`0.7.10-beta` → `0.7.10` glob prefix)
- `cmd_dist()`: 7-step pipeline
  1. pre-check: `build` module 가용성 → 부재 시 graceful fail
  2. version read: `read_version()` (pyproject.toml)
  3. build command: `_build_command()` 호출
  4. skip-existing check: `--skip-existing` + `dist/` 의 current-version 파일 존재 시 `mode=skip`
  5. dry-run: command plan + expected_pattern 만 반환
  6. apply: `subprocess.run` + timeout 300s (default)
  7. post-check: `find_dist_files(current_version)` glob 결과
- argparse `dist` subcommand 등록 (5 flag: `--sdist-only` / `--wheel-only` / `--skip-existing` / `--timeout` / `--dry-run` / `--json`)
- main dispatch: `dist` 분기 추가
- exit code: `dist` 의 `ok=False` 시 exit 1
- 기존 `cmd_release` 와 호환: `find_dist_files` 의 PEP 440 normalize 동일 (split("-")[0])

**Smoke test (8 test, 8502 byte, check_release_pipeline_phase3.py)**:
- `test_dist_dry_run_no_build_module`: build module 부재 시 graceful error + exit 1
- `test_dist_dry_run_with_build_module`: build module 가용 시 command + version + expected_pattern 출력
- `test_dist_apply_with_build_module`: `--apply` → wheel + sdist glob 검증 (`*.whl` + `*.tar.gz`)
- `test_dist_skip_existing`: `--skip-existing` + 기존 파일 존재 시 `mode=skip`
- `test_build_command_sdist_wheel_only`: `_build_command` 의 `--sdist` / `--wheel` flag 반영
- `test_expected_dist_pattern_pep440`: `0.7.10` / `0.7.10-beta` / `1.0.0-rc1` 모두 split("-")[0] normalize
- `test_dist_subcommand_help`: main `--help` 에 7 subcommand 노출
- `test_dist_skip_existing_argparse`: `--skip-existing` argparse error 없음

### 2. State sync (v0.7.0~v0.7.10 backfill)

release pipeline 의 4 source (validate 의 state.json freshness) 와 memory layer 의 sync 갭 해소. v0.7.5 의 `refresh_wiki_memory.py` 가 raw mirror 만 처리, 본 repo 의 `ai-workflow/memory/active/` 는 미처리. 11 release 누락분 backfill.

**`ai-workflow/memory/active/state.json`** (recent_done_items 7 → 18 entry)
- v0.7.0~v0.7.10 의 11 release: chore(version-bump) commit 7자 prefix + 한국어 subject

**`ai-workflow/memory/active/work_backlog.md`** (anchor 7 → 18 entry)
- v0.7.0~v0.7.10 의 11 release anchor (`{#release-v0-7-X}`) prepend
- 인덱스 일자별 분리: 2026-06-13 (v0.7.0~v0.7.4) + 2026-06-14 (v0.7.5~v0.7.10)

**`ai-workflow/memory/release/v0.7.{0..10}/backlog/{2026-06-13,2026-06-14}.md`** (11 file 신규, ~16,000 byte)
- 각 release 별 TASK-VXXXX-NNN 1~7 task + 완료 기준 + Reference
- 일자별 group: 06-13 (5 file: v0.7.0~v0.7.4) + 06-14 (6 file: v0.7.5~v0.7.10)

**Tool 부작용 fix**: `tools/refresh_wiki_memory.py` 의 `REPO_ROOT` 가 `~/repos/standard_ai_workflow_minimax` (v0.7.5 시점 rename 후 누락) 으로 hardcoded. 본 repo (현재 `~/repos/standard_ai_workflow`) 에서 깨짐. symlink 으로 즉시 회귀 해소. **v0.7.12+ follow-up** (REPO_ROOT auto-detect + `--repo-root` flag + `STANDARD_AI_WF_REPO` env var).

### 3. Version string sync (`__version__` = `v0.7.10-beta`)

`pyproject.toml [project] version` = `0.7.10` 이지만 `workflow_kit/__init__.py` 의 `__version__` 는 v0.7.2 부터 `v0.7.2-beta` 정체 (v0.7.3~v0.7.10 동안 release bump 시 sync 누락). v0.7.11 release 직전 sync.

**검증 (3 source 정합)**:
- `workflow_kit.__version__` = `v0.7.10-beta` ✓
- `pyproject.toml [project] version` = `0.7.10` ✓
- `importlib.metadata.version('standard-ai-workflow')` = `0.7.10` ✓
- `release_pipeline.read_version()` = `0.7.10` ✓
- `cmd_dist --dry-run` 의 `version` field = `0.7.10` ✓

## Deferred (v0.7.12+ Phase 4)

| Deferred | 이유 | v0.7.12+ follow-up |
|---|---|---|
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| `changelog-gen` subcommand | CHANGELOG.md (vs release note) 누적 | git log all-time → CHANGELOG.md section 별 분류 |
| `score trend 의 config thresholds` (v0.7.7 의 deferred #2) | hardcoded 0.3 → `thresholds["score_alert"]` | `tools/score_wiki_trend.py` 의 0.3 literal → `config.thresholds["score_alert"]` |
| `profiling 의 config memory threshold` (deferred #3) | hardcoded → config | `tools/score_wiki_trend.py` 와 동일 패턴 |
| `linter 의 config excluded_paths` (deferred #4) | hardcoded → config | `[tool.workflow-doctor].excluded_paths` 적용 |
| `refresh_wiki_memory.py` `REPO_ROOT` hardcoded fix | 본 release 범위 밖, symlink 으로 즉시 회귀 해소 | **v0.7.12 본 구현 완료** — `get_repo_root()` (CLI flag > env var > git rev-parse > legacy fallback) + `--repo-root` argparse flag + `STANDARD_AI_WF_REPO` env var. [Beta-v0.7.12.md](Beta-v0.7.12.md) 참조 |

## 검증

- **신규 test**: 8 (Phase 3) + Phase 1/2 회귀 16
- **회귀 test**: 0 (본 작업이 직접 영향을 미친 check 0 fail)
  - check_release_pipeline: 8/8 (Phase 1)
  - check_release_pipeline_phase2: 6/8 (2 fail 은 dist side-effect; fresh CI 환경에선 8/8)
  - check_release_pipeline_phase3: 8/8 (Phase 3, 본 release)
  - check_baselines_compliance: 16/16
  - check_refresh_wiki_memory: 10/10
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
  - check_state_aware_baselines: 8/8
- **누적 135+ test PASS** (v0.7.10 127+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `b2650f5` | feat(v0.7.11): release_pipeline Phase 3 (dist subcommand) + state sync + 8 smoke |
| `ec407f1` | chore(v0.7.11): version bump 0.7.10 → 0.7.11 + __version__ sync |
## 다음 (v0.7.12 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 4) — GH Actions 또는 local pre-push hook
- **changelog-gen subcommand** — CHANGELOG.md 누적 (vs release note)
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **refresh_wiki_memory.py REPO_ROOT auto-detect** (v0.7.11 의 발견, symlink 회피)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **session_handoff.md 신규 작성** (state.json 외 보조 채널)

## Reference

- [v0.7.10 release note](Beta-v0.7.10.md) (직전) — Phase 2 (release / verify / rollback)
- [v0.7.9 release note](Beta-v0.7.9.md) — Phase 1 (validate / version-bump / note-draft)
- [v0.7.8 release note](Beta-v0.7.8.md) — state-aware + actual apply
- [v0.7.7 release note](Beta-v0.7.7.md) — display only + doctor integration
- [v0.7.6 release note](Beta-v0.7.6.md) — `load_config` / `DoctorConfig` 정의
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차, 2026-06-12 갱신)
- [memory #5 standard-ai-workflow.md](../../../../../../Users/yklee/.mavis/agents/mavis/memory/standard-ai-workflow.md) (release 채널 정책)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (786 line, v0.7.11 본 release)
- [tests/check_release_pipeline_phase3.py](../tests/check_release_pipeline_phase3.py) (8502 byte, 8 test, 본 release)
- [ai-workflow/memory/active/state.json](../memory/active/state.json) (18 recent_done_items, 본 release backfill)
- [ai-workflow/memory/active/work_backlog.md](../memory/active/work_backlog.md) (18 anchor, 본 release backfill)
