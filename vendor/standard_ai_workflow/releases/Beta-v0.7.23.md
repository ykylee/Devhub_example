# Beta v0.7.23 — Wiki 운영 cross-link 1-command wrapper (2026-06-15)

> `tools/refresh_wiki_memory.py` + `tools/emit_wiki_l2_body.py` 의 *3-step 운영 cycle* 을
> 1-command 로 묶음. 운영 시 *3번의 별도 invoke* 부담 zero.
> 3-step cycle = (1) L1 raw mirror 갱신 → (2) L2 dense 본문 emit → (3) L2 stub last_touched 갱신.

## 핵심 추가 (1 follow-up 본, 0 collateral, 0 deferred)

### 1. tools/wiki_emit.py — 1-command wrapper

**3-step 운영 cycle** (v0.7.5/v0.7.0/v0.7.5 의 tool 분산 운영의 *1-command 통합*):
1. **`refresh_wiki_memory.py --refresh-raw --apply`** — git log → release 별 분류 → 1차 출처 (raw mirror) 의 4 file 자동 보강 (state.json / work_backlog.md / wiki/log.md / memory/log.md)
2. **`emit_wiki_l2_body.py --apply`** — L1 raw mirror 본문 발췌 + L2 dense (`ai-workflow/wiki/sources/<stem>.md`) emit
3. **`refresh_wiki_memory.py --emit-l2 --apply`** — L2 stub 의 frontmatter 의 `last_touched` 갱신 (1차 출처의 in-repo retrieval 일관성 보장)

**argparse (10 flag)**:
- `--refresh-wiki` / `--emit-l2` / `--reemit-stubs` (sub-step 1개만)
- `--full` (default, 3-step cycle 전체)
- `--skip-1` / `--skip-2` / `--skip-3` (fine-grained skip)
- `--project <slug>` (multi-project 대비, default `standard-ai-workflow`)
- `--since <date>` (refresh_wiki_memory 의 git log --since 기준, default `2026-06-10`)
- `--max-chars <int>` (L2 dense 본문 cap, default 2000)
- `--dry-run` / `--apply` (default --apply)
- `--json` (CI 통합)

**REPO_ROOT 결정** (v0.7.17/v0.7.18 정공법 동일):
- `git rev-parse --show-toplevel` 우선
- `Path.cwd().resolve()` fallback
- 외부 vault (`~/wiki/`) reference 0 (in-repo storage 정공법)

**3-step 의 command build (in `--apply` mode)**:
```python
cmd_1 = [py, REFRESH_WIKI_MEMORY, "--refresh-raw", "--since", since, "--project", project, "--apply", "--json"]
cmd_2 = [py, EMIT_WIKI_L2_BODY, "--project", project, "--max-chars", str(max_chars), "--apply", "--json"]
cmd_3 = [py, REFRESH_WIKI_MEMORY, "--emit-l2", "--since", since, "--project", project, "--apply", "--json"]
```

**error handling** (v0.7.21 release coordination 정공법 정합):
- dry-run: subprocess 호출 0
- apply: 1+ step fail 시 *즉시 중단* + step name + stderr 포함 error dict return (memory #6 §R-4 의 "graceful fail" 와 동일 정신)

### 2. Smoke test (5 test 신규, check_v0_7_23_wiki_cross_link.py)

| Test | 검증 |
|---|---|
| `test_wiki_emit_dry_run_full_cycle` | 3-step 전체 (default) dry-run + 3 step planned + skipped_steps=[] |
| `test_wiki_emit_refresh_wiki_only` | `--refresh-wiki` 시 1단계만 (2/3 skipped) |
| `test_wiki_emit_emit_l2_only` | `--emit-l2` 시 2단계만 |
| `test_wiki_emit_reemit_stubs_only` | `--reemit-stubs` 시 3단계만 |
| `test_wiki_emit_skip_combinations` | `--skip-1/2/3` 의 3가지 조합 |

**5/5 PASS**.

### 3. Cross-cutting 효과 (DevHub PR #602 와 정합)

v0.7.17 의 wiki in-repo storage isolation + DevHub 의 PR #602 (in-repo L0 Home + L1 5 page + L2 5 dense) 와 *양방향 정합*:
- 본 project (standard_ai_workflow) 의 `tools/wiki_emit.py` = *operator tool* (1-command wrapper)
- DevHub 의 PR #602 = *wiki content* (4 phase raw mirror → L1 → L2 → L0)
- **정합**: 본 project 의 `wiki_emit.py` 가 DevHub 의 4 phase 를 *자동화* 가능. `wiki_emit.py --apply` 1회 = DevHub 의 4 phase cycle.

**Cross-project 적용** (my_harness, devhub):
- operator tool (= `wiki_emit.py`) + content (= 4 phase page) 의 *분리* 정공법. tool 은 본 project 표준 (1-command wrapper) 으로 통일, content 는 각 project 의 *도메인* (L1 5종: concept/decision/entity/pattern/topic).
- `wiki_emit.py --project=<slug>` 의 multi-project flag 가 *다른 project 도 같은 tool* 의 basis.

## Deferred (v0.7.24+)

| Deferred | 이유 | v0.7.24+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` |
| legacy L2 → in-repo migration (v0.7.17) | 외부 vault 의 기존 L2 file 의 in-repo 이관 | `tools/migrate_vault_l2_to_inrepo.py` |
| `check_workflow_linter.py` branch detection | mavis data dir 격리 환경의 branch name resolution | `workflow_kit.common.paths.get_current_branch` 의 4-priority auto-detect 적용 |
| `wiki_emit.py` 의 --since / --max-chars 의 interactive mode | TUI/GUI 가 아닌 plain argparse. CI 통합 위주 | TUI prompt 또는 pre-flight check + 친절한 error message |

## 검증

- **신규 test**: 5 (위 §2)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_v0_7_23_wiki_cross_link: 5/5 (신규, 본 release)
- **누적 197+ test PASS** (v0.7.22 192+ + 5 신규)

## Commit

| Hash | Subject |
|---|---|
| `b4936a2` | feat(v0.7.23): tools/wiki_emit.py 1-command wrapper (3-step cycle: refresh_raw + emit_l2 + reemit_stubs) + 5 smoke |

## 다음 (v0.7.24 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5 잔여) — GH Actions 또는 local pre-push hook
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **legacy L2 → in-repo migration** (v0.7.17 발견) — `tools/migrate_vault_l2_to_inrepo.py`
- **`check_workflow_linter.py` branch detection fix** (v0.7.22 발견) — 4-priority auto-detect 적용

## Reference

- [v0.7.22 release note](Beta-v0.7.22.md) (직전) — linter symlink-resolve fix
- [v0.7.21 release note](Beta-v0.7.21.md) — F-4 design 결함 fix (--allow-existing-tag + tag push coupling)
- [v0.7.18 release note](Beta-v0.7.18.md) — release coordination observability
- [v0.7.17 release note](Beta-v0.7.17.md) — wiki in-repo storage isolation
- [v0.7.5 release note](Beta-v0.7.5.md) — refresh_wiki_memory 정식화
- [v0.7.0 release note](Beta-v0.7.0.md) — LLM Wiki Layer (L1/L2 분화)
- DevHub PR #602 — wiki A안 (in-repo L0 Home + L1 5 page + L2 5 dense, 16/16 PASS)
- [workflow-source/tools/wiki_emit.py](../tools/wiki_emit.py) (~250 line, v0.7.23 본 release, 1-command wrapper)
- [workflow-source/tools/refresh_wiki_memory.py](../tools/refresh_wiki_memory.py) (~545 line, 1+3 단계)
- [workflow-source/tools/emit_wiki_l2_body.py](../tools/emit_wiki_l2_body.py) (~365 line, 2 단계)
- [tests/check_v0_7_23_wiki_cross_link.py](../tests/check_v0_7_23_wiki_cross_link.py) (~180 line, 5 test, 본 release)
