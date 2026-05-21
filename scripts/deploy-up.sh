#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.deploy.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/docs/setup/deploy.env.example}"

"$ROOT_DIR/scripts/deploy-preflight.sh"

compose_profile_args=()
if [ -n "${COMPOSE_PROFILES:-}" ]; then
  IFS=',' read -r -a compose_profiles <<< "$COMPOSE_PROFILES"
  for profile in "${compose_profiles[@]}"; do
    profile="${profile// /}"
    if [ -n "$profile" ]; then
      compose_profile_args+=(--profile "$profile")
    fi
  done
fi

echo "pulling images..."
(
  cd "$ROOT_DIR"
  docker compose "${compose_profile_args[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull
)

echo "starting stack..."
(
  cd "$ROOT_DIR"
  docker compose "${compose_profile_args[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d
)

echo "services status"
(
  cd "$ROOT_DIR"
  docker compose "${compose_profile_args[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps
)

echo "deploy up complete"
