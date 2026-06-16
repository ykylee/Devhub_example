# Beta v0.7.32 — TASK-V0731-001 + TASK-V0732-001 (Log Rotation + Metrics Aggregation)

> **Status**: in-flight
> 본 release 의 변경. v0.7.31 의 metrics log 의 *2 follow-up* — log rotation (10000 line threshold) + metrics aggregation (weekly/monthly/daily/all).

## 본 release 의 1차 출처

1. **v0.7.31 release note** (TASK-V0729-001 + V0730-001 metrics log + idempotency) — 본 release 의 *2 follow-up* 의 1차 출처
2. **v0.7.18 release note** (destructive subcommand 정공법, --dry-run 필수) — 본 release 의 *log rotation 의 destructive subcommand* 의 정공법
3. **v0.7.25 release note** (F-6 closure, SHA256 idempotency) — 본 release 의 *rotation 의 truncated log 의 idempotency* 의 1차 출처

## 발견 (v0.7.31 의 2 follow-up)

### TASK-V0731-001: log rotation 부재
v0.7.31 의 `archive_stale_memory.log` 가 *unbounded* — *시간이 지나면* log file *누적* → disk 사용량 *선형 증가* → post-mortem 분석 시 *읽기* *느림*.

**v0.7.32 의 정공법**: `LOG_ROTATE_LINE_THRESHOLD = 10000` (default) line 초과 시 `archive_stale_memory.log.YYYY-MM-DDTHH-MM-SSZ.gz` 로 gzip + truncate. caller opt-in `--rotate-logs` flag + `--rotate-threshold=N` config.

### TASK-V0732-001: metrics aggregation 부재
v0.7.31 의 `--show-metrics` 가 *raw entries* 만 반환 — caller 가 *weekly/monthly* aggregation *수동* 계산 필요. *장기 추세* 분석 어려움.

**v0.7.32 의 정공법**: `--aggregate=weekly|monthly|daily|all` flag. `aggregate_metrics()` 가 *period* 별 bucket + total 반환. `--include-rotated` flag — rotated log (*.gz) entries 도 *include*.

## 본 release 의 변경

### 1. `archive_stale_memory.py` 의 2 subcommand + 3 helper (TASK-V0731-001 + TASK-V0732-001)

**3 신규 helper**:
- `rotate_log_if_needed(memory_dir, *, line_threshold)` — `line_count <= line_threshold` 시 no rotate. 초과 시 gzip + truncate.
- `read_rotated_logs(memory_dir)` — 모든 rotated log (*.gz) 의 entries read. `source` field 추가.
- `aggregate_metrics(entries, *, period)` — period 별 bucket + total. ISO 8601 week / month / day / all.

**2 신규 subcommand**:
- `cmd_rotate_logs(args)` + `--rotate-logs` + `--rotate-threshold` flag
- `cmd_aggregate_metrics(args)` + `--aggregate` + `--include-rotated` flag

**log rotation logic**:
```python
def rotate_log_if_needed(memory_dir, *, line_threshold=10000):
    log_path = memory_dir / METRICS_LOG_NAME
    if not log_path.exists(): return {"rotated": False, ...}
    with log_path.open("r") as f:
        line_count = sum(1 for _ in f)
    if line_count <= line_threshold: return {"rotated": False, ...}
    timestamp = dt.datetime.now(tz=dt.timezone.utc).strftime("%Y-%m-%dT%H-%M-%SZ")
    archive_path = memory_dir / f"{METRICS_LOG_NAME}.{timestamp}.gz"
    with log_path.open("rb") as f_in, gzip.open(archive_path, "wb") as f_out:
        shutil.copyfileobj(f_in, f_out)
    log_path.write_text("", encoding="utf-8")
    return {"rotated": True, "line_count": line_count, "archive_path": str(archive_path), ...}
```

**aggregation logic (period 별 bucket)**:
```python
def aggregate_metrics(entries, *, period="weekly"):
    if period == "all" or not entries:
        total = {sum(e.values() for e in entries)}
        return {"period": "all", "buckets": [], "total": total}
    def bucket_key(ts_str):
        ts = dt.datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
        if period == "daily": return ts.strftime("%Y-%m-%d")
        elif period == "weekly":
            iso = ts.isocalendar()
            return f"{iso.year}-W{iso.week:02d}"
        elif period == "monthly": return ts.strftime("%Y-%m")
    bucket_data = defaultdict(list)
    for e in entries:
        bucket_data[bucket_key(e["timestamp"])].append(e)
    buckets = [{"bucket": bk, "archived": sum(...), ...} for bk in sorted(bucket_data)]
    return {"period": period, "buckets": buckets, "total": total}
```

### 2. 10 smoke test (10/10 PASS, 5-run stable)

`tests/check_v0_7_32_rotate_aggregate.py` (in-process import + mock):

**TASK-V0731-001 (5 smoke)**:
1. `test_rotate_log_under_threshold` — line < threshold → no rotate
2. `test_rotate_log_at_threshold` — line = threshold → no rotate (`<=` strict)
3. `test_rotate_log_over_threshold` — line > threshold → gzip + truncate (log 0 byte)
4. `test_rotate_log_creates_gz_archive` — archive path = `archive_stale_memory.log.YYYY-MM-DDTHH-MM-SSZ.gz`
5. `test_rotate_log_read_rotated` — `read_rotated_logs` 가 gz 의 entries read + `source` field

**TASK-V0732-001 (5 smoke)**:
6. `test_aggregate_weekly` — 7 entries (1 week) → 1 bucket (ISO week)
7. `test_aggregate_monthly` — 30 entries (1 month) → 1 bucket
8. `test_aggregate_multiple_weeks` — 14 entries (2 weeks) → 2 buckets (W25 + W26)
9. `test_aggregate_include_rotated` — `--include-rotated` 시 main + rotated entries 모두 포함 (105 entries total)
10. `test_aggregate_invalid_period` — invalid `--aggregate=yearly` → ok=False, error message

**10/10 PASS** (247+ → 257+ 누적 test, 5-run stable).

### 3. cumulative 257+ test (회귀 0)

- v0.7.31 의 10 신규 test = 247+
- v0.7.32 의 10 신규 test (TASK-V0731-001 5 + TASK-V0732-001 5) = **257+**
- 기존 v0.7.28/v0.7.30/v0.7.31 의 test 회귀 0

## 발견된 cross-cutting lesson (v0.7.32)

- **line count 의 *`<=` (less-than-or-equal)* 정공법**: `line_count <= line_threshold` 시 no rotate. `line_count > line_threshold` 시 rotate. *strict greater-than* 만 trigger — *under-or-equal* 는 no rotate. **v0.7.32 의 `<=` (현재) vs `<` (이전) 의 *의미 차이* — *less than* (이전) 는 == 시 rotate, *less-or-equal* (현재) 는 == 시 no rotate. *correct* 정공법은 `<=`.
- **gzip rotation 의 *truncate-after-compress* 정공법**: gzip write 완료 후 `log_path.write_text("", ...)` 로 truncate. atomic 보장 *없음* (gzip write 와 truncate 사이 *crash* 시 *main log + archive 둘 다 손실*). v0.7.33 의 *atomic rotation* 의 follow-up.
- **ISO 8601 week 의 *Monday-Sunday***: `ts.isocalendar()` 가 `(year, week, weekday)` 반환. `f"{year}-W{week:02d}"` 으로 bucket key. **ISO 8601 week** 정공법 (US week (Sunday-Saturday) 와 다름).
- **`ts.isocalendar()` 의 *Python 3.9+* return type**: `IsoCalendarDate` namedtuple. *tup[0] = year*, `tup[1] = week`. **3.9+ 정공법**.
- **`--aggregate=yearly` 같은 *invalid period* 의 *명시적 error***: silent return None ❌, *명시적 error message* + 사용 가능 option 4개 명시 ✅. **v0.7.18 의 destructive subcommand 정공법** (명시적 error).
- **`read_rotated_logs` 의 *dedup* 정공법**: `cmd_aggregate_metrics` 가 *main_ts = {e.timestamp for e in main_entries}`* set → `rotated_filtered = [e for e in rotated if e.timestamp not in main_ts]`. main log 가 *최신* → *main 우선* dedup. **post-rotation 의 main empty 시* 모든 entry 가 rotated → main_ts = {} → 모든 rotated include.
- **5-run stable 의 *중요성***: `1 < 100` 같은 *boundary* 의 race condition 회피. **5 consecutive run PASS** = test 의 *stability* 보장.

## Reference (다른 release note)

- v0.7.31 release note (TASK-V0729-001 + V0730-001 metrics log, 본 release 의 *follow-up* 의 1차 출처)
- v0.7.30 release note (TASK-V0728-001 cron integration, 본 release 의 *cron subcommand 의 follow-up* 의 *역사* 의 일부)
- v0.7.28 release note (TASK-V0726-004 archive_stale_memory, 본 release 의 *archive_stale_memory.py 의 follow-up* 의 1차 출처)
- v0.7.25 release note (F-6 closure, 본 release 의 *read_rotated_logs 의 dedup* 의 1차 출처)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 *--include-rotated* flag 의 *caller opt-in* 정신의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *log rotation 의 destructive subcommand* 의 정공법)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 2 TASK (본 release)

### TASK-V0731-001: Log rotation (line > 10000 → gzip + truncate)
- **상태**: in-flight
- **commit**: TBD
- **scope**: `rotate_log_if_needed` / `read_rotated_logs` / `cmd_rotate_logs` (3 helper) + `--rotate-logs` / `--rotate-threshold` flag (~80 line)

### TASK-V0732-001: Metrics aggregation (weekly/monthly/daily/all)
- **상태**: in-flight
- **commit**: TBD
- **scope**: `aggregate_metrics` / `cmd_aggregate_metrics` (2 helper) + `--aggregate` / `--include-rotated` flag (~100 line)

## Follow-up (v0.7.33+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0733-001**: atomic rotation — `gzip write + truncate` 의 *crash safety* (write + truncate 사이 *crash* 시 *gzip archive + main log 둘 다 손실*)
- **TASK-V0734-001**: `--aggregate=yearly` 추가 (ISO 8601 year) + `bucket_key` 의 *daily/weekly/monthly/yearly* 의 *4-축* 지원

## Metric

- v0.7.32 = 1 feat commit (TBD) + 1 chore (TBD) = 2 commit (TASK-V0727-001 의 2-phase + amend 의 *1 commit* 정공법)
- archive_stale_memory.py 변경 (~180 line: 3 helper + 2 subcommand + argparse flags)
- 1 신규 test file (10 smoke, 10/10 PASS, 5-run stable)
- 누적 test 257+ (v0.7.31 247+ + 10 신규)
- 27 release 누적 (v0.7.5~v0.7.32)
- 94 commit code-repo (v0.7.31 까지) + 2 commit = **96 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
