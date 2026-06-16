# Beta v0.7.21 — cmd_release --allow-existing-tag + tag push coupling (2026-06-15)

> v0.7.18 의 release coordination observability 의 **design 결함** fix.
> `_check_remote_tag` 가 *release 시점* 에 tag 존재를 catch 하지만, **tag push 는 *별도 step*** — 정상 flow (tag push → release) 와 충돌.
> 본 release 에서 `--allow-existing-tag` flag + `cmd_release` 의 **tag push 자동화** 로 coupling fix.

## 핵심 추가 (1 follow-up 본, 0 collateral, 0 deferred)

### 1. tools/release_pipeline.py — 3-step coupling fix

**v0.7.18 의 design 결함**:
- `_check_remote_tag` 가 *release 시점* (`cmd_release` 의 step 3.5) 에서 tag 존재 여부 확인
- 정상 flow: `git push origin v0.7.X-beta` → `gh release create --verify-tag` (tag 가 remote 에 *있어야* verify 통과)
- 즉 *tag push* 가 release 의 *선행 step* 인데, *pre-check* 가 *release 시점* 에 *이미 push 된* tag 를 catch → 의도와 다른 fail

**v0.7.21 의 정공법 (3-step coupling)**:
1. **pre-check**: remote tag 존재 시 default = exit 1 + auto-bump hint (v0.7.18 유지)
2. **`--allow-existing-tag` flag** (v0.7.21+): caller 의 명시적 opt-in 으로 pre-check fail skip + 그대로 진행
3. **tag push 자동화** (v0.7.21+): `cmd_release` 가 `git push origin refs/tags/<tag>` 자동 실행 후 `gh release create --verify-tag`. pre-check + tag push + gh release 가 **한 cycle**

**구현**:
```python
# step 3.5: 원격 tag pre-check
if not args.dry_run:
    tag_check = _check_remote_tag(tag)
    results["tag_pre_check"] = tag_check
    if tag_check["exists"]:
        if not getattr(args, "allow_existing_tag", False):
            return {**results, "error": (
                f"remote tag {tag} already exists at {tag_check['remote_url']}. "
                f"v0.7.16 race 정공법: --auto-bump 으로 다음 version 자동 bump, "
                f"--allow-existing-tag 으로 *기존 tag* 에 re-attach, "
                f"또는 --version=<next> 명시."
            )}
        results["tag_pre_check_skipped"] = "allow-existing-tag"

# step 3.6: local tag push (v0.7.21+ coupling)
if not args.dry_run:
    push_tag_proc = subprocess.run(
        ["git", "push", "origin", f"refs/tags/{tag}"],
        cwd=str(REPO_ROOT), capture_output=True, text=True, timeout=30,
    )
    results["tag_push"] = {...}
    if push_tag_proc.returncode != 0 and not getattr(args, "allow_existing_tag", False):
        return {**results, "error": f"git push tag {tag} failed: ..."}
```

**argparse**:
- `--allow-existing-tag` (default False) — caller 명시적 opt-in
- help: "remote tag pre-check 가 'already exists' 일 때 *skip* + 그대로 release 진행. 의도된 tag re-push 또는 backfill 시에만 사용."

### 2. v0.7.18~v0.7.20 의 *auto-bump chain* (회고, 2026-06-15)

본 release 의 trigger = v0.7.18 cut 시 *auto-bump chain* 의 *3 release 의 tag push + pre-check fail cycle*:

| step | version | tag push | pre-check | 결과 |
|---|---|---|---|---|
| v0.7.18 cut | 0.7.18 | ✓ remote push | ✓ catch (이미 v0.7.18-beta 존재) | exit 1 |
| `--auto-bump` | 0.7.19 | ✓ remote push | ✓ catch (이미 v0.7.19-beta 존재) | exit 1 |
| `--auto-bump` | 0.7.20 | ✓ remote push | ✓ catch (이미 v0.7.20-beta 존재) | exit 1 |

3 release 의 pre-check 가 *모두* 의도대로 catch — **F-4 의 본질 (pre-check 가 catch 한 race) 은 *입증***. 다만 *tag push 와 release 의 coupling 부재* 가 *추가 fix* 필요 → v0.7.21.

**Cross-cutting 발견**:
- pre-check + auto-bump 만으로는 *불충분*. tag push 와 release 의 *순서* 가 *분리* 되어 있으면 pre-check 가 *정상 flow* 의 tag push 결과를 *catch* 함.
- **결정**: pre-check + tag push + gh release 가 *한 cycle* 이어야. v0.7.21 의 fix.

### 3. Smoke test (8 test, v0.7.18 의 7 + v0.7.21 의 1)

| Test | 검증 |
|---|---|
| `test_remote_tag_check_argparse` | `--auto-bump` argparse error 부재 (v0.7.18) |
| `test_allow_existing_tag_argparse` (v0.7.21+) | `--allow-existing-tag` argparse error 부재 |
| `test_remote_tag_check_dry_run_no_remote` | origin remote 부재 시 graceful (v0.7.18) |
| `test_remote_tag_check_dry_run_existing_tag` | dry-run 시 tag_pre_check 결과 (v0.7.18) |
| `test_version_sort_key` | PEP 440 + suffix 정렬 (v0.7.18) |
| `test_next_available_version_no_remote` | remote_tags=[] → local 그대로 (v0.7.18) |
| `test_next_available_version_remote_ahead` | local < remote max → +0.0.1 (v0.7.18) |
| `test_next_available_version_local_ahead` | local > remote max → 그대로 (v0.7.18) |

**8/8 PASS**.

## Deferred (v0.7.22+)

| Deferred | 이유 | v0.7.22+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| Wiki 운영 cross-link | `emit_wiki_l2_body.py` + `refresh_wiki_memory.py` 1-command 통합 | `tools/wiki_emit.py` wrapper |
| `cmd_release --notes-template` | GH release notes custom template | argparse `--notes-template` |
| linter 의 symlink-resolve fix (v0.7.16) | mavis data dir 격리 + macOS `/var` symlink | `.resolve()` → `.absolute()` |
| legacy L2 → in-repo migration (v0.7.17) | 외부 vault 의 기존 L2 file 의 in-repo 이관 | `tools/migrate_vault_l2_to_inrepo.py` |

## 검증

- **신규 test**: 1 (`test_allow_existing_tag_argparse`)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_release_pipeline_release_coordination: 8/8 (7 기존 + 1 신규, 본 release)
- **누적 189+ test PASS** (v0.7.20 의 188+ + 1 신규)

## Commit

| Hash | Subject |
|---|---|
| `0ef97db` | fix(v0.7.21): cmd_release --allow-existing-tag flag + tag push 자동화 (pre-check + release coupling) |

## 다음 (v0.7.22 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5 잔여) — GH Actions 또는 local pre-push hook
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합
- **`cmd_release` 의 `--notes-template`** — GH release notes custom template
- **linter 의 symlink-resolve fix** (v0.7.16 발견) — `.resolve()` → `.absolute()`
- **legacy L2 → in-repo migration** (v0.7.17 발견) — `tools/migrate_vault_l2_to_inrepo.py`
- **3 legacy tag cleanup** (v0.7.21 발견) — v0.7.18/v0.7.19/v0.7.20 의 *auto-bump chain* 의 tag 가 remote 에 *고아* 로 존재. *re-push* 또는 *delete* 결정.

## Reference

- [v0.7.20 release note](Beta-v0.7.20.md) (직전) — release coordination observability + auto-bump chain
- [v0.7.18 release note](Beta-v0.7.18.md) — release coordination observability 의 본질
- [v0.7.17 release note](Beta-v0.7.17.md) — wiki in-repo storage isolation
- [v0.7.16 release note](Beta-v0.7.16.md) — config + linter fix + **race 발견의 1차 출처**
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차)
- [workflow-source/tools/release_pipeline.py](../tools/release_pipeline.py) (~1100 line, v0.7.21 본 release, --allow-existing-tag + tag push coupling)
- [tests/check_release_pipeline_release_coordination.py](../tests/check_release_pipeline_release_coordination.py) (~200 line, 8 test, 본 release)
- memory #22 §Part C (release coordination race 정공법)
