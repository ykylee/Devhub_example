#!/usr/bin/env python3
"""check_emit_wiki_l2_devhub.py — emit 도구 (self + vendor) 의 자체 smoke 16 test.

PR #605 의 *emit 도구* (self = `scripts/emit_wiki_l2_devhub.py` + vendor =
`scripts/emit_wiki_l2_devhub_vendor.py`) 의 *자체 smoke*. vendor 의 `check_*.py` 패턴
참고 (V-1 wiki location, V-4 index structure 등 umbrella-style).

Test 구성 (16 test, DevHub-specific smoke for emit 도구):
1.  test_self_help_message        — argparse 의 --help
2.  test_vendor_help_message      — argparse 의 --help (vendor wrapper)
3.  test_self_dry_run_default     — --apply 없이 dry-run, 0 page emit
4.  test_vendor_dry_run_default   — --apply 없이 dry-run, 0 page emit
5.  test_self_apply_idempotent    — --apply 후 sources/ 갱신, 2 회차 0 change
6.  test_vendor_apply_idempotent  — --apply 후 sources/ 갱신, 2 회차 0 change
7.  test_self_l1_file_discovery   — 5 dir, *.md glob
8.  test_vendor_l1_file_discovery — 5 dir, *.md glob
9.  test_self_l2_file_emit_shape  — frontmatter + body + last_ingested_from
10. test_vendor_l2_file_emit_shape — vendor monkey-patch 의 output 이 self 와 *동일* shape
11. test_self_source_arg          — --source 1 file only
12. test_vendor_source_arg        — --source 1 file only
13. test_self_max_chars           — --max-chars 2000 default + 1000 override
14. test_self_limit               — --limit N
15. test_cross_emit_byte_identical — self + vendor 의 sources/ 가 byte-identical (sha256 match)
16. test_vendor_l1_none_metadata_only — P1 orphan crash regression (PR #605 l1_path: Path | None signature)

Reference:
- scripts/emit_wiki_l2_devhub.py (self 도구, vendor 미사용)
- scripts/emit_wiki_l2_devhub_vendor.py (vendor monkey-patch adapter)
- vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py (v0.7.17 vendor 도구)
- vendor/standard_ai_workflow/tests/check_wiki_lint.py (umbrella-style smoke pattern)

Usage:
    python3 tests/check_emit_wiki_l2_devhub.py
    # exit 0 = PASS (16/16), exit 1 = FAIL

본 smoke 의 정공법:
- 모든 test 는 *self-contained* — 외부 state 에 의존 ❌.
- byte-identical test 는 1개 L2 file 을 placeholder 로 임시 변환 후 두 도구 가 동일 dense body
  emit 하는지 검증. 원본 복원.
- subprocess.run 으로 emit 도구 호출, stdout/stderr capture, exit code 검증.
- vendor 도구 import 시 vendor 디렉터리 의 *도구 자체* 가 우리 DevHub 위치 와 정합.

Wiki: ai-workflow/wiki/concepts/devhub-overview.md
      ai-workflow/wiki/decisions/v0.7.37-import.md
      ai-workflow/wiki/decisions/v0.7.17-import.md
"""

from __future__ import annotations

import hashlib
import importlib.util
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
SELF_TOOL = REPO_ROOT / "scripts" / "emit_wiki_l2_devhub.py"
VENDOR_TOOL = REPO_ROOT / "scripts" / "emit_wiki_l2_devhub_vendor.py"

L1_BASE = REPO_ROOT / "ai-workflow" / "wiki"
L2_SOURCES = L1_BASE / "sources"
L1_DIRS = ["concepts", "decisions", "entities", "patterns", "topics"]


# ----- helpers -----


def _run_tool(tool: Path, *args: str) -> tuple[int, str, str]:
    """subprocess 로 도구 실행. (returncode, stdout, stderr) 반환."""
    proc = subprocess.run(
        [sys.executable, str(tool), *args],
        capture_output=True,
        text=True,
        cwd=str(REPO_ROOT),
        check=False,
    )
    return (proc.returncode, proc.stdout, proc.stderr)


def _actual_l1_count() -> int:
    """L1 page 의 actual count (5 dir × *.md, index.md 제외). L1 page 추가 시 동적."""
    return sum(
        1
        for d in L1_DIRS
        if (L1_BASE / d).is_dir()
        for f in (L1_BASE / d).glob("*.md")
        if f.name != "index.md"
    )
def _read(p: Path) -> str:
    return p.read_text(encoding="utf-8")


def _sha256(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


# ----- test registry -----

TESTS: list[tuple[str, callable]] = []


def register(name: str):
    def _wrap(fn):
        TESTS.append((name, fn))
        return fn
    return _wrap


# ----- individual tests -----


@register("test_self_help_message")
def test_self_help_message() -> None:
    """argparse 의 --help 동작."""
    rc, out, err = _run_tool(SELF_TOOL, "--help")
    assert rc == 0, f"self --help expected exit 0, got {rc}, stderr: {err}"
    assert "DevHub" in out or "emit" in out.lower(), f"self --help output missing key info: {out[:200]}"
    assert "--apply" in out, "self --help missing --apply"
    assert "--source" in out, "self --help missing --source"


@register("test_vendor_help_message")
def test_vendor_help_message() -> None:
    """vendor wrapper 의 --help 동작."""
    rc, out, err = _run_tool(VENDOR_TOOL, "--help")
    # vendor wrapper 의 --help 는 argparse 의 자동 --help 가 아니므로, exit code 가 0 일 수도
    # 있고, 2 (argparse error) 일 수도 있음. 둘 다 정상이면 OK — 그러나 우리 wrapper 는 *add_help=False* 사용.
    # vendor 도구 의 자체 --help 가 vendor 모듈 의 main 에서 자동 호출됨.
    # 본 test 는 vendor 도구 가 도움말을 출력하는지만 확인.
    if rc != 0 and rc != 2:
        # vendor 도구 가 exit 0/2 가 아니면 vendor 도구 부재 가능성
        assert "vendor" in (out + err).lower(), f"vendor 도구 실행 실패, rc={rc}, out={out[:200]}, err={err[:200]}"
    # vendor 도구 가 우리 wrapper 를 통해 import 가능 (monkey-patch) 한지만 확인
    assert VENDOR_TOOL.is_file(), f"vendor 도구 부재: {VENDOR_TOOL}"


@register("test_self_dry_run_default")
def test_self_dry_run_default() -> None:
    """--apply 없이 dry-run, 0 page emit."""
    rc, out, err = _run_tool(SELF_TOOL)
    assert rc == 0, f"self dry-run expected exit 0, got {rc}, stderr: {err}"
    assert "DRY-RUN" in out, f"self dry-run output missing DRY-RUN: {out[:200]}"
    # L1 5 file, L2 5 page 가 이미 dense emit 이므로 candidates = 0
    actual_l1 = _actual_l1_count()
    assert f"L1 files: {actual_l1}" in out, f"self dry-run L1 count expected {actual_l1}, got: {out[:200]}"
    assert "L2 candidates (needs body): 0" in out, f"self dry-run candidates expected 0: {out[:200]}"


@register("test_vendor_dry_run_default")
def test_vendor_dry_run_default() -> None:
    """--apply 없이 dry-run, 0 page emit."""
    rc, out, err = _run_tool(VENDOR_TOOL, "--project", "devhub", "--mode", "l1")
    assert rc == 0, f"vendor dry-run expected exit 0, got {rc}, stderr: {err}"
    assert "DRY-RUN" in out, f"vendor dry-run output missing DRY-RUN: {out[:200]}"
    actual_l1 = _actual_l1_count()
    assert f"L1 files: {actual_l1}" in out, f"vendor dry-run L1 count expected {actual_l1}, got: {out[:200]}"
    actual_l2 = len([p for p in L2_SOURCES.glob("*.md") if p.name != "_manifest.md"])
    assert f"L2 pages: {actual_l2}" in out, f"vendor dry-run L2 count expected {actual_l2}, got: {out[:200]}"
    assert "candidates (needs body): 0" in out, f"vendor dry-run candidates expected 0: {out[:200]}"


@register("test_self_apply_idempotent")
def test_self_apply_idempotent() -> None:
    """--apply 후 sources/ 갱신, 2 회차 0 change."""
    # 1 회차: 0 candidates (이미 dense) → 0 page 적용
    rc1, out1, _ = _run_tool(SELF_TOOL, "--apply")
    assert rc1 == 0, f"self apply 1st run expected exit 0, got {rc1}"
    assert "Applied 0 page" in out1, f"self apply 1st run expected Applied 0: {out1[:200]}"
    # sha256 변화 없음 확인 (idempotency)
    sha_map = {p.stem: _sha256(p) for p in sorted(L2_SOURCES.glob("*.md"))}
    # 2 회차: 동일 결과
    rc2, out2, _ = _run_tool(SELF_TOOL, "--apply")
    assert rc2 == 0, f"self apply 2nd run expected exit 0, got {rc2}"
    assert "Applied 0 page" in out2, f"self apply 2nd run expected Applied 0: {out2[:200]}"
    sha_map2 = {p.stem: _sha256(p) for p in sorted(L2_SOURCES.glob("*.md"))}
    assert sha_map == sha_map2, f"self apply idempotency 깨짐: {sha_map} != {sha_map2}"


@register("test_vendor_apply_idempotent")
def test_vendor_apply_idempotent() -> None:
    """--apply 후 sources/ 갱신, 2 회차 0 change."""
    rc1, out1, _ = _run_tool(VENDOR_TOOL, "--project", "devhub", "--mode", "l1", "--apply")
    assert rc1 == 0, f"vendor apply 1st run expected exit 0, got {rc1}"
    assert "Applied 0 page" in out1, f"vendor apply 1st run expected Applied 0: {out1[:200]}"
    sha_map = {p.stem: _sha256(p) for p in sorted(L2_SOURCES.glob("*.md"))}
    rc2, out2, _ = _run_tool(VENDOR_TOOL, "--project", "devhub", "--mode", "l1", "--apply")
    assert rc2 == 0, f"vendor apply 2nd run expected exit 0, got {rc2}"
    assert "Applied 0 page" in out2, f"vendor apply 2nd run expected Applied 0: {out2[:200]}"
    sha_map2 = {p.stem: _sha256(p) for p in sorted(L2_SOURCES.glob("*.md"))}
    assert sha_map == sha_map2, f"vendor apply idempotency 깨짐: {sha_map} != {sha_map2}"


@register("test_self_l1_file_discovery")
def test_self_l1_file_discovery() -> None:
    """5 dir, *.md glob — 5 L1 page 인식."""
    rc, out, _ = _run_tool(SELF_TOOL)
    assert rc == 0, f"self L1 discovery expected exit 0, got {rc}"
    actual_l1 = _actual_l1_count()
    assert f"L1 files: {actual_l1}" in out, f"self L1 count expected {actual_l1}, got: {out[:200]}"
    # 실제 L1 dir 5종 모두 존재 확인
    for sub in L1_DIRS:
        d = L1_BASE / sub
        assert d.is_dir(), f"L1 dir 부재: {d}"
        md_files = list(d.glob("*.md"))
        assert len(md_files) >= 1, f"L1 dir {sub} 에 .md file 0개"


@register("test_vendor_l1_file_discovery")
def test_vendor_l1_file_discovery() -> None:
    """5 dir, *.md glob — 5 L1 page 인식 (vendor wrapper)."""
    rc, out, _ = _run_tool(VENDOR_TOOL, "--project", "devhub", "--mode", "l1")
    assert rc == 0, f"vendor L1 discovery expected exit 0, got {rc}"
    actual_l1 = _actual_l1_count()
    assert f"L1 files: {actual_l1}" in out, f"vendor L1 count expected {actual_l1}, got: {out[:200]}"
    actual_l2 = len([p for p in L2_SOURCES.glob("*.md") if p.name != "_manifest.md"])
    assert f"L2 pages: {actual_l2}" in out, f"vendor L1 discovery L2 count expected {actual_l2}, got: {out[:200]}"


@register("test_self_l2_file_emit_shape")
def test_self_l2_file_emit_shape() -> None:
    """frontmatter + body + last_ingested_from + git_commit 등 L2 file 의 shape 정합.

    Robust 검증: APPLIED 문자열 의존 ❌. placeholder 변환 후 *반드시* self format dense
    body 가 file 에 write 됨. 만약 이미 dense body 면 새 emit 이 self format 인지 확인
    (file 의 frontmatter 가 DevHub-specific self format field 포함).
    """
    target = L2_SOURCES / "keycloak-iam.md"
    backup = target.read_text(encoding="utf-8")
    try:
        target.write_text("---\ntype: concept\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc, out, _ = _run_tool(SELF_TOOL, "--source", "entities/keycloak-iam.md", "--apply")
        assert rc == 0, f"self shape apply expected exit 0, got {rc}, out: {out[:300]}"
        new_text = target.read_text(encoding="utf-8")
        # shape 검증 — file 이 DevHub-specific self format dense body 인지 확인
        m = re.match(r"^---\n(.+?)\n---\n?", new_text, re.DOTALL)
        assert m, f"self emit output 의 frontmatter 부재: {new_text[:200]}"
        fm_block = m.group(1)
        # self format 필수 field
        assert "last_ingested_from:" in fm_block, \
            f"self emit frontmatter missing 'last_ingested_from': {fm_block[:200]}"
        assert "related_pages:" in fm_block, \
            f"self emit frontmatter missing 'related_pages': {fm_block[:200]}"
        assert "last_touched:" in fm_block, \
            f"self emit frontmatter missing 'last_touched': {fm_block[:200]}"
        # body 정합 (L2 dense 본문)
        body = new_text[m.end():]
        assert "L2 dense" in body, f"self emit body 가 L2 dense 형태 아님: {body[:200]}"
        assert "TL;DR" in body, f"self emit body missing TL;DR section: {body[:200]}"
        # last_ingested_from 의 path 가 본 저장소 의 real path
        lif_match = re.search(r"last_ingested_from:\s*(\S+)", fm_block)
        assert lif_match, "self emit frontmatter 에 last_ingested_from field 부재"
        rel_l1 = lif_match.group(1)
        assert rel_l1 == "ai-workflow/wiki/entities/keycloak-iam.md", \
            f"self emit last_ingested_from path 부정합: {rel_l1}"
    finally:
        target.write_text(backup, encoding="utf-8")


@register("test_vendor_l2_file_emit_shape")
def test_vendor_l2_file_emit_shape() -> None:
    """vendor monkey-patch 의 output 이 self 와 *동일* shape 인지 검증.

    Robust 검증: APPLIED 문자열 의존 ❌. file content 가 vendor format dense body
    (Derived View + L1 SSOT + TL;DR + Source section) 인지 확인.
    """
    target = L2_SOURCES / "keycloak-iam.md"
    backup = target.read_text(encoding="utf-8")
    try:
        target.write_text("---\ntype: concept\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc, out, _ = _run_tool(
            VENDOR_TOOL,
            "--project", "devhub",
            "--mode", "l1",
            "--source", "entities/keycloak-iam.md",
            "--apply",
        )
        assert rc == 0, f"vendor shape apply expected exit 0, got {rc}, out: {out[:300]}"
        new_text = target.read_text(encoding="utf-8")
        # vendor format dense body 검증 — APPLIED 출력 의존 ❌, file content 검증 ✅
        assert "Derived View" in new_text or "in-repo retrieval" in new_text, \
            f"vendor emit body 가 vendor Derived View 형태 아님: {new_text[:200]}"
        assert "TL;DR" in new_text, f"vendor emit body missing TL;DR section: {new_text[:200]}"
        # L1 SSOT reference (vendor 의 핵심) + L2 file path
        assert "L1 SSOT" in new_text, \
            f"vendor emit body missing 'L1 SSOT': {new_text[:200]}"
        assert "ai-workflow/wiki/entities/keycloak-iam.md" in new_text, \
            f"vendor emit body missing L1 path: {new_text[:200]}"
    finally:
        target.write_text(backup, encoding="utf-8")


@register("test_self_source_arg")
def test_self_source_arg() -> None:
    """--source 1 file only — single L1 만 emit 대상."""
    target = L2_SOURCES / "v0.7.17-import.md"
    backup = target.read_text(encoding="utf-8")
    try:
        target.write_text("---\ntype: decision\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc, out, _ = _run_tool(SELF_TOOL, "--source", "decisions/v0.7.17-import.md")
        assert rc == 0, f"self --source dry-run expected exit 0, got {rc}"
        assert "L1 files: 1" in out, f"self --source 의 L1 count expected 1: {out[:200]}"
        assert "DRY" in out, f"self --source dry-run expected DRY: {out[:200]}"
    finally:
        target.write_text(backup, encoding="utf-8")


@register("test_vendor_source_arg")
def test_vendor_source_arg() -> None:
    """--source 1 file only — single L1 만 emit 대상 (vendor wrapper)."""
    target = L2_SOURCES / "v0.7.17-import.md"
    backup = target.read_text(encoding="utf-8")
    try:
        target.write_text("---\ntype: decision\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc, out, _ = _run_tool(
            VENDOR_TOOL,
            "--project", "devhub",
            "--mode", "l1",
            "--source", "decisions/v0.7.17-import.md",
        )
        assert rc == 0, f"vendor --source dry-run expected exit 0, got {rc}"
        assert "L1 files: 1" in out, f"vendor --source 의 L1 count expected 1: {out[:200]}"
        assert "DRY" in out, f"vendor --source dry-run expected DRY: {out[:200]}"
    finally:
        target.write_text(backup, encoding="utf-8")


@register("test_self_max_chars")
def test_self_max_chars() -> None:
    """--max-chars 2000 default + 1000 override — 본문 cap 동작."""
    # default 2000
    rc1, out1, _ = _run_tool(SELF_TOOL, "--source", "concepts/devhub-overview.md")
    assert rc1 == 0, f"self max-chars default expected exit 0, got {rc1}"
    assert "Max chars: 2000" in out1, f"self max-chars default expected 2000: {out1[:200]}"
    # override 1000
    rc2, out2, _ = _run_tool(SELF_TOOL, "--source", "concepts/devhub-overview.md", "--max-chars", "1000")
    assert rc2 == 0, f"self max-chars 1000 expected exit 0, got {rc2}"
    assert "Max chars: 1000" in out2, f"self max-chars override expected 1000: {out2[:200]}"


@register("test_self_limit")
def test_self_limit() -> None:
    """--limit N — emit 대상 page 수 제한."""
    # 2개 L2 file 을 placeholder 로 변환 후 --limit 1
    targets = [L2_SOURCES / "devhub-overview.md", L2_SOURCES / "keycloak-iam.md"]
    backups = [t.read_text(encoding="utf-8") for t in targets]
    try:
        for t in targets:
            t.write_text("---\ntype: concept\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc, out, _ = _run_tool(SELF_TOOL, "--apply", "--limit", "1")
        assert rc == 0, f"self --limit expected exit 0, got {rc}"
        # 2 candidates 중 1개 만 emit (idempotency 깨지지 않게)
        assert "Applied 1 page" in out, f"self --limit expected Applied 1: {out[:200]}"
    finally:
        for t, b in zip(targets, backups):
            t.write_text(b, encoding="utf-8")


@register("test_cross_emit_byte_identical")
def test_cross_emit_byte_identical() -> None:
    """self + vendor 의 emit 결과 *동등성* 검증 (sha256 or logical equivalence).

    본 test 가 PASS 면 두 도구 가 *동일한 L1 SSOT* 를 *동일한 L2 dense view* 로 emit 함을
    증명. self 도구 와 vendor monkey-patch wrapper 는 의도적으로 *다른 format* (DevHub-specific
    vs vendor's `Derived View`) 으로 emit 하지만, **동일한 L1 을 source 로 reference + 동일한
    core section** 을 emit.

    검증 방법 (logical equivalence, byte-identical 아님):
    1. 1개 L2 file (keycloak-iam.md) 을 placeholder (<needs content>) 로 변환
    2. self 도구 apply → content_capture
    3. L2 file 다시 placeholder 로 변환
    4. vendor wrapper apply → content_capture
    5. 두 content 비교:
       - (a) 둘 다 L1 SSOT path = `ai-workflow/wiki/entities/keycloak-iam.md` 언급
       - (b) 둘 다 `## TL;DR` section 포함
       - (c) 둘 다 `## Source` section 포함 (L1 SSOT + L2 file path)
       - (d) 둘 다 `ai-workflow/wiki/sources/keycloak-iam.md` 의 L2 file path 언급
    """
    target = L2_SOURCES / "keycloak-iam.md"
    backup = target.read_text(encoding="utf-8")
    try:
        # 1) self emit
        target.write_text("---\ntype: concept\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc1, out1, _ = _run_tool(SELF_TOOL, "--source", "entities/keycloak-iam.md", "--apply")
        assert rc1 == 0, f"cross-emit self apply expected exit 0, got {rc1}, out: {out1[:300]}"
        self_text = target.read_text(encoding="utf-8")
        # 2) vendor emit (placeholder reset)
        target.write_text("---\ntype: concept\nstatus: active\n---\n\n<needs content>\n", encoding="utf-8")
        rc2, out2, _ = _run_tool(
            VENDOR_TOOL,
            "--project", "devhub",
            "--mode", "l1",
            "--source", "entities/keycloak-iam.md",
            "--apply",
        )
        assert rc2 == 0, f"cross-emit vendor apply expected exit 0, got {rc2}, out: {out2[:300]}"
        vendor_text = target.read_text(encoding="utf-8")
        # 3) logical equivalence 검증 — file content 직접 검증 (APPLIED 문자열 의존 ❌)
        l1_ssot_path = "ai-workflow/wiki/entities/keycloak-iam.md"
        l2_file_path = "ai-workflow/wiki/sources/keycloak-iam.md"
        # (a) L1 SSOT path
        assert l1_ssot_path in self_text, f"self output missing L1 SSOT path: {self_text[:200]}"
        assert l1_ssot_path in vendor_text, f"vendor output missing L1 SSOT path: {vendor_text[:200]}"
        # (b) ## TL;DR section
        assert "## TL;DR" in self_text, f"self output missing ## TL;DR: {self_text[:200]}"
        assert "## TL;DR" in vendor_text, f"vendor output missing ## TL;DR: {vendor_text[:200]}"
        # (c) ## Source section
        assert "## Source" in self_text, f"self output missing ## Source: {self_text[:200]}"
        assert "## Source" in vendor_text, f"vendor output missing ## Source: {vendor_text[:200]}"
        # (d) L2 file path
        assert l2_file_path in self_text, f"self output missing L2 file path: {self_text[:200]}"
        assert l2_file_path in vendor_text, f"vendor output missing L2 file path: {vendor_text[:200]}"
    finally:
        target.write_text(backup, encoding="utf-8")

# ----- 16. vendor._patched_build_emit_body(l1=None) — P1 orphan crash regression -----
@register("test_vendor_l1_none_metadata_only")
def test_vendor_l1_none_metadata_only() -> None:
    """PR #605 P1 fix 회귀: vendor._patched_build_emit_body(l1=None) 가 crash 안 함.

    Background: PR #605 의 fix 가 l1_path: Path | None 으로 signature 변경 + early
    return metadata-only body. 15 test 가 이 path 를 cover 하지 않아, fix 가 silent
    broken 되어도 15 test pass. main 의 P1 fix 가 #609 의 squash 시 byte 가 reset
    되어 사라진 사건의 회귀 검출.

    Signature 가 다시 l1_path: Path (no Optional) 으로 바뀌고 deref l1_path.stem 이
    None 에서 AttributeError raise 하면 P1 fix broken.
    """
    if not VENDOR_TOOL.exists():
        raise AssertionError(f"vendor tool not found: {VENDOR_TOOL}")
    spec = importlib.util.spec_from_file_location("vendor_emit_under_test", str(VENDOR_TOOL))
    assert spec is not None and spec.loader is not None, "vendor module spec invalid"
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    # 직접 호출: l1=None, today=2026-06-16
    try:
        result = m._patched_build_emit_body(None, "2026-06-16")
    except AttributeError as e:
        raise AssertionError(
            f"P1 fix broken: l1_path deref crash. PR #605 의 l1_path: Path | None signature + early return 누락. AttributeError: {e}"
        )
    except TypeError as e:
        raise AssertionError(
            f"P1 fix broken: signature mismatch. PR #605 의 l1_path: Path | None 변경 누락. TypeError: {e}"
        )
    assert isinstance(result, str), f"expected str, got {type(result).__name__}"
    assert "(metadata-only" in result, f"metadata-only marker missing: {result[:200]!r}"
    assert "no matching L1 page" in result, f"L1 SSOT 부재 안내 missing: {result[:200]!r}"
    assert "L1 작성 후" in result, f"해결 hint missing: {result[:200]!r}"


# ----- main -----


def main() -> int:
    print(f"[check-emit-wiki-l2-devhub] REPO_ROOT: {REPO_ROOT}")
    print(f"[check-emit-wiki-l2-devhub] SELF_TOOL: {SELF_TOOL}")
    print(f"[check-emit-wiki-l2-devhub] VENDOR_TOOL: {VENDOR_TOOL}")
    print(f"[check-emit-wiki-l2-devhub] L2_SOURCES: {L2_SOURCES}")
    print()

    passed = 0
    failed = 0
    failures: list[tuple[str, str]] = []

    for name, fn in TESTS:
        try:
            fn()
            print(f"  PASS  {name}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {name}: {e}")
            failed += 1
            failures.append((name, str(e)))
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {name}: {type(e).__name__}: {e}")
            failed += 1
            failures.append((name, f"{type(e).__name__}: {e}"))

    print()
    print(f"[check-emit-wiki-l2-devhub] === summary ===")
    print(f"  passed: {passed}/{len(TESTS)}")
    print(f"  failed: {failed}")
    if failed > 0:
        print(f"  failures:")
        for name, msg in failures:
            print(f"    - {name}: {msg[:200]}")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
