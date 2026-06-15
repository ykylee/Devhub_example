# Security Baseline Opt-In (v0.7.0+)

- 문서 목적: workflow 시작 시 사용자 opt-in prompt (security-baseline)
- 상태: stable (v0.7.0 도입, v0.7.0 step 7 SCHEMA 형식으로 정합)
- 최종 수정일: 2026-06-12
- 관련 문서: [`./security-baseline.md`](./security-baseline.md), [`./SCHEMA.md`](./SCHEMA.md) (extension system SSOT)
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.opt-in.md` (20 line, commit `b19c819`) — 우리 적응: P) Partial 옵션 추가 (SCHEMA §3.2 형식)

## 1. Question (workflow-start 시 1회 표시)

> Question: Should security baseline extension rules be enforced for this project?
>
> A) Yes — enforce all SEC-WF rules as blocking constraints (recommended for production-grade applications)
>
> B) No — skip all SEC-WF rules (suitable for PoCs, prototypes, experimental projects)
>
> P) Partial — enforce only SEC-WF-01 (Audit Log) + SEC-WF-02 (Stage Gate Approval)
>
> X) Other (please describe after [Answer]: tag below)
>
> [Answer]:

## 2. Response Processing

답변은 `ai-workflow/memory/active/security_baseline_status.md` 에 기록:

| 답변 | 기록 |
|---|---|
| `A) Yes` | `status: enabled, partial: false, partial_rules: []` |
| `B) No` | `status: disabled, partial: false, partial_rules: []` |
| `P) Partial` | `status: enabled, partial: true, partial_rules: [SEC-WF-01, SEC-WF-02]` |
| `X) ...` | 사용자 의도 분석 후 A/B/P 중 하나에 매핑 (또는 `status: custom`) |

## 3. State File Schema

```yaml
security_baseline:
  status: enabled | disabled | partial
  partial_rules: [SEC-WF-01, SEC-WF-02]  # partial 시만
  decided_at: "2026-06-12T22:30:15Z"  # ISO 8601
  decided_by: "yklee"  # 또는 "default: A" (Greenfield)
  standard_ref: "AIDLC extensions/security/baseline/security-baseline.md @ b19c819"
```

`ai-workflow/memory/active/security_baseline_status.md` 형식:

```markdown
# Security Baseline Status

- **Enabled**: <yes|no>
- **Decided At**: <ISO 8601>
- **Decided By**: <user|orchestrator>
- **Reason**: <optional free text>

## Extension Configuration

| Extension | Enabled | Decided At |
|---|---|---|
| security-baseline | yes | 2026-06-12T22:30:15Z |
```

`state.json` 의 `security_baseline` 필드에 자동 동기화:
- `status` 가 `enabled` 면 stage completion 의 `Security Compliance` 섹션 활성
- `partial_rules` 가 partial enforcement 대상 (나머지는 advisory)
- 1+ rule non-compliant 시 "Continue" 옵션 사라짐

## 4. Default 정책

| 모드 | 기본값 |
|---|---|
| Greenfield (신규) | A) Yes (production 권장) |
| Brownfield (기존) | 기존 status 존중 (없으면 B) No 후 점진 enable) |
| CI/CD | opt-in prompt 표시 안 함, 기존 `security_baseline_status.md` 값 사용 |

## 5. Workflow 통합

### 5.1 session-start (workflow 시작 시)

3종 extension (security / testing / performance) opt-in prompt 동시 표시:
- 한 응답으로 3종 모두 결정 가능
- 또는 extension 별 개별 응답 (사용자 선택)

### 5.2 stage completion (stage 완료 시)

- status = enabled → stage completion message 에 "Security Compliance" 섹션
- 6 rule 마다 compliant / non-compliant / N/A 표시
- partial mode: partial_rules 만 hard constraint, 나머지 advisory

### 5.3 audit log 통합

- opt-in 결정 시 `append_audit_log("opt_in", "security_baseline=A")`
- stage completion 의 compliance summary 가 audit.md 에 기록

## 6. 한계 / 예외

- **Opt-in 변경**: 사용자가 workflow 중 opt-in 변경 가능. 변경 시 audit log + state.json 갱신.
- **N/A 처리**: 우리 workflow 에 N/A 한 rule (HTTP API / Lambda / RDS 등 cloud-specific) 은 baseline 미포함. blocking 아님.
- **v0.7.1+ follow-up**: runtime `evaluate_compliance()` helper 구현. 본 step 7 SCHEMA 형식으로 정합.

## 7. References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.opt-in.md` (20 line, commit `b19c819`)
- 우리 SSOT: `extensions/SCHEMA.md` §3.2 (opt-in 형식)
- 우리 1차 출시: `extensions/security-baseline.md` (242 line, 6 rule, commit `dc2c22b`)
- 우리 follow-up: `extensions/testing-baseline.opt-in.md`, `extensions/performance-baseline.opt-in.md` (동일 형식 차용)
