<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { ReviewSortBy, TagFilterMode } from '$lib/client/generated/reviews/v1/reviews_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Star } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import TagPicker from '$lib/ui/components/TagPicker.svelte';

	let reviews = $state<ReviewProto[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleting = $state<Set<string>>(new Set());
	let searchedPlace = $state<Place | null>(null);
	let mounted = $state(false);

	// Filter state
	let showFilters = $state(false);
	let tagSlugs = $state<string[]>([]);
	let tagMode = $state<'or' | 'and'>('or');
	let minRating = $state(0);
	let maxRating = $state(0);
	let commentRaw = $state('');
	let commentSearch = $state('');
	let city = $state('');
	let country = $state('');
	let sortBy = $state('date-desc');

	// Debounce comment search 300 ms
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

	let activeFilterCount = $derived(
		(tagSlugs.length > 0 ? 1 : 0) +
			(minRating > 0 || maxRating > 0 ? 1 : 0) +
			(commentSearch.trim() !== '' ? 1 : 0) +
			(city.trim() !== '' ? 1 : 0) +
			(country.trim() !== '' ? 1 : 0) +
			(sortBy !== 'date-desc' ? 1 : 0)
	);

	function clearFilters() {
		tagSlugs = [];
		tagMode = 'or';
		minRating = 0;
		maxRating = 0;
		commentRaw = '';
		commentSearch = '';
		city = '';
		country = '';
		sortBy = 'date-desc';
	}

	function toSortByEnum(s: string): ReviewSortBy {
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
		loading = true;
		try {
			const res = await client.reviews.listReviews({
				tagSlugs,
				tagFilterMode: tagMode === 'and' ? TagFilterMode.AND : TagFilterMode.OR,
				minRating,
				maxRating,
				commentSearch,
				city,
				country,
				sortBy: toSortByEnum(sortBy)
			});
			reviews = res.reviews ?? [];
		} catch (e) {
			console.error('Failed to load reviews:', e);
		} finally {
			loading = false;
		}
	}

	// Reactive reload when any filter changes (only after mount + auth confirmed)
	$effect(() => {
		if (!mounted) return;
		if (ratingRangeError) return;
		void [tagSlugs, tagMode, minRating, maxRating, commentSearch, city, country, sortBy];
		loadReviews();
	});

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
		mounted = true;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Reviews</h2>

	<!-- Filter bar -->
	<div class="space-y-3">
		<div class="flex flex-wrap items-center gap-2">
			<Button
				variant={showFilters ? 'default' : 'outline'}
				size="sm"
				onclick={() => (showFilters = !showFilters)}
			>
				Filters{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
			</Button>
			{#if activeFilterCount > 0}
				<Button variant="ghost" size="sm" onclick={clearFilters}>Clear all</Button>
			{/if}
			<div class="ml-auto flex items-center gap-2">
				<label for="reviews-sort" class="text-sm text-gray-600">Sort:</label>
				<select
					id="reviews-sort"
					bind:value={sortBy}
					class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
				>
					<option value="date-desc">Newest first</option>
					<option value="date-asc">Oldest first</option>
					<option value="rating-desc">Highest rated</option>
					<option value="rating-asc">Lowest rated</option>
				</select>
			</div>
		</div>

		{#if showFilters}
			<div class="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
				<!-- Tags + AND/OR toggle -->
				<div>
					<div class="mb-1 flex items-center gap-3">
						<span class="text-sm font-medium text-gray-700">Tags</span>
						<div
							class="flex items-center gap-0.5 rounded-full border border-gray-300 bg-white p-0.5"
						>
							<button
								type="button"
								onclick={() => (tagMode = 'or')}
								class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode ===
								'or'
									? 'bg-blue-600 text-white'
									: 'text-gray-600 hover:bg-gray-100'}"
							>
								Any (OR)
							</button>
							<button
								type="button"
								onclick={() => (tagMode = 'and')}
								class="rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {tagMode ===
								'and'
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
					<label for="comment-search" class="mb-1 block text-sm font-medium text-gray-700">
						Comment contains
					</label>
					<input
						id="comment-search"
						type="text"
						bind:value={commentRaw}
						placeholder="Search in comments…"
						class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
					/>
				</div>

				<!-- City + Country -->
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="filter-city" class="mb-1 block text-sm font-medium text-gray-700"
							>City</label
						>
						<input
							id="filter-city"
							type="text"
							bind:value={city}
							placeholder="e.g. Paris"
							class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div>
						<label for="filter-country" class="mb-1 block text-sm font-medium text-gray-700"
							>Country</label
						>
						<input
							id="filter-country"
							type="text"
							bind:value={country}
							placeholder="e.g. France"
							class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<section class="space-y-3">
		<h3 class="text-lg font-medium text-gray-800">Add a review</h3>
		<RestaurantSearch
			placeholder="Search for a restaurant to review…"
			onSelect={handleSearchSelect}
		/>
		{#if searchedPlace}
			<div class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
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
			<div
				class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"
			></div>
			Loading…
		</div>
	{:else if reviews.length === 0}
		<p class="text-sm text-gray-500">
			{#if activeFilterCount > 0}
				No reviews match the current filters. <button
					type="button"
					onclick={clearFilters}
					class="text-blue-600 underline hover:no-underline">Clear filters</button
				>
			{:else}
				No reviews yet. Search for a restaurant above to leave one.
			{/if}
		</p>
	{:else}
		<ul class="space-y-3">
			{#each reviews as review (review.id)}
				<li class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
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

						<div class="space-y-2 border-t border-gray-100 pt-3">
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
									<span class="text-sm font-semibold text-gray-800">{review.rating.toFixed(1)}</span
									>
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
										<span
											class="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700"
											>{tag}</span
										>
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
