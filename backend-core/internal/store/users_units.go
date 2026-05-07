package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// ListUsers returns a paginated list of application users together with the
// total count after applying filters.
func (s *PostgresStore) ListUsers(ctx context.Context, opts domain.UserListOptions) ([]domain.AppUser, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const countQuery = `
SELECT COUNT(*)
FROM users
WHERE ($1 = '' OR role = $1)
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR primary_unit_id = $3)`

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, opts.Role, opts.Status, opts.PrimaryUnitID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	const query = `
SELECT
	id,
	user_id,
	email,
	display_name,
	role,
	status,
	COALESCE(primary_unit_id, ''),
	COALESCE(current_unit_id, ''),
	is_seconded,
	joined_at,
	created_at,
	updated_at
FROM users
WHERE ($3 = '' OR role = $3)
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR primary_unit_id = $5)
ORDER BY user_id ASC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.Role, opts.Status, opts.PrimaryUnitID)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.AppUser, 0, limit)
	userIDs := make([]string, 0, limit)
	for rows.Next() {
		var user domain.AppUser
		var role string
		var status string
		if err := rows.Scan(
			&user.ID,
			&user.UserID,
			&user.Email,
			&user.DisplayName,
			&role,
			&status,
			&user.PrimaryUnitID,
			&user.CurrentUnitID,
			&user.IsSeconded,
			&user.JoinedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		user.Role = domain.AppRole(role)
		user.Status = domain.UserStatus(status)
		users = append(users, user)
		userIDs = append(userIDs, user.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	if len(userIDs) > 0 {
		appointments, err := s.appointmentsForUsers(ctx, userIDs)
		if err != nil {
			return nil, 0, err
		}
		for i := range users {
			users[i].Appointments = appointments[users[i].UserID]
		}
	}

	return users, total, nil
}

// GetUser fetches a single user (including appointments) by user_id.
func (s *PostgresStore) GetUser(ctx context.Context, userID string) (domain.AppUser, error) {
	const query = `
SELECT
	id,
	user_id,
	email,
	display_name,
	role,
	status,
	COALESCE(primary_unit_id, ''),
	COALESCE(current_unit_id, ''),
	is_seconded,
	joined_at,
	created_at,
	updated_at
FROM users
WHERE user_id = $1
LIMIT 1`

	var user domain.AppUser
	var role string
	var status string
	err := s.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.UserID,
		&user.Email,
		&user.DisplayName,
		&role,
		&status,
		&user.PrimaryUnitID,
		&user.CurrentUnitID,
		&user.IsSeconded,
		&user.JoinedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AppUser{}, fmt.Errorf("user %s not found: %w", userID, err)
		}
		return domain.AppUser{}, fmt.Errorf("get user %s: %w", userID, err)
	}
	user.Role = domain.AppRole(role)
	user.Status = domain.UserStatus(status)

	appointments, err := s.GetUserAppointments(ctx, userID)
	if err != nil {
		return domain.AppUser{}, err
	}
	user.Appointments = appointments
	return user, nil
}

// GetUserAppointments returns the appointments (unit memberships and leader
// assignments) for a single user.
func (s *PostgresStore) GetUserAppointments(ctx context.Context, userID string) ([]domain.UnitAppointment, error) {
	const query = `
SELECT unit_id, user_id, appointment_role
FROM unit_appointments
WHERE user_id = $1
ORDER BY unit_id ASC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query appointments for user %s: %w", userID, err)
	}
	defer rows.Close()

	appointments := make([]domain.UnitAppointment, 0)
	for rows.Next() {
		var appointment domain.UnitAppointment
		var appointmentRole string
		if err := rows.Scan(&appointment.UnitID, &appointment.UserID, &appointmentRole); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		appointment.AppointmentRole = domain.AppointmentRole(appointmentRole)
		appointments = append(appointments, appointment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments: %w", err)
	}
	return appointments, nil
}

// appointmentsForUsers loads appointments for several users in a single query.
func (s *PostgresStore) appointmentsForUsers(ctx context.Context, userIDs []string) (map[string][]domain.UnitAppointment, error) {
	const query = `
SELECT unit_id, user_id, appointment_role
FROM unit_appointments
WHERE user_id = ANY($1)
ORDER BY user_id ASC, unit_id ASC`

	rows, err := s.pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("query appointments for users: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]domain.UnitAppointment, len(userIDs))
	for rows.Next() {
		var appointment domain.UnitAppointment
		var appointmentRole string
		if err := rows.Scan(&appointment.UnitID, &appointment.UserID, &appointmentRole); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		appointment.AppointmentRole = domain.AppointmentRole(appointmentRole)
		result[appointment.UserID] = append(result[appointment.UserID], appointment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments: %w", err)
	}
	return result, nil
}

// GetHierarchy returns the full org-unit hierarchy along with derived counts
// (direct members and total members in the descendant subtree).
func (s *PostgresStore) GetHierarchy(ctx context.Context) (domain.Hierarchy, error) {
	const unitsQuery = `
WITH RECURSIVE descendants AS (
	SELECT unit_id, unit_id AS root_id FROM org_units
	UNION ALL
	SELECT o.unit_id, d.root_id
	FROM org_units o JOIN descendants d ON o.parent_unit_id = d.unit_id
)
SELECT
	o.id,
	o.unit_id,
	COALESCE(o.parent_unit_id, ''),
	o.unit_type,
	o.label,
	COALESCE(o.leader_user_id, ''),
	o.position_x,
	o.position_y,
	o.created_at,
	o.updated_at,
	(SELECT COUNT(*) FROM unit_appointments a WHERE a.unit_id = o.unit_id) AS direct_count,
	(SELECT COUNT(DISTINCT a.user_id)
	   FROM descendants d JOIN unit_appointments a ON a.unit_id = d.unit_id
	   WHERE d.root_id = o.unit_id) AS total_count
FROM org_units o
ORDER BY o.unit_id ASC`

	rows, err := s.pool.Query(ctx, unitsQuery)
	if err != nil {
		return domain.Hierarchy{}, fmt.Errorf("query hierarchy units: %w", err)
	}
	defer rows.Close()

	units := make([]domain.OrgUnit, 0)
	for rows.Next() {
		var unit domain.OrgUnit
		var unitType string
		var directCount int64
		var totalCount int64
		if err := rows.Scan(
			&unit.ID,
			&unit.UnitID,
			&unit.ParentUnitID,
			&unitType,
			&unit.Label,
			&unit.LeaderUserID,
			&unit.PositionX,
			&unit.PositionY,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			&directCount,
			&totalCount,
		); err != nil {
			return domain.Hierarchy{}, fmt.Errorf("scan unit: %w", err)
		}
		unit.UnitType = domain.UnitType(unitType)
		unit.DirectCount = int(directCount)
		unit.TotalCount = int(totalCount)
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return domain.Hierarchy{}, fmt.Errorf("iterate units: %w", err)
	}

	const edgesQuery = `
SELECT parent_unit_id, unit_id
FROM org_units
WHERE parent_unit_id IS NOT NULL
ORDER BY parent_unit_id, unit_id`

	edgeRows, err := s.pool.Query(ctx, edgesQuery)
	if err != nil {
		return domain.Hierarchy{}, fmt.Errorf("query hierarchy edges: %w", err)
	}
	defer edgeRows.Close()

	edges := make([]domain.OrgEdge, 0)
	for edgeRows.Next() {
		var edge domain.OrgEdge
		if err := edgeRows.Scan(&edge.SourceUnitID, &edge.TargetUnitID); err != nil {
			return domain.Hierarchy{}, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, edge)
	}
	if err := edgeRows.Err(); err != nil {
		return domain.Hierarchy{}, fmt.Errorf("iterate edges: %w", err)
	}

	return domain.Hierarchy{Units: units, Edges: edges}, nil
}

// ListUnitMembers returns all users (with their appointments) attached to a
// specific unit.
func (s *PostgresStore) ListUnitMembers(ctx context.Context, unitID string) ([]domain.AppUser, error) {
	const query = `
SELECT DISTINCT
	u.id,
	u.user_id,
	u.email,
	u.display_name,
	u.role,
	u.status,
	COALESCE(u.primary_unit_id, ''),
	COALESCE(u.current_unit_id, ''),
	u.is_seconded,
	u.joined_at,
	u.created_at,
	u.updated_at
FROM users u
JOIN unit_appointments a ON a.user_id = u.user_id
WHERE a.unit_id = $1
ORDER BY u.user_id ASC`

	rows, err := s.pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("list unit %s members: %w", unitID, err)
	}
	defer rows.Close()

	users := make([]domain.AppUser, 0)
	userIDs := make([]string, 0)
	for rows.Next() {
		var user domain.AppUser
		var role string
		var status string
		if err := rows.Scan(
			&user.ID,
			&user.UserID,
			&user.Email,
			&user.DisplayName,
			&role,
			&status,
			&user.PrimaryUnitID,
			&user.CurrentUnitID,
			&user.IsSeconded,
			&user.JoinedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan unit member: %w", err)
		}
		user.Role = domain.AppRole(role)
		user.Status = domain.UserStatus(status)
		users = append(users, user)
		userIDs = append(userIDs, user.UserID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unit members: %w", err)
	}

	if len(userIDs) > 0 {
		appointments, err := s.appointmentsForUsers(ctx, userIDs)
		if err != nil {
			return nil, err
		}
		for i := range users {
			users[i].Appointments = appointments[users[i].UserID]
		}
	}
	return users, nil
}

// ReplaceUnitMembers replaces the member appointments for a unit while
// preserving leader appointments. The new set is provided as a list of user
// IDs and is applied atomically inside a single transaction.
func (s *PostgresStore) ReplaceUnitMembers(ctx context.Context, unitID string, userIDs []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace unit members tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists int
	if err := tx.QueryRow(ctx, `SELECT 1 FROM org_units WHERE unit_id = $1`, unitID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("unit %s not found: %w", unitID, ErrNotFound)
		}
		return fmt.Errorf("verify unit %s: %w", unitID, err)
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM unit_appointments WHERE unit_id = $1 AND appointment_role = 'member'`,
		unitID,
	); err != nil {
		return fmt.Errorf("clear members for unit %s: %w", unitID, err)
	}

	if len(userIDs) > 0 {
		const insertQuery = `
INSERT INTO unit_appointments (user_id, unit_id, appointment_role)
VALUES ($1, $2, 'member')
ON CONFLICT (user_id, unit_id) DO UPDATE
	SET appointment_role = CASE
		WHEN unit_appointments.appointment_role = 'leader' THEN 'leader'
		ELSE EXCLUDED.appointment_role
	END`
		for _, userID := range userIDs {
			if userID == "" {
				continue
			}
			if _, err := tx.Exec(ctx, insertQuery, userID, unitID); err != nil {
				return fmt.Errorf("insert member %s for unit %s: %w", userID, unitID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace unit members: %w", err)
	}
	return nil
}
