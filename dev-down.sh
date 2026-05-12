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
    # exit 0 => PID 파일이 있어 dev-up 가 spawn 한 서비스 (sweep 대상에 추가)
    # exit 1 => PID 파일이 없어 외부 관리 인스턴스로 간주 (그대로 둠)
    local name=$1
    local pid_file="$PID_DIR/$name.pid"

    if [ ! -f "$pid_file" ]; then
        return 1
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
    return 0
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

# Reverse start order: frontend -> backend -> hydra -> kratos. dev-up 가 실제로
# spawn 한 서비스(PID 파일 존재) 의 포트만 sweep 대상에 누적해, 외부에서 직접
# 띄운 Kratos/Hydra/backend/frontend 인스턴스는 건드리지 않는다.
declare -a sweep_ports=()

declare -a services=("frontend" "backend" "hydra" "kratos")
declare -A service_ports=(
    [frontend]="3000"
    [backend]="8080"
    [hydra]="4444 4445"
    [kratos]="4433 4434"
)

for name in "${services[@]}"; do
    if stop_service "$name"; then
        for p in ${service_ports[$name]}; do
            sweep_ports+=("$p")
        done
    else
        echo "  $name not started by this script; leaving any listener on port(s) ${service_ports[$name]} intact"
    fi
done

if [ "${#sweep_ports[@]}" -gt 0 ]; then
    for p in "${sweep_ports[@]}"; do
        sweep_port "$p"
    done
fi

echo -e "${RED}모든 서비스가 종료되었습니다.${NC}"
