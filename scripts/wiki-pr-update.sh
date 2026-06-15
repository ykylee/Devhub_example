#!/usr/bin/env bash
# wiki-pr-update.sh — DevHub 의 in-repo PR-vault update wrapper.
# PR (`gh pr view <num>`) 의 metadata + touched files → `ai-workflow/wiki/prs/<num>.md` 신규 + log.md append.
# PR-auto-update skill 의 본 저장소 측 thin wrapper (D-72 pattern).
#
# 사용:
#   bash scripts/wiki-pr-update.sh --pr 552                         # dry-run
#   bash scripts/wiki-pr-update.sh --pr 552 --apply                # actual update
#   bash scripts/wiki-pr-update.sh --pr 552 --reingest --apply     # PR touched file 마다 re-ingest
#
# 본 script 의 source-of-truth (v0.7.17+ in-repo redirect, 2026-06-15):
#   - in-repo vault:        ai-workflow/wiki/ (prs/<num>.md + index.md)
#   - log target:           ai-workflow/memory/log.md
#   - 자체 update 도구:      python3 inline (PR metadata → prs/<num>.md + log.md + index.md)
#   - trigger:              `gh pr merge` 후 사용자가 본 wrapper 수동 실행
#
# **Deprecated (2026-06-15)**: my_harness 측 run_wiki_pr_update.py 호출 제거. in-repo 만 운영.
# vendor 의 core/ 에 wiki-pr-update 동등 도구 미존재 → 자체 inline python 으로 대체.
#
# 결정적 단순: 3 단계.
#   1. PR metadata + touched files 추출 (`gh pr view <num> --json ...` + `gh pr diff --name-only`)
#   2. (--reingest) touched file 마다 `wiki-ingest-from-raw --source <file> --apply` re-run (idempotent)
#   3. vault side effects (자체 inline python) — prs/<num>.md 신규 + log.md 1 line + index.md 갱신
#
# Idempotency key: `pr-<num>-<head.sha>` — 이미 갱신된 PR 은 skip (vault state 확인).
#
# Exit code:
#   0 — success (dry-run 또는 apply 모두 성공, 0 touched files 도 success, 이미 갱신된 PR 도 success)
#   1 — gh pr view 실패, 또는 vault side effect 실패
#   2 — invalid option 또는 required option (--pr) 부재

set -euo pipefail

# ----- options -----
DRY_RUN=1
PR_NUM=""
PROJECT="devhub"
REINGEST=0
QUIET=0
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${VAULT_ROOT:-${SRC}/ai-workflow/wiki}"
LOG_TARGET="${SRC}/ai-workflow/memory/log.md"
WIKI_INGEST_WRAPPER="$SCRIPT_DIR/wiki-ingest-from-raw.sh"

usage() {
  cat <<'EOF'
Usage: bash scripts/wiki-pr-update.sh [options]

Options:
  --pr <num>                Required. PR number.
  --project <name>          Default: devhub. (in-repo 만, my_harness 미참조)
  --reingest                PR touched file 마다 wiki-ingest-from-raw --source <file> --apply re-run.
                            (touched file 이 mirror-list 의 source 와 매칭 시만)
  --apply                   실제 vault 갱신 (default = dry-run).
  --quiet                   stderr 메시지 최소화.
  -h, --help                도움말.

Examples:
  # 1. dry-run preview (PR metadata + side effect preview, no actual change)
  bash scripts/wiki-pr-update.sh --pr 552

  # 2. 실제 vault 갱신 (prs/<num>.md + log.md + index.md)
  bash scripts/wiki-pr-update.sh --pr 552 --apply

  # 3. touched file re-ingest + vault 갱신
  bash scripts/wiki-pr-update.sh --pr 552 --reingest --apply
EOF
}

# ----- parse options -----
while [[ $# -gt 0 ]]; do
  case "$1" in
    --pr)
      PR_NUM="$2"
      shift 2
      ;;
    --project)
      PROJECT="$2"
      shift 2
      ;;
    --reingest)
      REINGEST=1
      shift
      ;;
    --apply)
      DRY_RUN=0
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
      echo "[wiki-pr-update] error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

# ----- validation -----
if [[ -z "$PR_NUM" ]]; then
  echo "[wiki-pr-update] error: --pr required" >&2
  exit 2
fi
if ! [[ "$PR_NUM" =~ ^[0-9]+$ ]]; then
  echo "[wiki-pr-update] error: --pr must be numeric: $PR_NUM" >&2
  exit 2
fi
if [[ "$PROJECT" != "devhub" ]]; then
  echo "[wiki-pr-update] error: invalid --project: $PROJECT (must be devhub, in-repo only)" >&2
  echo "[wiki-pr-update] hint: 2026-06-15+ 결정 — my_harness wiki-* skill 미참조. in-repo 만 운영." >&2
  exit 2
fi

# ----- gh CLI 부재 확인 -----
if ! command -v gh >/dev/null 2>&1; then
  echo "[wiki-pr-update] error: gh CLI 미설치 (https://cli.github.com)" >&2
  exit 1
fi

# ----- gh 인증 확인 -----
if ! gh auth status >/dev/null 2>&1; then
  echo "[wiki-pr-update] error: gh CLI 미인증. 'gh auth login' 으로 인증 후 재실행" >&2
  exit 1
fi

# ----- log helper -----
log() { [[ $QUIET -eq 1 ]] || echo "$@"; }

# ----- step 1: PR metadata + touched files 추출 -----
log "[wiki-pr-update] step 1/3: PR metadata extract (#$PR_NUM, project=$PROJECT)"

# PR metadata JSON
PR_JSON=$(gh pr view "$PR_NUM" --json number,title,state,author,baseRefName,headRefName,headRefOid,mergedAt,additions,deletions,changedFiles,body,labels 2>&1) || {
  echo "[wiki-pr-update] error: gh pr view $PR_NUM failed:" >&2
  echo "$PR_JSON" >&2
  exit 1
}

PR_STATE=$(echo "$PR_JSON" | python3 -c "import json,sys; print(json.load(sys.stdin)['state'])")
PR_TITLE=$(echo "$PR_JSON" | python3 -c "import json,sys; print(json.load(sys.stdin)['title'])")
PR_HEAD_SHA=$(echo "$PR_JSON" | python3 -c "import json,sys; print(json.load(sys.stdin)['headRefOid'])")
PR_MERGED_AT=$(echo "$PR_JSON" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('mergedAt',''))")

log "[wiki-pr-update]   state: $PR_STATE"
log "[wiki-pr-update]   title: $PR_TITLE"
log "[wiki-pr-update]   head.sha: $PR_HEAD_SHA"
[[ -n "$PR_MERGED_AT" ]] && log "[wiki-pr-update]   mergedAt: $PR_MERGED_AT"

# touched files: PR diff (gh pr diff --name-only)
TOUCHED_FILES=$(gh pr diff "$PR_NUM" --name-only 2>/dev/null || true)
TOUCHED_COUNT=$(echo "$TOUCHED_FILES" | grep -c . 2>/dev/null || echo 0)
log "[wiki-pr-update]   touched files: $TOUCHED_COUNT"

# Idempotency: 이미 갱신된 PR 인지 확인
PR_PAGE_DIR="$VAULT_ROOT/prs/$PROJECT"
PR_PAGE_PATH="$PR_PAGE_DIR/prs/$PR_NUM.md"
IDEMPOTENCY_KEY="pr-$PR_NUM-$PR_HEAD_SHA"
if [[ -f "$PR_PAGE_PATH" ]]; then
  EXISTING_KEY=$(grep -oE "^idempotency_key:.*$" "$PR_PAGE_PATH" 2>/dev/null | head -1 | cut -d' ' -f2)
  if [[ "$EXISTING_KEY" == "$IDEMPOTENCY_KEY" ]]; then
    log "[wiki-pr-update]   idempotency: 이미 갱신된 PR (key=$IDEMPOTENCY_KEY, skip)"
    log "[wiki-pr-update] DONE (idempotent skip)"
    exit 0
  else
    log "[wiki-pr-update]   idempotency: 기존 key=$EXISTING_KEY, 신규 key=$IDEMPOTENCY_KEY (re-emit)"
  fi
fi

# ----- step 2: (--reingest) touched file 마다 re-ingest -----
if [[ $REINGEST -eq 1 && -n "$TOUCHED_FILES" ]]; then
  log "[wiki-pr-update] step 2/3: re-ingest ($TOUCHED_COUNT files)"
  REINGEST_COUNT=0
  SKIP_COUNT=0
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    case "$f" in
      docs/adr/*|docs/governance/*|docs/planning/*|docs/setup/*|docs/requirements.md|docs/openapi.yaml|ai-workflow/memory/state.json|ai-workflow/memory/session_handoff.md|ai-workflow/memory/work_backlog.md)
        log "[wiki-pr-update]   re-ingest: $f"
        if [[ $DRY_RUN -eq 0 ]]; then
          bash "$WIKI_INGEST_WRAPPER" --project "$PROJECT" --source "$f" --apply --skip-lint --quiet 2>/dev/null || \
          bash "$WIKI_INGEST_WRAPPER" --project "$PROJECT" --source "$f" --apply --skip-lint
        else
          bash "$WIKI_INGEST_WRAPPER" --project "$PROJECT" --source "$f" --skip-lint --quiet 2>/dev/null || \
          bash "$WIKI_INGEST_WRAPPER" --project "$PROJECT" --source "$f" --skip-lint
        fi
        REINGEST_COUNT=$((REINGEST_COUNT + 1))
        ;;
      *)
        log "[wiki-pr-update]   skip (not in mirror list): $f"
        SKIP_COUNT=$((SKIP_COUNT + 1))
        ;;
    esac
  done <<< "$TOUCHED_FILES"
  log "[wiki-pr-update]   re-ingested: $REINGEST_COUNT, skipped: $SKIP_COUNT"
fi

# ----- step 3: vault side effects (자체 inline python) -----
log "[wiki-pr-update] step 3/3: vault side effects (prs/<num>.md + log.md + index.md)"

# PR JSON 을 임시 파일로 전달
PR_JSON_TMP=$(mktemp -t pr-update-XXXXXX.json)
trap 'rm -f "$PR_JSON_TMP"' EXIT
echo "$PR_JSON" > "$PR_JSON_TMP"

if [[ -n "$TOUCHED_FILES" ]]; then
  TOUCHED_FILES_TMP=$(mktemp -t touched-XXXXXX.txt)
  trap 'rm -f "$PR_JSON_TMP" "$TOUCHED_FILES_TMP"' EXIT
  echo "$TOUCHED_FILES" > "$TOUCHED_FILES_TMP"
else
  TOUCHED_FILES_TMP=""
fi

if [[ $DRY_RUN -eq 0 ]]; then
  # apply: prs/<num>.md + log.md + index.md 갱신
  PR_NUM="$PR_NUM" PROJECT="$PROJECT" PR_JSON_TMP="$PR_JSON_TMP" TOUCHED_FILES_TMP="$TOUCHED_FILES_TMP" VAULT_ROOT="$VAULT_ROOT" LOG_TARGET="$LOG_TARGET" IDEMPOTENCY_KEY="$IDEMPOTENCY_KEY" python3 <<'PYEOF'
"""wiki-pr-update in-repo: prs/<num>.md + log.md + index.md 갱신."""
import json
import os
import re
from datetime import datetime, timezone
from pathlib import Path

pr_num = os.environ["PR_NUM"]
project = os.environ["PROJECT"]
vault = Path(os.environ["VAULT_ROOT"])
log_target = Path(os.environ["LOG_TARGET"])
idem = os.environ["IDEMPOTENCY_KEY"]

with open(os.environ["PR_JSON_TMP"], encoding="utf-8") as f:
    pr = json.load(f)
touched = []
tf = os.environ.get("TOUCHED_FILES_TMP", "")
if tf and Path(tf).is_file():
    touched = [line.strip() for line in Path(tf).read_text(encoding="utf-8").split("\n") if line.strip()]

# prs/<num>.md 생성
pr_page_dir = vault / "prs" / project / "prs"
pr_page_dir.mkdir(parents=True, exist_ok=True)
pr_page = pr_page_dir / f"{pr_num}.md"
body = (pr.get("body") or "").strip()
labels = [l["name"] for l in pr.get("labels", [])]
author = pr.get("author", {}).get("login", "unknown")
content = f"""---
type: comparison
status: active
idempotency_key: {idem}
related_pages: [sources/devhub-overview]
created: {datetime.now(timezone.utc).strftime("%Y-%m-%d")}
updated: {datetime.now(timezone.utc).strftime("%Y-%m-%d")}
last_touched: {datetime.now(timezone.utc).strftime("%Y-%m-%d")}
last_ingested_from: gh pr view {pr_num}
---

# PR #{pr_num}: {pr['title']}

## Metadata

- **state**: `{pr.get('state', 'unknown')}`
- **author**: `{author}`
- **baseRef**: `{pr.get('baseRefName', 'unknown')}`
- **headRef**: `{pr.get('headRefName', 'unknown')}`
- **head SHA**: `{pr.get('headRefOid', 'unknown')}`
- **mergedAt**: `{pr.get('mergedAt') or 'n/a'}`
- **additions**: {pr.get('additions', 0)} / **deletions**: {pr.get('deletions', 0)}
- **changedFiles**: {pr.get('changedFiles', 0)}
- **labels**: {labels}

## Body

{body or '(no body)'}

## Touched files ({len(touched)})

{chr(10).join(f"- `{f}`" for f in touched) if touched else '(no files)'}

## Source

- 본 prs/<num>.md: `ai-workflow/wiki/prs/{project}/prs/{pr_num}.md`
- 본 PR (GitHub): https://github.com/ykylee/Devhub_example/pull/{pr_num}
- 본 script: `scripts/wiki-pr-update.sh` (in-repo, my_harness 미참조, v0.7.17+)
"""
pr_page.write_text(content, encoding="utf-8")

# log.md append
log_target.parent.mkdir(parents=True, exist_ok=True)
ts = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
log_line = f"[{ts}] pr-update | PR #{pr_num} | {pr.get('headRefOid', 'unknown')[:7]} | state={pr.get('state', 'unknown')} | files={len(touched)} | idem={idem}"
with open(log_target, "a", encoding="utf-8") as f:
    f.write(log_line + "\n")

# index.md 갱신 (prs 섹션)
index_path = vault / "index.md"
if index_path.is_file():
    idx_text = index_path.read_text(encoding="utf-8")
    # 기존 prs 섹션 갱신: ## PRs 섹션이 없으면 추가
    if "## PRs" not in idx_text:
        idx_text += "\n\n## PRs\n\n"
    prs_section_re = re.compile(r"## PRs\n\n(.*?)(?=\n## |\Z)", re.DOTALL)
    m = prs_section_re.search(idx_text)
    if m:
        existing = m.group(1)
        if f"#{pr_num}" not in existing:
            new_pr_line = f"- [#{pr_num}: {pr['title']}](prs/{project}/prs/{pr_num}.md) (state={pr.get('state', 'unknown')}, head={pr.get('headRefOid', 'unknown')[:7]})\n"
            new_section = existing + new_pr_line
            idx_text = idx_text.replace(m.group(0), "## PRs\n\n" + new_section)
            index_path.write_text(idx_text, encoding="utf-8")

print(f"[wiki-pr-update]   prs/{project}/prs/{pr_num}.md: 1 file created")
print(f"[wiki-pr-update]   log.md: 1 line appended")
print(f"[wiki-pr-update]   index.md: PRs 섹션 갱신")
PYEOF
else
  log "[wiki-pr-update]   dry-run: prs/<num>.md + log.md + index.md preview (no actual change)"
  log "[wiki-pr-update]     - prs/$PROJECT/prs/$PR_NUM.md (1 file would be created)"
  log "[wiki-pr-update]     - log.md (1 line would be appended)"
  log "[wiki-pr-update]     - index.md (PRs 섹션 갱신)"
fi

log ""
log "[wiki-pr-update] DONE"
log "[wiki-pr-update]   prs/<num>.md: $VAULT_ROOT/prs/$PROJECT/prs/$PR_NUM.md"
log "[wiki-pr-update]   log.md: $LOG_TARGET (1 line appended)"
log "[wiki-pr-update]   index.md: PRs 섹션 갱신"
if [[ $REINGEST -eq 0 ]]; then
  log "[wiki-pr-update]   (no re-ingest — touched files 는 wiki/ 의 기존 sources/<title>.md 그대로 유지)"
fi
log "[wiki-pr-update]   my_harness 호출: 0 (2026-06-15+ 결정)"
