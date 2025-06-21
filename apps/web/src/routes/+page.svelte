<script lang="ts">
	import { onMount } from 'svelte';
	import { createPageLogger } from '$lib/logger';
	import { authStore } from '$lib/stores/auth';
	import { authService } from '$lib/services/auth.service';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';

	// Create a safe logger for this page
	const logger = createPageLogger('home');

	onMount(() => {
		// Test logging on page mount
		logger.info('Home page mounted successfully');
		logger.debug('Testing debug level logging');
	});

	async function handleGoogleLogin() {
		await authService.initiateGoogleLogin();
	}
</script>

<div class="container mx-auto max-w-4xl p-6">
	<h1 class="mb-8 text-3xl font-bold text-center tracking-tight">Resto Rate</h1>

	{#if $authStore.isAuthenticated}
		<!-- Authenticated user content -->
		<div class="text-center space-y-6">
			<p class="text-lg text-muted-foreground mb-8">
				Welcome back, {$authStore.user?.name || $authStore.user?.email || 'User'}!
			</p>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto">
				<Card class="hover:bg-accent transition-colors">
					<CardHeader>
						<CardTitle>Restaurants</CardTitle>
						<CardDescription>Manage and rate restaurants</CardDescription>
					</CardHeader>
					<CardContent>
						<a
							href="/restaurants"
							class="block w-full text-center px-4 py-2 text-sm font-medium text-primary hover:underline"
						>
							View Restaurants
						</a>
					</CardContent>
				</Card>

				<Card class="hover:bg-accent transition-colors">
					<CardHeader>
						<CardTitle>Users</CardTitle>
						<CardDescription>View all users</CardDescription>
					</CardHeader>
					<CardContent>
						<a
							href="/users"
							class="block w-full text-center px-4 py-2 text-sm font-medium text-primary hover:underline"
						>
							View Users
						</a>
					</CardContent>
				</Card>
			</div>
		</div>
	{:else}
		<!-- Unauthenticated user content -->
		<div class="text-center space-y-6">
			<p class="text-lg text-muted-foreground mb-8">
				Welcome to Resto Rate - Rate and review your favorite restaurants
			</p>

			<div class="max-w-md mx-auto">
				<Card class="hover:bg-accent transition-colors">
					<CardHeader>
						<CardTitle>Get Started</CardTitle>
						<CardDescription>Sign in with Google to start rating restaurants</CardDescription>
					</CardHeader>
					<CardContent>
						<button
							class="w-full px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground"
							onclick={handleGoogleLogin}
						>
							Sign in with Google
						</button>
					</CardContent>
				</Card>
			</div>
		</div>
	{/if}
</div>
