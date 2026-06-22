"""Settings (pydantic-settings).

umbrella doc §3.3 + §3.6 + §11 정합:
- §3.3 Python 3.13+ / FastAPI / Pydantic v2
- §3.6 Path Y caller-provided user context (5분 만료)
- §11 운영 runbook (env var 기반 설정)
"""

from __future__ import annotations

from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """backend-knowledge runtime config.

    All env var 로 override 가능. None 또는 default 의미는 README §6 참조.
    """

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    # Path Y caller-provided user context (umbrella doc §3.6.1)
    path_y_max_age_seconds: int = Field(default=300, description="Path Y header 만료 (5분)")

    # Runtime data directory (umbrella doc §2.1)
    var_dir: Path = Field(default=Path("./var"), description="Runtime data dir (var/raw, var/bundles, var/audit, var/log)")

    # Logging (structlog)
    log_level: str = Field(default="INFO", description="structlog log level (DEBUG/INFO/WARNING/ERROR)")
    log_format: str = Field(default="json", description="json | console")

    # Gitea integration (umbrella doc §3.8)
    gitea_url: str | None = Field(default=None, description="Gitea instance URL (None 이면 mock mode)")
    gitea_token: str | None = Field(default=None, description="Gitea access token (None 이면 mock mode)")
    gitea_default_owner: str = Field(default="devhub", description="Gitea repo owner (default)")
    gitea_default_repo: str = Field(default="example", description="Gitea repo name (default)")
    gitea_timeout_seconds: float = Field(default=30.0, description="Gitea API timeout")

    # 봉투 암호화 (umbrella doc §3.6 / ADR-0025, scope = raw + .env/KEK 만)
    # None 이면 plaintext mode (PoC default). M-v0.2.0+ production 시 KEK 필수.
    raw_encryption_key: str | None = Field(default=None, description="봉투 암호화 KEK base64 (None 이면 plaintext)")

    # Audit log (umbrella doc §3.6.6.1)
    audit_log_retention_days: int = Field(default=7, description="Audit log retention (일)")

    # Monitoring (M-v0.2.0+ Prometheus)
    enable_metrics: bool = Field(default=False, description="/metrics endpoint 활성화")

    hrdb_url: str | None = Field(default=None, description="HR DB PostgreSQL URL (None = mock mode)")
    hrdb_schema: str = Field(default="public", description="HR DB schema (default: public)")
    hrdb_timeout_seconds: float = Field(default=30.0, description="HR DB query timeout")
    hrdb_pii_field_types: list[str] = Field(
        default_factory=lambda: ["name", "email", "phone", "address", "employee_id"],
        description="PII field 자동 detection 대상 (5 종, §3.6.6.5)",
    )

    postgres_url: str | None = Field(default=None, description="PostgreSQL raw storage URL (None = file mode)")
    postgres_pool_size: int = Field(default=5, description="PostgreSQL connection pool size")
    postgres_pool_max_overflow: int = Field(default=10, description="PostgreSQL connection pool max overflow")


_settings: Settings | None = None


def get_settings() -> Settings:
    """Singleton settings accessor."""
    global _settings
    if _settings is None:
        _settings = Settings()
    return _settings
