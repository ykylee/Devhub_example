package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// DevRequestListOptions parameterizes ListDevRequests.
type DevRequestListOptions struct {
	// Status 가 비어 있으면 모든 status. 콤마 다중은 caller 가 split.
	Statuses []domain.DevRequestStatus
	// AssigneeUserID 가 비어 있으면 모든 assignee.
	AssigneeUserID string
	// SourceSystem 이 비어 있으면 모든 source.
	SourceSystem string
	Limit        int
	Offset       int
}

const devRequestsSelectColumns = `
	id::text,
	title,
	COALESCE(details, ''),
	requester,
	COALESCE(assignee_user_id, ''),
	source_system,
	COALESCE(external_ref, ''),
	status,
	COALESCE(registered_target_type, ''),
	COALESCE(registered_target_id, ''),
	COALESCE(rejected_reason, ''),
	received_at,
	created_at,
	updated_at`

func scanDevRequest(row pgx.Row) (domain.DevRequest, error) {
	var dr domain.DevRequest
	var (
		regType        string
		regTargetID    string
		rejectedReason string
		externalRef    string
		details        string
		status         string
	)
	if err := row.Scan(
		&dr.ID,
		&dr.Title,
		&details,
		&dr.Requester,
		&dr.AssigneeUserID,
		&dr.SourceSystem,
		&externalRef,
		&status,
		&regType,
		&regTargetID,
		&rejectedReason,
		&dr.ReceivedAt,
		&dr.CreatedAt,
		&dr.UpdatedAt,
	); err != nil {
		return domain.DevRequest{}, err
	}
	dr.Details = details
	dr.ExternalRef = externalRef
	dr.Status = domain.DevRequestStatus(status)
	if regType != "" {
		dr.RegisteredTargetType = domain.DevRequestTargetType(regType)
	}
	dr.RegisteredTargetID = regTargetID
	dr.RejectedReason = rejectedReason
	return dr, nil
}

// CreateDevRequest는 외부 수신 endpoint 에서 호출. idempotency 충돌 시 ErrConflict.
// 검증 실패한 의뢰(rejected, invalid_intake)도 audit 보존 목적으로 저장하므로
// caller 가 의뢰 row 자체는 무조건 insert 한다.
func (s *PostgresStore) CreateDevRequest(ctx context.Context, dr domain.DevRequest) (domain.DevRequest, error) {
	if dr.ReceivedAt.IsZero() {
		dr.ReceivedAt = time.Now().UTC()
	}
	const insertQuery = `
INSERT INTO dev_requests (
    title, details, requester, assignee_user_id, source_system, external_ref,
    status, registered_target_type, registered_target_id, rejected_reason, received_at
)
VALUES (
    $1, NULLIF($2, ''), $3, NULLIF($4, ''), $5, NULLIF($6, ''),
    $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11
)
RETURNING` + devRequestsSelectColumns

	row := s.pool.QueryRow(ctx, insertQuery,
		dr.Title, dr.Details, dr.Requester, dr.AssigneeUserID, dr.SourceSystem, dr.ExternalRef,
		string(dr.Status), string(dr.RegisteredTargetType), dr.RegisteredTargetID, dr.RejectedReason,
		dr.ReceivedAt,
	)
	created, err := scanDevRequest(row)
	if isUniqueViolation(err) {
		return domain.DevRequest{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return domain.DevRequest{}, ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("create dev_request: %w", err)
	}
	return created, nil
}

// GetDevRequest 는 단일 의뢰 조회.
func (s *PostgresStore) GetDevRequest(ctx context.Context, id string) (domain.DevRequest, error) {
	query := `SELECT` + devRequestsSelectColumns + ` FROM dev_requests WHERE id = $1::uuid`
	row := s.pool.QueryRow(ctx, query, id)
	dr, err := scanDevRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("get dev_request: %w", err)
	}
	return dr, nil
}

// GetDevRequestByExternalRef는 idempotency 검증용. (source_system, external_ref) UNIQUE 인덱스 사용.
func (s *PostgresStore) GetDevRequestByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequest, error) {
	if externalRef == "" {
		return domain.DevRequest{}, ErrNotFound
	}
	query := `SELECT` + devRequestsSelectColumns + ` FROM dev_requests WHERE source_system = $1 AND external_ref = $2`
	row := s.pool.QueryRow(ctx, query, sourceSystem, externalRef)
	dr, err := scanDevRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("lookup dev_request by external_ref: %w", err)
	}
	return dr, nil
}

// ListDevRequests 는 목록 조회. caller 가 row-level filter (assignee) 를 적용한다.
func (s *PostgresStore) ListDevRequests(ctx context.Context, opts DevRequestListOptions) ([]domain.DevRequest, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	// status 배열 → array, 빈 배열은 모든 status.
	statusStrings := make([]string, 0, len(opts.Statuses))
	for _, st := range opts.Statuses {
		statusStrings = append(statusStrings, string(st))
	}

	const countQuery = `
SELECT COUNT(*) FROM dev_requests
WHERE (cardinality($1::text[]) = 0 OR status = ANY($1::text[]))
  AND ($2 = '' OR assignee_user_id = $2)
  AND ($3 = '' OR source_system = $3)`

	var total int
	if err := s.pool.QueryRow(ctx, countQuery, statusStrings, opts.AssigneeUserID, opts.SourceSystem).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count dev_requests: %w", err)
	}

	const listQuery = `
SELECT` + devRequestsSelectColumns + `
FROM dev_requests
WHERE (cardinality($1::text[]) = 0 OR status = ANY($1::text[]))
  AND ($2 = '' OR assignee_user_id = $2)
  AND ($3 = '' OR source_system = $3)
ORDER BY received_at DESC
LIMIT $4 OFFSET $5`

	rows, err := s.pool.Query(ctx, listQuery, statusStrings, opts.AssigneeUserID, opts.SourceSystem, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list dev_requests: %w", err)
	}
	defer rows.Close()

	drs := make([]domain.DevRequest, 0, limit)
	for rows.Next() {
		dr, err := scanDevRequest(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan dev_request: %w", err)
		}
		drs = append(drs, dr)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate dev_requests: %w", err)
	}
	return drs, total, nil
}

// TransitionDevRequestStatus는 단순 상태 전이 (rejected / closed / reopen 등).
// register 는 별도 RegisterDevRequest 에서 트랜잭션 처리.
func (s *PostgresStore) TransitionDevRequestStatus(ctx context.Context, id string, to domain.DevRequestStatus, rejectedReason string) (domain.DevRequest, error) {
	const updateQuery = `
UPDATE dev_requests SET
    status = $2,
    rejected_reason = CASE WHEN $2 = 'rejected' THEN NULLIF($3, '') ELSE NULL END,
    registered_target_type = CASE WHEN $2 = 'registered' THEN registered_target_type ELSE NULL END,
    registered_target_id   = CASE WHEN $2 = 'registered' THEN registered_target_id   ELSE NULL END,
    updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + devRequestsSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery, id, string(to), rejectedReason)
	updated, err := scanDevRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("transition dev_request: %w", err)
	}
	return updated, nil
}

// ReassignDevRequest는 system_admin 의 담당자 변경. handler 가 role 검증.
func (s *PostgresStore) ReassignDevRequest(ctx context.Context, id, newAssigneeUserID string) (domain.DevRequest, error) {
	const updateQuery = `
UPDATE dev_requests SET
    assignee_user_id = $2,
    updated_at = NOW()
WHERE id = $1::uuid
RETURNING` + devRequestsSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery, id, newAssigneeUserID)
	updated, err := scanDevRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, ErrNotFound
	}
	if isForeignKeyViolation(err) {
		return domain.DevRequest{}, ErrConflict
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("reassign dev_request: %w", err)
	}
	return updated, nil
}

// MarkDevRequestRegistered는 RegisterDevRequest 트랜잭션 안에서 호출되는 일부 동작.
// (target 생성과 함께 단일 트랜잭션. application/project handler 가 store 호출
// 시점에 이미 row 가 생성되어 있어야 하므로, 본 store 는 status 갱신만 책임)
// caller 가 target row 생성 후 status='registered' + target_type/id 갱신을 위임.
//
// codex hotfix #4 / self-review P2 #1 (sprint claude/work_260515-n) —
// rejected_reason 을 명시적으로 NULL 로 비워 reopen 후 promote 흐름의 잔재
// 차단. dreqMarkRegisteredUpdateQuery (promote transaction) 과 동일 정책.
func (s *PostgresStore) MarkDevRequestRegistered(ctx context.Context, id string, targetType domain.DevRequestTargetType, targetID string) (domain.DevRequest, error) {
	const updateQuery = `
UPDATE dev_requests SET
    status = 'registered',
    registered_target_type = $2,
    registered_target_id   = $3,
    rejected_reason        = NULL,
    updated_at = NOW()
WHERE id = $1::uuid AND status IN ('pending', 'in_review')
RETURNING` + devRequestsSelectColumns

	row := s.pool.QueryRow(ctx, updateQuery, id, string(targetType), targetID)
	updated, err := scanDevRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequest{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequest{}, fmt.Errorf("mark dev_request registered: %w", err)
	}
	return updated, nil
}
