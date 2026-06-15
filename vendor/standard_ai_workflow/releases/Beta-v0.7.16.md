# Beta v0.7.16 — [tool.workflow-doctor] config thresholds/excluded_paths 적용 (2026-06-15)

> v0.7.7~v0.7.14 의 Deferred 3종 (score_alert / memory_alert_mb / excluded_paths) 한꺼번에 해소.
> v0.7.6 의 `[tool.workflow-doctor]` schema 의 5 field 중 3 field 의 *실 consumer* 부재. 본 release 에서 모두 active.
> + `run_workflow_linter.py` 의 *누락된 dead code* (IndentationError) + dict slice bug fix (linter blocker 회피).
>
> **release 번호 변경 사유**: 본 release 의 작업은 본래 v0.7.15 로 cut 예정이었으나, 다른 session 에서 atomic_write + changelog-gen filter 가 v0.7.15 로 먼저 release 됨 (2026-06-15 06:30:11Z~06:32:26Z). 본 작업은 v0.7.16 으로 bump.

## 핵심 추가 (3 follow-up 본, 2 collateral fix, 0 deferred)

### 1. tools/score_wiki_trend.py — config.thresholds['score_alert'] 적용 (v0.7.7 deferred #2)

v0.7.7 의 `metadata.py` 가 default config 에 `thresholds = {"score_alert": 0.3, "memory_alert_mb": 100.0}` 박았지만, 실제 consumer (`score_wiki_trend.py`) 는 hardcoded `0.3` 그대로. 본 release 에서 0.3 literal 제거.

**구현**:
- module-level `_load_config_thresholds() -> dict`: `load_config(REPO_ROOT / "workflow-source")` 호출 → `config.thresholds` dict
- `SCORE_ALERT_DEFAULT = THRESHOLDS.get("score_alert", 0.3)`: consumer 에서 default 로 사용
- `compare_scores(baseline, current, alert_threshold=None)`: explicit None 이면 `SCORE_ALERT_DEFAULT` 사용
- argparse `--threshold`: default `SCORE_ALERT_DEFAULT` (default 0.3)
- `MEMORY_ALERT_MB_DEFAULT = THRESHOLDS.get("memory_alert_mb", 100.0)`: B-3 의 RSS probe 의 alert threshold

**Pyproject.toml [tool.workflow-doctor] 기본값 (v0.7.6 부터)**:
```toml
[tool.workflow-doctor]
thresholds = { score_alert = 0.3, memory_alert_mb = 100.0 }
```

운영 시 override:
```toml
[tool.workflow-doctor]
thresholds = { score_alert = 0.5, memory_alert_mb = 200.0 }  # 더 엄격한 alert
```

### 2. tools/score_wiki_trend.py — runtime RSS probe (v0.7.7 deferred #3)

`memory_alert_mb` 의 consumer 부재. v0.7.7 부터 default config 에 `100.0` 박혀있었지만 어디서도 안 씀. 본 release 에서 `record_current` 가 RSS 측정 + alert.

**구현**:
- `_probe_rss_mb() -> float | None`: `resource.getrusage(RUSAGE_SELF).ru_maxrss` 사용. macOS = bytes, Linux = KB 분기. Windows 는 부재 시 `None` 반환 (graceful).
- `record_current(commit)` 이 `_probe_rss_mb()` 호출 + record dict 에 `rss_mb` field 추가
- `rss_mb > MEMORY_ALERT_MB_DEFAULT` 시 stderr `⚠️ RSS X.XMB > memory_alert_mb Y.YMB` 경고 (CI 가독성)

**Cross-platform note**:
- macOS: `ru_maxrss` 가 bytes. `1024*1024` 로 MB.
- Linux: `ru_maxrss` 가 KB. `1024` 로 MB.
- Windows: `resource` module 부재. `None` 반환 (alert skip). 추후 psutil fallback 가능 (본 release scope 밖).

### 3. workflow_kit/common/linter.py — excluded_paths glob match (v0.7.7 deferred #4)

`excluded_paths` 가 v0.7.6 의 `DoctorConfig` field 인데, linter 의 `check_workflow_consistency` 가 broken link check 에 *적용 안 함*. 본 release 에서 `excluded_paths` 인자 추가 + glob match helper.

**구현**:
- `_is_excluded(path, excluded_patterns) -> bool`: glob match helper
  - `Path.match(pattern)` 1차 시도 (단일 segment)
  - `path.parents` 의 `match` 2차 시도 (상위 dir match)
  - `**` 패턴 시 regex 변환 후 match
- `check_workflow_consistency(..., *, excluded_paths=None)`: keyword-only 인자
- broken link check 시 `_is_excluded(link_path, excluded_paths)` 면 skip
- result `summary` 에 `excluded_paths` field 추가 (CI 가독성)
- `run_workflow_linter.py` 가 `load_config(project_root)` 호출 + `config.excluded_paths` 전달

**Default excluded_paths** (`_default_config()` 의 v0.7.6 default):
```python
excluded_paths = ["build/*", ".venv/*", ".venv-build/*", "__pycache__/*"]
```

본 release 의 pyproject.toml 에 1개 추가:
```toml
excluded_paths = ["build/*", ".venv/*", ".venv-build/*", "__pycache__/*", ".worktrees/*"]
```

### 4. skills/workflow-linter/scripts/run_workflow_linter.py — IndentationError fix (collateral)

`check_workflow_consistency` 가 호출 가능해진 후 `run_workflow_linter.py` 의 **누락된 dead code** 발견. line 152 의 `result = output_model.model_dump()` 후 line 153~163 의 `merge_into_result` 가 *잘못된 indent* (같은 level 의 indent 가 line 152 의 statement 의 body 처럼 보임). Python parser 가 `IndentationError: unexpected indent` → linter 가 *한 번도 실행 안 됨* (v0.7.7 부터).

**증상**:
```
File ".../run_workflow_linter.py", line 154
    result = merge_into_result(
IndentationError: unexpected indent
```

**Fix**: line 153~163 의 indent 를 line 152 의 같은 level 로 (4 spaces). `print(json.dumps(result, ...))` 도 `result = merge_into_result(...)` 의 *merged version* 으로 변경.

**Cross-cutting 발견**: `result["status"]` 가 dict 의 status field, `result.get("summary", "")[:200]` 가 `summary` 가 dict 면 `dict[:200]` = `KeyError("slice(None, 200, None)")`. 본 release 에서 `str(summary_value)[:200]` 로 dict → str 변환 후 slice.

### 5. Smoke test (9 test 신규, check_v0_7_15_config_thresholds.py)

| Test | 검증 |
|---|---|
| `test_score_alert_default_from_config` | `[tool.workflow-doctor] thresholds.score_alert=0.5` 일 때 load_config 가 반영 |
| `test_memory_alert_default_from_config` | default `memory_alert_mb=100.0` (없을 때) |
| `test_compare_scores_default_uses_config` | `compare_scores(None)` 가 config default 사용 |
| `test_compare_scores_explicit_override` | explicit `alert_threshold` 가 config default 보다 우선 |
| `test_record_current_adds_rss_mb_field` | record dict 에 `rss_mb` 추가 |
| `test_probe_rss_mb` | RSS 측정 (None 또는 float) |
| `test_linter_is_excluded` | `_is_excluded` glob match (build/*, .venv/*, tests/fixtures/*) |
| `test_linter_excluded_paths_skip_broken_link` | `excluded_paths` 적용 시 broken link skip |
| `test_run_workflow_linter_loads_config` | `run_workflow_linter` 의 load_config + linter 통합 (static check) |

**9/9 PASS**.

### 발견된 v0.7.17+ follow-up (test 환경 의존성, 본 release scope 밖)

- `check_workflow_linter.py` 의 `test_linter_pass` 가 mavis data dir 격리 환경 (`/Users/yklee/.mavis` → `/Users/yklee/.minimax` symlink) 에서 broken link false positive. **원인**: `tempfile.TemporaryDirectory()` 의 parent 가 symlink 면 linter 의 `.resolve()` 가 다른 real path 로 normalize → 4-levels-up README 가 안 보임. **fix (v0.7.17+)**: linter 의 broken link check 에서 `.resolve()` 대신 `.absolute()` 사용 (symlink 보존). 또는 test 가 `Path(tmp_dir).resolve(strict=False)` 로 명시.

### 6. Release 번호 변경 사유 (cross-cutting)

본 session 의 작업은 본래 v0.7.15 로 cut 예정이었으나, **다른 session 에서 atomic_write + changelog-gen --from-tag/--to-tag 가 v0.7.15 로 먼저 release 됨** (2026-06-15 06:30:11Z~06:32:26Z, commit `5cd1fe1` + `a369e7c`). 

**`gh release create v0.7.15-beta` 시점**:
- 다른 session 이 이미 remote 에 v0.7.15-beta tag + GH release 발행
- 본 session 의 `cmd_release --apply` 가 `a release with the same tag name already exists` fail
- **해결**: 본 session 의 작업을 v0.7.16 으로 bump + main 을 origin/main 으로 reset 후 cherry-pick

**Cross-cutting lesson**:
- **release coordination race** — 같은 v0.7.X tag 가 다른 session 에서 발행되면 자동 detection + 자동 bump 필요. v0.7.17+ follow-up.
- **`gh release create` 의 *pre-check* 가 필요**: release cut 시점에 remote 의 같은 tag 가 있는지 verify → 있으면 exit 1 + bump 권장. `cmd_release` 의 inner validation (cmd_validate 의 `--skip-packaging`) 와 별개.

## Deferred (v0.7.17+)

| Deferred | 이유 | v0.7.17+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` flag |
| test 환경의 symlink normalization (위 §5) | mavis data dir 격리 + macOS `/var` symlink | linter 의 `.resolve()` → `.absolute()` 또는 test 의 explicit normalize |
| release coordination observability (§6) | session-start 시 remote tag check + 자동 bump | `cmd_release` pre-check + auto-bump helper |

## 검증

- **신규 test**: 9 (B-1: 2 + B-2: 2 + B-3: 2 + linter integration: 3)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_v0_7_15_config_thresholds: 9/9 (신규, 본 release)
  - check_cli_doctor: 8/8
  - check_metadata: 10/10
  - check_baselines_compliance: 16/16
  - check_state_aware_baselines: 8/8
  - check_atomic_write: 5/5 (v0.7.15 다른 session, 본 release 회귀 0)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 8/8
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_changelog_gen: 4/4 (v0.7.14 + v0.7.15 의 changelog-gen 회귀 0)
- **본 release 의 linter 의 dead code + dict slice bug fix**:
  - `check_workflow_linter.py` 의 3 test 는 mavis data dir 격리 환경의 symlink issue 로 fail. v0.7.17+ follow-up.
- **누적 167+ test PASS** (v0.7.15 158+ + 9 신규)

## Commit

| Hash | Subject |
|---|---|
| `33f5243` | feat(v0.7.16): [tool.workflow-doctor] config thresholds/excluded_paths 적용 (B-1/B-2/B-3) + linter IndentationError fix + 9 smoke |

## 다음 (v0.7.17 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5) — GH Actions 또는 local pre-push hook
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — **v0.7.16 본 release 완료**
- **profiling 의 config memory threshold** (deferred #3) — **v0.7.16 본 release 완료**
- **linter 의 config excluded_paths** (deferred #4) — **v0.7.16 본 release 완료**
- **changelog-gen filter** (`--from-tag` / `--to-tag`) — **v0.7.15 다른 session 완료**
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **linter 의 symlink-resolve fix** (v0.7.16 발견) — `.resolve()` → `.absolute()` 또는 test 의 explicit normalize
- **release coordination observability** (v0.7.16 발견) — session-start 시 remote tag check + 자동 bump

## Reference

- [v0.7.15 release note](Beta-v0.7.15.md) (직전, 다른 session) — atomic_write + changelog-gen filter
- [v0.7.14 release note](Beta-v0.7.14.md) — version-bump auto-sync + changelog-gen (Phase 4)
- [v0.7.11 release note](Beta-v0.7.11.md) — release_pipeline Phase 3 (dist subcommand)
- [v0.7.7 release note](Beta-v0.7.7.md) — `DoctorConfig` 도입 (5 field schema, *본 release* 의 source)
- [v0.7.6 release note](Beta-v0.7.6.md) — `load_config` / `DoctorConfig` 정의
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [workflow_kit/common/metadata.py](../workflow_kit/common/metadata.py) (DoctorConfig / load_config, 162 line)
- [workflow_kit/common/linter.py](../workflow_kit/common/linter.py) (~200 line, v0.7.16 본 release, excluded_paths 인자)
- [tools/score_wiki_trend.py](../tools/score_wiki_trend.py) (~390 line, v0.7.16 본 release, SCORE_ALERT_DEFAULT + rss_mb)
- [skills/workflow-linter/scripts/run_workflow_linter.py](../skills/workflow-linter/scripts/run_workflow_linter.py) (~180 line, v0.7.16 본 release, load_config + IndentationError fix)
- [tests/check_v0_7_15_config_thresholds.py](../tests/check_v0_7_15_config_thresholds.py) (~290 line, 9 test, 본 release)
- [pyproject.toml](../pyproject.toml) `[tool.workflow-doctor]` (v0.7.6+ 의 5 field, 본 release 의 consumer 3종 활성화)
