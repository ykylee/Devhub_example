package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

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
	COALESCE(user_type, 'human'),
	COALESCE(idp_subject, ''),
	COALESCE(primary_unit_id, ''),
	COALESCE(current_unit_id, ''),
	is_seconded,
	joined_at,
	onboarding_completed_at,
	COALESCE(review_status, ''),
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
		var userType string
		if err := rows.Scan(
			&user.ID,
			&user.UserID,
			&user.Email,
			&user.DisplayName,
			&role,
			&status,
			&userType,
			&user.IdPSubject,
			&user.PrimaryUnitID,
			&user.CurrentUnitID,
			&user.IsSeconded,
			&user.JoinedAt,
			&user.OnboardingCompletedAt,
			&user.ReviewStatus,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		user.Role = domain.AppRole(role)
		user.Status = domain.UserStatus(status)
		user.Type = domain.UserType(userType)
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
	COALESCE(idp_subject, ''),
	COALESCE(primary_unit_id, ''),
	COALESCE(current_unit_id, ''),
	is_seconded,
	joined_at,
	onboarding_completed_at,
	COALESCE(review_status, ''),
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
		&user.IdPSubject,
		&user.PrimaryUnitID,
		&user.CurrentUnitID,
		&user.IsSeconded,
		&user.JoinedAt,
		&user.OnboardingCompletedAt,
		&user.ReviewStatus,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Match the convention used elsewhere in this file
			// (UpdateUser, DeleteUser, SetIdPSubject): wrap with
			// the store.ErrNotFound sentinel so callers can `errors.Is`
			// portably. The earlier `%w pgx.ErrNoRows` shape silently
			// broke organization.go's 404 path and authenticateActor's
			// pre-onboarding suppression.
			return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, ErrNotFound)
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

// GetUserByIdPSubject — ADR-0020 sub-carve C (sprint -k, issue #212 codex P1
// hotfix). Keycloak admin event 처리 시 ResourcePath 의 identity_id (Keycloak
// UUID) 로 DevHub users 행을 lookup. USER:DELETE 이벤트 처리 시 Keycloak user
// 가 이미 gone 이라 admin client GetUserDetails 가 404 — 본 메서드가 유일 lookup
// 경로. idp_subject 컬럼은 UNIQUE (migration 000009 + 000030 rename) 이라
// O(1) lookup.
func (s *PostgresStore) GetUserByIdPSubject(ctx context.Context, identityID string) (domain.AppUser, error) {
	const query = `
SELECT
	id,
	user_id,
	email,
	display_name,
	role,
	status,
	COALESCE(idp_subject, ''),
	COALESCE(primary_unit_id, ''),
	COALESCE(current_unit_id, ''),
	is_seconded,
	joined_at,
	onboarding_completed_at,
	COALESCE(review_status, ''),
	created_at,
	updated_at
FROM users
WHERE idp_subject = $1
LIMIT 1`

	var user domain.AppUser
	var role string
	var status string
	err := s.pool.QueryRow(ctx, query, identityID).Scan(
		&user.ID,
		&user.UserID,
		&user.Email,
		&user.DisplayName,
		&role,
		&status,
		&user.IdPSubject,
		&user.PrimaryUnitID,
		&user.CurrentUnitID,
		&user.IsSeconded,
		&user.JoinedAt,
		&user.OnboardingCompletedAt,
		&user.ReviewStatus,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AppUser{}, fmt.Errorf("user idp_subject=%s: %w", identityID, ErrNotFound)
		}
		return domain.AppUser{}, fmt.Errorf("get user idp_subject=%s: %w", identityID, err)
	}
	user.Role = domain.AppRole(role)
	user.Status = domain.UserStatus(status)

	appointments, err := s.GetUserAppointments(ctx, user.UserID)
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

// GetHierarchy returns the full org-unit hierarchy along with derived counts.
// 사용자 보고 2026-05-26 (PR #335 frontend 보강 + 본 backend 정합):
// 한 user 가 multiple unit 에 속할 때 canonical unit 한 곳만 카운트.
// canonical = leader 직책 우선 + 동일 role 내에서 가장 상위 (depth 최소) unit.
// direct_count + total_count 모두 canonical 기반 dedupe. ADR-0024 와 정합.
func (s *PostgresStore) GetHierarchy(ctx context.Context) (domain.Hierarchy, error) {
	const unitsQuery = `
WITH RECURSIVE
descendants AS (
	SELECT unit_id, unit_id AS root_id FROM org_units
	UNION ALL
	SELECT o.unit_id, d.root_id
	FROM org_units o JOIN descendants d ON o.parent_unit_id = d.unit_id
),
depths AS (
	SELECT unit_id, 1 AS depth FROM org_units WHERE parent_unit_id IS NULL OR parent_unit_id = ''
	UNION ALL
	SELECT o.unit_id, d.depth + 1
	FROM org_units o JOIN depths d ON o.parent_unit_id = d.unit_id
),
ranked_appointments AS (
	SELECT
		a.user_id,
		a.unit_id,
		ROW_NUMBER() OVER (
			PARTITION BY a.user_id
			ORDER BY
				CASE WHEN a.appointment_role = 'leader' THEN 0 ELSE 1 END,
				d.depth ASC,
				a.unit_id ASC
		) AS rn
	FROM unit_appointments a JOIN depths d ON a.unit_id = d.unit_id
),
canonical AS (
	SELECT user_id, unit_id FROM ranked_appointments WHERE rn = 1
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
	(SELECT COUNT(*) FROM canonical c WHERE c.unit_id = o.unit_id) AS direct_count,
	(SELECT COUNT(DISTINCT c.user_id)
	   FROM descendants d JOIN canonical c ON c.unit_id = d.unit_id
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
	COALESCE(u.idp_subject, ''),
	COALESCE(u.primary_unit_id, ''),
	COALESCE(u.current_unit_id, ''),
	u.is_seconded,
	u.joined_at,
	u.onboarding_completed_at,
	COALESCE(u.review_status, ''),
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
			&user.IdPSubject,
			&user.PrimaryUnitID,
			&user.CurrentUnitID,
			&user.IsSeconded,
			&user.JoinedAt,
			&user.OnboardingCompletedAt,
			&user.ReviewStatus,
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

// nullableUnitID converts an empty string to nil so it can be inserted into
// nullable foreign-key columns without tripping the FK constraint.
func nullableUnitID(unitID string) any {
	if strings.TrimSpace(unitID) == "" {
		return nil
	}
	return unitID
}

// CreateUser inserts a new application user. Returns ErrConflict if user_id or
// email already exist.
func (s *PostgresStore) CreateUser(ctx context.Context, input domain.CreateUserInput) (domain.AppUser, error) {
	joinedAt := input.JoinedAt
	if joinedAt.IsZero() {
		joinedAt = time.Now().UTC()
	}

	const query = `
INSERT INTO users (
	user_id,
	email,
	display_name,
	role,
	status,
	user_type,
	primary_unit_id,
	current_unit_id,
	is_seconded,
	joined_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at, updated_at`

	var user domain.AppUser
	err := s.pool.QueryRow(ctx, query,
		input.UserID,
		input.Email,
		input.DisplayName,
		string(input.Role),
		string(input.Status),
		string(input.Type),
		nullableUnitID(input.PrimaryUnitID),
		nullableUnitID(input.CurrentUnitID),
		input.IsSeconded,
		joinedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.AppUser{}, fmt.Errorf("create user %s: %w", input.UserID, ErrConflict)
		}
		if isForeignKeyViolation(err) {
			return domain.AppUser{}, fmt.Errorf("create user %s references missing unit: %w", input.UserID, ErrNotFound)
		}
		return domain.AppUser{}, fmt.Errorf("insert user: %w", err)
	}
	user.UserID = input.UserID
	user.Email = input.Email
	user.DisplayName = input.DisplayName
	user.Role = input.Role
	user.Status = input.Status
	user.Type = input.Type
	user.PrimaryUnitID = input.PrimaryUnitID
	user.CurrentUnitID = input.CurrentUnitID
	user.IsSeconded = input.IsSeconded
	user.JoinedAt = joinedAt
	user.Appointments = []domain.UnitAppointment{}
	return user, nil
}

// UpdateUser updates a subset of user fields. Only non-nil fields on input are
// applied. Returns ErrNotFound if no row matches.
func (s *PostgresStore) UpdateUser(ctx context.Context, userID string, input domain.UpdateUserInput) (domain.AppUser, error) {
	setClauses := make([]string, 0, 8)
	args := make([]any, 0, 9)
	args = append(args, userID)
	idx := 2

	if input.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", idx))
		args = append(args, *input.Email)
		idx++
	}
	if input.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", idx))
		args = append(args, *input.DisplayName)
		idx++
	}
	if input.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", idx))
		args = append(args, string(*input.Role))
		idx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*input.Status))
		idx++
	}
	if input.PrimaryUnitID != nil {
		setClauses = append(setClauses, fmt.Sprintf("primary_unit_id = $%d", idx))
		args = append(args, nullableUnitID(*input.PrimaryUnitID))
		idx++
	}
	if input.CurrentUnitID != nil {
		setClauses = append(setClauses, fmt.Sprintf("current_unit_id = $%d", idx))
		args = append(args, nullableUnitID(*input.CurrentUnitID))
		idx++
	}
	if input.IsSeconded != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_seconded = $%d", idx))
		args = append(args, *input.IsSeconded)
		idx++
	}
	if input.JoinedAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("joined_at = $%d", idx))
		args = append(args, *input.JoinedAt)
		idx++
	}
	if input.ReviewStatus != nil {
		// RM-ONBOARD-01: admin 의 명시 transition (예: ConfirmUserReview)
		// 은 ConfirmUserReview method 사용 권장. UpdateUser 의
		// ReviewStatus 분기는 system_admin 의 admin/settings/users 직접
		// 갱신 backdoor (legacy / 예외 case) — onboarding_completed_at 변경 없음.
		setClauses = append(setClauses, fmt.Sprintf("review_status = $%d", idx))
		if *input.ReviewStatus == "" {
			args = append(args, nil) // bi-implication CHECK 으로 onboarding_completed_at NULL 일 때만 가능
		} else {
			args = append(args, *input.ReviewStatus)
		}
		idx++
	}

	if len(setClauses) == 0 {
		// Nothing to update — just return the current row.
		return s.GetUser(ctx, userID)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf(`UPDATE users SET %s WHERE user_id = $1`, strings.Join(setClauses, ", "))

	tag, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.AppUser{}, fmt.Errorf("update user %s: %w", userID, ErrConflict)
		}
		if isForeignKeyViolation(err) {
			return domain.AppUser{}, fmt.Errorf("update user %s references missing unit: %w", userID, ErrNotFound)
		}
		return domain.AppUser{}, fmt.Errorf("update user %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.AppUser{}, fmt.Errorf("user %s: %w", userID, ErrNotFound)
	}

	return s.GetUser(ctx, userID)
}

// DeleteUser removes a user along with their unit appointments (cascade). Returns
// ErrNotFound when no row matches.
func (s *PostgresStore) DeleteUser(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user %s: %w", userID, ErrNotFound)
	}
	return nil
}

// SubmitOnboarding — RM-ONBOARD-01 (ADR-0021 §3.3, API-83 §16.3) — onboarding
// 제출 시 단일 트랜잭션으로 (a) users row INSERT (DB 미등록 사용자) 또는
// UPDATE (관리자 사전 등록된 미완료 사용자), (b) display_name + primary_unit_id
// + email + idp_subject 설정, (c) onboarding_completed_at = NOW(), (d)
// review_status = 'pending_review'. role 은 caller (handler) 가 결정 — Keycloak
// claim 매핑 또는 fallback `developer` (REQ-FR-ONBOARD-002 / §3.1).
//
// 정합 정책 (codex P1 PR #270):
//   - existing row 가 이미 onboarding_completed_at IS NOT NULL → ErrConflict
//     (이미 완료된 사용자 중복 호출, API-83 의 409 분기).
//   - existing row 가 onboarding_completed_at IS NULL → UPDATE (pre-seeded
//     사용자가 첫 로그인 후 onboarding 화면에서 제출).
//   - row 없음 → INSERT.
//
// Returns ErrConflict (이미 완료) / ErrNotFound (primary_unit_id 가
// organization_units 에 없음, FK violation).
func (s *PostgresStore) SubmitOnboarding(ctx context.Context, input domain.OnboardingSubmitInput) (domain.AppUser, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return domain.AppUser{}, errors.New("user_id is required")
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		return domain.AppUser{}, errors.New("display_name is required")
	}
	if strings.TrimSpace(input.PrimaryUnitID) == "" {
		return domain.AppUser{}, errors.New("primary_unit_id is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.AppUser{}, fmt.Errorf("begin submit onboarding tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. existing row 확인 (FOR UPDATE row lock).
	var existing struct {
		exists                bool
		onboardingCompletedAt *time.Time
	}
	row := tx.QueryRow(ctx,
		`SELECT onboarding_completed_at FROM users WHERE user_id = $1 FOR UPDATE`,
		input.UserID,
	)
	scanErr := row.Scan(&existing.onboardingCompletedAt)
	switch {
	case scanErr == nil:
		existing.exists = true
		if existing.onboardingCompletedAt != nil {
			return domain.AppUser{}, fmt.Errorf("user %s already completed onboarding: %w", input.UserID, ErrConflict)
		}
	case errors.Is(scanErr, pgx.ErrNoRows):
		existing.exists = false
	default:
		return domain.AppUser{}, fmt.Errorf("submit onboarding lookup: %w", scanErr)
	}

	// 2. INSERT 또는 UPDATE.
	role := input.FallbackRole
	if role == "" {
		role = domain.AppRoleDeveloper
	}
	if existing.exists {
		_, err = tx.Exec(ctx, `
UPDATE users SET
	display_name = $2,
	email = COALESCE(NULLIF($3, ''), email),
	primary_unit_id = $4,
	current_unit_id = $4,
	idp_subject = COALESCE(NULLIF($5, ''), idp_subject),
	onboarding_completed_at = NOW(),
	review_status = 'pending_review',
	updated_at = NOW()
WHERE user_id = $1`,
			input.UserID,
			input.DisplayName,
			input.Email,
			input.PrimaryUnitID,
			input.IdPSubject,
		)
	} else {
		_, err = tx.Exec(ctx, `
INSERT INTO users (
	user_id, email, display_name, role, status, user_type,
	idp_subject, primary_unit_id, current_unit_id, is_seconded, joined_at,
	onboarding_completed_at, review_status
) VALUES (
	$1, $2, $3, $4, 'active', 'human',
	NULLIF($5, ''), $6, $6, false, NOW(),
	NOW(), 'pending_review'
)`,
			input.UserID,
			input.Email,
			input.DisplayName,
			string(role),
			input.IdPSubject,
			input.PrimaryUnitID,
		)
	}
	if err != nil {
		if isUniqueViolation(err) {
			return domain.AppUser{}, fmt.Errorf("submit onboarding %s: %w", input.UserID, ErrConflict)
		}
		if isForeignKeyViolation(err) {
			return domain.AppUser{}, fmt.Errorf("submit onboarding %s references missing unit %s: %w", input.UserID, input.PrimaryUnitID, ErrNotFound)
		}
		return domain.AppUser{}, fmt.Errorf("submit onboarding %s: %w", input.UserID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AppUser{}, fmt.Errorf("commit submit onboarding: %w", err)
	}

	return s.GetUser(ctx, input.UserID)
}

// ConfirmUserReview — RM-ONBOARD-01 (API-86 §16.7). system_admin 의 명시
// transition (pending_review → reviewed). 사용자의 onboarding 이 이미 완료된
// 상태에서만 동작 (onboarding_completed_at IS NOT NULL).
//
// Returns ErrNotFound (user 미존재) / ErrConflict (이미 reviewed 이거나
// onboarding 미제출).
func (s *PostgresStore) ConfirmUserReview(ctx context.Context, userID string) (domain.AppUser, error) {
	if strings.TrimSpace(userID) == "" {
		return domain.AppUser{}, errors.New("user_id is required")
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE users SET
	review_status = 'reviewed',
	updated_at = NOW()
WHERE user_id = $1
  AND onboarding_completed_at IS NOT NULL
  AND review_status = 'pending_review'`,
		userID,
	)
	if err != nil {
		return domain.AppUser{}, fmt.Errorf("confirm user review %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		// 정확한 분기는 caller (handler) 가 GetUser 로 다시 확인 후 결정.
		// 본 store layer 는 단순 affected=0 → ErrNotFound 반환.
		// handler 가 404 (user not found) 또는 409 (already reviewed) 또는
		// 422 (onboarding not completed) 로 분기.
		return domain.AppUser{}, fmt.Errorf("user %s review confirm: %w", userID, ErrNotFound)
	}
	return s.GetUser(ctx, userID)
}

// CountPendingReview — pending_review 상태의 user 수 (Onboarding SOP §4.4
// Prometheus Gauge 의 cron refresh 호출처). httpapi.RunOnboardingPendingReviewGauge
// 가 본 메서드를 interval 마다 호출 → devhub_onboarding_pending_review_count Gauge
// update. read-only SELECT COUNT — index 가 없으면 full scan 이라 row 100K+ 에서
// 부담 가능, 그러나 review_status partial index 는 별도 carve (현재 user 테이블
// 사이즈 < 10K 수준이라 무시 가능). 운영 중 row scale 증가 시 partial index 추가
// (`CREATE INDEX users_pending_review_idx ON users(review_status) WHERE review_status='pending_review'`).
func (s *PostgresStore) CountPendingReview(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM users
 WHERE onboarding_completed_at IS NOT NULL
   AND review_status = 'pending_review'`,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count pending_review: %w", err)
	}
	return count, nil
}

// SetIdPSubject caches the IdP identity_id on the DevHub users row so
// subsequent identity lookups can skip the IdP user-list scan. Migration
// 000009 added the column (as kratos_identity_id); 000030 renamed it to
// idp_subject after ADR-0019. This is the only writer. Returns ErrNotFound
// when no user matches the given user_id — best-effort callers (lazy
// backfill paths) typically ignore that case.
func (s *PostgresStore) SetIdPSubject(ctx context.Context, userID, identityID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(identityID) == "" {
		return errors.New("identity_id is required")
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET idp_subject = $2, updated_at = NOW() WHERE user_id = $1`,
		userID, identityID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("idp_subject %s already mapped: %w", identityID, ErrConflict)
		}
		return fmt.Errorf("set idp_subject for %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user %s: %w", userID, ErrNotFound)
	}
	return nil
}

// SearchOrgUnits — RM-ONBOARD-01 (API-84 §16.4) — typeahead 검색. case-
// insensitive substring match on org_units.label. q 는 caller (handler) 가
// >= 2 chars 검증. limit > 20 또는 <= 0 인 경우 20 으로 clamp.
//
// 응답 row 는 domain.OrgUnit 전체 shape — handler 가 unit_id + label 만 노출.
// 권한 가드 없음 (REQ-FR-ONBOARD-004) — 모든 사용자에게 모든 조직 후보 노출.
func (s *PostgresStore) SearchOrgUnits(ctx context.Context, q string, limit int) ([]domain.OrgUnit, error) {
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	pattern := "%" + strings.ToLower(q) + "%"
	const query = `
SELECT
	id,
	unit_id,
	COALESCE(parent_unit_id, ''),
	unit_type,
	label,
	COALESCE(leader_user_id, ''),
	position_x,
	position_y,
	created_at,
	updated_at
FROM org_units
WHERE LOWER(label) LIKE $1
ORDER BY label ASC, unit_id ASC
LIMIT $2`

	rows, err := s.pool.Query(ctx, query, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search org units: %w", err)
	}
	defer rows.Close()
	units := make([]domain.OrgUnit, 0, limit)
	for rows.Next() {
		var unit domain.OrgUnit
		var unitType string
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
		); err != nil {
			return nil, fmt.Errorf("scan org unit search: %w", err)
		}
		unit.UnitType = domain.UnitType(unitType)
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate org unit search: %w", err)
	}
	return units, nil
}

// GetOrgUnit fetches a single org unit (without descendants).
func (s *PostgresStore) GetOrgUnit(ctx context.Context, unitID string) (domain.OrgUnit, error) {
	const query = `
SELECT
	id,
	unit_id,
	COALESCE(parent_unit_id, ''),
	unit_type,
	label,
	COALESCE(leader_user_id, ''),
	position_x,
	position_y,
	created_at,
	updated_at
FROM org_units
WHERE unit_id = $1
LIMIT 1`

	var unit domain.OrgUnit
	var unitType string
	err := s.pool.QueryRow(ctx, query, unitID).Scan(
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
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", unitID, ErrNotFound)
		}
		return domain.OrgUnit{}, fmt.Errorf("get unit %s: %w", unitID, err)
	}
	unit.UnitType = domain.UnitType(unitType)
	return unit, nil
}

// normalizedLeaderUserID is the single normalize point so that
// org_units.leader_user_id and unit_appointments.user_id always agree.
// Returns "" for blank/whitespace-only input.
func normalizedLeaderUserID(raw string) string {
	return strings.TrimSpace(raw)
}

// CreateOrgUnit inserts a new organization unit.
func (s *PostgresStore) CreateOrgUnit(ctx context.Context, input domain.CreateOrgUnitInput) (domain.OrgUnit, error) {
	const query = `
INSERT INTO org_units (
	unit_id,
	parent_unit_id,
	unit_type,
	label,
	leader_user_id,
	position_x,
	position_y
) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
RETURNING id, created_at, updated_at`

	leader := normalizedLeaderUserID(input.LeaderUserID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.OrgUnit{}, fmt.Errorf("begin create unit tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var unit domain.OrgUnit
	err = tx.QueryRow(ctx, query,
		input.UnitID,
		nullableUnitID(input.ParentUnitID),
		string(input.UnitType),
		input.Label,
		leader,
		input.PositionX,
		input.PositionY,
	).Scan(&unit.ID, &unit.CreatedAt, &unit.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.OrgUnit{}, fmt.Errorf("create unit %s: %w", input.UnitID, ErrConflict)
		}
		if isForeignKeyViolation(err) {
			return domain.OrgUnit{}, fmt.Errorf("create unit %s references missing parent: %w", input.UnitID, ErrNotFound)
		}
		return domain.OrgUnit{}, fmt.Errorf("insert unit: %w", err)
	}

	if leader != "" {
		// Keep org_units.leader_user_id and unit_appointments in sync.
		if _, err := tx.Exec(ctx,
			`INSERT INTO unit_appointments (user_id, unit_id, appointment_role)
			 VALUES ($1, $2, 'leader')
			 ON CONFLICT (user_id, unit_id)
			 DO UPDATE SET appointment_role = 'leader'`,
			leader, input.UnitID,
		); err != nil {
			if isForeignKeyViolation(err) {
				return domain.OrgUnit{}, fmt.Errorf("create unit %s leader %s: %w", input.UnitID, leader, ErrNotFound)
			}
			return domain.OrgUnit{}, fmt.Errorf("sync leader appointment for unit %s: %w", input.UnitID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OrgUnit{}, fmt.Errorf("commit create unit tx: %w", err)
	}

	unit.UnitID = input.UnitID
	unit.ParentUnitID = input.ParentUnitID
	unit.UnitType = input.UnitType
	unit.Label = input.Label
	unit.LeaderUserID = leader
	unit.PositionX = input.PositionX
	unit.PositionY = input.PositionY
	return unit, nil
}

// UpdateOrgUnit updates a subset of unit fields.
func (s *PostgresStore) UpdateOrgUnit(ctx context.Context, unitID string, input domain.UpdateOrgUnitInput) (domain.OrgUnit, error) {
	setClauses := make([]string, 0, 6)
	args := make([]any, 0, 7)
	args = append(args, unitID)
	idx := 2

	if input.ParentUnitID != nil {
		if *input.ParentUnitID == unitID {
			return domain.OrgUnit{}, fmt.Errorf("unit %s cannot be its own parent: %w", unitID, ErrConflict)
		}
		setClauses = append(setClauses, fmt.Sprintf("parent_unit_id = $%d", idx))
		args = append(args, nullableUnitID(*input.ParentUnitID))
		idx++
	}
	if input.UnitType != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit_type = $%d", idx))
		args = append(args, string(*input.UnitType))
		idx++
	}
	if input.Label != nil {
		setClauses = append(setClauses, fmt.Sprintf("label = $%d", idx))
		args = append(args, *input.Label)
		idx++
	}
	var leaderNormalized string
	if input.LeaderUserID != nil {
		leaderNormalized = normalizedLeaderUserID(*input.LeaderUserID)
		setClauses = append(setClauses, fmt.Sprintf("leader_user_id = NULLIF($%d, '')", idx))
		args = append(args, leaderNormalized)
		idx++
	}
	if input.PositionX != nil {
		setClauses = append(setClauses, fmt.Sprintf("position_x = $%d", idx))
		args = append(args, *input.PositionX)
		idx++
	}
	if input.PositionY != nil {
		setClauses = append(setClauses, fmt.Sprintf("position_y = $%d", idx))
		args = append(args, *input.PositionY)
		idx++
	}

	if len(setClauses) == 0 {
		return s.GetOrgUnit(ctx, unitID)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf(`UPDATE org_units SET %s WHERE unit_id = $1`, strings.Join(setClauses, ", "))

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.OrgUnit{}, fmt.Errorf("begin update unit tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serialize concurrent updates on the same unit so leader demote-then-
	// promote cannot interleave between admins (which would leave two leaders
	// despite the partial unique index ensuring eventual consistency).
	var lockedID int64
	if err := tx.QueryRow(ctx,
		`SELECT id FROM org_units WHERE unit_id = $1 FOR UPDATE`, unitID,
	).Scan(&lockedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", unitID, ErrNotFound)
		}
		return domain.OrgUnit{}, fmt.Errorf("lock unit %s: %w", unitID, err)
	}

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.OrgUnit{}, fmt.Errorf("update unit %s references missing parent: %w", unitID, ErrNotFound)
		}
		return domain.OrgUnit{}, fmt.Errorf("update unit %s: %w", unitID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.OrgUnit{}, fmt.Errorf("unit %s: %w", unitID, ErrNotFound)
	}

	if input.LeaderUserID != nil {
		// Normalize existing leader appointments first so unit-level leader
		// remains a single source of truth.
		if _, err := tx.Exec(ctx,
			`UPDATE unit_appointments
			 SET appointment_role = 'member'
			 WHERE unit_id = $1 AND appointment_role = 'leader'`,
			unitID,
		); err != nil {
			return domain.OrgUnit{}, fmt.Errorf("demote existing leaders for unit %s: %w", unitID, err)
		}
		if leaderNormalized != "" {
			if _, err := tx.Exec(ctx,
				`INSERT INTO unit_appointments (user_id, unit_id, appointment_role)
				 VALUES ($1, $2, 'leader')
				 ON CONFLICT (user_id, unit_id)
				 DO UPDATE SET appointment_role = 'leader'`,
				leaderNormalized, unitID,
			); err != nil {
				if isForeignKeyViolation(err) {
					return domain.OrgUnit{}, fmt.Errorf("update unit %s leader %s: %w", unitID, leaderNormalized, ErrNotFound)
				}
				return domain.OrgUnit{}, fmt.Errorf("sync leader appointment for unit %s: %w", unitID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.OrgUnit{}, fmt.Errorf("commit update unit tx: %w", err)
	}
	return s.GetOrgUnit(ctx, unitID)
}

// DeleteOrgUnit removes an org unit. Children unit references and member
// appointments are handled by ON DELETE constraints in the schema (parent_unit_id
// becomes NULL, appointments cascade-delete).
func (s *PostgresStore) DeleteOrgUnit(ctx context.Context, unitID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM org_units WHERE unit_id = $1`, unitID)
	if err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("delete unit %s: %w", unitID, ErrConflict)
		}
		return fmt.Errorf("delete unit %s: %w", unitID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("unit %s: %w", unitID, ErrNotFound)
	}
	return nil
}

func (s *PostgresStore) UpdateHierarchy(ctx context.Context, hie domain.Hierarchy) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, unit := range hie.Units {
		const query = "UPDATE org_units SET position_x = $2, position_y = $3, updated_at = NOW() WHERE unit_id = $1"
		_, err := tx.Exec(ctx, query, unit.UnitID, unit.PositionX, unit.PositionY)
		if err != nil {
			return fmt.Errorf("update unit %s position: %w", unit.UnitID, err)
		}
	}

	return tx.Commit(ctx)
}

// ListOrgUnitIDsByLeader returns the unit_id of every org unit where the given
// user is the leader. Used by 6-P3 org_head scope enforcement.
func (s *PostgresStore) ListOrgUnitIDsByLeader(ctx context.Context, leaderUserID string) ([]string, error) {
	const query = `SELECT unit_id FROM org_units WHERE leader_user_id = $1`
	rows, err := s.pool.Query(ctx, query, leaderUserID)
	if err != nil {
		return nil, fmt.Errorf("list org units by leader: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan org unit id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetOrgUnitSubtreeIDs returns the unit_id of the given org unit and all its
// descendants (recursive CTE via parent_unit_id). Used by 6-P3 scope expansion.
func (s *PostgresStore) GetOrgUnitSubtreeIDs(ctx context.Context, unitID string) ([]string, error) {
	const query = `
WITH RECURSIVE subtree AS (
    SELECT unit_id FROM org_units WHERE unit_id = $1
    UNION ALL
    SELECT ou.unit_id FROM org_units ou JOIN subtree s ON ou.parent_unit_id = s.unit_id
)
SELECT unit_id FROM subtree`
	rows, err := s.pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("get org unit subtree: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan subtree unit id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
