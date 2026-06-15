# Question File Format Pattern

- 문서 목적: 표준 AI 워크플로우에서 사용자 결정을 받을 때 AIDLC 의 `common/question-format-guide.md` 패턴을 차용해 multi-choice + `[Answer]:` tag + contradiction/ambiguity detection 을 강제한다.
- 범위: question file 형식, `[Answer]:` tag 규약, "Other" 옵션, contradiction/ambiguity 자동 점검, clarification file, gate 정책
- 대상 독자: 워크플로우 skill 구현자, AI agent, 프로젝트 온보딩 담당자, 운영자
- 상태: stable (v0.6.4 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: `./workflow_adoption_entrypoints.md` §7.1, `./stage_gate_pattern.md`, `./output_schema_guide.md` §3.2, AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/question-format-guide.md` (1차 출처)
- 1차 출처: AIDLC `common/question-format-guide.md` (369 lines, 2026-06-08 commit `b19c819`)

## 1. 왜 inline Q&A 에서 question file 로

기존 패턴:
- 채팅이나 LLM 호출 안에서 직접 사용자에게 질문
- 모호한 응답 ("mix of", "depends on complexity") 무방비로 수용
- 결정 근거가 세션 메모리에만 남고 SSOT 부재

AIDLC 의 question-format-guide 가 해결하는 문제:
- **결정의 SSOT 화**: 모든 질문이 별도 question file 에 기록 → git/PR 검토 가능
- **모호함 자동 점검**: 모호 응답 → clarification file 자동 생성 → 결정이 명시적일 때까지 stage gate 정지
- **Multi-choice 일관성**: A/B/C/D + X (Other) 형식 강제. 옵션 사이 빈 줄 (CommonMark strict renderer 호환)
- **Audit 가능**: 모든 결정이 git history 에 남음

## 2. Question File 형식

### 2.1 파일 위치와 명명

- 위치: `ai-workflow/memory/active/questions/<phase>-questions.md` 또는 프로젝트별 결정 (예: `docs/onboarding/questions/`)
- 명명: `{phase-name}-questions.md`. 예: `classification-questions.md`, `requirements-questions.md`, `design-questions.md`, `onboarding-questions.md`
- 새 세션의 첫 결정 → `initial-questions.md` (선택)

### 2.2 Header 형식

```markdown
# [Phase Name] Clarification Questions

본 문서는 `<phase>` 단계의 사용자 결정을 받기 위한 질문 묶음입니다.
각 질문의 마지막 옵션 (X) Other 는 필수이며, 답은 `[Answer]:` tag 옆에 letter (A/B/C/...) 로 표기합니다.

## Question 1
[Clear, specific question text]

A) [First meaningful option]

B) [Second meaningful option]

[...additional options as needed...]

X) Other (please describe after [Answer]: tag below)

[Answer]:
```

### 2.3 형식 강제 규칙 (CRITICAL)

| 규칙 | 적용 | 근거 |
|---|---|---|
| **다른 옵션 사이 빈 줄** | mandatory | CommonMark strict renderer (IntelliJ, PyCharm, GitLab) 가 옵션이 붙어 있으면 한 paragraph 로 collapse 함 |
| **"Other" 옵션은 마지막** | mandatory | A/B/C/D/X 패턴. "Other" 없으면 custom 응답 불가 |
| **2-5 meaningful options** | 권장 | 옵션 너무 많으면 결정 피로. 1-2 옵션은 부적합 (Yes/No ❌) |
| **`[Answer]:` tag 한 칸 띄움** | mandatory | parser 가 tag 인식. `[Answer]:A` 처럼 붙이면 ❌ |
| **빈 답변은 gate 위반** | mandatory | 모든 tag 가 채워질 때까지 stage gate 정지 |
| **문서 overwrite 금지** | mandatory | append-only. 검증 후 history 보존 |

### 2.4 사용 예시 (3 question sample)

```markdown
# Onboarding Clarification Questions

신규 프로젝트 온보딩을 위한 결정 사항입니다. 각 질문에 letter 로 답해주세요.

## Question 1
프로젝트 유형은?

A) Greenfield (빈 저장소에서 시작)

B) Brownfield (기존 코드베이스 분석 후 도입)

C) Hybrid (기존 skeleton + workflow 부터 동시 시작)

X) Other (please describe after [Answer]: tag below)

[Answer]: B

## Question 2
주요 언어/스택은?

A) Python (Django/Flask/FastAPI)

B) TypeScript/Node.js (Next.js/Express)

C) Go (Gin/Echo)

D) Rust (Actix/Axum)

X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 3
CI/CD 환경은?

A) GitHub Actions (현재 보유)

B) GitLab CI

C) Jenkins (self-hosted)

D) 배포 자동화 없음 (수동 배포)

X) Other (please describe after [Answer]: tag below)

[Answer]:
```

## 3. Answer 검증과 Contradiction Detection

### 3.1 Workflow

1. **질문 emit**: stage 시작 시 question file 작성 → 사용자에게 알림
2. **사용자 응답**: question file 을 직접 편집해 `[Answer]:` 채움
3. **parser 실행**: `workflow_kit/common/contracts/question_format.py` 의 `parse_answers(file_path)` 호출
4. **4 가지 검증**:
   - **Missing**: `[Answer]:` tag 가 비어있으면 ❌
   - **Invalid**: tag 옆 letter 가 옵션에 없으면 ❌
   - **Ambiguous**: letter + 추가 설명이 같이 있으면 (예: `[Answer]: A — but also maybe B`) parser 가 follow-up trigger
   - **Contradiction**: 다른 답변끼리 모순 (예: Q1 "bug fix" + Q2 "entire codebase affected")
5. **Clarification file emit**: 검증 실패 시 `*-clarification-questions.md` 자동 생성, 모순/모호함 설명 + targeted follow-up 질문
6. **Gate**: 모든 검증 통과 시에만 stage 진행

### 3.2 모호 응답 자동 감지 패턴

다음 keyword/문구가 answer 또는 answer 의 context 에 있으면 ambiguous 로 마킹:

- "mix of", "between A and B", "somewhere between"
- "depends on", "it depends", "상황에 따라"
- "not sure", "잘 모르겠", "모르겠"
- "TBD", "to be decided", "미정"

각 case 별 follow-up 질문 예시:

| 모호 응답 패턴 | Follow-up 질문 |
|---|---|
| "mix of A and B" | "A 와 B 의 결정 기준은 무엇인가요? 어떤 조건에서 A 를 쓰고 어떤 조건에서 B 를 쓰나요?" |
| "depends on complexity" | "complexity level 을 어떻게 정의하시나요? (예: LOC > 1000, components > 5, etc.)" |
| "not sure" | "결정에 필요한 추가 정보는 무엇인가요?" |
| "between A and B" | "정확한 중간 지점은 어디인가요? A 의 몇 % + B 의 몇 % 같은 식으로요." |

### 3.3 Contradiction Detection 예시

| Question A | Question B | Contradiction |
|---|---|---|
| Q1: "Bug fix scope" → "single component" | Q2: "Impact area" → "entire codebase" | ❌ single component + entire codebase 모순 |
| Q1: "Risk level" → "low" | Q2: "Breaking changes" → "yes" | ❌ low risk + breaking changes 모순 |
| Q1: "Timeline" → "quick fix" | Q2: "Subsystem affected" → "more than 3" | ❌ quick fix + 다수 subsystem 모순 |

각 contradiction 별로 `*-clarification-questions.md` 에:
- "Q1 과 Q2 의 답이 모순됩니다. 이유: ..."
- "어떤 답변이 정확한가요? (A) Q1 유지, (B) Q2 유지, (C) 둘 다 수정"
- follow-up `[Answer]:` tag

## 4. 구현 위치

### 4.1 문서

- 본 spec: `workflow-source/core/question_file_format.md` (이 문서)
- 적용: `project_workflow_profile_template.md`, `session_handoff_template.md`, 각 skill spec 의 "입력" 섹션
- cross-ref: `workflow_adoption_entrypoints.md` §7.1

### 4.2 코드 (v0.6.4 신규)

- `workflow_kit/common/contracts/question_format.py`
  - `parse_answers(file_path: Path) -> dict[int, str]`
  - `validate_answers(answers: dict, options: list[list[str]]) -> list[ValidationError]`
  - `detect_ambiguity(answers: dict) -> list[Ambiguity]`
  - `detect_contradiction(answers: dict, metadata: dict) -> list[Contradiction]`
  - `generate_clarification_file(errors: list, output_path: Path)`

### 4.3 테스트

- `workflow-source/tests/check_question_format.py`
  - 정상 응답 (모든 tag 채움 + 유효한 letter) → PASS
  - 빈 응답 (1+ tag 비어있음) → FAIL (gate 정지)
  - 모호 응답 (mix of, depends on 등) → FAIL (clarification file emit)
  - Contradiction (cross-question 모순) → FAIL (clarification file emit)

### 4.4 프로젝트별 적용 예시

- `standard_ai_workflow_minimax` 자체: Q&A 가 거의 없음 (template 기반) — 적용 안 함
- `devhub_example`: PR review / sprint planning 시 Q&A 가 inline 으로 발생 → question file 적용 후보
- `my_harness` (Rust): TDD chain 별 design decision 시 적용 후보

## 5. 한계와 예외

- **긴급 Q&A (1회성, 기록 가치 없음)**: question file 안 만들고 inline Q&A 도 OK. 예: "이거 urgent 한데?" 같은 즉시성.
- **대화형 CLI**: 일부 harness 의 interactive picker 는 multi-choice 가 아닌 자유 입력. 이 경우 question file 의 gate 정책 적용 안 함.
- **자동화 흐름 (CI/CD)**: 사용자 응답이 없으면 gate 자동 fail. timeout 정책 별도 정의.

## 6. 다음에 읽을 문서

- [`./stage_gate_pattern.md`](./stage_gate_pattern.md) — Question Format 의 gate 와 결합되는 Stage Gate 명시화 패턴
- [`./workflow_adoption_entrypoints.md` §7.1](./workflow_adoption_entrypoints.md#71-question-file-format-패턴--권장)
- [`./output_schema_guide.md` §3.2](./output_schema_guide.md) — stage_completion field 와 정합
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/question-format-guide.md`
