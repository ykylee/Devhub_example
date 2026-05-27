package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

func nullableUUIDArg(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

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

// ProjectListOptions parameterizes ListProjects.
type ProjectListOptions struct {
	RepositoryID    int64
	ApplicationID   string
	Status          string
	IncludeArchived bool
	Limit           int
	Offset          int
}

// --- Applications ---

const applicationsSelectColumns = `
	id::text,
	key,
	name,
	COALESCE(description, ''),
	status,
	visibility,
	COALESCE(owner_user_id, ''),
	COALESCE(leader_user_id, ''),
	COALESCE(development_unit_id, ''),
	start_date,
	due_date,
	archived_at,
	created_at,
	updated_at`

// applicationsSearchPredicate은 ListApplications의 `q` 파라미터($3)를 키/이름/오너/리더/
// 부서 + 부서 라벨 + 연결된 repository · project 로 매칭한다. count/list 쿼리에서 공유.
const applicationsSearchPredicate = `
    $3 = ''
    OR key ILIKE '%' || $3 || '%'
    OR name ILIKE '%' || $3 || '%'
    OR owner_user_id ILIKE '%' || $3 || '%'
    OR leader_user_id ILIKE '%' || $3 || '%'
    OR development_unit_id ILIKE '%' || $3 || '%'
    OR EXISTS (
      SELECT 1 FROM org_units ou
      WHERE ou.unit_id = applications.development_unit_id
        AND ou.label ILIKE '%' || $3 || '%'
    )
    OR EXISTS (
      SELECT 1 FROM application_repositories ar
      WHERE ar.application_id = applications.id
        AND (
          ar.repo_full_name ILIKE '%' || $3 || '%'
          OR ar.external_repo_id ILIKE '%' || $3 || '%'
        )
    )
    OR EXISTS (
      SELECT 1 FROM projects p
      WHERE p.application_id = applications.id
        AND (
          p.key ILIKE '%' || $3 || '%'
          OR p.name ILIKE '%' || $3 || '%'
        )
    )
`

func scanApplication(row pgx.Row) (domain.Application, error) {
	var (
		app                domain.Application
		startDate, dueDate *time.Time
		archivedAt         *time.Time
	)
	if err := row.Scan(
		&app.ID,
		&app.Key,
		&app.Name,
		&app.Description,
		&app.Status,
		&app.Visibility,
		&app.OwnerUserID,
		&app.LeaderUserID,
		&app.DevelopmentUnitID,
		&startDate,
		&dueDate,
		&archivedAt,
		&app.CreatedAt,
		&app.UpdatedAt,
	); err != nil {
		return domain.Application{}, err
	}
	app.StartDate = startDate
	app.DueDate = dueDate
	app.ArchivedAt = archivedAt
	return app, nil
}

func (s *PostgresStore) ListApplications(ctx context.Context, opts ApplicationListOptions) ([]domain.Application, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const countQuery = `
SELECT COUNT(*) FROM applications
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + applicationsSearchPredicate + `)`

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, opts.Status, opts.IncludeArchived, opts.Query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	const listQuery = `
SELECT` + applicationsSelectColumns + `
FROM applications
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + applicationsSearchPredicate + `)
ORDER BY key ASC
LIMIT $4 OFFSET $5`

	rows, err := s.pool.Query(ctx, listQuery, opts.Status, opts.IncludeArchived, opts.Query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	apps := make([]domain.Application, 0, limit)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan application: %w", err)
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate applications: %w", err)
	}
	return apps, total, nil
}

func (s *PostgresStore) GetApplication(ctx context.Context, applicationID string) (domain.Application, error) {
	query := `SELECT` + applicationsSelectColumns + ` FROM applications WHERE id = $1::uuid`
	row := s.pool.QueryRow(ctx, query, applicationID)
	app, err := scanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("get application: %w", err)
	}
	return app, nil
}

// GetApplicationByKey lookup by user-facing key. Useful for create-time conflict checks
// and admin tools.
func (s *PostgresStore) GetApplicationByKey(ctx context.Context, key string) (domain.Application, error) {
	query := `SELECT` + applicationsSelectColumns + ` FROM applications WHERE key = $1`
	row := s.pool.QueryRow(ctx, query, key)
	app, err := scanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("get application by key: %w", err)
	}
	return app, nil
}

// applicationsInsertQuery is the canonical INSERT used by CreateApplication and by
// the DREQ promote transaction (dev_requests_promote.go). Sharing the query keeps the
// archived_at consistency CHECK invariant identical across entry points.
//
// PR #110 integration test fail #1 정정 (sprint claude/work_260514-f) —
// status='archived' 로 직접 생성 시 archived_at 자동 채움. CHECK
// applications_archived_consistency 위반 회피. UpdateApplication 의 archived_at
// 자동 갱신 패턴 (CASE WHEN status='archived') 과 대칭.
const applicationsInsertQuery = `
INSERT INTO applications (key, name, description, status, visibility, owner_user_id, leader_user_id, development_unit_id, start_date, due_date, archived_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10,
        CASE WHEN $4 = 'archived' THEN NOW() ELSE NULL END)
RETURNING` + applicationsSelectColumns

func (s *PostgresStore) CreateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	row := s.pool.QueryRow(ctx, applicationsInsertQuery,
		app.Key, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	created, err := scanApplication(row)
	if isUniqueViolation(err) {
		return domain.Application{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.Application{}, ErrConflict
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("create application: %w", err)
	}
	return created, nil
}

// UpdateApplication mutates allowed fields (name/description/owner/dates/visibility/status).
// `key` 는 immutable 이라 호출자가 별도 검증 (PATCH handler) — store 는 단순 UPDATE.
// archived consistency CHECK 위반 회피 위해 status=archived 전이 시 archived_at = NOW 자동 설정,
// 기타 status 전이 시 archived_at = NULL 로 재설정.
func (s *PostgresStore) UpdateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	const updateQuery = `
UPDATE applications SET
	name = $2,
	description = NULLIF($3, ''),
	status = $4,
	visibility = $5,
	owner_user_id = NULLIF($6, ''),
	leader_user_id = NULLIF($7, ''),
	development_unit_id = NULLIF($8, ''),
	start_date = $9,
	due_date = $10,
	archived_at = CASE WHEN $4 = 'archived' THEN COALESCE(archived_at, NOW()) ELSE NULL END,
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + applicationsSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery,
		app.ID, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	updated, err := scanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, ErrNotFound
	}
	if isForeignKeyViolation(err) {
		return domain.Application{}, ErrConflict
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("update application: %w", err)
	}
	return updated, nil
}

// ArchiveApplication is the soft-delete entry point (api §13.2 DELETE = archive).
// Sets status='archived' + archived_at=NOW. archived_reason 은 audit_logs payload 에 기록.
func (s *PostgresStore) ArchiveApplication(ctx context.Context, applicationID, archivedReason string) (domain.Application, error) {
	const archiveQuery = `
UPDATE applications SET
	status = 'archived',
	archived_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + applicationsSelectColumns

	row := s.pool.QueryRow(ctx, archiveQuery, applicationID)
	archived, err := scanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("archive application: %w", err)
	}
	_ = archivedReason // audit 기록은 handler 책임
	return archived, nil
}

// CountActiveApplicationRepositories — 상태 전이 가드 검증용 (planning→active 의 활성 repo ≥1).
func (s *PostgresStore) CountActiveApplicationRepositories(ctx context.Context, applicationID string) (int, error) {
	const query = `
SELECT COUNT(*) FROM application_repositories
WHERE application_id = $1::uuid AND sync_status = 'active'`
	var count int
	if err := s.pool.QueryRow(ctx, query, applicationID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active application repositories: %w", err)
	}
	return count, nil
}

// --- Application-Repository link ---

const applicationRepositoriesSelectColumns = `
	application_id::text,
	repo_provider,
	repo_full_name,
	COALESCE(external_repo_id, ''),
	role,
	sync_status,
	COALESCE(sync_error_code, ''),
	sync_error_retryable,
	sync_error_at,
	last_sync_at,
	linked_at`

func scanApplicationRepository(row pgx.Row) (domain.ApplicationRepository, error) {
	var (
		link        domain.ApplicationRepository
		syncErrCode string
		retryable   *bool
		syncErrAt   *time.Time
		lastSyncAt  *time.Time
	)
	if err := row.Scan(
		&link.ApplicationID,
		&link.RepoProvider,
		&link.RepoFullName,
		&link.ExternalRepoID,
		&link.Role,
		&link.SyncStatus,
		&syncErrCode,
		&retryable,
		&syncErrAt,
		&lastSyncAt,
		&link.LinkedAt,
	); err != nil {
		return domain.ApplicationRepository{}, err
	}
	link.SyncErrorCode = domain.SyncErrorCode(syncErrCode)
	link.SyncErrorRetryable = retryable
	link.SyncErrorAt = syncErrAt
	link.LastSyncAt = lastSyncAt
	return link, nil
}

func (s *PostgresStore) ListApplicationRepositories(ctx context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	query := `SELECT` + applicationRepositoriesSelectColumns + `
FROM application_repositories
WHERE application_id = $1::uuid
ORDER BY repo_provider ASC, repo_full_name ASC`

	rows, err := s.pool.Query(ctx, query, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list application repositories: %w", err)
	}
	defer rows.Close()

	links := make([]domain.ApplicationRepository, 0, 4)
	for rows.Next() {
		link, err := scanApplicationRepository(rows)
		if err != nil {
			return nil, fmt.Errorf("scan application repository: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate application repositories: %w", err)
	}
	return links, nil
}

// applicationRepositoriesInsertQuery is the canonical INSERT used by
// CreateApplicationRepository and by the DREQ promote transaction
// (dev_requests_promote.go) when a primary repository link is supplied alongside a
// new Application. Sharing it keeps the composite PK + sync_status default in sync.
const applicationRepositoriesInsertQuery = `
INSERT INTO application_repositories (
	application_id, repo_provider, repo_full_name, external_repo_id, role, sync_status
) VALUES (
	$1::uuid, $2, $3, NULLIF($4, ''), $5, COALESCE(NULLIF($6, ''), 'requested')
)
RETURNING` + applicationRepositoriesSelectColumns

func (s *PostgresStore) CreateApplicationRepository(ctx context.Context, link domain.ApplicationRepository) (domain.ApplicationRepository, error) {
	syncStatus := string(link.SyncStatus)
	row := s.pool.QueryRow(ctx, applicationRepositoriesInsertQuery,
		link.ApplicationID, link.RepoProvider, link.RepoFullName,
		link.ExternalRepoID, link.Role, syncStatus,
	)
	created, err := scanApplicationRepository(row)
	if isUniqueViolation(err) {
		return domain.ApplicationRepository{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.ApplicationRepository{}, ErrConflict
	}
	if err != nil {
		return domain.ApplicationRepository{}, fmt.Errorf("create application repository: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) DeleteApplicationRepository(ctx context.Context, key ApplicationRepositoryLinkKey) error {
	const deleteQuery = `
DELETE FROM application_repositories
WHERE application_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	tag, err := s.pool.Exec(ctx, deleteQuery, key.ApplicationID, key.RepoProvider, key.RepoFullName)
	if err != nil {
		return fmt.Errorf("delete application repository: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateApplicationRepositorySync — link 단위 sync_status / sync_error_code / retryable / at
// 일관 갱신. errorCode 가 빈 문자열이면 정상 상태 (retryable / at 도 NULL 로 reset).
//
// PR #110 integration test fail #2 정정 (sprint claude/work_260514-f) — CASE 의 NULL
// 분기 type 을 명시 cast 추가. PG type inference 가 NULL 의 default type 을 text 로
// 추론 → boolean / timestamptz 컬럼과 mismatch (SQLSTATE 42804). NULL::boolean /
// NULL::timestamptz 로 명시.
func (s *PostgresStore) UpdateApplicationRepositorySync(ctx context.Context, key ApplicationRepositoryLinkKey, status domain.ApplicationRepositorySyncStatus, errorCode domain.SyncErrorCode) error {
	const updateQuery = `
UPDATE application_repositories SET
	sync_status = $4,
	sync_error_code = NULLIF($5, ''),
	sync_error_retryable = CASE WHEN $5 = '' THEN NULL::boolean ELSE $6 END,
	sync_error_at = CASE WHEN $5 = '' THEN NULL::timestamptz ELSE NOW() END,
	last_sync_at = NOW()
WHERE application_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	retryable := domain.IsRetryableSyncError(errorCode)
	tag, err := s.pool.Exec(ctx, updateQuery,
		key.ApplicationID, key.RepoProvider, key.RepoFullName,
		string(status), string(errorCode), retryable,
	)
	if err != nil {
		return fmt.Errorf("update application repository sync: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- SCM Provider catalog ---

const scmProvidersSelectColumns = `
	provider_key,
	display_name,
	enabled,
	adapter_version,
	created_at,
	updated_at`

func scanSCMProvider(row pgx.Row) (domain.SCMProvider, error) {
	var p domain.SCMProvider
	if err := row.Scan(
		&p.ProviderKey,
		&p.DisplayName,
		&p.Enabled,
		&p.AdapterVersion,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return domain.SCMProvider{}, err
	}
	return p, nil
}

func (s *PostgresStore) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
	const query = `SELECT` + scmProvidersSelectColumns + `
FROM scm_providers
ORDER BY provider_key ASC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list scm providers: %w", err)
	}
	defer rows.Close()

	providers := make([]domain.SCMProvider, 0, 4)
	for rows.Next() {
		p, err := scanSCMProvider(rows)
		if err != nil {
			return nil, fmt.Errorf("scan scm provider: %w", err)
		}
		providers = append(providers, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scm providers: %w", err)
	}
	return providers, nil
}

// UpdateSCMProvider 는 enabled / display_name 만 갱신. adapter_version 은 배포 파이프라인
// 외 수정 불가 (api §13.1.1) — store 단에서 컬럼 자체를 UPDATE 에 포함하지 않음.
func (s *PostgresStore) UpdateSCMProvider(ctx context.Context, provider domain.SCMProvider) (domain.SCMProvider, error) {
	const updateQuery = `
UPDATE scm_providers SET
	display_name = $2,
	enabled = $3,
	updated_at = NOW()
WHERE provider_key = $1
RETURNING` + scmProvidersSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery,
		provider.ProviderKey, provider.DisplayName, provider.Enabled,
	)
	updated, err := scanSCMProvider(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SCMProvider{}, ErrNotFound
	}
	if err != nil {
		return domain.SCMProvider{}, fmt.Errorf("update scm provider: %w", err)
	}
	return updated, nil
}

// --- Projects ---

const projectsSelectColumns = `
	id::text,
	COALESCE(application_id::text, ''),
	COALESCE(repository_id, 0),
	key,
	name,
	COALESCE(description, ''),
	status,
	visibility,
	COALESCE(owner_user_id, ''),
	start_date,
	due_date,
	archived_at,
	created_at,
	updated_at`

func scanProject(row pgx.Row) (domain.Project, error) {
	var (
		p                  domain.Project
		startDate, dueDate *time.Time
		archivedAt         *time.Time
	)
	if err := row.Scan(
		&p.ID,
		&p.ApplicationID,
		&p.RepositoryID,
		&p.Key,
		&p.Name,
		&p.Description,
		&p.Status,
		&p.Visibility,
		&p.OwnerUserID,
		&startDate,
		&dueDate,
		&archivedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return domain.Project{}, err
	}
	p.StartDate = startDate
	p.DueDate = dueDate
	p.ArchivedAt = archivedAt
	return p, nil
}

func (s *PostgresStore) ListProjects(ctx context.Context, opts ProjectListOptions) ([]domain.Project, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const countQuery = `
SELECT COUNT(*) FROM projects
WHERE ($1::bigint = 0 OR repository_id = $1)
  AND ($2::uuid IS NULL OR application_id = $2::uuid)
  AND ($3 = '' OR status = $3)
  AND ($4 OR status <> 'archived')`

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, opts.RepositoryID, nullableUUIDArg(opts.ApplicationID), opts.Status, opts.IncludeArchived).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	query := `
SELECT` + projectsSelectColumns + `
FROM projects
WHERE ($3::bigint = 0 OR repository_id = $3)
  AND ($4::uuid IS NULL OR application_id = $4::uuid)
  AND ($5 = '' OR status = $5)
  AND ($6 OR status <> 'archived')
ORDER BY key ASC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.RepositoryID, nullableUUIDArg(opts.ApplicationID), opts.Status, opts.IncludeArchived)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0, limit)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, total, nil
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (domain.Project, error) {
	query := `SELECT` + projectsSelectColumns + ` FROM projects WHERE id = $1::uuid`
	row := s.pool.QueryRow(ctx, query, projectID)
	p, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// projectsInsertQuery is the canonical INSERT used by CreateProject and by the DREQ
// promote transaction (dev_requests_promote.go). Sharing it keeps the (repository_id,
// key) UNIQUE constraint and NULLIF semantics identical across entry points.
const projectsInsertQuery = `
INSERT INTO projects (
	application_id, repository_id, key, name, description, status, visibility,
	owner_user_id, start_date, due_date
) VALUES (
	NULLIF($1, '')::uuid, NULLIF($2::bigint, 0), $3, $4, NULLIF($5, ''), $6, $7,
	NULLIF($8, ''), $9, $10
)
RETURNING` + projectsSelectColumns

func (s *PostgresStore) CreateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	row := s.pool.QueryRow(ctx, projectsInsertQuery,
		project.ApplicationID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	created, err := scanProject(row)
	if isUniqueViolation(err) {
		return domain.Project{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.Project{}, ErrConflict
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}
	return created, nil
}

// RepositoryCreatePayload — project 생성 시 동반 생성할 repository 입력. project tx
// 안에서 함께 생성하기 위해 store 로 전달 (codex #349 P2 atomicity).
type RepositoryCreatePayload struct {
	Key         string
	Slug        string
	SCMProvider string
}

// CreateProjectWithRepositoryPayload creates the project — optionally creating and
// linking a companion repository — in ONE transaction (codex #349 P2 atomicity).
// repoPayload 가 주어지면 repository upsert + project insert + project_repositories
// link 가 모두 같은 tx 에서 일어나므로, project insert 실패(예: 중복 key) 시 repository
// 생성도 rollback 되어 "create and link" 가 원자적이다 (고아 repository 없음).
func (s *PostgresStore) CreateProjectWithRepositoryPayload(ctx context.Context, project domain.Project, repositoryIDs []int64, repoPayload *RepositoryCreatePayload) (domain.Project, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Project{}, fmt.Errorf("begin create project tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if repoPayload != nil {
		repoID, repoErr := createRepositoryTx(ctx, tx, repoPayload.Key, repoPayload.Slug, repoPayload.SCMProvider)
		if isUniqueViolation(repoErr) || isForeignKeyViolation(repoErr) || isCheckViolation(repoErr, "") {
			return domain.Project{}, ErrConflict
		}
		if repoErr != nil {
			return domain.Project{}, fmt.Errorf("create companion repository (tx): %w", repoErr)
		}
		if project.RepositoryID == 0 {
			project.RepositoryID = repoID
		}
		repositoryIDs = append(repositoryIDs, repoID)
	}

	row := tx.QueryRow(ctx, projectsInsertQuery,
		project.ApplicationID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	created, err := scanProject(row)
	if isUniqueViolation(err) || isForeignKeyViolation(err) {
		return domain.Project{}, ErrConflict
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project (tx): %w", err)
	}

	if len(repositoryIDs) == 0 && project.RepositoryID > 0 {
		repositoryIDs = []int64{project.RepositoryID}
	}
	seen := map[int64]struct{}{}
	for _, rid := range repositoryIDs {
		if rid <= 0 {
			continue
		}
		if _, ok := seen[rid]; ok {
			continue
		}
		seen[rid] = struct{}{}
		role := "linked"
		if rid == project.RepositoryID {
			role = "primary"
		}
		const insertLink = `
INSERT INTO project_repositories (project_id, repository_id, role)
VALUES ($1::uuid, $2, $3)`
		if _, err := tx.Exec(ctx, insertLink, created.ID, rid, role); err != nil {
			if isUniqueViolation(err) || isForeignKeyViolation(err) || isCheckViolation(err, "") {
				return domain.Project{}, ErrConflict
			}
			return domain.Project{}, fmt.Errorf("create project repository link (tx): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Project{}, fmt.Errorf("commit create project tx: %w", err)
	}
	return created, nil
}

// createRepositoryTx upserts a companion repository within the given transaction
// (codex #349 P2 atomicity — 이전 CreateRepositoryForProject 를 tx-aware 로 전환).
// full_name(slug) 충돌 시 upsert. raw error 를 반환 — caller 가 violation 분류 +
// rollback 처리. clone_url/html_url 은 scm+<provider>:// placeholder (실제 SCM
// 프로비저닝은 범위 밖, 내부 repositories row 만 생성).
func createRepositoryTx(ctx context.Context, tx pgx.Tx, key, slug, scmProvider string) (int64, error) {
	fullName := strings.TrimSpace(slug)
	name := strings.TrimSpace(key)
	if name == "" {
		name = fullName
	}
	// source='system' — 시스템에서 생성한 repository (SCM mirror 가 아님). 소유권 분리
	// (migration 000042) — codex #363 P2. 빈 Source 를 scm 으로 기본 취급하던 것 정정.
	const query = `
INSERT INTO repositories (
	full_name, name, owner_login, clone_url, html_url, default_branch, private, source, updated_at
) VALUES (
	$1, $2, NULLIF(split_part($1, '/', 1), ''), NULLIF($3, ''), NULLIF($4, ''), 'main', false, 'system', NOW()
)
ON CONFLICT (full_name) DO UPDATE SET
	name = EXCLUDED.name,
	updated_at = NOW()
RETURNING id`

	cloneURL := ""
	htmlURL := ""
	if strings.TrimSpace(scmProvider) != "" {
		cloneURL = "scm+" + strings.TrimSpace(scmProvider) + "://" + fullName + ".git"
		htmlURL = "scm+" + strings.TrimSpace(scmProvider) + "://" + fullName
	}

	var id int64
	err := tx.QueryRow(ctx, query, fullName, name, cloneURL, htmlURL).Scan(&id)
	return id, err
}

func (s *PostgresStore) UpdateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	const updateQuery = `
UPDATE projects SET
	name = $2,
	description = NULLIF($3, ''),
	status = $4,
	visibility = $5,
	owner_user_id = NULLIF($6, ''),
	start_date = $7,
	due_date = $8,
	archived_at = CASE WHEN $4 = 'archived' THEN COALESCE(archived_at, NOW()) ELSE NULL END,
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + projectsSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery,
		project.ID, project.Name, project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	updated, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project: %w", err)
	}
	return updated, nil
}

func (s *PostgresStore) ArchiveProject(ctx context.Context, projectID, archivedReason string) (domain.Project, error) {
	const archiveQuery = `
UPDATE projects SET
	status = 'archived',
	archived_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + projectsSelectColumns

	row := s.pool.QueryRow(ctx, archiveQuery, projectID)
	archived, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("archive project: %w", err)
	}
	_ = archivedReason
	return archived, nil
}

func (s *PostgresStore) ListProjectRepositories(ctx context.Context, projectID string) ([]domain.ProjectRepository, error) {
	const query = `
SELECT project_id::text, repository_id, role, linked_at
FROM project_repositories
WHERE project_id = $1::uuid
ORDER BY repository_id ASC`

	rows, err := s.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project repositories: %w", err)
	}
	defer rows.Close()

	links := make([]domain.ProjectRepository, 0)
	for rows.Next() {
		var link domain.ProjectRepository
		if err := rows.Scan(&link.ProjectID, &link.RepositoryID, &link.Role, &link.LinkedAt); err != nil {
			return nil, fmt.Errorf("scan project repository: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project repositories: %w", err)
	}
	return links, nil
}

func (s *PostgresStore) CreateProjectRepository(ctx context.Context, link domain.ProjectRepository) (domain.ProjectRepository, error) {
	const query = `
INSERT INTO project_repositories (project_id, repository_id, role)
VALUES ($1::uuid, $2, $3)
RETURNING project_id::text, repository_id, role, linked_at`

	row := s.pool.QueryRow(ctx, query, link.ProjectID, link.RepositoryID, link.Role)
	var created domain.ProjectRepository
	if err := row.Scan(&created.ProjectID, &created.RepositoryID, &created.Role, &created.LinkedAt); err != nil {
		if isUniqueViolation(err) || isForeignKeyViolation(err) || isCheckViolation(err, "") {
			return domain.ProjectRepository{}, ErrConflict
		}
		return domain.ProjectRepository{}, fmt.Errorf("create project repository: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) DeleteProjectRepository(ctx context.Context, projectID string, repositoryID int64) error {
	const query = `DELETE FROM project_repositories WHERE project_id = $1::uuid AND repository_id = $2`
	cmd, err := s.pool.Exec(ctx, query, projectID, repositoryID)
	if err != nil {
		return fmt.Errorf("delete project repository: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
