---
type: concept
status: active
last_ingested_from: workflow-source/extensions/SCHEMA.md + workflow-source/extensions/{security,testing,performance}-baseline.md
related_pages: [concepts/security-baseline, concepts/reverse-engineering, concepts/stage-gate-pattern, decisions/adr-005-r9-wiki-source-rule, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-13
updated: 2026-06-13
---

# Extension System (v0.7.0 step 7, AIDLC 차용)

- 문서 목적: standard_ai_workflow v0.7.0 step 7 의 Extension 시스템 (AIDLC `extensions/` 차용) 의 file format + opt-in pattern + lint rule + helper contract SSOT.
- 범위: 3종 baseline (security / testing / performance) + SCHEMA.md + opt-in pattern + 23 smoke test
- 최종 수정일: 2026-06-13

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/extensions/SCHEMA.md` (200 lines, v0.7.0 stable) |
| 2 | 3종 baseline | `security-baseline.md` (242 line, 6 SEC-WF rule) / `testing-baseline.md` (130 line, 6 TST-WF rule) / `performance-baseline.md` (130 line, 6 PERF-WF rule) |
| 3 | 3종 opt-in | `<name>-baseline.opt-in.md` (A/B/P/X 4 옵션 + State File Schema) |
| 4 | smoke test | `workflow-source/tests/check_extension_system.py` (389 lines, 23 test PASS) |
| 5 | Source 1차 출처 | AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/` 3종 (security / testing / resiliency, commit `b19c819`, 2026-06-08) |
| 6 | 도입 버전 | v0.7.0 (commit `0052da1`) — 1차 출시 (security step 8 + testing/performance step 7) |
| 7 | 관련 ADR | 없음 (신규) |
| 8 | 관련 토픽 | [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 B (Extension 시스템) |

## §2 왜 Extension 시스템이 필요한가  {#s2-why}

v0.6.4-5 의 stage_completion + audit log + question format 등 runtime 안전장치 도입. 그러나 **cross-cutting rule** (security, testing, performance) 의 SSOT 가 없었음. 각 rule 을 workflow 도메인에 맞게 정의/검증할 표준 contract 가 필요.

AIDLC 의 Extension 시스템 (3종: security / testing / resiliency):
- baseline `<name>.md` (rules) + `<name>.opt-in.md` (사용자 선택 prompt)
- 워크플로우 시작 시 opt-in prompt → 사용자가 enable/disable 결정
- enabled rule 은 **hard constraint** (stage completion 에서 위반 시 "Request Changes" 만 표시, "Continue" 옵션 사라짐)

## §3 Directory Layout  {#s3-directory-layout}

```
workflow-source/extensions/
├── SCHEMA.md                       # SSOT (200 line, 11 section)
├── security-baseline.md            # 1차 출시 (v0.7.0 step 8, commit dc2c22b)
├── security-baseline.opt-in.md
├── testing-baseline.md             # 1차 출시 (v0.7.0 step 7, commit 0052da1)
├── testing-baseline.opt-in.md
├── performance-baseline.md         # 1차 출시 (v0.7.0 step 7, commit 0052da1)
└── performance-baseline.opt-in.md
```

v0.7.0 flat structure. v0.8.0+ sub-cat 도입 (AIDLC 의 `security/auth/`, `testing/property-based/` 등).

## §4 File Format  {#s4-file-format}

### 4.1 `<name>-baseline.md`

```markdown
# <Name> Baseline Extension (v0.7.0+)
- 문서 목적 / 범위 / 상태 / 최종 수정일 / 1차 출처 / 관련 문서
## 1. 왜 <Name> Baseline 이 필요한가
## 2. Rule 정의
### 2.1 Rule <PREFIX>-01: <Title>
**Rule**: <rule body>
**Verification**:
- <bullet 1>
- <bullet 2>
- <bullet 3>
## N. Compliance Summary (✅/❌/⚠️/⏸/N/A)
## N+1. References
```

### 4.2 `<name>-baseline.opt-in.md`

```markdown
> Question: Should <name> baseline extension rules be enforced for this project?
> A) Yes — enforce all <PREFIX> rules as blocking constraints
> B) No — skip all <PREFIX> rules
> P) Partial — enforce only <PREFIX>-<subset>
> X) Other
> [Answer]:
```

## §5 Rule ID Prefix Convention  {#s5-prefix}

| Extension | Prefix | Example | AIDLC 대응 |
|---|---|---|---|
| security-baseline | `SEC-WF-NN` | SEC-WF-01 | SECURITY-NN (15 rule) |
| testing-baseline | `TST-WF-NN` | TST-WF-01 | PBT-NN (9 rule) |
| performance-baseline | `PERF-WF-NN` | PERF-WF-01 | RESILIENCY-NN (16 rule, 부분) |

Rule ID format: `^[A-Z]+-WF-\d{2}$` (2 digit zero-padded).

## §6 Hard Constraint 정책 (enabled 시)  {#s6-hard-constraint}

- stage completion message 에 `<Name> Compliance` 섹션 추가
- 각 rule 마다 compliant / non-compliant / N/A 표시
- 1+ rule non-compliant 시 "Continue" 옵션 사라짐, "Request Changes" 만 표시
- partial mode: partial_rules list 외 advisory (non-blocking)
- N/A: not applicable 한 rule 은 blocking 아님

## §7 우리 사용 패턴 적응  {#s7-adaptation}

| AIDLC 패턴 | 우리 적응 | 비율 |
|---|---|---|
| 15 SECURITY rule | 6 SEC-WF (workflow domain) | 40% |
| 9 PBT rule | 6 TST-WF (PBT-05/06/08 N/A) | 67% |
| 16 RESILIENCY rule | 6 PERF-WF (HA/DR/Incident N/A) | 38% |
| A/B/X opt-in (3 옵션) | A/B/P/X (4 옵션) — Partial mode 추가 | 적응 |
| HTTP / Lambda / CDK | harness / workflow_kit / smoke test | 적응 |

**N/A 처리 정공법**:
1. 명시적 N/A section ("§한계/예외")
2. N/A 인 rule 의 검증 skip (smoke test 가 N/A 처리)
3. Compliance Summary table 에 N/A 표시
4. follow-up 에 "extension 추가 시 N/A → applicable 전환 검토"

## §8 Helper Contract (v0.7.1+ follow-up)  {#s8-helper}

각 extension 마다 `workflow_kit.common.contracts.<name>_baseline.evaluate_compliance()` helper.

```python
# v0.7.1+ 의사 코드
def evaluate_compliance(
    project_root: Path,
    enabled_rules: list[str],
) -> ComplianceSummary:
    """각 enabled rule 마다 compliant / non-compliant / N/A 평가."""
```

**v0.7.0 follow-up (commit `71de1b0`)**: `workflow_kit/common/contracts/baselines.py` (516 line) 가 spec → runtime 격상. 3 baseline × 6 rule = 18 RuleResult + 12 test PASS (`check_baselines_compliance.py`).

## §9 한계 / 예외  {#s9-limitations}

- **N/A 처리**: 우리 workflow 에 N/A 한 rule (HTTP API / Lambda / RDS) 은 baseline 미포함
- **Partial mode**: 1차 출시 3종 모두 partial 옵션 제공
- **v0.7.0 sub-cat 미도입**: AIDLC 의 sub-cat 는 v0.8.0+ follow-up
- **Helper runtime**: 본 step 7 spec level → v0.7.0 follow-up 으로 runtime 격상 완료

## §10 Follow-up (v0.7.1+)  {#s10-followup}

- session-start 에 3종 opt-in prompt 통합 (commit `71de1b0` 의 spec level)
- v0.8.0: sub-cat 도입 (e.g. `extensions/security/auth/`, `extensions/testing/property-based/`)
- v0.8.0: 4종 (resiliency) 추가 — workflow_kit health check + 장애 대응

## §11 References  {#s11-references}

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/` 3종 (commit `b19c819`, 2026-06-08)
- 우리 L1 wiki: [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 B
- 우리 SSOT: `workflow-source/extensions/SCHEMA.md` (200 line)
- 우리 검증: `workflow-source/tests/check_extension_system.py` (23 test PASS)
- runtime helper: `workflow_kit/common/contracts/baselines.py` (516 line, commit `71de1b0`)
