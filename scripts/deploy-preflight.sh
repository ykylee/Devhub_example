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

compose_profile_args=()
if [ -n "${COMPOSE_PROFILES:-}" ]; then
  IFS=',' read -r -a compose_profiles <<< "$COMPOSE_PROFILES"
  for profile in "${compose_profiles[@]}"; do
    profile="${profile// /}"
    if [ -n "$profile" ]; then
      compose_profile_args+=(--profile "$profile")
    fi
  done
fi

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

validate_cidr() {
  local cidr="$1"
  if ! [[ "$cidr" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}/([0-9]|[12][0-9]|3[0-2])$ ]]; then
    return 1
  fi

  local ip="${cidr%/*}"
  local octet
  IFS='.' read -r -a octets <<< "$ip"
  if [ "${#octets[@]}" -ne 4 ]; then
    return 1
  fi
  for octet in "${octets[@]}"; do
    if [ "$octet" -lt 0 ] || [ "$octet" -gt 255 ]; then
      return 1
    fi
  done
}

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

if ! validate_cidr "${KEYCLOAK_ADMIN_ALLOW_CIDR:-127.0.0.1/32}"; then
  echo "ERROR: KEYCLOAK_ADMIN_ALLOW_CIDR must be CIDR form (e.g. 127.0.0.1/32, 10.0.0.0/8)" >&2
  exit 1
fi

if [[ "$DEVHUB_OIDC_ISSUER_URL" != "$OIDC_ISSUER_URL" ]]; then
  echo "WARN: DEVHUB_OIDC_ISSUER_URL and OIDC_ISSUER_URL differ." >&2
  echo "      This is valid only when issuer/JWKS split is intentional." >&2
fi

if [[ ",${COMPOSE_PROFILES:-}," == *",local-idp,"* ]]; then
  keycloak_hostname="${KEYCLOAK_HOSTNAME:-}"
  if [ -z "$keycloak_hostname" ]; then
    echo "ERROR: KEYCLOAK_HOSTNAME is required when local-idp profile is enabled" >&2
    exit 1
  fi
  if [[ "$keycloak_hostname" != */devhub/auth/keycloak ]] && [[ "$keycloak_hostname" != */devhub/auth/keycloak/ ]]; then
    echo "ERROR: KEYCLOAK_HOSTNAME must include /devhub/auth/keycloak when local-idp is enabled" >&2
    echo "       current: $keycloak_hostname" >&2
    exit 1
  fi
fi

if [ -n "${DEVHUB_OIDC_JWKS_URL:-}" ]; then
  if ! [[ "$DEVHUB_OIDC_JWKS_URL" =~ /protocol/openid-connect/certs$ ]]; then
    echo "WARN: DEVHUB_OIDC_JWKS_URL does not end with /protocol/openid-connect/certs" >&2
  fi
fi

echo "[0/3] nginx conf sync (auto-fix)"
"$ROOT_DIR/scripts/nginx-conf-sync.sh" --fix

echo "[1/3] Docker availability check"
docker --version >/dev/null

docker compose version >/dev/null

echo "[2/3] Compose render validation"
(
  cd "$ROOT_DIR"
  docker compose "${compose_profile_args[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/tmp/devhub_deploy_compose_rendered.yml
)

echo "[3/3] OIDC endpoint reachability checks"
# issuer reachability — `docs/setup/deploy.prod.env.example` 가 권장하는 issuer/JWKS
# 분리 시나리오 (issuer = public, JWKS = internal) 에서 deploy 머신이 internal
# 환경 (public DNS 차단) 일 경우 issuer URL 도달 불가. 해당 case 에서는
# `SKIP_OIDC_ISSUER_REACH=1` 로 issuer reachability 검증을 건너뛰고 JWKS reachability
# 만 강제. (PR #278 review P2 #2 정합 — claude follow-up)
if [ "${SKIP_OIDC_ISSUER_REACH:-0}" = "1" ]; then
  echo "  SKIP_OIDC_ISSUER_REACH=1 — issuer reachability 검증 skip (issuer/JWKS split scenario)"
else
  curl -fsS "$OIDC_ISSUER_URL/.well-known/openid-configuration" >/dev/null
fi
if [ -n "${DEVHUB_OIDC_JWKS_URL:-}" ]; then
  curl -fsS "$DEVHUB_OIDC_JWKS_URL" >/dev/null
fi

echo "preflight OK"
