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

echo -e "${BLUE}DevHub 로컬 서비스를 시작합니다...${NC}"

PID_DIR="$REPO_ROOT/.pids"
mkdir -p "$PID_DIR"

mask_dsn() {
    # Best-effort credential mask. Robustness 보다는 로그에서 패스워드 한 토막을 가리는 용도.
    echo "$1" | sed -E 's#:[^:@/]+@#:***@#'
}

idp_dsn() {
    # infra/idp/{kratos,hydra}.yaml 의 dsn 은 credential 없이 search_path 만
    # 가지고 있다 (operator/머신 간 평문 credential 공유 회피). Ory 바이너리는
    # 환경변수 DSN 으로 yaml dsn 을 override 하므로, dev-up 가 spawn 직전에
    # DB_URL 에 schema 만 덧붙여 DSN 을 주입한다.
    local dsn=$1
    local schema=$2
    if [[ "$dsn" == *\?* ]]; then
        printf '%s&search_path=%s' "$dsn" "$schema"
    else
        printf '%s?search_path=%s' "$dsn" "$schema"
    fi
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

# 1b. IdP schema + Kratos/Hydra migrations. 셋 다 idempotent (CREATE SCHEMA IF
#     NOT EXISTS; migrate up 은 applied version 을 skip) 이라 warm DB 에서는
#     ~1-2 s 추가 비용. DEVHUB_SKIP_IDP_MIGRATE=1 로 우회 가능.
if [ "${DEVHUB_SKIP_IDP_MIGRATE:-}" = "1" ]; then
    echo "[skip-idp-migrate] IdP schema + migrate 단계 건너뜀."
else
    echo "Applying IdP schemas (hydra, kratos)..."
    ( cd "$REPO_ROOT/backend-core" && go run ./cmd/idp-apply-schemas -dsn "$DB_URL" -sql ../infra/idp/sql/001_create_idp_schemas.sql )
    if command -v kratos >/dev/null 2>&1; then
        echo "Applying kratos migrations..."
        kratos migrate sql up --yes "$(idp_dsn "$DB_URL" kratos)"
    else
        echo -e "${YELLOW}kratos 명령을 찾을 수 없음 — kratos migrate skip.${NC}"
    fi
    if command -v hydra >/dev/null 2>&1; then
        echo "Applying hydra migrations..."
        hydra migrate sql up --yes "$(idp_dsn "$DB_URL" hydra)"
    else
        echo -e "${YELLOW}hydra 명령을 찾을 수 없음 — hydra migrate skip.${NC}"
    fi
fi

# 2. Kratos
if is_port_listening 4433; then
    echo "  external instance detected on port 4433; using existing kratos (PID 파일 미작성)"
elif command -v kratos >/dev/null 2>&1; then
    export DSN="$(idp_dsn "$DB_URL" kratos)"
    run_service "kratos" "kratos serve -c infra/idp/kratos.yaml --dev" "kratos.log" ""
    unset DSN
    wait_for_port "kratos-public" 4433
    wait_for_port "kratos-admin"  4434
else
    echo -e "${YELLOW}kratos 명령을 찾을 수 없습니다. 인증이 동작하지 않습니다.${NC}"
fi

# 3. Hydra
if is_port_listening 4444; then
    echo "  external instance detected on port 4444; using existing hydra (PID 파일 미작성)"
elif command -v hydra >/dev/null 2>&1; then
    export DSN="$(idp_dsn "$DB_URL" hydra)"
    run_service "hydra" "hydra serve all -c infra/idp/hydra.yaml --dev" "hydra.log" ""
    unset DSN
    wait_for_port "hydra-public" 4444
    wait_for_port "hydra-admin"  4445
else
    echo -e "${YELLOW}hydra 명령을 찾을 수 없습니다. OIDC 코드 흐름이 완성되지 않습니다.${NC}"
fi

# 3b. OIDC client 등록 (멱등 — DELETE 후 POST). hydra-admin (4445) 미가동
#     시 skip. DEVHUB_SKIP_OIDC_REGISTER=1 로 우회 가능.
if [ "${DEVHUB_SKIP_OIDC_REGISTER:-}" = "1" ]; then
    echo "[skip-oidc-register] OIDC client 등록 단계 건너뜀."
elif is_port_listening 4445; then
    echo "Registering OIDC client (devhub-frontend)..."
    "$REPO_ROOT/infra/idp/scripts/register-devhub-client.sh"
else
    echo -e "${YELLOW}Hydra admin (4445) 미가동 — OIDC client 등록 skip.${NC}"
fi

# 4. backend-core
export AUTH_DEV_FALLBACK=true
export DEVHUB_AUTH_DEV_FALLBACK=1
export DEVHUB_KRATOS_PUBLIC_URL="${DEVHUB_KRATOS_PUBLIC_URL:-http://localhost:4433}"
export DEVHUB_KRATOS_ADMIN_URL="${DEVHUB_KRATOS_ADMIN_URL:-http://localhost:4434}"
export DEVHUB_HYDRA_PUBLIC_URL="${DEVHUB_HYDRA_PUBLIC_URL:-http://localhost:4444}"
export DEVHUB_HYDRA_ADMIN_URL="${DEVHUB_HYDRA_ADMIN_URL:-http://localhost:4445}"
if is_port_listening 8080; then
    echo "  external instance detected on port 8080; using existing backend (PID 파일 미작성)"
else
    # `go run` 은 grandchild 프로세스를 띄워 실제 listener 가 launcher PID 와
    # 일치하지 않는다. 빌드 산출물을 직접 실행하면 .pids/backend.pid 가 8080
    # 의 owner 와 같아져 dev-down 의 PID-kill 이 실제 서버를 종료한다.
    backend_bin="$REPO_ROOT/dev-bin/backend-core"
    mkdir -p "$(dirname "$backend_bin")"
    echo "Compiling backend..."
    # backend-core 는 자체 go.mod 를 가진 모듈이라 빌드는 그 디렉터리 안에서
    # 실행해야 한다. 산출물은 위 dev-bin/ 으로 다시 돌려 둔다.
    ( cd "$REPO_ROOT/backend-core" && go build -o "$backend_bin" . )
    run_service "backend" "$backend_bin" "backend.log" "backend-core"
    wait_for_port "backend" 8080
fi

# 5. frontend
if is_port_listening 3000; then
    echo "  external instance detected on port 3000; using existing frontend (PID 파일 미작성)"
else
    run_service "frontend" "npm run dev" "frontend.log" "frontend"
    wait_for_port "frontend" 3000 60
fi

echo ""
echo -e "${BLUE}모든 서비스 기동 완료:${NC}"
echo "  kratos     public 4433, admin 4434"
echo "  hydra      public 4444, admin 4445"
echo "  backend    8080  (http://localhost:8080/health)"
echo "  frontend   3000  (http://localhost:3000/)"
echo ""
echo "종료: ./dev-down.sh"
