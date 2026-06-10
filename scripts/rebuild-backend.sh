#!/usr/bin/env bash
# rebuild-backend.sh — backend-core 정적 binary 재빌드 + docker image 재빌드 + container 재기동.
#
# 사용법:
#   ./scripts/rebuild-backend.sh                    # 전체 (build + docker rebuild + container recreate)
#   ./scripts/rebuild-backend.sh --binary-only      # bin/main 만 재빌드 (docker 미반영)
#   ./scripts/rebuild-backend.sh --compose-file PATH  # 다른 compose 파일 사용 (default: docker-compose.yml)
#
# 환경:
#   CGO_ENABLED=0 + -ldflags "-s -w -extldflags '-static'" 로 alpine-compatible 정적 binary build.
#   Dockerfile 의 `COPY bin/main ./main` 가 이 binary 를 사용하므로, Go source 변경 후 docker image
#   가 stale binary 를 사용하지 않도록 본 script 가 build → docker build → container recreate 순서로
#   강제. backend 만 변경 시 frontend 재빌드는 skip.
#
# 의존성:
#   - Go 1.25.x (backend-core/go.mod 의 `go 1.25.9` 또는 그 이상)
#   - Docker 20+ + docker compose v2
#   - curl (health check)

set -euo pipefail

BINARY_ONLY=false
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --binary-only) BINARY_ONLY=true; shift ;;
    --compose-file) COMPOSE_FILE="$2"; shift 2 ;;
    -h|--help)
      sed -n '3,18p' "$0"
      exit 0
      ;;
    *) echo "unknown flag: $1" >&2; exit 1 ;;
  esac
done

# 1. Static Go binary build
echo "==> [1/3] building backend-core (CGO_ENABLED=0, alpine-compatible static)"
cd "$(dirname "$0")/.."
CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o backend-core/bin/main ./backend-core

if ! file backend-core/bin/main | grep -q "statically linked"; then
  echo "ERROR: built binary is NOT statically linked. CGO_ENABLED=0 or ldflags likely wrong." >&2
  exit 1
fi
BINARY_SIZE=$(stat -c%s backend-core/bin/main 2>/dev/null || stat -f%z backend-core/bin/main)
echo "    bin/main: ${BINARY_SIZE} bytes (statically linked)"

if [[ "$BINARY_ONLY" == "true" ]]; then
  echo "==> --binary-only: skipping docker rebuild"
  exit 0
fi

# 2. Docker image rebuild (--no-cache 강제 — Dockerfile 의 COPY bin/main 가 stale 안 되도록)
echo "==> [2/3] rebuilding docker image (no-cache)"
docker compose -f "$COMPOSE_FILE" build --no-cache backend-core

# 3. Container recreate (새 image 로 교체)
echo "==> [3/3] recreating backend-core container"
docker compose -f "$COMPOSE_FILE" up -d backend-core

# 4. Health check (최대 30s 대기)
echo "==> waiting for backend-core health (up to 30s)"
for i in {1..30}; do
  if curl -sS -f -m 2 http://localhost:8080/health >/dev/null 2>&1; then
    echo "    backend-core is healthy (after ${i}s)"
    exit 0
  fi
  sleep 1
done
echo "ERROR: backend-core did not become healthy within 30s" >&2
echo "check logs: docker compose -f $COMPOSE_FILE logs --tail=50 backend-core" >&2
exit 1
