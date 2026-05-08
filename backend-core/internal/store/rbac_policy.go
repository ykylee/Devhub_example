package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetActiveRBACPolicy(ctx context.Context) (domain.RBACPolicy, error) {
	policy := domain.DefaultRBACPolicy()

	var versionID int64
	err := s.pool.QueryRow(ctx, `
SELECT id, policy_version
FROM rbac_policy_versions
WHERE status = 'active'
ORDER BY created_at DESC, id DESC
LIMIT 1`).Scan(&versionID, &policy.PolicyVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RBACPolicy{}, ErrNotFound
	}
	if err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("get active rbac policy version: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
SELECT role, resource, permission
FROM rbac_policy_rules
WHERE policy_version_id = $1
ORDER BY role ASC, resource ASC`, versionID)
	if err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("list rbac policy rules: %w", err)
	}
	defer rows.Close()

	matrix := map[string]map[string]domain.RBACPermission{}
	for rows.Next() {
		var role, resource, permission string
		if err := rows.Scan(&role, &resource, &permission); err != nil {
			return domain.RBACPolicy{}, fmt.Errorf("scan rbac policy rule: %w", err)
		}
		if matrix[role] == nil {
			matrix[role] = map[string]domain.RBACPermission{}
		}
		matrix[role][resource] = domain.RBACPermission(permission)
	}
	if err := rows.Err(); err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("iterate rbac policy rules: %w", err)
	}

	policy.Source = "postgres"
	policy.Editable = true
	policy.Matrix = matrix
	return policy, nil
}

func (s *PostgresStore) ReplaceRBACPolicy(ctx context.Context, input domain.ReplaceRBACPolicyInput) (domain.RBACPolicy, error) {
	policy := input.Policy
	if policy.PolicyVersion == "" {
		generated, err := randomPrefixedID("rbac")
		if err != nil {
			return domain.RBACPolicy{}, fmt.Errorf("generate rbac policy version: %w", err)
		}
		policy.PolicyVersion = generated
	}

	actor := strings.TrimSpace(input.ActorLogin)
	if actor == "" {
		actor = "system"
	}
	reason := strings.TrimSpace(input.Reason)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("begin rbac policy replace: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE rbac_policy_versions SET status = 'archived' WHERE status = 'active'`); err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("archive active rbac policy: %w", err)
	}

	var versionID int64
	if err := tx.QueryRow(ctx, `
INSERT INTO rbac_policy_versions (policy_version, status, actor_login, reason)
VALUES ($1, 'active', $2, $3)
RETURNING id`, policy.PolicyVersion, actor, reason).Scan(&versionID); err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("insert rbac policy version: %w", err)
	}

	for role, resources := range policy.Matrix {
		for resource, permission := range resources {
			if _, err := tx.Exec(ctx, `
INSERT INTO rbac_policy_rules (policy_version_id, role, resource, permission)
VALUES ($1, $2, $3, $4)`, versionID, role, resource, string(permission)); err != nil {
				return domain.RBACPolicy{}, fmt.Errorf("insert rbac policy rule %s/%s: %w", role, resource, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.RBACPolicy{}, fmt.Errorf("commit rbac policy replace: %w", err)
	}

	policy.Source = "postgres"
	policy.Editable = true
	return policy, nil
}
