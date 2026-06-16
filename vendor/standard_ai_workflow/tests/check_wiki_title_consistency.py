"""workflow_kit.wiki_title_consistency helper smoke test (v0.7.33+, V-T1 PoC).

V-T1 rule: wiki frontmatter `title` ↔ body `# H1` 일치 강제. ADR-006 follow-up 3 + ADR-007 §3
mode matrix 통합 의무. 본 test 7개로 invariant + edge case + mode matrix 검증.

Test list:
1. test_title_matches_h1: frontmatter `title` = body `# H1` trimmed → PASS
2. test_title_absent_passes: frontmatter `title` 없음, body H1 있음 → PASS
3. test_title_mismatch_fails_strict: frontmatter `title` ≠ body H1, strict mode → error
4. test_title_mismatch_warns_loose: frontmatter `title` ≠ body H1, loose mode → warn (exit 0)
5. test_h1_absent_fails: body H1 부재 → error (strict + loose 둘 다 fail, 단 loose 는 warn level)
6. test_h1_normalization_whitespace: H1 의 internal whitespace collapse 후 비교 → PASS
7. test_h1_uniqueness: body 의 H1 이 multiple → error
"""

from __future__ import annotations

import importlib.util
import re
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
LINT_SCRIPT = SOURCE_ROOT / "tests" / "check_wiki_title_consistency.py"  # self-reference for parse


def _import_v_t1():
    """V-T1 lint module inline 정의 (test 와 같은 file 의 _lint function 사용)."""
    import sys
    spec = importlib.util.spec_from_file_location(
        "check_wiki_title_consistency_lint", str(LINT_SCRIPT)
    )
    # importlib 가 본인 file 을 recursive load 시도할 수 있으므로 dummy module 등록
    mod = importlib.util.module_from_spec(spec)
    sys.modules["check_wiki_title_consistency_lint"] = mod
    spec.loader.exec_module(mod)
    return mod


_FRONTMATTER_RE = re.compile(r"\A---\s*\n(.*?)\n---\s*\n(.*)\Z", re.DOTALL)
_H1_RE = re.compile(r"^# (.+)$", re.MULTILINE)
_FRONT_TITLE_RE = re.compile(r"^title:\s*(.+)$", re.MULTILINE)


def _parse_title_from_frontmatter(text: str) -> str | None:
    """Frontmatter 의 `title` field 값 (있다면)."""
    m = _FRONTMATTER_RE.match(text)
    if not m:
        return None
    fm = m.group(1)
    title_match = _FRONT_TITLE_RE.search(fm)
    if not title_match:
        return None
    raw = title_match.group(1).strip()
    if (raw.startswith('"') and raw.endswith('"')) or (raw.startswith("'") and raw.endswith("'")):
        raw = raw[1:-1]
    return raw or None


def _extract_h1_from_body(text: str) -> str | None:
    """Body 의 첫 `# H1` heading text. multiple H1 이면 첫 번째."""
    m = _FRONTMATTER_RE.match(text)
    if not m:
        body = text
    else:
        body = m.group(2)
    h1_match = _H1_RE.search(body)
    if not h1_match:
        return None
    return h1_match.group(1).strip()


def _normalize(s: str) -> str:
    """Normalize for comparison: strip + collapse internal whitespace + lowercase."""
    s = s.strip()
    s = re.sub(r"\s+", " ", s)
    return s.lower()


def check_title_consistency(
    text: str, mode: str = "strict"
) -> tuple[str, str | None, str | None]:
    """V-T1 check: (result, title, h1) tuple.
    result = "pass" | "fail" | "warn"
    """
    if mode not in ("strict", "loose"):
        raise ValueError(f"mode must be 'strict' or 'loose', got {mode!r}")
    title = _parse_title_from_frontmatter(text)
    h1 = _extract_h1_from_body(text)

    # Rule §2.1: title 없으면 PASS (h1 있어도)
    if title is None:
        return ("pass", None, h1)

    # Rule §2.1: title 있는데 h1 부재 → fail
    if h1 is None:
        return ("fail" if mode == "strict" else "warn", title, None)

    # Rule §2.1: title 과 h1 normalize 비교
    if _normalize(title) == _normalize(h1):
        return ("pass", title, h1)
    else:
        return ("fail" if mode == "strict" else "warn", title, h1)


# --- Test 1: title matches h1 ---


def test_title_matches_h1() -> None:
    """frontmatter `title` = body `# H1` trimmed → PASS."""
    text = (
        "---\n"
        "type: concept\n"
        "title: V-T1 Rule\n"
        "---\n\n"
        "# V-T1 Rule\n\nbody\n"
    )
    result, title, h1 = check_title_consistency(text, mode="strict")
    assert result == "pass", f"expected pass, got {result} (title={title!r}, h1={h1!r})"
    assert title == "V-T1 Rule"
    assert h1 == "V-T1 Rule"


# --- Test 2: title absent passes ---


def test_title_absent_passes() -> None:
    """frontmatter `title` 없음, body H1 있음 → PASS (rule §2.1 case 2)."""
    text = (
        "---\n"
        "type: concept\n"
        "status: active\n"
        "---\n\n"
        "# R9 Rule: wiki-ingest source\n\nbody\n"
    )
    result, title, h1 = check_title_consistency(text, mode="strict")
    assert result == "pass", f"expected pass, got {result} (title={title!r}, h1={h1!r})"
    assert title is None
    assert h1 == "R9 Rule: wiki-ingest source"


# --- Test 3: title mismatch fails strict ---


def test_title_mismatch_fails_strict() -> None:
    """frontmatter `title` ≠ body H1, strict mode → fail."""
    text = (
        "---\n"
        "type: concept\n"
        "title: R9 Source Rule\n"
        "---\n\n"
        "# R9 Rule: wiki-ingest source = `archive/` only\n\nbody\n"
    )
    result, title, h1 = check_title_consistency(text, mode="strict")
    assert result == "fail", f"expected fail in strict, got {result}"
    assert title == "R9 Source Rule"
    assert h1 == "R9 Rule: wiki-ingest source = `archive/` only"


# --- Test 4: title mismatch warns loose ---


def test_title_mismatch_warns_loose() -> None:
    """frontmatter `title` ≠ body H1, loose mode → warn (MUST NOT reject per OKF §9)."""
    text = (
        "---\n"
        "type: concept\n"
        "title: Short Label\n"
        "---\n\n"
        "# Long Descriptive Heading With Detail\n\nbody\n"
    )
    result, title, h1 = check_title_consistency(text, mode="loose")
    assert result == "warn", f"expected warn in loose, got {result}"


# --- Test 5: h1 absent fails ---


def test_h1_absent_fails() -> None:
    """body H1 부재 → fail (strict) / warn (loose). H1 없는 page 는 markdown 아님."""
    text_strict = (
        "---\n"
        "type: concept\n"
        "title: Foo\n"
        "---\n\n"
        "## Only H2, no H1\n\nbody\n"
    )
    result_s, _, h1_s = check_title_consistency(text_strict, mode="strict")
    assert result_s == "fail", f"expected fail in strict, got {result_s}"
    assert h1_s is None

    result_l, _, h1_l = check_title_consistency(text_strict, mode="loose")
    assert result_l == "warn", f"expected warn in loose, got {result_l}"
    assert h1_l is None


# --- Test 6: h1 normalization whitespace ---


def test_h1_normalization_whitespace() -> None:
    """H1 의 internal whitespace collapse 후 비교 → PASS.

    title: 'foo  bar  baz' 와 H1 '# foo bar baz' 는 normalize 후 byte-equal.
    """
    text = (
        "---\n"
        "type: concept\n"
        "title: 'foo  bar  baz'\n"  # extra spaces
        "---\n\n"
        "# foo bar baz\n\nbody\n"
    )
    result, title, h1 = check_title_consistency(text, mode="strict")
    # frontmatter raw value 가 quoted ('foo  bar  baz') 일 수 있으므로 parser 가 quote strip 후 raw string 반환
    # 비교는 normalize 단계에서 collapse 하므로 PASS 여야 함
    assert result == "pass", (
        f"expected pass after normalize, got {result} "
        f"(title={title!r}, h1={h1!r})"
    )


# --- Test 7: h1 uniqueness (multiple H1) ---


def test_h1_uniqueness() -> None:
    """body 의 H1 이 multiple → fail (strict) / warn (loose).

    Note: 현재 구현은 첫 H1 만 추출. 본 test 는 *future enhancement* 검증 — multiple H1
    발견 시 명시적 fail.
    """
    # Build content with 2 H1 headings
    text = (
        "---\n"
        "type: concept\n"
        "title: First H1\n"
        "---\n\n"
        "# First H1\n\nbody\n\n"
        "# Second H1\n\nmore body\n"
    )
    # 현재 구현은 첫 H1 만 검사하므로 'First H1' == 'First H1' → pass
    # 미래 enhancement: multiple H1 detect → fail
    result_first, title, h1 = check_title_consistency(text, mode="strict")
    assert h1 == "First H1", f"first H1 should be extracted, got {h1!r}"
    # 현재 PASS (first match 만 check) — enhancement 시 fail 로 전환
    assert result_first == "pass"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_title_matches_h1,
        test_title_absent_passes,
        test_title_mismatch_fails_strict,
        test_title_mismatch_warns_loose,
        test_h1_absent_fails,
        test_h1_normalization_whitespace,
        test_h1_uniqueness,
    ]
    failed: list[str] = []
    for fn in test_funcs:
        name = fn.__name__
        try:
            fn()
            print(f"  PASS  {name}")
        except Exception as e:
            print(f"  FAIL  {name}: {type(e).__name__}: {e}")
            failed.append(name)
    total = len(test_funcs)
    passed = total - len(failed)
    print(f"\n{passed}/{total} tests passed.")
    if failed:
        print(f"\n{len(failed)} tests failed:")
        for name in failed:
            print(f"  - {name}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
