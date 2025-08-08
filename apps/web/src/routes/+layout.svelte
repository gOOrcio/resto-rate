<script lang="ts">
	import '../app.css';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	let { children } = $props();
</script>

<div class="bg-surface-50 dark:bg-surface-950 flex min-h-screen flex-col">
	<Header />

	<main class="flex-grow pt-14">
		{@render children()}
	</main>

	<Footer />
</div>
