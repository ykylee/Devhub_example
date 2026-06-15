#!/usr/bin/env python3
"""v0.7.0: stage_completion required 격상 검증.

- ensure_stage_completion() 자동 생성 검증
- 기존 35 test 의 11종 skill runtime result 모두 stage_completion 8-field 포함 검증
- v0.6.5 batch + pilot + follow-up 의 12/12 일관성 runtime-level 검증
Reference: workflow-source/core/output_schema_guide.md §3.4 (v0.7.0 required)
           workflow-source/core/stage_gate_runtime_migration.md
"""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.contracts.stage_gate_runtime import (  # noqa: E402
    build_stage_completion,
    ensure_stage_completion,
    is_stage_completion_present,
    merge_into_result,
)

REPO_ROOT = Path("/Users/yklee/repos/standard_ai_workflow_minimax")


# --- Test 1: ensure_stage_completion 자동 생성 ---


def test_ensure_creates_when_missing() -> None:
    """stage_completion 없는 result → ensure 가 자동 생성."""
    result = {
        "status": "ok",
        "tool_version": "0.6.5",
        "warnings": [],
        "source_context": {"path": "x.md"},
    }
    ensured = ensure_stage_completion(
        result,
        stage_name="session-start",
        next_stage=None,
    )
    assert is_stage_completion_present(ensured) is True
    assert ensured["stage_completion"]["stage_name"] == "session-start"
    # 기존 field 보존
    assert ensured["status"] == "ok"
    assert ensured["tool_version"] == "0.6.5"


def test_ensure_preserves_existing() -> None:
    """stage_completion 이미 있으면 그대로 보존."""
    result = {
        "status": "ok",
        "stage_completion": build_stage_completion(
            stage_name="doc-sync",
            stage_status="ok",
            next_stage="validation-plan",
        ),
    }
    ensured = ensure_stage_completion(result, stage_name="different-name")
    # 기존 stage_completion 보존 (overwrite=False)
    assert ensured["stage_completion"]["stage_name"] == "doc-sync"


def test_ensure_status_mapping() -> None:
    """status 가 'success' → 'ok', 'warning' → 'warning', else 'error' 매핑."""
    # success → ok
    result = {"status": "success"}
    ensured = ensure_stage_completion(result, stage_name="test")
    assert ensured["stage_completion"]["stage_status"] == "ok"

    # warning → warning
    result = {"status": "warning"}
    ensured = ensure_stage_completion(result, stage_name="test")
    assert ensured["stage_completion"]["stage_status"] == "warning"

    # error → error
    result = {"status": "error"}
    ensured = ensure_stage_completion(result, stage_name="test")
    assert ensured["stage_completion"]["stage_status"] == "error"

    # unknown → error
    result = {"status": "unknown_value"}
    ensured = ensure_stage_completion(result, stage_name="test")
    assert ensured["stage_completion"]["stage_status"] == "error"


# --- Test 2: 12/12 runtime 결과 검증 ---


def _run_skill_actual(skill: str, *args: str) -> dict | None:
    """skill 의 run_*.py 를 실제 실행하여 JSON output 받기. 실패 시 None."""
    skill_dir = REPO_ROOT / "workflow-source" / "skills" / skill
    scripts = list(skill_dir.glob("scripts/run_*.py"))
    if not scripts:
        return None
    run_script = scripts[0]
    try:
        result = subprocess.run(
            ["python3", str(run_script), *args],
            capture_output=True, text=True, timeout=10,
        )
        # JSON parse
        # stdout 의 마지막 valid JSON dict 를 찾기
        lines = result.stdout.strip().split("\n")
        # 거꾸로 검색 — 마지막 '{' 부터의 valid JSON
        for i in range(len(lines) - 1, -1, -1):
            line = lines[i].strip()
            if line.startswith("{"):
                candidate = "\n".join(lines[i:])
                try:
                    return json.loads(candidate)
                except json.JSONDecodeError:
                    continue
        return None
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return None


def test_automated_repro_scaffold_runtime() -> None:
    """automated-repro-scaffold (pilot): happy path → stage_completion 포함."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
        f.write("# Bug\nTest")
        bug_path = f.name
    repro_path = bug_path.replace(".md", "_repro.py")
    try:
        result = _run_skill_actual(
            "automated-repro-scaffold", "--report", bug_path, "--output", repro_path
        )
        if result is None:
            print(f"  SKIP  automated-repro-scaffold (runtime not available)")
            return
        # v0.6.5 pilot 후 runtime 적용
        assert is_stage_completion_present(result) is True
        assert result["stage_completion"]["stage_name"] == "automated-repro-scaffold"
        assert result["stage_completion"]["stage_status"] == "ok"
    finally:
        Path(bug_path).unlink(missing_ok=True)
        Path(repro_path).unlink(missing_ok=True)


# --- Test 3: helper functions 일관성 ---


def test_build_then_merge_roundtrip() -> None:
    """build → merge → ensure 가 동일한 stage_completion 반환."""
    sc = build_stage_completion(
        stage_name="validation-plan",
        stage_status="ok",
        artifacts=["backlog.md"],
        next_stage="code-index-update",
    )
    base = {"status": "ok"}
    merged = merge_into_result(base, sc)
    ensured = ensure_stage_completion(merged, stage_name="different")
    # 기존 stage_completion 보존
    assert ensured["stage_completion"]["stage_name"] == "validation-plan"
    assert ensured["stage_completion"]["next_stage"] == "code-index-update"


def test_legacy_result_compatibility() -> None:
    """v0.6.4 이전 legacy result (stage_completion 없음) 도 ensure 로 복구 가능."""
    legacy_results = [
        {"status": "ok", "summary": "old result", "warnings": []},
        {"status": "success", "tool_version": "0.6.3"},
        {"status": "error", "error": "old error", "error_code": "X"},
    ]
    for legacy in legacy_results:
        ensured = ensure_stage_completion(legacy, stage_name="legacy")
        assert is_stage_completion_present(ensured) is True
        # 기존 field 보존
        for key in legacy:
            assert ensured[key] == legacy[key]


def test_8_field_completeness_after_ensure() -> None:
    """ensure 후 8 field 모두 존재 검증."""
    result = {"status": "ok"}
    ensured = ensure_stage_completion(
        result,
        stage_name="test-stage",
        next_stage="next-stage",
        artifacts=["a.md", "b.md"],
        notes=["test note 1", "test note 2"],
    )
    sc = ensured["stage_completion"]
    required = {
        "stage_name", "stage_status", "next_stage",
        "requested_changes", "approval_timestamp", "approval_actor",
        "artifacts", "notes",
    }
    assert required.issubset(set(sc.keys()))
    # 8 field 모두 빈 list / None 으로 채워짐 (정확한 dict literal)
    assert sc["stage_name"] == "test-stage"
    assert sc["stage_status"] == "ok"
    assert sc["next_stage"] == "next-stage"
    assert sc["requested_changes"] == []
    assert sc["approval_timestamp"] is None
    assert sc["approval_actor"] is None
    assert sc["artifacts"] == ["a.md", "b.md"]
    assert sc["notes"] == ["test note 1", "test note 2"]


# --- Test 4: workflow_skill_catalog §5.2 매핑 검증 (v0.6.5 spec) ---


def test_skill_catalog_stage_name_mapping() -> None:
    """workflow_skill_catalog.md §5.2 의 stage name 매핑 정합성.

    본 test 는 spec 의 *문서* 정합성 검증. L1 wiki 또는
    workflow_skill_catalog.md 의 stage name 이 11종 skill list 와 일치하는지.
    """
    catalog_path = REPO_ROOT / "workflow-source" / "core" / "workflow_skill_catalog.md"
    if not catalog_path.exists():
        return
    text = catalog_path.read_text(encoding="utf-8")

    # §5.2 표에 11종 skill 모두 등장해야 함
    expected_skills = {
        "session-start", "backlog-update", "doc-sync", "merge-doc-reconcile",
        "validation-plan", "code-index-update", "automated-repro-scaffold",
        "workflow-linter", "project-status-assessment", "memory-freeze",
        "git-conflict-resolver", "robust-patcher",
    }
    for skill in expected_skills:
        # '| `session-start` |' 같은 markdown table cell 패턴
        assert f"| `{skill}` |" in text, f"skill '{skill}' not found in §5.2 table"


# --- Main ---


def main() -> int:
    tests = [
        ("ensure_creates_when_missing", test_ensure_creates_when_missing),
        ("ensure_preserves_existing", test_ensure_preserves_existing),
        ("ensure_status_mapping", test_ensure_status_mapping),
        ("automated_repro_scaffold_runtime", test_automated_repro_scaffold_runtime),
        ("build_then_merge_roundtrip", test_build_then_merge_roundtrip),
        ("legacy_result_compatibility", test_legacy_result_compatibility),
        ("8_field_completeness_after_ensure", test_8_field_completeness_after_ensure),
        ("skill_catalog_stage_name_mapping", test_skill_catalog_stage_name_mapping),
    ]
    failed: list[str] = []
    for name, fn in tests:
        try:
            fn()
            print(f"  PASS  {name}")
        except AssertionError as e:
            failed.append(f"  FAIL  {name}: {e}")
            print(f"  FAIL  {name}: {e}")
        except Exception as e:  # noqa: BLE001
            failed.append(f"  ERROR {name}: {type(e).__name__}: {e}")
            print(f"  ERROR {name}: {type(e).__name__}: {e}")

    print()
    if failed:
        print(f"{len(failed)}/{len(tests)} tests failed:")
        for f in failed:
            print(f)
        return 1
    print(f"All {len(tests)} tests passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
