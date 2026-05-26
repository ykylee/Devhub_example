#!/usr/bin/env bash
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

build_backend_core() {
  echo "[host build] backend-core"
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
  echo "[host build] backend-ai"
  # host python3 와 backend-ai/Dockerfile (python:3.12-slim) 의 메이저/마이너가
  # 일치해야 한다. mismatch 시 grpcio 등 컴파일 확장 모듈의 ABI 가 컨테이너에서
  # import 실패하거나 segfault 한다.
  mkdir -p "$ROOT_DIR/backend-ai/.build/site-packages"
  if command -v python3.12 >/dev/null 2>&1; then
    (
      cd "$ROOT_DIR/backend-ai"
      python3.12 -m pip install --upgrade --disable-pip-version-check --no-cache-dir --target .build/site-packages -r requirements.txt
    )
    return
  fi

  if python3 -c 'import sys; raise SystemExit(0 if sys.version_info[:2] == (3, 12) else 1)' 2>/dev/null; then
    (
      cd "$ROOT_DIR/backend-ai"
      python3 -m pip install --upgrade --disable-pip-version-check --no-cache-dir --target .build/site-packages -r requirements.txt
    )
    return
  fi

  echo "WARN: host python is not 3.12; fallback to dockerized python:3.12-slim for backend-ai deps"
  docker run --rm \
    -v "$ROOT_DIR/backend-ai":/work \
    -w /work \
    python:3.12-slim \
    bash -lc "python -m pip install --upgrade --disable-pip-version-check --no-cache-dir --target .build/site-packages -r requirements.txt"
}

build_frontend() {
  echo "[host build] frontend"
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
  build_backend_core
  build_backend_ai
  build_frontend
}

main "$@"
