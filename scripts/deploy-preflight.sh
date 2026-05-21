#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.deploy.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/docs/setup/deploy.env.example}"

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "ERROR: compose file not found: $COMPOSE_FILE" >&2
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: env file not found: $ENV_FILE" >&2
  exit 1
fi

# shellcheck disable=SC1090
set -a
source "$ENV_FILE"
set +a

required_vars=(
  IMAGE_TAG
  DB_URL
  DEVHUB_PUBLIC_BASE_URL
  DEVHUB_OIDC_ISSUER_URL
  DEVHUB_OIDC_CLIENT_SECRET
  DEVHUB_KEYCLOAK_ADMIN_URL
  DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET
  OIDC_ISSUER_URL
  OIDC_REDIRECT_URI
  NEXT_PUBLIC_OIDC_ISSUER_URL
  NEXT_PUBLIC_OIDC_REDIRECT_URI
)

for v in "${required_vars[@]}"; do
  if [ -z "${!v:-}" ]; then
    echo "ERROR: required env var is empty: $v" >&2
    exit 1
  fi
done

if [ "${DEVHUB_AUTH_DEV_FALLBACK:-0}" != "0" ]; then
  echo "ERROR: DEVHUB_AUTH_DEV_FALLBACK must be 0 in deploy profile" >&2
  exit 1
fi

if [[ "$OIDC_REDIRECT_URI" != */auth/callback ]]; then
  echo "ERROR: OIDC_REDIRECT_URI must end with /auth/callback" >&2
  exit 1
fi

if [[ "$NEXT_PUBLIC_OIDC_REDIRECT_URI" != */auth/callback ]]; then
  echo "ERROR: NEXT_PUBLIC_OIDC_REDIRECT_URI must end with /auth/callback" >&2
  exit 1
fi

public_base="${DEVHUB_PUBLIC_BASE_URL%/}"
base_path_raw="${NEXT_PUBLIC_BASE_PATH:-devhub}"
if [ -n "$base_path_raw" ]; then
  base_path="/${base_path_raw#/}"
  base_path="${base_path%/}"
else
  base_path=""
fi
expected_redirect_uri="${public_base}${base_path}/auth/callback"
if [[ "$OIDC_REDIRECT_URI" != "$expected_redirect_uri" ]]; then
  echo "ERROR: OIDC_REDIRECT_URI must equal ${expected_redirect_uri}" >&2
  exit 1
fi
if [[ "$NEXT_PUBLIC_OIDC_REDIRECT_URI" != "$expected_redirect_uri" ]]; then
  echo "ERROR: NEXT_PUBLIC_OIDC_REDIRECT_URI must equal ${expected_redirect_uri}" >&2
  exit 1
fi

if [[ "$DEVHUB_OIDC_ISSUER_URL" != "$OIDC_ISSUER_URL" ]]; then
  echo "WARN: DEVHUB_OIDC_ISSUER_URL and OIDC_ISSUER_URL differ." >&2
  echo "      This is valid only when issuer/JWKS split is intentional." >&2
fi

if [ -n "${DEVHUB_OIDC_JWKS_URL:-}" ]; then
  if ! [[ "$DEVHUB_OIDC_JWKS_URL" =~ /protocol/openid-connect/certs$ ]]; then
    echo "WARN: DEVHUB_OIDC_JWKS_URL does not end with /protocol/openid-connect/certs" >&2
  fi
fi

echo "[1/3] Docker availability check"
docker --version >/dev/null

docker compose version >/dev/null

echo "[2/3] Compose render validation"
(
  cd "$ROOT_DIR"
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/tmp/devhub_deploy_compose_rendered.yml
)

echo "[3/3] OIDC endpoint reachability checks"
curl -fsS "$OIDC_ISSUER_URL/.well-known/openid-configuration" >/dev/null
if [ -n "${DEVHUB_OIDC_JWKS_URL:-}" ]; then
  curl -fsS "$DEVHUB_OIDC_JWKS_URL" >/dev/null
fi

echo "preflight OK"
