#!/usr/bin/env python3
"""v0.7.17+ in-repo wiki redirect smoke test (DevHub-specific).

vendor 의 standard_ai_workflow v0.7.17 wiki in-repo isolation check (11/11 PASS) 와 별개로,
본 저장소 (Devhub_example_minimax) 의 자체 파일 (scripts/, docs/llm-wiki/, .mavis/) 에
남아있는 외부 vault (`~/wiki/`) reference 0 을 검증.

Test 구성 (4 test):
1. scripts/wiki-*.sh: VAULT_ROOT 가 in-repo path (ai-workflow/wiki), ${HOME}/wiki 0
2. docs/llm-wiki/*.md: 외부 vault `~/wiki/` literal 0
3. ai-workflow/wiki/{concepts,decisions,entities,patterns,topics,sources} 디렉터리 6개 존재
4. ai-workflow/memory/log.md 존재 (event log target, v0.7.17 신규)

Reference:
- vendor/standard_ai_workflow/tests/check_v0_7_17_wiki_in_repo_isolation.py (11/11 PASS)
- docs/llm-wiki/README.md (D-72 Phase 1+1.5+3 → in-repo redirect, 2026-06-15 갱신)
- ai-workflow/wiki/ (v0.7.17+ in-repo wiki, 외부 vault 미사용)
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
WIKI_DIR = REPO_ROOT / "ai-workflow" / "wiki"
SCRIPTS_DIR = REPO_ROOT / "scripts"
DOCS_LLM_WIKI = REPO_ROOT / "docs" / "llm-wiki"

# 1차 출처 = in-repo 만
INREPO_PATH = "ai-workflow/wiki"
LEGACY_VAULT = "${HOME}/wiki"
LEGACY_VAULT_LITERAL = "~/wiki"


def test_scripts_no_legacy_vault_root():
    """scripts/wiki-*.sh: VAULT_ROOT 가 in-repo path (${HOME}/wiki 0)."""
    legacy_count = 0
    legacy_files = []
    for f in sorted(SCRIPTS_DIR.glob("wiki-*.sh")):
        text = f.read_text(encoding="utf-8")
        if LEGACY_VAULT in text:
            legacy_count += text.count(LEGACY_VAULT)
            legacy_files.append(f.name)
    assert legacy_count == 0, (
        f"scripts/wiki-*.sh 의 VAULT_ROOT 가 legacy (${{HOME}}/wiki) {legacy_count} 회 사용: {legacy_files}"
    )


def test_docs_no_legacy_wiki_literal():
    """docs/llm-wiki/*.md: 외부 vault ~/wiki/ literal 0."""
    legacy_count = 0
    legacy_files = []
    for f in sorted(DOCS_LLM_WIKI.glob("*.md")):
        text = f.read_text(encoding="utf-8")
        if LEGACY_VAULT_LITERAL in text:
            legacy_count += text.count(LEGACY_VAULT_LITERAL)
            legacy_files.append(f.name)
    assert legacy_count == 0, (
        f"docs/llm-wiki/*.md 의 외부 vault literal (~/wiki/) {legacy_count} 회 사용: {legacy_files}"
    )


def test_inrepo_wiki_dirs_exist():
    """ai-workflow/wiki/{concepts,decisions,entities,patterns,topics,sources} 6개 dir 존재."""
    expected = ["concepts", "decisions", "entities", "patterns", "topics", "sources"]
    missing = [d for d in expected if not (WIKI_DIR / d).is_dir()]
    assert not missing, f"in-repo wiki subdir {len(missing)} 개 부재: {missing}"


def test_inrepo_memory_log_exists():
    """ai-workflow/memory/log.md 존재 (v0.7.17 wiki event log target)."""
    log_path = REPO_ROOT / "ai-workflow" / "memory" / "log.md"
    assert log_path.is_file(), f"in-repo memory event log 부재: {log_path}"


def main() -> int:
    tests = [
        test_scripts_no_legacy_vault_root,
        test_docs_no_legacy_wiki_literal,
        test_inrepo_wiki_dirs_exist,
        test_inrepo_memory_log_exists,
    ]
    passed = 0
    failed = 0
    for t in tests:
        try:
            t()
            print(f"  PASS  {t.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {t.__name__}: {e}")
            failed += 1
        except Exception as e:
            print(f"  ERROR {t.__name__}: {type(e).__name__}: {e}")
            failed += 1
    print(f"\n{passed} pass, {failed} fail")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
