#!/usr/bin/env python3
"""v0.7.17+: wiki in-repo storage isolation smoke test.

외부 vault (`~/wiki/`) 와 본 project 의 연결 완전 차단. 본 project 의 wiki 가
*전부 in-repo* (`ai-workflow/wiki/` + `ai-workflow/memory/active/`) 에서 관리됨.

Test 구성 (9 test):
1. tools/refresh_wiki_memory.py: VAULT_ROOT/RAW_BASE/L2_BASE 가 in-repo path
2. tools/refresh_wiki_memory.py: RAW_FILES 의 state_json/work_backlog/wiki_log/memory_log 가 in-repo
3. tools/refresh_wiki_memory.py: L2_STUBS 의 4 file 이 ai-workflow/wiki/sources/ 안
4. tools/emit_wiki_l2_body.py: VAULT_ROOT/RAW_MIRROR/L2_SOURCES 가 in-repo path
5. tools/emit_wiki_l2_body.py: REPO_ROOT 자동 검출 (git rev-parse)
6. tools/score_wiki_maintainability.py: L2_SOURCES 가 INREPO_WIKI/sources
7. tests/check_refresh_wiki_memory.py: VAULT_ROOT reference 없음
8. tests/check_wiki_drift.py: _raw_mtime 이 in-repo path 만 사용
9. 본 repo 의 ai-workflow/wiki/sources/ dir 존재 + .gitkeep 정합

Reference:
- workflow-source/tools/refresh_wiki_memory.py (v0.7.17 본 release)
- workflow-source/tools/emit_wiki_l2_body.py
- workflow-source/tools/score_wiki_maintainability.py
- workflow-source/tests/check_refresh_wiki_memory.py
- workflow-source/tests/check_wiki_drift.py
- ai-workflow/wiki/sources/ (L2 dense emit target, 본 release 신규)
- memory "wiki in-repo isolation" (cross-project storage SSOT 패턴)
"""

from __future__ import annotations

import importlib.util
import re
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = SOURCE_ROOT.parent
INREPO_WIKI = REPO_ROOT / "ai-workflow" / "wiki"
INREPO_MEMORY = REPO_ROOT / "ai-workflow" / "memory" / "active"


def _import_module(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, str(path))
    mod = importlib.util.module_from_spec(spec)
    sys.modules[name] = mod
    spec.loader.exec_module(mod)
    return mod


def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


# --- Test 1: refresh_wiki_memory 의 in-repo path ---


def test_refresh_wiki_memory_no_vault_root():
    """refresh_wiki_memory.py 에 VAULT_ROOT = Path.home() / 'wiki' 가 *없어야*."""
    src = _read(SOURCE_ROOT / "tools" / "refresh_wiki_memory.py")
    assert "Path.home() / \"wiki\"" not in src, (
        "refresh_wiki_memory.py: VAULT_ROOT = Path.home() / 'wiki' 가 외부 vault 참조. "
        "in-repo path 로 redirect 필요 (v0.7.17)."
    )
    assert "VAULT_ROOT" not in src, (
        "refresh_wiki_memory.py: VAULT_ROOT 변수가 남아 있음. 제거 또는 in-repo path 로."
    )


def test_refresh_wiki_memory_raw_files_in_repo():
    """refresh_wiki_memory.py 의 RAW_FILES 4 file 이 in-repo path."""
    src = _read(SOURCE_ROOT / "tools" / "refresh_wiki_memory.py")
    # RAW_FILES dict 의 value 가 in-repo path 사용 확인
    # in-repo path = "memory/active/state.json" (relative to REPO_ROOT/ai-workflow)
    assert "memory/active/state.json" in src, "state_json path 가 in-repo 가 아님"
    assert "memory/active/work_backlog.md" in src, "work_backlog path 가 in-repo 가 아님"
    assert "wiki/log.md" in src, "wiki_log path 가 in-repo 가 아님"
    assert "memory/log.md" in src, "memory_log path 가 in-repo 가 아님"
    # 외부 vault 의 raw/projects/.../ 또는 ~/wiki/ 가 본 dict 에 없어야
    # (단, docstring 의 reference 에는 mention 가능 → test 는 RAW_FILES 부분만)
    raw_files_section = re.search(
        r"RAW_FILES\s*=\s*\{(.*?)\}",
        src,
        re.DOTALL,
    )
    assert raw_files_section is not None, "RAW_FILES dict 없음"
    assert "raw/projects" not in raw_files_section.group(1), "RAW_FILES 에 외부 raw mirror path 남아 있음"


def test_refresh_wiki_memory_l2_stubs_in_repo():
    """L2_STUBS 의 4 file 이 ai-workflow/wiki/sources/ 안."""
    src = _read(SOURCE_ROOT / "tools" / "refresh_wiki_memory.py")
    l2_section = re.search(
        r"L2_STUBS\s*=\s*\{(.*?)\}",
        src,
        re.DOTALL,
    )
    assert l2_section is not None, "L2_STUBS dict 없음"
    # 4 stub name 이 있어야
    for stub in ["active-state", "active-work-backlog", "active-session-handoff", "wiki-log"]:
        assert stub in l2_section.group(1), f"L2_STUBS 에 {stub} 없음"
    # 외부 wiki/projects/.../sources/ 가 없어야
    assert "wiki/projects" not in l2_section.group(1), "L2_STUBS 에 외부 wiki/projects path 남아 있음"
    assert "Path.home() / \"wiki\"" not in l2_section.group(1), "L2_STUBS 에 VAULT_ROOT 흔적"


# --- Test 2: emit_wiki_l2_body 의 in-repo path ---


def test_emit_wiki_l2_body_no_vault_root():
    """emit_wiki_l2_body.py 에 VAULT_ROOT / RAW_MIRROR / L2_SOURCES 가 in-repo path."""
    src = _read(SOURCE_ROOT / "tools" / "emit_wiki_l2_body.py")
    # in-repo path = REPO_ROOT / "ai-workflow" / "wiki" / "sources"
    assert "REPO_ROOT" in src, "REPO_ROOT 변수 없음"
    # RAW_MIRROR / L2_SOURCES 가 in-repo path 사용 확인
    assert "RAW_MIRROR = L1_BASE / \"wiki\"" in src or "RAW_MIRROR = L1_BASE / 'wiki'" in src, (
        "RAW_MIRROR 가 in-repo path 가 아님"
    )
    assert "L2_SOURCES = L1_BASE / \"wiki\" / \"sources\"" in src, (
        "L2_SOURCES 가 in-repo path 가 아님"
    )
    # 외부 vault 흔적 없어야
    assert "Path.home() / \"wiki\"" not in src, (
        "emit_wiki_l2_body.py: VAULT_ROOT = Path.home() / 'wiki' 가 외부 vault 참조"
    )


def test_emit_wiki_l2_body_repo_root_auto_detect():
    """emit_wiki_l2_body.py 의 _detect_repo_root 가 git rev-parse 사용."""
    src = _read(SOURCE_ROOT / "tools" / "emit_wiki_l2_body.py")
    assert "_detect_repo_root" in src, "_detect_repo_root 함수 없음"
    assert "git rev-parse" in src, "git rev-parse auto-detect 없음"
    assert "show-toplevel" in src, "git rev-parse --show-toplevel 없음"


# --- Test 3: score_wiki_maintainability 의 in-repo path ---


def test_score_wiki_maintainability_l2_in_repo():
    """score_wiki_maintainability.py 의 L2_SOURCES 가 INREPO_WIKI/sources."""
    src = _read(SOURCE_ROOT / "tools" / "score_wiki_maintainability.py")
    assert "L2_SOURCES = INREPO_WIKI / \"sources\"" in src, (
        "L2_SOURCES 가 in-repo path 가 아님"
    )
    # 외부 VAULT_ROOT 흔적 없어야
    assert "Path.home() / \"wiki\"" not in src, (
        "score_wiki_maintainability.py: VAULT_ROOT 가 남아 있음"
    )


# --- Test 4: test 파일의 in-repo path ---


def test_check_refresh_wiki_memory_no_vault_root():
    """tests/check_refresh_wiki_memory.py 에 VAULT_ROOT = Path.home() / 'wiki' 없음."""
    src = _read(SOURCE_ROOT / "tests" / "check_refresh_wiki_memory.py")
    # docstring / comment 의 *legacy* mention 은 OK. active code 에만 없으면.
    # "VAULT_ROOT" 변수가 *active code* 에 남아 있는지 확인.
    # heuristic: import + line 35 (line number may shift) 의 `VAULT_ROOT = ` line
    # 단순하게: `VAULT_ROOT = ` 가 *단일* 발생 (line 35 의 comment 한정).
    active_assignment = re.search(r"^VAULT_ROOT\s*=\s*Path\.home\(\)\s*/\s*\"wiki\"", src, re.MULTILINE)
    assert active_assignment is None, (
        "check_refresh_wiki_memory.py: VAULT_ROOT = Path.home() / 'wiki' 가 active code 에 남아 있음"
    )


def test_check_wiki_drift_raw_mtime_in_repo():
    """tests/check_wiki_drift.py 의 _raw_mtime 이 REPO_ROOT 기반 in-repo path."""
    src = _read(SOURCE_ROOT / "tests" / "check_wiki_drift.py")
    # _raw_mtime 함수의 body 가 REPO_ROOT / raw_path 사용
    fn_match = re.search(
        r"def _raw_mtime.*?(?=\ndef |\nclass |\Z)",
        src,
        re.DOTALL,
    )
    assert fn_match is not None, "_raw_mtime 함수 없음"
    fn_body = fn_match.group(0)
    assert "REPO_ROOT" in fn_body, "_raw_mtime 이 REPO_ROOT 사용 안 함"
    assert "VAULT_ROOT" not in fn_body, "_raw_mtime 이 VAULT_ROOT 사용 (외부 vault 참조)"
    # docstring comment 의 *legacy* mention 은 OK
    assert "Path.home() / \"wiki\"" not in fn_body, (
        "_raw_mtime 이 Path.home() / 'wiki' 사용 (외부 vault)"
    )


# --- Test 5: in-repo dir 존재 + .gitkeep 정합 ---


def test_inrepo_sources_dir_exists():
    """ai-workflow/wiki/sources/ dir + .gitkeep 존재 (v0.7.17+ 신규)."""
    sources_dir = INREPO_WIKI / "sources"
    assert sources_dir.exists(), f"sources dir 없음: {sources_dir}"
    assert sources_dir.is_dir(), f"sources 가 dir 아님: {sources_dir}"
    gitkeep = sources_dir / ".gitkeep"
    assert gitkeep.exists(), f".gitkeep 없음: {gitkeep}"
    # .gitkeep 의 주석이 in-repo storage 정합성 명시
    content = gitkeep.read_text(encoding="utf-8")
    assert "in-repo" in content or "v0.7.17" in content, (
        ".gitkeep 주석에 in-repo storage 정합성 명시 필요"
    )


def test_inrepo_memory_log_exists():
    """ai-workflow/memory/log.md 존재 (refresh 가 갱신할 대상)."""
    log_path = REPO_ROOT / "ai-workflow" / "memory" / "log.md"
    assert log_path.exists(), f"memory/log.md 없음: {log_path}"


def test_inrepo_no_legacy_symlink_or_legacy_path():
    """본 repo 에 legacy symlink (~/repos/standard_ai_workflow_minimax) 흔적 없음."""
    # memory #21 의 REPO_ROOT 4-priority auto-detect 도입 이후 legacy fallback 의
    # _LEGACY_REPO_ROOT 상수만 정공법 (deprecation 경고 + symlink 제거 후).
    # in-repo dir 구조에 ~/repos/ 흔적이 *없어야*.
    legacy = REPO_ROOT / "ai-workflow" / "raw" / "projects"  # in-repo 에 *raw* dir 이면 외부 vault 흔적
    assert not legacy.exists(), f"in-repo 에 raw dir 흔적: {legacy} (외부 vault mirror 가 in-repo 로 들어옴)"


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_refresh_wiki_memory_no_vault_root,
        test_refresh_wiki_memory_raw_files_in_repo,
        test_refresh_wiki_memory_l2_stubs_in_repo,
        test_emit_wiki_l2_body_no_vault_root,
        test_emit_wiki_l2_body_repo_root_auto_detect,
        test_score_wiki_maintainability_l2_in_repo,
        test_check_refresh_wiki_memory_no_vault_root,
        test_check_wiki_drift_raw_mtime_in_repo,
        test_inrepo_sources_dir_exists,
        test_inrepo_memory_log_exists,
        test_inrepo_no_legacy_symlink_or_legacy_path,
    ]
    passed = 0
    failed = 0
    for func in test_funcs:
        try:
            func()
            print(f"  PASS  {func.__name__}")
            passed += 1
        except AssertionError as e:
            print(f"  FAIL  {func.__name__}: {e}")
            failed += 1
        except Exception as e:  # noqa: BLE001
            print(f"  ERROR {func.__name__}: {type(e).__name__}: {e}")
            failed += 1

    print()
    print(f"{passed} pass, {failed} fail")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
