"""Storage (raw + bundle + audit + log).

umbrella doc §2.1 + §10 정합:
- Raw storage: 봉투 암호화 (AES-256-GCM, scope = raw + .env/KEK 만) + file mode
- Bundle storage: var/bundles/{bundle}/{type}/{slug}.md (git-pushable, 봉투 암호화 ❌)
- Audit log: var/audit/audit-YYYY-MM-DD.jsonl (JSON Lines)
- Application log: var/log/backend-knowledge.jsonl (structlog)
"""

from .raw_store import RawRecord, RawStore, RawStoreResult, get_raw_store

__all__ = [
    "RawRecord",
    "RawStore",
    "RawStoreResult",
    "get_raw_store",
]
