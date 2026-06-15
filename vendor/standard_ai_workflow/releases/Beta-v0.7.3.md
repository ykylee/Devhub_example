# Beta v0.7.3 — Extension Runtime Helper 본 구현 (2026-06-13)

> v0.7.2 의 sub-cat 본 구현 + v0.7.3 의 4종 runtime helper 본 구현으로
> **7 baseline dispatcher (5 baseline × 평균 6.4 rule = 44 RuleResult)** 완성.

## 핵심 추가 (4 runtime helper + 4 dispatcher)

### 4 Runtime Helper (workflow_kit/common/{auth,testing,profiling,resiliency}.py)

| Helper | Rule | Evaluator |
|---|---|---|
| `auth.py` | 6 SEC-AUTH | macOS keyring / token rotation / OAuth scope / 2FA / entropy / audit log |
| `testing.py` | 6 PBT-WF | AST-based property test 명명 / round-trip / invariant / idempotency / generator / shrink |
| `profiling.py` | 6 PERF-MEM | tracemalloc peak / leak detect / GC pause / ref cycle / cProfile / regression |
| `resiliency.py` | 8 RES-WF | doctor / structured log / metrics / run_id / 5-tuple error / SIGTERM / max iter / health snapshot |

### 4 신규 Baseline Dispatcher (contracts/baselines.py)

```python
from workflow_kit.common.contracts.baselines import evaluate_all
result = evaluate_all(project_root)  # 7 baseline 모두 평가
# {
#   "security": ComplianceSummary(... 6 rule),
#   "testing": ComplianceSummary(... 6 rule),
#   "performance": ComplianceSummary(... 6 rule),
#   "security-auth": ComplianceSummary(... 6 rule),
#   "testing-property-based": ComplianceSummary(... 6 rule),
#   "performance-memory": ComplianceSummary(... 6 rule),
#   "resiliency": ComplianceSummary(... 8 rule),
# }
```

## 회귀 / 신규 Test

- **신규 12 test**: check_v0_7_3_runtime_helpers.py (4 helper × 3 = 12)
- **기존 12 test (check_baselines_compliance.py)** v0.7.3 dispatcher 갱신: 18 → 44 RuleResult, 3 → 7 baseline
- **기존 30 test (check_extension_system.py)** 변경 없음
- **누적 191 test PASS** (v0.7.2 179 + 12 신규) — 회귀 0

## v0.7.2 → v0.7.3 의 차이

| 영역 | v0.7.2 | v0.7.3 |
|---|---|---|
| sub-cat doc | ✅ 26 rule 본 구현 | ✅ (그대로) |
| runtime helper | ❌ spec 만 | ✅ 4 helper 본 구현 |
| baseline dispatcher | 3 (security/testing/performance) | 7 (위 + 4 sub-cat) |
| RuleResult | 18 (3×6) | 44 (6×6 + 8) |
| 4 helper 자체 test | 0 | 12 |

## 4 Helper 의 Design Detail

### auth.py (SEC-AUTH)
- **chmod 600** 검증: 3 well-known secret path (myharness / aws / gh)
- **git history leak**: `git log -p --all` 에서 API key pattern (sk-, ghp_, glpat-) grep
- **entropy**: Shannon entropy × length, 128 bit threshold
- **OAuth scope**: ≤ 5개, 광범위 scope (admin, *:all) 거부

### testing.py (PBT-WF)
- **AST-based**: `ast.parse()` + `ast.FunctionDef.name` walk
- **decorator detection**: `@pytest.mark.parametrize` + `@given` (hypothesis)
- **5 pattern search**: round_trip / encode_decode / invariant / idempotent / minimal + @example

### profiling.py (PERF-MEM)
- **tracemalloc + RSS hybrid**: heap peak + RSS max
- **200 MB threshold**: macOS resource.RUSAGE_SELF 가 byte
- **leak detect**: 10 회 반복 후 RSS growth < 5 MB
- **GC pause ratio**: collected object count × 10 µs / total elapsed ≤ 5%
- **regression**: baseline.json 대비 ±10%

### resiliency.py (RES-WF)
- **doctor detection**: tools/doctor.py / workflow_kit/cli/doctor.py
- **structured log**: json.dumps + extra={} / key=value
- **5-tuple error**: error_type / message / cause / location / remediation 4+ 개
- **SIGTERM handler**: signal.signal(SIGINT|SIGTERM) regex
- **resource limit**: max_iter / max_time / MAX_ITER 상수

## Cross-cutting 적용

- v0.7.2 의 26 rule spec 이 **runtime 으로 enforce 가능**해짐
- CI 통합: `evaluate_all(project_root)` → JSON → alert/hook
- partial mode: state.json 의 `partial_rules` 로 soft rule 지정 가능 (v0.7.1+ 정책 유지)
- 4 helper 는 *독립 import 가능* (`from workflow_kit.common.auth import evaluate_compliance`)

## v0.7.4+ follow-up (open)

1. v0.7.4: 4 helper 의 `--json` / `--pretty` CLI wrapper (workflow doctor)
2. v0.7.4: profiling 의 baseline 자동 capture (`.memory_baseline.json` commit + auto-update)
3. v0.7.4: resiliency 의 SIGTERM handler 등록 boilerplate (`@graceful_shutdown` decorator)
4. v0.7.4: PBT hypothesis + memory objgraph 의존성 옵션 (optional requirement)
5. v0.8.0: flat path migration (v0.7.0 flat → v0.7.2+ sub-cat 1 release 후 deprecate)

## 호환성

- Python 3.10+ (macOS / Linux / Windows)
- 외부 의존성 0 (stdlib only: `ast`, `gc`, `tracemalloc`, `resource`, `json`, `re`)
- 기존 3 baseline dispatcher API 변경 없음 (backward compat)
- 4 신규 dispatcher 는 `evaluate_compliance(baseline="security-auth" | ...)` 로 명시
