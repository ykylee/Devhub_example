#!/usr/bin/env python3
"""Generate checked-in read-only transport descriptor drafts."""

from __future__ import annotations

import json
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.server.read_only_registry import build_transport_tool_descriptors


def main() -> int:
    print(json.dumps(build_transport_tool_descriptors(), ensure_ascii=False, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
