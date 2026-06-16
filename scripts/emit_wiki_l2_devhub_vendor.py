#!/usr/bin/env python3
"""emit_wiki_l2_devhub_vendor.py — vendor 의 emit_wiki_l2_body.py 의 DevHub adapter (full monkey-patch).

vendor 의 `emit_wiki_l2_body.py` 의 *mini structure* (RAW_MIRROR / project / ai-workflow /
wiki) 가 우리 DevHub 의 *운영 위치* (ai-workflow/wiki/) 와 정합 안 함. 본 도구는 vendor 의
*도구 자체* 를 import 후 RAW_MIRROR / L1_BASE / L2_SOURCES / path_to_stem / find_l1_files /
find_l2_pages / build_emit_body / main 함수를 *monkey-patch* 하여 vendor 의 L2 emit 자동화를
우리 DevHub 의 in-repo 위치에서 동작.

vendor 도구 의 monkey-patch layer (DevHub-specific, 150+ line):
- RAW_MIRROR = 본 저장소 의 *RAW_MIRROR* (ai-workflow/wiki) — vendor 의 mini structure 의
  *inner path* (RAW_MIRROR / project / ai-workflow / wiki) 와 정합되도록 module-level 변수를
  *우리 운영 위치* 로 강제.
- L1_BASE = 동일.
- L2_SOURCES = 우리 L2 운영 위치 (ai-workflow/wiki/sources) 로 강제.
- path_to_stem = L1 file 의 *basename* (vendor 의 multi-segment stem 무시) — 우리 L2 의
  basename = L1 의 basename (PR #602 의 A안 5 page 정합).
- find_l1_files = 우리 운영 위치 의 L1 dir 5종 (concepts/decisions/entities/patterns/topics)
  만 인식 (vendor 의 RAW_MIRROR.rglob("*") 무시).
- find_l2_pages = 동일.
- build_emit_body = vendor 의 *build_emit_body* 함수 그대로 호출 — 우리 L1 file 의 *closure
  capture* 의 RAW_MIRROR 가 우리 운영 위치이므로 *relative_to* 가 정합.
- main = vendor 의 main 의 *candidates 수집 + emit 루프* 전체 monkey-patch — 우리 RAW_MIRROR 의
  child (= concepts/decisions/entities/patterns/topics) 인식.

본 wrapper 의 정공법 (DevHub-specific):
- 본 저장소 의 L1 page 의 *relative path* 가 vendor 의 mini structure 와 정합.
- L2 emit 자동화: 우리 L1 5 page 의 *in-place* L2 dense (sources/) emit.
- vendor 의 mini structure 의 *가장 안쪽* 가 우리 L1 의 *parent directory* 와 정합.

Usage:
    # Dry-run: vendor 의 도구 의 DevHub adapter 결과
    python3 scripts/emit_wiki_l2_devhub_vendor.py --project devhub --mode l1

    # Apply: 실제 L2 emit (DevHub 의 in-repo 위치)
    python3 scripts/emit_wiki_l2_devhub_vendor.py --project devhub --mode l1 --apply

    # 단일 file L1 만 emit
    python3 scripts/emit_wiki_l2_devhub_vendor.py --project devhub --mode l1 --source concepts/devhub-overview.md --apply

Reference:
- vendor/standard_ai_workflow/tools/emit_wiki_l2_body.py (v0.7.17 vendor 도구, mini structure)
- scripts/emit_wiki_l2_devhub.py (자체 도구, vendor 미사용)
- tests/check_wiki_drift_devhub.py (DevHub adapter 패턴, monkey-patch)
"""

from __future__ import annotations

import importlib.util
import re
import sys
from datetime import date
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[1]
L1_BASE_REAL = REPO_ROOT / "ai-workflow" / "wiki"
L2_SOURCES_REAL = L1_BASE_REAL / "sources"

# 우리 DevHub 의 L1 dir 5종 (in-repo 운영 위치)
L1_DIRS = ["concepts", "decisions", "entities", "patterns", "topics"]
L1_DIR_TO_TYPE = {
    "concepts": "concept",
    "decisions": "decision",
    "entities": "entity",
    "patterns": "pattern",
    "topics": "topic",
}
# local alias (devhub.py 와 동일 — monkey-patch isolation 위해 inline)
_l1_dir_to_type = L1_DIR_TO_TYPE

VENDOR_TOOL = REPO_ROOT / "vendor" / "standard_ai_workflow" / "tools" / "emit_wiki_l2_body.py"


# ----- monkey-patch helpers -----

def _patched_path_to_stem(rel_path: str) -> str:
    """L1 file 의 basename (without .md) = stem (vendor 의 multi-segment stem 무시).

    e.g. concepts/devhub-overview.md → devhub-overview (basename)
    """
    return Path(rel_path).stem


def _patched_find_l1_files(project: str) -> list[Path]:
    """DevHub 의 L1 dir 5종 의 *.md file 인식 (vendor 의 RAW_MIRROR.rglob 무시)."""
    files: list[Path] = []
    for sub in L1_DIRS:
        d = L1_BASE_REAL / sub
        if d.is_dir():
            files.extend(sorted(p for p in d.glob("*.md") if p.name != ".gitkeep"))
    return files


def _patched_find_l2_pages(project: str) -> dict[str, Path]:
    """DevHub 의 L2 sources/ 의 *.md 인식 (vendor 의 L2_SOURCES.glob 무시)."""
    if not L2_SOURCES_REAL.is_dir():
        return {}
    return {p.stem: p for p in sorted(L2_SOURCES_REAL.glob("*.md")) if p.name != ".gitkeep"}


def _patched_needs_body(l2_path: Path) -> bool:
    """L2 page 의 본문이 placeholder 또는 짧은지 확인."""
    if not l2_path.is_file():
        return True
    text = l2_path.read_text(encoding="utf-8")
    if "<needs content>" in text:
        return True
    m = re.match(r"^---\n(.+?)\n---\n?", text, re.DOTALL)
    body = text[m.end():] if m else text
    return len(body.strip()) < 200


def _patched_extract_tldr_from_l1(l1_path: Path) -> str:
    """vendor 의 extract_tldr_from_l1 그대로 — 우리 L1 file 의 *real path* 인식."""
    # vendor 의 함수 가 우리 RAW_MIRROR (= L1_BASE_REAL) 의 child 인식.
    # L1 file 의 *closure* 의 RAW_MIRROR (우리 patch 후 = L1_BASE_REAL) 의 relative_to 가 정합.
    content = l1_path.read_text(encoding="utf-8")
    m = re.search(r"^## (?:§\d+\s+)?TL;DR\s*\n+(.*?)(?=^##|\Z)", content, re.MULTILINE | re.DOTALL)
    if not m:
        return ""
    tldr_block = m.group(1).strip()
    lines = tldr_block.splitlines()
    table_lines = [l for l in lines if l.strip().startswith("|")]
    if len(table_lines) >= 4:
        return "\n".join(table_lines[:4])
    return "\n".join(lines[:5])


def _patched_extract_l1_body(l1_path: Path, max_chars: int = 2000) -> str:
    """vendor 의 extract_l1_body 와 동일 — L1 의 본문 일부 추출 (max_chars cap)."""
    content = l1_path.read_text(encoding="utf-8")
    m = re.match(r"^---\n(.+?)\n---\n?", content, re.DOTALL)
    body = content[m.end():] if m else content
    h1_match = re.search(r"^#\s+(.+)$", body, re.MULTILINE)
    if h1_match:
        body = body[h1_match.end():].lstrip("\n")
    if len(body) <= max_chars:
        return body
    return body[:max_chars] + "\n\n…(truncated)"


def _patched_build_emit_body(l1_path: Path | None, today: str, max_chars: int = 2000, mode: str = "l1") -> str:
    """vendor 의 build_emit_body 의 closure* (RAW_MIRROR) 가 우리 DevHub 의 L1_BASE_REAL 이 되도록
    *RAW_MIRROR 의 closure capture* 의 우리 wrapper 가 monkey-patch 후의 *RAW_MIRROR* (= 우리
    L1_BASE_REAL) 의 relative_to 가 정합.

    l1_path 가 None 인 경우 (metadata-only / orphan L2) 는 body 만 metadata-only 양식으로
    반환 — frontmatter 는 호출자 (_patched_main apply loop) 가 별도 preserve.
    """
    if l1_path is None:
        # metadata-only body: L1 SSOT 부재 / orphan. 본문 placeholder 만.
        return (
            f"# (metadata-only, {today})\n"
            "\n"
            "> **L1 SSOT**: (no matching L1 page — metadata-only candidate)\n"
            "> 본 L2 는 L1 SSOT 부재 / orphan. L1 작성 후 `--apply --mode l1` 로 본문 emit.\n"
            "\n"
            "## TL;DR\n"
            "\n"
            "(no TL;DR — L1 SSOT 부재)\n"
            "\n"
            "## Source\n"
            "\n"
            "- 본 wrapper: `scripts/emit_wiki_l2_devhub_vendor.py`\n"
        )
    title = l1_path.stem.replace("-", " ").title()
    l1_line_count = sum(1 for _ in l1_path.open(encoding="utf-8"))
    # vendor 의 build_emit_body 의 line 154 의 RAW_MIRROR.parts.index("raw") + 2 — 우리
    # L1_BASE_REAL (= ai-workflow/wiki) 에 'raw' segment 없음. 직접 *relative_to* 의 우리
    # L1_BASE_REAL 로 계산.
    try:
        rel_l1 = l1_path.relative_to(L1_BASE_REAL)
    except ValueError:
        rel_l1 = Path(l1_path.name)
    tldr = _patched_extract_tldr_from_l1(l1_path)
    body = _patched_extract_l1_body(l1_path, max_chars=max_chars)

    parts = [
        f"# {title} (Derived View, {today})",
        "",
        f"> **L1 SSOT**: `ai-workflow/wiki/{rel_l1}` ({l1_line_count} lines)",
        "> 본 L2 derived view 는 in-repo retrieval 용 압축 요약. dense content 는 L1 SSOT 참조.",
        "",
        "## TL;DR",
        "",
        tldr or "(no TL;DR section in L1)",
        "",
        "## L1 body (truncated)",
        "",
        body,
        "",
        "## Source",
        "",
        f"- 본 L2 file: `ai-workflow/wiki/sources/{l1_path.stem}.md`",
        f"- L1 SSOT: `ai-workflow/wiki/{rel_l1}`",
        "- L0 Home: `ai-workflow/wiki/index.md`",
        "- 본 wrapper: `scripts/emit_wiki_l2_devhub_vendor.py` (vendor 도구의 DevHub adapter)",
    ]
    return "\n".join(parts)


def _patched_main(orig_main: Any, module: Any) -> int:
    """vendor 의 main 함수의 *candidates 수집 + emit 루프* monkey-patch — 우리 DevHub 의
    L1_BASE_REAL 의 child 만 인식 + 우리 L2_SOURCES_REAL 에 emit.
    """
    # argparse 흉내: vendor 의 main 의 argparse 를 *우리가 직접* 사용 (인자 동일)
    import argparse
    parser = argparse.ArgumentParser(add_help=False)
    parser.add_argument("--project", type=str, default="devhub")
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--max-chars", type=int, default=2000)
    parser.add_argument("--limit", type=int, default=0)
    parser.add_argument("--mode", type=str, choices=["l1", "metadata-only", "all"], default="all")
    # vendor 의 argparse 와 별도로 우리 wrapper 의 추가 인자 (--source) 받음
    parser.add_argument("--source", type=str, default="")
    parser.add_argument("--dry-run", action="store_true", help="dry-run (default = !--apply)")
    args, _unknown = parser.parse_known_args()

    print(f"[emit-wiki-l2-devhub-vendor] vendor 도구: {VENDOR_TOOL}")
    print(f"[emit-wiki-l2-devhub-vendor] REPO_ROOT: {REPO_ROOT}")
    print(f"[emit-wiki-l2-devhub-vendor] L1_BASE_REAL (RAW_MIRROR patched): {L1_BASE_REAL}")
    print(f"[emit-wiki-l2-devhub-vendor] L2_SOURCES_REAL: {L2_SOURCES_REAL}")
    print(f"[emit-wiki-l2-devhub-vendor] project: {args.project} (in-repo only)")
    print()

    today = date.today().isoformat()
    dry_run = not args.apply

    # 우리 L1 + L2 인식
    l1_files = _patched_find_l1_files(args.project)
    if args.source:
        l1_target = L1_BASE_REAL / args.source
        if not l1_target.is_file():
            print(f"[emit-wiki-l2-devhub-vendor] error: --source not found: {l1_target}", file=sys.stderr)
            return 1
        l1_files = [l1_target]

    l2_pages = _patched_find_l2_pages(args.project)

    candidates: list[Path] = []
    for l1 in l1_files:
        # 우리 stem = basename (vendor 의 path_to_stem 의 우리 DevHub 버전)
        stem = _patched_path_to_stem(str(l1.relative_to(L1_BASE_REAL)))
        l2 = l2_pages.get(stem)
        if l2 and _patched_needs_body(l2):
            candidates.append((l1, l2, "l1"))

    if args.mode in ("metadata-only", "all"):
        for stem, l2 in l2_pages.items():
            if not _patched_needs_body(l2):
                continue
            if any(c[1] == l2 for c in candidates):
                continue
            candidates.append((None, l2, "metadata-only"))

    if args.limit > 0:
        candidates = candidates[:args.limit]

    print(f"[emit-wiki-l2-devhub-vendor] L1 files: {len(l1_files)}")
    print(f"[emit-wiki-l2-devhub-vendor] L2 pages: {len(l2_pages)}")
    print(f"[emit-wiki-l2-devhub-vendor] candidates (needs body): {len(candidates)}")
    print(f"[emit-wiki-l2-devhub-vendor] mode: {args.mode}")
    print(f"[emit-wiki-l2-devhub-vendor] apply: {'YES' if not dry_run else 'NO (DRY-RUN)'}")
    print(f"[emit-wiki-l2-devhub-vendor] max_chars: {args.max_chars}")
    print()

    emitted = 0
    for l1, l2, mode in candidates:
        rel_l2 = l2.relative_to(L1_BASE_REAL)
        if dry_run:
            print(f"  [DRY ({mode})] {rel_l2}")
        else:
            new_body = _patched_build_emit_body(l1, today, max_chars=args.max_chars, mode=mode)
            # 기존 L2 frontmatter 보존 (type, status, last_ingested_from, dates 등).
            # 새 frontmatter 는 기존 보존 + 빌드 결과 body 결합.
            existing_fm: dict[str, str] = {}
            if l2.is_file():
                _etext = l2.read_text(encoding="utf-8")
                _em = re.match(r"^---\n(.+?)\n---\n?", _etext, re.DOTALL)
                if _em:
                    for _line in _em.group(1).split("\n"):
                        if ":" in _line:
                            _k, _, _v = _line.partition(":")
                            existing_fm[_k.strip()] = _v.strip()
            # 새 frontmatter: 기존 보존 + update last_touched (apply 시점)
            if "last_touched" not in existing_fm:
                existing_fm["last_touched"] = today
            else:
                existing_fm["last_touched"] = today
            if "updated" not in existing_fm:
                existing_fm["updated"] = today
            if "type" not in existing_fm:
                # derive from L1 dir when possible, else metadata-only
                if l1 is not None:
                    try:
                        _r = l1.relative_to(L1_BASE_REAL)
                        _d = _r.parts[0] if _r.parts else ""
                    except ValueError:
                        _d = ""
                    existing_fm["type"] = _l1_dir_to_type.get(_d, "concept")
                else:
                    existing_fm["type"] = "concept"
            fm_lines = [f"{k}: {v}" for k, v in existing_fm.items()]
            fm_block = "---\n" + "\n".join(fm_lines) + "\n---\n\n"
            l2.write_text(fm_block + new_body, encoding="utf-8")
            print(f"  [APPLIED ({mode})] {rel_l2}")
            emitted += 1

    print()
    if dry_run:
        print(f"[emit-wiki-l2-devhub-vendor] Dry-run complete. {len(candidates)} page 가 emit 대상. --apply 로 실제 실행.")
    else:
        print(f"[emit-wiki-l2-devhub-vendor] Applied {emitted} page.")
    return 0


def main() -> int:
    if not VENDOR_TOOL.is_file():
        print(f"[emit-wiki-l2-devhub-vendor] error: vendor 도구 부재: {VENDOR_TOOL}", file=sys.stderr)
        return 2

    # vendor 도구 import
    spec = importlib.util.spec_from_file_location("vendor_emit_wiki_l2_body", str(VENDOR_TOOL))
    module = importlib.util.module_from_spec(spec)
    sys.modules["vendor_emit_wiki_l2_body"] = module
    spec.loader.exec_module(module)

    # Monkey-patch: module-level 의 RAW_MIRROR / L1_BASE / L2_SOURCES + path_to_stem /
    # find_l1_files / find_l2_pages / build_emit_body / main
    module.RAW_MIRROR = L1_BASE_REAL
    module.L1_BASE = L1_BASE_REAL
    module.L2_SOURCES = L2_SOURCES_REAL
    module.path_to_stem = _patched_path_to_stem
    module.find_l1_files = _patched_find_l1_files
    module.find_l2_pages = _patched_find_l2_pages
    module.build_emit_body = _patched_build_emit_body

    # vendor 의 main 함수를 *우리의 _patched_main* 으로 교체
    orig_main = module.main
    module.main = lambda: _patched_main(orig_main, module)

    # vendor 의 main() 호출 (=> 우리 _patched_main)
    return module.main()


if __name__ == "__main__":
    sys.exit(main())
