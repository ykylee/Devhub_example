---
type: topic
status: active
last_ingested_from: workflow-source/reverse-engineering/{01..09}-*.md + workflow-source/core/reverse_engineering.md
related_pages: [concepts/reverse-engineering, concepts/unit-of-work, concepts/audit-log-standard, concepts/extension-system, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-13
updated: 2026-06-13
---

# Reverse Engineering 9-Artifact Index (v0.7.1, brownfield project SSOT)

- 문서 목적: standard_ai_workflow v0.7.1 의 9-Artifact (brownfield project 의 주제별 SSOT) 의 *index* page. 각 artifact 의 본문은 `workflow-source/reverse-engineering/0N-*.md` 에 위치.
- 범위: 9 artifact index + auto-fill helper (v0.7.1 follow-up) + R-1~R9 lint 정합
- 최종 수정일: 2026-06-13

## §1 9-Artifact Index

| # | Artifact | 본 위치 | 주제 |
|---|---|---|---|
| 1 | Business Overview | `workflow-source/reverse-engineering/01-business-overview.md` | business transaction = workflow stage transition / business dictionary = workflow vocabulary |
| 2 | System Architecture | `workflow-source/reverse-engineering/02-architecture.md` | components = harness / skill / MCP / workflow_kit |
| 3 | Code Structure | `workflow-source/reverse-engineering/03-code-structure.md` | key classes = workflow_kit modules (contracts, parser, helper) |
| 4 | API Documentation | `workflow-source/reverse-engineering/04-api-documentation.md` | REST → MCP tool / Internal → workflow_kit Python |
| 5 | Component Inventory | `workflow-source/reverse-engineering/05-component-inventory.md` | 5 type (Harness/MCP/workflow_kit/Template/Test) |
| 6 | Technology Stack | `workflow-source/reverse-engineering/06-technology-stack.md` | Python + 5 harness + packaging |
| 7 | Dependencies | `workflow-source/reverse-engineering/07-dependencies.md` | internal + external + lock/checksum (SEC-WF-05) |
| 8 | Code Quality Assessment | `workflow-source/reverse-engineering/08-code-quality-assessment.md` | smoke test PASS + R-1~R9 lint + patterns/anti-patterns |
| 9 | Reverse Engineering Metadata | `workflow-source/reverse-engineering/09-reverse-engineering-metadata.md` | ISO 8601 + state.json sync + rerun stale check |

## §2 각 Artifact 의 Verification Subsection

각 artifact 의 `## Verification` subsection 은 smoke test 와 1:1 매핑. workflow_kit 의 artifact 검증 helper 가 자동 확인.

## §3 Auto-Fill Helper (v0.7.1 follow-up)

`workflow-source/tools/fill_reverse_engineering_artifacts.py`:
- workflow-source/reverse-engineering/ 의 template 사용
- `--info=<json>` 또는 `--project-root=<path>` 입력
- heuristic 기반 TODO marker 채움 (사용자가 직접 완성)
- 9 artifact 자동 emit → brownfield project 의 9-Artifact 즉시 사용 가능

## §4 13 Step Procedure (workflow-source/core/reverse_engineering.md)

1. Multi-Package Discovery
2-9. 9 Artifact 생성 (auto-fill helper + 사용자 완성)
10. Timestamp + Metadata (ISO 8601)
11. State Tracking Sync (state.json)
12. User Message (요약 + "Approve" / "Request Changes")
13. User Approval Wait (MANDATORY, audit log + Question Format)

## §5 Rerun Stale Check

- workspace_last_modified > artifact_last_generated → stale (regenerate)
- workspace_last_modified ≤ artifact_last_generated → fresh (reuse)
- 사용자 explicit rerun → 무조건 regenerate

## §6 우리 사용 패턴 적응

- HTTP API 중심 분석 → workflow API (MCP + workflow_kit)
- Lambda / ECS / DynamoDB → harness / MCP / state.json
- Business transaction = HTTP request → workflow stage transition
- 9 artifact 구조 100% 유지 (AIDLC 의 1차 출처 1:1)

## §7 한계 / 예외

- **Greenfield**: artifact 0개. `state.json.reverse_engineering.status = "skipped"` 만 기록
- **AIDLC 와 차이**: 우리 business transaction = workflow stage. 7 inception stage 와 1:1 매핑 아님
- **v0.7.1 follow-up**: auto-fill helper (heuristic 기반 TODO marker), 100% 자동 생성 ❌, *초안 + 사용자 fill* 패턴 ✅

## §8 References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, commit `b19c819`, 2026-06-08)
- 우리 SSOT: `workflow-source/core/reverse_engineering.md` (148 line, commit `4bbd391`)
- 우리 9 artifact: `workflow-source/reverse-engineering/{01..09}-*.md`
- 우리 helper: `workflow-source/tools/fill_reverse_engineering_artifacts.py` (v0.7.1 follow-up, 200 line)
- 우리 검증: `workflow-source/tests/check_reverse_engineering.py` (19 test PASS)
- 우리 위키: [[concepts/reverse-engineering]] (v0.7.0 step 6)
- 우리 위키: [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 D
