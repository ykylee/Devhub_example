<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import type { AuditEvent } from '$lib/types';

	let entries = $state<AuditEvent[] | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	// M-v0.2.1+ scope: cursor → offset 정합. backend 의 listAudit 가 `offset` query parameter 추가 (PR #662 별도).
	let cursor = $state<number | null>(null);
	let cursorHistory = $state<number[]>([]);

	async function load(append: boolean = false): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const params: { limit: number; offset?: number } = { limit: 50 };
			if (append && cursor !== null) params.offset = cursor;
			const r = await api.listAudit(params, opts);
			// M-v0.2.1+ scope: backend audit api 에 `next_offset` query parameter + response field 추가 후 정합.
			cursor = null;
			entries = append ? [...(entries ?? []), ...r.data.items] : r.data.items;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	function loadMore(): void {
		if (cursor !== null) {
			cursorHistory = [...cursorHistory, cursor];
			load(true);
		}
	}

	function loadPrev(): void {
		if (cursorHistory.length > 0) {
			cursor = cursorHistory[cursorHistory.length - 1] ?? null;
			cursorHistory = cursorHistory.slice(0, -1);
			load(true);
		} else {
			load(false);
		}
	}

	onMount(() => load(false));
</script>

<header class="page-header">
	<h1>Audit log</h1>
	<p>
		Append-only audit trail (umbrella doc §3.8). user_id / org_id / action / resource / created_at
		표시.
	</p>
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

<div class="toolbar">
	<button onclick={loadPrev} disabled={loading}>Refresh</button>
	<button
		class="secondary"
		onclick={loadMore}
		disabled={loading || !cursor}
	>
		{loading ? 'Loading...' : cursor ? 'Load more' : 'No more'}
	</button>
	<span class="muted">
		{#if entries}
			showing: <code>{entries.length}</code>
			{#if cursor}· next cursor available{/if}
		{/if}
	</span>
</div>

{#if loading && !entries}
	<p class="muted">Loading audit log...</p>
{:else if entries && entries.length > 0}
	<table>
		<thead>
			<tr>
				<th>created_at</th>
				<th>user_id</th>
				<th>org_id</th>
				<th>action</th>
				<th>resource_type</th>
				<th>resource_id</th>
				<th>result</th>
				<th>ip</th>
			</tr>
		</thead>
		<tbody>
			{#each entries as e (e.audit_id)}
				<tr>
					<td class="muted">{e.created_at}</td>
					<td><code>{e.user_id}</code></td>
					<td><code>{e.org_id}</code></td>
					<td><code>{e.action}</code></td>
					<td><code>{e.resource_type}</code></td>
					<td><code>{e.resource_id}</code></td>
					<td>
						<span class="result result-{e.result}">{e.result}</span>
					</td>
					<td class="muted"><code>{e.ip ?? '-'}</code></td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else if !loading}
	<p class="muted">No audit entries.</p>
{/if}

<style>
	.page-header h1 {
		font-size: 1.4rem;
		margin: 0 0 0.3rem 0;
	}

	.page-header p {
		color: var(--text-muted);
		margin: 0 0 1rem 0;
		font-size: 0.85rem;
	}

	.toolbar {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		margin-bottom: 1rem;
	}

	.toolbar .muted {
		margin-left: 0.5rem;
		font-size: 0.85rem;
	}

	.muted {
		color: var(--text-muted);
	}

	button.secondary {
		background: var(--bg-card);
	}

	button.secondary:hover {
		background: var(--border);
	}

	.error {
		background: var(--danger);
		color: white;
		padding: 0.8rem 1rem;
		border-radius: var(--radius);
		margin-bottom: 1rem;
	}

	.result {
		display: inline-block;
		padding: 0.15rem 0.5rem;
		border-radius: 999px;
		font-size: 0.75rem;
		font-family: var(--font-mono);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.result-success { background: #14432e; color: #4ade80; }
	.result-denied { background: #4a3514; color: #fbbf24; }
	.result-error { background: #4a1414; color: #f87171; }
</style>
