#!/usr/bin/env python3
"""v0.7.0 step 6: Reverse Engineering 9-Artifact 검증.

- workflow-source/reverse-engineering/ 9 artifact file 존재
- 각 artifact 의 Verification subsection 존재
- core/reverse_engineering.md step 가이드 존재
- artifact ID 정합성 (01..09 순서)
- AIDLC 1차 출처 cross-reference
- 우리 적응 주석 (Workflow domain 적응: ...) 9 file 모두 포함
- state.json reverse_engineering 필드 schema hint (spec level)
- Rerun stale check 의 3 분기 명시

Reference: workflow-source/core/reverse_engineering.md
           workflow-source/reverse-engineering/01..09-*.md
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

REVERSE_ENG_DIR = SOURCE_ROOT / "reverse-engineering"
GUIDE_PATH = SOURCE_ROOT / "core" / "reverse_engineering.md"

# 9 artifact 이름 (순서대로)
ARTIFACT_NAMES = [
    "01-business-overview.md",
    "02-architecture.md",
    "03-code-structure.md",
    "04-api-documentation.md",
    "05-component-inventory.md",
    "06-technology-stack.md",
    "07-dependencies.md",
    "08-code-quality-assessment.md",
    "09-reverse-engineering-metadata.md",
]

# 각 artifact 의 AIDLC 대응 (smoke test 가 1:1 검증)
AIDLC_CORRESPONDENCE = {
    "01-business-overview.md": "business-overview",
    "02-architecture.md": "architecture",
    "03-code-structure.md": "code-structure",
    "04-api-documentation.md": "api-documentation",
    "05-component-inventory.md": "component-inventory",
    "06-technology-stack.md": "technology-stack",
    "07-dependencies.md": "dependencies",
    "08-code-quality-assessment.md": "code-quality-assessment",
    "09-reverse-engineering-metadata.md": "reverse-engineering-timestamp",
}


# --- Test 1: 디렉토리 + 9 artifact 존재 ---


def test_reverse_eng_dir_exists() -> None:
    """reverse-engineering/ 디렉토리 존재."""
    assert REVERSE_ENG_DIR.exists(), f"dir not found: {REVERSE_ENG_DIR}"
    assert REVERSE_ENG_DIR.is_dir(), f"not a directory: {REVERSE_ENG_DIR}"


def test_guide_exists() -> None:
    """core/reverse_engineering.md step 가이드 존재."""
    assert GUIDE_PATH.exists(), f"guide not found: {GUIDE_PATH}"


def test_nine_artifacts_defined() -> None:
    """9 artifact 모두 존재 (01..09)."""
    for name in ARTIFACT_NAMES:
        path = REVERSE_ENG_DIR / name
        assert path.exists(), f"artifact missing: {name}"


def test_artifact_count_exactly_nine() -> None:
    """9 artifact 외 다른 file 없음 (e.g. .DS_Store / 임시 file)."""
    actual = sorted(p.name for p in REVERSE_ENG_DIR.iterdir() if p.is_file())
    expected = sorted(ARTIFACT_NAMES)
    assert actual == expected, f"mismatch: actual={actual} expected={expected}"


# --- Test 2: artifact 내용 검증 ---


def test_each_artifact_has_verification_section() -> None:
    """각 artifact 가 Verification subsection 포함."""
    for name in ARTIFACT_NAMES:
        content = (REVERSE_ENG_DIR / name).read_text(encoding="utf-8")
        assert "## Verification" in content, f"{name} missing ## Verification"


def test_each_artifact_has_workflow_adaptation_note() -> None:
    """각 artifact 가 우리 적응 주석 (`Workflow domain 적응:`) 포함."""
    for name in ARTIFACT_NAMES:
        content = (REVERSE_ENG_DIR / name).read_text(encoding="utf-8")
        assert "Workflow domain 적응" in content, \
            f"{name} missing 'Workflow domain 적응' note"


def test_each_artifact_has_aidlc_reference() -> None:
    """각 artifact 가 AIDLC 1차 출처 cross-reference 포함."""
    for name in ARTIFACT_NAMES:
        content = (REVERSE_ENG_DIR / name).read_text(encoding="utf-8")
        assert "AIDLC" in content and "reverse-engineering.md" in content, \
            f"{name} missing AIDLC cross-reference"


def test_artifact_sequential_numbering() -> None:
    """artifact 파일명 prefix 가 01..09 순차."""
    for i, name in enumerate(ARTIFACT_NAMES, start=1):
        prefix = f"{i:02d}"
        assert name.startswith(prefix + "-"), \
            f"{name} should start with {prefix}-"


# --- Test 3: guide 내용 검증 ---


def test_guide_has_thirteen_steps() -> None:
    """core 가이드가 Step 1-13 (AIDLC 13 step) 모두 언급."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    for n in range(1, 14):
        # "Step N" 또는 "Step N-M" range 표기 매칭
        # 예: "Step 2-9" 는 2, 3, 4, 5, 6, 7, 8, 9 모두 포함
        if re.search(rf"Step\s+{n}\b", content):
            continue
        # range 표기: "Step A-B" 에서 n 이 [A, B] 범위 내
        range_match = re.search(rf"Step\s+(\d+)-(\d+)\b", content)
        if range_match:
            a, b = int(range_match.group(1)), int(range_match.group(2))
            if a <= n <= b:
                continue
        # 별도 표기: "Step X: ..." 또는 "X.1" 등
        assert False, f"Step {n} missing in guide"


def test_guide_has_9_artifact_table() -> None:
    """가이드가 9-Artifact 매핑 table 포함."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    for name in ARTIFACT_NAMES:
        assert name in content, f"guide missing artifact reference: {name}"


def test_guide_has_aidlc_correspondence() -> None:
    """가이드가 AIDLC artifact name 매핑 명시."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    for aidlc_name in AIDLC_CORRESPONDENCE.values():
        assert aidlc_name in content, f"guide missing AIDLC ref: {aidlc_name}"


def test_guide_has_rerun_stale_check() -> None:
    """가이드가 rerun stale check 3 분기 명시 (fresh/stale/explicit rerun)."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    assert "fresh" in content and "stale" in content, \
        "guide missing fresh/stale branches"
    assert "explicit rerun" in content or "사용자 explicit" in content, \
        "guide missing explicit rerun branch"


def test_guide_has_state_json_schema() -> None:
    """가이드가 state.json reverse_engineering schema hint 포함."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    # JSON snippet with reverse_engineering key
    assert "reverse_engineering" in content, \
        "guide missing state.json reverse_engineering field"
    assert "last_generated" in content, \
        "guide missing last_generated field"
    assert "artifact_count" in content, \
        "guide missing artifact_count field"


def test_guide_has_workflow_pattern_adaptation() -> None:
    """가이드가 우리 사용 패턴 적응 section (§6) 포함."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    assert "우리 사용 패턴 적응" in content, "guide missing §6"
    # AIDLC 와 우리 차이점 명시
    assert "Lambda" in content or "HTTP" in content, \
        "guide missing AIDLC-vs-ours adaptation example"


# --- Test 4: cross-reference 검증 ---


def test_aidlc_1차_출처_path_valid() -> None:
    """AIDLC 1차 출처 (commit b19c819) cross-reference 검증."""
    aidlc_path = Path("/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md")
    if not aidlc_path.exists():
        # SSOT 외부 — spec level 만 검증, skip
        return
    # 1차 출처가 311 line 근처여야 함 (311 ± 10)
    line_count = sum(1 for _ in aidlc_path.open(encoding="utf-8"))
    assert 280 <= line_count <= 340, \
        f"AIDLC source line count drift: {line_count} (expected 311 ± 30)"


def test_aidlc_artifact_count_matches() -> None:
    """AIDLC 1차 출처가 9 artifact 모두 정의 (step 2-10)."""
    aidlc_path = Path("/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md")
    if not aidlc_path.exists():
        return
    content = aidlc_path.read_text(encoding="utf-8")
    # 9 artifact file name 이 create 명령에 등장
    artifact_files = [
        "business-overview.md",
        "architecture.md",
        "code-structure.md",
        "api-documentation.md",
        "component-inventory.md",
        "technology-stack.md",
        "dependencies.md",
        "code-quality-assessment.md",
        "reverse-engineering-timestamp.md",
    ]
    for fname in artifact_files:
        assert fname in content, f"AIDLC source missing artifact: {fname}"


# --- Test 5: R-1~R9 lint 정합 ---


def test_no_duplicate_artifact_filename() -> None:
    """9 artifact 외 중복 file name 없음."""
    actual = [p.name for p in REVERSE_ENG_DIR.iterdir() if p.is_file()]
    assert len(actual) == len(set(actual)), f"duplicate filenames: {actual}"


def test_artifact_consistent_naming() -> None:
    """모든 artifact 가 NN-kebab-case.md 형식."""
    pattern = re.compile(r"^\d{2}-[a-z]+(?:-[a-z]+)*\.md$")
    for name in ARTIFACT_NAMES:
        assert pattern.match(name), f"{name} does not match NN-kebab-case.md pattern"


def test_guide_links_to_artifact_dir() -> None:
    """가이드가 workflow-source/reverse-engineering/ 경로 명시."""
    content = GUIDE_PATH.read_text(encoding="utf-8")
    assert "workflow-source/reverse-engineering/" in content, \
        "guide missing artifact dir path"


# --- 메인 실행 ---


def main() -> int:
    """모든 test 실행 후 결과 출력."""
    test_funcs = [
        test_reverse_eng_dir_exists,
        test_guide_exists,
        test_nine_artifacts_defined,
        test_artifact_count_exactly_nine,
        test_each_artifact_has_verification_section,
        test_each_artifact_has_workflow_adaptation_note,
        test_each_artifact_has_aidlc_reference,
        test_artifact_sequential_numbering,
        test_guide_has_thirteen_steps,
        test_guide_has_9_artifact_table,
        test_guide_has_aidlc_correspondence,
        test_guide_has_rerun_stale_check,
        test_guide_has_state_json_schema,
        test_guide_has_workflow_pattern_adaptation,
        test_aidlc_1차_출처_path_valid,
        test_aidlc_artifact_count_matches,
        test_no_duplicate_artifact_filename,
        test_artifact_consistent_naming,
        test_guide_links_to_artifact_dir,
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
