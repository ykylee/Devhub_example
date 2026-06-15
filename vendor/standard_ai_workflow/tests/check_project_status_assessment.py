#!/usr/bin/env python3
import json
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]

def run_assessment(project_root: Path, extra_args: list[str] = []) -> dict:
    cmd = [
        sys.executable,
        str(REPO_ROOT / "workflow-source/skills/project-status-assessment/scripts/run_project_status_assessment.py"),
        "--project-root", str(project_root),
        "--json"
    ] + extra_args
    result = subprocess.run(cmd, capture_output=True, text=True)
    try:
        return json.loads(result.stdout)
    except:
        print("STDOUT:", result.stdout)
        print("STDERR:", result.stderr)
        return {"status": "error", "error": "Failed to parse JSON output"}

def test_assessment_python():
    print("Testing assessment for Python project...")
    with tempfile.TemporaryDirectory() as tmp_dir:
        root = Path(tmp_dir)
        (root / "src").mkdir()
        (root / "tests").mkdir()
        (root / "requirements.txt").write_text("pytest")
        (root / "README.md").write_text("# Test Project")

        result = run_assessment(root)
        if result["status"] != "ok":
            print(f"FAILED. Result: {json.dumps(result, indent=2, ensure_ascii=False)}")
        assert result["status"] == "ok"
        assessment = result["assessment"]
        assert assessment["primary_stack"] == "python"
        assert "src" in assessment["dirs"]["source"]
        assert "tests" in assessment["dirs"]["test"]
        print("✅ Python assessment successful.")

def test_assessment_node():
    print("Testing assessment for Node.js project...")
    with tempfile.TemporaryDirectory() as tmp_dir:
        root = Path(tmp_dir)
        (root / "package.json").write_text(json.dumps({
            "name": "test-node",
            "scripts": {"test": "jest"}
        }))
        (root / "docs").mkdir()

        result = run_assessment(root)
        assert result["status"] == "ok"
        assessment = result["assessment"]
        assert assessment["primary_stack"] == "node"
        assert "docs" in assessment["dirs"]["docs"]
        assert assessment["package_scripts"]["test"] == "jest"
        print("✅ Node.js assessment successful.")

def test_assessment_apply():
    print("Testing assessment with --apply...")
    with tempfile.TemporaryDirectory() as tmp_dir:
        root = Path(tmp_dir)
        (root / "requirements.txt").write_text("pytest")

        output_path = Path(root / "custom_assessment.md").resolve()
        result = run_assessment(root, ["--apply", "--output-path", str(output_path)])

        if result["status"] != "ok":
            print(f"FAILED. Result: {json.dumps(result, indent=2, ensure_ascii=False)}")
        assert result["status"] == "ok"

        resolved_written = [str(Path(p).resolve()) for p in result["written_paths"]]
        assert str(output_path) in resolved_written, f"Expected {output_path} in {resolved_written}"
        assert output_path.exists()
        content = output_path.read_text()
        assert "# Repository Assessment" in content
        assert "추정 기본 스택: `python`" in content
        print("✅ Apply mode successful.")

if __name__ == "__main__":
    try:
        test_assessment_python()
        test_assessment_node()
        test_assessment_apply()
        print("\n🎉 All project-status-assessment smoke tests passed!")
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        sys.exit(1)
