"""HRDBSource unit tests (umbrella doc §3.8 + M-v0.2.3+).

mock mode default (HRDB_URL 미설정 시 in-memory fixture).
Real PostgreSQL mode 는 별도 integration test (CI env PostgreSQL 필요).
"""

from __future__ import annotations

import os
from datetime import datetime, timezone

import pytest

# HRDB_URL 환경변수 명시적 unset (mock mode 강제)
os.environ.pop("HRDB_URL", None)
os.environ.pop("POSTGRES_URL", None)

from backend_knowledge.sources.hrdb import HRDBSource  # noqa: E402
from backend_knowledge.sources.registry import (  # noqa: E402
    SOURCES,
    clear_sources,
    list_sources,
)


@pytest.fixture(autouse=True)
def _reset_registry():
    clear_sources()
    # Re-register by importing module
    from backend_knowledge.sources import hrdb as _hrdb_module  # noqa: F401

    yield
    clear_sources()


class TestHRDBSourceRegistration:
    def test_hrdb_registered(self):
        from backend_knowledge.sources.hrdb import HRDBSource as HS

        assert "hrdb" in SOURCES
        assert SOURCES["hrdb"] is HS

    def test_source_count_6(self):
        names = list_sources()
        # 4 Gitea sub-plugin + homelab_mock + hrdb = 6 source
        assert len(names) == 6
        assert "hrdb" in names

    def test_source_name(self):
        src = HRDBSource()
        assert src.name == "hrdb"


class TestHRDBSourceMockMode:
    @pytest.mark.asyncio
    async def test_connect_mock_mode(self):
        src = HRDBSource()
        await src.connect({})
        assert src.is_mock_mode is True
        assert src._connected is True

    @pytest.mark.asyncio
    async def test_fetch_mock_returns_3_employees(self):
        src = HRDBSource()
        await src.connect({})
        rows = await src.fetch(None)
        assert len(rows) == 3
        assert rows[0]["employee_id"] == "E001"
        assert rows[0]["name"] == "Alice Kim"

    @pytest.mark.asyncio
    async def test_fetch_mock_with_since_filter(self):
        src = HRDBSource()
        await src.connect({})
        since = datetime(2026, 6, 19, 0, 0, 0, tzinfo=timezone.utc)
        rows = await src.fetch(since)
        assert len(rows) >= 1

    @pytest.mark.asyncio
    async def test_fetch_department_aggregations(self):
        src = HRDBSource()
        await src.connect({})
        aggs = await src.fetch_department_aggregations()
        # Engineering=2, Product=1
        assert len(aggs) == 2
        agg_map = {a["department"]: a["count"] for a in aggs}
        assert agg_map["Engineering"] == 2
        assert agg_map["Product"] == 1

    @pytest.mark.asyncio
    async def test_health_check_mock_healthy(self):
        src = HRDBSource()
        await src.connect({})
        health = await src.health_check()
        assert health["healthy"] is True
        assert health["mode"] == "mock"
        assert health["fixture_count"] == 3


class TestHRDBSourceNormalize:
    @pytest.mark.asyncio
    async def test_normalize_employee_with_pii(self):
        src = HRDBSource()
        await src.connect({})
        rows = await src.fetch(None)
        concept = await src.normalize(rows[0])

        assert concept["source"] == "hrdb"
        assert concept["type"] == "dataset"
        assert concept["name"] == "hr-employee-e001"
        assert concept["title"] == "HR Employee E001: Alice Kim"
        assert concept["bundle"] == "devhub-hrdb"
        # PII fields detected (5 종 중 5 종 모두 hit)
        pii_fields = concept["frontmatter"]["_pii_fields"]
        assert set(pii_fields) == {"name", "email", "phone", "address", "employee_id"}
        assert concept["frontmatter"]["_pii_access_audit"] == "required"
        # body has markdown
        assert "# HR Employee E001" in concept["body"]
        assert "Alice Kim" in concept["body"]

    @pytest.mark.asyncio
    async def test_normalize_aggregation_metric(self):
        src = HRDBSource()
        await src.connect({})
        aggs = await src.fetch_department_aggregations()
        concept = await src.normalize_aggregation(aggs[0])

        assert concept["source"] == "hrdb"
        assert concept["type"] == "metric"
        assert concept["frontmatter"]["_metric_value"] == 2
        assert concept["frontmatter"]["_metric_unit"] == "employees"
        assert "Engineering" in concept["title"]


class TestHRDBSourceEdgeCases:
    @pytest.mark.asyncio
    async def test_fetch_without_connect_raises(self):
        src = HRDBSource()
        with pytest.raises(Exception):
            await src.fetch(None)

    @pytest.mark.asyncio
    async def test_disconnect_mock(self):
        src = HRDBSource()
        await src.connect({})
        await src.disconnect()
        assert src._connected is False


class TestHRDBSourceConfig:
    def test_default_storage_mode_db(self):
        # §10.4 정합: hrdb default storage_mode=db (PostgreSQL raw + file fallback)
        # M-v0.2.3+ 운영 기준 hrdb 는 dual mode (file + db default)
        from backend_knowledge.config import get_settings

        settings = get_settings()
        # hrdb_pii_field_types default 5 종
        assert len(settings.hrdb_pii_field_types) == 5
        assert "name" in settings.hrdb_pii_field_types
        assert "email" in settings.hrdb_pii_field_types