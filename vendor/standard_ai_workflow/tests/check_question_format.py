#!/usr/bin/env python3
"""Question File Format (v0.6.4) regression check.

Verifies that workflow_kit.common.contracts.question_format correctly:
- parses [Answer]: tags from a well-formed question file
- detects missing answers (gate violation)
- detects invalid letter
- detects ambiguity patterns (mix of, depends on, not sure, etc.)
- detects "Other" option absence (mandatory per spec)
- generates clarification file with proper follow-up
Reference: workflow-source/core/question_file_format.md
"""

from __future__ import annotations

import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.common.contracts.question_format import (  # noqa: E402
    AnswerEntry,
    ValidationResult,
    detect_ambiguity,
    detect_contradiction,
    full_validation,
    generate_clarification_file,
    parse_answers,
    validate_answers,
)


# --- Baseline question file content ---

WELL_FORMED_QUESTIONS = """# Onboarding Clarification Questions

신규 프로젝트 온보딩을 위한 결정 사항입니다.

## Question 1
프로젝트 유형은?

A) Greenfield

B) Brownfield

X) Other (please describe after [Answer]: tag below)

[Answer]: B

## Question 2
주요 언어는?

A) Python

B) TypeScript

C) Go

X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 3
CI/CD 환경은?

A) GitHub Actions

B) GitLab CI

X) Other (please describe after [Answer]: tag below)

[Answer]:
"""


MISSING_ANSWER_QUESTIONS = """# Sample Questions

## Question 1
테스트?

A) Yes

B) No

X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 2
다른 질문?

A) Option A

B) Option B

X) Other (please describe after [Answer]: tag below)

[Answer]:
"""


AMBIGUOUS_QUESTIONS = """# Sample Questions

## Question 1
어떻게 진행?

A) A안

B) B안

X) Other (please describe after [Answer]: tag below)

[Answer]: A — mix of A and B

## Question 2
언제까지?

A) 오늘

B) 다음주

X) Other (please describe after [Answer]: tag below)

[Answer]: it depends on complexity
"""


CONTRADICTION_QUESTIONS = """# Sample Questions

## Question 1
수정 범위?

A) 단일 component

B) 전체 codebase

X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 2
영향 영역?

A) 단일 component

B) 전체 codebase

X) Other (please describe after [Answer]: tag below)

[Answer]: B
"""


NO_OTHER_QUESTIONS = """# Sample Questions

## Question 1
선택?

A) Yes

B) No

[Answer]: A
"""


def _write_tmp(content: str) -> Path:
    """임시 파일 작성 후 path 반환."""
    tmp = tempfile.NamedTemporaryFile(
        mode="w", suffix=".md", delete=False, encoding="utf-8"
    )
    tmp.write(content)
    tmp.close()
    return Path(tmp.name)


# --- Test cases ---


def test_well_formed_questions() -> None:
    """well-formed question file → 모든 tag 유효, gate 통과."""
    fp = _write_tmp(WELL_FORMED_QUESTIONS)
    try:
        options = {1: ["A", "B", "X"], 2: ["A", "B", "C", "X"], 3: ["A", "B", "X"]}
        result = full_validation(fp, options)
        # Q3 는 empty answer → missing 에러 1건
        assert len(result.errors) == 1, f"expected 1 missing error, got {len(result.errors)}"
        assert result.errors[0].question_num == 3
        assert result.errors[0].error_type == "missing"
        # Q1, Q2 는 유효
        assert 1 in result.answers
        assert result.answers[1].raw_letter == "B"
        assert 2 in result.answers
        assert result.answers[2].raw_letter == "A"
        # Q3 는 missing 이므로 is_valid False
        assert result.is_valid is False
    finally:
        fp.unlink(missing_ok=True)


def test_missing_answer() -> None:
    """empty [Answer]: tag → missing 에러."""
    fp = _write_tmp(MISSING_ANSWER_QUESTIONS)
    try:
        options = {1: ["A", "B", "X"], 2: ["A", "B", "X"]}
        result = full_validation(fp, options)
        assert len(result.errors) == 1
        assert result.errors[0].error_type == "missing"
        assert result.errors[0].question_num == 2
        assert result.is_valid is False
    finally:
        fp.unlink(missing_ok=True)


def test_ambiguity_detection() -> None:
    """'mix of', 'depends on' 등 모호 패턴 감지."""
    fp = _write_tmp(AMBIGUOUS_QUESTIONS)
    try:
        options = {1: ["A", "B", "X"], 2: ["A", "B", "X"]}
        result = full_validation(fp, options)
        assert len(result.ambiguities) == 2, f"expected 2 ambiguities, got {len(result.ambiguities)}"
        keywords = {a.matched_keyword for a in result.ambiguities}
        assert "mix of" in keywords
        assert "it depends" in keywords or "depends on" in keywords
        assert result.is_valid is False
    finally:
        fp.unlink(missing_ok=True)


def test_contradiction_detection() -> None:
    """cross-question 모순 자동 감지."""
    fp = _write_tmp(CONTRADICTION_QUESTIONS)
    try:
        options = {1: ["A", "B", "X"], 2: ["A", "B", "X"]}
        rule = {
            "qa": 1,
            "qb": 2,
            "qa_letter": "A",
            "qb_letter": "B",
            "description": "단일 component + 전체 codebase 모순",
        }
        result = full_validation(fp, options, contradiction_rules=[rule])
        assert len(result.contradictions) == 1
        assert result.contradictions[0].question_a == 1
        assert result.contradictions[0].question_b == 2
        assert "단일 component" in result.contradictions[0].description
        assert result.is_valid is False
    finally:
        fp.unlink(missing_ok=True)


def test_no_other_option() -> None:
    """'Other' 옵션 부재 → mandatory violation."""
    fp = _write_tmp(NO_OTHER_QUESTIONS)
    try:
        options = {1: ["A", "B"]}  # X 빠짐
        result = full_validation(fp, options)
        # 'Other' mandatory 위반
        other_errors = [e for e in result.errors if "Other" in e.message]
        assert len(other_errors) >= 1, f"expected Other mandatory error, got {result.errors}"
    finally:
        fp.unlink(missing_ok=True)


def test_generate_clarification_file() -> None:
    """clarification file 자동 emit — 발견된 이슈가 follow-up 으로 변환."""
    fp = _write_tmp(AMBIGUOUS_QUESTIONS)
    try:
        options = {1: ["A", "B", "X"], 2: ["A", "B", "X"]}
        result = full_validation(fp, options)

        out_path = Path(tempfile.gettempdir()) / "test-clarification-questions.md"
        try:
            generate_clarification_file(
                errors=result.errors,
                ambiguities=result.ambiguities,
                contradictions=result.contradictions,
                output_path=out_path,
            )
            content = out_path.read_text(encoding="utf-8")
            assert "Clarification Questions" in content
            assert "발견된 모호 응답" in content
            assert "mix of" in content or "depends on" in content
        finally:
            out_path.unlink(missing_ok=True)
    finally:
        fp.unlink(missing_ok=True)


def test_parse_answers_file_not_found() -> None:
    """존재하지 않는 file → FileNotFoundError."""
    try:
        parse_answers("/nonexistent/path/questions.md")
        raise AssertionError("expected FileNotFoundError")
    except FileNotFoundError:
        pass


# --- Main ---

def main() -> int:
    tests = [
        ("well_formed_questions", test_well_formed_questions),
        ("missing_answer", test_missing_answer),
        ("ambiguity_detection", test_ambiguity_detection),
        ("contradiction_detection", test_contradiction_detection),
        ("no_other_option", test_no_other_option),
        ("generate_clarification_file", test_generate_clarification_file),
        ("parse_answers_file_not_found", test_parse_answers_file_not_found),
    ]
    failed: list[str] = []
    for name, fn in tests:
        try:
            fn()
            print(f"  PASS  {name}")
        except AssertionError as e:
            failed.append(f"  FAIL  {name}: {e}")
            print(f"  FAIL  {name}: {e}")
        except Exception as e:  # noqa: BLE001
            failed.append(f"  ERROR {name}: {type(e).__name__}: {e}")
            print(f"  ERROR {name}: {type(e).__name__}: {e}")

    print()
    if failed:
        print(f"{len(failed)}/{len(tests)} tests failed:")
        for f in failed:
            print(f)
        return 1
    print(f"All {len(tests)} tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
