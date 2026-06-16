#!/usr/bin/env bash
# wiki-ingest-from-raw.sh — DevHub 의 raw mirror + L2 dense emit 통합 wrapper.
#
# 사용:
#   bash scripts/wiki-ingest-from-raw.sh --dry-run                  # raw mirror + L2 emit preview
#   bash scripts/wiki-ingest-from-raw.sh --apply                    # raw mirror + L2 emit apply
#   bash scripts/wiki-ingest-from-raw.sh --source <rel> --apply     # 1 file 만 emit
#
# 본 script 의 source-of-truth (v0.7.17+ in-repo redirect, 2026-06-15):
#   - DevHub mirror tool:  scripts/wiki-sync-devhub.sh (in-repo)
#   - L2 dense emit:       vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py (in-repo)
#   - drift check:         tests/check_wiki_drift_devhub.py (in-repo DevHub adapter)
#   - vault 운영 규약:     ai-workflow/wiki/ (L0/L1/L2 in-repo)
#
# **Deprecated (2026-06-15)**: 본 script 의 *my_harness 측 run_wiki_ingest.py / run_wiki_lint.py*
# 호출 제거됨. my_harness 의 wiki-* skill 의 *in-repo redirect* (v0.7.17 결정) 가 본 PR 의
# follow-up. 본 script 는 *raw mirror + L2 emit* 만 호출. my_harness 의 *본 저장소 미참조*.
#
# 결정적 단순: 3 단계 wrapper.
#   1. raw mirror (wiki-sync-devhub.sh) — raw/ 갱신 (in-repo)
#   2. L2 dense emit (DevHub adapter, PR #605 follow-up #3)
#      - self:  scripts/emit_wiki_l2_devhub.py          (vendor 미사용, 1차 출처)
#      - vendor: scripts/emit_wiki_l2_devhub_vendor.py  (vendor monkey-patch wrapper)
#      - 동등성 검증: 양쪽 결과 의 file count + byte-identical 동일 시 PASS
#   3. (optional) drift check (tests/check_wiki_drift_devhub.py)
#
# 본 script 의 본 저장소 (= DevHub) 측 책임:
#   - raw/ 갱신 + L2 dense emit 의 통합 entry point
#   - in-repo 만 사용 (외부 vault ~/wiki/ 미사용, my_harness 미참조)
#   - dry-run / apply 의 user confirm flow 일관성
#   - emit 도구 self/vendor 선택 (default = self, 동등성 검증)
#
# Exit code:
#   0 — success (raw mirror + L2 emit dry-run 또는 apply 모두 성공)
#   1 — raw mirror 실패, 또는 L2 emit 실패, 또는 vendor 도구 부재
#   2 — invalid option 또는 required option (--project) 부재

set -euo pipefail

# ----- options -----
DRY_RUN=1
PROJECT="devhub"
SOURCE=""
LIMIT=""
SKIP_LINT=0
QUIET=0
EMIT_TOOL_CHOICE="self"  # PR #603 follow-up #3: self | vendor (default = self, 동등성 검증)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${VAULT_ROOT:-${SRC}/ai-workflow/wiki}"
EMIT_TOOL_SELF="$SCRIPT_DIR/emit_wiki_l2_devhub.py"
EMIT_TOOL_VENDOR="$SCRIPT_DIR/emit_wiki_l2_devhub_vendor.py"
EMIT_TOOL=""  # resolved after --emit-tool parse
DRIFT_CHECK="$SRC/tests/check_wiki_drift_devhub.py"

usage() {
  cat <<'EOF'
Usage: bash scripts/wiki-ingest-from-raw.sh [options]

Options:
  --project <devhub>             Default = devhub. (in-repo 만, my_harness 미참조)
  --source <rel_path>            1 file 만 emit (L1 page 의 상대 경로).
  --limit N                      --all 시 최대 N건.
  --apply                        실제 emit (default = dry-run).
  --skip-lint                    post-emit drift check skip.
  --quiet                        stderr 메시지 최소화.
  -h, --help                     도움말.

Examples:
  # 1. dry-run preview (default emit tool = self)
  bash scripts/wiki-ingest-from-raw.sh

  # 2. 실제 emit (self 도구)
  bash scripts/wiki-ingest-from-raw.sh --apply

  # 3. 실제 emit (vendor monkey-patch wrapper, 동등성 검증)
  bash scripts/wiki-ingest-from-raw.sh --apply --emit-tool vendor

  # 4. 1 file 만 emit
  bash scripts/wiki-ingest-from-raw.sh --source concepts/devhub-overview.md --apply
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
    --emit-tool)
      EMIT_TOOL_CHOICE="$2"
      shift 2
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
  echo "[wiki-ingest-from-raw] error: --project required" >&2
  exit 2
fi
# 2026-06-15 결정: my-harness project 옵션은 제거. in-repo only.
if [[ "$PROJECT" != "devhub" ]]; then
  echo "[wiki-ingest-from-raw] error: invalid --project: $PROJECT (must be devhub, in-repo only)" >&2
  echo "[wiki-ingest-from-raw] hint: 2026-06-15+ 결정 — my_harness wiki-* skill 미참조. in-repo 만 운영." >&2
  exit 2
fi
if [[ -n "$SOURCE" && -n "$LIMIT" ]]; then
  echo "[wiki-ingest-from-raw] error: --source and --limit are mutually exclusive" >&2
  exit 2
fi

# ----- emit 도구 resolution (PR #603 follow-up #3) -----
case "$EMIT_TOOL_CHOICE" in
  self)
    EMIT_TOOL="$EMIT_TOOL_SELF"
    ;;
  vendor)
    EMIT_TOOL="$EMIT_TOOL_VENDOR"
    ;;
  *)
    echo "[wiki-ingest-from-raw] error: invalid --emit-tool: $EMIT_TOOL_CHOICE (must be self|vendor)" >&2
    exit 2
    ;;
esac

# ----- emit 도구 부재 확인 -----
if [[ ! -f "$EMIT_TOOL" ]]; then
  echo "[wiki-ingest-from-raw] error: emit 도구 부재: $EMIT_TOOL" >&2
  case "$EMIT_TOOL_CHOICE" in
    self)
      echo "[wiki-ingest-from-raw]   hint: PR #605 의 scripts/emit_wiki_l2_devhub.py 가 본 worktree 에 없음" >&2
      echo "[wiki-ingest-from-raw]   hint: 'git fetch origin feat/vendor-emit-devhub-adapter && git cherry-pick <commit>' 로 PR #605 흡수" >&2
      ;;
    vendor)
      echo "[wiki-ingest-from-raw]   hint: PR #605 의 scripts/emit_wiki_l2_devhub_vendor.py 가 본 worktree 에 없음" >&2
      ;;
  esac
  exit 1
fi

# ----- log helper -----
log() { [[ $QUIET -eq 1 ]] || echo "$@"; }

# ----- step 1: raw mirror (DevHub) -----
log "[wiki-ingest-from-raw] step 1/3: raw mirror ($SRC → $VAULT_ROOT/raw/projects/$PROJECT/)"
log "[wiki-ingest-from-raw]   source root: $SRC"
log "[wiki-ingest-from-raw]   target: $VAULT_ROOT (in-repo, v0.7.17+)"
log "[wiki-ingest-from-raw]   mode: $([[ $DRY_RUN -eq 1 ]] && echo 'dry-run' || echo 'apply')"

if [[ $DRY_RUN -eq 1 ]]; then
  bash "$SCRIPT_DIR/wiki-sync-devhub.sh" --dry-run
else
  bash "$SCRIPT_DIR/wiki-sync-devhub.sh"
fi

# ----- step 2: L2 dense emit (DevHub adapter, PR #605 follow-up #3) -----
# 본 step 의 emit 도구:
#   - self:   scripts/emit_wiki_l2_devhub.py          (vendor 미사용, 1차 출처)
#   - vendor: scripts/emit_wiki_l2_devhub_vendor.py  (vendor monkey-patch wrapper)
#
# 동등성 검증 (--emit-tool=vendor 인 경우, dry-run 무관): self 결과 와 vendor 결과 의
# file count + sha256 일치 시 PASS. 불일치 시 non-zero exit 1 (caller 가 명시적 실패 인지).
#
# Decision: PR #603 의 Codex P2 (silent skip 방지) 와 PR #605 의 emit 도구 통합.
# 본 step 의 silent skip 제거 — 정상 호출 + 결과 검증.
log ""
log "[wiki-ingest-from-raw] step 2/3: L2 dense emit (L1 → sources/)"
log "[wiki-ingest-from-raw]   tool: $EMIT_TOOL_CHOICE ($EMIT_TOOL)"
if [[ $DRY_RUN -eq 1 ]]; then
  log "[wiki-ingest-from-raw]   mode: dry-run (preview, write 안 함)"
  # dry-run: emit 도구 의 *default* dry-run 호출. self 는 --project/--mode 미지원 →
  # 그 option 을 넘기면 argparse exit 2 + silent skip. (PR #606 Codex P2 해소)
  if [[ "$EMIT_TOOL_CHOICE" == "vendor" ]]; then
    # vendor 는 --project / --mode 지원. default mode=all 이라 self 와 같은 후보 set.
    python3 "$EMIT_TOOL" --project "$PROJECT" --mode all 2>&1 | tail -10 || {
      rc=$?
      log "[wiki-ingest-from-raw]   dry-run warning: vendor emit 도구 exit code $rc (preview 만)" >&2
    }
  else
    # self: vendor-only flag 없이 호출 (default dry-run, args.apply=False).
    python3 "$EMIT_TOOL" 2>&1 | tail -10 || {
      rc=$?
      log "[wiki-ingest-from-raw]   dry-run warning: self emit 도구 exit code $rc (preview 만)" >&2
    }
  fi
  log "[wiki-ingest-from-raw]   dry-run done. L2 emit 의 실제 apply 는 --apply 시."
else
  # --apply: 정상 emit (silent skip 제거, Codex P2 해소)
  log "[wiki-ingest-from-raw]   mode: apply (실제 L2 dense page 작성)"

  # 동등성 검증: --emit-tool=vendor 인 경우, self 와 vendor 의 결과가 동일한지 검증
  # 검증 방법: emit 도구 의 *실제 sources/ 출력* 의 *file content* byte-identical 비교
  # (emit 도구가 같은 sources/<stem>.md 에 작성하므로, 한 도구 의 *출력 file* = 비교 대상)
  if [[ "$EMIT_TOOL_CHOICE" == "vendor" ]]; then
    log "[wiki-ingest-from-raw]   동등성 검증: self 와 vendor 의 sources/ 출력 비교 (byte-identical)"

    # self 실행 (--source 1 file 또는 all)
    if [[ -n "$SOURCE" ]]; then
      python3 "$EMIT_TOOL_SELF" --apply --source "$SOURCE" 2>&1 | tail -3
    else
      python3 "$EMIT_TOOL_SELF" --apply 2>&1 | tail -3
    fi
    SELF_RC=$?
    if [[ $SELF_RC -ne 0 ]]; then
      log "[wiki-ingest-from-raw] error: self emit 도구 실패 (rc=$SELF_RC)" >&2
      exit 1
    fi

    # self 가 작성한 file 들 의 sha256 capture
    SELF_TARGETS=$(find "$VAULT_ROOT/sources" -type f -name '*.md' 2>/dev/null | LC_ALL=C sort)
    SELF_HASHES=$(echo "$SELF_TARGETS" | xargs -I {} sh -c "sha256sum '{}'" | sha256sum | awk '{print $1}')

    # vendor 실행 전 sources/ 의 *file list + hash* snapshot — vendor activity 검증용
    # (PR #606 Codex P2 해소: vendor 가 self 가 채운 placeholder 위에서 0 file emit →
    # byte-identical PASS 가 self-only-pass 가 되는 false-positive 방지)
    SOURCES_BEFORE_HASHES=""
    if [[ -n "$SELF_TARGETS" ]]; then
      SOURCES_BEFORE_HASHES=$(echo "$SELF_TARGETS" | xargs sha256sum 2>/dev/null | LC_ALL=C sort)
    fi

    # vendor 실행
    if [[ -n "$SOURCE" ]]; then
      python3 "$EMIT_TOOL_VENDOR" --apply --source "$SOURCE" 2>&1 | tail -3
    else
      python3 "$EMIT_TOOL_VENDOR" --apply 2>&1 | tail -3
    fi
    VENDOR_RC=$?
    if [[ $VENDOR_RC -ne 0 ]]; then
      log "[wiki-ingest-from-raw] error: vendor emit 도구 실패 (rc=$VENDOR_RC)" >&2
      exit 1
    fi

    # vendor activity 검증: vendor 실행 후 변형된 file 이 0개면 false-positive.
    # (self 가 모든 placeholder 를 채워서 vendor 의 _patched_needs_body 가 모두 False →
    # vendor 가 0 file emit → byte-identical 이 self-only 결과.)
    SOURCES_AFTER_HASHES=$(echo "$SELF_TARGETS" | xargs sha256sum 2>/dev/null | LC_ALL=C sort)
    VENDOR_CHANGED=$(comm -23 <(echo "$SOURCES_AFTER_HASHES") <(echo "$SOURCES_BEFORE_HASHES") | wc -l | tr -d ' ')
    if [[ "$VENDOR_CHANGED" -eq 0 ]]; then
      log "[wiki-ingest-from-raw] error: vendor 가 0 file 변형 — _patched_needs_body 가 모두 False (false-positive byte-identical 방지)" >&2
      log "[wiki-ingest-from-raw] hint: self 가 이미 placeholder 를 채운 상태에서 vendor 가 nothing to do. --source 로 single file L1 만 emit 해보세요." >&2
      exit 1
    fi

    # vendor 가 작성한 file 들 의 sha256 capture + 비교
    VENDOR_HASHES=$(echo "$SELF_TARGETS" | xargs -I {} sh -c "sha256sum '{}'" | sha256sum | awk '{print $1}')
    if [[ "$SELF_HASHES" != "$VENDOR_HASHES" ]]; then
      log "[wiki-ingest-from-raw] error: 동등성 검증 FAIL — sha256 불일치" >&2
      log "  self:   $SELF_HASHES" >&2
      log "  vendor: $VENDOR_HASHES" >&2
      exit 1
    fi

    log "[wiki-ingest-from-raw]   동등성 검증 PASS ($(echo "$SELF_TARGETS" | wc -l | tr -d ' ') file, sha256 match, vendor_changed=$VENDOR_CHANGED)"
  else
    # self 만 (default)
    if [[ -n "$SOURCE" ]]; then
      log "[wiki-ingest-from-raw]   단일 file apply: $SOURCE"
      python3 "$EMIT_TOOL" --apply --source "$SOURCE" 2>&1 | tail -5
    else
      log "[wiki-ingest-from-raw]   전체 apply"
      python3 "$EMIT_TOOL" --apply 2>&1 | tail -5
    fi
    EMIT_RC=$?
    if [[ $EMIT_RC -ne 0 ]]; then
      log "[wiki-ingest-from-raw] error: emit 도구 exit code $EMIT_RC" >&2
      exit 1
    fi
  fi
  log "[wiki-ingest-from-raw]   L2 emit apply 완료 (sources/ 갱신)"
fi

# ----- step 3: drift check (DevHub 자체 adapter) -----
if [[ $SKIP_LINT -eq 0 ]]; then
  log ""
  log "[wiki-ingest-from-raw] step 3/3: drift check (in-repo DevHub adapter)"
  if [[ ! -f "$DRIFT_CHECK" ]]; then
    log "[wiki-ingest-from-raw]   warn: drift check 부재: $DRIFT_CHECK (skip)"
  else
    if [[ $QUIET -eq 1 ]]; then
      python3 "$DRIFT_CHECK" --quiet 2>/dev/null || python3 "$DRIFT_CHECK" 2>&1 | tail -5
    else
      python3 "$DRIFT_CHECK"
    fi
  fi
else
  log "[wiki-ingest-from-raw] step 3/3: drift check (skipped via --skip-lint)"
fi

log ""
log "[wiki-ingest-from-raw] DONE"
log "[wiki-ingest-from-raw]   raw mirror: $SRC → $VAULT_ROOT/raw/projects/$PROJECT/"
log "[wiki-ingest-from-raw]   L2 dense: $VAULT_ROOT/{concepts,decisions,entities,patterns,topics}/ → $VAULT_ROOT/sources/"
log "[wiki-ingest-from-raw]   mode: $([[ $DRY_RUN -eq 1 ]] && echo 'dry-run' || echo 'apply')"
log "[wiki-ingest-from-raw]   my_harness 호출: 0 (2026-06-15+ 결정)"
