#!/usr/bin/env bash
# check_vendor_smoke.sh — vendor (standard_ai_workflow) 동기화 시 4종 smoke 회귀.
#
# 사용:
#   bash scripts/check_vendor_smoke.sh          # real check (default)
#   bash scripts/check_vendor_smoke.sh --quiet  # quiet mode (1 line per check)
#
# 본 script 의 source-of-truth:
#   - ai-workflow/minimax_code_workflow.md §4.4 (vendor smoke 회귀 표)
#   - vendor/standard_ai_workflow/.upstream-url (vendor metadata)
#   - ai-workflow/wiki/RAW_MIRROR_MANIFEST.md (raw mirror 운영 가이드)
#
# 본 script 가 호출하는 4종 smoke:
#   1. tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py (DevHub invariant 5/5)
#      - 본 저장소 의 in-repo path 정합 + 6 dir + memory/log.md + WIKI_SOURCES flat
#   2. vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py (vendor smoke 11/11)
#      - vendor 의 5 file 의 in-repo path + REPO_ROOT auto-detect + legacy 0
#   3. tests/check_emit_wiki_l2_devhub.py (emit 도구 self + vendor smoke 16/16, PR #605 follow-up)
#      - emit_wiki_l2_devhub.py (self) + emit_wiki_l2_devhub_vendor.py (vendor monkey-patch)
#      의 자체 smoke 16 test (help, dry-run, apply, idempotent, L1 discovery, L2 shape,
#      source arg, max-chars, limit, cross-emit logical equivalence, P1 orphan crash regression)
#   4. tests/check_wiki_ingest_devhub.py (wiki-ingest wrapper smoke 3/3, PR #606 follow-up)
#      - wiki-ingest-from-raw.sh 의 wrapper smoke (dry-run self/vendor 분기, --apply vendor
#      의 false-positive byte-identical 검출 — 3 scenario: full body / 일부 placeholder /
#      single-file mode. runtime ~60s)
#
# 합 34/34 PASS 가 본 script 의 정공법.
# vendor release (v0.7.17 → v0.7.18+) 동기화 시 본 script 로 회귀 0 확인 필수.
#
# Exit code:
#   0 — 4종 smoke 모두 PASS (34/34)
#   1 — 1종 이상 FAIL
#   2 — python3 부재 또는 script 부재

set -euo pipefail

# ----- paths -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

QUIET=0
for arg in "$@"; do
  case "$arg" in
    --quiet|-q)
      QUIET=1
      ;;
    -h|--help)
      echo "usage: bash scripts/check_vendor_smoke.sh [--quiet]"
      echo "  --quiet   1 line per check (default: full output)"
      exit 0
      ;;
    *)
      echo "[check-vendor-smoke] invalid option: $arg" >&2
      exit 2
      ;;
  esac
done

# ----- python3 check -----
if ! command -v python3 >/dev/null 2>&1; then
  echo "[check-vendor-smoke] error: python3 not found" >&2
  exit 2
fi

# ----- smoke 1: DevHub invariant (5/5) -----
SMOKE1="$REPO_ROOT/tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py"
if [[ ! -f "$SMOKE1" ]]; then
  echo "[check-vendor-smoke] error: smoke 1 not found: $SMOKE1" >&2
  exit 2
fi

# ----- smoke 2: vendor smoke (11/11) -----
SMOKE2="$REPO_ROOT/vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py"
if [[ ! -f "$SMOKE2" ]]; then
  echo "[check-vendor-smoke] error: smoke 2 not found: $SMOKE2" >&2
  echo "[check-vendor-smoke] hint: vendor/ 가 import 안 됐을 수 있음. 'cp -R ~/repos/standard_ai_workflow_minimax/workflow-source/. vendor/standard_ai_workflow/' 후 재시도" >&2
  exit 2
fi

# ----- smoke 3: emit 도구 (self + vendor) smoke (15/15, PR #605 follow-up) -----
SMOKE3="$REPO_ROOT/tests/check_emit_wiki_l2_devhub.py"
if [[ ! -f "$SMOKE3" ]]; then
  echo "[check-vendor-smoke] error: smoke 3 not found: $SMOKE3" >&2
  exit 2
fi

# ----- smoke 4: wiki-ingest wrapper smoke (3/3, PR #606 P1 follow-up) -----
# wiki-ingest-from-raw.sh 의 wrapper smoke. --apply --emit-tool vendor 의 false-positive
# byte-identical 검출 (PR #606 P1 fix) + dry-run 의 self/vendor 분기 (PR #606 P2 fix) 검증.
# 3 scenario (full body / 일부 placeholder / single-file mode) 모두 exit 1 + false-positive
# message raise. runtime ~60s.
SMOKE4="$REPO_ROOT/tests/check_wiki_ingest_devhub.py"
if [[ ! -f "$SMOKE4" ]]; then
  echo "[check-vendor-smoke] error: smoke 4 not found: $SMOKE4" >&2
  exit 2
fi

# ----- run -----
PASS1=0
FAIL1=0
PASS2=0
FAIL2=0
PASS3=0
FAIL3=0
PASS4=0
FAIL4=0

if [[ $QUIET -eq 1 ]]; then
  echo "[check-vendor-smoke] smoke 1/4: tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py"
  if python3 "$SMOKE1" --quiet 2>/dev/null || python3 "$SMOKE1" >/dev/null 2>&1; then
    PASS1=1
  else
    FAIL1=1
  fi
  echo "[check-vendor-smoke] smoke 2/4: vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py"
  if python3 "$SMOKE2" >/dev/null 2>&1; then
    PASS2=1
  else
    FAIL2=1
  fi
  echo "[check-vendor-smoke] smoke 3/4: tests/check_emit_wiki_l2_devhub.py"
  if python3 "$SMOKE3" >/dev/null 2>&1; then
    PASS3=1
  else
    FAIL3=1
  fi
  echo "[check-vendor-smoke] smoke 4/4: tests/check_wiki_ingest_devhub.py"
  if python3 "$SMOKE4" >/dev/null 2>&1; then
    PASS4=1
  else
    FAIL4=1
  fi
else
  echo "[check-vendor-smoke] === smoke 1/4: DevHub invariant (5/5 expected) ==="
  echo ""
  if python3 "$SMOKE1"; then
    PASS1=1
  else
    FAIL1=1
  fi
  echo ""
  echo "[check-vendor-smoke] === smoke 2/4: vendor in-repo isolation (11/11 expected) ==="
  echo ""
  if python3 "$SMOKE2"; then
    PASS2=1
  else
    FAIL2=1
  fi
  echo ""
  echo "[check-vendor-smoke] === smoke 3/4: emit 도구 (self + vendor) smoke (16/16 expected) ==="
  echo ""
  if python3 "$SMOKE3"; then
    PASS3=1
  else
    FAIL3=1
  fi
  echo ""
  echo "[check-vendor-smoke] === smoke 4/4: wiki-ingest wrapper smoke (3/3 expected) ==="
  echo ""
  if python3 "$SMOKE4"; then
    PASS4=1
  else
    FAIL4=1
  fi
fi

# ----- summary -----
echo ""
echo "[check-vendor-smoke] === summary ==="
echo "  smoke 1 (DevHub invariant):    $([[ $PASS1 -eq 1 ]] && echo 'PASS' || echo 'FAIL') (expected 5/5)"
echo "  smoke 2 (vendor in-repo):      $([[ $PASS2 -eq 1 ]] && echo 'PASS' || echo 'FAIL') (expected 11/11)"
echo "  smoke 3 (emit 도구 self+vendor): $([[ $PASS3 -eq 1 ]] && echo 'PASS' || echo 'FAIL') (expected 16/16)"
echo "  smoke 4 (wiki-ingest wrapper):   $([[ $PASS4 -eq 1 ]] && echo 'PASS' || echo 'FAIL') (expected 3/3)"
echo "  total: $([[ $PASS1 -eq 1 && $PASS2 -eq 1 && $PASS3 -eq 1 && $PASS4 -eq 1 ]] && echo '34/34 PASS' || echo 'FAIL')"

if [[ $PASS1 -eq 1 && $PASS2 -eq 1 && $PASS3 -eq 1 && $PASS4 -eq 1 ]]; then
  exit 0
else
  exit 1
fi
