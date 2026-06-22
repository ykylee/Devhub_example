"""HR DB source plugin (umbrella doc §3.8 + §10.4 + M-v0.2.3+).

External system: HR DB (PostgreSQL) — employee/department/position records.
M-v0.2.3+ 운영 기준 7종 source 중 1종 (Gitea 4 + homelab + metrics + hrdb).

Mode:
- real (HRDB_URL 설정): asyncpg connection pool 로 PostgreSQL 직접 query
- mock (HRDB_URL 미설정): in-memory fixture (3 employee records + 3 department aggregations)

PII field 자동 detection (5 종, §3.6.6.5):
- name, email, phone, address, employee_id
- frontmatter `_pii_fields` list 로 노출 → curation 시 access log 별도 storage (§3.6.6.5)
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from ..config import get_settings
from ..logger import get_logger
from ._base import ConceptDict, SourcePlugin, SourcePluginError
from .registry import register_source

logger = get_logger(__name__)


# In-memory fixture (mock mode default, M-v0.2.3+ PoC)
_MOCK_EMPLOYEES: list[dict[str, Any]] = [
    {
        "id": 1,
        "employee_id": "E001",
        "name": "Alice Kim",
        "email": "alice@example.com",
        "phone": "+82-10-1234-5678",
        "address": "Seoul, Korea",
        "department_id": "D001",
        "department": "Engineering",
        "position": "Senior Engineer",
        "updated_at": "2026-06-19T00:00:00+00:00",
    },
    {
        "id": 2,
        "employee_id": "E002",
        "name": "Bob Lee",
        "email": "bob@example.com",
        "phone": "+82-10-2345-6789",
        "address": "Busan, Korea",
        "department_id": "D001",
        "department": "Engineering",
        "position": "Engineer",
        "updated_at": "2026-06-19T00:00:00+00:00",
    },
    {
        "id": 3,
        "employee_id": "E003",
        "name": "Charlie Park",
        "email": "charlie@example.com",
        "phone": "+82-10-3456-7890",
        "address": "Incheon, Korea",
        "department_id": "D002",
        "department": "Product",
        "position": "Product Manager",
        "updated_at": "2026-06-19T00:00:00+00:00",
    },
]


@register_source
class HRDBSource(SourcePlugin):
    """HR DB PostgreSQL source plugin.

    mode:
        - HRDB_URL 설정 시: asyncpg connection pool 로 PostgreSQL 직접 query
        - HRDB_URL 미설정 시: in-memory _MOCK_EMPLOYEES fixture (3 employee records)

    normalize: employee raw row → ConceptDict (type=dataset) + department aggregation → ConceptDict (type=metric)
    """

    name: str = "hrdb"
    query_table: str = "employees"
    aggregation_table: str = "employees"  # GROUP BY department_id

    def __init__(self) -> None:
        self.settings = get_settings()
        self._pool: Any = None
        self._connected: bool = False
        self._mock_mode: bool = False
        self._last_error: dict | None = None

    @property
    def is_mock_mode(self) -> bool:
        return self._mock_mode

    async def connect(self, credential: dict) -> None:
        """HR DB 연결 (mock mode 자동 fallback)."""
        url = self.settings.hrdb_url
        if not url:
            self._mock_mode = True
            self._connected = True
            logger.info(
                "hrdb.connect.mock_mode",
                hint="HRDB_URL 미설정 → in-memory fixture (3 employee records)",
            )
            return

        try:
            import asyncpg  # type: ignore[import-not-found]
        except ImportError as e:
            raise SourcePluginError(
                "asyncpg not installed. run: pip install 'asyncpg>=0.29.0'"
            ) from e

        try:
            self._pool = await asyncpg.create_pool(
                dsn=url,
                min_size=1,
                max_size=5,
                command_timeout=self.settings.hrdb_timeout_seconds,
            )
            self._connected = True
            self._mock_mode = False
            logger.info(
                "hrdb.connect.real",
                schema=self.settings.hrdb_schema,
                pool_size=5,
            )
        except Exception as e:
            self._last_error = {"code": "connect_failed", "message": str(e)}
            raise SourcePluginError(f"HRDB connection failed: {e}") from e

    async def disconnect(self) -> None:
        """Connection pool 종료."""
        if self._pool is not None:
            await self._pool.close()
            self._pool = None
            self._connected = False
            logger.info("hrdb.disconnected")

    async def fetch(self, since: datetime | None) -> list[dict]:
        """HR DB 에서 employee records fetch.

        Args:
            since: last sync timestamp (None = full sync)

        Returns:
            list of raw employee dicts (PostgreSQL row → dict)
        """
        if not self._connected:
            raise SourcePluginError("not connected. call connect() first")

        if self._mock_mode:
            return self._fetch_mock(since)

        assert self._pool is not None
        sql = f"""
            SELECT id, employee_id, name, email, phone, address,
                   department_id, department, position, updated_at
            FROM {self.settings.hrdb_schema}.{self.query_table}
            WHERE ($1::timestamptz IS NULL OR updated_at >= $1)
            ORDER BY updated_at DESC
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql, since)
        return [dict(row) for row in rows]

    def _fetch_mock(self, since: datetime | None) -> list[dict]:
        """In-memory mock fetch (3 employee records)."""
        if since is None:
            return list(_MOCK_EMPLOYEES)
        return [e for e in _MOCK_EMPLOYEES if e["updated_at"] >= since.isoformat()]

    async def fetch_department_aggregations(self) -> list[dict]:
        """Department 별 employee count aggregation (type=metric concept).

        Returns:
            list of {department, count} dicts
        """
        if self._mock_mode:
            agg: dict[str, int] = {}
            for e in _MOCK_EMPLOYEES:
                dept = e["department"]
                agg[dept] = agg.get(dept, 0) + 1
            return [{"department": k, "count": v} for k, v in sorted(agg.items())]

        assert self._pool is not None
        sql = f"""
            SELECT department, COUNT(*) AS count
            FROM {self.settings.hrdb_schema}.{self.aggregation_table}
            GROUP BY department
            ORDER BY count DESC
        """
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(sql)
        return [dict(row) for row in rows]

    async def normalize(self, raw: dict) -> ConceptDict:
        """HR DB employee raw row → ConceptDict (type=dataset).

        PII fields (5 종, §3.6.6.5):
        - name / email / phone / address / employee_id
        - frontmatter `_pii_fields` list 로 노출 → curation 시 access log 별도 storage

        body 에는 PII field 값 포함 (access 는 Path Y user/org 기반 owner_org_id match).
        """
        employee_id = raw.get("employee_id", "")
        name = raw.get("name", "")
        email = raw.get("email", "")
        department = raw.get("department", "")
        position = raw.get("position", "")
        updated_at = raw.get("updated_at", "")

        pii_fields: list[str] = []
        for field_name in self.settings.hrdb_pii_field_types:
            if raw.get(field_name):
                pii_fields.append(field_name)

        md_body = f"# HR Employee {employee_id}\n\n"
        md_body += f"**Name**: {name}\n\n"
        md_body += f"**Email**: {email}\n\n"
        md_body += f"**Department**: {department}\n\n"
        md_body += f"**Position**: {position}\n\n"
        md_body += f"**Updated**: {updated_at}\n\n"
        if pii_fields:
            md_body += f"**PII Fields**: {', '.join(pii_fields)} (5 종 자동 detection)\n"

        return ConceptDict(
            source=self.name,
            type="dataset",
            name=f"hr-employee-{employee_id.lower()}",
            title=f"HR Employee {employee_id}: {name}",
            body=md_body,
            frontmatter={
                "tags": ["hrdb", "employee", department.lower().replace(" ", "-")],
                "description": f"HR DB employee record: {name} ({position}, {department})",
                "_pii_fields": pii_fields,
                "_pii_access_audit": "required",
            },
            raw_refs=[],
            timestamp=updated_at,
            bundle="devhub-hrdb",
        )

    async def normalize_aggregation(self, raw: dict) -> ConceptDict:
        """Department aggregation raw → ConceptDict (type=metric)."""
        department = raw.get("department", "unknown")
        count = raw.get("count", 0)

        md_body = f"# HR Department Headcount: {department}\n\n"
        md_body += f"**Department**: {department}\n\n"
        md_body += f"**Employee Count**: {count}\n\n"
        md_body += f"**Source**: HR DB aggregation ({self.aggregation_table} GROUP BY department)\n"

        return ConceptDict(
            source=self.name,
            type="metric",
            name=f"hr-dept-headcount-{department.lower().replace(' ', '-')}",
            title=f"HR Department Headcount: {department}",
            body=md_body,
            frontmatter={
                "tags": ["hrdb", "metric", "headcount", department.lower().replace(" ", "-")],
                "description": f"Department headcount metric: {department} = {count} employees",
                "_metric_value": count,
                "_metric_unit": "employees",
            },
            raw_refs=[],
            timestamp=datetime.now(timezone.utc).isoformat(),
            bundle="devhub-hrdb",
        )

    async def health_check(self) -> dict:
        """Connectivity check."""
        try:
            if not self._connected:
                await self.connect({})
            if self._mock_mode:
                return {
                    "healthy": True,
                    "mode": "mock",
                    "fixture_count": len(_MOCK_EMPLOYEES),
                    "last_error": None,
                }
            assert self._pool is not None
            async with self._pool.acquire() as conn:
                result = await conn.fetchval("SELECT 1")
            return {
                "healthy": result == 1,
                "mode": "real",
                "pool_size": 5,
                "last_error": None,
            }
        except Exception as e:
            return {
                "healthy": False,
                "mode": "unknown",
                "last_error": {"code": "health_check_failed", "message": str(e)},
            }