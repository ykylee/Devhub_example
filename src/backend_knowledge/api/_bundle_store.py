"""Bundle / concept storage helper (architecture.md §3 layer 격리 정공법).

§13.3 refactor (Private helper cross-router): curate.py 의 8 private helper 를 public module 로 추출.
원래 curate.py 에 정의되어 있던 helper 8 종 (모두 §3.4 bundle directory + concept metadata sidecar 운영):
- bundle_dir / bundle_index_dir / bundle_meta_path / concept_meta_path
- save_concept_metadata / load_concept_metadata
- find_concept_by_id / build_concept_id

api/curate.py + api/query.py + api/graph.py 3 file 에서 import (cross-router call ❌ → public utility ✅).
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path

from ..config import get_settings


def bundle_dir(bundle_name: str) -> Path:
    """Return bundle directory path (var/bundles/{bundle_name}/)."""
    settings = get_settings()
    return settings.var_dir / "bundles" / bundle_name


def bundle_index_dir(bundle_name: str) -> Path:
    """Return bundle index directory (var/bundles/{bundle_name}/.index/)."""
    return bundle_dir(bundle_name) / ".index"


def bundle_meta_path(bundle_name: str) -> Path:
    """Return bundle metadata JSON path (var/bundles/{bundle_name}/.bundle_meta.json)."""
    return bundle_dir(bundle_name) / ".bundle_meta.json"


def concept_meta_path(bundle: str, type_: str, slug: str) -> Path:
    """Return concept metadata sidecar path.

    Stores type / name / registered_by / visibility / frontmatter for fast lookup
    without re-parsing the full Markdown file. Per Codex P1 review fix (PR 1).
    """
    return bundle_dir(bundle) / type_ / f"{slug}.meta.json"


def save_concept_metadata(
    bundle: str,
    type_: str,
    slug: str,
    sha256: str,
    source: str,
    raw_id: str | None,
    registered_by: str,
    visibility: str,
    frontmatter: dict,
) -> None:
    """Save concept metadata sidecar JSON (per Codex P1 review fix)."""
    meta_path = concept_meta_path(bundle, type_, slug)
    meta_path.parent.mkdir(parents=True, exist_ok=True)
    meta = {
        "bundle": bundle,
        "type": type_,
        "name": slug,
        "sha256": sha256,
        "source": source,
        "raw_id": raw_id,
        "registered_by": registered_by,
        "visibility": visibility,
        "frontmatter": frontmatter,
        "registered_at": datetime.now(timezone.utc).isoformat(),
    }
    meta_path.write_text(json.dumps(meta, ensure_ascii=False, indent=2), encoding="utf-8")


def load_concept_metadata(bundle: str, type_: str, slug: str) -> dict | None:
    """Load concept metadata sidecar JSON (per Codex P1 review fix)."""
    meta_path = concept_meta_path(bundle, type_, slug)
    if not meta_path.exists():
        return None
    try:
        return json.loads(meta_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return None


def find_concept_by_id(concept_id: str) -> tuple[str, str, str] | None:
    """Find concept (bundle, type, slug) by concept_id (e.g., 'devhub-gitea/dataset/foo').

    Returns (bundle, type, slug) or None.
    Searches all bundles' concept metadata.
    """
    settings = get_settings()
    bundles_dir = settings.var_dir / "bundles"
    if not bundles_dir.exists():
        return None

    parts = concept_id.split("/")
    if len(parts) >= 3:
        bundle, type_, slug = parts[0], parts[1], parts[-1]
        meta = load_concept_metadata(bundle, type_, slug)
        if meta:
            return (bundle, type_, slug)

    for bundle_meta_path in bundles_dir.glob("*/.bundle_meta.json"):
        bundle_name = bundle_meta_path.parent.name
        bundle_dir_path = bundle_meta_path.parent
        for concept_meta in bundle_dir_path.glob("*/*.meta.json"):
            type_ = concept_meta.parent.name
            slug = concept_meta.stem.replace(".meta", "")
            if concept_meta.stem.endswith(".meta"):
                slug = concept_meta.stem[:-5]
            full_id = f"{bundle_name}/{type_}/{slug}"
            if full_id == concept_id:
                return (bundle_name, type_, slug)
    return None


def build_concept_id(bundle: str, type_: str, slug: str) -> str:
    """Build concept_id from bundle + type + slug."""
    return f"{bundle}/{type_}/{slug}"
