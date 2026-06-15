# Beta v0.7.8 — State-Aware Baselines + Config Actual Apply (2026-06-14)

> v0.7.7 의 deferred #1 (state-aware evaluate_compliance) 즉시 후속.
> baselines.py 의 7 _eval_* / 2 public evaluate_compliance / evaluate_all 에 state override + doctor.py 의 config actual apply.

## 핵심 추가 (1 follow-up 본, 3 deferred)

### state-aware evaluate_compliance + config actual apply

v0.7.7 의 doctor.py *display only* ("config partial: ... / config opt_in: ... footer") 가 v0.7.8 에서 *actual apply* 로 격상. caller 가 in-memory state dict 를 주입 → partial_rules / opt_in 자동 merge → evaluate_compliance 의 *진짜* partial mode 동작.

```bash
# project 의 state.json 부재 시에도 config.partial_rules 가 hard constraint 로 동작
# pyproject.toml:
[tool.workflow-doctor]
partial_rules = { resiliency = ["RES-WF-01", "RES-WF-02"] }
opt_in = { "security-auth" = ["SEC-AUTH-04"] }

# doctor.py 가 config → state merge → evaluate_compliance(state=merged_state) 호출
python -m workflow_kit.cli.doctor --baseline=resiliency --json
# → cs.partial_rules = ["RES-WF-01", "RES-WF-02"] (config 의 partial_rules)
# → cs.partial_rules union ["SEC-AUTH-04"] (config.opt_in 도)
```

**변경 (workflow_kit/common/contracts/baselines.py)**:
- 7 `_eval_*(project_root, *, state: dict | None = None)` — keyword-only state arg
- `state = state if state is not None else _read_state_json(project_root)` — 7 곳 patch
- `evaluate_compliance(project_root, baseline, fn=None, baseline_path=None, *, state=None)`
- `evaluate_all(project_root, fn=None, baseline_path=None, *, state=None)`
- **backward compat 100%**: state=None (default) → _read_state_json 자동 호출, 기존 positional caller 영향 0
- **breaking change 0**: keyword-only + default None

**변경 (workflow_kit/cli/doctor.py, 224 → 273 line)**:
- `evaluate(project_root, baseline, config=None)` — config 인자 추가
- config 있으면 `_read_state_json(project_root)` → in-memory state dict → **merge**:
  - `config.partial_rules[baseline]` → `state[f"{baseline}_baseline"]["partial_rules"]` union
  - `config.opt_in[baseline]` → `state[f"{baseline}_baseline"]["status"] = "enabled"` + opt_in rule 도 partial_rules union
- **key 정규화 fix** (find): config 의 hyphen (`'security-auth'`) vs baselines.py 의 underscore (`'security_auth'`) 불일치. `_normalize_key(bl_name)` 가 hyphen→underscore 변환 후 `_get_partial_rules(state, "security_auth")` 가 `state["security_auth_baseline"]` 정확히 찾음
- `evaluate_compliance(state=merged_state)` / `evaluate_all(state=merged_state)` 호출

**Smoke test (8 test, 248 line)**:
- `test_evaluate_compliance_state_kwarg`: keyword arg + state 명시 → partial_rules 반영
- `test_evaluate_all_state_kwarg`: evaluate_all(state=...) 정상
- `test_eval_resiliency_state_override`: _eval_*(state=...) 의 partial_rules 반영
- `test_backward_compat_state_none`: state=None → _read_state_json 자동 호출
- `test_positional_caller_breaking_change_0`: positional caller (fn, baseline_path) 영향 0
- `test_doctor_config_partial_rules_applied`: config.partial_rules → actual apply
- `test_doctor_config_opt_in_applied`: config.opt_in → state.status=enabled + partial_rules union
- `test_doctor_config_partial_rules_and_opt_in_union`: partial_rules + opt_in 동시 union (dedup)

## Deferred (v0.7.9+ follow-up)

v0.7.8 의 본 integration 은 **state-aware variant + partial_rules / opt_in actual apply**. 나머지 deferred 3종:

| Deferred | 이유 | v0.7.9+ follow-up |
|---|---|---|
| `thresholds.score_alert` 의 score trend integration | `tools/score_wiki_trend.py` 가 hardcoded `0.3` | `load_config().thresholds["score_alert"]` 사용 |
| `thresholds.memory_alert_mb` 의 profiling integration | `workflow_kit.common.profiling` 가 hardcoded `100` | `load_config().thresholds["memory_alert_mb"]` 사용 |
| `excluded_paths` 의 lint skip | `workflow_kit.common.linter` 가 path glob 부재 | `excluded_paths` glob match → lint skip |

trigger: v0.7.7 release note 의 *"다음"* 항목 ("state-aware evaluate_compliance") 의 즉시 후속.

## 검증

- **신규 test**: 8 (state_aware_baselines)
- **회귀 test**: 0 (6 check / 62 test PASS via run_all_checks)
  - check_baselines_compliance: 16/16
  - check_refresh_wiki_memory: 10/10
  - check_run_all_checks: 10/10
  - check_metadata: 10/10
  - check_cli_doctor: 8/8
  - check_state_aware_baselines: 8/8 (본 release)
- **누적 111+ test PASS** (v0.7.7 103+ + 8 신규)

## Commit

| Hash | Subject |
|---|---|
| `d3235ad` | feat(v0.7.8): state-aware evaluate_compliance + config actual apply |
| `b67af83` | chore(v0.7.8): version bump 0.7.7 → 0.7.8 + release note |

## 다음 (v0.7.9 / v0.8.0 후보)

- **score trend 의 config thresholds** (v0.7.7 의 deferred #2) — `tools/score_wiki_trend.py` hardcoded 0.3 → `thresholds["score_alert"]`
- **profiling 의 config memory threshold** (v0.7.7 의 deferred #3) — `workflow_kit.common.profiling` hardcoded 100 → `thresholds["memory_alert_mb"]`
- **linter 의 config excluded_paths** (v0.7.7 의 deferred #4) — path glob match → lint skip
- **Release pipeline 정식화** — workflow doctor 의 release validator hook + PyPI 자동 publish + GH release note 자동 generate
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합 (`scripts/release_post.sh`)

## Reference

- [v0.7.7 release note](Beta-v0.7.7.md) (직전) — display only state-aware deferred
- [v0.7.6 release note](Beta-v0.7.6.md) — `load_config` / `DoctorConfig` 의 1차 source
- [v0.7.5 release note](Beta-v0.7.5.md) — `refresh_wiki_memory` tool 정식화
- [v0.7.4 release note](Beta-v0.7.4.md) — CLI wrapper (`workflow doctor`) 의 integration base
- [workflow_kit/common/contracts/baselines.py](../workflow_kit/common/contracts/baselines.py) (696+ line, v0.7.8 본 release)
- [workflow_kit/cli/doctor.py](../workflow_kit/cli/doctor.py) (273 line, v0.7.8 본 release)
- [tests/check_state_aware_baselines.py](../tests/check_state_aware_baselines.py) (248 line, 8 test, 본 release)
