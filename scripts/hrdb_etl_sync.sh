#!/bin/bash
# scripts/hrdb_etl_sync.sh — off-boarding Phase 1 운영 cron (sprint claude/work_260519-p).
#
# ⚠ DEPRECATED (2026-05-20, sprint claude/work_260520-q-215-hrdb-cancel)
# ──────────────────────────────────────────────────────────────────────────────
# 사용자 결정 (issue #215, 2026-05-20): DevHub 가 외부 Keycloak 시나리오 채택 —
# 사내 IdP 팀이 별도 운영하는 Keycloak. HR ↔ Keycloak sync 는 IdP 팀 책임 (Keycloak
# User Federation 또는 사내 ETL → Keycloak Admin REST). DevHub 가 자체 cron 으로
# HRDB sync 하지 않음.
#
# 대체 흐름 (ADR-0020 결정 A + sub-carve C):
#   1. 외부 IdP 팀이 HR 'terminated' → Keycloak user disable (Keycloak admin console
#      또는 사내 ETL → Keycloak Admin REST PUT /users/{id} {enabled: false})
#   2. Keycloak event listener (sprint -k PR #241 의 backend cron) 가 admin event
#      polling — USER:UPDATE event 감지
#   3. DevHub backend 가 user_sync.go::SyncUserProfile 호출 → `users.status =
#      'deactivated'` 자동 sync (PR #241 의 정공법)
#
# 본 script 는 historical reference 로 보존 (별도 운영팀이 사내 환경에서 참고용으로
# 활용 가능). DevHub 운영 cron 에 등록하지 않음.
# ──────────────────────────────────────────────────────────────────────────────
#
# [ADR-0019 §5.3](../docs/adr/0019-keycloak-only-idp.md#53-잔여-carve-out) off-boarding 즉시성 carve out
# 의 Phase 1 권장 옵션 C 구현. design 은 docs/planning/keycloak_offboarding_immediacy.md §3.
#
# 동작:
#   1. HR 시스템에서 active=false 사용자 list export (사내 HR export 도구 호출 — env $HR_EXPORT_CMD)
#   2. DevHub PG users.status 직접 UPDATE — codex review #9 정정 정합 (hrdb.persons no active column,
#      회수는 users.status = 'deactivated' 직접 UPDATE)
#   3. Keycloak Admin API 호출 — user disable + force logout (admin REST base path /admin/realms/...)
#
# worst case latency: ETL 주기 (hourly cron 권장) + access_token TTL (5분 권장) ≈ 1h
#
# Phase 2 carve: LDAP/AD federation 도입 시 본 script 폐기 + Keycloak User Federation 자연 sync
# (keycloak_offboarding_immediacy.md §4).
#
# 사용법 (운영 cron):
#   # /etc/cron.d/devhub-hrdb-etl
#   0 * * * * devhub-svc /opt/devhub/scripts/hrdb_etl_sync.sh >> /var/log/devhub/hrdb-etl.log 2>&1
#
# 필수 env 변수:
#   - HR_EXPORT_CMD       : 사내 HR export 도구 호출 — system_id list 1개씩 stdout 출력
#                          (예: "/opt/sahub/bin/list-deactivated.sh" 또는
#                          "curl -s https://sahub.example.com/api/deactivated/today")
#   - DEVHUB_DB_URL       : PostgreSQL DSN (예: postgres://devhub:secret@db:5432/devhub?sslmode=require)
#   - KC_ISSUER_URL       : Keycloak issuer (예: https://devhub.example.com/devhub/auth/keycloak/realms/devhub)
#   - KC_ADMIN_URL        : Keycloak admin base (예: https://devhub.example.com/devhub/auth/keycloak)
#   - KC_ADMIN_REALM      : default devhub
#   - KC_ADMIN_CLIENT_ID  : default devhub-backend (KC-PR-C service account)
#   - KC_ADMIN_CLIENT_SECRET : 사내 vault 보관 secret
#
# 의존성: bash 4+, psql, curl, jq.
#
# 검증:
#   - dry-run (HR_EXPORT_CMD 가 빈 list 반환) 시 정상 종료
#   - 사내 staging 1회 실행 → audit_logs 의 'USER:UPDATE' (enabled=false) + 'USER:ACTION' (LOGOUT) 발생 확인
#     (keycloak_event_audit_integration.md §4.2 정합)

set -euo pipefail

# --- env validation ---
: "${HR_EXPORT_CMD:?Required env: HR_EXPORT_CMD (사내 HR export 도구)}"
: "${DEVHUB_DB_URL:?Required env: DEVHUB_DB_URL}"
: "${KC_ISSUER_URL:?Required env: KC_ISSUER_URL}"
: "${KC_ADMIN_URL:?Required env: KC_ADMIN_URL}"
: "${KC_ADMIN_CLIENT_SECRET:?Required env: KC_ADMIN_CLIENT_SECRET}"

KC_ADMIN_REALM="${KC_ADMIN_REALM:-devhub}"
KC_ADMIN_CLIENT_ID="${KC_ADMIN_CLIENT_ID:-devhub-backend}"

# --- helpers ---
log() {
  echo "[$(date -u +'%Y-%m-%dT%H:%M:%SZ')] $*"
}

# Keycloak admin access_token 획득 (client_credentials grant)
fetch_kc_admin_token() {
  local token_resp
  token_resp=$(curl -sS -X POST \
    "${KC_ISSUER_URL}/protocol/openid-connect/token" \
    -d "grant_type=client_credentials" \
    -d "client_id=${KC_ADMIN_CLIENT_ID}" \
    -d "client_secret=${KC_ADMIN_CLIENT_SECRET}")
  echo "$token_resp" | jq -r '.access_token // empty'
}

KC_TOKEN=$(fetch_kc_admin_token)
if [ -z "$KC_TOKEN" ]; then
  log "ERROR: Keycloak admin token 획득 실패. KC_ISSUER_URL / KC_ADMIN_CLIENT_* 확인."
  exit 1
fi

log "Keycloak admin token 획득 완료. ETL 시작."

# --- main loop ---
processed=0
failed=0

# HR_EXPORT_CMD 가 stdout 으로 system_id 1개씩 line 출력 가정
while IFS= read -r system_id; do
  [ -z "$system_id" ] && continue
  log "Processing: ${system_id}"

  # Step 1: DevHub users.status 직접 UPDATE (codex review #9 정정 정합)
  # ADR-0008 hrdb.persons 는 HR master 의 사본 — 비활성화 user 는 daily ETL upsert 의 자연 결과로 미동기.
  # 회수는 users 도메인 master 의 status 컬럼 UPDATE.
  if ! psql "${DEVHUB_DB_URL}" -v ON_ERROR_STOP=1 -c \
    "UPDATE users SET status = 'deactivated', updated_at = NOW() WHERE system_id = '${system_id//\'/\'\'}'" \
    >/dev/null 2>&1; then
    log "  ERROR: users.status UPDATE 실패 — ${system_id}"
    failed=$((failed + 1))
    continue
  fi

  # Step 2: Keycloak user lookup (username == system_id 가정)
  kc_user_id=$(curl -sS -H "Authorization: Bearer ${KC_TOKEN}" \
    "${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users?username=${system_id}&exact=true" \
    | jq -r '.[0].id // empty')

  if [ -z "$kc_user_id" ]; then
    log "  WARN: Keycloak user 없음 — ${system_id} (DevHub users.status 만 갱신 완료)"
    processed=$((processed + 1))
    continue
  fi

  # Step 3: Keycloak user disable
  http_code=$(curl -sS -o /dev/null -w "%{http_code}" \
    -X PUT -H "Authorization: Bearer ${KC_TOKEN}" -H "Content-Type: application/json" \
    "${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users/${kc_user_id}" \
    -d '{"enabled": false}')
  if [ "$http_code" != "204" ]; then
    log "  ERROR: Keycloak user disable 실패 — ${system_id} (HTTP ${http_code})"
    failed=$((failed + 1))
    continue
  fi

  # Step 4: Keycloak force logout (모든 active session 종료)
  http_code=$(curl -sS -o /dev/null -w "%{http_code}" \
    -X POST -H "Authorization: Bearer ${KC_TOKEN}" \
    "${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users/${kc_user_id}/logout")
  if [ "$http_code" != "204" ]; then
    log "  WARN: Keycloak force logout 실패 — ${system_id} (HTTP ${http_code}) — disable 은 적용됨, access_token TTL 만료까지 잔여"
  fi

  log "  ✓ ${system_id} off-boarded (Keycloak user_id=${kc_user_id})"
  processed=$((processed + 1))
done < <(eval "${HR_EXPORT_CMD}")

log "ETL 완료 — processed=${processed} failed=${failed}"

if [ "$failed" -gt 0 ]; then
  log "ERROR: ${failed} user(s) failed. 사내 운영팀 확인 + 다음 cron 자동 재시도."
  exit 1
fi
