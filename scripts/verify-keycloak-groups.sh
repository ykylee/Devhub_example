#!/bin/bash
# scripts/verify-keycloak-groups.sh — Keycloak group staging/prod setup 검증 (read-only).
# issue #214 P1-3 + ADR-0019 §5.3 + keycloak_groups_rbac_mapping.md §6.2 / §6.3 정합.
#
# 사용 시점:
#   - 사내 운영자가 staging / prod Keycloak admin console 에서 group 4종 + composite role 적용 후
#     자동 검증 1회 실행
#   - 검증 결과 모두 OK 시 acceptance 충족 — issue #214 close 근거
#
# 환경변수:
#   KEYCLOAK_URL                Keycloak base URL (필수, 예: https://kc.staging.internal/auth)
#   KC_BOOTSTRAP_ADMIN_USERNAME Keycloak admin user (기본 admin)
#   KC_BOOTSTRAP_ADMIN_PASSWORD Keycloak admin password (기본 admin — 운영은 vault 권장)
#   DEVHUB_REALM                DevHub realm name (기본 devhub)
#
# 검증 항목 (read-only, write 없음):
#   1. realm `$DEVHUB_REALM` 존재
#   2. group 4종 존재 (devhub-developers / devhub-managers / devhub-pmo-managers / devhub-system-admins)
#   3. 각 group 의 Composite Realm Role 정합 (group ↔ realm role 1:1 매핑)
#   4. Default Groups 비어 있음 (codex review #9 정정 — multi-role order-dependency 위험 회피)
#
# Prerequisites (host tools):
#   - python3 (3.8+, json 표준 라이브러리만 사용) — admin token / group / role JSON 파싱
#   - curl                                        — Keycloak Admin REST API
#   - bash 4+                                     — array / [[ syntax
#
# Exit code:
#   0 — 4 항목 모두 OK
#   1 — 1건 이상 FAIL (FAIL 항목 stderr 출력 + summary 표 stdout)
#
# Idempotent: 모든 호출이 GET. write 없음. N회 호출해도 동일 결과.

set -euo pipefail

BASE_URL="${KEYCLOAK_URL:-}"
ADMIN_USER="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
REALM="${DEVHUB_REALM:-devhub}"

if [ -z "$BASE_URL" ]; then
  echo "ERROR: KEYCLOAK_URL 미설정. 예: KEYCLOAK_URL=https://kc.staging.internal/auth ./scripts/verify-keycloak-groups.sh" >&2
  exit 1
fi

# group ↔ expected realm role 1:1 매핑 (keycloak_groups_rbac_mapping.md §3.1)
EXPECTED_GROUPS=(
  "devhub-developers:developer"
  "devhub-managers:manager"
  "devhub-pmo-managers:pmo_manager"
  "devhub-system-admins:system_admin"
)

# 검증 결과 누적
PASS_COUNT=0
FAIL_COUNT=0
FAIL_DETAIL=()

mark_pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf "  [PASS] %s\n" "$1"
}

mark_fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  FAIL_DETAIL+=("$1")
  printf "  [FAIL] %s\n" "$1" >&2
}

echo "==> Verifying Keycloak groups setup"
echo "    Realm: $REALM"
echo "    BaseURL: $BASE_URL"

echo "==> Obtaining admin token..."
admin_token=$(
  curl -fsS "$BASE_URL/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=$ADMIN_USER" \
    -d "password=$ADMIN_PASS" \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])'
)

# === 검증 1: realm 존재 ===
echo "==> [1/4] Realm existence"
realm_code=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer ${admin_token}" \
  "$BASE_URL/admin/realms/$REALM")
if [ "$realm_code" = "200" ]; then
  mark_pass "realm '$REALM' exists"
else
  mark_fail "realm '$REALM' missing (HTTP $realm_code)"
  echo ""
  echo "==> Summary (early abort, realm missing)"
  exit 1
fi

# === 검증 2 + 3: group 4종 존재 + composite role 정합 ===
echo "==> [2/4] Group existence (4 groups)"
echo "==> [3/4] Composite realm role mapping (1:1)"
# codex review (#306 P2) — `GET /groups` 가 top-level 만 반환. nested
# hierarchy 환경 (e.g., `/devhub/developers`) 에서 group_id 못 찾음 →
# search query (`?search=<name>&exact=true`) + recursive subGroups
# fallback. Keycloak Admin REST 의 search 는 모든 level 매칭.
for entry in "${EXPECTED_GROUPS[@]}"; do
  group_name="${entry%%:*}"
  expected_role="${entry##*:}"

  # search query 로 모든 level 검색 (exact match).
  search_result=$(curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/groups?search=$(printf '%s' "$group_name" | sed 's/ /%20/g')&exact=true")

  group_id=$(printf "%s" "$search_result" | python3 -c "
import json,sys

def find_group(groups, target):
    for g in groups:
        if g.get('name') == target:
            return g.get('id', '')
        # recursive subGroups (Keycloak nested hierarchy).
        nested = find_group(g.get('subGroups', []) or [], target)
        if nested:
            return nested
    return ''

print(find_group(json.load(sys.stdin), '$group_name'))
")

  if [ -z "$group_id" ]; then
    mark_fail "group '$group_name' missing"
    continue
  fi
  mark_pass "group '$group_name' exists (id $group_id)"

  # composite role 검증 — group 의 realm role mappings 가 expected_role 정확히 1개
  realm_roles_json=$(curl -fsS -H "Authorization: Bearer ${admin_token}" \
    "$BASE_URL/admin/realms/$REALM/groups/$group_id/role-mappings/realm")

  role_check=$(printf "%s" "$realm_roles_json" | python3 -c "
import json,sys
roles = json.load(sys.stdin)
names = sorted([r.get('name', '') for r in roles])
expected = '$expected_role'
if names == [expected]:
    print('OK')
elif expected in names:
    extras = [n for n in names if n != expected]
    print(f'EXTRA:{\",\".join(extras)}')
else:
    print(f'MISSING:got=[{\",\".join(names) if names else \"empty\"}]')
")

  case "$role_check" in
    OK)
      mark_pass "group '$group_name' → composite role '$expected_role' (1:1)"
      ;;
    EXTRA:*)
      extras="${role_check#EXTRA:}"
      mark_fail "group '$group_name' has expected role + extras: $extras (1:1 매핑 위반)"
      ;;
    MISSING:*)
      got="${role_check#MISSING:}"
      mark_fail "group '$group_name' missing role '$expected_role' ($got)"
      ;;
  esac
done

# === 검증 4: Default Groups 비어 있음 ===
echo "==> [4/4] Default Groups empty"
default_groups_json=$(curl -fsS -H "Authorization: Bearer ${admin_token}" \
  "$BASE_URL/admin/realms/$REALM/default-groups")
default_count=$(printf "%s" "$default_groups_json" | python3 -c '
import json,sys
print(len(json.load(sys.stdin)))
')

if [ "$default_count" -eq 0 ]; then
  mark_pass "Default Groups empty (codex review #9 정합)"
else
  default_names=$(printf "%s" "$default_groups_json" | python3 -c '
import json,sys
print(",".join(g.get("name","") for g in json.load(sys.stdin)))
')
  mark_fail "Default Groups not empty: $default_names (multi-role order-dependency 위험)"
fi

# === Summary ===
echo ""
echo "==> Summary"
echo "    PASS: $PASS_COUNT"
echo "    FAIL: $FAIL_COUNT"

if [ "$FAIL_COUNT" -gt 0 ]; then
  echo ""
  echo "==> FAIL detail:"
  for detail in "${FAIL_DETAIL[@]}"; do
    echo "    - $detail"
  done
  exit 1
fi

echo ""
echo "==> ✅ All checks passed — Keycloak groups setup acceptance OK"
echo "    Next: token decode 검증 (test user login → realm_access.roles 정합)"
echo "    Reference: docs/planning/keycloak_groups_rbac_mapping.md §3.3"
exit 0
