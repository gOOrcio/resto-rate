<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/state/auth.svelte';
	import client from '$lib/client/client';
	import type { WishlistItemProto } from '$lib/client/generated/wishlist/v1/wishlist_item_pb';
	import { Button } from '$lib/components/ui/button/index.js';

	let items = $state<WishlistItemProto[]>([]);
	let loading = $state(true);
	let removing = $state<Set<string>>(new Set());

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
		if (auth.isLoggedIn) loadWishlist();
		else loading = false;
	});
</script>

<div class="container mx-auto max-w-3xl space-y-6 p-6">
	<h2 class="text-2xl font-semibold text-blue-800">My Wishlist</h2>

	{#if !auth.isLoggedIn}
		<p class="text-sm text-gray-500">Please sign in to view your wishlist.</p>
	{:else if loading}
		<div class="flex items-center gap-2 text-sm text-gray-500">
			<div class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500"></div>
			Loading…
		</div>
	{:else if items.length === 0}
		<p class="text-sm text-gray-500">
			Your wishlist is empty. Search for a restaurant on the <a href="/" class="text-blue-600 hover:underline">home page</a> to add one.
		</p>
	{:else}
		<ul class="space-y-3">
			{#each items as item (item.id)}
				<li class="flex items-start justify-between gap-4 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
					<div class="min-w-0 flex-1">
						<p class="truncate font-medium text-gray-900">{item.restaurantName}</p>
						{#if item.restaurantAddress}
							<p class="truncate text-sm text-gray-500">{item.restaurantAddress}</p>
						{/if}
						{#if item.city || item.country}
							<p class="text-xs text-gray-400">{[item.city, item.country].filter(Boolean).join(', ')}</p>
						{/if}
					</div>
					<Button
						variant="outline"
						size="sm"
						disabled={removing.has(item.googlePlacesId)}
						onclick={() => remove(item.googlePlacesId)}
					>
						{removing.has(item.googlePlacesId) ? 'Removing…' : 'Remove'}
					</Button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
