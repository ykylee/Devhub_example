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

type AuditLog struct {
	ID         int64
	AuditID    string
	ActorLogin string
	Action     string
	TargetType string
	TargetID   string
	CommandID  string
	Payload    map[string]any
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
	PrimaryUnitID string
	CurrentUnitID string
	IsSeconded    bool
	JoinedAt      time.Time
	Appointments  []UnitAppointment
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
