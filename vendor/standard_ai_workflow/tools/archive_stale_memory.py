#!/usr/bin/env python3
"""
v0.7.28: TASK-V0726-004 — detached HEAD memory dir cleanup (stale archive)

F-7 (v0.7.26) 의 fix 로 detached HEAD 시 memory dir name = 7-char short SHA.
*시간이 지나면* 옛 commit 의 short SHA dir 가 누적 (e.g. 30+ days).
본 tool 은 *age-based* auto-archive: N day 이전의 short SHA dir → archive/ 로 move.

Usage:
    # 1. dry-run: 어떤 dir 이 archive 후보인지 list
    python3 tools/archive_stale_memory.py --older-than=30 --dry-run

    # 2. apply: 실제 archive (move to archive/<date>/<sha>/)
    python3 tools/archive_stale_memory.py --older-than=30 --apply

    # 3. list: 모든 short SHA dir (age 무관)
    python3 tools/archive_stale_memory.py --list --dry-run

    # 4. cleanup: apply 와 동일, --apply alias
    python3 tools/archive_stale_memory.py --older-than=30 --cleanup

    # 5. REPO_ROOT override
    python3 tools/archive_stale_memory.py --repo-root=/path/to/repo --older-than=30 --dry-run

1차 출처:
- v0.7.17 release note (in-repo redirect, memory dir 의 SSOT)
- v0.7.26 release note (F-7 branch detection, detached HEAD → short SHA)
- v0.7.18 release note (destructive subcommand 정공법, --dry-run 필수)
- v0.7.25 release note (F-6 closure, SHA256 hash 기반 idempotency)
"""
from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

# === Path setup ===
_LEGACY_REPO_ROOT = Path.home() / "repos" / "standard_ai_workflow_minimax"
_DEPRECATION_WARNED = False


def get_repo_root(cli_value: str | os.PathLike[str] | None = None, *, _suppress_warning: bool = False) -> Path:
    """REPO_ROOT 결정 (priority: CLI flag > env var > git rev-parse > legacy fallback).
    v0.7.12 의 4-priority 정공법.
    """
    global _DEPRECATION_WARNED

    if cli_value is not None:
        p = Path(cli_value).expanduser().resolve()
        return p

    env_val = os.environ.get("STANDARD_AI_WF_REPO")
    if env_val:
        p = Path(env_val).expanduser().resolve()
        return p

    try:
        import subprocess as sp
        proc = sp.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, timeout=5,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return Path(proc.stdout.strip()).resolve()
    except (Exception,):
        pass

    if not _DEPRECATION_WARNED and not _suppress_warning:
        _DEPRECATION_WARNED = True
        print(
            f"[DEPRECATION] archive_stale_memory.py: REPO_ROOT auto-detect 실패 — legacy fallback 사용 ({_LEGACY_REPO_ROOT}).",
            file=sys.stderr,
        )
    return _LEGACY_REPO_ROOT


# === Memory dir path ===
def get_memory_dir(repo_root: Path) -> Path:
    """`ai-workflow/memory/` 의 path. v0.7.17 의 in-repo redirect 의 SSOT."""
    return repo_root / "ai-workflow" / "memory"


# === Detection helpers ===
SHORT_SHA_RE = re.compile(r"^[a-f0-9]{7}$")


def is_short_sha_dir(path: Path) -> bool:
    """dir name 이 7-char hex (short SHA) 인지 확인. F-7 (v0.7.26) 의 detached HEAD 의
    dir name pattern. named branch (e.g. 'main', 'feat/x') 는 *불일치*."""
    if not path.is_dir():
        return False
    return bool(SHORT_SHA_RE.match(path.name))


def get_dir_age_days(path: Path) -> int:
    """dir 의 mtime 기준 age (day)."""
    mtime = path.stat().st_mtime
    mtime_dt = dt.datetime.fromtimestamp(mtime, tz=dt.timezone.utc)
    now_dt = dt.datetime.now(tz=dt.timezone.utc)
    return (now_dt - mtime_dt).days


def build_archive_path(memory_dir: Path, sha: str, archive_date: str) -> Path:
    """archive target path: `archive/<date>/<sha>/`."""
    return memory_dir / "archive" / archive_date / sha


def check_already_archived(memory_dir: Path, sha: str) -> bool:
    """sha 가 *이미* archive/ 하위에 있는지 확인. idempotency.

    v0.7.31 fix: file 도 catch (이전은 dir 만). dst 가 file (blocker) 인 경우에도
    *candidates build* 시점에 skip.
    """
    archive_base = memory_dir / "archive"
    if not archive_base.exists():
        return False
    for archive_entry in archive_base.iterdir():
        # file 도 검사 (sub-dir 일 필요 없음)
        candidate = archive_entry / sha
        if candidate.exists():
            return True
    return False


def sha256_of_dir(path: Path) -> str:
    """dir 의 SHA256 (file 들의 SHA256 concat). idempotency 의 1차 출처 (v0.7.25)."""
    h = hashlib.sha256()
    for f in sorted(path.rglob("*")):
        if f.is_file():
            h.update(f.read_bytes())
    return h.hexdigest()


# === Subcommand logic ===
def cmd_list(args: argparse.Namespace) -> dict:
    """모든 short SHA dir list (age 무관, dry-run 만)."""
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"error": f"memory dir not found: {memory_dir}", "items": []}

    items = []
    for entry in memory_dir.iterdir():
        if not is_short_sha_dir(entry):
            continue
        if check_already_archived(memory_dir, entry.name):
            continue
        age = get_dir_age_days(entry)
        items.append({
            "sha": entry.name,
            "path": str(entry),
            "age_days": age,
            "size_bytes": sum(f.stat().st_size for f in entry.rglob("*") if f.is_file()),
        })

    items.sort(key=lambda x: -x["age_days"])
    return {"mode": "list", "memory_dir": str(memory_dir), "items": items, "count": len(items)}


def cmd_archive(args: argparse.Namespace) -> dict:
    """age-based archive. dry-run or apply.

    Args:
        args: argparse.Namespace with --older-than, --dry-run/--apply, --repo-root.
    """
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"error": f"memory dir not found: {memory_dir}", "items": []}

    older_than = args.older_than
    if older_than < 0:
        return {"error": f"--older-than must be >= 0 (got {older_than})", "items": []}

    today = dt.date.today().strftime("%Y-%m-%d")
    candidates = []
    archived = []
    skipped = []

    for entry in memory_dir.iterdir():
        if not is_short_sha_dir(entry):
            continue
        if check_already_archived(memory_dir, entry.name):
            skipped.append({"sha": entry.name, "reason": "already-archived", "path": str(entry)})
            continue
        age = get_dir_age_days(entry)
        if age < older_than:
            skipped.append({"sha": entry.name, "reason": "too-recent", "age_days": age, "path": str(entry)})
            continue
        candidates.append({
            "sha": entry.name,
            "path": str(entry),
            "age_days": age,
            "archive_target": str(build_archive_path(memory_dir, entry.name, today)),
        })

    if args.dry_run:
        return {
            "mode": "dry-run",
            "memory_dir": str(memory_dir),
            "older_than": older_than,
            "candidates": candidates,
            "skipped": skipped,
            "candidate_count": len(candidates),
        }

    # apply mode
    for cand in candidates:
        src = Path(cand["path"])
        dst = Path(cand["archive_target"])
        dst.parent.mkdir(parents=True, exist_ok=True)
        if dst.exists() and dst.is_dir():
            skipped.append({"sha": cand["sha"], "reason": "archive-target-exists", "path": str(dst)})
            continue
        if dst.exists() and dst.is_file():
            # dst 가 *file* (blocker) — move-fail 시뮬레이션 (덮어쓰기 방지)
            skipped.append({"sha": cand["sha"], "reason": "move-fail: target is file (blocker), not dir", "path": str(src)})
            continue
        try:
            shutil.move(str(src), str(dst))
            archived.append({
                "sha": cand["sha"],
                "src": str(src),
                "dst": str(dst),
                "age_days": cand["age_days"],
                "sha256": sha256_of_dir(dst),
            })
        except (OSError, shutil.Error) as e:
            skipped.append({"sha": cand["sha"], "reason": f"move-fail: {e}", "path": str(src)})

    # TASK-V0729-001: run-time metrics log append (post-mortem 분석용)
    error_count = sum(1 for s in skipped if "fail" in s.get("reason", ""))
    append_metrics_log(
        memory_dir=memory_dir,
        older_than=older_than,
        archived_count=len(archived),
        skipped_count=len(skipped),
        error_count=error_count,
    )

    return {
        "mode": "apply",
        "memory_dir": str(memory_dir),
        "older_than": older_than,
        "archived": archived,
        "skipped": skipped,
        "archived_count": len(archived),
        "skipped_count": len(skipped),
    }


# === Run-time metrics log (TASK-V0729-001) ===
METRICS_LOG_NAME = "archive_stale_memory.log"


def append_metrics_log(
    memory_dir: Path,
    older_than: int,
    archived_count: int,
    skipped_count: int,
    error_count: int,
) -> bool:
    """apply mode 의 metrics 를 memory_dir/archive_stale_memory.log 에 append.

    Args:
        memory_dir: REPO_ROOT / "ai-workflow" / "memory" 의 path.
        older_than: apply 시 사용된 N day.
        archived_count: 성공 archive 된 dir 수.
        skipped_count: skip 된 dir 수 (already-archived, too-recent, archive-target-exists, move-fail 모두).
        error_count: skipped 중 "fail" reason (move-fail) 의 수.

    Returns:
        True if log written, False on error.
    """
    log_path = memory_dir / METRICS_LOG_NAME
    timestamp = dt.datetime.now(tz=dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    line = (
        f"{timestamp}\tolder_than={older_than}\tarchived={archived_count}\t"
        f"skipped={skipped_count}\terror={error_count}\n"
    )
    try:
        log_path.parent.mkdir(parents=True, exist_ok=True)
        with log_path.open("a", encoding="utf-8") as f:
            f.write(line)
        return True
    except (OSError, IOError):
        return False


def read_metrics_log(memory_dir: Path) -> list[dict]:
    """metrics log 의 모든 entry 를 dict list 로 read (post-mortem 분석용).

    Args:
        memory_dir: REPO_ROOT / "ai-workflow" / "memory" 의 path.

    Returns:
        list of {"timestamp": str, "older_than": int, "archived": int, "skipped": int, "error": int}.
    """
    log_path = memory_dir / METRICS_LOG_NAME
    if not log_path.exists():
        return []
    entries = []
    for line in log_path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        if len(parts) != 5:
            continue
        ts, ot, ar, sk, er = parts
        try:
            entries.append({
                "timestamp": ts,
                "older_than": int(ot.split("=")[1]),
                "archived": int(ar.split("=")[1]),
                "skipped": int(sk.split("=")[1]),
                "error": int(er.split("=")[1]),
            })
        except (ValueError, IndexError):
            continue
    return entries


def cmd_show_metrics(args: argparse.Namespace) -> dict:
    """metrics log 의 모든 entry 를 read + return.

    Returns:
        dict with keys: ok (bool), memory_dir, log_path, entries (list of dict), entry_count.
    """
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"ok": False, "error": f"memory dir not found: {memory_dir}", "entries": []}
    entries = read_metrics_log(memory_dir)
    return {
        "ok": True,
        "memory_dir": str(memory_dir),
        "log_path": str(memory_dir / METRICS_LOG_NAME),
        "entries": entries,
        "entry_count": len(entries),
    }


# === Log rotation (TASK-V0731-001) ===
LOG_ROTATE_LINE_THRESHOLD = 10000  # 10K line 초과 시 rotate


def rotate_log_if_needed(memory_dir: Path, *, line_threshold: int = LOG_ROTATE_LINE_THRESHOLD) -> dict:
    """metrics log 의 line 수 > threshold 시 gzip + rotate.

    Args:
        memory_dir: REPO_ROOT / "ai-workflow" / "memory".
        line_threshold: rotate trigger. default 10000 line.

    Returns:
        dict with keys: rotated (bool), line_count (int), threshold (int),
        archive_path (str | None), archive_size_bytes (int | None).
    """
    import gzip
    import shutil

    log_path = memory_dir / METRICS_LOG_NAME
    if not log_path.exists():
        return {"rotated": False, "line_count": 0, "threshold": line_threshold,
                "archive_path": None, "archive_size_bytes": None}
    # line count
    with log_path.open("r", encoding="utf-8") as f:
        line_count = sum(1 for _ in f)
    if line_count <= line_threshold:
        return {"rotated": False, "line_count": line_count, "threshold": line_threshold,
                "archive_path": None, "archive_size_bytes": None}
    # rotate: archive_stale_memory.log.2026-06-15T13-00-00Z.gz
    timestamp = dt.datetime.now(tz=dt.timezone.utc).strftime("%Y-%m-%dT%H-%M-%SZ")
    archive_path = memory_dir / f"{METRICS_LOG_NAME}.{timestamp}.gz"
    with log_path.open("rb") as f_in, gzip.open(archive_path, "wb") as f_out:
        shutil.copyfileobj(f_in, f_out)
    archive_size = archive_path.stat().st_size
    # truncate log
    log_path.write_text("", encoding="utf-8")
    return {
        "rotated": True,
        "line_count": line_count,
        "threshold": line_threshold,
        "archive_path": str(archive_path),
        "archive_size_bytes": archive_size,
    }


def cmd_rotate_logs(args: argparse.Namespace) -> dict:
    """metrics log 의 line count > threshold 시 gzip + rotate.

    Args:
        args: argparse.Namespace with --rotate-logs, --rotate-threshold.

    Returns:
        dict with keys: ok (bool), rotated (bool), line_count, threshold, archive_path, archive_size_bytes.
    """
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"ok": False, "error": f"memory dir not found: {memory_dir}"}
    threshold = getattr(args, "rotate_threshold", LOG_ROTATE_LINE_THRESHOLD)
    result = rotate_log_if_needed(memory_dir, line_threshold=threshold)
    return {
        "ok": True,
        "rotated": result["rotated"],
        "line_count": result["line_count"],
        "threshold": result["threshold"],
        "archive_path": result["archive_path"],
        "archive_size_bytes": result["archive_size_bytes"],
        "memory_dir": str(memory_dir),
    }


def read_rotated_logs(memory_dir: Path) -> list[dict]:
    """모든 rotated log (*.gz) 의 entries 를 read.

    Returns:
        list of {"timestamp": str, "older_than": int, "archived": int, "skipped": int, "error": int, "source": str}.
    """
    import gzip

    if not memory_dir.exists():
        return []
    entries = []
    # main log
    entries.extend(read_metrics_log(memory_dir))
    # rotated gz logs
    for path in sorted(memory_dir.glob(f"{METRICS_LOG_NAME}.*.gz")):
        try:
            with gzip.open(path, "rt", encoding="utf-8") as f:
                for line in f:
                    parts = line.strip().split("\t")
                    if len(parts) != 5:
                        continue
                    ts, ot, ar, sk, er = parts
                    entries.append({
                        "timestamp": ts,
                        "older_than": int(ot.split("=")[1]),
                        "archived": int(ar.split("=")[1]),
                        "skipped": int(sk.split("=")[1]),
                        "error": int(er.split("=")[1]),
                        "source": path.name,
                    })
        except (OSError, gzip.BadGzipFile):
            continue
    return entries


# === Metrics aggregation (TASK-V0732-001) ===
def aggregate_metrics(entries: list[dict], *, period: str = "weekly") -> dict:
    """metrics entries 를 *weekly* / *monthly* aggregation.

    Args:
        entries: read_metrics_log / read_rotated_logs 의 entries list.
        period: "weekly" / "monthly" / "daily" / "all".

    Returns:
        dict with keys: period, buckets (list of {"start": str, "end": str, "archived": int,
        "skipped": int, "error": int, "count": int}), total ({"archived": int, "skipped": int, "error": int, "count": int}).
    """
    from collections import defaultdict

    if period == "all" or not entries:
        total = {
            "archived": sum(e.get("archived", 0) for e in entries),
            "skipped": sum(e.get("skipped", 0) for e in entries),
            "error": sum(e.get("error", 0) for e in entries),
            "count": len(entries),
        }
        return {"period": "all", "buckets": [], "total": total}

    # bucket key
    def bucket_key(ts_str: str) -> str:
        try:
            ts = dt.datetime.fromisoformat(ts_str.replace("Z", "+00:00"))
        except (ValueError, TypeError):
            return "unknown"
        if period == "daily":
            return ts.strftime("%Y-%m-%d")
        elif period == "weekly":
            # ISO week (Monday-Sunday)
            iso = ts.isocalendar()
            return f"{iso.year}-W{iso.week:02d}"
        elif period == "monthly":
            return ts.strftime("%Y-%m")
        return "unknown"

    bucket_data: dict[str, list[dict]] = defaultdict(list)
    for e in entries:
        bk = bucket_key(e.get("timestamp", ""))
        bucket_data[bk].append(e)

    buckets = []
    for bk in sorted(bucket_data.keys()):
        es = bucket_data[bk]
        buckets.append({
            "bucket": bk,
            "archived": sum(e.get("archived", 0) for e in es),
            "skipped": sum(e.get("skipped", 0) for e in es),
            "error": sum(e.get("error", 0) for e in es),
            "count": len(es),
        })

    total = {
        "archived": sum(b["archived"] for b in buckets),
        "skipped": sum(b["skipped"] for b in buckets),
        "error": sum(b["error"] for b in buckets),
        "count": sum(b["count"] for b in buckets),
    }
    return {"period": period, "buckets": buckets, "total": total}


def cmd_aggregate_metrics(args: argparse.Namespace) -> dict:
    """metrics log 의 entries 를 *period* 별 aggregation.

    Args:
        args: argparse.Namespace with --aggregate (period), --include-rotated.

    Returns:
        dict with keys: ok (bool), period, buckets, total, entry_count, included_rotated (int).
    """
    repo_root = get_repo_root(args.repo_root)
    memory_dir = get_memory_dir(repo_root)
    if not memory_dir.exists():
        return {"ok": False, "error": f"memory dir not found: {memory_dir}"}
    period = getattr(args, "aggregate", "weekly")
    if period not in ("daily", "weekly", "monthly", "all"):
        return {"ok": False, "error": f"invalid --aggregate: {period} (expected: daily|weekly|monthly|all)"}
    entries = read_metrics_log(memory_dir)
    included_rotated = 0
    if getattr(args, "include_rotated", False):
        rotated = read_rotated_logs(memory_dir)
        # rotated 의 entries 는 source field 가 있음. main log 와 dedup
        # main log entries 의 timestamp 와 rotated 의 timestamp 가 *중복 가능* — 그러나 main 가 최신 → 우선
        main_ts = {e.get("timestamp") for e in entries}
        rotated_filtered = [e for e in rotated if e.get("timestamp") not in main_ts]
        entries.extend(rotated_filtered)
        included_rotated = len(rotated_filtered)
    result = aggregate_metrics(entries, period=period)
    return {
        "ok": True,
        "memory_dir": str(memory_dir),
        "entry_count": len(entries),
        "included_rotated": included_rotated,
        **result,
    }


# === Cron subcommand (TASK-V0728-001) ===
def cmd_install_cron(args: argparse.Namespace) -> dict:
    """`mavis cron create <agent> <cronName> --schedule <interval> --prompt ...` 자동 호출.

    매 interval 마다 본 tool 의 --older-than=<N> --apply 를 자동 실행.
    caller 는 prompt 의 *plain text* + 본 tool 의 호출.

    TASK-V0730-001: idempotency — *이미* 같은 cron 이 있으면 skip + report (ok=True).
    caller 가 *재호출* 시 *재설치* 안 함.

    Args:
        args: argparse.Namespace with --cron-name, --cron-interval, --older-than, --repo-root, --agent, --force-install.

    Returns:
        dict with keys: ok (bool), cron_name (str), cron_interval (str), mavis_cron_stdout (str),
        mavis_cron_stderr (str), returncode (int), idempotency (str: 'skipped-existing' | 'created' | 'forced'),
        error (str | None).
    """
    cron_name = getattr(args, "cron_name", "archive-memory")
    cron_interval = getattr(args, "cron_interval", "7d")
    agent = getattr(args, "agent", "mavis")
    older_than = args.older_than
    force_install = getattr(args, "force_install", False)
    repo_root = get_repo_root(args.repo_root)

    # TASK-V0730-001: idempotency check — mavis cron info <agent> <cronName> 으로 existing check
    if not force_install:
        info_proc = subprocess.run(
            ["mavis", "cron", "info", agent, cron_name],
            capture_output=True, text=True, timeout=30, cwd=str(get_repo_root(args.repo_root)),
        )
        if info_proc.returncode == 0 and cron_name in info_proc.stdout:
            # cron 이미 존재 → skip
            return {
                "ok": True,
                "cron_name": cron_name,
                "cron_interval": cron_interval,
                "older_than": older_than,
                "agent": agent,
                "repo_root": str(repo_root),
                "mavis_cron_stdout": info_proc.stdout,
                "mavis_cron_stderr": info_proc.stderr,
                "returncode": 0,
                "idempotency": "skipped-existing",
                "error": None,
            }

    prompt = (
        f"Run: python3 workflow-source/tools/archive_stale_memory.py --older-than={older_than} --apply "
        f"--repo-root={repo_root}\n"
        f"(auto-triggered by mavis cron '{cron_name}' every {cron_interval})"
    )

    proc = subprocess.run(
        ["mavis", "cron", "create", agent, cron_name, f"--schedule={cron_interval}", f"--prompt={prompt}"],
        capture_output=True, text=True, timeout=30, cwd=str(get_repo_root(args.repo_root)),
    )
    return {
        "ok": proc.returncode == 0,
        "cron_name": cron_name,
        "cron_interval": cron_interval,
        "older_than": older_than,
        "agent": agent,
        "repo_root": str(repo_root),
        "mavis_cron_stdout": proc.stdout,
        "mavis_cron_stderr": proc.stderr,
        "returncode": proc.returncode,
        "idempotency": "forced" if force_install else "created",
        "error": None if proc.returncode == 0 else f"mavis cron create failed (returncode={proc.returncode}): {proc.stderr}",
    }


def cmd_uninstall_cron(args: argparse.Namespace) -> dict:
    """`mavis cron disable <agent> <cronName>` 자동 호출.

    Args:
        args: argparse.Namespace with --cron-name, --agent.

    Returns:
        dict with keys: ok (bool), cron_name (str), mavis_cron_stdout (str),
        mavis_cron_stderr (str), returncode (int), error (str | None).
    """
    cron_name = getattr(args, "cron_name", "archive-memory")
    agent = getattr(args, "agent", "mavis")

    proc = subprocess.run(
        ["mavis", "cron", "disable", agent, cron_name],
        capture_output=True, text=True, timeout=30, cwd=str(get_repo_root(args.repo_root)),
    )
    return {
        "ok": proc.returncode == 0,
        "cron_name": cron_name,
        "agent": agent,
        "mavis_cron_stdout": proc.stdout,
        "mavis_cron_stderr": proc.stderr,
        "returncode": proc.returncode,
        "error": None if proc.returncode == 0 else f"mavis cron disable failed (returncode={proc.returncode}): {proc.stderr}",
    }


def cmd_show_cron(args: argparse.Namespace) -> dict:
    """`mavis cron list <agent>` 의 output 에서 cron_name 매치 여부 확인.

    Args:
        args: argparse.Namespace with --cron-name, --agent.

    Returns:
        dict with keys: ok (bool), cron_name (str), mavis_cron_stdout (str), found (bool).
    """
    cron_name = getattr(args, "cron_name", "archive-memory")
    agent = getattr(args, "agent", "mavis")
    proc = subprocess.run(
        ["mavis", "cron", "list", agent],
        capture_output=True, text=True, timeout=30, cwd=str(get_repo_root(args.repo_root)),
    )
    found = cron_name in proc.stdout if proc.returncode == 0 else False
    return {
        "ok": proc.returncode == 0,
        "cron_name": cron_name,
        "agent": agent,
        "found": found,
        "mavis_cron_stdout": proc.stdout,
        "mavis_cron_stderr": proc.stderr,
        "returncode": proc.returncode,
        "error": None if proc.returncode == 0 else f"mavis cron list failed (returncode={proc.returncode}): {proc.stderr}",
    }


# === argparse ===
def build_argparser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="archive_stale_memory",
        description="TASK-V0726-004 (v0.7.28): detached HEAD memory dir age-based auto-archive (short SHA dir). "
                    "TASK-V0728-001 (v0.7.30): --install-cron/--uninstall-cron/--show-cron 으로 자동화.",
    )
    p.add_argument(
        "--older-than", type=int, default=30,
        help="N day 이전의 short SHA dir 를 archive (default: 30)",
    )
    p.add_argument(
        "--list", action="store_true",
        help="모든 short SHA dir list (--older-than 무시, dry-run 만).",
    )
    p.add_argument(
        "--dry-run", action="store_true",
        help="dry-run mode: print what will be done, no file move.",
    )
    p.add_argument(
        "--apply", action="store_true",
        help="apply mode: actually move dirs to archive/.",
    )
    p.add_argument(
        "--cleanup", action="store_true",
        help="--apply alias (semantic clarity).",
    )
    p.add_argument(
        "--repo-root", type=str, default=None,
        help="REPO_ROOT override (priority 1, v0.7.12 정공법).",
    )
    # TASK-V0728-001 cron subcommand flags
    p.add_argument(
        "--install-cron", action="store_true",
        help="mavis cron self 자동 호출 (TASK-V0728-001).",
    )
    p.add_argument(
        "--uninstall-cron", action="store_true",
        help="mavis cron self --disable 자동 호출.",
    )
    p.add_argument(
        "--show-cron", action="store_true",
        help="mavis cron self list 에서 cron_name 매치 여부 확인.",
    )
    p.add_argument(
        "--show-metrics", action="store_true",
        help="TASK-V0729-001: archive_stale_memory.log 의 모든 entry read + return.",
    )
    p.add_argument(
        "--cron-name", type=str, default="archive-memory",
        help="mavis cron self 의 cron name (default: archive-memory).",
    )
    p.add_argument(
        "--cron-interval", type=str, default="7d",
        help="mavis cron self 의 interval (default: 7d, e.g. '7d' / '1d' / '12h').",
    )
    p.add_argument(
        "--agent", type=str, default="mavis",
        help="mavis cron 의 agent name (default: mavis).",
    )
    p.add_argument(
        "--force-install", action="store_true",
        help="TASK-V0730-001: --install-cron 의 idempotency skip 우회 (force create even if existing).",
    )
    # TASK-V0731-001: log rotation
    p.add_argument(
        "--rotate-logs", action="store_true",
        help="metrics log 의 line > threshold 시 gzip + rotate (TASK-V0731-001).",
    )
    p.add_argument(
        "--rotate-threshold", type=int, default=10000,
        help="TASK-V0731-001: rotate trigger (default: 10000 line).",
    )
    # TASK-V0732-001: metrics aggregation
    p.add_argument(
        "--aggregate", type=str, default=None,
        help="TASK-V0732-001: metrics aggregation period (daily|weekly|monthly|all).",
    )
    p.add_argument(
        "--include-rotated", action="store_true",
        help="TASK-V0732-001: --aggregate 시 rotated log (*.gz) 도 include.",
    )
    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_argparser()
    args = parser.parse_args(argv)

    # Cron subcommand 우선
    if args.install_cron:
        result = cmd_install_cron(args)
    elif args.uninstall_cron:
        result = cmd_uninstall_cron(args)
    elif args.show_cron:
        result = cmd_show_cron(args)
    elif args.show_metrics:
        result = cmd_show_metrics(args)
    elif args.rotate_logs:
        result = cmd_rotate_logs(args)
    elif args.aggregate is not None:
        result = cmd_aggregate_metrics(args)
    elif args.list:
        if args.apply or args.cleanup:
            parser.error("--list 와 --apply/--cleanup 은 mutually exclusive (--list 는 dry-run 만)")
        # force dry-run
        args.dry_run = True
        result = cmd_list(args)
    else:
        if not args.dry_run and not args.apply and not args.cleanup:
            parser.error("at least one of --dry-run, --apply, --cleanup, --install-cron, --uninstall-cron, --show-cron, --show-metrics, --rotate-logs, --aggregate is required")
        if args.cleanup:
            args.apply = True
        result = cmd_archive(args)

    print(json.dumps(result, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
