#!/usr/bin/env python3
"""Regression test for the smart-update policy added in v0.5.10.1.

Covers the file-write decision used by both ``bootstrap_lib`` and
``apply_workflow_upgrade``: when the destination already exists, the
write must be skipped (ignored) or applied (updated) based on a
VERSION-marker comparison, with content-hash comparison as the
fallback. Preserved paths (user data under ``ai-workflow/memory/``
and friends) are never overwritten, even with ``--force``.

Run as a script. Exits 0 on success, raises AssertionError otherwise.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCRIPTS_ROOT = SOURCE_ROOT / "scripts"


def _run_bootstrap_lib(args: list[str], env: dict[str, str] | None = None) -> dict:
    """Run ``python3 -m bootstrap_lib`` and return the parsed manifest."""
    full_env = os.environ.copy()
    full_env["PYTHONPATH"] = (
        f"{SOURCE_ROOT}{os.pathsep}{SCRIPTS_ROOT}"
        + os.pathsep
        + full_env.get("PYTHONPATH", "")
    )
    if env:
        full_env.update(env)
    completed = subprocess.run(
        [sys.executable, "-m", "bootstrap_lib", *args],
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        env=full_env,
        check=True,
    )
    return json.loads(completed.stdout)


def _bucket(actions: dict, action_name: str) -> list[dict]:
    return list(actions.get(action_name, []))


def test_bootstrap_first_run_creates_files() -> None:
    with tempfile.TemporaryDirectory() as tmp:
        target = Path(tmp)
        manifest = _run_bootstrap_lib(
            [
                "--target-root", str(target),
                "--project-slug", "test",
                "--project-name", "Test",
                "--harness", "opencode",
                "--no-interactive",
                "--adoption-mode", "new",
            ]
        )
        actions = manifest.get("file_actions", {})
        created = _bucket(actions, "created")
        assert len(created) > 0, "first run should CREATE files"
        # All non-preserved harness files should appear in created.
        rels = {a["rel"] for a in created}
        assert "AGENTS.md" in rels
        assert "opencode.json" in rels
        # Verify the marker was stamped on a markdown file.
        agents = (target / "AGENTS.md").read_text(encoding="utf-8")
        assert "standard-ai-workflow-kit: v0.5.10.1-beta" in agents


def test_bootstrap_second_run_ignores_matching_files() -> None:
    with tempfile.TemporaryDirectory() as tmp:
        target = Path(tmp)
        common = [
            "--target-root", str(target),
            "--project-slug", "test",
            "--project-name", "Test",
            "--harness", "opencode",
            "--no-interactive",
            "--adoption-mode", "new",
        ]
        _run_bootstrap_lib(common)

        manifest2 = _run_bootstrap_lib(common)
        actions = manifest2["file_actions"]
        ignored = _bucket(actions, "ignored")
        assert len(ignored) > 0, "second run should IGNORE files with matching marker + hash"
        created2 = _bucket(actions, "created")
        assert len(created2) == 0, (
            f"second run should not CREATE; got {created2}"
        )


def test_bootstrap_preserves_user_data_with_force() -> None:
    with tempfile.TemporaryDirectory() as tmp:
        target = Path(tmp)
        # Pre-create a user file under preserved path.
        memory_dir = target / "ai-workflow" / "memory"
        memory_dir.mkdir(parents=True)
        user_file = memory_dir / "user_notes.md"
        user_file.write_text("# My own notes\n", encoding="utf-8")

        _run_bootstrap_lib(
            [
                "--target-root", str(target),
                "--project-slug", "test",
                "--project-name", "Test",
                "--harness", "opencode",
                "--no-interactive",
                "--adoption-mode", "new",
                "--force",
            ]
        )

        # File should still exist, untouched.
        assert user_file.exists()
        content = user_file.read_text(encoding="utf-8")
        assert content == "# My own notes\n", (
            f"PRESERVE path was overwritten: {content!r}"
        )


def test_apply_upgrade_smart_update() -> None:
    """apply_workflow_upgrade should use the same smart-update policy."""
    with tempfile.TemporaryDirectory() as tmp:
        bundle = Path(tmp) / "bundle"
        target = Path(tmp) / "target"
        (bundle / "ai-workflow").mkdir(parents=True)
        (bundle / "ai-workflow" / "VERSION").write_text("v0.5.10.1-beta\n", encoding="utf-8")
        (bundle / "ai-workflow" / "PROFILE.md").write_text(
            "<!-- standard-ai-workflow-kit: v0.5.10.1-beta -->\n# New profile\n",
            encoding="utf-8",
        )
        target.mkdir()

        env = {"PYTHONPATH": f"{SOURCE_ROOT}{os.pathsep}{SCRIPTS_ROOT}"}
        completed = subprocess.run(
            [
                sys.executable,
                str(SCRIPTS_ROOT / "apply_workflow_upgrade.py"),
                "--source-root", str(bundle),
                "--target-root", str(target),
            ],
            cwd=str(REPO_ROOT),
            env={**env, **os.environ},
            capture_output=True,
            text=True,
            check=True,
        )
        assert "Created:" in completed.stdout
        # First run creates both VERSION and PROFILE.md.
        assert (target / "ai-workflow" / "VERSION").exists()
        assert (target / "ai-workflow" / "PROFILE.md").exists()

        # Second run: should be all IGNORED via the kit-version short-circuit.
        completed2 = subprocess.run(
            [
                sys.executable,
                str(SCRIPTS_ROOT / "apply_workflow_upgrade.py"),
                "--source-root", str(bundle),
                "--target-root", str(target),
                "--dry-run",
            ],
            cwd=str(REPO_ROOT),
            env={**env, **os.environ},
            capture_output=True,
            text=True,
            check=True,
        )
        assert "Ignored:" in completed2.stdout
        assert "already up to date" in completed2.stdout


def test_upgrade_diff_unit() -> None:
    """Quick unit smoke of the underlying decision function."""
    sys.path.insert(0, str(SOURCE_ROOT))
    from workflow_kit.upgrade_diff import Action, decide_action  # type: ignore[import-not-found]

    dec = decide_action(
        src_text="<!-- standard-ai-workflow-kit: v0.5.10.1-beta -->\nNEW",
        dst_text="<!-- standard-ai-workflow-kit: v0.5.0-beta -->\nOLD",
    )
    assert dec.action == Action.UPDATED, dec

    dec = decide_action(
        src_text="<!-- standard-ai-workflow-kit: v0.5.10.1-beta -->\nNEW",
        dst_text="<!-- standard-ai-workflow-kit: v0.5.10.1-beta -->\nNEW",
    )
    assert dec.action == Action.IGNORED, dec

    dec = decide_action(src_text="x", dst_text="y", is_preserved_path=True, force=True)
    assert dec.action == Action.PRESERVED, dec


def test_bootstrap_wheel_install_graceful_when_no_source() -> None:
    """Verify graceful skip when SOURCE_ROOT is unresolvable (wheel install
    without bundled core-docs data) and that global_snippet_sources accepts
    None without crashing. Also sanity-checks dev-install still copies
    core docs.
    """
    sys.path.insert(0, str(SOURCE_ROOT))
    sys.path.insert(0, str(SCRIPTS_ROOT))
    import bootstrap_lib.__main__ as BM  # type: ignore[import-not-found]
    from bootstrap_lib.discovery import global_snippet_sources  # type: ignore[import-not-found]

    assert global_snippet_sources(None) == {}, (
        "global_snippet_sources must return empty dict when source_root is None"
    )

    real_source_root = BM.SOURCE_ROOT
    BM.SOURCE_ROOT = None
    try:
        with tempfile.TemporaryDirectory() as tmp:
            target = Path(tmp)
            _make_args(target, copy_core_docs=True, force=True)
            manifest = _run_bootstrap_lib(
                [
                    "--target-root", str(target),
                    "--project-slug", "test",
                    "--project-name", "Test",
                    "--harness", "opencode",
                    "--no-interactive",
                    "--adoption-mode", "new",
                    "--copy-core-docs",
                ],
                env={"BOOTSTRAP_LIB_NO_SOURCE": "1"},
            )
            assert manifest.get("copied_core_docs") == [], (
                f"wheel install with SOURCE_ROOT=None should skip core docs; got {manifest.get('copied_core_docs')!r}"
            )
            warnings = manifest.get("warnings", [])
            assert any("core docs are not bundled" in w for w in warnings), (
                f"expected a manifest warning about missing core docs; got {warnings!r}"
            )
    finally:
        BM.SOURCE_ROOT = real_source_root

    assert real_source_root is not None, (
        "sanity: the dev-install SOURCE_ROOT should resolve"
    )

    with tempfile.TemporaryDirectory() as tmp:
        target = Path(tmp)
        manifest = _run_bootstrap_lib(
            [
                "--target-root", str(target),
                "--project-slug", "test",
                "--project-name", "Test",
                "--harness", "opencode",
                "--no-interactive",
                "--adoption-mode", "new",
                "--copy-core-docs",
            ]
        )
        assert len(manifest.get("copied_core_docs", [])) > 0, (
            "dev-install should copy core docs"
        )


def _make_args(target_root: Path, copy_core_docs: bool = True, force: bool = False):
    return argparse.Namespace(
        target_root=str(target_root),
        kit_dir="ai-workflow",
        project_slug="test",
        project_name="Test",
        project_intro="Test",
        today="2026-06-09",
        adoption_mode="new",
        copy_core_docs=copy_core_docs,
        force=force,
        update_deps=False,
        enable_mcp=False,
        mcp_bridge="jsonrpc-bridge",
        harnesses=["opencode"],
        no_interactive=True,
        dry_run=False,
    )


def main() -> int:
    test_upgrade_diff_unit()
    test_bootstrap_first_run_creates_files()
    test_bootstrap_second_run_ignores_matching_files()
    test_bootstrap_preserves_user_data_with_force()
    test_apply_upgrade_smart_update()
    test_bootstrap_wheel_install_graceful_when_no_source()
    print(
        "Smart update regression check passed "
        "(6 checks: unit, first-run CREATE, second-run IGNORE, "
        "PRESERVE+force, apply_upgrade round-trip, wheel-install graceful)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
