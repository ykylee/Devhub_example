#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# =========================
# Inline Config (edit here)
# =========================
# If the same variable is exported in shell, exported value wins.
: "${IMAGE_TAG:=change-me-tag}"
: "${IMAGE_REPO_PREFIX:=ghcr.io/ykylee/devhub_example}"
: "${DEVHUB_PUBLIC_BASE_URL:=http://100.90.113.29:23000}"
: "${DB_URL:=postgres://user:pass@db-host:5432/devhub?sslmode=require}"
: "${DEVHUB_OIDC_CLIENT_SECRET:=change-me-oidc-secret}"
: "${DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET:=change-me-keycloak-admin-secret}"

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
#   IMAGE_REPO_PREFIX               (e.g. ghcr.io/ykylee/devhub_example)
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

build_env_file() {
  local public_base="${DEVHUB_PUBLIC_BASE_URL%/}"
  local base_path="${NEXT_PUBLIC_BASE_PATH:-devhub}"
  local base_path_norm="/${base_path#/}"
  base_path_norm="${base_path_norm%/}"

  DEVHUB_OIDC_ISSUER_URL="${DEVHUB_OIDC_ISSUER_URL:-$public_base/devhub/auth/keycloak/realms/devhub}"
  DEVHUB_KEYCLOAK_ADMIN_URL="${DEVHUB_KEYCLOAK_ADMIN_URL:-$public_base/devhub/auth/keycloak}"
  OIDC_ISSUER_URL="${OIDC_ISSUER_URL:-$DEVHUB_OIDC_ISSUER_URL}"
  OIDC_REDIRECT_URI="${OIDC_REDIRECT_URI:-$public_base$base_path_norm/auth/callback}"
  NEXT_PUBLIC_OIDC_ISSUER_URL="${NEXT_PUBLIC_OIDC_ISSUER_URL:-$OIDC_ISSUER_URL}"
  NEXT_PUBLIC_OIDC_REDIRECT_URI="${NEXT_PUBLIC_OIDC_REDIRECT_URI:-$OIDC_REDIRECT_URI}"

  cat >"$ENV_FILE" <<EOF
IMAGE_TAG=${IMAGE_TAG}
IMAGE_REPO_PREFIX=${IMAGE_REPO_PREFIX}

NGINX_HTTP_PORT=${NGINX_HTTP_PORT:-80}
NGINX_HTTPS_PORT=${NGINX_HTTPS_PORT:-443}
NGINX_TLS_CERT_PATH=${NGINX_TLS_CERT_PATH:-./infra/nginx/certs/tls.crt}
NGINX_TLS_KEY_PATH=${NGINX_TLS_KEY_PATH:-./infra/nginx/certs/tls.key}
DEVHUB_PUBLIC_BASE_URL=${DEVHUB_PUBLIC_BASE_URL}

KEYCLOAK_UPSTREAM=${KEYCLOAK_UPSTREAM:-keycloak:8080}
KEYCLOAK_ADMIN_ALLOW_CIDR=${KEYCLOAK_ADMIN_ALLOW_CIDR:-127.0.0.1}

DB_URL=${DB_URL}
BACKEND_API_URL=${BACKEND_API_URL:-http://backend-core:8080}
BACKEND_AI_URL=${BACKEND_AI_URL:-http://backend-ai:8000}
DEVHUB_AUTH_DEV_FALLBACK=${DEVHUB_AUTH_DEV_FALLBACK:-0}
DEVHUB_TRUSTED_PROXIES=${DEVHUB_TRUSTED_PROXIES:-172.16.0.0/12}

DEVHUB_IDP_PROVIDER=${DEVHUB_IDP_PROVIDER:-keycloak}
DEVHUB_OIDC_ISSUER_URL=${DEVHUB_OIDC_ISSUER_URL}
DEVHUB_OIDC_CLIENT_ID=${DEVHUB_OIDC_CLIENT_ID:-devhub-frontend}
DEVHUB_OIDC_CLIENT_SECRET=${DEVHUB_OIDC_CLIENT_SECRET}
DEVHUB_OIDC_AUDIENCE=${DEVHUB_OIDC_AUDIENCE:-devhub-frontend}
DEVHUB_OIDC_JWKS_URL=${DEVHUB_OIDC_JWKS_URL:-}
DEVHUB_KEYCLOAK_ADMIN_URL=${DEVHUB_KEYCLOAK_ADMIN_URL}
DEVHUB_KEYCLOAK_ADMIN_REALM=${DEVHUB_KEYCLOAK_ADMIN_REALM:-devhub}
DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID=${DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID:-devhub-backend}
DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=${DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET}

NEXT_PUBLIC_BASE_PATH=${NEXT_PUBLIC_BASE_PATH:-devhub}
OIDC_ISSUER_URL=${OIDC_ISSUER_URL}
OIDC_AUTH_URL=${OIDC_AUTH_URL:-}
OIDC_REDIRECT_URI=${OIDC_REDIRECT_URI}
NEXT_PUBLIC_OIDC_ISSUER_URL=${NEXT_PUBLIC_OIDC_ISSUER_URL}
NEXT_PUBLIC_OIDC_CLIENT_ID=${NEXT_PUBLIC_OIDC_CLIENT_ID:-devhub-frontend}
NEXT_PUBLIC_OIDC_REDIRECT_URI=${NEXT_PUBLIC_OIDC_REDIRECT_URI}
NEXT_PUBLIC_OIDC_SCOPE=${NEXT_PUBLIC_OIDC_SCOPE:-openid offline_access email profile}
EOF
}

build_images() {
  echo "[build] backend-core"
  docker build -t "${IMAGE_REPO_PREFIX}/backend-core:${IMAGE_TAG}" "$ROOT_DIR/backend-core"
  echo "[build] backend-ai"
  docker build -t "${IMAGE_REPO_PREFIX}/backend-ai:${IMAGE_TAG}" "$ROOT_DIR/backend-ai"
  echo "[build] frontend"
  docker build -t "${IMAGE_REPO_PREFIX}/frontend:${IMAGE_TAG}" "$ROOT_DIR/frontend"
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
  require DB_URL
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
