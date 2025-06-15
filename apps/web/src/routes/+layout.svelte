<script lang="ts">
	import '../app.css';
	import { theme } from '$lib/stores/theme';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { onMount } from 'svelte';

	const { children } = $props();

	onMount(() => {
		if (typeof window !== 'undefined') {
			const storedTheme = localStorage.getItem('theme');
			if (storedTheme) {
				theme.set(storedTheme as 'dark' | 'light');
			}
		}
	});
</script>

<div class="min-h-screen bg-background text-foreground">
	<nav class="container mx-auto flex h-16 items-center justify-between px-4">
		<a href="/" class="text-xl font-bold">RestoRate</a>
		<div class="flex items-center gap-4">
			<a href="/users" class="hover:text-accent-foreground">Users</a>
			<ThemeToggle />
		</div>
	</nav>
	<main class="container mx-auto px-4 py-8">
		{@render children()}
	</main>
</div>
