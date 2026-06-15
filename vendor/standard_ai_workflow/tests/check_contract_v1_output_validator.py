#!/usr/bin/env python3
"""Contract v1 §5 output validator regression check.

Verifies that workflow_kit.contract_v1.output_validator correctly:
- accepts a well-formed §5 payload
- rejects 5+ violation patterns (missing field, contract_version, status enum, role enum, validation_result format)
- handles the v0.5.5 pilot's 4 scenario outputs as a regression baseline (no false positives on real-world §5 outputs)
Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §5
           workflow-source/examples/pilot_phase11_devhub_contract_v1.md
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.contract_v1 import validate_output  # noqa: E402

# --- 4 pilot baseline payloads (extracted from v0.5.5 walkthrough, minimal
# shape — just enough to exercise the validator. Real walkthrough includes
# additional optional fields, but the schema-level checks only need the
# required ones.)


def _baseline_output(delegation_id: str, role: str = "doc-worker", status: str = "ok") -> dict[str, object]:
    return {
        "contract_version": "1.0",
        "delegation_id": delegation_id,
        "completed_at": "2026-06-07T21:30:00+09:00",
        "worker": {"session_id": "mvs_x", "role": role, "model_tier": "small"},
        "result": {
            "status": status,
            "summary": "pilot baseline payload",
            "artifacts": [{"path": "x.md", "kind": "markdown", "lines": 10, "action": "created"}],
            "written_paths": ["x.md"],
            "next_step": "orchestrator review",
            "validation_result": {"ran": True, "command": "test", "status": "pass", "details": "ok"},
        },
        "warnings": [],
        "risks": [],
    }


def _extract_json_blocks(text: str) -> list[dict[str, object]]:
    pattern = re.compile(r"```json\s*\n(.*?)\n```", re.DOTALL)
    return [json.loads(m.group(1)) for m in pattern.finditer(text)]


def check_valid_pilot_baseline() -> None:
    """The 4 v0.5.5 pilot §5 outputs must all validate cleanly."""
    for did, role in [
        ("del-2026-06-07-p493", "doc-worker"),
        ("del-2026-06-07-p492", "code-worker"),
        ("del-2026-06-07-p491-ui", "code-worker"),
        ("del-2026-06-07-p490", "doc-worker"),
    ]:
        payload = _baseline_output(did, role=role)
        result = validate_output(payload, expected_delegation_id=did)
        if not result.is_valid:
            raise AssertionError(
                f"Pilot baseline {did!r} should validate, got errors: {result.errors}"
            )


def check_violation_missing_field() -> None:
    payload = _baseline_output("del-test")
    payload.pop("completed_at")
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Missing 'completed_at' should fail validation")
    if not any(err.field == "completed_at" for err in result.errors):
        raise AssertionError(f"Error should mention 'completed_at', got: {result.errors}")


def check_violation_wrong_contract_version() -> None:
    payload = _baseline_output("del-test")
    payload["contract_version"] = "2.0"
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Wrong contract_version should fail")
    if not any(err.field == "contract_version" for err in result.errors):
        raise AssertionError(f"Error should mention 'contract_version', got: {result.errors}")


def check_violation_status_enum() -> None:
    payload = _baseline_output("del-test", status="completed")
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Invalid status enum should fail")
    if not any(err.field == "result.status" for err in result.errors):
        raise AssertionError(f"Error should mention 'result.status', got: {result.errors}")


def check_violation_role_enum() -> None:
    payload = _baseline_output("del-test", role="super-worker")
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Invalid role enum should fail")
    if not any(err.field == "worker.role" for err in result.errors):
        raise AssertionError(f"Error should mention 'worker.role', got: {result.errors}")


def check_violation_validation_result_format() -> None:
    payload = _baseline_output("del-test")
    payload["result"]["validation_result"] = {"ran": True, "status": "maybe"}  # type: ignore[index]
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Invalid validation_result.status should fail")
    if not any(err.field == "result.validation_result.status" for err in result.errors):
        raise AssertionError(
            f"Error should mention 'result.validation_result.status', got: {result.errors}"
        )


def check_violation_delegation_id_mismatch() -> None:
    """The orchestrator passes expected_delegation_id; mismatches must fail."""
    payload = _baseline_output("del-expected")
    result = validate_output(payload, expected_delegation_id="del-different")
    if result.is_valid:
        raise AssertionError("delegation_id mismatch should fail")
    if not any(err.field == "delegation_id" for err in result.errors):
        raise AssertionError(f"Error should mention 'delegation_id', got: {result.errors}")


def check_violation_empty_next_step() -> None:
    payload = _baseline_output("del-test")
    payload["result"]["next_step"] = "   "  # whitespace-only
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Empty next_step should fail")
    if not any(err.field == "result.next_step" for err in result.errors):
        raise AssertionError(f"Error should mention 'result.next_step', got: {result.errors}")


def check_violation_artifact_kind_enum() -> None:
    payload = _baseline_output("del-test")
    payload["result"]["artifacts"][0]["kind"] = "exe"  # type: ignore[index]
    result = validate_output(payload)
    if result.is_valid:
        raise AssertionError("Invalid artifact kind should fail")
    if not any(err.field == "result.artifacts[0].kind" for err in result.errors):
        raise AssertionError(
            f"Error should mention 'result.artifacts[0].kind', got: {result.errors}"
        )


def check_pilot_doc_outputs_validate() -> None:
    """The actual §5 outputs embedded in the v0.5.5 pilot doc must validate.
    This is the strongest regression: the pilot's documented payloads are
    expected real-world baselines."""
    pilot_doc = SOURCE_ROOT / "examples" / "pilot_phase11_devhub_contract_v1.md"
    if not pilot_doc.exists():
        raise AssertionError(f"Pilot doc not found: {pilot_doc}")
    text = pilot_doc.read_text(encoding="utf-8")
    blocks = _extract_json_blocks(text)
    if not blocks:
        raise AssertionError("No JSON blocks found in pilot doc")
    # Every block should at least parse as a dict (we already do this elsewhere).
    # For §5 outputs (have 'worker' + 'result' but no 'task'), validate.
    seen_outputs = 0
    for block in blocks:
        if "worker" in block and "result" in block and "task" not in block:
            result = validate_output(block)
            if not result.is_valid:
                raise AssertionError(
                    f"Pilot doc §5 output failed validation: {block.get('delegation_id')!r} errors={result.errors}"
                )
            seen_outputs += 1
    if seen_outputs < 4:
        raise AssertionError(
            f"Expected >= 4 §5 outputs in pilot doc to validate, got {seen_outputs}"
        )


def main() -> int:
    check_valid_pilot_baseline()
    check_violation_missing_field()
    check_violation_wrong_contract_version()
    check_violation_status_enum()
    check_violation_role_enum()
    check_violation_validation_result_format()
    check_violation_delegation_id_mismatch()
    check_violation_empty_next_step()
    check_violation_artifact_kind_enum()
    check_pilot_doc_outputs_validate()
    print(
        "Contract v1 §5 output validator smoke check passed "
        "(4 pilot baselines valid, 8 violations caught, pilot doc §5 outputs validate)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
