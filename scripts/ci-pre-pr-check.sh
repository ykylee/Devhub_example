#!/usr/bin/env bash
#
# ci-pre-pr-check.sh — pre-PR hook for DevHub CI failure recovery pattern
#                     (cross-reference: docs/planning/2026-06-15-pr-598-ci-recovery-retrospective.md §6)
#
# Implements retro §6 steps 1-3 + interface drift check (steps 6-7 partial):
#   1. detect common early failure: actions/setup-go implicit cache when go.mod is nested under backend-core/
#   2. check branch base freshness vs origin/main (stale base = analysis distortion)
#   3. detect Go interface drift in test fake stores (missing `var _ Interface = (*Fake)(nil)` guards)
#   4. detect duplicate type definitions in frontend schema (TS2300 'Duplicate identifier' class)
#
# Usage:
#   bash scripts/ci-pre-pr-check.sh                    # default: all checks, report-only mode
#   bash scripts/ci-pre-pr-check.sh --strict           # exit 1 if any HIGH severity finding
#   bash scripts/ci-pre-pr-check.sh --workflows-only   # only step 1
#   bash scripts/ci-pre-pr-check.sh --base-only        # only step 2
#   bash scripts/ci-pre-pr-check.sh --harness-only     # only steps 3-4
#   bash scripts/ci-pre-pr-check.sh --json             # machine-readable output
#
# Exit codes:
#   0 = no findings or only LOW/MED severity
#   1 = HIGH severity finding (--strict required to fail on these in default mode)
#   2 = script/invocation error
#
# Retro reference:
#   docs/planning/2026-06-15-pr-598-ci-recovery-retrospective.md §6 (7-step 대응 절차)
#
# Cross-project: also valid for any project that has:
#   - workflows under .github/workflows/*.yml
#   - nested Go modules (go.mod in subdirectory)
#   - test fake stores implementing production interfaces

set -euo pipefail

SCRIPT_NAME="ci-pre-pr-check"
SCRIPT_VERSION="1.0.0"
REPO_ROOT="$(git rev-parse --show-toplevel)"

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

MODE="all"        # all | workflows | base | harness
STRICT=0
JSON_OUTPUT=0

while [ $# -gt 0 ]; do
  case "$1" in
    --strict)            STRICT=1; shift ;;
    --workflows-only)    MODE="workflows"; shift ;;
    --base-only)         MODE="base"; shift ;;
    --harness-only)      MODE="harness"; shift ;;
    --json)              JSON_OUTPUT=1; shift ;;
    --help|-h)
      sed -n '2,32p' "$0"
      exit 0
      ;;
    *)  echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

# ---------------------------------------------------------------------------
# Finding buffer + JSON helpers
# ---------------------------------------------------------------------------

# findings stored as "SEVERITY|CHECK|MESSAGE" lines in $FINDINGS
FINDINGS=""

add_finding() {
  local severity="$1"
  local check="$2"
  local message="$3"
  FINDINGS="${FINDINGS}${severity}|${check}|${message}
"
}

severity_rank() {
  case "$1" in
    HIGH)   echo 3 ;;
    MED)    echo 2 ;;
    LOW)    echo 1 ;;
    *)      echo 0 ;;
  esac
}

# ---------------------------------------------------------------------------
# Check 1: workflow cache path
#
#   actions/setup-go@v5 implicit cache looks for go.sum at repo root.
#   If go.mod lives in a subdirectory, this fails with
#   "Dependencies file is not found ... go.sum" — see PR #598.
#
# Fix: add `cache: false` to actions/setup-go@v5 when go.mod is nested.
# ---------------------------------------------------------------------------

check_workflow_cache_path() {
  local workflows_dir="${REPO_ROOT}/.github/workflows"
  local go_mod_locations
  go_mod_locations="$(find "$REPO_ROOT" -name go.mod -not -path "*/node_modules/*" -not -path "*/.worktrees/*" 2>/dev/null || true)"

  if [ -z "$go_mod_locations" ]; then
    return 0
  fi

  # detect nested go.mod (not at repo root)
  local nested=0
  local go_mod
  for go_mod in $go_mod_locations; do
    local rel
    rel="$(realpath --relative-to="$REPO_ROOT" "$go_mod" 2>/dev/null || echo "$go_mod")"
    if [ "$rel" != "go.mod" ]; then
      nested=1
      break
    fi
  done

  if [ "$nested" = "0" ]; then
    return 0
  fi

  # check all workflow files for setup-go without cache: false
  local wf
  for wf in "$workflows_dir"/*.yml "$workflows_dir"/*.yaml; do
    [ -f "$wf" ] || continue
    if ! grep -q "actions/setup-go" "$wf"; then
      continue
    fi
    # detect setup-go block — heuristic: look for "with:" under setup-go
    if awk '
      /uses:[[:space:]]+actions\/setup-go@/ { in_block=1; next }
      in_block && /^[[:space:]]+with:/ { in_with=1; next }
      in_with && /^[[:space:]]+cache:/ { cache_set=1 }
      in_block && /^[[:space:]]*[^[:space:]].*:[[:space:]]*$/ && !/^[[:space:]]+(with|with:|with[^a-z])/ { in_block=0; in_with=0 }
      END { exit (cache_set ? 0 : 1) }
    ' "$wf"; then
      continue
    fi
    add_finding "HIGH" "workflow-cache" "$(realpath --relative-to="$REPO_ROOT" "$wf" 2>/dev/null || echo "$wf") uses actions/setup-go without explicit cache: false — nested go.mod will trigger 'Dependencies file is not found' failure (PR #598 root cause). Add 'cache: false' to the with: block or use explicit actions/cache step."
  done
}

# ---------------------------------------------------------------------------
# Check 2: branch base freshness
#
#   Stale local main = analysis distortion (PR #598 §5.1).
#   Compare HEAD's merge-base with origin/main; report how far behind.
# ---------------------------------------------------------------------------

check_base_freshness() {
  if ! git rev-parse --verify origin/main >/dev/null 2>&1; then
    add_finding "MED" "base-freshness" "origin/main not fetched locally — run 'git fetch origin main' before opening PR"
    return 0
  fi

  local head_sha origin_main_sha merge_base
  head_sha="$(git rev-parse HEAD)"
  origin_main_sha="$(git rev-parse origin/main)"
  merge_base="$(git merge-base HEAD origin/main)"

  if [ "$head_sha" = "$origin_main_sha" ]; then
    add_finding "LOW" "base-freshness" "HEAD is at origin/main (${head_sha:0:7}) — branch is empty or already merged"
    return 0
  fi

  if [ "$head_sha" = "$merge_base" ]; then
    add_finding "HIGH" "base-freshness" "HEAD is the merge-base of origin/main — no commits ahead of main. Open a feature branch first."
    return 0
  fi

  # count commits origin/main is ahead of HEAD (the danger: stale base)
  local ahead
  ahead="$(git rev-list --count HEAD..origin/main)"

  if [ "$ahead" -ge 1 ]; then
    local severity="MED"
    if [ "$ahead" -ge 5 ]; then
      severity="HIGH"
    fi
    add_finding "$severity" "base-freshness" "origin/main is ${ahead} commit(s) ahead of HEAD (merge-base: ${merge_base:0:7}, origin/main: ${origin_main_sha:0:7}). Stale base may hide latest regression — consider 'git rebase origin/main' before opening PR (PR #598 §5.1)."
  fi
}

# ---------------------------------------------------------------------------
# Check 3: Go test harness interface drift
#
#   PR #598 §5.3: memoryPlatformStore missed new IntegrationStore methods,
#   causing 503 'integration store is not configured' at test runtime.
#   Pattern fix: `var _ Interface = (*Fake)(nil)` compile-time guard.
#
# We look for test files (*_test.go) that define a fake/mock struct
# implementing some interface, and check whether the compile-time guard
# pattern is present. This is a heuristic — it won't catch every case.
# ---------------------------------------------------------------------------

check_go_harness_guard() {
  local test_files
  test_files="$(find "$REPO_ROOT" -name "*_test.go" -not -path "*/.worktrees/*" 2>/dev/null || true)"
  if [ -z "$test_files" ]; then
    return 0
  fi

  local tf
  for tf in $test_files; do
    # heuristic: file contains type fake*/memory*/mock* struct
    if ! grep -Eq 'type[[:space:]]+(fake|memory|mock|stub)[A-Za-z_]*[[:space:]]+(struct|=.*struct)' "$tf"; then
      continue
    fi

    # heuristic: file has methods (func (x *<Type>) Method()) — implies implements some interface
    if ! grep -Eq '^func[[:space:]]+\(.*\*[A-Za-z]+\)[[:space:]]+[A-Z]' "$tf"; then
      continue
    fi

    # heuristic: file references some production interface (uppercase type from same package or import)
    # the guard pattern is `var _ Interface = (*Fake)(nil)`
    if grep -Eq 'var _ [A-Z][A-Za-z]+[[:space:]]*=[[:space:]]*\(\*[A-Za-z]+\)\(nil\)' "$tf"; then
      continue
    fi

    add_finding "MED" "go-harness-guard" "$(realpath --relative-to="$REPO_ROOT" "$tf" 2>/dev/null || echo "$tf") defines a fake/mock struct with methods but no 'var _ Interface = (*Fake)(nil)' compile-time guard — risk of interface drift at runtime (PR #598 §5.3). Add the guard line."
  done
}

# ---------------------------------------------------------------------------
# Check 4: frontend duplicate type definitions
#
#   PR #598 §3.3: integration.types.ts had IntegrationProviderType defined
#   twice → TS2300 'Duplicate identifier' build error.
#   Detect same identifier declared in `export type X = ...` form more than
#   once within a single TS/TSX file.
# ---------------------------------------------------------------------------

check_frontend_duplicate_types() {
  local ts_files
  ts_files="$(find "$REPO_ROOT" -name "*.ts" -o -name "*.tsx" 2>/dev/null | grep -v node_modules | grep -v ".worktrees" || true)"
  if [ -z "$ts_files" ]; then
    return 0
  fi

  local tf
  for tf in $ts_files; do
    # extract identifiers in `export type X` form, count occurrences
    local dups
    dups="$(grep -Eoh 'export[[:space:]]+type[[:space:]]+[A-Z][A-Za-z0-9_]+' "$tf" | awk '{print $3}' | sort | uniq -d || true)"
    if [ -n "$dups" ]; then
      add_finding "HIGH" "frontend-dup-type" "$(realpath --relative-to="$REPO_ROOT" "$tf" 2>/dev/null || echo "$tf") declares the same type identifier more than once: $(echo "$dups" | tr '\n' ',' ). PR #598 §3.3 hit this with IntegrationProviderType — remove duplicate definition."
    fi
  done
}

# ---------------------------------------------------------------------------
# Run selected checks
# ---------------------------------------------------------------------------

case "$MODE" in
  all)
    check_workflow_cache_path
    check_base_freshness
    check_go_harness_guard
    check_frontend_duplicate_types
    ;;
  workflows)
    check_workflow_cache_path
    ;;
  base)
    check_base_freshness
    ;;
  harness)
    check_go_harness_guard
    check_frontend_duplicate_types
    ;;
  *)  echo "internal error: unknown mode $MODE" >&2; exit 2 ;;
esac

# ---------------------------------------------------------------------------
# Render output
# ---------------------------------------------------------------------------

if [ "$JSON_OUTPUT" = "1" ]; then
  # JSON array
  printf '{\n'
  printf '  "script": "%s",\n' "$SCRIPT_NAME"
  printf '  "version": "%s",\n' "$SCRIPT_VERSION"
  printf '  "mode": "%s",\n' "$MODE"
  printf '  "repo_root": "%s",\n' "$REPO_ROOT"
  printf '  "head_sha": "%s",\n' "$(git rev-parse HEAD)"
  printf '  "origin_main_sha": "%s",\n' "$(git rev-parse origin/main 2>/dev/null || echo none)"
  printf '  "findings": [\n'
  json_first=1
  while IFS='|' read -r sev check msg; do
    [ -z "$sev" ] && continue
    if [ "$json_first" = "1" ]; then json_first=0; else printf ',\n'; fi
    printf '    {"severity": "%s", "check": "%s", "message": "%s"}' \
      "$sev" "$check" "$(echo "$msg" | sed 's/"/\\"/g')"
  done <<< "$FINDINGS"
  printf '\n  ]\n}\n'
else
  echo "==> $SCRIPT_NAME v$SCRIPT_VERSION (mode=$MODE, strict=$STRICT)"
  echo "    repo: $REPO_ROOT"
  echo "    head: $(git rev-parse --short HEAD)"
  echo "    origin/main: $(git rev-parse --short origin/main 2>/dev/null || echo none)"
  echo

  if [ -z "$FINDINGS" ]; then
    echo "  ✓ no findings"
    echo
    echo "  retro §6 status: ready to open PR"
    exit 0
  fi

  echo "  Findings:"
  echo "  ----------"
  while IFS='|' read -r sev check msg; do
    [ -z "$sev" ] && continue
    printf '  [%s] %s\n      %s\n\n' "$sev" "$check" "$msg"
  done <<< "$FINDINGS"
fi

# ---------------------------------------------------------------------------
# Exit code
# ---------------------------------------------------------------------------

# find max severity in findings
max_rank=0
while IFS='|' read -r sev _check _msg; do
  [ -z "$sev" ] && continue
  r=$(severity_rank "$sev")
  if [ "$r" -gt "$max_rank" ]; then max_rank="$r"; fi
done <<< "$FINDINGS"

# default: only fail on HIGH if --strict
# without --strict: only HIGH still triggers exit 1 (safer default)
# rationale: this is pre-PR guard, HIGH = known PR-blocker → must fix
if [ "$max_rank" -ge 3 ]; then
  exit 1
fi
exit 0
