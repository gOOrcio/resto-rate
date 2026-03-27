<script lang="ts">
	import clients from '$lib/client/client';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { goto } from '$app/navigation';
	import PlacePreviewCard from './PlacePreviewCard.svelte';
	import RatingForm from './RatingForm.svelte';
	import ReviewSummary from './ReviewSummary.svelte';
	import RestaurantSearch from './RestaurantSearch.svelte';

	let selectedPlace = $state<Place | null>(null);
	let isCheckingReview = $state(false);
	let currentReview = $state<ReviewProto | null>(null);
	let isEditingReview = $state(false);
	let showRatingForm = $state(false);

	async function handleSelect(place: Place) {
		selectedPlace = place;
		currentReview = null;
		showRatingForm = false;
		isEditingReview = false;

		isCheckingReview = true;
		try {
			const res = await clients.reviews.listReviews({ googlePlacesId: place.name || '' });
			currentReview = res.reviews?.[0] ?? null;
		} catch {
			currentReview = null;
		} finally {
			isCheckingReview = false;
		}
	}

	async function addToWishlist() {
		if (!selectedPlace) return;
		try {
			await clients.wishlist.addToWishlist({
				googlePlacesId: selectedPlace.name || '',
				restaurantName: selectedPlace.displayName?.text || selectedPlace.name || '',
				restaurantAddress: selectedPlace.formattedAddress || '',
				city: selectedPlace.postalAddress?.locality ?? '',
				country: selectedPlace.postalAddress?.country ?? ''
			});
			goto('/wishlist');
		} catch (e) {
			console.error('Wishlist add error:', e);
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
					class="text-sm text-blue-600 hover:underline"
				>
					See reviews from you &amp; friends →
				</a>
			{/if}

			{#if isCheckingReview}
				<div class="flex items-center gap-2 text-sm text-gray-500">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
					Checking your review…
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
						isEditingReview = false;
					}}
				/>
			{:else if !currentReview}
				{#if !showRatingForm}
					<div class="flex gap-3">
						<button
							class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
							onclick={addToWishlist}
						>
							☆ Save to wishlist
						</button>
						<button
							class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
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
							showRatingForm = false;
						}}
					/>
					<button
						class="mt-2 text-sm text-gray-500 hover:text-gray-700 hover:underline"
						onclick={() => (showRatingForm = false)}
					>
						Cancel
					</button>
				{/if}
			{/if}
		</div>
	{/if}
</div>
