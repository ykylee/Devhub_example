# Beta v0.7.13 — cmd_release --version flag (2026-06-15)

> v0.7.12 release 시 발견한 follow-up 본 구현. backfill 시 `pyproject.toml` 의 version 을 `sed` patch 했던 비효율 제거.
> 기존 `cmd_release` 의 `read_version()` 강제를 `--version=<X.Y.Z>` flag 로 override.

## 핵심 추가 (1 follow-up 본, 4 deferred)

### tools/release_pipeline.py cmd_release --version flag

v0.7.12 release 의 v0.7.5~v0.7.10 backfill 시 6 release 모두 staging 필요. `cmd_release` 가 `read_version()` 으로만 version 결정 → staging 시 `pyproject.toml` 의 version 을 `sed` patch + release + `sed` restore. 6 release × 4 sed call = 24 line 의 보일러플레이트.

**`--version=<X.Y.Z>` flag 추가**:
- argparse: `p_rel.add_argument("--version", default=None, help="version override (e.g. 0.7.5 for backfill). default: pyproject.toml [project] version")`
- cmd_release: `if args.version: version = args.version; results["version_source"] = "cli-flag"` else `version = read_version(); results["version_source"] = "pyproject.toml"`
- `version_source` field 가 result dict 에 포함 (CLI/text/JSON 모두) — staging caller 가 검증 가능

**호출**:
```bash
# default (pyproject.toml 의 version)
python3 tools/release_pipeline.py release --skip-validate --apply
# → version_source: pyproject.toml

# backfill / staging (explicit override)
python3 tools/release_pipeline.py release --version=0.7.5 --skip-validate --apply
# → version_source: cli-flag, tag: v0.7.5-beta
```

**Smoke test (3 test 신규, check_release_pipeline_version_flag.py)**:
- `test_version_argparse_recognized`: `--version=<X.Y.Z>` argparse error 없음 + `version_source=cli-flag`
- `test_version_override_pyproject`: `--version=0.7.5` 일 때 `tag=v0.7.5-beta` + `notes_file=Beta-v0.7.5.md` + assets 가 0.7.5 wheel/sdist (실제 staging copy + 검증)
- `test_version_default_pyproject`: `--version` 미지정 시 `version_source=pyproject.toml` (default) + `tag=v0.7.11-beta` (직전 HEAD)

**부수 변경**:
- `find_dist_files(version)` 변경 없음 (이미 keyword/positional arg 정상)
- `cmd_release` 의 `--verify-tag` flag (GH release create 의 --target main) 변경 없음
- `cmd_verify` / `cmd_rollback` 변경 없음 (별개 subcommand)

**Test file side effect**:
- `check_release_pipeline_version_flag.py` 의 test 2 (staging test) 가 v0.7.11 dist 를 pre-stage 에서 보장 (있으면 backup, 없으면 `/tmp/dist_v0_7_11/` 에서 copy) + finally 에서 restore. test 1/3 는 staging 불필요

## Deferred (v0.7.14+ Phase 5+)

| Deferred | 이유 | v0.7.14+ follow-up |
|---|---|---|
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| `changelog-gen` subcommand | CHANGELOG.md (vs release note) 누적 | git log all-time → CHANGELOG.md section 별 분류 |
| `score trend 의 config thresholds` (v0.7.7 의 deferred #2) | hardcoded 0.3 → `thresholds["score_alert"]` | `tools/score_wiki_trend.py` 의 0.3 literal → `config.thresholds["score_alert"]` |
| `profiling 의 config memory threshold` (deferred #3) | hardcoded → config | `tools/score_wiki_trend.py` 와 동일 패턴 |
| `linter 의 config excluded_paths` (deferred #4) | hardcoded → config | `[tool.workflow-doctor].excluded_paths` 적용 |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper 또는 subcommand 통합 |

## 검증

- **신규 test**: 3 (--version flag 3종) + Phase 1~3 회귀
- **회귀 test**: 0 (4 check 모두 PASS, 30+ test)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 8/8
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_version_flag: 3/3 (신규)
- **누적 142+ test PASS** (v0.7.12 139+ + 3 신규)

## Commit

| Hash | Subject |
|---|---|
| `922ebc0` | feat(v0.7.13): cmd_release --version flag (staging backfill, pyproject 일시 patch 불필요) |
| `TBD` | chore(v0.7.13): version bump 0.7.11 → 0.7.13 + __version__ sync + Beta-v0.7.13.md + state/work_backlog sync |

## 다음 (v0.7.14 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 4) — GH Actions 또는 local pre-push hook
- **changelog-gen subcommand** — CHANGELOG.md 누적 (vs release note)
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합

## Commit

| Hash | Subject |
|---|---|
| `922ebc0` | feat(v0.7.13): cmd_release --version flag (staging backfill, pyproject 일시 patch 불필요) |
| `628bf93` | chore(v0.7.13): version bump 0.7.11 → 0.7.13 + __version__ sync + Beta-v0.7.13.md |
| `afc685a` | chore(v0.7.13): state sync (v0.7.12 + v0.7.13 backfill) + 2 daily backlog |

## Reference

- [v0.7.12 release note](Beta-v0.7.12.md) (직전) — REPO_ROOT auto-detect + v0.7.5~v0.7.10 backfill
- [v0.7.11 release note](Beta-v0.7.11.md) — Phase 3 (dist subcommand) + state sync + version sync
- [v0.7.10 release note](Beta-v0.7.10.md) — Phase 2 (release / verify / rollback)
- [v0.7.9 release note](Beta-v0.7.9.md) — Phase 1 (validate / version-bump / note-draft)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (~810 line, v0.7.13 본 release)
- [tests/check_release_pipeline_version_flag.py](../tests/check_release_pipeline_version_flag.py) (5668 byte, 3 test, 본 release)
