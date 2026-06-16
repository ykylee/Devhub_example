#!/usr/bin/env python3
"""
v0.7.25: Legacy L2 Page In-Repo Migration (F-6 정공법)

외부 wiki (`~/wiki/wiki/projects/standard-ai-workflow/sources/`) 의 legacy
L2 page (15 version: v0.1.0 ~ v0.6.3, 2026-06-13 이전의 releases-*.md file)
를 in-repo 의 `ai-workflow/memory/release/_external-wiki-legacy.md` 단일
mirror file 로 copy. in-repo 의 L2 SSOT 가 외부 wiki 와 *정합*.

Usage:
    # 1. dry-run: 어떤 file 이 어떻게 처리될지 미리 보기
    python3 tools/migrate_legacy_l2.py --dry-run

    # 2. apply: 실제 migrate
    python3 tools/migrate_legacy_l2.py --apply

    # 3. JSON output
    python3 tools/migrate_legacy_l2.py --dry-run --json

    # 4. REPO_ROOT override
    python3 tools/migrate_legacy_l2.py --repo-root=/path/to/repo --dry-run

v0.7.17 의 in-repo redirect 의 *closure* — 외부 wiki L2 page 의 *역사
보존* + in-repo 의 L2 SSOT 정합.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
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
            f"[DEPRECATION] migrate_legacy_l2.py: REPO_ROOT auto-detect 실패 — legacy fallback 사용 ({_LEGACY_REPO_ROOT}). "
            "v0.7.12+ 부터 --repo-root=<path> 또는 STANDARD_AI_WF_REPO env var 사용 권장.",
            file=sys.stderr,
        )
    return _LEGACY_REPO_ROOT


# === External wiki path resolution (legacy location) ===
def get_external_wiki_path() -> Path:
    """외부 wiki 의 standard-ai-workflow project sources/ path 결정.

    v0.7.17 의 in-repo redirect 이전의 legacy 위치. v0.7.17+ 의 *신규* L2
    page 는 in-repo `ai-workflow/memory/` 에 생성되지만, *legacy* L2 page
    (= 2026-06-13 이전의 releases-*.md file) 는 외부 wiki 에만 존재.

    Returns:
        Path. legacy location.
    """
    return Path.home() / "wiki" / "wiki" / "projects" / "standard-ai-workflow" / "sources"


# === Migration logic ===
RELEASE_FILE_RE = re.compile(r"^releases-(?:alpha|beta)-(v[\d.]+)\.md$")
VERSION_RE = re.compile(r"^v[\d.]+$")


def discover_legacy_release_files(external_wiki: Path) -> list[dict]:
    """외부 wiki 의 legacy L2 release page (releases-alpha-v*.md / releases-beta-v*.md) list.

    Returns:
        list of {"version": str, "external_path": Path, "file_size": int, "sha256": str}.
    """
    if not external_wiki.exists():
        return []

    results: list[dict] = []
    for path in sorted(external_wiki.iterdir()):
        if not path.is_file() or path.suffix != ".md":
            continue
        m = RELEASE_FILE_RE.match(path.name)
        if not m:
            continue
        version = m.group(1)
        content = path.read_bytes()
        sha = hashlib.sha256(content).hexdigest()
        results.append(
            {
                "version": version,
                "external_path": path,
                "file_size": len(content),
                "sha256": sha,
            }
        )
    return results


def is_legacy_version(version: str, inrepo_releases: set[str]) -> bool:
    """version 이 *legacy* (= external 에만 존재, in-repo 부재) 인지 확인.

    Args:
        version: e.g. "v0.6.3"
        inrepo_releases: in-repo 의 ai-workflow/memory/release/ 하위 dir 이름 set.

    Returns:
        True if external only (legacy), False if in-repo 또는 양쪽 모두.
    """
    return version not in inrepo_releases


def build_mirror_frontmatter(legacy_files: list[dict], migrated_at: str, commit: str) -> str:
    """Mirror file 의 frontmatter 생성.

    Args:
        legacy_files: discover_legacy_release_files() 의 result.
        migrated_at: ISO 8601 timestamp.
        commit: apply 시점의 HEAD commit hash (7자 또는 full).

    Returns:
        frontmatter + 1st heading line 의 string.
    """
    versions = sorted([f["version"] for f in legacy_files])
    lines = [
        "---",
        "title: External Wiki Legacy L2 Pages (v0.1.0 ~ v0.6.3)",
        "type: source",
        "tags: [external-wiki, legacy, l2-mirror]",
        "sources:",
    ]
    for f in legacy_files:
        # external path 를 repo-relative 로 표시 (best effort)
        ext_str = str(f["external_path"])
        lines.append(f"  - {ext_str}")
    lines.extend(
        [
            f"migrated_from: {get_external_wiki_path()}",
            f"migrated_at: {migrated_at}",
            f"commit: {commit}",
            f"versions: [{', '.join(versions)}]",
            f"version_count: {len(versions)}",
            "last_touched: 2026-06-15",
            "related: []",
            "status: reviewed",
            "contradictions: []",
            "---",
            "",
        ]
    )
    return "\n".join(lines)


def build_mirror_body(legacy_files: list[dict]) -> str:
    """Mirror file 의 body 생성 — 15 file 의 1:1 mirror + 목차.

    Args:
        legacy_files: discover_legacy_release_files() 의 result.

    Returns:
        body string (frontmatter 제외).
    """
    sections: list[str] = []
    # Table of contents
    sections.append("# External Wiki Legacy L2 Pages (v0.7.25 F-6 migration)")
    sections.append("")
    sections.append("외부 wiki (`~/wiki/wiki/projects/standard-ai-workflow/sources/`) 의")
    sections.append("legacy L2 release page (15 version: v0.1.0 ~ v0.6.3) 의 1:1 mirror.")
    sections.append("v0.7.17 의 in-repo redirect 의 *closure* — 외부 wiki L2 page 의")
    sections.append("*역사 보존* + in-repo 의 L2 SSOT 정합.")
    sections.append("")
    sections.append("## 목차 (Table of Contents)")
    sections.append("")
    for f in legacy_files:
        sections.append(f"- [{f['version']}](#{f['version'].replace('.', '')}) — {f['file_size']:,} bytes, sha256:`{f['sha256'][:16]}...`")
    sections.append("")
    sections.append("---")
    sections.append("")

    # 1:1 mirror
    for f in legacy_files:
        sections.append(f"## {f['version']}")
        sections.append("")
        sections.append(f"**Source**: `{f['external_path']}`")
        sections.append(f"**Size**: {f['file_size']:,} bytes")
        sections.append(f"**SHA256**: `{f['sha256']}`")
        sections.append("")
        sections.append("```markdown")
        sections.append(f["external_path"].read_text(encoding="utf-8"))
        sections.append("```")
        sections.append("")
        sections.append("---")
        sections.append("")

    return "\n".join(sections)


def detect_drift(inrepo_mirror: Path, new_content: str) -> dict:
    """기존 in-repo mirror 와 new content 비교 (idempotency + hash drift).

    Args:
        inrepo_mirror: in-repo 의 mirror file path.
        new_content: 새로 생성하려는 content (full, frontmatter + body).

    Returns:
        dict:
            - status: "fresh" (신규) | "identical" (idempotent skip) | "drift" (hash 다름)
            - existing_sha256: 기존 file 의 SHA256 (없으면 None)
            - new_sha256: new content 의 SHA256
            - drift_size: byte 크기 차이 (drift 시에만)
    """
    new_sha = hashlib.sha256(new_content.encode("utf-8")).hexdigest()
    if not inrepo_mirror.exists():
        return {
            "status": "fresh",
            "existing_sha256": None,
            "new_sha256": new_sha,
            "drift_size": None,
        }
    existing_content = inrepo_mirror.read_text(encoding="utf-8")
    existing_sha = hashlib.sha256(existing_content.encode("utf-8")).hexdigest()
    if existing_sha == new_sha:
        return {
            "status": "identical",
            "existing_sha256": existing_sha,
            "new_sha256": new_sha,
            "drift_size": 0,
        }
    return {
        "status": "drift",
        "existing_sha256": existing_sha,
        "new_sha256": new_sha,
        "drift_size": len(new_content) - len(existing_content),
    }


def cmd_migrate_legacy_l2(args: argparse.Namespace) -> dict:
    """Main subcommand: legacy L2 page migrate (dry-run or apply).

    Returns:
        dict with keys: mode, external_wiki, inrepo_mirror, legacy_count,
        inrepo_release_count, files (list), drift, action_performed.
    """
    repo_root = get_repo_root(args.repo_root)
    inrepo_release_dir = repo_root / "ai-workflow" / "memory" / "release"
    inrepo_mirror = inrepo_release_dir / "_external-wiki-legacy.md"
    external_wiki = get_external_wiki_path()

    # in-repo 의 release dir set
    inrepo_releases: set[str] = set()
    if inrepo_release_dir.exists():
        inrepo_releases = {p.name for p in inrepo_release_dir.iterdir() if p.is_dir()}

    # external legacy file 발견
    all_files = discover_legacy_release_files(external_wiki)
    legacy_files = [f for f in all_files if is_legacy_version(f["version"], inrepo_releases)]

    # build content
    migrated_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    # commit 은 apply 시점의 HEAD (v0.7.25 의 in-flight commit hash)
    # dry-run 시점에는 HEAD 그대로
    try:
        proc = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            capture_output=True, text=True, timeout=5,
            cwd=repo_root,
        )
        commit = proc.stdout.strip()[:7] if proc.returncode == 0 else "TBD"
    except (subprocess.TimeoutExpired, FileNotFoundError):
        commit = "TBD"

    frontmatter = build_mirror_frontmatter(legacy_files, migrated_at, commit)
    body = build_mirror_body(legacy_files)
    new_content = frontmatter + body

    # drift detection
    drift = detect_drift(inrepo_mirror, new_content)

    # action
    action_performed = "none"
    if args.dry_run:
        action_performed = "dry-run"
    else:
        if drift["status"] == "identical":
            action_performed = "skipped (identical)"
        elif drift["status"] == "fresh":
            inrepo_mirror.parent.mkdir(parents=True, exist_ok=True)
            inrepo_mirror.write_text(new_content, encoding="utf-8")
            action_performed = "written (fresh)"
        else:  # drift
            # hash mismatch → warning + skip (manual review 필요)
            # v0.7.18 의 destructive subcommand 정공법 (--dry-run 필수 + apply 시 graceful fail)
            # 는 destructive 작업에만 적용. 본 tool 은 *file write* (덮어쓰기) → drift 시
            # *manual review* 가 더 안전. skip.
            action_performed = "skipped (drift — manual review)"

    result = {
        "mode": "dry-run" if args.dry_run else "apply",
        "external_wiki": str(external_wiki),
        "inrepo_mirror": str(inrepo_mirror),
        "legacy_count": len(legacy_files),
        "inrepo_release_count": len(inrepo_releases),
        "files": [
            {
                "version": f["version"],
                "external_path": str(f["external_path"]),
                "file_size": f["file_size"],
                "sha256": f["sha256"],
            }
            for f in legacy_files
        ],
        "drift": drift,
        "action_performed": action_performed,
        "migrated_at": migrated_at,
        "commit": commit,
    }
    return result


# === argparse ===
def build_argparser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="migrate_legacy_l2",
        description="F-6: 외부 wiki legacy L2 page → in-repo mirror migration (v0.7.25)",
    )
    p.add_argument(
        "--dry-run",
        action="store_true",
        help="dry-run mode: print what will be done, no file write.",
    )
    p.add_argument(
        "--apply",
        action="store_true",
        help="apply mode: actually write the mirror file.",
    )
    p.add_argument(
        "--repo-root",
        type=str,
        default=None,
        help="REPO_ROOT override (priority 1, v0.7.12 정공법).",
    )
    p.add_argument(
        "--json",
        action="store_true",
        help="output as JSON.",
    )
    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_argparser()
    args = parser.parse_args(argv)

    if not args.dry_run and not args.apply:
        parser.error("at least one of --dry-run or --apply is required")

    result = cmd_migrate_legacy_l2(args)

    if args.json:
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print(f"mode: {result['mode']}")
        print(f"external_wiki: {result['external_wiki']}")
        print(f"inrepo_mirror: {result['inrepo_mirror']}")
        print(f"legacy_count: {result['legacy_count']}")
        print(f"inrepo_release_count: {result['inrepo_release_count']}")
        print(f"action: {result['action_performed']}")
        print(f"drift: {result['drift']['status']}")
        print()
        print("--- legacy files ---")
        for f in result["files"]:
            print(f"  {f['version']:12s}  {f['file_size']:>8,} bytes  sha256:{f['sha256'][:16]}...")

    return 0


if __name__ == "__main__":
    sys.exit(main())
