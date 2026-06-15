# Performance Baseline Opt-In (v0.7.0+)

- 문서 목적: workflow 시작 시 사용자 opt-in prompt (performance-baseline)
- 상태: stable (v0.7.0 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: [`./performance-baseline.md`](./performance-baseline.md), [`./SCHEMA.md`](./SCHEMA.md) (extension system SSOT)

## 1. Question (workflow-start 시 1회 표시)

> Question: Should performance baseline extension rules be enforced for this project?
>
> A) Yes — enforce all PERF-WF rules as blocking constraints (recommended for production-grade)
>
> B) No — skip all PERF-WF rules (suitable for PoCs, prototypes, experimental projects)
>
> P) Partial — enforce only PERF-WF-01 (Smoke Test Time) + PERF-WF-04 (Audit Log Latency)
>
> X) Other (please describe after [Answer]: tag below)
>
> [Answer]:

## 2. Response Processing

답변은 `ai-workflow/memory/active/performance_baseline_status.md` 에 기록:

| 답변 | 기록 |
|---|---|
| `A) Yes` | `status: enabled, partial: false, partial_rules: []` |
| `B) No` | `status: disabled, partial: false, partial_rules: []` |
| `P) Partial` | `status: enabled, partial: true, partial_rules: [PERF-WF-01, PERF-WF-04]` |
| `X) ...` | 사용자 의도 분석 후 A/B/P 중 하나에 매핑 |

## 3. State File Schema

```yaml
performance_baseline:
  status: enabled | disabled | partial
  partial_rules: [PERF-WF-01, PERF-WF-04]  # partial 시만
  decided_at: "2026-06-12T23:50:00+09:00"  # ISO 8601
  decided_by: "yklee"  # 또는 "default: A" (Greenfield)
  standard_ref: "internal: workflow domain 적응 (AIDLC RESILIENCY 2 pillar 참조)"
```

`ai-workflow/memory/active/performance_baseline_status.md` 형식:

```markdown
# Performance Baseline Status

- **Enabled**: <yes|no>
- **Decided At**: <ISO 8601>
- **Decided By**: <user|orchestrator>
- **Reason**: <optional free text>

## Extension Configuration

| Extension | Enabled | Decided At |
|---|---|---|
| performance-baseline | yes | 2026-06-12T23:50:00+09:00 |
```

## 4. Default 정책

| 모드 | 기본값 |
|---|---|
| Greenfield (신규) | P) Partial (PERF-WF-01 + 04 만 hard, 나머지 advisory) |
| Brownfield (기존) | 기존 status 존중 (없으면 B) No 후 점진 enable) |
| PoC/Prototype | B) No 권장 |

## 5. Workflow 통합

### 5.1 session-start (workflow 시작 시)

3종 extension (security / testing / performance) opt-in prompt 동시 표시:
- 한 응답으로 3종 모두 결정 가능
- 또는 extension 별 개별 응답 (사용자 선택)

### 5.2 stage completion (stage 완료 시)

- status = enabled → stage completion message 에 "Performance Compliance" 섹션
- 6 rule 마다 compliant / non-compliant / N/A 표시
- partial mode: partial_rules 만 hard constraint, 나머지 advisory

### 5.3 audit log 통합

- opt-in 결정 시 `append_audit_log("opt_in", "performance_baseline=P")`
- stage completion 의 compliance summary 가 audit.md 에 기록

## 6. 한계 / 예외

- **Opt-in 변경**: 사용자가 workflow 중 opt-in 변경 가능. 변경 시 audit log + state.json 갱신.
- **N/A 처리**: 우리 workflow 에 N/A 한 rule (cloud HA / DR / Incident Response) 은 baseline 미포함. blocking 아님.
- **v0.7.1+ follow-up**: runtime `evaluate_compliance()` helper + profiling decorator (PERF-WF-06) + memory 자동 측정 (PERF-WF-03)

## 7. References

- 1차 출처: 없음 (우리 domain 적응)
- AIDLC 참조: `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.opt-in.md` (26 line, 동일 형식)
- 우리 SSOT: `extensions/SCHEMA.md` §3.2 (opt-in 형식)
- 우리 운영: `extensions/security-baseline.opt-in.md`, `extensions/testing-baseline.opt-in.md`
