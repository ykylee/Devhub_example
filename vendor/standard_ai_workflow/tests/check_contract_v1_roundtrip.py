#!/usr/bin/env python3
"""S1: Orchestrator ↔ Sub-agent contract v1 round-trip check.

Verifies that the §4 input schema and §5 output schema can be serialized
and that delegation_id / contract_version / role mapping is preserved.
Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §8.1.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

CONTRACT_PATH = Path(__file__).resolve().parents[1] / "core" / "orchestrator_subagent_contract_v1.md"
TASK_TYPE_TO_ROLE = {
    "doc_draft": "doc-worker",
    "code_change": "code-worker",
    "validation_run": "validation-worker",
    "bounded_research": "doc-worker",
}

# --- Schema fragments (must match §4 / §5 of the contract) ---

INPUT_REQUIRED_FIELDS = (
    "contract_version",
    "delegation_id",
    "issued_at",
    "issued_by",
    "task",
    "context",
)

OUTPUT_REQUIRED_FIELDS = (
    "contract_version",
    "delegation_id",
    "completed_at",
    "worker",
    "result",
)


def _build_sample_input() -> dict[str, object]:
    return {
        "contract_version": "1.0",
        "delegation_id": "del-2026-06-07-001",
        "issued_at": "2026-06-07T18:30:00+09:00",
        "issued_by": {
            "session_id": "mvs_a96f8eb4990a482ca14e3e5223447bb7",
            "role": "orchestrator",
        },
        "task": {
            "task_id": "TASK-V054-001",
            "task_type": "doc_draft",
            "brief": "Draft v1 contract spec for orchestrator → sub-agent delegation.",
            "constraints": [
                "scope: workflow-source/core/orchestrator_subagent_contract_v1.md",
                "do not touch: workflow_kit, scripts, tests",
            ],
            "inputs": {
                "files": [
                    "workflow-source/core/workflow_agent_topology.md",
                    "workflow-source/core/global_workflow_standard.md",
                ],
                "context_paths": [
                    "ai-workflow/memory/release/v0.5.4/",
                ],
            },
            "expected_outputs": {
                "primary_artifact": "workflow-source/core/orchestrator_subagent_contract_v1.md",
                "artifact_kind": "markdown",
                "must_include": [
                    "위임 입력/출력 스키마",
                    "4개 역할 경계",
                    "위임 가능/불가 카탈로그",
                ],
            },
            "validation": {
                "required": True,
                "criteria": "linter status ok + check_docs.py PASS",
                "owner": "orchestrator",
            },
        },
        "context": {
            "branch": "release/v0.5.4",
            "memory_layer_root": "ai-workflow/memory/release/v0.5.4/",
            "project_root": "/Users/yklee/repos/standard_ai_workflow_minimax",
        },
    }


def _build_sample_output(input_payload: dict[str, object]) -> dict[str, object]:
    task_type = input_payload["task"]["task_type"]  # type: ignore[index]
    return {
        "contract_version": input_payload["contract_version"],
        "delegation_id": input_payload["delegation_id"],
        "completed_at": "2026-06-07T19:45:00+09:00",
        "worker": {
            "session_id": "mvs_child_xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
            "role": TASK_TYPE_TO_ROLE[task_type],  # type: ignore[index]
            "model_tier": "small",
        },
        "result": {
            "status": "ok",
            "summary": "Drafted contract v1 spec with all required sections.",
            "artifacts": [
                {
                    "path": input_payload["task"]["expected_outputs"]["primary_artifact"],  # type: ignore[index]
                    "kind": "markdown",
                    "lines": 280,
                    "action": "created",
                }
            ],
            "written_paths": [
                input_payload["task"]["expected_outputs"]["primary_artifact"]  # type: ignore[index]
            ],
            "validation_result": {
                "ran": True,
                "command": "python workflow-source/tests/check_docs.py",
                "status": "pass",
                "details": "102 markdown files PASS",
            },
            "next_step": "orchestrator 가 PR 본문 작성 + cross-link 갱신",
        },
    }


def _check_input_schema(payload: dict[str, object]) -> None:
    for field in INPUT_REQUIRED_FIELDS:
        if field not in payload:
            raise AssertionError(f"Input schema missing required field: {field!r}")
    if payload["contract_version"] != "1.0":
        raise AssertionError(f"Input contract_version must be '1.0', got {payload['contract_version']!r}")
    if not str(payload["delegation_id"]).startswith("del-"):
        raise AssertionError(f"Input delegation_id must start with 'del-', got {payload['delegation_id']!r}")
    if payload["issued_by"]["role"] != "orchestrator":  # type: ignore[index]
        raise AssertionError(f"Input issued_by.role must be 'orchestrator', got {payload['issued_by']['role']!r}")  # type: ignore[index]


def _check_output_schema(payload: dict[str, object], input_payload: dict[str, object]) -> None:
    for field in OUTPUT_REQUIRED_FIELDS:
        if field not in payload:
            raise AssertionError(f"Output schema missing required field: {field!r}")
    if payload["contract_version"] != input_payload["contract_version"]:
        raise AssertionError(
            f"Output contract_version mismatch: input={input_payload['contract_version']!r}, "
            f"output={payload['contract_version']!r}"
        )
    if payload["delegation_id"] != input_payload["delegation_id"]:
        raise AssertionError(
            f"Output delegation_id mismatch: input={input_payload['delegation_id']!r}, "
            f"output={payload['delegation_id']!r}"
        )
    expected_role = TASK_TYPE_TO_ROLE[input_payload["task"]["task_type"]]  # type: ignore[index]
    if payload["worker"]["role"] != expected_role:  # type: ignore[index]
        raise AssertionError(
            f"Output worker.role mismatch for task_type={input_payload['task']['task_type']!r}: "  # type: ignore[index]
            f"expected={expected_role!r}, got={payload['worker']['role']!r}"  # type: ignore[index]
        )


def check_input_output_roundtrip() -> None:
    input_payload = _build_sample_input()
    output_payload = _build_sample_output(input_payload)
    _check_input_schema(input_payload)
    _check_output_schema(output_payload, input_payload)
    # Round-trip serialization (matches real orchestrator <-> sub-agent exchange)
    input_json = json.dumps(input_payload, ensure_ascii=False, sort_keys=True)
    output_json = json.dumps(output_payload, ensure_ascii=False, sort_keys=True)
    json.loads(input_json)
    json.loads(output_json)


def check_contract_doc_present() -> None:
    if not CONTRACT_PATH.exists():
        raise AssertionError(f"Contract v1 spec doc not found: {CONTRACT_PATH}")
    text = CONTRACT_PATH.read_text(encoding="utf-8")
    if "## 4. 위임 입력 스키마" not in text:
        raise AssertionError("Contract doc missing §4 (위임 입력 스키마)")
    if "## 5. 위임 출력 스키마" not in text:
        raise AssertionError("Contract doc missing §5 (위임 출력 스키마)")
    if "## 8. 검증 시나리오" not in text:
        raise AssertionError("Contract doc missing §8 (검증 시나리오)")


def main() -> int:
    check_contract_doc_present()
    check_input_output_roundtrip()
    print("Contract v1 round-trip smoke check passed (input ↔ output schemas, role mapping, doc presence).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
