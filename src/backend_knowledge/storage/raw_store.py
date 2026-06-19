"""Raw storage (file mode + 봉투 암호화, umbrella doc §3.6 / ADR-0025).

scope = raw + .env/KEK 만. bundle / concept (.md) 는 봉투 암호화 ❌.

File path: var/raw/{source}/{sha256_prefix}-{uuid}.bin (envelope encrypted)
        or var/raw/{source}/{sha256_prefix}-{uuid}.json (plaintext fallback when RAW_ENCRYPTION_KEY missing)

Envelope format (AES-256-GCM):
- DEK per raw (per-message random, 32 byte)
- KEK from RAW_ENCRYPTION_KEY (base64, 32 byte)
- nonce 96 bit (random per encryption)
- auth tag 128 bit
- file structure: [version 1 byte][nonce 12 byte][ciphertext + auth tag]
"""

from __future__ import annotations

import base64
import hashlib
import os
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from cryptography.hazmat.primitives.ciphers.aead import AESGCM

from ..config import get_settings
from ..logger import get_logger

logger = get_logger(__name__)


class RawStoreError(Exception):
    """Raised on raw store failures (encryption, file write, etc)."""


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

    FILE_VERSION: int = 1

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

    def _generate_raw_id(self, sha256: str) -> str:
        """raw_id format: sha256[:7] + uuid[:8] (예: 'abc1234-def56789')."""
        prefix = sha256[:7]
        suffix = uuid.uuid4().hex[:8]
        return f"{prefix}-{suffix}"

    def _compute_sha256(self, body: bytes) -> str:
        return hashlib.sha256(body).hexdigest()

    def _encrypt(self, plaintext: bytes) -> bytes:
        """AES-256-GCM encrypt. Returns envelope: [version 1][nonce 12][ciphertext + auth tag 16]."""
        if self.kek is None:
            raise RawStoreError("encryption requested but KEK is not set")
        nonce = os.urandom(12)
        aesgcm = AESGCM(self.kek)
        ciphertext = aesgcm.encrypt(nonce, plaintext, None)  # AAD = None
        # ciphertext already includes 16-byte auth tag at end
        return bytes([self.FILE_VERSION]) + nonce + ciphertext

    def _decrypt(self, envelope: bytes) -> bytes:
        """AES-256-GCM decrypt. Reverse of _encrypt."""
        if self.kek is None:
            raise RawStoreError("decryption requested but KEK is not set")
        if len(envelope) < 1 + 12 + 16:
            raise RawStoreError(f"envelope too short: {len(envelope)} bytes")
        version = envelope[0]
        if version != self.FILE_VERSION:
            raise RawStoreError(f"unsupported envelope version: {version}")
        nonce = envelope[1:13]
        ciphertext_with_tag = envelope[13:]
        aesgcm = AESGCM(self.kek)
        return aesgcm.decrypt(nonce, ciphertext_with_tag, None)

    def save(
        self,
        source: str,
        type_: str,
        name: str,
        body: str,
        registered_by: str = "",
        frontmatter_override: dict | None = None,
        raw_refs: list[str] | None = None,
        visibility: str = "org",
    ) -> RawStoreResult:
        """Save raw data to file mode storage.

        Returns: RawStoreResult (raw_id, sha256, size, envelope_encrypted, registered_at)
        """
        body_bytes = body.encode("utf-8")
        sha256 = self._compute_sha256(body_bytes)
        raw_id = self._generate_raw_id(sha256)

        if self.kek is not None:
            # 봉투 암호화 mode
            envelope = self._encrypt(body_bytes)
            ext = "bin"
            envelope_encrypted = True
        else:
            # Plaintext mode (PoC default when KEK not set)
            envelope = body_bytes
            ext = "json"  # .json for plaintext readability
            envelope_encrypted = False

        # Write to var/raw/{source}/{raw_id}.{ext}
        source_dir = self.base_dir / source
        source_dir.mkdir(parents=True, exist_ok=True)
        file_path = source_dir / f"{raw_id}.{ext}"
        file_path.write_bytes(envelope)

        size = len(envelope)

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

        Raises FileNotFoundError if raw_id not found.
        """
        # Try .bin (encrypted) first, then .json (plaintext)
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
                return RawRecord(
                    raw_id=raw_id,
                    source=source,
                    type="",  # unknown on load (override required)
                    name="",
                    body=body,
                    sha256=sha256,
                    size=len(envelope),
                    received_at=datetime.now(timezone.utc),  # placeholder
                    registered_by="",
                )

        raise FileNotFoundError(f"raw not found: {source}/{raw_id}")

    def exists(self, source: str, raw_id: str) -> bool:
        """Check if raw_id exists in source."""
        return (self.base_dir / source / f"{raw_id}.bin").exists() or (
            self.base_dir / source / f"{raw_id}.json"
        ).exists()

    def list_source(self, source: str, limit: int = 100) -> list[str]:
        """List raw_id in source directory (no decryption)."""
        source_dir = self.base_dir / source
        if not source_dir.exists():
            return []
        ids: list[str] = []
        for path in sorted(source_dir.iterdir()):
            if path.is_file() and (path.suffix in (".bin", ".json")):
                ids.append(path.stem)
                if len(ids) >= limit:
                    break
        return ids


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
