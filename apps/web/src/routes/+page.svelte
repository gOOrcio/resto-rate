<script lang="ts">
	import RestaurantSearchSv from '$lib/ui/components/RestaurantSearchSv.svelte';
</script>

<div class="container mx-auto max-w-6xl space-y-8 p-6">
	<section class="space-y-6">
		<h2 class="text-primary-800 dark:text-primary-200 text-2xl font-semibold">
			Google Places API Search
		</h2>

		<div class="space-y-4">
			<RestaurantSearchSv />
		</div>
	</section>
</div>
