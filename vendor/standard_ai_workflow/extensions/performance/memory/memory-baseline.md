# Memory Baseline Extension (v0.7.2, sub-cat)

- 문서 목적: standard_ai_workflow v0.7.2 의 memory-baseline extension. **sub-cat** of `performance-baseline` — 메모리 footprint / leak / GC 측정 자동화.
- 범위: 6 PERF-MEM rule (PERF-WF-01~06 과 별도) + workflow_kit.common.profiling + 8 smoke test
- 상태: stable (v0.7.2 도입)
- 최종 수정일: 2026-06-13
- 관련 문서: [`../performance-baseline.md`](../performance-baseline.md) (parent), [`../SCHEMA.md`](../SCHEMA.md) (extension system SSOT)
- 1차 출처: 없음 (우리 domain 적응). AIDLC 의 `extensions/resiliency/baseline/` Observability pillar 참조.

## §1 왜 Memory Baseline 이 필요한가

v0.7.0 step 7 의 `performance-baseline` (6 PERF-WF rule) 은 *general performance* — smoke test time, import time, audit log latency, state.json R/W latency, profiling hook. 그러나 **메모리 footprint** / **leak** / **GC** 의 *세부* 측정은 *advisory* (PERF-WF-03) 였음.

본 1차 출시: **memory-baseline 1종 (sub-cat)**, 6 rule 의 *세부* 측정.

## §2 6 Rule 정의

### 2.1 Rule PERF-MEM-01: Peak Memory ≤ 200 MB

**Rule**: workflow_kit 의 *peak memory* ≤ 200 MB. tracemalloc / psutil / resource.getrusage() 측정.

**Verification**:
- `workflow_kit.common.profiling.measure_peak_memory(n=10)` 함수
- 10회 평균 + max ≤ 200 MB
- smoke test: `test_peak_memory_under_200mb`

### 2.2 Rule PERF-MEM-02: Memory Leak Detection

**Rule**: 1000회 반복 호출 시 메모리 *누적 증가* ≤ 10 MB. leak 자동 검출.

**Verification**:
- `workflow_kit.common.profiling.detect_leak(obj, op, n=1000)` 함수
- baseline (1회) vs after_1000 (1000회) 의 memory 차이
- leak > 10 MB 시 fail + report
- smoke test: `test_leak_detection_1000_calls`

### 2.3 Rule PERF-MEM-03: GC Pressure ≤ 5%

**Rule**: workflow 실행 중 GC pause time 이 *총 runtime* 의 5% 이하. gc.set_debug() 측정.

**Verification**:
- `workflow_kit.common.profiling.measure_gc_pressure(n=100)` 함수
- gc pause time / total runtime ≤ 0.05
- smoke test: `test_gc_pressure_under_5pct`

### 2.4 Rule PERF-MEM-04: Object Reference Cycle Free

**Rule**: workflow_kit module 의 *모든 public class* 가 gc 가비지 컬렉트 가능. reference cycle 없음.

**Verification**:
- `gc.collect()` 후 unreferenced 객체 0
- `workflow_kit.common.profiling.check_reference_cycle(obj)` 함수
- weakref.finalize 로 자동 검증
- smoke test: `test_no_reference_cycle`

### 2.5 Rule PERF-MEM-05: Memory Profiling Hook

**Rule**: 모든 workflow_kit hot path (stage transition / audit append / state save) 에 *optional memory profiling hook* 제공. `with profile_memory()` context manager.

**Verification**:
- `workflow_kit.common.profiling.profile_memory()` context manager
- 사용 시 tracemalloc 자동 start / stop
- 미사용 시 overhead 0 (zero-cost)
- smoke test: `test_memory_profiling_hook`

### 2.6 Rule PERF-MEM-06: Memory Regression Detection

**Rule**: PR 머지 시 *memory baseline* (commit 별 측정값) 와 비교, ≥ 10% 증가 시 alert.

**Verification**:
- `tools/measure_memory_baseline.py` (별도 tool) 가 commit 별 peak memory 기록
- `tools/.memory_history.jsonl` (v0.7.2+ 누적, score_history 와 유사)
- 현재 vs baseline (10 commit 전) 비교, +10% 시 exit 1
- smoke test: `test_memory_regression_alert` (synthetic +10% 시나리오)

## §3 Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| PERF-MEM-01 | Peak Memory ≤ 200 MB | ✅ | tracemalloc 측정 |
| PERF-MEM-02 | Memory Leak Detection | ✅ | 1000회 반복 |
| PERF-MEM-03 | GC Pressure ≤ 5% | ✅ | gc.set_debug |
| PERF-MEM-04 | Object Reference Cycle Free | ✅ | weakref.finalize |
| PERF-MEM-05 | Memory Profiling Hook | ✅ | context manager |
| PERF-MEM-06 | Memory Regression Detection | ✅ | commit 별 추적 |

## §4 우리 사용 패턴 적응

| AIDLC 패턴 | 우리 적응 |
|---|---|
| CloudWatch Memory Usage | tracemalloc + psutil (local runtime) |
| JVM heap dump | Python objgraph (의존성) — v0.7.3+ |
| Profiler (YourKit) | cProfile + snakeviz (workflow_kit.common.profiling) |
| Memory regression CI | `tools/measure_memory_baseline.py` + GitHub Action |

## §5 우리 runtime helper (workflow_kit.common.profiling)

```python
def measure_peak_memory(n: int = 10) -> dict:
    """n 회 측정 후 avg / max / min peak memory 반환."""

def detect_leak(obj, op, n: int = 1000) -> dict:
    """1000회 반복 시 memory 누적 증가 측정."""

def measure_gc_pressure(n: int = 100) -> float:
    """n 회 반복 시 GC pause time / total runtime 비율."""

def check_reference_cycle(obj) -> bool:
    """weakref.finalize 로 reference cycle 검사."""

@contextmanager
def profile_memory():
    """tracemalloc 자동 start/stop."""
```

## §6 한계 / 예외

- **PERF-MEM-02 의 10 MB 임계값**: 우리 workflow 의 *typical session* 기준. 1000회 호출 시 5-8 MB 누적 — *경계*
- **PERF-MEM-03 의 5% 임계값**: Python 의 GC pause 는 *cycle detector* 가 main loop. 큰 dict / list 일 때 spike 가능
- **weakref.finalize**: 일부 object (e.g. file handle) 는 weakref 지원 안 함

## §7 Follow-up (v0.7.3+)

- v0.7.3: objgraph 의존성 추가 (reference chain 시각화)
- v0.7.3: `memory_profiler` 의존성 추가 (line-by-line memory profile)
- v0.7.3: v0.7.2 의 regression detection 의 CI 통합

## §8 References

- 1차 출처: 없음 (우리 domain 적응). AIDLC `extensions/resiliency/baseline/resiliency-baseline.md` Observability pillar 참조.
- 우리 SSOT: `extensions/SCHEMA.md` (v0.7.0 step 7, 200 line)
- 우리 parent: `extensions/performance-baseline.md` (6 PERF-WF rule)
- 우리 wiki: `ai-workflow/wiki/concepts/extension-system.md` (210 line)
- 우리 검증: `tests/check_extension_system.py` (23 test PASS, v0.7.0)
- 우리 검증 (본 1차 출시): `tests/check_memory_baseline.py` (8 test PASS, v0.7.2)
- v0.7.1+ roadmap: `extensions/v0.7.1-roadmap.md`
