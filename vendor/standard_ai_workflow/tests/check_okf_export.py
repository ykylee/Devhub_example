"""workflow_kit.okf_export helper smoke test (v0.7.33+, OKF PoC).

OKF v0.1 spec 의 frontmatter mapping (SPEC.md §4.1) + body rewriting (§5.1) 검증.
PoC 단계: 7 test 로 핵심 mapping + body rewrite + CLI 동작 검증.

Test list:
1. test_frontmatter_parse_minimal: minimal wiki frontmatter (type 만) 파싱
2. test_frontmatter_parse_full: 모든 field 가 있는 wiki frontmatter 파싱
3. test_frontmatter_parse_missing_type_raises: type 없으면 InvalidFrontmatterError
4. test_map_to_okf_field_order: SPEC.md §4.1 priority order (type → title → description → resource → tags → timestamp)
5. test_map_to_okf_derives_title_from_body: frontmatter 에 title 없을 때 body H1 에서 derive
6. test_rewrite_wiki_links: [[path]] → [text](../path.md), [[path#anchor]] → [text](../path.md#anchor)
7. test_export_wiki_to_okf_end_to_end: 1 page export → OKF spec required field + body link rewrite 검증
"""

from __future__ import annotations

import importlib.util
import re
import sys
import tempfile
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
OKF_EXPORT = SOURCE_ROOT / "workflow_kit" / "okf_export.py"

def _import_okf_export():
    """okf_export module importlib 로 load. dataclass decorator 가
    sys.modules 에서 호출 module 을 lookup 하므로 명시적 register 필수."""
    import sys
    spec = importlib.util.spec_from_file_location("okf_export", str(OKF_EXPORT))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["okf_export"] = mod  # dataclass 가 require
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: minimal frontmatter 파싱 ---


def test_frontmatter_parse_minimal() -> None:
    """type 만 있는 minimal wiki frontmatter 파싱."""
    mod = _import_okf_export()
    text = "---\ntype: concept\n---\n\n# Title\n\nbody\n"
    fm = mod.Frontmatter.parse(text)
    assert fm.type == "concept", f"type mismatch: {fm.type!r}"
    assert fm.status is None, f"status should be None, got {fm.status!r}"
    assert fm.title is None
    assert fm.related_pages == (), f"related_pages should be empty, got {fm.related_pages!r}"


# --- Test 2: 모든 field 가 있는 full frontmatter ---


def test_frontmatter_parse_full() -> None:
    """모든 field 가 있는 wiki frontmatter 파싱 — list field + bool field + nested 구조 확인."""
    mod = _import_okf_export()
    text = (
        "---\n"
        "type: decision\n"
        "status: accepted\n"
        "title: ADR-001\n"
        "description: 3-layer separation\n"
        "last_ingested_from: docs/architecture/ADR-001.md\n"
        "created: 2026-05-01\n"
        "updated: 2026-05-15\n"
        "related_pages: [concepts/foo, concepts/bar]\n"
        "tags: [architecture, layer]\n"
        "adr_id: ADR-001\n"
        "r9_skip: true\n"
        "---\n\n"
        "# ADR-001\n\nbody\n"
    )
    fm = mod.Frontmatter.parse(text)
    assert fm.type == "decision"
    assert fm.status == "accepted"
    assert fm.title == "ADR-001"
    assert fm.description == "3-layer separation"
    assert fm.last_ingested_from == "docs/architecture/ADR-001.md"
    assert fm.created == "2026-05-01"
    assert fm.updated == "2026-05-15"
    assert fm.related_pages == ("concepts/foo", "concepts/bar"), fm.related_pages
    assert fm.tags == ("architecture", "layer"), fm.tags
    assert fm.adr_id == "ADR-001"
    assert fm.r9_skip is True


# --- Test 3: type 없으면 에러 ---


def test_frontmatter_parse_missing_type_raises() -> None:
    """OKF §4.1 required: `type` field 없거나 비면 InvalidFrontmatterError."""
    mod = _import_okf_export()
    # no type at all
    try:
        mod.Frontmatter.parse("---\nstatus: active\n---\nbody\n")
    except mod.InvalidFrontmatterError as e:
        assert "type" in str(e).lower()
    else:
        raise AssertionError("expected InvalidFrontmatterError for missing type")
    # empty type
    try:
        mod.Frontmatter.parse("---\ntype: \"  \"\n---\nbody\n")
    except mod.InvalidFrontmatterError as e:
        assert "type" in str(e).lower()
    else:
        raise AssertionError("expected InvalidFrontmatterError for empty type")


# --- Test 4: mapping field order (SPEC.md §4.1 priority) ---


def test_map_to_okf_field_order() -> None:
    """OKF §4.1 priority: type → title → description → resource → tags → timestamp.

    본 test 는 serialize 결과의 frontmatter lines 순서 검증.
    """
    mod = _import_okf_export()
    fm = mod.Frontmatter.parse(
        "---\n"
        "type: concept\n"
        "title: Test\n"
        "description: test desc\n"
        "last_ingested_from: https://example.com/x.md\n"
        "tags: [a, b]\n"
        "updated: 2026-06-16\n"
        "status: active\n"
        "related_pages: [foo]\n"
        "r9_skip: true\n"
        "---\n\nbody\n"
    )
    mapping = mod.map_frontmatter_to_okf(fm)
    keys_in_order: list[str] = []
    for line in mapping.frontmatter_lines:
        if line == "---" or ":" not in line:
            continue
        key = line.split(":", 1)[0].strip()
        if key:
            keys_in_order.append(key)
    # priority: type, title, description, resource, tags, timestamp, then extensions
    expected_priority_prefix = ["type", "title", "description", "resource", "tags", "timestamp"]
    actual_priority_prefix = [k for k in keys_in_order if k in expected_priority_prefix]
    assert actual_priority_prefix == expected_priority_prefix, (
        f"OKF §4.1 priority order violated: {actual_priority_prefix}"
    )


# --- Test 5: title/description body derivation ---


def test_map_to_okf_derives_title_from_body() -> None:
    """frontmatter 에 title/description 없을 때 body H1 + 첫 prose paragraph 에서 derive."""
    mod = _import_okf_export()
    fm = mod.Frontmatter.parse(
        "---\n"
        "type: pattern\n"
        "status: active\n"
        "---\n\n"
        "# R4 Anchor Index\n"
        "\n"
        "A knowledge index needs merge-safe structure. Free-form prose causes permanent conflicts.\n"
        "\n"
        "## When to Use\n"
        "\n"
        "- Master knowledge catalogs\n"
    )
    body = "# R4 Anchor Index\n\nA knowledge index needs merge-safe structure. Free-form prose causes permanent conflicts.\n\n## When to Use\n"
    mapping = mod.map_frontmatter_to_okf(fm, body=body)
    lines_str = "\n".join(mapping.frontmatter_lines)
    assert "title: R4 Anchor Index" in lines_str, f"title not derived: {lines_str!r}"
    assert "A knowledge index needs merge-safe structure" in lines_str, (
        f"description not derived: {lines_str!r}"
    )


# --- Test 6: body link rewrite (§5.1) ---


def test_rewrite_wiki_links() -> None:
    """[[path/to/page]] → [text](../path/to/page.md), [[path#anchor]] → [text](../path/to/page.md#anchor)."""
    mod = _import_okf_export()
    body_in = (
        "See [[concepts/foo]] for details. "
        "Also [[concepts/bar#section-2]]. "
        "And [[entities/simple-entity]] here."
    )
    body_out = mod.rewrite_wiki_links_to_okf(body_in)
    assert "[foo](../concepts/foo.md)" in body_out, f"foo link not rewritten: {body_out!r}"
    assert "[bar](../concepts/bar.md#section-2)" in body_out, f"anchor link not rewritten: {body_out!r}"
    assert "[simple-entity](../entities/simple-entity.md)" in body_out, (
        f"entity link not rewritten: {body_out!r}"
    )
    # round-trip: no [[...]] remaining
    assert "[[" not in body_out and "]]" not in body_out, f"wiki link not fully rewritten: {body_out!r}"


# --- Test 7: end-to-end export ---


def test_export_wiki_to_okf_end_to_end() -> None:
    """1 page export — OKF spec required (`type` non-empty) + body link rewrite 검증."""
    mod = _import_okf_export()
    with tempfile.TemporaryDirectory() as tmpdir:
        wiki_root = Path(tmpdir) / "wiki"
        out_bundle = Path(tmpdir) / "bundle"
        wiki_root.mkdir()
        # make a fake wiki page with a body wiki-link
        page = wiki_root / "concepts"
        page.mkdir()
        wiki_page = page / "test-concept.md"
        wiki_page.write_text(
            "---\n"
            "type: concept\n"
            "status: active\n"
            "related_pages: [concepts/other]\n"
            "---\n\n"
            "# Test Concept\n\nbody line one\nbody line two\n\nSee [[concepts/other]] for related.\n",
            encoding="utf-8",
        )
        # make sibling concept that the link refers to (OKF tolerates broken links, but for completeness)
        (page / "other.md").write_text(
            "---\ntype: concept\nstatus: active\n---\n\n# Other\n", encoding="utf-8"
        )
        report = mod.export_wiki_to_okf(wiki_root, out_bundle)
        assert report.pages_exported == 2, f"exported count: {report.pages_exported}, errors: {report.errors}"
        out_page = out_bundle / "concepts" / "test-concept.md"
        assert out_page.exists(), f"output not created: {out_page}"
        out_text = out_page.read_text(encoding="utf-8")
        # OKF spec required: `type` field present, non-empty
        assert re.search(r"^type: \S+", out_text, re.MULTILINE), f"`type` field missing: {out_text!r}"
        # wiki-link rewritten
        assert "[[concepts/other]]" not in out_text
        assert "[other](../concepts/other.md)" in out_text
        # related_pages emit
        assert "related_pages" in out_text
        # See Also section emitted (from related_pages)
        assert "## See Also" in out_text


# --- Test 8: OKF spec §4.1 full conformance ---


def test_okf_spec_4_1_full_conformance() -> None:
    """OKF SPEC.md §4.1 full conformance: 3 hard rule + 5 recommended field.

    Hard rule (§9 conformance):
    1. Every non-reserved `.md` file has parseable YAML frontmatter
    2. Every frontmatter has non-empty `type` field
    3. Reserved filenames (`index.md`, `log.md`) follow structure

    Recommended (priority order): type → title → description → resource → tags → timestamp

    본 test 는 export 된 bundle 의 모든 page 가 §4.1 conformance 충족 검증.
    """
    mod = _import_okf_export()
    with tempfile.TemporaryDirectory() as tmpdir:
        wiki_root = Path(tmpdir) / "wiki"
        out_bundle = Path(tmpdir) / "bundle"
        wiki_root.mkdir()
        # 3 page: 1 with full frontmatter, 1 minimal, 1 in different subdir
        (wiki_root / "concepts").mkdir()
        (wiki_root / "entities").mkdir()
        (wiki_root / "concepts" / "full.md").write_text(
            "---\n"
            "type: concept\n"
            "title: Full Page\n"
            "description: All fields populated.\n"
            "last_ingested_from: https://example.com/spec.md\n"
            "tags: [a, b]\n"
            "updated: 2026-06-16\n"
            "status: active\n"
            "---\n\n# Full Page\n\nbody\n",
            encoding="utf-8",
        )
        (wiki_root / "concepts" / "minimal.md").write_text(
            "---\ntype: concept\n---\n\n# Minimal\n\nbody\n",
            encoding="utf-8",
        )
        (wiki_root / "entities" / "entity.md").write_text(
            "---\ntype: entity\nstatus: active\n---\n\n# Entity\n\nbody\n",
            encoding="utf-8",
        )
        report = mod.export_wiki_to_okf(wiki_root, out_bundle)
        assert report.pages_exported == 3, f"exported count: {report.pages_exported}, errors: {report.errors}"
        assert not report.errors, f"export errors: {report.errors}"

        # verify every exported page meets §4.1 conformance
        for out_path in sorted(out_bundle.rglob("*.md")):
            # §6 index.md 는 reserved 이며 frontmatter 가 다름 (okf_version, generated_at, generator).
            # §4.1 hard rule 은 *non-reserved* page 에만 적용.
            if out_path.name == "index.md" and out_path.parent == out_bundle:
                # bundle-root index.md: OKF spec §6 + §11 형식 (okf_version field)
                text = out_path.read_text(encoding="utf-8")
                assert text.startswith("---\n"), f"index.md missing frontmatter: {out_path}"
                assert "okf_version" in text, f"index.md missing okf_version (SPEC §11): {out_path}"
                continue
            text = out_path.read_text(encoding="utf-8")
            # §4.1 hard rule 1: parseable YAML frontmatter (lines start with ---)
            assert text.startswith("---\n"), f"{out_path}: missing frontmatter"
            # §4.1 hard rule 2: non-empty `type` field
            type_match = re.search(r"^type: (\S.*)$", text, re.MULTILINE)
            assert type_match, f"{out_path}: missing `type` field"
            type_val = type_match.group(1).strip().strip('"').strip("'")
            assert type_val, f"{out_path}: empty `type` field"
            # §4.1 hard rule 3: reserved filename structure
            # (per-page index.md/log.md 는 우리 export 가 emit 안 함 — subdir 의 reserved 는 안 OK)
            assert out_path.name not in ("index.md", "log.md"), (
                f"{out_path}: reserved filename in concept export (subdir)"
            )

        # verify priority order (type → title → description → resource → tags → timestamp) for full page
        full_text = (out_bundle / "concepts" / "full.md").read_text(encoding="utf-8")
        # find positions of first occurrence
        priority = ["type", "title", "description", "resource", "tags", "timestamp"]
        positions: list[tuple[str, int]] = []
        for key in priority:
            m = re.search(rf"^{key}:", full_text, re.MULTILINE)
            if m:
                positions.append((key, m.start()))
        # sort by position
        positions.sort(key=lambda x: x[1])
        actual_order = [k for k, _ in positions]
        assert actual_order == priority, (
            f"OKF §4.1 priority order violated: {actual_order} != {priority}"
        )


# --- Test 9: bundle directory layout (reserved file isolation, subdir preservation) ---


def test_okf_bundle_directory_layout() -> None:
    """Bundle directory layout: reserved filename 격리 + subdirectory 보존.

    OKF SPEC.md §3.1: `index.md` 와 `log.md` 는 reserved. §2.2: directory hierarchy 보존.
    본 test 는 export 시:
    1. reserved filename (`index.md`, `log.md`) 이 wiki 의 reserved file 과 충돌 시 skip
    2. subdirectory hierarchy 보존
    3. 우리 wiki 의 SCHEMA.md / INGEST_GUIDE.md 같은 reserved file 도 skip
    """
    mod = _import_okf_export()
    with tempfile.TemporaryDirectory() as tmpdir:
        wiki_root = Path(tmpdir) / "wiki"
        out_bundle = Path(tmpdir) / "bundle"
        wiki_root.mkdir()
        # create SCHEMA.md (reserved) at wiki root + INGEST_GUIDE.md (reserved)
        (wiki_root / "SCHEMA.md").write_text("---\ntype: schema\n---\n\nschema\n", encoding="utf-8")
        (wiki_root / "INGEST_GUIDE.md").write_text("# Guide\n", encoding="utf-8")
        (wiki_root / "index.md").write_text("# Index\n", encoding="utf-8")
        (wiki_root / "log.md").write_text("# Log\n", encoding="utf-8")
        # nested subdirectories with pages
        (wiki_root / "concepts").mkdir()
        (wiki_root / "concepts" / "a.md").write_text(
            "---\ntype: concept\n---\n\n# A\n\nbody\n", encoding="utf-8"
        )
        (wiki_root / "concepts" / "sub").mkdir()
        (wiki_root / "concepts" / "sub" / "b.md").write_text(
            "---\ntype: concept\n---\n\n# B\n\nbody\n", encoding="utf-8"
        )
        (wiki_root / "decisions").mkdir()
        (wiki_root / "decisions" / "d.md").write_text(
            "---\ntype: decision\n---\n\n# D\n\nbody\n", encoding="utf-8"
        )
        report = mod.export_wiki_to_okf(wiki_root, out_bundle)
        # only 3 concept/decision pages exported, reserved + non-type-dir files skipped
        assert report.pages_exported == 3, (
            f"exported count: {report.pages_exported} (expected 3), errors: {report.errors}"
        )
        # reserved wiki-root files NOT in output
        assert not (out_bundle / "SCHEMA.md").exists(), "SCHEMA.md leaked to bundle"
        assert not (out_bundle / "INGEST_GUIDE.md").exists(), "INGEST_GUIDE.md leaked"
        # bundle-root index.md IS auto-emitted (OKF spec §6 + §11) — verify it has okf_version
        index_path = out_bundle / "index.md"
        assert index_path.exists(), "bundle-root index.md not auto-emitted"
        index_text = index_path.read_text(encoding="utf-8")
        assert "okf_version" in index_text, "bundle-root index.md missing okf_version"
        # log.md 는 export 가 emit 안 함 (별도 기능)
        assert not (out_bundle / "log.md").exists(), "log.md leaked (not emitted by our export)"
        # subdirectory hierarchy preserved
        assert (out_bundle / "concepts" / "a.md").exists(), "concepts/a.md missing"
        assert (out_bundle / "concepts" / "sub" / "b.md").exists(), "concepts/sub/b.md missing"
        assert (out_bundle / "decisions" / "d.md").exists(), "decisions/d.md missing"
        # exact file list: 3 concept/decision pages + 1 bundle-root index.md
        all_files = sorted(p.relative_to(out_bundle) for p in out_bundle.rglob("*") if p.is_file())
        expected = sorted(
            [
                Path("index.md"),
                Path("concepts/a.md"),
                Path("concepts/sub/b.md"),
                Path("decisions/d.md"),
            ]
        )
        assert all_files == expected, f"layout mismatch:\n  got: {all_files}\n  expected: {expected}"


# --- Test 10: bundle root index.md auto-emit (OKF SPEC §6 + §11) ---


def test_okf_bundle_root_index_md_emit() -> None:
    """Bundle root `index.md` 자동 emit (OKF SPEC.md §6 + §11).

    §6: index.md MAY appear in any directory to enumerate contents.
    §11: bundle-root `index.md` frontmatter 의 `okf_version` 으로 spec version 선언.
    """
    mod = _import_okf_export()
    with tempfile.TemporaryDirectory() as tmpdir:
        wiki_root = Path(tmpdir) / "wiki"
        out_bundle = Path(tmpdir) / "bundle"
        wiki_root.mkdir()
        (wiki_root / "concepts").mkdir()
        (wiki_root / "concepts" / "alpha.md").write_text(
            "---\ntype: concept\nstatus: active\n---\n\n# Alpha\n\nbody\n",
            encoding="utf-8",
        )
        (wiki_root / "decisions").mkdir()
        (wiki_root / "decisions" / "beta.md").write_text(
            "---\ntype: decision\nstatus: active\n---\n\n# Beta\n\nbody\n",
            encoding="utf-8",
        )
        report = mod.export_wiki_to_okf(wiki_root, out_bundle)
        assert report.pages_exported == 2, f"got {report.pages_exported}, errors: {report.errors}"
        index_path = out_bundle / "index.md"
        assert index_path.exists(), "bundle-root index.md not auto-emitted"
        text = index_path.read_text(encoding="utf-8")
        # §11: okf_version field
        assert 'okf_version: "0.1"' in text, f"index.md missing okf_version field:\n{text}"
        # generated_at + generator field
        assert "generated_at:" in text, "index.md missing generated_at"
        assert "generator:" in text, "index.md missing generator"
        # body: section heading per type + entries
        assert "## Concepts" in text, "index.md missing Concepts section"
        assert "## Decisions" in text, "index.md missing Decisions section"
        assert "alpha.md" in text, "index.md missing alpha.md entry"
        assert "beta.md" in text, "index.md missing beta.md entry"
        # bundle-root index.md 는 §4.1 hard rule 적용 안 됨 (no `type` field, has `okf_version`)
        assert "type:" not in text.split("---")[1], "index.md should NOT have `type` field (reserved)"


# --- Test 11: vcs_commit field → commit-pinned URL (ADR-018) ---


def test_vcs_commit_emits_pinned_url() -> None:
    """vcs_commit 명시 시 _derive_resource 가 commit-pinned URL emit (ADR-018)."""
    import importlib.util
    import sys
    # Patch via sys.modules BEFORE _derive_resource's lazy import
    pr_spec = importlib.util.spec_from_file_location(
        "workflow_kit.path_resolver", str(SOURCE_ROOT / "workflow_kit" / "path_resolver.py")
    )
    pr = importlib.util.module_from_spec(pr_spec)
    pr_spec.loader.exec_module(pr)
    sys.modules["workflow_kit.path_resolver"] = pr
    orig_url = pr.resolve_in_repo_path_to_url
    orig_pinned = pr.resolve_in_repo_path_to_url_pinned
    pr.resolve_in_repo_path_to_url = lambda path, root: "https://github.com/foo/bar/blob/main/" + path
    pr.resolve_in_repo_path_to_url_pinned = lambda path, root, commit_sha=None, ref=None: (
        f"https://github.com/foo/bar/blob/{commit_sha or ref}/" + path
    )
    try:
        mod = _import_okf_export()
        # in-repo path + vcs_commit → commit-pinned URL
        url = mod._derive_resource(
            "docs/spec.md", repo_root=Path("/fake"), vcs_commit="abc1234"
        )
        assert url == "https://github.com/foo/bar/blob/abc1234/docs/spec.md", f"got {url!r}"
        # in-repo path + vcs_ref → ref-pinned URL
        url = mod._derive_resource(
            "docs/spec.md", repo_root=Path("/fake"), vcs_ref="v0.7.37"
        )
        assert url == "https://github.com/foo/bar/blob/v0.7.37/docs/spec.md", f"got {url!r}"
        # URL form unchanged (vcs_commit ignored)
        url = mod._derive_resource("https://example.com/spec.md")
        assert url == "https://example.com/spec.md", f"got {url!r}"
    finally:
        pr.resolve_in_repo_path_to_url = orig_url
        pr.resolve_in_repo_path_to_url_pinned = orig_pinned



# --- 메인 실행 ---
def main() -> int:
    test_funcs = [
        test_frontmatter_parse_minimal,
        test_frontmatter_parse_full,
        test_frontmatter_parse_missing_type_raises,
        test_map_to_okf_field_order,
        test_map_to_okf_derives_title_from_body,
        test_rewrite_wiki_links,
        test_export_wiki_to_okf_end_to_end,
        test_okf_spec_4_1_full_conformance,
        test_okf_bundle_directory_layout,
        test_okf_bundle_root_index_md_emit,
        test_vcs_commit_emits_pinned_url,
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
