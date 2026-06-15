# AI-Workflow Vendor Import Notes (v0.7.17, 2026-06-15)

> 표준화 워크플로우 v0.7.17 (standard_ai_workflow) 의 도구/스킬 발매 패키지를 본 저장소 (DevHub) 에 vendor import 한 작업의 메모. 향후 vendor 갱신 + 운영 시 참조.

## 1. import 범위

**Source**: [`ykylee/standard_ai_workflow`](https://github.com/ykylee/standard_ai_workflow) main 브랜치 commit `4d09dee` (v0.7.17, 2026-06-15)
**Local path**: `vendor/standard_ai_workflow/`
**Upstream metadata**: [`vendor/standard_ai_workflow/.upstream-url`](../../vendor/standard_ai_workflow/.upstream-url)

**가져온 것** (전체 28M, raw 52M → dist/build/egg-info 제거 후):
- `workflow-source/tools/` (7 도구, 핵심):
  - `refresh_wiki_memory.py` (v0.7.5+, v0.7.17 in-repo redirect)
  - `emit_wiki_l2_body.py` (v0.7.6+, v0.7.17 in-repo redirect)
  - `score_wiki_maintainability.py` (v0.7.17 in-repo redirect)
  - `score_wiki_trend.py`
  - `check_packaging.py`
  - `fill_reverse_engineering_artifacts.py`
  - `release_pipeline.py` (4-phase, v0.7.15+ changelog-gen --from-tag/--to-tag)
- `workflow-source/tests/` (70+ check_*.py, 핵심 4 + v0.7.17 신규):
  - `check_v0_7_17_wiki_in_repo_isolation.py` (**11/11 PASS**)
  - `check_atomic_write.py` (3/3 PASS, v0.7.15 신규)
  - `check_refresh_wiki_memory.py` (v0.7.17 in-repo redirect)
  - `check_wiki_drift.py` (v0.7.17 in-repo redirect)
  - `check_release_pipeline_changelog_gen.py`
- `workflow-source/{core,workflow_kit,scripts,skills,harnesses,extensions,mcp_servers,templates,schemas,examples,prompts,global-snippets}/` (full source)
- `workflow-source/ai-workflow/` (= standard_ai_workflow 의 자체 active memory + wiki mirror, vendor 내 mini structure)
- `workflow-source/{CHANGELOG.md, MEMORY_GOVERNANCE.md, README.md, pyproject.toml}`
- `workflow-source/releases/Beta-v0.7.17.md` (v0.7.17 release note)
- `workflow-source/dist/`, `build/`, `*.egg-info/`, `__pycache__/*` (제거)

**제외**:
- `workflow-source/dist/` (52M wheel/sdist, build artifact)
- `workflow-source/standard_ai_workflow.egg-info/` (build metadata)
- `workflow-source/dist 오후 10.44.30/`, `dist 오후 11.10.58/` (소스 내 build 오염, 향후 mavis-trash)

## 2. 격리 이유 (vendor side-by-side)

**이유**: standard_ai_workflow 와 본 저장소 (DevHub) 는 **같은 이름의 `ai-workflow/` 디렉터리** 를 가지지만 의미가 다름.

| | standard_ai_workflow | DevHub |
|---|---|---|
| `ai-workflow/` 의미 | **도구/스킬 발매 패키지** (workflow-source 의 sibling) | **프로젝트 운영 메모리/스크립트/스킬/테스트 인프라** (AGENTS.md workflow layer) |
| `workflow-source/` 의미 | raw source code (build target) | (없음) |
| `vendor/standard_ai_workflow/` 의미 | (없음) | **발매자 패키지의 격리 import** (read-only reference + 도구 사용) |

**충돌 회피**: 통째 merge 시 동명 dir (`core/`, `memory/`, `scripts/`, `skills/`, `tests/`, `workflow_kit/`, `harnesses/`, `mcp_servers/`) 의 의미가 섞여서 *DevHub 자체 운영 workflow* 와 *vendor 의 도구* 가 구분 불가. `vendor/` 격리로 양쪽 무영향 + 정합성 검증 가능.

## 3. Wiki 운영 in-repo 전환 (v0.7.17 N/A → applicable)

**사용자 결정 (2026-06-15)**: "이 저장소에서 `~/wiki` 에 반영하거나 확인하는 부분은 끊고 이 저장소의 위키는 저장소 내 위키만 사용하는 것으로 정리"

**적용 (5 file + 7 script + 6 dir)**:
1. `docs/llm-wiki/{README.md, scope-and-rationale.md, mirror-list.md, lint-config.toml, operation-sop.md, ingest-skill.md, pr-update-skill.md, query-skill.md}` 8 file — `~/wiki/` literal 0 회, in-repo path 로 redirect (DONE)
2. `scripts/wiki-{sync-devhub,ingest-from-raw,query,pr-update,status-check,mass-ingest,frontmatter-update}.sh` 7 script — `VAULT_ROOT="${HOME}/wiki"` → `VAULT_ROOT="${SRC}/ai-workflow/wiki"` (DONE)
3. `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics,sources}/` 6 dir 신규 + .gitkeep (DONE)
4. `ai-workflow/memory/log.md` 신규 (v0.7.17 wiki event log target, DONE)
5. `tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py` 신규 (4/4 PASS) — 본 저장소 자체 invariant 검증

**Follow-up (별도 PR)**:
- 7 script 의 *my_harness 측 wiki-* skill 호출 (`~/repos/my_harness/ai-workflow/skills/wiki-*/...`) 도 *in-repo 의 `vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py --apply`* 또는 *in-repo 의 `vendor/standard_ai_workflow/tools/score_wiki_maintainability.py`* 로 redirect. **이번 PR scope 외** (PR body 의 Future work 섹션에 명시).
- 글로벌 mavis 4 wiki thin wrapper (`~/.mavis/skills/wiki-{event-sync,source-sync,prompt-log,query-helper}/`) 호출은 *본 저장소 컨텍스트 작업 시 회피* — 운영 SOP.

## 4. 검증 (smoke test 15/15 PASS)

```bash
# v0.7.17 vendor smoke (외부 vault reference 0 검증, 11/11 PASS)
python3 vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py

# DevHub 자체 invariant (in-repo path 정합, 4/4 PASS)
python3 tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py
```

**v0.7.17 vendor smoke 11 test**:
1. `refresh_wiki_memory` no VAULT_ROOT (외부 vault reference 0)
2. `refresh_wiki_memory` RAW_FILES in-repo
3. `refresh_wiki_memory` L2_STUBS in-repo (4 file)
4. `emit_wiki_l2_body` no VAULT_ROOT
5. `emit_wiki_l2_body` REPO_ROOT auto-detect (git rev-parse)
6. `score_wiki_maintainability` L2 in-repo
7. `check_refresh_wiki_memory` no VAULT_ROOT (test 자체)
8. `check_wiki_drift` raw_mtime in-repo
9. `vendor/ai-workflow/wiki/sources/` dir 존재
10. `vendor/ai-workflow/memory/log.md` 존재
11. legacy symlink / path 0

**DevHub invariant 4 test**:
1. `scripts/wiki-*.sh` 의 `VAULT_ROOT="${HOME}/wiki"` 0 회
2. `docs/llm-wiki/*.md` 의 `~/wiki/` literal 0 회
3. `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics,sources}/` 6 dir 존재
4. `ai-workflow/memory/log.md` 존재

**합 15/15 PASS** (2026-06-15 기준).

## 5. v0.7.15 / v0.7.16 의 범용 패턴 (이미 vendor 안에 포함)

v0.7.17 의 vendor import 시 v0.7.15 + v0.7.16 의 변경도 함께 들어왔어. **DevHub 의 자체 도구/스크립트에도 적용 가능한 범용 패턴**:

- **v0.7.15 `atomic_write` helper**: `tools/refresh_wiki_memory.py` 등 4 file 에 도입. **DevHub 의 `ai-workflow/scripts/generate_workflow_state.py` 등** 의 *메모리 file write* 부분에 도입 후보 (F-followup). 정공법 = read-then-write 의 *atomicity* 보장 (partial write 시 원본 보존).
- **v0.7.15 `changelog-gen --from-tag/--to-tag` filter**: `tools/release_pipeline.py` 의 Phase 4. DevHub 의 자체 release 노트/PR body 자동 생성 도구와 정합 가능.
- **v0.7.16 `tool.workflow-doctor` config thresholds/excluded_paths**: `config.toml` 의 `thresholds` + `excluded_paths` 적용. DevHub 의 자체 lint/config 도구 (e.g. `scripts/check_tier_separation.sh`) 의 *경계값* 정의에 정합 가능.
- **v0.7.16 linter IndentationError fix**: dead code 의 indent 가 IndentationError 유발. DevHub 의 *dead code lint* 도입 후보.

**이번 PR scope 외** (각 follow-up PR).

## 6. 운영 / 갱신 SOP

- **vendor 갱신**: standard_ai_workflow 의 새 release 시 `git pull` 후 `cp -R ~/repos/standard_ai_workflow_minimax/workflow-source/. vendor/standard_ai_workflow/` (build/dist/pycache 제외, §1 의 filter).
- **v0.7.18+ smoke 추가**: vendor 의 `tests/check_v0_7_18_*.py` 가 vendor import 시 자동 포함됨. 본 저장소 의 4-test invariant smoke (위 §4 의 DevHub invariant) 는 *수정* 이 필요한 경우만 갱신.
- **conflict resolution**: vendor 가 28M 이고 거의 read-only. 충돌 가능성 = *본 저장소 측의 5 file* (scripts/wiki-*.sh, docs/llm-wiki/*.md) 만. vendor 갱신 후 *본 저장소 측의 in-repo redirect* 가 회귀 안 됐는지 §4 의 4-test invariant 로 즉시 확인.

## 7. 메타 (commit hash + cross-reference)

- **vendor import commit**: 본 PR 의 HEAD commit (작성 시점 2026-06-15, commit hash 7자 prefix는 PR 머지 후 결정).
- **v0.7.17 upstream commit**: `4d09dee` ([ykylee/standard_ai_workflow main](https://github.com/ykylee/standard_ai_workflow/commit/4d09dee))
- **위키 운영 in-repo 결정**: 사용자 지시 2026-06-15.
- **smoke 15/15 PASS**: `python3 vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py` + `python3 tests/check_v0_7_17_devhub_wiki_in_repo_invariant.py`.
