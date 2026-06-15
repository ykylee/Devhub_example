# Beta v0.7.10 — Release Pipeline Phase 2 (2026-06-14)

> v0.7.9 Phase 1 (validate / version-bump / note-draft) 의 Phase 2.
> manual release 절차 (docs/RELEASE.md) 의 *trigger + verify + rollback* 자동화. 6 subcommand 정합.

## 핵심 추가 (1 follow-up 본, 1 deferred)

### tools/release_pipeline.py Phase 2 (release / verify / rollback)

v0.7.9 의 Phase 1 (사전 점검 + version + note) 가 *3-step pipeline* 의 **input layer** 였다면, v0.7.10 의 Phase 2 는 **execution layer** — gh CLI 통합 + read-only verify + destructive rollback.

**PyPI/TestPyPI 업로드 ❌** (memory #5 의 release 채널 정책 — GitHub Releases 만).

```bash
# dry-run: gh command 만 print, 호출 0
python3 tools/release_pipeline.py release --dry-run --skip-validate
python3 tools/release_pipeline.py verify --tag=v0.7.9-beta --dry-run
python3 tools/release_pipeline.py rollback --tag=v0.7.9-beta --dry-run

# apply (gh auth 인증된 환경 가정)
python3 tools/release_pipeline.py release --apply  # gh release create 호출
python3 tools/release_pipeline.py verify --tag=v0.7.9-beta --json  # gh release view 호출
python3 tools/release_pipeline.py rollback --tag=v0.7.9-beta --apply  # 3 commands 실행
```

**Tool 구조 확장 (430 → 656 line)**:
- `cmd_release` (Phase 2 / 130 line): `gh release create` 호출
  - 사전 점검: `--skip-validate` 미지정 시 validate 4 source 자동 호출. 1+ fail 시 release 중단
  - dist 파일 glob: PEP 440 normalize (0.7.10 → `standard_ai_workflow-0.7.10*` glob)
  - tag + notes_file + dist glob → gh command build (dry-run 시 print, apply 시 실행)
  - gh auth status 자동 검증 (--apply 시). 미인증 시 graceful fail
  - repo remote 자동 추출 (`git remote get-url origin`)
- `cmd_verify` (Phase 2 / 50 line): `gh release view` (read-only)
  - `--json`: tag / name / url / isPrerelease / createdAt / assets / asset_count
  - `--dry-run`: gh_command 만 print. `--apply` 시 gh release view 호출. release 부재 시 graceful fail
- `cmd_rollback` (Phase 2 / 80 line): `gh release delete` + `git tag -d` + `git push --delete` (destructive)
  - 3 commands: local tag delete / remote tag delete / gh release delete --yes
  - `--dry-run`: commands list print. `--apply` 시 순차 실행, 1+ fail 시 중단
  - destructive 작업의 `--dry-run` 명시 (Phase 1 의 `--apply` 정공법과 동일)
- main 분기: validate / release / verify / rollback 의 ok/error 기반 exit code
- `find_dist_files`: PEP 440 normalize helper
- `_get_repo`: git remote origin → 'owner/repo' 자동 추출

**Smoke test (8 test, 224 line, check_release_pipeline_phase2.py)**:
- `test_release_dry_run_no_dist`: dist 부재 시 graceful error + exit 1
- `test_release_skip_validate_dry_run`: `--skip-validate` argparse 정상
- `test_verify_dry_run`: gh_command plan + mode=read-only
- `test_verify_apply_release_not_found`: release 부재 시 exit 1 + error
- `test_rollback_dry_run`: 3 commands (git tag -d / git push --delete / gh release delete) print
- `test_find_dist_files_pep440`: PEP 440 normalize 검증
- `test_cli_phase2_subcommand_help`: 3 Phase 2 subcommand 노출
- `test_release_skip_validate_argparse`: argparse error 없음

## Deferred (v0.7.11+ Phase 3)

| Deferred | 이유 | v0.7.11+ follow-up |
|---|---|---|
| `dist` subcommand | wheel + sdist *자동 빌드* (`python3 -m build` + `twine check`) | `python3 -m build` 호출 + `dist/*.whl` glob 결과 + PEP 440 normalize |
| `ci-publish` subcommand | GH Actions 또는 local pre-push hook | `.github/workflows/release.yml` 자동화 + gh auth token 주입 |
| `changelog-gen` subcommand | CHANGELOG.md (vs release note) 누적 | git log all-time → CHANGELOG.md section 별 분류 |

trigger: v0.7.9 release note 의 deferred 3종 (release / verify / rollback) 의 즉시 후속. v0.7.10 으로 2 of 3 완료 (release / verify), 1 follow-up (Phase 3) 으로 dist / ci-publish / changelog-gen.

## 검증

- **신규 test**: 8 (Phase 2) + Phase 1 회귀 8
- **회귀 test**: 0 (8 check / 78 test PASS via run_all_checks)
  - check_baselines_compliance: 16/16
  - check_refresh_wiki_memory: 10/10
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
  - check_state_aware baselines: 8/8
  - check_release_pipeline: 8/8 (Phase 1)
  - check_release_pipeline_phase2: 8/8 (Phase 2, 본 release)
- **누적 127+ test PASS** (v0.7.9 119+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `fdf8159` | feat(v0.7.10): release_pipeline Phase 2 (release / verify / rollback) + 8 smoke test |
| `67d4a37` | chore(v0.7.10): version bump 0.7.9 → 0.7.10 + release note |

## 다음 (v0.7.11 / v0.8.0 후보)

- **dist subcommand** (v0.7.10 의 Phase 3) — wheel + sdist 자동 빌드 (`python3 -m build`)
- **ci-publish subcommand** — GH Actions 또는 local pre-push hook
- **changelog-gen subcommand** — CHANGELOG.md 누적 (vs release note)
- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` 의 hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (deferred #3)
- **linter 의 config excluded_paths** (deferred #4)
- **Wiki 운영 cross-link** (emit_wiki_l2_body + refresh_wiki_memory 1-command)

## Reference

- [v0.7.9 release note](Beta-v0.7.9.md) (직전) — Phase 1 (validate / version-bump / note-draft)
- [v0.7.8 release note](Beta-v0.7.8.md) — state-aware + actual apply
- [v0.7.7 release note](Beta-v0.7.7.md) — display only + doctor integration
- [v0.7.6 release note](Beta-v0.7.6.md) — `load_config` / `DoctorConfig` 정의
- [docs/RELEASE.md](../../../docs/RELEASE.md) (manual release 절차, 2026-06-12 갱신)
- [memory #5 standard-ai-workflow.md](../../../../../../Users/yklee/.mavis/agents/mavis/memory/standard-ai-workflow.md) (release 채널 정책)
- [tools/release_pipeline.py](../tools/release_pipeline.py) (656 line, v0.7.10 본 release)
- [tests/check_release_pipeline_phase2.py](../tests/check_release_pipeline_phase2.py) (224 line, 8 test, 본 release)
