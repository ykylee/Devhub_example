#!/usr/bin/env python3
"""Smoke test harness package export."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import zipfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION

EXPORT_SCRIPT = SOURCE_ROOT / "scripts" / "export_harness_package.py"
EXPECTED_VERSION = TOOL_VERSION


def main() -> int:
    with tempfile.TemporaryDirectory() as tmpdir:
        output_root = Path(tmpdir) / "dist"
        completed = subprocess.run(
            [
                sys.executable,
                str(EXPORT_SCRIPT),
                "--harness",
                "codex",
                "--harness",
                "opencode",
                "--output-dir",
                str(output_root),
            ],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=True,
        )
        payload = json.loads(completed.stdout)
        exports = payload["exports"]
        if len(exports) != 2:
            raise AssertionError("Expected two exported harness packages.")
        if payload["package_version"] != EXPECTED_VERSION:
            raise AssertionError("Top-level export payload should report the package version.")
        for export in exports:
            package_root = Path(export["package_root"])
            manifest_path = Path(export["manifest_path"])
            archive_path = Path(export["archive_path"])
            if not package_root.exists():
                raise AssertionError(f"Missing package root: {package_root}")
            if not manifest_path.exists():
                raise AssertionError(f"Missing manifest: {manifest_path}")
            if not archive_path.exists():
                raise AssertionError(f"Missing archive: {archive_path}")

            manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
            if not manifest["included_files"]:
                raise AssertionError("Manifest should include exported files.")
            if manifest["global_snippet_files"]:
                raise AssertionError("Minimal deployment export should exclude global snippet files by default.")
            if export["package_version"] != EXPECTED_VERSION:
                raise AssertionError("Each export result should carry the package version.")
            if manifest["package_version"] != EXPECTED_VERSION:
                raise AssertionError("Manifest should record the package version.")
            if manifest["optimization_profile"] != "agent_runtime_minimal":
                raise AssertionError("Manifest should record the deployment optimization profile.")
            if manifest["release_focus"] != "workflow_skill_onboarding":
                raise AssertionError("Manifest should record the current workflow/skill release focus.")
            if "bundle/ai-workflow/core/workflow_adoption_entrypoints.md" not in manifest["recommended_entrypoints"]:
                raise AssertionError("Manifest should include workflow adoption entrypoint guidance.")
            if "developer_source_docs" not in manifest["excluded_by_default"]:
                raise AssertionError("Manifest should explain excluded deployment-only context.")
            if "official_mcp_server_default_adoption" not in manifest["deferred_release_items"]:
                raise AssertionError("Manifest should record deferred MCP activation work.")
            included = set(manifest["included_files"])
            if "bundle/AGENTS.md" not in included:
                raise AssertionError("Every harness export should include AGENTS.md as the shared top-level instruction entrypoint.")
            if "bundle/ai-workflow/core/global_workflow_standard.md" not in included:
                raise AssertionError("Manifest should include global workflow standard runtime docs.")
            if "bundle/ai-workflow/core/workflow_skill_catalog.md" not in included:
                raise AssertionError("Manifest should include workflow skill catalog runtime docs.")
            if "bundle/ai-workflow/core/workflow_adoption_entrypoints.md" not in included:
                raise AssertionError("Manifest should include workflow adoption entrypoint runtime docs.")
            if "bundle/ai-workflow/memory/active/state.json" not in included:
                raise AssertionError("Manifest should include the workflow state cache.")
            if "bundle/ai-workflow/memory/active/PROJECT_PROFILE.md" not in included:
                raise AssertionError("Manifest should include the workflow project profile.")
            if export["harness"] == "codex" and "bundle/.codex/config.toml.example" not in included:
                raise AssertionError("Codex export should preserve the .codex config example path.")
            if export["harness"] == "opencode" and "bundle/.opencode/agents/workflow-orchestrator.md" not in included:
                raise AssertionError("OpenCode export should preserve the .opencode agent path.")
            if "bundle/source-docs/core/workflow_skill_catalog.md" in included:
                raise AssertionError("Minimal deployment export should not include source-docs by default.")
            if "bundle/source-docs/schemas/read_only_transport_descriptors.json" in included:
                raise AssertionError("Minimal deployment export should not include draft MCP assets by default.")
            if "bundle/global-snippets/codex/config.toml.snippet" in included:
                raise AssertionError("Minimal deployment export should not include global snippets by default.")

            expected_root = (output_root / "harnesses" / export["harness"] / EXPECTED_VERSION).resolve()
            if package_root != expected_root:
                raise AssertionError("Versioned package root path is incorrect.")

            with zipfile.ZipFile(archive_path) as archive:
                names = set(archive.namelist())
            if "manifest.json" not in names:
                raise AssertionError("Archive should include manifest.json at the package root.")
            if "bundle/AGENTS.md" not in names:
                raise AssertionError("Archive should include AGENTS.md for every harness package.")
            if "bundle/ai-workflow/README.md" not in names:
                raise AssertionError("Archive should include the runtime workflow README.")
            if "bundle/ai-workflow/memory/active/state.json" not in names:
                raise AssertionError("Archive should include the workflow state cache.")
            if "bundle/ai-workflow/memory/active/PROJECT_PROFILE.md" not in names:
                raise AssertionError("Archive should include the workflow project profile.")
            if "PACKAGE_CONTENTS.md" not in names:
                raise AssertionError("Archive should include package composition guidance.")
            if "APPLY_GUIDE.md" not in names:
                raise AssertionError("Archive should include package apply guidance.")
            if any(name.startswith("bundle/source-docs/") for name in names):
                raise AssertionError("Archive should exclude source-docs in the default deployment profile.")

    print("Harness package export smoke check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
