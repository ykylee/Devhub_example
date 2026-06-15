#!/usr/bin/env bash
# wiki-ingest-from-raw.sh — DevHub repo (~/repos/Devhub_example_minimax/) 의 raw mirror 결과 +
# my_harness 측 wiki-ingest-from-raw skill 호출 통합 wrapper.
#
# 사용:
#   bash scripts/wiki-ingest-from-raw.sh --dry-run                # raw mirror + wiki ingest preview
#   bash scripts/wiki-ingest-from-raw.sh --project devhub --apply # raw mirror + wiki ingest 실제 적용
#   bash scripts/wiki-ingest-from-raw.sh --source <rel> --apply   # 1 file 만 ingest
#
# 본 script 의 source-of-truth:
#   - my_harness 측 SSOT: ~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md
#   - my_harness skill:    ~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/
#   - DevHub mirror tool:  scripts/wiki-sync-devhub.sh
#   - vault 운영 규약:     ai-workflow/wiki/AGENTS.md (v1.5, D-71) 의 §2.1 Ingest
#
# 결정적 단순: 2 단계 wrapper.
#   1. raw mirror (wiki-sync-devhub.sh) — raw/ 갱신
#   2. wiki ingest (my_harness 의 run_wiki_ingest.py) — raw/ → wiki/ page 자동 작성
#
# 본 script 의 본 저장소 (= DevHub) 측 책임:
#   - raw/ 갱신 + wiki page 자동화 통합 entry point
#   - my_harness 측 skill 가 부재 시 명확한 에러
#   - dry-run / apply 의 user confirm flow 일관성
#
# Exit code:
#   0 — success (raw mirror + wiki ingest dry-run 또는 apply 모두 성공)
#   1 — raw mirror 실패, 또는 wiki ingest 실패, 또는 my_harness skill 부재
#   2 — invalid option 또는 required option (--project) 부재

set -euo pipefail

# ----- options -----
DRY_RUN=1
PROJECT=""
SOURCE=""
LIMIT=""
SKIP_LINT=0
QUIET=0
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
MYHARNESS_ROOT="${MYHARNESS_ROOT:-$HOME/repos/my_harness}"
VAULT_ROOT="${VAULT_ROOT:-${SRC}/ai-workflow/wiki}"
WIKI_INGEST_SCRIPT="$MYHARNESS_ROOT/ai-workflow/skills/wiki-ingest-from-raw/scripts/run_wiki_ingest.py"
WIKI_LINT_SCRIPT="$MYHARNESS_ROOT/ai-workflow/skills/wiki-lint/scripts/run_wiki_lint.py"

usage() {
  cat <<'EOF'
Usage: bash scripts/wiki-ingest-from-raw.sh [options]

Options:
  --project <devhub|my-harness>  Required. 대상 project.
  --source <rel_path>            1 file ingest (raw/ 상대 경로). 미지정 시 --all.
  --limit N                      --all 시 최대 N건.
  --apply                        실제 ingest (default = dry-run).
  --skip-lint                    post-ingest wiki-lint skip.
  --quiet                        stderr 메시지 최소화.
  -h, --help                     도움말.

Examples:
  # 1. dry-run preview (raw mirror + wiki ingest preview)
  bash scripts/wiki-ingest-from-raw.sh --project devhub

  # 2. 실제 ingest (raw mirror + wiki ingest apply)
  bash scripts/wiki-ingest-from-raw.sh --project devhub --apply

  # 3. 1 file 만 ingest
  bash scripts/wiki-ingest-from-raw.sh --project devhub --source docs/adr/0001-idp-selection.md --apply

  # 4. 5건만 부분 ingest (CI / 빠른 적용)
  bash scripts/wiki-ingest-from-raw.sh --project devhub --limit 5 --apply
EOF
}

# ----- parse options -----
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --apply)
      DRY_RUN=0
      shift
      ;;
    --project)
      PROJECT="$2"
      shift 2
      ;;
    --source)
      SOURCE="$2"
      shift 2
      ;;
    --limit)
      LIMIT="$2"
      shift 2
      ;;
    --skip-lint)
      SKIP_LINT=1
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
      echo "[wiki-ingest-from-raw] error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

# ----- validation -----
if [[ -z "$PROJECT" ]]; then
  echo "[wiki-ingest-from-raw] error: --project required (devhub|my-harness)" >&2
  exit 2
fi
if [[ "$PROJECT" != "devhub" && "$PROJECT" != "my-harness" ]]; then
  echo "[wiki-ingest-from-raw] error: invalid --project: $PROJECT (must be devhub|my-harness)" >&2
  exit 2
fi
if [[ -n "$SOURCE" && -n "$LIMIT" ]]; then
  echo "[wiki-ingest-from-raw] error: --source and --limit are mutually exclusive" >&2
  exit 2
fi

# ----- my_harness skill 부재 확인 -----
if [[ ! -f "$WIKI_INGEST_SCRIPT" ]]; then
  echo "[wiki-ingest-from-raw] error: wiki-ingest-from-raw skill 미설치: $WIKI_INGEST_SCRIPT" >&2
  echo "[wiki-ingest-from-raw]   my_harness 측 SSOT: ~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/" >&2
  echo "[wiki-ingest-from-raw]   또는 MYHARNESS_ROOT 환경 변수로 경로 명시" >&2
  exit 1
fi

# ----- step 1: raw mirror (DevHub) -----
log() { [[ $QUIET -eq 1 ]] || echo "$@"; }

log "[wiki-ingest-from-raw] step 1/2: raw mirror ($SRC → $VAULT_ROOT/raw/projects/$PROJECT/)"
log "[wiki-ingest-from-raw]   source root: $SRC"
log "[wiki-ingest-from-raw]   target vault: $VAULT_ROOT"
log "[wiki-ingest-from-raw]   mode: $([[ $DRY_RUN -eq 1 ]] && echo 'dry-run' || echo 'apply')"

if [[ $DRY_RUN -eq 1 ]]; then
  bash "$SCRIPT_DIR/wiki-sync-devhub.sh" --dry-run
else
  bash "$SCRIPT_DIR/wiki-sync-devhub.sh"
fi

# ----- step 2: wiki ingest (my_harness skill 호출) -----
log "[wiki-ingest-from-raw] step 2/2: wiki ingest (raw/projects/$PROJECT/ → wiki/projects/$PROJECT/)"
log "[wiki-ingest-from-raw]   skill: $WIKI_INGEST_SCRIPT"

INGEST_ARGS=(
  --vault-path "$VAULT_ROOT"
  --project "$PROJECT"
  --output both
)
if [[ -n "$SOURCE" ]]; then
  INGEST_ARGS+=(--source "$SOURCE")
else
  INGEST_ARGS+=(--all)
fi
if [[ -n "$LIMIT" ]]; then
  INGEST_ARGS+=(--limit "$LIMIT")
fi
if [[ $DRY_RUN -eq 0 ]]; then
  INGEST_ARGS+=(--apply)
fi
if [[ $SKIP_LINT -eq 1 ]]; then
  INGEST_ARGS+=(--skip-lint)
fi
if [[ $QUIET -eq 1 ]]; then
  INGEST_ARGS+=(--quiet)
fi

log "[wiki-ingest-from-raw]   command: python3 $WIKI_INGEST_SCRIPT ${INGEST_ARGS[*]}"
python3 "$WIKI_INGEST_SCRIPT" "${INGEST_ARGS[@]}"
INGEST_EXIT=$?

if [[ $INGEST_EXIT -ne 0 ]]; then
  echo "[wiki-ingest-from-raw] error: wiki ingest failed (exit $INGEST_EXIT)" >&2
  exit 1
fi

log "[wiki-ingest-from-raw] DONE"
log "[wiki-ingest-from-raw]   raw mirror: $SRC → $VAULT_ROOT/raw/projects/$PROJECT/"
log "[wiki-ingest-from-raw]   wiki ingest: $VAULT_ROOT/raw/projects/$PROJECT/ → $VAULT_ROOT/wiki/projects/$PROJECT/"
log "[wiki-ingest-from-raw]   lint report: $VAULT_ROOT/_lint/$PROJECT/ingest_$(date -u +%Y-%m-%d).md"
