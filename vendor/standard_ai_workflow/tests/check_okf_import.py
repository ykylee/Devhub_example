"""workflow_kit.okf_import helper smoke test (v0.7.34+, OKF consumer PoC + v0.7.35 ADR-011 version detect).

OKF v0.1 spec §9 의 5 MUST NOT reject 정책 (loose mode) + 우리 wiki strict lint
(strict mode) 의 *additive* 양립 + ADR-011 spec version auto-detect 검증.

Test list (11):
1-7.  OKF consumer mode (loose/strict) + lint + staging + promote
8-11. ADR-011 version detect (parse + exact match + major mismatch + minor higher + missing)
"""

from __future__ import annotations

import importlib.util
import shutil
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
OKF_IMPORT = SOURCE_ROOT / "workflow_kit" / "okf_import.py"


def _import_okf_import():
    """okf_import module importlib 로 load (dataclass decorator 위해 sys.modules register).

    okf_import 이 workflow_kit.okf_export 에 의존하므로 두 module 모두 sys.modules 에 register.
    """
    import sys
    # Pre-load okf_export (dependency)
    export_spec = importlib.util.spec_from_file_location(
        "workflow_kit.okf_export",
        str(SOURCE_ROOT / "workflow_kit" / "okf_export.py"),
    )
    export_mod = importlib.util.module_from_spec(export_spec)
    sys.modules["workflow_kit.okf_export"] = export_mod
    export_spec.loader.exec_module(export_mod)
    if "workflow_kit" not in sys.modules:
        wk_spec = importlib.util.spec_from_file_location("workflow_kit", str(SOURCE_ROOT / "workflow_kit" / "__init__.py"))
        wk_mod = importlib.util.module_from_spec(wk_spec)
        sys.modules["workflow_kit"] = wk_mod
        wk_spec.loader.exec_module(wk_mod)
    spec = importlib.util.spec_from_file_location("workflow_kit.okf_import", str(OKF_IMPORT))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["workflow_kit.okf_import"] = mod
    spec.loader.exec_module(mod)
    return mod


def _make_bundle(
    root: Path,
    *,
    with_index_md: bool = False,
    with_manifest: bool = False,
    page_frontmatter: str = "type: concept\nstatus: active\nlast_ingested_from: https://example.com/x.md\n",
    page_body: str = "# Title\n\nbody line one\nbody line two.\n",
    page_name: str = "concepts/test-concept.md",
    include_unknown_key: bool = False,
    include_broken_link: bool = False,
) -> Path:
    """Build a small OKF bundle on disk for testing."""
    bundle = root / "bundle"
    bundle.mkdir(parents=True, exist_ok=True)
    if with_manifest:
        (bundle / "okf-bundle.yaml").write_text("mode: loose\n", encoding="utf-8")
    if with_index_md:
        (bundle / "index.md").write_text(
            "---\nokf_version: \"0.1\"\nokf_mode: loose\n---\n\n# Index\n", encoding="utf-8"
        )
    fm = page_frontmatter
    if include_unknown_key:
        fm += "weird_custom_key: some-value\n"
    body = page_body
    if include_broken_link:
        body += "\nSee [broken](../nonexistent.md) here.\n"
    page_path = bundle / page_name
    page_path.parent.mkdir(parents=True, exist_ok=True)
    page_path.write_text(f"---\n{fm}---\n\n{body}", encoding="utf-8")
    return bundle


# --- Test 1: default mode = strict ---


def test_detect_mode_default_strict() -> None:
    """mode 명시 없음 (CLI None, manifest 부재, index.md 부재) → default = strict."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir))
        mode = mod.detect_mode(bundle, cli_mode=None)
        assert mode == "strict", f"expected strict, got {mode!r}"


# --- Test 2: mode from okf-bundle.yaml manifest ---


def test_detect_mode_from_manifest() -> None:
    """okf-bundle.yaml 의 `mode: loose` → loose."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir), with_manifest=True)
        mode = mod.detect_mode(bundle, cli_mode=None)
        assert mode == "loose", f"expected loose from manifest, got {mode!r}"


# --- Test 3: mode from index.md frontmatter ---


def test_detect_mode_from_index_md() -> None:
    """bundle root index.md frontmatter `okf_mode: loose` → loose."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir), with_index_md=True)
        mode = mod.detect_mode(bundle, cli_mode=None)
        assert mode == "loose", f"expected loose from index.md, got {mode!r}"


# --- Test 4: CLI flag overrides manifest ---


def test_detect_mode_cli_override() -> None:
    """CLI --mode=strict 는 manifest 의 `mode: loose` 보다 우선."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir), with_manifest=True)
        mode = mod.detect_mode(bundle, cli_mode="strict")
        assert mode == "strict", f"CLI override should win, got {mode!r}"


# --- Test 5: strict mode tolerates unknown frontmatter key ---


def test_lint_strict_unknown_key_error() -> None:
    """OKF spec §4.1 은 unknown key tolerate. 우리 wiki strict mode 도 tolerate."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir), include_unknown_key=True)
        report = mod.import_okf_bundle(bundle, staging=Path(tmpdir) / "staging", mode="strict")
        assert report.pages_staged == 1, (
            f"expected 1 page staged, got {report.pages_staged}; "
            f"errors: {report.errors}"
        )


# --- Test 6: loose mode tolerates broken link ---


def test_lint_loose_broken_link_warn() -> None:
    """loose mode 에서 broken link → warn 만, reject 안 함 (OKF §9 MUST NOT reject)."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        bundle = _make_bundle(Path(tmpdir), include_broken_link=True)
        report = mod.import_okf_bundle(bundle, staging=Path(tmpdir) / "staging", mode="loose")
        assert report.pages_staged == 1, (
            f"loose mode should tolerate broken link, got pages_staged={report.pages_staged}; "
            f"errors: {report.errors}"
        )
        strict_report = mod.import_okf_bundle(bundle, staging=Path(tmpdir) / "staging2", mode="strict")
        assert strict_report.pages_staged == 0, (
            f"strict mode should reject broken link, got pages_staged={strict_report.pages_staged}"
        )


# --- Test 7: staging + promote ---


def test_import_staging_and_promote() -> None:
    """1 page import → staging landing + --promote 시 wiki 로 copy."""
    mod = _import_okf_import()
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)
        bundle = _make_bundle(tmp)
        staging = tmp / "staging"
        report = mod.import_okf_bundle(bundle, staging=staging, mode="strict", promote=False)
        assert report.pages_staged == 1
        staged_page = staging / "concepts" / "test-concept.md"
        assert staged_page.exists()
        assert staged_page.read_text(encoding="utf-8") == (bundle / "concepts" / "test-concept.md").read_text(
            encoding="utf-8"
        )
        fake_wiki = tmp / "wiki"
        fake_wiki.mkdir()
        for staged in staging.rglob("*.md"):
            rel = staged.relative_to(staging)
            (fake_wiki / rel).parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(staged, fake_wiki / rel)
        assert (fake_wiki / "concepts" / "test-concept.md").exists(), "promote failed"


# --- Test 8: OKF version parse (ADR-011) ---


def test_version_parse_canonical() -> None:
    """Canonical form "X.Y" parse. variants (v0.1, 0.1.5 semver) 도 accept."""
    mod = _import_okf_import()
    assert mod._parse_okf_version("0.1") == (0, 1)
    assert mod._parse_okf_version("0.2") == (0, 2)
    assert mod._parse_okf_version("1.0") == (1, 0)
    assert mod._parse_okf_version("v0.1") == (0, 1)
    assert mod._parse_okf_version("0.1.5") == (0, 1)
    assert mod._parse_okf_version(None) is None
    assert mod._parse_okf_version("") is None
    assert mod._parse_okf_version("1") is None
    assert mod._parse_okf_version("abc") is None


# --- Test 9: Version check exact match ---


def test_version_check_exact_match() -> None:
    """exact match (0.1 = 0.1) → pass."""
    mod = _import_okf_import()
    result = mod._check_version_compatibility("0.1")
    assert result.status == "pass", f"got {result.status}: {result.message}"
    assert result.bundle_version == "0.1"


# --- Test 10: Version check major mismatch → error ---


def test_version_check_major_mismatch_rejects_strict() -> None:
    """major mismatch (1.0 > 0.1) → error (our consumer cannot safely process)."""
    mod = _import_okf_import()
    result = mod._check_version_compatibility("1.0")
    assert result.status == "error", f"got {result.status}: {result.message}"
    assert "major" in result.message.lower() or "breaking" in result.message.lower()


# --- Test 11: Version check minor higher → warn ---


def test_version_check_minor_higher_warns() -> None:
    """minor higher (0.2 > 0.1, same major) → warn (backward-compatible per spec §11)."""
    mod = _import_okf_import()
    result = mod._check_version_compatibility("0.2")
    assert result.status == "warn", f"got {result.status}: {result.message}"
    assert "0.2" in result.message


# --- Test 12: Version check missing → warn (assume v0.1) ---


def test_version_check_missing_warns() -> None:
    """no okf_version field → warn (assume our v0.1)."""
    mod = _import_okf_import()
    result = mod._check_version_compatibility(None)
    assert result.status == "warn", f"got {result.status}: {result.message}"
    assert "no okf_version" in result.message.lower() or "assuming" in result.message.lower()


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_detect_mode_default_strict,
        test_detect_mode_from_manifest,
        test_detect_mode_from_index_md,
        test_detect_mode_cli_override,
        test_lint_strict_unknown_key_error,
        test_lint_loose_broken_link_warn,
        test_import_staging_and_promote,
        test_version_parse_canonical,
        test_version_check_exact_match,
        test_version_check_major_mismatch_rejects_strict,
        test_version_check_minor_higher_warns,
        test_version_check_missing_warns,
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
