<script lang="ts">
	import '../app.css';
	import { theme } from '$lib/stores/theme';
	import { authStore } from '$lib/stores/auth';
	import { authService } from '$lib/services/auth.service';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { ModeWatcher } from 'tailwind-variants/svelte';
	import { onMount } from 'svelte';

	const { children } = $props();

	onMount(() => {
		if (typeof window !== 'undefined') {
			const storedTheme = localStorage.getItem('theme');
			if (storedTheme) {
				theme.set(storedTheme as 'dark' | 'light');
			}
		}

		authService.verifySession();
	});

	async function handleGoogleLogin() {
		await authService.initiateGoogleLogin();
	}

	async function handleLogout() {
		await authService.logout();
	}
</script>

<ModeWatcher />

<div class="min-h-screen bg-background text-foreground">
	<nav class="container mx-auto flex h-16 items-center justify-between px-4">
		<a href="/" class="text-xl font-bold">RestoRate</a>
		<div class="flex items-center gap-4">
			{#if $authStore.isAuthenticated}
				<span class="text-sm text-muted-foreground">
					Welcome, {$authStore.user?.name || $authStore.user?.email || 'User'}
				</span>
				<a href="/restaurants" class="hover:text-accent-foreground">Restaurants</a>
				<a href="/users" class="hover:text-accent-foreground">Users</a>
				<button
					class="px-3 py-1 text-sm rounded-md hover:bg-accent hover:text-accent-foreground"
					onclick={handleLogout}
				>
					Logout
				</button>
			{:else}
				<button
					class="px-3 py-1 text-sm rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground"
					onclick={handleGoogleLogin}
				>
					Sign in with Google
				</button>
			{/if}
			<ThemeToggle />
		</div>
	</nav>
	<main class="container mx-auto px-4 py-8">
		{@render children()}
	</main>
</div>
