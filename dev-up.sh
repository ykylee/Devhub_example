#!/bin/bash

# 색상 정의
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Restart 처리
if [ "$1" == "restart" ]; then
    echo -e "${YELLOW}🔄 서비스를 재시작합니다... (dev-down 실행 중)${NC}"
    ./dev-down.sh
    sleep 2
fi

echo -e "${BLUE}🚀 DevHub 로컬 통합 서비스를 시작합니다...${NC}"

# PID 파일 저장 경로
PID_DIR=".pids"
mkdir -p "$PID_DIR"

# 서비스 실행 함수
run_service() {
    local name=$1
    local cmd=$2
    local log_file=$3
    local working_dir=$4
    local pid_file="$PID_DIR/.$name.pid"

    echo -e "${GREEN}Starting $name in ${working_dir:-root}...${NC}"
    if [ -n "$working_dir" ]; then
        (cd "$working_dir" && $cmd) > "$log_file" 2>&1 &
    else
        $cmd > "$log_file" 2>&1 &
    fi
    echo $! > "$pid_file"
}

# 1. IDP 서비스 (Ory Kratos & Hydra)
# 루트에서 실행하여 infra/idp/ 경로를 유효하게 함
if command -v kratos &> /dev/null; then
    run_service "kratos" "kratos serve -c infra/idp/kratos.yaml --dev" "kratos.log" ""
else
    echo -e "${YELLOW}⚠️ kratos 명령어를 찾을 수 없습니다.${NC}"
fi

if command -v hydra &> /dev/null; then
    run_service "hydra" "hydra serve all -c infra/idp/hydra.yaml --dev" "hydra.log" ""
else
    echo -e "${YELLOW}⚠️ hydra 명령어를 찾을 수 없습니다.${NC}"
fi

# 2. 백엔드 실행
# DB_URL 은 머신마다 사용자/포트가 다르므로 환경에서 받음. 미설정 시
# `$USER@localhost` 기본값을 채워 macOS/Linux 의 trust 인증을 그대로
# 활용한다 (그래도 사용자 머신에 맞게 .env 또는 export 로 재정의 권장).
export AUTH_DEV_FALLBACK=true
export DB_URL="${DB_URL:-postgres://${USER}@localhost:5432/devhub?sslmode=disable}"
run_service "backend" "go run main.go" "backend.log" "backend-core"

# 3. 프론트엔드 실행
run_service "frontend" "npm run dev" "frontend.log" "frontend"

echo -e "${BLUE}✅ 서비스가 백그라운드에서 실행 중입니다.${NC}"
