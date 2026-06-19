"""Path Y caller-provided user context (umbrella doc §3.6.1 정합).

8 field schema:
- version: schema version ("v0")
- user_id: DevHub internal user PK
- org_id: primary organization
- org_unit_ids: org_head scope 의 subtree (recursive CTE 결과)
- project_ids: user 의 project memberships
- roles: system role + resource role 목록
- request_id: trace id (envelope trace_id 와 동일값 권장)
- issued_at: context 생성 시각. 만료 검증은 issued_at + PATH_Y_MAX_AGE_SECONDS (300초) 별도.

X-DevHub-User-Context header: base64url(json).

format 검증만 backend-knowledge 책임 (§3.6.1):
- JSON parse
- schema check (8 field 1:1)
- issued_at 만료 (5분)
"""

from __future__ import annotations

import base64
import binascii
import json
from datetime import datetime, timedelta, timezone
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator

# Default max age (5 minutes) per umbrella doc §3.6.1
DEFAULT_PATH_Y_MAX_AGE_SECONDS: int = 300


# 8 field schema (umbrella doc §3.6.1 1:1 정합)
class PathYUserContext(BaseModel):
    """X-DevHub-User-Context header 의 decoded JSON schema (8 field 1:1 정합)."""

    model_config = ConfigDict(
        extra="forbid",  # 8 field 외 reject (strict)
        populate_by_name=True,
    )

    version: Literal["v0"] = "v0"
    user_id: str = Field(..., min_length=1)
    org_id: str = Field(..., min_length=1)
    org_unit_ids: list[str] = Field(default_factory=list)
    project_ids: list[str] = Field(default_factory=list)
    roles: list[str] = Field(default_factory=list)
    request_id: str = Field(..., min_length=1)
    issued_at: datetime  # ISO 8601 with timezone

    @field_validator("issued_at", mode="before")
    @classmethod
    def _parse_issued_at(cls, v: Any) -> datetime:
        """Parse ISO 8601 with timezone. naive datetime 은 UTC 가정."""
        if isinstance(v, datetime):
            return v if v.tzinfo else v.replace(tzinfo=timezone.utc)
        if isinstance(v, str):
            # Python 3.11+ fromisoformat handles most ISO 8601
            try:
                dt = datetime.fromisoformat(v.replace("Z", "+00:00"))
                return dt if dt.tzinfo else dt.replace(tzinfo=timezone.utc)
            except ValueError as e:
                raise ValueError(f"invalid issued_at format: {v}") from e
        raise ValueError(f"issued_at must be ISO 8601 string or datetime, got {type(v)}")


class PathYExpiredError(Exception):
    """Raised when X-DevHub-User-Context 의 issued_at 이 만료 (5분 초과)."""

    def __init__(self, age_seconds: float, max_age_seconds: int):
        self.age_seconds = age_seconds
        self.max_age_seconds = max_age_seconds
        super().__init__(f"X-DevHub-User-Context expired: age={age_seconds:.1f}s, max={max_age_seconds}s")


class PathYValidationError(Exception):
    """Raised when X-DevHub-User-Context 의 base64url/json/schema 가 invalid."""

    def __init__(self, reason: str):
        self.reason = reason
        super().__init__(f"X-DevHub-User-Context invalid: {reason}")


class PathYValidator:
    """Path Y header validator (format 검증 only, 인증은 caller 책임).

    Singleton pattern via get_path_y_validator().
    """

    def __init__(self, max_age_seconds: int = DEFAULT_PATH_Y_MAX_AGE_SECONDS):
        self.max_age_seconds = max_age_seconds

    def validate(self, header_value: str | None) -> PathYUserContext:
        """Validate X-DevHub-User-Context header value.

        Args:
            header_value: base64url(json) string (None or empty → PathYValidationError)

        Returns: PathYUserContext (decoded + schema-validated + 만료 검증 pass)

        Raises:
            PathYValidationError: base64url decode / JSON parse / schema 검증 실패
            PathYExpiredError: issued_at + max_age_seconds 초과
        """
        if not header_value:
            raise PathYValidationError("missing header value")

        # 1. base64url decode
        try:
            # base64url with padding optional
            padding = "=" * (-len(header_value) % 4)
            decoded = base64.urlsafe_b64decode(header_value + padding)
            text = decoded.decode("utf-8")
        except (binascii.Error, UnicodeDecodeError) as e:
            raise PathYValidationError(f"base64url decode failed: {e}") from e

        # 2. JSON parse
        try:
            data = json.loads(text)
        except json.JSONDecodeError as e:
            raise PathYValidationError(f"JSON parse failed: {e}") from e

        if not isinstance(data, dict):
            raise PathYValidationError(f"JSON root must be object, got {type(data).__name__}")

        # 3. schema check (Pydantic v2 validation)
        try:
            ctx = PathYUserContext(**data)
        except Exception as e:
            raise PathYValidationError(f"schema validation failed: {e}") from e

        # 4. 만료 검증 (issued_at + max_age_seconds)
        now = datetime.now(timezone.utc)
        age = (now - ctx.issued_at).total_seconds()
        if age > self.max_age_seconds:
            raise PathYExpiredError(age_seconds=age, max_age_seconds=self.max_age_seconds)

        return ctx


# Singleton instance
_validator: PathYValidator | None = None


def get_path_y_validator() -> PathYValidator:
    """Get singleton PathYValidator."""
    global _validator
    if _validator is None:
        from backend_knowledge.config import get_settings
        settings = get_settings()
        _validator = PathYValidator(max_age_seconds=settings.path_y_max_age_seconds)
    return _validator


# Re-export constants
PATH_Y_MAX_AGE_SECONDS = DEFAULT_PATH_Y_MAX_AGE_SECONDS


# FastAPI dependency (used by ingest.py / health.py via Depends())
def require_path_y_context(
    x_devhub_user_context: str | None = None,
) -> PathYUserContext:
    """FastAPI dependency for Path Y (REQUIRED).

    Usage:
        from fastapi import Depends
        @router.get("/protected")
        async def handler(ctx: PathYUserContext = Depends(require_path_y_context)):
            ...

    Raises HTTPException 400 E_VALIDATION if X-DevHub-User-Context header missing or invalid.
    Raises HTTPException 401 E_UNAUTHORIZED if header expired.

    Note: This signature uses `x_devhub_user_context` parameter name which FastAPI
    maps to the `X-DevHub-User-Context` HTTP header. The actual extraction logic
    is in api/ingest.py:get_path_y_context (called by main.py dependencies).
    """
    from fastapi import HTTPException, status

    if not x_devhub_user_context:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": "X-DevHub-User-Context required"},
        )
    validator = get_path_y_validator()
    try:
        return validator.validate(x_devhub_user_context)
    except PathYExpiredError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"code": "E_UNAUTHORIZED", "message": f"X-DevHub-User-Context expired: {e}"},
        )
    except PathYValidationError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"code": "E_VALIDATION", "message": f"X-DevHub-User-Context invalid: {e.reason}"},
        )
