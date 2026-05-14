package domain

import "time"

// ApplicationStatus is the lifecycle status of an Application (concept §13.2).
type ApplicationStatus string

const (
	ApplicationStatusPlanning ApplicationStatus = "planning"
	ApplicationStatusActive   ApplicationStatus = "active"
	ApplicationStatusOnHold   ApplicationStatus = "on_hold"
	ApplicationStatusClosed   ApplicationStatus = "closed"
	ApplicationStatusArchived ApplicationStatus = "archived"
)

// ApplicationVisibility is the visibility classification (concept §5.1, api §13.1).
type ApplicationVisibility string

const (
	ApplicationVisibilityPublic     ApplicationVisibility = "public"
	ApplicationVisibilityInternal   ApplicationVisibility = "internal"
	ApplicationVisibilityRestricted ApplicationVisibility = "restricted"
)

// ApplicationRepositoryRole categorizes a linked Repository (concept §13.4 — rollup weight repo_role).
type ApplicationRepositoryRole string

const (
	ApplicationRepositoryRolePrimary ApplicationRepositoryRole = "primary"
	ApplicationRepositoryRoleSub     ApplicationRepositoryRole = "sub"
	ApplicationRepositoryRoleShared  ApplicationRepositoryRole = "shared"
)

// ApplicationRepositorySyncStatus represents the link-level sync lifecycle (concept §13.3).
type ApplicationRepositorySyncStatus string

const (
	SyncStatusRequested    ApplicationRepositorySyncStatus = "requested"
	SyncStatusVerifying    ApplicationRepositorySyncStatus = "verifying"
	SyncStatusActive       ApplicationRepositorySyncStatus = "active"
	SyncStatusDegraded     ApplicationRepositorySyncStatus = "degraded"
	SyncStatusDisconnected ApplicationRepositorySyncStatus = "disconnected"
)

// SyncErrorCode is the standardized link-level error code dictionary (api §13.3).
// Stored as link-scope latest-1 cache; per-event details belong in webhook_events / adapter_event_logs.
type SyncErrorCode string

const (
	SyncErrorProviderUnreachable    SyncErrorCode = "provider_unreachable"     // retryable=true
	SyncErrorAuthInvalid            SyncErrorCode = "auth_invalid"             // retryable=false
	SyncErrorPermissionDenied       SyncErrorCode = "permission_denied"        // retryable=false
	SyncErrorRateLimited            SyncErrorCode = "rate_limited"             // retryable=true
	SyncErrorWebhookSignatureInvalid SyncErrorCode = "webhook_signature_invalid" // retryable=false
	SyncErrorPayloadSchemaMismatch  SyncErrorCode = "payload_schema_mismatch"  // retryable=false
	SyncErrorResourceNotFound       SyncErrorCode = "resource_not_found"       // retryable=false
	SyncErrorInternalAdapterError   SyncErrorCode = "internal_adapter_error"   // retryable=true
)

// IsRetryableSyncError reports whether the given code is operationally retryable
// per the api §13.3 dictionary. Unknown codes default to non-retryable.
func IsRetryableSyncError(code SyncErrorCode) bool {
	switch code {
	case SyncErrorProviderUnreachable,
		SyncErrorRateLimited,
		SyncErrorInternalAdapterError:
		return true
	}
	return false
}

// ProjectMemberRole categorizes membership in a Project.
type ProjectMemberRole string

const (
	ProjectMemberRoleLead        ProjectMemberRole = "lead"
	ProjectMemberRoleContributor ProjectMemberRole = "contributor"
	ProjectMemberRoleObserver    ProjectMemberRole = "observer"
)

// IntegrationScope distinguishes Application-level vs Project-level integrations.
type IntegrationScope string

const (
	IntegrationScopeApplication IntegrationScope = "application"
	IntegrationScopeProject     IntegrationScope = "project"
)

// IntegrationType classifies the external system an integration connects to.
type IntegrationType string

const (
	IntegrationTypeJira       IntegrationType = "jira"
	IntegrationTypeConfluence IntegrationType = "confluence"
)

// IntegrationPolicy is the operational policy applied to an integration (concept §3.1 / api §13.7).
type IntegrationPolicy string

const (
	IntegrationPolicySummaryOnly     IntegrationPolicy = "summary_only"
	IntegrationPolicyExecutionSystem IntegrationPolicy = "execution_system"
)

// Application is the top-level governance entity for a product/service lifecycle.
type Application struct {
	ID          string // UUID
	Key         string // 10-char immutable identifier (REQ-FR-APP-003)
	Name        string
	Description string
	Status      ApplicationStatus
	Visibility  ApplicationVisibility
	OwnerUserID string // FK users.user_id, may be empty if unset
	StartDate   *time.Time
	DueDate     *time.Time
	ArchivedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ApplicationRepository is one link between an Application and an external Repository
// (composite PK = (ApplicationID, RepoProvider, RepoFullName) per concept §13.3).
type ApplicationRepository struct {
	ApplicationID      string // UUID
	RepoProvider       string
	RepoFullName       string
	ExternalRepoID     string // optional
	Role               ApplicationRepositoryRole
	SyncStatus         ApplicationRepositorySyncStatus
	SyncErrorCode      SyncErrorCode // empty if no error
	SyncErrorRetryable *bool
	SyncErrorAt        *time.Time
	LastSyncAt         *time.Time
	LinkedAt           time.Time
}

// Project is a time-bounded operational unit hosted under a Repository.
type Project struct {
	ID            string  // UUID
	ApplicationID string  // UUID, may be empty for repo-only projects
	RepositoryID  int64   // FK repositories.id (existing BIGSERIAL)
	Key           string  // unique within Repository (UNIQUE (repository_id, key))
	Name          string
	Description   string
	Status        ApplicationStatus // 동일 vocabulary 재사용
	Visibility    ApplicationVisibility
	OwnerUserID   string
	StartDate     *time.Time
	DueDate       *time.Time
	ArchivedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProjectMember is one membership row.
type ProjectMember struct {
	ProjectID   string // UUID
	UserID      string // FK users.user_id
	ProjectRole ProjectMemberRole
	JoinedAt    time.Time
}

// ProjectIntegration represents one external integration (Jira/Confluence) bound to
// either an Application or a Project (single-table polymorphism via Scope + nullable FKs).
type ProjectIntegration struct {
	ID              string // UUID
	Scope           IntegrationScope
	ProjectID       string // empty if Scope=application
	ApplicationID   string // empty if Scope=project
	IntegrationType IntegrationType
	ExternalKey     string
	URL             string
	Policy          IntegrationPolicy
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SCMProvider is one entry of the SCM adapter catalog (concept §12.2 + api §13.1.1).
type SCMProvider struct {
	ProviderKey    string
	DisplayName    string
	Enabled        bool
	AdapterVersion string // semver, set by deployment pipeline only
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
