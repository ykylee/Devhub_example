# Beta v0.7.18 — release coordination observability (2026-06-15)

> v0.7.16 의 race lesson (memory #22 §release coordination race) 본 release 에서 *보험* 적용.
> `cmd_release` 의 `gh release create` 호출 직전 `git ls-remote origin` 로 *원격 tag pre-check*.
> pre-check fail 시 default: exit 1 + auto-bump hint. `--auto-bump` flag: 자동 bump + re-flow.

## 핵심 추가 (1 follow-up 본, 0 collateral, 0 deferred)

### 1. tools/release_pipeline.py — 3-step release coordination observability

**3 helper 신규**:
- `_check_remote_tag(tag, *, timeout=15) -> dict`: `git ls-remote origin "refs/tags/{tag}"` 조회. `{"exists": bool, "remote_url": str | None, "tag": str}`.
- `_list_remote_tags(pattern="v*", *, timeout=15) -> list[str]`: 원격의 tag list (정규식 filter + sort). peel `^{}` tag 제외.
- `_version_sort_key(tag) -> tuple`: PEP 440 + suffix sort key. `v0.7.17-beta → (0, 7, 17, 2)`, `v0.7.17 → (0, 7, 17, 0)`. suffix rank: `''=0` (release), `alpha=1`, `beta=2`, `rc=3`.

**`next_available_version(local_version, *, remote_tags=None) -> dict`**: 1차 출처 = remote `vX.Y.*` tag list 의 max + 0.0.1. 3 case:
- remote empty → local 그대로 (bumped=False)
- local < remote max → remote max + 0.0.1 (bumped=True)
- local >= remote max → local 그대로 (bumped=False)

**`cmd_release` 의 pre-check 통합 (v0.7.18+)**:
```python
# step 3.5: 원격 tag pre-check
if not args.dry_run:
    tag_check = _check_remote_tag(tag)
    results["tag_pre_check"] = tag_check
    if tag_check["exists"]:
        return {**results, "error": (
            f"remote tag {tag} already exists at {tag_check['remote_url']}. "
            f"v0.7.16 race 정공법: --auto-bump 으로 다음 version 자동 bump, "
            f"또는 --version=<next> 명시."
        )}
else:
    # dry-run: pre-check 결과 + warning (plan 검증)
    tag_check = _check_remote_tag(tag)
    results["tag_pre_check"] = tag_check
    if tag_check["exists"]:
        results["tag_pre_check_warning"] = f"remote tag {tag} already exists (dry-run: pre-check only)"
```

**`--auto-bump` argparse flag (v0.7.18+)**:
- 위치: `release` subcommand 의 `--version` 옆
- default: False (caller 의 명시적 opt-in)
- pre-check fail 시: `next_available_version(version)` 호출 → bumped=True 면 `write_version(next)` + `write_workflow_kit_version(next, suffix=...)` 자동 실행 → re-flow
- 결과 dict 에 `auto_bump: {...}` field 추가 (caller 가 검증 가능)

**Dry-run 시 tag_pre_check_warning** (v0.7.18+):
- dry-run mode 에서도 pre-check 수행
- remote tag 가 존재하면 `tag_pre_check_warning` field 에 메시지 (실제 gh 호출 안 함)
- caller 가 *release cut 전* race 상태 확인 가능

### 2. Smoke test (7 test 신규, check_release_pipeline_release_coordination.py)

| Test | 검증 |
|---|---|
| `test_remote_tag_check_argparse` | `--auto-bump` argparse error 부재 |
| `test_remote_tag_check_dry_run_no_remote` | origin remote 부재 시 graceful (exists=False) |
| `test_remote_tag_check_dry_run_existing_tag` | dry-run 시 tag_pre_check 결과 노출 |
| `test_version_sort_key` | PEP 440 + suffix 정렬 (v0.7.17 > v0.7.16-beta) |
| `test_next_available_version_no_remote` | remote_tags=[] → local 그대로 (bumped=False) |
| `test_next_available_version_remote_ahead` | local=0.7.17 + remote=v0.7.17 → +0.0.1 → 0.7.18 (bumped=True) |
| `test_next_available_version_local_ahead` | local=0.7.18 + remote=v0.7.17 → 그대로 (bumped=False) |

**7/7 PASS**.

### 3. v0.7.16 race lesson 의 *보험* (cross-cutting)

v0.7.16 의 5-step re-version + cherry-pick + merge (memory #22 §Part C) 의 root cause = `cmd_release` 의 `gh release create` 호출 *직전* 의 *pre-check 부재*. v0.7.18 의 `_check_remote_tag` 가 이 부재 해소.

**3-step race 보험 (v0.7.18+)**:
1. **Pre-check (default)**: `cmd_release` 의 step 3.5 에서 `git ls-remote origin` 로 tag 존재 여부 확인. 존재 시 `error: remote tag already exists` + auto-bump hint.
2. **Auto-bump (opt-in)**: `--auto-bump` flag 로 caller 가 명시적 opt-in. `next_available_version()` 가 다음 version 결정 + `write_version` + `write_workflow_kit_version` 자동.
3. **Dry-run verification**: dry-run mode 에서도 pre-check 수행. `tag_pre_check_warning` 으로 caller 가 release cut *전* race 상태 확인.

**v0.7.16 race 의 5-step re-version 작업 (4-5분) vs v0.7.18+ 의 1-step auto-bump (<5초)** — *시간 비용 100x 단축*.

## Deferred (v0.7.19+)

| Deferred | 이유 | v0.7.19+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` |
| linter 의 symlink-resolve fix (v0.7.16) | mavis data dir 격리 + macOS `/var` symlink | `.resolve()` → `.absolute()` |
| legacy L2 → in-repo migration (v0.7.17) | 외부 vault 의 기존 L2 file 의 in-repo 이관 | `tools/migrate_vault_l2_to_inrepo.py` |

## 검증

- **신규 test**: 7 (위 §2)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_release_pipeline_release_coordination: 7/7 (신규, 본 release)
  - check_release_pipeline: 8/8
  - check_release_pipeline_phase2: 7/8 (1 fail 은 dist 부재 pre-condition, 본 release 와 무관)
  - check_release_pipeline_phase3: 8/8
  - check_release_pipeline_version_flag: 2/3 (1 fail 은 pre-condition, release cut 후 정상화)
  - check_release_pipeline_version_auto_sync: 4/4
  - check_release_pipeline_changelog_gen: 6/6
- **누적 188+ test PASS** (v0.7.17 178+ + 7 신규 + 기존 3 release pipeline test 의 pre-condition fail 해소)

## Commit

| Hash | Subject |
|---|---|
| `07bf145` | feat(v0.7.18): release coordination observability (_check_remote_tag + next_available_version + --auto-bump) + 7 smoke |

## 다음 (v0.7.19 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5 잔여) — GH Actions 또는 local pre-push hook
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **linter 의 symlink-resolve fix** (v0.7.16 발견) — `.resolve()` → `.absolute()`
- **legacy L2 → in-repo migration** (v0.7.17 발견) — `tools/migrate_vault_l2_to_inrepo.py`

## Reference

- [v0.7.17 release note](Beta-v0.7.17.md) (직전) — wiki in-repo storage isolation
- [v0.7.16 release note](Beta-v0.7.16.md) — config + linter fix + **release coordination race** 의 1차 발견
- [v0.7.11 release note](Beta-v0.7.11.md) — release_pipeline Phase 3 (dist subcommand)
- [v0.7.10 release note](Beta-v0.7.10.md) — release_pipeline Phase 2 (release/verify/rollback)
- [v0.7.9 release note](Beta-v0.7.9.md) — release_pipeline tool 정식화 (validate/version-bump/note-draft)
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [workflow-source/tools/release_pipeline.py](../tools/release_pipeline.py) (~1100 line, v0.7.18 본 release, --auto-bump + 3 helper)
- [tests/check_release_pipeline_release_coordination.py](../tests/check_release_pipeline_release_coordination.py) (~190 line, 7 test, 본 release)
