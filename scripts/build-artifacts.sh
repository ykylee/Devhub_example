#!/usr/bin/env bash
# scripts/build-artifacts.sh — host build pattern (docker network 의존성 없음).
#
# 본 repo 는 host 에서 모든 build artifact 산출 → docker build 는 COPY-only 로
# image 생성 (multi-stage 없음). 사내 환경의 docker container 안에서 외부 proxy
# 접근 불가 (host proxy 가 container 안으로 전파 안 됨) 시나리오 회피.
#
# Prerequisites (host tools, 사전 검증):
#   - go 1.25+              backend-core 빌드 (CGO_ENABLED=0 정적 binary).
#                          backend-core/go.mod 가 `go 1.25.9` 명시 — host go 1.25
#                          정확히 (또는 그 이상) 권장. 사내 proxy 환경은 GOTOOLCHAIN
#                          자동 bump 차단 가능 → host 1.25 사전 설치.
#   - python3.12 (정확히)   backend-ai 빌드 (backend-ai/Dockerfile 의 python:3.12-slim
#                          과 ABI 정합 강제 — 다른 minor ver 시 grpcio 등 C extension
#                          import 실패 / segfault risk)
#   - node 20+ + npm        frontend 빌드 (Next.js standalone)
#   - bash 4+               array / [[ syntax
#
# 산출물 (host fs):
#   backend-core/bin/main                       — Go 정적 binary
#   backend-ai/.build/site-packages/            — Python deps target dir
#   frontend/.next/standalone/                  — Next.js standalone server
#   frontend/.next/static/                      — Next.js static assets
#
# 환경변수 (optional override):
#   ENV_FILE                    deploy env 파일 path. set 시 자동 source.
#   BACKEND_API_URL             (default http://backend-core:8080) frontend build-time
#   NEXT_OUTPUT                 (default standalone) Next.js output mode
#   NEXT_PUBLIC_BASE_PATH       (default devhub) Next.js basePath
#   NEXT_PUBLIC_APP_ORIGIN      (default http://localhost:3000) build-time origin
#   NEXT_PUBLIC_OIDC_*          (default empty) OIDC config — runtime-config route fallback
#
# 사내 환경 proxy: host 의 HTTP_PROXY/HTTPS_PROXY/NO_PROXY 가 go/npm/pip 모두에
# 자연 전파됨 (set -a 도 아닌 user env 가 child process 로 inherit). docker
# container 안에서는 proxy 전파 안 됨 → 본 script 가 host build pattern 채택.
#
# Idempotent: N회 호출해도 동일 결과. `go build` / `npm ci` / `pip install` 모두
# 산출물 deterministic (lock file 또는 go.mod 기반).

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -n "${ENV_FILE:-}" ] && [ -f "${ENV_FILE}" ]; then
  # shellcheck disable=SC1090
  set -a
  source "$ENV_FILE"
  set +a
fi

: "${BACKEND_API_URL:=http://backend-core:8080}"
: "${NEXT_OUTPUT:=standalone}"
: "${NEXT_PUBLIC_BASE_PATH:=devhub}"
: "${NEXT_PUBLIC_APP_ORIGIN:=http://localhost:3000}"
: "${NEXT_PUBLIC_OIDC_ISSUER_URL:=}"
: "${NEXT_PUBLIC_OIDC_REDIRECT_URI:=}"
: "${NEXT_PUBLIC_OIDC_CLIENT_ID:=devhub-frontend}"
: "${NEXT_PUBLIC_OIDC_SCOPE:=openid offline_access email profile}"

# === Prerequisites verification (host tools) ===
verify_prerequisites() {
  local missing=()
  local detail=""

  if ! command -v go >/dev/null 2>&1; then
    missing+=("go")
    detail+="\n  - go 미설치. https://go.dev/dl/ 에서 1.25+ 설치 (backend-core/go.mod = 1.25.9)"
  fi

  if ! command -v node >/dev/null 2>&1; then
    missing+=("node")
    detail+="\n  - node 미설치. https://nodejs.org/ 에서 20.x 설치"
  fi

  if ! command -v npm >/dev/null 2>&1; then
    missing+=("npm")
    detail+="\n  - npm 미설치 (보통 node 와 함께 설치됨)"
  fi

  # python3.12 정확히 강제 — backend-ai/Dockerfile 의 python:3.12-slim 과 ABI 정합.
  # 다른 minor ver 사용 시 grpcio / pydantic-core 등 C extension 의 wheel 이 다른
  # ABI 로 빌드되어 컨테이너 안에서 import 실패 / segfault risk.
  local python_bin=""
  if command -v python3.12 >/dev/null 2>&1; then
    python_bin="python3.12"
  elif command -v python3 >/dev/null 2>&1 && python3 -c 'import sys; raise SystemExit(0 if sys.version_info[:2] == (3, 12) else 1)' 2>/dev/null; then
    python_bin="python3"
  else
    missing+=("python3.12")
    detail+="\n  - python3.12 미설치 (다른 minor ver 은 backend-ai/Dockerfile 의 python:3.12-slim 과 ABI mismatch — C extension import 실패 risk)."
    detail+="\n    설치: pyenv install 3.12 / OS 패키지 매니저 (apt install python3.12 / brew install python@3.12)"
    detail+="\n    검증: python3.12 --version 또는 python3 --version (3.12.x 출력)"
  fi
  PYTHON_BIN="$python_bin"

  if [ ${#missing[@]} -gt 0 ]; then
    echo "ERROR: 필수 host 도구 ${#missing[@]} 개 누락: ${missing[*]}" >&2
    echo -e "상세 안내:$detail" >&2
    echo "" >&2
    echo "전체 prerequisites: 본 script 헤더 또는 docs/setup/docker-packaging-deployment-guide.md §1.1 참조" >&2
    exit 1
  fi

  # 버전 표시 (debug + 운영자가 환경 확인용)
  echo "[prereqs] host build tools verified:"
  echo "  - go      : $(go version | awk '{print $3}')"
  echo "  - node    : $(node --version)"
  echo "  - npm     : $(npm --version)"
  echo "  - python  : $($PYTHON_BIN --version) ($PYTHON_BIN)"
}

# === Build steps ===
build_backend_core() {
  echo "[host build] backend-core (Go)"
  (
    cd "$ROOT_DIR/backend-core"
    mkdir -p bin
    # GOOS=linux is required because the binary is always packaged into a
    # Linux Docker image (alpine). Without explicit GOOS, macOS hosts
    # produce a darwin/arm64 (Mach-O) binary that causes "exec format
    # error" in the container. GOARCH is intentionally left default
    # (matches the Docker host architecture via the Go toolchain
    # default).
    GOOS=linux CGO_ENABLED=0 go build -o bin/main .
  )
}

build_backend_ai() {
  echo "[host build] backend-ai (Python 3.12 deps)"
  # host python 3.12 강제. dockerized fallback 제거 (사내 docker container 안
  # PyPI 접근 proxy 전파 안 됨 → 항상 host 에서만 install).
  mkdir -p "$ROOT_DIR/backend-ai/.build/site-packages"
  (
    cd "$ROOT_DIR/backend-ai"
    "$PYTHON_BIN" -m pip install --upgrade --disable-pip-version-check --no-cache-dir --target .build/site-packages -r requirements.txt
  )
}

build_frontend() {
  echo "[host build] frontend (Next.js standalone)"
  (
    cd "$ROOT_DIR/frontend"
    npm ci
    BACKEND_API_URL="$BACKEND_API_URL" \
      NEXT_OUTPUT="$NEXT_OUTPUT" \
      NEXT_PUBLIC_BASE_PATH="$NEXT_PUBLIC_BASE_PATH" \
      NEXT_PUBLIC_APP_ORIGIN="$NEXT_PUBLIC_APP_ORIGIN" \
      NEXT_PUBLIC_OIDC_ISSUER_URL="$NEXT_PUBLIC_OIDC_ISSUER_URL" \
      NEXT_PUBLIC_OIDC_REDIRECT_URI="$NEXT_PUBLIC_OIDC_REDIRECT_URI" \
      NEXT_PUBLIC_OIDC_CLIENT_ID="$NEXT_PUBLIC_OIDC_CLIENT_ID" \
      NEXT_PUBLIC_OIDC_SCOPE="$NEXT_PUBLIC_OIDC_SCOPE" \
      npm run build
  )
}

main() {
  verify_prerequisites
  build_backend_core
  build_backend_ai
  build_frontend
  echo ""
  echo "[done] all host artifacts built:"
  echo "  - $ROOT_DIR/backend-core/bin/main"
  echo "  - $ROOT_DIR/backend-ai/.build/site-packages/"
  echo "  - $ROOT_DIR/frontend/.next/standalone/"
  echo "  - $ROOT_DIR/frontend/.next/static/"
  echo ""
  echo "Next: docker build (Dockerfile 은 COPY-only + ARG base image, base image pull 외 network 의존 없음)"
  echo "      docker build -f backend-core/Dockerfile -t devhub/backend-core:\$IMAGE_TAG backend-core"
  echo "      docker build -f backend-ai/Dockerfile   -t devhub/backend-ai:\$IMAGE_TAG   backend-ai"
  echo "      docker build -f frontend/Dockerfile     -t devhub/frontend:\$IMAGE_TAG     frontend"
  echo ""
  echo "사내 mirror registry 사용 시 (base image pull 차단 우회):"
  echo "      docker build --build-arg BACKEND_CORE_BASE=internal-registry.example.com/alpine:3.21 ..."
  echo "      docker build --build-arg BACKEND_AI_BASE=internal-registry.example.com/python:3.12-slim ..."
  echo "      docker build --build-arg FRONTEND_BASE=internal-registry.example.com/node:20-alpine ..."
  echo "      자세한 절차: docs/setup/docker-packaging-deployment-guide.md §5.1"
}

main "$@"
