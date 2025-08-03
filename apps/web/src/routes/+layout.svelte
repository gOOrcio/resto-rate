<script lang="ts">
	import '../app.css';
	import { AppBar } from '@skeletonlabs/skeleton-svelte';
	import Footer from '$lib/ui/navigation/Footer.svelte';
	import Header from '$lib/ui/navigation/Header.svelte';
	import Calendar from '@lucide/svelte/icons/calendar';
	import CircleUser from '@lucide/svelte/icons/circle-user';
	let { children } = $props();

	let iconSize = 26;
</script>

<div class="flex min-h-screen flex-col">
	<Header />

	<main class="grid flex-grow justify-items-center pt-14">
		{@render children()}
	</main>

	<Footer />
</div>
