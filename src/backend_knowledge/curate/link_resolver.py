"""Cross-link auto-resolution (umbrella doc §3.5.7 + M-v0.2.3+).

Pi LLM 기반 unresolved cross-link 자동 recommend + 3 mode confirm workflow:
1. dry-run: recommend 만 (변경 ❌)
2. confirm: operator 가 1 row 선택 + 적용
3. auto-apply: confidence ≥ threshold (default 0.9) 자동 적용

Pi SDK/RPC mode (§10.3 정공법):
- SDK mode: M-v0.2.3+ default (Pi coding agent SDK 사용)
- RPC mode: M-v0.2.3+ production option (Pi RPC server 호출)

Mode 자동 fallback: PI_MODE 환경변수 미설정 시 SDK mode default.
"""

from __future__ import annotations

import os
from datetime import datetime, timezone
from enum import Enum
from pathlib import Path
from typing import Any

from pydantic import BaseModel, Field

from ..config import get_settings
from ..logger import get_logger
from ..okf.cross_link import CrossLink, extract_cross_links

logger = get_logger(__name__)


# ----------------------------------------------------------------------
# Pydantic models
# ----------------------------------------------------------------------


class PiMode(str, Enum):
    """Pi LLM invocation mode (§10.3)."""

    SDK = "sdk"
    RPC = "rpc"


class ResolutionMode(str, Enum):
    """3 mode confirm workflow (§3.5.7.4)."""

    DRY_RUN = "dry-run"
    CONFIRM = "confirm"
    AUTO_APPLY = "auto-apply"


class LinkRecommendation(BaseModel):
    """Pi LLM recommendation (1 row per candidate).

    §3.5.7.2 j2 prompt template output format:
    - target_slug: recommended concept slug
    - target_path: recommended concept path ({bundle}/{category}/{slug}.md)
    - reason: 왜 이 candidate 가 맞는지 (1-2 문장)
    - confidence: 0.0 ~ 1.0 (Pi LLM self-assessment)
    """

    rank: int = Field(..., description="1=best, 2=second, 3=third")
    target_slug: str
    target_path: str
    reason: str
    confidence: float = Field(..., ge=0.0, le=1.0)


class UnresolvedLink(BaseModel):
    """Unresolved cross-link (extracted from bundle .md file)."""

    source_path: str  # {bundle}/{category}/{slug}.md
    link_text: str  # raw link text (e.g., "../other-bundle/category/slug.md")
    link_target: str  # parsed target path
    context: str  # ±2 lines context around the link


class ResolutionResult(BaseModel):
    """resolve_unresolved_link() 결과.

    §3.5.7.5 audit log event pi_link_resolve.applied 의 payload.
    """

    source_path: str
    original_link: str
    selected: LinkRecommendation | None
    alternatives: list[LinkRecommendation]
    mode: ResolutionMode
    confidence_threshold: float
    applied: bool
    timestamp: str


# ----------------------------------------------------------------------
# j2 prompt template rendering
# ----------------------------------------------------------------------


J2_TEMPLATE = """You are resolving an unresolved cross-link in a DevHub knowledge bundle.

## Source concept (where the link appears)
- Path: {{ source_path }}
- Context (2 lines before/after the link):
```
{{ context }}
```

## Unresolved link
- Text: {{ link_text }}
- Parsed target: {{ link_target }}

## Candidate concepts (top 10 by name similarity)
{% for c in candidates %}
- Path: {{ c.path }}
  Title: {{ c.title }}
  Type: {{ c.type }}
{% endfor %}

## Task
Recommend the top 3 candidate concepts that best resolve this unresolved link.
For each, provide:
- target_slug (concept slug)
- target_path (full path)
- reason (1-2 sentences explaining the match)
- confidence (0.0-1.0, your self-assessment)

Return JSON in this exact format:
```json
{
  "recommendations": [
    {
      "rank": 1,
      "target_slug": "...",
      "target_path": "...",
      "reason": "...",
      "confidence": 0.92
    },
    {
      "rank": 2,
      "target_slug": "...",
      "target_path": "...",
      "reason": "...",
      "confidence": 0.78
    },
    {
      "rank": 3,
      "target_slug": "...",
      "target_path": "...",
      "reason": "...",
      "confidence": 0.65
    }
  ]
}
```
"""


def render_prompt(
    source_path: str,
    link_text: str,
    link_target: str,
    context: str,
    candidates: list[dict[str, Any]],
) -> str:
    """j2 prompt template 렌더링 (Pi LLM 입력).

    §3.5.7.2 정공법:
    - input: unresolved link context ±2 lines
    - output: 3 row recommendation + reason + confidence 0~1
    """
    template = J2_TEMPLATE
    candidates_block = "\n".join(
        f"- Path: {c.get('path', '?')}\n  Title: {c.get('title', '?')}\n  Type: {c.get('type', '?')}"
        for c in candidates[:10]
    )
    return (
        template.replace("{{ source_path }}", source_path)
        .replace("{{ context }}", context)
        .replace("{{ link_text }}", link_text)
        .replace("{{ link_target }}", link_target)
        .replace(
            "{% for c in candidates %}\n{% endfor %}",
            candidates_block,
        )
    )


# ----------------------------------------------------------------------
# LinkResolver core
# ----------------------------------------------------------------------


class LinkResolver:
    """Cross-link auto-resolution core (§3.5.7 + M-v0.2.3+).

    Usage:
        resolver = LinkResolver()
        unresolved = await resolver.find_unresolved_links(bundle_dir)
        for link in unresolved:
            result = await resolver.resolve(
                link=link,
                candidate_concepts=await resolver.list_candidates(link.link_target),
                mode=ResolutionMode.AUTO_APPLY,
                confidence_threshold=0.9,
            )
            if result.applied:
                logger.info("pi_link_resolve.applied", ...)
    """

    def __init__(self, pi_mode: PiMode | None = None) -> None:
        self.settings = get_settings()
        self.pi_mode = pi_mode or PiMode(os.environ.get("PI_MODE", "sdk"))
        self._last_error: dict | None = None

    async def find_unresolved_links(self, bundle_dir: Path) -> list[UnresolvedLink]:
        """Bundle 디렉터리 에서 unresolved cross-link 모두 추출 (§3.5.7.1).

        unresolved 정의:
        - extract_cross_links() 가 추출한 link
        - link_target 의 .md 파일이 존재하지 않는 경우

        Returns:
            list[UnresolvedLink]
        """
        if not bundle_dir.exists():
            logger.warning("link_resolver.bundle_dir_not_found", path=str(bundle_dir))
            return []

        unresolved: list[UnresolvedLink] = []
        for md_file in bundle_dir.rglob("*.md"):
            try:
                content = md_file.read_text(encoding="utf-8")
            except (UnicodeDecodeError, OSError) as e:
                logger.warning(
                    "link_resolver.read_failed",
                    path=str(md_file),
                    error=str(e),
                )
                continue

            links = extract_cross_links(content)
            source_path = str(md_file.relative_to(bundle_dir.parent.parent))
            for link in links:
                if not _target_exists(bundle_dir, link.target):
                    ctx = _extract_context(content, link.start, link.end, context_chars=100)
                    unresolved.append(
                        UnresolvedLink(
                            source_path=source_path,
                            link_text=link.text,
                            link_target=link.target,
                            context=ctx,
                        )
                    )
        logger.info(
            "link_resolver.found_unresolved",
            count=len(unresolved),
            bundle_dir=str(bundle_dir),
        )
        return unresolved

    async def list_candidates(
        self,
        link_target: str,
        limit: int = 10,
    ) -> list[dict[str, Any]]:
        """link_target 기반 candidate concept 검색 (§3.5.7.1).

        정공법: link_target 의 마지막 segment (slug) 를 기준으로 name fuzzy match.

        Returns:
            list of {path, title, type} dicts
        """
        target_slug = link_target.rstrip(".md").split("/")[-1]
        bundles_dir = self.settings.var_dir / "bundles"
        if not bundles_dir.exists():
            return []

        candidates: list[dict[str, Any]] = []
        for md_file in bundles_dir.rglob("*.md"):
            try:
                content = md_file.read_text(encoding="utf-8")
            except (UnicodeDecodeError, OSError):
                continue
            if target_slug.lower() in md_file.name.lower():
                title = _extract_title_from_md(content)
                candidates.append(
                    {
                        "path": str(md_file.relative_to(bundles_dir.parent)),
                        "title": title,
                        "type": _extract_type_from_md(content),
                    }
                )
                if len(candidates) >= limit:
                    break
        return candidates

    async def resolve(
        self,
        link: UnresolvedLink,
        candidate_concepts: list[dict[str, Any]],
        mode: ResolutionMode = ResolutionMode.DRY_RUN,
        confidence_threshold: float = 0.9,
        selected_rank: int = 1,
    ) -> ResolutionResult:
        """Unresolved link 를 Pi LLM 으로 resolution (§3.5.7.4).

        3 mode confirm workflow:
        - dry-run: recommend 만, 변경 ❌
        - confirm: operator 가 1 row 선택, confidence 무관 적용
        - auto-apply: confidence ≥ threshold 자동 적용, else 변경 ❌

        Returns:
            ResolutionResult (applied boolean 으로 caller 가 다음 action 결정)
        """
        prompt = render_prompt(
            source_path=link.source_path,
            link_text=link.link_text,
            link_target=link.link_target,
            context=link.context,
            candidates=candidate_concepts,
        )

        recommendations = await self._invoke_pi(prompt, candidate_concepts)

        applied = False
        selected: LinkRecommendation | None = None
        if mode == ResolutionMode.DRY_RUN:
            selected = recommendations[0] if recommendations else None
        elif mode == ResolutionMode.CONFIRM:
            selected = recommendations[selected_rank - 1] if recommendations else None
            applied = selected is not None
        elif mode == ResolutionMode.AUTO_APPLY:
            selected = recommendations[0] if recommendations else None
            if selected is not None and selected.confidence >= confidence_threshold:
                applied = True

        result = ResolutionResult(
            source_path=link.source_path,
            original_link=link.link_text,
            selected=selected,
            alternatives=recommendations,
            mode=mode,
            confidence_threshold=confidence_threshold,
            applied=applied,
            timestamp=datetime.now(timezone.utc).isoformat(),
        )

        # audit log (§3.5.7.5)
        if applied:
            logger.info(
                "pi_link_resolve.applied",
                source_path=link.source_path,
                original_link=link.link_text,
                selected_slug=selected.target_slug if selected else None,
                confidence=selected.confidence if selected else 0.0,
                mode=mode.value,
            )
        else:
            logger.info(
                "pi_link_resolve.dry_run" if mode == ResolutionMode.DRY_RUN else "pi_link_resolve.skipped",
                source_path=link.source_path,
                reason="below_threshold" if mode == ResolutionMode.AUTO_APPLY else "manual_review",
            )

        return result

    async def _invoke_pi(
        self,
        prompt: str,
        candidates: list[dict[str, Any]],
    ) -> list[LinkRecommendation]:
        """Pi LLM 호출 (§10.3 SDK/RPC mode).

        M-v0.2.3+ PoC: mock mode (Pi SDK 미설치/미연결 시 자동 fallback).
        """
        if self.pi_mode == PiMode.SDK:
            return await self._invoke_pi_sdk(prompt, candidates)
        elif self.pi_mode == PiMode.RPC:
            return await self._invoke_pi_rpc(prompt, candidates)
        else:
            self._last_error = {"code": "unknown_mode", "message": f"unknown PiMode: {self.pi_mode}"}
            return []

    async def _invoke_pi_sdk(
        self,
        prompt: str,
        candidates: list[dict[str, Any]],
    ) -> list[LinkRecommendation]:
        """Pi SDK mode (M-v0.2.3+ PoC default).

        pi-coding-agent SDK 가 설치되어 있으면 실제 호출, 없으면 mock fallback.
        """
        try:
            from pi_coding_agent import Agent  # type: ignore[import-not-found]
        except ImportError:
            return self._mock_pi_response(prompt, candidates)

        try:
            agent = Agent(system_prompt="You are a DevHub cross-link resolver.")
            response = await agent.run(prompt)
            return _parse_pi_response(response.text)
        except Exception as e:
            self._last_error = {"code": "pi_sdk_error", "message": str(e)}
            logger.warning("link_resolver.pi_sdk_failed", error=str(e))
            return self._mock_pi_response(prompt, candidates)

    async def _invoke_pi_rpc(
        self,
        prompt: str,
        candidates: list[dict[str, Any]],
    ) -> list[LinkRecommendation]:
        """Pi RPC mode (M-v0.2.3+ production option).

        RPC server URL: PI_RPC_URL (env var)
        """
        import httpx

        rpc_url = os.environ.get("PI_RPC_URL")
        if not rpc_url:
            return self._mock_pi_response(prompt, candidates)

        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{rpc_url}/resolve",
                    json={"prompt": prompt},
                    timeout=self.settings.hrdb_timeout_seconds,
                )
                response.raise_for_status()
                return _parse_pi_response(response.json()["text"])
        except Exception as e:
            self._last_error = {"code": "pi_rpc_error", "message": str(e)}
            logger.warning("link_resolver.pi_rpc_failed", error=str(e))
            return self._mock_pi_response(prompt, candidates)

    def _mock_pi_response(
        self,
        prompt: str,
        candidates: list[dict[str, Any]],
    ) -> list[LinkRecommendation]:
        """Pi LLM mock response (PoC default, SDK/RPC 미설치 시).

        첫 3 row candidate 를 confidence 0.95/0.85/0.75 로 변환.
        """
        return [
            LinkRecommendation(
                rank=i + 1,
                target_slug=c["path"].rstrip(".md").split("/")[-1],
                target_path=c["path"],
                reason=f"Mock recommendation #{i + 1} for {c['path']}",
                confidence=round(0.95 - i * 0.10, 2),
            )
            for i, c in enumerate(candidates[:3])
        ]


# ----------------------------------------------------------------------
# Helper functions
# ----------------------------------------------------------------------


def _target_exists(bundle_dir: Path, target: str) -> bool:
    """link_target 의 .md 파일 존재 여부."""
    # target = "../other/category/slug.md" or "category/slug.md"
    target_path = (bundle_dir / target).resolve()
    return target_path.exists()


def _extract_context(body: str, match_start: int, match_end: int, context_chars: int = 100) -> str:
    """link 주변 ±2 lines context (§3.5.7.2 정공법)."""
    # 줄 경계 찾기
    line_start = body.rfind("\n", 0, match_start) + 1
    line_end = body.find("\n", match_end)
    if line_end == -1:
        line_end = len(body)

    # 2 줄 전후
    pre_start = body.rfind("\n", 0, line_start - 1)
    if pre_start == -1:
        pre_start = 0
    else:
        pre_start = pre_start + 1

    post_end = body.find("\n", line_end + 1)
    if post_end == -1:
        post_end = len(body)
    else:
        post_end_line = body.rfind("\n", 0, post_end - 1)
        post_end = post_end_line if post_end_line != -1 else post_end

    return body[pre_start:post_end].strip()


def _extract_title_from_md(content: str) -> str:
    """Markdown 첫 # 헤더 추출."""
    for line in content.split("\n"):
        if line.startswith("# "):
            return line[2:].strip()
    return "(no title)"


def _extract_type_from_md(content: str) -> str:
    """Frontmatter `type` 필드 추출."""
    if not content.startswith("---"):
        return "unknown"
    try:
        end = content.index("---", 3)
        fm = content[3:end]
        for line in fm.split("\n"):
            if line.startswith("type:"):
                return line.split(":", 1)[1].strip()
    except ValueError:
        pass
    return "unknown"


def _parse_pi_response(text: str) -> list[LinkRecommendation]:
    """Pi LLM 응답 → LinkRecommendation list 파싱.

    JSON format expected (j2 prompt template 정의).
    """
    import json
    import re

    match = re.search(r"\{[\s\S]*\}", text)
    if not match:
        return []
    try:
        data = json.loads(match.group(0))
        return [
            LinkRecommendation(
                rank=r["rank"],
                target_slug=r["target_slug"],
                target_path=r["target_path"],
                reason=r["reason"],
                confidence=float(r["confidence"]),
            )
            for r in data.get("recommendations", [])
        ]
    except (json.JSONDecodeError, KeyError, ValueError):
        return []