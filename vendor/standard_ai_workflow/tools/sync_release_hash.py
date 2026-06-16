#!/usr/bin/env python3
"""
v0.7.26: Automated fix(state) tool (F-7+)

v0.7.25 의 infinite fix(state) loop 회피 (8 commit + squash + final fix).
chore commit 시 자동 호출 가능 + 1 commit 으로 state.json + backlog 의
hash = chore commit hash.

Subcommand:
    sync-release-hash: state.json + backlog/ 의 hash = latest commit (chore)
"""
from __future__ import annotations

import argparse
import json
import re
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
    import os

    if cli_value is not None:
        return Path(cli_value).expanduser().resolve()

    env_val = os.environ.get("STANDARD_AI_WF_REPO")
    if env_val:
        return Path(env_val).expanduser().resolve()

    try:
        proc = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, timeout=5,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return Path(proc.stdout.strip()).resolve()
    except (subprocess.TimeoutExpired, FileNotFoundError):
        pass

    if not _DEPRECATION_WARNED and not _suppress_warning:
        _DEPRECATION_WARNED = True
        print(
            f"[DEPRECATION] sync_release_hash.py: REPO_ROOT auto-detect 실패 — legacy fallback 사용 ({_LEGACY_REPO_ROOT}).",
            file=sys.stderr,
        )
    return _LEGACY_REPO_ROOT


def get_latest_commit_hash(repo_root: Path, *, since: str | None = None) -> str:
    """latest commit (or since~..HEAD) 의 short SHA (7자) 반환.

    Args:
        repo_root: REPO_ROOT path.
        since: 기준 commit hash. None 이면 HEAD 자체.

    Returns:
        7자 short SHA. e.g. "8a61bd3".
    """
    if since:
        # rev-list {since}..HEAD + short hash
        proc_rev = subprocess.run(
            ["git", "rev-list", f"{since}..HEAD"],
            capture_output=True, text=True, timeout=5, cwd=str(repo_root),
        )
        if proc_rev.returncode != 0 or not proc_rev.stdout.strip():
            raise RuntimeError(f"git rev-list failed: {proc_rev.stderr}")
        head_sha = proc_rev.stdout.strip().splitlines()[-1]
        cmd = ["git", "rev-parse", "--short=7", head_sha]
    else:
        # 2-step: full SHA → short=7
        proc_full = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            capture_output=True, text=True, timeout=5, cwd=str(repo_root),
        )
        if proc_full.returncode != 0 or not proc_full.stdout.strip():
            raise RuntimeError(f"git rev-parse failed: {proc_full.stderr}")
        head_sha = proc_full.stdout.strip()
        cmd = ["git", "rev-parse", "--short=7", head_sha]

    proc = subprocess.run(
        cmd, capture_output=True, text=True, timeout=5, cwd=str(repo_root),
    )
    if proc.returncode != 0:
        raise RuntimeError(f"git rev-parse --short failed: {proc.stderr}")
    sha = proc.stdout.strip()
    if not sha:
        raise RuntimeError(f"empty git rev-parse --short output: {proc.stdout!r}")
    return sha[:7]


def update_state_json(state_path: Path, version: str, new_hash: str, *, dry_run: bool = False) -> bool:
    """state.json 의 recent_done_items[0] 에서 `{version} ({old_hash_or_TBD})` → `{version} ({new_hash})`.

    `{version} (TBD):` 같은 placeholder 도 매치 (F-7+ fix: TBD 도 sync 대상).

    Returns:
        True if changed, False if no match.
    """
    if not state_path.exists():
        return False
    content = state_path.read_text(encoding="utf-8")
    # 7-char hex hash OR TBD placeholder
    pattern = re.compile(rf'"{re.escape(version)} \(([a-f0-9]{{7}}|TBD)\):')
    matches = pattern.findall(content)
    if not matches:
        return False
    new_content = pattern.sub(f'"{version} ({new_hash}):', content, count=1)
    if new_content == content:
        return False
    if not dry_run:
        state_path.write_text(new_content, encoding="utf-8")
    return True


def update_backlog(backlog_path: Path, new_hash: str, *, dry_run: bool = False) -> bool:
    """backlog 의 `**commit**: \\`old_hash_or_TBD\\`` → `**commit**: \\`new_hash\\``.

    TBD placeholder 도 매치 (F-7+ fix).

    Returns:
        True if changed, False if no match.
    """
    if not backlog_path.exists():
        return False
    content = backlog_path.read_text(encoding="utf-8")
    # 7-char hex hash OR TBD placeholder
    pattern = re.compile(r'\*\*commit\*\*: `([a-f0-9]{7}|TBD)`')
    matches = pattern.findall(content)
    if not matches:
        return False
    new_content = pattern.sub(f'**commit**: `{new_hash}`', content, count=1)
    if new_content == content:
        return False
    if not dry_run:
        backlog_path.write_text(new_content, encoding="utf-8")
    return True


def cmd_sync_release_hash(args: argparse.Namespace) -> dict:
    """Main subcommand: state.json + backlog 의 hash = latest commit 으로 sync.

    Returns:
        dict with keys: version, old_hash, new_hash, state_updated, backlog_updated, mode.
    """
    repo_root = get_repo_root(args.repo_root)
    version = args.version
    if not version:
        raise ValueError("--version is required (e.g. v0.7.26)")

    # 1. latest commit hash 결정
    new_hash = get_latest_commit_hash(repo_root)

    # 2. state.json + backlog path
    state_path = repo_root / "ai-workflow" / "memory" / "active" / "state.json"
    backlog_path = repo_root / "ai-workflow" / "memory" / "release" / version / "backlog" / "2026-06-15.md"
    if not backlog_path.exists():
        # 다른 날짜의 backlog 시도
        release_dir = repo_root / "ai-workflow" / "memory" / "release" / version
        if release_dir.exists():
            for f in release_dir.glob("backlog/*.md"):
                backlog_path = f
                break

    # 3. update
    state_updated = update_state_json(state_path, version, new_hash, dry_run=args.dry_run)
    backlog_updated = update_backlog(backlog_path, new_hash, dry_run=args.dry_run) if backlog_path.exists() else False

    return {
        "version": version,
        "new_hash": new_hash,
        "state_updated": state_updated,
        "backlog_updated": backlog_updated,
        "backlog_path": str(backlog_path) if backlog_path.exists() else None,
        "state_path": str(state_path),
        "mode": "dry-run" if args.dry_run else "apply",
    }


# === argparse ===
def build_argparser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="sync_release_hash",
        description="F-7+ (v0.7.26): state.json + backlog 의 hash 를 latest commit 으로 sync (infinite fix(state) loop 회피)",
    )
    p.add_argument(
        "--version",
        type=str,
        required=True,
        help="version (e.g. v0.7.26)",
    )
    p.add_argument(
        "--repo-root",
        type=str,
        default=None,
        help="REPO_ROOT override (priority 1, v0.7.12 정공법).",
    )
    p.add_argument(
        "--dry-run",
        action="store_true",
        help="dry-run mode: print what will be done, no file write.",
    )
    p.add_argument(
        "--apply",
        action="store_true",
        help="apply mode: actually update files.",
    )
    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_argparser()
    args = parser.parse_args(argv)

    if not args.dry_run and not args.apply:
        parser.error("at least one of --dry-run or --apply is required")

    result = cmd_sync_release_hash(args)

    if args.json if hasattr(args, "json") else False:
        import json
        print(json.dumps(result, indent=2))
    else:
        print(f"mode: {result['mode']}")
        print(f"version: {result['version']}")
        print(f"new_hash: {result['new_hash']}")
        print(f"state_updated: {result['state_updated']}")
        print(f"backlog_updated: {result['backlog_updated']}")
        if result.get("backlog_path"):
            print(f"backlog_path: {result['backlog_path']}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
