<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import { SOURCES, type IngestStatusData, type SourceName } from '$lib/types';

	let sources = $state<IngestStatusData[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let actionInFlight = $state<SourceName | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const r = await api.listSources(opts);
			sources = r.data.sources;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	async function runAction(source: SourceName, action: 'pull' | 'sync'): Promise<void> {
		actionInFlight = source;
		const opts = withStoredPathY();
		try {
			if (action === 'pull') {
				await api.pullSource(source, opts);
			} else {
				await api.syncSource(source, opts);
			}
			await load();
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			actionInFlight = null;
		}
	}

	onMount(load);
</script>

<header class="page-header">
	<h1>Ingest</h1>
	<p>5 source plugin 의 state + pull / sync trigger (umbrella doc §3.5 + §3.7.1)</p>
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

{#if loading}
	<p class="muted">Loading sources...</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>source</th>
				<th>status</th>
				<th>last_pulled_at</th>
				<th>last_synced_at</th>
				<th>items_count</th>
				<th>error_count</th>
				<th>actions</th>
			</tr>
		</thead>
		<tbody>
			{#each sources as s (s.source)}
				<tr>
					<td><code>{s.source}</code></td>
					<td>
						<span class="status status-{s.health}">{s.health}</span>
					</td>
					<td class="muted">{s.last_sync ?? '-'}</td>
					<td class="muted">{s.last_error ? `${s.last_error.code}: ${s.last_error.message}` : '—'}</td>
					<td class="num">{s.metrics?.items_count ?? '—'}</td>
					<td class="num {s.last_error ? 'danger' : ''}">{s.last_error ? 1 : 0}</td>
					<td class="actions">
						<button
							disabled={actionInFlight === s.source}
							onclick={() => runAction(s.source as SourceName, 'pull')}
						>
							{actionInFlight === s.source ? '...' : 'Pull'}
						</button>
						<button
							class="secondary"
							disabled={actionInFlight === s.source}
							onclick={() => runAction(s.source as SourceName, 'sync')}
						>
							{actionInFlight === s.source ? '...' : 'Sync'}
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>

	<section class="card">
		<h2>Available source (5 plugin, M-v0.2.3 운영 / M-v0.2.0 mock)</h2>
		<ul class="plugin-list">
			{#each SOURCES as s (s)}
				<li><code>{s}</code></li>
			{/each}
		</ul>
	</section>
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

	.status {
		display: inline-block;
		padding: 0.15rem 0.5rem;
		border-radius: 999px;
		font-size: 0.75rem;
		font-family: var(--font-mono);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-active { background: #14432e; color: #4ade80; }
	.status-degraded { background: #4a3514; color: #fbbf24; }
	.status-disabled { background: #2a2a2a; color: var(--text-muted); }
	.status-error { background: #4a1414; color: #f87171; }

	.muted { color: var(--text-muted); }
	.danger { color: var(--danger); }

	.num {
		font-family: var(--font-mono);
		text-align: right;
	}

	.actions {
		display: flex;
		gap: 0.3rem;
	}

	.actions button {
		padding: 0.3rem 0.7rem;
		font-size: 0.8rem;
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

	.card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.2rem;
		margin-top: 1rem;
	}

	.card h2 {
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--text-secondary);
		margin: 0 0 0.6rem 0;
		letter-spacing: 0.05em;
	}

	.plugin-list {
		display: flex;
		gap: 0.6rem;
		flex-wrap: wrap;
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.plugin-list li {
		background: var(--bg-card);
		padding: 0.3rem 0.6rem;
		border-radius: var(--radius);
		font-size: 0.85rem;
	}
</style>
