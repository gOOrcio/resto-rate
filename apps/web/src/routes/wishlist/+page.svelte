<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
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

	let searchedPlace = $state<Place | null>(null);
	let searchAction = $state<'review' | null>(null);
	let savingToWishlist = $state(false);

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

	function handleSearchReview(_review: ReviewProto) {
		// Item was reviewed — remove from wishlist list (backend auto-removes)
		// and close the search panel
		items = items.filter((i) => i.googlePlacesId !== (searchedPlace?.name || ''));
		searchedPlace = null;
		searchAction = null;
	}

	async function loadWishlist() {
		try {
			const res = await client.wishlist.listWishlist({});
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

	onMount(() => {
		if (!auth.isLoggedIn) {
			goto('/?login=1');
			return;
		}
		loadWishlist();
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Wishlist</h2>

	<section class="space-y-3">
		<h3 class="text-lg font-medium text-gray-800">Find a restaurant</h3>
		<RestaurantSearch
			placeholder="Search to add to wishlist or review…"
			onSelect={handleSearchSelect}
		/>
		{#if searchedPlace}
			<div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm space-y-3">
				<div>
					<p class="font-medium text-gray-900">{searchedPlace.displayName?.text || searchedPlace.name || ''}</p>
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
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if items.length === 0}
		<p class="text-sm text-gray-500">
			Your wishlist is empty. Search for a restaurant above to add one.
		</p>
	{:else}
		<ul class="space-y-3">
			{#each items as item (item.id)}
				<li class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm space-y-3">
					<ExpandableRestaurantInfo
						googlePlacesId={item.googlePlacesId}
						name={item.restaurantName}
						address={item.restaurantAddress}
						city={item.city}
						country={item.country}
					/>

					{#if ratingId !== item.id}
						<div class="flex gap-2 pt-1 border-t border-gray-100">
							<Button
								variant="outline"
								size="sm"
								class="text-red-600 hover:text-red-700 hover:border-red-300"
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
						<div class="pt-2 border-t border-gray-100 space-y-3">
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
