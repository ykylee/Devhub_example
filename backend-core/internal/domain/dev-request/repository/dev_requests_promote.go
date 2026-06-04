package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	appRepo "github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	"github.com/devhub/backend-core/internal/store"
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
    rejected_reason        = NULL,
    updated_at = NOW()
WHERE id = $1::uuid AND status IN ('pending', 'in_review')
RETURNING` + devRequestsSelectColumns

// RegisterDevRequestWithNewPlatform promotes a pending/in_review dev_request
// into a freshly-created Application (optionally with one primary repository
// link) inside a single Postgres transaction.
func (r *DevRequestRepository) RegisterDevRequestWithNewPlatform(
	ctx context.Context,
	drID string,
	app domain.Platform,
	primaryRepo *domain.PlatformRepository,
) (domain.DevRequest, domain.Platform, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.DevRequest{}, domain.Platform{}, fmt.Errorf("begin promote tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, appRepo.PlatformsInsertQuery,
		app.Key, app.Name, app.Description, app.Status, app.Visibility,
		app.OwnerUserID, app.LeaderUserID, app.DevelopmentUnitID, app.StartDate, app.DueDate,
	)
	createdApp, err := appRepo.ScanPlatform(row)
	if store.IsUniqueViolation(err) {
		return domain.DevRequest{}, domain.Platform{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.DevRequest{}, domain.Platform{}, store.ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, domain.Platform{}, fmt.Errorf("promote: create application: %w", err)
	}

	if primaryRepo != nil {
		syncStatus := string(primaryRepo.SyncStatus)
		linkRow := tx.QueryRow(ctx, appRepo.PlatformRepositoriesInsertQuery,
			createdApp.ID, primaryRepo.RepoProvider, primaryRepo.RepoFullName,
			primaryRepo.ExternalRepoID, primaryRepo.Role, syncStatus,
		)
		if _, err := appRepo.ScanPlatformRepository(linkRow); err != nil {
			if store.IsUniqueViolation(err) {
				return domain.DevRequest{}, domain.Platform{}, store.ErrConflict
			}
			if store.IsForeignKeyViolation(err) {
				return domain.DevRequest{}, domain.Platform{}, store.ErrConflict
			}
			if store.IsCheckViolation(err, "") {
				return domain.DevRequest{}, domain.Platform{}, store.ErrConflict
			}
			return domain.DevRequest{}, domain.Platform{}, fmt.Errorf("promote: link primary repo: %w", err)
		}
	}

	drRow := tx.QueryRow(ctx, dreqMarkRegisteredUpdateQuery, drID, string(domain.DevRequestTargetPlatform), createdApp.ID)
	updatedDR, err := scanDevRequest(drRow)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, domain.Platform{}, store.ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, domain.Platform{}, fmt.Errorf("promote: mark dev_request registered: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DevRequest{}, domain.Platform{}, fmt.Errorf("promote: commit tx: %w", err)
	}
	return updatedDR, createdApp, nil
}

// RegisterDevRequestWithNewProject promotes a pending/in_review dev_request into
// a freshly-created Project inside a single Postgres transaction.
func (r *DevRequestRepository) RegisterDevRequestWithNewProject(
	ctx context.Context,
	drID string,
	project domain.Project,
) (domain.DevRequest, domain.Project, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("begin promote tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, appRepo.ProjectsInsertQuery,
		project.PlatformID, project.RepositoryID, project.Key, project.Name,
		project.Description, project.Status, project.Visibility,
		project.OwnerUserID, project.StartDate, project.DueDate,
	)
	createdProject, err := appRepo.ScanProject(row)
	if store.IsUniqueViolation(err) {
		return domain.DevRequest{}, domain.Project{}, store.ErrConflict
	}
	if store.IsForeignKeyViolation(err) {
		return domain.DevRequest{}, domain.Project{}, store.ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: create project: %w", err)
	}

	drRow := tx.QueryRow(ctx, dreqMarkRegisteredUpdateQuery, drID, string(domain.DevRequestTargetProject), createdProject.ID)
	updatedDR, err := scanDevRequest(drRow)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, domain.Project{}, store.ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: mark dev_request registered: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DevRequest{}, domain.Project{}, fmt.Errorf("promote: commit tx: %w", err)
	}
	return updatedDR, createdProject, nil
}
