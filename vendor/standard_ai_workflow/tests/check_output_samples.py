#!/usr/bin/env python3
"""Validate JSON output samples referenced by the workflow kit docs."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SAMPLES_DIR = SOURCE_ROOT / "examples" / "output_samples"
README_PATH = SAMPLES_DIR / "README.md"
SCHEMA_PATH = SOURCE_ROOT / "schemas" / "output_sample_contracts.json"
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+\.json)\)")

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit import __version__ as TOOL_VERSION
from workflow_kit.common.output_contracts import (
    COMMON_REQUIRED_KEYS,
    ERROR_PATH_CONTRACTS,
    output_error_field_shapes_schema,
    output_field_shapes_schema,
    SUCCESS_PATH_CONTRACTS,
    detect_sample_family,
    validate_output_payload,
)


def listed_sample_paths() -> list[Path]:
    text = README_PATH.read_text(encoding="utf-8")
    results: list[Path] = []
    for match in LINK_RE.finditer(text):
        target = match.group(1).strip()
        results.append((README_PATH.parent / target).resolve())
    return results


def actual_sample_paths() -> list[Path]:
    return sorted(SAMPLES_DIR.glob("*.json"))


def load_schema_contract() -> dict[str, object]:
    return json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))


def compare_schema_and_runtime_contracts(schema: dict[str, object], failures: list[str]) -> None:
    if schema.get("tool_version_source") != "workflow_kit.__version__":
        failures.append("ai-workflow/schemas/output_sample_contracts.json has an unexpected tool_version_source.")

    if set(schema.get("common_required_keys", [])) != set(COMMON_REQUIRED_KEYS):
        failures.append("Schema common_required_keys do not match workflow_kit.common.output_contracts.")

    schema_success = {
        key: set(value)
        for key, value in dict(schema.get("success_required_keys", {})).items()
    }
    runtime_success = {key: set(value) for key, value in SUCCESS_PATH_CONTRACTS.items()}
    if schema_success != runtime_success:
        failures.append("Schema success_required_keys do not match workflow_kit.common.output_contracts.")

    schema_error = {
        key: set(value)
        for key, value in dict(schema.get("error_required_keys", {})).items()
    }
    runtime_error = {key: set(value) for key, value in ERROR_PATH_CONTRACTS.items()}
    if schema_error != runtime_error:
        failures.append("Schema error_required_keys do not match workflow_kit.common.output_contracts.")

    schema_shapes = dict(schema.get("field_shapes", {}))
    runtime_shapes = output_field_shapes_schema()
    if schema_shapes != runtime_shapes:
        failures.append("Schema field_shapes do not match workflow_kit.common.output_contracts.")

    schema_error_shapes = dict(schema.get("error_field_shapes", {}))
    runtime_error_shapes = output_error_field_shapes_schema()
    if schema_error_shapes != runtime_error_shapes:
        failures.append("Schema error_field_shapes do not match workflow_kit.common.output_contracts.")


def main() -> int:
    failures: list[str] = []
    schema = load_schema_contract()
    compare_schema_and_runtime_contracts(schema, failures)
    listed = listed_sample_paths()
    actual = actual_sample_paths()
    if not listed:
        failures.append("No JSON sample links were found in examples/output_samples/README.md.")

    listed_set = set(listed)
    actual_set = set(actual)
    missing_from_readme = sorted(actual_set - listed_set)
    extra_in_readme = sorted(listed_set - actual_set)
    for path in missing_from_readme:
        failures.append(f"Sample JSON is not listed in README: {path.relative_to(REPO_ROOT)}")
    for path in extra_in_readme:
        failures.append(f"README lists a missing JSON sample: {path.relative_to(REPO_ROOT)}")

    for path in actual:
        if not path.exists():
            failures.append(f"Missing listed sample: {path.relative_to(REPO_ROOT)}")
            continue
        try:
            payload = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            failures.append(f"Invalid JSON in {path.relative_to(REPO_ROOT)}: {exc}")
            continue
        if not isinstance(payload, dict):
            failures.append(f"Sample is not a JSON object: {path.relative_to(REPO_ROOT)}")
            continue
        for key in sorted(COMMON_REQUIRED_KEYS):
            if key not in payload:
                failures.append(f"Missing `{key}` in {path.relative_to(REPO_ROOT)}")
        if payload.get("tool_version") != TOOL_VERSION:
            failures.append(
                f"Unexpected `tool_version` in {path.relative_to(REPO_ROOT)}: "
                f"expected {TOOL_VERSION}, got {payload.get('tool_version')}"
            )

        family = detect_sample_family(path)
        if payload.get("status") == "error":
            if "error_code" not in payload:
                failures.append(f"Missing `error_code` for error sample {path.relative_to(REPO_ROOT)}")
            if family in ERROR_PATH_CONTRACTS:
                for key in sorted(ERROR_PATH_CONTRACTS[family]):
                    if key not in payload:
                        failures.append(
                            f"Missing `{key}` in error sample {path.relative_to(REPO_ROOT)}"
                        )
        elif family in SUCCESS_PATH_CONTRACTS:
            for key in sorted(SUCCESS_PATH_CONTRACTS[family]):
                if key not in payload:
                    failures.append(
                        f"Missing `{key}` in success sample {path.relative_to(REPO_ROOT)}"
                    )
        if family:
            for error in validate_output_payload(payload, family=family):
                failures.append(f"{path.relative_to(REPO_ROOT)}: {error}")

    if failures:
        print("Output sample check failed:")
        for failure in failures:
            print(f"- {failure}")
        return 1

    print(f"Output sample check passed for {len(actual)} JSON files.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
