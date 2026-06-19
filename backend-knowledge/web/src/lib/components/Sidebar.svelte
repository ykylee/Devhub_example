<script lang="ts">
	import { page } from '$app/stores';

	interface Route {
		path: string;
		label: string;
	}

	const routes: Route[] = [
		{ path: '/', label: 'Dashboard' },
		{ path: '/concepts', label: 'Concepts' },
		{ path: '/ingest', label: 'Ingest' },
		{ path: '/bundles', label: 'Bundles' },
		{ path: '/raw', label: 'Raw' },
		{ path: '/audit', label: 'Audit' }
	];

	function isActive(routePath: string, pathname: string): boolean {
		if (routePath === '/') return pathname === '/';
		return pathname === routePath || pathname.startsWith(routePath + '/');
	}
</script>

<nav class="sidebar">
	<header>
		<h1>backend-knowledge</h1>
		<p class="sub">v0.2.0 PoC</p>
	</header>

	<ul>
		{#each routes as route (route.path)}
			<li>
				<a href={route.path} class:active={isActive(route.path, $page.url.pathname)}>
					{route.label}
				</a>
			</li>
		{/each}
	</ul>

	<div class="path-y-slot">
		<slot />
	</div>
</nav>

<style>
	.sidebar {
		width: 220px;
		min-height: 100vh;
		background: var(--bg-secondary);
		border-right: 1px solid var(--border);
		padding: 1.2rem 0.8rem;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
	}

	header {
		margin-bottom: 1.5rem;
		padding: 0 0.4rem;
	}

	h1 {
		font-size: 0.95rem;
		margin: 0;
		font-weight: 600;
	}

	.sub {
		font-size: 0.7rem;
		color: var(--text-muted);
		margin: 0.2rem 0 0 0;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	li a {
		display: block;
		padding: 0.55rem 0.8rem;
		border-radius: var(--radius);
		color: var(--text-secondary);
		text-decoration: none;
		margin-bottom: 0.2rem;
		font-size: 0.9rem;
	}

	li a:hover,
	li a.active {
		background: var(--bg-card);
		color: var(--text-primary);
		text-decoration: none;
	}

	.path-y-slot {
		margin-top: auto;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}
</style>
