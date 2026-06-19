"""Auth (Path Y caller-provided user context only).

umbrella doc §3.6.1 + ADR-0035 §3.4 정합:
- backend-knowledge 는 자체 인증 안 함 (OIDC/Keycloak/backend-core 인증 위임 ❌)
- X-DevHub-User-Context header (base64url(json), 8 field) 로 user/org/project/roles 전달
- format 검증만 backend-knowledge 책임: JSON parse + schema check + 만료 (5분)
- caller (gateway / 별도 agent) 가 authentication + context 구성 책임
"""

from .path_y import (
    PATH_Y_MAX_AGE_SECONDS,
    PathYExpiredError,
    PathYUserContext,
    PathYValidationError,
    PathYValidator,
    get_path_y_validator,
    require_path_y_context,
)

__all__ = [
    "PATH_Y_MAX_AGE_SECONDS",
    "PathYExpiredError",
    "PathYUserContext",
    "PathYValidationError",
    "PathYValidator",
    "get_path_y_validator",
    "require_path_y_context",
]
