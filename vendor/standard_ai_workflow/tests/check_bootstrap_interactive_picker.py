#!/usr/bin/env python3
"""Regression tests for the interactive `--harness` picker (v0.5.8).

Verifies:
  1. ``prompt_for_harnesses`` parses comma-separated indices correctly.
  2. ``prompt_for_harnesses`` returns the current selection on empty input.
  3. ``prompt_for_harnesses`` returns the current selection on ``q``.
  4. ``prompt_for_harnesses`` selects every harness on ``a``.
  5. ``prompt_for_harnesses`` ignores out-of-range and non-numeric tokens.
  6. ``prompt_for_harnesses`` rejects unknown harness names on input that
     doesn't match the menu indices.
  7. ``enforce_harness_selection`` raises ``SystemExit`` when --harness is
     empty and the picker could not run (non-TTY / --no-interactive).
  8. Bootstrap CLI: --no-interactive + no --harness -> SystemExit(1) with the
     documented error message (this is the regression that motivated the
     picker in the first place — silent 0 overlay files).
  9. Bootstrap CLI: --harness explicitly set works without the picker.
  10. HARNESS_DEFINITIONS (legacy) is still importable but HARNESS_SPECS is
      the single source of truth and contains all 6 entries (including
      pi-dev, which the legacy dict was missing).
"""

from __future__ import annotations

import io
import json
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPTS_DIR = SOURCE_ROOT / "scripts"
BOOTSTRAP_CLI = [sys.executable, "-m", "bootstrap_lib"]


def _run_cli(*extra: str, stdin_text: str | None = None) -> subprocess.CompletedProcess:
    """Run ``python -m bootstrap_lib`` in a subprocess with optional stdin."""
    return subprocess.run(
        [*BOOTSTRAP_CLI, *extra],
        cwd=str(SCRIPTS_DIR),
        input=stdin_text,
        capture_output=True,
        text=True,
        timeout=60,
    )


# ---------------------------------------------------------------------------
# 1-6. prompt_for_harnesses unit tests (in-process)
# ---------------------------------------------------------------------------
def test_picker_parses_indices() -> None:
    from bootstrap_lib.__main__ import prompt_for_harnesses
    from bootstrap_lib.harnesses import SUPPORTED_HARNESSES

    out = io.StringIO()
    picker_input = io.StringIO("1,3\n")
    result = prompt_for_harnesses(
        input_stream=picker_input,
        output_stream=out,
    )
    # Index 1 = first catalog entry, index 3 = third
    expected = sorted({SUPPORTED_HARNESSES[0], SUPPORTED_HARNESSES[2]})
    if result != expected:
        raise AssertionError(f"Expected {expected}, got {result}")


def test_picker_empty_input_keeps_selection() -> None:
    from bootstrap_lib.__main__ import prompt_for_harnesses
    from bootstrap_lib.harnesses import HARNESS_SPECS

    out = io.StringIO()
    picker_input = io.StringIO("\n")
    base = [next(iter(HARNESS_SPECS))]
    result = prompt_for_harnesses(
        base,
        input_stream=picker_input,
        output_stream=out,
    )
    if result != sorted(base):
        raise AssertionError(f"Empty input should keep selection, got {result}")


def test_picker_q_keeps_selection() -> None:
    from bootstrap_lib.__main__ import prompt_for_harnesses
    from bootstrap_lib.harnesses import HARNESS_SPECS

    out = io.StringIO()
    picker_input = io.StringIO("q\n")
    base = [next(iter(HARNESS_SPECS))]
    result = prompt_for_harnesses(
        base,
        input_stream=picker_input,
        output_stream=out,
    )
    if result != sorted(base):
        raise AssertionError(f"'q' should keep selection, got {result}")


def test_picker_a_selects_all() -> None:
    from bootstrap_lib.__main__ import prompt_for_harnesses
    from bootstrap_lib.harnesses import SUPPORTED_HARNESSES

    out = io.StringIO()
    picker_input = io.StringIO("a\n")
    result = prompt_for_harnesses(
        input_stream=picker_input,
        output_stream=out,
    )
    if set(result) != set(SUPPORTED_HARNESSES):
        raise AssertionError(
            f"'a' should select all 6 harnesses, got {result}, "
            f"expected {set(SUPPORTED_HARNESSES)}"
        )
    if len(result) != len(SUPPORTED_HARNESSES):
        raise AssertionError(
            f"'a' should select exactly {len(SUPPORTED_HARNESSES)} unique harnesses, got {result}"
        )


def test_picker_ignores_garbage() -> None:
    from bootstrap_lib.__main__ import prompt_for_harnesses
    from bootstrap_lib.harnesses import SUPPORTED_HARNESSES

    out = io.StringIO()
    # 99 is out of range, abc is non-numeric, 1 is valid
    picker_input = io.StringIO("99,abc,1\n")
    result = prompt_for_harnesses(
        input_stream=picker_input,
        output_stream=out,
    )
    if result != [SUPPORTED_HARNESSES[0]]:
        raise AssertionError(
            f"Garbage tokens should be ignored, valid 1 kept, got {result}"
        )


def test_enforce_raises_on_empty() -> None:
    """enforce_harness_selection must fail fast in non-picker contexts."""
    import argparse

    from bootstrap_lib.__main__ import enforce_harness_selection

    args = argparse.Namespace(harnesses=[], no_interactive=True)
    try:
        enforce_harness_selection(args)
    except SystemExit as exc:
        if "--harness is required" not in str(exc):
            raise AssertionError(
                f"SystemExit message should mention --harness, got: {exc}"
            )
    else:
        raise AssertionError("enforce_harness_selection should SystemExit on empty")


# ---------------------------------------------------------------------------
# 7-8. CLI subprocess tests
# ---------------------------------------------------------------------------
def test_cli_no_interactive_no_harness_fails() -> None:
    """Regression: --no-interactive + missing --harness must fail loudly."""
    with tempfile.TemporaryDirectory() as tmpdir:
        target = Path(tmpdir) / "sample"
        target.mkdir()
        completed = _run_cli(
            "--project-slug", "saw-picker-ci",
            "--project-name", "Picker CI",
            "--target-root", str(target),
            "--no-interactive",
        )
    if completed.returncode == 0:
        raise AssertionError(
            "Expected non-zero exit when --no-interactive + no --harness, got 0. "
            f"stdout={completed.stdout!r}"
        )
    if "--harness is required" not in completed.stderr:
        raise AssertionError(
            f"Stderr should mention --harness requirement, got: {completed.stderr!r}"
        )


def test_cli_harness_explicit_works() -> None:
    """--harness codex + --no-interactive should bootstrap without prompting."""
    with tempfile.TemporaryDirectory() as tmpdir:
        target = Path(tmpdir) / "sample"
        target.mkdir()
        completed = _run_cli(
            "--project-slug", "saw-picker-explicit",
            "--project-name", "Picker Explicit",
            "--target-root", str(target),
            "--harness", "codex",
            "--no-interactive",
            "--dry-run",
        )
    if completed.returncode != 0:
        raise AssertionError(
            f"Explicit --harness should succeed, got rc={completed.returncode}, "
            f"stderr={completed.stderr!r}"
        )
    payload = json.loads(completed.stdout)
    harnesses = payload.get("harnesses", [])
    if harnesses != ["codex"]:
        raise AssertionError(f"Expected harnesses=['codex'], got {harnesses}")


# ---------------------------------------------------------------------------
# 9-10. SoT sync checks
# ---------------------------------------------------------------------------
def test_harness_specs_is_complete_sot() -> None:
    """HARNESS_SPECS must contain all 6 harnesses including pi-dev.

    The legacy HARNESS_DEFINITIONS dict (kept for back-compat) does not have
    pi-dev. New code must use HARNESS_SPECS so pi-dev is selectable.
    """
    from bootstrap_lib.harnesses import HARNESS_SPECS, SUPPORTED_HARNESSES
    from bootstrap_lib.__main__ import HARNESS_DEFINITIONS

    spec_keys = set(HARNESS_SPECS)
    supported = set(SUPPORTED_HARNESSES)
    if spec_keys != supported:
        raise AssertionError(
            f"HARNESS_SPECS must equal SUPPORTED_HARNESSES. "
            f"missing in SPECS: {supported - spec_keys}; "
            f"extra in SPECS: {spec_keys - supported}"
        )
    if "pi-dev" not in HARNESS_SPECS:
        raise AssertionError("pi-dev must be in HARNESS_SPECS")
    # Legacy dict should be present (back-compat shim) but is documented as
    # missing pi-dev. We verify the documented gap so future fixes are loud.
    if "pi-dev" in HARNESS_DEFINITIONS:
        raise AssertionError(
            "HARNESS_DEFINITIONS (legacy) was documented to miss pi-dev; if "
            "you re-added it, update workflow_harness_distribution.md and "
            "remove this assertion."
        )


def test_registry_consistency_check() -> None:
    """Importing renderers must trigger the SoT consistency check.

    If HARNESS_SPECS and HARNESS_FILE_BUILDERS drift, the import will raise
    ``RuntimeError`` (registered at the bottom of renderers.py). This test
    exercises the import path so the failure mode is observable.
    """
    from bootstrap_lib.harnesses import renderers  # noqa: F401

    # If the consistency check ran during import, the test passes. The fact
    # that we got here without RuntimeError is the assertion.
    if not hasattr(renderers, "_verify_harness_registry_consistency"):
        raise AssertionError(
            "renderers._verify_harness_registry_consistency should be defined"
        )


def main() -> int:
    test_picker_parses_indices()
    test_picker_empty_input_keeps_selection()
    test_picker_q_keeps_selection()
    test_picker_a_selects_all()
    test_picker_ignores_garbage()
    test_enforce_raises_on_empty()
    test_cli_no_interactive_no_harness_fails()
    test_cli_harness_explicit_works()
    test_harness_specs_is_complete_sot()
    test_registry_consistency_check()
    print("Interactive picker regression suite passed (10 checks).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
