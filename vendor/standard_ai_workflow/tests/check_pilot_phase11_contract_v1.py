#!/usr/bin/env python3
"""Phase 11 pilot regression check: Devhub Example × Contract v1.

Verifies that the pilot result artifact (workflow-source/examples/pilot_phase11_devhub_contract_v1.md)
exists, contains the required sections, references the 4 Devhub_example PR IDs,
and that the §4/§5 JSON examples in it are parseable.
Reference: workflow-source/examples/pilot_phase11_devhub_contract_v1.md
           workflow-source/core/orchestrator_subagent_contract_v1.md
"""

from __future__ import annotations

import json
import re
from pathlib import Path

PILOT_DOC_PATH = Path(__file__).resolve().parents[1] / "examples" / "pilot_phase11_devhub_contract_v1.md"

# Required sections (markdown headings)
REQUIRED_SECTIONS = (
    "## 1. 동기",
    "## 2. Pilot 데이터",
    "## 3. Contract v1 Round-Trip Walkthrough",
    "## 4. §6 카탈로그 정합성",
    "## 5. §4/§5 JSON 스키마",
    "## 6. v0.5.6 Enforcement",
    "## 7. S4 라이브 데모",
    "## 8. 결론",
)

# Required PR references (Devhub_example PR IDs covered by the pilot)
REQUIRED_PR_REFS = (
    "#493",  # chore: untrack antigravity
    "#492",  # feature: N:M contribution weights
    "#491",  # UI: dashboard + KPI
    "#490",  # docs: traceability
)

# Per-scenario delegations that must be present in the pilot doc.
EXPECTED_DELEGATIONS = (
    "del-2026-06-07-p493",
    "del-2026-06-07-p492",
    "del-2026-06-07-p491-ui",
    "del-2026-06-07-p490",
)


def _extract_json_blocks(text: str) -> list[dict[str, object]]:
    """Pull every ```json ... ``` block, parse it, and return the list of dicts."""
    blocks: list[dict[str, object]] = []
    pattern = re.compile(r"```json\s*\n(.*?)\n```", re.DOTALL)
    for match in pattern.finditer(text):
        try:
            payload = json.loads(match.group(1))
        except json.JSONDecodeError as exc:
            raise AssertionError(f"Bad JSON block in pilot doc: {exc}") from exc
        if not isinstance(payload, dict):
            raise AssertionError(
                f"Expected JSON block to be a dict, got {type(payload).__name__}"
            )
        blocks.append(payload)
    return blocks


def check_pilot_doc_present() -> None:
    if not PILOT_DOC_PATH.exists():
        raise AssertionError(f"Pilot result doc not found: {PILOT_DOC_PATH}")
    text = PILOT_DOC_PATH.read_text(encoding="utf-8")
    for section in REQUIRED_SECTIONS:
        if section not in text:
            raise AssertionError(f"Pilot doc missing required section: {section!r}")


def check_pilot_pr_references() -> None:
    text = PILOT_DOC_PATH.read_text(encoding="utf-8")
    for pr_ref in REQUIRED_PR_REFS:
        if pr_ref not in text:
            raise AssertionError(f"Pilot doc missing PR reference: {pr_ref!r}")


def check_pilot_json_roundtrip() -> None:
    text = PILOT_DOC_PATH.read_text(encoding="utf-8")
    blocks = _extract_json_blocks(text)
    # Expect 8 JSON blocks: 4 scenarios × 2 (input §4 + output §5)
    if len(blocks) < 8:
        raise AssertionError(
            f"Pilot doc should contain >= 8 JSON blocks (4 scenarios × 2 each), got {len(blocks)}"
        )

    # Each input block must have contract_version=1.0 + delegation_id
    seen_delegation_ids: set[str] = set()
    for block in blocks:
        if "contract_version" not in block:
            raise AssertionError("JSON block missing contract_version")
        if "delegation_id" not in block:
            raise AssertionError("JSON block missing delegation_id")
        seen_delegation_ids.add(str(block["delegation_id"]))

    for expected in EXPECTED_DELEGATIONS:
        if expected not in seen_delegation_ids:
            raise AssertionError(
                f"Pilot doc missing expected delegation_id: {expected!r}"
            )

    # Pair validation: each input/output pair must share delegation_id
    delegation_to_kind: dict[str, set[str]] = {}
    for block in blocks:
        did = str(block["delegation_id"])
        # input blocks have "issued_by.role == orchestrator" + "task" field
        is_input = (
            isinstance(block.get("issued_by"), dict)
            and block["issued_by"].get("role") == "orchestrator"  # type: ignore[index]
            and "task" in block
        )
        # output blocks have "worker" + "result" fields
        is_output = "worker" in block and "result" in block
        if is_input and is_output:
            raise AssertionError(
                f"Block {did!r} looks like BOTH input and output (cannot classify)"
            )
        if not is_input and not is_output:
            raise AssertionError(
                f"Block {did!r} looks like NEITHER input nor output"
            )
        kind = "input" if is_input else "output"
        delegation_to_kind.setdefault(did, set()).add(kind)

    for did, kinds in delegation_to_kind.items():
        if kinds != {"input", "output"}:
            raise AssertionError(
                f"Delegation {did!r} has unbalanced kinds: {kinds}"
            )


def check_pilot_v056_priorities() -> None:
    text = PILOT_DOC_PATH.read_text(encoding="utf-8")
    for marker in ("P0", "P1", "P2"):
        if marker not in text:
            raise AssertionError(f"Pilot doc missing priority marker: {marker!r}")
    for priority_text in (
        "sub-agent 측 출력 스키마 가드",
        "orchestrator 측 §6.1 자동 위임",
        "contract v2 fan-out/in",
    ):
        if priority_text not in text:
            raise AssertionError(f"Pilot doc missing v0.5.6 priority: {priority_text!r}")


def main() -> int:
    check_pilot_doc_present()
    check_pilot_pr_references()
    check_pilot_json_roundtrip()
    check_pilot_v056_priorities()
    print(
        "Phase 11 pilot (Devhub Example × Contract v1) regression check passed "
        "(8 required sections, 4 PR refs, 8 JSON blocks round-trip, v0.5.6 priorities)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
