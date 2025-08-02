<script lang="ts">
	import clients from '$lib/client/client';
	import type { RestaurantProto } from '$lib/client/generated/restaurants/v1/restaurant_pb';
	import { ProgressRing } from '@skeletonlabs/skeleton-svelte';
	import { onDestroy } from 'svelte';

	let restaurants: RestaurantProto[] = [];
	let loading = false;
	let showLoader = false;
	let loaderTimer: NodeJS.Timeout
	let error = '';

	async function fetchRestaurants(): Promise<void> {
		loading = true;
		showLoader = false;
		clearTimeout(loaderTimer);

		loaderTimer = setTimeout(() => {
			if (loading) showLoader = true;
		}, 500);

		error = '';
		try {
			await sleep(3000);
			const response = await clients.restaurants.listRestaurants({ page: 1, pageSize: 20 });
			restaurants = response.restaurants ?? [];
		} catch (e: any) {
			if (e.code && e.details) {
				error = `Error ${e.code}: ${e.details}`;
			} else if (e.message) {
				error = e.message;
			} else {
				error = 'Failed to fetch restaurants';
			}
		} finally {
			loading = false;
			clearTimeout(loaderTimer);
		}
	}

	onDestroy(() => clearTimeout(loaderTimer));

	function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
	}
</script>

<div class="animate-pulse">
	<button type="button" class="btn preset-filled-primary-500" onclick={fetchRestaurants}>Load Restaurants</button>
</div>

{#if loading && showLoader}
	<ProgressRing value={null} />
{:else if error}
	<p style="color: red;">{error}</p>
{:else if !loading && restaurants.length}
	<ul>
		{#each restaurants as restaurant}
			<li>{restaurant.name}</li>
		{/each}
	</ul>
{/if}
