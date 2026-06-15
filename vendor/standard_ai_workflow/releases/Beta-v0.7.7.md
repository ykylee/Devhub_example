# Beta v0.7.7 — Doctor × Metadata Integration (2026-06-14)

> v0.7.6 의 [tool.workflow-doctor] metadata 의 *1차 consumer* 통합.
> workflow_kit.cli.doctor (v0.7.4) 의 hardcoded CI threshold → 외부화 config 기반.

## 핵심 추가 (1 follow-up, 2 deferred)

### workflow_kit.cli.doctor × load_config() integration

v0.7.6 의 `pyproject.toml [tool.workflow-doctor]` 5 field 가 doctor.py 에서 *실제 사용* 되도록 integration. v0.7.4 의 `--exit-on-fail` 가 `non_compliant` 하드코딩 → v0.7.7 의 *enum 기반 threshold* (compliant | advisory | non_compliant).

```bash
# 기본 (config 부재 시 default fail_on=non_compliant)
python -m workflow_kit.cli.doctor --exit-on-fail

# custom config — fail_on=advisory
# (project 의 pyproject.toml)
[tool.workflow-doctor]
fail_on = "advisory"
python -m workflow_kit.cli.doctor --exit-on-fail  # exit 1 if advisory

# config dump (--show-config)
python -m workflow_kit.cli.doctor --show-config  # DoctorConfig.to_dict() JSON

# JSON output — config + results 둘 다
python -m workflow_kit.cli.doctor --json  # {config: {...}, results: {...}}
```

**변경 (workflow_kit/cli/doctor.py, 172 → 224 line)**:
- `load_config(project_root)` 호출 — pyproject.toml [tool.workflow-doctor] section 자동 load
- `should_fail(status, config)` integration — `--exit-on-fail` 의 *enum* 기반 exit code
- `--show-config` flag: load 된 config `to_dict()` JSON 출력
- `--json` output 구조 변경: `{config: ..., results: ...}` — config field 추가
- `render_pretty` footer: `Config: fail_on=...` + `thresholds:` + `excluded_paths:` 표시
- `config.partial_rules[baseline]` / `config.opt_in[baseline]` 가 있으면 baseline 별 `config partial: ...` / `config opt_in: ...` 표시

**Smoke test (8 test, 207 line)**:
- `test_show_config_outputs_doctor_config`: `--show-config` → DoctorConfig.to_dict() (5 field)
- `test_json_output_includes_config`: `--json` → `{config, results}` 구조
- `test_pretty_renders_config_footer`: `Config: fail_on=...` footer 표시
- `test_exit_on_fail_non_compliant_threshold`: fail_on=non_compliant (default) + 실제 non_compliant → exit 1
- `test_exit_on_fail_advisory_threshold_via_config`: fail_on=advisory (custom config) → exit 0
- `test_exit_on_fail_compliant_threshold_via_config`: fail_on=compliant + status=compliant → exit 1
- `test_pretty_renders_partial_rules_and_opt_in`: config.partial_rules / opt_in → `config partial:` / `config opt_in:` 표시
- `test_baseline_single_evaluation`: `--baseline=resiliency` → 단일 baseline 만

## Deferred (v0.7.8+ follow-up)

v0.7.7 의 본 integration 은 `fail_on` enum + display only. *실제 apply* 의 4종은 follow-up:

| Deferred | 이유 | v0.7.8+ follow-up |
|---|---|---|
| `evaluate_compliance` 의 state-aware variant | `_eval_*_baseline(state)` signature 변경 필요 (breaking change 회피) | state.json read 시 config.partial_rules / opt_in 자동 merge |
| `thresholds.score_alert` 의 score trend integration | `tools/score_wiki_trend.py` 가 hardcoded `0.3` | load_config() 의 `thresholds["score_alert"]` 사용 |
| `thresholds.memory_alert_mb` 의 profiling integration | `workflow_kit.common.profiling` 가 hardcoded `100` | load_config() 의 `thresholds["memory_alert_mb"]` 사용 |
| `excluded_paths` 의 lint skip | `workflow_kit.common.linter` 가 path glob 부재 | `excluded_paths` glob match → lint skip |

trigger: v0.7.6 release note 의 *"다음"* 항목 ("workflow_kit.cli.doctor integration") 의 즉시 후속.
fix: state-aware variant 가 breaking change 라서 v0.7.7 은 *display + fail_on* 만. 1차 source (state.json) 가 우선, config 는 fallback + summary 표시.

## 검증

- **신규 test**: 8 (cli_doctor)
- **회귀 test**: 0 (5 check / 54 test PASS via run_all_checks)
  - check_baselines_compliance: 16/16
  - check_refresh_wiki_memory: 10/10
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
- **누적 103+ test PASS** (v0.7.6 95+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `022672f` | feat(v0.7.7): workflow_kit.cli.doctor 에 load_config + should_fail integration |
| `3300e73` | chore(v0.7.7): version bump 0.7.6 → 0.7.7 + release note |

## 다음 (v0.7.8 / v0.8.0 후보)

- **state-aware evaluate_compliance** — `evaluate_with_state_override(state_overrides)` variant (v0.7.7 의 deferred #1)
- **score trend 의 config thresholds** — `tools/score_wiki_trend.py` 의 `--alert --threshold` 가 load_config().thresholds["score_alert"] 사용 (deferred #2)
- **profiling 의 config memory threshold** — `workflow_kit.common.profiling` 의 memory_alert 가 load_config().thresholds["memory_alert_mb"] 사용 (deferred #3)
- **linter 의 config excluded_paths** — `workflow_kit.common.linter` 가 path glob match → lint skip (deferred #4)
- **Release pipeline 정식화** — workflow doctor 의 release validator hook + PyPI 자동 publish + GH release note 자동 generate
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합 (`scripts/release_post.sh`)

## Reference

- [v0.7.6 release note](Beta-v0.7.6.md) (직전) — metadata module + pyproject.toml section
- [v0.7.5 release note](Beta-v0.7.5.md) — refresh_wiki_memory tool 정식화
- [v0.7.4 release note](Beta-v0.7.4.md) — CLI wrapper (`workflow doctor`) 의 v0.7.7 integration base
- [workflow_kit/cli/doctor.py](../workflow_kit/cli/doctor.py) (224 line, v0.7.7 본 release)
- [workflow_kit/common/metadata.py](../workflow_kit/common/metadata.py) (v0.7.6, 1차 출처)
- [tests/check_cli_doctor.py](../tests/check_cli_doctor.py) (8 test, 본 release)
