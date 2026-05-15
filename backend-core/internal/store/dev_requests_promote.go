package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// dreqMarkRegisteredUpdateQuery is the UPDATE used inside promote transactions to
// flip a dev_request row to `status='registered'` with the freshly-minted target
// (target_type, target_id). It is the transactional twin of
// MarkDevRequestRegistered (which runs on the pool, outside any tx), but reuses
// the same column list so a refactor of devRequestsSelectColumns picks both up.
const dreqMarkRegisteredUpdateQuery = `
UPDATE dev_requests SET
    status = 'registered',
    registered_target_type = $2,
    registered_target_id   = $3,
    updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + devRequestsSelectColumns

// RegisterDevRequestWithNewApplication promotes a pending/in_review dev_request
// into a freshly-created Application (optionally with one primary repository
// link) inside a single Postgres transaction. REQ-FR-DREQ-005 정합. ADR-0013 §5.
//
// Caller must have already validated dev_request status transition feasibility
// (handler-level, IsValidDevRequestTransition). FK violations on owner/leader/
// development_unit are mapped to ErrConflict so the handler can surface a
// 409 with a stable code.
func (s *PostgresStore) RegisterDevRequestWithNewApplication(
	ctx context.Context,
	drID string,
	app domain.Application,
	primaryRepo *domain.ApplicationRepository,
) (domain.DevRequest, domain.Application, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.DevRequest{}, domain.Application{}, fmt.Errorf("begin promote tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, applicationsInsertQuery,
		app.Key, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	createdApp, err := scanApplication(row)
	if isUniqueViolation(err) {
		return domain.DevRequest{}, domain.Application{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.DevRequest{}, domain.Application{}, ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, domain.Application{}, fmt.Errorf("promote: create application: %w", err)
	}

	if primaryRepo != nil {
		syncStatus := string(primaryRepo.SyncStatus)
		linkRow := tx.QueryRow(ctx, applicationRepositoriesInsertQuery,
			createdApp.ID, primaryRepo.RepoProvider, primaryRepo.RepoFullName,
			primaryRepo.ExternalRepoID, primaryRepo.Role, syncStatus,
		)
		if _, err := scanApplicationRepository(linkRow); err != nil {
			if isUniqueViolation(err) {
				return domain.DevRequest{}, domain.Application{}, ErrConflict
			}
			if isForeignKeyViolation(err) {
				return domain.DevRequest{}, domain.Application{}, ErrConflict
			}
			return domain.DevRequest{}, domain.Application{}, fmt.Errorf("promote: link primary repo: %w", err)
		}
	}

	drRow := tx.QueryRow(ctx, dreqMarkRegisteredUpdateQuery, drID, string(domain.DevRequestTargetApplication), createdApp.ID)
	updatedDR, err := scanDevRequest(drRow)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, domain.Application{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, domain.Application{}, fmt.Errorf("promote: mark dev_request registered: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DevRequest{}, domain.Application{}, fmt.Errorf("promote: commit tx: %w", err)
	}
	return updatedDR, createdApp, nil
}

// RegisterDevRequestWithNewProject promotes a pending/in_review dev_request into
// a freshly-created Project inside a single Postgres transaction. REQ-FR-DREQ-005
// 정합. ADR-0013 §5.
//
// project.RepositoryID is required (FK repositories.id). project.ApplicationID is
// optional ("" → NULL). FK or UNIQUE (repository_id, key) violations map to
// ErrConflict.
func (s *PostgresStore) RegisterDevRequestWithNewProject(
	ctx context.Context,
	drID string,
	project domain.Project,
) (domain.DevRequest, domain.Project, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("begin promote tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, projectsInsertQuery,
		project.ApplicationID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	createdProject, err := scanProject(row)
	if isUniqueViolation(err) {
		return domain.DevRequest{}, domain.Project{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.DevRequest{}, domain.Project{}, ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: create project: %w", err)
	}

	drRow := tx.QueryRow(ctx, dreqMarkRegisteredUpdateQuery, drID, string(domain.DevRequestTargetProject), createdProject.ID)
	updatedDR, err := scanDevRequest(drRow)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: mark dev_request registered: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: commit tx: %w", err)
	}
	return updatedDR, createdProject, nil
}
