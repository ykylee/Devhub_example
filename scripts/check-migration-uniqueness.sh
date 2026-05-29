#!/bin/bash
# scripts/check-migration-uniqueness.sh
set -euo pipefail

MIGRATIONS_DIR="backend-core/migrations"
echo "🔍 Checking migration versions in $MIGRATIONS_DIR..."

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "❌ Error: Migrations directory '$MIGRATIONS_DIR' not found!"
  exit 1
fi

# 1. 버전 일련번호 규격 린팅 (반드시 6자리 숫자 + 언더바 형태여야 함)
invalid_files=()
for file in "$MIGRATIONS_DIR"/*.sql; do
  [ -e "$file" ] || continue
  filename=$(basename "$file")
  prefix=$(echo "$filename" | cut -d'_' -f1)
  
  # 6자리 숫자가 아니거나 숫자가 아닌 문자가 섞인 경우 검출
  if [[ ! "$prefix" =~ ^[0-9]{6}$ ]]; then
    invalid_files+=("$filename")
  fi
done

if [ ${#invalid_files[@]} -ne 0 ]; then
  echo "::error::❌ Error: Invalid migration filename pattern detected! Must start with a 6-digit number prefix (e.g. 000042_...)."
  for invalid in "${invalid_files[@]}"; do
    echo "  - Invalid file: $invalid"
  done
  exit 1
fi

# 2. 중복 버전 prefix 검증 (up.sql 기준)
duplicates=$(
  find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name '*.up.sql' -print \
    | awk -F'/' '{print $NF}' \
    | awk -F'_' '{print $1}' \
    | sort \
    | uniq -d
)

if [ -n "$duplicates" ]; then
  echo "::error::❌ Error: Duplicate migration versions detected!"
  echo "$duplicates" | while read -r prefix; do
    echo "  - Version prefix '$prefix' is used by multiple migration files:"
    find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name "${prefix}_*.up.sql" -print | sed 's/^/    /'
  done
  echo "::error::See docs/setup/migration_000021_conflict_resolution.md for resolution SOP."
  exit 1
fi

echo "✅ All migration prefixes are valid and unique!"
exit 0
