---
type: concept
status: active
last_ingested_from: workflow-source/core/reverse_engineering.md + workflow-source/reverse-engineering/{01..09}-*.md
related_pages: [concepts/extension-system, concepts/audit-log-standard, concepts/unit-of-work, decisions/adr-004-wiki-layer, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-13
updated: 2026-06-13
---

# Reverse Engineering 9-Artifact (v0.7.0 step 6, AIDLC 차용)

- 문서 목적: standard_ai_workflow v0.7.0 step 6 의 Reverse Engineering 9-Artifact (AIDLC `inception/reverse-engineering.md` 311 line 차용) 의 brownfield 도입 절차.
- 범위: 9 md template + 13 step guide + state.json sync + user approval
- 최종 수정일: 2026-06-13

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/core/reverse_engineering.md` (148 line, v0.7.0 stable) |
| 2 | 9 artifact | `workflow-source/reverse-engineering/01..09-*.md` (~430 line total) |
| 3 | smoke test | `workflow-source/tests/check_reverse_engineering.py` (298 line, 19 test PASS) |
| 4 | Source 1차 출처 | AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, commit `b19c819`, 2026-06-08) |
| 5 | 도입 버전 | v0.7.0 (commit `4bbd391`) |
| 6 | Trigger | Brownfield (기존 project) — Greenfield 는 skip |
| 7 | 관련 토픽 | [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 D |

## §2 왜 Reverse Engineering 9-Artifact 가 필요한가  {#s2-why}

기존 프로젝트 (brownfield) 에 standard_ai_workflow 도입 시, codebase 를 모르는 상태에서 표준을 강제하면 friction 큼. 우리 기존 `repository_assessment.md` (단일 92 line file) 의 한계:
- 주제별 SSOT 부재 (architecture, dependencies, quality 가 한 file 에 섞임)
- link 부재 — 후속 onboarding/handoff 가 참조 어려움

AIDLC 의 inception phase 첫 단계가 **Reverse Engineering 9-Artifact** — 기존 codebase 를 9 md file 로 분석/문서화하여 후속 단계 (requirements, design, code generation) 의 입력으로 사용.

## §3 9-Artifact 구조  {#s3-9-artifact}

구조 100% 유지 (AIDLC 와 동일), 내용만 압축 (~50 line/file):

| # | File | 우리 적응 |
|---|---|---|
| 1 | `01-business-overview.md` | business transaction = workflow stage transition |
| 2 | `02-architecture.md` | components = harness / skill / MCP / workflow_kit |
| 3 | `03-code-structure.md` | key classes = workflow_kit modules (contracts, parser, helper) |
| 4 | `04-api-documentation.md` | REST → MCP tool / Internal → workflow_kit Python |
| 5 | `05-component-inventory.md` | 5 type (Harness/MCP/workflow_kit/Template/Test) |
| 6 | `06-technology-stack.md` | Python + 5 harness + packaging |
| 7 | `07-dependencies.md` | internal + external + lock/checksum |
| 8 | `08-code-quality-assessment.md` | smoke test PASS + R-1~R9 lint + patterns/anti-patterns |
| 9 | `09-reverse-engineering-metadata.md` | ISO 8601 + state.json sync + rerun stale check |

각 artifact 는 `## Verification` subsection (smoke test 와 1:1 매칭) + `Workflow domain 적응` 주석 + AIDLC 1차 출처 cross-reference 명시.

## §4 13 Step Procedure  {#s4-13-step}

| Step | 이름 | 산출물 |
|---|---|---|
| 1 | Multi-Package Discovery | workspace scan + business context + infrastructure + build + service + quality |
| 2-9 | 9 Artifact 생성 | `0N-<name>.md` (위 표) |
| 10 | Timestamp + Metadata | `09-reverse-engineering-metadata.md` 의 ISO 8601 + workspace metadata |
| 11 | State Tracking Sync | `ai-workflow/memory/active/state.json` 의 `reverse_engineering` 필드 |
| 12 | User Message | 요약 + "Approve & Continue" / "Request Changes" 옵션 |
| 13 | User Approval Wait (MANDATORY) | `append_audit_log("user_response", <raw>)` |

## §5 Rerun Stale Check  {#s5-rerun-stale}

```python
# pseudo-code
workspace_last_modified = git_log_max_mtime(workspace_root)
artifact_last_generated = state.json.reverse_engineering.last_generated
if artifact_last_generated and workspace_last_modified <= artifact_last_generated:
    status = "fresh"  # reuse
else:
    status = "stale"  # regenerate
```

| 조건 | 동작 |
|---|---|
| workspace 변경 없음 | Skip (artifact reuse) |
| workspace 변경 감지 | Re-run (regenerate) |
| 사용자 explicit rerun | 무조건 re-run |

## §6 우리 사용 패턴 적응  {#s6-adaptation}

| AIDLC 패턴 | 우리 적응 | 비고 |
|---|---|---|
| 9 artifact 구조 | 동일 (구조 100% 유지) | AIDLC 의 1차 출처 1:1 |
| HTTP API 중심 분석 | workflow API (MCP + workflow_kit) | HTTP API 없음 |
| Lambda / ECS / DynamoDB | harness / MCP / state.json | 인프라 = local runtime |
| CDK / Terraform | pyproject.toml + harness config | packaging 차원 |
| Business transaction = HTTP request | business transaction = workflow stage transition | 동일 추상화 |

**N/A 처리**: 우리 workflow 는 greenfield 또는 in-house. AIDLC 의 multi-cloud brownfield 시나리오 (CDK + Lambda + RDS) 와 차이 큼. 9 artifact 구조는 유지, 내용만 우리 도메인으로 적응.

## §7 한계 / 예외  {#s7-limitations}

- **Greenfield**: artifact 0개. `state.json.reverse_engineering.status = "skipped"` 만 기록
- **AIDLC 와 차이**: 우리 business transaction = workflow stage. AIDLC 의 7 inception stage 와 1:1 매핑 아님 — 단순화
- **v0.7.1+ follow-up**: `existing_project_onboarding.py` 가 9 artifact 자동 fill (현재 단일 `repository_assessment.md` 만 자동)

## §8 Follow-up (v0.7.1+)  {#s8-followup}

- `existing_project_onboarding.py` 가 9 artifact 자동 fill
- v0.7.1: SEC-WF-05 dependency integrity 검증 — `07-dependencies.md` 의 lock file + checksum 자동 확인
- v0.7.1: 9-Artifact 별 wiki L1 page (각 artifact 마다 L1 topic page + L2 sources/)
- v0.8.0: Artifact 별 version diff (이전 reverse engineering 대비 변경점)

## §9 References  {#s9-references}

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, commit `b19c819`, 2026-06-08)
- 우리 SSOT: `workflow-source/core/reverse_engineering.md` (148 line)
- 우리 9 artifact: `workflow-source/reverse-engineering/{01..09}-*.md`
- 우리 검증: `workflow-source/tests/check_reverse_engineering.py` (19 test PASS)
- 우리 wiki: [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 D
