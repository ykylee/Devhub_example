#!/usr/bin/env python3
"""v0.7.23+: wiki 운영 cross-link 1-command wrapper.

L1 raw mirror (state.json / work_backlog.md / wiki/log.md / memory/log.md) +
L2 dense (ai-workflow/wiki/sources/) + L2 stub last_touched 갱신 의 3-step cycle
을 *1 command* 로 묶음. 운영 시 *3번의 별도 invoke* 부담 zero.

**3-step 운영 cycle**:
1. `refresh_wiki_memory.py --refresh-raw --apply` — L1 raw mirror 4 file 갱신
   (git log → release 별 분류 → 1차 출처의 4 file 자동 보강)
2. `emit_wiki_l2_body.py --apply` — L1 raw mirror 본문 발췌 + L2 dense
   (`ai-workflow/wiki/sources/<stem>.md`) emit
3. `refresh_wiki_memory.py --emit-l2 --apply` — L2 stub 의 frontmatter 의
   `last_touched` 갱신 (1차 출처의 in-repo retrieval 일관성 보장)

Usage:
    # 3-step cycle (default)
    python3 wiki_emit.py --apply

    # 1단계만 (raw mirror 만)
    python3 wiki_emit.py --refresh-wiki --apply

    # 2단계만 (L2 dense 만)
    python3 wiki_emit.py --emit-l2 --apply

    # 3단계만 (stub last_touched 만)
    python3 wiki_emit.py --reemit-stubs --apply

    # 1단계 skip (2+3 만)
    python3 wiki_emit.py --skip-1 --apply

    # dry-run (3-step plan preview)
    python3 wiki_emit.py --dry-run

    # project 지정 (multi-project 대비)
    python3 wiki_emit.py --project=standard-ai-workflow --apply

    # L2 dense 본문 cap (default 2000)
    python3 wiki_emit.py --emit-l2 --max-chars=3000 --apply

Reference:
- tools/refresh_wiki_memory.py (3-step 의 1+3)
- tools/emit_wiki_l2_body.py (3-step 의 2)
- v0.7.5 release note (refresh_wiki_memory 정식화)
- v0.7.0 release note (LLM Wiki Layer + L1/L2 분화)
- v0.7.17 release note (in-repo storage isolation)
- v0.7.23 release note (wiki 운영 cross-link 1-command wrapper)
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

# v0.7.17+ in-repo storage: 모든 path 가 REPO_ROOT 기준.
def _detect_repo_root() -> Path:
    """REPO_ROOT 결정 (priority: git rev-parse > cwd 기준 상위 dir 탐색)."""
    try:
        proc = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, timeout=10,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return Path(proc.stdout.strip()).resolve()
    except Exception:
        pass
    return Path.cwd().resolve()


REPO_ROOT = _detect_repo_root()
TOOLS_DIR = REPO_ROOT / "workflow-source" / "tools"
REFRESH_WIKI_MEMORY = TOOLS_DIR / "refresh_wiki_memory.py"
EMIT_WIKI_L2_BODY = TOOLS_DIR / "emit_wiki_l2_body.py"

# 3-step 의 표준 subcommand / flag 매핑
REFRESH_WIKI_SUBCOMMAND = ["--refresh-raw"]
REFRESH_STUBS_SUBCOMMAND = ["--emit-l2"]
EMIT_L2_DENSELY = ["--apply"]  # emit_wiki_l2_body 는 mode=l1 (default) 가 dense 본문 emit


def _run_step(name: str, cmd: list[str], *, dry: bool, timeout: int = 120) -> dict:
    """3-step 중 1 step 실행. dry-run 면 subprocess 호출 안 함.

    Returns:
        {"name": str, "command": list[str], "mode": str, "returncode": int | None,
         "stdout_tail": str, "stderr_tail": str, "skipped": bool}
    """
    result = {
        "name": name,
        "command": cmd,
        "mode": "dry-run" if dry else "apply",
        "returncode": None,
        "stdout_tail": "",
        "stderr_tail": "",
        "skipped": False,
    }
    if dry:
        return result

    proc = subprocess.run(
        cmd,
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        timeout=timeout,
    )
    result["returncode"] = proc.returncode
    if proc.stdout:
        result["stdout_tail"] = proc.stdout.strip().split("\n")[-1]
    if proc.stderr:
        result["stderr_tail"] = proc.stderr.strip().split("\n")[-1]
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument("--refresh-wiki", action="store_true",
                        help="1단계만 (L1 raw mirror 갱신)")
    parser.add_argument("--emit-l2", action="store_true",
                        help="2단계만 (L2 dense 본문 emit)")
    parser.add_argument("--reemit-stubs", action="store_true",
                        help="3단계만 (L2 stub last_touched 갱신)")
    parser.add_argument("--full", action="store_true",
                        help="3-step cycle 전체 (default)")
    parser.add_argument("--skip-1", action="store_true", help="1단계 skip")
    parser.add_argument("--skip-2", action="store_true", help="2단계 skip")
    parser.add_argument("--skip-3", action="store_true", help="3단계 skip")
    parser.add_argument("--project", default="standard-ai-workflow",
                        help="multi-project 대비 (default: standard-ai-workflow)")
    parser.add_argument("--since", default="2026-06-10",
                        help="refresh_wiki_memory 의 git log --since 기준 (default: 2026-06-10)")
    parser.add_argument("--max-chars", type=int, default=2000,
                        help="L2 dense 본문 cap (default: 2000)")
    parser.add_argument("--dry-run", action="store_true", dest="dry_run",
                        help="3-step plan preview (subprocess 호출 0)")
    parser.add_argument("--apply", dest="apply", action="store_true", default=True)
    parser.add_argument("--json", action="store_true", help="JSON 출력 (CI 통합)")
    args = parser.parse_args()

    # 3-step 결정: --full (default) / --refresh-wiki / --emit-l2 / --reemit-stubs / --skip-N
    if not any([args.refresh_wiki, args.emit_l2, args.reemit_stubs]):
        # 아무 sub-step 도 지정 안 함 → --full (3-step cycle)
        run_1 = run_2 = run_3 = True
    else:
        # 1개 이상 지정 → 각각 sub-step 만
        run_1 = args.refresh_wiki
        run_2 = args.emit_l2
        run_3 = args.reemit_stubs

    # --skip-N 가 우선
    if args.skip_1:
        run_1 = False
    if args.skip_2:
        run_2 = False
    if args.skip_3:
        run_3 = False

    # python interpreter
    py = sys.executable

    # 3-step command build
    cmd_1 = [py, str(REFRESH_WIKI_MEMORY)] + REFRESH_WIKI_SUBCOMMAND + [
        "--since", args.since,
        "--project", args.project,
    ]
    if not args.dry_run:
        cmd_1 += ["--apply"]
    cmd_1 += ["--json"]

    cmd_2 = [py, str(EMIT_WIKI_L2_BODY), "--project", args.project,
             "--max-chars", str(args.max_chars)]
    if not args.dry_run:
        cmd_2 += ["--apply"]
    cmd_2 += ["--json"]

    cmd_3 = [py, str(REFRESH_WIKI_MEMORY)] + REFRESH_STUBS_SUBCOMMAND + [
        "--since", args.since,
        "--project", args.project,
    ]
    if not args.dry_run:
        cmd_3 += ["--apply"]
    cmd_3 += ["--json"]

    # 3-step 실행
    results: list[dict] = []
    if run_1:
        results.append(_run_step("1_refresh_raw", cmd_1, dry=args.dry_run))
    if run_2:
        results.append(_run_step("2_emit_l2_dense", cmd_2, dry=args.dry_run))
    if run_3:
        results.append(_run_step("3_reemit_stubs", cmd_3, dry=args.dry_run))

    output = {
        "mode": "dry-run" if args.dry_run else "apply",
        "project": args.project,
        "since": args.since,
        "max_chars": args.max_chars,
        "steps": results,
        "skipped_steps": [
            n for n, run in [("1_refresh_raw", run_1), ("2_emit_l2_dense", run_2), ("3_reemit_stubs", run_3)]
            if not run
        ],
    }
    # apply mode 에서 1+ step fail 시 exit 1
    if not args.dry_run:
        for r in results:
            if r["returncode"] not in (0, None):
                output["error"] = f"step {r['name']} failed (exit {r['returncode']}): {r.get('stderr_tail', '')}"
                if args.json:
                    print(json.dumps(output, ensure_ascii=False, indent=2))
                else:
                    print(f"ERROR: {output['error']}", file=sys.stderr)
                return 1

    if args.json:
        print(json.dumps(output, ensure_ascii=False, indent=2))
    else:
        print(f"Wiki cross-link: {output['mode']} mode, project={args.project}, steps={len(results)}, skipped={len(output['skipped_steps'])}")
        for r in results:
            mode_label = "(dry)" if args.dry_run else ("✓" if r["returncode"] == 0 else f"✗ {r['returncode']}")
            print(f"  [{mode_label}] {r['name']}: {' '.join(r['command'][:5])}...")
    return 0


if __name__ == "__main__":
    sys.exit(main())
