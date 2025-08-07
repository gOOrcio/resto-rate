<script lang="ts">
	import clients from '$lib/client/client';
	import type { RestaurantProto } from '$lib/client/generated/restaurants/v1/restaurant_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { ProgressRing } from '@skeletonlabs/skeleton-svelte';
	import { onMount, onDestroy } from 'svelte';

	let restaurants: RestaurantProto[] = [];
	let places: Place[] = [];
	let loading = false;
	let showLoader = false;
	let loaderTimer: NodeJS.Timeout;
	let error = '';
	let searchQuery = 'Provide restaurant name';
	let requestedFields = ['name', 'displayName', 'rating', 'formattedAddress'];

	async function fetchRestaurants(): Promise<void> {
		loading = true;
		showLoader = false;
		clearTimeout(loaderTimer);

		loaderTimer = setTimeout(() => {
			if (loading) showLoader = true;
		}, 500);

		error = '';
		try {
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

	async function searchPlaces(): Promise<void> {
		loading = true;
		showLoader = false;
		clearTimeout(loaderTimer);

		loaderTimer = setTimeout(() => {
			if (loading) showLoader = true;
		}, 500);

		error = '';
		try {
			const response = await clients.googleMaps.searchText({
				textQuery: searchQuery,
				includedType: 'restaurant',
				strictTypeFiltering: true,
				maxResultCount: 10,
				requestedFields: requestedFields
			});
			places = response.places ?? [];
		} catch (e: any) {
			if (e.code && e.details) {
				error = `Error ${e.code}: ${e.details}`;
			} else if (e.message) {
				error = e.message;
			} else {
				error = 'Failed to search places';
			}
		} finally {
			loading = false;
			clearTimeout(loaderTimer);
		}
	}

	onMount(() => {
		fetchRestaurants();
	});

	onDestroy(() => clearTimeout(loaderTimer));
</script>

<div class="prose">
	<h1>Resto Rate</h1>

	<h2>Restaurants from Database</h2>
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

	<h2>Google Places API with Dynamic Field Selection</h2>
	<div class="mb-4">
		<label for="searchQuery" class="mb-2 block text-sm font-medium">Search Query:</label>
		<input
			id="searchQuery"
			type="text"
			bind:value={searchQuery}
			class="w-full rounded border p-2"
		/>
	</div>

	<button
		on:click={searchPlaces}
		disabled={loading}
		class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 disabled:opacity-50"
	>
		{loading ? 'Searching...' : 'Search Places'}
	</button>

	{#if places.length > 0}
		<h3>Search Results:</h3>
		<div class="space-y-4">
			{#each places as place}
				<div class="rounded border p-4">
					<h4 class="font-bold">{place.displayName?.text || place.name}</h4>
					{#if place.rating}
						<p>Rating: {place.rating}/5</p>
					{/if}
					{#if place.formattedAddress}
						<p>Address: {place.formattedAddress}</p>
					{/if}
					{#if place.priceLevel}
						<p>Price Level: {place.priceLevel}</p>
					{/if}
					{#if place.businessStatus}
						<p>Status: {place.businessStatus}</p>
					{/if}
					{#if place.photos && place.photos.length > 0}
						<p>Photos: {place.photos.length} available</p>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
