# Beta v0.7.6 — Test 통합 Runner + Pyproject Metadata (2026-06-14)

> v0.7.5 의 Wiki 운영 자동화의 *테스트 layer + config layer* 보강.
> 77+ check_*.py 통합 runner + pyproject.toml [tool.workflow-doctor] 외부 config.

## 핵심 추가 (2 follow-up)

### 1. run_all_checks.py 통합 runner (D)

workflow-source 의 `tests/check_*.py` **77 file** 의 *운영 layer* 단일 진입점. 매번 `python3 tests/check_X.py` 77번 호출 부담 → 한 번의 `python3 tests/run_all_checks.py` 로 통합.

```bash
# 전체 77 check 실행
python3 tests/run_all_checks.py

# filter (e.g. baselines + wiki 만)
python3 tests/run_all_checks.py --filter=baselines,wiki

# JSON 출력 (CI 통합)
python3 tests/run_all_checks.py --json

# fail-fast (첫 실패 시 중단)
python3 tests/run_all_checks.py --fail-fast

# per-check timeout (default 60s)
python3 tests/run_all_checks.py --timeout=120
```

**Tool 구조 (230 line)**:
- argparse 5 flag: `--tests-dir` / `--filter` / `--fail-fast` / `--timeout` / `--json`
- `discover_checks`: glob + `--filter=<substring>(comma-separated OR)` 매칭
- `run_one`: subprocess + timeout + stdout/stderr 캡처 + duration 측정
- `parse_output`: `All N tests passed.` / `N/M tests failed:` regex parse
- `aggregate`: `CheckResult` list → `RunSummary` (total / passed / failed / total_passed_tests / total_failed_tests / total_duration_sec)
- `print_human` / `print_json`: 시각화 + CI 통합

**Smoke test (10 test, 226 line)**:
- `test_discover_checks_all`: 70+ check file
- `test_discover_checks_filter_substring`: filter substring OR match
- `test_discover_checks_filter_no_match`: 0 file → exit 2
- `test_parse_output_pass` / `test_parse_output_fail`: output regex
- `test_run_one_baselines` / `test_run_one_refresh_wiki`: 실제 check 실행 (16 + 10 = 26 test 검증)
- `test_aggregate`: passed/failed 합계 정합
- `test_cli_json_output`: `--json` exit 0
- `test_cli_no_match_filter_errors`: `--filter=zzz` → exit 2

**fix 발견**: Python 3.14 dataclass + `KW_ONLY` 가 `sys.modules.get(cls.__module__).__dict__` 호출 → importlib 명시 등록 필요. `_import_runner()` 의 `sys.modules["run_all_checks"] = mod` 1 line 추가 (memory #3 의 "dry-run → fix → apply" 패턴 1회 적용).

### 2. workflow_kit.common.metadata (E)

pyproject.toml 의 `[tool.workflow-doctor]` section 을 read 하는 loader. workflow_kit.cli.doctor (v0.7.4) 의 hardcoded threshold / partial_rules / opt-in 을 *외부화*. v0.7.4 decorator + optional dep 의 *config layer* 후속.

**Module 구조 (162 line)**:
- `DoctorConfig` dataclass: 5 field
  - `partial_rules: dict[str, list[str]]` — baseline 별 hard constraint rule
  - `opt_in: dict[str, list[str]]` — baseline 별 opt-in rule (default disable)
  - `thresholds: dict[str, float]` — dim 별 alert (score_alert 0.3 / memory_alert_mb 100)
  - `excluded_paths: list[str]` — lint skip glob (build, .venv, .worktrees 등)
  - `fail_on: str` — CI threshold (compliant | advisory | non_compliant)
- `load_config(project_root=None)`: section 부재 / file 부재 / invalid TOML / type mismatch 모두 default fallback (**운영 안정성** — 절대 실패 안 함)
- `should_fail(status, config)`: severity 비교 (non_compliant=3 > advisory=2 > compliant=1 > not_applicable=0)
- `to_dict()`: JSON-serializable (CI integration)

**pyproject.toml default (5 field)**:
```toml
[tool.workflow-doctor]
partial_rules = { resiliency = ["RES-WF-01", "RES-WF-02"] }
opt_in = { "security-auth" = ["SEC-AUTH-04"] }
thresholds = { score_alert = 0.3, memory_alert_mb = 100.0 }
excluded_paths = ["build/*", ".venv/*", ".venv-build/*", "__pycache__/*", ".worktrees/*"]
fail_on = "non_compliant"
```

**Smoke test (10 test, 233 line)**:
- `test_load_config_no_section` / `test_load_config_no_section_but_file`: default fallback
- `test_load_config_full_section`: 5 field 모두 parse (정합)
- `test_load_config_invalid_toml`: truncated TOML → default
- `test_load_config_type_mismatch`: `partial_rules = "not_a_dict"` → default
- `test_doctor_config_to_dict`: JSON serialize 검증
- `test_should_fail_non_compliant`: fail_on=non_compliant + status=non_compliant → True
- `test_should_fail_advisory_below_threshold`: fail_on=non_compliant + status=advisory → False
- `test_should_fail_advisory_above_threshold`: fail_on=advisory → True
- `test_should_fail_not_applicable`: severity 0 → 어떤 fail_on 이든 False

## 검증

- **신규 test**: 10 (run_all_checks) + 10 (metadata) = **20 신규 test**
- **회귀 test**: 0 (4 check / 46 test PASS, baselines_compliance 16 + refresh_wiki_memory 10 + run_all_checks 10 + metadata 10)
- **누적 95+ test PASS** (v0.7.5 75+ + 20 신규)

## Commit

| Hash | Subject |
|---|---|
| `53d5dc8` | feat(v0.7.6): run_all_checks 통합 runner + 10 smoke test |
| `0daf6da` | feat(v0.7.6): workflow_kit.metadata (pyproject.toml [tool.workflow-doctor] loader) + 10 smoke test |
| `b9ede19` | chore(v0.7.6): version bump 0.7.5 → 0.7.6 + release note |

## 다음 (v0.7.7 / v0.8.0 후보)

- **workflow_kit.cli.doctor integration** — load_config() 의 fail_on + partial_rules + excluded_paths 사용 (v0.7.6 의 metadata 의 1차 consumer)
- **Release pipeline 정식화** — `workflow doctor` 의 release validator hook + PyPI 자동 publish + GH release note 자동 generate
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합 (`scripts/release_post.sh`)
- **Extension 2차 확장** — observability / docs-quality / dependency-audit sub-cat 추가 (v0.7.1-roadmap §1 follow-up)

## Reference

- [v0.7.5 release note](Beta-v0.7.5.md) (직전)
- [v0.7.1-roadmap.md](../extensions/v0.7.1-roadmap.md) §1 sub-cat 도입
- [tools/refresh_wiki_memory.py](../tools/refresh_wiki_memory.py) (v0.7.5, R-3 단계 분리 패턴)
- [tests/run_all_checks.py](../tests/run_all_checks.py) (본 release, 통합 runner)
- [workflow_kit/common/metadata.py](../workflow_kit/common/metadata.py) (본 release, pyproject loader)
- [tests/check_run_all_checks.py](../tests/check_run_all_checks.py) (10 test, 본 release)
- [tests/check_metadata.py](../tests/check_metadata.py) (10 test, 본 release)
