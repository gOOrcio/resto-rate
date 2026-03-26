<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Star, X } from '@lucide/svelte';
	import client from '$lib/client/client';
	import type { ReviewProto } from '$lib/client/generated/reviews/v1/review_pb';

	const {
		googlePlacesId,
		restaurantName,
		restaurantAddress,
		existingReview,
		onSubmit
	} = $props<{
		googlePlacesId: string;
		restaurantName: string;
		restaurantAddress: string;
		existingReview?: ReviewProto;
		onSubmit: (review: ReviewProto) => void;
	}>();

	let rating = $state(existingReview?.rating ?? 0);
	let hoverRating = $state(0);
	let comment = $state(existingReview?.comment ?? '');
	let tags = $state<string[]>(existingReview?.tags ? [...existingReview.tags] : []);
	let tagInput = $state('');
	let loading = $state(false);
	let error = $state<string | null>(null);

	const isEdit = $derived(!!existingReview?.id);
	const displayRating = $derived(hoverRating || rating);

	function addTag() {
		const t = tagInput.trim().replace(/,$/, '');
		if (t && !tags.includes(t)) {
			tags = [...tags, t];
		}
		tagInput = '';
	}

	function removeTag(tag: string) {
		tags = tags.filter((t) => t !== tag);
	}

	function handleTagKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ',') {
			e.preventDefault();
			addTag();
		}
	}

	async function handleSubmit() {
		if (rating < 1) {
			error = 'Please select a star rating';
			return;
		}
		error = null;
		loading = true;
		try {
			if (isEdit && existingReview) {
				const res = await client.reviews.updateReview({
					id: existingReview.id,
					comment,
					rating,
					tags
				});
				if (res.review) onSubmit(res.review);
			} else {
				const res = await client.reviews.createReview({
					googlePlacesId,
					restaurantName,
					restaurantAddress,
					comment,
					rating,
					tags
				});
				if (res.review) onSubmit(res.review);
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to save review';
		} finally {
			loading = false;
		}
	}
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl">
	<h4 class="mb-4 text-base font-semibold text-gray-800">
		{isEdit ? 'Edit your rating' : 'Rate this place'}
	</h4>

	<!-- Star picker -->
	<div class="mb-4">
		<Label class="mb-1 block text-sm">Rating *</Label>
		<div class="flex gap-1">
			{#each Array(5) as _, i}
				<button
					type="button"
					onclick={() => (rating = i + 1)}
					onmouseenter={() => (hoverRating = i + 1)}
					onmouseleave={() => (hoverRating = 0)}
					class="transition-transform hover:scale-110"
					aria-label="Rate {i + 1} stars"
				>
					<Star
						class="h-7 w-7 {i < displayRating
							? 'fill-amber-400 text-amber-400'
							: 'fill-none text-gray-300'}"
					/>
				</button>
			{/each}
		</div>
	</div>

	<!-- Comment -->
	<div class="mb-4">
		<Label for="comment" class="mb-1 block text-sm">Comment (optional)</Label>
		<textarea
			id="comment"
			bind:value={comment}
			rows="3"
			placeholder="What did you think?"
			class="w-full resize-none rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
		></textarea>
	</div>

	<!-- Tags -->
	<div class="mb-5">
		<Label for="tag-input" class="mb-1 block text-sm">Tags (optional)</Label>
		<div class="flex flex-wrap gap-1.5 mb-2">
			{#each tags as tag}
				<span class="flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700">
					{tag}
					<button
						type="button"
						onclick={() => removeTag(tag)}
						class="text-blue-500 hover:text-blue-700"
						aria-label="Remove tag {tag}"
					>
						<X class="h-3 w-3" />
					</button>
				</span>
			{/each}
		</div>
		<input
			id="tag-input"
			type="text"
			bind:value={tagInput}
			onkeydown={handleTagKeydown}
			onblur={addTag}
			placeholder="Type a tag and press Enter"
			class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
		/>
		<p class="mt-1 text-xs text-gray-400">Press Enter or comma to add a tag</p>
	</div>

	{#if error}
		<p class="mb-3 text-sm text-red-600">{error}</p>
	{/if}

	<Button onclick={handleSubmit} disabled={loading || rating < 1} class="w-full">
		{loading ? 'Saving…' : isEdit ? 'Update rating' : 'Save rating'}
	</Button>
</div>
