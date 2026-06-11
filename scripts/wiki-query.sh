#!/usr/bin/env bash
# wiki-query.sh — DevHub repo (~/repos/Devhub_example_omp/) 의 vault query wrapper.
# vault (`~/wiki/`) 의 LLM wiki 패턴 (D-71, AGENTS.md v1.5) 의 §2.2 Query 오퍼레이션
# 6 step 자동화 (read + query/ 페이지 file + log.md append).
#
# 사용:
#   bash scripts/wiki-query.sh --query "Keycloak RBAC"             # read-only (default)
#   bash scripts/wiki-query.sh --query "ADR-0020" --file           # read + query/ 페이지 file
#   bash scripts/wiki-query.sh --query "rbac" --tag rbac --limit 5  # tag filter + limit
#   bash scripts/wiki-query.sh --query "Keycloak" --format json    # JSON output
#
# 본 script 의 source-of-truth:
#   - my_harness 측 SSOT: ~/repos/my_harness/ai-workflow/core/wiki_query_skill_spec.md
#   - my_harness skill:    ~/repos/my_harness/ai-workflow/skills/wiki-query/
#   - vault 운영 규약:     ~/wiki/AGENTS.md (v1.5, D-71) 의 §2.2 Query
#
# 결정적 단순: read 가 main. write 는 --file 옵션.
#   1. read (my_harness 의 run_wiki_query.py) — frontmatter + wikilink + full-text 검색
#   2. (--file) write (my_harness 의 run_wiki_query.py --file) — query/ 페이지 신규 + log.md 1 line append
#
# 본 script 의 본 저장소 (= DevHub) 측 책임:
#   - read + (optional) write 의 통합 entry point
#   - my_harness 측 skill 가 부재 시 명확한 에러
#   - --no-file / --file 의 user confirm flow 일관성
#
# Exit code:
#   0 — success (read 또는 read+write 모두 성공, 0 results 도 success)
#   1 — read 실패, 또는 write 실패, 또는 my_harness skill 부재
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
MYHARNESS_ROOT="${MYHARNESS_ROOT:-$HOME/repos/my_harness}"
VAULT_ROOT="${VAULT_ROOT:-$HOME/wiki}"
WIKI_QUERY_SCRIPT="$MYHARNESS_ROOT/ai-workflow/skills/wiki-query/scripts/run_wiki_query.py"

usage() {
  cat <<'EOF'
Usage: bash scripts/wiki-query.sh [options]

Options:
  --query <text>            Required. 검색어 (full-text / wikilink / frontmatter key 모두 매칭).
  --project <name>          devhub|my-harness. default: devhub.
  --tag <tag>               frontmatter tags: 필터 (AND).
  --type <concept|entity|topic|source|comparison|query>  frontmatter type: 필터.
  --limit N                 최대 결과 수. default: 20.
  --format <md|json|plain>  출력 형식. default: md.
  --file                    query/ 페이지 자동 file + log.md 1 line append (AGENTS.md §2.2 step 4-5).
                            default: --no-file (read-only).
  --no-file                 read-only. default.
  --quiet                   stderr 메시지 최소화.
  -h, --help                도움말.

Examples:
  # 1. read-only (default) — agent 가 결과 받아 활용
  bash scripts/wiki-query.sh --query "Keycloak RBAC"

  # 2. read + query/ 페이지 file (AGENTS.md §2.2 6 step 자동)
  bash scripts/wiki-query.sh --query "ADR-0020 결정 사항" --file

  # 3. tag filter + limit
  bash scripts/wiki-query.sh --query "rbac" --tag rbac --limit 5

  # 4. JSON output (다른 tool 입력용)
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
if [[ "$PROJECT" != "devhub" && "$PROJECT" != "my-harness" ]]; then
  echo "[wiki-query] error: invalid --project: $PROJECT (must be devhub|my-harness)" >&2
  exit 2
fi
if [[ "$FORMAT" != "md" && "$FORMAT" != "json" && "$FORMAT" != "plain" ]]; then
  echo "[wiki-query] error: invalid --format: $FORMAT (must be md|json|plain)" >&2
  exit 2
fi

# ----- my_harness skill 부재 확인 -----
if [[ ! -f "$WIKI_QUERY_SCRIPT" ]]; then
  echo "[wiki-query] error: wiki-query skill 미설치: $WIKI_QUERY_SCRIPT" >&2
  echo "[wiki-query]   my_harness 측 SSOT: ~/repos/my_harness/ai-workflow/skills/wiki-query/" >&2
  echo "[wiki-query]   또는 MYHARNESS_ROOT 환경 변수로 경로 명시" >&2
  exit 1
fi

# ----- step 1 (read) + step 2 (optional write) -----
log() { [[ $QUIET -eq 1 ]] || echo "$@"; }

log "[wiki-query] step: $FILE_MODE"
log "[wiki-query]   vault: $VAULT_ROOT"
log "[wiki-query]   project: $PROJECT"
log "[wiki-query]   query: $QUERY"
[[ -n "$TAG" ]] && log "[wiki-query]   tag filter: $TAG"
[[ -n "$TYPE" ]] && log "[wiki-query]   type filter: $TYPE"
[[ -n "$LIMIT" ]] && log "[wiki-query]   limit: $LIMIT"
log "[wiki-query]   format: $FORMAT"

QUERY_ARGS=(
  --vault-path "$VAULT_ROOT"
  --project "$PROJECT"
  --query "$QUERY"
  --format "$FORMAT"
)
[[ -n "$TAG" ]] && QUERY_ARGS+=(--tag "$TAG")
[[ -n "$TYPE" ]] && QUERY_ARGS+=(--type "$TYPE")
[[ -n "$LIMIT" ]] && QUERY_ARGS+=(--limit "$LIMIT")
if [[ "$FILE_MODE" == "file" ]]; then
  QUERY_ARGS+=(--file)
fi
[[ $QUIET -eq 1 ]] && QUERY_ARGS+=(--quiet)

log "[wiki-query]   command: python3 $WIKI_QUERY_SCRIPT ${QUERY_ARGS[*]}"
python3 "$WIKI_QUERY_SCRIPT" "${QUERY_ARGS[@]}"
QUERY_EXIT=$?

if [[ $QUERY_EXIT -ne 0 ]]; then
  echo "[wiki-query] error: vault query failed (exit $QUERY_EXIT)" >&2
  exit 1
fi

log "[wiki-query] DONE"
log "[wiki-query]   results: stdout (--format=$FORMAT)"
if [[ "$FILE_MODE" == "file" ]]; then
  log "[wiki-query]   query/ page: $VAULT_ROOT/wiki/projects/$PROJECT/query/$(date -u +%Y-%m-%d)-<topic>.md (1 file created)"
  log "[wiki-query]   log.md: $VAULT_ROOT/log.md (1 line appended)"
  log "[wiki-query]   filed as: [[query/$(date -u +%Y-%m-%d)-<topic>]] (AGENTS.md §2.2 step 3)"
fi
