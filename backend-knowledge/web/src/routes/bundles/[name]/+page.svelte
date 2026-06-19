<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import type { BundleDetail, Concept } from '$lib/types';

	// M-v0.2.1+ scope: backend GET /api/v0-2/bundles/{name} 가 concepts array 미포함 (단일 metadata).
	// concepts list 는 backend GET /api/v0-2/bundles/{name}/concepts 별도 endpoint (PR #662) 추가 시 load.
	let bundle = $state<BundleDetail | null>(null);
	let concepts = $state<Concept[] | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		bundle = null;
		concepts = null;

		const name = $page.params.name ?? '';
		if (!name) {
			error = 'Missing bundle name in URL';
			loading = false;
			return;
		}

		const opts = withStoredPathY();
		try {
			const r = await api.getBundle(name, opts);
			bundle = r.data;
			// concepts 는 M-v0.2.1+ scope — backend GET /bundles/{name}/concepts 별도 endpoint 추가 후 load.
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<header class="page-header">
	{#if bundle}
		<h1>{bundle.name}</h1>
		<p class="breadcrumb">
			<a href="/bundles">bundles</a> / <code>{bundle.name}</code>
		</p>
	{:else}
		<h1>Bundle detail</h1>
	{/if}
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

{#if loading}
	<p class="muted">Loading...</p>
{:else if bundle}
	<section class="card">
		<h2>Bundle metadata</h2>
		<table class="kv">
			<tbody>
				<tr><th>name</th><td><code>{bundle.name}</code></td></tr>
				<tr><th>version</th><td><code>{bundle.version}</code></td></tr>
				<tr><th>concept_count</th><td>{bundle.concept_count}</td></tr>
				<tr><th>updated_at</th><td>{bundle.updated_at}</td></tr>
				<tr><th>updated_by</th><td><code>{bundle.updated_by}</code></td></tr>
				<tr><th>visibility</th><td><code>{bundle.visibility}</code></td></tr>
				{#if bundle.description}
					<tr><th>description</th><td>{bundle.description}</td></tr>
				{/if}
			</tbody>
		</table>
	</section>

	<section class="card">
		<h2>Concepts in this bundle ({concepts?.length ?? 0})</h2>
		{#if concepts && concepts.length > 0}
			<table>
				<thead>
					<tr>
						<th>type</th>
						<th>name</th>
						<th>title</th>
						<th>version</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each concepts as c (c.concept_id)}
						<tr>
							<td><code>{c.type}</code></td>
							<td><code>{c.name}</code></td>
							<td>{c.frontmatter.title ?? '-'}</td>
							<td><code>{c.version}</code></td>
							<td>
								<a
									href="/concepts/{encodeURIComponent(c.type)}/{encodeURIComponent(c.name)}?bundle={encodeURIComponent(bundle.name)}"
								>
									open
								</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{:else}
			<p class="muted">No concepts yet. Run /ingest → Pull to register raw → /curate to curate.</p>
		{/if}
	</section>
{/if}

<style>
	.page-header h1 {
		font-size: 1.4rem;
		margin: 0 0 0.3rem 0;
	}

	.breadcrumb {
		color: var(--text-muted);
		margin: 0 0 1rem 0;
		font-size: 0.85rem;
	}

	.card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.2rem;
		margin-bottom: 1rem;
	}

	.card h2 {
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--text-secondary);
		margin: 0 0 0.6rem 0;
		letter-spacing: 0.05em;
	}

	.kv th {
		text-align: left;
		width: 200px;
		font-weight: normal;
		color: var(--text-muted);
	}

	.kv td {
		font-family: var(--font-mono);
		font-size: 0.85rem;
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
</style>
