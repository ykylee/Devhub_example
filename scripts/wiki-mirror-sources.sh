#!/usr/bin/env bash
# wiki-mirror-sources.sh — list_sources helper (Phase 1+1.5+3 mirror patterns).
#
# 본 file 의 source-of-truth:
#   - docs/llm-wiki/mirror-list.md (15 패턴 문서)
#   - scripts/wiki-sync-devhub.sh (mirror sync script)
#   - tests/check_mirror_list_devhub.py (byte-identity 검증 test, PR #616 follow-up)
#
# 본 file 은 bash function library 이며, 직접 실행 ❌ — 다른 script 에서
# \`source\` 로 import 후 \`list_sources\` 함수 호출.
#
# Usage (in scripts/wiki-sync-devhub.sh):
#   source "$(dirname "${BASH_SOURCE[0]}")/wiki-mirror-sources.sh"
#   list_sources  # stdout 에 mirror source path list emit
#
# Usage (in test, with subshell):
#   SRC="$REPO_ROOT" bash -c 'source scripts/wiki-mirror-sources.sh && list_sources'
#
# Phase 1 (7 패턴) + Phase 1.5 (6 패턴) + Phase 3 (2 패턴) = 15 패턴. Phase 1+1.5+3 의
# 정확한 정합은 docs/llm-wiki/mirror-list.md §1.7.1~§1.7.2 + §2 (Phase 1+1.5+3 패턴 표) 참조.

# ----- guard: prevent direct execution -----
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "[wiki-mirror-sources] error: 본 script 는 source-able library 입니다. 직접 실행 ❌" >&2
  echo "  사용법: 다른 script 에서 \`source scripts/wiki-mirror-sources.sh\` 후 \`list_sources\` 호출" >&2
  exit 1
fi

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
