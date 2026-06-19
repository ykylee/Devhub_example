import { encodePathY, loadPathY } from './path-y';
import type {
	ApiErrorBody,
	AuditEvent,
	Bundle,
	Concept,
	Envelope,
	PathYUserContext,
	RawRecord
} from './types';

export class ApiError extends Error {
	constructor(
		public status: number,
		public code: string,
		message: string
	) {
		super(`${status} ${code}: ${message}`);
		this.name = 'ApiError';
	}
}

export interface ApiOptions {
	baseUrl?: string;
	pathY?: PathYUserContext | null;
}

const DEFAULT_BASE_URL = 'http://localhost:8000';

function buildHeaders(options: ApiOptions, init?: HeadersInit): Headers {
	const headers = new Headers(init);
	headers.set('Accept', 'application/json');
	if (options.pathY) {
		headers.set('X-DevHub-User-Context', encodePathY(options.pathY));
	}
	return headers;
}

async function request<T>(
	path: string,
	method: string,
	options: ApiOptions,
	body?: unknown
): Promise<Envelope<T>> {
	const baseUrl = options.baseUrl ?? DEFAULT_BASE_URL;
	const headers = buildHeaders(options, body ? { 'Content-Type': 'application/json' } : undefined);
	const init: RequestInit = { method, headers };
	if (body !== undefined) {
		init.body = JSON.stringify(body);
	}
	const res = await fetch(`${baseUrl}${path}`, init);
	if (!res.ok) {
		const errBody = (await res.json().catch(() => ({}))) as ApiErrorBody;
		throw new ApiError(
			res.status,
			errBody.detail?.code ?? 'E_UNKNOWN',
			errBody.detail?.message ?? res.statusText
		);
	}
	return (await res.json()) as Envelope<T>;
}

export const api = {
	health: (opts: ApiOptions = {}) =>
		request<{ status: string; version: string }>('/health', 'GET', opts),

	getMe: (opts: ApiOptions = {}) =>
		request<{
			user_id: string;
			org_id: string;
			roles: string[];
			project_ids: string[];
		}>('/api/v0-2/health/protected', 'GET', opts),

	listSources: (opts: ApiOptions = {}) =>
		request<Array<{ name: string; health: string }>>(
			'/api/v0-2/ingest/homelab_mock/status',
			'GET',
			opts
		),

	syncSource: (source: string, dryRun = false, opts: ApiOptions = {}) =>
		request<{
			synced: number;
			failed: number;
			raw_ids: string[];
			errors: Array<{ raw_name: string; code: string; message: string }>;
		}>(`/api/v0-2/ingest/${source}/sync?dry_run=${dryRun}`, 'POST', opts),

	listConcepts: (params: { q?: string; bundle?: string; type?: string; limit?: number } = {}, opts: ApiOptions = {}) => {
		const search = new URLSearchParams();
		if (params.q) search.set('q', params.q);
		if (params.bundle) search.set('bundle', params.bundle);
		if (params.type) search.set('type', params.type);
		if (params.limit) search.set('limit', String(params.limit));
		return request<{ hits: Array<{
			concept_id: string;
			type: string;
			title: string;
			snippet: string;
			score: number;
			bundle: string;
			source: string;
		}>; total: number; next_offset: number | null }>(
			`/api/v0-2/search?${search.toString()}`,
			'GET',
			opts
		);
	},

	getConcept: (bundle: string, type: string, name: string, opts: ApiOptions = {}) =>
		request<Concept>(
			`/api/v0-2/concepts/${encodeURIComponent(type)}/${encodeURIComponent(name)}?bundle=${encodeURIComponent(bundle)}`,
			'GET',
			opts
		),

	listBundles: (opts: ApiOptions = {}) =>
		request<{ items: Bundle[]; total: number }>('/api/v0-2/bundles', 'GET', opts),

	createBundle: (
		body: { name: string; description?: string; owner_org_id: string; visibility: 'public' | 'org' | 'personal' | 'project' },
		opts: ApiOptions = {}
	) => request<{ name: string; created_at: string; created_by: string; visibility: string; path: string }>(
		'/api/v0-2/bundles',
		'POST',
		opts,
		body
	),

	rebuildBundle: (bundle: string, dryRun = false, opts: ApiOptions = {}) =>
		request<{
			bundle: string;
			concept_count: number;
			link_count: number;
			reverse_index_generated: boolean;
			index_md_generated: boolean;
			viz_html_generated: boolean;
			duration_ms: number;
			rebuilt_at: string;
		}>(`/api/v0-2/bundles/${bundle}/rebuild?dry_run=${dryRun}`, 'POST', opts),

	listRaw: (params: { source?: string; limit?: number } = {}, opts: ApiOptions = {}) => {
		const search = new URLSearchParams();
		if (params.source) search.set('source', params.source);
		if (params.limit) search.set('limit', String(params.limit));
		return request<{ items: RawRecord[]; total: number }>(
			`/api/v0-2/raw?${search.toString()}`,
			'GET',
			opts
		);
	},

	listAudit: (
		params: { event_type?: string; user_id?: string; from?: string; to?: string; limit?: number } = {},
		opts: ApiOptions = {}
	) => {
		const search = new URLSearchParams();
		if (params.event_type) search.set('event_type', params.event_type);
		if (params.user_id) search.set('user_id', params.user_id);
		if (params.from) search.set('from', params.from);
		if (params.to) search.set('to', params.to);
		if (params.limit) search.set('limit', String(params.limit));
		return request<{ items: AuditEvent[]; total: number; filters: Record<string, unknown> }>(
			`/api/v0-2/audit?${search.toString()}`,
			'GET',
			opts
		);
	},

	listAlerts: (opts: ApiOptions = {}) =>
		request<{
			alerts: Array<{
				metric: string;
				value: number;
				severity: 'info' | 'warning' | 'critical';
				threshold: [number | null, number | null] | null;
				message: string;
				evaluated_at: string;
			}>;
			total: number;
			by_severity: { info: number; warning: number; critical: number };
		}>('/api/v0-2/monitoring/alerts', 'GET', opts)
};

export function withStoredPathY(opts: ApiOptions = {}): ApiOptions {
	return { ...opts, pathY: opts.pathY ?? loadPathY() };
}
