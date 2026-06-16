r"""Mirror list byte-identity test (DevHub-specific, PR #616 follow-up).

Verifies the 15-pattern mirror list in docs/llm-wiki/mirror-list.md (Phase 1+1.5+3
scope) is byte-identical between source (SRC = docs/*/ai-workflow/memory branch)
and the raw mirror (ai-workflow/wiki/raw/projects/devhub). 4 test scenario:

1. test_list_sources_paths_exist: list_sources 의 모든 path 가 SRC 에 실제 존재.
2. test_mirror_byte_identity: list_sources 의 path 의 source byte == raw mirror byte.
3. test_mirror_scope_compliance: list_sources 의 path 가 raw mirror 의 subset (allowlist: _manifest.md).
4. test_mirror_list_doc_consistency: docs 의 "15 패턴" claim + list_sources file count 정합.

Refs: PR #604 (raw mirror full commit), PR #610 (manifest regen), PR #616 (본 test).
"""

from __future__ import annotations

import os
import re
import subprocess
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
MIRROR_SOURCES = REPO_ROOT / "scripts" / "wiki-mirror-sources.sh"
MIRROR_LIST_DOC = REPO_ROOT / "docs" / "llm-wiki" / "mirror-list.md"
WIKI_SYNC = REPO_ROOT / "scripts" / "wiki-sync-devhub.sh"
RAW_MIRROR = REPO_ROOT / "ai-workflow" / "wiki" / "raw" / "projects" / "devhub"

if sys.platform == "win32":
    print(f"[check-mirror-list] SKIP: POSIX only")
    sys.exit(0)

if not MIRROR_SOURCES.exists():
    print(f"[check-mirror-list] SKIP: {MIRROR_SOURCES} not found")
    sys.exit(0)

if not RAW_MIRROR.is_dir():
    print(f"[check-mirror-list] SKIP: raw mirror not found: {RAW_MIRROR}")
    sys.exit(0)


# ----- result tracker -----
PASS = 0
FAIL = 0
FAILURES: list[str] = []


def _ok(name: str) -> None:
    global PASS
    PASS += 1
    print(f"  [PASS] {name}")


def _fail(name: str, reason: str) -> None:
    global FAIL
    FAIL += 1
    FAILURES.append(f"{name}: {reason}")
    print(f"  [FAIL] {name}: {reason}")


def _run_list_sources() -> list[str]:
    """Run list_sources via bash subshell. Return list of paths ($SRC/<rel>).

    Env var SRC 는 Python env dict 로 전달 (assignment-prefix SRC=... source ... 는
    scope 문제로 list_sources 의 $SRC 가 empty). export SRC=... 형태가 subshell
    전체에 적용.
    """
    env = {**os.environ, "SRC": str(REPO_ROOT)}
    r = subprocess.run(
        ["bash", "-c", f'source "{MIRROR_SOURCES}" && list_sources'],
        capture_output=True,
        text=True,
        timeout=30,
        cwd=str(REPO_ROOT),
        env=env,
    )
    if r.returncode != 0:
        raise RuntimeError(f"list_sources exit {r.returncode}: {r.stderr[-200:]}")
    return [p for p in r.stdout.split("\n") if p]


def _to_rel_path(src_path: str) -> str:
    """Convert $SRC/<rel> → <rel> with path normalization (handles double slashes)."""
    rel = src_path
    if rel.startswith(str(REPO_ROOT) + "/"):
        rel = rel[len(str(REPO_ROOT)) + 1:]
    return str(Path(rel))


# ----- 1. list_sources paths exist in source -----
def test_list_sources_paths_exist():
    name = "test_list_sources_paths_exist"
    try:
        paths = _run_list_sources()
        if not paths:
            return _fail(name, "list_sources returned 0 paths — script logic error?")
        missing: list[str] = []
        for p in paths:
            abs_path = Path(p)
            if not abs_path.is_file():
                missing.append(str(p))
        if missing:
            return _fail(
                name,
                f"{len(missing)}/{len(paths)} paths do not exist in SRC. "
                f"first 5: {missing[:5]}",
            )
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 2. list_sources paths byte-identical to raw mirror -----
def test_mirror_byte_identity():
    name = "test_mirror_byte_identity"
    try:
        paths = _run_list_sources()
        diff_files: list[str] = []
        for src_path in paths:
            rel = _to_rel_path(src_path)
            mirror_path = RAW_MIRROR / rel
            if not mirror_path.is_file():
                diff_files.append(f"MISSING in raw mirror: {rel}")
                continue
            src_hash = subprocess.run(
                ["sha256sum", str(src_path)],
                capture_output=True,
                text=True,
                timeout=5,
            ).stdout.split()[0]
            mirror_hash = subprocess.run(
                ["sha256sum", str(mirror_path)],
                capture_output=True,
                text=True,
                timeout=5,
            ).stdout.split()[0]
            if src_hash != mirror_hash:
                diff_files.append(f"DIFF: {rel}")
        if diff_files:
            return _fail(
                name,
                f"{len(diff_files)}/{len(paths)} paths differ. first 5: {diff_files[:5]}",
            )
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 3. raw mirror scope compliance -----
def test_mirror_scope_compliance():
    name = "test_mirror_scope_compliance"
    try:
        paths = _run_list_sources()
        list_rel = {_to_rel_path(p) for p in paths}
        raw_files = set()
        for md in RAW_MIRROR.rglob("*"):
            if md.is_file():
                raw_files.add(str(md.relative_to(RAW_MIRROR)))
        ALLOWLIST = {"_manifest.md"}
        unexpected = raw_files - list_rel - ALLOWLIST
        missing_in_mirror = list_rel - raw_files
        if unexpected:
            return _fail(
                name,
                f"{len(unexpected)} raw mirror files NOT in list_sources (scope violation). "
                f"first 5: {sorted(unexpected)[:5]}",
            )
        if missing_in_mirror:
            return _fail(
                name,
                f"{len(missing_in_mirror)} list_sources paths MISSING in raw mirror. "
                f"first 5: {sorted(missing_in_mirror)[:5]}",
            )
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 4. mirror-list.md doc consistency (15 pattern claim) -----
def test_mirror_list_doc_consistency():
    name = "test_mirror_list_doc_consistency"
    try:
        if not MIRROR_LIST_DOC.is_file():
            return _fail(name, f"mirror list doc not found: {MIRROR_LIST_DOC}")
        text = MIRROR_LIST_DOC.read_text(encoding="utf-8")
        # "15 패턴" string sanity in doc header
        if "15 패턴" not in text:
            return _fail(
                name,
                '"15 패턴" claim missing from mirror-list.md header. '
                "Phase 1+1.5+3 의 7+6+2 = 15 가 문서와 정합해야.",
            )
        # wiki-sync-devhub.sh 의 L90 docstring "15 패턴" claim 도 verify
        if WIKI_SYNC.is_file():
            sync_text = WIKI_SYNC.read_text(encoding="utf-8")
            if "15 패턴" not in sync_text:
                return _fail(
                    name,
                    '"15 패턴" claim missing from scripts/wiki-sync-devhub.sh '
                    "header comment. SSOT drift.",
                )
        # Sanity: list_sources output 이 >= 100 file (mirror list scope)
        paths = _run_list_sources()
        if len(paths) < 100:
            return _fail(
                name,
                f"list_sources output {len(paths)} files < 100. Phase 1+1.5+3 의 "
                f"~220 file scope 와 mismatch.",
            )
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- main -----
def main() -> int:
    print(f"[check-mirror-list] === mirror list byte-identity test (DevHub) ===")
    print(f"[check-mirror-list]   mirror list doc: {MIRROR_LIST_DOC}")
    print(f"[check-mirror-list]   list_sources: {MIRROR_SOURCES}")
    print(f"[check-mirror-list]   raw mirror: {RAW_MIRROR}")
    print("")

    test_list_sources_paths_exist()
    test_mirror_byte_identity()
    test_mirror_scope_compliance()
    test_mirror_list_doc_consistency()
    print("")

    total = PASS + FAIL
    print(f"[check-mirror-list] === summary ===")
    print(f"  passed: {PASS}/{total}")
    print(f"  failed: {FAIL}/{total}")
    if FAIL:
        print(f"  failures:")
        for f in FAILURES:
            print(f"    - {f}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
