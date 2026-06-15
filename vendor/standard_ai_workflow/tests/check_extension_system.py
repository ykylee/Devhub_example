#!/usr/bin/env python3
"""v0.7.0 step 7: Extension System 검증.

- extensions/SCHEMA.md (extension system SSOT) 존재
- 3종 extension (security / testing / performance) baseline + opt-in file 페어
- 각 baseline 의 rule ID 가 PREFIX-WF-NN 형식
- 각 baseline 의 **Rule** + **Verification** (≥ 3 bullet) 형식
- 각 baseline 의 Compliance Summary table 에 모든 rule ID 명시
- 각 opt-in 이 `> Question:` + `[Answer]:` 형식
- 각 opt-in 의 A/B/P/X 4 옵션 모두 존재
- Rule ID prefix 별 topic coverage (중복 topic 없음)
- AIDLC cross-reference (1차 출처 file path 명시)

Reference: workflow-source/extensions/SCHEMA.md
           workflow-source/extensions/{security,testing,performance}-baseline.md
           workflow-source/extensions/{security,testing,performance}-baseline.opt-in.md
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

EXTENSIONS_DIR = SOURCE_ROOT / "extensions"
SCHEMA_PATH = EXTENSIONS_DIR / "SCHEMA.md"

# 3종 extension 정의
EXTENSIONS = [
    {
        "name": "security",
        "prefix": "SEC-WF",
        "baseline": "security-baseline.md",
        "opt_in": "security-baseline.opt-in.md",
        "rule_count": 6,
        "aidlc_path": "extensions/security/baseline/security-baseline.md",
        "sub_cat": None,
    },
    {
        "name": "testing",
        "prefix": "TST-WF",
        "baseline": "testing-baseline.md",
        "opt_in": "testing-baseline.opt-in.md",
        "rule_count": 6,
        "aidlc_path": "extensions/testing/property-based/property-based-testing.md",
        "sub_cat": None,
    },
    {
        "name": "performance",
        "prefix": "PERF-WF",
        "baseline": "performance-baseline.md",
        "opt_in": "performance-baseline.opt-in.md",
        "rule_count": 6,
        "aidlc_path": None,  # 우리 domain 적응 — AIDLC 없음
        "sub_cat": None,
    },
]

# v0.7.2+ sub-cat: 각 parent baseline 의 sub-category
SUB_CAT_EXTENSIONS = [
    {
        "name": "security-auth",
        "parent": "security",
        "prefix": "SEC-AUTH",
        "baseline_path": "security/auth/auth-baseline.md",
        "opt_in_path": "security/auth/auth-baseline.opt-in.md",
        "rule_count": 6,
        "aidlc_path": "extensions/security/baseline/security-baseline.md",  # AIDLC parent
    },
    {
        "name": "testing-pbt",
        "parent": "testing",
        "prefix": "PBT-WF",
        "baseline_path": "testing/property-based/property-based-testing.md",
        "opt_in_path": "testing/property-based/property-based-testing.opt-in.md",
        "rule_count": 6,
        "aidlc_path": "extensions/testing/property-based/property-based-testing.md",
    },
    {
        "name": "performance-memory",
        "parent": "performance",
        "prefix": "PERF-MEM",
        "baseline_path": "performance/memory/memory-baseline.md",
        "opt_in_path": "performance/memory/memory-baseline.opt-in.md",
        "rule_count": 6,
        "aidlc_path": None,  # 우리 domain 적응
    },
    {
        "name": "resiliency",
        "parent": None,  # 4종 (standalone)
        "prefix": "RES-WF",
        "baseline_path": "resiliency-baseline.md",
        "opt_in_path": "resiliency-baseline.opt-in.md",
        "rule_count": 8,
        "aidlc_path": "extensions/resiliency/baseline/resiliency-baseline.md",
    },
]

# 정규식 패턴
# v0.7.2+ prefix format: <CAT>(-<SUB>)?(-WF)?-<NN>
# 예: SEC-WF-NN (v0.7.0), SEC-AUTH-NN (v0.7.2 sub-cat), TST-WF-NN, PBT-WF-NN, PERF-WF-NN, PERF-MEM-NN, RES-WF-NN
RULE_ID_RE = re.compile(r"^[A-Z]+(?:-[A-Z]+)?(?:-WF)?-\d{2}$")
RULE_HEADER_RE = re.compile(r"^### \d+\.\d+ Rule ([A-Z]+(?:-[A-Z]+)?(?:-WF)?-\d{2}):\s+(.+)$", re.MULTILINE)
VERIFICATION_BULLET_RE = re.compile(r"^- (.+)$", re.MULTILINE)
COMPLIANCE_ROW_RE = re.compile(r"^\| ([A-Z]+(?:-[A-Z]+)?(?:-WF)?-\d{2}) \|", re.MULTILINE)


# --- Test 1: SCHEMA.md 존재 ---


def test_schema_exists() -> None:
    """extensions/SCHEMA.md 존재."""
    assert SCHEMA_PATH.exists(), f"SCHEMA.md not found: {SCHEMA_PATH}"


def test_schema_has_directory_layout() -> None:
    """SCHEMA.md 가 Directory Layout section 명시."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    assert "## 2. Directory Layout" in content, "SCHEMA missing §2"
    assert "extensions/" in content, "SCHEMA missing extensions/ path"


def test_schema_has_file_format() -> None:
    """SCHEMA.md 가 File Format section (3.1 baseline + 3.2 opt-in) 명시."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    assert "### 3.1" in content and "<name>-baseline.md" in content, \
        "SCHEMA missing §3.1 baseline format"
    assert "### 3.2" in content and "<name>-baseline.opt-in.md" in content, \
        "SCHEMA missing §3.2 opt-in format"


def test_schema_has_prefix_convention() -> None:
    """SCHEMA.md 가 Rule ID prefix convention (SEC/TST/PERF) 명시."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    for ext in EXTENSIONS:
        assert ext["prefix"] in content, \
            f"SCHEMA missing prefix {ext['prefix']}"


def test_schema_has_lint_rule_section() -> None:
    """SCHEMA.md 가 §7 Lint Rule section 명시 (10 rule)."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    assert "## 7. Lint Rule" in content, "SCHEMA missing §7"
    # 10 lint rule 1~10 모두 명시
    for n in range(1, 11):
        assert re.search(rf"^\d+\. ", content, re.MULTILINE), "SCHEMA missing lint rules"


# --- Test 2: 3종 extension baseline 존재 + 페어 ---


def test_three_extensions_present() -> None:
    """3종 extension (security / testing / performance) 모두 존재."""
    for ext in EXTENSIONS:
        baseline_path = EXTENSIONS_DIR / ext["baseline"]
        opt_in_path = EXTENSIONS_DIR / ext["opt_in"]
        assert baseline_path.exists(), f"baseline missing: {ext['name']}"
        assert opt_in_path.exists(), f"opt-in missing: {ext['name']}"


def test_no_extra_baseline_files() -> None:
    """3종 baseline + SCHEMA.md 외 다른 baseline file 없음."""
    expected_basenames = {ext["baseline"] for ext in EXTENSIONS} | {"SCHEMA.md"}
    actual = sorted(p.name for p in EXTENSIONS_DIR.iterdir() if p.is_file())
    # 다른 file 이 있어도 SCHEMA + 3종 baseline + 3종 opt-in 모두 있어야 함
    for ext in EXTENSIONS:
        assert ext["baseline"] in actual, f"missing baseline: {ext['baseline']}"
        assert ext["opt_in"] in actual, f"missing opt-in: {ext['opt_in']}"
    assert "SCHEMA.md" in actual, "missing SCHEMA.md"


# --- Test 3: 각 baseline 의 rule 정의 검증 ---


def test_baseline_rule_count() -> None:
    """각 baseline 이 6 rule 정의 (TST/PERF 의 경우 6 rule)."""
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue  # security 는 v0.7.0 step 8 에서 별도 검증 (check_security_baseline.py)
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        rule_ids = [m for m in RULE_HEADER_RE.findall(content) if m[0].startswith(ext["prefix"])]
        assert len(rule_ids) == ext["rule_count"], \
            f"{ext['name']}: expected {ext['rule_count']} rules, got {len(rule_ids)}"


def test_baseline_rule_id_format() -> None:
    """각 baseline 의 rule ID 가 PREFIX-WF-NN 형식."""
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        for rid, _ in RULE_HEADER_RE.findall(content):
            assert RULE_ID_RE.match(rid), f"invalid rule ID: {rid}"
            assert rid.startswith(ext["prefix"]), \
                f"{rid} should start with {ext['prefix']}"


def test_baseline_has_rule_and_verification() -> None:
    """각 baseline 의 rule 이 **Rule** + **Verification** 형식."""
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        for rid, _ in RULE_HEADER_RE.findall(content):
            # rid 의 rule block 추출
            rule_pos_match = re.search(
                rf"### \d+\.\d+ Rule {rid}:.*?(?=\n### \d+\.\d+|\n## |\Z)",
                content,
                re.DOTALL,
            )
            assert rule_pos_match is not None, f"{rid} block not found"
            block = rule_pos_match.group(0)
            assert "**Rule**" in block, f"{rid} missing **Rule**"
            assert "**Verification**" in block, f"{rid} missing **Verification**"
            # Verification 의 bullet ≥ 3
            verification_match = re.search(
                rf"\*\*Verification\*\*:?[ \t]*\n(.*?)(?=\n### \d+\.\d+|\n## |\Z)",
                block,
                re.DOTALL,
            )
            if verification_match:
                section = verification_match.group(1)
                bullets = VERIFICATION_BULLET_RE.findall(section)
                assert len(bullets) >= 3, \
                    f"{rid} has {len(bullets)} bullets (need ≥ 3)"


def test_baseline_compliance_summary() -> None:
    """각 baseline 의 Compliance Summary table 에 모든 rule ID 명시."""
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        rule_ids_in_table = {m for m in COMPLIANCE_ROW_RE.findall(content)}
        rule_ids_defined = {m for m, _ in RULE_HEADER_RE.findall(content) \
                            if m.startswith(ext["prefix"])}
        assert rule_ids_in_table == rule_ids_defined, \
            f"{ext['name']}: table rids {rule_ids_in_table} != defined rids {rule_ids_defined}"


# --- Test 4: 각 opt-in 의 형식 검증 ---


def test_opt_in_question_format() -> None:
    """각 opt-in 이 `> Question:` + `[Answer]:` 형식."""
    for ext in EXTENSIONS:
        content = (EXTENSIONS_DIR / ext["opt_in"]).read_text(encoding="utf-8")
        assert "> Question:" in content, \
            f"{ext['name']} opt-in missing '> Question:'"
        assert "[Answer]:" in content, \
            f"{ext['name']} opt-in missing '[Answer]:'"


def test_opt_in_four_options() -> None:
    """각 opt-in 이 A/B/P/X 4 옵션 모두 포함."""
    for ext in EXTENSIONS:
        content = (EXTENSIONS_DIR / ext["opt_in"]).read_text(encoding="utf-8")
        for opt in ["A)", "B)", "P)", "X)"]:
            assert opt in content, \
                f"{ext['name']} opt-in missing option {opt}"


def test_opt_in_response_processing() -> None:
    """각 opt-in 이 Response Processing section (table 형식) 명시."""
    for ext in EXTENSIONS:
        content = (EXTENSIONS_DIR / ext["opt_in"]).read_text(encoding="utf-8")
        assert "Response Processing" in content or "답변은" in content, \
            f"{ext['name']} opt-in missing response processing"


def test_opt_in_state_schema() -> None:
    """각 opt-in 이 State File Schema (YAML) 명시."""
    for ext in EXTENSIONS:
        content = (EXTENSIONS_DIR / ext["opt_in"]).read_text(encoding="utf-8")
        assert "state:" in content.lower() or "status:" in content, \
            f"{ext['name']} opt-in missing state schema"


# --- Test 5: Cross-reference 검증 ---


def test_aidlc_1차_출처_path_valid() -> None:
    """AIDLC 1차 출처 (security + testing) cross-reference 검증."""
    aidlc_root = Path("/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details")
    if not aidlc_root.exists():
        return  # SSOT 외부 — skip
    for ext in EXTENSIONS:
        if ext["aidlc_path"] is None:
            continue  # performance 는 우리 domain 적응
        path = aidlc_root / ext["aidlc_path"]
        assert path.exists(), f"AIDLC 1차 출처 not found: {path}"


def test_aidlc_artifact_count() -> None:
    """AIDLC 1차 출처의 3 extension 모두 존재 (security + testing + resiliency)."""
    aidlc_root = Path("/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions")
    if not aidlc_root.exists():
        return
    # AIDLC 3종 baseline 모두 존재
    for cat, sub in [
        ("security", "baseline"),
        ("testing", "property-based"),
        ("resiliency", "baseline"),
    ]:
        path = aidlc_root / cat / sub
        assert path.exists(), f"AIDLC {cat}/{sub} not found"


def test_baseline_aidlc_reference() -> None:
    """각 baseline 이 AIDLC 1차 출처 cross-reference (또는 우리 domain 적응 명시)."""
    for ext in EXTENSIONS:
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        if ext["aidlc_path"] is None:
            # performance 는 우리 domain 적응 — AIDLC 없음 명시
            assert "우리 domain 적응" in content or "AIDLC" in content, \
                f"{ext['name']} missing 'domain 적응' note"
        else:
            assert "AIDLC" in content, \
                f"{ext['name']} missing AIDLC cross-reference"


# --- Test 6: Rule topic coverage (중복 검증) ---


def test_no_duplicate_rule_ids() -> None:
    """3종 extension 의 rule ID 가 unique."""
    all_rids = set()
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        for rid, _ in RULE_HEADER_RE.findall(content):
            assert rid not in all_rids, f"duplicate rule ID: {rid}"
            all_rids.add(rid)


def test_extension_unique_prefix() -> None:
    """3종 extension 의 prefix 가 unique."""
    prefixes = [ext["prefix"] for ext in EXTENSIONS]
    assert len(prefixes) == len(set(prefixes)), f"duplicate prefix: {prefixes}"


def test_each_extension_has_six_rules() -> None:
    """testing + performance 가 6 rule (security 는 별도 test)."""
    for ext in EXTENSIONS:
        if ext["name"] == "security":
            continue
        content = (EXTENSIONS_DIR / ext["baseline"]).read_text(encoding="utf-8")
        rule_ids = [rid for rid, _ in RULE_HEADER_RE.findall(content)
                    if rid.startswith(ext["prefix"])]
        assert len(rule_ids) == 6, \
            f"{ext['name']}: expected 6 rules, got {len(rule_ids)}"


# --- Test 7: extension_helper_contract 가이드 (SPEC level) ---


def test_schema_documents_helper_contract() -> None:
    """SCHEMA.md 가 §6 helper contract (v0.7.1+ follow-up) 명시."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    assert "## 6. Helper Contract" in content, "SCHEMA missing §6"
    assert "evaluate_compliance" in content, "SCHEMA missing evaluate_compliance"


def test_schema_documents_follow_up() -> None:
    """SCHEMA.md 가 §10 Follow-up (v0.7.1+) 명시."""
    content = SCHEMA_PATH.read_text(encoding="utf-8")
    assert "## 10. Follow-up" in content, "SCHEMA missing §10"
    assert "v0.7.1+" in content, "SCHEMA missing v0.7.1+ reference"


# --- Test 7: v0.7.2+ sub-cat + 4종 baseline ---


def test_sub_cat_four_baselines_present() -> None:
    """v0.7.2+ sub-cat 3 + 4종 1 = 4 baseline 모두 존재 (auth / pbt / memory / resiliency)."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["baseline_path"]
        assert path.exists(), f"sub-cat baseline missing: {ext['baseline_path']}"


def test_sub_cat_four_opt_ins_present() -> None:
    """v0.7.2+ sub-cat 3 + 4종 1 = 4 opt-in 모두 존재."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["opt_in_path"]
        assert path.exists(), f"sub-cat opt-in missing: {ext['opt_in_path']}"


def test_sub_cat_rule_count() -> None:
    """v0.7.2+ sub-cat 4종 의 rule count = 6 (resiliency = 8) baseline."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["baseline_path"]
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        rule_ids = [
            m for m in RULE_HEADER_RE.findall(content)
            if m[0].startswith(ext["prefix"])
        ]
        expected = ext["rule_count"]
        assert len(rule_ids) == expected, \
            f"{ext['name']}: expected {expected} rules, got {len(rule_ids)}"


def test_sub_cat_rule_id_format() -> None:
    """sub-cat 4종 의 rule ID 가 PREFIX-WF-NN 형식."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["baseline_path"]
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        for rid, _ in RULE_HEADER_RE.findall(content):
            assert rid.startswith(ext["prefix"]), \
                f"{rid} should start with {ext['prefix']}"


def test_sub_cat_opt_in_question_format() -> None:
    """sub-cat 4종 의 opt-in 이 `> Question:` + `[Answer]:` 형식."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["opt_in_path"]
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        assert "> Question:" in content
        assert "[Answer]:" in content


def test_sub_cat_aidlc_reference() -> None:
    """sub-cat 4종 의 1차 출처 cross-reference (AIDLC 또는 우리 domain 적응 명시)."""
    for ext in SUB_CAT_EXTENSIONS:
        path = EXTENSIONS_DIR / ext["baseline_path"]
        if not path.exists():
            continue
        content = path.read_text(encoding="utf-8")
        if ext["aidlc_path"] is None:
            # 우리 domain 적응 — AIDLC 없음 명시
            assert "우리 domain 적응" in content or "domain 적응" in content, \
                f"{ext['name']} missing 'domain 적응' note"
        else:
            assert "AIDLC" in content, f"{ext['name']} missing AIDLC reference"


def test_sub_cat_unique_prefix() -> None:
    """sub-cat 4종 의 prefix 가 unique."""
    prefixes = [ext["prefix"] for ext in SUB_CAT_EXTENSIONS]
    assert len(prefixes) == len(set(prefixes)), f"duplicate prefix: {prefixes}"


# --- 메인 실행 ---


def main() -> int:
    """모든 test 실행 후 결과 출력."""
    test_funcs = [
        test_schema_exists,
        test_schema_has_directory_layout,
        test_schema_has_file_format,
        test_schema_has_prefix_convention,
        test_schema_has_lint_rule_section,
        test_three_extensions_present,
        test_no_extra_baseline_files,
        test_baseline_rule_count,
        test_baseline_rule_id_format,
        test_baseline_has_rule_and_verification,
        test_baseline_compliance_summary,
        test_opt_in_question_format,
        test_opt_in_four_options,
        test_opt_in_response_processing,
        test_opt_in_state_schema,
        test_aidlc_1차_출처_path_valid,
        test_sub_cat_four_baselines_present,
        test_sub_cat_four_opt_ins_present,
        test_sub_cat_rule_count,
        test_sub_cat_rule_id_format,
        test_sub_cat_opt_in_question_format,
        test_sub_cat_aidlc_reference,
        test_sub_cat_unique_prefix,
        test_aidlc_artifact_count,
        test_baseline_aidlc_reference,
        test_no_duplicate_rule_ids,
        test_extension_unique_prefix,
        test_each_extension_has_six_rules,
        test_schema_documents_helper_contract,
        test_schema_documents_follow_up,
    ]

    passed = 0
    failed = 0
    failures: list[tuple[str, str]] = []
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
            failures.append((func.__name__, str(e)))
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1
            failures.append((func.__name__, f"{type(e).__name__}: {e}"))

    print()
    if failed == 0:
        print(f"All {passed} tests passed.")
        return 0
    print(f"{failed}/{passed + failed} tests failed:")
    for name, err in failures:
        print(f"  - {name}: {err}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
