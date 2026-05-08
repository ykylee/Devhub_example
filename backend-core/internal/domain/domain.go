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
}

type CommandApprovalRequest struct {
	CommandID  string
	ActorLogin string
	Reason     string
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

type RBACPermission string

const (
	RBACPermissionNone  RBACPermission = "none"
	RBACPermissionRead  RBACPermission = "read"
	RBACPermissionWrite RBACPermission = "write"
	RBACPermissionAdmin RBACPermission = "admin"
)

type RBACRole struct {
	Role        AppRole
	Label       string
	Description string
}

type RBACResource struct {
	Resource    string
	Label       string
	Description string
}

type RBACPermissionLevel struct {
	Permission  RBACPermission
	Label       string
	Rank        int
	Description string
}

type RBACPolicy struct {
	PolicyVersion string
	Source        string
	Editable      bool
	Roles         []RBACRole
	Resources     []RBACResource
	Permissions   []RBACPermissionLevel
	Matrix        map[string]map[string]RBACPermission
}

type ReplaceRBACPolicyInput struct {
	PolicyVersion string
	ActorLogin    string
	Reason        string
	Policy        RBACPolicy
}

func DefaultRBACPolicy() RBACPolicy {
	return RBACPolicy{
		PolicyVersion: "2026-05-07.default",
		Source:        "static_default_policy",
		Editable:      false,
		Roles: []RBACRole{
			{
				Role:        AppRoleDeveloper,
				Label:       "Developer",
				Description: "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한",
			},
			{
				Role:        AppRoleManager,
				Label:       "Manager",
				Description: "팀 운영, risk triage, 승인 전 command 생성 권한",
			},
			{
				Role:        AppRoleSystemAdmin,
				Label:       "System Admin",
				Description: "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한",
			},
		},
		Resources: []RBACResource{
			{Resource: "repositories", Label: "Repositories", Description: "repository, issue, pull request metadata"},
			{Resource: "ci_runs", Label: "CI Runs", Description: "CI run, build log, deployment state"},
			{Resource: "risks", Label: "Risks", Description: "risk list, mitigation command, decision audit"},
			{Resource: "commands", Label: "Commands", Description: "service action and mitigation command lifecycle"},
			{Resource: "organization", Label: "Organization", Description: "users, org units, memberships"},
			{Resource: "system_config", Label: "System Config", Description: "runtime config, adapters, infrastructure controls"},
		},
		Permissions: []RBACPermissionLevel{
			{Permission: RBACPermissionNone, Label: "None", Rank: 0, Description: "접근 불가"},
			{Permission: RBACPermissionRead, Label: "Read", Rank: 10, Description: "조회 가능"},
			{Permission: RBACPermissionWrite, Label: "Write", Rank: 20, Description: "생성/수정 또는 command 생성 가능"},
			{Permission: RBACPermissionAdmin, Label: "Admin", Rank: 30, Description: "관리 작업과 위험 작업 승인 가능"},
		},
		Matrix: map[string]map[string]RBACPermission{
			string(AppRoleDeveloper): {
				"repositories":  RBACPermissionRead,
				"ci_runs":       RBACPermissionRead,
				"risks":         RBACPermissionRead,
				"commands":      RBACPermissionNone,
				"organization":  RBACPermissionNone,
				"system_config": RBACPermissionNone,
			},
			string(AppRoleManager): {
				"repositories":  RBACPermissionWrite,
				"ci_runs":       RBACPermissionRead,
				"risks":         RBACPermissionWrite,
				"commands":      RBACPermissionWrite,
				"organization":  RBACPermissionRead,
				"system_config": RBACPermissionNone,
			},
			string(AppRoleSystemAdmin): {
				"repositories":  RBACPermissionAdmin,
				"ci_runs":       RBACPermissionAdmin,
				"risks":         RBACPermissionAdmin,
				"commands":      RBACPermissionAdmin,
				"organization":  RBACPermissionAdmin,
				"system_config": RBACPermissionAdmin,
			},
		},
	}
}

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
