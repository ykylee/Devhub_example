#!/usr/bin/env bash
# wiki-sync-devhub.sh — DevHub repo (~/repos/Devhub_example_minimax/) 의 의미 있는 파일을
# ~/wiki/raw/projects/devhub/ 로 mirror. my_harness 의 wiki-sync-ai-workflow.sh 와 동일 패턴.
#
# 사용:
#   bash scripts/wiki-sync-devhub.sh           # real mirror
#   bash scripts/wiki-sync-devhub.sh --dry-run # source list 출력 (no actual mirror)
#
# 본 script 의 source-of-truth:
#   - mirror list: docs/llm-wiki/mirror-list.md (§3 의 source 패턴)
#   - lint config: docs/llm-wiki/lint-config.toml
#   - operation SOP: docs/llm-wiki/operation-sop.md
#
# 결정적이고 단순: find + mkdir -p + cp. rsync 의 include/exclude 충돌 회피 (my_harness 의
# wiki-sync-ai-workflow.sh 와 동일 pattern).
#
# D-72 Phase 1 (2026-06-10) 의 본 저장소 측 sync script. lint L11 (사내 패턴 검출) +
# sa-internal/ 격리 + mirror policy (T-d-72-2) 의 source.
#
# Exit code:
#   0 — success (real mirror 또는 dry-run)
#   1 — vault 부재 또는 source root 부재 또는 invalid option

set -euo pipefail

# ----- options -----
DRY_RUN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=1
      ;;
    -h|--help)
      echo "usage: bash scripts/wiki-sync-devhub.sh [--dry-run]"
      echo "  --dry-run  source list 출력 (no actual mirror)"
      exit 0
      ;;
    *)
      echo "[wiki-sync-devhub] invalid option: $arg" >&2
      echo "  hint: --dry-run 또는 (no option) 만 지원" >&2
      exit 1
      ;;
  esac
done

# ----- paths -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${HOME}/wiki"
DEST="$VAULT_ROOT/raw/projects/devhub"

# ----- validation -----
if [[ ! -d "$SRC" ]]; then
  echo "[wiki-sync-devhub] source root 부재: $SRC" >&2
  exit 1
fi

if [[ ! -d "$VAULT_ROOT" ]]; then
  echo "[wiki-sync-devhub] target vault not found: $VAULT_ROOT" >&2
  echo "[wiki-sync-devhub] hint: 'wiki-init' 명령으로 vault 초기화 (my_harness 측 D-71 §2.2)" >&2
  echo "[wiki-sync-devhub] hint: 또는 my_harness 의 D-72 응답 §4 의 next-step #1 참고" >&2
  exit 1
fi

echo "[wiki-sync-devhub] source root: $SRC"
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[wiki-sync-devhub] dry-run: True (no actual mirror)"
else
  echo "[wiki-sync-devhub] dry-run: False (real mirror)"
fi
echo "[wiki-sync-devhub] target vault: $VAULT_ROOT (Gitea private)"

# ----- mirror list (Phase 1 source patterns) -----
# 7 패턴 (docs/llm-wiki/mirror-list.md §3):
#   1. ADR: docs/adr/0[0-9][0-9][0-9]-*.md
#   2. Governance: docs/governance/*.md
#   3. Planning: docs/planning/*.md
#   4. Setup: docs/setup/*.md
#   5. Requirements: docs/requirements.md
#   6. OpenAPI: docs/openapi.yaml
#   7. AI-workflow memory (main flat): ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}

# ----- helper: list all mirror sources -----
list_sources() {
  find "$SRC/docs/adr" -type f -name "0[0-9][0-9][0-9]-*.md" 2>/dev/null
  find "$SRC/docs/governance" -type f -name "*.md" 2>/dev/null
  find "$SRC/docs/planning" -type f -name "*.md" 2>/dev/null
  find "$SRC/docs/setup" -type f -name "*.md" 2>/dev/null
  [[ -f "$SRC/docs/requirements.md" ]] && echo "$SRC/docs/requirements.md"
  [[ -f "$SRC/docs/openapi.yaml" ]] && echo "$SRC/docs/openapi.yaml"
  for f in state.json session_handoff.md work_backlog.md; do
    [[ -f "$SRC/ai-workflow/memory/$f" ]] && echo "$SRC/ai-workflow/memory/$f"
  done
}

# ----- 1. dry-run: source list 출력 -----
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[wiki-sync-devhub] collecting files (dry-run)..."

  echo ""
  echo "  ADR (docs/adr/0[0-9][0-9][0-9]-*.md):"
  find "$SRC/docs/adr" -type f -name "0[0-9][0-9][0-9]-*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Governance (docs/governance/*.md):"
  find "$SRC/docs/governance" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Planning (docs/planning/*.md):"
  find "$SRC/docs/planning" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Setup (docs/setup/*.md):"
  find "$SRC/docs/setup" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Requirements (docs/requirements.md):"
  [[ -f "$SRC/docs/requirements.md" ]] && echo "    docs/requirements.md"

  echo ""
  echo "  OpenAPI (docs/openapi.yaml):"
  [[ -f "$SRC/docs/openapi.yaml" ]] && echo "    docs/openapi.yaml"

  echo ""
  echo "  AI-workflow memory (main flat, 3 file):"
  for f in state.json session_handoff.md work_backlog.md; do
    [[ -f "$SRC/ai-workflow/memory/$f" ]] && echo "    ai-workflow/memory/$f"
  done

  COUNT=$(list_sources | wc -l | tr -d ' ')
  echo ""
  echo "[wiki-sync-devhub] total: $COUNT file (estimated)"
  echo "[wiki-sync-devhub] dry-run: no changes made"
  echo "[wiki-sync-devhub] PASS (dry-run)"
  exit 0
fi

# ----- 2. real mirror: DEST 의 clean mirror 후 cp + manifest 자동 -----
echo "[wiki-sync-devhub] cleaning $DEST"
mkdir -p "$DEST"
find "$DEST" -mindepth 1 -delete 2>/dev/null

echo "[wiki-sync-devhub] collecting files from $SRC"

COUNT=0
list_sources | while IFS= read -r f; do
  rel="${f#"$SRC"/}"
  dest_path="$DEST/$rel"
  mkdir -p "$(dirname "$dest_path")"
  cp -p "$f" "$dest_path"
  COUNT=$((COUNT + 1))
  echo "  copied: $rel"
done

# ----- manifest 자동 생성 -----
MANIFEST="$DEST/_manifest.md"
{
  echo "# raw/projects/devhub/_manifest.md (D-72 Phase 1)"
  echo ""
  echo "- source: $SRC"
  echo "- target: $DEST"
  echo "- vault: $VAULT_ROOT (Gitea private)"
  echo "- sync tool: scripts/wiki-sync-devhub.sh (BSD-rsync safe)"
  echo "- mirror list: docs/llm-wiki/mirror-list.md §3"
  echo "- lint config: docs/llm-wiki/lint-config.toml"
  echo ""
  echo "## Last sync"
  echo ""
  echo "- timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "- count: $(find "$DEST" -type f 2>/dev/null | wc -l | tr -d ' ') file"
  echo "- size: $(du -sh "$DEST" 2>/dev/null | cut -f1)"
  echo ""
  echo "## Mirror log (recent, 최신 위)"
  echo ""
  echo "| timestamp | rel_path | size (bytes) |"
  echo "| --- | --- | --- |"
  find "$DEST" -type f -not -name "_manifest.md" 2>/dev/null | sort | while IFS= read -r f; do
    rel="${f#"$DEST"/}"
    size=$(stat -f%z "$f" 2>/dev/null || stat -c%s "$f" 2>/dev/null || echo "?")
    echo "| $(date -u +%Y-%m-%dT%H:%M:%SZ) | $rel | $size |"
  done
} > "$MANIFEST"

echo "[wiki-sync-devhub] DONE"
echo "  files: $(find "$DEST" -type f 2>/dev/null | wc -l | tr -d ' ')"
echo "  size:  $(du -sh "$DEST" 2>/dev/null | cut -f1)"
echo "  manifest: $MANIFEST"
