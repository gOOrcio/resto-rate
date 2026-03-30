<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { WishlistSortBy, WishlistTagFilterMode } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
	import { ConnectError, Code } from '@connectrpc/connect';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { extractCity } from '$lib/utils/place';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';
	import TagPicker from '$lib/ui/components/TagPicker.svelte';
	import TagFilter from '$lib/ui/components/TagFilter.svelte';
	import * as m from '$lib/paraglide/messages';

	let items = $state<WishlistItemProto[]>([]);
	let loading = $state(true);
	let removing = $state<Set<string>>(new Set());
	let ratingId = $state<string | null>(null);
	let editingTagsId = $state<string | null>(null);
	let editingTags = $state<string[]>([]);
	let savingTags = $state(false);
	let mounted = $state(false);

	let searchedPlace = $state<Place | null>(null);
	let searchAction = $state<'review' | null>(null);
	let savingToWishlist = $state(false);
	let saveError = $state<string | null>(null);
	let showSearch = $state(false);
	let pendingTags = $state<string[]>([]);

	// Filter state
	let showFilters = $state(false);
	let city = $state('');
	let country = $state('');
	let sortBy = $state('date-desc');
	let tagSlugs = $state<string[]>([]);
	let tagMode = $state<'OR' | 'AND'>('OR');

	let uniqueCities = $derived(
		[...new Set(items.map((i) => i.city).filter(Boolean))].sort()
	);
	let uniqueCountries = $derived(
		[...new Set(items.map((i) => i.country).filter(Boolean))].sort()
	);

	let activeFilterCount = $derived(
		(city.trim() !== '' ? 1 : 0) +
			(country.trim() !== '' ? 1 : 0) +
			(sortBy !== 'date-desc' ? 1 : 0) +
			(tagSlugs.length > 0 ? 1 : 0)
	);

	function clearFilters() {
		city = '';
		country = '';
		sortBy = 'date-desc';
		tagSlugs = [];
		tagMode = 'OR';
	}

	function toSortByEnum(s: string): WishlistSortBy {
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

	function handleSearchSelect(place: Place) {
		searchedPlace = place;
		searchAction = null;
		pendingTags = [];
		saveError = null;
	}

	async function saveToWishlist() {
		if (!searchedPlace) return;
		savingToWishlist = true;
		saveError = null;
		try {
			await client.wishlist.addToWishlist({
				googlePlacesId: searchedPlace.name || '',
				restaurantName: searchedPlace.displayName?.text || '',
				restaurantAddress: searchedPlace.formattedAddress || '',
				city: extractCity(searchedPlace),
				country: searchedPlace.postalAddress?.country ?? '',
				tagSlugs: pendingTags,
				photoReference: searchedPlace.photos?.[0]?.name || ''
			});
			await loadWishlist();
			searchedPlace = null;
			showSearch = false;
			pendingTags = [];
		} catch (e) {
			if (ConnectError.from(e).code === Code.FailedPrecondition) {
				saveError = m.wishlist_already_reviewed();
			} else {
				console.error('Failed to add to wishlist:', e);
				saveError = m.wishlist_save_error();
			}
		} finally {
			savingToWishlist = false;
		}
	}

	function handleSearchReview(review: ReviewProto) {
		items = items.filter((i) => i.googlePlacesId !== review.googlePlacesId);
		searchedPlace = null;
		searchAction = null;
		showSearch = false;
	}

	async function loadWishlist() {
		loading = true;
		try {
			const res = await client.wishlist.listWishlist({
				city,
				country,
				sortBy: toSortByEnum(sortBy),
				tagSlugs,
				tagFilterMode:
					tagMode === 'AND'
						? WishlistTagFilterMode.AND
						: WishlistTagFilterMode.OR
			});
			items = res.items ?? [];
		} catch (e) {
			console.error('Failed to load wishlist:', e);
		} finally {
			loading = false;
		}
	}

	async function saveTags(item: WishlistItemProto) {
		savingTags = true;
		try {
			await client.wishlist.addToWishlist({
				googlePlacesId: item.googlePlacesId,
				restaurantName: item.restaurantName,
				restaurantAddress: item.restaurantAddress,
				tagSlugs: editingTags
			});
			items = items.map((i) =>
				i.googlePlacesId === item.googlePlacesId ? { ...i, tags: [...editingTags] } : i
			);
			editingTagsId = null;
		} catch (e) {
			console.error('Failed to update tags:', e);
		} finally {
			savingTags = false;
		}
	}

	async function remove(googlePlacesId: string) {
		removing = new Set([...removing, googlePlacesId]);
		try {
			await client.wishlist.removeFromWishlist({ googlePlacesId });
			items = items.filter((i) => i.googlePlacesId !== googlePlacesId);
		} catch (e) {
			console.error('Failed to remove from wishlist:', e);
		} finally {
			removing.delete(googlePlacesId);
			removing = new Set(removing);
		}
	}

	$effect(() => {
		if (!mounted) return;
		void [city, country, sortBy, tagSlugs, tagMode];
		loadWishlist();
	});

	$effect(() => {
		if (auth.loading || mounted) return;
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		mounted = true;
	});
</script>

<div class="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6">
	<!-- Page header -->
	<div class="flex items-start justify-between gap-4">
		<div>
			<h1 class="font-display text-3xl font-semibold text-foreground">{m.wishlist_page_title()}</h1>
			{#if !loading}
				<p class="mt-1 text-sm text-muted-foreground">
					{items.length === 0 && activeFilterCount === 0
						? 'No places saved yet'
						: `${items.length} place${items.length === 1 ? '' : 's'} saved`}
				</p>
			{/if}
		</div>
		<button
			class="shrink-0 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
			onclick={() => { showSearch = !showSearch; searchedPlace = null; searchAction = null; pendingTags = []; }}
		>
			{showSearch ? m.common_cancel() : m.wishlist_add()}
		</button>
	</div>

	<!-- Add place panel -->
	{#if showSearch}
		<div class="relative z-10 card-reveal rounded-lg border border-border bg-card p-5">
			<p class="mb-3 text-sm font-medium text-foreground">{m.wishlist_search_label()}</p>
			<RestaurantSearch
				placeholder="Restaurant name or address…"
				onSelect={handleSearchSelect}
			/>
			{#if searchedPlace}
				<div class="mt-4 space-y-3 border-t border-border pt-4">
					<div>
						<p class="font-medium text-foreground">
							{searchedPlace.displayName?.text || searchedPlace.name || ''}
						</p>
						<p class="text-sm text-muted-foreground">{searchedPlace.formattedAddress || ''}</p>
					</div>

					{#if !searchAction}
						<TagPicker bind:selected={pendingTags} />

						{#if saveError}
							<p class="text-sm text-destructive">{saveError}</p>
						{/if}
						<div class="flex flex-wrap gap-2">
							<button
								class="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
								onclick={saveToWishlist}
								disabled={savingToWishlist}
							>
								{savingToWishlist ? m.common_saving() : m.wishlist_save_btn()}
							</button>
							<button
								class="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
								onclick={() => (searchAction = 'review')}
							>
								{m.wishlist_rate_instead()}
							</button>
							<button
								class="text-sm text-muted-foreground hover:text-foreground"
								onclick={() => (searchedPlace = null)}
							>
								{m.common_clear()}
							</button>
						</div>
					{:else if searchAction === 'review'}
						<RatingForm
							googlePlacesId={searchedPlace.name || ''}
							restaurantName={searchedPlace.displayName?.text || ''}
							restaurantAddress={searchedPlace.formattedAddress || ''}
							photoReference={searchedPlace.photos?.[0]?.name || ''}
							onSubmit={handleSearchReview}
						/>
						<button
							class="text-sm text-muted-foreground hover:text-foreground"
							onclick={() => (searchAction = null)}
						>
							{m.common_back()}
						</button>
					{/if}
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
				<label for="wishlist-sort" class="text-sm text-muted-foreground">{m.common_sort()}</label>
				<select
					id="wishlist-sort"
					bind:value={sortBy}
					class="rounded-md border border-border bg-card py-1 pl-2 pr-6 text-sm text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
				>
					<option value="date-desc">{m.common_sort_newest()}</option>
					<option value="date-asc">{m.common_sort_oldest()}</option>
					<option value="name-asc">{m.common_sort_name_az()}</option>
					<option value="name-desc">{m.common_sort_name_za()}</option>
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
	{:else if items.length === 0}
		<div class="py-16 text-center">
			<p class="text-muted-foreground">
				{#if activeFilterCount > 0}
					{m.wishlist_empty_with_filters()}
					<button type="button" onclick={clearFilters} class="underline hover:no-underline">
						{m.common_clear_filters()}
					</button>
				{:else}
					{m.wishlist_empty_no_filters()}
				{/if}
			</p>
		</div>
	{:else}
		<ul class="grid grid-cols-1 items-start gap-4 sm:grid-cols-2 lg:gap-5">
			{#each items as item, i (item.id)}
				<li
					class="card-reveal flex flex-col rounded-lg border border-border bg-card"
					style="animation-delay: {Math.min(i * 50, 300)}ms"
				>
					{#if ratingId !== item.id}
						<div class="p-5">
							<ExpandableRestaurantInfo
								googlePlacesId={item.googlePlacesId}
								name={item.restaurantName}
								address={item.restaurantAddress}
								city={item.city}
								country={item.country}
								photoReference={item.restaurantPhotoReference || ''}
							/>

							<!-- Tags section -->
							<div class="mt-3 rounded-lg border border-border">
								<button
									type="button"
									class="flex w-full items-center justify-between px-3 py-2 text-sm font-medium text-foreground"
									onclick={() => {
										if (editingTagsId === item.id) {
											editingTagsId = null;
										} else {
											editingTagsId = item.id;
											editingTags = [...(item.tags ?? [])];
										}
									}}
								>
									<span>Tags{item.tags?.length ? ` (${item.tags.length})` : ''}</span>
									<span class="text-muted-foreground transition-transform {editingTagsId === item.id ? 'rotate-90' : ''}">›</span>
								</button>
								{#if editingTagsId === item.id}
									<div class="border-t border-border px-3 pb-3 pt-2">
										<TagPicker bind:selected={editingTags} />
										<div class="mt-3 flex gap-2">
											<button
												type="button"
												disabled={savingTags}
												onclick={() => saveTags(item)}
												class="rounded-md bg-primary px-3 py-1 text-xs font-medium text-primary-foreground disabled:opacity-50 hover:opacity-90"
											>
												{savingTags ? m.common_saving() : m.wishlist_save_tags()}
											</button>
											<button
												type="button"
												onclick={() => (editingTagsId = null)}
												class="text-xs text-muted-foreground hover:text-foreground"
											>
												{m.common_cancel()}
											</button>
										</div>
									</div>
								{:else if item.tags?.length}
									<div class="border-t border-border px-3 py-2">
										<div class="flex flex-wrap gap-1.5">
											{#each item.tags as tag}
												<span class="rounded-full bg-secondary px-2.5 py-0.5 text-xs font-medium text-secondary-foreground">
													{tag}
												</span>
											{/each}
										</div>
									</div>
								{/if}
							</div>
						</div>
						<div class="flex items-center justify-between border-t border-border px-5 py-3">
							<div class="flex items-center gap-3">
								{#if item.googlePlacesId}
									<a
										href="/restaurants/{encodeURIComponent(item.googlePlacesId)}"
										class="text-xs text-muted-foreground hover:text-foreground hover:underline"
									>
										{m.common_details_and_reviews()}
									</a>
								{/if}
							</div>
							<div class="flex items-center gap-3">
								<button
									class="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
									onclick={() => (ratingId = item.id)}
								>
									{m.wishlist_rate_place()}
								</button>
								<button
									class="text-xs text-muted-foreground transition-colors hover:text-destructive disabled:opacity-40"
									disabled={removing.has(item.googlePlacesId)}
									onclick={() => remove(item.googlePlacesId)}
								>
									{removing.has(item.googlePlacesId) ? m.common_removing() : m.common_delete()}
								</button>
							</div>
						</div>
					{:else}
						<div class="flex flex-col gap-3 p-5">
							<RatingForm
								googlePlacesId={item.googlePlacesId}
								restaurantName={item.restaurantName}
								restaurantAddress={item.restaurantAddress}
								initialTags={item.tags ?? []}
								onSubmit={() => {
									items = items.filter((i) => i.googlePlacesId !== item.googlePlacesId);
									ratingId = null;
								}}
							/>
							<button
								class="text-sm text-muted-foreground hover:text-foreground"
								onclick={() => (ratingId = null)}
							>
								{m.common_cancel()}
							</button>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
