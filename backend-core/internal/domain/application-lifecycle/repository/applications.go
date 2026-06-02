package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
)

func nullableUUIDArg(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

// ApplicationListOptions parameterizes ListApplications.
type ApplicationListOptions = store.ApplicationListOptions

// ApplicationRepositoryLinkKey identifies a single link row (composite PK).
type ApplicationRepositoryLinkKey = store.ApplicationRepositoryLinkKey

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

func ScanApplication(row pgx.Row) (domain.Application, error) {
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

func (r *ApplicationRepository) ListApplications(ctx context.Context, opts ApplicationListOptions) ([]domain.Application, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	rowFilter := `
  AND ($6 = '' OR $6 = 'system_admin'
       OR owner_user_id = $7 OR leader_user_id = $7
       OR EXISTS (SELECT 1 FROM projects WHERE application_id = applications.id
                  AND (owner_user_id = $7 OR EXISTS (SELECT 1 FROM project_members WHERE project_id = projects.id AND user_id = $7)))
       OR (array_length($8::text[], 1) > 0 AND development_unit_id = ANY($8))
       OR (array_length($9::text[], 1) > 0 AND development_unit_id = ANY($9)))`

	var countQuery = `
SELECT COUNT(*) FROM applications
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + applicationsSearchPredicate + `)` + rowFilter

	var total int
	if err := r.store.Pool().QueryRow(ctx, countQuery, opts.Status, opts.IncludeArchived, opts.Query, 0, 0, opts.ActorRole, opts.ActorLogin, opts.OrgUnitIDs, opts.PrimaryUnitIDs).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	var listQuery = `
SELECT` + applicationsSelectColumns + `
FROM applications
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + applicationsSearchPredicate + `)` + rowFilter + `
ORDER BY key ASC
LIMIT $4 OFFSET $5`

	rows, err := r.store.Pool().Query(ctx, listQuery, opts.Status, opts.IncludeArchived, opts.Query, limit, offset, opts.ActorRole, opts.ActorLogin, opts.OrgUnitIDs, opts.PrimaryUnitIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	apps := make([]domain.Application, 0, limit)
	for rows.Next() {
		app, err := ScanApplication(rows)
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

func (r *ApplicationRepository) GetApplication(ctx context.Context, applicationID string) (domain.Application, error) {
	query := `SELECT` + applicationsSelectColumns + ` FROM applications WHERE id = $1::uuid`
	row := r.store.Pool().QueryRow(ctx, query, applicationID)
	app, err := ScanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("get application: %w", err)
	}
	return app, nil
}

// GetApplicationByKey lookup by user-facing key. Useful for create-time conflict checks
// and admin tools.
func (r *ApplicationRepository) GetApplicationByKey(ctx context.Context, key string) (domain.Application, error) {
	query := `SELECT` + applicationsSelectColumns + ` FROM applications WHERE key = $1`
	row := r.store.Pool().QueryRow(ctx, query, key)
	app, err := ScanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("get application by key: %w", err)
	}
	return app, nil
}

// ApplicationsInsertQuery is the canonical INSERT used by CreateApplication and by
// the DREQ promote transaction (dev_requests_promote.go). Sharing the query keeps the
// archived_at consistency CHECK invariant identical across entry points.
const ApplicationsInsertQuery = `
INSERT INTO applications (key, name, description, status, visibility, owner_user_id, leader_user_id, development_unit_id, start_date, due_date, archived_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10,
        CASE WHEN $4 = 'archived' THEN NOW() ELSE NULL END)
RETURNING` + applicationsSelectColumns

func (r *ApplicationRepository) CreateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	row := r.store.Pool().QueryRow(ctx, ApplicationsInsertQuery,
		app.Key, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	created, err := ScanApplication(row)
	if store.IsUniqueViolation(err) {
		return domain.Application{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.Application{}, store.ErrConflict
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
func (r *ApplicationRepository) UpdateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
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

	row := r.store.Pool().QueryRow(ctx, updateQuery,
		app.ID, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	updated, err := ScanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, store.ErrNotFound
	}
	if store.IsForeignKeyViolation(err) {
		return domain.Application{}, store.ErrConflict
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("update application: %w", err)
	}
	return updated, nil
}

// ArchiveApplication is the soft-delete entry point (api §13.2 DELETE = archive).
// Sets status='archived' + archived_at=NOW. archived_reason 은 audit_logs payload 에 기록.
func (r *ApplicationRepository) ArchiveApplication(ctx context.Context, applicationID, archivedReason string) (domain.Application, error) {
	const archiveQuery = `
UPDATE applications SET
	status = 'archived',
	archived_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + applicationsSelectColumns

	row := r.store.Pool().QueryRow(ctx, archiveQuery, applicationID)
	archived, err := ScanApplication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("archive application: %w", err)
	}
	_ = archivedReason // audit 기록은 handler 책임
	return archived, nil
}

// DeleteApplication — hard-delete. archived 가드는 handler 책임. cascade:
// application_repositories ON DELETE CASCADE / projects.application_id ON DELETE SET NULL
// (migration 000013/000014/000015). handler 가 archived 상태 검증 후에만 호출.
func (r *ApplicationRepository) DeleteApplication(ctx context.Context, applicationID string) error {
	const query = `DELETE FROM applications WHERE id = $1::uuid`
	cmd, err := r.store.Pool().Exec(ctx, query, applicationID)
	if err != nil {
		return fmt.Errorf("delete application: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// CountActiveApplicationRepositories — 상태 전이 가드 검증용 (planning→active 의 활성 repo ≥1).
// 직접 link (application_repositories.sync_status='active') + 프로젝트 경유 간접 link
// (project_repositories 는 sync 상태 컬럼이 없으므로 link 존재 = 항상 active 로 간주).
// migration 000034 의 project_repositories 컬럼은 (project_id, repository_id BIGINT, role,
// linked_at) 만 존재 → repositories 테이블 JOIN 으로 repo_provider/full_name 매핑.
func (r *ApplicationRepository) CountActiveApplicationRepositories(ctx context.Context, applicationID string) (int, error) {
	const query = `
SELECT COUNT(*) FROM (
	SELECT repo_provider, repo_full_name FROM application_repositories
	WHERE application_id = $1::uuid AND sync_status = 'active'
	UNION
	SELECT ip.provider_key AS repo_provider, r.full_name AS repo_full_name
	FROM project_repositories pr
	JOIN projects p              ON p.id  = pr.project_id
	JOIN repositories r          ON r.id  = pr.repository_id
	JOIN integration_providers ip ON ip.provider_id = r.provider_id
	WHERE p.application_id = $1::uuid
) active_repos`
	var count int
	if err := r.store.Pool().QueryRow(ctx, query, applicationID).Scan(&count); err != nil {
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
	linked_at,
	'direct'::text AS link_source`

func ScanApplicationRepository(row pgx.Row) (domain.ApplicationRepository, error) {
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
		&link.LinkSource,
	); err != nil {
		return domain.ApplicationRepository{}, err
	}
	link.SyncErrorCode = domain.SyncErrorCode(syncErrCode)
	link.SyncErrorRetryable = retryable
	link.SyncErrorAt = syncErrAt
	link.LastSyncAt = lastSyncAt
	return link, nil
}

func (r *ApplicationRepository) ListApplicationRepositories(ctx context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	query := `SELECT ` + applicationRepositoriesSelectColumns + `
FROM application_repositories
WHERE application_id = $1::uuid
UNION
SELECT
	$1::text                                    AS application_id,
	ip.provider_key                             AS repo_provider,
	r.full_name                                 AS repo_full_name,
	COALESCE(r.gitea_repository_id::text, '')   AS external_repo_id,
	CASE pr.role
	  WHEN 'primary' THEN 'primary'
	  WHEN 'shared'  THEN 'shared'
	  ELSE 'sub'
	END                                         AS role,
	'active'::text                              AS sync_status,
	''::text                                    AS sync_error_code,
	NULL::boolean                               AS sync_error_retryable,
	NULL::timestamptz                           AS sync_error_at,
	pr.linked_at                                AS last_sync_at,
	pr.linked_at                                AS linked_at,
	'via_project'::text                         AS link_source
FROM project_repositories pr
JOIN projects p              ON p.id = pr.project_id
JOIN repositories r          ON r.id = pr.repository_id
JOIN integration_providers ip ON ip.provider_id = r.provider_id
WHERE p.application_id = $1::uuid
ORDER BY repo_provider ASC, repo_full_name ASC`

	rows, err := r.store.Pool().Query(ctx, query, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list application repositories: %w", err)
	}
	defer rows.Close()

	links := make([]domain.ApplicationRepository, 0, 4)
	for rows.Next() {
		link, err := ScanApplicationRepository(rows)
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

const ApplicationRepositoriesInsertQuery = `
INSERT INTO application_repositories (
	application_id, repo_provider, repo_full_name, external_repo_id, role, sync_status
) VALUES (
	$1::uuid, $2, $3, NULLIF($4, ''), $5, COALESCE(NULLIF($6, ''), 'requested')
)
RETURNING` + applicationRepositoriesSelectColumns

func (r *ApplicationRepository) CreateApplicationRepository(ctx context.Context, link domain.ApplicationRepository) (domain.ApplicationRepository, error) {
	syncStatus := string(link.SyncStatus)
	row := r.store.Pool().QueryRow(ctx, ApplicationRepositoriesInsertQuery,
		link.ApplicationID, link.RepoProvider, link.RepoFullName,
		link.ExternalRepoID, link.Role, syncStatus,
	)
	created, err := ScanApplicationRepository(row)
	if store.IsUniqueViolation(err) {
		return domain.ApplicationRepository{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.ApplicationRepository{}, store.ErrConflict
	}
	if err != nil {
		return domain.ApplicationRepository{}, fmt.Errorf("create application repository: %w", err)
	}
	return created, nil
}

func (r *ApplicationRepository) DeleteApplicationRepository(ctx context.Context, key ApplicationRepositoryLinkKey) error {
	const deleteQuery = `
DELETE FROM application_repositories
WHERE application_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	tag, err := r.store.Pool().Exec(ctx, deleteQuery, key.ApplicationID, key.RepoProvider, key.RepoFullName)
	if err != nil {
		return fmt.Errorf("delete application repository: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *ApplicationRepository) UpdateApplicationRepositorySync(ctx context.Context, key ApplicationRepositoryLinkKey, status domain.ApplicationRepositorySyncStatus, errorCode domain.SyncErrorCode) error {
	const updateQuery = `
UPDATE application_repositories SET
	sync_status = $4,
	sync_error_code = NULLIF($5, ''),
	sync_error_retryable = CASE WHEN $5 = '' THEN NULL::boolean ELSE $6 END,
	sync_error_at = CASE WHEN $5 = '' THEN NULL::timestamptz ELSE NOW() END,
	last_sync_at = NOW()
WHERE application_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	retryable := domain.IsRetryableSyncError(errorCode)
	tag, err := r.store.Pool().Exec(ctx, updateQuery,
		key.ApplicationID, key.RepoProvider, key.RepoFullName,
		string(status), string(errorCode), retryable,
	)
	if err != nil {
		return fmt.Errorf("update application repository sync: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
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

func (r *ApplicationRepository) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
	const query = `SELECT` + scmProvidersSelectColumns + `
FROM scm_providers
ORDER BY provider_key ASC`

	rows, err := r.store.Pool().Query(ctx, query)
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

func (r *ApplicationRepository) UpdateSCMProvider(ctx context.Context, provider domain.SCMProvider) (domain.SCMProvider, error) {
	const updateQuery = `
UPDATE scm_providers SET
	display_name = $2,
	enabled = $3,
	updated_at = NOW()
WHERE provider_key = $1
RETURNING` + scmProvidersSelectColumns

	row := r.store.Pool().QueryRow(ctx, updateQuery,
		provider.ProviderKey, provider.DisplayName, provider.Enabled,
	)
	updated, err := scanSCMProvider(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SCMProvider{}, store.ErrNotFound
	}
	if err != nil {
		return domain.SCMProvider{}, fmt.Errorf("update scm provider: %w", err)
	}
	return updated, nil
}
