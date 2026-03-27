<script lang="ts">
	import clients from '$lib/client/client';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import {
		PriceLevel,
		BusinessStatus
	} from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import {
		MapPin,
		Phone,
		Globe,
		Star,
		Loader2,
		Check,
		X,
		ChevronRight,
		ChevronUp
	} from '@lucide/svelte';

	const { googlePlacesId, name, address, city, country } = $props<{
		googlePlacesId: string;
		name: string;
		address?: string;
		city?: string;
		country?: string;
	}>();

	let isExpanded = $state(false);
	let googleData = $state<Place | null>(null);
	let isLoadingGoogle = $state(false);
	let googleError = $state<string | null>(null);

	let status = $derived(
		googleData
			? (() => {
					switch (googleData.businessStatus) {
						case BusinessStatus.OPERATIONAL:
							return { label: 'Operational', color: 'text-green-600' };
						case BusinessStatus.CLOSED_TEMPORARILY:
							return { label: 'Temporarily closed', color: 'text-yellow-600' };
						case BusinessStatus.CLOSED_PERMANENTLY:
							return { label: 'Permanently closed', color: 'text-red-600' };
						default:
							return null;
					}
				})()
			: null
	);

	let priceLabel = $derived(
		googleData
			? (() => {
					const map: Partial<Record<PriceLevel, string>> = {
						[PriceLevel.FREE]: 'Free',
						[PriceLevel.INEXPENSIVE]: '$',
						[PriceLevel.MODERATE]: '$$',
						[PriceLevel.EXPENSIVE]: '$$$',
						[PriceLevel.VERY_EXPENSIVE]: '$$$$'
					};
					return map[googleData.priceLevel] ?? '';
				})()
			: ''
	);

	let hoursToday = $derived(
		googleData?.regularOpeningHours
			? (() => {
					const today = new Date().getDay();
					const idx = today === 0 ? 6 : today - 1;
					return googleData.regularOpeningHours!.weekdayText[idx] ?? null;
				})()
			: null
	);

	let amenities = $derived(
		googleData
			? [
					{ label: 'Dine-in', value: googleData.dineIn },
					{ label: 'Takeout', value: googleData.takeout },
					{ label: 'Delivery', value: googleData.delivery },
					{ label: 'Outdoor seating', value: googleData.outdoorSeating },
					{ label: 'Reservations', value: googleData.reservable }
				].filter((a) => a.value !== undefined && a.value !== null)
			: []
	);

	async function fetchGoogleData() {
		isLoadingGoogle = true;
		googleError = null;
		try {
			googleData = await clients.googleMaps.getRestaurantDetails({
				name: googlePlacesId,
				languageCode: 'pl',
				regionCode: 'pl'
			});
		} catch {
			googleError = 'Failed to load Google details.';
		} finally {
			isLoadingGoogle = false;
		}
	}

	async function toggleExpand() {
		if (!isExpanded && !googleData) {
			await fetchGoogleData();
		}
		isExpanded = !isExpanded;
	}

	function safeHostname(uri: string): string {
		try {
			return new URL(uri).hostname;
		} catch {
			return uri;
		}
	}
</script>

<div class="flex flex-col">
	<!-- DB info section -->
	<div class="space-y-1">
		<h3 class="text-base leading-tight font-bold text-gray-900">{name}</h3>
		{#if address}
			<p class="text-sm text-gray-500">{address}</p>
		{/if}
		{#if city || country}
			<p class="text-xs text-gray-400">
				{[city, country].filter(Boolean).join(', ')}
			</p>
		{/if}
	</div>

	<!-- Toggle button -->
	<div class="mt-3">
		<button
			onclick={toggleExpand}
			class="flex items-center gap-1.5 text-sm font-medium text-blue-600 transition-colors hover:text-blue-800"
		>
			{#if isExpanded}
				Hide details
				<ChevronUp class="h-4 w-4" />
			{:else}
				Show Google details
				<ChevronRight class="h-4 w-4" />
			{/if}
		</button>
	</div>

	<!-- Collapsible Google data section -->
	{#if isExpanded}
		<hr class="mt-3 border-gray-200" />

		{#if isLoadingGoogle}
			<div class="flex items-center justify-center py-8">
				<Loader2 class="h-8 w-8 animate-spin text-gray-400" />
			</div>
		{:else if googleError}
			<div class="flex flex-col items-center gap-3 py-6 text-center">
				<p class="text-sm text-red-500">{googleError}</p>
				<button
					onclick={fetchGoogleData}
					class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white transition-colors hover:bg-blue-700"
				>
					Retry
				</button>
			</div>
		{:else if googleData}
			<div class="mt-4 space-y-5">
				<!-- Header -->
				<div class="flex items-center justify-between">
					<img src="/GoogleMaps_Logo_Gray.svg" alt="Google Maps" class="h-4 w-auto" />
					<span class="text-xs text-gray-400">Live · not cached</span>
				</div>

				<!-- Core info: rating, status, price -->
				<div class="space-y-2">
					{#if googleData.rating}
						<div class="flex items-center gap-3">
							<div class="flex items-center gap-0.5">
								{#each Array(5) as _, i}
									<Star
										class="h-4 w-4 {i < Math.round(googleData.rating)
											? 'fill-amber-400 text-amber-400'
											: 'fill-none text-gray-300'}"
									/>
								{/each}
								<span class="ml-1 text-sm font-semibold text-gray-800">
									{googleData.rating.toFixed(1)}
								</span>
							</div>
							{#if googleData.userRatingCount}
								<span class="text-xs text-gray-400">
									({googleData.userRatingCount.toLocaleString()} reviews)
								</span>
							{/if}
						</div>
					{/if}

					<div class="flex items-center gap-2">
						{#if status}
							<span class="text-sm font-medium {status.color}">{status.label}</span>
						{/if}
						{#if priceLabel}
							<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-semibold text-gray-700">
								{priceLabel}
							</span>
						{/if}
					</div>
				</div>

				<hr class="border-gray-200" />

				<!-- Contact & location -->
				<div class="space-y-3">
					{#if googleData.formattedAddress}
						<div class="flex items-start gap-2">
							<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-gray-400" />
							<p class="text-sm leading-relaxed text-gray-600">
								{googleData.formattedAddress}
							</p>
						</div>
					{/if}
					{#if googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
						<div class="flex items-center gap-2">
							<Phone class="h-4 w-4 shrink-0 text-gray-400" />
							<a
								href="tel:{googleData.internationalPhoneNumber || googleData.nationalPhoneNumber}"
								class="text-sm text-blue-600 hover:underline"
							>
								{googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
							</a>
						</div>
					{/if}
					{#if googleData.websiteUri}
						<div class="flex items-center gap-2 overflow-hidden">
							<Globe class="h-4 w-4 shrink-0 text-gray-400" />
							<a
								href={googleData.websiteUri}
								target="_blank"
								rel="noopener noreferrer"
								class="truncate text-sm text-blue-600 hover:underline"
							>
								{safeHostname(googleData.websiteUri)}
							</a>
						</div>
					{/if}
					{#if googleData.googleMapsUri}
						<div class="flex items-center gap-2">
							<img src="/GoogleMaps_Logo_Gray.svg" alt="" class="h-3.5 w-auto shrink-0" />
							<a
								href={googleData.googleMapsUri}
								target="_blank"
								rel="noopener noreferrer"
								class="text-sm text-blue-600 hover:underline"
							>
								Open in Maps
							</a>
						</div>
					{/if}
				</div>

				<!-- Opening hours: today -->
				{#if hoursToday}
					<div>
						<hr class="mb-4 border-gray-200" />
						<h4 class="mb-1.5 text-xs font-semibold tracking-wide text-gray-400 uppercase">
							Today's hours
						</h4>
						<p class="text-sm text-gray-600">{hoursToday}</p>
					</div>
				{/if}

				<!-- Amenities -->
				{#if amenities.length > 0}
					<div>
						<hr class="mb-4 border-gray-200" />
						<h4 class="mb-3 text-xs font-semibold tracking-wide text-gray-400 uppercase">
							Features
						</h4>
						<div class="grid grid-cols-2 gap-2">
							{#each amenities as feature}
								<div class="flex items-center gap-1.5">
									{#if feature.value}
										<Check class="h-3.5 w-3.5 shrink-0 text-green-500" />
									{:else}
										<X class="h-3.5 w-3.5 shrink-0 text-gray-300" />
									{/if}
									<span class="text-xs text-gray-600">{feature.label}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Attribution -->
				<div class="border-t border-gray-100 pt-3">
					<p class="text-xs text-gray-400">Data from Google Places API</p>
				</div>
			</div>
		{/if}
	{/if}
</div>
