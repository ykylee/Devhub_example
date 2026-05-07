package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate webhook event")
var ErrNotFound = errors.New("not found")

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

func (s *PostgresStore) UpsertRisk(ctx context.Context, risk domain.Risk) error {
	suggestedActions, err := json.Marshal(risk.SuggestedActions)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO risks (
	risk_key,
	title,
	reason,
	impact,
	status,
	owner_login,
	source_type,
	source_id,
	suggested_actions,
	detected_at,
	mitigated_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, NULLIF($8, ''), $9::jsonb, $10, $11, NOW())
ON CONFLICT (risk_key) DO UPDATE SET
	title = EXCLUDED.title,
	reason = EXCLUDED.reason,
	impact = EXCLUDED.impact,
	status = CASE
		WHEN risks.status = 'mitigated' THEN risks.status
		ELSE EXCLUDED.status
	END,
	owner_login = EXCLUDED.owner_login,
	source_type = EXCLUDED.source_type,
	source_id = EXCLUDED.source_id,
	suggested_actions = EXCLUDED.suggested_actions,
	detected_at = LEAST(risks.detected_at, EXCLUDED.detected_at),
	mitigated_at = CASE
		WHEN risks.status = 'mitigated' THEN risks.mitigated_at
		ELSE EXCLUDED.mitigated_at
	END,
	updated_at = NOW()`

	detectedAt := risk.DetectedAt
	if detectedAt.IsZero() {
		detectedAt = time.Now().UTC()
	}
	_, err = s.pool.Exec(
		ctx,
		query,
		risk.RiskKey,
		risk.Title,
		risk.Reason,
		risk.Impact,
		risk.Status,
		risk.OwnerLogin,
		risk.SourceType,
		risk.SourceID,
		string(suggestedActions),
		detectedAt,
		risk.MitigatedAt,
	)
	return err
}

func (s *PostgresStore) CreateRiskMitigationCommand(ctx context.Context, req domain.RiskMitigationCommandRequest) (domain.Command, domain.AuditLog, bool, error) {
	if req.IdempotencyKey != "" {
		command, auditLog, found, err := s.commandByIdempotencyKey(ctx, req.IdempotencyKey, "risk_mitigation")
		if err != nil {
			return domain.Command{}, domain.AuditLog{}, false, err
		}
		if found {
			return command, auditLog, true, nil
		}
	}

	if err := s.ensureRiskExists(ctx, req.RiskID); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	commandID, err := randomPrefixedID("cmd")
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	auditID, err := randomPrefixedID("audit")
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	requestPayload := req.RequestPayload
	if requestPayload == nil {
		requestPayload = map[string]any{}
	}
	commandPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	auditPayload, err := json.Marshal(map[string]any{
		"action_type": req.ActionType,
		"dry_run":     req.DryRun,
		"reason":      req.Reason,
	})
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	defer tx.Rollback(ctx)

	const commandQuery = `
INSERT INTO commands (
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	idempotency_key,
	request_payload,
	result_payload,
	updated_at
) VALUES ($1, 'risk_mitigation', 'risk', $2, $3, 'pending', $4, $5, $6, $7, NULLIF($8, ''), $9::jsonb, '{}'::jsonb, NOW())
RETURNING
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at`

	var command domain.Command
	var scannedCommandPayload []byte
	var scannedResultPayload []byte
	err = tx.QueryRow(
		ctx,
		commandQuery,
		commandID,
		req.RiskID,
		req.ActionType,
		req.ActorLogin,
		req.Reason,
		req.DryRun,
		req.RequiresApproval,
		req.IdempotencyKey,
		string(commandPayload),
	).Scan(
		&command.ID,
		&command.CommandID,
		&command.CommandType,
		&command.TargetType,
		&command.TargetID,
		&command.ActionType,
		&command.Status,
		&command.ActorLogin,
		&command.Reason,
		&command.DryRun,
		&command.RequiresApproval,
		&command.IdempotencyKey,
		&scannedCommandPayload,
		&scannedResultPayload,
		&command.CreatedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		if req.IdempotencyKey != "" {
			existingCommand, existingAuditLog, found, findErr := s.commandByIdempotencyKey(ctx, req.IdempotencyKey, "risk_mitigation")
			if findErr != nil {
				return domain.Command{}, domain.AuditLog{}, false, findErr
			}
			if found {
				return existingCommand, existingAuditLog, true, nil
			}
		}
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedCommandPayload, &command.RequestPayload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedResultPayload, &command.ResultPayload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	const auditQuery = `
INSERT INTO audit_logs (
	audit_id,
	actor_login,
	action,
	target_type,
	target_id,
	command_id,
	payload
) VALUES ($1, $2, 'risk_mitigation.requested', 'risk', $3, $4, $5::jsonb)
RETURNING
	id,
	audit_id,
	actor_login,
	action,
	target_type,
	target_id,
	COALESCE(command_id, ''),
	payload,
	created_at`

	var auditLog domain.AuditLog
	var scannedAuditPayload []byte
	if err := tx.QueryRow(
		ctx,
		auditQuery,
		auditID,
		req.ActorLogin,
		req.RiskID,
		command.CommandID,
		string(auditPayload),
	).Scan(
		&auditLog.ID,
		&auditLog.AuditID,
		&auditLog.ActorLogin,
		&auditLog.Action,
		&auditLog.TargetType,
		&auditLog.TargetID,
		&auditLog.CommandID,
		&scannedAuditPayload,
		&auditLog.CreatedAt,
	); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedAuditPayload, &auditLog.Payload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	return command, auditLog, false, nil
}

func (s *PostgresStore) CreateServiceActionCommand(ctx context.Context, req domain.ServiceActionCommandRequest) (domain.Command, domain.AuditLog, bool, error) {
	if req.IdempotencyKey != "" {
		command, auditLog, found, err := s.commandByIdempotencyKey(ctx, req.IdempotencyKey, "service_action")
		if err != nil {
			return domain.Command{}, domain.AuditLog{}, false, err
		}
		if found {
			return command, auditLog, true, nil
		}
	}

	commandID, err := randomPrefixedID("cmd")
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	auditID, err := randomPrefixedID("audit")
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	requestPayload := req.RequestPayload
	if requestPayload == nil {
		requestPayload = map[string]any{}
	}
	commandPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	auditPayload, err := json.Marshal(map[string]any{
		"service_id":  req.ServiceID,
		"action_type": req.ActionType,
		"dry_run":     req.DryRun,
		"force":       req.Force,
		"reason":      req.Reason,
	})
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	defer tx.Rollback(ctx)

	const commandQuery = `
INSERT INTO commands (
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	idempotency_key,
	request_payload,
	result_payload,
	updated_at
) VALUES ($1, 'service_action', 'service', $2, $3, 'pending', $4, $5, $6, $7, NULLIF($8, ''), $9::jsonb, '{}'::jsonb, NOW())
RETURNING
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at`

	var command domain.Command
	var scannedCommandPayload []byte
	var scannedResultPayload []byte
	err = tx.QueryRow(
		ctx,
		commandQuery,
		commandID,
		req.ServiceID,
		req.ActionType,
		req.ActorLogin,
		req.Reason,
		req.DryRun,
		req.RequiresApproval,
		req.IdempotencyKey,
		string(commandPayload),
	).Scan(
		&command.ID,
		&command.CommandID,
		&command.CommandType,
		&command.TargetType,
		&command.TargetID,
		&command.ActionType,
		&command.Status,
		&command.ActorLogin,
		&command.Reason,
		&command.DryRun,
		&command.RequiresApproval,
		&command.IdempotencyKey,
		&scannedCommandPayload,
		&scannedResultPayload,
		&command.CreatedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		if req.IdempotencyKey != "" {
			existingCommand, existingAuditLog, found, findErr := s.commandByIdempotencyKey(ctx, req.IdempotencyKey, "service_action")
			if findErr != nil {
				return domain.Command{}, domain.AuditLog{}, false, findErr
			}
			if found {
				return existingCommand, existingAuditLog, true, nil
			}
		}
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedCommandPayload, &command.RequestPayload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedResultPayload, &command.ResultPayload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	const auditQuery = `
INSERT INTO audit_logs (
	audit_id,
	actor_login,
	action,
	target_type,
	target_id,
	command_id,
	payload
) VALUES ($1, $2, 'service_action.requested', 'service', $3, $4, $5::jsonb)
RETURNING
	id,
	audit_id,
	actor_login,
	action,
	target_type,
	target_id,
	COALESCE(command_id, ''),
	payload,
	created_at`

	var auditLog domain.AuditLog
	var scannedAuditPayload []byte
	if err := tx.QueryRow(
		ctx,
		auditQuery,
		auditID,
		req.ActorLogin,
		req.ServiceID,
		command.CommandID,
		string(auditPayload),
	).Scan(
		&auditLog.ID,
		&auditLog.AuditID,
		&auditLog.ActorLogin,
		&auditLog.Action,
		&auditLog.TargetType,
		&auditLog.TargetID,
		&auditLog.CommandID,
		&scannedAuditPayload,
		&auditLog.CreatedAt,
	); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	if err := decodeJSONMap(scannedAuditPayload, &auditLog.Payload); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	return command, auditLog, false, nil
}

func (s *PostgresStore) GetCommand(ctx context.Context, commandID string) (domain.Command, error) {
	const query = `
SELECT
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at
FROM commands
WHERE command_id = $1
LIMIT 1`

	command, err := scanCommand(s.pool.QueryRow(ctx, query, commandID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Command{}, ErrNotFound
	}
	return command, err
}

func (s *PostgresStore) ListRunnableDryRunCommands(ctx context.Context, limit int) ([]domain.Command, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	const query = `
SELECT
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at
FROM commands
WHERE status = 'pending'
	AND dry_run = TRUE
	AND requires_approval = FALSE
ORDER BY created_at ASC, id ASC
LIMIT $1`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commands := []domain.Command{}
	for rows.Next() {
		command, err := scanCommand(rows)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return commands, nil
}

func (s *PostgresStore) UpdateCommandStatus(ctx context.Context, commandID, status string, resultPayload map[string]any) (domain.Command, error) {
	payload, err := json.Marshal(resultPayload)
	if err != nil {
		return domain.Command{}, err
	}

	const query = `
UPDATE commands
SET
	status = $2,
	result_payload = $3,
	updated_at = NOW()
WHERE command_id = $1
RETURNING
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at`

	command, err := scanCommand(s.pool.QueryRow(ctx, query, commandID, status, payload))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Command{}, ErrNotFound
	}
	return command, err
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

func (s *PostgresStore) ListRepositories(ctx context.Context, opts domain.ListOptions) ([]domain.Repository, error) {
	limit, offset := boundedList(opts)
	const query = `
SELECT
	id,
	COALESCE(gitea_repository_id, 0),
	full_name,
	COALESCE(owner_login, ''),
	name,
	COALESCE(clone_url, ''),
	COALESCE(html_url, ''),
	COALESCE(default_branch, ''),
	private,
	updated_at
FROM repositories
ORDER BY updated_at DESC, id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repositories := make([]domain.Repository, 0, limit)
	for rows.Next() {
		var repository domain.Repository
		if err := rows.Scan(
			&repository.ID,
			&repository.GiteaID,
			&repository.FullName,
			&repository.OwnerLogin,
			&repository.Name,
			&repository.CloneURL,
			&repository.HTMLURL,
			&repository.DefaultBranch,
			&repository.Private,
			&repository.UpdatedAt,
		); err != nil {
			return nil, err
		}
		repositories = append(repositories, repository)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return repositories, nil
}

func (s *PostgresStore) ListIssues(ctx context.Context, opts domain.ListOptions) ([]domain.Issue, error) {
	limit, offset := boundedList(opts)
	const query = `
SELECT
	i.id,
	COALESCE(i.gitea_issue_id, 0),
	COALESCE(r.gitea_repository_id, 0),
	r.full_name,
	i.number,
	i.title,
	i.state,
	COALESCE(i.author_login, ''),
	COALESCE(i.assignee_login, ''),
	COALESCE(i.html_url, ''),
	i.opened_at,
	i.closed_at,
	i.updated_at
FROM issues i
JOIN repositories r ON r.id = i.repository_id
WHERE ($3 = '' OR r.full_name = $3)
  AND ($4 = '' OR i.state = $4)
ORDER BY i.updated_at DESC, i.id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.RepositoryName, opts.State)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]domain.Issue, 0, limit)
	for rows.Next() {
		var issue domain.Issue
		if err := rows.Scan(
			&issue.ID,
			&issue.GiteaID,
			&issue.RepositoryGiteaID,
			&issue.RepositoryName,
			&issue.Number,
			&issue.Title,
			&issue.State,
			&issue.AuthorLogin,
			&issue.AssigneeLogin,
			&issue.HTMLURL,
			&issue.OpenedAt,
			&issue.ClosedAt,
			&issue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return issues, nil
}

func (s *PostgresStore) ListPullRequests(ctx context.Context, opts domain.ListOptions) ([]domain.PullRequest, error) {
	limit, offset := boundedList(opts)
	const query = `
SELECT
	pr.id,
	COALESCE(pr.gitea_pull_request_id, 0),
	COALESCE(r.gitea_repository_id, 0),
	r.full_name,
	pr.number,
	pr.title,
	pr.state,
	COALESCE(pr.author_login, ''),
	COALESCE(pr.head_branch, ''),
	COALESCE(pr.base_branch, ''),
	COALESCE(pr.head_sha, ''),
	COALESCE(pr.html_url, ''),
	pr.merged_at,
	pr.closed_at,
	pr.updated_at
FROM pull_requests pr
JOIN repositories r ON r.id = pr.repository_id
WHERE ($3 = '' OR r.full_name = $3)
  AND ($4 = '' OR pr.state = $4)
ORDER BY pr.updated_at DESC, pr.id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.RepositoryName, opts.State)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pullRequests := make([]domain.PullRequest, 0, limit)
	for rows.Next() {
		var pullRequest domain.PullRequest
		if err := rows.Scan(
			&pullRequest.ID,
			&pullRequest.GiteaID,
			&pullRequest.RepositoryGiteaID,
			&pullRequest.RepositoryName,
			&pullRequest.Number,
			&pullRequest.Title,
			&pullRequest.State,
			&pullRequest.AuthorLogin,
			&pullRequest.HeadBranch,
			&pullRequest.BaseBranch,
			&pullRequest.HeadSHA,
			&pullRequest.HTMLURL,
			&pullRequest.MergedAt,
			&pullRequest.ClosedAt,
			&pullRequest.UpdatedAt,
		); err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, pullRequest)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pullRequests, nil
}

func (s *PostgresStore) ListCIRuns(ctx context.Context, opts domain.ListOptions) ([]domain.CIRun, error) {
	limit, offset := boundedList(opts)
	const query = `
SELECT
	id,
	external_id,
	repository_name,
	COALESCE(branch, ''),
	COALESCE(commit_sha, ''),
	status,
	COALESCE(conclusion, ''),
	started_at,
	finished_at,
	duration_seconds,
	COALESCE(html_url, ''),
	updated_at
FROM ci_runs
WHERE ($3 = '' OR repository_name = $3)
  AND ($4 = '' OR status = $4)
ORDER BY COALESCE(started_at, updated_at) DESC, id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.RepositoryName, opts.Status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]domain.CIRun, 0, limit)
	for rows.Next() {
		var run domain.CIRun
		if err := rows.Scan(
			&run.ID,
			&run.ExternalID,
			&run.RepositoryName,
			&run.Branch,
			&run.CommitSHA,
			&run.Status,
			&run.Conclusion,
			&run.StartedAt,
			&run.FinishedAt,
			&run.DurationSeconds,
			&run.HTMLURL,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *PostgresStore) ListRisks(ctx context.Context, opts domain.ListOptions) ([]domain.Risk, error) {
	limit, offset := boundedList(opts)
	const query = `
SELECT
	id,
	risk_key,
	title,
	reason,
	impact,
	status,
	COALESCE(owner_login, ''),
	source_type,
	COALESCE(source_id, ''),
	suggested_actions,
	detected_at,
	mitigated_at,
	created_at,
	updated_at
FROM risks
WHERE ($3 = '' OR status = $3)
  AND ($4 = '' OR impact = $4)
ORDER BY
	CASE impact
		WHEN 'critical' THEN 0
		WHEN 'high' THEN 1
		WHEN 'medium' THEN 2
		ELSE 3
	END,
	updated_at DESC,
	id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset, opts.Status, opts.Impact)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	risks := make([]domain.Risk, 0, limit)
	for rows.Next() {
		var risk domain.Risk
		var suggestedActions []byte
		if err := rows.Scan(
			&risk.ID,
			&risk.RiskKey,
			&risk.Title,
			&risk.Reason,
			&risk.Impact,
			&risk.Status,
			&risk.OwnerLogin,
			&risk.SourceType,
			&risk.SourceID,
			&suggestedActions,
			&risk.DetectedAt,
			&risk.MitigatedAt,
			&risk.CreatedAt,
			&risk.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(suggestedActions) > 0 {
			if err := json.Unmarshal(suggestedActions, &risk.SuggestedActions); err != nil {
				return nil, err
			}
		}
		risks = append(risks, risk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return risks, nil
}

func boundedList(opts domain.ListOptions) (int, int) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (s *PostgresStore) ensureRiskExists(ctx context.Context, riskID string) error {
	const query = `SELECT 1 FROM risks WHERE risk_key = $1 LIMIT 1`
	var exists int
	if err := s.pool.QueryRow(ctx, query, riskID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PostgresStore) commandByIdempotencyKey(ctx context.Context, idempotencyKey, commandType string) (domain.Command, domain.AuditLog, bool, error) {
	const commandQuery = `
SELECT
	id,
	command_id,
	command_type,
	target_type,
	target_id,
	action_type,
	status,
	actor_login,
	reason,
	dry_run,
	requires_approval,
	COALESCE(idempotency_key, ''),
	request_payload,
	result_payload,
	created_at,
	updated_at
FROM commands
WHERE idempotency_key = $1 AND command_type = $2
LIMIT 1`

	command, err := scanCommand(s.pool.QueryRow(ctx, commandQuery, idempotencyKey, commandType))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Command{}, domain.AuditLog{}, false, nil
	}
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}

	auditLog, err := s.auditLogForCommand(ctx, command.CommandID)
	if err != nil {
		return domain.Command{}, domain.AuditLog{}, false, err
	}
	return command, auditLog, true, nil
}

func (s *PostgresStore) auditLogForCommand(ctx context.Context, commandID string) (domain.AuditLog, error) {
	const query = `
SELECT
	id,
	audit_id,
	actor_login,
	action,
	target_type,
	target_id,
	COALESCE(command_id, ''),
	payload,
	created_at
FROM audit_logs
WHERE command_id = $1
ORDER BY created_at ASC, id ASC
LIMIT 1`

	var auditLog domain.AuditLog
	var payload []byte
	err := s.pool.QueryRow(ctx, query, commandID).Scan(
		&auditLog.ID,
		&auditLog.AuditID,
		&auditLog.ActorLogin,
		&auditLog.Action,
		&auditLog.TargetType,
		&auditLog.TargetID,
		&auditLog.CommandID,
		&payload,
		&auditLog.CreatedAt,
	)
	if err != nil {
		return domain.AuditLog{}, err
	}
	if err := decodeJSONMap(payload, &auditLog.Payload); err != nil {
		return domain.AuditLog{}, err
	}
	return auditLog, nil
}

func scanCommand(row pgx.Row) (domain.Command, error) {
	var command domain.Command
	var requestPayload []byte
	var resultPayload []byte
	err := row.Scan(
		&command.ID,
		&command.CommandID,
		&command.CommandType,
		&command.TargetType,
		&command.TargetID,
		&command.ActionType,
		&command.Status,
		&command.ActorLogin,
		&command.Reason,
		&command.DryRun,
		&command.RequiresApproval,
		&command.IdempotencyKey,
		&requestPayload,
		&resultPayload,
		&command.CreatedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		return domain.Command{}, err
	}
	if err := decodeJSONMap(requestPayload, &command.RequestPayload); err != nil {
		return domain.Command{}, err
	}
	if err := decodeJSONMap(resultPayload, &command.ResultPayload); err != nil {
		return domain.Command{}, err
	}
	return command, nil
}

func decodeJSONMap(payload []byte, target *map[string]any) error {
	if len(payload) == 0 {
		*target = map[string]any{}
		return nil
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}
	if *target == nil {
		*target = map[string]any{}
	}
	return nil
}

func randomPrefixedID(prefix string) (string, error) {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(bytes[:]), nil
}
