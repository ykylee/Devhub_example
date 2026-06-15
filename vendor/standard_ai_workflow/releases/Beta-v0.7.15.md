# Beta v0.7.15 — atomic_write helper + changelog-gen filter (2026-06-15)

> v0.7.14 release note 의 Deferred 표 항목 중 2종 본 구현. (1) SSOT JSON/text 의 interrupted-write 문제 해소 (POSIX `os.replace` 기반 atomic write), (2) `changelog-gen` 의 `--from-tag/--to-tag` range filter (multi-release 누적 시 range scan).
> + 세션 중 발견한 state.json 의 manual write race condition 도 본 helper 로 해소.

## 핵심 추가 (2 follow-up 본, 3 deferred)

### 1. workflow_kit.common.atomic_write (신규 helper module)

v0.7.11~v0.7.14 release 동안 본 repo 의 `ai-workflow/memory/active/state.json` 가 manual write 중 *0 byte* 상태로 깨지는 현상 발견. AI agent 가 write 도중 crash / tool kill -9 시 발생. 본 helper 도입으로 *interrupted-write* 문제 해소.

**POSIX `os.replace` 의 atomic guarantee**:
```
1. tempfile.mkstemp(dir=path.parent) — unique tmp file in same dir
2. f.write(...) + f.flush() + os.fsync(f.fileno()) — disk 에 영구 기록
3. os.replace(tmp, target) — atomic rename(2) (POSIX)
```

중간에 crash 나도 target file 은 *이전 version* 또는 *없음* 상태 보장. 둘 다 valid.

**구현**:
- `atomic_write_json(path, data, *, indent=2, ensure_ascii=False)` — JSON file 의 atomic write
- `atomic_write_text(path, body, *, encoding="utf-8")` — text file 의 atomic write
- module-level `WORKFLOW_KIT_INIT` import + workflow_kit.common.__init__ export
- error handling: 모든 예외 시 tmp file cleanup (`os.unlink`)
- best-effort fsync (일부 filesystem 에서 fail 가능, non-fatal)

**Smoke test (3 test 신규, check_atomic_write.py)**:
- `test_atomic_write_json_creates_file`: 부재 dir + 부재 file → file 생성 + 내용 정합
- `test_atomic_write_json_replaces_existing`: 기존 file → 새 내용으로 replace + old content 부재 검증
- `test_atomic_write_json_atomicity`: source-level POSIX pattern 검증 (`tempfile` + `os.replace` + `mkstemp` + `os.fsync` import/사용)

**refresh_wiki_memory.py 5 call site 변경**:
- `try/except ImportError` fallback (atomic_write_json/text = None 시 `p.write_text` fallback — standalone test 환경 호환)
- `state.json` / `work_backlog.md` / `wiki/log.md` / `memory/log.md` / L2 stubs 의 5 file write 가 atomic 으로 변경
- dead code fix: `write → mutate` 의 mutate-after-write 순서 정상화 (`last_freeze` / `last_ingest` mutate 가 `atomic_write` 이전에 실행)
- 14/14 smoke test PASS (이전 14/14 + 무변동, 0 regression)

**release_pipeline.py 활용** (Phase 2 의 changelog-gen filter 본 구현 시):
- `cmd_changelog_gen` 의 `output_path.write_text(body)` → `atomic_write_text(output_path, body)` (fallback 포함)
- `try/except ImportError` fallback (release_pipeline.py 가 standalone script 로 실행 가능)

### 2. tools/release_pipeline.py changelog-gen filter (--from-tag / --to-tag)

v0.7.14 release note 의 Deferred 표 항목 본 구현. multi-release 누적 시 range scan 가능.

**구현**:
- `collect_commits_in_range(from_ref: str | None, to_ref: str = "HEAD")` 신규
  - `from_ref=None` → `collect_commits_all_time()` (deprecated wrapper)
  - `from_ref=<tag>` → `git log <from>..<to> --pretty=format:...`
  - invalid ref 시 `git log` exit != 0 → empty list + caller 의 error handling
- `_parse_git_log(pretty_output)` helper — RELEASE_RE parse + commit dict 변환
- argparse `--from-tag` (default: None) + `--to-tag` (default: HEAD) flag 추가
- result dict `from_tag` / `to_tag` field 추가
- `cmd_changelog_gen` 의 error case: invalid tag → `mode=error` + tag name 포함 error message
- `atomic_write_text` fallback to `output_path.write_text` (atomic_write import 가능 시 atomic write)

**호출**:
```bash
# 전체 history (default)
python3 tools/release_pipeline.py changelog-gen --dry-run --json

# range scan
python3 tools/release_pipeline.py changelog-gen --from-tag=v0.7.0-beta --to-tag=v0.7.10-beta --dry-run --json
# → 49 commit, 12 version

# smaller range
python3 tools/release_pipeline.py changelog-gen --from-tag=v0.7.5-beta --to-tag=v0.7.8-beta --dry-run --json
# → 14 commit, 5 version

# invalid tag (graceful fail)
python3 tools/release_pipeline.py changelog-gen --from-tag=v9.9.9-beta --dry-run --json
# → {"error": "no commits in range v9.9.9-beta..HEAD (from_tag 또는 to_tag invalid 할 수 있음)"}
```

**Smoke test (2 test 신규 + 1 update, check_release_pipeline_changelog_gen.py)**:
- `test_changelog_gen_range_filter`: full history vs v0.7.0~v0.7.10 vs v0.7.5~v0.7.8 의 commit count hierarchy 검증
- `test_changelog_gen_out_of_range_graceful`: invalid `--from-tag=v9.9.9-beta` 시 `mode=error` + error message 검증
- `test_changelog_gen_argparse` update: `--from-tag` / `--to-tag` 의 argparse error 없음 검증
- subprocess `PYTHONPATH=SOURCE_ROOT` env 추가 (workflow_kit import 위해)
- sys.path.insert 추가 (importlib _import_tool 의 module import 위해)

**기존 test dynamic update** (release 마다 깨지지 않도록):
- `check_release_pipeline_version_auto_sync.py`: hardcoded `0.7.13` / `v0.7.13-beta` → `_bump_patch(pre_py)` 동적 read
- `check_release_pipeline_version_flag.py`: hardcoded `v0.7.11-beta` → `f"v{current}-beta"` 동적
- `check_release_pipeline.py`: 3 test 의 result field `next` / `previous` / `current` → `next_pyproject` / `previous_pyproject` / `current_pyproject` + `no_init=False` Args
- `check_refresh_wiki_memory.py`: sys.path.insert 추가 (workflow_kit.common.atomic_write import 위해)

## Deferred (v0.7.16+ Phase 5+)

| Deferred | 이유 | v0.7.16+ follow-up |
|---|---|---|
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 (사용자 제외 결정) |
| `score trend 의 config thresholds` (v0.7.7 의 deferred #2) | hardcoded 0.3 → `thresholds["score_alert"]` | `tools/score_wiki_trend.py` 의 0.3 literal → `config.thresholds["score_alert"]` |

| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper 또는 subcommand 통합 |
| `cmd_release` 의 `--notes-template` | GH release notes 의 custom template | release note 를 `Beta-v0.7.15.md` 외 custom format 가능 |
| tag 부재 tag 의 `--from-tag` (v0.7.5~v0.7.10 의 6 tag 가 backfill 되어 해결) | local `git rev-parse` fail | 본 release 의 backfill 로 해결. v0.7.16+ 부터 정상 동작 |

## 검증

- **신규 test**: 8 (atomic_write 3 + changelog-gen 2 + refresh_wiki_memory unchanged 14 + version-auto-sync 4 + version-flag 3 + release_pipeline 8) — *effective 신규* 5
  - check_atomic_write: 3 신규
  - check_release_pipeline_changelog_gen: 2 신규 (range_filter / out_of_range_graceful) + 1 update
  - check_release_pipeline.py: 3 update
  - check_release_pipeline_version_flag.py: 1 update
  - check_release_pipeline_version_auto_sync.py: dynamic pre/current
  - check_refresh_wiki_memory.py: sys.path update
- **회귀 test**: 0 (8 check / 58 test PASS)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 8/8
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_version_flag: 3/3
  - check_release_pipeline_version_auto_sync: 4/4
  - check_release_pipeline_changelog_gen: 6/6
  - check_atomic_write: 3/3
  - check_refresh_wiki_memory: 14/14
- **누적 158+ test PASS** (v0.7.14 150+ + 5+ 신규 / 3+ update)

## Commit

| Hash | Subject |
|---|---|
| `5cd1fe1` | feat(v0.7.15): atomic_write helper + changelog-gen --from-tag/--to-tag filter + 5 smoke |
| `a369e7c` | chore(v0.7.15): version bump 0.7.14 → 0.7.15 (auto-sync verified) + Beta-v0.7.15.md |
| `3049651` | chore(v0.7.15): state sync (atomic_write 적용) + 1 daily backlog |
| `3dfb5a1` | fix(v0.7.15): Beta-v0.7.15.md Commit section + 3 commit hash |

## 다음 (v0.7.16 / v0.8.0 후보)

- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **`cmd_release` 의 `--notes-template`** — GH release notes 의 custom template

## Reference

- [v0.7.14 release note](Beta-v0.7.14.md) (직전) — version-bump auto-sync + changelog-gen (1st cut) + 8 smoke
- [v0.7.13 release note](Beta-v0.7.13.md) — cmd_release --version flag
- [v0.7.12 release note](Beta-v0.7.12.md) — REPO_ROOT auto-detect + v0.7.5~v0.7.10 backfill
- [POSIX rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html) (atomic write 의 원리)
- [Keep-a-Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) (CHANGELOG.md 형식 reference)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (~990 line, 8 subcommand, v0.7.15 본 release)
- [workflow_kit/common/atomic_write.py](../workflow_kit/common/atomic_write.py) (v0.7.15 신규 helper, 3360 bytes)
- [tests/check_atomic_write.py](../tests/check_atomic_write.py) (3 test, v0.7.15)
- [tests/check_release_pipeline_changelog_gen.py](../tests/check_release_pipeline_changelog_gen.py) (6 test, v0.7.14 + v0.7.15)
- [tests/check_refresh_wiki_memory.py](../tests/check_refresh_wiki_memory.py) (14 test, atomic_write 적용)
- [workflow_kit/__version__](../workflow_kit/__init__.py) = `v0.7.15-beta`
- [pyproject.toml](../pyproject.toml) version = `0.7.15`
- GH release URL: `https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.15-beta`
