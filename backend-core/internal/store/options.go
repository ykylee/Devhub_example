package store

import (
	"errors"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

var ErrAuditInvariantViolation = errors.New("audit resource cannot grant create/edit/delete")
var ErrSystemRoleImmutable = errors.New("system role is immutable")
var ErrRoleInUse = errors.New("role is still assigned to subjects")



type ListAuditLogsOptions struct {
	Limit      int
	Offset     int
	ActorLogin string
	Action     string
	TargetType string
	TargetID   string
	CommandID  string
}

type PlatformListOptions struct {
	Status          string
	IncludeArchived bool
	Query           string
	Limit           int
	Offset          int
	ActorLogin      string
	ActorRole       string
	OrgUnitIDs      []string
	PrimaryUnitIDs  []string
}

// PlatformRepositoryLinkKey identifies a single link row (composite PK).
type PlatformRepositoryLinkKey struct {
	PlatformID  string
	RepoProvider string
	RepoFullName string
}

// ProjectListOptions parameterizes ListProjects.
type ProjectListOptions struct {
	RepositoryID    int64
	PlatformID   string
	StandaloneOnly  bool
	Status          string
	IncludeArchived bool
	Limit           int
	Offset          int
	ActorLogin      string
	ActorRole       string
	OrgUnitIDs      []string
	PrimaryUnitIDs  []string
}

type RepositoryCreatePayload struct {
	Key         string
	Slug        string
	SCMProvider string
}

// RepositoryUpdateDraftParams — draft repository 부분 갱신 입력 (sprint g #470 후속).
// nil=unchanged, ProviderID=""=unlink (SET NULL), ProviderID=uuid=set.
type RepositoryUpdateDraftParams struct {
	Key        *string
	Slug       *string
	ProviderID *string
}

type IntegrationProviderListOptions struct {
	ProviderType domain.IntegrationProviderType
	Enabled      *bool
	Limit        int
	Offset       int
}

type IntegrationBindingListOptions struct {
	ScopeType    domain.IntegrationScopeType
	ScopeID      string
	ProviderType domain.IntegrationProviderType
	Enabled      *bool
	Limit        int
	Offset       int
}

// IntegrationSyncJobListOptions — X-1 System Admin 운영 대시보드
// (RM-M4-07) 의 integration_sync_jobs 조회 옵션. 빈 Status 면 모든 status.
type IntegrationSyncJobListOptions struct {
	Status domain.IntegrationSyncJobStatus
	Limit  int
	Offset int
}

type DevRequestListOptions struct {
	Statuses       []domain.DevRequestStatus
	AssigneeUserID string
	SourceSystem   string
	Limit          int
	Offset         int
}

type ExternalTaskListOptions struct {
	ProviderID       string
	RawStatus        string
	NormalizedStatus string
	Assignee         string
	Labels           []string
	IncludeDeleted   bool
	Limit            int
	Offset           int
}

type RealtimeTicket struct {
	Ticket     string
	ActorLogin string
	ActorRole  string
	SourceType string
	ExpiresAt  time.Time
}
