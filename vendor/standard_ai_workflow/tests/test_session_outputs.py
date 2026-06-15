from __future__ import annotations

from workflow_kit.common.session_outputs import build_session_summary


def test_build_session_summary_empty():
    summary = build_session_summary(
        changed_files=[],
        handoff_items=[],
        backlog_items=[],
    )
    assert "Session Handoff Summary" in summary
    assert "### Changed Files" in summary
    assert "### Handoff Items" not in summary
    assert "### Backlog Items" not in summary


def test_build_session_summary_with_files():
    summary = build_session_summary(
        changed_files=["file1.py", "file2.md"],
        handoff_items=[],
        backlog_items=[],
    )
    assert "- `file1.py`" in summary
    assert "- `file2.md`" in summary


def test_build_session_summary_full():
    summary = build_session_summary(
        changed_files=["app.py"],
        handoff_items=["Complete login", "Add logout"],
        backlog_items=["Refactor tests"],
    )
    assert "Session Handoff Summary" in summary
    assert "### Changed Files" in summary
    assert "- `app.py`" in summary
    assert "### Handoff Items" in summary
    assert "- Complete login" in summary
    assert "- Add logout" in summary
    assert "### Backlog Items" in summary
    assert "- Refactor tests" in summary
