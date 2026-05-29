package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
)

// CreateAuditLog inserts an audit_logs row without an associated command.
// The audit_id is generated with a "audit_" prefix when one is not provided
// by the caller, mirroring the convention used by command-driven audit logs.
func (r *AuditRepository) CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error) {
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
		auditID, err = store.RandomPrefixedID("audit")
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
INSERT INTO audit_logs (audit_id, actor_login, action, target_type, target_id, command_id, payload, source_ip, request_id, source_type, source_event_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''))
ON CONFLICT (source_type, source_event_id) WHERE source_event_id IS NOT NULL AND source_type IS NOT NULL DO NOTHING
RETURNING id, audit_id, actor_login, action, target_type, target_id, COALESCE(command_id, ''), payload, COALESCE(source_ip, ''), COALESCE(request_id, ''), COALESCE(source_type, ''), COALESCE(source_event_id, ''), created_at`

	var inserted domain.AuditLog
	var payloadJSON []byte
	var scannedSourceType string
	err = r.store.Pool().QueryRow(ctx, query,
		auditID,
		actor,
		log.Action,
		log.TargetType,
		log.TargetID,
		commandIDArg,
		payloadBytes,
		log.SourceIP,
		log.RequestID,
		string(log.SourceType),
		log.SourceEventID,
	).Scan(
		&inserted.ID,
		&inserted.AuditID,
		&inserted.ActorLogin,
		&inserted.Action,
		&inserted.TargetType,
		&inserted.TargetID,
		&inserted.CommandID,
		&payloadJSON,
		&inserted.SourceIP,
		&inserted.RequestID,
		&scannedSourceType,
		&inserted.SourceEventID,
		&inserted.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// ON CONFLICT DO NOTHING — dedup 차단됨. 기존 row 를 SELECT 해서 반환.
		existing, lookupErr := r.getAuditLogBySourceEventID(ctx, string(log.SourceType), log.SourceEventID)
		if lookupErr != nil {
			return domain.AuditLog{}, fmt.Errorf("audit_logs dedup conflict but existing lookup failed: %w", lookupErr)
		}
		return existing, nil
	}
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("insert audit log: %w", err)
	}
	inserted.SourceType = domain.AuditSourceType(scannedSourceType)
	if len(payloadJSON) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(payloadJSON, &decoded); err != nil {
			return domain.AuditLog{}, fmt.Errorf("decode audit payload: %w", err)
		}
		inserted.Payload = decoded
	}
	return inserted, nil
}

// getAuditLogBySourceEventID — ON CONFLICT DO NOTHING dedup 시 기존 row 조회.
func (r *AuditRepository) getAuditLogBySourceEventID(ctx context.Context, sourceType, sourceEventID string) (domain.AuditLog, error) {
	if sourceEventID == "" {
		return domain.AuditLog{}, errors.New("getAuditLogBySourceEventID: sourceEventID must be non-empty")
	}
	const q = `
SELECT id, audit_id, actor_login, action, target_type, target_id, COALESCE(command_id, ''), payload, COALESCE(source_ip, ''), COALESCE(request_id, ''), COALESCE(source_type, ''), COALESCE(source_event_id, ''), created_at
FROM audit_logs
WHERE source_type = $1 AND source_event_id = $2
LIMIT 1`
	var existing domain.AuditLog
	var payloadJSON []byte
	var scannedSourceType string
	err := r.store.Pool().QueryRow(ctx, q, sourceType, sourceEventID).Scan(
		&existing.ID,
		&existing.AuditID,
		&existing.ActorLogin,
		&existing.Action,
		&existing.TargetType,
		&existing.TargetID,
		&existing.CommandID,
		&payloadJSON,
		&existing.SourceIP,
		&existing.RequestID,
		&scannedSourceType,
		&existing.SourceEventID,
		&existing.CreatedAt,
	)
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("select existing audit_log (%s, %s): %w", sourceType, sourceEventID, err)
	}
	existing.SourceType = domain.AuditSourceType(scannedSourceType)
	if len(payloadJSON) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(payloadJSON, &decoded); err != nil {
			return domain.AuditLog{}, fmt.Errorf("decode existing audit payload: %w", err)
		}
		existing.Payload = decoded
	}
	return existing, nil
}

func (r *AuditRepository) ListAuditLogs(ctx context.Context, opts store.ListAuditLogsOptions) ([]domain.AuditLog, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const query = `
SELECT id, audit_id, actor_login, action, target_type, target_id, COALESCE(command_id, ''), payload, COALESCE(source_ip, ''), COALESCE(request_id, ''), COALESCE(source_type, ''), COALESCE(source_event_id, ''), created_at
FROM audit_logs
WHERE ($1 = '' OR actor_login = $1)
  AND ($2 = '' OR action = $2)
  AND ($3 = '' OR target_type = $3)
  AND ($4 = '' OR target_id = $4)
  AND ($5 = '' OR command_id = $5)
ORDER BY created_at DESC, id DESC
LIMIT $6 OFFSET $7`

	rows, err := r.store.Pool().Query(ctx, query,
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
		var scannedSourceType string
		if err := rows.Scan(
			&log.ID,
			&log.AuditID,
			&log.ActorLogin,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&log.CommandID,
			&payloadJSON,
			&log.SourceIP,
			&log.RequestID,
			&scannedSourceType,
			&log.SourceEventID,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		log.SourceType = domain.AuditSourceType(scannedSourceType)
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
