package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// Integration store (API-58, sprint claude/work_260514-c).
// project_integrations 테이블은 migration 000016. scope 컬럼으로 application/project
// polymorphism.

// IntegrationListOptions parameterizes ListIntegrations.
type IntegrationListOptions struct {
	Scope           domain.IntegrationScope // 빈 문자열이면 모든 scope
	ApplicationID   string                  // scope=application 일 때
	ProjectID       string                  // scope=project 일 때
	IntegrationType domain.IntegrationType  // 빈 문자열이면 모든 type
	Limit           int
	Offset          int
}

const integrationsSelectColumns = `
	id::text,
	scope,
	COALESCE(project_id::text, ''),
	COALESCE(application_id::text, ''),
	integration_type,
	external_key,
	url,
	policy,
	created_at,
	updated_at`

func scanIntegration(row pgx.Row) (domain.ProjectIntegration, error) {
	var i domain.ProjectIntegration
	if err := row.Scan(
		&i.ID, &i.Scope, &i.ProjectID, &i.ApplicationID,
		&i.IntegrationType, &i.ExternalKey, &i.URL, &i.Policy,
		&i.CreatedAt, &i.UpdatedAt,
	); err != nil {
		return domain.ProjectIntegration{}, err
	}
	return i, nil
}

func (s *PostgresStore) ListIntegrations(ctx context.Context, opts IntegrationListOptions) ([]domain.ProjectIntegration, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	const countQuery = `
SELECT COUNT(*) FROM project_integrations
WHERE ($1 = '' OR scope = $1)
  AND ($2 = '' OR application_id = NULLIF($2, '')::uuid)
  AND ($3 = '' OR project_id = NULLIF($3, '')::uuid)
  AND ($4 = '' OR integration_type = $4)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery,
		string(opts.Scope), opts.ApplicationID, opts.ProjectID, string(opts.IntegrationType)).
		Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count integrations: %w", err)
	}
	query := `
SELECT` + integrationsSelectColumns + `
FROM project_integrations
WHERE ($3 = '' OR scope = $3)
  AND ($4 = '' OR application_id = NULLIF($4, '')::uuid)
  AND ($5 = '' OR project_id = NULLIF($5, '')::uuid)
  AND ($6 = '' OR integration_type = $6)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset,
		string(opts.Scope), opts.ApplicationID, opts.ProjectID, string(opts.IntegrationType))
	if err != nil {
		return nil, 0, fmt.Errorf("list integrations: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ProjectIntegration, 0, limit)
	for rows.Next() {
		i, err := scanIntegration(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan integration: %w", err)
		}
		out = append(out, i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate integrations: %w", err)
	}
	return out, total, nil
}

func (s *PostgresStore) GetIntegration(ctx context.Context, id string) (domain.ProjectIntegration, error) {
	query := `SELECT` + integrationsSelectColumns + ` FROM project_integrations WHERE id = $1::uuid`
	row := s.pool.QueryRow(ctx, query, id)
	integration, err := scanIntegration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProjectIntegration{}, ErrNotFound
	}
	if err != nil {
		return domain.ProjectIntegration{}, fmt.Errorf("get integration: %w", err)
	}
	return integration, nil
}

func (s *PostgresStore) CreateIntegration(ctx context.Context, integration domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	const insertQuery = `
INSERT INTO project_integrations (
	scope, project_id, application_id, integration_type, external_key, url, policy
) VALUES (
	$1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, $4, $5, $6, $7
)
RETURNING` + integrationsSelectColumns
	row := s.pool.QueryRow(ctx, insertQuery,
		string(integration.Scope), integration.ProjectID, integration.ApplicationID,
		string(integration.IntegrationType), integration.ExternalKey, integration.URL,
		string(integration.Policy),
	)
	created, err := scanIntegration(row)
	if isUniqueViolation(err) {
		return domain.ProjectIntegration{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.ProjectIntegration{}, ErrConflict
	}
	if err != nil {
		return domain.ProjectIntegration{}, fmt.Errorf("create integration: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) UpdateIntegration(ctx context.Context, integration domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	const updateQuery = `
UPDATE project_integrations SET
	external_key = $2,
	url = $3,
	policy = $4,
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + integrationsSelectColumns
	row := s.pool.QueryRow(ctx, updateQuery,
		integration.ID, integration.ExternalKey, integration.URL, string(integration.Policy),
	)
	updated, err := scanIntegration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProjectIntegration{}, ErrNotFound
	}
	// PR #107 codex review P2 — UpdateIntegration 의 external_key 변경이 partial
	// UNIQUE 인덱스 (application_id/project_id + integration_type + external_key) 를
	// 위반할 수 있으므로 ErrConflict 매핑. createIntegration 의 대칭.
	if isUniqueViolation(err) {
		return domain.ProjectIntegration{}, ErrConflict
	}
	if err != nil {
		return domain.ProjectIntegration{}, fmt.Errorf("update integration: %w", err)
	}
	return updated, nil
}

func (s *PostgresStore) DeleteIntegration(ctx context.Context, id string) error {
	const deleteQuery = `DELETE FROM project_integrations WHERE id = $1::uuid`
	tag, err := s.pool.Exec(ctx, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("delete integration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
