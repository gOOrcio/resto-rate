<script lang="ts">
	import clients from '$lib/client/client';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import PlacePreviewCard from './PlacePreviewCard.svelte';
	import RatingForm from './RatingForm.svelte';
	import ReviewSummary from './ReviewSummary.svelte';
	import RestaurantSearch from './RestaurantSearch.svelte';

	let selectedPlace = $state<Place | null>(null);
	let isChecking = $state(false);
	let currentReview = $state<ReviewProto | null>(null);
	let currentWishlistItem = $state<WishlistItemProto | null>(null);
	let isEditingReview = $state(false);
	let showRatingForm = $state(false);
	let wishlistLoading = $state(false);

	let selectionToken = 0;

	async function handleSelect(place: Place) {
		const googlePlacesId = place.name || '';
		if (!googlePlacesId) return;

		selectedPlace = place;
		currentReview = null;
		currentWishlistItem = null;
		showRatingForm = false;
		isEditingReview = false;

		const token = ++selectionToken;
		isChecking = true;
		try {
			const [reviewResult, wishlistResult] = await Promise.allSettled([
				clients.reviews.listReviews({ googlePlacesId }),
				clients.wishlist.listWishlist({ googlePlacesId })
			]);
			if (token !== selectionToken) return;
			currentReview =
				reviewResult.status === 'fulfilled' ? (reviewResult.value.reviews?.[0] ?? null) : null;
			currentWishlistItem =
				wishlistResult.status === 'fulfilled' ? (wishlistResult.value.items?.[0] ?? null) : null;
		} finally {
			if (token === selectionToken) isChecking = false;
		}
	}

	async function addToWishlist() {
		if (!selectedPlace?.name) return;
		wishlistLoading = true;
		try {
			const res = await clients.wishlist.addToWishlist({
				googlePlacesId: selectedPlace.name,
				restaurantName: selectedPlace.displayName?.text || selectedPlace.name,
				restaurantAddress: selectedPlace.formattedAddress || '',
				city: selectedPlace.postalAddress?.locality ?? '',
				country: selectedPlace.postalAddress?.country ?? ''
			});
			currentWishlistItem = res.item ?? null;
		} catch (e) {
			console.error('Wishlist add error:', e);
		} finally {
			wishlistLoading = false;
		}
	}

	async function removeFromWishlist() {
		if (!selectedPlace?.name) return;
		wishlistLoading = true;
		try {
			await clients.wishlist.removeFromWishlist({ googlePlacesId: selectedPlace.name });
			currentWishlistItem = null;
		} catch (e) {
			console.error('Wishlist remove error:', e);
		} finally {
			wishlistLoading = false;
		}
	}
</script>

<div class="space-y-6">
	<RestaurantSearch onSelect={handleSelect} />

	{#if selectedPlace}
		<div class="mt-6 space-y-4">
			<PlacePreviewCard place={selectedPlace} />

			{#if selectedPlace.name}
				<a
					href="/restaurants/{encodeURIComponent(selectedPlace.name)}"
					class="text-sm text-primary hover:underline"
				>
					See reviews from you &amp; friends →
				</a>
			{/if}

			{#if isChecking}
				<div class="flex items-center gap-2 text-sm text-muted-foreground">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
					Checking your history…
				</div>
			{:else if currentReview && !isEditingReview}
				<ReviewSummary
					review={currentReview}
					onEdit={() => (isEditingReview = true)}
				/>
			{:else if currentReview && isEditingReview}
				<RatingForm
					googlePlacesId={selectedPlace.name || ''}
					restaurantName={selectedPlace.displayName?.text || selectedPlace.name || ''}
					restaurantAddress={selectedPlace.formattedAddress || ''}
					city={selectedPlace.postalAddress?.locality ?? ''}
					country={selectedPlace.postalAddress?.country ?? ''}
					existingReview={currentReview}
					onSubmit={(review) => {
						currentReview = review;
						currentWishlistItem = null;
						isEditingReview = false;
					}}
				/>
			{:else if !currentReview}
				{#if !showRatingForm}
					<div class="flex gap-3">
						{#if currentWishlistItem}
							<button
								class="rounded-lg border border-green-300 bg-green-50 px-4 py-2 text-sm font-medium text-green-700 transition-colors hover:bg-green-100 disabled:opacity-50"
								onclick={removeFromWishlist}
								disabled={wishlistLoading}
							>
								★ Wishlisted — remove
							</button>
						{:else}
							<button
								class="rounded-lg border border-border bg-card px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:opacity-50"
								onclick={addToWishlist}
								disabled={wishlistLoading}
							>
								☆ Save to wishlist
							</button>
						{/if}
						<button
							class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
							onclick={() => (showRatingForm = true)}
						>
							📝 Add review
						</button>
					</div>
				{:else}
					<RatingForm
						googlePlacesId={selectedPlace.name || ''}
						restaurantName={selectedPlace.displayName?.text || selectedPlace.name || ''}
						restaurantAddress={selectedPlace.formattedAddress || ''}
						city={selectedPlace.postalAddress?.locality ?? ''}
						country={selectedPlace.postalAddress?.country ?? ''}
						onSubmit={(review) => {
							currentReview = review;
							currentWishlistItem = null;
							showRatingForm = false;
						}}
					/>
					<button
						class="mt-2 text-sm text-muted-foreground hover:text-foreground hover:underline"
						onclick={() => (showRatingForm = false)}
					>
						Cancel
					</button>
				{/if}
			{/if}
		</div>
	{/if}
</div>
