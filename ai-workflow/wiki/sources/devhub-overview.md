---
type: concept
status: active
last_ingested_from: ai-workflow/wiki/concepts/devhub-overview.md
related_pages: [sources/devhub-overview]
created: 2026-06-15
updated: 2026-06-15
last_touched: 2026-06-22T04:24:49Z
git_commit: e91115f0
git_branch: chore/260622-wiki-drift-cleanup-2
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
mirror_dirty: |
---

# DevHub Overview (L2 dense, in-repo)

> L1 SSOT: `ai-workflow/wiki/concepts/devhub-overview.md`
> 본 L2 derived view 는 in-repo retrieval 용 압축 요약. dense content 는 L1 SSOT 참조.

## TL;DR

DevHub (ykylee/Devhub_example) = 풀스택 (frontend + backend + infra + governance) 의 데모/레퍼런스 플랫폼. v0.1.0 의 3 domain (인증·조직·대시보드 / Platform·Repository·Project / Cross-cut) + Keycloak OIDC + PostgreSQL + Gitea SCM. 본 저장소 의 *모든 운영/코드/문서* 가 이 위키 의 1차 layer 로 reference 가능 (v0.7.17 in-repo redirect, 2026-06-15).

## 1. 3 domain

- **인증·조직·대시보드** (Authentication/Organization/Dashboard): Keycloak OIDC + `/admin/settings/*` (users/organization/permissions/audit/dev-requests/dev-request-tokens/integrations/integration-bindings) + dashboard (developer/manager/admin) + 역할 routing. 4 role: system_admin/developer/manager/team_manager. sub-carve B (account/user 폐기 + lazy auto-create) = v0.1.0 의 마지막 큰 backend 변경.
- **Platform·Repository·Project** (Domain 2): Platform 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog. Gitea SCM hourly pull (PR #595 X-5) → `pr_activities` + `build_runs` + `quality_snapshots` + `repository_pull_state` upsert. Inbound Source Routing (X-2, PR #586 5차) = AutoRouter interface + Gitea/Jira/GitHub/GitLab 4 provider regex + graceful degradation.
- **Cross-cut** (Domain 3): Audit (Keycloak event webhook + puller + audit_logs X-Request-ID) + RBAC (rbac_policies + PermissionCache LISTEN/NOTIFY) + CI (PR #599 pre-PR hook 7-step) + v0.7.17 vendor import.

## 2. Tier 정책 (사외/사내 2-tier)

GitHub (사외) = single source-of-truth. 사내 SCM = read-only pull. 사내 한정 (DEVHUB_KEYCLOAK_*, internal-registry, 172.16.0.0/12 등) = `infrastructure/`, `infra/idp/`, `.env.deploy`, `docker-compose.{local,test,deploy,colima}.yml` 등 격리. `docs/governance/worker_division.md` §6 + AGENTS.md 2-tier 형상관리 참조.

## 3. 운영 메타 (2026-06-15)

- Repo: github.com/ykylee/Devhub_example, main `dd20266a` (PR #600 머지)
- System version: v0.1.1-alpha / Workflow version: v0.5.11-beta
- WIP PR: #601 (v0.7.17 SSOT 동기화, CI SUCCESS)
- Stack: Keycloak OIDC + PostgreSQL (migrations 000001~000045+) + Gitea private instance + Playwright e2e shard 1/2/3 + Vitest frontend unit

## 4. Wiki 운영 (v0.7.17 in-repo)

- **L1 raw mirror**: `ai-workflow/wiki/raw/projects/devhub/` = 964 file (8.0M) byte-identical (D-72 Phase 1+1.5+3, 15 패턴)
- **L1 page**: `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/<stem>.md` (A안 5 page: devhub-overview / v0.7.17-import / keycloak-iam / in-repo-redirect / standard-ai-workflow-vendor)
- **L2 dense**: `ai-workflow/wiki/sources/<stem>.md` (A안 5 page dense)
- **Event log**: `ai-workflow/memory/log.md`
- **외부 vault `~/wiki/`**: 사용 안 함 (v0.7.17 결정, 2026-06-15)

## 5. Follow-up (별도 PR)

- v0.7.15 atomic_write / v0.7.15 changelog-gen / v0.7.16 workflow-doctor config thresholds 의 DevHub 자체 도구 도입
- 7 script 의 my_harness wiki-* skill 호출의 in-repo vendor tool redirect
- L1/L2 page 의 자동 emit 도구 (vendor 의 emit_wiki_l2_body.py 의 devhub adapter)
- v0.1.0 sub-carve B 완료 → 정식 release
