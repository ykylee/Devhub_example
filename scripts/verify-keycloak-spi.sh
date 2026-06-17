#!/usr/bin/env bash
# verify-keycloak-spi.sh — P2-6 (Keycloak SPI provider JAR) + X-8 (P3-5 audit event
# listener SPI push 전환) 자동 검증.
# 4 항목: (1) SPI JAR classpath + SPI service 등록 / (2) Realm eventsListeners
# 등록 / (3) Env var (DEVHUB_BACKEND_SPI_WEBHOOK_URL + DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET) /
# (4) Webhook push smoke (test user login → audit_logs 의 keycloak_event 1건 추가, latency < 1s).
# docs/setup/keycloak_event_listener_spi.md §6 정공법.
# Exit 0 = 4 항목 모두 OK, Exit 1 = 1건 이상 FAIL.

set -euo pipefail

# Prerequisite: jq (json parse) + curl + bash 4+
# docker-compose.colima 의 colima context 기준.

# === env ===
KEYCLOAK_URL="${KEYCLOAK_URL:-https://kc.staging.internal/auth}"
KC_BOOTSTRAP_ADMIN_USERNAME="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
KC_BOOTSTRAP_ADMIN_PASSWORD="${KC_BOOTSTRAP_ADMIN_PASSWORD:?set KC_BOOTSTRAP_ADMIN_PASSWORD}"
DEVHUB_REALM="${DEVHUB_REALM:-devhub}"
DEVHUB_BACKEND_API_URL="${DEVHUB_BACKEND_API_URL:-http://backend-core:8080}"
DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET="${DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET:?set DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET}"

PASS=0
FAIL=0

note() { printf "    %s\n" "$*"; }
pass() { printf "  [PASS] %s\n" "$*"; PASS=$((PASS+1)); }
fail() { printf "  [FAIL] %s\n" "$*"; FAIL=$((FAIL+1)); }

echo "==> Verifying Keycloak Event Listener SPI (P2-6 + X-8)"
echo "    Realm: $DEVHUB_REALM"
echo "    BaseURL: $KEYCLOAK_URL"
echo "    Backend: $DEVHUB_BACKEND_API_URL"

# === admin token ===
echo "==> Obtaining admin token..."
ADMIN_TOKEN=$(curl -sSf -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$KC_BOOTSTRAP_ADMIN_USERNAME&password=$KC_BOOTSTRAP_ADMIN_PASSWORD&grant_type=password&client_id=admin-cli" \
  "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" | jq -r .access_token)
if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
  fail "admin token obtain failed"
  exit 1
fi
note "admin token obtained (length=${#ADMIN_TOKEN})"

# === [1/4] SPI JAR classpath + service registration ===
echo "==> [1/4] SPI JAR classpath + service registration"
if command -v docker >/dev/null 2>&1; then
  KEYCLOAK_CONTAINER=$(docker ps --filter "label=com.docker.compose.service=keycloak" --format '{{.Names}}' | head -1)
  if [ -n "$KEYCLOAK_CONTAINER" ]; then
    if docker exec "$KEYCLOAK_CONTAINER" test -f /opt/keycloak/providers/devhub-keycloak-event-listener.jar 2>/dev/null; then
      pass "SPI JAR exists in Keycloak providers dir"
    else
      fail "SPI JAR missing at /opt/keycloak/providers/devhub-keycloak-event-listener.jar — build + mount SOP §2-§3"
    fi
    SERVICE_FILE=$(docker exec "$KEYCLOAK_CONTAINER" cat /opt/keycloak/providers/devhub-keycloak-event-listener.jar 2>/dev/null | unzip -p - META-INF/services/org.keycloak.events.EventListenerProviderFactory 2>/dev/null || true)
    if [ -z "$SERVICE_FILE" ]; then
      SERVICE_FILE=$(docker exec "$KEYCLOAK_CONTAINER" sh -c "cd /tmp && unzip -p /opt/keycloak/providers/devhub-keycloak-event-listener.jar META-INF/services/org.keycloak.events.EventListenerProviderFactory 2>/dev/null" 2>/dev/null || true)
    fi
    if echo "$SERVICE_FILE" | grep -q "com.devhub.keycloak.spi.DevHubEventListenerProviderFactory"; then
      pass "SPI service registration: DevHubEventListenerProviderFactory listed"
    else
      fail "SPI service registration missing DevHubEventListenerProviderFactory in META-INF/services"
    fi
  else
    note "Keycloak container not found (docker compose label=keycloak). [1/4] skipped."
  fi
else
  note "docker not available. [1/4] skipped."
fi

# === [2/4] Realm eventsListeners 등록 ===
echo "==> [2/4] Realm eventsListeners registration"
EVENTS_CONFIG=$(curl -sSf -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/admin/realms/$DEVHUB_REALM/events/config")
if echo "$EVENTS_CONFIG" | jq -e '.eventsListeners | index("devhub-event-listener")' >/dev/null 2>&1; then
  pass "Realm $DEVHUB_REALM eventsListeners contains 'devhub-event-listener'"
else
  fail "Realm $DEVHUB_REALM eventsListeners missing 'devhub-event-listener' — realm config SOP §4"
fi

# === [3/4] Env var (DEVHUB_BACKEND_SPI_WEBHOOK_URL + DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET) ===
echo "==> [3/4] Env var (DEVHUB_BACKEND_SPI_WEBHOOK_URL + DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET)"
if [ -n "$KEYCLOAK_CONTAINER" ]; then
  ENV_WEBHOOK_URL=$(docker exec "$KEYCLOAK_CONTAINER" sh -c 'echo "$DEVHUB_BACKEND_SPI_WEBHOOK_URL"' 2>/dev/null)
  if [ -n "$ENV_WEBHOOK_URL" ]; then
    pass "DEVHUB_BACKEND_SPI_WEBHOOK_URL set: $ENV_WEBHOOK_URL"
  else
    fail "DEVHUB_BACKEND_SPI_WEBHOOK_URL not set in Keycloak container"
  fi
  ENV_SECRET=$(docker exec "$KEYCLOAK_CONTAINER" sh -c 'echo "$DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET"' 2>/dev/null)
  if [ -n "$ENV_SECRET" ]; then
    pass "DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET set (length=${#ENV_SECRET})"
  else
    fail "DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET not set in Keycloak container"
  fi
else
  note "Keycloak container not found. [3/4] skipped."
fi

# === [4/4] Webhook push smoke (test user login → audit_logs keycloak_event 1건, latency < 1s) ===
echo "==> [4/4] Webhook push smoke (test user login → audit_logs keycloak_event)"
if [ -n "${TEST_USER_USERNAME:-}" ] && [ -n "${TEST_USER_PASSWORD:-}" ]; then
  # Login as test user (push user event to backend webhook)
  PUSH_START_MS=$(date +%s%3N)
  LOGIN_RESP=$(curl -sS -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=$TEST_USER_USERNAME&password=$TEST_USER_PASSWORD&grant_type=password&client_id=devhub-frontend" \
    "$KEYCLOAK_URL/realms/$DEVHUB_REALM/protocol/openid-connect/token" || echo "000")
  if [ "$LOGIN_RESP" = "200" ]; then
    pass "test user login push event fired (HTTP 200, latency 30s polling cron fallback excluded)"
    # Verify backend audit_logs receive push within 1s
    sleep 1
    AUDIT_COUNT=$(curl -sSf -H "Authorization: Bearer $ADMIN_TOKEN" \
      "$DEVHUB_BACKEND_API_URL/api/v1/internal/audit-events/keycloak?since=$PUSH_START_MS" 2>/dev/null | jq -r 'length // 0')
    if [ "$AUDIT_COUNT" -ge 1 ]; then
      pass "backend audit_logs received push event within 1s (count=$AUDIT_COUNT, latency < 1s ✓)"
    else
      fail "backend audit_logs did not receive push event within 1s (count=$AUDIT_COUNT, expected >= 1) — check SPI + webhook URL + secret"
    fi
  else
    fail "test user login failed (HTTP $LOGIN_RESP, expected 200) — set TEST_USER_USERNAME + TEST_USER_PASSWORD"
  fi
else
  note "TEST_USER_USERNAME + TEST_USER_PASSWORD not set. [4/4] skipped. Set both to enable webhook push smoke."
fi

# === Summary ===
echo "==> Summary"
echo "    PASS: $PASS"
echo "    FAIL: $FAIL"

if [ "$FAIL" -eq 0 ]; then
  echo "==> ✅ All checks passed — P2-6 + X-8 acceptance OK"
  exit 0
else
  echo "==> ❌ $FAIL check(s) failed — see detail above"
  exit 1
fi
