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

// ProjectStatus is the lifecycle status of a Project. Currently identical
// vocabulary to ApplicationStatus (5종 — planning/active/on_hold/closed/archived).
// 별도 alias 로 정의해 두는 이유:
//   - Project lifecycle 이 향후 Application 과 분기할 가능성 (예: Project 전용
//     `cancelled` 상태 도입) 을 코드 구조 측면에서 미리 열어 두기 위함.
//   - 분기 시점에 `type ProjectStatus string` 로 본 alias 를 끊고 별도 상수 그룹을
//     정의하면 됨. 본 sprint 까지는 vocabulary 동일하므로 alias 유지.
type ProjectStatus = ApplicationStatus

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
	SyncErrorProviderUnreachable     SyncErrorCode = "provider_unreachable"      // retryable=true
	SyncErrorAuthInvalid             SyncErrorCode = "auth_invalid"              // retryable=false
	SyncErrorPermissionDenied        SyncErrorCode = "permission_denied"         // retryable=false
	SyncErrorRateLimited             SyncErrorCode = "rate_limited"              // retryable=true
	SyncErrorWebhookSignatureInvalid SyncErrorCode = "webhook_signature_invalid" // retryable=false
	SyncErrorPayloadSchemaMismatch   SyncErrorCode = "payload_schema_mismatch"   // retryable=false
	SyncErrorResourceNotFound        SyncErrorCode = "resource_not_found"        // retryable=false
	SyncErrorInternalAdapterError    SyncErrorCode = "internal_adapter_error"    // retryable=true
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
	ID                string // UUID
	Key               string // 10-char immutable identifier (REQ-FR-APP-003)
	Name              string
	Description       string
	Status            ApplicationStatus
	Visibility        ApplicationVisibility
	OwnerUserID       string // legacy ownership field (kept for compatibility)
	LeaderUserID      string // application leader, FK users.user_id
	DevelopmentUnitID string // development department unit_id, FK org_units.unit_id
	StartDate         *time.Time
	DueDate           *time.Time
	ArchivedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
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
	ID            string // UUID
	ApplicationID string // UUID, may be empty for repo-only projects
	RepositoryID  int64  // FK repositories.id (existing BIGSERIAL)
	Key           string // unique within Repository (UNIQUE (repository_id, key))
	Name          string
	Description   string
	Status        ProjectStatus // ApplicationStatus alias — Project lifecycle 분기 시 alias 끊기
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

// IntegrationProviderType classifies provider domain.
type IntegrationProviderType string

const (
	IntegrationProviderTypeALM   IntegrationProviderType = "alm"
	IntegrationProviderTypeSCM   IntegrationProviderType = "scm"
	IntegrationProviderTypeCICD  IntegrationProviderType = "ci_cd"
	IntegrationProviderTypeDoc   IntegrationProviderType = "doc"
	IntegrationProviderTypeInfra IntegrationProviderType = "infra"
)

// IntegrationAuthMode is the credentials mode for a provider.
type IntegrationAuthMode string

const (
	IntegrationAuthModeToken       IntegrationAuthMode = "token"
	IntegrationAuthModeBasic       IntegrationAuthMode = "basic"
	IntegrationAuthModeOAuth2      IntegrationAuthMode = "oauth2"
	IntegrationAuthModeAppPassword IntegrationAuthMode = "app_password"
	IntegrationAuthModeAgent       IntegrationAuthMode = "agent"
)

// IntegrationScopeType defines binding scope.
type IntegrationScopeType string

const (
	IntegrationScopeTypeApplication IntegrationScopeType = "application"
	IntegrationScopeTypeProject     IntegrationScopeType = "project"
)

// IntegrationProvider is one row in integration_providers.
type IntegrationProvider struct {
	ID             string
	ProviderKey    string
	ProviderType   IntegrationProviderType
	DisplayName    string
	Enabled        bool
	AuthMode       IntegrationAuthMode
	CredentialsRef string
	Capabilities   []string
	SyncStatus     string
	LastSyncAt     *time.Time
	LastErrorCode  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IntegrationBinding is one row in integration_bindings.
type IntegrationBinding struct {
	ID          string
	ScopeType   IntegrationScopeType
	ScopeID     string
	ProviderID  string
	ExternalKey string
	Policy      IntegrationPolicy
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

// PRActivity is one event row from pr_activities (REQ-FR-APP-006, migration 000017).
type PRActivity struct {
	ID           int64
	RepositoryID int64
	ExternalPRID string
	EventType    string // opened|reviewed|commented|closed|merged|reopened|updated
	ActorLogin   string
	OccurredAt   time.Time
	Payload      map[string]any
	CreatedAt    time.Time
}

// BuildRun is one build execution row from build_runs (REQ-FR-APP-007).
type BuildRun struct {
	ID              int64
	RepositoryID    int64
	RunExternalID   string
	Branch          string
	CommitSHA       string
	Status          string // queued|running|success|failed|cancelled|skipped|unknown
	DurationSeconds *int
	StartedAt       time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
}

// QualitySnapshot is one quality measurement row (REQ-FR-APP-008).
type QualitySnapshot struct {
	ID            int64
	RepositoryID  int64
	Tool          string
	RefName       string
	CommitSHA     string
	Score         *float64
	GatePassed    *bool
	MetricPayload map[string]any
	MeasuredAt    time.Time
	CreatedAt     time.Time
}

// RepositoryActivity is an aggregated activity snapshot for a Repository
// (REQ-FR-APP-005). 1차 구현은 pr_activities + build_runs 의 기간 집계 — commit
// activity 의 실제 commit 이벤트는 후속 ingest pipeline 도입 시점에 추가.
type RepositoryActivity struct {
	RepositoryID       int64
	WindowFrom         time.Time
	WindowTo           time.Time
	PREventCount       int      // pr_activities 의 event 수
	ActiveContributors []string // PR 이벤트의 distinct actor_login
	BuildRunCount      int
	BuildSuccessRate   float64 // 0.0~1.0
}

// --- Application 롤업 (REQ-FR-APP-012 / REQ-NFR-PROJ-006, concept §13.4) ---

// WeightPolicy is the rollup weight policy choice (concept §13.4 + api §13.6).
type WeightPolicy string

const (
	WeightPolicyEqual    WeightPolicy = "equal"
	WeightPolicyRepoRole WeightPolicy = "repo_role"
	WeightPolicyCustom   WeightPolicy = "custom"
)

// ApplicationRollupOptions parameterizes ComputeApplicationRollup.
type ApplicationRollupOptions struct {
	Policy        WeightPolicy
	CustomWeights map[string]float64 // repo_full_name → weight (sum = 1.0 ± tolerance)
	WindowFrom    time.Time
	WindowTo      time.Time
}

// CustomWeightTolerance is the ±0.001 허용오차 for WeightPolicyCustom 합계 1.0 검증
// (concept §13.4 + api §13.6).
const CustomWeightTolerance = 0.001

// ApplicationRollupMeta는 롤업 응답의 meta 필드 (api §13.6).
type ApplicationRollupMeta struct {
	Period         RollupPeriod       `json:"period"`
	Filters        map[string]any     `json:"filters"`
	WeightPolicy   WeightPolicy       `json:"weight_policy"`
	AppliedWeights map[string]float64 `json:"applied_weights"` // repo → final weight
	Fallbacks      []RollupFallback   `json:"fallbacks"`
	DataGaps       []RollupDataGap    `json:"data_gaps"`
}

// RollupPeriod is the time window of the rollup.
type RollupPeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// RollupFallback notes when a repo weight is filled in by a fallback policy.
type RollupFallback struct {
	RepoFullName  string  `json:"repo_full_name"`
	Provider      string  `json:"provider"`
	Reason        string  `json:"reason"` // e.g., "custom_weight_missing"
	AppliedWeight float64 `json:"applied_weight"`
}

// RollupDataGap notes when a repo is excluded from the rollup due to missing or
// unreachable data (concept §13.4 누락 데이터 처리).
type RollupDataGap struct {
	RepoFullName string `json:"repo_full_name"`
	Provider     string `json:"provider"`
	Reason       string `json:"reason"` // e.g., "provider_unreachable" | "no_data_in_window"
}

// ApplicationRollup is the aggregated rollup payload (api §13.6 응답 data).
type ApplicationRollup struct {
	PullRequestDistribution map[string]int        `json:"pull_request_distribution"`  // opened/merged/closed/...
	BuildSuccessRate        float64               `json:"build_success_rate"`         // weighted average 0.0~1.0
	BuildAvgDurationSeconds int                   `json:"build_avg_duration_seconds"` // weighted average
	QualityScore            float64               `json:"quality_score"`              // weighted average
	QualityGateFailedCount  int                   `json:"quality_gate_failed_count"`
	CriticalWarningCount    int                   `json:"critical_warning_count"` // active→closed 가드 의존
	Meta                    ApplicationRollupMeta `json:"-"`                      // 별도 meta 필드로 serialize
}
