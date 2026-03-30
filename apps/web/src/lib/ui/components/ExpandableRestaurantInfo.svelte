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
	import * as m from '$lib/paraglide/messages';

	const { googlePlacesId, name, address, city, country, photoReference = '', rating = undefined } = $props<{
		googlePlacesId: string;
		name: string;
		address?: string;
		city?: string;
		country?: string;
		photoReference?: string;
		rating?: number;
	}>();

	function ratingColor(r: number): string {
		if (r >= 4.5) return 'text-emerald-500 dark:text-emerald-400';
		if (r >= 3.5) return 'text-amber-500 dark:text-amber-400';
		if (r >= 2.5) return 'text-orange-500 dark:text-orange-400';
		return 'text-red-500 dark:text-red-400';
	}

	let isExpanded = $state(false);
	let googleData = $state<Place | null>(null);
	let isLoadingGoogle = $state(false);
	let googleError = $state<string | null>(null);
	let photoLoadFailed = $state(false);

	let status = $derived(
		googleData
			? (() => {
					switch (googleData.businessStatus) {
						case BusinessStatus.OPERATIONAL:
							return { label: 'Operational', color: 'text-green-600 dark:text-green-400' };
						case BusinessStatus.CLOSED_TEMPORARILY:
							return { label: 'Temporarily closed', color: 'text-yellow-600 dark:text-yellow-400' };
						case BusinessStatus.CLOSED_PERMANENTLY:
							return { label: 'Permanently closed', color: 'text-red-600 dark:text-red-400' };
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
			googleError = m.expandable_load_failed();
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
	<div class="relative mb-3 h-36 w-full overflow-hidden rounded-lg bg-muted">
		{#if photoReference && !photoLoadFailed}
			<img
				src="{import.meta.env.VITE_API_URL || 'http://localhost:3001'}/place-photo?name={encodeURIComponent(photoReference)}"
				alt="Restaurant cover"
				class="h-full w-full object-cover"
				onerror={() => { photoLoadFailed = true; }}
			/>
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-10 w-10 text-muted-foreground/30"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="1"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
					/>
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
				</svg>
			</div>
		{/if}
	</div>

	<!-- DB info section -->
	<div class="flex items-start gap-3">
		<div class="min-w-0 flex-1 space-y-1">
			<h3 class="text-base leading-tight font-bold text-foreground">{name}</h3>
			{#if address}
				<div class="flex items-start gap-1.5">
					<MapPin class="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
					<p class="text-sm text-muted-foreground">{address}</p>
				</div>
			{/if}
		</div>
		{#if rating !== undefined}
			<div class="flex shrink-0 items-center gap-1 rounded-md border border-current px-2 py-1 {ratingColor(rating)}">
				<span class="text-lg font-bold tabular-nums leading-none">{rating.toFixed(1)}</span>
				<Star class="h-4 w-4 fill-current" />
			</div>
		{/if}
	</div>

	<!-- Toggle button -->
	<div class="mt-3">
		<button
			onclick={toggleExpand}
			class="flex items-center gap-1.5 text-sm font-medium text-primary transition-colors hover:text-primary/80"
		>
			{#if isExpanded}
				{m.expandable_hide_details()}
				<ChevronUp class="h-4 w-4" />
			{:else}
				{m.expandable_show_google()}
				<ChevronRight class="h-4 w-4" />
			{/if}
		</button>
	</div>

	<!-- Collapsible Google data section -->
	{#if isExpanded}
		<div class="mt-3 rounded-lg border border-border bg-muted/30 p-4">
			{#if isLoadingGoogle}
				<div class="flex items-center justify-center py-8">
					<Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
				</div>
			{:else if googleError}
				<div class="flex flex-col items-center gap-3 py-6 text-center">
					<p class="text-sm text-destructive">{googleError}</p>
					<button
						onclick={fetchGoogleData}
						class="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground transition-colors hover:bg-primary/90"
					>
						{m.common_retry()}
					</button>
				</div>
			{:else if googleData}
				<div class="space-y-4">
					<!-- Header -->
					<div class="flex items-center justify-between">
						<img src="/GoogleMaps_Logo_Gray.svg" alt="Google Maps" class="h-4 w-auto" />
						<span class="text-xs text-muted-foreground">{m.expandable_google_source()}</span>
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
												: 'fill-none text-gray-300 dark:text-gray-600'}"
										/>
									{/each}
									<span class="ml-1 text-sm font-semibold text-foreground">
										{googleData.rating.toFixed(1)}
									</span>
								</div>
								{#if googleData.userRatingCount}
									<span class="text-xs text-muted-foreground">
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
								<span class="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold text-muted-foreground">
									{priceLabel}
								</span>
							{/if}
						</div>
					</div>

					<hr class="border-border" />

					<!-- Contact & location -->
					<div class="space-y-3">
						{#if googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
							<div class="flex items-center gap-2">
								<Phone class="h-4 w-4 shrink-0 text-muted-foreground" />
								<a
									href="tel:{googleData.internationalPhoneNumber || googleData.nationalPhoneNumber}"
									class="text-sm text-primary hover:underline"
								>
									{googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
								</a>
							</div>
						{/if}
						{#if googleData.websiteUri}
							<div class="flex items-center gap-2 overflow-hidden">
								<Globe class="h-4 w-4 shrink-0 text-muted-foreground" />
								<a
									href={googleData.websiteUri}
									target="_blank"
									rel="noopener noreferrer"
									class="truncate text-sm text-primary hover:underline"
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
									class="text-sm text-primary hover:underline"
								>
									{m.expandable_open_maps()}
								</a>
							</div>
						{/if}
					</div>

					<!-- Opening hours: today -->
					{#if hoursToday}
						<div>
							<hr class="border-border" />
							<h4 class="mb-1.5 mt-4 text-xs font-semibold tracking-wide text-muted-foreground uppercase">
								Today's hours
							</h4>
							<p class="text-sm text-muted-foreground">{hoursToday}</p>
						</div>
					{/if}

					<!-- Amenities -->
					{#if amenities.length > 0}
						<div>
							<hr class="border-border" />
							<h4 class="mb-3 mt-4 text-xs font-semibold tracking-wide text-muted-foreground uppercase">
								Features
							</h4>
							<div class="grid grid-cols-2 gap-2">
								{#each amenities as feature}
									<div class="flex items-center gap-1.5">
										{#if feature.value}
											<Check class="h-3.5 w-3.5 shrink-0 text-green-500 dark:text-green-400" />
										{:else}
											<X class="h-3.5 w-3.5 shrink-0 text-muted-foreground/40" />
										{/if}
										<span class="text-xs text-muted-foreground">{feature.label}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>
