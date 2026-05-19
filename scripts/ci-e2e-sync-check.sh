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
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET is required for e2e global setup"
)

required_e2e_tokens=(
  "DEVHUB_KEYCLOAK_ADMIN_URL"
  "DEVHUB_KEYCLOAK_ADMIN_REALM"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID"
  "DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET"
)

for token in "${required_workflow_tokens[@]}"; do
  if ! rg -q "$token" "$workflow_path" frontend/tests/e2e/global-setup.ts frontend/tests/e2e/fixtures.ts; then
    echo "E2E-CI sync contract missing token: $token" >&2
    exit 1
  fi
done

for token in "${required_e2e_tokens[@]}"; do
  if ! rg -q "$token" frontend/tests/e2e/global-setup.ts frontend/tests/e2e/fixtures.ts; then
    echo "E2E helper missing required env token: $token" >&2
    exit 1
  fi
done

# CI e2e job must not wire legacy Ory envs.
for forbidden in DEVHUB_HYDRA_ADMIN_URL DEVHUB_HYDRA_PUBLIC_URL DEVHUB_KRATOS_PUBLIC_URL DEVHUB_KRATOS_ADMIN_URL; do
  if rg -q "$forbidden" "$workflow_path"; then
    echo "E2E-CI sync contract violation: legacy env still present: $forbidden" >&2
    exit 1
  fi
done

echo "E2E-CI sync contract check passed."
