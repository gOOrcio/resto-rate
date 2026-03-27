<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';

	let reviews = $state<ReviewProto[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleting = $state<Set<string>>(new Set());
	let searchedPlace = $state<Place | null>(null);

	async function loadReviews() {
		try {
			const res = await client.reviews.listReviews({});
			reviews = res.reviews ?? [];
		} catch (e) {
			console.error('Failed to load reviews:', e);
		} finally {
			loading = false;
		}
	}

	async function deleteReview(id: string) {
		deleting = new Set([...deleting, id]);
		const removed = reviews.find((r) => r.id === id)!;
		reviews = reviews.filter((r) => r.id !== id);
		try {
			await client.reviews.deleteReview({ id });
		} catch (e) {
			console.error('Failed to delete review:', e);
			reviews = [...reviews, removed];
		} finally {
			deleting.delete(id);
			deleting = new Set(deleting);
		}
	}

	function handleSearchSelect(place: Place) {
		searchedPlace = place;
	}

	function handleNewReview(review: ReviewProto) {
		reviews = [review, ...reviews];
		searchedPlace = null;
	}

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		loadReviews();
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Reviews</h2>

	<section class="space-y-3">
		<h3 class="text-lg font-medium text-gray-800">Add a review</h3>
		<RestaurantSearch
			placeholder="Search for a restaurant to review…"
			onSelect={handleSearchSelect}
		/>
		{#if searchedPlace}
			<div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm space-y-3">
				<div>
					<p class="font-medium text-gray-900">{searchedPlace.displayName?.text || ''}</p>
					<p class="text-sm text-gray-500">{searchedPlace.formattedAddress || ''}</p>
				</div>
				<RatingForm
					googlePlacesId={searchedPlace.name || ''}
					restaurantName={searchedPlace.displayName?.text || ''}
					restaurantAddress={searchedPlace.formattedAddress || ''}
					onSubmit={handleNewReview}
				/>
				<Button variant="ghost" size="sm" onclick={() => (searchedPlace = null)}>Cancel</Button>
			</div>
		{/if}
	</section>

	{#if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if reviews.length === 0}
		<p class="text-sm text-gray-500">
			No reviews yet. Search for a restaurant above to leave one.
		</p>
	{:else}
		<ul class="space-y-3">
			{#each reviews as review (review.id)}
				<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm space-y-3">
					{#if editingId === review.id}
						<RatingForm
							googlePlacesId={review.googlePlacesId}
							restaurantName={review.restaurantName}
							restaurantAddress={review.restaurantAddress}
							existingReview={review}
							onSubmit={(updated) => {
								reviews = reviews.map((r) => (r.id === updated.id ? updated : r));
								editingId = null;
							}}
						/>
						<Button variant="ghost" size="sm" class="mt-2" onclick={() => (editingId = null)}>
							Cancel
						</Button>
					{:else}
						<ExpandableRestaurantInfo
							googlePlacesId={review.googlePlacesId}
							name={review.restaurantName}
							address={review.restaurantAddress}
							city={review.restaurantCity}
							country={review.restaurantCountry}
						/>

						<div class="border-t border-gray-100 pt-3 space-y-2">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<div class="flex gap-0.5">
										{#each Array(5) as _, i}
											<Star
												class="h-4 w-4 {i < review.rating
													? 'fill-amber-400 text-amber-400'
													: 'fill-none text-gray-300'}"
											/>
										{/each}
									</div>
									<span class="text-sm font-semibold text-gray-800">{review.rating.toFixed(1)}</span>
								</div>
								<div class="flex gap-1">
									<Button variant="outline" size="sm" onclick={() => (editingId = review.id)}>
										Edit
									</Button>
									<Button
										variant="outline"
										size="sm"
										disabled={deleting.has(review.id)}
										onclick={() => deleteReview(review.id)}
										class="text-red-600 hover:border-red-300 hover:text-red-700"
									>
										{deleting.has(review.id) ? 'Deleting…' : 'Delete'}
									</Button>
								</div>
							</div>

							{#if review.comment}
								<p class="text-sm leading-relaxed text-gray-600">{review.comment}</p>
							{/if}

							{#if review.tags && review.tags.length > 0}
								<div class="flex flex-wrap gap-1.5">
									{#each review.tags as tag}
										<span class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">{tag}</span>
									{/each}
								</div>
							{/if}
						</div>

						{#if review.googlePlacesId}
							<a
								href="/restaurants/{encodeURIComponent(review.googlePlacesId)}"
								class="text-xs text-blue-600 hover:underline"
							>
								See all reviews from friends →
							</a>
						{/if}
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
