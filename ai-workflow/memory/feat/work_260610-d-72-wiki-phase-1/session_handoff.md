# Session Handoff — feat/work_260610-d-72-wiki-phase-1

- 문서 목적: 본 sprint 의 작업 상태 + 핵심 사실 (D-72 6 질문 본 저장소 측 적용 + docs/llm-wiki/ 5 file + script 82 file source list) + 다음 세션 directive.
- 범위: 본 PR (`feat/work_260610-d-72-wiki-phase-1`) 의 sprint scope = (1) `docs/llm-wiki/{README.md, scope-and-rationale.md, mirror-list.md, lint-config.toml, operation-sop.md}` 5 file 신규. (2) `scripts/wiki-sync-devhub.sh` 1 file 신규. (3) `ai-workflow/memory/feat/work_260610-d-72-wiki-phase-1/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-10.md, pr_body.md}` 5 file 신규. (4) `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` main flat sync (post-merge). (5) **코드 0줄** (sprint -a follow-up PR1 PR #540 의 carry-over C-j 의 후속 정공법 — LLM Wiki 통합 D-72 Phase 1).
- sprint branch: feat/work_260610-d-72-wiki-phase-1
- 대상 독자: 후속 에이전트, PR reviewer, owner
- 상태: in_progress (사용자 confirm 후 commit + push + PR 발행 대기)
- 최종 수정일: 2026-06-10
- 관련 문서: [work_backlog.md](./work_backlog.md), [backlog/2026-06-10.md](./backlog/2026-06-10.md), [state.json](./state.json), [pr_body.md](./pr_body.md), [D-72 응답](../../../../scratch/devhub_wiki_integration_response/RESPONSE.md), [my_harness D-71 디자인](https://github.com/ykylee/my_harness/blob/main/docs/architecture/DETAILED_DESIGN_LLM_WIKI.md)

## 1. sprint 목표 (in_progress — commit + PR 발행 대기)

D-72 Phase 1 — `~/wiki/` LLM Wiki 통합 의 **in-repo source-of-truth** + sync script. **코드 0줄 변경** (스크립트 6.4KB + docs 5 file 신규).

sprint scope (사용자 결정 2026-06-10):
- **(Phase 1) 본 PR**: docs/llm-wiki/ 5 file + scripts/wiki-sync-devhub.sh + sprint memory 5 file + main flat memory 3 file
- **(OUT of scope)**: T-d-72-2 (`bash scripts/wiki-sync-devhub.sh` 1회 실행, real mirror, 사용자 confirm 후) / D-73 (my_harness 측 wiki-lint 옵션 추가) / D-74 (my_harness 측 _lint/ 셋업) / Phase 3 (mass ingest, ~100 file mirror + 30~50 wiki page) / wiki/cross/ (cross-project 종합) / v2.0 (full compile, LLM 호출 + BM25+vector+MCP) / N-13 release_v1_roadmap §3.5 정합

## 2. 사용자 결정 사항 (in-session)

- **D-72 권장안 전체 승인** (Q1~Q6): 단일 `~/wiki/` vault + `wiki/projects/{devhub,my-harness}/` 동거 + per-project raw/ 분리 + Q3 단순화 (lint L11 + sa-internal/ 격리 불요, wiki 는 Gitea private 만) + Q4 단순화 (L01~L10 + L07 ADR 면제, lint L11 미사용) + Q5 v1.5 동시 시작 (my_harness 선 → DevHub 후) + Q6 단일 AGENTS.md / index.md / log.md + per-project `_lint/<project>/` 분리.
- **Phase 1 scope = core subset ~82 file** (ADR/governance/planning/setup/requirements/openapi/ai-workflow memory main flat). domain (66) + architecture + infrastructure + validation (~100 file) 은 Phase 3 (mass ingest) 의 별도 PR.
- **`docs/llm-wiki/` 선택** (vs `docs/wiki/` 또는 `docs/wiki-integration/`): 기존 `docs/wiki/` (Public Wiki, GitHub Wiki 게시 source) 와 명확히 분리. 두 wiki 의 audience + source-of-truth 다름.

## 3. 완료된 작업

### 3.1 `docs/llm-wiki/` (5 file 신규) ✓

| file | size | 책임 |
| --- | --- | --- |
| `README.md` | 7.8KB | 5 file root index + 두 wiki 분리 + 디렉터리 구조 + Phase 1 의 의의/한계 + forward path + Tier 정책 정합 |
| `scope-and-rationale.md` | 10.6KB | D-72 6 질문의 본 저장소 측 적용 + Phase 1 in-repo scope + scope 외 (forward path) + trade-off |
| `mirror-list.md` | 10KB | Phase 1 source list (core subset ~82 file) + Phase 3 scope 외 + lint 영향 + forward path |
| `lint-config.toml` | 4.4KB | per-project config (TOML) — L07 ADR 면제 + L10 devhub_adr_source_pattern + [meta] |
| `operation-sop.md` | 10.7KB | sync trigger + sync 절차 (dry-run/real/vault 부재) + lint trigger + sync 위험 10건 + verification |

### 3.2 `scripts/wiki-sync-devhub.sh` (NEW, 6.4KB, executable) ✓

- BSD-rsync safe pattern (my_harness 의 `wiki-sync-ai-workflow.sh` 와 동일).
- 7 source 패턴 (82 file, ~3.5MB estimated).
- `--dry-run` mode (source list 출력, no actual mirror) + (no option) real mirror + `--help`.
- Exit 0 success / 1 vault 부재 또는 source root 부재 또는 invalid option.
- Vault 부재 시 명시적 error + hint (my_harness 의 D-71 §2.2 wiki-init 또는 D-72 응답 §4 #1).

### 3.3 lint 4종 all PASS ✓

- `bash scripts/check-tier-separation.sh` — ✅ no changes between origin/main and HEAD
- `bash scripts/check-openapi-yaml-lint.sh` — ✅ passed
- `bash scripts/check-migration-uniqueness.sh` — ✅ valid and unique
- `python3.13 ai-workflow/tests/check_docs.py` — 본 PR 의 5 file 정합
- `bash scripts/wiki-sync-devhub.sh --dry-run` — ✅ PASS (82 file, exit 0)
- `HOME=/tmp/fake_home bash scripts/wiki-sync-devhub.sh` — ✅ vault 부재 시 명시적 error + exit 1

## 4. 잔여 / 후속 작업

### 4.1 본 PR 잔여
- **사용자 confirm 후** commit + push + gh pr create --body-file pr_body.md + gh pr merge --squash --delete-branch.
- main flat memory 3 file sync (post-merge).

### 4.2 carry-over (별도 PR / 사용자 trigger)
- **T-d-72-2** (P3): `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) — 사용자 confirm 후.
- **D-73** (my_harness 측): wiki-lint skill 에 `--project` + `--project-config` 옵션 추가.
- **D-74** (my_harness 측): my_harness 의 `_lint/my-harness/` + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업.
- **Phase 3** (mass ingest, 별도 PR): domain (66) + architecture + infrastructure + validation (~100 file) mirror + 30~50 wiki page.
- **wiki/cross/** (Phase 3 후속): cross-project 종합 (my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection).
- **v2.0** (forward): LLM 호출 + BM25+vector+MCP — my_harness 의 v2.0 경험 보고 진입.
- **N-13 release_v1_roadmap §3.5 정합** (housekeeping): N-13 row status = done 마킹.

## 5. 핵심 파일 / 라인 참조 (본 PR 시작 시점)

- `docs/llm-wiki/{README.md, scope-and-rationale.md, mirror-list.md, lint-config.toml, operation-sop.md}` (5 file 신규)
- `scripts/wiki-sync-devhub.sh` (1 file 신규, executable, 6.4KB)
- `ai-workflow/memory/feat/work_260610-d-72-wiki-phase-1/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-10.md, pr_body.md}` (5 file 신규)

## 6. 알아둘 trade-off (의도적 결정)

- **`docs/llm-wiki/` 선택 (vs `docs/wiki/` 또는 `docs/wiki-integration/`)**: 기존 `docs/wiki/` = **Public Wiki** (GitHub Wiki 게시 source, 인간 큐레이션) 임. 본 Phase 1 의 **LLM Wiki SSOT** 와 audience 다름. 디렉터리 이름 분리 = 두 wiki 의 명확한 구분.
- **mirror list 의 scope = core subset ~82 file**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture + infrastructure + validation (~100 file) 은 **Phase 3 (mass ingest)** 에서 별도 PR. 본 PR 의 lint 검증 + mirror 실행의 검증 가능한 정공법 = 작은 core subset.
- **lint-config.toml 의 L07 ADR 면제 config 작성 (옵션 미사용)**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 config 의 source 만 제공. 옵션 추가 후 자동 활성.
- **`~/wiki/` out-of-repo 변경 미포함**: 본 PR scope 의 의도적 한계. `~/wiki/raw/projects/devhub/`, `~/wiki/AGENTS.md`, `~/wiki/index.md`, `~/wiki/log.md`, `~/wiki/_lint/devhub/`, `~/wiki/wiki/projects/devhub/`, `~/wiki/wiki/cross/`, `~/wiki/schema/` 등 모두 out-of-repo = 본 PR scope 외.
- **mirror 실행은 본 PR scope 외**: 본 PR 의 lint 검증은 `bash scripts/wiki-sync-devhub.sh --dry-run` 의 source list 정합만. 실제 mirror 는 사용자 (yklee) 가 Phase 1 mirror 실행 시점에 진행. **이유**: mirror 실행은 `~/wiki/` 의 out-of-repo 변경 — 본 PR 의 in-repo 검증 가능 영역 (CI 4/4) 의 검증 범위 외.
- **scratch/ exclude**: `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) 는 본 Phase 1 의 reference. **mirror 미포함** (vault 비대화 + D-72 응답은 my_harness 측 작업).
- **`backend-core/` / `frontend/` / `backend-ai/` 의 source code mirror 제외**: vault 비대화 + LLM agent 의 코드 정합은 `code-index-update` 의 영역. 단 `docs/` 하위의 의미 있는 file (markdown + yaml + json) 만 mirror.

## 7. 다음 세션이 가장 먼저 할 일

1. `git status` / `git log --oneline -3` / `git branch --show-current` 확인 (현재 `feat/work_260610-d-72-wiki-phase-1`).
2. **사용자 confirm 후** `git add . + git commit + git push + gh pr create --body-file pr_body.md + gh pr merge --squash --delete-branch`. PR body 의 "추적성 영향" 섹션에 변경 file 6 (5 docs + 1 script) + Phase 1 source list 82 file 명시.
3. main flat memory 3 file sync (post-merge).
4. 또는 T-d-72-2 (`bash scripts/wiki-sync-devhub.sh` 1회 실행, real mirror, 사용자 confirm 후).
5. 또는 D-73 (my_harness 측 wiki-lint 옵션 추가) — 본 저장소 의 lint-config.toml 활성.
6. 또는 다른 sprint (backend-integration matrix / N-10 RBAC E2E 6 TC / N-13 release_v1_roadmap §3.5 정합).
