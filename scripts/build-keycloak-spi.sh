#!/usr/bin/env bash
# build-keycloak-spi.sh — P2-6 (Keycloak SPI provider JAR) + X-8 (P3-5 audit event
# listener push 전환) 의 JAR build + push (staging/prod 사내 SCM hand-off).
# docs/setup/keycloak_event_listener_spi_staging.md §2 의 단계 2 정공법.
# Option A: 표준 mvn build + 산출물 확인.
# Option B: docker buildx build + 도커 이미지 push (harbor.internal 등 사내 registry).
# exit 0 = build success + push success, exit 1 = build/push fail.

set -euo pipefail

# === env ===
KEYCLOAK_SPI_VERSION="${KEYCLOAK_SPI_VERSION:-1.0.0}"
KEYCLOAK_SPI_REGISTRY="${KEYCLOAK_SPI_REGISTRY:-harbor.internal/devhub}"  # 사내 registry
PUSH_MODE="${PUSH_MODE:-docker}"  # docker (default) | mvn (maven build only, no push)
MIRROR_BASE_URL="${MIRROR_BASE_URL:-}"  # 선택: 사내 maven mirror (e.g. https://maven.internal/repository/maven-public)

PASS=0
FAIL=0
note() { printf "    %s\n" "$*"; }
pass() { printf "  [PASS] %s\n" "$*"; PASS=$((PASS+1)); }
fail() { printf "  [FAIL] %s\n" "$*"; FAIL=$((FAIL+1)); }

echo "==> Building Keycloak Event Listener SPI JAR"
echo "    Version: $KEYCLOAK_SPI_VERSION"
echo "    Registry: $KEYCLOAK_SPI_REGISTRY"
echo "    Push mode: $PUSH_MODE"

# === prerequisite: Java + Maven ===
echo "==> Prerequisite: Java 21 + Maven 3.13+"
if ! command -v java >/dev/null 2>&1; then
  fail "java not found. Install Java 21+ (e.g. temurin-21, openjdk-21)"
  exit 1
fi
JAVA_VERSION=$(java -version 2>&1 | head -1 | awk -F\" '{print $2}')
if [[ ! "$JAVA_VERSION" =~ ^2[1-9](\.|$) ]]; then
  fail "java version = $JAVA_VERSION (expected 21+)"
  exit 1
fi
pass "java = $JAVA_VERSION"

if ! command -v mvn >/dev/null 2>&1; then
  fail "mvn not found. Install Maven 3.13+"
  exit 1
fi
MVN_VERSION=$(mvn -version 2>&1 | head -1 | awk '{print $3}')
if [[ ! "$MVN_VERSION" =~ ^3\.1[3-9]\. ]]; then
  note "mvn = $MVN_VERSION (expected 3.13+, continuing — best-effort)"
fi
pass "mvn = $MVN_VERSION"

# === maven build (사내 maven mirror 적용 가능) ===
SPI_DIR="infra/idp/keycloak-event-listener-spi"
JAR_PATH="$SPI_DIR/target/devhub-keycloak-event-listener.jar"

echo "==> [1/3] Maven build: $SPI_DIR"
if [ ! -d "$SPI_DIR" ]; then
  fail "SPI source dir not found: $SPI_DIR"
  exit 1
fi

cd "$SPI_DIR"
MVN_ARGS=()
if [ -n "$MIRROR_BASE_URL" ]; then
  MVN_ARGS+=("-Dmaven.repo.remote=$MIRROR_BASE_URL")
fi
mvn clean package "${MVN_ARGS[@]}" -DskipTests
cd - >/dev/null

if [ ! -f "$JAR_PATH" ]; then
  fail "JAR build failed: $JAR_PATH not found"
  exit 1
fi
JAR_SIZE=$(stat -c%s "$JAR_PATH" 2>/dev/null || stat -f%z "$JAR_PATH")
pass "JAR built: $JAR_PATH ($JAR_SIZE bytes)"

# === [2/3] JAR 검증 (META-INF/services 등록 확인) ===
echo "==> [2/3] JAR META-INF/services verification"
SPI_FACTORY=$(unzip -p "$JAR_PATH" META-INF/services/org.keycloak.events.EventListenerProviderFactory 2>/dev/null || true)
if [ -z "$SPI_FACTORY" ]; then
  fail "META-INF/services/org.keycloak.events.EventListenerProviderFactory missing or empty"
  exit 1
fi
if echo "$SPI_FACTORY" | grep -q "com.devhub.keycloak.spi.DevHubEventListenerProviderFactory"; then
  pass "META-INF/services registers DevHubEventListenerProviderFactory"
else
  fail "META-INF/services does not register DevHubEventListenerProviderFactory (got: $SPI_FACTORY)"
  exit 1
fi

# === [3/3] Push (docker OR mvn deploy) ===
echo "==> [3/3] Push: mode=$PUSH_MODE"
if [ "$PUSH_MODE" = "mvn" ]; then
  note "Maven deploy to registry: $KEYCLOAK_SPI_REGISTRY (사용자 결정 — settings.xml 의 <distributionManagement> 정합 확인 필요)"
  cd "$SPI_DIR"
  mvn deploy -DskipTests "${MVN_ARGS[@]}" || {
    fail "mvn deploy failed"
    exit 1
  }
  cd - >/dev/null
  pass "Maven deploy to registry: $KEYCLOAK_SPI_REGISTRY"
elif [ "$PUSH_MODE" = "docker" ]; then
  # docker buildx build + push (Dockerfile.keycloak 의 multi-stage build 활용)
  # Dockerfile.keycloak 는 infra/idp/Dockerfile.keycloak 에 위치
  DOCKERFILE="infra/idp/Dockerfile.keycloak"
  IMAGE_NAME="$KEYCLOAK_SPI_REGISTRY/devhub-keycloak-spi:$KEYCLOAK_SPI_VERSION"

  if [ ! -f "$DOCKERFILE" ]; then
    fail "Dockerfile not found: $DOCKERFILE"
    exit 1
  fi

  if ! command -v docker >/dev/null 2>&1; then
    fail "docker not found. Install Docker 20.10+ with buildx"
    exit 1
  fi

  # docker buildx build + push
  if docker buildx version >/dev/null 2>&1; then
    docker buildx build \
      --file "$DOCKERFILE" \
      --tag "$IMAGE_NAME" \
      --platform linux/amd64 \
      --push \
      . || {
      fail "docker buildx build + push failed"
      exit 1
    }
    pass "Docker image built + pushed: $IMAGE_NAME"
  else
    # fallback: docker build + docker push
    docker build --file "$DOCKERFILE" --tag "$IMAGE_NAME" . || {
      fail "docker build failed"
      exit 1
    }
    docker push "$IMAGE_NAME" || {
      fail "docker push failed"
      exit 1
    }
    pass "Docker image built + pushed: $IMAGE_NAME"
  fi
else
  fail "Unknown PUSH_MODE: $PUSH_MODE (expected docker | mvn)"
  exit 1
fi

# === Summary ===
echo "==> Summary"
echo "    PASS: $PASS"
echo "    FAIL: $FAIL"

if [ "$FAIL" -eq 0 ]; then
  echo "==> ✅ Build success — staging hand-off ready (다음 단계: §3 compose mount + §4 verify 자동 실행)"
  echo "    Artifacts:"
  echo "    - JAR: $JAR_PATH"
  echo "    - Image: $KEYCLOAK_SPI_REGISTRY/devhub-keycloak-spi:$KEYCLOAK_SPI_VERSION"
  exit 0
else
  echo "==> ❌ $FAIL step(s) failed — see detail above"
  exit 1
fi
