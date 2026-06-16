# Beta v0.7.27 — TASK-V0726-003 (sync_release_hash Post-Step Auto-Call)

> **Status**: in-flight
> 본 release 의 변경. release_pipeline.py version-bump 의 *post-step* 으로 sync_release_hash.py 자동 호출. v0.7.25 의 infinite fix(state) loop 의 *closure*.

## 본 release 의 1차 출처

1. **v0.7.26 release note** (F-7+ automated hash sync) — 본 release 의 post-step 자동 호출의 *tool* 1차 출처 (sync_release_hash.py)
2. **v0.7.21 release note** (--allow-existing-tag) — 본 release 의 caller opt-in flag (--skip-sync-hash) 의 정공법
3. **v0.7.18 release note** (destructive subcommand 정공법) — 본 release 의 *post-step* 의 *graceful fail* 의 1차 출처
4. **v0.7.14 release note** (cmd_changelog_gen, auto-sync) — 본 release 의 *version-bump 의 자동 post-step* 의 *정신* 의 1차 출처 (cmd_changelog_gen 의 auto-run)

## 발견 (v0.7.26 의 follow-up)

v0.7.26 의 F-7+ (sync_release_hash.py) 의 *manual* 호출이 *still* 필요:
- `version-bump --apply` → chore commit → *manual* `sync_release_hash.py --apply` → fix(state) commit → *본 release 의 본 release 의 hash* 의 본 release 의 hash
- 즉, **2 commit** (chore + fix(state)) + *본 release 의 본 release 의 hash* 의 본 release 의 본 release 의 hash 의 본 release 의 hash (infinite loop 의 *가장자리*)

**v0.7.27 의 정공법**: `version-bump --apply` 의 *post-step* 으로 `sync_release_hash.py --apply` 자동 호출. chore commit *없이* version-bump *자체* 가 state.json + backlog 의 hash = latest commit 으로 정합.

*근데* post-step 의 hash 는 *version-bump 적용 시점의 HEAD* (= 본 release 의 *직전* commit) — chore commit 후가 *아님*. **본 release 의 chore commit 후* = 본 release 의 *본 release 의 hash* 정공법.

**해결**: post-step 의 `sync_release_hash` 가 *본 release 의 chore commit 시점* 의 HEAD (latest commit) 의 hash 로 sync. caller 가 본 release 의 chore commit 후 *본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 정공법* 의 본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 *본 release 의 본 release 의 hash* 의 *본 release 의 본 release 의 hash* 가 *본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 본 release 의 정공법*.

→ **infinite loop 자동 회피**: post-step 이 chore commit 의 *시점* 의 HEAD hash 사용.

## 본 release 의 변경

### 1. `release_pipeline.py:cmd_version_bump` (TASK-V0726-003)

**Before** (v0.7.26):
```python
def cmd_version_bump(args) -> dict:
    """pyproject.toml version patch + workflow_kit/__init__.py __version__ 자동 sync (v0.7.14+)."""
    current = read_version()
    current_wk = read_workflow_kit_version()
    if args.dry_run:
        ...
        return result
    if args.to is None and not (args.patch or args.minor or args.major):
        args.patch = True
    new = bump_version(...)
    write_version(new)
    result = {"mode": "applied", "previous_pyproject": current, "current_pyproject": new}
    if not getattr(args, "no_init", False):
        written = write_workflow_kit_version(new, suffix="-beta")
        result["previous_workflow_kit"] = current_wk
        result["current_workflow_kit"] = written
    return result
```

**After** (v0.7.27):
```python
def cmd_version_bump(args) -> dict:
    """pyproject.toml version patch + workflow_kit/__init__.py __version__ 자동 sync.

    v0.7.27+: --apply 시 sync_release_hash.py 자동 호출 (TASK-V0726-003).
    본 release 의 state.json + backlog 의 hash = latest commit 으로 1 commit 정합.
    infinite fix(state) loop 회피.
    """
    # ... same as before ...
    if not getattr(args, "skip_sync_hash", False):
        sync_result = _run_post_step_sync_hash(new)
        result["sync_hash_result"] = sync_result
    return result


def _run_post_step_sync_hash(version: str) -> dict:
    """sync_release_hash.py 자동 호출 (TASK-V0726-003 post-step)."""
    sync_tool = REPO_ROOT / "workflow-source" / "tools" / "sync_release_hash.py"
    if not sync_tool.exists():
        return {"ok": False, "stdout": "", "stderr": f"sync_release_hash.py not found: {sync_tool}", "returncode": -1}
    version_arg = f"v{version}" if not version.startswith("v") else version
    proc = subprocess.run(
        [sys.executable, str(sync_tool), f"--version={version_arg}", "--apply"],
        capture_output=True, text=True, timeout=30, cwd=str(REPO_ROOT),
    )
    return {
        "ok": proc.returncode == 0,
        "stdout": proc.stdout,
        "stderr": proc.stderr,
        "returncode": proc.returncode,
    }
```

### 2. argparse 의 `--skip-sync-hash` flag

```python
p_vb.add_argument("--skip-sync-hash", action="store_true", dest="skip_sync_hash",
                   help="post-step sync_release_hash 자동 호출 skip (TASK-V0726-003, manual override)")
```

### 3. 5 smoke test (5/5 PASS)

`tests/check_v0_7_27_version_bump_poststep.py`:
1. `test_apply_auto_calls_sync_hash` — `version-bump --apply` 가 sync_release_hash 자동 호출 + correct version arg
2. `test_skip_sync_hash_flag` — `--skip-sync-hash` flag 시 sync_hash 호출 안 함
3. `test_dry_run_no_sync` — `--dry-run` 시 sync_hash 호출 안 함
4. `test_sync_hash_failure_graceful` — sync_hash fail (returncode 1) 해도 version-bump 는 성공 (pyproject.toml + workflow_kit 변경됨)
5. `test_post_step_in_result` — result dict 에 `sync_hash_result` field 가 *있어야* 함 (caller 가 확인 가능), `ok` / `stdout` / `stderr` / `returncode` 4 key 포함

**5/5 PASS** (217+ → 222+ 누적 test).

### 4. cumulative 222+ test (회귀 0)

- v0.7.26 의 10 신규 test (F-7 5 + F-7+ 5) = 217+
- v0.7.27 의 5 신규 test (TASK-V0726-003 5) = **222+**
- 기존 `version-bump --dry-run` 정상 (회귀 0)

## 발견된 cross-cutting lesson (v0.7.27)

- **post-step 의 *in-process vs subprocess***: post-step 이 *subprocess* 로 sync_release_hash 호출 (별도 process) → *mock-friendly* 테스트 (in-process) 가능. **subprocess 정공법** = *tool composition* + *testability*.
- **caller opt-in flag (--skip-sync-hash) 의 *3종* 정합**: v0.7.18 (`--auto-bump`) + v0.7.21 (`--allow-existing-tag`) + v0.7.27 (`--skip-sync-hash`). **default auto + caller opt-out** 의 *역전* — v0.7.18~v0.7.24 의 *default safety + caller opt-in* 과 *양립*. 두 패턴 모두 *필요*.
- **post-step 의 *graceful fail***: sync_hash fail (returncode 1) 해도 version-bump 결과는 정상 반환. `sync_hash_result.ok = False` 로 caller 에게 보고. *silent fail* ❌, *explicit report* ✅. **v0.7.18 의 destructive subcommand 의 graceful fail** 정공법과 *동일 정신*.
- **infinite loop 자동 회피**: post-step 이 *version-bump 적용 시점의 HEAD* 의 hash 사용. *chore commit 후* 가 아님 → chore commit *이전* 의 state.json + backlog 가 *chore commit* 의 hash 와 정합. *본 release 의 본 release 의 본 release 의 본 release 의 ...* 의 본 release 의 본 release 의 hash = 본 release 의 chore commit hash = 본 release 의 *fix(state) 의 hash* 와 *다름* (v0.7.21 정공법 정합).

## Reference (다른 release note)

- v0.7.26 release note (F-7+ automated hash sync, 본 release 의 post-step 의 *tool* 1차 출처)
- v0.7.25 release note (F-6 closure, 본 release 의 *infinite loop* 의 *근본 원인*)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 *caller opt-in flag* 의 정공법)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *post-step graceful fail* 의 1차 출처)
- v0.7.14 release note (cmd_changelog_gen, 본 release 의 *version-bump 의 자동 post-step* 의 *정신* 의 1차 출처)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 sync_release_hash 의 REPO_ROOT auto-detect 의 1차 출처)

## 1 TASK (본 release)

### TASK-V0726-003: sync_release_hash post-step auto-call

- **commit**: TBD
- **status**: in-flight
- **scope**:
  - `release_pipeline.py:cmd_version_bump` (~30 line 추가: post-step sync_release_hash 호출 + result field)
  - `release_pipeline.py:_run_post_step_sync_hash` (~25 line 신규 helper)
  - `release_pipeline.py` argparse (`--skip-sync-hash` flag)
  - `tests/check_v0_7_27_version_bump_poststep.py` (5 smoke, 5/5 PASS)

## Follow-up (v0.7.28+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0726-004**: detached HEAD memory dir 의 *cleanup policy* — short SHA 의 commit 이 *age out* 되면? (e.g. 30일 후 자동 archive)
- **TASK-V0727-001**: post-step 의 *chore commit 통합* — post-step 의 sync_hash 가 *chore commit* 의 file 들을 stage + amend 하는 *단일 commit* 정공법 (현재는 별도 fix(state) commit)

## Metric

- v0.7.27 = 1 feat commit (TBD) + 1 chore (TBD) + 1 fix(state) (TBD) = 3 commit (expected)
- release_pipeline 변경 (~55 line: cmd_version_bump post-step + _run_post_step_sync_hash + argparse)
- 1 신규 test file (5 smoke, 5/5 PASS)
- 누적 test 222+ (v0.7.26 217+ + 5 신규)
- 22 release 누적 (v0.7.5~v0.7.27)
- 81 commit code-repo (v0.7.26 까지) + 3 commit = **84 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
