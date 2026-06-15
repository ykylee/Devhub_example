#!/usr/bin/env bash
# wiki-sync-devhub.sh — DevHub repo (~/repos/Devhub_example_minimax/) 의 의미 있는 파일을
# ai-workflow/wiki/raw/projects/devhub/ 로 mirror. my_harness 의 wiki-sync-ai-workflow.sh 와 동일 패턴.
#
# 사용:
#   bash scripts/wiki-sync-devhub.sh              # real mirror
#   bash scripts/wiki-sync-devhub.sh --dry-run    # source list 출력 (no actual mirror)
#   bash scripts/wiki-sync-devhub.sh --no-clean   # real mirror, no DEST clean (incremental)
#
# 본 script 의 source-of-truth:
#   - mirror list: docs/llm-wiki/mirror-list.md (§2 의 13 source 패턴)
#   - lint config: docs/llm-wiki/lint-config.toml
#   - operation SOP: docs/llm-wiki/operation-sop.md
#
# 결정적이고 단순: find + mkdir -p + cp. rsync 의 include/exclude 충돌 회피 (my_harness 의
# wiki-sync-ai-workflow.sh 와 동일 pattern).
#
# D-72 Phase 1 (2026-06-10) 의 본 저장소 측 sync script. lint L11 (사내 패턴 검출) +
# sa-internal/ 격리 + mirror policy (T-d-72-2) 의 source.
#
# Phase 1.5 (2026-06-13) 갱신:
#   - source code + workflow + scripts + branch memory + traceability 추가 (13 패턴, 6 추가)
#   - 본 sprint 의 22 file + maintenance critical subset ~33 file = ~55 file
#   - 본 저장소 한정 + wiki 만으로 코드 maintenance 가능 정공법
#
# Exit code:
#   0 — success (real mirror 또는 dry-run)
#   1 — vault 부재 또는 source root 부재 또는 invalid option

set -euo pipefail

# ----- options -----
DRY_RUN=0
NO_CLEAN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=1
      ;;
    --no-clean)
      NO_CLEAN=1
      ;;
    -h|--help)
      echo "usage: bash scripts/wiki-sync-devhub.sh [--dry-run] [--no-clean]"
      echo "  --dry-run   source list 출력 (no actual mirror)"
      echo "  --no-clean  real mirror, no DEST clean (incremental)"
      exit 0
      ;;
    *)
      echo "[wiki-sync-devhub] invalid option: $arg" >&2
      echo "  hint: --dry-run / --no-clean / (no option) 만 지원" >&2
      exit 1
      ;;
  esac
done

# ----- paths -----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$(cd "$SCRIPT_DIR/.." && pwd)"
VAULT_ROOT="${SRC}/ai-workflow/wiki"
DEST="$VAULT_ROOT/raw/projects/devhub"

# ----- validation -----
if [[ ! -d "$SRC" ]]; then
  echo "[wiki-sync-devhub] source root 부재: $SRC" >&2
  exit 1
fi

if [[ ! -d "$VAULT_ROOT" ]]; then
  echo "[wiki-sync-devhub] target vault not found: $VAULT_ROOT" >&2
  echo "[wiki-sync-devhub] hint: 'wiki-init' 명령으로 vault 초기화 (my_harness 측 D-71 §2.2)" >&2
  echo "[wiki-sync-devhub] hint: 또는 my_harness 의 D-72 응답 §4 의 next-step #1 참고" >&2
  exit 1
fi

echo "[wiki-sync-devhub] source root: $SRC"
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[wiki-sync-devhub] dry-run: True (no actual mirror)"
else
  echo "[wiki-sync-devhub] dry-run: False (real mirror)"
fi
if [[ $NO_CLEAN -eq 1 ]]; then
  echo "[wiki-sync-devhub] no-clean: True (incremental, $DEST 의 기존 file 유지)"
else
  echo "[wiki-sync-devhub] no-clean: False (clean $DEST, fresh mirror)"
fi
echo "[wiki-sync-devhub] target vault: $VAULT_ROOT (Gitea private)"

# ----- mirror list (Phase 1 + Phase 1.5 + Phase 3 source patterns) -----
# 15 패턴 (docs/llm-wiki/mirror-list.md §2):
#   Phase 1 (7 패턴):
#     1. ADR: docs/adr/0[0-9][0-9][0-9]-*.md
#     2. Governance: docs/governance/*.md
#     3. Planning: docs/planning/*.md
#     4. Setup: docs/setup/*.md
#     5. Requirements: docs/requirements.md
#     6. OpenAPI: docs/openapi.yaml
#     7. AI-workflow memory (main flat): ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}
#   Phase 1.5 (6 패턴):
#     8. Workflows: .github/workflows/*.yml
#     9. Scripts (화이트리스트): scripts/{wiki-sync-devhub,select-playwright-specs,ci-e2e-sync-check,check-migration-uniqueness}.sh
#    10. Backend critical Go (화이트리스트): backend-core/internal/{auth,domain,httpapi,audit,rbac,store,sso-integrations}/ 의 key file
#    11. Frontend e2e critical (화이트리스트): frontend/tests/e2e/{fixtures,signout,voc-*}.ts + frontend/tests/e2e-manifests/*.txt + frontend/lib/auth/{tokenStore,apiClient,role-routing}.ts
#    12. Traceability: docs/traceability/{README,conventions,report,sync-checklist}.md
#    13. Branch memory (active + 30일 이내 CLOSED): ai-workflow/memory/<agent>/<branch>/{state.json, session_handoff.md, work_backlog.md, backlog/YYYY-MM-DD.md, pr_body.md}
#   Phase 3 (2 패턴):
#    14. Domain: docs/domain/**/*.md (66 file)
#    15. Architecture + Infrastructure + Validation: docs/{architecture,infrastructure,validation}/*.md (12 file)

# ----- helper: list all mirror sources -----
list_sources() {
  # Phase 1 (7 패턴)
  find "$SRC/docs/adr" -type f -name "0[0-9][0-9][0-9]-*.md" 2>/dev/null
  find "$SRC/docs/governance" -type f -name "*.md" 2>/dev/null
  find "$SRC/docs/planning" -type f -name "*.md" 2>/dev/null
  find "$SRC/docs/setup" -maxdepth 1 -type f -name "*.md" 2>/dev/null
  [[ -f "$SRC/docs/requirements.md" ]] && echo "$SRC/docs/requirements.md"
  [[ -f "$SRC/docs/openapi.yaml" ]] && echo "$SRC/docs/openapi.yaml"
  for f in state.json session_handoff.md work_backlog.md; do
    [[ -f "$SRC/ai-workflow/memory/$f" ]] && echo "$SRC/ai-workflow/memory/$f"
  done

  # Phase 1.5: 8. Workflows
  [[ -d "$SRC/.github/workflows" ]] && find "$SRC/.github/workflows" -type f -name "*.yml" 2>/dev/null

  # Phase 1.5: 9. Scripts (화이트리스트)
  for f in wiki-sync-devhub.sh select-playwright-specs.sh ci-e2e-sync-check.sh check-migration-uniqueness.sh; do
    [[ -f "$SRC/scripts/$f" ]] && echo "$SRC/scripts/$f"
  done

  # Phase 1.5: 10. Backend critical Go (화이트리스트)
  # 본 sprint 의 verification 으로 정확한 경로 정합 (2026-06-13)
  # mirror-list.md §1.7.1 정합
  for f in \
    "backend-core/main.go" \
    "backend-core/internal/sso-integrations/keycloak/saovae_stub.go" \
    "backend-core/internal/domain/auth-session/integration/ports.go" \
    "backend-core/internal/domain/auth-session/view/auth.go" \
    "backend-core/internal/domain/auth-session/view/handler.go" \
    "backend-core/internal/domain/application-lifecycle/routing/auto_route.go" \
    "backend-core/internal/domain/dev-request/view/voc_handler.go" \
    "backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go" \
    "backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go" \
    "backend-core/internal/domain/audit-ops/view/audit.go" \
    "backend-core/internal/domain/audit-ops/repository/audit_logs.go" \
    "backend-core/internal/domain/rbac.go" \
    "backend-core/internal/domain/rbac-permissions/view/rbac.go" \
    "backend-core/internal/httpapi/repository_ops.go" \
    "backend-core/internal/store/repository_ops.go"; do
    [[ -f "$SRC/$f" ]] && echo "$SRC/$f"
  done

  # Phase 1.5: 11. Frontend e2e critical (화이트리스트)
  # mirror-list.md §1.7.2 정합
  for f in \
    "frontend/tests/e2e/fixtures.ts" \
    "frontend/tests/e2e/signout.spec.ts" \
    "frontend/tests/e2e/voc-auto-routing.spec.ts" \
    "frontend/tests/e2e-manifests/smoke.txt" \
    "frontend/tests/e2e-manifests/quarantine.txt" \
    "frontend/lib/store.ts" \
    "frontend/domain/auth-session/service/role-routing.ts"; do
    [[ -f "$SRC/$f" ]] && echo "$SRC/$f"
  done

  # Phase 1.5: 12. Traceability
  for f in README.md conventions.md report.md sync-checklist.md; do
    [[ -f "$SRC/docs/traceability/$f" ]] && echo "$SRC/docs/traceability/$f"
  done

  # Phase 1.5: 13. Branch memory (active + 30일 이내 CLOSED)
  # 본 sprint 의 active 4 branch + 본 sprint 진행 중 CLOSED branch
  # future: archive 시 PR 머지 + 30일 후 mavis-trash 권장
  if [[ -d "$SRC/ai-workflow/memory" ]]; then
    # agent prefix 디렉터리 (codex/, claude/, gemini/, deepseek/, opencode/, mvs/, ...)
    for agent_dir in "$SRC/ai-workflow/memory"/*/; do
      [[ ! -d "$agent_dir" ]] && continue
      agent_name=$(basename "$agent_dir")
      for branch_dir in "$agent_dir"*/; do
        [[ ! -d "$branch_dir" ]] && continue
        branch_name=$(basename "$branch_dir")
        # core memory 3 file
        for f in state.json session_handoff.md work_backlog.md; do
          [[ -f "$branch_dir$f" ]] && echo "$branch_dir$f"
        done
        # backlog/YYYY-MM-DD.md
        if [[ -d "$branch_dir/backlog" ]]; then
          find "$branch_dir/backlog" -type f -name "*.md" 2>/dev/null
        fi
        # pr_body.md
        [[ -f "$branch_dir/pr_body.md" ]] && echo "$branch_dir/pr_body.md"
      done
    done
  fi

  # Phase 3: 14. Domain (~66 file)
  [[ -d "$SRC/docs/domain" ]] && find "$SRC/docs/domain" -type f -name "*.md" 2>/dev/null

  # Phase 3: 15. Architecture + Infrastructure + Validation (~12 file)
  for d in architecture infrastructure validation; do
    [[ -d "$SRC/docs/$d" ]] && find "$SRC/docs/$d" -type f -name "*.md" 2>/dev/null
  done

  return 0
}

# ----- 1. dry-run: source list 출력 -----
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[wiki-sync-devhub] collecting files (dry-run)..."

  echo ""
  echo "  === Phase 1 (docs subset, 7 패턴) ==="

  echo ""
  echo "  ADR (docs/adr/0[0-9][0-9][0-9]-*.md):"
  find "$SRC/docs/adr" -type f -name "0[0-9][0-9][0-9]-*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Governance (docs/governance/*.md):"
  find "$SRC/docs/governance" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Planning (docs/planning/*.md):"
  find "$SRC/docs/planning" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Setup (docs/setup/*.md):"
  find "$SRC/docs/setup" -maxdepth 1 -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Requirements (docs/requirements.md):"
  [[ -f "$SRC/docs/requirements.md" ]] && echo "    docs/requirements.md"

  echo ""
  echo "  OpenAPI (docs/openapi.yaml):"
  [[ -f "$SRC/docs/openapi.yaml" ]] && echo "    docs/openapi.yaml"

  echo ""
  echo "  AI-workflow memory (main flat, 3 file):"
  for f in state.json session_handoff.md work_backlog.md; do
    [[ -f "$SRC/ai-workflow/memory/$f" ]] && echo "    ai-workflow/memory/$f"
  done

  echo ""
  echo "  === Phase 1.5 (source code + workflow + scripts + branch memory + traceability, 6 패턴) ==="

  echo ""
  echo "  Workflows (.github/workflows/*.yml):"
  [[ -d "$SRC/.github/workflows" ]] && find "$SRC/.github/workflows" -type f -name "*.yml" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Scripts (화이트리스트):"
  for f in wiki-sync-devhub.sh select-playwright-specs.sh ci-e2e-sync-check.sh check-migration-uniqueness.sh; do
    [[ -f "$SRC/scripts/$f" ]] && echo "    scripts/$f"
  done

  echo ""
  echo "  Backend critical Go (화이트리스트, 15 file):"
  for f in \
    "backend-core/main.go" \
    "backend-core/internal/sso-integrations/keycloak/saovae_stub.go" \
    "backend-core/internal/domain/auth-session/integration/ports.go" \
    "backend-core/internal/domain/auth-session/view/auth.go" \
    "backend-core/internal/domain/auth-session/view/handler.go" \
    "backend-core/internal/domain/application-lifecycle/routing/auto_route.go" \
    "backend-core/internal/domain/dev-request/view/voc_handler.go" \
    "backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go" \
    "backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go" \
    "backend-core/internal/domain/audit-ops/view/audit.go" \
    "backend-core/internal/domain/audit-ops/repository/audit_logs.go" \
    "backend-core/internal/domain/rbac.go" \
    "backend-core/internal/domain/rbac-permissions/view/rbac.go" \
    "backend-core/internal/httpapi/repository_ops.go" \
    "backend-core/internal/store/repository_ops.go"; do
    [[ -f "$SRC/$f" ]] && echo "    $f"
  done

  echo ""
  echo "  Frontend e2e critical (화이트리스트, 7 file):"
  for f in \
    "frontend/tests/e2e/fixtures.ts" \
    "frontend/tests/e2e/signout.spec.ts" \
    "frontend/tests/e2e/voc-auto-routing.spec.ts" \
    "frontend/tests/e2e-manifests/smoke.txt" \
    "frontend/tests/e2e-manifests/quarantine.txt" \
    "frontend/lib/store.ts" \
    "frontend/domain/auth-session/service/role-routing.ts"; do
    [[ -f "$SRC/$f" ]] && echo "    $f"
  done

  echo ""
  echo "  Traceability (4 file):"
  for f in README.md conventions.md report.md sync-checklist.md; do
    [[ -f "$SRC/docs/traceability/$f" ]] && echo "    docs/traceability/$f"
  done

  echo ""
  echo "  Branch memory (active + 30일 이내 CLOSED):"
  if [[ -d "$SRC/ai-workflow/memory" ]]; then
    for agent_dir in "$SRC/ai-workflow/memory"/*/; do
      [[ ! -d "$agent_dir" ]] && continue
      agent_name=$(basename "$agent_dir")
      for branch_dir in "$agent_dir"*/; do
        [[ ! -d "$branch_dir" ]] && continue
        branch_name=$(basename "$branch_dir")
        for f in state.json session_handoff.md work_backlog.md; do
          [[ -f "$branch_dir$f" ]] && echo "    ai-workflow/memory/$agent_name/$branch_name/$f"
        done
        if [[ -d "$branch_dir/backlog" ]]; then
          find "$branch_dir/backlog" -type f -name "*.md" 2>/dev/null | sed "s|^$SRC/||" | sed 's/^/    /' || true
        fi
        [[ -f "$branch_dir/pr_body.md" ]] && echo "    ai-workflow/memory/$agent_name/$branch_name/pr_body.md"
      done
    done
  fi

  echo ""
  echo "  === Phase 3 (mass ingest, 2 패턴) ==="

  echo ""
  echo "  Domain (docs/domain/**/*.md, ~66 file):"
  [[ -d "$SRC/docs/domain" ]] && find "$SRC/docs/domain" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Architecture (docs/architecture/*.md, 1 file):"
  [[ -d "$SRC/docs/architecture" ]] && find "$SRC/docs/architecture" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Infrastructure (docs/infrastructure/**/*.md, 9 file):"
  [[ -d "$SRC/docs/infrastructure" ]] && find "$SRC/docs/infrastructure" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  echo ""
  echo "  Validation (docs/validation/*.md, 2 file):"
  [[ -d "$SRC/docs/validation" ]] && find "$SRC/docs/validation" -type f -name "*.md" 2>/dev/null | sort | sed "s|^$SRC/||" | sed 's/^/    /' || true

  COUNT=$(list_sources | wc -l | tr -d ' ')
  echo ""
  echo "[wiki-sync-devhub] total: $COUNT file (estimated)"
  echo "[wiki-sync-devhub] dry-run: no changes made"
  echo "[wiki-sync-devhub] PASS (dry-run)"
  exit 0
fi

# ----- 2. real mirror: DEST 의 clean (--no-clean 시 skip) 후 cp + manifest 자동 -----
if [[ $NO_CLEAN -eq 0 ]]; then
  echo "[wiki-sync-devhub] cleaning $DEST"
  mkdir -p "$DEST"
  find "$DEST" -mindepth 1 -delete 2>/dev/null
else
  echo "[wiki-sync-devhub] --no-clean: skip cleaning $DEST"
fi

echo "[wiki-sync-devhub] collecting files from $SRC"

COUNT=0
list_sources | while IFS= read -r f; do
  rel="${f#"$SRC"/}"
  dest_path="$DEST/$rel"
  mkdir -p "$(dirname "$dest_path")"
  cp -p "$f" "$dest_path"
  COUNT=$((COUNT + 1))
  echo "  copied: $rel"
done

# ----- manifest 자동 생성 (provenance: commit hash + version + branch + dirty) -----
MANIFEST="$DEST/_manifest.md"

# git provenance capture (mirror 의 source 의 SSOT provenance)
GIT_COMMIT=$(git -C "$SRC" rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_COMMIT_SHORT=$(git -C "$SRC" rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git -C "$SRC" branch --show-current 2>/dev/null || echo "unknown")
GIT_DIRTY=$(git -C "$SRC" status --porcelain 2>/dev/null | head -1)
if [[ -n "$GIT_DIRTY" ]]; then
  GIT_DIRTY_FLAG="(dirty: uncommitted changes)"
else
  GIT_DIRTY_FLAG=""
fi
SYNC_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Version capture (DevHub 의 2 version system: /VERSION + ai-workflow/VERSION)
VERSION_SYSTEM=$(cat "$SRC/VERSION" 2>/dev/null | head -1 | tr -d '[:space:]' || echo "unknown")
VERSION_WORKFLOW=$(cat "$SRC/ai-workflow/VERSION" 2>/dev/null | head -1 | tr -d '[:space:]' || echo "unknown")

{
  echo "# raw/projects/devhub/_manifest.md (D-72 Phase 1 + Phase 1.5, provenance tracking)"
  echo ""
  echo "## Provenance"
  echo ""
  echo "본 manifest 는 wiki mirror 의 **provenance tracking** 의 SSOT. 위키의 모든 정보의 상태(시점/commit/version/branch/dirty)를 본 manifest 에서 검증 가능."
  echo ""
  echo "| field | value |"
  echo "| --- | --- |"
  echo "| source repo | $SRC |"
  echo "| target vault | $DEST |"
  echo "| sync tool | scripts/wiki-sync-devhub.sh (BSD-rsync safe, Phase 1.5 13 패턴) |"
  echo "| mirror list | docs/llm-wiki/mirror-list.md §2 |"
  echo "| lint config | docs/llm-wiki/lint-config.toml |"
  echo "| operation SOP | docs/llm-wiki/operation-sop.md |"
  echo ""
  echo "### Git state (mirror 의 source 의 provenance)"
  echo ""
  echo "| field | value |"
  echo "| --- | --- |"
  echo "| commit (full) | $GIT_COMMIT |"
  echo "| commit (short) | $GIT_COMMIT_SHORT |"
  echo "| branch | $GIT_BRANCH |"
  echo "| dirty | $GIT_DIRTY_FLAG |"
  echo ""
  echo "### Version (DevHub 의 2 version system)"
  echo ""
  echo "| field | value |"
  echo "| --- | --- |"
  echo "| system version (root /VERSION) | $VERSION_SYSTEM |"
  echo "| workflow version (ai-workflow/VERSION) | $VERSION_WORKFLOW |"
  echo ""
  echo "### Last sync"
  echo ""
  echo "- timestamp: $SYNC_TIMESTAMP"
  echo "- count: $(find "$DEST" -type f 2>/dev/null | wc -l | tr -d ' ') file"
  echo "- size: $(du -sh "$DEST" 2>/dev/null | cut -f1)"
  echo ""
  echo "## Mirror log (recent, 최신 위)"
  echo ""
  echo "| timestamp | rel_path | size (bytes) |"
  echo "| --- | --- | --- |"
  find "$DEST" -type f -not -name "_manifest.md" 2>/dev/null | sort | while IFS= read -r f; do
    rel="${f#"$DEST"/}"
    size=$(stat -f%z "$f" 2>/dev/null || stat -c%s "$f" 2>/dev/null || echo "?")
    echo "| $SYNC_TIMESTAMP | $rel | $size |"
  done
} > "$MANIFEST"

echo "[wiki-sync-devhub] DONE"
echo "  files: $(find "$DEST" -type f 2>/dev/null | wc -l | tr -d ' ')"
echo "  size:  $(du -sh "$DEST" 2>/dev/null | cut -f1)"
echo "  manifest: $MANIFEST"
echo "  commit: $GIT_COMMIT_SHORT ($GIT_BRANCH) $GIT_DIRTY_FLAG"
echo "  version: system=$VERSION_SYSTEM workflow=$VERSION_WORKFLOW"
