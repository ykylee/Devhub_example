# Phase 1 Scope + Rationale (D-72 Q1~Q6 본 저장소 측 적용)

- **문서 목적**: D-72 응답 (my_harness 의 작업 에이전트, RESPONSE.md, 15KB) 의 6 질문별 권장안을 본 저장소 (DevHub) 측에 적용한 정공법 + rationale. Phase 1 의 scope 한계 + Phase 3 (mass ingest) 까지의 forward path.
- **범위**: 본 file = 본 저장소 측의 D-72 Q1~Q6 적용 정공법. Q3 단순화 (lint L11 제거 + sa-internal/ 격리 제거) 의 본 저장소 측 영향 + 본 Phase 1 의 mirror list scope 결정 + lint-config 의 per-project config 적용.
- **대상 독자**: yklee (owner), LLM agent (Phase 3 mass ingest 시 wiki page 작성자), my_harness 작업 에이전트 (D-73 wiki-lint `--project` 옵션 추가 시).
- **상태**: in_progress (D-72 Phase 1, 2026-06-10)
- **최종 수정일**: 2026-06-10
- **관련 문서**:
  - [`../README.md`](../README.md) (5 file root index)
  - [`./mirror-list.md`](./mirror-list.md) (Phase 1 mirror list)
  - [`./lint-config.toml`](./lint-config.toml) (per-project config)
  - [`./operation-sop.md`](./operation-sop.md) (sync + lint SOP)
  - [`../../scratch/devhub_wiki_integration_response/RESPONSE.md`](../../scratch/devhub_wiki_integration_response/RESPONSE.md) (D-72 권장안, 15KB)

## 1. D-72 6 질문의 본 저장소 측 적용

D-72 응답 §2 의 6 질문 권장안을 본 저장소 측에 적용. 각 권장안 + 본 저장소 측 정합 + 본 Phase 1 의 deliverable.

### Q1 (vault 동거): D-72-A 채택 — `~/wiki/` 단일 + `wiki/projects/{devhub,my-harness}/` 동거

**권장 (D-72)**: A — 동거 가능, D-71.1 의도 유지. D-71.1 의 의도 = "second brain 은 out-of-repo, mobile-friendly, 사설 메모 흡수 가능".

**본 저장소 측 정합**:
- 본 저장소 (DevHub) 의 in-repo 변경 = `docs/llm-wiki/` (SSOT + config + SOP + mirror list) + `scripts/wiki-sync-devhub.sh` (sync script 의 source).
- `~/wiki/` 의 out-of-repo 변경 = `raw/projects/devhub/` (mirror 실행 후) — **본 PR scope 외**, 사용자가 `bash scripts/wiki-sync-devhub.sh` 실행 시 생성.
- 본 Phase 1 의 deliverable = **in-repo source-of-truth** (mirror 실행 trigger + 정책의 source). **mirror 실행 결과는 본 PR scope 외**.

**본 Phase 1 적용**:
- `docs/llm-wiki/{README.md, scope-and-rationale.md, mirror-list.md, lint-config.toml, operation-sop.md}` 5 file 작성 (본 PR scope).
- `scripts/wiki-sync-devhub.sh` 작성 (본 PR scope).
- `~/wiki/raw/projects/devhub/` 의 실제 mirror = **본 PR scope 외** (Phase 1 의 mirror list 의 source-of-truth 만 본 PR, mirror 실행은 사용자 confirm 후).

### Q2 (raw/ 정책): D-72-A+C 채택 — per-project 분리 + per-project manifest

**권장 (D-72)**: A+C — `raw/projects/{my-harness,devhub}/` per-project 격리 + `_manifest.md` per-project.

**본 저장소 측 정합**:
- `scripts/wiki-sync-devhub.sh` 의 DEST = `~/wiki/raw/projects/devhub/` (Q2-A 정합).
- `raw/projects/devhub/_manifest.md` 자동 생성 (Q2-C 정합). manifest 의 entry = `[YYYY-MM-DD HH:MM:SS] <project>=devhub <rel_path>=<src> size=<bytes>`.
- mirror list 의 source = `docs/llm-wiki/mirror-list.md` 의 "Phase 1 source list" (core subset).

**본 Phase 1 적용**:
- `scripts/wiki-sync-devhub.sh` 작성 (DEST = `~/wiki/raw/projects/devhub/`, manifest 자동 생성, BSD-rsync safe).
- `docs/llm-wiki/mirror-list.md` 의 source list = ADR (31) + governance (5) + planning (26) + setup (15) + requirements.md + openapi.yaml + ai-workflow-memory main flat 3 file = **~80 file (core subset)**.

### Q3 (3-tier 정책): D-72-A 단순화 채택 — 단일 vault, sa-internal/ 격리 + lint L11 불요

**권장 (D-72)**: A 단순화 — `~/wiki/` 가 yklee 개인 Gitea private repo 만 push. lint L11 + sa-internal/ 격리 불요.

**본 저장소 측 정합**:
- 본 저장소 의 3-tier 정책 (= GitHub vs 사내 SCM push 분리) 의 **적용 대상 = DevHub repo 자체** (`~/repos/Devhub_example_omp/`). `~/wiki/` 는 별도 Gitea private repo (my_harness 의 D-72 응답 §3) — 다른 정책.
- 따라서 `docs/llm-wiki/` 의 4 file + `scripts/wiki-sync-devhub.sh` 의 **모든 변경 = 공용** (사내 한정 정보 미포함). Tier 검증 PASS 예상.
- `mirror-list.md` 의 source list 는 **사외/공용 file 만** (ADR/governance/planning/setup/requirements/openapi/ai-workflow-memory main flat). 사내 한정 정보는 `infra/idp/_archive_*/`, `infra/idp/keycloak-realm.ci.json`, `docker-compose.{local,test,deploy,colima}.yml`, `scripts/setup-keycloak.sh` 등 — **mirror list 에 미포함**.

**잔여 주의사항 (D-72 응답 §3 + 본 저장소 측)**:
- **L07 (모순) false positive**: DevHub ADR-*.md 의 의도적 supersede (예: ADR-0030 → ADR-0031) 가 L07 false positive 가능 → Q4 의 per-project config 에서 ADR-*.md 만 L07 면제.
- **L08 (index 미등록)**: `wiki/projects/devhub/` 가 미작성 시 (`[[...]]` link 0) L08 false positive 가능 → Phase 3 의 wiki page 작성 후 해소.

**본 Phase 1 적용**:
- `mirror-list.md` 의 source list 에 **사내 한정 정보 미포함** (lint-config.toml 의 L11 도 미사용).
- `lint-config.toml` 에 L07 면제 (ADR-*.md) 만 포함.

### Q4 (schema + L01~L10): D-72-B 단순화 채택 — schema 단일, L01~L10 두 project 동시, L07 ADR 면제

**권장 (D-72)**: B 단순화 — schema/ 단일, L01~L10 동시 적용, 단 L07 만 DevHub ADR-*.md 면제.

**본 저장소 측 정합**:
- `docs/llm-wiki/lint-config.toml` 작성. Q4 의 per-project config 형식:
  ```toml
  [rules.L07]
  skip_paths = ["wiki/projects/devhub/sources/ADR-*.md"]
  ```
- yklee 의 본 저장소 측 결정 = lint-config 의 **project key = `devhub`**. my_harness 의 lint config 와 별개.
- wiki-lint skill 의 `--project` + `--project-config` 옵션은 **my_harness 측 D-73 의 작업** (my_harness 작업 에이전트가 wiki-lint skill 에 옵션 추가). 본 PR scope 외.

**본 Phase 1 적용**:
- `docs/llm-wiki/lint-config.toml` 작성 (L07 면제 for ADR-*.md). 옵션 추가 = my_harness 측 D-73.
- 본 PR 의 lint 검증은 **본 저장소 측 lint (check_docs.py) 만**. wiki-lint 의 L01~L10 은 Phase 1 의 mirror 실행 후 별도 lint (사용자 confirm 후, 본 PR scope 외).

### Q5 (v1.5 vs full compile): D-72 권장 — 두 project 동시 v1.5, v2.0 은 my_harness 선 → DevHub 후

**권장 (D-72)**: 동시 v1.5 시작. v2.0 full compile 은 my_harness 먼저 → DevHub.

**본 저장소 측 정합**:
- 본 Phase 1 = **v1.5 (schema + lint only)** 의 DevHub 측 진입. v1.5 의 scope = `docs/llm-wiki/` 의 4 file (SSOT + config + SOP + mirror list) + `scripts/wiki-sync-devhub.sh` (sync script 의 source).
- Phase 3 (mass ingest) = 30~50 wiki page 작성 — 별도 PR.
- v2.0 (full compile, LLM 호출 + BM25+vector+MCP) = **본 Phase 1 의 scope 외**, my_harness 의 v2.0 경험 보고 진입 (D-72 응답 §2 Q5).

**본 Phase 1 적용**:
- 본 PR 의 scope = v1.5 의 DevHub 측 in-repo 변경. v2.0 의 full compile 미실시.
- `docs/llm-wiki/operation-sop.md` 의 §3 (forward path) 에 v2.0 진입 시점 명시.

### Q6 (운영 메커니즘): D-72 권장 — 단일 AGENTS.md / index.md / log.md + per-project lint report

**권장 (D-72)**: 단일 AGENTS.md / index.md / log.md (out-of-repo) + per-project `_lint/<project>/` 분리.

**본 저장소 측 정합**:
- `~/wiki/AGENTS.md`, `~/wiki/index.md`, `~/wiki/log.md` = out-of-repo 변경. **본 PR scope 외**. my_harness 의 D-72 응답 §2 Q6 의 "단일 AGENTS.md" 결정 = my_harness 측 D-73 의 작업 (두 project 합의본 작성).
- `~/wiki/_lint/devhub/` 의 per-project lint report = **본 PR scope 외** (mirror 실행 + wiki-lint 의 `--project` 옵션 활성 후).
- 본 저장소 측의 정합 = **`docs/llm-wiki/lint-config.toml` 의 devhub config** 가 my_harness 의 D-73 작업 후 활성화. 본 PR 은 **config 의 source** 만 제공.

**본 Phase 1 적용**:
- 본 PR 의 scope = in-repo 변경. out-of-repo (`~/wiki/AGENTS.md` 등) 변경 미포함. my_harness 의 D-72 Q6 의 정공법 (단일 AGENTS.md) 은 my_harness 측 D-73 에서 진행.

## 2. Phase 1 scope 결정

### 2.1 본 PR scope (in-repo 변경)

**신규 file 6**:
1. `docs/llm-wiki/README.md` (5 file root index)
2. `docs/llm-wiki/scope-and-rationale.md` (본 file)
3. `docs/llm-wiki/mirror-list.md` (Phase 1 mirror list, core subset ~80 file)
4. `docs/llm-wiki/lint-config.toml` (per-project config, L07 ADR 면제)
5. `docs/llm-wiki/operation-sop.md` (sync + lint SOP)
6. `scripts/wiki-sync-devhub.sh` (sync script, BSD-rsync safe, dry-run + vault-absent no-op)

**수정 file 0** (본 PR).

**합 6 file 신규 + 0 file 수정 = 6 file 변경 (in-repo)**.

**Tier**: 모두 **공용** (사내 한정 정보 미포함). `bash scripts/check-tier-separation.sh` PASS 예상.

### 2.2 본 PR scope 외 (forward path)

| 단계 | 작업 | 의존 |
| --- | --- | --- |
| **Phase 1 mirror 실행** (본 PR 머지 후) | 사용자 (yklee) 가 `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~80 file mirror + `_manifest.md` 자동 생성 | 본 PR 머지 |
| **D-73 (my_harness 측)** | wiki-lint skill 에 `--project` + `--project-config` 옵션 추가 | Phase 1 mirror 실행 |
| **D-74 (my_harness 측)** | my_harness 의 `_lint/my-harness/` 디렉터리 + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업 | D-73 |
| **Phase 3 (mass ingest)** | 별도 PR — DevHub 의 docs/domain + docs/architecture + docs/infrastructure + docs/validation (총 ~100 file) 의 mirror + `wiki/projects/devhub/{concepts,entities,topics,sources}/` 30~50 wiki page 작성 | D-73 + D-74 |
| **wiki/cross/** | cross-project 종합 페이지 (my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection, 1~3 page) | Phase 3 |
| **wiki-lint CI integration** | `ci.yml` 의 별도 lint job 또는 e2e shard 의 lint step 추가 | D-73 |
| **v2.0 (full compile)** | LLM 호출 + BM25+vector+MCP — my_harness 의 v2.0 경험 보고 진입 (D-72 Q5) | my_harness v2.0 |
| **N-13 release_v1_roadmap §3.5 정합** | N-13 row status = done (D-72 D-73 D-74 D-75) 마킹, 별도 housekeeping PR | 본 PR 머지 + Phase 1 mirror |

## 3. 알려둘 trade-off (의도적 결정)

- **`docs/llm-wiki/` 선택 (vs `docs/wiki/` 또는 `docs/wiki-integration/`)**: 기존 `docs/wiki/` 가 **Public Wiki** (GitHub Wiki 게시 source, 인간 큐레이션) 임. 본 Phase 1 의 **LLM Wiki SSOT** 와 audience 다름. 디렉터리 이름 분리 = 두 wiki 의 명확한 구분. **`docs/wiki/` (Public) ↔ `docs/llm-wiki/` (LLM)** 의 cross-link 없음.
- **`docs/llm-wiki/` 의 5 file 모두 본 PR scope (in-repo)**: my_harness 의 D-72 응답은 단일 wiki vault + wiki/projects/{devhub,my-harness}/ 동거 였지만, 본 저장소 측의 in-repo SSOT 는 별도 위치. **본 저장소 의 SSOT 가 `docs/llm-wiki/`, my_harness 의 wiki vault 가 `~/wiki/wiki/projects/devhub/`**. 두 SSOT 의 일관성은 lint-config.toml + mirror list 로 유지.
- **mirror list 의 scope = core subset (~80 file)**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture + infrastructure + validation (~100 file) 은 **Phase 3 (mass ingest)** 에서 별도 PR. 본 PR 의 lint 검증 + mirror 실행의 **검증 가능한 정공법** = 작은 core subset.
- **lint-config.toml 의 L07 ADR 면제 config 작성 (옵션 미사용)**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 **config 의 source 만 제공**. 옵션 추가 후 자동 활성.
- **`~/wiki/` out-of-repo 변경 미포함**: 본 PR scope 의 의도적 한계. `~/wiki/raw/projects/devhub/`, `~/wiki/AGENTS.md`, `~/wiki/index.md`, `~/wiki/log.md`, `~/wiki/_lint/devhub/`, `~/wiki/wiki/projects/devhub/`, `~/wiki/wiki/cross/`, `~/wiki/schema/` 등 모두 out-of-repo = 본 PR scope 외. **본 PR 의 source-of-truth 만 in-repo**.
- **mirror 실행은 본 PR scope 외 (본 PR 의 lint 검증은 `bash scripts/wiki-sync-devhub.sh --dry-run` 의 source list dry-run 만)**: 실제 mirror 는 사용자 (yklee) 가 Phase 1 mirror 실행 시점에 진행. **이유**: mirror 실행은 `~/wiki/` 의 out-of-repo 변경 — 본 PR 의 in-repo 검증 가능 영역 (CI 4/4) 의 검증 범위 외.

## 4. 다음 세션 directive

1. `docs/llm-wiki/mirror-list.md` 작성 (Phase 1 source list, core subset).
2. `docs/llm-wiki/lint-config.toml` 작성 (per-project config, L07 ADR 면제).
3. `docs/llm-wiki/operation-sop.md` 작성 (sync + lint SOP).
4. `scripts/wiki-sync-devhub.sh` 작성 (sync script, BSD-rsync safe, dry-run + vault-absent no-op).
5. sprint memory 5 file + main flat memory 3 file.
6. lint 검증 (4종) + script smoke test (`bash scripts/wiki-sync-devhub.sh --dry-run`).
7. commit + push + PR 발행.
8. main flat memory finalize (post-merge sync).
9. 또는 T-d-72-2 (`bash scripts/wiki-sync-devhub.sh` 1회 실행, 사용자 confirm 후).
