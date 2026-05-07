package store

import (
	"context"
	"encoding/json"
	"fmt"

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
