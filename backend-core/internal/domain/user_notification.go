package domain

import "time"

// UserNotificationKind는 in-app notification 의 kind enum. ADR-0028 §3 결정.
//
//	kind         용도                                          ref_id
//	dev_voc      dev_request_vocs.id (voc 도착, 담당자 알림)        dev_request_voc_id
//	dev_request  dev_requests.id (dev-request 등록/업데이트)        dev_request_id
//	mention      user mention (post-MVP)                              N/A
type UserNotificationKind string

const (
	// UserNotificationKindDevVoc는 dev_request_vocs 의 새 row 등록 시
	// assignee 에게 발송. ref_id = dev_request_voc_id (string).
	UserNotificationKindDevVoc UserNotificationKind = "dev_voc"
	// UserNotificationKindDevRequest는 dev_request 의 routing 완료 시
	// assignee 에게 발송. ref_id = dev_request_id (string).
	UserNotificationKindDevRequest UserNotificationKind = "dev_request"
)

// UserNotification은 in-app notification row. ADR-0028 §3 결정.
// user_id → users.user_id (FK, ON DELETE CASCADE).
// read_at IS NULL → 미읽음, IS NOT NULL → 읽음. unread 조회 index 분리.
type UserNotification struct {
	ID        string                `json:"id"`
	UserID    string                `json:"user_id"`
	Kind      UserNotificationKind  `json:"kind"`
	RefID     string                `json:"ref_id,omitempty"`
	Title     string                `json:"title"`
	Body      string                `json:"body"`
	ReadAt    *time.Time            `json:"read_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}
