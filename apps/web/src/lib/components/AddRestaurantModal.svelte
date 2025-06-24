<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { apiClient } from '$lib/api';
	import { createPageLogger } from '$lib/logger';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
	} from '$lib/components/ui/dialog/index.js';

	export let open: boolean;

	let loading = false;
	let error: string | null = null;
	const logger = createPageLogger('AddRestaurantModal');
	const dispatch = createEventDispatcher();

	let formData = {
		name: '',
		address: '',
		rating: null as number | null,
		comment: '',
	};

	async function handleSubmit() {
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

			formData = { name: '', address: '', rating: null, comment: '' };
			dispatch('added');
			open = false;
		} catch (err) {
			error = `Failed to create restaurant: ${err}`;
			logger.error('Failed to create restaurant', { error: err, formData });
		} finally {
			loading = false;
		}
	}
</script>

<Dialog bind:open>
	<DialogContent class="sm:max-w-[425px]">
		<DialogHeader>
			<DialogTitle>Add Manually</DialogTitle>
			<DialogDescription>
				Add a new restaurant to your list if you can't find it on Google.
			</DialogDescription>
		</DialogHeader>

		{#if error}
			<Alert variant="destructive">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<form on:submit|preventDefault={handleSubmit} class="space-y-4 py-4">
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
				<Input type="text" id="address" bind:value={formData.address} placeholder="Enter address" />
			</div>
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

			<DialogFooter>
				<Button type="submit" disabled={loading}>
					{loading ? 'Adding...' : 'Add Restaurant'}
				</Button>
			</DialogFooter>
		</form>
	</DialogContent>
</Dialog>
