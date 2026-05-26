#!/usr/bin/env bash
set -euo pipefail

# =========================
# Prerequisites (host tools)
# =========================
# - python3 (3.8+, json 표준 라이브러리만 사용) — generate_local_realm_import() heredoc 및 setup-keycloak.sh 호출 chain.
# - curl                                          — Keycloak Admin REST + readiness wait (setup-keycloak.sh).
# - docker + docker compose (v2 plugin)           — runtime image build + scripts/deploy-up.sh.
# - bash 4+                                       — array / heredoc / [[ syntax.
# 검증: `python3 --version && curl --version && docker compose version`.
# 자세한 fail-mode + 우회 절차: docs/setup/docker-packaging-deployment-guide.md §1.1.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# =========================
# Inline Config (edit here)
# =========================
# If the same variable is exported in shell, exported value wins.
: "${IMAGE_TAG:=change-me-tag}"
: "${IMAGE_REPO_PREFIX:=local/devhub}"
: "${PUBLIC_ACCESS_SCHEME:=http}"
: "${PUBLIC_ACCESS_HOST:=localhost}"
: "${PUBLIC_ACCESS_PORT:=13000}"
: "${DEVHUB_PUBLIC_BASE_URL:=}"
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
: "${ACTION:=all}" # build|deploy|all
: "${ENV_FILE:=/tmp/devhub-deploy.env}"
: "${DOCKER_COMPOSE_FILE:=$ROOT_DIR/docker-compose.deploy.yml}"
: "${NEXT_PUBLIC_BASE_PATH:=devhub}"
: "${KEYCLOAK_UPSTREAM:=keycloak:8080}"
: "${KEYCLOAK_ADMIN_ALLOW_CIDR:=127.0.0.1/32}"
: "${NGINX_HTTP_PORT:=3000}"
: "${AUTO_CONFIGURE_KEYCLOAK_REDIRECTS:=1}"
# repo 안 안정 path (.gitignore 추적 외). 직전 /tmp/* 는 일부 호스트 (tmpfs / Docker
# Desktop 의 file mount 제한) 에서 container 재시작 시 사라지는 risk 가 있어 .build/
# 로 이전. .build/ 디렉토리는 generate_local_realm_import() 가 mkdir -p 로 자동 생성.
: "${GENERATED_KEYCLOAK_REALM_IMPORT:=$ROOT_DIR/.build/devhub-keycloak-realm.generated.json}"

# Simple one-shot deploy helper.
# Required env:
#   IMAGE_TAG
#   IMAGE_REPO_PREFIX               (default: local/devhub)
#   PUBLIC_ACCESS_HOST / PORT       external client access endpoint (e.g. host:13000)
#   DEVHUB_PUBLIC_BASE_URL          optional override, derived from PUBLIC_ACCESS_* when empty
#   DB_URL
#   DEVHUB_OIDC_CLIENT_SECRET
#   DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET
#
# Optional env:
#   ACTION                          build|deploy|all (default: all)
#   ENV_FILE                        deploy env file path (default: /tmp/devhub-deploy.env)
#   DOCKER_COMPOSE_FILE             compose file path (default: docker-compose.deploy.yml)
#   DEVHUB_OIDC_ISSUER_URL          default: ${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}:${PUBLIC_ACCESS_PORT}/devhub/auth/keycloak/realms/devhub
#   DEVHUB_KEYCLOAK_ADMIN_URL       default: ${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}:${PUBLIC_ACCESS_PORT}/devhub/auth/keycloak
#   OIDC_ISSUER_URL                 default: DEVHUB_OIDC_ISSUER_URL
#   OIDC_REDIRECT_URI               default: ${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}:${PUBLIC_ACCESS_PORT}/devhub/auth/callback
#   NEXT_PUBLIC_OIDC_ISSUER_URL     default: OIDC_ISSUER_URL
#   NEXT_PUBLIC_OIDC_REDIRECT_URI   default: OIDC_REDIRECT_URI
#   NEXT_PUBLIC_BASE_PATH           default: devhub
#   NEXT_PUBLIC_OIDC_CLIENT_ID      default: devhub-frontend
#   DEVHUB_OIDC_CLIENT_ID           default: devhub-frontend
#   DEVHUB_OIDC_AUDIENCE            default: devhub-frontend
#   DEVHUB_KEYCLOAK_ADMIN_REALM     default: devhub
#   DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID default: devhub-backend
#   DEVHUB_AUTH_DEV_FALLBACK        default: 0
#   DEVHUB_ONBOARDING_GATE_ENABLED  default: 1 (운영 안전). 내부 Keycloak 테스트 시
#                                    `0` 으로 두면 backend onboarding gate 해제 →
#                                    Keycloak 계정 자유 사용 가능 (frontend 는 첫
#                                    진입 시 /onboarding → Skip 1회 클릭 필요).
#   BACKEND_API_URL                 default: http://backend-core:8080
#   BACKEND_AI_URL                  default: http://backend-ai:8000
#   DEVHUB_TRUSTED_PROXIES          default: 172.16.0.0/12
#   KEYCLOAK_UPSTREAM               default: keycloak:8080
#   KEYCLOAK_ADMIN_ALLOW_CIDR       default: 127.0.0.1/32
#   DB_MODE                         external|docker (docker will bring up db/db-init via compose profile)
#   POSTGRES_USER/PASSWORD/DB       default: user/pass/devhub
#   DB_HOST/DB_PORT/DB_SSLMODE      default: db/5432/disable
#   NGINX_HTTP_PORT                 default: 3000 (VM ingress port)
#   AUTO_CONFIGURE_KEYCLOAK_REDIRECTS default: 1 (local-idp deploy 후 redirect/webOrigin 자동 동기화)
#   GENERATED_KEYCLOAK_REALM_IMPORT default: $ROOT_DIR/.build/devhub-keycloak-realm.generated.json (repo 안 안정 path, .gitignore 추적 외)

require() {
  local var_name="$1"
  if [ -z "${!var_name:-}" ]; then
    echo "ERROR: required env var is empty: $var_name" >&2
    exit 1
  fi
}

validate_cidr() {
  local cidr="$1"
  if ! [[ "$cidr" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}/([0-9]|[12][0-9]|3[0-2])$ ]]; then
    return 1
  fi

  local ip="${cidr%/*}"
  local octet
  IFS='.' read -r -a octets <<< "$ip"
  if [ "${#octets[@]}" -ne 4 ]; then
    return 1
  fi
  for octet in "${octets[@]}"; do
    if [ "$octet" -lt 0 ] || [ "$octet" -gt 255 ]; then
      return 1
    fi
  done
}

normalize_keycloak_hostname() {
  local hostname="$1"
  local normalized="${hostname%/}"
  if [[ "$normalized" == */devhub/auth/keycloak ]]; then
    printf "%s" "$normalized"
    return
  fi

  # If only scheme://host[:port] is provided, append the Keycloak proxy path.
  local remainder="${normalized#*://}"
  if [[ "$normalized" =~ ^https?:// ]] && [[ "$remainder" != */* ]]; then
    printf "%s/devhub/auth/keycloak" "$normalized"
    return
  fi

  printf "%s" "$normalized"
}

emit_env_line() {
  local key="$1"
  local value="$2"
  local escaped="${value//\'/\'\"\'\"\'}"
  printf "%s='%s'\n" "$key" "$escaped"
}

generate_local_realm_import() {
  local public_base="$1"
  local base_path_norm="$2"
  local source_realm="${KEYCLOAK_REALM_IMPORT_TEMPLATE:-$ROOT_DIR/infra/idp/keycloak-realm.dev.json}"
  local output_realm="${GENERATED_KEYCLOAK_REALM_IMPORT:-$ROOT_DIR/.build/devhub-keycloak-realm.generated.json}"
  local redirect_uri="${public_base}${base_path_norm}/auth/callback"
  local post_logout_a="${public_base}${base_path_norm}/*"
  local post_logout_b="${public_base}${base_path_norm}/"

  mkdir -p "$(dirname "$output_realm")"

  python3 - "$source_realm" "$output_realm" "$public_base" "$redirect_uri" "$post_logout_a" "$post_logout_b" <<'PY'
import json
import pathlib
import sys

src = pathlib.Path(sys.argv[1])
dst = pathlib.Path(sys.argv[2])
origin = sys.argv[3]
redirect_uri = sys.argv[4]
post_logout_a = sys.argv[5]
post_logout_b = sys.argv[6]

doc = json.loads(src.read_text(encoding="utf-8"))
for client in doc.get("clients", []):
    if client.get("clientId") != "devhub-frontend":
        continue

    redirect_uris = client.get("redirectUris") or []
    if redirect_uri not in redirect_uris:
        redirect_uris.append(redirect_uri)
    client["redirectUris"] = redirect_uris

    web_origins = client.get("webOrigins") or []
    if origin not in web_origins:
        web_origins.append(origin)
    client["webOrigins"] = web_origins

    attrs = client.get("attributes") or {}
    entries = [x for x in (attrs.get("post.logout.redirect.uris") or "").split("##") if x]
    for candidate in (post_logout_a, post_logout_b):
        if candidate not in entries:
            entries.append(candidate)
    attrs["post.logout.redirect.uris"] = "##".join(entries)
    client["attributes"] = attrs
    break

dst.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(str(dst))
PY
}

build_env_file() {
  local public_base="${DEVHUB_PUBLIC_BASE_URL%/}"
  if [ -z "$public_base" ]; then
    if [ -n "${PUBLIC_ACCESS_PORT:-}" ]; then
      public_base="${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}:${PUBLIC_ACCESS_PORT}"
    else
      public_base="${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}"
    fi
  fi
  local base_path="${NEXT_PUBLIC_BASE_PATH:-devhub}"
  local base_path_norm="/${base_path#/}"
  base_path_norm="${base_path_norm%/}"
  local nginx_http_port="${NGINX_HTTP_PORT:-}"
  if [ "$DB_MODE" = "docker" ]; then
    COMPOSE_PROFILES="local-db,local-idp"
    DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${DB_HOST}:${DB_PORT}/${POSTGRES_DB}?sslmode=${DB_SSLMODE}"
    KEYCLOAK_HOSTNAME="${KEYCLOAK_HOSTNAME:-$public_base/devhub/auth/keycloak}"
    KEYCLOAK_HOSTNAME="$(normalize_keycloak_hostname "$KEYCLOAK_HOSTNAME")"
    DEVHUB_KEYCLOAK_SSL_REQUIRED="${DEVHUB_KEYCLOAK_SSL_REQUIRED:-none}"
    KEYCLOAK_REALM_IMPORT_PATH="$(generate_local_realm_import "$public_base" "$base_path_norm")"
  elif [ "$DB_MODE" = "external" ]; then
    : "${DB_URL:?set DB_URL when DB_MODE=external}"
    COMPOSE_PROFILES="${COMPOSE_PROFILES:-}"
    DEVHUB_KEYCLOAK_SSL_REQUIRED="${DEVHUB_KEYCLOAK_SSL_REQUIRED:-external}"
  else
    echo "ERROR: invalid DB_MODE=$DB_MODE (use: external|docker)" >&2
    exit 1
  fi

  local internal_keycloak_base="${INTERNAL_KEYCLOAK_BASE_URL:-http://keycloak:8080/devhub/auth/keycloak}"
  if [[ ",${COMPOSE_PROFILES:-}," == *",local-idp,"* ]]; then
    # OIDC issuer must match the URL seen by browser tokens (iss claim),
    # while backend admin API uses the internal service DNS.
    DEVHUB_OIDC_ISSUER_URL="${DEVHUB_OIDC_ISSUER_URL:-$public_base/devhub/auth/keycloak/realms/devhub}"
    DEVHUB_KEYCLOAK_ADMIN_URL="${DEVHUB_KEYCLOAK_ADMIN_URL:-$internal_keycloak_base}"
    # backend-core runs in container network. Explicit JWKS avoids resolving
    # localhost to the container itself in local-idp deploys.
    local docker_host_base="http://host.docker.internal"
    if [ -n "${PUBLIC_ACCESS_PORT:-}" ]; then
      docker_host_base="${docker_host_base}:${PUBLIC_ACCESS_PORT}"
    fi
    DEVHUB_OIDC_JWKS_URL="${DEVHUB_OIDC_JWKS_URL:-$docker_host_base/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/certs}"
  else
    DEVHUB_OIDC_ISSUER_URL="${DEVHUB_OIDC_ISSUER_URL:-$public_base/devhub/auth/keycloak/realms/devhub}"
    DEVHUB_KEYCLOAK_ADMIN_URL="${DEVHUB_KEYCLOAK_ADMIN_URL:-$public_base/devhub/auth/keycloak}"
  fi
  OIDC_ISSUER_URL="${OIDC_ISSUER_URL:-$public_base/devhub/auth/keycloak/realms/devhub}"
  OIDC_REDIRECT_URI="${OIDC_REDIRECT_URI:-$public_base$base_path_norm/auth/callback}"
  NEXT_PUBLIC_OIDC_ISSUER_URL="${NEXT_PUBLIC_OIDC_ISSUER_URL:-$OIDC_ISSUER_URL}"
  NEXT_PUBLIC_OIDC_REDIRECT_URI="${NEXT_PUBLIC_OIDC_REDIRECT_URI:-$OIDC_REDIRECT_URI}"

  {
    emit_env_line IMAGE_TAG "$IMAGE_TAG"
    emit_env_line IMAGE_REPO_PREFIX "$IMAGE_REPO_PREFIX"
    emit_env_line PUBLIC_ACCESS_SCHEME "$PUBLIC_ACCESS_SCHEME"
    emit_env_line PUBLIC_ACCESS_HOST "$PUBLIC_ACCESS_HOST"
    emit_env_line PUBLIC_ACCESS_PORT "$PUBLIC_ACCESS_PORT"
    emit_env_line DB_MODE "$DB_MODE"
    if [ -n "${COMPOSE_PROFILES:-}" ]; then
      emit_env_line COMPOSE_PROFILES "$COMPOSE_PROFILES"
    fi
    printf "\n"
    emit_env_line NGINX_HTTP_PORT "$nginx_http_port"
    emit_env_line DEVHUB_PUBLIC_BASE_URL "$public_base"
    emit_env_line NEXT_PUBLIC_APP_ORIGIN "$public_base"
    printf "\n"
    emit_env_line KEYCLOAK_UPSTREAM "${KEYCLOAK_UPSTREAM:-keycloak:8080}"
    emit_env_line KEYCLOAK_ADMIN_ALLOW_CIDR "${KEYCLOAK_ADMIN_ALLOW_CIDR:-127.0.0.1/32}"
    emit_env_line KEYCLOAK_HOSTNAME "${KEYCLOAK_HOSTNAME:-localhost}"
    emit_env_line KC_BOOTSTRAP_ADMIN_USERNAME "${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
    emit_env_line KC_BOOTSTRAP_ADMIN_PASSWORD "${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
    emit_env_line KC_DB_URL "${KC_DB_URL:-jdbc:postgresql://db:5432/devhub}"
    emit_env_line KC_DB_USERNAME "${KC_DB_USERNAME:-user}"
    emit_env_line KC_DB_PASSWORD "${KC_DB_PASSWORD:-pass}"
    emit_env_line KC_DB_SCHEMA "${KC_DB_SCHEMA:-keycloak}"
    emit_env_line DEVHUB_KEYCLOAK_SSL_REQUIRED "${DEVHUB_KEYCLOAK_SSL_REQUIRED:-none}"
    # KEYCLOAK_REALM_IMPORT_PATH 는 docker-compose 의 keycloak service (profiles:
    # ["local-idp"]) volume mount 가 참조하므로 local-idp profile 일 때만 emit.
    # external mode (사내 운영 Keycloak) 에서는 keycloak container 자체가 미가동
    # → mount noop. env 파일에 dev.json fallback 이 emit 되면 운영자가 dev.json
    # 가 prod 에 적용된다고 오독할 risk 가 있어 분기.
    if [[ ",${COMPOSE_PROFILES:-}," == *",local-idp,"* ]]; then
      emit_env_line KEYCLOAK_REALM_IMPORT_PATH "${KEYCLOAK_REALM_IMPORT_PATH:-$ROOT_DIR/infra/idp/keycloak-realm.dev.json}"
    fi
    printf "\n"
    emit_env_line DB_URL "$DB_URL"
    emit_env_line POSTGRES_USER "$POSTGRES_USER"
    emit_env_line POSTGRES_PASSWORD "$POSTGRES_PASSWORD"
    emit_env_line POSTGRES_DB "$POSTGRES_DB"
    emit_env_line BACKEND_API_URL "${BACKEND_API_URL:-http://backend-core:8080}"
    emit_env_line BACKEND_AI_URL "${BACKEND_AI_URL:-http://backend-ai:8000}"
    emit_env_line DEVHUB_AUTH_DEV_FALLBACK "${DEVHUB_AUTH_DEV_FALLBACK:-0}"
    emit_env_line DEVHUB_ONBOARDING_GATE_ENABLED "${DEVHUB_ONBOARDING_GATE_ENABLED:-1}"
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
  echo "[stage] build host artifacts"
  ENV_FILE="$ENV_FILE" "$ROOT_DIR/scripts/build-artifacts.sh"

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

deploy_stack() {
  COMPOSE_FILE="$DOCKER_COMPOSE_FILE" ENV_FILE="$ENV_FILE" "$ROOT_DIR/scripts/deploy-up.sh"
}

sync_keycloak_redirects() {
  if [[ ",${COMPOSE_PROFILES:-}," != *",local-idp,"* ]]; then
    return 0
  fi
  if [ "${AUTO_CONFIGURE_KEYCLOAK_REDIRECTS:-1}" != "1" ]; then
    echo "[post-deploy] skip Keycloak redirect sync (AUTO_CONFIGURE_KEYCLOAK_REDIRECTS=${AUTO_CONFIGURE_KEYCLOAK_REDIRECTS})"
    return 0
  fi

  local public_base="${DEVHUB_PUBLIC_BASE_URL%/}"
  if [ -z "$public_base" ]; then
    if [ -n "${PUBLIC_ACCESS_PORT:-}" ]; then
      public_base="${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}:${PUBLIC_ACCESS_PORT}"
    else
      public_base="${PUBLIC_ACCESS_SCHEME}://${PUBLIC_ACCESS_HOST}"
    fi
  fi
  local base_path="/${NEXT_PUBLIC_BASE_PATH:-devhub}"
  base_path="${base_path%/}"

  local local_keycloak_url
  if [ -n "${PUBLIC_ACCESS_PORT:-}" ]; then
    local_keycloak_url="${PUBLIC_ACCESS_SCHEME}://127.0.0.1:${PUBLIC_ACCESS_PORT}/devhub/auth/keycloak"
  else
    local_keycloak_url="${PUBLIC_ACCESS_SCHEME}://127.0.0.1/devhub/auth/keycloak"
  fi

  echo "[post-deploy] sync Keycloak redirect/webOrigin with ${public_base}${base_path} (via ${local_keycloak_url})"
  KEYCLOAK_URL="$local_keycloak_url" \
    DEVHUB_FRONTEND_ORIGIN="$public_base" \
    DEVHUB_FRONTEND_BASEPATH="$base_path" \
    DEVHUB_KEYCLOAK_SSL_REQUIRED="${DEVHUB_KEYCLOAK_SSL_REQUIRED:-none}" \
    KC_BOOTSTRAP_ADMIN_USERNAME="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}" \
    KC_BOOTSTRAP_ADMIN_PASSWORD="${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}" \
    "$ROOT_DIR/scripts/setup-keycloak.sh"
}

main() {
  require IMAGE_TAG
  require IMAGE_REPO_PREFIX
  require DEVHUB_OIDC_CLIENT_SECRET
  require DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET
  if [ -z "${DEVHUB_PUBLIC_BASE_URL:-}" ]; then
    require PUBLIC_ACCESS_HOST
  fi
  if [ "$DB_MODE" = "external" ]; then
    require DB_URL
  fi
  if ! validate_cidr "${KEYCLOAK_ADMIN_ALLOW_CIDR:-}"; then
    echo "ERROR: KEYCLOAK_ADMIN_ALLOW_CIDR must be CIDR form (e.g. 127.0.0.1/32, 10.0.0.0/8)" >&2
    exit 1
  fi

  build_env_file
  echo "Generated deploy env: $ENV_FILE"

  case "$ACTION" in
    build)
      build_images
      ;;
    deploy)
      deploy_stack
      sync_keycloak_redirects
      ;;
    all)
      build_images
      deploy_stack
      sync_keycloak_redirects
      ;;
    *)
      echo "ERROR: invalid ACTION=$ACTION (use: build|deploy|all)" >&2
      exit 1
      ;;
  esac
}

main "$@"
