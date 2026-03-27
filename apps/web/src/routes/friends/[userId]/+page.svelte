<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { ReviewSortBy, TagFilterMode } from '$lib/client/generated/reviews/v1/reviews_service_pb';
	import { WishlistSortBy } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import type { FriendProto } from '$lib/client/generated/friendship/v1/friendship_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import TagPicker from '$lib/ui/components/TagPicker.svelte';

	const targetUserId = page.params.userId;

	let friend = $state<FriendProto | null>(null);
	let notFriends = $state(false);
	let mounted = $state(false);

	// Tab state
	let activeTab = $state<'reviews' | 'wishlist'>('reviews');

	// --- Reviews ---
	let reviews = $state<ReviewProto[]>([]);
	let reviewsLoading = $state(false);
	let showReviewFilters = $state(false);
	let tagSlugs = $state<string[]>([]);
	let tagMode = $state<'or' | 'and'>('or');
	let minRating = $state(0);
	let maxRating = $state(0);
	let commentRaw = $state('');
	let commentSearch = $state('');
	let reviewCity = $state('');
	let reviewCountry = $state('');
	let reviewSortBy = $state('date-desc');

	$effect(() => {
		const val = commentRaw;
		const id = setTimeout(() => {
			commentSearch = val;
		}, 300);
		return () => clearTimeout(id);
	});

	let ratingRangeError = $derived(
		minRating > 0 && maxRating > 0 && minRating > maxRating
			? 'Min rating cannot exceed max rating'
			: null
	);

	let activeReviewFilterCount = $derived(
		(tagSlugs.length > 0 ? 1 : 0) +
			(minRating > 0 || maxRating > 0 ? 1 : 0) +
			(commentSearch.trim() !== '' ? 1 : 0) +
			(reviewCity.trim() !== '' ? 1 : 0) +
			(reviewCountry.trim() !== '' ? 1 : 0) +
			(reviewSortBy !== 'date-desc' ? 1 : 0)
	);

	function clearReviewFilters() {
		tagSlugs = [];
		tagMode = 'or';
		minRating = 0;
		maxRating = 0;
		commentRaw = '';
		commentSearch = '';
		reviewCity = '';
		reviewCountry = '';
		reviewSortBy = 'date-desc';
	}

	function toReviewSortEnum(s: string): ReviewSortBy {
		switch (s) {
			case 'date-asc':
				return ReviewSortBy.DATE_ASC;
			case 'rating-desc':
				return ReviewSortBy.RATING_DESC;
			case 'rating-asc':
				return ReviewSortBy.RATING_ASC;
			default:
				return ReviewSortBy.DATE_DESC;
		}
	}

	async function loadReviews() {
		reviewsLoading = true;
		try {
			const res = await client.reviews.listReviews({
				targetUserId,
				tagSlugs,
				tagFilterMode: tagMode === 'and' ? TagFilterMode.AND : TagFilterMode.OR,
				minRating,
				maxRating,
				commentSearch,
				city: reviewCity,
				country: reviewCountry,
				sortBy: toReviewSortEnum(reviewSortBy)
			});
			reviews = res.reviews ?? [];
		} catch (e: unknown) {
			const msg = (e as Error).message ?? '';
			if (msg.includes('permission_denied') || msg.includes('PermissionDenied')) {
				notFriends = true;
			}
			console.error('Failed to load reviews:', e);
		} finally {
			reviewsLoading = false;
		}
	}

	$effect(() => {
		if (!mounted || activeTab !== 'reviews') return;
		if (ratingRangeError) return;
		void [tagSlugs, tagMode, minRating, maxRating, commentSearch, reviewCity, reviewCountry, reviewSortBy];
		loadReviews();
	});

	// --- Wishlist ---
	let wishlistItems = $state<WishlistItemProto[]>([]);
	let wishlistLoading = $state(false);
	let wishlistCity = $state('');
	let wishlistCountry = $state('');
	let wishlistSortBy = $state('date-desc');

	let activeWishlistFilterCount = $derived(
		(wishlistCity.trim() !== '' ? 1 : 0) +
			(wishlistCountry.trim() !== '' ? 1 : 0) +
			(wishlistSortBy !== 'date-desc' ? 1 : 0)
	);

	function clearWishlistFilters() {
		wishlistCity = '';
		wishlistCountry = '';
		wishlistSortBy = 'date-desc';
	}

	function toWishlistSortEnum(s: string): WishlistSortBy {
		switch (s) {
			case 'date-asc':
				return WishlistSortBy.DATE_ASC;
			case 'name-asc':
				return WishlistSortBy.NAME_ASC;
			case 'name-desc':
				return WishlistSortBy.NAME_DESC;
			default:
				return WishlistSortBy.DATE_DESC;
		}
	}

	async function loadWishlist() {
		wishlistLoading = true;
		try {
			const res = await client.wishlist.listWishlist({
				targetUserId,
				city: wishlistCity,
				country: wishlistCountry,
				sortBy: toWishlistSortEnum(wishlistSortBy)
			});
			wishlistItems = res.items ?? [];
		} catch (e: unknown) {
			const msg = (e as Error).message ?? '';
			if (msg.includes('permission_denied') || msg.includes('PermissionDenied')) {
				notFriends = true;
			}
			console.error('Failed to load wishlist:', e);
		} finally {
			wishlistLoading = false;
		}
	}

	$effect(() => {
		if (!mounted || activeTab !== 'wishlist') return;
		void [wishlistCity, wishlistCountry, wishlistSortBy];
		loadWishlist();
	});

	onMount(async () => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		// Load friend info (for the profile header)
		try {
			const res = await client.friendship.listFriends({});
			friend = res.friends.find((f) => f.userId === targetUserId) ?? null;
			if (!friend) {
				notFriends = true;
			}
		} catch (e) {
			console.error('Failed to load friends list:', e);
		}
		mounted = true;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<!-- Profile header -->
	<div class="flex items-start justify-between">
		<div>
			{#if friend}
				<h2 class="text-2xl font-semibold text-blue-800">{friend.name}</h2>
				<p class="text-sm text-gray-500">
					{#if friend.username}{friend.username} · {/if}{friend.email}
				</p>
			{:else if !mounted}
				<div class="h-7 w-48 animate-pulse rounded bg-gray-200"></div>
			{:else}
				<h2 class="text-2xl font-semibold text-blue-800">Friend's Profile</h2>
			{/if}
		</div>
		<Button variant="outline" size="sm" href="/friends">← Back to friends</Button>
	</div>

	{#if notFriends}
		<div class="rounded-lg border border-amber-200 bg-amber-50 p-6 text-center">
			<p class="font-medium text-amber-800">You need to be friends to view this profile.</p>
			<Button class="mt-4" href="/friends">Go to Friends</Button>
		</div>
	{:else}
		<!-- Tabs -->
		<div class="border-b border-gray-200">
			<nav class="-mb-px flex gap-6">
				<button
					type="button"
					onclick={() => (activeTab = 'reviews')}
					class="border-b-2 pb-3 text-sm font-medium transition-colors {activeTab === 'reviews'
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
				>
					Reviews
				</button>
				<button
					type="button"
					onclick={() => (activeTab = 'wishlist')}
					class="border-b-2 pb-3 text-sm font-medium transition-colors {activeTab === 'wishlist'
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
				>
					Wishlist
				</button>
			</nav>
		</div>

		<!-- Reviews tab -->
		{#if activeTab === 'reviews'}
			<!-- Reviews filter bar -->
			<div class="space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<Button
						variant={showReviewFilters ? 'default' : 'outline'}
						size="sm"
						onclick={() => (showReviewFilters = !showReviewFilters)}
					>
						Filters{activeReviewFilterCount > 0 ? ` (${activeReviewFilterCount})` : ''}
					</Button>
					{#if activeReviewFilterCount > 0}
						<Button variant="ghost" size="sm" onclick={clearReviewFilters}>Clear all</Button>
					{/if}
					<div class="ml-auto flex items-center gap-2">
						<label for="review-sort" class="text-sm text-gray-600">Sort:</label>
						<select
							id="review-sort"
							bind:value={reviewSortBy}
							class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						>
							<option value="date-desc">Newest first</option>
							<option value="date-asc">Oldest first</option>
							<option value="rating-desc">Highest rated</option>
							<option value="rating-asc">Lowest rated</option>
						</select>
					</div>
				</div>

				{#if showReviewFilters}
					<div class="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
						<!-- Tags + AND/OR toggle -->
						<div>
							<div class="mb-1 flex items-center gap-3">
								<span class="text-sm font-medium text-gray-700">Tags</span>
								<div class="flex items-center gap-0.5 rounded-full border border-gray-300 bg-white p-0.5">
									<button
										type="button"
										onclick={() => (tagMode = 'or')}
										class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'or'
											? 'bg-blue-600 text-white'
											: 'text-gray-600 hover:bg-gray-100'}"
									>
										Any (OR)
									</button>
									<button
										type="button"
										onclick={() => (tagMode = 'and')}
										class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'and'
											? 'bg-blue-600 text-white'
											: 'text-gray-600 hover:bg-gray-100'}"
									>
										All (AND)
									</button>
								</div>
							</div>
							<TagPicker bind:selected={tagSlugs} />
						</div>

						<!-- Rating range -->
						<div class="flex flex-wrap items-center gap-2">
							<span class="text-sm font-medium text-gray-700">Rating:</span>
							<select
								bind:value={minRating}
								class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
							>
								<option value={0}>Min ★</option>
								{#each [1, 2, 3, 4, 5] as n}
									<option value={n}>{n} ★</option>
								{/each}
							</select>
							<span class="text-sm text-gray-500">to</span>
							<select
								bind:value={maxRating}
								class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
							>
								<option value={0}>Max ★</option>
								{#each [1, 2, 3, 4, 5] as n}
									<option value={n}>{n} ★</option>
								{/each}
							</select>
							{#if ratingRangeError}
								<p class="w-full text-xs text-red-600">{ratingRangeError}</p>
							{/if}
						</div>

						<!-- Comment search -->
						<div>
							<label for="review-comment-search" class="mb-1 block text-sm font-medium text-gray-700">
								Comment contains
							</label>
							<input
								id="review-comment-search"
								type="text"
								bind:value={commentRaw}
								placeholder="Search in comments…"
								class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
							/>
						</div>

						<!-- City + Country -->
						<div class="grid grid-cols-2 gap-3">
							<div>
								<label for="review-city" class="mb-1 block text-sm font-medium text-gray-700">City</label>
								<input
									id="review-city"
									type="text"
									bind:value={reviewCity}
									placeholder="e.g. Paris"
									class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
								/>
							</div>
							<div>
								<label for="review-country" class="mb-1 block text-sm font-medium text-gray-700">Country</label>
								<input
									id="review-country"
									type="text"
									bind:value={reviewCountry}
									placeholder="e.g. France"
									class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
								/>
							</div>
						</div>
					</div>
				{/if}
			</div>

			{#if reviewsLoading}
				<div class="flex items-center gap-2 text-sm text-gray-500">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
					Loading…
				</div>
			{:else if reviews.length === 0}
				<p class="text-sm text-gray-500">
					{#if activeReviewFilterCount > 0}
						No reviews match the current filters.
						<button type="button" onclick={clearReviewFilters} class="text-blue-600 underline hover:no-underline">
							Clear filters
						</button>
					{:else}
						No reviews yet.
					{/if}
				</p>
			{:else}
				<ul class="space-y-3">
					{#each reviews as review (review.id)}
						<li class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
							<ExpandableRestaurantInfo
								googlePlacesId={review.googlePlacesId}
								name={review.restaurantName}
								address={review.restaurantAddress}
								city={review.restaurantCity}
								country={review.restaurantCountry}
							/>

							<div class="space-y-2 border-t border-gray-100 pt-3">
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

								{#if review.comment}
									<p class="text-sm leading-relaxed text-gray-600">{review.comment}</p>
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

								{#if review.googlePlacesId}
									<a
										href="/restaurants/{encodeURIComponent(review.googlePlacesId)}"
										class="text-xs text-blue-600 hover:underline"
									>
										See all reviews →
									</a>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		{/if}

		<!-- Wishlist tab -->
		{#if activeTab === 'wishlist'}
			<!-- Wishlist filter bar -->
			<div class="flex flex-wrap items-center gap-2">
				{#if activeWishlistFilterCount > 0}
					<Button variant="ghost" size="sm" onclick={clearWishlistFilters}>
						Clear filters ({activeWishlistFilterCount})
					</Button>
				{/if}
				<div class="ml-auto flex items-center gap-2">
					<label for="wishlist-sort" class="text-sm text-gray-600">Sort:</label>
					<select
						id="wishlist-sort"
						bind:value={wishlistSortBy}
						class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
					>
						<option value="date-desc">Newest first</option>
						<option value="date-asc">Oldest first</option>
						<option value="name-asc">Name A–Z</option>
						<option value="name-desc">Name Z–A</option>
					</select>
				</div>
				<div class="flex w-full gap-3">
					<input
						type="text"
						bind:value={wishlistCity}
						placeholder="Filter by city…"
						class="flex-1 rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
					/>
					<input
						type="text"
						bind:value={wishlistCountry}
						placeholder="Filter by country…"
						class="flex-1 rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
					/>
				</div>
			</div>

			{#if wishlistLoading}
				<div class="flex items-center gap-2 text-sm text-gray-500">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
					Loading…
				</div>
			{:else if wishlistItems.length === 0}
				<p class="text-sm text-gray-500">
					{#if activeWishlistFilterCount > 0}
						No wishlist items match the current filters.
						<button type="button" onclick={clearWishlistFilters} class="text-blue-600 underline hover:no-underline">
							Clear filters
						</button>
					{:else}
						Empty wishlist.
					{/if}
				</p>
			{:else}
				<ul class="space-y-3">
					{#each wishlistItems as item (item.id)}
						<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
							<ExpandableRestaurantInfo
								googlePlacesId={item.googlePlacesId}
								name={item.restaurantName}
								address={item.restaurantAddress}
								city={item.city}
								country={item.country}
							/>
						</li>
					{/each}
				</ul>
			{/if}
		{/if}
	{/if}
</div>
