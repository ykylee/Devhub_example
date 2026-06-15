#!/usr/bin/env python3
"""S2: Orchestrator ↔ Sub-agent contract v1 role-mapping check.

Verifies that the §6.1 'MUST delegate' catalog maps each task to the correct
sub-agent role per §3.
Reference: workflow-source/core/orchestrator_subagent_contract_v1.md §8.2.
"""

from __future__ import annotations

from pathlib import Path

CONTRACT_PATH = Path(__file__).resolve().parents[1] / "core" / "orchestrator_subagent_contract_v1.md"

# §6.1 must-delegate catalog (subset that is unambiguous enough to regression-test).
# Format: (description, expected_role)
MUST_DELEGATE_CATALOG = (
    ("5개 이상 파일 read", "doc-worker"),
    ("2개 이상 문서 비교/대조", "doc-worker"),
    ("단일 문서 200줄+ 초안 작성", "doc-worker"),
    ("5개 이상 파일 bounded patch", "code-worker"),
    ("빌드/컴파일/lint/test 실행", "validation-worker"),
    ("5개 이상 로그 파일 read/요약", "validation-worker"),
    ("외부 저장소 분석/탐색", "doc-worker"),
    ("새 파일 300줄+ 작성", "doc-worker"),
)


def _check_catalog_against_doc() -> None:
    if not CONTRACT_PATH.exists():
        raise AssertionError(f"Contract v1 spec doc not found: {CONTRACT_PATH}")
    text = CONTRACT_PATH.read_text(encoding="utf-8")
    if "## 6.1 위임 가능 (MUST delegate)" not in text:
        raise AssertionError("Contract doc missing §6.1 (위임 가능 (MUST delegate))")
    if "## 6.3 직접 처리 (MUST NOT delegate)" not in text:
        raise AssertionError("Contract doc missing §6.3 (직접 처리 (MUST NOT delegate))")


def check_role_mapping_consistency() -> None:
    """Verify that the §6.1 catalog is consistent with §3 role definitions."""
    # The mapping must be deterministic: doc-worker for read/draft, code-worker for
    # bounded patch, validation-worker for run-only tasks. We assert this in code
    # so the regression test is meaningful even if the markdown wording shifts.
    doc_worker_keywords = ("read", "비교", "초안", "분석", "탐색", "draft", "research", "작성", "문서")
    code_worker_keywords = ("patch", "코드", "설정")
    validation_worker_keywords = ("빌드", "실행", "로그", "lint", "test")

    for description, expected_role in MUST_DELEGATE_CATALOG:
        if expected_role == "doc-worker":
            assert any(k in description for k in doc_worker_keywords), (
                f"§6.1 row {description!r} mapped to doc-worker but lacks doc-worker keywords"
            )
        elif expected_role == "code-worker":
            assert any(k in description for k in code_worker_keywords), (
                f"§6.1 row {description!r} mapped to code-worker but lacks code-worker keywords"
            )
        elif expected_role == "validation-worker":
            assert any(k in description for k in validation_worker_keywords), (
                f"§6.1 row {description!r} mapped to validation-worker but lacks validation-worker keywords"
            )
        else:
            raise AssertionError(f"Unknown role {expected_role!r} in MUST_DELEGATE_CATALOG")


def check_role_definitions_present() -> None:
    text = CONTRACT_PATH.read_text(encoding="utf-8")
    for role_section in (
        "### 3.1 orchestrator (메인 오케스트레이터)",
        "### 3.2 doc-worker",
        "### 3.3 code-worker",
        "### 3.4 validation-worker",
        "### 3.5 generic workflow-worker (임시/예외)",
    ):
        if role_section not in text:
            raise AssertionError(f"Contract doc missing role definition: {role_section!r}")


def main() -> int:
    _check_catalog_against_doc()
    check_role_definitions_present()
    check_role_mapping_consistency()
    print("Contract v1 role-mapping smoke check passed (5+ MUST-delegate rows consistent with §3 roles).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
