<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { ReviewSortBy, TagFilterMode } from '$lib/client/generated/reviews/v1/reviews_service_pb';
	import { WishlistSortBy } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import type { FriendProto } from '$lib/client/generated/friendship/v1/friendship_pb';
	import { ConnectError, Code } from '@connectrpc/connect';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import TagPicker from '$lib/ui/components/TagPicker.svelte';
	import * as m from '$lib/paraglide/messages';

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

	let hasRatingRangeError = $derived(
		minRating > 0 && maxRating > 0 && minRating > maxRating
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
			if (ConnectError.from(e).code === Code.PermissionDenied) {
				notFriends = true;
			} else {
				console.error('Failed to load reviews:', e);
			}
		} finally {
			reviewsLoading = false;
		}
	}

	$effect(() => {
		if (!mounted || activeTab !== 'reviews' || notFriends) return;
		if (hasRatingRangeError) return;
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
			if (ConnectError.from(e).code === Code.PermissionDenied) {
				notFriends = true;
			} else {
				console.error('Failed to load wishlist:', e);
			}
		} finally {
			wishlistLoading = false;
		}
	}

	$effect(() => {
		if (!mounted || activeTab !== 'wishlist' || notFriends) return;
		void [wishlistCity, wishlistCountry, wishlistSortBy];
		loadWishlist();
	});

	async function initPage() {
		try {
			const res = await client.friendship.listFriends({});
			friend = res.friends.find((f) => f.userId === targetUserId) ?? null;
			if (!friend) notFriends = true;
		} catch (e) {
			console.error('Failed to load friends list:', e);
		}
		mounted = true;
	}

	$effect(() => {
		if (auth.loading || mounted) return;
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		void initPage();
	});
</script>

<div class="mx-auto max-w-3xl space-y-6 px-4 py-8 sm:px-6">
	<!-- Profile header -->
	<div class="flex items-start justify-between">
		<div>
			{#if friend}
				<h2 class="font-display text-3xl font-semibold text-foreground">{friend.name}</h2>
				<p class="text-sm text-muted-foreground">
					{#if friend.username}{friend.username} · {/if}{friend.email}
				</p>
			{:else if !mounted}
				<div class="h-7 w-48 animate-pulse rounded bg-muted"></div>
			{:else}
				<h2 class="font-display text-3xl font-semibold text-foreground">{m.friend_profile_title()}</h2>
			{/if}
		</div>
		<Button variant="outline" size="sm" href="/friends">{m.friend_profile_back()}</Button>
	</div>

	{#if notFriends}
		<div class="rounded-lg border border-border bg-muted p-6 text-center">
			<p class="font-medium text-foreground">{m.friend_profile_not_friends()}</p>
			<Button class="mt-4" href="/friends">{m.friend_profile_go_friends()}</Button>
		</div>
	{:else}
		<!-- Tabs -->
		<div class="border-b border-border">
			<nav role="tablist" aria-label="Friend profile sections" class="-mb-px flex gap-6">
				<button
					type="button"
					role="tab"
					id="tab-reviews"
					aria-controls="panel-reviews"
					aria-selected={activeTab === 'reviews'}
					onclick={() => (activeTab = 'reviews')}
					class="border-b-2 pb-3 text-sm font-medium transition-colors {activeTab === 'reviews'
						? 'border-primary text-primary'
						: 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'}"
				>
					{m.friend_profile_tab_reviews()}
				</button>
				<button
					type="button"
					role="tab"
					id="tab-wishlist"
					aria-controls="panel-wishlist"
					aria-selected={activeTab === 'wishlist'}
					onclick={() => (activeTab = 'wishlist')}
					class="border-b-2 pb-3 text-sm font-medium transition-colors {activeTab === 'wishlist'
						? 'border-primary text-primary'
						: 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'}"
				>
					{m.friend_profile_tab_wishlist()}
				</button>
			</nav>
		</div>

		<!-- Reviews tab -->
		{#if activeTab === 'reviews'}
		<div role="tabpanel" id="panel-reviews" aria-labelledby="tab-reviews">
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
						<Button variant="ghost" size="sm" onclick={clearReviewFilters}>{m.common_clear_all()}</Button>
					{/if}
					<div class="ml-auto flex items-center gap-2">
						<label for="review-sort" class="text-sm text-muted-foreground">{m.common_sort()}</label>
						<select
							id="review-sort"
							bind:value={reviewSortBy}
							class="rounded-md border border-border bg-card px-2 py-1 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
						>
							<option value="date-desc">{m.common_sort_newest()}</option>
							<option value="date-asc">{m.common_sort_oldest()}</option>
							<option value="rating-desc">{m.common_sort_rating_high()}</option>
							<option value="rating-asc">{m.common_sort_rating_low()}</option>
						</select>
					</div>
				</div>

				{#if showReviewFilters}
					<div class="space-y-4 rounded-lg border border-border bg-muted/40 p-4">
						<!-- Tags + AND/OR toggle -->
						<div>
							<div class="mb-1 flex items-center gap-3">
								<span class="text-sm font-medium text-foreground">{m.common_filter_tags()}</span>
								<div class="flex items-center gap-0.5 rounded-full border border-border bg-card p-0.5">
									<button
										type="button"
										onclick={() => (tagMode = 'or')}
										class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'or'
											? 'bg-primary text-primary-foreground'
											: 'text-muted-foreground hover:text-foreground'}"
									>
										{m.friend_profile_sort_any_or()}
									</button>
									<button
										type="button"
										onclick={() => (tagMode = 'and')}
										class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode === 'and'
											? 'bg-primary text-primary-foreground'
											: 'text-muted-foreground hover:text-foreground'}"
									>
										{m.friend_profile_sort_all_and()}
									</button>
								</div>
							</div>
							<TagPicker bind:selected={tagSlugs} />
						</div>

						<!-- Rating range -->
						<div class="flex flex-wrap items-center gap-2">
							<span class="text-sm font-medium text-foreground">{m.common_filter_rating()}</span>
							<select
								bind:value={minRating}
								class="rounded-md border border-border bg-card px-2 py-1 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
							>
								<option value={0}>{m.common_filter_min_rating()}</option>
								{#each [1, 2, 3, 4, 5] as n}
									<option value={n}>{n} ★</option>
								{/each}
							</select>
							<span class="text-sm text-muted-foreground">{m.common_filter_rating_to()}</span>
							<select
								bind:value={maxRating}
								class="rounded-md border border-border bg-card px-2 py-1 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
							>
								<option value={0}>{m.common_filter_max_rating()}</option>
								{#each [1, 2, 3, 4, 5] as n}
									<option value={n}>{n} ★</option>
								{/each}
							</select>
							{#if hasRatingRangeError}
								<p class="w-full text-xs text-destructive">{m.common_filter_rating_error()}</p>
							{/if}
						</div>

						<!-- Comment search -->
						<div>
							<label for="review-comment-search" class="mb-1 block text-sm font-medium text-foreground">
								{m.common_filter_comment()}
							</label>
							<input
								id="review-comment-search"
								type="text"
								bind:value={commentRaw}
								placeholder={m.common_filter_comment_placeholder()}
								class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
							/>
						</div>

						<!-- City + Country -->
						<div class="grid grid-cols-2 gap-3">
							<div>
								<label for="review-city" class="mb-1 block text-sm font-medium text-foreground">{m.common_filter_city()}</label>
								<input
									id="review-city"
									type="text"
									bind:value={reviewCity}
									placeholder={m.friend_profile_city_placeholder()}
									class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
								/>
							</div>
							<div>
								<label for="review-country" class="mb-1 block text-sm font-medium text-foreground">{m.common_filter_country()}</label>
								<input
									id="review-country"
									type="text"
									bind:value={reviewCountry}
									placeholder={m.friend_profile_country_placeholder()}
									class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
								/>
							</div>
						</div>
					</div>
				{/if}
			</div>

			{#if reviewsLoading}
				<div class="flex items-center gap-2 py-8 text-sm text-muted-foreground">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
					{m.common_loading()}
				</div>
			{:else if reviews.length === 0}
				<p class="text-sm text-muted-foreground">
					{#if activeReviewFilterCount > 0}
						{m.friend_profile_no_match_reviews()}
						<button type="button" onclick={clearReviewFilters} class="underline hover:no-underline">
							{m.common_clear_filters()}
						</button>
					{:else}
						{m.friend_profile_no_reviews()}
					{/if}
				</p>
			{:else}
				<ul class="space-y-3">
					{#each reviews as review (review.id)}
						<li class="card-reveal space-y-3 rounded-lg border border-border bg-card p-4">
							<ExpandableRestaurantInfo
								googlePlacesId={review.googlePlacesId}
								name={review.restaurantName}
								address={review.restaurantAddress}
								city={review.restaurantCity}
								country={review.restaurantCountry}
							/>

							<div class="space-y-2 border-t border-border pt-3">
								<div class="flex items-center gap-2">
									<div class="flex gap-0.5">
										{#each Array(5) as _, i}
											<Star
												class="h-4 w-4 {i < review.rating
													? 'fill-amber-400 text-amber-400'
													: 'fill-none text-muted-foreground/40'}"
											/>
										{/each}
									</div>
									<span class="text-sm font-semibold text-foreground">{review.rating.toFixed(1)}</span>
								</div>

								{#if review.comment}
									<p class="text-sm leading-relaxed text-muted-foreground">{review.comment}</p>
								{/if}

								{#if review.tags && review.tags.length > 0}
									<div class="flex flex-wrap gap-1.5">
										{#each review.tags as tag}
											<span class="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
												{tag}
											</span>
										{/each}
									</div>
								{/if}

								{#if review.googlePlacesId}
									<a
										href="/restaurants/{encodeURIComponent(review.googlePlacesId)}"
										class="text-xs text-muted-foreground hover:text-foreground hover:underline"
									>
										{m.friend_profile_see_all_reviews()}
									</a>
								{/if}
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
		{/if}

		<!-- Wishlist tab -->
		{#if activeTab === 'wishlist'}
		<div role="tabpanel" id="panel-wishlist" aria-labelledby="tab-wishlist">
			<div class="flex flex-wrap items-center gap-2">
				{#if activeWishlistFilterCount > 0}
					<Button variant="ghost" size="sm" onclick={clearWishlistFilters}>
						{m.common_clear_filters()} ({activeWishlistFilterCount})
					</Button>
				{/if}
				<div class="ml-auto flex items-center gap-2">
					<label for="wishlist-sort" class="text-sm text-muted-foreground">{m.common_sort()}</label>
					<select
						id="wishlist-sort"
						bind:value={wishlistSortBy}
						class="rounded-md border border-border bg-card px-2 py-1 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
					>
						<option value="date-desc">{m.common_sort_newest()}</option>
						<option value="date-asc">{m.common_sort_oldest()}</option>
						<option value="name-asc">{m.common_sort_name_az()}</option>
						<option value="name-desc">{m.common_sort_name_za()}</option>
					</select>
				</div>
				<div class="flex w-full gap-3">
					<input
						type="text"
						bind:value={wishlistCity}
						placeholder="Filter by city…"
						class="flex-1 rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
					/>
					<input
						type="text"
						bind:value={wishlistCountry}
						placeholder="Filter by country…"
						class="flex-1 rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
					/>
				</div>
			</div>

			{#if wishlistLoading}
				<div class="flex items-center gap-2 py-8 text-sm text-muted-foreground">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
					{m.common_loading()}
				</div>
			{:else if wishlistItems.length === 0}
				<p class="text-sm text-muted-foreground">
					{#if activeWishlistFilterCount > 0}
						{m.friend_profile_no_match_wishlist()}
						<button type="button" onclick={clearWishlistFilters} class="underline hover:no-underline">
							{m.common_clear_filters()}
						</button>
					{:else}
						{m.friend_profile_no_wishlist()}
					{/if}
				</p>
			{:else}
				<ul class="space-y-3">
					{#each wishlistItems as item (item.id)}
						<li class="card-reveal rounded-lg border border-border bg-card p-4">
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
		</div>
		{/if}
	{/if}
</div>
