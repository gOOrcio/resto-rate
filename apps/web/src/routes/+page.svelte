<script lang="ts">
	import { onMount } from 'svelte';
	import { createPageLogger } from '$lib/logger';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';

	// Create a safe logger for this page
	const logger = createPageLogger('home');

	onMount(() => {
		// Test logging on page mount
		logger.info('Home page mounted successfully');
		logger.debug('Testing debug level logging');
	});
</script>

<div class="container mx-auto max-w-4xl p-6">
	<h1 class="mb-8 text-3xl font-bold text-center tracking-tight">Resto Rate</h1>

	<div class="text-center space-y-6">
		<p class="text-lg text-muted-foreground mb-8">
			Welcome to Resto Rate - Rate and review your favorite restaurants
		</p>

		<div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-2xl mx-auto">
			<Card class="hover:bg-accent transition-colors">
				<CardHeader>
					<CardTitle>Login</CardTitle>
					<CardDescription>Sign in to your account</CardDescription>
				</CardHeader>
				<CardContent>
					<Button variant="link" href="/login" class="w-full">Go to Login</Button>
				</CardContent>
			</Card>

			<Card class="hover:bg-accent transition-colors">
				<CardHeader>
					<CardTitle>Users</CardTitle>
					<CardDescription>View all users</CardDescription>
				</CardHeader>
				<CardContent>
					<Button variant="link" href="/users" class="w-full">View Users</Button>
				</CardContent>
			</Card>

			<Card class="hover:bg-accent transition-colors">
				<CardHeader>
					<CardTitle>Restaurants</CardTitle>
					<CardDescription>Manage restaurants</CardDescription>
				</CardHeader>
				<CardContent>
					<Button 
						variant="link" 
						href="/restaurants" 
						class="w-full"
						on:click={() => logger.info('Navigating to restaurants page')}
					>
						View Restaurants
					</Button>
				</CardContent>
			</Card>
		</div>
	</div>
</div>
