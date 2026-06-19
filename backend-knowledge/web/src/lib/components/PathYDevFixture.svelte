<script lang="ts">
	import { onMount } from 'svelte';
	import { clearPathY, loadPathY, makeDevPathY, savePathY } from '$lib/path-y';
	import type { PathYUserContext } from '$lib/types';

	let ctx = $state<PathYUserContext | null>(null);
	let expanded = $state(false);

	onMount(() => {
		ctx = loadPathY();
	});

	function useDev(): void {
		const dev = makeDevPathY();
		savePathY(dev);
		ctx = dev;
		expanded = false;
	}

	function clear(): void {
		clearPathY();
		ctx = null;
		expanded = false;
	}
</script>

<div class="path-y-dev">
	<h3>Path Y dev fixture</h3>

	{#if ctx}
		<div class="info">
			<div><span class="key">user</span> <code>{ctx.user_id}</code></div>
			<div><span class="key">org</span> <code>{ctx.org_id}</code></div>
			<div><span class="key">roles</span> <code>{ctx.roles.join(', ')}</code></div>
			<button type="button" onclick={clear} class="small danger">Clear</button>
		</div>
	{:else}
		<button type="button" onclick={useDev} class="small">Use dev fixture</button>
		<p class="hint">standalone (gateway 없이). PoC default.</p>
	{/if}
</div>

<style>
	.path-y-dev {
		color: var(--text-secondary);
		font-size: 0.78rem;
	}

	h3 {
		font-size: 0.7rem;
		margin: 0 0 0.5rem 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
	}

	.info {
		background: var(--bg-primary);
		padding: 0.5rem;
		border-radius: var(--radius);
		margin-bottom: 0.4rem;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.info code {
		font-size: 0.7rem;
		color: var(--accent);
		word-break: break-all;
	}

	.key {
		color: var(--text-muted);
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.hint {
		font-size: 0.65rem;
		color: var(--text-muted);
		margin: 0.4rem 0 0 0;
		line-height: 1.4;
	}

	button.small {
		font-size: 0.7rem;
		padding: 0.25rem 0.5rem;
		margin-top: 0.4rem;
	}

	button.danger {
		background: var(--danger);
	}

	button.danger:hover {
		background: #b91c1c;
	}
</style>
