r"""L1 wiki page format SSOT edge case tests (DevHub-specific, PR #615 follow-up).

Verifies the module-level \`validate_l1_pages\` function (defined in
tests/check_wiki_drift_devhub.py) against 9 edge case scenarios. 본 test file 의
위치 의도: SSOT function 의 unit test. main repo 의 5 L1 page 도 smoke 으로
verify (real data 가 SSOT 와 정합).

Validation rules (SSOT — docs/governance/l1-format.md):
1. frontmatter delimiter (---) 존재
2. required field 4개: type, status, created, updated
3. type enum: {concept, decision, entity, pattern, topic}
4. status enum: {active, draft}
5. created / updated date format: YYYY-MM-DD

Wiki: ai-workflow/wiki/concepts/devhub-overview.md
      ai-workflow/wiki/decisions/v0.7.37-import.md
      ai-workflow/wiki/decisions/v0.7.17-import.md
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
DRIFT_TEST = REPO_ROOT / "tests" / "check_wiki_drift_devhub.py"

if sys.platform == "win32":
    print(f"[check-l1-format] SKIP: POSIX only")
    sys.exit(0)

if not DRIFT_TEST.exists():
    print(f"[check-l1-format] SKIP: drift test not found: {DRIFT_TEST}")
    sys.exit(0)


# ----- result tracker -----
PASS = 0
FAIL = 0
FAILURES: list[str] = []


def _ok(name: str) -> None:
    global PASS
    PASS += 1
    print(f"  [PASS] {name}")


def _fail(name: str, reason: str) -> None:
    global FAIL
    FAIL += 1
    FAILURES.append(f"{name}: {reason}")
    print(f"  [FAIL] {name}: {reason}")


def _import_validate_l1_pages():
    """check_wiki_drift_devhub.py 에서 validate_l1_pages module-level function import."""
    spec = importlib.util.spec_from_file_location("drift_under_test", str(DRIFT_TEST))
    assert spec is not None and spec.loader is not None
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    return m.validate_l1_pages


validate_l1_pages = _import_validate_l1_pages()


def _make_l1(wiki_root: Path, dir_name: str, name: str, body: str) -> None:
    """Create an L1 page in temp wiki_root/dir_name/name.md."""
    d = wiki_root / dir_name
    d.mkdir(parents=True, exist_ok=True)
    (d / name).write_text(body, encoding="utf-8")


# ----- 1. valid L1 page → 0 issues -----
def test_valid_l1_page():
    name = "test_valid_l1_page"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(wiki, "concepts", "ok.md", "---\ntype: concept\nstatus: active\ncreated: 2026-06-15\nupdated: 2026-06-15\n---\n# ok\n")
            count, issues = validate_l1_pages(wiki)
            if count != 0 or issues:
                return _fail(name, f"expected 0 issues, got {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 2. missing frontmatter → fail -----
def test_missing_frontmatter():
    name = "test_missing_frontmatter"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(wiki, "concepts", "no-fm.md", "# no frontmatter\njust body\n")
            count, issues = validate_l1_pages(wiki)
            if count != 1 or "missing frontmatter" not in issues[0]:
                return _fail(name, f"expected 1 missing-frontmatter issue, got {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 3. missing required field (type) → fail -----
def test_missing_required_field():
    name = "test_missing_required_field"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(
                wiki, "concepts", "no-type.md",
                "---\nstatus: active\ncreated: 2026-06-15\nupdated: 2026-06-15\n---\n# x\n",
            )
            count, issues = validate_l1_pages(wiki)
            # missing 'type' triggers: 1 missing required + 1 invalid type (None)
            if count < 1 or not any("missing type" in i for i in issues):
                return _fail(name, f"expected 'missing type' issue, got {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 4. invalid type (e.g. 'bogus') → fail -----
def test_invalid_type():
    name = "test_invalid_type"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(
                wiki, "concepts", "bad-type.md",
                "---\ntype: bogus\nstatus: active\ncreated: 2026-06-15\nupdated: 2026-06-15\n---\n# x\n",
            )
            count, issues = validate_l1_pages(wiki)
            if count != 1 or "invalid type" not in issues[0]:
                return _fail(name, f"expected 'invalid type' issue, got {count}: {issues}")
            if "'bogus'" not in issues[0]:
                return _fail(name, f"expected value 'bogus' in error, got: {issues[0]}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 5. invalid status (e.g. 'accepted') → fail -----
def test_invalid_status():
    name = "test_invalid_status"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(
                wiki, "decisions", "bad-status.md",
                "---\ntype: decision\nstatus: accepted\ncreated: 2026-06-15\nupdated: 2026-06-15\n---\n# x\n",
            )
            count, issues = validate_l1_pages(wiki)
            if count != 1 or "invalid status" not in issues[0]:
                return _fail(name, f"expected 'invalid status' issue, got {count}: {issues}")
            if "'accepted'" not in issues[0]:
                return _fail(name, f"expected value 'accepted' in error, got: {issues[0]}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 6. invalid date format (e.g. '2026-6-15') → fail -----
def test_invalid_date_format():
    name = "test_invalid_date_format"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(
                wiki, "entities", "bad-date.md",
                "---\ntype: entity\nstatus: active\ncreated: 2026-6-15\nupdated: 2026/06/15\n---\n# x\n",
            )
            count, issues = validate_l1_pages(wiki)
            # 2 invalid date (created + updated)
            if count < 2 or not all("invalid" in i and "date format" in i for i in issues):
                return _fail(name, f"expected 2 date format issues, got {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 7. multiple issues (multi-error detection) → all reported -----
def test_multiple_issues():
    name = "test_multiple_issues"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            _make_l1(
                wiki, "patterns", "multi.md",
                # missing type, invalid status, bad date — 4 issues
                "---\nstatus: deprecated\ncreated: 2026.06.15\nupdated: foo\n---\n# x\n",
            )
            count, issues = validate_l1_pages(wiki)
            if count < 3:
                return _fail(name, f"expected >=3 issues, got {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 8. skip non-L1 dirs (sources/, raw/) → no false positive -----
def test_skip_non_l1_dirs():
    name = "test_skip_non_l1_dirs"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            # sources/ and raw/ 의 invalid file 은 skip 대상
            (wiki / "sources").mkdir(parents=True, exist_ok=True)
            (wiki / "sources" / "no-fm.md").write_text("no frontmatter\n", encoding="utf-8")
            (wiki / "raw").mkdir(parents=True, exist_ok=True)
            (wiki / "raw" / "projects").mkdir(parents=True, exist_ok=True)
            (wiki / "raw" / "projects" / "devhub" / "no-fm.md").parent.mkdir(parents=True, exist_ok=True)
            (wiki / "raw" / "projects" / "devhub" / "no-fm.md").write_text("no frontmatter\n", encoding="utf-8")
            count, issues = validate_l1_pages(wiki)
            if count != 0:
                return _fail(name, f"sources/ or raw/ false positive: {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 9. skip index.md and RAW_MIRROR_MANIFEST.md → no false positive -----
def test_skip_index_and_manifest():
    name = "test_skip_index_and_manifest"
    try:
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            wiki = Path(d)
            wiki.mkdir(parents=True, exist_ok=True)
            (wiki / "index.md").write_text("# top\n", encoding="utf-8")
            (wiki / "RAW_MIRROR_MANIFEST.md").write_text("# manifest\n", encoding="utf-8")
            count, issues = validate_l1_pages(wiki)
            if count != 0:
                return _fail(name, f"index/manifest false positive: {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- 10. real main repo (5 L1 page) → 0 issues (smoke) -----
def test_real_main_repo_5_l1_pages():
    name = "test_real_main_repo_5_l1_pages"
    try:
        wiki = REPO_ROOT / "ai-workflow" / "wiki"
        if not wiki.is_dir():
            return _fail(name, f"main wiki not found: {wiki}")
        count, issues = validate_l1_pages(wiki)
        if count != 0:
            return _fail(name, f"real 5 L1 page fail: {count}: {issues}")
        _ok(name)
    except Exception as e:
        _fail(name, f"exception: {e}")


# ----- main -----
def main() -> int:
    print(f"[check-l1-format] === L1 format SSOT edge case tests (DevHub) ===")
    print(f"[check-l1-format]   SSOT: docs/governance/l1-format.md")
    print(f"[check-l1-format]   impl: tests/check_wiki_drift_devhub.py:validate_l1_pages")
    print("")

    test_valid_l1_page()
    test_missing_frontmatter()
    test_missing_required_field()
    test_invalid_type()
    test_invalid_status()
    test_invalid_date_format()
    test_multiple_issues()
    test_skip_non_l1_dirs()
    test_skip_index_and_manifest()
    test_real_main_repo_5_l1_pages()
    print("")

    total = PASS + FAIL
    print(f"[check-l1-format] === summary ===")
    print(f"  passed: {PASS}/{total}")
    print(f"  failed: {FAIL}/{total}")
    if FAIL:
        print(f"  failures:")
        for f in FAILURES:
            print(f"    - {f}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
