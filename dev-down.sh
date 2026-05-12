#!/bin/bash
#
# dev-down.sh — DevHub 로컬 서비스 종료 (macOS/Linux).
#
# dev-up.sh 가 .pids/<name>.pid 에 남긴 PID 를 역순으로 종료하고, 분실된
# PID 파일에 대비해 알려진 포트들도 한 번 더 쓸어담는다.
#
# Windows 환경은 dev-down.ps1 사용.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

RED='\033[0;31m'
NC='\033[0m'

PID_DIR="$REPO_ROOT/.pids"

stop_service() {
    local name=$1
    local pid_file="$PID_DIR/$name.pid"

    if [ ! -f "$pid_file" ]; then
        echo "  $name not tracked (no PID file)"
        return 0
    fi

    local svc_pid
    svc_pid=$(cat "$pid_file" 2>/dev/null | tr -d '[:space:]')
    rm -f "$pid_file"

    if [ -z "$svc_pid" ]; then
        return 0
    fi

    if kill -0 "$svc_pid" 2>/dev/null; then
        kill "$svc_pid" 2>/dev/null || true
        sleep 0.5
        if kill -0 "$svc_pid" 2>/dev/null; then
            kill -9 "$svc_pid" 2>/dev/null || true
        fi
        echo "  $name stopped (PID $svc_pid)"
    else
        echo "  $name (PID $svc_pid) already gone"
    fi
}

sweep_port() {
    # Safety net: any leftover process holding a known DevHub port gets killed.
    local port=$1
    if command -v lsof >/dev/null 2>&1; then
        local pids
        pids=$(lsof -ti:"$port" 2>/dev/null || true)
        if [ -n "$pids" ]; then
            echo "  sweeping pids on port $port: $pids"
            echo "$pids" | xargs -r kill -9 2>/dev/null || true
        fi
    fi
}

echo -e "${RED}DevHub 로컬 서비스 종료...${NC}"

# Reverse start order: frontend -> backend -> hydra -> kratos.
stop_service "frontend"
stop_service "backend"
stop_service "hydra"
stop_service "kratos"

for p in 3000 8080 4433 4434 4444 4445; do
    sweep_port "$p"
done

echo -e "${RED}모든 서비스가 종료되었습니다.${NC}"
