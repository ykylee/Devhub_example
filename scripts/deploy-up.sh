#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.deploy.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/docs/setup/deploy.env.example}"

"$ROOT_DIR/scripts/deploy-preflight.sh"

echo "pulling images..."
(
  cd "$ROOT_DIR"
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull
)

echo "starting stack..."
(
  cd "$ROOT_DIR"
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d
)

echo "services status"
(
  cd "$ROOT_DIR"
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps
)

echo "deploy up complete"
