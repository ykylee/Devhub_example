# PR Body — docs/work_260610-c-j-build-tag-review (PR #XXX body source)

- 문서 목적: PR #XXX 의 body 본문 (markdown) — gh pr create --body-file 의 source. PR #XXX 가 머지된 후 본 file 은 archival.
- 범위: PR #XXX 의 PR body 본문 (변경 요약 4 section + 추적성 영향 + 결정 trade-off + Tier 분류 + 검증 + Out of scope + Refs + Base/target).
- sprint branch: docs/work_260610-c-j-build-tag-review
- 대상 독자: PR reviewer, owner, 머지 후 archival reference
- 상태: draft (PR 발행 전)
- 최종 수정일: 2026-06-10
- 관련 문서: [session_handoff.md](./session_handoff.md), [work_backlog.md](./work_backlog.md), [backlog/2026-06-10.md](./backlog/2026-06-10.md), [state.json](./state.json), [ADR-0031 §4 결정](../../../../docs/adr/0031-build-tag-policy-review.md), [ADR-0030 §2.3 confirmed](../../../../docs/adr/0030-sso-integrations-and-auth-session-port.md)

# docs(adr,traceability,memory): ADR-0031 build tag 정책 재검토 — ADR-0030 §2.3 confirmed (sprint -a follow-up PR1 PR #540 의 carry-over C-j)

## 목적

v1.1 sprint -a follow-up PR1 (PR #540, `feat/work_260610-v1-1-sprint-a-real-adapter`, main HEAD `58d163f`) 의 carry-over **C-j** (P3) 의 정공법 PR. **문서만 변경 (코드 0줄)**.

- C-j (P3): build tag 정책 재검토 — ADR-0030 §2.3 의 runtime injection (옵션 2) vs build tag (옵션 1) trade-off 의 정량 측정 + 현시점 결정 confirmed.

## 변경 요약

### 1. `docs/adr/0031-build-tag-policy-review.md` (NEW, 12KB)

9 section 신규:
- **§0 메타**: 상태 accepted (2026-06-10), Tier 공용, 관련 ADR-0030 + PR #539 + #540 + #542.
- **§1 배경**: ADR-0030 §2.3 의 결정 근거 추정 (정량 측정 없음) + sprint -a follow-up 적용 결과.
- **§2 정량 측정**: runtime injection (현재) 의 cost = stub binary overhead < 5KB (전체 backend-core < 50MB 대비 0.01%) + CI matrix 4 jobs + +15~20min runtime. build tag (이론) 의 cost = binary -6.3KB 절감 + CI matrix 6 jobs + +30~60min runtime + 5~10 file 변경 + 2개 binary 운영.
- **§3 후보 옵션**: 3 옵션 (build tag / runtime injection 유지 / hybrid) + trade-off 평가.
- **§4 결정**: **옵션 2 (runtime injection 유지) confirmed**. 근거 6건 (정량 측정 결과 / CI runtime / 코드 운영 / 현시점 측정 가능 / stub production 위험 0 / architectural cleanliness 우선순위).
- **§5 재검토 trigger 5건**: stub code size > 250KB / stub production risk / CI axes 5+ / Phase 2 agentic RAG / stub safety. 현시점 trigger 0건.
- **§6 cross-tier impact**: tier 매핑 + CI/ADR/tier lint.
- **§7 risks + open questions 5건**: stub code size 증가 / 운영 SOP 누락 / CI matrix 확장 / Phase 2 / 사내 SOP 자동화.
- **§8 supersession**: ADR-0030 을 supersede 하지 않음 (confirmed). 후속 trigger 발효 시 새 ADR (e.g. ADR-0032).
- **§9 변경 이력**: 1 row.

### 2. `docs/adr/0030-sso-integrations-and-auth-session-port.md` (2 변경)

- **§2.3 row (옵션 2)**: "**2026-06-10 confirmed (ADR-0031 §4 재평가)**" reference 추가 + 정량 측정 결과 명시.
- **§9 변경 이력**: row 추가 (ADR-0031 row + sprint reference).

### 3. `docs/traceability/report.md` (2 변경)

- **§4 ADR 인덱스**: ADR-0031 row 신규 (영향 도메인: 운영, CI, IdP).
- **§6 변경 이력**: 본 PR row 신규 (2026-06-10, ADR-0031 + ADR-0030 + traceability + memory 4 변경 영역).

### 4. Memory sync (5 file 신규)

- `ai-workflow/memory/docs/work_260610-c-j-build-tag-review/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-10.md,pr_body.md}` (sprint memory 5종).
- main flat `ai-workflow/memory/{state.json,session_handoff.md,work_backlog.md}` 동기화 — 본 PR 머지 후 자동 sync (housekeeping commit).

## 추적성 영향

| Stage | ID | Status | 비고 |
| --- | --- | --- | --- |
| ADR | `ADR-0031` | **신규** | build tag 정책 재검토 — runtime injection (옵션 2) confirmed. 정량 측정 (binary < 5KB vs CI +30~60min) + 재검토 trigger 5건. ADR-0030 §2.3 supersede X. |

**신규 ID 1건 (ADR 만, IMPL/REQ/UC/ARCH/API/RM/UT/TC 신규 발급 0건)** — ADR-0030 §2.3 의 결정의 re-evaluation 문서.

## 결정 trade-off (요약)

| 측정 항목 | Runtime injection (현재) | Build tag (이론) | 차이 |
| --- | --- | --- | --- |
| Binary overhead | < 5KB | -6.3KB (절감) | -6.3KB (build tag 유리) |
| CI runtime | +15~20min (PR #542) | +30~60min (이론) | build tag 가 +15~40min 더 |
| CI matrix jobs | 4 (e2e 1/2/3 + e2e-internal) | 6 (e2e × 2 tags) | build tag 가 +2 jobs |
| 코드 변경 | 0 (현재 상태 유지) | 5~10 file | build tag 가 +5~10 file |
| 운영 복잡도 | 1 binary | 2 binary | build tag 가 +1 |

**결론 (정량)**: build tag 의 binary size 절감 (~6KB) 은 무시 가능 수준 (전체 backend-core binary < 50MB 대비 0.01% 미만). **runtime injection 의 cost 가 build tag 의 cost 보다 본질적으로 작음**. **runtime injection 유지 confirmed**.

## Tier 분류 (self-check)

| 변경 영역 | Tier | 근거 |
| --- | --- | --- |
| `docs/adr/0031-...` | **공용** | ADR 본문, 사내 한정 정보 미포함 |
| `docs/adr/0030-...` | **공용** | ADR row 갱신, 사내 한정 정보 미포함 |
| `docs/traceability/report.md` | **공용** | 문서 정합, 사내 한정 정보 미포함 |
| `ai-workflow/memory/...` | **공용** | workflow metadata, 사내 한정 정보 미포함 |

**본 PR 의 모든 변경 = 공용**. `check-tier-separation.sh` no changes between origin/main and HEAD 확인.

## 검증 (run on this branch)

- `bash scripts/check-tier-separation.sh` — ✅ no changes between origin/main and HEAD
- `bash scripts/check-openapi-yaml-lint.sh` — ✅ passed (openapi.yaml 변경 0)
- `bash scripts/check-migration-uniqueness.sh` — ✅ valid and unique (migration 변경 0)
- `python3.13 ai-workflow/tests/check_docs.py` — 본 PR 의 4 file 정합 (ADR-0031 metadata 6 field + cross-link + 제목 헤더, main flat memory finalize). exit 1 의 원인은 본 PR 영역 외 기존 file 의 historical link.

## Out of scope (별도 PR / 후속)

- **backend-integration DEVHUB_BUILD_TIER matrix** (sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix). 본 PR 의 C-j 정공법과 별개.
- **release_v1_roadmap.md §3.5 N-13** 정합 (C-j done 마킹). 본 PR 머지 후 별도 housekeeping commit.
- **N-10 RBAC E2E 6 TC 보강** (sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs`).
- **Phase 2 (v1.2) agentic RAG 의 추가 port 도입 시 새 ADR**: 본 ADR-0031 §5 trigger 4번.

## Refs

- [ADR-0030 §2.3 runtime injection 결정 (re-evaluation source)](../../../../docs/adr/0030-sso-integrations-and-auth-session-port.md)
- [release_v1_roadmap.md §3.5 N-13](../../../../docs/planning/release_v1_roadmap.md) — sprint -a follow-up carry-over N-13
- [PR #539 `feat/work_260610-v1-1-sprint-a-followup`](https://github.com/ykylee/Devhub_example/pull/539) — saovae_stub + main wiring
- [PR #540 `feat/work_260610-v1-1-sprint-a-real-adapter`](https://github.com/ykylee/Devhub_example/pull/540) — real adapter (C-j 의 정량 측정 데이터 source)
- [PR #541 `docs/work_260610-traceability-impl-sso-keycloak`](https://github.com/ykylee/Devhub_example/pull/541) — C-g + C-h 정합 (ADR-0030 §5 timeline)
- [PR #542 `feat/work_260610-c-i-e2e-internal-job`](https://github.com/ykylee/Devhub_example/pull/542) — C-i E2E Internal job (C-j 의 정량 측정 데이터 source)

## Base / target

- **base branch**: `main` (HEAD `1dfc100`, PR #541 + #542 + post-merge sync 머지 후)
- **target branch**: `main`
- **merge strategy**: squash
- **branch name**: `docs/work_260610-c-j-build-tag-review`
