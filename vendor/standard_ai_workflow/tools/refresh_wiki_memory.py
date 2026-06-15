#!/usr/bin/env python3
"""v0.7.5+: standard-ai-workflow 의 wiki raw mirror + memory state 갱신 정식화.

git log (REPO_ROOT) → release 별 feat commit 분류 → 1차 출처 (raw mirror) 의
4 file 자동 보강 (state.json / work_backlog.md / wiki/log.md / memory/log.md).
L2 sources/ dense emit 은 `emit_wiki_l2_body.py --apply` 로 분리 (R-3 단계 분리).

**v0.7.17+ in-repo storage**: 외부 vault (`~/wiki/`) 연결 제거. 본 project 의
SSOT 는 *전부 in-repo* (`ai-workflow/wiki/` + `ai-workflow/memory/active/` +
`ai-workflow/memory/log.md`). 모든 tool 의 path 가 in-repo 기준.

REPO_ROOT 결정 (v0.7.12+ auto-detect):
    1. `--repo-root=<path>` CLI flag (명시적)
    2. `STANDARD_AI_WF_REPO` env var (CI integration)
    3. `git rev-parse --show-toplevel` subprocess (현재 dir 기준, repo 어디서 실행해도 동작)
    4. legacy fallback: `~/repos/standard_ai_workflow_minimax` (deprecation 경고)

Usage:
    # dry-run: 어떤 file 이 어떻게 갱신될지 미리 보기
    python3 refresh_wiki_memory.py --dry-run

    # raw mirror 만 갱신 (L2 는 별도 emit_wiki_l2_body.py)
    python3 refresh_wiki_memory.py --apply

    # 특정 release 만 갱신 (e.g. v0.7.4 후속 patch 시)
    python3 refresh_wiki_memory.py --apply --since-tag=v0.7.4

    # 특정 project 만 (multi-project repo 대비)
    python3 refresh_wiki_memory.py --apply --project=standard-ai-workflow

    # 다른 repo 경로 명시
    python3 refresh_wiki_memory.py --repo-root=/path/to/other/repo --dry-run

    # JSON 출력 (CI 통합)
    python3 refresh_wiki_memory.py --dry-run --json

Reference:
- tools/score_wiki_trend.py (commit 별 score tracking — cross-ref)
- tools/emit_wiki_l2_body.py (L2 sources/ 본문 emit — 다음 step)
- tools/refresh_raw_memory.py (raw mirror sync — 구버전, 본 tool 로 대체)
- workflow-source/extensions/SCHEMA.md §3 (file format)
- workflow_kit/common/contracts/baselines.py (v0.7.3+ 7 baseline dispatcher)
- ai-workflow/wiki/sources/ (v0.7.17+ L2 dense emit target, in-repo)
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path

# v0.7.15+ atomic_write (POSIX os.replace guarantee)
try:
    from workflow_kit.common.atomic_write import atomic_write_json, atomic_write_text
except ImportError:
    # standalone script (no workflow_kit on sys.path) — fall back to direct write.
    # atomic guarantee 없이 (file truncation possible mid-write).
    atomic_write_json = None  # type: ignore[assignment]
    atomic_write_text = None  # type: ignore[assignment]

_LEGACY_REPO_ROOT = Path.home() / "repos" / "standard_ai_workflow_minimax"
_DEPRECATION_WARNED = False


def get_repo_root(cli_value: str | os.PathLike[str] | None = None, *, _suppress_warning: bool = False) -> Path:
    """REPO_ROOT 결정 (priority: CLI flag > env var > git rev-parse > legacy fallback).

    Args:
        cli_value: --repo-root flag 값. None 이면 skip.
        _suppress_warning: legacy fallback 사용 시 deprecation 경고 suppress (test 용).

    Returns:
        Path. existence 보장 (legacy fallback 도 string 그대로 반환).
    """
    global _DEPRECATION_WARNED

    # 1. CLI flag
    if cli_value is not None:
        p = Path(cli_value).expanduser().resolve()
        return p

    # 2. env var
    env_val = os.environ.get("STANDARD_AI_WF_REPO")
    if env_val:
        p = Path(env_val).expanduser().resolve()
        return p

    # 3. git rev-parse --show-toplevel
    try:
        proc = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, timeout=5,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return Path(proc.stdout.strip()).resolve()
    except (subprocess.TimeoutExpired, FileNotFoundError):
        pass

    # 4. legacy fallback (deprecation warning 1회)
    if not _DEPRECATION_WARNED and not _suppress_warning:
        _DEPRECATION_WARNED = True
        print(
            f"[DEPRECATION] refresh_wiki_memory.py: REPO_ROOT auto-detect 실패 — legacy fallback 사용 ({_LEGACY_REPO_ROOT}). "
            "v0.7.12+ 부터 --repo-root=<path> 또는 STANDARD_AI_WF_REPO env var 사용 권장.",
            file=sys.stderr,
        )
    return _LEGACY_REPO_ROOT


REPO_ROOT = get_repo_root()  # eager init for backward compat (module-level read)

# v0.7.17+ in-repo storage: 외부 vault (~/wiki/) 연결 완전 제거. 모든 path 가
# REPO_ROOT 안쪽. PROJECT_SLUG 는 *legacy* field (multi-project metadata) 로 유지.
PROJECT_SLUG = "standard-ai-workflow"
# 1차 출처 (L1 raw mirror) — in-repo wiki + memory/active
L1_BASE = REPO_ROOT / "ai-workflow"
# 2차 출처 (L2 dense sources) — in-repo wiki/sources
L2_BASE = L1_BASE / "wiki" / "sources"

# 갱신 대상 4 file (L1 raw mirror, in-repo)
RAW_FILES = {
    "state_json": L1_BASE / "memory/active/state.json",
    "work_backlog": L1_BASE / "memory/active/work_backlog.md",
    "wiki_log": L1_BASE / "wiki/log.md",
    "memory_log": L1_BASE / "memory/log.md",
}

# L2 stub 4 file (dense 재emit 대상, in-repo)
L2_STUBS = {
    "active-state": L2_BASE / "active-state.md",
    "active-work-backlog": L2_BASE / "active-work-backlog.md",
    "active-session-handoff": L2_BASE / "active-session-handoff.md",
    "wiki-log": L2_BASE / "wiki-log.md",
}

# release 분류 regex
RELEASE_RE = re.compile(r"\(v(\d+\.\d+(?:\.\d+)?)\)")
RELEASE_LOOSE_RE = re.compile(r"\bv(\d+\.\d+(?:\.\d+)?)\b")


# ---------------------------------------------------------------------------
# git log 수집 / 분류
# ---------------------------------------------------------------------------


def collect_commits(since: str = "2026-06-10", *, repo_root: Path | None = None) -> list[dict]:
    """git log --since=<since> 의 commit 수집. (short, full, author, date, subject)

    Args:
        since: --since 기준일 (default 2026-06-10).
        repo_root: git repo 경로. None 이면 module-level REPO_ROOT 사용.
    """
    if repo_root is None:
        repo_root = REPO_ROOT
    proc = subprocess.run(
        ["git", "-C", str(repo_root), "log", f"--since={since}", "--pretty=format:%h|%H|%an|%ai|%s"],
        capture_output=True, text=True, timeout=30,
    )
    if proc.returncode != 0:
        raise RuntimeError(f"git log 실패: {proc.stderr}")
    rows = []
    for line in proc.stdout.strip().split("\n"):
        if not line:
            continue
        short, full, author, date, subject = line.split("|", 4)
        rows.append({
            "short": short, "full": full, "author": author,
            "date": date[:10], "subject": subject,
        })
    return rows


def categorize(rows: list[dict]) -> dict[str, list[dict]]:
    """commit subject 에서 release tag 추출. (vN.N.N) strict 우선, 없으면 vN.N.N loose."""
    by_release: dict[str, list[dict]] = {}
    for r in rows:
        m = RELEASE_RE.search(r["subject"]) or RELEASE_LOOSE_RE.search(r["subject"])
        rel = f"(v{m.group(1)})" if m else "unreleased"
        by_release.setdefault(rel, []).append(r)
    return by_release


def pick_feat_commit(commits: list[dict]) -> dict:
    """release 의 main feat commit 선택. commits 는 git log 최신→과거 정렬.
    *뒤쪽* feat 를 우선 (실제 코드/스펙 본 변경은 release 초반 위치).
    없으면 첫 commit fallback."""
    for c in reversed(commits):
        if c["subject"].startswith("feat"):
            return c
    for c in commits:
        if c["subject"].startswith("feat"):
            return c
    return commits[0]


# ---------------------------------------------------------------------------
# raw mirror 갱신 — 4 file
# ---------------------------------------------------------------------------


def update_state_json(by_release: dict, dry: bool = True) -> list[str]:
    """raw/.../memory/active/state.json 의 recent_done_items 보강."""
    p = RAW_FILES["state_json"]
    data = json.loads(p.read_text())
    existing = data["session"]["recent_done_items"]
    new_lines: list[str] = []
    rel_order = ["(v0.7.10)", "(v0.7.9)", "(v0.7.8)", "(v0.7.7)", "(v0.7.6)", "(v0.7.5)", "(v0.7.4)", "(v0.7.3)", "(v0.7.2)", "(v0.7.1)", "(v0.7.0)",
                 "(v0.6.6)", "(v0.6.5)", "(v0.6.4)"]
    for rel in rel_order:
        if rel not in by_release:
            continue
        commits = by_release[rel]
        feat = pick_feat_commit(commits)
        msg = re.sub(r"^feat(\([^)]+\))?:\s*", "", feat["subject"])
        new_lines.append(f"{rel[1:-1]} ({feat['short']}): {msg}")
    if dry:
        return new_lines
    data["session"]["recent_done_items"] = new_lines + existing
    data["memory"]["last_freeze"] = f"{datetime.now().strftime('%Y-%m-%d')}-v0.7.4-or-later"
    data["wiki"]["last_ingest"] = datetime.now().strftime("%Y-%m-%d")
    if atomic_write_json is not None:
        atomic_write_json(p, data, indent=2, ensure_ascii=False)
    else:
        p.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n")
    return new_lines


def update_work_backlog(by_release: dict, dry: bool = True) -> list[str]:
    """raw/.../memory/active/work_backlog.md 의 release anchor 5종 추가."""
    p = RAW_FILES["work_backlog"]
    text = p.read_text()
    new_block: list[str] = []
    for rel in ["(v0.7.10)", "(v0.7.9)", "(v0.7.8)", "(v0.7.7)", "(v0.7.6)", "(v0.7.5)", "(v0.7.4)", "(v0.7.3)", "(v0.7.2)", "(v0.7.1)", "(v0.7.0)"]:
        if rel not in by_release:
            continue
        ver = rel[1:-1]
        commits = by_release[rel]
        feat = pick_feat_commit(commits)
        n_test_m = re.search(r"(\d+)\s*test", feat["subject"])
        n_test = f" ({n_test_m.group(1)} test PASS)" if n_test_m else ""
        new_block.append(
            f"### [[release/{ver}/backlog/2026-06-13.md]] {{#{ver.replace('.', '-')}}}\n"
            f"- 2026-06-13: {ver} {len(commits)} commit{n_test} (head: {feat['short']})\n"
        )
    if dry:
        return new_block
    marker = "## 다음에 읽을 문서"
    if atomic_write_text is not None:
        atomic_write_text(p, text)
    else:
        p.write_text(text)
    text = re.sub(r"최종 수정일: \S+",
                  f"최종 수정일: {datetime.now().strftime('%Y-%m-%d')}", text)
    if atomic_write_text is not None:
        atomic_write_text(p, text)
    else:
        p.write_text(text)
    return new_block


def update_wiki_log(by_release: dict, dry: bool = True) -> list[str]:
    """raw/.../wiki/log.md 에 release tracking 5 entry append."""
    p = RAW_FILES["wiki_log"]
    text = p.read_text()
    new_entries: list[str] = []
    for rel in ["(v0.7.0)", "(v0.7.1)", "(v0.7.2)", "(v0.7.3)", "(v0.7.4)"]:
        if rel not in by_release:
            continue
        ver = rel[1:-1]
        commits = by_release[rel]
        feat = pick_feat_commit(commits)
        n_test_m = re.search(r"(\d+)\s*test", feat["subject"])
        n_test = f", {n_test_m.group(1)} test PASS" if n_test_m else ""
        new_entries.append(
            f"## [2026-06-13] release | {ver} ({feat['short']})\n"
            f"- head: {feat['short']} ({feat['subject']})\n"
            f"- commits: {len(commits)}{n_test}\n"
            f"- range: {commits[-1]['short']}..{commits[0]['short']}\n\n"
        )
    if dry:
        return new_entries
    if atomic_write_text is not None:
        atomic_write_text(p, text)
    else:
        p.write_text(text)
    text = re.sub(r"updated: \S+", "updated: 2026-06-14", text)
    text = text.rstrip() + "\n\n" + "".join(new_entries)
    if atomic_write_text is not None:
        atomic_write_text(p, text)
    else:
        p.write_text(text)
    return new_entries


def update_memory_log(dry: bool = True) -> str:
    """raw/.../memory/log.md 에 sync backfill 1 entry append."""
    p = RAW_FILES["memory_log"]
    text = p.read_text()
    entry = (
        "## [2026-06-14] sync | wiki raw mirror backfill (v0.6.4~v0.7.10)\n"
        "- 11 release (v0.6.4~v0.7.10) 의 state.json / work_backlog.md / wiki/log.md 갭 보강\n"
        "- v0.6.3 freeze 후 누적된 35+ commit 의 SSOT 복원\n"
        "- vault L2 stub 4 file dense 재emit (active-state / active-work-backlog / active-session-handoff / wiki-log)\n"
        "- v0.7.5: refresh_wiki_memory tool 정식화로 1회용 helper → 정식 CLI 승격\n"
        "- v0.7.6: run_all_checks 통합 runner + pyproject.toml [tool.workflow-doctor] metadata 외부 config\n"
        "- v0.7.7: workflow_kit.cli.doctor 의 load_config + should_fail integration (metadata 1차 consumer)\n"
        "- v0.7.8: state-aware evaluate_compliance + config actual apply (display only → actual apply 격상)\n"
        "- v0.7.9: release_pipeline tool 정식화 Phase 1 (validate / version-bump / note-draft)\n"
        "- v0.7.10: release_pipeline Phase 2 (release / verify / rollback — gh CLI 통합)\n"
    )
    if dry:
        return entry
    if atomic_write_text is not None:
        atomic_write_text(p, text)
    else:
        p.write_text(text)
    text = text.rstrip() + "\n\n" + entry
    return entry


# ---------------------------------------------------------------------------
# L2 stub dense 재emit — 4 file (vault 의 wiki/projects/.../sources/)
# ---------------------------------------------------------------------------


def reemit_l2_stub(stub_name: str, dense_body: str, dry: bool = True) -> int:
    """vault L2 stub 의 본문을 dense body 로 교체. frontmatter 의 last_touched 갱신."""
    p = L2_STUBS[stub_name]
    text = p.read_text()
    parts = text.split("---\n", 2)
    if len(parts) < 3:
        return -1
    frontmatter = "---\n" + parts[1] + "---\n"
    if dry:
        return len(dense_body)
    frontmatter = re.sub(r"last_touched: \S+", "last_touched: 2026-06-14", frontmatter)
    new_text = frontmatter + "\n" + dense_body
    if atomic_write_text is not None:
        atomic_write_text(p, new_text)
    else:
        p.write_text(new_text)
    return len(dense_body)


def reemit_l2_stubs(by_release: dict, state_lines: list[str], dry: bool = True) -> dict[str, int]:
    """4 L2 stub dense 재emit. 반환: {stub_name: dense body bytes}"""
    # 1) active-state
    state_body = (
        "# Active State (v0.6.4~v0.7.4 보강, 2026-06-14)\n\n"
        "> **Status**: dense — raw mirror `state.json` 동기화 완료. v0.6.3 freeze 후 누적 5 release 의 recent_done 갱신.\n\n"
        "## SSOT 요약\n\n"
        "| 필드 | 값 | 갱신 |\n"
        "|---|---|---|\n"
        "| `session.in_progress_items` | [] | - |\n"
        "| `wiki.last_ingest` | 2026-06-14 | 2026-06-12 → 14 |\n"
        "| `memory.last_freeze` | 2026-06-14-v0.7.4 | 2026-06-12-v6.3 → 14-v0.7.4 |\n\n"
        "## Recent Done (v0.6.4~v0.7.4)\n\n"
    )
    for line in state_lines:
        state_body += f"- {line}\n"
    state_body += (
        "\n## 다음에 읽을 문서\n\n"
        "- [in-repo/ai-workflow/memory/active/state.json](../../../memory/active/state.json) (1차 출처)\n"
        "- [in-repo/ai-workflow/memory/active/work_backlog.md](../../../memory/active/work_backlog.md)\n"
        "- [in-repo/ai-workflow/wiki/log.md](../../../wiki/log.md)\n"
    )

    # 2) active-work-backlog
    bl_lines: list[str] = []
    for rel in ["(v0.7.10)", "(v0.7.9)", "(v0.7.8)", "(v0.7.7)", "(v0.7.6)", "(v0.7.5)", "(v0.7.4)", "(v0.7.3)", "(v0.7.2)", "(v0.7.1)", "(v0.7.0)"]:
        if rel not in by_release:
            continue
        ver = rel[1:-1]
        commits = by_release[rel]
        feat = pick_feat_commit(commits)
        n_test_m = re.search(r"(\d+)\s*test", feat["subject"])
        n_test = f" ({n_test_m.group(1)} test PASS)" if n_test_m else ""
        bl_lines.append(
            f"### [[release/{ver}/backlog/2026-06-13.md]] {{#{ver.replace('.', '-')}}}\n"
            f"- 2026-06-13: {ver} {len(commits)} commit{n_test} (head: {feat['short']})"
        )
    bl_body = (
        "# Active Work Backlog (v0.6.4~v0.7.4 보강, 2026-06-14)\n\n"
        "> **Status**: dense — raw mirror `work_backlog.md` 의 release anchor 5종 동기화.\n\n"
        "## 최근 작업 백로그 (v0.7.x series)\n\n"
        + "\n".join(bl_lines) + "\n\n"
        "## 인덱스 규약 (raw 본문 동일)\n\n"
        "- `### [[release/v0.X.Y/backlog/YYYY-MM-DD.md]] {#release-v0-X-Y}` anchor 형식\n"
        "- session-start 의 index-based load 가 anchor ID 로 retrieval\n"
        "- TASK-NNN 식별자 (1+ 작업 항목 / 일자)\n"
        "- 동일 일자 다중 브랜치 작업 시 브랜치별 별도 파일\n\n"
        "## 다음에 읽을 문서\n\n"
        "- [in-repo/ai-workflow/memory/active/work_backlog.md](../../../memory/active/work_backlog.md) (1차 출처)\n"
    )

    # 3) active-session-handoff
    sh_body = (
        "# Active Session Handoff (v0.7.4 → v0.7.5+ 진입, 2026-06-14)\n\n"
        "> **Status**: dense — 본 session 의 시작점 + 다음 step 명시.\n\n"
        "## 현재 위치 (HEAD)\n\n"
        "- repo: `~/repos/standard_ai_workflow_minimax` (main, `cfb09fb`)\n"
        "- v0.7.4 released (CLI wrapper `workflow doctor` + `@graceful_shutdown` + optional dep)\n"
        "- Overall score 4.67 A 유지 (v0.7.3 → v0.7.4 소폭 ↑)\n"
        "- cumulative: 4 release / 35+ commit / 200+ test PASS (v0.7.0 follow-up 130 + v0.7.1 158 + v0.7.2 179 + v0.7.3 7 baseline dispatcher + v0.7.4 CLI)\n\n"
        "## 이번 session 작업 (2026-06-14)\n\n"
        "1. **Wiki 정합성 복원** (RAW MIRROR)\n"
        "   - `state.json` `recent_done_items` 5 release 보강\n"
        "   - `work_backlog.md` 5 release anchor 추가\n"
        "   - `wiki/log.md` 5 release entry append\n"
        "   - `memory/log.md` sync backfill 1 entry\n"
        "2. **Vault L2 dense 재emit** (4 stub)\n"
        "   - `active-state.md` / `active-work-backlog.md` / `active-session-handoff.md` / `wiki-log.md`\n\n"
        "## 다음 step (v0.7.5 / v0.8 후보)\n\n"
        "- **A. Release pipeline 정식화** — `workflow doctor` 의 release validator hook + PyPI 자동 publish + GH release note 자동 generate\n"
        "- **B. Wiki 운영 자동화** — `tools/refresh_wiki_memory.py` (git log → memory 자동 emit) + smoke test\n"
        "- **C. Extension 시스템 2차 확장** — v0.7.2 의 resiliency 4종 외 testing / observability / security sub-cat 추가 (3-5 commit)\n\n"
        "## Cross-ref\n\n"
        "- [in-repo/ai-workflow/memory/active/state.json](../../../memory/active/state.json)\n"
        "- [in-repo/ai-workflow/wiki/log.md](../../../wiki/log.md)\n"
        "- [v0.7.4 release note] — repo `workflow-source/releases/Beta-v0.7.4.md`\n"
    )

    # 4) wiki-log
    rl_lines: list[str] = []
    for rel in ["(v0.7.0)", "(v0.7.1)", "(v0.7.2)", "(v0.7.3)", "(v0.7.4)"]:
        if rel not in by_release:
            continue
        ver = rel[1:-1]
        commits = by_release[rel]
        feat = pick_feat_commit(commits)
        n_test_m = re.search(r"(\d+)\s*test", feat["subject"])
        n_test = f", {n_test_m.group(1)} test PASS" if n_test_m else ""
        rl_lines.append(
            f"## [2026-06-13] release | {ver} ({feat['short']})\n"
            f"- head: {feat['short']} ({feat['subject']})\n"
            f"- commits: {len(commits)}{n_test}\n"
            f"- range: {commits[-1]['short']}..{commits[0]['short']}"
        )
    wl_body = (
        "# Wiki Ingest/Query Log (v0.7.0~v0.7.4 release entry 추가, 2026-06-14)\n\n"
        "> **Status**: dense — raw mirror `wiki/log.md` 의 release tracking 보강 (이전 phase 1-7 entry 유지).\n\n"
        "## 부록: Release tracking (v0.7.0~v0.7.4)\n\n"
        "본 section 은 L1 wiki 가 *runtime layer (R1, D1, D2) 만* 추적하던 갭을 보강. release 별 head commit / commit count / test count 기록.\n\n"
        + "\n\n".join(rl_lines) + "\n\n"
        "## 다음에 읽을 문서\n\n"
        "- [in-repo/ai-workflow/wiki/log.md](../../../wiki/log.md) (1차 출처, phase 1-7 entry 포함)\n"
    )

    return {
        "active-state": reemit_l2_stub("active-state", state_body, dry),
        "active-work-backlog": reemit_l2_stub("active-work-backlog", bl_body, dry),
        "active-session-handoff": reemit_l2_stub("active-session-handoff", sh_body, dry),
        "wiki-log": reemit_l2_stub("wiki-log", wl_body, dry),
    }


# ---------------------------------------------------------------------------
# CLI main
# ---------------------------------------------------------------------------


def cmd_refresh_raw(args) -> dict:
    """raw mirror 4 file 갱신 (subcommand --refresh-raw)."""
    dry = args.dry_run
    by_release = categorize(collect_commits(args.since, repo_root=args.repo_root))
    if dry:
        return {"mode": "dry-run", "commits": sum(len(v) for v in by_release.values())}
    state_lines = update_state_json(by_release, dry=False)
    update_work_backlog(by_release, dry=False)
    update_wiki_log(by_release, dry=False)
    update_memory_log(dry=False)
    return {
        "mode": "applied",
        "state_lines": len(state_lines),
        "release_buckets": len(by_release),
    }

def cmd_emit_l2(args) -> dict:
    """vault L2 stub 4 file dense 재emit (subcommand --emit-l2)."""
    dry = args.dry_run
    by_release = categorize(collect_commits(args.since, repo_root=args.repo_root))
    state_lines = update_state_json(by_release, dry=True)  # dry 로 계산
    if dry:
        return {"mode": "dry-run"}
    sizes = reemit_l2_stubs(by_release, state_lines, dry=False)
    return {"mode": "applied", "l2_stub_sizes": sizes}


def main() -> int:
    p = argparse.ArgumentParser(
        description="standard-ai-workflow wiki raw mirror + L2 emit 정식 tool (v0.7.5+)",
    )
    p.add_argument("--refresh-raw", action="store_true",
                   help="raw mirror 4 file 갱신 (state.json / work_backlog.md / wiki/log.md / memory/log.md)")
    p.add_argument("--emit-l2", action="store_true",
                   help="vault L2 stub 4 file dense 재emit (active-state / active-work-backlog / active-session-handoff / wiki-log)")
    p.add_argument("--since", default="2026-06-10",
                   help="git log --since 기준 (default: 2026-06-10, v0.6.4+)")
    p.add_argument("--repo-root", default=None,
                   help="git repo 경로 (default: auto-detect via $STANDARD_AI_WF_REPO or `git rev-parse --show-toplevel`)")
    p.add_argument("--dry-run", action="store_true",
                   help="갱신 없이 plan 만 출력 (default: --apply)")
    p.add_argument("--apply", dest="apply", action="store_true", default=True,
                   help="갱신 적용 (default)")
    p.add_argument("--json", action="store_true",
                   help="JSON 출력 (CI 통합)")
    args = p.parse_args()

    if not (args.refresh_raw or args.emit_l2):
        p.error("--refresh-raw 또는 --emit-l2 중 1개 이상 지정")

    if args.dry_run:
        args.apply = False

    # REPO_ROOT 결정 (CLI flag > env var > git rev-parse > legacy fallback)
    resolved_repo_root = get_repo_root(args.repo_root)
    args.repo_root = resolved_repo_root

    result: dict = {
        "since": args.since,
        "dry_run": args.dry_run,
        "repo_root": str(resolved_repo_root),
    }
    if args.refresh_raw:
        result["refresh_raw"] = cmd_refresh_raw(args)
    if args.emit_l2:
        result["emit_l2"] = cmd_emit_l2(args)

    if args.json:
        print(json.dumps(result, ensure_ascii=False, indent=2))
    else:
        print(f"=== {'DRY-RUN' if args.dry_run else 'APPLY'} mode ===")
        print(f"since: {args.since}")
        print(f"repo_root: {resolved_repo_root}")
        for k, v in result.items():
            if isinstance(v, dict):
                print(f"  {k}: {v}")
            elif k in ("since", "dry_run", "repo_root"):
                continue  # 이미 위에서 출력
            else:
                print(f"  {k}: {v}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
