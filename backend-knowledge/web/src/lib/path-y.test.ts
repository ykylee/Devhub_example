// @vitest-environment jsdom
import { describe, expect, it, beforeEach } from 'vitest';
import {
	clearPathY,
	encodePathY,
	loadPathY,
	makeDevPathY,
	savePathY
} from './path-y';
import type { PathYUserContext } from './types';

describe('encodePathY', () => {
	it('encodes a simple context as base64url', () => {
		const ctx: PathYUserContext = {
			version: 'v0',
			user_id: 'u_001',
			org_id: 'ou_root',
			org_unit_ids: ['ou_root'],
			project_ids: [],
			roles: ['developer'],
			request_id: 'req_001',
			issued_at: '2026-06-22T00:00:00Z'
		};
		const encoded = encodePathY(ctx);
		expect(encoded).not.toContain('+');
		expect(encoded).not.toContain('/');
		expect(encoded).not.toContain('=');
		expect(encoded.length).toBeGreaterThan(0);
	});

	it('handles Korean characters', () => {
		const ctx: PathYUserContext = {
			version: 'v0',
			user_id: 'u_한국',
			org_id: 'ou_한국',
			org_unit_ids: [],
			project_ids: [],
			roles: [],
			request_id: 'req',
			issued_at: '2026-06-22T00:00:00Z'
		};
		const encoded = encodePathY(ctx);
		expect(encoded).toMatch(/^[A-Za-z0-9_-]+$/);
	});

	it('roundtrips through btoa/JSON', () => {
		const ctx = makeDevPathY();
		const encoded = encodePathY(ctx);
		const decoded = JSON.parse(atob(encoded));
		expect(decoded.user_id).toBe(ctx.user_id);
		expect(decoded.org_id).toBe(ctx.org_id);
	});
});

describe('makeDevPathY', () => {
	it('creates a v0 context with required fields', () => {
		const ctx = makeDevPathY();
		expect(ctx.version).toBe('v0');
		expect(ctx.user_id).toBe('u_dev_001');
		expect(ctx.org_id).toBe('ou_dev_dept_a');
		expect(ctx.org_unit_ids).toContain('ou_dev_dept_a');
		expect(ctx.project_ids).toContain('prj_dev');
		expect(ctx.roles).toContain('developer');
	});

	it('generates unique request_id per call', async () => {
		const ctx1 = makeDevPathY();
		await new Promise((resolve) => setTimeout(resolve, 5));
		const ctx2 = makeDevPathY();
		expect(ctx1.request_id).not.toBe(ctx2.request_id);
	});

	it('uses ISO 8601 issued_at', () => {
		const ctx = makeDevPathY();
		expect(ctx.issued_at).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
	});
});

describe('localStorage integration (loadPathY/savePathY/clearPathY)', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns null when no context is stored', () => {
		expect(loadPathY()).toBeNull();
	});

	it('returns null when window is undefined (SSR safety)', () => {
		const originalWindow = globalThis.window;
		// @ts-expect-error - simulate SSR
		delete (globalThis as { window?: unknown }).window;
		expect(loadPathY()).toBeNull();
		globalThis.window = originalWindow;
	});

	it('saves and loads a context roundtrip', () => {
		const ctx = makeDevPathY();
		savePathY(ctx);
		const loaded = loadPathY();
		expect(loaded).toEqual(ctx);
	});

	it('clearPathY removes the stored context', () => {
		const ctx = makeDevPathY();
		savePathY(ctx);
		expect(loadPathY()).not.toBeNull();
		clearPathY();
		expect(loadPathY()).toBeNull();
	});

	it('returns null on corrupted JSON', () => {
		localStorage.setItem('bk_path_y_context', 'not-json{');
		expect(loadPathY()).toBeNull();
	});

	it('savePathY is no-op when window is undefined', () => {
		const originalWindow = globalThis.window;
		// @ts-expect-error - simulate SSR
		delete (globalThis as { window?: unknown }).window;
		const ctx = makeDevPathY();
		expect(() => savePathY(ctx)).not.toThrow();
		globalThis.window = originalWindow;
	});
});

describe('encodePathY cross-environment', () => {
	it('works in jsdom (btoa available)', () => {
		const ctx = makeDevPathY();
		const encoded = encodePathY(ctx);
		expect(typeof encoded).toBe('string');
		expect(encoded.length).toBeGreaterThan(0);
	});
});