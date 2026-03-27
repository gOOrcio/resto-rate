<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';

	let reviews = $state<ReviewProto[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleting = $state<Set<string>>(new Set());

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

	onMount(() => {
		if (auth.isLoggedIn) loadReviews();
		else loading = false;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Reviews</h2>

	{#if !auth.isLoggedIn}
		<p class="text-sm text-gray-500">Please sign in to view your reviews.</p>
	{:else if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if reviews.length === 0}
		<p class="text-sm text-gray-500">
			No reviews yet. Search for a restaurant on the <a href="/" class="text-blue-600 hover:underline">home page</a> to leave one.
		</p>
	{:else}
		<ul class="space-y-3">
			{#each reviews as review (review.id)}
				<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
					{#if editingId === review.id}
						<RatingForm
							googlePlacesId={review.googlePlacesId}
							restaurantName={review.restaurantName}
							restaurantAddress=""
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
						<div class="mb-2 flex items-start justify-between gap-2">
							<div class="min-w-0">
								{#if review.restaurantName}
									<p class="truncate font-medium text-gray-900">{review.restaurantName}</p>
								{/if}
								{#if review.restaurantAddress}
									<p class="truncate text-sm text-gray-500">{review.restaurantAddress}</p>
								{/if}
								{#if review.restaurantCity || review.restaurantCountry}
									<p class="text-xs text-gray-400">{[review.restaurantCity, review.restaurantCountry].filter(Boolean).join(', ')}</p>
								{/if}
							</div>
							<div class="flex shrink-0 gap-1">
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

						<div class="mb-2 flex items-center gap-2">
							<div class="flex items-center gap-0.5">
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

						{#if review.comment}
							<p class="mb-2 text-sm leading-relaxed text-gray-600">{review.comment}</p>
						{/if}

						{#if review.tags && review.tags.length > 0}
							<div class="flex flex-wrap gap-1.5">
								{#each review.tags as tag}
									<span class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">
										{tag}
									</span>
								{/each}
							</div>
						{/if}
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
