<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Card, Badge } from 'flowbite-svelte';

	const { place } = $props<{ place: Place }>();
</script>

<Card
	class="overflow-hidden rounded-2xl border-0 bg-white shadow-xl transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl dark:bg-gray-800 dark:shadow-gray-900/50"
>
	<div class="space-y-4 p-6">
		<!-- Restaurant Name -->
		<header class="space-y-2">
			<h3 class="text-xl font-bold leading-tight text-gray-900 dark:text-white">
				{place.displayName?.text || place.name}
			</h3>
		</header>

		<!-- Rating Section -->
		{#if place.rating}
			<div class="flex items-center space-x-2">
				<div class="flex items-center space-x-1">
					{#each Array(5) as _, i}
						<svg
							class="h-5 w-5 {i < Math.floor(place.rating!)
								? 'text-yellow-400'
								: 'text-gray-300 dark:text-gray-600'}"
							fill="currentColor"
							viewBox="0 0 20 20"
							aria-hidden="true"
						>
							<path
								d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"
							/>
						</svg>
					{/each}
				</div>
				<Badge color="primary" class="rounded-full px-3 py-1.5 font-semibold">
					{place.rating}/5
				</Badge>
			</div>
		{/if}

		<!-- Address Section -->
		{#if place.formattedAddress}
			<div class="space-y-2">
				<div class="flex items-start space-x-2">
					<svg
						class="mt-0.5 h-5 w-5 flex-shrink-0 text-gray-500 dark:text-gray-400"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						aria-hidden="true"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
						/>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
						/>
					</svg>
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
				<h4 class="sr-only">
					Restaurant Types
				</h4>
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
