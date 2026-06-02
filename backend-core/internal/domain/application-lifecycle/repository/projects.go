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

// ProjectListOptions parameterizes ListProjects.
type ProjectListOptions = store.ProjectListOptions

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

func ScanProject(row pgx.Row) (domain.Project, error) {
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

func (r *ApplicationRepository) ListProjects(ctx context.Context, opts ProjectListOptions) ([]domain.Project, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 5000 {
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
  AND ($4 OR status <> 'archived')
  AND ($5 = false OR application_id IS NULL)`

	var total int
	if err := r.store.Pool().QueryRow(ctx, countQuery, opts.RepositoryID, nullableUUIDArg(opts.ApplicationID), opts.Status, opts.IncludeArchived, opts.StandaloneOnly).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	query := `
SELECT` + projectsSelectColumns + `
FROM projects
WHERE ($3::bigint = 0 OR repository_id = $3)
  AND ($4::uuid IS NULL OR application_id = $4::uuid)
  AND ($5 = '' OR status = $5)
  AND ($6 OR status <> 'archived')
  AND ($7 = false OR application_id IS NULL)
ORDER BY key ASC
LIMIT $1 OFFSET $2`

	rows, err := r.store.Pool().Query(ctx, query, limit, offset, opts.RepositoryID, nullableUUIDArg(opts.ApplicationID), opts.Status, opts.IncludeArchived, opts.StandaloneOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0, limit)
	for rows.Next() {
		p, err := ScanProject(rows)
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

func (r *ApplicationRepository) GetProject(ctx context.Context, projectID string) (domain.Project, error) {
	query := `SELECT` + projectsSelectColumns + ` FROM projects WHERE id = $1::uuid`
	row := r.store.Pool().QueryRow(ctx, query, projectID)
	p, err := ScanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project: %w", err)
	}

	// Fetch project members
	const membersQuery = `
SELECT project_id::text, user_id, project_role, joined_at
FROM project_members
WHERE project_id = $1::uuid
ORDER BY joined_at ASC`
	rows, err := r.store.Pool().Query(ctx, membersQuery, projectID)
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project members: %w", err)
	}
	defer rows.Close()

	var members []domain.ProjectMember
	for rows.Next() {
		var m domain.ProjectMember
		var roleStr string
		if err := rows.Scan(&m.ProjectID, &m.UserID, &roleStr, &m.JoinedAt); err != nil {
			return domain.Project{}, fmt.Errorf("scan project member: %w", err)
		}
		m.ProjectRole = domain.ProjectMemberRole(roleStr)
		members = append(members, m)
	}
	p.ProjectMembers = members

	return p, nil
}

// ProjectsInsertQuery is the canonical INSERT used by CreateProject and by the DREQ
// promote transaction (dev_requests_promote.go). Sharing it keeps the (repository_id,
// key) UNIQUE constraint and NULLIF semantics identical across entry points.
const ProjectsInsertQuery = `
INSERT INTO projects (
	application_id, repository_id, key, name, description, status, visibility,
	owner_user_id, start_date, due_date, archived_at
) VALUES (
	NULLIF($1, '')::uuid, NULLIF($2::bigint, 0), $3, $4, NULLIF($5, ''), $6, $7,
	NULLIF($8, ''), $9, $10,
	CASE WHEN $6 = 'archived' THEN NOW() ELSE NULL END
)
RETURNING` + projectsSelectColumns

func (r *ApplicationRepository) CreateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.Project{}, fmt.Errorf("begin create project tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, ProjectsInsertQuery,
		project.ApplicationID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	created, err := ScanProject(row)
	if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) {
		return domain.Project{}, store.ErrConflict
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project (tx): %w", err)
	}

	// 1. Primary repository link insert
	if created.RepositoryID > 0 {
		const insertLink = `
INSERT INTO project_repositories (project_id, repository_id, role)
VALUES ($1::uuid, $2, 'primary')
ON CONFLICT (project_id, repository_id) DO NOTHING`
		if _, err := tx.Exec(ctx, insertLink, created.ID, created.RepositoryID); err != nil {
			if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
				return domain.Project{}, store.ErrConflict
			}
			return domain.Project{}, fmt.Errorf("create project primary link (tx): %w", err)
		}
	}

	// 2. Insert project members
	for _, member := range project.ProjectMembers {
		if strings.TrimSpace(member.UserID) == "" {
			continue
		}
		const insertMember = `
INSERT INTO project_members (project_id, user_id, project_role, joined_at)
VALUES ($1::uuid, $2, $3, COALESCE(NULLIF($4, '0001-01-01T00:00:00Z'::timestamptz), NOW()))
ON CONFLICT (project_id, user_id) DO UPDATE SET
	project_role = EXCLUDED.project_role`
		var joinedAt interface{} = member.JoinedAt
		if member.JoinedAt.IsZero() {
			joinedAt = nil
		}
		if _, err := tx.Exec(ctx, insertMember, created.ID, member.UserID, string(member.ProjectRole), joinedAt); err != nil {
			if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
				return domain.Project{}, store.ErrConflict
			}
			return domain.Project{}, fmt.Errorf("create project member (tx): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Project{}, fmt.Errorf("commit create project tx: %w", err)
	}

	return r.GetProject(ctx, created.ID)
}

// RepositoryCreatePayload — project 생성 시 동반 생성할 repository 입력. project tx
// 안에서 함께 생성하기 위해 store 로 전달.
type RepositoryCreatePayload = store.RepositoryCreatePayload

// CreateRepositoryDraft 는 draft repository 를 생성한다. providerID 는 연동 대상
// integration_providers FK. system 이 생성하므로 source='system'.
func (r *ApplicationRepository) CreateRepositoryDraft(ctx context.Context, key, slug, providerID string) (domain.Repository, error) {
	fullName := strings.TrimSpace(slug)
	name := strings.TrimSpace(key)
	if fullName == "" || name == "" {
		return domain.Repository{}, store.ErrConflict
	}

	const query = `
INSERT INTO repositories (
	full_name, name, owner_login, clone_url, html_url, default_branch, private,
	repository_status, source, provider_id, publish_requested_at, published_at, updated_at
) VALUES (
	$1, $2, NULLIF(split_part($1, '/', 1), ''), '', '', 'main', false,
	'draft', 'system', NULLIF($3, '')::uuid, NULL, NULL, NOW()
)
RETURNING
	id,
	COALESCE(gitea_repository_id, 0),
	full_name,
	COALESCE(owner_login, ''),
	name,
	COALESCE(clone_url, ''),
	COALESCE(html_url, ''),
	COALESCE(default_branch, ''),
	private,
	COALESCE(repository_status, 'draft'),
	COALESCE(provider_id::text, ''),
	publish_requested_at,
	published_at,
	updated_at`

	var repo domain.Repository
	if err := r.store.Pool().QueryRow(ctx, query, fullName, name, strings.TrimSpace(providerID)).Scan(
		&repo.ID,
		&repo.GiteaID,
		&repo.FullName,
		&repo.OwnerLogin,
		&repo.Name,
		&repo.CloneURL,
		&repo.HTMLURL,
		&repo.DefaultBranch,
		&repo.Private,
		&repo.Status,
		&repo.ProviderID,
		&repo.PublishRequestedAt,
		&repo.PublishedAt,
		&repo.UpdatedAt,
	); err != nil {
		if store.IsUniqueViolation(err) {
			return domain.Repository{}, store.ErrConflict
		}
		return domain.Repository{}, fmt.Errorf("create repository draft: %w", err)
	}
	return repo, nil
}

func (r *ApplicationRepository) MarkRepositoryDraftPublishRequested(ctx context.Context, repositoryID int64) (domain.Repository, error) {
	const query = `
UPDATE repositories
SET
	publish_requested_at = NOW(),
	updated_at = NOW()
WHERE id = $1
  AND repository_status = 'draft'
RETURNING
	id,
	COALESCE(gitea_repository_id, 0),
	full_name,
	COALESCE(owner_login, ''),
	name,
	COALESCE(clone_url, ''),
	COALESCE(html_url, ''),
	COALESCE(default_branch, ''),
	private,
	COALESCE(repository_status, 'draft'),
	COALESCE(provider_id::text, ''),
	publish_requested_at,
	published_at,
	updated_at`
	var repo domain.Repository
	if err := r.store.Pool().QueryRow(ctx, query, repositoryID).Scan(
		&repo.ID,
		&repo.GiteaID,
		&repo.FullName,
		&repo.OwnerLogin,
		&repo.Name,
		&repo.CloneURL,
		&repo.HTMLURL,
		&repo.DefaultBranch,
		&repo.Private,
		&repo.Status,
		&repo.ProviderID,
		&repo.PublishRequestedAt,
		&repo.PublishedAt,
		&repo.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Repository{}, store.ErrNotFound
		}
		return domain.Repository{}, fmt.Errorf("mark repository publish requested: %w", err)
	}
	return repo, nil
}

// GetRepositoryByID — detail page 용. ListRepositories 와 동일하게 linked count
// subquery 포함.
func (r *ApplicationRepository) GetRepositoryByID(ctx context.Context, repositoryID int64) (domain.Repository, error) {
	const query = `
SELECT
	r.id,
	COALESCE(r.gitea_repository_id, 0),
	r.full_name,
	COALESCE(r.owner_login, ''),
	r.name,
	COALESCE(r.clone_url, ''),
	COALESCE(r.html_url, ''),
	COALESCE(r.default_branch, ''),
	r.private,
	COALESCE(r.repository_status, 'active'),
	publish_requested_at,
	published_at,
	r.updated_at,
	COALESCE(r.source, 'scm'),
	COALESCE(r.provider_id::text, ''),
	COALESCE(p.provider_key, ''),
	COALESCE(r.description, ''),
	COALESCE((SELECT COUNT(*)
	          FROM application_repositories ar
	          WHERE ar.repo_full_name = r.full_name), 0)::int AS linked_applications_count,
	COALESCE((SELECT COUNT(*)
	          FROM project_repositories pr
	          WHERE pr.repository_id = r.id), 0)::int AS linked_projects_count
FROM repositories r
LEFT JOIN integration_providers p ON p.provider_id = r.provider_id
WHERE r.id = $1`

	var repo domain.Repository
	if err := r.store.Pool().QueryRow(ctx, query, repositoryID).Scan(
		&repo.ID,
		&repo.GiteaID,
		&repo.FullName,
		&repo.OwnerLogin,
		&repo.Name,
		&repo.CloneURL,
		&repo.HTMLURL,
		&repo.DefaultBranch,
		&repo.Private,
		&repo.Status,
		&repo.PublishRequestedAt,
		&repo.PublishedAt,
		&repo.UpdatedAt,
		&repo.Source,
		&repo.ProviderID,
		&repo.ProviderKey,
		&repo.Description,
		&repo.LinkedApplicationsCount,
		&repo.LinkedProjectsCount,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Repository{}, store.ErrNotFound
		}
		return domain.Repository{}, fmt.Errorf("get repository by id: %w", err)
	}
	return repo, nil
}

// CreateProjectWithRepositoryPayload creates the project — optionally creating and
// linking a companion repository — in ONE transaction.
func (r *ApplicationRepository) CreateProjectWithRepositoryPayload(ctx context.Context, project domain.Project, repositoryIDs []int64, repoPayload *RepositoryCreatePayload) (domain.Project, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.Project{}, fmt.Errorf("begin create project tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if repoPayload != nil {
		repoID, repoErr := createRepositoryTx(ctx, tx, repoPayload.Key, repoPayload.Slug, repoPayload.SCMProvider)
		if store.IsUniqueViolation(repoErr) || store.IsForeignKeyViolation(repoErr) || store.IsCheckViolation(repoErr, "") {
			return domain.Project{}, store.ErrConflict
		}
		if repoErr != nil {
			return domain.Project{}, fmt.Errorf("create companion repository (tx): %w", repoErr)
		}
		if project.RepositoryID == 0 {
			project.RepositoryID = repoID
		}
		repositoryIDs = append(repositoryIDs, repoID)
	}

	row := tx.QueryRow(ctx, ProjectsInsertQuery,
		project.ApplicationID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	created, err := ScanProject(row)
	if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) {
		return domain.Project{}, store.ErrConflict
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
			if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
				return domain.Project{}, store.ErrConflict
			}
			return domain.Project{}, fmt.Errorf("create project repository link (tx): %w", err)
		}
	}

	// Insert project members
	for _, member := range project.ProjectMembers {
		if strings.TrimSpace(member.UserID) == "" {
			continue
		}
		const insertMember = `
INSERT INTO project_members (project_id, user_id, project_role, joined_at)
VALUES ($1::uuid, $2, $3, COALESCE(NULLIF($4, '0001-01-01T00:00:00Z'::timestamptz), NOW()))
ON CONFLICT (project_id, user_id) DO UPDATE SET
	project_role = EXCLUDED.project_role`
		var joinedAt interface{} = member.JoinedAt
		if member.JoinedAt.IsZero() {
			joinedAt = nil
		}
		if _, err := tx.Exec(ctx, insertMember, created.ID, member.UserID, string(member.ProjectRole), joinedAt); err != nil {
			if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
				return domain.Project{}, store.ErrConflict
			}
			return domain.Project{}, fmt.Errorf("create project member (tx): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Project{}, fmt.Errorf("commit create project tx: %w", err)
	}
	return created, nil
}

// createRepositoryTx upserts a companion repository within the given transaction.
func createRepositoryTx(ctx context.Context, tx pgx.Tx, key, slug, scmProvider string) (int64, error) {
	fullName := strings.TrimSpace(slug)
	name := strings.TrimSpace(key)
	if name == "" {
		name = fullName
	}
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

func (r *ApplicationRepository) UpdateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.Project{}, fmt.Errorf("begin update project tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const updateQuery = `
UPDATE projects SET
	application_id = NULLIF($9, '')::uuid,
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

	row := tx.QueryRow(ctx, updateQuery,
		project.ID, project.Name, project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate, project.ApplicationID,
	)
	updated, err := ScanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project (tx): %w", err)
	}

	// Sync project members
	// 1. Delete existing members
	const deleteMembers = `DELETE FROM project_members WHERE project_id = $1::uuid`
	if _, err := tx.Exec(ctx, deleteMembers, project.ID); err != nil {
		return domain.Project{}, fmt.Errorf("delete project members (tx): %w", err)
	}

	// 2. Insert new/updated members
	for _, member := range project.ProjectMembers {
		if strings.TrimSpace(member.UserID) == "" {
			continue
		}
		const insertMember = `
INSERT INTO project_members (project_id, user_id, project_role, joined_at)
VALUES ($1::uuid, $2, $3, COALESCE(NULLIF($4, '0001-01-01T00:00:00Z'::timestamptz), NOW()))`
		var joinedAt interface{} = member.JoinedAt
		if member.JoinedAt.IsZero() {
			joinedAt = nil
		}
		if _, err := tx.Exec(ctx, insertMember, project.ID, member.UserID, string(member.ProjectRole), joinedAt); err != nil {
			if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
				return domain.Project{}, store.ErrConflict
			}
			return domain.Project{}, fmt.Errorf("update project member (tx): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Project{}, fmt.Errorf("commit update project tx: %w", err)
	}

	// Return updated project with freshly synced members loaded
	return r.GetProject(ctx, updated.ID)
}

func (r *ApplicationRepository) ArchiveProject(ctx context.Context, projectID, archivedReason string) (domain.Project, error) {
	const archiveQuery = `
UPDATE projects SET
	status = 'archived',
	archived_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + projectsSelectColumns

	row := r.store.Pool().QueryRow(ctx, archiveQuery, projectID)
	archived, err := ScanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("archive project: %w", err)
	}
	_ = archivedReason
	return archived, nil
}

func (r *ApplicationRepository) ListProjectRepositories(ctx context.Context, projectID string) ([]domain.ProjectRepository, error) {
	const query = `
SELECT project_id::text, repository_id, role, linked_at
FROM project_repositories
WHERE project_id = $1::uuid
ORDER BY repository_id ASC`

	rows, err := r.store.Pool().Query(ctx, query, projectID)
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

func (r *ApplicationRepository) CreateProjectRepository(ctx context.Context, link domain.ProjectRepository) (domain.ProjectRepository, error) {
	const query = `
INSERT INTO project_repositories (project_id, repository_id, role)
VALUES ($1::uuid, $2, $3)
RETURNING project_id::text, repository_id, role, linked_at`

	row := r.store.Pool().QueryRow(ctx, query, link.ProjectID, link.RepositoryID, link.Role)
	var created domain.ProjectRepository
	if err := row.Scan(&created.ProjectID, &created.RepositoryID, &created.Role, &created.LinkedAt); err != nil {
		if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
			return domain.ProjectRepository{}, store.ErrConflict
		}
		return domain.ProjectRepository{}, fmt.Errorf("create project repository: %w", err)
	}
	return created, nil
}

func (r *ApplicationRepository) DeleteProjectRepository(ctx context.Context, projectID string, repositoryID int64) error {
	const query = `DELETE FROM project_repositories WHERE project_id = $1::uuid AND repository_id = $2`
	cmd, err := r.store.Pool().Exec(ctx, query, projectID, repositoryID)
	if err != nil {
		return fmt.Errorf("delete project repository: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *ApplicationRepository) DeleteProject(ctx context.Context, projectID string) error {
	const query = `DELETE FROM projects WHERE id = $1::uuid`
	cmd, err := r.store.Pool().Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}
