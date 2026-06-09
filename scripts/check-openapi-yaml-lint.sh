#!/bin/bash
# scripts/check-openapi-yaml-lint.sh
#
# ADR-0029 §6 (d) P2 — openapi.yaml CI lint gate.
# 4 항목 검증:
#   1. yaml safe_load (OpenAPI 3.0.3 spec validity)
#   2. info.version semver format (X.Y.Z)
#   3. paths count >= 1 (openapi.yaml 비어있으면 fail)
#   4. docs/backend_api_contract.md ↔ docs/openapi.yaml cross-link
#      (contract.md 가 openapi.yaml link 1건 이상 인용 — ADR-0027 §3 결정)
#
# Pattern: scripts/check-migration-uniqueness.sh 와 정합 (별도 bash script
# + set -euo pipefail + ::error:: / ::warning:: GitHub Actions workflow
# command). 운영자가 로컬에서도 실행 가능.
set -euo pipefail

OPENAPI_YAML="backend-core/internal/httpapi/swaggerui/asset/openapi.yaml"
CONTRACT_MD="docs/backend_api_contract.md"

echo "🔍 Checking openapi.yaml lint ($OPENAPI_YAML)..."

if [ ! -f "$OPENAPI_YAML" ]; then
  echo "::error::❌ Error: openapi.yaml '$OPENAPI_YAML' not found!"
  exit 1
fi

if [ ! -f "$CONTRACT_MD" ]; then
  echo "::error::❌ Error: backend API contract '$CONTRACT_MD' not found!"
  exit 1
fi

errors=0

# 1. yaml safe_load (OpenAPI 3.0.3 spec validity)
echo "  1. yaml safe_load + openapi version check..."
if ! python3 -c "
import yaml, sys
with open('$OPENAPI_YAML') as f:
    spec = yaml.safe_load(f)
if not isinstance(spec, dict):
    print('::error::❌ openapi.yaml is not a valid YAML mapping (got ' + type(spec).__name__ + ')')
    sys.exit(1)
oa = spec.get('openapi', '')
if not oa.startswith('3.'):
    print('::error::❌ openapi.yaml openapi version is not 3.x (got ' + str(oa) + ')')
    sys.exit(1)
info = spec.get('info', {})
if 'title' not in info or 'version' not in info:
    print('::error::❌ openapi.yaml info.title or info.version missing')
    sys.exit(1)
paths = spec.get('paths', {})
if not paths:
    print('::error::❌ openapi.yaml paths is empty')
    sys.exit(1)
schemas = spec.get('components', {}).get('schemas', {})
if not schemas:
    print('::error::❌ openapi.yaml components.schemas is empty')
    sys.exit(1)
securities = spec.get('components', {}).get('securitySchemes', {})
if 'bearerAuth' not in securities:
    print('::error::❌ openapi.yaml components.securitySchemes.bearerAuth missing (ADR-0019 + ADR-0029 정합)')
    sys.exit(1)
print('    ok: openapi ' + oa + ' | ' + info['title'] + ' v' + info['version'] + ' | paths=' + str(len(paths)) + ' | schemas=' + str(len(schemas)))
"; then
  errors=$((errors + 1))
fi

# 2. info.version semver format
echo "  2. info.version semver check (X.Y.Z)..."
VERSION=$(python3 -c "
import yaml
with open('$OPENAPI_YAML') as f:
    spec = yaml.safe_load(f)
print(spec.get('info', {}).get('version', ''))
")
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
  echo "::error::❌ openapi.yaml info.version '$VERSION' does not match semver (X.Y.Z[-suffix])"
  errors=$((errors + 1))
else
  echo "    ok: info.version=$VERSION"
fi

# 3. paths count >= 1
echo "  3. paths count check (>= 1)..."
PATH_COUNT=$(python3 -c "
import yaml
with open('$OPENAPI_YAML') as f:
    spec = yaml.safe_load(f)
print(len(spec.get('paths', {})))
")
if [ "$PATH_COUNT" -lt 1 ]; then
  echo "::error::❌ openapi.yaml paths count is $PATH_COUNT (expected >= 1)"
  errors=$((errors + 1))
else
  echo "    ok: paths count=$PATH_COUNT"
fi

# 4. contract.md ↔ openapi.yaml cross-link
echo "  4. contract.md ↔ openapi.yaml cross-link check..."
if ! grep -q "openapi\.yaml" "$CONTRACT_MD"; then
  echo "::error::❌ $CONTRACT_MD does not reference openapi.yaml (cross-link missing — ADR-0027 §3 결정)"
  errors=$((errors + 1))
else
  CROSS_LINK_COUNT=$(grep -c "openapi\.yaml" "$CONTRACT_MD" || true)
  echo "    ok: contract.md has $CROSS_LINK_COUNT reference(s) to openapi.yaml"
fi

echo ""
if [ "$errors" -ne 0 ]; then
  echo "::error::❌ openapi.yaml lint failed: $errors error(s)"
  exit 1
fi

echo "✅ openapi.yaml lint passed: yaml valid + semver + paths>=$PATH_COUNT + cross-link ok"
