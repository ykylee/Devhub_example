# Beta v0.7.29 — TASK-V0727-001 (Post-Step 2-Phase + Amend Integration)

> **Status**: in-flight
> 본 release 의 변경. release_pipeline.py version-bump 의 post-step 을 *2-phase* (sync + amend) 로 확장. v0.7.27 의 TASK-V0726-003 post-step 의 *closure* — 별도 fix(state) commit 불필요.

## 본 release 의 1차 출처

1. **v0.7.27 release note** (TASK-V0726-003 sync_release_hash post-step) — 본 release 의 *post-step 의 phase 1* (sync_release_hash 자동 호출) 의 1차 출처
2. **v0.7.21 release note** (--allow-existing-tag) — 본 release 의 *amend 의 위험* 의 *caller opt-in* 의 정공법
3. **v0.7.18 release note** (destructive subcommand 정공법, --dry-run 필수 + apply 시 graceful fail) — 본 release 의 *amend 의 graceful fail* 의 1차 출처
4. **v0.7.25 release note** (F-6 closure, SHA256 hash 기반 idempotency) — 본 release 의 *2-phase 의 idempotency* 의 1차 출처
5. **v0.7.28 release note** (TASK-V0726-004 archive_stale_memory) — 본 release 의 *amend 후의 final_hash* 의 *post-step* 검증

## 발견 (v0.7.27 의 follow-up)

v0.7.27 의 TASK-V0726-003 (post-step sync_release_hash) 의 *문제점*:
- post-step 의 sync 가 *별도 commit* 없이 *file 변경만*
- *caller* 가 본 release 의 *fix(state) commit* 별도로 작성 (v0.7.26 1 fix(state), v0.7.27 1 fix(state), v0.7.28 1 fix(state))
- 즉, *본 release 의 본 release 의 hash* 의 정합 = `feat + chore + 1+ fix(state)` = **3 commit**

**v0.7.29 의 정공법**:
- post-step 의 *2-phase* 통합: (1) sync_release_hash 자동 호출 + (2) `git add` + `git commit --amend --no-edit`
- 결과: *별도 fix(state) commit 불필요* — `feat + chore` = **2 commit**

## 본 release 의 변경

### 1. `release_pipeline.py:_run_post_step_sync_hash` (TASK-V0727-001 2-phase)

**Before** (v0.7.27):
```python
def _run_post_step_sync_hash(version: str) -> dict:
    sync_tool = Path(__file__).resolve().parent / "sync_release_hash.py"
    version_arg = f"v{version}" if not version.startswith("v") else version
    proc = subprocess.run([sys.executable, sync_tool, f"--version={version_arg}", "--apply"], ...)
    return {"ok": proc.returncode == 0, "stdout": ..., "stderr": ..., "returncode": ...}
```

**After** (v0.7.29, 2-phase):
```python
def _run_post_step_sync_hash(version: str) -> dict:
    sync_tool = Path(__file__).resolve().parent / "sync_release_hash.py"
    version_arg = f"v{version}" if not version.startswith("v") else version

    # Phase 1: sync_release_hash.py 자동 호출 (TASK-V0726-003)
    proc_sync = subprocess.run([sys.executable, sync_tool, f"--version={version_arg}", "--apply"], ...)
    if proc_sync.returncode != 0:
        return {"ok": False, "error": f"sync_release_hash failed (returncode={proc_sync.returncode})", ...}

    # Phase 2: git add + git commit --amend (TASK-V0727-001)
    # amend 시 *HEAD* 의 *직전* commit (feat or chore) 이 amend 됨
    # sync_release_hash 의 변경 = state.json + backlog 의 TBD → HEAD hash
    # → 별도 fix(state) commit 불필요
    proc_add = subprocess.run(["git", "add", "-A"], ...)
    if proc_add.returncode != 0:
        return {"ok": False, "error": f"git add failed (returncode={proc_add.returncode})", ...}

    proc_amend = subprocess.run(["git", "commit", "--amend", "--no-edit"], ...)
    if proc_amend.returncode != 0:
        return {"ok": False, "error": f"git commit --amend failed (returncode={proc_amend.returncode})", ...}

    # final hash (amend 후의 HEAD)
    proc_hash = subprocess.run(["git", "rev-parse", "HEAD", "--short=7"], ...)
    final_hash = proc_hash.stdout.strip().splitlines()[-1][:7] if proc_hash.returncode == 0 else None

    return {"ok": True, "sync_result": ..., "amend_result": ..., "final_hash": final_hash, "error": None}
```

### 2. 5 smoke test (5/5 PASS)

`tests/check_v0_7_29_poststep_amend.py` (in-process import + mock):
1. `test_2_phase_sync_calls` — sync_release_hash + git add + git commit --amend + git rev-parse 모두 호출 (4 calls)
2. `test_amend_integration` — amend 후의 final_hash 가 결과 dict 에 포함
3. `test_no_tbd_skip` — TBD 부재 시 sync 가 *변경 0* (idempotent), amend 도 *no-op* (4 calls 모두 발생)
4. `test_sync_failure_graceful` — sync fail (returncode 1) 시 amend 호출 *안 함* (1 call), ok=False, error message 포함
5. `test_amend_failure_graceful` — git commit --amend fail (returncode 1) 시 ok=False, final_hash=None, error message 포함

**5/5 PASS** (227+ → 232+ 누적 test).

### 3. cumulative 232+ test (회귀 0)

- v0.7.28 의 5 신규 test = 227+
- v0.7.29 의 5 신규 test (TASK-V0727-001 5) = **232+**
- 기존 v0.7.27 의 5 test (TASK-V0726-003) 회귀 0
- 기존 `version-bump --dry-run` 정상 (회귀 0)

## 발견된 cross-cutting lesson (v0.7.29)

- **2-phase 의 *sequential dependency***: phase 1 (sync) fail → phase 2 (amend) 호출 *안 함*. phase 2 (amend) fail → *amend 후의 hash* = None. *sequential dependency* 의 *graceful fail* 정공법.
- **`git commit --amend` 의 *HEAD 기준***: amend 는 *HEAD* 의 *직전* commit 만 amend. feat commit 의 *직전* = feat commit 의 *직전* 의 *직전* commit (즉, 본 release 의 chore commit 의 *직전* = feat commit amend). 본 release 의 *2-commit* 정공법 (feat + chore) → *HEAD* 의 chore commit amend → *1 commit 통합*.
- **amend 의 *force push* 의 위험**: amend 가 remote 에 *이미 push* 된 commit 의 hash 변경 → force push 필요. 본 release 의 *local-only* amend (push *전*) → force push 불필요. **local amend** vs **remote amend** 의 *구분* 이 *critical*.
- **`--no-edit` flag 의 의미**: amend 시 commit message *변경 안 함*. 본 release 의 chore commit message 가 *그대로* 유지 (amend 가 file 변경만 추가). **commit message 의 *immutability*** 정공법.
- **post-step 의 *subprocess count* 의 *일관성***: 본 release 의 version-bump 가 4 subprocess call (sync + add + amend + rev-parse). v0.7.27 의 version-bump 가 1 subprocess call (sync). **subprocess count 의 *일관성* = test 의 *호출 횟수* 검증이 *정공법***.

## Reference (다른 release note)

- v0.7.28 release note (TASK-V0726-004 archive_stale_memory, 본 release 의 *amend 후의 final_hash* 의 *post-step* 검증)
- v0.7.27 release note (TASK-V0726-003 sync_release_hash post-step, 본 release 의 *phase 1* 의 1차 출처)
- v0.7.25 release note (F-6 closure, 본 release 의 *2-phase 의 idempotency* 의 1차 출처)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 *amend 의 위험* 의 *caller opt-in* 의 정공법)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *amend 의 graceful fail* 의 1차 출처)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 1 TASK (본 release)

### TASK-V0727-001: Post-step 2-phase + amend integration

- **commit**: TBD
- **status**: in-flight
- **scope**:
  - `release_pipeline.py:_run_post_step_sync_hash` (~50 line 추가: 2-phase implementation + amend + rev-parse + graceful fail)
  - `tests/check_v0_7_29_poststep_amend.py` (5 smoke, 5/5 PASS)

## Follow-up (v0.7.30+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0728-001**: `archive_stale_memory.py` 의 *cron* 자동 호출 (e.g. `mavis cron self archive-memory --every 7d --prompt "..."`) — *manual* → *automated*
- **TASK-V0729-001**: post-step 의 *interactive mode* — caller 가 *amend* vs *new commit* 을 *선택* (CLI flag `--commit-mode=amend|new`, default = amend)

## Metric

- v0.7.29 = 1 feat commit (TBD) + 1 chore (TBD) = 2 commit (v0.7.28 의 3 commit 대비 *33% 감소*)
- release_pipeline 변경 (~50 line: 2-phase implementation)
- 1 신규 test file (5 smoke, 5/5 PASS)
- 누적 test 232+ (v0.7.28 227+ + 5 신규)
- 24 release 누적 (v0.7.5~v0.7.29)
- 86 commit code-repo (v0.7.28 까지) + 2 commit = **88 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
