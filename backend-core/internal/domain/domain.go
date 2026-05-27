package domain

import (
	"context"
	"time"
)

type Repository struct {
	ID            int64
	GiteaID       int64
	FullName      string
	OwnerLogin    string
	Name          string
	CloneURL      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
	UpdatedAt     time.Time
	// 소유권 분리 (migration 000042). SCM mirror 필드(위)와 구분되는 메타.
	Source      string // "scm" | "system" (빈 값 = legacy, scm 으로 취급)
	ProviderID  string // 연동된 integration_providers(scm) FK, 빈 값 가능
	Description string // 시스템 소유 — SCM sync 가 덮어쓰지 않음
}

// Repository.Source 값.
const (
	RepositorySourceSCM    = "scm"
	RepositorySourceSystem = "system"
)

type User struct {
	GiteaID     int64
	Login       string
	DisplayName string
	AvatarURL   string
	HTMLURL     string
}

type Issue struct {
	ID                int64
	GiteaID           int64
	RepositoryGiteaID int64
	RepositoryName    string
	Number            int64
	Title             string
	State             string
	AuthorLogin       string
	AssigneeLogin     string
	HTMLURL           string
	OpenedAt          *time.Time
	ClosedAt          *time.Time
	UpdatedAt         time.Time
}

type PullRequest struct {
	ID                int64
	GiteaID           int64
	RepositoryGiteaID int64
	RepositoryName    string
	Number            int64
	Title             string
	State             string
	AuthorLogin       string
	HeadBranch        string
	BaseBranch        string
	HeadSHA           string
	HTMLURL           string
	MergedAt          *time.Time
	ClosedAt          *time.Time
	UpdatedAt         time.Time
}

type CIRun struct {
	ID              int64
	ExternalID      string
	RepositoryName  string
	Branch          string
	CommitSHA       string
	Status          string
	Conclusion      string
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationSeconds *int
	HTMLURL         string
	UpdatedAt       time.Time
}

type Risk struct {
	ID               int64
	RiskKey          string
	Title            string
	Reason           string
	Impact           string
	Status           string
	OwnerLogin       string
	SourceType       string
	SourceID         string
	SuggestedActions []string
	DetectedAt       time.Time
	MitigatedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CommandStatus mirrors the API contract §2 enum exactly. Backend code that
// constructs or compares command status should use these constants rather than
// raw string literals so a future rename or addition is mechanically tracked.
type CommandStatus string

const (
	CommandStatusPending   CommandStatus = "pending"
	CommandStatusRunning   CommandStatus = "running"
	CommandStatusSucceeded CommandStatus = "succeeded"
	CommandStatusFailed    CommandStatus = "failed"
	CommandStatusRejected  CommandStatus = "rejected"
	CommandStatusCancelled CommandStatus = "cancelled"
)

// commandTerminalStates is the set from which no further transitions are
// allowed. CommandStatus.IsTerminal returns true for these.
var commandTerminalStates = map[CommandStatus]bool{
	CommandStatusSucceeded: true,
	CommandStatusFailed:    true,
	CommandStatusRejected:  true,
	CommandStatusCancelled: true,
}

// commandValidTransitions encodes the allowed lifecycle. pending may go to
// running (worker pickup) or jump straight to rejected/cancelled when the
// approval workflow blocks it. running flows into the three success/failure
// terminal states. Terminal states never transition further.
var commandValidTransitions = map[CommandStatus]map[CommandStatus]bool{
	CommandStatusPending: {
		CommandStatusRunning:   true,
		CommandStatusRejected:  true,
		CommandStatusCancelled: true,
	},
	CommandStatusRunning: {
		CommandStatusSucceeded: true,
		CommandStatusFailed:    true,
		CommandStatusCancelled: true,
	},
}

// IsTerminal reports whether the status is one from which no transition is
// allowed (succeeded, failed, rejected, cancelled).
func (s CommandStatus) IsTerminal() bool { return commandTerminalStates[s] }

// CanTransitionTo reports whether moving from s to next is allowed by the
// 6-state lifecycle. Same-state transitions return false (the caller should
// treat that as a no-op).
func (s CommandStatus) CanTransitionTo(next CommandStatus) bool {
	if s == next {
		return false
	}
	allowed, ok := commandValidTransitions[s]
	if !ok {
		return false
	}
	return allowed[next]
}

type Command struct {
	ID               int64
	CommandID        string
	CommandType      string
	TargetType       string
	TargetID         string
	ActionType       string
	Status           string
	ActorLogin       string
	Reason           string
	DryRun           bool
	RequiresApproval bool
	IdempotencyKey   string
	RequestPayload   map[string]any
	ResultPayload    map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AuditSourceType classifies which authentication path produced an audit
// row. Per DEC-2 (work_26_05_11-c, T-M1-04) the vocabulary is bounded to
// oidc | webhook | kratos | system | keycloak_event at this stage. The
// "kratos" value is deprecated (ADR-0001 superseded by ADR-0019) but
// preserved so historical audit_logs rows decode cleanly. New actor
// classes (cli, api_token, ...) extend this enum when they become real.
type AuditSourceType string

const (
	AuditSourceOIDC          AuditSourceType = "oidc"           // Bearer-verified user request
	AuditSourceWebhook       AuditSourceType = "webhook"        // signed inbound webhook (e.g. Gitea)
	AuditSourceKratos        AuditSourceType = "kratos"         // Kratos self-service hook (legacy ADR-0001 — superseded by ADR-0019)
	AuditSourceSystem        AuditSourceType = "system"         // dev fallback or background job
	AuditSourceKeycloakEvent AuditSourceType = "keycloak_event" // Keycloak Admin REST event polling (ADR-0019 §5.3 (9), sprint -v PR-C)
)

type AuditLog struct {
	ID         int64
	AuditID    string
	ActorLogin string
	Action     string
	TargetType string
	TargetID   string
	CommandID  string
	Payload    map[string]any
	// SourceIP / RequestID / SourceType are populated by the request_id
	// middleware + recordAudit (T-M1-04). Existing rows persisted before
	// migration 000008 keep these as zero-value strings.
	SourceIP   string
	RequestID  string
	SourceType AuditSourceType
	// SourceEventID — deterministic dedup key for emitters that may
	// at-least-once deliver (예: Keycloak event listener cron, ADR-0019 §5.3
	// (9) Phase 2 PR-D). 같은 (SourceType, SourceEventID) 조합으로 audit_logs 의
	// partial UNIQUE INDEX (migration 000032) 가 중복 INSERT 차단. 빈 문자열
	// 인 row 는 unique 제약을 받지 않음 (partial WHERE NOT NULL).
	SourceEventID string
	CreatedAt     time.Time
}

type RiskMitigationCommandRequest struct {
	RiskID           string
	ActorLogin       string
	ActionType       string
	Reason           string
	DryRun           bool
	IdempotencyKey   string
	RequestPayload   map[string]any
	RequiresApproval bool
	// Audit actor enrichment (PR-D follow-up, work_260512-i). Handler picks
	// these up from requireRequestID + authenticateActor + ClientIP and the
	// store records them on the commands-flow audit_logs row so the
	// "commands generated this audit" path matches the audit_logs.go
	// standalone path. Empty values land as NULL.
	SourceIP   string
	RequestID  string
	SourceType AuditSourceType
}

type ServiceActionCommandRequest struct {
	ServiceID        string
	ActorLogin       string
	ActionType       string
	Reason           string
	Force            bool
	DryRun           bool
	IdempotencyKey   string
	RequestPayload   map[string]any
	RequiresApproval bool
	// Audit actor enrichment (PR-D follow-up, work_260512-i).
	SourceIP   string
	RequestID  string
	SourceType AuditSourceType
}

type CommandApprovalRequest struct {
	CommandID  string
	ActorLogin string
	Reason     string
	// Audit actor enrichment (PR-D follow-up, work_260512-i).
	SourceIP   string
	RequestID  string
	SourceType AuditSourceType
}

type ListOptions struct {
	Limit          int
	Offset         int
	RepositoryName string
	State          string
	Status         string
	Impact         string
}

type ChangeSet struct {
	Repository  *Repository
	Sender      *User
	Issue       *Issue
	PullRequest *PullRequest
	CIRun       *CIRun
	Risk        *Risk
	Ignored     bool
	Reason      string
}

type Sink interface {
	UpsertRepository(context.Context, Repository) error
	UpsertUser(context.Context, User) error
	UpsertIssue(context.Context, Issue) error
	UpsertPullRequest(context.Context, PullRequest) error
	UpsertCIRun(context.Context, CIRun) error
	UpsertRisk(context.Context, Risk) error
	MarkWebhookEventProcessed(context.Context, int64) error
	MarkWebhookEventIgnored(context.Context, int64, string) error
	MarkWebhookEventFailed(context.Context, int64, string) error
}

type AppRole string

const (
	AppRoleDeveloper   AppRole = "developer"
	AppRoleManager     AppRole = "manager"
	AppRoleSystemAdmin AppRole = "system_admin"
	// AppRolePMOManager — ADR-0011 §4.2 / REQ-FR-PROJ-010 (sprint claude/work_260515-d).
	// Application/Project 운영 위양 role. application Edit 수정만 + project 전체 + project 멤버.
	// 시스템 설정/RBAC 정책/계정 변경 금지.
	AppRolePMOManager AppRole = "pmo_manager"
)

type UserType string

const (
	UserTypeHuman  UserType = "human"
	UserTypeSystem UserType = "system"
)

type UserStatus string

const (
	UserStatusActive      UserStatus = "active"
	UserStatusPending     UserStatus = "pending"
	UserStatusDeactivated UserStatus = "deactivated"
)

type UnitType string

const (
	UnitTypeCompany  UnitType = "company"
	UnitTypeDivision UnitType = "division"
	UnitTypeTeam     UnitType = "team"
	UnitTypeGroup    UnitType = "group"
	UnitTypePart     UnitType = "part"
)

type AppointmentRole string

const (
	AppointmentRoleLeader AppointmentRole = "leader"
	AppointmentRoleMember AppointmentRole = "member"
)

type AppUser struct {
	ID          int64
	UserID      string
	Email       string
	DisplayName string
	Role        AppRole
	Status      UserStatus
	Type        UserType
	// IdPSubject caches the IdP identity_id (Keycloak user UUID since
	// ADR-0019, formerly Kratos identity.id) so handlers can skip the
	// O(n) /admin/users scan. Empty when the row has not been backfilled
	// yet. Populated eagerly on account.create and lazily on the first
	// admin/self-service action against the user (migration 000009 added
	// the column, 000030 renamed kratos_identity_id -> idp_subject).
	IdPSubject    string
	PrimaryUnitID string
	CurrentUnitID string
	IsSeconded    bool
	JoinedAt      time.Time
	Appointments  []UnitAppointment
	// OnboardingCompletedAt — RM-ONBOARD-01 (ADR-0021 §3.3). 사용자 onboarding
	// 제출 시 set (POST /api/v1/me/onboarding). nil = 미완료 (limited mode 또는
	// admin pre-seed 직후 미완료 상태). bi-implication CHECK 제약 — ReviewStatus
	// 와 동시 NULL 또는 동시 NOT NULL.
	OnboardingCompletedAt *time.Time
	// ReviewStatus — RM-ONBOARD-01. "pending_review" (제출 직후) 또는 "reviewed"
	// (system_admin transition 후). 빈 문자열 = NULL = onboarding 미제출.
	// `pending_review` 사용자는 시스템에서 무소속 취급 (ARCH-ONBOARD-02).
	ReviewStatus string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ReviewStatus 값 enum (ADR-0021 §3.2 / ARCH-ONBOARD-02).
const (
	ReviewStatusPendingReview = "pending_review"
	ReviewStatusReviewed      = "reviewed"
)

type OrgUnit struct {
	ID           int64
	UnitID       string
	ParentUnitID string
	UnitType     UnitType
	Label        string
	LeaderUserID string
	PositionX    int
	PositionY    int
	DirectCount  int
	TotalCount   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UnitAppointment struct {
	UnitID          string
	UserID          string
	AppointmentRole AppointmentRole
}

type OrgEdge struct {
	SourceUnitID string
	TargetUnitID string
}

type Hierarchy struct {
	Units []OrgUnit
	Edges []OrgEdge
}

type UserListOptions struct {
	Limit         int
	Offset        int
	Role          string
	Status        string
	PrimaryUnitID string
}

type CreateUserInput struct {
	UserID        string
	Email         string
	DisplayName   string
	Role          AppRole
	Status        UserStatus
	Type          UserType
	PrimaryUnitID string
	CurrentUnitID string
	IsSeconded    bool
	JoinedAt      time.Time
}

type UpdateUserInput struct {
	Email         *string
	DisplayName   *string
	Role          *AppRole
	Status        *UserStatus
	PrimaryUnitID *string
	CurrentUnitID *string
	IsSeconded    *bool
	JoinedAt      *time.Time
	// ReviewStatus — RM-ONBOARD-01. PATCH /api/v1/me 의 primary_unit_id
	// 변경 시 store layer 가 자동으로 ReviewStatus=pending_review 로 reset
	// (ADR-0021 §3.2 / UC-ONBOARD-07). admin 의 review confirm transition 은
	// 별도 SubmitOnboarding/ConfirmUserReview method 사용.
	ReviewStatus *string
}

// OnboardingSubmitInput — RM-ONBOARD-01 (API-83). POST /api/v1/me/onboarding
// 제출 시 SubmitOnboarding(ctx, login, input) 호출. store 가 단일 트랜잭션으로
// (a) row INSERT 또는 UPDATE (admin pre-seeded 사용자 정합), (b) display_name +
// primary_unit_id set, (c) onboarding_completed_at = NOW(), (d) review_status =
// 'pending_review' 처리. role 은 fallback (Keycloak claim 또는 default
// `developer`) 이 caller 책임.
type OnboardingSubmitInput struct {
	UserID        string
	Email         string
	DisplayName   string
	PrimaryUnitID string
	IdPSubject    string
	// FallbackRole — token claim 매핑 결과 또는 default `developer`.
	// onboarding payload 는 role 을 받지 않음 (REQ-FR-ONBOARD-002 / §3.1).
	FallbackRole AppRole
}

type CreateOrgUnitInput struct {
	UnitID       string
	ParentUnitID string
	UnitType     UnitType
	Label        string
	LeaderUserID string
	PositionX    int
	PositionY    int
}

type UpdateOrgUnitInput struct {
	ParentUnitID *string
	UnitType     *UnitType
	Label        *string
	LeaderUserID *string
	PositionX    *int
	PositionY    *int
}
