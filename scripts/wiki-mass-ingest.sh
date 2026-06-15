#!/usr/bin/env bash
# wiki-mass-ingest.sh — Phase 3 mass ingest 정공법.
#
# 사용:
#   bash scripts/wiki-mass-ingest.sh                    # dry-run (default)
#   bash scripts/wiki-mass-ingest.sh --apply           # real ingest
#   bash scripts/wiki-mass-ingest.sh --dir=<dir>       # 특정 directory 만 ingest
#
# 본 script 의 source-of-truth:
#   - mirror manifest: ai-workflow/wiki/raw/projects/devhub/_manifest.md
#   - 위키 sources/ pages: ai-workflow/wiki/wiki/projects/devhub/sources/*.md
#   - Phase 3 scope: docs/domain + docs/architecture + docs/infrastructure + docs/validation
#
# 본 script 의 정공법:
#   - raw/ 의 Phase 3 scope (~78 file) 의 wiki page 자동 생성
#   - frontmatter + raw body 1:1 mirror (Phase 1.5 의 1:1 mirror 정공법 정합)
#   - provenance 자동 반영 (git_commit + version + last_touched)
#   - wiki index.md 의 자동 갱신 (L08 fix)
#
# 결정적이고 단순: Python + 정규식 frontmatter parser.
#
# D-72 Phase 3 (2026-06-13) 의 본 저장소 측 script.
#
# Exit code:
#   0 — success
#   1 — manifest 부재 / wiki dir 부재 / source-of-truth drift

set -euo pipefail

# ----- options -----
APPLY=0
TARGET_DIR=""
for arg in "$@"; do
  case "$arg" in
    --apply)
      APPLY=1
      ;;
    --dry-run)
      APPLY=0
      ;;
    --dir=*)
      TARGET_DIR="${arg#--dir=}"
      ;;
    -h|--help)
      echo "usage: bash scripts/wiki-mass-ingest.sh [--apply] [--dir=<domain|architecture|infrastructure|validation>]"
      echo "  (default)  dry-run (no actual changes)"
      echo "  --apply    real ingest (wiki page 생성)"
      echo "  --dir=<d>  특정 directory 만 ingest"
      exit 0
      ;;
    *)
      echo "[wiki-mass-ingest] invalid option: $arg" >&2
      exit 1
      ;;
  esac
done

# ----- paths -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${SRC}/ai-workflow/wiki"
RAW_DIR="$VAULT_ROOT/raw/projects/devhub"
WIKI_SOURCES="$VAULT_ROOT/wiki/projects/devhub/sources"
MANIFEST="$RAW_DIR/_manifest.md"

# ----- validation -----
if [[ ! -f "$MANIFEST" ]]; then
  echo "[wiki-mass-ingest] manifest 부재: $MANIFEST" >&2
  exit 1
fi

if [[ ! -d "$WIKI_SOURCES" ]]; then
  echo "[wiki-mass-ingest] wiki sources/ 부재: $WIKI_SOURCES" >&2
  exit 1
fi

# ----- Phase 3 scope (default: all 4 directory) -----
if [[ -n "$TARGET_DIR" ]]; then
  SCOPE_DIRS="$TARGET_DIR"
else
  SCOPE_DIRS="domain architecture infrastructure validation"
fi

# ----- Python script for mass ingest -----
python3 << PYEOF
import os
import re
import glob

RAW_DIR = "$RAW_DIR"
WIKI_SOURCES = "$WIKI_SOURCES"
MANIFEST = "$MANIFEST"
SCOPE_DIRS = "$SCOPE_DIRS".split()
APPLY = $APPLY

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

print(f"[wiki-mass-ingest] provenance:")
print(f"  commit: {COMMIT_SHORT} ({BRANCH}) {DIRTY}")
print(f"  version: system={VERSION_SYSTEM} workflow={VERSION_WORKFLOW}")
print(f"  sync: {SYNC_TS}")
print(f"  apply: {bool(APPLY)}")
print(f"  scope: {SCOPE_DIRS}")
print()

# ----- collect raw files from Phase 3 scope -----
raw_files = []
for sub in SCOPE_DIRS:
    raw_subdir = f"{RAW_DIR}/docs/{sub}"
    if not os.path.isdir(raw_subdir):
        print(f"  SKIP (no raw/ dir): docs/{sub}")
        continue
    pattern = os.path.join(raw_subdir, "**/*.md")
    for f in sorted(glob.glob(pattern, recursive=True)):
        raw_files.append((f, sub))

print(f"[wiki-mass-ingest] raw files found: {len(raw_files)}")
print()

# ----- Wiki page 생성 (1:1 mirror 정공법) -----
def make_frontmatter(raw_rel_path, fname, sub):
    """frontmatter 작성 (provenance + 표준 8 field + AD-specific tags)"""
    # type 결정
    if sub == "domain":
        # domain/ 뒤의 sub-domain 추출 (예: docs/domain/auth-session/api.md → auth-session)
        parts = raw_rel_path.split("/")
        if len(parts) >= 4:
            domain_name = parts[3]  # domain/<name>/<file>.md
        else:
            domain_name = "core"
        title = fname.replace("-", " ").replace(".md", "")
        return f"""---
title: {title}
type: source
tags: [domain, {domain_name}, project-devhub]
sources: [raw/projects/devhub/{raw_rel_path}]
git_commit: {COMMIT_SHORT}
git_branch: {BRANCH}
version_system: {VERSION_SYSTEM}
version_workflow: {VERSION_WORKFLOW}
last_touched: {SYNC_TS}
mirror_dirty: {DIRTY}
related: [none]
status: draft
contradictions: [none]
---"""
    elif sub == "architecture":
        title = fname.replace("-", " ").replace(".md", "")
        return f"""---
title: {title}
type: source
tags: [architecture, project-devhub]
sources: [raw/projects/devhub/{raw_rel_path}]
git_commit: {COMMIT_SHORT}
git_branch: {BRANCH}
version_system: {VERSION_SYSTEM}
version_workflow: {VERSION_WORKFLOW}
last_touched: {SYNC_TS}
mirror_dirty: {DIRTY}
related: [none]
status: draft
contradictions: [none]
---"""
    elif sub == "infrastructure":
        # infrastructure/<sub-category>/<file>.md
        parts = raw_rel_path.split("/")
        if len(parts) >= 4:
            infra_name = parts[3]
        else:
            infra_name = "core"
        title = fname.replace("-", " ").replace(".md", "")
        return f"""---
title: {title}
type: source
tags: [infrastructure, {infra_name}, project-devhub]
sources: [raw/projects/devhub/{raw_rel_path}]
git_commit: {COMMIT_SHORT}
git_branch: {BRANCH}
version_system: {VERSION_SYSTEM}
version_workflow: {VERSION_WORKFLOW}
last_touched: {SYNC_TS}
mirror_dirty: {DIRTY}
related: [none]
status: draft
contradictions: [none]
---"""
    elif sub == "validation":
        title = fname.replace("-", " ").replace(".md", "")
        return f"""---
title: {title}
type: source
tags: [validation, project-devhub]
sources: [raw/projects/devhub/{raw_rel_path}]
git_commit: {COMMIT_SHORT}
git_branch: {BRANCH}
version_system: {VERSION_SYSTEM}
version_workflow: {VERSION_WORKFLOW}
last_touched: {SYNC_TS}
mirror_dirty: {DIRTY}
related: [none]
status: draft
contradictions: [none]
---"""
    else:
        return None

def extract_raw_body(raw_path):
    """raw/ 의 md 에서 frontmatter 제거 + body 추출 (1:1 mirror 정공법)"""
    with open(raw_path) as f:
        content = f.read()

    lines = content.split("\n")
    if lines[0].strip() == "---":
        end_idx = None
        for i, line in enumerate(lines[1:], 1):
            if line.strip() == "---":
                end_idx = i
                break
        if end_idx is not None:
            return "\n".join(lines[end_idx+1:]).lstrip("\n")
    return content

# Wiki page path 결정
def get_wiki_path(raw_path):
    """raw/ 의 path → wiki/ 의 sources/ 의 path"""
    rel = os.path.relpath(raw_path, RAW_DIR)
    return f"{WIKI_SOURCES}/{rel}"

# Process each file
CREATED = 0
UPDATED = 0
SKIPPED = 0
new_pages = []
for raw_path, sub in raw_files:
    rel = os.path.relpath(raw_path, RAW_DIR)
    wiki_path = get_wiki_path(raw_path)
    fname = os.path.basename(raw_path)

    frontmatter = make_frontmatter(rel, fname, sub)
    if frontmatter is None:
        SKIPPED += 1
        continue

    body = extract_raw_body(raw_path)
    new_content = frontmatter + "\n\n" + body

    existing = os.path.exists(wiki_path)
    if existing:
        # Check if already has same content
        with open(wiki_path) as f:
            current = f.read()
        if current == new_content:
            SKIPPED += 1
            continue
        if APPLY:
            with open(wiki_path, "w") as f:
                f.write(new_content)
            UPDATED += 1
        else:
            UPDATED += 1  # dry-run count
    else:
        if APPLY:
            os.makedirs(os.path.dirname(wiki_path), exist_ok=True)
            with open(wiki_path, "w") as f:
                f.write(new_content)
            CREATED += 1
        else:
            CREATED += 1  # dry-run count
        new_pages.append(rel)

# ----- wiki index.md 갱신 -----
INDEX_PATH = f"{WIKI_SOURCES[:-8]}/index.md"  # ../index.md
INDEX_PATH = f"{WIKI_SOURCES}/../index.md"
# Hmm, let's check
import os.path
candidate = os.path.dirname(WIKI_SOURCES) + "/index.md"
if os.path.exists(candidate):
    INDEX_PATH = candidate
else:
    INDEX_PATH = None

if APPLY and INDEX_PATH and os.path.exists(INDEX_PATH):
    with open(INDEX_PATH) as f:
        idx_content = f.read()

    added = 0
    for rel in new_pages:
        # Wiki sources/ 의 path
        wiki_source_path = f"sources/{rel}"
        if wiki_source_path in idx_content:
            continue  # already in index
        # Add a new line
        fname = os.path.basename(rel)
        new_line = f"- [wiki/projects/devhub/{wiki_source_path}]({wiki_source_path}) — ingested\n"
        # Insert alphabetically
        lines = idx_content.split("\n")
        inserted = False
        for i, line in enumerate(lines):
            if line.startswith("- [wiki/projects/devhub/sources/") and line > new_line:
                lines.insert(i, new_line.rstrip("\n"))
                inserted = True
                break
        if not inserted:
            lines.append(new_line.rstrip("\n"))
        idx_content = "\n".join(lines)
        added += 1

    if added > 0:
        with open(INDEX_PATH, "w") as f:
            f.write(idx_content)
        print(f"[wiki-mass-ingest] index.md: {added} new entries added")
    else:
        print(f"[wiki-mass-ingest] index.md: 0 new entries (all already present)")

print()
print(f"[wiki-mass-ingest] DONE")
print(f"  created: {CREATED} page(s)")
print(f"  updated: {UPDATED} page(s)")
print(f"  skipped: {SKIPPED} page(s) (no changes)")
print(f"  mode: {'APPLY' if APPLY else 'DRY-RUN'}")
PYEOF
