<script lang="ts">
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { MapPin, Star } from '@lucide/svelte';

	const { place } = $props<{ place: Place }>();

	const name = $derived(place.displayName?.text || place.name || '');
	const address = $derived(place.formattedAddress || '');
	const rating = $derived(place.rating ?? null);
	const reviewCount = $derived(place.userRatingCount ?? null);
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<div class="mb-2">
		<span class="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">
			Preview — not saved yet
		</span>
	</div>

	<h3 class="text-xl font-bold text-gray-900">{name}</h3>

	{#if address}
		<div class="mt-2 flex items-start gap-2 text-sm text-gray-600">
			<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-gray-400" />
			<span>{address}</span>
		</div>
	{/if}

	{#if rating !== null}
		<div class="mt-2 flex items-center gap-1.5">
			{#each Array(5) as _, i}
				<Star
					class="h-4 w-4 {i < Math.round(rating)
						? 'fill-amber-400 text-amber-400'
						: 'fill-none text-gray-300'}"
				/>
			{/each}
			<span class="text-sm text-gray-600">{rating.toFixed(1)}</span>
			{#if reviewCount}
				<span class="text-xs text-gray-400">({reviewCount.toLocaleString()} Google reviews)</span>
			{/if}
		</div>
	{/if}
</div>
