<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError, api, withStoredPathY } from '$lib/api';
	import type { RawRecord } from '$lib/types';

	let entries = $state<RawRecord[] | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let actionInFlight = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		error = null;
		const opts = withStoredPathY();
		try {
			const r = await api.listRaw({ limit: 100 }, opts);
			entries = r.data.items;
		} catch (e) {
			error = e instanceof ApiError ? `${e.code}: ${e.message}` : String(e);
		} finally {
			loading = false;
		}
	}

	async function remove(rawId: string): Promise<void> {
		if (!confirm(`Delete raw ${rawId}? This also removes the metadata sidecar.`)) return;
		actionInFlight = rawId;
		const opts = withStoredPathY();
		try {
			await api.deleteRaw(rawId, opts);
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
	<h1>Raw data</h1>
	<p>
		Per-source raw blob + sidecar metadata (umbrella doc §3.3.2 + Codex P1 fix). 봉투 암호화 v2 envelope
		로 저장.
	</p>
</header>

{#if error}
	<div class="error">Error: {error}</div>
{/if}

{#if loading}
	<p class="muted">Loading raw entries...</p>
{:else if entries && entries.length > 0}
	<table>
		<thead>
			<tr>
				<th>raw_id</th>
				<th>source</th>
				<th>type</th>
				<th>name</th>
				<th>visibility</th>
				<th>size</th>
				<th>registered_at</th>
				<th>sha256</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each entries as e (e.raw_id)}
				<tr>
					<td><code>{e.raw_id}</code></td>
					<td><code>{e.source}</code></td>
					<td><code>{e.type}</code></td>
					<td><code>{e.name}</code></td>
					<td><code>{e.visibility}</code></td>
				<td class="num">{e.size}</td>
				<td class="muted">{e.registered_at}</td>
				<td class="muted"><code>{e.sha256}</code></td>
					<td>
						<button
							class="danger"
							disabled={actionInFlight === e.raw_id}
							onclick={() => remove(e.raw_id)}
						>
							{actionInFlight === e.raw_id ? '...' : 'Delete'}
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else if !loading}
	<p class="muted">No raw data registered yet. Run /ingest → Pull to register from a source.</p>
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

	button.danger {
		background: var(--danger);
		padding: 0.3rem 0.7rem;
		font-size: 0.8rem;
	}

	button.danger:hover {
		opacity: 0.85;
	}

	button.danger:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
