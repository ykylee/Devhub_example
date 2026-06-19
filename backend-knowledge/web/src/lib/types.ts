export type ConceptType =
	| 'dataset'
	| 'metric'
	| 'api_endpoint'
	| 'runbook'
	| 'integration'
	| 'event'
	| 'reference'
	| 'decision';

export type Visibility = 'public' | 'org' | 'personal' | 'project';

export type XDevHubCurator =
	| 'rule-based'
	| 'llm-system_admin'
	| 'human-self-user'
	| 'human-org-head'
	| 'human-system-admin';

export interface PathYUserContext {
	version: 'v0';
	user_id: string;
	org_id: string;
	org_unit_ids: string[];
	project_ids: string[];
	roles: string[];
	request_id: string;
	issued_at: string;
}

export interface ConceptFrontmatter {
	type: ConceptType;
	title?: string;
	description?: string;
	resource?: string;
	tags?: string[];
	timestamp?: string;
	x_devhub_source?: string;
	x_devhub_bundle?: string;
	x_devhub_version: number;
	x_devhub_curator: XDevHubCurator;
	x_devhub_owner_org_id?: string;
	x_devhub_owner_user_id?: string;
	x_devhub_owner_org_unit_ids?: string[];
	x_devhub_owner_project_ids?: string[];
	x_devhub_visibility: Visibility;
}

export interface Concept {
	concept_id: string;
	type: string;
	name: string;
	bundle: string;
	frontmatter: ConceptFrontmatter;
	body: string;
	cross_links_out?: Array<{ target: string; type: string; section?: string }>;
	cross_links_in?: Array<{ source: string; type: string; section?: string }>;
	version: number;
	created_at?: string;
	updated_at?: string;
	updated_by: string;
	visibility: string;
}

export interface Bundle {
	name: string;
	description: string;
	owner_org_id: string;
	owner_user_id?: string;
	visibility: Visibility;
	concept_count: number;
	last_rebuild?: string;
	size_bytes: number;
}

export interface RawRecord {
	raw_id: string;
	type: ConceptType;
	name: string;
	source: string;
	sha256: string;
	size: number;
	registered_at: string;
	visibility: Visibility;
}

export type AuditEventType =
	| 'audit.user.login'
	| 'audit.concept.access'
	| 'audit.curation.edit'
	| 'audit.query'
	| 'audit.concept.archive'
	| 'audit.concept.publish'
	| 'audit.config.change';

export interface AuditEvent {
	event: AuditEventType;
	timestamp: string;
	user_id: string | null;
	org_id?: string | null;
	request_id?: string | null;
	ip?: string | null;
	success: boolean;
	[key: string]: unknown;
}

export interface EnvelopeMeta {
	request_id: string;
	timestamp: string;
	api_version: 'v0-2';
	caller_user_id: string | null;
	path_y_validated: boolean;
}

export interface Envelope<T> {
	envelope: EnvelopeMeta;
	data: T;
}

export interface ApiErrorDetail {
	code: string;
	message: string;
}

export interface ApiErrorBody {
	detail?: ApiErrorDetail;
}

export type SourceName =
	| 'homelab_mock'
	| 'gitea_repo_pull'
	| 'gitea_issue'
	| 'gitea_wiki'
	| 'gitea_action';

export type SourceState = 'idle' | 'syncing' | 'error' | 'disabled';

export type SourceHealthStatus = 'healthy' | 'degraded' | 'unhealthy';

export interface IngestStatusData {
	source: string;
	last_sync: string | null;
	next_sync: string | null;
	state: SourceState;
	last_error: { code: string; message: string } | null;
	health: SourceHealthStatus;
	metrics: Record<string, unknown>;
}

export interface IngestStatusListData {
	sources: IngestStatusData[];
	total: number;
}

export interface IngestPullData {
	pulled: number;
	failed: number;
	raw_ids: string[];
	next_pull_recommended: string | null;
	errors: Array<{ source: string; code: string; message: string }>;
}

export interface BundleDetail {
	name: string;
	description: string;
	version: number;
	concept_count: number;
	owner_org_id: string;
	owner_user_id: string | null;
	org_unit_ids: string[];
	project_ids: string[];
	visibility: string;
	created_at: string | null;
	updated_at: string | null;
	updated_by: string | null;
}

export const SOURCES: SourceName[] = [
	'homelab_mock',
	'gitea_repo_pull',
	'gitea_issue',
	'gitea_wiki',
	'gitea_action'
];

export const CONCEPT_TYPES: ConceptType[] = [
	'dataset',
	'metric',
	'api_endpoint',
	'runbook',
	'integration',
	'event',
	'reference',
	'decision'
];

export const AUDIT_EVENT_TYPES: AuditEventType[] = [
	'audit.user.login',
	'audit.concept.access',
	'audit.curation.edit',
	'audit.query',
	'audit.concept.archive',
	'audit.concept.publish',
	'audit.config.change'
];
