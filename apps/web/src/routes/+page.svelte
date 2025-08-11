<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { CardSv, BadgeSv } from '$lib/ui/components';
	import RestaurantSearchSv from '$lib/ui/components/RestaurantSearchSv.svelte';

	let places: Place[] = [];
	let selectedPlace: Place | null = null;
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

		<CardSv variant="outlined" color="surface" class="space-y-4">
			<div>
				<label
					for="searchQuery"
					class="text-surface-700 dark:text-surface-300 mb-2 block text-sm font-medium"
				>
					Search restaurant by name:
				</label>
				<RestaurantSearchSv onPlaceSelected={(place: Place) => selectedPlace = place} />
			</div>
		</CardSv>

		{#if selectedPlace}
			<div class="space-y-4">
				<h3 class="text-primary-800 dark:text-primary-200 text-xl font-semibold">
					Selected Restaurant:
				</h3>
				<CardSv variant="outlined" color="surface" class="space-y-3">
					<h4 class="text-primary-800 dark:text-primary-200 font-bold">
						{selectedPlace.displayName?.text || selectedPlace.name}
					</h4>
					{#if selectedPlace.rating}
						<div>
							<BadgeSv variant="filled" color="primary" size="sm">
								Rating: {selectedPlace.rating}/5
							</BadgeSv>
						</div>
					{/if}
					{#if selectedPlace.formattedAddress}
						<p class="text-surface-600 dark:text-surface-400 text-sm">
							<strong>Address:</strong>
							{selectedPlace.formattedAddress}
						</p>
					{/if}
					{#if selectedPlace.types && selectedPlace.types.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each selectedPlace.types as type}
								<BadgeSv variant="outlined" color="secondary" size="sm">
									{type}
								</BadgeSv>
							{/each}
						</div>
					{/if}
				</CardSv>
			</div>
		{/if}

		{#if places.length > 0}
			<div class="space-y-4">
				<h3 class="text-primary-800 dark:text-primary-200 text-xl font-semibold">
					Search Results:
				</h3>
				<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each places as place}
						<CardSv variant="outlined" color="surface" class="space-y-3">
							<h4 class="text-primary-800 dark:text-primary-200 font-bold">
								{place.displayName?.text || place.name}
							</h4>
							{#if place.rating}
								<div>
									<BadgeSv variant="filled" color="primary" size="sm">
										Rating: {place.rating}/5
									</BadgeSv>
								</div>
							{/if}
							{#if place.formattedAddress}
								<p class="text-surface-600 dark:text-surface-400 text-sm">
									<strong>Address:</strong>
									{place.formattedAddress}
								</p>
							{/if}
						</CardSv>
					{/each}
				</div>
			</div>
		{/if}
	</section>
</div>
