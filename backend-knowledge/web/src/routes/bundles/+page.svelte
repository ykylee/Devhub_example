<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import type { Bundle } from '$lib/types';

	let bundles = $state<Bundle[] | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const r = await api.listBundles(opts);
			bundles = r.data.items;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<header class="page-header">
	<h1>Bundles</h1>
	<p>OKF v0.1 bundle = 1 source plugin 의 curated concept set (umbrella doc §3.6)</p>
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

{#if loading}
	<p class="muted">Loading bundles...</p>
{:else if bundles}
	<table>
		<thead>
			<tr>
				<th>name</th>
				<th>version</th>
				<th>concept_count</th>
				<th>updated_at</th>
				<th>updated_by</th>
				<th>visibility</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each bundles as b (b.name + b.version)}
				<tr>
					<td><code>{b.name}</code></td>
					<td><code>{b.version}</code></td>
					<td class="num">{b.concept_count}</td>
					<td class="muted">{b.updated_at}</td>
					<td class="muted"><code>{b.updated_by}</code></td>
					<td><code>{b.visibility}</code></td>
					<td>
						<a href="/bundles/{encodeURIComponent(b.name)}">detail</a>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else}
	<p class="muted">No bundles.</p>
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

	.num {
		font-family: var(--font-mono);
		text-align: right;
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
