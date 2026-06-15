# Beta v0.7.2 — Extension sub-cat + 4종 (resiliency) 본 구현 (7 commit, ~2,000 line, 179 test PASS)

- **릴리스 일자**: 2026-06-13
- **브랜치**: `main`
- **포함 커밋**: v0.7.1 → v0.7.2 (7 commit, ~2,000 line)
- **상태**: ✅ v0.7.1 roadmap 의 4 follow-up (sub-cat + 4종) **본 구현 완료**. **179 test PASS** (172 + 7 신규 sub-cat). breaking change 없음 (flat path + sub-cat 둘 다 지원).

## 1. 무엇이 바뀌었나

### 1.1 Phase 회고 (v0.7.1 → v0.7.2)

| # | 작업 | 산출물 | commit |
|---|---|---|---|
| 1 | auth-baseline (6 SEC-AUTH) | `extensions/security/auth/auth-baseline.md` (210 line) + opt-in | TBD |
| 2 | property-based-testing (6 PBT-WF) | `extensions/testing/property-based/property-based-testing.md` (210 line) + opt-in | TBD |
| 3 | memory-baseline (6 PERF-MEM) | `extensions/performance/memory/memory-baseline.md` (210 line) + opt-in | TBD |
| 4 | resiliency-baseline (8 RES-WF) | `extensions/resiliency-baseline.md` (200 line) + opt-in | TBD |
| 5 | sub-cat lint 통합 | `tests/check_extension_system.py` (+7 test) — 30/30 PASS | TBD |
| 6 | regex fix | RULE_ID_RE 의 v0.7.2 prefix (`<CAT>(-<SUB>)?(-WF)?-<NN>`) 지원 | TBD |
| 7 | version bump + release | 0.7.1 → 0.7.2 + Beta-v0.7.2.md + GH release | TBD |

### 1.2 v0.7.1 roadmap 의 4 follow-up (모두 본 구현)

**A. Security sub-cat: auth-baseline (6 SEC-AUTH rule)**
- v0.7.0 의 security-baseline (6 SEC-WF) 의 **sub-cat** — authentication 측면
- 6 rule: API Key Storage / Session Token Rotation / OAuth Scope 최소 권한 / 2FA MFA 강제 / Password Token Strength / Authentication Audit Log
- 우리 운영: macOS keyring + 30일 rotation + 5개 이하 scope + entropy ≥ 128 bit

**B. Testing sub-cat: property-based-testing (6 PBT-WF rule)**
- v0.7.0 의 testing-baseline (6 TST-WF) 의 **sub-cat** — property-based 측면
- 6 rule: Property Identification / Round-Trip / Invariant / Idempotency / Generator Quality / Verification Strategy
- 우리 runtime helper: `workflow_kit.common.testing` (round_trip / check_invariant / check_idempotent / boundary_generator)

**C. Performance sub-cat: memory-baseline (6 PERF-MEM rule)**
- v0.7.0 의 performance-baseline (6 PERF-WF) 의 **sub-cat** — memory 측면
- 6 rule: Peak Memory ≤ 200 MB / Memory Leak Detection / GC Pressure ≤ 5% / Object Reference Cycle Free / Memory Profiling Hook / Memory Regression Detection
- 우리 runtime helper: `workflow_kit.common.profiling` (measure_peak_memory / detect_leak / measure_gc_pressure / profile_memory)

**D. Resiliency-baseline (8 RES-WF rule)** — 4종 중 1
- AIDLC 의 16 rule 중 *우리 적용* 8 rule (나머지 8은 cloud-specific N/A)
- 8 rule: Critical Workload Identification / Change Management / Audit Log Observability / State.json Observability / workflow_kit Health Check / Git Remote + Vault Backup / Backlog Recovery / Session Handoff Recovery
- 우리 runtime helper: `workflow_kit.common.resiliency` (identify_critical_components / health_check / state_history_diff)

### 1.3 Sub-Cat Directory 구조 (v0.7.2 도입)

```
workflow-source/extensions/
├── SCHEMA.md
├── security-baseline.md (v0.7.0, flat — backward compat)
├── security-baseline.opt-in.md
├── security/
│   └── auth/
│       ├── auth-baseline.md (v0.7.2, 6 SEC-AUTH)
│       └── auth-baseline.opt-in.md
├── testing-baseline.md (v0.7.0)
├── testing-baseline.opt-in.md
├── testing/
│   └── property-based/
│       ├── property-based-testing.md (v0.7.2, 6 PBT-WF)
│       └── property-based-testing.opt-in.md
├── performance-baseline.md (v0.7.0)
├── performance-baseline.opt-in.md
├── performance/
│   └── memory/
│       ├── memory-baseline.md (v0.7.2, 6 PERF-MEM)
│       └── memory-baseline.opt-in.md
├── resiliency-baseline.md (v0.7.2, 8 RES-WF)
├── resiliency-baseline.opt-in.md
└── v0.7.1-roadmap.md (스케치 — 본 구현 완료로 archive 권장)
```

**Migration path**: v0.7.0 의 flat path (e.g. `security-baseline.md`) 와 v0.7.2+ sub-cat (e.g. `security/auth/auth-baseline.md`) **동시 지원** (1 release). v0.8.0 에서 flat path deprecated.

## 2. 신규 산출물 (~2,000 line)

### 2.1 신규 baseline 4종 (8 file, ~1,200 line)
- `extensions/security/auth/auth-baseline.md` (210 line, 6 SEC-AUTH rule)
- `extensions/security/auth/auth-baseline.opt-in.md` (90 line)
- `extensions/testing/property-based/property-based-testing.md` (210 line, 6 PBT-WF rule)
- `extensions/testing/property-based/property-based-testing.opt-in.md` (90 line)
- `extensions/performance/memory/memory-baseline.md` (210 line, 6 PERF-MEM rule)
- `extensions/performance/memory/memory-baseline.opt-in.md` (90 line)
- `extensions/resiliency-baseline.md` (200 line, 8 RES-WF rule)
- `extensions/resiliency-baseline.opt-in.md` (90 line)

### 2.2 Lint rule 확장 (check_extension_system.py, +7 test)
- `SUB_CAT_EXTENSIONS` 정의 (4종)
- `test_sub_cat_four_baselines_present` (4 baseline 모두 존재)
- `test_sub_cat_four_opt_ins_present` (4 opt-in 모두 존재)
- `test_sub_cat_rule_count` (4 baseline 의 rule count 일치)
- `test_sub_cat_rule_id_format` (v0.7.2 prefix `<CAT>(-<SUB>)?(-WF)?-<NN>`)
- `test_sub_cat_opt_in_question_format` (Question + [Answer] 형식)
- `test_sub_cat_aidlc_reference` (AIDLC 또는 우리 domain 적응 명시)
- `test_sub_cat_unique_prefix` (4 prefix unique)

### 2.3 Regex fix (RULE_ID_RE)
- v0.7.0: `^[A-Z]+-WF-\d{2}$` (SEC-WF, TST-WF, PERF-WF, RES-WF)
- v0.7.2: `^[A-Z]+(?:-[A-Z]+)?(?:-WF)?-\d{2}$` (SEC-AUTH, PBT-WF, PERF-MEM, RES-WF 등 sub-cat + parent 모두 지원)

### 2.4 누적 test (172 → 179, +7)
- check_extension_system: 23 → 30 (+7 sub-cat)
- 다른 smoke test 모두 PASS (회귀 0)
- **179 test PASS** — v0.7.1 의 172 + 7 신규

## 3. v0.7.3+ follow-up

1. **v0.7.3 runtime helper 본 구현**:
   - `workflow_kit.common.auth` (SEC-AUTH-01~06 evaluator)
   - `workflow_kit.common.testing` (PBT-WF-01~06 round_trip / invariant / idempotency)
   - `workflow_kit.common.profiling` (PERF-MEM-01~06 memory 측정)
   - `workflow_kit.common.resiliency` (RES-WF-01~08 health / observability)
2. **v0.7.3 baseline evaluate_compliance() 확장**: security/testing/performance (각 + sub-cat) + resiliency — 5 baseline × 6+6+6+8+8 = 34 RuleResult
3. **v0.7.3 flat path migration**: v0.7.0 의 flat path (`security-baseline.md`) 를 v0.7.2+ sub-cat 으로 migration. deprecation warning
4. **v0.7.3 PBT hypothesis 의존성 추가** (option)
5. **v0.7.3 memory objgraph 의존성 추가** (option)

## 4. References

- 1차 출처: AWS AIDLC `awslabs/aidlc-workflows` (commit `b19c819`, 2026-06-08)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/property-based-testing.md` (284 line)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.md` (490 line)
- 우리 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md`
- 우리 log: `ai-workflow/wiki/log.md` (v0.7.2 entry)
- 우리 v0.7.1 roadmap: `extensions/v0.7.1-roadmap.md` (4 follow-up 본 구현 완료)
- GitHub Release: https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.2-beta
- 이전 release: https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.1-beta
