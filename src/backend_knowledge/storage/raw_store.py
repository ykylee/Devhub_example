"""Raw storage (file mode + 봉투 암호화, umbrella doc §3.6 / ADR-0025).

scope = raw + .env/KEK 만. bundle / concept (.md) 는 봉투 암호화 ❌.

File layout (per raw):
- var/raw/{source}/{raw_id}.bin — 봉투 암호화 body (KEK mode)
- var/raw/{source}/{raw_id}.json — plaintext body (KEK 미설정, PoC default)
- var/raw/{source}/{raw_id}.meta.json — metadata sidecar (type/name/owner/visibility/frontmatter)
- var/raw/{source}/{raw_id}.dek — 봉투 암호화 DEK wrap (KEK mode only, per-raw DEK)

Envelope format v2 (AES-256-GCM, per-raw DEK + KEK wrap):
- [version 2][kek_nonce 12][wrapped_dek 48][dek_nonce 12][ciphertext + auth tag 16+]
- DEK per raw (per-message random, 32 byte)
- KEK from RAW_ENCRYPTION_KEY (base64, 32 byte)
- kek_nonce: 12 byte random (DEK wrap with KEK)
- dek_nonce: 12 byte random (body encrypt with DEK)
- auth tag: 16 byte

Codex P2 review fix (PR 1): per-raw DEK generated, wrapped with KEK, never reused.
"""

from __future__ import annotations

import base64
import hashlib
import json
import os
import re
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from cryptography.hazmat.primitives.ciphers.aead import AESGCM

from ..config import get_settings
from ..logger import get_logger

logger = get_logger(__name__)


SOURCE_NAME_PATTERN = re.compile(r"^[a-z0-9][a-z0-9_-]{0,62}$")


class RawStoreError(Exception):
    """Raised on raw store failures (encryption, file write, etc)."""


class InvalidSourceNameError(RawStoreError):
    """Raised when source name fails whitelist validation (path traversal defense)."""


@dataclass
class RawRecord:
    """In-memory raw record (decrypted plaintext + metadata)."""

    raw_id: str  # sha256 prefix 7 + uuid suffix
    source: str
    type: str  # "dataset" | "metric" | ... (8종)
    name: str
    body: str  # decrypted plaintext (markdown, JSON, etc)
    frontmatter_override: dict = field(default_factory=dict)
    raw_refs: list[str] = field(default_factory=list)
    sha256: str = ""
    size: int = 0
    received_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))
    registered_by: str = ""  # Path Y user_id
    owner_org_id: str = ""  # Path Y org_id
    owner_project_ids: list[str] = field(default_factory=list)  # Path Y project_ids
    visibility: str = "org"  # public | org | personal | project


@dataclass
class RawStoreResult:
    """Raw store write result."""

    raw_id: str
    sha256: str
    size: int
    envelope_encrypted: bool
    registered_at: datetime


class RawStore:
    """File mode raw storage + 봉투 암호화 (AES-256-GCM).

    PoC default: file mode (var/raw/{source}/) + envelope encryption if RAW_ENCRYPTION_KEY set, plaintext otherwise.
    M-v0.2.3+ production: PostgreSQL optional (umbrella doc §10).
    """

    FILE_VERSION: int = 2

    def __init__(self, base_dir: Path | None = None, kek: bytes | None = None):
        """Initialize raw store.

        Args:
            base_dir: var/raw directory (default: get_settings().var_dir / 'raw')
            kek: KEK bytes (32 byte, base64 decoded from RAW_ENCRYPTION_KEY). None → plaintext mode.
        """
        settings = get_settings()
        self.base_dir = base_dir or (settings.var_dir / "raw")
        self.base_dir.mkdir(parents=True, exist_ok=True)
        self.kek = kek
        if kek is not None and len(kek) != 32:
            raise RawStoreError(f"KEK must be 32 bytes (AES-256), got {len(kek)}")

    @staticmethod
    def _validate_source_name(source: str) -> str:
        """Validate source name against whitelist (Codex P2 fix: path traversal defense).

        Allowed: ^[a-z0-9][a-z0-9_-]{0,62}$
        Reject: empty, '/', '..', absolute path components, uppercase, special chars.

        Returns: validated source (echo)
        Raises: InvalidSourceNameError if rejected.
        """
        if not isinstance(source, str) or not SOURCE_NAME_PATTERN.match(source):
            raise InvalidSourceNameError(
                f"invalid source name: {source!r}. must match {SOURCE_NAME_PATTERN.pattern}"
            )
        return source

    def _generate_raw_id(self, sha256: str) -> str:
        """raw_id format: sha256[:7] + uuid[:8] (예: 'abc1234-def56789')."""
        prefix = sha256[:7]
        suffix = uuid.uuid4().hex[:8]
        return f"{prefix}-{suffix}"

    def _compute_sha256(self, body: bytes) -> str:
        return hashlib.sha256(body).hexdigest()

    def _encrypt(self, plaintext: bytes) -> bytes:
        """AES-256-GCM 봉투 암호화 v2 (per-raw DEK).

        Returns: envelope bytes
            [version 2][kek_nonce 12][wrapped_dek 48][dek_nonce 12][ciphertext + auth tag 16+]
        """
        if self.kek is None:
            raise RawStoreError("encryption requested but KEK is not set")

        # Codex P2 fix (PR 1): generate per-raw DEK, wrap with KEK, never reuse.
        dek = os.urandom(32)  # 256-bit Data Encryption Key
        kek_nonce = os.urandom(12)
        dek_nonce = os.urandom(12)

        kek_aesgcm = AESGCM(self.kek)
        wrapped_dek = kek_aesgcm.encrypt(kek_nonce, dek, None)  # 32 + 16 = 48 bytes

        dek_aesgcm = AESGCM(dek)
        ciphertext = dek_aesgcm.encrypt(dek_nonce, plaintext, None)  # plaintext + 16 tag

        return (
            bytes([self.FILE_VERSION])
            + kek_nonce
            + wrapped_dek
            + dek_nonce
            + ciphertext
        )

    def _decrypt(self, envelope: bytes) -> bytes:
        """AES-256-GCM 봉투 복호화 v2. Reverse of _encrypt."""
        if self.kek is None:
            raise RawStoreError("decryption requested but KEK is not set")
        # version(1) + kek_nonce(12) + wrapped_dek(48) + dek_nonce(12) + min_tag(16) = 89
        if len(envelope) < 1 + 12 + 48 + 12 + 16:
            raise RawStoreError(f"envelope too short: {len(envelope)} bytes")
        version = envelope[0]
        if version != self.FILE_VERSION:
            raise RawStoreError(f"unsupported envelope version: {version}")
        kek_nonce = envelope[1:13]
        wrapped_dek = envelope[13:61]
        dek_nonce = envelope[61:73]
        ciphertext_with_tag = envelope[73:]

        kek_aesgcm = AESGCM(self.kek)
        dek = kek_aesgcm.decrypt(kek_nonce, wrapped_dek, None)

        dek_aesgcm = AESGCM(dek)
        return dek_aesgcm.decrypt(dek_nonce, ciphertext_with_tag, None)

    def _meta_path(self, source: str, raw_id: str) -> Path:
        return self.base_dir / source / f"{raw_id}.meta.json"

    def _save_meta_sidecar(
        self,
        source: str,
        raw_id: str,
        type_: str,
        name: str,
        sha256: str,
        size: int,
        registered_by: str,
        owner_org_id: str,
        owner_project_ids: list[str],
        visibility: str,
        frontmatter_override: dict,
        raw_refs: list[str],
    ) -> None:
        """Write metadata sidecar JSON (Codex P1 fix: persist raw metadata).

        Stores type/name/owner/visibility/frontmatter so load() can return a fully-populated
        RawRecord without re-parsing the body. Used by DELETE authorization and FR-I-004/005.
        """
        meta_path = self._meta_path(source, raw_id)
        meta = {
            "raw_id": raw_id,
            "source": source,
            "type": type_,
            "name": name,
            "sha256": sha256,
            "size": size,
            "registered_at": datetime.now(timezone.utc).isoformat(),
            "registered_by": registered_by,
            "owner_org_id": owner_org_id,
            "owner_project_ids": owner_project_ids,
            "visibility": visibility,
            "frontmatter_override": frontmatter_override,
            "raw_refs": raw_refs,
        }
        meta_path.write_text(json.dumps(meta, ensure_ascii=False, indent=2), encoding="utf-8")

    def _load_meta_sidecar(self, source: str, raw_id: str) -> dict | None:
        """Read metadata sidecar JSON. Returns None if not present (legacy record)."""
        meta_path = self._meta_path(source, raw_id)
        if not meta_path.exists():
            return None
        try:
            return json.loads(meta_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            return None

    def save(
        self,
        source: str,
        type_: str,
        name: str,
        body: str,
        registered_by: str = "",
        owner_org_id: str = "",
        owner_project_ids: list[str] | None = None,
        frontmatter_override: dict | None = None,
        raw_refs: list[str] | None = None,
        visibility: str = "org",
    ) -> RawStoreResult:
        """Save raw data to file mode storage.

        Returns: RawStoreResult (raw_id, sha256, size, envelope_encrypted, registered_at)
        """
        self._validate_source_name(source)
        body_bytes = body.encode("utf-8")
        sha256 = self._compute_sha256(body_bytes)
        raw_id = self._generate_raw_id(sha256)

        if self.kek is not None:
            envelope = self._encrypt(body_bytes)
            ext = "bin"
            envelope_encrypted = True
        else:
            envelope = body_bytes
            ext = "json"
            envelope_encrypted = False

        source_dir = self.base_dir / source
        source_dir.mkdir(parents=True, exist_ok=True)
        file_path = source_dir / f"{raw_id}.{ext}"
        file_path.write_bytes(envelope)

        size = len(envelope)

        self._save_meta_sidecar(
            source=source,
            raw_id=raw_id,
            type_=type_,
            name=name,
            sha256=sha256,
            size=size,
            registered_by=registered_by,
            owner_org_id=owner_org_id,
            owner_project_ids=list(owner_project_ids or []),
            visibility=visibility,
            frontmatter_override=dict(frontmatter_override or {}),
            raw_refs=list(raw_refs or []),
        )

        logger.info(
            "raw_saved",
            raw_id=raw_id,
            source=source,
            type=type_,
            name=name,
            sha256=sha256,
            size=size,
            envelope_encrypted=envelope_encrypted,
            registered_by=registered_by,
            owner_org_id=owner_org_id,
        )

        return RawStoreResult(
            raw_id=raw_id,
            sha256=sha256,
            size=size,
            envelope_encrypted=envelope_encrypted,
            registered_at=datetime.now(timezone.utc),
        )

    def load(self, source: str, raw_id: str) -> RawRecord:
        """Load raw data from file mode storage.

        Returns RawRecord with metadata populated from sidecar (if present).
        Raises FileNotFoundError if raw_id not found.
        """
        self._validate_source_name(source)
        meta = self._load_meta_sidecar(source, raw_id)

        for ext in ("bin", "json"):
            file_path = self.base_dir / source / f"{raw_id}.{ext}"
            if file_path.exists():
                envelope = file_path.read_bytes()
                if ext == "bin":
                    body_bytes = self._decrypt(envelope)
                else:
                    body_bytes = envelope
                body = body_bytes.decode("utf-8")
                sha256 = self._compute_sha256(body_bytes)

                if meta:
                    return RawRecord(
                        raw_id=raw_id,
                        source=source,
                        type=meta.get("type", ""),
                        name=meta.get("name", ""),
                        body=body,
                        frontmatter_override=meta.get("frontmatter_override", {}),
                        raw_refs=meta.get("raw_refs", []),
                        sha256=sha256,
                        size=len(envelope),
                        received_at=datetime.fromisoformat(meta["registered_at"]) if meta.get("registered_at") else datetime.now(timezone.utc),
                        registered_by=meta.get("registered_by", ""),
                        owner_org_id=meta.get("owner_org_id", ""),
                        owner_project_ids=meta.get("owner_project_ids", []),
                        visibility=meta.get("visibility", "org"),
                    )
                else:
                    return RawRecord(
                        raw_id=raw_id,
                        source=source,
                        type="",
                        name="",
                        body=body,
                        sha256=sha256,
                        size=len(envelope),
                        received_at=datetime.now(timezone.utc),
                        registered_by="",
                    )

        raise FileNotFoundError(f"raw not found: {source}/{raw_id}")

    def exists(self, source: str, raw_id: str) -> bool:
        """Check if raw_id exists in source."""
        self._validate_source_name(source)
        return (self.base_dir / source / f"{raw_id}.bin").exists() or (
            self.base_dir / source / f"{raw_id}.json"
        ).exists()

    def list_source(self, source: str, limit: int = 100) -> list[str]:
        """List raw_id in source directory (no decryption)."""
        self._validate_source_name(source)
        source_dir = self.base_dir / source
        if not source_dir.exists():
            return []
        ids: list[str] = []
        for path in sorted(source_dir.iterdir()):
            if path.is_file() and (path.suffix in (".bin", ".json")) and not path.name.endswith(".meta.json"):
                ids.append(path.stem)
                if len(ids) >= limit:
                    break
        return ids

    def delete(self, source: str, raw_id: str) -> bool:
        """Hard delete raw body + meta sidecar. Returns True if anything was deleted."""
        self._validate_source_name(source)
        deleted = False
        source_dir = self.base_dir / source
        for ext in ("bin", "json"):
            file_path = source_dir / f"{raw_id}.{ext}"
            if file_path.exists():
                file_path.unlink()
                deleted = True
        meta_path = self._meta_path(source, raw_id)
        if meta_path.exists():
            meta_path.unlink()
            deleted = True
        return deleted


_raw_store: RawStore | None = None


def get_raw_store() -> RawStore:
    """Singleton raw store accessor."""
    global _raw_store
    if _raw_store is None:
        settings = get_settings()
        kek_bytes: bytes | None = None
        if settings.raw_encryption_key:
            try:
                kek_bytes = base64.b64decode(settings.raw_encryption_key)
            except Exception as e:
                logger.warning("raw_encryption_key_decode_failed", error=str(e))
        _raw_store = RawStore(kek=kek_bytes)
    return _raw_store
