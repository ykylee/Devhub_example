#!/usr/bin/env bash
set -euo pipefail

# nginx conf sync — renders infra/nginx/devhub.deploy.conf.template into
# infra/nginx/devhub.deploy.conf as a derived artifact for non-docker reference
# (host nginx, manual inspection, etc.). The docker deploy compose mounts the
# .template directly and lets nginx official image's envsubst entrypoint
# substitute env vars at container startup, so the rendered .conf is NOT used
# by docker-compose.deploy.yml — it exists purely as a synced reference.
#
# Modes:
#   --fix   (default): regenerate devhub.deploy.conf when out of sync
#   --check          : exit 1 on drift, print diff
#
# Placeholder defaults (override via env when running standalone):
#   KEYCLOAK_UPSTREAM=keycloak:8080
#   KEYCLOAK_ADMIN_ALLOW_CIDR=127.0.0.1
#
# deploy-preflight.sh invokes this with --fix so deploy never proceeds with a
# stale rendered conf vs the template source.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE="$ROOT_DIR/infra/nginx/devhub.deploy.conf.template"
TARGET="$ROOT_DIR/infra/nginx/devhub.deploy.conf"

MODE="${1:---fix}"

if [ ! -f "$TEMPLATE" ]; then
  echo "ERROR: template not found: $TEMPLATE" >&2
  exit 1
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "ERROR: envsubst not found (install gettext-base / gettext)" >&2
  exit 1
fi

: "${KEYCLOAK_UPSTREAM:=keycloak:8080}"
: "${KEYCLOAK_ADMIN_ALLOW_CIDR:=127.0.0.1}"

if command -v sha256sum >/dev/null 2>&1; then
  template_sha="$(sha256sum "$TEMPLATE" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  template_sha="$(shasum -a 256 "$TEMPLATE" | awk '{print $1}')"
else
  template_sha="(sha256 unavailable)"
fi

rendered="$(mktemp)"
trap 'rm -f "$rendered"' EXIT

{
  cat <<EOF
# AUTO-GENERATED from infra/nginx/devhub.deploy.conf.template — DO NOT EDIT.
# Modify the .template and re-run scripts/nginx-conf-sync.sh (deploy-preflight
# auto-regenerates on drift).
#
# Source template SHA-256: $template_sha
# Placeholder defaults used:
#   KEYCLOAK_UPSTREAM=$KEYCLOAK_UPSTREAM
#   KEYCLOAK_ADMIN_ALLOW_CIDR=$KEYCLOAK_ADMIN_ALLOW_CIDR
#
EOF
  KEYCLOAK_UPSTREAM="$KEYCLOAK_UPSTREAM" \
  KEYCLOAK_ADMIN_ALLOW_CIDR="$KEYCLOAK_ADMIN_ALLOW_CIDR" \
    envsubst '${KEYCLOAK_UPSTREAM} ${KEYCLOAK_ADMIN_ALLOW_CIDR}' < "$TEMPLATE"
} > "$rendered"

case "$MODE" in
  --check)
    if [ ! -f "$TARGET" ]; then
      echo "ERROR: $TARGET missing — run with --fix to generate" >&2
      exit 1
    fi
    if ! diff -q "$TARGET" "$rendered" >/dev/null; then
      echo "ERROR: $(basename "$TARGET") out of sync with .template" >&2
      diff -u "$TARGET" "$rendered" >&2 || true
      exit 1
    fi
    echo "nginx conf sync OK ($(basename "$TARGET"))"
    ;;
  --fix)
    if [ -f "$TARGET" ] && diff -q "$TARGET" "$rendered" >/dev/null; then
      echo "nginx conf already in sync ($(basename "$TARGET"))"
    else
      cp "$rendered" "$TARGET"
      echo "regenerated $TARGET (template SHA-256: $template_sha)"
    fi
    ;;
  *)
    echo "ERROR: invalid mode: $MODE (use --check or --fix)" >&2
    exit 1
    ;;
esac
