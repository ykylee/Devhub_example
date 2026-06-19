<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';

	let health = $state<{ status: string; version: string } | null>(null);
	let alerts = $state<{ alerts: Array<unknown>; by_severity: { info: number; warning: number; critical: number } } | null>(null);
	let recentAudit = $state<{ items: Array<{ event: string; timestamp: string; user_id: string | null; success: boolean }>; total: number } | null>(null);
	let bundles = $state<{ items: Array<{ name: string; concept_count: number }>; total: number } | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const [h, a, au, bs] = await Promise.all([
				api.health(opts),
				api.listAlerts(opts),
				api.listAudit({ limit: 10 }, opts),
				api.listBundles(opts)
			]);
			health = h.data;
			alerts = a.data;
			recentAudit = au.data;
			bundles = bs.data;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function fmtTime(iso: string): string {
		try {
			return new Date(iso).toLocaleString();
		} catch {
			return iso;
		}
	}
</script>

<header class="page-header">
	<h1>Dashboard</h1>
	<p>v0.2.0 PoC standalone backend knowledge tool (umbrella doc §1.2 G7 + ADR-0035)</p>
	<button type="button" onclick={load} disabled={loading}>{loading ? 'Loading...' : 'Refresh'}</button>
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

<section class="cards">
	<article class="card">
		<h2>Health</h2>
		{#if health}
			<p>status: <code>{health.status}</code></p>
			<p>version: <code>{health.version}</code></p>
		{:else}
			<p class="muted">Loading...</p>
		{/if}
	</article>

	<article class="card">
		<h2>Alerts (3-tier)</h2>
		{#if alerts}
			<p>info: <code>{alerts.by_severity.info}</code></p>
			<p class="warn">warning: <code>{alerts.by_severity.warning}</code></p>
			<p class="crit">critical: <code>{alerts.by_severity.critical}</code></p>
			<p class="muted">total fired: {alerts.alerts.length}</p>
		{:else}
			<p class="muted">Loading...</p>
		{/if}
	</article>

	<article class="card">
		<h2>Bundles</h2>
		{#if bundles}
			<p>total: <code>{bundles.total}</code></p>
			<ul class="compact">
				{#each bundles.items as b (b.name)}
					<li><a href={`/bundles/${b.name}`}>{b.name}</a> ({b.concept_count})</li>
				{/each}
			</ul>
		{:else}
			<p class="muted">Loading...</p>
		{/if}
	</article>
</section>

<section class="card wide">
	<h2>Recent audit (10)</h2>
	{#if recentAudit}
		<p>total in window: <code>{recentAudit.total}</code></p>
		<table>
			<thead>
				<tr>
					<th>time</th>
					<th>event</th>
					<th>user</th>
					<th>success</th>
				</tr>
			</thead>
			<tbody>
				{#each recentAudit.items as e (e.timestamp + e.event)}
					<tr>
						<td class="time">{fmtTime(e.timestamp)}</td>
						<td><code>{e.event}</code></td>
						<td>{e.user_id ?? '-'}</td>
						<td class={e.success ? 'ok' : 'fail'}>{e.success ? '✓' : '✗'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p class="muted">Loading...</p>
	{/if}
</section>

<style>
	.page-header {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.page-header h1 {
		font-size: 1.4rem;
		margin: 0;
	}

	.page-header p {
		color: var(--text-muted);
		margin: 0;
		flex: 1;
		font-size: 0.85rem;
	}

	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.2rem;
	}

	.card.wide {
		grid-column: 1 / -1;
	}

	.card h2 {
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--text-secondary);
		margin: 0 0 0.6rem 0;
		letter-spacing: 0.05em;
	}

	.card p {
		margin: 0.2rem 0;
		font-size: 0.9rem;
	}

	.card code {
		color: var(--accent);
	}

	.warn code {
		color: var(--warning);
	}

	.crit code {
		color: var(--danger);
	}

	.muted {
		color: var(--text-muted);
	}

	.error {
		background: var(--danger);
		color: white;
		padding: 0.8rem 1rem;
		border-radius: var(--radius);
		margin-bottom: 1rem;
	}

	.compact {
		list-style: none;
		padding: 0;
		margin: 0.4rem 0 0 0;
		font-size: 0.85rem;
	}

	.compact li {
		padding: 0.15rem 0;
	}

	table {
		margin-top: 0.5rem;
	}

	.time {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.ok {
		color: var(--success);
	}

	.fail {
		color: var(--danger);
	}
</style>
