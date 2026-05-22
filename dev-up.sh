#!/bin/bash
#
# dev-up.sh — DevHub 로컬 서비스 launcher (macOS/Linux).
#
# 실행:
#   ./dev-up.sh              # 전체 기동
#   ./dev-up.sh restart      # dev-down.sh 로 정리 후 재기동
#
# 환경 변수:
#   DB_URL                   PostgreSQL DSN (기본: postgres://${USER}@localhost:5432/devhub?sslmode=disable)
#   DEVHUB_SKIP_MIGRATE      1 이면 migrate-up 단계를 건너뜀
#   DEVHUB_SKIP_READY_WAIT   1 이면 TCP readiness probe 를 건너뜀 (빠르지만 위험)
#
# Windows 환경은 dev-up.ps1 사용 (ASCII 영어, 동일 동작).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

if [ "${1:-}" = "restart" ]; then
    echo -e "${YELLOW}서비스 재시작 (dev-down.sh 먼저)...${NC}"
    ./dev-down.sh
    sleep 2
fi

echo -e "${BLUE}DevHub 로컬 서비스를 시작합니다 (Keycloak IdP 기반)...${NC}"

PID_DIR="$REPO_ROOT/.pids"
mkdir -p "$PID_DIR"

mask_dsn() {
    # Best-effort credential mask. Robustness 보다는 로그에서 패스워드 한 토막을 가리는 용도.
    echo "$1" | sed -E 's#:[^:@/]+@#:***@#'
}

is_port_listening() {
    # 외부 관리 인스턴스 존중용 즉시 체크. wait_for_port 와 달리 deadline 없이
    # 한 번만 시도한다. /dev/tcp 는 bash 내장이라 nc 의존성을 피한다.
    local port=$1
    (echo > "/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1
}

wait_for_port() {
    local name=$1
    local port=$2
    local timeout=${3:-30}
    if [ "${DEVHUB_SKIP_READY_WAIT:-}" = "1" ]; then
        echo "  [skip-ready-wait] $name on port $port"
        return 0
    fi
    local deadline=$(( $(date +%s) + timeout ))
    while [ "$(date +%s)" -lt "$deadline" ]; do
        # bash /dev/tcp 는 macOS/Linux 모두 지원. nc 의존성 회피.
        if (echo > "/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1; then
            echo "  $name ready on port $port"
            return 0
        fi
        sleep 0.25
    done
    echo -e "${RED}Timed out after ${timeout}s waiting for $name on port $port. 해당 .log 파일 확인.${NC}" >&2
    return 1
}

run_service() {
    local name=$1
    local cmd=$2
    local log_file=$3
    local working_dir=$4
    local pid_file="$PID_DIR/$name.pid"
    local abs_log
    if [[ "$log_file" = /* ]]; then
        abs_log="$log_file"
    else
        abs_log="$REPO_ROOT/$log_file"
    fi

    echo -e "${GREEN}Starting $name (cwd: ${working_dir:-root})...${NC}"
    if [ -n "$working_dir" ]; then
        ( cd "$working_dir" && exec $cmd > "$abs_log" 2>&1 ) &
    else
        $cmd > "$abs_log" 2>&1 &
    fi
    echo $! > "$pid_file"
}

# 1. DB migrations (idempotent). golang-migrate 없으면 경고만 내고 진행 — N1 회귀
#    가능성은 남지만 backend 가 부팅 자체는 가능. `make migrate-tools` 가 권장 경로.
DB_URL="${DB_URL:-postgres://${USER}@localhost:5432/devhub?sslmode=disable}"
export DB_URL

if [ "${DEVHUB_SKIP_MIGRATE:-}" = "1" ]; then
    echo "[skip-migrate] migrate-up 단계 건너뜀."
elif command -v migrate >/dev/null 2>&1; then
    echo "Applying migrations against $(mask_dsn "$DB_URL")..."
    migrate -path backend-core/migrations -database "$DB_URL" up
else
    echo -e "${YELLOW}migrate 명령을 찾을 수 없음. 한 번 \`make migrate-tools\` 실행 권장. (DEVHUB_SKIP_MIGRATE=1 로 무음 가능)${NC}"
fi

# 2. Keycloak (local dev)
# Keycloak 은 보통 docker 로 띄우는 것을 권장하나, 8180 포트가 비어있으면 경고를 낸다.
# KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak 를 명시해 dev 와 prod 의 issuer path 구조를 일관화
# (docs/reports/2026-05-20-network-docker-single-port-review.md §4.8 정정).
if is_port_listening 8180; then
    echo "  Keycloak detected on port 8180."
else
    echo -e "${YELLOW}Warning: Keycloak not detected on port 8180. 인증이 동작하지 않을 수 있습니다.${NC}"
    echo "  docker run -d -p 8180:8080 \\"
    echo "    -e KC_BOOTSTRAP_ADMIN_USERNAME=admin -e KC_BOOTSTRAP_ADMIN_PASSWORD=admin \\"
    echo "    -e KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak \\"
    echo "    quay.io/keycloak/keycloak:25.0 start-dev --http-enabled=true --hostname-strict=false"
fi

# 3. backend-core
export AUTH_DEV_FALLBACK=true
export DEVHUB_AUTH_DEV_FALLBACK=1
# Keycloak default envs for local dev — issuer path 가 prod 와 동일 prefix (/devhub/auth/keycloak/...).
export DEVHUB_IDP_PROVIDER="${DEVHUB_IDP_PROVIDER:-keycloak}"
export DEVHUB_OIDC_ISSUER_URL="${DEVHUB_OIDC_ISSUER_URL:-http://localhost:8180/devhub/auth/keycloak/realms/devhub}"
export DEVHUB_OIDC_CLIENT_ID="${DEVHUB_OIDC_CLIENT_ID:-devhub-frontend}"

if is_port_listening 8080; then
    echo "  external instance detected on port 8080; using existing backend (PID 파일 미작성)"
else
    backend_bin="$REPO_ROOT/dev-bin/backend-core"
    mkdir -p "$(dirname "$backend_bin")"
    echo "Compiling backend..."
    ( cd "$REPO_ROOT/backend-core" && go build -o "$backend_bin" . )
    run_service "backend" "$backend_bin" "backend.log" "backend-core"
    wait_for_port "backend" 8080
fi

# 4. frontend
if is_port_listening 3000; then
    echo "  external instance detected on port 3000; using existing frontend (PID 파일 미작성)"
else
    run_service "frontend" "npm run dev" "frontend.log" "frontend"
    wait_for_port "frontend" 3000 60
fi

echo ""
echo -e "${BLUE}모든 서비스 기동 완료:${NC}"
echo "  backend    8080  (http://localhost:8080/health)"
echo "  frontend   3000  (http://localhost:3000/)"
echo ""
echo "종료: ./dev-down.sh"
