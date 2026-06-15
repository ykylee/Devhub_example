#!/usr/bin/env bash
#
# smoke-test.sh — quick sanity checks for ci-pre-pr-check.sh
#
# Runs the script in 3 modes against a known-good checkout (PR #598 merged)
# and asserts exit codes + finding counts. Exits 0 if all pass.
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

echo "==> ci-pre-pr-check smoke test"
echo

# test 1: --help exits 0
echo "[1/5] --help"
out="$(bash "$SCRIPT" --help 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "shows usage" "Usage:" "$out"

# test 2: --workflows-only on PR #598 merged main (all 3 workflows have cache: false)
echo
echo "[2/5] --workflows-only (clean main: 0 findings)"
out="$(bash "$SCRIPT" --workflows-only 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "no findings" "no findings" "$out"

# test 3: --harness-only (PR #598 already has compile-time guards + 1 def of IntegrationProviderType)
echo
echo "[3/5] --harness-only (clean main: 0 findings)"
out="$(bash "$SCRIPT" --harness-only 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "no findings" "no findings" "$out"

# test 4: --json produces valid JSON shape
echo
echo "[4/5] --json shape"
out="$(bash "$SCRIPT" --json 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "has script field" '"script":' "$out"
assert_contains "has findings array" '"findings":' "$out"
assert_contains "has version" '"version": "1.0.0"' "$out"

# test 5: --base-only on HEAD = origin/main (LOW severity, exit 0)
echo
echo "[5/5] --base-only (HEAD = origin/main, expect LOW)"
out="$(bash "$SCRIPT" --base-only 2>&1)"
assert_exit "exit code" "0" "$?"
assert_contains "low severity" "[LOW]" "$out"

echo
echo "==> ${PASS} passed, ${FAIL} failed"
[ "$FAIL" = "0" ] && exit 0 || exit 1
