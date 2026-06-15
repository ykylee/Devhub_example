#!/usr/bin/env python3
"""Smoke test the safe harness update applier."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
BOOTSTRAP_SCRIPT = SOURCE_ROOT / "scripts" / "bootstrap_workflow_kit.py"
APPLY_SCRIPT = SOURCE_ROOT / "scripts" / "apply_harness_update.py"


def run_json(cmd: list[str], cwd: Path) -> dict[str, object]:
    completed = subprocess.run(
        [sys.executable, *cmd],
        cwd=cwd,
        capture_output=True,
        text=True,
        check=True,
    )
    return json.loads(completed.stdout)


def main() -> int:
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_root = Path(tmpdir)
        source_root = tmp_root / "source-package"
        bundle_root = source_root / "bundle"
        target_root = tmp_root / "target-repo"
        bundle_root.mkdir(parents=True, exist_ok=True)
        target_root.mkdir(parents=True, exist_ok=True)

        run_json(
            [
                str(BOOTSTRAP_SCRIPT),
                "--target-root",
                str(bundle_root),
                "--project-slug",
                "sample_repo",
                "--project-name",
                "Sample Repo",
                "--adoption-mode",
                "existing",
                "--harness",
                "codex",
            ],
            REPO_ROOT,
        )

        (target_root / "AGENTS.md").write_text(
            "# AGENTS.md\n\nold agents\n",
            encoding="utf-8",
        )
        (target_root / "ai-workflow" / "project").mkdir(parents=True, exist_ok=True)
        (target_root / "ai-workflow" / "project" / "state.json").write_text(
            "{\"old\": true}\n",
            encoding="utf-8",
        )
        (target_root / ".codex").write_text("legacy flat file\n", encoding="utf-8")

        dry_run_payload = run_json(
            [
                str(APPLY_SCRIPT),
                "--source-root",
                str(source_root),
                "--target-root",
                str(target_root),
                "--timestamp",
                "20260424T120000Z",
                "--dry-run",
            ],
            REPO_ROOT,
        )
        if dry_run_payload["status"] != "dry_run":
            raise AssertionError("Expected dry-run status.")
        if "AGENTS.md" not in dry_run_payload["updated_paths"]:
            raise AssertionError("Expected AGENTS.md to be scheduled for update.")
        if ".codex" not in dry_run_payload["updated_paths"]:
            raise AssertionError("Expected legacy .codex collision to be scheduled for update.")
        if (target_root / ".ai-workflow-backups").exists():
            raise AssertionError("Dry run should not create backup directories.")
        if (target_root / ".codex").read_text(encoding="utf-8") != "legacy flat file\n":
            raise AssertionError("Dry run should not modify target files.")

        apply_payload = run_json(
            [
                str(APPLY_SCRIPT),
                "--source-root",
                str(source_root),
                "--target-root",
                str(target_root),
                "--timestamp",
                "20260424T120000Z",
            ],
            REPO_ROOT,
        )
        if apply_payload["status"] != "applied":
            raise AssertionError("Expected applied status.")
        backup_dir = Path(str(apply_payload["backup_dir"]))
        if not backup_dir.exists():
            raise AssertionError("Expected backup directory to be created.")
        if (backup_dir / "AGENTS.md").read_text(encoding="utf-8") != "# AGENTS.md\n\nold agents\n":
            raise AssertionError("Expected prior AGENTS.md to be preserved in backup.")
        if (backup_dir / ".codex").read_text(encoding="utf-8") != "legacy flat file\n":
            raise AssertionError("Expected prior flat .codex file to be preserved in backup.")
        if not (target_root / ".codex").is_dir():
            raise AssertionError("Expected target .codex path to become a directory after apply.")
        if not (target_root / ".codex" / "config.toml.example").exists():
            raise AssertionError("Expected codex config example after apply.")

        source_agents = (bundle_root / "AGENTS.md").read_text(encoding="utf-8")
        target_agents = (target_root / "AGENTS.md").read_text(encoding="utf-8")
        if source_agents != target_agents:
            raise AssertionError("Expected AGENTS.md to match source bundle after apply.")

        second_apply_payload = run_json(
            [
                str(APPLY_SCRIPT),
                "--source-root",
                str(source_root),
                "--target-root",
                str(target_root),
                "--timestamp",
                "20260424T120500Z",
            ],
            REPO_ROOT,
        )
        if second_apply_payload["updated_paths"]:
            raise AssertionError("Expected no updates on second identical apply.")
        if second_apply_payload["backup_dir"] is not None:
            raise AssertionError("Expected no backup dir when nothing changes.")

    print("Harness update apply smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
