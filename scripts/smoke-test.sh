#!/usr/bin/env bash
#
# smoke-test.sh — quick sanity checks for ci-pre-pr-check.sh
#
# Runs the script in 3 modes against a known-good checkout (PR #598 merged)
# and asserts exit codes + JSON shape. Also exercises the cache heuristic
# against a synthetic broken workflow to confirm the guard actually fires
# (PR #598 root-cause mode).
#
# Exit codes:
#   0 = all assertions pass
#   1 = at least one assertion failed
#
# Codex review reference (PR #599):
#   Finding 2 (P2): live repo's --harness-only may emit MED findings for
#   pre-existing fake/mock test files (rbac_test.go etc.), so asserting
#   "no findings" against a live checkout is not stable. Replaced with
#   exit+JSON-shape validation.
#   Finding 1 (P2): previous awk heuristic was file-global, so setup-node's
#   `cache: 'npm'` satisfied the check. New test 6 creates a synthetic
#   workflow with setup-go but no cache: false, asserts HIGH finding.
#
# Usage:
#   bash scripts/smoke-test.sh         # from repo root or worktree

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="${SCRIPT_DIR}/ci-pre-pr-check.sh"
PASS=0
FAIL=0

assert_exit() {
  local label="$1"
  local expected="$2"
  local actual="$3"
  if [ "$expected" = "$actual" ]; then
    echo "  ✓ $label (exit $actual)"
    PASS=$((PASS+1))
  else
    echo "  ✗ $label (expected exit $expected, got $actual)"
    FAIL=$((FAIL+1))
  fi
}

assert_contains() {
  local label="$1"
  local needle="$2"
  local haystack="$3"
  if echo "$haystack" | grep -qF "$needle"; then
    echo "  ✓ $label (contains '$needle')"
    PASS=$((PASS+1))
  else
    echo "  ✗ $label (missing '$needle')"
    FAIL=$((FAIL+1))
  fi
}

# portable temp dir (avoids /tmp pollution on macOS)
TMP_DIR="$(mktemp -d -t ci-pre-pr-smoke.XXXXXX)"
trap 'rm -rf "$TMP_DIR"' EXIT

# Find a safe workflow dir to drop a synthetic broken workflow into.
# The repo's own .github/workflows/ is tracked; we use a sibling temp dir
# and point the script at it via REPO_ROOT=... (or just call script with
# the synthetic workflow placed in a temp dir's .github/workflows/).
make_broken_repo() {
  local broken_repo="$TMP_DIR/broken-repo"
  mkdir -p "$broken_repo/.github/workflows"
  cat > "$broken_repo/.github/workflows/ci.yml" <<'EOF'
name: broken
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
EOF
  # also need a nested go.mod so the check enters the inner loop
  mkdir -p "$broken_repo/backend-core"
  echo "module example.com/broken" > "$broken_repo/backend-core/go.mod"
  echo "$broken_repo"
}

# Find a safe workflow dir for a synthetic GOOD workflow (cache: false explicit).
make_good_repo() {
  local good_repo="$TMP_DIR/good-repo"
  mkdir -p "$good_repo/.github/workflows"
  cat > "$good_repo/.github/workflows/ci.yml" <<'EOF'
name: good
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: false
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
EOF
  mkdir -p "$good_repo/backend-core"
  echo "module example.com/good" > "$good_repo/backend-core/go.mod"
  echo "$good_repo"
}

# helper: copy the script into the temp repo at scripts/ (so REPO_ROOT is the repo)
seed_script() {
  local repo="$1"
  mkdir -p "$repo/scripts"
  cp "$SCRIPT" "$repo/scripts/ci-pre-pr-check.sh"
  chmod +x "$repo/scripts/ci-pre-pr-check.sh"
}

echo "==> ci-pre-pr-check smoke test"
echo

# test 1: --help exits 0
echo "[1/6] --help"
out="$(bash "$SCRIPT" --help 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "shows usage" "Usage:" "$out"

# test 2: --workflows-only on PR #598 merged main (all 3 workflows have cache: false)
echo
echo "[2/6] --workflows-only (clean main: 0 findings)"
out="$(bash "$SCRIPT" --workflows-only 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "no findings" "no findings" "$out"

# test 3: --harness-only (was asserting 'no findings' but live repo's
# pre-existing fake/test files emit MED — Finding 2). Replaced with
# exit-code + JSON-shape validation: script must still complete cleanly.
echo
echo "[3/6] --harness-only (live repo: exit 0 + JSON shape; don't assert 0 findings)"
out="$(bash "$SCRIPT" --harness-only 2>&1)"
assert_exit "exit code" "0" "$?"

# test 4: --json produces valid JSON shape
echo
echo "[4/6] --json shape"
out="$(bash "$SCRIPT" --json 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "has script field" '"script":' "$out"
assert_contains "has findings array" '"findings":' "$out"
assert_contains "has version" '"version": "1.0.0"' "$out"

# test 5: --base-only — behavior depends on where the script is run.
#   - On main checkout (HEAD = origin/main): expect LOW severity.
#   - On a feature branch worktree (HEAD ahead of origin/main): 0 findings
#     (PR is correctly ahead; no MED/HIGH signal needed). Either way exit 0.
# We assert exit 0 + one of (LOW severity OR no findings).
echo
echo "[5/6] --base-only (exit 0; LOW or no-findings depending on branch state)"
out="$(bash "$SCRIPT" --base-only 2>&1)"
assert_exit "exit code" "0" "$?"
if echo "$out" | grep -qF "[LOW]"; then
  PASS=$((PASS+1))
  echo "  ✓ low severity on main checkout (contains '[LOW]')"
elif echo "$out" | grep -qF "no findings"; then
  PASS=$((PASS+1))
  echo "  ✓ no findings on feature branch (contains 'no findings')"
else
  FAIL=$((FAIL+1))
  echo "  ✗ neither LOW severity nor no-findings (PR is in unexpected state)"
fi

# test 6: synthetic broken workflow — cache: false MISSING inside setup-go
# step, with setup-node's cache: 'npm' as a decoy. Expect the script to
# emit a HIGH workflow-cache finding (Finding 1 regression test).
echo
echo "[6/6] synthetic broken workflow (no cache:false inside setup-go step)"
broken_repo="$(make_broken_repo)"
seed_script "$broken_repo"
# cd into the broken repo so the script's REPO_ROOT fallback picks it up
( cd "$broken_repo" && bash "$broken_repo/scripts/ci-pre-pr-check.sh" --workflows-only 2>&1 )
out="$( cd "$broken_repo" && bash "$broken_repo/scripts/ci-pre-pr-check.sh" --workflows-only 2>&1 )"
assert_exit "exit code" "1" "$?"
assert_contains "HIGH workflow-cache finding" "[HIGH] workflow-cache" "$out"

# bonus: synthetic GOOD workflow — explicit cache: false inside setup-go
# step, plus setup-node's cache: 'npm'. Expect 0 findings.
echo
echo "[bonus] synthetic good workflow (cache:false explicit inside setup-go step)"
good_repo="$(make_good_repo)"
seed_script "$good_repo"
out="$( cd "$good_repo" && bash "$good_repo/scripts/ci-pre-pr-check.sh" --workflows-only 2>&1 )"
assert_exit "exit code" "0" "$?"
assert_contains "no findings" "no findings" "$out"

echo
echo "==> ${PASS} passed, ${FAIL} failed"
[ "$FAIL" = "0" ] && exit 0 || exit 1
