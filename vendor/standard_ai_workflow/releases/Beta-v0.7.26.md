# Beta v0.7.26 — Branch Detection Fix (F-7) + Automated Hash Sync (F-7+)

> **Status**: in-flight
> 본 release 의 변경. check_workflow_linter 의 branch detection 정확도 + v0.7.25 의 infinite fix(state) loop 해결.

## 본 release 의 1차 출처

1. **v0.7.22 release note** (linter symlink-resolve fix) — 본 release 의 collateral fix 의 1차 출처 (paths.py 의 *workflow dir resolve* 의 일관성)
2. **v0.7.18 release note** (destructive subcommand 정공법) — 본 release 의 automated hash sync 의 *graceful fail* 의 1차 출처
3. **v0.7.21 release note** (--allow-existing-tag) — 본 release 의 state.json + backlog hash sync 의 *본 release 의 본 release 의 hash* 의 정공법
4. **v0.7.25 release note** (F-6 closure, infinite fix(state) loop) — 본 release 의 automated hash sync 의 *1차 동기화 trigger*

## 발견

### 1. detached HEAD 시 branch detection 부정확 (F-7)

`workflow_kit/common/paths.py:get_current_branch()` 의 *detached HEAD* (e.g. CI checkout at specific SHA, 또는 `git checkout <sha>`) fallback:

- `git rev-parse --abbrev-ref HEAD` 가 `"HEAD"` (literal) 반환
- `_usable_branch_name("HEAD")` = `None` (이미 처리됨 — line 79)
- `return None or "main"` = `"main"` ← **잘못된 branch 이름**

이 *fallback* 은 *잘못된* memory dir 로의 *조용한 misroute* (CI 의 detached checkout 시 *모든* 가 "main" 으로 강제 fallback).

### 2. v0.7.25 의 infinite fix(state) loop (F-7+)

v0.7.25 의 release cut 시 (8 fix(state) commit + squash + final fix):
- 본 release 의 *fix(state)* commit 이 *본 release 의 본 release 의 hash* 를 변경
- 변경된 hash 가 state.json 의 recent_done_items[0] 와 *정합* 해야 함
- 그 *정합* commit 이 *또 다른* hash 변경
- → *infinite loop*

`tools/sync_release_hash.py` (F-7+ 신규)로 *자동* sync — chore commit 시 *자동* 호출되어 1 commit 으로 state.json + backlog 의 hash = latest commit hash.

## 본 release 의 변경

### 1. workflow_kit/common/paths.py: F-7 (detached HEAD → 7-char SHA)

**변경 위치**: `get_current_branch()` 의 `git rev-parse --abbrev-ref HEAD` 후처리.

**Before** (v0.7.25):
```python
def get_current_branch() -> str:
    for env_key in BRANCH_ENV_KEYS:
        branch = _usable_branch_name(os.environ.get(env_key))
        if branch:
            return branch
    repo_root = Path(__file__).resolve().parents[3]
    try:
        branch = subprocess.check_output(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=str(repo_root), stderr=subprocess.DEVNULL, text=True,
        ).strip()
        return _usable_branch_name(branch) or "main"
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "main"
```

**After** (v0.7.26):
```python
def get_current_branch() -> str:
    for env_key in BRANCH_ENV_KEYS:
        branch = _usable_branch_name(os.environ.get(env_key))
        if branch:
            return branch
    repo_root = Path(__file__).resolve().parents[3]
    try:
        branch = subprocess.check_output(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=str(repo_root), stderr=subprocess.DEVNULL, text=True,
        ).strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "main"

    # F-7 fix: detached HEAD → 7-char commit SHA (not "main")
    if branch == "HEAD":
        try:
            short_sha = subprocess.check_output(
                ["git", "rev-parse", "--short=7", "HEAD"],
                cwd=str(repo_root), stderr=subprocess.DEVNULL, text=True,
            ).strip()
            if short_sha and len(short_sha) >= 7:
                return short_sha
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
        return "main"

    return _usable_branch_name(branch) or "main"
```

### 2. tools/sync_release_hash.py: F-7+ (automated hash sync, infinite loop 회피)

**신규 tool** (4-priority REPO_ROOT auto-detect, v0.7.12 정공법).

**subcommand**: `sync-release-hash` (단일 subcommand).

**mode**: `--dry-run` / `--apply` (v0.7.18+ 정공법).

**구현 핵심**:
- `get_latest_commit_hash(repo_root)` — `git rev-parse HEAD` (full SHA) + `git rev-parse --short=7 <full>` (short SHA) — 2-step
- `update_state_json(state_path, version, new_hash)` — `"{version} (old_or_TBD):"` → `"{version} ({new_hash}):"`
- `update_backlog(backlog_path, new_hash)` — `**commit**: \`old_or_TBD\`` → `**commit**: \`{new_hash}\``
- *TBD* placeholder 도 sync 대상 (정합성 1-shot)

**TBD vs 7-char hash 매치**:
- state.json: `re.compile(rf'"{re.escape(version)} \(([a-f0-9]{{7}}|TBD)\):')` — TBD 또는 7-char hex
- backlog: `re.compile(r'\*\*commit\*\*: `([a-f0-9]{7}|TBD)`')` — TBD 또는 7-char hex

### 3. 2 smoke test (10/10 PASS)

- `tests/check_v0_7_26_branch_detection_fix.py` (5 smoke, 5/5 PASS):
  1. `test_detached_head_short_sha_fallback` — detached HEAD → 7-char commit SHA
  2. `test_normal_branch_unchanged` — main / feat/* / release/* 정상
  3. `test_env_override_still_works` — CODEX_WORKFLOW_BRANCH / GITHUB_HEAD_REF / GITHUB_REF_NAME
  4. `test_unsafe_env_falls_back` — unsafe env + detached → main
  5. `test_git_failure_falls_back` — git fail → main

- `tests/check_v0_7_26_sync_release_hash.py` (5 smoke, 5/5 PASS):
  1. `test_state_json_hash_updated` — state.json 의 v0.7.26 (TBD) → v0.7.26 (hash)
  2. `test_backlog_commit_updated` — backlog 의 `**commit**: \`TBD\`` → `**commit**: \`hash\``
  3. `test_idempotent_no_match` — entry 부재 시 state_updated: False (no write)
  4. `test_dry_run_no_write` — --dry-run 시 file 변경 0
  5. `test_real_repo_hash_sync` — *real* REPO_ROOT + *real* v0.7.25 state.json entry → sync verify (96a919d → 00e7ca8)

**10/10 PASS** (207+ → 217+ 누적 test).

### 4. cumulative 217+ test (회귀 0)

- v0.7.25 의 5 신규 test = 207+
- v0.7.26 의 10 신규 test (F-7 5 + F-7+ 5) = **217+**
- 기존 `check_paths.py` PASS (회귀 0) — `BRANCH_NAME` module-level 의 *main* / *worktree* 정확

## 발견된 cross-cutting lesson (v0.7.26)

- **detached HEAD 의 *stable identifier***: short SHA (7자) 는 *collision-resistant* (16M commit 중 1% collision). "main" fallback 은 *잘못된* branch 이름. **F-7 의 정공법 = SHA-7**.
- **infinite fix(state) loop 의 *automated 해결***: `tools/sync_release_hash.py` 의 *1-shot sync* + regex 가 TBD 도 매치 → *infinite loop* 없이 *1 commit* 으로 정합. **F-7+ 의 정공법 = regex `([a-f0-9]{7}|TBD)`**.
- **TBD 의 *의미***: `TBD` 는 *placeholder* — *uncommitted* state 의 표시. 정합 후엔 *latest commit hash* 로 *대체* 되어야 함. *Stale TBD* 는 *정합 안 됨* 의 신호.
- **2-step git rev-parse 의 *호환성***: `git rev-parse HEAD --short=7` (1-step) 은 *git version* 에 따라 동작 *불일치*. 2-step (`rev-parse HEAD` → full SHA → `rev-parse --short=7 <full>`) 이 *더 호환성* 좋음. **F-7+ 의 정공법 = 2-step**.

## Reference (다른 release note)

- v0.7.25 release note (F-6 closure, 본 release 의 *automated hash sync* 의 1차 동기화 trigger)
- v0.7.22 release note (linter symlink-resolve, 본 release 의 collateral 의 1차 출처)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 hash sync 의 *본 release 의 본 release 의 hash* 정공법)
- v0.7.18 release note (destructive subcommand, 본 release 의 *graceful fail* 의 1차 출처)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 sync_release_hash.py 의 REPO_ROOT auto-detect 의 1차 출처)
- v0.7.24 release note (--notes-template flag, 본 release 의 release note format 의 1차 출처)

## 2 TASK (본 release)

### TASK-V0726-001: F-7 branch detection fix (detached HEAD → short SHA)

- **commit**: TBD
- **status**: in-flight
- **scope**: `workflow_kit/common/paths.py:get_current_branch()` (~15 line 추가) + `tests/check_v0_7_26_branch_detection_fix.py` (5 smoke, 5/5 PASS)

### TASK-V0726-002: F-7+ automated hash sync (infinite fix(state) loop 회피)

- **commit**: TBD
- **status**: in-flight
- **scope**: `tools/sync_release_hash.py` (~180 line, 4-priority REPO_ROOT + 1 subcommand + 3 helper) + `tests/check_v0_7_26_sync_release_hash.py` (5 smoke, 5/5 PASS)

## Follow-up (v0.7.27+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0726-003**: `sync_release_hash.py` 를 `release_pipeline.py version-bump` 의 *post-step* 으로 자동 호출 (1 commit 통합)
- **TASK-V0726-004**: detached HEAD 시의 memory dir 의 *cleanup policy* — short SHA 의 commit 이 *age out* 되면? (e.g. 30일 후 자동 archive)

## Metric

- v0.7.26 = 1 feat commit (TBD) + 1 chore (TBD) = 2 commit
- 1 신규 tool (sync_release_hash.py, ~180 line)
- 1 workflow_kit 변경 (paths.py, ~15 line)
- 2 신규 test file (10 smoke, 10/10 PASS)
- 누적 test 217+ (v0.7.25 207+ + 10 신규)
- 21 release 누적 (v0.7.5~v0.7.26)
- 79 commit code-repo (v0.7.25 까지) + 2 commit = **81 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
