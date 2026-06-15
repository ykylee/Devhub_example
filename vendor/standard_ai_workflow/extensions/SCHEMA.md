# Extension System Schema (v0.7.0 step 7)

- 문서 목적: standard_ai_workflow 의 Extension 시스템 (B) 의 file format + opt-in pattern + lint rule 의 SSOT.
- 범위: extensions/ 디렉토리 layout, `<name>-baseline.md` + `<name>-baseline.opt-in.md` 형식, rule ID prefix convention, compliance summary, helper contract
- 상태: stable (v0.7.0 step 7)
- 최종 수정일: 2026-06-12
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/` 3종 (security / testing / resiliency) — commit `b19c819` (2026-06-08)
- 관련 문서: [`./security-baseline.md`](./security-baseline.md) (1차 출시), [`./security-baseline.opt-in.md`](./security-baseline.opt-in.md), [`./testing-baseline.md`](./testing-baseline.md), [`./testing-baseline.opt-in.md`](./testing-baseline.opt-in.md), [`./performance-baseline.md`](./performance-baseline.md), [`./performance-baseline.opt-in.md`](./performance-baseline.opt-in.md), [`../core/extension_helper_contract.md`](../core/extension_helper_contract.md) (v0.7.1+ follow-up), [`../tests/check_extension_system.py`](../tests/check_extension_system.py) (smoke test)

## 1. 왜 Extension 시스템이 필요한가

v0.6.4-5 에서 stage_completion + audit log + question format 등 runtime 안전장치 도입. 그러나 **cross-cutting rule** (security, testing, performance) 의 SSOT 가 없었음. 각 rule 을 workflow 도메인에 맞게 정의/검증할 표준 contract 가 필요.

AIDLC 의 Extension 시스템 (3종: security / testing / resiliency):
- baseline `<name>.md` (rules) + `<name>.opt-in.md` (사용자 선택 prompt)
- 워크플로우 시작 시 opt-in prompt → 사용자가 enable/disable 결정
- enabled rule 은 **hard constraint** (stage completion 에서 위반 시 "Request Changes" 만 표시, "Continue" 옵션 사라짐)

## 2. Directory Layout

```
workflow-source/extensions/
├── SCHEMA.md                       # 본 doc (SSOT)
├── security-baseline.md            # 1차 출시 (v0.7.0 step 8)
├── security-baseline.opt-in.md     # opt-in prompt
├── testing-baseline.md             # v0.7.0 step 7 1차 출시
├── testing-baseline.opt-in.md
├── performance-baseline.md         # v0.7.0 step 7 1차 출시
└── performance-baseline.opt-in.md
```

선택 sub-directory (AIDLC 와 동일, v0.8.0+ 사용):
- `extensions/security/<sub-cat>/` (e.g. `security/auth/`)
- `extensions/testing/property-based/` (AIDLC 1차 출처 위치)
- `extensions/performance/<sub-cat>/` (v0.8.0+)

본 v0.7.0 step 7 은 flat structure (security-baseline, testing-baseline, performance-baseline) 만. sub-cat 는 follow-up.

## 3. File Format

### 3.1 `<name>-baseline.md`

```markdown
# <Name> Baseline Extension (v0.7.0+)

- 문서 목적: ...
- 범위: N rule (<PREFIX>-NN)
- 상태: stable
- 최종 수정일: YYYY-MM-DD
- 1차 출처: AIDLC `extensions/<cat>/<sub-cat>/<name>.md` (line, commit)
- 관련 문서: ...

## 1. 왜 <Name> Baseline 이 필요한가
...

## 2. Rule 정의

### 2.1 Rule <PREFIX>-01: <Title>
**Rule**: <rule body>
**Verification**:
- <bullet 1>
- <bullet 2>
- <bullet 3>

### 2.2 Rule <PREFIX>-02: <Title>
...

## N. Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| <PREFIX>-01 | <Title> | ✅/❌/⚠️/⏸/N/A | <note> |
| ... |

## N+1. References
...
```

### 3.2 `<name>-baseline.opt-in.md`

```markdown
# <Name> Baseline Opt-In (v0.7.0+)

- 문서 목적: workflow 시작 시 사용자 opt-in prompt
- 상태: stable
- 최종 수정일: YYYY-MM-DD

## 1. Question (workflow-start 시 1회 표시)

> Question: Should <name> baseline extension rules be enforced for this project?
>
> A) Yes — enforce all <PREFIX> rules as blocking constraints (recommended for production-grade)
>
> B) No — skip all <PREFIX> rules (suitable for PoCs, prototypes, experimental projects)
>
> P) Partial — enforce only <PREFIX>-<subset> (e.g. PBT-02/03/07/08/09)
>
> X) Other (please describe after [Answer]: tag below)
>
> [Answer]:

## 2. Response Processing

답변은 `ai-workflow/memory/active/<name>_status.md` 에 기록:
- `A) Yes` → enabled = True, partial = False
- `B) No` → enabled = False
- `P) Partial` → enabled = True, partial = True, rules = subset

## 3. State File Schema

```yaml
<name>_baseline:
  status: enabled | disabled | partial
  partial_rules: [<PREFIX>-02, <PREFIX>-03, ...]  # partial 시만
  decided_at: "<ISO 8601>"
  decided_by: "<user|agent|default>"
  standard_ref: "AIDLC extensions/<cat>/<sub-cat>/<name>.md @ <commit>"
```
```

## 4. Rule ID Prefix Convention

| Extension | Prefix | Example | AIDLC 대응 |
|---|---|---|---|
| security-baseline | `SEC-WF-NN` | SEC-WF-01 | AIDLC 의 SECURITY-NN (15 rule) |
| testing-baseline | `TST-WF-NN` | TST-WF-01 | AIDLC 의 PBT-NN (9 rule) |
| performance-baseline | `PERF-WF-NN` | PERF-WF-01 | AIDLC 에 없음 — 우리 domain 적응 |

Rule ID format: `^[A-Z]+-WF-\d{2}$` (2 digit zero-padded).

## 5. Hard Constraint 정책 (enabled 시)

- stage completion message 에 `<Name> Compliance` 섹션 추가
- 각 rule 마다 compliant / non-compliant / N/A 표시
- 1+ rule 이 non-compliant 이면 "Continue" 옵션 사라짐, "Request Changes" 만 표시
- partial mode: 부분 enforcement rule 만 hard, 나머지 advisory
- N/A: not applicable 한 rule 은 blocking 아님

## 6. Helper Contract (v0.7.1+)

각 extension 마다 `workflow_kit.common.contracts.<name>_baseline.evaluate_compliance()` helper. 본 step 7 은 spec level 만 — runtime helper 는 v0.7.1 follow-up.

```python
# v0.7.1+ 의사 코드
def evaluate_compliance(
    project_root: Path,
    enabled_rules: list[str],
) -> ComplianceSummary:
    """각 enabled rule 마다 compliant / non-compliant / N/A 평가."""
    ...
```

## 7. Lint Rule (smoke test)

`tests/check_extension_system.py` 가 다음을 검증:
1. SCHEMA.md 존재
2. 모든 extension 이 `<name>-baseline.md` + `<name>-baseline.opt-in.md` 페어
3. 각 baseline 의 rule ID 가 `<PREFIX>-WF-NN` 형식
4. 각 baseline 의 rule 이 `**Rule**` + `**Verification**` (≥ 3 bullet) 형식
5. 각 baseline 의 `## N. Compliance Summary` table 에 모든 rule ID 명시
6. 각 opt-in 이 `> Question:` + `[Answer]:` 형식
7. 각 opt-in 의 A/B/P/X 4 옵션 모두 존재
8. AIDLC cross-reference (rule ID 또는 1차 출처 file path 명시)
9. Rule topic coverage (rule 간 중복 topic 없음)
10. AIDLC extensions/<cat>/<sub-cat>/<name>.md file 존재 검증 (1차 출처 SSOT)

## 8. 우리 사용 패턴 적응

| AIDLC 패턴 | 우리 적응 | 비고 |
|---|---|---|
| security-baseline 15 rule | 6 rule (SEC-WF-01~06) | workflow domain specific |
| testing/property-based 9 rule | 6 rule (TST-WF-01~06) | smoke test 위주 |
| resiliency-baseline 16 rule | N/A (v0.7.0), 6 rule (PERF-WF-01~06) | workflow runtime 성능 |
| HTTP / Lambda / CDK 위주 | harness / workflow_kit / smoke test | 인프라 = local runtime |
| Compliance audit + state.md | `ai-workflow/memory/active/<name>_status.md` + state.json | 우리 state.json + active memory |

## 9. 한계 / 예외

- **N/A 처리**: 우리 workflow 에 HTTP API / Lambda / RDS 등이 없음. SEC-WF-04 (HTTP headers) 같은 AIDLC rule 은 우리 도메인에서 N/A. 우리 적응은 "우리 운영에 applicable 한 rule" 만 emit.
- **Partial mode**: 1차 출시 3종 모두 partial 옵션 제공 (P). 사용자가 subset 만 enable 가능.
- **v0.7.0 sub-cat 미도입**: AIDLC 의 `testing/property-based/` 같은 sub-cat 는 v0.8.0+ follow-up.
- **Helper runtime**: 본 step 7 은 spec level (md + smoke test). runtime evaluate_compliance helper 는 v0.7.1+.

## 10. Follow-up (v0.7.1+)

- `workflow_kit.common.contracts.<name>_baseline.evaluate_compliance()` 구현 (3종)
- session-start 에 opt-in prompt 자동 통합
- `state.json` 의 `<name>_baseline` 필드 schema validation
- v0.8.0: sub-cat 도입 (e.g. `extensions/security/auth/`, `extensions/testing/property-based/`)
- v0.8.0: 4종 (resiliency) 추가 — workflow_kit health check + 장애 대응

## 11. References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/` (commit `b19c819`, 2026-06-08)
  - `security/baseline/security-baseline.md` (307 line, 15 rule)
  - `testing/property-based/property-based-testing.md` (284 line, 9 rule)
  - `resiliency/baseline/resiliency-baseline.md` (490 line, 16 rule)
- 우리 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md` (B = Extension 시스템)
- 우리 1차 출시: `extensions/security-baseline.md` (242 line, 6 rule, commit `dc2c22b`)
- 우리 검증: `tests/check_extension_system.py` (smoke test)
