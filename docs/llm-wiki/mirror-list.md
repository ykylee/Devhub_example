# Phase 1 Mirror List (D-72, core subset ~80 file)

- **문서 목적**: `scripts/wiki-sync-devhub.sh` 의 mirror source list. Phase 1 의 scope = **core subset ~80 file**. DevHub repo 의 의미 있는 file 중 wiki 의 LLM agent RAG source 로 적합한 file 의 명시.
- **범위**: 본 Phase 1 의 mirror list scope. domain (66 file) + architecture (1) + infrastructure + validation (~100 file) 의 **mirror 제외** (Phase 3 mass ingest 의 scope). mirror 제외 사유 = 본 Phase 1 의 검증 가능한 정공법 (작은 core subset) + mirror 실행 시간.
- **대상 독자**: yklee (owner), LLM agent (Phase 3 mass ingest 시 wiki page 작성자), `scripts/wiki-sync-devhub.sh` 의 자동 mirror 실행 시 source 결정.
- **상태**: in_progress (D-72 Phase 1, 2026-06-10)
- **최종 수정일**: 2026-06-10
- **관련 문서**:
  - [`./scope-and-rationale.md`](./scope-and-rationale.md) (Phase 1 scope + D-72 Q1~Q6 적용)
  - [`./operation-sop.md`](./operation-sop.md) (sync + lint SOP)
  - `scripts/wiki-sync-devhub.sh` (sync script 의 source)

## 1. Phase 1 source list (core subset ~80 file)

### 1.1 Architecture Decision Records (ADR) — 31 file

**소스 경로**: `docs/adr/*.md` (ADR-0001..ADR-0031)

| ADR ID | 제목 | mirror target |
| --- | --- | --- |
| ADR-0001 | idp-selection | `~/wiki/raw/projects/devhub/docs/adr/0001-idp-selection.md` |
| ADR-0002 | rbac-policy-edit-api | `~/wiki/raw/projects/devhub/docs/adr/0002-rbac-policy-edit-api.md` |
| ... | ... | ... |
| ADR-0030 | sso-integrations-and-auth-session-port | `~/wiki/raw/projects/devhub/docs/adr/0030-sso-integrations-and-auth-session-port.md` |
| ADR-0031 | build-tag-policy-review | `~/wiki/raw/projects/devhub/docs/adr/0031-build-tag-policy-review.md` |

**mirror 정책**: 31 file 모두 mirror (의미 있는 SSOT 결정). `infra/idp/_archive_*/` 의 immutable archive 미포함 (sprint -a follow-up PR #540 의 immutable archive 결정 정합).

**lint 영향**: ADR-*.md 는 L07 (모순) 의 의도적 supersede (예: ADR-0030 → ADR-0031) 가 false positive 가능 → `lint-config.toml` 의 `[rules.L07].skip_paths = ["wiki/projects/devhub/sources/ADR-*.md"]` 면제 (Phase 1 의 lint config).

### 1.2 Governance — 5 file

**소스 경로**: `docs/governance/*.md`

| file | mirror target |
| --- | --- |
| `code-taxonomy.md` | `~/wiki/raw/projects/devhub/docs/governance/code-taxonomy.md` |
| `document-standards.md` | `~/wiki/raw/projects/devhub/docs/governance/document-standards.md` |
| `keycloak_admin_responsibility.md` | `~/wiki/raw/projects/devhub/docs/governance/keycloak_admin_responsibility.md` |
| `README.md` | `~/wiki/raw/projects/devhub/docs/governance/README.md` |
| `worker_division.md` | `~/wiki/raw/projects/devhub/docs/governance/worker_division.md` |

**mirror 정책**: 5 file 모두 mirror. 단 `worker_division.md` 의 사내 한정 정보 (DEVHUB_KEYCLOAK_*, internal-registry, 172.16.0.0/12) **포함되지만** D-72 응답 §3 + yklee 결정으로 `sa-internal/` 격리 불요 + GitHub Wiki 가 아닌 Gitea private 만 push 이므로 mirror 허용.

### 1.3 Planning — 26 file

**소스 경로**: `docs/planning/*.md`

**mirror 정책**: 26 file 모두 mirror. 단 `infra/idp/_archive_*/` 의 immutable archive 미포함 (ADR-0001/0009 cross-ref 가능 archive 는 별도 위치).

**file list (자동 script 로 동적)**:
- 2026-06-12-inbound-source-routing-sprint-plan.md
- api-key-management-sprint-plan.md
- application_management_hotfix_2026-05-27.md
- external-integrations-agentic-rag-roadmap.md
- integrated_test_report_20260601.md
- integrated_test_scenarios.md
- keycloak_event_audit_integration.md
- migration_baseline_reset_plan_2026-06-04.md
- ops_ui_transition_plan.md
- project_creation_dreq_notification_concept.md
- release_v1_roadmap.md
- role-access-concept.md
- sprint-plan-20260601.md
- system_usecases.md
- rbac-hardening-implementation-readiness-20260602.md
- (11+ file; `find docs/planning -type f -name "*.md" | wc -l` = 26)

### 1.4 Setup — 15 file

**소스 경로**: `docs/setup/*.md`

**mirror 정책**: 15 file 모두 mirror. setup 의 운영 SOP (test-server-deployment, single-port-deployment, docker-packaging-deployment-guide 등) 가 LLM agent 의 RAG source 로 가치 높음.

### 1.5 Requirements + OpenAPI — 2 file

**소스 경로**:
- `docs/requirements.md` (DevHub 의 REQ SSOT)
- `docs/openapi.yaml` (DevHub 의 API contract)

**mirror 정책**: 2 file 모두 mirror. `openapi.yaml` 의 경로 = 81, schema = 78. **단 `openapi.yaml` 의 경로 (e.g. `internal-registry.example.com`) 가 사내 한정 정보 포함 가능** — D-72 응답 §3 의 lint L11 (사내 패턴 검출) 으로 자동 검출 권장 (D-73 작업, my_harness 측).

### 1.6 AI-workflow memory (main flat) — 3 file

**소스 경로**:
- `ai-workflow/memory/state.json` (head_commit + status)
- `ai-workflow/memory/session_handoff.md` (post-session handoff)
- `ai-workflow/memory/work_backlog.md` (변경 이력)

**mirror 정책**: 3 file 모두 mirror (main flat 만 — sprint branch 의 memory 는 본 Phase 1 의 scope 외, my_harness 의 wiki-sync-ai-workflow.sh 와 동일 pattern).

**본 저장소 의 main flat memory 의 위치**:
- `state.json` 의 `head_commit` = `ea8b4bf` (2026-06-10, PR #543 + 2 housekeeping commits)
- `session_handoff.md` 의 §6/§7/§8 (PR #541/#542/#543 row)
- `work_backlog.md` 의 §5 변경 이력 (PR #543 row)

## 2. Phase 1 scope 외 (Phase 3 mass ingest, 별도 PR)

### 2.1 Domain (66 file)

**소스 경로**: `docs/domain/**/*.md`

| sub-directory | file 수 | 비고 |
| --- | --- | --- |
| `docs/domain/auth-session/` | ~10 | 인증/계정 도메인 (RELEASE Sprint, N-12 등) |
| `docs/domain/rbac-permissions/` | ~8 | RBAC 도메인 (ARCH-04..05, REQ-FR-87..) |
| `docs/domain/organization-management/` | ~6 | 조직 도메인 (REQ-FR-39..) |
| `docs/domain/application-lifecycle/` | ~8 | Application 도메인 |
| `docs/domain/dev-request/` | ~5 | dev-request 도메인 (ADR-0028) |
| `docs/domain/onboarding/` | ~4 | onboarding 도메인 |
| `docs/domain/audit-ops/` | ~5 | audit-ops 도메인 (ADR-0030 정합) |
| `docs/domain/realtime/` | ~3 | realtime 도메인 |
| `docs/domain/auth-session-port/`, `auth-rbac/`, `auth-rbac-port/` 등 | ~17 | cross-cut 도메인 |

**Phase 3 의 mirror list** — `find docs/domain -type f -name "*.md"` 의 66 file 모두.

### 2.2 Architecture (1 file)

**소스 경로**: `docs/architecture/DETAILED_DESIGN_*.md` (총 7+ file)

**Phase 3 의 mirror list** — `find docs/architecture -type f -name "*.md"` 의 1 file.

### 2.3 Infrastructure (variable)

**소스 경로**: `docs/infrastructure/**` (sub: keycloak-idp/, integration/, monitoring/, ops/, etc.)

**Phase 3 의 mirror list** — `find docs/infrastructure -type f -name "*.md"` 의 variable file (현시점 count = `infrastructure/README.md` 만, sub-directory 별도 확인 필요).

### 2.4 Validation (1 file)

**소스 경로**: `docs/validation/N-*.md`

**Phase 3 의 mirror list** — N-12 (voc + notification) 등.

## 3. mirror 실행 정책 (script 의 source list)

`scripts/wiki-sync-devhub.sh` 의 mirror 실행 시 다음 4 패턴으로 file 매칭:

| 패턴 | mirror source | mirror target |
| --- | --- | --- |
| ADR | `docs/adr/ADR-*.md` | `~/wiki/raw/projects/devhub/docs/adr/ADR-*.md` |
| Governance | `docs/governance/*.md` | `~/wiki/raw/projects/devhub/docs/governance/*.md` |
| Planning | `docs/planning/*.md` | `~/wiki/raw/projects/devhub/docs/planning/*.md` |
| Setup | `docs/setup/*.md` | `~/wiki/raw/projects/devhub/docs/setup/*.md` |
| Requirements | `docs/requirements.md` | `~/wiki/raw/projects/devhub/docs/requirements.md` |
| OpenAPI | `docs/openapi.yaml` | `~/wiki/raw/projects/devhub/docs/openapi.yaml` |
| AI-workflow memory (main flat) | `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` | `~/wiki/raw/projects/devhub/ai-workflow-memory/{state.json, session_handoff.md, work_backlog.md}` |

**제외 패턴** (mirror 미실시):
- 빌드 산출물: `target/`, `backend-core/main`, `frontend/.next/`, `playwright-report/`, `test-results/`, `dist/`, `build/`, `__pycache__/`, `node_modules/`
- VCS + IDE: `.git/`, `.idea/`, `.vscode/`, `.DS_Store`
- Backend runtime: `backend-core/`, `frontend/`, `backend-ai/` 의 source code (wiki 의 RAG source 가 아님, 코드 정합은 `ai-workflow/memory/{code-index, ...}.md` 등 별도)
- Archive: `infra/idp/_archive_*/` (immutable archive, wiki 정합 불요)
- Public Wiki (기존): `docs/wiki/` (대외 공개용, LLM Wiki 와 cross-link 없음)
- LLM Wiki (본 Phase 1): `docs/llm-wiki/` (본 Phase 1 의 source-of-truth, mirror 미필요)
- Lint/scratch: `_lint/`, `scratch/`, `playwright-report/`

**mirror size 추정**:
- ADR: 31 file (≈ 700KB)
- Governance: 5 file (≈ 100KB)
- Planning: 26 file (≈ 1.5MB)
- Setup: 15 file (≈ 800KB)
- Requirements: 1 file (≈ 50KB)
- OpenAPI: 1 file (≈ 300KB)
- AI-workflow memory: 3 file (≈ 50KB)
- **합 ≈ 3.5MB, 82 file**

## 4. lint 영향

**Phase 1 mirror 실행 후 wiki-lint 의 L01~L10 검증** (D-73 wiki-lint `--project` 옵션 활성 후):

| L rule | 영향 | DevHub 적용 |
| --- | --- | --- |
| L01 | frontmatter 누락 | raw/ file 은 wiki page 아님, 적용 X |
| L02 | broken wiki link | wiki page 만, raw/ 적용 X |
| L03 | 고아 페이지 | wiki page 만, raw/ 적용 X |
| L04 | 중복 페이지 | wiki page 만, raw/ 적용 X |
| L05 | stale (90일+) | wiki page 만, raw/ 적용 X |
| L06 | sources: 경로 부재 | wiki page 만, raw/ 적용 X |
| L07 | 모순 (같은 title 두 페이지) | **DevHub ADR-*.md 면제** (lint-config.toml) |
| L08 | index.md 미등록 wiki 페이지 | Phase 3 의 wiki page 작성 후 해소 |
| L09 | log.md 1주일+ 미갱신 | out-of-repo, my_harness 측 관리 |
| L10 | raw/ source 0 | wiki page 의 raw mirror source 부재, Phase 3 의 wiki page 작성 후 해소 |

**Phase 1 의 lint 검증 = 본 PR scope 외 (mirror 실행 후 사용자 confirm 후)**.

## 5. forward path

| 단계 | mirror list | 정공법 |
| --- | --- | --- |
| **Phase 1 (본 PR)** | core subset ~80 file (ADR/governance/planning/setup/requirements/openapi/ai-workflow memory) | `docs/llm-wiki/mirror-list.md` + `scripts/wiki-sync-devhub.sh` |
| **Phase 3 (별도 PR)** | domain (66) + architecture (1) + infrastructure + validation (~100 file) | `docs/llm-wiki/mirror-list-phase-3.md` (별도 작성) + `scripts/wiki-sync-devhub.sh` 의 `--phase 3` 옵션 추가 (선택) |
| **Phase N (forward)** | ai-workflow memory 의 sprint branch 별 mirror (본 Phase 1 의 main flat 외) | `scripts/wiki-sync-devhub.sh` 의 `--branch <branch>` 옵션 (선택) |

## 6. 다음 세션 directive

1. `docs/llm-wiki/lint-config.toml` 작성 (L07 ADR 면제 + lint L11/L05/L09 의 DevHub 적용 정책).
2. `docs/llm-wiki/operation-sop.md` 작성 (sync + lint SOP + forward path + 위험).
3. `scripts/wiki-sync-devhub.sh` 작성 (BSD-rsync safe, dry-run + vault-absent no-op + mirror list 의 source list 동적 + manifest 자동).
4. sprint memory 5 file + main flat memory 3 file.
5. lint 검증 (4종) + script smoke test (dry-run mode).
6. commit + push + PR 발행.
7. main flat memory finalize (post-merge sync).
