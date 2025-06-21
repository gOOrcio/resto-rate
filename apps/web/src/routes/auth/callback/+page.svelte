<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authService } from '$lib/services/auth.service';
	import { createPageLogger } from '$lib/logger';
	
	const logger = createPageLogger('auth-callback');
	
	onMount(async () => {
		try {
			const code = $page.url.searchParams.get('code');
			const error = $page.url.searchParams.get('error');
			
			if (error) {
				logger.error('OAuth error received', { error });
				goto('/?error=auth_failed');
				return;
			}
			
			if (!code) {
				logger.error('No authorization code received');
				goto('/?error=no_code');
				return;
			}
			
			logger.info('Processing Google OAuth callback');
			
			// Handle the OAuth callback using the auth service
			await authService.handleGoogleCallback(code);
			
			logger.info('OAuth callback completed successfully');
			
			// Redirect to home page
			goto('/');
		} catch (error) {
			logger.error('Error processing OAuth callback', { error });
			goto('/?error=auth_failed');
		}
	});
</script>

<div class="flex items-center justify-center min-h-screen">
	<div class="text-center">
		<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
		<h2 class="text-xl font-semibold mb-2">Processing Authentication</h2>
		<p class="text-muted-foreground">Please wait while we complete your sign-in...</p>
	</div>
</div> 