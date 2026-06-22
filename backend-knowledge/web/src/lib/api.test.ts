// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, api, withStoredPathY } from './api';
import { encodePathY, makeDevPathY, savePathY } from './path-y';
import type { PathYUserContext } from './types';

function mockFetchOnce(response: { status: number; body: unknown }): ReturnType<typeof vi.fn> {
	const spy = vi.fn().mockResolvedValueOnce({
		ok: response.status >= 200 && response.status < 300,
		status: response.status,
		statusText: response.status === 200 ? 'OK' : 'ERR',
		json: async () => response.body
	});
	globalThis.fetch = spy as unknown as typeof fetch;
	return spy;
}

describe('ApiError', () => {
	it('constructs with status, code, message', () => {
		const e = new ApiError(404, 'E_NOT_FOUND', 'not found');
		expect(e.status).toBe(404);
		expect(e.code).toBe('E_NOT_FOUND');
		expect(e.message).toBe('not found');
		expect(e.name).toBe('ApiError');
	});

	it('inherits from Error', () => {
		const e = new ApiError(500, 'E_INTERNAL', 'oops');
		expect(e).toBeInstanceOf(Error);
		expect(e.toString()).toContain('500 E_INTERNAL');
	});
});

describe('withStoredPathY', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns opts unchanged when pathY is provided', () => {
		const ctx = makeDevPathY();
		const opts = { pathY: ctx };
		const result = withStoredPathY(opts);
		expect(result.pathY).toBe(ctx);
	});

	it('loads from localStorage when pathY is null', () => {
		const ctx = makeDevPathY();
		savePathY(ctx);
		const result = withStoredPathY();
		expect(result.pathY).toEqual(ctx);
	});

	it('returns null pathY when nothing stored', () => {
		const result = withStoredPathY();
		expect(result.pathY).toBeNull();
	});

	it('preserves other opts fields', () => {
		const result = withStoredPathY({ baseUrl: 'http://example.com' });
		expect(result.baseUrl).toBe('http://example.com');
	});
});

describe('api.health', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns health status', async () => {
		mockFetchOnce({
			status: 200,
			body: { data: { status: 'ok', version: '0.2.0' }, request_id: 'req_1' }
		});
		const result = await api.health();
		expect(result.data.status).toBe('ok');
		expect(result.data.version).toBe('0.2.0');
	});

	it('throws ApiError on 500', async () => {
		mockFetchOnce({
			status: 500,
			body: { detail: { code: 'E_INTERNAL', message: 'oops' } }
		});
		await expect(api.health()).rejects.toThrow(ApiError);
	});

	it('uses X-DevHub-User-Context header when pathY provided', async () => {
		const ctx = makeDevPathY();
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { status: 'ok', version: '0.2.0' }, request_id: 'req' }
		});
		await api.health({ pathY: ctx });
		const callArgs = fetchSpy.mock.calls[0] as [string, RequestInit];
		const headers = callArgs[1].headers as Headers;
		expect(headers.get('X-DevHub-User-Context')).toBe(encodePathY(ctx));
	});

	it('does not include X-DevHub-User-Context when pathY is null', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { status: 'ok', version: '0.2.0' }, request_id: 'req' }
		});
		await api.health();
		const callArgs = fetchSpy.mock.calls[0] as [string, RequestInit];
		const headers = callArgs[1].headers as Headers;
		expect(headers.get('X-DevHub-User-Context')).toBeNull();
	});
});

describe('api.syncSource', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('sends POST to /api/v0-2/ingest/{source}/sync', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { synced: 3, failed: 0, raw_ids: [], errors: [] }, request_id: 'req' }
		});
		await api.syncSource('homelab_mock');
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('/api/v0-2/ingest/homelab_mock/sync');
		expect(url).toContain('dry_run=false');
	});

	it('encodes dryRun=true as query param', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { synced: 0, failed: 0, raw_ids: [], errors: [] }, request_id: 'req' }
		});
		await api.syncSource('gitea_issue', {}, true);
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('dry_run=true');
	});
});

describe('api.listConcepts', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('builds query string from params', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { hits: [], total: 0, next_offset: null }, request_id: 'req' }
		});
		await api.listConcepts({ q: 'test', bundle: 'devhub', type: 'dataset', limit: 10 });
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('q=test');
		expect(url).toContain('bundle=devhub');
		expect(url).toContain('type=dataset');
		expect(url).toContain('limit=10');
	});

	it('handles empty params', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { hits: [], total: 0, next_offset: null }, request_id: 'req' }
		});
		await api.listConcepts();
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('/api/v0-2/search?');
	});
});

describe('api.getConcept', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('encodes type and name in URL', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { concept_id: 'devhub/dataset/foo' }, request_id: 'req' }
		});
		await api.getConcept('devhub', 'dataset', 'foo');
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('dataset');
		expect(url).toContain('foo');
		expect(url).toContain('bundle=devhub');
	});

	it('handles special characters in name', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { concept_id: 'x' }, request_id: 'req' }
		});
		await api.getConcept('devhub', 'dataset', '한글');
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(decodeURIComponent(url)).toContain('한글');
	});
});

describe('api.createBundle', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('sends POST with JSON body', async () => {
		const fetchSpy = mockFetchOnce({
			status: 201,
			body: { data: { name: 'new-bundle' }, request_id: 'req' }
		});
		await api.createBundle({ name: 'new-bundle', owner_org_id: 'ou_root', visibility: 'org' });
		const callArgs = fetchSpy.mock.calls[0] as [string, RequestInit];
		const headers = callArgs[1].headers as Headers;
		expect(headers.get('Content-Type')).toBe('application/json');
		const body = JSON.parse(callArgs[1].body as string) as Record<string, unknown>;
		expect(body['name']).toBe('new-bundle');
		expect(body['owner_org_id']).toBe('ou_root');
	});
});

describe('api.deleteRaw', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('sends DELETE to encoded raw_id', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { deleted: true, raw_id: 'abc1234' }, request_id: 'req' }
		});
		await api.deleteRaw('abc/1234+foo');
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('/api/v0-2/raw/');
		expect(url).not.toContain('abc/1234+foo');
		expect(decodeURIComponent(url)).toContain('abc/1234+foo');
	});
});

describe('api error handling', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('extracts code from error body detail', async () => {
		mockFetchOnce({
			status: 404,
			body: { detail: { code: 'E_NOT_FOUND', message: 'not found' } }
		});
		try {
			await api.health();
			expect.fail('should have thrown');
		} catch (e) {
			expect(e).toBeInstanceOf(ApiError);
			const err = e as ApiError;
			expect(err.status).toBe(404);
			expect(err.code).toBe('E_NOT_FOUND');
			expect(err.message).toBe('not found');
		}
	});

	it('falls back to E_UNKNOWN when body has no detail', async () => {
		mockFetchOnce({
			status: 500,
			body: {}
		});
		try {
			await api.health();
			expect.fail('should have thrown');
		} catch (e) {
			const err = e as ApiError;
			expect(err.code).toBe('E_UNKNOWN');
		}
	});

	it('handles non-JSON error body', async () => {
		const spy = vi.fn().mockResolvedValueOnce({
			ok: false,
			status: 503,
			statusText: 'Service Unavailable',
			json: async () => {
				throw new Error('parse error');
			}
		});
		globalThis.fetch = spy as unknown as typeof fetch;
		try {
			await api.health();
			expect.fail('should have thrown');
		} catch (e) {
			const err = e as ApiError;
			expect(err.status).toBe(503);
			expect(err.code).toBe('E_UNKNOWN');
		}
	});
});

describe('api.listAudit query params', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('builds query string from all params', async () => {
		const fetchSpy = mockFetchOnce({
			status: 200,
			body: { data: { items: [], total: 0, filters: {} }, request_id: 'req' }
		});
		await api.listAudit({
			event_type: 'audit.curation.edit',
			user_id: 'u_001',
			from: '2026-06-01',
			to: '2026-06-30',
			limit: 50
		});
		const url = fetchSpy.mock.calls[0]?.[0] as string;
		expect(url).toContain('event_type=audit.curation.edit');
		expect(url).toContain('user_id=u_001');
		expect(url).toContain('from=2026-06-01');
		expect(url).toContain('to=2026-06-30');
		expect(url).toContain('limit=50');
	});
});