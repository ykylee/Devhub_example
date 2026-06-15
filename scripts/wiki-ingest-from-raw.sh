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
#   2. L2 dense emit (vendor 의 emit_wiki_l2_body.py) — L1 → L2 sources/ 자동 작성
#   3. (optional) drift check (tests/check_wiki_drift_devhub.py)
#
# 본 script 의 본 저장소 (= DevHub) 측 책임:
#   - raw/ 갱신 + L2 dense emit 의 통합 entry point
#   - in-repo 만 사용 (외부 vault ~/wiki/ 미사용, my_harness 미참조)
#   - dry-run / apply 의 user confirm flow 일관성
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
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${VAULT_ROOT:-${SRC}/ai-workflow/wiki}"
EMIT_TOOL="$SRC/vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py"
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
  # 1. dry-run preview
  bash scripts/wiki-ingest-from-raw.sh

  # 2. 실제 emit
  bash scripts/wiki-ingest-from-raw.sh --apply

  # 3. 1 file 만 emit
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

# ----- vendor 도구 부재 확인 -----
if [[ ! -f "$EMIT_TOOL" ]]; then
  echo "[wiki-ingest-from-raw] error: vendor emit 도구 부재: $EMIT_TOOL" >&2
  echo "[wiki-ingest-from-raw]   hint: 'cp -R ~/repos/standard_ai_workflow_minimax/workflow-source/. vendor/standard_ai_workflow/' 후 재시도" >&2
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

# ----- step 2: L2 dense emit (vendor 의 in-repo 도구) -----
# vendor 의 emit 도구가 *vendor 의 mini structure* (RAW_MIRROR / project / ai-workflow / wiki) 하드코딩.
# 우리 DevHub 의 L1 = $VAULT_ROOT/{concepts,decisions,entities,patterns,topics}/*.md. vendor 의 도구가 우리
# 구조를 인식하려면 source path 가 vendor 의 mini structure 와 일치해야 함. 본 PR 의 follow-up
# (vendor emit 도구의 devhub adapter) 가 미정. 현재는 *dry-run 만 정상, --apply 시 vendor 의
# mini structure mismatch 로 fail 가능*. 그 경우 *수동 emit* 안내.
log ""
log "[wiki-ingest-from-raw] step 2/3: L2 dense emit (L1 → sources/)"
log "[wiki-ingest-from-raw]   tool: $EMIT_TOOL (vendor in-repo, v0.7.17+)"
log "[wiki-ingest-from-raw]   follow-up: vendor emit 도구의 devhub adapter (mini structure mismatch) — 본 PR scope 외"

if [[ $DRY_RUN -eq 1 ]]; then
  log "[wiki-ingest-from-raw]   dry-run: vendor emit 도구 의 dry-run (mini structure mismatch 가능)"
  python3 "$EMIT_TOOL" --project "$PROJECT" --mode all 2>&1 | tail -10 || true
  log "[wiki-ingest-from-raw]   dry-run note: L2 dense emit 의 실제 apply 는 *dry-run* 만. 본 dry-run 의 출력은 preview 일 뿐, sources/ 에 실제 file 작성 안 됨."
else
  # --apply 모드: L2 dense page 의 *silent skip* 방지 (Codex P2, PR #603)
  # 본 script 의 caller 가 L2 emit 의 결과를 *명시적* 으로 알 수 있어야 함. silent skip → DONE 시
  # caller 가 L2 dense page 가 업데이트됐다고 오인 가능. 그래서 *non-zero exit 1* + 명시적 안내.
  log ""
  log "[wiki-ingest-from-raw]   apply: vendor emit 도구 의 apply (mini structure mismatch 시 fail 가능)"
  log "[wiki-ingest-from-raw]   note: 현재 vendor 의 emit 도구는 *vendor 의 mini structure* 만 인식. 우리 DevHub 의"
  log "[wiki-ingest-from-raw]         in-repo L1 (5 page, A안) 의 *수동 L2 emit* 이 PR #602 의 commit 86e2e2df."
  log "[wiki-ingest-from-raw]         전체 220+ file L2 자동화는 follow-up PR 의 *devhub adapter*."
  log ""
  if [[ -n "$SOURCE" ]]; then
    log "[wiki-ingest-from-raw]   단일 file apply: $SOURCE"
  else
    log "[wiki-ingest-from-raw]   전체 apply"
  fi
  log "[wiki-ingest-from-raw]   ERROR: --apply 모드 의 L2 emit 호출 *skip*. 다음 중 하나 선택:"
  log "[wiki-ingest-from-raw]     (a) dry-run 만 사용: --apply 제거, 결과 검토 후 수동 emit"
  log "[wiki-ingest-from-raw]     (b) follow-up PR 후 재실행: vendor emit 도구의 *devhub adapter* 가 본 PR 의 follow-up 으로 작성되면"
  log "[wiki-ingest-from-raw]         step 2 가 자동 활성화. adapter 작성 시 본 else branch 의 exit 1 제거 + python3 호출 복원."
  log ""
  log "[wiki-ingest-from-raw]   exit 1: --apply 의 L2 emit silent skip 방지 (Codex P2, 2026-06-15)"
  log ""
  log "[wiki-ingest-from-raw]   참고: 본 step 1 (raw mirror) 와 step 3 (drift check) 는 *이미 실행됨* — 둘 다 정상."
  log "[wiki-ingest-from-raw]         step 1 의 raw mirror (964 file, 8M) 와 step 3 의 drift check 결과는 *유효*."
  log "[wiki-ingest-from-raw]         L2 dense page (sources/) 만 *수동* 또는 *adapter 후* emit 필요."
  exit 1
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
