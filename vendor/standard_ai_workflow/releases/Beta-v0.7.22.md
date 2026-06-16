# Beta v0.7.22 — linter symlink-resolve fix (2026-06-15)

> v0.7.16 release note 의 *linter symlink-resolve* 의 1차 발견 (memory #22 §Part D) 의 *본 fix*.
> `workflow_kit/common/linter.py` 의 `Path.resolve()` 가 mavis data dir 격리 환경
> (e.g. `.mavis` → `.minimax` symlink) + macOS `/var` symlink 환경에서 *정상 relative
> path* 를 *broken* 으로 false-positive 보고. `.absolute()` 로 변경 (symlink 보존 +
> cwd 기준 정규화만).

## 핵심 추가 (1 follow-up 본, 0 collateral, 0 deferred)

### 1. workflow_kit/common/linter.py — `.resolve()` → `.absolute()` (2 site)

**v0.7.16 의 1차 발견 (memory #22 §Part D)**:
- `Path.resolve()` 가 symlink 따라감 → `tmp_dir/README.md` 의 real path 가 `tmp_dir` 의 parent 로 normalize (mavis data dir 격리 환경)
- 정상 relative path 의 `5 levels up` 이 `4 levels up` 으로 계산되어 broken link false-positive
- `check_workflow_linter.py` 의 `test_linter_pass` 가 mavis data dir 격리 환경에서 fail

**v0.7.22 fix**: 2 site 의 `.resolve()` → `.absolute()`:
- `linter.py:59` (maturity consistency 의 test_path 검증): `(project_root / test_path_str).resolve()` → `.absolute()`
- `linter.py:181` (broken link check 의 link_path 계산): `(path.parent / clean_link).resolve()` → `.absolute()`

**`Path.resolve()` vs `Path.absolute()` 의 차이**:
- `.resolve()`: `realpath(3)` 와 동등 — symlink 따라가서 real path. *false-positive* 의 원인.
- `.absolute()`: `Path.cwd() / path` — symlink 보존 + cwd 기준 정규화만. *사용자 의도 보존*.
- `Path.resolve(strict=False)`: default 와 동일하게 symlink 따라감. *strict mode* 가 도움 안 됨.

### 2. Smoke test (3 test 신규, check_v0_7_22_linter_symlink_resolve.py)

| Test | 검증 |
|---|---|
| `test_linter_broken_link_symlink_aware` | symlink 환경 시뮬레이션 (cwd 가 symlink → real path) 의 broken link check 가 정상 relative path 를 broken 으로 false-positive 보고 안 함 |
| `test_linter_maturity_test_path_symlink_aware` | 동일 환경의 maturity consistency 의 test_path 검증 |
| `test_linter_resolve_to_absolute_invariant` | 정공법 자체 — `.resolve()` 와 `.absolute()` 가 symlink 환경에서 *다른* 결과. 즉 *본질* 검증 |

**3/3 PASS**.

### 3. Cross-cutting 적용 (memory #6 R-4 의 정공법)

본 fix 는 memory #6 의 R-4 (Runtime config mutate binary test 격리) 의 *symlink-aware* 정공법의 *동일 정신*:

- memory #6: `MYHARNESS_HOME` / `XDG_CONFIG_HOME=tempdir` override 시 `mavis-trash` recovery
- v0.7.22: `Path.absolute()` (symlink 보존) 으로 *사용자 의도* 의 relative path 보존
- **공통 원칙**: 환경 격리 (mavis data dir, XDG_CONFIG_HOME) 시 *symlink 의 follow* 가 *사용자 의도* 를 깨면 안 됨. *absolute or explicit override* 가 정공법.

## Deferred (v0.7.23+)

| Deferred | 이유 | v0.7.23+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` |
| legacy L2 → in-repo migration (v0.7.17) | 외부 vault 의 기존 L2 file 의 in-repo 이관 | `tools/migrate_vault_l2_to_inrepo.py` |
| `check_workflow_linter.py` 의 branch detection (별개) | mavis data dir 격리 환경의 branch name resolution | `workflow_kit.common.paths.get_current_branch` 의 4-priority auto-detect 적용 |

## 검증

- **신규 test**: 3 (위 §2)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_v0_7_22_linter_symlink_resolve: 3/3 (신규, 본 release)
- **누적 192+ test PASS** (v0.7.21 189+ + 3 신규)

## Commit

| Hash | Subject |
|---|---|
| `3c12950` | fix(v0.7.22): workflow_kit/common/linter.py .resolve() → .absolute() (mavis data dir 격리 환경 + macOS /var symlink fix) + 3 smoke |

## 다음 (v0.7.23 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5 잔여) — GH Actions 또는 local pre-push hook
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **legacy L2 → in-repo migration** (v0.7.17 발견) — `tools/migrate_vault_l2_to_inrepo.py`
- **check_workflow_linter.py branch detection fix** (v0.7.22 발견) — `workflow_kit.common.paths.get_current_branch` 의 4-priority auto-detect 적용

## Reference

- [v0.7.21 release note](Beta-v0.7.21.md) (직전) — F-4 design 결함 fix (--allow-existing-tag + tag push coupling)
- [v0.7.18 release note](Beta-v0.7.18.md) — release coordination observability 의 1차 출처
- [v0.7.17 release note](Beta-v0.7.17.md) — wiki in-repo storage + collateral 발견
- [v0.7.16 release note](Beta-v0.7.16.md) — config + linter fix + **linter symlink-resolve 발견의 1차 출처**
- memory #6 §R-4 (Runtime config mutate binary test 격리 — symlink-aware 정공법)
- memory #22 §Part D (linter 의 symlink-resolve follow-up)
- [workflow_kit/common/linter.py](../workflow_kit/common/linter.py) (~213 line, v0.7.22 본 release, .absolute() 적용)
- [tests/check_v0_7_22_linter_symlink_resolve.py](../tests/check_v0_7_22_linter_symlink_resolve.py) (~180 line, 3 test, 본 release)
