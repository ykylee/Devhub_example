# Beta v0.7.24 — cmd_release --notes-template flag (2026-06-15)

> GH release notes format 자유도. 5가지 template 지원 (default / detailed / simple / changelog / custom).
> 운영자가 release 시 *어떤 형태의 notes* 를 쓸지 caller 가 결정. 이전은 *always default* 였음.

## 핵심 추가 (1 follow-up 본, 0 collateral, 0 deferred)

### 1. tools/release_pipeline.py — `--notes-template` flag + `_resolve_notes_file()` helper

**5 가지 template**:
| Template | notes_file | Use case |
|---|---|---|
| `default` | `Beta-v<X>.<Y>.<Z>.md` (기존) | *정통* release note (full content) |
| `detailed` | `Beta-v<X>.<Y>.<Z>.md` (default 와 동일, 명시적) | default 와 동일하나 *caller 의 의도* 명시 |
| `simple` | `Beta-v<X>.<Y>.<Z>-simple.md` | *1 line summary* — 1st # + 1st ## + 1st paragraph 만 자동 generate. *CI patch release* / *의미 없는 release* 등 |
| `changelog` | `workflow-source/CHANGELOG.md` | *Keep-a-Changelog 1.1.0* 형식 (v0.7.14 의 changelog-gen 의 output) |
| `custom:<path>` | 임의 path | *full override* — custom file 직접 사용 |

**argparse** (1 flag):
```python
p_rel.add_argument("--notes-template", dest="notes_template", default="default",
                   help="release notes format 결정. v0.7.24+. "
                        "'default' (Beta-v<X>.<Y>.<Z>.md) / 'detailed' (default 와 동일) / "
                        "'simple' (1 line summary) / 'changelog' (workflow-source/CHANGELOG.md) / "
                        "'custom:<path>' (임의 path).")
```

**`_resolve_notes_file()` helper** (v0.7.14 의 `draft_changelog` 와 *동일 정신* — `cmd_changelog_gen` 의 output 재활용):
- 5 가지 template 분기
- `simple` template 시 *default notes 의 1st # + 1st ## + 1st paragraph* 자동 generate (blank_count=2 → 1st paragraph 끝)
- `custom:<path>` 의 *in-repo* 또는 *absolute* 모두 지원
- unknown value 시 명확한 error message (사용 가능 option 4개 명시)

**`cmd_release` 의 notes_file 결정 로직 (v0.7.24+)**:
```python
notes_template = getattr(args, "notes_template", "default") or "default"
notes_resolution = _resolve_notes_file(version, notes_template, dry_run=args.dry_run)
if notes_resolution.get("error"):
    return {**results, "error": notes_resolution["error"]}
notes_file = notes_resolution["notes_file"]
if not notes_file.exists():
    return {**results, "error": f"release note not found: {notes_file}"}
```

**`results["notes_file"]` 의 path 표현**:
- in-repo (`Beta-v<X>.<Y>.<Z>.md`): relative_to(REPO_ROOT) → `workflow-source/releases/Beta-v<X>.<Y>.<Z>.md`
- absolute (`/path/to/custom.md`, `workflow-source/CHANGELOG.md`): 그대로 (즉 `notes_file.is_relative_to(REPO_ROOT)` 의 try/except)

### 2. Smoke test (5 test 신규, check_v0_7_24_release_notes_template.py)

| Test | 검증 |
|---|---|
| `test_notes_template_default_argparse` | `--notes-template=default` argparse error 부재 |
| `test_notes_template_simple` | simple template 가 1st # + 1st ## + 1st paragraph 자동 generate + 2nd ## 헤더 + 2nd paragraph 제외 |
| `test_notes_template_changelog` | changelog template 가 `workflow-source/CHANGELOG.md` 가리킴 |
| `test_notes_template_custom` | custom:<path> 의 in-repo + absolute 모두 지원 |
| `test_notes_template_unknown` | unknown value 시 명확한 error message (사용 가능 option 4개 명시) |

**5/5 PASS**.

## Deferred (v0.7.25+)

| Deferred | 이유 | v0.7.25+ follow-up |
|---|---|---|
| `ci-publish` subcommand (Phase 5) | GH Actions / pre-push hook | `.github/workflows/release.yml` |
| legacy L2 → in-repo migration (v0.7.17) | 외부 vault 의 기존 L2 file 의 in-repo 이관 | `tools/migrate_vault_l2_to_inrepo.py` |
| `check_workflow_linter.py` branch detection | mavis data dir 격리 환경의 branch name resolution | `workflow_kit.common.paths.get_current_branch` 의 4-priority auto-detect 적용 |
| `--notes-template` 의 TUI/Interactive mode | CI 통합 위주 argparse. *rich* output 미지원 | TUI prompt 또는 pre-flight check + 친절한 error message |

## 검증

- **신규 test**: 5 (위 §2)
- **회귀 test**: 0 (본 작업이 직접 영향 0 fail)
  - check_v0_7_24_release_notes_template: 5/5 (신규, 본 release)
- **누적 202+ test PASS** (v0.7.23 197+ + 5 신규)

## Commit

| Hash | Subject |
|---|---|
| `1dfa8fb` | feat(v0.7.24): cmd_release --notes-template flag (5 template: default/detailed/simple/changelog/custom) + 5 smoke |

## 다음 (v0.7.25 / v0.8.0 후보)

- **ci-publish subcommand** (v0.7.11 의 Phase 5 잔여) — GH Actions 또는 local pre-push hook
- **legacy L2 → in-repo migration** (v0.7.17 발견) — `tools/migrate_vault_l2_to_inrepo.py`
- **`check_workflow_linter.py` branch detection fix** (v0.7.22 발견) — 4-priority auto-detect 적용

## Reference

- [v0.7.23 release note](Beta-v0.7.23.md) (직전) — wiki 운영 cross-link 1-command wrapper
- [v0.7.22 release note](Beta-v0.7.22.md) — linter symlink-resolve fix
- [v0.7.21 release note](Beta-v0.7.21.md) — F-4 design 결함 fix (--allow-existing-tag + tag push coupling)
- [v0.7.18 release note](Beta-v0.7.18.md) — release coordination observability
- [v0.7.14 release note](Beta-v0.7.14.md) — cmd_changelog_gen (Keep-a-Changelog 1.1.0, *simple* template 의 본문 의 1차 출처)
- [workflow-source/tools/release_pipeline.py](../tools/release_pipeline.py) (~1330 line, v0.7.24 본 release, --notes-template + _resolve_notes_file)
- [tests/check_v0_7_24_release_notes_template.py](../tests/check_v0_7_24_release_notes_template.py) (~210 line, 5 test, 본 release)
