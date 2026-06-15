#!/usr/bin/env python3
"""Smoke test the harness scaffold helper."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SCAFFOLD_SCRIPT = SOURCE_ROOT / "scripts" / "scaffold_harness.py"


def main() -> int:
    with tempfile.TemporaryDirectory() as tmpdir:
        scratch = Path(tmpdir)
        harness_name = "test-harness"
        repo_copy = scratch / "repo"
        subprocess.run(
            ["cp", "-R", str(REPO_ROOT), str(repo_copy)],
            check=True,
            cwd=scratch,
        )
        completed = subprocess.run(
            [
                sys.executable,
                str(repo_copy / "workflow-source" / "scripts" / "scaffold_harness.py"),
                "--harness-name",
                harness_name,
                "--display-name",
                "Test Harness",
                "--root-entrypoint",
                "TODO: test-entry.md",
                "--config-file",
                "TODO: test-config.json",
            ],
            cwd=repo_copy,
            capture_output=True,
            text=True,
            check=True,
        )
        payload = json.loads(completed.stdout)
        generated = payload["generated_files"]
        readme_path = Path(generated["readme"])
        overlay_spec_path = Path(generated["overlay_spec"])
        if not readme_path.exists() or not overlay_spec_path.exists():
            raise AssertionError("Harness scaffold did not create the expected files.")

        readme_text = readme_path.read_text(encoding="utf-8")
        if "Test Harness Harness Package" not in readme_text:
            raise AssertionError("Harness README did not include the display name.")
        if "read_only_transport_descriptors.json" not in readme_text:
            raise AssertionError("Harness README did not mention read-only transport descriptors.")

        overlay_spec_text = overlay_spec_path.read_text(encoding="utf-8")
        if "test-harness" not in overlay_spec_text:
            raise AssertionError("Overlay spec did not include the harness slug.")
        if "read_only_transport_descriptors.json" not in overlay_spec_text:
            raise AssertionError("Overlay spec did not mention read-only transport descriptors.")

    print("Harness scaffold smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
