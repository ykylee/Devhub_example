#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# =========================
# Inline Config (edit here)
# =========================
# If the same variable is exported in shell, exported value wins.
: "${IMAGE_TAG:=change-me-tag}"
: "${IMAGE_REPO_PREFIX:=local/devhub}"
: "${DEVHUB_PUBLIC_BASE_URL:=http://100.90.113.29:23000}"
: "${DB_URL:=}"
: "${DEVHUB_OIDC_CLIENT_SECRET:=change-me-oidc-secret}"
: "${DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET:=change-me-keycloak-admin-secret}"
: "${DB_MODE:=external}" # external|docker
: "${POSTGRES_USER:=user}"
: "${POSTGRES_PASSWORD:=pass}"
: "${POSTGRES_DB:=devhub}"
: "${DB_HOST:=db}"
: "${DB_PORT:=5432}"
: "${DB_SSLMODE:=disable}"

# Optional (override only when needed)
: "${ACTION:=all}" # build|push|deploy|all
: "${ENV_FILE:=/tmp/devhub-deploy.env}"
: "${DOCKER_COMPOSE_FILE:=$ROOT_DIR/docker-compose.deploy.yml}"
: "${NEXT_PUBLIC_BASE_PATH:=devhub}"
: "${KEYCLOAK_UPSTREAM:=keycloak:8080}"
: "${KEYCLOAK_ADMIN_ALLOW_CIDR:=127.0.0.1}"
: "${NGINX_HTTP_PORT:=80}"
: "${NGINX_HTTPS_PORT:=443}"
: "${NGINX_TLS_CERT_PATH:=./infra/nginx/certs/tls.crt}"
: "${NGINX_TLS_KEY_PATH:=./infra/nginx/certs/tls.key}"

# Simple one-shot deploy helper.
# Required env:
#   IMAGE_TAG
#   IMAGE_REPO_PREFIX               (default: local/devhub)
#   DEVHUB_PUBLIC_BASE_URL          (e.g. https://devhub.example.com or http://100.90.113.29:23000)
#   DB_URL
#   DEVHUB_OIDC_CLIENT_SECRET
#   DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET
#
# Optional env:
#   ACTION                          build|push|deploy|all (default: all)
#   ENV_FILE                        deploy env file path (default: /tmp/devhub-deploy.env)
#   DOCKER_COMPOSE_FILE             compose file path (default: docker-compose.deploy.yml)
#   DEVHUB_OIDC_ISSUER_URL          default: ${DEVHUB_PUBLIC_BASE_URL}/devhub/auth/keycloak/realms/devhub
#   DEVHUB_KEYCLOAK_ADMIN_URL       default: ${DEVHUB_PUBLIC_BASE_URL}/devhub/auth/keycloak
#   OIDC_ISSUER_URL                 default: DEVHUB_OIDC_ISSUER_URL
#   OIDC_REDIRECT_URI               default: ${DEVHUB_PUBLIC_BASE_URL}/devhub/auth/callback
#   NEXT_PUBLIC_OIDC_ISSUER_URL     default: OIDC_ISSUER_URL
#   NEXT_PUBLIC_OIDC_REDIRECT_URI   default: OIDC_REDIRECT_URI
#   NEXT_PUBLIC_BASE_PATH           default: devhub
#   NEXT_PUBLIC_OIDC_CLIENT_ID      default: devhub-frontend
#   DEVHUB_OIDC_CLIENT_ID           default: devhub-frontend
#   DEVHUB_OIDC_AUDIENCE            default: devhub-frontend
#   DEVHUB_KEYCLOAK_ADMIN_REALM     default: devhub
#   DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID default: devhub-backend
#   DEVHUB_AUTH_DEV_FALLBACK        default: 0
#   BACKEND_API_URL                 default: http://backend-core:8080
#   BACKEND_AI_URL                  default: http://backend-ai:8000
#   DEVHUB_TRUSTED_PROXIES          default: 172.16.0.0/12
#   KEYCLOAK_UPSTREAM               default: keycloak:8080
#   KEYCLOAK_ADMIN_ALLOW_CIDR       default: 127.0.0.1
#   DB_MODE                         external|docker (docker will bring up db/db-init via compose profile)
#   POSTGRES_USER/PASSWORD/DB       default: user/pass/devhub
#   DB_HOST/DB_PORT/DB_SSLMODE      default: db/5432/disable
#   NGINX_HTTP_PORT                 default: 80
#   NGINX_HTTPS_PORT                default: 443
#   NGINX_TLS_CERT_PATH             default: ./infra/nginx/certs/tls.crt
#   NGINX_TLS_KEY_PATH              default: ./infra/nginx/certs/tls.key

require() {
  local var_name="$1"
  if [ -z "${!var_name:-}" ]; then
    echo "ERROR: required env var is empty: $var_name" >&2
    exit 1
  fi
}

emit_env_line() {
  local key="$1"
  local value="$2"
  printf "%s=%q\n" "$key" "$value"
}

build_env_file() {
  local public_base="${DEVHUB_PUBLIC_BASE_URL%/}"
  local base_path="${NEXT_PUBLIC_BASE_PATH:-devhub}"
  local base_path_norm="/${base_path#/}"
  base_path_norm="${base_path_norm%/}"

  if [ "$DB_MODE" = "docker" ]; then
    COMPOSE_PROFILES="local-db"
    DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${DB_HOST}:${DB_PORT}/${POSTGRES_DB}?sslmode=${DB_SSLMODE}"
  elif [ "$DB_MODE" = "external" ]; then
    : "${DB_URL:?set DB_URL when DB_MODE=external}"
    COMPOSE_PROFILES="${COMPOSE_PROFILES:-}"
  else
    echo "ERROR: invalid DB_MODE=$DB_MODE (use: external|docker)" >&2
    exit 1
  fi

  DEVHUB_OIDC_ISSUER_URL="${DEVHUB_OIDC_ISSUER_URL:-$public_base/devhub/auth/keycloak/realms/devhub}"
  DEVHUB_KEYCLOAK_ADMIN_URL="${DEVHUB_KEYCLOAK_ADMIN_URL:-$public_base/devhub/auth/keycloak}"
  OIDC_ISSUER_URL="${OIDC_ISSUER_URL:-$DEVHUB_OIDC_ISSUER_URL}"
  OIDC_REDIRECT_URI="${OIDC_REDIRECT_URI:-$public_base$base_path_norm/auth/callback}"
  NEXT_PUBLIC_OIDC_ISSUER_URL="${NEXT_PUBLIC_OIDC_ISSUER_URL:-$OIDC_ISSUER_URL}"
  NEXT_PUBLIC_OIDC_REDIRECT_URI="${NEXT_PUBLIC_OIDC_REDIRECT_URI:-$OIDC_REDIRECT_URI}"

  {
    emit_env_line IMAGE_TAG "$IMAGE_TAG"
    emit_env_line IMAGE_REPO_PREFIX "$IMAGE_REPO_PREFIX"
    emit_env_line DB_MODE "$DB_MODE"
    if [ -n "${COMPOSE_PROFILES:-}" ]; then
      emit_env_line COMPOSE_PROFILES "$COMPOSE_PROFILES"
    fi
    printf "\n"
    emit_env_line NGINX_HTTP_PORT "${NGINX_HTTP_PORT:-80}"
    emit_env_line NGINX_HTTPS_PORT "${NGINX_HTTPS_PORT:-443}"
    emit_env_line NGINX_TLS_CERT_PATH "${NGINX_TLS_CERT_PATH:-./infra/nginx/certs/tls.crt}"
    emit_env_line NGINX_TLS_KEY_PATH "${NGINX_TLS_KEY_PATH:-./infra/nginx/certs/tls.key}"
    emit_env_line DEVHUB_PUBLIC_BASE_URL "$DEVHUB_PUBLIC_BASE_URL"
    printf "\n"
    emit_env_line KEYCLOAK_UPSTREAM "${KEYCLOAK_UPSTREAM:-keycloak:8080}"
    emit_env_line KEYCLOAK_ADMIN_ALLOW_CIDR "${KEYCLOAK_ADMIN_ALLOW_CIDR:-127.0.0.1}"
    emit_env_line KEYCLOAK_HOSTNAME "${KEYCLOAK_HOSTNAME:-localhost}"
    emit_env_line KC_BOOTSTRAP_ADMIN_USERNAME "${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
    emit_env_line KC_BOOTSTRAP_ADMIN_PASSWORD "${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
    emit_env_line KC_DB_URL "${KC_DB_URL:-jdbc:postgresql://db:5432/devhub}"
    emit_env_line KC_DB_USERNAME "${KC_DB_USERNAME:-user}"
    emit_env_line KC_DB_PASSWORD "${KC_DB_PASSWORD:-pass}"
    emit_env_line KC_DB_SCHEMA "${KC_DB_SCHEMA:-keycloak}"
    printf "\n"
    emit_env_line DB_URL "$DB_URL"
    emit_env_line POSTGRES_USER "$POSTGRES_USER"
    emit_env_line POSTGRES_PASSWORD "$POSTGRES_PASSWORD"
    emit_env_line POSTGRES_DB "$POSTGRES_DB"
    emit_env_line BACKEND_API_URL "${BACKEND_API_URL:-http://backend-core:8080}"
    emit_env_line BACKEND_AI_URL "${BACKEND_AI_URL:-http://backend-ai:8000}"
    emit_env_line DEVHUB_AUTH_DEV_FALLBACK "${DEVHUB_AUTH_DEV_FALLBACK:-0}"
    emit_env_line DEVHUB_TRUSTED_PROXIES "${DEVHUB_TRUSTED_PROXIES:-172.16.0.0/12}"
    printf "\n"
    emit_env_line DEVHUB_IDP_PROVIDER "${DEVHUB_IDP_PROVIDER:-keycloak}"
    emit_env_line DEVHUB_OIDC_ISSUER_URL "$DEVHUB_OIDC_ISSUER_URL"
    emit_env_line DEVHUB_OIDC_CLIENT_ID "${DEVHUB_OIDC_CLIENT_ID:-devhub-frontend}"
    emit_env_line DEVHUB_OIDC_CLIENT_SECRET "$DEVHUB_OIDC_CLIENT_SECRET"
    emit_env_line DEVHUB_OIDC_AUDIENCE "${DEVHUB_OIDC_AUDIENCE:-devhub-frontend}"
    emit_env_line DEVHUB_OIDC_JWKS_URL "${DEVHUB_OIDC_JWKS_URL:-}"
    emit_env_line DEVHUB_KEYCLOAK_ADMIN_URL "$DEVHUB_KEYCLOAK_ADMIN_URL"
    emit_env_line DEVHUB_KEYCLOAK_ADMIN_REALM "${DEVHUB_KEYCLOAK_ADMIN_REALM:-devhub}"
    emit_env_line DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID "${DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID:-devhub-backend}"
    emit_env_line DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET "$DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET"
    printf "\n"
    emit_env_line NEXT_PUBLIC_BASE_PATH "${NEXT_PUBLIC_BASE_PATH:-devhub}"
    emit_env_line OIDC_ISSUER_URL "$OIDC_ISSUER_URL"
    emit_env_line OIDC_AUTH_URL "${OIDC_AUTH_URL:-}"
    emit_env_line OIDC_REDIRECT_URI "$OIDC_REDIRECT_URI"
    emit_env_line NEXT_PUBLIC_OIDC_ISSUER_URL "$NEXT_PUBLIC_OIDC_ISSUER_URL"
    emit_env_line NEXT_PUBLIC_OIDC_CLIENT_ID "${NEXT_PUBLIC_OIDC_CLIENT_ID:-devhub-frontend}"
    emit_env_line NEXT_PUBLIC_OIDC_REDIRECT_URI "$NEXT_PUBLIC_OIDC_REDIRECT_URI"
    emit_env_line NEXT_PUBLIC_OIDC_SCOPE "${NEXT_PUBLIC_OIDC_SCOPE:-openid offline_access email profile}"
  } >"$ENV_FILE"
}

build_images() {
  echo "[build] backend-core"
  docker build \
    -f "$ROOT_DIR/backend-core/Dockerfile" \
    -t "${IMAGE_REPO_PREFIX}/backend-core:${IMAGE_TAG}" \
    "$ROOT_DIR/backend-core"
  echo "[build] backend-ai"
  docker build \
    -f "$ROOT_DIR/backend-ai/Dockerfile" \
    -t "${IMAGE_REPO_PREFIX}/backend-ai:${IMAGE_TAG}" \
    "$ROOT_DIR/backend-ai"
  echo "[build] frontend"
  docker build \
    -f "$ROOT_DIR/frontend/Dockerfile" \
    -t "${IMAGE_REPO_PREFIX}/frontend:${IMAGE_TAG}" \
    "$ROOT_DIR/frontend"
}

push_images() {
  echo "[push] backend-core"
  docker push "${IMAGE_REPO_PREFIX}/backend-core:${IMAGE_TAG}"
  echo "[push] backend-ai"
  docker push "${IMAGE_REPO_PREFIX}/backend-ai:${IMAGE_TAG}"
  echo "[push] frontend"
  docker push "${IMAGE_REPO_PREFIX}/frontend:${IMAGE_TAG}"
}

deploy_stack() {
  COMPOSE_FILE="$DOCKER_COMPOSE_FILE" ENV_FILE="$ENV_FILE" "$ROOT_DIR/scripts/deploy-up.sh"
}

main() {
  require IMAGE_TAG
  require IMAGE_REPO_PREFIX
  require DEVHUB_PUBLIC_BASE_URL
  require DEVHUB_OIDC_CLIENT_SECRET
  require DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET

  build_env_file
  echo "Generated deploy env: $ENV_FILE"

  case "$ACTION" in
    build)
      build_images
      ;;
    push)
      push_images
      ;;
    deploy)
      deploy_stack
      ;;
    all)
      build_images
      push_images
      deploy_stack
      ;;
    *)
      echo "ERROR: invalid ACTION=$ACTION (use: build|push|deploy|all)" >&2
      exit 1
      ;;
  esac
}

main "$@"
