package store

import (
	"context"
	"errors"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate webhook event")

type WebhookEvent struct {
	ID             int64
	EventType      string
	DeliveryID     string
	DedupeKey      string
	RepositoryID   *int64
	RepositoryName string
	SenderLogin    string
	Payload        []byte
	Status         string
	ReceivedAt     time.Time
	ValidatedAt    *time.Time
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dbURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) SaveWebhookEvent(ctx context.Context, event WebhookEvent) (int64, error) {
	const query = `
INSERT INTO webhook_events (
	event_type,
	delivery_id,
	dedupe_key,
	repository_id,
	repository_name,
	sender_login,
	payload,
	status,
	received_at,
	validated_at
) VALUES ($1, NULLIF($2, ''), $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7::jsonb, $8, $9, $10)
ON CONFLICT (dedupe_key) DO NOTHING
RETURNING id`

	var id int64
	err := s.pool.QueryRow(
		ctx,
		query,
		event.EventType,
		event.DeliveryID,
		event.DedupeKey,
		event.RepositoryID,
		event.RepositoryName,
		event.SenderLogin,
		string(event.Payload),
		event.Status,
		event.ReceivedAt,
		event.ValidatedAt,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrDuplicateEvent
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

type ListWebhookEventsOptions struct {
	Limit  int
	Offset int
}

func (s *PostgresStore) ListWebhookEvents(ctx context.Context, opts ListWebhookEventsOptions) ([]WebhookEvent, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const query = `
SELECT
	id,
	event_type,
	COALESCE(delivery_id, ''),
	dedupe_key,
	repository_id,
	COALESCE(repository_name, ''),
	COALESCE(sender_login, ''),
	payload,
	status,
	received_at,
	validated_at
FROM webhook_events
ORDER BY received_at DESC, id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]WebhookEvent, 0, limit)
	for rows.Next() {
		var event WebhookEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.DeliveryID,
			&event.DedupeKey,
			&event.RepositoryID,
			&event.RepositoryName,
			&event.SenderLogin,
			&event.Payload,
			&event.Status,
			&event.ReceivedAt,
			&event.ValidatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStore) UpsertRepository(ctx context.Context, repository domain.Repository) error {
	const query = `
INSERT INTO repositories (
	gitea_repository_id,
	full_name,
	owner_login,
	name,
	clone_url,
	html_url,
	default_branch,
	private,
	updated_at
) VALUES (NULLIF($1, 0), $2, NULLIF($3, ''), $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, NOW())
ON CONFLICT (full_name) DO UPDATE SET
	gitea_repository_id = COALESCE(EXCLUDED.gitea_repository_id, repositories.gitea_repository_id),
	owner_login = EXCLUDED.owner_login,
	name = EXCLUDED.name,
	clone_url = EXCLUDED.clone_url,
	html_url = EXCLUDED.html_url,
	default_branch = EXCLUDED.default_branch,
	private = EXCLUDED.private,
	updated_at = NOW()`

	_, err := s.pool.Exec(
		ctx,
		query,
		repository.GiteaID,
		repository.FullName,
		repository.OwnerLogin,
		repository.Name,
		repository.CloneURL,
		repository.HTMLURL,
		repository.DefaultBranch,
		repository.Private,
	)
	return err
}

func (s *PostgresStore) UpsertUser(ctx context.Context, user domain.User) error {
	const query = `
INSERT INTO gitea_users (
	gitea_user_id,
	login,
	display_name,
	avatar_url,
	html_url,
	updated_at
) VALUES (NULLIF($1, 0), $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NOW())
ON CONFLICT (login) DO UPDATE SET
	gitea_user_id = COALESCE(EXCLUDED.gitea_user_id, gitea_users.gitea_user_id),
	display_name = EXCLUDED.display_name,
	avatar_url = EXCLUDED.avatar_url,
	html_url = EXCLUDED.html_url,
	updated_at = NOW()`

	_, err := s.pool.Exec(
		ctx,
		query,
		user.GiteaID,
		user.Login,
		user.DisplayName,
		user.AvatarURL,
		user.HTMLURL,
	)
	return err
}

func (s *PostgresStore) UpsertIssue(ctx context.Context, issue domain.Issue) error {
	repositoryID, err := s.repositoryID(ctx, issue.RepositoryGiteaID, issue.RepositoryName)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO issues (
	gitea_issue_id,
	repository_id,
	number,
	title,
	state,
	author_login,
	assignee_login,
	html_url,
	opened_at,
	closed_at,
	updated_at
) VALUES (NULLIF($1, 0), $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10, NOW())
ON CONFLICT (repository_id, number) DO UPDATE SET
	gitea_issue_id = COALESCE(EXCLUDED.gitea_issue_id, issues.gitea_issue_id),
	title = EXCLUDED.title,
	state = EXCLUDED.state,
	author_login = EXCLUDED.author_login,
	assignee_login = EXCLUDED.assignee_login,
	html_url = EXCLUDED.html_url,
	opened_at = COALESCE(EXCLUDED.opened_at, issues.opened_at),
	closed_at = EXCLUDED.closed_at,
	updated_at = NOW()`

	_, err = s.pool.Exec(
		ctx,
		query,
		issue.GiteaID,
		repositoryID,
		issue.Number,
		issue.Title,
		issue.State,
		issue.AuthorLogin,
		issue.AssigneeLogin,
		issue.HTMLURL,
		issue.OpenedAt,
		issue.ClosedAt,
	)
	return err
}

func (s *PostgresStore) UpsertPullRequest(ctx context.Context, pullRequest domain.PullRequest) error {
	repositoryID, err := s.repositoryID(ctx, pullRequest.RepositoryGiteaID, pullRequest.RepositoryName)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO pull_requests (
	gitea_pull_request_id,
	repository_id,
	number,
	title,
	state,
	author_login,
	head_branch,
	base_branch,
	head_sha,
	html_url,
	merged_at,
	closed_at,
	updated_at
) VALUES (NULLIF($1, 0), $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, $12, NOW())
ON CONFLICT (repository_id, number) DO UPDATE SET
	gitea_pull_request_id = COALESCE(EXCLUDED.gitea_pull_request_id, pull_requests.gitea_pull_request_id),
	title = EXCLUDED.title,
	state = EXCLUDED.state,
	author_login = EXCLUDED.author_login,
	head_branch = EXCLUDED.head_branch,
	base_branch = EXCLUDED.base_branch,
	head_sha = EXCLUDED.head_sha,
	html_url = EXCLUDED.html_url,
	merged_at = EXCLUDED.merged_at,
	closed_at = EXCLUDED.closed_at,
	updated_at = NOW()`

	_, err = s.pool.Exec(
		ctx,
		query,
		pullRequest.GiteaID,
		repositoryID,
		pullRequest.Number,
		pullRequest.Title,
		pullRequest.State,
		pullRequest.AuthorLogin,
		pullRequest.HeadBranch,
		pullRequest.BaseBranch,
		pullRequest.HeadSHA,
		pullRequest.HTMLURL,
		pullRequest.MergedAt,
		pullRequest.ClosedAt,
	)
	return err
}

func (s *PostgresStore) UpsertCIRun(ctx context.Context, run domain.CIRun) error {
	const query = `
INSERT INTO ci_runs (
	external_id,
	repository_id,
	repository_name,
	branch,
	commit_sha,
	status,
	conclusion,
	started_at,
	finished_at,
	duration_seconds,
	html_url,
	updated_at
) VALUES (
	$1,
	(SELECT id FROM repositories WHERE full_name = $2 LIMIT 1),
	$2,
	NULLIF($3, ''),
	NULLIF($4, ''),
	$5,
	NULLIF($6, ''),
	$7,
	$8,
	$9,
	NULLIF($10, ''),
	NOW()
)
ON CONFLICT (external_id) DO UPDATE SET
	repository_id = COALESCE(EXCLUDED.repository_id, ci_runs.repository_id),
	repository_name = EXCLUDED.repository_name,
	branch = EXCLUDED.branch,
	commit_sha = EXCLUDED.commit_sha,
	status = EXCLUDED.status,
	conclusion = EXCLUDED.conclusion,
	started_at = COALESCE(EXCLUDED.started_at, ci_runs.started_at),
	finished_at = EXCLUDED.finished_at,
	duration_seconds = EXCLUDED.duration_seconds,
	html_url = EXCLUDED.html_url,
	updated_at = NOW()`

	_, err := s.pool.Exec(
		ctx,
		query,
		run.ExternalID,
		run.RepositoryName,
		run.Branch,
		run.CommitSHA,
		run.Status,
		run.Conclusion,
		run.StartedAt,
		run.FinishedAt,
		run.DurationSeconds,
		run.HTMLURL,
	)
	return err
}

func (s *PostgresStore) MarkWebhookEventProcessed(ctx context.Context, id int64) error {
	return s.markWebhookEvent(ctx, id, "processed", "", true)
}

func (s *PostgresStore) MarkWebhookEventIgnored(ctx context.Context, id int64, reason string) error {
	return s.markWebhookEvent(ctx, id, "ignored", reason, false)
}

func (s *PostgresStore) MarkWebhookEventFailed(ctx context.Context, id int64, reason string) error {
	return s.markWebhookEvent(ctx, id, "failed", reason, false)
}

func (s *PostgresStore) markWebhookEvent(ctx context.Context, id int64, status, message string, processed bool) error {
	const query = `
UPDATE webhook_events
SET
	status = $2,
	error_message = NULLIF($3, ''),
	processed_at = CASE WHEN $4 THEN NOW() ELSE processed_at END,
	retry_count = CASE WHEN $2 = 'failed' THEN retry_count + 1 ELSE retry_count END,
	updated_at = NOW()
WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, id, status, message, processed)
	return err
}

func (s *PostgresStore) repositoryID(ctx context.Context, giteaID int64, fullName string) (int64, error) {
	const query = `
SELECT id
FROM repositories
WHERE (gitea_repository_id = NULLIF($1, 0)) OR full_name = $2
ORDER BY CASE WHEN gitea_repository_id = NULLIF($1, 0) THEN 0 ELSE 1 END
LIMIT 1`

	var id int64
	if err := s.pool.QueryRow(ctx, query, giteaID, fullName).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
