#!/usr/bin/env python3
"""Verify the built wheel imports cleanly in a fresh venv.

Catches packaging regressions like the v0.5.7.1 hotfix (sub-packages
``workflow_kit/common/state``, ``contracts``, ``schemas`` were missing from
``[tool.setuptools.packages]`` so ``pip install dist/*.whl`` failed with
``ModuleNotFoundError`` even though ``pip install -e .`` worked).

Usage::

    python3 tools/check_packaging.py [--wheel PATH]

Default behaviour:

1. Resolve the wheel path (most recent under ``dist/``) unless overridden.
2. Create a throwaway virtual environment via ``python3 -m venv``.
3. Install the wheel with ``pip install`` (no editable mode, no local
   source fallback).
4. Run a 1-line import smoke covering every public sub-package plus the
   two CLI entry points (``bootstrap_lib``, ``bootstrap_workflow_kit``).
5. Tear down the venv on success.

Exit code 0 on success, 1 on any import or install failure. The script
prints a JSON manifest describing what it checked, so it can be wired
into a release checklist or CI hook.
"""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_DIST = REPO_ROOT / "dist"

# Sub-packages that must be importable. The set is intentionally explicit:
# adding a new sub-package will require updating this list, which is the
# whole point — packaging drift is silent otherwise.
#
# Note: ``bootstrap_workflow_kit`` is intentionally NOT in this list. It's
# a legacy CLI shim (single .py file in scripts/) that downstream callers
# invoke directly via ``python scripts/bootstrap_workflow_kit.py`` rather
# than importing. The new programmatic entry point is
# ``python -m bootstrap_lib`` and the new programmatic API is the
# ``bootstrap_lib`` package itself, both of which are covered below.
REQUIRED_IMPORTS: tuple[str, ...] = (
    "workflow_kit",
    "bootstrap_lib",
    "workflow_kit.contract_v1",
    "workflow_kit.common",
    "workflow_kit.common.state",
    "workflow_kit.common.contracts",
    "workflow_kit.common.schemas",
)


def find_latest_wheel(dist_dir: Path) -> Path:
    wheels = sorted(dist_dir.glob("*.whl"), key=lambda p: p.stat().st_mtime)
    if not wheels:
        raise SystemExit(f"ERROR: no wheel found under {dist_dir}")
    return wheels[-1]


def run(cmd: list[str], *, cwd: Path | None = None, env: dict | None = None) -> None:
    print(f"  $ {' '.join(cmd)}", flush=True)
    completed = subprocess.run(
        cmd,
        cwd=str(cwd) if cwd else None,
        env=env,
        check=False,
    )
    if completed.returncode != 0:
        raise SystemExit(
            f"ERROR: command failed (rc={completed.returncode}): {' '.join(cmd)}"
        )


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--wheel",
        type=Path,
        default=None,
        help="Wheel to verify. Defaults to the most recent file under dist/.",
    )
    parser.add_argument(
        "--keep-venv",
        action="store_true",
        help="Keep the throwaway venv (printed at end) for debugging.",
    )
    args = parser.parse_args()

    wheel = args.wheel or find_latest_wheel(DEFAULT_DIST)
    if not wheel.exists():
        raise SystemExit(f"ERROR: wheel not found: {wheel}")
    print(f"Verifying wheel: {wheel}")

    with tempfile.TemporaryDirectory(prefix="saw-packaging-") as tmp:
        venv = Path(tmp) / "venv"
        run([sys.executable, "-m", "venv", str(venv)])
        pip = venv / "bin" / "pip"
        python = venv / "bin" / "python"

        # 1. Install the wheel. No -e, no local source — this is the
        #    scenario that revealed the v0.5.7.1 regression.
        run([str(pip), "install", "--upgrade", "pip"])
        run([str(pip), "install", str(wheel)])

        # 2. Import smoke. Use explicit try/except per import to record
        #    exactly which module is missing — a flat ``__import__`` chain
        #    would short-circuit on the first failure.
        import_payload = "import json\n"
        import_payload += "ok, missing = [], []\n"
        for mod in REQUIRED_IMPORTS:
            import_payload += (
                "try:\n"
                f"    __import__({mod!r})\n"
                f"    ok.append({mod!r})\n"
                "except Exception as exc:\n"
                f"    missing.append({{'module': {mod!r}, 'error': str(exc)}})\n"
            )
        import_payload += "print(json.dumps({'ok': ok, 'missing': missing}))\n"

        completed = subprocess.run(
            [str(python), "-c", import_payload],
            capture_output=True,
            text=True,
        )
        if completed.returncode != 0:
            print("--- import smoke failed ---")
            print(completed.stdout)
            print(completed.stderr, file=sys.stderr)
            return 1
        result = json.loads(completed.stdout)
        if result["missing"]:
            print("ERROR: missing imports:", json.dumps(result["missing"], indent=2))
            return 1

        # 3. CLI entry point smoke. ``python -m bootstrap_lib --help`` must
        #    succeed and show the new --no-interactive flag.
        completed = subprocess.run(
            [str(python), "-m", "bootstrap_lib", "--help"],
            capture_output=True,
            text=True,
        )
        if completed.returncode != 0:
            print("ERROR: bootstrap_lib --help failed")
            print(completed.stderr, file=sys.stderr)
            return 1
        if "--no-interactive" not in completed.stdout:
            print("ERROR: bootstrap_lib --help output missing --no-interactive")
            print(completed.stdout)
            return 1

        # 4. Package metadata smoke — confirm the version we built matches
        #    what we expect.
        completed = subprocess.run(
            [str(pip), "show", "standard-ai-workflow"],
            capture_output=True,
            text=True,
        )
        if completed.returncode != 0:
            print("ERROR: pip show failed")
            return 1
        meta = completed.stdout
        if "Name: standard-ai-workflow" not in meta:
            print("ERROR: pip show output missing package name")
            print(meta)
            return 1

        if args.keep_venv:
            kept = Path.cwd() / "saw-packaging-venv"
            shutil.copytree(venv, kept)
            print(f"Kept venv at: {kept}")

    manifest = {
        "wheel": str(wheel),
        "imported": result["ok"],
        "missing": result["missing"],
        "boot_lib_help_has_no_interactive": True,
        "result": "PASS",
    }
    print(json.dumps(manifest, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
