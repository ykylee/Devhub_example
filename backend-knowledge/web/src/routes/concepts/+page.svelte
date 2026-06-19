<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import { CONCEPT_TYPES, type ConceptSearchHit, type ConceptType, type SourceName } from '$lib/types';
	import { SOURCES } from '$lib/types';

	let q = $state('');
	let typeFilter = $state<ConceptType | ''>('');
	let bundleFilter = $state<SourceName | ''>('');
	let loading = $state(true);
	let error = $state<string | null>(null);
	let hits = $state<ConceptSearchHit[] | null>(null);
	let total = $state(0);

	async function search(): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const params: { q?: string; type?: string; bundle?: string; limit?: number } = { limit: 100 };
			if (q) params.q = q;
			if (typeFilter) params.type = typeFilter;
			if (bundleFilter) params.bundle = bundleFilter;
			const r = await api.listConcepts(params, opts);
			hits = r.data.hits;
			total = r.data.total;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
			hits = null;
		} finally {
			loading = false;
		}
	}

	onMount(search);

	function reset(): void {
		q = '';
		typeFilter = '';
		bundleFilter = '';
		search();
	}

	function detailHref(hit: ConceptSearchHit): string {
		const slug = hit.title || hit.concept_id.split('/').pop() || '';
		return `/concepts/${encodeURIComponent(hit.type)}/${encodeURIComponent(slug)}?bundle=${encodeURIComponent(hit.bundle)}`;
	}
</script>

<header class="page-header">
	<h1>Concepts</h1>
	<p>List of all OKF v0.1 concept across 5 source plugin (umbrella doc §3.5 + §3.7.2)</p>
</header>

<form class="filters" onsubmit={(e) => { e.preventDefault(); search(); }}>
	<input
		type="text"
		placeholder="search query"
		bind:value={q}
		aria-label="search query"
	/>
	<select bind:value={typeFilter} aria-label="type filter">
		<option value="">all types</option>
		{#each CONCEPT_TYPES as t (t)}
			<option value={t}>{t}</option>
		{/each}
	</select>
	<select bind:value={bundleFilter} aria-label="bundle filter">
		<option value="">all bundles</option>
		{#each SOURCES as s (s)}
			<option value={s}>{s}</option>
		{/each}
	</select>
	<button type="submit" disabled={loading}>{loading ? 'Searching...' : 'Search'}</button>
	<button type="button" onclick={reset} class="secondary">Reset</button>
</form>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

<div class="result-meta">
	{#if hits}
		<span>total: <code>{total}</code></span>
		<span>showing: <code>{hits.length}</code></span>
	{:else if loading}
		<span class="muted">Loading...</span>
	{/if}
</div>

{#if hits && hits.length > 0}
	<table>
		<thead>
			<tr>
				<th>bundle</th>
				<th>type</th>
				<th>title</th>
				<th>snippet</th>
				<th>score</th>
			</tr>
		</thead>
		<tbody>
			{#each hits as h (h.concept_id)}
				<tr>
					<td><code>{h.bundle}</code></td>
					<td><code>{h.type}</code></td>
					<td>
						<a href={detailHref(h)}>{h.title}</a>
					</td>
					<td class="snippet">{h.snippet}</td>
					<td class="num">{h.score.toFixed(2)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else if !loading && hits && hits.length === 0}
	<p class="empty">No concepts match the current filter.</p>
{/if}

<style>
	.page-header {
		margin-bottom: 1rem;
	}

	.page-header h1 {
		font-size: 1.4rem;
		margin: 0 0 0.3rem 0;
	}

	.page-header p {
		color: var(--text-muted);
		margin: 0;
		font-size: 0.85rem;
	}

	.filters {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1rem;
		flex-wrap: wrap;
	}

	.filters input {
		flex: 1;
		min-width: 200px;
	}

	.filters select {
		min-width: 140px;
	}

	button.secondary {
		background: var(--bg-card);
	}

	button.secondary:hover {
		background: var(--border);
	}

	.result-meta {
		display: flex;
		gap: 1rem;
		font-size: 0.85rem;
		color: var(--text-secondary);
		margin-bottom: 0.6rem;
	}

	.result-meta code {
		color: var(--accent);
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

	.empty {
		color: var(--text-muted);
		text-align: center;
		padding: 3rem 1rem;
	}

	.snippet {
		color: var(--text-muted);
		font-size: 0.85rem;
		max-width: 360px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.num {
		font-family: var(--font-mono);
		color: var(--accent);
		text-align: right;
	}
</style>
