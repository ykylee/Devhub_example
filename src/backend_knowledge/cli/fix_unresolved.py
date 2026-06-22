"""CLI tool: fix_unresolved.py (umbrella doc §3.5.7.4 + M-v0.2.3+).

unresolved cross-link 일괄 resolution CLI. 3 mode confirm workflow:
1. dry-run: recommend 만 출력 (기본값, 변경 ❌)
2. confirm: operator 확인 후 적용 (stdin 'y' 입력 시)
3. auto-apply: confidence ≥ threshold 자동 적용 (변경 ⭕)

Usage:
    python -m backend_knowledge.cli.fix_unresolved \\
        --bundle-dir ./var/bundles/devhub-gitea \\
        --mode dry-run

    python -m backend_knowledge.cli.fix_unresolved \\
        --bundle-dir ./var/bundles/devhub-gitea \\
        --mode auto-apply \\
        --confidence-threshold 0.9
"""

from __future__ import annotations

import argparse
import asyncio
import sys
from pathlib import Path
from typing import Any

from ..config import get_settings
from ..curate.link_resolver import (
    LinkResolver,
    PiMode,
    ResolutionMode,
    UnresolvedLink,
)
from ..logger import configure_logging, get_logger

logger = get_logger(__name__)


def parse_args() -> argparse.Namespace:
    """CLI argument parsing."""
    parser = argparse.ArgumentParser(
        prog="fix_unresolved.py",
        description="DevHub cross-link auto-resolution CLI (§3.5.7.4)",
    )
    parser.add_argument(
        "--bundle-dir",
        type=Path,
        required=True,
        help="Bundle 디렉터리 경로 (e.g., ./var/bundles/devhub-gitea)",
    )
    parser.add_argument(
        "--mode",
        type=str,
        choices=[m.value for m in ResolutionMode],
        default=ResolutionMode.DRY_RUN.value,
        help="Resolution mode (default: dry-run)",
    )
    parser.add_argument(
        "--confidence-threshold",
        type=float,
        default=0.9,
        help="Auto-apply confidence threshold (default: 0.9, range: 0.0~1.0)",
    )
    parser.add_argument(
        "--selected-rank",
        type=int,
        default=1,
        help="confirm mode 시 선택 rank (1=best, 2=second, 3=third, default: 1)",
    )
    parser.add_argument(
        "--pi-mode",
        type=str,
        choices=[m.value for m in PiMode],
        default=None,
        help="Pi LLM invocation mode (PI_MODE env or sdk default)",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help="실제 file 에 link 변경 적용 (default: false, dry-run 만)",
    )
    return parser.parse_args()


async def resolve_all(
    resolver: LinkResolver,
    bundle_dir: Path,
    mode: ResolutionMode,
    confidence_threshold: float,
    selected_rank: int,
    apply: bool,
) -> dict[str, Any]:
    """Bundle 의 모든 unresolved link 처리.

    Returns:
        summary dict: total/found/applied/skipped/error count
    """
    unresolved = await resolver.find_unresolved_links(bundle_dir)
    summary = {
        "total_unresolved": len(unresolved),
        "applied": 0,
        "skipped": 0,
        "error": 0,
    }

    for link in unresolved:
        candidates = await resolver.list_candidates(link.link_target, limit=10)
        if not candidates:
            logger.warning(
                "fix_unresolved.no_candidates",
                source_path=link.source_path,
                link_target=link.link_target,
            )
            summary["skipped"] += 1
            continue

        result = await resolver.resolve(
            link=link,
            candidate_concepts=candidates,
            mode=mode,
            confidence_threshold=confidence_threshold,
            selected_rank=selected_rank,
        )

        if result.applied:
            summary["applied"] += 1
            if apply:
                _apply_resolution(bundle_dir, link, result.selected)
                logger.info(
                    "fix_unresolved.applied",
                    source_path=link.source_path,
                    new_link=result.selected.target_path,
                    confidence=result.selected.confidence,
                )
        else:
            summary["skipped"] += 1
            logger.info(
                "fix_unresolved.skipped",
                source_path=link.source_path,
                mode=mode.value,
                confidence=result.selected.confidence if result.selected else 0.0,
            )

    return summary


def _apply_resolution(
    bundle_dir: Path,
    link: UnresolvedLink,
    selected: Any,
) -> None:
    """Bundle .md file 에 link 변경 적용.

    link_text 를 selected.target_path 로 교체.
    """
    if selected is None:
        return
    source_path = bundle_dir / link.source_path
    if not source_path.exists():
        logger.warning("fix_unresolved.source_not_found", path=str(source_path))
        return
    content = source_path.read_text(encoding="utf-8")
    new_content = content.replace(link.link_text, selected.target_path, 1)
    source_path.write_text(new_content, encoding="utf-8")


def print_summary(summary: dict[str, Any], mode: ResolutionMode) -> None:
    """Summary 출력."""
    print("\n" + "=" * 60)
    print(f"Fix Unresolved Summary (mode={mode.value})")
    print("=" * 60)
    print(f"Total unresolved: {summary['total_unresolved']}")
    print(f"Applied:         {summary['applied']}")
    print(f"Skipped:         {summary['skipped']}")
    print(f"Error:           {summary['error']}")
    print("=" * 60)


def main() -> int:
    """CLI entry point."""
    configure_logging()
    args = parse_args()

    bundle_dir: Path = args.bundle_dir
    if not bundle_dir.exists():
        print(f"ERROR: bundle directory not found: {bundle_dir}", file=sys.stderr)
        return 1

    mode = ResolutionMode(args.mode)
    pi_mode = PiMode(args.pi_mode) if args.pi_mode else None

    resolver = LinkResolver(pi_mode=pi_mode)
    summary = asyncio.run(
        resolve_all(
            resolver=resolver,
            bundle_dir=bundle_dir,
            mode=mode,
            confidence_threshold=args.confidence_threshold,
            selected_rank=args.selected_rank,
            apply=args.apply,
        )
    )

    print_summary(summary, mode)
    return 0 if summary["error"] == 0 else 1


if __name__ == "__main__":
    sys.exit(main())