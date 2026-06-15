#!/usr/bin/env python3
"""v0.7.0 step 8: Security Baseline (extension) 검증.

- extensions/security-baseline.md 6 rule 정의 검증
- extensions/security-baseline.opt-in.md 형식 검증
- rule ID (SEC-WF-01 ~ 06) 정합성
- Compliance summary 형식 (status enum, N/A 처리)
- v0.7.0 follow-up: workflow_kit.common.contracts.security_baseline helper (v0.7.1+)
  본 test 는 spec level 만 검증. runtime evaluate_compliance 는 v0.7.1+ follow-up.
Reference: workflow-source/extensions/security-baseline.md
           workflow-source/extensions/security-baseline.opt-in.md
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

BASELINE_PATH = SOURCE_ROOT / "extensions" / "security-baseline.md"
OPT_IN_PATH = SOURCE_ROOT / "extensions" / "security-baseline.opt-in.md"

# 정규식 패턴
RULE_HEADER_RE = re.compile(r"^### \d+\.\d+ Rule (SEC-WF-\d{2}):\s+(.+)$", re.MULTILINE)
# Verification 은 bold text 형식 (**Verification**) — AIDLC 패턴 준수
VERIFICATION_HEADER_RE = re.compile(r"^\*\*Verification\*\*\s*$", re.MULTILINE)
# Verification: bullet list 의 일부. 첫 번째 bullet 만 capture.
VERIFICATION_BULLET_RE = re.compile(r"^- (.+)$", re.MULTILINE)
COMPLIANCE_ROW_RE = re.compile(r"^\| (SEC-WF-\d{2}) \((.+?)\) \| (.+?) \| (.+?) \|$", re.MULTILINE)


# --- Test 1: spec 존재 ---


def test_baseline_exists() -> None:
    """extensions/security-baseline.md 존재."""
    assert BASELINE_PATH.exists(), f"baseline not found: {BASELINE_PATH}"


def test_opt_in_exists() -> None:
    """extensions/security-baseline.opt-in.md 존재."""
    assert OPT_IN_PATH.exists(), f"opt-in not found: {OPT_IN_PATH}"


# --- Test 2: 6 rule 정의 ---


def test_six_rules_defined() -> None:
    """SEC-WF-01 ~ SEC-WF-06 6 rule 모두 정의."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    rule_headers = RULE_HEADER_RE.findall(content)
    rule_ids = {rid for rid, _ in rule_headers}
    expected = {f"SEC-WF-{i:02d}" for i in range(1, 7)}
    assert rule_ids == expected, f"expected {expected}, got {rule_ids}"


def test_rule_titles_meaningful() -> None:
    """각 rule 의 title 이 non-empty."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    for rid, title in RULE_HEADER_RE.findall(content):
        assert len(title.strip()) > 5, f"{rid} title too short: {title!r}"


def test_each_rule_has_verification_section() -> None:
    """각 rule 이 Verification subsection 가짐 (bold text **Verification** 형식)."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    for rid, _ in RULE_HEADER_RE.findall(content):
        # rule_id 다음의 **Verification** 마커 존재
        rule_pos_match = re.search(
            rf"### \d+\.\d+ Rule {rid}:.*?(?=\n### \d|\n## |\Z)",
            content,
            re.DOTALL,
        )
        assert rule_pos_match is not None, f"{rid} rule block not found"
        rule_block = rule_pos_match.group(0)
        assert "**Verification**" in rule_block, \
            f"{rid} missing **Verification** section"


def test_verification_has_bullets() -> None:
    """Verification subsection 이 bullet point ≥ 3개."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    for rid, _ in RULE_HEADER_RE.findall(content):
        # rule block 의 **Verification** 이후 bullet 추출
        # 다음 rule (### N.M Rule) 또는 다음 section (## N) 까지
        rule_pos_match = re.search(
            rf"### \d+\.\d+ Rule {rid}:.*?\*\*Verification\*\*:?[ \t]*\n(.*?)(?=\n### \d+\.\d+|\n## |\Z)",
            content,
            re.DOTALL,
        )
        assert rule_pos_match is not None, f"{rid} Verification not found"
        section = rule_pos_match.group(1)
        bullets = VERIFICATION_BULLET_RE.findall(section)
        assert len(bullets) >= 3, f"{rid} Verification has {len(bullets)} bullets (need ≥ 3)"


def test_each_rule_has_verification_section_legacy() -> None:
    """각 rule 이 Verification subsection 가짐."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    for rid, _ in RULE_HEADER_RE.findall(content):
        # 같은 rule id 가 Verification subsection 에도 등장
        v_matches = VERIFICATION_HEADER_RE.findall(content)
        assert rid in v_matches, f"{rid} missing Verification section"


def test_verification_has_bullets_legacy() -> None:
    """Verification subsection 이 bullet point ≥ 3개."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    for rid, _ in RULE_HEADER_RE.findall(content):
        # 해당 rule 의 Verification 섹션 추출
        m = re.search(
            rf"#### Rule {rid} Verification\s*\n(.*?)(?=\n###|\n##|\Z)",
            content,
            re.DOTALL,
        )
        assert m is not None, f"{rid} Verification not found"
        section = m.group(1)
        bullets = VERIFICATION_BULLET_RE.findall(section)
        assert len(bullets) >= 3, f"{rid} Verification has {len(bullets)} bullets (need ≥ 3)"


# --- Test 3: opt-in format ---


def test_opt_in_question_format() -> None:
    """opt-in prompt 가 multi-choice + [Answer]: tag 형식."""
    content = OPT_IN_PATH.read_text(encoding="utf-8")
    # A) Yes / B) No / X) Other 패턴
    assert "A) Yes" in content
    assert "B) No" in content
    assert "X) Other" in content
    # [Answer]: tag
    assert "[Answer]:" in content


def test_opt_in_response_format_documented() -> None:
    """opt-in prompt 가 응답 형식 (security_baseline_status.md) 명시."""
    content = OPT_IN_PATH.read_text(encoding="utf-8")
    assert "security_baseline_status.md" in content
    assert "Extension Configuration" in content


# --- Test 4: compliance summary 형식 ---


def test_compliance_summary_status_enum() -> None:
    """compliance summary 의 Status 가 valid enum (✅ / ❌ / ⚠️ / ⏸)."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    # §4 compliance summary 표 추출
    summary_match = re.search(
        r"## 4\. Compliance Summary.*?(?=\n## 5\.)",
        content,
        re.DOTALL,
    )
    assert summary_match is not None, "§4 Compliance Summary not found"
    summary = summary_match.group(0)
    valid_statuses = {"✅", "❌", "⚠️", "⏸"}
    # row 들에서 status emoji 추출
    for m in COMPLIANCE_ROW_RE.finditer(summary):
        rule_id, name, status, notes = m.groups()
        # status 의 첫 emoji 가 valid
        for s in valid_statuses:
            if s in status:
                break
        else:
            # status 에 valid emoji 없으면 fail
            assert False, f"{rule_id} status invalid: {status!r}"


def test_compliance_summary_hard_constraint_documented() -> None:
    """§4 의 hard constraint 정책 (1+ ❌ 시 gate 정지) 명시."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    assert "hard constraint" in content.lower() or "blocking" in content.lower()
    assert "Request Changes" in content  # blocking 시 option
    assert "Continue" in content  # normal 시 option


# --- Test 5: workflow_kit integration ---


def test_workflow_kit_security_helper_referenced() -> None:
    """baseline doc 가 workflow_kit.common.contracts.security_baseline helper 참조 (v0.7.1+)."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    assert "workflow_kit.common.contracts.security_baseline" in content


def test_existing_test_files_cross_ref() -> None:
    """baseline doc 가 기존 smoke test 4종 cross-ref (자동 검증 위임)."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    expected_refs = [
        "check_audit_log_compliance.py",
        "check_stage_gate_compliance.py",
        "check_question_format.py",
        "check_stage_completion_required.py",
    ]
    for ref in expected_refs:
        assert ref in content, f"baseline missing cross-ref to {ref}"


# --- Test 6: AIDLC 호환 ---


def test_aidlc_extensions_pattern_followed() -> None:
    """AIDLC extensions 시스템 패턴 준수: <name>.md + <name>.opt-in.md 페어."""
    assert BASELINE_PATH.exists()
    assert OPT_IN_PATH.exists()
    # 두 file 이 같은 디렉토리에 위치
    assert BASELINE_PATH.parent == OPT_IN_PATH.parent
    # 파일명 패턴: <name>.md + <name>.opt-in.md
    assert BASELINE_PATH.stem == "security-baseline"
    assert OPT_IN_PATH.stem == "security-baseline.opt-in"


def test_aidlc_security_baseline_md_cross_ref() -> None:
    """AIDLC 원본 security-baseline.md (307 line) cross-ref."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    assert "awslabs/aidlc-workflows" in content
    assert "security-baseline.md" in content
    assert "307 line" in content


# --- Test 7: rule 6 종 정합성 ---


def test_rule_topic_coverage() -> None:
    """6 rule 이 서로 다른 topic 커버 (중복 ❌)."""
    content = BASELINE_PATH.read_text(encoding="utf-8")
    rules = {}
    for m in RULE_HEADER_RE.finditer(content):
        rules[m.group(1)] = m.group(2)

    expected_topics = {
        "SEC-WF-01": ["audit", "log"],
        "SEC-WF-02": ["stage", "gate", "approval"],
        "SEC-WF-03": ["question", "format"],
        "SEC-WF-04": ["error", "fail-closed"],
        "SEC-WF-05": ["depend", "lock"],
        "SEC-WF-06": ["r-9", "wiki"],
    }
    for rid, keywords in expected_topics.items():
        title = rules[rid].lower()
        # 적어도 1 keyword 매칭
        assert any(kw in title for kw in keywords), \
            f"{rid} title {title!r} doesn't match expected topic keywords {keywords}"


# --- Main ---


def main() -> int:
    tests = [
        ("baseline_exists", test_baseline_exists),
        ("opt_in_exists", test_opt_in_exists),
        ("six_rules_defined", test_six_rules_defined),
        ("rule_titles_meaningful", test_rule_titles_meaningful),
        ("each_rule_has_verification_section", test_each_rule_has_verification_section),
        ("verification_has_bullets", test_verification_has_bullets),
        ("opt_in_question_format", test_opt_in_question_format),
        ("opt_in_response_format_documented", test_opt_in_response_format_documented),
        ("compliance_summary_status_enum", test_compliance_summary_status_enum),
        ("compliance_summary_hard_constraint_documented", test_compliance_summary_hard_constraint_documented),
        ("workflow_kit_security_helper_referenced", test_workflow_kit_security_helper_referenced),
        ("existing_test_files_cross_ref", test_existing_test_files_cross_ref),
        ("aidlc_extensions_pattern_followed", test_aidlc_extensions_pattern_followed),
        ("aidlc_security_baseline_md_cross_ref", test_aidlc_security_baseline_md_cross_ref),
        ("rule_topic_coverage", test_rule_topic_coverage),
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
    raise SystemExit(main())
