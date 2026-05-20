#!/bin/bash
# scripts/setup-keycloak.sh — Keycloak realm/client/role setup (로컬 + 외부 모드 공용).
# Based on .github/workflows/ci.yml logic.
#
# 환경변수:
#   KEYCLOAK_URL                Keycloak base URL (필수, fallback 은 dev-up.sh 정합 dev default).
#                                 - 로컬 모드 (dev-up.sh) : http://localhost:8180/devhub/auth/keycloak
#                                 - 외부 모드 (사내 Keycloak): https://kc.internal.example.com/auth (또는 사내 path)
#   KC_BOOTSTRAP_ADMIN_USERNAME Keycloak admin user (기본 admin)
#   KC_BOOTSTRAP_ADMIN_PASSWORD Keycloak admin password (기본 admin — 운영은 vault 권장)
#   DEVHUB_REALM                DevHub realm name (기본 devhub)
#   DEVHUB_FRONTEND_ORIGIN      devhub-frontend client 의 redirect_uri / webOrigin origin (필수).
#                                 - 로컬 모드: http://localhost:3000
#                                 - 외부 모드: https://devhub.example.com
#                                 - 단일 포트 reverse-proxy 정합: https://devhub.example.com (basePath /devhub)
#   DEVHUB_FRONTEND_BASEPATH    Next.js basePath (기본 /devhub). 빈 값이면 native dev 모드.
#
# 단일 포트 컨셉 가드 (ADR-0018, ADR-0019):
#   - redirectUris / webOrigins 에 wildcard "*" 또는 임의 host 허용을 자동 적용하지 않는다.
#   - DEVHUB_FRONTEND_ORIGIN 이 비어있으면 fail-fast.

set -euo pipefail

BASE_URL="${KEYCLOAK_URL:-http://localhost:8180/devhub/auth/keycloak}"
ADMIN_USER="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
REALM="${DEVHUB_REALM:-devhub}"
FRONTEND_ORIGIN="${DEVHUB_FRONTEND_ORIGIN:-}"
FRONTEND_BASEPATH="${DEVHUB_FRONTEND_BASEPATH:-/devhub}"

if [ -z "$FRONTEND_ORIGIN" ]; then
  echo "ERROR: DEVHUB_FRONTEND_ORIGIN 미설정. 단일 포트 컨셉 (ADR-0018) 정합을 위해 redirect_uri origin 을 명시해야 한다." >&2
  echo "예: DEVHUB_FRONTEND_ORIGIN=http://localhost:3000 (native dev)" >&2
  echo "    DEVHUB_FRONTEND_ORIGIN=https://devhub.example.com (단일 포트 reverse proxy)" >&2
  exit 1
fi

# Strip trailing slash from FRONTEND_ORIGIN; ensure FRONTEND_BASEPATH starts with /.
FRONTEND_ORIGIN="${FRONTEND_ORIGIN%/}"
case "$FRONTEND_BASEPATH" in
  /*) ;;
  "") ;;
  *) FRONTEND_BASEPATH="/$FRONTEND_BASEPATH" ;;
esac

# redirect_uris allowlist — 단일 origin 만 허용.
if [ -n "$FRONTEND_BASEPATH" ]; then
  REDIRECT_URIS="[\"${FRONTEND_ORIGIN}${FRONTEND_BASEPATH}/auth/callback\",\"${FRONTEND_ORIGIN}${FRONTEND_BASEPATH}/*\"]"
  POST_LOGOUT_URIS="${FRONTEND_ORIGIN}${FRONTEND_BASEPATH}/##${FRONTEND_ORIGIN}${FRONTEND_BASEPATH}/auth/login"
else
  REDIRECT_URIS="[\"${FRONTEND_ORIGIN}/auth/callback\",\"${FRONTEND_ORIGIN}/*\"]"
  POST_LOGOUT_URIS="${FRONTEND_ORIGIN}/##${FRONTEND_ORIGIN}/auth/login"
fi
WEB_ORIGINS="[\"${FRONTEND_ORIGIN}\"]"

echo "Waiting for Keycloak at $BASE_URL..."
timeout 120s bash -c "until curl -fsS \"$BASE_URL/realms/master/.well-known/openid-configuration\" >/dev/null; do echo -n '.'; sleep 2; done"
echo " Keycloak is up."

echo "Obtaining admin token..."
admin_token=$(
  curl -fsS "$BASE_URL/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=$ADMIN_USER" \
    -d "password=$ADMIN_PASS" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])'
)

echo "Checking if realm '$REALM' exists..."
http_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer ${admin_token}" \
  "$BASE_URL/admin/realms/$REALM")

if [ "$http_code" = "404" ]; then
  echo "Creating realm '$REALM'..."
  curl -fsS -X POST "$BASE_URL/admin/realms" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d "{\"realm\":\"$REALM\",\"enabled\":true,\"sslRequired\":\"none\"}"
else
  echo "Realm '$REALM' already exists. Updating sslRequired to none..."
  curl -fsS -X PUT "$BASE_URL/admin/realms/$REALM" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d "{\"sslRequired\":\"none\"}"
fi

echo "Creating roles..."
for role in developer manager pmo_manager system_admin; do
  code=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/roles/$role")
  if [ "$code" = "404" ]; then
    echo "  Creating role '$role'..."
    curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/roles" \
      -H "Authorization: Bearer ${admin_token}" \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"$role\"}"
  fi
done

echo "Configuring client 'devhub-frontend'..."
frontend_client_id=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients?clientId=devhub-frontend" \
    | python3 -c 'import json,sys; a=json.load(sys.stdin); print(a[0]["id"] if a else "")'
)
if [ -z "$frontend_client_id" ]; then
  echo "  Creating client 'devhub-frontend' (redirect_uris=${REDIRECT_URIS}, webOrigins=${WEB_ORIGINS})..."
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/clients" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d "{\"clientId\":\"devhub-frontend\",\"enabled\":true,\"publicClient\":true,\"standardFlowEnabled\":true,\"directAccessGrantsEnabled\":false,\"redirectUris\":${REDIRECT_URIS},\"webOrigins\":${WEB_ORIGINS},\"attributes\":{\"pkce.code.challenge.method\":\"S256\",\"post.logout.redirect.uris\":\"${POST_LOGOUT_URIS}\"}}"
  frontend_client_id=$(
    curl -fsS -H "Authorization: Bearer ${admin_token}" \
      "$BASE_URL/admin/realms/$REALM/clients?clientId=devhub-frontend" \
      | python3 -c 'import json,sys; a=json.load(sys.stdin); print(a[0]["id"] if a else "")'
  )
fi

echo "Adding audience mapper to 'devhub-frontend'..."
frontend_aud_mapper_id=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients/${frontend_client_id}/protocol-mappers/models" \
    | python3 -c 'import json,sys; m=next((x for x in json.load(sys.stdin) if x.get("name")=="devhub-frontend-audience"), None); print(m.get("id","") if m else "")'
)
if [ -z "$frontend_aud_mapper_id" ]; then
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/clients/${frontend_client_id}/protocol-mappers/models" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d '{"name":"devhub-frontend-audience","protocol":"openid-connect","protocolMapper":"oidc-audience-mapper","config":{"included.client.audience":"devhub-frontend","id.token.claim":"false","access.token.claim":"true"}}'
fi

echo "Configuring client 'devhub-backend'..."
backend_client_id=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients?clientId=devhub-backend" \
    | python3 -c 'import json,sys; a=json.load(sys.stdin); print(a[0]["id"] if a else "")'
)
if [ -z "$backend_client_id" ]; then
  echo "  Creating client 'devhub-backend'..."
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/clients" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d '{"clientId":"devhub-backend","enabled":true,"publicClient":false,"serviceAccountsEnabled":true,"standardFlowEnabled":false,"directAccessGrantsEnabled":false}'
  backend_client_id=$(
    curl -fsS -H "Authorization: Bearer ${admin_token}" \
      "$BASE_URL/admin/realms/$REALM/clients?clientId=devhub-backend" \
      | python3 -c 'import json,sys; a=json.load(sys.stdin); print(a[0]["id"] if a else "")'
  )
fi

echo "Granting Admin API permissions to 'devhub-backend' service account..."
realm_mgmt_client_id=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients?clientId=realm-management" \
    | python3 -c 'import json,sys; a=json.load(sys.stdin); print(a[0]["id"] if a else "")'
)
service_account_user_id=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients/${backend_client_id}/service-account-user" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])'
)
role_payload=$(
  {
    printf '['
    first=1
    # ADR-0020 sub-carve E (PR #244 merge 6810384) 정합 — service account
    # 는 view-users + view-events + view-realm 만 요구. manage-users 는 정
    # 공법 제거 (backend KeycloakAdminClient.CreateIdentity 등 write API 호
    # 출처 5건 모두 제거됨). docs/planning/keycloak_service_account_min_role.md.
    for role_name in view-users query-users view-events view-realm; do
      role_json=$(
        curl -fsS -H "Authorization: Bearer ${admin_token}" \
          "$BASE_URL/admin/realms/$REALM/clients/${realm_mgmt_client_id}/roles/${role_name}"
      )
      if [ "$first" -eq 1 ]; then
        printf '%s' "$role_json"
        first=0
      else
        printf ',%s' "$role_json"
      fi
    done
    printf ']'
  }
)
curl -fsS -X POST \
  "$BASE_URL/admin/realms/$REALM/users/${service_account_user_id}/role-mappings/clients/${realm_mgmt_client_id}" \
  -H "Authorization: Bearer ${admin_token}" \
  -H "Content-Type: application/json" \
  -d "$role_payload" >/dev/null

backend_secret=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/clients/${backend_client_id}/client-secret" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["value"])'
)

echo "Creating default user 'test'..."
user_exists=$(
  curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/users?username=test" \
    | python3 -c 'import json,sys; a=json.load(sys.stdin); print("true" if a else "false")'
)

if [ "$user_exists" = "false" ]; then
  echo "  User 'test' does not exist. Creating..."
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/users" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d '{"username":"test","enabled":true,"email":"test@example.com","firstName":"Test","lastName":"User","attributes":{"employee_id":"EMP001"}}'
  
  test_user_id=$(
    curl -fsS -H "Authorization: Bearer ${admin_token}" \
      "$BASE_URL/admin/realms/$REALM/users?username=test" \
      | python3 -c 'import json,sys; print(json.load(sys.stdin)[0]["id"])'
  )
  
  echo "  Setting password for 'test'..."
  curl -fsS -X PUT "$BASE_URL/admin/realms/$REALM/users/$test_user_id/reset-password" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d '{"type":"password","value":"test","temporary":false}'
  
  # infra/idp/sql/003_seed_test_admin.sql 정합 — DevHub users.role 이 system_admin
  # 으로 시드되므로 Keycloak 측도 일치시켜야 첫 로그인 시 lazy_auto_create
  # (sprint -i PR #239) 의 role merge 정합.
  echo "  Assigning 'system_admin' role to 'test'..."
  admin_role_json=$(
    curl -fsS -H "Authorization: Bearer ${admin_token}" \
      "$BASE_URL/admin/realms/$REALM/roles/system_admin"
  )
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/users/$test_user_id/role-mappings/realm" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d "[$admin_role_json]"
else
  echo "  User 'test' already exists."
fi

echo "Configuration complete."
echo "--------------------------------------------------"
echo "DEVHUB_OIDC_CLIENT_SECRET=$backend_secret"
echo "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=$backend_secret"
echo "--------------------------------------------------"
