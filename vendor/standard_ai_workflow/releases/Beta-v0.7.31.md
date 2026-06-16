# Beta v0.7.31 — TASK-V0729-001 + TASK-V0730-001 (Run-Time Metrics + Cron Idempotency)

> **Status**: in-flight
> 본 release 의 변경. v0.7.30 의 cron 통합의 *2 follow-up* — post-mortem 분석용 metrics log + cron install idempotency.

## 본 release 의 1차 출처

1. **v0.7.30 release note** (TASK-V0728-001 cron integration) — 본 release 의 *cron integration 의 follow-up* 의 1차 출처
2. **v0.7.18 release note** (destructive subcommand 정공법) — 본 release 의 *install-cron 의 idempotency skip* 의 정공법
3. **v0.7.25 release note** (F-6 closure, SHA256 hash 기반 idempotency) — 본 release 의 *read_metrics_log 의 idempotency* 의 1차 출처
4. **v0.7.27 release note** (TASK-V0726-003 sync_release_hash post-step) — 본 release 의 *post-step 의 caller opt-in flag* 정공법의 *역사* 의 일부

## 발견 (v0.7.30 의 2 follow-up)

### TASK-V0729-001: cron run-time metrics 부재

v0.7.30 의 cron auto-call 의 *문제점*: *caller* 가 *post-mortem 분석* 시 *archive_count* / *skip_count* / *error_count* 의 *log* 부재.
- *cron trigger* 가 성공했는지 *이후* 검증 어려움
- *누적* stats (월간 archive dir 수) *부재*

**v0.7.31 의 정공법**: `ai-workflow/memory/archive_stale_memory.log` 에 *append-only log*. caller 가 `--show-metrics` 로 read.

### TASK-V0730-001: install-cron 의 idempotency 부재

v0.7.30 의 `--install-cron` 의 *문제점*: *caller* 가 *재호출* 시 *매번 create 시도* → `mavis cron create` 의 *conflict* 가능.
- *수동* `mavis cron create` 로 만든 cron + *수동* `archive_stale_memory --install-cron` → conflict
- *caller discipline* 필요 (이전 cron *삭제* 후 install)

**v0.7.31 의 정공법**: `mavis cron info <agent> <cronName>` 으로 *existing check* → *이미* 존재 시 skip + report (ok=True, idempotency='skipped-existing'). caller opt-in `--force-install` flag 로 *skip 우회*.

## 본 release 의 변경

### 1. `archive_stale_memory.py` 의 5 subcommand + 2 helper (TASK-V0729-001 + TASK-V0730-001)

**2 신규 helper**:
- `append_metrics_log(memory_dir, older_than, archived_count, skipped_count, error_count)` — `archive_stale_memory.log` 에 1-line append. ISO 8601 timestamp + 4 count field (5 tab-separated).
- `read_metrics_log(memory_dir)` — log 의 모든 entry 를 dict list 로 read.

**1 신규 subcommand**:
- `cmd_show_metrics(args)` + `--show-metrics` flag — metrics log 의 모든 entry read + return.

**TASK-V0730-001: install-cron 의 idempotency**:
- `cmd_install_cron` 의 *first step* = `mavis cron info <agent> <cronName>` 으로 existing check
- *이미* 존재 시 skip + return (`ok=True`, `idempotency='skipped-existing'`)
- *신규* flag `--force-install` — skip 우회 (force create)
- result dict 에 `idempotency: 'skipped-existing' | 'created' | 'forced'` field 추가

**apply mode 의 metrics log append**:
```python
# apply loop 후
error_count = sum(1 for s in skipped if "fail" in s.get("reason", ""))
append_metrics_log(
    memory_dir=memory_dir,
    older_than=older_than,
    archived_count=len(archived),
    skipped_count=len(skipped),
    error_count=error_count,
)
```

**metrics log entry format** (5 tab-separated):
```
2026-06-15T13:00:00Z\tolder_than=30\tarchived=2\tskipped=1\terror=0\n
```

**TASK-V0729-001: cmd_show_metrics**:
```python
def cmd_show_metrics(args) -> dict:
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"ok": False, "error": f"memory dir not found: {memory_dir}", "entries": []}
    entries = read_metrics_log(memory_dir)
    return {"ok": True, "memory_dir": str(memory_dir), "log_path": str(memory_dir / METRICS_LOG_NAME),
            "entries": entries, "entry_count": len(entries)}
```

### 2. 10 smoke test (10/10 PASS)

`tests/check_v0_7_31_archive_metrics_idempotency.py` (in-process import + mock):

**TASK-V0729-001 (5 smoke)**:
1. `test_metrics_log_written` — apply 시 `archive_stale_memory.log` 에 entry append
2. `test_metrics_log_format` — entry format: timestamp \t older_than=N \t archived=N \t skipped=N \t error=N
3. `test_metrics_log_skipped_count` — file blocker 시뮬레이션 + skipped reason='already-archived' → error=0 (not 'fail')
4. `test_show_metrics_returns_entries` — `--show-metrics` 가 entries list 반환 (post-mortem 분석용)
5. `test_show_metrics_no_log` — log 부재 시 entry_count=0 + ok=True (graceful)

**TASK-V0730-001 (5 smoke)**:
6. `test_install_cron_idempotent_skip` — `mavis cron info` OK + cron_name in stdout → `skipped-existing` (no create)
7. `test_install_cron_creates_when_missing` — `mavis cron info` fail → `created` (2 calls)
8. `test_install_cron_force_install` — `--force-install` 시 info check skip (1 call only, idempotency='forced')
9. `test_uninstall_cron_idempotent` — `mavis cron disable` idempotent (no-op on disabled)
10. `test_metrics_log_idempotency` — `read_metrics_log` 가 동일 file 2번 read 시 동일 result (no state side-effect)

**10/10 PASS** (237+ → 247+ 누적 test).

### 3. cumulative 247+ test (회귀 0)

- v0.7.30 의 5 신규 test = 237+
- v0.7.31 의 10 신규 test (TASK-V0729-001 5 + TASK-V0730-001 5) = **247+**
- 기존 v0.7.28/v0.7.30 의 test 회귀 0 (`check_already_archived` fix 가 v0.7.28 의 test 3 의 setup 이 *file* blocker 시 *정상* 동작 보장)

## 발견된 cross-cutting lesson (v0.7.31)

- **append-only metrics log 의 *3-field-정합***: archived_count + skipped_count + error_count 의 *3-축* 정합. `error_count = sum(1 for s in skipped if "fail" in s.get("reason", ""))` — *reason* keyword match 정공법.
- **`mavis cron info` 의 *existence check* 정공법**: `info <agent> <name>` 의 returncode 0 + stdout 에 `cronName` 매치 → existing. 본 release 의 *idempotency skip* 의 1차 출처.
- **`--force-install` flag 의 *caller opt-in* 정신**: default safety + caller opt-in. v0.7.21 (`--allow-existing-tag`) 와 동일. *default idempotency + caller force*.
- **ISO 8601 timestamp + tab-separated format**: log entry 의 *parseability* + *machine-readable*. `str.split("\t")` + `key.split("=")` 의 *2-level parse* 정공법.
- **dst 가 file 인 경우의 *덮어쓰기 방지***: `dst.exists() and dst.is_file()` 분기 (TASK-V0730-001 의 *부수효과* — *archive target file* 시 *덮어쓰기* 안 함, *skip* + report). v0.7.28 의 `check_already_archived` 도 file 도 catch 로 강화.

## Reference (다른 release note)

- v0.7.30 release note (TASK-V0728-001 cron integration, 본 release 의 *follow-up* 의 1차 출처)
- v0.7.28 release note (TASK-V0726-004 archive_stale_memory, 본 release 의 *metrics log 의 file* 의 *in-repo path* 의 1차 출처)
- v0.7.27 release note (TASK-V0726-003 sync_release_hash post-step, 본 release 의 *post-step 의 caller opt-in flag* 정공법의 *역사* 의 일부)
- v0.7.25 release note (F-6 closure, 본 release 의 *SHA256 idempotency* 의 1차 출처)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 *--force-install* flag 의 *caller opt-in* 정신의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *install-cron 의 idempotency skip* 의 정공법)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 2 TASK (본 release)

### TASK-V0729-001: Run-time metrics log
- **상태**: in-flight
- **commit**: TBD
- **scope**: `append_metrics_log` / `read_metrics_log` / `cmd_show_metrics` (3 helper) + `--show-metrics` flag + `apply_metrics_log` 자동 호출 (~80 line)

### TASK-V0730-001: Install-cron idempotency
- **상태**: in-flight
- **commit**: TBD
- **scope**: `cmd_install_cron` 의 *idempotency check* (info + create) + `--force-install` flag + `check_already_archived` file catch (~30 line)

## Follow-up (v0.7.32+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0731-001**: `append_metrics_log` 의 *rotation* — log 가 *10000 line* 초과 시 *gzip + rotate* (post-mortem 의 *장기 보관*)
- **TASK-V0732-001**: `--show-metrics` 의 *aggregation* — 주간/월간 *sum* + *trend* (caller 가 *장기 추세* 분석 가능)

## Metric

- v0.7.31 = 1 feat commit (TBD) + 1 fix (TBD) = 2 commit (TASK-V0727-001 의 2-phase + amend 의 *1 commit* 정공법)
- archive_stale_memory.py 변경 (~110 line: 3 helper + 1 subcommand + idempotency + check_already_archived fix)
- 1 신규 test file (10 smoke, 10/10 PASS) + 1 v0.7.28 test fix (file catch)
- 누적 test 247+ (v0.7.30 237+ + 10 신규)
- 26 release 누적 (v0.7.5~v0.7.31)
- 91 commit code-repo (v0.7.30 까지) + 2 commit = **93 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
