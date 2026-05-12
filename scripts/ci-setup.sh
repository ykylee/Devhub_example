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
if ! command -v kratos &> /dev/null || ! command -v hydra &> /dev/null; then
  echo "Installing Ory binaries..."
  curl -L https://github.com/ory/kratos/releases/download/v1.1.0/kratos_1.1.0-linux_64bit.tar.gz | tar xz -C /usr/local/bin kratos
  curl -L https://github.com/ory/hydra/releases/download/v2.2.0/hydra_2.2.0-linux_64bit.tar.gz | tar xz -C /usr/local/bin hydra
fi

# 2. DB Schema 생성
echo "Applying IdP schemas..."
cd backend-core
go run ./cmd/idp-apply-schemas -sql ../infra/idp/sql/001_create_idp_schemas.sql
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
echo "Seeding E2E users..."
cd backend-core
go run ./cmd/idp-apply-schemas -sql ../infra/idp/sql/002_seed_e2e_users.sql
cd ..

echo "CI setup completed successfully."
