package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrSystemRoleImmutable is returned when a caller tries to delete or rename
// one of the seeded system roles. Permissions on system roles can still change
// via UpdateRBACRolePermissions.
var ErrSystemRoleImmutable = errors.New("system role is immutable")

// ErrRoleInUse is returned when DeleteRBACRole is called while at least one
// user still has the role assigned. The handler maps this to 422 role_in_use
// per docs/backend_api_contract.md section 12.5.
var ErrRoleInUse = errors.New("role is still assigned to subjects")

// ErrAuditInvariantViolation is returned when a permission matrix tries to
// grant create/edit/delete on the audit resource (section 12.0.4).
var ErrAuditInvariantViolation = errors.New("audit resource cannot grant create/edit/delete")

// rbacAuditInvariantConstraint matches the CHECK constraint name from
// migrations/000005_create_rbac_policies.up.sql.
const rbacAuditInvariantConstraint = "rbac_policies_audit_invariant"

func isCheckViolation(err error, name string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23514" {
		return false
	}
	if name == "" {
		return true
	}
	return pgErr.ConstraintName == name
}

// ListRBACRoles returns all roles ordered with system roles first (developer,
// manager, system_admin) and custom roles after, both in role_id ascending order.
func (s *PostgresStore) ListRBACRoles(ctx context.Context) ([]domain.RBACRole, error) {
	const query = `
SELECT role_id, name, description, is_system, permissions::text, created_at, updated_at
FROM rbac_policies
ORDER BY
    is_system DESC,
    CASE role_id
        WHEN 'developer'    THEN 0
        WHEN 'manager'      THEN 1
        WHEN 'system_admin' THEN 2
        ELSE 3
    END,
    role_id ASC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query rbac roles: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.RBACRole, 0)
	for rows.Next() {
		role, err := scanRBACRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rbac roles: %w", err)
	}
	return roles, nil
}

// GetRBACRole returns a single role by id. ErrNotFound when no row matches.
func (s *PostgresStore) GetRBACRole(ctx context.Context, roleID string) (domain.RBACRole, error) {
	const query = `
SELECT role_id, name, description, is_system, permissions::text, created_at, updated_at
FROM rbac_policies
WHERE role_id = $1`

	row := s.pool.QueryRow(ctx, query, roleID)
	role, err := scanRBACRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RBACRole{}, fmt.Errorf("rbac role %s: %w", roleID, ErrNotFound)
	}
	if err != nil {
		return domain.RBACRole{}, err
	}
	return role, nil
}

// CreateRBACRole inserts a custom role. The caller must have validated the
// id format with domain.ValidateRoleID. ErrConflict on duplicate id.
// ErrAuditInvariantViolation if permissions grant audit:create/edit/delete.
func (s *PostgresStore) CreateRBACRole(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error) {
	if domain.IsSystemRole(role.ID) {
		return domain.RBACRole{}, fmt.Errorf("create role %s: %w", role.ID, ErrConflict)
	}

	matrix := domain.EnforceAuditInvariant(role.Permissions)
	payload, err := json.Marshal(matrix)
	if err != nil {
		return domain.RBACRole{}, fmt.Errorf("marshal permissions: %w", err)
	}

	const query = `
INSERT INTO rbac_policies (role_id, name, description, is_system, permissions)
VALUES ($1, $2, $3, FALSE, $4::jsonb)
RETURNING role_id, name, description, is_system, permissions::text, created_at, updated_at`

	row := s.pool.QueryRow(ctx, query, role.ID, role.Name, role.Description, payload)
	created, err := scanRBACRole(row)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.RBACRole{}, fmt.Errorf("create role %s: %w", role.ID, ErrConflict)
		}
		if isCheckViolation(err, rbacAuditInvariantConstraint) {
			return domain.RBACRole{}, fmt.Errorf("create role %s: %w", role.ID, ErrAuditInvariantViolation)
		}
		return domain.RBACRole{}, fmt.Errorf("create role %s: %w", role.ID, err)
	}
	return created, nil
}

// UpdateRBACRolePermissions replaces the permission matrix for any role
// (system or custom). Audit invariant is enforced both client-side and by
// the DB CHECK constraint.
func (s *PostgresStore) UpdateRBACRolePermissions(ctx context.Context, roleID string, perms domain.PermissionMatrix) (domain.RBACRole, error) {
	matrix := domain.EnforceAuditInvariant(perms)
	payload, err := json.Marshal(matrix)
	if err != nil {
		return domain.RBACRole{}, fmt.Errorf("marshal permissions: %w", err)
	}

	const query = `
UPDATE rbac_policies
SET permissions = $2::jsonb, updated_at = NOW()
WHERE role_id = $1
RETURNING role_id, name, description, is_system, permissions::text, created_at, updated_at`

	row := s.pool.QueryRow(ctx, query, roleID, payload)
	updated, err := scanRBACRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RBACRole{}, fmt.Errorf("update role %s: %w", roleID, ErrNotFound)
	}
	if err != nil {
		if isCheckViolation(err, rbacAuditInvariantConstraint) {
			return domain.RBACRole{}, fmt.Errorf("update role %s: %w", roleID, ErrAuditInvariantViolation)
		}
		return domain.RBACRole{}, fmt.Errorf("update role %s: %w", roleID, err)
	}
	return updated, nil
}

// UpdateRBACRoleMetadata changes name/description on a custom role. System
// roles are rejected with ErrSystemRoleImmutable.
func (s *PostgresStore) UpdateRBACRoleMetadata(ctx context.Context, roleID, name, description string) (domain.RBACRole, error) {
	if domain.IsSystemRole(roleID) {
		return domain.RBACRole{}, fmt.Errorf("update role %s metadata: %w", roleID, ErrSystemRoleImmutable)
	}

	const query = `
UPDATE rbac_policies
SET name = $2, description = $3, updated_at = NOW()
WHERE role_id = $1 AND is_system = FALSE
RETURNING role_id, name, description, is_system, permissions::text, created_at, updated_at`

	row := s.pool.QueryRow(ctx, query, roleID, name, description)
	updated, err := scanRBACRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RBACRole{}, fmt.Errorf("update role %s metadata: %w", roleID, ErrNotFound)
	}
	if err != nil {
		return domain.RBACRole{}, fmt.Errorf("update role %s metadata: %w", roleID, err)
	}
	return updated, nil
}

// DeleteRBACRole removes a custom role. Returns ErrSystemRoleImmutable for
// system roles, ErrRoleInUse if any user still has the role assigned, and
// ErrNotFound otherwise.
func (s *PostgresStore) DeleteRBACRole(ctx context.Context, roleID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var isSystem bool
	err = tx.QueryRow(ctx, `SELECT is_system FROM rbac_policies WHERE role_id = $1`, roleID).Scan(&isSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("delete role %s: %w", roleID, ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("lookup role %s: %w", roleID, err)
	}
	if isSystem {
		return fmt.Errorf("delete role %s: %w", roleID, ErrSystemRoleImmutable)
	}

	var usageCount int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = $1`, roleID).Scan(&usageCount); err != nil {
		return fmt.Errorf("count role %s usage: %w", roleID, err)
	}
	if usageCount > 0 {
		return fmt.Errorf("delete role %s (%d assigned): %w", roleID, usageCount, ErrRoleInUse)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM rbac_policies WHERE role_id = $1`, roleID); err != nil {
		// COUNT-then-DELETE leaves a window where another transaction can
		// assign the role to a user; the FK to users.role then rejects the
		// delete with 23503. Surface that as ErrRoleInUse so the handler still
		// returns 422 role_in_use rather than a 500 (M1-FIX-A).
		if isForeignKeyViolation(err) {
			return fmt.Errorf("delete role %s (assigned during delete): %w", roleID, ErrRoleInUse)
		}
		return fmt.Errorf("delete role %s: %w", roleID, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete role %s: %w", roleID, err)
	}
	return nil
}

// GetSubjectRoles returns the roles assigned to a single user. In single-role
// mode (section 12.6) the slice is always length 0 (user not found returns
// ErrNotFound) or 1.
func (s *PostgresStore) GetSubjectRoles(ctx context.Context, userID string) ([]string, error) {
	var roleID string
	err := s.pool.QueryRow(ctx, `SELECT role FROM users WHERE user_id = $1`, userID).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("subject %s: %w", userID, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get subject %s roles: %w", userID, err)
	}
	return []string{roleID}, nil
}

// SetSubjectRole assigns a single role to a user. The role must already exist
// in rbac_policies (FK enforced). Returns ErrNotFound for missing user or
// missing role.
func (s *PostgresStore) SetSubjectRole(ctx context.Context, userID, roleID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE users SET role = $2, updated_at = NOW() WHERE user_id = $1`, userID, roleID)
	if err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("assign role %s to %s: %w", roleID, userID, ErrNotFound)
		}
		return fmt.Errorf("assign role %s to %s: %w", roleID, userID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("assign role to subject %s: %w", userID, ErrNotFound)
	}
	return nil
}

// rbacRowScanner is satisfied by both pgx.Rows and pgx.Row, letting scanRBACRole
// serve list and single-row paths uniformly.
type rbacRowScanner interface {
	Scan(dest ...any) error
}

func scanRBACRole(row rbacRowScanner) (domain.RBACRole, error) {
	var (
		role        domain.RBACRole
		permissions string
	)
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.System, &permissions, &role.CreatedAt, &role.UpdatedAt); err != nil {
		return domain.RBACRole{}, err
	}
	matrix := make(domain.PermissionMatrix)
	if err := json.Unmarshal([]byte(permissions), &matrix); err != nil {
		return domain.RBACRole{}, fmt.Errorf("decode permissions for %s: %w", role.ID, err)
	}
	role.Permissions = matrix
	return role, nil
}
