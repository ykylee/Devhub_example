#!/usr/bin/env python3
"""Update schemas/output_sample_contracts.json to match runtime contracts."""

import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.output_contracts import (
    COMMON_REQUIRED_KEYS,
    ERROR_PATH_CONTRACTS,
    output_error_field_shapes_schema,
    output_field_shapes_schema,
    SUCCESS_PATH_CONTRACTS,
)

def main():
    contract = {
        "tool_version_source": "workflow_kit.__version__",
        "common_required_keys": sorted(list(COMMON_REQUIRED_KEYS)),
        "success_required_keys": {k: sorted(list(v)) for k, v in SUCCESS_PATH_CONTRACTS.items()},
        "error_required_keys": {k: sorted(list(v)) for k, v in ERROR_PATH_CONTRACTS.items()},
        "field_shapes": output_field_shapes_schema(),
        "error_field_shapes": output_error_field_shapes_schema(),
    }
    print(json.dumps(contract, ensure_ascii=False, indent=2, sort_keys=True))

if __name__ == "__main__":
    main()
