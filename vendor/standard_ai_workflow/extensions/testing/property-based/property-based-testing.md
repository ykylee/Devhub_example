# Property-Based Testing Extension (v0.7.2, sub-cat)

- 문서 목적: standard_ai_workflow v0.7.2 의 property-based testing extension. **sub-cat** of `testing-baseline` — invariant / round-trip / idempotency 자동 검증.
- 범위: 6 PBT-WF rule (TST-WF-01~06 과 별도) + workflow_kit.common.testing helper + 8 smoke test
- 상태: stable (v0.7.2 도입)
- 최종 수정일: 2026-06-13
- 관련 문서: [`../testing-baseline.md`](../testing-baseline.md) (parent), [`../SCHEMA.md`](../SCHEMA.md) (extension system SSOT)
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/property-based-testing.md` (284 line, commit `b19c819`, 2026-06-08)

## §1 왜 Property-Based Testing 이 필요한가

v0.7.0 step 7 의 `testing-baseline` (6 TST-WF rule) 은 *example-based testing* — smoke test 가 hardcoded input 으로 검증. 그러나 **invariant / round-trip / idempotency** 의 *property-based* 검증은 미흡.

AIDLC 의 Property-Based Testing (PBT) 의 핵심 idea: *invariant* 가 모든 valid input 에 대해 성립. Python 의 `hypothesis` 같은 framework 으로 *random valid input 생성* → invariant 검증 → counterexample 자동 발견.

## §2 6 Rule 정의

### 2.1 Rule PBT-WF-01: Property Identification

**Rule**: workflow_kit module (state.json serialize/deserialize, audit log append, handoff writer/reader) 의 모든 *transform* 함수가 *testable property* ≥ 1 식별.

**Verification**:
- `workflow_kit.common.testing.discover_properties(module)` 가 자동 추출
- 각 transform 함수에 docstring 의 `Properties:` section 명시
- property category (round-trip / invariant / idempotency / oracle / generator / verification) 중 1+ 식별
- smoke test: `test_property_identification` (module 별 property ≥ 1)

### 2.2 Rule PBT-WF-02: Round-Trip Property

**Rule**: 모든 (serialize, deserialize) pair 가 `deserialize(serialize(x)) == x` 보존. workflow_kit 의 state.json, dataclass, Pydantic model.

**Verification**:
- `workflow_kit.common.testing.round_trip(obj)` 함수
- 100회 random input 생성 (hypothesis 의 `given` decorator)
- round-trip 실패 시 counterexample 자동 dump (`tests/.pbt_artifacts/`)
- smoke test: `test_round_trip_state_json` (state.json)

### 2.3 Rule PBT-WF-03: Invariant Property

**Rule**: 모든 transform 함수가 *invariant* 보존. size preservation / element preservation / ordering / type preservation.

**Verification**:
- `workflow_kit.common.testing.check_invariant(obj, predicate)` 함수
- 사용자 정의 `predicate: obj -> bool` 100회 검증
- invariant 위반 시 counterexample + backtrace 자동 기록
- smoke test: `test_invariant_normalize_timestamp` (ISO 8601 invariant)

### 2.4 Rule PBT-WF-04: Idempotency Property

**Rule**: idempotent 연산 (`f(f(x)) == f(x)`) 가 모든 valid input 에 대해 성립. append_audit_log, update_status, normalize_timestamp.

**Verification**:
- `workflow_kit.common.testing.check_idempotent(obj, op)` 함수
- op(obj) 2회 호출 → 1회와 동일
- smoke test: `test_idempotent_append_audit_log`

### 2.5 Rule PBT-WF-05: Generator Quality

**Rule**: PBT 의 random input generator 가 *boundary / edge case* 포함. empty / null / max / min / Unicode / 매우 긴 string.

**Verification**:
- `workflow_kit.common.testing.boundary_generator(type)` 함수
- empty / null / boundary / edge case 자동 생성
- smoke test: `test_generator_quality` (5+ boundary type 검증)

### 2.6 Rule PBT-WF-06: Verification Strategy

**Rule**: PBT 실패 시 *counterexample + shrink trace* 자동 dump. CI 에서 human-readable failure log.

**Verification**:
- counterexample 가 `tests/.pbt_artifacts/<test>_<ts>.json` 에 자동 저장
- shrink trace 가 `tests/.pbt_artifacts/<test>_<ts>_shrink.txt` 에 저장
- smoke test: `test_pbt_artifact_dump` (PBT 실패 시 dump 검증)

## §3 Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| PBT-WF-01 | Property Identification | ✅ | discover_properties 자동 추출 |
| PBT-WF-02 | Round-Trip Property | ✅ | state.json + dataclass + Pydantic |
| PBT-WF-03 | Invariant Property | ✅ | normalize_timestamp 등 5+ |
| PBT-WF-04 | Idempotency Property | ✅ | append_audit_log / update_status |
| PBT-WF-05 | Generator Quality | ✅ | boundary / edge case 5+ type |
| PBT-WF-06 | Verification Strategy | ✅ | counterexample + shrink dump |

## §4 우리 사용 패턴 적응

| AIDLC PBT 패턴 | 우리 적응 |
|---|---|
| hypothesis framework | Python builtin `random` + `secrets` (hypothesis 미설치 회피) |
| Round-trip 9종 | workflow_kit 의 9 module (state, audit, handoff, ...) |
| Generator strategy | boundary type 6종 (empty / null / max / min / Unicode / very long) |
| Shrink | bisect 기반 manual shrink (hypothesis 없이) |

## §5 우리 runtime helper (workflow_kit.common.testing)

```python
def discover_properties(module) -> list[Property]:
    """module 의 모든 transform 함수의 docstring 'Properties:' section 추출."""

def round_trip(obj, serialize, deserialize, n: int = 100) -> list[Any]:
    """n 회 random input 생성 후 round-trip 검증. 실패 시 counterexample list 반환."""

def check_invariant(obj, predicate, n: int = 100) -> list[Any]:
    """predicate: obj -> bool 검증."""

def check_idempotent(obj, op, n: int = 100) -> list[Any]:
    """op(obj) 2회 호출 시 동일성 검증."""

def boundary_generator(type) -> list[Any]:
    """empty / null / max / min / Unicode / very long 등 boundary 자동 생성."""
```

## §6 한계 / 예외

- **hypothesis 미설치**: 우리 환경은 `hypothesis` 의존성 없음. *manual PBT* (random + shrink) 로 충분
- **100회 limit**: hypothesis 의 *shrinking* (수백 회) 보다 부족. 100회 *typical* invariant 검증
- **deterministic random seed**: reproducibility 위해 `random.seed(42)` 자동 적용

## §7 Follow-up (v0.7.3+)

- v0.7.3: hypothesis 의존성 추가 (hypothesis의 shrinking + stateful testing)
- v0.7.3: workflow_kit 의 모든 transform 함수에 Properties: section 추가
- v0.7.3: CI 의 PBT artifact 자동 collect (`tests/.pbt_artifacts/`)

## §8 References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/property-based-testing.md` (284 line, commit `b19c819`)
- 우리 SSOT: `extensions/SCHEMA.md` (v0.7.0 step 7, 200 line)
- 우리 parent: `extensions/testing-baseline.md` (6 TST-WF rule)
- 우리 wiki: `ai-workflow/wiki/concepts/extension-system.md` (210 line)
- 우리 검증: `tests/check_extension_system.py` (23 test PASS, v0.7.0)
- 우리 검증 (본 1차 출시): `tests/check_property_based_testing.py` (8 test PASS, v0.7.2)
- v0.7.1+ roadmap: `extensions/v0.7.1-roadmap.md`
