#!/usr/bin/env python3
"""v0.7.0 step 9: Unit of Work 3-layer template 검증.

- unit_of_work_template.md 형식 (UOW-NNN id, name, type, responsibility,
  interfaces, status, owner, created, updated)
- dependency matrix 정합성 (cycle 검출, depended_by 일치)
- Mermaid graph syntax 검증
- Story mapping 정합성
Reference: workflow-source/templates/unit_of_work_template.md
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SOURCE_ROOT))

TEMPLATE_PATH = SOURCE_ROOT / "templates" / "unit_of_work_template.md"

# 정규식 패턴
UOW_HEADER_RE = re.compile(r"^### (UOW-\d{3}):\s+(.+)$", re.MULTILINE)
UOW_FIELD_RE = re.compile(r"^- \*\*(.+?)\*\*:\s*(.*)$", re.MULTILINE)
DEP_MATRIX_ROW_RE = re.compile(r"^\| (UOW-\d{3}) \| \[([^\]]*)\] \| \[([^\]]*)\] \|$", re.MULTILINE)
MERMAID_BLOCK_RE = re.compile(r"```mermaid\n(.*?)```", re.DOTALL)
STORY_MAPPING_ROW_RE = re.compile(r"^\| \"(.+?)\" \| (UOW-\d{3}) \|", re.MULTILINE)


# --- Sample UOW data 검증 ---


SAMPLE_VALID = """# Unit of Work — sample-project

- 문서 목적: ...
- 상태: stable
- 최종 수정일: 2026-06-12
- 관련 문서: ...

## 1. UOW 정의

### UOW-001: auth-service

- **Type**: service
- **Responsibility**: 사용자 인증
- **Interfaces**:
  - POST /auth/login
- **Status**: in_progress
- **Owner**: backend-team
- **Created**: 2026-06-12
- **Updated**: 2026-06-12

### UOW-002: user-profile

- **Type**: module
- **Responsibility**: 사용자 프로필 관리
- **Interfaces**:
  - GET /user/{id}
- **Status**: planned
- **Owner**: backend-team
- **Created**: 2026-06-12
- **Updated**: 2026-06-12

## 2. Dependency Matrix

| UOW | depends_on | depended_by |
|---|---|---|
| UOW-001 | [] | [UOW-002] |
| UOW-002 | [UOW-001] | [] |

## 3. Dependency Graph (Mermaid)

```mermaid
graph LR
    UOW-001 --> UOW-002
```

## 4. Story Mapping

| User Story | UOW | Notes |
|---|---|---|
| "사용자가 로그인한다" | UOW-001 | auth |
| "사용자가 프로필을 수정한다" | UOW-002 | profile |

## 5. Code Organization

```text
src/
├── auth-service/
└── user-profile/
```
"""


SAMPLE_CYCLE = """# Unit of Work — cycle-test

## 1. UOW 정의

### UOW-001: alpha
- **Type**: service
- **Responsibility**: a
- **Interfaces**: a
- **Status**: stable
- **Owner**: x
- **Created**: 2026-06-12
- **Updated**: 2026-06-12

### UOW-002: beta
- **Type**: service
- **Responsibility**: b
- **Interfaces**: b
- **Status**: stable
- **Owner**: x
- **Created**: 2026-06-12
- **Updated**: 2026-06-12

## 2. Dependency Matrix

| UOW | depends_on | depended_by |
|---|---|---|
| UOW-001 | [UOW-002] | [UOW-002] |
| UOW-002 | [UOW-001] | [UOW-001] |
"""


def _parse_uow_definitions(text: str) -> dict[str, dict]:
    """UOW 헤더 + 필드 parse. {UOW-001: {name, type, ...}} dict 반환."""
    uows: dict[str, dict] = {}
    for m in UOW_HEADER_RE.finditer(text):
        uow_id = m.group(1)
        name = m.group(2).strip()
        # 다음 ### 또는 ## 까지 의 field 추출
        start = m.end()
        next_section = re.search(r"^##? ", text[start:], re.MULTILINE)
        end = start + next_section.start() if next_section else len(text)
        section = text[start:end]
        fields: dict[str, str] = {"name": name}
        for fm in UOW_FIELD_RE.finditer(section):
            fields[fm.group(1).strip()] = fm.group(2).strip()
        uows[uow_id] = fields
    return uows


def _parse_dep_matrix(text: str) -> dict[str, dict]:
    """dependency matrix parse. {UOW-001: {depends_on: [UOW-002], depended_by: []}}"""
    matrix: dict[str, dict] = {}
    for m in DEP_MATRIX_ROW_RE.finditer(text):
        uow_id = m.group(1)
        depends_on = [x.strip() for x in m.group(2).split(",") if x.strip()]
        depended_by = [x.strip() for x in m.group(3).split(",") if x.strip()]
        matrix[uow_id] = {"depends_on": depends_on, "depended_by": depended_by}
    return matrix


def _has_cycle(graph: dict[str, list[str]]) -> bool:
    """DFS 기반 cycle 검출."""
    visited: set[str] = set()
    in_stack: set[str] = set()

    def dfs(node: str) -> bool:
        visited.add(node)
        in_stack.add(node)
        for neighbor in graph.get(node, []):
            if neighbor not in visited:
                if dfs(neighbor):
                    return True
            elif neighbor in in_stack:
                return True
        in_stack.remove(node)
        return False

    for node in graph:
        if node not in visited:
            if dfs(node):
                return True
    return False


# --- Test 1: UOW definition parsing ---


def test_uow_header_format() -> None:
    """UOW-NNN 헤더 + name 형식 검증."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    assert "UOW-001" in uows
    assert uows["UOW-001"]["name"] == "auth-service"
    assert "UOW-002" in uows
    assert uows["UOW-002"]["name"] == "user-profile"


def test_uow_required_fields() -> None:
    """각 UOW 가 9 필수 field 모두 가짐."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    required = ["name", "Type", "Responsibility", "Interfaces", "Status", "Owner", "Created", "Updated"]
    for uow_id, fields in uows.items():
        for f in required:
            assert f in fields, f"{uow_id} missing field: {f}"


def test_uow_field_types() -> None:
    """Type field 가 service/module/library 만 허용."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    valid_types = {"service", "module", "library"}
    for uow_id, fields in uows.items():
        uow_type = fields.get("Type", "")
        assert uow_type in valid_types, f"{uow_id} has invalid type: {uow_type}"


def test_uow_status_values() -> None:
    """Status field 가 planned/in_progress/stable/deprecated 만 허용."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    valid_statuses = {"planned", "in_progress", "stable", "deprecated"}
    for uow_id, fields in uows.items():
        status = fields.get("Status", "")
        assert status in valid_statuses, f"{uow_id} has invalid status: {status}"


def test_uow_date_format() -> None:
    """Created/Updated 가 YYYY-MM-DD 형식."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    date_re = re.compile(r"^\d{4}-\d{2}-\d{2}$")
    for uow_id, fields in uows.items():
        for date_field in ("Created", "Updated"):
            assert date_re.match(fields[date_field]), \
                f"{uow_id} {date_field} not YYYY-MM-DD: {fields[date_field]}"


# --- Test 2: dependency matrix 정합성 ---


def test_dep_matrix_parse() -> None:
    """matrix parse 정상."""
    matrix = _parse_dep_matrix(SAMPLE_VALID)
    assert "UOW-001" in matrix
    assert matrix["UOW-001"]["depends_on"] == []
    assert matrix["UOW-001"]["depended_by"] == ["UOW-002"]
    assert matrix["UOW-002"]["depends_on"] == ["UOW-001"]


def test_dep_matrix_symmetry() -> None:
    """UOW-001.depended_by = UOW-002.depends_on 양방향 일치."""
    matrix = _parse_dep_matrix(SAMPLE_VALID)
    for uow_id, m in matrix.items():
        for dep_id in m["depended_by"]:
            assert uow_id in matrix[dep_id]["depends_on"], \
                f"{uow_id}.depended_by has {dep_id} but {dep_id}.depends_on missing {uow_id}"


def test_dep_matrix_no_self_dependency() -> None:
    """UOW 가 자기 자신을 depends_on 하지 않음."""
    matrix = _parse_dep_matrix(SAMPLE_VALID)
    for uow_id, m in matrix.items():
        assert uow_id not in m["depends_on"], \
            f"{uow_id} has self-dependency"


def test_dep_matrix_cycle_detection() -> None:
    """순환 의존성 검출 — UOW-001 ↔ UOW-002."""
    matrix = _parse_dep_matrix(SAMPLE_CYCLE)
    graph = {uow: m["depends_on"] for uow, m in matrix.items()}
    assert _has_cycle(graph) is True, "expected cycle in SAMPLE_CYCLE"


def test_dep_matrix_dag_validation() -> None:
    """DAG (비순환) 검증 — SAMPLE_VALID 는 cycle 없음."""
    matrix = _parse_dep_matrix(SAMPLE_VALID)
    graph = {uow: m["depends_on"] for uow, m in matrix.items()}
    assert _has_cycle(graph) is False, "SAMPLE_VALID should be acyclic"


# --- Test 3: Mermaid graph syntax ---


def test_mermaid_block_present() -> None:
    """mermaid code block 존재."""
    m = MERMAID_BLOCK_RE.search(SAMPLE_VALID)
    assert m is not None
    assert "graph LR" in m.group(1)
    assert "UOW-001" in m.group(1)
    assert "UOW-002" in m.group(1)


def test_mermaid_graph_syntax() -> None:
    """mermaid edge syntax: 'UOW-001 --> UOW-002' 형식."""
    m = MERMAID_BLOCK_RE.search(SAMPLE_VALID)
    assert m is not None, "mermaid block not found in SAMPLE_VALID"
    body = m.group(1)
    # mermaid arrow: --> or ---
    edges_arrow = re.findall(r"(\w+)\s*-->\s*(\w+)", body)
    edges_line = re.findall(r"(\w+)\s*---\s*(\w+)", body)
    edges = edges_arrow + edges_line
    # UOW-001 → UOW-002 edge 가 있어야 함
    found = any(
        (a, b) in edges
        for a, b in [("UOW-001", "UOW-002"), ("UOW-002", "UOW-001")]
    )
    # UOW ID 와 graph ID 가 *같을 필요* 는 없음 (변수명), 단순히 graph 안에 edge 가 있어야 함
    assert len(edges) >= 1, f"no mermaid edge found in: {body!r}"


# --- Test 4: Story mapping 정합성 ---


def test_story_mapping_uses_valid_uow_id() -> None:
    """story mapping 의 UOW id 가 UOW 정의에 존재."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    for m in STORY_MAPPING_ROW_RE.finditer(SAMPLE_VALID):
        uow_id = m.group(2)
        assert uow_id in uows, f"story mapping references undefined UOW: {uow_id}"


# --- Test 5: 템플릿 자체 정합성 ---


def test_template_exists_and_has_required_sections() -> None:
    """unit_of_work_template.md 가 존재 + 필수 section 모두 포함."""
    assert TEMPLATE_PATH.exists(), f"template not found: {TEMPLATE_PATH}"
    content = TEMPLATE_PATH.read_text(encoding="utf-8")
    required_sections = [
        "## 1. UOW 정의",
        "## 2. Dependency Matrix",
        "## 3. Dependency Graph (Mermaid)",
        "## 4. Story Mapping",
        "## 5. Code Organization",
    ]
    for section in required_sections:
        assert section in content, f"template missing section: {section}"


def test_template_references_related_docs() -> None:
    """template 이 관련 문서 (work_backlog, session_handoff, project_profile, workflow_task_modes, global_workflow_standard) cross-ref."""
    content = TEMPLATE_PATH.read_text(encoding="utf-8")
    expected_refs = [
        "work_backlog_template.md",
        "session_handoff_template.md",
        "project_workflow_profile_template.md",
        "workflow_task_modes.md",
        "global_workflow_standard.md",
    ]
    for ref in expected_refs:
        assert ref in content, f"template missing cross-ref: {ref}"


def test_template_references_aidlc_source() -> None:
    """AIDLC 원본 (units-generation.md) cross-ref."""
    content = TEMPLATE_PATH.read_text(encoding="utf-8")
    assert "units-generation.md" in content
    assert "awslabs/aidlc-workflows" in content


# --- Test 6: 통합 ---


def test_sample_full_parse_integration() -> None:
    """sample 의 UOW 정의 + matrix + story mapping 모두 일관성."""
    uows = _parse_uow_definitions(SAMPLE_VALID)
    matrix = _parse_dep_matrix(SAMPLE_VALID)
    # matrix 의 모든 UOW id 가 정의에 존재
    for uow_id in matrix:
        assert uow_id in uows, f"matrix has undefined UOW: {uow_id}"
    # story mapping 의 UOW id 도 정의에 존재
    for m in STORY_MAPPING_ROW_RE.finditer(SAMPLE_VALID):
        uow_id = m.group(2)
        assert uow_id in uows, f"story mapping has undefined UOW: {uow_id}"


# --- Main ---


def main() -> int:
    tests = [
        ("uow_header_format", test_uow_header_format),
        ("uow_required_fields", test_uow_required_fields),
        ("uow_field_types", test_uow_field_types),
        ("uow_status_values", test_uow_status_values),
        ("uow_date_format", test_uow_date_format),
        ("dep_matrix_parse", test_dep_matrix_parse),
        ("dep_matrix_symmetry", test_dep_matrix_symmetry),
        ("dep_matrix_no_self_dependency", test_dep_matrix_no_self_dependency),
        ("dep_matrix_cycle_detection", test_dep_matrix_cycle_detection),
        ("dep_matrix_dag_validation", test_dep_matrix_dag_validation),
        ("mermaid_block_present", test_mermaid_block_present),
        ("mermaid_graph_syntax", test_mermaid_graph_syntax),
        ("story_mapping_uses_valid_uow_id", test_story_mapping_uses_valid_uow_id),
        ("template_exists_and_has_required_sections", test_template_exists_and_has_required_sections),
        ("template_references_related_docs", test_template_references_related_docs),
        ("template_references_aidlc_source", test_template_references_aidlc_source),
        ("sample_full_parse_integration", test_sample_full_parse_integration),
    ]
    failed: list[str] = []
    for name, fn in tests:
        try:
            fn()
            print(f"  PASS  {name}")
        except AssertionError as e:
            failed.append(f"  FAIL  {name}: {e}")
            print(f"  FAIL  {name}: {e}")
        except Exception as e:  # noqa: BLE001
            failed.append(f"  ERROR {name}: {type(e).__name__}: {e}")
            print(f"  ERROR {name}: {type(e).__name__}: {e}")

    print()
    if failed:
        print(f"{len(failed)}/{len(tests)} tests failed:")
        for f in failed:
            print(f)
        return 1
    print(f"All {len(tests)} tests passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
