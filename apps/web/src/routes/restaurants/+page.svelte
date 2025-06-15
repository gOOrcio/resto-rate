<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api';
	import { createPageLogger } from '$lib/logger';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { AuthApi } from '$lib/client';

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
	const api = new AuthApi();

	// Create page-specific logger - safe for SSR
	const logger = createPageLogger('restaurants');

	// Form data
	let formData = {
		name: '',
		address: '',
		rating: null as number | null,
		comment: '',
	};

	async function loadRestaurants() {
		loading = true;
		error = null;
		logger.debug('Loading restaurants list');
		try {
			const response = await api.getRestaurants();
			restaurants = response.data;
			logger.info('Restaurants loaded successfully', { count: restaurants.length });
		} catch (err) {
			error = `Failed to load restaurants: ${err}`;
			logger.error('Failed to load restaurants', { error: err });
		} finally {
			loading = false;
		}
	}

	async function createRestaurant() {
		if (!formData.name.trim()) {
			error = 'Restaurant name is required';
			return;
		}

		if (formData.rating && (formData.rating < 1 || formData.rating > 5)) {
			error = 'Rating must be between 1 and 5';
			return;
		}

		loading = true;
		error = null;
		logger.debug('Creating new restaurant', { name: formData.name });
		try {
			const restaurantData = {
				name: formData.name.trim(),
				address: formData.address.trim() || undefined,
				rating: formData.rating || undefined,
				comment: formData.comment.trim() || undefined,
			};

			await apiClient.createRestaurant(restaurantData);
			logger.info('Restaurant created successfully', { name: restaurantData.name });

			// Reset form
			formData = {
				name: '',
				address: '',
				rating: null,
				comment: '',
			};

			// Reload restaurants
			await loadRestaurants();
		} catch (err) {
			error = `Failed to create restaurant: ${err}`;
			logger.error('Failed to create restaurant', { error: err, formData });
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
			await api.deleteRestaurant(id);
			restaurants = restaurants.filter(r => r.id !== id);
			logger.info('Restaurant deleted successfully', { id });
		} catch (err) {
			error = `Failed to delete restaurant: ${err}`;
			console.error('Error deleting restaurant:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadRestaurants();
	});
</script>

<div class="container mx-auto max-w-6xl p-6 space-y-6">
	<h1 class="text-3xl font-bold tracking-tight">Restaurants</h1>

	{#if error}
		<Alert variant="destructive">
			{#snippet children()}
				<AlertDescription>{error}</AlertDescription>
			{/snippet}
		</Alert>
	{/if}

	<Card>
		<CardHeader>
			<CardTitle>Add New Restaurant</CardTitle>
			<CardDescription>Fill in the details to add a new restaurant to the list.</CardDescription>
		</CardHeader>
		<CardContent>
			<form on:submit|preventDefault={createRestaurant} class="space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="space-y-2">
						<Label for="name">Restaurant Name *</Label>
						<Input
							type="text"
							id="name"
							bind:value={formData.name}
							placeholder="Enter restaurant name"
							required
						/>
					</div>

					<div class="space-y-2">
						<Label for="address">Address</Label>
						<Input
							type="text"
							id="address"
							bind:value={formData.address}
							placeholder="Enter address"
						/>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="space-y-2">
						<Label for="rating">Rating (1-5)</Label>
						<Input
							type="number"
							id="rating"
							bind:value={formData.rating}
							min="1"
							max="5"
							placeholder="1-5"
						/>
					</div>

					<div class="space-y-2">
						<Label for="comment">Comment</Label>
						<Input
							type="text"
							id="comment"
							bind:value={formData.comment}
							placeholder="Enter your comment"
						/>
					</div>
				</div>

				<Button type="submit" disabled={loading}>
					{loading ? 'Adding...' : 'Add Restaurant'}
				</Button>
			</form>
		</CardContent>
	</Card>

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
										{#if restaurant.rating}
											<div class="text-foreground">
												{'★'.repeat(restaurant.rating)}{'☆'.repeat(5 - restaurant.rating)}
												<span class="ml-1 text-muted-foreground">({restaurant.rating}/5)</span>
											</div>
										{:else}
											<span class="text-muted-foreground">-</span>
										{/if}
									</TableCell>
									<TableCell class="max-w-xs truncate text-muted-foreground">
										{restaurant.comment || '-'}
									</TableCell>
									<TableCell class="text-muted-foreground">
										{new Date(restaurant.createdAt).toLocaleDateString()}
									</TableCell>
									<TableCell>
										<Button
											variant="destructive"
											size="sm"
											on:click={(e) => deleteRestaurant(restaurant.id)}
											disabled={loading}
										>
											Delete
										</Button>
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				</div>
			{/if}
		</CardContent>
	</Card>

	<div>
		<Button variant="link" href="/" class="p-0">← Back to Home</Button>
	</div>
</div>
