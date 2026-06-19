<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import type { Concept } from '$lib/types';

	let concept = $state<Concept | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		concept = null;

		const bundle = $page.url.searchParams.get('bundle') ?? 'homelab-mock';
		const type = $page.params.type ?? '';
		const name = $page.params.name ?? '';

		if (!type || !name) {
			error = 'Missing type or name in URL';
			loading = false;
			return;
		}

		const opts = withStoredPathY();
		try {
			const r = await api.getConcept(bundle, type, name, opts);
			concept = r.data;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<header class="page-header">
	{#if concept}
		<h1>{concept.frontmatter.title ?? concept.name}</h1>
		<p class="breadcrumb">
			<a href="/concepts">concepts</a> /
			<code>{concept.bundle}</code> /
			<code>{concept.type}</code> /
			<code>{concept.name}</code>
		</p>
	{:else}
		<h1>Concept detail</h1>
	{/if}
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

{#if loading}
	<p class="muted">Loading...</p>
{:else if concept}
	<section class="card">
		<h2>Frontmatter (12 field, OKF v0.1)</h2>
		<table class="kv">
			<tbody>
				<tr><th>type</th><td><code>{concept.frontmatter.type}</code></td></tr>
				<tr><th>title</th><td>{concept.frontmatter.title ?? '-'}</td></tr>
				<tr><th>description</th><td>{concept.frontmatter.description ?? '-'}</td></tr>
				<tr><th>resource</th><td>{concept.frontmatter.resource ?? '-'}</td></tr>
				<tr><th>tags</th><td><code>{concept.frontmatter.tags?.join(', ') ?? '-'}</code></td></tr>
				<tr><th>timestamp</th><td>{concept.frontmatter.timestamp ?? '-'}</td></tr>
				<tr><th>x_devhub_source</th><td><code>{concept.frontmatter.x_devhub_source ?? '-'}</code></td></tr>
				<tr><th>x_devhub_bundle</th><td><code>{concept.frontmatter.x_devhub_bundle ?? '-'}</code></td></tr>
				<tr><th>x_devhub_version</th><td><code>{concept.frontmatter.x_devhub_version}</code></td></tr>
				<tr><th>x_devhub_curator</th><td><code>{concept.frontmatter.x_devhub_curator}</code></td></tr>
				<tr><th>x_devhub_owner_org_id</th><td><code>{concept.frontmatter.x_devhub_owner_org_id ?? '-'}</code></td></tr>
				<tr><th>x_devhub_owner_user_id</th><td><code>{concept.frontmatter.x_devhub_owner_user_id ?? '-'}</code></td></tr>
				<tr><th>x_devhub_owner_org_unit_ids</th><td><code>{concept.frontmatter.x_devhub_owner_org_unit_ids?.join(', ') ?? '-'}</code></td></tr>
				<tr><th>x_devhub_owner_project_ids</th><td><code>{concept.frontmatter.x_devhub_owner_project_ids?.join(', ') ?? '-'}</code></td></tr>
				<tr><th>x_devhub_visibility</th><td><code>{concept.frontmatter.x_devhub_visibility}</code></td></tr>
			</tbody>
		</table>
	</section>

	<section class="card">
		<h2>Body (Markdown)</h2>
		<pre><code>{concept.body}</code></pre>
	</section>

	<section class="card">
		<h2>Cross-link</h2>
		<div class="links">
			<div>
				<h3>Out-links ({concept.cross_links_out?.length ?? 0})</h3>
				{#if concept.cross_links_out && concept.cross_links_out.length > 0}
					<ul>
						{#each concept.cross_links_out as l (l.target + (l.section ?? ''))}
							<li><code>{l.target}</code> <span class="muted">({l.type})</span></li>
						{/each}
					</ul>
				{:else}
					<p class="muted">none</p>
				{/if}
			</div>
			<div>
				<h3>In-links ({concept.cross_links_in?.length ?? 0})</h3>
				{#if concept.cross_links_in && concept.cross_links_in.length > 0}
					<ul>
						{#each concept.cross_links_in as l (l.source + (l.section ?? ''))}
							<li><code>{l.source}</code> <span class="muted">({l.type})</span></li>
						{/each}
					</ul>
				{:else}
					<p class="muted">none</p>
				{/if}
			</div>
		</div>
	</section>

	<section class="card">
		<h2>Metadata</h2>
		<table class="kv">
			<tbody>
				<tr><th>concept_id</th><td><code>{concept.concept_id}</code></td></tr>
				<tr><th>version</th><td><code>{concept.version}</code></td></tr>
				<tr><th>visibility</th><td><code>{concept.visibility}</code></td></tr>
				<tr><th>updated_by</th><td><code>{concept.updated_by}</code></td></tr>
			</tbody>
		</table>
	</section>
{/if}

<style>
	.page-header {
		margin-bottom: 1rem;
	}

	.page-header h1 {
		font-size: 1.4rem;
		margin: 0 0 0.3rem 0;
	}

	.breadcrumb {
		color: var(--text-muted);
		margin: 0;
		font-size: 0.85rem;
	}

	.breadcrumb code {
		color: var(--text-secondary);
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

	.card h3 {
		font-size: 0.78rem;
		color: var(--text-muted);
		margin: 0 0 0.4rem 0;
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

	pre {
		max-height: 400px;
		overflow: auto;
	}

	pre code {
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

	.links {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.links ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.links li {
		padding: 0.2rem 0;
		font-size: 0.85rem;
	}
</style>
