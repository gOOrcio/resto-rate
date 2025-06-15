<script lang="ts">
	import '../app.css';
	import { theme } from '$lib/stores/theme';
	import { onMount } from 'svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	const { children } = $props();

	onMount(() => {
		const stored = localStorage.getItem('theme');
		if (stored) {
			theme.set(stored as 'light' | 'dark');
		} else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
			theme.set('dark');
		}
	});
</script>

<div class="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors">
	<nav class="border-b border-gray-200 dark:border-gray-700">
		<div class="container mx-auto px-4 py-3 flex justify-between items-center">
			<a href="/" class="text-xl font-bold">Resto Rate</a>
			<ThemeToggle />
		</div>
	</nav>

	{@render children()}
</div>
