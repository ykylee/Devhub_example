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

// UserNotificationRepository는 user_notifications table CRUD. ADR-0028 §3.
type UserNotificationRepository struct {
	store *store.PostgresStore
}

func NewUserNotificationRepository(s *store.PostgresStore) *UserNotificationRepository {
	return &UserNotificationRepository{store: s}
}

const userNotificationSelectColumns = `id, user_id, kind, ref_id, title, body, read_at, created_at`

func scanUserNotification(row pgx.Row) (domain.UserNotification, error) {
	var n domain.UserNotification
	var (
		refID string
		read  *time.Time
	)
	if err := row.Scan(&n.ID, &n.UserID, &n.Kind, &refID, &n.Title, &n.Body, &read, &n.CreatedAt); err != nil {
		return domain.UserNotification{}, err
	}
	n.RefID = refID
	n.ReadAt = read
	return n, nil
}

// InsertNotification은 in-app notification 1건 등록. ADR-0028 §3.
func (r *UserNotificationRepository) InsertNotification(ctx context.Context, n domain.UserNotification) (domain.UserNotification, error) {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	const insertQuery = `
INSERT INTO user_notifications (user_id, kind, ref_id, title, body, read_at, created_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, ''), $7)
RETURNING ` + userNotificationSelectColumns
	row := r.store.Pool().QueryRow(ctx, insertQuery,
		n.UserID, string(n.Kind), n.RefID, n.Title, n.Body, n.ReadAt, n.CreatedAt,
	)
	created, err := scanUserNotification(row)
	if err != nil {
		return domain.UserNotification{}, fmt.Errorf("insert user_notification: %w", err)
	}
	return created, nil
}

// ListUnreadByUser는 미읽음 알림 최신순. ADR-0028 §3.
func (r *UserNotificationRepository) ListUnreadByUser(ctx context.Context, userID string, limit int) ([]domain.UserNotification, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const listQuery = `SELECT ` + userNotificationSelectColumns + ` FROM user_notifications WHERE user_id = $1 AND read_at IS NULL ORDER BY created_at DESC LIMIT $2`
	rows, err := r.store.Pool().Query(ctx, listQuery, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list unread user_notifications: %w", err)
	}
	defer rows.Close()
	var out []domain.UserNotification
	for rows.Next() {
		n, err := scanUserNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// MarkRead는 user_notification 의 read_at SET. ADR-0028 §3.
func (r *UserNotificationRepository) MarkRead(ctx context.Context, id, userID string) error {
	// user_id 검증: 타인의 notification 을 mark 하면 no-op (404 반환).
	const updateQuery = `
UPDATE user_notifications SET read_at = $3
WHERE id = $1::uuid AND user_id = $2 AND read_at IS NULL`
	tag, err := r.store.Pool().Exec(ctx, updateQuery, id, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// not found or already read or not owned
		return ErrNotificationNotFound
	}
	return nil
}

// ErrNotificationNotFound는 mark read 시 대상 부재.
var ErrNotificationNotFound = errors.New("user_notification: not found or not owned")
