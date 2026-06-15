import os
import json
import shutil
import re
from pathlib import Path
import subprocess

TEST_ROOT = Path("tmp/dep-test-final-2").resolve()

def setup_test_env():
    if TEST_ROOT.exists():
        shutil.rmtree(TEST_ROOT)
    TEST_ROOT.mkdir(parents=True)

    # Python setup
    (TEST_ROOT / "requirements.txt").write_text("pytest==8.0.0\nrequests\n", encoding="utf-8")

    # Node setup
    package_json = {
        "name": "test-project",
        "version": "1.0.0",
        "devDependencies": {
            "jest": "^29.0.0"
        }
    }
    (TEST_ROOT / "package.json").write_text(json.dumps(package_json, indent=2), encoding="utf-8")

def run_bootstrap():
    cmd = [
        "python3", "workflow-source/scripts/bootstrap_workflow_kit.py",
        "--target-root", str(TEST_ROOT),
        "--project-slug", "test-proj",
        "--project-name", "Test Project",
        "--adoption-mode", "existing",
        "--update-deps"
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    print(result.stdout)
    if result.returncode != 0:
        print(result.stderr)

def verify():
    # Verify Python
    reqs = (TEST_ROOT / "requirements.txt").read_text()
    print("--- requirements.txt ---")
    print(reqs)

    # Use regex to find exact package matches
    def has_pkg(pkg, text):
        pattern = rf"(^|\s|[,]){re.escape(pkg)}([>=<#\s]|$)"
        return bool(re.search(pattern, text, re.MULTILINE | re.IGNORECASE))

    # pytest should NOT be added as a new line because it exists as pytest==8.0.0
    # But it will exist in the file.
    assert has_pkg("pytest", reqs)
    assert has_pkg("mcp_servers", reqs)
    assert has_pkg("pytest-asyncio", reqs)

    # Check that 'pytest' wasn't added as a separate line if it already existed
    # The new dependencies are added after "# Standard AI Workflow Dependencies"
    new_part = reqs.split("# Standard AI Workflow Dependencies")[1]
    assert not has_pkg("pytest", new_part)
    assert has_pkg("mcp_servers", new_part)
    assert has_pkg("pytest-asyncio", new_part)

    # Verify Node
    pkg = json.loads((TEST_ROOT / "package.json").read_text())
    print("--- package.json devDependencies ---")
    print(json.dumps(pkg.get("devDependencies"), indent=2))
    assert "@modelcontextprotocol/sdk" in pkg["devDependencies"]
    assert pkg["devDependencies"]["@modelcontextprotocol/sdk"] == "latest"

if __name__ == "__main__":
    setup_test_env()
    run_bootstrap()
    verify()
    print("\nSUCCESS: Dependency update logic verified.")
