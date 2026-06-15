#!/usr/bin/env python3
"""v0.7.1: existing_project_onboarding 의 9-Artifact 자동 fill helper.

기존 `repository_assessment.md` (단일 92 line file) 의 한계 — 주제별 SSOT 부재.
AIDLC 의 `inception/reverse-engineering.md` 9-Artifact 구조를 차용하여
brownfield project 의 9 md file 자동 emit.

9 artifact:
1. 01-business-overview.md
2. 02-architecture.md
3. 03-code-structure.md
4. 04-api-documentation.md
5. 05-component-inventory.md
6. 06-technology-stack.md
7. 07-dependencies.md
8. 08-code-quality-assessment.md
9. 09-reverse-engineering-metadata.md

Usage:
    # dry-run
    python3 fill_reverse_engineering_artifacts.py --project-root=~/repos/foo --dry-run

    # 실제 emit (workflow-source/reverse-engineering/ 의 template 사용)
    python3 fill_reverse_engineering_artifacts.py --project-root=~/repos/foo --output-dir=./out --apply

Reference:
- workflow-source/reverse-engineering/{01..09}-*.md (template)
- workflow-source/core/reverse_engineering.md (13 step 가이드)
- workflow-source/scripts/run_existing_project_onboarding.py (기존 onboarding 흐름)
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import datetime
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TEMPLATE_DIR = SOURCE_ROOT / "reverse-engineering"

# 9 artifact 정의 (file_name → template path)
ARTIFACTS = [
    "01-business-overview.md",
    "02-architecture.md",
    "03-code-structure.md",
    "04-api-documentation.md",
    "05-component-inventory.md",
    "06-technology-stack.md",
    "07-dependencies.md",
    "08-code-quality-assessment.md",
    "09-reverse-engineering-metadata.md",
]

# 각 artifact 의 auto-fill 정책 (heuristic)
AUTO_FILL_RULES = {
    "01-business-overview.md": [
        ("Project Purpose", lambda info: info.get("purpose", "[TODO: describe the system's business purpose]")),
        ("Workflow Domain", lambda info: "workflow stage transition (inception → design → impl → validation → review → release)"),
        ("Business Transactions", lambda info: "\n".join(f"- {t}" for t in info.get("transactions", ["[TODO: list business transactions]"]))),
        ("Business Dictionary", lambda info: "\n".join(f"- **{k}**: {v}" for k, v in info.get("dictionary", {}).items()) or "- [TODO: define key terms]"),
    ],
    "02-architecture.md": [
        ("Components", lambda info: "\n".join(f"- **{c['name']}** ({c.get('type', 'Application')}): {c.get('purpose', '[TODO]')}" for c in info.get("components", [])) or "- [TODO: list components]"),
        ("Architecture", lambda info: "[TODO: paste Mermaid diagram or describe components + relationships]"),
        ("Data Flow", lambda info: "[TODO: paste sequence diagram of key flows]"),
        ("Integration Points", lambda info: "\n".join(f"- {p}" for p in info.get("integrations", ["[TODO: list integration points]"]))),
    ],
    "03-code-structure.md": [
        ("Build System", lambda info: f"- **Type**: {info.get('build_type', '[TODO: e.g. Python setuptools / npm]')}\n- **Configuration**: {info.get('build_config', '[TODO]')}"),
        ("Key Classes/Modules", lambda info: "[TODO: paste class diagram or module hierarchy]"),
        ("Existing Files Inventory", lambda info: "\n".join(f"- `{f}` - {info.get('file_purposes', {}).get(f, '[TODO: purpose]')}" for f in info.get("key_files", ["[TODO: list key files]"]))),
        ("Critical Dependencies", lambda info: "\n".join(f"- **{d['name']}** {d.get('version', '')}: {d.get('purpose', '[TODO]')}" for d in info.get("critical_deps", []))),
    ],
    "04-api-documentation.md": [
        ("REST APIs", lambda info: "\n".join(f"- **{e.get('name', '[TODO]')}**: `{e.get('method', 'GET')} {e.get('path', '/api/path')}`\n  - Purpose: {e.get('purpose', '[TODO]')}" for e in info.get("rest_apis", [])) or "- [TODO: list REST APIs (workflow: MCP tool interface)]"),
        ("MCP Tool APIs", lambda info: "\n".join(f"- **{t.get('name', '[TODO]')}**: `{t.get('tool_path', '[TODO: workflow-source/mcp_servers/<name>/]')}`" for t in info.get("mcp_tools", [])) or "- [TODO: list MCP tool APIs]"),
        ("Internal APIs (workflow_kit)", lambda info: "\n".join(f"- **{m.get('name', '[TODO]')}**: `workflow_kit.{m.get('path', '[TODO]')}`" for m in info.get("internal_modules", [])) or "- [TODO: list internal Python modules]"),
        ("Data Models", lambda info: "\n".join(f"- **{d.get('name', '[TODO]')}**: {d.get('description', '[TODO]')}" for d in info.get("data_models", [])) or "- [TODO: list data models]"),
    ],
    "05-component-inventory.md": [
        ("Harness Packages", lambda info: "\n".join(f"- {h}" for h in info.get("harnesses", ["[TODO: list supported harnesses]"]))),
        ("MCP Server Packages", lambda info: "\n".join(f"- {m}" for m in info.get("mcp_servers", ["[TODO: list MCP servers]"]))),
        ("workflow_kit Shared Packages", lambda info: "\n".join(f"- {m}" for m in info.get("wk_modules", ["[TODO: list workflow_kit modules]"]))),
        ("Template/Extension Packages", lambda info: "\n".join(f"- {t}" for t in info.get("templates", ["[TODO: list templates/extensions]"]))),
        ("Test Packages", lambda info: "\n".join(f"- {t}" for t in info.get("tests", ["[TODO: list smoke tests]"]))),
        ("Total Count", lambda info: f"- **Total Components**: {info.get('total', 0)}\n- **Harness**: {info.get('harness_count', 0)}\n- **MCP**: {info.get('mcp_count', 0)}\n- **workflow_kit**: {info.get('wk_count', 0)}\n- **Template/Extension**: {info.get('template_count', 0)}\n- **Test**: {info.get('test_count', 0)}"),
    ],
    "06-technology-stack.md": [
        ("Programming Languages", lambda info: "\n".join(f"- {l}" for l in info.get("languages", ["[TODO: list languages + versions"]))),
        ("Frameworks", lambda info: "\n".join(f"- {f}" for f in info.get("frameworks", ["[TODO: list frameworks]"]))),
        ("Infrastructure (Harness Runtime)", lambda info: "\n".join(f"- {i}" for i in info.get("infra", ["[TODO: list harness runtime]"]))),
        ("Build Tools", lambda info: "\n".join(f"- {t}" for t in info.get("build_tools", ["[TODO: list build tools]"]))),
        ("Testing Tools", lambda info: "\n".join(f"- {t}" for t in info.get("test_tools", ["[TODO: list testing tools]"]))),
    ],
    "07-dependencies.md": [
        ("Internal Dependencies", lambda info: "[TODO: paste Mermaid graph of internal dependencies]\n\n" + "\n".join(f"### [{a}] depends on [{b}]\n- **Type**: {t}\n- **Reason**: {r}" for a, b, t, r in info.get("internal_deps", []))),
        ("External Dependencies", lambda info: "\n".join(f"- **{d.get('name', '[TODO]')}**: {d.get('version', '[TODO]')} ({d.get('license', '[TODO]')})\n  - Purpose: {d.get('purpose', '[TODO]')}" for d in info.get("external_deps", []))),
        ("Dependency Resolution Policy", lambda info: f"- **Lock File**: {info.get('lock_file', '[TODO: None / requirements.txt / uv.lock / poetry.lock]')}\n- **Checksum Verification**: {info.get('checksum', '[TODO: None / SHA256 / PGP — SEC-WF-05 의도]')}\n- **Version Pinning**: {info.get('pinning', '[TODO: Loose >= / Strict == / Range]')}"),
    ],
    "08-code-quality-assessment.md": [
        ("Test Coverage", lambda info: f"- **Overall**: {info.get('test_coverage', '[TODO: percentage or Good/Fair/Poor/None]')}\n- **Smoke Tests**: {info.get('smoke_tests', '[TODO: PASS/FAIL count]')}\n- **Unit Tests**: {info.get('unit_tests', '[TODO: count or N/A]')}"),
        ("Code Quality Indicators", lambda info: f"- **Linting**: {info.get('linting', '[TODO: Configured/Not configured]')}\n- **R-1~R9 Compliance**: {info.get('r1_r9', '[TODO: rule별 pass/fail count]')}\n- **Type Hints**: {info.get('type_hints', '[TODO: percentage]')}"),
        ("Technical Debt", lambda info: "\n".join(f"- {d}" for d in info.get("debt", ["[TODO: list technical debt items]"]))),
        ("Good Patterns", lambda info: "\n".join(f"- {p}" for p in info.get("good_patterns", ["[TODO: list good patterns]"]))),
        ("Anti-patterns", lambda info: "\n".join(f"- {a}" for a in info.get("anti_patterns", ["[TODO: list anti-patterns]"]))),
    ],
    "09-reverse-engineering-metadata.md": [
        ("Analysis Date", lambda info: datetime.now().isoformat(timespec="seconds")),
        ("Analyzer", lambda info: info.get("analyzer", "[TODO: Human / AI agent / tool name]")),
        ("Workspace", lambda info: info.get("workspace", "[TODO: absolute path]")),
        ("Total Files Analyzed", lambda info: str(info.get("total_files", 0))),
        ("Workflow Version", lambda info: info.get("workflow_version", "v0.7.1-beta")),
        ("Reverse Engineering Standard", lambda info: "AIDLC inception/reverse-engineering.md @ b19c819"),
    ],
}


def load_template(artifact_name: str) -> str:
    """workflow-source/reverse-engineering/<name> 의 template 본문."""
    path = TEMPLATE_DIR / artifact_name
    if not path.exists():
        return ""
    content = path.read_text(encoding="utf-8")
    # body 만 (frontmatter 또는 헤더 제외)
    if content.startswith("---\n"):
        end = content.find("\n---\n", 4)
        if end >= 0:
            return content[end + 5:]
    return content


def auto_fill(artifact_name: str, info: dict) -> str:
    """heuristic 기반 auto-fill. TODO marker 는 사용자 채움 필요."""
    template = load_template(artifact_name)
    rules = AUTO_FILL_RULES.get(artifact_name, [])

    # template 의 §N 섹션별 placeholder 에 채우기
    # template 의 "## N.X" 또는 "## N." 로 시작하는 각 section title
    # heuristic: section title 다음의 bullet 1개를 rule 결과로 교체
    for sec_title, filler in rules:
        # template 에서 sec_title 등장 찾기
        pattern = re.compile(
            rf"(## [^#\n]*{re.escape(sec_title)}[^#\n]*\n)(.*?)(\n## |\Z)",
            re.DOTALL,
        )
        replacement = filler(info)
        m = pattern.search(template)
        if m:
            template = template[:m.start()] + m.group(1) + replacement + template[m.end():]

    return template


def fill_one(artifact_name: str, info: dict, output_dir: Path) -> Path:
    """단일 artifact 의 output file."""
    out = output_dir / artifact_name
    content = auto_fill(artifact_name, info)
    out.write_text(content, encoding="utf-8")
    return out


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument("--info", help="JSON 파일 path (heuristic 정보)")
    parser.add_argument("--project-root", help="project root (heuristic 자동 추출)")
    parser.add_argument("--output-dir", default="./out", help="output directory (default: ./out)")
    parser.add_argument("--apply", action="store_true", help="실제 파일 emit (default: dry-run)")
    parser.add_argument("--limit", type=int, default=0, help="max N artifact (default: 9)")
    args = parser.parse_args()

    dry_run = not args.apply

    # info dict 로드
    if args.info:
        info = json.loads(Path(args.info).read_text(encoding="utf-8"))
    elif args.project_root:
        info = {"project_root": args.project_root, "analyzer": "auto-fill", "workspace": args.project_root}
    else:
        info = {}

    artifacts = ARTIFACTS[:args.limit] if args.limit > 0 else ARTIFACTS
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    print(f"9-Artifact auto-fill helper (v0.7.1)")
    print(f"Output dir: {output_dir.resolve()}")
    print(f"Artifacts: {len(artifacts)}")
    print(f"Mode: {'APPLY' if args.apply else 'DRY-RUN'}")
    print()

    for name in artifacts:
        if dry_run:
            print(f"  [DRY] {name}")
        else:
            out = fill_one(name, info, output_dir)
            print(f"  [APPLIED] {out.relative_to(output_dir.parent)}")

    print()
    if dry_run:
        print(f"Dry-run complete. {len(artifacts)} artifact. --apply 로 실제 emit.")
    else:
        print(f"Emitted {len(artifacts)} artifact to {output_dir}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
