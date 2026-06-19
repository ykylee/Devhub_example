import type { PathYUserContext } from './types';

const STORAGE_KEY = 'bk_path_y_context';

export function loadPathY(): PathYUserContext | null {
	if (typeof window === 'undefined') return null;
	const raw = localStorage.getItem(STORAGE_KEY);
	if (!raw) return null;
	try {
		return JSON.parse(raw) as PathYUserContext;
	} catch {
		return null;
	}
}

export function savePathY(ctx: PathYUserContext): void {
	if (typeof window === 'undefined') return;
	localStorage.setItem(STORAGE_KEY, JSON.stringify(ctx));
}

export function clearPathY(): void {
	if (typeof window === 'undefined') return;
	localStorage.removeItem(STORAGE_KEY);
}

export function encodePathY(ctx: PathYUserContext): string {
	const json = JSON.stringify(ctx);
	if (typeof btoa === 'function') {
		return btoa(json).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
	}
	return Buffer.from(json, 'utf-8').toString('base64url');
}

export function makeDevPathY(): PathYUserContext {
	const now = new Date();
	return {
		version: 'v0',
		user_id: 'u_dev_001',
		org_id: 'ou_dev_dept_a',
		org_unit_ids: ['ou_dev_dept_a', 'ou_dev_dept_b1'],
		project_ids: ['prj_dev'],
		roles: ['developer', 'project_leader:prj_dev'],
		request_id: `req_dev_${now.getTime()}`,
		issued_at: now.toISOString()
	};
}
