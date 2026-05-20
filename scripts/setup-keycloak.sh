#!/bin/bash
# scripts/setup-keycloak.sh — Local Keycloak realm/client/role setup.
# Based on .github/workflows/ci.yml logic.

set -euo pipefail

BASE_URL="${KEYCLOAK_URL:-http://localhost:23000/devhub/auth/keycloak}"
ADMIN_USER="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
REALM="${DEVHUB_REALM:-devhub}"

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
  echo "  Creating client 'devhub-frontend'..."
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/clients" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d '{"clientId":"devhub-frontend","enabled":true,"publicClient":true,"standardFlowEnabled":true,"directAccessGrantsEnabled":false,"redirectUris":["*"],"webOrigins":["*"]}'
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
    for role_name in view-users query-users manage-users view-realm; do
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
  
  echo "  Assigning 'developer' role to 'test'..."
  dev_role_json=$(
    curl -fsS -H "Authorization: Bearer ${admin_token}" \
      "$BASE_URL/admin/realms/$REALM/roles/developer"
  )
  curl -fsS -X POST "$BASE_URL/admin/realms/$REALM/users/$test_user_id/role-mappings/realm" \
    -H "Authorization: Bearer ${admin_token}" \
    -H "Content-Type: application/json" \
    -d "[$dev_role_json]"
else
  echo "  User 'test' already exists."
fi

echo "Configuration complete."
echo "--------------------------------------------------"
echo "DEVHUB_OIDC_CLIENT_SECRET=$backend_secret"
echo "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=$backend_secret"
echo "--------------------------------------------------"
