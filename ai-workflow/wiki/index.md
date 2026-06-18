# DevHub Wiki (L0 Home, in-repo)

> 본 페이지는 DevHub 의 *in-repo wiki* 진입점. v0.7.17 (PR #600, 2026-06-15) 부터 외부 vault `~/wiki/` 미사용. 본 저장소 의 모든 운영/코드/문서 가 이 위키 의 1차 layer 로 reference 가능.

## 구조 (3 layer)

| layer | 위치 | 역할 | 비고 |
|---|---|---|---|
| **L0 Home** | `ai-workflow/wiki/index.md` (본 file) | 진입점 + 1차 layer | 본 file |
| **L1 raw mirror** | `ai-workflow/wiki/raw/projects/devhub/` | 1:1 byte mirror of source (964 file, 8.0M) | `scripts/wiki-sync-devhub.sh` 의 in-repo redirect 결과. *L1 SSOT* = 본 저장소 의 15 패턴 source file. LLM query 의 1차 layer reasoning. |
| **L1 page** | `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/<stem>.md` | 1차 layer 운영 문서 | frontmatter + 본문 (1-3 page). A안 5 page 작성 완료. |
| **L2 dense** | `ai-workflow/wiki/sources/<stem>.md` | dense derived view | L1 의 *compressed* (2000자 cap). in-repo retrieval 용. A안 5 page 작성 완료. |
| **Event log** | `ai-workflow/memory/log.md` | wiki event log | vendor 의 wiki-log target. PR/merge/release event append. |

## L1 page (A안, 5 page)

| Stem | 위치 | 내용 |
|---|---|---|
| `devhub-overview` | `concepts/devhub-overview.md` | DevHub 의 3 domain + tier 정책 + 운영 메타 + wiki 운영 + follow-up |
| `v0.7.17-import` | `decisions/v0.7.17-import.md` | v0.7.17 vendor import 결정 (4가지) + smoke 16/16 + follow-up + 메모리 anchor |
| `keycloak-iam` | `entities/keycloak-iam.md` | Keycloak IAM 의 4 layer + ADR + tier 분리 + 운영 정공법 + follow-up |
| `in-repo-redirect` | `patterns/in-repo-redirect.md` | 5 file 의 in-repo path redirect + 4-priority REPO_ROOT + 4 smoke + 갱신 SOP |
| `standard-ai-workflow-vendor` | `topics/standard-ai-workflow-vendor.md` | vendor 구조 + 발매자 vs 소비자 + Mavis 운영 + sub-agent 4 role + 작업 모드 6종 |

## L2 dense (A안, 5 page)

각 L1 의 *compressed derived view*. 본 file 의 L1 page list 와 1:1 매핑. L2 의 related_pages field = `sources/<stem>`.

## wiki 운영 정공법 (v0.7.17 in-repo)

### 1. L1 raw mirror 자동 갱신

```bash
# Dry-run (어떤 file 이 mirror 대상인지 확인)
bash scripts/wiki-sync-devhub.sh --dry-run

# Real mirror (DEST 의 clean 후 1:1 byte copy + manifest 자동)
bash scripts/wiki-sync-devhub.sh

# Incremental (DEST 의 기존 file 유지, --no-clean)
bash scripts/wiki-sync-devhub.sh --no-clean
```

`scripts/wiki-sync-devhub.sh` 의 15 패턴 (`docs/llm-wiki/mirror-list.md` §2): docs/adr, docs/governance, docs/planning, docs/setup, requirements.md, openapi.yaml, ai-workflow/memory main flat + branch memory, .github/workflows/*.yml, scripts/*.sh, backend-core critical Go, frontend/tests/e2e + manifests, docs/traceability, docs/domain, docs/architecture + docs/infrastructure + docs/validation.

### 2. L1 page 작성 (수동 또는 vendor 도구)

수동 작성: 본 file 의 L1 page 5 종 의 frontmatter 형식 + 본문.
- `type`: `concept` | `decision` | `entity` | `pattern` | `topic`
- `status`: `active` | `draft`
- `last_ingested_from`: 1차 출처 file path
- `related_pages`: wikilink list (`sources/...` 등)
- `created` / `updated` / `active_since` / `active_reason` field

vendor 도구 (향후):
- `python3 vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py --project=devhub --apply` — L1 page 의 본문 일부를 L2 dense 로 자동 emit
- `python3 vendor/standard_ai_workflow/tools/refresh_wiki_memory.py --refresh-raw --emit-l2` — active memory 의 4 file + wiki log 의 in-repo 갱신
- `python3 vendor/standard_ai_workflow/tools/score_wiki_maintainability.py` — wiki maintainability score
- 본 저장소 의 *devhub adapter* 작성 (vendor 의 emit 도구가 우리 mirror list 의 15 패턴 자동 ingest) = follow-up PR

### 3. wiki 검증 (smoke 16/16)

```bash
# DevHub invariant (5/5)
python3 tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py

# vendor smoke (11/11)
python3 vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py
```

### 4. wiki drift 검증

```bash
python3 vendor/standard_ai_workflow/tests/check_wiki_drift.py
```

L1 raw mirror mtime vs L2 last_touched 비교. 7일 임계값.

## 1차 출처 (R4 anchor)

- **본 저장소 의 1차 layer 운영 문서** (L1): `AGENTS.md` + `ai-workflow/MEMORY_GOVERNANCE.md` + `ai-workflow/minimax_code_workflow.md` + `ai-workflow/memory/PROJECT_PROFILE.md` + 도메인별 `requirements.md` / `architecture.md` / `adr/*`
- **vendor 의 1차 출처 SSOT** (in-repo): `vendor/standard_ai_workflow/core/maturity_matrix.json` + `vendor/standard_ai_workflow/harnesses/minimax-code/README.md` + `vendor/standard_ai_workflow/core/global_workflow_standard.md` + `vendor/standard_ai_workflow/core/orchestrator_subagent_contract_v1.md`
- **외부 vault `~/wiki/`**: 사용 안 함 (v0.7.17 결정, 2026-06-15)

## follow-up (별도 PR)

- A안 5 page 의 운영 검증 후 *전체 220+ file* L1 + L2 확장 (vendor 의 emit_wiki_l2_body.py 의 devhub adapter 작성)
- 7 script 의 my_harness wiki-* skill 호출의 in-repo vendor tool redirect
- v0.7.15 atomic_write / v0.7.15 changelog-gen / v0.7.16 workflow-doctor config thresholds 의 DevHub 자체 도구 도입
- vendor 갱신 (v0.7.18~v0.7.21 의 4 release 흡수 시)

## 메모리 anchor (cross-project)

- **MEMORY.md §12**: "Vendor side-by-side 격리 정공법" (commit 44f47baf)
- **MEMORY.md §13**: "Vendor SSOT → 본 저장소 운영 문서 동기화 정공법" (commit b8eb55d5)
- **MEMORY.md §14** (예정): "Wiki L0/L1/L2 self-bootstrap 정공법" (commit TBD)


## PRs

- [#600: feat(vendor): standard_ai_workflow v0.7.17 import + wiki in-repo redirect](prs/devhub/prs/600.md) (state=MERGED, head=f813493)

- [wiki/projects/devhub/sources/docs/architecture/README.md](sources/docs/architecture/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/README.md](sources/docs/domain/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/audit-ops/README.md](sources/docs/domain/audit-ops/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/audit-ops/api.md](sources/docs/domain/audit-ops/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/audit-ops/architecture.md](sources/docs/domain/audit-ops/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/audit-ops/requirements.md](sources/docs/domain/audit-ops/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/README.md](sources/docs/domain/auth-session/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/account_redesign.md](sources/docs/domain/auth-session/account_redesign.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/api.md](sources/docs/domain/auth-session/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/architecture.md](sources/docs/domain/auth-session/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/requirements.md](sources/docs/domain/auth-session/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/auth-session/test_cases.md](sources/docs/domain/auth-session/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/README.md](sources/docs/domain/dev-request/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/api.md](sources/docs/domain/dev-request/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/architecture.md](sources/docs/domain/dev-request/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/concept.md](sources/docs/domain/dev-request/concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/requirements.md](sources/docs/domain/dev-request/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/dev-request/test_cases.md](sources/docs/domain/dev-request/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/README.md](sources/docs/domain/integration-registry/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/api.md](sources/docs/domain/integration-registry/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/architecture.md](sources/docs/domain/integration-registry/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/capability_matrix.md](sources/docs/domain/integration-registry/capability_matrix.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/external_system_concept.md](sources/docs/domain/integration-registry/external_system_concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/homelab_pull_strategy.md](sources/docs/domain/integration-registry/homelab_pull_strategy.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/prometheus_homelab_alerts.md](sources/docs/domain/integration-registry/prometheus_homelab_alerts.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/requirements.md](sources/docs/domain/integration-registry/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/task_api.md](sources/docs/domain/integration-registry/task_api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/task_architecture.md](sources/docs/domain/integration-registry/task_architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/task_ingestion_concept.md](sources/docs/domain/integration-registry/task_ingestion_concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/task_requirements.md](sources/docs/domain/integration-registry/task_requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/integration-registry/test_cases.md](sources/docs/domain/integration-registry/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/README.md](sources/docs/domain/onboarding/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/api.md](sources/docs/domain/onboarding/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/architecture.md](sources/docs/domain/onboarding/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/concept.md](sources/docs/domain/onboarding/concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/impl_plan.md](sources/docs/domain/onboarding/impl_plan.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/requirements.md](sources/docs/domain/onboarding/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/onboarding/test_cases.md](sources/docs/domain/onboarding/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/README.md](sources/docs/domain/organization-management/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/api.md](sources/docs/domain/organization-management/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/architecture.md](sources/docs/domain/organization-management/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/backend_requirements_org_hierarchy.md](sources/docs/domain/organization-management/backend_requirements_org_hierarchy.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/org_chart_ux_spec.md](sources/docs/domain/organization-management/org_chart_ux_spec.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/organizational_hierarchy_spec.md](sources/docs/domain/organization-management/organizational_hierarchy_spec.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/requirements.md](sources/docs/domain/organization-management/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/organization-management/test_cases.md](sources/docs/domain/organization-management/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/README.md](sources/docs/domain/platform-lifecycle/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/api.md](sources/docs/domain/platform-lifecycle/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/architecture.md](sources/docs/domain/platform-lifecycle/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/dashboard_concept.md](sources/docs/domain/platform-lifecycle/dashboard_concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/project_concept.md](sources/docs/domain/platform-lifecycle/project_concept.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/requirements.md](sources/docs/domain/platform-lifecycle/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/platform-lifecycle/test_cases.md](sources/docs/domain/platform-lifecycle/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/README.md](sources/docs/domain/rbac-permissions/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/api.md](sources/docs/domain/rbac-permissions/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/architecture.md](sources/docs/domain/rbac-permissions/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/keycloak_groups_mapping.md](sources/docs/domain/rbac-permissions/keycloak_groups_mapping.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/requirements.md](sources/docs/domain/rbac-permissions/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/rbac-permissions/test_cases.md](sources/docs/domain/rbac-permissions/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/realtime/README.md](sources/docs/domain/realtime/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/realtime/api.md](sources/docs/domain/realtime/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/realtime/architecture.md](sources/docs/domain/realtime/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/realtime/requirements.md](sources/docs/domain/realtime/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/repository-integration/README.md](sources/docs/domain/repository-integration/README.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/repository-integration/api.md](sources/docs/domain/repository-integration/api.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/repository-integration/architecture.md](sources/docs/domain/repository-integration/architecture.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/repository-integration/requirements.md](sources/docs/domain/repository-integration/requirements.md) — ingested
- [wiki/projects/devhub/sources/docs/domain/repository-integration/test_cases.md](sources/docs/domain/repository-integration/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/README.md](sources/docs/infrastructure/README.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/commandworker/test_cases.md](sources/docs/infrastructure/commandworker/test_cases.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/deployment-automation/single_port_reverse_proxy.md](sources/docs/infrastructure/deployment-automation/single_port_reverse_proxy.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/e2e_migration.md](sources/docs/infrastructure/keycloak-idp/e2e_migration.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/failover.md](sources/docs/infrastructure/keycloak-idp/failover.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/offboarding_immediacy.md](sources/docs/infrastructure/keycloak-idp/offboarding_immediacy.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/refactor_execution_plan.md](sources/docs/infrastructure/keycloak-idp/refactor_execution_plan.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/service_account_min_role.md](sources/docs/infrastructure/keycloak-idp/service_account_min_role.md) — ingested
- [wiki/projects/devhub/sources/docs/infrastructure/keycloak-idp/sso_federation.md](sources/docs/infrastructure/keycloak-idp/sso_federation.md) — ingested
- [wiki/projects/devhub/sources/docs/validation/2026-06-12-n13-test2-rebase-verification.md](sources/docs/validation/2026-06-12-n13-test2-rebase-verification.md) — ingested
- [wiki/projects/devhub/sources/docs/validation/N-10-manager-rbac.md](sources/docs/validation/N-10-manager-rbac.md) — ingested