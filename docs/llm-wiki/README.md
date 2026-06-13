# DevHub ↔ my_harness `~/wiki/` LLM Wiki 통합 — Phase 1 + Phase 1.5 + Phase 3 (D-72)

- **문서 목적**: DevHub 저장소 (본 repo) 의 LLM Wiki 통합의 **본 저장소 측 SSOT**. my_harness 의 `~/wiki/` Obsidian vault (D-71, D-72 응답) 의 DevHub mirror 의 source-of-truth + lint config + sync script 의 source 위치.
- **범위**: Phase 1 + **Phase 1.5 (2026-06-13 추가)** + **Phase 3 (2026-06-13 추가)** = (1) `docs/llm-wiki/` 의 4 file (scope-and-rationale / mirror-list / lint-config / operation-sop) + (2) `scripts/wiki-sync-devhub.sh` (sync script, **Phase 1.5+3 갱신**) + **`scripts/wiki-mass-ingest.sh`** (Phase 3 NEW) + **`scripts/wiki-frontmatter-update.sh`** (provenance) + **`scripts/wiki-status-check.sh`** (status). **Phase 1.5 의 추가**: source code + workflow + scripts + branch memory + traceability 의 mirror scope. **Phase 3 의 추가**: docs/domain + architecture + infrastructure + validation mass ingest (~78 file) + 78 wiki page 신규 생성. (3) **mirror 실행은 본 PR scope 외** — 사용자가 `bash scripts/wiki-sync-devhub.sh` 실행 시 `~/wiki/raw/projects/devhub/` 에 실제 mirror (~925 file, ~7.6M).
- **대상 독자**: DevHub owner (yklee), LLM agent (**wiki 의 RAG source + 코드 maintenance 작업 시 source**), my_harness 작업 에이전트 (D-72 통합 협업).
- **상태**: active (D-72 Phase 1+1.5+3 + D-79 query + D-80 pr-update thin wrapper 작성 완료, **본 저장소 한정 + 위키만으로 코드 maintenance 가능 정공법 + mass ingest 정공법 + provenance tracking**)
- **최종 수정일**: 2026-06-13 (Phase 1.5+3 추가: mirror scope 8 패턴 ~133 file 확장, lint-config.toml 갱신, operation-sop.md 갱신, mirror 실행 결과 ~925 file / ~7.6M, 162 wiki page matched)
- **관련 문서**:
  - `docs/llm-wiki/scope-and-rationale.md` (**2026-06-13 갱신**: Phase 1+1.5+3 scope + D-72 Q1~Q6 적용)
  - `docs/llm-wiki/mirror-list.md` (**2026-06-13 갱신**: Phase 1+1.5+3 mirror list, 15 패턴 ~220 file)
  - `docs/llm-wiki/lint-config.toml` (**2026-06-13 갱신**: per-project lint config, config_version 3)
  - `docs/llm-wiki/operation-sop.md` (**2026-06-13 갱신**: sync + lint SOP, §0 위키 1:1 mirror 4 layer + Phase 3 mass ingest SOP + §9 향후 작업 지침)
  - `scripts/wiki-sync-devhub.sh` (**2026-06-13 갱신**: sync script, 15 패턴 + --no-clean 옵션)
  - `scripts/wiki-mass-ingest.sh` (**2026-06-13 NEW**: Phase 3 mass ingest, ~78 file wiki page 자동 생성)
  - `scripts/wiki-frontmatter-update.sh` (**2026-06-13**: provenance tracking, 위키 page 의 frontmatter 자동 갱신)
  - `scripts/wiki-status-check.sh` (**2026-06-13**: status 검증 command, 4 mode)
  - my_harness 의 [`scratch/devhub_wiki_integration_response/RESPONSE.md`](../../scratch/devhub_wiki_integration_response/RESPONSE.md) (D-72 권장안, 15KB)
  - my_harness 의 [`docs/architecture/DETAILED_DESIGN_LLM_WIKI.md`](https://github.com/ykylee/my_harness/blob/main/docs/architecture/DETAILED_DESIGN_LLM_WIKI.md) (D-71 디자인, 24KB)

## 1. 두 wiki 의 분리 (audience 차이)

본 저장소 (DevHub) 에는 **두 개의 wiki** 가 공존:

| Wiki | 위치 | Audience | Source | 게시 |
| --- | --- | --- | --- | --- |
| **Public Wiki** (기존) | `docs/wiki/` (in-repo) | 인간 — 외부 개발자, 잠재 사용자, 기술 블로그 독자 | `docs/` 의 대외 편집본 (사전 큐레이션) | GitHub Wiki |
| **LLM Wiki** (D-72 신규) | `~/wiki/raw/projects/devhub/` (out-of-repo) | LLM agent — RAG source, second brain 의 DevHub 부분 | DevHub repo 의 의미 있는 file (mirror) | Gitea private (yklee 개인) |

**두 wiki 의 source-of-truth 가 다름**:
- Public Wiki = `docs/wiki/Home.md`, `docs/wiki/columns/001-role-priority-ux.md` 등 — 인간 작성 + GitHub Wiki 게시 source.
- LLM Wiki = `~/wiki/raw/projects/devhub/docs/adr/0030-sso-integrations-and-auth-session-port.md` (mirror) — DevHub 의 **SSOT 직접 mirror**, 인간 큐레이션 X.

**두 wiki 의 cross-link 정책**: Public Wiki 는 `docs/` 의 SSOT 정합만 신경. LLM Wiki 는 `~/wiki/` 의 SSOT 정합만 신경. **cross-link 없음** (의도적 분리, yklee 결정 2026-06-10).

## 2. 디렉터리 구조 (Phase 1 scope)

### 2.1 DevHub in-repo 변경 (Phase 1 + Phase 3 + D-79 + D-80)

```
docs/llm-wiki/                            ← 본 저장소 측 SSOT (8 file root index)
├── README.md                             ← 본 file (8 file root index)
├── scope-and-rationale.md                ← Phase 1 scope + D-72 Q1~Q6 적용
├── mirror-list.md                        ← Phase 1 mirror list (어떤 file 을 ~/wiki/raw/projects/devhub/ 로 보낼지)
├── lint-config.toml                      ← per-project config (DevHub ADR-*.md L07 면제)
├── operation-sop.md                      ← sync + lint 실행 절차
├── ingest-skill.md                       ← D-72 Phase 3 — wiki-ingest-from-raw skill 본 저장소 측 사용법 가이드
├── query-skill.md                        ← D-79 — wiki-query skill 본 저장소 측 사용법 가이드 (read + --file 모드)
└── pr-update-skill.md                    ← D-80 — wiki-pr-update skill 본 저장소 측 사용법 가이드 (local manual trigger)

scripts/wiki-sync-devhub.sh               ← sync script (D-72 Phase 1, BSD-rsync safe, dry-run + vault-absent no-op)
scripts/wiki-ingest-from-raw.sh           ← wrapper (D-72 Phase 3, 2-step: raw mirror + my_harness skill dispatch)
scripts/wiki-query.sh                     ← wrapper (D-79, 9 option: --query/--tag/--type/--limit/--format/--file/--no-file/--quiet)
scripts/wiki-pr-update.sh                 ← wrapper (D-80, 5 option: --pr/--project/--reingest/--apply/--quiet, gh CLI dispatch)
```

### 2.2 `~/wiki/` out-of-repo 변경 (Phase 1 mirror 실행 후, 본 PR scope 외)

```
~/wiki/
├── AGENTS.md                             ← 단일 (my_harness 의 D-71 SSOT, 두 project 합의본)
├── index.md                              ← 단일 (두 project 카탈로그, 섹션 분리)
├── log.md                                ← 단일 (시계열 로그)
├── raw/
│   ├── projects/
│   │   ├── my-harness/                   ← (my_harness 측 작업, 본 PR scope 외)
│   │   └── devhub/                       ← (본 PR 의 mirror 실행 시 생성, 본 PR scope 외)
│   │       ├── docs/
│   │       │   ├── adr/                  ← (mirror-list.md 의 ADR-* 31 file)
│   │       │   ├── governance/           ← (mirror-list.md 의 governance 5 file)
│   │       │   ├── planning/             ← (26 file)
│   │       │   ├── setup/                ← (15 file)
│   │       │   ├── requirements.md
│   │       │   ├── openapi.yaml
│   │       │   └── ai-workflow-memory/   ← (state.json + session_handoff.md + work_backlog.md)
│   │       └── _manifest.md              ← (mirror 실행 시 자동 생성)
│   └── ... (기존 articles/, books/, personal/, clippings/ — 본 PR scope 외)
├── wiki/
│   ├── projects/
│   │   ├── my-harness/                   ← (my_harness 측 작업, 본 PR scope 외)
│   │   └── devhub/                       ← (Phase 3 mass ingest, 본 PR scope 외)
│   └── cross/                            ← (Phase 3, 본 PR scope 외)
├── schema/                               ← 단일 (my_harness 의 D-71 SSOT, 두 project 합의본)
└── _lint/
    ├── my-harness/                       ← (per-project lint report, my_harness 측)
    └── devhub/                           ← (per-project lint report, 본 PR scope 외 — Phase 3 부터)
```

## 3. Phase 1 의 의의 + 한계

### 3.1 의의

- **in-repo source-of-truth 정합**: 본 저장소 의 `docs/adr/`, `docs/governance/`, `docs/requirements.md` 등이 wiki 의 mirror source 로 자동 적격 — 별도 큐레이션 불요.
- **lint 자동화 정합**: `docs/llm-wiki/lint-config.toml` 의 L07 ADR 면제 config 가 `~/wiki/` 의 wiki-lint 가 자동으로 DevHub 의 의도적 supersede (예: ADR-0030 → ADR-0031) 의 false positive 회피.
- **operation SOP**: `docs/llm-wiki/operation-sop.md` 의 trigger / frequency / dry-run 절차로 owner (yklee) 가 mirror + lint 실행 시 정책 혼선 없음.

### 3.2 한계 (Phase 1 의 scope 외)

- **mass ingest 미실시**: 본 PR 의 mirror list 는 core subset (~80 file) 만. domain (66 file) + architecture + infrastructure + validation 의 100+ file 은 **Phase 3 (mass ingest, 별도 PR)** 에서 mirror.
- **wiki page 작성 미실시**: 본 PR 은 mirror + sync 만. **30~50 wiki page 작성 (concepts/entities/topics/sources)** 은 **Phase 3 (mass ingest)** 에서 진행.
- **`~/wiki/` out-of-repo 변경 미실시**: mirror 실행은 본 PR scope 외. `bash scripts/wiki-sync-devhub.sh` 실행은 사용자 (yklee) 가 직접.
- **`wiki-lint --project` 옵션 미도입**: 본 PR 의 lint-config.toml 은 future 사용 — my_harness 측 작업 에이전트가 wiki-lint skill 에 `--project` + `--project-config` 옵션 추가 후 본 lint-config 가 활성화. (D-72 응답 §4 #4)
- **wiki-lint CI integration 미실시**: 본 PR 의 lint-config.toml 의 lint 는 **사용자 수동 실행** (Phase 1). **CI integration 은 D-74+ (별도 PR)** 에서.

## 4. 다음 행동 (Phase 1 → Phase 3 → D-79 → D-80)

| ID | 우선순위 | 작업 | 의존 |
| --- | --- | --- | --- |
| T-d-72-1 | P3 | Phase 1 in-repo source-of-truth 정합 (PR #544 머지) | — |
| T-d-72-2 | P3 | `bash scripts/wiki-sync-devhub.sh` 1회 실행 (실제 mirror, dry-run 아닌 real) | T-d-72-1 |
| T-d-72-3 | P3 | `~/wiki/` 의 `wiki-lint` skill 에 `--project` + `--project-config` 옵션 추가 (my_harness 측 D-73) | T-d-72-2 |
| T-d-72-4 | P3 | `wiki/projects/devhub/` 의 wiki page 작성 (30~50 page, **Phase 3 mass ingest**) | T-d-72-3 |
| T-d-72-5 | P3 | `wiki/cross/` 의 cross-project 종합 페이지 작성 (1~3 page) | T-d-72-4 |
| T-d-72-6 | P3 | wiki-lint CI integration (`ci.yml` 의 e2e shard 또는 별도 lint job) | T-d-72-3 |
| **T-d-79-1** | P3 | **본 저장소 측 thin wrapper 작성** (`scripts/wiki-query.sh` + `docs/llm-wiki/query-skill.md`) | — |
| **T-d-79-2** | P3 | my_harness 측 `wiki_query_skill_spec.md` (§1-§11) + `skills/wiki-query/SKILL.md` + `scripts/run_wiki_query.py` 작성 (my_harness 측) | — |
| **T-d-79-3** | P3 | dry-run 검증 (5 query sample: full-text / wikilink / tag / type / json format) | T-d-79-2 |
| **T-d-79-4** | P3 | `--file` 옵션 1회 실행 검증 (query/ 페이지 + log.md + index.md side effect, AGENTS.md §2.2 6 step 자동) | T-d-79-3 |
| **T-d-79-5** | P3 | wiki-lint 통합 (`wiki-lint` skill 의 L01~L10 가 query/ 페이지 검증) | D-74 |
| **T-d-79-6** | P3 | v2.0 (BM25 + vector + RRF) — query 결과의 RAG rerank (my_harness 측) | my_harness v2.0 |
| **T-d-80-1** | P3 | **본 저장소 측 thin wrapper 작성** (`scripts/wiki-pr-update.sh` + `docs/llm-wiki/pr-update-skill.md`) | — |
| **T-d-80-2** | P3 | my_harness 측 `wiki_pr_update_skill_spec.md` (§1-§11) + `skills/wiki-pr-update/SKILL.md` + `scripts/run_wiki_pr_update.py` 작성 (my_harness 측) | — |
| **T-d-80-3** | P3 | dry-run 검증 (PR #552 — 실제 머지된 PR, 6 touched file) | T-d-80-2 |
| **T-d-80-4** | P3 | `--apply` 1회 실행 (prs/552.md + log.md 1 line + index.md prs 섹션) | T-d-80-3 |
| **T-d-80-5** | P3 | idempotency 검증 (동일 PR 재실행 → skip 확인, `pr-<num>-<head.sha>` key) | T-d-80-4 |
| **T-d-80-6** | P3 | `--reingest` 검증 (PR touched file 중 mirror-list 매칭 시 `wiki-ingest-from-raw --source <file> --apply` re-run) | T-d-80-4 |
| **T-d-80-7** | P3 | wiki-lint 통합 (`wiki-lint` skill 의 L01~L10 가 prs/<num>.md 검증) | D-74 |
| **T-d-80-8** | P3 | CI integration (사내 self-hosted runner 도입 시 forward path — `pull_request: closed` + `workflow_dispatch` trigger) | my_harness v2.0 |
| **T-d-80-9** | P3 | main branch 머지 후 `prs/<num>.md` 의 `mergedAt` + `mergeCommitSha` 자동 fill | T-d-80-4 |

## 5. Tier 정책 정합

**DevHub 의 3-tier 정책 (= GitHub vs 사내 SCM push 분리) 은 DevHub repo 자체에만 적용**. `~/wiki/` 는 별도 Gitea private repo (my_harness 의 D-72 응답 §3 결정) — 다른 정책.

따라서 본 PR (Phase 1) + Phase 3 + D-79 + D-80 의 모든 변경 = **공용**:
- `docs/llm-wiki/` 의 8 file — 대외 정합 가능 (D-72 응답 + lint/sync SOP + D-79 query 가이드 + D-80 pr-update 가이드).
- `scripts/wiki-sync-devhub.sh` + `scripts/wiki-ingest-from-raw.sh` + `scripts/wiki-query.sh` + `scripts/wiki-pr-update.sh` — 4 script 모두 sync/ingest/query/PR-update 의 thin wrapper, GitHub 공개 적격 (실제 vault 갱신 은 Gitea private 만).
- 단 `mirror-list.md` 의 mirror source list 에는 DevHub 의 사내 한정 정보 (`DEVHUB_KEYCLOAK_*`, `internal-registry.*`, `kc.internal.example.com`, `devhub.example.com`, `172.16.0.0/12`, `infra/idp/_archive_*/`) 가 포함되지 않음 — D-72 응답 §3 + yklee 결정 (wiki 는 Gitea private 만) 으로 `sa-internal/` 격리 불요, 사외/공용 정보만 mirror.

**Tier 검증 (예상)**: `bash scripts/check-tier-separation.sh` PASS — 본 PR 의 모든 변경 = **공용** (사내 한정 정보 미포함).

## 6. 다음 세션 directive

1. `git status` / `git log --oneline -3` / `git branch --show-current` 확인.
2. 본 PR 의 8 file + 4 script 의 lint / commit / push / PR 발행 (D-79 + D-80 추가분).
3. main flat memory 3 file finalize (post-merge sync).
4. 또는 T-d-72-2 (`bash scripts/wiki-sync-devhub.sh` 1회 실행, 사용자 confirm 후).
5. 또는 T-d-79-2 / T-d-80-2 (my_harness 측 SSOT spec/impl 작성 의뢰, [`./handoff-to-my-harness.md`](../../ai-workflow/memory/feat/work_260611-a-wiki-ingest-from-raw/handoff-to-my-harness.md) 참고).
6. 또는 다른 sprint (backend-integration matrix / N-10 RBAC E2E 6 TC / release_v1_roadmap §3.5 N-13 정합).
