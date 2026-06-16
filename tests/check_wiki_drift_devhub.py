#!/usr/bin/env python3
"""DevHub 의 wiki drift check — vendor 의 check_wiki_drift.py 의 *DevHub adapter*.

Codex P2 (PR #602, 2026-06-15) 의 본체: vendor 의 check_wiki_drift.py 가
SOURCE_ROOT 의 parents[1] (= 우리 DevHub 의 vendor/ 디렉터리) 를 REPO_ROOT 로
derive → vendor/ai-workflow/wiki/index.md 를 찾는데 우리 DevHub 의 wiki 는
저장소 root 의 ai-workflow/wiki/index.md (vendor 디렉터리 바깥).

본 adapter 는 vendor 의 test 의 REPO_ROOT / INREPO_WIKI 를 *저장소 root* 로
monkey-patch 한 뒤 test 의 main() 호출. vendor 의 test 자체 는 read-only
reference 라 *수정* ❌, *DevHub wrapper* ✅.

Test 구성 (4/4 PASS):
1. test_inrepo_wiki_l1_drift: L1 wiki 의 updated: < L1 code mtime + 7일 (drift 0)
2. test_vault_l2_drift: L2 sources/ 의 last_touched < L1 raw mirror mtime + 7일 (drift 0)
3. test_ingested_from_paths_exist: L1 의 last_ingested_from 의 path 가 실제 존재
4. test_l1_wiki_pages_format: L1 wiki 의 frontmatter + type/status field 정합

5. test_v070_concept_pages_indexed: vendor 의 test 가 우리 DevHub 의 *vendor/ai-workflow/wiki/index.md*
   를 찾는데 본 adapter 의 monkey-patch 로 skip. (DevHub 의 index.md 는 ai-workflow/wiki/index.md,
   vendor 디렉터리 바깥). 본 test 는 vendor 자체 repo 의 test 로서만 활성.

Reference:
- vendor/standard_ai_workflow/tests/check_wiki_drift.py (v0.7.17+, vendor 자체 drift check)
- ai-workflow/wiki/index.md (DevHub 의 L0 Home)
- ai-workflow/wiki/RAW_MIRROR_MANIFEST.md (raw mirror 운영 가이드)
- ai-workflow/minimax_code_workflow.md §4.4 (vendor smoke 회귀 표)
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
VENDOR_TEST = (
    REPO_ROOT
    / "vendor"
    / "standard_ai_workflow"
    / "tests"
    / "check_wiki_drift.py"
)


def _import_vendor_test():
    """vendor 의 check_wiki_drift module import (4-priority REPO_ROOT auto-detect 후 monkey-patch)."""
    spec = importlib.util.spec_from_file_location(
        "vendor_check_wiki_drift", str(VENDOR_TEST)
    )
    module = importlib.util.module_from_spec(spec)
    sys.modules["vendor_check_wiki_drift"] = module
    spec.loader.exec_module(module)

    # Monkey-patch: vendor 의 test 의 SOURCE_ROOT / REPO_ROOT / INREPO_WIKI 를
    # *본 저장소 root* 로 강제. vendor 의 SOURCE_ROOT = vendor/standard_ai_workflow/ 이지만
    # 우리 DevHub 의 wiki 위치 = REPO_ROOT/ai-workflow/wiki/. 따라서 REPO_ROOT = 본 저장소 root.
    # SOURCE_ROOT = vendor/ 디렉터리 (test 의 Path(__file__).parents[1] 의 의도 보존).
    module.SOURCE_ROOT = REPO_ROOT / "vendor" / "standard_ai_workflow"
    module.REPO_ROOT = REPO_ROOT
    module.INREPO_WIKI = REPO_ROOT / "ai-workflow" / "wiki"
    module.L2_SOURCES = module.INREPO_WIKI / "sources"
    module.RAW_MIRROR = module.INREPO_WIKI
    return module


# ----- L1 format SSOT (docs/governance/l1-format.md) -----
# DevHub 의 5 L1 dir (concepts/decisions/entities/patterns/topics) 의 *.md file 의
# frontmatter 검증. 본 function 은 module-level (import 가능) — tests/check_l1_format_devhub.py
# 의 edge case test 에서 직접 import + 호출.
#
# 검증 항목 (SSOT 와 1:1):
#   1. frontmatter delimiter (\`---\` ... \`---\`) 존재
#   2. required field 4개: type, status, created, updated
#   3. type enum: {concept, decision, entity, pattern, topic}
#   4. status enum: {active, draft}
#   5. created / updated date format: YYYY-MM-DD (ISO 8601)
#
# Page 분류: 5 L1 dir 만 검증 대상. index.md, RAW_MIRROR_MANIFEST.md, sources/, raw/ 는 skip.
#
# Returns: (issue_count, issues_list). issue_count == 0 면 PASS.
def validate_l1_pages(inrepo_wiki: Path) -> tuple[int, list[str]]:
    import re
    VALID_TYPES = {"concept", "decision", "entity", "pattern", "topic"}
    VALID_STATUS = {"active", "draft"}
    REQUIRED_FIELDS = ("type", "status", "created", "updated")
    DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
    L1_DIRS = ("concepts", "decisions", "entities", "patterns", "topics")
    issues: list[str] = []
    for md in inrepo_wiki.rglob("*.md"):
        if md.name in ("index.md", "RAW_MIRROR_MANIFEST.md"):
            continue
        rel = md.relative_to(inrepo_wiki)
        if rel.parts[0] not in L1_DIRS:
            continue
        text = md.read_text(encoding="utf-8")
        m = re.match(r"^---\n(.+?)\n---", text, re.DOTALL)
        if not m:
            issues.append(f"{md.name}: missing frontmatter")
            continue
        fm = m.group(1)
        field_map: dict[str, str] = {}
        for line in fm.split("\n"):
            if ":" in line:
                k, _, v = line.partition(":")
                field_map[k.strip()] = v.strip()
        for required in REQUIRED_FIELDS:
            if required not in field_map:
                issues.append(f"{md.name}: missing {required}")
        if field_map.get("type") not in VALID_TYPES:
            issues.append(
                f"{md.name}: invalid type: {field_map.get('type')!r} "
                f"(valid: {sorted(VALID_TYPES)})"
            )
        if field_map.get("status") not in VALID_STATUS:
            issues.append(
                f"{md.name}: invalid status: {field_map.get('status')!r} "
                f"(valid: {sorted(VALID_STATUS)})"
            )
        for date_field in ("created", "updated"):
            v = field_map.get(date_field)
            if v is not None and not DATE_RE.match(v):
                issues.append(
                    f"{md.name}: invalid {date_field} date format: {v!r} "
                    f"(expected YYYY-MM-DD)"
                )
    return (len(issues), issues)


def main() -> int:
    if not VENDOR_TEST.is_file():
        print(f"[check-wiki-drift-devhub] error: vendor test not found: {VENDOR_TEST}")
        return 2

    print(f"[check-wiki-drift-devhub] vendor test: {VENDOR_TEST}")
    print(f"[check-wiki-drift-devhub] REPO_ROOT: {REPO_ROOT}")
    print(f"[check-wiki-drift-devhub] INREPO_WIKI: {REPO_ROOT / 'ai-workflow' / 'wiki'}")
    print("")

    module = _import_vendor_test()
    # vendor 의 test 가 자체 pytest test_* 함수들을 정의하지만, main() 으로 직접 실행 가능.
    # 여기서는 pytest-style test_* 함수들을 직접 호출하여 PASS/FAIL 집계.
    # vendor 의 test 의 main() 가 없다면 (의미: vendor 의 test 는 *pytest runner* 가 호출), 우리가 수동 호출.
    test_functions = [
        ("test_inrepo_wiki_l1_drift", getattr(module, "test_inrepo_wiki_l1_drift", None)),
        ("test_vault_l2_drift", getattr(module, "test_vault_l2_drift", None)),
        ("test_ingested_from_paths_exist", getattr(module, "test_ingested_from_paths_exist", None)),
        ("test_l1_wiki_pages_format", getattr(module, "test_l1_wiki_pages_format", None)),
        # test_v070_concept_pages_indexed 는 *vendor 자체 repo 의 self-test* → DevHub 에서는 skip
    ]


    passed = 0
    failed = 0
    for name, fn in test_functions:
        if fn is None:
            print(f"  SKIP  {name} (function not present in vendor test)")
            continue
        # DevHub 의 wiki format 은 vendor 의 ADR-format 과 다름
        # (type 5종 + status:active/active_since, adr_id 없음)
        # → test_l1_wiki_pages_format 는 DevHub 자체 검증으로 대체 (vendor 의 test skip)
        if name == "test_l1_wiki_pages_format":
            print(f"  SKIP  {name} (DevHub 자체 format 검증 — vendor ADR-format 과 의미 다름)")
            # DevHub 자체 frontmatter 검증 (module-level SSOT 함수)
            self_validate_count, self_validate_issues = validate_l1_pages(module.INREPO_WIKI)
            if self_validate_count == 0:
                print(f"  PASS  _devhub_self_validate_l1_pages (DevHub 자체 format)")
                passed += 1
            else:
                print(f"  FAIL  _devhub_self_validate_l1_pages: {self_validate_count} issue")
                for issue in self_validate_issues:
                    print(f"    {issue}")
                failed += 1
            continue
        try:
            fn()
            print(f"  PASS  {name}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {name}: {e}")
            failed += 1
        except Exception as e:
            print(f"  ERROR {name}: {type(e).__name__}: {e}")
            failed += 1

    print("")
    print(f"[check-wiki-drift-devhub] {passed} pass, {failed} fail")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
