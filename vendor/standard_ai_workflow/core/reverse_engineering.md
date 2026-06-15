# Reverse Engineering 가이드 (v0.7.0 step 6)

- 문서 목적: brownfield (기존) 프로젝트에 standard_ai_workflow v0.7.0+ 도입 시, codebase 분석 결과를 9-Artifact 형식으로 자동 생성하는 절차를 정의한다.
- 범위: brownfield project detection, 9 artifact 작성, state tracking, user approval, rerun policy
- 상태: draft (v0.7.0 step 6)
- 최종 수정일: 2026-06-12
- 1차 출처: AIDLC `aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, commit b19c819, 2026-06-08)
- 관련 문서: `./global_workflow_standard.md`, `../templates/repository_assessment_template.md`, `../reverse-engineering/{01..09}-*.md` (9 artifact template)

## 1. 왜 필요한가 (Why)

기존 프로젝트 (brownfield) 에 standard_ai_workflow 도입 시, code-base 를 모르는 상태에서 표준을 강제하면 friction 큼. AIDLC 의 inception phase 첫 단계가 "reverse engineering" — 기존 codebase 를 9-Artifact 형식으로 분석/문서화하여 후속 단계 (requirements, design, code generation) 의 입력으로 사용.

우리 적응: `existing` mode 의 산출물 (현재 단일 `repository_assessment.md`) 을 9개 artifact 로 분화 → 각 주제별 SSOT 확보 + L1 wiki / 후속 onboarding / handoff 가 artifact 별로 link 가능.

## 2. Execute / Skip 조건

| 조건 | 동작 |
|---|---|
| **Brownfield** (기존 code 발견) | Step 1-13 전체 실행 |
| **Greenfield** (no existing code) | Skip — `state.json.reverse_engineering.status = "skipped"` 기록 |
| **Rerun**: 기존 artifact 존재 + workspace 변경 없음 | Skip (artifact reuse) |
| **Rerun**: 기존 artifact 존재 + workspace 변경 감지 | Re-run (regenerate) |
| **사용자 explicit rerun 요청** | 무조건 re-run |

## 3. Rerun Stale Check

```python
# pseudo-code
workspace_last_modified = git_log_max_mtime(workspace_root)
artifact_last_generated = state.json.reverse_engineering.last_generated
if artifact_last_generated and workspace_last_modified <= artifact_last_generated:
    status = "fresh"  # reuse
else:
    status = "stale"  # regenerate
```

## 4. 9-Artifact 구조

`workflow-source/reverse-engineering/` 디렉토리에 9 file 생성:

| # | File | AIDLC 대응 | 우리 적응 |
|---|---|---|---|
| 1 | `01-business-overview.md` | business-overview | workflow stage transition = business transaction |
| 2 | `02-architecture.md` | architecture | harness / skill / MCP / workflow_kit |
| 3 | `03-code-structure.md` | code-structure | workflow_kit module hierarchy |
| 4 | `04-api-documentation.md` | api-documentation | REST → MCP tool / Internal → workflow_kit Python |
| 5 | `05-component-inventory.md` | component-inventory | Harness/MCP/workflow_kit/Template/Test 5 type |
| 6 | `06-technology-stack.md` | technology-stack | Python + 5 harness runtime + packaging |
| 7 | `07-dependencies.md` | dependencies | internal workflow_kit + pyproject external |
| 8 | `08-code-quality-assessment.md` | code-quality-assessment | smoke test PASS + R-1~R9 lint |
| 9 | `09-reverse-engineering-metadata.md` | reverse-engineering-timestamp | ISO 8601 + state.json 동기화 |

각 artifact 는 `## Verification` subsection 을 가지며, smoke test 가 검증.

## 5. Step-by-Step Procedure

### Step 1: Multi-Package Discovery

- 1.1 workspace scan — top-level dir + sub-package
- 1.2 business context 파악 — "이 시스템이 하는 일" 1-2 문장
- 1.3 infrastructure 발견 — harness config, MCP config, pyproject.toml
- 1.4 build system — pyproject / package.json / Makefile
- 1.5 service architecture — MCP tool interface + workflow_kit module boundary
- 1.6 code quality — smoke test count + lint 설정

### Step 2-9: 9-Artifact 생성

각 step 은 대응되는 `0N-*.md` template 을 copy + placeholder 채우기. 각 artifact 끝의 `## Verification` 가 smoke test 와 1:1 매핑.

### Step 10: Timestamp + Metadata

`09-reverse-engineering-metadata.md` 작성:
- Analysis Date = ISO 8601
- Analyzer = 작업자 (human/AI agent)
- Workspace = 절대경로
- Total Files Analyzed = N
- 8 artifact 모두 `[x]` checkbox

### Step 11: State Tracking Sync

`ai-workflow/memory/active/state.json` 의 `reverse_engineering` 필드 갱신:

```json
{
  "reverse_engineering": {
    "status": "completed",
    "last_generated": "<ISO 8601>",
    "artifact_dir": "workflow-source/reverse-engineering/",
    "artifact_count": 9,
    "standard_ref": "AIDLC inception/reverse-engineering.md @ b19c819"
  }
}
```

### Step 12: User Message

```markdown
# 🔍 Reverse Engineering Complete

[Summary of key findings from 9 artifacts]

> **📋 REVIEW REQUIRED**
> Please examine the reverse engineering artifacts at: `workflow-source/reverse-engineering/`

> **🚀 WHAT'S NEXT?**
>
> 🔧 **Request Changes** — Ask for modifications to the analysis
> ✅ **Approve & Continue** — Approve and proceed to **Requirements Analysis** (v0.7.1+ 예정)
```

### Step 13: User Approval Wait (MANDATORY)

- `append_audit_log("user_response", <raw input>)` 호출
- 사용자 명시적 승인 전까지 후속 단계 진행 금지
- 승인 = `[Answer]: A) Approve & Continue` (Question Format 형식 따름, v0.6.4 Question File Format step A)

## 6. 우리 사용 패턴 적응

| AIDLC 패턴 | 우리 적응 | 비고 |
|---|---|---|
| HTTP API 중심 분석 | workflow API (MCP + workflow_kit) | HTTP API 없음 |
| Lambda / ECS / DynamoDB | harness / MCP / state.json | 인프라 = local runtime |
| CDK / Terraform | pyproject.toml + harness config | packaging 차원 |
| Business transaction = HTTP request | business transaction = workflow stage transition | 동일 추상화 |
| 9 artifact 각 ~50 line | 9 artifact 각 ~30-50 line (placeholder 위주) | template 형식, 실제 fill 은 tool/user |

## 7. 한계 / 예외

- **Template 상태**: 9 artifact 가 placeholder 위주 — 실제 fill 은 `existing_project_onboarding.py` 가 자동 생성 (v0.7.1+)
- **Greenfield**: artifact 0개. `state.json.reverse_engineering.status = "skipped"` 만 기록
- **AIDLC 와 차이**: 우리 business transaction = workflow stage. AIDLC 의 7 inception stage 와 1:1 매핑 아님 — 단순화.
- **N/A 처리**: HTTP API 없는 brownfield 도 04-api-documentation.md 작성 (MCP + workflow_kit 으로 채움)

## 8. Follow-up (v0.7.1+)

- `existing_project_onboarding.py` 가 9 artifact 자동 fill (현재 단일 `repository_assessment.md` 만 자동)
- v0.7.1: SEC-WF-05 dependency integrity 검증 — `07-dependencies.md` 의 lock file + checksum 자동 확인
- v0.7.1: 8-Artifact 별 wiki L1 page (각 artifact 마다 L1 topic page + L2 sources/)
- v0.8.0: Artifact 별 version diff (이전 reverse engineering 대비 변경점)

## 9. References

- 1차 출처: AIDLC `aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (b19c819, 2026-06-08)
- 우리 L1 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md` (D = 9-Artifact)
- 우리 기존: `workflow-source/scripts/run_existing_project_onboarding.py` (1차 분석)
- 우리 SSOT: `workflow-source/reverse-engineering/{01..09}-*.md` (9 artifact)
- 우리 검증: `workflow-source/tests/check_reverse_engineering.py` (smoke test)
