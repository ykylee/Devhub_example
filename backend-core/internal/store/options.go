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

type ApplicationListOptions struct {
	Status          string
	IncludeArchived bool
	Query           string
	Limit           int
	Offset          int
}

// ApplicationRepositoryLinkKey identifies a single link row (composite PK).
type ApplicationRepositoryLinkKey struct {
	ApplicationID string
	RepoProvider  string
	RepoFullName  string
}

// ProjectListOptions parameterizes ListProjects.
type ProjectListOptions struct {
	RepositoryID    int64
	ApplicationID   string
	StandaloneOnly  bool
	Status          string
	IncludeArchived bool
	Limit           int
	Offset          int
}

type RepositoryCreatePayload struct {
	Key         string
	Slug        string
	SCMProvider string
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
