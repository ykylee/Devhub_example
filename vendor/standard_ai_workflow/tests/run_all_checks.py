#!/usr/bin/env python3
"""v0.7.6+: workflow-source 의 check_*.py smoke test 통합 runner (77 file).

77+ check_*.py 를 한 번에 실행 + 결과 집계. 운영 layer 의 *단일 진입점*.
R-3 단계 분리 (refresh_wiki_memory 와 동일 패턴).

Usage:
    # 전체 77 check 실행
    python3 run_all_checks.py

    # filter (e.g. baselines + wiki 만)
    python3 run_all_checks.py --filter=baselines,wiki

    # JSON 출력 (CI 통합)
    python3 run_all_checks.py --json

    # fail-fast (첫 실패 시 중단)
    python3 run_all_checks.py --fail-fast

    # 특정 dir
    python3 run_all_checks.py --tests-dir=tests

Reference:
- tests/check_baselines_compliance.py (16 test) — v0.7.5
- tests/check_refresh_wiki_memory.py (10 test) — v0.7.5
- tools/refresh_wiki_memory.py (v0.7.5, 1차 출처 / 2차 출처 path 명시 패턴)
- workflow_kit/metadata.py (v0.7.6, [tool.workflow-doctor] config loader)
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import time
from dataclasses import dataclass, field, asdict
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
TESTS_DIR = SOURCE_ROOT / "tests"


@dataclass
class CheckResult:
    """개별 check_*.py 실행 결과."""
    name: str
    path: str
    exit_code: int
    duration_sec: float
    passed: int = 0
    failed: int = 0
    last_line: str = ""
    error_excerpt: str = ""


@dataclass
class RunSummary:
    """전체 run 집계."""
    total: int = 0
    passed: int = 0
    failed: int = 0
    total_duration_sec: float = 0.0
    total_passed_tests: int = 0
    total_failed_tests: int = 0
    results: list[CheckResult] = field(default_factory=list)


def discover_checks(tests_dir: Path, filter_pattern: str | None = None) -> list[Path]:
    """tests/check_*.py glob. --filter 적용 시 name substring match (comma-separated OR)."""
    all_checks = sorted(tests_dir.glob("check_*.py"))
    if not filter_pattern:
        return all_checks
    needles = [n.strip() for n in filter_pattern.split(",") if n.strip()]
    if not needles:
        return all_checks
    return [
        c for c in all_checks
        if any(n in c.stem for n in needles)
    ]


def parse_output(output: str) -> tuple[int, int, str]:
    """check_*.py 의 stdout 에서 (passed, failed, last_line) 추출.

    형식 예:
        All 16 tests passed.
        1/10 tests failed:
        All X tests passed.
    """
    passed = 0
    failed = 0
    last_line = ""
    for line in output.strip().split("\n"):
        line = line.rstrip()
        if not line:
            continue
        last_line = line
        m = re.match(r"^All (\d+) tests passed\.$", line)
        if m:
            passed = int(m.group(1))
            failed = 0
            continue
        m = re.match(r"^(\d+)/(\d+) tests failed:", line)
        if m:
            failed = int(m.group(1))
            # total = int(m.group(2))  # not used; check_* 가 PASS 도 함께 보고
            continue
    return passed, failed, last_line


def run_one(check_path: Path, timeout: int = 60) -> CheckResult:
    """단일 check_*.py 실행 + CheckResult 반환."""
    start = time.time()
    try:
        proc = subprocess.run(
            [sys.executable, str(check_path)],
            capture_output=True, text=True, timeout=timeout,
        )
        duration = time.time() - start
        output = proc.stdout + proc.stderr
        passed, failed, last_line = parse_output(output)
        return CheckResult(
            name=check_path.stem,
            path=str(check_path.relative_to(SOURCE_ROOT)),
            exit_code=proc.returncode,
            duration_sec=round(duration, 2),
            passed=passed,
            failed=failed,
            last_line=last_line,
            error_excerpt="".join(output.split("\n")[-3:])[:200] if proc.returncode != 0 else "",
        )
    except subprocess.TimeoutExpired:
        return CheckResult(
            name=check_path.stem,
            path=str(check_path.relative_to(SOURCE_ROOT)),
            exit_code=-1,
            duration_sec=float(timeout),
            passed=0,
            failed=0,
            last_line="TIMEOUT",
            error_excerpt=f"timeout after {timeout}s",
        )


def aggregate(results: list[CheckResult], duration: float) -> RunSummary:
    """CheckResult list → RunSummary 집계."""
    summary = RunSummary(
        total=len(results),
        passed=sum(1 for r in results if r.exit_code == 0),
        failed=sum(1 for r in results if r.exit_code != 0),
        total_duration_sec=round(duration, 2),
        total_passed_tests=sum(r.passed for r in results),
        total_failed_tests=sum(r.failed for r in results),
        results=results,
    )
    return summary


def print_human(summary: RunSummary) -> None:
    """사람이 읽기 좋은 출력."""
    print(f"=== workflow-source check runner (v0.7.6) ===\n")
    print(f"  total:   {summary.total}")
    print(f"  passed:  {summary.passed} (exit 0)")
    print(f"  failed:  {summary.failed} (exit != 0)")
    print(f"  duration: {summary.total_duration_sec}s")
    print(f"  test pass: {summary.total_passed_tests}")
    print(f"  test fail: {summary.total_failed_tests}\n")
    print(f"  --- per-check ---")
    for r in summary.results:
        status = "PASS" if r.exit_code == 0 else f"FAIL({r.exit_code})"
        line_excerpt = r.last_line[:60] if r.last_line else ""
        print(f"  [{status}] {r.name} ({r.duration_sec}s) {line_excerpt}")
    if summary.failed > 0:
        print(f"\n  --- failures ---")
        for r in summary.results:
            if r.exit_code != 0:
                print(f"  ✗ {r.name}: {r.error_excerpt}")


def main() -> int:
    p = argparse.ArgumentParser(
        description="workflow-source 의 check_*.py 통합 runner (v0.7.6+)",
    )
    p.add_argument("--tests-dir", default=str(TESTS_DIR),
                   help="check_*.py glob dir (default: tests/)")
    p.add_argument("--filter", default=None,
                   help="comma-separated name substring (e.g. baselines,wiki)")
    p.add_argument("--fail-fast", action="store_true",
                   help="첫 실패 시 중단")
    p.add_argument("--timeout", type=int, default=60,
                   help="per-check timeout (default: 60s)")
    p.add_argument("--json", action="store_true",
                   help="JSON 출력 (CI 통합)")
    args = p.parse_args()

    tests_dir = Path(args.tests_dir)
    if not tests_dir.exists():
        print(f"[error] tests dir 부재: {tests_dir}", file=sys.stderr)
        return 2

    checks = discover_checks(tests_dir, args.filter)
    if not checks:
        print(f"[error] check_*.py 0 file 매치: {tests_dir} (filter={args.filter})",
              file=sys.stderr)
        return 2

    start = time.time()
    results: list[CheckResult] = []
    for check_path in checks:
        result = run_one(check_path, timeout=args.timeout)
        results.append(result)
        if args.fail_fast and result.exit_code != 0:
            break
    summary = aggregate(results, time.time() - start)

    if args.json:
        print(json.dumps(asdict(summary), ensure_ascii=False, indent=2))
    else:
        print_human(summary)

    return 0 if summary.failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
