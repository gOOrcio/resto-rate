<script lang="ts">
	import clients from '$lib/client/client';
	import type { RestaurantProto } from '$lib/client/generated/restaurants/v1/restaurant_pb';

	let restaurants: RestaurantProto[] = [];
	let loading = false;
	let error = '';

	async function fetchRestaurants(): Promise<void> {
		loading = true;
		error = '';
		try {
			const response = await clients.restaurants.listRestaurants({ page: 1, pageSize: 20 });
			restaurants = response.restaurants ?? [];
		} catch (e: any) {
			if (e.code && e.details) {
				error = `Error ${e.code}: ${e.details}`;
			} else if (e.message) {
				error = e.message;
			} else {
				error = 'Failed to fetch restaurants';
			}
		} finally {
			loading = false;
		}
	}
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>
<button type="button" class="btn preset-filled-primary-500" onclick={fetchRestaurants}>Load Restaurants</button>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p style="color: red;">{error}</p>
{:else if restaurants.length}
	<ul>
		{#each restaurants as restaurant}
			<li>{restaurant.name}</li>
		{/each}
	</ul>
{/if}
