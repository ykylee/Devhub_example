#!/bin/bash
#
# scripts/ci-setup.sh — GitHub Actions 전용 IdP 및 DB 초기화 스크립트.
#

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# 환경변수 기본값 (CI 전용)
export DB_URL="${DB_URL:-postgres://runner@localhost:5432/devhub?sslmode=disable}"
export HYDRA_ADMIN_URL="${HYDRA_ADMIN_URL:-http://localhost:4445}"
export KRATOS_ADMIN_URL="${KRATOS_ADMIN_URL:-http://localhost:4434}"

echo "Preparing DevHub CI environment..."

# 1. Ory 바이너리 설치 (Linux 64bit 가정)
# Pin to v26.2.0 to match infra/idp/scripts/install-binaries.ps1 (prod /
# dev-up.sh canonical version). The older v2.2.0/v1.1.0 line predates the
# `migrate sql up` subcommand convention used by dev-up.sh and trips with
# "Please provide the database URL." even when DSN is supplied.
#
# The v26.x linux_64bit tarballs nest the binary one level deep (matching
# the Windows zip layout that install-binaries.ps1 handles via -Recurse).
# Extract into a temp dir and `find` the binary so this keeps working if
# upstream layout shifts again.
ORY_VERSION="${ORY_VERSION:-26.2.0}"
install_ory_binary() {
  local name="$1"
  local url="https://github.com/ory/${name}/releases/download/v${ORY_VERSION}/${name}_${ORY_VERSION}-linux_64bit.tar.gz"
  local tmpdir
  tmpdir="$(mktemp -d)"
  curl -fsSL "$url" -o "$tmpdir/${name}.tar.gz"
  tar xzf "$tmpdir/${name}.tar.gz" -C "$tmpdir"
  local bin
  bin="$(find "$tmpdir" -type f -name "$name" -perm -u=x | head -n1)"
  if [ -z "$bin" ]; then
    echo "Failed to locate ${name} binary inside ${name}_${ORY_VERSION}-linux_64bit.tar.gz" >&2
    echo "Tarball contents:" >&2
    tar tzf "$tmpdir/${name}.tar.gz" >&2
    exit 1
  fi
  install -m 0755 "$bin" "/usr/local/bin/${name}"
  rm -rf "$tmpdir"
}

if ! command -v kratos &> /dev/null || ! command -v hydra &> /dev/null; then
  echo "Installing Ory binaries (v${ORY_VERSION})..."
  install_ory_binary kratos
  install_ory_binary hydra
fi
kratos version
hydra version

# 2. DB Schema 생성
# `idp-apply-schemas` reads `DEVHUB_DB_URL` (not `DB_URL`) and otherwise falls
# back to a `postgres:postgres@...` default that does not exist on the GH
# runner (POSTGRES_USER=runner). Pass DSN explicitly via the `-dsn` flag so the
# correct credentials are used regardless of whether sudo strips the env.
echo "Applying IdP schemas..."
cd backend-core
go run ./cmd/idp-apply-schemas -dsn "$DB_URL" -sql ../infra/idp/sql/001_create_idp_schemas.sql
cd ..

# 3. Ory 마이그레이션
echo "Running Ory migrations..."
export DSN_HYDRA="${DB_URL}&search_path=hydra"
export DSN_KRATOS="${DB_URL}&search_path=kratos"
hydra migrate sql up --yes "$DSN_HYDRA"
kratos migrate sql up --yes "$DSN_KRATOS"

# 4. App 마이그레이션 (golang-migrate)
echo "Running app migrations..."
if ! command -v migrate &> /dev/null; then
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
fi
migrate -path backend-core/migrations -database "$DB_URL" up

# 5. E2E Seed 데이터 주입
# Same `-dsn` requirement as the schema-init step above.
echo "Seeding E2E users..."
cd backend-core
go run ./cmd/idp-apply-schemas -dsn "$DB_URL" -sql ../infra/idp/sql/002_seed_e2e_users.sql
cd ..

echo "CI setup completed successfully."
