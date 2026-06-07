#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env.dogfood}"

usage() {
  cat <<'EOF'
Usage:
  ./scripts/dogfood-create-user.sh <email> [password] [display-name]

Examples:
  ./scripts/dogfood-create-user.sh dogfood-user@example.com
  ./scripts/dogfood-create-user.sh dogfood-user@example.com 'ChangeMe-12345!' 'Dogfood User'
EOF
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

EMAIL="${1:-}"
PASSWORD="${2:-ChangeMe-12345!}"
DISPLAY_NAME="${3:-Dogfood User}"

if [ -z "$EMAIL" ]; then
  usage
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: env file not found: $ENV_FILE" >&2
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

KC_BASE_URL="${DEVHUB_KEYCLOAK_ADMIN_URL%/}"
KC_REALM="${DEVHUB_KEYCLOAK_ADMIN_REALM:-devhub}"
KC_CLIENT_ID="${DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID:-devhub-backend}"
KC_CLIENT_SECRET="${DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET:-}"

if [ -z "$KC_CLIENT_SECRET" ]; then
  echo "ERROR: DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET is required" >&2
  exit 1
fi

json_field() {
  local path="$1"
  node -e '
    const fs = require("fs");
    const path = process.argv[1].split(".");
    const input = fs.readFileSync(0, "utf8");
    const data = input ? JSON.parse(input) : {};
    let cur = data;
    for (const key of path) {
      cur = cur?.[key];
    }
    if (cur === undefined || cur === null) process.exit(1);
    if (typeof cur === "object") {
      process.stdout.write(JSON.stringify(cur));
    } else {
      process.stdout.write(String(cur));
    }
  ' "$path"
}

fetch_admin_token() {
  curl -fsS \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=client_credentials" \
    -d "client_id=${KC_CLIENT_ID}" \
    -d "client_secret=${KC_CLIENT_SECRET}" \
    "${KC_BASE_URL}/realms/${KC_REALM}/protocol/openid-connect/token" \
    | json_field access_token
}

find_user_id_by_email() {
  local token="$1"
  local response
  response="$(curl -fsS \
    -H "Authorization: Bearer ${token}" \
    -H "Accept: application/json" \
    "${KC_BASE_URL}/admin/realms/${KC_REALM}/users?email=${EMAIL}&exact=true")"

  printf '%s' "$response" | node -e '
    const fs = require("fs");
    const body = fs.readFileSync(0, "utf8");
    const data = body ? JSON.parse(body) : [];
    if (Array.isArray(data) && data[0]?.id) process.stdout.write(String(data[0].id));
  '
}

upsert_user() {
  local token="$1"
  local user_id="$2"
  local payload

  payload="$(node -e '
    const email = process.argv[1];
    const displayName = process.argv[2];
    process.stdout.write(JSON.stringify({
      username: email,
      email,
      enabled: true,
      emailVerified: true,
      firstName: displayName,
      lastName: "",
      attributes: {
        source: ["dogfood"],
      },
    }));
  ' "$EMAIL" "$DISPLAY_NAME")"

  if [ -n "$user_id" ]; then
    curl -fsS -X PUT \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      -d "$payload" \
      "${KC_BASE_URL}/admin/realms/${KC_REALM}/users/${user_id}" >/dev/null
    return
  fi

  curl -fsS -X POST \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "${KC_BASE_URL}/admin/realms/${KC_REALM}/users" >/dev/null
}

reset_password() {
  local token="$1"
  local user_id="$2"
  curl -fsS -X PUT \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    -d "{\"type\":\"password\",\"value\":\"${PASSWORD}\",\"temporary\":false}" \
    "${KC_BASE_URL}/admin/realms/${KC_REALM}/users/${user_id}/reset-password" >/dev/null
}

main() {
  local token
  local user_id

  token="$(fetch_admin_token)"
  user_id="$(find_user_id_by_email "$token" || true)"

  upsert_user "$token" "$user_id"

  if [ -z "$user_id" ]; then
    user_id="$(find_user_id_by_email "$token")"
  fi

  if [ -z "$user_id" ]; then
    echo "ERROR: failed to resolve Keycloak user id for $EMAIL" >&2
    exit 1
  fi

  reset_password "$token" "$user_id"

  cat <<EOF
Created/updated dogfood user
  email       : $EMAIL
  password    : $PASSWORD
  displayName : $DISPLAY_NAME
  keycloak id : $user_id
  login URL   : http://localhost:${FRONTEND_PORT:-13000}/login

Note:
  - This only creates the Keycloak account.
  - The first login should redirect to /onboarding.
  - After onboarding submit, the DevHub user row is created with review_status=pending_review.
EOF
}

main
