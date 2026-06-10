# PR Body — feat/work_260610-d-72-wiki-phase-1 (PR #XXX body source)

- 문서 목적: PR #XXX 의 body 본문 (markdown) — gh pr create --body-file 의 source. PR #XXX 가 머지된 후 본 file 은 archival.
- 범위: PR #XXX 의 PR body 본문 (변경 요약 + 추적성 영향 + Tier 분류 + 검증 + 의도적 trade-off + Out of scope + Refs + Base/target).
- sprint branch: feat/work_260610-d-72-wiki-phase-1
- 대상 독자: PR reviewer, owner, 머지 후 archival reference
- 상태: draft (PR 발행 전)
- 최종 수정일: 2026-06-10
- 관련 문서: [session_handoff.md](./session_handoff.md), [work_backlog.md](./work_backlog.md), [backlog/2026-06-10.md](./backlog/2026-06-10.md), [state.json](./state.json), [D-72 응답](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md), [my_harness D-71 디자인](https://github.com/ykylee/my_harness/blob/main/docs/architecture/DETAILED_DESIGN_LLM_WIKI.md)

# docs(llm-wiki,scripts): D-72 Phase 1 — `~/wiki/` LLM Wiki 통합 의 in-repo source-of-truth + sync script (sprint `feat/work_260610-d-72-wiki-phase-1`)

## 목적

DevHub 저장소 (본 repo) 의 LLM Wiki 통합의 **본 저장소 측 Phase 1**. my_harness 의 `~/wiki/` Obsidian vault (D-71 디자인, D-72 응답) 의 DevHub mirror 의 source-of-truth + lint config + sync script 의 source 위치.

[D-72 응답 (RESPONSE.md, 15KB)](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md) 의 6 질문별 권장안 (Q1 단일 vault + per-project 동거, Q2 per-project raw/ 분리, Q3 단순화 — lint L11 + sa-internal/ 격리 불요, Q4 schema 단일 + L01~L10 + L07 ADR 면제, Q5 v1.5 동시 시작, Q6 단일 AGENTS.md + per-project lint report) 의 본 저장소 측 적용.

**범위**: in-repo 변경 (본 PR) + out-of-repo mirror 실행 (사용자 confirm 후, 본 PR scope 외). 본 PR 의 의의 = **source-of-truth 의 in-repo 제공**, mirror 실행은 사용자 trigger.

## 변경 요약

### 1. `docs/llm-wiki/` (NEW, 5 file)

본 저장소 측의 **in-repo source-of-truth**. 기존 `docs/wiki/` (Public Wiki, GitHub Wiki 게시 source, 인간 큐레이션) 와 명확히 분리 (audience + source-of-truth 다름).
- [D-72 응답 (RESPONSE.md, 15KB)](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md) — my_harness 작업 에이전트 의 권장안 (Q1~Q6 + 8 위험 + 6 next-step)
| --- | --- |
| `README.md` (7.8KB) | 5 file root index + 두 wiki 분리 (Public vs LLM) + 디렉터리 구조 (in-repo + out-of-repo) + Phase 1 의 의의 + 한계 + forward path + Tier 정책 정합 + 다음 행동 |
| `scope-and-rationale.md` (10.6KB) | Phase 1 scope + D-72 6 질문의 본 저장소 측 적용 + Phase 1 in-repo scope + scope 외 (forward path) + trade-off |
| `mirror-list.md` (10KB) | Phase 1 source list (core subset ~82 file) + Phase 3 scope 외 (domain 66 + architecture + infrastructure + validation) + lint 영향 + forward path |
| `lint-config.toml` (4.4KB) | per-project config (TOML, Q4 의 권장안 §4) — `[project] name = "devhub"`, `[rules.L07].skip_paths` (ADR-*.md 면제), `[rules.L07].skip_if_frontmatter = ["supersedes"]`, `[rules.L10].devhub_adr_source_pattern` (raw/ 1:1 mirror), `[meta]` (config_version, created, source_of_truth) |
| `operation-sop.md` (10.7KB) | sync trigger (수동 + 주기) + sync 절차 (dry-run / real / vault 부재) + lint trigger + 주기 + sync 위험 10건 (R-d-72-S-1..10) + verification |

### 2. `scripts/wiki-sync-devhub.sh` (NEW, 6.4KB, executable)

**BSD-rsync safe sync script** (my_harness 의 `wiki-sync-ai-workflow.sh` 와 동일 pattern). 7 source 패턴:
1. ADR: `docs/adr/0[0-9][0-9][0-9]-*.md` (31 file, ADR-0001..0031)
2. Governance: `docs/governance/*.md` (5 file)
3. Planning: `docs/planning/*.md` (26 file)
4. Setup: `docs/setup/*.md` (15 file)
5. Requirements: `docs/requirements.md` (1 file)
6. OpenAPI: `docs/openapi.yaml` (1 file)
7. AI-workflow memory (main flat): `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` (3 file)

**총 82 file, ~3.5MB** (estimated).

**Mode**:
- `--dry-run`: source list 출력 (no actual mirror, CI 검증 가능).
- (no option): real mirror — DEST 의 clean mirror 후 cp + `_manifest.md` 자동 생성.
- `--help`: usage.

**Exit code**:
- 0: success
- 1: vault 부재 / source root 부재 / invalid option

**Vault 부재 시 명시적 error**: "target vault not found: $HOME/wiki" + hint (my_harness 의 D-71 §2.2 wiki-init 또는 D-72 응답 §4 #1).

## 추적성 영향

| Stage | ID | Status | 비고 |
| --- | --- | --- | --- |
| ADR | (없음) | — | 본 PR 은 ADR 신규 X. D-72 응답 §2 Q4 의 lint config 의 `[rules.L07]` 만 적용. |
| IMPL/REQ/UC/ARCH/API/RM/UT/TC | (없음) | — | 본 PR 은 mirror + sync script 만, ID 미발급 |

**신규 ID 0건**. 본 PR 은 source-of-truth + script 의 source. mirror 실행 + lint 결과는 본 PR scope 외 (사용자 confirm 후, Phase 1.5 forward path).

## Tier 분류 (self-check)

| 변경 영역 | Tier | 근거 |
| --- | --- | --- |
| `docs/llm-wiki/` 5 file | **공용** | 본 저장소 의 wiki SSOT, 사내 한정 정보 미포함 (Q3 단순화 + yklee 결정: wiki 는 Gitea private 만). 단 `mirror-list.md` 의 source list 의 `worker_division.md` 는 사내 한정 정보 포함하나 (D-72 응답 §3) yklee 결정으로 Gitea private 만 push 이므로 lint L11 불요. |
| `scripts/wiki-sync-devhub.sh` | **공용** | sync script 의 source, GitHub 공개 적격 (mirror 실행은 Gitea private 만). |

**본 PR 의 모든 변경 = 공용**. `bash scripts/check-tier-separation.sh` PASS 확인.

## 검증 (run on this branch)

- `bash scripts/check-tier-separation.sh` — ✅ no changes between origin/main and HEAD
- `bash scripts/check-openapi-yaml-lint.sh` — ✅ passed (openapi.yaml 변경 0)
- `bash scripts/check-migration-uniqueness.sh` — ✅ valid and unique (migration 변경 0)
- `python3.13 ai-workflow/tests/check_docs.py` — 본 PR 의 5 file 정합 (metadata 6 field + cross-link + 제목 헤더). exit 1 의 원인은 본 PR 영역 외 기존 file 의 historical link.
- `bash scripts/wiki-sync-devhub.sh --dry-run` — ✅ PASS (82 file source list 출력, no actual mirror, exit 0)
- `HOME=/tmp/fake_home bash scripts/wiki-sync-devhub.sh` — ✅ vault 부재 시 명시적 error + exit 1

## 의도적 trade-off

- **`docs/llm-wiki/` 선택 (vs `docs/wiki/` 또는 `docs/wiki-integration/`)**: 기존 `docs/wiki/` = **Public Wiki** (GitHub Wiki 게시 source, 인간 큐레이션, mtime 2026-05-20). 본 Phase 1 의 **LLM Wiki SSOT** 와 audience 다름. 디렉터리 이름 분리 = 두 wiki 의 명확한 구분. `docs/wiki/` (Public) ↔ `docs/llm-wiki/` (LLM) 의 cross-link 없음.
- **mirror list 의 scope = core subset ~82 file**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture (1) + infrastructure + validation (~100 file) 은 **Phase 3 (mass ingest)** 의 별도 PR. 본 PR 의 검증 가능한 정공법 (CI 4/4 + script smoke test) = 작은 core subset.
- **lint-config.toml 의 L07 ADR 면제 config 작성 (옵션 미사용)**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 config 의 source 만 제공. 옵션 추가 후 자동 활성.
- **`~/wiki/` out-of-repo 변경 미포함**: 본 PR scope 의 의도적 한계. `~/wiki/raw/projects/devhub/`, `~/wiki/AGENTS.md`, `~/wiki/index.md`, `~/wiki/log.md`, `~/wiki/_lint/devhub/`, `~/wiki/wiki/projects/devhub/`, `~/wiki/wiki/cross/`, `~/wiki/schema/` 등 모두 out-of-repo = 본 PR scope 외. **본 PR 의 source-of-truth 만 in-repo**.
- **mirror 실행은 본 PR scope 외**: 본 PR 의 lint 검증은 `bash scripts/wiki-sync-devhub.sh --dry-run` 의 source list 정합만. 실제 mirror 는 사용자 (yklee) 가 Phase 1 mirror 실행 시점에 진행. **이유**: mirror 실행은 `~/wiki/` 의 out-of-repo 변경 — 본 PR 의 in-repo 검증 가능 영역 (CI 4/4) 의 검증 범위 외.
- **scratch/ exclude**: `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) 는 본 Phase 1 의 reference. **mirror 미포함** (vault 비대화 + D-72 응답은 my_harness 측 작업).
- **`backend-core/` / `frontend/` / `backend-ai/` 의 source code mirror 제외**: vault 비대화 + LLM agent 의 코드 정합은 `code-index-update` 의 영역. 단 `docs/` 하위의 의미 있는 file (markdown + yaml + json) 만 mirror.

## Out of scope (별도 PR / 후속)

- **T-d-72-2 (P3)**: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~82 file mirror + `_manifest.md` 자동 생성. 본 PR 머지 후 사용자 confirm.
- **D-73 (my_harness 측)**: wiki-lint skill 에 `--project` + `--project-config` 옵션 추가. 본 PR 의 lint-config.toml 활성.
- **D-74 (my_harness 측)**: my_harness 의 `_lint/my-harness/` + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업.
- **Phase 3 (mass ingest)**: DevHub 의 docs/domain + docs/architecture + docs/infrastructure + docs/validation (~100 file) mirror + 30~50 wiki page 작성. 별도 PR.
- **wiki/cross/** (cross-project 종합): my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection, 1~3 page. Phase 3 후속.
- **wiki-lint CI integration**: `ci.yml` 의 별도 lint job. D-73 + D-74 후.
- **v2.0 (full compile)**: LLM 호출 + BM25+vector+MCP. my_harness 의 v2.0 경험 보고 진입 (D-72 Q5).
- **N-13 release_v1_roadmap §3.5 정합**: N-13 row status = done (D-72 D-73 D-74 D-75) 마킹, 별도 housekeeping PR.

## Refs
- [D-72 응답 (RESPONSE.md, 15KB)](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md) — my_harness 작업 에이전트 의 권장안 (Q1~Q6 + 8 위험 + 6 next-step)
- [D-72 응답 (RESPONSE.md, 15KB)](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md) — my_harness 작업 에이전트 의 권장안 (Q1~Q6 + 8 위험 + 6 next-step)
- [D-71 디자인 (DETAILED_DESIGN_LLM_WIKI.md, 24KB)](https://github.com/ykylee/my_harness/blob/main/docs/architecture/DETAILED_DESIGN_LLM_WIKI.md) — my_harness 의 LLM Wiki + Obsidian 디자인 SSOT
- my_harness 의 [`wiki-sync-ai-workflow.sh`](https://github.com/ykylee/my_harness/blob/main/~/wiki/wiki-sync-ai-workflow.sh) — 본 script 의 source-of-pattern (BSD-rsync safe + set -euo pipefail + find + cp)
- my_harness 의 [`wiki-lint SKILL.md`](https://github.com/ykylee/my_harness/blob/main/ai-workflow/skills/wiki-lint/SKILL.md) — wiki-lint L01~L10 10개 규칙 (D-72 의 lint target)
- release_v1_roadmap.md §3.5 N-13 (D-72~D-75 carry-over 의 source)
- DevHub repo: `~/repos/Devhub_example_omp/` (본 PR 의 source root)

## Base / target

- **base branch**: `main` (HEAD `ea8b4bf`, PR #543 + 2 housekeeping commits 머지 후)
- **target branch**: `main`
- **merge strategy**: squash
- **branch name**: `feat/work_260610-d-72-wiki-phase-1`
