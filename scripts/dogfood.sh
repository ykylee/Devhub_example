#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env.dogfood}"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.colima.yml}"
PROJECT_NAME="${PROJECT_NAME:-devhub-dogfood}"
PID_DIR="$ROOT_DIR/.pids/dogfood"
LOG_DIR="$ROOT_DIR/artifacts/dogfood/logs"
AI_VENV_DIR="$ROOT_DIR/backend-ai/.venv-dogfood"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

usage() {
  cat <<'EOF'
Usage:
  ./scripts/dogfood.sh up
  ./scripts/dogfood.sh up --build
  ./scripts/dogfood.sh down
  ./scripts/dogfood.sh restart
  ./scripts/dogfood.sh reset-db
  ./scripts/dogfood.sh reset-all
  ./scripts/dogfood.sh smoke
  ./scripts/dogfood.sh test-onboarding
  ./scripts/dogfood.sh test-integration-admin
  ./scripts/dogfood.sh test-organization-admin
  ./scripts/dogfood.sh test-repository-dashboard
  ./scripts/dogfood.sh test-self-dogfood
  ./scripts/dogfood.sh test-self-dogfood-dashboard
  ./scripts/dogfood.sh status
  ./scripts/dogfood.sh logs [backend|frontend|ai|all]
EOF
}

load_env() {
  if [ ! -f "$ENV_FILE" ]; then
    echo "ERROR: env file not found: $ENV_FILE" >&2
    exit 1
  fi
  # shellcheck disable=SC1090
  set -a
  source "$ENV_FILE"
  set +a
}

compose() {
  (
    cd "$ROOT_DIR"
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -p "$PROJECT_NAME" "$@"
  )
}

is_pid_running() {
  local pid="$1"
  kill -0 "$pid" 2>/dev/null
}

is_port_listening() {
  local port="$1"
  (echo > "/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1
}

listener_pid_for_port() {
  local port="$1"
  lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | head -n 1
}

command_for_pid() {
  local pid="$1"
  ps -p "$pid" -o command= 2>/dev/null || true
}

is_safe_reclaim_target() {
  local name="$1"
  local cmd="$2"

  case "$name" in
    backend)
      [[ "$cmd" == *"backend-core"* ]]
      ;;
    *)
      return 1
      ;;
  esac
}

kill_process_group_or_pid() {
  local pid="$1"
  kill -TERM -- "-$pid" 2>/dev/null || kill "$pid" 2>/dev/null || true
  sleep 1
  if is_pid_running "$pid"; then
    kill -KILL -- "-$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null || true
  fi
}

wait_for_port() {
  local name="$1"
  local port="$2"
  local timeout="${3:-60}"
  local deadline=$(( $(date +%s) + timeout ))

  while [ "$(date +%s)" -lt "$deadline" ]; do
    if is_port_listening "$port"; then
      echo "  $name ready on port $port"
      return 0
    fi
    sleep 1
  done

  echo "ERROR: timed out waiting for $name on port $port" >&2
  return 1
}

wait_for_http() {
  local name="$1"
  local url="$2"
  local timeout="${3:-90}"
  local deadline=$(( $(date +%s) + timeout ))

  while [ "$(date +%s)" -lt "$deadline" ]; do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "  $name ready at $url"
      return 0
    fi
    sleep 2
  done

  echo "ERROR: timed out waiting for $name at $url" >&2
  return 1
}

start_native() {
  local name="$1"
  local workdir="$2"
  local cmd="$3"
  local port="$4"
  local wait_mode="$5"
  local pid_file="$PID_DIR/$name.pid"
  local log_file="$LOG_DIR/$name.log"

  if [ -f "$pid_file" ]; then
    local existing_pid
    existing_pid="$(tr -d '[:space:]' < "$pid_file" 2>/dev/null || true)"
    if [ -n "$existing_pid" ] && is_pid_running "$existing_pid"; then
      echo "  $name already running (PID $existing_pid)"
      return 0
    fi
    rm -f "$pid_file"
  fi

  if is_port_listening "$port"; then
    local listener_pid listener_cmd
    listener_pid="$(listener_pid_for_port "$port" || true)"
    listener_cmd="$(command_for_pid "$listener_pid")"

    if [ -n "$listener_pid" ] && is_safe_reclaim_target "$name" "$listener_cmd"; then
      echo -e "${YELLOW}  reclaiming port $port from existing $name process (PID $listener_pid)${NC}"
      kill_process_group_or_pid "$listener_pid"
      local deadline=$(( $(date +%s) + 10 ))
      while [ "$(date +%s)" -lt "$deadline" ]; do
        if ! is_port_listening "$port"; then
          break
        fi
        sleep 1
      done
    fi

    if is_port_listening "$port"; then
      echo -e "${YELLOW}  port $port already in use; leaving existing listener intact for $name${NC}"
      return 0
    fi
  fi

  echo -e "${GREEN}Starting $name...${NC}"
  local child_pid
  child_pid="$(
    WORKDIR="$workdir" CMD="$cmd" LOG_FILE="$log_file" python3 <<'PY'
import os
import subprocess
import sys

workdir = os.environ["WORKDIR"]
cmd = os.environ["CMD"]
log_file = os.environ["LOG_FILE"]

with open(log_file, "ab", buffering=0) as log:
    proc = subprocess.Popen(
        ["/bin/bash", "-lc", cmd],
        cwd=workdir,
        stdin=subprocess.DEVNULL,
        stdout=log,
        stderr=subprocess.STDOUT,
        start_new_session=True,
        env=os.environ.copy(),
    )
    sys.stdout.write(str(proc.pid))
PY
  )"
  printf '%s\n' "$child_pid" > "$pid_file"

  case "$wait_mode" in
    http)
      wait_for_http "$name" "http://127.0.0.1:$port/health"
      ;;
    port)
      wait_for_port "$name" "$port"
      ;;
    *)
      echo "ERROR: unknown wait mode: $wait_mode" >&2
      exit 1
      ;;
  esac
}

stop_native() {
  local name="$1"
  local pid_file="$PID_DIR/$name.pid"

  if [ ! -f "$pid_file" ]; then
    echo "  $name not started by dogfood.sh"
    return 0
  fi

  local pid
  pid="$(tr -d '[:space:]' < "$pid_file" 2>/dev/null || true)"
  rm -f "$pid_file"

  if [ -z "$pid" ]; then
    return 0
  fi

  if is_pid_running "$pid"; then
    kill_process_group_or_pid "$pid"
    echo "  $name stopped (PID $pid)"
  else
    echo "  $name already stopped (PID $pid)"
  fi
}

setup_keycloak_clients() {
  echo -e "${BLUE}Syncing Keycloak realm/client settings...${NC}"
  (
    cd "$ROOT_DIR"
    KEYCLOAK_URL="$DEVHUB_KEYCLOAK_ADMIN_URL" \
    KC_BOOTSTRAP_ADMIN_USERNAME="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}" \
    KC_BOOTSTRAP_ADMIN_PASSWORD="${KC_BOOTSTRAP_ADMIN_PASSWORD:-admin}" \
    DEVHUB_FRONTEND_ORIGIN="$DEVHUB_PUBLIC_BASE_URL" \
    DEVHUB_FRONTEND_BASEPATH= \
    ./scripts/setup-keycloak.sh
  )
}

run_migrations() {
  if command -v migrate >/dev/null 2>&1; then
    echo -e "${BLUE}Running DB migrations...${NC}"
    (
      cd "$ROOT_DIR"
      migrate -path backend-core/migrations -database "$MIGRATE_DB_URL" up
    )
  else
    echo -e "${YELLOW}migrate not found; skipping migration step${NC}"
  fi
}

backend_ai_command() {
  echo "source \"$AI_VENV_DIR/bin/activate\" && if command -v uvicorn >/dev/null 2>&1; then uvicorn main:app --host 0.0.0.0 --port ${BACKEND_AI_PORT}; else python3 main.py; fi"
}

ensure_backend_ai_venv() {
  if [ ! -d "$AI_VENV_DIR" ]; then
    echo -e "${BLUE}Creating backend-ai virtualenv...${NC}"
    python3 -m venv "$AI_VENV_DIR"
  fi

  if [ ! -f "$AI_VENV_DIR/.deps-installed" ]; then
    echo -e "${BLUE}Installing backend-ai dependencies...${NC}"
    "$AI_VENV_DIR/bin/pip" install -r "$ROOT_DIR/backend-ai/requirements.txt"
    touch "$AI_VENV_DIR/.deps-installed"
  fi
}

up() {
  load_env
  mkdir -p "$PID_DIR" "$LOG_DIR"
  local compose_args=("up" "-d")
  if [ "${2:-}" = "--build" ]; then
    compose_args+=("--build")
  fi

  echo -e "${BLUE}Starting dogfood infra containers...${NC}"
  compose "${compose_args[@]}"

  wait_for_port "dogfood-db" "${POSTGRES_HOST_PORT}" 30
  wait_for_http "dogfood-keycloak" "${DEVHUB_OIDC_ISSUER_URL}/.well-known/openid-configuration" 120

  run_migrations
  setup_keycloak_clients
  ensure_backend_ai_venv

  start_native "backend" "$ROOT_DIR/backend-core" "PORT=${BACKEND_CORE_PORT} go run ." "${BACKEND_CORE_PORT}" "http"
  start_native "ai" "$ROOT_DIR/backend-ai" "$(backend_ai_command)" "${BACKEND_AI_PORT}" "http"
  start_native "frontend" "$ROOT_DIR/frontend" "PORT=${FRONTEND_PORT} npm run dev" "${FRONTEND_PORT}" "port"

  echo
  echo -e "${BLUE}Dogfood environment is up.${NC}"
  echo "  frontend : http://localhost:${FRONTEND_PORT}"
  echo "  backend  : http://localhost:${BACKEND_CORE_PORT}/health"
  echo "  ai       : http://localhost:${BACKEND_AI_PORT}/health"
  echo "  keycloak : ${DEVHUB_OIDC_ISSUER_URL}"
  echo "  postgres : localhost:${POSTGRES_HOST_PORT}"
}

down() {
  load_env
  echo -e "${RED}Stopping dogfood native apps...${NC}"
  stop_native "frontend"
  stop_native "ai"
  stop_native "backend"

  echo -e "${RED}Stopping dogfood containers...${NC}"
  compose down
}

reset_db() {
  load_env
  echo -e "${RED}Resetting dogfood DB/Keycloak data...${NC}"
  echo "  - native apps will be stopped"
  echo "  - dogfood containers will be removed"
  echo "  - dogfood compose volumes will be removed"

  down || true
  compose down -v --remove-orphans

  echo -e "${BLUE}Dogfood DB/Keycloak data reset complete.${NC}"
  echo "Run ./scripts/dogfood.sh up to recreate a fresh environment."
}

reset_all() {
  load_env
  reset_db

  echo -e "${RED}Cleaning dogfood runtime artifacts...${NC}"
  rm -rf "$PID_DIR"
  rm -rf "$LOG_DIR"

  echo -e "${BLUE}Dogfood runtime artifacts removed.${NC}"
  echo "Run ./scripts/dogfood.sh up to recreate everything from scratch."
}

smoke() {
  load_env
  echo -e "${BLUE}Running dogfood smoke checks...${NC}"

  wait_for_http "backend" "http://127.0.0.1:${BACKEND_CORE_PORT}/health" 15
  wait_for_http "ai" "http://127.0.0.1:${BACKEND_AI_PORT}/health" 15
  wait_for_http "frontend" "http://127.0.0.1:${FRONTEND_PORT}/login" 15
  wait_for_http "keycloak" "${DEVHUB_OIDC_ISSUER_URL}/.well-known/openid-configuration" 15
  wait_for_port "postgres" "${POSTGRES_HOST_PORT}" 15

  if [ -n "${GITEA_URL:-}" ] && [ -n "${GITEA_TOKEN:-}" ]; then
    local gitea_version_url="${GITEA_URL%/}/api/v1/version"
    if curl -fsS -H "Authorization: token ${GITEA_TOKEN}" "$gitea_version_url" >/dev/null 2>&1; then
      echo "  gitea ready at $gitea_version_url"
    else
      echo -e "${YELLOW}  gitea check skipped/failed at $gitea_version_url${NC}"
    fi
  fi

  echo -e "${BLUE}Dogfood smoke checks passed.${NC}"
}

test_onboarding() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood onboarding smoke test...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/dogfood-onboarding-smoke.spec.ts
  )
}

test_integration_admin() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood Gitea integration admin scenario...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/dogfood-gitea-integration-admin.spec.ts
  )
}

test_organization_admin() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood organization admin scenario...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/dogfood-organization-admin.spec.ts
  )
}

test_repository_dashboard() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood repository dashboard scenario...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/repository-dashboard.spec.ts
  )
}

test_self_dogfood() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood self-dogfood admin scenario...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/dogfood-self-dogfood-admin.spec.ts
  )
}

test_self_dogfood_dashboard() {
  load_env
  smoke

  echo -e "${BLUE}Running dogfood self-dogfood dashboard scenario...${NC}"
  (
    cd "$ROOT_DIR/frontend"
    export DSN="${DB_URL}"
    export PLAYWRIGHT_BASE_URL="http://localhost:${FRONTEND_PORT}"
    npm run e2e -- tests/e2e/dogfood-self-dogfood-dashboard.spec.ts
  )
}

status() {
  load_env
  echo -e "${BLUE}Compose status${NC}"
  compose ps || true
  echo
  echo -e "${BLUE}Native app status${NC}"
  for name in backend ai frontend; do
    pid_file="$PID_DIR/$name.pid"
    if [ -f "$pid_file" ]; then
      pid="$(tr -d '[:space:]' < "$pid_file" 2>/dev/null || true)"
      if [ -n "$pid" ] && is_pid_running "$pid"; then
        echo "  $name: running (PID $pid)"
      else
        echo "  $name: stale pid file"
      fi
    else
      echo "  $name: not started by dogfood.sh"
    fi
  done
}

logs() {
  local target="${1:-all}"
  case "$target" in
    backend|frontend|ai)
      tail -n 100 -f "$LOG_DIR/$target.log"
      ;;
    all)
      for name in backend ai frontend; do
        echo "===== $name ====="
        if [ -f "$LOG_DIR/$name.log" ]; then
          tail -n 40 "$LOG_DIR/$name.log"
        else
          echo "(no log yet)"
        fi
      done
      ;;
    *)
      echo "ERROR: unsupported log target: $target" >&2
      exit 1
      ;;
  esac
}

cmd="${1:-}"
case "$cmd" in
  up)
    up "${1:-}" "${2:-}"
    ;;
  down)
    down
    ;;
  restart)
    down || true
    up
    ;;
  reset-db)
    reset_db
    ;;
  reset-all)
    reset_all
    ;;
  smoke)
    smoke
    ;;
  test-onboarding)
    test_onboarding
    ;;
  test-integration-admin)
    test_integration_admin
    ;;
  test-organization-admin)
    test_organization_admin
    ;;
  test-repository-dashboard)
    test_repository_dashboard
    ;;
  test-self-dogfood)
    test_self_dogfood
    ;;
  test-self-dogfood-dashboard)
    test_self_dogfood_dashboard
    ;;
  status)
    status
    ;;
  logs)
    logs "${2:-all}"
    ;;
  *)
    usage
    exit 1
    ;;
esac
