# Beta v0.7.9 — Release Pipeline 정식화 (2026-06-14)

> v0.7.6 metadata + v0.7.8 state-aware variant 의 *release layer* 보강.
> manual release 절차 (memory #5 / docs/RELEASE.md) 의 *기계화 layer*.

## 핵심 추가 (1 follow-up, Phase 1 of 2)

### tools/release_pipeline.py 정식화 (3 subcommand)

release 절차 (validate → version-bump → note-draft → gh release create) 의 *기계화 layer*. v0.7.0~v0.7.8 의 4 release 모두 manual 절차 (pyproject.toml version edit + Beta-v<X>.<Y>.<Z>.md 작성 + check_packaging + doctor + git tag) 였음. v0.7.9 에서 1차 부분 정식화.

**PyPI/TestPyPI 업로드 ❌** (memory #5 의 release 채널 정책 — GitHub Releases 만).
**gh release create trigger** ❌ (v0.7.10+ follow-up, 수동 trigger 유지).

```bash
# dry-run: 모든 subcommand plan 만 출력
python3 tools/release_pipeline.py validate --dry-run
python3 tools/release_pipeline.py version-bump --patch --dry-run
python3 tools/release_pipeline.py note-draft --from=v0.7.4-beta --to=0.7.9 --dry-run

# apply
python3 tools/release_pipeline.py version-bump --patch --apply
python3 tools/release_pipeline.py note-draft --from=v0.7.4-beta --to=0.7.9 --apply

# JSON 출력 (CI integration)
python3 tools/release_pipeline.py validate --json
```

**Tool 구조 (430 line, tools/release_pipeline.py)**:
- argparse 3 subcommand: `validate` / `version-bump` / `note-draft`
- 1차 출처: REPO_ROOT (workflow-source) + PYPROJECT + RELEASES_DIR
- tomllib (3.11+) / tomli (3.10) 분기

**3 subcommand**:

1. **`validate`**: 4 source 의 release-readiness 통합 검증
   - **check_packaging.py** (pyproject `[tool.setuptools.packages]` ↔ 디스크 정합성) — v0.5.7.1 hotfix 의 sub-package 누락 패턴 방지
   - **workflow_kit.cli.doctor** (v0.7.8, state-aware variant) — 7 baseline compliance
   - **state.json freshness** (last_freeze / last_ingest) — refresh_wiki_memory 의 갭 보강 정합
   - **git status** (working tree clean) — release commit 의 clean state 보장
   - `--skip-{packaging,doctor,state,git}` flag 로 부분 skip
   - 1+ source fail 시 exit 1 (CI integration)

2. **`version-bump`**: pyproject.toml version patch
   - `--patch` (default) / `--minor` / `--major` / `--to=...`
   - `--dry-run` / `--apply` mode
   - `parse_version`: `X.Y.Z` / `X.Y.Z-suffix` 모두 정합
   - re.sub 으로 `version = "..."` 1-line replace (back-up 자동 ❌ — caller 책임)

3. **`note-draft`**: git log `<from_tag>..HEAD` → release note skeleton 자동 생성
   - `--from=<tag> --to=<version>` (dest=from_tag, 'from' keyword 회피)
   - feat / chore / docs 분류 + 30 commit 까지 table 표시
   - `Beta-v<X>.<Y>.<Z>.md` 형식 skeleton (TL;DR / 핵심 추가 / Commit / Reference)
   - skeleton 만, *수동 편집* 후 commit

**Smoke test (8 test, 223 line)**:
- `test_validate_json_output`: validate --json → 4 source dict
- `test_version_bump_patch_dry_run`: 0.7.8 → 0.7.9 patch dry-run
- `test_version_bump_to_explicit`: --to=0.8.0 명시
- `test_version_bump_apply_and_restore`: 실제 pyproject.toml 갱신 + 원복 정합
- `test_note_draft_dry_run`: output_path + commits count
- `test_parse_version_formats`: `0.7.8` / `0.7.8-beta` / invalid
- `test_bump_version_logic`: major/minor/patch/to 4 branch
- `test_cli_subcommand_help`: --help + 3 subcommand 노출

## Deferred (v0.7.10+ follow-up)

| Deferred | 이유 | v0.7.10+ follow-up |
|---|---|---|
| `release` subcommand (gh release create 통합) | gh CLI 인증 토큰 회전 부담 + 수동 trigger 정책 | `gh release create v<X>.<Y>.<Z>-beta --target main --verify-tag dist/standard_ai_workflow-*b0*.whl --notes-file workflow-source/releases/Beta-v<X>.<Y>.<Z>.md` 의 자동 trigger (memory #5) |
| `verify` subcommand (tag push + GH release URL 검증) | tag push 후 GH release page 가 *유효* 인지 검증 필요 | `gh release view <tag> --json url` + `requests` 로 page GET |
| `rollback` subcommand (version + release note revert) | release 직후 critical issue 발견 시 | git revert 자동 + version roll-back + note 추가 |

## 검증

- **신규 test**: 8 (release_pipeline)
- **회귀 test**: 0 (7 check / 70 test PASS via run_all_checks)
  - check_baselines_compliance: 16/16
  - check_refresh_wiki_memory: 10/10
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
  - check_state_aware_baselines: 8/8
  - check_release_pipeline: 8/8 (본 release)
- **누적 119+ test PASS** (v0.7.8 111+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `cb0a892` | feat(v0.7.9): release_pipeline tool 정식화 (validate / version-bump / note-draft) + 8 smoke test |
| `d39be44` | chore(v0.7.9): version bump 0.7.8 → 0.7.9 + release note |

## 다음 (v0.7.10 / v0.8.0 후보)

- **release subcommand** (v0.7.9 의 deferred) — gh release create 자동 trigger
- **verify subcommand** (deferred) — tag push + GH release URL 검증
- **rollback subcommand** (deferred) — release 직후 critical issue rollback
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합

## Reference

- [v0.7.8 release note](Beta-v0.7.8.md) (직전) — state-aware + actual apply
- [v0.7.7 release note](Beta-v0.7.7.md) — display only + doctor integration
- [v0.7.6 release note](Beta-v0.7.6.md) — `load_config` / `DoctorConfig` 정의
- [memory #5 standard-ai-workflow.md](../../../../../../Users/yklee/.mavis/agents/mavis/memory/standard-ai-workflow.md) (release 채널 정책: GitHub Releases 만)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (수동 release 절차, 2026-06-08 작성)
- [tools/check_packaging.py](../tools/check_packaging.py) (v0.5.8+, packaging 정합성 검증)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (430 line, v0.7.9 본 release)
- [tests/check_release_pipeline.py](../tests/check_release_pipeline.py) (223 line, 8 test, 본 release)
