<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	
	export let fallback: string = '/';
	
	onMount(() => {
		// Check authentication status on mount
		const unsubscribe = authStore.subscribe(state => {
			if (!state.isAuthenticated && !state.isLoading) {
				goto(fallback);
			}
		});
		
		return unsubscribe;
	});
</script>

{#if $authStore.isAuthenticated}
	<slot />
{:else if $authStore.isLoading}
	<div class="flex items-center justify-center min-h-[200px]">
		<div class="text-center">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
			<p class="text-muted-foreground">Loading...</p>
		</div>
	</div>
{/if} 