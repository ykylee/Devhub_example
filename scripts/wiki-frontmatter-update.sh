#!/usr/bin/env bash
# wiki-frontmatter-update.sh — wiki sources/ page 의 frontmatter 에 mirror 의 provenance 자동 반영.
#
# 사용:
#   bash scripts/wiki-frontmatter-update.sh
#
# 본 script 의 source-of-truth:
#   - mirror manifest: ~/wiki/raw/projects/devhub/_manifest.md
#   - 위키 sources/ pages: ~/wiki/wiki/projects/devhub/sources/*.md
#
# 본 script 의 정공법:
#   - mirror manifest 의 commit hash + version 정보 + last sync timestamp 를
#     위키의 sources/ page 의 frontmatter 에 자동 반영.
#   - 위키 정보의 provenance 자동 추적 가능.
#   - 위키가 stale 한지 검증 가능 (last_touched < commit 시간 → stale).
#
# 결정적이고 단순: Python + 정규식 frontmatter parser.
#
# D-72 Phase 1.5 + provenance tracking (2026-06-13) 의 본 저장소 측 script.
#
# Exit code:
#   0 — success
#   1 — manifest 부재 / wiki dir 부재 / source-of-truth drift

set -euo pipefail

# ----- paths -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${HOME}/wiki"
RAW_DIR="$VAULT_ROOT/raw/projects/devhub"
WIKI_SOURCES="$VAULT_ROOT/wiki/projects/devhub/sources"
MANIFEST="$RAW_DIR/_manifest.md"

# ----- validation -----
if [[ ! -f "$MANIFEST" ]]; then
  echo "[wiki-frontmatter-update] manifest 부재: $MANIFEST" >&2
  echo "[wiki-frontmatter-update] hint: 'bash scripts/wiki-sync-devhub.sh' 먼저 실행" >&2
  exit 1
fi

if [[ ! -d "$WIKI_SOURCES" ]]; then
  echo "[wiki-frontmatter-update] wiki sources/ 부재: $WIKI_SOURCES" >&2
  exit 1
fi

# ----- Python script for frontmatter update -----
python3 << PYEOF
import os
import re
import glob

RAW_DIR = "$RAW_DIR"
WIKI_SOURCES = "$WIKI_SOURCES"
MANIFEST = "$MANIFEST"

# ----- extract provenance from manifest -----
with open(MANIFEST) as f:
    content = f.read()

m = re.search(r'commit \(full\) \| (\S+)', content)
COMMIT_FULL = m.group(1) if m else "unknown"
m = re.search(r'commit \(short\) \| (\S+)', content)
COMMIT_SHORT = m.group(1) if m else "unknown"
m = re.search(r'branch \| (\S+)', content)
BRANCH = m.group(1) if m else "unknown"
m = re.search(r'dirty \| (.+)', content)
DIRTY = m.group(1).strip() if m else ""
m = re.search(r'system version.*?\| (\S+)', content)
VERSION_SYSTEM = m.group(1) if m else "unknown"
m = re.search(r'workflow version.*?\| (\S+)', content)
VERSION_WORKFLOW = m.group(1) if m else "unknown"
m = re.search(r'timestamp: (\S+)', content)
SYNC_TS = m.group(1) if m else "unknown"

print(f"[wiki-frontmatter-update] manifest provenance:")
print(f"  commit: {COMMIT_SHORT} ({BRANCH}) {DIRTY}")
print(f"  version: system={VERSION_SYSTEM} workflow={VERSION_WORKFLOW}")
print(f"  sync: {SYNC_TS}")

# ----- update each wiki sources/ page's frontmatter -----
UPDATED = 0
SKIPPED = 0
for page_path in sorted(glob.glob(f"{WIKI_SOURCES}/*.md")):
    fname = os.path.basename(page_path)
    if fname == "_manifest.md":
        continue

    with open(page_path) as f:
        content = f.read()

    # Parse frontmatter
    if not content.startswith("---\n"):
        SKIPPED += 1
        continue

    # Find end of frontmatter
    lines = content.split("\n")
    end_idx = None
    for i, line in enumerate(lines[1:], 1):
        if line.strip() == "---":
            end_idx = i
            break

    if end_idx is None:
        SKIPPED += 1
        continue

    body_lines = lines[end_idx+1:]
    body = "\n".join(body_lines).lstrip("\n")

    # Update or add fields
    frontmatter_lines = lines[:end_idx+1]

    def update_field(frontmatter_lines, key, value):
        new_lines = []
        found = False
        for line in frontmatter_lines:
            if line.startswith(f"{key}:"):
                new_lines.append(f"{key}: {value}")
                found = True
            else:
                new_lines.append(line)
        if not found:
            # Add before closing ---
            new_lines.insert(-1, f"{key}: {value}")
        return new_lines

    frontmatter_lines = update_field(frontmatter_lines, "git_commit", COMMIT_SHORT)
    frontmatter_lines = update_field(frontmatter_lines, "git_branch", BRANCH)
    frontmatter_lines = update_field(frontmatter_lines, "version_system", VERSION_SYSTEM)
    frontmatter_lines = update_field(frontmatter_lines, "version_workflow", VERSION_WORKFLOW)
    frontmatter_lines = update_field(frontmatter_lines, "last_touched", SYNC_TS)
    frontmatter_lines = update_field(frontmatter_lines, "mirror_dirty", DIRTY)

    # Write back
    new_content = "\n".join(frontmatter_lines) + "\n\n" + body
    with open(page_path, "w") as f:
        f.write(new_content)
    UPDATED += 1

print(f"[wiki-frontmatter-update] DONE")
print(f"  updated: {UPDATED} page(s)")
print(f"  skipped: {SKIPPED} page(s) (no frontmatter or _manifest.md)")
PYEOF
