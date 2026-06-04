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

// PlatformListOptions parameterizes ListPlatforms.
type PlatformListOptions = store.PlatformListOptions

// PlatformRepositoryLinkKey identifies a single link row (composite PK).
type PlatformRepositoryLinkKey = store.PlatformRepositoryLinkKey

// --- Platforms ---

const platformsSelectColumns = `
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
const platformsSearchPredicate = `
    $3 = ''
    OR key ILIKE '%' || $3 || '%'
    OR name ILIKE '%' || $3 || '%'
    OR owner_user_id ILIKE '%' || $3 || '%'
    OR leader_user_id ILIKE '%' || $3 || '%'
    OR development_unit_id ILIKE '%' || $3 || '%'
    OR EXISTS (
      SELECT 1 FROM org_units ou
      WHERE ou.unit_id = platforms.development_unit_id
        AND ou.label ILIKE '%' || $3 || '%'
    )
    OR EXISTS (
      SELECT 1 FROM platform_repositories ar
      WHERE ar.platform_id = platforms.id
        AND (
          ar.repo_full_name ILIKE '%' || $3 || '%'
          OR ar.external_repo_id ILIKE '%' || $3 || '%'
        )
    )
    OR EXISTS (
      SELECT 1 FROM projects p
      WHERE p.platform_id = platforms.id
        AND (
          p.key ILIKE '%' || $3 || '%'
          OR p.name ILIKE '%' || $3 || '%'
        )
    )
`

func ScanPlatform(row pgx.Row) (domain.Platform, error) {
	var (
		plat               domain.Platform
		startDate, dueDate *time.Time
		archivedAt         *time.Time
	)
	if err := row.Scan(
		&plat.ID,
		&plat.Key,
		&plat.Name,
		&plat.Description,
		&plat.Status,
		&plat.Visibility,
		&plat.OwnerUserID,
		&plat.LeaderUserID,
		&plat.DevelopmentUnitID,
		&startDate,
		&dueDate,
		&archivedAt,
		&plat.CreatedAt,
		&plat.UpdatedAt,
	); err != nil {
		return domain.Platform{}, err
	}
	plat.StartDate = startDate
	plat.DueDate = dueDate
	plat.ArchivedAt = archivedAt
	return plat, nil
}

func (r *PlatformRepository) ListPlatforms(ctx context.Context, opts PlatformListOptions) ([]domain.Platform, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	rowFilterCount := `
  AND ($4 = '' OR $4 = 'system_admin'
       OR owner_user_id = $5 OR leader_user_id = $5
       OR EXISTS (SELECT 1 FROM projects WHERE platform_id = platforms.id
                  AND (owner_user_id = $5 OR EXISTS (SELECT 1 FROM project_members WHERE project_id = projects.id AND user_id = $5)))
       OR (array_length($6::text[], 1) > 0 AND development_unit_id = ANY($6))
       OR (array_length($7::text[], 1) > 0 AND development_unit_id = ANY($7)))`

	rowFilterList := `
  AND ($6 = '' OR $6 = 'system_admin'
       OR owner_user_id = $7 OR leader_user_id = $7
       OR EXISTS (SELECT 1 FROM projects WHERE platform_id = platforms.id
                  AND (owner_user_id = $7 OR EXISTS (SELECT 1 FROM project_members WHERE project_id = projects.id AND user_id = $7)))
       OR (array_length($8::text[], 1) > 0 AND development_unit_id = ANY($8))
       OR (array_length($9::text[], 1) > 0 AND development_unit_id = ANY($9)))`

	var countQuery = `
SELECT COUNT(*) FROM platforms
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + platformsSearchPredicate + `)` + rowFilterCount

	var total int
	if err := r.store.Pool().QueryRow(ctx, countQuery, opts.Status, opts.IncludeArchived, opts.Query, opts.ActorRole, opts.ActorLogin, opts.OrgUnitIDs, opts.PrimaryUnitIDs).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count platforms: %w", err)
	}

	var listQuery = `
SELECT` + platformsSelectColumns + `
FROM platforms
WHERE ($1 = '' OR status = $1)
  AND ($2 OR status <> 'archived')
  AND (` + platformsSearchPredicate + `)` + rowFilterList + `
ORDER BY key ASC
LIMIT $4 OFFSET $5`

	rows, err := r.store.Pool().Query(ctx, listQuery, opts.Status, opts.IncludeArchived, opts.Query, limit, offset, opts.ActorRole, opts.ActorLogin, opts.OrgUnitIDs, opts.PrimaryUnitIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("list platforms: %w", err)
	}
	defer rows.Close()

	plats := make([]domain.Platform, 0, limit)
	for rows.Next() {
		plat, err := ScanPlatform(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan platform: %w", err)
		}
		plats = append(plats, plat)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate platforms: %w", err)
	}
	return plats, total, nil
}

func (r *PlatformRepository) GetPlatform(ctx context.Context, platformID string) (domain.Platform, error) {
	query := `SELECT` + platformsSelectColumns + ` FROM platforms WHERE id = $1::uuid`
	row := r.store.Pool().QueryRow(ctx, query, platformID)
	plat, err := ScanPlatform(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Platform{}, fmt.Errorf("get platform: %w", err)
	}
	return plat, nil
}

// GetPlatformByKey lookup by user-facing key. Useful for create-time conflict checks
// and admin tools.
func (r *PlatformRepository) GetPlatformByKey(ctx context.Context, key string) (domain.Platform, error) {
	query := `SELECT` + platformsSelectColumns + ` FROM platforms WHERE key = $1`
	row := r.store.Pool().QueryRow(ctx, query, key)
	plat, err := ScanPlatform(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Platform{}, fmt.Errorf("get platform by key: %w", err)
	}
	return plat, nil
}

// PlatformsInsertQuery is the canonical INSERT used by CreatePlatform and by
// the DREQ promote transaction (dev_requests_promote.go). Sharing the query keeps the
// archived_at consistency CHECK invariant identical across entry points.
const PlatformsInsertQuery = `
INSERT INTO platforms (key, name, description, status, visibility, owner_user_id, leader_user_id, development_unit_id, start_date, due_date, archived_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10,
        CASE WHEN $4 = 'archived' THEN NOW() ELSE NULL END)
RETURNING` + platformsSelectColumns

func (r *PlatformRepository) CreatePlatform(ctx context.Context, plat domain.Platform) (domain.Platform, error) {
	row := r.store.Pool().QueryRow(ctx, PlatformsInsertQuery,
		plat.Key, plat.Name, plat.Description, plat.Status, plat.Visibility,
		plat.OwnerUserID, plat.LeaderUserID, plat.DevelopmentUnitID, plat.StartDate, plat.DueDate,
	)
	created, err := ScanPlatform(row)
	if store.IsUniqueViolation(err) {
		return domain.Platform{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.Platform{}, store.ErrConflict
	}
	if err != nil {
		return domain.Platform{}, fmt.Errorf("create platform: %w", err)
	}
	return created, nil
}

// UpdatePlatform mutates allowed fields (name/description/owner/dates/visibility/status).
// `key` 는 immutable 이라 호출자가 별도 검증 (PATCH handler) — store 는 단순 UPDATE.
// archived consistency CHECK 위반 회피 위해 status=archived 전이 시 archived_at = NOW 자동 설정,
// 기타 status 전이 시 archived_at = NULL 로 재설정.
func (r *PlatformRepository) UpdatePlatform(ctx context.Context, plat domain.Platform) (domain.Platform, error) {
	const updateQuery = `
UPDATE platforms SET
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
RETURNING` + platformsSelectColumns

	row := r.store.Pool().QueryRow(ctx, updateQuery,
		plat.ID, plat.Name, plat.Description, plat.Status, plat.Visibility,
		plat.OwnerUserID, plat.LeaderUserID, plat.DevelopmentUnitID, plat.StartDate, plat.DueDate,
	)
	updated, err := ScanPlatform(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, store.ErrNotFound
	}
	if store.IsForeignKeyViolation(err) {
		return domain.Platform{}, store.ErrConflict
	}
	if err != nil {
		return domain.Platform{}, fmt.Errorf("update platform: %w", err)
	}
	return updated, nil
}

// ArchivePlatform is the soft-delete entry point (api §13.2 DELETE = archive).
// Sets status='archived' + archived_at=NOW. archived_reason 은 audit_logs payload 에 기록.
func (r *PlatformRepository) ArchivePlatform(ctx context.Context, platformID, archivedReason string) (domain.Platform, error) {
	const archiveQuery = `
UPDATE platforms SET
	status = 'archived',
	archived_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + platformsSelectColumns

	row := r.store.Pool().QueryRow(ctx, archiveQuery, platformID)
	archived, err := ScanPlatform(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Platform{}, fmt.Errorf("archive platform: %w", err)
	}
	_ = archivedReason // audit 기록은 handler 책임
	return archived, nil
}

// DeletePlatform — hard-delete. archived 가드는 handler 책임. cascade:
// platform_repositories ON DELETE CASCADE / projects.platform_id ON DELETE SET NULL
// (migration 000013/000014/000015). handler 가 archived 상태 검증 후에만 호출.
func (r *PlatformRepository) DeletePlatform(ctx context.Context, platformID string) error {
	const query = `DELETE FROM platforms WHERE id = $1::uuid`
	cmd, err := r.store.Pool().Exec(ctx, query, platformID)
	if err != nil {
		return fmt.Errorf("delete platform: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// CountActivePlatformRepositories — 상태 전이 가드 검증용 (planning→active 의 활성 repo ≥1).
// 직접 link (platform_repositories.sync_status='active') + 프로젝트 경유 간접 link
// (project_repositories 는 sync 상태 컬럼이 없으므로 link 존재 = 항상 active 로 간주).
// migration 000034 의 project_repositories 컬럼은 (project_id, repository_id BIGINT, role,
// linked_at) 만 존재 → repositories 테이블 JOIN 으로 repo_provider/full_name 매핑.
func (r *PlatformRepository) CountActivePlatformRepositories(ctx context.Context, platformID string) (int, error) {
	const query = `
SELECT COUNT(*) FROM (
	SELECT repo_provider, repo_full_name FROM platform_repositories
	WHERE platform_id = $1::uuid AND sync_status = 'active'
	UNION
	SELECT ip.provider_key AS repo_provider, r.full_name AS repo_full_name
	FROM project_repositories pr
	JOIN projects p              ON p.id  = pr.project_id
	JOIN repositories r          ON r.id  = pr.repository_id
	JOIN integration_providers ip ON ip.provider_id = r.provider_id
	WHERE p.platform_id = $1::uuid
) active_repos`
	var count int
	if err := r.store.Pool().QueryRow(ctx, query, platformID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active platform repositories: %w", err)
	}
	return count, nil
}

// --- Application-Repository link ---

const platformRepositoriesSelectColumns = `
	platform_id::text,
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

func ScanPlatformRepository(row pgx.Row) (domain.PlatformRepository, error) {
	var (
		link        domain.PlatformRepository
		syncErrCode string
		retryable   *bool
		syncErrAt   *time.Time
		lastSyncAt  *time.Time
	)
	if err := row.Scan(
		&link.PlatformID,
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
		return domain.PlatformRepository{}, err
	}
	link.SyncErrorCode = domain.SyncErrorCode(syncErrCode)
	link.SyncErrorRetryable = retryable
	link.SyncErrorAt = syncErrAt
	link.LastSyncAt = lastSyncAt
	return link, nil
}

func (r *PlatformRepository) ListPlatformRepositories(ctx context.Context, platformID string) ([]domain.PlatformRepository, error) {
	query := `SELECT ` + platformRepositoriesSelectColumns + `
FROM platform_repositories
WHERE platform_id = $1::uuid
UNION
SELECT
	$1::text                                    AS platform_id,
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
WHERE p.platform_id = $1::uuid
ORDER BY repo_provider ASC, repo_full_name ASC`

	rows, err := r.store.Pool().Query(ctx, query, platformID)
	if err != nil {
		return nil, fmt.Errorf("list platform repositories: %w", err)
	}
	defer rows.Close()

	links := make([]domain.PlatformRepository, 0, 4)
	for rows.Next() {
		link, err := ScanPlatformRepository(rows)
		if err != nil {
			return nil, fmt.Errorf("scan platform repository: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform repositories: %w", err)
	}
	return links, nil
}

const PlatformRepositoriesInsertQuery = `
INSERT INTO platform_repositories (
	platform_id, repo_provider, repo_full_name, external_repo_id, role, sync_status
) VALUES (
	$1::uuid, $2, $3, NULLIF($4, ''), $5, COALESCE(NULLIF($6, ''), 'requested')
)
RETURNING` + platformRepositoriesSelectColumns

func (r *PlatformRepository) CreatePlatformRepository(ctx context.Context, link domain.PlatformRepository) (domain.PlatformRepository, error) {
	syncStatus := string(link.SyncStatus)
	row := r.store.Pool().QueryRow(ctx, PlatformRepositoriesInsertQuery,
		link.PlatformID, link.RepoProvider, link.RepoFullName,
		link.ExternalRepoID, link.Role, syncStatus,
	)
	created, err := ScanPlatformRepository(row)
	if store.IsUniqueViolation(err) {
		return domain.PlatformRepository{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.PlatformRepository{}, store.ErrConflict
	}
	if err != nil {
		return domain.PlatformRepository{}, fmt.Errorf("create platform repository: %w", err)
	}
	return created, nil
}

func (r *PlatformRepository) DeletePlatformRepository(ctx context.Context, key PlatformRepositoryLinkKey) error {
	const deleteQuery = `
DELETE FROM platform_repositories
WHERE platform_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	tag, err := r.store.Pool().Exec(ctx, deleteQuery, key.PlatformID, key.RepoProvider, key.RepoFullName)
	if err != nil {
		return fmt.Errorf("delete platform repository: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *PlatformRepository) UpdatePlatformRepositorySync(ctx context.Context, key PlatformRepositoryLinkKey, status domain.PlatformRepositorySyncStatus, errorCode domain.SyncErrorCode) error {
	const updateQuery = `
UPDATE platform_repositories SET
	sync_status = $4,
	sync_error_code = NULLIF($5, ''),
	sync_error_retryable = CASE WHEN $5 = '' THEN NULL::boolean ELSE $6 END,
	sync_error_at = CASE WHEN $5 = '' THEN NULL::timestamptz ELSE NOW() END,
	last_sync_at = NOW()
WHERE platform_id = $1::uuid AND repo_provider = $2 AND repo_full_name = $3`

	retryable := domain.IsRetryableSyncError(errorCode)
	tag, err := r.store.Pool().Exec(ctx, updateQuery,
		key.PlatformID, key.RepoProvider, key.RepoFullName,
		string(status), string(errorCode), retryable,
	)
	if err != nil {
		return fmt.Errorf("update platform repository sync: %w", err)
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

func (r *PlatformRepository) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
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

func (r *PlatformRepository) UpdateSCMProvider(ctx context.Context, provider domain.SCMProvider) (domain.SCMProvider, error) {
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
