#!/usr/bin/env python3
"""v0.7.9+: standard-ai-workflow release pipeline 정식화 (7 subcommand).

release 절차 (validate → dist → version-bump → note-draft → release → verify → rollback) 의
*기계화 layer*. manual 절차 (memory #5 / docs/RELEASE.md) 의 *부족 부분* 자동화.

Phase 1 (v0.7.9): validate / version-bump / note-draft — 사전 점검 + version + note.
Phase 2 (v0.7.10): release / verify / rollback — gh CLI 통합 + read-only verify + destructive rollback.
Phase 3 (v0.7.11): dist — `python3 -m build` wheel + sdist 자동 빌드 (PEP 517/518).
Phase 5 (v0.7.18): release coordination observability — cmd_release 의 --auto-bump
  + remote tag pre-check (`git ls-remote origin`). v0.7.16 의 race lesson 반영.

PyPI/TestPyPI 업로드 ❌ (memory #5 의 release 채널 정책 — GitHub Releases 만).

Usage:
    # dry-run: 모든 subcommand plan 만 출력
    python3 tools/release_pipeline.py validate --dry-run
    python3 tools/release_pipeline.py version-bump --patch --dry-run
    python3 tools/release_pipeline.py note-draft --from=v0.7.8 --to=v0.7.9 --dry-run

    # apply
    python3 tools/release_pipeline.py version-bump --patch --apply
    python3 tools/release_pipeline.py note-draft --from=v0.7.8 --to=v0.7.9 --apply

    # JSON 출력 (CI integration)
    python3 tools/release_pipeline.py validate --json

Reference:
- tools/check_packaging.py (packaging 정합성 검증)
- tools/refresh_wiki_memory.py (v0.7.5, git log → memory emit 패턴)
- workflow_kit.cli.doctor (v0.7.8 state-aware baseline 검증)
- memory #5 standard-ai-workflow.md (release 채널 정책: GitHub Releases 만)
- docs/RELEASE.md (수동 release 절차)
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path

# v0.7.15+ atomic_write (POSIX os.replace guarantee)
try:
    from workflow_kit.common.atomic_write import atomic_write_json, atomic_write_text
except ImportError:
    # standalone script (no workflow_kit on sys.path) — fall back to direct write.
    atomic_write_json = None  # type: ignore[assignment]
    atomic_write_text = None  # type: ignore[assignment]
# 1차 출처
REPO_ROOT = Path(__file__).resolve().parents[1]
PYPROJECT = REPO_ROOT / "pyproject.toml"
RELEASES_DIR = REPO_ROOT / "releases"
WORKFLOW_KIT_INIT = REPO_ROOT / "workflow_kit" / "__init__.py"

# 표준 sub-package list (check_packaging.py 와 동일 — packaging 정합성)
EXPECTED_SUBPACKAGES = [
    "workflow_kit",
    "workflow_kit.common",
    "workflow_kit.common.modes",
    "workflow_kit.common.state",
    "workflow_kit.common.contracts",
    "workflow_kit.common.schemas",
    "workflow_kit.contract_v1",
    "workflow_kit.server",
    "workflow_kit.harness",
    "bootstrap_lib",
    "bootstrap_lib.harnesses",
]

# tomllib (3.11+) / tomli (3.10) 분기
if sys.version_info >= (3, 11):
    import tomllib  # type: ignore[import-not-found]
else:  # pragma: no cover
    import tomli as tomllib  # type: ignore[no-redef, import-not-found]


# ---------------------------------------------------------------------------
# 1. validate
# ---------------------------------------------------------------------------


def cmd_validate(args) -> dict:
    """4 source 의 release-readiness 검증.

    1. check_packaging.py: pyproject 의 [tool.setuptools.packages] ↔ 디스크 정합
    2. workflow_kit.cli.doctor: 7 baseline 모두 evaluate (state-aware variant)
    3. state.json freshness: v0.7.5+ refresh_wiki_memory 의 last_freeze / last_ingest
    4. git status: working tree clean (release commit 의 clean state 보장)
    """
    results: dict = {}

    # 1. check_packaging
    if not args.skip_packaging:
        proc = subprocess.run(
            [sys.executable, str(REPO_ROOT / "tools/check_packaging.py")],
            capture_output=True, text=True, timeout=120,
        )
        results["packaging"] = {
            "exit_code": proc.returncode,
            "ok": proc.returncode == 0,
            "last_line": proc.stdout.strip().split("\n")[-1] if proc.stdout else "",
        }
    else:
        results["packaging"] = {"ok": True, "skipped": True}

    # 2. workflow_kit.cli.doctor (v0.7.8)
    if not args.skip_doctor:
        proc = subprocess.run(
            [sys.executable, "-m", "workflow_kit.cli.doctor", "--json",
             "--project-root", str(REPO_ROOT)],
            capture_output=True, text=True, timeout=60,
        )
        if proc.returncode == 0:
            try:
                doctor_out = json.loads(proc.stdout)
                cs_status = {
                    bl: doctor_out["results"][bl]["status"]
                    for bl in doctor_out["results"]
                }
                non_compliant = [
                    bl for bl, st in cs_status.items() if st == "non_compliant"
                ]
                results["doctor"] = {
                    "ok": len(non_compliant) == 0,
                    "baselines": cs_status,
                    "non_compliant": non_compliant,
                }
            except (json.JSONDecodeError, KeyError) as e:
                results["doctor"] = {"ok": False, "error": str(e)}
        else:
            results["doctor"] = {"ok": False, "exit_code": proc.returncode}
    else:
        results["doctor"] = {"ok": True, "skipped": True}

    # 3. state.json freshness
    if not args.skip_state:
        state_path = REPO_ROOT / "ai-workflow" / "memory" / "active" / "state.json"
        if state_path.exists():
            data = json.loads(state_path.read_text())
            last_freeze = data.get("memory", {}).get("last_freeze", "")
            last_ingest = data.get("wiki", {}).get("last_ingest", "")
            # last_freeze 가 v0.7.9 cycle 의 commit 범위 내면 OK
            results["state"] = {
                "ok": bool(last_freeze),
                "last_freeze": last_freeze,
                "last_ingest": last_ingest,
            }
        else:
            # state.json 부재도 OK (default empty state 시 v0.7.8 정합)
            results["state"] = {"ok": True, "absent": True}
    else:
        results["state"] = {"ok": True, "skipped": True}

    # 4. git status
    if not args.skip_git:
        proc = subprocess.run(
            ["git", "status", "--porcelain"],
            cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=10,
        )
        clean = proc.stdout.strip() == ""
        results["git"] = {
            "ok": clean,
            "untracked_or_modified": proc.stdout.strip().split("\n") if not clean else [],
        }
    else:
        results["git"] = {"ok": True, "skipped": True}

    return results


# ---------------------------------------------------------------------------
# 1.5 release coordination observability (v0.7.18+)
# ---------------------------------------------------------------------------


def _check_remote_tag(tag: str, *, timeout: int = 15) -> dict:
    """원격 (origin) 에 주어진 tag 가 존재하는지 확인.

    Returns:
        {"exists": bool, "remote_url": str | None, "tag": str}
    """
    result: dict = {"exists": False, "remote_url": None, "tag": tag}
    # 1. remote URL 추출
    remote_proc = subprocess.run(
        ["git", "remote", "get-url", "origin"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=timeout,
    )
    if remote_proc.returncode != 0:
        return result
    result["remote_url"] = remote_proc.stdout.strip()
    # 2. ls-remote 로 tag 조회
    ls_proc = subprocess.run(
        ["git", "ls-remote", "origin", f"refs/tags/{tag}"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=timeout,
    )
    if ls_proc.returncode == 0 and ls_proc.stdout.strip():
        result["exists"] = True
    return result


def _list_remote_tags(pattern: str = "v*", *, timeout: int = 15) -> list[str]:
    """원격의 tag list (정규식 filter, sort -V)."""
    ls_proc = subprocess.run(
        ["git", "ls-remote", "--tags", "origin", pattern],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=timeout,
    )
    if ls_proc.returncode != 0:
        return []
    tags = []
    for line in ls_proc.stdout.strip().splitlines():
        # line: "<sha>\trefs/tags/<tagname>"
        parts = line.split("\t", 1)
        if len(parts) == 2:
            tag = parts[1].removeprefix("refs/tags/")
            # peel 된 ^{} tag 제외
            if not tag.endswith("^{}"):
                tags.append(tag)
    return sorted(tags, key=_version_sort_key)


def _version_sort_key(tag: str) -> tuple:
    """PEP 440 + suffix sort key. v0.7.17-beta → (0, 7, 17, 'beta'), v0.7.18 → (0, 7, 18, '').

    SemVer-ish + PEP 440 suffix 순서 (release < alpha < beta < rc). 정수 tuple 이므로
    `sorted(tags, key=_version_sort_key)` 가 *자동으로* numeric + suffix 순서.
    """
    # 'v' prefix 제거
    s = tag.lstrip("v")
    # '-suffix' 분리
    if "-" in s:
        base, suffix = s.split("-", 1)
    else:
        base, suffix = s, ""
    # base = 'X.Y.Z' → int tuple
    parts = base.split(".")
    nums = tuple(int(p) for p in parts if p.isdigit())
    # suffix sort: '' (release) < 'alpha' < 'beta' < 'rc'
    suffix_order = {"": 0, "alpha": 1, "beta": 2, "rc": 3}
    suffix_rank = suffix_order.get(suffix.split(".")[0], 99)
    return nums + (suffix_rank, suffix)


def next_available_version(local_version: str, *, remote_tags: list[str] | None = None) -> dict:
    """local_version 보다 큰, remote 에 없는 다음 version 결정.

    1차 출처: remote `git ls-remote --tags origin "vX.Y.*"` 의 latest + 0.0.1 bump.
    local_version 이 이미 remote 의 latest 보다 크면 그대로 (충돌 없음).
    같은 major.minor prefix 의 모든 tag → max + 0.0.1.

    Args:
        local_version: 현재 local pyproject 의 version (e.g. "0.7.17").
        remote_tags: pre-fetched list. None 이면 _list_remote_tags() 호출.

    Returns:
        {"next": "0.7.18", "current_local": "0.7.17", "remote_max": "0.7.17-beta", "bumped": True}
    """
    if remote_tags is None:
        remote_tags = _list_remote_tags()
    # local_version 의 major.minor prefix
    parts = local_version.split(".")
    if len(parts) < 2:
        major_minor_prefix = local_version
    else:
        major_minor_prefix = ".".join(parts[:2])
    # remote 의 같은 major.minor 의 tag 만 filter
    prefix = f"v{major_minor_prefix}."
    same_prefix = [t for t in remote_tags if t.startswith(prefix)]
    # numeric base 비교 (PEP 440 suffix 무시)
    def base_tuple(t: str) -> tuple:
        b = t.lstrip("v").split("-", 1)[0]
        try:
            return tuple(int(p) for p in b.split("."))
        except ValueError:
            return (0,)
    if same_prefix:
        remote_max = max(same_prefix, key=base_tuple)
    else:
        remote_max = None
    local_tuple = base_tuple(f"v{local_version}")
    if remote_max is None:
        # remote 에 같은 major.minor 부재 → local 그대로 (다음 patch 가 local 의 +1)
        next_v = local_version
        bumped = False
    else:
        remote_tuple = base_tuple(remote_max)
        if local_tuple > remote_tuple:
            # local 이 remote max 보다 큼 → 그대로
            next_v = local_version
            bumped = False
        elif local_tuple < remote_tuple:
            # local 이 remote max 보다 작음 → remote max + 0.0.1
            next_tuple = list(remote_tuple)
            next_tuple[-1] += 1
            next_v = ".".join(str(n) for n in next_tuple)
            bumped = True
        else:
            # local == remote max → patch bump
            next_tuple = list(remote_tuple)
            next_tuple[-1] += 1
            next_v = ".".join(str(n) for n in next_tuple)
            bumped = True
    return {
        "next": next_v,
        "current_local": local_version,
        "remote_max": remote_max,
        "bumped": bumped,
    }


# ---------------------------------------------------------------------------
# 2. version-bump
# ---------------------------------------------------------------------------


def read_version() -> str:
    """pyproject.toml 의 [project] version 읽기."""
    with PYPROJECT.open("rb") as f:
        data = tomllib.load(f)
    return data["project"]["version"]


def write_version(new_version: str) -> None:
    """pyproject.toml 의 [project] version 갱신."""
    text = PYPROJECT.read_text()
    text = re.sub(
        r'(version\s*=\s*)"[\d.]+([^"]*)"',
        rf'\1"{new_version}\2"',
        text,
        count=1,
    )
    PYPROJECT.write_text(text)


def read_workflow_kit_version() -> str:
    """workflow_kit/__init__.py 의 __version__ 읽기. e.g. 'v0.7.13-beta'."""
    text = WORKFLOW_KIT_INIT.read_text()
    m = re.search(r'__version__\s*=\s*"([^"]+)"', text)
    if not m:
        raise ValueError(f"__version__ not found in {WORKFLOW_KIT_INIT}")
    return m.group(1)


def write_workflow_kit_version(new_version: str, *, suffix: str = "-beta") -> str:
    """workflow_kit/__init__.py 의 __version__ 갱신.

    e.g. new_version='0.7.14' → '__version__ = "v0.7.14-beta"'.
    suffix 인자 (default '-beta') 로 suffix override 가능. None 시 suffix 제거.
    Returns:
        실제 기록된 __version__ string (e.g. 'v0.7.14-beta').
    """
    text = WORKFLOW_KIT_INIT.read_text()
    target = f'"v{new_version}{suffix or ""}"'
    new_text, n = re.subn(
        r'__version__\s*=\s*"v?\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?"',
        f'__version__ = {target}',
        text,
        count=1,
    )
    if n == 0:
        raise ValueError(f"__version__ pattern not found in {WORKFLOW_KIT_INIT}")
    WORKFLOW_KIT_INIT.write_text(new_text)
    return f"v{new_version}{suffix or ''}"

def bump_version(version: str, *, patch: bool = False, minor: bool = False, major: bool = False, to: str | None = None) -> str:
    """version bump.

    --to=... 명시 시 그대로 사용. 아니면 --major/--minor/--patch 중 1개 (default: --patch).
    """
    if to is not None:
        return to
    major_n, minor_n, patch_n = parse_version(version)
    if major:
        return f"{major_n + 1}.0.0"
    if minor:
        return f"{major_n}.{minor_n + 1}.0"
    # default: patch
    return f"{major_n}.{minor_n}.{patch_n + 1}"


def parse_version(version: str) -> tuple[int, int, int]:
    """'0.7.8' → (0, 7, 8). pre-release 식별자 (e.g. '-beta') 는 무시."""
    m = re.match(r"^(\d+)\.(\d+)\.(\d+)", version)
    if not m:
        raise ValueError(f"invalid version: {version}")
    return int(m.group(1)), int(m.group(2)), int(m.group(3))


def cmd_version_bump(args) -> dict:
    """pyproject.toml version patch + workflow_kit/__init__.py __version__ 자동 sync (v0.7.14+).

    --no-init flag 시 __init__.py sync skip (CI / override 시나리오).

    v0.7.27+: --apply 시 sync_release_hash.py 자동 호출 (TASK-V0726-003). 본 release 의
    state.json + backlog 의 hash = latest commit (apply 후의 chore commit) 으로 1 commit
    으로 정합. infinite fix(state) loop 회피.
    """
    current = read_version()
    current_wk = read_workflow_kit_version()
    if args.dry_run:
        new = bump_version(
            current,
            patch=args.patch, minor=args.minor, major=args.major, to=args.to,
        )
        result = {
            "mode": "dry-run",
            "current_pyproject": current,
            "current_workflow_kit": current_wk,
            "next_pyproject": new,
            "next_workflow_kit": f"v{new}-beta" if not getattr(args, "no_init", False) else "(skipped)",
        }
        return result
    if args.to is None and not (args.patch or args.minor or args.major):
        args.patch = True
    new = bump_version(
        current,
        patch=args.patch, minor=args.minor, major=args.major, to=args.to,
    )
    write_version(new)
    result = {
        "mode": "applied",
        "previous_pyproject": current,
        "current_pyproject": new,
    }
    if not getattr(args, "no_init", False):
        written = write_workflow_kit_version(new, suffix="-beta")
        result["previous_workflow_kit"] = current_wk
        result["current_workflow_kit"] = written
    else:
        result["workflow_kit_skipped"] = True

    # TASK-V0726-003 (v0.7.27): post-step 자동 sync — state.json + backlog 의 hash = latest
    # commit. --skip-sync-hash flag 시 skip (manual override).
    if not getattr(args, "skip_sync_hash", False):
        sync_result = _run_post_step_sync_hash(new)
        result["sync_hash_result"] = sync_result
    return result


def _run_post_step_sync_hash(version: str) -> dict:
    """sync_release_hash.py 자동 호출 (TASK-V0726-003 post-step) + amend 통합 (TASK-V0727-001).

    2-phase:
    1. sync_release_hash.py 자동 호출 — state.json + backlog 의 TBD → *current HEAD* hash
    2. `git add` (sync 의 변경) + `git commit --amend --no-edit` — 1 commit 통합 (별도 fix(state) commit 불필요)

    sync_release_hash.py 는 release_pipeline.py 와 같은 dir (workflow-source/tools/) 에
    위치. REPO_ROOT 와 무관하게 __file__ 의 parents[1] (workflow-source/tools/) 기준.

    Args:
        version: new version (e.g. "0.7.29").

    Returns:
        dict with keys: ok (bool), sync_result (subprocess result), amend_result (subprocess result),
        final_hash (amend 후의 HEAD short SHA, 또는 None).
        sync_release_hash.py 또는 git amend 의 returncode != 0 면 ok = False.
    """
    sync_tool = Path(__file__).resolve().parent / "sync_release_hash.py"
    if not sync_tool.exists():
        return {
            "ok": False, "sync_result": None, "amend_result": None, "final_hash": None,
            "error": f"sync_release_hash.py not found: {sync_tool}",
        }
    version_arg = f"v{version}" if not version.startswith("v") else version

    # Phase 1: sync_release_hash 호출
    proc_sync = subprocess.run(
        [sys.executable, str(sync_tool), f"--version={version_arg}", "--apply"],
        capture_output=True, text=True, timeout=30, cwd=str(REPO_ROOT),
    )
    sync_result = {
        "stdout": proc_sync.stdout,
        "stderr": proc_sync.stderr,
        "returncode": proc_sync.returncode,
    }
    if proc_sync.returncode != 0:
        return {
            "ok": False, "sync_result": sync_result, "amend_result": None, "final_hash": None,
            "error": f"sync_release_hash.py failed (returncode={proc_sync.returncode}): {proc_sync.stderr}",
        }

    # Phase 2: git add (sync 의 변경) + git commit --amend --no-edit (1 commit 통합)
    # amend 시 *HEAD* 의 *직전* commit (feat or chore) 이 amend 됨
    # sync_release_hash 의 변경 = state.json + backlog 의 TBD → HEAD hash
    # *이미* amend 후 의 *HEAD* 의 본 release 의 chore commit hash 와 정합
    proc_add = subprocess.run(
        ["git", "add", "-A"],
        capture_output=True, text=True, timeout=30, cwd=str(REPO_ROOT),
    )
    add_result = {
        "stdout": proc_add.stdout,
        "stderr": proc_add.stderr,
        "returncode": proc_add.returncode,
    }
    if proc_add.returncode != 0:
        return {
            "ok": False, "sync_result": sync_result, "amend_result": add_result, "final_hash": None,
            "error": f"git add failed (returncode={proc_add.returncode}): {proc_add.stderr}",
        }

    proc_amend = subprocess.run(
        ["git", "commit", "--amend", "--no-edit"],
        capture_output=True, text=True, timeout=30, cwd=str(REPO_ROOT),
    )
    amend_result = {
        "stdout": proc_amend.stdout,
        "stderr": proc_amend.stderr,
        "returncode": proc_amend.returncode,
    }
    if proc_amend.returncode != 0:
        return {
            "ok": False, "sync_result": sync_result, "amend_result": amend_result, "final_hash": None,
            "error": f"git commit --amend failed (returncode={proc_amend.returncode}): {proc_amend.stderr}",
        }

    # final hash (amend 후의 HEAD)
    # 2-step: full SHA → short=7 (F-7+ 의 정공법, v0.7.26)
    proc_full = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        capture_output=True, text=True, timeout=5, cwd=str(REPO_ROOT),
    )
    if proc_full.returncode == 0 and proc_full.stdout.strip():
        head_full = proc_full.stdout.strip()
        proc_short = subprocess.run(
            ["git", "rev-parse", "--short=7", head_full],
            capture_output=True, text=True, timeout=5, cwd=str(REPO_ROOT),
        )
        if proc_short.returncode == 0 and proc_short.stdout.strip():
            final_hash = proc_short.stdout.strip()[:7]
        else:
            final_hash = None
    else:
        final_hash = None

    return {
        "ok": True, "sync_result": sync_result, "amend_result": amend_result, "final_hash": final_hash,
        "error": None,
    }


# ---------------------------------------------------------------------------
# 3. note-draft
# ---------------------------------------------------------------------------


def collect_commits_since(from_tag: str) -> list[dict]:
    """git log <from_tag>..HEAD 의 commit 목록."""
    proc = subprocess.run(
        ["git", "log", f"{from_tag}..HEAD", "--pretty=format:%h|%s|%an|%ai"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=30,
    )
    if proc.returncode != 0:
        return []
    rows = []
    for line in proc.stdout.strip().split("\n"):
        if not line:
            continue
        parts = line.split("|", 3)
        if len(parts) == 4:
            rows.append({
                "short": parts[0], "subject": parts[1],
                "author": parts[2], "date": parts[3][:10],
            })
    return rows


def draft_release_note(to_version: str, commits: list[dict], from_tag: str) -> str:
    """release note skeleton 자동 생성.

    기존 Beta-v<X>.<Y>.<Z>.md 의 패턴 따름:
    - TL;DR + 핵심 추가
    - Commit
    - Reference
    """
    today = datetime.now().strftime("%Y-%m-%d")
    feat_commits = [c for c in commits if c["subject"].startswith("feat")]
    chore_commits = [c for c in commits if c["subject"].startswith("chore")]
    docs_commits = [c for c in commits if c["subject"].startswith("docs")]

    lines = [
        f"# Beta v{to_version} — (자동 생성, 편집 필요) ({today})",
        "",
        "> 본 release note 는 `tools/release_pipeline.py note-draft` 의 *skeleton*.",
        "> commit hash / 본문 / Reference 등을 *수동 편집* 후 `docs/v<X>.<Y>.<Z>-release.md` 로 commit.",
        f"> 범위: `{from_tag}..HEAD` ({len(commits)} commit)",
        "",
        "## TL;DR",
        "",
        f"- {len(feat_commits)} feat / {len(chore_commits)} chore / {len(docs_commits)} docs commit",
        f"- 범위: `{from_tag}..HEAD`",
        "",
        "## 핵심 추가",
        "",
        "### feat",
        "",
    ]
    for c in feat_commits:
        lines.append(f"- `{c['short']}` {c['subject']}")
    if chore_commits:
        lines += ["", "### chore", ""]
        for c in chore_commits:
            lines.append(f"- `{c['short']}` {c['subject']}")
    if docs_commits:
        lines += ["", "### docs", ""]
        for c in docs_commits:
            lines.append(f"- `{c['short']}` {c['subject']}")

    lines += [
        "",
        "## Commit",
        "",
        "| Hash | Subject |",
        "|---|---|",
    ]
    for c in commits[:30]:  # max 30
        lines.append(f"| `{c['short']}` | {c['subject']} |")
    if len(commits) > 30:
        lines.append(f"| ... | ({len(commits) - 30} more) |")

    lines += [
        "",
        "## Reference",
        "",
        f"- 이전 release note: `Beta-v{from_tag.replace('v', '')}.md`",
        f"- memory entry: 추후 추가",
        "",
        "---",
        "",
        f"<!-- Auto-generated by tools/release_pipeline.py note-draft --from={from_tag} --to={to_version} -->",
        "",
    ]
    return "\n".join(lines)


def cmd_note_draft(args) -> dict:
    """git log --since=<from_tag> → release note skeleton."""
    commits = collect_commits_since(args.from_tag)
    if not commits:
        return {"mode": "error", "error": f"no commits since {args.from_tag}"}
    note = draft_release_note(args.to, commits, args.from_tag)
    output_path = RELEASES_DIR / f"Beta-v{args.to}.md"
    if args.dry_run:
        return {
            "mode": "dry-run",
            "output_path": str(output_path.relative_to(REPO_ROOT)),
            "commits": len(commits),
            "preview_first_500": note[:500],
        }
    output_path.write_text(note)
    return {
        "mode": "applied",
        "output_path": str(output_path.relative_to(REPO_ROOT)),
        "commits": len(commits),
    }


# ---------------------------------------------------------------------------
# 3.5 changelog-gen (Phase 4 — v0.7.14+)
# ---------------------------------------------------------------------------


RELEASE_RE = re.compile(r"\(v(\d+\.\d+(?:\.\d+)?)\)")
# commit subject prefix → Keep-a-Changelog section mapping
SECTION_PREFIXES = {
    "feat": "Added",
    "fix": "Fixed",
    "docs": "Changed",  # docs 변경 → Changed (Keep-a-Changelog 의 "Changed" 섹션)
    "refactor": "Changed",
    "perf": "Changed",
    "chore": "Changed",  # chore 는 빌드/CI → Changed 로 흡수 (Keep-a-Changelog 표준)
    "test": "Changed",
    "build": "Changed",
    "ci": "Changed",
}


def collect_commits_all_time() -> list[dict]:
    """git log all-time 의 commit (subject 의 vX.Y.Z 추출).

    v0.7.15+ deprecation: prefer collect_commits_in_range(from_ref, to_ref).
    """
    proc = subprocess.run(
        ["git", "log", "--all", "--pretty=format:%h|%H|%an|%ai|%s"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=60,
    )
    if proc.returncode != 0:
        return []
    return _parse_git_log(proc.stdout)


def collect_commits_in_range(from_ref: str | None, to_ref: str = "HEAD") -> list[dict]:
    """git log <from>..<to> 의 commit. (v0.7.15+ filter).

    Args:
        from_ref: 시작 ref (tag or commit hash). None 이면 --all (전체 history).
        to_ref: 종료 ref (default HEAD).

    Returns:
        commit dict list. from_ref 가 invalid (e.g. unknown tag) 시 empty list + stderr 의 error.
    """
    if from_ref is None:
        return collect_commits_all_time()
    # git log <from>..<to>
    range_arg = f"{from_ref}..{to_ref}"
    proc = subprocess.run(
        ["git", "log", range_arg, "--pretty=format:%h|%H|%an|%ai|%s"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=60,
    )
    if proc.returncode != 0:
        # from_ref 또는 to_ref invalid. caller 가 error 처리.
        return []
    return _parse_git_log(proc.stdout)


def _parse_git_log(pretty_output: str) -> list[dict]:
    """`git log --pretty=format:...` output → commit dict list (RELEASE_RE parse)."""
    rows = []
    for line in pretty_output.strip().split("\n"):
        if not line:
            continue
        parts = line.split("|", 4)
        if len(parts) < 5:
            continue
        short, full, author, date, subject = parts
        m = RELEASE_RE.search(subject)
        version = m.group(1) if m else "unreleased"
        rows.append({
            "short": short, "full": full, "author": author,
            "date": date[:10], "subject": subject, "version": version,
        })
    return rows


def categorize_by_section(subject: str) -> str:
    """commit subject prefix → Keep-a-Changelog section."""
    # `feat(...)`, `fix:` 등 첫 token 추출
    m = re.match(r"^([a-zA-Z]+)", subject)
    if not m:
        return "Changed"
    prefix = m.group(1).lower()
    return SECTION_PREFIXES.get(prefix, "Changed")


def draft_changelog(commits: list[dict], unreleased_label: str = "Unreleased") -> str:
    """multi-release commit → Keep-a-Changelog 형식 CHANGELOG.md 본문.

    e.g.:
        # Changelog
        ...
        ## [0.7.10] - 2026-06-14
        ### Added
        - ...
        ### Fixed
        - ...
    """
    # group by version
    by_version: dict[str, list[dict]] = {}
    for c in commits:
        by_version.setdefault(c["version"], []).append(c)

    # version order: 사전 reverse (latest first)
    versions = sorted(by_version.keys(), reverse=True)

    lines = [
        "# Changelog",
        "",
        "All notable changes to this project will be documented in this file.",
        "",
        "본 파일은 `tools/release_pipeline.py changelog-gen` 으로 자동 생성됩니다 (v0.7.14+).",
        "수동 편집도 가능하나 다음 release 시 자동 갱신 시 충돌 가능.",
        "",
    ]

    for ver in versions:
        if ver == "unreleased":
            label = unreleased_label
        else:
            label = ver
        v_commits = by_version[ver]
        # head commit (latest)
        head = v_commits[0]
        # date = head commit 의 date
        lines += [
            f"## [{label}] - {head['date']}",
            "",
        ]
        # section 별 분류
        by_section: dict[str, list[dict]] = {}
        for c in v_commits:
            sec = categorize_by_section(c["subject"])
            by_section.setdefault(sec, []).append(c)
        # section 출력 (Keep-a-Changelog 표준 6 종)
        for sec in ["Added", "Changed", "Fixed", "Deprecated", "Removed", "Security"]:
            if sec not in by_section:
                continue
            lines += [f"### {sec}", ""]
            for c in by_section[sec][:30]:  # max 30
                lines.append(f"- {c['subject']} ({c['short']})")
            if len(by_section[sec]) > 30:
                lines.append(f"- ... ({len(by_section[sec]) - 30} more)")
            lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def _resolve_notes_file(version: str, template: str, *, dry_run: bool = False) -> dict:
    """v0.7.24+ --notes-template flag 의 release notes file 결정.

    Templates:
        - default: `Beta-v{version}.md` (기존 동작)
        - detailed: `Beta-v{version}.md` + 1st paragraph (default 와 동일, 명시적)
        - simple: `Beta-v{version}-simple.md` (1 line summary)
        - changelog: `CHANGELOG.md` (Keep-a-Changelog 1.1.0 형식, v0.7.14 의 changelog-gen 의 output)
        - custom:<path>: 임의 path

    Returns:
        {"notes_file": Path, "source": str, "error": str | None}
    """
    template = (template or "default").strip()
    if template == "default" or template == "detailed":
        notes_file = RELEASES_DIR / f"Beta-v{version}.md"
        return {"notes_file": notes_file, "source": template, "error": None}
    elif template == "simple":
        notes_file = RELEASES_DIR / f"Beta-v{version}-simple.md"
        if not notes_file.exists() and not dry_run:
            # simple: default notes 의 1st # 헤더 + 1st ## 헤더 + 1st paragraph 만 자동 generate
            # 본문 추출: 1st # + 1st ## + (1st blank skip) + 본문 line + 2nd blank (paragraph 끝)
            default_notes = RELEASES_DIR / f"Beta-v{version}.md"
            if default_notes.exists():
                content = default_notes.read_text(encoding="utf-8")
                lines = content.split("\n")
                # 1st # 헤더
                first_h1 = next((i for i, l in enumerate(lines) if l.startswith("# ")), -1)
                if first_h1 >= 0:
                    simple_lines: list[str] = []
                    seen_h1 = False
                    seen_first_h2 = False
                    # 1st # 헤더 + 1st ## 헤더 + 본문 (2nd ## 헤더 또는 2nd blank 전까지)
                    # 본문 = 1st ## 헤더 *후* 의 non-blank line 들
                    blank_count = 0
                    in_body = False
                    for i in range(first_h1, len(lines)):
                        line = lines[i]
                        if not seen_h1:
                            if line.startswith("# "):
                                simple_lines.append(line)
                                seen_h1 = True
                            continue
                        if line.startswith("## "):
                            if not seen_first_h2:
                                simple_lines.append(line)
                                seen_first_h2 = True
                                in_body = True
                            else:
                                # 2nd ## 헤더 → 끝
                                break
                        elif line.strip() == "":
                            if in_body:
                                blank_count += 1
                                if blank_count >= 2:
                                    # 2nd blank → 1st paragraph 끝
                                    break
                        else:
                            if in_body:
                                simple_lines.append(line)
                                blank_count = 0
                    notes_file.parent.mkdir(parents=True, exist_ok=True)
                    notes_file.write_text("\n".join(simple_lines).rstrip() + "\n", encoding="utf-8")
        return {"notes_file": notes_file, "source": template, "error": None}
    elif template == "changelog":
        notes_file = REPO_ROOT / "workflow-source" / "CHANGELOG.md"
        return {"notes_file": notes_file, "source": template, "error": None}
    elif template.startswith("custom:"):
        custom_path = Path(template[len("custom:"):])
        if not custom_path.is_absolute():
            custom_path = REPO_ROOT / custom_path
        return {"notes_file": custom_path, "source": template, "error": None}
    else:
        return {
            "notes_file": Path(),
            "source": template,
            "error": f"unknown --notes-template value: {template!r}. Use 'default' / 'detailed' / 'simple' / 'changelog' / 'custom:<path>'",
        }


def cmd_changelog_gen(args) -> dict:
    """multi-release git log → CHANGELOG.md 본문 생성 (Keep-a-Changelog 형식)."""
    from_tag = getattr(args, "from_tag", None)
    to_tag = getattr(args, "to_tag", "HEAD")
    commits = collect_commits_in_range(from_tag, to_tag)
    if not commits:
        if from_tag is not None:
            return {
                "error": f"no commits in range {from_tag}..{to_tag} (from_tag 또는 to_tag invalid 할 수 있음)",
            }
        return {"mode": "error", "error": "no commits in git log"}
    body = draft_changelog(commits, unreleased_label=getattr(args, "unreleased_label", "Unreleased"))
    output_path = Path(args.output) if args.output else (REPO_ROOT / "CHANGELOG.md")
    if args.dry_run:
        return {
            "mode": "dry-run",
            "output_path": str(output_path),
            "commits": len(commits),
            "versions": len(set(c["version"] for c in commits)),
            "from_tag": from_tag,
            "to_tag": to_tag,
            "preview_first_500": body[:500],
        }
    output_path.parent.mkdir(parents=True, exist_ok=True)
    if atomic_write_text is not None:
        atomic_write_text(output_path, body)
    else:
        output_path.write_text(body, encoding="utf-8")
    return {
        "mode": "applied",
        "output_path": str(output_path),
        "commits": len(commits),
        "versions": len(set(c["version"] for c in commits)),
        "from_tag": from_tag,
        "to_tag": to_tag,
    }


def find_dist_files(version: str) -> list[Path]:
    """dist/ 의 wheel + sdist glob. PEP 440 normalize: 0.7.10 → 0.7.10b0."""
    dist = REPO_ROOT / "dist"
    if not dist.exists():
        return []
    # PEP 440: 0.7.10-beta → 0.7.10b0
    base = version.split("-")[0]
    pep_version = base  # wheel 파일명은 X.Y.Z 형태
    return sorted(dist.glob(f"standard_ai_workflow-{pep_version}*"))


def cmd_release(args) -> dict:
    """GitHub Release 생성 (gh release create).

    사전 점검: --skip-validate 미지정 시 validate 4 source 자동 호출.
    1+ source fail 시 release 중단 (exit 1).

    **v0.7.18+ release coordination observability**:
    `tag` 결정 후 `git ls-remote origin` 로 *원격 tag 존재 여부* 확인. 존재 시
    - default: exit 1 + auto-bump hint
    - `--auto-bump`: `next_available_version()` 로 다음 version 결정 + version-bump 자동 + re-flow
    v0.7.16 의 race lesson 반영 (memory #22 §release coordination race).

    gh auth 인증된 환경 가정. token 회전 부담은 caller 책임.
    """
    results: dict = {"pre_check": {}, "gh_commands": [], "mode": "dry-run" if args.dry_run else "apply"}

    # 1. validate (사전 점검)
    if not args.skip_validate:
        val_result = cmd_validate(args)
        results["pre_check"] = val_result
        if not all(v.get("ok", False) for v in val_result.values()):
            return {**results, "error": "validate failed; abort release"}

    # 2. dist 파일 glob
    # v0.7.13+: --version override (backfill 시 staging 용도). default 는 read_version().
    if getattr(args, "version", None):
        version = args.version
        results["version_source"] = "cli-flag"
    else:
        version = read_version()
        results["version_source"] = "pyproject.toml"

    # v0.7.18+ auto-bump: pre-check 후 tag 결정 전에 호출
    if getattr(args, "auto_bump", False):
        bump_info = next_available_version(version)
        if bump_info["bumped"]:
            version = bump_info["next"]
            results["version_source"] = "auto-bump"
            results["auto_bump"] = bump_info
            # version-bump 자동 적용 (in-place). write_version + write_workflow_kit_version
            write_version(version)
            suffix = "beta"
            if read_workflow_kit_version().endswith("-beta"):
                suffix = "beta"
            elif read_workflow_kit_version().endswith("-alpha"):
                suffix = "alpha"
            else:
                suffix = ""  # default
            write_workflow_kit_version(version, suffix=("-" + suffix) if suffix else "")
        else:
            results["auto_bump"] = bump_info  # bumped=False, info only

    dist_files = find_dist_files(version)
    if not dist_files:
        return {**results, "error": f"no dist files found for version {version} (run `python3 -m build` first)"}

    # 3. tag 결정 + 원격 tag pre-check (v0.7.18+)
    tag = f"v{version}-beta"
    # v0.7.24+: --notes-template flag 로 release notes format 자유도
    notes_template = getattr(args, "notes_template", "default") or "default"
    notes_resolution = _resolve_notes_file(version, notes_template, dry_run=args.dry_run)
    if notes_resolution.get("error"):
        return {**results, "error": notes_resolution["error"]}
    notes_file = notes_resolution["notes_file"]
    if not notes_file.exists():
        return {**results, "error": f"release note not found: {notes_file}"}

    # 3.5 원격 tag pre-check + tag push (v0.7.18+ race lesson, v0.7.21+ follow-up)
    # v0.7.21 fix: tag push 와 release 의 coupling. *순서*:
    #   1. pre-check: remote 에 tag 가 이미 push 됐는지 확인
    #   2. tag push: pre-check fail 시 default = skip, --allow-existing-tag 면 skip + 진행, --auto-bump 면 bump
    #   3. gh release create: --verify-tag 가 tag 의 remote 존재 검증 (pre-check 와 *redundant* 한 부분)
    if not args.dry_run:
        tag_check = _check_remote_tag(tag)
        results["tag_pre_check"] = tag_check
        if tag_check["exists"]:
            if not getattr(args, "allow_existing_tag", False):
                return {
                    **results,
                    "error": (
                        f"remote tag {tag} already exists at {tag_check['remote_url']}. "
                        f"v0.7.16 race 정공법: --auto-bump 으로 다음 version 자동 bump, "
                        f"--allow-existing-tag 으로 *기존 tag* 에 re-attach, "
                        f"또는 --version=<next> 명시."
                    ),
                }
            # --allow-existing-tag: skip pre-check fail, 그대로 release 진행
            results["tag_pre_check_skipped"] = "allow-existing-tag"

    # 3.6 local tag push (v0.7.21+ — tag push 와 release 의 coupling)
    if not args.dry_run:
        push_tag_proc = subprocess.run(
            ["git", "push", "origin", f"refs/tags/{tag}"],
            cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=30,
        )
        results["tag_push"] = {
            "tag": tag,
            "returncode": push_tag_proc.returncode,
            "stdout_tail": push_tag_proc.stdout.strip().split("\n")[-1] if push_tag_proc.stdout else "",
            "stderr_tail": push_tag_proc.stderr.strip().split("\n")[-1] if push_tag_proc.stderr else "",
        }
        if push_tag_proc.returncode != 0 and not getattr(args, "allow_existing_tag", False):
            return {**results, "error": f"git push tag {tag} failed: {push_tag_proc.stderr.strip()}"}
    else:
        # dry-run: pre-check 결과 + warning (plan 검증)
        tag_check = _check_remote_tag(tag)
        results["tag_pre_check"] = tag_check
        if tag_check["exists"]:
            results["tag_pre_check_warning"] = f"remote tag {tag} already exists (dry-run: pre-check only)"

    rel_assets = [str(f.relative_to(REPO_ROOT)) for f in dist_files]
    results["tag"] = tag
    results["assets"] = rel_assets
    # v0.7.24+: notes_file 가 in-repo 면 relative path, 그 외 (예: changelog) 면 absolute
    try:
        results["notes_file"] = str(notes_file.relative_to(REPO_ROOT))
    except ValueError:
        results["notes_file"] = str(notes_file)

    # 4. gh command build
    repo_remote = subprocess.run(
        ["git", "remote", "get-url", "origin"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=10,
    )
    repo = repo_remote.stdout.strip().replace("https://github.com/", "").replace(".git", "")
    results["repo"] = repo

    gh_cmd = [
        "gh", "release", "create", tag,
        "--repo", repo,
        "--title", f"Beta v{version}",
        "--notes-file", str(notes_file),
        "--target", "main",
        "--verify-tag",
    ] + [str(f) for f in dist_files]
    results["gh_command"] = " ".join(gh_cmd)

    if args.dry_run:
        return results

    # 5. gh auth check + release create
    auth_proc = subprocess.run(["gh", "auth", "status"], capture_output=True, text=True, timeout=10)
    if auth_proc.returncode != 0:
        return {**results, "error": "gh auth not authenticated"}
    results["gh_auth_ok"] = True

    proc = subprocess.run(gh_cmd, capture_output=True, text=True, timeout=120)
    results["gh_exit_code"] = proc.returncode
    if proc.stdout:
        results["gh_stdout_tail"] = proc.stdout.strip().split("\n")[-1]
    if proc.stderr:
        results["gh_stderr_tail"] = proc.stderr.strip().split("\n")[-1]
    if proc.returncode != 0:
        return {**results, "error": f"gh release create failed: exit {proc.returncode}"}
    return results


# ---------------------------------------------------------------------------
# 5. verify (Phase 2 — v0.7.10)
# ---------------------------------------------------------------------------


def cmd_verify(args) -> dict:
    """GitHub Release 의 tag + asset 검증 (read-only)."""
    tag = args.tag
    if tag.startswith("v"):
        tag_full = tag
    else:
        tag_full = f"v{tag}"
    # 1. gh release view (--json tagName,name,url,assets,isPrerelease,publishedAt)
    gh_cmd = [
        "gh", "release", "view", tag_full,
        "--repo", _get_repo(),
        "--json", "tagName,name,url,assets,isPrerelease,publishedAt",
    ]
    results: dict = {"tag": tag_full, "gh_command": " ".join(gh_cmd), "mode": "read-only"}

    if args.dry_run:
        return results

    proc = subprocess.run(gh_cmd, capture_output=True, text=True, timeout=30)
    if proc.returncode != 0:
        return {**results, "error": f"release not found: {tag_full} (gh exit {proc.returncode})"}

    try:
        release_data = json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        return {**results, "error": f"gh release view JSON parse failed: {e}"}

    results["name"] = release_data.get("name")
    results["url"] = release_data.get("url")
    results["is_prerelease"] = release_data.get("isPrerelease")
    results["created_at"] = release_data.get("publishedAt")
    results["assets"] = [a.get("name") for a in release_data.get("assets", [])]
    return results


def _get_repo() -> str:
    """git remote origin → 'owner/repo' 추출."""
    proc = subprocess.run(
        ["git", "remote", "get-url", "origin"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=10,
    )
    return proc.stdout.strip().replace("https://github.com/", "").replace(".git", "")


# ---------------------------------------------------------------------------
# 6. rollback (Phase 2 — v0.7.10)
# ---------------------------------------------------------------------------


def cmd_rollback(args) -> dict:
    """GitHub Release + git tag 삭제 (destructive).

    --dry-run: 삭제 명령만 print, 실제 호출 0.
    --apply: gh release delete + git tag -d + git push --delete origin <tag>.
    """
    tag = args.tag if args.tag.startswith("v") else f"v{args.tag}"
    repo = _get_repo()

    commands = [
        # local tag delete
        ["git", "tag", "-d", tag],
        # remote tag delete
        ["git", "push", "--delete", "origin", tag],
        # gh release delete
        ["gh", "release", "delete", tag, "--repo", repo, "--yes"],
    ]
    results: dict = {
        "tag": tag,
        "repo": repo,
        "commands": [" ".join(c) for c in commands],
        "mode": "dry-run" if args.dry_run else "apply",
    }

    if args.dry_run:
        return results

    # 실제 실행
    executed: list[dict] = []
    for cmd in commands:
        proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        executed.append({
            "cmd": " ".join(cmd),
            "exit_code": proc.returncode,
            "stdout_tail": (proc.stdout or "").strip().split("\n")[-1] if proc.stdout else "",
        })
        if proc.returncode != 0:
            results["error"] = f"command failed: {' '.join(cmd)} (exit {proc.returncode})"
            break
    results["executed"] = executed
    results["ok"] = "error" not in results
    return results

    p_cl.add_argument("--unreleased-label", default="Unreleased",
                      help="unreleased commit group 의 label (default: 'Unreleased')")
    p_cl.add_argument("--from-tag", default=None,
                      help="git log 시작 ref (e.g. v0.7.0-beta). 미지정 시 --all (전체 history)")
    p_cl.add_argument("--to-tag", default="HEAD",
                      help="git log 종료 ref (default: HEAD)")
    p_cl.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_cl.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_cl.add_argument("--json", action="store_true")


# ---------------------------------------------------------------------------
# 7. dist (Phase 3 — v0.7.11)
# ---------------------------------------------------------------------------


def _check_build_module() -> dict:
    """`build` module 가용성 체크. 없으면 pip install 안내.

    Returns:
        {"available": bool, "hint": str (if not available)}
    """
    try:
        import build  # type: ignore[import-not-found]  # noqa: F401

        return {"available": True, "version": getattr(build, "__version__", "unknown")}
    except ImportError:
        return {
            "available": False,
            "hint": "pip install build (or `python3 -m pip install --user build`)",
        }


def _build_command(out_dir: Path, *, sdist_only: bool = False, wheel_only: bool = False) -> list[str]:
    """`python3 -m build` 호출 command list. PEP 517/518 build."""
    cmd = [sys.executable, "-m", "build", "--outdir", str(out_dir)]
    if sdist_only:
        cmd.append("--sdist")
    elif wheel_only:
        cmd.append("--wheel")
    cmd.append(str(REPO_ROOT))
    return cmd


def _expected_dist_pattern(version: str) -> str:
    """version (e.g. '0.7.10' or '0.7.10-beta') → dist file prefix (PEP 440 normalize)."""
    return version.split("-")[0]


def cmd_dist(args) -> dict:
    """Wheel + sdist 자동 빌드 (`python3 -m build`).

    pre-check: `build` module 가용성 → 부재 시 graceful fail.
    dry-run: command + PEP 440 normalize 만 print. exit 0.
    apply: subprocess `python3 -m build` 실행. exit code + dist glob 결과 report.
    """
    _dist_dir = REPO_ROOT / "dist"
    results: dict = {"mode": "dry-run" if args.dry_run else "apply", "out_dir": str(_dist_dir)}

    # 1) pre-check: build module 가용성
    build_check = _check_build_module()
    results["build_module"] = build_check
    if not build_check["available"]:
        results["error"] = f"build module not installed: {build_check['hint']}"
        return results

    # 2) version read (pyproject.toml)
    try:
        current_version = read_version()
    except Exception as e:  # pragma: no cover
        results["error"] = f"pyproject.toml version read 실패: {e}"
        return results
    results["version"] = current_version
    cmd = _build_command(
        _dist_dir,
        sdist_only=getattr(args, "sdist_only", False),
        wheel_only=getattr(args, "wheel_only", False),
    )
    results["command"] = " ".join(cmd)
    results["expected_pattern"] = f"standard_ai_workflow-{_expected_dist_pattern(current_version)}*"

    # 4) skip-existing check (--skip-existing)
    if getattr(args, "skip_existing", False) and _dist_dir.exists():
        existing = find_dist_files(current_version)
        if existing:
            results["mode"] = "skip"
            results["skipped"] = True
            results["existing"] = [f.name for f in existing]
            results["ok"] = True
            return results

    # 5) dry-run: command plan 만 반환
    if args.dry_run:
        results["ok"] = True
        return results

    # 6) apply: subprocess `python3 -m build` 실행
    _dist_dir.mkdir(parents=True, exist_ok=True)
    try:
        proc = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=args.timeout,
            cwd=str(REPO_ROOT),
        )
    except subprocess.TimeoutExpired:
        results["error"] = f"build timeout after {args.timeout}s"
        return results

    results["returncode"] = proc.returncode
    # 마지막 5 line 의 stdout/stderr 만 report (full log 은 debug 용)
    out_tail = proc.stdout.strip().splitlines()[-5:] if proc.stdout.strip() else []
    err_tail = proc.stderr.strip().splitlines()[-5:] if proc.stderr.strip() else []
    results["stdout_tail"] = out_tail
    results["stderr_tail"] = err_tail

    if proc.returncode != 0:
        results["error"] = f"build failed: exit {proc.returncode}"
        results["ok"] = False
        return results

    # 7) post-check: dist glob 결과
    built = find_dist_files(current_version)
    results["built"] = [f.name for f in built]
    results["ok"] = True
    return results

# ---------------------------------------------------------------------------
# main
# ---------------------------------------------------------------------------


def main() -> int:
    p = argparse.ArgumentParser(
        description="standard-ai-workflow release pipeline (v0.7.9+)",
    )
    sub = p.add_subparsers(dest="command", required=True)

    # validate
    p_val = sub.add_parser("validate", help="release-readiness 검증 (4 source)")
    p_val.add_argument("--skip-packaging", action="store_true", help="check_packaging skip")
    p_val.add_argument("--skip-doctor", action="store_true", help="doctor skip")
    p_val.add_argument("--skip-state", action="store_true", help="state.json check skip")
    p_val.add_argument("--skip-git", action="store_true", help="git status check skip")
    p_val.add_argument("--dry-run", action="store_true")
    p_val.add_argument("--json", action="store_true")

    # version-bump (v0.7.14+ auto-sync workflow_kit/__init__.py)
    p_vb = sub.add_parser("version-bump", help="pyproject.toml version patch + workflow_kit/__init__.py __version__ auto-sync")
    p_vb.add_argument("--patch", action="store_true", help="patch bump (default)")
    p_vb.add_argument("--minor", action="store_true", help="minor bump")
    p_vb.add_argument("--major", action="store_true", help="major bump")
    p_vb.add_argument("--to", help="explicit version (e.g. 0.7.9)")
    p_vb.add_argument("--no-init", action="store_true", dest="no_init",
                       help="workflow_kit/__init__.py __version__ sync skip (CI / override 시나리오)")
    p_vb.add_argument("--dry-run", action="store_true", dest="dry_run",
                       help="bump plan 만 출력 (default: --apply)")
    p_vb.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_vb.add_argument("--skip-sync-hash", action="store_true", dest="skip_sync_hash",
                       help="post-step sync_release_hash 자동 호출 skip (TASK-V0726-003, manual override)")
    p_vb.add_argument("--json", action="store_true", help="JSON output (CI integration)")

    # note-draft
    p_nd = sub.add_parser("note-draft", help="release note skeleton 자동 생성")
    p_nd.add_argument("--from", dest="from_tag", required=True, help="이전 release tag (e.g. v0.7.8)")
    p_nd.add_argument("--to", required=True, help="새 release version (e.g. 0.7.9)")
    p_nd.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_nd.add_argument("--apply", dest="apply", action="store_true", default=True)

    # changelog-gen (Phase 4 — v0.7.14+, Keep-a-Changelog 형식, v0.7.15+ filter)
    p_cl = sub.add_parser("changelog-gen", help="multi-release git log → CHANGELOG.md 본문 (Keep-a-Changelog 형식)")
    p_cl.add_argument("--output", default=None,
                      help="output file path (default: workflow-source/CHANGELOG.md)")
    p_cl.add_argument("--unreleased-label", default="Unreleased",
                      help="unreleased commit group 의 label (default: 'Unreleased')")
    p_cl.add_argument("--from-tag", default=None,
                      help="git log 시작 ref (e.g. v0.7.0-beta). 미지정 시 --all (전체 history)")
    p_cl.add_argument("--to-tag", default="HEAD",
                      help="git log 종료 ref (default: HEAD)")
    p_cl.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_cl.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_cl.add_argument("--json", action="store_true")

    # release (Phase 2 — v0.7.10, v0.7.13+ --version)
    p_rel = sub.add_parser("release", help="GitHub Release 생성 (gh release create)")
    p_rel.add_argument("--skip-validate", action="store_true", help="validate 사전 점검 skip")
    p_rel.add_argument("--version", default=None,
                       help="version override (e.g. 0.7.5 for backfill). default: pyproject.toml [project] version")
    p_rel.add_argument("--auto-bump", dest="auto_bump", action="store_true", default=False,
                       help="remote tag pre-check fail 시 다음 version 으로 자동 bump + re-flow. "
                            "v0.7.18+: release coordination observability.")
    p_rel.add_argument("--allow-existing-tag", dest="allow_existing_tag", action="store_true", default=False,
                       help="remote tag pre-check 가 'already exists' 일 때 *skip* + 그대로 release 진행. "
                            "v0.7.21+ follow-up: tag push 와 release 의 coupling fix. "
                            "*의도된* tag re-push (e.g. wheel re-attach) 또는 backfill 시에만 사용.")
    p_rel.add_argument("--notes-template", dest="notes_template", default="default",
                       help="release notes format 결정. v0.7.24+. "
                            "'default' (Beta-v<X>.<Y>.<Z>.md) / 'detailed' (default 와 동일) / "
                            "'simple' (1 line summary) / 'changelog' (workflow-source/CHANGELOG.md) / "
                            "'custom:<path>' (임의 path).")
    p_rel.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_rel.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_rel.add_argument("--json", action="store_true")

    # verify (Phase 2 — v0.7.10)
    p_ver = sub.add_parser("verify", help="GitHub Release 의 tag + asset 검증 (read-only)")
    p_ver.add_argument("--tag", required=True, help="tag 이름 (e.g. v0.7.9-beta 또는 0.7.9)")
    p_ver.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_ver.add_argument("--json", action="store_true")

    # rollback (Phase 2 — v0.7.10)
    p_rb = sub.add_parser("rollback", help="GitHub Release + git tag 삭제 (destructive)")
    p_rb.add_argument("--tag", required=True, help="tag 이름 (e.g. v0.7.9-beta 또는 0.7.9)")
    p_rb.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_rb.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_rb.add_argument("--json", action="store_true")

    # dist (Phase 3 — v0.7.11)
    p_dist = sub.add_parser("dist", help="wheel + sdist 자동 빌드 (`python3 -m build`)")
    p_dist.add_argument("--sdist-only", action="store_true", help="sdist 만 빌드")
    p_dist.add_argument("--wheel-only", action="store_true", help="wheel 만 빌드")
    p_dist.add_argument("--skip-existing", action="store_true", help="dist/ 의 current-version 파일 있으면 skip")
    p_dist.add_argument("--timeout", type=int, default=300, help="subprocess timeout in sec (default 300)")
    p_dist.add_argument("--dry-run", action="store_true", dest="dry_run")
    p_dist.add_argument("--apply", dest="apply", action="store_true", default=True)
    p_dist.add_argument("--json", action="store_true")

    args = p.parse_args()
    if getattr(args, "dry_run", False):
        args.apply = False

    if args.command == "validate":
        result = cmd_validate(args)
    elif args.command == "version-bump":
        result = cmd_version_bump(args)
    elif args.command == "note-draft":
        result = cmd_note_draft(args)
    elif args.command == "changelog-gen":
        result = cmd_changelog_gen(args)
    elif args.command == "release":
        result = cmd_release(args)
    elif args.command == "verify":
        result = cmd_verify(args)
    elif args.command == "rollback":
        result = cmd_rollback(args)
    elif args.command == "dist":
        result = cmd_dist(args)
    else:
        p.error(f"unknown command: {args.command}")
        return 2

    if getattr(args, "json", False):
        print(json.dumps(result, ensure_ascii=False, indent=2))
    else:
        dry_run = getattr(args, "dry_run", False)
        mode_label = "DRY-RUN" if dry_run else "APPLY"
        if args.command == "verify":
            mode_label = "READ-ONLY"
        print(f"=== {args.command} ({mode_label}) ===")
        for k, v in result.items():
            if isinstance(v, dict):
                print(f"  {k}:")
                for k2, v2 in v.items():
                    if isinstance(v2, list) and len(str(v2)) > 80:
                        print(f"    {k2}: [{', '.join(str(x) for x in v2[:3])}...]")
                    else:
                        print(f"    {k2}: {v2}")
            elif isinstance(v, list):
                if len(str(v)) > 80:
                    print(f"  {k}: [{', '.join(str(x) for x in v[:3])}...]")
                else:
                    print(f"  {k}: {v}")
            else:
                print(f"  {k}: {v}")

    # exit code: validate / release / verify / rollback / dist 의 ok/error 기반
    if args.command in ("validate", "release", "verify", "rollback", "dist"):
        if "error" in result:
            return 1
        if args.command in ("release", "rollback", "dist") and not result.get("ok", True):
            return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
