"""Path Y middleware test (umbrella doc §3.6.1)."""

from __future__ import annotations

import base64
import json
from datetime import datetime, timezone, timedelta

import pytest

from backend_knowledge.auth.path_y import (
    PathYExpiredError,
    PathYUserContext,
    PathYValidationError,
    PathYValidator,
)


class TestPathYValidator:
    """Test PathYValidator (format 검증 only, 인증은 caller 책임)."""

    def test_valid_header_passes(self) -> None:
        """Valid 8-field Path Y header should pass validation."""
        payload = {
            "version": "v0",
            "user_id": "u_abc123",
            "org_id": "ou_root_dept_a",
            "org_unit_ids": ["ou_root_dept_a"],
            "project_ids": ["prj_x"],
            "roles": ["developer"],
            "request_id": "req_20260618_001",
            "issued_at": datetime.now(timezone.utc).isoformat(),
        }
        json_str = json.dumps(payload)
        header_value = base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")

        validator = PathYValidator()
        ctx = validator.validate(header_value)
        assert ctx.user_id == "u_abc123"
        assert ctx.org_id == "ou_root_dept_a"
        assert ctx.roles == ["developer"]
        assert ctx.version == "v0"

    def test_missing_header_raises_validation_error(self) -> None:
        """Missing/empty header should raise PathYValidationError."""
        validator = PathYValidator()
        with pytest.raises(PathYValidationError, match="missing header value"):
            validator.validate(None)
        with pytest.raises(PathYValidationError, match="missing header value"):
            validator.validate("")

    def test_invalid_base64_raises_validation_error(self) -> None:
        """Invalid base64url should raise PathYValidationError."""
        validator = PathYValidator()
        with pytest.raises(PathYValidationError, match="base64url decode failed"):
            validator.validate("not-valid-base64-!!!")

    def test_invalid_json_raises_validation_error(self) -> None:
        """Invalid JSON (valid base64 but not JSON) should raise PathYValidationError."""
        not_json = base64.urlsafe_b64encode(b"not json at all").decode("ascii")
        validator = PathYValidator()
        with pytest.raises(PathYValidationError, match="JSON parse failed"):
            validator.validate(not_json)

    def test_missing_required_field_raises_validation_error(self) -> None:
        """Missing required field (e.g., user_id) should raise PathYValidationError."""
        payload = {
            "version": "v0",
            # user_id missing!
            "org_id": "ou_root_dept_a",
            "org_unit_ids": [],
            "project_ids": [],
            "roles": [],
            "request_id": "req_20260618_001",
            "issued_at": datetime.now(timezone.utc).isoformat(),
        }
        json_str = json.dumps(payload)
        header_value = base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")

        validator = PathYValidator()
        with pytest.raises(PathYValidationError, match="schema validation failed"):
            validator.validate(header_value)

    def test_extra_field_raises_validation_error(self) -> None:
        """Extra field (9th field) should raise (model_config extra='forbid')."""
        payload = {
            "version": "v0",
            "user_id": "u_abc123",
            "org_id": "ou_root_dept_a",
            "org_unit_ids": [],
            "project_ids": [],
            "roles": [],
            "request_id": "req_20260618_001",
            "issued_at": datetime.now(timezone.utc).isoformat(),
            "extra_field": "should be rejected",
        }
        json_str = json.dumps(payload)
        header_value = base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")

        validator = PathYValidator()
        with pytest.raises(PathYValidationError, match="schema validation failed"):
            validator.validate(header_value)

    def test_expired_raises_expired_error(self) -> None:
        """Expired header (issued_at > 5분 전) should raise PathYExpiredError."""
        expired_at = datetime.now(timezone.utc) - timedelta(minutes=10)
        payload = {
            "version": "v0",
            "user_id": "u_abc123",
            "org_id": "ou_root_dept_a",
            "org_unit_ids": [],
            "project_ids": [],
            "roles": [],
            "request_id": "req_20260618_001",
            "issued_at": expired_at.isoformat(),
        }
        json_str = json.dumps(payload)
        header_value = base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")

        validator = PathYValidator(max_age_seconds=300)
        with pytest.raises(PathYExpiredError) as exc_info:
            validator.validate(header_value)
        assert exc_info.value.age_seconds > 300

    def test_naive_datetime_assumed_utc(self) -> None:
        """Naive datetime (no tzinfo) should be assumed UTC."""
        # 1분 전의 naive datetime
        one_min_ago = (datetime.now(timezone.utc) - timedelta(minutes=1)).replace(tzinfo=None)
        payload = {
            "version": "v0",
            "user_id": "u_abc123",
            "org_id": "ou_root_dept_a",
            "org_unit_ids": [],
            "project_ids": [],
            "roles": [],
            "request_id": "req_20260618_001",
            "issued_at": one_min_ago.isoformat(),  # naive ISO 8601
        }
        json_str = json.dumps(payload)
        header_value = base64.urlsafe_b64encode(json_str.encode("utf-8")).decode("ascii").rstrip("=")

        validator = PathYValidator()
        ctx = validator.validate(header_value)
        # naive datetime treated as UTC → age ~60s, less than 300s max_age
        assert ctx.issued_at.tzinfo is not None
