<script lang="ts">
	import '../app.css';
	import { AppBar } from '@skeletonlabs/skeleton-svelte';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	let { children } = $props();

</script>

<div class="flex min-h-screen flex-col">
	<Header />

	<main class="grid flex-grow justify-items-center pt-14">
		{@render children()}
	</main>

	<Footer />
</div>
