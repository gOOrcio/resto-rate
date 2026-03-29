<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	const googlePlacesId = $derived(decodeURIComponent(page.params.googlePlacesId ?? ''));

	let reviews = $state<ReviewProto[]>([]);
	let averageRating = $state(0);
	let restaurantName = $state('');
	let restaurantAddress = $state('');
	let restaurantCity = $state('');
	let restaurantCountry = $state('');
	let loading = $state(true);
	let error = $state('');
	let isWishlisted = $state(false);
	let wishlistLoading = $state(false);

	async function loadRestaurantData() {
		try {
			const [reviewsResult, wishlistResult] = await Promise.allSettled([
				client.reviews.listRestaurantReviews({ googlePlacesId }),
				client.wishlist.listWishlist({ googlePlacesId }),
			]);

			if (reviewsResult.status === 'rejected') {
				throw reviewsResult.reason;
			}

			const reviewsRes = reviewsResult.value;
			reviews = reviewsRes.reviews;
			averageRating = reviewsRes.averageRating;
			restaurantName = reviewsRes.restaurantName;
			restaurantAddress = reviewsRes.restaurantAddress;
			restaurantCity = reviewsRes.restaurantCity;
			restaurantCountry = reviewsRes.restaurantCountry;

			if (wishlistResult.status === 'fulfilled') {
				isWishlisted = (wishlistResult.value.items?.length ?? 0) > 0;
			} else {
				console.error('Failed to load wishlist state:', wishlistResult.reason);
			}
		} catch (e) {
			console.error('Failed to load restaurant data:', e);
			error = 'Failed to load reviews.';
		} finally {
			loading = false;
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
				});
				isWishlisted = true;
			}
		} catch (e) {
			console.error('Wishlist toggle error:', e);
		} finally {
			wishlistLoading = false;
		}
	}

	onMount(() => {
		if (auth.isLoggedIn) loadRestaurantData();
		else loading = false;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	{#if !auth.isLoggedIn}
		<p class="text-sm text-muted-foreground">Please sign in to view restaurant reviews.</p>
	{:else if loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
			Loading…
		</div>
	{:else if error}
		<p class="text-sm text-destructive">{error}</p>
	{:else}
		<!-- Restaurant header -->
		<div class="rounded-lg border border-border bg-card p-5 shadow-sm">
			<div class="flex items-start justify-between gap-4">
				<div class="min-w-0 flex-1">
					<h2 class="text-2xl font-semibold text-foreground">{restaurantName || 'Restaurant'}</h2>
					{#if restaurantAddress}
						<p class="mt-1 text-sm text-muted-foreground">{restaurantAddress}</p>
					{/if}
					{#if restaurantCity || restaurantCountry}
						<p class="text-xs text-muted-foreground">{[restaurantCity, restaurantCountry].filter(Boolean).join(', ')}</p>
					{/if}
				</div>
				<Button
					variant={isWishlisted ? 'outline' : 'secondary'}
					size="sm"
					onclick={toggleWishlist}
					disabled={wishlistLoading}
					class="shrink-0 gap-1.5"
					aria-pressed={isWishlisted}
					aria-busy={wishlistLoading}
				>
					{#if wishlistLoading}
						<div class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
						<span class="sr-only">Updating wishlist…</span>
					{:else if isWishlisted}
						★ Wishlisted
					{:else}
						☆ Wishlist
					{/if}
				</Button>
			</div>
			{#if reviews.length > 0}
				<div class="mt-3 flex items-center gap-2">
					<div class="flex items-center gap-0.5">
						{#each Array(5) as _, i}
							<Star
								class="h-5 w-5 {i < Math.round(averageRating)
									? 'fill-amber-400 text-amber-400'
									: 'fill-none text-gray-300 dark:text-gray-600'}"
							/>
						{/each}
					</div>
					<span class="font-semibold text-foreground">{averageRating.toFixed(1)}</span>
					<span class="text-sm text-muted-foreground">({reviews.length} {reviews.length === 1 ? 'review' : 'reviews'})</span>
				</div>
			{/if}
		</div>

		<!-- Reviews -->
		{#if reviews.length === 0}
			<p class="text-sm text-muted-foreground">No reviews yet from you or your friends for this restaurant.</p>
		{:else}
			<ul class="space-y-3">
				{#each reviews as review (review.id)}
					<li class="rounded-lg border border-border bg-card p-4 shadow-sm">
						<div class="mb-2 flex items-start justify-between gap-2">
							<div>
								<p class="font-medium text-foreground">{review.authorName || 'Unknown'}</p>
								<div class="mt-1 flex items-center gap-1.5">
									<div class="flex items-center gap-0.5">
										{#each Array(5) as _, i}
											<Star
												class="h-3.5 w-3.5 {i < review.rating
													? 'fill-amber-400 text-amber-400'
													: 'fill-none text-gray-300 dark:text-gray-600'}"
											/>
										{/each}
									</div>
									<span class="text-sm font-semibold text-foreground">{review.rating.toFixed(1)}</span>
								</div>
							</div>
							<span class="shrink-0 text-xs text-muted-foreground">
								{new Date(Number(review.createdAt) * 1000).toLocaleDateString()}
							</span>
						</div>

						{#if review.comment}
							<p class="mb-2 text-sm leading-relaxed text-muted-foreground">{review.comment}</p>
						{/if}

						{#if review.tags && review.tags.length > 0}
							<div class="flex flex-wrap gap-1.5">
								{#each review.tags as tag}
									<span class="rounded-full bg-secondary px-2.5 py-0.5 text-xs font-medium text-secondary-foreground">
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
