"""wiki-ingest-from-raw.sh wrapper smoke test (DevHub-specific, PR #606 P1 follow-up).

Verifies:
1. dry-run (self, default) — step 2 preview 정상 출력, exit 0
2. dry-run (--emit-tool vendor) — vendor dry-run 정상, exit 0
3. --apply --emit-tool vendor false-positive detection — 3가지 scenario 모두
   vendor activity check (VENDOR_CHANGED=0) 가 exit 1 + error message raise

Refs: PR #606 (P1 false-positive detection), PR #610 (manifest regen fix 가 있어야
wiki-ingest 의 --apply 가 byte-identical 검증 가능).

Note on side effects: wrapper 의 --dry-run 도 wiki-sync-devhub.sh 의 `cp -p` 단계가
raw mirror 의 mtime 을 갱신하여 working dir 에 unstaged 변경을 만든다. dry-run 의
*logical* side-effect 는 0 (sources/ 미변경) 이지만 raw mirror 의 mtime 변경은
test 의 scope 밖. 따라서 test 1, 2 는 sources/ 의 content 변화 검증으로
단순화하지 않고, step 2 log + exit code 만 확인한다.
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

# ----- paths -----
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
WRAPPER = REPO_ROOT / "scripts" / "wiki-ingest-from-raw.sh"

# ----- POSIX guard -----
if sys.platform == "win32":
    print(f"[check-wiki-ingest] SKIP: POSIX only (current platform: {sys.platform})")
    print(f"[check-wiki-ingest] RESULT: 0/0 SKIP")
    sys.exit(0)

if not WRAPPER.exists():
    print(f"[check-wiki-ingest] SKIP: wrapper not found: {WRAPPER}")
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


def _run(args: list[str], timeout: int = 60) -> tuple[int, str, str]:
    """Run wrapper subprocess. Return (rc, stdout, stderr)."""
    r = subprocess.run(
        ["bash", str(WRAPPER)] + args,
        capture_output=True,
        text=True,
        timeout=timeout,
        cwd=str(REPO_ROOT),
    )
    return r.returncode, r.stdout, r.stderr


# ----- 1. dry-run (self, default) -----
def test_dry_run_self():
    """Default dry-run: step 2 preview 정상 출력, exit 0.

    Regression for PR #606 (Codex P2): dry-run self 가 vendor-only flag (
    --project/--mode) 미지원으로 argparse exit 2 + silent skip. fix 이후 dry-run
    이 default preview 만 정상 호출.
    """
    name = "test_dry_run_self"
    try:
        rc, stdout, stderr = _run([])
        if rc != 0:
            return _fail(name, f"exit {rc}, stderr: {stderr[-200:]}")
        combined = stdout + stderr
        if "step 2/3: L2 dense emit" not in combined:
            return _fail(name, "step 2 log missing — dry-run self preview 가 실행 안 됨")
        if "dry-run done" not in combined:
            return _fail(name, "'dry-run done' marker missing — preview loop 가 끝까지 안 돌음")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 2. dry-run (vendor) -----
def test_dry_run_vendor():
    """--emit-tool vendor dry-run: vendor 도구 정상 호출, exit 0.

    vendor 는 --project --mode 지원. dry-run 이 vendor 의 default dry-run 호출.
    """
    name = "test_dry_run_vendor"
    try:
        rc, stdout, stderr = _run(["--emit-tool", "vendor"])
        if rc != 0:
            return _fail(name, f"exit {rc}, stderr: {stderr[-200:]}")
        combined = stdout + stderr
        if "step 2/3: L2 dense emit" not in combined:
            return _fail(name, "step 2 log missing")
        if "tool: vendor" not in combined:
            return _fail(name, "vendor tool selected log missing — wrapper 가 vendor 분기 미진입")
        if "dry-run done" not in combined:
            return _fail(name, "'dry-run done' missing")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 3. --apply --emit-tool vendor false-positive detection (3 scenario) -----
def test_apply_vendor_false_positive_detection():
    """P1 fix 의 vendor activity check (VENDOR_CHANGED=0) 가 3가지 scenario 모두
    exit 1 + error message raise.

    Scenario A: 모든 sources/*.md 가 full body (placeholder 없음) → self 0 + vendor 0
    Scenario B: 일부만 placeholder → self 가 placeholder 만 emit → vendor skip → 0
    Scenario C: placeholder 1개 + --source 그 file → single-file mode 의 false positive
    """
    name = "test_apply_vendor_false_positive_detection"
    sources = REPO_ROOT / "ai-workflow" / "wiki" / "sources"
    if not sources.is_dir():
        return _fail(name, f"sources dir not found: {sources}")

    # backup
    backups: dict[Path, str] = {}
    for src in sorted(sources.glob("*.md")):
        if src.name in ("log.md", "_manifest.md"):
            continue
        backups[src] = src.read_text(encoding="utf-8")

    try:
        # ----- Scenario A: all sources full body (no placeholders) -----
        # main 의 L2 가 이미 full body 이므로 별도 변형 불필요. 그냥 apply.
        rc, stdout, stderr = _run(["--apply", "--emit-tool", "vendor", "--skip-lint"])
        combined = stdout + stderr
        if rc != 1:
            return _fail(
                name,
                f"Scenario A: expected exit 1 (false-positive detection), got {rc}. "
                f"combined: {combined[-300:]}",
            )
        if "false-positive" not in combined and "0 file 변형" not in combined:
            return _fail(
                name,
                f"Scenario A: exit 1 but false-positive error message missing. "
                f"combined: {combined[-300:]}",
            )

        # ----- Scenario B: 일부만 placeholder (concepts/devhub-overview.md 만) -----
        placeholder_target = sources / "devhub-overview.md"
        if placeholder_target not in backups:
            return _fail(name, f"Scenario B: {placeholder_target} not in backup — main 의 L2 page 변경됨?")
        placeholder_target.write_text(
            "---\ntype: concept\nstatus: active\n---\n\n<needs content>\n",
            encoding="utf-8",
        )
        rc, stdout, stderr = _run(["--apply", "--emit-tool", "vendor", "--skip-lint"])
        combined = stdout + stderr
        if rc != 1:
            return _fail(
                name,
                f"Scenario B: expected exit 1, got {rc}. combined: {combined[-300:]}",
            )
        if "false-positive" not in combined and "0 file 변형" not in combined:
            return _fail(
                name,
                f"Scenario B: exit 1 but false-positive error message missing. "
                f"combined: {combined[-300:]}",
            )

        # ----- Scenario C: --source single-file mode 의 false positive -----
        # placeholder 1개 + --source 그 file → self 1 + vendor 0
        placeholder_target.write_text(
            "<needs content>\n",
            encoding="utf-8",
        )
        rc, stdout, stderr = _run(
            [
                "--apply",
                "--emit-tool",
                "vendor",
                "--source",
                "concepts/devhub-overview.md",
                "--skip-lint",
            ]
        )
        combined = stdout + stderr
        if rc != 1:
            return _fail(
                name,
                f"Scenario C: expected exit 1, got {rc}. combined: {combined[-300:]}",
            )
        if "false-positive" not in combined and "0 file 변형" not in combined:
            return _fail(
                name,
                f"Scenario C: exit 1 but false-positive error message missing. "
                f"combined: {combined[-300:]}",
            )

        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")
    finally:
        # restore
        for src, content in backups.items():
            src.write_text(content, encoding="utf-8")


# ----- main -----
def main() -> int:
    print(f"[check-wiki-ingest] === wiki-ingest-from-raw.sh wrapper smoke (DevHub, v0.7.17+ + #606 P1 follow-up) ===")
    print(f"[check-wiki-ingest]   wrapper: {WRAPPER}")
    print(f"[check-wiki-ingest]   python: {sys.version.split()[0]}")
    print(f"[check-wiki-ingest]   platform: {sys.platform}")
    print("")

    test_dry_run_self()
    test_dry_run_vendor()
    test_apply_vendor_false_positive_detection()
    print("")

    # ----- summary -----
    total = PASS + FAIL
    print(f"[check-wiki-ingest] === summary ===")
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
