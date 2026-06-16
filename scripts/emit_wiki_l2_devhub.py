#!/usr/bin/env python3
"""emit_wiki_l2_devhub.py — DevHub 의 in-repo wiki L2 dense emit tool (자체).

vendor 의 `emit_wiki_l2_body.py` 의 *mini structure* 하드코딩 (RAW_MIRROR / project /
ai-workflow / wiki) 가 우리 DevHub 의 *운영 위치* (ai-workflow/wiki/) 와 정합 안 함. 본
도구는 *vendor 미사용* 으로 자체 구현. 우리 DevHub 의 L1 → L2 dense emit + A안 5 page
의 *전체 L2 자동화* + 향후 220+ file 의 L1 page 의 *L2 emit* 가능.

본 도구 의 정공법 (DevHub-specific):
- L1 = ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/*.md (본 저장소 의 운영 위치)
- L2 = ai-workflow/wiki/sources/<stem>.md (vendor 의 L2_BASE 와 같은 위치)
- stem = L1 file 의 basename (e.g. concepts/devhub-overview.md → devhub-overview)
- L1 의 `last_ingested_from` 의 path 가 *본 저장소 의 real path* (raw mirror 의 path X) — byte-identical 정합

Usage:
    # Dry-run: 어떤 page 가 emit 대상인지 preview
    python3 scripts/emit_wiki_l2_devhub.py

    # Apply: 실제 L2 dense emit
    python3 scripts/emit_wiki_l2_devhub.py --apply

    # 단일 file L1 만 emit
    python3 scripts/emit_wiki_l2_devhub.py --source concepts/devhub-overview.md --apply

    # Vendor monkey-patch wrapper (vendor 의 도구 의 mini structure 우회):
    python3 scripts/emit_wiki_l2_devhub_vendor.py --apply

Wiki: ai-workflow/wiki/decisions/v0.7.37-import.md
      ai-workflow/wiki/decisions/v0.7.17-import.md
      ai-workflow/wiki/concepts/devhub-overview.md
"""

from __future__ import annotations

import argparse
import re
import sys
from datetime import date
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
L1_BASE = REPO_ROOT / "ai-workflow" / "wiki"
L2_SOURCES = L1_BASE / "sources"

# 우리 DevHub 의 L1 dir 5종 (in-repo 운영 위치)
L1_DIRS = ["concepts", "decisions", "entities", "patterns", "topics"]


def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def _parse_frontmatter(text: str) -> tuple[dict, str]:
    """frontmatter + body 분리. (dict, body_string) 반환."""
    m = re.match(r"^---\n(.+?)\n---\n?", text, re.DOTALL)
    if not m:
        return ({}, text)
    fm_str = m.group(1)
    body = text[m.end():]
    fm = {}
    for line in fm_str.split("\n"):
        if ":" in line:
            k, _, v = line.partition(":")
            fm[k.strip()] = v.strip()
    return (fm, body)


def _find_l1_files() -> list[Path]:
    """L1 = {concepts,decisions,entities,patterns,topics}/*.md 의 모든 file."""
    files: list[Path] = []
    for sub in L1_DIRS:
        d = L1_BASE / sub
        if d.is_dir():
            files.extend(sorted(p for p in d.glob("*.md") if p.name != ".gitkeep"))
    return files


def _stem_from_l1(l1: Path) -> str:
    """L1 file 의 basename (without .md) = stem. e.g. concepts/devhub-overview.md → devhub-overview"""
    return l1.stem


def _l2_path_for(l1: Path) -> Path:
    """L1 file 에 대응하는 L2 path. concepts/devhub-overview.md → sources/devhub-overview.md"""
    return L2_SOURCES / f"{_stem_from_l1(l1)}.md"


def _build_l2_body(l1: Path, today: str, max_chars: int = 2000) -> str:
    """L1 page 의 frontmatter + body 일부 → L2 dense body.

    Vendor 의 build_emit_body 와 정합 (TL;DR + 1차 출처 reference + 본문 cap).
    """
    text = _read(l1)
    fm, body = _parse_frontmatter(text)
    fm_lines = [f"{k}: {v}" for k, v in fm.items()]

    # TL;DR: 본문 의 첫 # 헤더 + 그 다음 paragraph, 또는 frontmatter 의 summary field
    h1_match = re.search(r"^#\s+(.+)$", body, re.MULTILINE)
    title = h1_match.group(1).strip() if h1_match else l1.stem.replace("-", " ").title()

    # 본문 cap: 첫 paragraph (TL;DR section) 우선, 없으면 전체 body
    tl_dr_match = re.search(r"^##\s+TL;DR\s*\n+(.+?)(?=\n##\s|\Z)", body, re.DOTALL | re.MULTILINE)
    if tl_dr_match:
        snippet = tl_dr_match.group(1).strip()
    else:
        # 첫 paragraph (## 다음)
        first_para_match = re.search(r"^##\s+\S.*?\n\n(.+?)(?=\n##\s|\n#\s|\Z)", body, re.DOTALL | re.MULTILINE)
        if first_para_match:
            snippet = first_para_match.group(1).strip()
        else:
            # body 의 첫 1500자
            snippet = body.strip()[:1500]

    # L2 dense 본문 (max_chars cap)
    snippet_capped = snippet[:max_chars]
    if len(snippet) > max_chars:
        snippet_capped += "\n\n…(L1 본문 cap, 전체 본문은 L1 SSOT 참조)"

    l2_body = f"""# {title} (L2 dense, DevHub-specific)

> **L1 SSOT**: `ai-workflow/wiki/{l1.relative_to(L1_BASE)}`
> 본 L2 derived view 는 in-repo retrieval 용 압축 요약. dense content 는 L1 SSOT 참조.

## TL;DR

{snippet_capped}

## Source

- 본 L2 file: `ai-workflow/wiki/sources/{_stem_from_l1(l1)}.md`
- L1 SSOT: `ai-workflow/wiki/{l1.relative_to(L1_BASE)}`
- L0 Home: `ai-workflow/wiki/index.md`
- 본 도구: `scripts/emit_wiki_l2_devhub.py` (in-repo, vendor 미사용)
- 1차 출처: {fm.get('last_ingested_from', 'n/a')}
"""
    return l2_body


def _needs_body(l2: Path) -> bool:
    """L2 page 의 본문이 placeholder 또는 짧은지 확인 (emit 대상)."""
    if not l2.is_file():
        return True
    text = _read(l2)
    if "<needs content>" in text:
        return True
    # 본문이 200자 미만이면 emit
    _, body = _parse_frontmatter(text)
    return len(body.strip()) < 200


def main() -> int:
    parser = argparse.ArgumentParser(
        description="DevHub wiki L2 dense emit tool (자체). Vendor 의 mini structure 무시."
    )
    parser.add_argument("--apply", action="store_true", help="실제 L2 emit (default = dry-run)")
    parser.add_argument("--source", type=str, help="단일 L1 file 의 상대 경로 (concepts/devhub-overview.md)")
    parser.add_argument("--max-chars", type=int, default=2000, help="L1 본문 cap (default 2000)")
    parser.add_argument("--limit", type=int, default=0, help="max N page (default 0 = 무제한)")
    args = parser.parse_args()

    today = date.today().isoformat()
    dry_run = not args.apply

    l1_files = _find_l1_files()
    if args.source:
        # 단일 file
        l1_target = L1_BASE / args.source
        if not l1_target.is_file():
            print(f"[emit-wiki-l2-devhub] error: --source not found: {l1_target}", file=sys.stderr)
            return 1
        l1_files = [l1_target]

    candidates: list[Path] = []
    for l1 in l1_files:
        l2 = _l2_path_for(l1)
        if _needs_body(l2):
            candidates.append(l1)

    if args.limit > 0:
        candidates = candidates[:args.limit]

    print(f"[emit-wiki-l2-devhub] L1_BASE: {L1_BASE}")
    print(f"[emit-wiki-l2-devhub] L2_SOURCES: {L2_SOURCES}")
    print(f"[emit-wiki-l2-devhub] L1 files: {len(l1_files)} (전체 L1 page)")
    print(f"[emit-wiki-l2-devhub] L2 candidates (needs body): {len(candidates)}")
    print(f"[emit-wiki-l2-devhub] Mode: {'APPLY' if not dry_run else 'DRY-RUN'}")
    print(f"[emit-wiki-l2-devhub] Max chars: {args.max_chars}")
    print()

    emitted = 0
    for l1 in candidates:
        l2 = _l2_path_for(l1)
        rel_l1 = l1.relative_to(L1_BASE)
        rel_l2 = l2.relative_to(L1_BASE)
        if dry_run:
            print(f"  [DRY] L1={rel_l1} → L2={rel_l2}")
        else:
            L2_SOURCES.mkdir(parents=True, exist_ok=True)
            frontmatter = f"""---
type: concept
status: active
last_ingested_from: ai-workflow/wiki/{rel_l1}
related_pages: [sources/{_stem_from_l1(l1)}]
created: {today}
updated: {today}
last_touched: {today}
---

"""
            body = _build_l2_body(l1, today, max_chars=args.max_chars)
            l2.write_text(frontmatter + body, encoding="utf-8")
            print(f"  [APPLIED] L1={rel_l1} → L2={rel_l2}")
            emitted += 1

    print()
    if dry_run:
        print(f"[emit-wiki-l2-devhub] Dry-run complete. {len(candidates)} page 가 emit 대상. --apply 로 실제 실행.")
    else:
        print(f"[emit-wiki-l2-devhub] Applied {emitted} page.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
