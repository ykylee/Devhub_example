#!/usr/bin/env bash
# wiki-query.sh — DevHub 의 in-repo wiki L1/L2 query wrapper.
# vault (ai-workflow/wiki/) 의 LLM wiki 패턴 (D-71, AGENTS.md v1.5) 의 §2.2 Query 오퍼레이션
# 6 step 자동화 (read + query/ 페이지 file + log.md append).
#
# 사용:
#   bash scripts/wiki-query.sh --query "Keycloak RBAC"             # read-only (default)
#   bash scripts/wiki-query.sh --query "ADR-0020" --file           # read + query/ 페이지 file
#   bash scripts/wiki-query.sh --query "rbac" --tag rbac --limit 5  # tag filter + limit
#   bash scripts/wiki-query.sh --query "Keycloak" --format json    # JSON output
#
# 본 script 의 source-of-truth (v0.7.17+ in-repo redirect, 2026-06-15):
#   - in-repo wiki:        ai-workflow/wiki/ (L0 Home + L1 + L2 dense)
#   - 자체 query 도구:      python3 inline grep + frontmatter parsing (in-repo, no external skill)
#   - log target:          ai-workflow/memory/log.md (vendor 의 wiki-log target)
#
# **Deprecated (2026-06-15)**: my_harness 측 run_wiki_query.py 호출 제거. in-repo 만 운영.
# vendor 의 core/ 에 *wiki-query* 동등 도구 미존재 → 자체 inline grep + frontmatter parsing 으로 대체.
#
# 결정적 단순: read 가 main. write 는 --file 옵션.
#   1. read (자체 inline python) — frontmatter + wikilink + full-text 검색
#   2. (--file) write (자체 inline python) — query/ 페이지 신규 + log.md 1 line append
#
# 본 script 의 본 저장소 (= DevHub) 측 책임:
#   - read + (optional) write 의 통합 entry point
#   - in-repo 만 (외부 vault ~/wiki/ 미사용, my_harness 미참조)
#   - --no-file / --file 의 user confirm flow 일관성
#
# Exit code:
#   0 — success (read 또는 read+write 모두 성공, 0 results 도 success)
#   1 — read 실패, 또는 write 실패
#   2 — invalid option 또는 required option (--query) 부재

set -euo pipefail

# ----- options -----
QUERY=""
PROJECT="devhub"
TAG=""
TYPE=""
LIMIT=""
FORMAT="md"
FILE_MODE="no-file"   # no-file (default) | file
QUIET=0
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${VAULT_ROOT:-${SRC}/ai-workflow/wiki}"
LOG_TARGET="${SRC}/ai-workflow/memory/log.md"

usage() {
  cat <<'EOF'
Usage: bash scripts/wiki-query.sh [options]

Options:
  --query <text>            Required. 검색어 (full-text / wikilink / frontmatter key 모두 매칭).
  --project <name>          Default: devhub. (in-repo 만, my_harness 미참조)
  --tag <tag>               frontmatter tags: 필터 (AND).
  --type <concept|decision|entity|pattern|topic>  frontmatter type: 필터.
  --limit N                 최대 결과 수. default: 20.
  --format <md|json|plain>  출력 형식. default: md.
  --file                    query/ 페이지 자동 file + log.md 1 line append (AGENTS.md §2.2 step 4-5).
                            default: --no-file (read-only).
  --no-file                 read-only. default.
  --quiet                   stderr 메시지 최소화.
  -h, --help                도움말.

Examples:
  # 1. read-only (default)
  bash scripts/wiki-query.sh --query "Keycloak RBAC"

  # 2. read + query/ 페이지 file
  bash scripts/wiki-query.sh --query "ADR-0020 결정 사항" --file

  # 3. tag filter + limit
  bash scripts/wiki-query.sh --query "rbac" --tag rbac --limit 5

  # 4. JSON output
  bash scripts/wiki-query.sh --query "Keycloak" --format json --no-file
EOF
}

# ----- parse options -----
while [[ $# -gt 0 ]]; do
  case "$1" in
    --query)
      QUERY="$2"
      shift 2
      ;;
    --project)
      PROJECT="$2"
      shift 2
      ;;
    --tag)
      TAG="$2"
      shift 2
      ;;
    --type)
      TYPE="$2"
      shift 2
      ;;
    --limit)
      LIMIT="$2"
      shift 2
      ;;
    --format)
      FORMAT="$2"
      shift 2
      ;;
    --file)
      FILE_MODE="file"
      shift
      ;;
    --no-file)
      FILE_MODE="no-file"
      shift
      ;;
    --quiet)
      QUIET=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[wiki-query] error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

# ----- validation -----
if [[ -z "$QUERY" ]]; then
  echo "[wiki-query] error: --query required" >&2
  exit 2
fi
if [[ "$PROJECT" != "devhub" ]]; then
  echo "[wiki-query] error: invalid --project: $PROJECT (must be devhub, in-repo only)" >&2
  echo "[wiki-query] hint: 2026-06-15+ 결정 — my_harness wiki-* skill 미참조. in-repo 만 운영." >&2
  exit 2
fi
if [[ "$FORMAT" != "md" && "$FORMAT" != "json" && "$FORMAT" != "plain" ]]; then
  echo "[wiki-query] error: invalid --format: $FORMAT (must be md|json|plain)" >&2
  exit 2
fi

# ----- vault 부재 확인 -----
if [[ ! -d "$VAULT_ROOT" ]]; then
  echo "[wiki-query] error: vault 부재: $VAULT_ROOT" >&2
  exit 1
fi

# ----- log helper -----
log() { [[ $QUIET -eq 1 ]] || echo "$@"; }

# ----- read (in-repo, inline python) -----
log "[wiki-query] step: $FILE_MODE"
log "[wiki-query]   vault: $VAULT_ROOT (in-repo, v0.7.17+)"
log "[wiki-query]   project: $PROJECT"
log "[wiki-query]   query: $QUERY"
[[ -n "$TAG" ]] && log "[wiki-query]   tag filter: $TAG"
[[ -n "$TYPE" ]] && log "[wiki-query]   type filter: $TYPE"
[[ -n "$LIMIT" ]] && log "[wiki-query]   limit: $LIMIT"
log "[wiki-query]   format: $FORMAT"

# query Python 도구 (in-repo, inline): grep + frontmatter + wikilink + full-text
QUERY_RESULTS=$(QUERY="$QUERY" PROJECT="$PROJECT" TAG="$TAG" TYPE="$TYPE" LIMIT="$LIMIT" FORMAT="$FORMAT" VAULT_ROOT="$VAULT_ROOT" python3 <<'PYEOF'
"""wiki-query in-repo: in-repo L1/L2 page 의 frontmatter + wikilink + full-text 검색."""
import os
import re
import sys
from pathlib import Path

vault = Path(os.environ["VAULT_ROOT"])
project = os.environ["PROJECT"]
query = os.environ["QUERY"].lower()
tag_filter = os.environ.get("TAG", "").strip()
type_filter = os.environ.get("TYPE", "").strip()
limit = int(os.environ.get("LIMIT") or "20")
fmt = os.environ.get("FORMAT", "md")

# L1: concepts/decisions/entities/patterns/topics/*.md
# L2: sources/*.md
# (raw/ 는 제외 — 1:1 byte mirror, LLM query 대상 X)
l1_dirs = ["concepts", "decisions", "entities", "patterns", "topics"]
l2_dirs = ["sources"]

results = []
for sub in l1_dirs + l2_dirs:
    d = vault / sub
    if not d.is_dir():
        continue
    for md in sorted(d.rglob("*.md")):
        if md.name in ("index.md", "RAW_MIRROR_MANIFEST.md", ".gitkeep"):
            continue
        text = md.read_text(encoding="utf-8", errors="ignore")
        # frontmatter parse
        fm_match = re.match(r"^---\n(.+?)\n---", text, re.DOTALL)
        if fm_match:
            fm = fm_match.group(1)
            type_field = ""
            status_field = ""
            tags_field = ""
            for line in fm.split("\n"):
                if line.startswith("type:"):
                    type_field = line.split(":", 1)[1].strip()
                elif line.startswith("status:"):
                    status_field = line.split(":", 1)[1].strip()
                elif line.startswith("tags:"):
                    tags_field = line.split(":", 1)[1].strip()
            # filter
            if type_filter and type_field != type_filter:
                continue
            if tag_filter and tag_filter not in tags_field:
                continue
            # full-text 매칭
            content_lower = text.lower()
            if query in content_lower:
                rel = md.relative_to(vault)
                title_match = re.search(r"^#\s+(.+)$", text, re.MULTILINE)
                title = title_match.group(1).strip() if title_match else md.stem
                results.append({
                    "path": str(rel),
                    "type": type_field or "n/a",
                    "status": status_field or "n/a",
                    "title": title,
                    "snippet": text[:200].replace("\n", " "),
                })

# limit
results = results[:limit]

# output
if fmt == "json":
    import json
    print(json.dumps({"results": results, "count": len(results), "query": query, "project": project}, ensure_ascii=False, indent=2))
elif fmt == "plain":
    for r in results:
        print(f"{r['path']}\t{r['title']}")
else:  # md (default)
    print(f"# Query results: {query} (project={project}, count={len(results)})")
    print("")
    for r in results:
        print(f"## {r['title']}")
        print(f"- **path**: `ai-workflow/wiki/{r['path']}`")
        print(f"- **type**: {r['type']} / **status**: {r['status']}")
        print(f"- **snippet**: {r['snippet']}...")
        print("")
PYEOF
)

# emit
echo "$QUERY_RESULTS"

QUERY_EXIT=$?
if [[ $QUERY_EXIT -ne 0 ]]; then
  echo "[wiki-query] error: vault query failed (exit $QUERY_EXIT)" >&2
  exit 1
fi

# ----- (--file) write: query/ 페이지 + log.md append -----
if [[ "$FILE_MODE" == "file" ]]; then
  log ""
  log "[wiki-query] step 2: query/ page file + log.md append (AGENTS.md §2.2 step 4-5)"

  QUERY_DATE=$(date -u +%Y-%m-%d)
  QUERY_SLUG=$(echo "$QUERY" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9-' '-' | sed 's/^-\|-$//g' | head -c 50)
  QUERY_DIR="$VAULT_ROOT/query/$PROJECT/query"
  QUERY_PAGE="$QUERY_DIR/$QUERY_DATE-$QUERY_SLUG.md"

  mkdir -p "$QUERY_DIR"
  cat > "$QUERY_PAGE" <<EOF
# Query: $QUERY (project=$PROJECT, $QUERY_DATE)

## Context

- query: \`$QUERY\`
- project: \`$PROJECT\`
- vault: \`$VAULT_ROOT\` (in-repo, v0.7.17+)
- date: $QUERY_DATE
- file_mode: true (AGENTS.md §2.2 step 4-5)

## Results

$QUERY_RESULTS

## Source

- 본 query page: \`ai-workflow/wiki/query/$PROJECT/query/$QUERY_DATE-$QUERY_SLUG.md\`
- L0 Home: \`ai-workflow/wiki/index.md\`
EOF

  # log.md append (atomic, v0.7.15 follow-up D + #607 follow-up, P1 race fix)
  # 기존 atomic_write_text(read+append+os.replace) 는 두 process 동시 실행 시
  # read-modify-write race 로 한 줄 lost. atomic_append_text (O_APPEND + fsync) 로 교체.
  # POSIX 전용 (macOS/Linux). Windows fallback 으로 echo >> 유지.
  LOG_LINE="[$QUERY_DATE] query | $QUERY | project=$PROJECT | file=$QUERY_PAGE | results=$(echo "$QUERY_RESULTS" | grep -c "path" 2>/dev/null || echo 0)"
  mkdir -p "$(dirname "$LOG_TARGET")"
  _SCRIPT_DIR_LOG="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  if LOG_LINE="$LOG_LINE" LOG_TARGET="$LOG_TARGET" _SCRIPT_DIR_LOG="$_SCRIPT_DIR_LOG" python3 <<'PYEOF_LOG' 2>/dev/null
import os, sys
from pathlib import Path
sys.path.insert(0, os.environ['_SCRIPT_DIR_LOG'])
from atomic_write import atomic_append_text
target = Path(os.environ['LOG_TARGET'])
line = os.environ['LOG_LINE'] + '\n'
atomic_append_text(target, line)
PYEOF_LOG
  then
    :
  else
    # fallback: atomic_append_text 실패 시 (Windows 등) 기존 echo >> 로 graceful degrade
    echo "$LOG_LINE" >> "$LOG_TARGET"
  fi

  log "[wiki-query]   query/ page: $QUERY_PAGE (1 file created)"
  log "[wiki-query]   log.md: $LOG_TARGET (1 line appended)"
fi

log ""
log "[wiki-query] DONE"
log "[wiki-query]   results: stdout (--format=$FORMAT, count=stdout line 수)"
log "[wiki-query]   vault: $VAULT_ROOT (in-repo)"
log "[wiki-query]   my_harness 호출: 0 (2026-06-15+ 결정)"
