<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { ReviewSortBy, TagFilterMode } from '$lib/client/generated/reviews/v1/reviews_service_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Calendar, Receipt } from '@lucide/svelte';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import TagFilter from '$lib/ui/components/TagFilter.svelte';
	import { WouldVisitAgain } from '$lib/client/generated/reviews/v1/review_pb';
	import * as m from '$lib/paraglide/messages';

	let reviews = $state<ReviewProto[]>([]);
	let loading = $state(true);
	let editingId = $state<string | null>(null);
	let deleting = $state<Set<string>>(new Set());
	let searchedPlace = $state<Place | null>(null);
	let mounted = $state(false);
	let showAddReview = $state(false);

	// Filter state
	let showFilters = $state(false);
	let tagSlugs = $state<string[]>([]);
	let tagMode = $state<'OR' | 'AND'>('OR');
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

	let uniqueCities = $derived(
		[...new Set(reviews.map((r) => r.restaurantCity).filter(Boolean))].sort()
	);
	let uniqueCountries = $derived(
		[...new Set(reviews.map((r) => r.restaurantCountry).filter(Boolean))].sort()
	);

	let hasRatingRangeError = $derived(
		minRating > 0 && maxRating > 0 && minRating > maxRating
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
		tagMode = 'OR';
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
				tagFilterMode: tagMode === 'AND' ? TagFilterMode.AND : TagFilterMode.OR,
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

	$effect(() => {
		if (!mounted) return;
		if (hasRatingRangeError) return;
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
		showAddReview = false;
	}

	$effect(() => {
		if (auth.loading || mounted) return;
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		mounted = true;
	});

	const WOULD_VISIT_AGAIN_LABELS: Record<number, { text: () => string; cls: string }> = {
		[WouldVisitAgain.YES]: { text: m.common_would_visit_again_yes, cls: 'text-emerald-600 dark:text-emerald-400' },
		[WouldVisitAgain.MAYBE]: { text: m.common_would_visit_again_maybe, cls: 'text-amber-600 dark:text-amber-400' },
		[WouldVisitAgain.NO]: { text: m.common_would_visit_again_no, cls: 'text-red-600 dark:text-red-400' },
	};

	function formatVisitDate(ts: bigint): string {
		if (!ts) return '';
		return new Date(Number(ts) * 1000).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
	}
</script>

<div class="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6">
	<!-- Page header -->
	<div class="flex items-start justify-between gap-4">
		<div>
			<h1 class="font-display text-3xl font-semibold text-foreground">{m.reviews_page_title()}</h1>
			{#if !loading}
				<p class="mt-1 text-sm text-muted-foreground">
					{reviews.length === 0 && activeFilterCount === 0
						? 'No reviews yet'
						: `${reviews.length} restaurant${reviews.length === 1 ? '' : 's'}`}
				</p>
			{/if}
		</div>
		<button
			class="shrink-0 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
			onclick={() => { showAddReview = !showAddReview; searchedPlace = null; }}
		>
			{showAddReview ? m.common_cancel() : m.reviews_add()}
		</button>
	</div>

	<!-- Add review panel -->
	{#if showAddReview}
		<div class="relative z-10 card-reveal rounded-lg border border-border bg-card p-5">
			<p class="mb-3 text-sm font-medium text-foreground">{m.reviews_search_label()}</p>
			<RestaurantSearch
				placeholder="Restaurant name or address…"
				onSelect={handleSearchSelect}
			/>
			{#if searchedPlace}
				<div class="mt-4 space-y-3 border-t border-border pt-4">
					<div>
						<p class="font-medium text-foreground">{searchedPlace.displayName?.text || ''}</p>
						<p class="text-sm text-muted-foreground">{searchedPlace.formattedAddress || ''}</p>
					</div>
					<RatingForm
						googlePlacesId={searchedPlace.name || ''}
						restaurantName={searchedPlace.displayName?.text || ''}
						restaurantAddress={searchedPlace.formattedAddress || ''}
						city={searchedPlace.postalAddress?.locality || searchedPlace.postalAddress?.administrativeArea || ''}
						country={searchedPlace.postalAddress?.country || ''}
						photoReference={searchedPlace.photos?.[0]?.name || ''}
						onSubmit={handleNewReview}
					/>
					<button
						class="text-sm text-muted-foreground hover:text-foreground"
						onclick={() => (searchedPlace = null)}
					>
						{m.reviews_choose_different()}
					</button>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Filter bar -->
	<div class="space-y-3">
		<div class="flex flex-wrap items-center gap-2">
			<button
				class="rounded-md border px-3 py-1.5 text-sm font-medium transition-colors {showFilters
					? 'border-primary bg-primary text-primary-foreground'
					: 'border-border bg-card text-foreground hover:bg-muted'}"
				onclick={() => (showFilters = !showFilters)}
			>
				Filters{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
			</button>
			{#if activeFilterCount > 0}
				<button
					class="text-sm text-muted-foreground hover:text-foreground"
					onclick={clearFilters}
				>
					{m.common_clear_all()}
				</button>
			{/if}
			<div class="ml-auto flex items-center gap-2">
				<label for="reviews-sort" class="text-sm text-muted-foreground">{m.common_sort()}</label>
				<select
					id="reviews-sort"
					bind:value={sortBy}
					class="rounded-md border border-border bg-card py-1 pl-2 pr-6 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
				>
					<option value="date-desc">{m.common_sort_newest()}</option>
					<option value="date-asc">{m.common_sort_oldest()}</option>
					<option value="rating-desc">{m.common_sort_rating_high()}</option>
					<option value="rating-asc">{m.common_sort_rating_low()}</option>
				</select>
			</div>
		</div>

		{#if showFilters}
			<div class="card-reveal space-y-4 rounded-lg border border-border bg-card p-4">
				<!-- Tags -->
				<div>
					<span class="mb-2 block text-sm font-medium text-foreground">{m.common_filter_tags()}</span>
					<TagFilter bind:selected={tagSlugs} bind:mode={tagMode} />
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
					<label for="comment-search" class="mb-1 block text-sm font-medium text-foreground">
						{m.common_filter_comment()}
					</label>
					<input
						id="comment-search"
						type="text"
						bind:value={commentRaw}
						placeholder={m.common_filter_comment_placeholder()}
						class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:ring-1 focus:ring-ring focus:outline-none"
					/>
				</div>

				<!-- City + Country -->
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="filter-city" class="mb-1 block text-sm font-medium text-foreground">{m.common_filter_city()}</label>
						<select
							id="filter-city"
							bind:value={city}
							class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
						>
							<option value="">{m.common_filter_all_cities()}</option>
							{#each uniqueCities as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="filter-country" class="mb-1 block text-sm font-medium text-foreground">{m.common_filter_country()}</label>
						<select
							id="filter-country"
							bind:value={country}
							class="w-full rounded-md border border-border bg-card px-3 py-1.5 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
						>
							<option value="">{m.common_filter_all_countries()}</option>
							{#each uniqueCountries as c}
								<option value={c}>{c}</option>
							{/each}
						</select>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Content -->
	{#if loading}
		<div class="flex items-center gap-2 py-8 text-sm text-muted-foreground">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary"></div>
			Loading…
		</div>
	{:else if reviews.length === 0}
		<div class="py-16 text-center">
			<p class="text-muted-foreground">
				{#if activeFilterCount > 0}
					{m.reviews_empty_with_filters()}
					<button type="button" onclick={clearFilters} class="underline hover:no-underline">
						{m.common_clear_filters()}
					</button>
				{:else}
					{m.reviews_empty_no_filters()}
				{/if}
			</p>
		</div>
	{:else}
		<ul class="grid grid-cols-1 items-start gap-4 sm:grid-cols-2 lg:gap-5">
			{#each reviews as review, i (review.id)}
				<li
					class="card-reveal flex flex-col rounded-lg border border-border bg-card"
					style="animation-delay: {Math.min(i * 50, 300)}ms"
				>
					{#if editingId === review.id}
						<div class="p-5">
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
							<button
								class="mt-3 text-sm text-muted-foreground hover:text-foreground"
								onclick={() => (editingId = null)}
							>
								{m.common_cancel()}
							</button>
						</div>
					{:else}
						<div class="flex flex-col gap-3 p-5">
							<!-- Restaurant info + rating badge -->
							<ExpandableRestaurantInfo
								googlePlacesId={review.googlePlacesId}
								name={review.restaurantName}
								address={review.restaurantAddress}
								city={review.restaurantCity}
								country={review.restaurantCountry}
								photoReference={review.restaurantPhotoReference || ''}
								rating={review.rating}
							/>

							<!-- Comment -->
							{#if review.comment}
								<p class="line-clamp-3 text-sm leading-relaxed text-muted-foreground">
									{review.comment}
								</p>
							{/if}

							<!-- Tags -->
							{#if review.tags && review.tags.length > 0}
								<div class="flex flex-wrap gap-1.5">
									{#each review.tags as tag}
										<span
											class="rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground"
										>
											{tag}
										</span>
									{/each}
								</div>
							{/if}

							<!-- Extra detail fields -->
							{#if review.visitedAt || review.pricePaidPerPerson || review.wouldVisitAgain || review.dishHighlights}
								{@const visitDate = formatVisitDate(review.visitedAt)}
								{@const wvaEntry = WOULD_VISIT_AGAIN_LABELS[review.wouldVisitAgain]}
								<div class="space-y-2 border-t border-border pt-3">
									<div class="flex flex-wrap items-center gap-x-3 gap-y-1.5">
										{#if visitDate}
											<span class="flex items-center gap-1 text-xs text-muted-foreground">
												<Calendar class="h-3 w-3 shrink-0" />{visitDate}
											</span>
										{/if}
										{#if review.pricePaidPerPerson}
											<span class="flex items-center gap-1 text-xs text-muted-foreground">
												<Receipt class="h-3 w-3 shrink-0" />${review.pricePaidPerPerson}/person
											</span>
										{/if}
										{#if wvaEntry}
											<span class="rounded-full border border-current px-2 py-0.5 text-xs font-medium {wvaEntry.cls}">
												{wvaEntry.text()}
											</span>
										{/if}
									</div>
									{#if review.dishHighlights}
										<p class="text-xs text-muted-foreground">
											<span class="font-medium text-foreground">{m.common_highlights()}</span> {review.dishHighlights}
										</p>
									{/if}
								</div>
							{/if}
						</div>

						<!-- Card footer -->
						<div class="flex items-center justify-between border-t border-border px-5 py-3">
							{#if review.googlePlacesId}
								<a
									href="/restaurants/{encodeURIComponent(review.googlePlacesId)}"
									class="text-xs text-muted-foreground hover:text-foreground hover:underline"
								>
									{m.common_details_and_reviews()}
								</a>
							{:else}
								<span></span>
							{/if}
							<div class="flex items-center gap-3">
								<button
									class="text-xs text-muted-foreground transition-colors hover:text-foreground"
									onclick={() => (editingId = review.id)}
								>
									{m.common_edit()}
								</button>
								<button
									class="text-xs text-muted-foreground transition-colors hover:text-destructive disabled:opacity-40"
									disabled={deleting.has(review.id)}
									onclick={() => deleteReview(review.id)}
								>
									{deleting.has(review.id) ? m.common_deleting() : m.common_delete()}
								</button>
							</div>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
