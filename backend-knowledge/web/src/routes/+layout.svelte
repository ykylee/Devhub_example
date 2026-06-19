<script lang="ts">
	import '../app.css';
	import PathYDevFixture from '$lib/components/PathYDevFixture.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
</script>

<div class="app">
	<Sidebar>
		<PathYDevFixture />
	</Sidebar>
	<main>
		<slot />
	</main>
</div>

<style>
	.app {
		display: flex;
		min-height: 100vh;
	}

	main {
		flex: 1;
		padding: 1.5rem 2rem;
		overflow-x: auto;
		min-width: 0;
	}
</style>
