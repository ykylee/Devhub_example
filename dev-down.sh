#!/bin/bash

# 색상 정의
RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}🛑 DevHub 로컬 서비스를 종료합니다...${NC}"

PID_DIR=".pids"

# 프로세스 종료 함수
stop_service() {
    local name=$1
    local pid_file="$PID_DIR/.$name.pid"
    
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        kill $PID 2>/dev/null
        rm "$pid_file"
        echo "$name (PID: $PID) stopped."
    fi
}

# 1. 프론트엔드 종료
stop_service "frontend"

# 2. 백엔드 종료
stop_service "backend"

# 3. IDP 서비스 종료
stop_service "kratos"
stop_service "hydra"

# 남은 프로세스 정리 (포트 기준 - 안전장치)
lsof -ti:3000,8080,4433,4434,4444,4445 | xargs kill -9 2>/dev/null

echo -e "${RED}✅ 모든 서비스가 종료되었습니다.${NC}"
