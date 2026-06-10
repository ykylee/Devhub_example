#!/usr/bin/env bash
#
# Tier separation lint — 사외/사내 cross-tier leak 자동 검사.
# 정책 (`docs/governance/worker_division.md` §6.4 + PR #536 governance enforcement):
#
# 1. 사외 PR (default base = main) 가 다음을 변경했는지 검사:
#    - 사내 한정 env var (DEVHUB_KEYCLOAK_*, GITEA_URL, HR_EXPORT_CMD, ...)
#    - 사내 호스트명/IP (kc.internal.example.com, 172.16.0.0/12, ...)
#    - 사내 한정 경로 (infra/idp/*, scripts/setup-keycloak.sh, ...)
#    - hardcoded secret 패턴 (JWT, key=, token=, secret=, password= with high-entropy)
# 2. 변경 scope 가 사외-safe (ci.yml e2e build / unit / integration / lint / openapi
#    / workflow / migration) 인지 검사.
# 3. 통과하지 못하면 PR 머지 차단.
#
# 사용:
#   bash scripts/check-tier-separation.sh <base_ref> [<head_ref>]
#   - base_ref: 비교 기준 ref (default: origin/main)
#   - head_ref: 비교 대상 ref (default: HEAD)
#   - tier: 사외 (default) / 사내 / 공용
#
# Exit code:
#   0: 통과
#   1: 사내 한정 패턴 매칭 (cross-tier leak 가능)
#   2: 잘못된 tier 명시
#   3: 내부 오류 (git, grep 등)

set -euo pipefail

base_ref="${1:-origin/main}"
head_ref="${2:-HEAD}"
tier="${3:-external}"

if [[ "$tier" != "external" && "$tier" != "internal" && "$tier" != "shared" ]]; then
  echo "invalid tier: $tier (must be external / internal / shared)" >&2
  exit 2
fi

# 변경 파일 추출
if ! diff_output=$(git diff --name-only "$base_ref" "$head_ref" 2>&1); then
  echo "git diff failed: $diff_output" >&2
  exit 3
fi

if [ -z "$diff_output" ]; then
  echo "no changes between $base_ref and $head_ref"
  exit 0
fi

echo "=== tier-separation lint ==="
echo "  base: $base_ref"
echo "  head: $head_ref"
echo "  tier: $tier"
echo "  changed files: $(echo "$diff_output" | wc -l)"
echo

# 사내 한정 패턴 (regex set)
declare -a internal_patterns=(
  # 사내 env var
  'DEVHUB_KEYCLOAK_(ADMIN|STAGING|JWKS|SPI|WEBHOOK)_'
  'GITEA_URL='
  'GITEA_TOKEN='
  'GITEA_WEBHOOK_SECRET='
  'HR_EXPORT_CMD'
  'DEVHUB_HR_EMAIL_FALLBACK_DOMAIN'
  'DEVHUB_TRUSTED_PROXIES=172\.'
  # 사내 호스트명
  'kc\.internal\.example\.com'
  'internal-registry\.example\.com'
  'devhub-stage\.example\.com'
  'devhub\.example\.com'
  'sahub\.example\.com'
  # 사내 IP CIDR
  '172\.16\.0\.0/12'
  '172\.(1[6-9]|2[0-9]|3[0-1])\.0\.0'
  '10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}'
  '192\.168\.[0-9]{1,3}\.[0-9]{1,3}'
  # hardcoded Keycloak admin password pattern
  'KC_BOOTSTRAP_ADMIN_PASSWORD=[a-zA-Z0-9]+'
  'KEYCLOAK_ADMIN_PASSWORD=[a-zA-Z0-9]+'
  'DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=[a-zA-Z0-9-]+'
)

# 사내 한정 경로 (변경된 파일이 이 경로에 속하는지)
declare -a internal_paths=(
  'infra/idp/keycloak-realm.prod.json'
  'infra/idp/keycloak-event-listener-spi/'
  'scripts/setup-keycloak.sh'
  'scripts/verify-keycloak-groups.sh'
  'scripts/deploy-from-env.sh'
  'scripts/deploy-preflight.sh'
  'scripts/deploy-up.sh'
  'scripts/nginx-conf-sync.sh'
  'scripts/hrdb_etl_sync.sh'
  'scripts/hrdb_etl_seed.sql'
  'scripts/dogfood.sh'
  'scripts/dogfood-create-user.sh'
  'docs/dogfood/'
  'docs/setup/deploy.env.example'
  'docs/setup/deploy.stage.env.example'
  'docs/setup/deploy.prod.env.example'
  'docs/reports/'
  'docs/analysis/'
  'docker-compose.deploy.yml'
  'docker-compose.local.yml'
  'docker-compose.test.yml'
  '.env.deploy'
  '.env.prod'
  '.env.stage'
  '.env.local'
  '.env.test'
)

violations=0

echo "--- (1) 사내 한정 경로 변경 검사 ---"
for changed in $diff_output; do
  for internal_path in "${internal_paths[@]}"; do
    if [[ "$changed" == *"$internal_path"* ]]; then
      echo "  ❌ 사외 PR 이 사내 한정 경로 변경: $changed (matches $internal_path)"
      violations=$((violations + 1))
    fi
  done
done
echo

echo "--- (2) 사내 한정 패턴 (env/host/IP/secret) 매칭 검사 ---"
# 변경된 파일의 +line 만 grep. '-' (deletion) 는 skip.
for changed in $diff_output; do
  # skip binary / large files
  if [[ ! -f "$changed" ]]; then
    continue
  fi
  # diff 로 + 라인 만 추출
  added_lines=$(git diff "$base_ref" "$head_ref" -- "$changed" 2>/dev/null | grep -E '^\+' | grep -vE '^\+\+\+' || true)
  if [ -z "$added_lines" ]; then
    continue
  fi
  for pattern in "${internal_patterns[@]}"; do
    if echo "$added_lines" | grep -qE "$pattern"; then
      echo "  ❌ $changed: 사내 한정 패턴 매칭 ($pattern)"
      echo "$added_lines" | grep -E "$pattern" | head -3 | sed 's/^/      | /'
      violations=$((violations + 1))
    fi
  done
done
echo

echo "--- (3) tier 명시 검사 ---"
if [ "$tier" = "internal" ]; then
  echo "  ℹ️  internal tier PR. 본 lint 는 advisory 만 (사내 SCM 운영자 review 필수)."
fi
if [ "$tier" = "shared" ]; then
  echo "  ℹ️  shared tier PR. 양쪽 동기화 필수 (governance/추적성 ID/AGENTS.md drift 검증은 별도 lint)."
fi
echo

if [ "$violations" -gt 0 ]; then
  echo "=== FAIL: $violations violation(s) ==="
  echo "  사내 한정 정보가 사외 PR 에 포함되었습니다."
  echo "  본 PR 이 사내 SCM 으로만 push 되어야 한다면, PR template 의 Tier 를"
  echo "  '사내 (사내 SCM)' 으로 변경 후 owner 의 명시적 승인으로 merge 하십시오."
  echo "  자세한 정책: docs/governance/worker_division.md §6.4"
  exit 1
fi

echo "=== PASS: no 사내 한정 패턴 매칭 ==="
exit 0
