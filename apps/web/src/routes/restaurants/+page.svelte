<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api';

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
		try {
			const response = (await apiClient.getRestaurants()) as { restaurants: Restaurant[] };
			restaurants = response.restaurants || [];
		} catch (err) {
			error = `Failed to load restaurants: ${err}`;
			console.error('Error loading restaurants:', err);
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
		try {
			const restaurantData = {
				name: formData.name.trim(),
				address: formData.address.trim() || undefined,
				rating: formData.rating || undefined,
				comment: formData.comment.trim() || undefined,
			};

			await apiClient.createRestaurant(restaurantData);

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
			console.error('Error creating restaurant:', err);
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
			await loadRestaurants();
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

<div class="container mx-auto max-w-6xl p-6">
	<h1 class="mb-6 text-3xl font-bold">Restaurants</h1>

	{#if error}
		<div class="mb-4 rounded border border-red-400 bg-red-100 px-4 py-3 text-red-700">
			{error}
		</div>
	{/if}

	<!-- Create Restaurant Form -->
	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<h2 class="text-xl font-semibold mb-4">Add New Restaurant</h2>

		<form on:submit|preventDefault={createRestaurant} class="space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 mb-1">
						Restaurant Name *
					</label>
					<input
						type="text"
						id="name"
						bind:value={formData.name}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Enter restaurant name"
						required
					/>
				</div>

				<div>
					<label for="address" class="block text-sm font-medium text-gray-700 mb-1">
						Address
					</label>
					<input
						type="text"
						id="address"
						bind:value={formData.address}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Enter address"
					/>
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<label for="rating" class="block text-sm font-medium text-gray-700 mb-1">
						Rating (1-5)
					</label>
					<input
						type="number"
						id="rating"
						bind:value={formData.rating}
						min="1"
						max="5"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="1-5"
					/>
				</div>

				<div>
					<label for="comment" class="block text-sm font-medium text-gray-700 mb-1">
						Comment
					</label>
					<input
						type="text"
						id="comment"
						bind:value={formData.comment}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Enter your comment"
					/>
				</div>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition disabled:opacity-50"
			>
				{loading ? 'Adding...' : 'Add Restaurant'}
			</button>
		</form>
	</div>

	<!-- Restaurants Table -->
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<div class="px-6 py-4 border-b border-gray-200">
			<h2 class="text-xl font-semibold">All Restaurants ({restaurants.length})</h2>
		</div>

		{#if loading && restaurants.length === 0}
			<div class="px-6 py-8 text-center text-gray-500">Loading restaurants...</div>
		{:else if restaurants.length === 0}
			<div class="px-6 py-8 text-center text-gray-500">
				No restaurants found. Add one using the form above!
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Name
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Address
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Rating
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Comment
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Added
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each restaurants as restaurant (restaurant.id)}
							<tr>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm font-medium text-gray-900">{restaurant.name}</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm text-gray-500">{restaurant.address || '-'}</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									{#if restaurant.rating}
										<div class="text-sm text-gray-900">
											{'★'.repeat(restaurant.rating)}{'☆'.repeat(5 - restaurant.rating)}
											<span class="ml-1">({restaurant.rating}/5)</span>
										</div>
									{:else}
										<div class="text-sm text-gray-500">-</div>
									{/if}
								</td>
								<td class="px-6 py-4">
									<div class="text-sm text-gray-500 max-w-xs truncate">
										{restaurant.comment || '-'}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm text-gray-500">
										{new Date(restaurant.createdAt).toLocaleDateString()}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<button
										on:click={() => deleteRestaurant(restaurant.id)}
										disabled={loading}
										class="text-red-600 hover:text-red-900 disabled:opacity-50"
									>
										Delete
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<div class="mt-6">
		<a href="/" class="text-blue-500 hover:text-blue-700">← Back to Home</a>
	</div>
</div>
