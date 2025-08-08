<script lang="ts">
	import clients from '$lib/client/client';
	import type { RestaurantProto } from '$lib/client/generated/restaurants/v1/restaurant_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { ProgressRing } from '@skeletonlabs/skeleton-svelte';
	import { Button, Card, Input, Badge } from '$lib/ui/components';
	import { onMount, onDestroy } from 'svelte';

	let restaurants: RestaurantProto[] = [];
	let places: Place[] = [];
	let loading = false;
	let showLoader = false;
	let loaderTimer: NodeJS.Timeout;
	let error = '';
	let searchQuery = '';

	async function searchPlaces(): Promise<void> {
		loading = true;
		showLoader = false;
		clearTimeout(loaderTimer);

		loaderTimer = setTimeout(() => {
			if (loading) showLoader = true;
		}, 500);

		error = '';
		try {
			const response = await clients.googleMaps.searchRestaurants({
				textQuery: searchQuery,
				languageCode: 'pl',
				regionCode: 'pl'
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

	onDestroy(() => clearTimeout(loaderTimer));
</script>

<div class="container mx-auto max-w-6xl space-y-8 p-6">
	<header class="space-y-4 text-center">
		<h1 class="text-primary-900 dark:text-primary-100 text-4xl font-bold">Resto Rate</h1>
		<p class="text-surface-600 dark:text-surface-400 text-lg">
			Discover, rate, and review the best restaurants around you
		</p>
	</header>

	<section class="space-y-6">
		<h2 class="text-primary-800 dark:text-primary-200 text-2xl font-semibold">
			Google Places API Search
		</h2>

		<Card variant="outlined" color="surface" class="space-y-4">
			<div>
				<label
					for="searchQuery"
					class="text-surface-700 dark:text-surface-300 mb-2 block text-sm font-medium"
				>
					Search restaurant by name:
				</label>
				<input
					id="searchQuery"
					type="text"
					bind:value={searchQuery}
					class="input preset-outlined-surface-200-800 w-full"
				/>
			</div>

			<Button
				onclick={searchPlaces}
				disabled={loading}
				variant="filled"
				color="primary"
				size="md"
				class="w-full sm:w-auto"
			>
				{loading ? 'Searching...' : 'Search Places'}
			</Button>
		</Card>

		{#if places.length > 0}
			<div class="space-y-4">
				<h3 class="text-primary-800 dark:text-primary-200 text-xl font-semibold">
					Search Results:
				</h3>
				<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each places as place}
						<Card variant="outlined" color="surface" class="space-y-3">
							<h4 class="text-primary-800 dark:text-primary-200 font-bold">
								{place.displayName?.text || place.name}
							</h4>
							{#if place.rating}
								<div>
									<Badge variant="filled" color="primary" size="sm">
										Rating: {place.rating}/5
									</Badge>
								</div>
							{/if}
							{#if place.formattedAddress}
								<p class="text-surface-600 dark:text-surface-400 text-sm">
									<strong>Address:</strong>
									{place.formattedAddress}
								</p>
							{/if}
						</Card>
					{/each}
				</div>
			</div>
		{/if}
	</section>
</div>
