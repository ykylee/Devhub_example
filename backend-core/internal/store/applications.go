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
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrNotImplemented signals that a method is registered in the store interface but
// its PostgreSQL body has not been written yet. Sprint claude/work_260514-a 의 carve in 은
// store 인터페이스 + handler stub 까지이며, 실 SQL 구현은 후속 sprint 의 carve out.
var ErrNotImplemented = errors.New("not implemented")

// ApplicationListOptions parameterizes ListApplications.
type ApplicationListOptions struct {
	Status          string
	IncludeArchived bool
	Query           string
	Limit           int
	Offset          int
}

// ApplicationRepositoryLinkKey identifies a single link row (composite PK).
type ApplicationRepositoryLinkKey struct {
	ApplicationID string
	RepoProvider  string
	RepoFullName  string
}

// ProjectListOptions parameterizes ListProjects (within a Repository scope).
type ProjectListOptions struct {
	RepositoryID    int64
	Status          string
	IncludeArchived bool
	Limit           int
	Offset          int
}

// --- Applications ---

func (s *PostgresStore) ListApplications(ctx context.Context, opts ApplicationListOptions) ([]domain.Application, int, error) {
	return nil, 0, ErrNotImplemented
}

func (s *PostgresStore) GetApplication(ctx context.Context, applicationID string) (domain.Application, error) {
	query := `
		SELECT id, key, name, description, status, visibility, owner_user_id, start_date, due_date, archived_at, created_at, updated_at
		FROM applications
		WHERE id = $1
	`
	var app domain.Application
	var startDate, dueDate pgtype.Date
	var archivedAt pgtype.Timestamptz
	var ownerUserID, description pgtype.Text

	err := s.pool.QueryRow(ctx, query, applicationID).Scan(
		&app.ID,
		&app.Key,
		&app.Name,
		&description,
		&app.Status,
		&app.Visibility,
		&ownerUserID,
		&startDate,
		&dueDate,
		&archivedAt,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Application{}, ErrNotFound
		}
		return domain.Application{}, err
	}

	app.OwnerUserID = ownerUserID.String
	app.Description = description.String
	if startDate.Valid {
		app.StartDate = &startDate.Time
	}
	if dueDate.Valid {
		app.DueDate = &dueDate.Time
	}
	if archivedAt.Valid {
		app.ArchivedAt = &archivedAt.Time
	}

	return app, nil
}


func (s *PostgresStore) CreateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	query := `
		INSERT INTO applications (key, name, description, status, visibility, owner_user_id, start_date, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	var startDate, dueDate pgtype.Date
	if app.StartDate != nil {
		startDate = pgtype.Date{Time: *app.StartDate, Valid: true}
	}
	if app.DueDate != nil {
		dueDate = pgtype.Date{Time: *app.DueDate, Valid: true}
	}

	err := s.pool.QueryRow(ctx, query,
		app.Key,
		app.Name,
		app.Description,
		app.Status,
		app.Visibility,
		app.OwnerUserID,
		startDate,
		dueDate,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "applications_key_key") {
				return domain.Application{}, ErrConflict
			}
		}
		return domain.Application{}, err
	}

	return app, nil
}

// UpdateApplication mutates allowed fields and enforces the ADR-0011 §4.1 1차 정책
// (system_admin 일임) via the handler layer; this store method is unconditional.
// 상태 전이 가드 검증은 concept §13.2.1 에 따라 handler/service 에서 수행하고, store 는
// 단순히 단일 SQL UPDATE 만 책임진다.
func (s *PostgresStore) UpdateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	query := `
		UPDATE applications
		SET 
			name = $2,
			description = $3,
			status = $4,
			visibility = $5,
			owner_user_id = $6,
			start_date = $7,
			due_date = $8,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	var startDate, dueDate pgtype.Date
	if app.StartDate != nil {
		startDate = pgtype.Date{Time: *app.StartDate, Valid: true}
	}
	if app.DueDate != nil {
		dueDate = pgtype.Date{Time: *app.DueDate, Valid: true}
	}
	
	err := s.pool.QueryRow(ctx, query,
		app.ID,
		app.Name,
		app.Description,
		app.Status,
		app.Visibility,
		app.OwnerUserID,
		startDate,
		dueDate,
	).Scan(&app.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Application{}, ErrNotFound
		}
		return domain.Application{}, err
	}

	// Return the updated object by re-fetching it.
	return s.GetApplication(ctx, app.ID)
}


// ArchiveApplication is the soft-delete entry point (api §13.2 DELETE = archive).
func (s *PostgresStore) ArchiveApplication(ctx context.Context, applicationID, archivedReason string) (domain.Application, error) {
	query := `
		UPDATE applications
		SET 
			status = 'archived',
			archived_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND archived_at IS NULL
		RETURNING updated_at, archived_at
	`
	var updatedAt, archivedAt time.Time
	err := s.pool.QueryRow(ctx, query, applicationID).Scan(&updatedAt, &archivedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Application{}, ErrNotFound
		}
		return domain.Application{}, err
	}

	// TODO: Persist archivedReason somewhere, e.g. in a new audit log event.

	return s.GetApplication(ctx, applicationID)
}

// --- Application-Repository link ---

func (s *PostgresStore) ListApplicationRepositories(ctx context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	query := `
		SELECT application_id, repo_provider, repo_full_name, external_repo_id, role, 
		       sync_status, sync_error_code, sync_error_retryable, sync_error_at, 
			   last_sync_at, linked_at
		FROM application_repositories
		WHERE application_id = $1
		ORDER BY role, repo_full_name
	`
	rows, err := s.pool.Query(ctx, query, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []domain.ApplicationRepository
	for rows.Next() {
		var repo domain.ApplicationRepository
		var externalRepoID, syncErrorCode pgtype.Text
		var syncErrorRetryable pgtype.Bool
		var syncErrorAt, lastSyncAt pgtype.Timestamptz

		err := rows.Scan(
			&repo.ApplicationID,
			&repo.RepoProvider,
			&repo.RepoFullName,
			&externalRepoID,
			&repo.Role,
			&repo.SyncStatus,
			&syncErrorCode,
			&syncErrorRetryable,
			&syncErrorAt,
			&lastSyncAt,
			&repo.LinkedAt,
		)
		if err != nil {
			return nil, err
		}

		repo.ExternalRepoID = externalRepoID.String
		repo.SyncErrorCode = domain.SyncErrorCode(syncErrorCode.String)
		if syncErrorRetryable.Valid {
			repo.SyncErrorRetryable = &syncErrorRetryable.Bool
		}
		if syncErrorAt.Valid {
			repo.SyncErrorAt = &syncErrorAt.Time
		}
		if lastSyncAt.Valid {
			repo.LastSyncAt = &lastSyncAt.Time
		}
		repos = append(repos, repo)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return repos, nil
}


func (s *PostgresStore) CreateApplicationRepository(ctx context.Context, link domain.ApplicationRepository) (domain.ApplicationRepository, error) {
	query := `
		INSERT INTO application_repositories (application_id, repo_provider, repo_full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING linked_at
	`
	err := s.pool.QueryRow(ctx, query,
		link.ApplicationID,
		link.RepoProvider,
		link.RepoFullName,
		link.Role,
	).Scan(&link.LinkedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return domain.ApplicationRepository{}, ErrConflict
			}
		}
		return domain.ApplicationRepository{}, err
	}
	return link, nil
}

func (s *PostgresStore) DeleteApplicationRepository(ctx context.Context, key ApplicationRepositoryLinkKey) error {
	query := `
		DELETE FROM application_repositories
		WHERE application_id = $1 AND repo_provider = $2 AND repo_full_name = $3
	`
	cmdTag, err := s.pool.Exec(ctx, query, key.ApplicationID, key.RepoProvider, key.RepoFullName)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateApplicationRepositorySync 는 webhook/pull worker 가 sync_status / sync_error_code 를
// link 단위로 1건 캐시 갱신할 때 호출 (api §13.3 운영 룰).
func (s *PostgresStore) UpdateApplicationRepositorySync(ctx context.Context, key ApplicationRepositoryLinkKey, status domain.ApplicationRepositorySyncStatus, errorCode domain.SyncErrorCode) error {
	return ErrNotImplemented
}

// --- SCM Provider catalog ---

func (s *PostgresStore) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
	return nil, ErrNotImplemented
}

func (s *PostgresStore) UpdateSCMProvider(ctx context.Context, provider domain.SCMProvider) (domain.SCMProvider, error) {
	return domain.SCMProvider{}, ErrNotImplemented
}

// --- Projects ---

func (s *PostgresStore) ListProjects(ctx context.Context, opts ProjectListOptions) ([]domain.Project, int, error) {
	var whereClauses []string
	var args []interface{}
	argCount := 1

	if !opts.IncludeArchived {
		whereClauses = append(whereClauses, fmt.Sprintf("status <> 'archived'"))
	}
	if opts.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, opts.Status)
		argCount++
	}
	if opts.RepositoryID != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("repository_id = $%d", argCount))
		args = append(args, opts.RepositoryID)
		argCount++
	}

	where := ""
	if len(whereClauses) > 0 {
		where = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM projects " + where
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, application_id, repository_id, key, name, description, status, visibility, owner_user_id, start_date, due_date, archived_at, created_at, updated_at
		FROM projects
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, where, argCount, argCount+1)
	
	limit, offset := boundedList(domain.ListOptions{Limit: opts.Limit, Offset: opts.Offset})
	args = append(args, limit, offset)
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, 0, err
		}
		projects = append(projects, p)
	}

	return projects, total, nil
}


func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (domain.Project, error) {
	query := `
		SELECT id, application_id, repository_id, key, name, description, status, visibility, owner_user_id, start_date, due_date, archived_at, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	row := s.pool.QueryRow(ctx, query, projectID)
	p, err := scanProject(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}
	return p, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	query := `
		INSERT INTO projects (application_id, repository_id, key, name, description, status, visibility, owner_user_id, start_date, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	var appID pgtype.UUID
	if p.ApplicationID != nil {
		appID = pgtype.UUID{Bytes: *p.ApplicationID, Valid: true}
	}
	var startDate, dueDate pgtype.Date
	if p.StartDate != nil {
		startDate = pgtype.Date{Time: *p.StartDate, Valid: true}
	}
	if p.DueDate != nil {
		dueDate = pgtype.Date{Time: *p.DueDate, Valid: true}
	}

	err := s.pool.QueryRow(ctx, query,
		appID,
		p.RepositoryID,
		p.Key,
		p.Name,
		p.Description,
		p.Status,
		p.Visibility,
		p.OwnerUserID,
		startDate,
		dueDate,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return domain.Project{}, ErrConflict
			}
		}
		return domain.Project{}, err
	}

	return p, nil
}


func (s *PostgresStore) UpdateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	query := `
		UPDATE projects
		SET 
			name = $2,
			description = $3,
			status = $4,
			visibility = $5,
			owner_user_id = $6,
			start_date = $7,
			due_date = $8,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	var startDate, dueDate pgtype.Date
	if p.StartDate != nil {
		startDate = pgtype.Date{Time: *p.StartDate, Valid: true}
	}
	if p.DueDate != nil {
		dueDate = pgtype.Date{Time: *p.DueDate, Valid: true}
	}

	err := s.pool.QueryRow(ctx, query,
		p.ID,
		p.Name,
		p.Description,
		p.Status,
		p.Visibility,
		p.OwnerUserID,
		startDate,
		dueDate,
	).Scan(&p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}
	return s.GetProject(ctx, p.ID)
}

func (s *PostgresStore) ArchiveProject(ctx context.Context, projectID, archivedReason string) (domain.Project, error) {
	query := `
		UPDATE projects
		SET 
			status = 'archived',
			archived_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND archived_at IS NULL
	`
	cmdTag, err := s.pool.Exec(ctx, query, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.Project{}, ErrNotFound
	}
	
	// TODO: Persist archivedReason
	
	return s.GetProject(ctx, projectID)
}

func scanProject(row pgx.Row) (domain.Project, error) {
	var p domain.Project
	var appID pgtype.UUID
	var startDate, dueDate pgtype.Date
	var archivedAt pgtype.Timestamptz
	var ownerUserID, description pgtype.Text

	err := row.Scan(
		&p.ID,
		&appID,
		&p.RepositoryID,
		&p.Key,
		&p.Name,
		&description,
		&p.Status,
		&p.Visibility,
		&ownerUserID,
		&startDate,
		&dueDate,
		&archivedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return domain.Project{}, err
	}

	if appID.Valid {
		p.ApplicationID = &appID.Bytes
	}
	p.OwnerUserID = ownerUserID.String
	p.Description = description.String
	if startDate.Valid {
		p.StartDate = &startDate.Time
	}
	if dueDate.Valid {
		p.DueDate = &dueDate.Time
	}
	if archivedAt.Valid {
		p.ArchivedAt = &archivedAt.Time
	}

	return p, nil
}
