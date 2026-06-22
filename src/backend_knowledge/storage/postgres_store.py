"""PostgreSQL raw storage backend (umbrella doc §10.1 + M-v0.2.3+).

scope = raw + .env/KEK 만. bundle / concept (.md) 는 git-pushable + wiki review flow.

POSTGRES_URL 미설정 시 None 반환 (caller 가 file mode 로 fallback).
PostgreSQL connection pool (asyncpg) 기반 raw_records table 관리.

Schema (raw_records):
- raw_id TEXT PRIMARY KEY (sha256 prefix 7 + uuid suffix)
- source TEXT NOT NULL (source plugin name, e.g., "hrdb")
- type TEXT NOT NULL (8종 enum: dataset/metric/api_endpoint/runbook/integration/event/reference/decision)
- name TEXT NOT NULL (concept identifier slug)
- body BYTEA NOT NULL (봉투 암호화 body, KEK mode)
- body_plaintext TEXT (KEK 미설정 시, plaintext mode fallback)
- frontmatter JSONB (frontmatter metadata)
- owner_org_id TEXT (Path Y owner org)
- visibility TEXT (org/personal/project/public)
- ingested_at TIMESTAMPTZ NOT NULL
- ingest_lock BOOLEAN DEFAULT FALSE (M-v0.2.3+ Pi-driven periodic ingest lock)
- last_verified_at TIMESTAMPTZ
- sha256 TEXT NOT NULL (body integrity verification)
- concept_ids JSONB DEFAULT '[]' (1 raw → N concepts traceability, §3.6.6.4)

Index:
- idx_raw_records_source_received ON (source, ingested_at DESC)
- idx_raw_records_visibility ON (visibility)
- idx_raw_records_ingested ON (ingested_at DESC) — Pi periodic ingest query
- idx_raw_records_ingest_lock ON (ingest_lock) — partial index WHERE ingest_lock = TRUE

Codex P2 review fix (PR 11): schema 는 umbrella doc §10.1 raw_records 와 1:1 정합.
"""

from __future__ import annotations

import base64
import hashlib
import json
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any

from ..config import get_settings
from ..logger import get_logger
from .raw_store import RawRecord, RawStoreError

logger = get_logger(__name__)


# ----------------------------------------------------------------------
# Pydantic-style data class (raw_records table row)
# ----------------------------------------------------------------------


@dataclass
class PostgresRawRecord:
    """raw_records table row (decrypted form)."""

    raw_id: str
    source: str
    type: str
    name: str
    body: bytes  # decrypted body (raw bytes)
    frontmatter: dict[str, Any]
    owner_org_id: str | None
    visibility: str
    ingested_at: datetime
    ingest_lock: bool
    last_verified_at: datetime | None
    sha256: str
    concept_ids: list[str]


# ----------------------------------------------------------------------
# Connection pool management
# ----------------------------------------------------------------------


class PostgresStoreError(RawStoreError):
    """PostgreSQL store errors (connection, query, transaction)."""


class PostgresStore:
    """PostgreSQL raw storage backend (asyncpg pool).

    Usage:
        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            record = await store.insert(raw_record)
            fetched = await store.select(raw_id="abc1234")
            rows = await store.list_by_source(source="hrdb", limit=100)

    POSTGRES_URL 미설정 시 PostgresStore.connect() raises PostgresStoreError
    (caller 가 file mode fallback 결정).
    """

    def __init__(self, dsn: str, pool_size: int = 5, max_overflow: int = 10) -> None:
        self.dsn = dsn
        self.pool_size = pool_size
        self.max_overflow = max_overflow
        self._pool: Any = None  # asyncpg.Pool

    @classmethod
    @asynccontextmanager
    async def connect(cls) -> "PostgresStore":
        """Context manager: POSTGRES_URL 기반 pool 생성.

        Raises:
            PostgresStoreError: POSTGRES_URL 미설정 시
        """
        settings = get_settings()
        if not settings.postgres_url:
            raise PostgresStoreError("POSTGRES_URL not configured (file mode fallback)")

        store = cls(
            dsn=settings.postgres_url,
            pool_size=settings.postgres_pool_size,
            max_overflow=settings.postgres_pool_max_overflow,
        )
        try:
            # asyncpg pool 생성은 lazy import (optional dependency)
            try:
                import asyncpg  # type: ignore[import-not-found]
            except ImportError as e:
                raise PostgresStoreError(
                    "asyncpg not installed. run: pip install 'asyncpg>=0.29.0'"
                ) from e

            store._pool = await asyncpg.create_pool(
                dsn=store.dsn,
                min_size=1,
                max_size=store.pool_size + store.max_overflow,
                command_timeout=settings.hrdb_timeout_seconds,
            )
            logger.info(
                "postgres_store.connected",
                pool_size=store.pool_size,
                max_overflow=store.max_overflow,
            )
            yield store
        finally:
            if store._pool is not None:
                await store._pool.close()
                logger.info("postgres_store.closed")

    async def initialize_schema(self) -> None:
        """CREATE TABLE IF NOT EXISTS raw_records + indexes (§10.1 정합)."""
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            await conn.execute(
                """
                CREATE TABLE IF NOT EXISTS raw_records (
                    raw_id TEXT PRIMARY KEY,
                    source TEXT NOT NULL,
                    type TEXT NOT NULL,
                    name TEXT NOT NULL,
                    body BYTEA NOT NULL,
                    frontmatter JSONB NOT NULL DEFAULT '{}'::jsonb,
                    owner_org_id TEXT,
                    visibility TEXT NOT NULL DEFAULT 'org',
                    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    ingest_lock BOOLEAN NOT NULL DEFAULT FALSE,
                    last_verified_at TIMESTAMPTZ,
                    sha256 TEXT NOT NULL,
                    concept_ids JSONB NOT NULL DEFAULT '[]'::jsonb
                )
                """
            )
            await conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_raw_records_source_received "
                "ON raw_records (source, ingested_at DESC)"
            )
            await conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_raw_records_visibility "
                "ON raw_records (visibility)"
            )
            await conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_raw_records_ingested "
                "ON raw_records (ingested_at DESC)"
            )
            await conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_raw_records_ingest_lock "
                "ON raw_records (ingest_lock) WHERE ingest_lock = TRUE"
            )
        logger.info("postgres_store.schema_initialized")

    # ------------------------------------------------------------------
    # CRUD operations
    # ------------------------------------------------------------------

    async def insert(self, record: PostgresRawRecord) -> str:
        """raw_records INSERT (또는 ON CONFLICT UPDATE).

        Returns:
            raw_id
        """
        assert self._pool is not None, "pool not initialized"
        body_b64 = base64.b64encode(record.body).decode("ascii")
        frontmatter_json = json.dumps(record.frontmatter)
        concept_ids_json = json.dumps(record.concept_ids)

        async with self._pool.acquire() as conn:
            await conn.execute(
                """
                INSERT INTO raw_records
                    (raw_id, source, type, name, body, frontmatter, owner_org_id,
                     visibility, ingested_at, ingest_lock, last_verified_at, sha256, concept_ids)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
                ON CONFLICT (raw_id) DO UPDATE SET
                    source = EXCLUDED.source,
                    type = EXCLUDED.type,
                    name = EXCLUDED.name,
                    body = EXCLUDED.body,
                    frontmatter = EXCLUDED.frontmatter,
                    owner_org_id = EXCLUDED.owner_org_id,
                    visibility = EXCLUDED.visibility,
                    ingested_at = EXCLUDED.ingested_at,
                    last_verified_at = EXCLUDED.last_verified_at,
                    sha256 = EXCLUDED.sha256,
                    concept_ids = EXCLUDED.concept_ids
                """,
                record.raw_id,
                record.source,
                record.type,
                record.name,
                body_b64.encode("ascii"),
                frontmatter_json,
                record.owner_org_id,
                record.visibility,
                record.ingested_at,
                record.ingest_lock,
                record.last_verified_at,
                record.sha256,
                concept_ids_json,
            )
        logger.info(
            "postgres_store.inserted",
            raw_id=record.raw_id,
            source=record.source,
            type=record.type,
        )
        return record.raw_id

    async def select(self, raw_id: str) -> PostgresRawRecord | None:
        """raw_id 기준 SELECT.

        Returns:
            PostgresRawRecord | None (없으면 None)
        """
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                SELECT raw_id, source, type, name, body, frontmatter, owner_org_id,
                       visibility, ingested_at, ingest_lock, last_verified_at, sha256, concept_ids
                FROM raw_records
                WHERE raw_id = $1
                """,
                raw_id,
            )
        if row is None:
            return None

        return _row_to_record(row)

    async def list_by_source(
        self,
        source: str,
        limit: int = 100,
        since: datetime | None = None,
    ) -> list[PostgresRawRecord]:
        """source 기준 LIST (since optional filter).

        Returns:
            list[PostgresRawRecord]
        """
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            if since is None:
                rows = await conn.fetch(
                    """
                    SELECT raw_id, source, type, name, body, frontmatter, owner_org_id,
                           visibility, ingested_at, ingest_lock, last_verified_at, sha256, concept_ids
                    FROM raw_records
                    WHERE source = $1
                    ORDER BY ingested_at DESC
                    LIMIT $2
                    """,
                    source,
                    limit,
                )
            else:
                rows = await conn.fetch(
                    """
                    SELECT raw_id, source, type, name, body, frontmatter, owner_org_id,
                           visibility, ingested_at, ingest_lock, last_verified_at, sha256, concept_ids
                    FROM raw_records
                    WHERE source = $1 AND ingested_at >= $2
                    ORDER BY ingested_at DESC
                    LIMIT $3
                    """,
                    source,
                    since,
                    limit,
                )
        return [_row_to_record(row) for row in rows]

    async def list_ingest_locked(self, limit: int = 100) -> list[PostgresRawRecord]:
        """ingest_lock = TRUE 인 row LIST (Pi periodic ingest pipeline 용).

        M-v0.2.3+ PoC 운영: §10.3 Pi ingest pipeline 의 SELECT raw_records → set ingest_lock
        → decrypt → Pi LLM normalize → emit concept → update ingest_lock = FALSE 순환.
        """
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(
                """
                SELECT raw_id, source, type, name, body, frontmatter, owner_org_id,
                       visibility, ingested_at, ingest_lock, last_verified_at, sha256, concept_ids
                FROM raw_records
                WHERE ingest_lock = TRUE
                ORDER BY ingested_at ASC
                LIMIT $1
                """,
                limit,
            )
        return [_row_to_record(row) for row in rows]

    async def update_ingest_lock(self, raw_id: str, locked: bool) -> None:
        """ingest_lock 갱신 (Pi periodic ingest pipeline 용 atomic operation)."""
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            await conn.execute(
                "UPDATE raw_records SET ingest_lock = $1 WHERE raw_id = $2",
                locked,
                raw_id,
            )

    async def delete(self, raw_id: str) -> bool:
        """raw_id 기준 DELETE.

        Returns:
            True if deleted, False if not found
        """
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            result = await conn.execute(
                "DELETE FROM raw_records WHERE raw_id = $1",
                raw_id,
            )
        # result: "DELETE N" where N is rowcount
        deleted = result.endswith(" 1")
        if deleted:
            logger.info("postgres_store.deleted", raw_id=raw_id)
        return deleted

    async def verify_integrity(self, raw_id: str) -> bool:
        """body 의 sha256 해시 재계산 검증 (umbrella doc §4.7 정합).

        Returns:
            True if sha256 matches, False if mismatch or not found
        """
        assert self._pool is not None, "pool not initialized"
        record = await self.select(raw_id)
        if record is None:
            return False
        actual_sha256 = hashlib.sha256(record.body).hexdigest()
        return actual_sha256 == record.sha256

    async def update_concept_ids(self, raw_id: str, concept_ids: list[str]) -> None:
        """1 raw → N concepts traceability (§3.6.6.4 data lineage)."""
        assert self._pool is not None, "pool not initialized"
        async with self._pool.acquire() as conn:
            await conn.execute(
                """
                UPDATE raw_records
                SET concept_ids = $1, last_verified_at = NOW()
                WHERE raw_id = $2
                """,
                json.dumps(concept_ids),
                raw_id,
            )


# ----------------------------------------------------------------------
# Helpers
# ----------------------------------------------------------------------


def _row_to_record(row: Any) -> PostgresRawRecord:
    """asyncpg Record → PostgresRawRecord 변환."""
    # body 는 base64 encoded BYTEA
    body_b64 = row["body"]
    if isinstance(body_b64, str):
        body_bytes = base64.b64decode(body_b64.encode("ascii"))
    else:
        body_bytes = bytes(body_b64)

    frontmatter = row["frontmatter"]
    if isinstance(frontmatter, str):
        frontmatter = json.loads(frontmatter)

    concept_ids = row["concept_ids"]
    if isinstance(concept_ids, str):
        concept_ids = json.loads(concept_ids)

    return PostgresRawRecord(
        raw_id=row["raw_id"],
        source=row["source"],
        type=row["type"],
        name=row["name"],
        body=body_bytes,
        frontmatter=frontmatter,
        owner_org_id=row["owner_org_id"],
        visibility=row["visibility"],
        ingested_at=row["ingested_at"],
        ingest_lock=row["ingest_lock"],
        last_verified_at=row["last_verified_at"],
        sha256=row["sha256"],
        concept_ids=concept_ids,
    )


def raw_record_to_postgres(record: RawRecord) -> PostgresRawRecord:
    """RawRecord (file mode) → PostgresRawRecord 변환.

    봉투 암호화 body 의 envelope 유지 (§4.7 sha256 + 봉투 정합성).
    """
    body_bytes = record.body.encode("utf-8") if isinstance(record.body, str) else record.body
    return PostgresRawRecord(
        raw_id=record.raw_id,
        source=record.source,
        type=record.type,
        name=record.name,
        body=body_bytes,
        frontmatter=record.frontmatter,
        owner_org_id=record.owner_org_id,
        visibility=record.visibility,
        ingested_at=record.ingested_at,
        ingest_lock=False,
        last_verified_at=datetime.now(timezone.utc),
        sha256=record.sha256,
        concept_ids=record.concept_ids,
    )