package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
)

// ErrVocExternalRefConflict는 (source_system, external_ref) UNIQUE 위반.
// ADR-0028 §3: idempotency 보장을 위해 사전 SELECT + 200 반환으로 흡수하나,
// 동시성 race 시 이 에러를 받음.
var ErrVocExternalRefConflict = errors.New("dev_request_voc: external_ref conflict")

// ErrVocAlreadyRouted는 동시성 routing race 시 UPDATE 가 no-op (이미 routed/closed).
var ErrVocAlreadyRouted = errors.New("dev_request_voc: already routed or closed")

// DevRequestVocRepository는 dev_request_vocs table CRUD. ADR-0028 §3.
type DevRequestVocRepository struct {
	store *store.PostgresStore
}

func NewDevRequestVocRepository(s *store.PostgresStore) *DevRequestVocRepository {
	return &DevRequestVocRepository{store: s}
}

const devRequestVocSelectColumns = `id, external_ref, source_system, title, details, requester,
    req_department, assignee_user_id, dev_department, request_date, dev_schedule,
    status, project_id, dev_request_id, routed_at, created_at, updated_at`

func scanDevRequestVoc(row pgx.Row) (domain.DevRequestVoc, error) {
	var v domain.DevRequestVoc
	var (
		reqDept      string
		devDept      string
		details      string
		assignee     *string
		extRef       string
		projectID    *string
		devRequestID *string
		status       string
	)
	if err := row.Scan(
		&v.ID, &extRef, &v.SourceSystem, &v.Title, &details, &v.Requester,
		&reqDept, &assignee, &devDept, &v.RequestDate, &v.DevSchedule,
		&status, &projectID, &devRequestID, &v.RoutedAt, &v.CreatedAt, &v.UpdatedAt,
	); err != nil {
		return domain.DevRequestVoc{}, err
	}
	v.ExternalRef = extRef
	v.Details = details
	v.ReqDepartment = reqDept
	if assignee != nil {
		v.AssigneeUserID = *assignee
	}
	v.DevDepartment = devDept
	v.Status = domain.DevRequestVocStatus(status)
	if projectID != nil {
		v.ProjectID = *projectID
	}
	if devRequestID != nil {
		v.DevRequestID = *devRequestID
	}
	return v, nil
}

// CreateVoc는 새 voc 등록. (source_system, external_ref) UNIQUE 위반 시
// ErrVocExternalRefConflict. ADR-0028 §3: 외부 API 는 사전 Get 으로
// idempotent 처리하지만, 동시성 race 시 이 에러로 fallback.
func (r *DevRequestVocRepository) CreateVoc(ctx context.Context, v domain.DevRequestVoc) (domain.DevRequestVoc, error) {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	v.UpdatedAt = v.CreatedAt
	const insertQuery = `
INSERT INTO dev_request_vocs (
    external_ref, source_system, title, details, requester,
    req_department, assignee_user_id, dev_department, request_date, dev_schedule,
    status, project_id, dev_request_id, routed_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, NULLIF($7, ''), $8, $9, $10,
    $11, $12::uuid, $13::uuid, $14, $15, $16
)
RETURNING ` + devRequestVocSelectColumns
	row := r.store.Pool().QueryRow(ctx, insertQuery,
		v.ExternalRef, v.SourceSystem, v.Title, v.Details, v.Requester,
		v.ReqDepartment, v.AssigneeUserID, v.DevDepartment, v.RequestDate, v.DevSchedule,
		string(v.Status), nullableString(v.ProjectID), nullableString(v.DevRequestID), v.RoutedAt,
		v.CreatedAt, v.UpdatedAt,
	)
	created, err := scanDevRequestVoc(row)
	if err != nil {
		// pgx unique violation detection: 23505 SQLSTATE
		if isUniqueViolation(err) {
			return domain.DevRequestVoc{}, ErrVocExternalRefConflict
		}
		return domain.DevRequestVoc{}, fmt.Errorf("create dev_request_voc: %w", err)
	}
	return created, nil
}

// GetVocByExternalRef는 (source_system, external_ref) UNIQUE lookup.
// ADR-0028 §3: idempotency 의 1차 시도 (사전 SELECT).
func (r *DevRequestVocRepository) GetVocByExternalRef(ctx context.Context, sourceSystem, externalRef string) (domain.DevRequestVoc, bool, error) {
	query := `SELECT ` + devRequestVocSelectColumns + ` FROM dev_request_vocs WHERE source_system = $1 AND external_ref = $2`
	row := r.store.Pool().QueryRow(ctx, query, sourceSystem, externalRef)
	v, err := scanDevRequestVoc(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DevRequestVoc{}, false, nil
		}
		return domain.DevRequestVoc{}, false, fmt.Errorf("get dev_request_voc: %w", err)
	}
	return v, true, nil
}

// GetVocByID는 단건 조회.
func (r *DevRequestVocRepository) GetVocByID(ctx context.Context, id string) (domain.DevRequestVoc, error) {
	query := `SELECT ` + devRequestVocSelectColumns + ` FROM dev_request_vocs WHERE id = $1::uuid`
	row := r.store.Pool().QueryRow(ctx, query, id)
	return scanDevRequestVoc(row)
}

// RouteVoc는 voc 의 project_id 결정 + dev-request 자동 생성 + status=routed.
// ADR-0028 §3 결정. 단일 트랜잭션 (voc update + dev-request insert).
func (r *DevRequestVocRepository) RouteVoc(ctx context.Context, vocID, projectID string, dr domain.DevRequest) (domain.DevRequestVoc, domain.DevRequest, error) {
	tx, err := r.store.Pool().Begin(ctx)
	if err != nil {
		return domain.DevRequestVoc{}, domain.DevRequest{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1) dev-request insert (voc 의 9 field 복사 + project_id 결정)
	now := time.Now().UTC()
	if dr.ReceivedAt.IsZero() {
		dr.ReceivedAt = now
	}
	dr.CreatedAt = now
	dr.UpdatedAt = now
	const insertDr = `
INSERT INTO dev_requests (
    title, details, requester, req_department, assignee_user_id, dev_department,
    request_date, dev_schedule, source_system, external_ref, status,
    registered_target_type, registered_target_id, rejected_reason, received_at, created_at, updated_at
) VALUES (
    $1, NULLIF($2, ''), $3, $4, NULLIF($5, ''), $6,
    $7, $8, $9, $10, $11,
    NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''), $15, $16, $17
)
RETURNING id`
	var devRequestID string
	row := tx.QueryRow(ctx, insertDr,
		dr.Title, dr.Details, dr.Requester, dr.ReqDepartment, dr.AssigneeUserID, dr.DevDepartment,
		dr.RequestDate, dr.DevSchedule, dr.SourceSystem, dr.ExternalRef, string(dr.Status),
		string(dr.RegisteredTargetType), dr.RegisteredTargetID, dr.RejectedReason,
		dr.ReceivedAt, dr.CreatedAt, dr.UpdatedAt,
	)
	if err := row.Scan(&devRequestID); err != nil {
		return domain.DevRequestVoc{}, domain.DevRequest{}, fmt.Errorf("insert dev_request: %w", err)
	}

	// 2) voc update: status=routed, project_id, dev_request_id, routed_at
	// 동시성 race 방지: status='received' predicate — 두 동시 routing 시 한쪽만 성공.
	const updateVoc = `
UPDATE dev_request_vocs SET
    status = 'routed',
    project_id = $2::uuid,
    dev_request_id = $3::uuid,
    routed_at = $4,
    updated_at = $4
WHERE id = $1::uuid AND status = 'received'
RETURNING ` + devRequestVocSelectColumns
	v, err := scanDevRequestVoc(tx.QueryRow(ctx, updateVoc, vocID, projectID, devRequestID, now))
	if errors.Is(err, pgx.ErrNoRows) {
		// race: 다른 transaction 이 먼저 routed 로 변경.
		return domain.DevRequestVoc{}, domain.DevRequest{}, ErrVocAlreadyRouted
	}
	if err != nil {
		return domain.DevRequestVoc{}, domain.DevRequest{}, fmt.Errorf("update voc: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.DevRequestVoc{}, domain.DevRequest{}, fmt.Errorf("commit: %w", err)
	}
	dr.ID = devRequestID
	return v, dr, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func isUniqueViolation(err error) bool {
	// pgx error wrapping: type assertion to interface{ SQLState() string }
	type sqlStater interface{ SQLState() string }
	var s sqlStater
	if errors.As(err, &s) {
		return s.SQLState() == "23505"
	}
	return false
}
