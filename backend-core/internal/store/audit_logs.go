package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
)

// CreateAuditLog inserts an audit_logs row without an associated command.
// The audit_id is generated with a "audit_" prefix when one is not provided
// by the caller, mirroring the convention used by command-driven audit logs.
func (s *PostgresStore) CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	payload := log.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("marshal audit payload: %w", err)
	}

	auditID := log.AuditID
	if auditID == "" {
		auditID, err = randomPrefixedID("audit")
		if err != nil {
			return domain.AuditLog{}, fmt.Errorf("generate audit id: %w", err)
		}
	}

	actor := log.ActorLogin
	if actor == "" {
		actor = "system"
	}

	var commandIDArg any
	if log.CommandID != "" {
		commandIDArg = log.CommandID
	} else {
		commandIDArg = nil
	}

	const query = `
INSERT INTO audit_logs (audit_id, actor_login, action, target_type, target_id, command_id, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, audit_id, actor_login, action, target_type, target_id, COALESCE(command_id, ''), payload, created_at`

	var inserted domain.AuditLog
	var payloadJSON []byte
	err = s.pool.QueryRow(ctx, query,
		auditID,
		actor,
		log.Action,
		log.TargetType,
		log.TargetID,
		commandIDArg,
		payloadBytes,
	).Scan(
		&inserted.ID,
		&inserted.AuditID,
		&inserted.ActorLogin,
		&inserted.Action,
		&inserted.TargetType,
		&inserted.TargetID,
		&inserted.CommandID,
		&payloadJSON,
		&inserted.CreatedAt,
	)
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("insert audit log: %w", err)
	}
	if len(payloadJSON) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(payloadJSON, &decoded); err != nil {
			return domain.AuditLog{}, fmt.Errorf("decode audit payload: %w", err)
		}
		inserted.Payload = decoded
	}
	return inserted, nil
}

type ListAuditLogsOptions struct {
	Limit      int
	Offset     int
	ActorLogin string
	Action     string
	TargetType string
	TargetID   string
	CommandID  string
}

func (s *PostgresStore) ListAuditLogs(ctx context.Context, opts ListAuditLogsOptions) ([]domain.AuditLog, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const query = `
SELECT id, audit_id, actor_login, action, target_type, target_id, COALESCE(command_id, ''), payload, created_at
FROM audit_logs
WHERE ($1 = '' OR actor_login = $1)
  AND ($2 = '' OR action = $2)
  AND ($3 = '' OR target_type = $3)
  AND ($4 = '' OR target_id = $4)
  AND ($5 = '' OR command_id = $5)
ORDER BY created_at DESC, id DESC
LIMIT $6 OFFSET $7`

	rows, err := s.pool.Query(ctx, query,
		strings.TrimSpace(opts.ActorLogin),
		strings.TrimSpace(opts.Action),
		strings.TrimSpace(opts.TargetType),
		strings.TrimSpace(opts.TargetID),
		strings.TrimSpace(opts.CommandID),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.AuditLog, 0, limit)
	for rows.Next() {
		var log domain.AuditLog
		var payloadJSON []byte
		if err := rows.Scan(
			&log.ID,
			&log.AuditID,
			&log.ActorLogin,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&log.CommandID,
			&payloadJSON,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(payloadJSON) > 0 {
			var decoded map[string]any
			if err := json.Unmarshal(payloadJSON, &decoded); err != nil {
				return nil, fmt.Errorf("decode audit payload: %w", err)
			}
			log.Payload = decoded
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
