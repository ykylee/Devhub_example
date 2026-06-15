#!/usr/bin/env python3
import json
import subprocess
import sys
import tempfile
from pathlib import Path

def test_relative_paths_in_state():
    repo_root = Path(__file__).resolve().parents[2]
    sample_root = repo_root / "workflow-source" / "examples" / "acme_delivery_platform"

    with tempfile.TemporaryDirectory() as tmpdir:
        output_path = Path(tmpdir) / "state.json"

        cmd = [
            sys.executable, "workflow-source/scripts/generate_workflow_state.py",
            "--project-profile-path", str(sample_root / "PROJECT_PROFILE.md"),
            "--session-handoff-path", str(sample_root / "session_handoff.md"),
            "--work-backlog-index-path", str(sample_root / "work_backlog.md"),
            "--latest-backlog-path", str(sample_root / "backlog" / "2026-04-18.md"),
            "--output-path", str(output_path),
            "--workspace-root", str(sample_root),
            "--generated-at", "2026-05-01",
        ]

        print(f"Running: {' '.join(cmd)}")
        result = subprocess.run(cmd, capture_output=True, text=True, cwd=str(repo_root))
        if result.returncode != 0:
            print(f"Error: {result.stderr}")
            exit(1)

        state = json.loads(output_path.read_text(encoding="utf-8"))

        print("\nVerifying source_of_truth paths...")
        sot = state.get("source_of_truth", {})
        for key, val in sot.items():
            if val:
                print(f"  {key}: {val}")
                if val.startswith("/") or (":" in val and "\\" in val):
                    print(f"Error: {key} is an absolute path: {val}")
                    exit(1)
                else:
                    print(f"OK: {key} is relative")

        print("\nVerifying next_documents paths...")
        next_docs = state.get("next_documents", [])
        for doc in next_docs:
            print(f"  {doc}")
            if doc.startswith("/") or (":" in doc and "\\" in doc):
                print(f"Error: next_document is an absolute path: {doc}")
                exit(1)
            else:
                print("OK: relative")

    print("\nPath relative-ness check passed!")

if __name__ == "__main__":
    test_relative_paths_in_state()
