# Beta v0.7.4 — CLI Wrapper + Decorator + Optional Dep (2026-06-13)

> v0.7.3 의 4 runtime helper 에 *사용자 진입점* 추가.
> CLI `workflow doctor`, `@graceful_shutdown` decorator, optional `hypothesis`/`objgraph`.

## 핵심 추가 (3 follow-up)

### 1. workflow_kit/cli/doctor.py (CLI wrapper)

```bash
# 7 baseline 모두 평가 (pretty table)
python -m workflow_kit.cli.doctor --project-root .

# 단일 baseline
python -m workflow_kit.cli.doctor --baseline=resiliency --project-root .

# JSON 출력 (CI 통합)
python -m workflow_kit.cli.doctor --json --project-root .

# non_compliant 발견 시 exit 1
python -m workflow_kit.cli.doctor --exit-on-fail --project-root .
```

**Sample output**:
```
=================================================================
 Workflow Doctor — 7 Baseline Compliance Report
=================================================================

[△] security: advisory
   rule_id        title                                    status
   ----------------------------------------------------------------------
   ✓ SEC-WF-01     Audit Log Append-Only + ISO 8601         compliant
   △ SEC-WF-02     Stage Gate Approval Mandatory           advisory
   ...
[△] resiliency: advisory
   ✓ RES-WF-01     Health Check                             compliant
   ...

=================================================================
 Summary: 44 rule total, 12 compliant, 18 advisory, 1 non_compliant
=================================================================
```

### 2. workflow_kit/common/decorators.py (@graceful_shutdown + @v0_7_4_deprecated)

```python
from workflow_kit.common.decorators import graceful_shutdown, v0_7_4_deprecated

# SIGINT / SIGTERM handler 자동 등록 + cleanup
@graceful_shutdown(cleanup=lambda: print("cleanup done"))
def long_running_task():
    ...

# 또는 단순 decoration
@graceful_shutdown
def main():
    ...

# deprecation 표시
@v0_7_4_deprecated(reason="use new_function instead", version="0.8.0")
def old_function():
    ...
```

**@graceful_shutdown 동작**:
- SIGINT (Ctrl-C) / SIGTERM / SIGBREAK (Windows) handler 자동 등록
- cleanup callback 호출 후 sys.exit(0)
- 5초 timeout 후 강제 종료 (warning log)
- @functools.wraps 로 메타데이터 보존

### 3. Optional Dependency (hypothesis / objgraph)

**pyproject.toml**:
```toml
[project.optional-dependencies]
pbt = ["hypothesis>=6.0"]        # PBT framework
profiling = ["objgraph>=3.5"]    # reference chain visualization
```

**Install**:
```bash
pip install standard_ai_workflow[pbt]      # hypothesis 추가
pip install standard_ai_workflow[profiling] # objgraph 추가
```

**Runtime behavior**:
- `testing.py` 의 `check_generator_present`: hypothesis 미설치 시 fallback (minimal generator 또는 advisory)
- `profiling.py` 의 `check_reference_cycle`: objgraph 미설치 시 fallback (basic gc collect)
- 미설치 시 status = `advisory` + `pip install` 힌트

## 회귀 / 신규 Test

- **신규 9 test**: check_v0_7_4_followup.py (3 CLI + 3 decorator + 3 optional dep)
- **기존 test 그대로 PASS**: 12 check_baselines_compliance + 12 check_v0_7_3_runtime_helpers + 30 check_extension_system
- **누적 200 test PASS** (v0.7.3 191 + 9 신규) — 회귀 0

## Bug Fix (v0.7.3 → v0.7.4)

- **baselines.py line 434 `import time` 제거**: top-level `import time` (line 36) 가 이미 있어 local `import time` 이 UnboundLocalError 유발 (PERF-WF-02 의 `time.time()`). v0.7.3 의 doctor CLI 가 fail 했던 원인. v0.7.4 fix.

## v0.7.3 → v0.7.4 의 차이

| 영역 | v0.7.3 | v0.7.4 |
|---|---|---|
| runtime helper (4) | ✅ | ✅ (그대로) |
| CLI wrapper | ❌ | ✅ `workflow doctor` |
| Decorator | ❌ | ✅ `@graceful_shutdown` + `@v0_7_4_deprecated` |
| Optional dep | ❌ (수동 import) | ✅ `pip install [pbt]` / `[profiling]` |
| pyproject optional-deps | mcp-jsonrpc / mcp-sdk / dev | + `pbt` / `profiling` |
| 누적 test | 191 | 200 |

## Design Detail

### doctor.py
- **argparse mutual exclusive**: `--json` vs `--pretty` (default pretty)
- **pretty table**: 78 char width, 4 status icon (✓/△/✗/—)
- **summary footer**: total / compliant / advisory / non_compliant count
- **--exit-on-fail**: CI hook (alert/notification)
- **--project-root**: 명시 (default: import-time Path 계산)

### decorators.py
- **@graceful_shutdown**:
  - 직접 호출 `@graceful_shutdown` (no paren) + parameterized `@graceful_shutdown(cleanup=...)` 모두 지원
  - SIGINT / SIGTERM / SIGBREAK (Windows) 3 signal handler 등록
  - Cleanup callback: thread-safe Event + timeout_sec
  - Original handler 복원 (fn 종료 시)
- **@v0_7_4_deprecated**:
  - DeprecationWarning 발행 (stacklevel=2)
  - fn 본문은 그대로 실행 (backward compat)
  - `version` parameter 로 제거 예정 명시

### Optional Dep 전략
- **try-import** 패턴: `try: import hypothesis except ImportError: pass`
- **Runtime hint**: 미설치 시 status=`advisory` + `pip install <pkg>` 메시지
- **pyproject.toml** `optional-dependencies` 로 의도 명시
- **`--json` output**: 설치 여부도 status 에 영향 (선택적 enforce)

## Cross-cutting 적용

- 다른 project 도 동일 optional dep 전략 가능:
  1. pyproject 의 `optional-dependencies` 에 등록
  2. helper module 의 try-import
  3. status advisory + install hint
- CLI wrapper 정공법: argparse + 4 sub-mode (default / --json / --pretty / --exit-on-fail)
- Decorator 정공법: `@functools.wraps` + `__wrapped__` + custom marker (`__graceful_shutdown__`)

## v0.7.5+ follow-up (open)

1. v0.7.5: doctor.py 의 `--watch` mode (file 변경 시 자동 re-evaluate)
2. v0.7.5: doctor.py 의 `--baseline-filter=<prefix>` (SEC-* / PBT-* / PERF-* / RES-*)
3. v0.7.5: `@retry` decorator (transient error 자동 retry, AIDLC 의 circuit breaker pattern)
4. v0.7.5: `@timeout` decorator (max 실행 시간 + force kill)
5. v0.8.0: flat path migration (v0.7.0 flat → v0.7.2+ sub-cat deprecate)

## 호환성

- Python 3.10+ (macOS / Linux / Windows)
- 외부 의존성 0 (stdlib only: argparse, signal, time, functools, warnings)
- v0.7.3 의 4 helper / 7 dispatcher / 44 RuleResult API 변경 없음
- 3 follow-up 모두 **additive** (backward compat)
