---
type: concept
status: active
last_ingested_from: AGENTS.md + docs/architecture.md + docs/planning/release_v0-1_roadmap.md
related_pages: [decisions/v0.7.17-import, entities/keycloak-iam, patterns/in-repo-redirect, topics/standard-ai-workflow-vendor]
created: 2026-06-15
updated: 2026-06-15
active_since: 2026-06-15
active_reason: "v0.7.17 wiki in-repo redirect (PR #600 머지) + DevHub v0.1.0 roadmap active"
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
---

# DevHub Overview (L1 concept, in-repo)

## TL;DR

DevHub (ykylee/Devhub_example) 는 풀스택 (frontend + backend + infra + governance) 의 데모/레퍼런스 플랫폼. 1차 릴리즈 v0.1.0 의 3 domain (인증/조직/대시보드 / Platform·Repository·Project / Cross-cut) + Keycloak OIDC IdP + PostgreSQL + Gitea SCM 통합. 본 저장소 의 *모든 운영/코드/문서* 가 이 위키 의 1차 layer 로 reference 가능 (v0.7.17 redirect, 2026-06-15).

## 1. 구조 (3 domain)

### 1.1 인증/조직/대시보드
- Keycloak OIDC 로그인 + `/admin/settings/*` (users/organization/permissions/audit/dev-requests/dev-request-tokens/integrations/integration-bindings)
- 대시보드 (developer/manager/admin) + 역할 routing
- 4 role: system_admin / developer / manager / team_manager
- sub-carve B (`/api/v0-1/accounts/*` 폐기 + lazy auto-create) 가 v0.1.0 의 마지막 큰 backend 변경

### 1.2 Platform/Repository/Project 도메인
- Platform 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog
- Gitea SCM hourly pull (production wire, PR #595 X-5) → `pr_activities` + `build_runs` + `quality_snapshots` + `repository_pull_state` upsert
- Inbound Source Routing (X-2, PR #586 5차) — `AutoRouter` interface + Gitea/Jira/GitHub/GitLab 4 provider regex + graceful degradation

### 1.3 Cross-cut
- Audit (Keycloak event webhook + puller + audit_logs X-Request-ID propagation)
- RBAC (rbac_policies seeded + PermissionCache LISTEN/NOTIFY)
- CI (PR #599 pre-PR hook 7-step recovery, PR #598 retro)
- v0.7.17 표준화 워크플로우 vendor import (in-repo)

## 2. Tier 정책 (사외/사내 2-tier)

GitHub (사외) = single source-of-truth (push-only). 사내 SCM = GitHub 에서 read-only pull. 사내 한정 정보 (DEVHUB_KEYCLOAK_*, internal-registry.example.com, 172.16.0.0/12, kc.internal.example.com 등) 는 `infrastructure/` + `infra/idp/` + `.env.deploy` + `docker-compose.{local,test,deploy,colima}.yml` 등 사내 한정 경로에 격리. `docs/governance/worker_division.md` §6 + `AGENTS.md` 사외/사내 2-tier 형상관리 분리 참조.

## 3. 운영 메타

- **Repository**: https://github.com/ykylee/Devhub_example (GitHub, main branch, 7 active branches + 30+ sprint memory)
- **System version** (root /VERSION): v0.1.1-alpha
- **Workflow version** (ai-workflow/VERSION): v0.5.11-beta
- **Active main HEAD** (2026-06-15): `dd20266a` (PR #600 머지, v0.7.17 vendor import + wiki in-repo redirect)
- **WIP PR**: #601 (v0.7.17 SSOT 동기화, CI SUCCESS)
- **Mavis session ID (현재)**: `mvs_8952f7f57f9749a68171434a78f89960`
- **인프라**: Keycloak OIDC (sso-integrations) + PostgreSQL (migrations 000001~000045+) + Gitea (private instance) + Playwright e2e (shard 1/2/3) + Vitest frontend unit

## 4. Wiki 운영 (v0.7.17 in-repo redirect)

- **L1 raw mirror** (in-repo): `ai-workflow/wiki/raw/projects/devhub/` — 964 file (8.0M) byte-identical mirror of docs + code + workflow
- **L1 page** (in-repo): `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/<stem>.md` — frontmatter + 본문 = 1-3 page (5 page A안: devhub-overview, v0.7.17-import, keycloak-iam, in-repo-redirect, standard-ai-workflow-vendor)
- **L2 dense** (in-repo): `ai-workflow/wiki/sources/<stem>.md` — L1 의 dense version (compressed, in-repo retrieval 용)
- **Event log**: `ai-workflow/memory/log.md` (vendor 의 wiki-log target)
- **외부 vault `~/wiki/`** = 사용 안 함 (v0.7.17 redirect 결정, 사용자 2026-06-15)

## 5. 후속 정합 (follow-up)

- v0.7.15 atomic_write / v0.7.15 changelog-gen / v0.7.16 workflow-doctor config thresholds 의 DevHub 자체 도구 도입 (vendor 의 범용 패턴 흡수)
- 7 script 의 my_harness 측 wiki-* skill 호출의 in-repo vendor tool redirect
- L1/L2 page 의 자동 emit 도구 (vendor 의 emit_wiki_l2_body.py 의 *devhub project mode* adapter)
- v0.1.0 의 sub-carve B (account/user management 폐기) 완료 → v0.1.0 정식 release
