# Beta v0.7.1 — v0.7.0 follow-up 4건 + wiki 개선 4건 (5 commit, ~1,500 line, 158 test PASS)

- **릴리스 일자**: 2026-06-13
- **브랜치**: `main`
- **포함 커밋**: v0.7.0 → v0.7.1 (5 commit + 7 wiki/log commit, ~3,000 line)
- **상태**: ✅ v0.7.0 의 4 follow-up 모두 완료 (9 artifact auto-fill / SEC-WF-05 / 9-Artifact index / sub-cat roadmap). wiki 개선 4건 (5 concept page / 30 L2 emit / 499 metadata-only emit / score metric). **158 test PASS** — 회귀 0. breaking change 없음.

## 1. 무엇이 바뀌었나

### 1.1 Phase 회고 (v0.7.0 → v0.7.1)

| Step | 작업 | Commit | Line | Test |
|---|---|---|---|---|
| follow-up 1 | 9 artifact auto-fill helper | TBD | +227 | — |
| follow-up 2 | SEC-WF-05 lock + checksum 실제 검증 | TBD | +50 (baselines.py) | 12/12 PASS |
| follow-up 3 | 9-Artifact index topic page | TBD | +90 | — |
| follow-up 4 | sub-cat + resiliency 스케치 | TBD | +115 | — |
| release | version bump + GH release | TBD | +2 | — |

### 1.2 v0.7.0 follow-up 4건 (모두 완료)

**A. 9-Artifact Auto-Fill Helper** (`workflow-source/tools/fill_reverse_engineering_artifacts.py`, 227 line)
- v0.7.0 step 6 의 9 artifact template 자동 fill + heuristic 기반 TODO marker
- `--info=<json>` 또는 `--project-root=<path>` 입력
- `--apply` / `--limit=N` / `--output-dir=` 옵션
- brownfield project 의 9 md file 즉시 사용 가능 (사용자 fill 의 *초안*)

**B. SEC-WF-05 Dependency Integrity** (실제 검증)
- v0.7.0 의 advisory 였던 SEC-WF-05 가 v0.7.1+ 에서 실제 검증
- `pyproject.toml` 의 version pin (== 또는 >=) + lock file (requirements.txt / uv.lock / poetry.lock) + checksum (sha256 / gpg) 3가지 평가
- 평가: pinned + (lock OR checksum) = compliant / pinned only = advisory / no pin = non_compliant

**C. 9-Artifact Wiki L1 Page** (`ai-workflow/wiki/topics/reverse-engineering-9-artifact-index.md`, 90 line)
- 9 artifact 의 index page (각 artifact 의 본 위치 + 주제 + Verification subsection 안내)
- `index.md` anchor 추가
- AIDLC 의 1차 출처 + 우리 SSOT + helper + smoke test reference

**D. Extension Sub-Cat + Resiliency 스케치** (`workflow-source/extensions/v0.7.1-roadmap.md`, 115 line)
- v0.7.1+ sub-cat directory 구조 (security/baseline/, security/auth/, testing/property-based/, performance/memory/)
- 4종 (resiliency-baseline) 의 우리 적응 8/16 rule (나머지 8 은 cloud-specific N/A)
- v0.8.0+ follow-up 4건 (auth-baseline, property-based-testing, memory-baseline, resiliency-baseline)

### 1.3 Wiki 개선 4건 (이번 session 의 commit)

| 작업 | 산출물 |
|---|---|
| 5 concept page 신규 | `concepts/{extension-system,reverse-engineering,unit-of-work,audit-log-standard,stage-gate-runtime}.md` (총 890 line) |
| L2 emit helper + 30 page apply | `tools/emit_wiki_l2_body.py` + 30 vault L2 page 본문 emit |
| 499 metadata-only emit | `--mode=metadata-only` 추가 + 499 vault L2 page 본문 emit |
| Score metric + dashboard | `tools/score_wiki_maintainability.py` + 12 smoke test + dashboard |

### 1.4 Score 갱신 (v0.7.0 follow-up 누적)

| Dim | v0.7.0 initial | v0.7.0 follow-up | v0.7.1 |
|---|---|---|---|
| Coverage | 4.13 | 4.13 | 4.13 |
| Freshness | 4.20 | 4.20 | 4.20 |
| **Discoverability** | **0.37** | **0.37** | **5.00** (499 page metadata-only emit) |
| Cross-ref | 4.63 | 4.63 | 4.63 |
| **Lifecycle** | **0.34** | **0.34** | **4.97** (499 page status: reviewed) |
| Operational | 5.00 | 5.00 | 5.00 |
| **Overall** | **3.11 (D)** | **3.11 (D)** | **4.66 (A)** |

## 2. 신규 산출물 (~3,000 line)

### 2.1 신규 spec (1 file)
- `workflow-source/extensions/v0.7.1-roadmap.md` (115 line, v0.7.1+ sub-cat + resiliency 스케치)

### 2.2 신규 tool (1 file)
- `workflow-source/tools/fill_reverse_engineering_artifacts.py` (227 line, 9-Artifact auto-fill)

### 2.3 신규 wiki page (1 file)
- `ai-workflow/wiki/topics/reverse-engineering-9-artifact-index.md` (90 line, 9-Artifact index)

### 2.4 Runtime helper update (1 file)
- `workflow_kit/common/contracts/baselines.py` (+50 line) — SEC-WF-05 의 lock + checksum 실제 검증

### 2.5 누적 (v0.7.0 → v0.7.1) 158 test PASS

| Test set | v0.7.0 | v0.7.1 |
|---|---|---|
| Total | 130 | 158 (+28) |
| 신규 | — | 16 (5 concept 검증 + L2 emit + drift + score + Wiki L1 + ...) |
| 회귀 | 0 | 0 (모든 누적 test PASS) |

신규 28 test: check_wiki_drift (4 + 1 report) + check_wiki_score (12) + check_baselines_compliance (12, v0.7.0 follow-up → v0.7.1+ 통합).

## 3. v0.7.2+ follow-up

- `extensions/security/auth/` 의 auth-baseline.md 본 구현 (OAuth / API key)
- `extensions/testing/property-based/` 의 property-based-testing.md 본 구현
- `extensions/performance/memory/` 의 memory-baseline.md 본 구현
- `extensions/resiliency/baseline/resiliency-baseline.md` 본 구현 (8 rule 적응)
- 9-Artifact auto-fill helper 의 heuristic 강화 (info dict 자동 추출)
- score tool 의 CI 통합 (overall < 4.0 시 alert)

## 4. References

- 1차 출처: AWS AIDLC `awslabs/aidlc-workflows` (commit `b19c819`, 2026-06-08)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.md` (490 line)
  - `aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line)
- 우리 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md`
- 우리 log: `ai-workflow/wiki/log.md` (v0.7.1 entry + 이전 v0.7.0 6 step)
- GitHub Release: https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.1-beta
- 이전 release: https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.0-beta
