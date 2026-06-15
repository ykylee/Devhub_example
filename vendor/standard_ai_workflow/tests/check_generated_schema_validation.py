#!/usr/bin/env python3
"""Validate output samples against the generated JSON Schema drafts."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

from jsonschema import validate, ValidationError


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
SAMPLES_DIR = SOURCE_ROOT / "examples" / "output_samples"
README_PATH = SAMPLES_DIR / "README.md"
SCHEMA_PATH = SOURCE_ROOT / "schemas" / "generated_output_schemas.json"
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+\.json)\)")

if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import detect_sample_family


def listed_sample_paths() -> list[Path]:
    text = README_PATH.read_text(encoding="utf-8")
    return [((README_PATH.parent / match.group(1).strip()).resolve()) for match in LINK_RE.finditer(text)]


def main() -> int:
    if not SCHEMA_PATH.exists():
        print(f"Schema path {SCHEMA_PATH} does not exist.")
        return 1

    generated = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    outputs_bundle = generated.get("outputs", {})
    errors_bundle = generated.get("errors", {})
    failures: list[str] = []

    for path in listed_sample_paths():
        if not path.exists():
            failures.append(f"Sample file not found: {path}")
            continue

        payload = json.loads(path.read_text(encoding="utf-8"))
        family = detect_sample_family(path)
        if not family:
            failures.append(f"Could not infer family for sample {path.name}")
            continue
        
        status = payload.get("status")
        bundle = errors_bundle if status == "error" else outputs_bundle
        
        # In the new Pydantic bundle, errors might be keyed by family or just 'error'
        # Our bundle generator currently provides it for each family for compatibility.
        schema = bundle.get(family)
        
        if schema is None:
            failures.append(f"Missing generated JSON Schema for family {family} (status: {status})")
            continue
        
        try:
            validate(instance=payload, schema=schema)
        except ValidationError as e:
            failures.append(f"Sample {path.name} failed schema validation: {e.message}")

    if failures:
        print("Generated schema validation failed:")
        for failure in failures:
            print(f"- {failure}")
        return 1

    print("All output samples passed Pydantic-generated JSON Schema validation.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
