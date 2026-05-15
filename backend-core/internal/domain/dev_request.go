package domain

import "time"

// DevRequestStatus is the 6-state lifecycle of a dev request (Dev Request 컨셉 §2.3).
type DevRequestStatus string

const (
	// DevRequestStatusReceived은 외부 수신 직후의 transient 상태. 검증 통과 시 pending,
	// 실패 시 rejected(invalid_intake) 로 즉시 전이된다.
	DevRequestStatusReceived DevRequestStatus = "received"
	// DevRequestStatusPending은 담당자 처리 대기 상태.
	DevRequestStatusPending DevRequestStatus = "pending"
	// DevRequestStatusInReview는 담당자가 dashboard 에서 열람 또는 명시적으로
	// acknowledge 한 상태. 본 sprint 는 pending→in_review 자동 전이는 도입하지 않고
	// 명시적 acknowledge 만 (carve out).
	DevRequestStatusInReview DevRequestStatus = "in_review"
	// DevRequestStatusRegistered는 application/project 로 promote 완료된 상태.
	DevRequestStatusRegistered DevRequestStatus = "registered"
	// DevRequestStatusRejected는 명시적으로 거절된 상태 (rejected_reason 필수).
	DevRequestStatusRejected DevRequestStatus = "rejected"
	// DevRequestStatusClosed는 registered/rejected 이후 system_admin 이 닫은 상태.
	DevRequestStatusClosed DevRequestStatus = "closed"
)

// DevRequestTargetType은 register(promote) 결과로 생성된 entity 의 종류.
type DevRequestTargetType string

const (
	DevRequestTargetApplication DevRequestTargetType = "application"
	DevRequestTargetProject     DevRequestTargetType = "project"
)

// DevRequest는 외부 시스템에서 들어온 1건의 개발 작업 의뢰. 컨셉 §2.1.
type DevRequest struct {
	ID                   string // UUID
	Title                string
	Details              string // markdown raw; XSS sanitize 는 frontend 책임 (REQ-NFR-DREQ-004)
	Requester            string // 외부 시스템상의 의뢰자 식별자 (DevHub user 일 필요 없음)
	AssigneeUserID       string // DevHub users.user_id FK
	SourceSystem         string // 인증된 intake token 의 매핑값 (ADR-0012 §4.1.2 spoofing 방지)
	ExternalRef          string // 외부 ticket id 등; (SourceSystem, ExternalRef) UNIQUE for idempotency
	Status               DevRequestStatus
	RegisteredTargetType DevRequestTargetType // status=registered 일 때만 채워짐
	RegisteredTargetID   string               // application_id 또는 project_id
	RejectedReason       string               // status=rejected 일 때 필수
	ReceivedAt           time.Time            // 외부 시스템에서 수신된 시각
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// DevRequestStatusTransitions은 status 머신의 허용 전이 표. 컨셉 §2.3 + REQ-FR-DREQ-003.
//
//	from / to               received pending in_review registered rejected closed
//	received                   -        ✓       -          -         ✓       -
//	pending                    -        -       ✓          ✓         ✓       -
//	in_review                  -        -       -          ✓         ✓       -
//	rejected (reopen)          -        ✓       -          -         -       ✓
//	registered                 -        -       -          -         -       ✓
//	closed                     -        -       -          -         -       -
var DevRequestStatusTransitions = map[DevRequestStatus]map[DevRequestStatus]bool{
	DevRequestStatusReceived: {
		DevRequestStatusPending:  true,
		DevRequestStatusRejected: true, // invalid_intake
	},
	DevRequestStatusPending: {
		DevRequestStatusInReview:   true,
		DevRequestStatusRegistered: true,
		DevRequestStatusRejected:   true,
	},
	DevRequestStatusInReview: {
		DevRequestStatusRegistered: true,
		DevRequestStatusRejected:   true,
	},
	DevRequestStatusRejected: {
		DevRequestStatusPending: true, // reopen
		DevRequestStatusClosed:  true,
	},
	DevRequestStatusRegistered: {
		DevRequestStatusClosed: true,
	},
	DevRequestStatusClosed: {}, // 종착 상태
}

// IsValidDevRequestTransition reports whether the (from, to) status pair is permitted.
func IsValidDevRequestTransition(from, to DevRequestStatus) bool {
	allowed, ok := DevRequestStatusTransitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

// DevRequestIntakeToken은 외부 수신 endpoint 인증을 위한 토큰 row (ADR-0012 §4.1.1).
// plain token 은 절대 저장하지 않으며, SHA-256(plain) 만 HashedToken 으로 보관.
type DevRequestIntakeToken struct {
	TokenID      string // UUID
	ClientLabel  string // 운영용 식별자 (예: "ops_portal")
	HashedToken  string // SHA-256 hex of plain token
	AllowedIPs   []string // CIDR 배열
	SourceSystem string   // intake 요청의 source_system 자동 채움 값
	CreatedAt    time.Time
	CreatedBy    string // user_id FK
	LastUsedAt   *time.Time
	RevokedAt    *time.Time
}

// IsActive reports whether the token is still usable.
func (t DevRequestIntakeToken) IsActive() bool {
	return t.RevokedAt == nil
}
