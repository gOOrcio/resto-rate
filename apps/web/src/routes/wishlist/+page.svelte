<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import { WishlistSortBy } from '$lib/client/generated/wishlist/v1/wishlist_service_pb';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';
	import type { Place } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
	import { Button } from '$lib/components/ui/button/index.js';
	import ExpandableRestaurantInfo from '$lib/ui/components/ExpandableRestaurantInfo.svelte';
	import RatingForm from '$lib/ui/components/RatingForm.svelte';
	import RestaurantSearch from '$lib/ui/components/RestaurantSearch.svelte';

	let items = $state<WishlistItemProto[]>([]);
	let loading = $state(true);
	let removing = $state<Set<string>>(new Set());
	let ratingId = $state<string | null>(null);
	let mounted = $state(false);

	let searchedPlace = $state<Place | null>(null);
	let searchAction = $state<'review' | null>(null);
	let savingToWishlist = $state(false);

	// Filter state
	let city = $state('');
	let country = $state('');
	let sortBy = $state('date-desc');

	let activeFilterCount = $derived(
		(city.trim() !== '' ? 1 : 0) +
			(country.trim() !== '' ? 1 : 0) +
			(sortBy !== 'date-desc' ? 1 : 0)
	);

	function clearFilters() {
		city = '';
		country = '';
		sortBy = 'date-desc';
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
	}

	async function saveToWishlist() {
		if (!searchedPlace) return;
		savingToWishlist = true;
		try {
			await client.wishlist.addToWishlist({
				googlePlacesId: searchedPlace.name || '',
				restaurantName: searchedPlace.displayName?.text || '',
				restaurantAddress: searchedPlace.formattedAddress || '',
				city: searchedPlace.postalAddress?.locality ?? '',
				country: searchedPlace.postalAddress?.country ?? ''
			});
			// Reload the wishlist to show the new item
			await loadWishlist();
			searchedPlace = null;
		} catch (e) {
			console.error('Failed to add to wishlist:', e);
		} finally {
			savingToWishlist = false;
		}
	}

	function handleSearchReview(review: ReviewProto) {
		// Item was reviewed — remove from wishlist list (backend auto-removes)
		// and close the search panel
		items = items.filter((i) => i.googlePlacesId !== review.googlePlacesId);
		searchedPlace = null;
		searchAction = null;
	}

	async function loadWishlist() {
		loading = true;
		try {
			const res = await client.wishlist.listWishlist({
				city,
				country,
				sortBy: toSortByEnum(sortBy)
			});
			items = res.items ?? [];
		} catch (e) {
			console.error('Failed to load wishlist:', e);
		} finally {
			loading = false;
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

	// Reactive reload when filters change (only after auth + mount)
	$effect(() => {
		if (!mounted) return;
		void [city, country, sortBy];
		loadWishlist();
	});

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		mounted = true;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Wishlist</h2>

	<!-- Filter bar -->
	<div class="flex flex-wrap items-center gap-2">
		{#if activeFilterCount > 0}
			<Button variant="ghost" size="sm" onclick={clearFilters}
				>Clear filters ({activeFilterCount})</Button
			>
		{/if}
		<div class="ml-auto flex items-center gap-2">
			<label for="wishlist-sort" class="text-sm text-gray-600">Sort:</label>
			<select
				id="wishlist-sort"
				bind:value={sortBy}
				class="rounded-md border border-gray-300 px-2 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
			>
				<option value="date-desc">Newest first</option>
				<option value="date-asc">Oldest first</option>
				<option value="name-asc">Name A–Z</option>
				<option value="name-desc">Name Z–A</option>
			</select>
		</div>
		<div class="flex w-full gap-3">
			<div class="flex-1">
				<input
					type="text"
					bind:value={city}
					placeholder="Filter by city…"
					class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
				/>
			</div>
			<div class="flex-1">
				<input
					type="text"
					bind:value={country}
					placeholder="Filter by country…"
					class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
				/>
			</div>
		</div>
	</div>

	<section class="space-y-3">
		<h3 class="text-lg font-medium text-gray-800">Find a restaurant</h3>
		<RestaurantSearch
			placeholder="Search to add to wishlist or review…"
			onSelect={handleSearchSelect}
		/>
		{#if searchedPlace}
			<div class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
				<div>
					<p class="font-medium text-gray-900">
						{searchedPlace.displayName?.text || searchedPlace.name || ''}
					</p>
					<p class="text-sm text-gray-500">{searchedPlace.formattedAddress || ''}</p>
				</div>

				{#if !searchAction}
					<div class="flex gap-2">
						<Button onclick={saveToWishlist} disabled={savingToWishlist}>
							{savingToWishlist ? 'Saving…' : '☆ Save to wishlist'}
						</Button>
						<Button variant="secondary" onclick={() => (searchAction = 'review')}>
							📝 Add review
						</Button>
						<Button variant="ghost" onclick={() => (searchedPlace = null)}>Cancel</Button>
					</div>
				{:else if searchAction === 'review'}
					<RatingForm
						googlePlacesId={searchedPlace.name || ''}
						restaurantName={searchedPlace.displayName?.text || ''}
						restaurantAddress={searchedPlace.formattedAddress || ''}
						onSubmit={handleSearchReview}
					/>
					<Button variant="ghost" size="sm" onclick={() => (searchAction = null)}>Back</Button>
				{/if}
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
	{:else if items.length === 0}
		<p class="text-sm text-gray-500">
			{#if activeFilterCount > 0}
				No wishlist items match the current filters. <button
					type="button"
					onclick={clearFilters}
					class="text-blue-600 underline hover:no-underline">Clear filters</button
				>
			{:else}
				Your wishlist is empty. Search for a restaurant above to add one.
			{/if}
		</p>
	{:else}
		<ul class="space-y-3">
			{#each items as item (item.id)}
				<li class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
					<ExpandableRestaurantInfo
						googlePlacesId={item.googlePlacesId}
						name={item.restaurantName}
						address={item.restaurantAddress}
						city={item.city}
						country={item.country}
					/>

					{#if ratingId !== item.id}
						<div class="flex gap-2 border-t border-gray-100 pt-1">
							<Button
								variant="outline"
								size="sm"
								class="text-red-600 hover:border-red-300 hover:text-red-700"
								disabled={removing.has(item.googlePlacesId)}
								onclick={() => remove(item.googlePlacesId)}
							>
								{removing.has(item.googlePlacesId) ? 'Removing…' : 'Remove'}
							</Button>
							<Button variant="secondary" size="sm" onclick={() => (ratingId = item.id)}>
								Rate this place
							</Button>
						</div>
					{:else}
						<div class="space-y-3 border-t border-gray-100 pt-2">
							<RatingForm
								googlePlacesId={item.googlePlacesId}
								restaurantName={item.restaurantName}
								restaurantAddress={item.restaurantAddress}
								onSubmit={() => {
									items = items.filter((i) => i.googlePlacesId !== item.googlePlacesId);
									ratingId = null;
								}}
							/>
							<Button variant="ghost" size="sm" onclick={() => (ratingId = null)}>Cancel</Button>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
