#!/usr/bin/env bash

set -euo pipefail

workflow_path="${1:-.github/workflows/ci.yml}"

if [ ! -f "$workflow_path" ]; then
  echo "workflow file not found: $workflow_path" >&2
  exit 1
fi

required_workflow_tokens=(
  "DEVHUB_OIDC_ISSUER_URL"
  "DEVHUB_KEYCLOAK_ADMIN_URL"
  "DEVHUB_KEYCLOAK_ADMIN_REALM"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET"
  "NEXT_PUBLIC_OIDC_ISSUER_URL"
)

required_e2e_tokens=(
  "DEVHUB_KEYCLOAK_ADMIN_URL"
  "DEVHUB_KEYCLOAK_ADMIN_REALM"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET"
)
#
# DEVHUB_BUILD_TIER 는 의도적으로 required_e2e_tokens 에 미포함.
# e2e shard 1/2/3 (saovae_stub default) 의 env block 에는 미설정.
# 본 PR 의 e2e-internal job (DEVHUB_BUILD_TIER=internal) 만 별도로 token 노출.
# e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 actionlint + 실제 e2e run 이 검증.

for token in "${required_workflow_tokens[@]}"; do
  if ! grep -Rqs -- "$token" "$workflow_path" frontend/tests/e2e/global-setup.ts frontend/tests/e2e/fixtures.ts; then
    echo "E2E-CI sync contract missing token: $token" >&2
    exit 1
  fi
done

for token in "${required_e2e_tokens[@]}"; do
  if ! grep -Rqs -- "$token" frontend/tests/e2e/global-setup.ts frontend/tests/e2e/fixtures.ts; then
    echo "E2E helper missing required env token: $token" >&2
    exit 1
  fi
done

# CI e2e job must not wire legacy Ory envs.
for forbidden in DEVHUB_HYDRA_ADMIN_URL DEVHUB_HYDRA_PUBLIC_URL DEVHUB_KRATOS_PUBLIC_URL DEVHUB_KRATOS_ADMIN_URL; do
  if grep -Rqs -- "$forbidden" "$workflow_path"; then
    echo "E2E-CI sync contract violation: legacy env still present: $forbidden" >&2
    exit 1
  fi
done

echo "E2E-CI sync contract check passed."
