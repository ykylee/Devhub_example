package domain

import "time"

// DevRequestVocStatus는 voc 의 3-state lifecycle. ADR-0028 §3 결정.
//
//	from / to     received routed closed
//	received        -       ✓       ✓
//	routed          -       -       ✓
//	closed          -       -       -
type DevRequestVocStatus string

const (
	// DevRequestVocStatusReceived은 초기 접수. project_id NULL, dev_request_id NULL.
	// 알림: assignee 가 존재하면 in-app notification 발송.
	DevRequestVocStatusReceived DevRequestVocStatus = "received"
	// DevRequestVocStatusRouted는 project 결정 + dev-request 자동 생성 후.
	// project_id SET, dev_request_id SET, routed_at SET.
	DevRequestVocStatusRouted DevRequestVocStatus = "routed"
	// DevRequestVocStatusClosed는 명시적 close. 후속 promote 없음.
	DevRequestVocStatusClosed DevRequestVocStatus = "closed"
)

// DevRequestVocTransitions은 status 머신의 허용 전이 표. ADR-0028 §3 결정.
var DevRequestVocTransitions = map[DevRequestVocStatus]map[DevRequestVocStatus]bool{
	DevRequestVocStatusReceived: {
		DevRequestVocStatusRouted: true,
		DevRequestVocStatusClosed: true,
	},
	DevRequestVocStatusRouted: {
		DevRequestVocStatusClosed: true,
	},
	DevRequestVocStatusClosed: {},
}

// DevRequestVoc는 외부 시스템 의 의뢰 (voc = Voice of Customer) 가 project 결정
// 전 staging 단계의 row. ADR-0028 §3 결정. voc 의 9 field (title/details/requester/
// req_department/assignee/dev_department/request_date/dev_schedule/external_ref) +
// source_system (인증된 intake token 의 매핑값, ADR-0012 §4.1.2 spoofing 방지) +
// status 머신 (received/routed/closed) + project_id (routed 후 SET, NULL 허용) +
// dev_request_id (routed 후 자동 생성된 dev-request 의 FK, NULL 허용) + routed_at.
type DevRequestVoc struct {
	ID             string              `json:"id"`
	ExternalRef    string              `json:"external_ref"`
	SourceSystem   string              `json:"source_system"`
	Title          string              `json:"title"`
	Details        string              `json:"details"`
	Requester      string              `json:"requester"`
	ReqDepartment  string              `json:"req_department"`
	AssigneeUserID string              `json:"assignee_user_id,omitempty"`
	DevDepartment  string              `json:"dev_department"`
	RequestDate    *time.Time          `json:"request_date,omitempty"`
	DevSchedule    string              `json:"dev_schedule"`
	Status         DevRequestVocStatus `json:"status"`
	ProjectID      string              `json:"project_id,omitempty"`
	DevRequestID   string              `json:"dev_request_id,omitempty"`
	RoutedAt       *time.Time          `json:"routed_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}
