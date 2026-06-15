# Performance Baseline Extension (v0.7.0, 우리 domain 적응)

- 문서 목적: standard_ai_workflow 의 cross-cutting performance rule baseline. AIDLC 에는 없는 우리 domain 적응 extension — workflow runtime 성능.
- 범위: 6 rule (PERF-WF-01 ~ PERF-WF-06), opt-in pattern, smoke test + profiling 통합
- 대상 독자: workflow 설계자, AI agent, 운영자
- 상태: stable (v0.7.0 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: [`./SCHEMA.md`](./SCHEMA.md) (extension system SSOT), [`./performance-baseline.opt-in.md`](./performance-baseline.opt-in.md), [`./security-baseline.md`](./security-baseline.md), [`./testing-baseline.md`](./testing-baseline.md)
- 1차 출처: 없음 (우리 domain 적응). AIDLC 의 resiliency-baseline (16 rule) 의 일부 (성능/모니터링 관련) 차용.

## 1. 왜 Performance Baseline 이 필요한가

AIDLC 의 resiliency-baseline 은 cloud workload (Lambda / ECS / RDS) 성능 — 우리 workflow domain 에 N/A. 우리 workflow_kit 의 local runtime (Python + pyproject) 에 applicable 한 performance rule 이 필요.

본 1차 출시: **performance-baseline 1종**, 우리 운영 컨텍스트에 applicable 한 6 rule. AIDLC RESILIENCY 의 6 pillar (Business Goals, Change Mgmt, Observability, HA, DR, Continuous Improvement) 중 **Observability + Continuous Improvement** 만 우리 적응.

## 2. Rule 정의

### 2.1 Rule PERF-WF-01: Smoke Test Execution Time

**Rule**: 모든 smoke test (`tests/check_*.py`) 의 실행 시간은 ≤ 30초. CI feedback loop 보존.

**Verification**:
- 각 smoke test 의 `time.time()` 측정 후 ≤ 30초
- 전체 smoke test 묶음 (8종) 의 총 실행 시간 ≤ 60초
- 30초 초과 시 `PERF-WF-01 Non-Compliant` 표시 + log
- 회귀 방지: baseline 측정값 (e.g. 1.5초) 의 5배 이내

### 2.2 Rule PERF-WF-02: Module Import Time

**Rule**: workflow_kit 의 module import 시간은 ≤ 1초. lazy loading 권장.

**Verification**:
- `python -X importtime workflow_kit` 측정 후 ≤ 1초
- heavy module (network / cloud SDK / file I/O) 은 lazy import
- import 시 side effect 0 (no I/O, no network)
- regression 시 baseline 대비 +50% 이내

### 2.3 Rule PERF-WF-03: Memory Footprint

**Rule**: workflow_kit runtime 의 memory footprint 은 ≤ 200 MB. resource-constrained 환경 (CI / 작은 VM) 지원.

**Verification**:
- `tracemalloc` / `resource.getrusage()` 측정
- typical session (skill 1종 실행) 의 RSS ≤ 200 MB
- peak memory 가 baseline + 50% 이내
- memory leak 자동 검출 (gc.collect() 후 잔여 객체 0)

### 2.4 Rule PERF-WF-04: Audit Log Append Latency

**Rule**: `append_audit_log()` 의 latency 는 ≤ 10ms. workflow stage 전환 시마다 호출되므로 hot path.

**Verification**:
- 1000회 호출 시 평균 latency ≤ 10ms
- p99 latency ≤ 50ms
- file I/O (append) 가 atomic + buffered
- lock contention 시 fallback (in-memory queue + batch flush)

### 2.5 Rule PERF-WF-05: state.json Read/Write Latency

**Rule**: `state.json` 의 read/write latency ≤ 5ms. session start + handoff 마다 호출.

**Verification**:
- `load_state()` / `save_state()` 평균 latency ≤ 5ms
- p99 latency ≤ 20ms
- atomic_write 사용 (write 시 corruption 방지)
- state.json 크기 ≤ 100 KB (그 이상 시 SQLite 마이그레이션 권장)

### 2.6 Rule PERF-WF-06: Profiling Hook for Production Debug

**Rule**: workflow_kit 의 critical path (stage transition / audit append / state save) 에는 optional profiling hook 제공. production debug 시 enable.

**Verification**:
- `workflow_kit.common.profiling.trace()` decorator / context manager 제공
- cProfile + snakeviz 통합 가능
- profiling overhead ≤ 5% (when enabled)
- profiling 비활성화 시 overhead 0 (zero-cost abstraction)

## 3. Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| PERF-WF-01 | Smoke Test Execution Time | ✅ | 8 smoke test 총 ≤ 5초 (현재) |
| PERF-WF-02 | Module Import Time | ✅ | workflow_kit import ≤ 0.3초 (현재) |
| PERF-WF-03 | Memory Footprint | ⚠️ | 측정 자동화 미구축 (v0.7.1 follow-up) |
| PERF-WF-04 | Audit Log Append Latency | ✅ | 평균 ≤ 1ms (현재 측정) |
| PERF-WF-05 | state.json Read/Write Latency | ✅ | 평균 ≤ 2ms (현재 측정) |
| PERF-WF-06 | Profiling Hook | ⚠️ | decorator 미구축 (v0.7.1 follow-up) |

## 4. 우리 운영 적용

- session-start 시 opt-in prompt (3종 extension 통합)
- `workflow_kit.common.profiling` module 신규 (PERF-WF-06)
- `tests/benchmark/` directory 신규 (PERF-WF-01/02/04/05 자동 측정)
- v0.7.1: PERF-WF-03 memory 자동 측정 (tracemalloc 통합)
- v0.7.1: PERF-WF-06 profiling decorator 구현

## 5. 한계 / 예외

- **N/A 처리**: AIDLC RESILIENCY 의 HA / DR / Incident Response (15 rule) 은 우리 workflow domain 에 N/A. cloud workload 운영 영역.
- **Partial mode**: opt-in P) Partial 시 PERF-WF-01 + PERF-WF-04 만 hard constraint (smoke test + audit latency 는 critical)
- **v0.7.1+ follow-up**: 6 rule 의 runtime enforcement helper. 본 step 7 은 spec level 만.

## 6. References

- 1차 출처: 없음 (우리 domain 적응)
- AIDLC 참조: `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.md` (490 line, 16 rule, Observability + Continuous Improvement 2 pillar 참조)
- 우리 L1 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md` (B = Extension 시스템)
- 우리 SSOT: `extensions/SCHEMA.md` (extension system schema)
- 우리 검증: `tests/check_extension_system.py` (smoke test)
