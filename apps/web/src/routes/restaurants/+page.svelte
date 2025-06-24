<script lang="ts">
	import { onMount } from 'svelte';
	import type { Place } from '@googlemaps/google-maps-services-js';
	import { apiClient } from '$lib/api';
	import { createPageLogger } from '$lib/logger';
	import { buttonVariants } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
	} from '$lib/components/ui/card';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow,
	} from '$lib/components/ui/table';
	import ProtectedRoute from '$lib/components/ProtectedRoute.svelte';
	import AddRestaurantModal from '$lib/components/AddRestaurantModal.svelte';
	import RestaurantSearch from '$lib/components/RestaurantSearch.svelte';
	import StarRating from '$lib/components/StarRating.svelte';

	type Restaurant = {
		id: string;
		name: string;
		address: string | null;
		rating: number | null;
		comment: string | null;
		createdAt: string;
		updatedAt: string;
	};

	let restaurants: Restaurant[] = [];
	let loading = false;
	let error: string | null = null;
	let modalOpen = false;
	let searchResults: Partial<Place>[] = [];
	let searchError: string | null = null;

	const logger = createPageLogger('restaurants');

	async function loadRestaurants() {
		loading = true;
		error = null;
		logger.debug('Loading restaurants list');
		try {
			const response = (await apiClient.getRestaurants()) as { restaurants: Restaurant[] };
			restaurants = response.restaurants;
			logger.info('Restaurants loaded successfully', { count: restaurants.length });
		} catch (err) {
			error = `Failed to load restaurants: ${err}`;
			logger.error('Failed to load restaurants', { error: err });
		} finally {
			loading = false;
		}
	}

	async function deleteRestaurant(id: string) {
		if (!confirm('Are you sure you want to delete this restaurant?')) {
			return;
		}

		loading = true;
		error = null;
		try {
			await apiClient.deleteRestaurant(id);
			restaurants = restaurants.filter((r) => r.id !== id);
			logger.info('Restaurant deleted successfully', { id });
		} catch (err) {
			error = `Failed to delete restaurant: ${err}`;
			console.error('Error deleting restaurant:', err);
		} finally {
			loading = false;
		}
	}

	function handleSearchResults(event: CustomEvent<Partial<Place>[]>) {
		searchResults = event.detail;
		searchError = null;
	}

	function handleSearchError(event: CustomEvent<string>) {
		searchError = event.detail;
		searchResults = [];
	}

	async function addFromGoogle(place: Partial<Place>) {
		if (!place.name) return;

		loading = true;
		error = null;
		try {
			const restaurantData = {
				name: place.name,
				address: place.formatted_address || place.vicinity || 'N/A',
				// Google rating is 0-5, our rating is 1-5. We can show google rating separately.
				// For now, let's not set a rating.
			};
			await apiClient.createRestaurant(restaurantData);
			await loadRestaurants();
			searchResults = []; // Clear search results after adding
		} catch (err) {
			error = `Failed to add restaurant: ${err}`;
		} finally {
			loading = false;
		}
	}

	async function updateRating(restaurantId: string, newRating: number) {
		const originalRestaurants = [...restaurants];
		// Optimistically update UI
		restaurants = restaurants.map((r) => (r.id === restaurantId ? { ...r, rating: newRating } : r));

		try {
			await apiClient.updateRestaurant(restaurantId, { rating: newRating });
			logger.info('Restaurant rating updated', { id: restaurantId, rating: newRating });
		} catch (err) {
			// Revert on error
			restaurants = originalRestaurants;
			error = `Failed to update rating: ${err}`;
			logger.error('Failed to update rating', { error: err });
		}
	}

	onMount(() => {
		loadRestaurants();
	});
</script>

<ProtectedRoute>
	<div class="container mx-auto max-w-6xl p-6 space-y-6">
		<h1 class="text-3xl font-bold tracking-tight">Restaurants</h1>

		{#if error}
			<Alert variant="destructive">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<Card>
			<CardHeader>
				<CardTitle>Find a Restaurant</CardTitle>
				<CardDescription>Search for a restaurant on Google or add one manually.</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="flex items-center space-x-2">
					<div class="flex-grow">
						<RestaurantSearch
							on:searchresults={handleSearchResults}
							on:searcherror={handleSearchError}
						/>
					</div>
					<button
						class={cn(buttonVariants({ variant: 'default' }))}
						on:click={() => (modalOpen = true)}>Add Manually</button
					>
				</div>
			</CardContent>
		</Card>

		{#if searchError}
			<Alert variant="destructive">
				<AlertDescription>{searchError}</AlertDescription>
			</Alert>
		{/if}

		{#if searchResults.length > 0}
			<Card>
				<CardHeader>
					<CardTitle>Google Search Results</CardTitle>
				</CardHeader>
				<CardContent>
					<div class="rounded-md border">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Name</TableHead>
									<TableHead>Address</TableHead>
									<TableHead>Google Rating</TableHead>
									<TableHead>Actions</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each searchResults as place (place.place_id)}
									<TableRow>
										<TableCell>{place.name}</TableCell>
										<TableCell>{place.formatted_address || place.vicinity}</TableCell>
										<TableCell
											>{place.rating
												? `${place.rating} (${place.user_ratings_total} reviews)`
												: 'N/A'}</TableCell
										>
										<TableCell>
											<button
												class={cn(buttonVariants({ size: 'default' }))}
												on:click={() => addFromGoogle(place)}>Add</button
											>
										</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					</div>
				</CardContent>
			</Card>
		{/if}

		<Card>
			<CardHeader>
				<CardTitle>All Restaurants ({restaurants.length})</CardTitle>
			</CardHeader>
			<CardContent>
				{#if loading && restaurants.length === 0}
					<div class="text-center text-muted-foreground py-8">Loading restaurants...</div>
				{:else if restaurants.length === 0}
					<div class="text-center text-muted-foreground py-8">
						No restaurants found. Add one using the form above!
					</div>
				{:else}
					<div class="rounded-md border">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Name</TableHead>
									<TableHead>Address</TableHead>
									<TableHead>Rating</TableHead>
									<TableHead>Comment</TableHead>
									<TableHead>Added</TableHead>
									<TableHead>Actions</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each restaurants as restaurant (restaurant.id)}
									<TableRow>
										<TableCell class="font-medium">{restaurant.name}</TableCell>
										<TableCell class="text-muted-foreground">{restaurant.address || '-'}</TableCell>
										<TableCell>
											<StarRating
												rating={restaurant.rating}
												editable={true}
												on:rate={(e) => updateRating(restaurant.id, e.detail.rating)}
											/>
										</TableCell>
										<TableCell class="max-w-xs truncate text-muted-foreground">
											{restaurant.comment || '-'}
										</TableCell>
										<TableCell class="text-muted-foreground">
											{new Date(restaurant.createdAt).toLocaleDateString()}
										</TableCell>
										<TableCell>
											<button
												class={cn(buttonVariants({ variant: 'default', size: 'default' }))}
												on:click={() => deleteRestaurant(restaurant.id)}
												disabled={loading}>Delete</button
											>
										</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>

	<AddRestaurantModal bind:open={modalOpen} on:added={loadRestaurants} />
</ProtectedRoute>
