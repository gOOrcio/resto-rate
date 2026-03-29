<script lang="ts">
	import clients from '$lib/client/client';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import {
		PriceLevel,
		BusinessStatus
	} from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import type { RestaurantProto } from '$lib/client/generated/restaurants/v1/restaurant_pb';
	import {
		MapPin,
		Pencil,
		Check,
		X,
		ChevronRight,
		ChevronLeft,
		Phone,
		Globe,
		Star,
		Loader2
	} from '@lucide/svelte';
	import { fly } from 'svelte/transition';

	const { restaurant, initialGoogleData = undefined } = $props<{
		restaurant: RestaurantProto;
		initialGoogleData?: Place;
	}>();

	// Local copies of editable DB fields
	let localName = $state(restaurant.name);
	let localAddress = $state(restaurant.address);

	// Inline edit state
	let isEditingName = $state(false);
	let isEditingAddress = $state(false);
	let editedName = $state('');
	let editedAddress = $state('');
	let isSaving = $state(false);
	let saveError = $state<string | null>(null);

	// Google panel state
	let isExpanded = $state(false);
	let googleData = $state<Place | null>(initialGoogleData ?? null);
	let isLoadingGoogle = $state(false);
	let googleError = $state<string | null>(null);

	const hasGoogle = $derived(!!restaurant.googlePlacesId);
	const isEditing = $derived(isEditingName || isEditingAddress);

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
					const today = new Date().getDay(); // 0=Sun
					const idx = today === 0 ? 6 : today - 1; // Google: 0=Mon
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
				]
			: []
	);

	async function toggleExpand() {
		if (!isExpanded && !googleData && restaurant.googlePlacesId) {
			await fetchGoogleData();
		}
		isExpanded = !isExpanded;
	}

	async function fetchGoogleData() {
		isLoadingGoogle = true;
		googleError = null;
		try {
			googleData = await clients.googleMaps.getRestaurantDetails({
				name: restaurant.googlePlacesId,
				languageCode: 'pl',
				regionCode: 'pl'
			});
		} catch {
			googleError = 'Failed to load Google details.';
		} finally {
			isLoadingGoogle = false;
		}
	}

	function startEditName() {
		editedName = localName;
		isEditingName = true;
		isEditingAddress = false;
		saveError = null;
	}

	function startEditAddress() {
		editedAddress = localAddress;
		isEditingAddress = true;
		isEditingName = false;
		saveError = null;
	}

	function cancelEdit() {
		isEditingName = false;
		isEditingAddress = false;
		saveError = null;
	}

	async function saveChanges() {
		isSaving = true;
		saveError = null;
		try {
			const result = await clients.restaurants.updateRestaurant({
				id: restaurant.id,
				name: editedName,
				address: editedAddress,
				googlePlacesId: restaurant.googlePlacesId
			});
			localName = result.restaurant?.name ?? editedName;
			localAddress = result.restaurant?.address ?? editedAddress;
			isEditingName = false;
			isEditingAddress = false;
		} catch {
			saveError = 'Failed to save changes.';
		} finally {
			isSaving = false;
		}
	}

	function safeHostname(uri: string): string {
		try {
			return new URL(uri).hostname;
		} catch {
			return uri;
		}
	}

	function formatCreatedAt(unix: bigint | number): string {
		return new Date(Number(unix) * 1000).toLocaleDateString();
	}
</script>

<div
	class="flex w-fit overflow-hidden rounded-2xl bg-card shadow-xl transition-all duration-300 hover:shadow-2xl"
>
	<!-- ═══ Left panel: DB data ═══ -->
	<div class="flex w-80 flex-shrink-0 flex-col gap-5 p-6">
		<!-- Name -->
		<div class="group">
			{#if isEditingName}
				<input
					type="text"
					class="w-full rounded-lg border border-ring px-2 py-1 text-xl font-bold text-foreground focus:outline-none focus:ring-2 focus:ring-ring bg-background"
					bind:value={editedName}
					onkeydown={(e) => e.key === 'Escape' && cancelEdit()}
				/>
			{:else}
				<div class="flex items-start justify-between gap-2">
					<h3 class="text-xl font-bold leading-tight text-foreground">
						{localName}
					</h3>
					<button
						onclick={startEditName}
						class="mt-0.5 shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
						aria-label="Edit name"
					>
						<Pencil class="h-4 w-4" />
					</button>
				</div>
			{/if}
		</div>

		<!-- Address -->
		<div class="group flex-1">
			{#if isEditingAddress}
				<textarea
					class="w-full resize-none rounded-lg border border-ring px-2 py-1 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring bg-background"
					rows="3"
					bind:value={editedAddress}
					onkeydown={(e) => e.key === 'Escape' && cancelEdit()}
				></textarea>
			{:else}
				<div class="flex items-start gap-2">
					<MapPin class="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
					<div class="flex flex-1 items-start justify-between gap-2">
						<p class="text-sm leading-relaxed text-muted-foreground">
							{localAddress || '—'}
						</p>
						<button
							onclick={startEditAddress}
							class="mt-0.5 shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
							aria-label="Edit address"
						>
							<Pencil class="h-4 w-4" />
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Save / Cancel controls -->
		{#if isEditing}
			<div class="flex flex-col gap-2">
				<div class="flex items-center gap-2">
					<button
						onclick={saveChanges}
						disabled={isSaving}
						class="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
					>
						{#if isSaving}
							<Loader2 class="h-4 w-4 animate-spin" />
						{:else}
							<Check class="h-4 w-4" />
						{/if}
						Save
					</button>
					<button
						onclick={cancelEdit}
						class="flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
					>
						<X class="h-4 w-4" />
						Cancel
					</button>
				</div>
				{#if saveError}
					<p class="text-xs text-destructive">{saveError}</p>
				{/if}
			</div>
		{/if}

		<!-- Footer: created date + expand toggle -->
		<div class="mt-auto flex items-center justify-between pt-2">
			<span class="text-xs text-muted-foreground">
				{formatCreatedAt(restaurant.createdAt)}
			</span>
			{#if hasGoogle}
				<button
					onclick={toggleExpand}
					class="flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm font-medium text-muted-foreground transition-all hover:border-primary/40 hover:bg-accent hover:text-foreground"
				>
					{#if isExpanded}
						<ChevronLeft class="h-4 w-4" />
						Collapse
					{:else}
						<ChevronRight class="h-4 w-4" />
						Google details
					{/if}
				</button>
			{/if}
		</div>
	</div>

	<!-- ═══ Right panel: Google data (Svelte fly transition) ═══ -->
	{#if isExpanded}
		<div transition:fly={{ x: 50, duration: 300 }}>
			<!-- separator -->
			<div class="w-px self-stretch bg-border"></div>

			<div class="flex h-full w-96 flex-col overflow-y-auto">
				{#if isLoadingGoogle}
					<div class="flex flex-1 items-center justify-center py-16">
						<Loader2 class="h-10 w-10 animate-spin text-muted-foreground" />
					</div>
				{:else if googleError}
					<div class="flex flex-1 flex-col items-center justify-center gap-3 py-16 text-center">
						<p class="text-sm text-destructive">{googleError}</p>
						<button
							onclick={fetchGoogleData}
							class="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground transition-colors hover:bg-primary/90"
						>
							Retry
						</button>
					</div>
				{:else if googleData}
					<div class="space-y-5 p-6">
						<!-- Header -->
						<div class="flex items-center justify-between">
							<img src="/GoogleMaps_Logo_Gray.svg" alt="Google Maps" class="h-4 w-auto" />
							<span class="text-xs text-muted-foreground">Live · not cached</span>
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
							{#if googleData.formattedAddress}
								<div class="flex items-start gap-2">
									<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
									<p class="text-sm leading-relaxed text-muted-foreground">
										{googleData.formattedAddress}
									</p>
								</div>
							{/if}
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
										Open in Maps
									</a>
								</div>
							{/if}
						</div>

						<!-- Opening hours: today -->
						{#if hoursToday}
							<div>
								<hr class="mb-4 border-border" />
								<h4 class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
									Today's hours
								</h4>
								<p class="text-sm text-muted-foreground">{hoursToday}</p>
							</div>
						{/if}

						<!-- Amenities -->
						{#if amenities.length > 0}
							<div>
								<hr class="mb-4 border-border" />
								<h4 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
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

						<!-- Attribution -->
						<div class="border-t border-border pt-3">
							<p class="text-xs text-muted-foreground">Data from Google Places API</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
