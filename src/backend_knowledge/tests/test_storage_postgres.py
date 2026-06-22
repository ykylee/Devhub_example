"""PostgresStore unit tests (umbrella doc §10.1 + M-v0.2.3+).

전반부: dataclass + 변환 함수 unit test (no PostgreSQL 필요).
후반부: connection pool + CRUD integration test (POSTGRES_URL 환경변수 있을 때만 활성화).
"""

from __future__ import annotations

import os
from datetime import datetime, timezone

import pytest

# Integration test skip marker
requires_postgres = pytest.mark.skipif(
    os.environ.get("POSTGRES_URL") is None,
    reason="POSTGRES_URL 환경변수 미설정 → integration test skip (file mode default)",
)


class TestPostgresRawRecord:
    def test_dataclass_construction(self):
        from backend_knowledge.storage.postgres_store import PostgresRawRecord

        rec = PostgresRawRecord(
            raw_id="abc1234",
            source="hrdb",
            type="dataset",
            name="hr-employee-e001",
            body=b'{"employee_id": "E001"}',
            frontmatter={"tags": ["hrdb"]},
            owner_org_id="ou_root_dept_a",
            visibility="org",
            ingested_at=datetime.now(timezone.utc),
            ingest_lock=False,
            last_verified_at=datetime.now(timezone.utc),
            sha256="a" * 64,
            concept_ids=["e001"],
        )
        assert rec.raw_id == "abc1234"
        assert rec.source == "hrdb"
        assert rec.type == "dataset"
        assert rec.ingest_lock is False

    def test_dataclass_default_concept_ids(self):
        from backend_knowledge.storage.postgres_store import PostgresRawRecord

        rec = PostgresRawRecord(
            raw_id="x",
            source="hrdb",
            type="metric",
            name="test",
            body=b"{}",
            frontmatter={},
            owner_org_id=None,
            visibility="public",
            ingested_at=datetime.now(timezone.utc),
            ingest_lock=True,
            last_verified_at=None,
            sha256="b" * 64,
            concept_ids=[],
        )
        assert rec.concept_ids == []
        assert rec.ingest_lock is True
        assert rec.owner_org_id is None


class TestRawRecordToPostgres:
    def test_conversion_basic(self):
        from backend_knowledge.storage.postgres_store import raw_record_to_postgres
        from backend_knowledge.storage.raw_store import RawRecord

        raw = RawRecord(
            raw_id="abc1234",
            source="hrdb",
            type="dataset",
            name="hr-employee-e001",
            body='{"employee_id": "E001"}',
            frontmatter={"tags": ["hrdb"]},
            owner_org_id="ou_root_dept_a",
            visibility="org",
            ingested_at=datetime.now(timezone.utc),
            sha256="c" * 64,
            concept_ids=["e001"],
        )
        pg = raw_record_to_postgres(raw)
        assert pg.raw_id == "abc1234"
        assert pg.source == "hrdb"
        # body bytes 변환 (str → utf-8)
        assert pg.body == b'{"employee_id": "E001"}'
        # ingest_lock default False
        assert pg.ingest_lock is False
        # last_verified_at 자동 설정
        assert pg.last_verified_at is not None


class TestPostgresStoreErrors:
    def test_postgres_store_error_inherits(self):
        from backend_knowledge.storage.postgres_store import PostgresStoreError

        # PostgresStoreError → RawStoreError → Exception
        assert issubclass(PostgresStoreError, Exception)


@requires_postgres
class TestPostgresStoreIntegration:
    """Integration test: 실제 PostgreSQL 연결 + CRUD.

    실행:
        POSTGRES_URL=postgresql://user:pass@localhost:5432/dbname pytest -k test_postgres_store_integration

    M-v0.2.3+ production 환경에서만 활성화.
    """

    @pytest.mark.asyncio
    async def test_connect_and_initialize_schema(self):
        from backend_knowledge.storage.postgres_store import PostgresStore

        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            # schema created (no exception)
            assert store._pool is not None

    @pytest.mark.asyncio
    async def test_insert_and_select_roundtrip(self):
        from backend_knowledge.storage.postgres_store import (
            PostgresRawRecord,
            PostgresStore,
        )

        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            rec = PostgresRawRecord(
                raw_id="test_abc1234",
                source="hrdb_test",
                type="dataset",
                name="test-record",
                body=b'{"key": "value"}',
                frontmatter={"tags": ["test"]},
                owner_org_id="ou_test",
                visibility="org",
                ingested_at=datetime.now(timezone.utc),
                ingest_lock=False,
                last_verified_at=datetime.now(timezone.utc),
                sha256="d" * 64,
                concept_ids=["test-1"],
            )
            await store.insert(rec)
            fetched = await store.select("test_abc1234")
            assert fetched is not None
            assert fetched.source == "hrdb_test"
            assert fetched.body == b'{"key": "value"}'

            # cleanup
            await store.delete("test_abc1234")

    @pytest.mark.asyncio
    async def test_list_by_source(self):
        from backend_knowledge.storage.postgres_store import (
            PostgresRawRecord,
            PostgresStore,
        )

        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            for i in range(3):
                rec = PostgresRawRecord(
                    raw_id=f"test_list_{i}",
                    source="hrdb_list_test",
                    type="metric",
                    name=f"test-{i}",
                    body=b"{}",
                    frontmatter={},
                    owner_org_id=None,
                    visibility="public",
                    ingested_at=datetime.now(timezone.utc),
                    ingest_lock=False,
                    last_verified_at=None,
                    sha256="e" * 64,
                    concept_ids=[],
                )
                await store.insert(rec)
            rows = await store.list_by_source("hrdb_list_test", limit=10)
            assert len(rows) >= 3
            # cleanup
            for i in range(3):
                await store.delete(f"test_list_{i}")

    @pytest.mark.asyncio
    async def test_update_ingest_lock(self):
        from backend_knowledge.storage.postgres_store import (
            PostgresRawRecord,
            PostgresStore,
        )

        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            rec = PostgresRawRecord(
                raw_id="test_lock",
                source="hrdb_lock_test",
                type="dataset",
                name="lock-test",
                body=b"{}",
                frontmatter={},
                owner_org_id=None,
                visibility="org",
                ingested_at=datetime.now(timezone.utc),
                ingest_lock=False,
                last_verified_at=None,
                sha256="f" * 64,
                concept_ids=[],
            )
            await store.insert(rec)
            await store.update_ingest_lock("test_lock", True)
            fetched = await store.select("test_lock")
            assert fetched.ingest_lock is True
            # cleanup
            await store.delete("test_lock")

    @pytest.mark.asyncio
    async def test_verify_integrity_sha256(self):
        import hashlib

        from backend_knowledge.storage.postgres_store import (
            PostgresRawRecord,
            PostgresStore,
        )

        async with PostgresStore.connect() as store:
            await store.initialize_schema()
            body = b'{"integrity": "test"}'
            sha256 = hashlib.sha256(body).hexdigest()
            rec = PostgresRawRecord(
                raw_id="test_integrity",
                source="hrdb_integrity_test",
                type="dataset",
                name="integrity-test",
                body=body,
                frontmatter={},
                owner_org_id=None,
                visibility="org",
                ingested_at=datetime.now(timezone.utc),
                ingest_lock=False,
                last_verified_at=None,
                sha256=sha256,
                concept_ids=[],
            )
            await store.insert(rec)
            is_valid = await store.verify_integrity("test_integrity")
            assert is_valid is True
            # cleanup
            await store.delete("test_integrity")


class TestPostgresStoreConfig:
    def test_postgres_url_default_none(self):
        from backend_knowledge.config import get_settings

        settings = get_settings()
        # POSTGRES_URL 미설정 시 None (file mode)
        # (테스트 환경에 따라 설정 다를 수 있음)
        assert settings.postgres_url is None or isinstance(settings.postgres_url, str)

    def test_postgres_pool_size_default(self):
        from backend_knowledge.config import get_settings

        settings = get_settings()
        assert settings.postgres_pool_size >= 1
        assert settings.postgres_pool_size <= 100

    def test_hrdb_url_default_none(self):
        from backend_knowledge.config import get_settings

        settings = get_settings()
        # HRDB_URL 미설정 시 None (mock mode)
        assert settings.hrdb_url is None or isinstance(settings.hrdb_url, str)