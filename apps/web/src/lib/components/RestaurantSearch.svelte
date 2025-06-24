<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import { apiClient } from '$lib/api';
	import { apiLogger } from '$lib/logger';
	import { cn } from '$lib/utils';
	import { buttonVariants } from '$lib/components/ui/button';

	let query = '';
	let loading = false;
	const dispatch = createEventDispatcher();

	async function handleSearch() {
		if (!query.trim()) return;

		loading = true;
		apiLogger.debug('Searching for places', { query });
		try {
			const response = await apiClient.searchPlaces(query);
			dispatch('searchresults', response.results);
		} catch (error) {
			apiLogger.error('Failed to search places', { error });
			dispatch('searcherror', 'Failed to search for restaurants.');
		} finally {
			loading = false;
		}
	}
</script>

<form class="flex w-full items-center space-x-2" on:submit|preventDefault={handleSearch}>
	<Input
		type="search"
		placeholder="Search for a restaurant on Google..."
		bind:value={query}
		disabled={loading}
	/>
	<button type="submit" disabled={loading} class={cn(buttonVariants())}>
		{loading ? 'Searching...' : 'Search'}
	</button>
</form>
