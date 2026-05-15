#!/bin/bash
#
# register-devhub-client.sh — Hydra OIDC client 등록 (PS 카운터파트의 bash 버전)
#
# 목적: ADR-0001 §8.2 결정에 따라 DevHub frontend 를 Hydra 의 first-party
#       OIDC client 로 등록한다 (silent consent + PKCE + public client).
#
# 구현 노트: Hydra v26 CLI 의 `hydra create oauth2-client` 는 `--client-id` 를
#           무시하고 UUID 를 자동 발급해 PoC 재현성을 깨트린다. 따라서 본
#           스크립트는 Admin REST API (`POST /admin/clients`) 를 직접 호출해
#           client_id 를 명시한다.
#
# 전제 조건:
#   - Hydra admin (기본 http://localhost:4445) 이 응답 중
#   - `curl` 사용 가능 (macOS / Linux 표준)
#
# 사용:
#   ./infra/idp/scripts/register-devhub-client.sh
#
# 환경변수:
#   HYDRA_ADMIN_URL  Hydra admin endpoint (기본 http://localhost:4445)
#   DEVHUB_OIDC_CLIENT_ID  등록할 client id (기본 devhub-frontend)
#   DEVHUB_OIDC_REDIRECT_URI  redirect uri (기본 http://localhost:3000/auth/callback)
#   DEVHUB_OIDC_POST_LOGOUT_URI  post logout redirect uri (기본 http://localhost:3000/)
#
# 운영 진입 시 redirect / post-logout URI 와 secret rotation 정책 재검토.

set -euo pipefail

ADMIN_URL="${HYDRA_ADMIN_URL:-http://localhost:4445}"
CLIENT_ID="${DEVHUB_OIDC_CLIENT_ID:-devhub-frontend}"
REDIRECT_URI="${DEVHUB_OIDC_REDIRECT_URI:-http://localhost:3000/auth/callback}"
POST_LOGOUT_URI="${DEVHUB_OIDC_POST_LOGOUT_URI:-http://localhost:3000/}"

BODY=$(cat <<EOF
{
  "client_id": "$CLIENT_ID",
  "client_name": "DevHub Frontend (first-party)",
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "redirect_uris": ["$REDIRECT_URI"],
  "post_logout_redirect_uris": ["$POST_LOGOUT_URI"],
  "scope": "openid offline_access email profile",
  "token_endpoint_auth_method": "none",
  "skip_consent": true,
  "skip_logout_consent": true
}
EOF
)

echo "Hydra admin URL: $ADMIN_URL"
echo "Registering OIDC client: $CLIENT_ID"

# 기존 client 가 있으면 DELETE → POST 로 교체 (PS 카운터파트와 동일 의미).
http_status=$(curl -s -o /dev/null -w "%{http_code}" "$ADMIN_URL/admin/clients/$CLIENT_ID")
if [ "$http_status" = "200" ]; then
    echo "Existing client '$CLIENT_ID' found; deleting before recreate."
    curl -sf -X DELETE "$ADMIN_URL/admin/clients/$CLIENT_ID" >/dev/null
elif [ "$http_status" != "404" ]; then
    echo "Existence check returned $http_status; proceeding anyway." >&2
fi

curl -sf -X POST -H "Content-Type: application/json" -d "$BODY" "$ADMIN_URL/admin/clients" >/dev/null

echo "Created client '$CLIENT_ID'."
echo "Verify: hydra list oauth2-clients --endpoint $ADMIN_URL"
