<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import {
		PriceLevel,
		BusinessStatus
	} from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { PartySize, Occasion, WouldVisitAgain } from '$lib/client/generated/reviews/v1/review_pb';
	import {
		Star,
		MapPin,
		Phone,
		Globe,
		Check,
		X,
		Loader2,
		ChevronLeft,
		ChevronDown,
		ChevronUp
	} from '@lucide/svelte';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';

	const googlePlacesId = $derived(decodeURIComponent(page.params.googlePlacesId ?? ''));

	let reviews = $state<ReviewProto[]>([]);
	let averageRating = $state(0);
	let restaurantName = $state('');
	let restaurantAddress = $state('');
	let restaurantCity = $state('');
	let restaurantCountry = $state('');
	let restaurantPhoto = $state('');
	let loading = $state(true);
	let error = $state('');
	let isWishlisted = $state(false);
	let wishlistLoading = $state(false);
	let showRatingForm = $state(false);
	let myReviewExpanded = $state(false);
	let expandedFriendIds = $state<Set<string>>(new Set());
	let photoLoadFailed = $state(false);

	// Google Places data
	let googleData = $state<Place | null>(null);
	let googleLoading = $state(false);

	const myReview = $derived(reviews.find((r) => r.userId === auth.user?.id));
	const friendReviews = $derived(reviews.filter((r) => r.userId !== auth.user?.id));

	const PARTY_SIZE_LABELS: Record<number, string> = {
		[PartySize.SOLO]: 'Solo',
		[PartySize.COUPLE]: 'Couple',
		[PartySize.SMALL_GROUP]: 'Small group',
		[PartySize.LARGE_GROUP]: 'Large group'
	};
	const OCCASION_LABELS: Record<number, string> = {
		[Occasion.CASUAL]: 'Casual',
		[Occasion.DATE_NIGHT]: 'Date night',
		[Occasion.BUSINESS]: 'Business',
		[Occasion.CELEBRATION]: 'Celebration',
		[Occasion.QUICK_BITE]: 'Quick bite'
	};
	const WOULD_VISIT_AGAIN_LABELS: Record<number, { text: string; cls: string }> = {
		[WouldVisitAgain.YES]: { text: 'Would visit again', cls: 'text-emerald-600 dark:text-emerald-400' },
		[WouldVisitAgain.MAYBE]: { text: 'Maybe again', cls: 'text-amber-600 dark:text-amber-400' },
		[WouldVisitAgain.NO]: { text: "Wouldn't return", cls: 'text-red-600 dark:text-red-400' }
	};

	function formatDate(ts: bigint | number): string {
		if (!ts) return '';
		return new Date(Number(ts) * 1000).toLocaleDateString(undefined, {
			year: 'numeric', month: 'short', day: 'numeric'
		});
	}

	function toggleFriend(id: string) {
		const next = new Set(expandedFriendIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		expandedFriendIds = next;
	}

	async function loadRestaurantData() {
		const [reviewsResult, wishlistResult] = await Promise.allSettled([
			client.reviews.listRestaurantReviews({ googlePlacesId }),
			client.wishlist.listWishlist({ googlePlacesId })
		]);

		if (reviewsResult.status === 'rejected') throw reviewsResult.reason;

		const res = reviewsResult.value;
		reviews = res.reviews;
		averageRating = res.averageRating;
		restaurantName = res.restaurantName;
		restaurantAddress = res.restaurantAddress;
		restaurantCity = res.restaurantCity;
		restaurantCountry = res.restaurantCountry;
		restaurantPhoto = res.reviews.find((r) => r.restaurantPhotoReference)?.restaurantPhotoReference || '';

		if (wishlistResult.status === 'fulfilled') {
			isWishlisted = (wishlistResult.value.items?.length ?? 0) > 0;
		}
	}

	async function loadGoogleData() {
		googleLoading = true;
		try {
			googleData = await client.googleMaps.getRestaurantDetails({
				name: googlePlacesId,
				languageCode: 'en',
				regionCode: 'pl'
			});
			if (!restaurantPhoto && googleData.photos?.[0]?.name) {
				restaurantPhoto = googleData.photos[0].name;
			}
		} catch (e) {
			console.error('Failed to load Google data:', e);
		} finally {
			googleLoading = false;
		}
	}

	async function toggleWishlist() {
		wishlistLoading = true;
		try {
			if (isWishlisted) {
				await client.wishlist.removeFromWishlist({ googlePlacesId });
				isWishlisted = false;
			} else {
				await client.wishlist.addToWishlist({
					googlePlacesId,
					restaurantName,
					restaurantAddress,
					city: restaurantCity,
					country: restaurantCountry,
					photoReference: restaurantPhoto
				});
				isWishlisted = true;
			}
		} catch (e) {
			console.error('Wishlist toggle error:', e);
		} finally {
			wishlistLoading = false;
		}
	}

	function handleReviewSubmit(review: ReviewProto) {
		reviews = [review, ...reviews.filter((r) => r.userId !== auth.user?.id)];
		averageRating = reviews.reduce((s, r) => s + r.rating, 0) / (reviews.length || 1);
		showRatingForm = false;
		myReviewExpanded = true;
	}

	function safeHostname(uri: string): string {
		try { return new URL(uri).hostname; } catch { return uri; }
	}

	let status = $derived(
		googleData
			? (() => {
					switch (googleData.businessStatus) {
						case BusinessStatus.OPERATIONAL:
							return { label: 'Open', color: 'text-emerald-600 dark:text-emerald-400' };
						case BusinessStatus.CLOSED_TEMPORARILY:
							return { label: 'Temporarily closed', color: 'text-amber-600 dark:text-amber-400' };
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

	let initialized = $state(false);
	$effect(() => {
		if (auth.loading || initialized) return;
		initialized = true;
		if (!auth.isLoggedIn) { goto('/'); return; }
		loading = true;
		Promise.all([loadRestaurantData(), loadGoogleData()])
			.catch((e) => { console.error(e); error = 'Failed to load restaurant data.'; })
			.finally(() => { loading = false; });
	});
</script>

<div class="mx-auto max-w-3xl space-y-5 px-4 py-8 sm:px-6">
	<!-- Back -->
	<button
		onclick={() => history.back()}
		class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
	>
		<ChevronLeft class="h-4 w-4" />
		Back
	</button>

	{#if loading}
		<div class="flex items-center gap-2 py-16 text-sm text-muted-foreground">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
			Loading…
		</div>
	{:else if error}
		<p class="text-sm text-destructive">{error}</p>
	{:else}
		<!-- ── Hero card ── -->
		<div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
			<!-- Photo -->
			<div class="relative h-52 w-full bg-muted">
				{#if restaurantPhoto && !photoLoadFailed}
					<img
						src="{import.meta.env.VITE_API_URL || 'http://localhost:3001'}/place-photo?name={encodeURIComponent(restaurantPhoto)}"
						alt="Restaurant cover"
						class="h-full w-full object-cover"
						onerror={() => { photoLoadFailed = true; }}
					/>
				{:else}
					<div class="flex h-full w-full items-center justify-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-muted-foreground/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
					</div>
				{/if}
			</div>

			<div class="p-5">
				<!-- Name + address + wishlist -->
				<div class="flex items-start justify-between gap-4">
					<div class="min-w-0 flex-1">
						<h1 class="text-2xl font-bold text-foreground">{restaurantName || 'Restaurant'}</h1>
						{#if restaurantAddress}
							<div class="mt-1 flex items-start gap-1.5">
								<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
								<p class="text-sm text-muted-foreground">{restaurantAddress}</p>
							</div>
						{/if}
						{#if restaurantCity || restaurantCountry}
							<p class="mt-0.5 pl-5 text-xs text-muted-foreground">
								{[restaurantCity, restaurantCountry].filter(Boolean).join(', ')}
							</p>
						{/if}
					</div>
					{#if !myReview}
						<button
							onclick={toggleWishlist}
							disabled={wishlistLoading}
							class="flex shrink-0 items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50
								{isWishlisted
									? 'border-amber-300 bg-amber-50 text-amber-700 hover:bg-amber-100 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-300'
									: 'border-border bg-card text-foreground hover:bg-muted'}"
						>
							{#if wishlistLoading}
								<Loader2 class="h-3.5 w-3.5 animate-spin" />
							{:else}
								<Star class="h-3.5 w-3.5 {isWishlisted ? 'fill-amber-500 text-amber-500' : ''}" />
							{/if}
							{isWishlisted ? 'Wishlisted' : 'Wishlist'}
						</button>
					{/if}
				</div>

				<!-- Rating row — collapses into your review, or shows write CTA -->
				<div class="mt-4 border-t border-border pt-4">
					{#if showRatingForm}
						<div class="space-y-3">
							<RatingForm
								{googlePlacesId}
								restaurantName={restaurantName}
								restaurantAddress={restaurantAddress}
								existingReview={myReview}
								onSubmit={handleReviewSubmit}
							/>
							<button
								class="text-sm text-muted-foreground hover:text-foreground"
								onclick={() => (showRatingForm = false)}
							>
								Cancel
							</button>
						</div>
					{:else if myReview}
						<!-- Collapsible: summary row + expandable details -->
						<button
							class="flex w-full items-center justify-between gap-3 text-left"
							onclick={() => (myReviewExpanded = !myReviewExpanded)}
						>
							<div class="flex items-center gap-2">
								<div class="flex items-center gap-0.5">
									{#each Array(5) as _, i}
										<Star class="h-5 w-5 {i < Math.round(averageRating) ? 'fill-amber-400 text-amber-400' : 'fill-none text-gray-300 dark:text-gray-600'}" />
									{/each}
								</div>
								<span class="font-semibold text-foreground">{averageRating.toFixed(1)}</span>
								<span class="text-sm text-muted-foreground">
									({reviews.length} {reviews.length === 1 ? 'review' : 'reviews'} from you &amp; friends)
								</span>
							</div>
							{#if myReviewExpanded}
								<ChevronUp class="h-4 w-4 shrink-0 text-muted-foreground" />
							{:else}
								<ChevronDown class="h-4 w-4 shrink-0 text-muted-foreground" />
							{/if}
						</button>

						{#if myReviewExpanded}
							<div class="mt-4 space-y-3 rounded-lg border border-border bg-muted/30 p-4">
								<!-- Header: your rating + date + edit -->
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2">
										<div class="flex items-center gap-0.5">
											{#each Array(5) as _, i}
												<Star class="h-4 w-4 {i < myReview.rating ? 'fill-amber-400 text-amber-400' : 'fill-none text-gray-300 dark:text-gray-600'}" />
											{/each}
										</div>
										<span class="font-semibold text-foreground">{myReview.rating.toFixed(1)}</span>
									</div>
									<div class="flex items-center gap-3">
										<span class="text-xs text-muted-foreground">{formatDate(myReview.createdAt)}</span>
										<button
											class="text-xs text-primary hover:underline"
											onclick={(e) => { e.stopPropagation(); showRatingForm = true; myReviewExpanded = false; }}
										>
											Edit
										</button>
									</div>
								</div>

								{#if myReview.comment}
									<p class="text-sm leading-relaxed text-muted-foreground">{myReview.comment}</p>
								{/if}

								{#if myReview.tags?.length}
									<div class="flex flex-wrap gap-1.5">
										{#each myReview.tags as tag}
											<span class="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">{tag}</span>
										{/each}
									</div>
								{/if}

								<!-- Extra fields -->
								{#if myReview.visitedAt || myReview.partySize || myReview.occasion || myReview.pricePaidPerPerson || myReview.wouldVisitAgain || myReview.dishHighlights}
									{@const visitDate = formatDate(myReview.visitedAt)}
									{@const partyLabel = PARTY_SIZE_LABELS[myReview.partySize]}
									{@const occasionLabel = OCCASION_LABELS[myReview.occasion]}
									{@const wvaEntry = WOULD_VISIT_AGAIN_LABELS[myReview.wouldVisitAgain]}
									<div class="border-t border-border pt-3 space-y-1.5">
										<div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
											{#if visitDate}<span>📅 {visitDate}</span>{/if}
											{#if partyLabel}<span>👥 {partyLabel}</span>{/if}
											{#if occasionLabel}<span>🎉 {occasionLabel}</span>{/if}
											{#if myReview.pricePaidPerPerson}<span>💰 ${myReview.pricePaidPerPerson}/person</span>{/if}
											{#if wvaEntry}<span class={wvaEntry.cls}>{wvaEntry.text}</span>{/if}
										</div>
										{#if myReview.dishHighlights}
											<p class="text-xs text-muted-foreground">
												<span class="font-medium text-foreground">Highlights:</span> {myReview.dishHighlights}
											</p>
										{/if}
									</div>
								{/if}
							</div>
						{/if}
					{:else}
						<!-- No review yet -->
						<div class="flex items-center justify-between">
							<span class="text-sm text-muted-foreground">
								{reviews.length > 0
									? `${reviews.length} ${reviews.length === 1 ? 'review' : 'reviews'} from friends`
									: 'No reviews yet'}
							</span>
							<button
								class="rounded-lg bg-primary px-4 py-1.5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
								onclick={() => (showRatingForm = true)}
							>
								Write a review
							</button>
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- ── Google Places ── -->
		<div class="rounded-xl border border-border bg-card p-5 shadow-sm">
			<div class="mb-4 flex items-center justify-between">
				<img src="/GoogleMaps_Logo_Gray.svg" alt="Google Maps" class="h-4 w-auto" />
				<span class="text-xs text-muted-foreground">Google Places data</span>
			</div>

			{#if googleLoading}
				<div class="flex items-center gap-2 py-6 text-sm text-muted-foreground">
					<Loader2 class="h-4 w-4 animate-spin" />
					Loading Google details…
				</div>
			{:else if googleData}
				<div class="space-y-4">
					<div class="space-y-2">
						{#if googleData.rating}
							<div class="flex items-center gap-3">
								<div class="flex items-center gap-0.5">
									{#each Array(5) as _, i}
										<Star class="h-4 w-4 {i < Math.round(googleData.rating) ? 'fill-amber-400 text-amber-400' : 'fill-none text-gray-300 dark:text-gray-600'}" />
									{/each}
								</div>
								<span class="font-semibold text-foreground">{googleData.rating.toFixed(1)}</span>
								{#if googleData.userRatingCount}
									<span class="text-sm text-muted-foreground">({googleData.userRatingCount.toLocaleString()} Google reviews)</span>
								{/if}
							</div>
						{/if}
						<div class="flex items-center gap-2">
							{#if status}<span class="text-sm font-medium {status.color}">{status.label}</span>{/if}
							{#if priceLabel}<span class="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold text-muted-foreground">{priceLabel}</span>{/if}
						</div>
					</div>

					<hr class="border-border" />

					<div class="space-y-2">
						{#if googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
							<div class="flex items-center gap-2">
								<Phone class="h-4 w-4 shrink-0 text-muted-foreground" />
								<a href="tel:{googleData.internationalPhoneNumber || googleData.nationalPhoneNumber}" class="text-sm text-primary hover:underline">
									{googleData.nationalPhoneNumber || googleData.internationalPhoneNumber}
								</a>
							</div>
						{/if}
						{#if googleData.websiteUri}
							<div class="flex items-center gap-2 overflow-hidden">
								<Globe class="h-4 w-4 shrink-0 text-muted-foreground" />
								<a href={googleData.websiteUri} target="_blank" rel="noopener noreferrer" class="truncate text-sm text-primary hover:underline">
									{safeHostname(googleData.websiteUri)}
								</a>
							</div>
						{/if}
						{#if googleData.googleMapsUri}
							<div class="flex items-center gap-2">
								<img src="/GoogleMaps_Logo_Gray.svg" alt="" class="h-3.5 w-auto shrink-0" />
								<a href={googleData.googleMapsUri} target="_blank" rel="noopener noreferrer" class="text-sm text-primary hover:underline">
									Open in Google Maps
								</a>
							</div>
						{/if}
					</div>

					{#if hoursToday}
						<hr class="border-border" />
						<div>
							<h4 class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Today's hours</h4>
							<p class="text-sm text-muted-foreground">{hoursToday}</p>
						</div>
					{/if}

					{#if amenities.length > 0}
						<hr class="border-border" />
						<div>
							<h4 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Features</h4>
							<div class="grid grid-cols-2 gap-2">
								{#each amenities as feature}
									<div class="flex items-center gap-1.5">
										{#if feature.value}
											<Check class="h-3.5 w-3.5 shrink-0 text-emerald-500" />
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
			{:else}
				<p class="text-sm text-muted-foreground">Google details unavailable.</p>
			{/if}
		</div>

		<!-- ── Friends' reviews ── -->
		{#if friendReviews.length > 0}
			<div class="rounded-xl border border-border bg-card shadow-sm">
				<div class="border-b border-border px-5 py-3">
					<h2 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
						Friends' reviews ({friendReviews.length})
					</h2>
				</div>
				<ul class="divide-y divide-border">
					{#each friendReviews as review (review.id)}
						{@const expanded = expandedFriendIds.has(review.id)}
						{@const hasDetails = !!(review.comment || review.tags?.length || review.visitedAt || review.partySize || review.occasion || review.pricePaidPerPerson || review.wouldVisitAgain || review.dishHighlights)}
						<li>
							<button
								class="flex w-full items-center justify-between gap-3 px-5 py-4 text-left {hasDetails ? 'hover:bg-muted/40' : 'cursor-default'}"
								onclick={() => hasDetails && toggleFriend(review.id)}
								disabled={!hasDetails}
							>
								<div class="flex items-center gap-3">
									<div class="flex flex-col">
										<span class="font-medium text-foreground">{review.authorName || 'Friend'}</span>
										<div class="mt-0.5 flex items-center gap-1.5">
											<div class="flex items-center gap-0.5">
												{#each Array(5) as _, i}
													<Star class="h-3.5 w-3.5 {i < review.rating ? 'fill-amber-400 text-amber-400' : 'fill-none text-gray-300 dark:text-gray-600'}" />
												{/each}
											</div>
											<span class="text-sm font-semibold text-foreground">{review.rating.toFixed(1)}</span>
										</div>
									</div>
								</div>
								<div class="flex items-center gap-2 shrink-0">
									<span class="text-xs text-muted-foreground">{formatDate(review.createdAt)}</span>
									{#if hasDetails}
										{#if expanded}
											<ChevronUp class="h-4 w-4 text-muted-foreground" />
										{:else}
											<ChevronDown class="h-4 w-4 text-muted-foreground" />
										{/if}
									{/if}
								</div>
							</button>

							{#if expanded && hasDetails}
								<div class="space-y-3 border-t border-border bg-muted/20 px-5 py-4">
									{#if review.comment}
										<p class="text-sm leading-relaxed text-muted-foreground">{review.comment}</p>
									{/if}
									{#if review.tags?.length}
										<div class="flex flex-wrap gap-1.5">
											{#each review.tags as tag}
												<span class="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">{tag}</span>
											{/each}
										</div>
									{/if}
									{#if review.visitedAt || review.partySize || review.occasion || review.pricePaidPerPerson || review.wouldVisitAgain || review.dishHighlights}
										{@const visitDate = formatDate(review.visitedAt)}
										{@const partyLabel = PARTY_SIZE_LABELS[review.partySize]}
										{@const occasionLabel = OCCASION_LABELS[review.occasion]}
										{@const wvaEntry = WOULD_VISIT_AGAIN_LABELS[review.wouldVisitAgain]}
										<div class="space-y-1.5 border-t border-border pt-3">
											<div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
												{#if visitDate}<span>📅 {visitDate}</span>{/if}
												{#if partyLabel}<span>👥 {partyLabel}</span>{/if}
												{#if occasionLabel}<span>🎉 {occasionLabel}</span>{/if}
												{#if review.pricePaidPerPerson}<span>💰 ${review.pricePaidPerPerson}/person</span>{/if}
												{#if wvaEntry}<span class={wvaEntry.cls}>{wvaEntry.text}</span>{/if}
											</div>
											{#if review.dishHighlights}
												<p class="text-xs text-muted-foreground">
													<span class="font-medium text-foreground">Highlights:</span> {review.dishHighlights}
												</p>
											{/if}
										</div>
									{/if}
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	{/if}
</div>
