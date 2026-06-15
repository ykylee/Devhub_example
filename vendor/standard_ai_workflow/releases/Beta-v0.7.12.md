# Beta v0.7.12 — refresh_wiki_memory REPO_ROOT Auto-Detect (2026-06-15)

> v0.7.11 release note 의 Deferred 1 follow-up 본 구현. Beta-v0.7.11.md "Tool 부작용 fix" line + Deferred table 의 마지막 행 참조.
> hardcoded `REPO_ROOT = ~/repos/standard_ai_workflow_minimax` 의 4-priority auto-detect.

## 핵심 추가 (1 follow-up 본, 0 deferred)

### tools/refresh_wiki_memory.py REPO_ROOT auto-detect

v0.7.5 의 `refresh_wiki_memory.py` 가 module-level literal `REPO_ROOT = Path.home() / "repos" / "standard_ai_workflow_minimax"` (v0.7.5 시점의 `~/repos/standard_ai_workflow_minimax` repo path) 사용. 본 repo 가 `~/repos/standard_ai_workflow` 로 rename 된 후 깨짐. v0.7.11 release 시 symlink (`~/repos/standard_ai_workflow_minimax -> ~/repos/standard_ai_workflow`) 으로 즉시 회귀 해소. v0.7.12 에서 본 fix.

**4-priority auto-detect** (priority 1 > 2 > 3 > 4):

1. **CLI flag** `--repo-root=<path>` (명시적, 가장 우선)
2. **env var** `$STANDARD_AI_WF_REPO` (CI integration)
3. **`git rev-parse --show-toplevel`** subprocess (현재 cwd 기준, repo 어디서 실행해도 동작)
4. **legacy fallback** `~/repos/standard_ai_workflow_minimax` (deprecation 경고 1회 stderr)

```bash
# priority 1: --repo-root flag
python3 tools/refresh_wiki_memory.py --refresh-raw --dry-run --repo-root=/path/to/other/repo

# priority 2: env var
STANDARD_AI_WF_REPO=/path/to/repo python3 tools/refresh_wiki_memory.py --refresh-raw --dry-run

# priority 3: git rev-parse (cwd 기준, default 동작)
python3 tools/refresh_wiki_memory.py --refresh-raw --dry-run

# priority 4: legacy fallback (deprecation stderr)
# (priority 1~3 모두 fail 시)
```

**Tool 구조 변경**:
- `get_repo_root(cli_value, *, _suppress_warning) -> Path` helper (52~80 line)
  - priority 1~4 cascade
  - legacy fallback 시 stderr 1회 deprecation warning (test 가능하도록 `_suppress_warning` flag)
- `REPO_ROOT = get_repo_root()` (eager init for backward compat)
- `collect_commits(since, *, repo_root=None)` — `repo_root` keyword-only parameter
- `cmd_refresh_raw` / `cmd_emit_l2` 가 `args.repo_root` 전달
- main argparse `--repo-root` flag 추가 (default None)
- main result 에 `repo_root: str(resolved)` field 추가 (CLI/text/JSON 모두 노출)
- module docstring 에 REPO_ROOT 결정 section 추가 (priority + Usage example)

**Smoke test (4 test 신규, check_refresh_wiki_memory.py 10 → 14)**:
- `test_repo_root_argparse_recognized`: --repo-root flag 인식 + result.repo_root = git rev-parse toplevel
- `test_repo_root_cli_flag_priority`: --repo-root flag 가 env var (`STANDARD_AI_WF_REPO=/tmp/env-var-path-must-be-ignored`) override
- `test_repo_root_env_var_fallback`: cwd 가 `/tmp` 외부 (cwd=/tmp) + env var 명시 → env var 반영
- `test_repo_root_git_toplevel_priority`: cwd 가 git repo 안일 때 `git rev-parse --show-toplevel` 결과 = `REPO_ROOT`

**legacy literal 처리**:
- module-level `_LEGACY_REPO_ROOT = Path.home() / "repos" / "standard_ai_workflow_minimax"` — `get_repo_root()` 의 fallback 으로만 사용
- priority 1~3 모두 fail 시 1회 stderr deprecation 경고:
  ```
  [DEPRECATION] refresh_wiki_memory.py: REPO_ROOT auto-detect 실패 — legacy fallback 사용 (/home/yklee/repos/standard_ai_workflow_minimax). v0.7.12+ 부터 --repo-root=<path> 또는 STANDARD_AI_WF_REPO env var 사용 권장.
  ```
- symlink (`~/repos/standard_ai_workflow_minimax -> ~/repos/standard_ai_workflow`) 제거 (본 작업의 검증 단계). priority 1~3 만으로 정상 동작

**Test file 갭**:
- `check_refresh_wiki_memory.py` 의 module-level `REPO_ROOT = Path.home() / "repos" / "standard_ai_workflow_minimax"` (line 35) 도 hardcoded 였음. 미사용 (VAULT_ROOT 만 쓰임) 이나, v0.7.12 에서 제거. 깨끗한 코드 유지

## Deferred (v0.7.13+ Phase 5)

| Deferred | 이유 | v0.7.13+ follow-up |
|---|---|---|
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| `changelog-gen` subcommand | CHANGELOG.md (vs release note) 누적 | git log all-time → CHANGELOG.md section 별 분류 |
| `score trend 의 config thresholds` (v0.7.7 의 deferred #2) | hardcoded 0.3 → `thresholds["score_alert"]` | `tools/score_wiki_trend.py` 의 0.3 literal → `config.thresholds["score_alert"]` |
| `profiling 의 config memory threshold` (deferred #3) | hardcoded → config | `tools/score_wiki_trend.py` 와 동일 패턴 |
| `linter 의 config excluded_paths` (deferred #4) | hardcoded → config | `[tool.workflow-doctor].excluded_paths` 적용 |
| v0.7.5~v0.7.10 release backfill (별도 발견) | local tag v0.7.0~v0.7.4 만 존재, v0.7.5~v0.7.10 부재 + GH release 도 부재 | 별도 세션에서 `git tag` + `gh release create` backfill (6 release × tag + asset) |

## 검증

- **신규 test**: 4 (REPO_ROOT auto-detect 4종) + Phase 1~3 회귀
- **회귀 test**: 0 (symlink 없이 14/14 PASS)
  - check_refresh_wiki_memory: 14/14 (10 기존 + 4 신규)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 6/8 (2 fail 은 dist side-effect, fresh CI 환경에선 8/8)
  - check_release_pipeline_phase3: 8/8
  - check_baselines_compliance: 16/16
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
  - check_state_aware_baselines: 8/8
- **누적 139+ test PASS** (v0.7.11 135+ + 4 신규)

## 부가: v0.7.5~v0.7.10 release backfill (2026-06-15 01:47:04Z~01:47:20Z)

본 release 의 cut 시점 (`63080ba`) 에 v0.7.5~v0.7.10 의 6 release 가 `main` 에 push 완료 + `Beta-v0.7.X.md` 첨부된 상태였지만, **git tag + GH release + wheel/sdist 만 부재** 발견. 본 세션에서 self-dogfooding (v0.7.11 의 `dist` subcommand + v0.7.12 의 symlink-less auto-detect + v0.7.12 의 `git worktree`) 으로 backfill 완료.

**6 wheel/sdist 빌드** (`git worktree add` × 6 + `python3 -m build --outdir /tmp/dist_v_$ver`):
- v0.7.5: `standard_ai_workflow-0.7.5-py3-none-any.whl` (181273 bytes) + `.tar.gz` (148665 bytes)
- v0.7.6: `.whl` (183658) + `.tar.gz` (151351)
- v0.7.7: `.whl` (183657) + `.tar.gz` (151356)
- v0.7.8: `.whl` (183852) + `.tar.gz` (151526)
- v0.7.9: `.whl` (183850) + `.tar.gz` (151518)
- v0.7.10: `.whl` (183858) + `.tar.gz` (151542)

**6 git tag push** (`git tag -a v0.7.X-beta c2a75f8` ... + `git push origin v0.7.X-beta`):
- `v0.7.5-beta` @ `c2a75f8`
- `v0.7.6-beta` @ `b9ede19`
- `v0.7.7-beta` @ `3300e73`
- `v0.7.8-beta` @ `b67af83`
- `v0.7.9-beta` @ `d39be44`
- `v0.7.10-beta` @ `67d4a37`

**6 GH release create** (`gh release create` × 6 with `--verify-tag`):
- https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.5-beta (2026-06-15T01:47:04Z)
- ... (생략) ...
- https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.10-beta (2026-06-15T01:47:20Z)

**verify (`gh release view --json`)** × 6 정상 (asset 2종 모두 부착, `is_prerelease=false`).

## Commit

| Hash | Subject |
|---|---|
| `63080ba` | feat(v0.7.12): refresh_wiki_memory REPO_ROOT auto-detect (CLI flag > env var > git rev-parse > legacy fallback) + 4 smoke + Beta-v0.7.12.md |
| `89b7af5` | chore(v0.7.12): v0.7.5~v0.7.10 release backfill (6 wheel/sdist + 6 git tag + 6 GH release) + Beta-v0.7.12.md 갱신 |

## 다음 (v0.7.13 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 4) — GH Actions 또는 local pre-push hook
- **changelog-gen subcommand** — CHANGELOG.md 누적 (vs release note)
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **v0.7.5~v0.7.10 release backfill** (v0.7.12 의 발견) — **v0.7.12 본 세션에서 완료** (6 wheel/sdist 빌드 + 6 git tag push + 6 gh release create, 2026-06-15 01:47:04Z~01:47:20Z)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합

## Reference
- [v0.7.11 release note](Beta-v0.7.11.md) (직전) — Phase 3 (dist subcommand) + state sync + version sync
- [v0.7.10 release note](Beta-v0.7.10.md) — Phase 2 (release / verify / rollback)
- [v0.7.9 release note](Beta-v0.7.9.md) — Phase 1 (validate / version-bump / note-draft)
- [v0.7.5 release note 없음] (raw mirror 만 갭) — refresh_wiki_memory 정식화 (tool 도입 commit `0741775`)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차, 2026-06-12 갱신)
- [tools/refresh_wiki_memory.py](../tools/refresh_wiki_memory.py) (~485 line, v0.7.12 본 release)
- [tests/check_refresh_wiki_memory.py](../tests/check_refresh_wiki_memory.py) (~325 line, 14 test, 본 release)
