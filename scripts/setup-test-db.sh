#!/bin/bash
# scripts/setup-test-db.sh — backend integration test 의 PostgreSQL 환경 준비.
#
# 사용법:
#   export DEVHUB_TEST_DB_URL="postgres://runner@localhost:5432/devhub_test?sslmode=disable"
#   ./scripts/setup-test-db.sh
#   cd backend-core && go test -count=1 -run 'TestIntegration_' ./internal/store/...
#
# 동작:
#   1. DEVHUB_TEST_DB_URL 환경변수 검증
#   2. migrate (golang-migrate) 가 설치되어 있지 않으면 go install 로 설치
#   3. backend-core/migrations 의 모든 up.sql 을 DB 에 적용
#
# 참고:
#   - 본 스크립트는 새 DB 를 만들지 않는다. 사용자가 미리 createdb 또는 DSN 의
#     DB 가 비어있어야 한다.
#   - integration test 의 applicationsFixture 가 TRUNCATE CASCADE 로 매 test 마다
#     상태 격리. 본 스크립트는 1회 실행 (마이그레이션 적용) 으로 충분.
#   - CI 에서는 .github/workflows/ci.yml 의 backend-integration job 이 동일 로직을
#     수행 (scripts/ci-setup.sh 의 migrate up step 재사용).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if [[ -z "${DEVHUB_TEST_DB_URL:-}" ]]; then
  echo "ERROR: DEVHUB_TEST_DB_URL is not set."
  echo "  Example: export DEVHUB_TEST_DB_URL='postgres://runner@localhost:5432/devhub_test?sslmode=disable'"
  exit 1
fi

if ! command -v migrate &> /dev/null; then
  echo "golang-migrate not found; installing..."
  GOBIN=/usr/local/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
fi

echo "Applying migrations to $DEVHUB_TEST_DB_URL ..."
migrate -path backend-core/migrations -database "$DEVHUB_TEST_DB_URL" up

echo "Done. Run integration tests with:"
echo "  cd backend-core && go test -count=1 -run 'TestIntegration_' ./internal/store/..."
