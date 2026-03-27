<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { Star } from '@lucide/svelte';

	const googlePlacesId = $derived(decodeURIComponent(page.params.googlePlacesId));

	let reviews = $state<ReviewProto[]>([]);
	let averageRating = $state(0);
	let restaurantName = $state('');
	let restaurantAddress = $state('');
	let restaurantCity = $state('');
	let restaurantCountry = $state('');
	let loading = $state(true);
	let error = $state('');

	async function loadReviews() {
		try {
			const res = await client.reviews.listRestaurantReviews({ googlePlacesId });
			reviews = res.reviews;
			averageRating = res.averageRating;
			restaurantName = res.restaurantName;
			restaurantAddress = res.restaurantAddress;
			restaurantCity = res.restaurantCity;
			restaurantCountry = res.restaurantCountry;
		} catch (e) {
			console.error('Failed to load restaurant reviews:', e);
			error = 'Failed to load reviews.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		if (auth.isLoggedIn) loadReviews();
		else loading = false;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	{#if !auth.isLoggedIn}
		<p class="text-sm text-gray-500">Please sign in to view restaurant reviews.</p>
	{:else if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if error}
		<p class="text-sm text-red-600">{error}</p>
	{:else}
		<!-- Restaurant header -->
		<div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
			<h2 class="text-2xl font-semibold text-blue-800">{restaurantName || 'Restaurant'}</h2>
			{#if restaurantAddress}
				<p class="mt-1 text-sm text-gray-500">{restaurantAddress}</p>
			{/if}
			{#if restaurantCity || restaurantCountry}
				<p class="text-xs text-gray-400">{[restaurantCity, restaurantCountry].filter(Boolean).join(', ')}</p>
			{/if}
			{#if reviews.length > 0}
				<div class="mt-3 flex items-center gap-2">
					<div class="flex items-center gap-0.5">
						{#each Array(5) as _, i}
							<Star
								class="h-5 w-5 {i < Math.round(averageRating)
									? 'fill-amber-400 text-amber-400'
									: 'fill-none text-gray-300'}"
							/>
						{/each}
					</div>
					<span class="font-semibold text-gray-800">{averageRating.toFixed(1)}</span>
					<span class="text-sm text-gray-500">({reviews.length} {reviews.length === 1 ? 'review' : 'reviews'})</span>
				</div>
			{/if}
		</div>

		<!-- Reviews -->
		{#if reviews.length === 0}
			<p class="text-sm text-gray-500">No reviews yet from you or your friends for this restaurant.</p>
		{:else}
			<ul class="space-y-3">
				{#each reviews as review (review.id)}
					<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
						<div class="mb-2 flex items-start justify-between gap-2">
							<div>
								<p class="font-medium text-gray-900">{review.authorName || 'Unknown'}</p>
								<div class="mt-1 flex items-center gap-1.5">
									<div class="flex items-center gap-0.5">
										{#each Array(5) as _, i}
											<Star
												class="h-3.5 w-3.5 {i < review.rating
													? 'fill-amber-400 text-amber-400'
													: 'fill-none text-gray-300'}"
											/>
										{/each}
									</div>
									<span class="text-sm font-semibold text-gray-700">{review.rating.toFixed(1)}</span>
								</div>
							</div>
							<span class="shrink-0 text-xs text-gray-400">
								{new Date(Number(review.createdAt) * 1000).toLocaleDateString()}
							</span>
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
					</li>
				{/each}
			</ul>
		{/if}
	{/if}
</div>
