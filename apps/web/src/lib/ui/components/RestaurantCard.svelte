<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Card, Badge, Rating, Star, type RatingIconProps } from 'flowbite-svelte';
	import { MapPinAltOutline } from 'flowbite-svelte-icons';
	import { v4 as uuidv4 } from 'uuid';
	import { restaurantCardTheme, ratingTheme } from '$lib/ui/theme/components';

	const { place } = $props<{ place: Place }>();
	const wrapper = (props: RatingIconProps) => (anchor: any, _props: RatingIconProps) =>
		Star(anchor, { ..._props, ...props });
</script>

<Card class={restaurantCardTheme.base}>
	<div class="space-y-4 p-6">
		<!-- Restaurant Name -->
		<header class="space-y-2">
			<h3 class="text-xl font-bold leading-tight text-gray-900 dark:text-white">
				{place.displayName?.text || place.name}
			</h3>
		</header>

		<!-- Rating Section with Enhanced Background -->
		{#if place.rating}
			<div class={ratingTheme.base}>
				<div class="flex items-center justify-between">
					<div class="flex items-center space-x-3">
						<Rating
							id={uuidv4()}
							total={5}
							size={24}
							rating={place.rating}
							icon={wrapper({ fillColor: '#008800', strokeColor: '#008800' })}
							class="rating-primary flex items-center space-x-1"
						>
							{#snippet text()}
								<div class="ml-3">
									<p class={ratingTheme.text}>
										{place.rating.toFixed(1)}/5
									</p>
								</div>
							{/snippet}
						</Rating>
					</div>
				</div>
			</div>
		{/if}

		<!-- Address Section -->
		{#if place.formattedAddress}
			<div class="space-y-2">
				<div class="flex items-start space-x-2">
					<MapPinAltOutline class="h-6 w-6 shrink-0 text-gray-500 dark:text-gray-400" />
					<div class="min-w-0 flex-1">
						<p class="text-sm font-medium text-gray-700 dark:text-gray-300">
							<strong class="sr-only">Address:</strong>
							{place.formattedAddress}
						</p>
					</div>
				</div>
			</div>
		{/if}

		<!-- Restaurant Types -->
		{#if place.types && place.types.length > 0}
			<div class="space-y-2">
				<h4 class="sr-only">Restaurant Types</h4>
				<div class="flex flex-wrap gap-2">
					{#each place.types as type}
						<Badge
							color="secondary"
							class="rounded-full bg-blue-100 px-3 py-1.5 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-200"
						>
							{type}
						</Badge>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</Card>
