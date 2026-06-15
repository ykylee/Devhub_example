# Testing Baseline Extension (v0.7.0, AIDLC 차용)

- 문서 목적: standard_ai_workflow 의 cross-cutting testing rule baseline. AIDLC `extensions/testing/property-based/` 차용. v0.7.0 step 7 1차 출시 — testing-baseline 1종.
- 범위: 6 rule (TST-WF-01 ~ TST-WF-06), opt-in pattern, smoke test 통합
- 대상 독자: workflow 설계자, AI agent, 운영자, QA 검토자
- 상태: stable (v0.7.0 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: [`./SCHEMA.md`](./SCHEMA.md) (extension system SSOT), [`./testing-baseline.opt-in.md`](./testing-baseline.opt-in.md), [`./security-baseline.md`](./security-baseline.md), [`./performance-baseline.md`](./performance-baseline.md), [`../core/audit_log_standard.md`](../core/audit_log_standard.md) (test log 표준)
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/property-based-testing.md` (284 line, 2026-06-08 commit `b19c819`)

## 1. 왜 Testing Baseline 이 필요한가

v0.6.4-7 의 smoke test 시스템 (7종 check_*.py → 8종) 은 example-based test 위주. AIDLC 의 Property-Based Testing (PBT) 의 핵심 idea (invariant, round-trip, idempotency) 가 우리 smoke test 에 부분 적용.

본 1차 출시: **testing-baseline 1종**, 우리 운영 컨텍스트에 applicable 한 6 rule. AIDLC 의 9 PBT rule → 우리 6 TST-WF rule 로 도메인 적응.

## 2. 우리 적응 결정

AIDLC PBT 의 9 rule 중:
- PBT-01 Property Identification (design stage) — 우리: §2.1 TST-WF-01
- PBT-02 Round-Trip (serialize/deserialize) — 우리: §2.2 TST-WF-02
- PBT-03 Invariant (transformation) — 우리: §2.3 TST-WF-03
- PBT-04 Idempotency — 우리: §2.4 TST-WF-04
- PBT-05 Commutativity — N/A (workflow stage 는 순차, commutative 불가)
- PBT-06 Stateful Properties — N/A (workflow_kit state.json 은 persistent store, 별도 검증)
- PBT-07 Generator Quality — 우리: §2.5 TST-WF-05
- PBT-08 Oracle Properties — N/A (oracle 없음)
- PBT-09 Verification Strategy — 우리: §2.6 TST-WF-06

## 3. Rule 정의

### 3.1 Rule TST-WF-01: Smoke Test Coverage Required

**Rule**: 모든 신규 workflow component (skill / MCP / runner / template) 는 smoke test (`tests/check_*.py`) 를 가져야 한다. design stage 에서 testable property 가 식별되어야 한다.

**Verification**:
- 신규 component 추가 시 `tests/check_<component>.py` 가 함께 추가됨
- 각 smoke test 가 ≥ 5 test case 보유
- design doc (UOW, handoff) 에 "Testable Properties" section 명시
- property category (round-trip / invariant / idempotency / generator / verification) 중 1+ 식별

### 3.2 Rule TST-WF-02: Round-Trip Properties for State Serialization

**Rule**: workflow_kit 의 state.json ↔ dict ↔ JSON 변환은 round-trip 보존되어야 한다. parse + serialize 가 identity.

**Verification**:
- `state.json` parse → dict → serialize → state.json = 원본 (byte-equal)
- 모든 Pydantic / dataclass model 의 to_dict() + from_dict() 가 identity
- ISO 8601 timestamp 가 마이크로초 단위까지 보존 (audit log 정합)
- workflow_kit 의 `normalize_timestamp` + `serialize_state` 가 round-trip

### 3.3 Rule TST-WF-03: Invariant Properties for Transformations

**Rule**: workflow 의 transformation (handoff writer/reader, audit append, normalizer) 은 invariant 를 보존해야 한다. size / element / order / type preservation.

**Verification**:
- handoff reader 가 writer 의 output byte-count 보존
- audit log append 가 기존 entry 보존 (append-only)
- normalizer 가 원본 field 보존 (lossy transformation N/A 표시)
- 각 invariant 가 smoke test 의 `test_invariant_*` 1+ case 로 검증

### 3.4 Rule TST-WF-04: Idempotency Properties for State Operations

**Rule**: idempotent 연산 (audit log dedup, state.json 의 status field update, template re-emit) 은 `f(f(x)) = f(x)` 보장.

**Verification**:
- `append_audit_log()` 를 동일 entry 로 2번 호출 시 1번 entry 만 (dedup)
- `update_status()` 를 동일 status 로 2번 호출 시 side effect 0
- template re-emit 이 동일 output (byte-equal)
- idempotency violation 시 `TST-WF-04 Non-Compliant` + log

### 3.5 Rule TST-WF-05: Generator Quality for Smoke Test Inputs

**Rule**: smoke test 의 input data 는 hardcoded 가 아닌 generator (또는 fixture factory) 가 생성. boundary / edge case / random valid input 모두 포함.

**Verification**:
- 각 smoke test 가 hardcoded input ≤ 30% (≥ 70% generator/fixture)
- boundary test 1+ (empty / null / max / min)
- edge case 1+ (Unicode / very long string / special char)
- test data 가 `tests/fixtures/` 또는 `conftest.py` 의 factory 에서 생성

### 3.6 Rule TST-WF-06: Verification Strategy Documented

**Rule**: 모든 smoke test 는 "what is verified" 1-line docstring + verification methodology 명시. 단순 실행이 아닌 의도적 검증.

**Verification**:
- 각 test function 의 docstring 1+ line ("what" + "how")
- assertion message 가 expected vs actual 명시
- 실패 시 debug 가능한 traceback (assertion + source_context)
- smoke test README 또는 docstring 에 verification strategy section

## 4. Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| TST-WF-01 | Smoke Test Coverage Required | ✅ | 8 smoke test (107 test PASS) |
| TST-WF-02 | Round-Trip Properties | ✅ | state.json + audit log byte-equal |
| TST-WF-03 | Invariant Properties | ✅ | handoff + audit log + normalizer |
| TST-WF-04 | Idempotency Properties | ✅ | append_audit_log dedup 검증됨 |
| TST-WF-05 | Generator Quality | ⚠️ | boundary test 일부 부족 (v0.7.1 follow-up) |
| TST-WF-06 | Verification Strategy | ✅ | 모든 test docstring 1+ line |

## 5. 우리 운영 적용

- session-start 시 opt-in prompt (3종 extension 통합)
- workflow_kit.common.contracts.testing_baseline.evaluate_compliance() v0.7.1+ 구현
- v0.7.1+ smoke test generator (boundary / edge case 자동 생성)
- v0.7.1+ idempotency test 추가 (각 helper 마다)

## 6. 한계 / 예외

- **N/A 처리**: PBT-05 (commutativity) / PBT-06 (stateful) / PBT-08 (oracle) — 우리 workflow 에 N/A. 본 baseline 미포함.
- **PoC 모드**: opt-in B) No 선택 시 rule 모두 advisory. 단 TST-WF-01 (smoke test coverage) 은 PoC 라도 권장.
- **v0.7.1+ follow-up**: 6 rule 의 runtime enforcement helper (TST-WF-04 idempotency 자동 검증 등)

## 7. References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/property-based-testing.md` (284 line, commit `b19c819`)
- 우리 L1 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md` (B = Extension 시스템)
- 우리 운영: `tests/check_*.py` 8종 (smoke test)
- 우리 SSOT: `extensions/SCHEMA.md` (extension system schema)
- 우리 검증: `tests/check_extension_system.py` (smoke test)
