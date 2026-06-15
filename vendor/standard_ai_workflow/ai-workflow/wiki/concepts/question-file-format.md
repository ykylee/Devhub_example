---
type: concept
status: active
last_ingested_from: workflow-source/core/question_file_format.md + workflow_kit/common/contracts/question_format.py
related_pages: [concepts/stage-gate-pattern, decisions/adr-001-3-layer-separation, patterns/r4-anchor-index, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-12
updated: 2026-06-12
---

# Question File Format (v0.6.4, AIDLC 차용)

- 문서 목적: standard_ai_workflow v0.6.4 의 Question File Format 패턴 (AIDLC `common/question-format-guide.md` 차용) 의 외부 markdown spec + Python enforcement helper 결합 정책을 정리한다. multi-choice + `[Answer]:` tag + "Other" mandatory + contradiction/ambiguity auto-detection.
- 범위: 외부 spec 구조, `parse_answers` / `validate_answers` / `detect_ambiguity` / `detect_contradiction` / `generate_clarification_file` API, stage gate 와의 결합
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/core/question_file_format.md` (204 lines, v0.6.4 stable) |
| 2 | Python helper | `workflow_kit/common/contracts/question_format.py` (358 lines) |
| 3 | smoke test | `workflow-source/tests/check_question_format.py` (336 lines, 7 test PASS) |
| 4 | Source 1차 출처 | AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/question-format-guide.md` (369 lines, 2026-06-08 commit `b19c819`) |
| 5 | 도입 버전 | v0.6.4 (commit `25756bb` + `bc16d91`) |
| 6 | 관련 ADR | 없음 (신규) |
| 7 | 관련 패턴 | [[concepts/stage-gate-pattern]] (gate 와 결합) |
| 8 | 관련 토픽 | [[topics/aidlc-benchmark-analysis-2026-06-12]] §4 보완안 (A) |

## §2 왜 inline Q&A 에서 question file 로  {#s2-why}

기존 inline Q&A 의 약점:

- 채팅 안에서 직접 사용자에게 질문 → 결정 근거가 세션 메모리에만 남음
- 모호한 응답 ("mix of", "depends on") 무방비로 수용
- 결정의 SSOT 부재 → audit 어려움

AIDLC 의 `common/question-format-guide.md` 가 해결하는 문제:

- **결정의 SSOT 화**: 모든 질문이 별도 question file (예: `<phase>-questions.md`) 에 기록 → git/PR 검토 가능
- **모호함 자동 점검**: 모호 응답 → clarification file 자동 생성 → stage gate 정지
- **Multi-choice 일관성**: A/B/C/D + X (Other) 형식 강제
- **Audit 가능**: 모든 결정이 git history 에 남음

## §3 Question File 형식  {#s3-format}

### §3.1 파일 위치

`ai-workflow/memory/active/questions/<phase>-questions.md` 또는 프로젝트별 결정

### §3.2 형식 강제 규칙

| 규칙 | 적용 |
|---|---|
| 옵션 사이 빈 줄 | mandatory (CommonMark strict renderer 호환) |
| "Other" 옵션 마지막 | mandatory (A/B/C/D/X 패턴) |
| 2-5 meaningful options | 권장 |
| `[Answer]:` tag 한 칸 띄움 | mandatory (parser 인식) |
| 빈 답변은 gate 위반 | mandatory |
| 문서 overwrite 금지 | mandatory (append-only) |

### §3.3 표준 형식 예시

```markdown
# Onboarding Clarification Questions

## Question 1
프로젝트 유형은?

A) Greenfield

B) Brownfield

X) Other (please describe after [Answer]: tag below)

[Answer]: B
```

## §4 Python Enforcement  {#s4-python}

### §4.1 API 표면

| 함수 | 시그니처 | 용도 |
|---|---|---|
| `parse_answers` | `(file_path) -> dict[int, AnswerEntry]` | `[Answer]:` tag parse |
| `validate_answers` | `(answers, options) -> list[ValidationError]` | missing / invalid letter / "Other" mandatory 검증 |
| `detect_ambiguity` | `(answers) -> list[Ambiguity]` | "mix of" / "depends on" / "not sure" 등 자동 감지 |
| `detect_contradiction` | `(answers, rules) -> list[Contradiction]` | cross-question 모순 |
| `generate_clarification_file` | `(errors, ambiguities, contradictions, output_path)` | follow-up question file 자동 emit |
| `full_validation` | `(file_path, options, contradiction_rules) -> ValidationResult` | all-in-one convenience |

### §4.2 모호 응답 자동 감지 패턴

8 keyword: `mix of`, `between a and b`, `somewhere between`, `depends on`, `it depends`, `상황에 따라`, `not sure`, `잘 모르겠`, `모르겠`, `tbd`, `to be decided`, `미정`

6 follow-up 매핑:

| 모호 keyword | Follow-up |
|---|---|
| `mix of` | A 와 B 의 결정 기준은? 어떤 조건에서 A vs B? |
| `depends on` | complexity level 정의? (LOC, components, dependencies) |
| `not sure` | 결정에 필요한 추가 정보는? |
| `tbd` | 결정 가능한 시점? 그때까지 stage gate 정지? |
| `between a and b` | 정확한 중간 지점? (A 몇 % + B 몇 %) |
| 그 외 | 더 구체적으로 답변 요청 |

### §4.3 Contradiction Detection 규칙 예시

| Q1 | Q2 | Contradiction |
|---|---|---|
| "Bug fix scope" → "single component" | "Impact area" → "entire codebase" | ❌ 모순 |
| "Risk level" → "low" | "Breaking changes" → "yes" | ❌ 모순 |
| "Timeline" → "quick fix" | "Subsystem affected" → "more than 3" | ❌ 모순 |

## §5 Stage Gate 와의 결합  {#s5-gate-binding}

| 단계 | 동작 | Gate |
|---|---|---|
| Stage 시작 | Question File Format 으로 사용자 결정 입력 | 모든 `[Answer]:` 채워질 때까지 stage 시작 ❌ |
| Stage 실행 | (skill/MCP) | (중간) |
| Stage 종료 | Stage Gate 2-option 으로 사용자 결정 입력 | approval_timestamp/actor 모두 있어야 다음 stage 진행 |
| (선택) 다음 stage | Question File Format 다시 시작 | (반복) |

## §6 적용 위치  {#s6-apply}

- `project_workflow_profile_template.md` — 신규 프로젝트 profile 작성 시 clarification file 자동 emit
- `ask_user` 호출 빈도 ↓ (한 popup 에 multi-question 압축)
- `session_handoff_template.md` — 세션 시작 시 결정 사항 SSOT 화
- 11종 skill spec 의 "입력" 섹션에 question file 패턴 cross-ref

## §7 한계와 예외  {#s7-limitations}

- **긴급 Q&A (1회성, 기록 가치 없음)**: question file 안 만들고 inline Q&A 도 OK
- **대화형 CLI**: 일부 harness 의 interactive picker 는 자유 입력 — gate 정책 적용 안 함
- **자동화 흐름 (CI/CD)**: 사용자 응답이 없으면 gate 자동 fail. timeout 정책 별도 정의

## §8 Related Decisions / Patterns  {#s8-related}

- [[concepts/stage-gate-pattern]] — Stage Gate 명시화 (output 단계)
- [[decisions/adr-001-3-layer-separation]] — Source/Runtime/Project Docs 3-layer
- [[patterns/r4-anchor-index]] — wiki anchor 기반 index
- [[topics/aidlc-benchmark-analysis-2026-06-12]] — AIDLC 벤치마크 분석 (A.Question File Format 도입 근거)

## §9 References  {#s9-references}

- 외부 spec: [`../../workflow-source/core/question_file_format.md`](../../workflow-source/core/question_file_format.md)
- Python helper: [`../../workflow-source/workflow_kit/common/contracts/question_format.py`](../../workflow-source/workflow_kit/common/contracts/question_format.py)
- smoke test: [`../../workflow-source/tests/check_question_format.py`](../../workflow-source/tests/check_question_format.py)
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/question-format-guide.md`
