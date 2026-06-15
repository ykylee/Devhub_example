# Auth Baseline Opt-In (v0.7.2+)

- 문서 목적: workflow 시작 시 사용자 opt-in prompt (auth-baseline)
- 상태: stable (v0.7.2 도입)
- 최종 수정일: 2026-06-13
- 관련 문서: [`./auth-baseline.md`](./auth-baseline.md), [`../SCHEMA.md`](../SCHEMA.md)

## 1. Question (workflow-start 시 1회 표시)

> Question: Should auth baseline extension rules be enforced for this project?
>
> A) Yes — enforce all SEC-AUTH rules as blocking constraints (recommended for production-grade)
>
> B) No — skip all SEC-AUTH rules (suitable for PoCs, prototypes, experimental projects)
>
> P) Partial — enforce only SEC-AUTH-01 (API Key Storage) + SEC-AUTH-06 (Audit Log)
>
> X) Other (please describe after [Answer]: tag below)
>
> [Answer]:

## 2. Response Processing

답변은 `ai-workflow/memory/active/auth_baseline_status.md` 에 기록:

| 답변 | 기록 |
|---|---|
| `A) Yes` | `status: enabled, partial: false, partial_rules: []` |
| `B) No` | `status: disabled, partial: false, partial_rules: []` |
| `P) Partial` | `status: enabled, partial: true, partial_rules: [SEC-AUTH-01, SEC-AUTH-06]` |
| `X) ...` | 사용자 의도 분석 후 A/B/P 중 하나에 매핑 |

## 3. State File Schema

```yaml
auth_baseline:
  status: enabled | disabled | partial
  partial_rules: [SEC-AUTH-01, SEC-AUTH-06]  # partial 시만
  decided_at: "2026-06-13T22:50:00+09:00"  # ISO 8601
  decided_by: "yklee"  # 또는 "default: A" (Greenfield)
  standard_ref: "AIDLC extensions/security/baseline/security-baseline.md @ b19c819"
```

## 4. Default 정책

| 모드 | 기본값 |
|---|---|
| Greenfield (신규) | A) Yes (production 권장) |
| Brownfield (기존) | 기존 status 존중 (없으면 B) No 후 점진 enable) |
| CI/CD | opt-in prompt 표시 안 함, 기존 `auth_baseline_status.md` 값 사용 |
| PoC / Prototype | B) No 권장 |

## 5. Workflow 통합

### 5.1 session-start (workflow 시작 시)

3종 sub-cat baseline (security-auth / testing-pbt / performance-memory) opt-in prompt 동시 표시.

### 5.2 stage completion (stage 완료 시)

- status = enabled → stage completion message 에 "Auth Compliance" 섹션
- 6 rule 마다 compliant / non-compliant / N/A 표시
- partial mode: partial_rules 만 hard constraint, 나머지 advisory

### 5.3 audit log 통합

- opt-in 결정 시 `append_audit_log("opt_in", "auth_baseline=A")`
- stage completion 의 compliance summary 가 audit.md 에 기록

## 6. 한계 / 예외

- **Opt-in 변경**: 사용자가 workflow 중 opt-in 변경 가능. 변경 시 audit log + state.json 갱신.
- **N/A 처리**: 우리 workflow 에 N/A 한 rule (e.g. cross-account role) 은 baseline 미포함. blocking 아님.
- **v0.7.2+ runtime**: workflow_kit.common.auth.evaluate_compliance() 구현 (8 smoke test)

## 7. References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line)
- 우리 SSOT: `extensions/SCHEMA.md` §3.2 (opt-in 형식)
- 우리 1차 출시: `extensions/security/auth/auth-baseline.md` (6 rule, 200 line)
- 우리 follow-up: `extensions/security-baseline.opt-in.md`, `extensions/testing-baseline.opt-in.md`, `extensions/performance-baseline.opt-in.md`
