package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// RepositoryPullStore 의 state 6 method (X-5 production wire follow-up).
//
// 책임: repository_pull_state (migration 000043) 의 CRUD. cold start (state row 부재)
// 시 9 method 모두 자동 upsert 로 처리. ID slot: IMPL-GITEA-PULL-STORE-01 의 state 부분.
//
// 단일 RepositoryPullStore interface 의 6 method 만 본 file. ingest 3 method 는
// repository_pull_ingest.go. ListGiteaPullTargets 는 repository_pull_targets.go.

// parseRepositoryIDBigint converts RepositoryPullStore's string form to bigint and
// returns a typed error. Used by all 6 state methods.
func parseRepositoryIDBigint(repositoryID string) (int64, error) {
	id, err := strconv.ParseInt(repositoryID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse repository id %q: %w", repositoryID, err)
	}
	return id, nil
}

// UpdatePullState upserts the per-repository pull outcome (success/error/partial).
// status in {success, error, partial}. errMsg may be empty on success. lastPullAt is
// the caller's snapshot of `now()`; on partial/error the call site may opt to NOT
// advance the cursor (per adapter's X-5 contract — see gitea_pull.go:222 + :273).
func (s *PostgresStore) UpdatePullState(
	ctx context.Context,
	repositoryID, status, errMsg string,
	lastPullAt time.Time,
) error {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return err
	}
	var lastPullAtParam any
	if !lastPullAt.IsZero() {
		lastPullAtParam = lastPullAt.UTC()
	} else {
		lastPullAtParam = nil
	}
	var errMsgParam any
	if errMsg != "" {
		errMsgParam = errMsg
	} else {
		errMsgParam = nil
	}
	const query = `
INSERT INTO repository_pull_state (repository_id, last_pull_at, last_pull_status, last_pull_error, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (repository_id) DO UPDATE SET
    last_pull_at = COALESCE(EXCLUDED.last_pull_at, repository_pull_state.last_pull_at),
    last_pull_status = EXCLUDED.last_pull_status,
    last_pull_error = EXCLUDED.last_pull_error,
    updated_at = now()`
	if _, err := s.pool.Exec(ctx, query, repoIDInt, lastPullAtParam, status, errMsgParam); err != nil {
		return fmt.Errorf("upsert repository_pull_state: %w", err)
	}
	return nil
}

// IncrementConsecutiveFailures atomically increments consecutive_failures and returns
// the new value. If the row does not exist yet (cold start), it is created with
// consecutive_failures=1 (insert default) and the trigger refreshes updated_at.
//
// Loop 의 single source-of-truth (X-5 ADR-0034 §3.3 + §3.4) — adapter 는 호출하지 않음.
func (s *PostgresStore) IncrementConsecutiveFailures(
	ctx context.Context,
	repositoryID string,
) (int, error) {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return 0, err
	}
	const query = `
INSERT INTO repository_pull_state (repository_id, consecutive_failures, updated_at)
VALUES ($1, 1, now())
ON CONFLICT (repository_id) DO UPDATE SET
    consecutive_failures = repository_pull_state.consecutive_failures + 1,
    updated_at = now()
RETURNING consecutive_failures`
	var failures int
	if err := s.pool.QueryRow(ctx, query, repoIDInt).Scan(&failures); err != nil {
		return 0, fmt.Errorf("increment consecutive_failures: %w", err)
	}
	return failures, nil
}

// ResetConsecutiveFailures resets consecutive_failures=0 and clears backoff_until.
// Called by the adapter after a fully successful pull.
func (s *PostgresStore) ResetConsecutiveFailures(
	ctx context.Context,
	repositoryID string,
) error {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return err
	}
	const query = `
UPDATE repository_pull_state
SET consecutive_failures = 0,
    backoff_until = NULL,
    updated_at = now()
WHERE repository_id = $1`
	if _, err := s.pool.Exec(ctx, query, repoIDInt); err != nil {
		return fmt.Errorf("reset consecutive_failures: %w", err)
	}
	return nil
}

// SetBackoff sets the per-repository backoff deadline. Called by the loop after a
// failed cycle (with exponential delay capped at backoffCap, default 24h).
func (s *PostgresStore) SetBackoff(
	ctx context.Context,
	repositoryID string,
	until time.Time,
) error {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return err
	}
	const query = `
INSERT INTO repository_pull_state (repository_id, backoff_until, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (repository_id) DO UPDATE SET
    backoff_until = EXCLUDED.backoff_until,
    updated_at = now()`
	var untilParam any
	if until.IsZero() {
		untilParam = nil
	} else {
		untilParam = until.UTC()
	}
	if _, err := s.pool.Exec(ctx, query, repoIDInt, untilParam); err != nil {
		return fmt.Errorf("set backoff_until: %w", err)
	}
	return nil
}

// BackoffUntil returns the current backoff deadline. Zero time + nil error if the
// row does not exist or the backoff_until is NULL.
func (s *PostgresStore) BackoffUntil(
	ctx context.Context,
	repositoryID string,
) (time.Time, error) {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return time.Time{}, err
	}
	var until *time.Time
	const query = `SELECT backoff_until FROM repository_pull_state WHERE repository_id = $1`
	if err := s.pool.QueryRow(ctx, query, repoIDInt).Scan(&until); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("fetch backoff_until: %w", err)
	}
	if until == nil {
		return time.Time{}, nil
	}
	return *until, nil
}

// LastPullAt returns the most recent successful pull timestamp. Zero time if the
// row does not exist or last_pull_at is NULL.
func (s *PostgresStore) LastPullAt(
	ctx context.Context,
	repositoryID string,
) (time.Time, error) {
	repoIDInt, err := parseRepositoryIDBigint(repositoryID)
	if err != nil {
		return time.Time{}, err
	}
	var at *time.Time
	const query = `SELECT last_pull_at FROM repository_pull_state WHERE repository_id = $1`
	if err := s.pool.QueryRow(ctx, query, repoIDInt).Scan(&at); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("fetch last_pull_at: %w", err)
	}
	if at == nil {
		return time.Time{}, nil
	}
	return *at, nil
}
