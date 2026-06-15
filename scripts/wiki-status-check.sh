#!/usr/bin/env bash
# wiki-status-check.sh — wiki 의 모든 정보의 상태(provenance) 검증.
#
# 사용:
#   bash scripts/wiki-status-check.sh              # full status report
#   bash scripts/wiki-status-check.sh --stale       # stale pages only (last_touched < commit)
#   bash scripts/wiki-status-check.sh --diff       # manifest vs current git drift
#   bash scripts/wiki-status-check.sh --json       # JSON output
#
# 본 script 의 source-of-truth:
#   - mirror manifest: ai-workflow/wiki/raw/projects/devhub/_manifest.md
#   - 위키 sources/ pages: ai-workflow/wiki/wiki/projects/devhub/sources/*.md
#   - 본 저장소 의 git: HEAD commit + dirty flag
#
# 본 script 의 정공법:
#   - 위키의 모든 sources/ page 의 frontmatter 의 git_commit + version 정보 + last_touched
#     와 manifest 의 정보를 비교하여 위키가 stale 한지 검증.
#   - 본 저장소 의 git HEAD 와 manifest 의 commit 이 다른지 검증.
#   - 위키가 어느 commit 의 어느 시점의 정보인지 한눈에 파악.
#
# 결정적이고 단순: Python + 정규식 frontmatter parser.
#
# D-72 Phase 1.5 + provenance tracking (2026-06-13) 의 본 저장소 측 script.
#
# Exit code:
#   0 — success
#   1 — manifest 부재 / wiki dir 부재 / source-of-truth drift

set -euo pipefail

# ----- options -----
MODE="all"
for arg in "$@"; do
  case "$arg" in
    --stale)
      MODE="stale"
      ;;
    --diff)
      MODE="diff"
      ;;
    --json)
      MODE="json"
      ;;
    -h|--help)
      echo "usage: bash scripts/wiki-status-check.sh [--stale|--diff|--json]"
      echo "  (default)  full status report"
      echo "  --stale    stale pages only (last_touched < commit)"
      echo "  --diff     manifest vs current git drift"
      echo "  --json     JSON output"
      exit 0
      ;;
    *)
      echo "[wiki-status-check] invalid option: $arg" >&2
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
  echo "[wiki-status-check] manifest 부재: $MANIFEST" >&2
  exit 1
fi

if [[ ! -d "$WIKI_SOURCES" ]]; then
  echo "[wiki-status-check] wiki sources/ 부재: $WIKI_SOURCES" >&2
  exit 1
fi

# ----- extract provenance from manifest -----
COMMIT_FULL=$(grep -E "^\| commit \(full\) \|" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
COMMIT_SHORT=$(grep -E "^\| commit \(short\) \|" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
BRANCH=$(grep -E "^\| branch \|" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
DIRTY=$(grep -E "^\| dirty \|" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
VERSION_SYSTEM=$(grep -E "^\| system version" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
VERSION_WORKFLOW=$(grep -E "^\| workflow version" "$MANIFEST" | head -1 | sed 's/^.*| //' | tr -d '| ' | xargs)
SYNC_TIMESTAMP=$(grep -E "^\- timestamp: " "$MANIFEST" | head -1 | sed 's/^- timestamp: //' | xargs)

# ----- current git state -----
GIT_HEAD=$(git -C "$SRC" rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_HEAD_SHORT=$(git -C "$SRC" rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH_CUR=$(git -C "$SRC" branch --show-current 2>/dev/null || echo "unknown")
GIT_DIRTY_CUR=$(git -C "$SRC" status --porcelain 2>/dev/null | head -1)
if [[ -n "$GIT_DIRTY_CUR" ]]; then
  GIT_DIRTY_CUR_FLAG="(dirty: uncommitted changes)"
else
  GIT_DIRTY_CUR_FLAG=""
fi

# ----- mode: full report (default) -----
if [[ "$MODE" == "all" ]]; then
  echo "=== Wiki Status Report ==="
  echo ""
  echo "## Mirror manifest provenance"
  echo ""
  echo "| field | value |"
  echo "| --- | --- |"
  echo "| commit (full) | $COMMIT_FULL |"
  echo "| commit (short) | $COMMIT_SHORT |"
  echo "| branch | $BRANCH |"
  echo "| dirty | $DIRTY |"
  echo "| version (system) | $VERSION_SYSTEM |"
  echo "| version (workflow) | $VERSION_WORKFLOW |"
  echo "| last sync | $SYNC_TIMESTAMP |"
  echo ""
  echo "## Current git state"
  echo ""
  echo "| field | value |"
  echo "| --- | --- |"
  echo "| HEAD | $GIT_HEAD_SHORT |"
  echo "| branch | $GIT_BRANCH_CUR |"
  echo "| dirty | $GIT_DIRTY_CUR_FLAG |"
  echo ""

  # Drift check
  if [[ "$COMMIT_SHORT" != "$GIT_HEAD_SHORT" ]]; then
    echo "## ⚠️  DRIFT DETECTED"
    echo ""
    echo "  manifest commit: $COMMIT_SHORT"
    echo "  current HEAD:    $GIT_HEAD_SHORT"
    echo "  → mirror stale: re-run 'bash scripts/wiki-sync-devhub.sh' to update manifest"
    echo ""
  else
    echo "## ✅ Manifest matches current HEAD"
    echo ""
  fi

  # Wiki pages status (recursive: include sub-directory Phase 3 pages)
  echo "## Wiki pages status"
  echo ""
  total=0
  matched=0
  stale=0
  while IFS= read -r page; do
    [[ ! -f "$page" ]] && continue
    [[ "$(basename "$page")" == "_manifest.md" ]] && continue
    total=$((total + 1))

    if head -1 "$page" | grep -q "^---$"; then
      page_commit=$(awk 'NR>1 && /^---$/{exit} /^git_commit:/{print $2}' "$page")
      if [[ "$page_commit" == "$COMMIT_SHORT" ]]; then
        matched=$((matched + 1))
      else
        stale=$((stale + 1))
      fi
    else
      stale=$((stale + 1))
    fi
  done < <(find "$WIKI_SOURCES" -name "*.md" -type f)
  
  echo "| status | count |"
  echo "| --- | --- |"
  echo "| matched (commit=$COMMIT_SHORT) | $matched |"
  echo "| stale | $stale |"
  echo "| total | $total |"
  echo ""
  exit 0
fi

# ----- mode: stale -----
if [[ "$MODE" == "stale" ]]; then
  echo "=== Stale wiki pages ==="
  echo ""
  echo "manifest commit: $COMMIT_SHORT"
  echo ""
  stale_count=0
  while IFS= read -r page; do
    [[ ! -f "$page" ]] && continue
    [[ "$(basename "$page")" == "_manifest.md" ]] && continue

    if ! head -1 "$page" | grep -q "^---$"; then
      echo "  STALE (no frontmatter): ${page#$WIKI_SOURCES/}"
      stale_count=$((stale_count + 1))
      continue
    fi

    page_commit=$(awk 'NR>1 && /^---$/{exit} /^git_commit:/{print $2}' "$page")
    if [[ "$page_commit" != "$COMMIT_SHORT" ]]; then
      echo "  STALE (commit=$page_commit, expected=$COMMIT_SHORT): ${page#$WIKI_SOURCES/}"
      stale_count=$((stale_count + 1))
    fi
  done < <(find "$WIKI_SOURCES" -name "*.md" -type f)
  echo ""
  echo "Total stale: $stale_count"
  exit 0
fi

# ----- mode: diff -----
if [[ "$MODE" == "diff" ]]; then
  echo "=== Manifest vs current git drift ==="
  echo ""
  if [[ "$COMMIT_SHORT" != "$GIT_HEAD_SHORT" ]]; then
    echo "⚠️  DRIFT DETECTED"
    echo ""
    echo "  manifest: $COMMIT_SHORT"
    echo "  HEAD:     $GIT_HEAD_SHORT"
    echo ""
    echo "  → mirror stale: 'bash scripts/wiki-sync-devhub.sh' to update"
    exit 1
  else
    echo "✅ No drift: manifest matches HEAD"
    echo ""
    echo "  HEAD:     $GIT_HEAD_SHORT"
    echo "  manifest: $COMMIT_SHORT"
    echo "  branch:   $GIT_BRANCH_CUR"
    exit 0
  fi
fi

# ----- mode: json -----
if [[ "$MODE" == "json" ]]; then
  cat <<EOF
{
  "manifest": {
    "commit_full": "$COMMIT_FULL",
    "commit_short": "$COMMIT_SHORT",
    "branch": "$BRANCH",
    "dirty": "$DIRTY",
    "version_system": "$VERSION_SYSTEM",
    "version_workflow": "$VERSION_WORKFLOW",
    "sync_timestamp": "$SYNC_TIMESTAMP"
  },
  "git": {
    "head": "$GIT_HEAD",
    "head_short": "$GIT_HEAD_SHORT",
    "branch": "$GIT_BRANCH_CUR",
    "dirty": "$GIT_DIRTY_CUR_FLAG"
  },
  "drift": $([ "$COMMIT_SHORT" != "$GIT_HEAD_SHORT" ] && echo "true" || echo "false")
}
EOF
  exit 0
fi
