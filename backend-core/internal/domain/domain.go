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
}

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
// oidc | webhook | kratos | system at this stage. New actor classes (cli,
// api_token, ...) extend this enum when they become real.
type AuditSourceType string

const (
	AuditSourceOIDC    AuditSourceType = "oidc"    // Bearer-verified user request
	AuditSourceWebhook AuditSourceType = "webhook" // signed inbound webhook (e.g. Gitea)
	AuditSourceKratos  AuditSourceType = "kratos"  // Kratos self-service hook (e.g. settings/password/after)
	AuditSourceSystem  AuditSourceType = "system"  // dev fallback or background job
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
	CreatedAt  time.Time
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
	ID            int64
	UserID        string
	Email         string
	DisplayName   string
	Role          AppRole
	Status        UserStatus
	Type          UserType
	// KratosIdentityID caches the Kratos identity_id so handlers can skip
	// the O(n) /admin/identities scan. Empty when the row has not been
	// backfilled yet. Populated eagerly on account.create and lazily on
	// the first admin/self-service action against the user (migration
	// 000009).
	KratosIdentityID string
	PrimaryUnitID    string
	CurrentUnitID    string
	IsSeconded       bool
	JoinedAt         time.Time
	Appointments     []UnitAppointment
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

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
